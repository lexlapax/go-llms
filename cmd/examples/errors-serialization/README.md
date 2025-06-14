# Enhanced Error Handling Example

This example demonstrates the comprehensive error handling features in go-llms, including serializable errors, recovery strategies, and error aggregation.

## Features Demonstrated

### 1. Serializable Errors
- JSON serialization of errors for cross-language compatibility
- Rich error context with stack traces
- Error codes and types for categorization

### 2. Recovery Strategies
- **Exponential Backoff**: Retry with increasing delays
- **Linear Backoff**: Retry with fixed delay increments
- **Circuit Breaker**: Prevent cascading failures
- **Fallback Strategy**: Use alternative values/methods
- **Composite Strategy**: Combine multiple strategies

### 3. Error Context Enhancement
- Automatic context collection
- Operation, request, and resource context
- Runtime information capture
- Error builder pattern for fluent error creation

### 4. Error Aggregation
- Collect multiple errors from batch operations
- Preserve individual error contexts
- Serializable aggregated errors

## Running the Example

```bash
go run main.go
```

## Example Output

The example will demonstrate:

1. Basic error creation with JSON serialization
2. Retry logic with exponential backoff
3. Error aggregation from multiple operations
4. Context enrichment with operation details
5. Error builder for complex error creation
6. Circuit breaker pattern in action
7. Composite recovery strategies

## Key Concepts

### SerializableError Interface
```go
type SerializableError interface {
    error
    ToJSON() ([]byte, error)
    GetContext() map[string]interface{}
    GetRecoveryStrategy() RecoveryStrategy
}
```

### Creating Errors with Context
```go
err := errors.NewError("operation failed").
    WithCode("OP_FAILED").
    SetRetryable(true).
    WithContext("operation", "user_fetch").
    WithContext("user_id", "12345")
```

### Using Recovery Strategies
```go
strategy := errors.NewExponentialBackoffStrategy(
    5,                    // max attempts
    100*time.Millisecond, // base delay
    5*time.Second,        // max delay
)

err := errors.NewError("temporary failure").
    SetRetryable(true).
    WithRecovery(strategy)
```

### Error Aggregation
```go
agg := errors.NewErrorAggregator()
for _, op := range operations {
    if err := op.Execute(); err != nil {
        agg.AddWithContext(err, map[string]interface{}{
            "operation": op.Name,
            "timestamp": time.Now(),
        })
    }
}

if agg.HasErrors() {
    serializable := agg.ToSerializable()
    data, _ := serializable.ToJSON()
    fmt.Println(string(data))
}
```

## Use Cases

1. **API Client Libraries**: Serialize errors for debugging across services
2. **Batch Processing**: Aggregate errors from multiple operations
3. **Resilient Systems**: Implement retry logic with backoff strategies
4. **Monitoring**: Rich error context for better observability
5. **Scripting Integration**: JSON serialization for cross-language error handling

## Integration with go-llmspell

This error handling system is designed to work seamlessly with go-llmspell's scripting engine:

- Errors can be serialized to JSON for script consumption
- Recovery strategies can be configured from scripts
- Error context provides detailed debugging information
- Aggregation supports batch operations common in scripts