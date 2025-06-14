// ABOUTME: OpenAI provider metadata implementation with dynamic model loading
// ABOUTME: Provides capability discovery for OpenAI, models fetched from modelinfo service

package provider

import "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo"

// OpenAIMetadata implements ProviderMetadata for OpenAI
type OpenAIMetadata struct {
	*BaseProviderMetadata
}

// NewOpenAIMetadata creates metadata for OpenAI provider
func NewOpenAIMetadata() *OpenAIMetadata {
	return NewOpenAIMetadataWithService(nil)
}

// NewOpenAIMetadataWithService creates metadata with a specific model service
func NewOpenAIMetadataWithService(modelService *modelinfo.ModelInfoService) *OpenAIMetadata {
	return &OpenAIMetadata{
		BaseProviderMetadata: NewBaseProviderMetadata(
			"OpenAI",
			"OpenAI GPT models including GPT-4, GPT-3.5-turbo, and specialized models",
			"openai", // providerType for filtering
			[]Capability{
				CapabilityStreaming,
				CapabilityFunctionCalling,
				CapabilityVision,
				CapabilityStructuredOutput,
			},
			Constraints{
				MaxConcurrency: 100,
				RateLimit: &RateLimit{
					RequestsPerMinute: 3500,
					TokensPerMinute:   350000,
				},
				RequiredHeaders: []string{"Authorization"},
				MaxRetries:      3,
			},
			ConfigSchema{
				Version:     "1.0",
				Description: "Configuration for OpenAI provider",
				Fields: map[string]ConfigField{
					"api_key": {
						Name:        "api_key",
						Type:        "string",
						Description: "OpenAI API key",
						Required:    true,
						Secret:      true,
						EnvVar:      "OPENAI_API_KEY",
					},
					"organization": {
						Name:        "organization",
						Type:        "string",
						Description: "OpenAI organization ID",
						Required:    false,
						EnvVar:      "OPENAI_ORGANIZATION",
					},
					"base_url": {
						Name:        "base_url",
						Type:        "string",
						Description: "Base URL for API (for custom endpoints)",
						Required:    false,
						Default:     "https://api.openai.com",
					},
					"model": {
						Name:        "model",
						Type:        "string",
						Description: "Default model to use",
						Required:    false,
						Default:     "gpt-4",
						// Options will be populated dynamically from GetModels
					},
					"temperature": {
						Name:        "temperature",
						Type:        "number",
						Description: "Sampling temperature (0-2)",
						Required:    false,
						Default:     0.7,
						Validation:  ">=0 && <=2",
					},
					"max_tokens": {
						Name:        "max_tokens",
						Type:        "number",
						Description: "Maximum tokens to generate",
						Required:    false,
						Default:     2048,
					},
				},
			},
			modelService,
		),
	}
}
