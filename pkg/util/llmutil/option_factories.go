// Package llmutil provides utility functions for common LLM operations.
package llmutil

import (
	"net/http"
	"os"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// Common option factory functions

// WithPerformanceOptions creates a set of options optimized for performance.
// This includes a shorter timeout, retry settings, and a customized HTTP client.
func WithPerformanceOptions() []domain.ProviderOption {
	// Create a custom HTTP client with performance tuning
	httpClient := &http.Client{
		Timeout: 15,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			MaxConnsPerHost:     100,
			DisableKeepAlives:   false,
		},
	}

	return []domain.ProviderOption{
		domain.NewHTTPClientOption(httpClient),
		domain.NewTimeoutOption(15),
		domain.NewRetryOption(2, 300), // Retry quickly for performance-sensitive applications
	}
}

// WithReliabilityOptions creates a set of options optimized for reliability.
// This includes longer timeouts and more aggressive retry settings.
func WithReliabilityOptions() []domain.ProviderOption {
	// Create a custom HTTP client with reliability tuning
	httpClient := &http.Client{
		Timeout: 30,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 50,
			MaxConnsPerHost:     100,
			DisableKeepAlives:   false,
		},
	}

	return []domain.ProviderOption{
		domain.NewHTTPClientOption(httpClient),
		domain.NewTimeoutOption(30),
		domain.NewRetryOption(3, 1000), // More retries with longer delays for reliability
	}
}

// Provider-specific option factory functions

// WithOpenAIDefaultOptions creates a set of options commonly used with OpenAI.
// This includes reasonable defaults for organization, API headers, etc.
func WithOpenAIDefaultOptions(organizationID string) []domain.ProviderOption {
	// Start with common options
	options := WithReliabilityOptions()

	// Add OpenAI-specific options if an organization is provided
	if organizationID != "" {
		options = append(options, domain.NewOpenAIOrganizationOption(organizationID))
	}

	// Add common headers for OpenAI API
	headers := map[string]string{
		"User-Agent": "go-llms/1.0",
	}
	options = append(options, domain.NewHeadersOption(headers))

	return options
}

// WithAnthropicDefaultOptions creates a set of options commonly used with Anthropic.
// This includes a default system prompt and headers.
func WithAnthropicDefaultOptions(systemPrompt string) []domain.ProviderOption {
	// Start with common options
	options := WithReliabilityOptions()

	// Add Anthropic-specific options if a system prompt is provided
	if systemPrompt != "" {
		options = append(options, domain.NewAnthropicSystemPromptOption(systemPrompt))
	} else {
		// Default system prompt
		options = append(options, domain.NewAnthropicSystemPromptOption(
			"You are a helpful, harmless, and honest AI assistant."))
	}

	// Add common headers for Anthropic API
	headers := map[string]string{
		"User-Agent": "go-llms/1.0",
	}
	options = append(options, domain.NewHeadersOption(headers))

	return options
}

// WithGeminiDefaultOptions creates a set of options commonly used with Gemini.
// This includes a default generation config and safety settings.
func WithGeminiDefaultOptions() []domain.ProviderOption {
	// Start with common options
	options := WithReliabilityOptions()

	// Add Gemini-specific options
	generationConfig := domain.NewGeminiGenerationConfigOption().
		WithTemperature(0.7).
		WithTopK(40).
		WithTopP(0.95)

	options = append(options, generationConfig)

	// Default safety settings
	safetySettings := []map[string]interface{}{
		{
			"category":  "HARM_CATEGORY_HARASSMENT",
			"threshold": "BLOCK_MEDIUM_AND_ABOVE",
		},
		{
			"category":  "HARM_CATEGORY_HATE_SPEECH",
			"threshold": "BLOCK_MEDIUM_AND_ABOVE",
		},
		{
			"category":  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
			"threshold": "BLOCK_MEDIUM_AND_ABOVE",
		},
		{
			"category":  "HARM_CATEGORY_DANGEROUS_CONTENT",
			"threshold": "BLOCK_MEDIUM_AND_ABOVE",
		},
	}
	safetySettingsOption := domain.NewGeminiSafetySettingsOption(safetySettings)
	options = append(options, safetySettingsOption)

	return options
}

// Use case specific option factory functions

// WithStreamingOptions creates a set of options optimized for streaming responses.
// This includes a longer timeout and specific headers for streaming.
func WithStreamingOptions() []domain.ProviderOption {
	// Create a custom HTTP client optimized for streaming
	httpClient := &http.Client{
		Timeout: 60, // Longer timeout for streaming
		Transport: &http.Transport{
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     20,
			DisableKeepAlives:   false,
		},
	}

	// Add common headers for streaming APIs
	headers := map[string]string{
		"User-Agent":    "go-llms/1.0",
		"Accept":        "text/event-stream",
		"Cache-Control": "no-cache",
	}

	return []domain.ProviderOption{
		domain.NewHTTPClientOption(httpClient),
		domain.NewTimeoutOption(60),
		domain.NewHeadersOption(headers),
	}
}

// WithProxyOptions creates options for working with API proxies.
// This includes custom base URL and headers needed for proxy authentication.
func WithProxyOptions(baseURL string, proxyAPIKey string) []domain.ProviderOption {
	options := []domain.ProviderOption{}

	// Add base URL if provided
	if baseURL != "" {
		options = append(options, domain.NewBaseURLOption(baseURL))
	}

	// Add headers for proxy authentication if API key is provided
	if proxyAPIKey != "" {
		headers := map[string]string{
			"User-Agent":      "go-llms/1.0",
			"X-Proxy-API-Key": proxyAPIKey,
		}
		options = append(options, domain.NewHeadersOption(headers))
	}

	return options
}

// Combined options for specific scenarios

// WithOpenAIStreamingOptions combines OpenAI default options with streaming options.
func WithOpenAIStreamingOptions(organizationID string) []domain.ProviderOption {
	options := WithOpenAIDefaultOptions(organizationID)
	streamingOptions := WithStreamingOptions()

	// Replace HTTP client and timeout with streaming-optimized versions
	return append(options, streamingOptions...)
}

// WithAnthropicStreamingOptions combines Anthropic default options with streaming options.
func WithAnthropicStreamingOptions(systemPrompt string) []domain.ProviderOption {
	options := WithAnthropicDefaultOptions(systemPrompt)
	streamingOptions := WithStreamingOptions()

	// Replace HTTP client and timeout with streaming-optimized versions
	return append(options, streamingOptions...)
}

// CreateOptionFactoryFromEnv creates provider options combining environment variables and factory functions.
// This function provides a consolidated approach to creating provider options by:
// 1. Getting options from environment variables
// 2. Applying option factory functions based on the use case
// 3. Merging both sets of options with appropriate priority
//
// Parameters:
//   - providerType: The provider type ("openai", "anthropic", "gemini")
//   - useCase: The use case ("default", "performance", "reliability", "streaming")
//     If empty, the function will look for a use case in the environment variables
//
// The function determines which factory function to use based on the provider and use case,
// then merges those options with any options found in environment variables.
func CreateOptionFactoryFromEnv(providerType, useCase string) []domain.ProviderOption {
	var options []domain.ProviderOption

	// First, get options from environment variables
	envOptions := GetProviderOptionsFromEnv(providerType)

	// If useCase is empty, check the environment for provider-specific use case
	if useCase == "" {
		switch providerType {
		case "openai":
			useCase = os.Getenv(EnvOpenAIUseCase)
		case "anthropic":
			useCase = os.Getenv(EnvAnthropicUseCase)
		case "gemini":
			useCase = os.Getenv(EnvGeminiUseCase)
		}

		// If still empty after checking environment, default to "default"
		if useCase == "" {
			useCase = "default"
		}
	}

	// Next, apply factory function options based on provider and use case
	var factoryOptions []domain.ProviderOption

	switch providerType {
	case "openai":
		// Get organization ID from environment if available
		orgID := os.Getenv(EnvOpenAIOrganization)

		switch useCase {
		case "streaming":
			factoryOptions = WithOpenAIStreamingOptions(orgID)
		case "performance":
			factoryOptions = append(WithPerformanceOptions(), domain.NewOpenAIOrganizationOption(orgID))
		case "reliability":
			factoryOptions = append(WithReliabilityOptions(), domain.NewOpenAIOrganizationOption(orgID))
		default:
			factoryOptions = WithOpenAIDefaultOptions(orgID)
		}

	case "anthropic":
		// Get system prompt from environment if available
		systemPrompt := os.Getenv(EnvAnthropicSystemPrompt)

		switch useCase {
		case "streaming":
			factoryOptions = WithAnthropicStreamingOptions(systemPrompt)
		case "performance":
			factoryOptions = append(WithPerformanceOptions(), domain.NewAnthropicSystemPromptOption(systemPrompt))
		case "reliability":
			factoryOptions = append(WithReliabilityOptions(), domain.NewAnthropicSystemPromptOption(systemPrompt))
		default:
			factoryOptions = WithAnthropicDefaultOptions(systemPrompt)
		}

	case "gemini":
		switch useCase {
		case "streaming":
			factoryOptions = append(WithStreamingOptions(), WithGeminiDefaultOptions()...)
		case "performance":
			factoryOptions = append(WithPerformanceOptions(), WithGeminiDefaultOptions()...)
		case "reliability":
			factoryOptions = append(WithReliabilityOptions(), WithGeminiDefaultOptions()...)
		default:
			factoryOptions = WithGeminiDefaultOptions()
		}

	default:
		// For unknown providers, just use common options based on use case
		switch useCase {
		case "streaming":
			factoryOptions = WithStreamingOptions()
		case "performance":
			factoryOptions = WithPerformanceOptions()
		case "reliability":
			factoryOptions = WithReliabilityOptions()
		}
	}

	// Merge options, giving priority to environment variables
	// First add factory options as defaults
	options = append(options, factoryOptions...)

	// Then add environment options which may override factory options
	options = append(options, envOptions...)

	return options
}
