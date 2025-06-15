# Testing Framework and Infrastructure

> **[Documentation Home](/docs/README.md) / [Technical Documentation](/docs/technical/README.md) / Testing Framework**

This document provides comprehensive information on testing with the go-llms testing infrastructure, covering the framework, patterns, and best practices for effective testing.

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Core Components](#core-components)
   - [Fixtures](#fixtures)
   - [Pattern-Based Responses](#pattern-based-responses)
   - [Scenario Builder](#scenario-builder)
   - [Matchers](#matchers)
   - [Helpers](#helpers)
4. [Usage Patterns](#usage-patterns)
   - [Testing Provider Functionality](#testing-provider-functionality)
   - [Testing Tool Execution](#testing-tool-execution)
   - [Testing Agent Workflows](#testing-agent-workflows)
   - [Testing Error Conditions](#testing-error-conditions)
   - [Testing Concurrent Access](#testing-concurrent-access)
5. [Error Condition Test Suite](#error-condition-test-suite)
6. [Agent Testing Considerations](#agent-testing-considerations)
7. [Schema Validation Testing](#schema-validation-testing)
8. [Stress Testing](#stress-testing)
9. [Best Practices](#best-practices)
10. [Performance Considerations](#performance-considerations)
11. [Migration Guide](#migration-guide)
12. [Running Tests](#running-tests)
13. [Troubleshooting](#troubleshooting)
14. [Related Documentation](#related-documentation)

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

### Recent Changes (June 15, 2025)

- **47 files migrated** to use centralized test infrastructure
- **37+ new fixtures created**: 14 tools, 12 providers, 8 agents, 3 advanced
- **~200 lines net reduction** with vastly improved maintainability
- **Comprehensive fixture library** covering all major testing scenarios

See [TESTING_INFRA_CHANGES.md](../../TESTING_INFRA_CHANGES.md) for complete details.

## Quick Start

### Basic Provider Testing

```go
func TestBasicProvider(t *testing.T) {
    // Create an OpenAI-like provider
    provider := fixtures.OpenAIMockProvider()
    
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
// Basic providers
provider := fixtures.BasicMockProvider()                    // Simple responses
provider := fixtures.BasicMockProviderWithContent("Hello")  // Fixed content

// Provider-specific fixtures
provider := fixtures.OpenAIMockProvider()      // OpenAI-like responses
provider := fixtures.AnthropicMockProvider()   // Claude-like responses
provider := fixtures.GeminiMockProvider()      // Gemini-like responses

// Streaming providers
provider := fixtures.RealisticStreamingProvider()  // Variable delays
provider := fixtures.FastStreamingProvider()       // Minimal latency

// Error scenarios
provider := fixtures.RateLimitErrorProvider()     // Rate limit errors
provider := fixtures.AuthErrorProvider()          // Authentication errors
provider := fixtures.NetworkErrorProvider()       // Network failures
provider := fixtures.IntermittentErrorProvider()  // Random errors

// Configuration-specific
provider := fixtures.ConfiguredOpenAIProvider()     // With OpenAI config
provider := fixtures.ConfiguredAnthropicProvider()  // With Anthropic config
```

#### Tool Fixtures

```go
// File operation tools
readTool := fixtures.FileReadMockTool()      // Read files
writeTool := fixtures.FileWriteMockTool()    // Write files
listTool := fixtures.FileListMockTool()      // List directory
deleteTool := fixtures.FileDeleteMockTool()  // Delete files
moveTool := fixtures.FileMoveMockTool()      // Move/rename files

// Web tools
scrapeTool := fixtures.WebScrapeMockTool()        // Web scraping
fetchTool := fixtures.WebFetchMockTool()          // URL fetching
httpTool := fixtures.HTTPRequestMockTool()        // HTTP requests
searchTool := fixtures.WebSearchMockTool()        // Web search

// Data processing tools
jsonTool := fixtures.JSONProcessMockTool()    // JSON manipulation
csvTool := fixtures.CSVProcessMockTool()      // CSV processing
textTool := fixtures.TextProcessMockTool()    // Text analysis

// Calculator tool (from original fixtures)
calc := fixtures.CalculatorMockTool()         // Basic arithmetic
```

#### Agent Fixtures

```go
// Basic agents
agent := fixtures.SimpleMockAgent()      // Basic responses
agent := fixtures.ResearchMockAgent()    // Research workflows
agent := fixtures.WorkflowMockAgent()    // Multi-step workflows

// Stateful agents
agent := fixtures.StatefulMockAgent()                        // Maintains state
agent := fixtures.StateBuilderMockAgent("name", mods)       // Builds state
agent := fixtures.SharedDataBuilderMockAgent("n", "k", "v") // Shared data

// Specialized agents
agent := fixtures.TrackingMockAgent("name", delay)             // Execution tracking
agent := fixtures.SpecialistMockAgent("name", "specialty", t)  // Domain expert
agent := fixtures.CoordinatorMockAgent("name")                 // Multi-agent coord

// Error handling agents
agent := fixtures.ErrorSimulationMockAgent("name", "type", n)  // Error scenarios
agent := fixtures.TimeoutMockAgent("name", timeout)            // Timeout testing
agent := fixtures.QualityRefinementMockAgent("n", q, rate)     // Iterative improve
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
provider := fixtures.OpenAIMockProvider()

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
// Pre-built scenario templates
scenario := fixtures.SimpleScenarioTemplate(t)
scenario := fixtures.ResearchScenarioTemplate(t)
scenario := fixtures.CalculationScenarioTemplate(t)
scenario := fixtures.FileProcessingScenarioTemplate(t)
scenario := fixtures.ErrorHandlingScenarioTemplate(t)
scenario := fixtures.StreamingScenarioTemplate(t)
scenario := fixtures.MultiToolScenarioTemplate(t)
scenario := fixtures.ConversationScenarioTemplate(t)

// Custom scenario with fluent API
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
    ExpectToolCall("web_search", matchers.HasField("query", matchers.Contains("renewable"))).
    ExpectEvent("agent.complete", matchers.HasField("status", matchers.Equals("success"))).
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

// Composite matchers
matchers.AllOf(matchers.HasPrefix("test"), matchers.Contains("data"))
matchers.AnyOf(matchers.Equals("success"), matchers.Equals("completed"))
matchers.Not(matchers.IsEmpty())

// Field matchers
matchers.HasField("status", matchers.Equals("success"))
matchers.MatchesJSON(`{"status": "success", "count": 42}`)
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

// Event capture with filters
capture := helpers.NewEventCapture()
capture.WithFilter(func(event domain.Event) bool {
    return event.Type == "agent.complete"
})

// Event assertions
capture.AssertEventEmitted(t, "agent.start", matchers.HasField("agentID", matchers.Equals("test-agent")))
capture.AssertEventCount(t, "tool.execute", 3)
capture.AssertNoEvents(t, "agent.error")

// Pointer helpers (for optional fields)
stringPtr := helpers.StringPtr("value")
intPtr := helpers.IntPtr(42)
boolPtr := helpers.BoolPtr(true)
```

### Mock Registry

The testing infrastructure includes a centralized mock registry for managing mock instances:

```go
// Register and retrieve mocks by name
registry := mocks.NewMockRegistry()
registry.RegisterProvider("openai", provider)
registry.RegisterTool("calculator", tool)
registry.RegisterAgent("research", agent)

// Retrieve registered mocks
provider := registry.GetProvider("openai")
tool := registry.GetTool("calculator")
agent := registry.GetAgent("research")

// List all registered mocks
providers := registry.ListProviders()
tools := registry.ListTools()
agents := registry.ListAgents()
```

## Usage Patterns

### Testing Provider Functionality

```go
func TestProviderFunctionality(t *testing.T) {
    provider := fixtures.OpenAIMockProvider()
    
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
        provider := fixtures.RateLimitErrorProvider()
        
        _, err := provider.Generate(context.Background(), "test")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "rate limit")
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
    provider := fixtures.OpenAIMockProvider()
    
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

## Error Condition Test Suite

The error condition test suite systematically tests how the library handles various error scenarios:

### Provider Error Handling

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
            }
            // Additional error types...
        }))
        defer mockServer.Close()

        // Test providers with different error conditions
        testErrorConditions(t, mockServer.URL, "auth-error", domain.ErrAuthenticationFailure)
        testErrorConditions(t, mockServer.URL, "rate-limit", domain.ErrRateLimitExceeded)
    })
}
```

### Schema Validation Errors

```go
func TestSchemaValidationErrors(t *testing.T) {
    t.Run("TypeValidationErrors", func(t *testing.T) {
        schema := &domain.Schema{
            Type: "object",
            Properties: map[string]domain.Property{
                "name": {Type: "string"},
                "age": {Type: "integer"},
            },
            Required: []string{"name", "age"},
        }
        
        // Test with wrong types
        wrongTypesJSON := `{
            "name": 123,
            "age": "twenty"
        }`
        
        result, err := validator.Validate(schema, wrongTypesJSON)
        require.Error(t, err)
        require.False(t, result.Valid)
    })
}
```

## Agent Testing Considerations

### Mock Testing Limitations

When testing agent workflows with mock LLM providers, there are several limitations to be aware of:

1. **Direct Response Return**: Mock providers may return tool call JSON directly instead of executing tools
2. **Recursive Tool Calls**: Testing recursive depth limits requires special handling
3. **Sequential Tool Calls**: Testing tool call sequences requires careful setup

### Effective Testing Patterns

#### Direct Tool Extraction Testing

```go
func TestExtractToolCall(t *testing.T) {
    mockProvider := provider.NewMockProvider()
    agent := workflow.NewAgent(mockProvider)

    testJSON := `{
      "tool": "test_tool",
      "params": {
        "key": "value"
      }
    }`

    toolName, params, shouldCall := agent.ExtractToolCall(testJSON)
    assert.Equal(t, "test_tool", toolName)
    assert.True(t, shouldCall)
}
```

#### Testing Tool Functionality with Error Conditions

```go
func TestRecursionDepthLimit(t *testing.T) {
    recursionCount := 0
    maxRecursion := 5

    // Create a tool that errors at max depth
    recursiveErrorTool := tools.NewTool(
        "recursive_error_tool",
        "A tool that tracks calls and errors at max depth",
        func(params map[string]interface{}) (interface{}, error) {
            recursionCount++
            if recursionCount >= maxRecursion {
                return nil, fmt.Errorf("maximum recursion depth (%d) exceeded", maxRecursion)
            }
            return fmt.Sprintf("Success at depth %d", recursionCount), nil
        },
        &sdomain.Schema{},
    )

    // Configure mock provider to surface the error
    mockProvider := provider.NewMockProvider()
    // ... provider configuration

    agent := workflow.NewAgent(mockProvider)
    agent.AddTool(recursiveErrorTool)

    _, err := agent.Run(context.Background(), "Test recursive tool error")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "maximum recursion depth")
}
```

## Schema Validation Testing

The schema validation system provides comprehensive validation features:

### Implementation Status

#### Core Validation Features (Fully Implemented)
- ✅ Type validation (string, number, integer, boolean, object, array)
- ✅ Constraint validation (min/max length, min/max items, pattern, enum, etc.)
- ✅ Required fields validation
- ✅ Nested object validation
- ✅ Array item validation
- ✅ Format validation (email, uri, hostname, ipv4, uuid, etc.)
- ✅ Type coercion

#### Conditional Validation Features (Partially Implemented)
- 🔄 If/Then/Else conditional validation
- 🔄 AllOf validation
- 🔄 AnyOf validation
- 🔄 OneOf validation
- 🔄 Not validation

### Validation Test Structure

```go
func TestSchemaValidationErrors(t *testing.T) {
    t.Run("TypeValidationErrors", func(t *testing.T) {
        schema := &domain.Schema{
            Type: "object",
            Properties: map[string]domain.Property{
                "string_prop": {Type: "string"},
                "number_prop": {Type: "number"},
            },
            Required: []string{"string_prop", "number_prop"},
        }

        wrongTypesJSON := `{
            "string_prop": 123,
            "number_prop": "not a number"
        }`

        result, err := validator.Validate(schema, wrongTypesJSON)
        require.Error(t, err)
        require.False(t, result.Valid)
    })
}
```

## Stress Testing

Stress tests evaluate the library's behavior under high-concurrency and load conditions:

### Provider Stress Tests

```go
func TestProviderConcurrentRequests(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }
    
    providers := []struct {
        name     string
        provider domain.Provider
    }{
        {"OpenAI", mockOpenAIProvider()},
        {"Anthropic", mockAnthropicProvider()},
        {"Gemini", mockGeminiProvider()},
    }
    
    concurrencyLevels := []int{10, 50, 100, 250, 500}
    
    for _, p := range providers {
        for _, concurrency := range concurrencyLevels {
            t.Run(fmt.Sprintf("%s_Concurrency_%d", p.name, concurrency), func(t *testing.T) {
                // Run stress test with metrics
                // Track success rate, latency, throughput
            })
        }
    }
}
```

### Workflow Agent Stress Tests (New in June 2025)

Comprehensive workflow agent stress tests have been added:

```go
func TestWorkflowAgentsConcurrentExecution(t *testing.T) {
    workflowTypes := []struct {
        name        string
        createFunc  func() domain.BaseAgent
        complexity  string
    }{
        {"Sequential", createSequentialWorkflow, "low"},
        {"Parallel", createParallelWorkflow, "medium"},
        {"Conditional", createConditionalWorkflow, "medium"},
        {"Loop", createLoopWorkflow, "high"},
        {"Nested", createNestedWorkflow, "high"},
    }
    
    concurrencyLevels := []int{10, 50, 100}
    
    // Metrics tracked:
    // - Success rate
    // - Average latency
    // - Throughput (requests/sec)
    // - Memory usage
}
```

## Best Practices

### 1. Use Appropriate Fixtures

Choose the most specific fixture for your test case:

```go
// ✅ Good: Use specific fixture
provider := fixtures.OpenAIMockProvider()

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

### 4. Test Error Conditions

Always test error scenarios:

```go
func TestWithErrors(t *testing.T) {
    // Test normal case
    t.Run("success", func(t *testing.T) {
        // ... success test
    })
    
    // Test error cases
    t.Run("rate_limit_error", func(t *testing.T) {
        provider := fixtures.RateLimitErrorProvider()
        // ... error test
    })
}
```

### 5. Use Table-Driven Tests

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
    }
    
    provider := fixtures.OpenAIMockProvider()
    
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            response, err := provider.Generate(context.Background(), test.input)
            assert.NoError(t, err)
            assert.Contains(t, response, test.expected)
        })
    }
}
```

### 6. Clean Up Resources

Ensure proper cleanup in tests:

```go
func TestWithCleanup(t *testing.T) {
    provider := fixtures.OpenAIMockProvider()
    
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
    provider := fixtures.OpenAIMockProvider()
    
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
    
    provider := fixtures.OpenAIMockProvider()
    
    t.Run("subtest1", func(t *testing.T) {
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
    provider := fixtures.OpenAIMockProvider()
    defer func() {
        if len(provider.GetCallHistory()) > 100 {
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
    provider := fixtures.OpenAIMockProvider()  // Use specific provider fixture
    // Patterns are pre-configured for realistic behavior
    
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

### From Inline Agent Setup

```go
// OLD: Inline mock agent
func TestOldAgent(t *testing.T) {
    agent := &mockAgent{
        BaseAgentImpl: core.NewBaseAgent("test", "desc", domain.AgentTypeCustom),
        runFunc: func(ctx context.Context, state *domain.State) (*domain.State, error) {
            result := state.Clone()
            result.Set("output", "processed")
            return result, nil
        },
    }
    
    // ... test code
}

// NEW: Using fixtures
func TestNewAgent(t *testing.T) {
    // Use pre-built fixture
    agent := fixtures.SimpleMockAgent()
    
    // Or create specialized agent
    agent := fixtures.SpecialistMockAgent("analyzer", "data_analysis", 50*time.Millisecond)
    
    // ... test code
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

# Run new workflow stress tests (June 2025)
go test -v ./tests/stress/workflow_stress_test.go -run TestWorkflow

# Run benchmarks
make benchmark                 # Run all benchmarks
make benchmark-pkg PKG=agent   # Run benchmarks for specific package
make benchmark-specific BENCH=BenchmarkAgentCreation  # Run specific benchmark
```

### Test Skip Control

Go-LLMs uses environment variables to control which tests are run:

```bash
# Run all tests including OpenAI API compatible provider integration tests
ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 go test ./tests/integration/...

# Skip specific provider tests even when enabled
SKIP_OPEN_ROUTER=1 ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 go test ./tests/integration/...
SKIP_OLLAMA=1 ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 go test ./tests/integration/...
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
var sharedProvider = fixtures.OpenAIMockProvider()

func TestConcurrent1(t *testing.T) {
    sharedProvider.WithPatternResponse("test1", response1)
    // ... test
}

// ✅ Solution: Create separate instances
func TestConcurrent1(t *testing.T) {
    provider := fixtures.OpenAIMockProvider()
    provider.WithPatternResponse("test1", response1)
    // ... test
}
```

### Debugging Tips

#### Enable Debug Logging

```go
func TestWithDebug(t *testing.T) {
    provider := fixtures.OpenAIMockProvider()
    provider.EnableDebugLogging(t)
    
    // Now all provider interactions will be logged
    response, err := provider.Generate(context.Background(), "test")
}
```

#### Inspect Call History

```go
func TestWithHistory(t *testing.T) {
    provider := fixtures.OpenAIMockProvider()
    
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

## Future Work and TODO Items

### Phase 3: Scenario Builder Adoption - Remaining Tasks

#### Complex Test Migration
- Identify tests with 5+ setup steps for migration
- Migrate integration tests to scenario builder
- Migrate workflow tests to scenario builder
- Create migration patterns documentation

#### Integration Test Patterns
- Standardize provider integration test setup
- Create multi-component scenario templates
- Migrate end-to-end tests to new patterns
- Document integration testing best practices

#### Documentation
- Add comprehensive scenario builder examples
- Create scenario builder cookbook with common patterns
- Document advanced testing techniques
- Create video tutorials for complex scenarios

### Phase 4: Matcher Standardization (Optional)

#### Custom Assertion Migration
- Audit existing custom assertion logic across tests
- Create domain-specific matchers for:
  - String assertions (beyond basic contains/equals)
  - State assertions (complex state validation)
  - Event assertions (event sequence validation)
- Migrate tests from custom assertions to standardized matchers
- Create matcher composition patterns

#### Event Assertion Patterns
- Standardize event verification approaches
- Create event sequence matchers for complex workflows
- Create event data matchers for payload validation
- Document event testing patterns and anti-patterns
- Add event timeline visualization support

### Long-term Enhancements

1. **Property-based testing support** - Generate test cases based on properties
2. **Fuzzing integration** - Automated testing with random inputs
3. **Test data generation** - Smart generation of test data based on schemas
4. **Visual test reporting** - HTML reports with test execution visualization
5. **CI/CD pipeline integration** - Better integration with GitHub Actions and other CI systems
6. **Performance profiling** - Built-in profiling support for identifying bottlenecks
7. **Test coverage visualization** - Visual representation of code coverage
8. **Distributed testing** - Support for running tests across multiple machines
9. **Snapshot testing** - Support for snapshot-based testing of complex outputs
10. **Contract testing** - Support for testing provider/consumer contracts

## Related Documentation

- [TESTING_INFRA_CHANGES.md](../../TESTING_INFRA_CHANGES.md) - Complete summary of test infrastructure changes
- [MIGRATION_GUIDE.md](../../MIGRATION_GUIDE.md) - Step-by-step migration instructions
- [pkg/testutils/](../../pkg/testutils/) - Test utilities package
- [pkg/testutils/fixtures/](../../pkg/testutils/fixtures/) - All available fixtures
- [migration-analysis.md](../../migration-analysis.md) - Detailed migration analysis
- [Benchmarking Framework](benchmarks.md) - Detailed overview of performance benchmarks
- [Performance Optimization](performance.md) - Comprehensive overview of performance optimization strategies
- [Sync.Pool Implementation](sync-pool.md) - Detailed guide on sync.Pool usage for memory optimization
- [Concurrency Patterns](concurrency.md) - Documentation of thread safety and concurrent execution patterns