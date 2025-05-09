package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the GEMINI_API_KEY environment variable")
		os.Exit(1)
	}

	// Create a Gemini provider with default options
	model := "gemini-2.0-flash-lite" // Use flash-lite model
	geminiProvider := provider.NewGeminiProvider(apiKey, model)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("\n=== Simple Text Generation ===")
	simpleTextGeneration(ctx, geminiProvider)

	fmt.Println("\n=== Conversation ===")
	conversation(ctx, geminiProvider)

	fmt.Println("\n=== Structured Output ===")
	structuredOutput(ctx, geminiProvider)

	fmt.Println("\n=== Streaming Response ===")
	streamingResponse(ctx, geminiProvider)
}

// Simple text generation with a prompt
func simpleTextGeneration(ctx context.Context, provider *provider.GeminiProvider) {
	prompt := "Explain quantum computing in simple terms"
	
	// Generate text with the prompt
	response, err := provider.Generate(ctx, prompt, domain.WithTemperature(0.2))
	if err != nil {
		fmt.Printf("Error generating text: %v\n", err)
		return
	}
	
	fmt.Printf("Prompt: %s\n\nResponse:\n%s\n", prompt, response)
}

// Conversation with multiple messages
func conversation(ctx context.Context, provider *provider.GeminiProvider) {
	messages := []domain.Message{
		{Role: domain.RoleUser, Content: "Hello, I'd like to learn about machine learning."},
		{Role: domain.RoleAssistant, Content: "Hi there! I'd be happy to help you learn about machine learning. What specific aspects are you interested in?"},
		{Role: domain.RoleUser, Content: "Can you explain what neural networks are?"},
	}
	
	// Generate a response to the conversation
	response, err := provider.GenerateMessage(ctx, messages)
	if err != nil {
		fmt.Printf("Error generating conversation response: %v\n", err)
		return
	}
	
	// Display the conversation
	fmt.Println("Conversation:")
	for _, msg := range messages {
		role := "User"
		if msg.Role == domain.RoleAssistant {
			role = "Assistant"
		}
		fmt.Printf("%s: %s\n", role, msg.Content)
	}
	
	fmt.Printf("Assistant: %s\n", response.Content)
}

// Generate structured output with schema
func structuredOutput(ctx context.Context, provider *provider.GeminiProvider) {
	// Define a schema for a recipe
	recipeSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type:        "string",
				Description: "The name of the recipe",
			},
			"description": {
				Type:        "string",
				Description: "A brief description of the recipe",
			},
			"ingredients": {
				Type:        "array",
				Description: "List of ingredients",
				Items: &schemaDomain.Property{
					Type: "object",
					Properties: map[string]schemaDomain.Property{
						"name": {
							Type:        "string",
							Description: "Ingredient name",
						},
						"quantity": {
							Type:        "string",
							Description: "Quantity with unit",
						},
					},
					Required: []string{"name", "quantity"},
				},
			},
			"steps": {
				Type:        "array",
				Description: "Cooking instructions",
				Items: &schemaDomain.Property{
					Type: "string",
				},
			},
			"preparationTime": {
				Type:        "integer",
				Description: "Preparation time in minutes",
			},
			"cookingTime": {
				Type:        "integer",
				Description: "Cooking time in minutes",
			},
			"servings": {
				Type:        "integer",
				Description: "Number of servings",
			},
		},
		Required: []string{"name", "ingredients", "steps"},
	}
	
	// Generate a recipe
	prompt := "Create a recipe for a quick and healthy vegetarian pasta dish"
	result, err := provider.GenerateWithSchema(ctx, prompt, recipeSchema)
	if err != nil {
		fmt.Printf("Error generating structured output: %v\n", err)
		return
	}
	
	// Convert to map for easier access
	recipe, ok := result.(map[string]interface{})
	if !ok {
		fmt.Println("Error: Could not convert result to map")
		return
	}
	
	// Print the recipe details
	fmt.Printf("Recipe: %s\n", recipe["name"])
	if desc, ok := recipe["description"].(string); ok {
		fmt.Printf("Description: %s\n", desc)
	}
	
	fmt.Println("\nIngredients:")
	if ingredients, ok := recipe["ingredients"].([]interface{}); ok {
		for _, item := range ingredients {
			if ingredient, ok := item.(map[string]interface{}); ok {
				name := ingredient["name"].(string)
				quantity := ingredient["quantity"].(string)
				fmt.Printf("- %s (%s)\n", name, quantity)
			}
		}
	}
	
	fmt.Println("\nSteps:")
	if steps, ok := recipe["steps"].([]interface{}); ok {
		for i, step := range steps {
			fmt.Printf("%d. %s\n", i+1, step.(string))
		}
	}
	
	if prepTime, ok := recipe["preparationTime"].(float64); ok {
		fmt.Printf("\nPreparation Time: %d minutes\n", int(prepTime))
	}
	
	if cookTime, ok := recipe["cookingTime"].(float64); ok {
		fmt.Printf("Cooking Time: %d minutes\n", int(cookTime))
	}
	
	if servings, ok := recipe["servings"].(float64); ok {
		fmt.Printf("Servings: %d\n", int(servings))
	}
}

// Stream responses token by token
func streamingResponse(ctx context.Context, provider *provider.GeminiProvider) {
	prompt := "Write a short poem about artificial intelligence"
	
	// Stream responses
	stream, err := provider.Stream(ctx, prompt)
	if err != nil {
		fmt.Printf("Error creating stream: %v\n", err)
		return
	}
	
	fmt.Printf("Prompt: %s\n\nStreaming response:\n", prompt)
	
	// Process tokens as they arrive
	var fullResponse string
	fmt.Print("Beginning stream: ")
	tokenCount := 0
	for token := range stream {
		// Print token immediately with some debug info
		fmt.Print(token.Text)
		tokenCount++
		
		// Build full response
		fullResponse += token.Text
		
		// Small delay to simulate real-time display
		time.Sleep(5 * time.Millisecond)
	}
	
	fmt.Printf("\n\nStream complete. Received %d tokens.\n", tokenCount)
	fmt.Println("Full response:\n", fullResponse)
}

