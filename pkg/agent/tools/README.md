# Agent-Tool Integration

This package provides bidirectional conversion between Agents and Tools, enabling seamless integration between the state-based agent system and the parameter-based tool system.

## Overview

The agent-tool integration provides two main components:

1. **AgentTool**: Wraps a `BaseAgent` to expose it as a `Tool`
2. **ToolAgent**: Wraps a `Tool` to expose it as a `BaseAgent`

This allows for:
- Using agents as tools in LLM function calling
- Using tools as agents in workflows
- Composing complex systems with both agents and tools

## AgentTool

`AgentTool` wraps an agent to make it usable as a tool.

### Basic Usage

```go
// Create an agent
agent := myTextProcessingAgent()

// Wrap as tool
tool := tools.NewAgentTool(agent)

// Use as tool
result, err := tool.Execute(ctx, "input text")
```

### Custom Mappings

```go
tool := tools.NewAgentTool(agent).
    // Map tool parameters to state keys
    WithStateMapper(tools.CreateStateMapper(map[string]string{
        "text": "input",     // tool param "text" -> state key "input"
        "mode": "settings",  // tool param "mode" -> state key "settings"
    })).
    // Extract specific fields from result state
    WithResultMapper(tools.CreateResultMapper("output", "status"))
```

### Parameter Schema

```go
schema := &sdomain.Schema{
    Type: "object",
    Properties: map[string]sdomain.Property{
        "text": {Type: "string", Description: "Input text"},
        "mode": {Type: "string", Enum: []string{"upper", "lower"}},
    },
    Required: []string{"text"},
}

tool := tools.NewAgentTool(agent).WithParameterSchema(schema)
```

## ToolAgent

`ToolAgent` wraps a tool to make it usable as an agent.

### Basic Usage

```go
// Create a tool
tool := myCalculatorTool()

// Wrap as agent
agent := tools.NewToolAgent(tool)

// Use as agent
state := domain.NewState()
state.Set("input", params)
result, err := agent.Run(ctx, state)
```

### Custom Mappings

```go
agent := tools.NewToolAgent(tool).
    // Extract tool parameters from state
    WithParamMapper(tools.CreateParamMapper(map[string]string{
        "num1": "a",        // state key "num1" -> param "a"
        "num2": "b",        // state key "num2" -> param "b"
        "operation": "op",  // state key "operation" -> param "op"
    })).
    // Update state with prefixed results
    WithStateUpdater(tools.CreateStateUpdaterWithPrefix("calc"))
```

## Mapper Functions

### State Mappers (for AgentTool)

```go
// Default mapper - handles maps, strings, and State objects
DefaultStateMapper

// Custom field mapping
CreateStateMapper(map[string]string{
    "param_name": "state_key",
})

// Custom function
customMapper := func(ctx context.Context, params interface{}) (*domain.State, error) {
    // Custom mapping logic
    return state, nil
}
```

### Result Mappers (for AgentTool)

```go
// Default mapper - looks for "result", "output", or "response" keys
DefaultResultMapper

// Extract specific fields
CreateResultMapper("field1", "field2")

// Custom function
customMapper := func(ctx context.Context, state *domain.State) (interface{}, error) {
    // Custom extraction logic
    return result, nil
}
```

### Parameter Mappers (for ToolAgent)

```go
// Default mapper - looks for "params" or "input" keys
DefaultParamMapper

// Map state keys to parameter names
CreateParamMapper(map[string]string{
    "state_key": "param_name",
})

// Extract single parameter
CreateSingleParamMapper("input_text")
```

### State Updaters (for ToolAgent)

```go
// Default updater - sets "result" and "success" keys
DefaultStateUpdater

// Add prefix to all result keys
CreateStateUpdaterWithPrefix("tool_name")

// Custom function
customUpdater := func(ctx context.Context, state *domain.State, result interface{}, err error) (*domain.State, error) {
    // Custom update logic
    return state, nil
}
```

## Integration with LLMAgent

The agent-tool wrappers integrate seamlessly with LLMAgent:

```go
// Create LLM agent
llmAgent := core.NewLLMAgent("assistant", "AI assistant", provider)

// Add regular tools
llmAgent.AddTool(calculatorTool)

// Add agents as tools
textAgent := createTextProcessingAgent()
llmAgent.AddTool(tools.NewAgentTool(textAgent))

// The LLM can now call both tools and agents
```

## Use Cases

### 1. Expose Complex Agents as Simple Tools

```go
// Complex multi-step agent
researchAgent := workflow.NewSequentialAgent("researcher").
    AddSubAgent(searchAgent).
    AddSubAgent(analyzeAgent).
    AddSubAgent(summarizeAgent)

// Expose as simple tool
researchTool := tools.NewAgentTool(researchAgent).
    WithStateMapper(func(ctx context.Context, params interface{}) (*domain.State, error) {
        query := params.(string)
        state := domain.NewState()
        state.Set("query", query)
        return state, nil
    }).
    WithResultMapper(tools.CreateResultMapper("summary"))

// Now usable in LLM function calling
llmAgent.AddTool(researchTool)
```

### 2. Use Tools in Agent Workflows

```go
// Existing tools
calcTool := createCalculatorTool()
weatherTool := createWeatherTool()

// Wrap as agents
calcAgent := tools.NewToolAgent(calcTool)
weatherAgent := tools.NewToolAgent(weatherTool)

// Use in workflow
workflow := workflow.NewSequentialAgent("data-processor").
    AddSubAgent(weatherAgent).  // Get weather data
    AddSubAgent(calcAgent)      // Process temperature calculations
```

### 3. Bidirectional Conversion

```go
// Start with agent
agent := myAgent()

// Convert to tool for LLM use
tool := tools.NewAgentTool(agent)

// Convert back to agent for workflow use
agentAgain := tools.NewToolAgent(tool)
```

## Best Practices

1. **Clear Naming**: Use descriptive names for state keys and parameters
2. **Schema Definition**: Always define parameter schemas for better validation
3. **Error Handling**: Implement proper error handling in mappers
4. **State Isolation**: Be mindful of state modifications in workflows
5. **Testing**: Test both directions of conversion

## Examples

See `example_test.go` for complete working examples of:
- Wrapping agents as tools
- Wrapping tools as agents
- Bidirectional conversion
- Custom mappers and updaters