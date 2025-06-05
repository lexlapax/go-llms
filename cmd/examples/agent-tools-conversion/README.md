# Agent-Tool Conversion Example

This example demonstrates the bidirectional conversion utilities between agents and tools in the Go-LLMs library.

## Overview

The agent-tool conversion utilities enable seamless interoperability between the agent and tool systems:
- Convert any agent to be used as a tool
- Convert any tool to be used as an agent
- Register agents in tool registries
- Forward events from tools to agent event systems
- Automatic schema mapping and parameter conversion

## Running the Example

```bash
# Build the example
go build -o agent-tools-conversion .

# Run specific examples
./agent-tools-conversion basic      # Basic agent-tool conversion
./agent-tools-conversion registry   # Registry integration
./agent-tools-conversion events     # Event forwarding
./agent-tools-conversion schema     # Schema mapping
./agent-tools-conversion chain      # Tool chains
./agent-tools-conversion mapping    # Advanced mapping

# Run all examples
./agent-tools-conversion all
```

## Examples Included

### 1. Basic Conversion (`basic`)
Shows the fundamental pattern of converting an agent to a tool:
- Creates a calculator agent with custom logic
- Converts it to a tool using `NewAgentTool`
- Executes the tool with parameters

### 2. Registry Integration (`registry`)
Demonstrates bulk registration and discovery:
- Creates multiple agents
- Registers them as tools with a prefix
- Shows registry search capabilities

### 3. Event Forwarding (`events`)
Shows how tools can emit events through the agent system:
- Creates an event-emitting tool
- Wraps it as an agent with event dispatcher
- Demonstrates progress and message events

### 4. Schema Mapping (`schema`)
Automatic parameter and result mapping:
- Agent with input/output schemas
- Automatic schema derivation for tools
- Type-safe execution with validation

### 5. Tool Chains (`chain`)
Creating composite tools from multiple agents:
- Chain of transformation agents
- Sequential processing pipeline
- Single tool interface for multiple operations

### 6. Advanced Mapping (`mapping`)
Sophisticated parameter transformations:
- Path-based extraction from nested data
- Type conversions (string to int, etc.)
- Nested state flattening

## Key Concepts

### Agent to Tool Conversion
```go
// Any agent can become a tool
agent := core.NewBaseAgent("my-agent", "Description", domain.AgentTypeCustom)
tool := tools.NewAgentTool(agent)
```

### Tool to Agent Conversion
```go
// Any tool can become an agent
tool := &MyCustomTool{}
agent := tools.NewToolAgent(tool)
```

### Event Support
```go
// Tools can emit events when wrapped with dispatcher
dispatcher := core.NewEventDispatcher(100)
agent := tools.NewToolAgentWithEvents(tool, dispatcher)
```

### Registry Integration
```go
// Bulk convert and register agents
agents := []domain.BaseAgent{agent1, agent2, agent3}
tools.RegisterAgentsAsTools(agents, registry, options)
```

## Use Cases

1. **Unified Interface**: Use agents and tools interchangeably
2. **Legacy Integration**: Wrap existing tools as modern agents
3. **Tool Discovery**: Register agents in tool registries for discovery
4. **Event Monitoring**: Add observability to tools via agent events
5. **Composition**: Build complex tools from simple agents

## Related Documentation

- [Agent Architecture](../../../docs/technical/agents.md)
- [Tool System](../../../docs/api/agent.md#tools)
- [Built-in Tools](../../../docs/user-guide/built-in-components.md)