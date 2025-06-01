# Built-in Components Guide

Go-LLMs provides a comprehensive set of built-in tools, agents, and workflows that can be discovered and used directly in your applications. This guide covers how to use the built-in component system.

## Overview

The built-in components system provides:
- **Registry System**: Thread-safe component registry with search and discovery
- **Pre-built Tools**: Web, file, and system tools ready to use
- **Agent Templates**: Pre-configured agents for common tasks (coming soon)
- **Workflow Patterns**: Multi-agent coordination patterns (coming soon)

## Using Built-in Tools

### Tool Discovery

Built-in tools are automatically registered and can be discovered at runtime:

```go
import (
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
    // Import tool categories to register them
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
)

// List all available tools
allTools := tools.Tools.List()
for _, entry := range allTools {
    fmt.Printf("Tool: %s - %s\n", entry.Metadata.Name, entry.Metadata.Description)
}

// Find tools by category
webTools := tools.Tools.ListByCategory("web")
fileTools := tools.Tools.ListByCategory("file")
systemTools := tools.Tools.ListByCategory("system")

// Search for tools by name or description
searchResults := tools.Tools.Search("file")
```

### Using a Tool

Once you've discovered a tool, you can retrieve and use it:

```go
// Get a specific tool
tool, found := tools.GetTool("web_fetch")
if !found {
    log.Fatal("Tool not found")
}

// Or use MustGetTool (panics if not found)
tool := tools.MustGetTool("file_read")

// Execute the tool
ctx := context.Background()
result, err := tool.Execute(ctx, map[string]interface{}{
    "path": "/path/to/file.txt",
})
if err != nil {
    log.Fatalf("Tool execution failed: %v", err)
}

// Type assert to get specific result
readResult := result.(*file.ReadFileResult)
fmt.Printf("File content: %s\n", readResult.Content)
```

## Available Built-in Tools

### Web Tools

1. **web_fetch** - Fetches content from URLs
   - Supports custom timeouts and headers
   - Captures response metadata

2. **web_search** - Search the web using DuckDuckGo
   - Configurable result limits
   - Safe search filtering

3. **web_scrape** - Extract structured data from HTML
   - CSS-like selectors
   - Link extraction
   - Metadata parsing

4. **http_request** - Advanced HTTP operations
   - Full HTTP method support
   - Authentication (basic, bearer, API key)
   - Custom headers and body types

### File Tools

1. **file_read** - Enhanced file reading
   - Large file support with streaming
   - Binary detection
   - Line range reading
   - Metadata extraction

2. **file_write** - Enhanced file writing
   - Atomic operations
   - Append mode
   - Backup creation
   - Custom permissions

3. **file_list** - Directory listing
   - Pattern matching
   - Size/date filtering
   - Recursive traversal
   - Sorting options

4. **file_delete** - Safe file deletion
   - Safety checks for system directories
   - Confirmation requirements
   - Recursive directory deletion

5. **file_move** - Move/rename files
   - Atomic operations
   - Cross-device support
   - Directory moves

6. **file_search** - Search file contents
   - Regex support
   - Context lines
   - Binary file detection

### System Tools

1. **execute_command** - Run system commands
   - Environment variable support
   - Working directory control
   - Timeout management
   - Safety checks

2. **get_environment_variable** - Read environment variables
   - Pattern matching
   - Sensitive variable masking
   - Sorted output

3. **get_system_info** - System information
   - OS and architecture details
   - Memory statistics
   - Runtime information
   - Environment summary

4. **process_list** - List running processes
   - Cross-platform support
   - Filtering and sorting
   - Resource usage info

## Tool Metadata

Each tool provides rich metadata:

```go
entry := tools.Tools.Get("web_fetch")
meta := entry.Metadata

fmt.Printf("Name: %s\n", meta.Name)
fmt.Printf("Category: %s\n", meta.Category)
fmt.Printf("Description: %s\n", meta.Description)
fmt.Printf("Version: %s\n", meta.Version)
fmt.Printf("Author: %s\n", meta.Author)

// Resource usage hints
fmt.Printf("CPU Usage: %s\n", meta.ResourceUsage.CPUUsage)
fmt.Printf("Memory Usage: %s\n", meta.ResourceUsage.MemoryUsage)
fmt.Printf("Network Usage: %s\n", meta.ResourceUsage.NetworkUsage)

// Required permissions
for _, perm := range meta.Permissions {
    fmt.Printf("Permission: %s\n", perm)
}

// Examples
for _, example := range meta.Examples {
    fmt.Printf("Example: %s\n", example.Description)
    fmt.Printf("Input: %v\n", example.Input)
    fmt.Printf("Expected Output: %v\n", example.ExpectedOutput)
}
```

## Using Tools with Agents

Built-in tools integrate seamlessly with the agent system:

```go
import (
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
)

// Create an agent with built-in tools
agent := workflow.NewAgent(
    "research-agent",
    provider,
    workflow.WithTools(
        tools.MustGetTool("web_search"),
        tools.MustGetTool("web_fetch"),
        tools.MustGetTool("file_write"),
    ),
)

// The agent can now use these tools in its workflow
response, err := agent.Run(ctx, workflow.UserMessage(
    "Search for information about Go generics and save it to a file",
))
```

## Creating Custom Tools

You can create custom tools that integrate with the built-in registry:

```go
import (
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
    atools "github.com/lexlapax/go-llms/pkg/agent/tools"
)

func init() {
    // Create custom tool
    customTool := atools.NewTool(
        "my_custom_tool",
        "Does something custom",
        func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
            // Tool implementation
            return "result", nil
        },
    )
    
    // Register with metadata
    tools.Tools.Register("my_custom_tool", customTool, tools.Metadata{
        Name:        "my_custom_tool",
        Category:    "custom",
        Description: "A custom tool for specific tasks",
        Version:     "1.0.0",
        Author:      "Your Name",
    })
}
```

## Migration from common_tools.go

If you're migrating from the old common_tools.go approach:

```go
// Old approach
import "github.com/lexlapax/go-llms/pkg/agent/tools"
tool := tools.WebFetch()

// New approach
import (
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)
tool := tools.MustGetTool("web_fetch")
```

Benefits of the new approach:
- Dynamic discovery of available tools
- Rich metadata for each tool
- Version tracking
- Resource usage hints
- Permission declarations
- Better organization and modularity

## Best Practices

1. **Import Tool Categories**: Always import the tool categories you need to ensure registration
2. **Check Tool Availability**: Use the discovery features to verify tools are available
3. **Handle Errors**: Always check errors from tool execution
4. **Use Type Assertions**: Type assert results to access specific fields
5. **Respect Permissions**: Check tool permissions before use in restricted environments

## Future Components

The built-in component system will expand to include:

### Agent Templates (Coming Soon)
- WebResearcher - Web research with source tracking
- CodeReviewer - Code review and analysis
- DataAnalyst - Data analysis and insights
- And more...

### Workflow Patterns (Coming Soon)
- Pipeline - Sequential processing
- MapReduce - Parallel processing
- Consensus - Multi-agent agreement
- Research workflows
- Code review workflows
- Data processing pipelines

Stay tuned for updates as we expand the built-in component library!