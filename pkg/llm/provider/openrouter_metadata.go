// ABOUTME: OpenRouter provider metadata implementation with dynamic model loading
// ABOUTME: Provides capability discovery for OpenRouter aggregated models, fetched from modelinfo service

package provider

import "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo"

// OpenRouterMetadata implements ProviderMetadata for OpenRouter
type OpenRouterMetadata struct {
	*BaseProviderMetadata
}

// NewOpenRouterMetadata creates metadata for OpenRouter provider
func NewOpenRouterMetadata() *OpenRouterMetadata {
	return NewOpenRouterMetadataWithService(nil)
}

// NewOpenRouterMetadataWithService creates metadata with a specific model service
func NewOpenRouterMetadataWithService(modelService *modelinfo.ModelInfoService) *OpenRouterMetadata {
	return &OpenRouterMetadata{
		BaseProviderMetadata: NewBaseProviderMetadata(
			"OpenRouter",
			"Unified API for multiple LLM providers",
			"openrouter", // providerType for filtering
			[]Capability{
				CapabilityStreaming,
				CapabilityFunctionCalling,
				CapabilityVision,
				CapabilityStructuredOutput,
				// Note: Actual capabilities depend on the underlying model
			},
			Constraints{
				MaxConcurrency: 200, // High concurrency as it's an aggregator
				RateLimit: &RateLimit{
					RequestsPerMinute: 600,
					TokensPerMinute:   0, // No global token limit
				},
				RequiredHeaders: []string{"Authorization", "X-Title"},
				MaxRetries:      3,
			},
			ConfigSchema{
				Version:     "1.0",
				Description: "Configuration for OpenRouter provider",
				Fields: map[string]ConfigField{
					"api_key": {
						Name:        "api_key",
						Type:        "string",
						Description: "OpenRouter API key",
						Required:    true,
						Secret:      true,
						EnvVar:      "OPENROUTER_API_KEY",
					},
					"base_url": {
						Name:        "base_url",
						Type:        "string",
						Description: "OpenRouter API base URL",
						Required:    false,
						Default:     "https://openrouter.ai/api/v1",
					},
					"site_url": {
						Name:        "site_url",
						Type:        "string",
						Description: "Your site URL (for rankings)",
						Required:    false,
						EnvVar:      "OPENROUTER_SITE_URL",
					},
					"site_name": {
						Name:        "site_name",
						Type:        "string",
						Description: "Your site name (shown in rankings)",
						Required:    false,
						EnvVar:      "OPENROUTER_SITE_NAME",
					},
					"model": {
						Name:        "model",
						Type:        "string",
						Description: "Model to use (e.g., 'openai/gpt-4')",
						Required:    false,
						Default:     "auto", // OpenRouter can auto-select
						// Options will be populated dynamically from GetModels
					},
					"transforms": {
						Name:        "transforms",
						Type:        "array",
						Description: "Request transformations",
						Required:    false,
						Options:     []interface{}{"middle-out"}, // For better streaming
					},
					"route": {
						Name:        "route",
						Type:        "string",
						Description: "Routing preference",
						Required:    false,
						Options:     []interface{}{"fallback", "cheapest"},
						Default:     "fallback",
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
