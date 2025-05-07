package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
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
		apiKey:     apiKey,
		model:      model,
		baseURL:    defaultAnthropicBaseURL,
		httpClient: http.DefaultClient,
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

// GenerateMessage produces text from a list of messages
func (p *AnthropicProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	// Apply options
	providerOptions := domain.DefaultOptions()
	for _, option := range options {
		option(providerOptions)
	}

	// Convert domain messages to Anthropic messages format
	anthMessages := make([]map[string]interface{}, 0, len(messages))
	var systemMessage string

	for _, msg := range messages {
		if msg.Role == domain.RoleSystem {
			// Anthropic handles system message separately
			systemMessage = msg.Content
		} else {
			anthMessages = append(anthMessages, map[string]interface{}{
				"role":    string(msg.Role),
				"content": msg.Content,
			})
		}
	}

	// Prepare request body
	requestBody := map[string]interface{}{
		"model":    p.model,
		"messages": anthMessages,
	}

	// Add system message if present
	if systemMessage != "" {
		requestBody["system"] = systemMessage
	}

	// Add temperature if provided
	if providerOptions.Temperature != 0 {
		requestBody["temperature"] = providerOptions.Temperature
	}

	// Add max tokens if provided
	if providerOptions.MaxTokens != 0 {
		requestBody["max_tokens"] = providerOptions.MaxTokens
	}

	// Add top p if provided
	if providerOptions.TopP != 0 {
		requestBody["top_p"] = providerOptions.TopP
	}

	// Add stop sequences if provided
	if len(providerOptions.StopSequences) > 0 {
		requestBody["stop_sequences"] = providerOptions.StopSequences
	}

	// Marshal request body
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/messages", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestJSON))
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

	return domain.Response{Content: responseContent}, nil
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

	// Parse the JSON into a map
	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
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

	// Convert domain messages to Anthropic messages format
	anthMessages := make([]map[string]interface{}, 0, len(messages))
	var systemMessage string

	for _, msg := range messages {
		if msg.Role == domain.RoleSystem {
			// Anthropic handles system message separately
			systemMessage = msg.Content
		} else {
			anthMessages = append(anthMessages, map[string]interface{}{
				"role":    string(msg.Role),
				"content": msg.Content,
			})
		}
	}

	// Prepare request body
	requestBody := map[string]interface{}{
		"model":    p.model,
		"messages": anthMessages,
		"stream":   true, // Important for streaming
	}

	// Add system message if present
	if systemMessage != "" {
		requestBody["system"] = systemMessage
	}

	// Add temperature if provided
	if providerOptions.Temperature != 0 {
		requestBody["temperature"] = providerOptions.Temperature
	}

	// Add max tokens if provided
	if providerOptions.MaxTokens != 0 {
		requestBody["max_tokens"] = providerOptions.MaxTokens
	}

	// Add top p if provided
	if providerOptions.TopP != 0 {
		requestBody["top_p"] = providerOptions.TopP
	}

	// Add stop sequences if provided
	if len(providerOptions.StopSequences) > 0 {
		requestBody["stop_sequences"] = providerOptions.StopSequences
	}

	// Marshal request body
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/messages", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestJSON))
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

	// Create a response stream
	tokenCh := make(chan domain.Token)

	// Start a goroutine to read the stream
	go func() {
		defer resp.Body.Close()
		defer close(tokenCh)

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
			if err := json.Unmarshal([]byte(data), &event); err != nil {
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
				if err := json.Unmarshal([]byte(data), &deltaEvent); err != nil {
					continue
				}

				if deltaEvent.Delta.Type == "text_delta" && deltaEvent.Delta.Text != "" {
					// Send the token
					select {
					case <-ctx.Done():
						return
					case tokenCh <- domain.Token{Text: deltaEvent.Delta.Text, Finished: false}:
						// Sent successfully
					}
				}
			case "message_delta":
				var stopEvent struct {
					Delta struct {
						StopReason string `json:"stop_reason"`
					} `json:"delta"`
				}
				if err := json.Unmarshal([]byte(data), &stopEvent); err != nil {
					continue
				}

				if stopEvent.Delta.StopReason != "" {
					// Send final token
					select {
					case <-ctx.Done():
						return
					case tokenCh <- domain.Token{Text: "", Finished: true}:
						return
					}
				}
			case "message_stop":
				// Send final token if not already sent
				select {
				case <-ctx.Done():
					return
				case tokenCh <- domain.Token{Text: "", Finished: true}:
					return
				}
			}
		}

		// If the scanner ends without a message_stop event, send a final token
		select {
		case <-ctx.Done():
			return
		case tokenCh <- domain.Token{Text: "", Finished: true}:
			// Sent successfully
		}
	}()

	return tokenCh, nil
}

// enhancePromptWithSchema is shared with OpenAI provider
// Consider extracting to a common utility package
func enhancePromptWithAnthropicSchema(prompt string, schema *schemaDomain.Schema) string {
	// Convert schema to JSON
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		// If we can't marshal the schema, just return the original prompt
		return prompt
	}

	// Build enhanced prompt
	enhancedPrompt := fmt.Sprintf(`%s

You are to provide a JSON response that conforms to the following JSON schema. 
Respond ONLY with valid JSON that matches this schema:

%s

Your response must be valid JSON only, with no explanations, markdown code blocks, or any other text.`, prompt, schemaJSON)

	return enhancedPrompt
}
