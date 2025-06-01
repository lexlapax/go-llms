# Migration Guide: Using Built-in Tools

This guide shows how to migrate from creating custom tools to using the built-in tools from the registry.

## Key Changes

### 1. Imports

**Before (Custom Tools):**
```go
import (
    "github.com/lexlapax/go-llms/pkg/agent/tools"
    // ... other imports
)
```

**After (Built-in Tools):**
```go
import (
    "github.com/lexlapax/go-llms/pkg/agent/tools"
    
    // Import built-in tools to trigger registration
    builtinTools "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)
```

### 2. Tool Creation

**Before (Custom Implementation):**
```go
// Custom web fetch tool
agent.AddTool(tools.NewTool(
    "web_fetch",
    "Fetch content from a URL",
    func(params struct {
        URL string `json:"url"`
    }) (map[string]interface{}, error) {
        // Custom implementation with http.Get
        resp, err := http.Get(params.URL)
        // ... handle response
    },
    &schemaDomain.Schema{
        Type: "object",
        Properties: map[string]schemaDomain.Property{
            "url": {
                Type:        "string",
                Description: "The URL to fetch",
            },
        },
        Required: []string{"url"},
    },
))
```

**After (Built-in Tool):**
```go
// Use built-in web fetch tool
if webFetch, ok := builtinTools.GetTool("web_fetch"); ok {
    agent.AddTool(webFetch)
}
```

### 3. Enhanced Capabilities

The built-in tools provide enhanced features:

#### Web Fetch Tool
- **Custom timeout support**: Configure request timeout
- **Header capture**: Access response headers
- **Better error handling**: Context-aware cancellation
- **Resource metadata**: Track resource usage

#### File Tools
- **Large file handling**: Streaming with 4KB buffer
- **Binary file detection**: Automatic encoding detection
- **Line range reading**: Read specific line ranges
- **Atomic writes**: Write to temp file then rename
- **Backup creation**: Optional backup with timestamps

#### Web Search Tool (New)
- **Multiple search engines**: DuckDuckGo support (more coming)
- **Result limits**: Configure max results
- **Safe search**: Enable/disable safe search
- **Structured results**: Title, URL, description, snippets

## Migration Steps

1. **Update imports** to include built-in tool packages
2. **Remove custom tool implementations** for standard operations
3. **Replace with registry lookups**:
   ```go
   // Get tool from registry
   tool, ok := builtinTools.GetTool("tool_name")
   if ok {
       agent.AddTool(tool)
   }
   ```
4. **Keep custom tools** for domain-specific operations not in built-ins

## Benefits of Using Built-in Tools

1. **Consistency**: Standardized tool interfaces across projects
2. **Performance**: Optimized implementations with pooling
3. **Maintenance**: Updates and bug fixes handled centrally
4. **Discovery**: Easy to find available tools via registry
5. **Documentation**: Built-in examples and metadata

## Example: Complete Migration

See `main_builtin.go` for a complete example using built-in tools alongside custom tools for functionality not yet available in the registry.

## Tools Still Requiring Custom Implementation

These tools aren't yet available as built-ins and still need custom implementation:
- `get_current_date`: Date/time utilities
- `calculator`: Mathematical expressions
- `execute_command`: System command execution (coming soon)

## Next Steps

As more tools are added to the built-in registry, you can progressively replace custom implementations with registry lookups, reducing code maintenance burden while gaining enhanced capabilities.