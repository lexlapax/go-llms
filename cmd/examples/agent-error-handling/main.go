package main

// ABOUTME: Example demonstrating error handling patterns in agent workflows
// ABOUTME: Shows retry logic, error recovery, validation, and custom error types

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// Custom error types for better error handling
var (
	ErrNetworkTimeout = errors.New("network timeout")
	ErrRateLimited    = errors.New("rate limited")
	ErrInvalidInput   = errors.New("invalid input")
	ErrToolFailure    = errors.New("tool execution failed")
)

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// ErrorHandlingHook implements domain.Hook for error tracking
type ErrorHandlingHook struct {
	errors     []error
	recoveries int
}

func NewErrorHandlingHook() *ErrorHandlingHook {
	return &ErrorHandlingHook{
		errors: make([]error, 0),
	}
}

func (h *ErrorHandlingHook) BeforeRun(ctx context.Context, agent domain.BaseAgent, state *domain.State) (context.Context, error) {
	// Validate state before execution
	if state == nil {
		return ctx, ErrInvalidInput
	}

	// Check for required fields
	if _, exists := state.Get("user_input"); !exists {
		return ctx, fmt.Errorf("%w: missing user_input", ErrInvalidInput)
	}

	return ctx, nil
}

func (h *ErrorHandlingHook) AfterRun(ctx context.Context, agent domain.BaseAgent, state *domain.State, result *domain.State, err error) error {
	if err != nil {
		h.errors = append(h.errors, err)
		log.Printf("[%s] Error tracked: %v", agent.Name(), err)

		// Attempt recovery for certain errors
		if errors.Is(err, ErrRateLimited) {
			h.recoveries++
			log.Printf("[%s] Rate limit recovery attempt %d", agent.Name(), h.recoveries)
		}
	}
	return nil
}

func (h *ErrorHandlingHook) Summary() {
	fmt.Printf("\nError Summary:\n")
	fmt.Printf("Total errors: %d\n", len(h.errors))
	fmt.Printf("Recovery attempts: %d\n", h.recoveries)

	if len(h.errors) > 0 {
		fmt.Println("\nError details:")
		for i, err := range h.errors {
			fmt.Printf("  %d. %v\n", i+1, err)
		}
	}
}

// RunWithRetry executes an agent with exponential backoff retry
func RunWithRetry(ctx context.Context, agent domain.BaseAgent, state *domain.State, config RetryConfig) (*domain.State, error) {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		fmt.Printf("Attempt %d/%d...\n", attempt, config.MaxAttempts)

		result, err := agent.Run(ctx, state)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err) {
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't sleep on last attempt
		if attempt < config.MaxAttempts {
			fmt.Printf("Error: %v. Retrying in %v...\n", err, delay)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			// Exponential backoff
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// isRetryable determines if an error should trigger a retry
func isRetryable(err error) bool {
	// Network timeouts and rate limits are retryable
	if errors.Is(err, ErrNetworkTimeout) || errors.Is(err, ErrRateLimited) {
		return true
	}

	// Context errors are not retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for specific error messages
	errStr := err.Error()
	retryablePatterns := []string{
		"timeout",
		"rate limit",
		"temporary failure",
		"connection refused",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
}

// FlakyAgent simulates an agent that fails intermittently
type FlakyAgent struct {
	domain.BaseAgent
	failureRate float64
	attempts    int
}

func NewFlakyAgent(name string, failureRate float64) *FlakyAgent {
	return &FlakyAgent{
		BaseAgent:   core.NewBaseAgent(name, "Simulates failures", domain.AgentTypeCustom),
		failureRate: failureRate,
	}
}

func (f *FlakyAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	f.attempts++

	// Simulate different types of failures
	if f.attempts == 1 {
		return nil, ErrNetworkTimeout
	}
	if f.attempts == 2 {
		return nil, ErrRateLimited
	}

	// Success on third attempt
	result := state.Clone()
	result.Set("output", fmt.Sprintf("Success after %d attempts!", f.attempts))
	result.Set("attempts", f.attempts)

	return result, nil
}

// ValidatingAgent demonstrates input validation
type ValidatingAgent struct {
	domain.BaseAgent
}

func NewValidatingAgent(name string) *ValidatingAgent {
	return &ValidatingAgent{
		BaseAgent: core.NewBaseAgent(name, "Validates input", domain.AgentTypeCustom),
	}
}

func (v *ValidatingAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Simple validation logic
	userInput, exists := state.Get("user_input")
	if !exists {
		return nil, fmt.Errorf("%w: user_input is required", ErrInvalidInput)
	}

	if str, ok := userInput.(string); ok && len(str) == 0 {
		return nil, fmt.Errorf("%w: user_input cannot be empty", ErrInvalidInput)
	}

	// Check for max_tokens if provided
	if maxTokens, exists := state.Get("max_tokens"); exists {
		if tokens, ok := maxTokens.(int); ok {
			if tokens < 1 || tokens > 4096 {
				return nil, fmt.Errorf("%w: max_tokens must be between 1 and 4096", ErrInvalidInput)
			}
		}
	}

	result := state.Clone()
	result.Set("output", "Input validated successfully")
	result.Set("validated", true)

	return result, nil
}

// CircuitBreakerAgent demonstrates circuit breaker pattern
type CircuitBreakerAgent struct {
	domain.BaseAgent
	failureCount     int
	successCount     int
	lastFailureTime  time.Time
	state            string // "closed", "open", "half-open"
	failureThreshold int
	successThreshold int
	timeout          time.Duration
}

func NewCircuitBreakerAgent(name string) *CircuitBreakerAgent {
	return &CircuitBreakerAgent{
		BaseAgent:        core.NewBaseAgent(name, "Circuit breaker pattern", domain.AgentTypeCustom),
		state:            "closed",
		failureThreshold: 3,
		successThreshold: 2,
		timeout:          5 * time.Second,
	}
}

func (c *CircuitBreakerAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Check circuit breaker state
	switch c.state {
	case "open":
		// Check if timeout has passed
		if time.Since(c.lastFailureTime) > c.timeout {
			c.state = "half-open"
			c.successCount = 0
			fmt.Println("Circuit breaker: half-open (trying again)")
		} else {
			return nil, fmt.Errorf("circuit breaker is open, service unavailable")
		}

	case "half-open":
		// In half-open state, we're testing if the service is back
		fmt.Println("Circuit breaker: testing service...")
	}

	// Simulate service call (50% failure rate for demo)
	if time.Now().Unix()%2 == 0 {
		// Failure
		c.failureCount++
		c.successCount = 0
		c.lastFailureTime = time.Now()

		if c.failureCount >= c.failureThreshold {
			c.state = "open"
			fmt.Printf("Circuit breaker: OPEN (failures: %d)\n", c.failureCount)
		}

		return nil, fmt.Errorf("service temporarily unavailable")
	}

	// Success
	c.successCount++

	if c.state == "half-open" && c.successCount >= c.successThreshold {
		c.state = "closed"
		c.failureCount = 0
		fmt.Println("Circuit breaker: CLOSED (service recovered)")
	}

	result := state.Clone()
	result.Set("output", "Service call successful")
	result.Set("circuit_state", c.state)

	return result, nil
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Error Handling Examples ===")

	// Example 1: Basic retry with exponential backoff
	retryExample(ctx)

	// Example 2: Input validation
	validationExample(ctx)

	// Example 3: Error tracking with hooks
	errorTrackingExample(ctx)

	// Example 4: Circuit breaker pattern
	circuitBreakerExample(ctx)
}

func retryExample(ctx context.Context) {
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("Example 1: Retry with Exponential Backoff")
	fmt.Println(strings.Repeat("-", 40))

	// Create a flaky agent that fails initially
	agent := NewFlakyAgent("flaky-service", 0.7)

	// Configure retry behavior
	retryConfig := RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      2 * time.Second,
		BackoffFactor: 2.0,
	}

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Process this request")

	// Run with retry
	fmt.Println("\nRunning agent with retry logic...")
	result, err := RunWithRetry(ctx, agent, state, retryConfig)

	if err != nil {
		fmt.Printf("Failed after retries: %v\n", err)
	} else {
		if output, exists := result.Get("output"); exists {
			fmt.Printf("Success: %v\n", output)
		}
		if attempts, exists := result.Get("attempts"); exists {
			fmt.Printf("Total attempts: %v\n", attempts)
		}
	}

	fmt.Println()
}

func validationExample(ctx context.Context) {
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("Example 2: Input Validation")
	fmt.Println(strings.Repeat("-", 40))

	agent := NewValidatingAgent("input-validator")

	// Test with invalid input (missing required field)
	fmt.Println("\nTest 1: Missing required field")
	state1 := domain.NewState()
	_, err := agent.Run(ctx, state1)
	if err != nil {
		fmt.Printf("Validation error (expected): %v\n", err)
	}

	// Test with empty input
	fmt.Println("\nTest 2: Empty input")
	state2 := domain.NewState()
	state2.Set("user_input", "")
	_, err = agent.Run(ctx, state2)
	if err != nil {
		fmt.Printf("Validation error (expected): %v\n", err)
	}

	// Test with invalid max_tokens
	fmt.Println("\nTest 3: Invalid max_tokens")
	state3 := domain.NewState()
	state3.Set("user_input", "Valid input")
	state3.Set("max_tokens", 5000)
	_, err = agent.Run(ctx, state3)
	if err != nil {
		fmt.Printf("Validation error (expected): %v\n", err)
	}

	// Test with valid input
	fmt.Println("\nTest 4: Valid input")
	state4 := domain.NewState()
	state4.Set("user_input", "Valid input text")
	state4.Set("max_tokens", 1000)
	result, err := agent.Run(ctx, state4)
	if err != nil {
		fmt.Printf("Unexpected error: %v\n", err)
	} else {
		if output, exists := result.Get("output"); exists {
			fmt.Printf("Success: %v\n", output)
		}
	}

	fmt.Println()
}

func errorTrackingExample(ctx context.Context) {
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("Example 3: Error Tracking with Hooks")
	fmt.Println(strings.Repeat("-", 40))

	// Create error handling hook
	errorHook := NewErrorHandlingHook()

	// Create agent with hook
	agent := NewFlakyAgent("tracked-agent", 0.5)

	// Create a simple wrapper to apply hooks
	// In real usage, you would use agent.WithHook() if available
	hookedRun := func(ctx context.Context, state *domain.State) (*domain.State, error) {
		ctx, err := errorHook.BeforeRun(ctx, agent, state)
		if err != nil {
			return nil, err
		}

		result, runErr := agent.Run(ctx, state)
		_ = errorHook.AfterRun(ctx, agent, state, result, runErr)

		return result, runErr
	}

	// Run multiple times to accumulate errors
	fmt.Println("\nRunning agent multiple times...")
	for i := 1; i <= 3; i++ {
		fmt.Printf("\nRun %d:\n", i)
		state := domain.NewState()
		state.Set("user_input", fmt.Sprintf("Request %d", i))

		result, err := hookedRun(ctx, state)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else if output, exists := result.Get("output"); exists {
			fmt.Printf("Success: %v\n", output)
		}

		// Reset agent attempts for demo
		agent.attempts = 0
	}

	// Show error summary
	errorHook.Summary()
	fmt.Println()
}

func circuitBreakerExample(ctx context.Context) {
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("Example 4: Circuit Breaker Pattern")
	fmt.Println(strings.Repeat("-", 40))

	agent := NewCircuitBreakerAgent("protected-service")

	fmt.Println("\nTesting circuit breaker with multiple calls...")

	// Make multiple calls to demonstrate circuit breaker behavior
	for i := 1; i <= 10; i++ {
		fmt.Printf("\nCall %d: ", i)

		state := domain.NewState()
		state.Set("user_input", fmt.Sprintf("Request %d", i))

		result, err := agent.Run(ctx, state)
		if err != nil {
			fmt.Printf("Failed: %v", err)
		} else if output, exists := result.Get("output"); exists {
			fmt.Printf("Success: %v", output)
			if circuitState, exists := result.Get("circuit_state"); exists {
				fmt.Printf(" (circuit: %v)", circuitState)
			}
		}

		// Small delay between calls
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println()
}
