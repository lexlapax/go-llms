# Provider Interface Documentation

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [API Reference](../../technical/api-reference) / Providers**

Complete API reference for LLM provider interfaces and implementations in Go-LLMs, covering core provider interfaces, provider-specific implementations, configuration options, and usage patterns.

## Core Provider Interfaces

### Provider Interface

The base interface that all LLM providers must implement:

```go
package llm

// Provider defines the core interface for LLM providers
type Provider interface {
    // Complete generates a completion for the given request
    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
    
    // GetCapabilities returns the capabilities of this provider
    GetCapabilities() Capabilities
    
    // GetModels returns available models for this provider
    GetModels(ctx context.Context) ([]Model, error)
    
    // Close cleans up any resources used by the provider
    Close() error
}
```

#### Methods

##### Complete

```go
Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
```

Generates a completion based on the provided request.

**Parameters:**
- `ctx`: Context for cancellation and timeout control
- `request`: The completion request containing messages and parameters

**Returns:**
- `*CompletionResponse`: The generated completion
- `error`: Error if the request fails

**Example:**
```go
response, err := provider.Complete(ctx, &CompletionRequest{
    Messages: []Message{
        {Role: "user", Content: "What is the capital of France?"},
    },
    Model: "gpt-3.5-turbo",
}
```

##### GetCapabilities

```go
GetCapabilities() Capabilities
```

Returns the capabilities supported by this provider.

**Returns:**
- `Capabilities`: Provider capabilities including supported features

**Example:**
```go
caps := provider.GetCapabilities()
if caps.SupportsStreaming {
    // Use streaming
}
```

##### GetModels

```go
GetModels(ctx context.Context) ([]Model, error)
```

Retrieves the list of available models from the provider.

**Parameters:**
- `ctx`: Context for cancellation and timeout control

**Returns:**
- `[]Model`: List of available models
- `error`: Error if the request fails

### StreamingProvider Interface

Extends Provider with streaming capabilities:

```go
// StreamingProvider adds streaming support to the base Provider interface
type StreamingProvider interface {
    Provider
    
    // CompleteStream generates a streaming completion
    CompleteStream(ctx context.Context, request *CompletionRequest) (<-chan StreamChunk, error)
}
```

#### Methods

##### CompleteStream

```go
CompleteStream(ctx context.Context, request *CompletionRequest) (<-chan StreamChunk, error)
```

Generates a streaming completion, returning chunks as they become available.

**Parameters:**
- `ctx`: Context for cancellation and timeout control
- `request`: The completion request

**Returns:**
- `<-chan StreamChunk`: Channel of stream chunks
- `error`: Error if the stream setup fails

**Example:**
```go
stream, err := provider.CompleteStream(ctx, request)
if err != nil {
    return err
}

for chunk := range stream {
    if chunk.Error != nil {
        return chunk.Error
    }
    fmt.Print(chunk.Delta)
}
```

### EmbeddingProvider Interface

Provides text embedding generation capabilities:

```go
// EmbeddingProvider generates embeddings for text input
type EmbeddingProvider interface {
    // CreateEmbedding generates embeddings for the input text
    CreateEmbedding(ctx context.Context, request *EmbeddingRequest) (*EmbeddingResponse, error)
    
    // GetEmbeddingModels returns available embedding models
    GetEmbeddingModels(ctx context.Context) ([]EmbeddingModel, error)
}
```

#### Methods

##### CreateEmbedding

```go
CreateEmbedding(ctx context.Context, request *EmbeddingRequest) (*EmbeddingResponse, error)
```

Generates embeddings for the provided text input.

**Parameters:**
- `ctx`: Context for cancellation and timeout control
- `request`: The embedding request

**Returns:**
- `*EmbeddingResponse`: Generated embeddings
- `error`: Error if the request fails

### FunctionCallingProvider Interface

Adds function/tool calling capabilities:

```go
// FunctionCallingProvider supports function/tool calling
type FunctionCallingProvider interface {
    Provider
    
    // SupportsFunctionCalling indicates if the provider supports function calling
    SupportsFunctionCalling() bool
    
    // GetFunctionCallModels returns models that support function calling
    GetFunctionCallModels(ctx context.Context) ([]Model, error)
}
```

## Provider Implementations

### OpenAI Provider

```go
package openai

// Provider implements the OpenAI API
type Provider struct {
    config Config
    client *http.Client
}

// Config configures the OpenAI provider
type Config struct {
    APIKey         string        `json:"api_key"`
    BaseURL        string        `json:"base_url,omitempty"`
    Organization   string        `json:"organization,omitempty"`
    Model          string        `json:"model,omitempty"`
    DefaultModel   string        `json:"default_model,omitempty"`
    Timeout        time.Duration `json:"timeout,omitempty"`
    MaxRetries     int           `json:"max_retries,omitempty"`
    RetryDelay     time.Duration `json:"retry_delay,omitempty"`
}

// New creates a new OpenAI provider
func New(config Config) *Provider
```

#### Usage Example

```go
import "github.com/lexlapax/go-llms/pkg/llm/provider/openai"

provider := provider.NewOpenAIProvider(
    domain.NewTemperatureOption(&[]float64{0.7}[0]),
)
```

### Anthropic Provider

```go
package anthropic

// Provider implements the Anthropic Claude API
type Provider struct {
    config Config
    client *http.Client
}

// Config configures the Anthropic provider
type Config struct {
    APIKey       string        `json:"api_key"`
    BaseURL      string        `json:"base_url,omitempty"`
    Model        string        `json:"model,omitempty"`
    MaxTokens    int           `json:"max_tokens,omitempty"`
    Timeout      time.Duration `json:"timeout,omitempty"`
    Version      string        `json:"version,omitempty"`
}

// New creates a new Anthropic provider
func New(config Config) *Provider
```

#### Usage Example

```go
import "github.com/lexlapax/go-llms/pkg/llm/provider/anthropic"

provider := provider.NewAnthropicProvider(
    // APIKey: os.Getenv("ANTHROPIC_API_KEY"), // Moved to constructor parameters
    Model:  "claude-3-opus-20240229",
}

response, err := provider.Complete(ctx, &llm.CompletionRequest{
    Messages: []llm.Message{
        {Role: "user", Content: "Write a haiku about programming."},
    },
    MaxTokens: &[]int{100}[0],
}
```

### Google Gemini Provider

```go
package google

// Provider implements the Google Gemini API
type Provider struct {
    config Config
    client *genai.Client
}

// Config configures the Google provider
type Config struct {
    APIKey          string        `json:"api_key"`
    Model           string        `json:"model,omitempty"`
    Region          string        `json:"region,omitempty"`
    Timeout         time.Duration `json:"timeout,omitempty"`
    SafetySettings  []SafetySetting `json:"safety_settings,omitempty"`
}

// New creates a new Google Gemini provider
func New(config Config) *Provider
```

### Vertex AI Provider

```go
package vertexai

// Provider implements the Google Vertex AI API
type Provider struct {
    config Config
    client *aiplatform.PredictionClient
}

// Config configures the Vertex AI provider
type Config struct {
    ProjectID    string        `json:"project_id"`
    Location     string        `json:"location"`
    Model        string        `json:"model,omitempty"`
    Endpoint     string        `json:"endpoint,omitempty"`
    Credentials  string        `json:"credentials,omitempty"`
    Timeout      time.Duration `json:"timeout,omitempty"`
}

// New creates a new Vertex AI provider
func New(config Config) *Provider
```

### Ollama Provider

```go
package ollama

// Provider implements the Ollama local model API
type Provider struct {
    config Config
    client *http.Client
}

// Config configures the Ollama provider
type Config struct {
    BaseURL     string        `json:"base_url,omitempty"`
    Model       string        `json:"model"`
    Timeout     time.Duration `json:"timeout,omitempty"`
    KeepAlive   time.Duration `json:"keep_alive,omitempty"`
    NumPredict  int           `json:"num_predict,omitempty"`
    Temperature float64       `json:"temperature,omitempty"`
}

// New creates a new Ollama provider
func New(config Config) *Provider
```

### OpenRouter Provider

```go
package openrouter

// Provider implements the OpenRouter API
type Provider struct {
    config Config
    client *http.Client
}

// Config configures the OpenRouter provider
type Config struct {
    APIKey      string                 `json:"api_key"`
    BaseURL     string                 `json:"base_url,omitempty"`
    SiteURL     string                 `json:"site_url,omitempty"`
    SiteName    string                 `json:"site_name,omitempty"`
    Model       string                 `json:"model,omitempty"`
    Providers   []string               `json:"providers,omitempty"`
    Timeout     time.Duration          `json:"timeout,omitempty"`
    RouteParams map[string]interface{} `json:"route_params,omitempty"`
}

// New creates a new OpenRouter provider
func New(config Config) *Provider
```

## Request and Response Types

### CompletionRequest

```go
// CompletionRequest represents a request to generate a completion
type CompletionRequest struct {
    // Required fields
    Messages []Message `json:"messages"`
    
    // Model selection
    Model string `json:"model,omitempty"`
    
    // Generation parameters
    Temperature      *float64 `json:"temperature,omitempty"`
    TopP            *float64 `json:"top_p,omitempty"`
    MaxTokens       *int     `json:"max_tokens,omitempty"`
    StopSequences   []string `json:"stop,omitempty"`
    
    // Advanced features
    Stream          bool            `json:"stream,omitempty"`
    Tools           []ToolDefinition `json:"tools,omitempty"`
    ToolChoice      interface{}     `json:"tool_choice,omitempty"`
    ResponseFormat  *ResponseFormat `json:"response_format,omitempty"`
    
    // Provider-specific
    ProviderOptions map[string]interface{} `json:"provider_options,omitempty"`
    
    // Metadata
    User     string                 `json:"user,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents a conversation message
type Message struct {
    Role       string     `json:"role"`
    Content    string     `json:"content"`
    Name       string     `json:"name,omitempty"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
    ToolCallID string     `json:"tool_call_id,omitempty"`
}

// ToolDefinition defines a tool/function that can be called
type ToolDefinition struct {
    Type     string   `json:"type"`
    Function Function `json:"function"`
}

// Function defines a callable function
type Function struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Parameters  interface{} `json:"parameters"`
}
```

### CompletionResponse

```go
// CompletionResponse represents a completion response
type CompletionResponse struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Created int64  `json:"created"`
    Model   string `json:"model"`
    
    // Content
    Choices []Choice `json:"choices"`
    
    // Usage information
    Usage *Usage `json:"usage,omitempty"`
    
    // Provider-specific
    SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
    ProviderMetadata  map[string]interface{} `json:"provider_metadata,omitempty"`
}

// Choice represents a completion choice
type Choice struct {
    Index        int         `json:"index"`
    Message      Message     `json:"message"`
    FinishReason string      `json:"finish_reason"`
    LogProbs     interface{} `json:"logprobs,omitempty"`
}

// Usage represents token usage information
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

### StreamChunk

```go
// StreamChunk represents a chunk in a streaming response
type StreamChunk struct {
    ID      string    `json:"id"`
    Object  string    `json:"object"`
    Created int64     `json:"created"`
    Model   string    `json:"model"`
    Choices []StreamChoice `json:"choices"`
    Error   error     `json:"error,omitempty"`
}

// StreamChoice represents a choice in a streaming response
type StreamChoice struct {
    Index        int          `json:"index"`
    Delta        MessageDelta `json:"delta"`
    FinishReason string       `json:"finish_reason,omitempty"`
}

// MessageDelta represents incremental message content
type MessageDelta struct {
    Role      string     `json:"role,omitempty"`
    Content   string     `json:"content,omitempty"`
    ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}
```

## Provider Capabilities

```go
// Capabilities describes what a provider supports
type Capabilities struct {
    // Core features
    SupportsStreaming      bool `json:"supports_streaming"`
    SupportsFunctionCalling bool `json:"supports_function_calling"`
    SupportsEmbeddings     bool `json:"supports_embeddings"`
    SupportsImages         bool `json:"supports_images"`
    SupportsAudio          bool `json:"supports_audio"`
    
    // Model features
    SupportsSystemMessages bool `json:"supports_system_messages"`
    SupportsToolMessages   bool `json:"supports_tool_messages"`
    MaxContextLength       int  `json:"max_context_length"`
    
    // Advanced features
    SupportsLogProbs       bool `json:"supports_logprobs"`
    SupportsJSONMode       bool `json:"supports_json_mode"`
    SupportsSeed           bool `json:"supports_seed"`
    
    // Rate limits
    RequestsPerMinute      int  `json:"requests_per_minute,omitempty"`
    TokensPerMinute        int  `json:"tokens_per_minute,omitempty"`
    ConcurrentRequests     int  `json:"concurrent_requests,omitempty"`
}
```

## Model Information

```go
// Model represents an available model
type Model struct {
    ID               string    `json:"id"`
    Object           string    `json:"object"`
    Created          int64     `json:"created"`
    OwnedBy          string    `json:"owned_by"`
    
    // Capabilities
    ContextLength    int       `json:"context_length"`
    MaxOutputTokens  int       `json:"max_output_tokens,omitempty"`
    TrainingCutoff   time.Time `json:"training_cutoff,omitempty"`
    
    // Features
    SupportsFunctions bool     `json:"supports_functions"`
    SupportsVision    bool     `json:"supports_vision"`
    SupportsTools     bool     `json:"supports_tools"`
    
    // Pricing (optional)
    InputCostPer1K   float64  `json:"input_cost_per_1k,omitempty"`
    OutputCostPer1K  float64  `json:"output_cost_per_1k,omitempty"`
}
```

## Error Handling

### Provider Errors

```go
// ProviderError represents a provider-specific error
type ProviderError struct {
    Provider     string                 `json:"provider"`
    Code         string                 `json:"code"`
    Message      string                 `json:"message"`
    StatusCode   int                    `json:"status_code,omitempty"`
    Details      map[string]interface{} `json:"details,omitempty"`
    Retryable    bool                   `json:"retryable"`
    RetryAfter   time.Duration          `json:"retry_after,omitempty"`
}

// Error implements the error interface
func (e *ProviderError) Error() string

// Common error codes
const (
    ErrCodeRateLimit       = "rate_limit_exceeded"
    ErrCodeQuotaExceeded   = "quota_exceeded"
    ErrCodeInvalidRequest  = "invalid_request"
    ErrCodeAuthentication  = "authentication_failed"
    ErrCodeModelNotFound   = "model_not_found"
    ErrCodeContextTooLong  = "context_length_exceeded"
    ErrCodeServerError     = "server_error"
    ErrCodeTimeout         = "timeout"
)
```

### Error Handling Example

```go
response, err := provider.Complete(ctx, request)
if err != nil {
    var providerErr *ProviderError
    if errors.As(err, &providerErr) {
        switch providerErr.Code {
        case ErrCodeRateLimit:
            // Wait and retry
            time.Sleep(providerErr.RetryAfter)
            return retry(ctx, request)
        case ErrCodeAuthentication:
            // Handle auth error
            return nil, fmt.Errorf("authentication failed: %w", err)
        default:
            // Handle other errors
            return nil, err
        }
    }
    return nil, err
}
```

## Provider Options

### Common Options

```go
// CommonOptions are supported by most providers
type CommonOptions struct {
    // Request options
    Timeout        time.Duration `json:"timeout,omitempty"`
    MaxRetries     int           `json:"max_retries,omitempty"`
    RetryDelay     time.Duration `json:"retry_delay,omitempty"`
    
    // HTTP options
    HTTPClient     *http.Client  `json:"-"`
    HTTPHeaders    map[string]string `json:"http_headers,omitempty"`
    
    // Debug options
    Debug          bool          `json:"debug,omitempty"`
    LogLevel       string        `json:"log_level,omitempty"`
}
```

### Provider-Specific Options

```go
// OpenAI-specific options
type OpenAIOptions struct {
    CommonOptions
    Organization    string  `json:"organization,omitempty"`
    User           string  `json:"user,omitempty"`
    LogProbs       bool    `json:"logprobs,omitempty"`
    TopLogProbs    int     `json:"top_logprobs,omitempty"`
    Echo           bool    `json:"echo,omitempty"`
    PresencePenalty float64 `json:"presence_penalty,omitempty"`
    FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
}

// Anthropic-specific options
type AnthropicOptions struct {
    CommonOptions
    Version        string   `json:"version,omitempty"`
    TopK           int      `json:"top_k,omitempty"`
    StopSequences  []string `json:"stop_sequences,omitempty"`
}
```

## Best Practices

### 1. Context Management

Always use context for proper cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := provider.Complete(ctx, request)
```

### 2. Error Handling

Handle provider-specific errors appropriately:

```go
if err != nil {
    var providerErr *ProviderError
    if errors.As(err, &providerErr) {
        if providerErr.Retryable {
            // Implement retry logic
        }
    }
}
```

### 3. Resource Cleanup

Always close providers when done:

```go
provider := openai.New(config)
defer provider.Close()
```

### 4. Model Selection

Check model availability before use:

```go
models, err := provider.GetModels(ctx)
if err != nil {
    return err
}

// Find suitable model
var selectedModel *Model
for _, model := range models {
    if model.SupportsFunctions && model.ContextLength >= 8000 {
        selectedModel = &model
        break
    }
}
```

### 5. Streaming Best Practices

Handle streaming responses properly:

```go
stream, err := provider.CompleteStream(ctx, request)
if err != nil {
    return err
}

var fullResponse strings.Builder
for chunk := range stream {
    if chunk.Error != nil {
        return chunk.Error
    }
    
    for _, choice := range chunk.Choices {
        fullResponse.WriteString(choice.Delta.Content)
        
        // Process chunk immediately if needed
        fmt.Print(choice.Delta.Content)
    }
}
```

This comprehensive provider API documentation covers all aspects of working with LLM providers in Go-LLMs, providing the foundation for building robust applications with multiple LLM backends.