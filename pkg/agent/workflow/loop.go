// ABOUTME: Loop workflow agent that executes iterative processing with conditions and counters
// ABOUTME: Provides for/while loop logic for workflows with support for break conditions and iteration limits

package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// LoopAgent executes workflow steps in a loop until a condition is met.
// It provides various loop patterns including while loops, until loops, count loops,
// and forever loops. The agent supports iteration limits, time constraints, result
// collection, and state passing between iterations.
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

// getCurrentIteration safely returns the current iteration count
func (l *LoopAgent) getCurrentIteration() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.currentIteration
}

// setCurrentIteration safely sets the current iteration count
func (l *LoopAgent) setCurrentIteration(iter int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.currentIteration = iter
}

// incrementIteration safely increments the iteration count
func (l *LoopAgent) incrementIteration() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.currentIteration++
}

// getStartTime safely returns the start time
func (l *LoopAgent) getStartTime() time.Time {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.startTime
}

// setStartTime safely sets the start time
func (l *LoopAgent) setStartTime(t time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.startTime = t
}

// appendIterationResult safely appends a result to iteration results
func (l *LoopAgent) appendIterationResult(result interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.iterationResults = append(l.iterationResults, result)
}

// resetIterationResults safely resets the iteration results
func (l *LoopAgent) resetIterationResults() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.iterationResults = make([]interface{}, 0)
}

// getIterationResultsCopy safely returns a copy of iteration results
func (l *LoopAgent) getIterationResultsCopy() []interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	results := make([]interface{}, len(l.iterationResults))
	copy(results, l.iterationResults)
	return results
}

// LoopType represents different types of loop behavior.
// It defines the various loop patterns available for workflow execution.
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

// NewLoopAgent creates a new loop workflow agent.
// By default, the agent has unlimited iterations and duration, collects results,
// breaks on error, and passes state through iterations.
//
// Parameters:
//   - name: The name of the loop workflow
//
// Returns a new LoopAgent instance.
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

// SetLoopBody sets the step to execute in each iteration.
// This is the core logic that will be repeated during the loop.
//
// Parameters:
//   - step: The workflow step to execute in each iteration
//
// Returns the LoopAgent for method chaining.
func (l *LoopAgent) SetLoopBody(step WorkflowStep) *LoopAgent {
	l.loopBody = step
	return l
}

// SetLoopAgent sets an agent as the loop body.
// This is a convenience method that wraps the agent in an AgentStep.
//
// Parameters:
//   - agent: The agent to execute in each iteration
//
// Returns the LoopAgent for method chaining.
func (l *LoopAgent) SetLoopAgent(agent domain.BaseAgent) *LoopAgent {
	step := &AgentStep{
		name:  fmt.Sprintf("loop-%s", agent.Name()),
		agent: agent,
	}
	l.loopBody = step
	return l
}

// WithWhileCondition sets a condition that continues the loop while true.
// The loop will execute as long as this condition returns true.
//
// Parameters:
//   - condition: Function that evaluates the state and iteration count
//
// Returns the LoopAgent for method chaining.
func (l *LoopAgent) WithWhileCondition(condition func(state *domain.State, iteration int) bool) *LoopAgent {
	l.continueCondition = condition
	return l
}

// WithUntilCondition sets a condition that breaks the loop when true.
// The loop will execute until this condition returns true.
//
// Parameters:
//   - condition: Function that evaluates the state and iteration count
//
// Returns the LoopAgent for method chaining.
func (l *LoopAgent) WithUntilCondition(condition func(state *domain.State, iteration int) bool) *LoopAgent {
	l.breakCondition = condition
	return l
}

// WithMaxIterations sets the maximum number of iterations.
// The loop will terminate after executing this many times.
//
// Parameters:
//   - max: Maximum iteration count (0 = unlimited)
//
// Returns the LoopAgent for method chaining.
func (l *LoopAgent) WithMaxIterations(max int) *LoopAgent {
	l.maxIterations = max
	return l
}

// WithMaxDuration sets the maximum total duration for the loop.
// The loop will terminate after running for this duration.
//
// Parameters:
//   - duration: Maximum duration (0 = unlimited)
//
// Returns the LoopAgent for method chaining.
func (l *LoopAgent) WithMaxDuration(duration time.Duration) *LoopAgent {
	l.maxDuration = duration
	return l
}

// WithCollectResults configures whether to collect results from each iteration.
// When enabled, results are stored and available via GetIterationResults().
//
// Parameters:
//   - collect: If true, collects results from each iteration
//
// Returns the LoopAgent for method chaining.
func (l *LoopAgent) WithCollectResults(collect bool) *LoopAgent {
	l.collectResults = collect
	return l
}

// WithBreakOnError configures whether to break on first error.
// When enabled, the loop terminates immediately on any error.
//
// Parameters:
//   - breakOnErr: If true, breaks loop on first error
//
// Returns the LoopAgent for method chaining.
func (l *LoopAgent) WithBreakOnError(breakOnErr bool) *LoopAgent {
	l.breakOnError = breakOnErr
	return l
}

// WithPassStateThrough configures whether to pass state between iterations.
// When enabled, each iteration receives the output state from the previous iteration.
// When disabled, each iteration starts with the original input state.
//
// Parameters:
//   - passThrough: If true, passes state from one iteration to the next
//
// Returns the LoopAgent for method chaining.
func (l *LoopAgent) WithPassStateThrough(passThrough bool) *LoopAgent {
	l.passStateThrough = passThrough
	return l
}

// WithIterationDelay sets a delay between iterations.
// This can be useful for rate limiting or avoiding resource exhaustion.
//
// Parameters:
//   - delay: Duration to wait between iterations
//
// Returns the LoopAgent for method chaining.
func (l *LoopAgent) WithIterationDelay(delay time.Duration) *LoopAgent {
	l.iterationDelay = delay
	return l
}

// WithHook adds a monitoring hook to the workflow agent.
// Hooks allow monitoring and customization of loop execution.
//
// Parameters:
//   - hook: The hook to add
//
// Returns the LoopAgent for method chaining.
func (l *LoopAgent) WithHook(hook domain.Hook) *LoopAgent {
	l.BaseWorkflowAgent.WithHook(hook)
	return l
}

// WhileLoop creates a loop that continues while the condition is true.
// This is a convenience function for creating a while-style loop.
//
// Parameters:
//   - name: The name of the loop
//   - condition: Function that returns true to continue looping
//   - step: The workflow step to execute in each iteration
//
// Returns a configured LoopAgent.
func WhileLoop(name string, condition func(state *domain.State, iteration int) bool, step WorkflowStep) *LoopAgent {
	return NewLoopAgent(name).
		SetLoopBody(step).
		WithWhileCondition(condition)
}

// UntilLoop creates a loop that continues until the condition is true.
// This is a convenience function for creating an until-style loop.
//
// Parameters:
//   - name: The name of the loop
//   - condition: Function that returns true to stop looping
//   - step: The workflow step to execute in each iteration
//
// Returns a configured LoopAgent.
func UntilLoop(name string, condition func(state *domain.State, iteration int) bool, step WorkflowStep) *LoopAgent {
	return NewLoopAgent(name).
		SetLoopBody(step).
		WithUntilCondition(condition)
}

// CountLoop creates a loop that executes a specific number of times.
// This is a convenience function for creating a count-based loop.
//
// Parameters:
//   - name: The name of the loop
//   - count: Number of iterations to execute
//   - step: The workflow step to execute in each iteration
//
// Returns a configured LoopAgent.
func CountLoop(name string, count int, step WorkflowStep) *LoopAgent {
	return NewLoopAgent(name).
		SetLoopBody(step).
		WithMaxIterations(count)
}

// Run executes the loop workflow.
// It repeatedly executes the loop body until a termination condition is met.
// The loop can be terminated by: max iterations, max duration, while condition
// becoming false, until condition becoming true, or an error (if break on error).
//
// Parameters:
//   - ctx: The execution context
//   - input: The initial state
//
// Returns the final state after loop completion or an error.
func (l *LoopAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	// Validate before running
	if err := l.Validate(); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Initialize loop state
	l.setStartTime(time.Now())
	l.setCurrentIteration(0)
	l.resetIterationResults()

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
		currentIter := l.getCurrentIteration()

		if l.maxIterations > 0 && currentIter >= l.maxIterations {
			l.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
				"reason":    "max_iterations_reached",
				"iteration": currentIter,
			})
			break
		}

		// Check maximum duration
		startTime := l.getStartTime()
		if l.maxDuration > 0 && time.Since(startTime) >= l.maxDuration {
			l.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
				"reason":    "max_duration_reached",
				"iteration": l.getCurrentIteration(),
				"duration":  time.Since(startTime).String(),
			})
			break
		}

		// Check continue condition (while loop)
		currentIter = l.getCurrentIteration()
		if l.continueCondition != nil && !l.continueCondition(currentState.State, currentIter) {
			l.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
				"reason":    "continue_condition_false",
				"iteration": currentIter,
			})
			break
		}

		// Check break condition (until loop)
		currentIter = l.getCurrentIteration()
		if l.breakCondition != nil && l.breakCondition(currentState.State, currentIter) {
			l.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
				"reason":    "break_condition_true",
				"iteration": currentIter,
			})
			break
		}

		// Update iteration status
		currentIter = l.getCurrentIteration()
		iterationName := fmt.Sprintf("iteration-%d", currentIter)
		l.updateStatus(WorkflowStateRunning, iterationName, nil)
		l.updateStepStatus(iterationName, StepStatus{
			State:     StepStateRunning,
			StartTime: time.Now(),
		})

		// Emit iteration start event
		l.emitWorkflowEvent(domain.EventWorkflowStep, map[string]interface{}{
			"iteration": currentIter,
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
		// Get start time with lock
		l.mu.RLock()
		stepStartTime := l.status.Steps[iterationName].StartTime
		l.mu.RUnlock()

		stepStatus := StepStatus{
			StartTime: stepStartTime,
			EndTime:   time.Now(),
		}

		if err != nil {
			stepStatus.State = StepStateFailed
			stepStatus.Error = err
			l.updateStepStatus(iterationName, stepStatus)

			// Emit error event
			l.emitWorkflowEvent(domain.EventAgentError, map[string]interface{}{
				"iteration": currentIter,
				"error":     err.Error(),
				"duration":  iterationDuration.String(),
			})

			if l.breakOnError {
				l.updateStatus(WorkflowStateFailed, iterationName, err)
				if err := l.AfterRun(ctx, input, currentState.State, err); err != nil {
					// Log the after-run error but still return the original error
					_ = err // Explicitly ignore the error for linting
				}
				return nil, fmt.Errorf("loop failed at iteration %d: %w", currentIter, err)
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
				"iteration": currentIter,
				"duration":  iterationDuration,
				"state":     iterationState.State,
				"error":     err,
			}
			l.appendIterationResult(iterationResult)
		}

		// Emit iteration complete event
		l.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
			"iteration": currentIter,
			"step":      "iteration_complete",
			"duration":  iterationDuration.String(),
			"success":   err == nil,
		})

		// Increment iteration counter
		l.incrementIteration()

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
		currentState.Metadata["total_iterations"] = l.getCurrentIteration()
		currentState.Metadata["total_duration"] = time.Since(l.getStartTime())
		currentState.Metadata["loop_completed"] = true

		if l.collectResults {
			currentState.Metadata["iteration_results"] = l.getIterationResultsCopy()
		}
	}

	// Execute after hooks
	finalState := currentState.State
	if err := l.AfterRun(ctx, input, finalState, nil); err != nil {
		return finalState, err
	}

	// Emit workflow complete event
	l.emitWorkflowEvent(domain.EventAgentComplete, map[string]interface{}{
		"duration":         time.Since(l.getStartTime()),
		"total_iterations": l.getCurrentIteration(),
		"completed":        true,
	})

	return finalState, nil
}

// GetIterationResults returns the results from all iterations.
// Only available when WithCollectResults(true) is set.
//
// Returns a slice containing results from each iteration.
func (l *LoopAgent) GetIterationResults() []interface{} {
	return l.getIterationResultsCopy()
}

// GetCurrentIteration returns the current iteration number.
// The iteration counter starts at 0 and increments after each iteration.
//
// Returns the current iteration count.
func (l *LoopAgent) GetCurrentIteration() int {
	return l.getCurrentIteration()
}

// GetTotalDuration returns the total duration of the loop.
// This measures the time from when Run() was called to the current moment.
//
// Returns the elapsed duration or 0 if not started.
func (l *LoopAgent) GetTotalDuration() time.Duration {
	startTime := l.getStartTime()
	if startTime.IsZero() {
		return 0
	}
	return time.Since(startTime)
}

// Validate validates the loop workflow configuration.
// It ensures the loop has a body and at least one termination condition.
//
// Returns an error if validation fails.
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

// Reset resets the loop state for reuse.
// This clears the iteration counter, start time, and results,
// allowing the same loop agent to be executed again.
func (l *LoopAgent) Reset() {
	l.setCurrentIteration(0)
	l.setStartTime(time.Time{})
	l.resetIterationResults()
}

// GetLoopBody returns the loop body step.
// This is the workflow step that is executed in each iteration.
//
// Returns the configured loop body or nil.
func (l *LoopAgent) GetLoopBody() WorkflowStep {
	return l.loopBody
}
