package processor

import (
	"strings"
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestEnhancePromptWithSchema(t *testing.T) {
	enhancer := NewPromptEnhancer()

	t.Run("simple object schema", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string", Description: "Person's name"},
				"age":  {Type: "integer", Description: "Person's age", Minimum: float64Ptr(0)},
			},
			Required: []string{"name"},
		}

		prompt := "Generate information about a person"
		enhanced, err := enhancer.Enhance(prompt, schema)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check that the original prompt is included
		if !strings.Contains(enhanced, prompt) {
			t.Errorf("Enhanced prompt should contain the original prompt")
		}

		// Check that schema is included
		if !strings.Contains(enhanced, "JSON schema") {
			t.Errorf("Enhanced prompt should mention JSON schema")
		}

		// Check for specific schema elements
		if !strings.Contains(enhanced, "object") {
			t.Errorf("Enhanced prompt should contain schema type")
		}

		if !strings.Contains(enhanced, "name") {
			t.Errorf("Enhanced prompt should contain property name")
		}

		if !strings.Contains(enhanced, "Person's name") {
			t.Errorf("Enhanced prompt should contain property description")
		}

		if !strings.Contains(enhanced, "required") {
			t.Errorf("Enhanced prompt should mention required fields")
		}

		// Check for instructions on output format
		if !strings.Contains(enhanced, "valid JSON") {
			t.Errorf("Enhanced prompt should instruct to output valid JSON")
		}
	})

	t.Run("array schema", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "array",
			Properties: map[string]schemaDomain.Property{
				"": {
					Type: "array",
					Items: &schemaDomain.Property{
						Type: "string",
					},
				},
			},
		}

		prompt := "Generate a list of fruits"
		enhanced, err := enhancer.Enhance(prompt, schema)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check that the original prompt is included
		if !strings.Contains(enhanced, prompt) {
			t.Errorf("Enhanced prompt should contain the original prompt")
		}

		// Check for array-specific instructions
		if !strings.Contains(enhanced, "array") {
			t.Errorf("Enhanced prompt should mention array type")
		}
	})

	t.Run("complex nested schema", func(t *testing.T) {
		addressProperty := schemaDomain.Property{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"street":  {Type: "string", Description: "Street address"},
				"city":    {Type: "string", Description: "City name"},
				"zipCode": {Type: "string", Description: "ZIP or postal code"},
			},
		}

		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name":    {Type: "string", Description: "Person's name"},
				"age":     {Type: "integer", Description: "Person's age", Minimum: float64Ptr(0)},
				"address": addressProperty,
				"tags": {
					Type: "array",
					Items: &schemaDomain.Property{
						Type: "string",
					},
					Description: "List of tags",
				},
			},
			Required: []string{"name", "address"},
		}

		prompt := "Generate a person with address details"
		enhanced, err := enhancer.Enhance(prompt, schema)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check for nested schema elements
		if !strings.Contains(enhanced, "address") {
			t.Errorf("Enhanced prompt should contain nested object field")
		}

		// Check for array property
		if !strings.Contains(enhanced, "tags") {
			t.Errorf("Enhanced prompt should contain array field")
		}
	})

	t.Run("schema with enum", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string", Description: "Person's name"},
				"gender": {
					Type:        "string",
					Description: "Person's gender",
					Enum:        []string{"male", "female", "non-binary", "other"},
				},
			},
			Required: []string{"name", "gender"},
		}

		prompt := "Generate a person's information"
		enhanced, err := enhancer.Enhance(prompt, schema)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check for enum values
		for _, value := range []string{"male", "female", "non-binary", "other"} {
			if !strings.Contains(enhanced, value) {
				t.Errorf("Enhanced prompt should contain enum value: %s", value)
			}
		}
	})
}

func TestEnhancePromptWithOptions(t *testing.T) {
	enhancer := NewPromptEnhancer()

	t.Run("with custom instructions", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string", Description: "Person's name"},
				"age":  {Type: "integer", Description: "Person's age", Minimum: float64Ptr(0)},
			},
			Required: []string{"name"},
		}

		prompt := "Generate information about a person"
		options := map[string]interface{}{
			"instructions": "Make sure the name is realistic and age is appropriate",
			"format":       "markdown",
		}

		enhanced, err := enhancer.EnhanceWithOptions(prompt, schema, options)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check that custom instructions are included
		if !strings.Contains(enhanced, "realistic") {
			t.Errorf("Enhanced prompt should contain custom instructions")
		}

		// Check that format instructions are included
		if !strings.Contains(enhanced, "markdown") {
			t.Errorf("Enhanced prompt should contain format instructions")
		}
	})

	t.Run("with examples", func(t *testing.T) {
		schema := &schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"name": {Type: "string", Description: "Person's name"},
				"age":  {Type: "integer", Description: "Person's age", Minimum: float64Ptr(0)},
			},
			Required: []string{"name"},
		}

		prompt := "Generate information about a person"
		options := map[string]interface{}{
			"examples": []map[string]interface{}{
				{"name": "John Doe", "age": 30},
				{"name": "Jane Smith", "age": 25},
			},
		}

		enhanced, err := enhancer.EnhanceWithOptions(prompt, schema, options)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check that examples are included
		if !strings.Contains(enhanced, "John Doe") {
			t.Errorf("Enhanced prompt should contain example data")
		}

		if !strings.Contains(enhanced, "Jane Smith") {
			t.Errorf("Enhanced prompt should contain example data")
		}
	})
}
