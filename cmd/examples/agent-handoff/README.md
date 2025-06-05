# Agent Handoff Example

**Status: Placeholder - Awaiting Phase 5 Implementation**

This example will demonstrate agent-to-agent handoff patterns once Phase 5 of the agent architecture is complete.

## Current Status

The Handoff interface exists but execution is not implemented. The TODO in `handoff.go` states:
> "TODO: In Phase 2, this will use the agent registry to find and execute the target agent"

## What's Coming in Phase 5

Phase 5 will implement multi-agent enhancements inspired by Google's Agent Development Kit (ADK):

### 1. Automatic Sub-Agent Registration
```go
// Simple API for creating agents with sub-agents
coordinator := core.NewLLMAgentWithSubAgents("coordinator", "gpt-4", 
    []domain.BaseAgent{
        techSupport,
        billingSupport,
        seniorSupport,
    })

// Sub-agents automatically become available as tools!
```

### 2. Dynamic Agent Delegation
The LLM will be able to transfer control to sub-agents using a built-in tool:
```json
{
  "tool": "transfer_to_agent",
  "arguments": {
    "agent_name": "techSupport",
    "reason": "Customer reporting login errors"
  }
}
```

### 3. Shared State Context
- Automatic state sharing between parent and child agents
- Child agents can access parent state
- Updates can flow back to parent
- No manual state passing required

### 4. Simplified Handoff Patterns
```go
// Before Phase 5: Manual handoff execution
handoff := domain.NewHandoff().From("A").To("B").Build()
// ... manual registry lookup and execution ...

// After Phase 5: Automatic through sub-agents
// Just attach sub-agents and let the LLM decide!
```

## Use Cases (After Phase 5)

1. **Customer Service Router**
   - Main agent analyzes requests
   - Automatically transfers to tech, billing, or escalation agents
   - State flows seamlessly between agents

2. **Multi-Stage Processing**
   - Research agent → Analysis agent → Writer agent
   - Each stage is a sub-agent with automatic handoff

3. **Expert Consultation**
   - General agent consults specialist sub-agents
   - Aggregates responses from multiple experts

## Running the Example

```bash
# Currently shows placeholder information
go run main.go

# After Phase 5, will demonstrate full handoff capabilities
```

## Phase 5 Implementation Timeline

See TODO.md for detailed implementation plan:

1. **Phase 5.1**: Core Handoff Implementation (1-2 days)
2. **Phase 5.2**: Auto-Tool Registration (1 day)
3. **Phase 5.3**: Shared State Context (1 day)
4. **Phase 5.4**: API Simplification (1 day)
5. **Phase 5.5**: Examples and Documentation (1 day)

## Benefits After Phase 5

- **Simpler Code**: No manual handoff execution
- **Dynamic Routing**: LLM decides which agent to use
- **Better State Management**: Automatic state sharing
- **Google ADK Parity**: Similar capabilities and patterns

## Related Examples

- `agent-multi-coordination` - Will show advanced multi-agent patterns
- `agent-sub-agents` - New example coming in Phase 5.5
- `workflow-conditional` - Current alternative for routing logic