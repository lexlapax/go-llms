// ABOUTME: Type definitions for workflow agents including states, steps, and error handling
// ABOUTME: Provides the core types used across all workflow implementations

package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// WorkflowStateType represents the state of a workflow
type WorkflowStateType string

const (
	// WorkflowStatePending indicates the workflow hasn't started
	WorkflowStatePending WorkflowStateType = "pending"
	// WorkflowStateRunning indicates the workflow is currently executing
	WorkflowStateRunning WorkflowStateType = "running"
	// WorkflowStatePaused indicates the workflow is paused
	WorkflowStatePaused WorkflowStateType = "paused"
	// WorkflowStateCompleted indicates the workflow completed successfully
	WorkflowStateCompleted WorkflowStateType = "completed"
	// WorkflowStateFailed indicates the workflow failed
	WorkflowStateFailed WorkflowStateType = "failed"
	// WorkflowStateCanceled indicates the workflow was canceled
	WorkflowStateCanceled WorkflowStateType = "canceled"
)

// StepStateType represents the state of a workflow step
type StepStateType string

const (
	// StepStatePending indicates the step hasn't started
	StepStatePending StepStateType = "pending"
	// StepStateRunning indicates the step is currently executing
	StepStateRunning StepStateType = "running"
	// StepStateCompleted indicates the step completed successfully
	StepStateCompleted StepStateType = "completed"
	// StepStateFailed indicates the step failed
	StepStateFailed StepStateType = "failed"
	// StepStateSkipped indicates the step was skipped
	StepStateSkipped StepStateType = "skipped"
)

// ErrorAction defines what to do when an error occurs
type ErrorAction string

const (
	// ErrorActionRetry indicates the step should be retried
	ErrorActionRetry ErrorAction = "retry"
	// ErrorActionSkip indicates the step should be skipped
	ErrorActionSkip ErrorAction = "skip"
	// ErrorActionAbort indicates the workflow should be aborted
	ErrorActionAbort ErrorAction = "abort"
	// ErrorActionContinue indicates the workflow should continue
	ErrorActionContinue ErrorAction = "continue"
)

// WorkflowStep interface for workflow steps
type WorkflowStep interface {
	// Name returns the step name
	Name() string
	// Execute runs the step
	Execute(ctx context.Context, state *WorkflowState) (*WorkflowState, error)
	// Validate checks if the step is valid
	Validate() error
}

// WorkflowState wraps domain.State with workflow metadata
type WorkflowState struct {
	*domain.State
	Metadata map[string]interface{}
}

// StepStatus tracks the status of a workflow step
type StepStatus struct {
	State     StepStateType
	StartTime time.Time
	EndTime   time.Time
	Error     error
	Retries   int
}

// ErrorHandler interface for handling step errors
type ErrorHandler interface {
	// HandleError processes an error and returns the action to take
	HandleError(ctx context.Context, step WorkflowStep, state *WorkflowState, err error) ErrorAction
}

// DefaultErrorHandler provides basic error handling
type DefaultErrorHandler struct {
	MaxRetries int
	RetryDelay time.Duration
}

// HandleError implements ErrorHandler
func (h *DefaultErrorHandler) HandleError(ctx context.Context, step WorkflowStep, state *WorkflowState, err error) ErrorAction {
	// Simple implementation - can be extended
	return ErrorActionAbort
}

// WorkflowStatus represents the current state of workflow execution
type WorkflowStatus struct {
	State       WorkflowStateType
	StartTime   time.Time
	EndTime     time.Time
	CurrentStep string
	Error       error
	Steps       map[string]StepStatus
}

// WorkflowDefinition defines the structure and flow of a workflow
type WorkflowDefinition struct {
	Name           string
	Description    string
	Steps          []WorkflowStep
	Parallel       bool
	MaxConcurrency int
}

// AgentStep wraps an agent as a workflow step
type AgentStep struct {
	name  string
	agent domain.BaseAgent
}

// NewAgentStep creates a new AgentStep from a BaseAgent
func NewAgentStep(name string, agent domain.BaseAgent) WorkflowStep {
	return &AgentStep{
		name:  name,
		agent: agent,
	}
}

// Name returns the step name
func (s *AgentStep) Name() string {
	return s.name
}

// Execute runs the agent
func (s *AgentStep) Execute(ctx context.Context, state *WorkflowState) (*WorkflowState, error) {
	// Run the agent with the domain state
	result, err := s.agent.Run(ctx, state.State)
	if err != nil {
		return state, err
	}

	// Create new workflow state with the result
	newWorkflowState := &WorkflowState{
		State:    result,
		Metadata: make(map[string]interface{}),
	}

	// Copy workflow metadata to avoid concurrent map writes
	if state.Metadata != nil {
		for k, v := range state.Metadata {
			newWorkflowState.Metadata[k] = v
		}
	}

	// Add step execution metadata
	newWorkflowState.Metadata[fmt.Sprintf("step_%s_completed", s.name)] = time.Now()

	return newWorkflowState, nil
}

// Validate validates the step
func (s *AgentStep) Validate() error {
	if s.agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}
	return s.agent.Validate()
}
