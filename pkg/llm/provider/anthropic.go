package provider

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
	"github.com/lexlapax/go-llms/pkg/util/json"
)

const (
	defaultAnthropicBaseURL = "https://api.anthropic.com"
)

// AnthropicProvider implements the Provider interface for Anthropic
type AnthropicProvider struct {
	apiKey       string
	model        string
	baseURL      string
	httpClient   *http.Client
	systemPrompt string
	metadata     map[string]string
	// Optimization: cache for converted messages
	messageCache *MessageCache
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, model string, options ...domain.ProviderOption) *AnthropicProvider {
	provider := &AnthropicProvider{
		apiKey:       apiKey,
		model:        model,
		baseURL:      defaultAnthropicBaseURL,
		httpClient:   http.DefaultClient,
		metadata:     make(map[string]string),
		messageCache: NewMessageCache(),
	}

	for _, option := range options {
		// Check if the option is compatible with Anthropic
		if anthropicOption, ok := option.(domain.AnthropicOption); ok {
			anthropicOption.ApplyToAnthropic(provider)
		}
	}

	return provider
}

// Setter methods for options
// SetBaseURL sets the base URL for the Anthropic API
func (p *AnthropicProvider) SetBaseURL(url string) {
	p.baseURL = url
}

// SetHTTPClient sets the HTTP client
func (p *AnthropicProvider) SetHTTPClient(client *http.Client) {
	p.httpClient = client
}

// SetSystemPrompt sets the system prompt for Anthropic API calls
func (p *AnthropicProvider) SetSystemPrompt(systemPrompt string) {
	p.systemPrompt = systemPrompt
}

// SetMetadata sets the metadata for Anthropic API calls
func (p *AnthropicProvider) SetMetadata(metadata map[string]string) {
	p.metadata = metadata
}

// Generate produces text from a prompt
func (p *AnthropicProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	// Create a simple text message using the new structure
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, prompt),
	}
	response, err := p.GenerateMessage(ctx, messages, options...)
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// validateContentTypesForAnthropic checks if the content types in the messages are supported by Anthropic
func (p *AnthropicProvider) validateContentTypesForAnthropic(messages []domain.Message) error {
	for _, msg := range messages {
		if msg.Content != nil {
			for _, part := range msg.Content {
				// Anthropic currently supports text and image content types
				if part.Type != domain.ContentTypeText && part.Type != domain.ContentTypeImage {
					return domain.NewUnsupportedContentTypeError("Anthropic", part.Type)
				}
			}
		}
	}
	return nil
}

// ConvertMessagesToAnthropicFormat converts domain messages to Anthropic format
// Optimized version with caching and reduced allocations
// This method is exported for benchmarking purposes
func (p *AnthropicProvider) ConvertMessagesToAnthropicFormat(messages []domain.Message) ([]map[string]interface{}, string) {
	// Check cache first
	cacheKey := GenerateMessagesKey(messages)
	if cachedResult, found := p.messageCache.Get(cacheKey); found {
		result := cachedResult.(map[string]interface{})
		return result["messages"].([]map[string]interface{}), result["system"].(string)
	}

	// Pre-allocate the slice with reasonable capacity
	// Anthropic handles system messages separately, so capacity is potentially less
	anthMessages := make([]map[string]interface{}, 0, len(messages))
	var systemMessage string

	// Process all messages
	for _, msg := range messages {
		if msg.Role == domain.RoleSystem {
			// Extract text content for system message
			if msg.Content != nil && len(msg.Content) > 0 {
				for _, part := range msg.Content {
					if part.Type == domain.ContentTypeText {
						systemMessage = part.Text
						break
					}
				}
			}
		} else {
			// Regular message (user or assistant)
			message := make(map[string]interface{}, 2)
			message["role"] = string(msg.Role)

			// Handle multimodal content
			if msg.Content != nil && len(msg.Content) > 0 {
				contentParts := make([]map[string]interface{}, 0, len(msg.Content))

				for _, part := range msg.Content {
					switch part.Type {
					case domain.ContentTypeText:
						// Text part
						contentParts = append(contentParts, map[string]interface{}{
							"type": "text",
							"text": part.Text,
						})
					case domain.ContentTypeImage:
						// Image part - Anthropic format
						imagePart := map[string]interface{}{
							"type": "image",
						}

						sourcePart := make(map[string]interface{})
						if part.Image.Source.Type == domain.SourceTypeURL {
							// URL-based image
							sourcePart["type"] = "url"
							sourcePart["url"] = part.Image.Source.URL
						} else {
							// Base64-encoded image
							sourcePart["type"] = "base64"
							sourcePart["media_type"] = part.Image.Source.MediaType
							sourcePart["data"] = part.Image.Source.Data
						}

						imagePart["source"] = sourcePart
						contentParts = append(contentParts, imagePart)
					}
				}

				// Add content parts to the message
				message["content"] = contentParts
			}

			anthMessages = append(anthMessages, message)
		}
	}

	// Cache the result - store both the messages and system prompt
	result := map[string]interface{}{
		"messages": anthMessages,
		"system":   systemMessage,
	}
	p.messageCache.Set(cacheKey, result)

	return anthMessages, systemMessage
}

// buildAnthropicRequestBody creates a request body for the Anthropic API
func (p *AnthropicProvider) buildAnthropicRequestBody(
	messages []map[string]interface{},
	systemMessage string,
	options *domain.ProviderOptions,
) map[string]interface{} {
	// Pre-allocate the map with the right capacity (standard fields + possible options)
	// We need at least model and messages, plus potential options
	requestBody := make(map[string]interface{}, 6)

	// Add required fields
	requestBody["model"] = p.model
	requestBody["messages"] = messages

	// Add system message if present from messages or from provider configuration
	if systemMessage != "" {
		requestBody["system"] = systemMessage
	} else if p.systemPrompt != "" {
		requestBody["system"] = p.systemPrompt
	}

	// Add metadata if present
	if len(p.metadata) > 0 {
		requestBody["metadata"] = p.metadata
	}

	// Add temperature if it differs from default
	if options.Temperature != 0.7 {
		requestBody["temperature"] = options.Temperature
	}

	// Always add max_tokens - Anthropic requires this field
	requestBody["max_tokens"] = options.MaxTokens

	// Add top_p if it differs from default
	if options.TopP != 1.0 {
		requestBody["top_p"] = options.TopP
	}

	// Add stop sequences if provided
	if len(options.StopSequences) > 0 {
		requestBody["stop_sequences"] = options.StopSequences
	}

	return requestBody
}

// GenerateMessage produces text from a list of messages - optimized version
func (p *AnthropicProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	// Validate content types
	if err := p.validateContentTypesForAnthropic(messages); err != nil {
		return domain.Response{}, err
	}

	// Apply options
	providerOptions := domain.DefaultOptions()
	for _, option := range options {
		option(providerOptions)
	}

	// Convert messages to Anthropic format - optimized with caching
	anthMessages, systemMessage := p.ConvertMessagesToAnthropicFormat(messages)

	// Build request body - optimized with pre-allocation
	requestBody := p.buildAnthropicRequestBody(anthMessages, systemMessage, providerOptions)

	// Use optimized JSON marshaling with buffer reuse for request body
	requestBuffer := &bytes.Buffer{}
	err := json.MarshalWithBuffer(requestBody, requestBuffer)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request - reuse the buffer directly
	url := fmt.Sprintf("%s/v1/messages", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, requestBuffer)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01") // Use appropriate API version

	// Make the request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for error
	if resp.StatusCode != http.StatusOK {
		return domain.Response{}, ParseJSONError(body, resp.StatusCode, "anthropic", "GenerateMessage")
	}

	// Parse response
	var anthropicResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	// Use optimized unmarshaling which is ~2x faster than standard library
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return domain.Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text from response
	var responseContent string
	for _, content := range anthropicResp.Content {
		if content.Type == "text" {
			responseContent = content.Text
			break
		}
	}

	// Use the response pool to reduce allocations
	return domain.GetResponsePool().NewResponse(responseContent), nil
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *AnthropicProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	// Build a prompt that includes the schema
	enhancedPrompt := enhancePromptWithAnthropicSchema(prompt, schema)

	// Generate response
	response, err := p.Generate(ctx, enhancedPrompt, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Try to extract JSON from the response using optimized extractor
	jsonStr := processor.ExtractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("response does not contain valid JSON")
	}

	// Parse the JSON into a map - use optimized JSON unmarshaling
	var result interface{}
	if err := json.UnmarshalFromString(jsonStr, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return result, nil
}

// Stream streams responses token by token
func (p *AnthropicProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	// Create a simple text message using the new structure
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, prompt),
	}
	return p.StreamMessage(ctx, messages, options...)
}

// StreamMessage streams responses from a list of messages
func (p *AnthropicProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	// Validate content types
	if err := p.validateContentTypesForAnthropic(messages); err != nil {
		return nil, err
	}

	// Apply options
	providerOptions := domain.DefaultOptions()
	for _, option := range options {
		option(providerOptions)
	}

	// Convert messages to Anthropic format - optimized with caching
	anthMessages, systemMessage := p.ConvertMessagesToAnthropicFormat(messages)

	// Build request body - optimized with pre-allocation
	requestBody := p.buildAnthropicRequestBody(anthMessages, systemMessage, providerOptions)

	// Add streaming flag
	requestBody["stream"] = true

	// Use optimized JSON marshaling with buffer reuse for request body
	requestBuffer := &bytes.Buffer{}
	err := json.MarshalWithBuffer(requestBody, requestBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request - reuse the buffer directly
	url := fmt.Sprintf("%s/v1/messages", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, requestBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01") // Use appropriate API version
	req.Header.Set("Accept", "text/event-stream")

	// Make the request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, ParseJSONError(body, resp.StatusCode, "anthropic", "StreamMessage")
	}

	// Get a channel from the pool
	responseStream, tokenCh := domain.GetChannelPool().GetResponseStream()

	// Start a goroutine to read the stream
	go func() {
		defer resp.Body.Close()
		defer close(tokenCh)
		// Return the channel to the pool when done
		// Note: Put will avoid putting closed channels back

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			// Check if context is canceled
			select {
			case <-ctx.Done():
				return
			default:
				// Continue
			}

			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			// Extract the data part
			data := strings.TrimPrefix(line, "data: ")

			// Skip empty data lines
			if data == "" || data == "[DONE]" {
				continue
			}

			// Determine event type
			var event struct {
				Type string `json:"type"`
			}
			// Use optimized JSON unmarshaling - significantly faster than standard library
			if err := json.UnmarshalFromString(data, &event); err != nil {
				continue
			}

			// Process based on event type
			switch event.Type {
			case "content_block_delta":
				var deltaEvent struct {
					Delta struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"delta"`
				}
				// Use optimized JSON unmarshaling from string
				if err := json.UnmarshalFromString(data, &deltaEvent); err != nil {
					continue
				}

				if deltaEvent.Delta.Type == "text_delta" && deltaEvent.Delta.Text != "" {
					// Send the token - use token pool to reduce allocations
					select {
					case <-ctx.Done():
						return
					case tokenCh <- domain.GetTokenPool().NewToken(deltaEvent.Delta.Text, false):
						// Sent successfully
					}
				}
			case "message_delta":
				var stopEvent struct {
					Delta struct {
						StopReason string `json:"stop_reason"`
					} `json:"delta"`
				}
				// Use optimized JSON unmarshaling from string
				if err := json.UnmarshalFromString(data, &stopEvent); err != nil {
					continue
				}

				if stopEvent.Delta.StopReason != "" {
					// Send final token - use token pool to reduce allocations
					select {
					case <-ctx.Done():
						return
					case tokenCh <- domain.GetTokenPool().NewToken("", true):
						return
					}
				}
			case "message_stop":
				// Send final token if not already sent - use token pool to reduce allocations
				select {
				case <-ctx.Done():
					return
				case tokenCh <- domain.GetTokenPool().NewToken("", true):
					return
				}
			}
		}

		// If the scanner ends without a message_stop event, send a final token
		select {
		case <-ctx.Done():
			return
		case tokenCh <- domain.GetTokenPool().NewToken("", true):
			// Sent successfully
		}
	}()

	return responseStream, nil
}

// enhancePromptWithAnthropicSchema is shared with OpenAI provider
// Consider extracting to a common utility package
func enhancePromptWithAnthropicSchema(prompt string, schema *schemaDomain.Schema) string {
	// Reuse buffer for schema JSON - reduces allocations
	var schemaBuffer bytes.Buffer
	schemaBuffer.Grow(1024) // Pre-allocate reasonable buffer size for most schemas

	// Use optimized JSON marshaling with indentation
	err := json.MarshalIndentWithBuffer(schema, &schemaBuffer, "", "  ")
	if err != nil {
		// If we can't marshal the schema, just return the original prompt
		return prompt
	}

	// Build enhanced prompt
	// Estimate the size to pre-allocate the final string builder
	totalSize := len(prompt) + schemaBuffer.Len() + 200 // 200 for the template text

	var promptBuilder strings.Builder
	promptBuilder.Grow(totalSize)

	promptBuilder.WriteString(prompt)
	promptBuilder.WriteString("\n\nYou are to provide a JSON response that conforms to the following JSON schema.")
	promptBuilder.WriteString("\nRespond ONLY with valid JSON that matches this schema:\n\n")
	promptBuilder.Write(schemaBuffer.Bytes())
	promptBuilder.WriteString("\n\nYour response must be valid JSON only, with no explanations, markdown code blocks, or any other text.")

	return promptBuilder.String()
}
