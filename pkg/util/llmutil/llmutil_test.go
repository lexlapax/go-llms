package llmutil

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestCreateProvider(t *testing.T) {
	// Save original environment variables
	origOpenAIKey := os.Getenv("OPENAI_API_KEY")
	origAnthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	origGeminiKey := os.Getenv("GEMINI_API_KEY")

	// Clean up environment after test
	defer func() {
		os.Setenv("OPENAI_API_KEY", origOpenAIKey)
		os.Setenv("ANTHROPIC_API_KEY", origAnthropicKey)
		os.Setenv("GEMINI_API_KEY", origGeminiKey)
	}()

	// Clear all API keys to prevent environment interference
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("ANTHROPIC_API_KEY")
	os.Unsetenv("GEMINI_API_KEY")

	tests := []struct {
		name          string
		config        ModelConfig
		envSetup      map[string]string
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
			name: "Missing API key but available in env",
			config: ModelConfig{
				Provider: "openai",
				Model:    "gpt-4o",
			},
			envSetup: map[string]string{
				"OPENAI_API_KEY": "env-api-key",
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
		{
			name: "OpenAI with provider options",
			config: ModelConfig{
				Provider: "openai",
				Model:    "gpt-4o",
				APIKey:   "test-api-key",
				Options: []domain.ProviderOption{
					domain.NewTimeoutOption(15),
					domain.NewOpenAIOrganizationOption("test-org"),
				},
			},
			expectError: false,
		},
		{
			name: "Anthropic with provider options",
			config: ModelConfig{
				Provider: "anthropic",
				Model:    "claude-3-5-sonnet-latest",
				APIKey:   "test-api-key",
				Options: []domain.ProviderOption{
					domain.NewAnthropicSystemPromptOption("Test system prompt"),
					domain.NewRetryOption(3, 500),
				},
			},
			expectError: false,
		},
		{
			name: "Missing model with fallback from env",
			config: ModelConfig{
				Provider: "openai",
				APIKey:   "test-api-key",
			},
			envSetup: map[string]string{
				"OPENAI_MODEL": "gpt-4",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all API keys for each test to ensure consistent environment
			os.Unsetenv("OPENAI_API_KEY")
			os.Unsetenv("ANTHROPIC_API_KEY")
			os.Unsetenv("GEMINI_API_KEY")

			// Set environment variables for this test
			for k, v := range tt.envSetup {
				os.Setenv(k, v)
			}

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
		name            string
		err             error
		expectRetryable bool
	}{
		{
			name:            "nil error",
			err:             nil,
			expectRetryable: false,
		},
		{
			name:            "network connectivity error",
			err:             domain.ErrNetworkConnectivity,
			expectRetryable: true,
		},
		{
			name:            "rate limit error",
			err:             domain.ErrRateLimitExceeded,
			expectRetryable: true,
		},
		{
			name:            "invalid model parameters error",
			err:             domain.ErrInvalidModelParameters,
			expectRetryable: false,
		},
		{
			name:            "token quota exceeded error",
			err:             domain.ErrTokenQuotaExceeded,
			expectRetryable: false,
		},
		{
			name:            "other error",
			err:             errors.New("some other error"),
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

func TestWithProviderOptions(t *testing.T) {
	// Save original environment variables
	origOpenAIOrg := os.Getenv(EnvOpenAIOrganization)
	origAnthropicSystemPrompt := os.Getenv(EnvAnthropicSystemPrompt)

	// Clean up environment after test
	defer func() {
		os.Setenv(EnvOpenAIOrganization, origOpenAIOrg)
		os.Setenv(EnvAnthropicSystemPrompt, origAnthropicSystemPrompt)
	}()

	// Set environment variables for test
	os.Setenv(EnvOpenAIOrganization, "env-org-id")
	os.Setenv(EnvAnthropicSystemPrompt, "env-system-prompt")

	tests := []struct {
		name            string
		config          ModelConfig
		expectedOptions int
		checkOption     func([]domain.ProviderOption) bool
	}{
		{
			name: "Base URL Option",
			config: ModelConfig{
				Provider: "openai",
				BaseURL:  "https://custom.openai.com",
			},
			expectedOptions: 1,
			checkOption: func(options []domain.ProviderOption) bool {
				for _, opt := range options {
					if baseURLOpt, ok := opt.(*domain.BaseURLOption); ok {
						return baseURLOpt.URL == "https://custom.openai.com"
					}
				}
				return false
			},
		},
		{
			name: "With Use Case",
			config: ModelConfig{
				Provider: "openai",
				UseCase:  "streaming",
			},
			expectedOptions: 3, // At least 3 options from streaming
			checkOption: func(options []domain.ProviderOption) bool {
				hasTimeout := false
				hasHTTPClient := false

				for _, opt := range options {
					switch opt.(type) {
					case *domain.TimeoutOption:
						hasTimeout = true
					case *domain.HTTPClientOption:
						hasHTTPClient = true
					}
				}

				return hasTimeout && hasHTTPClient
			},
		},
		{
			name: "Custom Options Override",
			config: ModelConfig{
				Provider: "anthropic",
				UseCase:  "default",
				Options: []domain.ProviderOption{
					domain.NewAnthropicSystemPromptOption("custom-override-prompt"),
				},
			},
			expectedOptions: 1, // At least the one we added
			checkOption: func(options []domain.ProviderOption) bool {
				for _, opt := range options {
					if _, ok := opt.(*domain.AnthropicSystemPromptOption); ok {
						// Our custom option should be applied, but we can't check its value directly
						// Just verify it's present
						return true
					}
				}
				return false
			},
		},
		{
			name: "Unknown Provider",
			config: ModelConfig{
				Provider: "unknown",
				UseCase:  "performance",
			},
			expectedOptions: 3, // Should still get the performance options
			checkOption: func(options []domain.ProviderOption) bool {
				hasRetry := false

				for _, opt := range options {
					if _, ok := opt.(*domain.RetryOption); ok {
						hasRetry = true
						break
					}
				}

				return hasRetry
			},
		},
		{
			name: "Anthropic with base URL",
			config: ModelConfig{
				Provider: "anthropic",
				Model:    "claude-3-5-sonnet-latest",
				APIKey:   "test-api-key",
				BaseURL:  "https://custom-anthropic.example.com",
			},
			expectedOptions: 1,
			checkOption: func(options []domain.ProviderOption) bool {
				for _, opt := range options {
					if baseURLOpt, ok := opt.(*domain.BaseURLOption); ok {
						return baseURLOpt.URL == "https://custom-anthropic.example.com"
					}
				}
				return false
			},
		},
		{
			name: "OpenAI with base URL and custom options",
			config: ModelConfig{
				Provider: "openai",
				Model:    "gpt-4o",
				APIKey:   "test-api-key",
				BaseURL:  "https://custom-openai.example.com",
				Options: []domain.ProviderOption{
					domain.NewTimeoutOption(15),
					domain.NewOpenAIOrganizationOption("test-org"),
				},
			},
			expectedOptions: 3, // Base URL + 2 custom options
			checkOption: func(options []domain.ProviderOption) bool {
				hasBaseURL := false
				hasTimeout := false
				hasOrg := false

				for _, opt := range options {
					switch o := opt.(type) {
					case *domain.BaseURLOption:
						hasBaseURL = o.URL == "https://custom-openai.example.com"
					case *domain.TimeoutOption:
						hasTimeout = true
					case *domain.OpenAIOrganizationOption:
						hasOrg = true // Just check that organization is present
					}
				}

				return hasBaseURL && hasTimeout && hasOrg
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call function
			options, err := WithProviderOptions(tt.config)

			// Check error
			if err != nil {
				t.Errorf("WithProviderOptions() error = %v", err)
				return
			}

			// Check number of options
			if len(options) < tt.expectedOptions {
				t.Errorf("WithProviderOptions() returned %d options, expected at least %d",
					len(options), tt.expectedOptions)
			}

			// Check specific option contents
			if !tt.checkOption(options) {
				t.Errorf("WithProviderOptions() did not return expected option configuration")
			}
		})
	}
}

func TestProviderFromEnv(t *testing.T) {
	// Define variables for provider, name, and model
	var prov domain.Provider
	var provName, modelName string
	var err error

	// Store original environment variables
	originalOpenAI := os.Getenv("OPENAI_API_KEY")
	originalAnthropic := os.Getenv("ANTHROPIC_API_KEY")
	originalGemini := os.Getenv("GEMINI_API_KEY")
	originalOpenAIBaseURL := os.Getenv("OPENAI_BASE_URL")
	originalAnthropicBaseURL := os.Getenv("ANTHROPIC_BASE_URL")
	originalGeminiBaseURL := os.Getenv("GEMINI_BASE_URL")
	originalOpenAIOrg := os.Getenv("OPENAI_ORGANIZATION")
	originalAnthropicSystemPrompt := os.Getenv("ANTHROPIC_SYSTEM_PROMPT")
	originalHTTPTimeout := os.Getenv("LLM_HTTP_TIMEOUT")
	originalRetryAttempts := os.Getenv("LLM_RETRY_ATTEMPTS")

	// Clean up environment after the test
	defer func() {
		os.Setenv("OPENAI_API_KEY", originalOpenAI)
		os.Setenv("ANTHROPIC_API_KEY", originalAnthropic)
		os.Setenv("GEMINI_API_KEY", originalGemini)
		os.Setenv("OPENAI_BASE_URL", originalOpenAIBaseURL)
		os.Setenv("ANTHROPIC_BASE_URL", originalAnthropicBaseURL)
		os.Setenv("GEMINI_BASE_URL", originalGeminiBaseURL)
		os.Setenv("OPENAI_ORGANIZATION", originalOpenAIOrg)
		os.Setenv("ANTHROPIC_SYSTEM_PROMPT", originalAnthropicSystemPrompt)
		os.Setenv("LLM_HTTP_TIMEOUT", originalHTTPTimeout)
		os.Setenv("LLM_RETRY_ATTEMPTS", originalRetryAttempts)
	}()

	// Clear all environment variables for clean testing
	envVars := []string{
		"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GEMINI_API_KEY",
		"OPENAI_BASE_URL", "ANTHROPIC_BASE_URL", "GEMINI_BASE_URL",
		"OPENAI_ORGANIZATION", "ANTHROPIC_SYSTEM_PROMPT",
		"LLM_HTTP_TIMEOUT", "LLM_RETRY_ATTEMPTS",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	// Test with no API keys (should return mock provider)
	_, provName, _, err = ProviderFromEnv()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if provName != "mock" {
		t.Errorf("Expected 'mock' provider, got: %s", provName)
	}

	// Test with Gemini API key set
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	prov, provName, modelName, err = ProviderFromEnv()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if provName != "gemini" {
		t.Errorf("Expected 'gemini' provider, got: %s", provName)
	}
	if modelName != "gemini-2.0-flash-lite" {
		t.Errorf("Expected 'gemini-2.0-flash-lite' model, got: %s", modelName)
	}

	// Test that the right provider type is returned
	_, ok := prov.(*provider.GeminiProvider)
	if !ok {
		t.Errorf("Expected GeminiProvider, got: %T", prov)
	}

	// Clean up environment for next tests
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	// Test OpenAI provider with custom base URL and organization
	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	os.Setenv("OPENAI_BASE_URL", "https://custom-openai.example.com")
	os.Setenv("OPENAI_ORGANIZATION", "test-org")
	prov, provName, _, err = ProviderFromEnv()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if provName != "openai" {
		t.Errorf("Expected 'openai' provider, got: %s", provName)
	}

	// Test that the provider has the right type
	_, ok = prov.(*provider.OpenAIProvider)
	if !ok {
		t.Errorf("Expected OpenAIProvider, got: %T", prov)
	}

	// Clean up environment for next tests
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	// Test Anthropic provider with custom base URL and system prompt
	os.Setenv("ANTHROPIC_API_KEY", "test-anthropic-key")
	os.Setenv("ANTHROPIC_BASE_URL", "https://custom-anthropic.example.com")
	os.Setenv("ANTHROPIC_SYSTEM_PROMPT", "Test system prompt")
	prov, provName, _, err = ProviderFromEnv()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if provName != "anthropic" {
		t.Errorf("Expected 'anthropic' provider, got: %s", provName)
	}

	// Test that the provider has the right type
	_, ok = prov.(*provider.AnthropicProvider)
	if !ok {
		t.Errorf("Expected AnthropicProvider, got: %T", prov)
	}

	// Clean up environment for next tests
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	// Test provider with common options
	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	os.Setenv("LLM_HTTP_TIMEOUT", "15")
	os.Setenv("LLM_RETRY_ATTEMPTS", "3")
	prov, provName, _, err = ProviderFromEnv()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if provName != "openai" {
		t.Errorf("Expected 'openai' provider, got: %s", provName)
	}

	// Test that the provider has the right type
	_, ok = prov.(*provider.OpenAIProvider)
	if !ok {
		t.Errorf("Expected OpenAIProvider, got: %T", prov)
	}
}

func TestGenerateWithOptions(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	ctx := context.Background()
	prompt := "Test prompt"
	temperature := 0.7
	maxTokens := 100

	result, err := GenerateWithOptions(ctx, mockProvider, prompt, temperature, maxTokens)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == "" {
		t.Errorf("Expected non-empty result but got empty string")
	}
}

func TestConcurrentStreamMessages(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	messageGroups := [][]domain.Message{
		{
			{Role: "user", Content: "Message 1"},
		},
		{
			{Role: "user", Content: "Message 2"},
		},
	}

	streams, errors := ConcurrentStreamMessages(context.Background(), mockProvider, messageGroups)

	if len(streams) != len(messageGroups) {
		t.Errorf("Expected %d streams, got %d", len(messageGroups), len(streams))
	}
	if len(errors) != len(messageGroups) {
		t.Errorf("Expected %d errors, got %d", len(messageGroups), len(errors))
	}

	// Check if all streams are valid
	for i, stream := range streams {
		if stream == nil {
			t.Errorf("Stream %d is nil", i)
		}
		if errors[i] != nil {
			t.Errorf("Unexpected error for stream %d: %v", i, errors[i])
		}

		// Try to read from the stream
		select {
		case token, ok := <-stream:
			if !ok {
				t.Errorf("Stream %d closed unexpectedly", i)
			}
			if token.Text == "" {
				t.Errorf("Expected non-empty token text from stream %d", i)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timed out waiting for token from stream %d", i)
		}
	}
}

func TestProcessTypedWithProvider(t *testing.T) {
	// Skip test for now
	t.Skip("Skipping test that requires mock provider")
}