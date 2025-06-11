package main

// ABOUTME: Example demonstrating the dedicated Ollama provider functionality
// ABOUTME: Shows model listing, generation, streaming, and multimodal features

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/fetchers"
)

func main() {
	log.Println("=== Ollama Provider Example ===")
	log.Println("This example demonstrates the Ollama provider's features:")
	log.Println("1. Listing available models")
	log.Println("2. Basic text generation")
	log.Println("3. Streaming responses")
	log.Println("4. Conversation with context")
	log.Println("5. Custom configuration options")
	log.Println()

	// Check if Ollama is running
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}
	log.Printf("Using Ollama host: %s\n", ollamaHost)
	log.Println()

	// Example 1: List available models
	log.Println("--- Example 1: Listing Available Models ---")
	listAvailableModels(ollamaHost)

	// Get model from environment or use default
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.2:3b"
		log.Printf("No OLLAMA_MODEL set, using default: %s\n", model)
	}
	log.Println()

	// Example 2: Basic text generation with convenience wrapper
	log.Println("--- Example 2: Basic Text Generation ---")
	basicGeneration(model, ollamaHost)

	// Example 3: Streaming response
	log.Println("\n--- Example 3: Streaming Response ---")
	streamingExample(model, ollamaHost)

	// Example 4: Conversation with context
	log.Println("\n--- Example 4: Conversation with Context ---")
	conversationExample(model, ollamaHost)

	// Example 5: Custom configuration
	log.Println("\n--- Example 5: Custom Configuration ---")
	customConfigExample(model)

	// Example 6: Using standard OpenAI provider (for comparison)
	log.Println("\n--- Example 6: Using OpenAI Provider with Ollama ---")
	openAIProviderExample(model, ollamaHost)
}

func listAvailableModels(host string) {
	fetcher := fetchers.NewOllamaFetcher(host, nil)
	models, err := fetcher.FetchModels()
	if err != nil {
		log.Printf("Error fetching models: %v\n", err)
		log.Println("Make sure Ollama is running at", host)
		return
	}

	log.Printf("Found %d models:\n", len(models))
	for _, model := range models {
		log.Printf("  - %s: %s\n", model.Name, model.Description)
		log.Printf("    Context Window: %d tokens\n", model.ContextWindow)
		log.Printf("    Capabilities: Streaming=%v, Functions=%v, Vision=%v\n",
			model.Capabilities.Streaming,
			model.Capabilities.FunctionCalling,
			model.Capabilities.Image.Read)
	}
}

func basicGeneration(model, host string) {
	// Create Ollama provider with convenience wrapper
	provider := provider.NewOllamaProvider(model,
		provider.WithOllamaHost(host),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	prompt := "Explain quantum computing in 2-3 sentences, suitable for a 10-year-old."
	response, err := provider.Generate(ctx, prompt,
		domain.WithTemperature(0.7),
		domain.WithMaxTokens(150),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	log.Printf("Response:\n%s\n", response)
}

func streamingExample(model, host string) {
	provider := provider.NewOllamaProvider(model,
		provider.WithOllamaHost(host),
		provider.WithOllamaTimeout(60*time.Second),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	prompt := "Write a haiku about programming in Go."
	stream, err := provider.Stream(ctx, prompt,
		domain.WithTemperature(0.8),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	log.Println("Streaming response:")
	for token := range stream {
		fmt.Print(token.Text)
		if token.Finished {
			fmt.Println()
		}
	}
}

func conversationExample(model, host string) {
	provider := provider.NewOllamaProvider(model,
		provider.WithOllamaHost(host),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleSystem, "You are a helpful programming tutor."),
		domain.NewTextMessage(domain.RoleUser, "What is a goroutine in Go?"),
		domain.NewTextMessage(domain.RoleAssistant, "A goroutine is a lightweight thread of execution in Go. It's one of Go's most powerful features for concurrent programming. You can start a goroutine by using the 'go' keyword before a function call."),
		domain.NewTextMessage(domain.RoleUser, "Can you show me a simple example?"),
	}

	response, err := provider.GenerateMessage(ctx, messages,
		domain.WithMaxTokens(200),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	log.Printf("Assistant: %s\n", response.Content)
}

func customConfigExample(model string) {
	// Example with custom host and timeout
	customHost := "http://192.168.1.100:11434" // Example custom host
	customTimeout := 2 * time.Minute

	provider := provider.NewOllamaProvider(model,
		provider.WithOllamaHost(customHost),
		provider.WithOllamaTimeout(customTimeout),
	)

	// This will likely fail unless you actually have Ollama at that address
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := provider.Generate(ctx, "Hello", domain.WithMaxTokens(10))
	if err != nil {
		log.Printf("Expected error (custom host likely not available): %v\n", err)
	} else {
		log.Println("Custom configuration worked!")
	}
}

func openAIProviderExample(model, host string) {
	// For comparison, showing how to use the standard OpenAI provider
	// This is what NewOllamaProvider does under the hood
	provider := provider.NewOpenAIProvider(
		"dummy-key",
		model,
		domain.NewBaseURLOption(host),
		domain.NewHTTPClientOption(&http.Client{Timeout: 60 * time.Second}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := provider.Generate(ctx, "What makes Go a good language for backend development? Give me 3 bullet points.",
		domain.WithMaxTokens(150),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	log.Printf("Response using OpenAI provider:\n%s\n", response)
}
