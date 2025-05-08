package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/schema/validation"
	structuredProcessor "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// We can use the Recipe type from main.go since we're in the same package

// TestAnthropicWithMock tests the Anthropic example functionality using a mock provider
func TestAnthropicWithMock(t *testing.T) {
	// Create a mock provider that simulates Anthropic responses
	mockProvider := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			// For recipe schema (simulating a structured response)
			if len(prompt) > 100 {
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
			return "Mock Anthropic response for: " + prompt, nil
		},
	)

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

	// Test basic text generation
	t.Run("BasicGeneration", func(t *testing.T) {
		response, err := mockProvider.Generate(context.Background(), "Explain what Go channels are")
		if err != nil {
			t.Fatalf("Generation failed: %v", err)
		}
		if response == "" {
			t.Error("Expected non-empty response")
		}
	})

	// Test message-based conversation
	t.Run("MessageConversation", func(t *testing.T) {
		messages := []domain.Message{
			{Role: domain.RoleSystem, Content: "You are a helpful coding assistant."},
			{Role: domain.RoleUser, Content: "What's the difference between a slice and an array in Go?"},
		}
		response, err := mockProvider.GenerateMessage(context.Background(), messages)
		if err != nil {
			t.Fatalf("Message generation failed: %v", err)
		}
		if response.Content == "" {
			t.Error("Expected non-empty message response")
		}
	})

	// Test structured recipe generation
	t.Run("StructuredRecipeGeneration", func(t *testing.T) {
		validator := validation.NewValidator()
		processor := structuredProcessor.NewStructuredProcessor(validator)
		promptEnhancer := structuredProcessor.NewPromptEnhancer()

		// Enhance the prompt
		prompt := "Generate a simple vegetarian pasta recipe"
		enhancedPrompt, err := promptEnhancer.Enhance(prompt, recipeSchema)
		if err != nil {
			t.Fatalf("Prompt enhancement failed: %v", err)
		}

		// Generate response
		response, err := mockProvider.Generate(context.Background(), enhancedPrompt)
		if err != nil {
			t.Fatalf("Structured generation failed: %v", err)
		}

		// Process the response
		recipeData, err := processor.Process(recipeSchema, response)
		if err != nil {
			t.Fatalf("Processing failed: %v", err)
		}

		// Validate the recipe data
		jsonData, _ := json.Marshal(recipeData)
		var recipe Recipe
		if err := json.Unmarshal(jsonData, &recipe); err != nil {
			t.Fatalf("JSON unmarshaling failed: %v", err)
		}

		// Verify the recipe
		if recipe.Title == "" {
			t.Error("Expected non-empty recipe title")
		}
		if len(recipe.Ingredients) == 0 {
			t.Error("Expected non-empty ingredients list")
		}
		if len(recipe.Steps) == 0 {
			t.Error("Expected non-empty steps list")
		}
	})

	// Test streaming
	t.Run("Streaming", func(t *testing.T) {
		stream, err := mockProvider.Stream(context.Background(), "List 3 benefits of Go's garbage collector")
		if err != nil {
			t.Fatalf("Stream creation failed: %v", err)
		}

		tokens := []string{}
		for token := range stream {
			tokens = append(tokens, token.Text)
		}

		if len(tokens) == 0 {
			t.Error("Expected at least one token from the stream")
		}
	})
}

// float64Ptr is already defined in main.go