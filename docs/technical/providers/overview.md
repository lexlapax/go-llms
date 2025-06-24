# Provider System Overview

> **[Documentation Home](../README.md) / [Providers](README.md) / Overview**

## Introduction

The provider system is the foundation of go-llms, offering a unified interface for interacting with various Large Language Model services. This abstraction allows applications to switch between providers with minimal code changes while maintaining provider-specific optimizations.

## Architecture

### Provider Interface

The core `Provider` interface defines the contract all implementations must fulfill:

```go
type Provider interface {
    // Basic text generation
    Generate(ctx context.Context, prompt string, options ...Option) (Response, error)
    
    // Generation with message history
    GenerateMessage(ctx context.Context, messages []Message, options ...Option) (Response, error)
    
    // Structured output with schema validation
    GenerateWithSchema(ctx context.Context, prompt string, schema *schema.Schema, options ...Option) (any, error)
    
    // Streaming responses
    Stream(ctx context.Context, prompt string, options ...Option) (<-chan StreamResponse, error)
    StreamMessage(ctx context.Context, messages []Message, options ...Option) (<-chan StreamResponse, error)
}
```

### MetadataProvider Interface

Providers that support capability discovery implement the `MetadataProvider` interface:

```go
type MetadataProvider interface {
    GetMetadata() ProviderMetadata
}
```

## Provider Lifecycle

### 1. Initialization
```go
provider := provider.NewOpenAIProvider(apiKey, model, options...)
```

During initialization:
- API credentials are validated
- HTTP client is configured
- Base URLs are set
- Default options are applied

### 2. Configuration
Providers support configuration through functional options:

```go
provider := provider.NewOpenAIProvider(apiKey, model,
    domain.WithTimeout(30 * time.Second),
    domain.WithMaxRetries(3),
    domain.WithHTTPClient(customClient),
    provider.WithOrganization(orgID), // Provider-specific option
)
```

### 3. Request Execution
When a request is made:

```
Application Request
    ↓
Validate Options
    ↓
Build API Request
    ↓
Apply Retry Logic
    ↓
Send HTTP Request
    ↓
Parse Response
    ↓
Handle Errors
    ↓
Return Result
```

### 4. Error Handling
Providers translate API-specific errors into standardized error types:

![Provider Registry System](../images/provider-registry.svg)
*Figure 1: Provider registration and discovery architecture showing how providers are managed and selected*

```go
type ProviderError struct {
    Provider string
    Type     ErrorType
    Message  string
    Details  map[string]interface{}
}
```

## Message Format

### Message Structure
```go
type Message struct {
    Role    MessageRole    // user, assistant, system
    Content []ContentPart  // Text, images, files
}

type ContentPart struct {
    Type ContentType // text, image_url, file
    Text string
    // Additional fields for different content types
}
```

### Message Roles
- **System**: Sets behavior and context
- **User**: Input from the user
- **Assistant**: Previous model responses
- **Tool**: Results from function calls (OpenAI/Anthropic)

## Options System

### Common Options
All providers support these standard options:

```go
// Model selection
domain.WithModel("gpt-4-turbo-preview")

// Temperature control
domain.WithTemperature(0.7)

// Token limits
domain.WithMaxTokens(2000)

// Response format
domain.WithJSONMode(true)

// Timeout
domain.WithTimeout(30 * time.Second)
```

### Provider-Specific Options
Each provider may have unique options:

```go
// OpenAI specific
provider.WithOrganization("org-123")
provider.WithTools(tools...)

// Anthropic specific
provider.WithMaxRetries(5)
provider.WithBetaFeatures("computer-use")

// Gemini specific
provider.WithSafetySettings(settings...)
```

## Streaming

### Stream Response Structure
```go
type StreamResponse struct {
    Content  string // Incremental content
    Done     bool   // Stream completed
    Error    error  // Any error encountered
    Metadata map[string]interface{} // Optional metadata
}
```

### Stream Usage Pattern
```go
stream, err := provider.Stream(ctx, prompt)
if err != nil {
    return err
}

for response := range stream {
    if response.Error != nil {
        return response.Error
    }
    
    fmt.Print(response.Content)
    
    if response.Done {
        break
    }
}
```

## Schema Validation

### Structured Output Generation
Providers can generate outputs conforming to JSON schemas:

```go
schema := &schema.Schema{
    Type: "object",
    Properties: map[string]schema.Property{
        "name": {Type: "string"},
        "age":  {Type: "integer"},
    },
    Required: []string{"name"},
}

result, err := provider.GenerateWithSchema(ctx, prompt, schema)
```

### Schema Enforcement Strategies
Different providers use different approaches:
- **OpenAI**: JSON mode + prompt engineering
- **Anthropic**: Strong prompt engineering
- **Gemini**: Response validation + retry

## Rate Limiting and Retry

### Built-in Retry Logic
Providers implement exponential backoff for transient errors:

```go
// Retry configuration
type RetryConfig struct {
    MaxRetries     int
    InitialDelay   time.Duration
    MaxDelay       time.Duration
    Multiplier     float64
    RetryableErrors []ErrorType
}
```

### Rate Limit Handling
```go
response, err := provider.Generate(ctx, prompt)
if err != nil {
    if providerErr, ok := err.(*ProviderError); ok {
        if providerErr.Type == ErrorTypeRateLimit {
            // Extract retry-after from Details
            retryAfter := providerErr.Details["retry_after"].(int)
            time.Sleep(time.Duration(retryAfter) * time.Second)
            // Retry request
        }
    }
}
```

## Provider Selection Considerations

### Performance Factors
- **Latency**: Gemini Flash < OpenAI GPT-3.5 < Anthropic Claude
- **Throughput**: Depends on rate limits and model
- **Context Length**: Anthropic (200k) > Gemini (1M) > OpenAI (128k)

### Feature Requirements
- **Function Calling**: OpenAI and Anthropic have best support
- **Vision**: All major providers support multimodal inputs
- **Streaming**: Universal support with provider-specific optimizations

### Cost Optimization
- **Token Pricing**: Varies significantly between providers
- **Batch Processing**: Some providers offer discounts
- **Caching**: Anthropic offers prompt caching

## Best Practices

### 1. Provider Initialization
```go
// Store providers as singletons
var (
    openaiProvider    domain.Provider
    anthropicProvider domain.Provider
    providerOnce      sync.Once
)

func GetProvider() domain.Provider {
    providerOnce.Do(func() {
        openaiProvider = provider.NewOpenAIProvider(
            os.Getenv("OPENAI_API_KEY"),
            "gpt-4",
        )
}
    return openaiProvider
}
```

### 2. Error Handling
```go
func handleProviderError(err error) error {
    var providerErr *ProviderError
    if errors.As(err, &providerErr) {
        log.Printf("Provider error: %s - %s", providerErr.Provider, providerErr.Type)
        
        switch providerErr.Type {
        case ErrorTypeRateLimit:
            return fmt.Errorf("rate limited, please retry later")
        case ErrorTypeContextLength:
            return fmt.Errorf("input too long, please shorten")
        default:
            return fmt.Errorf("provider error: %s", providerErr.Message)
        }
    }
    return err
}
```

### 3. Multi-Provider Fallback
```go
providers := []domain.Provider{
    primaryProvider,
    fallbackProvider,
}

for _, p := range providers {
    response, err := p.Generate(ctx, prompt)
    if err == nil {
        return response, nil
    }
    log.Printf("Provider failed: %v, trying next", err)
}
return nil, fmt.Errorf("all providers failed")
```

## Testing with Providers

### Mock Provider
```go
mockProvider := provider.NewMockProvider()
mockProvider.WithGenerateFunc(func(ctx context.Context, prompt string, options ...Option) (Response, error) {
    return Response{Content: "Mock response"}, nil
}
```

### Provider Testing Pattern
```go
func TestMyFunction(t *testing.T) {
    // Use mock for unit tests
    provider := provider.NewMockProvider()
    provider.AddResponse("Expected response")
    
    // Test your function
    result := MyFunction(provider)
    
    // Verify calls
    calls := provider.GetCalls()
    assert.Equal(t, 1, len(calls))
}
```

## Next Steps

- Learn how to [Implement Custom Providers](implementing-providers.md)
- Explore the [Provider Registry](provider-registry.md) for dynamic management
- Understand [Provider Metadata](metadata.md) for capability discovery
- See [API Reference](../api-reference/providers.md) for detailed documentation