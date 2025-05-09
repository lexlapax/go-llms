package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
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
}

func TestMain(m *testing.M) {
	// Run tests
	m.Run()
}