// ABOUTME: Example demonstrating advanced error handling patterns for agents
// ABOUTME: Shows retry logic, fallback strategies, and error recovery techniques

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
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
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	BackoffFactor float64
}

// ErrorHandlingHook implements error tracking and recovery
type ErrorHandlingHook struct {
	errors []error
	recoveries int
}

func (h *ErrorHandlingHook) BeforeGenerate(ctx context.Context, messages []domain.Message) {
	// Could validate messages here
}

func (h *ErrorHandlingHook) AfterGenerate(ctx context.Context, response domain.Response, err error) {
	if err != nil {
		h.errors = append(h.errors, err)
		log.Printf("Generation error tracked: %v", err)
	}
}

func (h *ErrorHandlingHook) BeforeToolCall(ctx context.Context, tool string, params map[string]interface{}) {
	log.Printf("Calling tool: %s", tool)
}

func (h *ErrorHandlingHook) AfterToolCall(ctx context.Context, tool string, result interface{}, err error) {
	if err != nil {
		h.errors = append(h.errors, err)
		log.Printf("Tool error tracked: %s - %v", tool, err)
	}
}

// RunWithRetry executes an agent with exponential backoff retry
func RunWithRetry(agent domain.BaseAgent, state *domain.State, config RetryConfig) (*domain.State, error) {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		log.Printf("Attempt %d/%d", attempt, config.MaxAttempts)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Run the agent
		result, err := agent.Run(ctx, state)
		if err == nil {
			return result, nil
		}

		lastErr = err
		log.Printf("Attempt %d failed: %v", attempt, err)

		// Check if error is retryable
		if !isRetryableError(err) {
			log.Printf("Error is not retryable: %v", err)
			return nil, err
		}

		// Don't sleep after last attempt
		if attempt < config.MaxAttempts {
			log.Printf("Waiting %v before retry...", delay)
			time.Sleep(delay)

			// Exponential backoff
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	// Network timeouts and rate limits are retryable
	if errors.Is(err, ErrNetworkTimeout) || errors.Is(err, ErrRateLimited) {
		return true
	}

	// Context deadline exceeded is retryable
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check error message for common retryable patterns
	errMsg := err.Error()
	retryablePatterns := []string{
		"timeout",
		"temporary failure",
		"connection reset",
		"rate limit",
	}

	for _, pattern := range retryablePatterns {
		if contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}

// FallbackAgent wraps multiple agents with fallback logic
type FallbackAgent struct {
	*core.BaseAgentImpl
	primaryAgent   domain.BaseAgent
	fallbackAgents []domain.BaseAgent
}

func NewFallbackAgent(name string, primary domain.BaseAgent, fallbacks ...domain.BaseAgent) *FallbackAgent {
	return &FallbackAgent{
		BaseAgentImpl:  core.NewBaseAgent(name, "Fallback agent with error recovery", domain.AgentTypeCustom),
		primaryAgent:   primary,
		fallbackAgents: fallbacks,
	}
}

func (f *FallbackAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Try primary agent first
	log.Printf("Trying primary agent: %s", f.primaryAgent.Name())
	result, err := f.primaryAgent.Run(ctx, state)
	if err == nil {
		return result, nil
	}

	log.Printf("Primary agent failed: %v", err)

	// Try fallback agents
	for i, fallback := range f.fallbackAgents {
		log.Printf("Trying fallback agent %d: %s", i+1, fallback.Name())
		result, err = fallback.Run(ctx, state)
		if err == nil {
			log.Printf("Fallback agent %d succeeded", i+1)
			return result, nil
		}
		log.Printf("Fallback agent %d failed: %v", i+1, err)
	}

	return nil, fmt.Errorf("all agents failed, last error: %w", err)
}

func main() {
	fmt.Println("=== Advanced Error Handling Example ===\n")

	// Create a flaky tool that sometimes fails
	flakyTool := tools.NewTool(
		"flaky_api",
		"A tool that sometimes fails to demonstrate error handling",
		func(ctx domain.ToolContext, params struct {
			Query string `json:"query"`
		}) (map[string]interface{}, error) {
			// Simulate different failure modes
			switch params.Query {
			case "timeout":
				return nil, ErrNetworkTimeout
			case "ratelimit":
				return nil, ErrRateLimited
			case "invalid":
				return nil, ErrInvalidInput
			case "crash":
				return nil, ErrToolFailure
			default:
				return map[string]interface{}{
					"result": fmt.Sprintf("Success for query: %s", params.Query),
				}, nil
			}
		},
		&schemaDomain.Schema{
			Type: "object",
			Properties: map[string]schemaDomain.Property{
				"query": {
					Type:        "string",
					Description: "The query to process",
				},
			},
			Required: []string{"query"},
		},
	)

	// Create agents
	primaryAgent, _ := core.NewAgentFromString("primary", "mock")
	primaryAgent.AddTool(flakyTool)

	fallback1, _ := core.NewAgentFromString("fallback1", "mock")
	fallback2, _ := core.NewAgentFromString("fallback2", "mock")

	// Example 1: Retry with exponential backoff
	fmt.Println("Example 1: Retry with Exponential Backoff")
	fmt.Println("-" * 40)

	retryConfig := RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
	}

	state := domain.NewState()
	state.Set("user_input", "Call the flaky_api with query 'timeout'")

	result, err := RunWithRetry(primaryAgent, state, retryConfig)
	if err != nil {
		fmt.Printf("Failed after retries: %v\n", err)
	} else {
		output, _ := result.Get("output")
		fmt.Printf("Success: %v\n", output)
	}

	// Example 2: Fallback chain
	fmt.Println("\nExample 2: Fallback Chain")
	fmt.Println("-" * 40)

	fallbackAgent := NewFallbackAgent("fallback-chain", primaryAgent, fallback1, fallback2)

	state2 := domain.NewState()
	state2.Set("user_input", "Process this request")

	result2, err := fallbackAgent.Run(context.Background(), state2)
	if err != nil {
		fmt.Printf("All agents failed: %v\n", err)
	} else {
		output, _ := result2.Get("output")
		fmt.Printf("Success with fallback: %v\n", output)
	}

	// Example 3: Error tracking with hooks
	fmt.Println("\nExample 3: Error Tracking with Hooks")
	fmt.Println("-" * 40)

	errorHook := &ErrorHandlingHook{}
	trackingAgent, _ := core.NewAgentFromString("tracking", "mock")
	trackingAgent.AddTool(flakyTool)
	trackingAgent.WithHook(errorHook)

	// Try various operations
	queries := []string{"success", "timeout", "invalid", "crash"}
	for _, query := range queries {
		state := domain.NewState()
		state.Set("user_input", fmt.Sprintf("Call flaky_api with query '%s'", query))

		result, err := trackingAgent.Run(context.Background(), state)
		if err != nil {
			fmt.Printf("Query '%s' failed: %v\n", query, err)
		} else {
			output, _ := result.Get("output")
			fmt.Printf("Query '%s' succeeded: %v\n", query, output)
		}
	}

	fmt.Printf("\nTotal errors tracked: %d\n", len(errorHook.errors))

	// Example 4: Circuit breaker pattern
	fmt.Println("\nExample 4: Circuit Breaker Pattern")
	fmt.Println("-" * 40)

	circuitBreaker := NewCircuitBreaker(3, 5*time.Second)
	
	for i := 0; i < 5; i++ {
		err := circuitBreaker.Call(func() error {
			// Simulate failures
			if i < 3 {
				return ErrToolFailure
			}
			return nil
		})

		if err != nil {
			fmt.Printf("Call %d failed: %v\n", i+1, err)
		} else {
			fmt.Printf("Call %d succeeded\n", i+1)
		}
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	failureThreshold int
	resetTimeout     time.Duration
	failures         int
	lastFailureTime  time.Time
	state            string // "closed", "open", "half-open"
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: threshold,
		resetTimeout:     timeout,
		state:            "closed",
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	// Check if circuit should be reset
	if cb.state == "open" && time.Since(cb.lastFailureTime) > cb.resetTimeout {
		cb.state = "half-open"
		cb.failures = 0
		log.Println("Circuit breaker: half-open")
	}

	// If circuit is open, fail fast
	if cb.state == "open" {
		return fmt.Errorf("circuit breaker is open")
	}

	// Try the call
	err := fn()
	if err != nil {
		cb.failures++
		cb.lastFailureTime = time.Now()

		if cb.failures >= cb.failureThreshold {
			cb.state = "open"
			log.Printf("Circuit breaker: open (failures: %d)", cb.failures)
		}
		return err
	}

	// Success - reset the circuit
	if cb.state == "half-open" {
		cb.state = "closed"
		log.Println("Circuit breaker: closed")
	}
	cb.failures = 0
	return nil
}