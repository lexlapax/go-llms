# Core Concepts

> **[Documentation Home](README.md) / Core Concepts**

## Overview

This guide explains the fundamental concepts and abstractions in go-llms. Understanding these concepts is essential for effectively using and extending the library.

## Provider

A **Provider** is an abstraction over different LLM services. It handles communication with LLM APIs and provides a unified interface regardless of the underlying service.

### Key Characteristics
- **Unified Interface**: Same methods work across all providers
- **Configuration**: Provider-specific options via functional options pattern
- **Capabilities**: Different providers support different features (streaming, function calling, etc.)

### Example
```go
// Creating providers
openai := provider.NewOpenAIProvider(apiKey, "gpt-4")
anthropic := provider.NewAnthropicProvider(apiKey, "claude-3-5-sonnet-latest")
gemini := provider.NewGeminiProvider(apiKey, "gemini-2.0-flash")

// Using any provider
response, err := provider.Generate(ctx, "What is Go?")
```

## Agent

An **Agent** is an autonomous entity that can process inputs, make decisions, and produce outputs. Agents can use tools, maintain state, and coordinate with other agents.

### Types of Agents

#### LLM Agent
Powered by a language model, can use tools and maintain conversation context.
```go
agent := core.NewLLMAgent("assistant", "gpt-4", core.LLMDeps{
    Provider: provider,
}
agent.SetSystemPrompt("You are a helpful assistant")
agent.AddTool(calculatorTool)
```

#### Workflow Agents
Orchestrate multiple agents in specific patterns:
- **Sequential**: Execute agents one after another
- **Parallel**: Execute agents concurrently
- **Conditional**: Route to agents based on conditions
- **Loop**: Iterate until a condition is met

### Agent Lifecycle
```
Initialize → Configure → Add Tools/Sub-agents → Run → Process State → Return Result
```

## Tool

A **Tool** is a function that agents can call to perform specific tasks. Tools extend agent capabilities beyond text generation.

### Tool Components
1. **Name**: Unique identifier
2. **Description**: What the tool does (used by LLM for selection)
3. **Schema**: JSON Schema defining input parameters
4. **Execute**: The actual function implementation

### Example Tool
```go
weatherTool := tools.NewTool(
    "get_weather",
    "Get current weather for a location",
    func(params struct {
        Location string `json:"location"`
    }) (map[string]interface{}, error) {
        // Implementation
        return map[string]interface{}{
            "temperature": 72,
            "condition": "sunny",
        }, nil
    },
    &schema.Schema{
        Type: "object",
        Properties: map[string]schema.Property{
            "location": {
                Type: "string",
                Description: "City name",
            },
        },
        Required: []string{"location"},
    },
)
```

## State

**State** represents the data flowing through agents. It's a thread-safe container for passing information between agents and tools.

### State Features
- **Thread-Safe**: Safe for concurrent access
- **Immutable Operations**: Clone for modifications
- **Metadata Support**: Attach additional context
- **Change Tracking**: Monitor state evolution

### State Operations
```go
// Create state
state := domain.NewState()

// Set values
state.Set("user_input", "Hello")
state.Set("context", contextData)

// Get values
input, exists := state.Get("user_input")

// Clone and modify
newState := state.Clone()
newState.Set("processed", true)

// Metadata
state.SetMetadata("source", "user")
```

![State Management](../images/state-management.svg)
*Figure 1: State lifecycle and management patterns showing how state flows through agents and tools*

## Event System

The **Event System** provides observability into agent operations. Events are emitted throughout execution for monitoring, debugging, and integration.

### Event Types
- `EventAgentStart`: Agent begins execution
- `EventAgentComplete`: Agent finishes successfully
- `EventAgentError`: Agent encounters an error
- `EventToolCall`: Tool is invoked
- `EventStateChange`: State is modified

### Event Handling
```go
// Subscribe to events
agent.OnEvent(func(event domain.Event) {
    switch event.Type {
    case domain.EventToolCall:
        log.Printf("Tool called: %s", event.Data)
    case domain.EventAgentError:
        log.Printf("Error: %v", event.Data)
    }
}
```

## Schema System

The **Schema System** provides JSON Schema validation for structured inputs and outputs. It ensures type safety and enables reliable tool calling.

### Schema Components
- **Type Definitions**: Define expected data types
- **Validation Rules**: Constraints on values
- **Property Descriptions**: Documentation for LLMs
- **Required Fields**: Enforce mandatory parameters

### Schema Usage
```go
schema := &schema.Schema{
    Type: "object",
    Properties: map[string]schema.Property{
        "name": {
            Type: "string",
            MinLength: ptr(1),
            MaxLength: ptr(100),
        },
        "age": {
            Type: "integer",
            Minimum: ptr(0.0),
            Maximum: ptr(150.0),
        },
    },
    Required: []string{"name"},
}
```

## Error Handling

Go-LLMs uses a structured error system that provides context and enables proper error handling across different layers.

### Error Categories
- **Provider Errors**: API communication issues
- **Validation Errors**: Schema or input validation failures
- **Tool Errors**: Tool execution failures
- **Agent Errors**: Agent-level failures
- **Bridge Errors**: Serializable errors for external systems

### Error Handling Pattern
```go
result, err := agent.Run(ctx, state)
if err != nil {
    switch e := err.(type) {
    case *errors.ProviderError:
        // Handle provider-specific error
        if e.IsRateLimit() {
            // Implement retry logic
        }
    case *errors.ValidationError:
        // Handle validation error
        log.Printf("Invalid input: %v", e.Details)
    default:
        // Handle generic error
    }
}
```

## Provider Metadata

**Provider Metadata** describes capabilities, models, and constraints of LLM providers. This enables dynamic discovery and adaptation.

### Metadata Components
```go
type ProviderMetadata struct {
    Name         string
    Description  string
    Capabilities []Capability  // streaming, function_calling, vision
    Models       []ModelInfo   // available models
    Constraints  Constraints   // rate limits, max tokens
}
```

## Workflow Patterns

![Workflow Patterns](../images/workflow-patterns.svg)
*Figure 2: Agent workflow orchestration patterns including sequential, parallel, conditional, and loop patterns*

Workflow patterns define how multiple agents collaborate:

## Type Conversion

The **Type Conversion Registry** enables seamless conversion between different type systems, particularly important for bridge integration.

### Conversion Flow
```
Go Type → Converter → Bridge Type → External System
```

### Registration
```go
registry.RegisterConverter(
    reflect.TypeOf(MyType{}),
    func(v interface{}) (interface{}, error) {
        // Conversion logic
    },
)
```

## Best Practices

### 1. **State Management**
- Keep state minimal and focused
- Use metadata for auxiliary information
- Clone state when modifications are needed

### 2. **Tool Design**
- Make tools focused and composable
- Provide clear descriptions for LLM understanding
- Use schemas for robust validation

### 3. **Agent Composition**
- Start simple, compose for complexity
- Use workflow agents for orchestration
- Leverage events for observability

### 4. **Error Handling**
- Handle errors at appropriate levels
- Provide context in error messages
- Use structured errors for better handling

### 5. **Performance**
- Use streaming for long responses
- Implement caching where appropriate
- Monitor with events and metrics

## Next Steps

- Explore specific component documentation:
  - [Providers](providers/README.md)
  - [Agents](agents/README.md)
  - [Tools](tools/README.md)
- See [API Reference](api-reference/README.md) for detailed APIs
- Check [Examples](/cmd/examples/) for practical usage