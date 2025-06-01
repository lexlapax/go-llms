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

## Future Built-in Examples (Planned)

### builtins-web-tools/
Will demonstrate:
- Web fetching with timeouts and headers
- Web search across multiple engines
- Web scraping (when implemented)
- HTTP request tools (POST, PUT, etc.)

### builtins-system-tools/
Will demonstrate:
- Command execution with safety controls
- Environment variable access
- System information gathering

## Using Built-in Tools

All built-in tools follow the same pattern:

1. **Import the category** to trigger registration:
   ```go
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
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