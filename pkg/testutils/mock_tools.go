package testutils

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// MockTool is a mock implementation of the Tool interface for testing
type MockTool struct {
	ToolName        string
	ToolDescription string
	Schema          *sdomain.Schema
	Executor        func(ctx context.Context, params interface{}) (interface{}, error)
}

func (t MockTool) Name() string {
	return t.ToolName
}

func (t MockTool) Description() string {
	return t.ToolDescription
}

func (t MockTool) Execute(ctx context.Context, params interface{}) (interface{}, error) {
	if t.Executor != nil {
		return t.Executor(ctx, params)
	}
	return nil, nil
}

func (t MockTool) ParameterSchema() *sdomain.Schema {
	return t.Schema
}

// CreateCalculatorTool is a helper function to create a calculator tool for tests
func CreateCalculatorTool() domain.Tool {
	return MockTool{
		ToolName:        "calculator",
		ToolDescription: "Perform mathematical calculations",
		Executor: func(ctx context.Context, params interface{}) (interface{}, error) {
			return map[string]interface{}{
				"result": 4,
			}, nil
		},
		Schema: &sdomain.Schema{
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
}