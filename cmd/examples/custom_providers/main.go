package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

func main() {
	fmt.Println("=== Custom Providers Example ===")
	fmt.Println("This example demonstrates how to use custom LLM providers with go-llms")
	fmt.Println("1. OpenRouter API (OpenAI API-compatible)")
	fmt.Println("2. Ollama (local LLM provider)")
	fmt.Println()

	// Check which examples to run
	runOpenRouter := os.Getenv("OPENROUTER_API_KEY") != ""
	runOllama := os.Getenv("OLLAMA_HOST") != ""

	if !runOpenRouter && !runOllama {
		fmt.Println("No API keys or configuration found. Please set one of the following:")
		fmt.Println("- OPENROUTER_API_KEY for OpenRouter API")
		fmt.Println("- OLLAMA_HOST for Ollama (e.g., http://localhost:11434)")
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
	fmt.Println("\n--- OpenRouter Example ---")
	fmt.Println("OpenRouter provides access to various LLM providers with an OpenAI-compatible API")

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	
	// Get the model name from environment variable or use default
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "anthropic/claude-3-5-sonnet"
	}

	// Method 1: Direct provider instantiation with options
	fmt.Println("\nMethod 1: Direct provider instantiation")
	openRouterProvider := provider.NewOpenAIProvider(
		apiKey,
		model,
		provider.WithBaseURL("https://openrouter.ai/api"),
	)

	// Use the provider to generate a response
	response, err := openRouterProvider.Generate(
		context.Background(),
		"What models do you provide access to?",
		domain.WithMaxTokens(150),
	)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response from model %s:\n%s\n", model, response)
	}

	// Method 2: Using ModelConfig and CreateProvider
	fmt.Println("\nMethod 2: Using ModelConfig and CreateProvider")
	config := llmutil.ModelConfig{
		Provider: "openai",
		Model:    model,
		APIKey:   apiKey,
		BaseURL:  "https://openrouter.ai/api",
	}

	openRouterProvider2, err := llmutil.CreateProvider(config)
	if err != nil {
		fmt.Printf("Error creating provider: %v\n", err)
		return
	}

	// Use the provider to generate a response
	response2, err := openRouterProvider2.Generate(
		context.Background(),
		"What is your latency like?",
		domain.WithMaxTokens(150),
	)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response:\n%s\n", response2)
	}

	// Method 3: Using environment variables
	fmt.Println("\nMethod 3: Using environment variables")
	fmt.Println("Set OPENAI_BASE_URL and OPENAI_API_KEY in your environment")
	fmt.Println("Example:")
	fmt.Println("export OPENAI_BASE_URL=https://openrouter.ai/api")
	fmt.Println("export OPENAI_API_KEY=your_openrouter_key")
	fmt.Println("export OPENAI_MODEL=anthropic/claude-3-5-sonnet")
}

// Ollama Example
// Ollama allows running LLMs locally
func runOllamaExample() {
	fmt.Println("\n--- Ollama Example ---")
	fmt.Println("Ollama allows you to run LLMs locally")

	// Get Ollama host and model from environment variables
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}

	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "llama3"
	}

	// No API key is typically needed for Ollama
	apiKey := ""

	// Method 1: Direct provider instantiation with options
	fmt.Println("\nMethod 1: Direct provider instantiation")
	ollamaProvider := provider.NewOpenAIProvider(
		apiKey,
		ollamaModel,
		provider.WithBaseURL(ollamaHost),
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
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response from model %s:\n%s\n", ollamaModel, response)
	}

	// Method 2: Using ModelConfig and CreateProvider
	fmt.Println("\nMethod 2: Using ModelConfig and CreateProvider")
	config := llmutil.ModelConfig{
		Provider: "openai",
		Model:    ollamaModel,
		APIKey:   apiKey, // Empty string for Ollama
		BaseURL:  ollamaHost,
	}

	ollamaProvider2, err := llmutil.CreateProvider(config)
	if err != nil {
		fmt.Printf("Error creating provider: %v\n", err)
		return
	}

	// Use the provider for streaming
	fmt.Println("\nStreaming with Ollama:")
	
	stream, err := ollamaProvider2.Stream(
		ctx,
		"List three projects that use Ollama",
		domain.WithMaxTokens(150),
	)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Streaming response:")
	for token := range stream {
		fmt.Print(token.Text)
		if token.Finished {
			fmt.Println()
		}
	}

	// Method 3: Using environment variables
	fmt.Println("\nMethod 3: Using environment variables")
	fmt.Println("Set OPENAI_BASE_URL, OPENAI_API_KEY (empty), and OPENAI_MODEL in your environment")
	fmt.Println("Example:")
	fmt.Println("export OPENAI_BASE_URL=http://localhost:11434")
	fmt.Println("export OPENAI_API_KEY=\"\"")
	fmt.Println("export OPENAI_MODEL=llama3")
}