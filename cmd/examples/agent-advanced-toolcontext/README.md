# Advanced ToolContext Example

This example demonstrates the advanced features of ToolContext when tools are called by LLM agents. It showcases how tools can leverage the rich context provided by the framework for enhanced functionality.

## Overview

ToolContext provides tools with access to:
- **Event Emission**: Tools can emit progress, debug, info, warning, error, and custom events
- **Progress Reporting**: Long-running tools can report their progress
- **Retry Information**: Tools can detect retry attempts and adjust behavior
- **State Access**: Tools can read agent state and execution context
- **Agent Information**: Tools know which agent called them and the run ID

## Features Demonstrated

### 1. Progress Reporting (`progress`)
Shows how tools can report progress during long-running operations:
```go
toolCtx.Events.EmitProgress(current, total, "Processing...")
```

### 2. Event Emission (`events`)
Demonstrates various event types:
- Message events for general updates
- Debug events for detailed information
- Warning/Error events for issues
- Info events for important notifications
- Custom events for metrics or specialized data

### 3. Retry Handling (`retry`)
Shows how tools can:
- Detect retry attempts via `toolCtx.Retry`
- Adjust behavior based on retry count
- Report retry status through events

### 4. State Access (`state`)
Demonstrates accessing:
- Agent information (name, ID, type)
- Run context (run ID)
- State values through StateReader interface
- Using state to customize behavior

### 5. All Features (`all`)
A comprehensive example combining all features in a realistic scenario.

## Running the Examples

```bash
# Build the example
go build -o agent-advanced-toolcontext .

# Run individual examples
./agent-advanced-toolcontext progress
./agent-advanced-toolcontext events
./agent-advanced-toolcontext retry
./agent-advanced-toolcontext state
./agent-advanced-toolcontext all

# With custom prompts
./agent-advanced-toolcontext progress "Process 50 items with detailed progress"
./agent-advanced-toolcontext state "What context information is available?"
```

## Key Concepts

### ToolContext Structure
```go
type ToolContext struct {
    Context     context.Context
    State       StateReader    // Read-only access to state
    Agent       BaseAgent      // The calling agent
    RunID       string         // Unique run identifier
    Retry       int           // Retry attempt number (0 = first try)
    Events      EventEmitter  // Event emission interface
}
```

### Event Types
- `EventProgress`: Progress updates with current/total/message
- `EventMessage`: General messages
- `EventDebug`: Debug information
- `EventInfo`: Important information
- `EventWarning`: Warning conditions
- `EventError`: Error conditions
- Custom events: Any string prefixed with "custom:"

### Best Practices

1. **Always Check ToolContext**: Verify the context is a ToolContext before casting
2. **Nil Check Events**: Check if `toolCtx.Events` is nil before emitting
3. **Meaningful Progress**: Report progress at meaningful intervals
4. **Structured Custom Events**: Use structured data for custom events
5. **State is Read-Only**: Tools can read but not modify agent state

## Example Output

### Progress Example
```
[15:04:05.123] [TOOL CALL] Starting 'data_processor'
[15:04:05.124] [PROGRESS] 5/20 - Processing item 5 of 20
[15:04:05.324] [MESSAGE] Reached 25% completion
[15:04:05.524] [PROGRESS] 10/20 - Processing item 10 of 20
[15:04:05.624] [MESSAGE] Reached 50% completion
...
[15:04:06.124] [MESSAGE] Processing complete!
[15:04:06.125] [TOOL RESULT] 'data_processor' completed
  - processed: 20
  - status: completed
```

### Retry Example
```
[15:04:05.123] [TOOL CALL] Starting 'unreliable_service'
[15:04:05.124] [ERROR] service temporarily unavailable
[15:04:06.125] [TOOL CALL] Starting 'unreliable_service'
[15:04:06.125] [INFO] Retry attempt #1
[15:04:06.126] [MESSAGE] Success after 1 retries!
[15:04:06.126] [TOOL RESULT] 'unreliable_service' completed
  - data: Successfully retrieved data
  - retries: 1
  - status: success
```

## Integration with LLM Agents

The examples show how LLM agents can:
1. Call tools that use advanced ToolContext features
2. System prompts guide the LLM on when to use tools
3. Event monitoring provides real-time feedback
4. State sharing enables context-aware tool execution

## Extending the Examples

To create your own tools with advanced features:

1. **Create the tool with ToolContext handling**:
```go
tool := atools.NewTool("my_tool", "description", func(ctx context.Context, params interface{}) (interface{}, error) {
    toolCtx, ok := ctx.(*domain.ToolContext)
    if !ok {
        return nil, fmt.Errorf("context is not a ToolContext")
    }
    
    // Use toolCtx features...
    if toolCtx.Events != nil {
        toolCtx.Events.EmitMessage("Starting work...")
    }
    
    return result, nil
})
```

2. **Add to agent and configure system prompt**:
```go
agent.AddTool(tool)
agent.SetSystemPrompt("You have access to my_tool which...")
```

3. **Enable event monitoring**:
```go
dispatcher := core.NewEventDispatcher(100)
agent.BaseAgentImpl.SetEventDispatcher(dispatcher)
```

## Common Patterns

### Progress for Batch Operations
```go
for i, item := range items {
    if toolCtx.Events != nil {
        toolCtx.Events.EmitProgress(i+1, len(items), fmt.Sprintf("Processing %s", item.Name))
    }
    // Process item...
}
```

### Retry-Aware Behavior
```go
timeout := 5 * time.Second
if toolCtx.Retry > 0 {
    // Increase timeout on retries
    timeout = timeout * time.Duration(toolCtx.Retry+1)
    toolCtx.Events.EmitInfo(fmt.Sprintf("Using %v timeout for retry #%d", timeout, toolCtx.Retry))
}
```

### State-Driven Configuration
```go
config := defaultConfig
if toolCtx.State != nil {
    if userConfig, exists := toolCtx.State.Get("tool_config"); exists {
        config = mergeConfigs(config, userConfig)
    }
}
```

## Next Steps

- Explore creating tools that coordinate through events
- Build tools that adapt behavior based on retry patterns
- Create state-aware tools for personalized experiences
- Implement progress reporting for real long-running operations