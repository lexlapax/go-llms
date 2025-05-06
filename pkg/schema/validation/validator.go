package validation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// DefaultValidator implements schema validation
type DefaultValidator struct{}

// NewValidator creates a new validator
func NewValidator() *DefaultValidator {
	return &DefaultValidator{}
}

// Validate validates a JSON string against a schema
func (v *DefaultValidator) Validate(schema *domain.Schema, jsonStr string) (*domain.ValidationResult, error) {
	var data interface{}

	// Parse JSON
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Validate against schema
	result := &domain.ValidationResult{Valid: true}
	errors := v.validateValue("", schema, data)

	if len(errors) > 0 {
		result.Valid = false
		result.Errors = errors
	}

	return result, nil
}

// ValidateStruct validates a Go struct against a schema
func (v *DefaultValidator) ValidateStruct(schema *domain.Schema, obj interface{}) (*domain.ValidationResult, error) {
	// Convert struct to map first using reflection
	// For now, we'll use JSON marshaling as a simple implementation
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct: %w", err)
	}

	return v.Validate(schema, string(jsonData))
}

// validateValue validates a value against a schema
func (v *DefaultValidator) validateValue(path string, schema *domain.Schema, data interface{}) []string {
	var errors []string

	// Validate type first
	if schema.Type != "" {
		if typeErr := v.validateType(path, schema.Type, data); typeErr != "" {
			errors = append(errors, typeErr)
			// If type is wrong, don't proceed with further validation
			return errors
		}
	}

	// Based on the schema type, validate the appropriate constraints
	switch schema.Type {
	case "object":
		errors = append(errors, v.validateObject(path, schema, data)...)
	case "array":
		errors = append(errors, v.validateArray(path, schema, data)...)
	case "string":
		errors = append(errors, v.validateString(path, schema, data)...)
	case "integer", "number":
		errors = append(errors, v.validateNumber(path, schema, data)...)
	case "boolean":
		// No additional validation needed for booleans
	}

	return errors
}

// validateType validates the type of a value
func (v *DefaultValidator) validateType(path string, expectedType string, value interface{}) string {
	displayPath := path
	if displayPath == "" {
		displayPath = "value"
	}

	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Sprintf("%s must be a string", displayPath)
		}
	case "integer":
		if _, ok := value.(float64); !ok {
			return fmt.Sprintf("%s must be an integer", displayPath)
		}
		// In JSON, all numbers are float64, so check if it's a whole number
		if float64(int(value.(float64))) != value.(float64) {
			return fmt.Sprintf("%s must be an integer (no decimal part)", displayPath)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Sprintf("%s must be a number", displayPath)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Sprintf("%s must be a boolean", displayPath)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Sprintf("%s must be an object", displayPath)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Sprintf("%s must be an array", displayPath)
		}
	}

	return ""
}

// validateObject validates an object against a schema
func (v *DefaultValidator) validateObject(path string, schema *domain.Schema, data interface{}) []string {
	var errors []string
	obj, ok := data.(map[string]interface{})
	if !ok {
		// This should never happen as we already validated the type
		return errors
	}

	// Check required properties
	for _, req := range schema.Required {
		if _, exists := obj[req]; !exists {
			propPath := req
			if path != "" {
				propPath = path + "." + req
			}
			errors = append(errors, fmt.Sprintf("property %s is required", propPath))
		}
	}

	// Validate each property against its schema
	for name, propSchema := range schema.Properties {
		propPath := name
		if path != "" {
			propPath = path + "." + name
		}

		if value, exists := obj[name]; exists {
			// Create a sub-schema for the property
			subSchema := &domain.Schema{
				Type:        propSchema.Type,
				Properties:  propSchema.Properties,
				Required:    propSchema.Required,
				Description: propSchema.Description,
			}

			// Copy constraints to sub-schema
			if propSchema.Minimum != nil {
				if subSchema.Properties == nil {
					subSchema.Properties = make(map[string]domain.Property)
				}
				prop := domain.Property{Minimum: propSchema.Minimum}
				subSchema.Properties[""] = prop
			}
			if propSchema.Maximum != nil {
				if subSchema.Properties == nil {
					subSchema.Properties = make(map[string]domain.Property)
				}
				prop := subSchema.Properties[""]
				prop.Maximum = propSchema.Maximum
				subSchema.Properties[""] = prop
			}
			if propSchema.MinLength != nil {
				if subSchema.Properties == nil {
					subSchema.Properties = make(map[string]domain.Property)
				}
				prop := subSchema.Properties[""]
				prop.MinLength = propSchema.MinLength
				subSchema.Properties[""] = prop
			}
			if propSchema.MaxLength != nil {
				if subSchema.Properties == nil {
					subSchema.Properties = make(map[string]domain.Property)
				}
				prop := subSchema.Properties[""]
				prop.MaxLength = propSchema.MaxLength
				subSchema.Properties[""] = prop
			}
			if propSchema.Pattern != "" {
				if subSchema.Properties == nil {
					subSchema.Properties = make(map[string]domain.Property)
				}
				prop := subSchema.Properties[""]
				prop.Pattern = propSchema.Pattern
				subSchema.Properties[""] = prop
			}
			if len(propSchema.Enum) > 0 {
				if subSchema.Properties == nil {
					subSchema.Properties = make(map[string]domain.Property)
				}
				prop := subSchema.Properties[""]
				prop.Enum = propSchema.Enum
				subSchema.Properties[""] = prop
			}
			if propSchema.Format != "" {
				if subSchema.Properties == nil {
					subSchema.Properties = make(map[string]domain.Property)
				}
				prop := subSchema.Properties[""]
				prop.Format = propSchema.Format
				subSchema.Properties[""] = prop
			}
			if propSchema.Items != nil {
				if subSchema.Properties == nil {
					subSchema.Properties = make(map[string]domain.Property)
				}
				prop := subSchema.Properties[""]
				prop.Items = propSchema.Items
				subSchema.Properties[""] = prop
			}

			// Validate the property value
			propErrors := v.validateValue(propPath, subSchema, value)
			errors = append(errors, propErrors...)
		}
	}

	return errors
}

// validateArray validates an array against a schema
func (v *DefaultValidator) validateArray(path string, schema *domain.Schema, data interface{}) []string {
	var errors []string
	arr, ok := data.([]interface{})
	if !ok {
		// This should never happen as we already validated the type
		return errors
	}

	// If schema.Properties is empty or schema.Properties[""] is not set
	if schema.Properties == nil {
		return errors
	}

	itemsProp, exists := schema.Properties[""]
	if !exists || itemsProp.Items == nil {
		return errors
	}

	// Get the schema for array items
	itemSchema := itemsProp.Items
	subSchema := &domain.Schema{
		Type:        itemSchema.Type,
		Properties:  itemSchema.Properties,
		Required:    itemSchema.Required,
		Description: itemSchema.Description,
	}

	// Apply additional constraints from the item schema
	if itemSchema.Minimum != nil || itemSchema.Maximum != nil ||
		itemSchema.MinLength != nil || itemSchema.MaxLength != nil ||
		itemSchema.Pattern != "" || itemSchema.Format != "" ||
		len(itemSchema.Enum) > 0 || itemSchema.Items != nil {

		if subSchema.Properties == nil {
			subSchema.Properties = make(map[string]domain.Property)
		}

		// Create a property with constraints
		prop := domain.Property{}

		// Copy numeric constraints
		if itemSchema.Minimum != nil {
			prop.Minimum = itemSchema.Minimum
		}
		if itemSchema.Maximum != nil {
			prop.Maximum = itemSchema.Maximum
		}

		// Copy string constraints
		if itemSchema.MinLength != nil {
			prop.MinLength = itemSchema.MinLength
		}
		if itemSchema.MaxLength != nil {
			prop.MaxLength = itemSchema.MaxLength
		}
		if itemSchema.Pattern != "" {
			prop.Pattern = itemSchema.Pattern
		}
		if len(itemSchema.Enum) > 0 {
			prop.Enum = itemSchema.Enum
		}
		if itemSchema.Format != "" {
			prop.Format = itemSchema.Format
		}

		// Copy nested array constraints
		if itemSchema.Items != nil {
			prop.Items = itemSchema.Items
		}

		// Add the property
		subSchema.Properties[""] = prop
	}

	// Validate each item in the array
	for i, item := range arr {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		itemErrors := v.validateValue(itemPath, subSchema, item)
		errors = append(errors, itemErrors...)
	}

	return errors
}

// validateString validates a string against constraints
func (v *DefaultValidator) validateString(path string, schema *domain.Schema, data interface{}) []string {
	var errors []string
	str, ok := data.(string)
	if !ok {
		// This should never happen as we already validated the type
		return errors
	}

	displayPath := path
	if displayPath == "" {
		displayPath = "value"
	}

	// Get string constraints
	var minLength, maxLength *int
	var pattern, format string
	var enum []string

	if schema.Properties != nil {
		prop, exists := schema.Properties[""]
		if exists {
			minLength = prop.MinLength
			maxLength = prop.MaxLength
			pattern = prop.Pattern
			format = prop.Format
			enum = prop.Enum
		}
	}

	// Validate min length
	if minLength != nil && len(str) < *minLength {
		errors = append(errors, fmt.Sprintf("%s must be at least %d characters long", displayPath, *minLength))
	}

	// Validate max length
	if maxLength != nil && len(str) > *maxLength {
		errors = append(errors, fmt.Sprintf("%s must be no more than %d characters long", displayPath, *maxLength))
	}

	// Validate pattern
	if pattern != "" {
		re, err := regexp.Compile(pattern)
		if err != nil {
			errors = append(errors, fmt.Sprintf("invalid pattern: %s", pattern))
		} else if !re.MatchString(str) {
			errors = append(errors, fmt.Sprintf("%s must match pattern: %s", displayPath, pattern))
		}
	}

	// Validate enum
	if len(enum) > 0 {
		valid := false
		for _, enumValue := range enum {
			if str == enumValue {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, fmt.Sprintf("%s must be one of: %s", displayPath, strings.Join(enum, ", ")))
		}
	}

	// Validate format
	if format != "" {
		switch format {
		case "email":
			emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			re, err := regexp.Compile(emailPattern)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid email pattern: %v", err))
			} else if !re.MatchString(str) {
				errors = append(errors, fmt.Sprintf("%s must be a valid email address", displayPath))
			}
		case "date-time":
			// Simplified ISO8601 date-time validation
			dateTimePattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`
			re, err := regexp.Compile(dateTimePattern)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid date-time pattern: %v", err))
			} else if !re.MatchString(str) {
				errors = append(errors, fmt.Sprintf("%s must be a valid ISO8601 date-time", displayPath))
			}
		case "uri":
			// Simplified URI validation
			uriPattern := `^(https?|ftp)://[^\s/$.?#].[^\s]*$`
			re, err := regexp.Compile(uriPattern)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid URI pattern: %v", err))
			} else if !re.MatchString(str) {
				errors = append(errors, fmt.Sprintf("%s must be a valid URI", displayPath))
			}
		default:
			errors = append(errors, fmt.Sprintf("unsupported format: %s", format))
		}
	}

	return errors
}

// validateNumber validates a number against constraints
func (v *DefaultValidator) validateNumber(path string, schema *domain.Schema, data interface{}) []string {
	var errors []string
	num, ok := data.(float64)
	if !ok {
		// This should never happen as we already validated the type
		return errors
	}

	displayPath := path
	if displayPath == "" {
		displayPath = "value"
	}

	// Get number constraints
	var minimum, maximum *float64

	if schema.Properties != nil {
		prop, exists := schema.Properties[""]
		if exists {
			minimum = prop.Minimum
			maximum = prop.Maximum
		}
	}

	// Validate minimum
	if minimum != nil && num < *minimum {
		errors = append(errors, fmt.Sprintf("%s must be at least %g", displayPath, *minimum))
	}

	// Validate maximum
	if maximum != nil && num > *maximum {
		errors = append(errors, fmt.Sprintf("%s must be at most %g", displayPath, *maximum))
	}

	return errors
}
