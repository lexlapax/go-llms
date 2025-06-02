# Built-in Tools Examples

This directory contains examples demonstrating the built-in tools system. The examples are organized by focus area:

## builtins-discovery/
**Focus**: Registry discovery and basic usage
- How to discover available built-in tools
- Search and filter tools by category/tags
- Basic tool usage with agents
- Migration from custom tools to built-ins

**When to use**: Start here to understand the registry system and available tools.

## builtins-file-tools/
**Focus**: Deep dive into file tool capabilities
- Enhanced file reading (streaming, metadata, line ranges)
- Atomic file writing with backups
- Binary file detection
- Large file handling
- Agent integration for file operations

**When to use**: When you need to understand the full capabilities of file tools.

## builtins-datetime-tools/
**Focus**: Comprehensive date/time operations
- Current time in various formats and timezones
- Date parsing with auto-detection and relative dates
- Date arithmetic and business day calculations
- Formatting with localization support
- Timezone conversions with DST handling
- Date comparisons and sorting

**When to use**: When you need to work with dates, times, and timezones in your applications.

## builtins-web-tools/
**Focus**: Web interaction and HTTP operations
- Web fetching with timeouts and headers
- Web search using DuckDuckGo
- Web scraping with CSS selectors
- Advanced HTTP requests (all methods, auth, custom headers)
- Response metadata and timing information

**When to use**: When you need to interact with web services, APIs, or scrape web content.

## builtins-system-tools/
**Focus**: System interaction and management
- Command execution with safety controls and timeouts
- Environment variable access with pattern matching
- Comprehensive system information gathering
- Process listing and filtering
- Cross-platform compatibility

**When to use**: When you need to interact with the operating system, run commands, or gather system information.

## builtins-data-tools/
**Focus**: Structured data processing
- JSON processing with JSONPath queries
- CSV parsing, filtering, and statistics
- XML parsing with XPath and JSON conversion
- Common data transformations (filter, map, reduce, sort, group)
- Type conversions and aggregations

**When to use**: When you need to process, transform, or analyze structured data in various formats.

## Using Built-in Tools

All built-in tools follow the same pattern:

1. **Import the category** to trigger registration:
   ```go
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
   ```

2. **Discover available tools**:
   ```go
   tools.Tools.List()                    // All tools
   tools.Tools.ListByCategory("file")    // By category
   tools.Tools.Search("read")            // By search term
   ```

3. **Use the tools**:
   ```go
   tool, _ := tools.GetTool("file_read")
   result, err := tool.Execute(ctx, params)
   ```

## Benefits Over Custom Tools

- **Consistency**: Standardized interfaces across all projects
- **Features**: Enhanced capabilities (streaming, timeouts, metadata)
- **Discovery**: Easy to find and understand available tools
- **Maintenance**: Updates and fixes handled centrally
- **Performance**: Optimized implementations with pooling