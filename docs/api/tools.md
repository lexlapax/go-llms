# Tools API Reference

The tools package (`pkg/agent/tools`) provides the foundation for creating tools that LLMs can invoke. It offers a flexible system for wrapping functions as tools with rich metadata, performance optimizations, and seamless agent integration.

## Overview

Tools in go-llms are functions that agents can invoke to perform specific tasks. The package provides:
- Simple function wrapping with automatic parameter handling
- Rich metadata support for LLM guidance
- Performance optimizations with object pooling
- Bidirectional agent-tool conversion
- MCP (Model Context Protocol) compatibility

## Core Types

### Tool

The main tool implementation that wraps functions for agent use.

```go
type Tool struct {
    // Contains wrapped function, metadata, and performance optimizations
}
```

### ToolBuilder

Fluent interface for creating tools with comprehensive metadata.

```go
type ToolBuilder struct {
    // Builder for creating tools with rich metadata
}
```

## Creating Tools

### Basic Tool Creation

```go
import "github.com/lexlapax/go-llms/pkg/agent/tools"

// Simple function wrapping
fn := func(a, b int) int {
    return a + b
}

tool := tools.NewTool(
    "add",
    "Adds two numbers",
    fn,
    schema, // Parameter schema
)
```

### Using ToolBuilder

The ToolBuilder provides a fluent interface for creating tools with rich metadata:

```go
tool := tools.NewToolBuilder("calculator", "Performs mathematical calculations").
    WithFunction(calculateFn).
    WithParameterSchema(paramSchema).
    WithOutputSchema(outputSchema).
    WithUsageInstructions(`
        Use this tool for mathematical calculations.
        Supports basic arithmetic operations.
    `).
    WithExamples([]string{
        `{"operation": "add", "a": 5, "b": 3}`,
        `{"operation": "multiply", "a": 4, "b": 7}`,
    }).
    WithCategory("math").
    WithTags([]string{"arithmetic", "calculation"}).
    WithVersion("1.0.0").
    WithBehavior(
        true,  // deterministic
        false, // destructive
        false, // requiresConfirmation
        "low", // latency
    ).
    Build()
```

### Context Support

Tools can accept different types of context for advanced functionality:

```go
// With ToolContext for state and event access
fn := func(ctx *domain.ToolContext, params struct{Text string}) (string, error) {
    // Emit progress events
    ctx.Events.EmitProgress(1, 10, "Processing text")
    
    // Access state
    config, _ := ctx.State.Get("config")
    
    // Access agent info
    agentName := ctx.AgentInfo.Name
    
    return processText(params.Text), nil
}

// With standard context
fn := func(ctx context.Context, input string) (string, error) {
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
        return process(input), nil
    }
}
```

## Tool Metadata

### Enhanced Metadata Options

```go
tool := builder.
    // Usage guidance for LLMs
    WithUsageInstructions("Detailed instructions for when and how to use this tool").
    
    // Example inputs/outputs
    WithExamples([]string{
        `Input: {"query": "weather in NYC"} Output: {"temp": 72, "conditions": "sunny"}`,
    }).
    
    // Constraints and limitations
    WithConstraints([]string{
        "Maximum 1000 characters input",
        "Rate limited to 10 requests per minute",
    }).
    
    // Error handling guidance
    WithErrorGuidance(map[string]string{
        "rate_limit": "Wait 60 seconds before retrying",
        "invalid_input": "Check input format matches schema",
    }).
    
    // Categorization
    WithCategory("web").
    WithTags([]string{"api", "search", "external"}).
    
    // Behavioral hints
    WithBehavior(
        false,  // non-deterministic (API calls may vary)
        false,  // non-destructive
        true,   // requires confirmation for sensitive operations
        "high", // high latency due to network calls
    ).
    Build()
```

## Agent-Tool Conversion

### Converting Agents to Tools

```go
// Basic conversion
agentTool := tools.NewAgentTool(myAgent)

// With custom mapping
agentTool := tools.NewAgentTool(myAgent).
    WithStateMapper(func(params map[string]interface{}) domain.State {
        state := domain.NewState()
        state.Set("input", params["text"])
        state.Set("options", params["options"])
        return state
    }).
    WithResultMapper(func(state domain.State) (interface{}, error) {
        return map[string]interface{}{
            "result": state.Get("output"),
            "metadata": state.Get("metadata"),
        }, nil
    })
```

### Converting Tools to Agents

```go
// Basic conversion
toolAgent := tools.NewToolAgent(myTool)

// With custom mapping
toolAgent := tools.NewToolAgent(myTool).
    WithParamMapper(func(state domain.State) map[string]interface{} {
        return map[string]interface{}{
            "query": state.Get("input.query"),
            "limit": state.Get("options.limit"),
        }
    }).
    WithStateUpdater(func(state domain.State, result interface{}) domain.State {
        state.Set("result", result)
        state.Set("processed", true)
        return state
    })
```

## Utility Functions

### State and Parameter Mapping

```go
// Create field-based state mapper
mapper := tools.CreateStateMapper(map[string]string{
    "input_text": "text",      // params["input_text"] -> state["text"]
    "max_length": "options.max_length",
})

// Create result extractor
resultMapper := tools.CreateResultMapper("output", "summary", "metadata")

// Create parameter mapper with paths
paramMapper := tools.CreatePathMapper(map[string]string{
    "query": "search.query",    // state["search"]["query"] -> params["query"]
    "limit": "options.limit",
})

// Type conversion mapper
conversionMapper := tools.CreateTypeConversionMapper(
    map[string]string{"id": "userID"},
    map[string]string{"id": "string"}, // Convert to string
)
```

### Tool Chain Creation

```go
// Create a chain of agents as a single tool
toolChain := tools.CreateToolChainFromAgents(
    "research-chain",
    "Performs research by chaining search, analysis, and summary",
    searchAgent,
    analyzeAgent,
    summaryAgent,
)
```

### LLM Agent Wrapping

```go
// Wrap an LLM agent as a tool with sensible defaults
llmTool := tools.WrapLLMAgentAsTool(
    llmAgent,
    "text-processor",
    "Processes text using LLM capabilities",
)
```

## Performance Considerations

The tools package includes several performance optimizations:

1. **Object Pooling**: Reuses argument slices to reduce allocations
2. **Reflection Caching**: Pre-computes reflection data for faster execution
3. **Type Conversion**: Optimized paths for common type conversions
4. **Thread Safety**: Safe for concurrent use

```go
// Tools automatically use sync.Pool for performance
// No special configuration needed
tool := tools.NewTool("example", "description", fn, schema)
// Pool usage is handled internally
```

## Error Handling

Tools provide structured error information:

```go
result, err := tool.Execute(params)
if err != nil {
    // Error includes context about parameter validation,
    // execution failures, or schema mismatches
    log.Printf("Tool execution failed: %v", err)
}
```

## Examples

### Calculator Tool

```go
type CalcParams struct {
    Operation string  `json:"operation"`
    A         float64 `json:"a"`
    B         float64 `json:"b"`
}

calcFn := func(params CalcParams) (float64, error) {
    switch params.Operation {
    case "add":
        return params.A + params.B, nil
    case "multiply":
        return params.A * params.B, nil
    default:
        return 0, fmt.Errorf("unknown operation: %s", params.Operation)
    }
}

calculator := tools.NewToolBuilder("calculator", "Basic calculator").
    WithFunction(calcFn).
    WithParameterSchema(schema).
    WithExamples([]string{
        `{"operation": "add", "a": 5, "b": 3} -> 8`,
    }).
    Build()
```

### File Reader Tool with Context

```go
readFile := func(ctx *domain.ToolContext, params struct{Path string}) (string, error) {
    // Check permissions from state
    perms, _ := ctx.State.Get("permissions")
    if !canRead(perms, params.Path) {
        return "", fmt.Errorf("permission denied")
    }
    
    // Emit progress
    ctx.Events.EmitProgress(1, 2, "Reading file")
    
    content, err := os.ReadFile(params.Path)
    if err != nil {
        return "", err
    }
    
    ctx.Events.EmitProgress(2, 2, "Complete")
    return string(content), nil
}

fileReader := tools.NewToolBuilder("read-file", "Reads file content").
    WithFunction(readFile).
    WithConstraints([]string{
        "Only reads text files",
        "Maximum 10MB file size",
    }).
    Build()
```

## Tool Discovery (v0.3.4+)

The enhanced discovery system enables metadata-first tool exploration without requiring imports:

### Discovery API

```go
// Create discovery instance
discovery := tools.NewDiscovery()

// List all available tools without imports
tools := discovery.ListTools()
for _, info := range tools {
    fmt.Printf("%s (%s): %s\n", info.Name, info.Category, info.Description)
}

// Search tools by keyword
searchResults := discovery.SearchTools("json")

// Filter by category
webTools := discovery.ListByCategory("web")

// Get tool schema without loading
schema, err := discovery.GetToolSchema("calculator")
if err == nil {
    fmt.Printf("Parameters: %v\n", schema.Parameters)
    fmt.Printf("Output: %v\n", schema.Output)
}

// Get tool examples
examples, err := discovery.GetToolExamples("web_search")

// Create tool on demand
calculator, err := discovery.CreateTool("calculator")
if err == nil {
    result, _ := calculator.Execute(ctx, params)
}

// Batch creation
tools, err := discovery.CreateTools("web_fetch", "json_process")
```

### ToolMetadata Structure

```go
type ToolMetadata struct {
    Name            string
    Description     string
    Category        string
    Tags            []string
    ParameterSchema json.RawMessage  // JSON schema
    OutputSchema    json.RawMessage  // JSON schema
    Examples        []Example
    UsageHint       string
}
```

### Use Cases

- **Scripting Engines**: Dynamic tool discovery for Lua/JavaScript bridges
- **CLI Applications**: Interactive tool exploration without imports
- **Documentation**: Auto-generated tool catalogs
- **Reduced Binary Size**: Only import tools you actually use

## See Also

- [Agent API Reference](agent.md) - Core agent concepts
- [Workflow API Reference](workflows.md) - Composing tools into workflows
- [Built-in Tools](builtins.md) - Pre-built tools for common tasks
- [Tool Development Guide](../user-guide/tool-development.md) - Detailed guide for creating tools