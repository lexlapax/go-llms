# Custom Providers: Creating Custom LLM Providers

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Advanced Topics](/docs/user-guide/advanced/) / Custom Providers**

Learn how to create custom LLM providers for Go-LLMs, including implementing the provider interface, handling authentication, supporting streaming, and integrating with the provider registry.

## Provider Architecture Overview

Go-LLMs providers must implement the core `Provider` interface and optionally support additional capabilities:

```go
// Core provider interface
type Provider interface {
    // Basic information
    Name() string
    
    // Model operations
    ListModels(ctx context.Context) ([]Model, error)
    
    // Completion operations
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    CompleteStream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error)
    
    // Embedding operations (optional)
    Embed(ctx context.Context, req *EmbedRequest) (*EmbedResponse, error)
}

// Extended capabilities
type ToolProvider interface {
    Provider
    SupportsTools() bool
    CompleteWithTools(ctx context.Context, req *ToolCompletionRequest) (*ToolCompletionResponse, error)
}

type VisionProvider interface {
    Provider
    SupportsVision() bool
    CompleteWithImages(ctx context.Context, req *VisionRequest) (*VisionResponse, error)
}
```

---

## Basic Provider Implementation

### Step 1: Provider Structure

```go
package customprovider

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// CustomProvider implements a custom LLM provider
type CustomProvider struct {
    apiKey     string
    baseURL    string
    httpClient *http.Client
    options    CustomOptions
}

// CustomOptions contains provider-specific configuration
type CustomOptions struct {
    APIKey          string
    BaseURL         string
    OrganizationID  string
    DefaultModel    string
    Timeout         time.Duration
    MaxRetries      int
    RateLimit       int // requests per minute
    CustomHeaders   map[string]string
}

// NewCustomProvider creates a new custom provider instance
func NewCustomProvider(opts CustomOptions) (*CustomProvider, error) {
    if opts.APIKey == "" {
        return nil, fmt.Errorf("API key is required")
    }
    
    if opts.BaseURL == "" {
        opts.BaseURL = "https://api.custom-llm.com/v1"
    }
    
    if opts.Timeout == 0 {
        opts.Timeout = 60 * time.Second
    }
    
    httpClient := &http.Client{
        Timeout: opts.Timeout,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }
    
    return &CustomProvider{
        apiKey:     opts.APIKey,
        baseURL:    opts.BaseURL,
        httpClient: httpClient,
        options:    opts,
    }, nil
}

// Name returns the provider name
func (p *CustomProvider) Name() string {
    return "custom"
}
```

### Step 2: Model Management

```go
// Model represents an available model
type Model struct {
    ID           string    `json:"id"`
    Name         string    `json:"name"`
    Description  string    `json:"description"`
    MaxTokens    int       `json:"max_tokens"`
    InputCost    float64   `json:"input_cost"`
    OutputCost   float64   `json:"output_cost"`
    Capabilities []string  `json:"capabilities"`
    Created      time.Time `json:"created"`
}

// ListModels returns available models
func (p *CustomProvider) ListModels(ctx context.Context) ([]provider.Model, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", 
        fmt.Sprintf("%s/models", p.baseURL), nil)
    if err != nil {
        return nil, err
    }
    
    p.setHeaders(req)
    
    resp, err := p.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to list models: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, p.handleErrorResponse(resp)
    }
    
    var modelsResp struct {
        Models []Model `json:"models"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
        return nil, fmt.Errorf("failed to decode models: %w", err)
    }
    
    // Convert to provider.Model
    models := make([]provider.Model, len(modelsResp.Models))
    for i, m := range modelsResp.Models {
        models[i] = provider.Model{
            ID:          m.ID,
            Name:        m.Name,
            Description: m.Description,
            Context:     m.MaxTokens,
            Input:       m.InputCost,
            Output:      m.OutputCost,
        }
    }
    
    return models, nil
}

// GetModel returns information about a specific model
func (p *CustomProvider) GetModel(ctx context.Context, modelID string) (*Model, error) {
    models, err := p.ListModels(ctx)
    if err != nil {
        return nil, err
    }
    
    for _, model := range models {
        if model.ID == modelID {
            return &Model{
                ID:          model.ID,
                Name:        model.Name,
                Description: model.Description,
                MaxTokens:   model.Context,
            }, nil
        }
    }
    
    return nil, fmt.Errorf("model %s not found", modelID)
}
```

### Step 3: Completion Implementation

```go
// Complete performs a completion request
func (p *CustomProvider) Complete(ctx context.Context, req *provider.CompletionRequest) (*provider.CompletionResponse, error) {
    // Validate request
    if err := p.validateRequest(req); err != nil {
        return nil, err
    }
    
    // Convert to API format
    apiReq := p.convertToAPIRequest(req)
    
    // Make API call with retry logic
    var resp *provider.CompletionResponse
    var lastErr error
    
    for attempt := 0; attempt <= p.options.MaxRetries; attempt++ {
        resp, lastErr = p.doComplete(ctx, apiReq)
        if lastErr == nil {
            return resp, nil
        }
        
        // Check if error is retryable
        if !p.isRetryableError(lastErr) {
            return nil, lastErr
        }
        
        // Calculate backoff
        backoff := p.calculateBackoff(attempt)
        select {
        case <-time.After(backoff):
            // Continue with retry
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (p *CustomProvider) doComplete(ctx context.Context, apiReq *customAPIRequest) (*provider.CompletionResponse, error) {
    jsonData, err := json.Marshal(apiReq)
    if err != nil {
        return nil, err
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST",
        fmt.Sprintf("%s/completions", p.baseURL),
        bytes.NewReader(jsonData))
    if err != nil {
        return nil, err
    }
    
    p.setHeaders(req)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := p.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, p.handleErrorResponse(resp)
    }
    
    var apiResp customAPIResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return p.convertFromAPIResponse(&apiResp), nil
}

// Request/Response conversion
type customAPIRequest struct {
    Model       string                 `json:"model"`
    Messages    []customMessage        `json:"messages"`
    Temperature float64                `json:"temperature,omitempty"`
    MaxTokens   int                    `json:"max_tokens,omitempty"`
    Stream      bool                   `json:"stream"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type customMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type customAPIResponse struct {
    ID      string            `json:"id"`
    Choices []customChoice    `json:"choices"`
    Usage   customUsage       `json:"usage"`
    Model   string            `json:"model"`
}

type customChoice struct {
    Message      customMessage `json:"message"`
    FinishReason string        `json:"finish_reason"`
}

type customUsage struct {
    InputTokens  int `json:"input_tokens"`
    OutputTokens int `json:"output_tokens"`
    TotalTokens  int `json:"total_tokens"`
}

func (p *CustomProvider) convertToAPIRequest(req *provider.CompletionRequest) *customAPIRequest {
    apiReq := &customAPIRequest{
        Model:       req.Model,
        Temperature: req.Temperature,
        MaxTokens:   req.MaxTokens,
        Stream:      false,
        Messages:    make([]customMessage, len(req.Messages)),
    }
    
    for i, msg := range req.Messages {
        apiReq.Messages[i] = customMessage{
            Role:    msg.Role,
            Content: msg.Content,
        }
    }
    
    return apiReq
}

func (p *CustomProvider) convertFromAPIResponse(apiResp *customAPIResponse) *provider.CompletionResponse {
    if len(apiResp.Choices) == 0 {
        return &provider.CompletionResponse{}
    }
    
    return &provider.CompletionResponse{
        ID:      apiResp.ID,
        Model:   apiResp.Model,
        Content: apiResp.Choices[0].Message.Content,
        Usage: &provider.Usage{
            PromptTokens:     apiResp.Usage.InputTokens,
            CompletionTokens: apiResp.Usage.OutputTokens,
            TotalTokens:      apiResp.Usage.TotalTokens,
        },
        FinishReason: apiResp.Choices[0].FinishReason,
    }
}
```

### Step 4: Streaming Support

```go
// CompleteStream performs a streaming completion request
func (p *CustomProvider) CompleteStream(ctx context.Context, req *provider.CompletionRequest) (<-chan provider.StreamChunk, error) {
    // Validate request
    if err := p.validateRequest(req); err != nil {
        return nil, err
    }
    
    // Create stream channel
    stream := make(chan provider.StreamChunk, 100)
    
    // Start streaming in goroutine
    go func() {
        defer close(stream)
        
        if err := p.doStreamComplete(ctx, req, stream); err != nil {
            stream <- provider.StreamChunk{
                Error: err,
            }
        }
    }()
    
    return stream, nil
}

func (p *CustomProvider) doStreamComplete(ctx context.Context, req *provider.CompletionRequest, stream chan<- provider.StreamChunk) error {
    apiReq := p.convertToAPIRequest(req)
    apiReq.Stream = true
    
    jsonData, err := json.Marshal(apiReq)
    if err != nil {
        return err
    }
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST",
        fmt.Sprintf("%s/completions", p.baseURL),
        bytes.NewReader(jsonData))
    if err != nil {
        return err
    }
    
    p.setHeaders(httpReq)
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Accept", "text/event-stream")
    
    resp, err := p.httpClient.Do(httpReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return p.handleErrorResponse(resp)
    }
    
    // Parse SSE stream
    scanner := bufio.NewScanner(resp.Body)
    var currentEvent strings.Builder
    
    for scanner.Scan() {
        line := scanner.Text()
        
        if line == "" {
            // End of event
            if currentEvent.Len() > 0 {
                if err := p.processStreamEvent(currentEvent.String(), stream); err != nil {
                    return err
                }
                currentEvent.Reset()
            }
        } else if strings.HasPrefix(line, "data: ") {
            currentEvent.WriteString(strings.TrimPrefix(line, "data: "))
        }
    }
    
    return scanner.Err()
}

func (p *CustomProvider) processStreamEvent(data string, stream chan<- provider.StreamChunk) error {
    if data == "[DONE]" {
        return nil
    }
    
    var event customStreamEvent
    if err := json.Unmarshal([]byte(data), &event); err != nil {
        return fmt.Errorf("failed to parse stream event: %w", err)
    }
    
    if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
        stream <- provider.StreamChunk{
            Content: event.Choices[0].Delta.Content,
            Model:   event.Model,
        }
    }
    
    return nil
}

type customStreamEvent struct {
    ID      string               `json:"id"`
    Model   string               `json:"model"`
    Choices []customStreamChoice `json:"choices"`
}

type customStreamChoice struct {
    Delta        customStreamDelta `json:"delta"`
    FinishReason *string           `json:"finish_reason"`
}

type customStreamDelta struct {
    Content string `json:"content"`
}
```

---

## Advanced Features

### Tool/Function Calling Support

```go
// Implement ToolProvider interface
func (p *CustomProvider) SupportsTools() bool {
    return true
}

func (p *CustomProvider) CompleteWithTools(ctx context.Context, req *provider.ToolCompletionRequest) (*provider.ToolCompletionResponse, error) {
    // Convert tools to API format
    apiTools := make([]customTool, len(req.Tools))
    for i, tool := range req.Tools {
        apiTools[i] = customTool{
            Type: "function",
            Function: customFunction{
                Name:        tool.Name,
                Description: tool.Description,
                Parameters:  tool.Parameters,
            },
        }
    }
    
    // Create API request with tools
    apiReq := &customToolRequest{
        Model:       req.Model,
        Messages:    p.convertMessages(req.Messages),
        Tools:       apiTools,
        ToolChoice:  req.ToolChoice,
        Temperature: req.Temperature,
        MaxTokens:   req.MaxTokens,
    }
    
    // Make API call
    jsonData, err := json.Marshal(apiReq)
    if err != nil {
        return nil, err
    }
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST",
        fmt.Sprintf("%s/completions", p.baseURL),
        bytes.NewReader(jsonData))
    if err != nil {
        return nil, err
    }
    
    p.setHeaders(httpReq)
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := p.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, p.handleErrorResponse(resp)
    }
    
    var apiResp customToolResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, err
    }
    
    return p.convertToolResponse(&apiResp), nil
}

type customTool struct {
    Type     string         `json:"type"`
    Function customFunction `json:"function"`
}

type customFunction struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"`
}

type customToolRequest struct {
    Model       string          `json:"model"`
    Messages    []customMessage `json:"messages"`
    Tools       []customTool    `json:"tools"`
    ToolChoice  interface{}     `json:"tool_choice,omitempty"`
    Temperature float64         `json:"temperature,omitempty"`
    MaxTokens   int             `json:"max_tokens,omitempty"`
}

type customToolResponse struct {
    ID      string             `json:"id"`
    Model   string             `json:"model"`
    Choices []customToolChoice `json:"choices"`
    Usage   customUsage        `json:"usage"`
}

type customToolChoice struct {
    Message      customToolMessage `json:"message"`
    FinishReason string            `json:"finish_reason"`
}

type customToolMessage struct {
    Role         string           `json:"role"`
    Content      string           `json:"content,omitempty"`
    ToolCalls    []customToolCall `json:"tool_calls,omitempty"`
}

type customToolCall struct {
    ID       string                 `json:"id"`
    Type     string                 `json:"type"`
    Function customFunctionCall     `json:"function"`
}

type customFunctionCall struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}
```

### Vision/Multimodal Support

```go
// Implement VisionProvider interface
func (p *CustomProvider) SupportsVision() bool {
    return true
}

func (p *CustomProvider) CompleteWithImages(ctx context.Context, req *provider.VisionRequest) (*provider.VisionResponse, error) {
    // Convert messages with images
    apiMessages := make([]customVisionMessage, len(req.Messages))
    
    for i, msg := range req.Messages {
        content := []customVisionContent{
            {
                Type: "text",
                Text: msg.Content,
            },
        }
        
        // Add images
        for _, img := range msg.Images {
            if img.URL != "" {
                content = append(content, customVisionContent{
                    Type: "image_url",
                    ImageURL: &customImageURL{
                        URL: img.URL,
                    },
                })
            } else if img.Base64 != "" {
                content = append(content, customVisionContent{
                    Type: "image_url",
                    ImageURL: &customImageURL{
                        URL: fmt.Sprintf("data:image/jpeg;base64,%s", img.Base64),
                    },
                })
            }
        }
        
        apiMessages[i] = customVisionMessage{
            Role:    msg.Role,
            Content: content,
        }
    }
    
    // Create vision request
    apiReq := &customVisionRequest{
        Model:       req.Model,
        Messages:    apiMessages,
        MaxTokens:   req.MaxTokens,
        Temperature: req.Temperature,
    }
    
    // Make API call and process response
    // ... (similar to regular completion)
    
    return nil, nil
}

type customVisionMessage struct {
    Role    string                 `json:"role"`
    Content []customVisionContent  `json:"content"`
}

type customVisionContent struct {
    Type     string            `json:"type"`
    Text     string            `json:"text,omitempty"`
    ImageURL *customImageURL   `json:"image_url,omitempty"`
}

type customImageURL struct {
    URL string `json:"url"`
}
```

### Embeddings Support

```go
// Embed creates embeddings for input text
func (p *CustomProvider) Embed(ctx context.Context, req *provider.EmbedRequest) (*provider.EmbedResponse, error) {
    apiReq := &customEmbedRequest{
        Model:  req.Model,
        Input:  req.Input,
        Format: "float",
    }
    
    jsonData, err := json.Marshal(apiReq)
    if err != nil {
        return nil, err
    }
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST",
        fmt.Sprintf("%s/embeddings", p.baseURL),
        bytes.NewReader(jsonData))
    if err != nil {
        return nil, err
    }
    
    p.setHeaders(httpReq)
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := p.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, p.handleErrorResponse(resp)
    }
    
    var apiResp customEmbedResponse
    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, err
    }
    
    // Convert to provider format
    embeddings := make([][]float64, len(apiResp.Data))
    for i, data := range apiResp.Data {
        embeddings[i] = data.Embedding
    }
    
    return &provider.EmbedResponse{
        Embeddings: embeddings,
        Model:      apiResp.Model,
        Usage: &provider.Usage{
            PromptTokens: apiResp.Usage.PromptTokens,
            TotalTokens:  apiResp.Usage.TotalTokens,
        },
    }, nil
}

type customEmbedRequest struct {
    Model  string   `json:"model"`
    Input  []string `json:"input"`
    Format string   `json:"encoding_format,omitempty"`
}

type customEmbedResponse struct {
    Data  []customEmbedData `json:"data"`
    Model string            `json:"model"`
    Usage customEmbedUsage  `json:"usage"`
}

type customEmbedData struct {
    Embedding []float64 `json:"embedding"`
    Index     int       `json:"index"`
}

type customEmbedUsage struct {
    PromptTokens int `json:"prompt_tokens"`
    TotalTokens  int `json:"total_tokens"`
}
```

---

## Error Handling and Resilience

### Custom Error Types

```go
// Define provider-specific errors
type CustomProviderError struct {
    StatusCode int
    ErrorCode  string
    Message    string
    Details    map[string]interface{}
}

func (e *CustomProviderError) Error() string {
    return fmt.Sprintf("[%s] %s (status: %d)", e.ErrorCode, e.Message, e.StatusCode)
}

// Error categories
var (
    ErrInvalidAPIKey     = &CustomProviderError{StatusCode: 401, ErrorCode: "invalid_api_key"}
    ErrRateLimitExceeded = &CustomProviderError{StatusCode: 429, ErrorCode: "rate_limit_exceeded"}
    ErrModelNotFound     = &CustomProviderError{StatusCode: 404, ErrorCode: "model_not_found"}
    ErrInvalidRequest    = &CustomProviderError{StatusCode: 400, ErrorCode: "invalid_request"}
    ErrServerError       = &CustomProviderError{StatusCode: 500, ErrorCode: "server_error"}
)

func (p *CustomProvider) handleErrorResponse(resp *http.Response) error {
    var errorResp struct {
        Error struct {
            Code    string                 `json:"code"`
            Message string                 `json:"message"`
            Details map[string]interface{} `json:"details"`
        } `json:"error"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
        return &CustomProviderError{
            StatusCode: resp.StatusCode,
            ErrorCode:  "unknown_error",
            Message:    fmt.Sprintf("HTTP %d", resp.StatusCode),
        }
    }
    
    return &CustomProviderError{
        StatusCode: resp.StatusCode,
        ErrorCode:  errorResp.Error.Code,
        Message:    errorResp.Error.Message,
        Details:    errorResp.Error.Details,
    }
}

func (p *CustomProvider) isRetryableError(err error) bool {
    var customErr *CustomProviderError
    if errors.As(err, &customErr) {
        switch customErr.StatusCode {
        case 429, 502, 503, 504:
            return true
        case 500:
            // Retry server errors with specific codes
            if customErr.ErrorCode == "temporary_failure" {
                return true
            }
        }
    }
    
    // Network errors are retryable
    var netErr net.Error
    if errors.As(err, &netErr) {
        return netErr.Temporary()
    }
    
    return false
}
```

### Rate Limiting

```go
// Rate limiter implementation
type RateLimiter struct {
    limiter  *rate.Limiter
    mu       sync.Mutex
    counters map[string]*rate.Limiter
}

func NewRateLimiter(rps int) *RateLimiter {
    return &RateLimiter{
        limiter:  rate.NewLimiter(rate.Limit(rps), rps),
        counters: make(map[string]*rate.Limiter),
    }
}

func (rl *RateLimiter) Wait(ctx context.Context, key string) error {
    rl.mu.Lock()
    limiter, exists := rl.counters[key]
    if !exists {
        limiter = rate.NewLimiter(rate.Limit(10), 10) // Per-key limit
        rl.counters[key] = limiter
    }
    rl.mu.Unlock()
    
    // Wait for both global and per-key limits
    if err := rl.limiter.Wait(ctx); err != nil {
        return err
    }
    
    return limiter.Wait(ctx)
}

// Add to provider
func (p *CustomProvider) applyRateLimit(ctx context.Context, userID string) error {
    if p.rateLimiter != nil {
        return p.rateLimiter.Wait(ctx, userID)
    }
    return nil
}
```

### Retry Logic

```go
// Exponential backoff with jitter
func (p *CustomProvider) calculateBackoff(attempt int) time.Duration {
    base := float64(100 * time.Millisecond)
    max := float64(30 * time.Second)
    
    // Exponential backoff
    backoff := base * math.Pow(2, float64(attempt))
    
    // Add jitter
    jitter := rand.Float64() * 0.3 * backoff
    backoff = backoff + jitter
    
    // Cap at maximum
    if backoff > max {
        backoff = max
    }
    
    return time.Duration(backoff)
}

// Retry with circuit breaker
type CircuitBreaker struct {
    failures     int
    lastFailTime time.Time
    state        CircuitState
    threshold    int
    timeout      time.Duration
    mu           sync.RWMutex
}

type CircuitState int

const (
    CircuitClosed CircuitState = iota
    CircuitOpen
    CircuitHalfOpen
)

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.RLock()
    state := cb.state
    cb.mu.RUnlock()
    
    if state == CircuitOpen {
        cb.mu.RLock()
        if time.Since(cb.lastFailTime) > cb.timeout {
            cb.mu.RUnlock()
            cb.mu.Lock()
            cb.state = CircuitHalfOpen
            cb.mu.Unlock()
        } else {
            cb.mu.RUnlock()
            return errors.New("circuit breaker is open")
        }
    }
    
    err := fn()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()
        
        if cb.failures >= cb.threshold {
            cb.state = CircuitOpen
        }
        
        return err
    }
    
    // Success - reset
    cb.failures = 0
    cb.state = CircuitClosed
    
    return nil
}
```

---

## Provider Registration

### Static Registration

```go
// Register with the provider registry
func init() {
    provider.Register("custom", func(config map[string]interface{}) (provider.Provider, error) {
        opts := CustomOptions{}
        
        // Parse configuration
        if apiKey, ok := config["api_key"].(string); ok {
            opts.APIKey = apiKey
        } else if apiKey := os.Getenv("CUSTOM_API_KEY"); apiKey != "" {
            opts.APIKey = apiKey
        } else {
            return nil, errors.New("API key not provided")
        }
        
        if baseURL, ok := config["base_url"].(string); ok {
            opts.BaseURL = baseURL
        }
        
        if timeout, ok := config["timeout"].(float64); ok {
            opts.Timeout = time.Duration(timeout) * time.Second
        }
        
        return NewCustomProvider(opts)
    })
}
```

### Dynamic Registration

```go
// Provider metadata for discovery
type ProviderMetadata struct {
    Name         string
    DisplayName  string
    Description  string
    Capabilities []string
    Models       []ModelInfo
    ConfigSchema map[string]interface{}
}

func (p *CustomProvider) GetMetadata() ProviderMetadata {
    return ProviderMetadata{
        Name:        "custom",
        DisplayName: "Custom LLM Provider",
        Description: "Custom implementation of LLM provider",
        Capabilities: []string{
            "completion",
            "streaming",
            "tools",
            "vision",
            "embeddings",
        },
        Models: []ModelInfo{
            {
                ID:          "custom-large",
                Name:        "Custom Large Model",
                MaxTokens:   100000,
                InputCost:   0.01,
                OutputCost:  0.02,
            },
            {
                ID:          "custom-small",
                Name:        "Custom Small Model",
                MaxTokens:   50000,
                InputCost:   0.001,
                OutputCost:  0.002,
            },
        },
        ConfigSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "api_key": map[string]interface{}{
                    "type":        "string",
                    "description": "API authentication key",
                    "required":    true,
                },
                "base_url": map[string]interface{}{
                    "type":        "string",
                    "description": "API base URL",
                    "default":     "https://api.custom-llm.com/v1",
                },
                "timeout": map[string]interface{}{
                    "type":        "integer",
                    "description": "Request timeout in seconds",
                    "default":     60,
                },
            },
        },
    }
}

// Discovery interface
type DiscoverableProvider interface {
    provider.Provider
    GetMetadata() ProviderMetadata
    ValidateConfig(config map[string]interface{}) error
    HealthCheck(ctx context.Context) error
}

func (p *CustomProvider) ValidateConfig(config map[string]interface{}) error {
    if _, ok := config["api_key"].(string); !ok {
        return errors.New("api_key is required")
    }
    
    if baseURL, ok := config["base_url"].(string); ok {
        if _, err := url.Parse(baseURL); err != nil {
            return fmt.Errorf("invalid base_url: %w", err)
        }
    }
    
    return nil
}

func (p *CustomProvider) HealthCheck(ctx context.Context) error {
    // Try to list models as health check
    _, err := p.ListModels(ctx)
    return err
}
```

---

## Testing Custom Providers

### Unit Tests

```go
package customprovider_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCustomProvider_Complete(t *testing.T) {
    // Mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/v1/completions", r.URL.Path)
        assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{
            "id": "test-123",
            "model": "custom-large",
            "choices": [{
                "message": {"role": "assistant", "content": "Hello!"},
                "finish_reason": "stop"
            }],
            "usage": {"input_tokens": 10, "output_tokens": 5, "total_tokens": 15}
        }`))
    }))
    defer server.Close()
    
    // Create provider
    provider, err := NewCustomProvider(CustomOptions{
        APIKey:  "test-key",
        BaseURL: server.URL + "/v1",
    })
    require.NoError(t, err)
    
    // Test completion
    resp, err := provider.Complete(context.Background(), &provider.CompletionRequest{
        Model: "custom-large",
        Messages: []provider.Message{
            {Role: "user", Content: "Hello"},
        },
    })
    
    require.NoError(t, err)
    assert.Equal(t, "Hello!", resp.Content)
    assert.Equal(t, 15, resp.Usage.TotalTokens)
}

func TestCustomProvider_StreamComplete(t *testing.T) {
    // Mock SSE server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/event-stream")
        w.WriteHeader(http.StatusOK)
        
        // Send stream events
        fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"delta":{"content":"Hello"}}]}`)
        fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"delta":{"content":" world"}}]}`)
        fmt.Fprintf(w, "data: [DONE]\n\n")
    }))
    defer server.Close()
    
    provider, err := NewCustomProvider(CustomOptions{
        APIKey:  "test-key",
        BaseURL: server.URL + "/v1",
    })
    require.NoError(t, err)
    
    // Test streaming
    stream, err := provider.CompleteStream(context.Background(), &provider.CompletionRequest{
        Model: "custom-large",
        Messages: []provider.Message{
            {Role: "user", Content: "Hello"},
        },
    })
    require.NoError(t, err)
    
    // Collect chunks
    var content string
    for chunk := range stream {
        if chunk.Error != nil {
            t.Fatal(chunk.Error)
        }
        content += chunk.Content
    }
    
    assert.Equal(t, "Hello world", content)
}
```

### Integration Tests

```go
func TestCustomProvider_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Use real API
    provider, err := NewCustomProvider(CustomOptions{
        APIKey: os.Getenv("CUSTOM_API_KEY"),
    })
    require.NoError(t, err)
    
    ctx := context.Background()
    
    // Test model listing
    t.Run("ListModels", func(t *testing.T) {
        models, err := provider.ListModels(ctx)
        require.NoError(t, err)
        assert.NotEmpty(t, models)
    })
    
    // Test completion
    t.Run("Complete", func(t *testing.T) {
        resp, err := provider.Complete(ctx, &provider.CompletionRequest{
            Model: "custom-small",
            Messages: []provider.Message{
                {Role: "user", Content: "What is 2+2?"},
            },
            MaxTokens: 10,
        })
        
        require.NoError(t, err)
        assert.NotEmpty(t, resp.Content)
        assert.Contains(t, resp.Content, "4")
    })
    
    // Test error handling
    t.Run("InvalidModel", func(t *testing.T) {
        _, err := provider.Complete(ctx, &provider.CompletionRequest{
            Model: "non-existent-model",
            Messages: []provider.Message{
                {Role: "user", Content: "Test"},
            },
        })
        
        require.Error(t, err)
        var customErr *CustomProviderError
        assert.ErrorAs(t, err, &customErr)
        assert.Equal(t, 404, customErr.StatusCode)
    })
}
```

---

## Provider Configuration Guide

### Environment Variables

```bash
# Required
export CUSTOM_API_KEY="your-api-key"

# Optional
export CUSTOM_BASE_URL="https://api.custom-llm.com/v1"
export CUSTOM_TIMEOUT="60"
export CUSTOM_MAX_RETRIES="3"
export CUSTOM_RATE_LIMIT="100"
```

### Configuration File

```yaml
# config.yaml
providers:
  custom:
    api_key: ${CUSTOM_API_KEY}
    base_url: https://api.custom-llm.com/v1
    timeout: 60
    max_retries: 3
    rate_limit: 100
    default_model: custom-large
    custom_headers:
      X-Custom-Header: value
    models:
      - id: custom-large
        max_tokens: 100000
      - id: custom-small
        max_tokens: 50000
```

### Usage Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    _ "github.com/yourorg/custom-provider" // Import registers provider
)

func main() {
    // Create provider from config
    p, err := provider.NewFromConfig(map[string]interface{}{
        "provider": "custom",
        "api_key":  os.Getenv("CUSTOM_API_KEY"),
        "timeout":  30,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Use provider
    resp, err := p.Complete(context.Background(), &provider.CompletionRequest{
        Model: "custom-large",
        Messages: []provider.Message{
            {Role: "system", Content: "You are a helpful assistant."},
            {Role: "user", Content: "Hello!"},
        },
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(resp.Content)
}
```

---

## Best Practices

### 1. Error Handling
- Implement comprehensive error types
- Provide detailed error messages
- Support error recovery and retries
- Log errors appropriately

### 2. Performance
- Use connection pooling
- Implement proper timeouts
- Support request cancellation
- Add caching where appropriate

### 3. Security
- Secure API key storage
- Validate all inputs
- Implement rate limiting
- Use TLS for all connections

### 4. Observability
- Add structured logging
- Implement metrics collection
- Support distributed tracing
- Provide health check endpoints

### 5. Testing
- Write comprehensive unit tests
- Include integration tests
- Test error scenarios
- Benchmark performance

### 6. Documentation
- Document all configuration options
- Provide usage examples
- Explain error codes
- Include troubleshooting guide

---

## Next Steps

- **[Custom Tools](custom-tools.md)** - Create custom tools for agents
- **[Workflow Orchestration](workflow-orchestration.md)** - Advanced workflow patterns
- **[Production Deployment](production-deployment.md)** - Deploy custom providers
- **[Provider Comparison](/docs/user-guide/reference/provider-comparison.md)** - Compare with built-in providers
- **[API Reference](/docs/technical/api-reference/providers.md)** - Provider interface details