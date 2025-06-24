# Error Codes Reference: Complete Error Handling Guide

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Reference](/docs/user-guide/reference/) / Error Codes**

Comprehensive guide to all error codes, types, and handling strategies in Go-LLMs. Learn how to identify, handle, and recover from various error conditions.

## Error Type Hierarchy

```
errors.Error
├── ProviderError
│   ├── AuthenticationError
│   ├── RateLimitError
│   ├── QuotaExceededError
│   ├── ModelNotFoundError
│   └── InvalidRequestError
├── AgentError
│   ├── ToolExecutionError
│   ├── WorkflowError
│   ├── MemoryError
│   └── TimeoutError
├── ValidationError
│   ├── SchemaValidationError
│   ├── InputValidationError
│   └── OutputValidationError
└── SystemError
    ├── ConfigurationError
    ├── NetworkError
    └── InternalError
```

---

## Provider Errors

### Authentication Errors (1xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 1001 | `InvalidAPIKey` | API key is invalid or malformed | Check API key format and validity |
| 1002 | `ExpiredAPIKey` | API key has expired | Renew or rotate API key |
| 1003 | `MissingAPIKey` | API key not provided | Set environment variable or config |
| 1004 | `UnauthorizedAccess` | Access denied for resource | Check permissions and scopes |
| 1005 | `InvalidOrganization` | Organization ID invalid | Verify organization settings |

**Example Handling:**
```go
err := provider.Complete(ctx, request)
if err != nil {
    var authErr *errors.AuthenticationError
    if errors.As(err, &authErr) {
        switch authErr.Code {
        case errors.InvalidAPIKey:
            // Prompt for new API key
        case errors.ExpiredAPIKey:
            // Attempt key rotation
        default:
            // Generic auth error handling
        }
    }
}
```

### Rate Limit Errors (2xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 2001 | `RateLimitExceeded` | Too many requests | Implement exponential backoff |
| 2002 | `TokenLimitExceeded` | Token quota exceeded | Wait for quota reset |
| 2003 | `ConcurrentLimitExceeded` | Too many concurrent requests | Reduce parallelism |
| 2004 | `DailyQuotaExceeded` | Daily limit reached | Wait until next day |
| 2005 | `MinuteQuotaExceeded` | Per-minute limit hit | Brief wait required |

**Example Handling:**
```go
func handleRateLimit(err error) {
    var rateErr *errors.RateLimitError
    if errors.As(err, &rateErr) {
        // Extract retry-after header
        retryAfter := rateErr.RetryAfter
        if retryAfter > 0 {
            time.Sleep(time.Duration(retryAfter) * time.Second)
        } else {
            // Exponential backoff
            backoff := time.Second * time.Duration(math.Pow(2, float64(attempt)))
            time.Sleep(backoff)
        }
    }
}
```

### Model Errors (3xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 3001 | `ModelNotFound` | Requested model doesn't exist | Use fallback model |
| 3002 | `ModelDeprecated` | Model is deprecated | Migrate to newer model |
| 3003 | `ModelUnavailable` | Model temporarily unavailable | Retry or use alternative |
| 3004 | `ModelCapabilityError` | Model lacks required capability | Switch to capable model |
| 3005 | `ModelLoadError` | Failed to load model (Ollama) | Check model installation |

**Example Handling:**
```go
// Fallback model strategy
models := []string{"gpt-4o", "gpt-4o-mini", "gpt-3.5-turbo"}
var lastErr error

for _, model := range models {
    err := agent.Complete(ctx, request, WithModel(model))
    if err == nil {
        break
    }
    
    var modelErr *errors.ModelNotFoundError
    if errors.As(err, &modelErr) {
        lastErr = err
        continue // Try next model
    }
    return err // Other error, don't retry
}
```

### Request Errors (4xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 4001 | `InvalidRequest` | Malformed request | Fix request structure |
| 4002 | `ContentTooLarge` | Request exceeds size limit | Chunk or summarize content |
| 4003 | `InvalidParameters` | Invalid parameter values | Validate parameters |
| 4004 | `UnsupportedOperation` | Operation not supported | Use alternative approach |
| 4005 | `ContentPolicyViolation` | Content flagged by safety filters | Modify content |

---

## Agent Errors

### Tool Execution Errors (5xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 5001 | `ToolNotFound` | Tool doesn't exist | Check tool registration |
| 5002 | `ToolExecutionFailed` | Tool execution error | Check tool input/state |
| 5003 | `ToolTimeout` | Tool execution timeout | Increase timeout or optimize |
| 5004 | `ToolPermissionDenied` | Insufficient permissions | Check tool permissions |
| 5005 | `ToolInputInvalid` | Invalid tool input | Validate against schema |

**Example Handling:**
```go
result, err := agent.ExecuteTool(ctx, "file_read", input)
if err != nil {
    var toolErr *errors.ToolExecutionError
    if errors.As(err, &toolErr) {
        // Log detailed error information
        log.Printf("Tool %s failed: %s", toolErr.ToolName, toolErr.Details)
        
        // Attempt recovery
        if toolErr.Code == errors.ToolTimeout {
            // Retry with increased timeout
            ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
            defer cancel()
            result, err = agent.ExecuteTool(ctx, "file_read", input)
        }
    }
}
```

### Workflow Errors (6xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 6001 | `WorkflowStepFailed` | Individual step failed | Retry step or skip |
| 6002 | `WorkflowTimeout` | Entire workflow timeout | Increase timeout or checkpoint |
| 6003 | `WorkflowCancelled` | Workflow was cancelled | Clean up and exit |
| 6004 | `CircularDependency` | Circular workflow detected | Fix workflow design |
| 6005 | `InvalidWorkflowState` | Corrupted workflow state | Reset workflow |

### Memory Errors (7xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 7001 | `MemoryFull` | Memory capacity exceeded | Clear old entries |
| 7002 | `MemoryCorrupted` | Memory state corrupted | Reset memory |
| 7003 | `MemoryPersistenceFailed` | Failed to save memory | Check storage |
| 7004 | `MemoryLoadFailed` | Failed to load memory | Use fresh memory |
| 7005 | `MemoryTypeInvalid` | Invalid memory type | Check configuration |

---

## Validation Errors

### Schema Validation Errors (8xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 8001 | `SchemaInvalid` | JSON Schema is invalid | Fix schema definition |
| 8002 | `ValidationFailed` | Data doesn't match schema | Fix data or relax schema |
| 8003 | `RequiredFieldMissing` | Required field not present | Add missing field |
| 8004 | `TypeMismatch` | Wrong data type | Convert or coerce type |
| 8005 | `PatternMismatch` | String doesn't match pattern | Fix string format |

**Example Handling:**
```go
// Validation with recovery
validator := schema.NewValidator(jsonSchema)
err := validator.Validate(data)
if err != nil {
    var valErr *errors.SchemaValidationError
    if errors.As(err, &valErr) {
        // Attempt auto-correction
        for _, issue := range valErr.Issues {
            if issue.Type == "missing_required" {
                // Add default value
                data[issue.Field] = issue.DefaultValue
            } else if issue.Type == "type_mismatch" {
                // Attempt type conversion
                data[issue.Field] = convertType(data[issue.Field], issue.ExpectedType)
            }
        }
        
        // Retry validation
        err = validator.Validate(data)
    }
}
```

### Input Validation Errors (9xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 9001 | `InputTooLong` | Input exceeds max length | Truncate or summarize |
| 9002 | `InputTooShort` | Input below min length | Provide more context |
| 9003 | `InvalidFormat` | Input format incorrect | Reformat input |
| 9004 | `UnsupportedLanguage` | Language not supported | Translate or switch provider |
| 9005 | `InvalidEncoding` | Character encoding issue | Fix encoding |

---

## System Errors

### Configuration Errors (10xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 10001 | `ConfigNotFound` | Configuration file missing | Use defaults or create |
| 10002 | `ConfigInvalid` | Invalid configuration | Fix configuration |
| 10003 | `ConfigVersionMismatch` | Wrong config version | Migrate configuration |
| 10004 | `MissingRequiredConfig` | Required setting missing | Add required setting |
| 10005 | `ConfigAccessDenied` | Can't read config | Check permissions |

### Network Errors (11xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 11001 | `ConnectionTimeout` | Connection timed out | Retry with backoff |
| 11002 | `ConnectionRefused` | Connection refused | Check service status |
| 11003 | `DNSResolutionFailed` | DNS lookup failed | Check network/DNS |
| 11004 | `SSLHandshakeFailed` | SSL/TLS error | Check certificates |
| 11005 | `NetworkUnreachable` | Network unreachable | Check connectivity |

**Example Handling:**
```go
// Network error recovery
func withNetworkRetry(fn func() error) error {
    backoff := []time.Duration{
        1 * time.Second,
        2 * time.Second,
        4 * time.Second,
        8 * time.Second,
    }
    
    var lastErr error
    for i, delay := range backoff {
        err := fn()
        if err == nil {
            return nil
        }
        
        var netErr *errors.NetworkError
        if errors.As(err, &netErr) {
            if netErr.Temporary {
                log.Printf("Attempt %d failed, retrying in %v", i+1, delay)
                time.Sleep(delay)
                lastErr = err
                continue
            }
        }
        
        return err // Non-retryable error
    }
    
    return lastErr
}
```

### Internal Errors (12xxx)

| Code | Name | Description | Recovery Strategy |
|------|------|-------------|-------------------|
| 12001 | `PanicRecovered` | Recovered from panic | Report bug, restart |
| 12002 | `MemoryAllocationFailed` | Out of memory | Reduce memory usage |
| 12003 | `GoroutineLeaked` | Goroutine leak detected | Fix leak, restart |
| 12004 | `DeadlockDetected` | Potential deadlock | Fix code, restart |
| 12005 | `InternalInconsistency` | Internal state corrupted | Reset state |

---

## Error Handling Patterns

### Basic Error Handling

```go
// Simple error check
result, err := agent.Complete(ctx, request)
if err != nil {
    return fmt.Errorf("agent failed: %w", err)
}

// Type assertion
if authErr, ok := err.(*errors.AuthenticationError); ok {
    // Handle authentication error
}

// Using errors.As (recommended)
var rateErr *errors.RateLimitError
if errors.As(err, &rateErr) {
    // Handle rate limit error
}
```

### Comprehensive Error Handler

```go
func HandleError(err error) error {
    if err == nil {
        return nil
    }
    
    // Log error with context
    log.Printf("Error occurred: %v", err)
    
    // Check for specific error types
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        return fmt.Errorf("operation timed out")
        
    case errors.Is(err, context.Canceled):
        return fmt.Errorf("operation cancelled")
    }
    
    // Handle provider errors
    var provErr *errors.ProviderError
    if errors.As(err, &provErr) {
        return handleProviderError(provErr)
    }
    
    // Handle validation errors
    var valErr *errors.ValidationError
    if errors.As(err, &valErr) {
        return handleValidationError(valErr)
    }
    
    // Default error
    return fmt.Errorf("unexpected error: %w", err)
}
```

### Error Recovery Strategies

```go
// Retry with exponential backoff
func RetryWithBackoff(operation func() error, maxAttempts int) error {
    for attempt := 0; attempt < maxAttempts; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        if !isRetryable(err) {
            return err
        }
        
        delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
        time.Sleep(delay)
    }
    
    return fmt.Errorf("max attempts exceeded")
}

// Circuit breaker pattern
type CircuitBreaker struct {
    failures     int
    lastFailTime time.Time
    threshold    int
    timeout      time.Duration
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.isOpen() {
        return fmt.Errorf("circuit breaker is open")
    }
    
    err := fn()
    if err != nil {
        cb.recordFailure()
        return err
    }
    
    cb.reset()
    return nil
}
```

### Error Aggregation

```go
// Collect multiple errors
type ErrorCollector struct {
    errors []error
    mu     sync.Mutex
}

func (ec *ErrorCollector) Add(err error) {
    if err == nil {
        return
    }
    
    ec.mu.Lock()
    defer ec.mu.Unlock()
    ec.errors = append(ec.errors, err)
}

func (ec *ErrorCollector) Error() error {
    ec.mu.Lock()
    defer ec.mu.Unlock()
    
    if len(ec.errors) == 0 {
        return nil
    }
    
    return fmt.Errorf("multiple errors occurred: %v", ec.errors)
}
```

---

## Error Logging and Monitoring

### Structured Error Logging

```go
// Enhanced error with context
type ErrorContext struct {
    Code       string
    Message    string
    Provider   string
    Model      string
    Operation  string
    Timestamp  time.Time
    TraceID    string
    Details    map[string]interface{}
}

func LogError(err error, ctx ErrorContext) {
    log.WithFields(log.Fields{
        "error_code":  ctx.Code,
        "error_msg":   ctx.Message,
        "provider":    ctx.Provider,
        "model":       ctx.Model,
        "operation":   ctx.Operation,
        "timestamp":   ctx.Timestamp,
        "trace_id":    ctx.TraceID,
        "details":     ctx.Details,
    }).Error("Operation failed")
}
```

### Error Metrics

```go
// Prometheus metrics for errors
var (
    errorCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "gollms_errors_total",
            Help: "Total number of errors",
        },
        []string{"type", "code", "provider"},
    )
    
    errorDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "gollms_error_recovery_duration_seconds",
            Help: "Time taken to recover from errors",
        },
        []string{"type", "recovery_method"},
    )
)
```

---

## Best Practices

### 1. Always Check Errors
```go
// Bad
result, _ := agent.Complete(ctx, request)

// Good
result, err := agent.Complete(ctx, request)
if err != nil {
    return fmt.Errorf("completion failed: %w", err)
}
```

### 2. Use Error Wrapping
```go
// Preserve error context
if err != nil {
    return fmt.Errorf("failed to process request %s: %w", requestID, err)
}
```

### 3. Handle Specific Error Types
```go
// Check specific error types first
var authErr *errors.AuthenticationError
if errors.As(err, &authErr) {
    // Handle auth error specifically
    return handleAuthError(authErr)
}

// Then handle general errors
return fmt.Errorf("unexpected error: %w", err)
```

### 4. Implement Graceful Degradation
```go
// Try primary provider, fall back to secondary
result, err := primaryProvider.Complete(ctx, request)
if err != nil {
    log.Printf("Primary provider failed: %v, trying fallback", err)
    result, err = fallbackProvider.Complete(ctx, request)
    if err != nil {
        return nil, fmt.Errorf("all providers failed: %w", err)
    }
}
```

### 5. Log Errors Appropriately
```go
// Log with appropriate level
switch {
case errors.Is(err, context.Canceled):
    log.Debug("Operation cancelled by user")
case isRetryable(err):
    log.Warn("Retryable error occurred", "error", err)
default:
    log.Error("Fatal error occurred", "error", err)
}
```

---

## Troubleshooting Guide

### Common Issues and Solutions

| Issue | Error Codes | Solution |
|-------|-------------|----------|
| "API key invalid" | 1001-1003 | Check environment variables, verify key validity |
| "Rate limit exceeded" | 2001-2005 | Implement backoff, reduce request frequency |
| "Model not found" | 3001 | Check model name spelling, use ListModels |
| "Tool execution failed" | 5001-5005 | Verify tool inputs, check permissions |
| "Validation failed" | 8001-8005 | Review schema, check data format |
| "Connection timeout" | 11001 | Increase timeout, check network |

### Debug Mode

```go
// Enable debug mode for detailed errors
os.Setenv("GOLLMS_DEBUG", "true")
os.Setenv("GOLLMS_LOG_LEVEL", "debug")

// Custom error handler with stack traces
func DebugErrorHandler(err error) {
    if os.Getenv("GOLLMS_DEBUG") == "true" {
        log.Printf("Error: %+v", err) // Prints with stack trace
    }
}
```

---

## Next Steps

- **[Best Practices Checklist](best-practices-checklist.md)** - Production readiness guide
- **[Configuration Reference](configuration-reference.md)** - Error-related configuration
- **[Troubleshooting Guide](/docs/user-guide/advanced/troubleshooting.md)** - Advanced debugging
- **[Provider Comparison](provider-comparison.md)** - Provider-specific errors
- **[Monitoring Guide](/docs/user-guide/advanced/production-deployment.md#monitoring)** - Error monitoring setup