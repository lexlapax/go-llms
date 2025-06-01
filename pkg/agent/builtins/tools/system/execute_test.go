// ABOUTME: Tests for the ExecuteCommand built-in tool
// ABOUTME: Validates command execution, safety checks, timeouts, and environment handling

package system

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
)

// Helper function to convert tool result to map
func resultToMap(t *testing.T, result interface{}) map[string]interface{} {
	// Try direct map assertion first
	if resultMap, ok := result.(map[string]interface{}); ok {
		return resultMap
	}
	
	// Try JSON bytes
	if resultBytes, ok := result.([]byte); ok {
		var resultData map[string]interface{}
		if err := json.Unmarshal(resultBytes, &resultData); err != nil {
			t.Fatalf("Failed to unmarshal result bytes: %v", err)
		}
		return resultData
	}
	
	// Try ExecuteCommandResult struct
	if execResult, ok := result.(*ExecuteCommandResult); ok {
		jsonBytes, err := json.Marshal(execResult)
		if err != nil {
			t.Fatalf("Failed to marshal ExecuteCommandResult: %v", err)
		}
		var resultData map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &resultData); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}
		return resultData
	}
	
	t.Fatalf("Cannot convert result to map, unexpected type: %T", result)
	return nil
}

func TestExecuteCommandRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("execute_command")
	if !ok {
		t.Fatal("ExecuteCommand tool not registered")
	}
	if tool == nil {
		t.Fatal("ExecuteCommand tool is nil")
	}

	// Test tool name
	if tool.Name() != "execute_command" {
		t.Errorf("Expected tool name 'execute_command', got '%s'", tool.Name())
	}

	// Test that we can retrieve it via MustGetTool
	mustGetTool := tools.MustGetTool("execute_command")
	if mustGetTool == nil {
		t.Fatal("MustGetTool returned nil")
	}
}

func TestExecuteCommandBasic(t *testing.T) {
	tool := ExecuteCommand()
	ctx := context.Background()

	testCases := []struct {
		name        string
		params      map[string]interface{}
		checkStdout string
		checkStderr string
		expectError bool
	}{
		{
			name: "Simple echo command",
			params: map[string]interface{}{
				"command": "echo 'Hello World'",
			},
			checkStdout: "Hello World",
			expectError: false,
		},
		{
			name: "Command with exit code",
			params: map[string]interface{}{
				"command": "echo 'test' && exit 0",
			},
			checkStdout: "test",
			expectError: false,
		},
		{
			name: "Print working directory",
			params: map[string]interface{}{
				"command": "pwd",
			},
			expectError: false,
		},
		{
			name: "List files",
			params: map[string]interface{}{
				"command": "ls",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, tc.params)
			if err != nil && !tc.expectError {
				t.Fatalf("Unexpected error: %v", err)
			}
			if err == nil && tc.expectError {
				t.Fatal("Expected error but got none")
			}

			if err == nil && result != nil {
				resultMap := resultToMap(t, result)

				// Check stdout if specified
				if tc.checkStdout != "" {
					stdout, _ := resultMap["stdout"].(string)
					if !strings.Contains(stdout, tc.checkStdout) {
						t.Errorf("Expected stdout to contain '%s', got '%s'", tc.checkStdout, stdout)
					}
				}

				// Check stderr if specified
				if tc.checkStderr != "" {
					stderr, _ := resultMap["stderr"].(string)
					if !strings.Contains(stderr, tc.checkStderr) {
						t.Errorf("Expected stderr to contain '%s', got '%s'", tc.checkStderr, stderr)
					}
				}

				// Verify success flag
				success, _ := resultMap["success"].(bool)
				if tc.expectError && success {
					t.Error("Expected success to be false")
				}
				if !tc.expectError && !success {
					t.Error("Expected success to be true")
				}
			}
		})
	}
}

func TestExecuteCommandSafety(t *testing.T) {
	tool := ExecuteCommand()
	ctx := context.Background()

	// Test dangerous commands that should be blocked in safe mode
	dangerousCommands := []string{
		"rm -rf /tmp/test",
		"sudo ls",
		"chmod 777 /tmp",
		"echo 'test' > /tmp/file",
		"cat /etc/passwd | grep root",
		"ls; rm file",
		"ls && rm file",
		"echo `rm file`",
		"echo $(rm file)",
	}

	for _, cmd := range dangerousCommands {
		t.Run("Blocking: "+cmd, func(t *testing.T) {
			params := map[string]interface{}{
				"command":   cmd,
				"safe_mode": true,
			}
			_, err := tool.Execute(ctx, params)
			if err == nil {
				t.Errorf("Expected error for dangerous command '%s', but got none", cmd)
			}
		})
	}

	// Test that safe commands are allowed
	safeCommands := []string{
		"echo 'test'",
		"ls",
		"pwd",
		"date",
		"whoami",
	}

	for _, cmd := range safeCommands {
		t.Run("Allowing: "+cmd, func(t *testing.T) {
			params := map[string]interface{}{
				"command":   cmd,
				"safe_mode": true,
			}
			_, err := tool.Execute(ctx, params)
			if err != nil {
				t.Errorf("Unexpected error for safe command '%s': %v", cmd, err)
			}
		})
	}
}

func TestExecuteCommandTimeout(t *testing.T) {
	tool := ExecuteCommand()
	ctx := context.Background()

	// Use appropriate sleep command based on OS
	sleepCmd := "sleep 3"
	if runtime.GOOS == "windows" {
		sleepCmd = "timeout /t 3"
	}

	params := map[string]interface{}{
		"command": sleepCmd,
		"timeout": 1, // 1 second timeout
	}

	start := time.Now()
	result, err := tool.Execute(ctx, params)
	duration := time.Since(start)

	// Should not return an error, but result should indicate timeout
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that it actually timed out around 1 second
	if duration > 2*time.Second {
		t.Errorf("Command took too long: %v", duration)
	}

	// Check result
	if result != nil {
		resultMap := resultToMap(t, result)
		if timedOut, _ := resultMap["timed_out"].(bool); !timedOut {
			t.Error("Expected timed_out to be true")
		}
		if success, _ := resultMap["success"].(bool); success {
			t.Error("Expected success to be false for timed out command")
		}
	}
}

func TestExecuteCommandEnvironment(t *testing.T) {
	tool := ExecuteCommand()
	ctx := context.Background()

	// Use appropriate echo command based on OS
	echoCmd := "echo $TEST_VAR"
	if runtime.GOOS == "windows" {
		echoCmd = "echo %TEST_VAR%"
	}

	params := map[string]interface{}{
		"command": echoCmd,
		"environment": map[string]interface{}{
			"TEST_VAR": "test_value_123",
		},
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that environment variable was set
	resultMap := resultToMap(t, result)
	stdout, _ := resultMap["stdout"].(string)
	if !strings.Contains(stdout, "test_value_123") {
		t.Errorf("Expected stdout to contain 'test_value_123', got '%s'", stdout)
	}
}

func TestExecuteCommandWorkingDirectory(t *testing.T) {
	tool := ExecuteCommand()
	ctx := context.Background()

	// Test with temp directory
	params := map[string]interface{}{
		"command":     "pwd",
		"working_dir": "/tmp",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that working directory was used
	resultMap := resultToMap(t, result)
	stdout, _ := resultMap["stdout"].(string)
	workingDir, _ := resultMap["working_dir"].(string)
	
	// Verify working_dir in result
	if !strings.Contains(workingDir, "tmp") {
		t.Errorf("Expected working_dir to contain 'tmp', got '%s'", workingDir)
	}
	
	// Verify pwd output
	if !strings.Contains(stdout, "tmp") {
		t.Errorf("Expected pwd output to contain 'tmp', got '%s'", stdout)
	}
}

func TestExecuteCommandInput(t *testing.T) {
	tool := ExecuteCommand()
	ctx := context.Background()

	// Use cat to echo stdin
	params := map[string]interface{}{
		"command": "cat",
		"input":   "Hello from stdin\nLine 2",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that input was provided via stdin
	resultMap := resultToMap(t, result)
	stdout, _ := resultMap["stdout"].(string)
	if !strings.Contains(stdout, "Hello from stdin") {
		t.Errorf("Expected stdout to contain 'Hello from stdin', got '%s'", stdout)
	}
	if !strings.Contains(stdout, "Line 2") {
		t.Errorf("Expected stdout to contain 'Line 2', got '%s'", stdout)
	}
}

func TestExecuteCommandDirectExecution(t *testing.T) {
	tool := ExecuteCommand()
	ctx := context.Background()

	// Test direct execution without shell
	params := map[string]interface{}{
		"command": "/bin/echo test direct",
		"shell":   "none",
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check output
	resultMap := resultToMap(t, result)
	stdout, _ := resultMap["stdout"].(string)
	if !strings.Contains(stdout, "test direct") {
		t.Errorf("Expected stdout to contain 'test direct', got '%s'", stdout)
	}
}

func TestValidateCommandSafety(t *testing.T) {
	testCases := []struct {
		command     string
		shouldError bool
		description string
	}{
		// Safe commands
		{"echo test", false, "Simple echo"},
		{"ls -la", false, "List files"},
		{"pwd", false, "Print working directory"},
		{"date", false, "Show date"},
		{"git status", false, "Git command"},
		{"/usr/bin/echo test", false, "Full path to safe command"},
		
		// Dangerous commands
		{"rm -rf /", true, "Dangerous rm command"},
		{"sudo apt-get update", true, "Sudo command"},
		{"chmod 777 file", true, "Permission change"},
		{"echo test > file", true, "File redirect"},
		{"cat file | grep test", true, "Pipe command"},
		{"ls; rm file", true, "Command separator"},
		{"ls && rm file", true, "Command chaining"},
		{"echo `date`", true, "Command substitution"},
		{"echo $(date)", true, "Command substitution"},
		{"format c:", true, "Format command"},
		{"shutdown -h now", true, "Shutdown command"},
		{"kill -9 1234", true, "Kill command"},
		{"dd if=/dev/zero of=/dev/sda", true, "DD command"},
		{"unknown_command", true, "Unknown command not in allowlist"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := validateCommandSafety(tc.command)
			if tc.shouldError && err == nil {
				t.Errorf("Expected error for command '%s' but got none", tc.command)
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Unexpected error for command '%s': %v", tc.command, err)
			}
		})
	}
}