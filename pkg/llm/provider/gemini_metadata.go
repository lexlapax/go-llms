// ABOUTME: Gemini provider metadata implementation with dynamic model loading
// ABOUTME: Provides capability discovery for Google Gemini, models fetched from modelinfo service

package provider

import "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo"

// GeminiMetadata implements ProviderMetadata for Google Gemini
type GeminiMetadata struct {
	*BaseProviderMetadata
}

// NewGeminiMetadata creates metadata for Gemini provider
func NewGeminiMetadata() *GeminiMetadata {
	return NewGeminiMetadataWithService(nil)
}

// NewGeminiMetadataWithService creates metadata with a specific model service
func NewGeminiMetadataWithService(modelService *modelinfo.ModelInfoService) *GeminiMetadata {
	return &GeminiMetadata{
		BaseProviderMetadata: NewBaseProviderMetadata(
			"Gemini",
			"Google Gemini models with multimodal capabilities",
			"google", // providerType for filtering
			[]Capability{
				CapabilityStreaming,
				CapabilityVision,
				CapabilityAudio,
				CapabilityVideo,
				CapabilityStructuredOutput,
				CapabilityFunctionCalling,
			},
			Constraints{
				MaxConcurrency: 100,
				RateLimit: &RateLimit{
					RequestsPerMinute: 60,
					TokensPerMinute:   1000000,
				},
				RequiredHeaders: []string{"Authorization"},
				MaxRetries:      3,
			},
			ConfigSchema{
				Version:     "1.0",
				Description: "Configuration for Google Gemini provider",
				Fields: map[string]ConfigField{
					"api_key": {
						Name:        "api_key",
						Type:        "string",
						Description: "Google AI API key",
						Required:    true,
						Secret:      true,
						EnvVar:      "GOOGLE_API_KEY",
					},
					"model": {
						Name:        "model",
						Type:        "string",
						Description: "Model to use",
						Required:    false,
						Default:     "gemini-pro",
						// Options will be populated dynamically from GetModels
					},
					"safety_settings": {
						Name:        "safety_settings",
						Type:        "object",
						Description: "Safety settings for content filtering",
						Required:    false,
						Properties: map[string]ConfigField{
							"harassment": {
								Name:        "harassment",
								Type:        "string",
								Description: "Harassment filter level",
								Options:     []interface{}{"BLOCK_NONE", "BLOCK_LOW_AND_ABOVE", "BLOCK_MEDIUM_AND_ABOVE", "BLOCK_ONLY_HIGH"},
								Default:     "BLOCK_MEDIUM_AND_ABOVE",
							},
							"hate_speech": {
								Name:        "hate_speech",
								Type:        "string",
								Description: "Hate speech filter level",
								Options:     []interface{}{"BLOCK_NONE", "BLOCK_LOW_AND_ABOVE", "BLOCK_MEDIUM_AND_ABOVE", "BLOCK_ONLY_HIGH"},
								Default:     "BLOCK_MEDIUM_AND_ABOVE",
							},
							"sexually_explicit": {
								Name:        "sexually_explicit",
								Type:        "string",
								Description: "Sexually explicit content filter level",
								Options:     []interface{}{"BLOCK_NONE", "BLOCK_LOW_AND_ABOVE", "BLOCK_MEDIUM_AND_ABOVE", "BLOCK_ONLY_HIGH"},
								Default:     "BLOCK_MEDIUM_AND_ABOVE",
							},
							"dangerous_content": {
								Name:        "dangerous_content",
								Type:        "string",
								Description: "Dangerous content filter level",
								Options:     []interface{}{"BLOCK_NONE", "BLOCK_LOW_AND_ABOVE", "BLOCK_MEDIUM_AND_ABOVE", "BLOCK_ONLY_HIGH"},
								Default:     "BLOCK_MEDIUM_AND_ABOVE",
							},
						},
					},
					"temperature": {
						Name:        "temperature",
						Type:        "number",
						Description: "Sampling temperature (0-1)",
						Required:    false,
						Default:     0.7,
						Validation:  ">=0 && <=1",
					},
					"max_output_tokens": {
						Name:        "max_output_tokens",
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
