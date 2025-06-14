// ABOUTME: Anthropic provider metadata implementation with dynamic model loading
// ABOUTME: Provides capability discovery for Anthropic Claude, models fetched from modelinfo service

package provider

import "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo"

// AnthropicMetadata implements ProviderMetadata for Anthropic
type AnthropicMetadata struct {
	*BaseProviderMetadata
}

// NewAnthropicMetadata creates metadata for Anthropic provider
func NewAnthropicMetadata() *AnthropicMetadata {
	return NewAnthropicMetadataWithService(nil)
}

// NewAnthropicMetadataWithService creates metadata with a specific model service
func NewAnthropicMetadataWithService(modelService *modelinfo.ModelInfoService) *AnthropicMetadata {
	return &AnthropicMetadata{
		BaseProviderMetadata: NewBaseProviderMetadata(
			"Anthropic",
			"Anthropic Claude models focused on safety and capability",
			"anthropic", // providerType for filtering
			[]Capability{
				CapabilityStreaming,
				CapabilityVision,
				CapabilityStructuredOutput,
				// Note: Anthropic doesn't support function calling yet
			},
			Constraints{
				MaxConcurrency: 50,
				RateLimit: &RateLimit{
					RequestsPerMinute: 1000,
					TokensPerMinute:   100000,
				},
				RequiredHeaders: []string{"x-api-key", "anthropic-version"},
				MaxRetries:      3,
			},
			ConfigSchema{
				Version:     "1.0",
				Description: "Anthropic provider configuration",
				Fields: map[string]ConfigField{
					"api_key": {
						Name:        "api_key",
						Type:        "string",
						Description: "Anthropic API key",
						Required:    true,
						Secret:      true,
						EnvVar:      "ANTHROPIC_API_KEY",
					},
					"model": {
						Name:        "model",
						Type:        "string",
						Description: "Model to use",
						Required:    false,
						Default:     "claude-3-opus-20240229",
						// Options will be populated dynamically from GetModels
					},
					"max_tokens": {
						Name:        "max_tokens",
						Type:        "number",
						Description: "Maximum tokens to generate",
						Required:    false,
						Default:     4096,
					},
				},
			},
			modelService,
		),
	}
}
