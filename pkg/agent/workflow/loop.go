// ABOUTME: Loop workflow agent that executes iterative processing with conditions and counters
// ABOUTME: Provides for/while loop logic for workflows with support for break conditions and iteration limits

package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// LoopAgent executes workflow steps in a loop until a condition is met
type LoopAgent struct {
	*BaseWorkflowAgent

	// Loop body - the step(s) to execute in each iteration
	loopBody WorkflowStep

	// Loop conditions
	continueCondition func(state *domain.State, iteration int) bool // Continue while this is true
	breakCondition    func(state *domain.State, iteration int) bool // Break when this is true

	// Loop limits
	maxIterations int           // Maximum number of iterations (0 = unlimited)
	maxDuration   time.Duration // Maximum total duration (0 = unlimited)

	// Loop behavior
	collectResults   bool          // If true, collects results from each iteration
	breakOnError     bool          // If true, breaks loop on first error
	passStateThrough bool          // If true, passes state from one iteration to the next
	iterationDelay   time.Duration // Delay between iterations

	// Loop metadata
	currentIteration int
	startTime        time.Time
	iterationResults []interface{}
}

// LoopType represents different types of loop behavior
type LoopType string

const (
	// LoopTypeWhile continues while condition is true
	LoopTypeWhile LoopType = "while"
	// LoopTypeUntil continues until condition is true
	LoopTypeUntil LoopType = "until"
	// LoopTypeCount executes a specific number of times
	LoopTypeCount LoopType = "count"
	// LoopTypeForever executes indefinitely (use with break conditions)
	LoopTypeForever LoopType = "forever"
)

// NewLoopAgent creates a new loop workflow agent
func NewLoopAgent(name string) *LoopAgent {
	return &LoopAgent{
		BaseWorkflowAgent: NewBaseWorkflowAgent(
			name,
			fmt.Sprintf("Loop workflow: %s", name),
			domain.AgentTypeLoop,
		),
		maxIterations:    0, // Unlimited by default
		maxDuration:      0, // Unlimited by default
		collectResults:   true,
		breakOnError:     true,
		passStateThrough: true,
		iterationResults: make([]interface{}, 0),
	}
}

// SetLoopBody sets the step to execute in each iteration
func (l *LoopAgent) SetLoopBody(step WorkflowStep) *LoopAgent {
	l.loopBody = step
	return l
}

// SetLoopAgent sets an agent as the loop body
func (l *LoopAgent) SetLoopAgent(agent domain.BaseAgent) *LoopAgent {
	step := &AgentStep{
		name:  fmt.Sprintf("loop-%s", agent.Name()),
		agent: agent,
	}
	l.loopBody = step
	return l
}

// WithWhileCondition sets a condition that continues the loop while true
func (l *LoopAgent) WithWhileCondition(condition func(state *domain.State, iteration int) bool) *LoopAgent {
	l.continueCondition = condition
	return l
}

// WithUntilCondition sets a condition that breaks the loop when true
func (l *LoopAgent) WithUntilCondition(condition func(state *domain.State, iteration int) bool) *LoopAgent {
	l.breakCondition = condition
	return l
}

// WithMaxIterations sets the maximum number of iterations
func (l *LoopAgent) WithMaxIterations(max int) *LoopAgent {
	l.maxIterations = max
	return l
}

// WithMaxDuration sets the maximum total duration for the loop
func (l *LoopAgent) WithMaxDuration(duration time.Duration) *LoopAgent {
	l.maxDuration = duration
	return l
}

// WithCollectResults configures whether to collect results from each iteration
func (l *LoopAgent) WithCollectResults(collect bool) *LoopAgent {
	l.collectResults = collect
	return l
}

// WithBreakOnError configures whether to break on first error
func (l *LoopAgent) WithBreakOnError(breakOnErr bool) *LoopAgent {
	l.breakOnError = breakOnErr
	return l
}

// WithPassStateThrough configures whether to pass state between iterations
func (l *LoopAgent) WithPassStateThrough(passThrough bool) *LoopAgent {
	l.passStateThrough = passThrough
	return l
}

// WithIterationDelay sets a delay between iterations
func (l *LoopAgent) WithIterationDelay(delay time.Duration) *LoopAgent {
	l.iterationDelay = delay
	return l
}

// WithHook adds a monitoring hook to the workflow agent
func (l *LoopAgent) WithHook(hook domain.Hook) *LoopAgent {
	l.BaseWorkflowAgent.WithHook(hook)
	return l
}

// WhileLoop creates a loop that continues while the condition is true
func WhileLoop(name string, condition func(state *domain.State, iteration int) bool, step WorkflowStep) *LoopAgent {
	return NewLoopAgent(name).
		SetLoopBody(step).
		WithWhileCondition(condition)
}

// UntilLoop creates a loop that continues until the condition is true
func UntilLoop(name string, condition func(state *domain.State, iteration int) bool, step WorkflowStep) *LoopAgent {
	return NewLoopAgent(name).
		SetLoopBody(step).
		WithUntilCondition(condition)
}

// CountLoop creates a loop that executes a specific number of times
func CountLoop(name string, count int, step WorkflowStep) *LoopAgent {
	return NewLoopAgent(name).
		SetLoopBody(step).
		WithMaxIterations(count)
}

// Run executes the loop workflow
func (l *LoopAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	// Validate before running
	if err := l.Validate(); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Initialize loop state
	l.startTime = time.Now()
	l.currentIteration = 0
	l.iterationResults = make([]interface{}, 0)

	// Update status to running
	l.updateStatus(WorkflowStateRunning, "", nil)
	l.emitWorkflowEvent(domain.EventWorkflowStart, map[string]interface{}{
		"maxIterations":  l.maxIterations,
		"maxDuration":    l.maxDuration.String(),
		"collectResults": l.collectResults,
	})

	// Create workflow state
	workflowState := l.createWorkflowState(input)
	currentState := workflowState

	// Execute before hooks
	if err := l.BeforeRun(ctx, input); err != nil {
		l.updateStatus(WorkflowStateFailed, "", err)
		return nil, err
	}

	// Main loop
	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			l.updateStatus(WorkflowStateCanceled, "", ctx.Err())
			return currentState.State, ctx.Err()
		default:
		}

		// Check maximum iterations
		if l.maxIterations > 0 && l.currentIteration >= l.maxIterations {
			l.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
				"reason":    "max_iterations_reached",
				"iteration": l.currentIteration,
			})
			break
		}

		// Check maximum duration
		if l.maxDuration > 0 && time.Since(l.startTime) >= l.maxDuration {
			l.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
				"reason":    "max_duration_reached",
				"iteration": l.currentIteration,
				"duration":  time.Since(l.startTime).String(),
			})
			break
		}

		// Check continue condition (while loop)
		if l.continueCondition != nil && !l.continueCondition(currentState.State, l.currentIteration) {
			l.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
				"reason":    "continue_condition_false",
				"iteration": l.currentIteration,
			})
			break
		}

		// Check break condition (until loop)
		if l.breakCondition != nil && l.breakCondition(currentState.State, l.currentIteration) {
			l.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
				"reason":    "break_condition_true",
				"iteration": l.currentIteration,
			})
			break
		}

		// Update iteration status
		iterationName := fmt.Sprintf("iteration-%d", l.currentIteration)
		l.updateStatus(WorkflowStateRunning, iterationName, nil)
		l.updateStepStatus(iterationName, StepStatus{
			State:     StepStateRunning,
			StartTime: time.Now(),
		})

		// Emit iteration start event
		l.emitWorkflowEvent(domain.EventWorkflowStep, map[string]interface{}{
			"iteration": l.currentIteration,
			"step":      "iteration_start",
		})

		// Execute loop body
		iterationStart := time.Now()
		var iterationState *WorkflowState
		var err error

		if l.passStateThrough {
			iterationState, err = l.loopBody.Execute(ctx, currentState)
		} else {
			// Create fresh state for each iteration
			freshState := l.createWorkflowState(input)
			iterationState, err = l.loopBody.Execute(ctx, freshState)
		}

		iterationDuration := time.Since(iterationStart)

		// Update step status
		stepStatus := StepStatus{
			StartTime: l.status.Steps[iterationName].StartTime,
			EndTime:   time.Now(),
		}

		if err != nil {
			stepStatus.State = StepStateFailed
			stepStatus.Error = err
			l.updateStepStatus(iterationName, stepStatus)

			// Emit error event
			l.emitWorkflowEvent(domain.EventAgentError, map[string]interface{}{
				"iteration": l.currentIteration,
				"error":     err.Error(),
				"duration":  iterationDuration.String(),
			})

			if l.breakOnError {
				l.updateStatus(WorkflowStateFailed, iterationName, err)
				if err := l.AfterRun(ctx, input, currentState.State, err); err != nil {
					// Log the after-run error but still return the original error
					_ = err // Explicitly ignore the error for linting
				}
				return nil, fmt.Errorf("loop failed at iteration %d: %w", l.currentIteration, err)
			}

			// Continue with current state if not breaking on error
			iterationState = currentState
		} else {
			stepStatus.State = StepStateCompleted
			l.updateStepStatus(iterationName, stepStatus)

			// Update current state for next iteration
			if l.passStateThrough && iterationState != nil {
				currentState = iterationState
			}
		}

		// Collect results if enabled
		if l.collectResults && iterationState != nil {
			// Store the iteration result
			iterationResult := map[string]interface{}{
				"iteration": l.currentIteration,
				"duration":  iterationDuration,
				"state":     iterationState.State,
				"error":     err,
			}
			l.iterationResults = append(l.iterationResults, iterationResult)
		}

		// Emit iteration complete event
		l.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
			"iteration": l.currentIteration,
			"step":      "iteration_complete",
			"duration":  iterationDuration.String(),
			"success":   err == nil,
		})

		// Increment iteration counter
		l.currentIteration++

		// Apply iteration delay
		if l.iterationDelay > 0 {
			select {
			case <-time.After(l.iterationDelay):
			case <-ctx.Done():
				l.updateStatus(WorkflowStateCanceled, "", ctx.Err())
				return currentState.State, ctx.Err()
			}
		}
	}

	// Update final status
	l.updateStatus(WorkflowStateCompleted, "", nil)

	// Add loop metadata to final state
	if currentState != nil && currentState.Metadata != nil {
		currentState.Metadata["total_iterations"] = l.currentIteration
		currentState.Metadata["total_duration"] = time.Since(l.startTime)
		currentState.Metadata["loop_completed"] = true

		if l.collectResults {
			currentState.Metadata["iteration_results"] = l.iterationResults
		}
	}

	// Execute after hooks
	finalState := currentState.State
	if err := l.AfterRun(ctx, input, finalState, nil); err != nil {
		return finalState, err
	}

	// Emit workflow complete event
	l.emitWorkflowEvent(domain.EventAgentComplete, map[string]interface{}{
		"duration":         time.Since(l.startTime),
		"total_iterations": l.currentIteration,
		"completed":        true,
	})

	return finalState, nil
}

// GetIterationResults returns the results from all iterations
func (l *LoopAgent) GetIterationResults() []interface{} {
	results := make([]interface{}, len(l.iterationResults))
	copy(results, l.iterationResults)
	return results
}

// GetCurrentIteration returns the current iteration number
func (l *LoopAgent) GetCurrentIteration() int {
	return l.currentIteration
}

// GetTotalDuration returns the total duration of the loop
func (l *LoopAgent) GetTotalDuration() time.Duration {
	if l.startTime.IsZero() {
		return 0
	}
	return time.Since(l.startTime)
}

// Validate validates the loop workflow configuration
func (l *LoopAgent) Validate() error {
	// Validate base agent but skip the step validation since we use loop body
	if err := l.BaseAgentImpl.Validate(); err != nil {
		return err
	}

	if l.loopBody == nil {
		return fmt.Errorf("loop workflow must have a loop body")
	}

	if err := l.loopBody.Validate(); err != nil {
		return fmt.Errorf("loop body validation failed: %w", err)
	}

	// Validate that we have at least one termination condition
	hasTermination := l.maxIterations > 0 ||
		l.maxDuration > 0 ||
		l.continueCondition != nil ||
		l.breakCondition != nil

	if !hasTermination {
		return fmt.Errorf("loop workflow must have at least one termination condition (maxIterations, maxDuration, continueCondition, or breakCondition)")
	}

	// Validate iteration delay
	if l.iterationDelay < 0 {
		return fmt.Errorf("iteration delay cannot be negative")
	}

	return nil
}

// Reset resets the loop state for reuse
func (l *LoopAgent) Reset() {
	l.currentIteration = 0
	l.startTime = time.Time{}
	l.iterationResults = make([]interface{}, 0)
}

// GetLoopBody returns the loop body step
func (l *LoopAgent) GetLoopBody() WorkflowStep {
	return l.loopBody
}
