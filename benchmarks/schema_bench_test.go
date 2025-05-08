package benchmarks

import (
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/testutils"
)

// Use helper functions from testutils package
var intPtr = testutils.IntPtr
var float64Ptr = testutils.Float64Ptr

// BenchmarkStringValidation benchmarks string validation performance
func BenchmarkStringValidation(b *testing.B) {
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type:      "string",
				MinLength: intPtr(3),
				MaxLength: intPtr(50),
				Pattern:   "^[a-zA-Z ]+$",
			},
			"email": {
				Type:   "string",
				Format: "email",
			},
			"website": {
				Type:   "string",
				Format: "uri",
			},
			"category": {
				Type: "string",
				Enum: []string{"A", "B", "C", "D"},
			},
		},
		Required: []string{"name", "email"},
	}

	validJSON := `{
		"name": "John Doe",
		"email": "john@example.com",
		"website": "https://example.com",
		"category": "A"
	}`

	validator := validation.NewValidator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.Validate(schema, validJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNestedObjectValidation benchmarks nested object validation performance
func BenchmarkNestedObjectValidation(b *testing.B) {
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"person": {
				Type: "object",
				Properties: map[string]schemaDomain.Property{
					"name": {
						Type:      "string",
						MinLength: intPtr(3),
					},
					"age": {
						Type:    "integer",
						Minimum: float64Ptr(0),
						Maximum: float64Ptr(120),
					},
					"address": {
						Type: "object",
						Properties: map[string]schemaDomain.Property{
							"street": {Type: "string"},
							"city":   {Type: "string"},
							"state":  {Type: "string"},
							"zip":    {Type: "string"},
						},
						Required: []string{"street", "city"},
					},
				},
				Required: []string{"name", "age"},
			},
			"contacts": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"type":  {Type: "string", Enum: []string{"phone", "email", "social"}},
						"value": {Type: "string"},
					},
					Required: []string{"type", "value"},
				},
			},
		},
		Required: []string{"person"},
	}

	validJSON := `{
		"person": {
			"name": "John Doe",
			"age": 35,
			"address": {
				"street": "123 Main St",
				"city": "Anytown",
				"state": "CA",
				"zip": "12345"
			}
		},
		"contacts": [
			{
				"type": "phone",
				"value": "555-123-4567"
			},
			{
				"type": "email",
				"value": "john@example.com"
			}
		]
	}`

	validator := validation.NewValidator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.Validate(schema, validJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkArrayValidation benchmarks array validation performance
func BenchmarkArrayValidation(b *testing.B) {
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"items": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"id":    {Type: "integer"},
						"name":  {Type: "string"},
						"value": {Type: "number"},
						"tags": {
							Type: "array",
							Items: &schemaDomain.Property{
								Type: "string",
							},
						},
					},
					Required: []string{"id", "name"},
				},
			},
		},
		Required: []string{"items"},
	}

	validJSON := `{
		"items": [
			{
				"id": 1,
				"name": "Item 1",
				"value": 10.5,
				"tags": ["tag1", "tag2"]
			},
			{
				"id": 2,
				"name": "Item 2",
				"value": 20.75,
				"tags": ["tag2", "tag3", "tag4"]
			},
			{
				"id": 3,
				"name": "Item 3",
				"value": 15.25,
				"tags": ["tag1", "tag5"]
			}
		]
	}`

	validator := validation.NewValidator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.Validate(schema, validJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidationWithErrors benchmarks validation with errors
func BenchmarkValidationWithErrors(b *testing.B) {
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type:      "string",
				MinLength: intPtr(5), // Will cause an error with the data below
				Pattern:   "^[a-zA-Z ]+$",
			},
			"age": {
				Type:    "integer",
				Minimum: float64Ptr(18), // Will cause an error with the data below
			},
			"email": {
				Type:   "string",
				Format: "email",
			},
		},
		Required: []string{"name", "age", "email", "missing"}, // Missing field will cause error
	}

	invalidJSON := `{
		"name": "Joe", 
		"age": 16,
		"email": "not-an-email"
	}`

	validator := validation.NewValidator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := validator.Validate(schema, invalidJSON)
		if err != nil {
			b.Fatal(err)
		}
		if result.Valid {
			b.Fatal("expected validation to fail")
		}
	}
}

// BenchmarkRepeatedValidation tests validators with repeated validations (caching benefit)
func BenchmarkRepeatedValidation(b *testing.B) {
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
		Required: []string{"name"},
	}

	validJSONs := []string{
		`{"name": "Alice", "age": 25}`,
		`{"name": "Bob", "age": 30}`,
		`{"name": "Charlie", "age": 35}`,
		`{"name": "David", "age": 40}`,
	}

	validator := validation.NewValidator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate between different JSONs to test caching
		jsonStr := validJSONs[i%len(validJSONs)]
		_, err := validator.Validate(schema, jsonStr)
		if err != nil {
			b.Fatal(err)
		}
	}
}
