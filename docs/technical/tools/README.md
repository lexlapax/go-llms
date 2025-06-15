# Tool Documentation

> **[Documentation Home](../README.md) / Tools**

## Overview

Tools extend agent capabilities by providing functions that can be called during execution. They bridge the gap between LLM reasoning and real-world actions, enabling agents to interact with external systems, perform calculations, and access information.

## Documentation Structure

### Core Documentation
- [**Tool System Overview**](overview.md) - Architecture and core concepts
- [**Creating Tools**](creating-tools.md) - Build custom tools step by step
- [**Tool Discovery**](tool-discovery.md) - Runtime registration and metadata
- [**Built-in Tools**](built-in-tools.md) - Available tools and examples

## What are Tools?

Tools are functions that agents can invoke to:
- Access external APIs and services
- Perform computations and data processing
- Interact with databases and file systems
- Execute business logic
- Bridge to other systems

### Tool Components
1. **Name** - Unique identifier for the tool
2. **Description** - Human-readable description for LLM understanding
3. **Schema** - JSON Schema defining input parameters
4. **Function** - The actual implementation

## Quick Start

### Creating a Simple Tool
```go
// Simple calculator tool
calculator := tools.NewTool(
    "calculator",
    "Perform basic arithmetic operations",
    func(params struct {
        Operation string  `json:"operation"`
        A         float64 `json:"a"`
        B         float64 `json:"b"`
    }) (map[string]interface{}, error) {
        var result float64
        switch params.Operation {
        case "add":
            result = params.A + params.B
        case "subtract":
            result = params.A - params.B
        case "multiply":
            result = params.A * params.B
        case "divide":
            if params.B == 0 {
                return nil, errors.New("division by zero")
            }
            result = params.A / params.B
        default:
            return nil, fmt.Errorf("unknown operation: %s", params.Operation)
        }
        
        return map[string]interface{}{
            "result": result,
            "operation": params.Operation,
            "a": params.A,
            "b": params.B,
        }, nil
    },
    &schema.Schema{
        Type: "object",
        Properties: map[string]schema.Property{
            "operation": {
                Type: "string",
                Enum: []string{"add", "subtract", "multiply", "divide"},
                Description: "The arithmetic operation to perform",
            },
            "a": {
                Type: "number",
                Description: "First operand",
            },
            "b": {
                Type: "number",
                Description: "Second operand",
            },
        },
        Required: []string{"operation", "a", "b"},
    },
)
```

### Adding Tools to Agents
```go
agent := core.NewLLMAgent("assistant", "gpt-4", deps)

// Add individual tools
agent.AddTool(calculator)
agent.AddTool(weatherTool)
agent.AddTool(databaseTool)

// Tools are automatically available to the agent
state := domain.NewState()
state.Set("user_input", "What is 25 times 4?")
result, err := agent.Run(ctx, state)
```

## Tool Categories

### Data Access Tools
```go
// Database query tool
dbTool := tools.NewTool(
    "query_database",
    "Query customer database",
    func(params struct {
        Query string `json:"query"`
        Limit int    `json:"limit"`
    }) (interface{}, error) {
        // Execute database query
        return db.Query(params.Query, params.Limit)
    },
    schema,
)
```

### API Integration Tools
```go
// Weather API tool
weatherTool := tools.NewTool(
    "get_weather",
    "Get current weather for a location",
    func(params struct {
        Location string `json:"location"`
        Units    string `json:"units"`
    }) (interface{}, error) {
        // Call weather API
        return weatherAPI.GetWeather(params.Location, params.Units)
    },
    schema,
)
```

### Processing Tools
```go
// Text analysis tool
analysisTool := tools.NewTool(
    "analyze_text",
    "Analyze text for sentiment and entities",
    func(params struct {
        Text string `json:"text"`
    }) (interface{}, error) {
        // Perform text analysis
        return analyzer.Analyze(params.Text)
    },
    schema,
)
```

![Tool Discovery](../images/tool-discovery.svg)
*Figure 1: Tool discovery and registration system showing how tools are dynamically registered and made available to agents*

## Advanced Tool Features

### Tool with Context
```go
// Tool that uses context
contextTool := tools.NewContextTool(
    "context_aware",
    "Tool that respects context cancellation",
    func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case result := <-doWork(params):
            return result, nil
        }
    },
    schema,
)
```

### Tool with State Access
```go
// Tool that accesses agent state
stateTool := tools.NewStateTool(
    "state_reader",
    "Read from agent state",
    func(state *domain.State, params map[string]interface{}) (interface{}, error) {
        key := params["key"].(string)
        value, exists := state.Get(key)
        if !exists {
            return nil, fmt.Errorf("key not found: %s", key)
        }
        return value, nil
    },
    schema,
)
```

### Tool Composition
```go
// Tool that uses other tools
compositeTool := tools.NewTool(
    "research_and_summarize",
    "Research a topic and provide summary",
    func(params struct {
        Topic string `json:"topic"`
    }) (interface{}, error) {
        // Use search tool
        searchResults, err := searchTool.Execute(ctx, map[string]interface{}{
            "query": params.Topic,
        })
        if err != nil {
            return nil, err
        }
        
        // Use summary tool
        summary, err := summaryTool.Execute(ctx, map[string]interface{}{
            "text": searchResults,
        })
        
        return summary, err
    },
    schema,
)
```

## Tool Discovery

### Runtime Registration
```go
// Register tool for discovery
discovery := tools.GetDiscovery()
discovery.RegisterTool(myTool)

// Discover tools by capability
webTools := discovery.FindTools(tools.WithTag("web"))
dataTools := discovery.FindTools(tools.WithTag("data"))

// Get tool metadata
metadata := discovery.GetToolMetadata("calculator")
```

### Tool Metadata
```go
// Tool with rich metadata
tool := tools.NewTool(
    "enhanced_tool",
    "Tool with metadata",
    handler,
    schema,
).WithMetadata(tools.Metadata{
    Version: "1.0.0",
    Author: "TeamName",
    Tags: []string{"data", "analysis"},
    Examples: []tools.Example{
        {
            Description: "Analyze sales data",
            Input: map[string]interface{}{
                "data_type": "sales",
                "period": "Q1-2024",
            },
            Output: "Analysis results...",
        },
    },
})
```

## Testing Tools

### Unit Testing
```go
func TestCalculatorTool(t *testing.T) {
    // Test successful calculation
    result, err := calculator.Execute(context.Background(), map[string]interface{}{
        "operation": "multiply",
        "a": 5,
        "b": 6,
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 30.0, result.(map[string]interface{})["result"])
    
    // Test error case
    _, err = calculator.Execute(context.Background(), map[string]interface{}{
        "operation": "divide",
        "a": 10,
        "b": 0,
    })
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "division by zero")
}
```

### Integration Testing
```go
func TestToolWithAgent(t *testing.T) {
    // Create mock provider
    mockProvider := provider.NewMockProvider()
    
    // Create agent with tool
    agent := core.NewLLMAgent("test", "model", core.LLMDeps{
        Provider: mockProvider,
    })
    agent.AddTool(calculator)
    
    // Set up mock to call tool
    mockProvider.AddResponse(`{"tool": "calculator", "params": {"operation": "add", "a": 10, "b": 20}}`)
    
    // Run agent
    state := domain.NewState()
    state.Set("user_input", "Add 10 and 20")
    
    result, err := agent.Run(context.Background(), state)
    assert.NoError(t, err)
    
    // Verify tool was called
    output, _ := result.Get("output")
    assert.Contains(t, output, "30")
}
```

## Best Practices

### 1. Tool Design
- **Single Responsibility**: Each tool should do one thing well
- **Clear Naming**: Use descriptive, action-oriented names
- **Comprehensive Descriptions**: Help LLMs understand when to use the tool
- **Error Handling**: Return clear, actionable error messages

### 2. Schema Definition
- **Be Explicit**: Define all parameters clearly
- **Use Constraints**: Add validation rules (min/max, enum, pattern)
- **Provide Examples**: Include example values in descriptions
- **Required Fields**: Only mark truly required fields

### 3. Implementation
- **Input Validation**: Validate beyond schema requirements
- **Error Recovery**: Handle errors gracefully
- **Resource Management**: Clean up resources properly
- **Performance**: Consider caching for expensive operations

### 4. Testing
- **Edge Cases**: Test boundary conditions
- **Error Paths**: Verify error handling
- **Integration**: Test with actual agents
- **Performance**: Benchmark expensive operations

## Common Patterns

### Retry Pattern
```go
func retryableTool(params map[string]interface{}) (interface{}, error) {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        result, err := doOperation(params)
        if err == nil {
            return result, nil
        }
        
        if !isRetryable(err) {
            return nil, err
        }
        
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return nil, fmt.Errorf("operation failed after %d retries", maxRetries)
}
```

### Caching Pattern
```go
var cache = make(map[string]cacheEntry)
var cacheMux sync.RWMutex

func cachedTool(params map[string]interface{}) (interface{}, error) {
    key := generateCacheKey(params)
    
    // Check cache
    cacheMux.RLock()
    if entry, exists := cache[key]; exists && !entry.expired() {
        cacheMux.RUnlock()
        return entry.value, nil
    }
    cacheMux.RUnlock()
    
    // Compute result
    result, err := computeExpensiveOperation(params)
    if err != nil {
        return nil, err
    }
    
    // Update cache
    cacheMux.Lock()
    cache[key] = cacheEntry{
        value:     result,
        timestamp: time.Now(),
    }
    cacheMux.Unlock()
    
    return result, nil
}
```

## Next Steps

- Understand the [Tool System Overview](overview.md)
- Learn [Creating Tools](creating-tools.md) in detail
- Explore [Tool Discovery](tool-discovery.md) for runtime registration
- See [Built-in Tools](built-in-tools.md) for ready-to-use tools
- Check the [API Reference](../api-reference/tools.md) for detailed APIs