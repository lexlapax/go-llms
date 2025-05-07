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
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	// Optimization: cache for converted messages
	messageCache *MessageCache
}

// AnthropicOption configures the Anthropic provider
type AnthropicOption func(*AnthropicProvider)

// WithAnthropicBaseURL sets a custom base URL for the Anthropic API
func WithAnthropicBaseURL(url string) AnthropicOption {
	return func(p *AnthropicProvider) {
		p.baseURL = url
	}
}

// WithAnthropicHTTPClient sets a custom HTTP client
func WithAnthropicHTTPClient(client *http.Client) AnthropicOption {
	return func(p *AnthropicProvider) {
		p.httpClient = client
	}
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, model string, options ...AnthropicOption) *AnthropicProvider {
	provider := &AnthropicProvider{
		apiKey:      apiKey,
		model:       model,
		baseURL:     defaultAnthropicBaseURL,
		httpClient:  http.DefaultClient,
		messageCache: NewMessageCache(),
	}

	for _, option := range options {
		option(provider)
	}

	return provider
}

// Generate produces text from a prompt
func (p *AnthropicProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	messages := []domain.Message{
		{Role: domain.RoleUser, Content: prompt},
	}
	response, err := p.GenerateMessage(ctx, messages, options...)
	if err != nil {
		return "", err
	}
	return response.Content, nil
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
	
	// Fast path for single message
	if len(messages) == 1 {
		if messages[0].Role == domain.RoleSystem {
			systemMessage = messages[0].Content
			// Cache and return
			result := map[string]interface{}{
				"messages": anthMessages,
				"system":   systemMessage,
			}
			p.messageCache.Set(cacheKey, result)
			return anthMessages, systemMessage
		} else {
			message := make(map[string]interface{}, 2)
			message["role"] = string(messages[0].Role)
			message["content"] = messages[0].Content
			anthMessages = append(anthMessages, message)
			
			// Cache and return
			result := map[string]interface{}{
				"messages": anthMessages,
				"system":   systemMessage,
			}
			p.messageCache.Set(cacheKey, result)
			return anthMessages, systemMessage
		}
	}
	
	// Process all messages
	for _, msg := range messages {
		if msg.Role == domain.RoleSystem {
			// Anthropic handles system message separately
			systemMessage = msg.Content
		} else {
			// Regular message (user or assistant)
			message := make(map[string]interface{}, 2)
			message["role"] = string(msg.Role)
			message["content"] = msg.Content
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
	
	// Add system message if present
	if systemMessage != "" {
		requestBody["system"] = systemMessage
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
		var errorResp struct {
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}
		// Use optimized JSON unmarshaling - significantly faster than standard library
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return domain.Response{}, fmt.Errorf("API error: %s: %s", errorResp.Error.Type, errorResp.Error.Message)
		}
		return domain.Response{}, fmt.Errorf("API error: status code %d", resp.StatusCode)
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
	messages := []domain.Message{
		{Role: domain.RoleUser, Content: prompt},
	}
	return p.StreamMessage(ctx, messages, options...)
}

// StreamMessage streams responses from a list of messages
func (p *AnthropicProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
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
		resp.Body.Close()
		return nil, fmt.Errorf("API error: status code %d", resp.StatusCode)
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