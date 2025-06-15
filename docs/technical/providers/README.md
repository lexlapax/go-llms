# Provider Documentation

> **[Documentation Home](../README.md) / Providers**

## Overview

Providers are the foundation of go-llms, abstracting different LLM services behind a unified interface. This section covers everything you need to know about working with and implementing providers.

## Documentation Structure

### Core Documentation
- [**Provider System Overview**](overview.md) - Understanding how providers work
- [**Implementing Providers**](implementing-providers.md) - Create custom provider implementations
- [**Provider Registry**](provider-registry.md) - Dynamic registration and discovery
- [**Provider Metadata**](metadata.md) - Capabilities and configuration

### Available Providers

#### Production-Ready Providers
- **OpenAI** - GPT-4, GPT-3.5, and other OpenAI models
- **Anthropic** - Claude 3.5 Sonnet, Claude 3 Opus, and other Anthropic models
- **Google Gemini** - Gemini 2.0 Flash, Gemini Pro, and other Google models
- **Google Vertex AI** - Enterprise Google AI with additional features
- **Ollama** - Local model execution
- **OpenRouter** - Multi-provider gateway

#### Planned Providers (v0.4.x)
- **Mistral AI** - Mistral and Mixtral models
- **AWS Bedrock** - Multiple models through AWS
- **Azure OpenAI** - OpenAI models via Azure

## Quick Start

### Using a Provider
```go
// Create a provider
provider := provider.NewOpenAIProvider(apiKey, "gpt-4")

// Generate a response
response, err := provider.Generate(ctx, "Hello, world!")

// Generate with streaming
stream, err := provider.Stream(ctx, "Tell me a story")
for chunk := range stream {
    fmt.Print(chunk.Content)
}

// Generate with schema validation
result, err := provider.GenerateWithSchema(ctx, prompt, schema)
```

### Provider Selection Guide

| Provider | Best For | Key Features |
|----------|----------|--------------|
| OpenAI | General purpose, function calling | GPT-4, wide tool support |
| Anthropic | Long context, nuanced tasks | 200k context, Claude 3.5 |
| Google Gemini | Fast responses, multimodal | Very fast, vision support |
| Vertex AI | Enterprise, Google Cloud integration | SLA, private endpoints |
| Ollama | Local/private deployment | No API costs, full control |
| OpenRouter | Provider flexibility | Switch models easily |

## Common Patterns

### Provider Options
```go
// With custom HTTP client
provider := provider.NewOpenAIProvider(apiKey, model,
    domain.WithHTTPClient(customClient),
)

// With custom base URL
provider := provider.NewOpenAIProvider(apiKey, model,
    domain.WithBaseURL("https://custom.openai.com"),
)

// With timeout
provider := provider.NewOpenAIProvider(apiKey, model,
    domain.WithTimeout(30 * time.Second),
)
```

### Provider Selection Guide

![Provider Registry](../images/provider-registry.svg)
*Figure 1: Provider registration and discovery system showing how providers are dynamically registered and selected*

### Error Handling
```go
response, err := provider.Generate(ctx, prompt)
if err != nil {
    var providerErr *errors.ProviderError
    if errors.As(err, &providerErr) {
        switch providerErr.Type {
        case errors.ErrTypeRateLimit:
            // Handle rate limiting
        case errors.ErrTypeAuthentication:
            // Handle auth errors
        case errors.ErrTypeContextLength:
            // Handle context too long
        }
    }
}
```

### Multi-Provider Setup
```go
// Create multiple providers
providers := []domain.Provider{
    provider.NewOpenAIProvider(openaiKey, "gpt-4"),
    provider.NewAnthropicProvider(anthropicKey, "claude-3-5-sonnet-latest"),
}

// Use with multi-provider strategies
multiProvider := provider.NewMultiProvider(
    provider.WithProviders(providers...),
    provider.WithStrategy(provider.StrategyFastest),
)
```

## Provider Features Matrix

| Feature | OpenAI | Anthropic | Gemini | Vertex AI | Ollama | OpenRouter |
|---------|--------|-----------|---------|-----------|---------|------------|
| Text Generation | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Streaming | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Function Calling | ✅ | ✅ | ✅ | ✅ | ⚠️ | ⚠️ |
| Vision | ✅ | ✅ | ✅ | ✅ | ⚠️ | ⚠️ |
| JSON Mode | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Context Length | 128k | 200k | 1M | 1M | Varies | Varies |

Legend: ✅ Full support | ⚠️ Model dependent | ❌ Not supported

## Next Steps

- Read [Provider System Overview](overview.md) to understand the architecture
- Follow [Implementing Providers](implementing-providers.md) to create custom providers
- Explore [Provider Registry](provider-registry.md) for dynamic provider management
- Check [Provider Metadata](metadata.md) for capability discovery