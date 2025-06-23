// ABOUTME: Type definitions for workflow agents including states, steps, and error handling
// ABOUTME: Provides the core types used across all workflow implementations

package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// WorkflowStateType represents the state of a workflow.
// It tracks the lifecycle of workflow execution from pending to completion.
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

// StepStateType represents the state of a workflow step.
// It tracks individual step execution within a workflow.
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

// ErrorAction defines what to do when an error occurs.
// It determines the workflow's response to step failures.
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

// WorkflowStep interface for workflow steps.
// Implementations define discrete units of work that can be
// executed as part of a workflow.
type WorkflowStep interface {
	// Name returns the step name
	Name() string
	// Execute runs the step
	Execute(ctx context.Context, state *WorkflowState) (*WorkflowState, error)
	// Validate checks if the step is valid
	Validate() error
}

// WorkflowState wraps domain.State with workflow metadata.
// It extends the base state with workflow-specific information
// needed during execution.
type WorkflowState struct {
	*domain.State
	Metadata map[string]interface{}
}

// StepStatus tracks the status of a workflow step.
// It records execution details including timing, errors, and retry attempts.
type StepStatus struct {
	State     StepStateType
	StartTime time.Time
	EndTime   time.Time
	Error     error
	Retries   int
}

// ErrorHandler interface for handling step errors.
// Implementations determine how workflows respond to failures,
// including retry logic and error recovery strategies.
type ErrorHandler interface {
	// HandleError processes an error and returns the action to take
	HandleError(ctx context.Context, step WorkflowStep, state *WorkflowState, err error) ErrorAction
}

// DefaultErrorHandler provides basic error handling.
// It implements a simple abort-on-error strategy with
// configurable retry behavior.
type DefaultErrorHandler struct {
	MaxRetries int
	RetryDelay time.Duration
}

// HandleError implements ErrorHandler.
// Currently returns ErrorActionAbort for all errors.
// This can be extended to implement more sophisticated error handling.
func (h *DefaultErrorHandler) HandleError(ctx context.Context, step WorkflowStep, state *WorkflowState, err error) ErrorAction {
	// Simple implementation - can be extended
	return ErrorActionAbort
}

// WorkflowStatus represents the current state of workflow execution.
// It provides a complete view of the workflow's progress, including
// individual step statuses and timing information.
type WorkflowStatus struct {
	State       WorkflowStateType
	StartTime   time.Time
	EndTime     time.Time
	CurrentStep string
	Error       error
	Steps       map[string]StepStatus
}

// WorkflowDefinition defines the structure and flow of a workflow.
// It specifies the steps to execute and how they should be coordinated
// (sequential or parallel execution).
type WorkflowDefinition struct {
	Name           string
	Description    string
	Steps          []WorkflowStep
	Parallel       bool
	MaxConcurrency int
}

// AgentStep wraps an agent as a workflow step.
// This allows any BaseAgent to be used as a step in a workflow,
// providing seamless integration between agents and workflows.
type AgentStep struct {
	name  string
	agent domain.BaseAgent
}

// NewAgentStep creates a new AgentStep from a BaseAgent.
//
// Parameters:
//   - name: The name for this step
//   - agent: The agent to wrap as a step
//
// Returns a WorkflowStep that executes the agent.
func NewAgentStep(name string, agent domain.BaseAgent) WorkflowStep {
	return &AgentStep{
		name:  name,
		agent: agent,
	}
}

// Name returns the step name.
// Implements WorkflowStep.Name.
func (s *AgentStep) Name() string {
	return s.name
}

// Execute runs the agent.
// It executes the wrapped agent with the workflow state and returns
// a new workflow state with the agent's results and updated metadata.
// Implements WorkflowStep.Execute.
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

// Validate validates the step.
// It ensures the wrapped agent is not nil and is itself valid.
// Implements WorkflowStep.Validate.
func (s *AgentStep) Validate() error {
	if s.agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}
	return s.agent.Validate()
}
