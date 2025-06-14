package provider

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// mockListener implements RegistryListener for testing
type mockListener struct {
	registered   []string
	unregistered []string
	updated      []string
}

func (l *mockListener) OnProviderRegistered(name string, provider domain.Provider) {
	l.registered = append(l.registered, name)
}

func (l *mockListener) OnProviderUnregistered(name string) {
	l.unregistered = append(l.unregistered, name)
}

func (l *mockListener) OnProviderUpdated(name string, provider domain.Provider) {
	l.updated = append(l.updated, name)
}

func TestDynamicRegistry(t *testing.T) {
	t.Run("Basic Registration", func(t *testing.T) {
		registry := NewDynamicRegistry()

		// Register a provider
		mockProvider := NewMockProvider()
		err := registry.RegisterProvider("test", mockProvider, nil)
		assert.NoError(t, err)

		// Get the provider
		p, err := registry.GetProvider("test")
		assert.NoError(t, err)
		assert.Equal(t, mockProvider, p)

		// List providers
		providers := registry.ListProviders()
		assert.Contains(t, providers, "test")

		// Unregister
		err = registry.UnregisterProvider("test")
		assert.NoError(t, err)

		// Should not find after unregister
		_, err = registry.GetProvider("test")
		assert.Error(t, err)
	})

	t.Run("Factory Registration", func(t *testing.T) {
		registry := NewDynamicRegistry()
		factory := &MockFactory{}

		// Register factory
		err := registry.RegisterFactory("mock", factory)
		assert.NoError(t, err)

		// Get template
		template, err := registry.GetTemplate("mock")
		assert.NoError(t, err)
		assert.Equal(t, "mock", template.Type)
		assert.Equal(t, "Mock Provider", template.Name)

		// List templates
		templates := registry.ListTemplates()
		assert.Len(t, templates, 1)
	})

	t.Run("Create Provider From Template", func(t *testing.T) {
		registry := NewDynamicRegistry()
		factory := &MockFactory{}

		err := registry.RegisterFactory("mock", factory)
		require.NoError(t, err)

		// Create provider
		config := map[string]interface{}{
			"default_response": "test response",
		}

		err = registry.CreateProviderFromTemplate("mock", "test-mock", config)
		assert.NoError(t, err)

		// Verify provider exists
		p, err := registry.GetProvider("test-mock")
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("Provider With Metadata", func(t *testing.T) {
		registry := NewDynamicRegistry()

		mockProvider := NewMockProvider()
		metadata := &BaseProviderMetadata{
			ProviderName: "Test",
			Capabilities: []Capability{CapabilityStreaming},
		}

		err := registry.RegisterProvider("test", mockProvider, metadata)
		assert.NoError(t, err)

		// Get metadata
		meta, err := registry.GetMetadata("test")
		assert.NoError(t, err)
		assert.Equal(t, "Test", meta.Name())

		// List by capability
		providers := registry.ListProvidersByCapability(CapabilityStreaming)
		assert.Contains(t, providers, "test")

		providers = registry.ListProvidersByCapability(CapabilityVision)
		assert.Empty(t, providers)
	})

	t.Run("Registry Listeners", func(t *testing.T) {
		registry := NewDynamicRegistry()
		listener := &mockListener{}

		registry.AddListener(listener)

		// Register provider
		mockProvider := NewMockProvider()
		err := registry.RegisterProvider("test", mockProvider, nil)
		assert.NoError(t, err)
		assert.Contains(t, listener.registered, "test")

		// Unregister
		err = registry.UnregisterProvider("test")
		assert.NoError(t, err)
		assert.Contains(t, listener.unregistered, "test")

		// Remove listener
		registry.RemoveListener(listener)

		// Further operations should not notify
		err = registry.RegisterProvider("test2", mockProvider, nil)
		assert.NoError(t, err)
		assert.NotContains(t, listener.registered, "test2")
	})

	t.Run("Update Provider", func(t *testing.T) {
		registry := NewDynamicRegistry()
		factory := &MockFactory{}
		listener := &mockListener{}

		registry.AddListener(listener)
		err := registry.RegisterFactory("mock", factory)
		require.NoError(t, err)

		// Create initial provider
		config := map[string]interface{}{
			"default_response": "initial",
		}
		err = registry.CreateProviderFromTemplate("mock", "test", config)
		assert.NoError(t, err)

		// Update provider
		newConfig := map[string]interface{}{
			"default_response": "updated",
		}
		err = registry.UpdateProvider("test", newConfig)
		assert.NoError(t, err)

		assert.Contains(t, listener.updated, "test")
	})

	t.Run("Export Import Config", func(t *testing.T) {
		registry := NewDynamicRegistry()
		factory := &MockFactory{}

		err := registry.RegisterFactory("mock", factory)
		require.NoError(t, err)

		// Create provider with config
		config := map[string]interface{}{
			"default_response": "test",
			"responses": map[string]interface{}{
				"hello": "world",
			},
		}
		err = registry.CreateProviderFromTemplate("mock", "test-provider", config)
		assert.NoError(t, err)

		// Export config
		exported, err := registry.ExportConfig()
		assert.NoError(t, err)
		assert.Contains(t, exported, "providers")
		assert.Contains(t, exported, "version")

		providers := exported["providers"].(map[string]interface{})
		assert.Contains(t, providers, "test-provider")

		// Create new registry and import
		newRegistry := NewDynamicRegistry()
		err = newRegistry.RegisterFactory("mock", factory)
		require.NoError(t, err)

		err = newRegistry.ImportConfig(exported)
		assert.NoError(t, err)

		// Verify provider exists in new registry
		p, err := newRegistry.GetProvider("test-provider")
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("ModelRegistry Interface", func(t *testing.T) {
		registry := NewDynamicRegistry()
		var modelRegistry domain.ModelRegistry = registry

		// Register model
		mockProvider := NewMockProvider()
		err := modelRegistry.RegisterModel("test-model", mockProvider)
		assert.NoError(t, err)

		// Get model
		p, err := modelRegistry.GetModel("test-model")
		assert.NoError(t, err)
		assert.Equal(t, mockProvider, p)

		// List models
		models := modelRegistry.ListModels()
		assert.Contains(t, models, "test-model")
	})

	t.Run("Invalid Operations", func(t *testing.T) {
		registry := NewDynamicRegistry()

		// Empty provider name
		err := registry.RegisterProvider("", NewMockProvider(), nil)
		assert.Error(t, err)

		// Nil provider
		err = registry.RegisterProvider("test", nil, nil)
		assert.Error(t, err)

		// Empty factory type
		err = registry.RegisterFactory("", &MockFactory{})
		assert.Error(t, err)

		// Nil factory
		err = registry.RegisterFactory("test", nil)
		assert.Error(t, err)

		// Get non-existent provider
		_, err = registry.GetProvider("non-existent")
		assert.Error(t, err)

		// Get metadata for provider without metadata
		err = registry.RegisterProvider("no-meta", NewMockProvider(), nil)
		require.NoError(t, err)
		_, err = registry.GetMetadata("no-meta")
		assert.Error(t, err)

		// Create from non-existent template
		err = registry.CreateProviderFromTemplate("non-existent", "test", nil)
		assert.Error(t, err)

		// Update non-existent provider
		err = registry.UpdateProvider("non-existent", nil)
		assert.Error(t, err)
	})
}

func TestProviderFactories(t *testing.T) {
	t.Run("OpenAI Factory", func(t *testing.T) {
		// Save and unset env var for test
		oldKey := os.Getenv("OPENAI_API_KEY")
		_ = os.Unsetenv("OPENAI_API_KEY")
		defer func() {
			if oldKey != "" {
				_ = os.Setenv("OPENAI_API_KEY", oldKey)
			}
		}()

		factory := &OpenAIFactory{}

		// Test template
		template := factory.GetTemplate()
		assert.Equal(t, "openai", template.Type)
		assert.Equal(t, "OpenAI", template.Name)
		assert.NotEmpty(t, template.Schema.Fields)

		// Validate config - should fail without API key
		config := map[string]interface{}{}
		err := factory.ValidateConfig(config)
		assert.Error(t, err)

		// Valid config with API key
		config["api_key"] = "test-key"
		err = factory.ValidateConfig(config)
		assert.NoError(t, err)

		// Create provider (will fail without real API key in actual use)
		p, err := factory.CreateProvider(config)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("Anthropic Factory", func(t *testing.T) {
		factory := &AnthropicFactory{}

		template := factory.GetTemplate()
		assert.Equal(t, "anthropic", template.Type)
		assert.Contains(t, template.Schema.Fields, "api_key")
		assert.Contains(t, template.Schema.Fields, "model")

		// Test with config
		config := map[string]interface{}{
			"api_key": "test-key",
			"model":   "claude-3-opus-20240229",
		}

		err := factory.ValidateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("Mock Factory", func(t *testing.T) {
		factory := &MockFactory{}

		template := factory.GetTemplate()
		assert.Equal(t, "mock", template.Type)

		// Create with responses
		config := map[string]interface{}{
			"default_response": "default",
			"responses": map[string]interface{}{
				"hello": "world",
				"test":  "response",
			},
		}

		err := factory.ValidateConfig(config)
		assert.NoError(t, err)

		p, err := factory.CreateProvider(config)
		assert.NoError(t, err)
		assert.NotNil(t, p)
	})
}

func TestRegisterDefaultFactories(t *testing.T) {
	registry := NewDynamicRegistry()

	err := RegisterDefaultFactories(registry)
	assert.NoError(t, err)

	// Verify all factories are registered
	templates := registry.ListTemplates()
	assert.Len(t, templates, 7) // openai, anthropic, gemini, ollama, openrouter, vertexai, mock

	// Verify each template
	for _, template := range templates {
		switch template.Type {
		case "openai":
			assert.Equal(t, "OpenAI", template.Name)
		case "anthropic":
			assert.Equal(t, "Anthropic", template.Name)
		case "gemini":
			assert.Equal(t, "Gemini", template.Name)
		case "ollama":
			assert.Equal(t, "Ollama", template.Name)
		case "openrouter":
			assert.Equal(t, "OpenRouter", template.Name)
		case "vertexai":
			assert.Equal(t, "VertexAI", template.Name)
		case "mock":
			assert.Equal(t, "Mock Provider", template.Name)
		default:
			t.Errorf("Unexpected template type: %s", template.Type)
		}
	}
}
