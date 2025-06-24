# Implementing Providers

> **[Documentation Home](../README.md) / [Providers](README.md) / Implementing Providers**

## Overview

This guide walks through implementing a custom LLM provider for go-llms. Whether you're adding support for a new LLM service or creating a specialized provider, this document covers all the steps.

## Provider Interface Requirements

Your provider must implement the `domain.Provider` interface:

```go
type Provider interface {
    // Basic generation
    Generate(ctx context.Context, prompt string, options ...Option) (Response, error)
    
    // Message-based generation
    GenerateMessage(ctx context.Context, messages []Message, options ...Option) (Response, error)
    
    // Schema-constrained generation
    GenerateWithSchema(ctx context.Context, prompt string, schema *schema.Schema, options ...Option) (any, error)
    
    // Streaming generation
    Stream(ctx context.Context, prompt string, options ...Option) (<-chan StreamResponse, error)
    StreamMessage(ctx context.Context, messages []Message, options ...Option) (<-chan StreamResponse, error)
}
```

## Step-by-Step Implementation

### 1. Create Provider Structure

```go
package provider

import (
    "context"
    "net/http"
    "time"
    
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/schema/domain"
)

type MyProvider struct {
    apiKey     string
    model      string
    baseURL    string
    httpClient *http.Client
    options    *providerOptions
}

type providerOptions struct {
    timeout      time.Duration
    maxRetries   int
    temperature  float64
    maxTokens    int
    // Provider-specific options
    customParam  string
}
```

### 2. Implement Constructor

```go
func NewMyProvider(apiKey, model string, opts ...domain.Option) *MyProvider {
    p := &MyProvider{
        apiKey:  apiKey,
        model:   model,
        baseURL: "https://api.myprovider.com/v1",
        options: &providerOptions{
            timeout:     30 * time.Second,
            maxRetries:  3,
            temperature: 0.7,
            maxTokens:   2000,
        },
    }
    
    // Apply common options
    commonOpts := &domain.CommonOptions{}
    for _, opt := range opts {
        opt(commonOpts)
    }
    
    // Configure HTTP client
    if commonOpts.HTTPClient != nil {
        p.httpClient = commonOpts.HTTPClient
    } else {
        p.httpClient = &http.Client{
            Timeout: p.options.timeout,
        }
    }
    
    // Apply other common options
    if commonOpts.BaseURL != "" {
        p.baseURL = commonOpts.BaseURL
    }
    if commonOpts.MaxTokens > 0 {
        p.options.maxTokens = commonOpts.MaxTokens
    }
    if commonOpts.Temperature != nil {
        p.options.temperature = *commonOpts.Temperature
    }
    
    return p
}
```

### 3. Implement Generate Method

```go
func (p *MyProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (domain.Response, error) {
    // Convert to message format
    messages := []domain.Message{
        {
            Role: domain.RoleUser,
            Content: []domain.ContentPart{
                {Type: domain.ContentTypeText, Text: prompt},
            },
        },
    }
    
    return p.GenerateMessage(ctx, messages, options...)
}
```

### 4. Implement GenerateMessage Method

```go
func (p *MyProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
    // Apply runtime options
    opts := p.applyOptions(options...)
    
    // Build API request
    apiReq := p.buildAPIRequest(messages, opts)
    
    // Execute with retry
    var lastErr error
    for attempt := 0; attempt <= opts.maxRetries; attempt++ {
        if attempt > 0 {
            delay := p.calculateBackoff(attempt)
            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return domain.Response{}, ctx.Err()
            }
        }
        
        resp, err := p.doRequest(ctx, apiReq)
        if err == nil {
            return p.parseResponse(resp)
        }
        
        lastErr = err
        if !p.isRetryable(err) {
            break
        }
    }
    
    return domain.Response{}, p.wrapError(lastErr)
}

// API request structure (provider-specific)
type apiRequest struct {
    Model       string      `json:"model"`
    Messages    []apiMessage `json:"messages"`
    Temperature float64     `json:"temperature"`
    MaxTokens   int         `json:"max_tokens"`
    Stream      bool        `json:"stream"`
}

type apiMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

func (p *MyProvider) buildAPIRequest(messages []domain.Message, opts *providerOptions) *apiRequest {
    // Convert messages to API format
    apiMessages := make([]apiMessage, len(messages))
    for i, msg := range messages {
        apiMessages[i] = apiMessage{
            Role:    p.convertRole(msg.Role),
            Content: p.extractTextContent(msg.Content),
        }
    }
    
    return &apiRequest{
        Model:       p.model,
        Messages:    apiMessages,
        Temperature: opts.temperature,
        MaxTokens:   opts.maxTokens,
        Stream:      false,
    }
}
```

### 5. Implement HTTP Request Handling

```go
func (p *MyProvider) doRequest(ctx context.Context, apiReq *apiRequest) (*http.Response, error) {
    // Marshal request
    body, err := json.Marshal(apiReq)
    if err != nil {
        return nil, err
    }
    
    // Create HTTP request
    req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    
    // Set headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+p.apiKey)
    
    // Execute request
    resp, err := p.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    
    // Check status code
    if resp.StatusCode != http.StatusOK {
        defer resp.Body.Close()
        return nil, p.parseErrorResponse(resp)
    }
    
    return resp, nil
}
```

### 6. Implement Response Parsing

```go
// API response structure (provider-specific)
type apiResponse struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Created int64  `json:"created"`
    Choices []struct {
        Message struct {
            Role    string `json:"role"`
            Content string `json:"content"`
        } `json:"message"`
        FinishReason string `json:"finish_reason"`
    } `json:"choices"`
    Usage struct {
        PromptTokens     int `json:"prompt_tokens"`
        CompletionTokens int `json:"completion_tokens"`
        TotalTokens      int `json:"total_tokens"`
    } `json:"usage"`
}

func (p *MyProvider) parseResponse(httpResp *http.Response) (domain.Response, error) {
    defer httpResp.Body.Close()
    
    var apiResp apiResponse
    if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
        return domain.Response{}, fmt.Errorf("failed to decode response: %w", err)
    }
    
    if len(apiResp.Choices) == 0 {
        return domain.Response{}, fmt.Errorf("no choices in response")
    }
    
    return domain.Response{
        Content: apiResp.Choices[0].Message.Content,
        Usage: &domain.Usage{
            PromptTokens:     apiResp.Usage.PromptTokens,
            CompletionTokens: apiResp.Usage.CompletionTokens,
            TotalTokens:      apiResp.Usage.TotalTokens,
        },
        Metadata: map[string]interface{}{
            "id":            apiResp.ID,
            "created":       apiResp.Created,
            "finish_reason": apiResp.Choices[0].FinishReason,
        },
    }, nil
}
```

### 7. Implement Streaming

```go
func (p *MyProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (<-chan domain.StreamResponse, error) {
    messages := []domain.Message{
        {
            Role: domain.RoleUser,
            Content: []domain.ContentPart{
                {Type: domain.ContentTypeText, Text: prompt},
            },
        },
    }
    
    return p.StreamMessage(ctx, messages, options...)
}

func (p *MyProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (<-chan domain.StreamResponse, error) {
    opts := p.applyOptions(options...)
    
    // Build streaming request
    apiReq := p.buildAPIRequest(messages, opts)
    apiReq.Stream = true
    
    // Create response channel
    responseChan := make(chan domain.StreamResponse, 100)
    
    go func() {
        defer close(responseChan)
        
        resp, err := p.doStreamRequest(ctx, apiReq)
        if err != nil {
            responseChan <- domain.StreamResponse{
                Error: err,
                Done:  true,
            }
            return
        }
        defer resp.Body.Close()
        
        // Parse SSE stream
        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            line := scanner.Text()
            
            // Parse SSE data
            if strings.HasPrefix(line, "data: ") {
                data := strings.TrimPrefix(line, "data: ")
                if data == "[DONE]" {
                    responseChan <- domain.StreamResponse{Done: true}
                    return
                }
                
                chunk, err := p.parseStreamChunk(data)
                if err != nil {
                    responseChan <- domain.StreamResponse{
                        Error: err,
                        Done:  true,
                    }
                    return
                }
                
                responseChan <- chunk
            }
        }
    }()
    
    return responseChan, nil
}
```

### 8. Implement Schema Validation

```go
func (p *MyProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schema.Schema, options ...domain.Option) (any, error) {
    // Add schema instructions to prompt
    schemaPrompt := p.buildSchemaPrompt(prompt, schema)
    
    // Generate response
    response, err := p.Generate(ctx, schemaPrompt, options...)
    if err != nil {
        return nil, err
    }
    
    // Extract and validate JSON
    jsonStr := p.extractJSON(response.Content)
    
    // Parse JSON
    var result interface{}
    if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
        return nil, fmt.Errorf("failed to parse JSON response: %w", err)
    }
    
    // Validate against schema
    validator := validation.NewValidator()
    if err := validator.ValidateData(schema, result); err != nil {
        return nil, fmt.Errorf("response doesn't match schema: %w", err)
    }
    
    return result, nil
}

func (p *MyProvider) buildSchemaPrompt(prompt string, schema *schema.Schema) string {
    schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
    
    return fmt.Sprintf(`%s

You must respond with valid JSON that matches this schema:
%s

Important: Respond ONLY with the JSON object, no additional text.`, prompt, schemaJSON)
}
```

### 9. Implement Error Handling

```go
func (p *MyProvider) parseErrorResponse(resp *http.Response) error {
    body, _ := io.ReadAll(resp.Body)
    
    // Try to parse provider-specific error format
    var apiError struct {
        Error struct {
            Message string `json:"message"`
            Type    string `json:"type"`
            Code    string `json:"code"`
        } `json:"error"`
    }
    
    if err := json.Unmarshal(body, &apiError); err == nil {
        return p.categorizeError(resp.StatusCode, apiError.Error.Type, apiError.Error.Message)
    }
    
    // Fallback to generic error
    return &domain.ProviderError{
        Provider: "MyProvider",
        Type:     p.errorTypeFromStatus(resp.StatusCode),
        Message:  fmt.Sprintf("API error: %s", body),
        Details: map[string]interface{}{
            "status_code": resp.StatusCode,
            "body":        string(body),
        },
    }
}

func (p *MyProvider) categorizeError(statusCode int, errorType, message string) error {
    var errType domain.ErrorType
    
    switch statusCode {
    case 401:
        errType = domain.ErrorTypeAuthentication
    case 429:
        errType = domain.ErrorTypeRateLimit
    case 400:
        if strings.Contains(message, "context") || strings.Contains(message, "token") {
            errType = domain.ErrorTypeContextLength
        } else {
            errType = domain.ErrorTypeInvalidRequest
        }
    case 500, 502, 503:
        errType = domain.ErrorTypeProvider
    default:
        errType = domain.ErrorTypeUnknown
    }
    
    return &domain.ProviderError{
        Provider: "MyProvider",
        Type:     errType,
        Message:  message,
        Details: map[string]interface{}{
            "status_code": statusCode,
            "error_type":  errorType,
        },
    }
}
```

### 10. Implement Metadata Provider (Optional)

```go
func (p *MyProvider) GetMetadata() domain.ProviderMetadata {
    return domain.ProviderMetadata{
        Name:        "MyProvider",
        Description: "Custom LLM Provider implementation",
        Capabilities: []domain.Capability{
            domain.CapabilityTextGeneration,
            domain.CapabilityStreaming,
            domain.CapabilityFunctionCalling,
        },
        Models: []domain.ModelInfo{
            {
                ID:          "model-v1",
                Name:        "Model V1",
                Description: "General purpose model",
                Context:     8192,
                Input:       0.01,  // Cost per 1K tokens
                Output:      0.02,  // Cost per 1K tokens
            },
        },
        Constraints: domain.Constraints{
            MaxTokens:      4096,
            MaxContextSize: 8192,
            RateLimit: &domain.RateLimit{
                RequestsPerMinute: 60,
                TokensPerMinute:   90000,
            },
        },
    }
}
```

## Testing Your Provider

### Unit Tests

```go
func TestMyProvider_Generate(t *testing.T) {
    // Create test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/v1/chat/completions", r.URL.Path)
        
        // Return mock response
        response := apiResponse{
            Choices: []struct{
                Message struct{
                    Content string `json:"content"`
                } `json:"message"`
            }{
                {Message: struct{Content string `json:"content"`}{Content: "Test response"}},
            },
        }
        
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()
    
    // Create provider with test server
    provider := NewMyProvider("test-key", "test-model",
        domain.WithBaseURL(server.URL),
    )
    
    // Test generation
    resp, err := provider.Generate(context.Background(), "Test prompt")
    assert.NoError(t, err)
    assert.Equal(t, "Test response", resp.Content)
}
```

### Integration Tests

```go
func TestMyProvider_Integration(t *testing.T) {
    // Skip if no API key
    apiKey := os.Getenv("MYPROVIDER_API_KEY")
    if apiKey == "" {
        t.Skip("MYPROVIDER_API_KEY not set")
    }
    
    provider := NewMyProvider(apiKey, "model-v1")
    
    t.Run("Generate", func(t *testing.T) {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        resp, err := provider.Generate(ctx, "Say hello")
        require.NoError(t, err)
        assert.NotEmpty(t, resp.Content)
}
    
    t.Run("Stream", func(t *testing.T) {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        stream, err := provider.Stream(ctx, "Count to 5")
        require.NoError(t, err)
        
        var content strings.Builder
        for chunk := range stream {
            if chunk.Error != nil {
                t.Fatal(chunk.Error)
            }
            content.WriteString(chunk.Content)
        }
        
        assert.Contains(t, content.String(), "5")
}
}
```

## Provider Registration

### Static Registration

Add your provider to the package initialization:

```go
// In pkg/llm/provider/init.go
func init() {
    // Register provider factory
    registry.RegisterProvider("myprovider", func(config map[string]interface{}) (domain.Provider, error) {
        apiKey, _ := config["api_key"].(string)
        model, _ := config["model"].(string)
        return NewMyProvider(apiKey, model), nil
}
}
```

### Dynamic Registration

```go
// In your application
registry := provider.GetRegistry()
registry.RegisterProvider("myprovider", func(config map[string]interface{}) (domain.Provider, error) {
    // Create provider from config
    return NewMyProvider(config["api_key"].(string), config["model"].(string)), nil
}
```

## Best Practices

### 1. Consistent Error Handling
Always wrap errors with context and categorize them properly:
```go
if err != nil {
    return nil, fmt.Errorf("myprovider: failed to generate: %w", err)
}
```

### 2. Context Handling
Respect context cancellation throughout:
```go
select {
case <-ctx.Done():
    return nil, ctx.Err()
case result := <-resultChan:
    return result, nil
}
```

### 3. Resource Management
Always clean up resources:
```go
defer func() {
    if resp != nil && resp.Body != nil {
        resp.Body.Close()
    }
}()
```

### 4. Configuration Validation
Validate configuration early:
```go
func NewMyProvider(apiKey, model string, opts ...Option) (*MyProvider, error) {
    if apiKey == "" {
        return nil, fmt.Errorf("API key is required")
    }
    if model == "" {
        return nil, fmt.Errorf("model is required")
    }
    // ... rest of initialization
}
```

## Checklist

Before considering your provider complete:

- [ ] All interface methods implemented
- [ ] Error handling with proper categorization
- [ ] Context cancellation support
- [ ] Retry logic for transient failures
- [ ] Rate limit handling
- [ ] Streaming support (if applicable)
- [ ] Schema validation support
- [ ] Unit tests with mocked responses
- [ ] Integration tests (with skip if no API key)
- [ ] Documentation with examples
- [ ] Provider metadata implementation
- [ ] Registration in provider registry

## Next Steps

- Review [Provider Registry](provider-registry.md) for registration details
- Implement [Provider Metadata](metadata.md) for capability discovery
- Add comprehensive tests following [Testing Guide](../development/testing.md)
- Submit your provider following [Contributing Guide](../development/contributing.md)