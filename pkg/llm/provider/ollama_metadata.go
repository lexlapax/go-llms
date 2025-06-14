// ABOUTME: Ollama provider metadata implementation with dynamic model loading
// ABOUTME: Provides capability discovery for Ollama local models, fetched from modelinfo service

package provider

import "github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo"

// OllamaMetadata implements ProviderMetadata for Ollama
type OllamaMetadata struct {
	*BaseProviderMetadata
}

// NewOllamaMetadata creates metadata for Ollama provider
func NewOllamaMetadata() *OllamaMetadata {
	return NewOllamaMetadataWithService(nil)
}

// NewOllamaMetadataWithService creates metadata with a specific model service
func NewOllamaMetadataWithService(modelService *modelinfo.ModelInfoService) *OllamaMetadata {
	return &OllamaMetadata{
		BaseProviderMetadata: NewBaseProviderMetadata(
			"Ollama",
			"Local LLM models running via Ollama",
			"ollama", // providerType for filtering
			[]Capability{
				CapabilityStreaming,
				CapabilityEmbeddings,
				// Note: Other capabilities depend on the specific model
			},
			Constraints{
				MaxConcurrency: 10,  // Local inference is more limited
				RateLimit:      nil, // No rate limits for local models
				MaxRetries:     2,
			},
			ConfigSchema{
				Version:     "1.0",
				Description: "Configuration for Ollama local model provider",
				Fields: map[string]ConfigField{
					"base_url": {
						Name:        "base_url",
						Type:        "string",
						Description: "Ollama server URL",
						Required:    false,
						Default:     "http://localhost:11434",
						EnvVar:      "OLLAMA_HOST",
					},
					"model": {
						Name:        "model",
						Type:        "string",
						Description: "Model to use",
						Required:    false,
						Default:     "llama2",
						// Options will be populated dynamically from GetModels
					},
					"keep_alive": {
						Name:        "keep_alive",
						Type:        "string",
						Description: "How long to keep model loaded in memory",
						Required:    false,
						Default:     "5m",
						Validation:  "duration string (e.g., '5m', '1h')",
					},
					"num_ctx": {
						Name:        "num_ctx",
						Type:        "number",
						Description: "Context window size",
						Required:    false,
						Default:     2048,
					},
					"num_gpu": {
						Name:        "num_gpu",
						Type:        "number",
						Description: "Number of GPUs to use (0 = CPU only)",
						Required:    false,
						Default:     1,
					},
					"num_thread": {
						Name:        "num_thread",
						Type:        "number",
						Description: "Number of CPU threads",
						Required:    false,
						Default:     0, // 0 means auto-detect
					},
					"temperature": {
						Name:        "temperature",
						Type:        "number",
						Description: "Sampling temperature (0-2)",
						Required:    false,
						Default:     0.8,
						Validation:  ">=0 && <=2",
					},
				},
			},
			modelService,
		),
	}
}
