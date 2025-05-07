package main

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TestSimpleExample ensures that the simple example can be executed
func TestSimpleExample(t *testing.T) {
	// Initialize the mock provider for testing
	mockProvider := provider.NewMockProvider()
	
	// Test text generation
	t.Run("TextGeneration", func(t *testing.T) {
		response, err := mockProvider.Generate(nil, "Tell me a joke")
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		
		if response == "" {
			t.Error("Expected non-empty response")
		}
	})
	
	// Test schema validation
	t.Run("SchemaValidation", func(t *testing.T) {
		schema := &domain.Schema{
			Type: "object",
			Properties: map[string]domain.Property{
				"name": {Type: "string"},
				"age":  {Type: "integer"},
			},
			Required: []string{"name"},
		}
		
		_, err := mockProvider.GenerateWithSchema(nil, "Generate a person", schema)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})
}