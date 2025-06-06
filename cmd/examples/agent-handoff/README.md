# Agent Handoff Example - Customer Support System

This example demonstrates a complete customer support system using agent handoff patterns, showcasing the multi-agent features from Phase 5.

## Overview

The example implements a customer support system with:
- **Support Coordinator**: Main agent that routes issues
- **Tech Support Agent**: Handles technical issues
- **Billing Support Agent**: Handles billing and payment issues  
- **Senior Support Agent**: Handles escalated complex issues

## Key Features Demonstrated

### 1. **Simplified Multi-Agent Creation**
```go
coordinator, err := core.NewLLMAgentWithSubAgentsFromString(
    "supportCoordinator",
    "openai/gpt-4",
    techSupport,
    billingSupport,
    seniorSupport,
)
```

### 2. **Automatic Tool Registration**
Sub-agents are automatically available as tools:
- `techSupport` - Technical troubleshooting
- `billingSupport` - Billing and payments
- `seniorSupport` - Escalated issues
- `transfer_to_agent` - Built-in delegation tool

### 3. **Direct Handoff with TransferTo()**
```go
result, err := coordinator.TransferTo(ctx, "techSupport", 
    "Customer reporting internet issues", 
    map[string]interface{}{
        "issue": "Internet disconnecting",
        "customer_id": "CUST-12345",
    })
```

### 4. **State Preservation**
Customer context (ID, issue details) is maintained across handoffs.

### 5. **Conditional Escalation**
Agents can determine when escalation is needed:
```go
if needsEscalation, ok := result.Get("needs_escalation"); ok && needsEscalation.(bool) {
    // Escalate to senior support with context
}
```

## Running the Example

```bash
go run cmd/examples/agent-handoff/main.go
```

## Example Scenarios

### Scenario 1: Technical Issue
- Customer reports internet connectivity problem
- Routed to tech support
- Tech support provides troubleshooting steps

### Scenario 2: Billing Issue  
- Customer requests refund for double charge
- Routed to billing support
- Billing processes refund ($49.99)

### Scenario 3: Complex Issue with Escalation
- Customer reports API authentication failure
- Tech support attempts resolution
- Determines escalation needed
- Senior support provides custom solution with service credit ($25)

## Architecture Benefits

1. **Specialization**: Each agent focuses on specific domain expertise
2. **Scalability**: Easy to add new specialist agents
3. **Context Preservation**: Customer information flows through handoffs
4. **Flexible Routing**: Coordinator can intelligently route based on issue type
5. **Escalation Paths**: Built-in support for tiered support levels

## Real-World Applications

This pattern is ideal for:
- Customer support systems
- IT helpdesk implementations
- Multi-tier technical support
- Any system requiring specialized handling based on request type

## Next Steps

- Replace mock provider with actual LLM (OpenAI, Anthropic, etc.)
- Add more sophisticated routing logic in coordinator
- Implement actual integration with support ticket systems
- Add metrics and logging for handoff tracking
- Create more specialized agents (security, network, database, etc.)