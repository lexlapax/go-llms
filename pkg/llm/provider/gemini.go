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
	defaultGeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"
)

// GeminiProvider implements the Provider interface for Google Gemini
type GeminiProvider struct {
	apiKey         string
	model          string
	baseURL        string
	httpClient     *http.Client
	messageCache   *MessageCache
	topK           int
	safetySettings []map[string]interface{}
}

// NewGeminiProvider creates a new Google Gemini provider
// Default model is "gemini-2.0-flash-lite"
func NewGeminiProvider(apiKey, model string, options ...domain.ProviderOption) *GeminiProvider {
	// Default to Gemini 2.0 Flash Lite if no model is specified
	if model == "" {
		model = "gemini-2.0-flash-lite"
	}

	provider := &GeminiProvider{
		apiKey:         apiKey,
		model:          model,
		baseURL:        defaultGeminiBaseURL,
		httpClient:     http.DefaultClient,
		messageCache:   NewMessageCache(),
		topK:           40, // Default topK value
		safetySettings: nil,
	}

	for _, option := range options {
		// Check if the option is compatible with Gemini
		if geminiOption, ok := option.(domain.GeminiOption); ok {
			geminiOption.ApplyToGemini(provider)
		}
	}

	return provider
}

// Setter methods for options
// SetBaseURL sets the base URL for the Gemini API
func (p *GeminiProvider) SetBaseURL(url string) {
	p.baseURL = url
}

// SetHTTPClient sets the HTTP client
func (p *GeminiProvider) SetHTTPClient(client *http.Client) {
	p.httpClient = client
}

// SetTopK sets the topK parameter for Gemini API calls
func (p *GeminiProvider) SetTopK(topK int) {
	p.topK = topK
}

// SetSafetySettings sets the safety settings for Gemini API calls
func (p *GeminiProvider) SetSafetySettings(settings []map[string]interface{}) {
	p.safetySettings = settings
}

// ConvertMessagesToGeminiFormat converts domain messages to Gemini API format
func (p *GeminiProvider) ConvertMessagesToGeminiFormat(messages []domain.Message) []map[string]interface{} {
	// Check cache first
	cacheKey := GenerateMessagesKey(messages)
	if cachedResult, found := p.messageCache.Get(cacheKey); found {
		return cachedResult.([]map[string]interface{})
	}

	// For Gemini, we simplify compared to other providers since their format is more straightforward
	// Create initial contents array with typical capacity
	contents := make([]map[string]interface{}, 0, len(messages))

	for _, msg := range messages {
		// Create a message object
		message := make(map[string]interface{})

		// Map our roles to Gemini roles
		// Gemini API uses "user" and "model" roles
		var role string
		switch msg.Role {
		case domain.RoleUser:
			role = "user"
		case domain.RoleAssistant:
			role = "model"
		case domain.RoleSystem:
			// System messages are special in Gemini
			// For now, we'll treat them as user messages with a prefix
			// This is a simplification - proper handling will be implemented later
			role = "user"
		default:
			role = "user" // Default to user for unknown roles
		}

		message["role"] = role

		// Handle multimodal content
		if msg.Content != nil && len(msg.Content) > 0 {
			parts := make([]map[string]interface{}, 0, len(msg.Content))
			
			for _, part := range msg.Content {
				switch part.Type {
				case domain.ContentTypeText:
					// Text part
					parts = append(parts, map[string]interface{}{
						"text": part.Text,
					})
				case domain.ContentTypeImage:
					// Image part
					if part.Image.Source.Type == domain.SourceTypeURL {
						// URL-based image
						parts = append(parts, map[string]interface{}{
							"inline_data": map[string]interface{}{
								"mime_type": part.Image.Source.MediaType,
								"url": part.Image.Source.URL,
							},
						})
					} else {
						// Base64-encoded image
						parts = append(parts, map[string]interface{}{
							"inline_data": map[string]interface{}{
								"mime_type": part.Image.Source.MediaType,
								"data": part.Image.Source.Data,
							},
						})
					}
				case domain.ContentTypeVideo:
					// Video part
					if part.Video.Source.Type == domain.SourceTypeURL {
						// URL-based video
						parts = append(parts, map[string]interface{}{
							"inline_data": map[string]interface{}{
								"mime_type": part.Video.Source.MediaType,
								"url": part.Video.Source.URL,
							},
						})
					} else {
						// Base64-encoded video
						parts = append(parts, map[string]interface{}{
							"inline_data": map[string]interface{}{
								"mime_type": part.Video.Source.MediaType,
								"data": part.Video.Source.Data,
							},
						})
					}
				}
			}
			
			message["parts"] = parts
		} else {
			// Legacy compatibility for old message format
			message["parts"] = []map[string]interface{}{
				{
					"text": "",
				},
			}
		}
		
		contents = append(contents, message)
	}

	// Cache the result
	p.messageCache.Set(cacheKey, contents)

	return contents
}

// buildGeminiRequestBody creates a request body for the Gemini API
func (p *GeminiProvider) buildGeminiRequestBody(
	contents []map[string]interface{},
	options *domain.ProviderOptions,
) map[string]interface{} {
	// Pre-allocate the map with the right capacity for efficiency
	requestBody := make(map[string]interface{}, 3)

	// Add contents to the request body
	requestBody["contents"] = contents

	// Add generation config with various parameters
	generationConfig := make(map[string]interface{}, 5)

	if options.Temperature != 0.7 {
		generationConfig["temperature"] = options.Temperature
	}

	if options.MaxTokens != 1024 {
		generationConfig["maxOutputTokens"] = options.MaxTokens
	}

	if options.TopP != 1.0 {
		generationConfig["topP"] = options.TopP
	}

	// Add top K if it's set
	generationConfig["topK"] = p.topK

	// Add stop sequences if provided
	if len(options.StopSequences) > 0 {
		generationConfig["stopSequences"] = options.StopSequences
	}

	// Only add the generationConfig if it has entries
	if len(generationConfig) > 0 {
		requestBody["generationConfig"] = generationConfig
	}

	// Add safety settings if configured
	if len(p.safetySettings) > 0 {
		requestBody["safetySettings"] = p.safetySettings
	}

	return requestBody
}

// validateContentTypesForGemini checks if the content types in the messages are supported by Gemini
func (p *GeminiProvider) validateContentTypesForGemini(messages []domain.Message) error {
	for _, msg := range messages {
		if msg.Content != nil {
			for _, part := range msg.Content {
				// Gemini currently supports text, image, and video content types
				if part.Type != domain.ContentTypeText && 
				   part.Type != domain.ContentTypeImage && 
				   part.Type != domain.ContentTypeVideo {
					return domain.NewUnsupportedContentTypeError("Gemini", part.Type)
				}
			}
		}
	}
	return nil
}

// Generate produces text from a prompt
func (p *GeminiProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
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

// GenerateMessage produces text from a list of messages
func (p *GeminiProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	// Validate content types
	if err := p.validateContentTypesForGemini(messages); err != nil {
		return domain.Response{}, err
	}

	// Apply options
	providerOptions := domain.DefaultOptions()
	for _, option := range options {
		option(providerOptions)
	}

	// Convert messages to Gemini format - with caching
	geminiContents := p.ConvertMessagesToGeminiFormat(messages)

	// Build request body
	requestBody := p.buildGeminiRequestBody(geminiContents, providerOptions)

	// Use optimized JSON marshaling with buffer reuse for request body
	requestBuffer := &bytes.Buffer{}
	err := json.MarshalWithBuffer(requestBody, requestBuffer)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request with API key in URL
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, requestBuffer)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

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

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		// Extract error information
		var errorResponse struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return domain.Response{}, mapGeminiErrorToStandard(
				resp.StatusCode,
				errorResponse.Error.Status,
				errorResponse.Error.Message,
				"GenerateMessage",
			)
		}
		return domain.Response{}, ParseJSONError(body, resp.StatusCode, "gemini", "GenerateMessage")
	}

	// Parse response
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		PromptFeedback struct {
			BlockReason   string `json:"blockReason"`
			SafetyRatings []struct {
				Category    string `json:"category"`
				Probability string `json:"probability"`
			} `json:"safetyRatings"`
		} `json:"promptFeedback"`
	}

	// Use optimized unmarshaling
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return domain.Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for blocked content
	if geminiResp.PromptFeedback.BlockReason != "" {
		return domain.Response{}, domain.NewProviderError(
			"gemini",
			"GenerateMessage",
			resp.StatusCode,
			fmt.Sprintf("content blocked: %s", geminiResp.PromptFeedback.BlockReason),
			domain.ErrContentFiltered,
		)
	}

	// Check if we have candidates
	if len(geminiResp.Candidates) == 0 {
		return domain.Response{}, fmt.Errorf("no candidates in response")
	}

	// Extract text from response - combine all parts
	var responseBuilder strings.Builder
	for _, part := range geminiResp.Candidates[0].Content.Parts {
		responseBuilder.WriteString(part.Text)
	}
	responseContent := responseBuilder.String()

	// Use the response pool to reduce allocations
	return domain.GetResponsePool().NewResponse(responseContent), nil
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *GeminiProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	// Build a prompt that includes the schema
	enhancedPrompt := enhancePromptWithGeminiSchema(prompt, schema)

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
func (p *GeminiProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	// Create a simple text message using the new structure
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, prompt),
	}
	return p.StreamMessage(ctx, messages, options...)
}

// StreamMessage streams responses from a list of messages
func (p *GeminiProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	// Validate content types
	if err := p.validateContentTypesForGemini(messages); err != nil {
		return nil, err
	}

	// Apply options
	providerOptions := domain.DefaultOptions()
	for _, option := range options {
		option(providerOptions)
	}

	// Convert messages to Gemini format - with caching
	geminiContents := p.ConvertMessagesToGeminiFormat(messages)

	// Build request body
	requestBody := p.buildGeminiRequestBody(geminiContents, providerOptions)

	// Use optimized JSON marshaling with buffer reuse for request body
	requestBuffer := &bytes.Buffer{}
	err := json.MarshalWithBuffer(requestBody, requestBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request with API key in URL - use streamGenerateContent endpoint for streaming
	// IMPORTANT: The alt=sse parameter is REQUIRED for the Gemini API to return responses in Server-Sent Events format.
	// Without this parameter, the API returns standard JSON responses that don't conform to SSE protocol,
	// causing the streaming implementation to fail. This requirement is verified in TestGeminiAltSSEParameter.
	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s", p.baseURL, p.model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, requestBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	// Add Accept header for SSE (Server-Sent Events)
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

		// Extract error information
		var errorResponse struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return nil, mapGeminiErrorToStandard(
				resp.StatusCode,
				errorResponse.Error.Status,
				errorResponse.Error.Message,
				"StreamMessage",
			)
		}
		return nil, ParseJSONError(body, resp.StatusCode, "gemini", "StreamMessage")
	}

	// Get a channel from the pool
	responseStream, tokenCh := domain.GetChannelPool().GetResponseStream()

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
			if data == "" {
				continue
			}

			// Handle special case for end of stream marker if present
			if data == "[DONE]" {
				select {
				case <-ctx.Done():
					return
				case tokenCh <- domain.GetTokenPool().NewToken("", true):
					// Sent finish token
				}
				return
			}

			// Parse the data as JSON
			var streamResponse struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
					FinishReason string `json:"finishReason"`
				} `json:"candidates"`
			}

			// Use optimized JSON unmarshaling
			if err := json.UnmarshalFromString(data, &streamResponse); err != nil {
				continue
			}

			// Check if we have candidates
			if len(streamResponse.Candidates) == 0 {
				continue
			}

			// Extract text from response - combine all parts
			var text string
			for _, part := range streamResponse.Candidates[0].Content.Parts {
				text += part.Text
			}

			// Skip empty text
			if text == "" {
				continue
			}

			// Check if this is the final message with a finish reason
			isFinished := streamResponse.Candidates[0].FinishReason != ""

			// Send the token - use token pool to reduce allocations
			select {
			case <-ctx.Done():
				return
			case tokenCh <- domain.GetTokenPool().NewToken(text, isFinished):
				// Sent successfully
			}

			// If this was the final message, we're done
			if isFinished {
				return
			}
		}

		// If the scanner ends without a finish reason, send a final token
		select {
		case <-ctx.Done():
			return
		case tokenCh <- domain.GetTokenPool().NewToken("", true):
			// Sent successfully
		}
	}()

	return responseStream, nil
}

// mapGeminiErrorToStandard maps Gemini API error messages to standard error types
func mapGeminiErrorToStandard(statusCode int, errorType, errorMsg string, operation string) error {
	// Convert error message and type to lowercase for case-insensitive matching
	lowerErrorMsg := strings.ToLower(errorMsg)
	lowerErrorType := strings.ToLower(errorType)

	// Common error patterns for Gemini
	switch {
	case statusCode == http.StatusUnauthorized ||
		strings.Contains(lowerErrorType, "authentication") ||
		strings.Contains(lowerErrorMsg, "api key") ||
		strings.Contains(lowerErrorMsg, "unauthorized"):
		return domain.NewProviderError("gemini", operation, statusCode, errorMsg, domain.ErrAuthenticationFailed)

	case statusCode == http.StatusTooManyRequests ||
		strings.Contains(lowerErrorType, "rate_limit") ||
		strings.Contains(lowerErrorMsg, "rate limit"):
		return domain.NewProviderError("gemini", operation, statusCode, errorMsg, domain.ErrRateLimitExceeded)

	case strings.Contains(lowerErrorType, "invalid_argument") ||
		strings.Contains(lowerErrorMsg, "token") ||
		strings.Contains(lowerErrorMsg, "context length") ||
		strings.Contains(lowerErrorMsg, "too long"):
		return domain.NewProviderError("gemini", operation, statusCode, errorMsg, domain.ErrContextTooLong)

	case strings.Contains(lowerErrorType, "permissions_denied") ||
		strings.Contains(lowerErrorMsg, "content filtered") ||
		strings.Contains(lowerErrorMsg, "content policy") ||
		strings.Contains(lowerErrorMsg, "safety"):
		return domain.NewProviderError("gemini", operation, statusCode, errorMsg, domain.ErrContentFiltered)

	case strings.Contains(lowerErrorType, "not_found") ||
		strings.Contains(lowerErrorMsg, "model not found"):
		return domain.NewProviderError("gemini", operation, statusCode, errorMsg, domain.ErrModelNotFound)

	case strings.Contains(lowerErrorType, "resource_exhausted") ||
		strings.Contains(lowerErrorMsg, "quota") ||
		strings.Contains(lowerErrorMsg, "billing") ||
		strings.Contains(lowerErrorMsg, "payment"):
		return domain.NewProviderError("gemini", operation, statusCode, errorMsg, domain.ErrTokenQuotaExceeded)

	case strings.Contains(lowerErrorType, "invalid") ||
		strings.Contains(lowerErrorMsg, "invalid parameter") ||
		strings.Contains(lowerErrorMsg, "invalid request"):
		return domain.NewProviderError("gemini", operation, statusCode, errorMsg, domain.ErrInvalidModelParameters)

	case statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusGatewayTimeout ||
		strings.Contains(lowerErrorMsg, "network") ||
		strings.Contains(lowerErrorMsg, "connection") ||
		strings.Contains(lowerErrorType, "connection"):
		return domain.NewProviderError("gemini", operation, statusCode, errorMsg, domain.ErrNetworkConnectivity)

	case statusCode >= 500:
		return domain.NewProviderError("gemini", operation, statusCode, errorMsg, domain.ErrProviderUnavailable)

	default:
		return domain.NewProviderError("gemini", operation, statusCode, errorMsg, domain.ErrRequestFailed)
	}
}

// enhancePromptWithGeminiSchema adds schema information to the prompt for structured output
func enhancePromptWithGeminiSchema(prompt string, schema *schemaDomain.Schema) string {
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
