# Metrics and Monitoring Example

This example demonstrates how to use hooks for monitoring and collecting metrics in Go-LLMs agents.

## Overview

The example showcases:

1. **Metrics Collection**: Using the `MetricsHook` to gather statistics about agent operations
2. **Detailed Logging**: Using the `LoggingHook` for real-time visibility into agent actions
3. **Custom Tools**: Creating tools with configurable performance characteristics for testing
4. **Combined Hooks**: Using multiple hooks simultaneously for comprehensive monitoring

## Key Components

### Hooks

The example uses two hook implementations:

1. **MetricsHook**: Collects quantitative data about agent operations:
   - Request counts
   - Tool call counts
   - Error counts
   - Token estimates
   - Response generation times
   - Tool execution statistics

2. **LoggingHook**: Provides real-time qualitative information about agent operations:
   - Generation start/completion events
   - Tool execution events
   - Error events
   - Content details (configurable verbosity)

### Simulated Tools

The example includes test tools with configurable characteristics:

- **Fast Tool**: Responds quickly with minimal latency
- **Slow Tool**: Simulates a high-latency external API
- **Unreliable Tool**: Simulates occasional failures (configurable failure rate)
- **Calculator Tool**: A real functional tool for calculations

## How Hooks Work

Hooks in Go-LLMs provide callbacks at key points in the agent workflow:

```
┌─────────────────┐  BeforeGenerate   ┌──────────────┐  BeforeToolCall   ┌──────────────┐
│                 │ ────────────────► │              │ ────────────────► │              │
│  Agent starts   │                   │ LLM generates│                   │ Tool executes│
│                 │ ◄──────────────── │              │ ◄──────────────── │              │
└─────────────────┘  AfterGenerate    └──────────────┘  AfterToolCall    └──────────────┘
```

Each hook implements callbacks for these four events:

1. `BeforeGenerate`: Called before sending a request to the LLM
2. `AfterGenerate`: Called after receiving a response from the LLM (or error)
3. `BeforeToolCall`: Called before executing a tool
4. `AfterToolCall`: Called after a tool completes (or errors)

## Running the Example

Build and run the example:

```bash
make example EXAMPLE=metrics
./bin/metrics
```

The output will show:

1. Detailed logs of agent operations in real-time
2. A summary of metrics collected during agent operations
3. A demonstration of metrics reset functionality

## Implementing Your Own Hooks

You can implement custom hooks by satisfying the `domain.Hook` interface:

```go
type Hook interface {
    // BeforeGenerate is called before generating a response
    BeforeGenerate(ctx context.Context, messages []domain.Message)

    // AfterGenerate is called after generating a response
    AfterGenerate(ctx context.Context, response domain.Response, err error)

    // BeforeToolCall is called before executing a tool
    BeforeToolCall(ctx context.Context, tool string, params map[string]interface{})

    // AfterToolCall is called after executing a tool
    AfterToolCall(ctx context.Context, tool string, result interface{}, err error)
}
```

Common hook use cases:

1. **Telemetry**: Send metrics to monitoring systems
2. **Logging**: Record agent activities for debugging
3. **Cost Tracking**: Monitor token usage and API costs
4. **Performance Analysis**: Track response times and bottlenecks
5. **Auditing**: Record all agent actions for compliance
6. **Rate Limiting**: Enforce usage limits
7. **Caching**: Record patterns for potential caching optimization

## Example Metrics Output

The metrics report provides insights like:

```
📊 Agent Metrics Report
====================
Total Requests:      5
Total Tool Calls:    7
Error Count:         1
Estimated Tokens:    1529
Avg Generation Time: 35.42 ms

🔧 Tool Statistics
-----------------
{
  "calculator": {
    "Calls": 1,
    "AverageTimeMs": 2.0,
    "FastestCallMs": 2.0,
    "SlowestCallMs": 2.0
  },
  "fastTool": {
    "Calls": 2,
    "AverageTimeMs": 50.5,
    "FastestCallMs": 50.0,
    "SlowestCallMs": 51.0
  },
  "slowTool": {
    "Calls": 2,
    "AverageTimeMs": 200.5,
    "FastestCallMs": 200.0,
    "SlowestCallMs": 201.0
  },
  "unreliableTool": {
    "Calls": 2,
    "AverageTimeMs": 100.0,
    "FastestCallMs": 100.0,
    "SlowestCallMs": 100.0
  }
}
```

## Best Practices

1. **Minimal Impact**: Keep hooks lightweight to avoid affecting agent performance
2. **Error Handling**: Hooks should never panic or throw errors
3. **Concurrency Safety**: Ensure thread safety for metrics collection
4. **Context Usage**: Use context values for tracking operation timing
5. **Reset Capability**: Provide ways to reset metrics for interval measurements
6. **Selective Detail**: Configure verbosity levels for logging hooks