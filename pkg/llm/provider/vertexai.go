package provider

// File vertexai.go implements the Provider interface for Google Vertex AI service.
// It supports Gemini and partner models through the Vertex AI REST API, providing
// both standard and streaming generation capabilities with OAuth2 authentication
// using service accounts or Application Default Credentials (ADC).

// ABOUTME: Google Vertex AI provider implementation using REST API
// ABOUTME: Supports Gemini and partner models with OAuth2 authentication

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
	"github.com/lexlapax/go-llms/pkg/util/json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	defaultVertexAIScope = "https://www.googleapis.com/auth/cloud-platform"
)

// VertexAIProvider implements the Provider interface for Google Vertex AI
type VertexAIProvider struct {
	projectID          string
	location           string
	model              string
	httpClient         *http.Client
	tokenSource        oauth2.TokenSource
	serviceAccountPath string
	// Optimization: cache for converted messages
	messageCache *MessageCache
}

// NewVertexAIProvider creates a new Vertex AI provider instance.
// The projectID parameter specifies the Google Cloud project. The location parameter
// specifies the region (e.g., "us-central1"). The model parameter specifies which
// model to use (e.g., "gemini-1.0-pro", "claude-3-sonnet@20240229").
// Authentication is handled automatically using Application Default Credentials or
// service account credentials if specified via options.
func NewVertexAIProvider(projectID, location, model string, options ...domain.ProviderOption) (*VertexAIProvider, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required for Vertex AI provider")
	}
	if location == "" {
		return nil, fmt.Errorf("location is required for Vertex AI provider")
	}
	if model == "" {
		return nil, fmt.Errorf("model is required for Vertex AI provider")
	}

	provider := &VertexAIProvider{
		projectID:    projectID,
		location:     location,
		model:        model,
		httpClient:   &http.Client{Timeout: 300 * time.Second},
		messageCache: NewMessageCache(),
	}

	// Apply options
	for _, option := range options {
		// Check if the option is compatible with Vertex AI
		if vertexOption, ok := option.(domain.VertexAIOption); ok {
			vertexOption.ApplyToVertexAI(provider)
		}
	}

	// Initialize authentication
	if err := provider.initAuth(); err != nil {
		return nil, fmt.Errorf("failed to initialize Vertex AI authentication: %w", err)
	}

	return provider, nil
}

// initAuth initializes authentication for Vertex AI
func (p *VertexAIProvider) initAuth() error {
	ctx := context.Background()

	// Option 1: Service Account JSON file
	if p.serviceAccountPath != "" {
		keyData, err := os.ReadFile(p.serviceAccountPath)
		if err != nil {
			return fmt.Errorf("failed to read service account file: %w", err)
		}

		config, err := google.JWTConfigFromJSON(keyData, defaultVertexAIScope)
		if err != nil {
			return fmt.Errorf("failed to parse service account JSON: %w", err)
		}

		p.tokenSource = config.TokenSource(ctx)
		return nil
	}

	// Option 2: Check GOOGLE_APPLICATION_CREDENTIALS environment variable
	if credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		keyData, err := os.ReadFile(credPath)
		if err != nil {
			return fmt.Errorf("failed to read GOOGLE_APPLICATION_CREDENTIALS file: %w", err)
		}

		config, err := google.JWTConfigFromJSON(keyData, defaultVertexAIScope)
		if err != nil {
			return fmt.Errorf("failed to parse service account JSON: %w", err)
		}

		p.tokenSource = config.TokenSource(ctx)
		return nil
	}

	// Option 3: Application Default Credentials (ADC)
	credentials, err := google.FindDefaultCredentials(ctx, defaultVertexAIScope)
	if err != nil {
		return fmt.Errorf("failed to find default credentials: %w", err)
	}

	p.tokenSource = credentials.TokenSource
	return nil
}

// SetServiceAccountFile sets the service account file path
func (p *VertexAIProvider) SetServiceAccountFile(path string) {
	p.serviceAccountPath = path
}

// SetHTTPClient sets the HTTP client
func (p *VertexAIProvider) SetHTTPClient(client *http.Client) {
	p.httpClient = client
}

// buildGenerateURL builds the URL for the generateContent endpoint
func (p *VertexAIProvider) buildGenerateURL() string {
	return fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent",
		p.location, p.projectID, p.location, p.model,
	)
}

// buildStreamURL builds the URL for the streamGenerateContent endpoint
func (p *VertexAIProvider) buildStreamURL() string {
	return fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:streamGenerateContent",
		p.location, p.projectID, p.location, p.model,
	)
}

// Generate produces text from a prompt
func (p *VertexAIProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
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
func (p *VertexAIProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	// Apply options
	providerOptions := domain.DefaultOptions()
	for _, option := range options {
		option(providerOptions)
	}

	// Convert messages to Vertex AI format
	requestBody := p.convertToVertexFormat(messages, providerOptions)

	// Marshal request body
	requestBuffer := &bytes.Buffer{}
	err := json.MarshalWithBuffer(requestBody, requestBuffer)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.buildGenerateURL(), requestBuffer)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Apply authentication
	if err := p.applyAuth(req); err != nil {
		return domain.Response{}, fmt.Errorf("failed to apply authentication: %w", err)
	}

	// Make the request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.Response{}, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return domain.Response{}, ParseJSONError(body, resp.StatusCode, "vertexai", "GenerateMessage")
	}

	// Parse response
	var vertexResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &vertexResp); err != nil {
		return domain.Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if there are candidates
	if len(vertexResp.Candidates) == 0 {
		return domain.Response{}, fmt.Errorf("vertexai provider (%s): API returned no candidates in response", p.model)
	}

	// Extract text from the first candidate
	var responseText string
	if len(vertexResp.Candidates[0].Content.Parts) > 0 {
		responseText = vertexResp.Candidates[0].Content.Parts[0].Text
	}

	return domain.GetResponsePool().NewResponse(responseText), nil
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *VertexAIProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	// Build a prompt that includes the schema
	enhancedPrompt := enhancePromptWithSchema(prompt, schema)

	// Generate response
	response, err := p.Generate(ctx, enhancedPrompt, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Try to extract JSON from the response
	jsonStr := processor.ExtractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("response does not contain valid JSON")
	}

	// Parse the JSON
	var result interface{}
	if err := json.UnmarshalFromString(jsonStr, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return result, nil
}

// Stream streams responses token by token
func (p *VertexAIProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, prompt),
	}
	return p.StreamMessage(ctx, messages, options...)
}

// StreamMessage streams responses from a list of messages
func (p *VertexAIProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	// Apply options
	providerOptions := domain.DefaultOptions()
	for _, option := range options {
		option(providerOptions)
	}

	// Convert messages to Vertex AI format
	requestBody := p.convertToVertexFormat(messages, providerOptions)

	// Marshal request body
	requestBuffer := &bytes.Buffer{}
	err := json.MarshalWithBuffer(requestBody, requestBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.buildStreamURL(), requestBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// Apply authentication
	if err := p.applyAuth(req); err != nil {
		return nil, fmt.Errorf("failed to apply authentication: %w", err)
	}

	// Make the request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, ParseJSONError(body, resp.StatusCode, "vertexai", "StreamMessage")
	}

	// Get a channel from the pool
	responseStream, tokenCh := domain.GetChannelPool().GetResponseStream()

	// Start a goroutine to read the stream
	go func() {
		defer func() {
			_ = resp.Body.Close()
		}()
		defer close(tokenCh)

		p.parseSSEStream(resp.Body, tokenCh, ctx)
	}()

	return responseStream, nil
}

// applyAuth applies authentication to the request
func (p *VertexAIProvider) applyAuth(req *http.Request) error {
	token, err := p.tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	return nil
}

// convertToVertexFormat converts domain messages to Vertex AI format
func (p *VertexAIProvider) convertToVertexFormat(messages []domain.Message, options *domain.ProviderOptions) map[string]interface{} {
	// Check cache first
	cacheKey := GenerateMessagesKey(messages)
	if cachedRequest, found := p.messageCache.Get(cacheKey); found {
		// Add generation config to cached request
		if request, ok := cachedRequest.(map[string]interface{}); ok {
			request["generationConfig"] = p.buildGenerationConfig(options)
			return request
		}
	}

	contents := make([]map[string]interface{}, 0, len(messages))

	for _, msg := range messages {
		content := map[string]interface{}{
			"role":  p.mapRole(msg.Role),
			"parts": p.convertParts(msg),
		}
		contents = append(contents, content)
	}

	request := map[string]interface{}{
		"contents":         contents,
		"generationConfig": p.buildGenerationConfig(options),
	}

	// Cache the result (without generation config)
	cacheRequest := map[string]interface{}{
		"contents": contents,
	}
	p.messageCache.Set(cacheKey, cacheRequest)

	return request
}

// mapRole maps domain roles to Vertex AI roles
func (p *VertexAIProvider) mapRole(role domain.Role) string {
	switch role {
	case domain.RoleUser:
		return "user"
	case domain.RoleAssistant:
		return "model"
	case domain.RoleSystem:
		// Vertex AI doesn't have a system role, convert to user
		return "user"
	case domain.RoleTool:
		// Tool responses are part of the model's response in Vertex AI
		return "model"
	default:
		return "user"
	}
}

// convertParts converts message content to Vertex AI parts format
func (p *VertexAIProvider) convertParts(msg domain.Message) []map[string]interface{} {
	// If the message has structured content
	if len(msg.Content) > 0 {
		parts := make([]map[string]interface{}, 0, len(msg.Content))

		for _, content := range msg.Content {
			switch content.Type {
			case domain.ContentTypeText:
				parts = append(parts, map[string]interface{}{
					"text": content.Text,
				})
			case domain.ContentTypeImage:
				// Vertex AI expects inline image data
				if content.Image.Source.Type == domain.SourceTypeBase64 {
					parts = append(parts, map[string]interface{}{
						"inlineData": map[string]interface{}{
							"mimeType": content.Image.Source.MediaType,
							"data":     content.Image.Source.Data,
						},
					})
				} else if content.Image.Source.Type == domain.SourceTypeURL {
					// Vertex AI doesn't support image URLs directly
					// Would need to download and convert to base64
					// For now, we'll skip URL images
					continue
				}
			case domain.ContentTypeFile:
				// Vertex AI supports file uploads through a different API
				// For now, treat file content as text if available
				if content.File.FileData != "" {
					parts = append(parts, map[string]interface{}{
						"text": fmt.Sprintf("File: %s\nContent: %s", content.File.FileName, content.File.FileData),
					})
				}
				// Video and Audio are not directly supported in the same way
			}
		}

		return parts
	}

	// Legacy format - single text content
	// Extract text from first content part if available
	for _, content := range msg.Content {
		if content.Type == domain.ContentTypeText {
			return []map[string]interface{}{
				{"text": content.Text},
			}
		}
	}

	// No content found
	return []map[string]interface{}{}
}

// buildGenerationConfig builds the generation configuration
func (p *VertexAIProvider) buildGenerationConfig(options *domain.ProviderOptions) map[string]interface{} {
	config := make(map[string]interface{})

	if options.Temperature != 0.7 {
		config["temperature"] = options.Temperature
	}

	if options.MaxTokens != 1024 {
		config["maxOutputTokens"] = options.MaxTokens
	}

	if options.TopP != 1.0 {
		config["topP"] = options.TopP
	}

	if options.TopK != 0 {
		config["topK"] = options.TopK
	}

	if len(options.StopSequences) > 0 {
		config["stopSequences"] = options.StopSequences
	}

	return config
}

// parseSSEStream parses Server-Sent Events for streaming
func (p *VertexAIProvider) parseSSEStream(reader io.Reader, tokenCh chan<- domain.Token, ctx context.Context) {
	scanner := bufio.NewScanner(reader)
	var dataBuffer strings.Builder

	for scanner.Scan() {
		// Check if context is canceled
		select {
		case <-ctx.Done():
			return
		default:
			// Continue
		}

		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}
			dataBuffer.WriteString(data)
		} else if line == "" && dataBuffer.Len() > 0 {
			// Process accumulated data
			var response struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
				} `json:"candidates"`
			}

			if err := json.UnmarshalFromString(dataBuffer.String(), &response); err == nil {
				if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
					text := response.Candidates[0].Content.Parts[0].Text
					if text != "" {
						select {
						case <-ctx.Done():
							return
						case tokenCh <- domain.GetTokenPool().NewToken(text, false):
							// Sent successfully
						}
					}
				}
			}

			dataBuffer.Reset()
		}
	}
}
