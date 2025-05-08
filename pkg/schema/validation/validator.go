package validation

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// RegexCache stores compiled regular expressions to avoid recompilation
var RegexCache = sync.Map{}

// Validator implements schema validation with performance enhancements
type Validator struct {
	// errorBufferPool provides reusable string buffers for errors
	// Uses pointers to slices to avoid allocations during Put
	errorBufferPool sync.Pool

	// validationResultPool provides reusable validation results
	validationResultPool sync.Pool
	
	// enableCoercion controls whether the validator attempts to coerce values to the expected type
	enableCoercion bool
	
	// enableCustomValidation controls whether the validator supports custom validation functions
	enableCustomValidation bool
}

// NewValidator creates a new validator with performance enhancements
// This function returns a validator with improved performance through
// object pooling, regex caching, and efficient validation paths.
func NewValidator(options ...func(*Validator)) *Validator {
	v := &Validator{
		errorBufferPool: sync.Pool{
			New: func() interface{} {
				// Preallocate a slice with reasonable capacity to avoid reallocation
				// Return a pointer to avoid allocations during Put
				slice := make([]string, 0, 8)
				return &slice
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
		enableCoercion: false, // Disabled by default for backward compatibility
		enableCustomValidation: false, // Disabled by default for backward compatibility
	}
	
	// Apply options
	for _, option := range options {
		option(v)
	}
	
	return v
}

// WithCoercion enables or disables type coercion during validation
func WithCoercion(enable bool) func(*Validator) {
	return func(v *Validator) {
		v.enableCoercion = enable
	}
}

// Validate validates a JSON string against a schema
func (v *Validator) Validate(schema *domain.Schema, jsonStr string) (*domain.ValidationResult, error) {
	var data interface{}

	// Parse JSON
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Get a validation result from the pool
	result := v.validationResultPool.Get().(*domain.ValidationResult)
	result.Valid = true
	result.Errors = result.Errors[:0] // Reset the errors slice but keep capacity

	// Get an error buffer from the pool (pointer to slice)
	errorsPtr := v.errorBufferPool.Get().(*[]string)
	errors := *errorsPtr
	errors = errors[:0] // Reset slice but keep capacity

	// Validate against schema
	errors = v.validateValue("", schema, data, errors)

	if len(errors) > 0 {
		result.Valid = false
		result.Errors = append(result.Errors, errors...) // Copy errors to result
	}

	// Update the pointer's underlying slice
	*errorsPtr = errors

	// Return the error buffer to the pool
	v.errorBufferPool.Put(errorsPtr)

	return result, nil
}

// ValidateStruct validates a Go struct against a schema
func (v *Validator) ValidateStruct(schema *domain.Schema, obj interface{}) (*domain.ValidationResult, error) {
	// Convert struct to map first using JSON marshaling
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct: %w", err)
	}

	return v.Validate(schema, string(jsonData))
}

// validateValue validates a value against a schema
func (v *Validator) validateValue(path string, schema *domain.Schema, data interface{}, errors []string) []string {
	// Fast path for nil schema
	if schema == nil {
		return errors
	}
	
	// Process conditional validation first
	errors = v.validateConditional(path, schema, data, errors)

	// If schema doesn't have a type but has conditionals, we're done
	if schema.Type == "" {
		return errors
	}

	// Try to get format from properties
	var format string
	if schema.Properties != nil {
		if prop, exists := schema.Properties[""]; exists {
			format = prop.Format
		}
	}

	// Try to coerce the value to the expected type if coercion is enabled
	if v.enableCoercion {
		coercedValue, coerced := v.Coerce(schema.Type, data, format)
		if coerced {
			data = coercedValue
		}
	}

	// Validate type with fast simple type checks
	if !v.isCorrectType(schema.Type, data) {
		displayPath := path
		if displayPath == "" {
			displayPath = "value"
		}
		errors = append(errors, fmt.Sprintf("%s must be a %s", displayPath, schema.Type))
		// If type is wrong, don't proceed with further validation
		return errors
	}

	// Type-based validation with specific handling for each type
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

// validateConditional validates a value against conditional schema requirements
func (v *Validator) validateConditional(path string, schema *domain.Schema, data interface{}, errors []string) []string {
	// If-Then-Else validation
	if schema.If != nil {
		// Create a copy of errors to check if If schema validation produces errors
		ifErrors := make([]string, len(errors))
		copy(ifErrors, errors)
		
		// Validate against If schema
		ifErrors = v.validateValue(path, schema.If, data, ifErrors)
		
		// If If schema is valid (no new errors were added), apply Then schema
		if len(ifErrors) == len(errors) && schema.Then != nil {
			errors = v.validateValue(path, schema.Then, data, errors)
		} else if len(ifErrors) > len(errors) && schema.Else != nil {
			// If If schema is invalid, apply Else schema
			errors = v.validateValue(path, schema.Else, data, errors)
		}
	}

	// AllOf validation - data must be valid against all schemas
	if schema.AllOf != nil && len(schema.AllOf) > 0 {
		for _, subSchema := range schema.AllOf {
			errors = v.validateValue(path, subSchema, data, errors)
		}
	}

	// AnyOf validation - data must be valid against at least one schema
	if schema.AnyOf != nil && len(schema.AnyOf) > 0 {
		validAgainstAny := false
		
		// Try all schemas
		for _, subSchema := range schema.AnyOf {
			// Make a copy of errors for this schema
			subErrors := make([]string, len(errors))
			copy(subErrors, errors)
			
			// Validate against this schema
			subErrors = v.validateValue(path, subSchema, data, subErrors)
			
			// If no new errors were added, this schema validated
			if len(subErrors) == len(errors) {
				validAgainstAny = true
				break
			}
		}
		
		// If not valid against any schema, add a general error
		if !validAgainstAny {
			displayPath := path
			if displayPath == "" {
				displayPath = "value"
			}
			errors = append(errors, fmt.Sprintf("%s does not match any of the required schemas", displayPath))
		}
	}

	// OneOf validation - data must be valid against exactly one schema
	if schema.OneOf != nil && len(schema.OneOf) > 0 {
		validSchemaCount := 0
		
		// Try all schemas
		for _, subSchema := range schema.OneOf {
			// Make a copy of errors for this schema
			subErrors := make([]string, len(errors))
			copy(subErrors, errors)
			
			// Validate against this schema
			subErrors = v.validateValue(path, subSchema, data, subErrors)
			
			// If no new errors were added, this schema validated
			if len(subErrors) == len(errors) {
				validSchemaCount++
			}
		}
		
		// Must be valid against exactly one schema
		if validSchemaCount != 1 {
			displayPath := path
			if displayPath == "" {
				displayPath = "value"
			}
			if validSchemaCount == 0 {
				errors = append(errors, fmt.Sprintf("%s does not match any of the required schemas", displayPath))
			} else {
				errors = append(errors, fmt.Sprintf("%s matches more than one schema when it should match exactly one", displayPath))
			}
		}
	}

	// Not validation - data must NOT be valid against the schema
	if schema.Not != nil {
		// Make a copy of errors for Not schema
		notErrors := make([]string, len(errors))
		copy(notErrors, errors)
		
		// Validate against Not schema
		notErrors = v.validateValue(path, schema.Not, data, notErrors)
		
		// If no new errors were added, the Not schema validated, which is wrong
		if len(notErrors) == len(errors) {
			displayPath := path
			if displayPath == "" {
				displayPath = "value"
			}
			errors = append(errors, fmt.Sprintf("%s matches a schema that it should not match", displayPath))
		}
	}

	return errors
}

// isCorrectType checks if a value is of the expected type
func (v *Validator) isCorrectType(expectedType string, value interface{}) bool {
	// Fast path for nil values
	if value == nil {
		return false
	}

	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "integer":
		// Accept both int64 and float64 for integer type
		switch v := value.(type) {
		case int, int64:
			return true
		case float64:
			// Fast integer check
			return float64(int(v)) == v
		}
		return false
	case "number":
		switch value.(type) {
		case float64, int, int64:
			return true
		}
		return false
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	}
	return false
}

// Note: The validateType function has been replaced by isCorrectType 
// and direct type validation in validateValue

// validateObject validates an object against a schema
func (v *Validator) validateObject(path string, schema *domain.Schema, data interface{}, errors []string) []string {
	// Try to coerce to object if coercion is enabled
	if v.enableCoercion && !v.isCorrectType("object", data) {
		coercedObj, ok := CoerceToObject(data)
		if ok {
			data = coercedObj
		}
	}
	
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
					prop.Maximum != nil || prop.ExclusiveMinimum != nil || prop.ExclusiveMaximum != nil ||
					prop.MinItems != nil || prop.MaxItems != nil || prop.UniqueItems != nil ||
					prop.Items != nil {

					// Create a copy of the property in the schema's Properties map
					// to handle constraints in the validateX methods
					if subSchema.Properties == nil {
						subSchema.Properties = map[string]domain.Property{}
					}
					subSchema.Properties[""] = prop
				}

				// Validate the property value
				errors = v.validateValue(propPath, subSchema, value, errors)
				
				// If custom validation is enabled, run custom validators
				if v.enableCustomValidation && prop.CustomValidator != "" {
					errors = v.validateWithCustomValidator(propPath, prop, value, errors)
				}
			}
		}
	}

	return errors
}

// validateArray validates an array against a schema
func (v *Validator) validateArray(path string, schema *domain.Schema, data interface{}, errors []string) []string {
	// Try to coerce to array if coercion is enabled
	if v.enableCoercion && !v.isCorrectType("array", data) {
		coercedArr, ok := CoerceToArray(data)
		if ok {
			data = coercedArr
		}
	}
	
	arr, ok := data.([]interface{})
	if !ok {
		// This should never happen as we already validated the type
		return errors
	}

	// Get array-specific constraints from the special "" property
	var minItems, maxItems *int
	var uniqueItems *bool

	if schema.Properties != nil {
		if prop, exists := schema.Properties[""]; exists {
			minItems = prop.MinItems
			maxItems = prop.MaxItems
			uniqueItems = prop.UniqueItems
		}
	}

	// Validate minItems
	if minItems != nil && len(arr) < *minItems {
		displayPath := path
		if displayPath == "" {
			displayPath = "array"
		}
		errors = append(errors, fmt.Sprintf("%s must contain at least %d items", displayPath, *minItems))
	}

	// Validate maxItems
	if maxItems != nil && len(arr) > *maxItems {
		displayPath := path
		if displayPath == "" {
			displayPath = "array"
		}
		errors = append(errors, fmt.Sprintf("%s must contain no more than %d items", displayPath, *maxItems))
	}

	// Validate uniqueItems if required
	if uniqueItems != nil && *uniqueItems && len(arr) > 1 {
		// Performance optimization for small arrays: use a direct comparison
		if len(arr) <= 10 {
			for i := 0; i < len(arr)-1; i++ {
				for j := i + 1; j < len(arr); j++ {
					// Simple equality check for primitive types
					if equalValues(arr[i], arr[j]) {
						displayPath := path
						if displayPath == "" {
							displayPath = "array"
						}
						errors = append(errors, fmt.Sprintf("%s must contain unique items", displayPath))
						// Only report the error once
						i = len(arr)
						break
					}
				}
			}
		} else {
			// For larger arrays, use a map-based approach for better performance
			seen := make(map[string]bool)
			hasDuplicates := false
			
			for _, item := range arr {
				// Convert item to string for map key
				key, err := json.Marshal(item)
				if err != nil {
					continue // Skip this item if it can't be marshaled
				}
				
				keyStr := string(key)
				if seen[keyStr] {
					hasDuplicates = true
					break
				}
				seen[keyStr] = true
			}
			
			if hasDuplicates {
				displayPath := path
				if displayPath == "" {
					displayPath = "array"
				}
				errors = append(errors, fmt.Sprintf("%s must contain unique items", displayPath))
			}
		}
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
				items.Maximum != nil || items.ExclusiveMinimum != nil || items.ExclusiveMaximum != nil ||
				items.MinItems != nil || items.MaxItems != nil || items.UniqueItems != nil ||
				items.Items != nil {

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

// equalValues provides a basic equality check for common value types
func equalValues(a, b interface{}) bool {
	// Fast path for identical types
	switch a := a.(type) {
	case string:
		if b, ok := b.(string); ok {
			return a == b
		}
	case float64:
		if b, ok := b.(float64); ok {
			return a == b
		}
	case bool:
		if b, ok := b.(bool); ok {
			return a == b
		}
	case nil:
		return b == nil
	}
	
	// For complex types, use reflection or JSON marshaling
	aJson, aErr := json.Marshal(a)
	bJson, bErr := json.Marshal(b)
	
	if aErr != nil || bErr != nil {
		return false
	}
	
	return string(aJson) == string(bJson)
}

// validateStringFormat validates a string against a specific format
func (v *Validator) validateStringFormat(format string, str string, displayPath string, errors []string) []string {
	if v.enableCoercion {
		// Use coercion utilities for format validation
		switch format {
		case "email":
			if _, ok := CoerceToEmail(str); !ok {
				errors = append(errors, fmt.Sprintf("%s must be a valid email address", displayPath))
			}
		case "date", "date-time":
			if _, ok := CoerceToDate(str); !ok {
				errors = append(errors, fmt.Sprintf("%s must be a valid ISO8601 date-time", displayPath))
			}
		case "uri", "url":
			if _, ok := CoerceToURL(str); !ok {
				errors = append(errors, fmt.Sprintf("%s must be a valid URI", displayPath))
			}
		case "uuid":
			if _, ok := CoerceToUUID(str); !ok {
				errors = append(errors, fmt.Sprintf("%s must be a valid UUID", displayPath))
			}
		case "duration":
			if _, ok := CoerceToDuration(str); !ok {
				errors = append(errors, fmt.Sprintf("%s must be a valid duration", displayPath))
			}
		case "ip":
			if _, ok := CoerceToIP(str); !ok {
				errors = append(errors, fmt.Sprintf("%s must be a valid IP address", displayPath))
			}
		case "ipv4":
			if _, ok := CoerceToIPv4(str); !ok {
				errors = append(errors, fmt.Sprintf("%s must be a valid IPv4 address", displayPath))
			}
		case "ipv6":
			if _, ok := CoerceToIPv6(str); !ok {
				errors = append(errors, fmt.Sprintf("%s must be a valid IPv6 address", displayPath))
			}
		case "hostname":
			if _, ok := CoerceToHostname(str); !ok {
				errors = append(errors, fmt.Sprintf("%s must be a valid hostname", displayPath))
			}
		case "base64":
			if _, ok := CoerceToBase64(str); !ok {
				errors = append(errors, fmt.Sprintf("%s must be a valid base64 string", displayPath))
			}
		case "json":
			if _, ok := CoerceToJSON(str); !ok {
				errors = append(errors, fmt.Sprintf("%s must be a valid JSON string", displayPath))
			}
		default:
			// No error for unsupported formats when coercion is enabled
		}
	} else {
		// Use strict regex patterns for format validation
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
			// Strict ISO8601 date-time validation
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
		case "uri", "url":
			// Strict URI validation
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
		case "uuid":
			// Strict UUID validation
			uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
			var re *regexp.Regexp
			if cached, found := RegexCache.Load(uuidPattern); found {
				re = cached.(*regexp.Regexp)
			} else {
				var err error
				re, err = regexp.Compile(uuidPattern)
				if err != nil {
					errors = append(errors, fmt.Sprintf("invalid UUID pattern: %v", err))
					return errors
				}
				RegexCache.Store(uuidPattern, re)
			}
			if !re.MatchString(str) {
				errors = append(errors, fmt.Sprintf("%s must be a valid UUID", displayPath))
			}
		case "hostname":
			// Hostname validation based on RFC 1123
			hostnamePattern := `^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`
			var re *regexp.Regexp
			if cached, found := RegexCache.Load(hostnamePattern); found {
				re = cached.(*regexp.Regexp)
			} else {
				var err error
				re, err = regexp.Compile(hostnamePattern)
				if err != nil {
					errors = append(errors, fmt.Sprintf("invalid hostname pattern: %v", err))
					return errors
				}
				RegexCache.Store(hostnamePattern, re)
			}
			if !re.MatchString(str) {
				errors = append(errors, fmt.Sprintf("%s must be a valid hostname", displayPath))
			}
		case "ipv4":
			// IPv4 validation
			ipv4Pattern := `^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$`
			var re *regexp.Regexp
			if cached, found := RegexCache.Load(ipv4Pattern); found {
				re = cached.(*regexp.Regexp)
			} else {
				var err error
				re, err = regexp.Compile(ipv4Pattern)
				if err != nil {
					errors = append(errors, fmt.Sprintf("invalid IPv4 pattern: %v", err))
					return errors
				}
				RegexCache.Store(ipv4Pattern, re)
			}
			if !re.MatchString(str) {
				errors = append(errors, fmt.Sprintf("%s must be a valid IPv4 address", displayPath))
				return errors
			}
			
			// Validate each octet
			parts := strings.Split(str, ".")
			for _, part := range parts {
				if num, err := strconv.Atoi(part); err != nil || num < 0 || num > 255 {
					errors = append(errors, fmt.Sprintf("%s must be a valid IPv4 address", displayPath))
					break
				}
			}
		case "ipv6":
			// Just check if net.ParseIP parses it as a valid IPv6 address
			ip := net.ParseIP(str)
			if ip == nil || ip.To4() != nil {
				errors = append(errors, fmt.Sprintf("%s must be a valid IPv6 address", displayPath))
			}
		case "base64":
			// Validate base64 encoding
			_, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				// Try URL-safe base64
				_, err = base64.URLEncoding.DecodeString(str)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s must be a valid base64 string", displayPath))
				}
			}
		case "json":
			// Validate JSON
			var j interface{}
			if err := json.Unmarshal([]byte(str), &j); err != nil {
				errors = append(errors, fmt.Sprintf("%s must be a valid JSON string", displayPath))
			}
		default:
			errors = append(errors, fmt.Sprintf("unsupported format: %s", format))
		}
	}
	
	return errors
}

// validateString validates a string against constraints
func (v *Validator) validateString(path string, schema *domain.Schema, data interface{}, errors []string) []string {
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

	// Validate enum with efficient loop
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

	// Validate format with improved format support
	if format != "" {
		// Check for multiple formats (separated by comma or pipe)
		if strings.Contains(format, ",") || strings.Contains(format, "|") {
			var separator string
			if strings.Contains(format, ",") {
				separator = ","
			} else {
				separator = "|"
			}
			
			formats := strings.Split(format, separator)
			validAgainstAny := false
			
			// Try validating against each format until one succeeds
			for _, fmt := range formats {
				fmt = strings.TrimSpace(fmt)
				// Make a copy of errors for this format
				tmpErrors := make([]string, len(errors))
				copy(tmpErrors, errors)
				
				// Validate against this format
				tmpErrors = v.validateStringFormat(fmt, str, displayPath, tmpErrors)
				
				// If no new errors were added, this format is valid
				if len(tmpErrors) == len(errors) {
					validAgainstAny = true
					break
				}
			}
			
			// If not valid against any format, add a general error
			if !validAgainstAny {
				errors = append(errors, fmt.Sprintf("%s must match one of these formats: %s", displayPath, format))
			}
		} else {
			// Single format validation
			errors = v.validateStringFormat(format, str, displayPath, errors)
		}
	}

	return errors
}

// validateNumber validates a number against constraints
func (v *Validator) validateNumber(path string, schema *domain.Schema, data interface{}, errors []string) []string {
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
	var minimum, maximum, exclusiveMinimum, exclusiveMaximum *float64

	if schema.Properties != nil {
		if prop, exists := schema.Properties[""]; exists {
			minimum = prop.Minimum
			maximum = prop.Maximum
			exclusiveMinimum = prop.ExclusiveMinimum
			exclusiveMaximum = prop.ExclusiveMaximum
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

	// Validate exclusive minimum
	if exclusiveMinimum != nil && num <= *exclusiveMinimum {
		errors = append(errors, fmt.Sprintf("%s must be greater than %g", displayPath, *exclusiveMinimum))
	}

	// Validate exclusive maximum
	if exclusiveMaximum != nil && num >= *exclusiveMaximum {
		errors = append(errors, fmt.Sprintf("%s must be less than %g", displayPath, *exclusiveMaximum))
	}

	return errors
}