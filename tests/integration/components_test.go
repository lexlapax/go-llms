package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
)

// TestComponentsIntegration tests that components work together correctly
// in realistic usage scenarios. This test was previously called TestOptimizedComponentsIntegration
// but since the optimized implementations are now the standard implementation, it's been renamed.
func TestComponentsIntegration(t *testing.T) {
	// Create validator
	validator := validation.NewValidator()

	// Create schema for user info
	userSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"name":  {Type: "string", MinLength: intPtrOpt(3)},
			"age":   {Type: "integer", Minimum: float64PtrOpt(18)},
			"email": {Type: "string", Format: "email"},
			"tags": {
				Type: "array",
				Items: &sdomain.Property{
					Type: "string",
				},
			},
		},
		Required: []string{"name", "age", "email"},
	}

	// Create a function that uses validated user data
	processUser := func(userData map[string]interface{}) (string, error) {
		// Validate the user data
		result, err := validator.Validate(userSchema, marshal(userData))
		if err != nil {
			return "", err
		}
		if !result.Valid {
			return "", errorFor("Invalid user data: " + result.Errors[0])
		}

		// Process the validated data
		name := userData["name"].(string)
		// Handle both int and float64 since JSON unmarshaling could produce either
		var age int
		switch v := userData["age"].(type) {
		case float64:
			age = int(v)
		case int:
			age = v
		}
		email := userData["email"].(string)

		// Include tags if present
		tags := ""
		if tagsArr, ok := userData["tags"].([]interface{}); ok && len(tagsArr) > 0 {
			tags = " with tags: "
			for i, tag := range tagsArr {
				if i > 0 {
					tags += ", "
				}
				tags += tag.(string)
			}
		}

		return name + " (" + email + ") is " + fmt.Sprintf("%d", age) + " years old" + tags, nil
	}

	// Create tool with the user processing function
	userTool := tools.NewTool("processUser", "Process user information", processUser, userSchema)

	// Test with valid user data
	t.Run("ValidUserData", func(t *testing.T) {
		userData := map[string]interface{}{
			"name":  "John Doe",
			"age":   30,
			"email": "john@example.com",
			"tags":  []interface{}{"customer", "premium"},
		}

		result, err := userTool.Execute(context.Background(), userData)
		if err != nil {
			t.Fatalf("Error executing tool with valid data: %v", err)
		}

		expected := "John Doe (john@example.com) is 30 years old with tags: customer, premium"
		if result != expected {
			t.Errorf("Expected result '%s', got '%v'", expected, result)
		}
	})

	// Test with missing required field
	t.Run("MissingRequiredField", func(t *testing.T) {
		userData := map[string]interface{}{
			"name": "John Doe",
			// Missing "age" field
			"email": "john@example.com",
		}

		_, err := userTool.Execute(context.Background(), userData)
		if err == nil {
			t.Fatalf("Expected error for missing required field, got none")
		}
	})

	// Test with invalid field values
	t.Run("InvalidFieldValues", func(t *testing.T) {
		userData := map[string]interface{}{
			"name":  "Jo",            // Too short
			"age":   16,              // Below minimum
			"email": "invalid-email", // Invalid format
		}

		_, err := userTool.Execute(context.Background(), userData)
		if err == nil {
			t.Fatalf("Expected error for invalid field values, got none")
		}
	})

	// Test integration between validator and tool
	t.Run("SimpleIntegration", func(t *testing.T) {
		// Create a composite function that uses both validator and tool
		validateAndProcess := func(jsonData string) (string, error) {
			// First validate with the validator
			validationResult, err := validator.Validate(userSchema, jsonData)
			if err != nil {
				return "", err
			}
			if !validationResult.Valid {
				return "", fmt.Errorf("validation failed: %v", validationResult.Errors)
			}

			// Parse the JSON to a map
			var userData map[string]interface{}
			if err := json.Unmarshal([]byte(jsonData), &userData); err != nil {
				return "", err
			}

			// Process with the tool
			toolResult, err := userTool.Execute(context.Background(), userData)
			if err != nil {
				return "", err
			}

			return toolResult.(string), nil
		}

		// Valid JSON
		validJSON := `{
			"name": "John Doe",
			"age": 30,
			"email": "john@example.com",
			"tags": ["customer", "premium"]
		}`

		// Process the valid data
		result, err := validateAndProcess(validJSON)
		if err != nil {
			t.Fatalf("Error with valid data: %v", err)
		}

		expected := "John Doe (john@example.com) is 30 years old with tags: customer, premium"
		if result != expected {
			t.Errorf("Expected result '%s', got '%s'", expected, result)
		}

		// Invalid JSON (missing required field)
		invalidJSON := `{
			"name": "John Doe",
			"email": "john@example.com"
		}`

		// Should fail validation
		_, err = validateAndProcess(invalidJSON)
		if err == nil {
			t.Errorf("Expected error for missing required field, got none")
		}

		// Invalid JSON (format error)
		invalidFormatJSON := `{
			"name": "John Doe",
			"age": 30,
			"email": "not-an-email"
		}`

		// Should fail validation
		_, err = validateAndProcess(invalidFormatJSON)
		if err == nil {
			t.Errorf("Expected error for invalid email format, got none")
		}
	})

	// Test performance with repeated calls
	t.Run("PerformanceWithRepeatedCalls", func(t *testing.T) {
		// Skip in short mode
		if testing.Short() {
			t.Skip("Skipping performance test in short mode")
		}

		userData := map[string]interface{}{
			"name":  "John Doe",
			"age":   30,
			"email": "john@example.com",
			"tags":  []interface{}{"customer", "premium"},
		}

		// Make repeated calls to exercise object pooling and caching
		for i := 0; i < 1000; i++ {
			_, err := userTool.Execute(context.Background(), userData)
			if err != nil {
				t.Fatalf("Error on iteration %d: %v", i, err)
			}
		}
	})
}

// Helper functions

// intPtrOpt returns a pointer to an int value
func intPtrOpt(v int) *int {
	return &v
}

// float64PtrOpt returns a pointer to a float64 value
func float64PtrOpt(v float64) *float64 {
	return &v
}

// errorFor creates a custom error for testing
func errorFor(msg string) error {
	return &testError{msg}
}

// testError is a custom error for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// marshal marshals an interface to JSON string
func marshal(v interface{}) string {
	jsonBytes, _ := json.Marshal(v)
	return string(jsonBytes)
}
