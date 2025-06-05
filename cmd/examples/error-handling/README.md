# Advanced Error Handling Example

This example demonstrates advanced error handling patterns for agents, including retry logic, fallback strategies, circuit breakers, and error recovery techniques.

## Features

- **Exponential Backoff Retry**: Automatically retry failed operations with increasing delays
- **Fallback Chain**: Try multiple agents in sequence until one succeeds
- **Error Tracking**: Use hooks to monitor and log errors
- **Circuit Breaker**: Prevent cascading failures by failing fast
- **Custom Error Types**: Define specific error types for better handling

## Running the Example

```bash
go run main.go
```

## Key Concepts

### 1. Retry with Exponential Backoff

Implements intelligent retry logic for transient failures:

```go
retryConfig := RetryConfig{
    MaxAttempts:   3,
    InitialDelay:  1 * time.Second,
    MaxDelay:      10 * time.Second,
    BackoffFactor: 2.0,
}
```

The retry mechanism:
- Identifies retryable errors (timeouts, rate limits)
- Increases delay between attempts
- Caps maximum delay to prevent excessive waiting

### 2. Fallback Chain

Create a chain of agents that are tried in sequence:

```go
fallbackAgent := NewFallbackAgent("chain", 
    primaryAgent,    // Try this first
    fallback1,       // If primary fails
    fallback2,       // Last resort
)
```

Useful for:
- Multiple provider redundancy
- Gradual degradation of service
- Cost optimization (try cheaper providers first)

### 3. Error Tracking Hooks

Monitor errors across agent execution:

```go
type ErrorHandlingHook struct {
    errors []error
    recoveries int
}
```

Tracks:
- Generation errors
- Tool execution failures
- Recovery attempts
- Error patterns

### 4. Circuit Breaker Pattern

Prevent cascading failures:

```go
circuitBreaker := NewCircuitBreaker(
    3,              // Failure threshold
    5*time.Second,  // Reset timeout
)
```

States:
- **Closed**: Normal operation
- **Open**: Failing fast, not attempting calls
- **Half-Open**: Testing if service recovered

## Error Types

The example defines custom error types for specific handling:

```go
var (
    ErrNetworkTimeout = errors.New("network timeout")
    ErrRateLimited    = errors.New("rate limited")
    ErrInvalidInput   = errors.New("invalid input")
    ErrToolFailure    = errors.New("tool execution failed")
)
```

## Best Practices

1. **Identify Retryable Errors**: Not all errors should trigger retries
2. **Set Reasonable Timeouts**: Balance between patience and responsiveness
3. **Log Errors Appropriately**: Track patterns without flooding logs
4. **Fail Fast When Appropriate**: Don't retry permanent failures
5. **Provide Fallback Options**: Graceful degradation is better than complete failure

## Use Cases

- **API Integration**: Handle rate limits and temporary outages
- **Multi-Provider Systems**: Automatic failover between providers
- **Production Systems**: Maintain service availability
- **Cost Optimization**: Try cheaper options first, fallback to premium
- **Development/Testing**: Simulate and handle various failure modes

## Extensions

You could extend this example to:

- Implement request queuing during outages
- Add metrics collection for error rates
- Create provider health checks
- Implement adaptive retry strategies
- Add request deduplication