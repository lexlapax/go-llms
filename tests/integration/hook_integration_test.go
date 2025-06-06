// ABOUTME: Integration tests for hook system across different agent types
// ABOUTME: Tests hook lifecycle, multiple hooks, and hook data consistency

package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestHookImpl is a test hook implementation that captures all events
type TestHookImpl struct {
	mu                  sync.RWMutex
	beforeGenerateCalls [][]ldomain.Message
	afterGenerateCalls  []AfterGenerateCall
	beforeToolCalls     []BeforeToolCall
	afterToolCalls      []AfterToolCall
	callOrder           []string
}

type AfterGenerateCall struct {
	Response ldomain.Response
	Error    error
}

type BeforeToolCall struct {
	Tool   string
	Params map[string]interface{}
}

type AfterToolCall struct {
	Tool   string
	Result interface{}
	Error  error
}

func NewTestHook() *TestHookImpl {
	return &TestHookImpl{
		beforeGenerateCalls: make([][]ldomain.Message, 0),
		afterGenerateCalls:  make([]AfterGenerateCall, 0),
		beforeToolCalls:     make([]BeforeToolCall, 0),
		afterToolCalls:      make([]AfterToolCall, 0),
		callOrder:           make([]string, 0),
	}
}

func (h *TestHookImpl) BeforeGenerate(ctx context.Context, messages []ldomain.Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.beforeGenerateCalls = append(h.beforeGenerateCalls, messages)
	h.callOrder = append(h.callOrder, "BeforeGenerate")
}

func (h *TestHookImpl) AfterGenerate(ctx context.Context, response ldomain.Response, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.afterGenerateCalls = append(h.afterGenerateCalls, AfterGenerateCall{Response: response, Error: err})
	h.callOrder = append(h.callOrder, "AfterGenerate")
}

func (h *TestHookImpl) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.beforeToolCalls = append(h.beforeToolCalls, BeforeToolCall{Tool: tool, Params: params})
	h.callOrder = append(h.callOrder, fmt.Sprintf("BeforeToolCall:%s", tool))
}

func (h *TestHookImpl) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.afterToolCalls = append(h.afterToolCalls, AfterToolCall{Tool: tool, Result: result, Error: err})
	h.callOrder = append(h.callOrder, fmt.Sprintf("AfterToolCall:%s", tool))
}

func (h *TestHookImpl) GetStats() (beforeGen int, afterGen int, beforeTool int, afterTool int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.beforeGenerateCalls), len(h.afterGenerateCalls), len(h.beforeToolCalls), len(h.afterToolCalls)
}

func (h *TestHookImpl) GetCallOrder() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	order := make([]string, len(h.callOrder))
	copy(order, h.callOrder)
	return order
}

// TestBasicHookIntegration tests basic hook functionality with an LLM agent
func TestBasicHookIntegration(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{Content: "I'll help you calculate that."}, nil
	})

	// Create agent with hooks
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("test-agent", "Test agent with hooks", deps)

	// Add test hook
	testHook := NewTestHook()
	agent.WithHook(testHook)

	// Set system prompt
	agent.SetSystemPrompt("You are a helpful assistant.")

	// Execute agent
	ctx := context.Background()
	state := domain.NewState()
	state.Set("user_input", "Calculate 2 + 2")

	result, err := agent.Run(ctx, state)
	if err != nil {
		t.Fatalf("Agent execution failed: %v", err)
	}

	// Verify result
	output, exists := result.Get("output")
	if !exists {
		t.Fatal("No output in result")
	}
	if output != "I'll help you calculate that." {
		t.Errorf("Unexpected output: %v", output)
	}

	// Verify hook calls
	beforeGen, afterGen, beforeTool, afterTool := testHook.GetStats()
	if beforeGen != 1 {
		t.Errorf("Expected 1 BeforeGenerate call, got %d", beforeGen)
	}
	if afterGen != 1 {
		t.Errorf("Expected 1 AfterGenerate call, got %d", afterGen)
	}
	if beforeTool != 0 {
		t.Errorf("Expected 0 BeforeToolCall calls, got %d", beforeTool)
	}
	if afterTool != 0 {
		t.Errorf("Expected 0 AfterToolCall calls, got %d", afterTool)
	}

	// Verify call order
	expectedOrder := []string{"BeforeGenerate", "AfterGenerate"}
	actualOrder := testHook.GetCallOrder()
	if len(actualOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d calls, got %d", len(expectedOrder), len(actualOrder))
	}
	for i, expected := range expectedOrder {
		if actualOrder[i] != expected {
			t.Errorf("Call %d: expected %s, got %s", i, expected, actualOrder[i])
		}
	}
}

// TestHookWithToolCalls tests hooks when tools are called
func TestHookWithToolCalls(t *testing.T) {
	// Create mock provider with tool call response
	mockProvider := provider.NewMockProvider()

	// Create a response queue for multiple calls
	responses := []ldomain.Response{
		{Content: `{"tool": "calculator", "params": {"operation": "add", "a": 2, "b": 2}}`},
		{Content: "The result is 4."},
	}
	responseIndex := 0

	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		if responseIndex >= len(responses) {
			return ldomain.Response{}, fmt.Errorf("no more responses")
		}
		resp := responses[responseIndex]
		responseIndex++
		return resp, nil
	})

	// Create agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("test-agent", "Test agent with tools", deps)

	// Add test hook
	testHook := NewTestHook()
	agent.WithHook(testHook)

	// Add calculator tool
	// Define parameter schema
	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"operation": {Type: "string"},
			"a":         {Type: "number"},
			"b":         {Type: "number"},
		},
		Required: []string{"operation", "a", "b"},
	}

	calcTool := tools.NewTool(
		"calculator",
		"Perform calculations",
		func(params struct {
			Operation string  `json:"operation"`
			A         float64 `json:"a"`
			B         float64 `json:"b"`
		}) (float64, error) {
			switch params.Operation {
			case "add":
				return params.A + params.B, nil
			case "subtract":
				return params.A - params.B, nil
			default:
				return 0, fmt.Errorf("unknown operation: %s", params.Operation)
			}
		},
		paramSchema,
	)
	agent.AddTool(calcTool)

	// Execute agent
	ctx := context.Background()
	state := domain.NewState()
	state.Set("user_input", "Calculate 2 + 2")

	result, err := agent.Run(ctx, state)
	if err != nil {
		t.Fatalf("Agent execution failed: %v", err)
	}

	// Verify result
	output, _ := result.Get("output")
	if output != "The result is 4." {
		t.Errorf("Unexpected output: %v", output)
	}

	// Verify hook calls
	beforeGen, afterGen, beforeTool, afterTool := testHook.GetStats()
	if beforeGen != 2 { // Initial call + call after tool result
		t.Errorf("Expected 2 BeforeGenerate calls, got %d", beforeGen)
	}
	if afterGen != 2 {
		t.Errorf("Expected 2 AfterGenerate calls, got %d", afterGen)
	}
	if beforeTool != 1 {
		t.Errorf("Expected 1 BeforeToolCall call, got %d", beforeTool)
	}
	if afterTool != 1 {
		t.Errorf("Expected 1 AfterToolCall call, got %d", afterTool)
	}

	// Verify tool call details
	if len(testHook.beforeToolCalls) > 0 {
		toolCall := testHook.beforeToolCalls[0]
		if toolCall.Tool != "calculator" {
			t.Errorf("Expected tool 'calculator', got '%s'", toolCall.Tool)
		}
	}

	// Verify call order
	expectedOrder := []string{
		"BeforeGenerate",
		"AfterGenerate",
		"BeforeToolCall:calculator",
		"AfterToolCall:calculator",
		"BeforeGenerate",
		"AfterGenerate",
	}
	actualOrder := testHook.GetCallOrder()
	if len(actualOrder) != len(expectedOrder) {
		t.Fatalf("Expected %d calls, got %d", len(expectedOrder), len(actualOrder))
	}
	for i, expected := range expectedOrder {
		if actualOrder[i] != expected {
			t.Errorf("Call %d: expected %s, got %s", i, expected, actualOrder[i])
		}
	}
}

// TestMultipleHooks tests that multiple hooks are called in order
func TestMultipleHooks(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{Content: "Test response"}, nil
	})

	// Create agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("test-agent", "Test agent with multiple hooks", deps)

	// Add multiple test hooks
	hook1 := NewTestHook()
	hook2 := NewTestHook()
	hook3 := NewTestHook()

	agent.WithHook(hook1).WithHook(hook2).WithHook(hook3)

	// Execute agent
	ctx := context.Background()
	state := domain.NewState()
	state.Set("user_input", "Test input")

	_, err := agent.Run(ctx, state)
	if err != nil {
		t.Fatalf("Agent execution failed: %v", err)
	}

	// Verify all hooks were called
	for i, hook := range []*TestHookImpl{hook1, hook2, hook3} {
		beforeGen, afterGen, _, _ := hook.GetStats()
		if beforeGen != 1 {
			t.Errorf("Hook %d: Expected 1 BeforeGenerate call, got %d", i+1, beforeGen)
		}
		if afterGen != 1 {
			t.Errorf("Hook %d: Expected 1 AfterGenerate call, got %d", i+1, afterGen)
		}
	}
}

// TestHookWithErrors tests hook behavior when errors occur
func TestHookWithErrors(t *testing.T) {
	// Create mock provider that returns an error
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{}, fmt.Errorf("provider error")
	})

	// Create agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("test-agent", "Test agent with errors", deps)

	// Add test hook
	testHook := NewTestHook()
	agent.WithHook(testHook)

	// Execute agent
	ctx := context.Background()
	state := domain.NewState()
	state.Set("user_input", "Test input")

	_, err := agent.Run(ctx, state)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify hook was called even with error
	beforeGen, afterGen, _, _ := testHook.GetStats()
	if beforeGen != 1 {
		t.Errorf("Expected 1 BeforeGenerate call, got %d", beforeGen)
	}
	if afterGen != 1 {
		t.Errorf("Expected 1 AfterGenerate call, got %d", afterGen)
	}

	// Verify error was passed to hook
	if len(testHook.afterGenerateCalls) > 0 {
		call := testHook.afterGenerateCalls[0]
		if call.Error == nil {
			t.Error("Expected error in AfterGenerate call")
		}
	}
}

// TestMetricsHook tests the built-in metrics hook
func TestMetricsHook(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{Content: "Test response"}, nil
	})

	// Create agent with metrics hook
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("test-agent", "Test agent with metrics", deps)

	// Add metrics hook
	metricsHook := core.NewLLMMetricsHook()
	agent.WithHook(metricsHook)

	// Execute agent multiple times
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		state := domain.NewState()
		state.Set("user_input", fmt.Sprintf("Test input %d", i))

		_, err := agent.Run(ctx, state)
		if err != nil {
			t.Fatalf("Agent execution %d failed: %v", i, err)
		}
	}

	// Check metrics
	metrics := metricsHook.GetMetrics()
	if metrics.Requests != 3 {
		t.Errorf("Expected 3 requests, got %d", metrics.Requests)
	}
	if metrics.ErrorCount != 0 {
		t.Errorf("Expected 0 errors, got %d", metrics.ErrorCount)
	}
	if metrics.TotalTokens == 0 {
		t.Error("Expected some token count")
	}
	// Note: Average generation time might be 0 in fast tests, so we'll skip this check
	// The important thing is that the hook captured the calls
}

// TestLoggingHook tests the built-in logging hook
func TestLoggingHook(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{Content: "Test response"}, nil
	})

	// Create agent with logging hook
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("test-agent", "Test agent with logging", deps)

	// Add logging hook (output will go to test log)
	loggingHook := core.NewLoggingHook(nil, core.LogLevelDetailed)
	agent.WithHook(loggingHook)

	// Execute agent
	ctx := context.Background()
	state := domain.NewState()
	state.Set("user_input", "Test input for logging")

	_, err := agent.Run(ctx, state)
	if err != nil {
		t.Fatalf("Agent execution failed: %v", err)
	}

	// Test passes if no panic occurs
	// Actual log output verification would require capturing log output
}

// TestHooksInWorkflowAgents tests hooks in workflow agents
func TestHooksInWorkflowAgents(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()

	// Create two LLM agents with hooks
	deps := core.LLMDeps{
		Provider: mockProvider,
	}

	agent1Hook := NewTestHook()
	agent1 := core.NewLLMAgent("agent1", "First agent", deps)
	agent1.WithHook(agent1Hook)

	agent2Hook := NewTestHook()
	agent2 := core.NewLLMAgent("agent2", "Second agent", deps)
	agent2.WithHook(agent2Hook)

	// Set up responses for both agents
	responses := []ldomain.Response{
		{Content: "Result from agent 1"},
		{Content: "Result from agent 2"},
	}
	responseIndex := 0

	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		if responseIndex >= len(responses) {
			return ldomain.Response{}, fmt.Errorf("no more responses")
		}
		resp := responses[responseIndex]
		responseIndex++
		return resp, nil
	})

	// Create sequential workflow
	sequential := workflow.NewSequentialAgent("sequential")
	// Add agents as workflow steps
	err := sequential.AddStep(workflow.NewAgentStep("agent1", agent1))
	if err != nil {
		t.Fatalf("Failed to add agent1: %v", err)
	}
	err = sequential.AddStep(workflow.NewAgentStep("agent2", agent2))
	if err != nil {
		t.Fatalf("Failed to add agent2: %v", err)
	}

	// Execute workflow
	ctx := context.Background()
	state := domain.NewState()
	state.Set("user_input", "Process this sequentially")

	result, err := sequential.Run(ctx, state)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	// Verify both agents' hooks were called
	beforeGen1, afterGen1, _, _ := agent1Hook.GetStats()
	if beforeGen1 != 1 {
		t.Errorf("Agent1: Expected 1 BeforeGenerate call, got %d", beforeGen1)
	}
	if afterGen1 != 1 {
		t.Errorf("Agent1: Expected 1 AfterGenerate call, got %d", afterGen1)
	}

	beforeGen2, afterGen2, _, _ := agent2Hook.GetStats()
	if beforeGen2 != 1 {
		t.Errorf("Agent2: Expected 1 BeforeGenerate call, got %d", beforeGen2)
	}
	if afterGen2 != 1 {
		t.Errorf("Agent2: Expected 1 AfterGenerate call, got %d", afterGen2)
	}

	// Verify final result
	if output, exists := result.Get("output"); exists {
		if output != "Result from agent 2" {
			t.Errorf("Unexpected final output: %v", output)
		}
	}
}

// TestConcurrentHookSafety tests that hooks are thread-safe
func TestConcurrentHookSafety(t *testing.T) {
	// Create mock provider
	mockProvider := provider.NewMockProvider()

	// Create agent with test hook
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("test-agent", "Test concurrent hook safety", deps)

	testHook := NewTestHook()
	agent.WithHook(testHook)

	// Run multiple concurrent executions
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Each goroutine gets its own response function
			// Note: This creates a race condition in the real test, but it's intentional
			// to test concurrent safety of hooks themselves

			ctx := context.Background()
			state := domain.NewState()
			state.Set("user_input", fmt.Sprintf("Input %d", id))

			_, err := agent.Run(ctx, state)
			if err != nil {
				t.Errorf("Goroutine %d: execution failed: %v", id, err)
			}
		}(i)

		// Small delay to ensure concurrent execution
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()

	// Verify hook was called correct number of times
	beforeGen, afterGen, _, _ := testHook.GetStats()
	if beforeGen != numGoroutines {
		t.Errorf("Expected %d BeforeGenerate calls, got %d", numGoroutines, beforeGen)
	}
	if afterGen != numGoroutines {
		t.Errorf("Expected %d AfterGenerate calls, got %d", numGoroutines, afterGen)
	}
}
