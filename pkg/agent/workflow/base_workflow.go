// ABOUTME: Defines the WorkflowAgent interface for orchestrating multi-step processes and multi-agent interactions.
// ABOUTME: This interface extends BaseAgent with workflow-specific capabilities for complex task coordination.

package workflow

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// Note: WorkflowStep is defined in types.go as an interface

// Note: ErrorHandler is defined in types.go as an interface

// Note: WorkflowDefinition is defined in types.go

// Note: WorkflowState is defined in types.go as a concrete struct

// WorkflowAgent extends BaseAgent with workflow orchestration capabilities.
// It provides advanced features for executing complex multi-step workflows,
// managing agent coordination, and handling workflow lifecycle operations.
type WorkflowAgent interface {
	domain.BaseAgent

	// ExecuteWorkflow executes a workflow definition with the given initial state
	ExecuteWorkflow(ctx context.Context, workflow *WorkflowDefinition, initialState *domain.State) (*domain.State, error)

	// RegisterAgent registers an agent that can be used in workflow steps
	RegisterAgent(id string, agent domain.BaseAgent) error

	// GetRegisteredAgent retrieves a registered agent by ID
	GetRegisteredAgent(id string) (domain.BaseAgent, bool)

	// SetStepTimeout sets the timeout for individual workflow steps
	SetStepTimeout(stepName string, timeout int) error

	// GetWorkflowStatus returns the current status of workflow execution
	GetWorkflowStatus() WorkflowStatus

	// PauseWorkflow pauses the current workflow execution
	PauseWorkflow() error

	// ResumeWorkflow resumes a paused workflow
	ResumeWorkflow() error

	// CancelWorkflow cancels the current workflow execution
	CancelWorkflow() error
}

// Note: WorkflowStatus is defined in types.go

// BranchingWorkflowAgent extends WorkflowAgent with conditional branching
type BranchingWorkflowAgent interface {
	WorkflowAgent

	// AddBranch adds a conditional branch to the workflow
	AddBranch(condition func(state *domain.State) bool, branch *WorkflowDefinition) error

	// AddLoop adds a loop construct to the workflow
	AddLoop(condition func(state *domain.State) bool, loopSteps []WorkflowStep) error
}

// WorkflowOrchestrator manages multiple workflow executions
type WorkflowOrchestrator interface {
	// ExecuteWorkflows executes multiple workflows with coordination
	ExecuteWorkflows(ctx context.Context, workflows []*WorkflowDefinition, coordination CoordinationStrategy) ([]*domain.State, error)

	// GetWorkflowResults retrieves results from all executed workflows
	GetWorkflowResults() map[string]*domain.State

	// SetCoordinationStrategy updates how workflows are coordinated
	SetCoordinationStrategy(strategy CoordinationStrategy)
}

// CoordinationStrategy defines how multiple workflows are coordinated
type CoordinationStrategy struct {
	// Type specifies the coordination type (sequential, parallel, dependent)
	Type string

	// Dependencies maps workflow names to their dependencies
	Dependencies map[string][]string

	// SharedStateHandler manages state sharing between workflows
	SharedStateHandler func(states map[string]*domain.State) *domain.State
}
