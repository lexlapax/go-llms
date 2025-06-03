# Agent Architecture Restructuring Plan

## Overview

This document outlines a comprehensive plan to restructure our agent and workflow interfaces based on Google's Agent Development Kit (ADK) architecture. The goal is to create a more elegant, composable, and flexible agent system that clearly separates concerns between different agent types while maintaining a unified interface.

## Current Architecture Analysis

### Current State
- Single `Agent` interface mixing LLM and workflow concerns
- `DefaultAgent` and `MultiAgent` are both LLM-focused implementations
- No clear workflow agent abstractions
- Tools are separate entities that cannot easily become agents
- No agent hierarchy or composition model
- Implicit state management through context

### Limitations
1. Cannot create deterministic workflow agents without LLM dependency
2. Agents cannot be composed hierarchically
3. No clear separation between LLM agents and orchestration agents
4. Tools and agents are completely separate concepts
5. Limited reusability and composability

## Proposed Architecture

### Core Design Principles
1. **Unified Base Agent**: All agents derive from a common `BaseAgent` interface
2. **Separation of Concerns**: Clear distinction between LLM agents and workflow agents
3. **Composability**: Agents can contain sub-agents and be wrapped as tools
4. **State Management**: Explicit state passing between agents
5. **Lifecycle Management**: Standardized execution lifecycle with hooks

### Interface Hierarchy

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

## Detailed Interface Definitions

### 1. BaseAgent Interface

```go
// pkg/agent/domain/base_agent.go

type BaseAgent interface {
    // Core identification
    Name() string
    Description() string
    
    // Hierarchy management
    Parent() BaseAgent
    SetParent(parent BaseAgent)
    SubAgents() []BaseAgent
    AddSubAgent(agent BaseAgent) error
    FindAgent(name string) BaseAgent
    
    // Execution methods
    Run(ctx context.Context, state *State) (*State, error)
    RunAsync(ctx context.Context, state *State) (<-chan Event, error)
    
    // Lifecycle hooks
    BeforeRun(ctx context.Context, state *State) error
    AfterRun(ctx context.Context, state *State, err error) error
    
    // State management
    InputSchema() *schema.Schema
    OutputSchema() *schema.Schema
    
    // Configuration
    WithConfig(config AgentConfig) BaseAgent
}

// State represents the shared state between agents
type State struct {
    // Key-value store for agent communication
    values map[string]interface{}
    // Artifacts like files, images, etc.
    artifacts map[string]Artifact
    // Conversation history if applicable
    messages []Message
}

// Event represents an event during agent execution
type Event struct {
    Type      EventType
    AgentName string
    Timestamp time.Time
    Data      interface{}
}

type EventType string

const (
    EventStart       EventType = "start"
    EventProgress    EventType = "progress"
    EventToolCall    EventType = "tool_call"
    EventComplete    EventType = "complete"
    EventError       EventType = "error"
    EventStateUpdate EventType = "state_update"
)
```

### 2. LLMAgent Implementation

```go
// pkg/agent/llm/llm_agent.go

type LLMAgent struct {
    baseAgent
    
    // LLM-specific fields
    provider      llm.Provider
    model         string
    systemPrompt  string
    tools         map[string]Tool
    
    // Configuration
    temperature   float64
    maxTokens     int
    
    // Callbacks
    beforeGenerate []GenerateCallback
    afterGenerate  []GenerateCallback
    beforeTool     []ToolCallback
    afterTool      []ToolCallback
}

// LLM-specific methods
func (a *LLMAgent) AddTool(tool Tool) *LLMAgent
func (a *LLMAgent) SetSystemPrompt(prompt string) *LLMAgent
func (a *LLMAgent) WithModel(model string) *LLMAgent
func (a *LLMAgent) WithTemperature(temp float64) *LLMAgent
```

### 3. Workflow Agents

```go
// pkg/agent/workflow/sequential_agent.go

type SequentialAgent struct {
    baseAgent
    
    // Configuration
    stopOnError bool
    passState   bool // Whether to pass state between agents
}

func (a *SequentialAgent) Run(ctx context.Context, state *State) (*State, error) {
    // Execute each sub-agent in order
    currentState := state
    for _, agent := range a.SubAgents() {
        newState, err := agent.Run(ctx, currentState)
        if err != nil && a.stopOnError {
            return currentState, err
        }
        if a.passState {
            currentState = newState
        }
    }
    return currentState, nil
}
```

```go
// pkg/agent/workflow/parallel_agent.go

type ParallelAgent struct {
    baseAgent
    
    // Configuration
    waitForAll    bool
    mergeStrategy MergeStrategy
}

type MergeStrategy func(states []*State) *State
```

### 4. Agent as Tool

```go
// pkg/agent/tools/agent_tool.go

type AgentTool struct {
    baseTool
    agent BaseAgent
    
    // Configuration
    inputMapping  map[string]string // Map tool params to agent state
    outputMapping map[string]string // Map agent state to tool output
}

func NewAgentTool(agent BaseAgent) *AgentTool {
    return &AgentTool{
        baseTool: baseTool{
            name:        agent.Name() + "_tool",
            description: agent.Description(),
        },
        agent: agent,
    }
}

func (t *AgentTool) Execute(ctx context.Context, params interface{}) (interface{}, error) {
    // Convert params to State
    state := t.paramsToState(params)
    
    // Run the agent
    resultState, err := t.agent.Run(ctx, state)
    if err != nil {
        return nil, err
    }
    
    // Convert state back to tool output
    return t.stateToOutput(resultState), nil
}
```

### 5. Enhanced Tool Interface

```go
// pkg/agent/domain/tool.go

type Tool interface {
    // Core identification
    Name() string
    Description() string
    
    // Execution
    Execute(ctx context.Context, params interface{}) (interface{}, error)
    ExecuteAsync(ctx context.Context, params interface{}) (<-chan interface{}, error)
    
    // Schema
    ParameterSchema() *schema.Schema
    OutputSchema() *schema.Schema
    
    // Configuration
    IsLongRunning() bool
    WithTimeout(timeout time.Duration) Tool
    
    // Tool context for advanced features
    WithContext(tc ToolContext) Tool
}

// ToolContext provides additional context for tool execution
type ToolContext interface {
    // Access to the calling agent
    Agent() BaseAgent
    
    // Access to shared state
    State() *State
    
    // Ability to emit events
    EmitEvent(event Event)
}
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1-2)
1. Define new interfaces in `pkg/agent/domain/`
   - `base_agent.go`
   - `state.go`
   - `events.go`
   - `tool.go` (enhanced)
2. Implement base agent functionality
   - `pkg/agent/core/base_agent.go`
   - State management utilities
   - Event system

### Phase 2: LLM Agent Migration (Week 2-3)
1. Implement new LLMAgent based on current DefaultAgent
2. Migrate tool integration to new interface
3. Add state management capabilities
4. Implement agent hierarchy support
5. Remove old superfluos code, examples and tests

### Phase 3: Workflow Agents (Week 3-4)
1. Implement workflow agent base
2. Create SequentialAgent
3. Create ParallelAgent
4. Create ConditionalAgent
5. Create LoopAgent

### Phase 4: Agent-Tool Integration (Week 4)
1. Implement AgentTool wrapper
2. Create tool context system
3. Add bidirectional agent-tool conversion utilities

### Phase 5: Advanced Features (Week 5)
1. State persistence and serialization
2. Agent discovery and registry
3. Advanced merge strategies for parallel agents
4. Streaming support for long-running agents

### Phase 6: Migration and Testing (Week 5-6)
1. Create migration guide
2. Update all examples
3. Comprehensive testing
4. Performance benchmarking

## Migration Strategy

### Backward Compatibility Approach
1. Do not keep existing interfaces in deprecated package
2. do not Provide adapters from old to new interfaces
3. remove old code
5. Clean up documentation for changed code under docs/

### Example Migration

```go
// Old code
agent := workflow.NewAgent(provider).
    AddTool(myTool).
    SetSystemPrompt("You are a helpful assistant")

// New code
agent := llm.NewLLMAgent("assistant", "A helpful assistant", provider).
    AddTool(myTool).
    SetSystemPrompt("You are a helpful assistant")

// Or compose agents
workflow := workflow.NewSequentialAgent("workflow", "Process user request").
    AddSubAgent(preprocessor).
    AddSubAgent(agent).
    AddSubAgent(postprocessor)
```

## Benefits of New Architecture

1. **Clear Separation of Concerns**
   - LLM logic separate from orchestration logic
   - Tools and agents have clear boundaries

2. **Enhanced Composability**
   - Agents can be nested arbitrarily
   - Any agent can become a tool
   - Workflow patterns are first-class citizens

3. **Better State Management**
   - Explicit state passing
   - State isolation between agents
   - Clear data flow

4. **Improved Testing**
   - Mock workflow agents for testing
   - Test agents in isolation
   - Clear interfaces for mocking

5. **Future Extensibility**
   - Easy to add new agent types
   - Support for different execution modes
   - Plugin architecture for custom agents

## Example Use Cases

### 1. Research Assistant
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

### 3. Interactive Assistant with Tools
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

## Conclusion

This architecture provides a clean, extensible foundation for building complex agent systems. By separating concerns and providing clear abstractions, we enable developers to create sophisticated multi-agent applications while maintaining code clarity and testability.

The phased implementation approach ensures we can deliver value incrementally while maintaining system stability. The migration strategy protects existing users while providing a clear path to the new architecture.