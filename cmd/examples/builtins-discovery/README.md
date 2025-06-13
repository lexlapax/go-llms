# Built-in Tools Discovery Example

This example demonstrates the comprehensive built-in tools registry system, showing how to discover, search, and use all pre-built tools across all categories.

## Overview

The built-in tools system provides:
- Automatic tool registration via imports
- Discovery capabilities (list, search, filter by category)
- Standardized tool interfaces
- Enhanced features over custom implementations
- Complete coverage of web, file, system, data, datetime, and feed operations

## Running the Example

```bash
# Optional: Set API key for agent demonstration
export OPENAI_API_KEY="your-api-key"

# Run the example
go run main.go
```

## What This Example Shows

1. **Complete Tool Registry**
   - Lists all 36+ registered tools across 6 categories
   - Shows total tool count and category breakdown
   - Displays tool descriptions and metadata

2. **Tool Discovery Methods**
   - Search tools by keyword (e.g., "fetch")
   - Filter tools by category
   - Find tools by tags (e.g., "json")
   - View tool parameters and schemas

3. **Tool Usage Examples**
   - Direct tool execution (datetime_now, data_transform)
   - Adding multiple tools to agents
   - Combining tools for complex workflows

4. **Tool Statistics**
   - Total tools available
   - Tools per category breakdown
   - Tag-based tool discovery

## Key Concepts

### Auto-Registration
Tools are automatically registered when you import their packages:
```go
import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed"
```

### Tool Discovery
```go
// List all tools
allTools := tools.Tools.List()

// Search for tools
fetchTools := tools.Tools.Search("fetch")

// Filter by category
fileTools := tools.Tools.ListByCategory("file")

// Find by tags
for _, entry := range tools.Tools.List() {
    if contains(entry.Metadata.Tags, "json") {
        // Tool works with JSON
    }
}
```

### Enhanced Discovery (v0.3.4+)
The new discovery system allows metadata-first exploration without imports:
```go
// Create discovery instance (no imports needed!)
discovery := tools.NewDiscovery()

// List all available tools without loading them
availableTools := discovery.ListTools()
fmt.Printf("Found %d tools available\n", len(availableTools))

// Search tools by keyword
jsonTools := discovery.SearchTools("json")
fmt.Printf("Tools that work with JSON: %d\n", len(jsonTools))

// Get tool details without creating it
schema, _ := discovery.GetToolSchema("calculator")
examples, _ := discovery.GetToolExamples("calculator")

// Create tools only when needed
calculator, _ := discovery.CreateTool("calculator")

// Perfect for scripting engines and dynamic environments
```

### Using Tools
```go
// Get a specific tool
tool, ok := tools.GetTool("datetime_now")
if ok {
    result, err := tool.Execute(ctx, params)
}

// Add to agent
agent.AddTool(webFetch).
      AddTool(jsonProcess).
      AddTool(dateCalc)
```

## Available Built-in Tool Categories

### Web Tools (4 tools)
- **web_fetch**: Fetch content from URLs with timeout support
- **web_search**: Search the web using various engines
- **web_scrape**: Extract structured data from web pages
- **http_request**: Make HTTP requests with full control

### File Tools (6 tools)
- **read_file**: Read files with streaming support
- **write_file**: Write files with atomic operations
- **delete_file**: Delete files and directories
- **list_files**: List directory contents
- **move_file**: Move or rename files
- **search_files**: Search files by pattern

### System Tools (4 tools)
- **execute_command**: Execute system commands
- **get_environment_variable**: Read environment variables
- **get_system_info**: Get system information
- **process_list**: List running processes

### Data Tools (4 tools)
- **json_process**: Parse and query JSON data
- **csv_process**: Parse and transform CSV data
- **xml_process**: Parse and query XML data
- **data_transform**: Transform data with operations like filter, map, sort

### DateTime Tools (7 tools)
- **datetime_now**: Get current date/time
- **datetime_info**: Get detailed date/time information
- **datetime_calculate**: Perform date calculations
- **datetime_parse**: Parse date strings
- **datetime_format**: Format dates
- **datetime_convert**: Convert between timezones
- **datetime_compare**: Compare dates

### Feed Tools (6 tools)
- **feed_fetch**: Fetch RSS/Atom/JSON feeds
- **feed_discover**: Discover feeds from websites
- **feed_filter**: Filter feed items
- **feed_aggregate**: Combine multiple feeds
- **feed_convert**: Convert between feed formats
- **feed_extract**: Extract specific fields from feeds

## Benefits of Built-in Tools

- **Standardized Interfaces**: Consistent API across all tools
- **Enhanced Features**: Timeouts, streaming, retries built-in
- **Automatic Registration**: No manual setup required
- **Easy Discovery**: Search and filter capabilities
- **Comprehensive Documentation**: Each tool has examples
- **Regular Updates**: Maintained and improved over time

## Next Steps

- Explore category-specific examples:
  - [Web Tools Example](../builtins-web-tools/)
  - [File Tools Example](../builtins-file-tools/)
  - [System Tools Example](../builtins-system-tools/)
  - [Data Tools Example](../builtins-data-tools/)
  - [DateTime Tools Example](../builtins-datetime-tools/)
  - [Feed Tools Example](../builtins-feed-tools/)
- Check the [agent example](../agent/) to see tools in complex workflows
- Review the [built-in components guide](../../../docs/user-guide/built-in-components.md)