# Working with Tools

This guide covers everything you need to know about tools in go-llms - from using built-in tools to creating your own.

## Overview

Tools extend what agents can do by providing functions they can call. Go-llms includes 32 built-in tools across 7 categories, plus the ability to create custom tools.

## Quick Start

### Using Built-in Tools

```go
import (
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
    // Import categories you need
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
)

// Get a tool
searchTool, _ := tools.GetTool("web_search")

// Use it directly
result, err := searchTool.Execute(ctx, map[string]interface{}{
    "query": "latest AI news",
})

// Or add to an agent
agent.AddTool(searchTool)
```

### Tool Discovery

```go
// List all available tools
allTools := tools.Tools.List()
for _, entry := range allTools {
    fmt.Printf("%s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
}

// Find tools by category
webTools := tools.Tools.ListByCategory("web")
fileTools := tools.Tools.ListByCategory("file")

// Search for tools
results := tools.Tools.Search("json")
```

### Enhanced Discovery (v0.3.4+)

The new discovery system allows exploring tools without importing them:

```go
// Create discovery instance
discovery := tools.NewDiscovery()

// List all tools without imports
availableTools := discovery.ListTools()
for _, info := range availableTools {
    fmt.Printf("%s (%s): %s\n", info.Name, info.Category, info.Description)
    fmt.Printf("  Tags: %v\n", info.Tags)
}

// Search tools by keyword
jsonTools := discovery.SearchTools("json")
webTools := discovery.SearchTools("web")

// Get tool help without loading it
help, _ := discovery.GetToolHelp("calculator")
fmt.Println(help)

// Load tool only when needed
calculator, _ := discovery.CreateTool("calculator")
result, _ := calculator.Execute(ctx, params)
```

This is especially useful for scripting environments where you want to discover tools dynamically.

## Built-in Tool Categories

### Web Tools (5 tools)

#### API Client (`api_client`) - v3.0.0
Advanced REST and GraphQL API client with authentication support.

**Key Features:**
- REST API support with all HTTP methods
- GraphQL queries and mutations
- OpenAPI/Swagger discovery
- Multiple authentication methods
- Automatic auth detection from agent state

**Example:**
```go
// REST API call
result, _ := apiClient.Execute(ctx, map[string]interface{}{
    "base_url": "https://api.github.com",
    "endpoint": "/repos/lexlapax/go-llms",
    "method": "GET",
})

// GraphQL query
result, _ := apiClient.Execute(ctx, map[string]interface{}{
    "base_url": "https://api.github.com",
    "endpoint": "/graphql",
    "graphql_query": "query { viewer { login name } }",
})
```

See [Built-in Tools Guide](builtin-tools.md#api-client-tool) for comprehensive documentation.

#### Web Search (`web_search`)
Search multiple engines with automatic fallback.

**Supported Engines:**
- DuckDuckGo (free, no API key)
- Brave Search (comprehensive, requires BRAVE_API_KEY)
- Tavily (LLM-optimized, requires TAVILY_API_KEY)
- Serper.dev (fast Google results, requires SERPERDEV_API_KEY)
- Serpapi (Google results, requires SERPAPI_API_KEY)

**Example:**
```go
result, _ := webSearch.Execute(ctx, map[string]interface{}{
    "query": "machine learning trends 2024",
    "max_results": 10,
    "engine": "tavily", // Optional, auto-selects by default
})
```

#### Web Fetch (`web_fetch`)
Fetch and extract content from web pages.

```go
result, _ := webFetch.Execute(ctx, map[string]interface{}{
    "url": "https://example.com/article",
    "timeout": 30,
    "extract_metadata": true,
})
```

#### Web Scrape (`web_scrape`)
Extract structured data using CSS selectors.

```go
result, _ := webScrape.Execute(ctx, map[string]interface{}{
    "url": "https://example.com",
    "selectors": map[string]string{
        "title": "h1",
        "price": ".price",
        "description": ".product-desc",
    },
})
```

#### HTTP Request (`http_request`)
Low-level HTTP operations with full control.

```go
result, _ := httpRequest.Execute(ctx, map[string]interface{}{
    "url": "https://api.example.com/data",
    "method": "POST",
    "headers": map[string]string{
        "Authorization": "Bearer token",
    },
    "body": `{"key": "value"}`,
})
```

### File Tools (6 tools)

#### File Read (`file_read`)
Read files with advanced options.

```go
result, _ := fileRead.Execute(ctx, map[string]interface{}{
    "path": "/path/to/file.txt",
    "start_line": 100,
    "end_line": 200,
})
```

#### File Write (`file_write`)
Write files safely with atomic operations.

```go
result, _ := fileWrite.Execute(ctx, map[string]interface{}{
    "path": "/path/to/output.txt",
    "content": "Hello, World!",
    "mode": "overwrite", // or "append"
    "create_dirs": true,
})
```

#### File List (`file_list`)
List directory contents with filtering.

```go
result, _ := fileList.Execute(ctx, map[string]interface{}{
    "path": "/home/user/documents",
    "pattern": "*.pdf",
    "recursive": true,
    "include_hidden": false,
})
```

#### File Search (`file_search`)
Search file contents with regex.

```go
result, _ := fileSearch.Execute(ctx, map[string]interface{}{
    "path": "/project",
    "pattern": "TODO|FIXME",
    "file_pattern": "*.go",
    "recursive": true,
})
```

### System Tools (4 tools)

#### Execute Command (`execute_command`)
Run system commands safely.

```go
result, _ := executeCommand.Execute(ctx, map[string]interface{}{
    "command": "git",
    "args": []string{"status", "--short"},
    "working_dir": "/project",
    "timeout": 10,
})
```

#### System Info (`get_system_info`)
Get comprehensive system information.

```go
result, _ := systemInfo.Execute(ctx, map[string]interface{}{
    "include_env": false,
    "include_network": true,
    "include_disk": true,
})
```

### Data Tools (4 tools)

#### JSON Process (`json_process`)
Process JSON with JSONPath queries.

```go
result, _ := jsonProcess.Execute(ctx, map[string]interface{}{
    "data": jsonString,
    "operation": "query",
    "query": "$.users[?(@.age > 18)].name",
})
```

#### CSV Process (`csv_process`)
Handle CSV data with transformations.

```go
result, _ := csvProcess.Execute(ctx, map[string]interface{}{
    "data": csvString,
    "operation": "filter",
    "filter": map[string]interface{}{
        "column": "age",
        "operator": ">",
        "value": 18,
    },
})
```

### DateTime Tools (7 tools)

#### DateTime Now (`datetime_now`)
Get current date/time in any timezone.

```go
result, _ := datetimeNow.Execute(ctx, map[string]interface{}{
    "timezone": "America/New_York",
    "format": "RFC3339",
})
```

#### DateTime Calculate (`datetime_calculate`)
Perform date arithmetic.

```go
result, _ := datetimeCalc.Execute(ctx, map[string]interface{}{
    "date": "2024-01-15",
    "operation": "add",
    "value": 30,
    "unit": "days",
})
```

### Feed Tools (6 tools)

#### Feed Fetch (`feed_fetch`)
Fetch and parse RSS/Atom/JSON feeds.

```go
result, _ := feedFetch.Execute(ctx, map[string]interface{}{
    "url": "https://example.com/rss",
    "max_items": 20,
})
```

#### Feed Filter (`feed_filter`)
Filter feed items by criteria.

```go
result, _ := feedFilter.Execute(ctx, map[string]interface{}{
    "feed": feedData,
    "keywords": []string{"technology", "AI"},
    "after": "2024-01-01",
})
```

### Math Tools (1 tool)

#### Calculator (`calculator`)
Perform mathematical calculations.

```go
result, _ := calculator.Execute(ctx, map[string]interface{}{
    "expression": "sqrt(16) + 3 * 4",
})
```

## Creating Custom Tools

### Basic Tool

```go
import "github.com/lexlapax/go-llms/pkg/agent/tools"

// Create a simple tool
myTool := tools.NewToolBuilder("my_tool", "Does something useful").
    WithFunction(func(ctx *domain.ToolContext, params map[string]interface{}) (interface{}, error) {
        // Your tool logic here
        name := params["name"].(string)
        return fmt.Sprintf("Hello, %s!", name), nil
    }).
    WithParameterSchema(&domain.Schema{
        Type: "object",
        Properties: map[string]*domain.Schema{
            "name": {
                Type:        "string",
                Description: "Name to greet",
            },
        },
        Required: []string{"name"},
    }).
    Build()
```

### Advanced Tool with Context

```go
// Tool that uses agent context
contextTool := tools.NewToolBuilder("context_aware", "Uses agent state").
    WithFunction(func(ctx *domain.ToolContext, params map[string]interface{}) (interface{}, error) {
        // Access agent state
        if ctx.State != nil {
            userId, _ := ctx.State.Get("user_id")
            // Use userId in your logic
        }
        
        // Emit progress events
        if ctx.Events != nil {
            ctx.Events.EmitProgress(50, 100, "Processing...")
        }
        
        return "result", nil
    }).
    WithCategory("custom").
    WithUsageInstructions("Use when you need user context").
    Build()
```

### Tool with Validation

```go
// Tool with comprehensive validation
validatedTool := tools.NewToolBuilder("validated", "Validates input").
    WithFunction(func(ctx *domain.ToolContext, params map[string]interface{}) (interface{}, error) {
        age := int(params["age"].(float64))
        email := params["email"].(string)
        
        // Process validated input
        return map[string]interface{}{
            "valid": true,
            "age": age,
            "email": email,
        }, nil
    }).
    WithParameterSchema(&domain.Schema{
        Type: "object",
        Properties: map[string]*domain.Schema{
            "age": {
                Type:        "integer",
                Minimum:     float64Ptr(0),
                Maximum:     float64Ptr(150),
                Description: "Person's age",
            },
            "email": {
                Type:        "string",
                Format:      "email",
                Description: "Email address",
            },
        },
        Required: []string{"age", "email"},
    }).
    Build()
```

## Agent-Tool Interoperability

### Converting Agents to Tools

Any agent can be used as a tool:

```go
// Convert agent to tool
myAgent := createCustomAgent()
agentTool := tools.NewAgentTool(myAgent)

// Use in another agent
coordinator.AddTool(agentTool)
```

### Converting Tools to Agents

Tools can become agents:

```go
// Convert tool to agent
webFetchTool := tools.MustGetTool("web_fetch")
fetchAgent := tools.NewToolAgent(webFetchTool)

// Use in workflows
workflow.AddAgent(fetchAgent)
```

### Tool Chains

Create composite tools:

```go
// Chain multiple operations
pipelineTool := tools.CreateToolChain(
    tools.MustGetTool("web_fetch"),
    tools.MustGetTool("json_process"),
    tools.MustGetTool("file_write"),
)

// Single tool that fetches, processes, and saves
result, _ := pipelineTool.Execute(ctx, params)
```

## Tool Patterns

### Tool with Retry Logic

```go
retryTool := tools.NewToolBuilder("retry_example", "Tool with retries").
    WithFunction(func(ctx *domain.ToolContext, params map[string]interface{}) (interface{}, error) {
        maxRetries := 3
        var lastErr error
        
        for i := 0; i < maxRetries; i++ {
            result, err := performOperation(params)
            if err == nil {
                return result, nil
            }
            lastErr = err
            time.Sleep(time.Second * time.Duration(i+1))
        }
        
        return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
    }).
    Build()
```

### Tool with Progress Reporting

```go
progressTool := tools.NewToolBuilder("progress_example", "Reports progress").
    WithFunction(func(ctx *domain.ToolContext, params map[string]interface{}) (interface{}, error) {
        items := params["items"].([]interface{})
        results := make([]interface{}, 0)
        
        for i, item := range items {
            // Report progress
            if ctx.Events != nil {
                progress := (i + 1) * 100 / len(items)
                ctx.Events.EmitProgress(progress, 100, fmt.Sprintf("Processing item %d", i+1))
            }
            
            // Process item
            result := processItem(item)
            results = append(results, result)
        }
        
        return results, nil
    }).
    Build()
```

### Tool with Caching

```go
var cache = make(map[string]interface{})
var cacheMu sync.RWMutex

cacheTool := tools.NewToolBuilder("cached_example", "Tool with caching").
    WithFunction(func(ctx *domain.ToolContext, params map[string]interface{}) (interface{}, error) {
        key := fmt.Sprintf("%v", params["key"])
        
        // Check cache
        cacheMu.RLock()
        if cached, ok := cache[key]; ok {
            cacheMu.RUnlock()
            return cached, nil
        }
        cacheMu.RUnlock()
        
        // Compute result
        result := expensiveOperation(params)
        
        // Store in cache
        cacheMu.Lock()
        cache[key] = result
        cacheMu.Unlock()
        
        return result, nil
    }).
    Build()
```

## Best Practices

### 1. Tool Design
- **Single Responsibility**: Each tool should do one thing well
- **Clear Parameters**: Use descriptive parameter names and schemas
- **Error Handling**: Return meaningful error messages
- **Documentation**: Provide examples and usage instructions

### 2. Performance
- **Timeouts**: Respect context timeouts
- **Resource Usage**: Declare resource requirements in metadata
- **Caching**: Cache expensive operations when appropriate
- **Concurrency**: Make tools thread-safe

### 3. Integration
- **State Access**: Use ToolContext to access agent state
- **Event Emission**: Report progress for long operations
- **Validation**: Validate parameters using schemas
- **Type Safety**: Handle type conversions carefully

### 4. Security
- **Input Validation**: Never trust user input
- **Path Traversal**: Validate file paths
- **Command Injection**: Sanitize command arguments
- **API Keys**: Use agent state for credentials

## Tool Metadata

Each tool provides rich metadata:

```go
meta := tool.GetMetadata()
fmt.Printf("Name: %s\n", meta.Name)
fmt.Printf("Category: %s\n", meta.Category)
fmt.Printf("Version: %s\n", meta.Version)
fmt.Printf("CPU Usage: %s\n", meta.ResourceUsage.CPUUsage)

// Check examples
for _, example := range meta.Examples {
    fmt.Printf("Example: %s\n", example.Description)
    fmt.Printf("Input: %v\n", example.Input)
}
```

## Registering Custom Tools

Add your tools to the registry:

```go
import "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"

func init() {
    tools.Tools.Register("my_tool", myTool, tools.Metadata{
        Name:        "my_tool",
        Category:    "custom",
        Description: "My custom tool",
        Version:     "1.0.0",
        Author:      "Your Name",
        Examples: []tools.Example{
            {
                Description: "Basic usage",
                Input: map[string]interface{}{
                    "param": "value",
                },
                ExpectedOutput: "result",
            },
        },
    })
}
```

## Tool Testing

Test your tools thoroughly:

```go
func TestMyTool(t *testing.T) {
    tool := createMyTool()
    
    // Test normal operation
    result, err := tool.Execute(ctx, map[string]interface{}{
        "input": "test",
    })
    assert.NoError(t, err)
    assert.Equal(t, "expected", result)
    
    // Test error cases
    _, err = tool.Execute(ctx, map[string]interface{}{
        "invalid": "param",
    })
    assert.Error(t, err)
    
    // Test with context
    ctx := &domain.ToolContext{
        Context: context.Background(),
        State:   domain.NewStateReader(state),
    }
    result, err = tool.Execute(ctx, params)
}
```

## Next Steps

Now that you understand tools:
- Explore [Workflows](workflows.md) to orchestrate multiple agents and tools
- See [Examples Gallery](examples-gallery.md) for real-world usage
- Check [API Reference](../api/tools.md) for detailed documentation

Ready to extend your agents with powerful tools? Let's build! 🛠️