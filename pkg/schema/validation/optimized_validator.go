package validation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// RegexCache stores compiled regular expressions to avoid recompilation
var RegexCache = sync.Map{}

// OptimizedValidator implements schema validation with optimizations
type OptimizedValidator struct {
	// errorBufferPool provides reusable string buffers for errors
	errorBufferPool sync.Pool

	// validationResultPool provides reusable validation results
	validationResultPool sync.Pool
}

// NewOptimizedValidator creates a new optimized validator
func NewOptimizedValidator() *OptimizedValidator {
	v := &OptimizedValidator{
		errorBufferPool: sync.Pool{
			New: func() interface{} {
				// Preallocate a slice with reasonable capacity to avoid reallocation
				return make([]string, 0, 8)
			},
		},
		validationResultPool: sync.Pool{
			New: func() interface{} {
				return &domain.ValidationResult{
					Valid:  true,
					Errors: make([]string, 0, 8),
				}
			},
		},
	}
	return v
}

// Validate validates a JSON string against a schema
func (v *OptimizedValidator) Validate(schema *domain.Schema, jsonStr string) (*domain.ValidationResult, error) {
	var data interface{}

	// Parse JSON
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Get a validation result from the pool
	result := v.validationResultPool.Get().(*domain.ValidationResult)
	result.Valid = true
	result.Errors = result.Errors[:0] // Reset the errors slice but keep capacity

	// Get an error buffer from the pool
	errors := v.errorBufferPool.Get().([]string)
	errors = errors[:0] // Reset slice but keep capacity

	// Validate against schema
	errors = v.validateValue("", schema, data, errors)

	if len(errors) > 0 {
		result.Valid = false
		result.Errors = append(result.Errors, errors...) // Copy errors to result
	}

	// Return the error buffer to the pool
	v.errorBufferPool.Put(errors)

	return result, nil
}

// ValidateStruct validates a Go struct against a schema
func (v *OptimizedValidator) ValidateStruct(schema *domain.Schema, obj interface{}) (*domain.ValidationResult, error) {
	// Convert struct to map first using JSON marshaling
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct: %w", err)
	}

	return v.Validate(schema, string(jsonData))
}

// validateValue validates a value against a schema
func (v *OptimizedValidator) validateValue(path string, schema *domain.Schema, data interface{}, errors []string) []string {
	// Fast path for nil schema or empty type
	if schema == nil || schema.Type == "" {
		return errors
	}

	// Validate type first with fast simple type checks
	typeError := v.validateType(path, schema.Type, data)
	if typeError != "" {
		errors = append(errors, typeError)
		// If type is wrong, don't proceed with further validation
		return errors
	}

	// Optimized validation based on type
	switch schema.Type {
	case "object":
		errors = v.validateObject(path, schema, data, errors)
	case "array":
		errors = v.validateArray(path, schema, data, errors)
	case "string":
		errors = v.validateString(path, schema, data, errors)
	case "integer", "number":
		errors = v.validateNumber(path, schema, data, errors)
	}

	return errors
}

// validateType validates the type of a value
func (v *OptimizedValidator) validateType(path string, expectedType string, value interface{}) string {
	displayPath := path
	if displayPath == "" {
		displayPath = "value"
	}

	// Fast path for nil values
	if value == nil {
		return fmt.Sprintf("%s is null but expected %s", displayPath, expectedType)
	}

	// Use type switches for more efficient type checking
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Sprintf("%s must be a string", displayPath)
		}
	case "integer":
		num, ok := value.(float64)
		if !ok {
			return fmt.Sprintf("%s must be an integer", displayPath)
		}
		// Fast integer check
		if float64(int(num)) != num {
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
func (v *OptimizedValidator) validateObject(path string, schema *domain.Schema, data interface{}, errors []string) []string {
	obj, ok := data.(map[string]interface{})
	if !ok {
		// This should never happen as we already validated the type
		return errors
	}

	// Check required properties - fast path with direct iteration
	if len(schema.Required) > 0 {
		for _, req := range schema.Required {
			if _, exists := obj[req]; !exists {
				propPath := req
				if path != "" {
					propPath = path + "." + req
				}
				errors = append(errors, fmt.Sprintf("property %s is required", propPath))
			}
		}
	}

	// Only validate defined properties
	if schema.Properties != nil {
		for name, prop := range schema.Properties {
			if value, exists := obj[name]; exists {
				propPath := name
				if path != "" {
					propPath = path + "." + name
				}

				// Create a schema for this property
				subSchema := &domain.Schema{
					Type:        prop.Type,
					Properties:  prop.Properties,
					Required:    prop.Required,
					Description: prop.Description,
				}

				// For simple constraint validations in string, number, etc.
				// we need to handle the constraints in a special way
				if prop.MinLength != nil || prop.MaxLength != nil || prop.Pattern != "" || 
				   len(prop.Enum) > 0 || prop.Format != "" || prop.Minimum != nil ||
				   prop.Maximum != nil || prop.Items != nil {
					
					// Create a copy of the property in the schema's Properties map
					// to handle constraints in the validateX methods
					if subSchema.Properties == nil {
						subSchema.Properties = map[string]domain.Property{}
					}
					subSchema.Properties[""] = prop
				}

				// Validate the property value
				errors = v.validateValue(propPath, subSchema, value, errors)
			}
		}
	}

	return errors
}

// validateArray validates an array against a schema
func (v *OptimizedValidator) validateArray(path string, schema *domain.Schema, data interface{}, errors []string) []string {
	arr, ok := data.([]interface{})
	if !ok {
		// This should never happen as we already validated the type
		return errors
	}

	// Get the items schema if available
	var itemsSchema *domain.Schema
	if schema.Properties != nil {
		if prop, exists := schema.Properties[""]; exists && prop.Items != nil {
			items := prop.Items
			itemsSchema = &domain.Schema{
				Type:        items.Type,
				Properties:  items.Properties,
				Required:    items.Required,
				Description: items.Description,
			}

			// Inherit constraints from the item schema if needed
			if items.MinLength != nil || items.MaxLength != nil || items.Pattern != "" || 
			   len(items.Enum) > 0 || items.Format != "" || items.Minimum != nil ||
			   items.Maximum != nil || items.Items != nil {
				
				if itemsSchema.Properties == nil {
					itemsSchema.Properties = map[string]domain.Property{}
				}
				itemsSchema.Properties[""] = *items
			}
		}
	}

	// If no items schema, nothing to validate
	if itemsSchema == nil {
		return errors
	}

	// Cache the path prefix for better performance
	pathPrefix := path
	if pathPrefix != "" {
		pathPrefix = pathPrefix + "["
	} else {
		pathPrefix = "["
	}

	// Validate each item
	for i, item := range arr {
		itemPath := fmt.Sprintf("%s%d]", pathPrefix, i)
		errors = v.validateValue(itemPath, itemsSchema, item, errors)
	}

	return errors
}

// validateString validates a string against constraints
func (v *OptimizedValidator) validateString(path string, schema *domain.Schema, data interface{}, errors []string) []string {
	str, ok := data.(string)
	if !ok {
		// This should never happen as we already validated the type
		return errors
	}

	displayPath := path
	if displayPath == "" {
		displayPath = "value"
	}

	// Get string constraints from the special "" property
	var minLength, maxLength *int
	var pattern, format string
	var enum []string

	if schema.Properties != nil {
		if prop, exists := schema.Properties[""]; exists {
			minLength = prop.MinLength
			maxLength = prop.MaxLength
			pattern = prop.Pattern
			format = prop.Format
			enum = prop.Enum
		}
	}

	// Validate min length - fast path
	if minLength != nil && len(str) < *minLength {
		errors = append(errors, fmt.Sprintf("%s must be at least %d characters long", displayPath, *minLength))
	}

	// Validate max length - fast path
	if maxLength != nil && len(str) > *maxLength {
		errors = append(errors, fmt.Sprintf("%s must be no more than %d characters long", displayPath, *maxLength))
	}

	// Validate pattern using regex cache
	if pattern != "" {
		// Get or compile the regex
		var re *regexp.Regexp
		if cached, found := RegexCache.Load(pattern); found {
			re = cached.(*regexp.Regexp)
		} else {
			var err error
			re, err = regexp.Compile(pattern)
			if err != nil {
				errors = append(errors, fmt.Sprintf("invalid pattern: %s", pattern))
				return errors
			}
			RegexCache.Store(pattern, re)
		}

		if !re.MatchString(str) {
			errors = append(errors, fmt.Sprintf("%s must match pattern: %s", displayPath, pattern))
		}
	}

	// Validate enum with optimized loop
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

	// Validate format with cached regex patterns
	if format != "" {
		switch format {
		case "email":
			emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
			var re *regexp.Regexp
			if cached, found := RegexCache.Load(emailPattern); found {
				re = cached.(*regexp.Regexp)
			} else {
				var err error
				re, err = regexp.Compile(emailPattern)
				if err != nil {
					errors = append(errors, fmt.Sprintf("invalid email pattern: %v", err))
					return errors
				}
				RegexCache.Store(emailPattern, re)
			}
			if !re.MatchString(str) {
				errors = append(errors, fmt.Sprintf("%s must be a valid email address", displayPath))
			}
		case "date-time":
			// Simplified ISO8601 date-time validation
			dateTimePattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`
			var re *regexp.Regexp
			if cached, found := RegexCache.Load(dateTimePattern); found {
				re = cached.(*regexp.Regexp)
			} else {
				var err error
				re, err = regexp.Compile(dateTimePattern)
				if err != nil {
					errors = append(errors, fmt.Sprintf("invalid date-time pattern: %v", err))
					return errors
				}
				RegexCache.Store(dateTimePattern, re)
			}
			if !re.MatchString(str) {
				errors = append(errors, fmt.Sprintf("%s must be a valid ISO8601 date-time", displayPath))
			}
		case "uri":
			// Simplified URI validation
			uriPattern := `^(https?|ftp)://[^\s/$.?#].[^\s]*$`
			var re *regexp.Regexp
			if cached, found := RegexCache.Load(uriPattern); found {
				re = cached.(*regexp.Regexp)
			} else {
				var err error
				re, err = regexp.Compile(uriPattern)
				if err != nil {
					errors = append(errors, fmt.Sprintf("invalid URI pattern: %v", err))
					return errors
				}
				RegexCache.Store(uriPattern, re)
			}
			if !re.MatchString(str) {
				errors = append(errors, fmt.Sprintf("%s must be a valid URI", displayPath))
			}
		default:
			errors = append(errors, fmt.Sprintf("unsupported format: %s", format))
		}
	}

	return errors
}

// validateNumber validates a number against constraints
func (v *OptimizedValidator) validateNumber(path string, schema *domain.Schema, data interface{}, errors []string) []string {
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
		if prop, exists := schema.Properties[""]; exists {
			minimum = prop.Minimum
			maximum = prop.Maximum
		}
	}

	// Validate minimum - fast path
	if minimum != nil && num < *minimum {
		errors = append(errors, fmt.Sprintf("%s must be at least %g", displayPath, *minimum))
	}

	// Validate maximum - fast path
	if maximum != nil && num > *maximum {
		errors = append(errors, fmt.Sprintf("%s must be at most %g", displayPath, *maximum))
	}

	return errors
}