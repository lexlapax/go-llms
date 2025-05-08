package main

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/testutils"
)

// TestAgentExample tests the basic agent functionality
func TestAgentExample(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	agent := workflow.NewAgent(mockProvider)
	
	// Add a calculator tool
	agent.AddTool(testutils.CreateCalculatorTool())
	
	// Set a system prompt
	agent.SetSystemPrompt("You are a helpful math assistant.")
	
	// Run the agent
	result, err := agent.Run(context.Background(), "What is 2+2?")
	if err != nil {
		t.Fatalf("Agent failed to run: %v", err)
	}
	
	// The result should not be empty
	if result == "" {
		t.Errorf("Expected non-empty result, got empty string")
	}
}

// TestCachedAgentExample tests the cached agent capabilities
func TestCachedAgentExample(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	agent := workflow.NewCachedAgent(mockProvider)
	
	// Add a calculator tool
	agent.AddTool(testutils.CreateCalculatorTool())
	
	// Set a system prompt
	agent.SetSystemPrompt("You are a helpful assistant.")
	
	// Run the agent twice with the same query to test caching
	result1, err := agent.Run(context.Background(), "What is 2+2?")
	if err != nil {
		t.Fatalf("First agent run failed: %v", err)
	}
	
	result2, err := agent.Run(context.Background(), "What is 2+2?")
	if err != nil {
		t.Fatalf("Second agent run failed: %v", err)
	}
	
	// Results should not be empty
	if result1 == "" || result2 == "" {
		t.Errorf("Expected non-empty results")
	}
	
	// Check cache stats to verify caching worked
	stats := agent.GetCacheStats()
	if stats["hits"].(int) < 1 {
		t.Logf("Cache stats: %v", stats)
		t.Errorf("Expected at least 1 cache hit, got %d", stats["hits"].(int))
	}
}

// TestMessageManagerExample tests the message manager functionality
func TestMessageManagerExample(t *testing.T) {
	// Test message manager with modest limits
	config := workflow.MessageManagerConfig{
		UseTokenTruncation:   true,
		KeepAllSystemMessages: true,
	}
	manager := workflow.NewMessageManager(5, 500, config)
	
	// Test message management (rest of test implementation)
	if manager.GetMessageCount() != 0 {
		t.Errorf("Expected 0 messages initially, got %d", manager.GetMessageCount())
	}
}

// TestToolExecutorExample tests the tool executor functionality
func TestToolExecutorExample(t *testing.T) {
	// Create tool map
	toolMap := make(map[string]domain.Tool)
	
	// Add calculator tool
	toolMap["calculator"] = testutils.MockTool{
		ToolName:        "calculator",
		ToolDescription: "Perform mathematical calculations",
		Executor: func(ctx context.Context, params interface{}) (interface{}, error) {
			return map[string]interface{}{
				"result": 4,
			}, nil
		},
	}
	
	// Add date tool
	toolMap["date"] = testutils.MockTool{
		ToolName:        "date",
		ToolDescription: "Get the current date",
		Executor: func(ctx context.Context, params interface{}) (interface{}, error) {
			return map[string]string{
				"date": time.Now().Format("2006-01-02"),
			}, nil
		},
	}
	
	// Create executor
	executor := workflow.NewToolExecutor(toolMap, 2, 1*time.Second, nil)
	
	// Execute multiple tools
	toolNames := []string{"calculator", "date"}
	params := []interface{}{
		map[string]interface{}{"expression": "2+2"},
		nil,
	}
	
	results := executor.ExecuteToolsParallel(context.Background(), toolNames, params)
	
	// Check we have results for both tools
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	
	// Calculator tool should have succeeded
	calcResult, exists := results["calculator"]
	if !exists || calcResult.Status != workflow.ToolStatusSuccess {
		t.Errorf("Expected calculator success, got %v", calcResult.Status)
	}
	
	// Date tool should have succeeded
	dateResult, exists := results["date"]
	if !exists || dateResult.Status != workflow.ToolStatusSuccess {
		t.Errorf("Expected date success, got %v", dateResult.Status)
	}
}