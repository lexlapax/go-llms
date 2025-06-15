# Error Handling Migration Guide

## Overview

This guide covers the migration from standard Go errors to the enhanced serializable error system in go-llms v0.3.5.2. This change enables JSON serialization for all errors, making them compatible with downstream projects like go-llmspell.

## What Changed

### Before (v0.3.5.1 and earlier)
```go
// Standard Go errors
var ErrAgentNotFound = errors.New("agent not found")

// Custom error structs
type ProviderError struct {
    Provider string
    Message  string
    Err      error
}
```

### After (v0.3.5.2+)
```go
// Serializable errors with metadata
var ErrAgentNotFound = errors.NewErrorWithCode("agent_not_found", "agent not found").SetFatal(true)

// Enhanced error structs with BaseError embedding
type ProviderError struct {
    *errors.BaseError
    Provider   string `json:"provider"`
    StatusCode int    `json:"status_code,omitempty"`
}
```

## Key Benefits

1. **JSON Serialization**: All errors can be serialized to JSON for logging, debugging, and bridge compatibility
2. **Rich Context**: Errors include structured context, stack traces, and metadata
3. **Recovery Strategies**: Built-in support for retry logic and error recovery
4. **Bridge Compatibility**: Seamless integration with scripting engines and external systems

## Migration Steps

### 1. Update Error Creation

#### Old Pattern
```go
return fmt.Errorf("failed to connect to %s: %w", provider, err)
```

#### New Pattern
```go
return errors.Wrap(err, "failed to connect").
    WithContext("provider", provider).
    SetRetryable(true)
```

### 2. Convert Custom Error Types

#### Old Pattern
```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}
```

#### New Pattern
```go
type ValidationError struct {
    *errors.BaseError
    Field string `json:"field"`
    Value interface{} `json:"value,omitempty"`
}

func NewValidationError(field string, value interface{}, message string) *ValidationError {
    baseErr := errors.Wrap(ErrSchemaValidation, message)
    baseErr.WithContext("field", field).
        WithContext("value", value).
        WithType("ValidationError").
        SetFatal(true)

    return &ValidationError{
        BaseError: baseErr,
        Field:     field,
        Value:     value,
    }
}
```

### 3. Update Error Checking

#### Old Pattern
```go
if errors.Is(err, ErrAgentNotFound) {
    // handle error
}
```

#### New Pattern
```go
// Same pattern works, but now with enhanced capabilities
if errors.As(err, &ErrAgentNotFound) {
    // handle error
}

// Or use new capabilities
if errors.IsRetryableError(err) {
    // retry logic
}

if errors.IsFatalError(err) {
    // stop execution
}
```

## Error Categories and Codes

### LLM Provider Errors
- `llm_request_failed` - General request failure (retryable)
- `llm_auth_failed` - Authentication failed (fatal)
- `llm_rate_limit` - Rate limit exceeded (retryable)
- `llm_timeout` - Request timeout (retryable)
- `llm_provider_unavailable` - Provider unavailable (retryable)
- `llm_model_not_found` - Model not found (fatal)
- `llm_quota_exceeded` - Token quota exceeded (fatal)

### Agent Errors
- `agent_not_found` - Agent not found (fatal)
- `agent_timeout` - Execution timeout (retryable)
- `agent_cancelled` - Execution canceled
- `tool_not_found` - Tool not found (fatal)
- `tool_execution_failed` - Tool execution failed (retryable)
- `schema_validation_failed` - Schema validation failed (fatal)

## Working with Serializable Errors

### JSON Serialization
```go
// Create an enhanced error
err := domain.NewProviderError("openai", "Generate", 429, "Rate limit exceeded", nil)

// Serialize to JSON
if serErr, ok := err.(errors.SerializableError); ok {
    jsonData, err := serErr.ToJSON()
    if err == nil {
        fmt.Printf("Error JSON: %s\n", jsonData)
    }
}
```

### Context Extraction
```go
// Extract structured context
context := errors.GetErrorContext(err)
if context != nil {
    provider := context["provider"]
    statusCode := context["status_code"]
    // Use context for logging, debugging, etc.
}
```

### Recovery Strategies
```go
// Create error with recovery strategy
err := errors.NewError("temporary failure").
    SetRetryable(true).
    WithRecovery(errors.ExponentialBackoffStrategy)

// Check if error has recovery strategy
if strategy := err.GetRecoveryStrategy(); strategy != nil {
    if strategy.CanRecover(err) {
        // Attempt recovery
        recoveryErr := strategy.Recover(err, context)
    }
}
```

## Bridge Compatibility

The enhanced error system is designed for seamless integration with bridge layers (like go-llmspell):

```go
// All errors are bridge-compatible
func handleError(err error) map[string]interface{} {
    if serErr, ok := err.(errors.SerializableError); ok {
        jsonData, _ := serErr.ToJSON()
        
        var bridgeData map[string]interface{}
        json.Unmarshal(jsonData, &bridgeData)
        
        return bridgeData // Ready for script consumption
    }
    
    // Fallback for non-serializable errors
    return map[string]interface{}{
        "error": err.Error(),
        "type": "unknown",
    }
}
```

## Best Practices

### 1. Use Appropriate Error Codes
- Choose descriptive, consistent error codes
- Follow the pattern: `{domain}_{specific_error}`
- Example: `llm_rate_limit`, `agent_timeout`, `tool_not_found`

### 2. Set Retryability Appropriately
```go
// Retryable errors (temporary failures)
errors.NewErrorWithCode("network_timeout", "Network timeout").SetRetryable(true)

// Fatal errors (permanent failures)
errors.NewErrorWithCode("invalid_config", "Invalid configuration").SetFatal(true)
```

### 3. Include Rich Context
```go
err.WithContext("user_id", userID).
    WithContext("operation", "generate").
    WithContext("model", modelName).
    WithContext("retry_count", retryCount)
```

### 4. Preserve Error Chains
```go
// Always wrap, don't replace
return errors.Wrap(originalErr, "operation failed").
    WithContext("step", "validation")
```

## Testing Serializable Errors

```go
func TestErrorSerialization(t *testing.T) {
    err := domain.NewProviderError("openai", "Generate", 500, "Server error", nil)
    
    // Test JSON serialization
    if serErr, ok := err.(errors.SerializableError); ok {
        jsonData, err := serErr.ToJSON()
        require.NoError(t, err)
        
        // Verify JSON structure
        var data map[string]interface{}
        err = json.Unmarshal(jsonData, &data)
        require.NoError(t, err)
        
        assert.Equal(t, "ProviderError", data["type"])
        assert.Equal(t, "openai", data["context"].(map[string]interface{})["provider"])
    }
}
```

## Common Migration Issues

### 1. Type Assertions
**Problem**: Old type assertions may fail
```go
if pe, ok := err.(*ProviderError); ok { // May fail
```

**Solution**: Use errors.As
```go
var pe *ProviderError
if errors.As(err, &pe) { // Always works
```

### 2. Error Wrapping
**Problem**: Lost error context
```go
return fmt.Errorf("failed: %w", err) // Loses metadata
```

**Solution**: Use enhanced wrapping
```go
return errors.Wrap(err, "failed").WithContext("operation", "migrate")
```

### 3. Nil Error Handling
**Problem**: Creating errors with nil BaseError
```go
return &CustomError{BaseError: nil} // Will panic
```

**Solution**: Always create BaseError properly
```go
baseErr := errors.NewError("message")
return &CustomError{BaseError: baseErr}
```

## Backwards Compatibility

The migration maintains backwards compatibility:

1. **Error Interface**: All enhanced errors still implement `error`
2. **Error Checking**: `errors.Is()` and `errors.As()` continue to work
3. **Error Messages**: Human-readable messages remain unchanged
4. **Wrapping**: `fmt.Errorf()` still works, but loses enhanced features

## Performance Considerations

The enhanced error system has minimal performance impact:

- **Stack Trace**: Captured only once at creation
- **Context**: Stored as `map[string]interface{}` (efficient for small contexts)
- **Serialization**: Only performed when explicitly requested
- **Memory**: Slightly higher due to additional metadata (~100-200 bytes per error)

## Conclusion

The enhanced error system provides significant benefits for debugging, monitoring, and system integration while maintaining full backwards compatibility. The migration is designed to be gradual and non-breaking, allowing teams to adopt enhanced features at their own pace.

For questions or issues with the migration, refer to the error handling examples in `cmd/examples/enhanced-errors/` or consult the API documentation.