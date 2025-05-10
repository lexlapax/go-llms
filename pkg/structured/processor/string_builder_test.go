package processor

import (
	"strings"
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestOptimizedStringBuilder(t *testing.T) {
	// Create a simple test schema
	schema := &schemaDomain.Schema{
		Type:        "object",
		Description: "A test schema",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string", Description: "The name"},
			"age":  {Type: "integer", Description: "The age"},
		},
		Required: []string{"name"},
	}

	// Get schema JSON for capacity estimation
	schemaJSON, err := getSchemaJSON(schema)
	if err != nil {
		t.Fatalf("Failed to get schema JSON: %v", err)
	}

	// Test standard string builder vs optimized builder
	prompt := "Generate a user profile"
	t.Run("CompareBuilders", func(t *testing.T) {
		// Standard builder
		var stdBuilder strings.Builder
		if _, err := stdBuilder.WriteString(prompt); err != nil {
			t.Fatalf("Failed to write to standard builder: %v", err)
		}
		if _, err := stdBuilder.WriteString("\n"); err != nil {
			t.Fatalf("Failed to write newline to standard builder: %v", err)
		}
		if _, err := stdBuilder.WriteString("Schema type: "); err != nil {
			t.Fatalf("Failed to write schema type prefix to standard builder: %v", err)
		}
		if _, err := stdBuilder.WriteString(schema.Type); err != nil {
			t.Fatalf("Failed to write schema type to standard builder: %v", err)
		}
		if _, err := stdBuilder.WriteString("\n"); err != nil {
			t.Fatalf("Failed to write final newline to standard builder: %v", err)
		}
		stdResult := stdBuilder.String()

		// Optimized builder
		optBuilder := NewSchemaPromptBuilder(prompt, schema, len(schemaJSON))
		if _, err := optBuilder.WriteString(prompt); err != nil {
			t.Fatalf("Failed to write to optimized builder: %v", err)
		}
		if _, err := optBuilder.WriteString("\n"); err != nil {
			t.Fatalf("Failed to write newline to optimized builder: %v", err)
		}
		if _, err := optBuilder.WriteString("Schema type: "); err != nil {
			t.Fatalf("Failed to write schema type prefix to optimized builder: %v", err)
		}
		if _, err := optBuilder.WriteString(schema.Type); err != nil {
			t.Fatalf("Failed to write schema type to optimized builder: %v", err)
		}
		if _, err := optBuilder.WriteString("\n"); err != nil {
			t.Fatalf("Failed to write final newline to optimized builder: %v", err)
		}
		optResult := optBuilder.String()

		// Check that the results are identical
		if stdResult != optResult {
			t.Errorf("Builder outputs don't match:\nStd: %s\nOpt: %s", stdResult, optResult)
		}

		// Check that the optimized builder has a reasonable capacity
		if optBuilder.Cap() < len(optResult) {
			t.Errorf("Optimized builder capacity is too small: cap=%d, len=%d",
				optBuilder.Cap(), len(optResult))
		}
	})

	t.Run("EnhanceWithOptions", func(t *testing.T) {
		// Create a prompt enhancer
		enhancer := NewPromptEnhancer()

		// Test with examples option
		examples := []map[string]interface{}{
			{
				"name": "John Doe",
				"age":  30,
			},
			{
				"name": "Jane Smith",
				"age":  25,
			},
		}

		options := map[string]interface{}{
			"instructions": "Provide a complete profile",
			"format":       "JSON",
			"examples":     examples,
		}

		// Test enhancing with options
		result, err := enhancer.EnhanceWithOptions(prompt, schema, options)
		if err != nil {
			t.Fatalf("Failed to enhance prompt with options: %v", err)
		}

		// Verify the result contains all the expected parts
		checks := []string{
			prompt,
			"Please provide your response as a valid JSON object",
			"The required fields are: name",
			"Additional instructions: Provide a complete profile",
			"Format your response as JSON",
			"Example 1:",
			"Example 2:",
			"John Doe",
			"Jane Smith",
		}

		for _, check := range checks {
			if !strings.Contains(result, check) {
				t.Errorf("Expected result to contain '%s', but it didn't", check)
			}
		}
	})

	t.Run("CapacityEstimation", func(t *testing.T) {
		complexSchema := createComplexTestSchema()

		// Get schema JSON for capacity estimation
		complexSchemaJSON, err := getSchemaJSON(complexSchema)
		if err != nil {
			t.Fatalf("Failed to get schema JSON: %v", err)
		}

		// Test with a complex schema
		prompt := "Generate a detailed recipe"
		optBuilder := NewSchemaPromptBuilder(prompt, complexSchema, len(complexSchemaJSON))

		// Write a large amount of content
		for i := 0; i < 50; i++ {
			if _, err := optBuilder.WriteString("Line " + strings.Repeat("abcdefghij", 2) + "\n"); err != nil {
				t.Fatalf("Failed to write content line: %v", err)
			}
		}

		// Check that we didn't have to resize (capacity >= length)
		finalLength := optBuilder.Len()
		finalCapacity := optBuilder.Cap()

		if finalCapacity < finalLength {
			t.Errorf("Builder needed to resize: cap=%d, len=%d", finalCapacity, finalLength)
		}

		// For debugging: see how much extra capacity we allocated
		// t.Logf("Length: %d, Capacity: %d, Ratio: %.2f",
		//    finalLength, finalCapacity, float64(finalCapacity)/float64(finalLength))
	})
}

// Helper to create a complex test schema
func createComplexTestSchema() *schemaDomain.Schema {
	return &schemaDomain.Schema{
		Type:        "object",
		Description: "A recipe schema",
		Properties: map[string]schemaDomain.Property{
			"title":       {Type: "string", Description: "Recipe title"},
			"description": {Type: "string", Description: "Brief description"},
			"ingredients": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"name":     {Type: "string", Description: "Ingredient name"},
						"quantity": {Type: "string", Description: "Amount needed"},
						"unit":     {Type: "string", Description: "Unit of measurement"},
					},
				},
				Description: "List of ingredients",
			},
			"steps": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "string",
				},
				Description: "Preparation steps",
			},
			"prepTime": {Type: "integer", Description: "Preparation time in minutes"},
			"cookTime": {Type: "integer", Description: "Cooking time in minutes"},
			"servings": {Type: "integer", Description: "Number of servings"},
			"difficulty": {
				Type:        "string",
				Enum:        []string{"easy", "medium", "hard"},
				Description: "Recipe difficulty",
			},
		},
		Required: []string{"title", "ingredients", "steps"},
	}
}
