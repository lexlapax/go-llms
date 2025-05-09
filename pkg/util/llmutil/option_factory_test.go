package llmutil

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

func TestCreateOptionFactoryFromEnvSimple(t *testing.T) {
	// Test that the factory creates options for different providers and use cases
	tests := []struct {
		name             string
		provider         string
		useCase          string
		expectedMinCount int
		expectedTypes    []string
	}{
		{
			name:             "OpenAI Default",
			provider:         "openai",
			useCase:          "default",
			expectedMinCount: 1,
			expectedTypes:    []string{"*domain.HeadersOption"},
		},
		{
			name:             "OpenAI Streaming",
			provider:         "openai",
			useCase:          "streaming",
			expectedMinCount: 3,
			expectedTypes:    []string{"*domain.HeadersOption", "*domain.HTTPClientOption", "*domain.TimeoutOption"},
		},
		{
			name:             "Anthropic Performance",
			provider:         "anthropic",
			useCase:          "performance",
			expectedMinCount: 3,
			expectedTypes:    []string{"*domain.HTTPClientOption", "*domain.TimeoutOption", "*domain.RetryOption"},
		},
		{
			name:             "Gemini Reliability",
			provider:         "gemini",
			useCase:          "reliability",
			expectedMinCount: 3,
			expectedTypes:    []string{"*domain.HTTPClientOption", "*domain.TimeoutOption", "*domain.RetryOption"},
		},
		{
			name:             "Unknown Provider",
			provider:         "unknown",
			useCase:          "default",
			expectedMinCount: 0,
			expectedTypes:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := CreateOptionFactoryFromEnv(tt.provider, tt.useCase)

			// Check minimum count
			if len(options) < tt.expectedMinCount {
				t.Errorf("Expected at least %d options, got %d", tt.expectedMinCount, len(options))
			}

			// Check that all expected types are present
			for _, typeName := range tt.expectedTypes {
				found := false
				for _, opt := range options {
					if nameOfType(opt) == typeName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected option type %s but not found", typeName)
				}
			}
		})
	}
}

// Helper function to get the name of a type
func nameOfType(v interface{}) string {
	switch v.(type) {
	case *domain.HTTPClientOption:
		return "*domain.HTTPClientOption"
	case *domain.TimeoutOption:
		return "*domain.TimeoutOption"
	case *domain.RetryOption:
		return "*domain.RetryOption"
	case *domain.HeadersOption:
		return "*domain.HeadersOption"
	case *domain.BaseURLOption:
		return "*domain.BaseURLOption"
	case *domain.OpenAIOrganizationOption:
		return "*domain.OpenAIOrganizationOption"
	case *domain.AnthropicSystemPromptOption:
		return "*domain.AnthropicSystemPromptOption"
	case *domain.GeminiGenerationConfigOption:
		return "*domain.GeminiGenerationConfigOption"
	case *domain.GeminiSafetySettingsOption:
		return "*domain.GeminiSafetySettingsOption"
	default:
		return "unknown"
	}
}
