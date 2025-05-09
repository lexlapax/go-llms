package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
	// Check if at least one API key is provided
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	geminiKey := os.Getenv("GEMINI_API_KEY")

	if openaiKey == "" && anthropicKey == "" && geminiKey == "" {
		fmt.Println("No API keys found. Using mock providers for demonstration.")
		runWithMockProviders()
		return
	}

	// Create real providers with available API keys
	var providers []provider.ProviderWeight

	if openaiKey != "" {
		openaiProvider := provider.NewOpenAIProvider(
			openaiKey,
			"gpt-4o", // You can change this to your preferred model
		)
		providers = append(providers, provider.ProviderWeight{
			Provider: openaiProvider,
			Weight:   1.0,
			Name:     "openai",
		})
	}

	if anthropicKey != "" {
		anthropicProvider := provider.NewAnthropicProvider(
			anthropicKey,
			"claude-3-5-sonnet-latest", // You can change this to your preferred model
		)
		providers = append(providers, provider.ProviderWeight{
			Provider: anthropicProvider,
			Weight:   1.0,
			Name:     "anthropic",
		})
	}

	if geminiKey != "" {
		geminiProvider := provider.NewGeminiProvider(
			geminiKey,
			"gemini-2.0-flash-lite", // You can change this to your preferred model
		)
		providers = append(providers, provider.ProviderWeight{
			Provider: geminiProvider,
			Weight:   1.0,
			Name:     "gemini",
		})
	}

	if len(providers) < 2 {
		// Add a mock provider to ensure we have at least two providers for demonstration
		mockProvider := provider.NewMockProvider().WithGenerateFunc(
			func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
				return "[MOCK PROVIDER] Response to: " + prompt, nil
			})
		providers = append(providers, provider.ProviderWeight{
			Provider: mockProvider,
			Weight:   1.0,
			Name:     "mock",
		})
	}

	runWithRealProviders(providers)
}

func runWithMockProviders() {
	// Create mock providers with different latency characteristics
	fastMockProvider := createDelayedMockProvider(100*time.Millisecond, "FAST")
	mediumMockProvider := createDelayedMockProvider(300*time.Millisecond, "MEDIUM")
	slowMockProvider := createDelayedMockProvider(500*time.Millisecond, "SLOW")

	providers := []provider.ProviderWeight{
		{Provider: slowMockProvider, Weight: 1.0, Name: "slow"},
		{Provider: mediumMockProvider, Weight: 1.0, Name: "medium"},
		{Provider: fastMockProvider, Weight: 1.0, Name: "fast"},
	}

	// Create a multi-provider with the fastest strategy
	fmt.Println("\n=== Multi-Provider Example with Simulated Providers ===")
	fmt.Println("\nStrategy: FASTEST (returns the fastest response)")
	fastestProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

	// Record start time
	start := time.Now()
	response, err := fastestProvider.Generate(context.Background(), "What is Go's concurrency model?")
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", response)
		fmt.Printf("Time taken: %v\n", elapsed)
	}

	// Create a multi-provider with the primary strategy
	fmt.Println("\nStrategy: PRIMARY (tries primary first, falls back to others)")
	primaryProvider := provider.NewMultiProvider(providers, provider.StrategyPrimary).
		WithPrimaryProvider(0) // Use slow provider as primary

	// Record start time
	start = time.Now()
	response, err = primaryProvider.Generate(context.Background(), "What is Go's concurrency model?")
	elapsed = time.Since(start)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", response)
		fmt.Printf("Time taken: %v\n", elapsed)
	}

	// Demonstrate streaming with MultiProvider
	fmt.Println("\nStreaming with MultiProvider:")
	stream, err := fastestProvider.Stream(context.Background(), "List three benefits of Go's garbage collector")
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
}

func runWithRealProviders(providers []provider.ProviderWeight) {
	fmt.Println("\n=== Multi-Provider Example with Real Providers ===")

	// Create a multi-provider with the fastest strategy
	fmt.Println("\nStrategy: FASTEST (returns the fastest response)")
	fastestProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

	// Record start time
	start := time.Now()
	response, err := fastestProvider.Generate(context.Background(), "What is Go's concurrency model?")
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", response)
		fmt.Printf("Time taken: %v\n", elapsed)
	}

	// Create a multi-provider with the primary strategy
	fmt.Println("\nStrategy: PRIMARY (tries primary first, falls back to others)")
	primaryProvider := provider.NewMultiProvider(providers, provider.StrategyPrimary).
		WithPrimaryProvider(0)

	// Record start time
	start = time.Now()
	response, err = primaryProvider.Generate(context.Background(), "Explain the difference between goroutines and OS threads.")
	elapsed = time.Since(start)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", response)
		fmt.Printf("Time taken: %v\n", elapsed)
	}

	// Test message-based conversation
	fmt.Println("\nMessage-based conversation:")
	messages := []domain.Message{
		{Role: domain.RoleSystem, Content: "You are a Go programming expert."},
		{Role: domain.RoleUser, Content: "What are some best practices for error handling in Go?"},
	}

	start = time.Now()
	messageResponse, err := fastestProvider.GenerateMessage(context.Background(), messages)
	elapsed = time.Since(start)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", messageResponse.Content)
		fmt.Printf("Time taken: %v\n", elapsed)
	}
}

// createDelayedMockProvider creates a mock provider with specified delay
func createDelayedMockProvider(delay time.Duration, prefix string) domain.Provider {
	return provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
				return fmt.Sprintf("[%s PROVIDER] Response after %v delay to: %s",
					prefix, delay, prompt), nil
			}
		}).WithStreamFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
			tokenCh := make(chan domain.Token, 5)

			go func() {
				defer close(tokenCh)

				// Wait for initial delay
				select {
				case <-ctx.Done():
					return
				case <-time.After(delay):
					// Continue after delay
				}

				// Send tokens with small delays between them
				tokens := []string{
					fmt.Sprintf("[%s PROVIDER] ", prefix),
					"First benefit: Automatic memory management. ",
					"Second benefit: Low pause times. ",
					"Third benefit: Concurrent collection.",
				}

				for i, token := range tokens {
					// Check for cancellation
					select {
					case <-ctx.Done():
						return
					case tokenCh <- domain.GetTokenPool().NewToken(token, i == len(tokens)-1):
						// Token sent successfully
						time.Sleep(50 * time.Millisecond) // Small delay between tokens
					}
				}
			}()

			return tokenCh, nil
		})
}
