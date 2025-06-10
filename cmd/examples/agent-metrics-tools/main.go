package main

// ABOUTME: Example demonstrating metrics collection for LLM agent operations with built-in tools
// ABOUTME: Shows how to track latency, token usage, and tool execution metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"

	// Import tool categories
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/math"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)

func main() {
	// Setup structured logging
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Create provider - try to use a real provider if available
	var llmProvider llmdomain.Provider
	var providerName, modelName string

	// Try OpenAI first
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		providerName = "openai"
		modelName = "gpt-4o-mini"
		llmProvider = provider.NewOpenAIProvider(apiKey, modelName)
	} else if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		providerName = "anthropic"
		modelName = "claude-3-haiku-20240307"
		llmProvider = provider.NewAnthropicProvider(apiKey, modelName)
	} else if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		providerName = "gemini"
		modelName = "gemini-1.5-flash"
		llmProvider = provider.NewGeminiProvider(apiKey, modelName)
	} else {
		fmt.Println("Note: No LLM API keys found. Using mock provider for demonstration.")
		fmt.Println("Set ANTHROPIC_API_KEY, OPENAI_API_KEY, or GEMINI_API_KEY for real LLM usage.")
		fmt.Println()
		providerName = "mock"
		modelName = "mock-model"
		llmProvider = createMockProvider()
	}

	fmt.Printf("Using %s provider with model %s\n\n", providerName, modelName)

	// Create the metrics hook
	metricsHook := core.NewLLMMetricsHook()

	// Create the logging hook
	loggingHook := core.NewLoggingHook(logger, core.LogLevelDetailed)

	// Create agent with both hooks
	deps := core.LLMDeps{
		Provider: llmProvider,
	}
	agent := core.NewLLMAgent("metrics-agent", "Metrics Demo Agent", deps).
		WithHook(metricsHook).
		WithHook(loggingHook)

	// Add some tools with different characteristics
	agent.AddTool(tools.MustGetTool("calculator"))
	agent.AddTool(tools.MustGetTool("web_fetch"))
	agent.AddTool(tools.MustGetTool("file_list"))
	agent.AddTool(tools.MustGetTool("datetime_now"))

	// Set a system prompt to help the agent understand the tools
	agent.SetSystemPrompt(`You are a helpful assistant with access to several tools:
- calculator: Can perform basic math operations (add, subtract, multiply, divide)
- web_fetch: Fetch content from web URLs
- file_list: List files in directories
- datetime_now: Get current date and time

When asked to calculate, use the calculator tool.
When asked about time, use the datetime_now tool.
When asked about files, use the file_list tool.
Be concise in your responses.`)

	// Setup context
	ctx := context.Background()

	fmt.Println("🔍 Running agent with metrics collection")
	fmt.Println("==========================================")

	// Run several agent operations
	runAgentOperations(ctx, agent, 5)

	// Get and display metrics
	metrics := metricsHook.GetMetrics()
	printMetrics(metrics)

	// Reset metrics for a new test
	metricsHook.Reset()
	fmt.Println("\n🔄 Metrics reset, running more operations...")
	fmt.Println("==========================================")

	// Run more operations
	runAgentOperations(ctx, agent, 3)

	// Get and display metrics again
	metrics = metricsHook.GetMetrics()
	printMetrics(metrics)
}

// runAgentOperations runs the agent multiple times with different prompts
func runAgentOperations(ctx context.Context, agent *core.LLMAgent, count int) {
	prompts := []string{
		"Calculate 123 + 456 using the calculator",
		"Calculate 50 * 20 using the calculator tool",
		"What time is it right now?",
		"List files in the current directory",
		"Calculate 100 / 4 using the calculator",
		"Calculate 999 - 333 using the calculator tool",
		"Show me the current date and time",
		"List .go files in the current directory",
	}

	for i := 0; i < count && i < len(prompts); i++ {
		fmt.Printf("\n➡️ Running operation %d: %s\n", i+1, prompts[i])

		// Create state with the prompt
		state := domain.NewState()
		state.Set("user_input", prompts[i])

		resultState, err := agent.Run(ctx, state)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		// Extract result from state
		if result, exists := resultState.Get("result"); exists {
			fmt.Printf("✅ Result: %s\n", truncateString(fmt.Sprintf("%v", result), 60))
		} else {
			fmt.Printf("⚠️ No result found in state\n")
		}

		// Add a small delay between operations
		time.Sleep(100 * time.Millisecond)
	}
}

// printMetrics prints the collected metrics
func printMetrics(metrics core.Metrics) {
	fmt.Println("\n📊 Agent Metrics Report")
	fmt.Println("====================")
	fmt.Printf("Total Requests:      %d\n", metrics.Requests)
	fmt.Printf("Total Tool Calls:    %d\n", metrics.ToolCalls)
	fmt.Printf("Error Count:         %d\n", metrics.ErrorCount)
	fmt.Printf("Estimated Tokens:    %d\n", metrics.TotalTokens)
	fmt.Printf("Avg Generation Time: %.2f ms\n", metrics.AverageGenTimeMs)

	if len(metrics.ToolStats) > 0 {
		fmt.Println("\n🔧 Tool Statistics")
		fmt.Println("-----------------")

		// Convert to JSON for pretty formatting
		toolStatsJSON, _ := json.MarshalIndent(metrics.ToolStats, "", "  ")
		fmt.Println(string(toolStatsJSON))
	}
}

// truncateString truncates a string if it's too long
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// createMockProvider creates a mock provider for demonstration
func createMockProvider() llmdomain.Provider {
	mockProvider := provider.NewMockProvider()
	toolCallCount := 0

	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []llmdomain.Message, options ...llmdomain.Option) (llmdomain.Response, error) {
		// Track tool calls
		toolCallCount++

		// Check if this is a tool result
		var hasToolResult bool
		for _, msg := range messages {
			if msg.Role == "user" {
				for _, part := range msg.Content {
					if part.Type == llmdomain.ContentTypeText && strings.Contains(part.Text, "Tool results:") {
						hasToolResult = true
						break
					}
				}
			}
		}

		if hasToolResult {
			// Return final response after tool execution
			return llmdomain.Response{
				Content: "I've successfully executed the requested operation. The metrics show the tool execution time and results.",
			}, nil
		}

		// Extract the last user message
		var lastUserMsg string
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				for _, part := range messages[i].Content {
					if part.Type == llmdomain.ContentTypeText {
						lastUserMsg = part.Text
						break
					}
				}
				if lastUserMsg != "" {
					break
				}
			}
		}

		// Generate tool calls based on the prompt
		if strings.Contains(lastUserMsg, "Calculate") || strings.Contains(lastUserMsg, "calculate") {
			if strings.Contains(lastUserMsg, "+") || strings.Contains(lastUserMsg, "add") {
				return llmdomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "add", "a": 123, "b": 456}}`,
				}, nil
			} else if strings.Contains(lastUserMsg, "*") || strings.Contains(lastUserMsg, "multiply") {
				return llmdomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "multiply", "a": 50, "b": 20}}`,
				}, nil
			} else if strings.Contains(lastUserMsg, "/") || strings.Contains(lastUserMsg, "divide") {
				return llmdomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "divide", "a": 100, "b": 4}}`,
				}, nil
			} else if strings.Contains(lastUserMsg, "-") || strings.Contains(lastUserMsg, "subtract") {
				return llmdomain.Response{
					Content: `{"tool": "calculator", "params": {"operation": "subtract", "a": 999, "b": 333}}`,
				}, nil
			}
		} else if strings.Contains(lastUserMsg, "time") || strings.Contains(lastUserMsg, "date") {
			return llmdomain.Response{
				Content: `{"tool": "datetime_now", "params": {}}`,
			}, nil
		} else if strings.Contains(lastUserMsg, "files") || strings.Contains(lastUserMsg, "List") {
			if strings.Contains(lastUserMsg, ".go") {
				return llmdomain.Response{
					Content: `{"tool": "file_list", "params": {"path": ".", "pattern": "*.go"}}`,
				}, nil
			}
			return llmdomain.Response{
				Content: `{"tool": "file_list", "params": {"path": "."}}`,
			}, nil
		}

		// Default response
		return llmdomain.Response{
			Content: "I understand your request. Let me help you with that.",
		}, nil
	})

	return mockProvider
}
