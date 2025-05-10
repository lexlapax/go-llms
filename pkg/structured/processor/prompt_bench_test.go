package processor

import (
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/testutils"
)

// BenchmarkGetSchemaJSON benchmarks the schema JSON generation with caching
func BenchmarkGetSchemaJSON(b *testing.B) {
	// Create a test schema
	schema := createBenchmarkSchema()

	// Clear the cache before starting
	cache := getSchemaCache()
	cache.Clear()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := getSchemaJSON(schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetSchemaJSONCached benchmarks schema JSON retrieval from cache
func BenchmarkGetSchemaJSONCached(b *testing.B) {
	// Create a test schema
	schema := createBenchmarkSchema()

	// Prime the cache before starting
	cache := getSchemaCache()
	cache.Clear()
	_, err := getSchemaJSON(schema)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := getSchemaJSON(schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPromptEnhancer benchmarks the prompt enhancement process
func BenchmarkPromptEnhancer(b *testing.B) {
	schema := createBenchmarkSchema()
	prompt := "Write a recipe with ingredients and steps."

	enhancer := NewPromptEnhancer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := enhancer.Enhance(prompt, schema)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPromptEnhancerWithOptions benchmarks the prompt enhancement with options
func BenchmarkPromptEnhancerWithOptions(b *testing.B) {
	schema := createBenchmarkSchema()
	prompt := "Write a recipe with ingredients and steps."
	
	options := map[string]interface{}{
		"instructions": "Make sure the recipe is vegetarian.",
		"format": "JSON",
		"examples": []map[string]interface{}{
			{
				"name": "Pasta Primavera",
				"ingredients": []string{"pasta", "vegetables", "olive oil"},
				"steps": []string{"Cook pasta", "Sauté vegetables", "Mix together"},
			},
		},
	}

	enhancer := NewPromptEnhancer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := enhancer.EnhanceWithOptions(prompt, schema, options)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function to create a schema for benchmarking
func createBenchmarkSchema() *schemaDomain.Schema {
	return &schemaDomain.Schema{
		Type:        "object",
		Description: "A recipe object",
		Title:       "Recipe Schema",
		Required:    []string{"name", "ingredients", "steps"},
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type:        "string",
				Description: "The name of the recipe",
				MinLength:   testutils.IntPtr(3),
				MaxLength:   testutils.IntPtr(100),
			},
			"description": {
				Type:        "string",
				Description: "A brief description of the recipe",
			},
			"ingredients": {
				Type:        "array",
				Description: "List of ingredients required",
				Items: &schemaDomain.Property{
					Type:        "string",
					Description: "An ingredient with amount",
				},
				MinItems:    testutils.IntPtr(1),
				MaxItems:    testutils.IntPtr(30),
				UniqueItems: testutils.BoolPtr(true),
			},
			"steps": {
				Type:        "array",
				Description: "List of steps to prepare the recipe",
				Items: &schemaDomain.Property{
					Type:        "string",
					Description: "A step in the cooking process",
				},
				MinItems: testutils.IntPtr(1),
			},
			"prepTime": {
				Type:        "integer",
				Description: "Preparation time in minutes",
				Minimum:     testutils.Float64Ptr(0),
				Maximum:     testutils.Float64Ptr(180),
			},
			"cookTime": {
				Type:        "integer",
				Description: "Cooking time in minutes",
				Minimum:     testutils.Float64Ptr(0),
				Maximum:     testutils.Float64Ptr(300),
			},
			"servings": {
				Type:        "integer",
				Description: "Number of servings",
				Minimum:     testutils.Float64Ptr(1),
				Maximum:     testutils.Float64Ptr(20),
			},
			"difficulty": {
				Type:        "string",
				Description: "The difficulty level of the recipe",
				Enum:        []string{"easy", "medium", "hard"},
			},
			"tags": {
				Type:        "array",
				Description: "Categories or tags for the recipe",
				Items: &schemaDomain.Property{
					Type: "string",
				},
				UniqueItems: testutils.BoolPtr(true),
			},
		},
	}
}