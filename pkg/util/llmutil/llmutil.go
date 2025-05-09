// Package llmutil provides utility functions for common LLM operations.
package llmutil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// ModelConfig represents a configuration for an LLM model
type ModelConfig struct {
	Provider  string // Provider identifier (e.g., "openai", "anthropic")
	Model     string // Model name
	APIKey    string // API key
	BaseURL   string // Optional base URL override
	MaxTokens int    // Optional max tokens override
}

// WithProviderOptions creates provider-specific options for initialization
func WithProviderOptions(config ModelConfig) ([]domain.ProviderOption, error) {
	var interfaceOptions []domain.ProviderOption
	
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
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	var llmProvider domain.Provider
	
	options, err := WithProviderOptions(config)
	if err != nil {
		return nil, err
	}
	
	switch config.Provider {
	case "openai":
		llmProvider = provider.NewOpenAIProvider(config.APIKey, config.Model, options...)
	
	case "anthropic":
		// TODO: Refactor Anthropic provider to use interface-based options
		// For now, fall back to the old functional options
		var anthropicOptions []provider.AnthropicOption
		
		// Manually convert interface options to functional options
		for _, opt := range options {
			if baseURLOpt, ok := opt.(*domain.BaseURLOption); ok {
				anthropicOptions = append(anthropicOptions, provider.WithAnthropicBaseURL(baseURLOpt.URL))
			}
		}
		
		llmProvider = provider.NewAnthropicProvider(config.APIKey, config.Model, anthropicOptions...)
	
	case "gemini":
		// TODO: Refactor Gemini provider to use interface-based options
		// For now, fall back to the old functional options
		var geminiOptions []provider.GeminiOption
		
		// Manually convert interface options to functional options
		for _, opt := range options {
			if baseURLOpt, ok := opt.(*domain.BaseURLOption); ok {
				geminiOptions = append(geminiOptions, provider.WithGeminiBaseURL(baseURLOpt.URL))
			}
		}
		
		llmProvider = provider.NewGeminiProvider(config.APIKey, config.Model, geminiOptions...)
	
	case "mock":
		llmProvider = provider.NewMockProvider()
	
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	return llmProvider, nil
}

// ProviderFromEnv creates a provider using environment variables
// It looks for API keys (OPENAI_API_KEY, ANTHROPIC_API_KEY, GEMINI_API_KEY) and
// custom base URLs (OPENAI_BASE_URL, ANTHROPIC_BASE_URL, GEMINI_BASE_URL)
func ProviderFromEnv() (domain.Provider, string, string, error) {
	// Check for API keys in environment variables
	openAIKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY") 
	geminiKey := os.Getenv("GEMINI_API_KEY")
	
	// Check for base URL overrides in environment variables
	openAIBaseURL := os.Getenv("OPENAI_BASE_URL")
	anthropicBaseURL := os.Getenv("ANTHROPIC_BASE_URL")
	geminiBaseURL := os.Getenv("GEMINI_BASE_URL")
	
	// Default models for each provider
	openAIModel := os.Getenv("OPENAI_MODEL")
	if openAIModel == "" {
		openAIModel = "gpt-4o"
	}
	
	anthropicModel := os.Getenv("ANTHROPIC_MODEL")
	if anthropicModel == "" {
		anthropicModel = "claude-3-5-sonnet-latest"
	}
	
	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		geminiModel = "gemini-2.0-flash-lite"
	}
	
	// Try to create a provider in order of preference
	if openAIKey != "" {
		var options []domain.ProviderOption
		
		// Add base URL option if specified
		if openAIBaseURL != "" {
			options = append(options, domain.NewBaseURLOption(openAIBaseURL))
		}
		
		llmProvider := provider.NewOpenAIProvider(openAIKey, openAIModel, options...)
		return llmProvider, "openai", openAIModel, nil
	}
	
	if anthropicKey != "" {
		var options []provider.AnthropicOption
		
		// Add base URL option if specified
		if anthropicBaseURL != "" {
			options = append(options, provider.WithAnthropicBaseURL(anthropicBaseURL))
		}
		
		llmProvider := provider.NewAnthropicProvider(anthropicKey, anthropicModel, options...)
		return llmProvider, "anthropic", anthropicModel, nil
	}
	
	if geminiKey != "" {
		var options []provider.GeminiOption
		
		// Add base URL option if specified
		if geminiBaseURL != "" {
			options = append(options, provider.WithGeminiBaseURL(geminiBaseURL))
		}
		
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