package provider

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderMetadata(t *testing.T) {
	t.Run("BaseProviderMetadata", func(t *testing.T) {
		metadata := &BaseProviderMetadata{
			ProviderName:        "Test Provider",
			ProviderDescription: "A test provider",
			Capabilities:        []Capability{CapabilityStreaming, CapabilityVision},
			ProviderConstraints: Constraints{
				MaxConcurrency: 10,
			},
			Schema: ConfigSchema{
				Version: "1.0",
			},
			// Note: Models are now loaded dynamically via GetModels()
		}

		assert.Equal(t, "Test Provider", metadata.Name())
		assert.Equal(t, "A test provider", metadata.Description())
		assert.Len(t, metadata.GetCapabilities(), 2)
		// Note: GetModels() is now async, test removed
		assert.Equal(t, 10, metadata.GetConstraints().MaxConcurrency)
		assert.Equal(t, "1.0", metadata.GetConfigSchema().Version)
	})

	t.Run("HasCapability", func(t *testing.T) {
		metadata := &BaseProviderMetadata{
			Capabilities: []Capability{CapabilityStreaming, CapabilityVision},
		}

		assert.True(t, HasCapability(metadata, CapabilityStreaming))
		assert.True(t, HasCapability(metadata, CapabilityVision))
		assert.False(t, HasCapability(metadata, CapabilityFunctionCalling))
		assert.False(t, HasCapability(metadata, CapabilityEmbeddings))
	})

	t.Run("FindModelByID", func(t *testing.T) {
		// Create metadata with cached models for testing
		metadata := &BaseProviderMetadata{
			cachedModels: []ModelInfo{
				{ID: "model-1", Name: "Model 1"},
				{ID: "model-2", Name: "Model 2"},
			},
			cacheExpiry: time.Now().Add(time.Hour), // Cache is still valid
		}

		ctx := context.Background()
		model, found, err := FindModelByID(metadata, "model-1", ctx)
		require.NoError(t, err)
		assert.True(t, found)
		assert.NotNil(t, model)
		assert.Equal(t, "Model 1", model.Name)

		model, found, err = FindModelByID(metadata, "model-3", ctx)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, model)
	})

	t.Run("JSON Marshaling", func(t *testing.T) {
		metadata := &BaseProviderMetadata{
			ProviderName:        "Test",
			ProviderDescription: "Test provider",
			Capabilities:        []Capability{CapabilityStreaming},
			cachedModels: []ModelInfo{
				{
					ID:            "test-1",
					Name:          "Test Model",
					ContextWindow: 2048,
					InputPricing: &PricingInfo{
						Currency:  "USD",
						PerTokens: 1000,
						Price:     0.01,
					},
				},
			},
			cacheExpiry: time.Now().Add(time.Hour), // Cache is still valid
		}

		data, err := json.Marshal(metadata)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, "Test", result["name"])
		assert.Equal(t, "Test provider", result["description"])

		models := result["models"].([]interface{})
		assert.Len(t, models, 1)
	})
}

func TestOpenAIMetadata(t *testing.T) {
	metadata := NewOpenAIMetadata()

	t.Run("Provider Info", func(t *testing.T) {
		assert.Equal(t, "OpenAI", metadata.Name())
		assert.Contains(t, metadata.Description(), "OpenAI GPT models")
	})

	t.Run("Capabilities", func(t *testing.T) {
		caps := metadata.GetCapabilities()
		assert.Contains(t, caps, CapabilityStreaming)
		assert.Contains(t, caps, CapabilityFunctionCalling)
		assert.Contains(t, caps, CapabilityVision)
		assert.Contains(t, caps, CapabilityStructuredOutput)
	})

	t.Run("Models", func(t *testing.T) {
		ctx := context.Background()
		// Note: Models are now fetched dynamically from modelinfo service
		// Since we don't have a model service configured in tests, this will return empty
		models, err := metadata.GetModels(ctx)
		require.NoError(t, err)
		// Should return empty slice when no model service is configured
		assert.Empty(t, models)
	})

	t.Run("Constraints", func(t *testing.T) {
		constraints := metadata.GetConstraints()
		assert.NotNil(t, constraints.RateLimit)
		assert.Greater(t, constraints.RateLimit.RequestsPerMinute, 0)
		assert.Contains(t, constraints.RequiredHeaders, "Authorization")
	})

	t.Run("Config Schema", func(t *testing.T) {
		schema := metadata.GetConfigSchema()
		assert.NotEmpty(t, schema.Fields)

		apiKey, exists := schema.Fields["api_key"]
		assert.True(t, exists)
		assert.True(t, apiKey.Required)
		assert.True(t, apiKey.Secret)
		assert.Equal(t, "OPENAI_API_KEY", apiKey.EnvVar)
	})
}

func TestAnthropicMetadata(t *testing.T) {
	metadata := NewAnthropicMetadata()

	t.Run("Provider Info", func(t *testing.T) {
		assert.Equal(t, "Anthropic", metadata.Name())
		assert.Contains(t, metadata.Description(), "Claude")
	})

	t.Run("Capabilities", func(t *testing.T) {
		caps := metadata.GetCapabilities()
		assert.Contains(t, caps, CapabilityStreaming)
		assert.Contains(t, caps, CapabilityVision)
		assert.NotContains(t, caps, CapabilityFunctionCalling) // Claude doesn't support function calling
	})

	t.Run("Models", func(t *testing.T) {
		ctx := context.Background()
		// Note: Models are now fetched dynamically from modelinfo service
		// Since we don't have a model service configured in tests, this will return empty
		models, err := metadata.GetModels(ctx)
		require.NoError(t, err)
		// Should return empty slice when no model service is configured
		assert.Empty(t, models)
	})

	t.Run("Config Schema", func(t *testing.T) {
		schema := metadata.GetConfigSchema()

		model, exists := schema.Fields["model"]
		assert.True(t, exists)
		// Options are now populated dynamically, not statically
		assert.Equal(t, "claude-3-opus-20240229", model.Default)
	})
}

func TestConfigField(t *testing.T) {
	t.Run("Basic Field", func(t *testing.T) {
		field := ConfigField{
			Name:        "test",
			Type:        "string",
			Description: "Test field",
			Required:    true,
			Default:     "default",
			EnvVar:      "TEST_VAR",
		}

		data, err := json.Marshal(field)
		require.NoError(t, err)

		var result ConfigField
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, field.Name, result.Name)
		assert.Equal(t, field.Type, result.Type)
		assert.Equal(t, field.Required, result.Required)
		assert.Equal(t, field.Default, result.Default)
		assert.Equal(t, field.EnvVar, result.EnvVar)
	})

	t.Run("Object Field with Properties", func(t *testing.T) {
		field := ConfigField{
			Name:        "config",
			Type:        "object",
			Description: "Configuration object",
			Properties: map[string]ConfigField{
				"host": {
					Name:     "host",
					Type:     "string",
					Required: true,
				},
				"port": {
					Name:    "port",
					Type:    "number",
					Default: 8080,
				},
			},
		}

		assert.Len(t, field.Properties, 2)
		assert.True(t, field.Properties["host"].Required)
		assert.Equal(t, 8080, field.Properties["port"].Default)
	})
}

func TestModelInfo(t *testing.T) {
	t.Run("Complete Model Info", func(t *testing.T) {
		model := ModelInfo{
			ID:            "test-model",
			Name:          "Test Model",
			Description:   "A test model",
			Capabilities:  []Capability{CapabilityStreaming, CapabilityVision},
			ContextWindow: 32768,
			MaxTokens:     4096,
			InputPricing: &PricingInfo{
				Currency:  "USD",
				PerTokens: 1000,
				Price:     0.01,
			},
			OutputPricing: &PricingInfo{
				Currency:  "USD",
				PerTokens: 1000,
				Price:     0.02,
			},
			Deprecated: false,
			Metadata: map[string]interface{}{
				"version": "1.0",
			},
		}

		data, err := json.Marshal(model)
		require.NoError(t, err)

		var result ModelInfo
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, model.ID, result.ID)
		assert.Equal(t, model.Name, result.Name)
		assert.Len(t, result.Capabilities, 2)
		assert.Equal(t, model.ContextWindow, result.ContextWindow)
		assert.NotNil(t, result.InputPricing)
		assert.Equal(t, 0.01, result.InputPricing.Price)
	})

	t.Run("Deprecated Model", func(t *testing.T) {
		deprecationDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		model := ModelInfo{
			ID:              "old-model",
			Name:            "Old Model",
			Deprecated:      true,
			DeprecationDate: &deprecationDate,
		}

		assert.True(t, model.Deprecated)
		assert.NotNil(t, model.DeprecationDate)
		assert.Equal(t, 2024, model.DeprecationDate.Year())
	})
}

func TestGeminiMetadata(t *testing.T) {
	metadata := NewGeminiMetadata()

	t.Run("Provider Info", func(t *testing.T) {
		assert.Equal(t, "Gemini", metadata.Name())
		assert.Contains(t, metadata.Description(), "Google Gemini")
	})

	t.Run("Capabilities", func(t *testing.T) {
		caps := metadata.GetCapabilities()
		assert.Contains(t, caps, CapabilityStreaming)
		assert.Contains(t, caps, CapabilityVision)
		assert.Contains(t, caps, CapabilityAudio)
		assert.Contains(t, caps, CapabilityVideo)
		assert.Contains(t, caps, CapabilityStructuredOutput)
		assert.Contains(t, caps, CapabilityFunctionCalling)
	})

	t.Run("Models", func(t *testing.T) {
		ctx := context.Background()
		// Note: Models are now fetched dynamically from modelinfo service
		// Since we don't have a model service configured in tests, this will return empty
		models, err := metadata.GetModels(ctx)
		require.NoError(t, err)
		// Should return empty slice when no model service is configured
		assert.Empty(t, models)
	})

	t.Run("Config Schema", func(t *testing.T) {
		schema := metadata.GetConfigSchema()

		apiKey, exists := schema.Fields["api_key"]
		assert.True(t, exists)
		assert.True(t, apiKey.Required)
		assert.True(t, apiKey.Secret)
		assert.Equal(t, "GOOGLE_API_KEY", apiKey.EnvVar)

		safety, exists := schema.Fields["safety_settings"]
		assert.True(t, exists)
		assert.Equal(t, "object", safety.Type)
		assert.NotEmpty(t, safety.Properties)
	})
}

func TestOllamaMetadata(t *testing.T) {
	metadata := NewOllamaMetadata()

	t.Run("Provider Info", func(t *testing.T) {
		assert.Equal(t, "Ollama", metadata.Name())
		assert.Contains(t, metadata.Description(), "Local LLM")
	})

	t.Run("Capabilities", func(t *testing.T) {
		caps := metadata.GetCapabilities()
		assert.Contains(t, caps, CapabilityStreaming)
		assert.Contains(t, caps, CapabilityEmbeddings)
		// Ollama has limited capabilities compared to cloud providers
		assert.Len(t, caps, 2)
	})

	t.Run("Constraints", func(t *testing.T) {
		constraints := metadata.GetConstraints()
		assert.Nil(t, constraints.RateLimit)            // No rate limits for local models
		assert.Equal(t, 10, constraints.MaxConcurrency) // Limited concurrency
	})

	t.Run("Config Schema", func(t *testing.T) {
		schema := metadata.GetConfigSchema()

		baseURL, exists := schema.Fields["base_url"]
		assert.True(t, exists)
		assert.False(t, baseURL.Required)
		assert.Equal(t, "http://localhost:11434", baseURL.Default)
		assert.Equal(t, "OLLAMA_HOST", baseURL.EnvVar)

		numGPU, exists := schema.Fields["num_gpu"]
		assert.True(t, exists)
		assert.Equal(t, "number", numGPU.Type)
	})
}

func TestOpenRouterMetadata(t *testing.T) {
	metadata := NewOpenRouterMetadata()

	t.Run("Provider Info", func(t *testing.T) {
		assert.Equal(t, "OpenRouter", metadata.Name())
		assert.Contains(t, metadata.Description(), "Unified API")
	})

	t.Run("Capabilities", func(t *testing.T) {
		caps := metadata.GetCapabilities()
		// OpenRouter supports most capabilities as an aggregator
		assert.Contains(t, caps, CapabilityStreaming)
		assert.Contains(t, caps, CapabilityFunctionCalling)
		assert.Contains(t, caps, CapabilityVision)
		assert.Contains(t, caps, CapabilityStructuredOutput)
	})

	t.Run("Constraints", func(t *testing.T) {
		constraints := metadata.GetConstraints()
		assert.NotNil(t, constraints.RateLimit)
		assert.Equal(t, 600, constraints.RateLimit.RequestsPerMinute)
		assert.Equal(t, 200, constraints.MaxConcurrency) // High concurrency
		assert.Contains(t, constraints.RequiredHeaders, "X-Title")
	})

	t.Run("Config Schema", func(t *testing.T) {
		schema := metadata.GetConfigSchema()

		route, exists := schema.Fields["route"]
		assert.True(t, exists)
		assert.Contains(t, route.Options, "fallback")
		assert.Contains(t, route.Options, "cheapest")

		transforms, exists := schema.Fields["transforms"]
		assert.True(t, exists)
		assert.Equal(t, "array", transforms.Type)
	})
}

func TestVertexAIMetadata(t *testing.T) {
	metadata := NewVertexAIMetadata()

	t.Run("Provider Info", func(t *testing.T) {
		assert.Equal(t, "VertexAI", metadata.Name())
		assert.Contains(t, metadata.Description(), "enterprise")
	})

	t.Run("Capabilities", func(t *testing.T) {
		caps := metadata.GetCapabilities()
		// Vertex AI has same capabilities as Gemini
		assert.Contains(t, caps, CapabilityStreaming)
		assert.Contains(t, caps, CapabilityVision)
		assert.Contains(t, caps, CapabilityAudio)
		assert.Contains(t, caps, CapabilityVideo)
	})

	t.Run("Constraints", func(t *testing.T) {
		constraints := metadata.GetConstraints()
		assert.NotEmpty(t, constraints.AllowedRegions)
		assert.Contains(t, constraints.AllowedRegions, "us-central1")
		assert.Contains(t, constraints.AllowedRegions, "europe-west1")
	})

	t.Run("Config Schema", func(t *testing.T) {
		schema := metadata.GetConfigSchema()

		projectID, exists := schema.Fields["project_id"]
		assert.True(t, exists)
		assert.True(t, projectID.Required)
		assert.Equal(t, "GOOGLE_CLOUD_PROJECT", projectID.EnvVar)

		location, exists := schema.Fields["location"]
		assert.True(t, exists)
		assert.NotEmpty(t, location.Options)
		assert.Equal(t, "us-central1", location.Default)
	})
}
