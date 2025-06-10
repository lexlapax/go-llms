// ABOUTME: Tests for LoopAgent workflow implementation
// ABOUTME: Validates iterative processing, condition evaluation, and termination logic

package workflow

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestLoopAgent_CountLoop(t *testing.T) {
	// Create a simple step that increments a counter
	incrementStep := &counterStep{
		name: "increment",
	}

	// Create count loop that runs 5 times
	loop := CountLoop("count-test", 5, incrementStep)

	ctx := context.Background()
	initialState := domain.NewState()
	initialState.Set("counter", 0)

	result, err := loop.Run(ctx, initialState)
	if err != nil {
		t.Fatalf("Loop failed: %v", err)
	}

	// Check final counter value
	if counter, exists := result.Get("counter"); !exists || counter != 5 {
		t.Errorf("Expected counter to be 5, got: %v", counter)
	}

	// Check iteration count
	if loop.GetCurrentIteration() != 5 {
		t.Errorf("Expected 5 iterations, got: %d", loop.GetCurrentIteration())
	}
}

func TestLoopAgent_WhileLoop(t *testing.T) {
	// Create a step that increments a counter
	incrementStep := &counterStep{
		name: "increment",
	}

	// Create while loop that continues while counter < 3
	loop := WhileLoop("while-test", func(state *domain.State, iteration int) bool {
		if counter, exists := state.Get("counter"); exists {
			return counter.(int) < 3
		}
		return false
	}, incrementStep)

	ctx := context.Background()
	initialState := domain.NewState()
	initialState.Set("counter", 0)

	result, err := loop.Run(ctx, initialState)
	if err != nil {
		t.Fatalf("Loop failed: %v", err)
	}

	// Check final counter value
	if counter, exists := result.Get("counter"); !exists || counter != 3 {
		t.Errorf("Expected counter to be 3, got: %v", counter)
	}

	// Check iteration count
	if loop.GetCurrentIteration() != 3 {
		t.Errorf("Expected 3 iterations, got: %d", loop.GetCurrentIteration())
	}
}

func TestLoopAgent_UntilLoop(t *testing.T) {
	// Create a step that increments a counter
	incrementStep := &counterStep{
		name: "increment",
	}

	// Create until loop that continues until counter >= 4
	loop := UntilLoop("until-test", func(state *domain.State, iteration int) bool {
		if counter, exists := state.Get("counter"); exists {
			return counter.(int) >= 4
		}
		return false
	}, incrementStep)

	ctx := context.Background()
	initialState := domain.NewState()
	initialState.Set("counter", 0)

	result, err := loop.Run(ctx, initialState)
	if err != nil {
		t.Fatalf("Loop failed: %v", err)
	}

	// Check final counter value
	if counter, exists := result.Get("counter"); !exists || counter != 4 {
		t.Errorf("Expected counter to be 4, got: %v", counter)
	}

	// Check iteration count
	if loop.GetCurrentIteration() != 4 {
		t.Errorf("Expected 4 iterations, got: %d", loop.GetCurrentIteration())
	}
}

func TestLoopAgent_MaxDuration(t *testing.T) {
	// Create a step with delay
	delayStep := &delayStep{
		name:  "delay",
		delay: 50 * time.Millisecond,
	}

	// Create loop with max duration of 100ms
	loop := NewLoopAgent("duration-test").
		SetLoopBody(delayStep).
		WithMaxDuration(100 * time.Millisecond).
		WithMaxIterations(10) // Set high iteration limit so duration limit kicks in first

	ctx := context.Background()
	initialState := domain.NewState()

	start := time.Now()
	result, err := loop.Run(ctx, initialState)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Loop failed: %v", err)
	}

	// Should have stopped due to duration limit, not iteration limit
	if loop.GetCurrentIteration() >= 10 {
		t.Errorf("Expected less than 10 iterations due to duration limit, got: %d", loop.GetCurrentIteration())
	}

	// Duration should be around 100ms (allow some tolerance)
	if duration < 90*time.Millisecond || duration > 200*time.Millisecond {
		t.Errorf("Expected duration around 100ms, got: %v", duration)
	}

	if result == nil {
		t.Error("Expected result but got nil")
	}
}

func TestLoopAgent_BreakOnError(t *testing.T) {
	// Create a step that errors on iteration 2
	errorStep := &errorStep{
		name:             "error",
		errorOnIteration: 2,
	}

	// Create loop that breaks on error
	loop := NewLoopAgent("error-test").
		SetLoopBody(errorStep).
		WithMaxIterations(5).
		WithBreakOnError(true)

	ctx := context.Background()
	initialState := domain.NewState()

	result, err := loop.Run(ctx, initialState)
	if err == nil {
		t.Fatal("Expected error but got none")
	}

	if result != nil {
		t.Error("Expected nil result on error")
	}

	// Should have stopped at iteration 2
	if loop.GetCurrentIteration() != 2 {
		t.Errorf("Expected to stop at iteration 2, got: %d", loop.GetCurrentIteration())
	}
}

func TestLoopAgent_ContinueOnError(t *testing.T) {
	// Create a step that errors on iteration 2
	errorStep := &errorStep{
		name:             "continue-error",
		errorOnIteration: 2,
	}

	// Create loop that continues on error
	loop := NewLoopAgent("continue-error-test").
		SetLoopBody(errorStep).
		WithMaxIterations(5).
		WithBreakOnError(false)

	ctx := context.Background()
	initialState := domain.NewState()

	result, err := loop.Run(ctx, initialState)
	if err != nil {
		t.Fatalf("Loop should not fail when continuing on error: %v", err)
	}

	// Should have completed all 5 iterations despite error
	if loop.GetCurrentIteration() != 5 {
		t.Errorf("Expected 5 iterations, got: %d", loop.GetCurrentIteration())
	}

	if result == nil {
		t.Error("Expected result but got nil")
	}
}

func TestLoopAgent_CollectResults(t *testing.T) {
	// Create a step that increments a counter
	incrementStep := &counterStep{
		name: "collect",
	}

	// Create loop with result collection enabled
	loop := NewLoopAgent("collect-test").
		SetLoopBody(incrementStep).
		WithMaxIterations(3).
		WithCollectResults(true)

	ctx := context.Background()
	initialState := domain.NewState()
	initialState.Set("counter", 0)

	result, err := loop.Run(ctx, initialState)
	if err != nil {
		t.Fatalf("Loop failed: %v", err)
	}

	// Check that results were collected
	results := loop.GetIterationResults()
	if len(results) != 3 {
		t.Errorf("Expected 3 iteration results, got: %d", len(results))
	}

	// Check individual iteration results
	for i, resultItem := range results {
		resultMap := resultItem.(map[string]interface{})
		if resultMap["iteration"] != i {
			t.Errorf("Expected iteration %d, got: %v", i, resultMap["iteration"])
		}
		if resultMap["error"] != nil {
			t.Errorf("Expected no error for iteration %d, got: %v", i, resultMap["error"])
		}
	}

	if result == nil {
		t.Error("Expected result but got nil")
	}
}

func TestLoopAgent_NoPassStateThrough(t *testing.T) {
	// Create a step that increments a counter
	incrementStep := &counterStep{
		name: "no-pass",
	}

	// Create loop that doesn't pass state between iterations
	loop := NewLoopAgent("no-pass-test").
		SetLoopBody(incrementStep).
		WithMaxIterations(3).
		WithPassStateThrough(false)

	ctx := context.Background()
	initialState := domain.NewState()
	initialState.Set("counter", 10) // Start with 10

	result, err := loop.Run(ctx, initialState)
	if err != nil {
		t.Fatalf("Loop failed: %v", err)
	}

	// Since state is not passed through, each iteration should start fresh
	// The final result should be the original state
	if counter, exists := result.Get("counter"); !exists || counter != 10 {
		t.Errorf("Expected counter to remain 10 (no state passthrough), got: %v", counter)
	}
}

func TestLoopAgent_IterationDelay(t *testing.T) {
	// Create a simple increment step
	incrementStep := &counterStep{
		name: "delay-test",
	}

	// Create loop with iteration delay
	delay := 50 * time.Millisecond
	loop := NewLoopAgent("delay-test").
		SetLoopBody(incrementStep).
		WithMaxIterations(3).
		WithIterationDelay(delay)

	ctx := context.Background()
	initialState := domain.NewState()
	initialState.Set("counter", 0)

	start := time.Now()
	result, err := loop.Run(ctx, initialState)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Loop failed: %v", err)
	}

	// Should take at least 2 * delay (delay between 3 iterations = 2 delays)
	expectedMinDuration := 2 * delay
	if duration < expectedMinDuration {
		t.Errorf("Expected duration at least %v, got: %v", expectedMinDuration, duration)
	}

	// Check final result
	if counter, exists := result.Get("counter"); !exists || counter != 3 {
		t.Errorf("Expected counter to be 3, got: %v", counter)
	}
}

func TestLoopAgent_Validation(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *LoopAgent
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid loop",
			setup: func() *LoopAgent {
				step := &counterStep{name: "valid"}
				return NewLoopAgent("valid").
					SetLoopBody(step).
					WithMaxIterations(5)
			},
			expectError: false,
		},
		{
			name: "No loop body",
			setup: func() *LoopAgent {
				return NewLoopAgent("no-body").
					WithMaxIterations(5)
			},
			expectError: true,
			errorMsg:    "loop workflow must have a loop body",
		},
		{
			name: "No termination condition",
			setup: func() *LoopAgent {
				step := &counterStep{name: "no-termination"}
				return NewLoopAgent("no-termination").
					SetLoopBody(step)
				// No termination conditions set
			},
			expectError: true,
			errorMsg:    "loop workflow must have at least one termination condition (maxIterations, maxDuration, continueCondition, or breakCondition)",
		},
		{
			name: "Negative iteration delay",
			setup: func() *LoopAgent {
				step := &counterStep{name: "negative-delay"}
				return NewLoopAgent("negative-delay").
					SetLoopBody(step).
					WithMaxIterations(5).
					WithIterationDelay(-1 * time.Second)
			},
			expectError: true,
			errorMsg:    "iteration delay cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loop := tt.setup()
			err := loop.Validate()

			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error but got: %v", err)
			}
			if tt.expectError && err != nil && err.Error() != tt.errorMsg {
				t.Errorf("Expected error message '%s', got: %s", tt.errorMsg, err.Error())
			}
		})
	}
}

func TestLoopAgent_Reset(t *testing.T) {
	incrementStep := &counterStep{name: "reset-test"}
	loop := NewLoopAgent("reset-test").
		SetLoopBody(incrementStep).
		WithMaxIterations(3)

	ctx := context.Background()
	initialState := domain.NewState()
	initialState.Set("counter", 0)

	// Run first time
	_, err := loop.Run(ctx, initialState)
	if err != nil {
		t.Fatalf("First run failed: %v", err)
	}

	// Check state before reset
	if loop.GetCurrentIteration() != 3 {
		t.Errorf("Expected 3 iterations before reset, got: %d", loop.GetCurrentIteration())
	}

	// Reset and run again
	loop.Reset()

	if loop.GetCurrentIteration() != 0 {
		t.Errorf("Expected 0 iterations after reset, got: %d", loop.GetCurrentIteration())
	}

	// Run again
	_, err = loop.Run(ctx, initialState)
	if err != nil {
		t.Fatalf("Second run failed: %v", err)
	}

	if loop.GetCurrentIteration() != 3 {
		t.Errorf("Expected 3 iterations after reset and rerun, got: %d", loop.GetCurrentIteration())
	}
}

// Helper test steps

type counterStep struct {
	name string
}

func (c *counterStep) Name() string {
	return c.name
}

func (c *counterStep) Execute(ctx context.Context, state *WorkflowState) (*WorkflowState, error) {
	currentCounter := 0
	if counter, exists := state.Get("counter"); exists {
		currentCounter = counter.(int)
	}

	newState := state.Clone()
	newState.Set("counter", currentCounter+1)

	return &WorkflowState{
		State:    newState,
		Metadata: make(map[string]interface{}),
	}, nil
}

func (c *counterStep) Validate() error {
	return nil
}

type delayStep struct {
	name  string
	delay time.Duration
}

func (d *delayStep) Name() string {
	return d.name
}

func (d *delayStep) Execute(ctx context.Context, state *WorkflowState) (*WorkflowState, error) {
	if d.delay > 0 {
		select {
		case <-time.After(d.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return state, nil
}

func (d *delayStep) Validate() error {
	return nil
}

type errorStep struct {
	name             string
	errorOnIteration int
	currentIteration int
}

func (e *errorStep) Name() string {
	return e.name
}

func (e *errorStep) Execute(ctx context.Context, state *WorkflowState) (*WorkflowState, error) {
	if e.currentIteration == e.errorOnIteration {
		return nil, fmt.Errorf("intentional error at iteration %d", e.errorOnIteration)
	}
	e.currentIteration++
	return state, nil
}

func (e *errorStep) Validate() error {
	return nil
}
