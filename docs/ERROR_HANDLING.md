# Error Handling in Go-LLMs

This document describes the error handling patterns used in the Go-LLMs library.

## Core Principles

1. **Standardized Error Types**: Use a set of predefined error types for common error cases.
2. **Error Wrapping**: Wrap errors with additional context while preserving the original error.
3. **Error Classification**: Provide helpers to classify errors by type.
4. **Provider-specific Information**: Include provider-specific details in errors.
5. **Multiple Provider Handling**: Special handling for aggregating errors from multiple providers.

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

## Usage Examples

### Handling Provider Errors

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

```go
response, err := multiProvider.Generate(ctx, prompt)
if err != nil {
    // Check if it's a multi-provider error
    var multiErr *provider.MultiProviderError
    if errors.As(err, &multiErr) {
        fmt.Println("Errors from multiple providers:")
        for providerName, providerErr := range multiErr.ProviderErrors {
            fmt.Printf("  - %s: %v\n", providerName, providerErr)
        }
    } else {
        // Handle other errors
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

## Provider-Specific Error Mapping

Each provider implements its own error mapping function to convert provider-specific errors to the standard error types. For example:

```go
// mapOpenAIErrorToStandard maps OpenAI API error messages to standard error types
func mapOpenAIErrorToStandard(statusCode int, errorMsg string, operation string) error {
    lowerErrorMsg := strings.ToLower(errorMsg)

    switch {
    case statusCode == http.StatusUnauthorized || strings.Contains(lowerErrorMsg, "invalid api key"):
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrAuthenticationFailed)
        
    case statusCode == http.StatusTooManyRequests || strings.Contains(lowerErrorMsg, "rate limit"):
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrRateLimitExceeded)
        
    // More mappings...
    
    default:
        return domain.NewProviderError("openai", operation, statusCode, errorMsg, domain.ErrRequestFailed)
    }
}
```

## Best Practices

When implementing new provider integrations or client code, follow these best practices:

1. **Use Standard Error Types**: Use the standard error types defined in `pkg/llm/domain/errors.go` when possible.
2. **Include Context**: When creating new errors, include enough context to help debugging.
3. **Preserve Original Errors**: Use `fmt.Errorf("error message: %w", err)` or `domain.NewProviderError()` to preserve the original error.
4. **Check Error Types**: Use `errors.Is()` and `errors.As()` to check error types, rather than string matching.
5. **Return Specific Errors**: Return the most specific error type possible to help callers handle errors appropriately.

## Conclusion

The consistent error handling pattern in Go-LLMs simplifies error handling for clients and improves the debugging experience. By using standardized error types and wrapping with context, we can provide more helpful error messages while maintaining the ability to programmatically check error types.