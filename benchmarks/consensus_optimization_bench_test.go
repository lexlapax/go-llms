package benchmarks

import (
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// BenchmarkConsensusOptimization tests the performance of different consensus algorithms
// under various conditions to identify optimization opportunities
func BenchmarkConsensusOptimization(b *testing.B) {
	// Create test data for different scenarios
	shortResponses := []provider.FallbackResult{
		{Provider: "provider1", Content: "The capital of France is Paris.", ElapsedTime: 100 * time.Millisecond, Weight: 1.0},
		{Provider: "provider2", Content: "Paris is the capital city of France.", ElapsedTime: 150 * time.Millisecond, Weight: 1.0},
		{Provider: "provider3", Content: "The capital of France is Paris.", ElapsedTime: 120 * time.Millisecond, Weight: 1.0},
		{Provider: "provider4", Content: "France's capital city is Paris.", ElapsedTime: 110 * time.Millisecond, Weight: 1.0},
		{Provider: "provider5", Content: "Berlin is the capital of Germany.", ElapsedTime: 90 * time.Millisecond, Weight: 0.5},
	}

	mediumResponses := []provider.FallbackResult{
		{
			Provider:    "provider1",
			Content:     "Object-oriented programming (OOP) is a programming paradigm based on the concept of objects, which can contain data and code. The main principles of OOP are encapsulation, inheritance, polymorphism, and abstraction.",
			ElapsedTime: 120 * time.Millisecond,
			Weight:      1.0,
		},
		{
			Provider:    "provider2",
			Content:     "The four main principles of object-oriented programming are encapsulation, inheritance, polymorphism, and abstraction. These concepts help organize code and make it more reusable.",
			ElapsedTime: 150 * time.Millisecond,
			Weight:      1.0,
		},
		{
			Provider:    "provider3",
			Content:     "Object-oriented programming has four key principles: encapsulation (hiding implementation details), inheritance (creating new classes from existing ones), polymorphism (using interfaces to handle different types), and abstraction (focusing on essential features).",
			ElapsedTime: 180 * time.Millisecond,
			Weight:      1.0,
		},
		{
			Provider:    "provider4",
			Content:     "The principles of functional programming include pure functions, immutability, function composition, and higher-order functions.",
			ElapsedTime: 130 * time.Millisecond,
			Weight:      0.7,
		},
	}

	longResponses := []provider.FallbackResult{
		{
			Provider: "provider1",
			Content: `Machine learning is a subset of artificial intelligence (AI) that provides systems the ability to automatically learn and improve from experience without being explicitly programmed. Machine learning focuses on the development of computer programs that can access data and use it to learn for themselves.

The process of learning begins with observations or data, such as examples, direct experience, or instruction, in order to look for patterns in data and make better decisions in the future based on the examples that we provide. The primary aim is to allow the computers to learn automatically without human intervention or assistance and adjust actions accordingly.

Some popular machine learning algorithms include linear regression, logistic regression, decision trees, random forests, support vector machines, and neural networks.`,
			ElapsedTime: 250 * time.Millisecond,
			Weight:      1.0,
		},
		{
			Provider: "provider2",
			Content: `Machine learning is a field of artificial intelligence that enables computer systems to learn from data and improve through experience, without explicit programming. It focuses on developing algorithms that can analyze data, recognize patterns, and make predictions or decisions.

The learning process typically begins with observations, data, or examples, which are used to identify patterns and relationships. The goal is to enable computers to learn autonomously and adapt their behavior based on the data they process.

Common machine learning techniques include supervised learning (using labeled training data), unsupervised learning (identifying patterns in unlabeled data), and reinforcement learning (learning through trial and error with rewards and penalties).`,
			ElapsedTime: 280 * time.Millisecond,
			Weight:      1.0,
		},
		{
			Provider: "provider3",
			Content: `Blockchain technology is a decentralized, distributed ledger that records transactions across multiple computers. The technology ensures that each transaction is verified by multiple participants and cannot be altered retroactively, providing a secure and transparent system for recording information.

Key features of blockchain include:
1. Decentralization: No single entity controls the blockchain
2. Transparency: All transactions are visible to all network participants
3. Immutability: Once recorded, data cannot be altered
4. Security: Cryptographic principles ensure data integrity

Blockchain has applications beyond cryptocurrencies, including supply chain management, voting systems, identity verification, and smart contracts.`,
			ElapsedTime: 300 * time.Millisecond,
			Weight:      1.0,
		},
	}

	conflictingResponses := []provider.FallbackResult{
		{Provider: "provider1", Content: "JavaScript is the best programming language for beginners.", ElapsedTime: 100 * time.Millisecond, Weight: 1.0},
		{Provider: "provider2", Content: "Python is the best programming language for beginners.", ElapsedTime: 120 * time.Millisecond, Weight: 1.0},
		{Provider: "provider3", Content: "Python is ideal for beginners due to its simple syntax.", ElapsedTime: 110 * time.Millisecond, Weight: 1.0},
		{Provider: "provider4", Content: "For beginners, I recommend starting with Python.", ElapsedTime: 130 * time.Millisecond, Weight: 0.8},
		{Provider: "provider5", Content: "JavaScript is recommended for beginners who want to learn web development.", ElapsedTime: 140 * time.Millisecond, Weight: 0.9},
	}

	weightedResponses := []provider.FallbackResult{
		{Provider: "provider1", Content: "Climate change is primarily caused by human activities.", ElapsedTime: 100 * time.Millisecond, Weight: 3.0},
		{Provider: "provider2", Content: "Climate change has both natural and human causes.", ElapsedTime: 120 * time.Millisecond, Weight: 1.0},
		{Provider: "provider3", Content: "Human activities are the main driver of climate change.", ElapsedTime: 110 * time.Millisecond, Weight: 1.5},
		{Provider: "provider4", Content: "Climate change is a natural cycle not influenced by humans.", ElapsedTime: 130 * time.Millisecond, Weight: 0.5},
		{Provider: "provider5", Content: "The primary cause of climate change is human emission of greenhouse gases.", ElapsedTime: 150 * time.Millisecond, Weight: 2.0},
	}

	// Test majority consensus with different response sets
	b.Run("MajorityConsensus_Short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectMajorityConsensus(shortResponses)
		}
	})

	b.Run("MajorityConsensus_Medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectMajorityConsensus(mediumResponses)
		}
	})

	b.Run("MajorityConsensus_Conflicting", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectMajorityConsensus(conflictingResponses)
		}
	})

	// Test similarity consensus with different response sets and thresholds
	thresholds := []float64{0.5, 0.7, 0.9}
	for _, threshold := range thresholds {
		b.Run("SimilarityConsensus_Short_"+fmt.Sprintf("%.1f", threshold), func(b *testing.B) {
			// Reset similarity cache for consistent benchmarking
			provider.ResetSimilarityCache()
			
			for i := 0; i < b.N; i++ {
				_, _ = provider.PublicSelectSimilarityConsensus(shortResponses, threshold)
			}
		})

		b.Run("SimilarityConsensus_Medium_"+fmt.Sprintf("%.1f", threshold), func(b *testing.B) {
			// Reset similarity cache for consistent benchmarking
			provider.ResetSimilarityCache()
			
			for i := 0; i < b.N; i++ {
				_, _ = provider.PublicSelectSimilarityConsensus(mediumResponses, threshold)
			}
		})

		b.Run("SimilarityConsensus_Long_"+fmt.Sprintf("%.1f", threshold), func(b *testing.B) {
			// Reset similarity cache for consistent benchmarking
			provider.ResetSimilarityCache()
			
			for i := 0; i < b.N; i++ {
				_, _ = provider.PublicSelectSimilarityConsensus(longResponses, threshold)
			}
		})
	}

	// Test similarity consensus with caching benefits
	b.Run("SimilarityConsensus_WithCache", func(b *testing.B) {
		// Reset cache initially
		provider.ResetSimilarityCache()
		
		// Pre-warm the cache with one call
		_, _ = provider.PublicSelectSimilarityConsensus(mediumResponses, 0.7)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectSimilarityConsensus(mediumResponses, 0.7)
		}
	})

	// Test weighted consensus with different response sets
	b.Run("WeightedConsensus_Equal", func(b *testing.B) {
		// Create equal weights
		equalWeightResponses := make([]provider.FallbackResult, len(conflictingResponses))
		copy(equalWeightResponses, conflictingResponses)
		for i := range equalWeightResponses {
			equalWeightResponses[i].Weight = 1.0
		}
		
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectWeightedConsensus(equalWeightResponses)
		}
	})

	b.Run("WeightedConsensus_Varied", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectWeightedConsensus(weightedResponses)
		}
	})

	// Test similarity calculation separately
	b.Run("SimilarityCalculation_Short", func(b *testing.B) {
		// Reset similarity cache
		provider.ResetSimilarityCache()
		
		text1 := shortResponses[0].Content
		text2 := shortResponses[1].Content
		
		for i := 0; i < b.N; i++ {
			_ = provider.PublicCalculateSimilarity(text1, text2)
		}
	})

	b.Run("SimilarityCalculation_Medium", func(b *testing.B) {
		// Reset similarity cache
		provider.ResetSimilarityCache()
		
		text1 := mediumResponses[0].Content
		text2 := mediumResponses[1].Content
		
		for i := 0; i < b.N; i++ {
			_ = provider.PublicCalculateSimilarity(text1, text2)
		}
	})

	b.Run("SimilarityCalculation_Long", func(b *testing.B) {
		// Reset similarity cache
		provider.ResetSimilarityCache()
		
		text1 := longResponses[0].Content
		text2 := longResponses[1].Content
		
		for i := 0; i < b.N; i++ {
			_ = provider.PublicCalculateSimilarity(text1, text2)
		}
	})

	// Test cached similarity calculation
	b.Run("SimilarityCalculation_Cached", func(b *testing.B) {
		// Reset cache initially
		provider.ResetSimilarityCache()
		
		text1 := mediumResponses[0].Content
		text2 := mediumResponses[1].Content
		
		// Pre-warm the cache
		_ = provider.PublicCalculateSimilarity(text1, text2)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = provider.PublicCalculateSimilarity(text1, text2)
		}
	})

	// Test SelectBestFromGroup function
	b.Run("SelectBestFromGroup", func(b *testing.B) {
		// Create a group of similar responses
		group := []provider.FallbackResult{
			shortResponses[0],
			shortResponses[1],
			shortResponses[2],
		}
		
		for i := 0; i < b.N; i++ {
			_ = provider.PublicSelectBestFromGroup(group)
		}
	})
}

// BenchmarkStructuredConsensus tests the performance of consensus for structured outputs
func BenchmarkStructuredConsensus(b *testing.B) {
	// Create test structured responses (simulating JSON marshaled data)
	simpleStructured := []provider.FallbackResult{
		{
			Provider: "provider1",
			Content:  `{"name":"John","age":30,"city":"New York"}`,
			Weight:   1.0,
		},
		{
			Provider: "provider2",
			Content:  `{"name":"John","age":30,"city":"New York"}`,
			Weight:   1.0,
		},
		{
			Provider: "provider3",
			Content:  `{"name":"John","age":30,"city":"NYC"}`,
			Weight:   1.0,
		},
	}

	complexStructured := []provider.FallbackResult{
		{
			Provider: "provider1",
			Content: `{
				"person": {
					"name": "John Smith",
					"age": 30,
					"address": {
						"street": "123 Main St",
						"city": "New York",
						"zip": "10001"
					},
					"hobbies": ["reading", "hiking", "photography"]
				}
			}`,
			Weight: 1.0,
		},
		{
			Provider: "provider2",
			Content: `{
				"person": {
					"name": "John Smith",
					"age": 30,
					"address": {
						"street": "123 Main Street",
						"city": "New York",
						"zip": "10001"
					},
					"hobbies": ["reading", "hiking", "photography"]
				}
			}`,
			Weight: 1.0,
		},
		{
			Provider: "provider3",
			Content: `{
				"person": {
					"name": "Jane Doe",
					"age": 28,
					"address": {
						"street": "456 Park Ave",
						"city": "Chicago",
						"zip": "60601"
					},
					"hobbies": ["painting", "swimming", "cooking"]
				}
			}`,
			Weight: 0.7,
		},
	}

	// Benchmark simple structured consensus
	b.Run("SimpleStructured_Majority", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectMajorityConsensus(simpleStructured)
		}
	})

	b.Run("SimpleStructured_Similarity", func(b *testing.B) {
		provider.ResetSimilarityCache()
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectSimilarityConsensus(simpleStructured, 0.7)
		}
	})

	b.Run("SimpleStructured_Weighted", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectWeightedConsensus(simpleStructured)
		}
	})

	// Benchmark complex structured consensus
	b.Run("ComplexStructured_Majority", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectMajorityConsensus(complexStructured)
		}
	})

	b.Run("ComplexStructured_Similarity", func(b *testing.B) {
		provider.ResetSimilarityCache()
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectSimilarityConsensus(complexStructured, 0.7)
		}
	})

	b.Run("ComplexStructured_Weighted", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = provider.PublicSelectWeightedConsensus(complexStructured)
		}
	})

	// Test similarity calculation for JSON content
	b.Run("JSONSimilarity", func(b *testing.B) {
		provider.ResetSimilarityCache()
		json1 := complexStructured[0].Content
		json2 := complexStructured[1].Content
		
		for i := 0; i < b.N; i++ {
			_ = provider.PublicCalculateSimilarity(json1, json2)
		}
	})
}

// BenchmarkConcurrentConsensus tests the performance of consensus algorithms under concurrent access
func BenchmarkConcurrentConsensus(b *testing.B) {
	// Create test responses
	responses := []provider.FallbackResult{
		{Provider: "provider1", Content: "The capital of France is Paris.", Weight: 1.0},
		{Provider: "provider2", Content: "Paris is the capital city of France.", Weight: 1.0},
		{Provider: "provider3", Content: "The capital of France is Paris.", Weight: 1.0},
		{Provider: "provider4", Content: "France's capital city is Paris.", Weight: 1.0},
		{Provider: "provider5", Content: "Berlin is the capital of Germany.", Weight: 0.5},
	}

	// Test concurrent access to consensus algorithms
	// Test similarity consensus in parallel
	b.Run("ConcurrentSimilarityConsensus", func(b *testing.B) {
		provider.ResetSimilarityCache()
		
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = provider.PublicSelectSimilarityConsensus(responses, 0.7)
			}
		})
	})
	
	// Test majority consensus in parallel
	b.Run("ConcurrentMajorityConsensus", func(b *testing.B) {
		provider.ResetSimilarityCache()
		
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = provider.PublicSelectMajorityConsensus(responses)
			}
		})
	})
	
	// Test weighted consensus in parallel
	b.Run("ConcurrentWeightedConsensus", func(b *testing.B) {
		provider.ResetSimilarityCache()
		
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = provider.PublicSelectWeightedConsensus(responses)
			}
		})
	})

	// Test concurrent access to similarity calculation
	b.Run("ConcurrentSimilarityCalculation", func(b *testing.B) {
		provider.ResetSimilarityCache()
		
		text1 := responses[0].Content
		text2 := responses[1].Content
		
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = provider.PublicCalculateSimilarity(text1, text2)
			}
		})
	})
}