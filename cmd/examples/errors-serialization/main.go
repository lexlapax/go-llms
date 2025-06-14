// Package main demonstrates enhanced error handling features
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lexlapax/go-llms/pkg/errors"
)

// simulateAPICall simulates an API call that might fail
func simulateAPICall(failureType string) error {
	switch failureType {
	case "retryable":
		return errors.NewError("temporary network issue").
			WithCode("NET_TIMEOUT").
			SetRetryable(true).
			WithContext("endpoint", "/api/users").
			WithContext("timeout", "30s")

	case "fatal":
		return errors.NewError("database connection lost").
			WithCode("DB_FATAL").
			SetFatal(true).
			WithContext("database", "primary").
			WithContext("connection_pool", "exhausted")

	case "validation":
		return errors.NewErrorBuilder("invalid request data").
			WithCode("VALIDATION_ERROR").
			WithContext("field", "email").
			WithContext("value", "invalid-email").
			WithContext("constraint", "must be valid email format").
			Build()

	default:
		return nil
	}
}

// processWithRetry demonstrates retry logic with recovery strategies
func processWithRetry(_ context.Context) error {
	// Create an error with exponential backoff recovery
	err := errors.NewError("service temporarily unavailable").
		WithCode("SERVICE_UNAVAILABLE").
		SetRetryable(true).
		WithRecovery(errors.NewExponentialBackoffStrategy(5, 100*time.Millisecond, 5*time.Second))

	// Get recovery strategy
	if strategy := err.GetRecoveryStrategy(); strategy != nil {
		fmt.Printf("Using recovery strategy: %s\n", strategy.Name())
		fmt.Printf("Max attempts: %d\n", strategy.MaxAttempts())

		// Simulate retry loop
		for attempt := 1; attempt <= strategy.MaxAttempts(); attempt++ {
			fmt.Printf("\nAttempt %d/%d\n", attempt, strategy.MaxAttempts())

			// Check if we can recover
			if !strategy.CanRecover(err) {
				fmt.Println("Cannot recover from this error")
				return err
			}

			// Simulate operation
			if attempt < 3 {
				fmt.Println("Operation failed, will retry...")

				// Calculate backoff
				backoff := strategy.BackoffDuration(attempt)
				fmt.Printf("Waiting %v before retry\n", backoff)
				time.Sleep(backoff)
			} else {
				fmt.Println("Operation succeeded!")
				return nil
			}
		}
	}

	return err
}

// demonstrateErrorAggregation shows error aggregation features
func demonstrateErrorAggregation() {
	fmt.Println("\n=== Error Aggregation Demo ===")

	agg := errors.NewErrorAggregator()

	// Simulate multiple operations that might fail
	operations := []struct {
		name string
		err  error
	}{
		{
			name: "user_validation",
			err:  simulateAPICall("validation"),
		},
		{
			name: "api_call",
			err:  simulateAPICall("retryable"),
		},
		{
			name: "database_update",
			err:  nil, // Success
		},
	}

	// Collect errors with context
	for _, op := range operations {
		if op.err != nil {
			agg.AddWithContext(op.err, map[string]interface{}{
				"operation": op.name,
				"timestamp": time.Now(),
			})
		}
	}

	// Convert to serializable error
	if agg.HasErrors() {
		serializable := agg.ToSerializable()

		// Serialize to JSON
		data, _ := serializable.ToJSON()
		fmt.Println("\nAggregated errors as JSON:")
		fmt.Println(string(data))
	}
}

// demonstrateErrorContext shows context enrichment
func demonstrateErrorContext() {
	fmt.Println("\n=== Error Context Demo ===")

	// Create a base error
	err := fmt.Errorf("connection refused")

	// Enrich with operation context
	enriched := errors.EnrichErrorWithOperation(err, "FetchUserProfile")

	// Further enrich with request context
	enriched = errors.EnrichErrorWithRequest(enriched, "GET", "https://api.example.com/users/123", 500)

	// Add resource context
	enriched = errors.EnrichErrorWithResource(enriched, "User", "123")

	// Add runtime context
	enriched = errors.EnrichErrorWithRuntime(enriched)

	// Convert to BaseError to access context
	if be, ok := enriched.(*errors.BaseError); ok {
		fmt.Println("\nError with enriched context:")
		data, _ := be.ToJSON()

		// Pretty print JSON
		var pretty map[string]interface{}
		if err := json.Unmarshal(data, &pretty); err != nil {
			fmt.Printf("Error unmarshaling JSON: %v\n", err)
			return
		}
		prettyJSON, _ := json.MarshalIndent(pretty, "", "  ")
		fmt.Println(string(prettyJSON))
	}
}

// demonstrateErrorBuilder shows the builder pattern
func demonstrateErrorBuilder() {
	fmt.Println("\n=== Error Builder Demo ===")

	// Build a comprehensive error
	err := errors.NewErrorBuilder("failed to process payment").
		WithCode("PAYMENT_FAILED").
		WithType("PaymentError").
		WithCause(fmt.Errorf("insufficient funds")).
		WithContext("amount", 150.00).
		WithContext("currency", "USD").
		WithContext("merchant", "ACME Corp").
		WithContextMap(map[string]interface{}{
			"transaction_id": "TXN-12345",
			"customer_id":    "CUST-67890",
		}).
		WithRetryable(true).
		WithRecovery(errors.NewLinearBackoffStrategy(3, 2*time.Second)).
		Build()

	fmt.Println("Built error:")
	data, _ := err.ToJSON()
	fmt.Println(string(data))
}

// demonstrateCircuitBreaker shows circuit breaker pattern
func demonstrateCircuitBreaker() {
	fmt.Println("\n=== Circuit Breaker Demo ===")

	circuit := errors.NewCircuitBreakerStrategy(3, 5*time.Second)

	// Simulate failures
	for i := 1; i <= 5; i++ {
		fmt.Printf("\nRequest %d:\n", i)

		err := fmt.Errorf("service error")

		if circuit.CanRecover(err) {
			fmt.Println("Circuit allows request")

			// Simulate failure
			result := circuit.Recover(err, nil)
			if result != nil {
				fmt.Printf("Request failed: %v\n", result)
			}
		} else {
			fmt.Println("Circuit is OPEN - blocking request")
			backoff := circuit.BackoffDuration(1)
			fmt.Printf("Must wait %v before retry\n", backoff)

			if i == 4 {
				// Wait for circuit to reset
				fmt.Println("\nWaiting for circuit reset...")
				time.Sleep(backoff)
			}
		}
	}
}

// demonstrateCompositeStrategy shows composite recovery strategies
func demonstrateCompositeStrategy() {
	fmt.Println("\n=== Composite Strategy Demo ===")

	// Create multiple strategies
	fallback := errors.NewFallbackStrategy(func(err error, ctx map[string]interface{}) error {
		fmt.Println("Using fallback value")
		return nil
	})

	exponential := errors.NewExponentialBackoffStrategy(3, 100*time.Millisecond, 1*time.Second)

	// Combine strategies
	composite := errors.NewCompositeStrategy(exponential, fallback)

	err := errors.NewError("primary service failed").SetRetryable(true)

	fmt.Printf("Composite strategy with %d total attempts\n", composite.MaxAttempts())

	// Simulate recovery attempts
	for i := 1; i <= 4; i++ {
		fmt.Printf("\nAttempt %d:\n", i)

		if composite.CanRecover(err) {
			result := composite.Recover(err, nil)
			if result == nil {
				fmt.Println("Recovery successful!")
				break
			}

			backoff := composite.BackoffDuration(i)
			if backoff > 0 {
				fmt.Printf("Waiting %v before next attempt\n", backoff)
				time.Sleep(backoff)
			}
		}
	}
}

func main() {
	fmt.Println("=== Enhanced Error Handling Examples ===")

	// Demonstrate different error scenarios
	fmt.Println("\n1. Basic Error Creation and Serialization:")
	err := simulateAPICall("retryable")
	if serr, ok := err.(*errors.BaseError); ok {
		data, _ := serr.ToJSON()
		fmt.Println(string(data))
	}

	// Demonstrate retry with recovery
	fmt.Println("\n2. Retry with Recovery Strategy:")
	ctx := context.Background()
	if err := processWithRetry(ctx); err != nil {
		log.Printf("Process failed after retries: %v", err)
	}

	// Demonstrate error aggregation
	demonstrateErrorAggregation()

	// Demonstrate error context
	demonstrateErrorContext()

	// Demonstrate error builder
	demonstrateErrorBuilder()

	// Demonstrate circuit breaker
	demonstrateCircuitBreaker()

	// Demonstrate composite strategy
	demonstrateCompositeStrategy()

	fmt.Println("\n=== Examples Complete ===")
}
