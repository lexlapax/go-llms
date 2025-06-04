package main

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

func TestProviderCreation(t *testing.T) {
	// Test mock provider creation (always works without API keys)
	config := llmutil.ModelConfig{
		Provider: "mock",
		Model:    "mock-model",
		APIKey:   "not-needed",
	}

	llmProvider, err := llmutil.CreateProvider(config)
	if err != nil {
		t.Errorf("Failed to create mock provider: %v", err)
	}

	if llmProvider == nil {
		t.Error("Provider is nil despite no error")
	}
}

func TestBatchGenerate(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	prompts := []string{
		"What is the capital of France?",
		"Give me a recipe for pancakes",
		"How many planets are in our solar system?",
	}

	results, errors := llmutil.BatchGenerate(context.Background(), mockProvider, prompts)

	// Check that we got the right number of results and errors
	if len(results) != len(prompts) {
		t.Errorf("Expected %d results, got %d", len(prompts), len(results))
	}

	if len(errors) != len(prompts) {
		t.Errorf("Expected %d errors, got %d", len(prompts), len(errors))
	}

	// Check that all results were generated successfully
	for i, err := range errors {
		if err != nil {
			t.Errorf("Error in batch generation for prompt %d: %v", i, err)
		}
		if results[i] == "" {
			t.Errorf("Empty result for prompt %d", i)
		}
	}
}

func TestProviderPool(t *testing.T) {
	// Create mock providers
	provider1 := provider.NewMockProvider()
	provider2 := provider.NewMockProvider()
	provider3 := provider.NewMockProvider()

	// Create provider pool with round-robin strategy
	providerPool := llmutil.NewProviderPool(
		[]domain.Provider{provider1, provider2, provider3},
		llmutil.StrategyRoundRobin,
	)

	// Test multiple generations to exercise the pool
	for i := 0; i < 5; i++ {
		result, err := providerPool.Generate(
			context.Background(),
			"Test prompt",
		)

		if err != nil {
			t.Errorf("Error in pool generation %d: %v", i, err)
		}

		if result == "" {
			t.Errorf("Empty result for generation %d", i)
		}
	}

	// Check metrics
	metrics := providerPool.GetMetrics()
	totalRequests := 0
	for _, m := range metrics {
		totalRequests += m.Requests
	}

	if totalRequests != 5 {
		t.Errorf("Expected 5 total requests in metrics, got %d", totalRequests)
	}
}

func TestGenerateWithRetry(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	result, err := llmutil.GenerateWithRetry(
		context.Background(),
		mockProvider,
		"Test prompt",
		3, // max retries
	)

	if err != nil {
		t.Errorf("Error in generate with retry: %v", err)
	}

	if result == "" {
		t.Error("Empty result from generate with retry")
	}
}

func TestAgentCreation(t *testing.T) {
	mockProvider := provider.NewMockProvider()

	// Create a simple calculator tool
	calculatorTool := tools.NewTool(
		"calculator",
		"Perform mathematical calculations",
		func(params struct {
			Expression string `json:"expression"`
		}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"success":    true,
				"expression": params.Expression,
				"result":     42, // Fixed result for testing
			}, nil
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate",
				},
			},
			Required: []string{"expression"},
		},
	)

	// Create the agent using core.LLMAgent
	agent := core.NewAgent("test-agent", mockProvider)
	agent.SetSystemPrompt("You are a helpful assistant with access to tools.")
	agent.AddTool(calculatorTool)

	if agent == nil {
		t.Error("Agent creation failed, agent is nil")
	}

	// Test running the agent with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create state for the agent
	state := agentDomain.NewState()
	state.Set("prompt", "What is 7 * 6?")

	resultState, err := agent.Run(ctx, state)

	if err != nil {
		t.Errorf("Error running agent: %v", err)
	}

	if resultState == nil {
		t.Error("Agent result state is nil")
	} else {
		if result, exists := resultState.Get("result"); !exists {
			t.Error("No result in state")
		} else if result == nil {
			t.Error("Agent result is nil")
		}
	}
}

func TestRunWithTimeout(t *testing.T) {
	mockProvider := provider.NewMockProvider()

	// Create a simple agent
	agent := core.NewAgent("test-agent", mockProvider)
	agent.SetSystemPrompt("You are a helpful assistant.")

	// Run with a reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	state := agentDomain.NewState()
	state.Set("prompt", "What is 7 * 6?")

	resultState, err := agent.Run(ctx, state)

	if err != nil {
		t.Errorf("Error running with timeout: %v", err)
	}

	if resultState == nil {
		t.Error("Result state from agent is nil")
	} else {
		if result, exists := resultState.Get("result"); !exists {
			t.Error("No result in state")
		} else if result == nil {
			t.Error("Result from agent is nil")
		}
	}

	// Run with a very short timeout to test timeout handling
	// Note: This is expected to fail with a timeout error
	shortCtx, shortCancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer shortCancel()

	state2 := agentDomain.NewState()
	state2.Set("prompt", "Complex question that requires thinking")

	_, err = agent.Run(shortCtx, state2)

	// This may or may not timeout depending on the mock implementation speed
	// So we don't test the error condition here as it's not deterministic
}

// Use the float64Ptr function defined in main.go
