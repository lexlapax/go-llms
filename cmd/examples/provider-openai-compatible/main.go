package main

// ABOUTME: Example demonstrating usage with OpenAI API compatible providers
// ABOUTME: Shows how to use alternative providers like OpenRouter and Ollama

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

func main() {
	log.Println("=== OpenAI API Compatible Providers Example ===")
	log.Println("This example demonstrates how to use providers that implement the OpenAI API specification")
	log.Println("1. OpenRouter API (OpenAI API-compatible)")
	log.Println("2. Ollama (local LLM provider with OpenAI API compatibility)")
	log.Println()

	// Check which examples to run
	runOpenRouter := os.Getenv("OPENROUTER_API_KEY") != ""
	runOllama := os.Getenv("OLLAMA_HOST") != ""

	if !runOpenRouter && !runOllama {
		log.Println("No API keys or configuration found. Please set one of the following:")
		log.Println("- OPENROUTER_API_KEY for OpenRouter API")
		log.Println("- OLLAMA_HOST for Ollama (e.g., http://localhost:11434)")
		return
	}

	// Run examples
	if runOpenRouter {
		runOpenRouterExample()
	}

	if runOllama {
		runOllamaExample()
	}
}

// OpenRouter Example
// OpenRouter provides access to many models with an OpenAI-compatible API
func runOpenRouterExample() {
	log.Println("\n--- OpenRouter Example ---")
	log.Println("OpenRouter provides access to various LLM providers with an OpenAI-compatible API")

	apiKey := os.Getenv("OPENROUTER_API_KEY")

	// Get the model name from environment variable or use default
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "mistralai/mistral-small-3.1-24b-instruct:free"
	}

	// Method 1: Direct provider instantiation with interface-based options
	log.Println("\nMethod 1: Direct provider instantiation with interface-based options")

	// Create a custom HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Create the provider options
	// For OpenRouter, we need to omit the "/v1" as the OpenAI provider will add it
	baseURLOption := domain.NewBaseURLOption("https://openrouter.ai/api")
	httpClientOption := domain.NewHTTPClientOption(httpClient)
	headersOption := domain.NewHeadersOption(map[string]string{
		"HTTP-Referer": "https://github.com/lexlapax/go-llms", // OpenRouter attribution
		"X-Title":      "Go-LLMs Example",                     // Additional OpenRouter headers
	})

	// Create the provider with multiple options
	openRouterProvider := provider.NewOpenAIProvider(
		apiKey,
		model,
		baseURLOption,
		httpClientOption,
		headersOption,
	)

	// Use the provider to generate a response with messages (preferred for OpenRouter)
	messages := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "What models do you provide access to?"),
	}

	messageResp, err := openRouterProvider.GenerateMessage(
		context.Background(),
		messages,
		domain.WithMaxTokens(150),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		log.Printf("Response from model %s:\n%s\n", model, messageResp.Content)
	}

	// Method 2: Using ModelConfig and CreateProvider
	log.Println("\nMethod 2: Using ModelConfig and CreateProvider")

	// NOTE: With future ModelConfig improvements, we'd use the HeadersOption with ModelConfig
	// For now, this is just shown as an example of the option pattern
	// openRouterHeaders := domain.NewHeadersOption(map[string]string{
	//    "HTTP-Referer": "https://github.com/lexlapax/go-llms",
	// })

	config := llmutil.ModelConfig{
		Provider: "openai",
		Model:    model,
		APIKey:   apiKey,
		BaseURL:  "https://openrouter.ai/api",
	}

	openRouterProvider2, err := llmutil.CreateProvider(config)
	if err != nil {
		log.Printf("Error creating provider: %v\n", err)
		return
	}

	// Use the provider to generate a response with messages (preferred for OpenRouter)
	messages2 := []domain.Message{
		domain.NewTextMessage(domain.RoleUser, "What is your latency like?"),
	}

	messageResp2, err := openRouterProvider2.GenerateMessage(
		context.Background(),
		messages2,
		domain.WithMaxTokens(150),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		log.Printf("Response:\n%s\n", messageResp2.Content)
	}

	// Method 3: Using environment variables
	log.Println("\nMethod 3: Using environment variables")
	log.Println("Set OPENAI_BASE_URL and OPENAI_API_KEY in your environment")
	log.Println("Example:")
	log.Println("export OPENAI_BASE_URL=https://openrouter.ai/api")
	log.Println("export OPENAI_API_KEY=your_openrouter_key")
	log.Println("export OPENAI_MODEL=mistralai/mistral-small-3.1-24b-instruct:free")
}

// Ollama Example
// Ollama allows running LLMs locally
func runOllamaExample() {
	log.Println("\n--- Ollama Example ---")
	log.Println("Ollama allows you to run LLMs locally")

	// Get Ollama host and model from environment variables
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}

	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "llama3.2:3b"
	}

	// Ollama doesn't need a real API key, but the OpenAI provider requires a non-empty string
	apiKey := "dummy-key" // This key is ignored by Ollama but prevents errors in the provider

	// Method 1: Direct provider instantiation with interface-based options
	log.Println("\nMethod 1: Direct provider instantiation with interface-based options")

	// Create a custom HTTP client with timeout
	ollamaClient := &http.Client{
		Timeout: 60 * time.Second, // Longer timeout for local models
	}

	// Create the provider options
	ollamaBaseURLOption := domain.NewBaseURLOption(ollamaHost)
	ollamaHTTPClientOption := domain.NewHTTPClientOption(ollamaClient)

	// Create the provider with multiple options
	ollamaProvider := provider.NewOpenAIProvider(
		apiKey,
		ollamaModel,
		ollamaBaseURLOption,
		ollamaHTTPClientOption,
	)

	// Use the provider to generate a response
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := ollamaProvider.Generate(
		ctx,
		"What are the benefits of running LLMs locally?",
		domain.WithMaxTokens(150),
		domain.WithTemperature(0.7),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		log.Printf("Response from model %s:\n%s\n", ollamaModel, response)
	}

	// Method 2: Using ModelConfig and CreateProvider
	log.Println("\nMethod 2: Using ModelConfig and CreateProvider")
	config := llmutil.ModelConfig{
		Provider: "openai",
		Model:    ollamaModel,
		APIKey:   apiKey, // Dummy key for Ollama
		BaseURL:  ollamaHost,
	}

	ollamaProvider2, err := llmutil.CreateProvider(config)
	if err != nil {
		log.Printf("Error creating provider: %v\n", err)
		return
	}

	// Use the provider for streaming
	log.Println("\nStreaming with Ollama:")

	stream, err := ollamaProvider2.Stream(
		ctx,
		"List three projects that use llama",
		domain.WithMaxTokens(150),
	)

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	log.Println("Streaming response:")
	for token := range stream {
		log.Print(token.Text)
		if token.Finished {
			log.Println()
		}
	}

	// Method 3: Using environment variables
	log.Println("\nMethod 3: Using environment variables")
	log.Println("Set OPENAI_BASE_URL, OPENAI_API_KEY (empty), and OPENAI_MODEL in your environment")
	log.Println("Example:")
	log.Println("export OPENAI_BASE_URL=http://localhost:11434")
	log.Println("export OPENAI_API_KEY=\"\"")
	log.Println("export OPENAI_MODEL=llama3.2:3b")
}
