# Provider Interface Documentation

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [API Reference](../../technical/api-reference) / Providers**

Complete API reference for LLM provider interfaces and implementations in Go-LLMs, covering core provider interfaces, provider-specific implementations, configuration options, and usage patterns.

## Core Provider Interfaces

### Provider Interface

The base interface that all LLM providers must implement (`pkg/llm/domain`):

```go
// Provider defines the contract that all LLM providers must implement.
type Provider interface {
    // Generate produces text from a prompt
    Generate(ctx context.Context, prompt string, options ...Option) (string, error)

    // GenerateMessage produces text from a list of messages
    GenerateMessage(ctx context.Context, messages []Message, options ...Option) (Response, error)

    // GenerateWithSchema produces structured output conforming to a schema
    GenerateWithSchema(ctx context.Context, prompt string, schema *schema.Schema, options ...Option) (interface{}, error)

    // Stream streams responses token by token
    Stream(ctx context.Context, prompt string, options ...Option) (ResponseStream, error)

    // StreamMessage streams responses from a list of messages
    StreamMessage(ctx context.Context, messages []Message, options ...Option) (ResponseStream, error)
}
```

`ResponseStream` is a read-only token channel: `type ResponseStream <-chan Token`

#### Methods

##### Generate

```go
Generate(ctx context.Context, prompt string, options ...Option) (string, error)
```

Generates text from a prompt string.

**Parameters:**
- `ctx`: Context for cancellation and timeout control
- `prompt`: The input prompt
- `options`: Optional provider options (temperature, max tokens, etc.)

**Returns:**
- `string`: The generated text
- `error`: Error if the request fails

**Example:**
```go
text, err := provider.Generate(ctx, "What is the capital of France?")
```

##### GenerateMessage

```go
GenerateMessage(ctx context.Context, messages []Message, options ...Option) (Response, error)
```

Generates a response from a list of messages (chat format).

**Parameters:**
- `ctx`: Context for cancellation and timeout control
- `messages`: Conversation messages
- `options`: Optional provider options

**Returns:**
- `Response`: The generated response
- `error`: Error if the request fails

##### GenerateWithSchema

```go
GenerateWithSchema(ctx context.Context, prompt string, schema *schema.Schema, options ...Option) (interface{}, error)
```

Generates structured output that conforms to the provided JSON schema.

##### Stream

```go
Stream(ctx context.Context, prompt string, options ...Option) (ResponseStream, error)
```

Streams response tokens as they are generated.

**Returns:**
- `ResponseStream`: A read-only `<-chan Token` channel
- `error`: Error if the stream setup fails

**Example:**
```go
stream, err := provider.Stream(ctx, "Explain Go interfaces")
if err != nil {
    return err
}
for token := range stream {
    fmt.Print(token.Text)
}
```

##### StreamMessage

```go
StreamMessage(ctx context.Context, messages []Message, options ...Option) (ResponseStream, error)
```

Streams a response from a list of messages.

### ModelRegistry Interface

Optional interface for registering and discovering models by name (`pkg/llm/domain`):

```go
type ModelRegistry interface {
    RegisterModel(name string, provider Provider) error
    GetModel(name string) (Provider, error)
    ListModels() []string
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
import "github.com/lexlapax/go-llms/pkg/llm/provider"

p := provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), "gpt-4")
text, err := p.Generate(ctx, "Hello!")
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
import "github.com/lexlapax/go-llms/pkg/llm/provider"

p := provider.NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"), "claude-3-opus-20240229")
text, err := p.Generate(ctx, "Write a haiku about programming.")
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

text, err := provider.Generate(ctx, prompt)
```

### 2. Error Handling

Handle provider-specific errors appropriately:

```go
if err != nil {
    var providerErr *domain.ProviderError
    if errors.As(err, &providerErr) {
        if providerErr.Retryable {
            // Implement retry logic
        }
    }
}
```

### 3. Streaming Best Practices

Handle streaming responses properly:

```go
stream, err := provider.Stream(ctx, prompt)
if err != nil {
    return err
}

var fullResponse strings.Builder
for token := range stream {
    fullResponse.WriteString(token.Text)
    fmt.Print(token.Text)
}
```

This comprehensive provider API documentation covers all aspects of working with LLM providers in Go-LLMs, providing the foundation for building robust applications with multiple LLM backends.