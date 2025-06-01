# Built-in File Tools Example

This example provides a deep dive into the enhanced file operation capabilities of the built-in file tools.

## Overview

The built-in file tools provide advanced features beyond basic file I/O:
- **file_read**: Streaming, metadata, binary detection, line ranges
- **file_write**: Atomic operations, backups, directory creation

## Running the Example

```bash
# Optional: Set API key for agent demonstration
export OPENAI_API_KEY="your-api-key"

# Run the example
go run main.go
```

## Features Demonstrated

### 1. Enhanced File Reading
- **Large file streaming**: Handles files larger than memory
- **Binary detection**: Automatically detects binary vs text files
- **Line range reading**: Read specific line ranges (e.g., lines 10-20)
- **File metadata**: Size, permissions, modification time
- **Encoding handling**: Proper UTF-8 validation

### 2. Atomic File Writing
- **Write-rename pattern**: Ensures file integrity
- **Automatic backups**: Creates timestamped backups before overwriting
- **Directory creation**: Creates parent directories as needed
- **Custom permissions**: Set file mode on creation
- **Append mode**: Add content to existing files

### 3. Agent Integration
Shows how to use file tools with agents for complex file operations:
- Reading and summarizing files
- Transforming file content
- Creating reports from multiple files

## Example Operations

### Reading with Metadata
```go
result, _ := readTool.Execute(ctx, map[string]interface{}{
    "path": "config.json",
    "include_metadata": true,
})
```

### Line Range Reading
```go
result, _ := readTool.Execute(ctx, map[string]interface{}{
    "path": "large.log",
    "start_line": 100,
    "end_line": 200,
})
```

### Atomic Write with Backup
```go
result, _ := writeTool.Execute(ctx, map[string]interface{}{
    "path": "config.json",
    "content": newConfig,
    "create_backup": true,
    "atomic": true,
})
```

## Performance Benefits

- **Streaming**: 4KB buffer for efficient memory usage
- **Lazy loading**: Only reads what's needed
- **Concurrent safe**: Multiple agents can use tools simultaneously

## Use Cases

1. **Configuration Management**: Safe config updates with backups
2. **Log Processing**: Read specific portions of large log files
3. **Data Pipeline**: Stream processing of large datasets
4. **Report Generation**: Create files with proper error handling

## Next Steps

- Explore the [discovery example](../builtins-discovery/) to find more tools
- Check upcoming examples for web and system tools
- Review the implementation in `pkg/agent/builtins/tools/file/`