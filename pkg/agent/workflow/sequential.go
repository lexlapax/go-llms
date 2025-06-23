// ABOUTME: Sequential workflow agent that executes steps one after another
// ABOUTME: Passes state from each step to the next, with error handling and hooks

package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// SequentialAgent executes workflow steps in sequence.
// It runs each step one after another, passing the output state from
// each step as input to the next. Supports error handling, retries,
// and conditional execution based on error behavior.
type SequentialAgent struct {
	*BaseWorkflowAgent

	// Options
	stopOnError bool
	maxRetries  int
}

// NewSequentialAgent creates a new sequential workflow agent.
// By default, it stops on the first error and does not retry failed steps.
//
// Parameters:
//   - name: The name of the sequential workflow
//
// Returns a new SequentialAgent instance.
func NewSequentialAgent(name string) *SequentialAgent {
	return &SequentialAgent{
		BaseWorkflowAgent: NewBaseWorkflowAgent(
			name,
			fmt.Sprintf("Sequential workflow: %s", name),
			domain.AgentTypeSequential,
		),
		stopOnError: true,
		maxRetries:  0,
	}
}

// WithStopOnError configures whether to stop on first error.
// When true (default), the workflow stops at the first failed step.
// When false, the workflow continues with subsequent steps even if one fails.
//
// Parameters:
//   - stop: If true, stop on first error
//
// Returns the SequentialAgent for method chaining.
func (s *SequentialAgent) WithStopOnError(stop bool) *SequentialAgent {
	s.stopOnError = stop
	return s
}

// WithMaxRetries sets the maximum number of retries per step.
// Failed steps will be retried up to this many times before
// being considered permanently failed.
//
// Parameters:
//   - retries: Maximum retry attempts (0 = no retries)
//
// Returns the SequentialAgent for method chaining.
func (s *SequentialAgent) WithMaxRetries(retries int) *SequentialAgent {
	s.maxRetries = retries
	return s
}

// WithHook adds a monitoring hook to the workflow agent.
// Hooks allow monitoring and customization of sequential execution.
//
// Parameters:
//   - hook: The hook to add
//
// Returns the SequentialAgent for method chaining.
func (s *SequentialAgent) WithHook(hook domain.Hook) *SequentialAgent {
	s.BaseWorkflowAgent.WithHook(hook)
	return s
}

// AddAgent adds an agent as a workflow step.
// This is a convenience method that wraps the agent in an AgentStep.
//
// Parameters:
//   - agent: The agent to add to the sequence
//
// Returns the SequentialAgent for method chaining.
func (s *SequentialAgent) AddAgent(agent domain.BaseAgent) *SequentialAgent {
	step := &AgentStep{
		name:  agent.Name(),
		agent: agent,
	}
	if err := s.AddStep(step); err != nil {
		// Log error but continue for fluent interface
		// In practice, this error should be very rare
		_ = err // Explicitly ignore the error for linting
	}
	return s
}

// Run executes the sequential workflow.
// It runs each step in order, passing the output of each step as input
// to the next. Execution stops on error if configured to do so.
// Steps may be retried based on the retry configuration.
//
// Parameters:
//   - ctx: The execution context
//   - input: The initial state
//
// Returns the final state after all steps or an error.
func (s *SequentialAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	// Validate before running
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Update status to running
	s.updateStatus(WorkflowStateRunning, "", nil)
	s.emitWorkflowEvent(domain.EventWorkflowStart, map[string]interface{}{
		"steps": len(s.steps),
	})

	// Create workflow state
	workflowState := s.createWorkflowState(input)

	// Execute before hooks
	if err := s.BeforeRun(ctx, input); err != nil {
		s.updateStatus(WorkflowStateFailed, "", err)
		return nil, err
	}

	// Execute each step in sequence
	for i, step := range s.steps {
		stepName := step.Name()

		// Update current step
		s.updateStatus(WorkflowStateRunning, stepName, nil)
		s.updateStepStatus(stepName, StepStatus{
			State:     StepStateRunning,
			StartTime: time.Now(),
		})

		// Emit step start event
		s.emitWorkflowEvent(domain.EventWorkflowStep, map[string]interface{}{
			"step":  stepName,
			"index": i,
			"total": len(s.steps),
		})

		// Execute step with retries
		var stepErr error
		var result *WorkflowState

		for attempt := 0; attempt <= s.maxRetries; attempt++ {
			if attempt > 0 {
				s.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
					"step":    stepName,
					"retry":   attempt,
					"message": fmt.Sprintf("Retrying step %s (attempt %d/%d)", stepName, attempt, s.maxRetries),
				})
			}

			result, stepErr = step.Execute(ctx, workflowState)
			if stepErr == nil {
				break
			}

			// Handle error
			if s.errorHandler != nil {
				result, stepErr = s.handleStepError(ctx, step, workflowState, stepErr)
				if stepErr == nil {
					break
				}
			}
		}

		// Update step status
		// Get start time with lock
		s.mu.RLock()
		startTime := s.status.Steps[stepName].StartTime
		s.mu.RUnlock()

		stepStatus := StepStatus{
			StartTime: startTime,
			EndTime:   time.Now(),
		}

		if stepErr != nil {
			stepStatus.State = StepStateFailed
			stepStatus.Error = stepErr
			s.updateStepStatus(stepName, stepStatus)

			// Emit error event
			s.emitWorkflowEvent(domain.EventAgentError, map[string]interface{}{
				"step":  stepName,
				"error": stepErr.Error(),
			})

			if s.stopOnError {
				s.updateStatus(WorkflowStateFailed, stepName, stepErr)
				if err := s.AfterRun(ctx, input, nil, stepErr); err != nil {
					// Log the after-run error but still return the original error
					// In practice, this should be logged properly
					_ = err // Explicitly ignore the error for linting
				}
				return nil, fmt.Errorf("step %s failed: %w", stepName, stepErr)
			}

			// Continue with current state if not stopping on error
			continue
		}

		stepStatus.State = StepStateCompleted
		s.updateStepStatus(stepName, stepStatus)

		// Update workflow state for next step
		if result != nil {
			workflowState = result
		}

		// Emit step complete event
		s.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
			"step":     stepName,
			"index":    i + 1,
			"total":    len(s.steps),
			"complete": true,
		})
	}

	// Update final status
	s.updateStatus(WorkflowStateCompleted, "", nil)

	// Execute after hooks
	finalState := workflowState.State
	if err := s.AfterRun(ctx, input, finalState, nil); err != nil {
		return finalState, err
	}

	// Emit workflow complete event
	s.emitWorkflowEvent(domain.EventAgentComplete, map[string]interface{}{
		"duration": time.Since(s.status.StartTime),
		"steps":    len(s.steps),
	})

	return finalState, nil
}

// RunAsync executes the workflow asynchronously.
// It returns a channel that emits events as the workflow progresses,
// including step completions and the final result. The workflow
// execution happens in a separate goroutine.
//
// Parameters:
//   - ctx: The execution context
//   - input: The initial state
//
// Returns an event channel and any validation error.
func (s *SequentialAgent) RunAsync(ctx context.Context, input *domain.State) (<-chan domain.Event, error) {
	// Validate before running
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Create event channel
	eventChan := make(chan domain.Event, 100)

	// Subscribe to internal events
	subscriptionID := s.OnEvent(func(event *domain.Event) {
		select {
		case eventChan <- *event:
		case <-ctx.Done():
		}
	})

	// Run workflow in goroutine
	go func() {
		defer close(eventChan)
		defer s.Unsubscribe(subscriptionID) // Clean up subscription

		result, err := s.Run(ctx, input)

		// Send final event
		finalEvent := domain.Event{
			Type:      domain.EventAgentComplete,
			AgentID:   s.ID(),
			AgentName: s.Name(),
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"result": result,
				"error":  err,
			},
		}

		if err != nil {
			finalEvent.Type = domain.EventAgentError
			finalEvent.Error = err
		}

		select {
		case eventChan <- finalEvent:
		case <-ctx.Done():
		}
	}()

	return eventChan, nil
}
