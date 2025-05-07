package workflow

import (
	"context"
	"fmt"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// MockProvider is a mock implementation of the Provider interface
type MockProvider struct {
	generateMessageFunc    func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error)
	generateWithSchemaFunc func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error)
}

// Generate produces text from a prompt
func (p *MockProvider) Generate(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
	resp, err := p.GenerateMessage(ctx, []ldomain.Message{{Role: ldomain.RoleUser, Content: prompt}}, options...)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// GenerateMessage generates a response to a sequence of messages
func (p *MockProvider) GenerateMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
	if p.generateMessageFunc != nil {
		return p.generateMessageFunc(ctx, messages, options...)
	}
	return ldomain.Response{Content: "Mock response"}, nil
}

// GenerateWithSchema produces structured output conforming to a schema
func (p *MockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
	if p.generateWithSchemaFunc != nil {
		return p.generateWithSchemaFunc(ctx, prompt, schema, options...)
	}
	return map[string]interface{}{"result": "Mock structured response"}, nil
}

// Stream streams responses token by token
func (p *MockProvider) Stream(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	ch := make(chan ldomain.Token)
	go func() {
		defer close(ch)
		ch <- ldomain.Token{Text: "Mock", Finished: false}
		ch <- ldomain.Token{Text: " response", Finished: true}
	}()
	return ch, nil
}

// StreamMessage streams responses token by token with messages
func (p *MockProvider) StreamMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	ch := make(chan ldomain.Token)
	go func() {
		defer close(ch)
		ch <- ldomain.Token{Text: "Mock", Finished: false}
		ch <- ldomain.Token{Text: " response", Finished: true}
	}()
	return ch, nil
}

// TestDefaultAgent_Run tests the Run method of DefaultAgent
func TestDefaultAgent_Run(t *testing.T) {
	t.Run("simple response without tools", func(t *testing.T) {
		mockProvider := &MockProvider{
			generateMessageFunc: func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
				return ldomain.Response{Content: "This is a simple response without tool calls"}, nil
			},
		}

		agent := NewAgent(mockProvider)

		result, err := agent.Run(context.Background(), "Test query")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// The result should be the simple response
		expectedResult := "This is a simple response without tool calls"
		if result != expectedResult {
			t.Errorf("Expected result to be '%s', got '%v'", expectedResult, result)
		}
	})

	t.Run("with tool call", func(t *testing.T) {
		// First response calls a tool, second response is final
		callCount := 0
		mockProvider := &MockProvider{
			generateMessageFunc: func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
				callCount++
				if callCount == 1 {
					return ldomain.Response{Content: "I'll need to calculate something.\n\n```json\n{\"tool\": \"calculator\", \"params\": {\"expression\": \"2+2\"}}\n```"}, nil
				}
				return ldomain.Response{Content: "The result of 2+2 is 4"}, nil
			},
		}

		agent := NewAgent(mockProvider)

		// Add a calculator tool
		agent.AddTool(tools.NewTool(
			"calculator",
			"Calculates mathematical expressions",
			func(params struct {
				Expression string `json:"expression"`
			}) (int, error) {
				// Simple mock calculator that always returns 4
				return 4, nil
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

		result, err := agent.Run(context.Background(), "What is 2+2?")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// The result should be the final response after tool call
		expectedResult := "The result of 2+2 is 4"
		if result != expectedResult {
			t.Errorf("Expected result to be '%s', got '%v'", expectedResult, result)
		}

		// There should have been 2 calls to the provider
		if callCount != 2 {
			t.Errorf("Expected 2 calls to the provider, got %d", callCount)
		}
	})

	t.Run("with schema validation", func(t *testing.T) {
		schema := &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
			Required: []string{"name"},
		}

		mockProvider := &MockProvider{
			generateWithSchemaFunc: func(ctx context.Context, prompt string, s *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
				return map[string]interface{}{
					"name": "Alice",
					"age":  30,
				}, nil
			},
		}

		agent := NewAgent(mockProvider)

		result, err := agent.RunWithSchema(context.Background(), "Get me a person", schema)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// The result should be a map with name and age
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected result to be a map, got %T", result)
		}

		if resultMap["name"] != "Alice" {
			t.Errorf("Expected name to be 'Alice', got '%v'", resultMap["name"])
		}

		if resultMap["age"] != 30 {
			t.Errorf("Expected age to be 30, got '%v'", resultMap["age"])
		}
	})
}

// TestDefaultAgent_ToolCall tests the agent's tool call handling
func TestDefaultAgent_ToolCall(t *testing.T) {
	// Create an agent instance to test ExtractToolCall
	agent := &DefaultAgent{}

	// Test JSON format in message content
	tool, _, shouldCall := agent.ExtractToolCall(`{"tool": "calculator", "params": {"expression": "2+2"}}`)
	if !shouldCall {
		t.Errorf("Expected shouldCall to be true for JSON format")
	}
	if tool != "calculator" {
		t.Errorf("Expected tool to be 'calculator', got '%s'", tool)
	}

	// Test JSON in code block
	tool, _, shouldCall = agent.ExtractToolCall("I need to use the calculator.\n\n```json\n{\"tool\": \"calculator\", \"params\": {\"expression\": \"2+2\"}}\n```")
	if !shouldCall {
		t.Errorf("Expected shouldCall to be true for code block format")
	}
	if tool != "calculator" {
		t.Errorf("Expected tool to be 'calculator', got '%s'", tool)
	}

	// Test extract JSON blocks
	blocks := extractJSONBlocks("Here's some JSON:\n\n```json\n{\"tool\": \"calculator\"}\n```\n\nAnd some more:\n```\n{\"value\": 42}\n```")
	if len(blocks) != 2 {
		t.Errorf("Expected 2 JSON blocks, got %d", len(blocks))
	}
}

// MockHook is a mock implementation of the Hook interface
type MockHook struct {
	beforeGenerateCalled bool
	afterGenerateCalled  bool
	beforeToolCallCalled bool
	afterToolCallCalled  bool
}

// BeforeGenerate is called before generating a response
func (h *MockHook) BeforeGenerate(ctx context.Context, messages []ldomain.Message) {
	h.beforeGenerateCalled = true
}

// AfterGenerate is called after generating a response
func (h *MockHook) AfterGenerate(ctx context.Context, response ldomain.Response, err error) {
	h.afterGenerateCalled = true
}

// BeforeToolCall is called before executing a tool
func (h *MockHook) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	h.beforeToolCallCalled = true
}

// AfterToolCall is called after executing a tool
func (h *MockHook) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	h.afterToolCallCalled = true
}

// TestDefaultAgent_Hooks tests hooks in the agent
func TestDefaultAgent_Hooks(t *testing.T) {
	mockProvider := &MockProvider{
		generateMessageFunc: func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
			return ldomain.Response{Content: "{\"tool\": \"test_tool\", \"params\": {\"name\": \"Test\"}}"}, nil
		},
	}

	agent := NewAgent(mockProvider)

	// Add a test tool
	agent.AddTool(tools.NewTool(
		"test_tool",
		"Test tool",
		func(params struct {
			Name string `json:"name"`
		}) (string, error) {
			return "Hello, " + params.Name, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"name": {
					Type:        "string",
					Description: "A name",
				},
			},
			Required: []string{"name"},
		},
	))

	// Add a mock hook
	mockHook := &MockHook{}
	agent.WithHook(mockHook)

	// Run the agent
	_, err := agent.Run(context.Background(), "Test hook")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify that all hook methods were called
	if !mockHook.beforeGenerateCalled {
		t.Errorf("Expected BeforeGenerate to be called")
	}

	if !mockHook.afterGenerateCalled {
		t.Errorf("Expected AfterGenerate to be called")
	}

	if !mockHook.beforeToolCallCalled {
		t.Errorf("Expected BeforeToolCall to be called")
	}

	if !mockHook.afterToolCallCalled {
		t.Errorf("Expected AfterToolCall to be called")
	}
}

// TestDefaultAgent_ExtractToolCall tests the ExtractToolCall method
func TestDefaultAgent_ExtractToolCall(t *testing.T) {
	agent := &DefaultAgent{}

	tests := []struct {
		name           string
		content        string
		expectedTool   string
		expectedParams interface{}
		shouldCall     bool
	}{
		{
			name:           "JSON format",
			content:        "{\"tool\": \"test_tool\", \"params\": {\"name\": \"Test\"}}",
			expectedTool:   "test_tool",
			expectedParams: map[string]interface{}{"name": "Test"},
			shouldCall:     true,
		},
		{
			name:           "Code block format",
			content:        "I'll use the test tool.\n\n```json\n{\"tool\": \"test_tool\", \"params\": {\"name\": \"Test\"}}\n```",
			expectedTool:   "test_tool",
			expectedParams: map[string]interface{}{"name": "Test"},
			shouldCall:     true,
		},
		{
			name:           "Text format",
			content:        "Tool: test_tool\nParams: {\"name\": \"Test\"}",
			expectedTool:   "test_tool",
			expectedParams: map[string]interface{}{"name": "Test"},
			shouldCall:     true,
		},
		{
			name:           "No tool call",
			content:        "This is just a regular response with no tool call.",
			expectedTool:   "",
			expectedParams: nil,
			shouldCall:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, params, shouldCall := agent.ExtractToolCall(tt.content)

			if shouldCall != tt.shouldCall {
				t.Errorf("Expected shouldCall to be %v, got %v", tt.shouldCall, shouldCall)
			}

			if !tt.shouldCall {
				return
			}

			if tool != tt.expectedTool {
				t.Errorf("Expected tool to be '%s', got '%s'", tt.expectedTool, tool)
			}

			// Check params - convert both to string for easier comparison
			expectedParamsStr := fmt.Sprintf("%v", tt.expectedParams)
			paramsStr := fmt.Sprintf("%v", params)

			if expectedParamsStr != paramsStr {
				t.Errorf("Expected params to be '%v', got '%v'", tt.expectedParams, params)
			}
		})
	}
}
