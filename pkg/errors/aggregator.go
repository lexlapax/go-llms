package errors

// ABOUTME: Error aggregator for collecting and managing multiple errors
// ABOUTME: Provides thread-safe error collection with JSON serialization

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ErrorAggregatorImpl implements ErrorAggregator interface.
// It provides thread-safe collection and management of multiple errors,
// supporting context tracking, serialization, and error analysis.
type ErrorAggregatorImpl struct {
	errors []errorEntry
	mu     sync.RWMutex
}

// errorEntry holds an error with its context.
// It captures the error along with metadata such as timestamp
// and contextual information for debugging and analysis.
type errorEntry struct {
	Err       error                  `json:"-"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewErrorAggregator creates a new error aggregator.
// The aggregator collects multiple errors and provides methods
// to analyze, serialize, and manage them as a group.
//
// Returns a new ErrorAggregator instance.
func NewErrorAggregator() ErrorAggregator {
	return &ErrorAggregatorImpl{
		errors: make([]errorEntry, 0),
	}
}

// Add adds an error to the aggregator.
// If the error has associated context (via GetErrorContext),
// it will be automatically extracted and stored.
//
// Parameters:
//   - err: The error to add (nil errors are ignored)
func (a *ErrorAggregatorImpl) Add(err error) {
	if err == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	entry := errorEntry{
		Err:       err,
		Message:   err.Error(),
		Timestamp: time.Now(),
	}

	// Extract context if available
	if ctx := GetErrorContext(err); ctx != nil {
		entry.Context = ctx
	}

	a.errors = append(a.errors, entry)
}

// AddWithContext adds an error with context.
// The provided context is merged with any existing context
// from the error itself, with the provided context taking precedence.
//
// Parameters:
//   - err: The error to add (nil errors are ignored)
//   - context: Additional context information
func (a *ErrorAggregatorImpl) AddWithContext(err error, context map[string]interface{}) {
	if err == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	entry := errorEntry{
		Err:       err,
		Message:   err.Error(),
		Context:   context,
		Timestamp: time.Now(),
	}

	// Merge with existing context if available
	if existingCtx := GetErrorContext(err); existingCtx != nil {
		if entry.Context == nil {
			entry.Context = existingCtx
		} else {
			// Merge contexts
			for k, v := range existingCtx {
				if _, exists := entry.Context[k]; !exists {
					entry.Context[k] = v
				}
			}
		}
	}

	a.errors = append(a.errors, entry)
}

// Errors returns all collected errors.
// Returns a copy of the error slice to prevent external modifications.
//
// Returns a slice of all collected errors.
func (a *ErrorAggregatorImpl) Errors() []error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	errs := make([]error, len(a.errors))
	for i, entry := range a.errors {
		errs[i] = entry.Err
	}
	return errs
}

// Error returns the aggregated error.
// If no errors were collected, returns nil.
// If one error was collected, returns that error.
// If multiple errors were collected, returns a formatted error
// containing all error messages.
//
// Returns the aggregated error or nil.
func (a *ErrorAggregatorImpl) Error() error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(a.errors) == 0 {
		return nil
	}

	if len(a.errors) == 1 {
		return a.errors[0].Err
	}

	// Build aggregated error message
	var messages []string
	for _, entry := range a.errors {
		messages = append(messages, entry.Message)
	}

	return fmt.Errorf("multiple errors occurred (%d):\n%s", len(a.errors), strings.Join(messages, "\n"))
}

// HasErrors checks if any errors were collected.
//
// Returns true if at least one error has been added.
func (a *ErrorAggregatorImpl) HasErrors() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.errors) > 0
}

// Clear removes all errors.
// This resets the aggregator to its initial empty state.
func (a *ErrorAggregatorImpl) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errors = make([]errorEntry, 0)
}

// ToSerializable converts to a serializable error.
// It creates a structured error that includes all collected errors
// with their contexts, timestamps, and metadata. The resulting error
// can be serialized to JSON for transmission or storage.
//
// Returns a SerializableError or nil if no errors were collected.
func (a *ErrorAggregatorImpl) ToSerializable() SerializableError {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(a.errors) == 0 {
		return nil
	}

	// Create aggregated error
	aggErr := &BaseError{
		Type:      "AggregatedError",
		Message:   a.Error().Error(),
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}

	// Add error count
	aggErr.Context["error_count"] = len(a.errors)

	// Add individual errors as context
	errorDetails := make([]map[string]interface{}, len(a.errors))
	for i, entry := range a.errors {
		detail := map[string]interface{}{
			"message":   entry.Message,
			"timestamp": entry.Timestamp,
		}

		if entry.Context != nil {
			detail["context"] = entry.Context
		}

		// Check if error is serializable
		if se, ok := entry.Err.(SerializableError); ok {
			if data, err := se.ToJSON(); err == nil {
				var parsed map[string]interface{}
				if json.Unmarshal(data, &parsed) == nil {
					detail["serialized"] = parsed
				}
			}
		}

		errorDetails[i] = detail
	}

	aggErr.Context["errors"] = errorDetails

	// Check if any error is retryable or fatal
	for _, entry := range a.errors {
		if IsRetryableError(entry.Err) {
			aggErr.Retryable = true
		}
		if IsFatalError(entry.Err) {
			aggErr.Fatal = true
			break // Fatal takes precedence
		}
	}

	return aggErr
}

// AggregateErrors creates an aggregated error from multiple errors.
// This is a convenience function that creates an aggregator,
// adds all errors, and returns the aggregated result.
//
// Parameters:
//   - errs: Variable number of errors to aggregate
//
// Returns the aggregated error or nil if no errors provided.
func AggregateErrors(errs ...error) error {
	agg := NewErrorAggregator()
	for _, err := range errs {
		agg.Add(err)
	}
	return agg.Error()
}

// AggregateErrorsWithContext creates an aggregated error with context.
// Similar to AggregateErrors but adds the same context to all errors.
//
// Parameters:
//   - context: Context to add to all errors
//   - errs: Variable number of errors to aggregate
//
// Returns the aggregated error or nil if no errors provided.
func AggregateErrorsWithContext(context map[string]interface{}, errs ...error) error {
	agg := NewErrorAggregator()
	for _, err := range errs {
		agg.AddWithContext(err, context)
	}
	return agg.Error()
}

// ErrorList is a simple list of errors that implements error interface.
// Unlike ErrorAggregator, this provides a lightweight, non-thread-safe
// way to collect and manage multiple errors.
type ErrorList struct {
	errors []error
}

// NewErrorList creates a new error list.
//
// Parameters:
//   - errs: Initial errors to add to the list
//
// Returns a new ErrorList instance.
func NewErrorList(errs ...error) *ErrorList {
	return &ErrorList{
		errors: errs,
	}
}

// Add adds an error to the list.
// Nil errors are ignored.
//
// Parameters:
//   - err: The error to add
func (e *ErrorList) Add(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

// Errors returns all errors.
// Returns the internal error slice directly (not a copy).
//
// Returns all errors in the list.
func (e *ErrorList) Errors() []error {
	return e.errors
}

// Error implements error interface.
// Returns a formatted string representation of all errors.
// Single errors are returned as-is, multiple errors are
// formatted as a semicolon-separated list.
//
// Returns the error message string.
func (e *ErrorList) Error() string {
	if len(e.errors) == 0 {
		return "no errors"
	}

	if len(e.errors) == 1 {
		return e.errors[0].Error()
	}

	var messages []string
	for _, err := range e.errors {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("multiple errors: [%s]", strings.Join(messages, "; "))
}

// IsEmpty checks if the list is empty.
//
// Returns true if no errors have been added.
func (e *ErrorList) IsEmpty() bool {
	return len(e.errors) == 0
}

// First returns the first error or nil.
//
// Returns the first error in the list or nil if empty.
func (e *ErrorList) First() error {
	if len(e.errors) > 0 {
		return e.errors[0]
	}
	return nil
}

// Last returns the last error or nil.
//
// Returns the last error in the list or nil if empty.
func (e *ErrorList) Last() error {
	if len(e.errors) > 0 {
		return e.errors[len(e.errors)-1]
	}
	return nil
}
