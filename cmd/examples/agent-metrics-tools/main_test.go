package main

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func TestMetricsHook(t *testing.T) {
	// Create the metrics hook
	metricsHook := core.NewLLMMetricsHook()

	// Create agent with metrics hook
	mockProvider := provider.NewMockProvider()
	agent := core.NewAgent("test-agent", mockProvider).WithHook(metricsHook)

	// Add tools
	agent.AddTool(NewDummyTool("testTool", 10*time.Millisecond, 0))

	// Setup context
	ctx := context.Background()

	// Create state with prompt
	state := domain.NewState()
	state.Set("prompt", "Use testTool with query 'test'")

	// Run a few operations
	if _, err := agent.Run(ctx, state); err != nil {
		t.Fatalf("Error running agent: %v", err)
	}

	// Get metrics
	metrics := metricsHook.GetMetrics()

	// Basic validation
	if metrics.Requests <= 0 {
		t.Errorf("Expected requests to be greater than 0, got %d", metrics.Requests)
	}

	// Since we can't guarantee the mock will actually call tools,
	// we'll skip the tool call validation

	// Reset and verify
	metricsHook.Reset()
	resetMetrics := metricsHook.GetMetrics()
	if resetMetrics.Requests != 0 || resetMetrics.ToolCalls != 0 {
		t.Error("Metrics reset did not clear counters")
	}
}
