package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Person defines a struct for our schema
type Person struct {
	Name  string `json:"name" validate:"required" description:"Person's name"`
	Age   int    `json:"age" validate:"min=0,max=120" description:"Age in years"`
	Email string `json:"email" validate:"required,email" description:"Email address"`
}

func main() {
	// Create a schema manually
	schema := &domain.Schema{
		Type: "object",
		Properties: map[string]domain.Property{
			"name": {Type: "string", Description: "Person's name"},
			"age": {Type: "integer", Minimum: float64Ptr(0), Maximum: float64Ptr(120)},
			"email": {Type: "string", Format: "email", Description: "Email address"},
		},
		Required: []string{"name", "email"},
	}

	// Create a mock LLM provider
	llmProvider := provider.NewMockProvider()

	fmt.Println("Go-LLMs Simple Example")
	fmt.Println("======================")

	// Example 1: Using the mock provider with a simple prompt
	fmt.Println("\nExample 1: Simple generation")
	response, err := llmProvider.Generate(context.Background(), "Tell me about Go")
	if err != nil {
		log.Fatalf("Generation error: %v", err)
	}
	fmt.Printf("Response: %s\n", response)

	// Example 2: Using the mock provider with schema
	fmt.Println("\nExample 2: Structured generation with schema")
	result, err := llmProvider.GenerateWithSchema(
		context.Background(),
		"Generate information about a person",
		schema,
	)
	if err != nil {
		log.Fatalf("Structured generation error: %v", err)
	}

	// Pretty print the result
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Structured Response:\n%s\n", string(resultJSON))

	// Example 3: Streaming responses
	fmt.Println("\nExample 3: Streaming response")
	stream, err := llmProvider.Stream(context.Background(), "Tell me about Go")
	if err != nil {
		log.Fatalf("Stream error: %v", err)
	}

	fmt.Print("Streamed Response: ")
	for token := range stream {
		fmt.Print(token.Text)
		if token.Finished {
			fmt.Println()
		}
	}
}

// Helper function for creating float pointers
func float64Ptr(v float64) *float64 {
	return &v
}