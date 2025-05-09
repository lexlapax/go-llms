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

	// Create a Gemini provider with generation config options
	model := "gemini-2.0-flash-lite" // Use flash-lite model

	// Create generation config option with multiple settings
	generationConfigOption := domain.NewGeminiGenerationConfigOption().
		WithTemperature(0.7).
		WithTopK(40).
		WithMaxOutputTokens(1024).
		WithTopP(0.95)

	// Create safety settings option (optional)
	safetySettings := []map[string]interface{}{
		{
			"category":  "HARM_CATEGORY_HARASSMENT",
			"threshold": "BLOCK_MEDIUM_AND_ABOVE",
		},
		{
			"category":  "HARM_CATEGORY_HATE_SPEECH",
			"threshold": "BLOCK_MEDIUM_AND_ABOVE",
		},
	}
	safetySettingsOption := domain.NewGeminiSafetySettingsOption(safetySettings)

	// Create the provider with options
	geminiProvider := provider.NewGeminiProvider(
		apiKey,
		model,
		generationConfigOption,
		safetySettingsOption,
	)

	fmt.Println("Using Gemini with generation config and safety settings options")

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

	fmt.Println("\n=== Generation Config Options Comparison ===")
	demonstrateGenerationConfigOptions(ctx, apiKey, model)
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

// demonstrateGenerationConfigOptions shows the impact of different generation config options
func demonstrateGenerationConfigOptions(ctx context.Context, apiKey, modelName string) {
	fmt.Println("Demonstrating the impact of different GeminiGenerationConfigOption settings")

	// Common prompt for comparison
	prompt := "Tell me a joke about programming"

	// 1. Create a provider with high temperature (more creative/random)
	highTempConfig := domain.NewGeminiGenerationConfigOption().
		WithTemperature(1.0).
		WithTopK(40)

	highTempProvider := provider.NewGeminiProvider(
		apiKey,
		modelName,
		highTempConfig,
	)

	// 2. Create a provider with low temperature (more focused/deterministic)
	lowTempConfig := domain.NewGeminiGenerationConfigOption().
		WithTemperature(0.1).
		WithTopK(40)

	lowTempProvider := provider.NewGeminiProvider(
		apiKey,
		modelName,
		lowTempConfig,
	)

	// 3. Create a provider with top-P sampling (nucleus sampling)
	topPConfig := domain.NewGeminiGenerationConfigOption().
		WithTemperature(0.7).
		WithTopP(0.5) // Lower top-P means more focused on higher probability tokens

	topPProvider := provider.NewGeminiProvider(
		apiKey,
		modelName,
		topPConfig,
	)

	// Generate with high temperature
	fmt.Println("\n--- High Temperature (1.0) ---")
	fmt.Println("More creative and varied outputs, potentially less coherent")

	highTempResponse, err := highTempProvider.Generate(ctx, prompt)
	if err != nil {
		fmt.Printf("Error generating with high temperature: %v\n", err)
	} else {
		fmt.Printf("Response:\n%s\n", highTempResponse)
	}

	// Generate with low temperature
	fmt.Println("\n--- Low Temperature (0.1) ---")
	fmt.Println("More focused and deterministic outputs, potentially more repetitive")

	lowTempResponse, err := lowTempProvider.Generate(ctx, prompt)
	if err != nil {
		fmt.Printf("Error generating with low temperature: %v\n", err)
	} else {
		fmt.Printf("Response:\n%s\n", lowTempResponse)
	}

	// Generate with top-P sampling
	fmt.Println("\n--- Top-P Sampling (0.5) ---")
	fmt.Println("Nucleus sampling focuses on the most likely tokens")

	topPResponse, err := topPProvider.Generate(ctx, prompt)
	if err != nil {
		fmt.Printf("Error generating with top-P sampling: %v\n", err)
	} else {
		fmt.Printf("Response:\n%s\n", topPResponse)
	}

	// Explain the differences
	fmt.Println("\nSummary:")
	fmt.Println("- Temperature controls randomness. Higher values (e.g., 1.0) increase diversity")
	fmt.Println("- Top-K limits the tokens considered to the K most likely")
	fmt.Println("- Top-P (nucleus sampling) considers the smallest set of tokens whose cumulative probability exceeds P")
	fmt.Println("- Adjusting these parameters lets you balance creativity vs determinism")
}

