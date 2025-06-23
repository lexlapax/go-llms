package errors

// ABOUTME: Error context implementation for enriching errors with metadata
// ABOUTME: Provides thread-safe context management and builder pattern

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ErrorContextImpl implements ErrorContext interface.
// It provides thread-safe storage and management of error context data,
// supporting key-value pairs and common enrichment patterns.
type ErrorContextImpl struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewErrorContext creates a new error context.
// The context starts empty and can be enriched using the fluent API.
//
// Returns a new ErrorContext instance.
func NewErrorContext() ErrorContext {
	return &ErrorContextImpl{
		data: make(map[string]interface{}),
	}
}

// Add adds a key-value pair to the context.
// Thread-safe operation that supports method chaining.
//
// Parameters:
//   - key: The context key
//   - value: The context value
//
// Returns self for method chaining.
func (c *ErrorContextImpl) Add(key string, value interface{}) ErrorContext {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return c
}

// AddAll adds multiple key-value pairs.
// Thread-safe batch operation for adding multiple context values.
//
// Parameters:
//   - values: Map of key-value pairs to add
//
// Returns self for method chaining.
func (c *ErrorContextImpl) AddAll(values map[string]interface{}) ErrorContext {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range values {
		c.data[k] = v
	}
	return c
}

// Get retrieves a value by key.
// Thread-safe read operation.
//
// Parameters:
//   - key: The context key to retrieve
//
// Returns the value and a boolean indicating if the key exists.
func (c *ErrorContextImpl) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

// GetAll returns all context values.
// Returns a copy of the internal map to prevent external modifications.
// Thread-safe operation.
//
// Returns a copy of all context key-value pairs.
func (c *ErrorContextImpl) GetAll() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Create a copy to prevent external modifications
	copy := make(map[string]interface{}, len(c.data))
	for k, v := range c.data {
		copy[k] = v
	}
	return copy
}

// WithStackTrace adds stack trace to context.
// Captures the current stack trace and stores it under "stack_trace" key.
//
// Returns self for method chaining.
func (c *ErrorContextImpl) WithStackTrace() ErrorContext {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data["stack_trace"] = captureStackTrace(2)
	return c
}

// WithTimestamp adds timestamp to context.
// Stores the current time under "timestamp" key.
//
// Returns self for method chaining.
func (c *ErrorContextImpl) WithTimestamp() ErrorContext {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data["timestamp"] = time.Now()
	return c
}

// Clone creates a copy of the context.
// The clone is independent and modifications won't affect the original.
//
// Returns a new ErrorContext with copied data.
func (c *ErrorContextImpl) Clone() ErrorContext {
	c.mu.RLock()
	defer c.mu.RUnlock()
	newContext := &ErrorContextImpl{
		data: make(map[string]interface{}, len(c.data)),
	}
	for k, v := range c.data {
		newContext.data[k] = v
	}
	return newContext
}

// ContextualError wraps an error with context.
// It implements the error interface and provides methods
// to enrich errors with contextual information.
type ContextualError struct {
	err     error
	context ErrorContext
}

// NewContextualError creates a new contextual error.
// The error is automatically enriched with a timestamp.
//
// Parameters:
//   - err: The error to wrap with context
//
// Returns a new ContextualError instance.
func NewContextualError(err error) *ContextualError {
	return &ContextualError{
		err:     err,
		context: NewErrorContext().WithTimestamp(),
	}
}

// Error implements error interface.
// Returns the error message from the wrapped error.
//
// Returns the error message string.
func (e *ContextualError) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped error.
// Supports Go 1.13+ error unwrapping.
//
// Returns the wrapped error.
func (e *ContextualError) Unwrap() error {
	return e.err
}

// Context returns the error context.
//
// Returns the ErrorContext associated with this error.
func (e *ContextualError) Context() ErrorContext {
	return e.context
}

// WithContext adds context to the error.
// Fluent method for adding contextual information.
//
// Parameters:
//   - key: The context key
//   - value: The context value
//
// Returns self for method chaining.
func (e *ContextualError) WithContext(key string, value interface{}) *ContextualError {
	e.context.Add(key, value)
	return e
}

// WithContextMap adds multiple context values.
// Batch operation for adding multiple context entries.
//
// Parameters:
//   - values: Map of key-value pairs to add
//
// Returns self for method chaining.
func (e *ContextualError) WithContextMap(values map[string]interface{}) *ContextualError {
	e.context.AddAll(values)
	return e
}

// ProvideContext implements ContextProvider.
// Merges this error's context into the provided context.
//
// Parameters:
//   - ctx: The context to merge into
//
// Returns the enriched context.
func (e *ContextualError) ProvideContext(ctx ErrorContext) ErrorContext {
	// Merge existing context with provided context
	for k, v := range e.context.GetAll() {
		ctx.Add(k, v)
	}
	return ctx
}

// ErrorBuilder provides a fluent interface for building errors.
// It supports setting various error properties including message,
// cause, code, type, context, and recovery strategies.
type ErrorBuilder struct {
	message   string
	cause     error
	code      string
	errorType string
	context   map[string]interface{}
	retryable bool
	fatal     bool
	recovery  RecoveryStrategy
}

// NewErrorBuilder creates a new error builder.
//
// Parameters:
//   - message: The error message
//
// Returns a new ErrorBuilder instance.
func NewErrorBuilder(message string) *ErrorBuilder {
	return &ErrorBuilder{
		message: message,
		context: make(map[string]interface{}),
	}
}

// WithCause sets the cause error.
// Used for error chaining to track the root cause.
//
// Parameters:
//   - err: The cause error
//
// Returns self for method chaining.
func (b *ErrorBuilder) WithCause(err error) *ErrorBuilder {
	b.cause = err
	return b
}

// WithCode sets the error code.
// Error codes can be used for programmatic error handling.
//
// Parameters:
//   - code: The error code
//
// Returns self for method chaining.
func (b *ErrorBuilder) WithCode(code string) *ErrorBuilder {
	b.code = code
	return b
}

// WithType sets the error type.
// Types categorize errors (e.g., "ValidationError", "NetworkError").
//
// Parameters:
//   - errorType: The error type
//
// Returns self for method chaining.
func (b *ErrorBuilder) WithType(errorType string) *ErrorBuilder {
	b.errorType = errorType
	return b
}

// WithContext adds context.
// Adds a single key-value pair to the error context.
//
// Parameters:
//   - key: The context key
//   - value: The context value
//
// Returns self for method chaining.
func (b *ErrorBuilder) WithContext(key string, value interface{}) *ErrorBuilder {
	b.context[key] = value
	return b
}

// WithContextMap adds multiple context values.
// Batch operation for adding multiple context entries.
//
// Parameters:
//   - values: Map of key-value pairs to add
//
// Returns self for method chaining.
func (b *ErrorBuilder) WithContextMap(values map[string]interface{}) *ErrorBuilder {
	for k, v := range values {
		b.context[k] = v
	}
	return b
}

// WithRetryable sets retryable flag.
// Indicates whether the operation that caused this error can be retried.
//
// Parameters:
//   - retryable: Whether the error is retryable
//
// Returns self for method chaining.
func (b *ErrorBuilder) WithRetryable(retryable bool) *ErrorBuilder {
	b.retryable = retryable
	return b
}

// WithFatal sets fatal flag.
// Fatal errors indicate unrecoverable conditions.
//
// Parameters:
//   - fatal: Whether the error is fatal
//
// Returns self for method chaining.
func (b *ErrorBuilder) WithFatal(fatal bool) *ErrorBuilder {
	b.fatal = fatal
	return b
}

// WithRecovery sets recovery strategy.
// Defines how the system should attempt to recover from this error.
//
// Parameters:
//   - strategy: The recovery strategy
//
// Returns self for method chaining.
func (b *ErrorBuilder) WithRecovery(strategy RecoveryStrategy) *ErrorBuilder {
	b.recovery = strategy
	return b
}

// Build creates the error.
// Constructs a BaseError with all configured properties
// and captures the current stack trace.
//
// Returns a new BaseError instance.
func (b *ErrorBuilder) Build() *BaseError {
	err := &BaseError{
		Type:      b.errorType,
		Message:   b.message,
		Code:      b.code,
		Cause:     b.cause,
		Context:   b.context,
		Timestamp: time.Now(),
		Retryable: b.retryable,
		Fatal:     b.fatal,
		Recovery:  b.recovery,
		Stack:     captureStackTrace(2),
	}

	if b.errorType == "" {
		err.Type = "Error"
	}

	if b.cause != nil {
		err.CauseText = b.cause.Error()
	}

	return err
}

// EnrichError adds context to any error.
// If the error is already a BaseError, context is added directly.
// Otherwise, the error is wrapped in a BaseError first.
//
// Parameters:
//   - err: The error to enrich
//   - context: Context to add
//
// Returns the enriched error or nil if err is nil.
func EnrichError(err error, context map[string]interface{}) error {
	if err == nil {
		return nil
	}

	// If it's already a BaseError, add context directly
	if be, ok := err.(*BaseError); ok {
		return be.WithContextMap(context)
	}

	// Otherwise wrap it
	return Wrap(err, err.Error()).WithContextMap(context)
}

// EnrichErrorWithOperation adds operation context.
// Convenience function for adding operation name and timestamp.
//
// Parameters:
//   - err: The error to enrich
//   - operation: The operation name
//
// Returns the enriched error.
func EnrichErrorWithOperation(err error, operation string) error {
	return EnrichError(err, map[string]interface{}{
		"operation": operation,
		"timestamp": time.Now(),
	})
}

// EnrichErrorWithRequest adds request context.
// Convenience function for adding HTTP request information.
//
// Parameters:
//   - err: The error to enrich
//   - method: HTTP method
//   - url: Request URL
//   - statusCode: HTTP status code
//
// Returns the enriched error.
func EnrichErrorWithRequest(err error, method, url string, statusCode int) error {
	return EnrichError(err, map[string]interface{}{
		"request_method": method,
		"request_url":    url,
		"status_code":    statusCode,
		"timestamp":      time.Now(),
	})
}

// EnrichErrorWithResource adds resource context.
// Convenience function for adding resource identification.
//
// Parameters:
//   - err: The error to enrich
//   - resourceType: Type of resource
//   - resourceID: Resource identifier
//
// Returns the enriched error.
func EnrichErrorWithResource(err error, resourceType, resourceID string) error {
	return EnrichError(err, map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"timestamp":     time.Now(),
	})
}

// RuntimeContext captures runtime information.
// Contains Go runtime details including version, architecture,
// CPU count, goroutine count, and memory statistics.
type RuntimeContext struct {
	GoVersion    string                 `json:"go_version"`
	GOOS         string                 `json:"goos"`
	GOARCH       string                 `json:"goarch"`
	NumCPU       int                    `json:"num_cpu"`
	NumGoroutine int                    `json:"num_goroutine"`
	MemStats     map[string]interface{} `json:"mem_stats,omitempty"`
}

// CaptureRuntimeContext captures current runtime information.
// Gathers runtime statistics including memory usage, goroutine count,
// and system information.
//
// Returns a RuntimeContext with current runtime information.
func CaptureRuntimeContext() RuntimeContext {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return RuntimeContext{
		GoVersion:    runtime.Version(),
		GOOS:         runtime.GOOS,
		GOARCH:       runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		MemStats: map[string]interface{}{
			"alloc":       fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024),
			"total_alloc": fmt.Sprintf("%.2f MB", float64(m.TotalAlloc)/1024/1024),
			"sys":         fmt.Sprintf("%.2f MB", float64(m.Sys)/1024/1024),
			"num_gc":      m.NumGC,
		},
	}
}

// EnrichErrorWithRuntime adds runtime context.
// Captures and adds current runtime information to the error.
// Useful for debugging performance and resource issues.
//
// Parameters:
//   - err: The error to enrich
//
// Returns the enriched error.
func EnrichErrorWithRuntime(err error) error {
	ctx := CaptureRuntimeContext()
	return EnrichError(err, map[string]interface{}{
		"runtime": ctx,
	})
}
