# Testing Framework and Strategy

> **[Documentation Home](/REFERENCE.md) / [Technical Documentation](/docs/technical/) / Testing Framework**

This document outlines the testing approach implemented in Go-LLMs, focusing on comprehensive test coverage, error condition handling, and stress testing for high-load scenarios.

## Table of Contents

1. [Introduction](#introduction)
2. [Error Condition Test Suite](#error-condition-test-suite)
3. [Stress Testing](#stress-testing)
4. [Running Tests](#running-tests)
5. [Related Documentation](#related-documentation)

## Introduction

The Go-LLMs library implements a comprehensive testing strategy with several layers:

- **Unit tests**: Verify individual component behavior
- **Integration tests**: Test the interaction between multiple components
- **Error condition tests**: Validate error handling and recovery mechanisms
- **Benchmarks**: Measure performance characteristics and identify bottlenecks (see [Benchmarking Framework](benchmarks.md))
- **Stress tests**: Evaluate behavior under high-load scenarios

This layered approach ensures that the library is robust, performant, and reliable in various usage scenarios. This document focuses specifically on error condition testing and stress testing aspects of the framework.

## Error Condition Test Suite

The error condition test suite systematically tests how the library handles various error scenarios, focusing on:

1. Provider error conditions
2. Schema validation error conditions
3. Agent error conditions

### Provider Error Handling

Provider error tests validate that the library correctly handles various API error scenarios:

```go
func TestProviderErrors(t *testing.T) {
    t.Run("MockErrorServer", func(t *testing.T) {
        mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Error simulation based on URL path
            if strings.Contains(r.URL.Path, "auth-error") {
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"error":{"message":"Invalid API key"}}`))
            } else if strings.Contains(r.URL.Path, "rate-limit") {
                w.WriteHeader(http.StatusTooManyRequests)
                w.Write([]byte(`{"error":{"message":"Rate limit exceeded"}}`))
            } else if strings.Contains(r.URL.Path, "context-length") {
                w.WriteHeader(http.StatusBadRequest)
                w.Write([]byte(`{"error":{"message":"Context length exceeded"}}`))
            }
            // Additional error types handled...
        }))
        defer mockServer.Close()

        // Test providers with different error conditions
        testErrorConditions(t, mockServer.URL, "auth-error", domain.ErrAuthenticationFailure, "OpenAI")
        testErrorConditions(t, mockServer.URL, "rate-limit", domain.ErrRateLimitExceeded, "Anthropic")
        testErrorConditions(t, mockServer.URL, "context-length", domain.ErrContextLengthExceeded, "Gemini")
        // Additional error types tested...
    })

    // Additional error scenarios tested
    t.Run("NetworkFailure", func(t *testing.T) { /* Test network failures */ })
    t.Run("RetryMechanism", func(t *testing.T) { /* Test retry behavior */ })
}
```

This approach tests:

- Authentication errors
- Rate limiting
- Context length errors
- Content filtering
- Server errors
- Malformed responses
- Network connectivity issues
- Timeout handling
- Retry mechanisms

### Schema Validation Errors

Schema validation error tests ensure that the library properly validates input data and provides helpful error messages:

```go
func TestSchemaValidationErrors(t *testing.T) {
    t.Run("TypeValidationErrors", func(t *testing.T) {
        schema := `{
            "type": "object",
            "properties": {
                "name": {"type": "string"},
                "age": {"type": "integer"},
                "active": {"type": "boolean"}
            }
        }`
        
        // Test invalid types
        data := `{
            "name": 123,
            "age": "twenty",
            "active": "yes"
        }`
        
        result, err := validation.ValidateJSON(data, schema)
        
        // Verify appropriate error types and messages
        require.Error(t, err)
        require.False(t, result.Valid)
        require.Contains(t, err.Error(), "name: expected string, got number")
        require.Contains(t, err.Error(), "age: expected integer, got string")
        require.Contains(t, err.Error(), "active: expected boolean, got string")
    })

    // Additional validation error types tested
    t.Run("ConstraintValidationErrors", func(t *testing.T) { /* ... */ })
    t.Run("RequiredFieldsValidation", func(t *testing.T) { /* ... */ })
    t.Run("NestedObjectValidation", func(t *testing.T) { /* ... */ })
    t.Run("ArrayItemValidation", func(t *testing.T) { /* ... */ })
    // More validation scenarios...
}
```

These tests cover:

- Type validation errors
- Constraint validation (min/max, pattern, format)
- Required field validation
- Nested object validation
- Array item validation
- Complex schema validation scenarios

### Agent Error Conditions

Agent error tests validate how the agent system handles tool errors, invalid schemas, timeouts, and more:

```go
func TestAgentErrors(t *testing.T) {
    t.Run("FailingProvider", func(t *testing.T) {
        mockProvider := provider.NewMockProvider()
        mockProvider.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
            return "", errors.New("simulated provider error")
        })
        
        agent := workflow.NewBaseAgent(mockProvider, nil)
        _, err := agent.Run(context.Background(), "test prompt")
        
        require.Error(t, err)
        require.Contains(t, err.Error(), "simulated provider error")
    })

    // Additional agent error scenarios
    t.Run("FailingTool", func(t *testing.T) { /* ... */ })
    t.Run("InvalidToolName", func(t *testing.T) { /* ... */ })
    t.Run("InvalidToolParams", func(t *testing.T) { /* ... */ })
    t.Run("InvalidSchema", func(t *testing.T) { /* ... */ })
    t.Run("Timeout", func(t *testing.T) { /* ... */ })
    // More error scenarios...
}
```

These tests verify:

- Provider failure handling
- Tool execution errors
- Invalid tool name handling
- Parameter validation
- Schema validation errors
- Timeout handling
- Error propagation through hooks

```

## Stress Testing

Stress tests evaluate the library's behavior under high-concurrency and load conditions, ensuring stability and reliability in production environments.

### Provider Stress Tests

```go
func TestProviderConcurrentRequests(t *testing.T) {
    // Skip in short test mode
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }
    
    // Create providers to test
    providers := []struct {
        name     string
        provider domain.Provider
    }{
        {"OpenAI", mockOpenAIProvider()},
        {"Anthropic", mockAnthropicProvider()},
        {"Gemini", mockGeminiProvider()},
        {"Multi", mockMultiProvider()},
    }
    
    // Test different concurrency levels
    concurrencyLevels := []int{10, 50, 100, 250, 500}
    
    // Run tests for each provider and concurrency level
    for _, p := range providers {
        for _, concurrency := range concurrencyLevels {
            t.Run(fmt.Sprintf("%s_Concurrency_%d", p.name, concurrency), func(t *testing.T) {
                var (
                    wg            sync.WaitGroup
                    successful    int32
                    failed        int32
                    totalLatencyMs int64
                )
                
                // Create a semaphore to limit concurrent goroutines
                sem := make(chan struct{}, concurrency)
                
                // Track peak goroutine count
                initialGoroutines := runtime.NumGoroutine()
                
                // Launch concurrent requests
                for i := 0; i < concurrency*2; i++ {
                    wg.Add(1)
                    sem <- struct{}{} // Acquire semaphore
                    go func(id int) {
                        defer func() {
                            <-sem // Release semaphore
                            wg.Done()
                        }()
                        
                        // Select a prompt randomly
                        prompt := selectRandomPrompt()
                        
                        // Measure request time
                        requestStart := time.Now()
                        _, err := p.provider.Generate(context.Background(), prompt)
                        latencyMs := time.Since(requestStart).Milliseconds()
                        
                        // Update metrics atomically
                        atomic.AddInt64(&totalLatencyMs, latencyMs)
                        
                        if err != nil {
                            atomic.AddInt32(&failed, 1)
                        } else {
                            atomic.AddInt32(&successful, 1)
                        }
                    }(i)
                }
                
                // Wait for all requests to complete
                wg.Wait()
                
                // Record results with comprehensive metrics
                t.Logf("Results for %s at concurrency %d:", p.name, concurrency)
                t.Logf("  Success rate: %.2f%% (%d/%d)", 
                    float64(successful)/float64(successful+failed)*100, successful, successful+failed)
                t.Logf("  Average latency: %.2f ms", float64(totalLatencyMs)/float64(successful+failed))
                t.Logf("  Total duration: %v", totalDuration)
                t.Logf("  Goroutines: %d initial, %d peak", initialGoroutines, peakGoroutines)
            })
        }
    }
}
```

### Agent Workflow Stress Tests

```go
func TestAgentConcurrentRequests(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }
    
    // Test different agent types
    agents := []struct {
        name  string
        agent domain.Agent
    }{
        {"BaseAgent", createBaseAgent()},
        {"CachedAgent", createCachedAgent()},
        {"MultiAgent", createMultiAgent()},
    }
    
    // Test different concurrency levels
    concurrencyLevels := []int{5, 20, 50, 100}
    
    // Run tests for each agent and concurrency level
    for _, a := range agents {
        for _, concurrency := range concurrencyLevels {
            t.Run(fmt.Sprintf("%s_Concurrency_%d", a.name, concurrency), func(t *testing.T) {
                // Create thread-safe tool counter
                toolCounter := &safeToolCounter{}
                
                // Add tool counters to agent
                addToolCounterHook(a.agent, toolCounter)
                
                // Create a semaphore to limit concurrent goroutines
                sem := make(chan struct{}, concurrency)
                
                // Run concurrency test with comprehensive metrics
                // Similar pattern to provider stress test...
                
                // Report specific agent metrics
                t.Logf("  Average tool invocations per request: %.2f", 
                    float64(toolCounter.Count())/float64(successful+failed))
            })
        }
    }
}
```

### Memory Pool Stress Tests

```go
func TestResponsePoolStress(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }
    
    // Test different content sizes
    contentSizes := []struct {
        name        string
        size        string
        contentFunc func() string
    }{
        {"Tiny", "small", generateTinyContent},
        {"Small", "small", generateSmallContent},
        {"Medium", "medium", generateMediumContent},
        {"Large", "large", generateLargeContent},
        {"XLarge", "large", generateXLargeContent},
    }
    
    // Test different concurrency levels
    concurrencyLevels := []int{10, 100, 1000}
    
    // Response pool
    pool := domain.NewResponsePool()
    
    // Run stress tests
    for _, size := range contentSizes {
        for _, concurrency := range concurrencyLevels {
            t.Run(fmt.Sprintf("ResponsePool_Concurrency_%d_Size_%s", concurrency, size.name), func(t *testing.T) {
                // Run pool stress test with metrics for acquisition time, release time, throughput
                // Similar pattern to other stress tests...
                
                // Additional metrics specific to pools
                t.Logf("  Total processed: %.2f MB", float64(totalBytesProcessed)/1024/1024)
                t.Logf("  Throughput: %.2f operations/sec", float64(operationsCompleted)/totalDuration.Seconds())
            })
        }
    }
}
```

### Structured Output Processor Stress Tests

```go
func TestStructuredProcessorConcurrentRequests(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }
    
    // Test different schema complexities
    schemas := []struct {
        name        string
        description string
        schema      string
        mockFunc    func() string
    }{
        {"SmallSchema", "small schema, low complexity", smallSchema, mockSmallResponse},
        {"MediumSchema", "medium schema, medium complexity", mediumSchema, mockMediumResponse},
        {"LargeSchema", "large schema, high complexity", largeSchema, mockLargeResponse},
    }
    
    // Test different concurrency levels
    concurrencyLevels := []int{10, 50, 100, 200}
    
    // Run stress tests for each schema and concurrency level
    for _, s := range schemas {
        for _, concurrency := range concurrencyLevels {
            t.Run(fmt.Sprintf("%s_Concurrency_%d", s.name, concurrency), func(t *testing.T) {
                // Run stress test with metrics
                // Similar pattern to other stress tests...
                
                // Structured processor specific metrics
                t.Logf("  Validation error rate: %.2f%% (%d/%d)", 
                    float64(validationErrors)/float64(total)*100, validationErrors, total)
                t.Logf("  Average LLM latency: %.2f ms (%.2f%%)", 
                    float64(llmLatencyMs)/float64(total), percentLLM)
                t.Logf("  Average processing latency: %.2f ms (%.2f%%)", 
                    float64(processingLatencyMs)/float64(total), percentProcessing)
            })
        }
    }
}
```

## Running Tests

The Go-LLMs library provides comprehensive Makefile targets for running different test suites:

```bash
# Run all tests (excluding integration, multi-provider, and stress tests)
make test

# Run all tests including integration, multi-provider, and stress tests
make test-all

# Run specific test suites
make test-integration      # Run integration tests
make test-multi-provider   # Run multi-provider tests
make test-stress           # Run all stress tests

# Run specific stress test categories
make test-stress-provider      # Run provider stress tests
make test-stress-agent         # Run agent workflow stress tests
make test-stress-structured    # Run structured output processor stress tests
make test-stress-pool          # Run memory pool stress tests
```

## Related Documentation

For more detailed information on various aspects of testing and performance:

- [Benchmarking Framework](benchmarks.md) - Detailed overview of performance benchmarks
- [Performance Optimization](performance.md) - Comprehensive overview of performance optimization strategies
- [Sync.Pool Implementation](sync-pool.md) - Detailed guide on sync.Pool usage for memory optimization
- [Concurrency Patterns](concurrency.md) - Documentation of thread safety and concurrent execution patterns