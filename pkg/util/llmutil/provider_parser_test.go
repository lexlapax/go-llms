// ABOUTME: Tests for provider string parser functionality
// ABOUTME: Validates parsing of various provider/model string formats and aliases

package llmutil

import (
	"os"
	"testing"
)

func TestParseProviderModelString(t *testing.T) {
	tests := []struct {
		name         string
		spec         string
		wantProvider string
		wantModel    string
		wantErr      bool
	}{
		// Basic provider/model format
		{"openai with model", "openai/gpt-4", "openai", "gpt-4", false},
		{"anthropic with model", "anthropic/claude-3-opus-20240229", "anthropic", "claude-3-opus-20240229", false},
		{"google with model", "google/gemini-1.5-pro", "google", "gemini-1.5-pro", false},

		// Provider normalization (no longer normalizes gemini to google)
		{"gemini provider normalization", "gemini/gemini-1.5-pro", "gemini", "gemini-1.5-pro", false},

		// Provider only
		{"openai provider only", "openai", "openai", "", false},
		{"anthropic provider only", "anthropic", "anthropic", "", false},
		{"google provider only", "google", "google", "", false},
		{"gemini provider only", "gemini", "gemini", "", false},

		// Model aliases
		{"gpt-4 alias", "gpt-4", "openai", "gpt-4", false},
		{"claude alias", "claude", "anthropic", "claude-3-7-sonnet-latest", false},
		// Note: "gemini" is now treated as a provider, not an alias

		// Model inference
		{"gpt model inference", "gpt-4o-mini", "openai", "gpt-4o-mini", false},
		{"claude model inference", "claude-3-haiku-20240307", "anthropic", "claude-3-haiku-20240307", false},
		{"gemini model inference", "gemini-2.0-flash", "gemini", "gemini-2.0-flash", false},

		// Special cases
		{"mock provider", "mock", "mock", "", false},
		{"empty spec", "", "", "", true},
		{"unknown spec", "unknown-model-xyz", "", "", true},

		// Turbo keyword detection
		{"turbo keyword", "gpt-3.5-turbo", "openai", "gpt-3.5-turbo", false},
		{"turbo in middle", "gpt-4-turbo", "openai", "gpt-4-turbo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, model, err := ParseProviderModelString(tt.spec)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProviderModelString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if provider != tt.wantProvider {
				t.Errorf("ParseProviderModelString() provider = %v, want %v", provider, tt.wantProvider)
			}

			if model != tt.wantModel {
				t.Errorf("ParseProviderModelString() model = %v, want %v", model, tt.wantModel)
			}
		})
	}
}

func TestParseProviderModelWithOptions(t *testing.T) {
	tests := []struct {
		name         string
		spec         string
		wantProvider string
		wantModel    string
		wantUseCase  string
		wantErr      bool
	}{
		{"basic with option", "openai/gpt-4:streaming", "openai", "gpt-4", "streaming", false},
		{"alias with option", "claude:reliability", "anthropic", "claude-3-7-sonnet-latest", "reliability", false},
		{"no option", "openai/gpt-4", "openai", "gpt-4", "", false},
		{"provider only with option", "openai:performance", "openai", "", "performance", false},
		{"invalid spec with option", "unknown:streaming", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, model, useCase, err := ParseProviderModelWithOptions(tt.spec)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseProviderModelWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if provider != tt.wantProvider {
				t.Errorf("ParseProviderModelWithOptions() provider = %v, want %v", provider, tt.wantProvider)
			}

			if model != tt.wantModel {
				t.Errorf("ParseProviderModelWithOptions() model = %v, want %v", model, tt.wantModel)
			}

			if useCase != tt.wantUseCase {
				t.Errorf("ParseProviderModelWithOptions() useCase = %v, want %v", useCase, tt.wantUseCase)
			}
		})
	}
}

func TestNewProviderFromString(t *testing.T) {
	// Set up mock API key for testing
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	// Ensure Anthropic key is not set
	oldAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	defer func() {
		if oldAnthropicKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", oldAnthropicKey)
		}
	}()

	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		{"mock provider", "mock", false},
		{"openai with key", "openai/gpt-4", false},
		{"missing api key", "anthropic/claude-3-opus-latest", true}, // Should fail without API key
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProviderFromString(tt.spec)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewProviderFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && provider == nil {
				t.Error("NewProviderFromString() returned nil provider without error")
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("normalizeProvider", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
		}{
			{"OpenAI", "openai"},
			{"ANTHROPIC", "anthropic"},
			{"gemini", "gemini"},
			{"Gemini", "gemini"},
			{"google", "google"},
		}

		for _, tt := range tests {
			if got := normalizeProvider(tt.input); got != tt.want {
				t.Errorf("normalizeProvider(%q) = %q, want %q", tt.input, got, tt.want)
			}
		}
	})

	t.Run("isKnownProvider", func(t *testing.T) {
		knownProviders := []string{"openai", "OpenAI", "anthropic", "google", "gemini", "mock"}
		for _, provider := range knownProviders {
			if !isKnownProvider(provider) {
				t.Errorf("isKnownProvider(%q) = false, want true", provider)
			}
		}

		unknownProviders := []string{"azure", "huggingface", "unknown"}
		for _, provider := range unknownProviders {
			if isKnownProvider(provider) {
				t.Errorf("isKnownProvider(%q) = true, want false", provider)
			}
		}
	})

	t.Run("getAPIKeyEnvVar", func(t *testing.T) {
		tests := []struct {
			provider string
			want     string
		}{
			{"openai", "OPENAI_API_KEY"},
			{"anthropic", "ANTHROPIC_API_KEY"},
			{"google", "GEMINI_API_KEY"},
			{"gemini", "GEMINI_API_KEY"},
			{"custom", "CUSTOM_API_KEY"},
		}

		for _, tt := range tests {
			if got := getAPIKeyEnvVar(tt.provider); got != tt.want {
				t.Errorf("getAPIKeyEnvVar(%q) = %q, want %q", tt.provider, got, tt.want)
			}
		}
	})
}
