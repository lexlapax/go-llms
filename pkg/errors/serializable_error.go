package errors

// ABOUTME: Implementation of SerializableError with JSON support and context
// ABOUTME: Provides stack trace capture and rich error context

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// BaseError is the standard implementation of SerializableError
type BaseError struct {
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Code      string                 `json:"code,omitempty"`
	Cause     error                  `json:"-"` // Original error, not serialized directly
	CauseText string                 `json:"cause,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Stack     []StackFrame           `json:"stack,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Retryable bool                   `json:"retryable"`
	Fatal     bool                   `json:"fatal"`
	Recovery  RecoveryStrategy       `json:"-"` // Recovery strategy, not serialized
}

// NewError creates a new BaseError
func NewError(message string) *BaseError {
	return &BaseError{
		Type:      "BaseError",
		Message:   message,
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
		Stack:     captureStackTrace(2), // Skip NewError and its caller
	}
}

// NewErrorWithCode creates a new BaseError with an error code
func NewErrorWithCode(code, message string) *BaseError {
	err := NewError(message)
	err.Code = code
	err.Type = code // Use code as type if specified
	return err
}

// Wrap wraps an existing error
func Wrap(err error, message string) *BaseError {
	if err == nil {
		return nil
	}

	baseErr := &BaseError{
		Type:      "WrappedError",
		Message:   message,
		Cause:     err,
		CauseText: err.Error(),
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
		Stack:     captureStackTrace(2), // Skip Wrap and its caller
	}

	// Inherit properties from BaseError if wrapping one
	if be, ok := err.(*BaseError); ok {
		baseErr.Retryable = be.Retryable
		baseErr.Fatal = be.Fatal
		if be.Code != "" {
			baseErr.Code = be.Code
		}
		// Merge contexts
		for k, v := range be.Context {
			baseErr.Context[k] = v
		}
	}

	return baseErr
}

// Error implements the error interface
func (e *BaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Cause.Error())
	}
	return e.Message
}

// ToJSON implements SerializableError
func (e *BaseError) ToJSON() ([]byte, error) {
	return json.MarshalIndent(e, "", "  ")
}

// GetContext implements SerializableError
func (e *BaseError) GetContext() map[string]interface{} {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	return e.Context
}

// GetRecoveryStrategy implements SerializableError
func (e *BaseError) GetRecoveryStrategy() RecoveryStrategy {
	return e.Recovery
}

// WithContext adds context to the error
func (e *BaseError) WithContext(key string, value interface{}) *BaseError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithContextMap adds multiple context values
func (e *BaseError) WithContextMap(context map[string]interface{}) *BaseError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	for k, v := range context {
		e.Context[k] = v
	}
	return e
}

// WithCode sets the error code
func (e *BaseError) WithCode(code string) *BaseError {
	e.Code = code
	return e
}

// WithType sets the error type
func (e *BaseError) WithType(errorType string) *BaseError {
	e.Type = errorType
	return e
}

// SetRetryable marks the error as retryable
func (e *BaseError) SetRetryable(retryable bool) *BaseError {
	e.Retryable = retryable
	return e
}

// SetFatal marks the error as fatal
func (e *BaseError) SetFatal(fatal bool) *BaseError {
	e.Fatal = fatal
	return e
}

// WithRecovery sets the recovery strategy
func (e *BaseError) WithRecovery(strategy RecoveryStrategy) *BaseError {
	e.Recovery = strategy
	return e
}

// WithStackTrace ensures stack trace is captured
func (e *BaseError) WithStackTrace() *BaseError {
	if len(e.Stack) == 0 {
		e.Stack = captureStackTrace(2)
	}
	return e
}

// Unwrap returns the wrapped error
func (e *BaseError) Unwrap() error {
	return e.Cause
}

// Is implements error matching
func (e *BaseError) Is(target error) bool {
	if target == nil {
		return false
	}

	// Check if target is same type
	if te, ok := target.(*BaseError); ok {
		if e.Code != "" && e.Code == te.Code {
			return true
		}
		if e.Type == te.Type && e.Message == te.Message {
			return true
		}
	}

	// Check wrapped error
	if e.Cause != nil {
		return e.Cause == target
	}

	return false
}

// MarshalJSON customizes JSON serialization
func (e *BaseError) MarshalJSON() ([]byte, error) {
	type Alias BaseError
	return json.Marshal(&struct {
		*Alias
		TimestampStr string `json:"timestamp_str"`
		RecoveryName string `json:"recovery_strategy,omitempty"`
	}{
		Alias:        (*Alias)(e),
		TimestampStr: e.Timestamp.Format(time.RFC3339),
		RecoveryName: e.getRecoveryName(),
	})
}

// UnmarshalJSON customizes JSON deserialization
func (e *BaseError) UnmarshalJSON(data []byte) error {
	type Alias BaseError
	aux := &struct {
		*Alias
		TimestampStr string `json:"timestamp_str"`
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse timestamp if string provided
	if aux.TimestampStr != "" {
		if t, err := time.Parse(time.RFC3339, aux.TimestampStr); err == nil {
			e.Timestamp = t
		}
	}

	// Recreate cause error from text
	if e.CauseText != "" && e.Cause == nil {
		e.Cause = fmt.Errorf("%s", e.CauseText)
	}

	return nil
}

// getRecoveryName returns the recovery strategy name if available
func (e *BaseError) getRecoveryName() string {
	if e.Recovery != nil {
		return e.Recovery.Name()
	}
	return ""
}

// captureStackTrace captures the current stack trace
func captureStackTrace(skip int) []StackFrame {
	var frames []StackFrame

	// Capture up to 32 frames
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip+1, pcs)

	if n == 0 {
		return frames
	}

	// Get the function names and file info
	callerFrames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := callerFrames.Next()

		// Skip runtime frames
		if strings.Contains(frame.Function, "runtime.") {
			if !more {
				break
			}
			continue
		}

		frames = append(frames, StackFrame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})

		if !more {
			break
		}

		// Limit stack depth
		if len(frames) >= 20 {
			break
		}
	}

	return frames
}

// ErrorFromJSON deserializes a BaseError from JSON
func ErrorFromJSON(data []byte) (*BaseError, error) {
	var err BaseError
	if jsonErr := json.Unmarshal(data, &err); jsonErr != nil {
		return nil, jsonErr
	}
	return &err, nil
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if be, ok := err.(*BaseError); ok {
		return be.Retryable
	}

	// Check wrapped errors
	var be *BaseError
	if As(err, &be) {
		return be.Retryable
	}

	return false
}

// IsFatalError checks if an error is fatal
func IsFatalError(err error) bool {
	if err == nil {
		return false
	}

	if be, ok := err.(*BaseError); ok {
		return be.Fatal
	}

	// Check wrapped errors
	var be *BaseError
	if As(err, &be) {
		return be.Fatal
	}

	return false
}

// GetErrorContext extracts context from an error
func GetErrorContext(err error) map[string]interface{} {
	if err == nil {
		return nil
	}

	if se, ok := err.(SerializableError); ok {
		return se.GetContext()
	}

	if be, ok := err.(*BaseError); ok {
		return be.GetContext()
	}

	// Check wrapped errors
	var be *BaseError
	if As(err, &be) {
		return be.GetContext()
	}

	return nil
}

// As is a wrapper around errors.As for convenience
func As(err error, target interface{}) bool {
	if err == nil {
		return false
	}

	// Direct type assertion
	switch t := target.(type) {
	case **BaseError:
		if be, ok := err.(*BaseError); ok {
			*t = be
			return true
		}
		// Check wrapped error
		if ue, ok := err.(interface{ Unwrap() error }); ok {
			return As(ue.Unwrap(), target)
		}
	}

	return false
}
