package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

func TestAnthropicProviderOptions(t *testing.T) {
	// Create a mock HTTP server to simulate Anthropic API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check various headers and request body
		apiKey := r.Header.Get("x-api-key")
		_ = r.Header.Get("Content-Type") // We check but don't use this in the test
		
		// Read the body to extract system prompt and metadata
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		// Extract system prompt and metadata from request body
		systemPrompt, _ := requestBody["system"].(string)
		metadata, _ := requestBody["metadata"].(map[string]interface{})
		
		// Just a simple response for the test
		w.Header().Set("Content-Type", "application/json")
		response := fmt.Sprintf(`{
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"content": [
				{
					"type": "text",
					"text": "This is a test response. API Key: %s, System Prompt: %s, Metadata Present: %t"
				}
			],
			"model": "claude-3-5-sonnet-latest",
			"stop_reason": "end_turn",
			"stop_sequence": null
		}`, apiKey, systemPrompt, metadata != nil)
		
		w.Write([]byte(response))
	}))
	defer mockServer.Close()

	t.Run("BaseURLOption", func(t *testing.T) {
		baseURLOption := domain.NewBaseURLOption(mockServer.URL)
		provider := NewAnthropicProvider("test-key", "claude-3-5-sonnet-latest", baseURLOption)
		
		if provider.baseURL != mockServer.URL {
			t.Errorf("Expected baseURL to be %s, got %s", mockServer.URL, provider.baseURL)
		}
	})
	
	t.Run("HTTPClientOption", func(t *testing.T) {
		customClient := &http.Client{Timeout: 30 * time.Second}
		clientOption := domain.NewHTTPClientOption(customClient)
		provider := NewAnthropicProvider("test-key", "claude-3-5-sonnet-latest", clientOption)
		
		if provider.httpClient != customClient {
			t.Error("HTTPClientOption failed to set the HTTP client")
		}
	})
	
	t.Run("SystemPromptOption", func(t *testing.T) {
		systemPrompt := "You are a helpful assistant that answers questions concisely."
		systemPromptOption := domain.NewAnthropicSystemPromptOption(systemPrompt)
		baseURLOption := domain.NewBaseURLOption(mockServer.URL)
		
		provider := NewAnthropicProvider("test-key", "claude-3-5-sonnet-latest", systemPromptOption, baseURLOption)
		
		// Check that the option field was updated
		if provider.systemPrompt != systemPrompt {
			t.Errorf("Expected systemPrompt to be %s, got %s", systemPrompt, provider.systemPrompt)
		}
		
		// Make a request and check if the system prompt is included
		ctx := context.Background()
		response, err := provider.Generate(ctx, "Test prompt")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		// Response should contain the system prompt
		if !strings.Contains(response, systemPrompt) {
			t.Errorf("Expected response to contain system prompt %s, got: %s", systemPrompt, response)
		}
	})
	
	t.Run("MetadataOption", func(t *testing.T) {
		metadata := map[string]string{
			"user_id": "test123",
			"session_id": "abc456",
		}
		metadataOption := domain.NewAnthropicMetadataOption(metadata)
		baseURLOption := domain.NewBaseURLOption(mockServer.URL)
		
		provider := NewAnthropicProvider("test-key", "claude-3-5-sonnet-latest", metadataOption, baseURLOption)
		
		// Check that the option field was updated
		if len(provider.metadata) != len(metadata) {
			t.Errorf("Expected metadata to have %d items, got %d", len(metadata), len(provider.metadata))
		}
		
		for k, v := range metadata {
			if provider.metadata[k] != v {
				t.Errorf("Expected metadata[%s] to be %s, got %s", k, v, provider.metadata[k])
			}
		}
		
		// Make a request and check if metadata was included
		ctx := context.Background()
		response, err := provider.Generate(ctx, "Test prompt")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		// Response should indicate metadata was present
		if !strings.Contains(response, "Metadata Present: true") {
			t.Errorf("Expected response to indicate metadata was present, got: %s", response)
		}
	})
	
	t.Run("Multiple Options", func(t *testing.T) {
		customClient := &http.Client{Timeout: 30 * time.Second}
		systemPrompt := "You are a helpful assistant that answers questions concisely."
		metadata := map[string]string{
			"user_id": "test123",
			"session_id": "abc456",
		}
		
		options := []domain.ProviderOption{
			domain.NewBaseURLOption(mockServer.URL),
			domain.NewHTTPClientOption(customClient),
			domain.NewAnthropicSystemPromptOption(systemPrompt),
			domain.NewAnthropicMetadataOption(metadata),
		}
		
		provider := NewAnthropicProvider("test-key", "claude-3-5-sonnet-latest", options...)
		
		// Check that all options were applied correctly
		if provider.baseURL != mockServer.URL {
			t.Errorf("Expected baseURL to be %s, got %s", mockServer.URL, provider.baseURL)
		}
		
		if provider.httpClient != customClient {
			t.Error("HTTPClientOption failed to set the HTTP client")
		}
		
		if provider.systemPrompt != systemPrompt {
			t.Errorf("Expected systemPrompt to be %s, got %s", systemPrompt, provider.systemPrompt)
		}
		
		if len(provider.metadata) != len(metadata) {
			t.Errorf("Expected metadata to have %d items, got %d", len(metadata), len(provider.metadata))
		}
		
		// Make a request to verify all options are applied
		ctx := context.Background()
		response, err := provider.Generate(ctx, "Test prompt")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		// Response should contain indicators that options were applied
		if !strings.Contains(response, systemPrompt) {
			t.Errorf("Expected response to contain system prompt, got: %s", response)
		}
		
		if !strings.Contains(response, "Metadata Present: true") {
			t.Errorf("Expected response to indicate metadata was present, got: %s", response)
		}
	})
}