package benchmarks

import (
	"encoding/json"
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// BenchmarkPromptTemplateProcessing benchmarks various aspects of prompt template processing
func BenchmarkPromptTemplateProcessing(b *testing.B) {
	// Test data setup
	// Define schemas of different types and complexities
	simpleObjectSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
		Required: []string{"name"},
	}

	// Helper functions to create pointers to primitive types
	intPtr := func(i int) *int { return &i }
	float64Ptr := func(f float64) *float64 { return &f }
	boolPtr := func(b bool) *bool { return &b }

	complexObjectSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type:        "string",
				Description: "The name of the person",
				MinLength:   intPtr(2),
				MaxLength:   intPtr(50),
			},
			"age": {
				Type:        "integer",
				Description: "Age in years",
				Minimum:     float64Ptr(0),
				Maximum:     float64Ptr(120),
			},
			"email": {
				Type:        "string",
				Description: "Email address",
				Format:      "email",
			},
			"tags": {
				Type:        "array",
				Description: "List of tags",
				Items:       &schemaDomain.Property{Type: "string"},
			},
			"interests": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "string",
					Enum: []string{"sports", "music", "reading", "travel", "cooking"},
				},
				UniqueItems: boolPtr(true),
			},
			"address": {
				Type:        "object",
				Description: "User's address",
				Properties: map[string]schemaDomain.Property{
					"street":  {Type: "string"},
					"city":    {Type: "string"},
					"zipCode": {Type: "string", Pattern: "^\\d{5}(-\\d{4})?$"},
					"country": {Type: "string"},
				},
				Required: []string{"street", "city"},
			},
		},
		Required: []string{"name", "email"},
	}

	simpleArraySchema := &schemaDomain.Schema{
		Type: "array",
		Properties: map[string]schemaDomain.Property{
			"": {
				Type:  "array",
				Items: &schemaDomain.Property{Type: "string"},
			},
		},
	}

	complexArraySchema := &schemaDomain.Schema{
		Type: "array",
		Properties: map[string]schemaDomain.Property{
			"": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"id":    {Type: "integer"},
						"name":  {Type: "string"},
						"value": {Type: "number"},
						"tags":  {Type: "array", Items: &schemaDomain.Property{Type: "string"}},
					},
					Required: []string{"id", "name"},
				},
			},
		},
	}

	// Define prompts of various lengths
	shortPrompt := "Generate a user profile."
	mediumPrompt := "Please generate a detailed user profile with name, age, email address, interests, and location information. Make sure to include all required fields and follow the formatting guidelines specified in the schema."
	longPrompt := `Generate a comprehensive user profile for our database system. The profile should include:
1. Full name (first and last name)
2. Age (must be between 0 and 120)
3. Valid email address for contact purposes
4. A list of user interests/hobbies from the predefined categories
5. Complete address information including street, city, ZIP code, and country

Please ensure that the generated data is realistic and follows all constraints specified in the schema. 
Names should be properly capitalized and between 2-50 characters. 
Email addresses must be in valid format. 
ZIP codes should follow the US format (5 digits or 5+4 format).
The user's interests should be selected from the available options and not contain duplicates.
All required fields must be present and properly formatted according to the schema.`

	// 1. Benchmark singleton pattern with GetDefaultPromptEnhancer()
	b.Run("SingletonEnhancer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			enhancer := processor.GetDefaultPromptEnhancer()
			_ = enhancer
		}
	})

	// 2. Benchmark creating a new enhancer
	b.Run("NewEnhancer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			enhancer := processor.NewPromptEnhancer()
			_ = enhancer
		}
	})

	// 3. Benchmark enhancing with singleton vs. new instance
	b.Run("EnhanceWithSingleton", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			enhancer := processor.GetDefaultPromptEnhancer()
			_, err := enhancer.Enhance(shortPrompt, simpleObjectSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("EnhanceWithNewInstance", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			enhancer := processor.NewPromptEnhancer()
			_, err := enhancer.Enhance(shortPrompt, simpleObjectSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 4. Benchmark convenience function
	b.Run("ConvenienceFunction", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := processor.EnhancePromptWithSchema(shortPrompt, simpleObjectSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 5. Benchmark enhancing with different prompt lengths
	b.Run("ShortPrompt", func(b *testing.B) {
		enhancer := processor.GetDefaultPromptEnhancer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.Enhance(shortPrompt, simpleObjectSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MediumPrompt", func(b *testing.B) {
		enhancer := processor.GetDefaultPromptEnhancer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.Enhance(mediumPrompt, simpleObjectSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("LongPrompt", func(b *testing.B) {
		enhancer := processor.GetDefaultPromptEnhancer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.Enhance(longPrompt, simpleObjectSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 6. Benchmark enhancing with different schema complexities
	b.Run("SimpleObjectSchema", func(b *testing.B) {
		enhancer := processor.GetDefaultPromptEnhancer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.Enhance(mediumPrompt, simpleObjectSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ComplexObjectSchema", func(b *testing.B) {
		enhancer := processor.GetDefaultPromptEnhancer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.Enhance(mediumPrompt, complexObjectSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SimpleArraySchema", func(b *testing.B) {
		enhancer := processor.GetDefaultPromptEnhancer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.Enhance(mediumPrompt, simpleArraySchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ComplexArraySchema", func(b *testing.B) {
		enhancer := processor.GetDefaultPromptEnhancer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.Enhance(mediumPrompt, complexArraySchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 7. Benchmark schema caching
	b.Run("SchemaCache_FirstAccess", func(b *testing.B) {
		// Generate a series of slightly different schemas to avoid cache hits
		schemas := make([]*schemaDomain.Schema, b.N)
		for i := 0; i < b.N; i++ {
			// Clone the complex schema but modify a field slightly
			schema := *complexObjectSchema
			schema.Required = append(schema.Required, "age") // Add an extra required field
			schemas[i] = &schema
		}
		
		enhancer := processor.GetDefaultPromptEnhancer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.Enhance(shortPrompt, schemas[i])
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SchemaCache_RepeatedAccess", func(b *testing.B) {
		enhancer := processor.GetDefaultPromptEnhancer()
		// Warm up cache
		_, _ = enhancer.Enhance(shortPrompt, complexObjectSchema)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.Enhance(shortPrompt, complexObjectSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 8. Benchmark options processing
	b.Run("EnhanceWithSimpleOptions", func(b *testing.B) {
		enhancer := processor.GetDefaultPromptEnhancer()
		options := map[string]interface{}{
			"instructions": "Focus on accuracy.",
			"format":       "a detailed profile",
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.EnhanceWithOptions(mediumPrompt, simpleObjectSchema, options)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("EnhanceWithComplexOptions", func(b *testing.B) {
		enhancer := processor.GetDefaultPromptEnhancer()
		// Create complex options with examples
		examples := make([]map[string]interface{}, 3)
		
		examples[0] = map[string]interface{}{
			"name":  "John Doe",
			"age":   30,
			"email": "john.doe@example.com",
			"tags":  []string{"vip", "premium"},
			"interests": []string{"sports", "music"},
			"address": map[string]interface{}{
				"street":  "123 Main St",
				"city":    "Anytown",
				"zipCode": "12345",
				"country": "USA",
			},
		}
		
		examples[1] = map[string]interface{}{
			"name":  "Jane Smith",
			"email": "jane.smith@example.com",
			"age":   25,
			"interests": []string{"reading", "travel"},
			"address": map[string]interface{}{
				"street":  "456 Oak Ave",
				"city":    "Somewhere",
				"country": "Canada",
			},
		}
		
		examples[2] = map[string]interface{}{
			"name":  "Bob Johnson",
			"email": "bob@example.com",
			"age":   40,
			"tags":  []string{"new", "trial"},
			"address": map[string]interface{}{
				"street": "789 Pine St",
				"city":   "Elsewhere",
			},
		}
		
		options := map[string]interface{}{
			"instructions": "Generate realistic data for a diverse set of users with various interests and demographics.",
			"format":       "a detailed user profile with all available fields properly filled out",
			"examples":     examples,
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.EnhanceWithOptions(mediumPrompt, complexObjectSchema, options)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 9. Benchmark schema JSON marshaling
	b.Run("SchemaMarshaling_Simple", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(simpleObjectSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SchemaMarshaling_Complex", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(complexObjectSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}