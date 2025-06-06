package integration

// ABOUTME: Integration tests for workflow agents (Sequential, Parallel, Conditional, Loop)
// ABOUTME: Tests workflow execution patterns with mock providers

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// TestSequentialWorkflow tests sequential agent execution
func TestSequentialWorkflow(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()

	// Set up response sequence
	responses := []string{
		"Step 1: Analyzed the topic",
		"Step 2: Generated insights",
		"Step 3: Created summary",
	}
	responseIndex := 0

	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		if responseIndex >= len(responses) {
			return ldomain.Response{}, fmt.Errorf("unexpected call")
		}
		response := responses[responseIndex]
		responseIndex++
		return ldomain.Response{
			Content: response,
		}, nil
	})

	// Create agents
	analyzer := core.NewLLMAgent("analyzer", "test", core.LLMDeps{Provider: mockProvider})
	analyzer.SetSystemPrompt("Analyze the given topic")

	generator := core.NewLLMAgent("generator", "test", core.LLMDeps{Provider: mockProvider})
	generator.SetSystemPrompt("Generate insights based on analysis")

	summarizer := core.NewLLMAgent("summarizer", "test", core.LLMDeps{Provider: mockProvider})
	summarizer.SetSystemPrompt("Summarize the insights")

	// Create sequential workflow
	sequential := workflow.NewSequentialAgent("test-sequential").
		WithStopOnError(true).
		AddAgent(analyzer).
		AddAgent(generator).
		AddAgent(summarizer)

	// Execute workflow
	ctx := context.Background()
	state := domain.NewState()
	state.Set("user_input", "Analyze climate change impacts")

	result, err := sequential.Run(ctx, state)
	if err != nil {
		t.Fatalf("Sequential workflow failed: %v", err)
	}

	// Verify results
	outputVal, _ := result.Get("output")
	output, ok := outputVal.(string)
	if !ok {
		t.Fatal("Expected output to be string")
	}

	if !strings.Contains(output, "Created summary") {
		t.Errorf("Expected final output to contain summary, got: %s", output)
	}

	// Verify all steps were executed
	if responseIndex != 3 {
		t.Errorf("Expected 3 agent calls, got %d", responseIndex)
	}
}

// TestParallelWorkflow tests parallel agent execution
func TestParallelWorkflow(t *testing.T) {
	// Create mock providers for parallel execution
	mockProvider1 := provider.NewMockProvider()
	mockProvider2 := provider.NewMockProvider()
	mockProvider3 := provider.NewMockProvider()

	// Set up responses
	mockProvider1.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{
			Content: "Analysis: Climate change is real",
		}, nil
	})

	mockProvider2.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Simulate slower response
		time.Sleep(50 * time.Millisecond)
		return ldomain.Response{
			Content: "Research: 97% of scientists agree",
		}, nil
	})

	mockProvider3.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{
			Content: "Data: Temperature rising 1.1°C",
		}, nil
	})

	// Create agents
	analyst := core.NewLLMAgent("analyst", "test", core.LLMDeps{Provider: mockProvider1})
	researcher := core.NewLLMAgent("researcher", "test", core.LLMDeps{Provider: mockProvider2})
	dataScientist := core.NewLLMAgent("data-scientist", "test", core.LLMDeps{Provider: mockProvider3})

	// Create parallel workflow with merge all strategy
	parallel := workflow.NewParallelAgent("test-parallel").
		WithMaxConcurrency(3).
		WithMergeStrategy(workflow.MergeAll).
		AddAgent(analyst).
		AddAgent(researcher).
		AddAgent(dataScientist)

	// Execute workflow
	ctx := context.Background()
	state := domain.NewState()
	state.Set("user_input", "Analyze climate change")

	result, err := parallel.Run(ctx, state)
	if err != nil {
		t.Fatalf("Parallel workflow failed: %v", err)
	}

	// Verify all results are merged
	resultsVal, exists := result.Get("parallel_results")
	if !exists {
		t.Fatal("Expected parallel_results to exist")
	}

	results, ok := resultsVal.(map[string]interface{})
	if !ok {
		t.Fatal("Expected parallel_results to be map[string]interface{}")
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Verify each result
	expectedContent := []string{
		"Climate change is real",
		"97% of scientists agree",
		"Temperature rising 1.1°C",
	}

	foundCount := 0
	for agentName, r := range results {
		resultMap, ok := r.(map[string]interface{})
		if !ok {
			t.Errorf("Result for %s is not a map", agentName)
			continue
		}

		// Check for response, result, or output keys
		var output string
		if resp, ok := resultMap["response"].(string); ok {
			output = resp
		} else if res, ok := resultMap["result"].(string); ok {
			output = res
		} else {
			// LLMAgent sets both "result" and "output", but parallel workflow only extracts "response" and "result"
			// So output should be in result
			t.Logf("Agent %s result map: %+v", agentName, resultMap)
		}

		for _, expected := range expectedContent {
			if strings.Contains(output, expected) {
				foundCount++
				break
			}
		}
	}

	if foundCount != 3 {
		t.Errorf("Expected to find all 3 expected contents, found %d", foundCount)
	}
}

// TestConditionalWorkflow tests conditional branching with real agents
func TestConditionalWorkflow(t *testing.T) {
	// Create mock providers
	mockProvider1 := provider.NewMockProvider()
	mockProvider2 := provider.NewMockProvider()
	mockProvider3 := provider.NewMockProvider()

	// Track which branch was executed
	executedBranches := []string{}

	// Set up responses
	mockProvider1.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		executedBranches = append(executedBranches, "urgent")
		return ldomain.Response{
			Content: "Handled as urgent request",
		}, nil
	})

	mockProvider2.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		executedBranches = append(executedBranches, "normal")
		return ldomain.Response{
			Content: "Handled as normal request",
		}, nil
	})

	mockProvider3.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		executedBranches = append(executedBranches, "default")
		return ldomain.Response{
			Content: "Handled as default request",
		}, nil
	})

	// Create agents
	urgentHandler := core.NewLLMAgent("urgent-handler", "test", core.LLMDeps{Provider: mockProvider1})
	normalHandler := core.NewLLMAgent("normal-handler", "test", core.LLMDeps{Provider: mockProvider2})
	defaultHandler := core.NewLLMAgent("default-handler", "test", core.LLMDeps{Provider: mockProvider3})

	// Create conditional workflow
	conditional := workflow.NewConditionalAgent("test-conditional").
		AddAgent("urgent", func(state *domain.State) bool {
			priorityVal, _ := state.Get("priority")
			priority, _ := priorityVal.(string)
			return priority == "urgent"
		}, urgentHandler).
		AddAgent("normal", func(state *domain.State) bool {
			priorityVal, _ := state.Get("priority")
			priority, _ := priorityVal.(string)
			return priority == "normal"
		}, normalHandler).
		SetDefaultAgent(defaultHandler)

	// Test urgent priority
	ctx := context.Background()
	state := domain.NewState()
	state.Set("priority", "urgent")
	state.Set("user_input", "Handle this request")

	result, err := conditional.Run(ctx, state)
	if err != nil {
		t.Fatalf("Conditional workflow failed: %v", err)
	}

	outputVal, _ := result.Get("output")
	output, _ := outputVal.(string)
	if !strings.Contains(output, "urgent") {
		t.Errorf("Expected urgent handler, got: %s", output)
	}

	if len(executedBranches) != 1 || executedBranches[0] != "urgent" {
		t.Errorf("Expected only urgent branch to execute, got: %v", executedBranches)
	}
}

// TestLoopWorkflow tests iterative execution
func TestLoopWorkflow(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()
	iterationCount := 0

	// Set up responses that simulate quality improvement
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		iterationCount++
		quality := 0.3 + float64(iterationCount)*0.2
		return ldomain.Response{
			Content: fmt.Sprintf("Iteration %d: Improved quality to %.1f", iterationCount, quality),
		}, nil
	})

	// Create refinement agent
	refiner := core.NewLLMAgent("refiner", "test", core.LLMDeps{Provider: mockProvider})
	refiner.SetSystemPrompt("Improve the quality of the work")

	// Create a custom agent that updates quality in state
	qualityUpdater := &qualityAgent{
		BaseAgentImpl: core.NewBaseAgent("quality-updater", "Updates quality metric", domain.AgentTypeCustom),
		refiner:       refiner,
	}

	// Create loop workflow that runs until quality > 0.8
	loop := workflow.NewLoopAgent("test-loop").
		WithUntilCondition(func(state *domain.State, iteration int) bool {
			qualityVal, _ := state.Get("quality")
			quality, _ := qualityVal.(float64)
			return quality > 0.8
		}).
		WithMaxIterations(5).
		SetLoopAgent(qualityUpdater)

	// Execute workflow
	ctx := context.Background()
	state := domain.NewState()
	state.Set("quality", 0.0)
	state.Set("user_input", "Improve this work")

	result, err := loop.Run(ctx, state)
	if err != nil {
		t.Fatalf("Loop workflow failed: %v", err)
	}

	// Verify final quality
	qualityVal, _ := result.Get("quality")
	quality, _ := qualityVal.(float64)
	if quality <= 0.8 {
		t.Errorf("Expected quality > 0.8, got %.1f", quality)
	}

	// Verify iteration count (should be 3: 0.5, 0.7, 0.9)
	if iterationCount != 3 {
		t.Errorf("Expected 3 iterations, got %d", iterationCount)
	}
}

// qualityAgent is a test agent that updates quality metric
type qualityAgent struct {
	*core.BaseAgentImpl
	refiner domain.BaseAgent
}

func (q *qualityAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Run the refiner
	result, err := q.refiner.Run(ctx, state)
	if err != nil {
		return nil, err
	}

	// Update quality based on iteration
	qualityVal, _ := state.Get("quality")
	quality, _ := qualityVal.(float64)
	if quality == 0 {
		quality = 0.5
	} else {
		quality += 0.2
	}

	result.Set("quality", quality)
	return result, nil
}
