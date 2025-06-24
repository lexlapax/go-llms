# Tools API

Create and manage agent tools - ToolBuilder pattern and agent-tool conversion

## Package Information

- **Import Path**: `github.com/lexlapax/go-llms/pkg/agent/tools`
- **Category**: Agent Framework
- **Stability**: Stable (v0.3.x)

## Overview

The Tools package defines the interfaces and patterns for creating agent tools. It provides the ToolBuilder pattern for rich metadata, automatic documentation generation, and seamless integration with agents.

Key features:
- ToolBuilder for declarative tool creation
- Automatic schema generation
- Tool discovery and registration
- Metadata and documentation support
- Performance tracking
- MCP (Model Context Protocol) compatibility

## Core Types

### Tool Interface

All tools implement this interface:

```go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    GetInputSchema() *schema.Schema
    GetOutputSchema() *schema.Schema
}
```

### ToolBuilder Pattern

Create tools with rich metadata:

```go
tool := tools.NewToolBuilder("my-tool").
    WithDescription("Does something useful").
    WithInputSchema(inputSchema).
    WithOutputSchema(outputSchema).
    WithExecutor(func(ctx context.Context, input interface{}) (interface{}, error) {
        // Tool logic here
        return result, nil
    }).
    Build()
```

### Tool Registry

Manage and discover tools:

```go
registry := tools.GetGlobalRegistry()
registry.Register(tool)

// Discover tools by category
webTools := registry.GetByCategory("web")
```
## Examples

See the examples directory for usage examples.
## Best Practices

Follow Go best practices and the patterns shown in examples.
## Error Handling

Check error types and implement appropriate recovery strategies.