package examples

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// TestMultiAgent_InitAndRun tests the initialization and basic running of a MultiAgent
func TestMultiAgent_InitAndRun(t *testing.T) {
	// Create a mock provider with custom response behavior
	mockProvider := &CustomMockProvider{
		generateMessageFunc: func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
			if len(messages) > 0 && messages[len(messages)-1].Role == ldomain.RoleUser {
				// Simple response without tool calls
				return ldomain.Response{Content: "This is a response to: " + messages[len(messages)-1].Content}, nil
			}
			return ldomain.Response{Content: "Default response"}, nil
		},
	}

	// Create a multi agent
	agent := workflow.NewMultiAgent(mockProvider)
	agent.SetSystemPrompt("You are a helpful assistant.")

	// Test basic functionality
	response, err := agent.Run(context.Background(), "Hello, world!")
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	// Check the response
	expectedPart := "This is a response to: Hello, world!"
	if response != expectedPart {
		t.Errorf("Expected response to contain %q, got %q", expectedPart, response)
	}
}

// TestMultiAgent_WithTools tests the MultiAgent with tool usage
func TestMultiAgent_WithTools(t *testing.T) {
	// Create a mock provider that uses tools
	mockProvider := &CustomMockProvider{
		generateMessageFunc: func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
			// Check if this is the first response or a follow-up
			if len(messages) == 2 { // system + user
				// First response - call the calculator tool
				return ldomain.Response{Content: `{"tool": "calculator", "params": {"expression": "2+2"}}`}, nil
			} else {
				// Follow-up response after tool execution
				return ldomain.Response{Content: "The result of 2+2 is 4"}, nil
			}
		},
	}

	// Create a multi agent
	agent := workflow.NewMultiAgent(mockProvider)
	agent.SetSystemPrompt("You are a helpful assistant.")

	// Add calculator tool
	agent.AddTool(CreateCalculatorTool())

	// Run the agent
	response, err := agent.Run(context.Background(), "What is 2+2?")
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	// Check the response
	expected := "The result of 2+2 is 4"
	if response != expected {
		t.Errorf("Expected %q, got %q", expected, response)
	}
}

// TestCachedAgent_BasicCaching tests the CachedAgent's caching capabilities
func TestCachedAgent_BasicCaching(t *testing.T) {
	// Counter to track how many times the mock provider is called
	callCount := 0

	// Create a mock provider
	mockProvider := &CustomMockProvider{
		generateMessageFunc: func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
			callCount++
			return ldomain.Response{Content: "Response #" + string(rune(callCount+'0'))}, nil
		},
	}

	// Create a cached agent
	agent := workflow.NewCachedAgent(mockProvider)
	agent.SetSystemPrompt("You are a helpful assistant.")

	// First call - should go to the provider
	response1, err := agent.Run(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	// Second call with the same input - should use cache
	response2, err := agent.Run(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	// Different input - should go to the provider
	response3, err := agent.Run(context.Background(), "Different")
	if err != nil {
		t.Fatalf("Third call failed: %v", err)
	}

	// Check responses
	if response1 != response2 {
		t.Errorf("Cache not working: Second response should match first, got %q vs %q", response1, response2)
	}

	if response2 == response3 {
		t.Errorf("Cache issue: Third response should be different, got %q for both", response2)
	}

	// Check call count - should be 2, not 3
	if callCount != 2 {
		t.Errorf("Expected 2 provider calls, got %d", callCount)
	}

	// Check cache stats
	stats := agent.GetCacheStats()
	if stats["hits"].(int) < 1 {
		t.Errorf("Expected at least 1 cache hit, got %d", stats["hits"].(int))
	}
}

// TestMessageManager tests the message manager's functionality
func TestMessageManager_Truncation(t *testing.T) {
	// Create a message manager with a small message limit
	config := workflow.MessageManagerConfig{
		UseTokenTruncation:    false, // Use simple message count truncation
		KeepAllSystemMessages: true,
	}
	manager := workflow.NewMessageManager(3, 1000, config)

	// Add a system message
	manager.SetSystemPrompt("System prompt")

	// Add two user messages
	manager.AddMessage(ldomain.Message{
		Role:    ldomain.RoleUser,
		Content: "User message 1",
	})
	manager.AddMessage(ldomain.Message{
		Role:    ldomain.RoleUser,
		Content: "User message 2",
	})

	// Get messages - should have all 3
	messages := manager.GetMessages()
	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	// Add one more message to trigger truncation
	manager.AddMessage(ldomain.Message{
		Role:    ldomain.RoleAssistant,
		Content: "Assistant response",
	})

	// Get messages again - should still have 3, but the oldest user message should be gone
	messages = manager.GetMessages()
	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages after truncation, got %d", len(messages))
	}

	// Check if system message is retained
	var hasSystemMsg, hasUserMsg2, hasAssistantMsg bool
	for _, msg := range messages {
		if msg.Role == ldomain.RoleSystem {
			hasSystemMsg = true
		}
		if msg.Role == ldomain.RoleUser && msg.Content == "User message 2" {
			hasUserMsg2 = true
		}
		if msg.Role == ldomain.RoleAssistant {
			hasAssistantMsg = true
		}
	}
	
	if !hasSystemMsg {
		t.Errorf("Expected to find system message after truncation")
	}
	if !hasUserMsg2 {
		t.Errorf("Expected to find 'User message 2' after truncation")
	}
	if !hasAssistantMsg {
		t.Errorf("Expected to find assistant message after truncation")
	}
}

// TestToolExecutor tests the parallel tool execution functionality
func TestToolExecutor_ParallelExecution(t *testing.T) {
	// Create a map of tools
	toolMap := make(map[string]domain.Tool)

	// Add a fast tool using our mock tool implementation
	toolMap["fast"] = MockTool{
		name:        "fast",
		description: "Fast tool",
		executor: func(ctx context.Context, params interface{}) (interface{}, error) {
			return "Fast result", nil
		},
	}

	// Add a slow tool
	toolMap["slow"] = MockTool{
		name:        "slow",
		description: "Slow tool",
		executor: func(ctx context.Context, params interface{}) (interface{}, error) {
			time.Sleep(100 * time.Millisecond)
			return "Slow result", nil
		},
	}

	// Create a tool executor with concurrency limit
	executor := workflow.NewToolExecutor(toolMap, 2, 1*time.Second, nil)

	// Execute tools in parallel
	toolNames := []string{"fast", "slow"}
	params := []interface{}{nil, nil}

	startTime := time.Now()
	results := executor.ExecuteToolsParallel(context.Background(), toolNames, params)
	duration := time.Since(startTime)

	// Verify we got both results
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Check individual results
	fastResult, ok := results["fast"]
	if !ok || fastResult.Status != workflow.ToolStatusSuccess {
		t.Errorf("Fast tool execution failed or not found")
	}

	slowResult, ok := results["slow"]
	if !ok || slowResult.Status != workflow.ToolStatusSuccess {
		t.Errorf("Slow tool execution failed or not found")
	}

	// Execution should take approximately the time of the slowest tool (~100ms)
	// But allow some margin for test environment variability
	if duration < 90*time.Millisecond || duration > 200*time.Millisecond {
		t.Errorf("Parallel execution time unexpected: %v (expected ~100ms)", duration)
	}
}

// TestToolExecutor_Timeout tests that the tool executor respects timeouts
func TestToolExecutor_Timeout(t *testing.T) {
	// Create a map with a very slow tool that respects context cancellation
	toolMap := make(map[string]domain.Tool)
	toolMap["very_slow"] = MockTool{
		name:        "very_slow",
		description: "Very slow tool",
		executor: func(ctx context.Context, params interface{}) (interface{}, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(300 * time.Millisecond):
				return "Completed execution", nil
			}
		},
	}

	// Use a mutex to protect concurrent access to status flag
	var mu sync.Mutex
	var resultStatus workflow.ToolExecutionStatus
	
	// Create executor with a very short timeout
	// Using a reasonable timeout that's still short enough to trigger but not race-prone
	executor := workflow.NewToolExecutor(toolMap, 1, 50*time.Millisecond, nil)

	// Execute the slow tool in a separate goroutine to avoid blocking the test
	done := make(chan bool)
	go func() {
		results := executor.ExecuteToolsParallel(context.Background(), []string{"very_slow"}, []interface{}{nil})

		// Store the result status safely
		mu.Lock()
		if result, ok := results["very_slow"]; ok {
			resultStatus = result.Status
		}
		mu.Unlock()
		
		done <- true
	}()
	
	// Wait for the execution to complete with a generous timeout
	select {
	case <-done:
		// Success, proceed to check results
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Test execution timed out")
	}

	// Check the result status - should be either timeout or error depending on exact timing
	mu.Lock()
	defer mu.Unlock()
	
	// Accept either timeout or error, as timing can be unpredictable across environments
	if resultStatus != workflow.ToolStatusTimeout && resultStatus != workflow.ToolStatusError {
		t.Errorf("Expected tool status to be timeout or error, got: %s", resultStatus)
	}
}

// The common mock helpers are now in mock_helpers.go