package tools

import (
	"context"
	"fmt"
	"strings"
	"testing"

	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestToolEdgeCases tests edge cases for the tool implementation
func TestToolEdgeCases(t *testing.T) {
	// Test cases for empty/nil parameters
	t.Run("NilParameters", func(t *testing.T) {
		// A function that takes no parameters
		noParamFunc := func() string {
			return "success"
		}
		
		tool := NewTool("noParams", "Test tool with no params", noParamFunc, nil)
		
		// Execute with nil params should succeed
		result, err := tool.Execute(context.Background(), nil)
		if err != nil {
			t.Errorf("Expected success with nil params, got error: %v", err)
		}
		if result != "success" {
			t.Errorf("Expected 'success', got %v", result)
		}
	})
	
	t.Run("RequiredParamsButNil", func(t *testing.T) {
		// A function that requires parameters
		paramFunc := func(name string) string {
			return "Hello, " + name
		}
		
		tool := NewTool("withParams", "Test tool with params", paramFunc, nil)
		
		// Execute with nil params should fail
		_, err := tool.Execute(context.Background(), nil)
		if err == nil {
			t.Errorf("Expected error when providing nil params to function requiring params")
		}
	})

	// Test cases for type mismatches
	t.Run("TypeMismatch", func(t *testing.T) {
		// Function expecting a specific struct
		type UserParams struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		
		userFunc := func(params UserParams) string {
			return fmt.Sprintf("%s is %d years old", params.Name, params.Age)
		}
		
		tool := NewTool("userTool", "Tool for user info", userFunc, nil)
		
		// Test with completely wrong type
		wrongType := 123
		_, err := tool.Execute(context.Background(), wrongType)
		if err == nil {
			t.Errorf("Expected error with wrong parameter type")
		}
		
		// Test with partial match (map missing a field)
		partialMap := map[string]interface{}{
			"name": "John",
			// missing age
		}
		// This should still work with default values
		result, err := tool.Execute(context.Background(), partialMap)
		if err != nil {
			t.Errorf("Expected success with partial map, got error: %v", err)
		}
		if result != "John is 0 years old" {
			t.Errorf("Expected default age of 0, got: %v", result)
		}
	})
	
	// Test with schema validation
	t.Run("SchemaValidation", func(t *testing.T) {
		// Define a schema for parameters
		schema := &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"name": {
					Type:      "string",
					MinLength: intPtr(2),
					MaxLength: intPtr(50),
				},
				"age": {
					Type:    "integer",
					Minimum: float64Ptr(0),
					Maximum: float64Ptr(120),
				},
			},
			Required: []string{"name"},
		}
		
		// Function using the parameters
		userFunc := func(params struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}) string {
			return fmt.Sprintf("%s is %d years old", params.Name, params.Age)
		}
		
		tool := NewTool("userTool", "Tool for user info", userFunc, schema)
		
		// Valid parameters
		validParams := map[string]interface{}{
			"name": "John",
			"age":  30,
		}
		result, err := tool.Execute(context.Background(), validParams)
		if err != nil {
			t.Errorf("Expected success with valid parameters, got error: %v", err)
		}
		if result != "John is 30 years old" {
			t.Errorf("Expected 'John is 30 years old', got %v", result)
		}
		
		// Note: Schema validation isn't fully implemented in this tool yet,
		// but this test case demonstrates how it would be used
	})
	
	// Test function receiving map[string]interface{}
	t.Run("MapParameter", func(t *testing.T) {
		// Function that directly accepts a map
		mapFunc := func(data map[string]interface{}) string {
			if name, ok := data["name"].(string); ok {
				return "Hello, " + name
			}
			return "Hello, unknown"
		}
		
		tool := NewTool("mapTool", "Tool accepting map", mapFunc, nil)
		
		// Test with map
		mapParam := map[string]interface{}{
			"name": "John",
			"age":  30,
		}
		result, err := tool.Execute(context.Background(), mapParam)
		if err != nil {
			t.Errorf("Expected success with map parameter, got error: %v", err)
		}
		if result != "Hello, John" {
			t.Errorf("Expected 'Hello, John', got %v", result)
		}
	})
	
	// Test complex type conversions
	t.Run("ComplexTypeConversions", func(t *testing.T) {
		// Function taking various types
		conversionFunc := func(params struct {
			IntValue    int     `json:"int"`
			FloatValue  float64 `json:"float"`
			StringValue string  `json:"string"`
			BoolValue   bool    `json:"bool"`
		}) string {
			return fmt.Sprintf("int: %d, float: %.1f, string: %s, bool: %t",
				params.IntValue, params.FloatValue, params.StringValue, params.BoolValue)
		}
		
		tool := NewTool("conversionTool", "Tool with type conversions", conversionFunc, nil)
		
		// Test with various string representations that should be converted
		strParams := map[string]interface{}{
			"int":    "42",
			"float":  "3.14",
			"string": 123,       // number to string
			"bool":   "true",
		}
		result, err := tool.Execute(context.Background(), strParams)
		if err != nil {
			t.Errorf("Expected success with string conversions, got error: %v", err)
		}
		expected := "int: 42, float: 3.1, string: 123, bool: true"
		if result != expected {
			t.Errorf("Expected '%s', got '%v'", expected, result)
		}
		
		// Test with mixed number types
		numParams := map[string]interface{}{
			"int":    42.0,      // float to int
			"float":  3,         // int to float
			"string": "hello",
			"bool":   1,         // number to bool
		}
		result, err = tool.Execute(context.Background(), numParams)
		if err != nil {
			t.Errorf("Expected success with number conversions, got error: %v", err)
		}
		expected = "int: 42, float: 3.0, string: hello, bool: true"
		if result != expected {
			t.Errorf("Expected '%s', got '%v'", expected, result)
		}
	})
	
	// Test error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		// Function that returns an error
		errorFunc := func(shouldError bool) (string, error) {
			if shouldError {
				return "", fmt.Errorf("requested error")
			}
			return "success", nil
		}
		
		tool := NewTool("errorTool", "Tool that may return error", errorFunc, nil)
		
		// Test without error
		result, err := tool.Execute(context.Background(), false)
		if err != nil {
			t.Errorf("Expected success without error, got: %v", err)
		}
		if result != "success" {
			t.Errorf("Expected 'success', got %v", result)
		}
		
		// Test with error
		_, err = tool.Execute(context.Background(), true)
		if err == nil {
			t.Errorf("Expected error to be returned, got nil")
		}
		if !strings.Contains(err.Error(), "requested error") {
			t.Errorf("Error message not forwarded correctly, got: %v", err)
		}
	})
}

// Helper function to create integer pointers
func intPtr(i int) *int {
	return &i
}

// Helper function to create float64 pointers
func float64Ptr(f float64) *float64 {
	return &f
}