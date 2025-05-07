package validation

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestValidatorEquivalence tests that validator instances produce consistent results
// This test was originally used to compare optimized and unoptimized implementations,
// but now tests that multiple instances of the same (optimized) validator behave identically
func TestValidatorEquivalence(t *testing.T) {
	// Create two instances of the validator to ensure they behave consistently
	validator1 := NewValidator()
	validator2 := NewValidator()

	// Define a set of test cases covering various schema types and validations
	testCases := []struct {
		name   string
		schema *domain.Schema
		json   string
		valid  bool // expected validation result
	}{
		{
			name: "Simple Object Validation (Valid)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"name": {Type: "string"},
					"age":  {Type: "integer"},
				},
				Required: []string{"name"},
			},
			json:  `{"name": "John", "age": 30}`,
			valid: true,
		},
		{
			name: "Simple Object Validation (Invalid - Missing Required)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"name": {Type: "string"},
					"age":  {Type: "integer"},
				},
				Required: []string{"name", "age"},
			},
			json:  `{"name": "John"}`,
			valid: false,
		},
		{
			name: "String Constraints (Valid)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"username": {
						Type:      "string",
						MinLength: intPtr(3),
						MaxLength: intPtr(20),
						Pattern:   "^[a-zA-Z0-9_]+$",
					},
				},
				Required: []string{"username"},
			},
			json:  `{"username": "john_doe123"}`,
			valid: true,
		},
		{
			name: "String Constraints (Invalid - Too Short)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"username": {
						Type:      "string",
						MinLength: intPtr(3),
						MaxLength: intPtr(20),
						Pattern:   "^[a-zA-Z0-9_]+$",
					},
				},
				Required: []string{"username"},
			},
			json:  `{"username": "jo"}`,
			valid: false,
		},
		{
			name: "String Constraints (Invalid - Too Long)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"username": {
						Type:      "string",
						MinLength: intPtr(3),
						MaxLength: intPtr(10),
						Pattern:   "^[a-zA-Z0-9_]+$",
					},
				},
				Required: []string{"username"},
			},
			json:  `{"username": "johndoe1234567890"}`,
			valid: false,
		},
		{
			name: "String Constraints (Invalid - Pattern)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"username": {
						Type:      "string",
						MinLength: intPtr(3),
						MaxLength: intPtr(20),
						Pattern:   "^[a-zA-Z0-9_]+$",
					},
				},
				Required: []string{"username"},
			},
			json:  `{"username": "john@doe"}`,
			valid: false,
		},
		{
			name: "Numeric Constraints (Valid)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"age": {
						Type:    "integer",
						Minimum: float64Ptr(18),
						Maximum: float64Ptr(120),
					},
					"score": {
						Type:    "number",
						Minimum: float64Ptr(0),
						Maximum: float64Ptr(100),
					},
				},
				Required: []string{"age", "score"},
			},
			json:  `{"age": 35, "score": 92.5}`,
			valid: true,
		},
		{
			name: "Numeric Constraints (Invalid - Below Minimum)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"age": {
						Type:    "integer",
						Minimum: float64Ptr(18),
						Maximum: float64Ptr(120),
					},
				},
				Required: []string{"age"},
			},
			json:  `{"age": 16}`,
			valid: false,
		},
		{
			name: "Numeric Constraints (Invalid - Above Maximum)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"age": {
						Type:    "integer",
						Minimum: float64Ptr(18),
						Maximum: float64Ptr(120),
					},
				},
				Required: []string{"age"},
			},
			json:  `{"age": 150}`,
			valid: false,
		},
		{
			name: "Numeric Constraints (Invalid - Integer Type)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"age": {
						Type: "integer",
					},
				},
				Required: []string{"age"},
			},
			json:  `{"age": 35.5}`,
			valid: false,
		},
		{
			name: "Enum Validation (Valid)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"color": {
						Type: "string",
						Enum: []string{"red", "green", "blue"},
					},
				},
				Required: []string{"color"},
			},
			json:  `{"color": "green"}`,
			valid: true,
		},
		{
			name: "Enum Validation (Invalid)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"color": {
						Type: "string",
						Enum: []string{"red", "green", "blue"},
					},
				},
				Required: []string{"color"},
			},
			json:  `{"color": "yellow"}`,
			valid: false,
		},
		{
			name: "Format Validation (Valid)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"email": {
						Type:   "string",
						Format: "email",
					},
					"website": {
						Type:   "string",
						Format: "uri",
					},
				},
				Required: []string{"email", "website"},
			},
			json:  `{"email": "user@example.com", "website": "https://example.com"}`,
			valid: true,
		},
		{
			name: "Format Validation (Invalid - Email)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"email": {
						Type:   "string",
						Format: "email",
					},
				},
				Required: []string{"email"},
			},
			json:  `{"email": "invalid-email"}`,
			valid: false,
		},
		{
			name: "Format Validation (Invalid - URI)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"website": {
						Type:   "string",
						Format: "uri",
					},
				},
				Required: []string{"website"},
			},
			json:  `{"website": "example.com"}`,
			valid: false,
		},
		{
			name: "Array Validation (Valid)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"tags": {
						Type: "array",
						Items: &domain.Property{
							Type: "string",
						},
					},
				},
				Required: []string{"tags"},
			},
			json:  `{"tags": ["one", "two", "three"]}`,
			valid: true,
		},
		{
			name: "Array Validation (Invalid - Item Type)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"tags": {
						Type: "array",
						Items: &domain.Property{
							Type: "string",
						},
					},
				},
				Required: []string{"tags"},
			},
			json:  `{"tags": ["one", 2, "three"]}`,
			valid: false,
		},
		{
			name: "Complex Nested Object (Valid)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"user": {
						Type: "object",
						Properties: map[string]domain.Property{
							"name": {Type: "string"},
							"address": {
								Type: "object",
								Properties: map[string]domain.Property{
									"street": {Type: "string"},
									"city":   {Type: "string"},
									"zip":    {Type: "string"},
								},
								Required: []string{"street", "city"},
							},
						},
						Required: []string{"name", "address"},
					},
					"active": {Type: "boolean"},
				},
				Required: []string{"user", "active"},
			},
			json: `{
				"user": {
					"name": "John Doe",
					"address": {
						"street": "123 Main St",
						"city": "Anytown",
						"zip": "12345"
					}
				},
				"active": true
			}`,
			valid: true,
		},
		{
			name: "Complex Nested Object (Invalid - Missing Nested Required)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"user": {
						Type: "object",
						Properties: map[string]domain.Property{
							"name": {Type: "string"},
							"address": {
								Type: "object",
								Properties: map[string]domain.Property{
									"street": {Type: "string"},
									"city":   {Type: "string"},
									"zip":    {Type: "string"},
								},
								Required: []string{"street", "city"},
							},
						},
						Required: []string{"name", "address"},
					},
				},
				Required: []string{"user"},
			},
			json: `{
				"user": {
					"name": "John Doe",
					"address": {
						"street": "123 Main St"
					}
				}
			}`,
			valid: false,
		},
		{
			name: "Complex Array of Objects (Valid)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"items": {
						Type: "array",
						Items: &domain.Property{
							Type: "object",
							Properties: map[string]domain.Property{
								"id":    {Type: "integer"},
								"name":  {Type: "string"},
								"price": {Type: "number"},
							},
							Required: []string{"id", "name"},
						},
					},
				},
				Required: []string{"items"},
			},
			json: `{
				"items": [
					{"id": 1, "name": "Item 1", "price": 10.99},
					{"id": 2, "name": "Item 2", "price": 20.50},
					{"id": 3, "name": "Item 3", "price": 5.25}
				]
			}`,
			valid: true,
		},
		{
			name: "Complex Array of Objects (Invalid - Missing Required in Array Item)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"items": {
						Type: "array",
						Items: &domain.Property{
							Type: "object",
							Properties: map[string]domain.Property{
								"id":    {Type: "integer"},
								"name":  {Type: "string"},
								"price": {Type: "number"},
							},
							Required: []string{"id", "name", "price"},
						},
					},
				},
				Required: []string{"items"},
			},
			json: `{
				"items": [
					{"id": 1, "name": "Item 1", "price": 10.99},
					{"id": 2, "name": "Item 2"},
					{"id": 3, "name": "Item 3", "price": 5.25}
				]
			}`,
			valid: false,
		},
		{
			name: "Various Types Validation (Valid)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"string":  {Type: "string"},
					"integer": {Type: "integer"},
					"number":  {Type: "number"},
					"boolean": {Type: "boolean"},
					"object":  {Type: "object"},
					"array":   {Type: "array"},
				},
				Required: []string{"string", "integer", "number", "boolean", "object", "array"},
			},
			json: `{
				"string": "test",
				"integer": 42,
				"number": 3.14,
				"boolean": true,
				"object": {"key": "value"},
				"array": [1, 2, 3]
			}`,
			valid: true,
		},
		{
			name: "Various Types Validation (Invalid - Wrong Types)",
			schema: &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"string":  {Type: "string"},
					"integer": {Type: "integer"},
					"number":  {Type: "number"},
					"boolean": {Type: "boolean"},
					"object":  {Type: "object"},
					"array":   {Type: "array"},
				},
				Required: []string{"string", "integer", "number", "boolean", "object", "array"},
			},
			json: `{
				"string": 123,
				"integer": "42",
				"number": true,
				"boolean": 1,
				"object": [1, 2, 3],
				"array": {"key": "value"}
			}`,
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run first validator
			validator1Result, validator1Err := validator1.Validate(tc.schema, tc.json)
			if validator1Err != nil {
				t.Fatalf("Validator 1 error: %v", validator1Err)
			}

			// Run second validator
			validator2Result, validator2Err := validator2.Validate(tc.schema, tc.json)
			if validator2Err != nil {
				t.Fatalf("Validator 2 error: %v", validator2Err)
			}

			// Test validity result matches expected
			if validator1Result.Valid != tc.valid {
				t.Errorf("Validator 1: expected valid=%v, got %v", tc.valid, validator1Result.Valid)
			}
			if validator2Result.Valid != tc.valid {
				t.Errorf("Validator 2: expected valid=%v, got %v", tc.valid, validator2Result.Valid)
			}

			// Test that both validators gave the same validity result
			if validator1Result.Valid != validator2Result.Valid {
				t.Errorf("Validators disagree: validator1=%v, validator2=%v", validator1Result.Valid, validator2Result.Valid)
			}

			// If invalid, check that both validators report errors
			if !tc.valid {
				if len(validator1Result.Errors) == 0 {
					t.Error("Validator 1: expected errors, got none")
				}
				if len(validator2Result.Errors) == 0 {
					t.Error("Validator 2: expected errors, got none")
				}

				// Check for equivalent error content (exact messages may differ)
				for _, v1Err := range validator1Result.Errors {
					// For each error from validator 1, check if validator 2 has a similar one
					found := false
					for _, v2Err := range validator2Result.Errors {
						// Check if they contain the same property references and error types
						if errorsEquivalent(v1Err, v2Err) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Validator 1 error not matched in validator 2: %s", v1Err)
						t.Errorf("Validator 2 errors: %v", validator2Result.Errors)
					}
				}
			}
		})
	}
}

// TestMemoryReuse tests that the validator properly reuses memory
// This verifies that the optimized validator's memory pooling works correctly
func TestMemoryReuse(t *testing.T) {
	validator := NewValidator()
	
	// Schema for validation
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
		Required: []string{"name", "age"},
	}
	
	// Valid JSON
	validJSON := `{"name": "John", "age": 30}`
	
	// First validation should succeed
	result1, err := validator.Validate(schema, validJSON)
	if err != nil {
		t.Fatalf("First validation failed: %v", err)
	}
	if !result1.Valid {
		t.Errorf("Expected valid result for first validation")
	}
	
	// Invalid JSON (missing required field)
	invalidJSON := `{"name": "Jane"}`
	
	// Second validation should fail
	result2, err := validator.Validate(schema, invalidJSON)
	if err != nil {
		t.Fatalf("Second validation failed: %v", err)
	}
	if result2.Valid {
		t.Errorf("Expected invalid result for second validation")
	}
	if len(result2.Errors) == 0 {
		t.Errorf("Expected errors for second validation")
	}
	
	// Check if the first result was modified (it shouldn't be)
	if !result1.Valid {
		t.Errorf("First result was modified incorrectly")
	}
	if len(result1.Errors) > 0 {
		t.Errorf("First result errors were modified incorrectly: %v", result1.Errors)
	}
	
	// Validate a third time
	result3, err := validator.Validate(schema, validJSON)
	if err != nil {
		t.Fatalf("Third validation failed: %v", err)
	}
	if !result3.Valid {
		t.Errorf("Expected valid result for third validation")
	}
	
	// Second result should still show invalid
	if result2.Valid {
		t.Errorf("Second result was modified incorrectly")
	}
}

// TestRegexCache tests that the regex cache is working
// This verifies that the optimized validator's regex pattern caching works correctly
func TestRegexCache(t *testing.T) {
	// Clear the regex cache
	RegexCache = sync.Map{}
	
	validator := NewValidator()
	
	// Schema with patterns
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"username": {
				Type:    "string",
				Pattern: "^[a-zA-Z0-9_]+$",
			},
			"email": {
				Type:   "string",
				Format: "email",
			},
		},
		Required: []string{"username", "email"},
	}
	
	// Valid JSON
	validJSON := `{"username": "john_doe", "email": "john@example.com"}`
	
	// First validation should populate the cache
	_, err := validator.Validate(schema, validJSON)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}
	
	// Check that patterns are in cache
	count := 0
	RegexCache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	
	if count == 0 {
		t.Errorf("Expected items in regex cache, found none")
	}
}

// errorsEquivalent checks if two validation error messages are semantically equivalent
func errorsEquivalent(err1, err2 string) bool {
	// Normalize both error strings
	err1Lower := strings.ToLower(strings.TrimSpace(err1))
	err2Lower := strings.ToLower(strings.TrimSpace(err2))
	
	// If the normalized strings are identical, they're equivalent
	if err1Lower == err2Lower {
		return true
	}
	
	// Extract property names from both errors
	prop1 := extractPropertyName(err1Lower)
	prop2 := extractPropertyName(err2Lower)
	
	// If both errors have property names, they should match
	if prop1 != "" && prop2 != "" {
		if prop1 != prop2 {
			// If property names don't match, errors are about different properties
			return false
		}
		// Property names match, now check error types
	}
	
	// Check if both errors contain the same validation term
	validationTerms := []string{
		"required", "missing", "string", "integer", "number", "boolean", "object", "array",
		"minimum", "maximum", "minlength", "maxlength", "pattern", "format", "enum",
	}
	
	// For each validation term, check if both errors contain it
	for _, term := range validationTerms {
		if strings.Contains(err1Lower, term) && strings.Contains(err2Lower, term) {
			return true
		}
	}
	
	// Special case for "must be" patterns that might differ in formatting
	if strings.Contains(err1Lower, "must be") && strings.Contains(err2Lower, "must be") {
		// Both errors are about constraints, likely the same one
		return true
	}
	
	// If both errors contain numeric values (especially for min/max constraints)
	containsNumber1 := containsNumeric(err1Lower)
	containsNumber2 := containsNumeric(err2Lower)
	if containsNumber1 && containsNumber2 {
		// If they both mention the same property and both have numbers,
		// they're likely about the same constraint
		if prop1 != "" && prop1 == prop2 {
			return true
		}
	}
	
	return false
}

// extractPropertyName tries to extract a property name from an error message
func extractPropertyName(err string) string {
	// First check for standard property patterns
	words := strings.Fields(err)
	for _, word := range words {
		// Look for property names, which might contain dots or brackets
		if strings.Contains(word, ".") || strings.Contains(word, "[") {
			return word
		}
	}
	
	// Also check for the first word that might be a property name
	for _, word := range words {
		// Skip common error terms
		if word == "property" || word == "field" || word == "is" || 
		   word == "must" || word == "be" || word == "a" || word == "an" ||
		   word == "the" || word == "required" || word == "missing" {
			continue
		}
		
		// The first word that's not a common term might be the property name
		return word
	}
	
	return ""
}

// containsNumeric checks if a string contains any numeric values
func containsNumeric(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

// containsAny checks if a string contains any of the given terms
func containsAny(s string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(strings.ToLower(s), term) {
			return true
		}
	}
	return false
}

// TestFormatValidation specifically tests format validation
func TestFormatValidation(t *testing.T) {
	validator := NewValidator()
	
	// Test scenarios for each format
	testCases := []struct {
		format  string
		valid   []string
		invalid []string
	}{
		{
			format: "email",
			valid: []string{
				"user@example.com",
				"john.doe@example.co.uk",
				"info+test@example.org",
			},
			invalid: []string{
				"invalid-email",
				"@example.com",
				"user@",
				"user@.com",
			},
		},
		{
			format: "uri",
			valid: []string{
				"https://example.com",
				"http://example.org/path",
				"ftp://files.example.net/file.txt",
			},
			invalid: []string{
				"example.com",
				"http:/example.com",
				"https:/",
				"://example.com",
			},
		},
		{
			format: "date-time",
			valid: []string{
				"2023-01-01T12:30:45Z",
				"2023-01-01T12:30:45.123Z",
				"2023-01-01T12:30:45+01:00",
			},
			invalid: []string{
				"2023-01-01",
				"12:30:45",
				"2023/01/01T12:30:45Z",
				"2023-01-01 12:30:45",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Format_%s", tc.format), func(t *testing.T) {
			// Schema with the format
			schema := &domain.Schema{
				Type: "object",
				Properties: map[string]domain.Property{
					"value": {
						Type:   "string",
						Format: tc.format,
					},
				},
				Required: []string{"value"},
			}
			
			// Test valid values
			for i, val := range tc.valid {
				jsonStr := fmt.Sprintf(`{"value": "%s"}`, val)
				result, err := validator.Validate(schema, jsonStr)
				if err != nil {
					t.Errorf("Valid %s #%d failed: %v", tc.format, i, err)
					continue
				}
				if !result.Valid {
					t.Errorf("Expected valid for '%s' format with '%s', got invalid: %v", 
						tc.format, val, result.Errors)
				}
			}
			
			// Test invalid values
			for i, val := range tc.invalid {
				jsonStr := fmt.Sprintf(`{"value": "%s"}`, val)
				result, err := validator.Validate(schema, jsonStr)
				if err != nil {
					t.Errorf("Invalid %s #%d failed: %v", tc.format, i, err)
					continue
				}
				if result.Valid {
					t.Errorf("Expected invalid for '%s' format with '%s', got valid", 
						tc.format, val)
				}
			}
		})
	}
}