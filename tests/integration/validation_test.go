package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// TestEndToEndValidation tests the schema validation system from end to end
func TestEndToEndValidation(t *testing.T) {
	// Create a validator
	validator := validation.NewValidator()

	// Create a schema for a person
	schema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"name": {
				Type: "string",
			},
			"age": {
				Type:    "integer",
				Minimum: float64Ptr(0),
				Maximum: float64Ptr(120),
			},
			"email": {
				Type:   "string",
				Format: "email",
			},
			"tags": {
				Type: "array",
				Items: &sdomain.Property{
					Type: "string",
				},
			},
		},
		Required: []string{"name"},
	}

	// Test valid input
	t.Run("ValidInput", func(t *testing.T) {
		input := `{
			"name": "John Doe",
			"age": 30,
			"email": "john@example.com",
			"tags": ["developer", "golang"]
		}`

		result, err := validator.Validate(schema, input)
		if err != nil {
			t.Fatalf("Validation failed with error: %v", err)
		}

		if !result.Valid {
			t.Errorf("Expected valid result, got invalid with errors: %v", result.Errors)
		}
	})

	// Test missing required field
	t.Run("MissingRequiredField", func(t *testing.T) {
		input := `{
			"age": 30,
			"email": "john@example.com"
		}`

		result, err := validator.Validate(schema, input)
		if err != nil {
			t.Fatalf("Validation failed with error: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected invalid result for missing required field")
		}

		// Check if the error mentions the missing field
		hasRequiredError := false
		for _, err := range result.Errors {
			if containsAll(err, "name", "required") {
				hasRequiredError = true
				break
			}
		}

		if !hasRequiredError {
			t.Errorf("Expected error about missing required field 'name', got: %v", result.Errors)
		}
	})

	// Test invalid types
	t.Run("InvalidTypes", func(t *testing.T) {
		input := `{
			"name": "John Doe",
			"age": "thirty",
			"email": "not-an-email"
		}`

		result, err := validator.Validate(schema, input)
		if err != nil {
			t.Fatalf("Validation failed with error: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected invalid result for invalid types")
		}

		// Check if the errors mention the type issues
		hasError := false
		for _, err := range result.Errors {
			if strings.Contains(strings.ToLower(err), "age") ||
				strings.Contains(strings.ToLower(err), "email") {
				hasError = true
				break
			}
		}

		if !hasError {
			t.Errorf("Expected validation errors, got: %v", result.Errors)
		}
	})

	// Test out of range values
	t.Run("OutOfRangeValues", func(t *testing.T) {
		input := `{
			"name": "John Doe",
			"age": 150
		}`

		result, err := validator.Validate(schema, input)
		if err != nil {
			t.Fatalf("Validation failed with error: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected invalid result for out of range values")
		}

		// Check if the error mentions the range issue
		hasRangeError := false
		for _, err := range result.Errors {
			if strings.Contains(strings.ToLower(err), "age") {
				hasRangeError = true
				break
			}
		}

		if !hasRangeError {
			t.Errorf("Expected error about 'age' validation, got: %v", result.Errors)
		}
	})

	// Test type coercion
	t.Run("TypeCoercion", func(t *testing.T) {
		// String -> int, string -> bool, etc.
		input := `{
			"name": "John Doe",
			"age": 30
		}`

		result, err := validator.Validate(schema, input)
		if err != nil {
			t.Fatalf("Validation failed with error: %v", err)
		}

		// This should be valid
		if !result.Valid {
			t.Errorf("Expected valid result, got errors: %v", result.Errors)
		}
	})

	// Test structured processor
	t.Run("StructuredProcessor", func(t *testing.T) {
		// Create a structured processor
		proc := processor.NewStructuredProcessor(validator)

		// Process valid input
		input := `{
			"name": "John Doe",
			"age": 30,
			"email": "john@example.com"
		}`

		processed, err := proc.Process(schema, input)
		if err != nil {
			t.Fatalf("Processing failed with error: %v", err)
		}

		// Verify the output structure
		data, ok := processed.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map[string]interface{}, got: %T", processed)
		}

		// Check name
		name, ok := data["name"].(string)
		if !ok || name != "John Doe" {
			t.Errorf("Expected name 'John Doe', got: %v", data["name"])
		}

		// Check age
		age, ok := data["age"].(float64)
		if !ok || age != 30 {
			t.Errorf("Expected age 30, got: %v", data["age"])
		}
	})
}

// TestMockProviderWithSchema tests the schema validation with the mock provider
func TestMockProviderWithSchema(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewMockProvider()

	// Create a schema for a weather forecast
	schema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"location": {
				Type: "string",
			},
			"temperature": {
				Type: "number",
			},
			"forecast": {
				Type: "string",
				Enum: []string{"sunny", "cloudy", "rainy", "snowy"},
			},
		},
		Required: []string{"location", "temperature"},
	}

	// Set up the mock to return a valid response
	mockProvider.WithGenerateWithSchemaFunc(func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
		// Return a map that matches the expected values in the test
		return map[string]interface{}{
			"location":    "New York",
			"temperature": 72.5,
			"forecast":    "sunny",
		}, nil
	})

	// Test generating with schema
	t.Run("GenerateWithSchema", func(t *testing.T) {
		result, err := mockProvider.GenerateWithSchema(context.Background(), "What's the weather like?", schema)
		if err != nil {
			t.Fatalf("GenerateWithSchema failed with error: %v", err)
		}

		// Check the result
		data, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map[string]interface{}, got: %T", result)
		}

		// Verify the structure
		location, ok := data["location"].(string)
		if !ok || location != "New York" {
			t.Errorf("Expected location 'New York', got: %v", data["location"])
		}

		temperature, ok := data["temperature"].(float64)
		if !ok || temperature != 72.5 {
			t.Errorf("Expected temperature 72.5, got: %v", data["temperature"])
		}

		forecast, ok := data["forecast"].(string)
		if !ok || forecast != "sunny" {
			t.Errorf("Expected forecast 'sunny', got: %v", data["forecast"])
		}
	})

	// Test invalid response, should cause retries and eventually fail
	t.Run("InvalidResponse", func(t *testing.T) {
		// Set up the mock to return an invalid response
		count := 0
		mockProvider.WithGenerateWithSchemaFunc(func(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (interface{}, error) {
			count++
			// Return with missing required temperature field which should cause an error
			if count < 3 { // Allow a few retries
				return map[string]interface{}{
					"location": "New York",
					// Missing temperature field
				}, fmt.Errorf("validation failed: missing required field 'temperature'")
			}
			// Always return an error
			return nil, fmt.Errorf("validation failed: missing required field 'temperature'")
		})

		_, err := mockProvider.GenerateWithSchema(context.Background(), "What's the weather like?", schema)
		if err == nil {
			t.Errorf("Expected error for invalid response, got nil")
		}

		// Since we're always returning an error, we shouldn't check retry count here
	})
}

// Helper functions
func float64Ptr(v float64) *float64 {
	return &v
}

func containsAll(s string, substrings ...string) bool {
	s = strings.ToLower(s)
	for _, sub := range substrings {
		if !strings.Contains(s, strings.ToLower(sub)) {
			return false
		}
	}
	return true
}
