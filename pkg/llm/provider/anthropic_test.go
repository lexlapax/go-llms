package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestAnthropicProvider(t *testing.T) {
	// Create a mock HTTP server to simulate Anthropic API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if authorization header is present
		if r.Header.Get("x-api-key") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, `{"error":{"type":"auth_error","message":"Invalid API key"}}`)
			return
		}

		// Handle different API endpoints
		switch r.URL.Path {
		case "/v1/messages":
			// Standard completion response for messages API
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{
				"id": "msg_123",
				"type": "message",
				"role": "assistant",
				"content": [
					{
						"type": "text",
						"text": "This is a test response from the Anthropic API."
					}
				],
				"model": "claude-3-sonnet-20240229",
				"stop_reason": "end_turn"
			}`)
		case "/v1/complete":
			// Legacy completion API
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{
				"completion": "This is a test response from the Anthropic API.",
				"stop_reason": "stop_sequence",
				"model": "claude-2.0"
			}`)
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, `{"error":{"type":"invalid_request_error","message":"Not found"}}`)
		}
	}))
	defer mockServer.Close()

	ctx := context.Background()

	t.Run("NewAnthropicProvider", func(t *testing.T) {
		provider := NewAnthropicProvider("test-api-key", "claude-3-sonnet-20240229", WithAnthropicBaseURL(mockServer.URL))
		if provider == nil {
			t.Fatal("Expected non-nil provider")
		}
		if provider.apiKey != "test-api-key" {
			t.Errorf("Expected API key 'test-api-key', got '%s'", provider.apiKey)
		}
		if provider.model != "claude-3-sonnet-20240229" {
			t.Errorf("Expected model 'claude-3-sonnet-20240229', got '%s'", provider.model)
		}
	})

	t.Run("Generate", func(t *testing.T) {
		provider := NewAnthropicProvider("test-api-key", "claude-3-sonnet-20240229", WithAnthropicBaseURL(mockServer.URL))
		response, err := provider.Generate(ctx, "Tell me a joke")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		expected := "This is a test response from the Anthropic API."
		if response != expected {
			t.Errorf("Expected response '%s', got '%s'", expected, response)
		}
	})

	t.Run("GenerateMessage", func(t *testing.T) {
		provider := NewAnthropicProvider("test-api-key", "claude-3-sonnet-20240229", WithAnthropicBaseURL(mockServer.URL))
		messages := []domain.Message{
			{Role: domain.RoleSystem, Content: "You are a helpful assistant"},
			{Role: domain.RoleUser, Content: "Tell me a joke"},
		}
		response, err := provider.GenerateMessage(ctx, messages)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		expected := "This is a test response from the Anthropic API."
		if response.Content != expected {
			t.Errorf("Expected response content '%s', got '%s'", expected, response.Content)
		}
	})

	t.Run("Invalid API key", func(t *testing.T) {
		provider := NewAnthropicProvider("invalid-key", "claude-3-sonnet-20240229", WithAnthropicBaseURL(mockServer.URL))
		_, err := provider.Generate(ctx, "Tell me a joke")
		if err == nil {
			t.Fatal("Expected error for invalid API key, got nil")
		}
	})

	t.Run("Generate with options", func(t *testing.T) {
		provider := NewAnthropicProvider("test-api-key", "claude-3-sonnet-20240229", WithAnthropicBaseURL(mockServer.URL))

		// Set custom options
		options := []domain.Option{
			domain.WithTemperature(0.2),
			domain.WithMaxTokens(100),
		}

		response, err := provider.Generate(ctx, "Tell me a joke", options...)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expected := "This is a test response from the Anthropic API."
		if response != expected {
			t.Errorf("Expected response '%s', got '%s'", expected, response)
		}
	})

	// Add a mock HTTP server for streaming responses
	streamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if authorization header is present
		if r.Header.Get("x-api-key") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, `{"error":{"type":"auth_error","message":"Invalid API key"}}`)
			return
		}

		// Check if correct content type and accept headers are set
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, `{"error":{"type":"invalid_request_error","message":"Invalid content type"}}`)
			return
		}

		// Stream response
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Simulate a streaming response for Anthropic
		fmt.Fprint(w, "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_123\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[],\"model\":\"claude-3-sonnet-20240229\"}}\n\n")
		fmt.Fprint(w, "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\"}}\n\n")
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"This \"}}\n\n")
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"is \"}}\n\n")
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"a \"}}\n\n")
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"streaming \"}}\n\n")
		fmt.Fprint(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"response.\"}}\n\n")
		fmt.Fprint(w, "event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n")
		fmt.Fprint(w, "event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\",\"stop_sequence\":null},\"usage\":{\"output_tokens\":5}}\n\n")
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	}))
	defer streamServer.Close()

	t.Run("Stream", func(t *testing.T) {
		provider := NewAnthropicProvider("test-api-key", "claude-3-sonnet-20240229", WithAnthropicBaseURL(streamServer.URL))
		stream, err := provider.Stream(ctx, "Tell me a joke")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Collect all tokens
		var tokens []domain.Token
		var tokensWithContent []domain.Token
		for token := range stream {
			tokens = append(tokens, token)
			// Filter only tokens with content
			if token.Text != "" {
				tokensWithContent = append(tokensWithContent, token)
			}
		}

		if len(tokensWithContent) == 0 {
			t.Fatal("Expected at least one token with content")
		}

		// At minimum, make sure we have some tokens and the last one is marked as finished
		if len(tokens) > 0 {
			lastToken := tokens[len(tokens)-1]
			if !lastToken.Finished {
				t.Error("Expected last token to be marked as finished")
			}
		}
	})

	t.Run("StreamMessage", func(t *testing.T) {
		provider := NewAnthropicProvider("test-api-key", "claude-3-sonnet-20240229", WithAnthropicBaseURL(streamServer.URL))
		messages := []domain.Message{
			{Role: domain.RoleSystem, Content: "You are a helpful assistant"},
			{Role: domain.RoleUser, Content: "Tell me a joke"},
		}
		stream, err := provider.StreamMessage(ctx, messages)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Collect all tokens
		var tokens []domain.Token
		var tokensWithContent []domain.Token
		for token := range stream {
			tokens = append(tokens, token)
			// Filter only tokens with content
			if token.Text != "" {
				tokensWithContent = append(tokensWithContent, token)
			}
		}

		if len(tokensWithContent) == 0 {
			t.Fatal("Expected at least one token with content")
		}

		// At minimum, make sure we have some tokens and the last one is marked as finished
		if len(tokens) > 0 {
			lastToken := tokens[len(tokens)-1]
			if !lastToken.Finished {
				t.Error("Expected last token to be marked as finished")
			}
		}
	})

	// Setup schema testing server
	schemaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a valid JSON response that conforms to our schema
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"id": "msg_123",
			"type": "message",
			"role": "assistant",
			"content": [
				{
					"type": "text",
					"text": "{\"name\":\"John Doe\",\"age\":30,\"email\":\"john@example.com\"}"
				}
			],
			"model": "claude-3-sonnet-20240229",
			"stop_reason": "end_turn"
		}`)
	}))
	defer schemaServer.Close()

	t.Run("GenerateWithSchema", func(t *testing.T) {
		provider := NewAnthropicProvider("test-api-key", "claude-3-sonnet-20240229", WithAnthropicBaseURL(schemaServer.URL))

		// Define a simple schema
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name":  {Type: "string", Description: "Person's name"},
				"age":   {Type: "integer", Description: "Person's age"},
				"email": {Type: "string", Format: "email", Description: "Person's email"},
			},
			Required: []string{"name", "email"},
		}

		result, err := provider.GenerateWithSchema(ctx, "Generate a person", schema)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check result is a map
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got %T", result)
		}

		// Check required fields
		name, hasName := resultMap["name"]
		if !hasName {
			t.Error("Expected 'name' field in result")
		}
		nameStr, ok := name.(string)
		if !ok || nameStr != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%v'", name)
		}

		age, hasAge := resultMap["age"]
		if !hasAge {
			t.Error("Expected 'age' field in result")
		}
		if age != float64(30) {
			t.Errorf("Expected age 30, got '%v'", age)
		}

		email, hasEmail := resultMap["email"]
		if !hasEmail {
			t.Error("Expected 'email' field in result")
		}
		emailStr, ok := email.(string)
		if !ok || emailStr != "john@example.com" {
			t.Errorf("Expected email 'john@example.com', got '%v'", email)
		}
	})
}
