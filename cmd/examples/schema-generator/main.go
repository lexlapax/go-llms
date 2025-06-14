// Example demonstrating schema generation
package main

// ABOUTME: Example showing how to generate JSON schemas from Go structs
// ABOUTME: Demonstrates both reflection and tag-based generation approaches

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/generator"
)

// User demonstrates basic struct with validation tags
type User struct {
	ID        string                 `json:"id" validate:"required,uuid" format:"uuid" description:"Unique user identifier"`
	Username  string                 `json:"username" validate:"required,min=3,max=20" pattern:"^[a-zA-Z0-9_]+$"`
	Email     string                 `json:"email" validate:"required,email"`
	Age       int                    `json:"age,omitempty" validate:"min=13,max=120"`
	Premium   bool                   `json:"premium"`
	CreatedAt time.Time              `json:"created_at" format:"date-time"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Product demonstrates complex nested structures
type Product struct {
	SKU         string      `json:"sku" pattern:"^[A-Z]{3}-[0-9]{4}$" validate:"required"`
	Name        string      `json:"name" minLength:"1" maxLength:"200"`
	Description string      `json:"description,omitempty" maxLength:"1000"`
	Price       Money       `json:"price" validate:"required"`
	Categories  []string    `json:"categories" minItems:"1" uniqueItems:"true"`
	Attributes  []Attribute `json:"attributes,omitempty"`
	Reviews     []Review    `json:"reviews,omitempty"`
}

type Money struct {
	Amount   float64 `json:"amount" minimum:"0"`
	Currency string  `json:"currency" enum:"USD,EUR,GBP,JPY" validate:"required,oneof=USD EUR GBP JPY"`
}

type Attribute struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type Review struct {
	Rating  int    `json:"rating" minimum:"1" maximum:"5"`
	Comment string `json:"comment,omitempty"`
	Author  string `json:"author"`
}

// SchemaTaggedExample uses schema-specific tags
type SchemaTaggedExample struct {
	StringField string   `schema:"type=string,format=email,description=Email address"`
	NumberField float64  `schema:"type=number,minimum=0,maximum=100"`
	IntField    int      `schema:"type=integer,minimum=1"`
	BoolField   bool     `schema:"type=boolean"`
	ArrayField  []string `schema:"type=array,minItems=1,maxItems=10"`
}

// RecursiveNode demonstrates handling of recursive structures
type RecursiveNode struct {
	ID       string           `json:"id"`
	Value    interface{}      `json:"value"`
	Children []*RecursiveNode `json:"children,omitempty"`
}

func main() {
	fmt.Println("Schema Generation Examples")
	fmt.Println("=========================")

	// Example 1: Basic reflection-based generation
	fmt.Println("1. Reflection-Based Generation")
	demoReflectionGenerator()

	// Example 2: Tag-based generation
	fmt.Println("\n2. Tag-Based Generation")
	demoTagGenerator()

	// Example 3: Custom type handlers
	fmt.Println("\n3. Custom Type Handlers")
	demoCustomTypeHandlers()

	// Example 4: Recursive structures
	fmt.Println("\n4. Recursive Structure Handling")
	demoRecursiveStructures()

	// Example 5: Comparison of generators
	fmt.Println("\n5. Generator Comparison")
	compareGenerators()
}

func demoReflectionGenerator() {
	gen := generator.NewReflectionSchemaGenerator()

	schema, err := gen.GenerateSchema(User{})
	if err != nil {
		log.Fatalf("Failed to generate schema: %v", err)
	}

	fmt.Println("Generated schema for User struct:")
	printSchema(schema)

	// With options
	gen2 := generator.NewReflectionSchemaGenerator().
		WithOptions(true, false, 5) // preserveGoTypes, includePrivate, maxDepth

	schema2, _ := gen2.GenerateSchema(Product{})
	fmt.Println("\nGenerated schema for Product struct (with Go type info):")
	printSchema(schema2)
}

func demoTagGenerator() {
	gen := generator.NewTagSchemaGenerator()

	// Set custom tag priority
	gen.SetTagPriority([]string{"schema", "validate", "json"})

	schema, err := gen.GenerateSchema(SchemaTaggedExample{})
	if err != nil {
		log.Fatalf("Failed to generate schema: %v", err)
	}

	fmt.Println("Generated schema using tag-based generator:")
	printSchema(schema)
}

func demoCustomTypeHandlers() {
	gen := generator.NewReflectionSchemaGenerator()

	// Register custom handler for time.Duration
	gen.RegisterTypeHandler(reflect.TypeOf(time.Duration(0)),
		func(t reflect.Type, tag reflect.StructTag) (domain.Property, error) {
			return domain.Property{
				Type:        "string",
				Format:      "duration",
				Description: "Duration in string format (e.g., '1h30m')",
			}, nil
		})

	type Config struct {
		Timeout     time.Duration `json:"timeout" description:"Request timeout"`
		RetryDelay  time.Duration `json:"retry_delay" description:"Delay between retries"`
		MaxIdleTime time.Duration `json:"max_idle_time" description:"Maximum idle time"`
	}

	schema, _ := gen.GenerateSchema(Config{})
	fmt.Println("Schema with custom Duration handler:")
	printSchema(schema)
}

func demoRecursiveStructures() {
	gen := generator.NewReflectionSchemaGenerator().
		WithOptions(true, false, 3) // Max depth of 3

	schema, err := gen.GenerateSchema(RecursiveNode{})
	if err != nil {
		log.Fatalf("Failed to generate schema: %v", err)
	}

	fmt.Println("Schema for recursive structure (max depth 3):")
	// For recursive structures, we'll just print key properties to avoid cycles
	fmt.Printf("  Type: %s\n", schema.Type)
	fmt.Printf("  Title: %s\n", schema.Title)
	fmt.Printf("  Properties: %d\n", len(schema.Properties))
	if childProp, ok := schema.Properties["children"]; ok {
		fmt.Printf("  Children type: %s (array of %s)\n", childProp.Type, childProp.Items.Type)
	}
}

func compareGenerators() {
	type ComparisonStruct struct {
		Name  string   `json:"name" validate:"required,min=1" description:"Object name"`
		Value int      `json:"value" minimum:"0" maximum:"100"`
		Tags  []string `json:"tags,omitempty" uniqueItems:"true"`
	}

	reflectionGen := generator.NewReflectionSchemaGenerator()
	tagGen := generator.NewTagSchemaGenerator()

	schema1, _ := reflectionGen.GenerateSchema(ComparisonStruct{})
	schema2, _ := tagGen.GenerateSchema(ComparisonStruct{})

	fmt.Println("Reflection-based generator output:")
	printSchema(schema1)

	fmt.Println("\nTag-based generator output:")
	printSchema(schema2)

	fmt.Println("\nKey differences:")
	fmt.Println("- Reflection generator extracts more type information")
	fmt.Println("- Tag generator prioritizes explicit tags over type inference")
	fmt.Println("- Both handle validation tags but may interpret them differently")
}

func printSchema(schema interface{}) {
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal schema: %v", err)
		return
	}
	fmt.Println(string(data))
}
