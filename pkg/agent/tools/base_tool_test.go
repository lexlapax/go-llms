// ABOUTME: Tests for enhanced base Tool implementation with comprehensive LLM guidance
// ABOUTME: Validates tool metadata, builder pattern, and all interface methods

package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestEnhancedToolImplementation(t *testing.T) {
	t.Run("basic tool with new metadata", func(t *testing.T) {
		// Create a simple function that accepts a struct
		type AddParams struct {
			X int `json:"x"`
			Y int `json:"y"`
		}
		fn := func(params AddParams) int {
			return params.X + params.Y
		}

		// Create schema
		schema := &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"x": {Type: "integer", Description: "First number"},
				"y": {Type: "integer", Description: "Second number"},
			},
			Required: []string{"x", "y"},
		}

		// Create tool with builder pattern
		tool := tools.NewToolBuilder("add", "Adds two numbers").
			WithFunction(fn).
			WithParameterSchema(schema).
			WithOutputSchema(&sdomain.Schema{
				Type:        "integer",
				Description: "Sum of the two numbers",
			}).
			WithUsageInstructions("Use this tool when you need to add two numbers together").
			WithCategory("math").
			WithTags([]string{"arithmetic", "addition"}).
			WithVersion("1.0.0").
			Build()

		// Test basic properties
		if tool.Name() != "add" {
			t.Errorf("Expected name 'add', got '%s'", tool.Name())
		}
		if tool.Description() != "Adds two numbers" {
			t.Errorf("Expected description 'Adds two numbers', got '%s'", tool.Description())
		}
		if tool.Category() != "math" {
			t.Errorf("Expected category 'math', got '%s'", tool.Category())
		}
		if tool.Version() != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", tool.Version())
		}

		// Test execution
		ctx := &domain.ToolContext{Context: context.Background()}
		result, err := tool.Execute(ctx, map[string]interface{}{"x": 5, "y": 3})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != 8 {
			t.Errorf("Expected result 8, got %v", result)
		}
	})

	t.Run("tool with full LLM guidance", func(t *testing.T) {
		// Create calculator function that accepts a struct
		type CalcParams struct {
			Operation string  `json:"operation"`
			A         float64 `json:"a"`
			B         float64 `json:"b"`
		}
		calcFn := func(params CalcParams) (float64, error) {
			switch params.Operation {
			case "add":
				return params.A + params.B, nil
			case "subtract":
				return params.A - params.B, nil
			case "multiply":
				return params.A * params.B, nil
			case "divide":
				if params.B == 0 {
					return 0, errors.New("division_by_zero")
				}
				return params.A / params.B, nil
			default:
				return 0, errors.New("invalid_operation")
			}
		}

		// Create comprehensive tool
		tool := tools.NewToolBuilder("calculator", "Performs basic arithmetic operations").
			WithFunction(calcFn).
			WithParameterSchema(&sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"operation": {
						Type:        "string",
						Description: "The operation to perform",
						Enum:        []string{"add", "subtract", "multiply", "divide"},
					},
					"a": {
						Type:        "number",
						Description: "First operand",
					},
					"b": {
						Type:        "number",
						Description: "Second operand",
					},
				},
				Required: []string{"operation", "a", "b"},
			}).
			WithOutputSchema(&sdomain.Schema{
				Type:        "number",
				Description: "Result of the arithmetic operation",
			}).
			WithUsageInstructions(`Use this calculator tool when you need to:
- Perform basic arithmetic operations (add, subtract, multiply, divide)
- Calculate numeric results from user queries
- Process mathematical expressions step by step`).
			WithExamples([]domain.ToolExample{
				{
					Name:        "Basic addition",
					Description: "Add two numbers together",
					Scenario:    "User asks: What is 15 plus 27?",
					Input:       map[string]interface{}{"operation": "add", "a": 15, "b": 27},
					Output:      42.0,
					Explanation: "Simple addition of two numbers",
				},
				{
					Name:        "Division example",
					Description: "Divide one number by another",
					Scenario:    "User asks: What is 100 divided by 4?",
					Input:       map[string]interface{}{"operation": "divide", "a": 100, "b": 4},
					Output:      25.0,
					Explanation: "Division operation with non-zero divisor",
				},
			}).
			WithConstraints([]string{
				"Only supports basic arithmetic operations",
				"Cannot handle complex expressions directly",
				"Division by zero returns an error",
			}).
			WithErrorGuidance(map[string]string{
				"division_by_zero":  "Cannot divide by zero. Please check the divisor value.",
				"invalid_operation": "Operation must be one of: add, subtract, multiply, divide",
			}).
			WithCategory("math").
			WithTags([]string{"arithmetic", "calculator", "math"}).
			WithVersion("2.0.0").
			WithBehavior(true, false, false, "fast").
			Build()

		// Test all metadata methods
		if tool.UsageInstructions() == "" {
			t.Error("Expected usage instructions")
		}
		if len(tool.Examples()) != 2 {
			t.Errorf("Expected 2 examples, got %d", len(tool.Examples()))
		}
		if len(tool.Constraints()) != 3 {
			t.Errorf("Expected 3 constraints, got %d", len(tool.Constraints()))
		}
		if len(tool.ErrorGuidance()) != 2 {
			t.Errorf("Expected 2 error guidance entries, got %d", len(tool.ErrorGuidance()))
		}

		// Test behavioral metadata
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

		// Test execution success case
		ctx := &domain.ToolContext{Context: context.Background()}
		result, err := tool.Execute(ctx, map[string]interface{}{
			"operation": "multiply",
			"a":         7.0,
			"b":         6.0,
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != 42.0 {
			t.Errorf("Expected result 42.0, got %v", result)
		}

		// Test error case
		_, err = tool.Execute(ctx, map[string]interface{}{
			"operation": "divide",
			"a":         10.0,
			"b":         0.0,
		})
		if err == nil {
			t.Error("Expected division by zero error")
		}
	})

	t.Run("tool with MCP export", func(t *testing.T) {
		fn := func(text string) string {
			return "Processed: " + text
		}

		tool := tools.NewToolBuilder("text_processor", "Processes text input").
			WithFunction(fn).
			WithParameterSchema(&sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"text": {
						Type:        "string",
						Description: "Text to process",
					},
				},
				Required: []string{"text"},
			}).
			WithUsageInstructions("Use this to process text").
			WithCategory("text").
			WithTags([]string{"text", "processing"}).
			WithVersion("1.0.0").
			WithBehavior(true, false, false, "fast").
			Build()

		// Test MCP export
		mcp := tool.ToMCPDefinition()
		if mcp.Name != "text_processor" {
			t.Errorf("MCP name mismatch: expected 'text_processor', got '%s'", mcp.Name)
		}
		if mcp.Description != "Processes text input" {
			t.Error("MCP description mismatch")
		}
		if mcp.InputSchema == nil {
			t.Error("MCP input schema should not be nil")
		}

		// Check annotations
		if mcp.Annotations == nil {
			t.Error("MCP annotations should not be nil")
		}
		if mcp.Annotations["category"] != "text" {
			t.Errorf("Expected category 'text', got '%v'", mcp.Annotations["category"])
		}
		if mcp.Annotations["version"] != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%v'", mcp.Annotations["version"])
		}
		if mcp.Annotations["deterministic"] != true {
			t.Error("Expected deterministic annotation to be true")
		}
	})

	t.Run("tool with context support", func(t *testing.T) {
		// Test function that uses ToolContext
		type GreetParams struct {
			Msg string `json:"msg"`
		}
		contextFn := func(ctx *domain.ToolContext, params GreetParams) string {
			if ctx.State != nil {
				if greeting, ok := ctx.State.Get("greeting"); ok {
					return greeting.(string) + " " + params.Msg
				}
			}
			return "Hello " + params.Msg
		}

		tool := tools.NewToolBuilder("greeter", "Greets with context").
			WithFunction(contextFn).
			WithParameterSchema(&sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"msg": {Type: "string", Description: "Message to append"},
				},
				Required: []string{"msg"},
			}).
			Build()

		// Test with context
		state := domain.NewState()
		state.Set("greeting", "Hi")
		ctx := &domain.ToolContext{
			Context: context.Background(),
			State:   state,
		}

		result, err := tool.Execute(ctx, map[string]interface{}{"msg": "World"})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != "Hi World" {
			t.Errorf("Expected 'Hi World', got '%v'", result)
		}
	})

	t.Run("destructive tool with confirmation", func(t *testing.T) {
		deleteFn := func(path string) error {
			if path == "" {
				return errors.New("empty_path")
			}
			// Simulate deletion
			return nil
		}

		tool := tools.NewToolBuilder("file_delete", "Deletes a file").
			WithFunction(deleteFn).
			WithParameterSchema(&sdomain.Schema{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"path": {
						Type:        "string",
						Description: "Path to file to delete",
					},
				},
				Required: []string{"path"},
			}).
			WithUsageInstructions("Use with caution - this permanently deletes files").
			WithConstraints([]string{
				"Cannot be undone",
				"Requires valid file path",
				"User confirmation recommended",
			}).
			WithErrorGuidance(map[string]string{
				"empty_path": "Path cannot be empty",
			}).
			WithCategory("file").
			WithBehavior(true, true, true, "fast").
			Build()

		// Test destructive flags
		if !tool.IsDestructive() {
			t.Error("Expected destructive tool")
		}
		if !tool.RequiresConfirmation() {
			t.Error("Expected confirmation required")
		}

		// Test MCP includes destructive metadata
		mcp := tool.ToMCPDefinition()
		if mcp.Annotations["destructive"] != true {
			t.Error("MCP should include destructive annotation")
		}
		if mcp.Annotations["requires_confirmation"] != true {
			t.Error("MCP should include requires_confirmation annotation")
		}
	})

	t.Run("validation of tool metadata", func(t *testing.T) {
		// Test validation catches missing required fields
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for missing function")
			}
		}()

		// This should panic because no function is set
		_ = tools.NewToolBuilder("invalid", "Invalid tool").Build()
	})
}

// TestToolBuilderEdgeCases tests edge cases and error scenarios
func TestToolBuilderEdgeCases(t *testing.T) {
	t.Run("minimal tool creation", func(t *testing.T) {
		fn := func() string { return "minimal" }

		// Create minimal tool with just required fields
		tool := tools.NewToolBuilder("minimal", "Minimal tool").
			WithFunction(fn).
			Build()

		// Should have default values for optional fields
		if tool.Category() != "" {
			t.Errorf("Expected empty category, got '%s'", tool.Category())
		}
		if tool.Version() != "1.0.0" {
			t.Errorf("Expected default version '1.0.0', got '%s'", tool.Version())
		}
		if !tool.IsDeterministic() {
			t.Error("Expected deterministic by default")
		}
		if tool.IsDestructive() {
			t.Error("Expected non-destructive by default")
		}
		if tool.EstimatedLatency() != "medium" {
			t.Errorf("Expected default latency 'medium', got '%s'", tool.EstimatedLatency())
		}
	})

	t.Run("tool with nil schemas", func(t *testing.T) {
		fn := func(x interface{}) interface{} { return x }

		tool := tools.NewToolBuilder("echo", "Echoes input").
			WithFunction(fn).
			WithParameterSchema(nil). // Explicitly set nil
			WithOutputSchema(nil).    // Explicitly set nil
			Build()

		// Should handle nil schemas gracefully
		if tool.ParameterSchema() != nil {
			t.Error("Expected nil parameter schema")
		}
		if tool.OutputSchema() != nil {
			t.Error("Expected nil output schema")
		}

		// Should still execute
		ctx := &domain.ToolContext{Context: context.Background()}
		result, err := tool.Execute(ctx, "test")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != "test" {
			t.Errorf("Expected 'test', got '%v'", result)
		}
	})

	t.Run("tool with empty metadata", func(t *testing.T) {
		fn := func() {}

		tool := tools.NewToolBuilder("empty", "Empty tool").
			WithFunction(fn).
			WithUsageInstructions("").   // Empty string
			WithExamples(nil).           // Nil examples
			WithConstraints([]string{}). // Empty constraints
			WithErrorGuidance(nil).      // Nil error guidance
			WithTags([]string{}).        // Empty tags
			Build()

		// Should handle empty metadata gracefully
		if tool.UsageInstructions() != "" {
			t.Error("Expected empty usage instructions")
		}
		if tool.Examples() != nil {
			t.Error("Expected nil examples")
		}
		if len(tool.Constraints()) != 0 {
			t.Error("Expected empty constraints")
		}
		if tool.ErrorGuidance() != nil {
			t.Error("Expected nil error guidance")
		}
		if len(tool.Tags()) != 0 {
			t.Error("Expected empty tags")
		}
	})

	t.Run("legacy NewTool still works", func(t *testing.T) {
		// Test that the old NewTool function still works
		type LegacyParams struct {
			A int `json:"a"`
			B int `json:"b"`
		}
		fn := func(params LegacyParams) int { return params.A + params.B }
		schema := &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"a": {Type: "integer"},
				"b": {Type: "integer"},
			},
		}

		tool := tools.NewTool("legacy_add", "Legacy addition", fn, schema)

		// Should have basic functionality
		if tool.Name() != "legacy_add" {
			t.Errorf("Expected name 'legacy_add', got '%s'", tool.Name())
		}
		if tool.Description() != "Legacy addition" {
			t.Errorf("Expected description 'Legacy addition', got '%s'", tool.Description())
		}

		// Should have default values for new fields
		if tool.Version() != "1.0.0" {
			t.Errorf("Expected default version '1.0.0', got '%s'", tool.Version())
		}
		if !tool.IsDeterministic() {
			t.Error("Expected deterministic by default")
		}

		// Test execution
		ctx := &domain.ToolContext{Context: context.Background()}
		result, err := tool.Execute(ctx, map[string]interface{}{"a": 3, "b": 4})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != 7 {
			t.Errorf("Expected result 7, got %v", result)
		}
	})
}
