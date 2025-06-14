package stress

// ABOUTME: Stress tests for workflow agents including sequential, parallel, conditional, and loop agents
// ABOUTME: Tests concurrent execution, memory leaks, and state management under high load

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// TestWorkflowAgentsConcurrentExecution tests workflow agents under concurrent load
func TestWorkflowAgentsConcurrentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Track memory stats
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Create mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []llmdomain.Message, options ...llmdomain.Option) (llmdomain.Response, error) {
		// Simulate varying response times
		delay := time.Duration(rand.Intn(10)) * time.Millisecond
		select {
		case <-time.After(delay):
			return llmdomain.Response{
				Content: fmt.Sprintf("Processed: %v", messages[len(messages)-1].Content),
			}, nil
		case <-ctx.Done():
			return llmdomain.Response{}, ctx.Err()
		}
	})

	// Create base agents for workflows
	createAgent := func(name string) domain.BaseAgent {
		deps := core.LLMDeps{Provider: mockProvider}
		agent := core.NewLLMAgent(name, "mock", deps)
		agent.SetSystemPrompt(fmt.Sprintf("You are %s agent", name))
		return agent
	}

	// Test configurations
	concurrencyLevels := []int{10, 50, 100}
	workflowTypes := []struct {
		name       string
		createFunc func() domain.BaseAgent
		complexity string
	}{
		{
			name: "Sequential",
			createFunc: func() domain.BaseAgent {
				seq := workflow.NewSequentialAgent("seq-workflow")
				seq.AddAgent(createAgent("step1"))
				seq.AddAgent(createAgent("step2"))
				seq.AddAgent(createAgent("step3"))
				return seq
			},
			complexity: "low",
		},
		{
			name: "Parallel",
			createFunc: func() domain.BaseAgent {
				par := workflow.NewParallelAgent("par-workflow").
					WithMaxConcurrency(2).
					WithMergeStrategy(workflow.MergeAll)
				par.AddAgent(createAgent("parallel1"))
				par.AddAgent(createAgent("parallel2"))
				par.AddAgent(createAgent("parallel3"))
				par.AddAgent(createAgent("parallel4"))
				return par
			},
			complexity: "medium",
		},
		{
			name: "Conditional",
			createFunc: func() domain.BaseAgent {
				cond := workflow.NewConditionalAgent("cond-workflow")
				cond.AddAgent(
					"low-value",
					func(state *domain.State) bool {
						val, _ := state.Get("value")
						if num, ok := val.(int); ok {
							return num < 50
						}
						return false
					},
					createAgent("low-branch"),
				)
				cond.AddAgent(
					"high-value",
					func(state *domain.State) bool {
						val, _ := state.Get("value")
						if num, ok := val.(int); ok {
							return num >= 50
						}
						return false
					},
					createAgent("high-branch"),
				)
				return cond
			},
			complexity: "medium",
		},
		{
			name: "Loop",
			createFunc: func() domain.BaseAgent {
				loop := workflow.NewLoopAgent("loop-workflow").
					WithMaxIterations(5).
					WithWhileCondition(func(state *domain.State, iteration int) bool {
						return iteration < 5
					})
				loop.SetLoopAgent(createAgent("loop-body"))
				return loop
			},
			complexity: "high",
		},
		{
			name: "Nested",
			createFunc: func() domain.BaseAgent {
				// Create a complex nested workflow
				innerSeq := workflow.NewSequentialAgent("inner-seq")
				innerSeq.AddAgent(createAgent("inner1"))
				innerSeq.AddAgent(createAgent("inner2"))

				innerPar := workflow.NewParallelAgent("inner-par")
				innerPar.AddAgent(createAgent("par1"))
				innerPar.AddAgent(createAgent("par2"))

				nested := workflow.NewSequentialAgent("nested-workflow")
				nested.AddAgent(createAgent("start"))
				nested.AddAgent(innerSeq)
				nested.AddAgent(innerPar)
				nested.AddAgent(createAgent("end"))
				return nested
			},
			complexity: "high",
		},
	}

	// Run tests for each workflow type and concurrency level
	for _, wf := range workflowTypes {
		for _, concurrency := range concurrencyLevels {
			t.Run(fmt.Sprintf("%s_Concurrency_%d", wf.name, concurrency), func(t *testing.T) {
				var (
					wg           sync.WaitGroup
					successCount int64
					failureCount int64
					totalLatency int64
				)

				// Create workflow agent
				workflowAgent := wf.createFunc()

				// Run concurrent executions
				startTime := time.Now()
				for i := 0; i < concurrency; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()

						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						// Create state with varying data
						state := domain.NewState()
						state.Set("request_id", id)
						state.Set("value", rand.Intn(100))
						state.Set("user_input", fmt.Sprintf("Request %d", id))
						state.Set("iteration", 0)

						// Execute workflow
						execStart := time.Now()
						_, err := workflowAgent.Run(ctx, state)
						latency := time.Since(execStart).Milliseconds()
						atomic.AddInt64(&totalLatency, latency)

						if err != nil {
							atomic.AddInt64(&failureCount, 1)
						} else {
							atomic.AddInt64(&successCount, 1)
						}
					}(i)
				}

				wg.Wait()
				duration := time.Since(startTime)

				// Calculate metrics
				total := successCount + failureCount
				successRate := float64(successCount) / float64(total) * 100
				avgLatency := float64(totalLatency) / float64(total)

				t.Logf("Results for %s workflow (complexity: %s) at concurrency %d:",
					wf.name, wf.complexity, concurrency)
				t.Logf("  Total requests: %d", total)
				t.Logf("  Success rate: %.2f%%", successRate)
				t.Logf("  Average latency: %.2f ms", avgLatency)
				t.Logf("  Total duration: %v", duration)
				t.Logf("  Throughput: %.2f requests/sec", float64(total)/duration.Seconds())

				// Ensure high success rate
				if successRate < 95.0 {
					t.Errorf("Success rate too low: %.2f%% (expected > 95%%)", successRate)
				}
			})
		}
	}

	// Check memory usage
	runtime.ReadMemStats(&memStatsAfter)
	memoryIncrease := memStatsAfter.Alloc - memStatsBefore.Alloc
	t.Logf("Memory increase: %.2f MB", float64(memoryIncrease)/(1024*1024))
}

// TestWorkflowStateManagementStress tests state management under high concurrency
func TestWorkflowStateManagementStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []llmdomain.Message, options ...llmdomain.Option) (llmdomain.Response, error) {
		return llmdomain.Response{Content: "Processed"}, nil
	})

	// Create an agent that heavily uses state
	deps := core.LLMDeps{Provider: mockProvider}
	stateAgent := core.NewLLMAgent("state-test", "mock", deps)

	// Add a tool that modifies state
	stateAgent.AddTool(mocks.NewMockTool("state-modifier", "Modifies state"))

	// Test concurrent state operations
	concurrency := 100
	operations := 50
	var wg sync.WaitGroup
	var stateErrors int64

	// Test concurrent state operations without shared parent

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operations; j++ {
				// Create local state
				state := domain.NewState()

				// Set multiple values
				state.Set(fmt.Sprintf("worker_%d_key_%d", workerID, j), rand.Intn(1000))
				state.Set("shared_counter", workerID*operations+j)

				// Create nested structures
				state.Set("nested", map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"value": rand.Float64(),
							"array": []int{1, 2, 3, 4, 5},
						},
					},
				})

				// Test concurrent reads and writes
				for k := 0; k < 10; k++ {
					key := fmt.Sprintf("concurrent_key_%d", rand.Intn(10))
					if rand.Float32() < 0.5 {
						state.Set(key, rand.Intn(100))
					} else {
						_, _ = state.Get(key)
					}
				}

				// Verify state integrity
				val, exists := state.Get(fmt.Sprintf("worker_%d_key_%d", workerID, j))
				if !exists {
					atomic.AddInt64(&stateErrors, 1)
				}
				if _, ok := val.(int); !ok && val != nil {
					atomic.AddInt64(&stateErrors, 1)
				}
			}
		}(i)
	}

	wg.Wait()

	t.Logf("State management stress test results:")
	t.Logf("  Concurrent workers: %d", concurrency)
	t.Logf("  Operations per worker: %d", operations)
	t.Logf("  Total operations: %d", concurrency*operations)
	t.Logf("  State errors: %d", stateErrors)

	if stateErrors > 0 {
		t.Errorf("State integrity errors detected: %d", stateErrors)
	}
}

// TestWorkflowMemoryLeakDetection tests for memory leaks in workflow execution
func TestWorkflowMemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	// Force GC before starting
	runtime.GC()
	runtime.GC()

	var memStatsBefore runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	// Create mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []llmdomain.Message, options ...llmdomain.Option) (llmdomain.Response, error) {
		// Simulate response with large content
		content := make([]byte, 1024) // 1KB per response
		for i := range content {
			content[i] = byte('A' + i%26)
		}
		return llmdomain.Response{Content: string(content)}, nil
	})

	iterations := 500

	// Test different workflow types for memory leaks
	for i := 0; i < iterations; i++ {
		// Create agents
		deps := core.LLMDeps{Provider: mockProvider}
		agent1 := core.NewLLMAgent(fmt.Sprintf("agent1-%d", i), "mock", deps)
		agent2 := core.NewLLMAgent(fmt.Sprintf("agent2-%d", i), "mock", deps)
		agent3 := core.NewLLMAgent(fmt.Sprintf("agent3-%d", i), "mock", deps)

		// Create different workflow types
		seq := workflow.NewSequentialAgent(fmt.Sprintf("seq-%d", i))
		seq.AddAgent(agent1)
		seq.AddAgent(agent2)
		seq.AddAgent(agent3)

		par := workflow.NewParallelAgent(fmt.Sprintf("par-%d", i))
		par.AddAgent(agent1)
		par.AddAgent(agent2)

		loop := workflow.NewLoopAgent(fmt.Sprintf("loop-%d", i)).WithMaxIterations(3)
		loop.SetLoopAgent(agent1)

		workflows := []domain.BaseAgent{seq, par, loop}

		// Execute each workflow
		for _, wf := range workflows {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			state := domain.NewState()
			state.Set("iteration", i)
			state.Set("large_data", make([]byte, 10*1024)) // 10KB of data

			_, _ = wf.Run(ctx, state)
			cancel()
		}

		// Periodically force GC and check memory
		if i%50 == 0 {
			runtime.GC()
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			t.Logf("Iteration %d - Memory: %.2f MB, Goroutines: %d",
				i,
				float64(memStats.Alloc)/(1024*1024),
				runtime.NumGoroutine(),
			)
		}
	}

	// Final GC and memory check
	runtime.GC()
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	var memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsAfter)

	// Safe conversion avoiding potential overflow
	var memoryIncrease int64
	if memStatsAfter.Alloc >= memStatsBefore.Alloc {
		memoryIncrease = int64(memStatsAfter.Alloc - memStatsBefore.Alloc)
	} else {
		memoryIncrease = -int64(memStatsBefore.Alloc - memStatsAfter.Alloc)
	}
	memoryIncreasePerIteration := float64(memoryIncrease) / float64(iterations)

	t.Logf("Workflow memory leak detection results:")
	t.Logf("  Initial memory: %.2f MB", float64(memStatsBefore.Alloc)/(1024*1024))
	t.Logf("  Final memory: %.2f MB", float64(memStatsAfter.Alloc)/(1024*1024))
	t.Logf("  Memory increase per iteration: %.2f KB", memoryIncreasePerIteration/1024)

	// Check for excessive memory growth (more than 50KB per iteration suggests a leak)
	if memoryIncrease > 0 && memoryIncreasePerIteration > 50*1024 {
		t.Errorf("Possible memory leak detected: %.2f KB per iteration", memoryIncreasePerIteration/1024)
	}
}

// TestWorkflowErrorHandlingUnderLoad tests error handling in workflows under high load
func TestWorkflowErrorHandlingUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create mock provider that sometimes fails
	mockProvider := provider.NewMockProvider()
	var callCount int64
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []llmdomain.Message, options ...llmdomain.Option) (llmdomain.Response, error) {
		count := atomic.AddInt64(&callCount, 1)

		// Extract agent name from messages to provide deterministic behavior
		agentName := ""
		for _, msg := range messages {
			// Extract text content from message
			for _, part := range msg.Content {
				if part.Type == llmdomain.ContentTypeText && strings.Contains(part.Text, "You are") {
					// Extract agent name between "You are " and " agent"
					text := part.Text
					startIdx := strings.Index(text, "You are ")
					endIdx := strings.Index(text, " agent")
					if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
						startIdx += len("You are ")
						agentName = text[startIdx:endIdx]
					}
					break
				}
			}
			if agentName != "" {
				break
			}
		}

		// Make failure deterministic based on agent name
		// For parallel tests, ensure at least one agent always succeeds
		if strings.Contains(agentName, "always-succeeds") {
			return llmdomain.Response{Content: fmt.Sprintf("Success %d for %s", count, agentName)}, nil
		}

		// Agents with "might-fail" in name fail on every 3rd call FROM THAT AGENT
		// Use agent-specific counters to ensure deterministic behavior per agent
		if strings.Contains(agentName, "might-fail") {
			// Simple hash of agent name to get consistent behavior
			agentHash := int64(0)
			for _, c := range agentName {
				agentHash += int64(c)
			}
			// This agent fails every 3rd time it's called
			if (count+agentHash)%3 == 0 {
				return llmdomain.Response{}, fmt.Errorf("simulated error for %s at count %d", agentName, count)
			}
		}

		// Sequential agents always fail on step2-might-fail
		if agentName == "step2-might-fail" {
			return llmdomain.Response{}, fmt.Errorf("simulated error for %s", agentName)
		}

		return llmdomain.Response{Content: fmt.Sprintf("Success %d for %s", count, agentName)}, nil
	})

	// Create agents with error handling
	createAgent := func(name string) domain.BaseAgent {
		deps := core.LLMDeps{Provider: mockProvider}
		agent := core.NewLLMAgent(name, "mock", deps)
		agent.SetSystemPrompt(fmt.Sprintf("You are %s agent", name))
		return agent
	}

	// Test different error scenarios
	scenarios := []struct {
		name          string
		workflow      domain.BaseAgent
		expectFailure bool
	}{
		{
			name: "Sequential_WithErrors",
			workflow: func() domain.BaseAgent {
				seq := workflow.NewSequentialAgent("seq-errors")
				seq.AddAgent(createAgent("step1"))
				seq.AddAgent(createAgent("step2-might-fail"))
				seq.AddAgent(createAgent("step3"))
				return seq
			}(),
			expectFailure: true,
		},
		{
			name: "Parallel_PartialFailure",
			workflow: func() domain.BaseAgent {
				par := workflow.NewParallelAgent("par-errors")
				par.WithMergeStrategy(workflow.MergeFirst) // Use first successful result
				// Create agents that won't all fail at once
				par.AddAgent(createAgent("par1-always-succeeds"))
				par.AddAgent(createAgent("par2-might-fail"))
				par.AddAgent(createAgent("par3-always-succeeds"))
				return par
			}(),
			expectFailure: false, // MergeFirst allows workflow to succeed with partial failures
		},
		{
			name: "Loop_WithRetries",
			workflow: func() domain.BaseAgent {
				loop := workflow.NewLoopAgent("loop-retry").
					WithMaxIterations(3).
					WithBreakOnError(false) // Continue iterating even on errors
				loop.SetLoopAgent(createAgent("retry-body"))
				return loop
			}(),
			expectFailure: false,
		},
	}

	concurrency := 50
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			var wg sync.WaitGroup
			var successCount, failureCount, partialSuccessCount int64

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()

					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()

					state := domain.NewState()
					state.Set("request_id", id)
					state.Set("user_input", fmt.Sprintf("Process request %d", id))

					result, err := scenario.workflow.Run(ctx, state)

					if err != nil {
						// Check for partial success
						hasPartialSuccess := false

						// Check if we have any results in the state
						if result != nil {
							// For parallel workflows with MergeFirst, success means we got a result even with errors
							if _, exists := result.Get("response"); exists {
								hasPartialSuccess = true
							} else if _, exists := result.Get("result"); exists {
								hasPartialSuccess = true
							} else if parallelResults, exists := result.Get("parallel_results"); exists && parallelResults != nil {
								hasPartialSuccess = true
							}
						}

						if hasPartialSuccess {
							atomic.AddInt64(&partialSuccessCount, 1)
						} else {
							atomic.AddInt64(&failureCount, 1)
						}
					} else {
						atomic.AddInt64(&successCount, 1)
					}
				}(i)
			}

			wg.Wait()

			total := successCount + failureCount + partialSuccessCount
			successRate := float64(successCount) / float64(total) * 100
			partialRate := float64(partialSuccessCount) / float64(total) * 100
			failureRate := float64(failureCount) / float64(total) * 100

			t.Logf("Error handling results for %s:", scenario.name)
			t.Logf("  Total requests: %d", total)
			t.Logf("  Full success: %.2f%% (%d)", successRate, successCount)
			t.Logf("  Partial success: %.2f%% (%d)", partialRate, partialSuccessCount)
			t.Logf("  Complete failure: %.2f%% (%d)", failureRate, failureCount)

			// Verify expectations
			if scenario.expectFailure && successRate > 50 {
				t.Errorf("Expected high failure rate but got %.2f%% success", successRate)
			}
			if !scenario.expectFailure && failureRate > 50 {
				t.Errorf("Expected low failure rate but got %.2f%% failure", failureRate)
			}
		})
	}
}
