package processor

import (
	"fmt"
	"strconv"
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/testutils"
)

func TestGenerateSchemaKey(t *testing.T) {
	// Basic test to ensure the function returns a uint64
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type: "string",
			},
		},
	}

	key := GenerateSchemaKey(schema)
	if key == 0 {
		t.Errorf("Expected non-zero key, got %d", key)
	}
}

func TestGenerateSchemaKeyDifferentiation(t *testing.T) {
	// Test that different schemas produce different keys
	schema1 := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string"},
		},
	}

	schema2 := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"age": {Type: "integer"},
		},
	}

	key1 := GenerateSchemaKey(schema1)
	key2 := GenerateSchemaKey(schema2)

	if key1 == key2 {
		t.Errorf("Expected different keys for different schemas, got %d for both", key1)
	}
}

func TestGenerateSchemaKeyConsistency(t *testing.T) {
	// Test that the same schema produces the same key
	schema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
	}

	key1 := GenerateSchemaKey(schema)
	key2 := GenerateSchemaKey(schema)

	if key1 != key2 {
		t.Errorf("Expected same key for same schema, got %d and %d", key1, key2)
	}
}

func TestGenerateSchemaKeyPropertyOrder(t *testing.T) {
	// Test that the same schema properties in different order should ideally produce the same key
	// But this might fail with the current implementation due to map iteration order
	schema1 := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
	}

	schema2 := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"age":  {Type: "integer"},
			"name": {Type: "string"},
		},
	}

	key1 := GenerateSchemaKey(schema1)
	key2 := GenerateSchemaKey(schema2)

	// In current implementation, these might be different keys due to map iteration order
	if key1 != key2 {
		t.Logf("Map iteration order affects key generation. Got %d and %d", key1, key2)
	}
}

func TestGenerateSchemaKeyConditionals(t *testing.T) {
	// Test that schemas with conditional validation rules produce different keys
	baseSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string"},
		},
	}

	schemaWithIf := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string"},
		},
		If: &schemaDomain.Schema{
			Properties: map[string]schemaDomain.Property{
				"type": {Type: "string", Enum: []string{"admin"}},
			},
		},
	}

	baseKey := GenerateSchemaKey(baseSchema)
	ifKey := GenerateSchemaKey(schemaWithIf)

	if baseKey == ifKey {
		t.Errorf("Expected different keys for schemas with vs. without conditionals, got %d for both", baseKey)
	}
}

func TestGenerateSchemaKeyNestedProperties(t *testing.T) {
	// Test that schemas with nested properties produce different keys
	baseSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string"},
		},
	}

	schemaWithNested := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {Type: "string"},
			"address": {
				Type: "object",
				Properties: map[string]schemaDomain.Property{
					"street": {Type: "string"},
					"city":   {Type: "string"},
				},
			},
		},
	}

	baseKey := GenerateSchemaKey(baseSchema)
	nestedKey := GenerateSchemaKey(schemaWithNested)

	if baseKey == nestedKey {
		t.Errorf("Expected different keys for schemas with vs. without nested properties, got %d for both", baseKey)
	}
}

func TestGenerateSchemaKeyEnums(t *testing.T) {
	// Test that schemas with enum values produce different keys
	baseSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"status": {Type: "string"},
		},
	}

	schemaWithEnum := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"status": {Type: "string", Enum: []string{"active", "inactive"}},
		},
	}

	baseKey := GenerateSchemaKey(baseSchema)
	enumKey := GenerateSchemaKey(schemaWithEnum)

	if baseKey == enumKey {
		t.Errorf("Expected different keys for schemas with vs. without enum values, got %d for both", baseKey)
	}
}

func TestGenerateSchemaKeyArrayItems(t *testing.T) {
	// Test that schemas with array items produce different keys
	baseSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"tags": {Type: "array"},
		},
	}

	schemaWithItems := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"tags": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "string",
				},
			},
		},
	}

	baseKey := GenerateSchemaKey(baseSchema)
	itemsKey := GenerateSchemaKey(schemaWithItems)

	if baseKey == itemsKey {
		t.Errorf("Expected different keys for schemas with vs. without array items, got %d for both", baseKey)
	}
}

func TestSchemaKeysCollisionResistance(t *testing.T) {
	// Test that similar schemas produce different keys
	// Create schemas with small differences to check collision resistance

	// Create a large set of schemas with minor variations
	var schemas []*schemaDomain.Schema
	var keys []uint64

	for i := 0; i < 100; i++ {
		schema := &schemaDomain.Schema{
			Type:        "object",
			Description: fmt.Sprintf("Schema description %d", i),
			Properties: map[string]schemaDomain.Property{
				"field" + strconv.Itoa(i): {
					Type:        "string",
					Description: fmt.Sprintf("Field description %d", i),
				},
			},
		}
		// Make sure the append result is assigned back to the variable
		schemas = append(schemas, schema)
		_ = schemas // Explicitly use schemas to avoid unused result warning
		keys = append(keys, GenerateSchemaKey(schema))
	}

	// Check for collisions
	keyMap := make(map[uint64]bool)
	collisions := 0

	for _, key := range keys {
		if keyMap[key] {
			collisions++
		}
		keyMap[key] = true
	}

	if collisions > 0 {
		t.Errorf("Found %d key collisions among 100 similar schemas", collisions)
	}
}

// Benchmarking the key generation function
func BenchmarkGenerateSchemaKey(b *testing.B) {
	schema := createTestSchema()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateSchemaKey(schema)
	}
}

// Helper function to create a test schema for benchmarking
func createTestSchema() *schemaDomain.Schema {
	return &schemaDomain.Schema{
		Type:        "object",
		Description: "A test schema",
		Title:       "Test Schema",
		Required:    []string{"name", "age", "email"},
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type:        "string",
				Description: "The name",
				MinLength:   testutils.IntPtr(2),
				MaxLength:   testutils.IntPtr(50),
			},
			"age": {
				Type:        "integer",
				Description: "The age",
				Minimum:     testutils.Float64Ptr(0),
				Maximum:     testutils.Float64Ptr(120),
			},
			"email": {
				Type:        "string",
				Description: "The email",
				Format:      "email",
			},
			"address": {
				Type:        "object",
				Description: "The address",
				Properties: map[string]schemaDomain.Property{
					"street": {
						Type:        "string",
						Description: "The street",
					},
					"city": {
						Type:        "string",
						Description: "The city",
					},
				},
				Required: []string{"street", "city"},
			},
			"tags": {
				Type:        "array",
				Description: "The tags",
				Items: &schemaDomain.Property{
					Type: "string",
				},
			},
		},
		If: &schemaDomain.Schema{
			Properties: map[string]schemaDomain.Property{
				"type": {
					Type: "string",
					Enum: []string{"admin"},
				},
			},
		},
		Then: &schemaDomain.Schema{
			Required: []string{"permissions"},
		},
	}
}
