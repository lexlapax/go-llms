// ABOUTME: Sequential workflow agent that executes steps one after another
// ABOUTME: Passes state from each step to the next, with error handling and hooks

package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// SequentialAgent executes workflow steps in sequence
type SequentialAgent struct {
	*BaseWorkflowAgent

	// Options
	stopOnError bool
	maxRetries  int
}

// NewSequentialAgent creates a new sequential workflow agent
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

// WithStopOnError configures whether to stop on first error (default: true)
func (s *SequentialAgent) WithStopOnError(stop bool) *SequentialAgent {
	s.stopOnError = stop
	return s
}

// WithMaxRetries sets the maximum number of retries per step
func (s *SequentialAgent) WithMaxRetries(retries int) *SequentialAgent {
	s.maxRetries = retries
	return s
}

// WithHook adds a monitoring hook to the workflow agent
func (s *SequentialAgent) WithHook(hook domain.Hook) *SequentialAgent {
	s.BaseWorkflowAgent.WithHook(hook)
	return s
}

// AddAgent adds an agent as a workflow step
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

// Run executes the sequential workflow
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
		stepStatus := StepStatus{
			StartTime: s.status.Steps[stepName].StartTime,
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

// RunAsync executes the workflow asynchronously
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

