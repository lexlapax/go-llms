package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	agentDomain "github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
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
	// Example 1: Simple provider creation with convenience function
	fmt.Println("\n=== Example 1: Provider Creation ===")

	// Try to create a provider using environment variables
	llmProvider, providerName, modelName, err := llmutil.ProviderFromEnv()
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	fmt.Printf("Using %s provider with model %s\n", providerName, modelName)

	// Example 2: Using batch generation
	fmt.Println("\n=== Example 2: Batch Generation ===")
	prompts := []string{
		"What is the capital of France?",
		"Give me a recipe for pancakes",
		"How many planets are in our solar system?",
	}

	results, errors := llmutil.BatchGenerate(context.Background(), llmProvider, prompts)

	for i, result := range results {
		if errors[i] != nil {
			fmt.Printf("Prompt %d error: %v\n", i+1, errors[i])
		} else {
			fmt.Printf("Prompt %d result snippet: %s...\n", i+1, truncate(result, 50))
		}
	}

	// Example 3: Generation with retry
	fmt.Println("\n=== Example 3: Generation with Retry ===")
	result, err := llmutil.GenerateWithRetry(
		context.Background(),
		llmProvider,
		"Write a haiku about programming",
		3, // max retries
	)

	if err != nil {
		fmt.Printf("Generation with retry failed: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}

	// Example 4: Provider pool
	fmt.Println("\n=== Example 4: Provider Pool ===")

	// Create multiple providers (for demonstration purposes)
	mockProvider1 := provider.NewMockProvider()
	mockProvider2 := provider.NewMockProvider()
	mockProvider3 := provider.NewMockProvider()

	// Create a provider pool with round-robin strategy
	providerPool := llmutil.NewProviderPool(
		[]domain.Provider{mockProvider1, mockProvider2, mockProvider3},
		llmutil.StrategyRoundRobin,
	)

	// Generate multiple responses using the pool
	for i := 0; i < 5; i++ {
		poolResult, poolErr := providerPool.Generate(
			context.Background(),
			fmt.Sprintf("This is test prompt %d", i+1),
		)

		if poolErr != nil {
			fmt.Printf("Pool generation %d error: %v\n", i+1, poolErr)
		} else {
			fmt.Printf("Pool generation %d result snippet: %s...\n", i+1, truncate(poolResult, 50))
		}
	}

	// Example 5: Typed generation
	fmt.Println("\n=== Example 5: Typed Generation ===")

	// Generate a product with typed output
	productPrompt := "Create a product listing for a high-end coffee maker"

	// Define a schema for the product
	productSchema := &schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"id": {
				Type:        "string",
				Description: "Unique product identifier",
			},
			"name": {
				Type:        "string",
				Description: "Product name",
			},
			"description": {
				Type:        "string",
				Description: "Product description",
			},
			"price": {
				Type:        "number",
				Description: "Product price in USD",
				Minimum:     float64Ptr(0),
			},
			"categories": {
				Type:        "array",
				Description: "Product categories",
				Items: &schemaDomain.Property{
					Type: "string",
				},
			},
			"inStock": {
				Type:        "boolean",
				Description: "Whether the product is in stock",
			},
		},
		Required: []string{"id", "name", "description", "price"},
	}

	// Generate with schema
	productResult, productErr := llmProvider.GenerateWithSchema(
		context.Background(),
		productPrompt,
		productSchema,
	)

	if productErr != nil {
		fmt.Printf("Product generation error: %v\n", productErr)
	} else {
		// Convert to JSON for display
		productJSON, _ := json.MarshalIndent(productResult, "", "  ")
		fmt.Printf("Generated product:\n%s\n", string(productJSON))

		// Convert result to product
		var product Product

		// This is simplified - in a real application we would use the structured processor
		productBytes, _ := json.Marshal(productResult)
		if err := json.Unmarshal(productBytes, &product); err != nil {
			fmt.Printf("Error converting to product: %v\n", err)
		} else {
			fmt.Printf("\nProduct Name: %s\n", product.Name)
			fmt.Printf("Product Price: $%.2f\n", product.Price)
		}
	}

	// Example 6: Agent creation with convenience function
	fmt.Println("\n=== Example 6: Agent Creation ===")

	// Create a simple calculator tool
	calculatorTool := tools.NewTool(
		"calculator",
		"Perform mathematical calculations",
		func(params struct {
			Expression string `json:"expression"`
		}) (map[string]interface{}, error) {
			return map[string]interface{}{
				"success":    true,
				"expression": params.Expression,
				"result":     42, // Placeholder - would call a real calculator function
			}, nil
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"expression": {
					Type:        "string",
					Description: "The mathematical expression to evaluate",
				},
			},
			Required: []string{"expression"},
		},
	)

	// Create an agent config
	agentConfig := llmutil.AgentConfig{
		Provider:      llmProvider,
		SystemPrompt:  "You are a helpful assistant with access to tools.",
		EnableCaching: true,
		Tools:         []agentDomain.Tool{calculatorTool},
		Hooks:         []agentDomain.Hook{workflow.NewMetricsHook()},
	}

	// Create the agent using the convenience function
	agent := llmutil.CreateAgent(agentConfig)

	// Run the agent with a timeout
	agentResult, agentErr := llmutil.RunWithTimeout(
		agent,
		"What is 7 * 6?",
		10*time.Second, // timeout
	)

	if agentErr != nil {
		fmt.Printf("Agent error: %v\n", agentErr)
	} else {
		fmt.Printf("Agent result: %v\n", agentResult)
	}

	fmt.Println("\nUtility Functions Demo Completed")
}

// Helper function to create float pointers
func float64Ptr(v float64) *float64 {
	return &v
}

// Helper function to truncate strings
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
