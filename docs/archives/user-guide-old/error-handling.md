# Error Handling Guide

Learn how to handle errors effectively in your go-llms applications.

## Overview

Building robust LLM applications requires proper error handling. This guide shows you how to handle common errors, implement retry logic, and build resilient applications.

## Common Error Types

Go-llms provides standardized error types for consistent error handling across all providers.

### Authentication Errors

```go
response, err := provider.Generate(ctx, prompt)
if err != nil {
    if domain.IsAuthenticationError(err) {
        // API key is invalid or missing
        log.Fatal("Please check your API key")
    }
}
```

### Rate Limiting

```go
if domain.IsRateLimitError(err) {
    // Too many requests
    log.Printf("Rate limited, waiting before retry...")
    time.Sleep(time.Minute)
    // Retry the request
}
```

### Context Length

```go
if domain.IsContextTooLongError(err) {
    // Input is too long for the model
    log.Printf("Input too long, truncating...")
    truncatedPrompt := truncatePrompt(prompt, maxTokens)
    response, err = provider.Generate(ctx, truncatedPrompt)
}
```

### Network Issues

```go
if domain.IsNetworkConnectivityError(err) || domain.IsTimeoutError(err) {
    // Network or timeout issues
    log.Printf("Network issue: %v", err)
    // Implement exponential backoff
}
```

## Error Classification

Use helper functions to classify errors:

```go
// Check specific error types
if domain.IsAuthenticationError(err) {
    // Handle auth error
}

if domain.IsRateLimitError(err) {
    // Handle rate limit
}

if domain.IsTimeoutError(err) {
    // Handle timeout
}

if domain.IsProviderUnavailableError(err) {
    // Provider is down
}

if domain.IsContentFilteredError(err) {
    // Content was filtered
}

if domain.IsTokenQuotaExceededError(err) {
    // Out of tokens/credits
}
```

## Provider Errors

Get detailed information about provider errors:

```go
var provErr *domain.ProviderError
if errors.As(err, &provErr) {
    log.Printf("Provider: %s", provErr.Provider)
    log.Printf("Operation: %s", provErr.Operation)
    log.Printf("Status Code: %d", provErr.StatusCode)
    log.Printf("Message: %s", provErr.Message)
    
    // Handle based on status code
    switch provErr.StatusCode {
    case 401:
        // Unauthorized
    case 429:
        // Rate limited
    case 500, 502, 503:
        // Server errors - retry
    }
}
```

## Retry Strategies

### Basic Retry with Backoff

```go
func retryWithBackoff(fn func() error, maxRetries int) error {
    var err error
    
    for i := 0; i < maxRetries; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        
        // Only retry certain errors
        if !isRetryable(err) {
            return err
        }
        
        // Exponential backoff
        backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
        log.Printf("Retry %d/%d after %v", i+1, maxRetries, backoff)
        time.Sleep(backoff)
    }
    
    return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}

func isRetryable(err error) bool {
    return domain.IsRateLimitError(err) ||
           domain.IsNetworkConnectivityError(err) ||
           domain.IsTimeoutError(err) ||
           domain.IsProviderUnavailableError(err)
}
```

### Advanced Retry with Jitter

```go
func retryWithJitter(fn func() error, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        if err := fn(); err == nil {
            return nil
        } else if !isRetryable(err) {
            return err
        }
        
        // Add jitter to prevent thundering herd
        baseDelay := time.Duration(math.Pow(2, float64(i))) * time.Second
        jitter := time.Duration(rand.Int63n(int64(baseDelay / 2)))
        
        time.Sleep(baseDelay + jitter)
    }
    
    return fmt.Errorf("max retries exceeded")
}
```

## Multi-Provider Error Handling

When using multiple providers, handle failures gracefully:

```go
// Check for multi-provider errors
var multiErr *domain.MultiProviderError
if errors.As(err, &multiErr) {
    log.Printf("Multiple providers failed:")
    
    for provider, providerErr := range multiErr.ProviderErrors {
        log.Printf("  %s: %v", provider, providerErr)
        
        // Check each provider's error type
        if domain.IsRateLimitError(providerErr) {
            log.Printf("    %s is rate limited", provider)
        }
    }
}
```

### Fallback Strategy

```go
func generateWithFallback(providers []domain.Provider, prompt string) (string, error) {
    var lastErr error
    
    for _, provider := range providers {
        response, err := provider.Generate(context.Background(), prompt)
        if err == nil {
            return response, nil
        }
        
        lastErr = err
        log.Printf("Provider %T failed: %v, trying next", provider, err)
        
        // Don't retry on auth errors
        if domain.IsAuthenticationError(err) {
            continue
        }
        
        // Wait before trying next provider for rate limits
        if domain.IsRateLimitError(err) {
            time.Sleep(5 * time.Second)
        }
    }
    
    return "", fmt.Errorf("all providers failed, last error: %w", lastErr)
}
```

## Structured Output Errors

Handle validation errors when using structured output:

```go
type Product struct {
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

var product Product
err := provider.GenerateWithSchema(ctx, prompt, &product)
if err != nil {
    if domain.IsInvalidJSON(err) {
        log.Printf("LLM didn't return valid JSON")
        // Retry with clearer instructions
        enhancedPrompt := prompt + "\nPlease respond with valid JSON only."
        err = provider.GenerateWithSchema(ctx, enhancedPrompt, &product)
    }
}
```

### Schema Validation Recovery

```go
func generateStructuredWithRetry(provider domain.Provider, prompt string, schema interface{}) error {
    maxAttempts := 3
    
    for i := 0; i < maxAttempts; i++ {
        err := provider.GenerateWithSchema(ctx, prompt, schema)
        if err == nil {
            return nil
        }
        
        // Enhance prompt based on error
        if domain.IsInvalidJSON(err) {
            prompt = fmt.Sprintf("%s\n\nIMPORTANT: Respond with valid JSON matching this structure: %+v", 
                prompt, schema)
        } else if validationErr, ok := err.(*validation.Error); ok {
            prompt = fmt.Sprintf("%s\n\nFix these validation errors: %v", 
                prompt, validationErr.Details)
        }
    }
    
    return fmt.Errorf("failed to get valid structured output after %d attempts", maxAttempts)
}
```

## Agent Error Handling

Agents handle errors differently for tools vs LLM calls:

### Tool Errors

Tool errors are communicated back to the LLM so it can try alternatives:

```go
// In agent implementation
result, err := tool.Execute(ctx, params)
if err != nil {
    // Tool error becomes part of conversation
    toolMessage := domain.Message{
        Role:    domain.RoleTool,
        Content: fmt.Sprintf("Tool error: %v", err),
    }
    
    // LLM sees the error and can try something else
    messages = append(messages, toolMessage)
    continue // Let agent try alternative approach
}
```

### Agent Retry Logic

```go
func runAgentWithRetry(agent domain.BaseAgent, state *domain.State) (*domain.State, error) {
    maxRetries := 3
    
    for i := 0; i < maxRetries; i++ {
        result, err := agent.Run(context.Background(), state)
        if err == nil {
            return result, nil
        }
        
        // Only retry transient errors
        if !isRetryable(err) {
            return nil, err
        }
        
        log.Printf("Agent failed (attempt %d/%d): %v", i+1, maxRetries, err)
        
        // Add retry context to state
        state.Set("retry_attempt", i+1)
        state.Set("previous_error", err.Error())
        
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    
    return nil, fmt.Errorf("agent failed after %d retries", maxRetries)
}
```

## Context and Timeouts

Always use context for proper timeout handling:

```go
// Set operation timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := provider.Generate(ctx, prompt)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Printf("Operation timed out after 30s")
        // Try with longer timeout or simpler prompt
    }
}
```

### Cascading Timeouts

```go
// Parent context with overall timeout
parentCtx, parentCancel := context.WithTimeout(ctx, 2*time.Minute)
defer parentCancel()

// Individual operation timeouts
for _, task := range tasks {
    opCtx, opCancel := context.WithTimeout(parentCtx, 30*time.Second)
    
    err := processTask(opCtx, task)
    opCancel() // Clean up immediately
    
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            if parentCtx.Err() != nil {
                return fmt.Errorf("overall timeout exceeded")
            }
            log.Printf("Task %s timed out, skipping", task.Name)
            continue
        }
        return err
    }
}
```

## Error Patterns by Use Case

### Chatbot Applications

```go
func handleChatError(err error) string {
    switch {
    case domain.IsRateLimitError(err):
        return "I'm receiving too many requests. Please try again in a moment."
    case domain.IsContentFilteredError(err):
        return "I can't process that request. Please rephrase your question."
    case domain.IsTimeoutError(err):
        return "The response is taking longer than expected. Please try again."
    case domain.IsAuthenticationError(err):
        return "There's a configuration issue. Please contact support."
    default:
        return "I encountered an error. Please try again later."
    }
}
```

### Data Processing Pipelines

```go
func processBatch(items []Item) error {
    var errors []error
    successCount := 0
    
    for _, item := range items {
        if err := processItem(item); err != nil {
            errors = append(errors, fmt.Errorf("item %s: %w", item.ID, err))
            
            // Continue processing other items
            if !isCriticalError(err) {
                continue
            }
            
            // Stop on critical errors
            return fmt.Errorf("critical error processing item %s: %w", item.ID, err)
        }
        successCount++
    }
    
    // Report partial success
    if len(errors) > 0 {
        log.Printf("Processed %d/%d items successfully", successCount, len(items))
        for _, err := range errors {
            log.Printf("Error: %v", err)
        }
    }
    
    return nil
}
```

## Best Practices

### 1. Use Typed Errors
```go
// Good: Check error type
if domain.IsRateLimitError(err) {
    // Handle rate limit
}

// Avoid: String matching
if strings.Contains(err.Error(), "rate limit") {
    // Fragile
}
```

### 2. Add Context
```go
// Good: Wrap with context
return fmt.Errorf("failed to generate response for user %s: %w", userID, err)

// Avoid: Bare error
return err
```

### 3. Log at the Right Level
```go
// Log transient errors as warnings
if domain.IsRateLimitError(err) {
    log.Warn("Rate limited, will retry", "provider", providerName)
}

// Log permanent errors as errors
if domain.IsAuthenticationError(err) {
    log.Error("Authentication failed", "provider", providerName)
}
```

### 4. Fail Fast on Unrecoverable Errors
```go
// Don't retry auth errors
if domain.IsAuthenticationError(err) {
    return fmt.Errorf("invalid credentials: %w", err)
}

// Do retry transient errors
if domain.IsNetworkConnectivityError(err) {
    return retryWithBackoff(operation, 3)
}
```

### 5. Provide User-Friendly Messages
```go
func userFriendlyError(err error) string {
    if technical := os.Getenv("DEBUG"); technical == "true" {
        return err.Error() // Full error in debug mode
    }
    
    // User-friendly messages in production
    switch {
    case domain.IsRateLimitError(err):
        return "Service is busy, please try again later"
    case domain.IsTokenQuotaExceededError(err):
        return "Monthly usage limit reached"
    default:
        return "An error occurred, please try again"
    }
}
```

## Testing Error Scenarios

```go
func TestErrorHandling(t *testing.T) {
    // Test with mock provider that returns specific errors
    mockProvider := provider.NewMockProvider()
    
    // Test rate limit handling
    mockProvider.SetError(domain.ErrRateLimitExceeded)
    _, err := generateWithRetry(mockProvider, "test")
    assert.True(t, domain.IsRateLimitError(err))
    
    // Test authentication error (should not retry)
    mockProvider.SetError(domain.ErrAuthenticationFailed)
    _, err = generateWithRetry(mockProvider, "test")
    assert.True(t, domain.IsAuthenticationError(err))
    assert.Equal(t, 1, mockProvider.CallCount()) // No retries
}
```

## Next Steps

- Learn about [Multi-Provider](providers.md#multi-provider-setup) strategies for reliability
- Explore [Agent](agents.md) error recovery patterns
- See [Examples Gallery](examples-gallery.md) for error handling in practice

Remember: Good error handling is what separates a demo from a production-ready application! 🛡️