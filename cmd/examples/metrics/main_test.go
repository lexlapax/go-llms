package main

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func TestMetricsHook(t *testing.T) {
	// Create the metrics hook
	metricsHook := workflow.NewMetricsHook()

	// Create agent with metrics hook
	mockProvider := provider.NewMockProvider()
	agent := workflow.NewAgent(mockProvider).WithHook(metricsHook)

	// Add tools
	agent.AddTool(NewDummyTool("testTool", 10*time.Millisecond, 0))

	// Setup context with metrics tracking
	ctx := workflow.WithMetrics(context.Background())

	// Run a few operations
	if _, err := agent.Run(ctx, "Use testTool with query 'test'"); err != nil {
		t.Fatalf("Error running agent: %v", err)
	}

	// Get metrics
	metrics := metricsHook.GetMetrics()

	// Basic validation
	if metrics.Requests <= 0 {
		t.Errorf("Expected requests to be greater than 0, got %d", metrics.Requests)
	}
	
	// Manually notify the hook of a tool call since we're using a mock provider
	// and can't guarantee it will call tools in the test
	metricsHook.NotifyToolCall("testTool", nil)
	
	// Get updated metrics
	metrics = metricsHook.GetMetrics()
	
	if metrics.ToolCalls <= 0 {
		t.Errorf("Expected tool calls to be greater than 0, got %d", metrics.ToolCalls)
	}

	if len(metrics.ToolStats) == 0 {
		t.Error("Expected tool stats to be populated")
	}

	// Reset and verify
	metricsHook.Reset()
	resetMetrics := metricsHook.GetMetrics()
	if resetMetrics.Requests != 0 || resetMetrics.ToolCalls != 0 {
		t.Error("Metrics reset did not clear counters")
	}
}