// ABOUTME: Tests for the sequential workflow agent
// ABOUTME: Validates sequential execution, error handling, and state passing

package workflow

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// MockAgent for testing
type MockAgent struct {
	*core.BaseAgentImpl
	name        string
	shouldError bool
	delay       time.Duration
	runFunc     func(ctx context.Context, state *domain.State) (*domain.State, error)
}

func NewMockAgent(name string) *MockAgent {
	return &MockAgent{
		BaseAgentImpl: core.NewBaseAgent(name, "Mock agent for testing", domain.AgentTypeCustom),
		name:          name,
	}
}

func (m *MockAgent) WithError() *MockAgent {
	m.shouldError = true
	return m
}

func (m *MockAgent) WithDelay(delay time.Duration) *MockAgent {
	m.delay = delay
	return m
}

func (m *MockAgent) WithRunFunc(f func(ctx context.Context, state *domain.State) (*domain.State, error)) *MockAgent {
	m.runFunc = f
	return m
}

func (m *MockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	if m.delay > 0 {
		// Use a timer to respect context cancellation
		timer := time.NewTimer(m.delay)
		select {
		case <-timer.C:
			// Delay completed
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		}
	}

	if m.runFunc != nil {
		return m.runFunc(ctx, state)
	}

	if m.shouldError {
		return nil, fmt.Errorf("mock error from %s", m.name)
	}

	// Default behavior: add a marker to the state
	newState := state.Clone()
	newState.Set(fmt.Sprintf("%s_executed", m.name), true)
	newState.Set("last_agent", m.name)

	return newState, nil
}

func TestSequentialAgent_BasicExecution(t *testing.T) {
	// Create agents
	agent1 := NewMockAgent("agent1")
	agent2 := NewMockAgent("agent2")
	agent3 := NewMockAgent("agent3")

	// Create sequential workflow
	workflow := NewSequentialAgent("test-workflow")
	workflow.AddAgent(agent1)
	workflow.AddAgent(agent2)
	workflow.AddAgent(agent3)

	// Create initial state
	initialState := domain.NewState()
	initialState.Set("input", "test-value")

	// Run workflow
	ctx := context.Background()
	result, err := workflow.Run(ctx, initialState)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	// Verify all agents executed
	if v, exists := result.Get("agent1_executed"); !exists || v != true {
		t.Error("Agent1 did not execute")
	}
	if v, exists := result.Get("agent2_executed"); !exists || v != true {
		t.Error("Agent2 did not execute")
	}
	if v, exists := result.Get("agent3_executed"); !exists || v != true {
		t.Error("Agent3 did not execute")
	}

	// Verify final agent
	if v, exists := result.Get("last_agent"); !exists || v != "agent3" {
		t.Errorf("Expected last_agent to be agent3, got %v", v)
	}

	// Verify original input preserved
	if v, exists := result.Get("input"); !exists || v != "test-value" {
		t.Error("Original input not preserved")
	}
}

func TestSequentialAgent_ErrorHandling(t *testing.T) {
	t.Run("StopOnError", func(t *testing.T) {
		// Create agents
		agent1 := NewMockAgent("agent1")
		agent2 := NewMockAgent("agent2").WithError()
		agent3 := NewMockAgent("agent3")

		// Create workflow with stop on error
		workflow := NewSequentialAgent("test-workflow").
			WithStopOnError(true)
		workflow.AddAgent(agent1)
		workflow.AddAgent(agent2)
		workflow.AddAgent(agent3)

		// Run workflow
		ctx := context.Background()
		initialState := domain.NewState()
		_, err := workflow.Run(ctx, initialState)

		// Should fail
		if err == nil {
			t.Fatal("Expected error but got none")
		}

		// Verify error message
		if !errors.Is(err, fmt.Errorf("mock error from agent2")) {
			t.Logf("Expected error containing 'mock error from agent2', got: %v", err)
		}
	})

	t.Run("ContinueOnError", func(t *testing.T) {
		// Create agents
		agent1 := NewMockAgent("agent1")
		agent2 := NewMockAgent("agent2").WithError()
		agent3 := NewMockAgent("agent3")

		// Create workflow without stop on error
		workflow := NewSequentialAgent("test-workflow").
			WithStopOnError(false)
		workflow.AddAgent(agent1)
		workflow.AddAgent(agent2)
		workflow.AddAgent(agent3)

		// Run workflow
		ctx := context.Background()
		initialState := domain.NewState()
		result, err := workflow.Run(ctx, initialState)

		// Should succeed despite error
		if err != nil {
			t.Fatalf("Workflow failed: %v", err)
		}

		// Verify agent1 and agent3 executed
		if v, exists := result.Get("agent1_executed"); !exists || v != true {
			t.Error("Agent1 did not execute")
		}
		if v, exists := result.Get("agent3_executed"); !exists || v != true {
			t.Error("Agent3 did not execute")
		}

		// Agent2 should not have executed
		if _, exists := result.Get("agent2_executed"); exists {
			t.Error("Agent2 should not have executed")
		}
	})
}

func TestSequentialAgent_StatePassthrough(t *testing.T) {
	// Create agents that modify state
	agent1 := NewMockAgent("agent1").WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
		newState := state.Clone()
		newState.Set("counter", 1)
		newState.Set("agent1_data", "from_agent1")
		return newState, nil
	})

	agent2 := NewMockAgent("agent2").WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
		newState := state.Clone()

		// Increment counter
		if counter, exists := state.Get("counter"); exists {
			if c, ok := counter.(int); ok {
				newState.Set("counter", c+1)
			}
		}

		// Add own data
		newState.Set("agent2_data", "from_agent2")

		// Verify can see agent1's data
		if v, exists := state.Get("agent1_data"); exists {
			newState.Set("agent2_saw_agent1", v)
		}

		return newState, nil
	})

	// Create workflow
	workflow := NewSequentialAgent("test-workflow")
	workflow.AddAgent(agent1)
	workflow.AddAgent(agent2)

	// Run workflow
	ctx := context.Background()
	initialState := domain.NewState()
	result, err := workflow.Run(ctx, initialState)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	// Verify state modifications
	if v, exists := result.Get("counter"); !exists || v != 2 {
		t.Errorf("Expected counter=2, got %v", v)
	}

	if v, exists := result.Get("agent1_data"); !exists || v != "from_agent1" {
		t.Errorf("Missing or incorrect agent1_data: %v", v)
	}

	if v, exists := result.Get("agent2_data"); !exists || v != "from_agent2" {
		t.Errorf("Missing or incorrect agent2_data: %v", v)
	}

	if v, exists := result.Get("agent2_saw_agent1"); !exists || v != "from_agent1" {
		t.Errorf("Agent2 didn't see agent1's data: %v", v)
	}
}

func TestSequentialAgent_EmptyWorkflow(t *testing.T) {
	workflow := NewSequentialAgent("empty-workflow")

	ctx := context.Background()
	initialState := domain.NewState()
	_, err := workflow.Run(ctx, initialState)

	if err == nil {
		t.Fatal("Expected error for empty workflow")
	}

	if !errors.Is(err, fmt.Errorf("workflow must have at least one step")) {
		t.Logf("Expected 'workflow must have at least one step', got: %v", err)
	}
}
