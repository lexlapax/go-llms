package main

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/testutils"
)

// TestAgentExample tests the basic agent functionality
func TestAgentExample(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	agent := core.NewAgent("test-agent", mockProvider)

	// Add a calculator tool
	agent.AddTool(testutils.CreateCalculatorTool())

	// Set a system prompt
	agent.SetSystemPrompt("You are a helpful math assistant.")

	// Run the agent
	state := domain.NewState()
	state.Set("prompt", "What is 2+2?")
	resultState, err := agent.Run(context.Background(), state)
	if err != nil {
		t.Fatalf("Agent failed to run: %v", err)
	}

	// The result should not be empty
	result, exists := resultState.Get("result")
	if !exists || result == "" {
		t.Errorf("Expected non-empty result, got empty or missing")
	}
}

// TestLLMAgentExample tests the LLM agent capabilities
func TestLLMAgentExample(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	agent := core.NewAgent("test-agent", mockProvider)

	// Add a calculator tool
	agent.AddTool(testutils.CreateCalculatorTool())

	// Set a system prompt
	agent.SetSystemPrompt("You are a helpful assistant.")

	// Run the agent twice with the same query
	state1 := domain.NewState()
	state1.Set("prompt", "What is 2+2?")
	result1State, err := agent.Run(context.Background(), state1)
	if err != nil {
		t.Fatalf("First agent run failed: %v", err)
	}

	state2 := domain.NewState()
	state2.Set("prompt", "What is 2+2?")
	result2State, err := agent.Run(context.Background(), state2)
	if err != nil {
		t.Fatalf("Second agent run failed: %v", err)
	}

	// Results should not be empty
	result1, exists1 := result1State.Get("result")
	result2, exists2 := result2State.Get("result")
	if !exists1 || !exists2 || result1 == "" || result2 == "" {
		t.Errorf("Expected non-empty results")
	}
}

// TestAgentWithTools tests agent with multiple tools
func TestAgentWithTools(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	agent := core.NewAgent("test-agent", mockProvider)

	// Add multiple tools
	agent.AddTool(testutils.CreateCalculatorTool())

	// Create a mock date tool
	agent.AddTool(testutils.MockTool{
		ToolName:        "date",
		ToolDescription: "Get the current date",
		Executor: func(ctx context.Context, params interface{}) (interface{}, error) {
			return map[string]string{
				"date": "2025-02-03",
			}, nil
		},
	})

	// Set system prompt
	agent.SetSystemPrompt("You are a helpful assistant with access to tools.")

	// Test listing tools
	tools := agent.ListTools()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	// Test getting a specific tool
	calcTool, exists := agent.GetTool("calculator")
	if !exists {
		t.Errorf("Expected calculator tool to exist")
	}
	if calcTool.Name() != "calculator" {
		t.Errorf("Expected tool name 'calculator', got '%s'", calcTool.Name())
	}

	// Run agent with a query that might use tools
	state := domain.NewState()
	state.Set("prompt", "What is 10 * 5 and what's today's date?")
	resultState, err := agent.Run(context.Background(), state)
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	// Check result exists
	result, exists := resultState.Get("result")
	if !exists || result == nil {
		t.Errorf("Expected result in state")
	}
}
