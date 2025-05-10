package benchmarks

import (
	"strings"
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
	"github.com/lexlapax/go-llms/pkg/testutils"
)

// BenchmarkStringBuilderCapacity tests different capacity estimation strategies for string builders
func BenchmarkStringBuilderCapacity(b *testing.B) {
	// Setup test schemas of different complexities
	simpleSchema := createSimpleSchema()
	mediumSchema := createMediumSchema()
	complexSchema := createComplexSchema()

	// Test different prompt sizes
	smallPrompt := "Generate a recipe"
	mediumPrompt := "Generate a detailed recipe with ingredients, instructions, and nutritional information."
	largePrompt := "Generate a comprehensive recipe including ingredients, step-by-step instructions, nutritional information, variations, serving suggestions, wine pairings, and historical background of the dish. Include notes about possible allergens and substitutions for different dietary restrictions."

	// Benchmark with default string builder (no pre-allocation)
	b.Run("DefaultBuilder", func(b *testing.B) {
		b.Run("Simple_SmallPrompt", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buildPromptDefault(smallPrompt, simpleSchema)
			}
		})
		b.Run("Medium_MediumPrompt", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buildPromptDefault(mediumPrompt, mediumSchema)
			}
		})
		b.Run("Complex_LargePrompt", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buildPromptDefault(largePrompt, complexSchema)
			}
		})
	})

	// Benchmark with current pre-allocation strategy
	b.Run("CurrentPreallocation", func(b *testing.B) {
		b.Run("Simple_SmallPrompt", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buildPromptCurrentStrategy(smallPrompt, simpleSchema)
			}
		})
		b.Run("Medium_MediumPrompt", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buildPromptCurrentStrategy(mediumPrompt, mediumSchema)
			}
		})
		b.Run("Complex_LargePrompt", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buildPromptCurrentStrategy(largePrompt, complexSchema)
			}
		})
	})

	// Benchmark with real implementation via the Enhancer
	b.Run("RealImplementation", func(b *testing.B) {
		b.Run("Simple_SmallPrompt", func(b *testing.B) {
			enhancer := processor.NewPromptEnhancer()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = enhancer.Enhance(smallPrompt, simpleSchema)
			}
		})
		b.Run("Medium_MediumPrompt", func(b *testing.B) {
			enhancer := processor.NewPromptEnhancer()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = enhancer.Enhance(mediumPrompt, mediumSchema)
			}
		})
		b.Run("Complex_LargePrompt", func(b *testing.B) {
			enhancer := processor.NewPromptEnhancer()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = enhancer.Enhance(largePrompt, complexSchema)
			}
		})
	})
}

// Sample implementation with no pre-allocation (baseline)
func buildPromptDefault(prompt string, schema *schemaDomain.Schema) string {
	var builder strings.Builder
	
	// Add base prompt
	builder.WriteString(prompt)
	builder.WriteString("\n\n")
	builder.WriteString("Please provide a JSON response according to this schema:\n\n")
	
	// Add schema details
	builder.WriteString("Type: ")
	builder.WriteString(schema.Type)
	builder.WriteString("\n")
	
	// Add properties
	if schema.Type == "object" && len(schema.Properties) > 0 {
		builder.WriteString("Properties:\n")
		for name, prop := range schema.Properties {
			builder.WriteString("- ")
			builder.WriteString(name)
			builder.WriteString(": ")
			builder.WriteString(prop.Type)
			if prop.Description != "" {
				builder.WriteString(" (")
				builder.WriteString(prop.Description)
				builder.WriteString(")")
			}
			builder.WriteString("\n")
		}
	}
	
	// Add required fields
	if len(schema.Required) > 0 {
		builder.WriteString("Required fields: ")
		builder.WriteString(strings.Join(schema.Required, ", "))
		builder.WriteString("\n")
	}
	
	return builder.String()
}

// Sample implementation with current capacity estimation
func buildPromptCurrentStrategy(prompt string, schema *schemaDomain.Schema) string {
	// Calculate initial capacity based on input sizes
	initialCapacity := len(prompt) + 500  // Base prompt + standard text
	
	// Account for property descriptions (est. ~50 bytes per property)
	if schema.Type == "object" {
		initialCapacity += len(schema.Properties) * 50
	}
	
	var builder strings.Builder
	builder.Grow(initialCapacity)
	
	// Add base prompt
	builder.WriteString(prompt)
	builder.WriteString("\n\n")
	builder.WriteString("Please provide a JSON response according to this schema:\n\n")
	
	// Add schema details
	builder.WriteString("Type: ")
	builder.WriteString(schema.Type)
	builder.WriteString("\n")
	
	// Add properties
	if schema.Type == "object" && len(schema.Properties) > 0 {
		builder.WriteString("Properties:\n")
		for name, prop := range schema.Properties {
			builder.WriteString("- ")
			builder.WriteString(name)
			builder.WriteString(": ")
			builder.WriteString(prop.Type)
			if prop.Description != "" {
				builder.WriteString(" (")
				builder.WriteString(prop.Description)
				builder.WriteString(")")
			}
			builder.WriteString("\n")
		}
	}
	
	// Add required fields
	if len(schema.Required) > 0 {
		builder.WriteString("Required fields: ")
		builder.WriteString(strings.Join(schema.Required, ", "))
		builder.WriteString("\n")
	}
	
	return builder.String()
}

// Helper to create a simple schema
func createSimpleSchema() *schemaDomain.Schema {
	return &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string", Description: "The name"},
			"age":  {Type: "integer", Description: "The age"},
		},
		Required: []string{"name"},
	}
}

// Helper to create a medium-complexity schema
func createMediumSchema() *schemaDomain.Schema {
	return &schemaDomain.Schema{
		Type: "object",
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
					Required: []string{"name"},
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
		},
		Required: []string{"title", "ingredients", "steps"},
	}
}

// Helper to create a complex schema
func createComplexSchema() *schemaDomain.Schema {
	schema := &schemaDomain.Schema{
		Type:        "object",
		Description: "A complete recipe schema",
		Title:       "Recipe",
		Required:    []string{"title", "ingredients", "instructions", "nutrition"},
		Properties: map[string]schemaDomain.Property{
			"title": {
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
				Description: "List of ingredients with amounts",
				Items: &schemaDomain.Property{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"name": {
							Type:        "string",
							Description: "Name of the ingredient",
						},
						"amount": {
							Type:        "string",
							Description: "Amount with unit (e.g., '2 cups')",
						},
						"preparation": {
							Type:        "string",
							Description: "Optional preparation instruction (e.g., 'finely chopped')",
						},
						"substitutes": {
							Type: "array",
							Items: &schemaDomain.Property{
								Type: "string",
							},
							Description: "Possible substitutions",
						},
					},
					Required: []string{"name", "amount"},
				},
				MinItems: testutils.IntPtr(1),
			},
			"instructions": {
				Type:        "array",
				Description: "Step-by-step cooking instructions",
				Items: &schemaDomain.Property{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"step": {
							Type:        "integer",
							Description: "Step number",
							Minimum:     testutils.Float64Ptr(1),
						},
						"description": {
							Type:        "string",
							Description: "Description of the step",
						},
						"timer": {
							Type:        "integer",
							Description: "Optional timer in minutes for this step",
							Minimum:     testutils.Float64Ptr(0),
						},
						"note": {
							Type:        "string",
							Description: "Optional chef's note for this step",
						},
					},
					Required: []string{"step", "description"},
				},
				MinItems: testutils.IntPtr(1),
			},
			"nutrition": {
				Type:        "object",
				Description: "Nutritional information per serving",
				Properties: map[string]schemaDomain.Property{
					"calories": {
						Type:        "integer",
						Description: "Calories per serving",
						Minimum:     testutils.Float64Ptr(0),
					},
					"protein": {
						Type:        "integer",
						Description: "Protein in grams",
						Minimum:     testutils.Float64Ptr(0),
					},
					"carbohydrates": {
						Type:        "integer",
						Description: "Carbohydrates in grams",
						Minimum:     testutils.Float64Ptr(0),
					},
					"fat": {
						Type:        "integer",
						Description: "Fat in grams",
						Minimum:     testutils.Float64Ptr(0),
					},
					"fiber": {
						Type:        "integer",
						Description: "Fiber in grams",
						Minimum:     testutils.Float64Ptr(0),
					},
					"sugar": {
						Type:        "integer",
						Description: "Sugar in grams",
						Minimum:     testutils.Float64Ptr(0),
					},
				},
				Required: []string{"calories"},
			},
			"servings": {
				Type:        "integer",
				Description: "Number of servings this recipe yields",
				Minimum:     testutils.Float64Ptr(1),
				Maximum:     testutils.Float64Ptr(50),
			},
			"cookingTime": {
				Type:        "object",
				Description: "Time needed to prepare and cook",
				Properties: map[string]schemaDomain.Property{
					"preparation": {
						Type:        "integer",
						Description: "Preparation time in minutes",
						Minimum:     testutils.Float64Ptr(0),
					},
					"cooking": {
						Type:        "integer",
						Description: "Cooking time in minutes",
						Minimum:     testutils.Float64Ptr(0),
					},
					"total": {
						Type:        "integer",
						Description: "Total time in minutes",
						Minimum:     testutils.Float64Ptr(0),
					},
				},
				Required: []string{"total"},
			},
			"difficulty": {
				Type:        "string",
				Enum:        []string{"easy", "medium", "hard", "expert"},
				Description: "The difficulty level of the recipe",
			},
			"cuisine": {
				Type:        "string",
				Description: "The type of cuisine (e.g., Italian, Mexican, etc.)",
			},
			"tags": {
				Type:        "array",
				Description: "Tags to categorize the recipe",
				Items: &schemaDomain.Property{
					Type: "string",
				},
				UniqueItems: testutils.BoolPtr(true),
			},
			"notes": {
				Type:        "array",
				Description: "Additional chef's notes or tips",
				Items: &schemaDomain.Property{
					Type: "string",
				},
			},
			"equipment": {
				Type:        "array",
				Description: "Required kitchen equipment",
				Items: &schemaDomain.Property{
					Type: "string",
				},
			},
		},
	}

	// Add conditional validation
	schema.If = &schemaDomain.Schema{
		Properties: map[string]schemaDomain.Property{
			"difficulty": {
				Type: "string",
				Enum: []string{"expert"},
			},
		},
	}
	schema.Then = &schemaDomain.Schema{
		Properties: map[string]schemaDomain.Property{
			"warnings": {
				Type:        "array",
				Description: "Warnings about difficult techniques",
				Items: &schemaDomain.Property{
					Type: "string",
				},
				MinItems: testutils.IntPtr(1),
			},
		},
		Required: []string{"warnings"},
	}

	return schema
}