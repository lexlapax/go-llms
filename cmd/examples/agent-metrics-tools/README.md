# Agent Metrics with Tools Example

This example demonstrates how to collect detailed metrics from an LLM agent that uses tools, including real-time monitoring and comprehensive statistics.

## Overview

The example showcases:

1. **Real LLM Providers**: Works with OpenAI, Anthropic, Gemini, or falls back to mock provider
2. **Built-in Tools**: Uses production-ready tools from the registry (calculator, web_fetch, file_list, datetime_now)
3. **Metrics Collection**: Using the `MetricsHook` to gather statistics about agent operations
4. **Detailed Logging**: Using the `LoggingHook` for real-time visibility into agent actions
5. **LLMAgent Pattern**: Uses the new `core.NewLLMAgent` with LLMDeps structure
6. **Combined Hooks**: Using multiple hooks simultaneously for comprehensive monitoring

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

### Built-in Tools

The example uses production tools from the built-in registry:

- **Calculator**: Performs mathematical operations (add, subtract, multiply, divide)
- **Web Fetch**: Retrieves content from web URLs
- **File List**: Lists files in directories with pattern matching
- **DateTime Now**: Gets the current date and time

All tools are registered and available through:
```go
tools.MustGetTool("calculator")
```

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

The example automatically detects available LLM providers:

```bash
# With OpenAI
export OPENAI_API_KEY=your-key-here
go run main.go

# With Anthropic
export ANTHROPIC_API_KEY=your-key-here
go run main.go

# With Gemini
export GEMINI_API_KEY=your-key-here
go run main.go

# Or build and run
make example EXAMPLE=agent-metrics-tools
./bin/agent-metrics-tools
```

The output will show:

1. Which LLM provider is being used
2. Detailed logs of agent operations in real-time
3. Actual tool calls being made with real calculations
4. A summary of metrics collected during agent operations
5. A demonstration of metrics reset functionality

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

With a real LLM provider, the metrics report shows actual tool execution:

```
Using openai provider with model gpt-4o-mini

📊 Agent Metrics Report
====================
Total Requests:      10
Total Tool Calls:    8
Error Count:         0
Estimated Tokens:    3063
Avg Generation Time: 953.90 ms

🔧 Tool Statistics
-----------------
{
  "calculator": {
    "Calls": 4,
    "AverageTimeMs": 11,
    "FastestCallMs": 10,
    "SlowestCallMs": 12
  },
  "datetime_now": {
    "Calls": 2,
    "AverageTimeMs": 5,
    "FastestCallMs": 4,
    "SlowestCallMs": 6
  },
  "file_list": {
    "Calls": 2,
    "AverageTimeMs": 15,
    "FastestCallMs": 14,
    "SlowestCallMs": 16
  }
}
```

The agent successfully performs operations like:
- Calculate 123 + 456 = 579
- Calculate 50 * 20 = 1000
- Show current date and time
- List files in directories

## Best Practices

1. **Minimal Impact**: Keep hooks lightweight to avoid affecting agent performance
2. **Error Handling**: Hooks should never panic or throw errors
3. **Concurrency Safety**: Ensure thread safety for metrics collection
4. **Context Usage**: Use context values for tracking operation timing
5. **Reset Capability**: Provide ways to reset metrics for interval measurements
6. **Selective Detail**: Configure verbosity levels for logging hooks