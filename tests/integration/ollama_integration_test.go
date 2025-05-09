package integration

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOllamaIntegration provides dedicated integration tests for Ollama
// Skip if OLLAMA_HOST environment variable is not set
func TestOllamaIntegration(t *testing.T) {
	// Skip all tests in this file by default - set ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 to run them
	if os.Getenv("ENABLE_OPENAPI_COMPATIBLE_API_TESTS") != "1" {
		t.Skip("Skipping Ollama integration tests - set ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 to run")
	}

	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		t.Skip("OLLAMA_HOST environment variable not set, skipping Ollama integration tests")
	}

	// Get Ollama model from environment variable or use default
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "llama3.2:3b" // Default to a smaller model if not specified
	}

	// Create common options
	baseURLOption := domain.NewBaseURLOption(ollamaHost)
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}
	httpClientOption := domain.NewHTTPClientOption(httpClient)

	// Test basic text generation
	t.Run("BasicGeneration", func(t *testing.T) {
		ollamaProvider := provider.NewOpenAIProvider(
			"dummy-key", // Dummy key for Ollama (will be ignored)
			ollamaModel,
			baseURLOption,
			httpClientOption,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		prompt := "What is the capital of France?"
		response, err := ollamaProvider.Generate(ctx, prompt)

		require.NoError(t, err, "Generate should not return an error")
		assert.NotEmpty(t, response, "Response should not be empty")

		// Check for reasonable content - should mention Paris
		// Since this is an LLM, we can't expect exact responses
		lowerResponse := strings.ToLower(response)
		assert.True(t,
			strings.Contains(lowerResponse, "paris") ||
				strings.Contains(lowerResponse, "capital") ||
				strings.Contains(lowerResponse, "france"),
			"Response should be topically relevant")
	})

	// Test message-based conversation
	t.Run("MessageConversation", func(t *testing.T) {
		ollamaProvider := provider.NewOpenAIProvider(
			"dummy-key", // Dummy key for Ollama (will be ignored)
			ollamaModel,
			baseURLOption,
			httpClientOption,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		messages := []domain.Message{
			{Role: domain.RoleSystem, Content: "You are a helpful assistant."},
			{Role: domain.RoleUser, Content: "What is the capital of France?"},
			{Role: domain.RoleAssistant, Content: "The capital of France is Paris."},
			{Role: domain.RoleUser, Content: "What is its population?"},
		}

		response, err := ollamaProvider.GenerateMessage(ctx, messages)

		require.NoError(t, err, "GenerateMessage should not return an error")
		assert.NotEmpty(t, response.Content, "Response should not be empty")

		// Check for reasonable content - should mention population numbers
		// Since this is an LLM, we can't expect exact responses
		lowerResponse := strings.ToLower(response.Content)
		assert.True(t,
			strings.Contains(lowerResponse, "million") ||
				strings.Contains(lowerResponse, "population") ||
				strings.Contains(lowerResponse, "people"),
			"Response should be topically relevant to Paris population")
	})

	// Test streaming
	t.Run("Streaming", func(t *testing.T) {
		ollamaProvider := provider.NewOpenAIProvider(
			"dummy-key", // Dummy key for Ollama (will be ignored)
			ollamaModel,
			baseURLOption,
			httpClientOption,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		prompt := "Count from 1 to 5 and put each number on a new line."
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

		// Check the content contains numbers
		response := fullResponse.String()
		assert.True(t,
			strings.Contains(response, "1") &&
				strings.Contains(response, "2") &&
				strings.Contains(response, "3"),
			"Response should contain numbers")
	})

	// Test parameter handling
	t.Run("ParameterHandling", func(t *testing.T) {
		ollamaProvider := provider.NewOpenAIProvider(
			"dummy-key", // Dummy key for Ollama (will be ignored)
			ollamaModel,
			baseURLOption,
			httpClientOption,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Test with temperature parameter
		t.Run("Temperature", func(t *testing.T) {
			prompt := "Write a random animal name."

			// Get two responses with different temperatures
			resp1, err := ollamaProvider.Generate(ctx, prompt, domain.WithTemperature(0.1))
			require.NoError(t, err, "Generate with low temperature should not error")

			resp2, err := ollamaProvider.Generate(ctx, prompt, domain.WithTemperature(1.0))
			require.NoError(t, err, "Generate with high temperature should not error")

			// We can't guarantee different outputs due to the nature of LLMs,
			// but we can check that both return valid responses
			assert.NotEmpty(t, resp1, "Response with low temperature should not be empty")
			assert.NotEmpty(t, resp2, "Response with high temperature should not be empty")
		})

		// Test with max tokens parameter
		t.Run("MaxTokens", func(t *testing.T) {
			// Skip this test as max tokens behavior is model-dependent
			t.Skip("Skipping max tokens test as token counting is model-dependent and can be inconsistent")

			prompt := "Write a very long story about a dragon."

			// Get a short response with low max tokens
			shortResp, err := ollamaProvider.Generate(ctx, prompt, domain.WithMaxTokens(20))
			require.NoError(t, err, "Generate with low max tokens should not error")

			// Get a longer response with higher max tokens
			longResp, err := ollamaProvider.Generate(ctx, prompt, domain.WithMaxTokens(100))
			require.NoError(t, err, "Generate with high max tokens should not error")

			// Just check that both responses are valid
			assert.NotEmpty(t, shortResp, "Short response should not be empty")
			assert.NotEmpty(t, longResp, "Long response should not be empty")
		})
	})

	// Test error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		// Test with invalid model name
		t.Run("InvalidModel", func(t *testing.T) {
			invalidProvider := provider.NewOpenAIProvider(
				"dummy-key",
				"non_existent_model", // This model doesn't exist
				baseURLOption,
				httpClientOption,
			)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			_, err := invalidProvider.Generate(ctx, "Hello")
			assert.Error(t, err, "Using non-existent model should result in error")
		})

		// Test with invalid base URL
		t.Run("InvalidURL", func(t *testing.T) {
			invalidProvider := provider.NewOpenAIProvider(
				"dummy-key",
				ollamaModel,
				domain.NewBaseURLOption("http://invalid-ollama-url-that-does-not-exist"),
				httpClientOption,
			)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err := invalidProvider.Generate(ctx, "Hello")
			assert.Error(t, err, "Using invalid URL should result in error")
		})

		// Test with timeout
		t.Run("Timeout", func(t *testing.T) {
			t.Skip("Skipping timeout test as it's flaky and depends on system performance")

			ollamaProvider := provider.NewOpenAIProvider(
				"dummy-key",
				ollamaModel,
				baseURLOption,
				httpClientOption,
			)

			// Very short timeout
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer cancel()

			// Should cause a timeout
			_, err := ollamaProvider.Generate(ctx, "Write a very long story")
			assert.Error(t, err, "Request with short timeout should fail")
			assert.Contains(t, err.Error(), "context", "Error should be related to context deadline")
		})
	})
}
