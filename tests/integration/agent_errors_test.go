package integration

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/workflow"
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock provider
			mockProvider := provider.NewMockProvider()

			// Set up the mock to return the error
			tc.setupError(mockProvider)

			// Create an agent
			agent := workflow.NewAgent(mockProvider)

			// Run the agent and expect an error
			_, err := agent.Run(context.Background(), "This should trigger an error")
			if err == nil {
				t.Fatal("Expected an error but got nil")
			}

			// Check that the error message contains the expected error
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.expectedErr)) {
				t.Errorf("Expected error to contain '%s', got: %v", tc.expectedErr, err)
			}
		})
	}
}