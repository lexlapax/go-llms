// ABOUTME: Base implementation of WorkflowAgent providing common workflow functionality
// ABOUTME: Embeds BaseAgentImpl and adds workflow-specific behavior for step execution

package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// BaseWorkflowAgent provides common functionality for all workflow agents.
// It extends BaseAgentImpl with workflow-specific features including step
// management, status tracking, error handling, and hooks. This type serves
// as the foundation for all concrete workflow implementations.
type BaseWorkflowAgent struct {
	*core.BaseAgentImpl

	steps        []WorkflowStep
	errorHandler ErrorHandler
	workflowDef  *WorkflowDefinition
	status       *WorkflowStatus
	hooks        []domain.Hook
	mu           sync.RWMutex
}

// NewBaseWorkflowAgent creates a new base workflow agent.
// It initializes the workflow with empty steps and pending status.
//
// Parameters:
//   - name: The workflow name
//   - description: The workflow description
//   - agentType: The specific type of workflow agent
//
// Returns a new BaseWorkflowAgent instance.
func NewBaseWorkflowAgent(name, description string, agentType domain.AgentType) *BaseWorkflowAgent {
	return &BaseWorkflowAgent{
		BaseAgentImpl: core.NewBaseAgent(name, description, agentType),
		steps:         make([]WorkflowStep, 0),
		hooks:         make([]domain.Hook, 0),
		status: &WorkflowStatus{
			State:     WorkflowStatePending,
			StartTime: time.Time{},
			Steps:     make(map[string]StepStatus),
		},
	}
}

// WithHook adds a monitoring hook to the workflow agent.
// Hooks are notified before and after workflow execution.
// Nil hooks are ignored.
//
// Parameters:
//   - hook: The hook to add
//
// Returns the BaseWorkflowAgent for method chaining.
func (w *BaseWorkflowAgent) WithHook(hook domain.Hook) *BaseWorkflowAgent {
	if hook == nil {
		return w
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.hooks = append(w.hooks, hook)
	return w
}

// AddStep adds a step to the workflow.
// Steps must have unique names within the workflow.
//
// Parameters:
//   - step: The workflow step to add
//
// Returns an error if the step is nil or has a duplicate name.
func (w *BaseWorkflowAgent) AddStep(step WorkflowStep) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if step == nil {
		return fmt.Errorf("step cannot be nil")
	}

	// Validate step doesn't already exist
	for _, existing := range w.steps {
		if existing.Name() == step.Name() {
			return fmt.Errorf("step with name %s already exists", step.Name())
		}
	}

	w.steps = append(w.steps, step)
	return nil
}

// Steps returns all workflow steps.
// Returns a copy to prevent external modifications.
//
// Returns a slice of all workflow steps.
func (w *BaseWorkflowAgent) Steps() []WorkflowStep {
	w.mu.RLock()
	defer w.mu.RUnlock()

	steps := make([]WorkflowStep, len(w.steps))
	copy(steps, w.steps)
	return steps
}

// SetErrorHandler sets the error handler for the workflow.
// The error handler determines how the workflow responds to step failures.
//
// Parameters:
//   - handler: The error handler to use
func (w *BaseWorkflowAgent) SetErrorHandler(handler ErrorHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.errorHandler = handler
}

// WorkflowDefinition returns the workflow definition.
// Returns nil if no definition has been set.
//
// Returns the current workflow definition.
func (w *BaseWorkflowAgent) WorkflowDefinition() *WorkflowDefinition {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.workflowDef
}

// SetWorkflowDefinition sets the workflow definition.
//
// Parameters:
//   - def: The workflow definition to set
//
// Returns an error if the definition is nil.
func (w *BaseWorkflowAgent) SetWorkflowDefinition(def *WorkflowDefinition) error {
	if def == nil {
		return fmt.Errorf("workflow definition cannot be nil")
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.workflowDef = def
	return nil
}

// Status returns the current workflow status.
// Returns a copy to prevent external modifications.
//
// Returns the current workflow execution status.
func (w *BaseWorkflowAgent) Status() *WorkflowStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Return a copy to prevent external modification
	statusCopy := &WorkflowStatus{
		State:       w.status.State,
		StartTime:   w.status.StartTime,
		EndTime:     w.status.EndTime,
		CurrentStep: w.status.CurrentStep,
		Error:       w.status.Error,
		Steps:       make(map[string]StepStatus),
	}

	for k, v := range w.status.Steps {
		statusCopy.Steps[k] = v
	}

	return statusCopy
}

// updateStatus updates the workflow status
func (w *BaseWorkflowAgent) updateStatus(state WorkflowStateType, currentStep string, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.status.State = state
	w.status.CurrentStep = currentStep
	w.status.Error = err

	if state == WorkflowStateRunning && w.status.StartTime.IsZero() {
		w.status.StartTime = time.Now()
	} else if state == WorkflowStateCompleted || state == WorkflowStateFailed {
		w.status.EndTime = time.Now()
	}
}

// updateStepStatus updates the status of a specific step
func (w *BaseWorkflowAgent) updateStepStatus(stepName string, status StepStatus) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.status.Steps == nil {
		w.status.Steps = make(map[string]StepStatus)
	}

	w.status.Steps[stepName] = status
}

// createWorkflowState creates a WorkflowState from domain.State
func (w *BaseWorkflowAgent) createWorkflowState(state *domain.State) *WorkflowState {
	return &WorkflowState{
		State:    state,
		Metadata: make(map[string]interface{}),
	}
}

// handleStepError handles errors from step execution
func (w *BaseWorkflowAgent) handleStepError(ctx context.Context, step WorkflowStep, state *WorkflowState, err error) (*WorkflowState, error) {
	if w.errorHandler == nil {
		return state, err
	}

	action := w.errorHandler.HandleError(ctx, step, state, err)

	switch action {
	case ErrorActionRetry:
		// Retry the step
		return step.Execute(ctx, state)
	case ErrorActionSkip:
		// Skip this step and continue
		return state, nil
	case ErrorActionAbort:
		// Abort the workflow
		return state, fmt.Errorf("workflow aborted due to step error: %w", err)
	default:
		return state, err
	}
}

// emitWorkflowEvent emits a workflow-specific event
func (w *BaseWorkflowAgent) emitWorkflowEvent(eventType domain.EventType, data interface{}) {
	w.EmitEvent(eventType, data)
}

// Validate validates the workflow configuration.
// It ensures the workflow has at least one step and all steps
// are valid with non-empty names.
//
// Returns an error if validation fails.
func (w *BaseWorkflowAgent) Validate() error {
	if err := w.BaseAgentImpl.Validate(); err != nil {
		return err
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	if len(w.steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	// Validate each step
	for i, step := range w.steps {
		if step == nil {
			return fmt.Errorf("step %d is nil", i)
		}
		if step.Name() == "" {
			return fmt.Errorf("step %d has empty name", i)
		}
	}

	return nil
}

// Clone creates a copy of the workflow agent.
// The clone has the same configuration but reset status.
// Steps and handlers are shared (not deep copied).
//
// Returns a new BaseWorkflowAgent instance.
func (w *BaseWorkflowAgent) Clone() *BaseWorkflowAgent {
	w.mu.RLock()
	defer w.mu.RUnlock()

	clone := &BaseWorkflowAgent{
		BaseAgentImpl: w.BaseAgentImpl,
		steps:         make([]WorkflowStep, len(w.steps)),
		errorHandler:  w.errorHandler,
		workflowDef:   w.workflowDef,
		status: &WorkflowStatus{
			State:     WorkflowStatePending,
			StartTime: time.Time{},
			Steps:     make(map[string]StepStatus),
		},
	}

	copy(clone.steps, w.steps)
	return clone
}

// OnEvent registers an event handler using the BaseAgentImpl functionality.
// This allows monitoring workflow events during execution.
//
// Parameters:
//   - handler: The event handler function
//
// Returns a subscription ID that can be used to unsubscribe.
func (w *BaseWorkflowAgent) OnEvent(handler func(event *domain.Event)) string {
	return w.BaseAgentImpl.OnEvent(handler)
}

// BeforeRun overrides BaseAgentImpl to call hooks before workflow execution.
// It first calls the parent implementation, then notifies all registered hooks.
//
// Parameters:
//   - ctx: The execution context
//   - state: The initial state
//
// Returns an error if the parent BeforeRun fails.
func (w *BaseWorkflowAgent) BeforeRun(ctx context.Context, state *domain.State) error {
	// Call parent implementation first
	if err := w.BaseAgentImpl.BeforeRun(ctx, state); err != nil {
		return err
	}

	// Notify hooks
	w.notifyHooksBeforeRun(ctx, state)
	return nil
}

// AfterRun overrides BaseAgentImpl to call hooks after workflow execution.
// It first notifies all registered hooks, then calls the parent implementation.
//
// Parameters:
//   - ctx: The execution context
//   - state: The initial state
//   - result: The result state (may be nil on error)
//   - err: Any error that occurred during execution
//
// Returns an error from the parent AfterRun.
func (w *BaseWorkflowAgent) AfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) error {
	// Notify hooks first
	w.notifyHooksAfterRun(ctx, state, result, err)

	// Call parent implementation
	return w.BaseAgentImpl.AfterRun(ctx, state, result, err)
}

// notifyHooksBeforeRun calls all hooks' BeforeRun method
func (w *BaseWorkflowAgent) notifyHooksBeforeRun(ctx context.Context, state *domain.State) {
	w.mu.RLock()
	hooks := w.hooks
	w.mu.RUnlock()

	for _, hook := range hooks {
		// Workflow agents implement BaseAgent interface, so we can call BeforeRun
		if beforeRunHook, ok := hook.(interface {
			BeforeRun(ctx context.Context, agent domain.BaseAgent, state *domain.State) (context.Context, error)
		}); ok {
			if _, err := beforeRunHook.BeforeRun(ctx, w, state); err != nil {
				// Log the hook error but continue execution
				// In practice, this should be logged properly
				_ = err // Explicitly ignore the error for linting
			}
		}
	}
}

// notifyHooksAfterRun calls all hooks' AfterRun method
func (w *BaseWorkflowAgent) notifyHooksAfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) {
	w.mu.RLock()
	hooks := w.hooks
	w.mu.RUnlock()

	for _, hook := range hooks {
		// Workflow agents implement BaseAgent interface, so we can call AfterRun
		if afterRunHook, ok := hook.(interface {
			AfterRun(ctx context.Context, agent domain.BaseAgent, state *domain.State, result *domain.State, err error) error
		}); ok {
			if hookErr := afterRunHook.AfterRun(ctx, w, state, result, err); hookErr != nil {
				// Log the hook error but continue execution
				// In practice, this should be logged properly
				_ = hookErr // Explicitly ignore the error for linting
			}
		}
	}
}
