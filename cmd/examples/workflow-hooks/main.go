// ABOUTME: Example demonstrating hook integration with workflow agents
// ABOUTME: Shows how metrics and logging hooks work with sequential and parallel workflows

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
)

func main() {
	// Example 1: Sequential workflow with hooks
	sequentialWithHooksExample()

	// Example 2: Parallel workflow with hooks
	parallelWithHooksExample()
}

func sequentialWithHooksExample() {
	fmt.Println("=== Sequential Workflow with Hooks Example ===")
	fmt.Println("Running a sequential workflow with metrics and logging hooks...")
	fmt.Println()

	ctx := context.Background()

	// Create mock agents
	agent1 := createMockAgent("data-processor", "Processes input data", 100*time.Millisecond)
	agent2 := createMockAgent("data-analyzer", "Analyzes processed data", 150*time.Millisecond)
	agent3 := createMockAgent("data-formatter", "Formats analysis results", 50*time.Millisecond)

	// Create hooks
	metricsHook := core.NewLLMMetricsHook()
	loggingHook := core.NewLoggingHook(slog.Default(), core.LogLevelBasic)

	// Create sequential workflow with hooks
	workflow := workflow.NewSequentialAgent("data-pipeline").
		WithHook(metricsHook).
		WithHook(loggingHook).
		AddAgent(agent1).
		AddAgent(agent2).
		AddAgent(agent3)

	// Create initial state
	initialState := domain.NewState()
	initialState.Set("data", "raw input data")

	// Run workflow
	fmt.Println("Starting sequential workflow...")
	start := time.Now()

	result, err := workflow.Run(ctx, initialState)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("Sequential workflow completed in %v\n", duration)

	// Display results
	if data, exists := result.Get("processed_data"); exists {
		fmt.Printf("Final result: %v\n", data)
	}

	// Display metrics
	fmt.Printf("\n--- Metrics Summary ---\n")
	metrics := metricsHook.GetMetrics()
	fmt.Printf("Total requests: %d\n", metrics.Requests)
	fmt.Printf("Total errors: %d\n", metrics.ErrorCount)
	fmt.Printf("Total tokens: %d\n", metrics.TotalTokens)
	fmt.Printf("Average generation time: %.2f ms\n", metrics.AverageGenTimeMs)

	fmt.Println()
}

func parallelWithHooksExample() {
	fmt.Println("=== Parallel Workflow with Hooks Example ===")
	fmt.Println("Running a parallel workflow with metrics and logging hooks...")
	fmt.Println()

	ctx := context.Background()

	// Create mock agents with different delays
	agent1 := createMockAgent("fast-processor", "Fast processing", 100*time.Millisecond)
	agent2 := createMockAgent("medium-processor", "Medium processing", 200*time.Millisecond)
	agent3 := createMockAgent("slow-processor", "Slow processing", 300*time.Millisecond)

	// Create hooks
	metricsHook := core.NewLLMMetricsHook()
	loggingHook := core.NewLoggingHook(slog.Default(), core.LogLevelBasic)

	// Create parallel workflow with hooks
	workflow := workflow.NewParallelAgent("parallel-processors").
		WithMaxConcurrency(3).
		WithMergeStrategy(workflow.MergeAll).
		WithHook(metricsHook).
		WithHook(loggingHook).
		AddAgent(agent1).
		AddAgent(agent2).
		AddAgent(agent3)

	// Create initial state
	initialState := domain.NewState()
	initialState.Set("task", "parallel processing task")

	// Run workflow
	fmt.Println("Starting parallel workflow...")
	start := time.Now()

	result, err := workflow.Run(ctx, initialState)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("Parallel workflow completed in %v\n", duration)

	// Display results
	if parallelResults, exists := result.Get("parallel_results"); exists {
		results := parallelResults.(map[string]interface{})
		fmt.Printf("Number of parallel results: %d\n", len(results))

		for agentName := range results {
			fmt.Printf("- %s completed\n", agentName)
		}
	}

	// Display metrics
	fmt.Printf("\n--- Metrics Summary ---\n")
	metrics := metricsHook.GetMetrics()
	fmt.Printf("Total requests: %d\n", metrics.Requests)
	fmt.Printf("Total errors: %d\n", metrics.ErrorCount)
	fmt.Printf("Total tokens: %d\n", metrics.TotalTokens)
	fmt.Printf("Average generation time: %.2f ms\n", metrics.AverageGenTimeMs)

	fmt.Println()
}

// Helper function to create mock agents
func createMockAgent(name, description string, delay time.Duration) domain.BaseAgent {
	agent := &mockAgent{
		BaseAgent:   core.NewBaseAgent(name, description, domain.AgentTypeCustom),
		delay:       delay,
		description: description,
	}
	return agent
}

type mockAgent struct {
	domain.BaseAgent
	delay       time.Duration
	description string
}

func (m *mockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Simulate processing delay
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	newState := state.Clone()

	// Add processed data
	if data, exists := state.Get("data"); exists {
		newState.Set("processed_data", fmt.Sprintf("%s -> processed by %s", data, m.Name()))
	} else if task, exists := state.Get("task"); exists {
		newState.Set("result", fmt.Sprintf("%s -> processed by %s", task, m.Name()))
	} else {
		newState.Set("result", fmt.Sprintf("Processed by %s (%s)", m.Name(), m.description))
	}

	return newState, nil
}
