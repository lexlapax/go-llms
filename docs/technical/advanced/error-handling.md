# Error Handling: Error Types and Recovery Strategies

> **[Project Root](/) / [Documentation](/docs/) / [Technical Documentation](/docs/technical/) / [Advanced Topics](/docs/technical/advanced/) / Error Handling**

Comprehensive guide to error handling in Go-LLMs, covering structured error types, error classification, recovery strategies, retry mechanisms, circuit breakers, graceful degradation patterns, and robust error reporting for building resilient LLM applications.

## Error Architecture and Classification

### 1. Structured Error System

```go
// BaseError provides the foundation for all Go-LLMs errors
type BaseError struct {
    Type        ErrorType              `json:"type"`
    Code        string                 `json:"code"`
    Message     string                 `json:"message"`
    Details     map[string]interface{} `json:"details,omitempty"`
    Cause       error                  `json:"-"`
    Timestamp   time.Time              `json:"timestamp"`
    Context     ErrorContext           `json:"context,omitempty"`
    Retryable   bool                   `json:"retryable"`
    Severity    ErrorSeverity          `json:"severity"`
}

type ErrorType string

const (
    ErrorTypeValidation   ErrorType = "validation"
    ErrorTypeAuth        ErrorType = "authentication"
    ErrorTypePermission  ErrorType = "permission"
    ErrorTypeRateLimit   ErrorType = "rate_limit"
    ErrorTypeQuota       ErrorType = "quota"
    ErrorTypeNetwork     ErrorType = "network"
    ErrorTypeTimeout     ErrorType = "timeout"
    ErrorTypeProvider    ErrorType = "provider"
    ErrorTypeInternal    ErrorType = "internal"
    ErrorTypeResource    ErrorType = "resource"
    ErrorTypeDependency  ErrorType = "dependency"
    ErrorTypeConfiguration ErrorType = "configuration"
)

type ErrorSeverity string

const (
    SeverityLow      ErrorSeverity = "low"
    SeverityMedium   ErrorSeverity = "medium"
    SeverityHigh     ErrorSeverity = "high"
    SeverityCritical ErrorSeverity = "critical"
)

type ErrorContext struct {
    RequestID   string                 `json:"request_id,omitempty"`
    UserID      string                 `json:"user_id,omitempty"`
    Provider    string                 `json:"provider,omitempty"`
    Model       string                 `json:"model,omitempty"`
    Operation   string                 `json:"operation,omitempty"`
    Component   string                 `json:"component,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Error implements the error interface
func (e *BaseError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s:%s] %s (caused by: %v)", e.Type, e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *BaseError) Unwrap() error {
    return e.Cause
}

// Is checks if the error matches a target error type
func (e *BaseError) Is(target error) bool {
    if t, ok := target.(*BaseError); ok {
        return e.Type == t.Type && e.Code == t.Code
    }
    return false
}

// As attempts to convert the error to a specific type
func (e *BaseError) As(target interface{}) bool {
    if t, ok := target.(**BaseError); ok {
        *t = e
        return true
    }
    return false
}

// WithContext adds context to an error
func (e *BaseError) WithContext(ctx ErrorContext) *BaseError {
    e.Context = ctx
    return e
}

// WithCause sets the underlying cause
func (e *BaseError) WithCause(cause error) *BaseError {
    e.Cause = cause
    return e
}

// WithDetail adds a detail field
func (e *BaseError) WithDetail(key string, value interface{}) *BaseError {
    if e.Details == nil {
        e.Details = make(map[string]interface{})
    }
    e.Details[key] = value
    return e
}
```

### 2. Specific Error Types

```go
// ValidationError represents input validation failures
type ValidationError struct {
    *BaseError
    Field       string      `json:"field"`
    Value       interface{} `json:"value,omitempty"`
    Constraint  string      `json:"constraint"`
    Violations  []Violation `json:"violations,omitempty"`
}

type Violation struct {
    Field   string      `json:"field"`
    Value   interface{} `json:"value,omitempty"`
    Rule    string      `json:"rule"`
    Message string      `json:"message"`
}

// NewValidationError creates a validation error
func NewValidationError(field, constraint, message string) *ValidationError {
    return &ValidationError{
        BaseError: &BaseError{
            Type:      ErrorTypeValidation,
            Code:      "VALIDATION_FAILED",
            Message:   message,
            Timestamp: time.Now(),
            Retryable: false,
            Severity:  SeverityMedium,
        },
        Field:      field,
        Constraint: constraint,
    }
}

// ProviderError represents LLM provider-specific errors
type ProviderError struct {
    *BaseError
    Provider     string `json:"provider"`
    Model        string `json:"model,omitempty"`
    StatusCode   int    `json:"status_code,omitempty"`
    ProviderCode string `json:"provider_code,omitempty"`
    Quota        *QuotaInfo `json:"quota,omitempty"`
}

type QuotaInfo struct {
    Limit     int64     `json:"limit"`
    Used      int64     `json:"used"`
    Remaining int64     `json:"remaining"`
    ResetTime time.Time `json:"reset_time"`
}

// NewProviderError creates a provider-specific error
func NewProviderError(provider, code, message string, retryable bool) *ProviderError {
    return &ProviderError{
        BaseError: &BaseError{
            Type:      ErrorTypeProvider,
            Code:      code,
            Message:   message,
            Timestamp: time.Now(),
            Retryable: retryable,
            Severity:  SeverityHigh,
        },
        Provider: provider,
    }
}

// RateLimitError represents rate limiting errors
type RateLimitError struct {
    *BaseError
    Limit       int           `json:"limit"`
    Remaining   int           `json:"remaining"`
    ResetTime   time.Time     `json:"reset_time"`
    RetryAfter  time.Duration `json:"retry_after"`
    Window      time.Duration `json:"window"`
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(limit, remaining int, resetTime time.Time) *RateLimitError {
    retryAfter := time.Until(resetTime)
    if retryAfter < 0 {
        retryAfter = time.Minute // Default retry after 1 minute
    }
    
    return &RateLimitError{
        BaseError: &BaseError{
            Type:      ErrorTypeRateLimit,
            Code:      "RATE_LIMIT_EXCEEDED",
            Message:   fmt.Sprintf("rate limit exceeded: %d/%d requests", limit-remaining, limit),
            Timestamp: time.Now(),
            Retryable: true,
            Severity:  SeverityMedium,
        },
        Limit:      limit,
        Remaining:  remaining,
        ResetTime:  resetTime,
        RetryAfter: retryAfter,
    }
}

// NetworkError represents network connectivity issues
type NetworkError struct {
    *BaseError
    Host        string        `json:"host,omitempty"`
    Port        int           `json:"port,omitempty"`
    Timeout     time.Duration `json:"timeout,omitempty"`
    DNSFailure  bool          `json:"dns_failure,omitempty"`
    ConnRefused bool          `json:"connection_refused,omitempty"`
}

// NewNetworkError creates a network error
func NewNetworkError(host string, port int, cause error) *NetworkError {
    return &NetworkError{
        BaseError: &BaseError{
            Type:      ErrorTypeNetwork,
            Code:      "NETWORK_ERROR",
            Message:   fmt.Sprintf("network error connecting to %s:%d", host, port),
            Cause:     cause,
            Timestamp: time.Now(),
            Retryable: true,
            Severity:  SeverityHigh,
        },
        Host: host,
        Port: port,
    }
}

// ConfigurationError represents configuration issues
type ConfigurationError struct {
    *BaseError
    ConfigPath string      `json:"config_path,omitempty"`
    Parameter  string      `json:"parameter,omitempty"`
    Expected   interface{} `json:"expected,omitempty"`
    Actual     interface{} `json:"actual,omitempty"`
}

// NewConfigurationError creates a configuration error
func NewConfigurationError(parameter, message string) *ConfigurationError {
    return &ConfigurationError{
        BaseError: &BaseError{
            Type:      ErrorTypeConfiguration,
            Code:      "INVALID_CONFIGURATION",
            Message:   message,
            Timestamp: time.Now(),
            Retryable: false,
            Severity:  SeverityHigh,
        },
        Parameter: parameter,
    }
}
```

### 3. Error Context and Tracing

```go
// ErrorTracker provides error tracking and correlation
type ErrorTracker struct {
    correlationID string
    breadcrumbs   []Breadcrumb
    context       map[string]interface{}
    mu            sync.RWMutex
}

type Breadcrumb struct {
    Timestamp time.Time              `json:"timestamp"`
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Category  string                 `json:"category"`
    Data      map[string]interface{} `json:"data,omitempty"`
}

// NewErrorTracker creates a new error tracker
func NewErrorTracker() *ErrorTracker {
    return &ErrorTracker{
        correlationID: generateCorrelationID(),
        breadcrumbs:   make([]Breadcrumb, 0, 50),
        context:       make(map[string]interface{}),
    }
}

// AddBreadcrumb adds a breadcrumb to the error trail
func (et *ErrorTracker) AddBreadcrumb(level, message, category string, data map[string]interface{}) {
    et.mu.Lock()
    defer et.mu.Unlock()
    
    breadcrumb := Breadcrumb{
        Timestamp: time.Now(),
        Level:     level,
        Message:   message,
        Category:  category,
        Data:      data,
    }
    
    et.breadcrumbs = append(et.breadcrumbs, breadcrumb)
    
    // Keep only last 50 breadcrumbs
    if len(et.breadcrumbs) > 50 {
        et.breadcrumbs = et.breadcrumbs[1:]
    }
}

// SetContext sets contextual information
func (et *ErrorTracker) SetContext(key string, value interface{}) {
    et.mu.Lock()
    defer et.mu.Unlock()
    et.context[key] = value
}

// WrapError wraps an error with tracking information
func (et *ErrorTracker) WrapError(err error, operation string) error {
    if err == nil {
        return nil
    }
    
    et.mu.RLock()
    defer et.mu.RUnlock()
    
    // If it's already a BaseError, enhance it
    if baseErr, ok := err.(*BaseError); ok {
        baseErr.Context.RequestID = et.correlationID
        baseErr.Context.Operation = operation
        if baseErr.Context.Metadata == nil {
            baseErr.Context.Metadata = make(map[string]interface{})
        }
        baseErr.Context.Metadata["breadcrumbs"] = et.breadcrumbs
        baseErr.Context.Metadata["context"] = et.context
        return baseErr
    }
    
    // Wrap other errors
    return &BaseError{
        Type:      ErrorTypeInternal,
        Code:      "WRAPPED_ERROR",
        Message:   err.Error(),
        Cause:     err,
        Timestamp: time.Now(),
        Context: ErrorContext{
            RequestID: et.correlationID,
            Operation: operation,
            Metadata: map[string]interface{}{
                "breadcrumbs": et.breadcrumbs,
                "context":     et.context,
            },
        },
        Retryable: false,
        Severity:  SeverityMedium,
    }
}
```

## Recovery Strategies

### 1. Retry Mechanisms

```go
// RetryPolicy defines retry behavior
type RetryPolicy struct {
    MaxAttempts     int           `yaml:"max_attempts" json:"max_attempts"`
    InitialDelay    time.Duration `yaml:"initial_delay" json:"initial_delay"`
    MaxDelay        time.Duration `yaml:"max_delay" json:"max_delay"`
    BackoffFactor   float64       `yaml:"backoff_factor" json:"backoff_factor"`
    Jitter          bool          `yaml:"jitter" json:"jitter"`
    RetryableErrors []string      `yaml:"retryable_errors" json:"retryable_errors"`
}

// RetryableExecutor executes operations with retry logic
type RetryableExecutor struct {
    policy   RetryPolicy
    logger   Logger
    metrics  *RetryMetrics
}

type RetryMetrics struct {
    TotalAttempts    int64
    SuccessfulRetries int64
    FailedRetries    int64
    MaxRetriesReached int64
}

// Execute runs an operation with retry logic
func (re *RetryableExecutor) Execute(ctx context.Context, operation Operation) (interface{}, error) {
    var lastError error
    
    for attempt := 1; attempt <= re.policy.MaxAttempts; attempt++ {
        // Add attempt tracking to context
        attemptCtx := context.WithValue(ctx, "attempt", attempt)
        
        atomic.AddInt64(&re.metrics.TotalAttempts, 1)
        
        result, err := operation(attemptCtx)
        if err == nil {
            if attempt > 1 {
                atomic.AddInt64(&re.metrics.SuccessfulRetries, 1)
            }
            return result, nil
        }
        
        lastError = err
        
        // Check if error is retryable
        if !re.isRetryable(err) {
            return nil, err
        }
        
        // Don't retry on last attempt
        if attempt == re.policy.MaxAttempts {
            atomic.AddInt64(&re.metrics.MaxRetriesReached, 1)
            break
        }
        
        // Calculate delay for next attempt
        delay := re.calculateDelay(attempt)
        
        re.logger.Warn("operation failed, retrying",
            "attempt", attempt,
            "max_attempts", re.policy.MaxAttempts,
            "delay", delay,
            "error", err,
        )
        
        // Wait before retry
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(delay):
            atomic.AddInt64(&re.metrics.FailedRetries, 1)
        }
    }
    
    return nil, fmt.Errorf("operation failed after %d attempts: %w", re.policy.MaxAttempts, lastError)
}

// isRetryable determines if an error should trigger a retry
func (re *RetryableExecutor) isRetryable(err error) bool {
    // Check if it's a BaseError with retryable flag
    if baseErr, ok := err.(*BaseError); ok {
        return baseErr.Retryable
    }
    
    // Check specific error types
    switch {
    case isNetworkError(err):
        return true
    case isTimeoutError(err):
        return true
    case isRateLimitError(err):
        return true
    case isTemporaryServerError(err):
        return true
    default:
        return false
    }
}

// calculateDelay computes the delay for the next retry attempt
func (re *RetryableExecutor) calculateDelay(attempt int) time.Duration {
    delay := time.Duration(float64(re.policy.InitialDelay) * math.Pow(re.policy.BackoffFactor, float64(attempt-1)))
    
    if delay > re.policy.MaxDelay {
        delay = re.policy.MaxDelay
    }
    
    // Add jitter if enabled
    if re.policy.Jitter {
        jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
        delay = delay + jitter
    }
    
    return delay
}

type Operation func(ctx context.Context) (interface{}, error)

// RetryableProvider wraps a provider with retry logic
type RetryableProvider struct {
    Provider
    executor *RetryableExecutor
}

func (rp *RetryableProvider) Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error) {
    result, err := rp.executor.Execute(ctx, func(ctx context.Context) (interface{}, error) {
        return rp.Provider.Complete(ctx, request)
    })
    
    if err != nil {
        return nil, err
    }
    
    return result.(*CompletionResponse), nil
}
```

### 2. Circuit Breaker Pattern

```go
// CircuitBreaker prevents cascading failures
type CircuitBreaker struct {
    name           string
    config         CircuitBreakerConfig
    state          CircuitState
    failures       int64
    successes      int64
    lastFailTime   time.Time
    lastStateChange time.Time
    mu             sync.RWMutex
    onStateChange  func(from, to CircuitState)
}

type CircuitBreakerConfig struct {
    FailureThreshold   int           `yaml:"failure_threshold" json:"failure_threshold"`
    SuccessThreshold   int           `yaml:"success_threshold" json:"success_threshold"`
    Timeout            time.Duration `yaml:"timeout" json:"timeout"`
    MaxRequests        int           `yaml:"max_requests" json:"max_requests"`
    ResetTimeout       time.Duration `yaml:"reset_timeout" json:"reset_timeout"`
}

type CircuitState string

const (
    StateClosed     CircuitState = "closed"
    StateOpen       CircuitState = "open"
    StateHalfOpen   CircuitState = "half_open"
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
    return &CircuitBreaker{
        name:            name,
        config:          config,
        state:           StateClosed,
        lastStateChange: time.Now(),
    }
}

// Execute runs an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, operation Operation) (interface{}, error) {
    // Check if circuit breaker allows execution
    if err := cb.allowExecution(); err != nil {
        return nil, err
    }
    
    // Execute operation
    start := time.Now()
    result, err := operation(ctx)
    duration := time.Since(start)
    
    // Record result
    if err != nil {
        cb.recordFailure(err)
        return nil, err
    }
    
    cb.recordSuccess(duration)
    return result, nil
}

// allowExecution checks if the operation should be executed
func (cb *CircuitBreaker) allowExecution() error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    switch cb.state {
    case StateClosed:
        return nil
        
    case StateOpen:
        // Check if timeout has elapsed
        if time.Since(cb.lastStateChange) >= cb.config.ResetTimeout {
            cb.setState(StateHalfOpen)
            return nil
        }
        return NewCircuitBreakerError(cb.name, "circuit breaker is open")
        
    case StateHalfOpen:
        // Allow limited number of requests
        if cb.successes < int64(cb.config.MaxRequests) {
            return nil
        }
        return NewCircuitBreakerError(cb.name, "circuit breaker is half-open with max requests reached")
        
    default:
        return NewCircuitBreakerError(cb.name, "unknown circuit breaker state")
    }
}

// recordFailure records a failed operation
func (cb *CircuitBreaker) recordFailure(err error) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.failures++
    cb.lastFailTime = time.Now()
    
    switch cb.state {
    case StateClosed:
        if cb.failures >= int64(cb.config.FailureThreshold) {
            cb.setState(StateOpen)
        }
        
    case StateHalfOpen:
        cb.setState(StateOpen)
        
    case StateOpen:
        // Already open, no state change needed
    }
}

// recordSuccess records a successful operation
func (cb *CircuitBreaker) recordSuccess(duration time.Duration) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.successes++
    
    switch cb.state {
    case StateHalfOpen:
        if cb.successes >= int64(cb.config.SuccessThreshold) {
            cb.setState(StateClosed)
            cb.failures = 0
            cb.successes = 0
        }
        
    case StateClosed:
        // Reset failure count on success
        if cb.failures > 0 {
            cb.failures = 0
        }
        
    case StateOpen:
        // Should not reach here, but handle gracefully
        cb.setState(StateClosed)
        cb.failures = 0
        cb.successes = 0
    }
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState CircuitState) {
    oldState := cb.state
    cb.state = newState
    cb.lastStateChange = time.Now()
    
    if cb.onStateChange != nil {
        go cb.onStateChange(oldState, newState)
    }
}

// CircuitBreakerError represents circuit breaker errors
type CircuitBreakerError struct {
    *BaseError
    CircuitName string `json:"circuit_name"`
}

func NewCircuitBreakerError(circuitName, message string) *CircuitBreakerError {
    return &CircuitBreakerError{
        BaseError: &BaseError{
            Type:      ErrorTypeDependency,
            Code:      "CIRCUIT_BREAKER_OPEN",
            Message:   message,
            Timestamp: time.Now(),
            Retryable: true,
            Severity:  SeverityHigh,
        },
        CircuitName: circuitName,
    }
}

// ProtectedProvider wraps a provider with circuit breaker protection
type ProtectedProvider struct {
    Provider
    circuitBreaker *CircuitBreaker
}

func (pp *ProtectedProvider) Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error) {
    result, err := pp.circuitBreaker.Execute(ctx, func(ctx context.Context) (interface{}, error) {
        return pp.Provider.Complete(ctx, request)
    })
    
    if err != nil {
        return nil, err
    }
    
    return result.(*CompletionResponse), nil
}
```

### 3. Graceful Degradation

```go
// DegradationManager handles graceful service degradation
type DegradationManager struct {
    levels    []DegradationLevel
    current   int
    triggers  []DegradationTrigger
    actions   map[string]DegradationAction
    metrics   *DegradationMetrics
    mu        sync.RWMutex
}

type DegradationLevel struct {
    Name        string   `yaml:"name" json:"name"`
    Severity    int      `yaml:"severity" json:"severity"`
    Description string   `yaml:"description" json:"description"`
    Actions     []string `yaml:"actions" json:"actions"`
    AutoRevert  bool     `yaml:"auto_revert" json:"auto_revert"`
}

type DegradationTrigger struct {
    Name        string        `yaml:"name" json:"name"`
    Condition   string        `yaml:"condition" json:"condition"`
    Threshold   float64       `yaml:"threshold" json:"threshold"`
    Duration    time.Duration `yaml:"duration" json:"duration"`
    Level       int           `yaml:"level" json:"level"`
}

type DegradationAction interface {
    Name() string
    Execute(ctx context.Context, level DegradationLevel) error
    Revert(ctx context.Context) error
    IsActive() bool
}

type DegradationMetrics struct {
    CurrentLevel     int
    ActivatedActions []string
    TriggerHistory   []TriggerEvent
    LastChange       time.Time
}

type TriggerEvent struct {
    Trigger   string    `json:"trigger"`
    Level     int       `json:"level"`
    Timestamp time.Time `json:"timestamp"`
    Action    string    `json:"action"` // activated, deactivated
}

// NewDegradationManager creates a degradation manager
func NewDegradationManager(levels []DegradationLevel) *DegradationManager {
    return &DegradationManager{
        levels:  levels,
        current: 0, // Normal operation
        actions: make(map[string]DegradationAction),
        metrics: &DegradationMetrics{
            TriggerHistory: make([]TriggerEvent, 0),
        },
    }
}

// RegisterAction registers a degradation action
func (dm *DegradationManager) RegisterAction(action DegradationAction) {
    dm.mu.Lock()
    defer dm.mu.Unlock()
    dm.actions[action.Name()] = action
}

// TriggerDegradation activates a degradation level
func (dm *DegradationManager) TriggerDegradation(ctx context.Context, level int, trigger string) error {
    dm.mu.Lock()
    defer dm.mu.Unlock()
    
    if level <= dm.current {
        return nil // Already at this level or higher
    }
    
    if level >= len(dm.levels) {
        return fmt.Errorf("invalid degradation level: %d", level)
    }
    
    degradationLevel := dm.levels[level]
    
    // Execute degradation actions
    for _, actionName := range degradationLevel.Actions {
        action, exists := dm.actions[actionName]
        if !exists {
            continue
        }
        
        if err := action.Execute(ctx, degradationLevel); err != nil {
            return fmt.Errorf("failed to execute degradation action %s: %w", actionName, err)
        }
    }
    
    // Update state
    dm.current = level
    dm.metrics.CurrentLevel = level
    dm.metrics.LastChange = time.Now()
    
    // Record trigger event
    dm.metrics.TriggerHistory = append(dm.metrics.TriggerHistory, TriggerEvent{
        Trigger:   trigger,
        Level:     level,
        Timestamp: time.Now(),
        Action:    "activated",
    })
    
    return nil
}

// RevertDegradation returns to normal operation
func (dm *DegradationManager) RevertDegradation(ctx context.Context) error {
    dm.mu.Lock()
    defer dm.mu.Unlock()
    
    if dm.current == 0 {
        return nil // Already at normal level
    }
    
    // Revert actions in reverse order
    for level := dm.current; level > 0; level-- {
        degradationLevel := dm.levels[level]
        
        for i := len(degradationLevel.Actions) - 1; i >= 0; i-- {
            actionName := degradationLevel.Actions[i]
            action, exists := dm.actions[actionName]
            if !exists {
                continue
            }
            
            if err := action.Revert(ctx); err != nil {
                return fmt.Errorf("failed to revert degradation action %s: %w", actionName, err)
            }
        }
    }
    
    // Update state
    dm.current = 0
    dm.metrics.CurrentLevel = 0
    dm.metrics.LastChange = time.Now()
    dm.metrics.ActivatedActions = nil
    
    return nil
}

// Built-in degradation actions
type RateLimitAction struct {
    name     string
    limiter  RateLimiter
    original RateLimit
    degraded RateLimit
    active   bool
}

func (rla *RateLimitAction) Name() string {
    return rla.name
}

func (rla *RateLimitAction) Execute(ctx context.Context, level DegradationLevel) error {
    if err := rla.limiter.SetLimit(rla.degraded); err != nil {
        return err
    }
    rla.active = true
    return nil
}

func (rla *RateLimitAction) Revert(ctx context.Context) error {
    if err := rla.limiter.SetLimit(rla.original); err != nil {
        return err
    }
    rla.active = false
    return nil
}

func (rla *RateLimitAction) IsActive() bool {
    return rla.active
}

// CacheBypassAction disables caching during degradation
type CacheBypassAction struct {
    name   string
    cache  Cache
    active bool
}

func (cba *CacheBypassAction) Name() string {
    return cba.name
}

func (cba *CacheBypassAction) Execute(ctx context.Context, level DegradationLevel) error {
    cba.cache.Disable()
    cba.active = true
    return nil
}

func (cba *CacheBypassAction) Revert(ctx context.Context) error {
    cba.cache.Enable()
    cba.active = false
    return nil
}

func (cba *CacheBypassAction) IsActive() bool {
    return cba.active
}

// FallbackProviderAction switches to a fallback provider
type FallbackProviderAction struct {
    name            string
    primary         Provider
    fallback        Provider
    providerManager *ProviderManager
    active          bool
}

func (fpa *FallbackProviderAction) Name() string {
    return fpa.name
}

func (fpa *FallbackProviderAction) Execute(ctx context.Context, level DegradationLevel) error {
    fpa.providerManager.SetActiveProvider(fpa.fallback)
    fpa.active = true
    return nil
}

func (fpa *FallbackProviderAction) Revert(ctx context.Context) error {
    fpa.providerManager.SetActiveProvider(fpa.primary)
    fpa.active = false
    return nil
}

func (fpa *FallbackProviderAction) IsActive() bool {
    return fpa.active
}
```

## Error Monitoring and Alerting

### 1. Error Aggregation and Analysis

```go
// ErrorAggregator collects and analyzes errors
type ErrorAggregator struct {
    errors    map[string]*ErrorStats
    window    time.Duration
    thresholds map[string]AlertThreshold
    alerter   Alerter
    mu        sync.RWMutex
}

type ErrorStats struct {
    Count       int64                 `json:"count"`
    FirstSeen   time.Time             `json:"first_seen"`
    LastSeen    time.Time             `json:"last_seen"`
    Occurrences []time.Time           `json:"occurrences"`
    Samples     []*BaseError          `json:"samples"`
    Rate        float64               `json:"rate"`
    Trend       string                `json:"trend"`
    Impact      string                `json:"impact"`
}

type AlertThreshold struct {
    ErrorType     ErrorType     `yaml:"error_type" json:"error_type"`
    Count         int64         `yaml:"count" json:"count"`
    Rate          float64       `yaml:"rate" json:"rate"`
    Window        time.Duration `yaml:"window" json:"window"`
    Severity      AlertSeverity `yaml:"severity" json:"severity"`
}

type AlertSeverity string

const (
    AlertSeverityInfo     AlertSeverity = "info"
    AlertSeverityWarning  AlertSeverity = "warning"
    AlertSeverityError    AlertSeverity = "error"
    AlertSeverityCritical AlertSeverity = "critical"
)

// Record adds an error to the aggregator
func (ea *ErrorAggregator) Record(err error) {
    if err == nil {
        return
    }
    
    ea.mu.Lock()
    defer ea.mu.Unlock()
    
    key := ea.getErrorKey(err)
    now := time.Now()
    
    stats, exists := ea.errors[key]
    if !exists {
        stats = &ErrorStats{
            FirstSeen:   now,
            Occurrences: make([]time.Time, 0),
            Samples:     make([]*BaseError, 0, 5),
        }
        ea.errors[key] = stats
    }
    
    stats.Count++
    stats.LastSeen = now
    stats.Occurrences = append(stats.Occurrences, now)
    
    // Keep only recent occurrences within window
    cutoff := now.Add(-ea.window)
    filtered := make([]time.Time, 0, len(stats.Occurrences))
    for _, occurrence := range stats.Occurrences {
        if occurrence.After(cutoff) {
            filtered = append(filtered, occurrence)
        }
    }
    stats.Occurrences = filtered
    
    // Add sample if we have room
    if len(stats.Samples) < 5 {
        if baseErr, ok := err.(*BaseError); ok {
            stats.Samples = append(stats.Samples, baseErr)
        }
    }
    
    // Calculate current rate
    stats.Rate = float64(len(stats.Occurrences)) / ea.window.Seconds()
    
    // Check alert thresholds
    ea.checkAlertThresholds(key, stats)
}

// getErrorKey generates a unique key for error aggregation
func (ea *ErrorAggregator) getErrorKey(err error) string {
    if baseErr, ok := err.(*BaseError); ok {
        return fmt.Sprintf("%s:%s", baseErr.Type, baseErr.Code)
    }
    
    // Fallback to error type
    return fmt.Sprintf("unknown:%T", err)
}

// checkAlertThresholds checks if any alert thresholds are exceeded
func (ea *ErrorAggregator) checkAlertThresholds(key string, stats *ErrorStats) {
    for _, threshold := range ea.thresholds {
        if threshold.Count > 0 && stats.Count >= threshold.Count {
            ea.sendAlert(key, stats, threshold, "count")
        }
        
        if threshold.Rate > 0 && stats.Rate >= threshold.Rate {
            ea.sendAlert(key, stats, threshold, "rate")
        }
    }
}

// sendAlert sends an alert notification
func (ea *ErrorAggregator) sendAlert(key string, stats *ErrorStats, threshold AlertThreshold, triggerType string) {
    alert := Alert{
        ID:          generateAlertID(),
        Type:        "error_threshold",
        Severity:    threshold.Severity,
        Title:       fmt.Sprintf("Error threshold exceeded: %s", key),
        Description: fmt.Sprintf("Error %s %s threshold: %v", key, triggerType, getThresholdValue(threshold, triggerType)),
        Timestamp:   time.Now(),
        Data: map[string]interface{}{
            "error_key":    key,
            "stats":        stats,
            "threshold":    threshold,
            "trigger_type": triggerType,
        },
    }
    
    if ea.alerter != nil {
        ea.alerter.Send(alert)
    }
}

// GetStats returns error statistics
func (ea *ErrorAggregator) GetStats() map[string]*ErrorStats {
    ea.mu.RLock()
    defer ea.mu.RUnlock()
    
    // Create a copy to avoid race conditions
    stats := make(map[string]*ErrorStats)
    for key, stat := range ea.errors {
        statCopy := *stat
        stats[key] = &statCopy
    }
    
    return stats
}

// ErrorReporter generates error reports
type ErrorReporter struct {
    aggregator *ErrorAggregator
    templates  map[string]*ReportTemplate
}

type ReportTemplate struct {
    Name        string                 `yaml:"name" json:"name"`
    Format      string                 `yaml:"format" json:"format"`
    Sections    []ReportSection        `yaml:"sections" json:"sections"`
    Recipients  []string               `yaml:"recipients" json:"recipients"`
    Schedule    string                 `yaml:"schedule" json:"schedule"`
    Filters     map[string]interface{} `yaml:"filters" json:"filters"`
}

type ReportSection struct {
    Title   string   `yaml:"title" json:"title"`
    Type    string   `yaml:"type" json:"type"`
    Queries []string `yaml:"queries" json:"queries"`
    Limit   int      `yaml:"limit" json:"limit"`
}

// GenerateReport creates an error report
func (er *ErrorReporter) GenerateReport(templateName string, startTime, endTime time.Time) (*ErrorReport, error) {
    template, exists := er.templates[templateName]
    if !exists {
        return nil, fmt.Errorf("report template %s not found", templateName)
    }
    
    stats := er.aggregator.GetStats()
    
    report := &ErrorReport{
        Title:     template.Name,
        StartTime: startTime,
        EndTime:   endTime,
        Generated: time.Now(),
        Sections:  make([]ReportSectionData, len(template.Sections)),
    }
    
    for i, section := range template.Sections {
        sectionData, err := er.generateSection(section, stats, startTime, endTime)
        if err != nil {
            return nil, fmt.Errorf("failed to generate section %s: %w", section.Title, err)
        }
        report.Sections[i] = *sectionData
    }
    
    return report, nil
}

type ErrorReport struct {
    Title     string              `json:"title"`
    StartTime time.Time           `json:"start_time"`
    EndTime   time.Time           `json:"end_time"`
    Generated time.Time           `json:"generated"`
    Sections  []ReportSectionData `json:"sections"`
    Summary   ReportSummary       `json:"summary"`
}

type ReportSectionData struct {
    Title string      `json:"title"`
    Data  interface{} `json:"data"`
    Chart interface{} `json:"chart,omitempty"`
}

type ReportSummary struct {
    TotalErrors    int64   `json:"total_errors"`
    UniqueErrors   int     `json:"unique_errors"`
    ErrorRate      float64 `json:"error_rate"`
    TopErrors      []string `json:"top_errors"`
    CriticalErrors int     `json:"critical_errors"`
    Trend          string  `json:"trend"`
}
```

This comprehensive error handling guide provides the foundation for building resilient LLM applications with robust error management, recovery strategies, and monitoring capabilities in Go-LLMs.