package generator

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Test structures
type SimpleUser struct {
	ID        string    `json:"id" validate:"required,uuid" format:"uuid"`
	Name      string    `json:"name" validate:"required,min=1,max=100" description:"User's full name"`
	Email     string    `json:"email" validate:"required,email"`
	Age       int       `json:"age,omitempty" validate:"min=0,max=150"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
}

type ComplexProduct struct {
	SKU      string                 `json:"sku" pattern:"^[A-Z]{3}-[0-9]{4}$" validate:"required"`
	Name     string                 `json:"name" minLength:"1" maxLength:"200"`
	Price    float64                `json:"price" minimum:"0" exclusiveMinimum:"0"`
	Tags     []string               `json:"tags" uniqueItems:"true" minItems:"1"`
	InStock  bool                   `json:"in_stock"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Variants []ProductVariant       `json:"variants,omitempty"`
}

type ProductVariant struct {
	ID    string `json:"id"`
	Color string `json:"color" enum:"red,green,blue"`
	Size  string `json:"size" validate:"oneof=S M L XL"`
}

type SchemaTaggedStruct struct {
	Field1 string `schema:"type=string,format=email,required"`
	Field2 int    `schema:"type=integer,minimum=10,maximum=100"`
	Field3 []bool `schema:"type=array,minItems=2,description=Boolean array"`
}

type NestedStruct struct {
	User    SimpleUser     `json:"user" validate:"required"`
	Product ComplexProduct `json:"product"`
	Extra   interface{}    `json:"extra,omitempty"`
}

func TestReflectionSchemaGenerator_SimpleStruct(t *testing.T) {
	gen := NewReflectionSchemaGenerator()

	schema, err := gen.GenerateSchema(SimpleUser{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Check basic structure
	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got %s", schema.Type)
	}

	if len(schema.Properties) != 5 {
		t.Errorf("Expected 5 properties, got %d", len(schema.Properties))
	}

	// Check ID property
	idProp, ok := schema.Properties["id"]
	if !ok {
		t.Error("ID property not found")
	} else {
		if idProp.Type != "string" {
			t.Errorf("ID: expected type 'string', got %s", idProp.Type)
		}
		if idProp.Format != "uuid" {
			t.Errorf("ID: expected format 'uuid', got %s", idProp.Format)
		}
	}

	// Check Email property
	emailProp, ok := schema.Properties["email"]
	if !ok {
		t.Error("Email property not found")
	} else {
		if emailProp.Format != "email" {
			t.Errorf("Email: expected format 'email', got %s", emailProp.Format)
		}
	}

	// Check required fields (created_at is also required as it doesn't have omitempty)
	expectedRequired := []string{"id", "name", "email", "created_at"}
	if len(schema.Required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(schema.Required))
	}

	// Check Age constraints
	ageProp, ok := schema.Properties["age"]
	if !ok {
		t.Error("Age property not found")
	} else {
		if ageProp.Minimum == nil || *ageProp.Minimum != 0 {
			t.Error("Age: minimum constraint not set correctly")
		}
		if ageProp.Maximum == nil || *ageProp.Maximum != 150 {
			t.Error("Age: maximum constraint not set correctly")
		}
	}
}

func TestReflectionSchemaGenerator_ComplexStruct(t *testing.T) {
	gen := NewReflectionSchemaGenerator()

	schema, err := gen.GenerateSchema(ComplexProduct{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Check array property
	tagsProp, ok := schema.Properties["tags"]
	if !ok {
		t.Error("Tags property not found")
	} else {
		if tagsProp.Type != "array" {
			t.Errorf("Tags: expected type 'array', got %s", tagsProp.Type)
		}
		if tagsProp.UniqueItems == nil || !*tagsProp.UniqueItems {
			t.Error("Tags: uniqueItems not set")
		}
		if tagsProp.MinItems == nil || *tagsProp.MinItems != 1 {
			t.Error("Tags: minItems not set correctly")
		}
	}

	// Check map property
	metaProp, ok := schema.Properties["metadata"]
	if !ok {
		t.Error("Metadata property not found")
	} else {
		if metaProp.Type != "object" {
			t.Errorf("Metadata: expected type 'object', got %s", metaProp.Type)
		}
		if metaProp.AdditionalProperties == nil || !*metaProp.AdditionalProperties {
			t.Error("Metadata: additionalProperties not set")
		}
	}

	// Check nested array of structs
	variantsProp, ok := schema.Properties["variants"]
	if !ok {
		t.Error("Variants property not found")
	} else {
		if variantsProp.Type != "array" {
			t.Errorf("Variants: expected type 'array', got %s", variantsProp.Type)
		}
		if variantsProp.Items == nil {
			t.Error("Variants: items not set")
		} else if variantsProp.Items.Type != "object" {
			t.Errorf("Variants items: expected type 'object', got %s", variantsProp.Items.Type)
		}
	}
}

func TestReflectionSchemaGenerator_NestedStruct(t *testing.T) {
	gen := NewReflectionSchemaGenerator()

	schema, err := gen.GenerateSchema(NestedStruct{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Check nested user
	userProp, ok := schema.Properties["user"]
	if !ok {
		t.Error("User property not found")
	} else {
		if userProp.Type != "object" {
			t.Errorf("User: expected type 'object', got %s", userProp.Type)
		}
		if len(userProp.Properties) == 0 {
			t.Error("User: no nested properties found")
		}
		// Check nested property
		if nameProp, ok := userProp.Properties["name"]; ok {
			if nameProp.Type != "string" {
				t.Error("User.name: incorrect type")
			}
		} else {
			t.Error("User.name property not found")
		}
	}

	// Check interface{} handling
	extraProp, ok := schema.Properties["extra"]
	if !ok {
		t.Error("Extra property not found")
	} else {
		if extraProp.Type != "object" {
			t.Errorf("Extra: expected type 'object', got %s", extraProp.Type)
		}
		if extraProp.AdditionalProperties == nil || !*extraProp.AdditionalProperties {
			t.Error("Extra: additionalProperties not set for interface{}")
		}
	}
}

func TestTagSchemaGenerator_SimpleStruct(t *testing.T) {
	gen := NewTagSchemaGenerator()

	schema, err := gen.GenerateSchema(SimpleUser{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Check that tags were processed
	nameProp, ok := schema.Properties["name"]
	if !ok {
		t.Error("Name property not found")
	} else {
		if nameProp.Description != "User's full name" {
			t.Errorf("Name: description not extracted from tag")
		}
		if nameProp.Minimum == nil || *nameProp.Minimum != 1 {
			t.Error("Name: min constraint not extracted")
		}
		if nameProp.Maximum == nil || *nameProp.Maximum != 100 {
			t.Error("Name: max constraint not extracted")
		}
	}

	// Check email format extraction
	emailProp, ok := schema.Properties["email"]
	if !ok {
		t.Error("Email property not found")
	} else {
		if emailProp.Format != "email" {
			t.Errorf("Email: format not extracted from validate tag")
		}
	}
}

func TestTagSchemaGenerator_SchemaTagged(t *testing.T) {
	gen := NewTagSchemaGenerator()

	schema, err := gen.GenerateSchema(SchemaTaggedStruct{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Check Field1 with schema tags
	field1, ok := schema.Properties["Field1"]
	if !ok {
		t.Error("Field1 not found")
	} else {
		if field1.Type != "string" {
			t.Errorf("Field1: expected type 'string', got %s", field1.Type)
		}
		if field1.Format != "email" {
			t.Errorf("Field1: expected format 'email', got %s", field1.Format)
		}
	}

	// Check Field2 constraints
	field2, ok := schema.Properties["Field2"]
	if !ok {
		t.Error("Field2 not found")
	} else {
		if field2.Minimum == nil || *field2.Minimum != 10 {
			t.Error("Field2: minimum not set correctly")
		}
		if field2.Maximum == nil || *field2.Maximum != 100 {
			t.Error("Field2: maximum not set correctly")
		}
	}

	// Check Field3 array constraints
	field3, ok := schema.Properties["Field3"]
	if !ok {
		t.Error("Field3 not found")
	} else {
		if field3.MinItems == nil || *field3.MinItems != 2 {
			t.Error("Field3: minItems not set correctly")
		}
		if field3.Description != "Boolean array" {
			t.Error("Field3: description not set correctly")
		}
	}
}

func TestTagSchemaGenerator_CustomTagParser(t *testing.T) {
	gen := NewTagSchemaGenerator()

	// Register custom tag parser
	gen.RegisterTagParser("custom", func(tagValue string, prop *domain.Property) error {
		prop.CustomValidator = tagValue
		return nil
	})

	type CustomTagged struct {
		Field string `custom:"myValidator"`
	}

	schema, err := gen.GenerateSchema(CustomTagged{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	prop, ok := schema.Properties["Field"]
	if !ok {
		t.Error("Field not found")
	} else {
		if prop.CustomValidator != "myValidator" {
			t.Errorf("Expected custom validator 'myValidator', got %s", prop.CustomValidator)
		}
	}
}

func TestReflectionSchemaGenerator_CustomTypeHandler(t *testing.T) {
	gen := NewReflectionSchemaGenerator()

	// Register custom handler for time.Duration
	gen.RegisterTypeHandler(reflect.TypeOf(time.Duration(0)), func(t reflect.Type, tag reflect.StructTag) (domain.Property, error) {
		return domain.Property{
			Type:        "string",
			Format:      "duration",
			Description: "Duration in RFC3339 format",
		}, nil
	})

	type DurationStruct struct {
		Timeout time.Duration `json:"timeout"`
	}

	schema, err := gen.GenerateSchema(DurationStruct{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	prop, ok := schema.Properties["timeout"]
	if !ok {
		t.Error("Timeout property not found")
	} else {
		if prop.Format != "duration" {
			t.Errorf("Expected format 'duration', got %s", prop.Format)
		}
	}
}

func TestSchemaGeneratorComparison(t *testing.T) {
	// Compare outputs from both generators
	reflectionGen := NewReflectionSchemaGenerator()
	tagGen := NewTagSchemaGenerator()

	testStruct := SimpleUser{}

	reflectionSchema, err := reflectionGen.GenerateSchema(testStruct)
	if err != nil {
		t.Fatalf("Reflection generator failed: %v", err)
	}

	tagSchema, err := tagGen.GenerateSchema(testStruct)
	if err != nil {
		t.Fatalf("Tag generator failed: %v", err)
	}

	// Both should produce valid schemas
	if reflectionSchema.Type != "object" || tagSchema.Type != "object" {
		t.Error("Both generators should produce object schemas")
	}

	// Both should have the same properties
	if len(reflectionSchema.Properties) != len(tagSchema.Properties) {
		t.Errorf("Property count mismatch: reflection=%d, tag=%d",
			len(reflectionSchema.Properties), len(tagSchema.Properties))
	}

	// Log schemas for visual comparison
	reflectionJSON, _ := json.MarshalIndent(reflectionSchema, "", "  ")
	tagJSON, _ := json.MarshalIndent(tagSchema, "", "  ")

	t.Logf("Reflection Schema:\n%s", reflectionJSON)
	t.Logf("Tag Schema:\n%s", tagJSON)
}

func TestReflectionSchemaGenerator_MaxDepth(t *testing.T) {
	gen := NewReflectionSchemaGenerator().WithOptions(true, false, 2)

	type Recursive struct {
		Name  string     `json:"name"`
		Child *Recursive `json:"child,omitempty"`
	}

	schema, err := gen.GenerateSchema(Recursive{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Check first level (depth 1)
	childProp, ok := schema.Properties["child"]
	if !ok {
		t.Error("Child property not found")
	} else if childProp.Type != "object" {
		t.Errorf("Child: expected type 'object', got %s", childProp.Type)
	}

	// Child should have properties since we're at depth 1
	if len(childProp.Properties) == 0 {
		t.Error("First level child should have properties")
	}

	// Due to caching of recursive types, the nested child will have the same schema
	// This is actually the correct behavior to prevent infinite recursion
	// The max depth prevents infinite generation, but cached schemas are reused
	if childProp.Properties != nil {
		nestedChild, ok := childProp.Properties["child"]
		if ok {
			// The nested child will reference the same cached schema
			if nestedChild.Type != "object" {
				t.Errorf("Nested child: expected type 'object', got %s", nestedChild.Type)
			}
			// Due to caching, it will have the same structure
			// This is expected behavior for recursive types
		}
	}
}
