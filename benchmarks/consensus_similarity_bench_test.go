package benchmarks

import (
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// BenchmarkSimilarityCalculation tests the performance of calculateSimilarity function
func BenchmarkSimilarityCalculation(b *testing.B) {
	// Test pairs with various characteristics
	testPairs := []struct {
		name  string
		textA string
		textB string
	}{
		{
			name:  "Identical",
			textA: "The capital of France is Paris.",
			textB: "The capital of France is Paris.",
		},
		{
			name:  "Similar",
			textA: "The capital of France is Paris.",
			textB: "Paris is the capital city of France.",
		},
		{
			name:  "Different",
			textA: "The capital of France is Paris.",
			textB: "Berlin is the capital of Germany.",
		},
		{
			name:  "Long_Similar",
			textA: "The Large Language Model (LLM) is a type of artificial intelligence algorithm that uses deep learning techniques and massive datasets of text to generate human-like text. These models can engage in conversations, answer questions, summarize documents, translate languages, and create various forms of content.",
			textB: "A Large Language Model is an AI system trained on vast text datasets using deep learning methods to generate text that resembles human writing. LLMs can answer questions, have conversations, summarize texts, translate between languages, and create different types of content.",
		},
		{
			name:  "Long_Different",
			textA: "The Large Language Model (LLM) is a type of artificial intelligence algorithm that uses deep learning techniques and massive datasets of text to generate human-like text. These models can engage in conversations, answer questions, summarize documents, translate languages, and create various forms of content.",
			textB: "Quantum computing is a rapidly-emerging technology that uses the principles of quantum mechanics to perform computations. Unlike classical computers that use bits to represent either 0 or 1, quantum computers use quantum bits or qubits that can represent both 0 and 1 simultaneously due to superposition.",
		},
	}

	b.Run("First_Call", func(b *testing.B) {
		// Ensure a clean cache state for accurate first-call measurement
		provider.ResetSimilarityCache()
		
		// Just run once to measure first-call overhead
		pair := testPairs[1] // Use the Similar pair
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			provider.PublicCalculateSimilarity(pair.textA, pair.textB)
		}
	})

	// Run individual benchmarks for each test pair
	for _, pair := range testPairs {
		b.Run(pair.name, func(b *testing.B) {
			// Reset cache before each test
			provider.ResetSimilarityCache()
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				provider.PublicCalculateSimilarity(pair.textA, pair.textB)
			}
		})
		
		// Test with cache benefits
		b.Run(pair.name+"_Cached", func(b *testing.B) {
			// Prime the cache with one call
			provider.PublicCalculateSimilarity(pair.textA, pair.textB)
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				provider.PublicCalculateSimilarity(pair.textA, pair.textB)
			}
		})
	}
	
	// Test cache performance under concurrent access
	b.Run("Concurrent_Access", func(b *testing.B) {
		provider.ResetSimilarityCache()
		
		b.RunParallel(func(pb *testing.PB) {
			// Choose a random test pair for each goroutine
			// Simple timer-based "random" selection
			idx := time.Now().UnixNano() % int64(len(testPairs))
			pair := testPairs[idx]
			
			for pb.Next() {
				provider.PublicCalculateSimilarity(pair.textA, pair.textB)
			}
		})
	})
}