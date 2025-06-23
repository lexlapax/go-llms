// ABOUTME: Provider factory implementations for common LLM providers
// ABOUTME: Enables dynamic provider creation from configuration

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// OpenAIFactory creates OpenAI provider instances from configuration.
// It supports API key from config or OPENAI_API_KEY environment variable.
type OpenAIFactory struct{}

// CreateProvider implements ProviderFactory
func (f *OpenAIFactory) CreateProvider(config map[string]interface{}) (domain.Provider, error) {
	apiKey, _ := config["api_key"].(string)
	if apiKey == "" {
		// Try environment variable
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	// Get model or use default
	model, _ := config["model"].(string)
	if model == "" {
		model = "gpt-4"
	}

	// Create provider options
	var providerOptions []domain.ProviderOption

	// Add OpenAI-specific options if available
	if org, ok := config["organization"].(string); ok {
		providerOptions = append(providerOptions, domain.NewOpenAIOrganizationOption(org))
	}
	if baseURL, ok := config["base_url"].(string); ok {
		providerOptions = append(providerOptions, domain.NewBaseURLOption(baseURL))
	}

	return NewOpenAIProvider(apiKey, model, providerOptions...), nil
}

// ValidateConfig implements ProviderFactory
func (f *OpenAIFactory) ValidateConfig(config map[string]interface{}) error {
	apiKey, _ := config["api_key"].(string)
	if apiKey == "" && os.Getenv("OPENAI_API_KEY") == "" {
		return fmt.Errorf("api_key is required (or set OPENAI_API_KEY environment variable)")
	}
	return nil
}

// GetTemplate implements ProviderFactory
func (f *OpenAIFactory) GetTemplate() ProviderTemplate {
	return ProviderTemplate{
		Type:        "openai",
		Name:        "OpenAI",
		Description: "OpenAI GPT models provider",
		Schema: ConfigSchema{
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
					Options:     []interface{}{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"},
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
			Examples: []ConfigExample{
				{
					Name:        "Basic configuration",
					Description: "Minimal configuration with API key",
					Config: map[string]interface{}{
						"api_key": "sk-...",
					},
				},
				{
					Name:        "Advanced configuration",
					Description: "Configuration with custom settings",
					Config: map[string]interface{}{
						"api_key":      "sk-...",
						"organization": "org-...",
						"model":        "gpt-4-turbo",
						"temperature":  0.5,
						"max_tokens":   4096,
					},
				},
			},
		},
		Defaults: map[string]interface{}{
			"model":       "gpt-4",
			"temperature": 0.7,
			"max_tokens":  2048,
		},
	}
}

// AnthropicFactory creates Anthropic provider instances from configuration.
// It supports API key from config or ANTHROPIC_API_KEY environment variable.
type AnthropicFactory struct{}

// CreateProvider implements ProviderFactory
func (f *AnthropicFactory) CreateProvider(config map[string]interface{}) (domain.Provider, error) {
	apiKey, _ := config["api_key"].(string)
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	// Get model or use default
	model, _ := config["model"].(string)
	if model == "" {
		model = "claude-3-opus-20240229"
	}

	// Create provider options
	var providerOptions []domain.ProviderOption

	return NewAnthropicProvider(apiKey, model, providerOptions...), nil
}

// ValidateConfig implements ProviderFactory
func (f *AnthropicFactory) ValidateConfig(config map[string]interface{}) error {
	apiKey, _ := config["api_key"].(string)
	if apiKey == "" && os.Getenv("ANTHROPIC_API_KEY") == "" {
		return fmt.Errorf("api_key is required (or set ANTHROPIC_API_KEY environment variable)")
	}
	return nil
}

// GetTemplate implements ProviderFactory
func (f *AnthropicFactory) GetTemplate() ProviderTemplate {
	return ProviderTemplate{
		Type:        "anthropic",
		Name:        "Anthropic",
		Description: "Anthropic Claude models provider",
		Schema: ConfigSchema{
			Version:     "1.0",
			Description: "Configuration for Anthropic provider",
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
					Description: "Default model to use",
					Required:    false,
					Default:     "claude-3-opus-20240229",
					Options:     []interface{}{"claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"},
				},
				"temperature": {
					Name:        "temperature",
					Type:        "number",
					Description: "Sampling temperature (0-1)",
					Required:    false,
					Default:     0.7,
					Validation:  ">=0 && <=1",
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
		Defaults: map[string]interface{}{
			"model":       "claude-3-opus-20240229",
			"temperature": 0.7,
			"max_tokens":  4096,
		},
	}
}

// MockFactory creates mock provider instances for testing.
// It supports configurable responses based on prompt patterns.
type MockFactory struct{}

// CreateProvider implements ProviderFactory
func (f *MockFactory) CreateProvider(config map[string]interface{}) (domain.Provider, error) {
	// Create a basic mock provider
	mockProvider := NewMockProvider()

	// If responses are provided, set them up
	if respMap, ok := config["responses"].(map[string]interface{}); ok {
		responses := make(map[string]string)
		for k, v := range respMap {
			if str, ok := v.(string); ok {
				responses[k] = str
			}
		}

		// Set up custom response function
		if len(responses) > 0 || config["default_response"] != nil {
			defaultResponse := "This is a mock response"
			if def, ok := config["default_response"].(string); ok {
				defaultResponse = def
			}

			mockProvider.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
				// Check if we have a specific response for this prompt
				for pattern, response := range responses {
					if strings.Contains(prompt, pattern) {
						return response, nil
					}
				}
				return defaultResponse, nil
			})
		}
	}

	return mockProvider, nil
}

// ValidateConfig implements ProviderFactory
func (f *MockFactory) ValidateConfig(config map[string]interface{}) error {
	// Mock provider has no required fields
	return nil
}

// GetTemplate implements ProviderFactory
func (f *MockFactory) GetTemplate() ProviderTemplate {
	return ProviderTemplate{
		Type:        "mock",
		Name:        "Mock Provider",
		Description: "Mock provider for testing",
		Schema: ConfigSchema{
			Version:     "1.0",
			Description: "Configuration for mock provider",
			Fields: map[string]ConfigField{
				"default_response": {
					Name:        "default_response",
					Type:        "string",
					Description: "Default response for unmatched prompts",
					Required:    false,
					Default:     "This is a mock response",
				},
				"responses": {
					Name:        "responses",
					Type:        "object",
					Description: "Map of prompt patterns to responses",
					Required:    false,
					Properties: map[string]ConfigField{
						"*": {
							Name:        "*",
							Type:        "string",
							Description: "Response for matching prompt",
						},
					},
				},
				"error": {
					Name:        "error",
					Type:        "string",
					Description: "Error message to return (for testing error handling)",
					Required:    false,
				},
			},
		},
		Defaults: map[string]interface{}{
			"default_response": "This is a mock response",
			"responses":        map[string]interface{}{},
		},
	}
}

// GeminiFactory creates Gemini provider instances from configuration.
// It supports API key from config or GOOGLE_API_KEY environment variable.
type GeminiFactory struct{}

// CreateProvider implements ProviderFactory
func (f *GeminiFactory) CreateProvider(config map[string]interface{}) (domain.Provider, error) {
	apiKey, _ := config["api_key"].(string)
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	// Get model or use default
	model, _ := config["model"].(string)
	if model == "" {
		model = "gemini-pro"
	}

	return NewGeminiProvider(apiKey, model), nil
}

// ValidateConfig implements ProviderFactory
func (f *GeminiFactory) ValidateConfig(config map[string]interface{}) error {
	apiKey, _ := config["api_key"].(string)
	if apiKey == "" && os.Getenv("GOOGLE_API_KEY") == "" {
		return fmt.Errorf("api_key is required (or set GOOGLE_API_KEY environment variable)")
	}
	return nil
}

// GetTemplate implements ProviderFactory
func (f *GeminiFactory) GetTemplate() ProviderTemplate {
	metadata := NewGeminiMetadata()
	return ProviderTemplate{
		Type:        "gemini",
		Name:        metadata.Name(),
		Description: metadata.Description(),
		Schema:      metadata.GetConfigSchema(),
		Defaults: map[string]interface{}{
			"model":       "gemini-pro",
			"temperature": 0.7,
		},
	}
}

// OllamaFactory creates Ollama provider instances from configuration.
// It supports base URL from config or OLLAMA_HOST environment variable.
type OllamaFactory struct{}

// CreateProvider implements ProviderFactory
func (f *OllamaFactory) CreateProvider(config map[string]interface{}) (domain.Provider, error) {
	// Get base URL or use default
	baseURL, _ := config["base_url"].(string)
	if baseURL == "" {
		baseURL = os.Getenv("OLLAMA_HOST")
	}
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	// Get model or use default
	model, _ := config["model"].(string)
	if model == "" {
		model = "llama2"
	}

	// Ollama provider doesn't use baseURL directly in constructor
	// It uses model and options
	var options []domain.ProviderOption
	if baseURL != "http://localhost:11434" {
		options = append(options, domain.NewBaseURLOption(baseURL))
	}
	return NewOllamaProvider(model, options...), nil
}

// ValidateConfig implements ProviderFactory
func (f *OllamaFactory) ValidateConfig(config map[string]interface{}) error {
	// Ollama doesn't require any specific config
	return nil
}

// GetTemplate implements ProviderFactory
func (f *OllamaFactory) GetTemplate() ProviderTemplate {
	metadata := NewOllamaMetadata()
	return ProviderTemplate{
		Type:        "ollama",
		Name:        metadata.Name(),
		Description: metadata.Description(),
		Schema:      metadata.GetConfigSchema(),
		Defaults: map[string]interface{}{
			"base_url": "http://localhost:11434",
			"model":    "llama2",
		},
	}
}

// OpenRouterFactory creates OpenRouter provider instances from configuration.
// It supports API key from config or OPENROUTER_API_KEY environment variable.
type OpenRouterFactory struct{}

// CreateProvider implements ProviderFactory
func (f *OpenRouterFactory) CreateProvider(config map[string]interface{}) (domain.Provider, error) {
	apiKey, _ := config["api_key"].(string)
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	// Get model or use default
	model, _ := config["model"].(string)
	if model == "" {
		model = "auto"
	}

	return NewOpenRouterProvider(apiKey, model), nil
}

// ValidateConfig implements ProviderFactory
func (f *OpenRouterFactory) ValidateConfig(config map[string]interface{}) error {
	apiKey, _ := config["api_key"].(string)
	if apiKey == "" && os.Getenv("OPENROUTER_API_KEY") == "" {
		return fmt.Errorf("api_key is required (or set OPENROUTER_API_KEY environment variable)")
	}
	return nil
}

// GetTemplate implements ProviderFactory
func (f *OpenRouterFactory) GetTemplate() ProviderTemplate {
	metadata := NewOpenRouterMetadata()
	return ProviderTemplate{
		Type:        "openrouter",
		Name:        metadata.Name(),
		Description: metadata.Description(),
		Schema:      metadata.GetConfigSchema(),
		Defaults: map[string]interface{}{
			"model":       "auto",
			"temperature": 0.7,
		},
	}
}

// VertexAIFactory creates Vertex AI provider instances from configuration.
// It requires a Google Cloud project ID and supports regional endpoints.
type VertexAIFactory struct{}

// CreateProvider implements ProviderFactory
func (f *VertexAIFactory) CreateProvider(config map[string]interface{}) (domain.Provider, error) {
	projectID, _ := config["project_id"].(string)
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	if projectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}

	// Get location or use default
	location, _ := config["location"].(string)
	if location == "" {
		location = "us-central1"
	}

	// Get model or use default
	model, _ := config["model"].(string)
	if model == "" {
		model = "gemini-1.0-pro"
	}

	return NewVertexAIProvider(projectID, location, model)
}

// ValidateConfig implements ProviderFactory
func (f *VertexAIFactory) ValidateConfig(config map[string]interface{}) error {
	projectID, _ := config["project_id"].(string)
	if projectID == "" && os.Getenv("GOOGLE_CLOUD_PROJECT") == "" {
		return fmt.Errorf("project_id is required (or set GOOGLE_CLOUD_PROJECT environment variable)")
	}
	return nil
}

// GetTemplate implements ProviderFactory
func (f *VertexAIFactory) GetTemplate() ProviderTemplate {
	metadata := NewVertexAIMetadata()
	return ProviderTemplate{
		Type:        "vertexai",
		Name:        metadata.Name(),
		Description: metadata.Description(),
		Schema:      metadata.GetConfigSchema(),
		Defaults: map[string]interface{}{
			"location":    "us-central1",
			"model":       "gemini-1.0-pro",
			"temperature": 0.7,
		},
	}
}

// RegisterDefaultFactories registers all built-in provider factories with the given registry.
// This includes factories for OpenAI, Anthropic, Gemini, Ollama, OpenRouter, VertexAI, and Mock providers.
// Returns an error if any factory registration fails.
func RegisterDefaultFactories(registry *DynamicRegistry) error {
	factories := map[string]ProviderFactory{
		"openai":     &OpenAIFactory{},
		"anthropic":  &AnthropicFactory{},
		"gemini":     &GeminiFactory{},
		"ollama":     &OllamaFactory{},
		"openrouter": &OpenRouterFactory{},
		"vertexai":   &VertexAIFactory{},
		"mock":       &MockFactory{},
	}

	for name, factory := range factories {
		if err := registry.RegisterFactory(name, factory); err != nil {
			return fmt.Errorf("failed to register %s factory: %w", name, err)
		}
	}

	return nil
}

// CreateProviderFromEnvironment creates a provider instance using environment variables.
// It looks for provider-specific environment variables prefixed with the uppercase provider type
// (e.g., OPENAI_API_KEY, ANTHROPIC_MODEL). Returns an error if the provider cannot be created.
func CreateProviderFromEnvironment(providerType string) (domain.Provider, error) {
	envPrefix := strings.ToUpper(providerType)
	config := make(map[string]interface{})

	// Common environment variables
	if apiKey := os.Getenv(envPrefix + "_API_KEY"); apiKey != "" {
		config["api_key"] = apiKey
	}
	if model := os.Getenv(envPrefix + "_MODEL"); model != "" {
		config["model"] = model
	}
	if org := os.Getenv(envPrefix + "_ORGANIZATION"); org != "" {
		config["organization"] = org
	}
	if baseURL := os.Getenv(envPrefix + "_BASE_URL"); baseURL != "" {
		config["base_url"] = baseURL
	}

	// Get factory from global registry
	factory, exists := globalRegistry.factories[providerType]
	if !exists {
		return nil, fmt.Errorf("no factory registered for provider type: %s", providerType)
	}

	return factory.CreateProvider(config)
}
