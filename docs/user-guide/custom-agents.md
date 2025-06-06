# Custom Agents

> **[Documentation Home](/REFERENCE.md) / [User Guide](/docs/user-guide/) / Custom Agents**

Custom agents provide ultimate flexibility for creating arbitrary orchestration logic beyond the predefined workflow patterns. They allow you to implement complex state management, external integrations, and unique workflow patterns by directly implementing the `BaseAgent` interface.

## Table of Contents

1. [Overview](#overview)
2. [When to Use Custom Agents](#when-to-use-custom-agents)
3. [Architecture](#architecture)
4. [Implementation Patterns](#implementation-patterns)
5. [Examples](#examples)
6. [Best Practices](#best-practices)
7. [Integration with Workflow Agents](#integration-with-workflow-agents)

## Overview

Custom agents inherit from `BaseAgentImpl` and implement their own `Run` method to provide completely custom orchestration logic. Unlike workflow agents that follow predefined patterns, custom agents can implement any logic using standard Go language constructs.

```go
type CustomAgent struct {
    *core.BaseAgentImpl
    // Custom fields
    subAgents map[string]domain.BaseAgent
    config    CustomConfig
}

func (c *CustomAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
    // Your custom orchestration logic here
    return result, nil
}
```

## When to Use Custom Agents

### Choose Custom Agents When You Need:

- **Complex Conditional Logic**: Multi-branched decision trees that don't fit conditional workflow patterns
- **Dynamic Agent Selection**: Runtime selection of agents based on complex criteria
- **External System Integration**: Database operations, API calls, file system interactions
- **Stateful Processing**: Complex state management across multiple steps
- **Unique Orchestration Patterns**: Workflow patterns not covered by standard agents

### Choose Workflow Agents When You Need:

- **Sequential Processing**: Use `SequentialAgent`
- **Parallel Execution**: Use `ParallelAgent` 
- **Simple Branching**: Use `ConditionalAgent`
- **Iteration Patterns**: Use `LoopAgent`

### Choose LLM Agents When You Need:

- **Direct LLM Interaction**: Simple prompt → response patterns
- **Tool Calling**: LLM agents with function calling capabilities

## Architecture

### Core Components

```go
// Custom agent implements BaseAgent interface
type MyCustomAgent struct {
    *core.BaseAgentImpl
    
    // Sub-agents for delegation
    llmAgent     domain.BaseAgent
    dbAgent      domain.BaseAgent
    validatorAgent domain.BaseAgent
    
    // Configuration
    config       MyConfig
    
    // State management
    stateManager domain.StateManager
}
```

### Key Patterns

1. **Sub-Agent Orchestration**: Coordinate multiple specialized agents
2. **State Management**: Use the state system for data flow between steps
3. **Event Emission**: Emit events for monitoring and debugging
4. **Hook Integration**: Support hooks for metrics and logging
5. **Error Handling**: Implement robust error handling and recovery

## Implementation Patterns

### 1. Basic Custom Agent

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

type CalculatorAgent struct {
    *core.BaseAgentImpl
}

func NewCalculatorAgent(name string) *CalculatorAgent {
    return &CalculatorAgent{
        BaseAgentImpl: core.NewBaseAgent(name, "Calculator agent", domain.AgentTypeCustom),
    }
}

func (c *CalculatorAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
    // Extract operation and operands from state
    operation, _ := input.Get("operation")
    operand1, _ := input.Get("operand1")
    operand2, _ := input.Get("operand2")
    
    // Perform calculation
    var result float64
    switch operation.(string) {
    case "add":
        result = operand1.(float64) + operand2.(float64)
    case "multiply":
        result = operand1.(float64) * operand2.(float64)
    default:
        return nil, fmt.Errorf("unsupported operation: %s", operation)
    }
    
    // Create result state
    resultState := input.Clone()
    resultState.Set("result", result)
    
    return resultState, nil
}
```

### 2. Multi-Agent Orchestration

```go
type StoryAgent struct {
    *core.BaseAgentImpl
    writer    domain.BaseAgent
    reviewer  domain.BaseAgent
    editor    domain.BaseAgent
}

func (s *StoryAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
    currentState := input.Clone()
    
    // Step 1: Generate initial story
    writerState := domain.NewState()
    writerState.Set("prompt", "Write a short story based on: " + input.GetString("topic"))
    
    storyResult, err := s.writer.Run(ctx, writerState)
    if err != nil {
        return nil, fmt.Errorf("story generation failed: %w", err)
    }
    
    story := storyResult.GetString("response")
    currentState.Set("story", story)
    
    // Step 2: Review tone and quality
    reviewState := domain.NewState()
    reviewState.Set("prompt", "Analyze the tone of this story: " + story)
    
    reviewResult, err := s.reviewer.Run(ctx, reviewState)
    if err != nil {
        return nil, fmt.Errorf("story review failed: %w", err)
    }
    
    tone := reviewResult.GetString("response")
    currentState.Set("tone_analysis", tone)
    
    // Step 3: Conditional editing based on tone
    if strings.Contains(strings.ToLower(tone), "negative") {
        editState := domain.NewState()
        editState.Set("prompt", "Make this story more positive: " + story)
        
        editResult, err := s.editor.Run(ctx, editState)
        if err != nil {
            return nil, fmt.Errorf("story editing failed: %w", err)
        }
        
        currentState.Set("story", editResult.GetString("response"))
        currentState.Set("edited", true)
    }
    
    return currentState, nil
}
```

### 3. External Integration Pattern

```go
type DataPipelineAgent struct {
    *core.BaseAgentImpl
    dbClient    DatabaseClient
    validator   domain.BaseAgent
    processor   domain.BaseAgent
}

func (d *DataPipelineAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
    // Step 1: Fetch data from database
    query := input.GetString("query")
    data, err := d.dbClient.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("database query failed: %w", err)
    }
    
    // Step 2: Validate data using sub-agent
    validationState := domain.NewState()
    validationState.Set("data", data)
    validationState.Set("schema", input.Get("validation_schema"))
    
    validResult, err := d.validator.Run(ctx, validationState)
    if err != nil {
        return nil, fmt.Errorf("data validation failed: %w", err)
    }
    
    if !validResult.GetBool("valid") {
        return nil, fmt.Errorf("data validation failed: %s", validResult.GetString("errors"))
    }
    
    // Step 3: Process data using sub-agent
    processState := domain.NewState()
    processState.Set("data", data)
    processState.Set("operations", input.Get("processing_operations"))
    
    processResult, err := d.processor.Run(ctx, processState)
    if err != nil {
        return nil, fmt.Errorf("data processing failed: %w", err)
    }
    
    // Return results
    result := input.Clone()
    result.Set("processed_data", processResult.Get("result"))
    result.Set("record_count", len(data))
    
    return result, nil
}
```

## Examples

The Go-LLMs library includes comprehensive custom agent examples:

- **[Story Agent](../../cmd/examples/agent-custom-story/README.md)**: Multi-LLM coordination with conditional logic
- **[Data Pipeline Agent](../../cmd/examples/agent-custom-data-pipeline/README.md)**: Database + processing + validation workflow
- **[API Orchestrator Agent](../../cmd/examples/agent-custom-api-orchestrator/README.md)**: Multiple API calls with retries
- **[Calculator Agent](../../cmd/examples/agent-custom-calculator/README.md)**: Pure computational logic

## Best Practices

### 1. State Management

```go
// Good: Use state for data flow
func (c *CustomAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
    // Clone input to avoid mutations
    currentState := input.Clone()
    
    // Add data progressively
    currentState.Set("step1_result", result1)
    currentState.Set("step2_result", result2)
    
    return currentState, nil
}

// Bad: Don't mutate input state directly
func (c *CustomAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
    input.Set("result", "modified") // Mutates caller's state
    return input, nil
}
```

### 2. Error Handling

```go
func (c *CustomAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
    // Use context for timeout handling
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Wrap errors with context
    result, err := c.subAgent.Run(ctx, someState)
    if err != nil {
        return nil, fmt.Errorf("sub-agent execution failed at step X: %w", err)
    }
    
    return result, nil
}
```

### 3. Event Emission

```go
func (c *CustomAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
    // Emit progress events
    c.EmitEvent(domain.EventAgentStart, map[string]interface{}{
        "agent": c.Name(),
        "input_keys": len(input.Keys()),
    })
    
    // ... processing ...
    
    c.EmitEvent(domain.EventProgress, map[string]interface{}{
        "step": "data_processing",
        "progress": 0.5,
    })
    
    // ... more processing ...
    
    c.EmitEvent(domain.EventAgentComplete, map[string]interface{}{
        "agent": c.Name(),
        "duration": time.Since(start),
    })
    
    return result, nil
}
```

### 4. Hook Integration

```go
// Ensure your custom agent supports hooks
func (c *CustomAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
    // Execute before hooks
    if err := c.BeforeRun(ctx, input); err != nil {
        return nil, err
    }
    
    // Your custom logic here
    result, err := c.customLogic(ctx, input)
    
    // Execute after hooks
    if afterErr := c.AfterRun(ctx, input, result, err); afterErr != nil {
        // Log the after-run error but return original error
        return result, err // or afterErr if you prefer
    }
    
    return result, err
}
```

## Integration with Workflow Agents

Custom agents can be seamlessly integrated with workflow agents:

```go
// Use custom agent as a step in sequential workflow
sequential := core.NewSequentialAgent("data-processing", "Sequential data processing agent")
sequential.AddAgent(customDataExtractor)
sequential.AddAgent(customDataValidator)
sequential.AddAgent(customDataProcessor)

// Use custom agent in parallel workflow
parallel := core.NewParallelAgent("analysis", "Parallel analysis agent")
parallel.AddAgent(customTextAnalyzer)
parallel.AddAgent(customImageAnalyzer)
parallel.AddAgent(customDataAnalyzer)

// Use custom agent in conditional workflow
conditional := core.NewConditionalAgent("smart-processor", "Smart content processor")
conditional.AddBranch("text", func(state *domain.State) bool {
    dataType, _ := state.Get("content_type")
    return dataType == "text"
}, customTextProcessor)
conditional.AddBranch("image", func(state *domain.State) bool {
    dataType, _ := state.Get("content_type")
    return dataType == "image"  
}, customImageProcessor)
```

## Custom Hooks for Monitoring

Custom agents can implement custom hooks for monitoring, logging, and metrics collection:

```go
// CustomHook implements monitoring for agent events
type CustomHook struct {
    name      string
    startTime time.Time
    events    []string
}

// NewCustomHook creates a new monitoring hook
func NewCustomHook(name string) *CustomHook {
    return &CustomHook{
        name:      name,
        startTime: time.Now(),
        events:    make([]string, 0),
    }
}

// BeforeGenerate is called before LLM generation
func (h *CustomHook) BeforeGenerate(ctx context.Context, messages []llmDomain.Message) {
    h.events = append(h.events, fmt.Sprintf("[%s] BeforeGenerate: %d messages", h.name, len(messages)))
}

// AfterGenerate is called after LLM generation
func (h *CustomHook) AfterGenerate(ctx context.Context, response llmDomain.Response, err error) {
    if err != nil {
        h.events = append(h.events, fmt.Sprintf("[%s] AfterGenerate Error: %v", h.name, err))
    } else {
        h.events = append(h.events, fmt.Sprintf("[%s] AfterGenerate: Response received", h.name))
    }
}

// BeforeToolCall is called before tool execution
func (h *CustomHook) BeforeToolCall(ctx context.Context, toolName string, params map[string]interface{}) {
    paramJSON, _ := json.Marshal(params)
    h.events = append(h.events, fmt.Sprintf("[%s] BeforeToolCall: %s with params: %s", h.name, toolName, paramJSON))
}

// AfterToolCall is called after tool execution
func (h *CustomHook) AfterToolCall(ctx context.Context, toolName string, result interface{}, err error) {
    if err != nil {
        h.events = append(h.events, fmt.Sprintf("[%s] AfterToolCall: %s error: %v", h.name, toolName, err))
    } else {
        resultJSON, _ := json.Marshal(result)
        h.events = append(h.events, fmt.Sprintf("[%s] AfterToolCall: %s result: %s", h.name, toolName, resultJSON))
    }
}

// GetEvents returns all collected events
func (h *CustomHook) GetEvents() []string {
    return h.events
}

// PrintSummary prints a summary of all events
func (h *CustomHook) PrintSummary() {
    fmt.Printf("\nCustom Hook (%s) Summary:\n", h.name)
    fmt.Printf("Total events: %d\n", len(h.events))
    fmt.Printf("Total duration: %v\n", time.Since(h.startTime))
}
```

### Using Custom Hooks

```go
// Create custom agent with monitoring
type MonitoredAgent struct {
    *core.BaseAgentImpl
    hook *CustomHook
}

func NewMonitoredAgent(name string) *MonitoredAgent {
    return &MonitoredAgent{
        BaseAgentImpl: core.NewBaseAgent(name, "Monitored agent", domain.AgentTypeCustom),
        hook:         NewCustomHook(name),
    }
}

func (m *MonitoredAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
    // Hook will automatically track LLM and tool events
    // if the agent uses LLM providers or tools
    
    // Custom logic here
    result := input.Clone()
    result.Set("processed", true)
    
    // Print monitoring summary
    m.hook.PrintSummary()
    
    return result, nil
}
```

## Validation

Always implement validation for your custom agents:

```go
func (c *CustomAgent) Validate() error {
    // Call base validation
    if err := c.BaseAgentImpl.Validate(); err != nil {
        return err
    }
    
    // Custom validation
    if c.subAgent == nil {
        return fmt.Errorf("sub-agent cannot be nil")
    }
    
    if c.config.Timeout <= 0 {
        return fmt.Errorf("timeout must be positive")
    }
    
    return nil
}
```

## Multi-Agent Patterns (New in Phase 5)

The Phase 5 multi-agent enhancement provides powerful new patterns for custom agents:

### Automatic Sub-Agent Tools

When creating LLM agents with sub-agents, they automatically become available as tools:

```go
// Create custom agent with automatic tool registration
coordinator, err := core.NewLLMAgentWithSubAgentsFromString(
    "coordinator",
    "openai/gpt-4",
    customAgent1,
    customAgent2,
    customAgent3,
)

// All sub-agents are now available as tools to the LLM
```

### Building Hierarchical Custom Agents

Create multi-level hierarchies of custom agents:

```go
// Custom research agent
type ResearchAgent struct {
    *core.BaseAgentImpl
}

func (r *ResearchAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
    // Custom research logic
    return state, nil
}

// Team lead managing multiple custom agents
teamLead, _ := core.NewLLMAgentWithSubAgents(
    "teamLead",
    provider,
    &ResearchAgent{BaseAgentImpl: core.NewBaseAgent("researcher1", "Web researcher", domain.AgentTypeCustom)},
    &ResearchAgent{BaseAgentImpl: core.NewBaseAgent("researcher2", "Academic researcher", domain.AgentTypeCustom)},
)
```

### Shared State with Custom Agents

Enable state sharing between your custom agents:

```go
// Custom agent that uses shared state
type StateAwareAgent struct {
    *core.BaseAgentImpl
}

func (s *StateAwareAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
    // Access shared state from parent
    if sharedValue, ok := state.Get("shared_context"); ok {
        fmt.Printf("Using shared context: %v\n", sharedValue)
    }
    
    // Process with awareness of parent state
    result := state.Clone()
    result.Set("processed_with_context", true)
    
    return result, nil
}

// Enable shared state for sub-agents
mainAgent.EnableSharedState(true)
mainAgent.ConfigureStateInheritance(true, true, true)
```

### Direct Transfer Between Custom Agents

Use the TransferTo method for easy delegation:

```go
// Transfer control to a custom sub-agent
result, err := coordinator.TransferTo(
    ctx,
    "customProcessor",
    "Process this data with custom logic",
    map[string]interface{}{
        "data": rawData,
        "options": processingOptions,
    },
)
```

### Custom Agents as Sub-Agents

Any custom agent can be used as a sub-agent:

```go
// Mix custom agents with standard agents
mainAgent, _ := core.NewLLMAgentWithSubAgents(
    "main",
    provider,
    myCustomAgent,      // Your custom agent
    standardLLMAgent,   // Standard LLM agent
    workflowAgent,      // Workflow agent
)
```

Custom agents provide the ultimate flexibility for implementing sophisticated agent workflows while maintaining compatibility with the broader Go-LLMs agent ecosystem.