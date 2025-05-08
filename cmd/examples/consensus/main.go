package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func main() {
	// Set up basic logging
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))

	// Check for API keys
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")

	providers := []provider.ProviderWeight{}

	// Validate required API keys
	if openaiKey == "" && anthropicKey == "" {
		fmt.Println("No API keys found. Using mock providers for demonstration.")
		fmt.Println("For accurate results with real models, provide OPENAI_API_KEY and/or ANTHROPIC_API_KEY.")
		fmt.Println()

		// Create mock providers with varying characteristics for testing
		providers = createMockProviders()
	} else {
		// Add real providers based on available API keys
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
			fmt.Println("Using OpenAI GPT-4o provider")
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
			fmt.Println("Using Anthropic Claude provider")
		}

		// Add a mock provider to ensure we have enough providers for demonstration
		if len(providers) < 3 {
			mockProvider := provider.NewMockProvider()
			providers = append(providers, provider.ProviderWeight{
				Provider: mockProvider,
				Weight:   0.5, // Lower weight for the mock provider
				Name:     "mock",
			})
			fmt.Println("Added mock provider to ensure 3+ providers for better consensus demonstration")
		}
	}

	fmt.Println("\n===== Multi-Provider Consensus Strategy Example =====")
	
	// Show basic usage first
	demonstrateBasicConsensus(providers)
	
	// Demonstrate different consensus strategies
	demonstrateConsensusStrategies(providers)
	
	// Demonstrate handling contradictions
	demonstrateContradictions(providers)
	
	// Demonstrate weighted consensus
	demonstrateWeightedConsensus(providers)
}

// demonstrateBasicConsensus shows the basic usage of the consensus strategy
func demonstrateBasicConsensus(providers []provider.ProviderWeight) {
	fmt.Println("\n== Basic Consensus Usage ==")
	
	// Create multi-provider with default consensus configuration
	consensusProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus)
	
	fmt.Println("Asking a factual question to multiple providers...")
	start := time.Now()
	
	// Use a factual question that should have consistent answers
	response, err := consensusProvider.Generate(
		context.Background(),
		"What is the capital of France? Provide a one-word answer.",
	)
	
	elapsed := time.Since(start)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Consensus response: %s\n", response)
	fmt.Printf("Time taken: %v\n", elapsed)
	fmt.Println("Note: The consensus strategy uses majority voting by default.")
}

// demonstrateConsensusStrategies shows the different consensus strategies
func demonstrateConsensusStrategies(providers []provider.ProviderWeight) {
	fmt.Println("\n== Different Consensus Strategies ==")
	
	// Define the prompt
	prompt := "Name three principles of object-oriented programming. List only the names, separated by commas."
	
	// 1. Majority consensus
	majorityProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus).
		WithConsensusStrategy(provider.ConsensusMajority)
	
	fmt.Println("\n1. MAJORITY STRATEGY")
	fmt.Println("Strategy: Return the most common response")
	
	start := time.Now()
	majorityResponse, err := majorityProvider.Generate(context.Background(), prompt)
	elapsed := time.Since(start)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", majorityResponse)
		fmt.Printf("Time taken: %v\n", elapsed)
	}
	
	// 2. Similarity consensus
	similarityProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus).
		WithConsensusStrategy(provider.ConsensusSimilarity).
		WithSimilarityThreshold(0.7) // 70% similarity threshold
	
	fmt.Println("\n2. SIMILARITY STRATEGY")
	fmt.Println("Strategy: Group responses by similarity and return most common group")
	fmt.Println("Similarity threshold: 70%")
	
	start = time.Now()
	similarityResponse, err := similarityProvider.Generate(context.Background(), prompt)
	elapsed = time.Since(start)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", similarityResponse)
		fmt.Printf("Time taken: %v\n", elapsed)
	}
	
	// 3. Weighted consensus
	weightedProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus).
		WithConsensusStrategy(provider.ConsensusWeighted)
	
	fmt.Println("\n3. WEIGHTED STRATEGY")
	fmt.Println("Strategy: Consider provider weights in addition to response similarity")
	
	start = time.Now()
	weightedResponse, err := weightedProvider.Generate(context.Background(), prompt)
	elapsed = time.Since(start)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Response: %s\n", weightedResponse)
		fmt.Printf("Time taken: %v\n", elapsed)
	}
}

// demonstrateContradictions shows how consensus handles contradictory information
func demonstrateContradictions(providers []provider.ProviderWeight) {
	fmt.Println("\n== Handling Contradictions ==")
	
	// Create specialized mock providers for this test
	contradictionProviders := []provider.ProviderWeight{
		{
			Provider: provider.NewMockProvider().WithGenerateFunc(
				func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
					return "The Earth is approximately 4.54 billion years old.", nil
				}),
			Weight: 1.0,
			Name:   "correct-provider",
		},
		{
			Provider: provider.NewMockProvider().WithGenerateFunc(
				func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
					return "The Earth is approximately 4.5 billion years old.", nil
				}),
			Weight: 1.0,
			Name:   "similar-provider",
		},
		{
			Provider: provider.NewMockProvider().WithGenerateFunc(
				func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
					return "The Earth is 6,000 years old.", nil
				}),
			Weight: 1.0,
			Name:   "incorrect-provider",
		},
	}
	
	// Using real providers if available
	if len(providers) >= 3 {
		contradictionProviders = providers
	}
	
	// Create a provider with similarity consensus
	consensusProvider := provider.NewMultiProvider(contradictionProviders, provider.StrategyConsensus).
		WithConsensusStrategy(provider.ConsensusSimilarity).
		WithSimilarityThreshold(0.7)
	
	fmt.Println("Asking a question that might produce contradictory responses...")
	fmt.Println("Prompt: How old is the Earth?")
	
	start := time.Now()
	response, err := consensusProvider.Generate(
		context.Background(),
		"How old is the Earth?",
	)
	elapsed := time.Since(start)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Consensus response: %s\n", response)
	fmt.Printf("Time taken: %v\n", elapsed)
	
	fmt.Println("\nNote: With similarity consensus, responses are grouped by similarity,")
	fmt.Println("and the largest group is chosen. This helps filter out outlier responses.")
}

// demonstrateWeightedConsensus shows how to weight different providers
func demonstrateWeightedConsensus(providers []provider.ProviderWeight) {
	fmt.Println("\n== Weighted Consensus ==")
	
	// Create a copy of providers and adjust weights
	weightedProviders := make([]provider.ProviderWeight, len(providers))
	copy(weightedProviders, providers)
	
	// Adjust weights for demonstration (first provider has higher weight)
	if len(weightedProviders) > 0 {
		weightedProviders[0].Weight = 2.0
		fmt.Printf("Giving %s provider double weight (2.0)\n", weightedProviders[0].Name)
	}
	
	// Create a provider with weighted consensus
	weightedConsensusProvider := provider.NewMultiProvider(weightedProviders, provider.StrategyConsensus).
		WithConsensusStrategy(provider.ConsensusWeighted)
	
	fmt.Println("\nAsking a subjective question where weighting might matter...")
	fmt.Println("Prompt: What's the best programming language for beginners?")
	
	start := time.Now()
	response, err := weightedConsensusProvider.Generate(
		context.Background(),
		"What's the best programming language for beginners? Keep your answer brief.",
	)
	elapsed := time.Since(start)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Weighted consensus response: %s\n", response)
	fmt.Printf("Time taken: %v\n", elapsed)
	
	fmt.Println("\nNote: With weighted consensus, providers with higher weights")
	fmt.Println("have more influence on the final result.")
}

// createMockProviders creates a set of mock providers with different behaviors
func createMockProviders() []provider.ProviderWeight {
	// Provider 1 - Returns consistently formatted answers
	consistentProvider := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			lowerPrompt := strings.ToLower(prompt)
			
			if strings.Contains(lowerPrompt, "capital of france") {
				return "Paris", nil
			} else if strings.Contains(lowerPrompt, "principles of object-oriented programming") {
				return "Encapsulation, Inheritance, Polymorphism", nil
			} else if strings.Contains(lowerPrompt, "earth") {
				return "The Earth is approximately 4.54 billion years old.", nil
			} else if strings.Contains(lowerPrompt, "programming language for beginners") {
				return "Python is the best programming language for beginners.", nil
			}
			
			return "This is a consistent mock response to: " + prompt, nil
		})
	
	// Provider 2 - Returns slightly different formatting/wording
	similarProvider := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			lowerPrompt := strings.ToLower(prompt)
			
			if strings.Contains(lowerPrompt, "capital of france") {
				return "The capital is Paris.", nil
			} else if strings.Contains(lowerPrompt, "principles of object-oriented programming") {
				return "Encapsulation, Inheritance, and Polymorphism", nil
			} else if strings.Contains(lowerPrompt, "earth") {
				return "Earth is about 4.5 billion years old.", nil
			} else if strings.Contains(lowerPrompt, "programming language for beginners") {
				return "For beginners, Python is ideal.", nil
			}
			
			return "This is a similar but differently worded response to: " + prompt, nil
		})
	
	// Provider 3 - Returns different answers to create some disagreement
	differentProvider := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			lowerPrompt := strings.ToLower(prompt)
			
			if strings.Contains(lowerPrompt, "capital of france") {
				return "Paris, France", nil
			} else if strings.Contains(lowerPrompt, "principles of object-oriented programming") {
				return "Inheritance, Polymorphism, Abstraction", nil
			} else if strings.Contains(lowerPrompt, "earth") {
				return "Scientists estimate the Earth to be around 4.6 billion years old.", nil
			} else if strings.Contains(lowerPrompt, "programming language for beginners") {
				return "JavaScript is best for beginners because it's used everywhere.", nil
			}
			
			return "This is a completely different response to: " + prompt, nil
		})
	
	// Provider 4 - Returns very different answers to test outlier rejection
	outlierProvider := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			lowerPrompt := strings.ToLower(prompt)
			
			if strings.Contains(lowerPrompt, "capital of france") {
				return "The capital city of the Republic of France is Paris, a global center for art, fashion, gastronomy and culture.", nil
			} else if strings.Contains(lowerPrompt, "principles of object-oriented programming") {
				return "Abstraction, Encapsulation, Inheritance, Polymorphism, and additionally, some include: Modularity, Reusability", nil
			} else if strings.Contains(lowerPrompt, "earth") {
				return "The current scientific consensus is that Earth formed approximately 4.54 billion years ago, with an uncertainty of about 1%.", nil
			} else if strings.Contains(lowerPrompt, "programming language for beginners") {
				return "Scratch is designed specifically for beginners with no coding experience.", nil
			}
			
			return "This is a very detailed and different response to: " + prompt, nil
		})
	
	// Return the providers with weights
	return []provider.ProviderWeight{
		{Provider: consistentProvider, Weight: 1.0, Name: "consistent"},
		{Provider: similarProvider, Weight: 1.0, Name: "similar"},
		{Provider: differentProvider, Weight: 1.0, Name: "different"},
		{Provider: outlierProvider, Weight: 0.5, Name: "outlier"}, // Lower weight for the outlier
	}
}