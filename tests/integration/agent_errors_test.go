package integration

// ABOUTME: Integration tests for agent error handling scenarios
// ABOUTME: Tests provider errors, context cancellation, and timeout handling

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// TestAgentErrors tests error handling for agent workflows
func TestAgentErrors(t *testing.T) {
	testCases := []struct {
		name        string
		setupError  func(*provider.MockProvider)
		expectedErr string
	}{
		{
			name: "LLM provider error",
			setupError: func(mock *provider.MockProvider) {
				mock.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
					return ldomain.Response{}, errors.New("provider error")
				})
			},
			expectedErr: "provider error",
		},
		{
			name: "Context canceled",
			setupError: func(mock *provider.MockProvider) {
				mock.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
					return ldomain.Response{}, context.Canceled
				})
			},
			expectedErr: "context canceled",
		},
		{
			name: "Context deadline exceeded",
			setupError: func(mock *provider.MockProvider) {
				mock.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
					return ldomain.Response{}, context.DeadlineExceeded
				})
			},
			expectedErr: "context deadline exceeded",
		},
		{
			name: "Network error",
			setupError: func(mock *provider.MockProvider) {
				mock.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
					return ldomain.Response{}, errors.New("network connection failed")
				})
			},
			expectedErr: "network connection failed",
		},
		{
			name: "Rate limit error",
			setupError: func(mock *provider.MockProvider) {
				mock.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
					return ldomain.Response{}, errors.New("rate limit exceeded")
				})
			},
			expectedErr: "rate limit exceeded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock provider
			mockProvider := provider.NewMockProvider()
			tc.setupError(mockProvider)

			// Create an agent
			deps := core.LLMDeps{
				Provider: mockProvider,
			}
			agent := core.NewLLMAgent("error-test-agent", "test", deps)
			agent.SetSystemPrompt("You are a helpful assistant.")

			// Create test context
			ctx := context.Background()

			// Create initial state
			state := domain.NewState()
			state.Set("user_input", "Hello, how are you?")

			// Run the agent - should return error
			_, err := agent.Run(ctx, state)
			if err == nil {
				t.Fatal("Expected error from agent run, got nil")
			}

			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Errorf("Expected error to contain '%s', got: %v", tc.expectedErr, err)
			}
		})
	}
}

// TestAgentTimeoutHandling tests agent behavior with timeouts
func TestAgentTimeoutHandling(t *testing.T) {
	// Create a mock provider that delays response
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		// Simulate a slow response
		select {
		case <-time.After(2 * time.Second):
			return ldomain.Response{Content: "This should not be returned"}, nil
		case <-ctx.Done():
			return ldomain.Response{}, ctx.Err()
		}
	})

	// Create an agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("timeout-test-agent", "test", deps)
	agent.SetSystemPrompt("You are a helpful assistant.")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Hello!")

	// Run the agent - should timeout
	_, err := agent.Run(ctx, state)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded error, got: %v", err)
	}
}

// TestAgentProviderPanic tests agent behavior when provider panics
func TestAgentProviderPanic(t *testing.T) {
	// Create a mock provider that panics
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		panic("provider panic!")
	})

	// Create an agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("panic-test-agent", "test", deps)
	agent.SetSystemPrompt("You are a helpful assistant.")

	// Create test context
	ctx := context.Background()

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Hello!")

	// Run the agent - should handle panic gracefully
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic to propagate, but no panic occurred")
			}
		}()

		// This should panic
		_, _ = agent.Run(ctx, state)
	}()
}

// TestAgentEmptyResponse tests agent behavior with empty provider responses
func TestAgentEmptyResponse(t *testing.T) {
	// Create a mock provider that returns empty response
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{Content: ""}, nil
	})

	// Create an agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("empty-response-agent", "test", deps)
	agent.SetSystemPrompt("You are a helpful assistant.")

	// Create test context
	ctx := context.Background()

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Hello!")

	// Run the agent
	finalState, err := agent.Run(ctx, state)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check final output
	output, ok := finalState.Get("output")
	if !ok {
		t.Fatal("No output in final state")
	}

	outputStr, ok := output.(string)
	if !ok {
		t.Fatal("Output is not a string")
	}

	// Empty response should be preserved
	if outputStr != "" {
		t.Errorf("Expected empty output, got: %s", outputStr)
	}
}

// TestAgentInvalidState tests agent behavior with invalid initial state
func TestAgentInvalidState(t *testing.T) {
	// Create a mock provider
	mockProvider := provider.NewMockProvider()
	mockProvider.WithGenerateMessageFunc(func(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
		return ldomain.Response{Content: "Response"}, nil
	})

	// Create an agent
	deps := core.LLMDeps{
		Provider: mockProvider,
	}
	agent := core.NewLLMAgent("invalid-state-agent", "test", deps)
	agent.SetSystemPrompt("You are a helpful assistant.")

	// Create test context
	ctx := context.Background()

	// Test with nil state
	_, err := agent.Run(ctx, nil)
	if err == nil {
		t.Error("Expected error with nil state, got nil")
	}

	// Test with state missing user_input
	state := domain.NewState()
	// Don't set user_input
	
	// Run may succeed but with no user input
	finalState, err := agent.Run(ctx, state)
	if err != nil {
		// Some implementations may error, which is also fine
		return
	}

	// If no error, verify output exists (even if empty)
	_, ok := finalState.Get("output")
	if !ok {
		t.Error("Expected output field in final state")
	}
}