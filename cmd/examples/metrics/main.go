package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// DummyTool is a simple tool that just waits a bit and returns a result
type DummyTool struct {
	name        string
	delay       time.Duration
	failPercent int
}

// NewDummyTool creates a dummy tool
func NewDummyTool(name string, delay time.Duration, failPercent int) *DummyTool {
	return &DummyTool{
		name:        name,
		delay:       delay,
		failPercent: failPercent,
	}
}

// CalculatorTool is a simple calculator tool
type CalculatorTool struct{}

// NewCalculatorTool creates a new calculator tool
func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

// Name returns the tool's name
func (t *CalculatorTool) Name() string {
	return "calculator"
}

// Description provides information about the tool
func (t *CalculatorTool) Description() string {
	return "A simple calculator that can add, subtract, multiply and divide numbers"
}

// CalculatorParams defines parameters for the calculator tool
type CalculatorParams struct {
	Operation string  `json:"operation"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
}

// Execute performs the calculation
func (t *CalculatorTool) Execute(ctx context.Context, params interface{}) (interface{}, error) {
	// Quick simulation of processing time
	time.Sleep(10 * time.Millisecond)
	
	// Try to convert params to CalculatorParams
	var calcParams CalculatorParams
	
	// If params is a map, try to extract the values
	if paramsMap, ok := params.(map[string]interface{}); ok {
		if op, ok := paramsMap["operation"].(string); ok {
			calcParams.Operation = op
		}
		
		if a, ok := paramsMap["a"].(float64); ok {
			calcParams.A = a
		} else if aStr, ok := paramsMap["a"].(string); ok {
			if aVal, err := strconv.ParseFloat(aStr, 64); err == nil {
				calcParams.A = aVal
			}
		}
		
		if b, ok := paramsMap["b"].(float64); ok {
			calcParams.B = b
		} else if bStr, ok := paramsMap["b"].(string); ok {
			if bVal, err := strconv.ParseFloat(bStr, 64); err == nil {
				calcParams.B = bVal
			}
		}
	}
	
	// Perform the calculation
	switch calcParams.Operation {
	case "add":
		return map[string]interface{}{
			"result": calcParams.A + calcParams.B,
		}, nil
	case "subtract":
		return map[string]interface{}{
			"result": calcParams.A - calcParams.B,
		}, nil
	case "multiply":
		return map[string]interface{}{
			"result": calcParams.A * calcParams.B,
		}, nil
	case "divide":
		if calcParams.B == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return map[string]interface{}{
			"result": calcParams.A / calcParams.B,
		}, nil
	default:
		return nil, fmt.Errorf("unknown operation: %s", calcParams.Operation)
	}
}

// ParameterSchema returns the schema for the calculator parameters
func (t *CalculatorTool) ParameterSchema() *sdomain.Schema {
	return &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"operation": {
				Type:        "string",
				Description: "The operation to perform (add, subtract, multiply, divide)",
			},
			"a": {
				Type:        "number",
				Description: "The first number",
			},
			"b": {
				Type:        "number",
				Description: "The second number",
			},
		},
		Required: []string{"operation", "a", "b"},
	}
}

// Name returns the tool's name
func (t *DummyTool) Name() string {
	return t.name
}

// Description provides information about the tool
func (t *DummyTool) Description() string {
	return fmt.Sprintf("A dummy tool that waits for %v and has a %d%% chance of failing", t.delay, t.failPercent)
}

// Execute runs the tool with parameters
func (t *DummyTool) Execute(ctx context.Context, params interface{}) (interface{}, error) {
	// Simulate processing time
	select {
	case <-time.After(t.delay):
		// Continue execution
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Random failure based on failPercent
	if t.failPercent > 0 && time.Now().UnixNano()%100 < int64(t.failPercent) {
		return nil, fmt.Errorf("tool %s failed (simulated failure)", t.name)
	}

	// Return a dummy result
	return map[string]interface{}{
		"tool":      t.name,
		"timestamp": time.Now().Format(time.RFC3339),
		"params":    params,
	}, nil
}

// ParameterSchema returns the schema for tool parameters
func (t *DummyTool) ParameterSchema() *sdomain.Schema {
	return &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"query": {
				Type:        "string",
				Description: "The input query",
			},
		},
		Required: []string{"query"},
	}
}

func main() {
	// Setup structured logging
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Create provider
	mockProvider := provider.NewMockProvider()

	// Create the metrics hook
	metricsHook := workflow.NewMetricsHook()

	// Create the logging hook
	loggingHook := workflow.NewLoggingHook(logger, workflow.LogLevelDetailed)

	// Create agent with both hooks
	agent := workflow.NewAgent(mockProvider).
		WithHook(metricsHook).
		WithHook(loggingHook)

	// Add some tools with different characteristics
	agent.AddTool(NewDummyTool("fastTool", 50*time.Millisecond, 0))
	agent.AddTool(NewDummyTool("slowTool", 200*time.Millisecond, 0))
	agent.AddTool(NewDummyTool("unreliableTool", 100*time.Millisecond, 30))
	agent.AddTool(NewCalculatorTool())

	// Setup context with metrics tracking
	ctx := workflow.WithMetrics(context.Background())

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
func runAgentOperations(ctx context.Context, agent domain.Agent, count int) {
	prompts := []string{
		"Calculate 123 + 456",
		"What's the capital of France?",
		"Use the fastTool with query 'test'",
		"Use the slowTool with query 'analysis'",
		"Use the unreliableTool with query 'risky'",
		"Tell me about machine learning",
		"Explain quantum computing",
		"What is the square root of 144?",
	}

	for i := 0; i < count && i < len(prompts); i++ {
		fmt.Printf("\n➡️ Running operation %d: %s\n", i+1, prompts[i])
		
		result, err := agent.Run(ctx, prompts[i])
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}
		
		fmt.Printf("✅ Result: %s\n", truncateString(fmt.Sprintf("%v", result), 60))
		
		// Add a small delay between operations
		time.Sleep(100 * time.Millisecond)
	}
}

// printMetrics prints the collected metrics
func printMetrics(metrics workflow.Metrics) {
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