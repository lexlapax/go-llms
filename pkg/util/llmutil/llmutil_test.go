package llmutil

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestCreateProvider(t *testing.T) {
	tests := []struct {
		name          string
		config        ModelConfig
		expectError   bool
		expectedError string
	}{
		{
			name: "Valid OpenAI config",
			config: ModelConfig{
				Provider: "openai",
				Model:    "gpt-4o",
				APIKey:   "test-api-key",
			},
			expectError: false,
		},
		{
			name: "Valid Anthropic config",
			config: ModelConfig{
				Provider: "anthropic",
				Model:    "claude-3-5-sonnet-latest",
				APIKey:   "test-api-key",
			},
			expectError: false,
		},
		{
			name: "Valid mock config",
			config: ModelConfig{
				Provider: "mock",
				Model:    "mock-model",
				APIKey:   "not-needed",
			},
			expectError: false,
		},
		{
			name: "Missing API key",
			config: ModelConfig{
				Provider: "openai",
				Model:    "gpt-4o",
			},
			expectError:   true,
			expectedError: "API key is required",
		},
		{
			name: "Unsupported provider",
			config: ModelConfig{
				Provider: "unsupported",
				Model:    "model",
				APIKey:   "test-api-key",
			},
			expectError:   true,
			expectedError: "unsupported provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := CreateProvider(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.expectedError != "" && !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error with %q, got %q", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if provider == nil {
					t.Errorf("Expected provider but got nil")
				}
			}
		})
	}
}

func TestBatchGenerate(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	prompts := []string{
		"Prompt 1",
		"Prompt 2",
		"Prompt 3",
	}

	results, errors := BatchGenerate(context.Background(), mockProvider, prompts)

	if len(results) != len(prompts) {
		t.Errorf("Expected %d results, got %d", len(prompts), len(results))
	}

	if len(errors) != len(prompts) {
		t.Errorf("Expected %d errors, got %d", len(prompts), len(errors))
	}

	for i, err := range errors {
		if err != nil {
			t.Errorf("Unexpected error for prompt %d: %v", i, err)
		}
	}

	for i, result := range results {
		if result == "" {
			t.Errorf("Expected non-empty result for prompt %d", i)
		}
	}
}

func TestGenerateWithRetry(t *testing.T) {
	t.Run("Success on first try", func(t *testing.T) {
		mockProvider := provider.NewMockProvider()
		result, err := GenerateWithRetry(context.Background(), mockProvider, "Test prompt", 3)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == "" {
			t.Errorf("Expected non-empty result")
		}
	})

	t.Run("Non-retryable error", func(t *testing.T) {
		// Create a failing provider that returns a non-retryable error
		failingProvider := &mockFailingProvider{
			err:       domain.ErrInvalidModelParameters,
			failCount: 1,
		}

		_, err := GenerateWithRetry(context.Background(), failingProvider, "Test prompt", 3)
		if err == nil {
			t.Errorf("Expected error but got nil")
		}
	})

	t.Run("Retryable error with eventual success", func(t *testing.T) {
		// Create a provider that fails with retryable errors but succeeds after N attempts
		failingProvider := &mockFailingProvider{
			err:       domain.ErrNetworkConnectivity,
			failCount: 2, // Fail twice then succeed
		}

		result, err := GenerateWithRetry(context.Background(), failingProvider, "Test prompt", 3)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == "" {
			t.Errorf("Expected non-empty result")
		}
	})

	t.Run("Retryable error with max retries exceeded", func(t *testing.T) {
		// Create a provider that always fails with retryable errors
		failingProvider := &mockFailingProvider{
			err:       domain.ErrNetworkConnectivity,
			failCount: 10, // More than max retries
		}

		_, err := GenerateWithRetry(context.Background(), failingProvider, "Test prompt", 3)
		if err == nil {
			t.Errorf("Expected error but got nil")
		}
		if !strings.Contains(err.Error(), "max retries reached") {
			t.Errorf("Expected 'max retries reached' error, got: %v", err)
		}
	})
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectRetryable bool
	}{
		{
			name:           "nil error",
			err:            nil,
			expectRetryable: false,
		},
		{
			name:           "network connectivity error",
			err:            domain.ErrNetworkConnectivity,
			expectRetryable: true,
		},
		{
			name:           "rate limit error",
			err:            domain.ErrRateLimitExceeded,
			expectRetryable: true,
		},
		{
			name:           "invalid model parameters error",
			err:            domain.ErrInvalidModelParameters,
			expectRetryable: false,
		},
		{
			name:           "token quota exceeded error",
			err:            domain.ErrTokenQuotaExceeded,
			expectRetryable: false,
		},
		{
			name:           "other error",
			err:            errors.New("some other error"),
			expectRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.expectRetryable {
				t.Errorf("Expected IsRetryableError to return %v, got %v", tt.expectRetryable, result)
			}
		})
	}
}

// Helper functions for the test

// Mock provider implementation for testing retries
type mockFailingProvider struct {
	attempts  int
	failCount int
	err       error
}

func (m *mockFailingProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
	m.attempts++
	if m.attempts <= m.failCount {
		return "", m.err
	}
	return "Success after retries", nil
}

func (m *mockFailingProvider) GenerateMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.Response, error) {
	result, err := m.Generate(ctx, "message-based", options...)
	return domain.Response{Content: result}, err
}

func (m *mockFailingProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *schemaDomain.Schema, options ...domain.Option) (interface{}, error) {
	result, err := m.Generate(ctx, prompt, options...)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"result": result}, nil
}

func (m *mockFailingProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
	_, err := m.Generate(ctx, prompt, options...)
	if err != nil {
		return nil, err
	}
	return makeMockStream("Success"), nil
}

func (m *mockFailingProvider) StreamMessage(ctx context.Context, messages []domain.Message, options ...domain.Option) (domain.ResponseStream, error) {
	_, err := m.Generate(ctx, "message-based", options...)
	if err != nil {
		return nil, err
	}
	return makeMockStream("Success"), nil
}

// Helper function to create a simple mock response stream
func makeMockStream(content string) domain.ResponseStream {
	ch := make(chan domain.Token, 1)
	
	// Start a goroutine to send the content to the channel
	go func() {
		defer close(ch)
		ch <- domain.Token{Text: content, Finished: true}
	}()
	
	return ch
}