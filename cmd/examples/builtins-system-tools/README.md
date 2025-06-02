# Built-in System Tools Example

This example demonstrates the system tools available in the go-llms library.

## Overview

The built-in system tools provide functionality for:
- Executing system commands with safety controls
- Reading environment variables with pattern matching
- Getting comprehensive system information
- Listing and filtering running processes

## Running the Example

```bash
go run main.go
```

## Available System Tools

1. **execute_command** - Run system commands
   - Shell selection (sh, bash, zsh, or direct)
   - Environment variable support
   - Working directory control
   - Timeout management
   - Safe mode with command allowlist/blocklist
   - Separate stdout/stderr capture

2. **get_environment_variable** - Read environment variables
   - Pattern matching (prefix*, *suffix, *contains*)
   - Sensitive variable masking
   - Optional value hiding
   - Sorted output

3. **get_system_info** - System information
   - OS and architecture details
   - CPU and memory statistics
   - Go runtime information
   - Environment summary

4. **process_list** - List running processes
   - Cross-platform support
   - Name-based filtering
   - Sorting by PID, name, CPU, or memory
   - Result limiting
   - Include/exclude self

## Example Usage

### Execute Command
```go
execTool := tools.MustGetTool("execute_command")
result, _ := execTool.Execute(ctx, map[string]interface{}{
    "command": "ls -la",
    "timeout": 5,
    "safe_mode": true,
})
```

### Execute with Environment
```go
result, _ := execTool.Execute(ctx, map[string]interface{}{
    "command": "echo Hello $NAME",
    "shell": "bash",
    "env": map[string]string{
        "NAME": "World",
    },
    "working_dir": "/tmp",
})
```

### Get Environment Variables
```go
envTool := tools.MustGetTool("get_environment_variable")
result, _ := envTool.Execute(ctx, map[string]interface{}{
    "pattern": "PATH*",  // All PATH-related variables
})
```

### Get System Info
```go
sysInfoTool := tools.MustGetTool("get_system_info")
result, _ := sysInfoTool.Execute(ctx, map[string]interface{}{})
```

### List Processes
```go
procTool := tools.MustGetTool("process_list")
result, _ := procTool.Execute(ctx, map[string]interface{}{
    "filter": "chrome",
    "sort_by": "memory",
    "limit": 10,
})
```

## Key Features

### Command Execution
- **Safety Controls**: Safe mode prevents dangerous commands
- **Shell Support**: Choose between different shells or direct execution
- **Environment**: Set custom environment variables
- **Working Directory**: Execute commands in specific directories
- **Timeout**: Prevent long-running commands (max 5 minutes)
- **Output Capture**: Separate stdout and stderr

### Environment Variables
- **Pattern Matching**: Find variables by pattern
- **Security**: Automatic masking of sensitive values (API keys, tokens)
- **Flexibility**: Show or hide values as needed

### System Information
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Runtime Info**: Go version, goroutines, memory usage
- **Resource Stats**: CPU count, memory statistics

### Process Management
- **Filtering**: Find processes by name
- **Sorting**: Order by various metrics
- **Performance**: Limit results for efficiency

## Security Considerations

### Command Execution Safety
- Safe mode blocks dangerous commands by default
- Commands are sanitized to prevent injection
- Timeout prevents resource exhaustion
- Working directory restrictions

### Environment Variable Security
- Sensitive variables are automatically masked
- Pattern matching prevents accidental exposure
- Option to hide values completely

## Integration with Agents

System tools can be used with agents for automation:

```go
agent := workflow.NewAgent(
    "system-admin",
    provider,
    workflow.WithTools(
        tools.MustGetTool("execute_command"),
        tools.MustGetTool("get_system_info"),
        tools.MustGetTool("process_list"),
    ),
)

// Agent can now perform system administration tasks
response, _ := agent.Run(ctx, workflow.UserMessage(
    "Check if the web server is running and report system resources",
))
```

## Platform-Specific Notes

- **Windows**: Some commands may differ (e.g., `dir` instead of `ls`)
- **macOS**: Process information may be limited without elevated privileges
- **Linux**: Full process information available to the current user

## Best Practices

1. Always use safe mode unless you need specific dangerous commands
2. Set appropriate timeouts for commands
3. Use pattern matching carefully with environment variables
4. Limit process list results for better performance
5. Handle platform differences in your commands