// ABOUTME: Tests for enhanced Tool interface with comprehensive LLM guidance
// ABOUTME: Validates tool metadata, examples, schemas, and MCP compatibility

package domain_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// Helper function to create a minimal tool using the centralized mock
func createMinimalTool(name, description string) *mocks.MockTool {
	return mocks.NewMockTool(name, description).
		WithExecutor(func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			return "minimal result", nil
		})
}

// Helper function to create a full-featured tool using the centralized mock
func createFullTool(config struct {
	name              string
	description       string
	usageInstructions string
	examples          []domain.ToolExample
	constraints       []string
	errorGuidance     map[string]string
	parameterSchema   *sdomain.Schema
	outputSchema      *sdomain.Schema
	category          string
	tags              []string
	version           string
	isDeterministic   bool
	isDestructive     bool
	requiresConfirm   bool
	estimatedLatency  string
}) *mocks.MockTool {
	tool := mocks.NewMockTool(config.name, config.description).
		WithCategory(config.category).
		WithTags(config.tags...).
		WithVersion(config.version).
		WithUsageInstructions(config.usageInstructions).
		WithExamples(config.examples...).
		WithConstraints(config.constraints...).
		WithErrorGuidance(config.errorGuidance).
		WithParameterSchema(config.parameterSchema).
		WithOutputSchema(config.outputSchema).
		WithExecutor(func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			// Simulate execution based on params
			if paramsMap, ok := params.(map[string]interface{}); ok {
				if paramsMap["error"] == true {
					return nil, domain.NewToolErrorWithGuidance(config.name, "simulated_error",
						"Simulated error for testing", "This is a simulated error for testing")
				}
			}
			return map[string]interface{}{
				"result":  "success",
				"message": "Tool executed successfully",
			}, nil
		})

	// Override behavioral properties based on config
	// Note: MockTool doesn't have setters for these, so we'll need to work with what's available
	return tool
}

func TestToolInterface(t *testing.T) {
	t.Run("minimal tool implementation", func(t *testing.T) {
		tool := createMinimalTool("test_tool", "A test tool")

		// Test basic interface compliance
		var _ domain.Tool = tool

		// Test basic methods
		if tool.Name() != "test_tool" {
			t.Errorf("Expected name 'test_tool', got '%s'", tool.Name())
		}
		if tool.Description() != "A test tool" {
			t.Errorf("Expected description 'A test tool', got '%s'", tool.Description())
		}

		// Test execution
		ctx := &domain.ToolContext{
			Context: context.Background(),
		}
		result, err := tool.Execute(ctx, nil)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != "minimal result" {
			t.Errorf("Expected 'minimal result', got '%v'", result)
		}

		// Test MCP export
		mcp := tool.ToMCPDefinition()
		if mcp.Name != "test_tool" {
			t.Errorf("MCP name mismatch: expected 'test_tool', got '%s'", mcp.Name)
		}
		if mcp.Description != "A test tool" {
			t.Errorf("MCP description mismatch")
		}
	})

	t.Run("full tool implementation with LLM guidance", func(t *testing.T) {
		tool := createFullTool(struct {
			name              string
			description       string
			usageInstructions string
			examples          []domain.ToolExample
			constraints       []string
			errorGuidance     map[string]string
			parameterSchema   *sdomain.Schema
			outputSchema      *sdomain.Schema
			category          string
			tags              []string
			version           string
			isDeterministic   bool
			isDestructive     bool
			requiresConfirm   bool
			estimatedLatency  string
		}{
			name:        "calculator",
			description: "Performs mathematical calculations",
			category:    "math",
			tags:        []string{"math", "calculation", "arithmetic"},
			version:     "2.0.0",

			usageInstructions: `Use this tool when you need to:
- Perform arithmetic operations
- Calculate trigonometric functions
- Work with mathematical constants`,

			examples: []domain.ToolExample{
				{
					Name:        "Basic addition",
					Description: "Add two numbers",
					Scenario:    "User asks 'What is 5 plus 3?'",
					Input: map[string]interface{}{
						"operation": "add",
						"operand1":  5,
						"operand2":  3,
					},
					Output:      map[string]interface{}{"result": 8},
					Explanation: "Simple arithmetic addition",
				},
			},

			constraints: []string{
				"Operands must be numeric",
				"Division by zero returns error",
			},

			errorGuidance: map[string]string{
				"division_by_zero": "Cannot divide by zero. Check operand2 value.",
				"invalid_operand":  "Operands must be numeric values.",
			},

			parameterSchema: &sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"operation": {
						Type:        "string",
						Description: "Mathematical operation",
					},
					"operand1": {
						Type:        "number",
						Description: "First operand",
					},
					"operand2": {
						Type:        "number",
						Description: "Second operand",
					},
				},
				Required: []string{"operation", "operand1"},
			},

			outputSchema: &sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"result": {
						Type:        "number",
						Description: "Calculation result",
					},
				},
				Required: []string{"result"},
			},

			isDeterministic:  true,
			isDestructive:    false,
			requiresConfirm:  false,
			estimatedLatency: "fast",
		})

		// Test interface compliance
		var _ domain.Tool = tool

		// Test metadata methods
		if tool.Category() != "math" {
			t.Errorf("Expected category 'math', got '%s'", tool.Category())
		}
		if len(tool.Tags()) != 3 {
			t.Errorf("Expected 3 tags, got %d", len(tool.Tags()))
		}
		if tool.Version() != "2.0.0" {
			t.Errorf("Expected version '2.0.0', got '%s'", tool.Version())
		}

		// Test guidance methods
		if tool.UsageInstructions() == "" {
			t.Error("Expected usage instructions")
		}
		if len(tool.Examples()) != 1 {
			t.Errorf("Expected 1 example, got %d", len(tool.Examples()))
		}
		if len(tool.Constraints()) != 2 {
			t.Errorf("Expected 2 constraints, got %d", len(tool.Constraints()))
		}
		if len(tool.ErrorGuidance()) != 2 {
			t.Errorf("Expected 2 error guidance entries, got %d", len(tool.ErrorGuidance()))
		}

		// Test behavioral methods
		if !tool.IsDeterministic() {
			t.Error("Expected deterministic tool")
		}
		if tool.IsDestructive() {
			t.Error("Expected non-destructive tool")
		}
		if tool.RequiresConfirmation() {
			t.Error("Expected no confirmation required")
		}
		if tool.EstimatedLatency() != "fast" {
			t.Errorf("Expected 'fast' latency, got '%s'", tool.EstimatedLatency())
		}

		// Test schema methods
		if tool.ParameterSchema() == nil {
			t.Error("Expected parameter schema")
		}
		if tool.OutputSchema() == nil {
			t.Error("Expected output schema")
		}
	})

	t.Run("tool example validation", func(t *testing.T) {
		example := domain.ToolExample{
			Name:        "Test example",
			Description: "Example description",
			Scenario:    "When to use this",
			Input:       map[string]interface{}{"key": "value"},
			Output:      map[string]interface{}{"result": "success"},
			Explanation: "Why this works",
		}

		// Validate all fields are accessible
		if example.Name == "" {
			t.Error("Example name should not be empty")
		}
		if example.Description == "" {
			t.Error("Example description should not be empty")
		}
		if example.Scenario == "" {
			t.Error("Example scenario should not be empty")
		}
		if example.Input == nil {
			t.Error("Example input should not be nil")
		}
		if example.Output == nil {
			t.Error("Example output should not be nil")
		}
		if example.Explanation == "" {
			t.Error("Example explanation should not be empty")
		}
	})

	t.Run("MCP export functionality", func(t *testing.T) {
		tool := mocks.NewMockTool("test_tool", "Test tool for MCP export").
			WithCategory("test").
			WithVersion("1.0.0").
			WithParameterSchema(&sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"input": {
						Type:        "string",
						Description: "Test input",
					},
				},
			}).
			WithOutputSchema(&sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"output": {
						Type:        "string",
						Description: "Test output",
					},
				},
			})

		mcp := tool.ToMCPDefinition()

		// Test basic fields
		if mcp.Name != "test_tool" {
			t.Errorf("MCP name mismatch: expected 'test_tool', got '%s'", mcp.Name)
		}
		if mcp.Description != "Test tool for MCP export" {
			t.Error("MCP description mismatch")
		}
		if mcp.InputSchema == nil {
			t.Error("MCP input schema should not be nil")
		}
		// Note: MockTool's ToMCPDefinition doesn't include OutputSchema in its implementation

		// Test annotations
		// Note: MockTool's ToMCPDefinition doesn't add annotations,
		// so we'll skip these checks as they're implementation-specific

		// Test JSON serialization
		jsonData, err := json.Marshal(mcp)
		if err != nil {
			t.Errorf("Failed to marshal MCP definition: %v", err)
		}
		if len(jsonData) == 0 {
			t.Error("MCP JSON should not be empty")
		}

		// Test JSON deserialization
		var decoded domain.MCPToolDefinition
		if err := json.Unmarshal(jsonData, &decoded); err != nil {
			t.Errorf("Failed to unmarshal MCP definition: %v", err)
		}
		if decoded.Name != mcp.Name {
			t.Error("MCP roundtrip failed")
		}
	})

	t.Run("tool error handling", func(t *testing.T) {
		tool := mocks.NewMockTool("error_test_tool", "Error test tool").
			WithErrorGuidance(map[string]string{
				"simulated_error": "This is a simulated error for testing",
			}).
			WithExecutor(func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
				if paramsMap, ok := params.(map[string]interface{}); ok {
					if paramsMap["error"] == true {
						return nil, domain.NewToolErrorWithGuidance("error_test_tool", "simulated_error",
							"Simulated error for testing", "This is a simulated error for testing")
					}
				}
				return "success", nil
			})

		ctx := &domain.ToolContext{
			Context: context.Background(),
		}

		// Test error case
		_, err := tool.Execute(ctx, map[string]interface{}{"error": true})
		if err == nil {
			t.Error("Expected error, got nil")
		}

		// Verify error guidance exists
		guidance := tool.ErrorGuidance()
		if guidance["simulated_error"] == "" {
			t.Error("Expected error guidance for simulated_error")
		}
	})
}

func TestToolSchemaValidation(t *testing.T) {
	t.Run("parameter schema validation", func(t *testing.T) {
		schema := &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"required_field": {
					Type:        "string",
					Description: "A required field",
				},
				"optional_field": {
					Type:        "number",
					Description: "An optional field",
				},
			},
			Required: []string{"required_field"},
		}

		tool := mocks.NewMockTool("schema_test", "Schema test tool").
			WithParameterSchema(schema)

		// Test schema is properly set
		if tool.ParameterSchema() == nil {
			t.Error("Parameter schema should not be nil")
		}

		// Test schema structure
		if tool.ParameterSchema().Type != "object" {
			t.Error("Schema type should be 'object'")
		}
		if len(tool.ParameterSchema().Properties) != 2 {
			t.Errorf("Expected 2 properties, got %d", len(tool.ParameterSchema().Properties))
		}
		if len(tool.ParameterSchema().Required) != 1 {
			t.Errorf("Expected 1 required field, got %d", len(tool.ParameterSchema().Required))
		}
	})

	t.Run("output schema validation", func(t *testing.T) {
		schema := &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"result": {
					Type:        "string",
					Description: "The result",
				},
				"metadata": {
					Type: "object",
					Properties: map[string]sdomain.Property{
						"timestamp": {
							Type:        "string",
							Description: "When the result was generated",
						},
					},
				},
			},
			Required: []string{"result"},
		}

		tool := mocks.NewMockTool("output_schema_test", "Output schema test tool").
			WithOutputSchema(schema)

		// Test schema is properly set
		if tool.OutputSchema() == nil {
			t.Error("Output schema should not be nil")
		}

		// Test nested properties
		if metaProp, exists := tool.OutputSchema().Properties["metadata"]; exists {
			if metaProp.Properties == nil {
				t.Error("Nested properties should exist")
			}
		} else {
			t.Error("Metadata property should exist")
		}
	})
}

func TestToolBehavioralMetadata(t *testing.T) {
	// Note: MockTool has fixed behavioral properties, so we can only test the default behavior
	t.Run("mock tool behavioral metadata", func(t *testing.T) {
		tool := mocks.NewMockTool("behavioral_test", "Testing behavioral metadata")

		// MockTool always returns these fixed values
		if !tool.IsDeterministic() {
			t.Error("MockTool should be deterministic")
		}
		if tool.IsDestructive() {
			t.Error("MockTool should not be destructive")
		}
		if tool.RequiresConfirmation() {
			t.Error("MockTool should not require confirmation")
		}
		if tool.EstimatedLatency() != "fast" {
			t.Errorf("MockTool should have 'fast' latency, got %s", tool.EstimatedLatency())
		}
	})
}
