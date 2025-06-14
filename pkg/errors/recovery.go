package errors

// ABOUTME: Implementation of error recovery strategies with retry logic
// ABOUTME: Provides built-in strategies and framework for custom recoveries

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// BaseRecoveryStrategy provides common functionality for recovery strategies
type BaseRecoveryStrategy struct {
	name        string
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
}

// Name returns the strategy name
func (s *BaseRecoveryStrategy) Name() string {
	return s.name
}

// MaxAttempts returns the maximum number of attempts
func (s *BaseRecoveryStrategy) MaxAttempts() int {
	return s.maxAttempts
}

// ExponentialBackoffStrategy implements exponential backoff with jitter
type ExponentialBackoffStrategy struct {
	BaseRecoveryStrategy
	factor float64
	jitter bool
}

// NewExponentialBackoffStrategy creates a new exponential backoff strategy
func NewExponentialBackoffStrategy(maxAttempts int, baseDelay, maxDelay time.Duration) *ExponentialBackoffStrategy {
	return &ExponentialBackoffStrategy{
		BaseRecoveryStrategy: BaseRecoveryStrategy{
			name:        "exponential_backoff",
			maxAttempts: maxAttempts,
			baseDelay:   baseDelay,
			maxDelay:    maxDelay,
		},
		factor: 2.0,
		jitter: true,
	}
}

// CanRecover checks if the error is recoverable
func (s *ExponentialBackoffStrategy) CanRecover(err error) bool {
	// Check if error is marked as retryable
	return IsRetryableError(err)
}

// Recover attempts recovery (returns error as retry is handled externally)
func (s *ExponentialBackoffStrategy) Recover(err error, context map[string]interface{}) error {
	// This strategy doesn't perform recovery, just determines if retry is possible
	return fmt.Errorf("retry required: %w", err)
}

// BackoffDuration calculates the backoff duration for an attempt
func (s *ExponentialBackoffStrategy) BackoffDuration(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	// Calculate exponential delay
	delay := float64(s.baseDelay) * math.Pow(s.factor, float64(attempt-1))

	// Cap at max delay
	if delay > float64(s.maxDelay) {
		delay = float64(s.maxDelay)
	}

	// Add jitter if enabled
	if s.jitter {
		// Add up to 20% jitter
		jitterRange := delay * 0.2
		jitter := (rand.Float64() - 0.5) * jitterRange
		delay += jitter
	}

	return time.Duration(delay)
}

// LinearBackoffStrategy implements linear backoff
type LinearBackoffStrategy struct {
	BaseRecoveryStrategy
	increment time.Duration
}

// NewLinearBackoffStrategy creates a new linear backoff strategy
func NewLinearBackoffStrategy(maxAttempts int, increment time.Duration) *LinearBackoffStrategy {
	return &LinearBackoffStrategy{
		BaseRecoveryStrategy: BaseRecoveryStrategy{
			name:        "linear_backoff",
			maxAttempts: maxAttempts,
			baseDelay:   increment,
			maxDelay:    increment * time.Duration(maxAttempts),
		},
		increment: increment,
	}
}

// CanRecover checks if the error is recoverable
func (s *LinearBackoffStrategy) CanRecover(err error) bool {
	return IsRetryableError(err)
}

// Recover attempts recovery
func (s *LinearBackoffStrategy) Recover(err error, context map[string]interface{}) error {
	return fmt.Errorf("retry required: %w", err)
}

// BackoffDuration calculates the backoff duration
func (s *LinearBackoffStrategy) BackoffDuration(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	return s.increment * time.Duration(attempt)
}

// NoRetryStrategy never retries
type NoRetryStrategy struct {
	BaseRecoveryStrategy
}

// NewNoRetryStrategy creates a strategy that never retries
func NewNoRetryStrategy() *NoRetryStrategy {
	return &NoRetryStrategy{
		BaseRecoveryStrategy: BaseRecoveryStrategy{
			name:        "no_retry",
			maxAttempts: 0,
		},
	}
}

// CanRecover always returns false
func (s *NoRetryStrategy) CanRecover(err error) bool {
	return false
}

// Recover returns the error unchanged
func (s *NoRetryStrategy) Recover(err error, context map[string]interface{}) error {
	return err
}

// BackoffDuration always returns 0
func (s *NoRetryStrategy) BackoffDuration(attempt int) time.Duration {
	return 0
}

// FallbackStrategy attempts to use a fallback value or function
type FallbackStrategy struct {
	BaseRecoveryStrategy
	fallbackFunc func(error, map[string]interface{}) error
}

// NewFallbackStrategy creates a strategy that uses a fallback
func NewFallbackStrategy(fallbackFunc func(error, map[string]interface{}) error) *FallbackStrategy {
	return &FallbackStrategy{
		BaseRecoveryStrategy: BaseRecoveryStrategy{
			name:        "fallback",
			maxAttempts: 1,
		},
		fallbackFunc: fallbackFunc,
	}
}

// CanRecover checks if fallback is available
func (s *FallbackStrategy) CanRecover(err error) bool {
	return s.fallbackFunc != nil
}

// Recover attempts to use the fallback
func (s *FallbackStrategy) Recover(err error, context map[string]interface{}) error {
	if s.fallbackFunc == nil {
		return err
	}
	return s.fallbackFunc(err, context)
}

// BackoffDuration returns 0 (immediate fallback)
func (s *FallbackStrategy) BackoffDuration(attempt int) time.Duration {
	return 0
}

// CircuitBreakerStrategy implements circuit breaker pattern
type CircuitBreakerStrategy struct {
	BaseRecoveryStrategy
	failureThreshold int
	resetTimeout     time.Duration
	state            string
	failures         int
	lastFailureTime  time.Time
}

// NewCircuitBreakerStrategy creates a circuit breaker strategy
func NewCircuitBreakerStrategy(failureThreshold int, resetTimeout time.Duration) *CircuitBreakerStrategy {
	return &CircuitBreakerStrategy{
		BaseRecoveryStrategy: BaseRecoveryStrategy{
			name:        "circuit_breaker",
			maxAttempts: 1,
		},
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:            "closed", // closed, open, half-open
	}
}

// CanRecover checks circuit breaker state
func (s *CircuitBreakerStrategy) CanRecover(err error) bool {
	now := time.Now()

	switch s.state {
	case "open":
		// Check if we should transition to half-open
		if now.Sub(s.lastFailureTime) > s.resetTimeout {
			s.state = "half-open"
			return true
		}
		return false
	case "half-open", "closed":
		return true
	default:
		return false
	}
}

// Recover updates circuit breaker state
func (s *CircuitBreakerStrategy) Recover(err error, context map[string]interface{}) error {
	if err != nil {
		s.failures++
		s.lastFailureTime = time.Now()

		if s.failures >= s.failureThreshold {
			s.state = "open"
		}
		return fmt.Errorf("circuit breaker: %w", err)
	}

	// Success - reset state
	s.failures = 0
	s.state = "closed"
	return nil
}

// BackoffDuration returns 0 or reset timeout
func (s *CircuitBreakerStrategy) BackoffDuration(attempt int) time.Duration {
	if s.state == "open" {
		remaining := s.resetTimeout - time.Since(s.lastFailureTime)
		if remaining > 0 {
			return remaining
		}
	}
	return 0
}

// CompositeStrategy combines multiple strategies
type CompositeStrategy struct {
	strategies []RecoveryStrategy
	current    int
}

// NewCompositeStrategy creates a strategy that tries multiple strategies in order
func NewCompositeStrategy(strategies ...RecoveryStrategy) *CompositeStrategy {
	return &CompositeStrategy{
		strategies: strategies,
		current:    0,
	}
}

// Name returns the composite strategy name
func (s *CompositeStrategy) Name() string {
	return "composite"
}

// CanRecover checks if any strategy can recover
func (s *CompositeStrategy) CanRecover(err error) bool {
	for _, strategy := range s.strategies {
		if strategy.CanRecover(err) {
			return true
		}
	}
	return false
}

// Recover attempts recovery with current strategy
func (s *CompositeStrategy) Recover(err error, context map[string]interface{}) error {
	if s.current >= len(s.strategies) {
		return fmt.Errorf("all strategies exhausted: %w", err)
	}

	strategy := s.strategies[s.current]
	if !strategy.CanRecover(err) {
		s.current++
		return s.Recover(err, context)
	}

	result := strategy.Recover(err, context)
	if result != nil {
		// Try next strategy if current one failed
		if s.current < len(s.strategies)-1 {
			s.current++
		}
	}
	return result
}

// MaxAttempts returns sum of all strategy attempts
func (s *CompositeStrategy) MaxAttempts() int {
	total := 0
	for _, strategy := range s.strategies {
		total += strategy.MaxAttempts()
	}
	return total
}

// BackoffDuration delegates to current strategy
func (s *CompositeStrategy) BackoffDuration(attempt int) time.Duration {
	if s.current >= len(s.strategies) {
		return 0
	}
	return s.strategies[s.current].BackoffDuration(attempt)
}

// RecoveryRegistry manages recovery strategies
type recoveryRegistry struct {
	strategies map[string]RecoveryStrategy
}

var defaultRegistry = &recoveryRegistry{
	strategies: make(map[string]RecoveryStrategy),
}

// RegisterRecoveryStrategy registers a recovery strategy
func RegisterRecoveryStrategy(name string, strategy RecoveryStrategy) {
	defaultRegistry.strategies[name] = strategy
}

// GetRecoveryStrategy retrieves a recovery strategy by name
func GetRecoveryStrategy(name string) (RecoveryStrategy, bool) {
	strategy, ok := defaultRegistry.strategies[name]
	return strategy, ok
}

// DefaultRecoveryStrategies returns commonly used strategies
func DefaultRecoveryStrategies() map[string]RecoveryStrategy {
	return map[string]RecoveryStrategy{
		"exponential": NewExponentialBackoffStrategy(5, 100*time.Millisecond, 30*time.Second),
		"linear":      NewLinearBackoffStrategy(3, 1*time.Second),
		"no_retry":    NewNoRetryStrategy(),
		"circuit":     NewCircuitBreakerStrategy(5, 60*time.Second),
	}
}

// init registers default strategies
func init() {
	for name, strategy := range DefaultRecoveryStrategies() {
		RegisterRecoveryStrategy(name, strategy)
	}
}
