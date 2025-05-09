package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestNewGeminiProvider(t *testing.T) {
	// Test creation with minimal parameters
	provider := NewGeminiProvider("test-api-key", "")
	if provider.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got %s", provider.apiKey)
	}
	if provider.model != "gemini-2.0-flash-lite" {
		t.Errorf("Expected default model to be 'gemini-2.0-flash-lite', got %s", provider.model)
	}
	if provider.baseURL != defaultGeminiBaseURL {
		t.Errorf("Expected default baseURL, got %s", provider.baseURL)
	}
	if provider.topK != 40 {
		t.Errorf("Expected default topK to be 40, got %d", provider.topK)
	}
}

func TestConvertMessagesToGeminiFormat(t *testing.T) {
	provider := NewGeminiProvider("test-api-key", "")

	// Test with single user message
	messages := []domain.Message{
		{Role: domain.RoleUser, Content: "Hello, world!"},
	}

	geminiFormat := provider.ConvertMessagesToGeminiFormat(messages)

	if len(geminiFormat) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(geminiFormat))
	}

	if geminiFormat[0]["role"] != "user" {
		t.Errorf("Expected role 'user', got %s", geminiFormat[0]["role"])
	}

	parts, ok := geminiFormat[0]["parts"].([]map[string]interface{})
	if !ok {
		t.Fatalf("Expected parts to be []map[string]interface{}")
	}

	if len(parts) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(parts))
	}

	if parts[0]["text"] != "Hello, world!" {
		t.Errorf("Expected text 'Hello, world!', got %s", parts[0]["text"])
	}

	// Test with conversation
	messages = []domain.Message{
		{Role: domain.RoleUser, Content: "Hello"},
		{Role: domain.RoleAssistant, Content: "Hi there!"},
		{Role: domain.RoleUser, Content: "How are you?"},
	}

	geminiFormat = provider.ConvertMessagesToGeminiFormat(messages)

	if len(geminiFormat) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(geminiFormat))
	}

	// Check first message
	if geminiFormat[0]["role"] != "user" {
		t.Errorf("Expected role 'user', got %s", geminiFormat[0]["role"])
	}

	// Check second message
	if geminiFormat[1]["role"] != "model" {
		t.Errorf("Expected role 'model', got %s", geminiFormat[1]["role"])
	}

	// Check third message
	if geminiFormat[2]["role"] != "user" {
		t.Errorf("Expected role 'user', got %s", geminiFormat[2]["role"])
	}

	// Test with system message
	messages = []domain.Message{
		{Role: domain.RoleSystem, Content: "You are a helpful assistant"},
		{Role: domain.RoleUser, Content: "Hello"},
	}

	geminiFormat = provider.ConvertMessagesToGeminiFormat(messages)

	if len(geminiFormat) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(geminiFormat))
	}

	// System message should be converted to user message
	if geminiFormat[0]["role"] != "user" {
		t.Errorf("Expected system message to be converted to role 'user', got %s", geminiFormat[0]["role"])
	}

	// Check caching
	cachedGeminiFormat := provider.ConvertMessagesToGeminiFormat(messages)
	if len(cachedGeminiFormat) != len(geminiFormat) {
		t.Errorf("Cache returned different result")
	}
}

func TestBuildGeminiRequestBody(t *testing.T) {
	provider := NewGeminiProvider("test-api-key", "")

	// Create simple contents
	contents := []map[string]interface{}{
		{
			"role": "user",
			"parts": []map[string]interface{}{
				{"text": "Hello"},
			},
		},
	}

	// Test with default options
	options := domain.DefaultOptions()
	body := provider.buildGeminiRequestBody(contents, options)

	// Verify contents is set
	bodyContents, ok := body["contents"]
	if !ok {
		t.Errorf("Expected contents to be present in the request body")
	}
	
	// Check if it's the same type
	_, ok = bodyContents.([]map[string]interface{})
	if !ok {
		t.Errorf("Expected contents to be a slice of maps")
	}

	// Test with custom options
	options.Temperature = 0.2
	options.MaxTokens = 500
	options.TopP = 0.9
	options.StopSequences = []string{"END"}

	body = provider.buildGeminiRequestBody(contents, options)

	// Verify generation config is set
	config, ok := body["generationConfig"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected generationConfig to be a map")
	}

	if config["temperature"] != 0.2 {
		t.Errorf("Expected temperature 0.2, got %v", config["temperature"])
	}

	if config["maxOutputTokens"] != 500 {
		t.Errorf("Expected maxOutputTokens 500, got %v", config["maxOutputTokens"])
	}

	if config["topP"] != 0.9 {
		t.Errorf("Expected topP 0.9, got %v", config["topP"])
	}

	stopSeq, ok := config["stopSequences"].([]string)
	if !ok || len(stopSeq) != 1 || stopSeq[0] != "END" {
		t.Errorf("Expected stopSequences [\"END\"], got %v", config["stopSequences"])
	}
}

func TestGeminiGenerate(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be application/json")
		}

		// Create response
		response := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{"text": "Hello, I am a test response."},
						},
					},
					"finishReason": "STOP",
				},
			},
		}

		// Encode response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider with test server URL
	baseURLOption := domain.NewBaseURLOption(server.URL)
	provider := NewGeminiProvider(
		"test-api-key",
		"gemini-2.0-flash-lite",
		baseURLOption,
	)

	// Test Generate method
	resp, err := provider.Generate(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if resp != "Hello, I am a test response." {
		t.Errorf("Expected 'Hello, I am a test response.', got '%s'", resp)
	}
}

func TestGeminiGenerateWithSchema(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create response with JSON embedded in text
		response := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{"text": `Here's the information you requested:
{"name": "John Doe", "age": 30, "email": "john@example.com"}`},
						},
					},
				},
			},
		}

		// Encode response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider with test server URL
	baseURLOption := domain.NewBaseURLOption(server.URL)
	provider := NewGeminiProvider(
		"test-api-key",
		"gemini-2.0-flash-lite",
		baseURLOption,
	)

	// Create schema
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name":  {Type: "string"},
			"age":   {Type: "integer"},
			"email": {Type: "string"},
		},
		Required: []string{"name", "email"},
	}

	// Test GenerateWithSchema method
	resp, err := provider.GenerateWithSchema(context.Background(), "Generate info about a person", schema)
	if err != nil {
		t.Fatalf("GenerateWithSchema returned error: %v", err)
	}

	// Check response
	data, ok := resp.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", resp)
	}

	if name, ok := data["name"].(string); !ok || name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %v", data["name"])
	}

	// Check age as a float64 (JSON numbers are decoded as float64)
	if age, ok := data["age"].(float64); !ok || age != 30 {
		t.Errorf("Expected age 30, got %v", data["age"])
	}

	if email, ok := data["email"].(string); !ok || email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got %v", data["email"])
	}
}

func TestGeminiStreamMessage(t *testing.T) {
	// Create a test server for streaming
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's using streamGenerateContent
		if r.URL.Path != "/models/gemini-2.0-flash-lite:streamGenerateContent" {
			t.Errorf("Expected URL path to contain streamGenerateContent, got %s", r.URL.Path)
		}

		// Set up SSE response
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Write chunks of streamed data
		chunks := []string{
			`data: {"candidates":[{"content":{"parts":[{"text":"Hello"}]},"finishReason":""}]}`,
			`data: {"candidates":[{"content":{"parts":[{"text":", "}]},"finishReason":""}]}`,
			`data: {"candidates":[{"content":{"parts":[{"text":"world"}]},"finishReason":""}]}`,
			`data: {"candidates":[{"content":{"parts":[{"text":"!"}]},"finishReason":"STOP"}]}`,
		}

		for _, chunk := range chunks {
			if _, err := w.Write([]byte(chunk + "\n\n")); err != nil {
				t.Errorf("Failed to write chunk: %v", err)
				return
			}
			w.(http.Flusher).Flush()
		}
	}))
	defer server.Close()

	// Create provider with test server URL
	baseURLOption := domain.NewBaseURLOption(server.URL)
	provider := NewGeminiProvider(
		"test-api-key",
		"gemini-2.0-flash-lite",
		baseURLOption,
	)

	// Test StreamMessage method
	messages := []domain.Message{
		{Role: domain.RoleUser, Content: "Hello"},
	}

	stream, err := provider.StreamMessage(context.Background(), messages)
	if err != nil {
		t.Fatalf("StreamMessage returned error: %v", err)
	}

	// Collect tokens from stream
	var tokens []domain.Token
	for token := range stream {
		tokens = append(tokens, token)
	}

	// Check tokens
	if len(tokens) != 4 {
		t.Fatalf("Expected 4 tokens, got %d", len(tokens))
	}

	// Check first token
	if tokens[0].Text != "Hello" || tokens[0].Finished {
		t.Errorf("First token: expected 'Hello', finished=false, got '%s', finished=%v",
			tokens[0].Text, tokens[0].Finished)
	}

	// Check second token
	if tokens[1].Text != ", " || tokens[1].Finished {
		t.Errorf("Second token: expected ', ', finished=false, got '%s', finished=%v",
			tokens[1].Text, tokens[1].Finished)
	}

	// Check third token
	if tokens[2].Text != "world" || tokens[2].Finished {
		t.Errorf("Third token: expected 'world', finished=false, got '%s', finished=%v",
			tokens[2].Text, tokens[2].Finished)
	}

	// Check fourth token
	if tokens[3].Text != "!" || !tokens[3].Finished {
		t.Errorf("Fourth token: expected '!', finished=true, got '%s', finished=%v",
			tokens[3].Text, tokens[3].Finished)
	}
}

func TestGeminiError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return an invalid API key error
		errorResponse := map[string]interface{}{
			"error": map[string]interface{}{
				"code":    401,
				"message": "Invalid API key",
				"status":  "UNAUTHENTICATED",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorResponse)
	}))
	defer server.Close()

	// Create provider with test server URL
	baseURLOption := domain.NewBaseURLOption(server.URL)
	provider := NewGeminiProvider(
		"invalid-api-key",
		"gemini-2.0-flash-lite",
		baseURLOption,
	)

	// Test error handling
	_, err := provider.Generate(context.Background(), "Hello")

	// Check error
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Check error type
	var providerErr *domain.ProviderError
	if !errors.As(err, &providerErr) || !errors.Is(err, domain.ErrAuthenticationFailed) {
		t.Errorf("Expected ErrAuthenticationFailed, got %v", err)
	}

	// Check provider name in error
	if providerErr != nil && providerErr.Provider != "gemini" {
		t.Errorf("Expected provider 'gemini', got '%s'", providerErr.Provider)
	}
}