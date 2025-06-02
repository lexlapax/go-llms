# Built-in File Tools Example

This comprehensive example demonstrates all 6 file operation tools available in the go-llms library, showcasing their full capabilities with practical examples.

## Overview

The built-in file tools provide advanced features beyond basic file I/O:
- **file_read**: Streaming, metadata, binary detection, line ranges
- **file_write**: Atomic operations, backups, directory creation, append mode
- **file_list**: Directory listing with metadata and filtering
- **file_search**: Pattern matching and content search capabilities
- **file_move**: Safe file relocation with validation
- **file_delete**: Safe file removal with existence checks

## Running the Example

```bash
# Optional: Set API key for agent demonstration
export OPENAI_API_KEY="your-api-key"

# Run the example
go run main.go
```

## Features Demonstrated

### 1. Enhanced File Reading (file_read)
- **Large file streaming**: Handles files larger than memory
- **Binary detection**: Automatically detects binary vs text files
- **Line range reading**: Read specific line ranges (e.g., lines 10-20)
- **File metadata**: Size, permissions, modification time, directory status
- **Encoding handling**: Proper UTF-8 validation

### 2. Atomic File Writing (file_write)
- **Write-rename pattern**: Ensures file integrity
- **Automatic backups**: Creates timestamped backups before overwriting
- **Directory creation**: Creates parent directories as needed
- **Custom permissions**: Set file mode on creation
- **Append mode**: Add content to existing files

### 3. Directory Operations (file_list)
- **Recursive listing**: Traverse directory trees
- **File filtering**: Filter by type, size, or modification time
- **Metadata collection**: Get comprehensive file information
- **Total size calculation**: Sum up directory contents

### 4. File Search (file_search)
- **Pattern matching**: Use glob patterns for file selection
- **Content search**: Search within file contents
- **Line number tracking**: Show exact match locations
- **Recursive search**: Search through directory hierarchies

### 5. File Movement (file_move)
- **Safe relocation**: Validates source and destination
- **Cross-directory moves**: Handle moves across filesystems
- **Existence checking**: Prevents overwriting without confirmation
- **Byte counting**: Track exactly how much data was moved

### 6. File Deletion (file_delete)
- **Existence validation**: Check if file exists before deletion
- **Safe removal**: Proper error handling and reporting
- **Path validation**: Ensure deletion targets are correct

### 7. Agent Integration
Shows how to use file tools with agents for complex file operations:
- Reading and summarizing files
- Transforming file content
- Creating reports from multiple files

## Example Operations

### Reading with Metadata
```go
result, _ := readTool.Execute(ctx, map[string]interface{}{
    "path":         "config.json",
    "include_meta": true,  // Fixed parameter name
})
if readResult, ok := result.(*file.ReadFileResult); ok {
    fmt.Printf("Content: %s\n", readResult.Content)
    fmt.Printf("Size: %d bytes\n", readResult.Metadata.Size)
}
```

### Line Range Reading
```go
result, _ := readTool.Execute(ctx, map[string]interface{}{
    "path":       "large.log",
    "line_start": 100,  // Fixed parameter names
    "line_end":   200,
})
```

### Atomic Write with Backup
```go
result, _ := writeTool.Execute(ctx, map[string]interface{}{
    "path":    "config.json",
    "content": newConfig,
    "backup":  true,  // Fixed parameter name
    "atomic":  true,
})
```

### Directory Listing
```go
result, _ := listTool.Execute(ctx, map[string]interface{}{
    "path":      "/path/to/directory",
    "recursive": false,
})
if listResult, ok := result.(*file.FileListResult); ok {
    fmt.Printf("Found %d files\n", len(listResult.Files))
}
```

### Pattern Search
```go
result, _ := searchTool.Execute(ctx, map[string]interface{}{
    "path":    "/path/to/search",
    "pattern": "*.json",
})
```

### Content Search
```go
result, _ := searchTool.Execute(ctx, map[string]interface{}{
    "path":         "/path/to/search",
    "pattern":      "error",      // Search pattern (content to find)
    "file_pattern": "*.txt",      // File name pattern
})
```

### File Move
```go
result, _ := moveTool.Execute(ctx, map[string]interface{}{
    "source":      "/path/to/source.txt",
    "destination": "/path/to/destination.txt",  // Fixed parameter name
})
```

### File Delete
```go
result, _ := deleteTool.Execute(ctx, map[string]interface{}{
    "path": "/path/to/file.txt",
})
```

## Important Parameter Names

Each file tool uses specific parameter names that must be matched exactly:

### file_read Tool
- `path`: File path to read (required)
- `include_meta`: Include file metadata (not `include_metadata`)
- `line_start`: Starting line number for range reading
- `line_end`: Ending line number for range reading

### file_write Tool
- `path`: File path to write (required)
- `content`: Content to write (required)
- `atomic`: Use atomic write operation
- `backup`: Create backup before overwriting
- `append`: Append to existing file

### file_list Tool
- `path`: Directory path to list (required)
- `recursive`: Recursively list subdirectories

### file_search Tool
- `path`: Directory path to search (required)
- `pattern`: Search pattern (content to find)
- `file_pattern`: Glob pattern for file name matching (e.g., "*.txt")
- `recursive`: Search recursively

### file_move Tool
- `source`: Source file path (required)
- `destination`: Destination file path (required)

### file_delete Tool
- `path`: File path to delete (required)

## Type Assertions

When handling file tool outputs, use the correct struct types:

```go
// file_read results
if readResult, ok := result.(*file.ReadFileResult); ok {
    fmt.Printf("Content: %s\n", readResult.Content)
    fmt.Printf("Lines: %d\n", readResult.Lines)
    if readResult.Metadata != nil {
        fmt.Printf("Size: %d\n", readResult.Metadata.Size)
    }
}

// file_write results
if writeResult, ok := result.(*file.WriteFileResult); ok {
    fmt.Printf("Bytes written: %d\n", writeResult.BytesWritten)
    fmt.Printf("Backup path: %s\n", writeResult.BackupPath)
}

// file_list results
if listResult, ok := result.(*file.FileListResult); ok {
    fmt.Printf("Found %d files\n", len(listResult.Files))
    fmt.Printf("Total size: %d bytes\n", listResult.TotalSize)
}

// file_search results
if searchResult, ok := result.(*file.FileSearchResult); ok {
    fmt.Printf("Found %d matches\n", len(searchResult.Matches))
    for _, match := range searchResult.Matches {
        fmt.Printf("- %s (%d bytes)\n", match.Path, match.Size)
    }
}

// file_move results
if moveResult, ok := result.(*file.FileMoveResult); ok {
    fmt.Printf("Moved from %s to %s\n", moveResult.SourcePath, moveResult.DestPath)
    fmt.Printf("Bytes moved: %d\n", moveResult.BytesMoved)
}

// file_delete results
if deleteResult, ok := result.(*file.FileDeleteResult); ok {
    fmt.Printf("Deleted: %s (existed: %v)\n", deleteResult.Path, deleteResult.Existed)
}
```

## Performance Benefits

- **Streaming**: 4KB buffer for efficient memory usage
- **Lazy loading**: Only reads what's needed
- **Concurrent safe**: Multiple agents can use tools simultaneously
- **Memory efficient**: Large files don't consume excessive memory

## Use Cases

1. **Configuration Management**: Safe config updates with backups
2. **Log Processing**: Read specific portions of large log files
3. **Data Pipeline**: Stream processing of large datasets
4. **Report Generation**: Create files with proper error handling
5. **File System Maintenance**: Organize, move, and clean up files
6. **Content Analysis**: Search and analyze file contents

## Integration with Agents

```go
agent := workflow.NewAgent(provider).
    SetSystemPrompt("You are a file management assistant.").
    AddTool(tools.MustGetTool("file_read")).
    AddTool(tools.MustGetTool("file_write")).
    AddTool(tools.MustGetTool("file_list")).
    AddTool(tools.MustGetTool("file_search")).
    AddTool(tools.MustGetTool("file_move")).
    AddTool(tools.MustGetTool("file_delete"))

result, _ := agent.Run(ctx, "Organize all .txt files in /tmp into subdirectories by creation date")
```

## Next Steps

- Explore the [discovery example](../builtins-discovery/) to find more tools
- Check the [data tools example](../builtins-data-tools/) for data processing
- Review the [built-in components guide](../../../docs/user-guide/built-in-components.md) for all tools
- Examine tool source code in `pkg/agent/builtins/tools/file/` for advanced usage