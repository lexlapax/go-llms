package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/fetchers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOllamaProviderIntegration tests the new Ollama convenience provider
func TestOllamaProviderIntegration(t *testing.T) {
	// Skip all tests in this file by default - set ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 to run them
	if os.Getenv("ENABLE_OPENAPI_COMPATIBLE_API_TESTS") != "1" {
		t.Skip("Skipping Ollama provider integration tests - set ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 to run")
	}

	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}

	// Get Ollama model from environment variable or use default
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "llama3.2:3b" // Default to a smaller model if not specified
	}

	// Test convenience wrapper
	t.Run("ConvenienceWrapper", func(t *testing.T) {
		t.Run("DefaultConfiguration", func(t *testing.T) {
			// Test with default configuration
			ollamaProvider := provider.NewOllamaProvider(ollamaModel)

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			prompt := "What is 2+2?"
			response, err := ollamaProvider.Generate(ctx, prompt)

			require.NoError(t, err, "Generate should not return an error")
			assert.NotEmpty(t, response, "Response should not be empty")

			// The response should mention 4
			assert.True(t,
				strings.Contains(response, "4") || strings.Contains(response, "four"),
				"Response should contain the answer")
		})

		t.Run("WithCustomHost", func(t *testing.T) {
			// Test with custom host
			ollamaProvider := provider.NewOllamaProvider(
				ollamaModel,
				provider.WithOllamaHost(ollamaHost),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			prompt := "Say hello in one word"
			response, err := ollamaProvider.Generate(ctx, prompt,
				domain.WithMaxTokens(10),
			)

			require.NoError(t, err, "Generate should not return an error")
			assert.NotEmpty(t, response, "Response should not be empty")
		})

		t.Run("WithCustomTimeout", func(t *testing.T) {
			// Test with custom timeout
			customTimeout := 90 * time.Second
			ollamaProvider := provider.NewOllamaProvider(
				ollamaModel,
				provider.WithOllamaHost(ollamaHost),
				provider.WithOllamaTimeout(customTimeout),
			)

			// This test just verifies the provider is created correctly
			// Actual timeout behavior is tested elsewhere
			assert.NotNil(t, ollamaProvider)
		})

		t.Run("WithMixedOptions", func(t *testing.T) {
			// Test mixing Ollama-specific options with standard options
			ollamaProvider := provider.NewOllamaProvider(
				ollamaModel,
				provider.WithOllamaHost(ollamaHost),
				provider.WithOllamaTimeout(60*time.Second),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			messages := []domain.Message{
				domain.NewTextMessage(domain.RoleUser, "What is Go programming language?"),
			}

			response, err := ollamaProvider.GenerateMessage(ctx, messages,
				domain.WithMaxTokens(100),
				domain.WithTemperature(0.5),
			)

			require.NoError(t, err, "GenerateMessage should not return an error")
			assert.NotEmpty(t, response.Content, "Response should not be empty")

			// Check for reasonable content
			lowerResponse := strings.ToLower(response.Content)
			assert.True(t,
				strings.Contains(lowerResponse, "go") ||
					strings.Contains(lowerResponse, "golang") ||
					strings.Contains(lowerResponse, "programming"),
				"Response should be about Go programming")
		})
	})

	// Test model listing
	t.Run("ModelListing", func(t *testing.T) {
		fetcher := fetchers.NewOllamaFetcher(ollamaHost, nil)
		models, err := fetcher.FetchModels()

		require.NoError(t, err, "FetchModels should not return an error")
		assert.NotEmpty(t, models, "Should have at least one model")

		// Check that the test model is in the list
		modelFound := false
		for _, model := range models {
			if model.Name == ollamaModel {
				modelFound = true
				// Verify model properties
				assert.Equal(t, "ollama", model.Provider)
				assert.NotEmpty(t, model.DisplayName)
				assert.NotEmpty(t, model.Description)
				assert.Greater(t, model.ContextWindow, 0)
				assert.True(t, model.Capabilities.Streaming)
				assert.True(t, model.Capabilities.FunctionCalling)
				assert.Equal(t, 0.0, model.Pricing.InputPer1kTokens)
				assert.Equal(t, 0.0, model.Pricing.OutputPer1kTokens)
				break
			}
		}
		assert.True(t, modelFound, "Test model should be in the fetched models list")
	})

	// Test streaming with convenience wrapper
	t.Run("StreamingWithWrapper", func(t *testing.T) {
		ollamaProvider := provider.NewOllamaProvider(
			ollamaModel,
			provider.WithOllamaHost(ollamaHost),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		prompt := "List three colors, one per line."
		stream, err := ollamaProvider.Stream(ctx, prompt)

		require.NoError(t, err, "Stream should not return an error")

		// Collect tokens
		tokens := make([]domain.Token, 0)
		for token := range stream {
			tokens = append(tokens, token)
		}

		// Check streaming behavior
		assert.NotEmpty(t, tokens, "Should receive tokens")
		assert.True(t, tokens[len(tokens)-1].Finished, "Last token should be marked as finished")

		// Combine all tokens
		var fullResponse strings.Builder
		for _, token := range tokens {
			fullResponse.WriteString(token.Text)
		}

		// Check the response mentions colors
		response := strings.ToLower(fullResponse.String())
		colorWords := []string{"red", "blue", "green", "yellow", "orange", "purple", "black", "white", "pink", "brown"}
		colorCount := 0
		for _, color := range colorWords {
			if strings.Contains(response, color) {
				colorCount++
			}
		}
		assert.GreaterOrEqual(t, colorCount, 1, "Response should contain at least one color")
	})
}
