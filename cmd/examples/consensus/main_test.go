package main

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

func TestBasicConsensus(t *testing.T) {
	// Create mock providers with predictable responses
	provider1 := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return "Paris", nil
		})
	
	provider2 := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return "Paris", nil
		})
	
	provider3 := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return "Paris, France", nil
		})
	
	providers := []provider.ProviderWeight{
		{Provider: provider1, Weight: 1.0, Name: "provider1"},
		{Provider: provider2, Weight: 1.0, Name: "provider2"},
		{Provider: provider3, Weight: 1.0, Name: "provider3"},
	}
	
	// Create consensus provider
	consensusProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus)
	
	// Test basic consensus
	response, err := consensusProvider.Generate(
		context.Background(),
		"What is the capital of France?",
	)
	
	if err != nil {
		t.Fatalf("Error in consensus generation: %v", err)
	}
	
	// Since two providers return "Paris", that should be the consensus
	if response != "Paris" {
		t.Errorf("Expected consensus response 'Paris', got: %s", response)
	}
}

func TestConsensusSimilarity(t *testing.T) {
	// Create mock providers with similar but not identical responses
	provider1 := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return "The Earth is approximately 4.54 billion years old.", nil
		})
	
	provider2 := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return "The Earth is about 4.5 billion years old.", nil
		})
	
	provider3 := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return "The Earth is 6,000 years old.", nil
		})
	
	providers := []provider.ProviderWeight{
		{Provider: provider1, Weight: 1.0, Name: "provider1"},
		{Provider: provider2, Weight: 1.0, Name: "provider2"},
		{Provider: provider3, Weight: 1.0, Name: "provider3"},
	}
	
	// Create consensus provider with similarity strategy
	consensusProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus).
		WithConsensusStrategy(provider.ConsensusSimilarity).
		WithSimilarityThreshold(0.7) // 70% similarity threshold
	
	// Test similarity consensus
	response, err := consensusProvider.Generate(
		context.Background(),
		"How old is the Earth?",
	)
	
	if err != nil {
		t.Fatalf("Error in similarity consensus generation: %v", err)
	}
	
	// First two responses are similar, so one of them should be returned
	// The third response is very different and should be considered an outlier
	if response != "The Earth is approximately 4.54 billion years old." && 
	   response != "The Earth is about 4.5 billion years old." {
		t.Errorf("Expected consensus from similar responses about Earth's age, got: %s", response)
	}
}

func TestWeightedConsensus(t *testing.T) {
	// Create mock providers with different opinions
	provider1 := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return "Python is the best for beginners.", nil
		})
	
	provider2 := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return "JavaScript is best for beginners.", nil
		})
	
	provider3 := provider.NewMockProvider().WithGenerateFunc(
		func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
			return "JavaScript is best for beginners.", nil
		})
	
	// Case 1: Equal weights - consensus should be JavaScript
	providers1 := []provider.ProviderWeight{
		{Provider: provider1, Weight: 1.0, Name: "provider1"},
		{Provider: provider2, Weight: 1.0, Name: "provider2"},
		{Provider: provider3, Weight: 1.0, Name: "provider3"},
	}
	
	// Create consensus provider with weighted strategy
	equalWeightProvider := provider.NewMultiProvider(providers1, provider.StrategyConsensus).
		WithConsensusStrategy(provider.ConsensusWeighted)
	
	// Test equal weight consensus
	response1, err := equalWeightProvider.Generate(
		context.Background(),
		"What's the best programming language for beginners?",
	)
	
	if err != nil {
		t.Fatalf("Error in equal weight consensus generation: %v", err)
	}
	
	// With equal weights, majority should win (JavaScript)
	if response1 != "JavaScript is best for beginners." {
		t.Errorf("Expected consensus 'JavaScript is best for beginners.', got: %s", response1)
	}
	
	// Case 2: Higher weight for Python provider
	providers2 := []provider.ProviderWeight{
		{Provider: provider1, Weight: 3.0, Name: "provider1-high-weight"}, // Triple weight
		{Provider: provider2, Weight: 1.0, Name: "provider2"},
		{Provider: provider3, Weight: 1.0, Name: "provider3"},
	}
	
	// Create consensus provider with weighted strategy
	weightedProvider := provider.NewMultiProvider(providers2, provider.StrategyConsensus).
		WithConsensusStrategy(provider.ConsensusWeighted)
	
	// Test weighted consensus
	response2, err := weightedProvider.Generate(
		context.Background(),
		"What's the best programming language for beginners?",
	)
	
	if err != nil {
		t.Fatalf("Error in weighted consensus generation: %v", err)
	}
	
	// With higher weight for Python provider, it should win
	if response2 != "Python is the best for beginners." {
		t.Errorf("Expected weighted consensus 'Python is the best for beginners.', got: %s", response2)
	}
}