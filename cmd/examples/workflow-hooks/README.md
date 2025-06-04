# Workflow Hooks Example

This example demonstrates how to use hooks with workflow agents for monitoring and logging.

## Overview

Hooks provide a way to monitor and instrument workflow execution. This example shows how to:

- Add metrics hooks to track execution statistics
- Add logging hooks to capture workflow events  
- Use hooks with both sequential and parallel workflows
- Collect and display metrics after workflow completion

## Features Demonstrated

1. **Metrics Collection**: Track requests, errors, and execution times
2. **Logging Integration**: Capture workflow events and progress
3. **Hook Composition**: Use multiple hooks simultaneously
4. **Workflow Monitoring**: Monitor both sequential and parallel workflows

## Usage

```bash
# Run the example
go run main.go
```

## Examples

### 1. Sequential Workflow with Hooks
Demonstrates a data processing pipeline with three sequential steps:
- Data Processor (100ms)
- Data Analyzer (150ms) 
- Data Formatter (50ms)

Total expected time: ~300ms + overhead

### 2. Parallel Workflow with Hooks
Demonstrates parallel processing with three concurrent agents:
- Fast Processor (100ms)
- Medium Processor (200ms)
- Slow Processor (300ms)

Total expected time: ~300ms (slowest agent) + overhead

## Code Structure

```go
// Create hooks
metricsHook := core.NewLLMMetricsHook()
loggingHook := core.NewLoggingHook()

// Add hooks to workflow
workflow := workflow.NewSequentialAgent("pipeline").
    WithHook(metricsHook).
    WithHook(loggingHook).
    AddAgent(agent1).
    AddAgent(agent2)

// Run workflow and collect metrics
result, err := workflow.Run(ctx, initialState)
metrics := metricsHook.GetMetrics()
```

## Available Hooks

### LLMMetricsHook
Tracks execution statistics:
- Total requests
- Total errors  
- Total execution time
- Average execution time
- Request/error counts per operation

### LoggingHook
Provides structured logging:
- Agent start/completion events
- Error logging with context
- Execution timing information

## Metrics Output

After workflow completion, you can access:

```go
metrics := metricsHook.GetMetrics()
fmt.Printf("Total requests: %d\n", metrics.TotalRequests)
fmt.Printf("Total errors: %d\n", metrics.TotalErrors) 
fmt.Printf("Average time: %v\n", metrics.AverageExecutionTime())
```

## Next Steps

- See the conditional workflow example for branching logic
- Explore custom hook implementation for specialized monitoring
- Try combining workflows with tool execution for more complex pipelines