package examples

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestAgentExample tests the basic agent example
func TestAgentExample(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	agent := workflow.NewAgent(mockProvider)
	
	// Add a calculator tool
	agent.AddTool(tools.NewTool(
		"calculator",
		"Perform mathematical calculations",
		func(params struct {
			Expression string `json:"expression"`
		}) (map[string]interface{}, error) {
			// Simple mock calculator that returns 4 for every expression
			return map[string]interface{}{
				"result": 4,
			}, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate",
				},
			},
			Required: []string{"expression"},
		},
	))
	
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

// TestMultiAgentExample tests the multi-agent capabilities
func TestMultiAgentExample(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	agent := workflow.NewMultiAgent(mockProvider)
	
	// Add a calculator tool
	agent.AddTool(tools.NewTool(
		"calculator",
		"Perform mathematical calculations",
		func(params struct {
			Expression string `json:"expression"`
		}) (map[string]interface{}, error) {
			// Simple mock calculator that returns 4 for every expression
			return map[string]interface{}{
				"result": 4,
			}, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate",
				},
			},
			Required: []string{"expression"},
		},
	))
	
	// Add a date tool
	agent.AddTool(tools.NewTool(
		"date",
		"Get the current date",
		func() map[string]string {
			return map[string]string{
				"date": time.Now().Format("2006-01-02"),
			}
		},
		nil,
	))
	
	// Set a system prompt
	agent.SetSystemPrompt("You are a helpful assistant.")
	
	// Run the agent with a prompt that requires multiple tools
	result, err := agent.Run(context.Background(), "What is today's date? Also, calculate 2+2.")
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
	agent.AddTool(tools.NewTool(
		"calculator",
		"Perform mathematical calculations",
		func(params struct {
			Expression string `json:"expression"`
		}) (map[string]interface{}, error) {
			// Simple mock calculator that returns 4 for every expression
			return map[string]interface{}{
				"result": 4,
			}, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate",
				},
			},
			Required: []string{"expression"},
		},
	))
	
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

// TestAgentWithSchema tests running an agent with schema validation
func TestAgentWithSchema(t *testing.T) {
	// Create a mock provider that returns structured data
	mockProvider := &MockStructuredProvider{
		data: map[string]interface{}{
			"name": "John Doe",
			"age":  30,
		},
	}
	
	agent := workflow.NewAgent(mockProvider)
	
	// Define a schema for person data
	personSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
		Required: []string{"name", "age"},
	}
	
	// Run the agent with schema
	result, err := agent.RunWithSchema(context.Background(), "Give me information about a person", personSchema)
	if err != nil {
		t.Fatalf("Agent failed to run with schema: %v", err)
	}
	
	// Check the result
	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map result, got %T", result)
	}
	
	if data["name"] != "John Doe" {
		t.Errorf("Expected name to be 'John Doe', got %v", data["name"])
	}
	
	if data["age"] != 30 {
		t.Errorf("Expected age to be 30, got %v", data["age"])
	}
}

// TestMessageManagerExample tests the message manager functionality
func TestMessageManagerExample(t *testing.T) {
	// Create a message manager with modest limits
	config := workflow.MessageManagerConfig{
		UseTokenTruncation:   true,
		KeepAllSystemMessages: true,
	}
	manager := workflow.NewMessageManager(5, 500, config)
	
	// Add a system message
	manager.SetSystemPrompt("You are a helpful assistant.")
	
	// Add some user and assistant messages
	messages := []ldomain.Message{
		{Role: ldomain.RoleUser, Content: "Hello!"},
		{Role: ldomain.RoleAssistant, Content: "Hi there! How can I help you today?"},
		{Role: ldomain.RoleUser, Content: "What's the weather like?"},
		{Role: ldomain.RoleAssistant, Content: "I'm an AI assistant and don't have access to real-time weather information. To get the current weather, please check a weather service or app."},
	}
	
	manager.AddMessages(messages)
	
	// Check that we have the expected number of messages (system + 4 messages)
	if manager.GetMessageCount() != 5 {
		t.Errorf("Expected 5 messages, got %d", manager.GetMessageCount())
	}
	
	// Add one more message to trigger truncation
	manager.AddMessage(ldomain.Message{
		Role:    ldomain.RoleUser,
		Content: "Can you help me with a math problem?",
	})
	
	// Check that we have the right messages after adding the new one
	currentMsgs := manager.GetMessages()
	t.Logf("Got %d messages after adding new message", len(currentMsgs))
	
	// The first message should be the system message
	if currentMsgs[0].Role != ldomain.RoleSystem {
		t.Errorf("Expected first message to be system, got %s", currentMsgs[0].Role)
	}
}

// TestToolExecutorExample tests the tool executor functionality
func TestToolExecutorExample(t *testing.T) {
	// Create a set of tools
	toolMap := make(map[string]domain.Tool)
	
	// Create a calculator tool
	calculatorTool := MockTool{
		name:        "calculator",
		description: "Perform mathematical calculations",
		executor: func(ctx context.Context, params interface{}) (interface{}, error) {
			// Simple mock calculator that returns 4 for every expression
			return map[string]interface{}{
				"result": 4,
			}, nil
		},
		schema: &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate",
				},
			},
			Required: []string{"expression"},
		},
	}
	toolMap["calculator"] = calculatorTool
	
	// Create a date tool
	dateTool := MockTool{
		name:        "date",
		description: "Get the current date",
		executor: func(ctx context.Context, params interface{}) (interface{}, error) {
			return map[string]string{
				"date": time.Now().Format("2006-01-02"),
			}, nil
		},
	}
	toolMap["date"] = dateTool
	
	// Create a tool executor
	executor := workflow.NewToolExecutor(toolMap, 2, 1*time.Second, nil)
	
	// Execute multiple tools in parallel
	toolNames := []string{"calculator", "date", "unknown_tool"}
	params := []interface{}{
		struct{ Expression string }{Expression: "2+2"},
		nil,
		nil,
	}
	
	results := executor.ExecuteToolsParallel(context.Background(), toolNames, params)
	
	// Check we have results for all tools
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	
	// Calculator tool should have succeeded
	calcResult, exists := results["calculator"]
	if !exists {
		t.Fatalf("Expected calculator result, but not found")
	}
	if calcResult.Status != workflow.ToolStatusSuccess {
		t.Errorf("Expected calculator success, got %s", calcResult.Status)
	}
	
	// Date tool should have succeeded
	dateResult, exists := results["date"]
	if !exists {
		t.Fatalf("Expected date result, but not found")
	}
	if dateResult.Status != workflow.ToolStatusSuccess {
		t.Errorf("Expected date success, got %s", dateResult.Status)
	}
	
	// Unknown tool should have failed
	unknownResult, exists := results["unknown_tool"]
	if !exists {
		t.Fatalf("Expected unknown_tool result, but not found")
	}
	if unknownResult.Status != workflow.ToolStatusNotFound {
		t.Errorf("Expected unknown_tool to be not found, got %s", unknownResult.Status)
	}
}

// The common mock helpers are now in mock_helpers.go