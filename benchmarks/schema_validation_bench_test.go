package benchmarks

import (
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	"github.com/lexlapax/go-llms/pkg/testutils"
)

// BenchmarkStringValidation benchmarks string validation performance
func BenchmarkStringValidation(b *testing.B) {
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type:      "string",
				MinLength: testutils.IntPtr(3),
				MaxLength: testutils.IntPtr(50),
				Pattern:   "^[a-zA-Z ]+$",
			},
			"email": {
				Type:   "string",
				Format: "email",
			},
		},
		Required: []string{"name", "email"},
	}

	validator := validation.NewValidator()
	validJSON := `{"name": "John Doe", "email": "john@example.com"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.Validate(schema, validJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNestedObjectValidation benchmarks validation of complex nested objects
func BenchmarkNestedObjectValidation(b *testing.B) {
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"person": {
				Type: "object",
				Properties: map[string]schemaDomain.Property{
					"name": {
						Type:      "string",
						MinLength: testutils.IntPtr(3),
					},
					"age": {
						Type:    "integer",
						Minimum: testutils.Float64Ptr(0),
						Maximum: testutils.Float64Ptr(120),
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

	validator := validation.NewValidator()
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.Validate(schema, validJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkArrayValidation benchmarks validation of array items
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
				// Note: MinItems and MaxItems are not supported in the current schema implementation
			},
		},
		Required: []string{"items"},
	}

	validator := validation.NewValidator()
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.Validate(schema, validJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

