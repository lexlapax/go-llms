package benchmarks

import (
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	structProcessor "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// BenchmarkJSONExtraction benchmarks extraction of JSON from text
func BenchmarkJSONExtraction(b *testing.B) {
	validator := validation.NewValidator()
	processor := structProcessor.NewStructuredProcessor(validator)

	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name":  {Type: "string"},
			"age":   {Type: "integer"},
			"email": {Type: "string"},
		},
		Required: []string{"name", "email"},
	}

	// Simple text with JSON
	simpleText := `Here's the information you requested:
{
  "name": "John Doe",
  "age": 30,
  "email": "john@example.com"
}`

	// More complex text with JSON in markdown code block
	markdownText := `# User Information

Here's the user information you requested:

` + "```json" + `
{
  "name": "Jane Smith",
  "age": 28,
  "email": "jane@example.com"
}
` + "```" + `

I hope this helps!`

	b.Run("SimpleTextJSON", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := processor.Process(schema, simpleText)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MarkdownCodeBlock", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := processor.Process(schema, markdownText)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkPromptEnhancement benchmarks the prompt enhancement process
func BenchmarkPromptEnhancement(b *testing.B) {
	enhancer := structProcessor.NewPromptEnhancer()

	// Simple schema
	simpleSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name":  {Type: "string", Description: "Person's name"},
			"age":   {Type: "integer", Description: "Person's age"},
			"email": {Type: "string", Description: "Email address"},
		},
		Required: []string{"name", "email"},
	}

	// Complex schema
	complexSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"recipe": {
				Type: "object",
				Properties: map[string]schemaDomain.Property{
					"title":       {Type: "string", Description: "Recipe title"},
					"description": {Type: "string", Description: "Recipe description"},
					"prep_time":   {Type: "integer", Description: "Preparation time in minutes"},
					"cook_time":   {Type: "integer", Description: "Cooking time in minutes"},
					"servings":    {Type: "integer", Description: "Number of servings"},
					"difficulty":  {Type: "string", Enum: []string{"easy", "medium", "hard"}},
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
					},
					"steps": {
						Type: "array",
						Items: &schemaDomain.Property{
							Type: "string",
						},
						Description: "Steps to prepare the recipe",
					},
				},
				Required: []string{"title", "ingredients", "steps"},
			},
		},
		Required: []string{"recipe"},
	}

	// Test with simple prompt and simple schema
	simplePrompt := "Generate information about a person"
	b.Run("SimplePromptSimpleSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.Enhance(simplePrompt, simpleSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Test with simple prompt and complex schema
	b.Run("SimplePromptComplexSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.Enhance(simplePrompt, complexSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
