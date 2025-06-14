package errors

import (
	"fmt"
	"testing"
	"time"
)

// TestExponentialBackoffStrategy tests exponential backoff
func TestExponentialBackoffStrategy(t *testing.T) {
	strategy := NewExponentialBackoffStrategy(5, 100*time.Millisecond, 5*time.Second)

	if strategy.Name() != "exponential_backoff" {
		t.Errorf("expected name 'exponential_backoff', got %s", strategy.Name())
	}

	if strategy.MaxAttempts() != 5 {
		t.Errorf("expected max attempts 5, got %d", strategy.MaxAttempts())
	}

	// Test backoff durations
	prevDuration := time.Duration(0)
	for i := 1; i <= 5; i++ {
		duration := strategy.BackoffDuration(i)

		// Should increase (allowing for jitter)
		if i > 1 && duration <= prevDuration {
			// With jitter, might occasionally be slightly less
			if duration < prevDuration/2 {
				t.Errorf("expected increasing duration at attempt %d", i)
			}
		}

		// Should not exceed max
		if duration > strategy.maxDelay {
			t.Errorf("duration %v exceeds max %v", duration, strategy.maxDelay)
		}

		prevDuration = duration
	}

	// Test with 0 attempt
	if strategy.BackoffDuration(0) != 0 {
		t.Error("expected 0 duration for 0 attempt")
	}
}

// TestLinearBackoffStrategy tests linear backoff
func TestLinearBackoffStrategy(t *testing.T) {
	strategy := NewLinearBackoffStrategy(3, 500*time.Millisecond)

	if strategy.Name() != "linear_backoff" {
		t.Errorf("expected name 'linear_backoff', got %s", strategy.Name())
	}

	if strategy.MaxAttempts() != 3 {
		t.Errorf("expected max attempts 3, got %d", strategy.MaxAttempts())
	}

	// Test backoff durations
	for i := 1; i <= 3; i++ {
		expected := time.Duration(i) * 500 * time.Millisecond
		if strategy.BackoffDuration(i) != expected {
			t.Errorf("expected duration %v for attempt %d, got %v",
				expected, i, strategy.BackoffDuration(i))
		}
	}
}

// TestNoRetryStrategy tests no retry strategy
func TestNoRetryStrategy(t *testing.T) {
	strategy := NewNoRetryStrategy()

	if strategy.Name() != "no_retry" {
		t.Errorf("expected name 'no_retry', got %s", strategy.Name())
	}

	if strategy.MaxAttempts() != 0 {
		t.Errorf("expected max attempts 0, got %d", strategy.MaxAttempts())
	}

	err := fmt.Errorf("test error")

	if strategy.CanRecover(err) {
		t.Error("expected CanRecover to return false")
	}

	if strategy.Recover(err, nil) != err {
		t.Error("expected Recover to return original error")
	}

	if strategy.BackoffDuration(1) != 0 {
		t.Error("expected BackoffDuration to return 0")
	}
}

// TestFallbackStrategy tests fallback strategy
func TestFallbackStrategy(t *testing.T) {
	fallbackCalled := false
	fallbackFunc := func(err error, context map[string]interface{}) error {
		fallbackCalled = true
		return nil
	}

	strategy := NewFallbackStrategy(fallbackFunc)

	if strategy.Name() != "fallback" {
		t.Errorf("expected name 'fallback', got %s", strategy.Name())
	}

	if strategy.MaxAttempts() != 1 {
		t.Errorf("expected max attempts 1, got %d", strategy.MaxAttempts())
	}

	err := fmt.Errorf("test error")

	if !strategy.CanRecover(err) {
		t.Error("expected CanRecover to return true")
	}

	result := strategy.Recover(err, nil)
	if result != nil {
		t.Error("expected Recover to return nil")
	}

	if !fallbackCalled {
		t.Error("expected fallback function to be called")
	}

	if strategy.BackoffDuration(1) != 0 {
		t.Error("expected BackoffDuration to return 0")
	}

	// Test with nil fallback
	nilStrategy := NewFallbackStrategy(nil)
	if nilStrategy.CanRecover(err) {
		t.Error("expected CanRecover to return false for nil fallback")
	}
}

// TestCircuitBreakerStrategy tests circuit breaker
func TestCircuitBreakerStrategy(t *testing.T) {
	strategy := NewCircuitBreakerStrategy(2, 100*time.Millisecond)

	if strategy.Name() != "circuit_breaker" {
		t.Errorf("expected name 'circuit_breaker', got %s", strategy.Name())
	}

	err := fmt.Errorf("test error")

	// Initially closed
	if !strategy.CanRecover(err) {
		t.Error("expected CanRecover to return true when closed")
	}

	// First failure
	result := strategy.Recover(err, nil)
	if result == nil {
		t.Error("expected error from Recover")
	}

	// Second failure - should open
	result = strategy.Recover(err, nil)
	if result == nil {
		t.Error("expected error from Recover")
	}

	// Circuit should be open
	if strategy.state != "open" {
		t.Error("expected circuit to be open")
	}

	if strategy.CanRecover(err) {
		t.Error("expected CanRecover to return false when open")
	}

	// Wait for reset timeout
	time.Sleep(150 * time.Millisecond)

	// Should transition to half-open
	if !strategy.CanRecover(err) {
		t.Error("expected CanRecover to return true after timeout")
	}

	if strategy.state != "half-open" {
		t.Error("expected circuit to be half-open")
	}

	// Success should close circuit
	result = strategy.Recover(nil, nil)
	if result != nil {
		t.Error("expected nil from successful Recover")
	}

	if strategy.state != "closed" {
		t.Error("expected circuit to be closed after success")
	}
}

// TestCompositeStrategy tests composite strategy
func TestCompositeStrategy(t *testing.T) {
	strategy1 := NewNoRetryStrategy()
	strategy2 := NewLinearBackoffStrategy(2, time.Millisecond)

	composite := NewCompositeStrategy(strategy1, strategy2)

	if composite.Name() != "composite" {
		t.Errorf("expected name 'composite', got %s", composite.Name())
	}

	if composite.MaxAttempts() != 2 {
		t.Errorf("expected max attempts 2, got %d", composite.MaxAttempts())
	}

	// Create a retryable error for testing
	retryableErr := NewError("test").SetRetryable(true)

	if !composite.CanRecover(retryableErr) {
		t.Error("expected CanRecover to return true")
	}

	// First strategy (no retry) should fail, move to second
	result := composite.Recover(retryableErr, nil)
	if result == nil {
		t.Error("expected error from Recover")
	}

	// Should be using second strategy now
	if composite.current != 1 {
		t.Error("expected to move to second strategy")
	}

	// Test backoff delegation
	backoff := composite.BackoffDuration(1)
	if backoff != time.Millisecond {
		t.Errorf("expected backoff %v, got %v", time.Millisecond, backoff)
	}
}

// TestRecoveryRegistry tests strategy registration
func TestRecoveryRegistry(t *testing.T) {
	// Test default strategies
	strategy, ok := GetRecoveryStrategy("exponential")
	if !ok || strategy == nil {
		t.Error("expected to find exponential strategy")
	}

	strategy, ok = GetRecoveryStrategy("linear")
	if !ok || strategy == nil {
		t.Error("expected to find linear strategy")
	}

	strategy, ok = GetRecoveryStrategy("no_retry")
	if !ok || strategy == nil {
		t.Error("expected to find no_retry strategy")
	}

	strategy, ok = GetRecoveryStrategy("circuit")
	if !ok || strategy == nil {
		t.Error("expected to find circuit strategy")
	}

	// Test custom registration
	custom := NewNoRetryStrategy()
	RegisterRecoveryStrategy("custom", custom)

	strategy, ok = GetRecoveryStrategy("custom")
	if !ok || strategy != custom {
		t.Error("expected to find custom strategy")
	}

	// Test non-existent strategy
	_, ok = GetRecoveryStrategy("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent strategy")
	}
}

// TestDefaultRecoveryStrategies tests default strategies
func TestDefaultRecoveryStrategies(t *testing.T) {
	defaults := DefaultRecoveryStrategies()

	expectedStrategies := []string{"exponential", "linear", "no_retry", "circuit"}

	for _, name := range expectedStrategies {
		if _, ok := defaults[name]; !ok {
			t.Errorf("expected default strategy %s", name)
		}
	}

	if len(defaults) != len(expectedStrategies) {
		t.Errorf("expected %d default strategies, got %d",
			len(expectedStrategies), len(defaults))
	}
}

// TestCanRecoverWithRetryableError tests CanRecover with retryable errors
func TestCanRecoverWithRetryableError(t *testing.T) {
	retryableErr := NewError("test").SetRetryable(true)
	nonRetryableErr := NewError("test").SetRetryable(false)

	// Exponential backoff should check retryability
	strategy := NewExponentialBackoffStrategy(3, time.Millisecond, time.Second)

	if !strategy.CanRecover(retryableErr) {
		t.Error("expected CanRecover to return true for retryable error")
	}

	if strategy.CanRecover(nonRetryableErr) {
		t.Error("expected CanRecover to return false for non-retryable error")
	}
}

// TestBackoffDurationEdgeCases tests edge cases for backoff
func TestBackoffDurationEdgeCases(t *testing.T) {
	strategy := NewExponentialBackoffStrategy(5, 100*time.Millisecond, 1*time.Second)

	// Test negative attempt
	if strategy.BackoffDuration(-1) != 0 {
		t.Error("expected 0 duration for negative attempt")
	}

	// Test very large attempt (should cap at max, accounting for jitter)
	duration := strategy.BackoffDuration(100)
	// Allow up to 10% over max due to jitter
	maxWithJitter := time.Duration(float64(strategy.maxDelay) * 1.1)
	if duration > maxWithJitter {
		t.Errorf("expected duration to be capped at max (with jitter), got %v", duration)
	}
}

// TestCircuitBreakerBackoffDuration tests circuit breaker backoff
func TestCircuitBreakerBackoffDuration(t *testing.T) {
	strategy := NewCircuitBreakerStrategy(1, 100*time.Millisecond)

	// Trigger circuit open
	err := fmt.Errorf("test")
	_ = strategy.Recover(err, nil)

	// Should return remaining time when open
	backoff := strategy.BackoffDuration(1)
	if backoff <= 0 || backoff > 100*time.Millisecond {
		t.Errorf("expected positive backoff less than reset timeout, got %v", backoff)
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Should return 0 after timeout
	backoff = strategy.BackoffDuration(1)
	if backoff != 0 {
		t.Errorf("expected 0 backoff after timeout, got %v", backoff)
	}
}

// TestCompositeStrategyExhaustion tests when all strategies fail
func TestCompositeStrategyExhaustion(t *testing.T) {
	strategy1 := NewNoRetryStrategy()
	strategy2 := NewNoRetryStrategy()

	composite := NewCompositeStrategy(strategy1, strategy2)

	err := fmt.Errorf("test")

	// Move through all strategies
	for i := 0; i < 3; i++ {
		result := composite.Recover(err, nil)
		if result == nil {
			t.Error("expected error from Recover")
		}
	}

	// Should be exhausted
	if composite.current < len(composite.strategies) {
		t.Error("expected all strategies to be exhausted")
	}

	// Backoff should be 0 when exhausted
	if composite.BackoffDuration(1) != 0 {
		t.Error("expected 0 backoff when exhausted")
	}
}
