// ABOUTME: Conditional workflow agent that executes different branches based on state conditions
// ABOUTME: Provides if/else logic for workflow branching with support for multiple conditions

package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// ConditionalAgent executes different workflow branches based on conditions.
// It provides if/else logic for workflows, allowing dynamic execution paths
// based on state evaluation. Multiple branches can be configured with priorities,
// and a default branch can be specified for when no conditions match.
type ConditionalAgent struct {
	*BaseWorkflowAgent

	// Conditional branches
	branches []ConditionalBranch

	// Default branch (executed if no conditions match)
	defaultBranch WorkflowStep

	// Options
	evaluateAllConditions bool // If true, evaluates all conditions even after finding a match
	allowMultipleMatches  bool // If true, allows multiple branches to execute
}

// ConditionalBranch represents a condition and its associated workflow step.
// Each branch has a condition function that evaluates the current state
// and a step to execute if the condition is true.
type ConditionalBranch struct {
	// Name of the branch for identification
	Name string

	// Condition function that evaluates the state
	Condition func(state *domain.State) bool

	// Step to execute if condition is true
	Step WorkflowStep

	// Priority for ordering (higher numbers execute first)
	Priority int
}

// NewConditionalAgent creates a new conditional workflow agent.
// By default, it evaluates conditions in order and executes only the first match.
//
// Parameters:
//   - name: The name of the conditional workflow
//
// Returns a new ConditionalAgent instance.
func NewConditionalAgent(name string) *ConditionalAgent {
	return &ConditionalAgent{
		BaseWorkflowAgent: NewBaseWorkflowAgent(
			name,
			fmt.Sprintf("Conditional workflow: %s", name),
			domain.AgentTypeConditional,
		),
		branches:              make([]ConditionalBranch, 0),
		evaluateAllConditions: false,
		allowMultipleMatches:  false,
	}
}

// AddBranch adds a conditional branch to the workflow.
// Branches are evaluated in the order they are added unless priorities are set.
//
// Parameters:
//   - name: Unique name for the branch
//   - condition: Function that evaluates the state and returns true if branch should execute
//   - step: The workflow step to execute if condition is true
//
// Returns the ConditionalAgent for method chaining.
func (c *ConditionalAgent) AddBranch(name string, condition func(state *domain.State) bool, step WorkflowStep) *ConditionalAgent {
	branch := ConditionalBranch{
		Name:      name,
		Condition: condition,
		Step:      step,
		Priority:  0,
	}
	c.branches = append(c.branches, branch)
	return c
}

// AddBranchWithPriority adds a conditional branch with priority.
// Higher priority branches are evaluated first.
//
// Parameters:
//   - name: Unique name for the branch
//   - condition: Function that evaluates the state and returns true if branch should execute
//   - step: The workflow step to execute if condition is true
//   - priority: Branch priority (higher numbers execute first)
//
// Returns the ConditionalAgent for method chaining.
func (c *ConditionalAgent) AddBranchWithPriority(name string, condition func(state *domain.State) bool, step WorkflowStep, priority int) *ConditionalAgent {
	branch := ConditionalBranch{
		Name:      name,
		Condition: condition,
		Step:      step,
		Priority:  priority,
	}
	c.branches = append(c.branches, branch)
	return c
}

// AddAgent adds an agent as a conditional branch.
// This is a convenience method that wraps the agent in an AgentStep.
//
// Parameters:
//   - name: Unique name for the branch
//   - condition: Function that evaluates the state and returns true if agent should execute
//   - agent: The agent to execute if condition is true
//
// Returns the ConditionalAgent for method chaining.
func (c *ConditionalAgent) AddAgent(name string, condition func(state *domain.State) bool, agent domain.BaseAgent) *ConditionalAgent {
	step := &AgentStep{
		name:  fmt.Sprintf("%s-%s", name, agent.Name()),
		agent: agent,
	}
	return c.AddBranch(name, condition, step)
}

// SetDefaultBranch sets the default branch (executed when no conditions match).
// Only one default branch can be set; setting a new one replaces the previous.
//
// Parameters:
//   - step: The workflow step to execute when no conditions match
//
// Returns the ConditionalAgent for method chaining.
func (c *ConditionalAgent) SetDefaultBranch(step WorkflowStep) *ConditionalAgent {
	c.defaultBranch = step
	return c
}

// SetDefaultAgent sets an agent as the default branch.
// This is a convenience method that wraps the agent in an AgentStep.
//
// Parameters:
//   - agent: The agent to execute when no conditions match
//
// Returns the ConditionalAgent for method chaining.
func (c *ConditionalAgent) SetDefaultAgent(agent domain.BaseAgent) *ConditionalAgent {
	step := &AgentStep{
		name:  fmt.Sprintf("default-%s", agent.Name()),
		agent: agent,
	}
	c.defaultBranch = step
	return c
}

// WithEvaluateAllConditions configures whether to evaluate all conditions.
// By default, evaluation stops after the first match unless this is enabled.
//
// Parameters:
//   - evaluate: If true, all conditions are evaluated regardless of matches
//
// Returns the ConditionalAgent for method chaining.
func (c *ConditionalAgent) WithEvaluateAllConditions(evaluate bool) *ConditionalAgent {
	c.evaluateAllConditions = evaluate
	return c
}

// WithAllowMultipleMatches configures whether multiple branches can execute.
// By default, only the first matching branch executes.
//
// Parameters:
//   - allow: If true, all matching branches will execute
//
// Returns the ConditionalAgent for method chaining.
func (c *ConditionalAgent) WithAllowMultipleMatches(allow bool) *ConditionalAgent {
	c.allowMultipleMatches = allow
	return c
}

// WithHook adds a monitoring hook to the workflow agent.
// Hooks allow monitoring and customization of workflow execution.
//
// Parameters:
//   - hook: The hook to add
//
// Returns the ConditionalAgent for method chaining.
func (c *ConditionalAgent) WithHook(hook domain.Hook) *ConditionalAgent {
	c.BaseWorkflowAgent.WithHook(hook)
	return c
}

// Run executes the conditional workflow.
// It evaluates conditions in priority order and executes matching branches.
// The workflow fails if any executed branch fails (unless error handling is configured).
//
// Parameters:
//   - ctx: The execution context
//   - input: The initial state
//
// Returns the final state after executing matching branches or an error.
func (c *ConditionalAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	// Validate before running
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Update status to running
	c.updateStatus(WorkflowStateRunning, "", nil)
	c.emitWorkflowEvent(domain.EventWorkflowStart, map[string]interface{}{
		"branches":   len(c.branches),
		"hasDefault": c.defaultBranch != nil,
	})

	// Create workflow state
	workflowState := c.createWorkflowState(input)

	// Execute before hooks
	if err := c.BeforeRun(ctx, input); err != nil {
		c.updateStatus(WorkflowStateFailed, "", err)
		return nil, err
	}

	// Sort branches by priority (highest first)
	branches := make([]ConditionalBranch, len(c.branches))
	copy(branches, c.branches)

	// Simple sort by priority (highest first)
	for i := 0; i < len(branches); i++ {
		for j := i + 1; j < len(branches); j++ {
			if branches[j].Priority > branches[i].Priority {
				branches[i], branches[j] = branches[j], branches[i]
			}
		}
	}

	// Evaluate conditions and execute matching branches
	var executedBranches []string
	var lastResult = workflowState
	var hasMatch bool

	for _, branch := range branches {
		// Emit condition evaluation event
		c.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
			"evaluating": branch.Name,
			"priority":   branch.Priority,
		})

		// Evaluate condition
		matches := branch.Condition(workflowState.State)

		if matches {
			hasMatch = true
			executedBranches = append(executedBranches, branch.Name)

			// Update status
			c.updateStatus(WorkflowStateRunning, branch.Name, nil)
			c.updateStepStatus(branch.Name, StepStatus{
				State:     StepStateRunning,
				StartTime: time.Now(),
			})

			// Execute branch
			result, err := branch.Step.Execute(ctx, lastResult)

			// Update step status
			// Get start time with lock
			c.mu.RLock()
			startTime := c.status.Steps[branch.Name].StartTime
			c.mu.RUnlock()

			stepStatus := StepStatus{
				StartTime: startTime,
				EndTime:   time.Now(),
			}

			if err != nil {
				stepStatus.State = StepStateFailed
				stepStatus.Error = err
				c.updateStepStatus(branch.Name, stepStatus)

				// Emit error event
				c.emitWorkflowEvent(domain.EventAgentError, map[string]interface{}{
					"branch": branch.Name,
					"error":  err.Error(),
				})

				// Handle error based on error handler
				if c.errorHandler != nil {
					result, err = c.handleStepError(ctx, branch.Step, lastResult, err)
					if err != nil {
						c.updateStatus(WorkflowStateFailed, branch.Name, err)
						if err := c.AfterRun(ctx, input, nil, err); err != nil {
							// Log the after-run error but still return the original error
							_ = err // Explicitly ignore the error for linting
						}
						return nil, fmt.Errorf("branch %s failed: %w", branch.Name, err)
					}
				} else {
					c.updateStatus(WorkflowStateFailed, branch.Name, err)
					if err := c.AfterRun(ctx, input, nil, err); err != nil {
						// Log the after-run error but still return the original error
						_ = err // Explicitly ignore the error for linting
					}
					return nil, fmt.Errorf("branch %s failed: %w", branch.Name, err)
				}
			}

			stepStatus.State = StepStateCompleted
			c.updateStepStatus(branch.Name, stepStatus)

			// Update workflow state for next iteration
			if result != nil {
				lastResult = result
			}

			// Emit branch complete event
			c.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
				"branch":    branch.Name,
				"completed": true,
			})

			// If not allowing multiple matches, break after first match
			if !c.allowMultipleMatches && !c.evaluateAllConditions {
				break
			}
		}

		// If not evaluating all conditions and we found a match, break
		if !c.evaluateAllConditions && hasMatch && !c.allowMultipleMatches {
			break
		}
	}

	// Execute default branch if no conditions matched
	if !hasMatch && c.defaultBranch != nil {
		executedBranches = append(executedBranches, "default")

		// Update status
		c.updateStatus(WorkflowStateRunning, "default", nil)
		c.updateStepStatus("default", StepStatus{
			State:     StepStateRunning,
			StartTime: time.Now(),
		})

		// Execute default branch
		result, err := c.defaultBranch.Execute(ctx, lastResult)

		// Update step status
		// Get start time with lock
		c.mu.RLock()
		startTime := c.status.Steps["default"].StartTime
		c.mu.RUnlock()

		stepStatus := StepStatus{
			StartTime: startTime,
			EndTime:   time.Now(),
		}

		if err != nil {
			stepStatus.State = StepStateFailed
			stepStatus.Error = err
			c.updateStepStatus("default", stepStatus)

			c.updateStatus(WorkflowStateFailed, "default", err)
			if err := c.AfterRun(ctx, input, nil, err); err != nil {
				// Log the after-run error but still return the original error
				_ = err // Explicitly ignore the error for linting
			}
			return nil, fmt.Errorf("default branch failed: %w", err)
		}

		stepStatus.State = StepStateCompleted
		c.updateStepStatus("default", stepStatus)

		if result != nil {
			lastResult = result
		}

		// Emit default branch complete event
		c.emitWorkflowEvent(domain.EventProgress, map[string]interface{}{
			"branch":    "default",
			"completed": true,
		})
	}

	// Update final status
	c.updateStatus(WorkflowStateCompleted, "", nil)

	// Add execution metadata to result
	if lastResult != nil && lastResult.Metadata != nil {
		lastResult.Metadata["executed_branches"] = executedBranches
		lastResult.Metadata["total_branches"] = len(c.branches)
		lastResult.Metadata["has_default"] = c.defaultBranch != nil
	}

	// Execute after hooks
	finalState := lastResult.State
	if err := c.AfterRun(ctx, input, finalState, nil); err != nil {
		return finalState, err
	}

	// Emit workflow complete event
	c.emitWorkflowEvent(domain.EventAgentComplete, map[string]interface{}{
		"duration":          time.Since(c.status.StartTime),
		"executed_branches": executedBranches,
		"total_branches":    len(c.branches),
	})

	return finalState, nil
}

// Validate validates the conditional workflow configuration.
// It ensures at least one branch or default branch exists and validates
// all branch configurations.
//
// Returns an error if validation fails.
func (c *ConditionalAgent) Validate() error {
	// Validate base agent but skip the step validation since we use branches
	if err := c.BaseAgentImpl.Validate(); err != nil {
		return err
	}

	if len(c.branches) == 0 && c.defaultBranch == nil {
		return fmt.Errorf("conditional workflow must have at least one branch or a default branch")
	}

	// Validate each branch
	for i, branch := range c.branches {
		if branch.Name == "" {
			return fmt.Errorf("branch %d has empty name", i)
		}
		if branch.Condition == nil {
			return fmt.Errorf("branch %s has nil condition", branch.Name)
		}
		if branch.Step == nil {
			return fmt.Errorf("branch %s has nil step", branch.Name)
		}
		if err := branch.Step.Validate(); err != nil {
			return fmt.Errorf("branch %s validation failed: %w", branch.Name, err)
		}
	}

	// Validate default branch if present
	if c.defaultBranch != nil {
		if err := c.defaultBranch.Validate(); err != nil {
			return fmt.Errorf("default branch validation failed: %w", err)
		}
	}

	return nil
}

// GetBranches returns all conditional branches.
// The returned slice is a copy to prevent external modifications.
//
// Returns a copy of all configured branches.
func (c *ConditionalAgent) GetBranches() []ConditionalBranch {
	branches := make([]ConditionalBranch, len(c.branches))
	copy(branches, c.branches)
	return branches
}

// GetDefaultBranch returns the default branch.
// Returns nil if no default branch is configured.
//
// Returns the default workflow step or nil.
func (c *ConditionalAgent) GetDefaultBranch() WorkflowStep {
	return c.defaultBranch
}
