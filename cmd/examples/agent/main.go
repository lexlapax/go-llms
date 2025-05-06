package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func main() {
	// Check for API keys in environment variables
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENAI_API_KEY environment variable")
		os.Exit(1)
	}

	// Create an LLM provider
	llmProvider := provider.NewOpenAIProvider(apiKey, "gpt-4")

	// Create an agent
	agent := workflow.NewAgent(llmProvider)

	// Add a system prompt
	agent.SetSystemPrompt("You are a helpful assistant that can answer questions and use tools when necessary.")

	// Set the model name
	agent.WithModel("gpt-4")

	// Create and add a logger
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	agent.WithHook(workflow.NewLoggingHook(logger, workflow.LogLevelDetailed))

	// Add metrics hook
	metricsHook := workflow.NewMetricsHook()
	agent.WithHook(metricsHook)

	// Add tools to the agent
	addTools(agent)

	// Run the agent with a user query
	fmt.Println("Running the agent with a query...")
	// Use context with metrics
	ctx := workflow.WithMetrics(context.Background())
	result, err := agent.Run(ctx, "What is the current year? Then tell me what 25 * 42 is.")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nAgent result:")
	fmt.Println("-------------")
	fmt.Println(result)

	// Now run with a schema
	schema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"answer": {
				Type:        "string",
				Description: "The answer to the user's question",
			},
			"calculations": {
				Type:        "array",
				Description: "Any calculations performed",
				Items: &sdomain.Property{
					Type: "string",
				},
			},
		},
		Required: []string{"answer"},
	}

	fmt.Println("\nRunning the agent with a schema...")
	structuredResult, err := agent.RunWithSchema(ctx, "Calculate 15 + 27 and 33 * 4", schema)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nStructured result:")
	fmt.Println("-----------------")
	resultMap, ok := structuredResult.(map[string]interface{})
	if !ok {
		fmt.Println("Unexpected result type")
		os.Exit(1)
	}

	fmt.Printf("Answer: %s\n", resultMap["answer"])
	if calculations, ok := resultMap["calculations"].([]interface{}); ok {
		fmt.Println("Calculations:")
		for _, calc := range calculations {
			fmt.Printf("  - %s\n", calc)
		}
	}

	// Display metrics
	fmt.Println("\nMetrics:")
	fmt.Println("-----------------")
	metrics := metricsHook.GetMetrics()
	fmt.Printf("Total requests: %d\n", metrics.Requests)
	fmt.Printf("Tool calls: %d\n", metrics.ToolCalls)
	fmt.Printf("Errors: %d\n", metrics.ErrorCount)
	fmt.Printf("Estimated tokens: %d\n", metrics.TotalTokens)
	fmt.Printf("Avg generation time: %.2f ms\n", metrics.AverageGenTimeMs)

	if len(metrics.ToolStats) > 0 {
		fmt.Println("\nTool Statistics:")
		for tool, stats := range metrics.ToolStats {
			fmt.Printf("- %s: %d calls, avg time: %.2f ms\n",
				tool, stats.Calls, stats.AverageTimeMs)
		}
	}
}

// addTools adds various tools to the agent
func addTools(agent domain.Agent) {
	// Add a date tool
	agent.AddTool(tools.NewTool(
		"get_current_date",
		"Get the current date",
		func() map[string]string {
			return map[string]string{
				"date": fmt.Sprintf("%s", time.Now().Format("2006-01-02")),
				"time": fmt.Sprintf("%s", time.Now().Format("15:04:05")),
				"year": fmt.Sprintf("%d", time.Now().Year()),
			}
		},
		&sdomain.Schema{
			Type:        "object",
			Description: "Returns the current date and time",
		},
	))

	// Add a calculator tool
	agent.AddTool(tools.NewTool(
		"calculator",
		"Perform mathematical calculations",
		func(params struct {
			Expression string `json:"expression"`
		}) (float64, error) {
			// In a real implementation, this would use a proper expression parser
			// For simplicity, here we'll just handle some basic cases
			parts := strings.Split(params.Expression, "*")
			if len(parts) == 2 {
				// Multiplication
				a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				if err != nil {
					return 0, fmt.Errorf("invalid number: %s", parts[0])
				}
				b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err != nil {
					return 0, fmt.Errorf("invalid number: %s", parts[1])
				}
				return a * b, nil
			}

			parts = strings.Split(params.Expression, "+")
			if len(parts) == 2 {
				// Addition
				a, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				if err != nil {
					return 0, fmt.Errorf("invalid number: %s", parts[0])
				}
				b, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err != nil {
					return 0, fmt.Errorf("invalid number: %s", parts[1])
				}
				return a + b, nil
			}

			// Handle other operations as needed...
			return 0, fmt.Errorf("unsupported operation in expression: %s", params.Expression)
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate",
				},
			},
			Required: []string{"expression"},
		},
	))
}
