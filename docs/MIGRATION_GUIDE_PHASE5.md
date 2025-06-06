# Migration Guide: Multi-Agent System Enhancement (Phase 5)

This guide helps you migrate existing code to use the new multi-agent features introduced in Phase 5, inspired by Google's Agent Development Kit (ADK).

## Overview

Phase 5 introduces powerful multi-agent capabilities:
- Automatic sub-agent to tool conversion
- Simplified API for multi-agent creation
- Shared state between parent and child agents
- Dynamic agent delegation via LLM

## Key Changes

### 1. Sub-Agents as Tools

**Before:**
```go
// Manual tool creation for each sub-agent
calculatorTool := tools.NewAgentTool(calculatorAgent)
researcherTool := tools.NewAgentTool(researcherAgent)

mainAgent := core.NewLLMAgent("assistant", provider)
mainAgent.AddTool(calculatorTool)
mainAgent.AddTool(researcherTool)
```

**After:**
```go
// Automatic tool registration
mainAgent, err := core.NewLLMAgentWithSubAgents(
    "assistant",
    provider,
    calculatorAgent,
    researcherAgent,
)
// Sub-agents are automatically available as tools!
```

### 2. Simplified Agent Creation

**Before:**
```go
// Complex setup
provider := llm.NewOpenAIProvider(apiKey, "gpt-4")
mainAgent := core.NewLLMAgent("coordinator", provider)

// Manual sub-agent management
subAgents := []domain.BaseAgent{techSupport, billing, senior}
for _, agent := range subAgents {
    mainAgent.AddSubAgent(agent)
    tool := tools.NewAgentTool(agent)
    mainAgent.AddTool(tool)
}
```

**After:**
```go
// One-line creation with provider string
mainAgent, err := core.NewLLMAgentWithSubAgentsFromString(
    "coordinator",
    "openai/gpt-4",  // or "anthropic/claude-3", "gemini/pro", etc.
    techSupport,
    billing,
    senior,
)
```

### 3. Agent Handoff

**Before:**
```go
// Manual handoff implementation
handoff := domain.NewHandoff().
    WithTarget("techSupport").
    WithReason("Customer needs technical help").
    WithInputTransform(func(s *domain.State) *domain.State {
        // Manual state transformation
        newState := domain.NewState()
        newState.Set("issue", s.Get("customer_issue"))
        return newState
    })

result, err := handoff.Execute(ctx, state)
```

**After:**
```go
// Direct transfer with convenience method
result, err := mainAgent.TransferTo(
    ctx,
    "techSupport",
    "Customer needs technical help",
    map[string]interface{}{
        "issue": "Internet disconnecting",
        "customer_id": "CUST-12345",
    },
)
```

### 4. Shared State Context

**Before:**
```go
// Manual state passing between agents
parentState := domain.NewState()
parentState.Set("context", "AI research project")

// Copy state manually for child
childState := domain.NewState()
childState.Set("context", parentState.Get("context"))
childState.Set("query", "transformer models")

result, err := researcher.Run(ctx, childState)
```

**After:**
```go
// Automatic state inheritance
mainAgent.EnableSharedState(true)
mainAgent.ConfigureStateInheritance(true, true, true) // messages, artifacts, metadata

// Child agents automatically see parent state
result, err := mainAgent.TransferTo(ctx, "researcher", "Research needed", 
    "transformer models")
// Child has access to parent's context automatically!
```

### 5. Sub-Agent Discovery

**Before:**
```go
// Manual tracking of sub-agents
subAgentMap := make(map[string]domain.BaseAgent)
subAgentMap["calculator"] = calculatorAgent
subAgentMap["researcher"] = researcherAgent

// Manual lookup
if agent, ok := subAgentMap["calculator"]; ok {
    // Use agent
}
```

**After:**
```go
// Built-in discovery
calcAgent := mainAgent.GetSubAgentByName("calculator")
if calcAgent != nil {
    // Use agent
}

// List all sub-agents
for _, agent := range mainAgent.SubAgents() {
    fmt.Printf("%s: %s\n", agent.Name(), agent.Description())
}
```

### 6. Builder Pattern

**Before:**
```go
mainAgent := core.NewLLMAgent("assistant", provider)
mainAgent.SetSystemPrompt("You are a helpful assistant")
mainAgent.AddTool(tool1)
mainAgent.AddTool(tool2)
mainAgent.EnableMetrics(true)
```

**After:**
```go
mainAgent := core.NewLLMAgent("assistant", provider).
    WithSubAgents(agent1, agent2, agent3).
    SetSystemPrompt("You are a helpful assistant").
    EnableSharedState(true).
    ConfigureStateInheritance(true, true, false)
```

## Migration Steps

### Step 1: Update Agent Creation

Replace manual agent creation with the new constructors:

```go
// If you have a provider instance
agent, err := core.NewLLMAgentWithSubAgents(name, provider, subAgents...)

// If you want to use provider strings
agent, err := core.NewLLMAgentWithSubAgentsFromString(name, "openai/gpt-4", subAgents...)
```

### Step 2: Remove Manual Tool Creation

Delete code that manually wraps sub-agents as tools:

```go
// Remove this pattern
for _, subAgent := range subAgents {
    tool := tools.NewAgentTool(subAgent)
    mainAgent.AddTool(tool)
}
```

### Step 3: Update Handoff Logic

Replace manual handoff implementations with TransferTo():

```go
// Instead of creating Handoff objects
result, err := mainAgent.TransferTo(ctx, targetName, reason, input)
```

### Step 4: Enable Shared State (Optional)

If your agents need to share context:

```go
mainAgent.EnableSharedState(true)
mainAgent.ConfigureStateInheritance(
    true,  // inherit messages
    true,  // inherit artifacts  
    false, // inherit metadata (optional)
)
```

### Step 5: Update Tool References

The built-in transfer tool is now available:

```go
// The agent automatically has a "transfer_to_agent" tool
// that the LLM can use to delegate to sub-agents
```

## Common Patterns

### Multi-Level Hierarchy

```go
// Create team leads with their teams
researchLead, _ := core.NewLLMAgentWithSubAgentsFromString(
    "researchLead",
    "openai/gpt-4",
    webResearcher,
    academicResearcher,
)

analysisLead, _ := core.NewLLMAgentWithSubAgentsFromString(
    "analysisLead",
    "openai/gpt-4",
    dataAnalyst,
    trendAnalyst,
)

// Create top coordinator
topCoordinator, _ := core.NewLLMAgentWithSubAgentsFromString(
    "topCoordinator",
    "openai/gpt-4o",
    researchLead,
    analysisLead,
    reportWriter,
)
```

### Dynamic Routing

```go
coordinator.SetSystemPrompt(`Route requests to the appropriate specialist:
- techSupport: For technical issues
- billingSupport: For payment issues
- seniorSupport: For escalations

Use the transfer_to_agent tool to delegate.`)

// The LLM will automatically choose the right sub-agent
```

### State Context Preservation

```go
// Enable full state sharing
agent.EnableSharedState(true)
agent.ConfigureStateInheritance(true, true, true)

// All sub-agents will see the parent's state
// Useful for maintaining conversation context, user info, etc.
```

## Troubleshooting

### Issue: Sub-agents not appearing as tools

**Solution:** Ensure sub-agents are registered in the global registry:
```go
core.Register(subAgent)
```

### Issue: Circular dependencies with agent registration

**Solution:** Only register leaf agents (sub-agents), not parent agents:
```go
// Register sub-agents
core.Register(calculator)
core.Register(researcher)
// Don't register the parent agent
```

### Issue: State not being shared

**Solution:** Enable shared state explicitly:
```go
mainAgent.EnableSharedState(true)
```

### Issue: Transfer failing

**Solution:** Check that the target agent name matches exactly:
```go
// Agent names are case-sensitive
result, err := agent.TransferTo(ctx, "techSupport", reason, input)
```

## Benefits of Migration

1. **Simpler Code**: Less boilerplate for multi-agent systems
2. **Automatic Features**: Sub-agents as tools without manual setup
3. **Better State Management**: Automatic context sharing
4. **Type Safety**: Provider strings validated at creation
5. **Cleaner APIs**: Intuitive methods for common operations

## Examples

See the following examples for complete implementations:
- `cmd/examples/agent-sub-agents/` - Basic multi-agent patterns
- `cmd/examples/agent-handoff/` - Customer support system with handoffs
- `cmd/examples/agent-multi-coordination/` - Hierarchical team coordination