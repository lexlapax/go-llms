package stress

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// TestProviderConcurrentRequests tests provider stability under high concurrency
func TestProviderConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Track memory stats before and after test
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Create a variety of providers to test
	openaiProvider := provider.NewMockProvider()
	// Set up predictable responses for the mock
	openaiProvider.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		return "Response from OpenAI", nil
	})

	anthropicProvider := provider.NewMockProvider()
	// Set up predictable responses for the mock
	anthropicProvider.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		return "Response from Anthropic", nil
	})

	geminiProvider := provider.NewMockProvider()
	// Set up predictable responses for the mock
	geminiProvider.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
		return "Response from Gemini", nil
	})

	// Create a multi-provider with consensus strategy
	multiProvider := provider.NewMultiProvider(
		[]provider.ProviderWeight{
			{Provider: openaiProvider, Weight: 1.0},
			{Provider: anthropicProvider, Weight: 1.0},
			{Provider: geminiProvider, Weight: 1.0},
		},
		provider.StrategyConsensus,
	)
	// Set consensus strategy
	multiProvider.WithConsensusStrategy(provider.ConsensusMajority)

	// Provider configs for testing
	providers := []struct {
		name     string
		provider domain.Provider
	}{
		{"OpenAI", openaiProvider},
		{"Anthropic", anthropicProvider},
		{"Gemini", geminiProvider},
		{"Multi", multiProvider},
	}

	// Define concurrency levels to test
	concurrencyLevels := []int{10, 50, 100, 250, 500}

	// Define various prompts to simulate real-world variety
	prompts := []string{
		"Tell me about artificial intelligence",
		"What are the benefits of quantum computing?",
		"Explain the theory of relativity",
		"How do neural networks work?",
		"Describe the process of photosynthesis",
		"What are the major programming paradigms?",
		"Explain how blockchain technology works",
		"What is the impact of climate change?",
		"How does the human immune system function?",
		"What are the principles of object-oriented programming?",
	}

	// Run tests for each provider and concurrency level
	for _, p := range providers {
		for _, concurrency := range concurrencyLevels {
			t.Run(fmt.Sprintf("%s_Concurrency_%d", p.name, concurrency), func(t *testing.T) {
				var (
					wg            sync.WaitGroup
					successful    int32
					failed        int32
					totalLatencyMs int64
					maxLatencyMs   int64
					minLatencyMs   int64 = 999999
				)

				// Set a reasonable timeout
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cancel()

				// Create a semaphore to limit concurrent goroutines
				// Using a buffered channel as a semaphore
				sem := make(chan struct{}, concurrency)

				// Track goroutine count
				initialGoroutines := runtime.NumGoroutine()

				// Launch concurrent requests
				startTime := time.Now()
				for i := 0; i < concurrency*2; i++ {
					wg.Add(1)
					sem <- struct{}{} // Acquire semaphore
					go func(id int) {
						defer func() {
							<-sem // Release semaphore
							wg.Done()
						}()

						// Select a prompt randomly
						prompt := prompts[rand.Intn(len(prompts))]

						// Measure request time
						requestStart := time.Now()
						_, err := p.provider.Generate(ctx, prompt)
						requestDuration := time.Since(requestStart)
						latencyMs := requestDuration.Milliseconds()

						// Update metrics atomically
						atomic.AddInt64(&totalLatencyMs, latencyMs)
						
						// Update min/max latency
						for {
							current := atomic.LoadInt64(&maxLatencyMs)
							if latencyMs <= current {
								break
							}
							if atomic.CompareAndSwapInt64(&maxLatencyMs, current, latencyMs) {
								break
							}
						}
						
						for {
							current := atomic.LoadInt64(&minLatencyMs)
							if latencyMs >= current {
								break
							}
							if atomic.CompareAndSwapInt64(&minLatencyMs, current, latencyMs) {
								break
							}
						}

						if err != nil {
							atomic.AddInt32(&failed, 1)
							if ctx.Err() == context.DeadlineExceeded {
								t.Logf("Request %d timed out: %v", id, err)
							} else {
								t.Logf("Request %d failed: %v", id, err)
							}
						} else {
							atomic.AddInt32(&successful, 1)
						}
					}(i)
				}

				// Wait for all requests to complete
				wg.Wait()
				totalDuration := time.Since(startTime)

				// Check goroutine count
				peakGoroutines := runtime.NumGoroutine()

				// Record results
				successRate := float64(successful) / float64(successful+failed) * 100
				avgLatencyMs := float64(totalLatencyMs) / float64(successful+failed)

				t.Logf("Results for %s at concurrency %d:", p.name, concurrency)
				t.Logf("  Success rate: %.2f%% (%d/%d)", successRate, successful, successful+failed)
				t.Logf("  Average latency: %.2f ms", avgLatencyMs)
				t.Logf("  Min latency: %d ms", minLatencyMs)
				t.Logf("  Max latency: %d ms", maxLatencyMs)
				t.Logf("  Total duration: %v", totalDuration)
				t.Logf("  Goroutines: %d initial, %d peak", initialGoroutines, peakGoroutines)

				// Basic validation
				if successful == 0 {
					t.Errorf("No successful requests for %s at concurrency %d", p.name, concurrency)
				}
			})
		}
	}

	// Collect final memory stats
	runtime.ReadMemStats(&memStatsAfter)

	// Report memory usage
	t.Logf("Memory usage before: %.2f MB", float64(memStatsBefore.Alloc)/1024/1024)
	t.Logf("Memory usage after: %.2f MB", float64(memStatsAfter.Alloc)/1024/1024)

	// Calculate memory difference safely
	var memDiff float64
	if memStatsAfter.Alloc >= memStatsBefore.Alloc {
		memDiff = float64(memStatsAfter.Alloc-memStatsBefore.Alloc) / 1024 / 1024
	} else {
		// If we get a negative difference (due to GC), report 0
		memDiff = 0
	}
	t.Logf("Memory difference: %.2f MB", memDiff)
	t.Logf("Total allocations: %d objects", memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}