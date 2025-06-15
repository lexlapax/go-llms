# Tool Development Guide

> **[Documentation Home](/docs/README.md) / [Technical Documentation](/docs/technical/README.md) / Tool Development**

This guide covers the internal architecture and development patterns for creating tools in go-llms. It explains the Tool interface, ToolBuilder pattern, registration system, and best practices for tool implementation.

## Overview

Tools in go-llms are functions that agents can invoke to perform specific operations. The tool system has evolved to include rich metadata, schema validation, and MCP (Model Context Protocol) compatibility.

## Tool Interface

The enhanced Tool interface provides comprehensive functionality:

```go
type Tool interface {
    // Core functionality
    Name() string                                                      // Unique identifier
    Description() string                                               // Brief description
    Execute(ctx *ToolContext, params interface{}) (interface{}, error) // Execute the tool

    // Schema definitions
    ParameterSchema() *domain.Schema // JSON Schema for input parameters
    OutputSchema() *domain.Schema    // JSON Schema for output structure

    // LLM guidance
    UsageInstructions() string        // Detailed instructions on when and how to use
    Examples() []ToolExample          // Concrete examples showing tool usage
    Constraints() []string            // Limitations and constraints
    ErrorGuidance() map[string]string // Map of error types to helpful guidance

    // Metadata
    Category() string // Category for grouping (e.g., "math", "web", "file")
    Tags() []string   // Tags for discovery and filtering
    Version() string  // Tool version for compatibility tracking

    // Behavioral hints
    IsDeterministic() bool      // Same input always produces same output
    IsDestructive() bool        // Tool modifies state or has side effects
    RequiresConfirmation() bool // User confirmation needed before execution
    EstimatedLatency() string   // Expected execution time: "fast", "medium", "slow"

    // MCP compatibility
    ToMCPDefinition() MCPToolDefinition // Export tool definition in MCP format
}
```

## ToolBuilder Pattern

The ToolBuilder provides a fluent interface for creating tools:

### Basic Usage

```go
tool := tools.NewToolBuilder("my_tool", "Brief description").
    WithFunction(myExecuteFunc).
    WithParameterSchema(paramSchema).
    Build()
```

### Complete Example

```go
tool := tools.NewToolBuilder("weather", "Get current weather information").
    WithCategory("web").
    WithTags("weather", "api", "temperature").
    WithVersion("1.0.0").
    WithFunction(weatherExecute).
    WithParameterSchema(paramSchema).
    WithOutputSchema(outputSchema).
    WithUsageInstructions(`Use this tool to get current weather information.
The tool returns temperature, conditions, humidity, and wind speed.`).
    WithConstraints(
        "Requires weather_api_key in agent state",
        "City names must be in English",
        "Temperature units limited to celsius or fahrenheit",
    ).
    WithExamples(
        tools.Example{
            Name:        "Basic weather query",
            Description: "Get weather for a city",
            Input:       map[string]interface{}{"city": "London"},
            Output:      map[string]interface{}{"temperature": 15.5, "description": "Cloudy"},
        },
    ).
    WithErrorGuidance(map[string]string{
        "city is required": "Provide a city name in the 'city' parameter",
        "api key not found": "Set weather_api_key in agent state before using this tool",
    }).
    WithDeterministic(false).        // Weather changes
    WithDestructive(false).          // Read-only operation
    WithRequiresConfirmation(false). // Safe to call
    WithEstimatedLatency("medium").  // API call required
    Build()
```

## Function Signatures

The ToolBuilder supports various function signatures through reflection:

### Supported Signatures

```go
// 1. Basic function with struct parameter
type Params struct {
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
}
func myFunc(params Params) (Result, error)

// 2. Function with context
func myFunc(ctx context.Context, params Params) (Result, error)

// 3. Function with ToolContext
func myFunc(ctx *domain.ToolContext, params Params) (*Result, error)

// 4. Interface{} parameters and results
func myFunc(ctx context.Context, params interface{}) (interface{}, error)
```

### Function Wrapping

The ToolBuilder automatically wraps your function to match the Tool interface:

```go
// Your function
func calculate(params CalcParams) (float64, error) {
    switch params.Operation {
    case "add":
        return params.A + params.B, nil
    case "multiply":
        return params.A * params.B, nil
    default:
        return 0, fmt.Errorf("unknown operation: %s", params.Operation)
    }
}

// Wrapped by ToolBuilder
tool := tools.NewToolBuilder("calculator", "Performs calculations").
    WithFunction(calculate).
    WithParameterSchema(calcSchema).
    Build()
```

## Schema Definition

Tools use JSON Schema for parameter and output validation:

### Parameter Schema

```go
paramSchema := &sdomain.Schema{
    Type: "object",
    Properties: map[string]sdomain.Property{
        "operation": {
            Type:        "string",
            Description: "Mathematical operation",
            Enum:        []interface{}{"add", "subtract", "multiply", "divide"},
        },
        "a": {
            Type:        "number",
            Description: "First operand",
        },
        "b": {
            Type:        "number", 
            Description: "Second operand",
        },
    },
    Required: []string{"operation", "a", "b"},
}
```

### Output Schema

```go
outputSchema := &sdomain.Schema{
    Type: "object",
    Properties: map[string]sdomain.Property{
        "result": {
            Type:        "number",
            Description: "Calculation result",
        },
        "formula": {
            Type:        "string",
            Description: "Formula used",
        },
    },
    Required: []string{"result"},
}
```

## Tool Context

Tools receive a ToolContext containing execution context:

```go
type ToolContext struct {
    Context context.Context    // Standard Go context
    State   StateReader       // Read-only state access
    Agent   *AgentInfo       // Information about calling agent
    RunID   string           // Unique execution identifier
    Events  ToolEventEmitter // Optional event emitter
}
```

### Accessing State

```go
func myToolExecute(ctx *domain.ToolContext, params MyParams) (*MyResult, error) {
    // Read configuration from state
    if apiKey, ok := ctx.State.Get("api_key"); ok {
        // Use API key
    }
    
    // Access previous results
    if lastResult, ok := ctx.State.Get("last_result"); ok {
        // Reference previous execution
    }
    
    return result, nil
}
```

### Emitting Events

```go
// Emit start event
ctx.EmitEvent(domain.EventTypeToolStart, map[string]interface{}{
    "tool": "my_tool",
    "params": params,
})

// Emit progress
ctx.EmitProgress(50, 100, "Processing halfway complete")

// Emit completion
ctx.EmitEvent(domain.EventTypeToolComplete, result)
```

## Tool Registration

Tools are registered in a global registry for discovery:

### Registration Pattern

```go
func init() {
    tools.MustRegisterTool("my_tool", CreateMyTool(), tools.ToolMetadata{
        Category:    "utilities",
        Tags:        []string{"processing", "data"},
        Description: "Processes data in various ways",
        Version:     "1.0.0",
        RequiredPermissions: []string{"data:read"},
        ResourceUsage: tools.ResourceInfo{
            Memory:      "medium",
            Network:     false,
            FileSystem:  false,
            Concurrency: true,
        },
    })
}
```

### Registry Usage

```go
// Get a tool by name
tool, ok := tools.GetTool("calculator")

// List tools by category
mathTools := tools.Tools.ListByCategory("math")

// Search tools
jsonTools := tools.Tools.Search("json")

// Check if tool exists
if tools.Tools.Has("web_search") {
    // Tool is available
}
```

## Tool Discovery (v0.3.4+)

The enhanced tool discovery system provides metadata-first access to tools without requiring imports:

### Metadata-First Discovery

Tools expose their metadata separately from implementation, enabling discovery without importing tool packages:

```go
// Get all available tool metadata without imports
metadata := tools.GetToolMetadata()
for name, meta := range metadata {
    fmt.Printf("%s: %s (category: %s)\n", name, meta.Description, meta.Category)
}

// Search tools by criteria
discovery := tools.NewDiscovery()
webTools := discovery.SearchTools("web") // Search by name, description, or tags
mathTools := discovery.ListByCategory("math")
```

### Tool Schema Access

Access detailed tool schemas for validation and documentation:

```go
// Get parameter schema for a specific tool
schema, err := discovery.GetToolSchema("calculator")
if err == nil {
    // Use schema for validation or documentation
    fmt.Printf("Parameters: %v\n", schema.Parameters)
    fmt.Printf("Output: %v\n", schema.Output)
}

// Get examples for a tool
examples, err := discovery.GetToolExamples("web_search")
for _, example := range examples {
    fmt.Printf("Example: %s\n", example.Description)
    fmt.Printf("Input: %v\n", example.Input)
    fmt.Printf("Expected: %v\n", example.Output)
}
```

### Lazy Tool Loading

Create tools only when needed:

```go
// List available tools without loading them
availableTools := discovery.ListTools()

// Load specific tools on demand
tool, err := discovery.CreateTool("calculator")
if err == nil {
    result, _ := tool.Execute(ctx, params)
}

// Load multiple tools
tools, err := discovery.CreateTools("calculator", "web_search", "file_read")
```

### Scripting Integration

The discovery system is designed for scripting engines and dynamic environments:

```go
// Example integration for scripting bridges
type ScriptingBridge struct {
    discovery *tools.ToolDiscovery
}

func (b *ScriptingBridge) ListTools() []ToolInfo {
    return b.discovery.ListTools()
}

func (b *ScriptingBridge) GetToolHelp(name string) (string, error) {
    schema, err := b.discovery.GetToolSchema(name)
    if err != nil {
        return "", err
    }
    return schema.GenerateHelp(), nil
}

func (b *ScriptingBridge) ExecuteTool(name string, params map[string]interface{}) (interface{}, error) {
    tool, err := b.discovery.CreateTool(name)
    if err != nil {
        return nil, err
    }
    return tool.Execute(ctx, params)
}
```

## Error Handling

Tools should return meaningful errors with context:

### Error Patterns

```go
// Domain-specific errors
type InvalidParameterError struct {
    Parameter string
    Reason    string
}

func (e InvalidParameterError) Error() string {
    return fmt.Sprintf("invalid parameter %s: %s", e.Parameter, e.Reason)
}

// In tool execution
func execute(ctx *ToolContext, params Params) (*Result, error) {
    if params.Value < 0 {
        return nil, InvalidParameterError{
            Parameter: "value",
            Reason:    "must be non-negative",
        }
    }
    
    // Wrap external errors with context
    data, err := fetchData(params.URL)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch data from %s: %w", params.URL, err)
    }
    
    return processData(data)
}
```

### Error Guidance

Provide guidance for common errors:

```go
.WithErrorGuidance(map[string]string{
    "connection timeout": "Check network connectivity and try again",
    "invalid api key": "Ensure API key is set in agent state",
    "rate limit exceeded": "Wait a few minutes before retrying",
})
```

## Performance Considerations

### Object Pooling

For frequently allocated objects:

```go
var resultPool = sync.Pool{
    New: func() interface{} {
        return &MyResult{}
    },
}

func execute(ctx *ToolContext, params Params) (*Result, error) {
    result := resultPool.Get().(*MyResult)
    defer func() {
        // Clear result before returning to pool
        *result = MyResult{}
        resultPool.Put(result)
    }()
    
    // Use result
    return result, nil
}
```

### Concurrent Execution

Tools should be safe for concurrent use:

```go
type MyTool struct {
    // Avoid shared mutable state
    config Config // Immutable after creation
    
    // Use sync primitives if needed
    mu    sync.RWMutex
    cache map[string]*Result
}

func (t *MyTool) Execute(ctx *ToolContext, params interface{}) (interface{}, error) {
    // Safe concurrent execution
    t.mu.RLock()
    if cached, ok := t.cache[key]; ok {
        t.mu.RUnlock()
        return cached, nil
    }
    t.mu.RUnlock()
    
    // Compute result...
}
```

## Testing Tools

### Unit Testing

```go
func TestCalculatorTool(t *testing.T) {
    tool := CreateCalculatorTool()
    
    // Test parameter validation
    t.Run("InvalidOperation", func(t *testing.T) {
        params := CalcParams{
            Operation: "invalid",
            A: 5,
            B: 3,
        }
        
        ctx := &domain.ToolContext{
            Context: context.Background(),
            State:   domain.NewState(),
        }
        
        _, err := tool.Execute(ctx, params)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "unknown operation")
    })
    
    // Test successful execution
    t.Run("Addition", func(t *testing.T) {
        params := CalcParams{
            Operation: "add",
            A: 5,
            B: 3,
        }
        
        result, err := tool.Execute(ctx, params)
        assert.NoError(t, err)
        assert.Equal(t, 8.0, result)
    })
}
```

### Integration Testing

```go
func TestToolWithAgent(t *testing.T) {
    // Create agent with tool
    agent := core.NewLLMAgent("test", "Test agent", provider)
    agent.AddTool(CreateCalculatorTool())
    
    // Test tool invocation
    state := domain.NewState()
    state.Set("task", "Calculate 15 + 27")
    
    result, err := agent.Run(context.Background(), state)
    assert.NoError(t, err)
    assert.Contains(t, result.Get("response"), "42")
}
```

## Best Practices

### 1. Use Structured Parameters

```go
// Good: Structured parameters
type SearchParams struct {
    Query    string   `json:"query"`
    Filters  []string `json:"filters,omitempty"`
    MaxItems int      `json:"max_items,omitempty"`
}

// Avoid: Multiple parameters
func search(query string, filters []string, maxItems int) // Don't do this
```

### 2. Provide Rich Metadata

```go
// Good: Comprehensive metadata
.WithUsageInstructions("Detailed usage instructions...")
.WithExamples(multiple, realistic, examples)
.WithConstraints("Clear limitations...")
.WithErrorGuidance(commonErrors)

// Avoid: Minimal metadata
.Build() // Don't skip metadata
```

### 3. Handle Errors Gracefully

```go
// Good: Contextual errors
return nil, fmt.Errorf("failed to parse JSON at line %d: %w", line, err)

// Avoid: Generic errors
return nil, errors.New("parsing failed")
```

### 4. Make Tools Deterministic When Possible

```go
// For deterministic tools
.WithDeterministic(true)

// Document non-determinism
.WithConstraints("Results may vary due to external API changes")
```

### 5. Validate Early

```go
func execute(ctx *ToolContext, params interface{}) (interface{}, error) {
    // Validate parameters first
    p, ok := params.(MyParams)
    if !ok {
        return nil, fmt.Errorf("invalid parameter type: expected MyParams, got %T", params)
    }
    
    // Additional validation
    if err := p.Validate(); err != nil {
        return nil, fmt.Errorf("parameter validation failed: %w", err)
    }
    
    // Proceed with execution
}
```

## MCP Integration

Tools can export to Model Context Protocol format:

```go
func (t *MyTool) ToMCPDefinition() MCPToolDefinition {
    return MCPToolDefinition{
        Name:        t.Name(),
        Description: t.Description(),
        InputSchema: t.ParameterSchema().ToMCP(),
        Category:    t.Category(),
        Examples:    convertExamplesToMCP(t.Examples()),
    }
}
```

## Advanced Patterns

### Streaming Results

For tools that produce incremental results:

```go
type StreamingTool struct {
    // ... tool implementation
}

func (t *StreamingTool) ExecuteStream(ctx *ToolContext, params interface{}) (<-chan Result, error) {
    results := make(chan Result)
    
    go func() {
        defer close(results)
        
        // Stream results
        for i := 0; i < totalItems; i++ {
            select {
            case <-ctx.Context.Done():
                return
            case results <- processItem(i):
                ctx.EmitProgress(i+1, totalItems, fmt.Sprintf("Processed %d/%d", i+1, totalItems))
            }
        }
    }()
    
    return results, nil
}
```

### Caching Results

For expensive operations:

```go
type CachedTool struct {
    cache *lru.Cache
    ttl   time.Duration
}

func (t *CachedTool) Execute(ctx *ToolContext, params interface{}) (interface{}, error) {
    key := computeCacheKey(params)
    
    // Check cache
    if cached, ok := t.cache.Get(key); ok {
        if entry := cached.(*cacheEntry); time.Since(entry.timestamp) < t.ttl {
            return entry.result, nil
        }
    }
    
    // Compute result
    result, err := t.compute(ctx, params)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    t.cache.Add(key, &cacheEntry{
        result:    result,
        timestamp: time.Now(),
    })
    
    return result, nil
}
```

### Composable Tools

Create tools that use other tools:

```go
func CreateResearchTool(searchTool, extractTool, summarizeTool domain.Tool) domain.Tool {
    return tools.NewToolBuilder("research", "Comprehensive research tool").
        WithFunction(func(ctx *ToolContext, params ResearchParams) (*ResearchResult, error) {
            // Use search tool
            searchResults, err := searchTool.Execute(ctx, SearchParams{
                Query: params.Topic,
            })
            if err != nil {
                return nil, fmt.Errorf("search failed: %w", err)
            }
            
            // Use extract tool on results
            extracted, err := extractTool.Execute(ctx, ExtractParams{
                Content: searchResults,
                Fields:  params.Fields,
            })
            if err != nil {
                return nil, fmt.Errorf("extraction failed: %w", err)
            }
            
            // Use summarize tool
            summary, err := summarizeTool.Execute(ctx, SummarizeParams{
                Content: extracted,
                Length:  params.SummaryLength,
            })
            if err != nil {
                return nil, fmt.Errorf("summarization failed: %w", err)
            }
            
            return &ResearchResult{
                Summary: summary,
                Sources: searchResults.Sources,
                Data:    extracted,
            }, nil
        }).
        Build()
}
```

## Migration from Legacy Tools

If migrating from older tool patterns:

### Old Pattern
```go
func WebFetch(url string) (string, error) {
    // Direct function
}
```

### New Pattern
```go
tool := tools.NewToolBuilder("web_fetch", "Fetches web content").
    WithFunction(func(ctx *ToolContext, params WebFetchParams) (*WebFetchResult, error) {
        // Implementation with context and structured params
    }).
    WithParameterSchema(schema).
    WithUsageInstructions("...").
    WithExamples(...).
    Build()
```

## Debugging Tools

### Enable Debug Logging

```go
func (t *MyTool) Execute(ctx *ToolContext, params interface{}) (interface{}, error) {
    if debug.Enabled() {
        debug.Log("Tool %s executing with params: %+v", t.Name(), params)
    }
    
    result, err := t.execute(ctx, params)
    
    if debug.Enabled() {
        if err != nil {
            debug.Log("Tool %s failed: %v", t.Name(), err)
        } else {
            debug.Log("Tool %s succeeded: %+v", t.Name(), result)
        }
    }
    
    return result, err
}
```

### Tool Metrics

```go
var (
    toolExecutions = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "tool_executions_total",
            Help: "Total number of tool executions",
        },
        []string{"tool", "status"},
    )
    
    toolDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "tool_execution_duration_seconds",
            Help: "Tool execution duration",
        },
        []string{"tool"},
    )
)

func instrumentedExecute(tool domain.Tool) domain.Tool {
    return &instrumentedTool{
        Tool: tool,
    }
}

func (t *instrumentedTool) Execute(ctx *ToolContext, params interface{}) (interface{}, error) {
    start := time.Now()
    
    result, err := t.Tool.Execute(ctx, params)
    
    duration := time.Since(start)
    toolDuration.WithLabelValues(t.Name()).Observe(duration.Seconds())
    
    if err != nil {
        toolExecutions.WithLabelValues(t.Name(), "error").Inc()
    } else {
        toolExecutions.WithLabelValues(t.Name(), "success").Inc()
    }
    
    return result, err
}
```

## Next Steps

- See [Tools API Reference](/docs/api/tools.md) for the complete API
- Check [Built-in Tools](/docs/api/builtins.md) for implementation examples
- Review [Agent Integration](/docs/technical/agents.md) for tool usage in agents
- Explore [Examples](/cmd/examples/) for real-world tool implementations