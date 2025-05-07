package validation

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestOptimizedValidator_SimpleValidation tests basic validation with the optimized validator
func TestOptimizedValidator_SimpleValidation(t *testing.T) {
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
		Required: []string{"name"},
	}

	validator := NewOptimizedValidator()

	// Test valid data
	validJSON := `{"name": "John", "age": 30}`
	result, err := validator.Validate(schema, validJSON)
	if err != nil {
		t.Fatalf("Validation failed with error: %v", err)
	}
	if !result.Valid {
		t.Fatalf("Expected valid result, got invalid with errors: %v", result.Errors)
	}

	// Test missing required field
	invalidJSON := `{"age": 30}`
	result, err = validator.Validate(schema, invalidJSON)
	if err != nil {
		t.Fatalf("Validation failed with error: %v", err)
	}
	if result.Valid {
		t.Fatalf("Expected invalid result for missing required field")
	}
	if len(result.Errors) == 0 {
		t.Fatalf("Expected validation errors for missing required field")
	}

	// Test wrong type
	invalidJSON = `{"name": "John", "age": "thirty"}`
	result, err = validator.Validate(schema, invalidJSON)
	if err != nil {
		t.Fatalf("Validation failed with error: %v", err)
	}
	if result.Valid {
		t.Fatalf("Expected invalid result for wrong type")
	}
	if len(result.Errors) == 0 {
		t.Fatalf("Expected validation errors for wrong type")
	}
}

// TestOptimizedValidator_ComplexValidation tests more complex validations
func TestOptimizedValidator_ComplexValidation(t *testing.T) {
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"name": {
				Type:      "string",
				MinLength: intPtr(3),
				MaxLength: intPtr(10),
				Pattern:   "^[a-zA-Z]+$",
			},
			"age": {
				Type:    "integer",
				Minimum: float64Ptr(18),
				Maximum: float64Ptr(120),
			},
			"email": {
				Type:   "string",
				Format: "email",
			},
			"category": {
				Type: "string",
				Enum: []string{"A", "B", "C"},
			},
		},
		Required: []string{"name", "age"},
	}

	validator := NewOptimizedValidator()

	// Test valid data
	validJSON := `{
		"name": "John",
		"age": 30,
		"email": "john@example.com",
		"category": "A"
	}`
	result, err := validator.Validate(schema, validJSON)
	if err != nil {
		t.Fatalf("Validation failed with error: %v", err)
	}
	if !result.Valid {
		t.Fatalf("Expected valid result, got invalid with errors: %v", result.Errors)
	}

	// Test various constraints
	testCases := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "String too short",
			json:    `{"name": "Jo", "age": 30}`,
			wantErr: true,
		},
		{
			name:    "String too long",
			json:    `{"name": "JohnJacobJingleheimer", "age": 30}`,
			wantErr: true,
		},
		{
			name:    "String pattern violation",
			json:    `{"name": "John123", "age": 30}`,
			wantErr: true,
		},
		{
			name:    "Number too small",
			json:    `{"name": "John", "age": 15}`,
			wantErr: true,
		},
		{
			name:    "Number too large",
			json:    `{"name": "John", "age": 130}`,
			wantErr: true,
		},
		{
			name:    "Invalid email format",
			json:    `{"name": "John", "age": 30, "email": "not-an-email"}`,
			wantErr: true,
		},
		{
			name:    "Invalid enum value",
			json:    `{"name": "John", "age": 30, "category": "D"}`,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.Validate(schema, tc.json)
			if err != nil {
				t.Fatalf("Validation failed with error: %v", err)
			}
			if tc.wantErr && result.Valid {
				t.Fatalf("Expected invalid result, got valid")
			}
			if !tc.wantErr && !result.Valid {
				t.Fatalf("Expected valid result, got invalid with errors: %v", result.Errors)
			}
		})
	}
}

// TestOptimizedValidator_ArrayValidation tests validation of arrays
func TestOptimizedValidator_ArrayValidation(t *testing.T) {
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"items": {
				Type: "array",
				Items: &domain.Property{
					Type: "object",
					Properties: map[string]domain.Property{
						"id":   {Type: "integer"},
						"name": {Type: "string"},
					},
					Required: []string{"id"},
				},
			},
		},
		Required: []string{"items"},
	}

	validator := NewOptimizedValidator()

	// Test valid array
	validJSON := `{
		"items": [
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"}
		]
	}`
	result, err := validator.Validate(schema, validJSON)
	if err != nil {
		t.Fatalf("Validation failed with error: %v", err)
	}
	if !result.Valid {
		t.Fatalf("Expected valid result, got invalid with errors: %v", result.Errors)
	}

	// Test invalid array item
	invalidJSON := `{
		"items": [
			{"id": 1, "name": "Item 1"},
			{"name": "Missing ID"},
			{"id": 3, "name": "Item 3"}
		]
	}`
	result, err = validator.Validate(schema, invalidJSON)
	if err != nil {
		t.Fatalf("Validation failed with error: %v", err)
	}
	if result.Valid {
		t.Fatalf("Expected invalid result for array with invalid item")
	}
	if len(result.Errors) == 0 {
		t.Fatalf("Expected validation errors for array with invalid item")
	}
}

// Note: Using helper functions from validator_test.go