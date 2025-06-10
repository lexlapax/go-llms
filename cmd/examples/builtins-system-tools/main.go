// ABOUTME: Example demonstrating the use of built-in system tools
// ABOUTME: Shows both direct tool usage and LLM agent integration with minimal prompting

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
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
		ID:          "system-test-agent",
		Name:        "System Tools Test",
		Description: "Testing system tools",
		Type:        domain.AgentTypeCustom,
	}

	state := domain.NewState()
	// Enable all system info options for demos
	state.Set("system_info_include_memory", true)
	state.Set("system_info_include_runtime", true)
	state.Set("system_info_include_environment", true)

	stateReader := domain.NewStateReader(state)

	return &domain.ToolContext{
		Context: ctx,
		State:   stateReader,
		Agent:   agentInfo,
		RunID:   "system-test-run-" + fmt.Sprintf("%d", time.Now().Unix()),
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

// createMockProvider creates a mock provider that simulates system tool usage
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
			if strings.Contains(messages[len(messages)-3].Content[0].Text, "operating system") || strings.Contains(messages[len(messages)-3].Content[0].Text, "architecture") {
				return ldomain.Response{
					Content: "This system is running on Linux with x86_64 architecture. It has 16 GB of total memory with 8 GB available.",
				}, nil
			} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "memory") {
				return ldomain.Response{
					Content: "The current memory usage shows 8 GB used out of 16 GB total, with 8 GB available. The system is operating efficiently.",
				}, nil
			} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "Go-related") || strings.Contains(messages[len(messages)-3].Content[0].Text, "GO") {
				return ldomain.Response{
					Content: "I found several Go-related environment variables: GOPATH is set to /home/user/go, GOROOT is /usr/local/go, and GO111MODULE is set to 'on'.",
				}, nil
			} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "top") && strings.Contains(messages[len(messages)-3].Content[0].Text, "memory") {
				return ldomain.Response{
					Content: "The top 3 processes by memory usage are: 1) Chrome (2.5 GB), 2) VS Code (1.8 GB), and 3) Docker Desktop (1.2 GB).",
				}, nil
			} else if strings.Contains(messages[len(messages)-3].Content[0].Text, "directory") && strings.Contains(messages[len(messages)-3].Content[0].Text, "list") {
				return ldomain.Response{
					Content: "The current directory is /home/user/projects and contains: README.md, go.mod, go.sum, main.go, and several subdirectories.",
				}, nil
			}

			return ldomain.Response{
				Content: "The system operation has been completed successfully.",
			}, nil
		}

		// Initial tool call based on the prompt
		if lastUserMsg != "" && !hasToolResult {
			if contains(lastUserMsg, "operating system") || contains(lastUserMsg, "architecture") {
				return ldomain.Response{
					Content: `{"tool": "get_system_info", "params": {"include_memory": true, "include_runtime": true}}`,
				}, nil
			} else if contains(lastUserMsg, "memory") && contains(lastUserMsg, "usage") {
				return ldomain.Response{
					Content: `{"tool": "get_system_info", "params": {"include_memory": true}}`,
				}, nil
			} else if contains(lastUserMsg, "Go-related") || contains(lastUserMsg, "GO") {
				return ldomain.Response{
					Content: `{"tool": "get_environment_variable", "params": {"pattern": "*GO*"}}`,
				}, nil
			} else if contains(lastUserMsg, "top") && contains(lastUserMsg, "memory") {
				return ldomain.Response{
					Content: `{"tool": "process_list", "params": {"sort_by": "memory", "limit": 3}}`,
				}, nil
			} else if contains(lastUserMsg, "current directory") && contains(lastUserMsg, "list") {
				return ldomain.Response{
					Content: `{"tool": "execute_command", "params": {"command": "pwd && ls -la", "timeout": 5}}`,
				}, nil
			} else if contains(lastUserMsg, "system") && contains(lastUserMsg, "CPUs") {
				return ldomain.Response{
					Content: `{"tool": "get_system_info", "params": {"include_memory": true, "include_runtime": true, "include_environment": true}}`,
				}, nil
			}
		}

		return ldomain.Response{
			Content: "I can help you with system information, environment variables, process management, and command execution.",
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

	fmt.Println("=== Built-in System Tools Example ===")
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

	// Get all system tools
	execTool, ok := tools.GetTool("execute_command")
	if !ok {
		log.Fatal("Failed to get execute_command tool")
	}

	envTool, ok := tools.GetTool("get_environment_variable")
	if !ok {
		log.Fatal("Failed to get environment variable tool")
	}

	sysInfoTool, ok := tools.GetTool("get_system_info")
	if !ok {
		log.Fatal("Failed to get system info tool")
	}

	procTool, ok := tools.GetTool("process_list")
	if !ok {
		log.Fatal("Failed to get process list tool")
	}

	// Create tool context
	toolCtx := createToolContext()

	// Example 1: System Information
	fmt.Println("--- System Information ---")

	// Basic system info
	result, err := sysInfoTool.Execute(toolCtx, map[string]interface{}{})
	printResult("Basic system info", result, err)

	// Full system info with all details
	result, err = sysInfoTool.Execute(toolCtx, map[string]interface{}{
		"include_memory":      true,
		"include_runtime":     true,
		"include_environment": true,
	})
	printResult("Full system info", result, err)

	// Example 2: Environment Variables
	fmt.Println("\n--- Environment Variables ---")

	// Get PATH variable
	result, err = envTool.Execute(toolCtx, map[string]interface{}{
		"name": "PATH",
	})
	printResult("Get PATH", result, err)

	// Search for Go-related variables
	result, err = envTool.Execute(toolCtx, map[string]interface{}{
		"pattern": "*GO*",
	})
	printResult("Search GO* variables", result, err)

	// Get all variables (names only)
	result, err = envTool.Execute(toolCtx, map[string]interface{}{
		"list_all":  true,
		"no_values": true,
	})
	printResult("List all variable names", result, err)

	// Example 3: Process Management
	fmt.Println("\n--- Process Management ---")

	// Get current process
	result, err = procTool.Execute(toolCtx, map[string]interface{}{
		"current_only": true,
	})
	printResult("Current process", result, err)

	// Top 5 by memory
	result, err = procTool.Execute(toolCtx, map[string]interface{}{
		"sort_by": "memory",
		"limit":   5,
	})
	printResult("Top 5 by memory", result, err)

	// Search for Go processes
	result, err = procTool.Execute(toolCtx, map[string]interface{}{
		"filter":       "go",
		"include_self": true,
	})
	printResult("Go processes", result, err)

	// Example 4: Command Execution
	fmt.Println("\n--- Command Execution ---")

	// Simple command
	result, err = execTool.Execute(toolCtx, map[string]interface{}{
		"command": "echo 'Hello from system tools!'",
		"timeout": 5,
	})
	printResult("Echo command", result, err)

	// Platform-specific commands
	if runtime.GOOS == "windows" {
		result, err = execTool.Execute(toolCtx, map[string]interface{}{
			"command":   "dir",
			"safe_mode": true,
			"timeout":   5,
		})
		printResult("Windows dir", result, err)
	} else {
		result, err = execTool.Execute(toolCtx, map[string]interface{}{
			"command":   "ls -la",
			"safe_mode": true,
			"timeout":   5,
		})
		printResult("Unix ls", result, err)
	}

	// Command with environment variables
	result, err = execTool.Execute(toolCtx, map[string]interface{}{
		"command": "echo \"User: $DEMO_USER, Time: $(date)\"",
		"shell":   "bash",
		"env": map[string]string{
			"DEMO_USER": "SystemToolsDemo",
		},
		"timeout": 5,
	})
	printResult("Command with env", result, err)

	// Example 5: Error Handling
	fmt.Println("\n--- Error Handling ---")

	// Command timeout
	result, err = execTool.Execute(toolCtx, map[string]interface{}{
		"command": "sleep 10",
		"timeout": 1, // 1 second timeout
	})
	printResult("Timeout test", result, err)

	// Non-existent command
	result, err = execTool.Execute(toolCtx, map[string]interface{}{
		"command": "nonexistentcommand123",
		"timeout": 5,
	})
	printResult("Non-existent command", result, err)

	// Invalid environment variable name
	result, err = envTool.Execute(toolCtx, map[string]interface{}{
		"name": "", // Empty name
	})
	printResult("Invalid env var", result, err)

	// Example 6: Advanced Features
	fmt.Println("\n--- Advanced Features ---")

	// Command with working directory
	result, err = execTool.Execute(toolCtx, map[string]interface{}{
		"command":     "pwd",
		"working_dir": "/tmp",
		"timeout":     5,
	})
	printResult("Command in /tmp", result, err)

	// Capture both stdout and stderr
	result, err = execTool.Execute(toolCtx, map[string]interface{}{
		"command":        "echo 'stdout' && echo 'stderr' >&2",
		"shell":          "bash",
		"capture_stderr": true,
		"timeout":        5,
	})
	printResult("Capture stdout/stderr", result, err)

	// Process list with all fields
	result, err = procTool.Execute(toolCtx, map[string]interface{}{
		"include_command": true,
		"include_user":    true,
		"include_state":   true,
		"limit":           3,
	})
	printResult("Detailed process info", result, err)

	// Display available operations from tool metadata
	fmt.Println("\n--- Available System Tools ---")
	systemTools := tools.Tools.ListByCategory("system")
	for _, entry := range systemTools {
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
	fmt.Println("=== LLM Agent with System Tools ===")
	fmt.Println()

	// Try to get a real provider, fall back to mock
	providerString := "anthropic/claude-3-5-sonnet"
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		providerString = "anthropic/claude-3-5-sonnet"
	} else if os.Getenv("OPENAI_API_KEY") != "" {
		providerString = "openai/gpt-4o"
	} else if os.Getenv("GEMINI_API_KEY") != "" {
		providerString = "gemini/gemini-2.0-flash"
	}

	provider, err := llmutil.NewProviderFromString(providerString)
	if err != nil {
		// Fall back to mock provider
		fmt.Println("Note: No LLM API keys found. Using mock provider for demonstration.")
		fmt.Println("The mock will simulate system tool usage.")
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

	// Get system tools
	execTool, _ := tools.GetTool("execute_command")
	envTool, _ := tools.GetTool("get_environment_variable")
	sysInfoTool, _ := tools.GetTool("get_system_info")
	procTool, _ := tools.GetTool("process_list")

	// Create LLM agent
	deps := core.LLMDeps{
		Provider: provider,
	}
	agent := core.NewLLMAgent("system-assistant", "System Information Assistant", deps)

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

	// Add all system tools
	agent.AddTool(execTool)
	agent.AddTool(envTool)
	agent.AddTool(sysInfoTool)
	agent.AddTool(procTool)

	// Set minimal system prompt - let the tools guide the LLM
	agent.SetSystemPrompt(`You are a helpful system information assistant. You have access to tools for:
- Getting system information (OS, architecture, memory, etc.)
- Reading environment variables
- Listing and analyzing processes
- Executing system commands safely

When asked about the system, use the appropriate tools to get accurate information. The tools will guide you on their proper usage.`)

	// Example queries
	examples := []struct {
		title  string
		prompt string
	}{
		{
			title:  "System Overview",
			prompt: "What operating system and architecture is this running on?",
		},
		{
			title:  "Memory Information",
			prompt: "Show me the current memory usage statistics",
		},
		{
			title:  "Environment Check",
			prompt: "What Go-related environment variables are set?",
		},
		{
			title:  "Process Analysis",
			prompt: "What are the top 3 processes using the most memory?",
		},
		{
			title:  "Command Execution",
			prompt: "Show me the current directory and list its contents",
		},
		{
			title:  "Multi-Tool Query",
			prompt: "Tell me about this system: OS version, number of CPUs, current user (from environment), and how many processes are running",
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
		fmt.Println("=== System Tool Information ===")
		systemTools := tools.Tools.ListByCategory("system")
		for _, entry := range systemTools {
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
	fmt.Println("The LLM agent successfully used system tools to:")
	fmt.Println("• Get system information without explicit tool instructions")
	fmt.Println("• Query environment variables based on patterns")
	fmt.Println("• Analyze running processes")
	fmt.Println("• Execute safe system commands")
	fmt.Println("• Combine multiple tools to answer complex queries")
	fmt.Println("\nThe tools' built-in documentation guided the LLM on proper usage.")
}
