# Built-in Tools Reference: Complete Tool Catalog

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Reference](/docs/user-guide/reference/) / Built-in Tools**

Comprehensive catalog of all built-in tools available in Go-LLMs, organized by category with detailed parameters, examples, and best practices.

## Tool System Overview

Go-LLMs provides 30+ built-in tools across 6 major categories, enabling agents to interact with files, web services, systems, data formats, and more.

### Tool Interface

All tools implement the standard interface:

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() interface{}
    OutputSchema() interface{}
    Execute(ctx context.Context, input interface{}) (interface{}, error)
}
```

### Using Tools with Agents

```go
// Basic tool usage
agent := core.NewLLMAgent("assistant", provider, core.WithTools(
    tools.NewFileReadTool(),
    tools.NewWebFetchTool(),
    tools.NewCalculatorTool(),
))

// Tool discovery
registry := discover.NewToolRegistry()
allTools := registry.ListTools()
```

---

## File System Tools

### file_read
**Purpose:** Read contents of files from the file system

**Parameters:**
- `path` (string, required): Path to the file to read
- `encoding` (string, optional): File encoding (default: "utf-8")

**Example:**
```go
tool := tools.NewFileReadTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "path": "/path/to/file.txt",
})
```

**Output:** File contents as string

**Best Practices:**
- Always check file existence before reading
- Handle large files with streaming when possible
- Validate file permissions

---

### file_write
**Purpose:** Write content to files on the file system

**Parameters:**
- `path` (string, required): Path where to write the file
- `content` (string, required): Content to write
- `mode` (string, optional): Write mode - "write", "append" (default: "write")
- `encoding` (string, optional): File encoding (default: "utf-8")

**Example:**
```go
tool := tools.NewFileWriteTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "path": "/path/to/output.txt",
    "content": "Hello, World!",
    "mode": "append",
})
```

**Output:** Success confirmation with bytes written

**Best Practices:**
- Create parent directories if they don't exist
- Use append mode for logs
- Implement file locking for concurrent writes

---

### file_list
**Purpose:** List files and directories at a given path

**Parameters:**
- `path` (string, required): Directory path to list
- `recursive` (bool, optional): List recursively (default: false)
- `pattern` (string, optional): Glob pattern to filter files

**Example:**
```go
tool := tools.NewFileListTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "path": "/home/user/documents",
    "recursive": true,
    "pattern": "*.pdf",
})
```

**Output:** Array of file information objects

**Best Practices:**
- Use patterns to filter large directories
- Limit recursion depth for performance
- Handle permission errors gracefully

---

### file_search
**Purpose:** Search for files matching specific criteria

**Parameters:**
- `path` (string, required): Starting directory for search
- `pattern` (string, optional): Name pattern to match
- `content` (string, optional): Content to search within files
- `modified_after` (string, optional): ISO 8601 date string
- `size_greater_than` (int, optional): Minimum file size in bytes

**Example:**
```go
tool := tools.NewFileSearchTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "path": "/var/log",
    "pattern": "*.log",
    "content": "ERROR",
    "modified_after": "2024-01-01T00:00:00Z",
})
```

**Output:** Array of matching file paths with metadata

---

### file_move
**Purpose:** Move or rename files and directories

**Parameters:**
- `source` (string, required): Source file path
- `destination` (string, required): Destination file path
- `overwrite` (bool, optional): Overwrite if exists (default: false)

**Example:**
```go
tool := tools.NewFileMoveTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "source": "/tmp/data.csv",
    "destination": "/data/processed/data_2024.csv",
})
```

---

### file_delete
**Purpose:** Delete files or directories

**Parameters:**
- `path` (string, required): Path to delete
- `recursive` (bool, optional): Delete directories recursively (default: false)

**Example:**
```go
tool := tools.NewFileDeleteTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "path": "/tmp/old_files",
    "recursive": true,
})
```

---

## Web Tools

### http_request
**Purpose:** Make HTTP requests to web services

**Parameters:**
- `url` (string, required): Target URL
- `method` (string, optional): HTTP method (default: "GET")
- `headers` (map[string]string, optional): Request headers
- `body` (string, optional): Request body
- `timeout` (int, optional): Timeout in seconds (default: 30)

**Example:**
```go
tool := tools.NewHTTPRequestTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "url": "https://api.example.com/data",
    "method": "POST",
    "headers": map[string]string{
        "Content-Type": "application/json",
        "Authorization": "Bearer token",
    },
    "body": `{"query": "test"}`,
})
```

**Output:** Response with status, headers, and body

**Best Practices:**
- Always set appropriate timeouts
- Handle rate limiting with retries
- Validate SSL certificates in production

---

### web_fetch
**Purpose:** Fetch and parse web pages

**Parameters:**
- `url` (string, required): URL to fetch
- `selector` (string, optional): CSS selector to extract specific content
- `wait_for` (string, optional): CSS selector to wait for (JavaScript sites)
- `user_agent` (string, optional): Custom user agent

**Example:**
```go
tool := tools.NewWebFetchTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "url": "https://example.com/article",
    "selector": "article.main-content",
})
```

**Output:** Extracted text content or full page HTML

---

### web_scrape
**Purpose:** Advanced web scraping with pattern extraction

**Parameters:**
- `url` (string, required): URL to scrape
- `patterns` (map[string]string, required): Named extraction patterns
- `pagination` (object, optional): Pagination configuration
- `rate_limit` (int, optional): Requests per second limit

**Example:**
```go
tool := tools.NewWebScrapeTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "url": "https://shop.example.com/products",
    "patterns": map[string]string{
        "title": "h1.product-title",
        "price": "span.price",
        "description": "div.description",
    },
    "pagination": map[string]interface{}{
        "next_selector": "a.next-page",
        "max_pages": 10,
    },
})
```

---

### web_search
**Purpose:** Search the web using search engines

**Parameters:**
- `query` (string, required): Search query
- `num_results` (int, optional): Number of results (default: 10)
- `search_type` (string, optional): "web", "image", "news" (default: "web")
- `language` (string, optional): Language code

**Example:**
```go
tool := tools.NewWebSearchTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "query": "golang concurrency patterns",
    "num_results": 20,
})
```

---

### api_client
**Purpose:** Interact with REST APIs with authentication

**Parameters:**
- `base_url` (string, required): API base URL
- `endpoint` (string, required): API endpoint path
- `method` (string, optional): HTTP method
- `auth_type` (string, optional): "bearer", "basic", "api_key"
- `auth_value` (string, optional): Authentication value
- `params` (map[string]string, optional): Query parameters
- `body` (interface{}, optional): Request body

**Example:**
```go
tool := tools.NewAPIClientTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "base_url": "https://api.github.com",
    "endpoint": "/repos/golang/go/issues",
    "auth_type": "bearer",
    "auth_value": "github_token",
    "params": map[string]string{
        "state": "open",
        "labels": "bug",
    },
})
```

---

## System Tools

### get_system_info
**Purpose:** Retrieve system information

**Parameters:**
- `info_type` (string, optional): "all", "cpu", "memory", "disk", "network"

**Example:**
```go
tool := tools.NewSystemInfoTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "info_type": "memory",
})
```

**Output:** System information based on requested type

---

### execute_command
**Purpose:** Execute system commands (use with caution)

**Parameters:**
- `command` (string, required): Command to execute
- `args` ([]string, optional): Command arguments
- `timeout` (int, optional): Timeout in seconds
- `working_dir` (string, optional): Working directory

**Example:**
```go
tool := tools.NewCommandExecutorTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "command": "grep",
    "args": []string{"-r", "ERROR", "/var/log"},
    "timeout": 60,
})
```

**Security Warning:** This tool can execute arbitrary commands. Always validate inputs and use with extreme caution.

---

### get_environment_variable
**Purpose:** Read environment variables

**Parameters:**
- `name` (string, required): Variable name
- `default` (string, optional): Default value if not found

**Example:**
```go
tool := tools.NewEnvVarTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "name": "API_ENDPOINT",
    "default": "https://api.example.com",
})
```

---

### process_list
**Purpose:** List running processes

**Parameters:**
- `filter` (string, optional): Filter by process name
- `sort_by` (string, optional): "cpu", "memory", "pid"

**Example:**
```go
tool := tools.NewProcessListTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "filter": "python",
    "sort_by": "memory",
})
```

---

## Math Tools

### calculator
**Purpose:** Perform mathematical calculations

**Parameters:**
- `expression` (string, required): Mathematical expression to evaluate
- `precision` (int, optional): Decimal precision (default: 2)

**Example:**
```go
tool := tools.NewCalculatorTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "expression": "(100 * 1.21) / 2 + sqrt(16)",
    "precision": 4,
})
```

**Supported Operations:**
- Basic: +, -, *, /, %, ^
- Functions: sqrt, sin, cos, tan, log, ln, abs, ceil, floor, round
- Constants: pi, e

---

## Date/Time Tools

### datetime_now
**Purpose:** Get current date and time

**Parameters:**
- `timezone` (string, optional): Timezone (default: "UTC")
- `format` (string, optional): Output format

**Example:**
```go
tool := tools.NewDateTimeNowTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "timezone": "America/New_York",
    "format": "2006-01-02 15:04:05",
})
```

---

### datetime_parse
**Purpose:** Parse date/time strings

**Parameters:**
- `date_string` (string, required): Date string to parse
- `format` (string, optional): Input format
- `timezone` (string, optional): Timezone

**Example:**
```go
tool := tools.NewDateTimeParseTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "date_string": "2024-01-15 14:30:00",
    "format": "2006-01-02 15:04:05",
})
```

---

### datetime_format
**Purpose:** Format date/time values

**Parameters:**
- `timestamp` (string/int, required): Unix timestamp or ISO string
- `format` (string, required): Output format
- `timezone` (string, optional): Target timezone

---

### datetime_calculate
**Purpose:** Perform date/time calculations

**Parameters:**
- `base_date` (string, required): Starting date
- `operation` (string, required): "add" or "subtract"
- `duration` (string, required): Duration (e.g., "2d", "3h", "1w")

**Example:**
```go
tool := tools.NewDateTimeCalculateTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "base_date": "2024-01-15",
    "operation": "add",
    "duration": "30d",
})
```

---

### datetime_compare
**Purpose:** Compare two dates

**Parameters:**
- `date1` (string, required): First date
- `date2` (string, required): Second date
- `unit` (string, optional): Comparison unit (days, hours, minutes)

---

### datetime_convert
**Purpose:** Convert between timezones

**Parameters:**
- `datetime` (string, required): Date/time to convert
- `from_timezone` (string, required): Source timezone
- `to_timezone` (string, required): Target timezone

---

### datetime_info
**Purpose:** Get detailed date/time information

**Parameters:**
- `datetime` (string, required): Date/time to analyze
- `timezone` (string, optional): Timezone for analysis

**Output:** Detailed breakdown including day of week, week number, quarter, etc.

---

## Data Processing Tools

### json_process
**Purpose:** Process and transform JSON data

**Parameters:**
- `data` (interface{}, required): JSON data to process
- `operation` (string, required): Operation to perform
- `path` (string, optional): JSONPath expression
- `value` (interface{}, optional): Value for set operations

**Operations:**
- `get`: Extract value at path
- `set`: Set value at path
- `delete`: Remove value at path
- `transform`: Apply transformation
- `validate`: Validate against schema

**Example:**
```go
tool := tools.NewJSONProcessTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "data": jsonData,
    "operation": "get",
    "path": "$.users[?(@.age > 18)].name",
})
```

---

### csv_process
**Purpose:** Process CSV data

**Parameters:**
- `data` (string, required): CSV data or file path
- `operation` (string, required): Processing operation
- `headers` (bool, optional): Has headers (default: true)
- `delimiter` (string, optional): Field delimiter (default: ",")

**Operations:**
- `parse`: Convert to structured data
- `filter`: Filter rows by condition
- `transform`: Transform columns
- `aggregate`: Perform aggregations
- `convert`: Convert to other formats

**Example:**
```go
tool := tools.NewCSVProcessTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "data": "/path/to/data.csv",
    "operation": "filter",
    "condition": "age > 25 AND department = 'Sales'",
})
```

---

### xml_process
**Purpose:** Process XML data

**Parameters:**
- `data` (string, required): XML data
- `operation` (string, required): Processing operation
- `xpath` (string, optional): XPath expression
- `namespaces` (map[string]string, optional): Namespace mappings

**Operations:**
- `parse`: Convert to structured data
- `query`: Query with XPath
- `transform`: Apply XSLT transformation
- `validate`: Validate against XSD

---

### data_transform
**Purpose:** General data transformation tool

**Parameters:**
- `data` (interface{}, required): Input data
- `from_format` (string, required): Source format
- `to_format` (string, required): Target format
- `options` (map[string]interface{}, optional): Format-specific options

**Supported Formats:**
- JSON, CSV, XML, YAML, TOML
- SQL, Markdown tables
- Custom formats with templates

**Example:**
```go
tool := tools.NewDataTransformTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "data": csvData,
    "from_format": "csv",
    "to_format": "json",
    "options": map[string]interface{}{
        "pretty": true,
        "camel_case": true,
    },
})
```

---

## Feed Processing Tools

### feed_fetch
**Purpose:** Fetch RSS/Atom feeds

**Parameters:**
- `url` (string, required): Feed URL
- `format` (string, optional): Expected format (auto-detected)
- `limit` (int, optional): Maximum items to fetch

**Example:**
```go
tool := tools.NewFeedFetchTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "url": "https://example.com/feed.xml",
    "limit": 20,
})
```

---

### feed_filter
**Purpose:** Filter feed items by criteria

**Parameters:**
- `feed` (interface{}, required): Feed data
- `criteria` (map[string]interface{}, required): Filter criteria
- `mode` (string, optional): "include" or "exclude"

**Filter Options:**
- `keywords`: Match keywords in title/content
- `authors`: Filter by author
- `categories`: Filter by category
- `date_range`: Filter by publication date

---

### feed_aggregate
**Purpose:** Aggregate multiple feeds

**Parameters:**
- `feeds` ([]string, required): List of feed URLs
- `dedup` (bool, optional): Remove duplicates (default: true)
- `sort` (string, optional): Sort order ("date", "title")
- `limit` (int, optional): Total items limit

---

### feed_convert
**Purpose:** Convert between feed formats

**Parameters:**
- `feed` (interface{}, required): Feed data
- `to_format` (string, required): Target format (rss, atom, json)
- `version` (string, optional): Format version

---

### feed_extract
**Purpose:** Extract specific data from feeds

**Parameters:**
- `feed` (interface{}, required): Feed data
- `fields` ([]string, required): Fields to extract
- `format` (string, optional): Output format

---

### feed_discover
**Purpose:** Discover feeds from a website

**Parameters:**
- `url` (string, required): Website URL
- `types` ([]string, optional): Feed types to discover

**Example:**
```go
tool := tools.NewFeedDiscoverTool()
result, err := tool.Execute(ctx, map[string]interface{}{
    "url": "https://blog.example.com",
    "types": []string{"rss", "atom"},
})
```

---

## Tool Best Practices

### Error Handling
```go
result, err := tool.Execute(ctx, input)
if err != nil {
    switch e := err.(type) {
    case *ValidationError:
        // Handle validation errors
    case *TimeoutError:
        // Handle timeouts
    default:
        // Handle other errors
    }
}
```

### Tool Chaining
```go
// Chain tools for complex operations
files, _ := fileListTool.Execute(ctx, map[string]interface{}{
    "path": "/data",
    "pattern": "*.json",
})

for _, file := range files.([]string) {
    data, _ := fileReadTool.Execute(ctx, map[string]interface{}{
        "path": file,
    })
    
    processed, _ := jsonProcessTool.Execute(ctx, map[string]interface{}{
        "data": data,
        "operation": "transform",
    })
}
```

### Parallel Execution
```go
// Execute tools in parallel for performance
var wg sync.WaitGroup
results := make([]interface{}, len(urls))

for i, url := range urls {
    wg.Add(1)
    go func(idx int, u string) {
        defer wg.Done()
        results[idx], _ = webFetchTool.Execute(ctx, map[string]interface{}{
            "url": u,
        })
    }(i, url)
}

wg.Wait()
```

### Custom Tool Creation
```go
type CustomTool struct {
    name string
}

func (t *CustomTool) Name() string {
    return t.name
}

func (t *CustomTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Custom implementation
    return nil, nil
}
```

---

## Performance Considerations

### Caching
- Implement caching for expensive operations
- Use TTL-based cache invalidation
- Consider distributed caching for scale

### Rate Limiting
- Respect external API rate limits
- Implement internal rate limiting for resource protection
- Use exponential backoff for retries

### Resource Management
- Set appropriate timeouts for all operations
- Limit concurrent operations
- Monitor memory usage for large data processing

### Optimization Tips
- Batch operations where possible
- Use streaming for large files
- Implement connection pooling for HTTP tools
- Profile tool execution for bottlenecks

---

## Security Guidelines

### Input Validation
- Always validate and sanitize inputs
- Use parameterized queries for database operations
- Validate file paths to prevent directory traversal

### Authentication
- Store credentials securely (environment variables, secrets manager)
- Use least privilege principle
- Rotate credentials regularly

### Network Security
- Validate SSL certificates
- Use secure protocols (HTTPS)
- Implement request signing where needed

### File System Security
- Restrict file operations to allowed directories
- Validate file types and sizes
- Implement virus scanning for uploads

---

## Next Steps

- **[Configuration Reference](configuration-reference.md)** - Detailed configuration options
- **[Error Codes Reference](error-codes-reference.md)** - Complete error handling guide
- **[Best Practices Checklist](best-practices-checklist.md)** - Production readiness
- **[Creating Tools Guide](/docs/user-guide/guides/agent-tools.md)** - Build custom tools
- **[Tool Examples](/docs/user-guide/examples/)** - Real-world tool usage