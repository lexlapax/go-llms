# Multi-Agent System with Sub-Agents Example

This example demonstrates the powerful multi-agent features introduced in Phase 5 of the agent architecture, including:

## Key Features Demonstrated

### 1. **Automatic Tool Registration**
When you add sub-agents to a parent agent, they are automatically registered as tools that the parent can use. This means:
- Each sub-agent becomes callable through the parent's tool system
- The parent agent gains a `transfer_to_agent` tool for dynamic delegation
- Sub-agents appear in the parent's tool list with their descriptions

### 2. **Simplified API**
The example shows the new Google ADK-inspired API:
```go
// Create agent with sub-agents in one call
mainAgent, err := core.NewLLMAgentWithSubAgentsFromString(
    "assistant",
    "openai/gpt-4", 
    calculator,
    researcher,
    summarizer,
)

// Or use the builder pattern
mainAgent.WithSubAgents(newAgent1, newAgent2)
```

### 3. **Convenient Transfer Methods**
Transfer control to sub-agents easily:
```go
// Simple string input
result, err := mainAgent.TransferTo(ctx, "calculator", "reason", "5 + 3")

// Structured input
result, err := mainAgent.TransferTo(ctx, "researcher", "reason", map[string]interface{}{
    "query": "quantum computing",
    "depth": "detailed",
})
```

### 4. **Shared State Context**
Sub-agents can inherit state from their parent:
```go
// Enable shared state
mainAgent.EnableSharedState(true)
mainAgent.ConfigureStateInheritance(true, true, true)

// Create shared context
sharedCtx := domain.NewSharedStateContext(parentState)
```

### 5. **Agent Discovery**
Find and interact with sub-agents:
```go
// Get sub-agent by name
calcAgent := mainAgent.GetSubAgentByName("calculator")

// List all sub-agents
for _, agent := range mainAgent.SubAgents() {
    fmt.Printf("%s: %s\n", agent.Name(), agent.Description())
}
```

## Running the Example

```bash
go run cmd/examples/agent-sub-agents/main.go
```

## Example Output

The example demonstrates:
1. Automatic registration of sub-agents as tools
2. Direct transfer to calculator agent for math operations
3. Structured transfers with complex inputs
4. Research delegation
5. Shared state context between parent and child agents
6. Chaining multiple agents (research → summarize)
7. Sub-agent discovery and listing

## Architecture Benefits

This multi-agent architecture provides:

1. **Modularity**: Each agent focuses on a specific capability
2. **Reusability**: Sub-agents can be shared across different parent agents
3. **Dynamic Delegation**: The LLM can choose which sub-agent to use based on the task
4. **State Sharing**: Parent and child agents can share context efficiently
5. **Tool Integration**: Sub-agents seamlessly integrate with the parent's tool system

## Use Cases

This pattern is ideal for:
- Complex assistants with specialized capabilities
- Multi-stage workflows (research → analyze → summarize)
- Systems that need to delegate to domain experts
- Applications requiring modular, extensible architectures

## Next Steps

- Add real LLM providers instead of mock implementations
- Implement more sophisticated sub-agents with actual capabilities
- Use the built-in tools (web, file, system) within sub-agents
- Create hierarchical agent structures (agents with their own sub-agents)
- Implement custom merge strategies for parallel agent execution