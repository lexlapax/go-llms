# Testing Infrastructure Examples

This document contains example test code demonstrating the testing infrastructure introduced in v0.3.5.9.

## Basic Provider Testing

This example shows how to test LLM provider functionality using fixtures.

```go
func ExampleProviderTesting(t *testing.T) {
    // Use a pre-configured ChatGPT-like provider
    provider := fixtures.ChatGPTMockProvider()

    // Test basic generation
    response, err := provider.Generate(context.Background(), "Hello!")
    assert.NoError(t, err)
    assert.Contains(t, response, "Hello")

    // Test pattern-based responses
    provider.WithPatternResponse("(?i).*weather.*", mocks.Response{
        Content: "Today is sunny with 75°F",
        Metadata: map[string]interface{}{
            "location":   "test-city",
            "confidence": 0.95,
        },
    })

    response, err = provider.Generate(context.Background(), "What's the weather like?")
    assert.NoError(t, err)
    assert.Contains(t, response, "sunny")

    // Verify call history
    history := provider.GetCallHistory()
    assert.Len(t, history, 2)
    assert.Equal(t, "Hello!", history[0].Input)
    assert.Contains(t, history[1].Input, "weather")
}
```

## Tool Testing with Context

This example demonstrates testing tools with proper context setup.

```go
func ExampleToolTesting(t *testing.T) {
    // Create a calculator tool fixture
    calc := fixtures.CalculatorMockTool()
    
    // Create test context using helpers
    ctx := helpers.CreateTestToolContext()

    // Test addition
    result, err := calc.Execute(ctx, map[string]interface{}{
        "operation": "add",
        "a":         5.0,
        "b":         3.0,
    })
    assert.NoError(t, err)
    assert.Equal(t, 8.0, result["result"])

    // Test multiplication with custom state
    stateData := map[string]interface{}{
        "precision": 2,
        "mode":      "scientific",
    }
    ctx = helpers.CreateToolContextWithState(stateData)

    result, err = calc.Execute(ctx, map[string]interface{}{
        "operation": "multiply",
        "a":         2.5,
        "b":         4.0,
    })
    assert.NoError(t, err)
    assert.Equal(t, 10.0, result["result"])
}
```

## Agent Workflow Testing

This example shows testing agent workflows with different agent types.

```go
func ExampleAgentTesting(t *testing.T) {
    // Test simple agent
    agent := fixtures.SimpleMockAgent()
    input := fixtures.BasicTestState()
    input.Set("message", "Process this data")

    result, err := agent.Run(context.Background(), input)
    assert.NoError(t, err)
    
    message, exists := result.Get("message")
    assert.True(t, exists)
    assert.Equal(t, "Simple agent response", message)

    // Test research agent with specific query
    researchAgent := fixtures.ResearchMockAgent()
    researchInput := fixtures.BasicTestState()
    researchInput.Set("query", "quantum computing trends")
    researchInput.Set("depth", "comprehensive")

    result, err = researchAgent.Run(context.Background(), researchInput)
    assert.NoError(t, err)

    taskType, _ := result.Get("task_type")
    assert.Equal(t, "research", taskType)

    query, _ := result.Get("processed_query")
    assert.Contains(t, query.(string), "quantum")
}
```

## State Management Testing

This example demonstrates testing with different state configurations.

```go
func ExampleStateTesting(t *testing.T) {
    // Test with basic state
    basicState := fixtures.BasicTestState()
    
    value, exists := basicState.Get("test_key")
    assert.True(t, exists)
    assert.Equal(t, "test_value", value)

    // Test with conversation state
    convState := fixtures.ConversationTestState()
    
    messages, exists := convState.Get("conversation_history")
    assert.True(t, exists)
    assert.IsType(t, []interface{}{}, messages)

    // Test with artifacts
    artifactState := fixtures.StateWithArtifacts()
    artifacts := artifactState.GetArtifacts()
    assert.GreaterOrEqual(t, len(artifacts), 2)

    // Find specific artifacts
    var reportFound, dataFound bool
    for _, artifact := range artifacts {
        switch artifact.Name {
        case "Test Report":
            reportFound = true
            assert.Equal(t, "application/pdf", artifact.MimeType)
        case "Test Data":
            dataFound = true
            assert.Equal(t, "application/json", artifact.MimeType)
        }
    }
    assert.True(t, reportFound, "Test Report artifact should be present")
    assert.True(t, dataFound, "Test Data artifact should be present")
}
```

## Error Simulation Testing

This example shows how to test error conditions and edge cases.

```go
func ExampleErrorTesting(t *testing.T) {
    // Test provider errors
    errorProvider := fixtures.ErrorMockProvider("rate_limit")
    
    _, err := errorProvider.Generate(context.Background(), "test prompt")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "rate_limit")

    // Test slow provider with timeout
    slowProvider := fixtures.SlowMockProvider(2 * time.Second)
    
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    _, err = slowProvider.Generate(ctx, "test prompt")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "context deadline exceeded")

    // Test error tool
    errorTool := fixtures.ErrorMockTool(0.5) // 50% error rate
    toolCtx := helpers.CreateTestToolContext()

    // Run multiple times to test error rate
    var errorCount int
    for i := 0; i < 10; i++ {
        _, err := errorTool.Execute(toolCtx, map[string]interface{}{"test": "data"})
        if err != nil {
            errorCount++
        }
    }
    
    // Should have some errors but not all (due to randomness, this is approximate)
    assert.Greater(t, errorCount, 0)
    assert.Less(t, errorCount, 10)
}
```

## Complex Scenario Testing

This example demonstrates the scenario builder for complex test setups.

```go
func ExampleComplexScenario(t *testing.T) {
    scenario.NewScenario(t).
        WithMockProvider("chatgpt", map[string]mocks.Response{
            "(?i).*analyze.*": {
                Content: "Analysis shows positive trends in renewable energy sector.",
                Metadata: map[string]interface{}{
                    "confidence": 0.92,
                    "sources":    3,
                },
            },
            "(?i).*summarize.*": {
                Content: "Summary: Market shows 15% growth YoY with strong Q4 performance.",
                Metadata: map[string]interface{}{
                    "wordCount": 12,
                    "readTime":  "30s",
                },
            },
        }).
        WithTool(fixtures.WebSearchMockTool()).
        WithTool(fixtures.CalculatorMockTool()).
        WithAgent(fixtures.WorkflowMockAgent()).
        WithInput("task", "analyze market trends and provide summary").
        WithInput("sector", "renewable energy").
        ExpectOutput("status", matchers.Equals("completed")).
        ExpectOutput("confidence", matchers.GreaterThan(0.9)).
        ExpectOutput("task_type", matchers.Equals("workflow")).
        ExpectMetadata("execution_time", matchers.IsType[time.Duration]()).
        ExpectNoError().
        Run()
}
```

## Event Testing

This example shows how to test event emission and capture.

```go
func ExampleEventTesting(t *testing.T) {
    // Create event capture
    eventCapture := helpers.NewEventCapture()
    
    // Simulate some events (in real code, these would be emitted by your system)
    eventCapture.EmitEvent("agent.start", map[string]interface{}{
        "agent_id": "test-agent",
        "task":     "research",
    })
    
    eventCapture.EmitEvent("tool.execute", map[string]interface{}{
        "tool":   "web_search",
        "query":  "quantum computing",
        "result": "found 42 results",
    })
    
    eventCapture.EmitEvent("agent.complete", map[string]interface{}{
        "agent_id": "test-agent",
        "status":   "success",
        "duration": "2.5s",
    })

    // Assert on captured events
    events := eventCapture.GetEvents()
    assert.Len(t, events, 3)

    // Use the event assertion helper
    helpers.AssertEvents(t, events).
        HasType("agent.start").
        HasType("tool.execute").
        HasType("agent.complete").
        InOrder()
}
```

## Performance Testing

This example shows how to test performance characteristics.

```go
func ExamplePerformanceTesting(t *testing.T) {
    provider := fixtures.ChatGPTMockProvider()
    
    // Measure response time for multiple requests
    start := time.Now()
    
    for i := 0; i < 100; i++ {
        _, err := provider.Generate(context.Background(), "test prompt")
        assert.NoError(t, err)
    }
    
    duration := time.Since(start)
    averageTime := duration / 100
    
    // Assert performance requirements
    assert.Less(t, averageTime, 10*time.Millisecond, "Average response time should be under 10ms")
    
    // Test concurrent access
    start = time.Now()
    
    done := make(chan bool, 10)
    for i := 0; i < 10; i++ {
        go func() {
            for j := 0; j < 10; j++ {
                provider.Generate(context.Background(), "concurrent test")
            }
            done <- true
        }()
    }
    
    // Wait for all goroutines to complete
    for i := 0; i < 10; i++ {
        <-done
    }
    
    concurrentDuration := time.Since(start)
    t.Logf("Concurrent execution of 100 requests took: %v", concurrentDuration)
    
    // Verify call history is thread-safe
    history := provider.GetCallHistory()
    assert.Len(t, history, 200) // 100 + 100 from concurrent test
}
```

## Custom Matchers

This example shows how to use various matchers for flexible assertions.

```go
func ExampleMatcherUsage(t *testing.T) {
    testData := map[string]interface{}{
        "name":        "John Doe",
        "age":         30,
        "email":       "john.doe@example.com",
        "scores":      []int{85, 92, 78, 96},
        "metadata":    map[string]string{"role": "admin"},
        "lastLogin":   time.Now(),
        "isActive":    true,
        "description": nil,
    }

    // String matchers
    assert.True(t, matchers.Equals("John Doe").Match(testData["name"]))
    assert.True(t, matchers.Contains("John").Match(testData["name"]))
    assert.True(t, matchers.HasPrefix("John").Match(testData["name"]))
    assert.True(t, matchers.HasSuffix("Doe").Match(testData["name"]))
    assert.True(t, matchers.MatchesRegex(`^[A-Z][a-z]+ [A-Z][a-z]+$`).Match(testData["name"]))

    // Numeric matchers
    assert.True(t, matchers.GreaterThan(25).Match(testData["age"]))
    assert.True(t, matchers.LessThan(35).Match(testData["age"]))
    assert.True(t, matchers.GreaterThanOrEqual(30).Match(testData["age"]))

    // Type matchers
    assert.True(t, matchers.IsType[string]().Match(testData["name"]))
    assert.True(t, matchers.IsType[int]().Match(testData["age"]))
    assert.True(t, matchers.IsType[[]int]().Match(testData["scores"]))

    // Collection matchers
    assert.True(t, matchers.HasLength(4).Match(testData["scores"]))
    assert.True(t, matchers.Contains(92).Match(testData["scores"]))

    // Boolean and nil matchers
    assert.True(t, matchers.IsTrue().Match(testData["isActive"]))
    assert.True(t, matchers.IsNil().Match(testData["description"]))
    assert.True(t, matchers.IsNotNil().Match(testData["lastLogin"]))
}
```

## Migration from Legacy Code

This example shows before/after migration patterns.

```go
func ExampleMigrationPatterns(t *testing.T) {
    // === BEFORE: Manual mock setup (legacy) ===
    // This is how tests used to be written
    /*
    type OldMockProvider struct {
        generateFunc func(ctx context.Context, prompt string) (string, error)
    }
    
    oldMock := &OldMockProvider{
        generateFunc: func(ctx context.Context, prompt string) (string, error) {
            if strings.Contains(prompt, "error") {
                return "", errors.New("simulated error")
            }
            return "manual mock response", nil
        },
    }
    */

    // === AFTER: Using new testing infrastructure ===
    // Much cleaner and more maintainable
    
    // Simple case: Basic provider with pattern responses
    provider := fixtures.ChatGPTMockProvider()
    provider.WithPatternResponse("(?i).*error.*", mocks.Response{
        Content: "",
        Error:   "simulated error",
    })
    provider.WithPatternResponse(".*", mocks.Response{
        Content: "intelligent mock response based on patterns",
    })

    // Test normal case
    response, err := provider.Generate(context.Background(), "hello world")
    assert.NoError(t, err)
    assert.Contains(t, response, "Hello")

    // Test error case
    _, err = provider.Generate(context.Background(), "trigger error")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "simulated error")

    // Complex case: Multi-component scenario
    scenario.NewScenario(t).
        WithMockProvider("claude", map[string]mocks.Response{
            "(?i).*complex.*": {
                Content: "Handled complex scenario successfully",
                Metadata: map[string]interface{}{
                    "complexity": "high",
                    "processed":  true,
                },
            },
        }).
        WithAgent(fixtures.WorkflowMockAgent()).
        WithInput("task", "handle complex workflow").
        ExpectOutput("status", matchers.Equals("completed")).
        ExpectMetadata("execution_time", matchers.IsNotNil()).
        ExpectNoError().
        Run()
}
```