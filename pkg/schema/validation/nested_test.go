package validation

import (
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestNestedObjectValidation(t *testing.T) {
	// Define a schema with nested objects
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"name": {Type: "string"},
			"address": {
				Type: "object",
				Properties: map[string]domain.Property{
					"street":  {Type: "string"},
					"city":    {Type: "string"},
					"zipCode": {Type: "string", Pattern: "^\\d{5}(-\\d{4})?$"},
				},
				Required: []string{"street", "city"},
			},
			"contacts": {
				Type: "array",
				Items: &domain.Property{
					Type: "object",
					Properties: map[string]domain.Property{
						"type":  {Type: "string", Enum: []string{"email", "phone", "social"}},
						"value": {Type: "string"},
					},
					Required: []string{"type", "value"},
				},
			},
		},
		Required: []string{"name", "address"},
	}

	validator := NewValidator()

	// Helper function to check if an error array contains a specific error
	localContainsError := func(errors []string, field, errType string) bool {
		for _, err := range errors {
			if strings.Contains(strings.ToLower(err), strings.ToLower(field)) &&
				strings.Contains(strings.ToLower(err), strings.ToLower(errType)) {
				return true
			}
		}
		return false
	}

	// Valid case
	t.Run("valid nested object", func(t *testing.T) {
		input := `{
			"name": "John Doe",
			"address": {
				"street": "123 Main St",
				"city": "Anytown",
				"zipCode": "12345"
			},
			"contacts": [
				{"type": "email", "value": "john@example.com"},
				{"type": "phone", "value": "555-1234"}
			]
		}`

		result, err := validator.Validate(schema, input)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
		}
	})

	// Invalid case - nested object missing required field
	t.Run("invalid nested object - missing required field", func(t *testing.T) {
		input := `{
			"name": "John Doe",
			"address": {
				"street": "123 Main St"
			},
			"contacts": [
				{"type": "email", "value": "john@example.com"}
			]
		}`

		result, err := validator.Validate(schema, input)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for missing required field in nested object")
		}
		if !localContainsError(result.Errors, "address.city", "required") {
			t.Errorf("Expected 'required' error for 'address.city', got: %v", result.Errors)
		}
	})

	// Invalid case - nested object property constraint violation
	t.Run("invalid nested object - property constraint violation", func(t *testing.T) {
		input := `{
			"name": "John Doe",
			"address": {
				"street": "123 Main St",
				"city": "Anytown",
				"zipCode": "invalid-zip"
			},
			"contacts": [
				{"type": "email", "value": "john@example.com"}
			]
		}`

		result, err := validator.Validate(schema, input)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for pattern constraint violation in nested object")
		}
		if !localContainsError(result.Errors, "address.zipCode", "pattern") {
			t.Errorf("Expected 'pattern' error for 'address.zipCode', got: %v", result.Errors)
		}
	})

	// Invalid case - nested array item validation
	t.Run("invalid nested array item", func(t *testing.T) {
		input := `{
			"name": "John Doe",
			"address": {
				"street": "123 Main St",
				"city": "Anytown",
				"zipCode": "12345"
			},
			"contacts": [
				{"type": "email", "value": "john@example.com"},
				{"type": "invalid-type", "value": "555-1234"}
			]
		}`

		result, err := validator.Validate(schema, input)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for enum constraint violation in array item")
		}
		if !localContainsError(result.Errors, "contacts[1].type", "one of") {
			t.Errorf("Expected 'enum' error for 'contacts[1].type', got: %v", result.Errors)
		}
	})

	// Invalid case - missing required field in array item
	t.Run("invalid nested array item - missing required field", func(t *testing.T) {
		input := `{
			"name": "John Doe",
			"address": {
				"street": "123 Main St",
				"city": "Anytown",
				"zipCode": "12345"
			},
			"contacts": [
				{"type": "email", "value": "john@example.com"},
				{"type": "phone"}
			]
		}`

		result, err := validator.Validate(schema, input)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for missing required field in array item")
		}
		if !localContainsError(result.Errors, "contacts[1].value", "required") {
			t.Errorf("Expected 'required' error for 'contacts[1].value', got: %v", result.Errors)
		}
	})
}
