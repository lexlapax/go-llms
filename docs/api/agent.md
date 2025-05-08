# Agent Package

The `agent` package provides functionality for building LLM-powered agents that can use tools to interact with external systems. It includes support for tool definition, agent workflows, and monitoring hooks.

## Core Components

### Domain

#### Tool Interface

```go
type Tool interface {
    // Name returns the tool's name
    Name() string
    
    // Description provides information about the tool
    Description() string
    
    // Execute runs the tool with parameters
    Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
    
    // ParameterSchema returns the schema for the tool parameters
    ParameterSchema() interface{}
}
```

The `Tool` interface defines methods for tools that agents can invoke, with support for parameter validation and execution.

#### Hook Interface

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

The `Hook` interface provides callbacks for monitoring agent operations, such as generation and tool execution.

#### RunContext

```go
type RunContext[D any] struct {
    ctx  context.Context
    deps D
}

// NewRunContext creates a new run context
func NewRunContext[D any](ctx context.Context, deps D) *RunContext[D]

// Deps returns the dependencies
func (r *RunContext[D]) Deps() D

// Context returns the context
func (r *RunContext[D]) Context() context.Context
```

The `RunContext` provides a way to carry dependencies through the agent execution flow, with generic support for different dependency types.

## Tools Package

The tools package provides base implementations and utilities for creating and registering tools.

### BaseTool

```go
// Create a new tool
tool := tools.NewTool(
    "weather",
    "Get the weather for a location",
    func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
        location, _ := params["location"].(string)
        // Implementation...
        return map[string]interface{}{
            "temperature": 72.5,
            "condition": "Sunny",
            "location": location,
        }, nil
    },
)
```

The `BaseTool` provides a foundation for tool implementations with support for parameter validation and execution.

### Parameter Schemas

```go
// Define a parameter schema
paramSchema := &schemaDomain.Schema{
    Type: "object",
    Properties: map[string]schemaDomain.Property{
        "location": {
            Type:        "string",
            Description: "The location to get weather for",
        },
    },
    Required: []string{"location"},
}

// Create a tool with a parameter schema
tool := tools.NewToolWithSchema(
    "weather",
    "Get the weather for a location",
    paramSchema,
    func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
        // Implementation...
    },
)
```

Tools can define parameter schemas to validate inputs before execution.

## Workflow Package

The workflow package provides agent implementations and execution flows.

### Agent

```go
// Create an agent
func NewAgent[D any, O any](provider domain.Provider) *Agent[D, O]

// Run the agent
func (a *Agent[D, O]) Run(ctx context.Context, input string, deps D) (O, error)

// Run with schema
func (a *Agent[D, O]) RunWithSchema(ctx context.Context, input string, schema *schemaDomain.Schema, deps D) (O, error)
```

The `Agent` type provides the main functionality for running LLM agents, with support for dependencies and structured outputs.

### Configuration

```go
// Add a tool to the agent
func (a *Agent[D, O]) AddTool(tool domain.Tool) *Agent[D, O]

// Set the system prompt
func (a *Agent[D, O]) SetSystemPrompt(prompt string) *Agent[D, O]

// Add a hook
func (a *Agent[D, O]) AddHook(hook domain.Hook) *Agent[D, O]
```

Agents can be configured with tools, system prompts, and hooks.

### Hooks Implementation

#### LoggingHook

```go
// Create a logging hook
hook := workflow.NewLoggingHook(slog.Default(), workflow.LogLevelDetailed)

// Log levels
const (
    // LogLevelBasic logs basic information
    LogLevelBasic
    // LogLevelDetailed logs detailed information including message content
    LogLevelDetailed
    // LogLevelDebug logs everything including full message content and tool data
    LogLevelDebug
)
```

The `LoggingHook` provides logging for agent operations, with configurable detail levels.

#### MetricsHook

```go
// Create a metrics hook
hook := workflow.NewMetricsHook()

// Get metrics
metrics := hook.GetMetrics()

// Reset metrics
hook.Reset()
```

The `MetricsHook` collects performance metrics for agent operations, such as request counts, tool calls, and response times.

### Tool Executor

```go
// The ToolExecutor is used internally by the agent
// You generally don't need to interact with it directly
executor := workflow.NewToolExecutor(tools)
```

The `ToolExecutor` handles tool execution and parameter validation.

### Message Manager

```go
// The MessageManager is used internally by the agent
// You generally don't need to interact with it directly
manager := workflow.NewMessageManager()
```

The `MessageManager` manages the conversation history for the agent.

## Additional Components

### MultiAgent

```go
// Create a multi-agent
multiAgent := workflow.NewMultiAgent[Deps, Output](providers)
```

The `MultiAgent` allows using multiple agents together, similar to the multi-provider approach.

### CachedAgent

```go
// Create a cached agent
cachedAgent := workflow.NewCachedAgent(agent)
```

The `CachedAgent` provides caching for agent responses to improve performance.

## Example Usage

### Basic Agent with Tools

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/tools"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// Define dependencies
type ExampleDeps struct {
    APIClient *APIClient
}

// Mock API client
type APIClient struct{}

func (c *APIClient) GetData(query string) (string, error) {
    return fmt.Sprintf("Data for query: %s", query), nil
}

func main() {
    // Create a provider
    llmProvider := provider.NewOpenAIProvider("your-api-key", "gpt-4o")
    
    // Create an agent
    agent := workflow.NewAgent[ExampleDeps, string](llmProvider).
        SetSystemPrompt("You are a helpful assistant with access to tools.")
    
    // Add a search tool
    agent.AddTool(tools.NewTool(
        "search",
        "Search for information",
        func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
            query, _ := params["query"].(string)
            
            // Get dependencies from context
            runCtx := ctx.Value("runContext").(*domain.RunContext[ExampleDeps])
            apiClient := runCtx.Deps().APIClient
            
            // Use the API client to get data
            result, err := apiClient.GetData(query)
            if err != nil {
                return nil, err
            }
            
            return result, nil
        },
    ))
    
    // Add a logging hook
    agent.AddHook(workflow.NewLoggingHook(slog.Default(), workflow.LogLevelDetailed))
    
    // Create dependencies
    deps := ExampleDeps{
        APIClient: &APIClient{},
    }
    
    // Run the agent
    result, err := agent.Run(context.Background(), "What can you tell me about Go programming?", deps)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Result: %s\n", result)
}
```

### Agent with Structured Output

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Define a structured output type
type RecipeOutput struct {
    Title       string   `json:"title"`
    Ingredients []string `json:"ingredients"`
    Steps       []string `json:"steps"`
    PrepTime    int      `json:"prep_time"`
}

func main() {
    // Create a provider
    llmProvider := provider.NewOpenAIProvider("your-api-key", "gpt-4o")
    
    // Create an agent with structured output
    agent := workflow.NewAgent[struct{}, RecipeOutput](llmProvider).
        SetSystemPrompt("You are a helpful cooking assistant.")
    
    // Define output schema
    schema := &domain.Schema{
        Type: "object",
        Properties: map[string]domain.Property{
            "title": {Type: "string", Description: "The recipe title"},
            "ingredients": {
                Type: "array",
                Items: &domain.Property{Type: "string"},
                Description: "List of ingredients",
            },
            "steps": {
                Type: "array",
                Items: &domain.Property{Type: "string"},
                Description: "Preparation steps",
            },
            "prep_time": {Type: "integer", Description: "Preparation time in minutes"},
        },
        Required: []string{"title", "ingredients", "steps"},
    }
    
    // Run the agent with schema
    recipe, err := agent.RunWithSchema(
        context.Background(),
        "Give me a simple recipe for chocolate chip cookies.",
        schema,
        struct{}{},
    )
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    // Use the structured output
    fmt.Printf("Recipe: %s\n\n", recipe.Title)
    
    fmt.Println("Ingredients:")
    for _, ingredient := range recipe.Ingredients {
        fmt.Printf("- %s\n", ingredient)
    }
    
    fmt.Println("\nSteps:")
    for i, step := range recipe.Steps {
        fmt.Printf("%d. %s\n", i+1, step)
    }
    
    fmt.Printf("\nPreparation time: %d minutes\n", recipe.PrepTime)
}
```

### Agent with Metrics

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
    // Create a provider
    llmProvider := provider.NewOpenAIProvider("your-api-key", "gpt-4o")
    
    // Create an agent
    agent := workflow.NewAgent[struct{}, string](llmProvider).
        SetSystemPrompt("You are a helpful assistant.")
    
    // Add a metrics hook
    metricsHook := workflow.NewMetricsHook()
    agent.AddHook(metricsHook)
    
    // Prepare context with metrics tracking
    ctx := workflow.WithMetrics(context.Background())
    
    // Run the agent
    result, err := agent.Run(ctx, "What is the capital of France?", struct{}{})
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Result: %s\n\n", result)
    
    // Get metrics
    metrics := metricsHook.GetMetrics()
    
    fmt.Println("Metrics:")
    fmt.Printf("Requests: %d\n", metrics.Requests)
    fmt.Printf("Tool calls: %d\n", metrics.ToolCalls)
    fmt.Printf("Errors: %d\n", metrics.ErrorCount)
    fmt.Printf("Total tokens: %d\n", metrics.TotalTokens)
    fmt.Printf("Average generation time: %.2fms\n", metrics.AverageGenTimeMs)
}
```
