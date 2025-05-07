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
)

const (
	defaultBaseURL = "https://api.openai.com"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// OpenAIOption configures the OpenAI provider
type OpenAIOption func(*OpenAIProvider)

// WithBaseURL sets a custom base URL for the OpenAI API
func WithBaseURL(url string) OpenAIOption {
	return func(p *OpenAIProvider) {
		p.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) OpenAIOption {
	return func(p *OpenAIProvider) {
		p.httpClient = client
	}
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string, options ...OpenAIOption) *OpenAIProvider {
	provider := &OpenAIProvider{
		apiKey:     apiKey,
		model:      model,
		baseURL:    defaultBaseURL,
		httpClient: http.DefaultClient,
	}

	for _, option := range options {
		option(provider)
	}

	return provider
}

// Generate produces text from a prompt
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
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
func (p *OpenAIProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	// Apply options
	providerOptions := domain.DefaultOptions()
	for _, option := range options {
		option(providerOptions)
	}

	// Convert domain messages to OpenAI messages
	oaiMessages := make([]map[string]interface{}, 0, len(messages))
	var lastAssistantIdx int = -1

	for i, msg := range messages {
		if msg.Role == domain.RoleAssistant {
			lastAssistantIdx = i
		}
	}

	for i, msg := range messages {
		// Special handling for tool messages - they must follow an assistant message with tool_calls
		if msg.Role == domain.RoleTool {
			// If this is a tool message without a preceding assistant message with tool_calls,
			// we need to convert it to a different format or skip it
			if lastAssistantIdx == -1 || i == 0 || messages[i-1].Role != domain.RoleAssistant {
				// Convert to user message instead as a fallback
				oaiMessages = append(oaiMessages, map[string]interface{}{
					"role":    string(domain.RoleUser),
					"content": fmt.Sprintf("Tool result: %s", msg.Content),
				})
			} else {
				// This is a valid tool message following an assistant
				oaiMessages = append(oaiMessages, map[string]interface{}{
					"role":         string(msg.Role),
					"content":      msg.Content,
					"tool_call_id": fmt.Sprintf("call_%d", i), // Generate a tool call ID
				})
			}
		} else if msg.Role == domain.RoleAssistant && i < len(messages)-1 && messages[i+1].Role == domain.RoleTool {
			// This assistant message is followed by a tool message, add tool_calls
			oaiMessages = append(oaiMessages, map[string]interface{}{
				"role":    string(msg.Role),
				"content": msg.Content,
				"tool_calls": []map[string]interface{}{
					{
						"id":   fmt.Sprintf("call_%d", i+1),
						"type": "function",
						"function": map[string]interface{}{
							"name":      "generic_tool",
							"arguments": "{}",
						},
					},
				},
			})
		} else {
			// Regular message
			oaiMessages = append(oaiMessages, map[string]interface{}{
				"role":    string(msg.Role),
				"content": msg.Content,
			})
		}
	}

	// Prepare request body
	requestBody := map[string]interface{}{
		"model":       p.model,
		"messages":    oaiMessages,
		"temperature": providerOptions.Temperature,
		"max_tokens":  providerOptions.MaxTokens,
		"top_p":       providerOptions.TopP,
	}

	if len(providerOptions.StopSequences) > 0 {
		requestBody["stop"] = providerOptions.StopSequences
	}

	if providerOptions.FrequencyPenalty != 0 {
		requestBody["frequency_penalty"] = providerOptions.FrequencyPenalty
	}

	if providerOptions.PresencePenalty != 0 {
		requestBody["presence_penalty"] = providerOptions.PresencePenalty
	}

	// Marshal request body
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/chat/completions", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestJSON))
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

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
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return domain.Response{}, fmt.Errorf("API error: %s", errorResp.Error.Message)
		}
		return domain.Response{}, fmt.Errorf("API error: status code %d", resp.StatusCode)
	}

	// Parse response
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return domain.Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if there are choices
	if len(openAIResp.Choices) == 0 {
		return domain.Response{}, fmt.Errorf("API returned no choices")
	}

	return domain.Response{Content: openAIResp.Choices[0].Message.Content}, nil
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

	// Try to extract JSON from the response
	jsonStr := extractJSON(response)
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
func (p *OpenAIProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	messages := []domain.Message{
		{Role: domain.RoleUser, Content: prompt},
	}
	return p.StreamMessage(ctx, messages, options...)
}

// StreamMessage streams responses from a list of messages
func (p *OpenAIProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	// Apply options
	providerOptions := domain.DefaultOptions()
	for _, option := range options {
		option(providerOptions)
	}

	// Convert domain messages to OpenAI messages
	oaiMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		oaiMessages[i] = map[string]string{
			"role":    string(msg.Role),
			"content": msg.Content,
		}
	}

	// Prepare request body
	requestBody := map[string]interface{}{
		"model":       p.model,
		"messages":    oaiMessages,
		"temperature": providerOptions.Temperature,
		"max_tokens":  providerOptions.MaxTokens,
		"top_p":       providerOptions.TopP,
		"stream":      true, // Important for streaming
	}

	if len(providerOptions.StopSequences) > 0 {
		requestBody["stop"] = providerOptions.StopSequences
	}

	if providerOptions.FrequencyPenalty != 0 {
		requestBody["frequency_penalty"] = providerOptions.FrequencyPenalty
	}

	if providerOptions.PresencePenalty != 0 {
		requestBody["presence_penalty"] = providerOptions.PresencePenalty
	}

	// Marshal request body
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/chat/completions", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
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

			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
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
				case tokenCh <- domain.Token{Text: "", Finished: true}:
					return
				}
			}

			// Skip empty content
			if content == "" {
				continue
			}

			// Send the token
			select {
			case <-ctx.Done():
				return
			case tokenCh <- domain.Token{Text: content, Finished: false}:
				// Sent successfully
			}
		}
	}()

	return tokenCh, nil
}

// enhancePromptWithSchema adds schema information to a prompt
func enhancePromptWithSchema(prompt string, schema *schemaDomain.Schema) string {
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

Output only valid JSON without any explanations, markdown code blocks, or any other text.`, prompt, schemaJSON)

	return enhancedPrompt
}

// extractJSON attempts to find and extract JSON from a string
func extractJSON(s string) string {
	// Look for JSON object between curly braces
	startIdx := strings.Index(s, "{")
	endIdx := strings.LastIndex(s, "}")

	if startIdx >= 0 && endIdx > startIdx {
		return s[startIdx : endIdx+1]
	}

	// Look for JSON array between square brackets
	startIdx = strings.Index(s, "[")
	endIdx = strings.LastIndex(s, "]")

	if startIdx >= 0 && endIdx > startIdx {
		return s[startIdx : endIdx+1]
	}

	// No JSON found
	return ""
}
