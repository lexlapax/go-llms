package benchmarks

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// BenchmarkResponseCreation benchmarks the creation of domain.Response objects
func BenchmarkResponseCreation(b *testing.B) {
	// Test with different content sizes to simulate different response sizes
	shortContent := "This is a short response."
	mediumContent := "This is a medium-length response that contains more words and characters to better simulate a real-world scenario with more text content."
	longContent := "This is a long response text that simulates a detailed answer from an LLM. " +
		"It contains multiple sentences and provides substantial information to better approximate the behavior " +
		"of a real LLM response. The purpose is to measure how the response creation performs with larger content " +
		"sizes, which is common in production scenarios. This response also includes various language features " +
		"like punctuation, spaces, and different sentence structures to make it more realistic for benchmarking purposes. " +
		"We want to ensure that our optimizations work well regardless of the size or complexity of the content being processed."

	// Benchmark with short content
	b.Run("ShortResponse_WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create response directly without using the pool
			_ = domain.Response{Content: shortContent}
		}
	})

	b.Run("ShortResponse_WithPool", func(b *testing.B) {
		pool := domain.GetResponsePool()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create response using the pool
			_ = pool.NewResponse(shortContent)
		}
	})

	// Benchmark with medium content
	b.Run("MediumResponse_WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create response directly without using the pool
			_ = domain.Response{Content: mediumContent}
		}
	})

	b.Run("MediumResponse_WithPool", func(b *testing.B) {
		pool := domain.GetResponsePool()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create response using the pool
			_ = pool.NewResponse(mediumContent)
		}
	})

	// Benchmark with long content
	b.Run("LongResponse_WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create response directly without using the pool
			_ = domain.Response{Content: longContent}
		}
	})

	b.Run("LongResponse_WithPool", func(b *testing.B) {
		pool := domain.GetResponsePool()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create response using the pool
			_ = pool.NewResponse(longContent)
		}
	})
}

// BenchmarkTokenCreation benchmarks the creation of domain.Token objects
func BenchmarkTokenCreation(b *testing.B) {
	// Test with different content sizes to simulate different token sizes
	shortToken := "a"
	mediumToken := "this is a medium token"
	longToken := "this is a much longer token that might be seen in some streaming responses from certain models"

	// Benchmark with short token
	b.Run("ShortToken_WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create token directly without using the pool
			_ = domain.Token{Text: shortToken, Finished: false}
		}
	})

	b.Run("ShortToken_WithPool", func(b *testing.B) {
		pool := domain.GetTokenPool()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create token using the pool
			_ = pool.NewToken(shortToken, false)
		}
	})

	// Benchmark with medium token
	b.Run("MediumToken_WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create token directly without using the pool
			_ = domain.Token{Text: mediumToken, Finished: false}
		}
	})

	b.Run("MediumToken_WithPool", func(b *testing.B) {
		pool := domain.GetTokenPool()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create token using the pool
			_ = pool.NewToken(mediumToken, false)
		}
	})

	// Benchmark with long token
	b.Run("LongToken_WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create token directly without using the pool
			_ = domain.Token{Text: longToken, Finished: false}
		}
	})

	b.Run("LongToken_WithPool", func(b *testing.B) {
		pool := domain.GetTokenPool()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create token using the pool
			_ = pool.NewToken(longToken, false)
		}
	})

	// Benchmark with final token (finished=true)
	b.Run("FinalToken_WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create final token directly without using the pool
			_ = domain.Token{Text: "", Finished: true}
		}
	})

	b.Run("FinalToken_WithPool", func(b *testing.B) {
		pool := domain.GetTokenPool()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create final token using the pool
			_ = pool.NewToken("", true)
		}
	})
}

// BenchmarkStreamingSimulation simulates a typical streaming scenario
func BenchmarkStreamingSimulation(b *testing.B) {
	// Create a set of tokens that represents a typical streaming response
	tokens := []string{
		"I", "'", "ll", " help", " you", " with", " that", " question", ".",
		" First", ",", " let", "'", "s", " understand", " the", " problem", ".",
		" The", " key", " concept", " here", " is", " to", " break", " it", " down",
		" into", " smaller", " steps", ".", " Let", "'", "s", " start", " by", "...",
	}

	// Simulate streaming without pool
	b.Run("StreamingSimulation_WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Use a dummy channel to simulate streaming
			ch := make(chan domain.Token, 5)
			go func() {
				defer close(ch)
				// Send all tokens except the last one
				for j := 0; j < len(tokens)-1; j++ {
					ch <- domain.Token{Text: tokens[j], Finished: false}
				}
				// Send the final token
				ch <- domain.Token{Text: tokens[len(tokens)-1], Finished: true}
			}()

			// Consume all tokens from the channel
			for range ch {
				// Just read the tokens, don't do anything with them
			}
		}
	})

	// Simulate streaming with pool
	b.Run("StreamingSimulation_WithPool", func(b *testing.B) {
		pool := domain.GetTokenPool()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Use a dummy channel to simulate streaming
			ch := make(chan domain.Token, 5)
			go func() {
				defer close(ch)
				// Send all tokens except the last one
				for j := 0; j < len(tokens)-1; j++ {
					ch <- pool.NewToken(tokens[j], false)
				}
				// Send the final token
				ch <- pool.NewToken(tokens[len(tokens)-1], true)
			}()

			// Consume all tokens from the channel
			for range ch {
				// Just read the tokens, don't do anything with them
			}
		}
	})
}