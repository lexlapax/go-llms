// ABOUTME: Example demonstrating the built-in file tools
// ABOUTME: Shows both direct tool usage and LLM agent integration with minimal prompting

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

// createToolContext creates a ToolContext for direct tool execution
func createToolContext() *domain.ToolContext {
	ctx := context.Background()
	agentInfo := domain.AgentInfo{
		ID:          "file-test-agent",
		Name:        "File Tools Test",
		Description: "Testing file tools",
		Type:        domain.AgentTypeCustom,
	}

	state := domain.NewState()
	// Enable useful options for file operations
	state.Set("file_read_include_meta", true)
	state.Set("file_write_atomic", true)
	state.Set("file_list_recursive", false)

	stateReader := domain.NewStateReader(state)

	return &domain.ToolContext{
		Context: ctx,
		State:   stateReader,
		Agent:   agentInfo,
		RunID:   "file-test-run-" + fmt.Sprintf("%d", time.Now().Unix()),
	}
}

// printResult prints a tool execution result in a formatted way
func printResult(operation string, result interface{}, err error) {
	if err != nil {
		fmt.Printf("❌ %s failed: %v\n", operation, err)
		return
	}
	fmt.Printf("✓ %s successful\n", operation)
	fmt.Printf("  Result: %+v\n", result)
}

// createMockProvider creates a mock provider that simulates file tool usage
func createMockProvider() ldomain.Provider {
	mockProvider := provider.NewMockProvider()
	hasToolResult := false

	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Extract the last user message
		var lastUserMsg string
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				for _, part := range messages[i].Content {
					if part.Type == ldomain.ContentTypeText {
						lastUserMsg = part.Text
						break
					}
				}
				if lastUserMsg != "" {
					break
				}
			}
		}

		// Check if this is a tool result response
		if strings.Contains(lastUserMsg, "Tool results:") && strings.Contains(lastUserMsg, "Result:") {
			hasToolResult = true
			// Extract result from tool result message
			if strings.Contains(messages[len(messages)-3].Content[0].Text, "read") && strings.Contains(messages[len(messages)-3].Content[0].Text, "config") {
				return ldomain.Response{
					Content: "The configuration file has been successfully read. It contains JSON settings with debug mode enabled and log level set to 'info'.",
				}, nil
			} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "write") || strings.Contains(messages[len(messages)-3].Content[0].Text, "create") {
				return ldomain.Response{
					Content: "I've successfully created the todo list file with 5 important tasks.",
				}, nil
			} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "list") && strings.Contains(messages[len(messages)-3].Content[0].Text, "JSON") {
				return ldomain.Response{
					Content: "I found 2 JSON files in the directory: config.json and settings.json. Both files are configuration files for the application.",
				}, nil
			}

			return ldomain.Response{
				Content: "The file operation has been completed successfully.",
			}, nil
		}

		// Initial tool call based on the prompt
		if lastUserMsg != "" && !hasToolResult {
			if contains(lastUserMsg, "read") && contains(lastUserMsg, "config") {
				return ldomain.Response{
					Content: `{"tool": "file_read", "params": {"path": "/tmp/demo/config.json", "include_meta": true}}`,
				}, nil
			} else if (contains(lastUserMsg, "create") || contains(lastUserMsg, "write")) && contains(lastUserMsg, "todo") {
				return ldomain.Response{
					Content: `{"tool": "file_write", "params": {"path": "/tmp/demo/todo.md", "content": "# TODO List\n\n1. Complete file tools migration\n2. Test all examples\n3. Update documentation\n4. Run benchmarks\n5. Create release notes\n", "atomic": true}}`,
				}, nil
			} else if contains(lastUserMsg, "list") && contains(lastUserMsg, "JSON") {
				return ldomain.Response{
					Content: `{"tool": "file_list", "params": {"path": "/tmp/demo", "pattern": "*.json"}}`,
				}, nil
			} else if contains(lastUserMsg, "search") && contains(lastUserMsg, "error") {
				return ldomain.Response{
					Content: `{"tool": "file_search", "params": {"path": "/tmp/demo", "pattern": "error", "file_pattern": "*.log"}}`,
				}, nil
			}
		}

		return ldomain.Response{
			Content: "I can help you with file operations like reading, writing, listing, searching, moving, and deleting files.",
		}, nil
	})

	return mockProvider
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func main() {
	var llmMode bool
	flag.BoolVar(&llmMode, "llm", false, "Run LLM agent examples instead of direct tool usage")
	flag.Parse()

	fmt.Println("=== Built-in File Tools Example ===")
	fmt.Println()

	if llmMode {
		runLLMExample()
	} else {
		runDirectExample()
	}
}

func runDirectExample() {
	fmt.Println("=== Direct Tool Usage Examples ===")
	fmt.Println()

	// Get all file tools
	readTool, ok := tools.GetTool("file_read")
	if !ok {
		log.Fatal("Failed to get file_read tool")
	}

	writeTool, ok := tools.GetTool("file_write")
	if !ok {
		log.Fatal("Failed to get file_write tool")
	}

	listTool, ok := tools.GetTool("file_list")
	if !ok {
		log.Fatal("Failed to get file_list tool")
	}

	deleteTool, ok := tools.GetTool("file_delete")
	if !ok {
		log.Fatal("Failed to get file_delete tool")
	}

	moveTool, ok := tools.GetTool("file_move")
	if !ok {
		log.Fatal("Failed to get file_move tool")
	}

	searchTool, ok := tools.GetTool("file_search")
	if !ok {
		log.Fatal("Failed to get file_search tool")
	}

	// Create demo directory
	demoDir, err := os.MkdirTemp("", "file_tools_demo_*")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(demoDir); err != nil {
			log.Printf("Failed to remove demo directory: %v", err)
		}
	}()
	fmt.Printf("Using demo directory: %s\n\n", demoDir)

	// Create tool context
	toolCtx := createToolContext()

	// Example 1: Write a configuration file
	fmt.Println("--- Example 1: Write Configuration File ---")

	configFile := filepath.Join(demoDir, "config.json")
	configContent := `{
  "version": "1.0",
  "settings": {
    "debug": true,
    "log_level": "info",
    "port": 8080
  }
}`

	result, err := writeTool.Execute(toolCtx, map[string]interface{}{
		"path":    configFile,
		"content": configContent,
		"atomic":  true,
	})
	printResult("Write config.json", result, err)

	// Example 2: Read file with metadata
	fmt.Println("\n--- Example 2: Read with Metadata ---")

	result, err = readTool.Execute(toolCtx, map[string]interface{}{
		"path":         configFile,
		"include_meta": true,
	})
	printResult("Read config with metadata", result, err)

	// Example 3: Create and append to log file
	fmt.Println("\n--- Example 3: Append to Log File ---")

	logFile := filepath.Join(demoDir, "app.log")

	// Initial log entry
	result, err = writeTool.Execute(toolCtx, map[string]interface{}{
		"path":    logFile,
		"content": fmt.Sprintf("[%s] Application started\n", time.Now().Format("2006-01-02 15:04:05")),
	})
	printResult("Create log file", result, err)

	// Append entries
	result, err = writeTool.Execute(toolCtx, map[string]interface{}{
		"path":    logFile,
		"content": fmt.Sprintf("[%s] Configuration loaded\n", time.Now().Format("2006-01-02 15:04:05")),
		"append":  true,
	})
	printResult("Append to log", result, err)

	// Example 4: List directory contents
	fmt.Println("\n--- Example 4: List Directory ---")

	// Create a few more files for listing
	_, _ = writeTool.Execute(toolCtx, map[string]interface{}{
		"path":    filepath.Join(demoDir, "settings.json"),
		"content": `{"theme": "dark"}`,
	})

	_, _ = writeTool.Execute(toolCtx, map[string]interface{}{
		"path":    filepath.Join(demoDir, "README.md"),
		"content": "# File Tools Demo\n\nThis is a demonstration.",
	})

	result, err = listTool.Execute(toolCtx, map[string]interface{}{
		"path":      demoDir,
		"pattern":   "*.json",
		"recursive": false,
	})
	printResult("List JSON files", result, err)

	// Example 5: Search file content
	fmt.Println("\n--- Example 5: Search Content ---")

	// Add some content to search
	_, _ = writeTool.Execute(toolCtx, map[string]interface{}{
		"path":    filepath.Join(demoDir, "errors.log"),
		"content": "Error: Connection timeout\nWarning: Low memory\nError: File not found\n",
	})

	result, err = searchTool.Execute(toolCtx, map[string]interface{}{
		"path":         demoDir,
		"pattern":      "Error",
		"file_pattern": "*.log",
	})
	printResult("Search for 'Error' in logs", result, err)

	// Example 6: Move/Rename file
	fmt.Println("\n--- Example 6: Move/Rename File ---")

	oldPath := filepath.Join(demoDir, "settings.json")
	newPath := filepath.Join(demoDir, "app-settings.json")

	result, err = moveTool.Execute(toolCtx, map[string]interface{}{
		"source":      oldPath,
		"destination": newPath,
	})
	printResult("Rename settings.json", result, err)

	// Example 7: Atomic write with backup
	fmt.Println("\n--- Example 7: Atomic Write with Backup ---")

	// Update config with backup
	updatedConfig := `{
  "version": "1.1",
  "settings": {
    "debug": false,
    "log_level": "warn",
    "port": 9090,
    "new_feature": true
  }
}`

	result, err = writeTool.Execute(toolCtx, map[string]interface{}{
		"path":    configFile,
		"content": updatedConfig,
		"atomic":  true,
		"backup":  true,
	})
	printResult("Update config with backup", result, err)

	// Example 8: Read specific lines from large file
	fmt.Println("\n--- Example 8: Read Line Range ---")

	// Create a file with many lines
	largeFile := filepath.Join(demoDir, "large.txt")
	var content strings.Builder
	for i := 1; i <= 50; i++ {
		content.WriteString(fmt.Sprintf("Line %d: This is line number %d of the file\n", i, i))
	}

	_, _ = writeTool.Execute(toolCtx, map[string]interface{}{
		"path":    largeFile,
		"content": content.String(),
	})

	// Read only lines 10-15
	result, err = readTool.Execute(toolCtx, map[string]interface{}{
		"path":       largeFile,
		"line_start": 10,
		"line_end":   15,
	})
	printResult("Read lines 10-15", result, err)

	// Example 9: Safe delete
	fmt.Println("\n--- Example 9: Safe Delete ---")

	// Create a temporary file to delete
	tempFile := filepath.Join(demoDir, "temp.txt")
	_, _ = writeTool.Execute(toolCtx, map[string]interface{}{
		"path":    tempFile,
		"content": "Temporary file",
	})

	result, err = deleteTool.Execute(toolCtx, map[string]interface{}{
		"path": tempFile,
	})
	printResult("Delete temp.txt", result, err)

	// Example 10: Error handling
	fmt.Println("\n--- Example 10: Error Handling ---")

	// Try to read non-existent file
	result, err = readTool.Execute(toolCtx, map[string]interface{}{
		"path": filepath.Join(demoDir, "nonexistent.txt"),
	})
	printResult("Read non-existent file", result, err)

	// Try to write to invalid path
	invalidPath := "/root/cannot-write-here.txt"
	if runtime.GOOS == "windows" {
		invalidPath = "C:\\Windows\\System32\\cannot-write.txt"
	}
	result, err = writeTool.Execute(toolCtx, map[string]interface{}{
		"path":    invalidPath,
		"content": "test",
	})
	printResult("Write to protected location", result, err)

	// Display available operations from tool metadata
	fmt.Println("\n--- Available File Tools ---")
	fileTools := tools.Tools.ListByCategory("file")
	for _, entry := range fileTools {
		fmt.Printf("\n%s (v%s): %s\n", entry.Metadata.Name, entry.Metadata.Version, entry.Metadata.Description)
		fmt.Printf("  Tags: %v\n", entry.Metadata.Tags)
		if entry.Component.EstimatedLatency() != "" {
			fmt.Printf("  Latency: %s\n", entry.Component.EstimatedLatency())
		}
		if entry.Component.IsDestructive() {
			fmt.Printf("  ⚠️  Destructive operation\n")
		}
		if entry.Component.RequiresConfirmation() {
			fmt.Printf("  ⚠️  Requires confirmation\n")
		}
	}
}

func runLLMExample() {
	fmt.Println("=== LLM Agent with File Tools ===")
	fmt.Println()

	// Try to get a real provider, fall back to mock
	providerString := "anthropic/claude-3-7-sonnet-latest"
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		providerString = "anthropic/claude-3-7-sonnet-latest"
	} else if os.Getenv("OPENAI_API_KEY") != "" {
		providerString = "openai/gpt-4o"
	} else if os.Getenv("GEMINI_API_KEY") != "" {
		providerString = "gemini/gemini-2.0-flash"
	}

	provider, err := llmutil.NewProviderFromString(providerString)
	if err != nil {
		// Fall back to mock provider
		fmt.Println("Note: No LLM API keys found. Using mock provider for demonstration.")
		fmt.Println("The mock will simulate file tool usage.")
		fmt.Println("Tip: Set DEBUG=1 to see detailed logging of agent execution.")
		fmt.Println()
		provider = createMockProvider()
	}

	// Parse provider info
	providerName, modelName, _ := llmutil.ParseProviderModelString(providerString)
	fmt.Printf("Provider: %s\n", providerName)
	if modelName != "" {
		fmt.Printf("Model: %s\n\n", modelName)
	}

	// Get file tools
	readTool, _ := tools.GetTool("file_read")
	writeTool, _ := tools.GetTool("file_write")
	listTool, _ := tools.GetTool("file_list")
	searchTool, _ := tools.GetTool("file_search")
	moveTool, _ := tools.GetTool("file_move")
	deleteTool, _ := tools.GetTool("file_delete")

	// Create LLM agent
	deps := core.LLMDeps{
		Provider: provider,
	}
	agent := core.NewLLMAgent("file-assistant", "File Management Assistant", deps)

	// Add logging if DEBUG is enabled
	if os.Getenv("DEBUG") == "1" {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
		loggingHook := core.NewLoggingHook(logger, core.LogLevelDebug)
		agent.WithHook(loggingHook)
		log.Println("Debug logging enabled")
	}

	// Add all file tools
	agent.AddTool(readTool)
	agent.AddTool(writeTool)
	agent.AddTool(listTool)
	agent.AddTool(searchTool)
	agent.AddTool(moveTool)
	agent.AddTool(deleteTool)

	// Set minimal system prompt - let the tools guide the LLM
	agent.SetSystemPrompt(`You are a helpful file management assistant. You have access to tools for:
- Reading files with metadata and line ranges
- Writing files with atomic operations and backups
- Listing directory contents with pattern matching
- Searching for files and content within files
- Moving and renaming files safely
- Deleting files with confirmation

When asked about files, use the appropriate tools to perform the requested operations. The tools will guide you on their proper usage.`)

	// Create a demo directory
	demoDir, err := os.MkdirTemp("", "llm_file_demo_*")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(demoDir); err != nil {
			log.Printf("Failed to remove demo directory: %v", err)
		}
	}()

	// Pre-populate with some files
	configPath := filepath.Join(demoDir, "config.json")
	_ = os.WriteFile(configPath, []byte(`{"debug": true, "log_level": "info"}`), 0644)

	_ = os.WriteFile(filepath.Join(demoDir, "app.log"), []byte("[2024-01-01] Started\n[2024-01-01] Error: Connection failed\n"), 0644)
	_ = os.WriteFile(filepath.Join(demoDir, "settings.json"), []byte(`{"theme": "dark"}`), 0644)

	// Example queries
	examples := []struct {
		title  string
		prompt string
	}{
		{
			title:  "Read Configuration",
			prompt: fmt.Sprintf("Read the configuration file at %s and tell me what settings are enabled", configPath),
		},
		{
			title:  "Create TODO List",
			prompt: fmt.Sprintf("Create a todo list file at %s with 5 important tasks", filepath.Join(demoDir, "todo.md")),
		},
		{
			title:  "List JSON Files",
			prompt: fmt.Sprintf("List all JSON files in %s and tell me what each one is for", demoDir),
		},
		{
			title:  "Search for Errors",
			prompt: fmt.Sprintf("Search for any errors in log files in %s", demoDir),
		},
		{
			title:  "File Operations",
			prompt: fmt.Sprintf("In %s: rename settings.json to app-settings.json, then verify it was renamed by listing JSON files", demoDir),
		},
		{
			title:  "Multi-Step Task",
			prompt: fmt.Sprintf("Read the config at %s, create a backup of it, then update the debug setting to false", configPath),
		},
	}

	// Run examples
	ctx := context.Background()
	for i, example := range examples {
		fmt.Printf("--- Example %d: %s ---\n", i+1, example.title)
		fmt.Printf("Query: %s\n", example.prompt)

		state := domain.NewState()
		state.Set("user_input", example.prompt)

		result, err := agent.Run(ctx, state)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		// Get the response
		if response, exists := result.Get("response"); exists {
			fmt.Printf("Response: %v\n\n", response)
		} else if messages := result.Messages(); len(messages) > 0 {
			// Check messages for response
			lastMsg := messages[len(messages)-1]
			fmt.Printf("Response: %s\n\n", lastMsg.Content)
		}
	}

	// Display tool information if requested
	if len(os.Args) > 2 && os.Args[2] == "info" {
		fmt.Println("=== File Tool Information ===")
		fileTools := tools.Tools.ListByCategory("file")
		for _, entry := range fileTools {
			tool := entry.Component
			fmt.Printf("\n%s:\n", tool.Name())
			fmt.Printf("  Description: %s\n", tool.Description())
			fmt.Printf("  Version: %s\n", tool.Version())
			fmt.Printf("  Category: %s\n", tool.Category())
			fmt.Printf("  Tags: %v\n", tool.Tags())
			fmt.Printf("  Deterministic: %v\n", tool.IsDeterministic())
			fmt.Printf("  Destructive: %v\n", tool.IsDestructive())
			fmt.Printf("  Requires Confirmation: %v\n", tool.RequiresConfirmation())
			fmt.Printf("  Estimated Latency: %s\n", tool.EstimatedLatency())

			if len(tool.Examples()) > 0 {
				fmt.Printf("  Examples: %d available\n", len(tool.Examples()))
			}
		}
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("The LLM agent successfully used file tools to:")
	fmt.Println("• Read files and extract configuration information")
	fmt.Println("• Create new files with structured content")
	fmt.Println("• List and analyze directory contents")
	fmt.Println("• Search for patterns in files")
	fmt.Println("• Perform multi-step file operations")
	fmt.Println("\nThe tools' built-in documentation guided the LLM on proper usage.")
}
