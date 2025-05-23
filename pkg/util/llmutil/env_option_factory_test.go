package llmutil

import (
	"os"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func TestCreateProviderWithUseCase(t *testing.T) {
	// Save original environment variables
	origOpenAIKey := os.Getenv(EnvOpenAIAPIKey)
	origAnthropicKey := os.Getenv(EnvAnthropicAPIKey)
	origGeminiKey := os.Getenv(EnvGeminiAPIKey)
	origOpenAIUseCase := os.Getenv(EnvOpenAIUseCase)
	origAnthropicUseCase := os.Getenv(EnvAnthropicUseCase)
	origGeminiUseCase := os.Getenv(EnvGeminiUseCase)

	// Clean up environment after test
	defer func() {
		os.Setenv(EnvOpenAIAPIKey, origOpenAIKey)
		os.Setenv(EnvAnthropicAPIKey, origAnthropicKey)
		os.Setenv(EnvGeminiAPIKey, origGeminiKey)
		os.Setenv(EnvOpenAIUseCase, origOpenAIUseCase)
		os.Setenv(EnvAnthropicUseCase, origAnthropicUseCase)
		os.Setenv(EnvGeminiUseCase, origGeminiUseCase)
	}()

	// Clear environment variables
	envVars := []string{
		EnvOpenAIAPIKey, EnvAnthropicAPIKey, EnvGeminiAPIKey,
		EnvOpenAIUseCase, EnvAnthropicUseCase, EnvGeminiUseCase,
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	// Test creating provider with use case
	tests := []struct {
		name           string
		config         ModelConfig
		expectSuccess  bool
		expectedResult string
	}{
		{
			name: "OpenAI with streaming use case",
			config: ModelConfig{
				Provider: "openai",
				Model:    "gpt-4o",
				APIKey:   "test-api-key",
				UseCase:  "streaming",
			},
			expectSuccess:  true,
			expectedResult: "*provider.OpenAIProvider",
		},
		{
			name: "Anthropic with performance use case",
			config: ModelConfig{
				Provider: "anthropic",
				Model:    "claude-3-5-sonnet-latest",
				APIKey:   "test-api-key",
				UseCase:  "performance",
			},
			expectSuccess:  true,
			expectedResult: "*provider.AnthropicProvider",
		},
		{
			name: "Gemini with reliability use case",
			config: ModelConfig{
				Provider: "gemini",
				Model:    "gemini-2.0-flash-lite",
				APIKey:   "test-api-key",
				UseCase:  "reliability",
			},
			expectSuccess:  true,
			expectedResult: "*provider.GeminiProvider",
		},
		{
			name: "Unknown provider with use case",
			config: ModelConfig{
				Provider: "unknown",
				Model:    "unknown-model",
				APIKey:   "test-api-key",
				UseCase:  "default",
			},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := CreateProvider(tt.config)

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
					return
				}

				if provider == nil {
					t.Errorf("Expected provider but got nil")
					return
				}

				// Check the type of the provider
				typeName := typeName(provider)
				if typeName != tt.expectedResult {
					t.Errorf("Expected provider type %s but got %s", tt.expectedResult, typeName)
				}

				// For use case tests, we just verify the provider was created successfully
				// We can't check internal option state directly
			} else {
				if err == nil {
					t.Errorf("Expected error but got success")
				}
			}
		})
	}
}

// Helper function to get type name as string
func typeName(v interface{}) string {
	return "*provider." + getProviderType(v)
}

// Helper function to extract provider type name
func getProviderType(v interface{}) string {
	switch v.(type) {
	case *provider.OpenAIProvider:
		return "OpenAIProvider"
	case *provider.AnthropicProvider:
		return "AnthropicProvider"
	case *provider.GeminiProvider:
		return "GeminiProvider"
	case *provider.MockProvider:
		return "MockProvider"
	default:
		return "UnknownProvider"
	}
}

// Helper function to extract provider options - this is just for testing
// This is an unused function but kept for future testing purposes, marked with _ prefix
