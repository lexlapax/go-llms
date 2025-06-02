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
	fmt.Printf("Total file tools: %d\n", len(fileTools))
	for _, entry := range fileTools {
		fmt.Printf("  - %s: %s\n", entry.Metadata.Name, entry.Metadata.Description)
		fmt.Printf("    Version: %s\n", entry.Metadata.Version)
		fmt.Printf("    Tags: %v\n", entry.Metadata.Tags)
		fmt.Println()
	}

	// Get all the tools
	readTool := tools.MustGetTool("file_read")
	writeTool := tools.MustGetTool("file_write")
	listTool := tools.MustGetTool("file_list")
	deleteTool := tools.MustGetTool("file_delete")
	moveTool := tools.MustGetTool("file_move")
	searchTool := tools.MustGetTool("file_search")

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
		if writeResult, ok := result.(*file.WriteFileResult); ok {
			fmt.Printf("Initial write successful: %d bytes written\n", writeResult.BytesWritten)
			fmt.Printf("File path: %s\n", writeResult.AbsolutePath)
		} else {
			fmt.Printf("Unexpected result type for write operation\n")
		}
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
		if writeResult, ok := result.(*file.WriteFileResult); ok {
			fmt.Printf("Update successful with atomic write\n")
			fmt.Printf("  Bytes written: %d\n", writeResult.BytesWritten)
			fmt.Printf("  Backup created at: %s\n", writeResult.BackupPath)
			fmt.Printf("  File existed before: %v\n", writeResult.FileExisted)
		} else {
			fmt.Printf("Unexpected result type for atomic write operation\n")
		}
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
		if readResult, ok := result.(*file.ReadFileResult); ok {
			fmt.Printf("File content:\n%s\n", readResult.Content)
			fmt.Printf("Characters read: %d\n", len(readResult.Content))
			if readResult.Metadata != nil {
				fmt.Printf("File metadata:\n")
				fmt.Printf("  Size: %d bytes\n", readResult.Metadata.Size)
				fmt.Printf("  Extension: %s\n", readResult.Metadata.Extension)
				fmt.Printf("  Modified: %v\n", readResult.Metadata.ModTime)
				fmt.Printf("  Is directory: %v\n", readResult.Metadata.IsDir)
				fmt.Printf("  Permissions: %v\n", readResult.Metadata.Mode)
			}
		} else {
			fmt.Printf("Unexpected result type for read operation\n")
		}
	}

	// Example 3: Append to log file
	fmt.Println("\n=== Example 3: Append Mode ===")
	logFile := filepath.Join(demoDir, "app.log")

	// Write initial log entry
	_, err = writeTool.Execute(ctx, map[string]interface{}{
		"path":    logFile,
		"content": "2024-01-31 10:00:00 - Application started\n",
	})
	if err != nil {
		log.Printf("Error writing initial log entry: %v", err)
	}

	// Append more log entries
	_, err = writeTool.Execute(ctx, map[string]interface{}{
		"path":    logFile,
		"content": "2024-01-31 10:00:01 - Configuration loaded\n",
		"append":  true,
	})
	if err != nil {
		log.Printf("Error appending to log: %v", err)
	}

	_, err = writeTool.Execute(ctx, map[string]interface{}{
		"path":    logFile,
		"content": "2024-01-31 10:00:02 - Server listening on port 8080\n",
		"append":  true,
	})
	if err != nil {
		log.Printf("Error appending to log: %v", err)
	}

	// Read the full log
	result, err = readTool.Execute(ctx, map[string]interface{}{
		"path": logFile,
	})
	if err != nil {
		log.Printf("Error reading log: %v", err)
	} else {
		if readResult, ok := result.(*file.ReadFileResult); ok {
			fmt.Printf("Full log:\n%s", readResult.Content)
			fmt.Printf("Log file size: %d characters\n", len(readResult.Content))
		} else {
			fmt.Printf("Unexpected result type for log read operation\n")
		}
	}

	// Example 4: Read specific lines from large file
	fmt.Println("\n=== Example 4: Line Range Reading ===")
	largeFile := filepath.Join(demoDir, "large.txt")

	// Create a file with many lines
	var content string
	for i := 1; i <= 20; i++ {
		content += fmt.Sprintf("Line %d: This is some content on line %d\n", i, i)
	}
	_, err = writeTool.Execute(ctx, map[string]interface{}{
		"path":    largeFile,
		"content": content,
	})
	if err != nil {
		log.Printf("Error creating large file: %v", err)
	}

	// Read only lines 5-10
	result, err = readTool.Execute(ctx, map[string]interface{}{
		"path":       largeFile,
		"line_start": 5,
		"line_end":   10,
	})
	if err != nil {
		log.Printf("Error reading lines: %v", err)
	} else {
		if readResult, ok := result.(*file.ReadFileResult); ok {
			fmt.Printf("Lines 5-10:\n%s\n", readResult.Content)
			fmt.Printf("Lines read: %d\n", readResult.Lines)
			fmt.Printf("Character count: %d\n", len(readResult.Content))
		} else {
			fmt.Printf("Unexpected result type for line range read operation\n")
		}
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

	// Example 6: File List operation
	fmt.Println("\n=== Example 6: File List ===")
	result, err = listTool.Execute(ctx, map[string]interface{}{
		"path":      demoDir,
		"recursive": false,
	})
	if err != nil {
		log.Printf("Error listing files: %v", err)
	} else {
		if listResult, ok := result.(*file.FileListResult); ok {
			fmt.Printf("Found %d files:\n", len(listResult.Files))
			var totalSize int64
			for _, fileInfo := range listResult.Files {
				fmt.Printf("  - %s (%d bytes, modified: %v)\n",
					fileInfo.Name, fileInfo.Size, fileInfo.ModifiedTime.Format("2006-01-02 15:04:05"))
				totalSize += fileInfo.Size
			}
			fmt.Printf("Total size: %d bytes\n", totalSize)
		} else {
			fmt.Printf("Unexpected result type for list operation\n")
		}
	}

	// Example 7: File Search operation
	fmt.Println("\n=== Example 7: File Search ===")
	result, err = searchTool.Execute(ctx, map[string]interface{}{
		"path":      demoDir,
		"pattern":   "*.json",
		"recursive": false,
	})
	if err != nil {
		log.Printf("Error searching files: %v", err)
	} else {
		if searchResult, ok := result.(*file.FileSearchResult); ok {
			fmt.Printf("Found %d JSON files:\n", len(searchResult.Matches))
			for _, match := range searchResult.Matches {
				fmt.Printf("  - %s\n", match.File)
			}
			fmt.Printf("Total matches: %d\n", searchResult.TotalMatches)
			fmt.Printf("Files searched: %d\n", searchResult.FilesSearched)
		} else {
			fmt.Printf("Unexpected result type for search operation\n")
		}
	}

	// Example 8: File Move operation
	fmt.Println("\n=== Example 8: File Move ===")
	// Create a test file to move
	moveTestFile := filepath.Join(demoDir, "move_test.txt")
	_, err = writeTool.Execute(ctx, map[string]interface{}{
		"path":    moveTestFile,
		"content": "This file will be moved",
	})
	if err != nil {
		log.Printf("Error creating test file for move: %v", err)
	}

	// Move the file
	movedFile := filepath.Join(demoDir, "moved_file.txt")
	result, err = moveTool.Execute(ctx, map[string]interface{}{
		"source":      moveTestFile,
		"destination": movedFile,
	})
	if err != nil {
		log.Printf("Error moving file: %v", err)
	} else {
		if moveResult, ok := result.(*file.FileMoveResult); ok {
			fmt.Printf("File moved successfully\n")
			fmt.Printf("  From: %s\n", moveResult.Source)
			fmt.Printf("  To: %s\n", moveResult.Destination)
			fmt.Printf("  Was rename: %v\n", moveResult.WasRename)
		} else {
			fmt.Printf("Unexpected result type for move operation\n")
		}
	}

	// Example 9: File Delete operation
	fmt.Println("\n=== Example 9: File Delete ===")
	// Create a test file to delete
	deleteTestFile := filepath.Join(demoDir, "delete_test.txt")
	_, err = writeTool.Execute(ctx, map[string]interface{}{
		"path":    deleteTestFile,
		"content": "This file will be deleted",
	})
	if err != nil {
		log.Printf("Error creating test file for delete: %v", err)
	}

	// Delete the file
	result, err = deleteTool.Execute(ctx, map[string]interface{}{
		"path": deleteTestFile,
	})
	if err != nil {
		log.Printf("Error deleting file: %v", err)
	} else {
		if deleteResult, ok := result.(*file.FileDeleteResult); ok {
			fmt.Printf("File deleted successfully\n")
			fmt.Printf("  Path: %s\n", deleteResult.Path)
			fmt.Printf("  Was deleted: %v\n", deleteResult.Deleted)
			fmt.Printf("  Was directory: %v\n", deleteResult.WasDirectory)
		} else {
			fmt.Printf("Unexpected result type for delete operation\n")
		}
	}

	// Example 10: Advanced read with content search
	fmt.Println("\n=== Example 10: Content Search in Files ===")
	result, err = searchTool.Execute(ctx, map[string]interface{}{
		"path":         demoDir,
		"pattern":      "Line",
		"file_pattern": "*.txt",
		"recursive":    false,
	})
	if err != nil {
		log.Printf("Error searching file content: %v", err)
	} else {
		if searchResult, ok := result.(*file.FileSearchResult); ok {
			fmt.Printf("Found %d content matches for 'Line':\n", len(searchResult.Matches))
			for i, match := range searchResult.Matches {
				if i < 5 { // Show first 5 matches
					fmt.Printf("  - %s line %d: %s\n", match.File, match.LineNumber, match.Line)
				}
			}
			if len(searchResult.Matches) > 5 {
				fmt.Printf("  ... and %d more matches\n", len(searchResult.Matches)-5)
			}
			fmt.Printf("Total matches: %d\n", searchResult.TotalMatches)
			fmt.Printf("Files searched: %d\n", searchResult.FilesSearched)
		} else {
			fmt.Printf("Unexpected result type for content search operation\n")
		}
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

	fmt.Println("\n=== Summary ===")
	fmt.Println("This example demonstrated all 6 file tools:")
	fmt.Println("• file_read: Read with metadata, line ranges, and content validation")
	fmt.Println("• file_write: Atomic writes, append mode, and backup creation")
	fmt.Println("• file_list: Directory listing with file metadata")
	fmt.Println("• file_search: Pattern matching and content search")
	fmt.Println("• file_move: File relocation with validation")
	fmt.Println("• file_delete: Safe file removal with existence checks")
	fmt.Println("\nAll operations include comprehensive error handling and result validation.")
}
