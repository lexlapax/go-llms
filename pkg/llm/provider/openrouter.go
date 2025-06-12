package provider

// ABOUTME: OpenRouter provider implementation using OpenAI-compatible API
// ABOUTME: Provides access to 400+ models through a unified interface

import (
	"os"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

const (
	defaultOpenRouterHost = "https://openrouter.ai/api"
)

// NewOpenRouterProvider creates a new OpenRouter provider using the OpenAI-compatible API
// OpenRouter provides access to 400+ models from various providers through a single API.
// It automatically handles fallbacks and selects cost-effective options.
func NewOpenRouterProvider(apiKey, model string, options ...domain.ProviderOption) *OpenAIProvider {
	// Set default options for OpenRouter
	defaultOptions := []domain.ProviderOption{
		domain.NewBaseURLOption(defaultOpenRouterHost),
	}

	// Check for OPENROUTER_API_BASE environment variable
	if base := os.Getenv("OPENROUTER_API_BASE"); base != "" {
		defaultOptions[0] = domain.NewBaseURLOption(base)
	}

	// Apply user-provided options (can override defaults)
	allOptions := append(defaultOptions, options...)

	// Create OpenAI provider with OpenRouter configuration
	return NewOpenAIProvider(apiKey, model, allOptions...)
}
