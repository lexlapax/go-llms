package main

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

func TestProviderOptions(t *testing.T) {
	// Skip if not running with -short flag
	if testing.Short() {
		t.Skip("Skipping provider options test in short mode")
	}

	// Create a test context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test common options with mock provider
	t.Run("CommonOptions", func(t *testing.T) {
		// Create custom HTTP client
		customClient := &http.Client{
			Timeout: 30 * time.Second,
		}

		// Create custom headers
		customHeaders := map[string]string{
			"X-Test-Header": "test-value",
			"User-Agent":    "test-agent",
		}

		// Create options
		httpClientOption := domain.NewHTTPClientOption(customClient)
		headersOption := domain.NewHeadersOption(customHeaders)
		baseURLOption := domain.NewBaseURLOption("https://mock-api.example.com")

		// Create a mock provider with options
		mockProvider := provider.NewMockProvider(
			httpClientOption,
			headersOption,
			baseURLOption,
		)

		// Verify the provider is created correctly
		if mockProvider == nil {
			t.Fatal("Failed to create mock provider")
		}

		// Test generation
		response, err := mockProvider.Generate(ctx, "Hello")
		if err != nil {
			t.Fatalf("Mock provider generate failed: %v", err)
		}

		if response == "" {
			t.Fatal("Empty response from mock provider")
		}
	})

	// Test provider-specific options
	t.Run("ProviderSpecificOptions", func(t *testing.T) {
		// Create safety settings for testing
		safetySettings := []map[string]interface{}{
			{
				"category":  "HARM_CATEGORY_DANGEROUS",
				"threshold": "BLOCK_LOW_AND_ABOVE",
			},
		}

		// Create a GeminiSafetySettingsOption
		safetyOption := domain.NewGeminiSafetySettingsOption(safetySettings)

		// Create a mock provider with specific options
		mockProvider := provider.NewMockProvider(safetyOption)

		// Verify the provider is created correctly
		if mockProvider == nil {
			t.Fatal("Failed to create mock provider")
		}

		// Test generation
		response, err := mockProvider.Generate(ctx, "Hello")
		if err != nil {
			t.Fatalf("Mock provider generate failed: %v", err)
		}

		if response == "" {
			t.Fatal("Empty response from mock provider")
		}
	})

	// Test ModelConfig with Options field
	t.Run("ModelConfigWithOptions", func(t *testing.T) {
		// Create provider-specific options
		orgOption := domain.NewOpenAIOrganizationOption("test-org")
		timeoutOption := domain.NewTimeoutOption(15)

		// Create a ModelConfig with explicit options
		config := llmutil.ModelConfig{
			Provider:  "mock", // Use mock for testing
			Model:     "mock-model",
			MaxTokens: 100,
			Options:   []domain.ProviderOption{orgOption, timeoutOption},
		}

		// Create provider from config
		llmProvider, err := llmutil.CreateProvider(config)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		// Verify the provider is created correctly
		if llmProvider == nil {
			t.Fatal("Failed to create provider from ModelConfig")
		}

		// Test generation
		response, err := llmProvider.Generate(ctx, "Hello")
		if err != nil {
			t.Fatalf("Provider generate failed: %v", err)
		}

		if response == "" {
			t.Fatal("Empty response from provider")
		}
	})

	// Test environment variable options
	t.Run("EnvironmentVariableOptions", func(t *testing.T) {
		// Save original environment variables
		origTimeout := os.Getenv("LLM_HTTP_TIMEOUT")
		origOpenAIKey := os.Getenv("OPENAI_API_KEY")
		origOpenAIOrg := os.Getenv("OPENAI_ORGANIZATION")
		origOpenAIUseCase := os.Getenv("OPENAI_USE_CASE")

		// Restore environment variables after test
		defer func() {
			os.Setenv("LLM_HTTP_TIMEOUT", origTimeout)
			os.Setenv("OPENAI_API_KEY", origOpenAIKey)
			os.Setenv("OPENAI_ORGANIZATION", origOpenAIOrg)
			os.Setenv("OPENAI_USE_CASE", origOpenAIUseCase)
		}()

		// Set test environment variables
		os.Setenv("LLM_HTTP_TIMEOUT", "15")
		os.Setenv("OPENAI_API_KEY", "test-key-from-env")
		os.Setenv("OPENAI_ORGANIZATION", "test-org-from-env")

		// Create a ModelConfig without explicit options
		config := llmutil.ModelConfig{
			Provider: "mock", // Use mock for testing
			Model:    "mock-model",
		}

		// Create provider from config (should use env vars)
		llmProvider, err := llmutil.CreateProvider(config)
		if err != nil {
			t.Fatalf("Error creating provider: %v", err)
		}

		// Verify the provider is created correctly
		if llmProvider == nil {
			t.Fatal("Failed to create provider from environment variables")
		}

		// Test generation
		response, err := llmProvider.Generate(ctx, "Hello")
		if err != nil {
			t.Fatalf("Provider generate failed: %v", err)
		}

		if response == "" {
			t.Fatal("Empty response from provider")
		}

		// Test with use case-specific environment variables
		os.Setenv("OPENAI_USE_CASE", "streaming")

		// Create a new config with the same parameters
		streamingConfig := llmutil.ModelConfig{
			Provider: "mock", // Use mock for testing
			Model:    "mock-model",
		}

		// Create provider from config (should use streaming options)
		streamingProvider, err := llmutil.CreateProvider(streamingConfig)
		if err != nil {
			t.Fatalf("Error creating streaming provider: %v", err)
		}

		// Verify the provider is created correctly
		if streamingProvider == nil {
			t.Fatal("Failed to create provider with use case from environment variables")
		}

		// Test generation with streaming provider
		streamingResponse, err := streamingProvider.Generate(ctx, "Hello")
		if err != nil {
			t.Fatalf("Streaming provider generate failed: %v", err)
		}

		if streamingResponse == "" {
			t.Fatal("Empty response from streaming provider")
		}
	})

	// Test option factory functions
	t.Run("OptionFactoryFunctions", func(t *testing.T) {
		// Test performance options
		performanceOptions := llmutil.WithPerformanceOptions()
		if len(performanceOptions) == 0 {
			t.Fatal("Expected non-empty performance options")
		}

		// Test OpenAI default options
		openAIOptions := llmutil.WithOpenAIDefaultOptions("test-org")
		if len(openAIOptions) == 0 {
			t.Fatal("Expected non-empty OpenAI options")
		}

		// Test streaming options
		streamingOptions := llmutil.WithStreamingOptions()
		if len(streamingOptions) == 0 {
			t.Fatal("Expected non-empty streaming options")
		}

		// Create a mock provider with factory options
		mockProvider := provider.NewMockProvider(performanceOptions...)

		// Verify the provider is created correctly
		if mockProvider == nil {
			t.Fatal("Failed to create mock provider with factory options")
		}

		// Test generation
		response, err := mockProvider.Generate(ctx, "Hello")
		if err != nil {
			t.Fatalf("Mock provider generate failed: %v", err)
		}

		if response == "" {
			t.Fatal("Empty response from mock provider")
		}
	})
}

func TestMain(m *testing.M) {
	// Run tests
	m.Run()
}
