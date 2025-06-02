// ABOUTME: Example demonstrating the use of built-in system tools
// ABOUTME: Shows command execution, environment variables, system info, and process management

package main

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
)

func main() {
	ctx := context.Background()

	// List all system tools
	fmt.Println("=== Available System Tools ===")
	systemTools := tools.Tools.ListByCategory("system")
	for _, entry := range systemTools {
		fmt.Printf("- %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
	}
	fmt.Println()

	// Example 1: Get System Information
	fmt.Println("=== Example 1: System Information ===")
	sysInfoTool := tools.MustGetTool("get_system_info")
	sysInfoResult, err := sysInfoTool.Execute(ctx, map[string]interface{}{})
	if err != nil {
		log.Fatalf("Failed to get system info: %v", err)
	}
	fmt.Printf("System information: %+v\n\n", sysInfoResult)

	// Example 2: Get Environment Variables
	fmt.Println("=== Example 2: Environment Variables ===")
	envTool := tools.MustGetTool("get_environment_variable")

	// Get all PATH-related variables
	pathResult, err := envTool.Execute(ctx, map[string]interface{}{
		"pattern": "PATH*",
	})
	if err != nil {
		log.Printf("Failed to get PATH variables: %v", err)
	} else {
		fmt.Printf("PATH variables: %+v\n", pathResult)
	}

	// Get Go-related variables with values hidden
	goResult, err := envTool.Execute(ctx, map[string]interface{}{
		"pattern":   "*GO*",
		"no_values": true,
	})
	if err != nil {
		log.Printf("Failed to get Go variables: %v", err)
	} else {
		fmt.Printf("Go variables (names only): %+v\n\n", goResult)
	}

	// Example 3: Execute Commands
	fmt.Println("=== Example 3: Execute Commands ===")
	execTool := tools.MustGetTool("execute_command")

	// Simple command - list files
	var listCmd string
	if runtime.GOOS == "windows" {
		listCmd = "dir"
	} else {
		listCmd = "ls -la"
	}

	listResult, err := execTool.Execute(ctx, map[string]interface{}{
		"command":   listCmd,
		"timeout":   5,
		"safe_mode": true,
	})
	if err != nil {
		log.Printf("Failed to list files: %v", err)
	} else {
		fmt.Printf("Directory listing:\n%+v\n", listResult)
	}

	// Command with environment variables
	echoResult, err := execTool.Execute(ctx, map[string]interface{}{
		"command": "echo Hello from $USER at $(date)",
		"shell":   "bash",
		"env": map[string]string{
			"USER": "go-llms-example",
		},
		"timeout": 5,
	})
	if err != nil {
		log.Printf("Failed to run echo command: %v", err)
	} else {
		fmt.Printf("Echo result: %+v\n\n", echoResult)
	}

	// Example 4: Process List
	fmt.Println("=== Example 4: Process List ===")
	procTool := tools.MustGetTool("process_list")

	// Get all Go processes
	goProcsResult, err := procTool.Execute(ctx, map[string]interface{}{
		"filter":       "go",
		"sort_by":      "cpu",
		"include_self": true,
		"limit":        10,
	})
	if err != nil {
		log.Printf("Failed to list processes: %v", err)
	} else {
		fmt.Printf("Go processes: %+v\n", goProcsResult)
	}

	// Get top 5 processes by memory
	topMemResult, err := procTool.Execute(ctx, map[string]interface{}{
		"sort_by": "memory",
		"limit":   5,
	})
	if err != nil {
		log.Printf("Failed to get top memory processes: %v", err)
	} else {
		fmt.Printf("Top 5 processes by memory: %+v\n\n", topMemResult)
	}

	// Example 5: Safe Command Execution
	fmt.Println("=== Example 5: Safe Command Execution ===")

	// Try to run a command with safe mode
	safeResult, err := execTool.Execute(ctx, map[string]interface{}{
		"command":   "pwd",
		"safe_mode": true,
		"timeout":   5,
	})
	if err != nil {
		log.Printf("Failed to run safe command: %v", err)
	} else {
		fmt.Printf("Current directory: %+v\n", safeResult)
	}

	// Example 6: Command with Working Directory
	fmt.Println("=== Example 6: Command with Working Directory ===")

	// Create a temp file in /tmp
	tempResult, err := execTool.Execute(ctx, map[string]interface{}{
		"command":     "echo 'Hello from temp' > test.txt && cat test.txt",
		"shell":       "bash",
		"working_dir": "/tmp",
		"timeout":     5,
	})
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
	} else {
		fmt.Printf("Temp file result: %+v\n", tempResult)
	}

	// Clean up
	_, _ = execTool.Execute(ctx, map[string]interface{}{
		"command": "rm -f /tmp/test.txt",
		"timeout": 5,
	})
}
