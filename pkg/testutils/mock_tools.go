package testutils

// ABOUTME: Mock tool implementations for agent testing
// ABOUTME: Includes configurable tools with success/failure modes

import (
	"fmt"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// MockTool is a mock implementation of the Tool interface for testing
type MockTool struct {
	ToolName        string
	ToolDescription string
	Schema          *sdomain.Schema
	Executor        func(ctx *domain.ToolContext, params interface{}) (interface{}, error)
}

func (t MockTool) Name() string {
	return t.ToolName
}

func (t MockTool) Description() string {
	return t.ToolDescription
}

func (t MockTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	if t.Executor != nil {
		return t.Executor(ctx, params)
	}
	return nil, nil
}

func (t MockTool) ParameterSchema() *sdomain.Schema {
	return t.Schema
}

func (t MockTool) OutputSchema() *sdomain.Schema {
	return nil
}

func (t MockTool) UsageInstructions() string {
	return ""
}

func (t MockTool) Examples() []domain.ToolExample {
	return nil
}

func (t MockTool) Constraints() []string {
	return nil
}

func (t MockTool) ErrorGuidance() map[string]string {
	return nil
}

func (t MockTool) Category() string {
	return "test"
}

func (t MockTool) Tags() []string {
	return []string{"test", "mock"}
}

func (t MockTool) Version() string {
	return "1.0.0"
}

func (t MockTool) IsDeterministic() bool {
	return true
}

func (t MockTool) IsDestructive() bool {
	return false
}

func (t MockTool) RequiresConfirmation() bool {
	return false
}

func (t MockTool) EstimatedLatency() string {
	return "fast"
}

func (t MockTool) ToMCPDefinition() domain.MCPToolDefinition {
	return domain.MCPToolDefinition{
		Name:        t.ToolName,
		Description: t.ToolDescription,
		InputSchema: t.Schema,
	}
}

// CreateCalculatorTool is a helper function to create a calculator tool for tests
func CreateCalculatorTool() domain.Tool {
	return MockTool{
		ToolName:        "calculator",
		ToolDescription: "Perform mathematical calculations",
		Executor: func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
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

// CreateMockTool is a helper function to create a mock tool for tests
func CreateMockTool(name string, description string, schema *sdomain.Schema) domain.Tool {
	return MockTool{
		ToolName:        name,
		ToolDescription: description,
		Executor: func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			return fmt.Sprintf("Executed %s tool with params: %v", name, params), nil
		},
		Schema: schema,
	}
}
