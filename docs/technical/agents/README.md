# Agent Documentation

> **[Documentation Home](../README.md) / Agents**

## Overview

Agents are the core abstraction for autonomous processing in go-llms. They can range from simple function wrappers to complex AI-powered entities that use tools, coordinate with other agents, and maintain state.

## Documentation Structure

### Core Documentation
- [**Agent System Overview**](overview.md) - Understanding the agent architecture
- [**LLM Agents**](llm-agents.md) - AI-powered agents with tool support
- [**Workflow Agents**](workflow-agents.md) - Orchestration patterns for agent coordination
- [**Multi-Agent Systems**](multi-agent-systems.md) - Building complex agent hierarchies
- [**State Management**](state-management.md) - Managing data flow between agents

## Agent Types

### LLM Agent
AI-powered agents that can:
- Process natural language inputs
- Use tools to extend capabilities
- Maintain conversation context
- Coordinate with sub-agents

### Workflow Agents
Orchestration agents that coordinate multiple agents:
- **Sequential Agent** - Execute agents in order
- **Parallel Agent** - Execute agents concurrently
- **Conditional Agent** - Route based on conditions
- **Loop Agent** - Iterate until conditions are met

### Custom Agents
Build your own agents by:
- Implementing the `BaseAgent` interface
- Extending existing agent types
- Composing agents together

![Workflow Patterns](../images/workflow-patterns.svg)
*Figure 1: Agent workflow orchestration patterns showing how different agent types coordinate*

## Quick Start

### Creating an LLM Agent
```go
// Create agent with provider
agent := core.NewLLMAgent("assistant", "gpt-4", core.LLMDeps{
    Provider: provider,
})

// Configure agent
agent.SetSystemPrompt("You are a helpful assistant")

// Add tools
agent.AddTool(weatherTool)
agent.AddTool(calculatorTool)

// Run agent
state := domain.NewState()
state.Set("user_input", "What's the weather in NYC?")

result, err := agent.Run(ctx, state)
```

### Creating a Workflow
```go
// Sequential workflow
workflow := workflow.NewSequentialAgent("process-pipeline")
workflow.AddAgent(validateAgent)
workflow.AddAgent(processAgent)
workflow.AddAgent(formatAgent)

// Parallel workflow
parallel := workflow.NewParallelAgent("multi-process")
parallel.AddAgent(agent1)
parallel.AddAgent(agent2)
parallel.WithMergeStrategy(workflow.MergeAll)

// Run workflow
result, err := workflow.Run(ctx, initialState)
```

### Multi-Agent Coordination
```go
// Create coordinator with sub-agents
coordinator := core.NewLLMAgent("coordinator", "gpt-4", deps)
coordinator.AddSubAgent(researchAgent)
coordinator.AddSubAgent(analysisAgent)
coordinator.AddSubAgent(reportAgent)

// Agents can transfer control
result, err := coordinator.TransferTo(ctx, "research-agent", 
    "Research climate change impacts", inputData)
```

## Common Patterns

### Tool-Using Agent
```go
agent := core.NewLLMAgent("tool-user", "gpt-4", deps)

// Add multiple tools
agent.AddTool(webSearchTool)
agent.AddTool(calculatorTool)
agent.AddTool(databaseTool)

// Agent automatically selects and uses tools based on the task
```

### Agent with Guardrails
```go
agent := core.NewLLMAgent("safe-agent", "gpt-4", deps)

// Add input validation
agent.WithInputGuardrails(inputValidator)

// Add output validation
agent.WithOutputGuardrails(outputValidator)

// Add state transformations
agent.WithInputTransforms(preprocessor)
agent.WithOutputTransforms(postprocessor)
```

### Event-Driven Agent
```go
agent := core.NewLLMAgent("observable-agent", "gpt-4", deps)

// Subscribe to events
agent.OnEvent(func(event domain.Event) {
    switch event.Type {
    case domain.EventToolCall:
        log.Printf("Tool called: %v", event.Data)
    case domain.EventStateChange:
        log.Printf("State changed: %v", event.Data)
    }
})
```

![State Management](../images/state-management.svg)
*Figure 2: State management in agent systems showing how data flows between agents and tools*

## Agent Features

### Core Features
- **State Management** - Thread-safe state handling
- **Tool Integration** - Extend capabilities with tools
- **Event System** - Observable operations
- **Error Handling** - Structured error management
- **Retry Logic** - Automatic retry with backoff

### Advanced Features
- **Sub-Agent Management** - Hierarchical agent systems
- **Guardrails** - Input/output validation
- **Transformations** - State preprocessing/postprocessing
- **Handoffs** - Agent-to-agent transfers
- **Async Execution** - Non-blocking agent runs

## Performance Considerations

### Concurrency
- Parallel agents can process concurrently
- State access is thread-safe
- Tools can be executed in parallel

### Resource Management
- Agents are lightweight
- State is cloned to prevent mutations
- Events are buffered for performance

### Optimization Tips
- Use workflow agents for orchestration
- Cache tool results when appropriate
- Limit agent depth in hierarchies
- Use streaming for long responses

## Testing Agents

### Unit Testing
```go
func TestAgent(t *testing.T) {
    // Use mock provider
    mockProvider := provider.NewMockProvider()
    mockProvider.AddResponse("Expected response")
    
    // Create agent with mock
    agent := core.NewLLMAgent("test", "model", core.LLMDeps{
        Provider: mockProvider,
    })
    
    // Test agent behavior
    state := domain.NewState()
    state.Set("input", "test")
    
    result, err := agent.Run(context.Background(), state)
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Integration Testing
```go
func TestAgentIntegration(t *testing.T) {
    // Use real provider (skip if no API key)
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        t.Skip("No API key")
    }
    
    provider := provider.NewOpenAIProvider(apiKey, "gpt-4")
    agent := core.NewLLMAgent("test", "gpt-4", core.LLMDeps{
        Provider: provider,
    })
    
    // Test with real LLM
    // ...
}
```

## Best Practices

### 1. Agent Design
- Keep agents focused on specific tasks
- Use composition over complex agents
- Leverage workflow agents for orchestration

### 2. State Management
- Keep state minimal
- Use metadata for auxiliary data
- Clone state when modifying

### 3. Tool Integration
- Make tools reusable
- Provide clear descriptions
- Validate inputs with schemas

### 4. Error Handling
- Handle errors at appropriate levels
- Use structured errors
- Implement retry strategies

### 5. Testing
- Use mocks for unit tests
- Test error scenarios
- Verify state transformations

## Next Steps

- Read [Agent System Overview](overview.md) for architecture details
- Learn about [LLM Agents](llm-agents.md) for AI-powered agents
- Explore [Workflow Agents](workflow-agents.md) for orchestration
- Understand [Multi-Agent Systems](multi-agent-systems.md) for complex scenarios
- Master [State Management](state-management.md) for data flow