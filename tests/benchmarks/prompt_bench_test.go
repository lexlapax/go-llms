package benchmarks

import (
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/structured/processor"
)

// BenchmarkPromptProcessing benchmarks the prompt enhancement process
func BenchmarkPromptProcessing(b *testing.B) {
	// Create test schemas of different complexities
	simpleSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
		Required: []string{"name"},
	}

	mediumSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name":        {Type: "string", Description: "Full name of the person"},
			"age":         {Type: "integer", Description: "Age in years"},
			"email":       {Type: "string", Description: "Email address", Format: "email"},
			"phoneNumber": {Type: "string", Description: "Contact phone number"},
			"address": {
				Type: "object",
				Properties: map[string]schemaDomain.Property{
					"street":  {Type: "string"},
					"city":    {Type: "string"},
					"state":   {Type: "string"},
					"zipCode": {Type: "string"},
				},
			},
		},
		Required: []string{"name", "email"},
	}

	complexSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"id":       {Type: "string", Description: "Unique identifier"},
			"type":     {Type: "string", Description: "Type of entity"},
			"name":     {Type: "string", Description: "Full name of the person or organization"},
			"age":      {Type: "integer", Description: "Age in years (for individuals)"},
			"email":    {Type: "string", Description: "Primary email address", Format: "email"},
			"verified": {Type: "boolean", Description: "Whether the account is verified"},
			"tags": {
				Type:        "array",
				Description: "List of tags",
				Items:       &schemaDomain.Property{Type: "string"},
			},
			"contact": {
				Type: "object",
				Properties: map[string]schemaDomain.Property{
					"phoneNumbers": {
						Type: "array",
						Items: &schemaDomain.Property{
							Type: "object",
							Properties: map[string]schemaDomain.Property{
								"type":   {Type: "string"},
								"number": {Type: "string"},
							},
						},
					},
					"preferredMethod": {Type: "string"},
				},
			},
			"addresses": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"type":    {Type: "string"},
						"street":  {Type: "string"},
						"city":    {Type: "string"},
						"state":   {Type: "string"},
						"zipCode": {Type: "string"},
						"country": {Type: "string"},
						"default": {Type: "boolean"},
					},
				},
			},
		},
		Required: []string{"id", "type", "name", "email"},
	}

	// Test prompts
	shortPrompt := "Provide information about a person."
	mediumPrompt := "Provide detailed information about a person including their contact information and address. Make sure to include all required fields and use appropriate formats for email addresses and phone numbers."
	longPrompt := "Generate a comprehensive profile for a customer including personal information, contact details, multiple addresses, account status, and preferred contact methods. The profile should include all personal identifiers, contact preferences, multiple addresses tagged by purpose, and any relevant account flags. This information will be used for a CRM system so ensure it follows the schema guidelines precisely."

	// Benchmark simple schema with short prompt
	b.Run("SimpleSchema_ShortPrompt", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := processor.EnhancePromptWithSchema(shortPrompt, simpleSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark medium schema with medium prompt
	b.Run("MediumSchema_MediumPrompt", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := processor.EnhancePromptWithSchema(mediumPrompt, mediumSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark complex schema with long prompt
	b.Run("ComplexSchema_LongPrompt", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := processor.EnhancePromptWithSchema(longPrompt, complexSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark repeated enhancement with the same schema
	b.Run("RepeatedEnhancement_SameSchema", func(b *testing.B) {
		prompts := []string{
			"Tell me about a person named John.",
			"What is Mary's information?",
			"Create a profile for Sarah.",
			"Generate data for a new customer.",
			"I need information about Alex.",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			promptIdx := i % len(prompts)
			_, err := processor.EnhancePromptWithSchema(prompts[promptIdx], mediumSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark with EnhanceWithOptions
	b.Run("EnhanceWithOptions", func(b *testing.B) {
		options := map[string]interface{}{
			"instructions": "Focus on providing accurate contact information.",
			"format":       "a complete profile",
			"examples": []map[string]interface{}{
				{
					"name":        "John Doe",
					"age":         30,
					"email":       "john.doe@example.com",
					"phoneNumber": "555-1234",
					"address": map[string]string{
						"street":  "123 Main St",
						"city":    "Anytown",
						"state":   "CA",
						"zipCode": "12345",
					},
				},
			},
		}
		enhancer := processor.NewPromptEnhancer()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := enhancer.EnhanceWithOptions(mediumPrompt, mediumSchema, options)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
