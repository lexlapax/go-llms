# Built-in System Tools Example

This example demonstrates the system tools available in the go-llms library, showing both direct tool usage and LLM agent integration.

## Overview

The built-in system tools provide functionality for:
- Executing system commands with safety controls
- Reading environment variables with pattern matching
- Getting comprehensive system information
- Listing and filtering running processes

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

## Available System Tools

1. **execute_command** (v2.0.0) - Execute system commands safely
   - Shell selection (sh, bash, zsh, or direct)
   - Environment variable support
   - Working directory control
   - Timeout management
   - Safe mode with command allowlist/blocklist
   - Separate stdout/stderr capture
   - ⚠️ Destructive operation
   - ⚠️ Requires confirmation

2. **get_environment_variable** (v2.0.0) - Get environment variables
   - Pattern matching (prefix*, *suffix, *contains*)
   - Sensitive variable masking
   - Optional value hiding
   - Sorted output

3. **get_system_info** (v2.0.0) - Get system information
   - OS and architecture details
   - CPU and memory statistics
   - Go runtime information
   - Environment summary

4. **process_list** (v2.0.0) - List running processes
   - Cross-platform support
   - Name-based filtering
   - Sorting by PID, name, CPU, or memory
   - Result limiting
   - Include/exclude self

## Example Modes

### Direct Tool Usage

The default mode demonstrates direct usage of all system tools:

```go
// Get all system tools
execTool, _ := tools.GetTool("execute_command")
envTool, _ := tools.GetTool("get_environment_variable")
sysInfoTool, _ := tools.GetTool("get_system_info")
procTool, _ := tools.GetTool("process_list")

// Execute a command
result, err := execTool.Execute(toolCtx, map[string]interface{}{
    "command": "echo 'Hello from system tools!'",
    "timeout": 5,
})
```

### LLM Agent Mode

The `-llm` flag demonstrates using system tools with an LLM agent:

```go
// Create LLM agent with system tools
agent := core.NewLLMAgent("system-assistant", "System Information Assistant", deps)
agent.AddTool(execTool)
agent.AddTool(envTool)
agent.AddTool(sysInfoTool)
agent.AddTool(procTool)

// The agent uses minimal prompting and relies on tool documentation
agent.SetSystemPrompt(`You are a helpful system information assistant...`)
```

The LLM mode includes example queries like:
- "What operating system and architecture is this running on?"
- "Show me the current memory usage statistics"
- "What Go-related environment variables are set?"
- "What are the top 3 processes using the most memory?"

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

### Command Execution Safety
- Safe mode blocks dangerous commands by default
- Commands are sanitized to prevent injection
- Timeout prevents resource exhaustion (max 5 minutes)
- Working directory restrictions

### Environment Variable Security
- Sensitive variables are automatically masked
- Pattern matching prevents accidental exposure
- Option to hide values completely

## Security Considerations

### Command Execution
- Safe mode is recommended for production use
- Dangerous commands require explicit allowlisting
- All commands have timeout limits
- Shell injection protection is built-in

### Environment Variables
- API keys and tokens are automatically masked
- Use pattern matching carefully to avoid exposing secrets
- Consider using no_values option for listing

## Platform-Specific Notes

- **Windows**: Some commands may differ (e.g., `dir` instead of `ls`)
- **macOS**: Process information may be limited without elevated privileges
- **Linux**: Full process information available to the current user

## Integration with New Architecture

This example uses the new agent architecture:

```go
// Uses core.NewLLMAgent instead of workflow.NewAgent
agent := core.NewLLMAgent("system-assistant", "System Information Assistant", deps)

// Tools are added individually
agent.AddTool(tool)

// State-based execution
state := domain.NewState()
state.Set("user_input", prompt)
result, err := agent.Run(ctx, state)
```

## Best Practices

1. Always use safe mode unless you need specific dangerous commands
2. Set appropriate timeouts for commands
3. Use pattern matching carefully with environment variables
4. Limit process list results for better performance
5. Handle platform differences in your commands
6. Let tools guide the LLM with their built-in documentation
7. Use DEBUG=1 to see detailed agent execution logs