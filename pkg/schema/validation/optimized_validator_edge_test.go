package validation

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	
	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestValidatorEdgeCases tests edge cases for the optimized validator
func TestValidatorEdgeCases(t *testing.T) {
	validator := NewOptimizedValidator()
	
	// Test invalid JSON input
	t.Run("InvalidJSON", func(t *testing.T) {
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"name": {Type: "string"},
			},
		}
		
		invalidJSON := `{"name": "test" "invalid"}`
		
		_, err := validator.Validate(schema, invalidJSON)
		if err == nil {
			t.Errorf("Expected error with invalid JSON, got nil")
		}
	})
	
	// Test nil schema
	t.Run("NilSchema", func(t *testing.T) {
		// No schema validation should pass
		result, err := validator.Validate(nil, `{"name": "test"}`)
		if err != nil {
			t.Errorf("Error with nil schema: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result with nil schema")
		}
	})
	
	// Test empty schema
	t.Run("EmptySchema", func(t *testing.T) {
		// Empty schema should pass
		schema := &domain.Schema{}
		
		result, err := validator.Validate(schema, `{"name": "test"}`)
		if err != nil {
			t.Errorf("Error with empty schema: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result with empty schema")
		}
	})
	
	// Test very large / complex schema
	t.Run("LargeComplexSchema", func(t *testing.T) {
		// Create a deep nested schema
		schema := createDeepNestedSchema(3) // 3 levels deep
		
		// Create valid data for this schema
		data := createValidDataForSchema(schema)
		jsonData, _ := json.Marshal(data)
		
		// Validate
		result, err := validator.Validate(schema, string(jsonData))
		if err != nil {
			t.Errorf("Error validating large schema: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result with large schema: %v", result.Errors)
		}
		
		// Validate with invalid data (missing required field)
		dataInvalid := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"value": "test",
					// missing "required" field
				},
				"value":    "test",
				"required": "present",
			},
			"value":    "test",
			"required": "present",
		}
		
		jsonDataInvalid, _ := json.Marshal(dataInvalid)
		
		resultInvalid, err := validator.Validate(schema, string(jsonDataInvalid))
		if err != nil {
			t.Errorf("Error validating invalid large schema: %v", err)
		}
		if resultInvalid.Valid {
			t.Errorf("Expected invalid result with broken large schema")
		}
	})
	
	// Test schema with all types
	t.Run("AllTypes", func(t *testing.T) {
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"string":  {Type: "string"},
				"integer": {Type: "integer"},
				"number":  {Type: "number"},
				"boolean": {Type: "boolean"},
				"array":   {Type: "array", Items: &domain.Property{Type: "string"}},
				"object":  {Type: "object", Properties: map[string]domain.Property{"key": {Type: "string"}}},
			},
			Required: []string{"string", "integer", "number", "boolean", "array", "object"},
		}
		
		validJSON := `{
			"string": "test",
			"integer": 42,
			"number": 3.14,
			"boolean": true,
			"array": ["one", "two"],
			"object": {"key": "value"}
		}`
		
		result, err := validator.Validate(schema, validJSON)
		if err != nil {
			t.Errorf("Error validating all types: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result with all types: %v", result.Errors)
		}
	})
	
	// Test regex pattern error handling
	t.Run("InvalidRegexPattern", func(t *testing.T) {
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"field": {
					Type:    "string",
					Pattern: "[[[ invalid regex", // Invalid regex pattern
				},
			},
			Required: []string{"field"},
		}
		
		json := `{"field": "test"}`
		
		result, err := validator.Validate(schema, json)
		if err != nil {
			t.Errorf("Error validating with invalid regex: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result with invalid regex pattern")
		}
		
		// Check that an error related to the invalid pattern is reported
		foundPatternError := false
		for _, e := range result.Errors {
			if containsAny(e, "pattern", "invalid") {
				foundPatternError = true
				break
			}
		}
		if !foundPatternError {
			t.Errorf("Expected error about invalid pattern, got: %v", result.Errors)
		}
	})
	
	// Test consecutive validation results are independent
	t.Run("ConsecutiveValidationIndependent", func(t *testing.T) {
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"field": {Type: "string"},
			},
			Required: []string{"field"},
		}
		
		// First validation (valid)
		validJSON := `{"field": "test"}`
		result1, _ := validator.Validate(schema, validJSON)
		
		// Second validation (invalid)
		invalidJSON := `{}`
		result2, _ := validator.Validate(schema, invalidJSON)
		
		// Third validation (valid again)
		result3, _ := validator.Validate(schema, validJSON)
		
		// Ensure each result is as expected and not affected by others
		if !result1.Valid {
			t.Errorf("First validation should be valid")
		}
		if result2.Valid {
			t.Errorf("Second validation should be invalid")
		}
		if !result3.Valid {
			t.Errorf("Third validation should be valid")
		}
		
		// Check result1 and result3 are distinct objects
		if result1 == result3 {
			t.Errorf("Results should be separate objects")
		}
	})
	
	// Test multiple format validation
	t.Run("MultipleFormatValidation", func(t *testing.T) {
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"email": {
					Type:   "string",
					Format: "email",
				},
				"uri": {
					Type:   "string",
					Format: "uri",
				},
				"datetime": {
					Type:   "string",
					Format: "date-time",
				},
			},
			Required: []string{"email", "uri", "datetime"},
		}
		
		// All valid formats
		validJSON := `{
			"email": "user@example.com",
			"uri": "https://example.com",
			"datetime": "2023-01-01T12:00:00Z"
		}`
		
		result, err := validator.Validate(schema, validJSON)
		if err != nil {
			t.Errorf("Error validating formats: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result with all formats: %v", result.Errors)
		}
		
		// All invalid formats
		invalidJSON := `{
			"email": "not-an-email",
			"uri": "example.com",
			"datetime": "2023-01-01"
		}`
		
		result, err = validator.Validate(schema, invalidJSON)
		if err != nil {
			t.Errorf("Error validating invalid formats: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result with all invalid formats")
		}
		
		// Should have 3 errors (one for each invalid format)
		if len(result.Errors) != 3 {
			t.Errorf("Expected 3 format errors, got %d: %v", len(result.Errors), result.Errors)
		}
	})
	
	// Test with unsupported format
	t.Run("UnsupportedFormat", func(t *testing.T) {
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"field": {
					Type:   "string",
					Format: "unsupported-format",
				},
			},
			Required: []string{"field"},
		}
		
		json := `{"field": "test"}`
		
		result, err := validator.Validate(schema, json)
		if err != nil {
			t.Errorf("Error validating with unsupported format: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result with unsupported format")
		}
		
		// Check that an error about unsupported format is reported
		foundFormatError := false
		for _, e := range result.Errors {
			if containsAny(e, "format", "unsupported") {
				foundFormatError = true
				break
			}
		}
		if !foundFormatError {
			t.Errorf("Expected error about unsupported format, got: %v", result.Errors)
		}
	})
	
	// Test with very long error messages (pool reuse)
	t.Run("VeryLongPropertyPaths", func(t *testing.T) {
		// Create a deep nested schema with very long property names
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"veryLongPropertyNameThatMightCauseIssuesWithBuffers": {
					Type: "object",
					Properties: map[string]domain.Property{
						"anotherVeryLongPropertyNameThatAddsToThePathLength": {
							Type: "object",
							Properties: map[string]domain.Property{
								"yetAnotherLongPropertyToMakePathsEvenLonger": {
									Type: "string",
									MinLength: intPtr(100),
								},
							},
							Required: []string{"yetAnotherLongPropertyToMakePathsEvenLonger"},
						},
					},
					Required: []string{"anotherVeryLongPropertyNameThatAddsToThePathLength"},
				},
			},
			Required: []string{"veryLongPropertyNameThatMightCauseIssuesWithBuffers"},
		}
		
		// Create JSON with invalid string (too short)
		json := `{
			"veryLongPropertyNameThatMightCauseIssuesWithBuffers": {
				"anotherVeryLongPropertyNameThatAddsToThePathLength": {
					"yetAnotherLongPropertyToMakePathsEvenLonger": "short"
				}
			}
		}`
		
		// Run validation - this should produce errors with very long property paths
		result, err := validator.Validate(schema, json)
		if err != nil {
			t.Errorf("Error validating with long paths: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result with string length violation")
		}
		
		// No need to check specific error messages, just ensure validation completed
		
		// Do a second validation to test pool reuse with these long strings
		result2, err := validator.Validate(schema, json)
		if err != nil {
			t.Errorf("Error on second validation with long paths: %v", err)
		}
		if result2.Valid {
			t.Errorf("Expected invalid result on second validation")
		}
	})
	
	// Test regex cache functionality
	t.Run("RegexCacheReuse", func(t *testing.T) {
		// Clear regex cache
		RegexCache = sync.Map{}
		
		// Create a schema with a regex pattern
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"username": {
					Type:    "string",
					Pattern: "^[a-zA-Z0-9_]{3,20}$",
				},
			},
			Required: []string{"username"},
		}
		
		// Valid JSON
		validJSON := `{"username": "valid_user123"}`
		
		// First validation should add to cache
		_, err := validator.Validate(schema, validJSON)
		if err != nil {
			t.Errorf("Error validating username: %v", err)
		}
		
		// Check that the pattern is in cache
		count := 0
		RegexCache.Range(func(key, value interface{}) bool {
			if key.(string) == "^[a-zA-Z0-9_]{3,20}$" {
				count++
			}
			return true
		})
		
		if count == 0 {
			t.Errorf("Expected pattern to be in regex cache")
		}
		
		// Run multiple validations with the same pattern
		for i := 0; i < 10; i++ {
			_, err := validator.Validate(schema, validJSON)
			if err != nil {
				t.Errorf("Error on validation %d: %v", i, err)
			}
		}
		
		// Cache should still have the same number of entries for this pattern
		newCount := 0
		RegexCache.Range(func(key, value interface{}) bool {
			if key.(string) == "^[a-zA-Z0-9_]{3,20}$" {
				newCount++
			}
			return true
		})
		
		if newCount != count {
			t.Errorf("Expected same number of cache entries, got %d instead of %d", newCount, count)
		}
	})
	
	// Test struct validation
	t.Run("StructValidation", func(t *testing.T) {
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"name":  {Type: "string"},
				"age":   {Type: "integer", Minimum: float64Ptr(18)},
				"email": {Type: "string", Format: "email"},
			},
			Required: []string{"name", "age", "email"},
		}
		
		// Define a struct matching the schema
		type Person struct {
			Name  string `json:"name"`
			Age   int    `json:"age"`
			Email string `json:"email"`
		}
		
		// Valid struct
		validPerson := Person{
			Name:  "John Doe",
			Age:   30,
			Email: "john@example.com",
		}
		
		result, err := validator.ValidateStruct(schema, validPerson)
		if err != nil {
			t.Errorf("Error validating valid struct: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result for valid struct: %v", result.Errors)
		}
		
		// Invalid struct
		invalidPerson := Person{
			Name:  "Young Person",
			Age:   16, // Below minimum
			Email: "invalid-email", // Invalid email format
		}
		
		result, err = validator.ValidateStruct(schema, invalidPerson)
		if err != nil {
			t.Errorf("Error validating invalid struct: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for invalid struct")
		}
		
		// Should have 2 errors (age and email)
		if len(result.Errors) != 2 {
			t.Errorf("Expected 2 errors, got %d: %v", len(result.Errors), result.Errors)
		}
	})
}


// createDeepNestedSchema creates a deeply nested schema for testing
func createDeepNestedSchema(depth int) *domain.Schema {
	if depth <= 0 {
		return &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"value":    {Type: "string"},
				"required": {Type: "string"},
			},
			Required: []string{"required"},
		}
	}
	
	levelName := fmt.Sprintf("level%d", depth)
	nextSchema := createDeepNestedSchema(depth - 1)
	
	return &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			levelName: {
				Type:       "object",
				Properties: nextSchema.Properties,
				Required:   nextSchema.Required,
			},
			"value":    {Type: "string"},
			"required": {Type: "string"},
		},
		Required: []string{"required", levelName},
	}
}

// createValidDataForSchema creates valid test data for a nested schema
func createValidDataForSchema(schema *domain.Schema) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Add required fields
	for _, req := range schema.Required {
		if req == "required" {
			result["required"] = "present"
		} else if prop, ok := schema.Properties[req]; ok {
			switch prop.Type {
			case "object":
				if prop.Properties != nil {
					// Create a schema for the nested object
					nestedSchema := &domain.Schema{
						Type:       "object",
						Properties: prop.Properties,
						Required:   prop.Required,
					}
					result[req] = createValidDataForSchema(nestedSchema)
				} else {
					result[req] = map[string]interface{}{"key": "value"}
				}
			case "array":
				result[req] = []interface{}{"item1", "item2"}
			case "string":
				result[req] = "test string"
			case "integer":
				result[req] = 42
			case "number":
				result[req] = 3.14
			case "boolean":
				result[req] = true
			default:
				result[req] = "fallback value"
			}
		}
	}
	
	// Add a generic value field if present in schema
	if _, ok := schema.Properties["value"]; ok {
		result["value"] = "test value"
	}
	
	return result
}