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
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
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
dataTools := tools.Tools.ListByCategory("data")
dateTimeTools := tools.Tools.ListByCategory("datetime")

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

### Data Tools

1. **json_process** - Process JSON data
   - Parse and validate JSON
   - Query with JSONPath expressions
   - Transform operations (flatten, prettify, minify, extract keys/values)

2. **csv_process** - Handle CSV data
   - Parse with configurable delimiters and headers
   - Filter with multiple operators
   - Transform operations (select columns, sort, statistics)
   - Convert to JSON

3. **xml_process** - Process XML data
   - Parse with attribute support
   - Query with simplified XPath
   - Convert to JSON with configurable attribute handling

4. **data_transform** - Common data transformations
   - Filter with complex conditions
   - Map operations (extract field, case conversion, type conversion)
   - Reduce operations (sum, count, min, max, average, concat)
   - Additional operations (sort, group_by, unique, reverse)

### Date Time Tools

1. **datetime_now** - Get current date/time
   - UTC and local timezone support
   - Custom timezone selection
   - Date components extraction
   - Week/quarter/year day calculations
   - Unix timestamps (all precision levels)
   - Custom format output

2. **datetime_info** - Get date information
   - Day/week/month/year information
   - Leap year detection
   - Days in month calculation
   - Period boundaries (start/end of week/month/quarter/year)
   - Configurable week start (Sunday/Monday)
   - Full timezone support

3. **datetime_calculate** - Date arithmetic
   - Add/subtract time units (days, hours, minutes, seconds)
   - Month/year arithmetic with proper handling
   - Duration calculations between dates
   - Business day calculations
   - Next/previous weekday finding
   - Age calculations
   - Timezone-aware operations

4. **datetime_parse** - Parse date/time strings
   - Auto-detect common formats
   - Custom format support
   - Relative date parsing ("yesterday", "next Monday", "in 3 days")
   - Unix timestamp parsing
   - Timezone parsing
   - Comprehensive validation

5. **datetime_format** - Format dates
   - Standard formats (RFC3339, ISO 8601, Kitchen, etc.)
   - Custom format strings
   - Localized formatting (Spanish, French, German, Italian, Portuguese, Russian)
   - Relative time ("2 hours ago", "in 3 days")
   - Multiple formats in single call
   - Timezone-aware formatting

6. **datetime_convert** - Convert dates/times
   - Timezone conversions with DST handling
   - List available timezones
   - Unix timestamp conversions (all precision levels)
   - DST information and detection
   - Timezone offset information

7. **datetime_compare** - Compare dates
   - Before/after/equal comparisons
   - Same period checks (day/week/month/year)
   - Range checks
   - Sort multiple dates
   - Find earliest/latest dates
   - Human-readable time differences

### Feed Tools

1. **feed_fetch** - Retrieve and parse feeds
   - Support for RSS 2.0, Atom 1.0, JSON Feed 1.1
   - Automatic format detection
   - HTTP handling with proper user agents
   - Conditional requests (If-Modified-Since, ETags)
   - Size and timeout limits
   - Unified output format for all feed types

2. **feed_discover** - Auto-discover feed URLs
   - Parse HTML for feed link tags
   - Check common feed URL patterns (/feed, /rss, etc.)
   - Validate discovered feeds with HEAD requests
   - Return feed type and metadata
   - Support for relative URL resolution

3. **feed_filter** - Filter feed items
   - Filter by date range (published/updated)
   - Keyword matching in title/content/description
   - Author filtering
   - Category/tag filtering
   - Match all/any logic for multiple criteria
   - Case-sensitive/insensitive options
   - Limit number of results

4. **feed_aggregate** - Combine multiple feeds
   - Merge feeds while detecting duplicates
   - Sort by date or title
   - Remove duplicates by URL or content hash
   - Optional metadata merging
   - Configurable item limits

5. **feed_convert** - Convert between formats
   - RSS 2.0 ↔ Atom 1.0 conversion
   - JSON Feed ↔ RSS/Atom conversion
   - Pretty-print option for human-readable output
   - Content inclusion control
   - Preserve maximum information during conversion

6. **feed_extract** - Extract data from feeds
   - Extract specific fields from items
   - Nested field extraction (author.name)
   - Flatten nested fields option
   - Get categories and tags
   - Extract media content (enclosures)
   - Feed-level metadata extraction
   - Custom field extraction with dot notation

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
    _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
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

## Using Feed Tools

Feed tools are designed to work together for comprehensive feed processing:

```go
import (
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed"
)

// Discover feeds from a website
discover := feed.FeedDiscover()
result, _ := discover.Execute(ctx, map[string]interface{}{
    "url": "https://example.com",
})

// Fetch and parse a feed
fetch := feed.FeedFetch()
result, _ := fetch.Execute(ctx, map[string]interface{}{
    "url": "https://example.com/rss",
    "max_items": 10,
})

// Filter items by criteria
filter := feed.FeedFilter()
result, _ := filter.Execute(ctx, map[string]interface{}{
    "feed": fetchResult.Feed,
    "keywords": []string{"technology", "AI"},
    "after_date": "2024-01-01",
})

// Convert to different format
convert := feed.FeedConvert()
result, _ := convert.Execute(ctx, map[string]interface{}{
    "feed": filterResult.Feed,
    "target_type": "json",
    "pretty": true,
})
```

### Common Feed Processing Workflows

1. **News Aggregation**: Fetch multiple feeds → Aggregate → Filter by date → Extract titles and links
2. **Content Monitoring**: Discover feeds → Filter by keywords → Convert to unified format
3. **Podcast Management**: Fetch podcast feeds → Extract enclosures → Filter by date
4. **Feed Migration**: Fetch old format → Convert to new format → Validate output

Stay tuned for updates as we expand the built-in component library!