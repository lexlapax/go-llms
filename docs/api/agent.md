# Agent API

Build autonomous agents - Agent lifecycle, state management, and event-driven architecture

## Package Information

- **Import Path**: `github.com/lexlapax/go-llms/pkg/agent`
- **Category**: Agent Framework
- **Stability**: Stable (v0.3.x)

## Overview

The Agent package provides a framework for building autonomous agents that can use tools, maintain state, and execute complex workflows. It supports both simple reactive agents and sophisticated multi-agent systems.

Key features:
- Flexible agent architecture
- Tool integration and management
- State persistence and recovery
- Event-driven lifecycle hooks
- Performance monitoring
- Multi-agent coordination

## Core Types

### Agent Interface

The core agent abstraction:

```go
type Agent interface {
    // Execute runs the agent with given input
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    
    // GetMetadata returns agent metadata
    GetMetadata() AgentMetadata
    
    // SetConfig updates configuration
    SetConfig(config AgentConfig) error
}
```

### Tool-Enabled Agents

Agents that can use tools:

```go
type ToolEnabledAgent interface {
    Agent
    RegisterTool(tool Tool) error
    ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error)
}
```

### Creating Agents

```go
// Create a simple LLM agent
agent := agent.NewLLMAgent(agent.Config{
    Provider: provider,
    SystemPrompt: "You are a helpful assistant.",
    Tools: []tools.Tool{
        tools.NewHTTPTool(),
        tools.NewFileTool(),
    },
})
```
## Examples

### Simple Agent

```go
agent := agent.NewSimpleAgent(agent.Config{
    Name: "helper",
    Handler: func(ctx context.Context, input interface{}) (interface{}, error) {
        // Process input
        return "Processed: " + input.(string), nil
    },
})

result, err := agent.Execute(ctx, "Hello")
```
## Best Practices

1. **Keep agents focused**: Single responsibility principle
2. **Use appropriate tools**: Only register necessary tools
3. **Implement error handling**: Graceful degradation for tool failures
4. **Monitor performance**: Track execution time and resource usage
5. **Test thoroughly**: Unit test agent logic and integration test with tools
## Error Handling

Handle agent execution errors:

```go
result, err := agent.Execute(ctx, input)
if err != nil {
    var agentErr *agent.ExecutionError
    if errors.As(err, &agentErr) {
        log.Printf("Agent %s failed: %s", agentErr.Agent, agentErr.Reason)
        // Implement recovery strategy
    }
}
```