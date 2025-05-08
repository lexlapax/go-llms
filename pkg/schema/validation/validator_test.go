package validation

import (
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Helper function to check if an error array contains a specific error
func containsError(errors []string, field, errType string) bool {
	for _, err := range errors {
		if strings.Contains(strings.ToLower(err), strings.ToLower(field)) &&
			strings.Contains(strings.ToLower(err), strings.ToLower(errType)) {
			return true
		}
	}
	return false
}

func TestObjectValidation(t *testing.T) {
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"name": {Type: "string", Required: []string{}},
			"age":  {Type: "integer", Minimum: float64Ptr(0)},
		},
		Required: []string{"name"},
	}

	validator := NewValidator()

	// Valid case
	t.Run("valid object", func(t *testing.T) {
		input := `{"name": "John", "age": 30}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
		}
	})

	// Invalid case - missing required field
	t.Run("invalid object - missing required", func(t *testing.T) {
		input := `{"age": 30}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for missing required field")
		}
		if !containsError(result.Errors, "name", "required") {
			t.Errorf("Expected 'required' error for 'name', got: %v", result.Errors)
		}
	})

	// Invalid case - type mismatch
	t.Run("invalid object - type mismatch", func(t *testing.T) {
		input := `{"name": "John", "age": "thirty"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for type mismatch")
		}
		if !containsError(result.Errors, "age", "integer") {
			t.Errorf("Expected type mismatch error for 'age', got: %v", result.Errors)
		}
	})

	// Invalid case - value constraint
	t.Run("invalid object - minimum constraint", func(t *testing.T) {
		input := `{"name": "John", "age": -5}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for minimum constraint violation")
		}
		if !containsError(result.Errors, "age", "at least") {
			t.Errorf("Expected minimum constraint error for 'age', got: %v", result.Errors)
		}
	})
}

func TestStringValidation(t *testing.T) {
	// Test schema with string constraints
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"username": {
				Type:      "string",
				MinLength: intPtr(3),
				MaxLength: intPtr(20),
				Pattern:   "^[a-zA-Z0-9_]+$",
			},
			"email": {
				Type:   "string",
				Format: "email",
			},
		},
		Required: []string{"username", "email"},
	}

	validator := NewValidator()

	// Valid case
	t.Run("valid string properties", func(t *testing.T) {
		input := `{"username": "john_doe", "email": "john@example.com"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
		}
	})

	// Test too short
	t.Run("invalid string - too short", func(t *testing.T) {
		input := `{"username": "jo", "email": "john@example.com"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for short username")
		}
		if !containsError(result.Errors, "username", "at least") {
			t.Errorf("Expected minLength error for 'username', got: %v", result.Errors)
		}
	})

	// Test too long
	t.Run("invalid string - too long", func(t *testing.T) {
		input := `{"username": "john_doe_with_very_long_name", "email": "john@example.com"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for long username")
		}
		if !containsError(result.Errors, "username", "no more than") {
			t.Errorf("Expected maxLength error for 'username', got: %v", result.Errors)
		}
	})

	// Test pattern mismatch
	t.Run("invalid string - pattern mismatch", func(t *testing.T) {
		input := `{"username": "john-doe", "email": "john@example.com"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for pattern mismatch")
		}
		if !containsError(result.Errors, "username", "pattern") {
			t.Errorf("Expected pattern error for 'username', got: %v", result.Errors)
		}
	})

	// Test invalid email format
	t.Run("invalid string - email format", func(t *testing.T) {
		input := `{"username": "john_doe", "email": "invalid-email"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for email format")
		}
		if !containsError(result.Errors, "email", "valid email") {
			t.Errorf("Expected format error for 'email', got: %v", result.Errors)
		}
	})
}

func TestDateTimeValidation(t *testing.T) {
	// Test schema with date-time format
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"timestamp": {
				Type:   "string",
				Format: "date-time",
			},
		},
		Required: []string{"timestamp"},
	}

	validator := NewValidator()

	// Valid date-time format
	t.Run("valid date-time format", func(t *testing.T) {
		input := `{"timestamp": "2023-05-01T14:30:00Z"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
		}
	})

	// Valid date-time with decimal seconds
	t.Run("valid date-time with decimal seconds", func(t *testing.T) {
		input := `{"timestamp": "2023-05-01T14:30:00.123Z"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
		}
	})

	// Valid date-time with timezone offset
	t.Run("valid date-time with timezone offset", func(t *testing.T) {
		input := `{"timestamp": "2023-05-01T14:30:00+01:00"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
		}
	})

	// Invalid date-time format
	t.Run("invalid date-time format", func(t *testing.T) {
		input := `{"timestamp": "2023-05-01 14:30:00"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for incorrect date-time format")
		}
		if !containsError(result.Errors, "timestamp", "valid ISO8601") {
			t.Errorf("Expected date-time format error for 'timestamp', got: %v", result.Errors)
		}
	})

	// Missing timezone in date-time
	t.Run("invalid date-time - missing timezone", func(t *testing.T) {
		input := `{"timestamp": "2023-05-01T14:30:00"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for date-time without timezone")
		}
		if !containsError(result.Errors, "timestamp", "valid ISO8601") {
			t.Errorf("Expected date-time format error for 'timestamp', got: %v", result.Errors)
		}
	})
}

func TestURIValidation(t *testing.T) {
	// Test schema with URI format
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"website": {
				Type:   "string",
				Format: "uri",
			},
		},
		Required: []string{"website"},
	}

	validator := NewValidator()

	// Valid HTTP URI
	t.Run("valid http URI", func(t *testing.T) {
		input := `{"website": "http://example.com"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
		}
	})

	// Valid HTTPS URI with path and query
	t.Run("valid https URI with path and query", func(t *testing.T) {
		input := `{"website": "https://example.com/path/to/resource?query=param"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
		}
	})

	// Valid FTP URI
	t.Run("valid ftp URI", func(t *testing.T) {
		input := `{"website": "ftp://files.example.com"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
		}
	})

	// Invalid URI - missing scheme
	t.Run("invalid URI - missing scheme", func(t *testing.T) {
		input := `{"website": "example.com"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for URI without scheme")
		}
		if !containsError(result.Errors, "website", "valid URI") {
			t.Errorf("Expected URI format error for 'website', got: %v", result.Errors)
		}
	})

	// Invalid URI - unsupported scheme
	t.Run("invalid URI - unsupported scheme", func(t *testing.T) {
		input := `{"website": "file:///path/to/file"}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for URI with unsupported scheme")
		}
		if !containsError(result.Errors, "website", "valid URI") {
			t.Errorf("Expected URI format error for 'website', got: %v", result.Errors)
		}
	})
}

func TestNumberValidation(t *testing.T) {
	// Test schema with number constraints
	schema := &domain.Schema{
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
	}

	validator := NewValidator()

	// Valid case
	t.Run("valid number properties", func(t *testing.T) {
		input := `{"age": 30, "score": 85.5}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
		}
	})

	// Test below minimum
	t.Run("invalid number - below minimum", func(t *testing.T) {
		input := `{"age": 15, "score": 85.5}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for age below minimum")
		}
		if !containsError(result.Errors, "age", "at least") {
			t.Errorf("Expected minimum error for 'age', got: %v", result.Errors)
		}
	})

	// Test above maximum
	t.Run("invalid number - above maximum", func(t *testing.T) {
		input := `{"age": 30, "score": 120}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for score above maximum")
		}
		if !containsError(result.Errors, "score", "at most") {
			t.Errorf("Expected maximum error for 'score', got: %v", result.Errors)
		}
	})

	// Test type mismatch (string instead of number)
	t.Run("invalid number - type mismatch", func(t *testing.T) {
		input := `{"age": "thirty", "score": 85.5}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for type mismatch")
		}
		if !containsError(result.Errors, "age", "integer") {
			t.Errorf("Expected type error for 'age', got: %v", result.Errors)
		}
	})
}

func TestArrayValidation(t *testing.T) {
	// Test schema with array constraints
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"tags": {
				Type: "array",
				Items: &domain.Property{
					Type: "string",
				},
			},
			"scores": {
				Type: "array",
				Items: &domain.Property{
					Type:    "number",
					Minimum: float64Ptr(0),
					Maximum: float64Ptr(100),
				},
			},
		},
	}

	validator := NewValidator()

	// Valid case
	t.Run("valid array properties", func(t *testing.T) {
		input := `{"tags": ["golang", "validation", "json"], "scores": [85, 90, 75.5]}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Valid {
			t.Errorf("Expected valid result, got validation errors: %v", result.Errors)
		}
	})

	// Test array with invalid item type
	t.Run("invalid array - item type mismatch", func(t *testing.T) {
		input := `{"tags": ["golang", 123, "json"], "scores": [85, 90, 75.5]}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for item type mismatch")
		}
		if !containsError(result.Errors, "tags[1]", "string") {
			t.Errorf("Expected type error for 'tags[1]', got: %v", result.Errors)
		}
	})

	// Test array with item constraint violation
	t.Run("invalid array - item constraint violation", func(t *testing.T) {
		input := `{"tags": ["golang", "validation", "json"], "scores": [85, 110, 75.5]}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for item constraint violation")
		}
		if !containsError(result.Errors, "scores[1]", "at most") {
			t.Errorf("Expected maximum error for 'scores[1]', got: %v", result.Errors)
		}
	})

	// Test non-array value
	t.Run("invalid array - not an array", func(t *testing.T) {
		input := `{"tags": "not-an-array", "scores": [85, 90, 75.5]}`
		result, err := validator.Validate(schema, input)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Valid {
			t.Errorf("Expected invalid result for non-array value")
		}
		if !containsError(result.Errors, "tags", "array") {
			t.Errorf("Expected type error for 'tags', got: %v", result.Errors)
		}
	})
}
