package integration

// ABOUTME: Integration tests for agent functionality with mock providers
// ABOUTME: Tests end-to-end agent behavior including tool usage and message handling

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestEndToEndAgent tests the agent system from end to end with mocks
func TestEndToEndAgent(t *testing.T) {
	// Create a mock provider with controlled responses
	mockProvider := provider.NewMockProvider()

	// Create an agent with new architecture
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("test-agent", "test", deps)

	// Set system prompt
	agent.SetSystemPrompt("You are a helpful assistant that can answer questions and use tools.")

	// Create a calculator tool
	calculatorTool := tools.NewTool(
		"calculator",
		"Perform arithmetic calculations",
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
			case "multiply":
				return params.A * params.B, nil
			case "divide":
				if params.B == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				return params.A / params.B, nil
			default:
				return 0, fmt.Errorf("unknown operation: %s", params.Operation)
			}
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"operation": {
					Type:        "string",
					Description: "The operation to perform (add, subtract, multiply, divide)",
					Enum:        []string{"add", "subtract", "multiply", "divide"},
				},
				"a": {
					Type:        "number",
					Description: "The first operand",
				},
				"b": {
					Type:        "number",
					Description: "The second operand",
				},
			},
			Required: []string{"operation", "a", "b"},
		},
	)

	// Add the calculator tool to the agent
	agent.AddTool(calculatorTool)

	// Mock the provider's GenerateMessage response
	// Set up a sequential behavior pattern:
	// 1. First call - We'll simulate this as the LLM deciding to use the calculator
	// 2. Second call - After getting the calculation result, returns the final answer
	callCount := 0
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		callCount++

		if callCount == 1 {
			// First call: LLM decides to use the calculator tool
			return ldomain.Response{
				Content: `I'll help you calculate that. Let me use the calculator.

<tool_calls>
[
  {
    "name": "calculator",
    "arguments": {
      "operation": "multiply",
      "a": 15,
      "b": 7
    }
  }
]
</tool_calls>

Let me calculate 15 × 7 for you.`,
			}, nil
		}

		// Second call: LLM responds with the final answer after getting the tool result
		// At this point, the messages will include the tool's result
		// We look for the tool result in the messages to form our response
		var toolResult string
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == ldomain.RoleUser {
				// Check if this message contains tool results
				if len(messages[i].Content) > 0 && messages[i].Content[0].Type == ldomain.ContentTypeText {
					content := messages[i].Content[0].Text
					if strings.Contains(content, "Tool results:") && strings.Contains(content, "105") {
						toolResult = "105"
						break
					}
				}
			}
		}

		return ldomain.Response{
			Content: fmt.Sprintf("The result of 15 × 7 is %s.", toolResult),
		}, nil
	})

	// Create test context
	ctx := context.Background()

	// Create initial state with user input
	state := domain.NewState()
	state.Set("user_input", "What is 15 times 7?")

	// Run the agent
	finalState, err := agent.Run(ctx, state)
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	// Check final output
	output, ok := finalState.Get("output")
	if !ok {
		t.Fatal("No output in final state")
	}

	outputStr, ok := output.(string)
	if !ok {
		t.Fatal("Output is not a string")
	}

	// Verify the response contains the expected answer
	if !strings.Contains(outputStr, "105") {
		t.Errorf("Expected output to contain '105', got: %s", outputStr)
	}

	// Verify tool was called
	if callCount != 2 {
		t.Errorf("Expected 2 LLM calls (one for tool decision, one for final answer), got %d", callCount)
	}
}

// TestAgentWithMultipleTools tests an agent that can use multiple tools
func TestAgentWithMultipleTools(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewMockProvider()

	// Create an agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("multi-tool-agent", "test", deps)
	agent.SetSystemPrompt("You are a helpful assistant with access to multiple tools.")

	// Create a weather tool
	weatherTool := tools.NewTool(
		"get_weather",
		"Get the current weather for a location",
		func(params struct {
			Location string `json:"location"`
		}) (map[string]interface{}, error) {
			// Simulate weather data
			weatherData := map[string]interface{}{
				"temperature": 22,
				"condition":   "sunny",
				"humidity":    65,
				"location":    params.Location,
			}
			return weatherData, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"location": {
					Type:        "string",
					Description: "The location to get weather for",
				},
			},
			Required: []string{"location"},
		},
	)

	// Create a time tool
	timeTool := tools.NewTool(
		"get_time",
		"Get the current time in a timezone",
		func(params struct {
			Timezone string `json:"timezone"`
		}) (string, error) {
			// Simulate time data
			return fmt.Sprintf("The time in %s is 3:45 PM", params.Timezone), nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"timezone": {
					Type:        "string",
					Description: "The timezone to get time for",
				},
			},
			Required: []string{"timezone"},
		},
	)

	// Add tools to agent
	agent.AddTool(weatherTool)
	agent.AddTool(timeTool)

	// Mock the provider to use both tools
	callCount := 0
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		callCount++

		if callCount == 1 {
			// First call: Use both tools
			return ldomain.Response{
				Content: `I'll check the weather and time for you.

<tool_calls>
[
  {
    "name": "get_weather",
    "arguments": {
      "location": "London"
    }
  },
  {
    "name": "get_time",
    "arguments": {
      "timezone": "Europe/London"
    }
  }
]
</tool_calls>

Let me fetch that information.`,
			}, nil
		}

		// Second call: Respond with the combined results
		return ldomain.Response{
			Content: "In London, it's currently 22°C and sunny with 65% humidity. The time in Europe/London is 3:45 PM.",
		}, nil
	})

	// Create test context
	ctx := context.Background()

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "What's the weather and time in London?")

	// Run the agent
	finalState, err := agent.Run(ctx, state)
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	// Check final output
	output, ok := finalState.Get("output")
	if !ok {
		t.Fatal("No output in final state")
	}

	outputStr, ok := output.(string)
	if !ok {
		t.Fatal("Output is not a string")
	}

	// Verify the response contains expected information
	expectedParts := []string{"22°C", "sunny", "3:45 PM", "London"}
	for _, part := range expectedParts {
		if !strings.Contains(outputStr, part) {
			t.Errorf("Expected output to contain '%s', got: %s", part, outputStr)
		}
	}
}

// TestAgentErrorHandling tests how the agent handles errors
func TestAgentErrorHandling(t *testing.T) {
	// Create a mock provider that returns an error
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{}, fmt.Errorf("simulated provider error")
	})

	// Create an agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("error-test-agent", "test", deps)

	// Create test context
	ctx := context.Background()

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Hello, how are you?")

	// Run the agent - should return error
	_, err := agent.Run(ctx, state)
	if err == nil {
		t.Fatal("Expected error from agent run, got nil")
	}

	if !strings.Contains(err.Error(), "simulated provider error") {
		t.Errorf("Expected error to contain 'simulated provider error', got: %v", err)
	}
}

// TestAgentHooks tests the hook functionality
func TestAgentHooks(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{
			Content: "Hello! I'm doing well, thank you for asking.",
		}, nil
	})

	// Create an agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("hook-test-agent", "test", deps)

	// Create a test hook to track calls
	var beforeGenerateCalled bool
	var afterGenerateCalled bool
	var capturedMessages []ldomain.Message
	var capturedResponse ldomain.Response

	testHook := &testAgentHook{
		beforeGenerate: func(ctx context.Context, messages []ldomain.Message) {
			beforeGenerateCalled = true
			capturedMessages = messages
		},
		afterGenerate: func(ctx context.Context, response ldomain.Response, err error) {
			afterGenerateCalled = true
			capturedResponse = response
		},
	}

	agent.WithHook(testHook)

	// Create test context
	ctx := context.Background()

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Hello, how are you?")

	// Run the agent
	finalState, err := agent.Run(ctx, state)
	if err != nil {
		t.Fatalf("Agent run failed: %v", err)
	}

	// Verify hooks were called
	if !beforeGenerateCalled {
		t.Error("BeforeGenerate hook was not called")
	}

	if !afterGenerateCalled {
		t.Error("AfterGenerate hook was not called")
	}

	// Verify captured data
	if len(capturedMessages) == 0 {
		t.Error("No messages captured in BeforeGenerate hook")
	}

	if capturedResponse.Content == "" {
		t.Error("No response captured in AfterGenerate hook")
	}

	// Check final output
	output, ok := finalState.Get("output")
	if !ok {
		t.Fatal("No output in final state")
	}

	if output != capturedResponse.Content {
		t.Errorf("Output mismatch: expected %s, got %s", capturedResponse.Content, output)
	}
}

// testAgentHook is a test implementation of the Hook interface
type testAgentHook struct {
	beforeGenerate func(ctx context.Context, messages []ldomain.Message)
	afterGenerate  func(ctx context.Context, response ldomain.Response, err error)
	beforeToolCall func(ctx context.Context, tool string, params map[string]interface{})
	afterToolCall  func(ctx context.Context, tool string, result interface{}, err error)
}

func (h *testAgentHook) BeforeGenerate(ctx context.Context, messages []ldomain.Message) {
	if h.beforeGenerate != nil {
		h.beforeGenerate(ctx, messages)
	}
}

func (h *testAgentHook) AfterGenerate(ctx context.Context, response ldomain.Response, err error) {
	if h.afterGenerate != nil {
		h.afterGenerate(ctx, response, err)
	}
}

func (h *testAgentHook) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	if h.beforeToolCall != nil {
		h.beforeToolCall(ctx, tool, params)
	}
}

func (h *testAgentHook) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	if h.afterToolCall != nil {
		h.afterToolCall(ctx, tool, result, err)
	}
}