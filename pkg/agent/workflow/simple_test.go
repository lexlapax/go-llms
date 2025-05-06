package workflow

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestSimpleToolExecution tests a simple tool execution flow
func TestSimpleToolExecution(t *testing.T) {
	// We'll create a simple series of responses:
	// 1. First response calls calculator tool
	// 2. Second response is the final answer
	responses := []ldomain.Response{
		{Content: `{"tool": "calculator", "params": {"expression": "2+2"}}`},
		{Content: "The answer is 4"},
	}

	responseIdx := 0
	mockProvider := &MockProvider{
		generateMessageFunc: func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
			resp := responses[responseIdx]
			responseIdx++
			return resp, nil
		},
	}

	agent := NewAgent(mockProvider)

	// Add calculator tool
	agent.AddTool(tools.NewTool(
		"calculator",
		"Calculates mathematical expressions",
		func(params struct {
			Expression string `json:"expression"`
		}) (int, error) {
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

	// Run the agent
	result, err := agent.Run(context.Background(), "What is 2+2?")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// The result should be the final response
	expectedResult := "The answer is 4"
	if result != expectedResult {
		t.Errorf("Expected result to be '%s', got '%v'", expectedResult, result)
	}

	// There should have been 2 calls to the provider
	if responseIdx != 2 {
		t.Errorf("Expected 2 calls to the provider, got %d", responseIdx)
	}
}
