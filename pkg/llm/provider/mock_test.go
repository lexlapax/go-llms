package provider

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestMockProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("default generate response", func(t *testing.T) {
		mock := NewMockProvider()

		response, err := mock.Generate(ctx, "Tell me about Go")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if response == "" {
			t.Error("Expected non-empty response")
		}

		// Check default response has expected format (JSON)
		if response[0] != '{' || response[len(response)-1] != '}' {
			t.Errorf("Expected JSON response, got: %s", response)
		}
	})

	t.Run("custom generate response", func(t *testing.T) {
		mock := NewMockProvider()
		customResponse := "This is a custom response"

		// Set custom response function
		mock.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return customResponse, nil
		})

		response, err := mock.Generate(ctx, "Tell me about Go")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if response != customResponse {
			t.Errorf("Expected: %s, got: %s", customResponse, response)
		}
	})

	t.Run("generate with message", func(t *testing.T) {
		mock := NewMockProvider()

		messages := []domain.Message{
			{Role: domain.RoleSystem, Content: "You are a helpful assistant."},
			{Role: domain.RoleUser, Content: "Tell me about Go"},
		}

		response, err := mock.GenerateMessage(ctx, messages)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if response.Content == "" {
			t.Error("Expected non-empty response content")
		}
	})

	t.Run("streaming response", func(t *testing.T) {
		mock := NewMockProvider()

		stream, err := mock.Stream(ctx, "Tell me about Go")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Collect all tokens
		var tokens []domain.Token
		for token := range stream {
			tokens = append(tokens, token)
		}

		if len(tokens) == 0 {
			t.Error("Expected at least one token")
		}

		// Last token should be marked as finished
		if !tokens[len(tokens)-1].Finished {
			t.Error("Expected last token to be marked as finished")
		}
	})

	t.Run("custom streaming response", func(t *testing.T) {
		mock := NewMockProvider()
		expectedTokens := []domain.Token{
			{Text: "This", Finished: false},
			{Text: " is", Finished: false},
			{Text: " a", Finished: false},
			{Text: " custom", Finished: false},
			{Text: " stream", Finished: true},
		}

		// Set custom stream function
		mock.WithStreamFunc(func(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
			ch := make(chan domain.Token)
			go func() {
				defer close(ch)
				for _, token := range expectedTokens {
					select {
					case <-ctx.Done():
						return
					case ch <- token:
						time.Sleep(10 * time.Millisecond) // Simulate delay
					}
				}
			}()
			return ch, nil
		})

		stream, err := mock.Stream(ctx, "Tell me about Go")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Collect all tokens
		var tokens []domain.Token
		for token := range stream {
			tokens = append(tokens, token)
		}

		if len(tokens) != len(expectedTokens) {
			t.Errorf("Expected %d tokens, got %d", len(expectedTokens), len(tokens))
		}

		for i, token := range tokens {
			if token.Text != expectedTokens[i].Text {
				t.Errorf("Token %d: expected text '%s', got '%s'", i, expectedTokens[i].Text, token.Text)
			}
			if token.Finished != expectedTokens[i].Finished {
				t.Errorf("Token %d: expected finished '%v', got '%v'", i, expectedTokens[i].Finished, token.Finished)
			}
		}
	})

	t.Run("generate with schema", func(t *testing.T) {
		mock := NewMockProvider()

		// Define a simple schema
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
			Required: []string{"name"},
		}

		// Set custom schema generation function
		mock.WithGenerateWithSchemaFunc(func(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
			return map[string]interface{}{
				"name": "John Doe",
				"age":  30,
			}, nil
		})

		result, err := mock.GenerateWithSchema(ctx, "Generate a person", schema)
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
		if name != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%v'", name)
		}

		age, hasAge := resultMap["age"]
		if !hasAge {
			t.Error("Expected 'age' field in result")
		}
		if age != 30 {
			t.Errorf("Expected age 30, got '%v'", age)
		}
	})
}
