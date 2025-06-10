# Built-in File Tools Example

This example demonstrates all 6 file operation tools available in the go-llms library, showing both direct tool usage and LLM agent integration.

## Overview

The built-in file tools provide advanced features beyond basic file I/O:
- **file_read**: Streaming, metadata, binary detection, line ranges
- **file_write**: Atomic operations, backups, directory creation, append mode
- **file_list**: Directory listing with metadata and filtering
- **file_search**: Pattern matching and content search capabilities
- **file_move**: Safe file relocation with validation
- **file_delete**: Safe file removal with existence checks

## Running the Example

### Direct Tool Usage (Default)
```bash
go run main.go
```

### LLM Agent Mode
```bash
go run main.go -llm
```

### Tool Information Display
```bash
go run main.go -llm info
```

### With Debug Logging
```bash
DEBUG=1 go run main.go -llm
```

## Available File Tools

1. **file_read** (v2.0.0) - Read files with advanced features
   - Large file streaming support
   - Binary file detection
   - Line range reading
   - File metadata extraction
   - UTF-8 encoding validation

2. **file_write** (v2.0.0) - Write files safely
   - Atomic write operations
   - Automatic backup creation
   - Directory auto-creation
   - Append mode support
   - Custom permissions
   - ⚠️ Destructive operation

3. **file_list** (v2.0.0) - List directory contents
   - Recursive directory traversal
   - Pattern-based filtering
   - File metadata collection
   - Size calculations

4. **file_search** (v2.0.0) - Search files and content
   - File name pattern matching
   - Content search within files
   - Line number tracking
   - Recursive search support

5. **file_move** (v2.0.0) - Move/rename files
   - Safe file relocation
   - Cross-filesystem support
   - Validation and error handling
   - Byte counting
   - ⚠️ Destructive operation

6. **file_delete** (v2.0.0) - Delete files safely
   - Existence validation
   - Safe removal with error handling
   - Path validation
   - ⚠️ Destructive operation
   - ⚠️ Requires confirmation

## Example Modes

### Direct Tool Usage

The default mode demonstrates direct usage of all file tools:

```go
// Get all file tools
readTool, _ := tools.GetTool("file_read")
writeTool, _ := tools.GetTool("file_write")
listTool, _ := tools.GetTool("file_list")
searchTool, _ := tools.GetTool("file_search")
moveTool, _ := tools.GetTool("file_move")
deleteTool, _ := tools.GetTool("file_delete")

// Write a configuration file
result, err := writeTool.Execute(toolCtx, map[string]interface{}{
    "path":    "config.json",
    "content": configContent,
    "atomic":  true,
})
```

### LLM Agent Mode

The `-llm` flag demonstrates using file tools with an LLM agent:

```go
// Create LLM agent with file tools
agent := core.NewLLMAgent("file-assistant", "File Management Assistant", deps)
agent.AddTool(readTool)
agent.AddTool(writeTool)
agent.AddTool(listTool)
agent.AddTool(searchTool)
agent.AddTool(moveTool)
agent.AddTool(deleteTool)

// The agent uses minimal prompting and relies on tool documentation
agent.SetSystemPrompt(`You are a helpful file management assistant...`)
```

The LLM mode includes example queries like:
- "Read the configuration file and tell me what settings are enabled"
- "Create a todo list file with 5 important tasks"
- "List all JSON files and tell me what each one is for"
- "Search for any errors in log files"
- "Rename settings.json to app-settings.json"

## Key Features

### Two Execution Modes
- **Direct Mode**: Demonstrates all tool capabilities with comprehensive examples
- **LLM Mode**: Shows how tools integrate with LLM agents using minimal prompting

### Mock Provider Support
- If no API keys are set, the example uses a mock provider
- Mock provider simulates tool usage for demonstration purposes
- Set ANTHROPIC_API_KEY, OPENAI_API_KEY, or GEMINI_API_KEY for real LLM usage

### Tool Documentation
- All tools include comprehensive metadata and usage instructions
- Tools provide examples, constraints, and error guidance
- LLM agents can leverage this built-in documentation

## Example Operations

### Reading with Metadata
```go
result, _ := readTool.Execute(ctx, map[string]interface{}{
    "path":         "config.json",
    "include_meta": true,
})
```

### Line Range Reading
```go
result, _ := readTool.Execute(ctx, map[string]interface{}{
    "path":       "large.log",
    "line_start": 100,
    "line_end":   200,
})
```

### Atomic Write with Backup
```go
result, _ := writeTool.Execute(ctx, map[string]interface{}{
    "path":    "config.json",
    "content": newConfig,
    "backup":  true,
    "atomic":  true,
})
```

### Content Search
```go
result, _ := searchTool.Execute(ctx, map[string]interface{}{
    "path":         "/path/to/search",
    "pattern":      "error",      // Search pattern (content to find)
    "file_pattern": "*.log",      // File name pattern
})
```

## Important Parameter Names

Each file tool uses specific parameter names that must be matched exactly:

### file_read Tool
- `path`: File path to read (required)
- `include_meta`: Include file metadata
- `line_start`: Starting line number for range reading
- `line_end`: Ending line number for range reading

### file_write Tool
- `path`: File path to write (required)
- `content`: Content to write (required)
- `atomic`: Use atomic write operation
- `backup`: Create backup before overwriting
- `append`: Append to existing file
- `mode`: File permissions (Unix mode)
- `create_dirs`: Create parent directories if needed

### file_list Tool
- `path`: Directory path to list (required)
- `pattern`: Glob pattern for filtering
- `recursive`: Recursively list subdirectories

### file_search Tool
- `path`: Directory path to search (required)
- `pattern`: Search pattern (content to find)
- `file_pattern`: Glob pattern for file name matching
- `recursive`: Search recursively

### file_move Tool
- `source`: Source file path (required)
- `destination`: Destination file path (required)

### file_delete Tool
- `path`: File path to delete (required)

## Performance Benefits

- **Streaming**: 4KB buffer for efficient memory usage
- **Lazy loading**: Only reads what's needed
- **Concurrent safe**: Multiple agents can use tools simultaneously
- **Memory efficient**: Large files don't consume excessive memory

## Security Considerations

- **Atomic writes**: Prevent partial file corruption
- **Backup creation**: Preserve original data before modifications
- **Path validation**: Prevent directory traversal attacks
- **Safe deletion**: Confirmation required for destructive operations

## Integration with New Architecture

This example uses the new agent architecture:

```go
// Uses core.NewLLMAgent instead of workflow.NewAgent
agent := core.NewLLMAgent("file-assistant", "File Management Assistant", deps)

// Tools are added individually
agent.AddTool(tool)

// State-based execution
state := domain.NewState()
state.Set("user_input", prompt)
result, err := agent.Run(ctx, state)
```

## Use Cases

1. **Configuration Management**: Safe config updates with backups
2. **Log Processing**: Read specific portions of large log files
3. **Data Pipeline**: Stream processing of large datasets
4. **Report Generation**: Create files with proper error handling
5. **File System Maintenance**: Organize, move, and clean up files
6. **Content Analysis**: Search and analyze file contents

## Best Practices

1. Always use atomic writes for critical files
2. Create backups before modifying important data
3. Use line ranges for large file processing
4. Validate paths before file operations
5. Handle errors appropriately in production
6. Let tools guide the LLM with their built-in documentation
7. Use DEBUG=1 to see detailed agent execution logs

## Next Steps

- Explore the [system tools example](../builtins-system-tools/) for system operations
- Check the [data tools example](../builtins-data-tools/) for data processing
- Review the [built-in components guide](../../../docs/user-guide/built-in-components.md) for all tools
- Examine tool source code in `pkg/agent/builtins/tools/file/` for advanced usage