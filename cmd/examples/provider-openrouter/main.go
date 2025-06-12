package main

// ABOUTME: Example of using the OpenRouter provider to access 400+ models
// ABOUTME: Demonstrates basic usage, streaming, and model discovery with OpenRouter

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/fetchers"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set OPENROUTER_API_KEY environment variable")
	}

	// Example 1: Basic usage with a free model
	fmt.Println("=== Example 1: Basic Generation with Free Model ===")
	if err := basicGeneration(apiKey); err != nil {
		log.Printf("Basic generation error: %v", err)
	}

	// Example 2: Streaming with a more powerful model
	fmt.Println("\n=== Example 2: Streaming Generation ===")
	if err := streamingGeneration(apiKey); err != nil {
		log.Printf("Streaming generation error: %v", err)
	}

	// Example 3: Model discovery
	fmt.Println("\n=== Example 3: Model Discovery ===")
	if err := modelDiscovery(apiKey); err != nil {
		log.Printf("Model discovery error: %v", err)
	}

	// Example 4: Using specific provider models
	fmt.Println("\n=== Example 4: Using Specific Provider Models ===")
	if err := specificProviderModels(apiKey); err != nil {
		log.Printf("Specific provider error: %v", err)
	}
}

func basicGeneration(apiKey string) error {
	// Create provider with a free model
	llm := provider.NewOpenRouterProvider(apiKey, "deepseek/deepseek-chat:free",
		domain.NewHeadersOption(map[string]string{
			"HTTP-Referer": "https://github.com/lexlapax/go-llms",
			"X-Title":      "Go-LLMs Example",
		}))

	// Create messages
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "What are the benefits of using Go for backend development? List 3 key points."),
	}

	// Generate response
	ctx := context.Background()
	response, err := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(200))
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Printf("Response from %s:\n%s\n", "deepseek/deepseek-chat:free", response.Content)
	return nil
}

func streamingGeneration(apiKey string) error {
	// Create provider with OpenAI GPT-4 through OpenRouter
	llm := provider.NewOpenRouterProvider(apiKey, "deepseek/deepseek-chat:free",
		domain.NewHeadersOption(map[string]string{
			"HTTP-Referer": "https://github.com/lexlapax/go-llms",
			"X-Title":      "Go-LLMs Example",
		}))

	// Create messages
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "Write a haiku about programming in Go."),
	}

	// Generate streaming response
	ctx := context.Background()
	stream, err := llm.StreamMessage(ctx, messages, domain.WithMaxTokens(100))
	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	fmt.Print("Streaming response: ")
	// Read the stream
	for token := range stream {
		if token.Text != "" {
			fmt.Print(token.Text)
		}
	}
	fmt.Println()

	return nil
}

func modelDiscovery(apiKey string) error {
	// Create fetcher
	fetcher := fetchers.NewOpenRouterFetcher(apiKey)

	// Fetch available models
	models, err := fetcher.FetchModels()
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}

	fmt.Printf("Found %d models available through OpenRouter\n", len(models))

	// Show some example models from different providers
	providers := make(map[string]int)
	freeModels := 0

	for _, model := range models {
		// Extract provider from model name
		parts := strings.SplitN(model.Name, "/", 2)
		if len(parts) > 0 {
			provider := strings.TrimPrefix(model.Provider, "openrouter:")
			providers[provider]++
		}

		// Check if it's a free model
		if strings.HasSuffix(model.Name, ":free") {
			freeModels++
		}
	}

	fmt.Printf("\nModels by provider:\n")
	for provider, count := range providers {
		fmt.Printf("  %s: %d models\n", provider, count)
	}
	fmt.Printf("\nFree models available: %d\n", freeModels)

	// Show a few interesting models
	fmt.Println("\nSome interesting models:")
	shown := 0
	for _, model := range models {
		if shown >= 5 {
			break
		}
		if model.Capabilities.Image.Read || strings.Contains(model.Name, "gpt-4") || strings.HasSuffix(model.Name, ":free") {
			fmt.Printf("  - %s (Context: %d tokens", model.Name, model.ContextWindow)
			if model.MaxOutputTokens > 0 {
				fmt.Printf(", Max output: %d", model.MaxOutputTokens)
			}
			if model.Capabilities.Image.Read {
				fmt.Printf(", Multimodal")
			}
			if strings.HasSuffix(model.Name, ":free") {
				fmt.Printf(", FREE")
			}
			fmt.Println(")")
			shown++
		}
	}

	// Show ALL free models
	fmt.Println("\nAll free models available:")
	for _, model := range models {
		if strings.HasSuffix(model.Name, ":free") {
			fmt.Printf("  - %s\n", model.Name)
		}
	}

	return nil
}

func specificProviderModels(apiKey string) error {
	// Test different provider models through OpenRouter
	// NOTE: Model IDs can change. Check https://openrouter.ai/models for current model list
	// If you get 405 errors, the model ID may be incorrect or deprecated
	models := []string{
		"deepseek/deepseek-chat:free",                  // deepseek model
		"nvidia/llama-3.1-nemotron-ultra-253b-v1:free", // nvidia model
		"meta-llama/llama-3.3-8b-instruct:free",        // Meta model
	}

	prompt := "Complete this sentence: The key to happiness is"

	for _, modelName := range models {
		fmt.Printf("\nTesting %s:\n", modelName)

		// Create provider
		llm := provider.NewOpenRouterProvider(apiKey, modelName,
			domain.NewHeadersOption(map[string]string{
				"HTTP-Referer": "https://github.com/lexlapax/go-llms",
				"X-Title":      "Go-LLMs Example",
			}))

		// Create messages
		messages := []domain.Message{
			domain.NewTextMessage(domain.RoleUser, prompt),
		}

		ctx := context.Background()
		response, err := llm.GenerateMessage(ctx, messages, domain.WithMaxTokens(50))
		if err != nil {
			// Some models might not be available or require specific permissions
			errStr := err.Error()
			if strings.Contains(errStr, "405") {
				fmt.Printf("  Error: Model '%s' returned 405 Method Not Allowed.\n", modelName)
				fmt.Printf("         This usually means the model ID is incorrect or the model has been deprecated.\n")
				fmt.Printf("         Check https://openrouter.ai/models for valid model IDs.\n")
			} else {
				fmt.Printf("  Error: %v\n", err)
			}
			continue
		}

		fmt.Printf("  Response: %s\n", response.Content)
	}

	return nil
}
