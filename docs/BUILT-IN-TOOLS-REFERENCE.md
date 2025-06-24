# Go-LLMs Built-in Tools Reference

This document provides a comprehensive reference for all built-in tools available in the Go-LLMs library, including their purpose, parameters, usage examples, and integration patterns.

## Table of Contents

1. [Overview](#overview)
2. [Tool Interface](#tool-interface)
3. [Tool Discovery and Registration](#tool-discovery-and-registration)
4. [Tool Categories](#tool-categories)
   - [File System Tools](#file-system-tools)
   - [Web Tools](#web-tools)
   - [System Tools](#system-tools)
   - [Math Tools](#math-tools)
   - [Date/Time Tools](#datetime-tools)
   - [Data Processing Tools](#data-processing-tools)
   - [Feed Processing Tools](#feed-processing-tools)
5. [Usage Examples](#usage-examples)
6. [Tool Integration Patterns](#tool-integration-patterns)

## Overview

Go-LLMs provides a comprehensive set of built-in tools that enable LLM agents to interact with various systems and perform complex operations. All tools follow a consistent interface and provide:

- **Structured Parameters**: JSON Schema-based parameter validation
- **Type Safety**: Strongly typed inputs and outputs
- **Error Handling**: Comprehensive error guidance for LLMs
- **Event Support**: Progress tracking and custom events
- **State Integration**: Context-aware execution with state management
- **MCP Compatibility**: Export to Model Context Protocol format

## Tool Interface

All tools implement the `domain.Tool` interface:

```go
type Tool interface {
    // Core functionality
    Name() string
    Description() string
    Execute(ctx *ToolContext, params interface{}) (interface{}, error)
    
    // Schema definitions
    ParameterSchema() *domain.Schema
    OutputSchema() *domain.Schema
    
    // LLM guidance
    UsageInstructions() string
    Examples() []ToolExample
    Constraints() []string
    ErrorGuidance() map[string]string
    
    // Metadata
    Category() string
    Tags() []string
    Version() string
    
    // Behavioral hints
    IsDeterministic() bool
    IsDestructive() bool
    RequiresConfirmation() bool
    EstimatedLatency() string
    
    // MCP compatibility
    ToMCPDefinition() MCPToolDefinition
}
```

## Tool Discovery and Registration

### Registry System

Tools are registered in a global registry with metadata:

```go
// Access the global tool registry
registry := tools.Tools

// Register a new tool
tools.MustRegisterTool("tool_name", tool, tools.ToolMetadata{
    Metadata: builtins.Metadata{
        Name:        "tool_name",
        Category:    "category",
        Tags:        []string{"tag1", "tag2"},
        Description: "Tool description",
        Version:     "1.0.0",
    },
    RequiredPermissions: []string{"permission:action"},
    ResourceUsage: tools.ResourceInfo{
        Memory:      "low", // low, medium, high
        Network:     false,
        FileSystem:  true,
        Concurrency: true,
    },
})
```

### Discovery Methods

```go
// List all tools
allTools := registry.List()

// Get tool by name
tool, found := registry.Get("tool_name")

// Filter by category
categoryTools := registry.ListByCategory("file")

// Filter by permission
permTools := registry.ListByPermission("file:read")

// Filter by resource usage
lowMemTools := registry.ListByResourceUsage(tools.ResourceCriteria{
    MaxMemory: "low",
})
```

## Tool Categories

### File System Tools

#### file_read
Reads file contents with support for large files, line ranges, and metadata.

**Parameters:**
- `path` (string, required): File path to read
- `max_size` (int64): Maximum bytes to read (default: 10MB)
- `line_start` (int): Start reading from this line (1-based)
- `line_end` (int): Stop reading at this line (inclusive)
- `include_meta` (bool): Include file metadata

**Example:**
```go
result, err := tool.Execute(ctx, ReadFileParams{
    Path: "/path/to/file.txt",
    LineStart: 100,
    LineEnd: 200,
    IncludeMeta: true,
})
```

#### file_write
Writes content to files with backup options and atomic operations.

**Parameters:**
- `path` (string, required): File path to write
- `content` (string, required): Content to write
- `mode` (string): Write mode - "overwrite", "append", "create"
- `create_dirs` (bool): Create parent directories if needed
- `backup` (bool): Create backup before overwriting

#### file_list
Lists directory contents with filtering and metadata.

**Parameters:**
- `path` (string, required): Directory path
- `recursive` (bool): Include subdirectories
- `include_hidden` (bool): Include hidden files
- `pattern` (string): Glob pattern for filtering
- `sort_by` (string): Sort by "name", "size", "time"

#### file_search
Searches for files by name or content.

**Parameters:**
- `path` (string, required): Start directory
- `pattern` (string): Name pattern (glob)
- `content` (string): Content to search for
- `recursive` (bool): Search subdirectories
- `max_results` (int): Maximum results to return

#### file_move
Moves or renames files safely.

**Parameters:**
- `source` (string, required): Source file path
- `destination` (string, required): Destination path
- `overwrite` (bool): Overwrite if destination exists
- `create_dirs` (bool): Create parent directories

#### file_delete
Deletes files with optional confirmation.

**Parameters:**
- `path` (string, required): File to delete
- `recursive` (bool): Delete directories recursively
- `confirm` (bool): Require confirmation

### Web Tools

#### http_request
Makes HTTP requests with full control over headers, body, and authentication.

**Parameters:**
- `url` (string, required): Request URL
- `method` (string): HTTP method (GET, POST, PUT, DELETE, etc.)
- `headers` (map[string]string): HTTP headers
- `body` (string): Request body
- `body_type` (string): Body type (json, form, text, xml)
- `auth_type` (string): Authentication type (basic, bearer, api_key)
- `timeout` (int): Timeout in seconds

**Example:**
```go
result, err := tool.Execute(ctx, HTTPRequestParams{
    URL: "https://api.example.com/data",
    Method: "POST",
    Headers: map[string]string{
        "Content-Type": "application/json",
    },
    Body: `{"key": "value"}`,
    AuthType: "bearer",
    AuthToken: "your-token",
})
```

#### web_fetch
Fetches and extracts content from web pages.

**Parameters:**
- `url` (string, required): Page URL
- `selector` (string): CSS selector for extraction
- `extract` (string): What to extract (text, html, links, images)
- `timeout` (int): Timeout in seconds
- `user_agent` (string): Custom user agent

#### web_scrape
Extracts structured data from HTML pages.

**Parameters:**
- `url` (string, required): Page URL
- `rules` (map[string]string): Extraction rules (selector -> field mapping)
- `pagination` (object): Pagination configuration
- `wait_for` (string): CSS selector to wait for
- `javascript` (bool): Enable JavaScript rendering

#### web_search
Searches the web using search engines.

**Parameters:**
- `query` (string, required): Search query
- `engine` (string): Search engine (google, bing, duckduckgo)
- `results` (int): Number of results
- `region` (string): Region code
- `safe_search` (bool): Enable safe search

#### api_client
Advanced API client with OpenAPI/GraphQL discovery and caching.

**Parameters:**
- `url` (string, required): API endpoint
- `operation` (string): Operation ID or GraphQL query
- `variables` (map): Query/path parameters or GraphQL variables
- `discovery` (bool): Enable API discovery
- `cache` (bool): Enable response caching

### System Tools

#### get_system_info
Retrieves comprehensive system information.

**Parameters:**
- `include_environment` (bool): Include environment summary
- `include_memory` (bool): Include memory statistics
- `include_runtime` (bool): Include Go runtime info

**Output includes:**
- OS details (name, platform, version)
- Architecture and CPU count
- Memory usage statistics
- Runtime information
- Environment summary

#### execute_command
Executes system commands safely.

**Parameters:**
- `command` (string, required): Command to execute
- `args` ([]string): Command arguments
- `working_dir` (string): Working directory
- `env` (map[string]string): Environment variables
- `timeout` (int): Timeout in seconds
- `capture_output` (bool): Capture stdout/stderr

#### get_environment_variable
Gets environment variable values.

**Parameters:**
- `name` (string, required): Variable name
- `default` (string): Default value if not found
- `expand` (bool): Expand embedded variables

#### process_list
Lists running processes with filtering.

**Parameters:**
- `filter` (string): Process name filter
- `include_children` (bool): Include child processes
- `sort_by` (string): Sort by cpu, memory, pid
- `limit` (int): Maximum processes to return

### Math Tools

#### calculator
Performs mathematical calculations including arithmetic, trigonometry, and logarithms.

**Parameters:**
- `operation` (string, required): Mathematical operation
- `operand1` (float64, required): First operand
- `operand2` (float64): Second operand (for binary operations)

**Supported operations:**
- Basic: add, subtract, multiply, divide, mod, power
- Trigonometry: sin, cos, tan, asin, acos, atan
- Logarithms: log, log10, exp
- Other: sqrt, abs, ceil, floor, round
- Constants: pi, e, phi

**Example:**
```go
result, err := tool.Execute(ctx, CalculatorParams{
    Operation: "sin",
    Operand1: math.Pi/2,
})
```

### Date/Time Tools

#### datetime_now
Provides current date/time in various formats.

**Parameters:**
- `timezone` (string): Timezone (e.g., "America/New_York")
- `include_components` (bool): Include date/time components
- `include_week_info` (bool): Include week information
- `include_timestamps` (bool): Include unix timestamps
- `format` (string): Custom format string

#### datetime_parse
Parses date/time strings with automatic format detection.

**Parameters:**
- `input` (string, required): Date/time string to parse
- `format` (string): Expected format (optional)
- `timezone` (string): Timezone for parsing
- `strict` (bool): Strict parsing mode

#### datetime_format
Formats date/time values in various formats.

**Parameters:**
- `timestamp` (int64, required): Unix timestamp
- `format` (string, required): Output format
- `timezone` (string): Target timezone
- `locale` (string): Locale for formatting

#### datetime_calculate
Performs date/time calculations.

**Parameters:**
- `base` (string, required): Base date/time
- `operation` (string, required): add, subtract, diff
- `duration` (string): Duration string (e.g., "2h30m")
- `unit` (string): Unit for diff (days, hours, etc.)

#### datetime_compare
Compares two date/time values.

**Parameters:**
- `date1` (string, required): First date
- `date2` (string, required): Second date
- `precision` (string): Comparison precision
- `timezone` (string): Timezone for comparison

#### datetime_convert
Converts between timezones.

**Parameters:**
- `datetime` (string, required): Date/time to convert
- `from_timezone` (string, required): Source timezone
- `to_timezone` (string, required): Target timezone
- `format` (string): Output format

#### datetime_info
Gets detailed information about a date/time.

**Parameters:**
- `datetime` (string, required): Date/time to analyze
- `timezone` (string): Timezone for analysis
- `include_holidays` (bool): Include holiday information
- `include_solar` (bool): Include sunrise/sunset

### Data Processing Tools

#### json_process
Processes JSON data with parsing, querying, and transformation.

**Parameters:**
- `data` (string, required): JSON data
- `operation` (string, required): parse, query, transform
- `jsonpath` (string): JSONPath expression for queries
- `transform` (string): Transformation type

**Transform types:**
- extract_keys: Get all keys
- extract_values: Get all values
- flatten: Flatten nested structure
- prettify: Format with indentation
- minify: Remove whitespace

#### csv_process
Processes CSV data with parsing, filtering, and transformation.

**Parameters:**
- `data` (string, required): CSV data
- `operation` (string, required): parse, filter, transform
- `headers` (bool): First row contains headers
- `delimiter` (string): Field delimiter
- `filter` (object): Filter conditions
- `columns` ([]string): Columns to select

#### xml_process
Processes XML data with parsing and transformation.

**Parameters:**
- `data` (string, required): XML data
- `operation` (string, required): parse, query, transform
- `xpath` (string): XPath expression
- `namespaces` (map): Namespace mappings
- `output_format` (string): json, text, xml

#### data_transform
Performs general data transformations.

**Parameters:**
- `data` (interface{}, required): Input data
- `transform` (string, required): Transformation type
- `options` (map): Transform-specific options

**Transform types:**
- sort: Sort arrays/objects
- filter: Filter by conditions
- map: Apply transformations
- reduce: Aggregate data
- pivot: Pivot table operations

### Feed Processing Tools

#### feed_fetch
Fetches and parses RSS/Atom/JSON feeds.

**Parameters:**
- `url` (string, required): Feed URL
- `timeout` (int): Timeout in seconds
- `max_items` (int): Maximum items to return
- `if_modified` (string): If-Modified-Since header
- `etag` (string): ETag for conditional requests

#### feed_filter
Filters feed items based on criteria.

**Parameters:**
- `feed` (object, required): Feed data
- `criteria` (object): Filter criteria
- `date_range` (object): Date range filter
- `keywords` ([]string): Keyword filters
- `categories` ([]string): Category filters

#### feed_aggregate
Aggregates multiple feeds into one.

**Parameters:**
- `feeds` ([]string, required): Feed URLs
- `sort_by` (string): Sort field (date, title)
- `deduplicate` (bool): Remove duplicates
- `max_items` (int): Total items limit

#### feed_convert
Converts between feed formats.

**Parameters:**
- `feed` (object, required): Feed data
- `from_format` (string, required): Source format
- `to_format` (string, required): Target format
- `options` (map): Conversion options

#### feed_extract
Extracts specific data from feeds.

**Parameters:**
- `feed` (object, required): Feed data
- `extract` (string, required): What to extract
- `format` (string): Output format
- `template` (string): Extraction template

#### feed_discover
Discovers feed URLs from web pages.

**Parameters:**
- `url` (string, required): Page URL
- `types` ([]string): Feed types to find
- `follow_links` (bool): Follow page links
- `max_depth` (int): Maximum link depth

## Usage Examples

### Basic Tool Usage

```go
// Create a tool instance
tool := file.ReadFile()

// Create context
ctx := &domain.ToolContext{
    Context: context.Background(),
    State:   domain.NewState(),
    Events:  domain.NewEventEmitter(),
}

// Execute with parameters
result, err := tool.Execute(ctx, file.ReadFileParams{
    Path: "/path/to/file.txt",
    IncludeMeta: true,
})

if err != nil {
    // Handle error with guidance
    guidance := tool.ErrorGuidance()
    if hint, ok := guidance["permission denied"]; ok {
        fmt.Printf("Error hint: %s\n", hint)
    }
}
```

### Using Tools with Agents

```go
// Create an agent with tools
agent := core.NewLLMAgent(provider, "gpt-4").
    AddTool(file.ReadFile()).
    AddTool(web.HTTPRequest()).
    AddTool(math.Calculator()).
    SetSystemPrompt("You are a helpful assistant with file and web access.")

// Run agent with tool access
result, err := agent.Run(ctx, "Read the config.json file and calculate the sum of all numeric values")
```

### Tool Discovery and Dynamic Loading

```go
// Use tool discovery
discovery := tools.NewDiscovery()

// List tools by category
fileTools := discovery.ListByCategory("file")
for _, info := range fileTools {
    fmt.Printf("Tool: %s - %s\n", info.Name, info.Description)
}

// Create tools dynamically
toolMap, err := discovery.CreateTools("file_read", "calculator", "web_fetch")
```

### State-Based Tool Configuration

```go
// Configure tool behavior via state
ctx.State.Set("file_read_max_size", int64(50*1024*1024)) // 50MB limit
ctx.State.Set("file_restricted_paths", []string{"/etc", "/sys"})
ctx.State.Set("file_allowed_paths", []string{"/home/user/data"})

// Tools will respect these settings
result, err := tool.Execute(ctx, params)
```

### Event Handling

```go
// Set up event listener
ctx.Events.On("progress", func(event domain.Event) {
    progress := event.Data.(domain.ProgressData)
    fmt.Printf("Progress: %d/%d - %s\n", 
        progress.Current, progress.Total, progress.Message)
})

ctx.Events.On("file_read_complete", func(event domain.Event) {
    data := event.Data.(map[string]interface{})
    fmt.Printf("Read %d bytes in %s\n", 
        data["bytes_read"], data["elapsed_time"])
})
```

## Tool Integration Patterns

### Error Handling Pattern

```go
func executeToolSafely(tool domain.Tool, ctx *domain.ToolContext, params interface{}) (interface{}, error) {
    // Pre-execution validation
    if err := validateParams(tool, params); err != nil {
        return nil, fmt.Errorf("invalid parameters: %w", err)
    }
    
    // Execute with timeout
    ctx, cancel := context.WithTimeout(ctx.Context, 30*time.Second)
    defer cancel()
    
    // Execute tool
    result, err := tool.Execute(ctx, params)
    if err != nil {
        // Check error guidance
        for errType, guidance := range tool.ErrorGuidance() {
            if strings.Contains(err.Error(), errType) {
                return nil, fmt.Errorf("%w\nGuidance: %s", err, guidance)
            }
        }
        return nil, err
    }
    
    return result, nil
}
```

### Tool Chaining Pattern

```go
func processDataPipeline(ctx *domain.ToolContext) error {
    // Step 1: Read file
    readTool := file.ReadFile()
    data, err := readTool.Execute(ctx, file.ReadFileParams{
        Path: "/data/input.json",
    })
    if err != nil {
        return err
    }
    
    // Step 2: Process JSON
    jsonTool := data.JSONProcess()
    processed, err := jsonTool.Execute(ctx, data.JSONProcessInput{
        Data: data.(*file.ReadFileResult).Content,
        Operation: "transform",
        Transform: "flatten",
    })
    if err != nil {
        return err
    }
    
    // Step 3: Write result
    writeTool := file.WriteFile()
    _, err = writeTool.Execute(ctx, file.WriteFileParams{
        Path: "/data/output.json",
        Content: processed.(*data.JSONProcessOutput).Result.(string),
    })
    
    return err
}
```

### Conditional Tool Selection

```go
func selectToolForTask(task string) domain.Tool {
    switch {
    case strings.Contains(task, "calculate"):
        return math.Calculator()
    case strings.Contains(task, "read") && strings.Contains(task, "file"):
        return file.ReadFile()
    case strings.Contains(task, "fetch") && strings.Contains(task, "web"):
        return web.WebFetch()
    case strings.Contains(task, "current time"):
        return datetime.DateTimeNow()
    default:
        return nil
    }
}
```

### Tool Registry Extension

```go
// Create custom tool
type CustomTool struct {
    atools.BaseTool
}

func (t *CustomTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
    // Custom implementation
    return nil, nil
}

// Register custom tool
func RegisterCustomTools() {
    tools.MustRegisterTool("custom_tool", &CustomTool{
        BaseTool: atools.NewToolBuilder("custom_tool", "Custom tool description").
            WithCategory("custom").
            WithTags([]string{"custom", "example"}).
            Build().(*atools.BaseTool),
    }, tools.ToolMetadata{
        Metadata: builtins.Metadata{
            Name: "custom_tool",
            Category: "custom",
        },
        ResourceUsage: tools.ResourceInfo{
            Memory: "low",
        },
    })
}
```

## Best Practices

1. **Always provide context**: Use ToolContext with proper state and event emitters
2. **Handle errors gracefully**: Check error guidance for helpful hints
3. **Set resource limits**: Configure max sizes and timeouts via state
4. **Monitor progress**: Use event listeners for long-running operations
5. **Validate inputs**: Use parameter schemas for validation
6. **Chain tools carefully**: Consider dependencies and error propagation
7. **Use appropriate tools**: Select tools based on their categories and capabilities
8. **Respect permissions**: Tools enforce permission boundaries
9. **Cache when possible**: Some tools support caching for performance
10. **Test with examples**: Use provided examples as templates

## Summary

The Go-LLMs built-in tools provide a comprehensive toolkit for LLM agents to interact with various systems. With consistent interfaces, strong typing, and extensive metadata, these tools enable safe and efficient automation of complex tasks. The tool system's extensibility allows for custom tools while maintaining compatibility with the existing ecosystem.