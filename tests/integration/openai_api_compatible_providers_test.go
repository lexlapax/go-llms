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
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

// TestOpenAIAPICompatibleProvidersIntegration tests integration with providers that implement the OpenAI API specification
// These tests require actual API keys and are skipped by default
func TestOpenAIAPICompatibleProvidersIntegration(t *testing.T) {
	// Skip all tests in this file by default - set ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 to run them
	if os.Getenv("ENABLE_OPENAPI_COMPATIBLE_API_TESTS") != "1" {
		t.Skip("Skipping OpenAI API Compatible Providers tests - set ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 to run")
	}

	// Each sub-test will automatically skip if the required environment variables are missing

	// Test OpenRouter integration
	t.Run("OpenRouter", func(t *testing.T) {
		apiKey := os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			t.Skip("OPENROUTER_API_KEY environment variable not set, skipping test")
		}
		// Get the model name from environment variable or use default
		model := os.Getenv("OPENROUTER_MODEL")
		if model == "" {
			model = "mistralai/mistral-small-3.1-24b-instruct:free"
		}

		// Method 1: Direct provider instantiation with interface-based options
		t.Run("DirectInstantiation", func(t *testing.T) {
			// Create a custom HTTP client with timeout
			httpClient := &http.Client{
				Timeout: 30 * time.Second,
			}

			// Create the provider options
			// For OpenRouter, we need to omit the "/v1" as the OpenAI provider will add it
			baseURLOption := domain.NewBaseURLOption("https://openrouter.ai/api")
			httpClientOption := domain.NewHTTPClientOption(httpClient)
			headersOption := domain.NewHeadersOption(map[string]string{
				"HTTP-Referer": "https://github.com/lexlapax/go-llms", // OpenRouter attribution
				"X-Title":      "Go-LLMs Test",                        // Additional OpenRouter headers
				"Content-Type": "application/json",
			})

			// Create the provider with multiple options
			openRouterProvider := provider.NewOpenAIProvider(
				apiKey,
				model,
				baseURLOption,
				httpClientOption,
				headersOption,
			)

			// Use the provider to generate a response with messages (preferred for OpenRouter)
			messages := []domain.Message{
				{Role: domain.RoleUser, Content: "What is the capital of France?"},
			}

			resp, err := openRouterProvider.GenerateMessage(
				context.Background(),
				messages,
				domain.WithMaxTokens(100),
			)

			if err != nil {
				t.Fatalf("OpenRouter GenerateMessage failed: %v", err)
			}

			if resp.Content == "" {
				t.Errorf("Expected non-empty response, got empty string")
			}

			// Check that the response mentions Paris
			if !strings.Contains(resp.Content, "Paris") {
				t.Errorf("Expected response to mention 'Paris', got: %s", resp.Content)
			}
		})

		// Method 2: Using ModelConfig and CreateProvider
		t.Run("ModelConfig", func(t *testing.T) {
			config := llmutil.ModelConfig{
				Provider: "openai",
				Model:    model,
				APIKey:   apiKey,
				BaseURL:  "https://openrouter.ai/api",
			}

			openRouterProvider, err := llmutil.CreateProvider(config)
			if err != nil {
				t.Fatalf("Error creating provider: %v", err)
			}

			// Use the provider to generate a response with messages (preferred for OpenRouter)
			messages := []domain.Message{
				{Role: domain.RoleUser, Content: "What is the capital of France?"},
			}

			resp, err := openRouterProvider.GenerateMessage(
				context.Background(),
				messages,
				domain.WithMaxTokens(100),
			)

			if err != nil {
				t.Fatalf("OpenRouter GenerateMessage failed: %v", err)
			}

			if resp.Content == "" {
				t.Errorf("Expected non-empty response, got empty string")
			}

			// Check that the response mentions Paris
			if !strings.Contains(resp.Content, "Paris") {
				t.Errorf("Expected response to mention 'Paris', got: %s", resp.Content)
			}
		})
	})

	// Test Ollama integration
	t.Run("Ollama", func(t *testing.T) {
		ollamaHost := os.Getenv("OLLAMA_HOST")
		if ollamaHost == "" {
			t.Skip("OLLAMA_HOST environment variable not set, skipping test")
		}

		// Get Ollama model from environment variable or use default
		ollamaModel := os.Getenv("OLLAMA_MODEL")
		if ollamaModel == "" {
			ollamaModel = "llama3.2:3b"
		}

		// Method 1: Direct provider instantiation with interface-based options
		t.Run("DirectInstantiation", func(t *testing.T) {
			// Create a custom HTTP client with timeout
			ollamaClient := &http.Client{
				Timeout: 60 * time.Second, // Longer timeout for local models
			}

			// Create the provider options
			ollamaBaseURLOption := domain.NewBaseURLOption(ollamaHost)
			ollamaHTTPClientOption := domain.NewHTTPClientOption(ollamaClient)

			// Create the provider with multiple options
			ollamaProvider := provider.NewOpenAIProvider(
				"dummy-key", // Dummy key for Ollama - needed for OpenAI provider but will be ignored
				ollamaModel,
				ollamaBaseURLOption,
				ollamaHTTPClientOption,
			)

			// Use the provider to generate a response
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			resp, err := ollamaProvider.Generate(
				ctx,
				"What is the capital of France?",
				domain.WithMaxTokens(100),
				domain.WithTemperature(0.7),
			)

			if err != nil {
				t.Fatalf("Ollama Generate failed: %v", err)
			}

			if resp == "" {
				t.Errorf("Expected non-empty response, got empty string")
			}

			// Check that the response mentions Paris
			// Note: Local models may be less deterministic, so this check is more relaxed
			if !strings.Contains(strings.ToLower(resp), "paris") {
				t.Logf("Note: Response doesn't contain 'Paris'. This may be acceptable for local models.")
				t.Logf("Received response: %s", resp)
			}
		})

		// Method 2: Using ModelConfig and CreateProvider
		t.Run("ModelConfig", func(t *testing.T) {
			config := llmutil.ModelConfig{
				Provider: "openai",
				Model:    ollamaModel,
				APIKey:   "dummy-key", // Dummy key for Ollama - needed for OpenAI provider but will be ignored
				BaseURL:  ollamaHost,
			}

			ollamaProvider, err := llmutil.CreateProvider(config)
			if err != nil {
				t.Fatalf("Error creating provider: %v", err)
			}

			// Use the provider to generate a response
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			resp, err := ollamaProvider.Generate(
				ctx,
				"What is the capital of France?",
				domain.WithMaxTokens(100),
			)

			if err != nil {
				t.Fatalf("Ollama Generate failed: %v", err)
			}

			if resp == "" {
				t.Errorf("Expected non-empty response, got empty string")
			}

			// Check that the response mentions Paris
			// Note: Local models may be less deterministic, so this check is more relaxed
			if !strings.Contains(strings.ToLower(resp), "paris") {
				t.Logf("Note: Response doesn't contain 'Paris'. This may be acceptable for local models.")
				t.Logf("Received response: %s", resp)
			}
		})

		// Test streaming with Ollama
		t.Run("Streaming", func(t *testing.T) {
			config := llmutil.ModelConfig{
				Provider: "openai",
				Model:    ollamaModel,
				APIKey:   "dummy-key", // Dummy key for Ollama - needed for OpenAI provider but will be ignored
				BaseURL:  ollamaHost,
			}

			ollamaProvider, err := llmutil.CreateProvider(config)
			if err != nil {
				t.Fatalf("Error creating provider: %v", err)
			}

			// Use the provider for streaming
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			stream, err := ollamaProvider.Stream(
				ctx,
				"What is the capital of France?",
				domain.WithMaxTokens(100),
			)

			if err != nil {
				t.Fatalf("Ollama Stream failed: %v", err)
			}

			fullResponse := ""
			for token := range stream {
				fullResponse += token.Text
			}

			if fullResponse == "" {
				t.Errorf("Expected non-empty streamed response, got empty string")
			}

			// Check that the response mentions Paris
			// Note: Local models may be less deterministic, so this check is more relaxed
			if !strings.Contains(strings.ToLower(fullResponse), "paris") {
				t.Logf("Note: Response doesn't contain 'Paris'. This may be acceptable for local models.")
				t.Logf("Received response: %s", fullResponse)
			}
		})
	})
}
