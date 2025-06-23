package errors

// ABOUTME: Implementation of error recovery strategies with retry logic
// ABOUTME: Provides built-in strategies and framework for custom recoveries

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// BaseRecoveryStrategy provides common functionality for recovery strategies.
// It serves as a foundation for implementing specific recovery patterns
// with configurable retry attempts and delay parameters.
type BaseRecoveryStrategy struct {
	name        string
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
}

// Name returns the strategy name.
// Implements RecoveryStrategy.Name.
//
// Returns the strategy identifier.
func (s *BaseRecoveryStrategy) Name() string {
	return s.name
}

// MaxAttempts returns the maximum number of attempts.
// Implements RecoveryStrategy.MaxAttempts.
//
// Returns the configured maximum retry attempts.
func (s *BaseRecoveryStrategy) MaxAttempts() int {
	return s.maxAttempts
}

// ExponentialBackoffStrategy implements exponential backoff with jitter.
// It increases delay exponentially between retries with optional
// randomization to prevent thundering herd problems.
type ExponentialBackoffStrategy struct {
	BaseRecoveryStrategy
	factor float64
	jitter bool
}

// NewExponentialBackoffStrategy creates a new exponential backoff strategy.
// The delay grows exponentially (delay * 2^attempt) up to maxDelay.
//
// Parameters:
//   - maxAttempts: Maximum number of retry attempts
//   - baseDelay: Initial delay duration
//   - maxDelay: Maximum delay duration cap
//
// Returns a configured ExponentialBackoffStrategy.
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

// CanRecover checks if the error is recoverable.
// Uses IsRetryableError to determine if retry is appropriate.
// Implements RecoveryStrategy.CanRecover.
//
// Parameters:
//   - err: The error to check
//
// Returns true if the error is retryable.
func (s *ExponentialBackoffStrategy) CanRecover(err error) bool {
	// Check if error is marked as retryable
	return IsRetryableError(err)
}

// Recover attempts recovery (returns error as retry is handled externally).
// This strategy doesn't perform actual recovery but signals retry requirement.
// Implements RecoveryStrategy.Recover.
//
// Parameters:
//   - err: The error to recover from
//   - context: Additional context information
//
// Returns a wrapped error indicating retry is needed.
func (s *ExponentialBackoffStrategy) Recover(err error, context map[string]interface{}) error {
	// This strategy doesn't perform recovery, just determines if retry is possible
	return fmt.Errorf("retry required: %w", err)
}

// BackoffDuration calculates the backoff duration for an attempt.
// Uses exponential growth with optional jitter (±20% randomization).
// Implements RecoveryStrategy.BackoffDuration.
//
// Parameters:
//   - attempt: The attempt number (1-based)
//
// Returns the calculated backoff duration.
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

// LinearBackoffStrategy implements linear backoff.
// It increases delay linearly between retries, providing
// predictable and uniform retry intervals.
type LinearBackoffStrategy struct {
	BaseRecoveryStrategy
	increment time.Duration
}

// NewLinearBackoffStrategy creates a new linear backoff strategy.
// The delay grows linearly (increment * attempt).
//
// Parameters:
//   - maxAttempts: Maximum number of retry attempts
//   - increment: Delay increment per attempt
//
// Returns a configured LinearBackoffStrategy.
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

// CanRecover checks if the error is recoverable.
// Implements RecoveryStrategy.CanRecover.
//
// Parameters:
//   - err: The error to check
//
// Returns true if the error is retryable.
func (s *LinearBackoffStrategy) CanRecover(err error) bool {
	return IsRetryableError(err)
}

// Recover attempts recovery.
// Implements RecoveryStrategy.Recover.
//
// Parameters:
//   - err: The error to recover from
//   - context: Additional context information
//
// Returns a wrapped error indicating retry is needed.
func (s *LinearBackoffStrategy) Recover(err error, context map[string]interface{}) error {
	return fmt.Errorf("retry required: %w", err)
}

// BackoffDuration calculates the backoff duration.
// Uses linear growth (increment * attempt).
// Implements RecoveryStrategy.BackoffDuration.
//
// Parameters:
//   - attempt: The attempt number (1-based)
//
// Returns the calculated backoff duration.
func (s *LinearBackoffStrategy) BackoffDuration(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}
	return s.increment * time.Duration(attempt)
}

// NoRetryStrategy never retries.
// This strategy immediately fails without any retry attempts,
// useful for non-recoverable errors or when retries are disabled.
type NoRetryStrategy struct {
	BaseRecoveryStrategy
}

// NewNoRetryStrategy creates a strategy that never retries.
// Useful for fatal errors or when retry logic should be disabled.
//
// Returns a configured NoRetryStrategy.
func NewNoRetryStrategy() *NoRetryStrategy {
	return &NoRetryStrategy{
		BaseRecoveryStrategy: BaseRecoveryStrategy{
			name:        "no_retry",
			maxAttempts: 0,
		},
	}
}

// CanRecover always returns false.
// Implements RecoveryStrategy.CanRecover.
//
// Parameters:
//   - err: The error to check (ignored)
//
// Always returns false.
func (s *NoRetryStrategy) CanRecover(err error) bool {
	return false
}

// Recover returns the error unchanged.
// Implements RecoveryStrategy.Recover.
//
// Parameters:
//   - err: The error to return
//   - context: Additional context (ignored)
//
// Returns the original error.
func (s *NoRetryStrategy) Recover(err error, context map[string]interface{}) error {
	return err
}

// BackoffDuration always returns 0.
// Implements RecoveryStrategy.BackoffDuration.
//
// Parameters:
//   - attempt: The attempt number (ignored)
//
// Always returns zero duration.
func (s *NoRetryStrategy) BackoffDuration(attempt int) time.Duration {
	return 0
}

// FallbackStrategy attempts to use a fallback value or function.
// It provides an alternative path when the primary operation fails,
// implementing the fallback pattern for graceful degradation.
type FallbackStrategy struct {
	BaseRecoveryStrategy
	fallbackFunc func(error, map[string]interface{}) error
}

// NewFallbackStrategy creates a strategy that uses a fallback.
// The fallback function is called when the primary operation fails.
//
// Parameters:
//   - fallbackFunc: Function to call for fallback behavior
//
// Returns a configured FallbackStrategy.
func NewFallbackStrategy(fallbackFunc func(error, map[string]interface{}) error) *FallbackStrategy {
	return &FallbackStrategy{
		BaseRecoveryStrategy: BaseRecoveryStrategy{
			name:        "fallback",
			maxAttempts: 1,
		},
		fallbackFunc: fallbackFunc,
	}
}

// CanRecover checks if fallback is available.
// Implements RecoveryStrategy.CanRecover.
//
// Parameters:
//   - err: The error to check (ignored)
//
// Returns true if a fallback function is configured.
func (s *FallbackStrategy) CanRecover(err error) bool {
	return s.fallbackFunc != nil
}

// Recover attempts to use the fallback.
// Calls the configured fallback function or returns the original error.
// Implements RecoveryStrategy.Recover.
//
// Parameters:
//   - err: The error that triggered fallback
//   - context: Additional context for the fallback
//
// Returns the result of the fallback or the original error.
func (s *FallbackStrategy) Recover(err error, context map[string]interface{}) error {
	if s.fallbackFunc == nil {
		return err
	}
	return s.fallbackFunc(err, context)
}

// BackoffDuration returns 0 (immediate fallback).
// Fallback strategies execute immediately without delay.
// Implements RecoveryStrategy.BackoffDuration.
//
// Parameters:
//   - attempt: The attempt number (ignored)
//
// Always returns zero duration.
func (s *FallbackStrategy) BackoffDuration(attempt int) time.Duration {
	return 0
}

// CircuitBreakerStrategy implements circuit breaker pattern.
// It prevents cascading failures by temporarily blocking requests
// after a threshold of failures, allowing the system to recover.
type CircuitBreakerStrategy struct {
	BaseRecoveryStrategy
	failureThreshold int
	resetTimeout     time.Duration
	state            string
	failures         int
	lastFailureTime  time.Time
}

// NewCircuitBreakerStrategy creates a circuit breaker strategy.
// The circuit opens after failureThreshold failures and resets after resetTimeout.
//
// Parameters:
//   - failureThreshold: Number of failures before opening circuit
//   - resetTimeout: Duration before attempting to close circuit
//
// Returns a configured CircuitBreakerStrategy.
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

// CanRecover checks circuit breaker state.
// Returns true if circuit is closed or half-open.
// Implements RecoveryStrategy.CanRecover.
//
// Parameters:
//   - err: The error to check (used for state transitions)
//
// Returns true if recovery attempt is allowed.
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

// Recover updates circuit breaker state.
// Tracks failures and manages state transitions between closed, open, and half-open.
// Implements RecoveryStrategy.Recover.
//
// Parameters:
//   - err: The error (nil indicates success)
//   - context: Additional context information
//
// Returns wrapped error or nil on success.
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

// BackoffDuration returns 0 or reset timeout.
// Returns remaining time before circuit can transition to half-open.
// Implements RecoveryStrategy.BackoffDuration.
//
// Parameters:
//   - attempt: The attempt number (ignored)
//
// Returns duration to wait or zero if ready.
func (s *CircuitBreakerStrategy) BackoffDuration(attempt int) time.Duration {
	if s.state == "open" {
		remaining := s.resetTimeout - time.Since(s.lastFailureTime)
		if remaining > 0 {
			return remaining
		}
	}
	return 0
}

// CompositeStrategy combines multiple strategies.
// It tries strategies in order until one succeeds or all fail,
// enabling complex recovery patterns through composition.
type CompositeStrategy struct {
	strategies []RecoveryStrategy
	current    int
}

// NewCompositeStrategy creates a strategy that tries multiple strategies in order.
// Strategies are attempted sequentially until one succeeds.
//
// Parameters:
//   - strategies: Variable number of strategies to compose
//
// Returns a configured CompositeStrategy.
func NewCompositeStrategy(strategies ...RecoveryStrategy) *CompositeStrategy {
	return &CompositeStrategy{
		strategies: strategies,
		current:    0,
	}
}

// Name returns the composite strategy name.
// Implements RecoveryStrategy.Name.
//
// Returns "composite".
func (s *CompositeStrategy) Name() string {
	return "composite"
}

// CanRecover checks if any strategy can recover.
// Returns true if at least one composed strategy can recover.
// Implements RecoveryStrategy.CanRecover.
//
// Parameters:
//   - err: The error to check
//
// Returns true if any strategy can handle the error.
func (s *CompositeStrategy) CanRecover(err error) bool {
	for _, strategy := range s.strategies {
		if strategy.CanRecover(err) {
			return true
		}
	}
	return false
}

// Recover attempts recovery with current strategy.
// Tries strategies in order, advancing to the next on failure.
// Implements RecoveryStrategy.Recover.
//
// Parameters:
//   - err: The error to recover from
//   - context: Additional context information
//
// Returns error from the current strategy or exhaustion error.
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

// MaxAttempts returns sum of all strategy attempts.
// Implements RecoveryStrategy.MaxAttempts.
//
// Returns total attempts across all composed strategies.
func (s *CompositeStrategy) MaxAttempts() int {
	total := 0
	for _, strategy := range s.strategies {
		total += strategy.MaxAttempts()
	}
	return total
}

// BackoffDuration delegates to current strategy.
// Implements RecoveryStrategy.BackoffDuration.
//
// Parameters:
//   - attempt: The attempt number
//
// Returns duration from the current strategy or zero.
func (s *CompositeStrategy) BackoffDuration(attempt int) time.Duration {
	if s.current >= len(s.strategies) {
		return 0
	}
	return s.strategies[s.current].BackoffDuration(attempt)
}

// recoveryRegistry manages recovery strategies.
// Internal registry for storing and retrieving named strategies.
type recoveryRegistry struct {
	strategies map[string]RecoveryStrategy
}

var defaultRegistry = &recoveryRegistry{
	strategies: make(map[string]RecoveryStrategy),
}

// RegisterRecoveryStrategy registers a recovery strategy.
// Strategies can be retrieved by name for reuse across the application.
//
// Parameters:
//   - name: The strategy identifier
//   - strategy: The strategy implementation
func RegisterRecoveryStrategy(name string, strategy RecoveryStrategy) {
	defaultRegistry.strategies[name] = strategy
}

// GetRecoveryStrategy retrieves a recovery strategy by name.
//
// Parameters:
//   - name: The strategy identifier
//
// Returns the strategy and a boolean indicating if it was found.
func GetRecoveryStrategy(name string) (RecoveryStrategy, bool) {
	strategy, ok := defaultRegistry.strategies[name]
	return strategy, ok
}

// DefaultRecoveryStrategies returns commonly used strategies.
// Provides pre-configured strategies for common recovery patterns:
// - "exponential": Exponential backoff with jitter
// - "linear": Linear backoff
// - "no_retry": No retry strategy
// - "circuit": Circuit breaker pattern
//
// Returns a map of strategy names to implementations.
func DefaultRecoveryStrategies() map[string]RecoveryStrategy {
	return map[string]RecoveryStrategy{
		"exponential": NewExponentialBackoffStrategy(5, 100*time.Millisecond, 30*time.Second),
		"linear":      NewLinearBackoffStrategy(3, 1*time.Second),
		"no_retry":    NewNoRetryStrategy(),
		"circuit":     NewCircuitBreakerStrategy(5, 60*time.Second),
	}
}

// init registers default strategies.
// Automatically registers common recovery strategies on package initialization.
func init() {
	for name, strategy := range DefaultRecoveryStrategies() {
		RegisterRecoveryStrategy(name, strategy)
	}
}
