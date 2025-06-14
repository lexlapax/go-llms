package stress

// ABOUTME: Stress tests for agent stability under high load and concurrency
// ABOUTME: Tests memory usage, concurrent execution, and system stability

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
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
		mocks.CreateCalculatorTool(),
		mocks.NewMockTool("weather", "Gets weather information"),
		mocks.NewMockTool("search", "Searches for information"),
	}

	// Create a base mock provider for all agents
	baseProvider := provider.NewMockProvider()

	// Create agent with new architecture
	deps := core.LLMDeps{
		Provider: baseProvider,
	}
	baseAgent := core.NewLLMAgent("stress-test-agent", "mock", deps)
	for _, tool := range mockTools {
		baseAgent.AddTool(tool)
	}

	// Add a shared thread-safe tool counter for each agent type
	baseAgentToolCounter := &safeToolCounter{}
	baseAgent.WithHook(baseAgentToolCounter)

	// Set system prompts
	baseAgent.SetSystemPrompt("You are a helpful assistant.")

	// Number of concurrent goroutines and requests per goroutine
	concurrency := 50
	requestsPerGoroutine := 20

	// Track successes and failures
	var successCount, failureCount int64
	var wg sync.WaitGroup

	// Mock provider response generator
	baseProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []llmdomain.Message, options ...llmdomain.Option) (llmdomain.Response, error) {
		// Simulate varying response times
		delay := time.Duration(rand.Intn(50)) * time.Millisecond //nolint:gosec // Non-crypto random is fine for test delays
		select {
		case <-time.After(delay):
			// Randomly choose to use a tool or not
			if rand.Float32() < 0.3 { //nolint:gosec // Non-crypto random is fine for test scenarios
				// Return a tool-using response
				return llmdomain.Response{
					Content: `I'll help you with that calculation.

<tool_calls>
[
  {
    "name": "calculator",
    "arguments": {
      "operation": "add",
      "a": 10,
      "b": 20
    }
  }
]
</tool_calls>`,
				}, nil
			}
			// Return a simple response
			return llmdomain.Response{
				Content: "Here's a helpful response to your query.",
			}, nil
		case <-ctx.Done():
			return llmdomain.Response{}, ctx.Err()
		}
	})

	// Run concurrent requests
	startTime := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

				// Create state with varying inputs
				state := domain.NewState()
				queries := []string{
					"What is 2 + 2?",
					"What's the weather like?",
					"Search for information about Go programming",
					"Calculate 15 * 7",
					"Tell me a joke",
				}
				state.Set("user_input", queries[rand.Intn(len(queries))]) //nolint:gosec // Non-crypto random is fine for test input selection

				// Run the agent
				_, err := baseAgent.Run(ctx, state)
				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					t.Logf("Worker %d request %d failed: %v", workerID, j, err)
				} else {
					atomic.AddInt64(&successCount, 1)
				}

				cancel()

				// Small random delay between requests
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	duration := time.Since(startTime)

	// Check results
	totalRequests := int64(concurrency * requestsPerGoroutine)
	t.Logf("Stress test completed in %v", duration)
	t.Logf("Total requests: %d", totalRequests)
	t.Logf("Successful: %d (%.2f%%)", successCount, float64(successCount)/float64(totalRequests)*100)
	t.Logf("Failed: %d (%.2f%%)", failureCount, float64(failureCount)/float64(totalRequests)*100)
	t.Logf("Requests/second: %.2f", float64(totalRequests)/duration.Seconds())

	// Log tool usage stats
	t.Logf("Tool calls - Base agent: %d", baseAgentToolCounter.getCount())

	// Check memory usage
	runtime.ReadMemStats(&memStatsAfter)
	memoryIncrease := memStatsAfter.Alloc - memStatsBefore.Alloc
	t.Logf("Memory increase: %d bytes (%.2f MB)", memoryIncrease, float64(memoryIncrease)/(1024*1024))

	// Ensure most requests succeeded
	successRate := float64(successCount) / float64(totalRequests)
	if successRate < 0.95 {
		t.Errorf("Success rate too low: %.2f%% (expected > 95%%)", successRate*100)
	}
}

// TestAgentMemoryLeaks tests for memory leaks in agent workflows
func TestAgentMemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	// Force garbage collection before starting
	runtime.GC()
	runtime.GC()

	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Create a mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []llmdomain.Message, options ...llmdomain.Option) (llmdomain.Response, error) {
		return llmdomain.Response{
			Content: "This is a test response that should be cleaned up properly.",
		}, nil
	})

	// Number of iterations
	iterations := 1000

	// Run many agent instances sequentially
	for i := 0; i < iterations; i++ {
		deps := core.LLMDeps{
			Provider: mockProvider,
		}
		agent := core.NewLLMAgent(fmt.Sprintf("leak-test-agent-%d", i), "mock", deps)
		agent.SetSystemPrompt("You are a helpful assistant for testing memory leaks.")

		// Add some tools
		agent.AddTool(mocks.CreateCalculatorTool())
		agent.AddTool(mocks.NewMockTool("tool1", "Test tool 1"))
		agent.AddTool(mocks.NewMockTool("tool2", "Test tool 2"))

		// Create and run with state
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		state := domain.NewState()
		state.Set("user_input", fmt.Sprintf("Test query %d", i))
		state.Set("metadata", map[string]interface{}{
			"iteration": i,
			"timestamp": time.Now(),
			"data":      make([]byte, 1024), // 1KB of data
		})

		_, _ = agent.Run(ctx, state)
		cancel()

		// Periodically force GC and check memory
		if i%100 == 0 {
			runtime.GC()
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			t.Logf("Iteration %d - Alloc: %.2f MB, TotalAlloc: %.2f MB, NumGC: %d",
				i,
				float64(memStats.Alloc)/(1024*1024),
				float64(memStats.TotalAlloc)/(1024*1024),
				memStats.NumGC,
			)
		}
	}

	// Force final garbage collection
	runtime.GC()
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Check final memory usage
	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Use signed integers to handle cases where memory decreases after GC
	// Safe conversion avoiding potential overflow
	var memoryIncrease int64
	if memStatsAfter.Alloc >= memStatsBefore.Alloc {
		memoryIncrease = int64(memStatsAfter.Alloc - memStatsBefore.Alloc)
	} else {
		memoryIncrease = -int64(memStatsBefore.Alloc - memStatsAfter.Alloc)
	}
	memoryIncreasePerIteration := float64(memoryIncrease) / float64(iterations)

	t.Logf("Memory statistics:")
	t.Logf("  Initial memory: %.2f MB", float64(memStatsBefore.Alloc)/(1024*1024))
	t.Logf("  Final memory: %.2f MB", float64(memStatsAfter.Alloc)/(1024*1024))
	t.Logf("  Total increase: %.2f MB", float64(memoryIncrease)/(1024*1024))
	t.Logf("  Increase per iteration: %.2f KB", memoryIncreasePerIteration/1024)
	t.Logf("  Number of GC runs: %d", memStatsAfter.NumGC-memStatsBefore.NumGC)

	// Check for excessive memory growth (more than 10KB per iteration suggests a leak)
	// Only check if memory actually increased
	if memoryIncrease > 0 && memoryIncreasePerIteration > 10*1024 {
		t.Errorf("Possible memory leak detected: %.2f KB per iteration", memoryIncreasePerIteration/1024)
	}
}

// TestAgentRapidContextCancellation tests agent behavior with rapid context cancellations
func TestAgentRapidContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rapid cancellation test in short mode")
	}

	// Create a mock provider that simulates slow responses
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []llmdomain.Message, options ...llmdomain.Option) (llmdomain.Response, error) {
		// Simulate a slow response
		select {
		case <-time.After(500 * time.Millisecond):
			return llmdomain.Response{Content: "Slow response"}, nil
		case <-ctx.Done():
			return llmdomain.Response{}, ctx.Err()
		}
	})

	// Create agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("cancellation-test-agent", "mock", deps)
	agent.SetSystemPrompt("Test agent for cancellation")

	// Track cancellation behavior
	var cancelledCount, completedCount int64
	iterations := 100

	var wg sync.WaitGroup
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			// Create context with random timeout
			timeout := time.Duration(rand.Intn(100)+10) * time.Millisecond
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Create state
			state := domain.NewState()
			state.Set("user_input", fmt.Sprintf("Request %d", iteration))

			// Run agent
			_, err := agent.Run(ctx, state)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					atomic.AddInt64(&cancelledCount, 1)
				}
			} else {
				atomic.AddInt64(&completedCount, 1)
			}

			// Sometimes cancel early
			if rand.Float32() < 0.3 {
				cancel()
			}
		}(i)

		// Small delay between launches
		time.Sleep(5 * time.Millisecond)
	}

	wg.Wait()

	t.Logf("Rapid cancellation test results:")
	t.Logf("  Total requests: %d", iterations)
	t.Logf("  Completed: %d", completedCount)
	t.Logf("  Canceled: %d", cancelledCount)
	t.Logf("  Other: %d", iterations-int(completedCount+cancelledCount))

	// Ensure the system handled cancellations properly
	if cancelledCount == 0 {
		t.Error("Expected some requests to be canceled but none were")
	}
}

// safeToolCounter is a thread-safe counter for hook testing
type safeToolCounter struct {
	count int64
}

func (h *safeToolCounter) BeforeGenerate(ctx context.Context, messages []llmdomain.Message) {
	// No-op
}

func (h *safeToolCounter) AfterGenerate(ctx context.Context, response llmdomain.Response, err error) {
	// No-op
}

func (h *safeToolCounter) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	atomic.AddInt64(&h.count, 1)
}

func (h *safeToolCounter) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	// No-op
}

func (h *safeToolCounter) getCount() int64 {
	return atomic.LoadInt64(&h.count)
}
