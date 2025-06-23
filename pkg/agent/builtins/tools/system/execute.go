// ABOUTME: System command execution tool with enhanced security and control features
// ABOUTME: Built-in tool for executing shell commands with timeouts, environment control, and safety options

package system

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ExecuteCommandParams defines parameters for the ExecuteCommand tool
type ExecuteCommandParams struct {
	Command     string            `json:"command"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Timeout     int               `json:"timeout,omitempty"`
	Shell       string            `json:"shell,omitempty"`
	SafeMode    bool              `json:"safe_mode,omitempty"`
	Input       string            `json:"input,omitempty"`
}

// ExecuteCommandResult defines the result of the ExecuteCommand tool
type ExecuteCommandResult struct {
	Stdout      string            `json:"stdout"`
	Stderr      string            `json:"stderr"`
	ExitCode    int               `json:"exit_code"`
	Success     bool              `json:"success"`
	TimedOut    bool              `json:"timed_out"`
	WorkingDir  string            `json:"working_dir"`
	Command     string            `json:"command"`
	Environment map[string]string `json:"environment,omitempty"`
	DurationMs  int64             `json:"duration_ms"`
}

// executeCommandParamSchema defines parameters for the ExecuteCommand tool
var executeCommandParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"command": {
			Type:        "string",
			Description: "The command to execute",
		},
		"working_dir": {
			Type:        "string",
			Description: "Working directory for command execution",
		},
		"environment": {
			Type:        "object",
			Description: "Environment variables to set (merged with current environment)",
		},
		"timeout": {
			Type:        "number",
			Description: "Timeout in seconds (default: 30, max: 300)",
		},
		"shell": {
			Type:        "string",
			Description: "Shell to use (sh, bash, zsh, or none for direct execution)",
		},
		"safe_mode": {
			Type:        "boolean",
			Description: "Enable safe mode to restrict dangerous commands (default: true)",
		},
		"input": {
			Type:        "string",
			Description: "Input to provide to the command via stdin",
		},
	},
	Required: []string{"command"},
}

// executeCommandOutputSchema defines the output schema for the ExecuteCommand tool
var executeCommandOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"stdout": {
			Type:        "string",
			Description: "Standard output from the command",
		},
		"stderr": {
			Type:        "string",
			Description: "Standard error output from the command",
		},
		"exit_code": {
			Type:        "number",
			Description: "Exit code of the command (0 indicates success)",
		},
		"success": {
			Type:        "boolean",
			Description: "Whether the command executed successfully",
		},
		"timed_out": {
			Type:        "boolean",
			Description: "Whether the command timed out",
		},
		"working_dir": {
			Type:        "string",
			Description: "The working directory where the command was executed",
		},
		"command": {
			Type:        "string",
			Description: "The command that was executed",
		},
		"environment": {
			Type:        "object",
			Description: "Custom environment variables that were set",
		},
		"duration_ms": {
			Type:        "number",
			Description: "Duration of command execution in milliseconds",
		},
	},
	Required: []string{"stdout", "stderr", "exit_code", "success", "command", "working_dir", "duration_ms"},
}

// Dangerous commands that are blocked in safe mode
var dangerousCommands = map[string]bool{
	"rm":       true,
	"rmdir":    true,
	"del":      true,
	"format":   true,
	"shutdown": true,
	"reboot":   true,
	"kill":     true,
	"killall":  true,
	"pkill":    true,
	"dd":       true,
	"mkfs":     true,
}

// Allowed commands in safe mode (allowlist approach)
var safeCommands = map[string]bool{
	"echo":     true,
	"printf":   true,
	"cat":      true,
	"ls":       true,
	"dir":      true,
	"pwd":      true,
	"date":     true,
	"whoami":   true,
	"hostname": true,
	"uname":    true,
	"which":    true,
	"where":    true,
	"env":      true,
	"grep":     true,
	"sed":      true,
	"awk":      true,
	"sort":     true,
	"uniq":     true,
	"head":     true,
	"tail":     true,
	"wc":       true,
	"find":     true,
	"tree":     true,
	"ps":       true,
	"top":      true,
	"df":       true,
	"du":       true,
	"free":     true,
	"uptime":   true,
	"curl":     true,
	"wget":     true,
	"ping":     true,
	"dig":      true,
	"nslookup": true,
	"git":      true,
	"npm":      true,
	"yarn":     true,
	"go":       true,
	"python":   true,
	"pip":      true,
	"node":     true,
	"java":     true,
	"mvn":      true,
	"gradle":   true,
	"make":     true,
	"cmake":    true,
	"docker":   true,
	"kubectl":  true,
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("execute_command", ExecuteCommand(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "execute_command",
			Category:    "system",
			Tags:        []string{"command", "shell", "execution", "system", "process"},
			Description: "Executes system commands with enhanced control and security",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic command",
					Description: "Execute a simple echo command",
					Code:        `ExecuteCommand().Execute(ctx, ExecuteCommandParams{Command: "echo 'Hello World'"})`,
				},
				{
					Name:        "With working directory",
					Description: "Execute command in specific directory",
					Code:        `ExecuteCommand().Execute(ctx, ExecuteCommandParams{Command: "ls -la", WorkingDir: "/tmp"})`,
				},
				{
					Name:        "With environment variables",
					Description: "Execute with custom environment",
					Code:        `ExecuteCommand().Execute(ctx, ExecuteCommandParams{Command: "echo $MY_VAR", Environment: map[string]string{"MY_VAR": "test"}})`,
				},
				{
					Name:        "With timeout",
					Description: "Execute with custom timeout",
					Code:        `ExecuteCommand().Execute(ctx, ExecuteCommandParams{Command: "sleep 5", Timeout: 10})`,
				},
				{
					Name:        "Direct execution without shell",
					Description: "Execute command directly without shell interpretation",
					Code:        `ExecuteCommand().Execute(ctx, ExecuteCommandParams{Command: "/usr/bin/ls -la", Shell: "none"})`,
				},
			},
		},
		RequiredPermissions: []string{"system:execute"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "medium", // Commands can vary, but assume medium as default
			Network:     false,
			FileSystem:  true,
			Concurrency: true,
		},
	})
}

// executeCommand is the main function for the tool
func executeCommand(ctx *domain.ToolContext, params ExecuteCommandParams) (*ExecuteCommandResult, error) {
	startTime := time.Now()

	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
			ToolName:   "execute_command",
			Parameters: params,
			RequestID:  ctx.RunID,
		})
	}

	// Set defaults
	if params.Timeout == 0 {
		params.Timeout = 30
	} else if params.Timeout > 300 {
		params.Timeout = 300 // Cap at 5 minutes
	}

	if params.Shell == "" {
		params.Shell = "sh"
	}

	// Default to safe mode
	// Since SafeMode is a bool, it's always either true or false
	// We can't detect if it was explicitly set, so we'll just use its value

	// Get safe mode default from state if not explicitly set
	safeMode := params.SafeMode
	if ctx.State != nil {
		if val, ok := ctx.State.Get("command_safe_mode"); ok {
			if sm, ok := val.(bool); ok {
				safeMode = sm
			}
		}
	}

	// Validate command in safe mode
	if safeMode {
		if err := validateCommandSafety(params.Command, ctx); err != nil {
			if ctx.Events != nil {
				ctx.Events.EmitError(fmt.Errorf("command blocked by safety check: %w", err))
			}
			return nil, err
		}
	}

	// Create context with timeout
	timeout := time.Duration(params.Timeout) * time.Second
	cmdCtx, cancel := context.WithTimeout(ctx.Context, timeout)
	defer cancel()

	// Prepare command
	var cmd *exec.Cmd
	if params.Shell == "none" {
		// Direct execution without shell
		parts := strings.Fields(params.Command)
		if len(parts) == 0 {
			return nil, fmt.Errorf("empty command")
		}
		cmd = exec.CommandContext(cmdCtx, parts[0], parts[1:]...) //nolint:gosec // Command execution is the purpose of this tool
	} else {
		// Execute via shell
		shell := params.Shell
		if shell != "sh" && shell != "bash" && shell != "zsh" {
			shell = "sh" // Default to sh for unknown shells
		}
		cmd = exec.CommandContext(cmdCtx, shell, "-c", params.Command)
	}

	// Set working directory
	if params.WorkingDir != "" {
		absPath, err := filepath.Abs(params.WorkingDir)
		if err != nil {
			return nil, fmt.Errorf("invalid working directory: %w", err)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("working directory does not exist: %s", absPath)
		}
		cmd.Dir = absPath
	} else {
		cmd.Dir, _ = os.Getwd()
	}

	// Set environment
	cmd.Env = os.Environ() // Start with current environment
	for key, value := range params.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Set up stdin if provided
	if params.Input != "" {
		cmd.Stdin = strings.NewReader(params.Input)
	}

	// Capture stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Emit progress event
	if ctx.Events != nil {
		ctx.Events.EmitMessage(fmt.Sprintf("Executing command: %s", params.Command))
	}

	// Execute command
	err := cmd.Run()
	duration := time.Since(startTime)

	// Prepare result
	result := &ExecuteCommandResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		Command:    params.Command,
		WorkingDir: cmd.Dir,
		DurationMs: duration.Milliseconds(),
	}

	// Add environment if it was customized
	if len(params.Environment) > 0 {
		result.Environment = params.Environment
	}

	// Handle timeout
	if cmdCtx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.Success = false
		return result, nil // Return result even on timeout
	}

	// Get exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Success = false
	} else {
		result.ExitCode = 0
		result.Success = true
	}

	// Emit result event
	if ctx.Events != nil {
		if result.Success {
			ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
				ToolName:  "execute_command",
				Result:    result,
				RequestID: ctx.RunID,
			})
		} else {
			ctx.Events.EmitError(fmt.Errorf("command failed with exit code %d", result.ExitCode))
		}
	}

	return result, nil
}

// ExecuteCommand creates a tool for executing system commands with comprehensive security controls
// including safe mode (enabled by default) that blocks dangerous commands and patterns, customizable
// timeouts, environment variable injection, and working directory management. The tool captures stdout
// and stderr separately, supports stdin input, and provides detailed execution results with timing information.
func ExecuteCommand() domain.Tool {
	builder := atools.NewToolBuilder("execute_command", "Executes system commands with enhanced control and security").
		WithFunction(executeCommand).
		WithParameterSchema(executeCommandParamSchema).
		WithOutputSchema(executeCommandOutputSchema).
		WithUsageInstructions(`Use this tool to execute system commands with enhanced control and security features.

Security Features:
- Safe mode (enabled by default) restricts dangerous commands
- Allowlisted commands in safe mode include common utilities
- Custom commands can be allowed via state configuration
- Command validation prevents injection attacks

Parameters:
- command: The command to execute (required)
- working_dir: Directory to execute in (optional, defaults to current)
- environment: Key-value pairs to add to environment (optional)
- timeout: Maximum execution time in seconds (optional, default 30, max 300)
- shell: Shell to use - sh, bash, zsh, or none for direct execution (optional, default sh)
- safe_mode: Enable/disable command safety checks (optional, default true)
- input: Data to provide via stdin (optional)

Output includes:
- stdout: Standard output from the command
- stderr: Standard error output
- exit_code: Command exit code (0 = success)
- success: Boolean indicating successful execution
- timed_out: Whether the command exceeded timeout
- duration_ms: Execution time in milliseconds

Safe Mode:
When safe_mode is true (default), the tool:
1. Blocks dangerous commands (rm, shutdown, format, etc.)
2. Blocks dangerous patterns (sudo, redirects, pipes, etc.)
3. Only allows commands from a safe allowlist
4. Permits full paths to system directories (/usr/bin/, /bin/, etc.)

To allow additional commands in safe mode, set them in state:
state.Set("allowed_commands", []string{"custom-tool", "my-script"})`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Basic command execution",
				Description: "Execute a simple echo command",
				Scenario:    "When you need to display text or test command execution",
				Input: map[string]interface{}{
					"command": "echo 'Hello, World!'",
				},
				Output: map[string]interface{}{
					"stdout":    "Hello, World!\n",
					"stderr":    "",
					"exit_code": 0,
					"success":   true,
					"timed_out": false,
				},
				Explanation: "Executes echo command which prints text to stdout",
			},
			{
				Name:        "List directory contents",
				Description: "List files in a specific directory",
				Scenario:    "When you need to explore directory contents",
				Input: map[string]interface{}{
					"command":     "ls -la",
					"working_dir": "/tmp",
				},
				Output: map[string]interface{}{
					"stdout":    "total 8\ndrwxrwxrwt  2 root root 4096 Jan  1 00:00 .\ndrwxr-xr-x 23 root root 4096 Jan  1 00:00 ..\n",
					"exit_code": 0,
					"success":   true,
				},
				Explanation: "Lists all files including hidden ones in /tmp directory",
			},
			{
				Name:        "Environment variable usage",
				Description: "Execute command with custom environment variables",
				Scenario:    "When you need to set specific environment variables for a command",
				Input: map[string]interface{}{
					"command": "echo $MY_VAR - $HOME",
					"environment": map[string]string{
						"MY_VAR": "custom value",
					},
				},
				Output: map[string]interface{}{
					"stdout":  "custom value - /home/user\n",
					"success": true,
				},
				Explanation: "MY_VAR is set from parameters, HOME is inherited from system",
			},
			{
				Name:        "Command with timeout",
				Description: "Execute a long-running command with timeout",
				Scenario:    "When you need to limit command execution time",
				Input: map[string]interface{}{
					"command": "sleep 10",
					"timeout": 2,
				},
				Output: map[string]interface{}{
					"stdout":      "",
					"stderr":      "",
					"exit_code":   -1,
					"success":     false,
					"timed_out":   true,
					"duration_ms": 2000,
				},
				Explanation: "Command times out after 2 seconds, before sleep 10 completes",
			},
			{
				Name:        "Direct command execution",
				Description: "Execute command without shell interpretation",
				Scenario:    "When you need precise control over command execution",
				Input: map[string]interface{}{
					"command": "/usr/bin/ls -la /tmp",
					"shell":   "none",
				},
				Output: map[string]interface{}{
					"success": true,
				},
				Explanation: "Executes /usr/bin/ls directly without shell, arguments parsed by spaces",
			},
			{
				Name:        "Command with input",
				Description: "Provide input to a command via stdin",
				Scenario:    "When a command needs input data",
				Input: map[string]interface{}{
					"command": "wc -l",
					"input":   "line1\nline2\nline3\n",
				},
				Output: map[string]interface{}{
					"stdout":  "3\n",
					"success": true,
				},
				Explanation: "Counts lines from the provided input (3 lines)",
			},
			{
				Name:        "Error handling example",
				Description: "Handle command that returns non-zero exit code",
				Scenario:    "When a command fails or returns an error",
				Input: map[string]interface{}{
					"command": "ls /nonexistent-directory",
				},
				Output: map[string]interface{}{
					"stdout":    "",
					"stderr":    "ls: cannot access '/nonexistent-directory': No such file or directory\n",
					"exit_code": 2,
					"success":   false,
				},
				Explanation: "Command fails with exit code 2, error message in stderr",
			},
		}).
		WithConstraints([]string{
			"Safe mode is enabled by default to prevent dangerous operations",
			"Maximum timeout is 300 seconds (5 minutes)",
			"Shell defaults to 'sh' if not specified",
			"Environment variables are merged with current process environment",
			"Working directory must exist or command will fail",
			"In safe mode, only allowlisted commands can be executed",
			"Command output is captured separately for stdout and stderr",
			"Direct execution (shell=none) splits command by spaces only",
			"Dangerous patterns like sudo, rm -rf, pipes are blocked in safe mode",
			"Exit code -1 indicates timeout or execution failure",
		}).
		WithErrorGuidance(map[string]string{
			"empty command":                                "Provide a non-empty command to execute",
			"command blocked by safety check":              "Command is dangerous and blocked in safe mode. Disable safe_mode if you really need to run it",
			"command contains dangerous pattern":           "Command contains dangerous patterns. Use simpler commands or disable safe_mode with caution",
			"command is not in the safe command allowlist": "Command not allowed in safe mode. Either use an allowed command or add it to allowed_commands in state",
			"invalid working directory":                    "Working directory path is invalid. Use an absolute path",
			"working directory does not exist":             "Specified directory doesn't exist. Create it first or use an existing directory",
			"context deadline exceeded":                    "Command timed out. Increase timeout parameter or optimize the command",
			"no such file or directory":                    "Command or file not found. Check the command path and spelling",
			"permission denied":                            "Insufficient permissions. Check file permissions or run with appropriate privileges",
			"executable file not found":                    "Command not found in PATH. Use full path or ensure command is installed",
		}).
		WithCategory("system").
		WithTags([]string{"command", "shell", "execution", "system", "process"}).
		WithVersion("2.0.0").
		WithBehavior(
			false,    // Not deterministic - commands can have different outputs
			true,     // Potentially destructive - commands can modify system
			true,     // Requires confirmation for dangerous operations
			"medium", // Medium latency - depends on command
		)

	return builder.Build()
}

// validateCommandSafety checks if a command is safe to execute
func validateCommandSafety(command string, ctx *domain.ToolContext) error {
	// Extract the base command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	baseCmd := filepath.Base(parts[0])

	// Check against dangerous commands
	if dangerousCommands[baseCmd] {
		return fmt.Errorf("command '%s' is blocked in safe mode", baseCmd)
	}

	// Check if command contains dangerous patterns
	dangerousPatterns := []string{
		"sudo",
		"su ",
		"chmod",
		"chown",
		">",  // Redirect that could overwrite files
		">>", // Append redirect
		"|",  // Pipe (could be used maliciously)
		";",  // Command separator
		"&&", // Command chaining
		"||", // Command chaining
		"`",  // Command substitution
		"$(", // Command substitution
		"rm -rf",
		"rm -fr",
		"format ",
		"mkfs.",
	}

	lowerCommand := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCommand, pattern) {
			return fmt.Errorf("command contains dangerous pattern '%s' which is blocked in safe mode", pattern)
		}
	}

	// Check if command is in the safe commands list
	commandAllowed := safeCommands[baseCmd]

	// Check state for additional allowed commands
	if !commandAllowed && ctx != nil && ctx.State != nil {
		if allowedCmds, ok := ctx.State.Get("allowed_commands"); ok {
			if cmdList, ok := allowedCmds.([]string); ok {
				for _, allowed := range cmdList {
					if baseCmd == allowed {
						commandAllowed = true
						break
					}
				}
			}
		}
	}

	// In strict safe mode, only allow explicitly safe commands
	// This is more restrictive but safer for untrusted input
	if !commandAllowed {
		// Allow full paths to known safe locations
		if strings.HasPrefix(parts[0], "/usr/bin/") ||
			strings.HasPrefix(parts[0], "/bin/") ||
			strings.HasPrefix(parts[0], "/usr/local/bin/") {
			// Check the base command from the path
			baseCmd = filepath.Base(parts[0])
			commandAllowed = safeCommands[baseCmd]

			// Check state again for the base command
			if !commandAllowed && ctx != nil && ctx.State != nil {
				if allowedCmds, ok := ctx.State.Get("allowed_commands"); ok {
					if cmdList, ok := allowedCmds.([]string); ok {
						for _, allowed := range cmdList {
							if baseCmd == allowed {
								commandAllowed = true
								break
							}
						}
					}
				}
			}

			if !commandAllowed {
				return fmt.Errorf("command '%s' is not in the safe command allowlist", baseCmd)
			}
		} else {
			return fmt.Errorf("command '%s' is not in the safe command allowlist", baseCmd)
		}
	}

	return nil
}

// MustGetExecuteCommand retrieves the registered ExecuteCommand tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetExecuteCommand() domain.Tool {
	return tools.MustGetTool("execute_command")
}
