# API Design: Design Principles and Patterns

> **[Project Root](/) / [Documentation](/docs/) / [Technical Documentation](/docs/technical/) / [Development](/docs/technical/development/) / API Design**

Comprehensive guide to API design principles and patterns used in Go-LLMs, covering interface design, package architecture, error handling strategies, configuration patterns, extensibility mechanisms, and best practices for building maintainable, user-friendly APIs.

## Core Design Principles

### 1. Simplicity and Clarity

**Principle**: APIs should be simple to understand and use correctly, with clear intent and minimal cognitive overhead.

```go
// Good: Clear, focused interface
type Provider interface {
    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
}

// Bad: Complex, unfocused interface
type ComplexProvider interface {
    CompleteWithModel(ctx context.Context, model string, prompt string, temp float64, maxTokens int, stop []string) (string, map[string]interface{}, error)
    CompleteWithFullOptions(ctx context.Context, options map[string]interface{}) (interface{}, error)
}
```

**Implementation Guidelines:**
- One responsibility per interface
- Self-documenting method names
- Consistent parameter ordering
- Minimal required parameters

### 2. Composability and Modularity

**Principle**: Components should work together seamlessly and be easily combinable.

```go
// Composable agent architecture
type Agent interface {
    Execute(ctx context.Context, input interface{}) (interface{}, error)
}

type ToolEnabledAgent struct {
    Agent
    tools ToolRegistry
}

type StatefulAgent struct {
    Agent
    state StateManager
}

type WorkflowAgent struct {
    Agent
    orchestrator WorkflowOrchestrator
}

// Composition example
func NewAdvancedAgent(base Agent, tools ToolRegistry, state StateManager) Agent {
    return &AdvancedAgent{
        ToolEnabledAgent: &ToolEnabledAgent{Agent: base, tools: tools},
        state:           state,
    }
}
```

### 3. Extensibility and Flexibility

**Principle**: APIs should be extensible without breaking existing code.

```go
// Extensible configuration with options pattern
type ProviderConfig struct {
    APIKey  string
    BaseURL string
    Model   string
    // Core required fields
}

type ProviderOption func(*ProviderConfig)

func WithModel(model string) ProviderOption {
    return func(c *ProviderConfig) {
        c.Model = model
    }
}

func WithTimeout(timeout time.Duration) ProviderOption {
    return func(c *ProviderConfig) {
        c.Timeout = timeout
    }
}

func WithRetry(config RetryConfig) ProviderOption {
    return func(c *ProviderConfig) {
        c.RetryConfig = &config
    }
}

// Usage allows easy extension
provider := NewProvider("api-key", 
    WithModel("gpt-4"),
    WithTimeout(30*time.Second),
    WithRetry(RetryConfig{MaxAttempts: 3}),
)
```

### 4. Consistent Error Handling

**Principle**: Errors should be consistent, informative, and actionable.

```go
// Structured error hierarchy
type ErrorType string

const (
    ErrorTypeValidation   ErrorType = "validation"
    ErrorTypeAuth        ErrorType = "authentication"
    ErrorTypeRateLimit   ErrorType = "rate_limit"
    ErrorTypeNetwork     ErrorType = "network"
    ErrorTypeProvider    ErrorType = "provider"
    ErrorTypeInternal    ErrorType = "internal"
)

type APIError struct {
    Type       ErrorType              `json:"type"`
    Code       string                 `json:"code"`
    Message    string                 `json:"message"`
    Details    map[string]interface{} `json:"details,omitempty"`
    Retryable  bool                   `json:"retryable"`
    Cause      error                  `json:"-"`
    Timestamp  time.Time              `json:"timestamp"`
}

func (e *APIError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *APIError) Unwrap() error {
    return e.Cause
}

// Error constructors for consistency
func NewValidationError(field, message string) *APIError {
    return &APIError{
        Type:      ErrorTypeValidation,
        Code:      "VALIDATION_FAILED",
        Message:   fmt.Sprintf("validation failed for field '%s': %s", field, message),
        Retryable: false,
        Timestamp: time.Now(),
        Details: map[string]interface{}{
            "field": field,
        },
    }
}

func NewProviderError(provider, code, message string, retryable bool) *APIError {
    return &APIError{
        Type:      ErrorTypeProvider,
        Code:      code,
        Message:   message,
        Retryable: retryable,
        Timestamp: time.Now(),
        Details: map[string]interface{}{
            "provider": provider,
        },
    }
}
```

## Interface Design Patterns

### 1. Core Interface Hierarchy

```go
// Base interfaces define essential contracts
type Provider interface {
    // Core functionality
    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
    
    // Metadata
    GetCapabilities() Capabilities
    GetModels(ctx context.Context) ([]Model, error)
}

// Extended interfaces add optional functionality
type StreamingProvider interface {
    Provider
    CompleteStream(ctx context.Context, request *CompletionRequest) (<-chan StreamChunk, error)
}

type FunctionCallingProvider interface {
    Provider
    SupportsFunctionCalling() bool
    GetFunctionSchema() []FunctionSchema
}

type EmbeddingProvider interface {
    CreateEmbedding(ctx context.Context, request *EmbeddingRequest) (*EmbeddingResponse, error)
}

// Composition for full-featured providers
type AdvancedProvider interface {
    Provider
    StreamingProvider
    FunctionCallingProvider
    EmbeddingProvider
}
```

### 2. Builder Pattern for Complex Objects

```go
// Builder for complex configuration
type AgentBuilder struct {
    config      AgentConfig
    tools       []Tool
    middleware  []Middleware
    interceptors []Interceptor
    validators  []Validator
}

func NewAgentBuilder() *AgentBuilder {
    return &AgentBuilder{
        config: DefaultAgentConfig(),
    }
}

func (b *AgentBuilder) WithProvider(provider Provider) *AgentBuilder {
    b.config.Provider = provider
    return b
}

func (b *AgentBuilder) WithTool(tool Tool) *AgentBuilder {
    b.tools = append(b.tools, tool)
    return b
}

func (b *AgentBuilder) WithMiddleware(middleware Middleware) *AgentBuilder {
    b.middleware = append(b.middleware, middleware)
    return b
}

func (b *AgentBuilder) WithSystemPrompt(prompt string) *AgentBuilder {
    b.config.SystemPrompt = prompt
    return b
}

func (b *AgentBuilder) WithMemory(memory AgentMemory) *AgentBuilder {
    b.config.Memory = memory
    return b
}

func (b *AgentBuilder) Build() (Agent, error) {
    if err := b.validate(); err != nil {
        return nil, fmt.Errorf("agent configuration validation failed: %w", err)
    }
    
    agent := &DefaultAgent{
        config:      b.config,
        tools:       NewToolRegistry(),
        middleware:  b.middleware,
        interceptors: b.interceptors,
    }
    
    // Register tools
    for _, tool := range b.tools {
        if err := agent.tools.Register(tool); err != nil {
            return nil, fmt.Errorf("failed to register tool %s: %w", tool.Name(), err)
        }
    }
    
    return agent, nil
}

// Usage example
agent, err := NewAgentBuilder().
    WithProvider(openaiProvider).
    WithTool(httpTool).
    WithTool(fileTool).
    WithSystemPrompt("You are a helpful assistant").
    WithMemory(conversationMemory).
    Build()
```

### 3. Plugin Architecture

```go
// Plugin interface for extensibility
type Plugin interface {
    Name() string
    Version() string
    Initialize(ctx context.Context, config map[string]interface{}) error
    Shutdown(ctx context.Context) error
}

// Specific plugin types
type ProviderPlugin interface {
    Plugin
    CreateProvider(config ProviderConfig) (Provider, error)
}

type ToolPlugin interface {
    Plugin
    GetTools() []Tool
}

type MiddlewarePlugin interface {
    Plugin
    CreateMiddleware(config map[string]interface{}) (Middleware, error)
}

// Plugin registry with discovery
type PluginRegistry struct {
    plugins     map[string]Plugin
    factories   map[string]PluginFactory
    mu          sync.RWMutex
}

type PluginFactory func(config map[string]interface{}) (Plugin, error)

func (r *PluginRegistry) RegisterFactory(pluginType string, factory PluginFactory) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.factories[pluginType] = factory
}

func (r *PluginRegistry) CreatePlugin(pluginType string, config map[string]interface{}) (Plugin, error) {
    r.mu.RLock()
    factory, exists := r.factories[pluginType]
    r.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("unknown plugin type: %s", pluginType)
    }
    
    return factory(config)
}

// Plugin discovery and loading
func (r *PluginRegistry) DiscoverPlugins(searchPaths []string) error {
    for _, path := range searchPaths {
        if err := r.discoverInPath(path); err != nil {
            return fmt.Errorf("plugin discovery failed in %s: %w", path, err)
        }
    }
    return nil
}
```

## Configuration Design Patterns

### 1. Hierarchical Configuration

```go
// Configuration hierarchy with inheritance
type Config struct {
    Global   GlobalConfig   `yaml:"global" json:"global"`
    Providers ProviderConfigs `yaml:"providers" json:"providers"`
    Agents   AgentConfigs   `yaml:"agents" json:"agents"`
    Tools    ToolConfigs    `yaml:"tools" json:"tools"`
}

type GlobalConfig struct {
    LogLevel     string        `yaml:"log_level" json:"log_level" default:"info"`
    Timeout      time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
    MaxRetries   int           `yaml:"max_retries" json:"max_retries" default:"3"`
    RateLimit    *RateLimit    `yaml:"rate_limit,omitempty" json:"rate_limit,omitempty"`
    Security     SecurityConfig `yaml:"security" json:"security"`
}

type ProviderConfigs map[string]ProviderConfig

type ProviderConfig struct {
    Type        string                 `yaml:"type" json:"type" validate:"required"`
    APIKey      string                 `yaml:"api_key" json:"api_key" validate:"required"`
    BaseURL     string                 `yaml:"base_url,omitempty" json:"base_url,omitempty"`
    Model       string                 `yaml:"model" json:"model" validate:"required"`
    Timeout     *time.Duration         `yaml:"timeout,omitempty" json:"timeout,omitempty"`
    MaxRetries  *int                   `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`
    Options     map[string]interface{} `yaml:"options,omitempty" json:"options,omitempty"`
}

// Configuration inheritance and merging
func (c *ProviderConfig) MergeWithGlobal(global GlobalConfig) {
    if c.Timeout == nil {
        c.Timeout = &global.Timeout
    }
    if c.MaxRetries == nil {
        c.MaxRetries = &global.MaxRetries
    }
}

// Environment variable override support
type ConfigLoader struct {
    envPrefix   string
    configPaths []string
    validators  []ConfigValidator
}

func (l *ConfigLoader) Load() (*Config, error) {
    config := &Config{}
    
    // Load from files
    for _, path := range l.configPaths {
        if err := l.loadFromFile(path, config); err != nil {
            return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
        }
    }
    
    // Override with environment variables
    if err := l.loadFromEnv(config); err != nil {
        return nil, fmt.Errorf("failed to load environment overrides: %w", err)
    }
    
    // Apply defaults
    l.applyDefaults(config)
    
    // Validate configuration
    if err := l.validate(config); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return config, nil
}
```

### 2. Type-Safe Configuration

```go
// Strongly typed configuration with validation
type ModelConfig struct {
    Name         string    `yaml:"name" json:"name" validate:"required"`
    Temperature  *float64  `yaml:"temperature,omitempty" json:"temperature,omitempty" validate:"omitempty,min=0,max=2"`
    MaxTokens    *int      `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty" validate:"omitempty,min=1"`
    TopP         *float64  `yaml:"top_p,omitempty" json:"top_p,omitempty" validate:"omitempty,min=0,max=1"`
    StopSequences []string `yaml:"stop_sequences,omitempty" json:"stop_sequences,omitempty"`
}

// Configuration validation
func (c *ModelConfig) Validate() error {
    if c.Name == "" {
        return NewValidationError("name", "model name is required")
    }
    
    if c.Temperature != nil {
        if *c.Temperature < 0 || *c.Temperature > 2 {
            return NewValidationError("temperature", "must be between 0 and 2")
        }
    }
    
    if c.MaxTokens != nil {
        if *c.MaxTokens < 1 {
            return NewValidationError("max_tokens", "must be greater than 0")
        }
    }
    
    return nil
}

// Configuration with defaults
func DefaultModelConfig() ModelConfig {
    return ModelConfig{
        Name:        "gpt-3.5-turbo",
        Temperature: &[]float64{0.7}[0],
        MaxTokens:   &[]int{1000}[0],
        TopP:        &[]float64{1.0}[0],
    }
}
```

## Request/Response Design

### 1. Consistent Request Structure

```go
// Base request structure with common fields
type BaseRequest struct {
    ID        string                 `json:"id,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Options   RequestOptions         `json:"options,omitempty"`
    Timestamp time.Time              `json:"timestamp,omitempty"`
}

type RequestOptions struct {
    Timeout   time.Duration `json:"timeout,omitempty"`
    Priority  Priority      `json:"priority,omitempty"`
    Async     bool          `json:"async,omitempty"`
    Streaming bool          `json:"streaming,omitempty"`
}

// Specific request types extend base
type CompletionRequest struct {
    BaseRequest
    Messages       []Message  `json:"messages" validate:"required,min=1"`
    Model         string     `json:"model,omitempty"`
    Temperature   *float64   `json:"temperature,omitempty"`
    MaxTokens     *int       `json:"max_tokens,omitempty"`
    Tools         []Tool     `json:"tools,omitempty"`
    ToolChoice    string     `json:"tool_choice,omitempty"`
}

type EmbeddingRequest struct {
    BaseRequest
    Input []string `json:"input" validate:"required,min=1"`
    Model string   `json:"model,omitempty"`
}
```

### 2. Structured Response Format

```go
// Base response with common fields
type BaseResponse struct {
    ID        string                 `json:"id"`
    Success   bool                   `json:"success"`
    Error     *APIError              `json:"error,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
    Duration  time.Duration          `json:"duration"`
}

// Specific response types
type CompletionResponse struct {
    BaseResponse
    Content      string       `json:"content"`
    Model        string       `json:"model"`
    Usage        UsageInfo    `json:"usage"`
    ToolCalls    []ToolCall   `json:"tool_calls,omitempty"`
    FinishReason string       `json:"finish_reason"`
}

type EmbeddingResponse struct {
    BaseResponse
    Embeddings [][]float64 `json:"embeddings"`
    Model      string      `json:"model"`
    Usage      UsageInfo   `json:"usage"`
}

type UsageInfo struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

### 3. Streaming Response Design

```go
// Streaming response with typed chunks
type StreamChunk interface {
    Type() ChunkType
    Timestamp() time.Time
}

type ChunkType string

const (
    ChunkTypeContent   ChunkType = "content"
    ChunkTypeToolCall  ChunkType = "tool_call"
    ChunkTypeMetadata  ChunkType = "metadata"
    ChunkTypeError     ChunkType = "error"
    ChunkTypeComplete  ChunkType = "complete"
)

type ContentChunk struct {
    Type      ChunkType `json:"type"`
    Content   string    `json:"content"`
    Delta     string    `json:"delta"`
    Index     int       `json:"index"`
    Timestamp time.Time `json:"timestamp"`
}

type ToolCallChunk struct {
    Type      ChunkType `json:"type"`
    ToolCall  ToolCall  `json:"tool_call"`
    Complete  bool      `json:"complete"`
    Timestamp time.Time `json:"timestamp"`
}

type ErrorChunk struct {
    Type      ChunkType `json:"type"`
    Error     APIError  `json:"error"`
    Fatal     bool      `json:"fatal"`
    Timestamp time.Time `json:"timestamp"`
}

// Streaming interface
type StreamReader interface {
    Read() (StreamChunk, error)
    Close() error
}

// Usage pattern
func ProcessStream(stream StreamReader) error {
    defer stream.Close()
    
    for {
        chunk, err := stream.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("stream read error: %w", err)
        }
        
        switch chunk.Type() {
        case ChunkTypeContent:
            content := chunk.(*ContentChunk)
            fmt.Print(content.Delta)
            
        case ChunkTypeToolCall:
            toolCall := chunk.(*ToolCallChunk)
            if toolCall.Complete {
                executeToolCall(toolCall.ToolCall)
            }
            
        case ChunkTypeError:
            errorChunk := chunk.(*ErrorChunk)
            if errorChunk.Fatal {
                return fmt.Errorf("fatal stream error: %v", errorChunk.Error)
            }
            log.Printf("Stream warning: %v", errorChunk.Error)
            
        case ChunkTypeComplete:
            return nil
        }
    }
    
    return nil
}
```

## Middleware and Interceptor Patterns

### 1. Middleware Architecture

```go
// Middleware interface for request/response processing
type Middleware interface {
    Name() string
    Process(ctx context.Context, request interface{}, next MiddlewareFunc) (interface{}, error)
}

type MiddlewareFunc func(ctx context.Context, request interface{}) (interface{}, error)

// Middleware chain
type MiddlewareChain struct {
    middlewares []Middleware
}

func (c *MiddlewareChain) Add(middleware Middleware) {
    c.middlewares = append(c.middlewares, middleware)
}

func (c *MiddlewareChain) Execute(ctx context.Context, request interface{}, final MiddlewareFunc) (interface{}, error) {
    if len(c.middlewares) == 0 {
        return final(ctx, request)
    }
    
    // Build chain from right to left
    handler := final
    for i := len(c.middlewares) - 1; i >= 0; i-- {
        middleware := c.middlewares[i]
        currentHandler := handler
        handler = func(ctx context.Context, req interface{}) (interface{}, error) {
            return middleware.Process(ctx, req, currentHandler)
        }
    }
    
    return handler(ctx, request)
}

// Built-in middleware examples
type LoggingMiddleware struct {
    logger Logger
}

func (m *LoggingMiddleware) Name() string {
    return "logging"
}

func (m *LoggingMiddleware) Process(ctx context.Context, request interface{}, next MiddlewareFunc) (interface{}, error) {
    start := time.Now()
    
    m.logger.Debug("request started", "request", request)
    
    response, err := next(ctx, request)
    
    duration := time.Since(start)
    if err != nil {
        m.logger.Error("request failed", "duration", duration, "error", err)
    } else {
        m.logger.Info("request completed", "duration", duration)
    }
    
    return response, err
}

type RateLimitingMiddleware struct {
    limiter RateLimiter
}

func (m *RateLimitingMiddleware) Name() string {
    return "rate_limiting"
}

func (m *RateLimitingMiddleware) Process(ctx context.Context, request interface{}, next MiddlewareFunc) (interface{}, error) {
    if err := m.limiter.Wait(ctx); err != nil {
        return nil, NewRateLimitError("rate limit exceeded")
    }
    
    return next(ctx, request)
}
```

### 2. Interceptor Pattern

```go
// Interceptor for cross-cutting concerns
type Interceptor interface {
    Before(ctx context.Context, request interface{}) (context.Context, interface{}, error)
    After(ctx context.Context, request interface{}, response interface{}, err error) (interface{}, error)
}

// Metrics interceptor
type MetricsInterceptor struct {
    collector MetricsCollector
}

func (i *MetricsInterceptor) Before(ctx context.Context, request interface{}) (context.Context, interface{}, error) {
    // Add metrics context
    ctx = context.WithValue(ctx, "metrics_start", time.Now())
    
    // Increment request counter
    i.collector.IncrementCounter("requests_total", getRequestLabels(request))
    
    return ctx, request, nil
}

func (i *MetricsInterceptor) After(ctx context.Context, request interface{}, response interface{}, err error) (interface{}, error) {
    start := ctx.Value("metrics_start").(time.Time)
    duration := time.Since(start)
    
    labels := getRequestLabels(request)
    if err != nil {
        labels["status"] = "error"
        i.collector.IncrementCounter("requests_errors", labels)
    } else {
        labels["status"] = "success"
    }
    
    i.collector.RecordDuration("request_duration", duration, labels)
    
    return response, err
}
```

## Versioning and Compatibility

### 1. API Versioning Strategy

```go
// Version-aware interfaces
type ProviderV1 interface {
    Complete(ctx context.Context, request *CompletionRequestV1) (*CompletionResponseV1, error)
}

type ProviderV2 interface {
    ProviderV1 // Backward compatibility
    CompleteWithTools(ctx context.Context, request *CompletionRequestV2) (*CompletionResponseV2, error)
}

// Version negotiation
type VersionedProvider struct {
    v1 ProviderV1
    v2 ProviderV2
}

func (p *VersionedProvider) GetSupportedVersions() []string {
    versions := []string{"v1"}
    if p.v2 != nil {
        versions = append(versions, "v2")
    }
    return versions
}

func (p *VersionedProvider) CompleteV1(ctx context.Context, request *CompletionRequestV1) (*CompletionResponseV1, error) {
    if p.v2 != nil {
        // Convert to v2 request
        v2Request := convertV1ToV2Request(request)
        v2Response, err := p.v2.CompleteWithTools(ctx, v2Request)
        if err != nil {
            return nil, err
        }
        // Convert back to v1 response
        return convertV2ToV1Response(v2Response), nil
    }
    
    return p.v1.Complete(ctx, request)
}
```

### 2. Deprecation Handling

```go
// Deprecation warnings
type DeprecatedAPI struct {
    replacement string
    removeIn    string
    logger      Logger
}

func (d *DeprecatedAPI) WarnDeprecation(method string) {
    d.logger.Warn("deprecated API usage",
        "method", method,
        "replacement", d.replacement,
        "remove_in", d.removeIn,
    )
}

// Deprecated methods with warnings
func (p *Provider) LegacyComplete(prompt string) (string, error) {
    deprecation := &DeprecatedAPI{
        replacement: "Complete(ctx, request)",
        removeIn:    "v1.0.0",
        logger:      p.logger,
    }
    deprecation.WarnDeprecation("LegacyComplete")
    
    // Convert to new API
    request := &CompletionRequest{
        Messages: []Message{
            {Role: "user", Content: prompt},
        },
    }
    
    response, err := p.Complete(context.Background(), request)
    if err != nil {
        return "", err
    }
    
    return response.Content, nil
}
```

This comprehensive API design guide provides the foundation for building consistent, maintainable, and user-friendly APIs throughout the Go-LLMs project, ensuring long-term success and developer satisfaction.