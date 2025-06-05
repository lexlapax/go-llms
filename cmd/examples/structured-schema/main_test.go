package main

import (
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/adapter/reflection"
)

func TestSchemaGeneration(t *testing.T) {
	// Test Address schema generation
	t.Run("address schema", func(t *testing.T) {
		schema, err := reflection.GenerateSchema(Address{})
		if err != nil {
			t.Fatalf("Error generating Address schema: %v", err)
		}

		// Validate schema
		if schema.Type != "object" {
			t.Errorf("Expected schema type to be 'object', got %s", schema.Type)
		}

		// Check that required fields are present
		requiredFields := map[string]bool{
			"street":     false,
			"city":       false,
			"state":      false,
			"postalCode": false,
			"country":    false,
		}

		for _, field := range schema.Required {
			requiredFields[field] = true
		}

		for field, found := range requiredFields {
			if !found {
				t.Errorf("Required field '%s' not marked as required in schema", field)
			}
		}

		// Check property types
		if prop, ok := schema.Properties["street"]; ok {
			if prop.Type != "string" {
				t.Errorf("Expected street property type to be 'string', got %s", prop.Type)
			}
		} else {
			t.Errorf("Street property not found in schema")
		}

		// Check string length constraints
		if prop, ok := schema.Properties["state"]; ok {
			if prop.MinLength == nil {
				t.Logf("MinLength not set for state property")
			}
			if prop.MaxLength == nil {
				t.Logf("MaxLength not set for state property")
			}
		} else {
			t.Errorf("State property not found in schema")
		}
	})

	// Test Product schema generation
	t.Run("product schema", func(t *testing.T) {
		schema, err := reflection.GenerateSchema(Product{})
		if err != nil {
			t.Fatalf("Error generating Product schema: %v", err)
		}

		// Check category property exists and has the right type
		if prop, ok := schema.Properties["category"]; ok {
			if prop.Type != "string" {
				t.Errorf("Expected category property type to be 'string', got %s", prop.Type)
			}
		} else {
			t.Errorf("Category property not found in schema")
		}
	})

	// Test order schema with nested objects and arrays
	t.Run("order schema with nested types", func(t *testing.T) {
		schema, err := reflection.GenerateSchema(Order{})
		if err != nil {
			t.Fatalf("Error generating Order schema: %v", err)
		}

		// Check items array
		if prop, ok := schema.Properties["items"]; ok {
			if prop.Type != "array" {
				t.Errorf("Expected items property type to be 'array', got %s", prop.Type)
			}

			if prop.Items == nil {
				t.Errorf("Expected items to have an items schema")
			}
		} else {
			t.Errorf("Items property not found in schema")
		}

		// Check nested address object
		if prop, ok := schema.Properties["shippingAddress"]; ok {
			if prop.Type != "object" {
				t.Errorf("Expected shippingAddress property type to be 'object', got %s", prop.Type)
			}

			if len(prop.Properties) == 0 {
				t.Errorf("Expected shippingAddress to have properties")
			}
		} else {
			t.Errorf("ShippingAddress property not found in schema")
		}
	})
}

// TestGenerationPerformance measures the performance of schema generation
func TestGenerationPerformance(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	var orderSchema, productSchema, customerSchema interface{}
	var err error

	// Benchmark Order schema generation
	t.Run("order schema performance", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 1000; i++ {
			orderSchema, err = reflection.GenerateSchema(Order{})
			if err != nil {
				t.Fatalf("Error generating Order schema: %v", err)
			}
		}
		elapsed := time.Since(start)
		t.Logf("Generated 1000 Order schemas in %s (%s per schema)", elapsed, elapsed/1000)
	})

	// Benchmark Product schema generation
	t.Run("product schema performance", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 1000; i++ {
			productSchema, err = reflection.GenerateSchema(Product{})
			if err != nil {
				t.Fatalf("Error generating Product schema: %v", err)
			}
		}
		elapsed := time.Since(start)
		t.Logf("Generated 1000 Product schemas in %s (%s per schema)", elapsed, elapsed/1000)
	})

	// Benchmark Customer schema generation
	t.Run("customer schema performance", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 1000; i++ {
			customerSchema, err = reflection.GenerateSchema(Customer{})
			if err != nil {
				t.Fatalf("Error generating Customer schema: %v", err)
			}
		}
		elapsed := time.Since(start)
		t.Logf("Generated 1000 Customer schemas in %s (%s per schema)", elapsed, elapsed/1000)
	})

	// Use the schemas to prevent compiler optimization
	_ = orderSchema
	_ = productSchema
	_ = customerSchema
}
