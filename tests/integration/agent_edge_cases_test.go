package integration

// ABOUTME: Integration tests for agent edge cases and error handling
// ABOUTME: Tests tool failures, missing tools, and other edge conditions

import (
	"context"
	"errors"
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

// TestAgentEdgeCases tests various edge cases for the agent system
func TestAgentEdgeCases(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewMockProvider()

	// Create an agent with new architecture
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("edge-case-agent", "test", deps)

	// Set system prompt
	agent.SetSystemPrompt("You are a helpful assistant that can answer questions and use tools.")

	// Create a calculator tool that will sometimes fail
	calculatorTool := tools.NewTool(
		"calculator",
		"Perform arithmetic calculations",
		func(params struct {
			Operation string  `json:"operation"`
			A         float64 `json:"a"`
			B         float64 `json:"b"`
		}) (interface{}, error) {
			// Add some edge cases
			switch params.Operation {
			case "add":
				return params.A + params.B, nil
			case "subtract":
				return params.A - params.B, nil
			case "multiply":
				return params.A * params.B, nil
			case "divide":
				if params.B == 0 {
					return nil, errors.New("division by zero")
				}
				return params.A / params.B, nil
			case "will_fail":
				return nil, errors.New("this operation is designed to fail")
			default:
				return nil, fmt.Errorf("unknown operation: %s", params.Operation)
			}
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"operation": {
					Type:        "string",
					Description: "The operation to perform",
					Enum:        []string{"add", "subtract", "multiply", "divide", "will_fail"},
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

	// Add the tool to the agent
	agent.AddTool(calculatorTool)

	// Test cases
	testCases := []struct {
		name           string
		query          string
		shouldContain  string
		mockResponses  []ldomain.Response
		expectedStages int
	}{
		{
			name:          "Tool not found",
			query:         "Use a nonexistent tool",
			shouldContain: "Available tools", // The agent should list available tools
			mockResponses: []ldomain.Response{
				{
					Content: `I'll try to use the nonexistent tool.

<tool_calls>
[
  {
    "name": "nonexistent_tool",
    "arguments": {}
  }
]
</tool_calls>`,
				},
				{
					Content: "I couldn't find that tool. Available tools include: calculator.",
				},
			},
			expectedStages: 2,
		},
		{
			name:          "Tool fails",
			query:         "Use a tool that will fail",
			shouldContain: "failed",
			mockResponses: []ldomain.Response{
				{
					Content: `I'll use the calculator with the will_fail operation.

<tool_calls>
[
  {
    "name": "calculator",
    "arguments": {
      "operation": "will_fail",
      "a": 1,
      "b": 1
    }
  }
]
</tool_calls>`,
				},
				{
					Content: "The calculator tool failed with error: this operation is designed to fail",
				},
			},
			expectedStages: 2,
		},
		{
			name:          "Divide by zero",
			query:         "Divide by zero",
			shouldContain: "division by zero",
			mockResponses: []ldomain.Response{
				{
					Content: `I'll attempt to divide by zero.

<tool_calls>
[
  {
    "name": "calculator",
    "arguments": {
      "operation": "divide",
      "a": 42,
      "b": 0
    }
  }
]
</tool_calls>`,
				},
				{
					Content: "The calculator tool failed with error: division by zero",
				},
			},
			expectedStages: 2,
		},
		{
			name:          "Invalid tool parameters",
			query:         "Use calculator with invalid params",
			shouldContain: "invalid",
			mockResponses: []ldomain.Response{
				{
					Content: `I'll use the calculator with invalid parameters.

<tool_calls>
[
  {
    "name": "calculator",
    "arguments": {
      "operation": "add",
      "x": 1,
      "y": 2
    }
  }
]
</tool_calls>`,
				},
				{
					Content: "The calculator tool failed due to invalid parameters.",
				},
			},
			expectedStages: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset provider for each test case
			stage := 0
			mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
				if stage >= len(tc.mockResponses) {
					return ldomain.Response{}, fmt.Errorf("unexpected stage %d", stage)
				}
				response := tc.mockResponses[stage]
				stage++
				return response, nil
			})

			// Create test context
			ctx := context.Background()

			// Create initial state
			state := domain.NewState()
			state.Set("user_input", tc.query)

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

			// Verify the response contains expected content
			if !strings.Contains(outputStr, tc.shouldContain) {
				t.Errorf("Expected output to contain '%s', got: %s", tc.shouldContain, outputStr)
			}

			// Verify number of stages
			if stage != tc.expectedStages {
				t.Errorf("Expected %d stages, got %d", tc.expectedStages, stage)
			}
		})
	}
}

// TestAgentWithNoTools tests agent behavior when no tools are available
func TestAgentWithNoTools(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{
			Content: "I don't have any tools available, but I can help answer your question directly.",
		}, nil
	})

	// Create an agent with no tools
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("no-tools-agent", "test", deps)
	agent.SetSystemPrompt("You are a helpful assistant.")

	// Create test context
	ctx := context.Background()

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Can you use a tool to help me?")

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

	// Verify the response
	if !strings.Contains(outputStr, "don't have any tools") {
		t.Errorf("Expected output to mention no tools available, got: %s", outputStr)
	}
}

// TestAgentConcurrentToolCalls tests agent handling multiple tool calls
func TestAgentConcurrentToolCalls(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewMockProvider()

	// Create an agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("concurrent-tools-agent", "test", deps)
	agent.SetSystemPrompt("You are a helpful assistant with multiple tools.")

	// Create multiple tools
	addTool := tools.NewTool(
		"add",
		"Add two numbers",
		func(params struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}) (float64, error) {
			return params.A + params.B, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"a": {Type: "number"},
				"b": {Type: "number"},
			},
			Required: []string{"a", "b"},
		},
	)

	multiplyTool := tools.NewTool(
		"multiply",
		"Multiply two numbers",
		func(params struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}) (float64, error) {
			return params.A * params.B, nil
		},
		&sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"a": {Type: "number"},
				"b": {Type: "number"},
			},
			Required: []string{"a", "b"},
		},
	)

	// Add tools to agent
	agent.AddTool(addTool)
	agent.AddTool(multiplyTool)

	// Mock responses
	callCount := 0
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		callCount++
		if callCount == 1 {
			// First call: Use multiple tools
			return ldomain.Response{
				Content: `I'll calculate both operations for you.

<tool_calls>
[
  {
    "name": "add",
    "arguments": {
      "a": 10,
      "b": 5
    }
  },
  {
    "name": "multiply",
    "arguments": {
      "a": 3,
      "b": 7
    }
  }
]
</tool_calls>`,
			}, nil
		}
		// Second call: Return results
		return ldomain.Response{
			Content: "The results are: 10 + 5 = 15, and 3 × 7 = 21.",
		}, nil
	})

	// Create test context
	ctx := context.Background()

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Calculate 10+5 and 3*7")

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

	// Verify both results are in the output
	expectedResults := []string{"15", "21"}
	for _, result := range expectedResults {
		if !strings.Contains(outputStr, result) {
			t.Errorf("Expected output to contain '%s', got: %s", result, outputStr)
		}
	}
}