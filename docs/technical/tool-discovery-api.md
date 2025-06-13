# Tool Discovery API

The Tool Discovery API provides metadata-first access to tools without requiring imports, making it perfect for scripting engines and dynamic environments.

## Overview

The discovery system separates tool metadata from implementation, enabling:
- **Zero-import exploration** - Browse tools without loading packages
- **Lazy instantiation** - Create tools only when needed  
- **Rich metadata access** - Schemas, examples, help text
- **Build tag isolation** - Avoid import cycles
- **Bridge integration** - Designed for go-llmspell

## Basic Usage

```go
import discoveryTools "github.com/lexlapax/go-llms/pkg/agent/tools"

// Create discovery instance
discovery := discoveryTools.NewDiscovery()

// List all available tools
tools := discovery.ListTools()

// Search by keyword
jsonTools := discovery.SearchTools("json")

// Filter by category
mathTools := discovery.ListByCategory("math")

// Get tool schema
schema, _ := discovery.GetToolSchema("calculator")

// Get examples
examples, _ := discovery.GetToolExamples("calculator")

// Create tool when needed
tool, _ := discovery.CreateTool("calculator")
```

## Discovery Interface

```go
type ToolDiscovery interface {
    // ListTools returns all available tools without loading them
    ListTools() []ToolInfo
    
    // SearchTools searches tools by keyword in name, description, or tags
    SearchTools(query string) []ToolInfo
    
    // ListByCategory returns tools in a specific category
    ListByCategory(category string) []ToolInfo
    
    // GetToolSchema returns detailed schema for a specific tool
    GetToolSchema(name string) (*ToolSchema, error)
    
    // GetToolExamples returns examples for a specific tool
    GetToolExamples(name string) ([]domain.ToolExample, error)
    
    // CreateTool instantiates a tool by name
    CreateTool(name string) (domain.Tool, error)
    
    // CreateTools instantiates multiple tools
    CreateTools(names ...string) (map[string]domain.Tool, error)
    
    // GetToolHelp generates help text for a tool
    GetToolHelp(name string) (string, error)
}
```

## Metadata Schema

### ToolInfo Structure

The `ToolInfo` struct provides lightweight metadata for tool discovery:

```go
type ToolInfo struct {
    Name            string          `json:"name"`
    Description     string          `json:"description"`
    Category        string          `json:"category"`
    Tags            []string        `json:"tags"`
    Version         string          `json:"version"`
    ParameterSchema json.RawMessage `json:"parameter_schema,omitempty"`
    OutputSchema    json.RawMessage `json:"output_schema,omitempty"`
    Examples        []Example       `json:"examples,omitempty"`
    UsageHint       string          `json:"usage_hint,omitempty"`
    Package         string          `json:"package,omitempty"`
}
```

### Example Structure

Tool examples include both input and expected output:

```go
type Example struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Input       json.RawMessage `json:"input"`
    Output      json.RawMessage `json:"output,omitempty"`
}
```

### ToolSchema Structure

Detailed schema information for tool usage:

```go
type ToolSchema struct {
    Name          string               `json:"name"`
    Description   string               `json:"description"`
    Parameters    interface{}          `json:"parameters,omitempty"`
    Output        interface{}          `json:"output,omitempty"`
    Examples      []domain.ToolExample `json:"examples,omitempty"`
    Constraints   []string             `json:"constraints,omitempty"`
    ErrorGuidance map[string]string    `json:"error_guidance,omitempty"`
}
```

## Bridge Integration

### For go-llmspell

The discovery API is designed for scripting bridge integration:

```go
// Get all tool metadata for bridge exposure
metadata := discoveryTools.GetToolMetadata()

// Convert to bridge-friendly format
for name, info := range metadata {
    toolData := map[string]interface{}{
        "name":        name,
        "description": info.Description,
        "category":    info.Category,
        "tags":        info.Tags,
    }
    
    // Parse schemas for script access
    if len(info.ParameterSchema) > 0 {
        var params interface{}
        json.Unmarshal(info.ParameterSchema, &params)
        toolData["parameters"] = params
    }
    
    // Expose to Lua/JavaScript
    bridge.ExposeToolMetadata(name, toolData)
}
```

### Lua Integration Example

```lua
-- List available tools (exposed via bridge)
local tools = llms.list_tools()
print("Found " .. #tools .. " tools")

-- Search for specific tools
local jsonTools = llms.search_tools("json")
for _, tool in ipairs(jsonTools) do
    print("JSON tool: " .. tool.name .. " - " .. tool.description)
end

-- Get tool schema before use
local calcSchema = llms.get_tool_schema("calculator")
print("Calculator parameters: " .. json.encode(calcSchema.parameters))

-- Use tool dynamically
local result = llms.use_tool("calculator", {
    operation = "add",
    operand1 = 10,
    operand2 = 5
})
print("Result: " .. result.result)
```

## Build Tags

The discovery system works with different build configurations:

```bash
# Metadata-only (recommended for exploration)
go run main.go
# Tools discoverable but not loadable

# Full tool loading  
go run -tags=tools main.go
# Tools both discoverable and loadable
```

### Build Tag Implementation

Tools are conditionally imported using build tags:

```go
//go:build tools
// +build tools

package tools

func init() {
    toolFactories["calculator"] = func() (domain.Tool, error) {
        return math.Calculator(), nil
    }
}
```

## Error Handling

### Tool Not Loaded

When tools aren't available (no build tags):

```go
tool, err := discovery.CreateTool("calculator")
if err != nil {
    // Expected: "tool calculator not yet loaded - import the tool package to use it"
    log.Printf("Tool not available: %v", err)
}
```

### Tool Not Found

When tool doesn't exist:

```go
schema, err := discovery.GetToolSchema("nonexistent")
if err != nil {
    // "tool nonexistent not found"
    log.Printf("Tool not found: %v", err)
}
```

## Performance Considerations

- **Metadata is cached** - No repeated parsing overhead
- **Lazy loading** - Tools created only when needed
- **Singleton pattern** - Single discovery instance
- **No import cycles** - Build tags prevent unwanted dependencies

## Migration Guide

### From Legacy Registry

**Before:**
```go
import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math"

tool, found := tools.GetTool("calculator")
```

**After:**
```go
import discoveryTools "github.com/lexlapax/go-llms/pkg/agent/tools"

discovery := discoveryTools.NewDiscovery()
tool, err := discovery.CreateTool("calculator")
```

### Benefits of Migration

- **Dynamic discovery** - Find tools at runtime
- **Reduced imports** - Only load what you need
- **Rich metadata** - Access schemas and examples
- **Bridge compatibility** - Works with scripting engines

## Examples

- **Basic Discovery**: [builtins-discovery example](../../cmd/examples/builtins-discovery/)
- **Scripting Integration**: See go-llmspell bridge implementation
- **Dynamic Loading**: Runtime tool selection patterns

## Related Documentation

- [Tool Development Guide](tool-development.md)
- [Built-in Tools Overview](../user-guide/tools.md)
- [Agent Integration](../user-guide/agents.md)