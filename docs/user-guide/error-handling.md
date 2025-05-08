# Error Handling in Go-LLMs

> **[Documentation Home](/REFERENCE.md) / [User Guide](/docs/user-guide/) / Error Handling**

This document describes the error handling patterns and best practices used in the Go-LLMs library.

*Related: [Getting Started](getting-started.md) | [Multi-Provider Guide](multi-provider.md) | [API Reference](/docs/api/README.md)*

## Core Principles

1. **Standardized Error Types**: Use a set of predefined error types for common error cases.
2. **Error Wrapping**: Wrap errors with additional context while preserving the original error.
3. **Error Classification**: Provide helpers to classify errors by type.
4. **Provider-specific Information**: Include provider-specific details in errors.
5. **Multiple Provider Handling**: Special handling for aggregating errors from multiple providers.
6. **Graceful Degradation**: Implement fallback mechanisms to handle errors in a way that preserves functionality.
7. **Contextual Error Handling**: Handle errors according to the specific context of the operation.

## Standard Error Types

All standard error types are defined in `pkg/llm/domain/errors.go`:

```go
var (
    // ErrRequestFailed is returned when a request to an LLM provider fails.
    ErrRequestFailed = errors.New("request to LLM provider failed")
    
    // ErrResponseParsing is returned when the response from an LLM provider cannot be parsed.
    ErrResponseParsing = errors.New("failed to parse LLM provider response")
    
    // ErrInvalidJSON is returned when the LLM response does not contain valid JSON.
    ErrInvalidJSON = errors.New("response does not contain valid JSON")
    
    // ErrAuthenticationFailed is returned when authentication with the LLM provider fails.
    ErrAuthenticationFailed = errors.New("authentication with LLM provider failed")
    
    // ErrRateLimitExceeded is returned when the LLM provider rate limit is exceeded.
    ErrRateLimitExceeded = errors.New("rate limit exceeded")
    
    // ErrContextTooLong is returned when the input context is too long for the model.
    ErrContextTooLong = errors.New("input context too long")
    
    // ErrProviderUnavailable is returned when the LLM provider is unavailable.
    ErrProviderUnavailable = errors.New("LLM provider unavailable")
    
    // ErrInvalidConfiguration is returned when the provider configuration is invalid.
    ErrInvalidConfiguration = errors.New("invalid provider configuration")
    
    // ErrNoResponse is returned when the LLM provider returns no response.
    ErrNoResponse = errors.New("no response from LLM provider")
    
    // ErrTimeout is returned when a request to an LLM provider times out.
    ErrTimeout = errors.New("LLM provider request timed out")
    
    // ErrContentFiltered is returned when content is filtered by the LLM provider.
    ErrContentFiltered = errors.New("content filtered by LLM provider")
    
    // ErrModelNotFound is returned when the requested model is not found.
    ErrModelNotFound = errors.New("model not found")
    
    // ErrNetworkConnectivity is returned when there are network issues connecting to the provider.
    ErrNetworkConnectivity = errors.New("network connectivity issues")
    
    // ErrTokenQuotaExceeded is returned when the user has exceeded their token quota.
    ErrTokenQuotaExceeded = errors.New("token quota exceeded")
    
    // ErrInvalidModelParameters is returned when provided model parameters are invalid.
    ErrInvalidModelParameters = errors.New("invalid model parameters")
)
```

## Provider Error Type

The `ProviderError` type provides detailed information about errors from providers:

```go
// ProviderError represents an error from an LLM provider with additional context.
type ProviderError struct {
    // Provider is the name of the LLM provider (e.g., "openai", "anthropic").
    Provider string
    
    // Operation is the operation that failed (e.g., "Generate", "Stream").
    Operation string
    
    // StatusCode is the HTTP status code if applicable.
    StatusCode int
    
    // Message is the error message from the provider.
    Message string
    
    // Err is the underlying error.
    Err error
}
```

The `ProviderError` implements:
- `Error() string`: Returns a formatted error message including provider, operation, status code, and message
- `Unwrap() error`: Returns the underlying error for error wrapping and checking with `errors.Is()`

## Multi-Provider Error Handling

For the `MultiProvider` implementation, we use a special error type that aggregates errors from multiple providers:

```go
// MultiProviderError represents an error from multiple providers
type MultiProviderError struct {
    // ProviderErrors contains the errors from each provider
    ProviderErrors map[string]error
    
    // Message is the overall error message
    Message string
}
```

The `MultiProviderError` implements:
- `Error() string`: Returns a formatted error message including all provider errors
- `Unwrap() error`: Returns the first error in the map (for compatibility with unwrap)
- `Is(target error) bool`: Checks if any of the provider errors match the target error

## Error Classification

Helper functions are provided to classify errors:

```go
// IsAuthenticationError checks if the error is an authentication error.
func IsAuthenticationError(err error) bool {
    return errors.Is(err, ErrAuthenticationFailed)
}

// IsRateLimitError checks if the error is a rate limit error.
func IsRateLimitError(err error) bool {
    return errors.Is(err, ErrRateLimitExceeded)
}

// IsTimeoutError checks if the error is a timeout error.
func IsTimeoutError(err error) bool {
    return errors.Is(err, ErrTimeout)
}

// IsProviderUnavailableError checks if the error is a provider unavailable error.
func IsProviderUnavailableError(err error) bool {
    return errors.Is(err, ErrProviderUnavailable)
}

// IsContentFilteredError checks if the error is a content filtered error.
func IsContentFilteredError(err error) bool {
    return errors.Is(err, ErrContentFiltered)
}

// IsNetworkConnectivityError checks if the error is a network connectivity error.
func IsNetworkConnectivityError(err error) bool {
    return errors.Is(err, ErrNetworkConnectivity)
}

// IsTokenQuotaExceededError checks if the error is a token quota exceeded error.
func IsTokenQuotaExceededError(err error) bool {
    return errors.Is(err, ErrTokenQuotaExceeded)
}

// IsInvalidModelParametersError checks if the error is an invalid model parameters error.
func IsInvalidModelParametersError(err error) bool {
    return errors.Is(err, ErrInvalidModelParameters)
}
```

## Provider-Specific Error Mapping

Each provider implements its own error mapping function to convert provider-specific errors to the standard error types. 

### OpenAI Error Mapping

```go
// mapOpenAIErrorToStandard maps OpenAI API error messages to standard error types
func mapOpenAIErrorToStandard(statusCode int, errorMsg string, operation string) error {
    // Convert error message to lowercase for case-insensitive matching
    lowerErrorMsg := strings.ToLower(errorMsg)

    // Common error patterns for OpenAI
    switch {
    case statusCode == http.StatusUnauthorized || strings.Contains(lowerErrorMsg, "invalid api key"):
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrAuthenticationFailed)
        
    case statusCode == http.StatusTooManyRequests || strings.Contains(lowerErrorMsg, "rate limit"):
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrRateLimitExceeded)
        
    case strings.Contains(lowerErrorMsg, "context length"):
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrContextTooLong)
        
    case strings.Contains(lowerErrorMsg, "content filter"):
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrContentFiltered)
        
    case strings.Contains(lowerErrorMsg, "model not found"):
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrModelNotFound)
        
    case strings.Contains(lowerErrorMsg, "quota") || strings.Contains(lowerErrorMsg, "billing"):
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrTokenQuotaExceeded)
        
    case strings.Contains(lowerErrorMsg, "invalid parameter") || strings.Contains(lowerErrorMsg, "invalid request"):
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrInvalidModelParameters)
        
    case statusCode == http.StatusServiceUnavailable || 
         statusCode == http.StatusBadGateway || 
         statusCode == http.StatusGatewayTimeout:
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrNetworkConnectivity)
        
    case statusCode >= 500:
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrProviderUnavailable)
        
    default:
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrRequestFailed)
    }
}
```

### Anthropic Error Mapping

```go
// mapAnthropicErrorToStandard maps Anthropic API error messages to standard error types
func mapAnthropicErrorToStandard(statusCode int, errorType, errorMsg string, operation string) error {
    lowerErrorMsg := strings.ToLower(errorMsg)
    lowerErrorType := strings.ToLower(errorType)

    switch {
    case statusCode == http.StatusUnauthorized || 
         strings.Contains(lowerErrorType, "authentication") || 
         strings.Contains(lowerErrorMsg, "api key"):
        return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrAuthenticationFailed)
        
    case statusCode == http.StatusTooManyRequests || 
         strings.Contains(lowerErrorType, "rate_limit") || 
         strings.Contains(lowerErrorMsg, "rate limit"):
        return domain.NewProviderError("anthropic", operation, statusCode, errorMsg, domain.ErrRateLimitExceeded)
    
    // Additional mappings for other error types...
    }
}
```

## Common Error Handling Patterns

### Basic Error Handling

The most basic pattern is to check for errors and handle them appropriately:

```go
response, err := provider.Generate(ctx, prompt)
if err != nil {
    if domain.IsAuthenticationError(err) {
        // Handle authentication errors
        fmt.Println("Authentication failed, please check your API key")
    } else if domain.IsRateLimitError(err) {
        // Handle rate limit errors
        fmt.Println("Rate limit exceeded, please try again later")
    } else if domain.IsTimeoutError(err) {
        // Handle timeout errors
        fmt.Println("Request timed out, please try again")
    } else if domain.IsNetworkConnectivityError(err) {
        // Handle network connectivity errors
        fmt.Println("Network connectivity issues, check your internet connection")
    } else if domain.IsTokenQuotaExceededError(err) {
        // Handle token quota exceeded errors
        fmt.Println("Token quota exceeded, check your billing and usage limits")
    } else if domain.IsInvalidModelParametersError(err) {
        // Handle invalid model parameters errors
        fmt.Println("Invalid model parameters provided, please check your request")
    } else if domain.IsContentFilteredError(err) {
        // Handle content filtered errors
        fmt.Println("Content was filtered by the provider, please revise your prompt")
    } else {
        // Handle other errors
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

### Handling Multi-Provider Errors

When using the `MultiProvider`, you should check for `MultiProviderError` and handle it accordingly:

```go
response, err := multiProvider.Generate(ctx, prompt)
if err != nil {
    // Check if it's a multi-provider error
    var multiErr *provider.MultiProviderError
    if errors.As(err, &multiErr) {
        fmt.Println("Errors from multiple providers:")
        for providerName, providerErr := range multiErr.ProviderErrors {
            fmt.Printf("  - %s: %v\n", providerName, providerErr)
            
            // You can further classify each provider's error
            if domain.IsRateLimitError(providerErr) {
                fmt.Printf("    Provider %s is rate limited\n", providerName)
            }
        }
    } else {
        // Handle other errors
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

### Retry Logic for Transient Errors

Some errors are transient and can be retried:

```go
const maxRetries = 3

var response domain.Response
var err error

for i := 0; i < maxRetries; i++ {
    response, err = provider.Generate(ctx, prompt)
    
    // If successful, break out of the retry loop
    if err == nil {
        break
    }
    
    // Only retry on certain errors
    if domain.IsRateLimitError(err) || 
       domain.IsNetworkConnectivityError(err) || 
       domain.IsProviderUnavailableError(err) || 
       domain.IsTimeoutError(err) {
        
        // Exponential backoff
        backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
        fmt.Printf("Retrying after %v due to error: %v\n", backoff, err)
        time.Sleep(backoff)
        continue
    }
    
    // Don't retry on other errors
    break
}

if err != nil {
    // Handle the error after all retries failed
    fmt.Printf("Failed after %d retries: %v\n", maxRetries, err)
    return
}
```

### Graceful Degradation with Fallbacks

Implement fallbacks for when primary providers fail:

```go
// Try the primary provider first
response, err := primaryProvider.Generate(ctx, prompt)
if err != nil {
    fmt.Printf("Primary provider failed: %v, falling back to secondary\n", err)
    
    // Fall back to the secondary provider
    response, err = secondaryProvider.Generate(ctx, prompt)
    if err != nil {
        fmt.Printf("Secondary provider also failed: %v\n", err)
        return nil, fmt.Errorf("all providers failed: %w", err)
    }
}
```

## Structured Output Error Handling

When working with structured outputs, additional error handling is needed for schema validation:

```go
// Define a schema for the output
schema := &domain.Schema{
    Type: domain.ObjectType,
    Properties: map[string]*domain.Schema{
        "name": {Type: domain.StringType},
        "age":  {Type: domain.IntegerType},
    },
    Required: []string{"name", "age"},
}

// Process with schema validation
result, err := processor.Process(ctx, llmResponse, schema)
if err != nil {
    if domain.IsInvalidJSON(err) {
        // Handle invalid JSON errors
        fmt.Println("LLM response did not contain valid JSON")
    } else if err is a validation error {
        // Handle schema validation errors
        fmt.Println("LLM response did not match the expected schema")
        fmt.Printf("Validation error: %v\n", err)
        
        // You might want to retry with a more explicit prompt
        enhancedPrompt := prompt + "\nPlease ensure your response includes 'name' as a string and 'age' as an integer."
        return retryWithEnhancedPrompt(ctx, provider, enhancedPrompt, schema)
    } else {
        // Handle other errors
        fmt.Printf("Error processing response: %v\n", err)
    }
    return nil, err
}
```

## Agent Error Handling

The Agent system has specific error handling considerations:

1. **Tool Execution Errors**:
   - Tool errors are communicated back to the LLM to allow the agent to try alternate approaches
   - They don't necessarily fail the entire agent process

```go
// Execute a tool and handle errors
result, err := tool.Execute(ctx, params)
if err != nil {
    // Log the error for debugging
    log.Printf("Tool %s failed: %v", tool.Name(), err)
    
    // Format the error for the LLM
    toolResponse := fmt.Sprintf("Error: %v", err)
    
    // Add to the conversation so the LLM can try an alternative
    conversation = append(conversation, domain.Message{
        Role:    "function",
        Content: toolResponse,
        Name:    tool.Name(),
    })
    
    // Continue the agent process
    return continueWithLLM(ctx, conversation)
}
```

2. **LLM Generation Errors**:
   - These are more severe and typically fail the agent process
   - Implement retries for transient errors

```go
// Get a response from the LLM
response, err := llm.GenerateMessage(ctx, conversation)
if err != nil {
    // Check for retryable errors
    if domain.IsRateLimitError(err) || domain.IsNetworkConnectivityError(err) {
        // Retry after a backoff
        time.Sleep(backoff)
        return retry(ctx, conversation, backoff*2)
    }
    
    // Non-retryable errors fail the agent process
    return nil, fmt.Errorf("agent failed due to LLM error: %w", err)
}
```

## Best Practices

### Provider Integration Best Practices

When implementing new provider integrations, follow these best practices:

1. **Use Standard Error Types**: Map provider-specific errors to the standard error types defined in `pkg/llm/domain/errors.go`.

2. **Include Context**: When creating new errors, include enough context to help debugging:
   - Provider name
   - Operation that failed
   - HTTP status code if applicable
   - Original error message

3. **Preserve Original Errors**: Use `fmt.Errorf("error message: %w", err)` or `domain.NewProviderError()` to preserve the original error for error checking.

4. **Error Mapping Logic**: Implement a provider-specific error mapping function that:
   - Analyzes status codes
   - Looks for patterns in error messages
   - Maps to the most specific standard error type

5. **Handle Rate Limits**: Implement proper handling for rate limit errors:
   - Include information about retry-after intervals
   - Consider implementing automatic retries with backoff

### Client Code Best Practices

When using the library in client code, follow these best practices:

1. **Check Error Types**: Use `errors.Is()` and `errors.As()` to check error types, rather than string matching:
   ```go
   if domain.IsRateLimitError(err) {
       // Handle rate limiting
   }
   ```

2. **Return Specific Errors**: Return the most specific error type possible to help callers handle errors appropriately.

3. **Implement Retries**: Implement retry logic for transient errors:
   - Rate limit errors
   - Network connectivity issues
   - Provider unavailability
   - Timeouts

4. **Use Timeouts**: Always set appropriate context timeouts for LLM operations:
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   
   response, err := provider.Generate(ctx, prompt)
   ```

5. **Graceful Degradation**: Implement fallback mechanisms when possible:
   - Try alternative providers
   - Retry with different model parameters
   - Use cached results as a last resort

6. **Log Detailed Errors**: Log detailed error information for debugging:
   ```go
   if err != nil {
       log.Printf("LLM error: %v (%T)", err, err)
       
       // Extract provider-specific details
       var provErr *domain.ProviderError
       if errors.As(err, &provErr) {
           log.Printf("Provider: %s, Operation: %s, Status: %d, Message: %s",
               provErr.Provider, provErr.Operation, provErr.StatusCode, provErr.Message)
       }
   }
   ```

7. **Handle Context Cancellation**: Always check for context cancellation:
   ```go
   if errors.Is(err, context.Canceled) {
       // Handle cancellation
       return nil, fmt.Errorf("operation was canceled: %w", err)
   }
   ```

## Conclusion

The Go-LLMs library provides a comprehensive error handling system that balances:

1. **Standardization**: Common error types across providers
2. **Context**: Detailed information about what went wrong
3. **Specificity**: Fine-grained error classification
4. **Preservability**: Error wrapping to maintain context
5. **Usability**: Helper functions for error checking

By following the patterns and best practices outlined in this document, you can build robust applications that gracefully handle errors from LLM providers and provide a better user experience.