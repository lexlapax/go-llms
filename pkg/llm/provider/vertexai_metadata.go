// ABOUTME: Vertex AI provider metadata implementation with dynamic model loading
// ABOUTME: Provides capability discovery for Google Vertex AI, models fetched from modelinfo service

package provider

import "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo"

// VertexAIMetadata implements ProviderMetadata for Google Vertex AI
type VertexAIMetadata struct {
	*BaseProviderMetadata
}

// NewVertexAIMetadata creates metadata for Vertex AI provider
func NewVertexAIMetadata() *VertexAIMetadata {
	return NewVertexAIMetadataWithService(nil)
}

// NewVertexAIMetadataWithService creates metadata with a specific model service
func NewVertexAIMetadataWithService(modelService *modelinfo.ModelInfoService) *VertexAIMetadata {
	return &VertexAIMetadata{
		BaseProviderMetadata: NewBaseProviderMetadata(
			"VertexAI",
			"Google Vertex AI enterprise LLM platform",
			"google", // providerType for filtering (same as Gemini)
			[]Capability{
				CapabilityStreaming,
				CapabilityVision,
				CapabilityAudio,
				CapabilityVideo,
				CapabilityStructuredOutput,
				CapabilityFunctionCalling,
			},
			Constraints{
				MaxConcurrency: 50,
				RateLimit: &RateLimit{
					RequestsPerMinute: 300,
					TokensPerMinute:   2000000,
				},
				RequiredHeaders: []string{"Authorization"},
				AllowedRegions: []string{
					"us-central1", "us-east1", "us-west1", "us-west4",
					"europe-west1", "europe-west2", "europe-west4",
					"asia-east1", "asia-northeast1", "asia-southeast1",
				},
				MaxRetries: 3,
			},
			ConfigSchema{
				Version:     "1.0",
				Description: "Configuration for Google Vertex AI provider",
				Fields: map[string]ConfigField{
					"project_id": {
						Name:        "project_id",
						Type:        "string",
						Description: "Google Cloud project ID",
						Required:    true,
						EnvVar:      "GOOGLE_CLOUD_PROJECT",
					},
					"location": {
						Name:        "location",
						Type:        "string",
						Description: "Vertex AI location/region",
						Required:    false,
						Default:     "us-central1",
						Options: []interface{}{
							"us-central1", "us-east1", "us-west1", "us-west4",
							"europe-west1", "europe-west2", "europe-west4",
							"asia-east1", "asia-northeast1", "asia-southeast1",
						},
					},
					"credentials_json": {
						Name:        "credentials_json",
						Type:        "string",
						Description: "Service account credentials JSON",
						Required:    false,
						Secret:      true,
						EnvVar:      "GOOGLE_APPLICATION_CREDENTIALS",
					},
					"model": {
						Name:        "model",
						Type:        "string",
						Description: "Model to use",
						Required:    false,
						Default:     "gemini-1.0-pro",
						// Options will be populated dynamically from GetModels
					},
					"endpoint": {
						Name:        "endpoint",
						Type:        "string",
						Description: "Custom endpoint (for private endpoints)",
						Required:    false,
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
