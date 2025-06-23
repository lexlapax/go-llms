// ABOUTME: Defines domain-specific errors for the agent system
// ABOUTME: Provides typed errors for better error handling and debugging

package domain

import (
	"fmt"

	"github.com/lexlapax/go-llms/pkg/errors"
)

// Common errors
var (
	// ErrAgentNotFound is returned when an agent cannot be found
	ErrAgentNotFound = errors.NewErrorWithCode("agent_not_found", "agent not found").SetFatal(true)

	// ErrCircularDependency is returned when adding an agent would create a circular dependency
	ErrCircularDependency = errors.NewErrorWithCode("agent_circular_dependency", "circular dependency detected").SetFatal(true)

	// ErrInvalidState is returned when state is invalid or corrupted
	ErrInvalidState = errors.NewErrorWithCode("agent_invalid_state", "invalid state").SetFatal(true)

	// ErrInvalidConfiguration is returned when agent configuration is invalid
	ErrInvalidConfiguration = errors.NewErrorWithCode("agent_invalid_config", "invalid configuration").SetFatal(true)

	// ErrExecutionTimeout is returned when agent execution times out
	ErrExecutionTimeout = errors.NewErrorWithCode("agent_timeout", "execution timeout").SetRetryable(true)

	// ErrMaxRetriesExceeded is returned when maximum retries are exceeded
	ErrMaxRetriesExceeded = errors.NewErrorWithCode("agent_max_retries", "maximum retries exceeded").SetFatal(true)

	// ErrAgentInitialization is returned when agent initialization fails
	ErrAgentInitialization = errors.NewErrorWithCode("agent_init_failed", "agent initialization failed").SetFatal(true)

	// ErrToolNotFound is returned when a tool cannot be found
	ErrToolNotFound = errors.NewErrorWithCode("tool_not_found", "tool not found").SetFatal(true)

	// ErrToolExecution is returned when tool execution fails
	ErrToolExecution = errors.NewErrorWithCode("tool_execution_failed", "tool execution failed").SetRetryable(true)

	// ErrSchemaValidation is returned when schema validation fails
	ErrSchemaValidation = errors.NewErrorWithCode("schema_validation_failed", "schema validation failed").SetFatal(true)

	// ErrEventDispatch is returned when event dispatch fails
	ErrEventDispatch = errors.NewErrorWithCode("event_dispatch_failed", "event dispatch failed").SetRetryable(true)

	// ErrStateAccess is returned when state access fails
	ErrStateAccess = errors.NewErrorWithCode("state_access_failed", "state access failed").SetRetryable(true)

	// ErrArtifactAccess is returned when artifact access fails
	ErrArtifactAccess = errors.NewErrorWithCode("artifact_access_failed", "artifact access failed").SetRetryable(true)

	// ErrWorkflowExecution is returned when workflow execution fails
	ErrWorkflowExecution = errors.NewErrorWithCode("workflow_execution_failed", "workflow execution failed").SetRetryable(true)

	// ErrAgentCancelled is returned when agent execution is canceled
	ErrAgentCancelled = errors.NewErrorWithCode("agent_cancelled", "agent execution canceled")

	// ErrStateReadOnly is returned when trying to modify a read-only state
	ErrStateReadOnly = errors.NewErrorWithCode("state_readonly", "state is read-only").SetFatal(true)
)

// AgentError represents an error that occurred during agent execution.
// It includes agent identification and execution phase information
// to help with debugging and error handling decisions.
type AgentError struct {
	*errors.BaseError

	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
	Phase     string `json:"phase"` // "initialize", "before_run", "run", "after_run", "cleanup"
}

// Error implements the error interface for AgentError.
// Provides a formatted error message with agent and phase context.
func (e *AgentError) Error() string {
	return fmt.Sprintf("agent error [%s/%s] in %s: %v", e.AgentID, e.AgentName, e.Phase, e.Message)
}

// NewAgentError creates a new agent error with context information.
// Automatically sets retryability based on the execution phase:
// initialization and cleanup errors are fatal, others are retryable.
func NewAgentError(agentID, agentName, phase string, err error) *AgentError {
	baseErr := errors.Wrap(err, fmt.Sprintf("agent error in %s phase", phase))
	_ = baseErr.WithContext("agent_id", agentID).
		WithContext("agent_name", agentName).
		WithContext("phase", phase).
		WithType("AgentError")

	// Set retryability based on phase
	if phase == "initialize" || phase == "cleanup" {
		_ = baseErr.SetFatal(true)
	} else {
		_ = baseErr.SetRetryable(true)
	}

	return &AgentError{
		BaseError: baseErr,
		AgentID:   agentID,
		AgentName: agentName,
		Phase:     phase,
	}
}

// WithContext adds additional context information to the agent error.
// Returns the error for method chaining.
func (e *AgentError) WithContext(key string, value interface{}) *AgentError {
	_ = e.BaseError.WithContext(key, value)
	return e
}

// ValidationError represents a validation error for configuration or data.
// Includes the field name and value that failed validation
// for better debugging and user feedback.
type ValidationError struct {
	*errors.BaseError

	Field string      `json:"field"`
	Value interface{} `json:"value,omitempty"`
}

// Error implements the error interface for ValidationError.
// Provides a formatted error message with field context.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new validation error for a specific field.
// Validation errors are always fatal since they indicate configuration issues.
// Includes the field name, invalid value, and descriptive message.
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	baseErr := errors.Wrap(ErrSchemaValidation, message)
	_ = baseErr.WithContext("field", field).
		WithContext("value", value).
		WithType("ValidationError").
		SetFatal(true)

	return &ValidationError{
		BaseError: baseErr,
		Field:     field,
		Value:     value,
	}
}

// MultiError represents multiple errors that occurred together.
// Deprecated: Use ErrorAggregator from the errors package for new code.
// This type is maintained for backward compatibility.
type MultiError struct {
	*errors.BaseError
	Errors []error `json:"errors"`
}

// Error implements the error interface for MultiError.
// Provides a summary message about the number of errors collected.
func (e *MultiError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("multiple errors occurred (%d errors)", len(e.Errors))
}

// Add adds an error to the multi-error collection.
// Automatically updates the error count in the context.
func (e *MultiError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
		// Update context with error count
		if e.BaseError != nil {
			_ = e.WithContext("error_count", len(e.Errors))
		}
	}
}

// HasErrors returns true if the multi-error contains any errors.
// Useful for checking if any errors were collected.
func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

// Unwrap returns the collected errors as a slice.
// Implements the standard library's error unwrapping interface.
func (e *MultiError) Unwrap() []error {
	return e.Errors
}

// NewMultiError creates a new MultiError instance.
// Initializes with an empty error collection ready for use.
func NewMultiError() *MultiError {
	baseErr := errors.NewErrorWithCode("multi_error", "multiple errors occurred")
	_ = baseErr.WithType("MultiError")

	return &MultiError{
		BaseError: baseErr,
		Errors:    make([]error, 0),
	}
}

// ToolError represents an error that occurred during tool execution.
// Includes tool name, execution phase, and optional input/output context
// for comprehensive debugging information.
type ToolError struct {
	*errors.BaseError

	ToolName string      `json:"tool_name"`
	Phase    string      `json:"phase"` // "validation", "execution", "result_processing"
	Input    interface{} `json:"input,omitempty"`
	Output   interface{} `json:"output,omitempty"`
}

// Error implements the error interface for ToolError.
// Provides a formatted error message with tool and phase context.
func (e *ToolError) Error() string {
	return fmt.Sprintf("tool error [%s] in %s: %v", e.ToolName, e.Phase, e.Message)
}

// NewToolError creates a new tool error with phase information.
// Execution phase errors are retryable, validation errors are fatal.
// Automatically sets appropriate retryability based on the phase.
func NewToolError(toolName, phase string, err error) *ToolError {
	baseErr := errors.Wrap(err, fmt.Sprintf("tool error in %s phase", phase))
	_ = baseErr.WithContext("tool_name", toolName).
		WithContext("phase", phase).
		WithType("ToolError")

	// Set retryability based on phase
	if phase == "execution" {
		_ = baseErr.SetRetryable(true)
	} else {
		_ = baseErr.SetFatal(true)
	}

	return &ToolError{
		BaseError: baseErr,
		ToolName:  toolName,
		Phase:     phase,
	}
}

// NewToolErrorWithGuidance creates a new tool error with helpful guidance.
// The guidance provides suggestions for fixing the error condition.
// Useful for providing actionable feedback to users or agents.
func NewToolErrorWithGuidance(toolName, errorType, message, guidance string) error {
	// Create a custom error that includes guidance
	errMsg := message
	if guidance != "" {
		errMsg = fmt.Sprintf("%s (Guidance: %s)", message, guidance)
	}

	baseErr := errors.Wrap(ErrToolExecution, errMsg)
	_ = baseErr.WithContext("tool_name", toolName).
		WithContext("error_type", errorType).
		WithContext("guidance", guidance).
		WithType("ToolError").
		SetRetryable(true)

	return &ToolError{
		BaseError: baseErr,
		ToolName:  toolName,
		Phase:     errorType, // Using Phase field to store error type
	}
}

// WithInput adds the tool input parameters to the error context.
// Useful for debugging what parameters caused the tool to fail.
func (e *ToolError) WithInput(input interface{}) *ToolError {
	e.Input = input
	_ = e.WithContext("input", input)
	return e
}

// WithOutput adds the tool output to the error context.
// Useful when the tool succeeded but output processing failed.
func (e *ToolError) WithOutput(output interface{}) *ToolError {
	e.Output = output
	_ = e.WithContext("output", output)
	return e
}

// IsRetryable checks if an error should be retried.
// Uses the enhanced errors package and checks domain-specific error types.
// Returns true for transient errors that might succeed on retry.
func IsRetryable(err error) bool {
	// Use the enhanced errors package's retryable check
	if errors.IsRetryableError(err) {
		return true
	}

	// Check for agent errors in specific phases
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		return agentErr.Retryable
	}

	// Check for tool errors
	var toolErr *ToolError
	if errors.As(err, &toolErr) {
		return toolErr.Retryable
	}

	// Default to not retryable
	return false
}

// IsFatal checks if an error is fatal and should stop execution.
// Fatal errors indicate unrecoverable conditions like configuration issues.
// Returns true for errors that should not be retried.
func IsFatal(err error) bool {
	// Use the enhanced errors package's fatal check
	if errors.IsFatalError(err) {
		return true
	}

	// Check for agent errors
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		return agentErr.Fatal
	}

	// Check for tool errors
	var toolErr *ToolError
	if errors.As(err, &toolErr) {
		return toolErr.Fatal
	}

	// Check for validation errors
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return validationErr.Fatal
	}

	return false
}
