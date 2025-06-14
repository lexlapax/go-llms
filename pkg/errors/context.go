package errors

// ABOUTME: Error context implementation for enriching errors with metadata
// ABOUTME: Provides thread-safe context management and builder pattern

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ErrorContextImpl implements ErrorContext interface
type ErrorContextImpl struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewErrorContext creates a new error context
func NewErrorContext() ErrorContext {
	return &ErrorContextImpl{
		data: make(map[string]interface{}),
	}
}

// Add adds a key-value pair to the context
func (c *ErrorContextImpl) Add(key string, value interface{}) ErrorContext {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return c
}

// AddAll adds multiple key-value pairs
func (c *ErrorContextImpl) AddAll(values map[string]interface{}) ErrorContext {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range values {
		c.data[k] = v
	}
	return c
}

// Get retrieves a value by key
func (c *ErrorContextImpl) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

// GetAll returns all context values
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

// WithStackTrace adds stack trace to context
func (c *ErrorContextImpl) WithStackTrace() ErrorContext {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data["stack_trace"] = captureStackTrace(2)
	return c
}

// WithTimestamp adds timestamp to context
func (c *ErrorContextImpl) WithTimestamp() ErrorContext {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data["timestamp"] = time.Now()
	return c
}

// Clone creates a copy of the context
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

// ContextualError wraps an error with context
type ContextualError struct {
	err     error
	context ErrorContext
}

// NewContextualError creates a new contextual error
func NewContextualError(err error) *ContextualError {
	return &ContextualError{
		err:     err,
		context: NewErrorContext().WithTimestamp(),
	}
}

// Error implements error interface
func (e *ContextualError) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped error
func (e *ContextualError) Unwrap() error {
	return e.err
}

// Context returns the error context
func (e *ContextualError) Context() ErrorContext {
	return e.context
}

// WithContext adds context to the error
func (e *ContextualError) WithContext(key string, value interface{}) *ContextualError {
	e.context.Add(key, value)
	return e
}

// WithContextMap adds multiple context values
func (e *ContextualError) WithContextMap(values map[string]interface{}) *ContextualError {
	e.context.AddAll(values)
	return e
}

// ProvideContext implements ContextProvider
func (e *ContextualError) ProvideContext(ctx ErrorContext) ErrorContext {
	// Merge existing context with provided context
	for k, v := range e.context.GetAll() {
		ctx.Add(k, v)
	}
	return ctx
}

// ErrorBuilder provides a fluent interface for building errors
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

// NewErrorBuilder creates a new error builder
func NewErrorBuilder(message string) *ErrorBuilder {
	return &ErrorBuilder{
		message: message,
		context: make(map[string]interface{}),
	}
}

// WithCause sets the cause error
func (b *ErrorBuilder) WithCause(err error) *ErrorBuilder {
	b.cause = err
	return b
}

// WithCode sets the error code
func (b *ErrorBuilder) WithCode(code string) *ErrorBuilder {
	b.code = code
	return b
}

// WithType sets the error type
func (b *ErrorBuilder) WithType(errorType string) *ErrorBuilder {
	b.errorType = errorType
	return b
}

// WithContext adds context
func (b *ErrorBuilder) WithContext(key string, value interface{}) *ErrorBuilder {
	b.context[key] = value
	return b
}

// WithContextMap adds multiple context values
func (b *ErrorBuilder) WithContextMap(values map[string]interface{}) *ErrorBuilder {
	for k, v := range values {
		b.context[k] = v
	}
	return b
}

// WithRetryable sets retryable flag
func (b *ErrorBuilder) WithRetryable(retryable bool) *ErrorBuilder {
	b.retryable = retryable
	return b
}

// WithFatal sets fatal flag
func (b *ErrorBuilder) WithFatal(fatal bool) *ErrorBuilder {
	b.fatal = fatal
	return b
}

// WithRecovery sets recovery strategy
func (b *ErrorBuilder) WithRecovery(strategy RecoveryStrategy) *ErrorBuilder {
	b.recovery = strategy
	return b
}

// Build creates the error
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

// EnrichError adds context to any error
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

// EnrichErrorWithOperation adds operation context
func EnrichErrorWithOperation(err error, operation string) error {
	return EnrichError(err, map[string]interface{}{
		"operation": operation,
		"timestamp": time.Now(),
	})
}

// EnrichErrorWithRequest adds request context
func EnrichErrorWithRequest(err error, method, url string, statusCode int) error {
	return EnrichError(err, map[string]interface{}{
		"request_method": method,
		"request_url":    url,
		"status_code":    statusCode,
		"timestamp":      time.Now(),
	})
}

// EnrichErrorWithResource adds resource context
func EnrichErrorWithResource(err error, resourceType, resourceID string) error {
	return EnrichError(err, map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"timestamp":     time.Now(),
	})
}

// RuntimeContext captures runtime information
type RuntimeContext struct {
	GoVersion    string                 `json:"go_version"`
	GOOS         string                 `json:"goos"`
	GOARCH       string                 `json:"goarch"`
	NumCPU       int                    `json:"num_cpu"`
	NumGoroutine int                    `json:"num_goroutine"`
	MemStats     map[string]interface{} `json:"mem_stats,omitempty"`
}

// CaptureRuntimeContext captures current runtime information
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

// EnrichErrorWithRuntime adds runtime context
func EnrichErrorWithRuntime(err error) error {
	ctx := CaptureRuntimeContext()
	return EnrichError(err, map[string]interface{}{
		"runtime": ctx,
	})
}
