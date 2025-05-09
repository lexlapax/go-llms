// Package llmutil provides utility functions for common LLM operations.
package llmutil

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// ModelConfig represents a configuration for an LLM model
type ModelConfig struct {
	Provider  string                 // Provider identifier (e.g., "openai", "anthropic")
	Model     string                 // Model name
	APIKey    string                 // API key
	BaseURL   string                 // Optional base URL override
	MaxTokens int                    // Optional max tokens override
	Options   []domain.ProviderOption // Optional provider-specific options
}

// WithProviderOptions creates provider-specific options for initialization
func WithProviderOptions(config ModelConfig) ([]domain.ProviderOption, error) {
	var interfaceOptions []domain.ProviderOption

	// Add user-provided options
	if config.Options != nil {
		interfaceOptions = append(interfaceOptions, config.Options...)
	}

	// Add base URL option if specified
	if config.BaseURL != "" {
		// Only add interface options for valid providers
		if config.Provider == "openai" || config.Provider == "anthropic" || config.Provider == "gemini" {
			baseURLOption := domain.NewBaseURLOption(config.BaseURL)
			interfaceOptions = append(interfaceOptions, baseURLOption)
		}
	}

	return interfaceOptions, nil
}

// CreateProvider creates an LLM provider based on configuration
func CreateProvider(config ModelConfig) (domain.Provider, error) {
	// Skip API key check for mock provider
	if config.Provider != "mock" && config.APIKey == "" {
		// Try to get API key from environment if not provided in config
		apiKey := GetAPIKeyFromEnv(config.Provider)
		if apiKey == "" {
			return nil, fmt.Errorf("API key is required (not provided in config or environment)")
		}
		config.APIKey = apiKey
	}

	// If model is not specified, try to get it from environment
	if config.Model == "" {
		config.Model = GetModelFromEnv(config.Provider)
	}

	var llmProvider domain.Provider

	// Get options from configuration
	options, err := WithProviderOptions(config)
	if err != nil {
		return nil, err
	}

	// If no options provided in config, try to get them from environment
	if config.Options == nil || len(config.Options) == 0 {
		envOptions := GetProviderOptionsFromEnv(config.Provider)
		options = append(options, envOptions...)
	}

	switch config.Provider {
	case "openai":
		llmProvider = provider.NewOpenAIProvider(config.APIKey, config.Model, options...)

	case "anthropic":
		llmProvider = provider.NewAnthropicProvider(config.APIKey, config.Model, options...)

	case "gemini":
		llmProvider = provider.NewGeminiProvider(config.APIKey, config.Model, options...)

	case "mock":
		llmProvider = provider.NewMockProvider()

	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	return llmProvider, nil
}

// ProviderFromEnv creates a provider using environment variables
// It looks for API keys, base URLs, models, and other provider-specific options
// from environment variables and applies them when creating the provider.
func ProviderFromEnv() (domain.Provider, string, string, error) {
	// Check for API keys in environment variables
	openAIKey := GetAPIKeyFromEnv("openai")
	anthropicKey := GetAPIKeyFromEnv("anthropic")
	geminiKey := GetAPIKeyFromEnv("gemini")

	// Get model names from environment variables (with defaults)
	openAIModel := GetModelFromEnv("openai")
	anthropicModel := GetModelFromEnv("anthropic")
	geminiModel := GetModelFromEnv("gemini")

	// Try to create a provider in order of preference
	if openAIKey != "" {
		// Get OpenAI-specific options from environment variables
		options := GetOpenAIOptionsFromEnv()

		// Create provider with options
		llmProvider := provider.NewOpenAIProvider(openAIKey, openAIModel, options...)
		return llmProvider, "openai", openAIModel, nil
	}

	if anthropicKey != "" {
		// Get Anthropic-specific options from environment variables
		options := GetAnthropicOptionsFromEnv()

		// Create provider with options
		llmProvider := provider.NewAnthropicProvider(anthropicKey, anthropicModel, options...)
		return llmProvider, "anthropic", anthropicModel, nil
	}

	if geminiKey != "" {
		// Get Gemini-specific options from environment variables
		options := GetGeminiOptionsFromEnv()

		// Create provider with options
		llmProvider := provider.NewGeminiProvider(geminiKey, geminiModel, options...)
		return llmProvider, "gemini", geminiModel, nil
	}

	// If no API keys are found, create a mock provider
	mockProvider := provider.NewMockProvider()
	return mockProvider, "mock", "default", nil
}

// BatchGenerate generates responses for multiple prompts in parallel
func BatchGenerate(ctx context.Context, provider domain.Provider, prompts []string, options ...domain.Option) ([]string, []error) {
	results := make([]string, len(prompts))
	errors := make([]error, len(prompts))
	var wg sync.WaitGroup
	
	for i, prompt := range prompts {
		wg.Add(1)
		go func(idx int, p string) {
			defer wg.Done()
			result, err := provider.Generate(ctx, p, options...)
			results[idx] = result
			errors[idx] = err
		}(i, prompt)
	}
	
	wg.Wait()
	return results, errors
}

// GenerateWithRetry attempts generation with automatic retries
func GenerateWithRetry(ctx context.Context, provider domain.Provider, prompt string, maxRetries int, options ...domain.Option) (string, error) {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		result, err := provider.Generate(ctx, prompt, options...)
		if err == nil {
			return result, nil
		}
		
		// Check if this is a retryable error
		if !IsRetryableError(err) {
			return "", err
		}
		
		lastErr = err
	}
	
	return "", fmt.Errorf("max retries reached: %w", lastErr)
}

// IsRetryableError determines if an error can be retried
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Network connectivity errors are retryable
	if domain.IsNetworkConnectivityError(err) {
		return true
	}
	
	// Rate limit errors are retryable
	if domain.IsRateLimitError(err) {
		return true
	}
	
	// Other errors are not retryable
	return false
}

// ProcessTypedWithProvider is a convenience function to generate and process structured output in one step
func ProcessTypedWithProvider[T any](
	ctx context.Context, 
	provider domain.Provider,
	prompt string, 
	target *T,
	options ...domain.Option,
) error {
	// Create a processor with a validator
	validator := validation.NewValidator()
	proc := processor.NewStructuredProcessor(validator)
	
	// In a real implementation, we would generate schema from the type
	// For now, use reflection.GenerateSchema which exists in the codebase
	
	// Create a simple schema for demonstration purposes
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{},
	}
	
	// Generate with schema
	result, err := provider.GenerateWithSchema(ctx, prompt, schema, options...)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}
	
	// Convert result to target type using the processor
	resultStr, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	
	return proc.ProcessTyped(schema, string(resultStr), target)
}

// GenerateWithOptions is a convenience function to generate with common option patterns
func GenerateWithOptions(
	ctx context.Context,
	provider domain.Provider,
	prompt string,
	temperature float64,
	maxTokens int,
) (string, error) {
	options := []domain.Option{
		domain.WithTemperature(temperature),
		domain.WithMaxTokens(maxTokens),
	}
	
	return provider.Generate(ctx, prompt, options...)
}

// ConcurrentStreamMessages streams multiple message sequences concurrently and merges results
func ConcurrentStreamMessages(
	ctx context.Context,
	provider domain.Provider,
	messageGroups [][]domain.Message,
	options ...domain.Option,
) ([]domain.ResponseStream, []error) {
	streams := make([]domain.ResponseStream, len(messageGroups))
	errors := make([]error, len(messageGroups))
	var wg sync.WaitGroup
	
	for i, messages := range messageGroups {
		wg.Add(1)
		go func(idx int, msgs []domain.Message) {
			defer wg.Done()
			stream, err := provider.StreamMessage(ctx, msgs, options...)
			streams[idx] = stream
			errors[idx] = err
		}(i, messages)
	}
	
	wg.Wait()
	return streams, errors
}