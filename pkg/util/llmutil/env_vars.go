// Package llmutil provides utility functions for common LLM operations.
package llmutil

import (
	"os"
	"strconv"
	"strings"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// Environment variable constants for common provider options
const (
	// Common options
	EnvHTTPTimeout   = "LLM_HTTP_TIMEOUT"   // Timeout in seconds for HTTP client
	EnvRetryAttempts = "LLM_RETRY_ATTEMPTS" // Number of retry attempts for failed requests
	EnvRetryDelay    = "LLM_RETRY_DELAY"    // Delay between retries in milliseconds

	// OpenAI options
	EnvOpenAIOrganization = "OPENAI_ORGANIZATION" // Organization ID for OpenAI
	EnvOpenAIModel        = "OPENAI_MODEL"        // Model to use for OpenAI
	EnvOpenAIBaseURL      = "OPENAI_BASE_URL"     // Base URL for OpenAI API
	EnvOpenAIAPIKey       = "OPENAI_API_KEY"      // API key for OpenAI

	// Anthropic options
	EnvAnthropicSystemPrompt = "ANTHROPIC_SYSTEM_PROMPT" // System prompt for Anthropic
	EnvAnthropicModel        = "ANTHROPIC_MODEL"         // Model to use for Anthropic
	EnvAnthropicBaseURL      = "ANTHROPIC_BASE_URL"      // Base URL for Anthropic API
	EnvAnthropicAPIKey       = "ANTHROPIC_API_KEY"       // API key for Anthropic

	// Gemini options
	EnvGeminiGenerationConfig = "GEMINI_GENERATION_CONFIG" // JSON string with generation config for Gemini
	EnvGeminiSafetySettings   = "GEMINI_SAFETY_SETTINGS"   // JSON string with safety settings for Gemini
	EnvGeminiModel            = "GEMINI_MODEL"             // Model to use for Gemini
	EnvGeminiBaseURL          = "GEMINI_BASE_URL"          // Base URL for Gemini API
	EnvGeminiAPIKey           = "GEMINI_API_KEY"           // API key for Gemini
)

// GetCommonOptionsFromEnv retrieves common provider options from environment variables.
// These options can be applied to any provider.
func GetCommonOptionsFromEnv() []domain.ProviderOption {
	var options []domain.ProviderOption

	// HTTP Timeout option
	if timeoutStr := os.Getenv(EnvHTTPTimeout); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout > 0 {
			options = append(options, domain.NewTimeoutOption(timeout))
		}
	}

	// Retry option
	retryAttempts := 0
	retryDelay := 0

	if attemptsStr := os.Getenv(EnvRetryAttempts); attemptsStr != "" {
		if attempts, err := strconv.Atoi(attemptsStr); err == nil && attempts > 0 {
			retryAttempts = attempts
		}
	}

	if delayStr := os.Getenv(EnvRetryDelay); delayStr != "" {
		if delay, err := strconv.Atoi(delayStr); err == nil && delay > 0 {
			retryDelay = delay
		}
	}

	if retryAttempts > 0 {
		options = append(options, domain.NewRetryOption(retryAttempts, retryDelay))
	}

	return options
}

// GetOpenAIOptionsFromEnv retrieves OpenAI-specific options from environment variables.
func GetOpenAIOptionsFromEnv() []domain.ProviderOption {
	var options []domain.ProviderOption

	// First, add common options
	options = append(options, GetCommonOptionsFromEnv()...)

	// Base URL option
	if baseURL := os.Getenv(EnvOpenAIBaseURL); baseURL != "" {
		options = append(options, domain.NewBaseURLOption(baseURL))
	}

	// Organization option
	if org := os.Getenv(EnvOpenAIOrganization); org != "" {
		options = append(options, domain.NewOpenAIOrganizationOption(org))
	}

	// Logit bias is more complex and would typically come from a JSON string in an env var
	// We'll skip it for now as it's less common

	return options
}

// GetAnthropicOptionsFromEnv retrieves Anthropic-specific options from environment variables.
func GetAnthropicOptionsFromEnv() []domain.ProviderOption {
	var options []domain.ProviderOption

	// First, add common options
	options = append(options, GetCommonOptionsFromEnv()...)

	// Base URL option
	if baseURL := os.Getenv(EnvAnthropicBaseURL); baseURL != "" {
		options = append(options, domain.NewBaseURLOption(baseURL))
	}

	// System prompt option
	if systemPrompt := os.Getenv(EnvAnthropicSystemPrompt); systemPrompt != "" {
		options = append(options, domain.NewAnthropicSystemPromptOption(systemPrompt))
	}

	// Metadata is more complex and would typically come from a JSON string in an env var
	// We'll skip it for now as it's less common

	return options
}

// GetGeminiOptionsFromEnv retrieves Gemini-specific options from environment variables.
func GetGeminiOptionsFromEnv() []domain.ProviderOption {
	var options []domain.ProviderOption

	// First, add common options
	options = append(options, GetCommonOptionsFromEnv()...)

	// Base URL option
	if baseURL := os.Getenv(EnvGeminiBaseURL); baseURL != "" {
		options = append(options, domain.NewBaseURLOption(baseURL))
	}

	// Generation config and safety settings are more complex and would typically come
	// from a JSON string in an env var - we'll skip detailed parsing for now

	return options
}

// GetProviderOptionsFromEnv retrieves options for a specific provider from environment variables.
func GetProviderOptionsFromEnv(providerType string) []domain.ProviderOption {
	providerType = strings.ToLower(providerType)

	switch providerType {
	case "openai":
		return GetOpenAIOptionsFromEnv()
	case "anthropic":
		return GetAnthropicOptionsFromEnv()
	case "gemini":
		return GetGeminiOptionsFromEnv()
	default:
		return GetCommonOptionsFromEnv()
	}
}

// GetAPIKeyFromEnv retrieves the API key for a specific provider from environment variables.
func GetAPIKeyFromEnv(providerType string) string {
	providerType = strings.ToLower(providerType)

	switch providerType {
	case "openai":
		return os.Getenv(EnvOpenAIAPIKey)
	case "anthropic":
		return os.Getenv(EnvAnthropicAPIKey)
	case "gemini":
		return os.Getenv(EnvGeminiAPIKey)
	default:
		return ""
	}
}

// GetModelFromEnv retrieves the model name for a specific provider from environment variables.
func GetModelFromEnv(providerType string) string {
	providerType = strings.ToLower(providerType)

	switch providerType {
	case "openai":
		model := os.Getenv(EnvOpenAIModel)
		if model == "" {
			return "gpt-4o"
		}
		return model
	case "anthropic":
		model := os.Getenv(EnvAnthropicModel)
		if model == "" {
			return "claude-3-5-sonnet-latest"
		}
		return model
	case "gemini":
		model := os.Getenv(EnvGeminiModel)
		if model == "" {
			return "gemini-2.0-flash-lite"
		}
		return model
	default:
		return ""
	}
}
