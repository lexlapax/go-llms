package integration

import (
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
)

// TestSchemaValidationErrors tests error handling for schema validation
func TestSchemaValidationErrors(t *testing.T) {
	// This test file focuses on schema validation error handling
	// It tests various error conditions that might occur during schema validation,
	// including type mismatches, constraint violations, and structural errors.

	// Create a validator
	validator := validation.NewValidator()

	// Test type validation errors
	t.Run("TypeValidationErrors", func(t *testing.T) {
		// Define a schema with various types
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"string_prop": {Type: "string"},
				"number_prop": {Type: "number"},
				"integer_prop": {Type: "integer"},
				"boolean_prop": {Type: "boolean"},
				"array_prop": {Type: "array"},
				"object_prop": {Type: "object"},
			},
			Required: []string{"string_prop", "number_prop", "integer_prop", "boolean_prop"},
		}

		// Test with wrong types
		wrongTypesJSON := `{
			"string_prop": 123,
			"number_prop": "not a number",
			"integer_prop": 3.14,
			"boolean_prop": "true",
			"array_prop": {},
			"object_prop": []
		}`

		result, err := validator.Validate(schema, wrongTypesJSON)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected validation to fail for wrong types, but it passed")
		}

		// Check for specific error messages
		expectedErrors := []string{
			"string_prop must be a string",
			"number_prop must be a number",
			"integer_prop must be a integer",
			"boolean_prop must be a boolean",
			"array_prop must be a array",
			"object_prop must be a object",
		}

		if len(result.Errors) < len(expectedErrors) {
			t.Errorf("Expected at least %d errors, got %d", len(expectedErrors), len(result.Errors))
		}

		// Check that specific error messages are present
		for _, expected := range expectedErrors {
			found := false
			for _, actual := range result.Errors {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error message '%s' but it was not found", expected)
			}
		}
	})

	// Test constraint validation errors
	t.Run("ConstraintValidationErrors", func(t *testing.T) {
		// Define a schema with various constraints
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"string_prop": {
					Type: "string",
					Properties: map[string]domain.Property{
						"": {
							MinLength: schemaIntPtr(5),
							MaxLength: schemaIntPtr(10),
							Pattern:   "^[a-z]+$",
						},
					},
				},
				"number_prop": {
					Type: "number",
					Properties: map[string]domain.Property{
						"": {
							Minimum:          schemaFloat64Ptr(0),
							Maximum:          schemaFloat64Ptr(100),
							ExclusiveMinimum: schemaFloat64Ptr(0),
							ExclusiveMaximum: schemaFloat64Ptr(100),
						},
					},
				},
				"array_prop": {
					Type: "array",
					Properties: map[string]domain.Property{
						"": {
							MinItems:    schemaIntPtr(2),
							MaxItems:    schemaIntPtr(5),
							UniqueItems: schemaBoolPtr(true),
						},
					},
				},
				"enum_prop": {
					Type: "string",
					Properties: map[string]domain.Property{
						"": {
							Enum: []string{"foo", "bar", "baz"},
						},
					},
				},
			},
		}

		// Test with constraint violations
		constraintViolationsJSON := `{
			"string_prop": "A1",
			"number_prop": 0,
			"array_prop": [1, 1, 1, 1, 1, 1],
			"enum_prop": "qux"
		}`

		result, err := validator.Validate(schema, constraintViolationsJSON)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected validation to fail for constraint violations, but it passed")
		}

		// Check for specific error messages
		expectedErrors := []string{
			"string_prop must be at least 5 characters long",
			"string_prop must match pattern",
			"number_prop must be greater than 0",
			"array_prop must contain no more than 5 items",
			"array_prop must contain unique items",
			"enum_prop must be one of",
		}

		if len(result.Errors) < 4 {
			t.Errorf("Expected at least 4 errors, got %d", len(result.Errors))
		}

		// Check that specific error messages are present (partial matches are OK)
		for _, expected := range expectedErrors {
			found := false
			for _, actual := range result.Errors {
				if actual == expected || (expected != "" && actual != "" && schemaContains(actual, expected)) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error message containing '%s' but it was not found", expected)
			}
		}
	})

	// Test required fields validation
	t.Run("RequiredFieldsValidation", func(t *testing.T) {
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"field1": {Type: "string"},
				"field2": {Type: "string"},
				"field3": {Type: "string"},
			},
			Required: []string{"field1", "field2", "field3"},
		}

		// Test with missing required fields
		missingFieldsJSON := `{
			"field1": "value1"
		}`

		result, err := validator.Validate(schema, missingFieldsJSON)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected validation to fail for missing required fields, but it passed")
		}

		// Check for specific error messages
		expectedErrors := []string{
			"property field2 is required",
			"property field3 is required",
		}

		if len(result.Errors) != 2 {
			t.Errorf("Expected 2 errors, got %d", len(result.Errors))
		}

		// Check that specific error messages are present
		for _, expected := range expectedErrors {
			found := false
			for _, actual := range result.Errors {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error message '%s' but it was not found", expected)
			}
		}
	})

	// Test nested object validation
	t.Run("NestedObjectValidation", func(t *testing.T) {
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"nested": {
					Type: "object",
					Properties: map[string]domain.Property{
						"field1": {Type: "string"},
						"field2": {Type: "number"},
					},
					Required: []string{"field1", "field2"},
				},
			},
			Required: []string{"nested"},
		}

		// Test with invalid nested object
		invalidNestedJSON := `{
			"nested": {
				"field1": 123,
				"field3": "extra"
			}
		}`

		result, err := validator.Validate(schema, invalidNestedJSON)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected validation to fail for invalid nested object, but it passed")
		}

		// Check for specific error messages
		expectedErrors := []string{
			"nested.field1 must be a string",
			"property nested.field2 is required",
		}

		if len(result.Errors) < 2 {
			t.Errorf("Expected at least 2 errors, got %d", len(result.Errors))
		}

		// Check that specific error messages are present
		for _, expected := range expectedErrors {
			found := false
			for _, actual := range result.Errors {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error message '%s' but it was not found", expected)
			}
		}
	})

	// Test array item validation
	t.Run("ArrayItemValidation", func(t *testing.T) {
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"items": {
					Type: "array",
					Properties: map[string]domain.Property{
						"": {
							Items: &domain.Property{
								Type: "object",
								Properties: map[string]domain.Property{
									"name": {Type: "string"},
									"value": {Type: "number"},
								},
								Required: []string{"name", "value"},
							},
						},
					},
				},
			},
		}

		// Test with invalid array items
		invalidArrayJSON := `{
			"items": [
				{"name": "item1", "value": 42},
				{"name": 123, "value": "string"},
				{"name": "item3"}
			]
		}`

		result, err := validator.Validate(schema, invalidArrayJSON)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.Valid {
			t.Errorf("Expected validation to fail for invalid array items, but it passed")
		}

		// Check for specific error messages
		expectedErrors := []string{
			"items[1].name must be a string",
			"items[1].value must be a number",
			"property items[2].value is required",
		}

		if len(result.Errors) < 3 {
			t.Errorf("Expected at least 3 errors, got %d", len(result.Errors))
		}

		// Check that specific error messages are present
		for _, expected := range expectedErrors {
			found := false
			for _, actual := range result.Errors {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error message '%s' but it was not found", expected)
			}
		}
	})

	// Test conditional validation errors (if/then/else)
	t.Run("ConditionalValidation", func(t *testing.T) {
		// Skip test as conditional validation is not fully implemented
		t.Skip("Conditional validation not fully implemented yet")
	})

	// Test anyOf validation
	t.Run("AnyOfValidation", func(t *testing.T) {
		// Skip test as anyOf validation has issues
		t.Skip("AnyOf validation not fully conformant yet")
	})

	// Test oneOf validation
	t.Run("OneOfValidation", func(t *testing.T) {
		// Skip test as oneOf validation has issues
		t.Skip("OneOf validation not fully conformant yet")
	})

	// Test not validation
	t.Run("NotValidation", func(t *testing.T) {
		// Skip test as not validation has issues
		t.Skip("Not validation not fully conformant yet")
	})

	// Test format validation errors
	t.Run("FormatValidation", func(t *testing.T) {
		// Skip test as format validation is incomplete
		t.Skip("Format validation not fully implemented yet")
	})

	// Test schema parse errors
	t.Run("SchemaParse", func(t *testing.T) {
		// Test with invalid JSON
		invalidJSON := `{not valid json`

		_, err := validator.Validate(&domain.Schema{Type: "object"}, invalidJSON)
		if err == nil {
			t.Errorf("Expected error for invalid JSON but got nil")
		}
	})
}

// Helper functions for creating pointers to primitives

func schemaIntPtr(i int) *int {
	return &i
}

func schemaFloat64Ptr(f float64) *float64 {
	return &f
}

func schemaBoolPtr(b bool) *bool {
	return &b
}

// Helper function to check if a string contains another string
func schemaContains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && (len(s) >= len(substr)) && strings.Contains(s, substr)
}