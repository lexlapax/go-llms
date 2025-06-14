// ABOUTME: Output validator for verifying parsed outputs against schemas
// ABOUTME: Provides detailed validation results with error information and fix suggestions

package outputs

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// ValidationResult contains the result of validation
type ValidationResult struct {
	// Valid indicates if the output is valid
	Valid bool

	// Errors contains validation errors
	Errors []ValidationError

	// Warnings contains validation warnings
	Warnings []ValidationWarning

	// Suggestions contains fix suggestions
	Suggestions []FixSuggestion
}

// ValidationError represents a validation error
type ValidationError struct {
	// Path is the JSON path to the error location
	Path string

	// Field is the field name that failed validation
	Field string

	// Message describes the error
	Message string

	// Code is an error code for programmatic handling
	Code string

	// Expected describes what was expected
	Expected string

	// Actual describes what was found
	Actual string
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	// Path is the JSON path to the warning location
	Path string

	// Message describes the warning
	Message string

	// Code is a warning code
	Code string
}

// FixSuggestion represents a suggested fix
type FixSuggestion struct {
	// Path is the JSON path to apply the fix
	Path string

	// Description describes the fix
	Description string

	// Fix is the suggested fix action
	Fix string

	// Example shows an example of the fix
	Example string
}

// Validator validates outputs against schemas
type Validator struct {
	// customRules holds custom validation rules
	customRules map[string]ValidationRule
}

// ValidationRule defines a custom validation rule
type ValidationRule func(path string, value interface{}, schema *OutputSchema) *ValidationError

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		customRules: make(map[string]ValidationRule),
	}
}

// AddCustomRule adds a custom validation rule
func (v *Validator) AddCustomRule(name string, rule ValidationRule) {
	v.customRules[name] = rule
}

// Validate validates output against a schema
func (v *Validator) Validate(ctx context.Context, output interface{}, schema *OutputSchema) (*ValidationResult, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema cannot be nil")
	}

	result := &ValidationResult{
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationWarning{},
		Suggestions: []FixSuggestion{},
	}

	// Validate the output
	v.validateValue("$", output, schema, result)

	// Apply custom rules
	for name, rule := range v.customRules {
		if err := rule("$", output, schema); err != nil {
			err.Code = "custom." + name
			result.Errors = append(result.Errors, *err)
			result.Valid = false
		}
	}

	// Generate fix suggestions based on errors
	v.generateSuggestions(result)

	return result, nil
}

// validateValue recursively validates a value against a schema
func (v *Validator) validateValue(path string, value interface{}, schema *OutputSchema, result *ValidationResult) {
	// Check required
	if schema.Required != nil && *schema.Required && value == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     path,
			Message:  "required field is missing",
			Code:     "required",
			Expected: "non-null value",
			Actual:   "null",
		})
		return
	}

	// Skip validation if value is nil and not required
	if value == nil {
		return
	}

	// Validate based on type
	switch schema.Type {
	case TypeString:
		v.validateString(path, value, schema, result)
	case TypeNumber:
		v.validateNumber(path, value, schema, result)
	case TypeInteger:
		v.validateInteger(path, value, schema, result)
	case TypeBoolean:
		v.validateBoolean(path, value, schema, result)
	case TypeArray:
		v.validateArray(path, value, schema, result)
	case TypeObject:
		v.validateObject(path, value, schema, result)
	default:
		result.Warnings = append(result.Warnings, ValidationWarning{
			Path:    path,
			Message: fmt.Sprintf("unknown schema type: %s", schema.Type),
			Code:    "unknown_type",
		})
	}
}

// validateString validates a string value
func (v *Validator) validateString(path string, value interface{}, schema *OutputSchema, result *ValidationResult) {
	str, ok := value.(string)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     path,
			Message:  "expected string",
			Code:     "type_mismatch",
			Expected: "string",
			Actual:   fmt.Sprintf("%T", value),
		})
		return
	}

	// Check enum
	if len(schema.Enum) > 0 {
		found := false
		for _, e := range schema.Enum {
			if e == str {
				found = true
				break
			}
		}
		if !found {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Path:     path,
				Message:  "value not in enum",
				Code:     "enum_violation",
				Expected: fmt.Sprintf("one of %v", schema.Enum),
				Actual:   str,
			})
		}
	}

	// Check pattern
	// TODO: Add pattern validation when needed

	// Check format
	if schema.Format != "" {
		v.validateFormat(path, str, schema.Format, result)
	}
}

// validateNumber validates a number value
func (v *Validator) validateNumber(path string, value interface{}, schema *OutputSchema, result *ValidationResult) {
	var num float64
	switch n := value.(type) {
	case float64:
		num = n
	case float32:
		num = float64(n)
	case int:
		num = float64(n)
	case int64:
		num = float64(n)
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     path,
			Message:  "expected number",
			Code:     "type_mismatch",
			Expected: "number",
			Actual:   fmt.Sprintf("%T", value),
		})
		return
	}

	// Check minimum
	if schema.Minimum != nil && num < *schema.Minimum {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     path,
			Message:  "value below minimum",
			Code:     "minimum_violation",
			Expected: fmt.Sprintf(">= %v", *schema.Minimum),
			Actual:   fmt.Sprintf("%v", num),
		})
	}

	// Check maximum
	if schema.Maximum != nil && num > *schema.Maximum {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     path,
			Message:  "value above maximum",
			Code:     "maximum_violation",
			Expected: fmt.Sprintf("<= %v", *schema.Maximum),
			Actual:   fmt.Sprintf("%v", num),
		})
	}
}

// validateInteger validates an integer value
func (v *Validator) validateInteger(path string, value interface{}, schema *OutputSchema, result *ValidationResult) {
	// First check if it's a valid integer type
	var intVal int64
	switch n := value.(type) {
	case int:
		intVal = int64(n)
	case int64:
		intVal = n
	case float64:
		if n != float64(int64(n)) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Path:     path,
				Message:  "expected integer, got float",
				Code:     "type_mismatch",
				Expected: "integer",
				Actual:   fmt.Sprintf("%v", n),
			})
			return
		}
		intVal = int64(n)
	default:
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     path,
			Message:  "expected integer",
			Code:     "type_mismatch",
			Expected: "integer",
			Actual:   fmt.Sprintf("%T", value),
		})
		return
	}

	// Validate as number with integer constraint
	v.validateNumber(path, float64(intVal), schema, result)
}

// validateBoolean validates a boolean value
func (v *Validator) validateBoolean(path string, value interface{}, schema *OutputSchema, result *ValidationResult) {
	_, ok := value.(bool)
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     path,
			Message:  "expected boolean",
			Code:     "type_mismatch",
			Expected: "boolean",
			Actual:   fmt.Sprintf("%T", value),
		})
	}
}

// validateArray validates an array value
func (v *Validator) validateArray(path string, value interface{}, schema *OutputSchema, result *ValidationResult) {
	arr, ok := value.([]interface{})
	if !ok {
		// Try to convert from typed slice
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Slice {
			arr = make([]interface{}, val.Len())
			for i := 0; i < val.Len(); i++ {
				arr[i] = val.Index(i).Interface()
			}
		} else {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Path:     path,
				Message:  "expected array",
				Code:     "type_mismatch",
				Expected: "array",
				Actual:   fmt.Sprintf("%T", value),
			})
			return
		}
	}

	// Check array length constraints
	if schema.MinItems != nil && len(arr) < *schema.MinItems {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     path,
			Message:  "array too short",
			Code:     "min_items_violation",
			Expected: fmt.Sprintf(">= %d items", *schema.MinItems),
			Actual:   fmt.Sprintf("%d items", len(arr)),
		})
	}

	if schema.MaxItems != nil && len(arr) > *schema.MaxItems {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     path,
			Message:  "array too long",
			Code:     "max_items_violation",
			Expected: fmt.Sprintf("<= %d items", *schema.MaxItems),
			Actual:   fmt.Sprintf("%d items", len(arr)),
		})
	}

	// Validate items
	if schema.Items != nil {
		for i, item := range arr {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			v.validateValue(itemPath, item, schema.Items, result)
		}
	}
}

// validateObject validates an object value
func (v *Validator) validateObject(path string, value interface{}, schema *OutputSchema, result *ValidationResult) {
	obj, ok := value.(map[string]interface{})
	if !ok {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     path,
			Message:  "expected object",
			Code:     "type_mismatch",
			Expected: "object",
			Actual:   fmt.Sprintf("%T", value),
		})
		return
	}

	// Check required properties
	if schema.RequiredProperties != nil {
		for _, req := range schema.RequiredProperties {
			if _, exists := obj[req]; !exists {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Path:     fmt.Sprintf("%s.%s", path, req),
					Field:    req,
					Message:  "required property missing",
					Code:     "required_property",
					Expected: "property to exist",
					Actual:   "missing",
				})
			}
		}
	}

	// Validate properties
	if schema.Properties != nil {
		for propName, propSchema := range schema.Properties {
			propPath := fmt.Sprintf("%s.%s", path, propName)
			if propValue, exists := obj[propName]; exists {
				v.validateValue(propPath, propValue, propSchema, result)
			} else if propSchema.Required != nil && *propSchema.Required {
				// Only add error if not already in RequiredProperties
				// to avoid duplicate errors
				alreadyRequired := false
				if schema.RequiredProperties != nil {
					for _, req := range schema.RequiredProperties {
						if req == propName {
							alreadyRequired = true
							break
						}
					}
				}
				if !alreadyRequired {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Path:     propPath,
						Field:    propName,
						Message:  "required property missing",
						Code:     "required_property",
						Expected: "property to exist",
						Actual:   "missing",
					})
				}
			}
		}
	}

	// Check for additional properties
	if schema.AdditionalProperties != nil && !*schema.AdditionalProperties {
		for propName := range obj {
			if _, hasProp := schema.Properties[propName]; !hasProp {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Path:    fmt.Sprintf("%s.%s", path, propName),
					Message: "additional property not allowed",
					Code:    "additional_property",
				})
			}
		}
	}
}

// validateFormat validates string formats
func (v *Validator) validateFormat(path, value, format string, result *ValidationResult) {
	valid := true
	switch format {
	case "email":
		// Simple email check
		valid = strings.Contains(value, "@") && strings.Contains(value, ".")
	case "date":
		// Simple date format check (YYYY-MM-DD)
		valid = len(value) == 10 && value[4] == '-' && value[7] == '-'
	case "time":
		// Simple time format check (HH:MM:SS)
		valid = len(value) >= 8 && value[2] == ':' && value[5] == ':'
	case "date-time":
		// Simple datetime check
		valid = strings.Contains(value, "T") || strings.Contains(value, " ")
	case "uri":
		// Simple URI check
		valid = strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") ||
			strings.HasPrefix(value, "ftp://") || strings.Contains(value, "://")
	case "uuid":
		// Simple UUID check
		valid = len(value) == 36 && strings.Count(value, "-") == 4
	}

	if !valid {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:     path,
			Message:  fmt.Sprintf("invalid format for %s", format),
			Code:     "format_violation",
			Expected: format,
			Actual:   value,
		})
	}
}

// generateSuggestions generates fix suggestions based on validation errors
func (v *Validator) generateSuggestions(result *ValidationResult) {
	for _, err := range result.Errors {
		switch err.Code {
		case "type_mismatch":
			result.Suggestions = append(result.Suggestions, FixSuggestion{
				Path:        err.Path,
				Description: fmt.Sprintf("Convert value to %s", err.Expected),
				Fix:         fmt.Sprintf("Ensure the value at %s is of type %s", err.Path, err.Expected),
				Example:     v.getTypeExample(err.Expected),
			})

		case "enum_violation":
			result.Suggestions = append(result.Suggestions, FixSuggestion{
				Path:        err.Path,
				Description: "Use one of the allowed values",
				Fix:         fmt.Sprintf("Change the value to %s", err.Expected),
				Example:     strings.Split(err.Expected, " ")[2], // Extract first enum value
			})

		case "required", "required_property":
			result.Suggestions = append(result.Suggestions, FixSuggestion{
				Path:        err.Path,
				Description: "Add the missing required field",
				Fix:         fmt.Sprintf("Add a value for %s", err.Path),
				Example:     fmt.Sprintf(`"%s": "value"`, err.Field),
			})

		case "format_violation":
			result.Suggestions = append(result.Suggestions, FixSuggestion{
				Path:        err.Path,
				Description: fmt.Sprintf("Fix the format to match %s", err.Expected),
				Fix:         fmt.Sprintf("Ensure the value matches the %s format", err.Expected),
				Example:     v.getFormatExample(err.Expected),
			})
		}
	}
}

// getTypeExample returns an example for a type
func (v *Validator) getTypeExample(typeName string) string {
	examples := map[string]string{
		"string":  `"example string"`,
		"number":  "42.5",
		"integer": "42",
		"boolean": "true",
		"array":   `["item1", "item2"]`,
		"object":  `{"key": "value"}`,
	}
	return examples[typeName]
}

// getFormatExample returns an example for a format
func (v *Validator) getFormatExample(format string) string {
	examples := map[string]string{
		"email":     "user@example.com",
		"date":      "2024-01-15",
		"time":      "14:30:00",
		"date-time": "2024-01-15T14:30:00Z",
		"uri":       "https://example.com/path",
		"uuid":      "550e8400-e29b-41d4-a716-446655440000",
	}
	return examples[format]
}

// Validate is a convenience function that creates a validator and validates output
func Validate(ctx context.Context, output interface{}, schema *OutputSchema) (*ValidationResult, error) {
	validator := NewValidator()
	return validator.Validate(ctx, output, schema)
}
