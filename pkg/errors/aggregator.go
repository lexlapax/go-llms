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

// ErrorAggregatorImpl implements ErrorAggregator interface
type ErrorAggregatorImpl struct {
	errors []errorEntry
	mu     sync.RWMutex
}

// errorEntry holds an error with its context
type errorEntry struct {
	Err       error                  `json:"-"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewErrorAggregator creates a new error aggregator
func NewErrorAggregator() ErrorAggregator {
	return &ErrorAggregatorImpl{
		errors: make([]errorEntry, 0),
	}
}

// Add adds an error to the aggregator
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

// AddWithContext adds an error with context
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

// Errors returns all collected errors
func (a *ErrorAggregatorImpl) Errors() []error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	errs := make([]error, len(a.errors))
	for i, entry := range a.errors {
		errs[i] = entry.Err
	}
	return errs
}

// Error returns the aggregated error
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

// HasErrors checks if any errors were collected
func (a *ErrorAggregatorImpl) HasErrors() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.errors) > 0
}

// Clear removes all errors
func (a *ErrorAggregatorImpl) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errors = make([]errorEntry, 0)
}

// ToSerializable converts to a serializable error
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

// AggregateErrors creates an aggregated error from multiple errors
func AggregateErrors(errs ...error) error {
	agg := NewErrorAggregator()
	for _, err := range errs {
		agg.Add(err)
	}
	return agg.Error()
}

// AggregateErrorsWithContext creates an aggregated error with context
func AggregateErrorsWithContext(context map[string]interface{}, errs ...error) error {
	agg := NewErrorAggregator()
	for _, err := range errs {
		agg.AddWithContext(err, context)
	}
	return agg.Error()
}

// ErrorList is a simple list of errors that implements error interface
type ErrorList struct {
	errors []error
}

// NewErrorList creates a new error list
func NewErrorList(errs ...error) *ErrorList {
	return &ErrorList{
		errors: errs,
	}
}

// Add adds an error to the list
func (e *ErrorList) Add(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}

// Errors returns all errors
func (e *ErrorList) Errors() []error {
	return e.errors
}

// Error implements error interface
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

// IsEmpty checks if the list is empty
func (e *ErrorList) IsEmpty() bool {
	return len(e.errors) == 0
}

// First returns the first error or nil
func (e *ErrorList) First() error {
	if len(e.errors) > 0 {
		return e.errors[0]
	}
	return nil
}

// Last returns the last error or nil
func (e *ErrorList) Last() error {
	if len(e.errors) > 0 {
		return e.errors[len(e.errors)-1]
	}
	return nil
}
