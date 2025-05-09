package llmutil

import (
	"os"
	"testing"
)

func TestGetAPIKeyFromEnv(t *testing.T) {
	// Save original environment variables
	origOpenAIKey := os.Getenv(EnvOpenAIAPIKey)
	origAnthropicKey := os.Getenv(EnvAnthropicAPIKey)
	origGeminiKey := os.Getenv(EnvGeminiAPIKey)

	// Clean up environment after test
	defer func() {
		os.Setenv(EnvOpenAIAPIKey, origOpenAIKey)
		os.Setenv(EnvAnthropicAPIKey, origAnthropicKey)
		os.Setenv(EnvGeminiAPIKey, origGeminiKey)
	}()

	tests := []struct {
		name        string
		provider    string
		envVars     map[string]string
		expectedKey string
	}{
		{
			name:     "OpenAI Key",
			provider: "openai",
			envVars: map[string]string{
				EnvOpenAIAPIKey: "test-openai-key",
			},
			expectedKey: "test-openai-key",
		},
		{
			name:     "Anthropic Key",
			provider: "anthropic",
			envVars: map[string]string{
				EnvAnthropicAPIKey: "test-anthropic-key",
			},
			expectedKey: "test-anthropic-key",
		},
		{
			name:     "Gemini Key",
			provider: "gemini",
			envVars: map[string]string{
				EnvGeminiAPIKey: "test-gemini-key",
			},
			expectedKey: "test-gemini-key",
		},
		{
			name:        "No Key",
			provider:    "openai",
			envVars:     map[string]string{},
			expectedKey: "",
		},
		{
			name:        "Unknown Provider",
			provider:    "unknown",
			envVars:     map[string]string{},
			expectedKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables first
			os.Unsetenv(EnvOpenAIAPIKey)
			os.Unsetenv(EnvAnthropicAPIKey)
			os.Unsetenv(EnvGeminiAPIKey)

			// Set environment variables for test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Call function
			key := GetAPIKeyFromEnv(tt.provider)

			// Check result
			if key != tt.expectedKey {
				t.Errorf("GetAPIKeyFromEnv(%s) = %s, want %s", tt.provider, key, tt.expectedKey)
			}
		})
	}
}

func TestGetModelFromEnv(t *testing.T) {
	// Save original environment variables
	origOpenAIModel := os.Getenv(EnvOpenAIModel)
	origAnthropicModel := os.Getenv(EnvAnthropicModel)
	origGeminiModel := os.Getenv(EnvGeminiModel)

	// Clean up environment after test
	defer func() {
		os.Setenv(EnvOpenAIModel, origOpenAIModel)
		os.Setenv(EnvAnthropicModel, origAnthropicModel)
		os.Setenv(EnvGeminiModel, origGeminiModel)
	}()

	tests := []struct {
		name          string
		provider      string
		envVars       map[string]string
		expectedModel string
	}{
		{
			name:     "OpenAI Model",
			provider: "openai",
			envVars: map[string]string{
				EnvOpenAIModel: "gpt-4",
			},
			expectedModel: "gpt-4",
		},
		{
			name:          "OpenAI Default Model",
			provider:      "openai",
			envVars:       map[string]string{},
			expectedModel: "gpt-4o",
		},
		{
			name:     "Anthropic Model",
			provider: "anthropic",
			envVars: map[string]string{
				EnvAnthropicModel: "claude-3-opus-20240229",
			},
			expectedModel: "claude-3-opus-20240229",
		},
		{
			name:          "Anthropic Default Model",
			provider:      "anthropic",
			envVars:       map[string]string{},
			expectedModel: "claude-3-5-sonnet-latest",
		},
		{
			name:     "Gemini Model",
			provider: "gemini",
			envVars: map[string]string{
				EnvGeminiModel: "gemini-1.5-pro",
			},
			expectedModel: "gemini-1.5-pro",
		},
		{
			name:          "Gemini Default Model",
			provider:      "gemini",
			envVars:       map[string]string{},
			expectedModel: "gemini-2.0-flash-lite",
		},
		{
			name:          "Unknown Provider",
			provider:      "unknown",
			envVars:       map[string]string{},
			expectedModel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables first
			os.Unsetenv(EnvOpenAIModel)
			os.Unsetenv(EnvAnthropicModel)
			os.Unsetenv(EnvGeminiModel)

			// Set environment variables for test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Call function
			model := GetModelFromEnv(tt.provider)

			// Check result
			if model != tt.expectedModel {
				t.Errorf("GetModelFromEnv(%s) = %s, want %s", tt.provider, model, tt.expectedModel)
			}
		})
	}
}

func TestGetCommonOptionsFromEnv(t *testing.T) {
	// Save original environment variables
	origHTTPTimeout := os.Getenv(EnvHTTPTimeout)
	origRetryAttempts := os.Getenv(EnvRetryAttempts)
	origRetryDelay := os.Getenv(EnvRetryDelay)

	// Clean up environment after test
	defer func() {
		os.Setenv(EnvHTTPTimeout, origHTTPTimeout)
		os.Setenv(EnvRetryAttempts, origRetryAttempts)
		os.Setenv(EnvRetryDelay, origRetryDelay)
	}()

	tests := []struct {
		name          string
		envVars       map[string]string
		expectedCount int
	}{
		{
			name:          "No Options",
			envVars:       map[string]string{},
			expectedCount: 0,
		},
		{
			name: "HTTP Timeout Only",
			envVars: map[string]string{
				EnvHTTPTimeout: "10",
			},
			expectedCount: 1, // Just the timeout option
		},
		{
			name: "Retry Settings",
			envVars: map[string]string{
				EnvRetryAttempts: "3",
				EnvRetryDelay:    "500",
			},
			expectedCount: 1, // One retry option with both settings
		},
		{
			name: "All Options",
			envVars: map[string]string{
				EnvHTTPTimeout:   "10",
				EnvRetryAttempts: "3",
				EnvRetryDelay:    "500",
			},
			expectedCount: 2, // Timeout + retry options
		},
		{
			name: "Invalid Timeout",
			envVars: map[string]string{
				EnvHTTPTimeout: "not-a-number",
			},
			expectedCount: 0, // Invalid timeout should be ignored
		},
		{
			name: "Invalid Retry Attempts",
			envVars: map[string]string{
				EnvRetryAttempts: "not-a-number",
				EnvRetryDelay:    "500",
			},
			expectedCount: 0, // Invalid retry should be ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables first
			os.Unsetenv(EnvHTTPTimeout)
			os.Unsetenv(EnvRetryAttempts)
			os.Unsetenv(EnvRetryDelay)

			// Set environment variables for test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Call function
			options := GetCommonOptionsFromEnv()

			// Check result count
			if len(options) != tt.expectedCount {
				t.Errorf("GetCommonOptionsFromEnv() returned %d options, want %d", len(options), tt.expectedCount)
			}

			// Additional type checks could be added here
		})
	}
}

func TestGetProviderOptionsFromEnv(t *testing.T) {
	// Save original environment variables
	origOpenAIBaseURL := os.Getenv(EnvOpenAIBaseURL)
	origAnthropicBaseURL := os.Getenv(EnvAnthropicBaseURL)
	origGeminiBaseURL := os.Getenv(EnvGeminiBaseURL)
	origHTTPTimeout := os.Getenv(EnvHTTPTimeout)
	origAnthropicSystemPrompt := os.Getenv(EnvAnthropicSystemPrompt)
	origOpenAIOrganization := os.Getenv(EnvOpenAIOrganization)

	// Clean up environment after test
	defer func() {
		os.Setenv(EnvOpenAIBaseURL, origOpenAIBaseURL)
		os.Setenv(EnvAnthropicBaseURL, origAnthropicBaseURL)
		os.Setenv(EnvGeminiBaseURL, origGeminiBaseURL)
		os.Setenv(EnvHTTPTimeout, origHTTPTimeout)
		os.Setenv(EnvAnthropicSystemPrompt, origAnthropicSystemPrompt)
		os.Setenv(EnvOpenAIOrganization, origOpenAIOrganization)
	}()

	tests := []struct {
		name          string
		provider      string
		envVars       map[string]string
		expectedCount int
	}{
		{
			name:          "OpenAI No Options",
			provider:      "openai",
			envVars:       map[string]string{},
			expectedCount: 0,
		},
		{
			name:     "OpenAI Base URL",
			provider: "openai",
			envVars: map[string]string{
				EnvOpenAIBaseURL: "https://custom.openai.api",
			},
			expectedCount: 1,
		},
		{
			name:     "OpenAI Base URL and Organization",
			provider: "openai",
			envVars: map[string]string{
				EnvOpenAIBaseURL:      "https://custom.openai.api",
				EnvOpenAIOrganization: "test-org",
			},
			expectedCount: 2,
		},
		{
			name:     "Anthropic Base URL and System Prompt",
			provider: "anthropic",
			envVars: map[string]string{
				EnvAnthropicBaseURL:      "https://custom.anthropic.api",
				EnvAnthropicSystemPrompt: "Test system prompt",
			},
			expectedCount: 2,
		},
		{
			name:     "Gemini Base URL",
			provider: "gemini",
			envVars: map[string]string{
				EnvGeminiBaseURL: "https://custom.gemini.api",
			},
			expectedCount: 1,
		},
		{
			name:     "Common Options",
			provider: "openai",
			envVars: map[string]string{
				EnvHTTPTimeout: "10",
			},
			expectedCount: 1,
		},
		{
			name:     "Unknown Provider",
			provider: "unknown",
			envVars: map[string]string{
				EnvHTTPTimeout: "10",
			},
			expectedCount: 1, // Should still get common options
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables first
			clearEnvVars := []string{
				EnvOpenAIBaseURL, EnvAnthropicBaseURL, EnvGeminiBaseURL,
				EnvHTTPTimeout, EnvAnthropicSystemPrompt, EnvOpenAIOrganization,
			}
			for _, v := range clearEnvVars {
				os.Unsetenv(v)
			}

			// Set environment variables for test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Call function
			options := GetProviderOptionsFromEnv(tt.provider)

			// Check result count
			if len(options) != tt.expectedCount {
				t.Errorf("GetProviderOptionsFromEnv(%s) returned %d options, want %d",
					tt.provider, len(options), tt.expectedCount)
			}
		})
	}
}
