# Agent framework architecture patterns for Go implementation

## Executive Summary

After analyzing Google's Agent Development Kit (ADK), OpenAI's Python agents SDK, and Pydantic AI SDK, several architectural patterns emerge that would translate well to a Go LLM library. Each framework takes a distinct approach: ADK emphasizes event-driven orchestration and hierarchical agent composition, OpenAI focuses on minimal abstractions with four core primitives, while Pydantic AI leverages type safety and validation as foundational principles. The analysis reveals key design patterns around concurrency, tool integration, and agent composition that align well with Go's strengths.

## Scalability comparison across frameworks

### Google ADK leads in enterprise-scale architecture

Google ADK demonstrates the most comprehensive scalability features with its **event-driven asynchronous architecture** built for production workloads. The framework supports deep agent hierarchies with automatic lifecycle management:

```python
# ADK's concurrent execution pattern
parallel_agent = ParallelAgent(
    name="concurrent_tasks",
    sub_agents=[research_agent, analysis_agent, summary_agent]
)
```

For Go implementation, this translates to channel-based coordination:

```go
func (p *ParallelAgent) RunAsync(ctx context.Context, invCtx *InvocationContext) <-chan Event {
    eventCh := make(chan Event, len(p.subAgents))
    var wg sync.WaitGroup
    
    for _, agent := range p.subAgents {
        wg.Add(1)
        go func(a Agent) {
            defer wg.Done()
            for event := range a.RunAsync(ctx, invCtx) {
                eventCh <- event
            }
        }(agent)
    }
    
    return eventCh
}
```

### OpenAI SDK prioritizes simplicity over scale

OpenAI's approach is deliberately minimal with **async-first design** but limited built-in scaling features. Resource control happens through simple parameters:

```python
result = await Runner.run(
    agent,
    input="query",
    max_turns=10,  # Prevents runaway execution
)
```

The SDK lacks explicit horizontal scaling patterns, focusing instead on single-process efficiency. This simplicity would map well to Go's philosophy of explicit over implicit behavior.

### Pydantic AI balances type safety with resource management

Pydantic AI provides **usage limits and request throttling** but is constrained by Python's GIL:

```python
result = agent.run_sync(
    'Query', 
    usage_limits=UsageLimits(response_tokens_limit=1000, request_limit=5)
)
```

## Flexibility and architectural patterns

### ADK's hierarchical agent composition

Google ADK's most elegant pattern is **"agents as functions"** - treating complex agent orchestration like composable software components:

```python
# Agent wrapping as tool
specialist_tool = AgentTool(
    agent=specialist_agent,
    description="Calls specialist for domain-specific tasks"
)

coordinator = LlmAgent(
    tools=[specialist_tool, other_tools]
)
```

This pattern translates beautifully to Go interfaces:

```go
type Tool interface {
    Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
    Schema() ToolSchema
}

type AgentTool struct {
    agent Agent
    schema ToolSchema
}

func (at *AgentTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    invCtx := NewInvocationContext(args)
    events := at.agent.RunAsync(ctx, invCtx)
    
    var result interface{}
    for event := range events {
        if event.IsFinal() {
            result = event.Content
        }
    }
    return result, nil
}
```

### OpenAI's compositional primitives

OpenAI SDK's **four core primitives** (Agents, Handoffs, Guardrails, Tracing) create a clean mental model:

```python
# Agents-as-Tools Pattern
orchestrator_agent = Agent(
    name="orchestrator",
    tools=[
        spanish_agent.as_tool(tool_name="translate_to_spanish"),
        french_agent.as_tool(tool_name="translate_to_french")
    ]
)
```

### Pydantic AI's type-driven flexibility

Pydantic AI achieves flexibility through **generic type parameters** and dependency injection:

```python
support_agent = Agent[SupportDependencies, SupportOutput](
    'openai:gpt-4o',
    deps_type=SupportDependencies,
    result_type=SupportOutput,
)
```

## Ease of use and developer experience

### ADK provides comprehensive tooling

Google ADK offers the richest developer experience with built-in UI tools:
- **adk web**: Interactive browser-based testing
- **adk eval**: Built-in evaluation framework
- **Agent Garden**: Curated sample repository

### OpenAI emphasizes minimal setup

OpenAI's strength lies in immediate productivity:

```python
agent = Agent(name="Assistant", instructions="You are a helpful assistant")
result = Runner.run_sync(agent, "Write a haiku about recursion")
```

### Pydantic AI leverages IDE integration

Pydantic AI excels in type-safe development with full IDE support and FastAPI-inspired ergonomics.

## Core capabilities comparison

### Tool integration patterns

All three frameworks converge on similar tool integration patterns, but with different philosophies:

**ADK's Universal Tool Interface:**
```python
class BaseTool:
    async def run_async(self, context: ToolContext) -> Any:
        raise NotImplementedError
```

**OpenAI's Function Tool Pattern:**
```python
@function_tool
async def fetch_weather(location: Location) -> str:
    """Fetch weather for a location."""
    return f"Weather data for {location}"
```

**Pydantic AI's Type-Safe Tools:**
```python
@agent.tool
async def get_account_balance(ctx: RunContext[SupportDependencies], account_id: str) -> str:
    return await ctx.deps.db.get_balance(ctx.deps.customer_id, account_id)
```

### State management approaches

- **ADK**: Persistent session state with artifact storage
- **OpenAI**: Run-scoped state, no built-in persistence
- **Pydantic AI**: Dependency injection for external state management

## Architectural insights for Go implementation

### Event-driven communication pattern

Borrowing from ADK's event-driven architecture:

```go
type Event struct {
    Author       string    `json:"author"`
    Content      Content   `json:"content"`
    Actions      Actions   `json:"actions"`
    InvocationID string    `json:"invocation_id"`
    Timestamp    time.Time `json:"timestamp"`
}

type Agent interface {
    RunAsync(ctx context.Context, invCtx *InvocationContext) <-chan Event
}
```

### Context-based dependency injection

Inspired by Pydantic AI's approach:

```go
type RunContext[T any] struct {
    context.Context
    Deps T
    Retry int
}

type Agent[TDeps any, TOutput any] struct {
    model      Model
    tools      []Tool[TDeps, any]
    depsType   TDeps
    outputType TOutput
}
```

### Minimal abstraction philosophy

Following OpenAI's approach of few, powerful primitives:

```go
// Core primitives only
type Agent interface { /* ... */ }
type Tool interface { /* ... */ }
type Handoff interface { /* ... */ }
type Guardrail interface { /* ... */ }
```

### Type-safe structured outputs

Adapting Pydantic AI's validation patterns:

```go
type Validatable interface {
    Validate() error
}

type StructuredOutput struct {
    validator *validator.Validate
}

func (s *StructuredOutput) UnmarshalJSON(data []byte) error {
    // Unmarshal and validate
    if err := json.Unmarshal(data, s); err != nil {
        return err
    }
    return s.Validate()
}
```

## Google ADK's special considerations

### Supporting agents via simpler interfaces

ADK's approach to simplifying agent interfaces through declarative configuration is particularly elegant:

```python
agent = LlmAgent(
    name="coordinator",
    model="gemini-2.0-flash",
    instruction="Delegate tasks based on user intent",
    sub_agents=[booking_agent, info_agent],
    transfer_enabled=True  # Automatic sub-agent routing
)
```

For Go, this suggests a builder pattern with functional options:

```go
agent := NewAgent(
    WithName("coordinator"),
    WithModel("gemini-2.0-flash"),
    WithInstruction("Delegate tasks based on user intent"),
    WithSubAgents(bookingAgent, infoAgent),
    WithTransferEnabled(true),
)
```

### Agent wrapping as tools

ADK's ability to wrap agents as tools creates powerful composition patterns. In Go:

```go
func WrapAgentAsTool(agent Agent) Tool {
    return &AgentTool{
        agent: agent,
        schema: generateSchemaFromAgent(agent),
    }
}
```

### Architectural elegance through event streams

ADK's event-driven architecture provides clean separation of concerns and natural concurrency boundaries, which aligns perfectly with Go's channel-based concurrency model.

## Key recommendations for go-llms library

Based on this analysis, I recommend the following architectural decisions for your Go LLM library:

1. **Adopt event-driven architecture** with channels for agent communication
2. **Use interface-based design** for maximum flexibility and testability
3. **Implement context-based dependency injection** for clean resource management
4. **Provide minimal core primitives** that compose well
5. **Support structured outputs** with validation interfaces
6. **Enable agent-as-tool wrapping** for hierarchical composition
7. **Use functional options pattern** for configuration
8. **Leverage Go's type system** for compile-time safety where possible

The combination of ADK's event-driven orchestration, OpenAI's minimal abstractions, and Pydantic AI's type safety patterns creates a powerful foundation for a Go-based agent framework that would be both performant and developer-friendly.