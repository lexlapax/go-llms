// Package provider implements LLM provider integrations for OpenAI, Anthropic, Google, and others.
//
// This package provides a unified interface for interacting with different LLM services,
// handling authentication, request/response formatting, error mapping, and capability discovery.
//
// # Provider Types
//
// The package supports several provider implementations:
//   - OpenAI (GPT-3.5, GPT-4, etc.)
//   - Anthropic (Claude models)
//   - Google (Gemini models via Google AI and Vertex AI)
//   - Ollama (local models)
//   - OpenRouter (unified access to multiple providers)
//   - Mock (for testing)
//
// # Core Concepts
//
// Provider: The main interface for interacting with an LLM service.
// Each provider handles its specific API format while exposing a common interface.
//
// Registry: A dynamic registry system for discovering and creating providers at runtime.
// Supports factory registration, templates, and configuration management.
//
// Metadata: Each provider exposes metadata about its capabilities, models, and constraints.
// This enables runtime discovery of features like vision support, function calling, etc.
//
// # Usage
//
// Basic provider creation:
//
//	provider := provider.NewOpenAIProvider(apiKey, "gpt-4")
//	response, err := provider.Generate(ctx, "Hello, world!")
//
// Using the registry:
//
//	registry := provider.GetGlobalRegistry()
//	p, err := registry.CreateProviderFromTemplate("openai", config)
//
// # Error Handling
//
// The package maps provider-specific errors to common error types:
//   - RateLimitError: API rate limits exceeded
//   - AuthenticationError: Invalid credentials
//   - InvalidRequestError: Malformed requests
//   - ModelNotFoundError: Requested model doesn't exist
//
// # Configuration
//
// Providers can be configured through:
//   - Direct constructor parameters
//   - Environment variables
//   - Registry templates
//   - Runtime options
//
// See individual provider documentation for specific configuration options.
package provider
