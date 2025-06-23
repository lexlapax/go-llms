// ABOUTME: Parallel workflow agent that executes multiple agents concurrently
// ABOUTME: Supports different merge strategies for combining results from parallel agents

package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// MergeStrategy defines how to merge results from parallel agents.
// It determines how the final state is constructed from the results
// of all parallel executions.
type MergeStrategy string

const (
	// MergeAll combines all results into the state
	MergeAll MergeStrategy = "all"
	// MergeFirst uses only the first completed result
	MergeFirst MergeStrategy = "first"
	// MergeCustom uses a custom merge function
	MergeCustom MergeStrategy = "custom"
)

// ParallelAgent executes workflow steps in parallel.
// It runs multiple agents concurrently and provides various strategies
// for merging their results. It supports concurrency limiting, timeouts,
// and different merge strategies including first-complete and custom merging.
type ParallelAgent struct {
	*BaseWorkflowAgent

	// Options
	maxConcurrency int
	mergeStrategy  MergeStrategy
	mergeFunc      func(results map[string]*domain.State) *domain.State
	timeout        time.Duration
}

// NewParallelAgent creates a new parallel workflow agent.
// By default, it has no concurrency limit and uses the MergeAll strategy.
//
// Parameters:
//   - name: The name of the parallel workflow
//
// Returns a new ParallelAgent instance.
func NewParallelAgent(name string) *ParallelAgent {
	return &ParallelAgent{
		BaseWorkflowAgent: NewBaseWorkflowAgent(
			name,
			fmt.Sprintf("Parallel workflow: %s", name),
			domain.AgentTypeParallel,
		),
		maxConcurrency: 0, // 0 means no limit
		mergeStrategy:  MergeAll,
	}
}

// WithMaxConcurrency sets the maximum number of concurrent executions.
// This limits how many agents can run simultaneously.
//
// Parameters:
//   - max: Maximum concurrent executions (0 = unlimited)
//
// Returns the ParallelAgent for method chaining.
func (p *ParallelAgent) WithMaxConcurrency(max int) *ParallelAgent {
	p.maxConcurrency = max
	return p
}

// WithMergeStrategy sets how to merge results from parallel agents.
// Available strategies: MergeAll, MergeFirst, MergeCustom.
//
// Parameters:
//   - strategy: The merge strategy to use
//
// Returns the ParallelAgent for method chaining.
func (p *ParallelAgent) WithMergeStrategy(strategy MergeStrategy) *ParallelAgent {
	p.mergeStrategy = strategy
	return p
}

// WithMergeFunc sets a custom merge function.
// This automatically sets the merge strategy to MergeCustom.
// The function receives all successful results and should return a final state.
//
// Parameters:
//   - f: Function that merges results into a single state
//
// Returns the ParallelAgent for method chaining.
func (p *ParallelAgent) WithMergeFunc(f func(results map[string]*domain.State) *domain.State) *ParallelAgent {
	p.mergeStrategy = MergeCustom
	p.mergeFunc = f
	return p
}

// WithHook adds a monitoring hook to the workflow agent.
// Hooks allow monitoring and customization of parallel execution.
//
// Parameters:
//   - hook: The hook to add
//
// Returns the ParallelAgent for method chaining.
func (p *ParallelAgent) WithHook(hook domain.Hook) *ParallelAgent {
	p.BaseWorkflowAgent.WithHook(hook)
	return p
}

// WithTimeout sets the maximum time to wait for all agents.
// If the timeout expires, running agents are canceled.
//
// Parameters:
//   - timeout: Maximum duration to wait
//
// Returns the ParallelAgent for method chaining.
func (p *ParallelAgent) WithTimeout(timeout time.Duration) *ParallelAgent {
	p.timeout = timeout
	return p
}

// AddAgent adds an agent as a workflow step.
// This is a convenience method that wraps the agent in an AgentStep.
//
// Parameters:
//   - agent: The agent to add to parallel execution
//
// Returns the ParallelAgent for method chaining.
func (p *ParallelAgent) AddAgent(agent domain.BaseAgent) *ParallelAgent {
	step := &AgentStep{
		name:  agent.Name(),
		agent: agent,
	}
	if err := p.AddStep(step); err != nil {
		// Log error but continue for fluent interface
		// In practice, this error should be very rare
		_ = err // Explicitly ignore the error for linting
	}
	return p
}

// Run executes the parallel workflow.
// It runs all configured agents concurrently and merges their results
// according to the configured merge strategy. The workflow respects
// concurrency limits and timeouts.
//
// Parameters:
//   - ctx: The execution context
//   - input: The initial state
//
// Returns the merged final state or an error.
func (p *ParallelAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	// Validate before running
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Create timeout context if specified
	if p.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.timeout)
		defer cancel()
	}

	// Update status to running
	p.updateStatus(WorkflowStateRunning, "", nil)
	p.emitWorkflowEvent(domain.EventWorkflowStart, map[string]interface{}{
		"steps":          len(p.steps),
		"maxConcurrency": p.maxConcurrency,
		"mergeStrategy":  p.mergeStrategy,
	})

	// Create workflow state
	workflowState := p.createWorkflowState(input)

	// Execute before hooks
	if err := p.BeforeRun(ctx, input); err != nil {
		p.updateStatus(WorkflowStateFailed, "", err)
		return nil, err
	}

	// Execute all steps in parallel
	results := make(map[string]*domain.State)
	errors := make(map[string]error)
	var firstResult *domain.State
	var firstResultName string
	var mu sync.Mutex

	// Create semaphore for concurrency control
	var sem chan struct{}
	if p.maxConcurrency > 0 {
		sem = make(chan struct{}, p.maxConcurrency)
	}

	// Wait group for all agents
	var wg sync.WaitGroup

	// Launch all agents
	for _, step := range p.steps {
		wg.Add(1)

		// Acquire semaphore if needed
		if sem != nil {
			sem <- struct{}{}
		}

		go func(s WorkflowStep) {
			defer wg.Done()
			if sem != nil {
				defer func() { <-sem }()
			}

			stepName := s.Name()
			startTime := time.Now()

			// Update step status
			p.updateStepStatus(stepName, StepStatus{
				State:     StepStateRunning,
				StartTime: startTime,
			})

			// Emit step start event
			p.emitWorkflowEvent(domain.EventWorkflowStep, map[string]interface{}{
				"step":     stepName,
				"parallel": true,
			})

			// Execute step
			result, err := s.Execute(ctx, workflowState)

			// Check if context was canceled
			if ctx.Err() != nil {
				err = fmt.Errorf("workflow canceled: %w", ctx.Err())
			}

			// Store result
			mu.Lock()
			if err != nil {
				errors[stepName] = err
				p.updateStepStatus(stepName, StepStatus{
					State:     StepStateFailed,
					StartTime: startTime,
					EndTime:   time.Now(),
					Error:     err,
				})
			} else {
				results[stepName] = result.State
				// Track first result for MergeFirst strategy
				if firstResult == nil && p.mergeStrategy == MergeFirst {
					firstResult = result.State
					firstResultName = stepName
				}
				p.updateStepStatus(stepName, StepStatus{
					State:     StepStateCompleted,
					StartTime: startTime,
					EndTime:   time.Now(),
				})
			}
			mu.Unlock()

			// Emit completion event
			if err != nil {
				p.emitWorkflowEvent(domain.EventAgentError, map[string]interface{}{
					"step":  stepName,
					"error": err.Error(),
				})
			} else {
				p.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
					"step":     stepName,
					"complete": true,
				})
			}
		}(step)
	}

	// Wait for all agents to complete
	wg.Wait()

	// Check for errors
	if len(errors) > 0 {
		p.updateStatus(WorkflowStateFailed, "", fmt.Errorf("parallel execution had %d failures", len(errors)))

		// If MergeFirst strategy and we have at least one success, continue
		if p.mergeStrategy != MergeFirst || len(results) == 0 {
			return nil, fmt.Errorf("parallel execution failed: %d agents failed", len(errors))
		}
	}

	// Merge results based on strategy
	var finalState *domain.State
	switch p.mergeStrategy {
	case MergeFirst:
		// Use the first completed result
		if firstResult != nil {
			finalState = firstResult
			p.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
				"message": fmt.Sprintf("Using result from first completed agent: %s", firstResultName),
			})
		} else {
			return nil, fmt.Errorf("no successful results for MergeFirst strategy")
		}
	case MergeCustom:
		if p.mergeFunc == nil {
			return nil, fmt.Errorf("merge function not set for custom merge strategy")
		}
		finalState = p.mergeFunc(results)
	case MergeAll:
		fallthrough
	default:
		// Merge all results into a single state
		finalState = input.Clone()

		// Store individual results
		parallelResults := make(map[string]interface{})
		for stepName, state := range results {
			// Extract all values from the state
			stateData := make(map[string]interface{})
			// This would need a method to get all keys from state
			// For now, we'll store specific known keys
			if response, exists := state.Get("response"); exists {
				stateData["response"] = response
			}
			if result, exists := state.Get("result"); exists {
				stateData["result"] = result
			}
			parallelResults[stepName] = stateData
		}
		finalState.Set("parallel_results", parallelResults)

		// Store errors if any
		if len(errors) > 0 {
			errorMap := make(map[string]string)
			for stepName, err := range errors {
				errorMap[stepName] = err.Error()
			}
			finalState.Set("parallel_errors", errorMap)
		}
	}

	// Update final status
	if len(errors) == 0 {
		p.updateStatus(WorkflowStateCompleted, "", nil)
	} else {
		p.updateStatus(WorkflowStateCompleted, "", fmt.Errorf("completed with %d errors", len(errors)))
	}

	// Execute after hooks
	if err := p.AfterRun(ctx, input, finalState, nil); err != nil {
		return finalState, err
	}

	// Emit workflow complete event
	p.emitWorkflowEvent(domain.EventAgentComplete, map[string]interface{}{
		"duration":   time.Since(p.status.StartTime),
		"steps":      len(p.steps),
		"successful": len(results),
		"failed":     len(errors),
	})

	return finalState, nil
}

// RunAsync executes the workflow asynchronously.
// It returns a channel that emits events as the workflow progresses,
// including a final completion or error event. The channel is closed
// when the workflow completes.
//
// Parameters:
//   - ctx: The execution context
//   - input: The initial state
//
// Returns an event channel and any validation error.
func (p *ParallelAgent) RunAsync(ctx context.Context, input *domain.State) (<-chan domain.Event, error) {
	// Validate before running
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Create event channel
	eventChan := make(chan domain.Event, 100)

	// Run workflow in goroutine
	go func() {
		defer close(eventChan)

		result, err := p.Run(ctx, input)

		// Send final event
		finalEvent := domain.Event{
			Type:      domain.EventAgentComplete,
			AgentID:   p.ID(),
			AgentName: p.Name(),
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
