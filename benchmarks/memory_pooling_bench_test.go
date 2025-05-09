package benchmarks

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

// BenchmarkMemoryPooling provides comprehensive benchmarks for all memory pool implementations
func BenchmarkMemoryPooling(b *testing.B) {
	// Different content sizes for response/token testing
	var (
		smallContent  = "Short text"
		mediumContent = "This is a medium length text that has a reasonable number of characters"
		largeContent  = "This is a much larger text that simulates a comprehensive response from an LLM model. " +
			"It contains multiple sentences and paragraphs to better approximate real-world scenarios where " +
			"responses might be quite extensive. When dealing with larger texts, memory allocations become " +
			"more significant, making the benefits of pooling more apparent. This benchmark helps us evaluate " +
			"how well our pooling strategy performs with larger memory allocations and whether it provides " +
			"meaningful benefits in such scenarios compared to standard allocation patterns."
	)

	// 1. Benchmark creating and getting from the pool (pool instantiation)
	b.Run("PoolInstantiation", func(b *testing.B) {
		b.Run("ResponsePool_GetGlobal", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.GetResponsePool()
			}
		})

		b.Run("TokenPool_GetGlobal", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.GetTokenPool()
			}
		})

		b.Run("ChannelPool_GetGlobal", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.GetChannelPool()
			}
		})

		b.Run("ResponsePool_NewPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.NewResponsePool()
			}
		})

		b.Run("TokenPool_NewPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.NewTokenPool()
			}
		})

		b.Run("ChannelPool_NewPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.NewChannelPool()
			}
		})
	})

	// 2. Benchmark getting and putting objects back to the pool (basic usage)
	b.Run("BasicPoolOperations", func(b *testing.B) {
		b.Run("ResponsePool_Get_Put", func(b *testing.B) {
			pool := domain.GetResponsePool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				resp := pool.Get()
				pool.Put(resp)
			}
		})

		b.Run("TokenPool_Get_Put", func(b *testing.B) {
			pool := domain.GetTokenPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				token := pool.Get()
				pool.Put(token)
			}
		})

		b.Run("ChannelPool_Get_Put", func(b *testing.B) {
			pool := domain.GetChannelPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ch := pool.Get()
				pool.Put(ch)
			}
		})
	})

	// 3. Benchmark the high-level NewX operations for different content sizes
	b.Run("ObjectCreation", func(b *testing.B) {
		// Response creation
		b.Run("Response_Small_WithoutPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.Response{Content: smallContent}
			}
		})

		b.Run("Response_Small_WithPool", func(b *testing.B) {
			pool := domain.GetResponsePool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = pool.NewResponse(smallContent)
			}
		})

		b.Run("Response_Medium_WithoutPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.Response{Content: mediumContent}
			}
		})

		b.Run("Response_Medium_WithPool", func(b *testing.B) {
			pool := domain.GetResponsePool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = pool.NewResponse(mediumContent)
			}
		})

		b.Run("Response_Large_WithoutPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.Response{Content: largeContent}
			}
		})

		b.Run("Response_Large_WithPool", func(b *testing.B) {
			pool := domain.GetResponsePool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = pool.NewResponse(largeContent)
			}
		})

		// Token creation
		b.Run("Token_Small_WithoutPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.Token{Text: smallContent, Finished: false}
			}
		})

		b.Run("Token_Small_WithPool", func(b *testing.B) {
			pool := domain.GetTokenPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = pool.NewToken(smallContent, false)
			}
		})

		b.Run("Token_Medium_WithoutPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.Token{Text: mediumContent, Finished: false}
			}
		})

		b.Run("Token_Medium_WithPool", func(b *testing.B) {
			pool := domain.GetTokenPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = pool.NewToken(mediumContent, false)
			}
		})

		b.Run("Token_Large_WithoutPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = domain.Token{Text: largeContent, Finished: false}
			}
		})

		b.Run("Token_Large_WithPool", func(b *testing.B) {
			pool := domain.GetTokenPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = pool.NewToken(largeContent, false)
			}
		})
	})

	// 4. Benchmark response streaming with and without pools
	b.Run("ResponseStreaming", func(b *testing.B) {
		// Arrays of tokens for streaming simulation
		smallTokens := []string{"This", " is", " a", " short", " message", "."}
		mediumTokens := []string{
			"This", " is", " a", " medium", " length", " message", " that", " simulates",
			" a", " conversation", " with", " an", " LLM", " model", ".",
		}
		largeTokens := []string{
			"This", " is", " the", " beginning", " of", " a", " much", " longer", " response", ".",
			" It", " contains", " multiple", " sentences", " to", " better", " approximate",
			" a", " realistic", " streaming", " scenario", " from", " an", " LLM", " model", ".",
			" When", " models", " stream", " tokens", ",", " they", " often", " break", " text",
			" into", " subword", " units", " rather", " than", " complete", " words", ".",
		}

		// Small streaming
		b.Run("Stream_Small_WithoutPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ch := make(chan domain.Token, len(smallTokens))
				go func() {
					defer close(ch)
					for j, token := range smallTokens {
						ch <- domain.Token{
							Text:     token,
							Finished: j == len(smallTokens)-1,
						}
					}
				}()

				// Consume all tokens
				for range ch {
					// Just consume
				}
			}
		})

		b.Run("Stream_Small_WithPool", func(b *testing.B) {
			tokenPool := domain.GetTokenPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ch := make(chan domain.Token, len(smallTokens))
				go func() {
					defer close(ch)
					for j, token := range smallTokens {
						ch <- tokenPool.NewToken(token, j == len(smallTokens)-1)
					}
				}()

				// Consume all tokens
				for range ch {
					// Just consume
				}
			}
		})

		b.Run("Stream_Small_FullPool", func(b *testing.B) {
			tokenPool := domain.GetTokenPool()
			channelPool := domain.GetChannelPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, ch := channelPool.GetResponseStream()
				go func() {
					defer close(ch)
					for j, token := range smallTokens {
						ch <- tokenPool.NewToken(token, j == len(smallTokens)-1)
					}
				}()

				// Consume all tokens
				for range ch {
					// Just consume
				}
			}
		})

		// Medium streaming
		b.Run("Stream_Medium_WithoutPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ch := make(chan domain.Token, len(mediumTokens))
				go func() {
					defer close(ch)
					for j, token := range mediumTokens {
						ch <- domain.Token{
							Text:     token,
							Finished: j == len(mediumTokens)-1,
						}
					}
				}()

				// Consume all tokens
				for range ch {
					// Just consume
				}
			}
		})

		b.Run("Stream_Medium_WithPool", func(b *testing.B) {
			tokenPool := domain.GetTokenPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ch := make(chan domain.Token, len(mediumTokens))
				go func() {
					defer close(ch)
					for j, token := range mediumTokens {
						ch <- tokenPool.NewToken(token, j == len(mediumTokens)-1)
					}
				}()

				// Consume all tokens
				for range ch {
					// Just consume
				}
			}
		})

		b.Run("Stream_Medium_FullPool", func(b *testing.B) {
			tokenPool := domain.GetTokenPool()
			channelPool := domain.GetChannelPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, ch := channelPool.GetResponseStream()
				go func() {
					defer close(ch)
					for j, token := range mediumTokens {
						ch <- tokenPool.NewToken(token, j == len(mediumTokens)-1)
					}
				}()

				// Consume all tokens
				for range ch {
					// Just consume
				}
			}
		})

		// Large streaming
		b.Run("Stream_Large_WithoutPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ch := make(chan domain.Token, len(largeTokens))
				go func() {
					defer close(ch)
					for j, token := range largeTokens {
						ch <- domain.Token{
							Text:     token,
							Finished: j == len(largeTokens)-1,
						}
					}
				}()

				// Consume all tokens
				for range ch {
					// Just consume
				}
			}
		})

		b.Run("Stream_Large_WithPool", func(b *testing.B) {
			tokenPool := domain.GetTokenPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ch := make(chan domain.Token, len(largeTokens))
				go func() {
					defer close(ch)
					for j, token := range largeTokens {
						ch <- tokenPool.NewToken(token, j == len(largeTokens)-1)
					}
				}()

				// Consume all tokens
				for range ch {
					// Just consume
				}
			}
		})

		b.Run("Stream_Large_FullPool", func(b *testing.B) {
			tokenPool := domain.GetTokenPool()
			channelPool := domain.GetChannelPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, ch := channelPool.GetResponseStream()
				go func() {
					defer close(ch)
					for j, token := range largeTokens {
						ch <- tokenPool.NewToken(token, j == len(largeTokens)-1)
					}
				}()

				// Consume all tokens
				for range ch {
					// Just consume
				}
			}
		})
	})

	// 5. Benchmark concurrent access to pools
	b.Run("ConcurrentPoolAccess", func(b *testing.B) {
		// Response pool concurrent access
		b.Run("ResponsePool_Concurrent", func(b *testing.B) {
			pool := domain.GetResponsePool()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					resp := pool.Get()
					resp.Content = "Test content"
					pool.Put(resp)
				}
			})
		})

		// Token pool concurrent access
		b.Run("TokenPool_Concurrent", func(b *testing.B) {
			pool := domain.GetTokenPool()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					token := pool.Get()
					token.Text = "Test token"
					token.Finished = false
					pool.Put(token)
				}
			})
		})

		// Channel pool concurrent access
		b.Run("ChannelPool_Concurrent", func(b *testing.B) {
			pool := domain.GetChannelPool()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					ch := pool.Get()
					// Send and receive a token to ensure the channel works
					go func() {
						ch <- domain.Token{Text: "Test", Finished: true}
						close(ch)
					}()

					// Consume the token
					for range ch {
						// Just consume
					}

					// Now put the channel back
					pool.Put(ch)
				}
			})
		})
	})

	// 6. Benchmark high-throughput multi-stream scenarios
	b.Run("HighThroughputStreaming", func(b *testing.B) {
		const numStreams = 50 // Number of concurrent streams
		tokens := []string{"The", " quick", " brown", " fox", " jumps"}

		b.Run("MultiStream_WithoutPool", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				wg.Add(numStreams)

				// Create channels
				channels := make([]chan domain.Token, numStreams)
				for j := 0; j < numStreams; j++ {
					channels[j] = make(chan domain.Token, 10)
				}

				// Start producers
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()

				for j := 0; j < numStreams; j++ {
					go func(ch chan domain.Token, id int) {
						defer close(ch)
						defer wg.Done()

						for k, token := range tokens {
							select {
							case <-ctx.Done():
								return
							case ch <- domain.Token{
								Text:     token + " " + string(rune('A'+id%26)),
								Finished: k == len(tokens)-1,
							}:
								// Token sent successfully
							}
						}
					}(channels[j], j)
				}

				// Consume all channels
				for j := 0; j < numStreams; j++ {
					for range channels[j] {
						// Just consume
					}
				}

				wg.Wait()
			}
		})

		b.Run("MultiStream_WithPool", func(b *testing.B) {
			tokenPool := domain.GetTokenPool()
			channelPool := domain.GetChannelPool()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				wg.Add(numStreams)

				// Create channels
				channels := make([]chan domain.Token, numStreams)
				for j := 0; j < numStreams; j++ {
					_, ch := channelPool.GetResponseStream()
					channels[j] = ch
				}

				// Start producers
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()

				for j := 0; j < numStreams; j++ {
					go func(ch chan domain.Token, id int) {
						defer close(ch)
						defer wg.Done()

						for k, token := range tokens {
							select {
							case <-ctx.Done():
								return
							case ch <- tokenPool.NewToken(
								token+" "+string(rune('A'+id%26)),
								k == len(tokens)-1,
							):
								// Token sent successfully
							}
						}
					}(channels[j], j)
				}

				// Consume all channels
				for j := 0; j < numStreams; j++ {
					for range channels[j] {
						// Just consume
					}
				}

				wg.Wait()
			}
		})
	})

	// 7. Benchmark clear operations for checking overhead
	b.Run("ClearOperations", func(b *testing.B) {
		b.Run("Response_Clear", func(b *testing.B) {
			pool := domain.GetResponsePool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				resp := pool.Get()
				resp.Content = largeContent
				// Clear the response (what happens in Put)
				resp.Content = ""
				pool.Put(resp)
			}
		})

		b.Run("Token_Clear", func(b *testing.B) {
			pool := domain.GetTokenPool()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				token := pool.Get()
				token.Text = largeContent
				token.Finished = true
				// Clear the token (what happens in Put)
				token.Text = ""
				token.Finished = false
				pool.Put(token)
			}
		})
	})
}
