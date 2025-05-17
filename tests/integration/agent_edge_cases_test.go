package integration

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestAgentEdgeCases tests various edge cases for the agent system
func TestAgentEdgeCases(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewMockProvider()

	// Create an agent
	agent := workflow.NewAgent(mockProvider)

	// Set system prompt
	agent.SetSystemPrompt("You are a helpful assistant that can answer questions and use tools.")

	// Create a calculator tool that will sometimes fail
	calculatorTool := tools.NewTool(
		"calculator",
		"Perform arithmetic calculations",
		func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			operation, ok := params["operation"].(string)
			if !ok {
				return nil, fmt.Errorf("operation must be a string")
			}

			a, ok := params["a"].(float64)
			if !ok {
				return nil, fmt.Errorf("a must be a number")
			}

			b, ok := params["b"].(float64)
			if !ok {
				return nil, fmt.Errorf("b must be a number")
			}

			// Add some edge cases
			switch operation {
			case "add":
				return a + b, nil
			case "subtract":
				return a - b, nil
			case "multiply":
				return a * b, nil
			case "divide":
				if b == 0 {
					return nil, errors.New("division by zero")
				}
				return a / b, nil
			case "will_fail":
				return nil, errors.New("this operation is designed to fail")
			default:
				return nil, fmt.Errorf("unknown operation: %s", operation)
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
			shouldContain: "available tools", // Changed from "not found" to "available tools" which is in the response
			mockResponses: []ldomain.Response{
				{
					Content: `{"tool": "nonexistent_tool", "params": {}}`,
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
					Content: `{"tool": "calculator", "params": {"operation": "will_fail", "a": 1, "b": 1}}`,
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
					Content: `{"tool": "calculator", "params": {"operation": "divide", "a": 42, "b": 0}}`,
				},
				{
					Content: "The calculator tool failed with error: division by zero",
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
				// Ensure we don't go beyond the provided responses
				if stage >= len(tc.mockResponses) {
					t.Fatalf("Unexpected call to GenerateMessage (stage %d, responses: %d)", stage, len(tc.mockResponses))
					return ldomain.Response{}, errors.New("unexpected call")
				}

				// Return the next response
				response := tc.mockResponses[stage]
				stage++
				return response, nil
			})

			// Run the agent
			result, err := agent.Run(context.Background(), tc.query)
			if err != nil {
				t.Fatalf("Agent Run failed unexpectedly: %v", err)
			}

			// Check the result
			strResult, ok := result.(string)
			if !ok {
				t.Fatalf("Expected string result, got: %T", result)
			}

			if !strings.Contains(strings.ToLower(strResult), strings.ToLower(tc.shouldContain)) {
				t.Errorf("Expected result to contain '%s', got: %s", tc.shouldContain, strResult)
			}

			// Verify we went through the expected number of stages
			if stage != tc.expectedStages {
				t.Errorf("Expected %d stages, got %d", tc.expectedStages, stage)
			}
		})
	}
}
