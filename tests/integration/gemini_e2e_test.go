package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestGeminiE2E tests basic Gemini provider functionality
func TestGeminiE2E(t *testing.T) {
	// Skip test if GEMINI_API_KEY is not set
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set, skipping Gemini e2e test")
	}

	// Create Gemini provider with default model (gemini-2.0-flash-lite)
	geminiProvider := provider.NewGeminiProvider(apiKey, "")

	// Set timeout for all tests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("BasicGeneration", func(t *testing.T) {
		prompt := "What is the capital of France?"

		response, err := geminiProvider.Generate(ctx, prompt)
		if err != nil {
			t.Fatalf("Failed to generate: %v", err)
		}

		if !strings.Contains(strings.ToLower(response), "paris") {
			t.Errorf("Expected response to contain 'Paris', got: %s", response)
		}

		t.Logf("Basic generation: %s", response)
	})

	t.Run("MessageGeneration", func(t *testing.T) {
		messages := []domain.Message{
			{Role: domain.RoleUser, Content: "Tell me about the Eiffel Tower"},
			{Role: domain.RoleAssistant, Content: "The Eiffel Tower is a wrought-iron lattice tower in Paris, France."},
			{Role: domain.RoleUser, Content: "How tall is it?"},
		}

		response, err := geminiProvider.GenerateMessage(ctx, messages)
		if err != nil {
			t.Fatalf("Failed to generate with messages: %v", err)
		}

		// Check that response contains height information
		if !strings.Contains(strings.ToLower(response.Content), "meter") &&
			!strings.Contains(strings.ToLower(response.Content), "feet") {
			t.Errorf("Expected response to contain height information, got: %s", response.Content)
		}

		t.Logf("Message generation: %s", response.Content)
	})

	t.Run("SchemaValidation", func(t *testing.T) {
		// Define a simple schema for a country
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {
					Type:        "string",
					Description: "The name of the country",
				},
				"capital": {
					Type:        "string",
					Description: "The capital city of the country",
				},
				"population": {
					Type:        "integer",
					Description: "The population of the country",
				},
				"languages": {
					Type:        "array",
					Description: "Official languages spoken in the country",
					Items: &schemaDomain.Property{
						Type: "string",
					},
				},
			},
			Required: []string{"name", "capital"},
		}

		prompt := "Generate information about France"

		result, err := geminiProvider.GenerateWithSchema(ctx, prompt, schema)
		if err != nil {
			t.Fatalf("Failed to generate with schema: %v", err)
		}

		// Check that result is a map
		countryData, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be a map, got: %T", result)
		}

		// Check required fields
		if name, ok := countryData["name"].(string); !ok || name == "" {
			t.Errorf("Expected 'name' field to be a non-empty string")
		}

		if capital, ok := countryData["capital"].(string); !ok || capital == "" {
			t.Errorf("Expected 'capital' field to be a non-empty string")
		}

		// Log the schema-validated result
		t.Logf("Generated country data: %v", countryData)
	})

	t.Run("StreamGeneration", func(t *testing.T) {
		prompt := "Count from 1 to 5"

		stream, err := geminiProvider.Stream(ctx, prompt)
		if err != nil {
			t.Fatalf("Failed to create stream: %v", err)
		}

		var fullResponse strings.Builder
		tokenCount := 0

		for token := range stream {
			tokenCount++
			fullResponse.WriteString(token.Text)

			// Log the token
			t.Logf("Stream token %d: %s (finished: %v)",
				tokenCount, token.Text, token.Finished)

			// If token indicates finished, check if we should end early
			if token.Finished && tokenCount > 10 {
				t.Logf("Stream finished after %d tokens", tokenCount)
				break
			}
		}

		// Validate we got some response
		if tokenCount == 0 {
			t.Errorf("Expected at least some tokens in the stream")
		}

		t.Logf("Full streamed response: %s", fullResponse.String())
	})

	t.Run("OptionsConfiguration", func(t *testing.T) {
		// Test with custom temperature
		prompt := "Tell me a joke about programming"

		// First with low temperature (more deterministic)
		lowTempResponse, err := geminiProvider.Generate(ctx, prompt,
			domain.WithTemperature(0.1),
			domain.WithMaxTokens(100),
		)
		if err != nil {
			t.Fatalf("Failed to generate with low temperature: %v", err)
		}

		t.Logf("Response with low temperature (0.1): %s", lowTempResponse)

		// Try with another temperature setting to see difference
		// Note: This isn't a guaranteed test as even with different temperatures,
		// responses might be similar. This is more for demonstration.
		highTempResponse, err := geminiProvider.Generate(ctx, prompt,
			domain.WithTemperature(1.0),
			domain.WithMaxTokens(100),
		)
		if err != nil {
			t.Fatalf("Failed to generate with high temperature: %v", err)
		}

		t.Logf("Response with high temperature (1.0): %s", highTempResponse)

		// Check they're not identical (though they could be by chance)
		if lowTempResponse == highTempResponse {
			t.Logf("Note: Responses with different temperatures were identical, " +
				"which can happen but is less likely with larger temperature differences")
		}
	})
}
