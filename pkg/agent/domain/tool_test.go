// ABOUTME: Tests for enhanced Tool interface with comprehensive LLM guidance
// ABOUTME: Validates tool metadata, examples, schemas, and MCP compatibility

package domain_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// mockMinimalTool implements the bare minimum Tool interface
type mockMinimalTool struct {
	name        string
	description string
}

func (t *mockMinimalTool) Name() string        { return t.name }
func (t *mockMinimalTool) Description() string { return t.description }
func (t *mockMinimalTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	return "minimal result", nil
}
func (t *mockMinimalTool) ParameterSchema() *sdomain.Schema { return nil }
func (t *mockMinimalTool) OutputSchema() *sdomain.Schema    { return nil }
func (t *mockMinimalTool) UsageInstructions() string        { return "" }
func (t *mockMinimalTool) Examples() []domain.ToolExample   { return nil }
func (t *mockMinimalTool) Constraints() []string            { return nil }
func (t *mockMinimalTool) ErrorGuidance() map[string]string { return nil }
func (t *mockMinimalTool) Category() string                 { return "" }
func (t *mockMinimalTool) Tags() []string                   { return nil }
func (t *mockMinimalTool) Version() string                  { return "1.0.0" }
func (t *mockMinimalTool) IsDeterministic() bool            { return true }
func (t *mockMinimalTool) IsDestructive() bool              { return false }
func (t *mockMinimalTool) RequiresConfirmation() bool       { return false }
func (t *mockMinimalTool) EstimatedLatency() string         { return "fast" }
func (t *mockMinimalTool) ToMCPDefinition() domain.MCPToolDefinition {
	return domain.MCPToolDefinition{
		Name:        t.name,
		Description: t.description,
	}
}

// mockFullTool implements Tool interface with full LLM guidance
type mockFullTool struct {
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
}

func (t *mockFullTool) Name() string        { return t.name }
func (t *mockFullTool) Description() string { return t.description }
func (t *mockFullTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	// Simulate execution based on params
	if paramsMap, ok := params.(map[string]interface{}); ok {
		if paramsMap["error"] == true {
			return nil, domain.NewToolErrorWithGuidance(t.name, "simulated_error",
				"Simulated error for testing", "This is a simulated error for testing")
		}
	}
	return map[string]interface{}{
		"result":  "success",
		"message": "Tool executed successfully",
	}, nil
}
func (t *mockFullTool) ParameterSchema() *sdomain.Schema { return t.parameterSchema }
func (t *mockFullTool) OutputSchema() *sdomain.Schema    { return t.outputSchema }
func (t *mockFullTool) UsageInstructions() string        { return t.usageInstructions }
func (t *mockFullTool) Examples() []domain.ToolExample   { return t.examples }
func (t *mockFullTool) Constraints() []string            { return t.constraints }
func (t *mockFullTool) ErrorGuidance() map[string]string { return t.errorGuidance }
func (t *mockFullTool) Category() string                 { return t.category }
func (t *mockFullTool) Tags() []string                   { return t.tags }
func (t *mockFullTool) Version() string                  { return t.version }
func (t *mockFullTool) IsDeterministic() bool            { return t.isDeterministic }
func (t *mockFullTool) IsDestructive() bool              { return t.isDestructive }
func (t *mockFullTool) RequiresConfirmation() bool       { return t.requiresConfirm }
func (t *mockFullTool) EstimatedLatency() string         { return t.estimatedLatency }
func (t *mockFullTool) ToMCPDefinition() domain.MCPToolDefinition {
	annotations := make(map[string]interface{})

	// Add behavioral metadata
	annotations["deterministic"] = t.isDeterministic
	annotations["destructive"] = t.isDestructive
	annotations["requires_confirmation"] = t.requiresConfirm
	annotations["estimated_latency"] = t.estimatedLatency
	annotations["category"] = t.category
	annotations["tags"] = t.tags
	annotations["version"] = t.version

	// Add guidance if present
	if t.usageInstructions != "" {
		annotations["usage_instructions"] = t.usageInstructions
	}
	if len(t.examples) > 0 {
		annotations["examples"] = t.examples
	}
	if len(t.constraints) > 0 {
		annotations["constraints"] = t.constraints
	}
	if len(t.errorGuidance) > 0 {
		annotations["error_guidance"] = t.errorGuidance
	}

	return domain.MCPToolDefinition{
		Name:         t.name,
		Description:  t.description,
		InputSchema:  t.parameterSchema,
		OutputSchema: t.outputSchema,
		Annotations:  annotations,
	}
}

func TestToolInterface(t *testing.T) {
	t.Run("minimal tool implementation", func(t *testing.T) {
		tool := &mockMinimalTool{
			name:        "test_tool",
			description: "A test tool",
		}

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
		tool := &mockFullTool{
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
		}

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
		tool := &mockFullTool{
			name:        "test_tool",
			description: "Test tool for MCP export",
			category:    "test",
			version:     "1.0.0",
			parameterSchema: &sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"input": {
						Type:        "string",
						Description: "Test input",
					},
				},
			},
			outputSchema: &sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"output": {
						Type:        "string",
						Description: "Test output",
					},
				},
			},
			isDeterministic: true,
			isDestructive:   false,
		}

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
		if mcp.OutputSchema == nil {
			t.Error("MCP output schema should not be nil")
		}

		// Test annotations
		if mcp.Annotations == nil {
			t.Error("MCP annotations should not be nil")
		}
		if mcp.Annotations["category"] != "test" {
			t.Error("MCP category annotation mismatch")
		}
		if mcp.Annotations["version"] != "1.0.0" {
			t.Error("MCP version annotation mismatch")
		}
		if mcp.Annotations["deterministic"] != true {
			t.Error("MCP deterministic annotation mismatch")
		}

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
		tool := &mockFullTool{
			name: "error_test_tool",
			errorGuidance: map[string]string{
				"simulated_error": "This is a simulated error for testing",
			},
		}

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

		tool := &mockFullTool{
			name:            "schema_test",
			parameterSchema: schema,
		}

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

		tool := &mockFullTool{
			name:         "output_schema_test",
			outputSchema: schema,
		}

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
	testCases := []struct {
		name             string
		isDeterministic  bool
		isDestructive    bool
		requiresConfirm  bool
		estimatedLatency string
	}{
		{
			name:             "deterministic fast tool",
			isDeterministic:  true,
			isDestructive:    false,
			requiresConfirm:  false,
			estimatedLatency: "fast",
		},
		{
			name:             "destructive slow tool",
			isDeterministic:  true,
			isDestructive:    true,
			requiresConfirm:  true,
			estimatedLatency: "slow",
		},
		{
			name:             "non-deterministic medium tool",
			isDeterministic:  false,
			isDestructive:    false,
			requiresConfirm:  false,
			estimatedLatency: "medium",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tool := &mockFullTool{
				name:             tc.name,
				isDeterministic:  tc.isDeterministic,
				isDestructive:    tc.isDestructive,
				requiresConfirm:  tc.requiresConfirm,
				estimatedLatency: tc.estimatedLatency,
			}

			if tool.IsDeterministic() != tc.isDeterministic {
				t.Errorf("IsDeterministic mismatch: expected %v, got %v",
					tc.isDeterministic, tool.IsDeterministic())
			}
			if tool.IsDestructive() != tc.isDestructive {
				t.Errorf("IsDestructive mismatch: expected %v, got %v",
					tc.isDestructive, tool.IsDestructive())
			}
			if tool.RequiresConfirmation() != tc.requiresConfirm {
				t.Errorf("RequiresConfirmation mismatch: expected %v, got %v",
					tc.requiresConfirm, tool.RequiresConfirmation())
			}
			if tool.EstimatedLatency() != tc.estimatedLatency {
				t.Errorf("EstimatedLatency mismatch: expected %s, got %s",
					tc.estimatedLatency, tool.EstimatedLatency())
			}

			// Verify these are reflected in MCP export
			mcp := tool.ToMCPDefinition()
			if mcp.Annotations["deterministic"] != tc.isDeterministic {
				t.Error("MCP deterministic annotation mismatch")
			}
			if mcp.Annotations["destructive"] != tc.isDestructive {
				t.Error("MCP destructive annotation mismatch")
			}
			if mcp.Annotations["requires_confirmation"] != tc.requiresConfirm {
				t.Error("MCP requires_confirmation annotation mismatch")
			}
			if mcp.Annotations["estimated_latency"] != tc.estimatedLatency {
				t.Error("MCP estimated_latency annotation mismatch")
			}
		})
	}
}
