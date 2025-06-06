package main

// ABOUTME: Example demonstrating provider-level convenience functions from llmutil
// ABOUTME: Shows simplified provider initialization, batch generation, and typed output

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

// Product is a sample struct for typed generation
type Product struct {
	ID          string   `json:"id" validate:"required" description:"Unique product identifier"`
	Name        string   `json:"name" validate:"required" description:"Product name"`
	Description string   `json:"description" validate:"required" description:"Product description"`
	Price       float64  `json:"price" validate:"min=0" description:"Product price in USD"`
	Categories  []string `json:"categories" description:"Product categories"`
	InStock     bool     `json:"inStock" description:"Whether the product is in stock"`
}

// Review is a sample struct for reviews
type Review struct {
	ID        string  `json:"id" validate:"required"`
	ProductID string  `json:"productId" validate:"required"`
	UserName  string  `json:"userName" validate:"required"`
	Rating    float64 `json:"rating" validate:"min=1,max=5"`
	Comment   string  `json:"comment"`
	Date      string  `json:"date" validate:"required"`
}

func main() {
	fmt.Println("=== Provider-Level Convenience Functions ===")

	// Example 1: Simple provider creation with convenience function
	fmt.Println("Example 1: Provider Creation from Environment")
	fmt.Println("-------------------------------------------")

	// Try to create a provider using environment variables
	llmProvider, providerName, modelName, err := llmutil.ProviderFromEnv()
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	fmt.Printf("Created %s provider with model %s\n", providerName, modelName)

	// Example 2: Using batch generation
	fmt.Println("\nExample 2: Batch Generation")
	fmt.Println("---------------------------")

	prompts := []string{
		"What is the capital of France? Answer in one word.",
		"What is 2+2? Answer with just the number.",
		"What color is the sky on a clear day? Answer in one word.",
	}

	fmt.Println("Sending batch of prompts...")
	results, errors := llmutil.BatchGenerate(context.Background(), llmProvider, prompts)

	for i, result := range results {
		if errors[i] != nil {
			fmt.Printf("  Prompt %d error: %v\n", i+1, errors[i])
		} else {
			fmt.Printf("  Prompt %d: \"%s\" -> %s\n", i+1, prompts[i], truncate(result, 50))
		}
	}

	// Example 3: Generation with retry
	fmt.Println("\nExample 3: Generation with Retry")
	fmt.Println("--------------------------------")

	result, err := llmutil.GenerateWithRetry(
		context.Background(),
		llmProvider,
		"Write a haiku about programming",
		3, // max retries
	)

	if err != nil {
		fmt.Printf("Generation with retry failed: %v\n", err)
	} else {
		fmt.Printf("Generated haiku:\n%s\n", result)
	}

	// Example 4: Provider pool with round-robin
	fmt.Println("\nExample 4: Provider Pool (Round-Robin)")
	fmt.Println("--------------------------------------")

	// Create multiple providers for pooling
	// In real usage, these might be different models or providers
	var providers []domain.Provider

	// Add the main provider
	providers = append(providers, llmProvider)

	// Add mock providers for demonstration
	providers = append(providers,
		provider.NewMockProvider(),
		provider.NewMockProvider(),
	)

	// Create a provider pool with round-robin strategy
	providerPool := llmutil.NewProviderPool(
		providers,
		llmutil.StrategyRoundRobin,
	)

	// Generate multiple responses using the pool
	fmt.Println("Making requests through the pool...")
	for i := 0; i < 5; i++ {
		poolResult, poolErr := providerPool.Generate(
			context.Background(),
			fmt.Sprintf("This is request %d. What provider am I using?", i+1),
		)

		if poolErr != nil {
			fmt.Printf("  Request %d error: %v\n", i+1, poolErr)
		} else {
			fmt.Printf("  Request %d handled (response: %s...)\n", i+1, truncate(poolResult, 30))
		}
	}

	// Example 5: Typed generation with schema
	fmt.Println("\nExample 5: Typed Generation with Schema")
	fmt.Println("---------------------------------------")

	// Generate a product with typed output
	productPrompt := "Create a detailed product listing for a premium espresso machine. Include all required fields: id, name, description, price, categories (array), and inStock (boolean)."

	// Use ProcessTypedWithProvider for structured output
	var product Product
	typedErr := llmutil.ProcessTypedWithProvider(
		context.Background(),
		llmProvider,
		productPrompt,
		&product,
	)

	if typedErr != nil {
		fmt.Printf("Typed generation error: %v\n", typedErr)
	} else {
		fmt.Println("Generated product:")
		jsonBytes, _ := json.MarshalIndent(product, "  ", "  ")
		fmt.Printf("  %s\n", jsonBytes)
	}

	// Example 6: Provider configuration with ModelConfig
	fmt.Println("\nExample 6: Custom Provider Configuration")
	fmt.Println("----------------------------------------")

	// Create a provider with custom configuration
	config := llmutil.ModelConfig{
		Provider:  providerName,
		Model:     modelName,
		APIKey:    "", // Will use from environment
		MaxTokens: 500,
		Options:   []domain.ProviderOption{
			// Options will be provider-specific
		},
	}

	customProvider, err := llmutil.CreateProvider(config)
	if err != nil {
		fmt.Printf("Failed to create custom provider: %v\n", err)
	} else {
		fmt.Println("Created provider with custom configuration")

		// Use the custom provider
		response, err := customProvider.Generate(
			context.Background(),
			"Tell me a fun fact about coffee in one sentence.",
		)

		if err != nil {
			fmt.Printf("Custom provider error: %v\n", err)
		} else {
			fmt.Printf("Fun fact: %s\n", response)
		}
	}

	fmt.Println("\nProvider Convenience Functions Demo Completed!")
}

// Helper function to truncate strings
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
