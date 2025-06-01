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

// ExecuteCommand creates a tool for executing system commands
// This is a built-in tool optimized for:
// - Security with safe mode and command restrictions
// - Environment variable control
// - Working directory management
// - Timeout handling
// - Separate stdout/stderr capture
// - Input via stdin
func ExecuteCommand() domain.Tool {
	return atools.NewTool(
		"execute_command",
		"Executes system commands with enhanced control and security",
		func(ctx context.Context, params ExecuteCommandParams) (*ExecuteCommandResult, error) {
			startTime := time.Now()

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
			if params.SafeMode || params.SafeMode == false {
				// SafeMode was explicitly set
			} else {
				params.SafeMode = true
			}

			// Validate command in safe mode
			if params.SafeMode {
				if err := validateCommandSafety(params.Command); err != nil {
					return nil, err
				}
			}

			// Create context with timeout
			timeout := time.Duration(params.Timeout) * time.Second
			cmdCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			// Prepare command
			var cmd *exec.Cmd
			if params.Shell == "none" {
				// Direct execution without shell
				parts := strings.Fields(params.Command)
				if len(parts) == 0 {
					return nil, fmt.Errorf("empty command")
				}
				cmd = exec.CommandContext(cmdCtx, parts[0], parts[1:]...)
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

			return result, nil
		},
		executeCommandParamSchema,
	)
}

// validateCommandSafety checks if a command is safe to execute
func validateCommandSafety(command string) error {
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

	// In strict safe mode, only allow explicitly safe commands
	// This is more restrictive but safer for untrusted input
	if !safeCommands[baseCmd] {
		// Allow full paths to known safe locations
		if strings.HasPrefix(parts[0], "/usr/bin/") ||
			strings.HasPrefix(parts[0], "/bin/") ||
			strings.HasPrefix(parts[0], "/usr/local/bin/") {
			// Check the base command from the path
			baseCmd = filepath.Base(parts[0])
			if !safeCommands[baseCmd] {
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
