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
	defaultBaseURL = "https://api.openai.com"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey       string
	model        string
	baseURL      string
	httpClient   *http.Client
	organization string
	logitBias    map[string]float64
	// Optimization: cache for converted messages
	messageCache *MessageCache
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string, options ...domain.ProviderOption) *OpenAIProvider {
	provider := &OpenAIProvider{
		apiKey:       apiKey,
		model:        model,
		baseURL:      defaultBaseURL,
		httpClient:   http.DefaultClient,
		messageCache: NewMessageCache(),
	}

	for _, option := range options {
		// Check if the option is compatible with OpenAI
		if openAIOption, ok := option.(domain.OpenAIOption); ok {
			openAIOption.ApplyToOpenAI(provider)
		}
	}

	return provider
}

// Setter methods for options
// SetBaseURL sets the base URL for the OpenAI API
func (p *OpenAIProvider) SetBaseURL(url string) {
	p.baseURL = url
}

// SetHTTPClient sets the HTTP client
func (p *OpenAIProvider) SetHTTPClient(client *http.Client) {
	p.httpClient = client
}

// SetOrganization sets the organization ID
func (p *OpenAIProvider) SetOrganization(org string) {
	p.organization = org
}

// SetLogitBias sets the logit bias
func (p *OpenAIProvider) SetLogitBias(logitBias map[string]float64) {
	p.logitBias = logitBias
}

// Generate produces text from a prompt
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
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

// ConvertMessagesToOpenAIFormat converts domain messages to OpenAI format
// Optimized version with caching and reduced allocations
// This method is exported for benchmarking purposes
func (p *OpenAIProvider) ConvertMessagesToOpenAIFormat(messages []domain.Message) []map[string]interface{} {
	// Check cache first
	cacheKey := GenerateMessagesKey(messages)
	if cachedMessages, found := p.messageCache.Get(cacheKey); found {
		return cachedMessages.([]map[string]interface{})
	}

	// Pre-allocate the slice with exact capacity
	oaiMessages := make([]map[string]interface{}, 0, len(messages))

	// Process all messages
	for i, msg := range messages {
		// Create the basic message with role
		message := make(map[string]interface{})
		message["role"] = string(msg.Role)

		// Check if this is a multimodal message (has Content slice)
		if len(msg.Content) > 0 {
			// This is a multimodal message
			contentParts := make([]map[string]interface{}, 0, len(msg.Content))

			// Process each content part
			for _, part := range msg.Content {
				switch part.Type {
				case domain.ContentTypeText:
					// Text part
					contentParts = append(contentParts, map[string]interface{}{
						"type": "text",
						"text": part.Text,
					})
				case domain.ContentTypeImage:
					// Image part - OpenAI uses image_url format
					imagePart := map[string]interface{}{
						"type": "image_url",
					}

					imageURL := make(map[string]interface{})
					if part.Image.Source.Type == domain.SourceTypeURL {
						// URL-based image
						imageURL["url"] = part.Image.Source.URL
					} else {
						// Base64-encoded image - requires data URL format
						imageURL["url"] = fmt.Sprintf(
							"data:%s;base64,%s",
							part.Image.Source.MediaType,
							part.Image.Source.Data,
						)
					}

					imagePart["image_url"] = imageURL
					contentParts = append(contentParts, imagePart)
				case domain.ContentTypeFile:
					// File part
					contentParts = append(contentParts, map[string]interface{}{
						"type": "file",
						"file": map[string]interface{}{
							"file_name": part.File.FileName,
							"file_data": part.File.FileData,
						},
					})
				case domain.ContentTypeVideo:
					// Video part
					contentParts = append(contentParts, map[string]interface{}{
						"type": "video",
						"video": map[string]interface{}{
							"media_type": part.Video.Source.MediaType,
							"data":       part.Video.Source.Data,
						},
					})
				case domain.ContentTypeAudio:
					// Audio part
					contentParts = append(contentParts, map[string]interface{}{
						"type": "audio",
						"audio": map[string]interface{}{
							"media_type": part.Audio.Source.MediaType,
							"data":       part.Audio.Source.Data,
						},
					})
				}
			}

			// Add the content parts to the message
			message["content"] = contentParts
		} else if msg.Role == domain.RoleTool {
			// Special handling for tool messages - they must follow an assistant message with tool_calls
			// Find the last assistant message index
			var lastAssistantIdx int = -1
			for j, m := range messages {
				if m.Role == domain.RoleAssistant {
					lastAssistantIdx = j
				}
			}

			if lastAssistantIdx == -1 || i == 0 || messages[i-1].Role != domain.RoleAssistant {
				// If this is a tool message without a preceding assistant message with tool_calls,
				// convert to user message as a fallback
				message["role"] = string(domain.RoleUser)
				// Legacy format for backward compatibility - uses first text content part or empty string
				textContent := ""
				if len(msg.Content) > 0 {
					for _, part := range msg.Content {
						if part.Type == domain.ContentTypeText {
							textContent = part.Text
							break
						}
					}
				}
				message["content"] = "Tool result: " + textContent
			} else {
				// This is a valid tool message following an assistant
				// Legacy format for backward compatibility - uses first text content part or empty string
				textContent := ""
				if len(msg.Content) > 0 {
					for _, part := range msg.Content {
						if part.Type == domain.ContentTypeText {
							textContent = part.Text
							break
						}
					}
				}
				message["content"] = textContent
				message["tool_call_id"] = "call_" + string(rune(i))
			}
		} else if msg.Role == domain.RoleAssistant && i < len(messages)-1 && messages[i+1].Role == domain.RoleTool {
			// This assistant message is followed by a tool message, add tool_calls
			// Legacy format for backward compatibility - uses first text content part or empty string
			textContent := ""
			if len(msg.Content) > 0 {
				for _, part := range msg.Content {
					if part.Type == domain.ContentTypeText {
						textContent = part.Text
						break
					}
				}
			}
			message["content"] = textContent

			// Create a single tool call
			functionMap := make(map[string]interface{}, 2)
			functionMap["name"] = "generic_tool"
			functionMap["arguments"] = "{}"

			toolCall := make(map[string]interface{}, 3)
			toolCall["id"] = "call_" + string(rune(i+1))
			toolCall["type"] = "function"
			toolCall["function"] = functionMap

			toolCalls := []map[string]interface{}{toolCall}
			message["tool_calls"] = toolCalls
		} else {
			// Legacy format for backward compatibility - uses first text content part or empty string
			textContent := ""
			if len(msg.Content) > 0 {
				for _, part := range msg.Content {
					if part.Type == domain.ContentTypeText {
						textContent = part.Text
						break
					}
				}
			}
			message["content"] = textContent
		}

		oaiMessages = append(oaiMessages, message)
	}

	// Cache the result
	p.messageCache.Set(cacheKey, oaiMessages)
	return oaiMessages
}

// validateContentTypesForOpenAI checks if the content types in the messages are supported by OpenAI
func (p *OpenAIProvider) validateContentTypesForOpenAI(messages []domain.Message) error {
	// OpenAI supports all content types in our implementation as of now
	// If there are limitations in the future, we can add checks here
	return nil
}

// buildOpenAIRequestBody creates a request body for the OpenAI API
func (p *OpenAIProvider) buildOpenAIRequestBody(
	messages []map[string]interface{},
	options *domain.ProviderOptions,
) map[string]interface{} {
	// Pre-allocate the map with the right capacity (standard fields + possible options)
	requestBody := make(map[string]interface{}, 8)

	// Add required fields
	requestBody["model"] = p.model
	requestBody["messages"] = messages

	// Add common options if they're not default values
	if options.Temperature != 0.7 {
		requestBody["temperature"] = options.Temperature
	}

	if options.MaxTokens != 1024 {
		requestBody["max_tokens"] = options.MaxTokens
	}

	if options.TopP != 1.0 {
		requestBody["top_p"] = options.TopP
	}

	// Add optional fields only if they have values
	if len(options.StopSequences) > 0 {
		requestBody["stop"] = options.StopSequences
	}

	if options.FrequencyPenalty != 0 {
		requestBody["frequency_penalty"] = options.FrequencyPenalty
	}

	if options.PresencePenalty != 0 {
		requestBody["presence_penalty"] = options.PresencePenalty
	}

	// Add logit bias if provided
	if len(p.logitBias) > 0 {
		requestBody["logit_bias"] = p.logitBias
	}

	return requestBody
}

// GenerateMessage produces text from a list of messages - optimized version
func (p *OpenAIProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	// Validate content types
	if err := p.validateContentTypesForOpenAI(messages); err != nil {
		return domain.Response{}, err
	}

	// Apply options - reuse the same options object for all requests
	providerOptions := domain.DefaultOptions()
	for _, option := range options {
		option(providerOptions)
	}

	// Convert messages to OpenAI format - optimized with caching
	oaiMessages := p.ConvertMessagesToOpenAIFormat(messages)

	// Build request body - optimized with pre-allocation
	requestBody := p.buildOpenAIRequestBody(oaiMessages, providerOptions)

	// Use optimized JSON marshaling with buffer reuse for request body
	requestBuffer := &bytes.Buffer{}
	err := json.MarshalWithBuffer(requestBody, requestBuffer)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request - reuse the buffer directly
	url := fmt.Sprintf("%s/v1/chat/completions", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, requestBuffer)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	// Set organization header if provided
	if p.organization != "" {
		req.Header.Set("OpenAI-Organization", p.organization)
	}

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
		return domain.Response{}, ParseJSONError(body, resp.StatusCode, "openai", "GenerateMessage")
	}

	// Parse response - use optimized JSON unmarshaling
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	// Use optimized unmarshaling which is ~2x faster than standard library
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return domain.Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if there are choices
	if len(openAIResp.Choices) == 0 {
		return domain.Response{}, fmt.Errorf("API returned no choices")
	}

	// Use the response pool to reduce allocations
	return domain.GetResponsePool().NewResponse(openAIResp.Choices[0].Message.Content), nil
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *OpenAIProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	// Build a prompt that includes the schema
	enhancedPrompt := enhancePromptWithSchema(prompt, schema)

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
func (p *OpenAIProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	// Create a simple text message using the new structure
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, prompt),
	}
	return p.StreamMessage(ctx, messages, options...)
}

// StreamMessage streams responses from a list of messages
func (p *OpenAIProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	// Validate content types
	if err := p.validateContentTypesForOpenAI(messages); err != nil {
		return nil, err
	}

	// Apply options
	providerOptions := domain.DefaultOptions()
	for _, option := range options {
		option(providerOptions)
	}

	// Convert messages to OpenAI format - optimized with caching
	oaiMessages := p.ConvertMessagesToOpenAIFormat(messages)

	// Build request body - optimized with pre-allocation
	requestBody := p.buildOpenAIRequestBody(oaiMessages, providerOptions)

	// Add streaming flag
	requestBody["stream"] = true

	// Use optimized JSON marshaling with buffer reuse for request body
	requestBuffer := &bytes.Buffer{}
	err := json.MarshalWithBuffer(requestBody, requestBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request - reuse the buffer directly
	url := fmt.Sprintf("%s/v1/chat/completions", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, requestBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	req.Header.Set("Accept", "text/event-stream")

	// Set organization header if provided
	if p.organization != "" {
		req.Header.Set("OpenAI-Organization", p.organization)
	}

	// Make the request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, ParseJSONError(body, resp.StatusCode, "openai", "StreamMessage")
	}

	// Get a channel from the pool
	responseStream, tokenCh := domain.GetChannelPool().GetResponseStream()

	// Start a goroutine to read the stream
	go func() {
		defer resp.Body.Close()
		defer close(tokenCh)
		// Return the channel to the pool when done
		// Note: Put will avoid putting closed channels back

		reader := bufio.NewReader(resp.Body)
		for {
			// Check if context is canceled
			select {
			case <-ctx.Done():
				return
			default:
				// Continue
			}

			// Read a line from the response
			line, err := reader.ReadString('\n')
			if err != nil {
				// Just exit the loop on any error, as it could be EOF
				return
			}

			// Process the line if it contains data
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			// Extract the data part
			data := strings.TrimPrefix(line, "data: ")

			// Check for end of stream
			if data == "[DONE]" {
				return
			}

			// Parse the JSON
			var streamResp struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
			}

			// Use optimized JSON unmarshaling from string - significantly faster than standard library
			if err := json.UnmarshalFromString(data, &streamResp); err != nil {
				// Skip invalid JSON
				continue
			}

			// Check if there are choices
			if len(streamResp.Choices) == 0 {
				continue
			}

			choice := streamResp.Choices[0]
			content := choice.Delta.Content

			// If content is empty and finish_reason is set, it means we're done
			if content == "" && choice.FinishReason != nil {
				// Send final token if needed
				select {
				case <-ctx.Done():
					return
				case tokenCh <- domain.GetTokenPool().NewToken("", true):
					return
				}
			}

			// Skip empty content
			if content == "" {
				continue
			}

			// Send the token - use token pool to reduce allocations
			select {
			case <-ctx.Done():
				return
			case tokenCh <- domain.GetTokenPool().NewToken(content, false):
				// Sent successfully
			}
		}
	}()

	return responseStream, nil
}

// enhancePromptWithSchema adds schema information to a prompt
func enhancePromptWithSchema(prompt string, schema *schemaDomain.Schema) string {
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
	promptBuilder.WriteString("\n\nOutput only valid JSON without any explanations, markdown code blocks, or any other text.")

	return promptBuilder.String()
}

// Note: extractJSON has been replaced with processor.ExtractJSON for better performance and reliability
