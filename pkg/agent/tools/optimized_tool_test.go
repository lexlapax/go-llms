package tools

import (
	"context"
	"reflect"
	"testing"
)

// TestOptimizedToolBasicExecution tests the basic execution of an optimized tool
func TestOptimizedToolBasicExecution(t *testing.T) {
	// Create a tool with a simple function that takes a struct parameter
	tool := NewOptimizedTool(
		"add",
		"Add two numbers",
		func(params struct {
			A int `json:"a"`
			B int `json:"b"`
		}) int {
			return params.A + params.B
		},
		nil,
	)

	// Check tool metadata
	if tool.Name() != "add" {
		t.Errorf("Expected name 'add', got '%s'", tool.Name())
	}
	if tool.Description() != "Add two numbers" {
		t.Errorf("Expected description 'Add two numbers', got '%s'", tool.Description())
	}

	// Execute the tool with a map of parameters
	params := map[string]interface{}{
		"a": 5,
		"b": 7,
	}
	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Check the result
	if result != 12 {
		t.Errorf("Expected result 12, got %v", result)
	}
}

// TestOptimizedToolStructParams tests the optimized tool with struct parameters
func TestOptimizedToolStructParams(t *testing.T) {
	// Create a tool with a struct parameter
	tool := NewOptimizedTool(
		"multiply",
		"Multiply two numbers",
		func(params struct {
			A float64 `json:"a"`
			B float64 `json:"b"`
		}) float64 {
			return params.A * params.B
		},
		nil,
	)

	// Test with map parameters
	mapParams := map[string]interface{}{
		"a": 3.5,
		"b": 2.0,
	}

	result, err := tool.Execute(context.Background(), mapParams)
	if err != nil {
		t.Fatalf("Execution with map params failed: %v", err)
	}

	expected := 7.0
	if result != expected {
		t.Errorf("Expected result %v, got %v", expected, result)
	}

	// Test with differently typed parameters that should be converted
	mixedParams := map[string]interface{}{
		"a": "3.5", // string that should be converted to float64
		"b": 2,     // int that should be converted to float64
	}

	result, err = tool.Execute(context.Background(), mixedParams)
	if err != nil {
		t.Fatalf("Execution with mixed params failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected result %v, got %v", expected, result)
	}
}

// TestOptimizedToolContextParam tests the optimized tool with a context parameter
func TestOptimizedToolContextParam(t *testing.T) {
	// Create a tool that accepts context
	tool := NewOptimizedTool(
		"context_tool",
		"A tool that uses context",
		func(ctx context.Context, name string) string {
			// Verify we got a valid context
			if ctx == nil {
				return "context is nil"
			}
			return "Hello, " + name
		},
		nil,
	)

	// Execute with context
	ctx := context.WithValue(context.Background(), "test", "value")
	result, err := tool.Execute(ctx, "World")
	
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result != "Hello, World" {
		t.Errorf("Expected 'Hello, World', got '%v'", result)
	}
}

// Helper functions for tests
func sumInts(nums []int) int {
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return sum
}

func concatStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += "," + strs[i]
	}
	return result
}

// TestOptimizedSimpleTools tests simple tool types
func TestOptimizedSimpleTools(t *testing.T) {
	// Test no parameters
	noParamTool := NewOptimizedTool(
		"no_param",
		"A tool with no parameters",
		func() string {
			return "success"
		},
		nil,
	)

	result, err := noParamTool.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("No param execution failed: %v", err)
	}
	if result != "success" {
		t.Errorf("Expected 'success', got '%v'", result)
	}

	// Test multiple return values
	multiReturnTool := NewOptimizedTool(
		"multi_return",
		"A tool with multiple return values",
		func(x int) (int, error) {
			return x * 2, nil
		},
		nil,
	)

	result, err = multiReturnTool.Execute(context.Background(), 5)
	if err != nil {
		t.Fatalf("Multi-return execution failed: %v", err)
	}
	if result != 10 {
		t.Errorf("Expected 10, got %v", result)
	}
}

// TestOptimizedToolComplexStruct tests a complex struct parameter
func TestOptimizedToolComplexStruct(t *testing.T) {
	// Test complex struct conversion by creating a simpler test
	complexTool := NewOptimizedTool(
		"complex",
		"A tool with complex struct parameters",
		func(params struct {
			Name    string   `json:"name"`
			Age     int      `json:"age"`
			Enabled bool     `json:"enabled"`
		}) map[string]interface{} {
			return map[string]interface{}{
				"processed_name": params.Name,
				"doubled_age":    params.Age * 2,
				"enabled":        params.Enabled,
			}
		},
		nil,
	)

	complexParams := map[string]interface{}{
		"name":    "John",
		"age":     25,
		"enabled": true,
	}

	result, err := complexTool.Execute(context.Background(), complexParams)
	if err != nil {
		t.Fatalf("Complex execution failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map result, got %T", result)
	}

	// Check expected values
	expectedResults := map[string]interface{}{
		"processed_name": "John",
		"doubled_age":    50,
		"enabled":        true,
	}

	for key, expected := range expectedResults {
		actual, exists := resultMap[key]
		if !exists {
			t.Errorf("Missing expected key '%s' in result", key)
			continue
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("For key '%s': expected %v (%T), got %v (%T)", key, expected, expected, actual, actual)
		}
	}
}