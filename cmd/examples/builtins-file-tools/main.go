// ABOUTME: Example demonstrating the enhanced built-in file tools
// ABOUTME: Shows file reading with metadata, line ranges, and atomic writing with backups

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	// Import built-in file tools - this triggers auto-registration
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
	fmt.Println("=== File Tools Example ===")
	fmt.Println()

	// Demonstrate tool discovery
	fmt.Println("Available file tools:")
	fileTools := tools.Tools.ListByCategory("file")
	for _, entry := range fileTools {
		fmt.Printf("  - %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
		fmt.Printf("    Version: %s\n", entry.Metadata.Version)
		fmt.Printf("    Tags: %v\n", entry.Metadata.Tags)
		fmt.Println()
	}

	// Get the tools
	readTool, _ := tools.GetTool("file_read")
	writeTool, _ := tools.GetTool("file_write")

	// Demonstrate direct tool usage
	ctx := context.Background()
	// Use a temp directory to avoid leaving files if interrupted
	demoDir, err := os.MkdirTemp("", "file_tools_demo_*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(demoDir) // Clean up
	fmt.Printf("Using demo directory: %s\n\n", demoDir)

	// Example 1: Write a file with atomic operation and backup
	fmt.Println("=== Example 1: Atomic Write ===")
	configFile := filepath.Join(demoDir, "config.json")

	// Write initial config
	initialConfig := `{
  "version": "1.0",
  "settings": {
    "debug": false
  }
}`

	result, err := writeTool.Execute(ctx, map[string]interface{}{
		"path":    configFile,
		"content": initialConfig,
	})
	if err != nil {
		log.Printf("Error writing initial config: %v", err)
	} else {
		writeResult := result.(*file.WriteFileResult)
		fmt.Printf("Initial write successful: %d bytes written\n", writeResult.BytesWritten)
	}

	// Update config with backup
	updatedConfig := `{
  "version": "1.1",
  "settings": {
    "debug": true,
    "log_level": "info"
  }
}`

	result, err = writeTool.Execute(ctx, map[string]interface{}{
		"path":    configFile,
		"content": updatedConfig,
		"atomic":  true,
		"backup":  true,
	})
	if err != nil {
		log.Printf("Error updating config: %v", err)
	} else {
		writeResult := result.(*file.WriteFileResult)
		fmt.Printf("Update successful with atomic write\n")
		fmt.Printf("  Backup created at: %v\n", writeResult.BackupPath)
	}

	// Example 2: Read file with metadata
	fmt.Println("\n=== Example 2: Read with Metadata ===")
	result, err = readTool.Execute(ctx, map[string]interface{}{
		"path":         configFile,
		"include_meta": true,
	})
	if err != nil {
		log.Printf("Error reading config: %v", err)
	} else {
		readResult := result.(*file.ReadFileResult)
		fmt.Printf("File content:\n%s\n", readResult.Content)
		if readResult.Metadata != nil {
			fmt.Printf("File metadata:\n")
			fmt.Printf("  Size: %v bytes\n", readResult.Metadata.Size)
			fmt.Printf("  Extension: %v\n", readResult.Metadata.Extension)
			fmt.Printf("  Modified: %v\n", readResult.Metadata.ModTime)
		}
	}

	// Example 3: Append to log file
	fmt.Println("\n=== Example 3: Append Mode ===")
	logFile := filepath.Join(demoDir, "app.log")

	// Write initial log entry
	writeTool.Execute(ctx, map[string]interface{}{
		"path":    logFile,
		"content": "2024-01-31 10:00:00 - Application started\n",
	})

	// Append more log entries
	writeTool.Execute(ctx, map[string]interface{}{
		"path":    logFile,
		"content": "2024-01-31 10:00:01 - Configuration loaded\n",
		"append":  true,
	})

	writeTool.Execute(ctx, map[string]interface{}{
		"path":    logFile,
		"content": "2024-01-31 10:00:02 - Server listening on port 8080\n",
		"append":  true,
	})

	// Read the full log
	result, err = readTool.Execute(ctx, map[string]interface{}{
		"path": logFile,
	})
	if err != nil {
		log.Printf("Error reading log: %v", err)
	} else {
		readResult := result.(*file.ReadFileResult)
		fmt.Printf("Full log:\n%s", readResult.Content)
	}

	// Example 4: Read specific lines from large file
	fmt.Println("\n=== Example 4: Line Range Reading ===")
	largeFile := filepath.Join(demoDir, "large.txt")

	// Create a file with many lines
	var content string
	for i := 1; i <= 20; i++ {
		content += fmt.Sprintf("Line %d: This is some content on line %d\n", i, i)
	}
	writeTool.Execute(ctx, map[string]interface{}{
		"path":    largeFile,
		"content": content,
	})

	// Read only lines 5-10
	result, err = readTool.Execute(ctx, map[string]interface{}{
		"path":       largeFile,
		"line_start": 5,
		"line_end":   10,
	})
	if err != nil {
		log.Printf("Error reading lines: %v", err)
	} else {
		readResult := result.(*file.ReadFileResult)
		fmt.Printf("Lines 5-10:\n%s\n", readResult.Content)
		fmt.Printf("Lines read: %v\n", readResult.Lines)
	}

	// Example 5: Using with an agent
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		fmt.Println("\n=== Example 5: Agent with File Tools ===")

		// Create provider and agent
		p := provider.NewOpenAIProvider(apiKey, "gpt-4o-mini")
		agent := workflow.NewAgent(p).
			SetSystemPrompt("You are a helpful assistant that can read and write files.").
			AddTool(readTool).
			AddTool(writeTool)

		// Use the agent
		todoFile := filepath.Join(demoDir, "todo.md")
		result, err := agent.Run(ctx, fmt.Sprintf(
			"Create a todo list file at %s with 3 tasks, then read it back and tell me what's in it.",
			todoFile,
		))
		if err != nil {
			log.Printf("Error running agent: %v", err)
		} else {
			fmt.Printf("Agent response: %v\n", result)
		}
	} else {
		fmt.Println("\n=== Agent Example Skipped (no API key) ===")
	}

	// Show tool examples
	fmt.Println("\n=== Tool Usage Examples ===")
	for _, entry := range fileTools {
		fmt.Printf("\n%s examples:\n", entry.Metadata.Name)
		for _, example := range entry.Metadata.Examples {
			fmt.Printf("  %s: %s\n", example.Name, example.Description)
			fmt.Printf("    %s\n", example.Code)
		}
	}
}
