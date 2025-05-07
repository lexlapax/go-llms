package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// BenchmarkChannelPooling benchmarks the channel pooling for streaming operations
func BenchmarkChannelPooling(b *testing.B) {
	// Simulation parameters
	numTokens := 50
	simulationDuration := 100 * time.Millisecond

	// Benchmark without pooling
	b.Run("StreamingWithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := make(chan domain.Token, 20)
			go func() {
				defer close(ch)
				
				// Simulate streaming tokens at intervals
				for j := 0; j < numTokens; j++ {
					select {
					case ch <- domain.Token{Text: "token", Finished: j == numTokens-1}:
					case <-time.After(simulationDuration):
						return
					}
				}
			}()
			
			// Consume all tokens
			for token := range ch {
				_ = token // Just consume the token
			}
		}
	})
	
	// Benchmark with pooling
	b.Run("StreamingWithPool", func(b *testing.B) {
		pool := domain.GetChannelPool()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, ch := pool.GetResponseStream()
			go func() {
				defer close(ch)
				
				// Simulate streaming tokens at intervals
				for j := 0; j < numTokens; j++ {
					select {
					case ch <- domain.Token{Text: "token", Finished: j == numTokens-1}:
					case <-time.After(simulationDuration):
						return
					}
				}
			}()
			
			// Consume all tokens
			for token := range ch {
				_ = token // Just consume the token
			}
		}
	})
	
	// Benchmark high-throughput scenario with many streams
	const numStreams = 100
	
	b.Run("MultipleStreamsWithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), simulationDuration*3)
			
			// Create multiple streams
			streams := make([]chan domain.Token, numStreams)
			for j := 0; j < numStreams; j++ {
				streams[j] = make(chan domain.Token, 20)
				go func(ch chan domain.Token) {
					defer close(ch)
					
					// Stream fewer tokens in the high-throughput scenario
					for k := 0; k < 10; k++ {
						select {
						case <-ctx.Done():
							return
						case ch <- domain.Token{Text: "token", Finished: k == 9}:
						}
					}
				}(streams[j])
			}
			
			// Consume all streams
			for _, stream := range streams {
				for range stream {
					// Just consume
				}
			}
			
			cancel()
		}
	})
	
	b.Run("MultipleStreamsWithPool", func(b *testing.B) {
		pool := domain.GetChannelPool()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), simulationDuration*3)
			
			// Create multiple streams
			streams := make([]chan domain.Token, numStreams)
			for j := 0; j < numStreams; j++ {
				_, ch := pool.GetResponseStream()
				streams[j] = ch
				go func(ch chan domain.Token) {
					defer close(ch)
					
					// Stream fewer tokens in the high-throughput scenario
					for k := 0; k < 10; k++ {
						select {
						case <-ctx.Done():
							return
						case ch <- domain.Token{Text: "token", Finished: k == 9}:
						}
					}
				}(ch)
			}
			
			// Consume all streams
			for _, stream := range streams {
				for range stream {
					// Just consume
				}
			}
			
			cancel()
		}
	})
	
	// Benchmark realistic streaming scenario with token pools
	b.Run("RealisticStreamingWithoutPooling", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			ch := make(chan domain.Token, 20)
			
			go func() {
				defer close(ch)
				
				tokens := []string{
					"The", " quick", " brown", " fox", " jumps", " over", " the", " lazy", " dog", ".",
				}
				
				for j, text := range tokens {
					token := domain.Token{
						Text:     text,
						Finished: j == len(tokens)-1,
					}
					
					select {
					case ch <- token:
					case <-time.After(simulationDuration):
						return
					}
				}
			}()
			
			// Accumulate tokens
			var fullText string
			for token := range ch {
				fullText += token.Text
			}
		}
	})
	
	b.Run("RealisticStreamingWithPooling", func(b *testing.B) {
		channelPool := domain.GetChannelPool()
		tokenPool := domain.GetTokenPool()
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			_, ch := channelPool.GetResponseStream()
			
			go func() {
				defer close(ch)
				
				tokens := []string{
					"The", " quick", " brown", " fox", " jumps", " over", " the", " lazy", " dog", ".",
				}
				
				for j, text := range tokens {
					token := tokenPool.NewToken(text, j == len(tokens)-1)
					
					select {
					case ch <- token:
					case <-time.After(simulationDuration):
						return
					}
				}
			}()
			
			// Accumulate tokens
			var fullText string
			for token := range ch {
				fullText += token.Text
			}
		}
	})
}