package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"golang.org/x/oauth2"
)

// mockTokenSource implements oauth2.TokenSource for testing
type mockTokenSource struct {
	token string
}

func (m *mockTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: m.token,
	}, nil
}

// testTransport wraps http.RoundTripper to rewrite URLs for testing
type testTransport struct {
	base      http.RoundTripper
	serverURL string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to use test server
	testReq := req.Clone(req.Context())
	testReq.URL.Scheme = "http"
	testReq.URL.Host = strings.TrimPrefix(t.serverURL, "http://")

	if t.base != nil {
		return t.base.RoundTrip(testReq)
	}
	return http.DefaultTransport.RoundTrip(testReq)
}

// createTestVertexAIProvider creates a provider configured for testing
func createTestVertexAIProvider(serverURL string) *VertexAIProvider {
	return &VertexAIProvider{
		projectID:    "test-project",
		location:     "us-central1",
		model:        "gemini-1.5-pro",
		httpClient:   &http.Client{},
		tokenSource:  &mockTokenSource{token: "test-token"},
		messageCache: NewMessageCache(),
	}
}

func TestVertexAIProvider_buildURLs(t *testing.T) {
	provider := &VertexAIProvider{
		projectID: "test-project",
		location:  "us-central1",
		model:     "gemini-1.5-pro",
	}

	expectedGenerateURL := "https://us-central1-aiplatform.googleapis.com/v1/projects/test-project/locations/us-central1/publishers/google/models/gemini-1.5-pro:generateContent"
	if url := provider.buildGenerateURL(); url != expectedGenerateURL {
		t.Errorf("Expected generate URL %s, got %s", expectedGenerateURL, url)
	}

	expectedStreamURL := "https://us-central1-aiplatform.googleapis.com/v1/projects/test-project/locations/us-central1/publishers/google/models/gemini-1.5-pro:streamGenerateContent"
	if url := provider.buildStreamURL(); url != expectedStreamURL {
		t.Errorf("Expected stream URL %s, got %s", expectedStreamURL, url)
	}
}

func TestVertexAIProvider_mapRole(t *testing.T) {
	provider := &VertexAIProvider{}

	tests := []struct {
		input    domain.Role
		expected string
	}{
		{domain.RoleUser, "user"},
		{domain.RoleAssistant, "model"},
		{domain.RoleSystem, "user"}, // Vertex AI doesn't have system role
		{domain.RoleTool, "model"},  // Tool responses are part of model's response
	}

	for _, test := range tests {
		result := provider.mapRole(test.input)
		if result != test.expected {
			t.Errorf("mapRole(%s) = %s, want %s", test.input, result, test.expected)
		}
	}
}

func TestVertexAIProvider_convertParts(t *testing.T) {
	provider := &VertexAIProvider{}

	// Test text message using NewTextMessage
	textMsg := domain.NewTextMessage(domain.RoleUser, "Hello, world!")

	parts := provider.convertParts(textMsg)
	if len(parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(parts))
	}
	if parts[0]["text"] != "Hello, world!" {
		t.Errorf("Expected text 'Hello, world!', got %v", parts[0]["text"])
	}

	// Test image message using NewImageMessage
	imageData := []byte("test image data")
	imageMsg := domain.NewImageMessage(domain.RoleUser, imageData, "image/png", "")

	parts = provider.convertParts(imageMsg)
	if len(parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(parts))
	}
	inlineData, ok := parts[0]["inlineData"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected inlineData to be a map")
	}
	if inlineData["mimeType"] != "image/png" {
		t.Errorf("Expected mimeType 'image/png', got %v", inlineData["mimeType"])
	}
	// The data should be base64 encoded by NewImageMessage
	expectedData := "dGVzdCBpbWFnZSBkYXRh" // base64 of "test image data"
	if inlineData["data"] != expectedData {
		t.Errorf("Expected data '%s', got %v", expectedData, inlineData["data"])
	}
}

func TestVertexAIProvider_buildGenerationConfig(t *testing.T) {
	provider := &VertexAIProvider{}

	options := &domain.ProviderOptions{
		Temperature:   0.8,
		MaxTokens:     2048,
		TopP:          0.95,
		TopK:          40,
		StopSequences: []string{"END", "STOP"},
	}

	config := provider.buildGenerationConfig(options)

	if config["temperature"] != 0.8 {
		t.Errorf("Expected temperature 0.8, got %v", config["temperature"])
	}
	if config["maxOutputTokens"] != 2048 {
		t.Errorf("Expected maxOutputTokens 2048, got %v", config["maxOutputTokens"])
	}
	if config["topP"] != 0.95 {
		t.Errorf("Expected topP 0.95, got %v", config["topP"])
	}
	if config["topK"] != 40 {
		t.Errorf("Expected topK 40, got %v", config["topK"])
	}

	stopSeqs, ok := config["stopSequences"].([]string)
	if !ok || len(stopSeqs) != 2 {
		t.Errorf("Expected stopSequences with 2 items, got %v", config["stopSequences"])
	}
}

func TestVertexAIProvider_GenerateMessage(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got %s", authHeader)
		}

		// Check URL path
		expectedPath := "/v1/projects/test-project/locations/us-central1/publishers/google/models/gemini-1.5-pro:generateContent"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Parse request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		// Check contents
		contents, ok := requestBody["contents"].([]interface{})
		if !ok || len(contents) != 1 {
			t.Errorf("Expected contents array with 1 item, got %v", requestBody["contents"])
		}

		// Return mock response
		response := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{"text": "Hello from Vertex AI!"},
						},
					},
					"finishReason": "STOP",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider for testing
	provider := createTestVertexAIProvider(server.URL)
	provider.httpClient = server.Client()

	// Create a custom HTTP client that rewrites the URL
	originalClient := provider.httpClient
	provider.httpClient = &http.Client{
		Transport: &testTransport{
			base:      originalClient.Transport,
			serverURL: server.URL,
		},
	}

	// Test message generation
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "Hello"),
	}

	response, err := provider.GenerateMessage(context.Background(), messages)
	if err != nil {
		t.Fatalf("GenerateMessage failed: %v", err)
	}

	if response.Content != "Hello from Vertex AI!" {
		t.Errorf("Expected response 'Hello from Vertex AI!', got %s", response.Content)
	}
}

func TestVertexAIProvider_ErrorHandling(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := map[string]interface{}{
			"error": map[string]interface{}{
				"code":    400,
				"message": "Invalid request",
				"status":  "INVALID_ARGUMENT",
			},
		}
		_ = json.NewEncoder(w).Encode(errorResponse)
	}))
	defer server.Close()

	// Create provider for testing
	provider := createTestVertexAIProvider(server.URL)
	provider.httpClient = &http.Client{
		Transport: &testTransport{
			base:      server.Client().Transport,
			serverURL: server.URL,
		},
	}

	// Test error handling
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "Hello"),
	}

	_, err := provider.GenerateMessage(context.Background(), messages)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Check if it's a ProviderError
	providerErr, ok := err.(*domain.ProviderError)
	if !ok {
		t.Fatalf("Expected ProviderError, got %T: %v", err, err)
	}

	// Check the provider name
	if providerErr.Provider != "vertexai" {
		t.Errorf("Expected provider 'vertexai', got %s", providerErr.Provider)
	}

	// Check status code
	if providerErr.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, providerErr.StatusCode)
	}

	// The error message may have been mapped
	if !strings.Contains(strings.ToLower(err.Error()), "invalid") {
		t.Errorf("Expected error to contain 'invalid', got %v", err)
	}
}

func TestVertexAIProvider_StreamMessage(t *testing.T) {
	// Create a test server that returns SSE stream
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("Expected Accept header 'text/event-stream', got %s", r.Header.Get("Accept"))
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Send SSE data
		flusher, _ := w.(http.Flusher)

		// First chunk
		response1 := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{"text": "Hello "},
						},
					},
				},
			},
		}
		data1, _ := json.Marshal(response1)
		_, _ = w.Write([]byte("data: " + string(data1) + "\n\n"))
		flusher.Flush()

		// Second chunk
		response2 := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{"text": "from streaming!"},
						},
					},
				},
			},
		}
		data2, _ := json.Marshal(response2)
		_, _ = w.Write([]byte("data: " + string(data2) + "\n\n"))
		flusher.Flush()

		// End signal
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	// Create provider for testing
	provider := createTestVertexAIProvider(server.URL)
	provider.httpClient = &http.Client{
		Transport: &testTransport{
			base:      server.Client().Transport,
			serverURL: server.URL,
		},
	}

	// Test streaming
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "Hello"),
	}

	stream, err := provider.StreamMessage(context.Background(), messages)
	if err != nil {
		t.Fatalf("StreamMessage failed: %v", err)
	}

	// Collect tokens
	var result strings.Builder
	for token := range stream {
		result.WriteString(token.Text)
	}

	expectedResult := "Hello from streaming!"
	if result.String() != expectedResult {
		t.Errorf("Expected streaming result '%s', got '%s'", expectedResult, result.String())
	}
}
