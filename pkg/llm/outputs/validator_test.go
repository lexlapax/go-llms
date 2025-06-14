package outputs

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_ValidateString(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	testCases := []struct {
		name    string
		value   interface{}
		schema  *OutputSchema
		valid   bool
		errCode string
	}{
		{
			name:  "Valid string",
			value: "test",
			schema: &OutputSchema{
				Type: TypeString,
			},
			valid: true,
		},
		{
			name:  "Invalid type",
			value: 123,
			schema: &OutputSchema{
				Type: TypeString,
			},
			valid:   false,
			errCode: "type_mismatch",
		},
		{
			name:  "Valid enum value",
			value: "red",
			schema: &OutputSchema{
				Type: TypeString,
				Enum: []string{"red", "green", "blue"},
			},
			valid: true,
		},
		{
			name:  "Invalid enum value",
			value: "yellow",
			schema: &OutputSchema{
				Type: TypeString,
				Enum: []string{"red", "green", "blue"},
			},
			valid:   false,
			errCode: "enum_violation",
		},
		{
			name:  "Valid email format",
			value: "test@example.com",
			schema: &OutputSchema{
				Type:   TypeString,
				Format: "email",
			},
			valid: true,
		},
		{
			name:  "Invalid email format",
			value: "not-an-email",
			schema: &OutputSchema{
				Type:   TypeString,
				Format: "email",
			},
			valid:   false,
			errCode: "format_violation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.Validate(ctx, tc.value, tc.schema)
			require.NoError(t, err)
			assert.Equal(t, tc.valid, result.Valid)

			if !tc.valid && tc.errCode != "" {
				require.NotEmpty(t, result.Errors)
				assert.Equal(t, tc.errCode, result.Errors[0].Code)
			}
		})
	}
}

func TestValidator_ValidateNumber(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	min := 0.0
	max := 100.0

	testCases := []struct {
		name    string
		value   interface{}
		schema  *OutputSchema
		valid   bool
		errCode string
	}{
		{
			name:  "Valid number",
			value: 42.5,
			schema: &OutputSchema{
				Type: TypeNumber,
			},
			valid: true,
		},
		{
			name:  "Valid integer as number",
			value: 42,
			schema: &OutputSchema{
				Type: TypeNumber,
			},
			valid: true,
		},
		{
			name:  "Number within range",
			value: 50.0,
			schema: &OutputSchema{
				Type:    TypeNumber,
				Minimum: &min,
				Maximum: &max,
			},
			valid: true,
		},
		{
			name:  "Number below minimum",
			value: -10.0,
			schema: &OutputSchema{
				Type:    TypeNumber,
				Minimum: &min,
			},
			valid:   false,
			errCode: "minimum_violation",
		},
		{
			name:  "Number above maximum",
			value: 150.0,
			schema: &OutputSchema{
				Type:    TypeNumber,
				Maximum: &max,
			},
			valid:   false,
			errCode: "maximum_violation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.Validate(ctx, tc.value, tc.schema)
			require.NoError(t, err)
			assert.Equal(t, tc.valid, result.Valid)

			if !tc.valid && tc.errCode != "" {
				require.NotEmpty(t, result.Errors)
				assert.Equal(t, tc.errCode, result.Errors[0].Code)
			}
		})
	}
}

func TestValidator_ValidateObject(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	required := true
	additionalProps := false

	testCases := []struct {
		name     string
		value    interface{}
		schema   *OutputSchema
		valid    bool
		errCount int
	}{
		{
			name: "Valid object with all required properties",
			value: map[string]interface{}{
				"name":  "John",
				"age":   30,
				"email": "john@example.com",
			},
			schema: &OutputSchema{
				Type: TypeObject,
				Properties: map[string]*OutputSchema{
					"name": {
						Type:     TypeString,
						Required: &required,
					},
					"age": {
						Type: TypeInteger,
					},
					"email": {
						Type:   TypeString,
						Format: "email",
					},
				},
				RequiredProperties: []string{"name"},
			},
			valid: true,
		},
		{
			name: "Missing required property",
			value: map[string]interface{}{
				"age": 30,
			},
			schema: &OutputSchema{
				Type: TypeObject,
				Properties: map[string]*OutputSchema{
					"name": {
						Type:     TypeString,
						Required: &required,
					},
					"age": {
						Type: TypeInteger,
					},
				},
				RequiredProperties: []string{"name"},
			},
			valid:    false,
			errCount: 1,
		},
		{
			name: "Additional properties not allowed",
			value: map[string]interface{}{
				"name":  "John",
				"extra": "not allowed",
			},
			schema: &OutputSchema{
				Type: TypeObject,
				Properties: map[string]*OutputSchema{
					"name": {
						Type: TypeString,
					},
				},
				AdditionalProperties: &additionalProps,
			},
			valid:    true, // Additional properties generate warnings, not errors
			errCount: 0,
		},
		{
			name: "Nested object validation",
			value: map[string]interface{}{
				"person": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
			schema: &OutputSchema{
				Type: TypeObject,
				Properties: map[string]*OutputSchema{
					"person": {
						Type: TypeObject,
						Properties: map[string]*OutputSchema{
							"name": {
								Type: TypeString,
							},
							"age": {
								Type: TypeInteger,
							},
						},
					},
				},
			},
			valid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.Validate(ctx, tc.value, tc.schema)
			require.NoError(t, err)
			assert.Equal(t, tc.valid, result.Valid)

			if tc.errCount > 0 {
				assert.Len(t, result.Errors, tc.errCount)
			}
		})
	}
}

func TestValidator_ValidateArray(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	minItems := 2
	maxItems := 5

	testCases := []struct {
		name    string
		value   interface{}
		schema  *OutputSchema
		valid   bool
		errCode string
	}{
		{
			name:  "Valid array of strings",
			value: []interface{}{"a", "b", "c"},
			schema: &OutputSchema{
				Type: TypeArray,
				Items: &OutputSchema{
					Type: TypeString,
				},
			},
			valid: true,
		},
		{
			name:  "Array with correct length",
			value: []interface{}{1, 2, 3},
			schema: &OutputSchema{
				Type:     TypeArray,
				MinItems: &minItems,
				MaxItems: &maxItems,
			},
			valid: true,
		},
		{
			name:  "Array too short",
			value: []interface{}{1},
			schema: &OutputSchema{
				Type:     TypeArray,
				MinItems: &minItems,
			},
			valid:   false,
			errCode: "min_items_violation",
		},
		{
			name:  "Array too long",
			value: []interface{}{1, 2, 3, 4, 5, 6},
			schema: &OutputSchema{
				Type:     TypeArray,
				MaxItems: &maxItems,
			},
			valid:   false,
			errCode: "max_items_violation",
		},
		{
			name:  "Array with invalid item type",
			value: []interface{}{"a", "b", 123},
			schema: &OutputSchema{
				Type: TypeArray,
				Items: &OutputSchema{
					Type: TypeString,
				},
			},
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.Validate(ctx, tc.value, tc.schema)
			require.NoError(t, err)
			assert.Equal(t, tc.valid, result.Valid)

			if !tc.valid && tc.errCode != "" {
				require.NotEmpty(t, result.Errors)
				assert.Equal(t, tc.errCode, result.Errors[0].Code)
			}
		})
	}
}

func TestValidator_GenerateSuggestions(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	schema := &OutputSchema{
		Type: TypeObject,
		Properties: map[string]*OutputSchema{
			"name": {
				Type: TypeString,
			},
			"email": {
				Type:   TypeString,
				Format: "email",
			},
			"status": {
				Type: TypeString,
				Enum: []string{"active", "inactive"},
			},
		},
		RequiredProperties: []string{"name"},
	}

	value := map[string]interface{}{
		// Missing required "name"
		"email":  "not-an-email",
		"status": "pending", // Not in enum
	}

	result, err := validator.Validate(ctx, value, schema)
	require.NoError(t, err)
	assert.False(t, result.Valid)

	// Check suggestions
	assert.NotEmpty(t, result.Suggestions)

	// Should have suggestions for:
	// - Missing required field
	// - Invalid email format
	// - Invalid enum value
	suggestionTypes := make(map[string]bool)
	for _, s := range result.Suggestions {
		if s.Description == "Add the missing required field" {
			suggestionTypes["required"] = true
		}
		if strings.Contains(s.Description, "format") {
			suggestionTypes["format"] = true
		}
		if s.Description == "Use one of the allowed values" {
			suggestionTypes["enum"] = true
		}
	}

	assert.True(t, suggestionTypes["required"])
	assert.True(t, suggestionTypes["format"])
	assert.True(t, suggestionTypes["enum"])
}

func TestValidator_CustomRules(t *testing.T) {
	ctx := context.Background()
	validator := NewValidator()

	// Add custom rule that requires strings to be uppercase
	validator.AddCustomRule("uppercase", func(path string, value interface{}, schema *OutputSchema) *ValidationError {
		if schema.Type != TypeString {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return nil
		}

		if str != strings.ToUpper(str) {
			return &ValidationError{
				Path:     path,
				Message:  "string must be uppercase",
				Code:     "custom.uppercase",
				Expected: "uppercase string",
				Actual:   str,
			}
		}

		return nil
	})

	schema := &OutputSchema{
		Type: TypeString,
	}

	// Test with lowercase string
	result, err := validator.Validate(ctx, "hello", schema)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)
	assert.Equal(t, "custom.uppercase", result.Errors[0].Code)

	// Test with uppercase string
	result, err = validator.Validate(ctx, "HELLO", schema)
	require.NoError(t, err)
	assert.True(t, result.Valid)
}
