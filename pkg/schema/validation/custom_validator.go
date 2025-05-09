package validation

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// CustomValidator is a function type for custom property validation
type CustomValidator func(value interface{}, displayPath string) []string

// CustomValidators is a registry of globally registered custom validators
var CustomValidators = sync.Map{}

// RegisterCustomValidator registers a custom validator function with a given name
func RegisterCustomValidator(name string, validator CustomValidator) {
	CustomValidators.Store(name, validator)
}

// GetCustomValidator retrieves a custom validator by name
func GetCustomValidator(name string) (CustomValidator, bool) {
	if val, ok := CustomValidators.Load(name); ok {
		if validator, ok := val.(CustomValidator); ok {
			return validator, true
		}
	}
	return nil, false
}

// Property extension for custom validators
type ExtendedProperty struct {
	CustomValidator string `json:"customValidator,omitempty"`
}

// WithCustomValidation adds support for custom validator functions
func WithCustomValidation(enable bool) func(*Validator) {
	return func(v *Validator) {
		v.enableCustomValidation = enable
	}
}

// validateWithCustomValidator runs property validation through a custom validator if specified
func (v *Validator) validateWithCustomValidator(path string, prop domain.Property, value interface{}, errors []string) []string {
	// Skip if custom validation is not enabled
	if !v.enableCustomValidation {
		return errors
	}

	// Try to extract custom validator name from the property
	// This requires unmarshaling the raw property to access extensions
	// For simplicity, we'll assume the customValidator field is added to the Property type
	var customValidatorName string

	// Get the property value and check for a JSON field named "customValidator"
	propValue := reflect.ValueOf(prop)
	if propValue.Kind() == reflect.Struct {
		for i := 0; i < propValue.NumField(); i++ {
			field := propValue.Type().Field(i)
			if field.Tag.Get("json") == "customValidator,omitempty" {
				if !propValue.Field(i).IsZero() {
					customValidatorName = propValue.Field(i).String()
				}
			}
		}
	}

	// If no custom validator is specified, return the original errors
	if customValidatorName == "" {
		return errors
	}

	// Try to get the custom validator
	validator, ok := GetCustomValidator(customValidatorName)
	if !ok {
		// Add an error indicating the custom validator wasn't found
		errors = append(errors, fmt.Sprintf("custom validator '%s' not found", customValidatorName))
		return errors
	}

	// Calculate the display path for the property
	displayPath := path
	if displayPath == "" {
		displayPath = "value"
	}

	// Run the custom validator
	customErrors := validator(value, displayPath)
	if len(customErrors) > 0 {
		errors = append(errors, customErrors...)
	}

	return errors
}

// Common standard custom validators that can be used out of the box

// ValidateNonEmpty checks if a string is not empty
func ValidateNonEmpty(value interface{}, displayPath string) []string {
	var errors []string

	if str, ok := value.(string); ok {
		if strings.TrimSpace(str) == "" {
			errors = append(errors, fmt.Sprintf("%s cannot be empty or whitespace only", displayPath))
		}
	} else {
		errors = append(errors, fmt.Sprintf("%s must be a string", displayPath))
	}

	return errors
}

// ValidateAlphanumeric checks if a string contains only alphanumeric characters
func ValidateAlphanumeric(value interface{}, displayPath string) []string {
	var errors []string

	if str, ok := value.(string); ok {
		for _, char := range str {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
				errors = append(errors, fmt.Sprintf("%s must contain only alphanumeric characters", displayPath))
				break
			}
		}
	} else {
		errors = append(errors, fmt.Sprintf("%s must be a string", displayPath))
	}

	return errors
}

// ValidateNoWhitespace checks if a string contains no whitespace
func ValidateNoWhitespace(value interface{}, displayPath string) []string {
	var errors []string

	if str, ok := value.(string); ok {
		if strings.ContainsAny(str, " \t\n\r") {
			errors = append(errors, fmt.Sprintf("%s must not contain whitespace", displayPath))
		}
	} else {
		errors = append(errors, fmt.Sprintf("%s must be a string", displayPath))
	}

	return errors
}

// ValidatePositive checks if a number is positive (greater than zero)
func ValidatePositive(value interface{}, displayPath string) []string {
	var errors []string

	switch v := value.(type) {
	case float64:
		if v <= 0 {
			errors = append(errors, fmt.Sprintf("%s must be positive", displayPath))
		}
	case int:
		if v <= 0 {
			errors = append(errors, fmt.Sprintf("%s must be positive", displayPath))
		}
	case int64:
		if v <= 0 {
			errors = append(errors, fmt.Sprintf("%s must be positive", displayPath))
		}
	default:
		errors = append(errors, fmt.Sprintf("%s must be a number", displayPath))
	}

	return errors
}

// ValidateNonNegative checks if a number is non-negative (greater than or equal to zero)
func ValidateNonNegative(value interface{}, displayPath string) []string {
	var errors []string

	switch v := value.(type) {
	case float64:
		if v < 0 {
			errors = append(errors, fmt.Sprintf("%s must be non-negative", displayPath))
		}
	case int:
		if v < 0 {
			errors = append(errors, fmt.Sprintf("%s must be non-negative", displayPath))
		}
	case int64:
		if v < 0 {
			errors = append(errors, fmt.Sprintf("%s must be non-negative", displayPath))
		}
	default:
		errors = append(errors, fmt.Sprintf("%s must be a number", displayPath))
	}

	return errors
}

// Register standard validators
func init() {
	RegisterCustomValidator("nonEmpty", ValidateNonEmpty)
	RegisterCustomValidator("alphanumeric", ValidateAlphanumeric)
	RegisterCustomValidator("noWhitespace", ValidateNoWhitespace)
	RegisterCustomValidator("positive", ValidatePositive)
	RegisterCustomValidator("nonNegative", ValidateNonNegative)
}
