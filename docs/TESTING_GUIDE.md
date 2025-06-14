# Go-LLMs Testing Guide

This guide provides comprehensive information on testing with the go-llms testing infrastructure introduced in v0.3.5.9.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Core Components](#core-components)
- [Usage Patterns](#usage-patterns)
- [Best Practices](#best-practices)
- [Performance Considerations](#performance-considerations)
- [Migration Guide](#migration-guide)
- [Troubleshooting](#troubleshooting)

## Overview

The go-llms testing infrastructure provides a comprehensive set of tools for testing LLM-related code:

- **Fixtures**: Pre-configured mock objects for common scenarios
- **Scenario Builder**: Fluent API for complex test setups
- **Matchers**: Flexible assertion capabilities
- **Helpers**: Utilities for context creation and state management
- **Mocks**: Core mock implementations with pattern matching

### Key Benefits

1. **Reduced Boilerplate**: Eliminate repetitive mock setup code
2. **Realistic Behavior**: Pattern-based responses simulate real LLM behavior
3. **Thread Safety**: All components work correctly with concurrent tests
4. **Performance**: Optimized for fast test execution
5. **Maintainability**: Consistent patterns across the codebase

## Quick Start

### Basic Provider Testing

```go
func TestBasicProvider(t *testing.T) {
    // Create a ChatGPT-like provider
    provider := fixtures.ChatGPTMockProvider()
    
    // Test basic generation
    response, err := provider.Generate(context.Background(), "Hello!")
    assert.NoError(t, err)
    assert.Contains(t, response, "Hello")
}
```

### Basic Tool Testing

```go
func TestBasicTool(t *testing.T) {
    // Create a calculator tool
    calc := fixtures.CalculatorMockTool()
    ctx := helpers.CreateTestToolContext()
    
    // Test addition
    result, err := calc.Execute(ctx, map[string]interface{}{
        "operation": "add",
        "a": 5.0,
        "b": 3.0,
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 8.0, result["result"])
}
```

### Basic Agent Testing

```go
func TestBasicAgent(t *testing.T) {
    // Create a research agent
    agent := fixtures.ResearchMockAgent()
    input := fixtures.BasicTestState()
    input.Set("query", "AI trends")
    
    result, err := agent.Run(context.Background(), input)
    assert.NoError(t, err)
    
    taskType, _ := result.Get("task_type")
    assert.Equal(t, "research", taskType)
}
```

## Core Components

### Fixtures

Fixtures provide pre-configured mock objects that behave realistically:

#### Provider Fixtures

```go
// ChatGPT-like responses with conversation patterns
provider := fixtures.ChatGPTMockProvider()

// Claude-like responses with analytical focus
provider := fixtures.ClaudeMockProvider()

// Error simulation for testing failure scenarios
provider := fixtures.ErrorMockProvider("rate_limit")

// Slow response simulation for timeout testing
provider := fixtures.SlowMockProvider(2 * time.Second)

// Streaming response simulation
provider := fixtures.StreamingMockProvider()
```

#### Tool Fixtures

```go
// Arithmetic calculator with all basic operations
calc := fixtures.CalculatorMockTool()

// Web search simulation with realistic results
web := fixtures.WebSearchMockTool()

// File operations with virtual filesystem
file := fixtures.FileMockTool()

// Error tool for testing error handling
error := fixtures.ErrorMockTool(0.3) // 30% error rate
```

#### Agent Fixtures

```go
// Simple agent for basic testing
agent := fixtures.SimpleMockAgent()

// Research agent with query processing
agent := fixtures.ResearchMockAgent()

// Workflow agent for complex processes
agent := fixtures.WorkflowMockAgent()

// Stateful agent with memory
agent := fixtures.StatefulMockAgent()
```

#### State Fixtures

```go
// Empty state for basic testing
state := fixtures.EmptyTestState()

// Basic state with common test data
state := fixtures.BasicTestState()

// Workflow state with execution context
state := fixtures.WorkflowTestState()

// Conversation state with message history
state := fixtures.ConversationTestState()

// State with artifacts (files, documents)
state := fixtures.StateWithArtifacts()

// State with metadata
state := fixtures.StateWithMetadata()

// Error state for testing error conditions
state := fixtures.ErrorTestState()

// Large state for performance testing
state := fixtures.LargeTestState()
```

### Pattern-Based Responses

Fixtures support regex patterns for realistic response matching:

```go
provider := fixtures.ChatGPTMockProvider()

// Add custom pattern responses
provider.WithPatternResponse("(?i).*weather.*", mocks.Response{
    Content: "Today is sunny with 75°F",
    Metadata: map[string]interface{}{
        "location": "test-city",
        "confidence": 0.95,
    },
})

provider.WithPatternResponse("(?i).*programming.*", mocks.Response{
    Content: "Programming is the art of solving problems with code.",
    Metadata: map[string]interface{}{
        "topic": "software-development",
        "complexity": "beginner",
    },
})
```

### Scenario Builder

The scenario builder provides a fluent API for complex test setups:

```go
scenario.NewScenario(t).
    WithMockProvider("chatgpt", map[string]mocks.Response{
        "(?i).*analyze.*": {
            Content: "Analysis complete: Market shows growth potential.",
            Metadata: map[string]interface{}{
                "confidence": 0.92,
                "data_points": 127,
            },
        },
    }).
    WithTool(fixtures.WebSearchMockTool()).
    WithTool(fixtures.CalculatorMockTool()).
    WithAgent(fixtures.WorkflowMockAgent()).
    WithInput("task", "analyze market trends").
    WithInput("sector", "renewable energy").
    ExpectOutput("status", matchers.Equals("completed")).
    ExpectOutput("confidence", matchers.GreaterThan(0.9)).
    ExpectMetadata("execution_time", matchers.IsType[time.Duration]()).
    ExpectNoError().
    Run()
```

### Matchers

Matchers provide flexible assertion capabilities:

```go
// String matchers
matchers.Equals("exact value")
matchers.Contains("substring")
matchers.HasPrefix("prefix")
matchers.HasSuffix("suffix")
matchers.MatchesRegex(`^\d{4}-\d{2}-\d{2}$`)

// Numeric matchers
matchers.GreaterThan(10)
matchers.LessThan(100)
matchers.GreaterThanOrEqual(10)
matchers.LessThanOrEqual(100)

// Type matchers
matchers.IsType[string]()
matchers.IsType[[]int]()
matchers.IsType[map[string]interface{}]()

// Collection matchers
matchers.HasLength(5)
matchers.Contains("item")
matchers.IsEmpty()

// Boolean and nil matchers
matchers.IsTrue()
matchers.IsFalse()
matchers.IsNil()
matchers.IsNotNil()
```

### Helpers

Helpers provide utilities for common testing tasks:

```go
// Context creation
ctx := helpers.CreateTestToolContext()
ctx := helpers.CreateToolContextWithState(map[string]interface{}{
    "mode": "test",
    "debug": true,
})

// Event testing
capture := helpers.NewEventCapture()
// ... run code that emits events
events := capture.GetEvents()

helpers.AssertEvents(t, events).
    HasType("agent.start").
    HasType("tool.execute").
    HasType("agent.complete").
    InOrder()

// Pointer helpers (for optional fields)
stringPtr := helpers.StringPtr("value")
intPtr := helpers.IntPtr(42)
boolPtr := helpers.BoolPtr(true)
```

## Usage Patterns

### Testing Provider Functionality

```go
func TestProviderFunctionality(t *testing.T) {
    provider := fixtures.ChatGPTMockProvider()
    
    // Test different message types
    messages := []domain.Message{
        {Role: "system", Content: "You are a helpful assistant"},
        {Role: "user", Content: "What is Go programming?"},
    }
    
    response, err := provider.GenerateMessage(context.Background(), messages)
    assert.NoError(t, err)
    assert.Contains(t, response.Content, "Go")
    
    // Test streaming
    stream, err := provider.Stream(context.Background(), "Count to 5")
    assert.NoError(t, err)
    
    var tokens []string
    for token := range stream {
        if token.Text != "" {
            tokens = append(tokens, token.Text)
        }
    }
    assert.Greater(t, len(tokens), 3)
}
```

### Testing Tool Execution

```go
func TestToolExecution(t *testing.T) {
    // Test calculator operations
    calc := fixtures.CalculatorMockTool()
    ctx := helpers.CreateTestToolContext()
    
    operations := []struct {
        name     string
        op       string
        a, b     float64
        expected float64
    }{
        {"addition", "add", 5, 3, 8},
        {"subtraction", "subtract", 10, 4, 6},
        {"multiplication", "multiply", 3, 7, 21},
        {"division", "divide", 15, 3, 5},
    }
    
    for _, test := range operations {
        t.Run(test.name, func(t *testing.T) {
            result, err := calc.Execute(ctx, map[string]interface{}{
                "operation": test.op,
                "a": test.a,
                "b": test.b,
            })
            
            assert.NoError(t, err)
            assert.Equal(t, test.expected, result["result"])
        })
    }
}
```

### Testing Agent Workflows

```go
func TestAgentWorkflows(t *testing.T) {
    tests := []struct {
        name       string
        agent      func() *mocks.MockAgent
        input      *domain.State
        assertions func(t *testing.T, result *domain.State, err error)
    }{
        {
            name:  "simple processing",
            agent: fixtures.SimpleMockAgent,
            input: fixtures.BasicTestState(),
            assertions: func(t *testing.T, result *domain.State, err error) {
                assert.NoError(t, err)
                message, _ := result.Get("message")
                assert.Equal(t, "Simple agent response", message)
            },
        },
        {
            name: "research workflow",
            agent: fixtures.ResearchMockAgent,
            input: func() *domain.State {
                state := fixtures.BasicTestState()
                state.Set("query", "machine learning trends")
                return state
            }(),
            assertions: func(t *testing.T, result *domain.State, err error) {
                assert.NoError(t, err)
                taskType, _ := result.Get("task_type")
                assert.Equal(t, "research", taskType)
            },
        },
    }
    
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            agent := test.agent()
            result, err := agent.Run(context.Background(), test.input)
            test.assertions(t, result, err)
        })
    }
}
```

### Testing Error Conditions

```go
func TestErrorConditions(t *testing.T) {
    // Test provider errors
    t.Run("provider rate limiting", func(t *testing.T) {
        provider := fixtures.ErrorMockProvider("rate_limit")
        
        _, err := provider.Generate(context.Background(), "test")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "rate_limit")
    })
    
    // Test timeout scenarios
    t.Run("provider timeout", func(t *testing.T) {
        provider := fixtures.SlowMockProvider(2 * time.Second)
        
        ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
        defer cancel()
        
        _, err := provider.Generate(ctx, "test")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "context deadline exceeded")
    })
    
    // Test tool errors
    t.Run("tool execution error", func(t *testing.T) {
        errorTool := fixtures.ErrorMockTool(1.0) // 100% error rate
        ctx := helpers.CreateTestToolContext()
        
        _, err := errorTool.Execute(ctx, map[string]interface{}{})
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "simulated error")
    })
}
```

### Testing Concurrent Access

```go
func TestConcurrentAccess(t *testing.T) {
    provider := fixtures.ChatGPTMockProvider()
    
    // Test thread safety
    const numGoroutines = 10
    const requestsPerGoroutine = 10
    
    var wg sync.WaitGroup
    wg.Add(numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        go func(routineID int) {
            defer wg.Done()
            
            for j := 0; j < requestsPerGoroutine; j++ {
                prompt := fmt.Sprintf("Request %d-%d", routineID, j)
                response, err := provider.Generate(context.Background(), prompt)
                assert.NoError(t, err)
                assert.NotEmpty(t, response)
            }
        }(i)
    }
    
    wg.Wait()
    
    // Verify all requests were recorded
    history := provider.GetCallHistory()
    assert.Len(t, history, numGoroutines*requestsPerGoroutine)
}
```

## Best Practices

### 1. Use Appropriate Fixtures

Choose the most specific fixture for your test case:

```go
// ✅ Good: Use specific fixture
provider := fixtures.ChatGPTMockProvider()

// ❌ Avoid: Generic mock requiring manual setup
provider := mocks.NewMockProvider("generic")
```

### 2. Leverage Pattern Matching

Use patterns for realistic behavior simulation:

```go
// ✅ Good: Pattern-based responses
provider.WithPatternResponse("(?i).*weather.*", mocks.Response{
    Content: "Weather information response",
})

// ❌ Avoid: Single static response
provider.WithDefaultResponse("Static response")
```

### 3. Use Scenario Builder for Complex Tests

For multi-component tests, use the scenario builder:

```go
// ✅ Good: Scenario builder for complex setup
scenario.NewScenario(t).
    WithMockProvider("chatgpt", responses).
    WithTool(tool).
    WithAgent(agent).
    Run()

// ❌ Avoid: Manual setup for complex scenarios
// ... lots of manual setup code
```

### 4. Apply Appropriate Matchers

Use the most specific matcher for your assertion:

```go
// ✅ Good: Specific matcher
assert.True(t, matchers.HasPrefix("Hello").Match(response))

// ❌ Avoid: Generic assertion that could be more specific
assert.True(t, strings.HasPrefix(response, "Hello"))
```

### 5. Test Error Conditions

Always test error scenarios:

```go
func TestWithErrors(t *testing.T) {
    // Test normal case
    t.Run("success", func(t *testing.T) {
        // ... success test
    })
    
    // Test error cases
    t.Run("rate_limit_error", func(t *testing.T) {
        provider := fixtures.ErrorMockProvider("rate_limit")
        // ... error test
    })
    
    t.Run("timeout_error", func(t *testing.T) {
        provider := fixtures.SlowMockProvider(2 * time.Second)
        // ... timeout test
    })
}
```

### 6. Use Table-Driven Tests

For testing multiple similar scenarios:

```go
func TestMultipleScenarios(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"greeting", "hello", "Hello"},
        {"question", "what is", "What is"},
        {"command", "please do", "I'll help"},
    }
    
    provider := fixtures.ChatGPTMockProvider()
    
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            response, err := provider.Generate(context.Background(), test.input)
            assert.NoError(t, err)
            assert.Contains(t, response, test.expected)
        })
    }
}
```

### 7. Clean Up Resources

Ensure proper cleanup in tests:

```go
func TestWithCleanup(t *testing.T) {
    provider := fixtures.ChatGPTMockProvider()
    
    // Use t.Cleanup for automatic cleanup
    t.Cleanup(func() {
        provider.Reset() // Reset call history if needed
    })
    
    // ... test code
}
```

## Performance Considerations

### Fixture Reuse

Fixtures are designed for reuse within tests:

```go
func TestFixtureReuse(t *testing.T) {
    // Create once, use multiple times
    provider := fixtures.ChatGPTMockProvider()
    
    t.Run("test1", func(t *testing.T) {
        response, err := provider.Generate(context.Background(), "hello")
        assert.NoError(t, err)
    })
    
    t.Run("test2", func(t *testing.T) {
        response, err := provider.Generate(context.Background(), "goodbye")
        assert.NoError(t, err)
    })
}
```

### Parallel Testing

Enable parallel testing for better performance:

```go
func TestParallel(t *testing.T) {
    t.Parallel() // Enable parallel execution
    
    provider := fixtures.ChatGPTMockProvider()
    
    t.Run("subtest1", func(t *testing.T) {
        t.Parallel()
        // ... test code
    })
    
    t.Run("subtest2", func(t *testing.T) {
        t.Parallel()
        // ... test code
    })
}
```

### Memory Usage

For large test suites, consider memory usage:

```go
func TestMemoryEfficient(t *testing.T) {
    // Use smaller state fixtures when possible
    state := fixtures.BasicTestState() // Instead of LargeTestState()
    
    // Reset call history if not needed
    provider := fixtures.ChatGPTMockProvider()
    defer func() {
        history := provider.GetCallHistory()
        if len(history) > 100 {
            provider.Reset()
        }
    }()
}
```

## Migration Guide

### From Legacy MockProvider

```go
// OLD: Manual mock setup
func TestOldWay(t *testing.T) {
    mock := provider.NewMockProvider()
    mock.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
        if strings.Contains(prompt, "hello") {
            return "Hello response", nil
        }
        return "Default response", nil
    })
    
    response, err := mock.Generate(context.Background(), "hello world")
    assert.NoError(t, err)
    assert.Contains(t, response, "Hello")
}

// NEW: Using fixtures
func TestNewWay(t *testing.T) {
    provider := fixtures.ChatGPTMockProvider()
    provider.WithPatternResponse("(?i).*hello.*", mocks.Response{
        Content: "Hello response",
    })
    
    response, err := provider.Generate(context.Background(), "hello world")
    assert.NoError(t, err)
    assert.Contains(t, response, "Hello")
}
```

### From Manual Context Creation

```go
// OLD: Manual context creation
func TestOldContext(t *testing.T) {
    mockAgent := &MockAgent{
        BaseAgentImpl: core.NewBaseAgent("test", "desc", domain.AgentTypeCustom),
    }
    ctx := domain.NewToolContext(
        context.Background(),
        domain.NewStateReader(domain.NewState()),
        mockAgent,
        "test-run",
    )
    
    // ... test code
}

// NEW: Using helpers
func TestNewContext(t *testing.T) {
    ctx := helpers.CreateTestToolContext()
    
    // Or with custom state
    ctx := helpers.CreateToolContextWithState(map[string]interface{}{
        "key": "value",
    })
    
    // ... test code
}
```

## Troubleshooting

### Common Issues

#### Pattern Not Matching

```go
// ❌ Problem: Pattern doesn't match
provider.WithPatternResponse("hello", mocks.Response{
    Content: "Hello response",
})

// Input: "Hello world!" (capital H)
// Result: No match, uses default response

// ✅ Solution: Use case-insensitive regex
provider.WithPatternResponse("(?i).*hello.*", mocks.Response{
    Content: "Hello response",
})
```

#### Context Creation Errors

```go
// ❌ Problem: Wrong context creation
ctx := domain.NewToolContext(context.Background(), nil, nil, "")

// ✅ Solution: Use helper
ctx := helpers.CreateTestToolContext()
```

#### Thread Safety Issues

```go
// ❌ Problem: Shared mutable state
var sharedProvider = fixtures.ChatGPTMockProvider()

func TestConcurrent1(t *testing.T) {
    sharedProvider.WithPatternResponse("test1", response1)
    // ... test
}

func TestConcurrent2(t *testing.T) {
    sharedProvider.WithPatternResponse("test2", response2)
    // ... test - may interfere with TestConcurrent1
}

// ✅ Solution: Create separate instances
func TestConcurrent1(t *testing.T) {
    provider := fixtures.ChatGPTMockProvider()
    provider.WithPatternResponse("test1", response1)
    // ... test
}

func TestConcurrent2(t *testing.T) {
    provider := fixtures.ChatGPTMockProvider()
    provider.WithPatternResponse("test2", response2)
    // ... test
}
```

### Debugging Tips

#### Enable Debug Logging

```go
func TestWithDebug(t *testing.T) {
    // Enable debug logging for fixtures
    provider := fixtures.ChatGPTMockProvider()
    provider.EnableDebugLogging(t)
    
    // Now all provider interactions will be logged
    response, err := provider.Generate(context.Background(), "test")
    // Output: [DEBUG] MockProvider: Generated response for pattern "test"
}
```

#### Inspect Call History

```go
func TestWithHistory(t *testing.T) {
    provider := fixtures.ChatGPTMockProvider()
    
    // Make some calls
    provider.Generate(context.Background(), "test1")
    provider.Generate(context.Background(), "test2")
    
    // Inspect history for debugging
    history := provider.GetCallHistory()
    for i, call := range history {
        t.Logf("Call %d: %s -> %s", i, call.Input, call.Output)
    }
}
```

#### Check Pattern Matching

```go
func TestPatternMatching(t *testing.T) {
    provider := fixtures.ChatGPTMockProvider()
    
    // Add a catch-all pattern for debugging
    provider.WithPatternResponse(".*", mocks.Response{
        Content: "DEBUG: Received input: {{.Input}}",
    })
    
    response, err := provider.Generate(context.Background(), "test input")
    t.Logf("Response: %s", response)
    // Output: "DEBUG: Received input: test input"
}
```

### Performance Debugging

#### Measure Test Performance

```go
func TestPerformance(t *testing.T) {
    start := time.Now()
    
    provider := fixtures.ChatGPTMockProvider()
    
    for i := 0; i < 1000; i++ {
        provider.Generate(context.Background(), fmt.Sprintf("test %d", i))
    }
    
    duration := time.Since(start)
    t.Logf("1000 requests took: %v (avg: %v)", duration, duration/1000)
    
    if duration > time.Second {
        t.Errorf("Performance issue: took %v, expected < 1s", duration)
    }
}
```

## Further Reading

- [TESTING_MIGRATION_GUIDE.md](./TESTING_MIGRATION_GUIDE.md) - Detailed migration instructions
- [pkg/testutils/doc.go](../pkg/testutils/doc.go) - Package documentation
- [docs/examples/testing_examples_test.go](./examples/testing_examples_test.go) - Comprehensive examples
- [Makefile](../Makefile) - Available test commands