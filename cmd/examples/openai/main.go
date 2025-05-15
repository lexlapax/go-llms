package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	structuredProcessor "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// Recipe represents a cooking recipe
type Recipe struct {
	Title       string   `json:"title"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
	PrepTime    int      `json:"prepTime"`
	CookTime    int      `json:"cookTime"`
	Servings    int      `json:"servings"`
	Difficulty  string   `json:"difficulty"`
}

func main() {
	// Check if API key is provided
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// If no API key, use mock provider instead
		fmt.Println("No OPENAI_API_KEY environment variable found. Using mock provider instead.")
		runWithMockProvider()
		return
	}

	// Optional organization ID
	orgID := os.Getenv("OPENAI_ORGANIZATION")

	// Create the OpenAI provider with organization option
	var openaiProvider *provider.OpenAIProvider

	if orgID != "" {
		// Include the organization option if provided
		orgOption := domain.NewOpenAIOrganizationOption(orgID)

		openaiProvider = provider.NewOpenAIProvider(
			apiKey,
			"gpt-4o",
			orgOption,
		)

		fmt.Println("Using OpenAI organization ID:", orgID)
	} else {
		// No organization ID provided
		openaiProvider = provider.NewOpenAIProvider(
			apiKey,
			"gpt-4o",
		)
	}

	// Create structured processor components
	validator := validation.NewValidator()
	processor := structuredProcessor.NewStructuredProcessor(validator)
	promptEnhancer := structuredProcessor.NewPromptEnhancer()

	// Create a schema for recipes
	recipeSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"title": {
				Type:        "string",
				Description: "The name of the recipe",
			},
			"ingredients": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "string",
				},
				Description: "List of ingredients with quantities",
			},
			"steps": {
				Type: "array",
				Items: &schemaDomain.Property{
					Type: "string",
				},
				Description: "List of cooking instructions",
			},
			"prepTime": {
				Type:        "integer",
				Description: "Preparation time in minutes",
				Minimum:     float64Ptr(0),
			},
			"cookTime": {
				Type:        "integer",
				Description: "Cooking time in minutes",
				Minimum:     float64Ptr(0),
			},
			"servings": {
				Type:        "integer",
				Description: "Number of servings",
				Minimum:     float64Ptr(1),
			},
			"difficulty": {
				Type:        "string",
				Description: "Difficulty level",
				Enum:        []string{"easy", "medium", "hard"},
			},
		},
		Required: []string{"title", "ingredients", "steps", "cookTime", "servings"},
	}

	fmt.Println("Go-LLMs OpenAI Example")
	fmt.Println("======================")

	// Example 1: Simple generation
	fmt.Println("\nExample 1: Simple generation")
	response, err := openaiProvider.Generate(
		context.Background(),
		"Explain what Go channels are in a paragraph",
	)
	if err != nil {
		log.Fatalf("Generation error: %v", err)
	}
	fmt.Printf("Response: %s\n", response)

	// Example 2: Using message-based conversation with system role
	fmt.Println("\nExample 2: Message-based conversation")
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleSystem, "You are a helpful coding assistant specializing in Go."),
		domain.NewTextMessage(domain.RoleUser, "What's the difference between a slice and an array in Go?"),
	}
	messageResponse, err := openaiProvider.GenerateMessage(context.Background(), messages)
	if err != nil {
		log.Fatalf("Message generation error: %v", err)
	}
	fmt.Printf("Response: %s\n", messageResponse.Content)

	// Example 3: Structured output with schema
	fmt.Println("\nExample 3: Structured recipe generation with schema")

	// Enhance the prompt with schema information
	prompt := "Generate a simple vegetarian pasta recipe"
	enhancedPrompt, err := promptEnhancer.Enhance(prompt, recipeSchema)
	if err != nil {
		log.Fatalf("Prompt enhancement error: %v", err)
	}

	// Generate the structured output
	structuredResponse, err := openaiProvider.Generate(context.Background(), enhancedPrompt)
	if err != nil {
		log.Fatalf("Structured generation error: %v", err)
	}

	// Process the raw response
	recipeData, err := processor.Process(recipeSchema, structuredResponse)
	if err != nil {
		log.Fatalf("Processing error: %v", err)
	}

	// Convert to our Recipe struct
	var recipe Recipe
	recipeJSON, _ := json.Marshal(recipeData)
	if err := json.Unmarshal(recipeJSON, &recipe); err != nil {
		log.Fatalf("JSON unmarshaling error: %v", err)
	}

	// Display the recipe
	fmt.Printf("Recipe: %s\n", recipe.Title)
	fmt.Printf("Difficulty: %s\n", recipe.Difficulty)
	fmt.Printf("Prep time: %d minutes, Cook time: %d minutes, Servings: %d\n",
		recipe.PrepTime, recipe.CookTime, recipe.Servings)

	fmt.Println("\nIngredients:")
	for _, ingredient := range recipe.Ingredients {
		fmt.Printf("- %s\n", ingredient)
	}

	fmt.Println("\nSteps:")
	for i, step := range recipe.Steps {
		fmt.Printf("%d. %s\n", i+1, step)
	}

	// Example 4: Stream the response
	fmt.Println("\nExample 4: Streaming response")
	stream, err := openaiProvider.Stream(
		context.Background(),
		"List 3 benefits of Go's garbage collector in short bullet points",
	)
	if err != nil {
		log.Fatalf("Stream error: %v", err)
	}

	fmt.Println("Streamed Response:")
	for token := range stream {
		fmt.Print(token.Text)
		if token.Finished {
			fmt.Println()
		}
	}
}

// runWithMockProvider runs the example with a mock provider
func runWithMockProvider() {
	// Create a mock provider with organization option
	orgOption := domain.NewOpenAIOrganizationOption("mock-org-id")

	mockProvider := provider.NewMockProvider(orgOption)

	// Set a custom response for recipe generation
	mockProvider.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		if len(prompt) > 100 && prompt[:100] != "" {
			// This is likely the recipe prompt
			return `{
				"title": "Creamy Mushroom Pasta",
				"ingredients": [
					"8 oz pasta",
					"8 oz mushrooms, sliced",
					"2 cloves garlic, minced",
					"1 cup heavy cream",
					"1/2 cup grated Parmesan cheese",
					"2 tbsp olive oil",
					"Salt and pepper to taste",
					"Fresh parsley, chopped"
				],
				"steps": [
					"Cook pasta according to package directions.",
					"Heat olive oil in a large skillet over medium heat.",
					"Add mushrooms and cook for 5 minutes until browned.",
					"Add garlic and cook for 1 minute until fragrant.",
					"Pour in cream and bring to a simmer.",
					"Add Parmesan cheese and stir until melted.",
					"Drain pasta and add to the sauce.",
					"Season with salt and pepper.",
					"Garnish with parsley and serve."
				],
				"prepTime": 10,
				"cookTime": 20,
				"servings": 4,
				"difficulty": "easy"
			}`, nil
		}
		return "This is a mock response for prompt: " + prompt, nil
	})

	fmt.Println("Go-LLMs Mock Example (simulating OpenAI)")
	fmt.Println("========================================")

	fmt.Println("\nExample 1: Simple generation")
	response, _ := mockProvider.Generate(
		context.Background(),
		"Explain what Go channels are in a paragraph",
	)
	fmt.Printf("Response: %s\n", response)

	// See the rest of the examples in the main function...
}

// Helper function for creating float pointers
func float64Ptr(v float64) *float64 {
	return &v
}
