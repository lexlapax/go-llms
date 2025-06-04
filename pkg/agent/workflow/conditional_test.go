// ABOUTME: Tests for ConditionalAgent workflow implementation
// ABOUTME: Validates branching logic, condition evaluation, and error handling

package workflow

import (
	"context"
	"fmt"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestConditionalAgent_BasicExecution(t *testing.T) {
	// Create test agents
	agent1 := createMockAgentWithResponse("agent1", "branch1 executed")
	agent2 := createMockAgentWithResponse("agent2", "branch2 executed")
	defaultAgent := createMockAgentWithResponse("default", "default executed")

	// Create conditional workflow
	conditional := NewConditionalAgent("test-conditional").
		AddAgent("branch1", func(state *domain.State) bool {
			if value, exists := state.Get("condition"); exists {
				return value == "branch1"
			}
			return false
		}, agent1).
		AddAgent("branch2", func(state *domain.State) bool {
			if value, exists := state.Get("condition"); exists {
				return value == "branch2"
			}
			return false
		}, agent2).
		SetDefaultAgent(defaultAgent)

	ctx := context.Background()

	// Test branch1 execution
	state1 := domain.NewState()
	state1.Set("condition", "branch1")

	result1, err := conditional.Run(ctx, state1)
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	if response, exists := result1.Get("response"); !exists || response != "branch1 executed" {
		t.Errorf("Expected 'branch1 executed', got: %v", response)
	}

	// Test branch2 execution
	state2 := domain.NewState()
	state2.Set("condition", "branch2")

	result2, err := conditional.Run(ctx, state2)
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	if response, exists := result2.Get("response"); !exists || response != "branch2 executed" {
		t.Errorf("Expected 'branch2 executed', got: %v", response)
	}

	// Test default execution
	state3 := domain.NewState()
	state3.Set("condition", "unknown")

	result3, err := conditional.Run(ctx, state3)
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	if response, exists := result3.Get("response"); !exists || response != "default executed" {
		t.Errorf("Expected 'default executed', got: %v", response)
	}
}

func TestConditionalAgent_Priority(t *testing.T) {
	executionOrder := make([]string, 0)

	// Create test steps that record execution order
	step1 := &mockStep{
		name: "low-priority",
		exec: func(ctx context.Context, state *WorkflowState) (*WorkflowState, error) {
			executionOrder = append(executionOrder, "low")
			newState := state.State.Clone()
			newState.Set("executed", append(getExecuted(newState), "low"))
			return &WorkflowState{State: newState, Metadata: make(map[string]interface{})}, nil
		},
	}

	step2 := &mockStep{
		name: "high-priority",
		exec: func(ctx context.Context, state *WorkflowState) (*WorkflowState, error) {
			executionOrder = append(executionOrder, "high")
			newState := state.State.Clone()
			newState.Set("executed", append(getExecuted(newState), "high"))
			return &WorkflowState{State: newState, Metadata: make(map[string]interface{})}, nil
		},
	}

	// Create conditional workflow with priorities
	conditional := NewConditionalAgent("priority-test").
		WithAllowMultipleMatches(true). // Allow both to execute
		AddBranchWithPriority("low", func(state *domain.State) bool { return true }, step1, 1).
		AddBranchWithPriority("high", func(state *domain.State) bool { return true }, step2, 10)

	ctx := context.Background()
	state := domain.NewState()
	state.Set("executed", make([]string, 0))

	result, err := conditional.Run(ctx, state)
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	// High priority should execute first
	if len(executionOrder) != 2 || executionOrder[0] != "high" || executionOrder[1] != "low" {
		t.Errorf("Expected execution order [high, low], got: %v", executionOrder)
	}

	executed, _ := result.Get("executed")
	execList := executed.([]string)
	if len(execList) != 2 || execList[0] != "high" || execList[1] != "low" {
		t.Errorf("Expected result execution order [high, low], got: %v", execList)
	}
}

func TestConditionalAgent_MultipleMatches(t *testing.T) {
	// Create test agents
	agent1 := createMockAgentWithResponse("agent1", "result1")
	agent2 := createMockAgentWithResponse("agent2", "result2")

	// Create conditional workflow that allows multiple matches
	conditional := NewConditionalAgent("multi-match").
		WithAllowMultipleMatches(true).
		AddAgent("branch1", func(state *domain.State) bool {
			return true // Always matches
		}, agent1).
		AddAgent("branch2", func(state *domain.State) bool {
			return true // Always matches
		}, agent2)

	ctx := context.Background()
	state := domain.NewState()

	result, err := conditional.Run(ctx, state)
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	// The final result should be from the last executed branch (agent2)
	if response, exists := result.Get("response"); !exists || response != "result2" {
		t.Errorf("Expected 'result2', got: %v", response)
	}
}

func TestConditionalAgent_NoMatches(t *testing.T) {
	// Create test agents
	agent1 := createMockAgentWithResponse("agent1", "result1")

	// Create conditional workflow without default
	conditional := NewConditionalAgent("no-matches").
		AddAgent("branch1", func(state *domain.State) bool {
			return false // Never matches
		}, agent1)

	ctx := context.Background()
	state := domain.NewState()

	result, err := conditional.Run(ctx, state)
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	// Should return original state since no branches executed
	if response, exists := result.Get("response"); exists {
		t.Errorf("Expected no response key, but got: %v", response)
	}
}

func TestConditionalAgent_ErrorHandling(t *testing.T) {
	// Create agent that returns an error
	errorAgent := &mockAgent{
		BaseAgent: createBaseAgent("error-agent", "Error agent", domain.AgentTypeCustom),
		shouldErr: true,
		response:  "should not see this",
	}

	// Create conditional workflow
	conditional := NewConditionalAgent("error-test").
		AddAgent("error-branch", func(state *domain.State) bool {
			return true // Always matches
		}, errorAgent)

	ctx := context.Background()
	state := domain.NewState()

	result, err := conditional.Run(ctx, state)
	if err == nil {
		t.Fatal("Expected error but got none")
	}

	if result != nil {
		t.Error("Expected nil result on error")
	}

	expectedError := "branch error-branch failed: mock error from error-agent"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: %s", expectedError, err.Error())
	}
}

func TestConditionalAgent_Validation(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *ConditionalAgent
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid workflow",
			setup: func() *ConditionalAgent {
				agent := createMockAgentWithResponse("test", "result")
				return NewConditionalAgent("valid").
					AddAgent("branch1", func(state *domain.State) bool { return true }, agent)
			},
			expectError: false,
		},
		{
			name: "No branches or default",
			setup: func() *ConditionalAgent {
				return NewConditionalAgent("empty")
			},
			expectError: true,
			errorMsg:    "conditional workflow must have at least one branch or a default branch",
		},
		{
			name: "Branch with nil condition",
			setup: func() *ConditionalAgent {
				agent := createMockAgentWithResponse("test", "result")
				conditional := NewConditionalAgent("nil-condition")
				// Manually add branch with nil condition
				conditional.branches = append(conditional.branches, ConditionalBranch{
					Name:      "nil-condition",
					Condition: nil,
					Step:      &AgentStep{name: "test", agent: agent},
				})
				return conditional
			},
			expectError: true,
			errorMsg:    "branch nil-condition has nil condition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditional := tt.setup()
			err := conditional.Validate()

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

func TestConditionalAgent_Metadata(t *testing.T) {
	// Create test agents
	agent1 := createMockAgentWithResponse("agent1", "result1")
	agent2 := createMockAgentWithResponse("agent2", "result2")

	// Create conditional workflow
	conditional := NewConditionalAgent("metadata-test").
		AddAgent("branch1", func(state *domain.State) bool {
			if value, exists := state.Get("execute"); exists {
				return value == "branch1"
			}
			return false
		}, agent1).
		AddAgent("branch2", func(state *domain.State) bool {
			return false // Never matches
		}, agent2)

	ctx := context.Background()
	state := domain.NewState()
	state.Set("execute", "branch1")

	// Run workflow
	result, err := conditional.Run(ctx, state)
	if err != nil {
		t.Fatalf("Workflow failed: %v", err)
	}

	// Check that we got a result
	if result == nil {
		t.Fatal("Expected result but got nil")
	}

	// Check metadata is present (this is set in the workflow state metadata, not the domain state)
	// We need to check the workflow status or other indicators
	status := conditional.Status()
	if status.State != WorkflowStateCompleted {
		t.Errorf("Expected workflow state completed, got: %v", status.State)
	}

	// Check that branch1 step was executed
	if stepStatus, exists := status.Steps["branch1"]; !exists || stepStatus.State != StepStateCompleted {
		t.Errorf("Expected branch1 to be completed")
	}

	// Check that branch2 step was not executed
	if stepStatus, exists := status.Steps["branch2"]; exists && stepStatus.State != StepStatePending {
		t.Errorf("Expected branch2 to not be executed, but status is: %v", stepStatus.State)
	}
}

// Helper functions

func getExecuted(state *domain.State) []string {
	if exec, exists := state.Get("executed"); exists {
		// Handle both []string and []interface{} cases
		switch v := exec.(type) {
		case []string:
			return v
		case []interface{}:
			result := make([]string, len(v))
			for i, item := range v {
				if s, ok := item.(string); ok {
					result[i] = s
				}
			}
			return result
		}
	}
	return make([]string, 0)
}

type mockStep struct {
	name string
	exec func(ctx context.Context, state *WorkflowState) (*WorkflowState, error)
}

func (m *mockStep) Name() string {
	return m.name
}

func (m *mockStep) Execute(ctx context.Context, state *WorkflowState) (*WorkflowState, error) {
	return m.exec(ctx, state)
}

func (m *mockStep) Validate() error {
	if m.exec == nil {
		return fmt.Errorf("mock step has nil exec function")
	}
	return nil
}

func createMockAgentWithResponse(name, response string) domain.BaseAgent {
	return &mockAgent{
		BaseAgent: createBaseAgent(name, "Mock agent", domain.AgentTypeCustom),
		response:  response,
	}
}

// Helper function to create base agent
func createBaseAgent(name, description string, agentType domain.AgentType) domain.BaseAgent {
	return core.NewBaseAgent(name, description, agentType)
}

// mockAgent for testing
type mockAgent struct {
	domain.BaseAgent
	response  string
	shouldErr bool
}

func (m *mockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	if m.shouldErr {
		return nil, fmt.Errorf("mock error from %s", m.Name())
	}

	newState := state.Clone()
	newState.Set("response", m.response)
	return newState, nil
}
