# Agent Architecture

> **[Documentation Home](/REFERENCE.md) / [Technical Documentation](/docs/technical/) / Agent Architecture**

This document describes the comprehensive agent architecture implementation in Go-LLMs, covering the redesign principles, core infrastructure, and enhanced components that enable sophisticated agent workflows.

## Table of Contents

1. [Overview](#overview)
2. [Architecture Redesign](#architecture-redesign)
3. [Design Principles](#design-principles)
4. [Architecture Phases](#architecture-phases)
5. [Core Components (Phase 1)](#core-components-phase-1)
6. [Enhanced Components (Phase 1.5)](#enhanced-components-phase-15)
7. [Component Integration](#component-integration)
8. [Usage Examples](#usage-examples)
9. [Testing Strategy](#testing-strategy)
10. [Future Phases](#future-phases)

## Overview

The Go-LLMs agent architecture represents a comprehensive restructuring based on Google's Agent Development Kit patterns, OpenAI's approach, and Pydantic AI's type-safety concepts. This architecture provides:

- **Type-safe agent interfaces** for consistent behavior
- **Flexible state management** with concurrency safety  
- **Event-driven architecture** for monitoring and debugging
- **Composable components** for validation, transformation, and delegation
- **OpenTelemetry-compatible tracing** for observability
- **Functional programming patterns** for event processing
- **Hierarchical agent composition** for complex workflows

The architecture follows a modular design where each component has a specific responsibility and can be used independently or composed together.

## Architecture Redesign

### Previous Limitations

The original architecture had several limitations that necessitated the restructuring:

1. **Mixed Concerns**: Single `Agent` interface mixing LLM and workflow concerns
2. **Limited Composability**: Agents could not be composed hierarchically
3. **Tool-Agent Separation**: Tools and agents were completely separate concepts
4. **Implicit State Management**: State was managed implicitly through context
5. **No Workflow Abstractions**: No clear workflow agent abstractions

### Redesign Goals

The restructuring aims to create:

1. **Unified Base Agent**: All agents derive from a common `BaseAgent` interface
2. **Clear Separation**: Distinction between LLM agents and workflow agents
3. **Enhanced Composability**: Agents can contain sub-agents and be wrapped as tools
4. **Explicit State Management**: Clear state passing between agents
5. **Standardized Lifecycle**: Execution lifecycle with hooks and events

## Design Principles

The new architecture follows these core principles:

1. **Unified Base Agent**: All agents derive from a common `BaseAgent` interface
2. **Separation of Concerns**: Clear distinction between LLM agents and workflow agents
3. **Composability**: Agents can contain sub-agents and be wrapped as tools
4. **State Management**: Explicit state passing between agents with validation
5. **Lifecycle Management**: Standardized execution lifecycle with hooks
6. **Minimal Core Primitives**: Following OpenAI's approach - Agent, Tool, Handoff, Guardrail, Tracer
7. **Type-Safe Dependency Injection**: Inspired by Pydantic AI's approach with generics
8. **Channel-Based Concurrency**: Leveraging Go's strengths for event-driven architecture
9. **Event Streams**: Functional programming patterns for event processing
10. **Built-in Observability**: OpenTelemetry integration from the start

### Planned Interface Hierarchy

```
BaseAgent (interface)
├── LLMAgent (struct) - Agents powered by language models
├── WorkflowAgent (interface) - Agents that orchestrate other agents
│   ├── SequentialAgent (struct) - Execute agents in sequence
│   ├── ParallelAgent (struct) - Execute agents in parallel
│   ├── ConditionalAgent (struct) - Conditional execution
│   └── LoopAgent (struct) - Iterative execution
└── CustomAgent (interface) - User-defined agent logic
```

## Architecture Phases

### Phase 1: Core Infrastructure ✅ COMPLETED
- Base agent interfaces and implementations
- State management with thread safety
- Event system for monitoring
- Artifact handling
- Agent hierarchy and registry

### Phase 1.5: Enhanced Core Infrastructure ✅ COMPLETED
- Agent delegation (Handoff)
- Input/output validation (Guardrails)
- Type-safe dependency injection (RunContext)
- Functional event stream processing
- State validation and transformation
- OpenTelemetry-compatible tracing

### Phase 2: LLM Agent Migration ✅ COMPLETED (February 3, 2025)
- LLM agent implementation based on existing DefaultAgent
- Tool integration with new interface
- State-based prompt engineering
- Error handling and retries
- Agent hierarchy support
- Full hook system implementation
- Ultra-simple agent creation from strings
- Provider string parsing with aliases
- Removed old workflow package

### Phase 3: Workflow Agents 🎯 NEXT PRIORITY
- Sequential Agent: Execute agents in sequence
- Parallel Agent: Execute agents in parallel with merge strategies
- Conditional Agent: Conditional execution based on state
- Loop Agent: Iterative execution patterns

### Phase 4: Agent-Tool Integration 📋 PLANNED
- AgentTool wrapper: Convert agents to tools
- Tool context system: Enhanced tool execution context
- Bidirectional conversion utilities

### Phase 5: Advanced Features 📋 PLANNED
- State persistence and serialization
- Agent discovery and registry
- Advanced merge strategies for parallel agents
- Streaming support for long-running agents

### Phase 6: Migration and Testing 📋 PLANNED
- Migration guide and documentation
- Update all examples to new architecture
- Comprehensive testing and benchmarking
- Performance optimization

## Core Components (Phase 1)

### BaseAgent Interface

The foundation for all agent types, providing:

```go
type BaseAgent interface {
    // Identification
    ID() string
    Name() string
    Description() string
    Type() AgentType
    
    // Hierarchy Management
    Parent() BaseAgent
    SetParent(parent BaseAgent) error
    SubAgents() []BaseAgent
    AddSubAgent(agent BaseAgent) error
    
    // Execution
    Run(ctx context.Context, input *State) (*State, error)
    RunAsync(ctx context.Context, input *State) (<-chan Event, error)
    
    // Lifecycle Hooks
    Initialize(ctx context.Context) error
    BeforeRun(ctx context.Context, state *State) error
    AfterRun(ctx context.Context, state *State, result *State, err error) error
    Cleanup(ctx context.Context) error
    
    // Schema and Configuration
    InputSchema() *schema.Schema
    OutputSchema() *schema.Schema
    Config() AgentConfig
    Validate() error
}
```

**Key Features:**
- **Type Safety**: Strong typing with compile-time checks
- **Hierarchy Support**: Parent-child agent relationships
- **Lifecycle Management**: Initialization, execution, and cleanup hooks
- **Schema Validation**: Input/output schema definitions
- **Metadata Support**: Extensible metadata system

### State Management

Thread-safe state container that flows between agents:

```go
type State struct {
    // Core state data
    values     map[string]interface{}
    artifacts  map[string]*Artifact
    messages   []Message
    
    // Metadata and lineage
    metadata   map[string]interface{}
    parentID   string
    version    int
}
```

**Key Features:**
- **Thread Safety**: Concurrent read/write protection
- **Immutability**: Clone operations for safe state passing
- **Versioning**: State version tracking for debugging
- **Artifact Support**: File and data artifact management
- **Message History**: Conversation message storage

### Event System

Comprehensive event system for monitoring and debugging:

```go
type Event struct {
    ID        string
    Type      EventType
    AgentID   string
    AgentName string
    Timestamp time.Time
    Data      interface{}
    Error     error
}
```

**Event Types:**
- `EventAgentStart`, `EventAgentComplete`, `EventAgentError`
- `EventStateUpdate`, `EventProgress`, `EventMessage`
- `EventToolCall`, `EventToolResult`, `EventToolError`
- `EventSubAgentStart`, `EventSubAgentEnd`, `EventWorkflowStep`

**Key Features:**
- **Non-blocking Dispatch**: Events don't block execution
- **Filtering**: Subscribe to specific event types
- **Error Propagation**: Error events for debugging
- **Structured Data**: Rich event data for analysis

## Enhanced Components (Phase 1.5)

### Handoff Interface

Agent delegation mechanism with input transformation:

```go
type Handoff interface {
    Name() string
    Description() string
    TargetAgent() string
    
    Execute(ctx context.Context, state *State) (*State, error)
    TransformInput(state *State) *State
    FilterMessages(messages []Message) []Message
}
```

**Pre-built Patterns:**
- `NewSimpleHandoff`: Passes state unchanged
- `NewFilteredHandoff`: Filters specific keys
- `NewMessagesOnlyHandoff`: Only passes messages
- `NewLastNMessagesHandoff`: Passes last N messages

**Usage:**
```go
handoff := NewHandoffBuilder("summarizer", "summary-agent").
    WithDescription("Handoff to summarization agent").
    WithInputFilter(func(state *State) *State {
        return LimitMessagesTransform(5)(context.Background(), state)
    }).
    Build()
```

### Guardrails Interface

Input/output validation for agent safety:

```go
type Guardrail interface {
    Name() string
    Type() GuardrailType // input, output, both
    
    Validate(ctx context.Context, state *State) error
    ValidateAsync(ctx context.Context, state *State, timeout time.Duration) <-chan error
}
```

**Pre-built Guardrails:**
- `RequiredKeysGuardrail`: Ensures required keys exist
- `MaxStateSizeGuardrail`: Limits state size
- `MessageCountGuardrail`: Limits message count
- `ContentModerationGuardrail`: Content filtering

**Usage:**
```go
chain := NewGuardrailChain("input-validation", GuardrailTypeInput, true).
    Add(RequiredKeysGuardrail("prompt", "user_id")).
    Add(MaxStateSizeGuardrail("size-limit", 1024*1024)).
    Add(ContentModerationGuardrail("content-check", prohibitedWords))
```

### RunContext with Dependency Injection

Type-safe dependency injection for agents:

```go
type RunContext[D any] struct {
    context.Context
    
    Deps      D
    RunID     string
    Retry     int
    StartTime time.Time
    State     *State
    EmitEvent func(Event)
}
```

**Example Dependencies:**
```go
type LLMDeps struct {
    Provider llm.Provider
    Logger   *slog.Logger
    Tracer   trace.Tracer
}

type DatabaseDeps struct {
    DB    *sql.DB
    Cache *redis.Client
}

// Usage
deps := LLMDeps{Provider: provider, Logger: logger}
runCtx := NewRunContext(ctx, deps, state)

// Access dependencies
response, err := runCtx.Deps.Provider.Complete(ctx, request)
```

### FunctionalEventStream

Functional programming patterns for event processing:

```go
type FunctionalEventStream interface {
    // Core operations
    Filter(predicate EventPredicate) FunctionalEventStream
    Map(transform EventTransform) FunctionalEventStream
    Reduce(reducer EventReducer, initial interface{}) interface{}
    
    // Stream control
    Take(n int) FunctionalEventStream
    TakeUntil(predicate EventPredicate) FunctionalEventStream
    Timeout(duration time.Duration) FunctionalEventStream
    
    // Consumption
    ForEach(handler EventHandler) error
    Collect() ([]Event, error)
    First() (Event, error)
}
```

**Usage:**
```go
stream := NewFunctionalEventStream(ctx, eventChan).
    Filter(ByType(EventToolCall, EventToolResult)).
    Filter(ByAgent("data-processor")).
    Map(WithMetadata("processed", true)).
    Take(100).
    Timeout(5 * time.Minute)

events, err := stream.Collect()
```

### StateValidator

Comprehensive state validation system:

```go
type StateValidator interface {
    Validate(state *State) error
}
```

**Built-in Validators:**
- `RequiredKeysValidator`: Ensure keys exist
- `SchemaValidator`: JSON schema validation
- `TypeValidator`: Type checking
- `RangeValidator`: Numeric ranges
- `RegexValidator`: Pattern matching
- `LengthValidator`: String/slice length
- `EnumValidator`: Allowed values

**Composition:**
```go
validator := CompositeValidator(
    RequiredKeysValidator("name", "email"),
    RegexValidator("email", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
    LengthValidator("name", 1, 100),
)
```

### StateTransforms

Pre-built state transformation functions:

```go
type StateTransform func(ctx context.Context, state *State) (*State, error)
```

**Available Transforms:**
- `FilterTransform`: Remove keys by pattern
- `MapTransform`: Apply function to values
- `PrefixKeysTransform`: Add prefix to keys
- `SelectKeysTransform`: Keep specific keys
- `RenameKeysTransform`: Rename keys
- `ValidateTransform`: Ensure validity
- `ChainTransforms`: Sequential transforms
- `FlattenTransform`: Flatten nested structures

**Usage:**
```go
transform := ChainTransforms(
    SelectKeysTransform("user", "prompt", "config"),
    NormalizeKeysTransform(),
    ValidateTransform(validator, schema),
    PrefixKeysTransform("input_"),
)

newState, err := transform(ctx, state)
```

### TracingHook

OpenTelemetry-compatible distributed tracing:

```go
type TracingHook interface {
    BeforeRun(ctx context.Context, agent BaseAgent, state *State) (context.Context, error)
    AfterRun(ctx context.Context, agent BaseAgent, state *State, result *State, err error) error
}
```

**Components:**
- `TracingHook`: Agent lifecycle tracing
- `ToolCallTracingHook`: Tool execution tracing
- `EventTracingHook`: Event dispatch tracing
- `CompositeTracingHook`: Combined tracing

**Usage:**
```go
tracer := otel.Tracer("my-agent")
hook := NewTracingHook("my-agent", tracer)

// Integrate with agent lifecycle
ctx, err = hook.BeforeRun(ctx, agent, state)
result, err := agent.Run(ctx, state)
err = hook.AfterRun(ctx, agent, state, result, err)
```

## Component Integration

### Agent Lifecycle Integration

Components integrate at different lifecycle points:

```go
func (a *Agent) Run(ctx context.Context, state *State) (*State, error) {
    // 1. Guardrails: Input validation
    if err := a.inputGuardrails.Validate(ctx, state); err != nil {
        return nil, err
    }
    
    // 2. Tracing: Start span
    ctx, err := a.tracingHook.BeforeRun(ctx, a, state)
    if err != nil {
        return nil, err
    }
    
    // 3. RunContext: Dependency injection
    runCtx := NewRunContext(ctx, a.deps, state)
    
    // 4. State transforms: Input processing
    processedState, err := a.inputTransform(runCtx, state)
    if err != nil {
        return nil, err
    }
    
    // 5. Core execution
    result, err := a.execute(runCtx, processedState)
    
    // 6. Guardrails: Output validation
    if err == nil && a.outputGuardrails != nil {
        err = a.outputGuardrails.Validate(ctx, result)
    }
    
    // 7. Tracing: Complete span
    _ = a.tracingHook.AfterRun(ctx, a, state, result, err)
    
    return result, err
}
```

### Event Flow Integration

Events flow through the system with processing capabilities:

```go
// Event emission from agent
agent.EmitEvent(EventAgentStart, startData)

// Event stream processing
stream := NewFunctionalEventStream(ctx, agent.Events()).
    Filter(And(ByAgent("data-processor"), IsError)).
    Map(WithMetadata("retry_candidate", true)).
    ForEach(errorHandler)
```

## Usage Examples

### Complete Agent with All Components

```go
// Define dependencies
type MyAgentDeps struct {
    LLM    llm.Provider
    DB     *sql.DB
    Logger *slog.Logger
}

// Create agent with all components
agent := NewBaseAgent("data-processor", "Processes data", AgentTypeLLM).
    WithDependencies(MyAgentDeps{
        LLM:    llmProvider,
        DB:     database,
        Logger: logger,
    }).
    WithInputGuardrails(
        NewGuardrailChain("input", GuardrailTypeInput, true).
            Add(RequiredKeysGuardrail("data", "schema")).
            Add(MaxStateSizeGuardrail("size", 1024*1024)),
    ).
    WithOutputGuardrails(
        NewGuardrailChain("output", GuardrailTypeOutput, false).
            Add(RequiredKeysGuardrail("result")).
            Add(SchemaValidator(outputSchema)),
    ).
    WithInputTransform(
        ChainTransforms(
            ValidateTransform(inputValidator, inputSchema),
            NormalizeKeysTransform(),
            SelectKeysTransform("data", "schema", "config"),
        ),
    ).
    WithHandoff("next-step", 
        NewHandoffBuilder("summarizer", "summary-agent").
            WithInputFilter(LimitMessagesTransform(10)).
            Build(),
    ).
    WithTracing(NewTracingHook("data-processor", tracer))

// Execute
state := NewState()
state.Set("data", inputData)
state.Set("schema", dataSchema)

result, err := agent.Run(ctx, state)
```

### Event Stream Monitoring

```go
// Monitor agent execution
eventStream := NewFunctionalEventStream(ctx, agent.Events()).
    Filter(Or(IsError, ByType(EventProgress))).
    Map(WithTimestamp).
    TakeUntil(IsComplete).
    Timeout(5 * time.Minute)

go func() {
    err := eventStream.ForEach(EventHandlerFunc(func(e Event) error {
        switch e.Type {
        case EventAgentError:
            logger.Error("Agent error", "error", e.Error, "agent", e.AgentName)
        case EventProgress:
            if data, ok := e.Data.(ProgressEventData); ok {
                logger.Info("Progress", "current", data.Current, "total", data.Total)
            }
        }
        return nil
    }))
    if err != nil {
        logger.Error("Event stream error", "error", err)
    }
}()
```

## Testing Strategy

### Unit Testing

Each component includes comprehensive unit tests:

```go
func TestHandoffBuilder(t *testing.T) {
    handoff := NewHandoffBuilder("test", "target").
        WithDescription("Test handoff").
        WithInputFilter(func(s *State) *State {
            return s.Clone()
        }).
        Build()
    
    assert.Equal(t, "test", handoff.Name())
    assert.Equal(t, "target", handoff.TargetAgent())
}

func TestGuardrailChain(t *testing.T) {
    chain := NewGuardrailChain("test", GuardrailTypeInput, true).
        Add(RequiredKeysGuardrail("key1")).
        Add(MaxStateSizeGuardrail("size", 100))
    
    state := NewState()
    state.Set("key1", "value")
    
    err := chain.Validate(context.Background(), state)
    assert.NoError(t, err)
}
```

### Integration Testing

Test component interactions:

```go
func TestAgentWithComponents(t *testing.T) {
    // Create mock dependencies
    deps := TestDeps{MockLLM: mockProvider}
    
    // Create agent with components
    agent := NewTestAgent().
        WithDependencies(deps).
        WithGuardrails(testGuardrails).
        WithTracing(mockTracer)
    
    // Test execution
    state := NewState()
    state.Set("input", "test")
    
    result, err := agent.Run(context.Background(), state)
    assert.NoError(t, err)
    assert.NotNil(t, result)
    
    // Verify tracing
    assert.Equal(t, 1, mockTracer.SpanCount())
}
```

### Performance Testing

Benchmark critical paths:

```go
func BenchmarkStateTransforms(b *testing.B) {
    state := createLargeState()
    transform := ChainTransforms(
        SelectKeysTransform("key1", "key2"),
        NormalizeKeysTransform(),
        ValidateTransform(validator, schema),
    )
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := transform(context.Background(), state)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Future Phases

### Phase 2: LLM Agent Migration (Completed)

Successfully leveraged the enhanced components:

- **RunContext**: Injected LLM providers and configuration
- **Guardrails**: Validate prompts and responses
- **StateTransforms**: Engineer prompts from state
- **TracingHook**: Trace LLM calls and token usage
- **Handoff**: Delegate between specialized LLM agents
- **Hook System**: Full implementation with metrics and logging
- **Provider Parsing**: Ultra-simple agent creation from strings

### Phase 3: Workflow Agents (Completed)

Successfully built on the foundation:

- **Sequential Agent**: Uses handoffs for step delegation with error handling and state passthrough
- **Parallel Agent**: Uses event streams for coordination with configurable merge strategies
- **Conditional Agent**: Uses state validators for decisions with priority-based evaluation  
- **Loop Agent**: Uses state management for iterative processing with count/while/until patterns
- **Complete Implementation**: All four workflow agent types with comprehensive tests and examples
- **Production Ready**: Error handling, timeout support, hook integration, and metadata collection

### Phase 4: Agent-Tool Integration (Next)

The next phase will focus on:

- **AgentTool Wrapper**: Bidirectional conversion between agents and tools
- **Tool Context System**: Enhanced context management for tool execution
- **Built-in Tool Integration**: Ensure all built-in tools work seamlessly with agents

### Phase 5-6: Advanced Features

- **State Persistence**: Checkpoint and resume capabilities
- **Advanced Patterns**: Circuit breakers, retries, timeouts
- **Multi-Agent Coordination**: Complex workflow patterns
- **Agent Discovery**: Registry and service discovery

## Future Use Cases

The agent architecture is designed to enable sophisticated multi-agent applications. Here are some planned use cases:

### 1. Research Assistant Workflow

```go
researcher := workflow.NewSequentialAgent("researcher", "Research a topic").
    AddSubAgent(
        llm.NewLLMAgent("query_analyzer", "Analyze the research query", provider).
            SetSystemPrompt("Extract key topics from the query"),
    ).
    AddSubAgent(
        workflow.NewParallelAgent("data_gatherer", "Gather data from sources").
            AddSubAgent(webSearchAgent).
            AddSubAgent(databaseAgent).
            AddSubAgent(fileSearchAgent),
    ).
    AddSubAgent(
        llm.NewLLMAgent("synthesizer", "Synthesize findings", provider).
            SetSystemPrompt("Combine research findings into a report"),
    )
```

### 2. Code Review System

```go
codeReviewer := workflow.NewParallelAgent("code_reviewer", "Review code changes").
    AddSubAgent(syntaxChecker).
    AddSubAgent(securityScanner).
    AddSubAgent(
        llm.NewLLMAgent("style_checker", "Check code style", provider).
            AddTool(astAnalyzer),
    ).
    WithMergeStrategy(CombineReviewResults)
```

### 3. Interactive Assistant with Agent Tools

```go
assistant := llm.NewLLMAgent("assistant", "Interactive assistant", provider).
    AddTool(calculatorTool).
    AddTool(weatherTool).
    AddTool(
        // Wrap another agent as a tool
        NewAgentTool(codeReviewer).
            WithInputMapping(map[string]string{
                "code": "code_to_review",
            }),
    )
```

### 4. Data Processing Pipeline

```go
dataPipeline := workflow.NewSequentialAgent("data_pipeline", "Process data").
    AddSubAgent(dataValidatorAgent).
    AddSubAgent(
        workflow.NewParallelAgent("processors", "Process in parallel").
            AddSubAgent(dataCleanerAgent).
            AddSubAgent(dataEnricherAgent).
            AddSubAgent(dataTransformerAgent),
    ).
    AddSubAgent(dataOutputAgent).
    WithErrorHandler(dataErrorRecoveryAgent)
```

## Component Summary

| Component | Purpose | Key Features |
|-----------|---------|--------------|
| BaseAgent | Core interface | Hierarchy, lifecycle, type safety |
| State | Data container | Thread-safe, versioned, immutable |
| Event System | Monitoring | Non-blocking, filterable, structured |
| Handoff | Delegation | Input transformation, message filtering |
| Guardrails | Validation | Sync/async, composable, content filtering |
| RunContext | Dependencies | Type-safe injection, generic support |
| EventStream | Processing | Functional ops, stream control, composition |
| StateValidator | Validation | Schema-based, composable, specialized |
| StateTransforms | Transformation | Immutable, chainable, pre-built functions |
| TracingHook | Observability | OpenTelemetry-compatible, comprehensive |

The agent architecture provides a solid foundation for building sophisticated LLM-powered applications while maintaining Go's simplicity and performance characteristics.