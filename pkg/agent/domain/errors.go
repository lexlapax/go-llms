// ABOUTME: Defines domain-specific errors for the agent system
// ABOUTME: Provides typed errors for better error handling and debugging

package domain

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrAgentNotFound is returned when an agent cannot be found
	ErrAgentNotFound = errors.New("agent not found")

	// ErrCircularDependency is returned when adding an agent would create a circular dependency
	ErrCircularDependency = errors.New("circular dependency detected")

	// ErrInvalidState is returned when state is invalid or corrupted
	ErrInvalidState = errors.New("invalid state")

	// ErrInvalidConfiguration is returned when agent configuration is invalid
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrExecutionTimeout is returned when agent execution times out
	ErrExecutionTimeout = errors.New("execution timeout")

	// ErrMaxRetriesExceeded is returned when maximum retries are exceeded
	ErrMaxRetriesExceeded = errors.New("maximum retries exceeded")

	// ErrAgentInitialization is returned when agent initialization fails
	ErrAgentInitialization = errors.New("agent initialization failed")

	// ErrToolNotFound is returned when a tool cannot be found
	ErrToolNotFound = errors.New("tool not found")

	// ErrToolExecution is returned when tool execution fails
	ErrToolExecution = errors.New("tool execution failed")

	// ErrSchemaValidation is returned when schema validation fails
	ErrSchemaValidation = errors.New("schema validation failed")

	// ErrEventDispatch is returned when event dispatch fails
	ErrEventDispatch = errors.New("event dispatch failed")

	// ErrStateAccess is returned when state access fails
	ErrStateAccess = errors.New("state access failed")

	// ErrArtifactAccess is returned when artifact access fails
	ErrArtifactAccess = errors.New("artifact access failed")

	// ErrWorkflowExecution is returned when workflow execution fails
	ErrWorkflowExecution = errors.New("workflow execution failed")

	// ErrAgentCancelled is returned when agent execution is cancelled
	ErrAgentCancelled = errors.New("agent execution cancelled")

	// ErrStateReadOnly is returned when trying to modify a read-only state
	ErrStateReadOnly = errors.New("state is read-only")
)

// AgentError represents an error that occurred during agent execution
type AgentError struct {
	AgentID   string
	AgentName string
	Phase     string // "initialize", "before_run", "run", "after_run", "cleanup"
	Err       error
	Context   map[string]interface{}
}

// Error implements the error interface
func (e *AgentError) Error() string {
	return fmt.Sprintf("agent error [%s/%s] in %s: %v", e.AgentID, e.AgentName, e.Phase, e.Err)
}

// Unwrap returns the underlying error
func (e *AgentError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches target
func (e *AgentError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// NewAgentError creates a new agent error
func NewAgentError(agentID, agentName, phase string, err error) *AgentError {
	return &AgentError{
		AgentID:   agentID,
		AgentName: agentName,
		Phase:     phase,
		Err:       err,
		Context:   make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *AgentError) WithContext(key string, value interface{}) *AgentError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// MultiError represents multiple errors
type MultiError struct {
	Errors []error
}

// Error implements the error interface
func (e *MultiError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("multiple errors occurred (%d errors)", len(e.Errors))
}

// Add adds an error to the multi-error
func (e *MultiError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

// HasErrors returns true if there are any errors
func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

// Unwrap returns the errors as a slice
func (e *MultiError) Unwrap() []error {
	return e.Errors
}

// ToolError represents an error that occurred during tool execution
type ToolError struct {
	ToolName string
	Phase    string // "validation", "execution", "result_processing"
	Err      error
	Input    interface{}
	Output   interface{}
}

// Error implements the error interface
func (e *ToolError) Error() string {
	return fmt.Sprintf("tool error [%s] in %s: %v", e.ToolName, e.Phase, e.Err)
}

// Unwrap returns the underlying error
func (e *ToolError) Unwrap() error {
	return e.Err
}

// NewToolError creates a new tool error
func NewToolError(toolName, phase string, err error) *ToolError {
	return &ToolError{
		ToolName: toolName,
		Phase:    phase,
		Err:      err,
	}
}

// WithInput adds input context to the error
func (e *ToolError) WithInput(input interface{}) *ToolError {
	e.Input = input
	return e
}

// WithOutput adds output context to the error
func (e *ToolError) WithOutput(output interface{}) *ToolError {
	e.Output = output
	return e
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	// Check for specific retryable errors
	if errors.Is(err, ErrExecutionTimeout) {
		return true
	}

	// Check for agent errors in specific phases
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		// Initialization errors are generally not retryable
		if agentErr.Phase == "initialize" {
			return false
		}
		// Most other phases can be retried
		return true
	}

	// Tool errors are generally retryable
	var toolErr *ToolError
	if errors.As(err, &toolErr) {
		return toolErr.Phase == "execution"
	}

	// Default to not retryable
	return false
}

// IsFatal checks if an error is fatal and should stop execution
func IsFatal(err error) bool {
	// Check for specific fatal errors
	if errors.Is(err, ErrCircularDependency) ||
		errors.Is(err, ErrInvalidConfiguration) ||
		errors.Is(err, ErrAgentInitialization) ||
		errors.Is(err, ErrSchemaValidation) {
		return true
	}

	// Validation errors are fatal
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}
