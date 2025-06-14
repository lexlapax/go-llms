// Package errors provides enhanced error handling with serialization support
package errors

// ABOUTME: Core interfaces for enhanced error handling with serialization
// ABOUTME: Provides JSON serialization, context tracking, and recovery strategies

import (
	"encoding/json"
	"time"
)

// SerializableError defines the interface for errors that can be serialized to JSON
type SerializableError interface {
	error
	ToJSON() ([]byte, error)
	GetContext() map[string]interface{}
	GetRecoveryStrategy() RecoveryStrategy
}

// RecoveryStrategy defines how to recover from an error
type RecoveryStrategy interface {
	// Name returns the strategy name
	Name() string

	// CanRecover checks if recovery is possible for this error
	CanRecover(err error) bool

	// Recover attempts to recover from the error
	// Returns nil if recovery succeeded, or an error if it failed
	Recover(err error, context map[string]interface{}) error

	// MaxAttempts returns the maximum number of recovery attempts
	MaxAttempts() int

	// BackoffDuration returns the duration to wait before retry attempt
	BackoffDuration(attempt int) time.Duration
}

// ErrorContext provides rich context for errors
type ErrorContext interface {
	// Add adds a key-value pair to the context
	Add(key string, value interface{}) ErrorContext

	// AddAll adds multiple key-value pairs
	AddAll(values map[string]interface{}) ErrorContext

	// Get retrieves a value by key
	Get(key string) (interface{}, bool)

	// GetAll returns all context values
	GetAll() map[string]interface{}

	// WithStackTrace adds stack trace to context
	WithStackTrace() ErrorContext

	// WithTimestamp adds timestamp to context
	WithTimestamp() ErrorContext

	// Clone creates a copy of the context
	Clone() ErrorContext
}

// ContextProvider allows errors to provide additional context
type ContextProvider interface {
	// ProvideContext adds context to the error
	ProvideContext(ctx ErrorContext) ErrorContext
}

// ErrorSerializer handles error serialization
type ErrorSerializer interface {
	// Serialize converts an error to the specified format
	Serialize(err error, format string) ([]byte, error)

	// Deserialize reconstructs an error from serialized data
	Deserialize(data []byte, format string) (error, error)

	// SupportedFormats returns supported serialization formats
	SupportedFormats() []string
}

// ErrorRegistry manages error types and their serialization
type ErrorRegistry interface {
	// Register registers an error type with its serializer
	Register(errorType string, serializer ErrorSerializer) error

	// Get retrieves a serializer for an error type
	Get(errorType string) (ErrorSerializer, bool)

	// SerializeAny serializes any registered error type
	SerializeAny(err error) ([]byte, error)

	// DeserializeAny deserializes to the appropriate error type
	DeserializeAny(data []byte) (error, error)
}

// ErrorMetadata contains metadata about an error
type ErrorMetadata struct {
	Type       string                 `json:"type"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Context    map[string]interface{} `json:"context,omitempty"`
	StackTrace []StackFrame           `json:"stack_trace,omitempty"`
	Retryable  bool                   `json:"retryable"`
	Fatal      bool                   `json:"fatal"`
}

// StackFrame represents a single frame in a stack trace
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// ErrorWrapper wraps errors with additional functionality
type ErrorWrapper interface {
	// Wrap wraps an error with context
	Wrap(err error, message string) SerializableError

	// WrapWithContext wraps with message and context
	WrapWithContext(err error, message string, context map[string]interface{}) SerializableError

	// WrapWithRecovery wraps with recovery strategy
	WrapWithRecovery(err error, message string, strategy RecoveryStrategy) SerializableError
}

// ErrorMatcher matches errors based on patterns
type ErrorMatcher interface {
	// Match checks if an error matches the pattern
	Match(err error) bool

	// Extract extracts information from a matched error
	Extract(err error) map[string]interface{}
}

// ErrorAggregator collects multiple errors
type ErrorAggregator interface {
	// Add adds an error to the aggregator
	Add(err error)

	// AddWithContext adds an error with context
	AddWithContext(err error, context map[string]interface{})

	// Errors returns all collected errors
	Errors() []error

	// Error returns the aggregated error
	Error() error

	// HasErrors checks if any errors were collected
	HasErrors() bool

	// Clear removes all errors
	Clear()

	// ToSerializable converts to a serializable error
	ToSerializable() SerializableError
}

// Custom JSON marshaling for ErrorMetadata
func (em ErrorMetadata) MarshalJSON() ([]byte, error) {
	type Alias ErrorMetadata
	return json.Marshal(&struct {
		Alias
		TimestampStr string `json:"timestamp_str"`
	}{
		Alias:        (Alias)(em),
		TimestampStr: em.Timestamp.Format(time.RFC3339),
	})
}
