package benchmarks

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// Test multi-provider with different consensus strategies
// This is a simple end-to-end benchmark of the MultiProvider with different strategies
// Helper function to create a mock multi-provider for testing
func createMockMultiProvider() *provider.MultiProvider {
	// Create 3 mock providers
	mp1 := provider.NewMockProvider().WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		if prompt == "What is the capital of France?" {
			return "The capital of France is Paris.", nil
		}
		return "Default response", nil
	})
	
	mp2 := provider.NewMockProvider().WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		if prompt == "What is the capital of France?" {
			return "Paris is the capital city of France.", nil
		}
		return "Default response", nil
	})
	
	mp3 := provider.NewMockProvider().WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		if prompt == "What is the capital of France?" {
			return "France's capital is Paris.", nil
		}
		return "Default response", nil
	})
	
	// Create a MultiProvider with all 3 providers
	return provider.NewMultiProvider([]provider.ProviderWeight{
		{Provider: mp1, Weight: 1.0, Name: "mock1"},
		{Provider: mp2, Weight: 1.0, Name: "mock2"},
		{Provider: mp3, Weight: 1.0, Name: "mock3"},
	}, provider.StrategyConsensus)
}

func BenchmarkMultiProviderWithConsensus(b *testing.B) {
	
	// Test with various strategies and multiple inputs
	testCases := []struct {
		name     string
		setupFn  func() *provider.MultiProvider
		inputs   []string  // Different inputs to test with
	}{
		{
			name: "Fastest",
			setupFn: func() *provider.MultiProvider {
				return createMockMultiProvider()
			},
			inputs: []string{
				"What is the capital of France?",
			},
		},
		{
			name: "Primary",
			setupFn: func() *provider.MultiProvider {
				return createMockMultiProvider().
					WithPrimaryProvider(0)
			},
			inputs: []string{
				"What is the capital of France?",
			},
		},
		{
			name: "Consensus_Majority",
			setupFn: func() *provider.MultiProvider {
				return createMockMultiProvider().
					WithConsensusStrategy(provider.ConsensusMajority)
			},
			inputs: []string{
				"What is the capital of France?",
			},
		},
		{
			name: "Consensus_Similarity",
			setupFn: func() *provider.MultiProvider {
				return createMockMultiProvider().
					WithConsensusStrategy(provider.ConsensusSimilarity).
					WithSimilarityThreshold(0.6)
			},
			inputs: []string{
				"What is the capital of France?",
			},
		},
		{
			name: "Consensus_Similarity_WithCache",
			setupFn: func() *provider.MultiProvider {
				// This run should benefit from the caching we implemented
				return createMockMultiProvider().
					WithConsensusStrategy(provider.ConsensusSimilarity).
					WithSimilarityThreshold(0.6)
			},
			inputs: []string{
				"What is the capital of France?", // Same input multiple times to test caching
				"What is the capital of France?",
				"What is the capital of France?",
			},
		},
		{
			name: "Consensus_Weighted",
			setupFn: func() *provider.MultiProvider {
				// Create individual providers with weights
				mp1 := provider.NewMockProvider().WithPredefinedResponses(map[string]string{
					"What is the capital of France?": "The capital of France is Paris.",
				})
				
				mp2 := provider.NewMockProvider().WithPredefinedResponses(map[string]string{
					"What is the capital of France?": "Paris is the capital city of France.",
				})
				
				mp3 := provider.NewMockProvider().WithPredefinedResponses(map[string]string{
					"What is the capital of France?": "France's capital is Paris.",
				})
				
				// Create weighted providers
				return provider.NewMultiProvider([]provider.ProviderWeight{
					{Provider: mp1, Weight: 1.0, Name: "mock1"},
					{Provider: mp2, Weight: 0.7, Name: "mock2"},
					{Provider: mp3, Weight: 0.5, Name: "mock3"},
				}, provider.StrategyConsensus).
					WithConsensusStrategy(provider.ConsensusWeighted)
			},
			inputs: []string{
				"What is the capital of France?",
			},
		},
		{
			name: "Consensus_Combined_Strategies",
			setupFn: func() *provider.MultiProvider {
				// Test a benchmark that exercises all consensus strategies sequentially
				return createMockMultiProvider().
					WithConsensusStrategy(provider.ConsensusMajority) // Will be changed during benchmark
			},
			inputs: []string{
				"What is the capital of France?",
			},
		},
	}
	
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			mp := tc.setupFn()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// For the combined test, cycle through all strategies
				if tc.name == "Consensus_Combined_Strategies" {
					// Change strategy based on iteration
					switch i % 3 {
					case 0:
						mp.WithConsensusStrategy(provider.ConsensusMajority)
					case 1:
						mp.WithConsensusStrategy(provider.ConsensusSimilarity)
					case 2:
						mp.WithConsensusStrategy(provider.ConsensusWeighted)
					}
				}
				
				// For tests with multiple inputs, rotate through them
				inputIdx := i % len(tc.inputs)
				input := tc.inputs[inputIdx]
				
				_, _ = mp.Generate(context.Background(), input)
			}
		})
	}
}