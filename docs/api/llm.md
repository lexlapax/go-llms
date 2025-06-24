# Llm API

Language model provider integration - Unified interface for OpenAI, Anthropic, Google Gemini

## Package Information

- **Import Path**: `github.com/lexlapax/go-llms/pkg/llm`
- **Category**: Core
- **Stability**: Stable (v0.3.x)

## Overview

The LLM package provides a unified interface for interacting with various language model providers including OpenAI, Anthropic, Google (Gemini and Vertex AI), Ollama, and OpenRouter. It abstracts provider-specific differences while exposing common functionality like completions, streaming, and function calling.

Key features:
- Provider abstraction with consistent API
- Automatic retry and error handling
- Token counting and rate limiting
- Streaming support for real-time responses
- Function/tool calling capabilities
- Multi-modal support (text and images)

## Core Types

### Provider Interface

The core abstraction for all LLM providers:

```go
type Provider interface {
    // Complete generates a completion for the given request
    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
    
    // GetCapabilities returns provider capabilities
    GetCapabilities() Capabilities
    
    // GetModels returns available models
    GetModels(ctx context.Context) ([]Model, error)
    
    // Close cleans up resources
    Close() error
}
```

### Streaming Support

For providers that support streaming:

```go
type StreamingProvider interface {
    Provider
    CompleteStream(ctx context.Context, request *CompletionRequest) (<-chan StreamChunk, error)
}
```

### Provider Factory

Creating providers with the factory pattern:

```go
// Create an OpenAI provider
provider, err := llm.NewProvider("openai", llm.ProviderConfig{
    APIKey: "your-api-key",
    Model: "gpt-4",
})

// Create an Anthropic provider
provider, err := llm.NewProvider("anthropic", llm.ProviderConfig{
    APIKey: "your-api-key",
    Model: "claude-3-opus-20240229",
})
```
## Examples

### Basic Completion

```go
provider, err := openai.New(openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})

response, err := provider.Complete(ctx, &llm.CompletionRequest{
    Messages: []llm.Message{
        {Role: "user", Content: "Hello, how are you?"},
    },
    Model: "gpt-3.5-turbo",
})
```

### Streaming Response

```go
stream, err := provider.CompleteStream(ctx, request)
for chunk := range stream {
    fmt.Print(chunk.Content)
}
```
## Best Practices

1. **Always use context**: Pass context for cancellation and timeouts
2. **Handle rate limits**: Implement exponential backoff for rate limit errors
3. **Monitor token usage**: Track token consumption to manage costs
4. **Use appropriate models**: Choose models based on task complexity
5. **Implement fallbacks**: Use multi-provider strategies for reliability
## Error Handling

Common errors and handling strategies:

```go
response, err := provider.Complete(ctx, request)
if err != nil {
    var apiErr *llm.APIError
    if errors.As(err, &apiErr) {
        switch apiErr.Type {
        case llm.ErrTypeRateLimit:
            // Implement backoff and retry
        case llm.ErrTypeInvalidRequest:
            // Fix request and retry
        case llm.ErrTypeAuthentication:
            // Check API key
        }
    }
}
```