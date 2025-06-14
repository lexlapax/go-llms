// ABOUTME: Integration tests for library-wide error serialization
// ABOUTME: Tests ensure all domain errors are properly serializable for downstream usage

package errors_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/errors"
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
)

func TestLLMErrorSerialization(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{
			name:  "ProviderError",
			error: llmdomain.NewProviderError("openai", "Generate", 429, "Rate limit exceeded", nil),
		},
		{
			name:  "UnsupportedContentTypeError",
			error: llmdomain.NewUnsupportedContentTypeError("openai", llmdomain.ContentTypeImage),
		},
		{
			name:  "AuthenticationError",
			error: llmdomain.ErrAuthenticationFailed,
		},
		{
			name:  "RateLimitError",
			error: llmdomain.ErrRateLimitExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			if serErr, ok := tt.error.(interface {
				error
				ToJSON() ([]byte, error)
			}); ok {
				data, err := serErr.ToJSON()
				if err != nil {
					t.Fatalf("Failed to serialize error: %v", err)
				}

				// Verify JSON is valid
				var jsonData map[string]interface{}
				if err := json.Unmarshal(data, &jsonData); err != nil {
					t.Fatalf("Invalid JSON output: %v", err)
				}

				// Verify required fields
				if jsonData["type"] == nil {
					t.Error("Missing 'type' field in serialized error")
				}
				if jsonData["message"] == nil {
					t.Error("Missing 'message' field in serialized error")
				}
				if jsonData["timestamp"] == nil {
					t.Error("Missing 'timestamp' field in serialized error")
				}

				t.Logf("Serialized %s: %s", tt.name, string(data))
			} else {
				t.Errorf("Error %s does not implement SerializableError", tt.name)
			}
		})
	}
}

func TestAgentErrorSerialization(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{
			name:  "AgentError",
			error: domain.NewAgentError("agent-123", "test-agent", "run", domain.ErrToolExecution),
		},
		{
			name:  "ValidationError",
			error: domain.NewValidationError("field1", "invalid-value", "field is required"),
		},
		{
			name:  "ToolError",
			error: domain.NewToolError("weather", "execution", domain.ErrToolExecution),
		},
		{
			name: "MultiError",
			error: func() *domain.MultiError {
				me := domain.NewMultiError()
				me.Add(domain.ErrAgentNotFound)
				me.Add(domain.ErrToolNotFound)
				return me
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			if serErr, ok := tt.error.(interface {
				error
				ToJSON() ([]byte, error)
			}); ok {
				data, err := serErr.ToJSON()
				if err != nil {
					t.Fatalf("Failed to serialize error: %v", err)
				}

				// Verify JSON is valid
				var jsonData map[string]interface{}
				if err := json.Unmarshal(data, &jsonData); err != nil {
					t.Fatalf("Invalid JSON output: %v", err)
				}

				// Verify required fields
				if jsonData["type"] == nil {
					t.Error("Missing 'type' field in serialized error")
				}
				if jsonData["message"] == nil {
					t.Error("Missing 'message' field in serialized error")
				}

				t.Logf("Serialized %s: %s", tt.name, string(data))
			} else {
				t.Errorf("Error %s does not implement SerializableError", tt.name)
			}
		})
	}
}

func TestErrorRetryability(t *testing.T) {
	tests := []struct {
		name      string
		error     error
		retryable bool
		fatal     bool
	}{
		{
			name:      "RateLimitError",
			error:     llmdomain.ErrRateLimitExceeded,
			retryable: true,
			fatal:     false,
		},
		{
			name:      "AuthError",
			error:     llmdomain.ErrAuthenticationFailed,
			retryable: false,
			fatal:     true,
		},
		{
			name:      "AgentTimeout",
			error:     domain.ErrExecutionTimeout,
			retryable: true,
			fatal:     false,
		},
		{
			name:      "InvalidConfig",
			error:     domain.ErrInvalidConfiguration,
			retryable: false,
			fatal:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.IsRetryableError(tt.error); got != tt.retryable {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.retryable)
			}
			if got := errors.IsFatalError(tt.error); got != tt.fatal {
				t.Errorf("IsFatalError() = %v, want %v", got, tt.fatal)
			}
		})
	}
}

func TestErrorContextExtraction(t *testing.T) {
	// Create a provider error with context
	providerErr := llmdomain.NewProviderError("openai", "Generate", 429, "Rate limit exceeded", nil)

	// Extract context
	context := errors.GetErrorContext(providerErr)
	if context == nil {
		t.Fatal("Expected non-nil context")
	}

	// Verify context contains expected fields
	if context["provider"] != "openai" {
		t.Errorf("Expected provider='openai', got %v", context["provider"])
	}
	if context["operation"] != "Generate" {
		t.Errorf("Expected operation='Generate', got %v", context["operation"])
	}
	if context["status_code"] != 429 {
		t.Errorf("Expected status_code=429, got %v", context["status_code"])
	}

	t.Logf("Provider error context: %+v", context)
}

func TestErrorRecoveryStrategies(t *testing.T) {
	// Create an error with retry strategy
	retryableErr := llmdomain.ErrRateLimitExceeded
	if retryableErr.GetRecoveryStrategy() != nil {
		t.Logf("Error has recovery strategy: %s", retryableErr.GetRecoveryStrategy().Name())
	}

	// Test custom error with recovery
	customErr := errors.NewError("custom error").
		SetRetryable(true).
		WithRecovery(&exponentialBackoffStrategy{})

	if strategy := customErr.GetRecoveryStrategy(); strategy != nil {
		t.Logf("Custom error recovery strategy: %s", strategy.Name())
	} else {
		t.Error("Expected recovery strategy for retryable error")
	}
}

// exponentialBackoffStrategy is a simple recovery strategy for testing
type exponentialBackoffStrategy struct{}

func (e *exponentialBackoffStrategy) Name() string {
	return "exponential_backoff"
}

func (e *exponentialBackoffStrategy) CanRecover(err error) bool {
	return errors.IsRetryableError(err)
}

func (e *exponentialBackoffStrategy) Recover(err error, context map[string]interface{}) error {
	// Simple implementation for testing
	return nil
}

func (e *exponentialBackoffStrategy) MaxAttempts() int {
	return 3
}

func (e *exponentialBackoffStrategy) BackoffDuration(attempt int) time.Duration {
	return time.Duration(attempt*attempt) * time.Second
}

func TestBridgeCompatibility(t *testing.T) {
	// Test that all errors can be converted to map[string]interface{} for bridge usage
	errors := []error{
		llmdomain.NewProviderError("openai", "Generate", 500, "Server error", nil),
		domain.NewAgentError("agent-1", "test", "run", domain.ErrToolExecution),
		domain.NewValidationError("field", "value", "invalid"),
	}

	for i, err := range errors {
		t.Run(fmt.Sprintf("error_%d", i), func(t *testing.T) {
			if serErr, ok := err.(interface {
				error
				ToJSON() ([]byte, error)
				GetContext() map[string]interface{}
			}); ok {
				// Test JSON serialization (bridge-compatible)
				jsonData, err := serErr.ToJSON()
				if err != nil {
					t.Fatalf("Failed to serialize error for bridge: %v", err)
				}

				// Test deserialization to map
				var bridgeData map[string]interface{}
				if err := json.Unmarshal(jsonData, &bridgeData); err != nil {
					t.Fatalf("Failed to deserialize to bridge format: %v", err)
				}

				// Verify bridge-required fields
				requiredFields := []string{"type", "message", "timestamp", "context"}
				for _, field := range requiredFields {
					if bridgeData[field] == nil {
						t.Errorf("Missing required bridge field: %s", field)
					}
				}

				t.Logf("Bridge-compatible data: %+v", bridgeData)
			} else {
				t.Errorf("Error is not bridge-compatible (not SerializableError)")
			}
		})
	}
}
