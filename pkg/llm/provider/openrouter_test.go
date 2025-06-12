package provider

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOpenRouterProvider(t *testing.T) {
	tests := []struct {
		name          string
		apiKey        string
		model         string
		envBase       string
		expectedBase  string
		expectBaseURL string
	}{
		{
			name:          "default configuration",
			apiKey:        "test-key",
			model:         "openai/gpt-4",
			envBase:       "",
			expectedBase:  defaultOpenRouterHost,
			expectBaseURL: "https://openrouter.ai/api",
		},
		{
			name:          "with custom base URL from env",
			apiKey:        "test-key",
			model:         "anthropic/claude-3-opus",
			envBase:       "https://custom.openrouter.ai/api",
			expectedBase:  "https://custom.openrouter.ai/api",
			expectBaseURL: "https://custom.openrouter.ai/api",
		},
		{
			name:          "with different model",
			apiKey:        "test-key",
			model:         "google/gemini-pro",
			envBase:       "",
			expectedBase:  defaultOpenRouterHost,
			expectBaseURL: "https://openrouter.ai/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if needed
			if tt.envBase != "" {
				_ = os.Setenv("OPENROUTER_API_BASE", tt.envBase)
				defer func() { _ = os.Unsetenv("OPENROUTER_API_BASE") }()
			}

			// Create provider
			provider := NewOpenRouterProvider(tt.apiKey, tt.model)

			// Verify it's an OpenAI provider
			assert.NotNil(t, provider)
			assert.Equal(t, tt.apiKey, provider.apiKey)
			assert.Equal(t, tt.model, provider.model)
			assert.Equal(t, tt.expectBaseURL, provider.baseURL)
		})
	}
}

func TestOpenRouterProviderIntegration(t *testing.T) {
	// Skip if no API key is provided
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenRouter integration test: OPENROUTER_API_KEY not set")
	}

	// Test with a free model
	provider := NewOpenRouterProvider(apiKey, "huggingface/zephyr-7b-beta:free")
	assert.NotNil(t, provider)

	// Note: Actual API calls would be tested in integration tests
}
