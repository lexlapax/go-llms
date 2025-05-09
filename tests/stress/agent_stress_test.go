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

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/testutils"
)

// TestAgentConcurrentRequests tests agent workflow stability under high concurrency
func TestAgentConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Track memory stats before and after test
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Define a set of mock tools for the agent to use
	mockTools := []domain.Tool{
		testutils.CreateCalculatorTool(),
		testutils.CreateMockTool("weather", "Gets weather information", nil),
		testutils.CreateMockTool("search", "Searches for information", nil),
	}

	// Create a base mock provider for all agents
	baseProvider := provider.NewMockProvider()

	// Create a variety of agent configurations to test
	baseAgent := workflow.NewAgent(baseProvider)
	for _, tool := range mockTools {
		baseAgent.AddTool(tool)
	}

	// Add a shared thread-safe tool counter for each agent type
	baseAgentToolCounter := &safeToolCounter{}
	baseAgent.WithHook(baseAgentToolCounter)

	cachedAgent := workflow.NewCachedAgent(baseProvider)
	for _, tool := range mockTools {
		cachedAgent.AddTool(tool)
	}

	// Add a shared thread-safe tool counter for the cached agent
	cachedAgentToolCounter := &safeToolCounter{}
	cachedAgent.WithHook(cachedAgentToolCounter)

	// Create providers for the multi-agent
	mockProvider1 := provider.NewMockProvider()
	mockProvider2 := provider.NewMockProvider()
	mockProvider3 := provider.NewMockProvider()

	// Create a multi-agent setup with the providers
	multiProvider := provider.NewMultiProvider(
		[]provider.ProviderWeight{
			{Provider: mockProvider1, Weight: 1.0},
			{Provider: mockProvider2, Weight: 1.0},
			{Provider: mockProvider3, Weight: 1.0},
		},
		provider.StrategyFastest,
	)

	multiAgent := workflow.NewAgent(multiProvider)
	for _, tool := range mockTools {
		multiAgent.AddTool(tool)
	}

	// Add a shared thread-safe tool counter for the multi agent
	multiAgentToolCounter := &safeToolCounter{}
	multiAgent.WithHook(multiAgentToolCounter)

	// Agent configs for testing
	agents := []struct {
		name      string
		agent     domain.Agent
		toolCount *safeToolCounter
	}{
		{"BaseAgent", baseAgent, baseAgentToolCounter},
		{"CachedAgent", cachedAgent, cachedAgentToolCounter},
		{"MultiAgent", multiAgent, multiAgentToolCounter},
	}

	// Define concurrency levels to test
	concurrencyLevels := []int{5, 20, 50, 100}

	// Define various prompts to simulate real-world variety
	prompts := []string{
		"What is 123 + 456?",
		"What's the weather like in San Francisco?",
		"Find information about the history of computers",
		"Calculate 987 - 654",
		"What's the weather like in Tokyo?",
		"Search for information about artificial intelligence",
		"What's 42 * 7?",
		"Weather forecast for London",
		"Find details about quantum computing",
		"Calculate the square root of 144",
	}

	// Run tests for each agent type and concurrency level
	for _, a := range agents {
		for _, concurrency := range concurrencyLevels {
			t.Run(fmt.Sprintf("%s_Concurrency_%d", a.name, concurrency), func(t *testing.T) {
				var (
					wg             sync.WaitGroup
					successful     int32
					failed         int32
					totalLatencyMs int64
					maxLatencyMs   int64
					minLatencyMs   int64 = 999999
				)

				// Reset the tool counter for each test
				a.toolCount.Reset()

				// Set a reasonable timeout
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
				defer cancel()

				// Create a semaphore to limit concurrent goroutines
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
						_, err := a.agent.Run(ctx, prompt)
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

				// Get final tool invocation count
				toolInvocations := a.toolCount.Count()

				// Check goroutine count after test
				peakGoroutines := runtime.NumGoroutine()

				// Record results
				successRate := float64(successful) / float64(successful+failed) * 100
				avgLatencyMs := float64(totalLatencyMs) / float64(successful+failed)
				avgToolsPerRequest := float64(toolInvocations) / float64(successful+failed)

				t.Logf("Results for %s at concurrency %d:", a.name, concurrency)
				t.Logf("  Success rate: %.2f%% (%d/%d)", successRate, successful, successful+failed)
				t.Logf("  Average latency: %.2f ms", avgLatencyMs)
				t.Logf("  Min latency: %d ms", minLatencyMs)
				t.Logf("  Max latency: %d ms", maxLatencyMs)
				t.Logf("  Total duration: %v", totalDuration)
				t.Logf("  Average tool invocations per request: %.2f", avgToolsPerRequest)
				t.Logf("  Goroutines: %d initial, %d peak", initialGoroutines, peakGoroutines)

				// Basic validation
				if successful == 0 {
					t.Errorf("No successful requests for %s at concurrency %d", a.name, concurrency)
				}
			})
		}
	}

	// Collect final memory stats
	runtime.ReadMemStats(&memStatsAfter)

	// Report memory usage
	t.Logf("Memory usage before: %.2f MB", float64(memStatsBefore.Alloc)/1024/1024)
	t.Logf("Memory usage after: %.2f MB", float64(memStatsAfter.Alloc)/1024/1024)
	t.Logf("Memory difference: %.2f MB", float64(memStatsAfter.Alloc-memStatsBefore.Alloc)/1024/1024)
	t.Logf("Total allocations: %d objects", memStatsAfter.Mallocs-memStatsBefore.Mallocs)
}

// safeToolCounter is a thread-safe hook for counting tool invocations
type safeToolCounter struct {
	count int32
	// mu is not used since we're using atomic operations
}

func (t *safeToolCounter) Reset() {
	atomic.StoreInt32(&t.count, 0)
}

func (t *safeToolCounter) Count() int32 {
	return atomic.LoadInt32(&t.count)
}

func (t *safeToolCounter) BeforeToolCall(ctx context.Context, tool string, args map[string]interface{}) {
}

func (t *safeToolCounter) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	atomic.AddInt32(&t.count, 1)
}

func (t *safeToolCounter) BeforeGenerate(ctx context.Context, messages []llmdomain.Message) {}
func (t *safeToolCounter) AfterGenerate(ctx context.Context, response llmdomain.Response, err error) {
}
