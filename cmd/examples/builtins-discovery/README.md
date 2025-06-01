# Built-in Tools Discovery Example

This example demonstrates the built-in tools registry system, showing how to discover, search, and use pre-built tools.

## Overview

The built-in tools system provides:
- Automatic tool registration via imports
- Discovery capabilities (list, search, filter by category)
- Standardized tool interfaces
- Enhanced features over custom implementations

## Running the Example

```bash
# Optional: Set API key for agent demonstration
export OPENAI_API_KEY="your-api-key"

# Run the example
go run main.go
```

## What This Example Shows

1. **Tool Discovery**
   - List all registered tools
   - Search tools by keyword
   - Filter tools by category
   - View tool metadata (version, tags, examples)

2. **Using Built-in Tools**
   - Retrieve tools from the registry
   - Add tools to agents
   - Execute tools with proper parameters

3. **Migration Guidance**
   - Comparison with custom tool creation
   - Benefits of using built-in tools
   - How to transition existing code

## Key Concepts

### Auto-Registration
Tools are automatically registered when you import their packages:
```go
import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
```

### Tool Discovery
```go
// List all tools
allTools := tools.Tools.List()

// Search for tools
webTools := tools.Tools.Search("web")

// Filter by category
fileTools := tools.Tools.ListByCategory("file")
```

### Using Tools
```go
// Get a specific tool
tool, ok := tools.GetTool("web_fetch")
if ok {
    agent.AddTool(tool)
}
```

## Available Built-in Tools

- **web_fetch**: Fetch content from URLs with timeout support
- **web_search**: Search the web using various engines
- **file_read**: Read files with streaming and metadata
- **file_write**: Write files with atomic operations and backups

## Next Steps

- Explore the [file tools example](../builtins-file-tools/) for advanced file operations
- Check the [agent example](../agent/) to see tools in complex workflows
- Review tool metadata for resource usage and permissions