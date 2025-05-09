# Provider Options System

> **[Documentation Home](/REFERENCE.md) / [User Guide](/docs/user-guide/README.md) / Provider Options**

This guide explains the provider options system in Go-LLMs, which allows you to configure LLM providers with both common and provider-specific options using a type-safe, interface-based approach.

## Overview

The provider option system in Go-LLMs allows you to:

1. Configure common settings across all providers (HTTP client, base URL, timeout, etc.)
2. Apply provider-specific options for unique features (organization ID, system prompt, generation config, etc.)
3. Combine multiple options when creating providers
4. Use a type-safe approach with Go's interface system

## Core Interfaces

The options system is built around several key interfaces:

```go
// Base interface for all provider options
type ProviderOption interface {
    // Identifies which provider type this option is for
    ProviderType() string
}

// Provider-specific option interfaces
type OpenAIOption interface {
    ProviderOption
    ApplyToOpenAI(provider interface{})
}

type AnthropicOption interface {
    ProviderOption
    ApplyToAnthropic(provider interface{})
}

type GeminiOption interface {
    ProviderOption
    ApplyToGemini(provider interface{})
}

type MockOption interface {
    ProviderOption
    ApplyToMock(provider interface{})
}

// Common options implement all provider interfaces
type CommonOption interface {
    OpenAIOption
    AnthropicOption
    GeminiOption
    MockOption
}
```

## Common Options

Go-LLMs provides several common options that work across all providers:

### BaseURLOption

Sets a custom base URL for the provider API:

```go
// Set custom API endpoint
baseURLOption := domain.NewBaseURLOption("https://custom-endpoint.example.com")
```

### HTTPClientOption

Sets a custom HTTP client for the provider:

```go
// Create a custom HTTP client with specific timeout
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}
httpClientOption := domain.NewHTTPClientOption(httpClient)
```

### TimeoutOption

Sets the timeout duration for API requests:

```go
// Set timeout to 15 seconds
timeoutOption := domain.NewTimeoutOption(15000) // milliseconds
```

### RetryOption

Configures retry behavior for API requests:

```go
// Retry up to 3 times with 1 second delay
retryOption := domain.NewRetryOption(3, 1000) // max retries, delay in milliseconds
```

### HeadersOption

Sets custom HTTP headers for API requests:

```go
// Set custom headers
headersOption := domain.NewHeadersOption(map[string]string{
    "X-Custom-Header": "custom-value",
})
```

### ModelOption

Sets the model name for the provider:

```go
// Set model name
modelOption := domain.NewModelOption("gpt-4o")
```

## Provider-Specific Options

Each provider has its own set of options for unique features:

### OpenAI-Specific Options

#### OpenAIOrganizationOption

Sets the organization ID for OpenAI API calls:

```go
// Set organization ID
orgOption := domain.NewOpenAIOrganizationOption("org-123456")
```

#### OpenAILogitBiasOption

Sets the logit bias for token selection:

```go
// Discourage certain tokens
logitBiasOption := domain.NewOpenAILogitBiasOption(map[string]float64{
    "50256": -100, // Discourage token 50256
})
```

### Anthropic-Specific Options

#### AnthropicSystemPromptOption

Sets the system prompt for all Anthropic API calls:

```go
// Set system prompt
systemPromptOption := domain.NewAnthropicSystemPromptOption(
    "You are a helpful coding assistant specializing in Go programming.")
```

#### AnthropicMetadataOption

Sets metadata for Anthropic API calls:

```go
// Set metadata for tracking
metadataOption := domain.NewAnthropicMetadataOption(map[string]string{
    "user_id": "user123",
    "session_id": "session456",
})
```

### Gemini-Specific Options

#### GeminiGenerationConfigOption

Configures generation parameters for Gemini:

```go
// Set generation parameters
generationConfigOption := domain.NewGeminiGenerationConfigOption().
    WithTemperature(0.7).
    WithTopK(40).
    WithMaxOutputTokens(1024).
    WithTopP(0.95)
```

#### GeminiSafetySettingsOption

Configures content filtering settings:

```go
// Set safety filters
safetySettings := []map[string]interface{}{
    {
        "category": "HARM_CATEGORY_HARASSMENT",
        "threshold": "BLOCK_MEDIUM_AND_ABOVE",
    },
    {
        "category": "HARM_CATEGORY_HATE_SPEECH",
        "threshold": "BLOCK_MEDIUM_AND_ABOVE",
    },
}
safetySettingsOption := domain.NewGeminiSafetySettingsOption(safetySettings)
```

## Using Options with Providers

Options can be passed when creating new providers:

```go
// Create OpenAI provider with options
openaiProvider := provider.NewOpenAIProvider(
    apiKey,
    "gpt-4o",
    domain.NewHTTPClientOption(httpClient),
    domain.NewOpenAIOrganizationOption("org-123456"),
)

// Create Anthropic provider with options
anthropicProvider := provider.NewAnthropicProvider(
    apiKey,
    "claude-3-5-sonnet-latest",
    domain.NewAnthropicSystemPromptOption("You are a helpful assistant."),
    domain.NewBaseURLOption("https://custom-endpoint.example.com"),
)

// Create Gemini provider with options
geminiProvider := provider.NewGeminiProvider(
    apiKey,
    "gemini-2.0-flash-lite",
    domain.NewGeminiGenerationConfigOption().WithTemperature(0.7),
    domain.NewGeminiSafetySettingsOption(safetySettings),
)
```

## Combining Multiple Options

You can combine multiple options when creating a provider:

```go
// Create an HTTP client with custom timeout
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}

// Gather all common options
commonOptions := []domain.ProviderOption{
    domain.NewHTTPClientOption(httpClient),
    domain.NewBaseURLOption("https://api.custom-endpoint.com"),
    domain.NewTimeoutOption(15000),
    domain.NewRetryOption(3, 1000),
    domain.NewHeadersOption(map[string]string{
        "X-Custom-Header": "custom-value",
    }),
}

// Add provider-specific options
providerOptions := append(
    commonOptions,
    domain.NewOpenAIOrganizationOption("org-123456"),
    domain.NewOpenAILogitBiasOption(map[string]float64{
        "50256": -100,
    }),
)

// Create provider with all options
openaiProvider := provider.NewOpenAIProvider(
    apiKey,
    "gpt-4o",
    providerOptions...,
)
```

## Using Options with ModelConfig

When using `llmutil.ModelConfig`, you can pass provider options:

```go
// Create ModelConfig
config := llmutil.ModelConfig{
    Provider: "openai",
    Model:    "gpt-4o",
    APIKey:   apiKey,
    BaseURL:  "https://api.openai.com", // This creates a BaseURLOption automatically
}

// Create provider from config with additional options
orgOption := domain.NewOpenAIOrganizationOption("org-123456")
provider, err := llmutil.CreateProviderWithOptions(config, orgOption)
```

## Best Practices

### Provider Type Safety

Each option is designed to work with specific providers:

- `OpenAIOption` only works with `OpenAIProvider`
- `AnthropicOption` only works with `AnthropicProvider`
- `GeminiOption` only works with `GeminiProvider`
- `MockOption` only works with `MockProvider`
- `CommonOption` works with all providers

The system will automatically ignore options that don't apply to a provider, making it safe to mix options.

### Prefer Builder Pattern for Complex Options

For options with multiple settings like `GeminiGenerationConfigOption`, use the builder pattern:

```go
// Builder pattern for complex options
generationConfigOption := domain.NewGeminiGenerationConfigOption().
    WithTemperature(0.7).
    WithTopK(40).
    WithMaxOutputTokens(1024).
    WithTopP(0.95)
```

### Using Options with Multiple Providers

When working with the multi-provider system, you need to apply options to each provider separately:

```go
// Create providers with their specific options
openaiProvider := provider.NewOpenAIProvider(
    openaiKey,
    "gpt-4o",
    domain.NewOpenAIOrganizationOption("org-123456"),
)

anthropicProvider := provider.NewAnthropicProvider(
    anthropicKey,
    "claude-3-5-sonnet-latest",
    domain.NewAnthropicSystemPromptOption("You are a helpful assistant."),
)

// Create provider weights
providers := []provider.ProviderWeight{
    {Provider: openaiProvider, Weight: 1.0, Name: "openai"},
    {Provider: anthropicProvider, Weight: 1.0, Name: "anthropic"},
}

// Create multi-provider
multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)
```

## Examples

### Simple Example with Common Options

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "time"

    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        fmt.Println("Please set OPENAI_API_KEY")
        return
    }

    // Create custom HTTP client
    httpClient := &http.Client{
        Timeout: 30 * time.Second,
    }
    
    // Create options
    httpClientOption := domain.NewHTTPClientOption(httpClient)
    headersOption := domain.NewHeadersOption(map[string]string{
        "X-Custom-Header": "custom-value",
    })
    
    // Create provider with options
    openaiProvider := provider.NewOpenAIProvider(
        apiKey,
        "gpt-4o",
        httpClientOption,
        headersOption,
    )
    
    // Use the provider
    response, err := openaiProvider.Generate(
        context.Background(),
        "Hello, world!",
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Response: %s\n", response)
}
```

### Provider-Specific Options Example

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
    if anthropicKey == "" {
        fmt.Println("Please set ANTHROPIC_API_KEY")
        return
    }

    // Create system prompt option
    systemPromptOption := domain.NewAnthropicSystemPromptOption(
        "You are a helpful coding assistant specializing in Go programming. You provide clear, concise responses focused on best practices.")
    
    // Create metadata option
    metadataOption := domain.NewAnthropicMetadataOption(map[string]string{
        "user_id": "user123",
        "session_id": "session456",
    })
    
    // Create provider with options
    anthropicProvider := provider.NewAnthropicProvider(
        anthropicKey,
        "claude-3-5-sonnet-latest",
        systemPromptOption,
        metadataOption,
    )
    
    // Use the provider with message-based API
    response, err := anthropicProvider.Generate(
        context.Background(),
        "Explain Go interfaces in simple terms",
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Response: %s\n", response)
}
```

### Combined Options Example

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "time"

    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    geminiKey := os.Getenv("GEMINI_API_KEY")
    if geminiKey == "" {
        fmt.Println("Please set GEMINI_API_KEY")
        return
    }

    // Create common options
    httpClient := &http.Client{
        Timeout: 45 * time.Second,
    }
    httpClientOption := domain.NewHTTPClientOption(httpClient)
    timeoutOption := domain.NewTimeoutOption(30000) // 30 seconds
    
    // Create Gemini-specific options
    generationConfigOption := domain.NewGeminiGenerationConfigOption().
        WithTemperature(0.7).
        WithTopK(40).
        WithMaxOutputTokens(1024)
    
    safetySettings := []map[string]interface{}{
        {
            "category": "HARM_CATEGORY_HARASSMENT",
            "threshold": "BLOCK_MEDIUM_AND_ABOVE",
        },
    }
    safetySettingsOption := domain.NewGeminiSafetySettingsOption(safetySettings)
    
    // Create provider with all options
    geminiProvider := provider.NewGeminiProvider(
        geminiKey,
        "gemini-2.0-flash-lite",
        httpClientOption,
        timeoutOption,
        generationConfigOption,
        safetySettingsOption,
    )
    
    // Use the provider
    response, err := geminiProvider.Generate(
        context.Background(),
        "Tell me about the Go programming language",
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Response: %s\n", response)
}
```

## API Reference

For a complete reference of all available provider options, see the [LLM API documentation](/docs/api/llm.md#provider-options).

## Option Implementation

If you're creating your own provider option, here's an example implementation:

```go
// Define a new option type
type CustomOption struct {
    Value string
}

// Create a constructor function
func NewCustomOption(value string) *CustomOption {
    return &CustomOption{Value: value}
}

// Implement the ProviderType method
func (o *CustomOption) ProviderType() string { 
    return "openai" // or "all" for common options
}

// Implement the provider-specific apply method
func (o *CustomOption) ApplyToOpenAI(provider interface{}) {
    if p, ok := provider.(interface{ SetCustomValue(value string) }); ok {
        p.SetCustomValue(o.Value)
    }
}
```

Then in your provider implementation, add the corresponding setter method:

```go
func (p *OpenAIProvider) SetCustomValue(value string) {
    p.customValue = value
}
```

This pattern ensures type safety and extensibility for all provider options.