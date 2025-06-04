// ABOUTME: Tests for the parallel workflow agent
// ABOUTME: Validates concurrent execution, merge strategies, and error handling

package workflow

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestParallelAgent_BasicExecution(t *testing.T) {
	// Create agents that track execution
	var executionOrder []string
	var orderMutex sync.Mutex

	createTrackingAgent := func(name string, delay time.Duration) *MockAgent {
		return NewMockAgent(name).
			WithDelay(delay).
			WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
				orderMutex.Lock()
				executionOrder = append(executionOrder, name)
				orderMutex.Unlock()

				newState := state.Clone()
				newState.Set(fmt.Sprintf("%s_result", name), fmt.Sprintf("data_from_%s", name))
				return newState, nil
			})
	}

	agent1 := createTrackingAgent("agent1", 50*time.Millisecond)
	agent2 := createTrackingAgent("agent2", 20*time.Millisecond)
	agent3 := createTrackingAgent("agent3", 30*time.Millisecond)

	// Create parallel workflow
	workflow := NewParallelAgent("test-parallel")
	workflow.AddAgent(agent1)
	workflow.AddAgent(agent2)
	workflow.AddAgent(agent3)

	// Run workflow
	ctx := context.Background()
	initialState := domain.NewState()
	initialState.Set("input", "test")

	start := time.Now()
	result, err := workflow.Run(ctx, initialState)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	// Verify parallel execution (should take ~50ms, not 100ms)
	if duration > 80*time.Millisecond {
		t.Errorf("Execution took too long (%v), agents may not be running in parallel", duration)
	}

	// Verify all agents executed
	if len(executionOrder) != 3 {
		t.Errorf("Expected 3 agents to execute, got %d", len(executionOrder))
	}

	// With parallel execution, agent2 should typically finish first (shortest delay)
	if len(executionOrder) > 0 && executionOrder[0] != "agent2" {
		t.Logf("Expected agent2 to finish first, but order was: %v", executionOrder)
	}

	// Verify results are merged
	parallelResults, exists := result.Get("parallel_results")
	if !exists {
		t.Fatal("No parallel_results in final state")
	}

	results, ok := parallelResults.(map[string]interface{})
	if !ok {
		t.Fatalf("parallel_results has wrong type: %T", parallelResults)
	}

	// Check each agent's results
	for _, agentName := range []string{"agent1", "agent2", "agent3"} {
		agentResult, exists := results[agentName]
		if !exists {
			t.Errorf("No result for %s", agentName)
			continue
		}

		resultMap, ok := agentResult.(map[string]interface{})
		if !ok {
			t.Errorf("Result for %s has wrong type: %T", agentName, agentResult)
		}

		// For now we're only storing response/result keys
		// In a real implementation, we'd need a way to iterate all state keys
		t.Logf("Agent %s results: %v", agentName, resultMap)
	}
}

func TestParallelAgent_MaxConcurrency(t *testing.T) {
	// Track concurrent executions
	var currentConcurrent int32
	var maxConcurrent int32

	createConcurrencyAgent := func(name string) *MockAgent {
		return NewMockAgent(name).WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
			// Increment concurrent count
			current := atomic.AddInt32(&currentConcurrent, 1)

			// Track max
			for {
				max := atomic.LoadInt32(&maxConcurrent)
				if current <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
					break
				}
			}

			// Simulate work
			time.Sleep(50 * time.Millisecond)

			// Decrement concurrent count
			atomic.AddInt32(&currentConcurrent, -1)

			return state.Clone(), nil
		})
	}

	// Create 5 agents
	workflow := NewParallelAgent("test-concurrency").
		WithMaxConcurrency(2) // Limit to 2 concurrent

	for i := 0; i < 5; i++ {
		workflow.AddAgent(createConcurrencyAgent(fmt.Sprintf("agent%d", i)))
	}

	// Run workflow
	ctx := context.Background()
	initialState := domain.NewState()

	_, err := workflow.Run(ctx, initialState)
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	// Verify max concurrency was respected
	if maxConcurrent > 2 {
		t.Errorf("Max concurrency exceeded limit: got %d, want <= 2", maxConcurrent)
	}
}

func TestParallelAgent_MergeStrategies(t *testing.T) {
	t.Run("MergeFirst", func(t *testing.T) {
		// Create agents with different delays
		fastAgent := NewMockAgent("fast").
			WithDelay(10 * time.Millisecond).
			WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
				newState := state.Clone()
				newState.Set("winner", "fast")
				return newState, nil
			})

		slowAgent := NewMockAgent("slow").
			WithDelay(100 * time.Millisecond).
			WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
				newState := state.Clone()
				newState.Set("winner", "slow")
				return newState, nil
			})

		workflow := NewParallelAgent("test-merge-first").
			WithMergeStrategy(MergeFirst)
		workflow.AddAgent(slowAgent)
		workflow.AddAgent(fastAgent)

		ctx := context.Background()
		initialState := domain.NewState()

		result, err := workflow.Run(ctx, initialState)
		if err != nil {
			t.Fatalf("Workflow failed: %v", err)
		}

		// Should get the fast agent's result
		winner, exists := result.Get("winner")
		if !exists || winner != "fast" {
			t.Errorf("Expected fast agent to win, got: %v", winner)
		}
	})

	t.Run("MergeCustom", func(t *testing.T) {
		agent1 := NewMockAgent("agent1").WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
			newState := state.Clone()
			newState.Set("value", 10)
			return newState, nil
		})

		agent2 := NewMockAgent("agent2").WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
			newState := state.Clone()
			newState.Set("value", 20)
			return newState, nil
		})

		agent3 := NewMockAgent("agent3").WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
			newState := state.Clone()
			newState.Set("value", 30)
			return newState, nil
		})

		// Custom merge function that sums values
		customMerge := func(results map[string]*domain.State) *domain.State {
			merged := domain.NewState()
			sum := 0

			for agentName, state := range results {
				if value, exists := state.Get("value"); exists {
					if v, ok := value.(int); ok {
						sum += v
					}
				}
				merged.Set(fmt.Sprintf("%s_processed", agentName), true)
			}

			merged.Set("sum", sum)
			merged.Set("agent_count", len(results))
			return merged
		}

		workflow := NewParallelAgent("test-custom-merge").
			WithMergeFunc(customMerge)
		workflow.AddAgent(agent1)
		workflow.AddAgent(agent2)
		workflow.AddAgent(agent3)

		ctx := context.Background()
		initialState := domain.NewState()

		result, err := workflow.Run(ctx, initialState)
		if err != nil {
			t.Fatalf("Workflow failed: %v", err)
		}

		// Verify custom merge worked
		sum, exists := result.Get("sum")
		if !exists || sum != 60 {
			t.Errorf("Expected sum=60, got %v", sum)
		}

		count, exists := result.Get("agent_count")
		if !exists || count != 3 {
			t.Errorf("Expected agent_count=3, got %v", count)
		}
	})
}

func TestParallelAgent_ErrorHandling(t *testing.T) {
	t.Run("PartialFailure", func(t *testing.T) {
		agent1 := NewMockAgent("agent1")
		agent2 := NewMockAgent("agent2").WithError()
		agent3 := NewMockAgent("agent3")

		workflow := NewParallelAgent("test-partial-failure")
		workflow.AddAgent(agent1)
		workflow.AddAgent(agent2)
		workflow.AddAgent(agent3)

		ctx := context.Background()
		initialState := domain.NewState()

		// With MergeAll strategy, should fail
		_, err := workflow.Run(ctx, initialState)
		if err == nil {
			t.Fatal("Expected error for partial failure")
		}

		// Try with MergeFirst strategy - should succeed if any agent succeeds
		workflow2 := NewParallelAgent("test-partial-success").
			WithMergeStrategy(MergeFirst)
		workflow2.AddAgent(agent1)
		workflow2.AddAgent(agent2)
		workflow2.AddAgent(agent3)

		result, err := workflow2.Run(ctx, initialState)
		if err != nil {
			t.Fatalf("Expected success with MergeFirst: %v", err)
		}

		// Should have a result from one of the successful agents
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		slowAgent := NewMockAgent("slow").WithDelay(200 * time.Millisecond)

		workflow := NewParallelAgent("test-timeout").
			WithTimeout(50 * time.Millisecond)
		workflow.AddAgent(slowAgent)

		ctx := context.Background()
		initialState := domain.NewState()

		start := time.Now()
		_, err := workflow.Run(ctx, initialState)
		duration := time.Since(start)

		// Should timeout
		if err == nil {
			t.Fatal("Expected timeout error")
		}

		// Should take about 50ms, not 200ms
		if duration > 100*time.Millisecond {
			t.Errorf("Timeout took too long: %v", duration)
		}
	})
}
