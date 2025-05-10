package benchmarks

import (
	"bytes"
	"encoding/json"
	"testing"

	optimizedJson "github.com/lexlapax/go-llms/pkg/util/json"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/testutils"
)

// createTestSchema creates a schema for benchmark testing
func createTestSchema() *schemaDomain.Schema {
	// Create a simple schema to test
	schema := &schemaDomain.Schema{
		Type:        "object",
		Description: "A test schema for benchmark testing",
		Title:       "Test Schema",
		Required:    []string{"name", "age", "email"},
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type:        "string",
				Description: "The name of the user",
				MinLength:   testutils.IntPtr(2),
				MaxLength:   testutils.IntPtr(50),
			},
			"age": {
				Type:        "integer",
				Description: "The age of the user",
				Minimum:     testutils.Float64Ptr(0),
				Maximum:     testutils.Float64Ptr(120),
			},
			"email": {
				Type:        "string",
				Description: "The email address of the user",
				Format:      "email",
				Pattern:     "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			},
			"address": {
				Type:        "object",
				Description: "The address of the user",
				Properties: map[string]schemaDomain.Property{
					"street": {
						Type:        "string",
						Description: "The street address",
					},
					"city": {
						Type:        "string",
						Description: "The city",
					},
					"state": {
						Type:        "string",
						Description: "The state or province",
					},
					"zip": {
						Type:        "string",
						Description: "The ZIP or postal code",
					},
				},
				Required: []string{"street", "city"},
			},
			"tags": {
				Type:        "array",
				Description: "Tags associated with the user",
				Items: &schemaDomain.Property{
					Type: "string",
				},
				MinItems:    testutils.IntPtr(0),
				MaxItems:    testutils.IntPtr(10),
				UniqueItems: testutils.BoolPtr(true),
			},
		},
	}

	return schema
}

// createComplexTestSchema creates a more complex schema for benchmark testing
func createComplexTestSchema() *schemaDomain.Schema {
	// Base schema from simple test
	schema := createTestSchema()

	// Add conditional validation
	schema.If = &schemaDomain.Schema{
		Properties: map[string]schemaDomain.Property{
			"type": {
				Type: "string",
				Enum: []string{"admin"},
			},
		},
	}

	schema.Then = &schemaDomain.Schema{
		Required: []string{"permissions"},
		Properties: map[string]schemaDomain.Property{
			"permissions": {
				Type:        "array",
				Description: "Admin permissions",
				Items: &schemaDomain.Property{
					Type: "string",
					Enum: []string{"read", "write", "delete", "manage"},
				},
				MinItems: testutils.IntPtr(1),
			},
		},
	}

	schema.Else = &schemaDomain.Schema{
		Properties: map[string]schemaDomain.Property{
			"role": {
				Type:        "string",
				Description: "User role",
				Enum:        []string{"user", "guest", "viewer"},
			},
		},
	}

	// Add allOf, anyOf, oneOf conditions
	schema.AllOf = []*schemaDomain.Schema{
		{
			Properties: map[string]schemaDomain.Property{
				"metadata": {
					Type:        "object",
					Description: "User metadata",
					Properties: map[string]schemaDomain.Property{
						"created": {
							Type:   "string",
							Format: "date-time",
						},
					},
				},
			},
		},
		{
			Properties: map[string]schemaDomain.Property{
				"active": {
					Type: "boolean",
				},
			},
		},
	}

	schema.AnyOf = []*schemaDomain.Schema{
		{
			Properties: map[string]schemaDomain.Property{
				"phone": {
					Type:        "string",
					Description: "Phone number",
					Pattern:     "^\\+?[0-9]{10,15}$",
				},
			},
		},
		{
			Properties: map[string]schemaDomain.Property{
				"mobile": {
					Type:        "string",
					Description: "Mobile number",
					Pattern:     "^\\+?[0-9]{10,15}$",
				},
			},
		},
	}

	schema.OneOf = []*schemaDomain.Schema{
		{
			Properties: map[string]schemaDomain.Property{
				"internal": {
					Type: "boolean",
					Enum: []string{"true"},
				},
			},
		},
		{
			Properties: map[string]schemaDomain.Property{
				"external": {
					Type: "boolean",
					Enum: []string{"true"},
				},
			},
		},
	}

	// Add property with conditional validation
	schema.Properties["preferences"] = schemaDomain.Property{
		Type:        "object",
		Description: "User preferences",
		Properties: map[string]schemaDomain.Property{
			"theme": {
				Type:        "string",
				Description: "UI theme",
				Enum:        []string{"light", "dark", "system"},
			},
			"notifications": {
				Type:        "boolean",
				Description: "Enable notifications",
			},
		},
		OneOf: []*schemaDomain.Schema{
			{
				Properties: map[string]schemaDomain.Property{
					"notifications": {
						Type: "boolean",
						Enum: []string{"true"},
					},
					"notificationSettings": {
						Type:        "object",
						Description: "Notification settings",
						Required:    []string{"email"},
						Properties: map[string]schemaDomain.Property{
							"email": {
								Type: "boolean",
							},
							"push": {
								Type: "boolean",
							},
							"sms": {
								Type: "boolean",
							},
						},
					},
				},
				Required: []string{"notificationSettings"},
			},
			{
				Properties: map[string]schemaDomain.Property{
					"notifications": {
						Type: "boolean",
						Enum: []string{"false"},
					},
				},
			},
		},
	}

	return schema
}

// Use helper functions directly from the testutils package or reuse package-level variables

// BenchmarkSchemaJSONMarshal benchmarks the performance of schema JSON marshaling
func BenchmarkSchemaJSONMarshal(b *testing.B) {
	simpleSchema := createTestSchema()
	complexSchema := createComplexTestSchema()

	// Benchmark standard JSON marshal for simple schema
	b.Run("StandardJSON_SimpleSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(simpleSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark optimized JSON marshal for simple schema
	b.Run("OptimizedJSON_SimpleSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.Marshal(simpleSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark standard JSON marshal with indent for simple schema
	b.Run("StandardJSON_MarshalIndent_SimpleSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.MarshalIndent(simpleSchema, "", "  ")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark optimized JSON marshal with indent for simple schema
	b.Run("OptimizedJSON_MarshalIndent_SimpleSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.MarshalIndent(simpleSchema, "", "  ")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark standard JSON marshal for complex schema
	b.Run("StandardJSON_ComplexSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(complexSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark optimized JSON marshal for complex schema
	b.Run("OptimizedJSON_ComplexSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.Marshal(complexSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark standard JSON marshal with indent for complex schema
	b.Run("StandardJSON_MarshalIndent_ComplexSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := json.MarshalIndent(complexSchema, "", "  ")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark optimized JSON marshal with indent for complex schema
	b.Run("OptimizedJSON_MarshalIndent_ComplexSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.MarshalIndent(complexSchema, "", "  ")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark optimized JSON with buffer reuse for simple schema
	b.Run("OptimizedJSON_WithBuffer_SimpleSchema", func(b *testing.B) {
		buf := &bytes.Buffer{}
		buf.Grow(2048) // Pre-allocate buffer
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := optimizedJson.MarshalWithBuffer(simpleSchema, buf)
			if err != nil {
				b.Fatal(err)
			}
			// Don't do anything with the buffer - it will be reused
		}
	})

	// Benchmark optimized JSON with buffer reuse for complex schema
	b.Run("OptimizedJSON_WithBuffer_ComplexSchema", func(b *testing.B) {
		buf := &bytes.Buffer{}
		buf.Grow(4096) // Pre-allocate larger buffer for complex schema
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := optimizedJson.MarshalWithBuffer(complexSchema, buf)
			if err != nil {
				b.Fatal(err)
			}
			// Don't do anything with the buffer - it will be reused
		}
	})

	// Benchmark optimized JSON with buffer reuse and indentation for simple schema
	b.Run("OptimizedJSON_WithBuffer_Indent_SimpleSchema", func(b *testing.B) {
		buf := &bytes.Buffer{}
		buf.Grow(2048) // Pre-allocate buffer
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := optimizedJson.MarshalIndentWithBuffer(simpleSchema, buf, "", "  ")
			if err != nil {
				b.Fatal(err)
			}
			// Don't do anything with the buffer - it will be reused
		}
	})

	// Benchmark optimized JSON with buffer reuse and indentation for complex schema
	b.Run("OptimizedJSON_WithBuffer_Indent_ComplexSchema", func(b *testing.B) {
		buf := &bytes.Buffer{}
		buf.Grow(4096) // Pre-allocate larger buffer for complex schema
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := optimizedJson.MarshalIndentWithBuffer(complexSchema, buf, "", "  ")
			if err != nil {
				b.Fatal(err)
			}
			// Don't do anything with the buffer - it will be reused
		}
	})

	// Benchmark converting to string
	b.Run("OptimizedJSON_ToString_SimpleSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.MarshalToString(simpleSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("OptimizedJSON_ToString_ComplexSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.MarshalToString(complexSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark schema-specific optimized marshalers
	b.Run("SchemaSpecific_Fast_SimpleSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.MarshalSchemaFast(simpleSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SchemaSpecific_Indent_SimpleSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.MarshalSchemaIndent(simpleSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SchemaSpecific_Fast_ComplexSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.MarshalSchemaFast(complexSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SchemaSpecific_Indent_ComplexSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.MarshalSchemaIndent(complexSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SchemaSpecific_ToString_ComplexSchema", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := optimizedJson.MarshalSchemaToString(complexSchema)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}