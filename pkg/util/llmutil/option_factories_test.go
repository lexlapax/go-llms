package llmutil

import (
	"fmt"
	"os"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

func TestWithPerformanceOptions(t *testing.T) {
	options := WithPerformanceOptions()

	// Check that we have the expected number of options
	if len(options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(options))
	}

	// Check that we have the expected option types
	hasHTTPClient := false
	hasTimeout := false
	hasRetry := false

	for _, opt := range options {
		switch opt.(type) {
		case *domain.HTTPClientOption:
			hasHTTPClient = true
		case *domain.TimeoutOption:
			hasTimeout = true
		case *domain.RetryOption:
			hasRetry = true
		}
	}

	if !hasHTTPClient {
		t.Errorf("Expected HTTPClientOption but not found")
	}
	if !hasTimeout {
		t.Errorf("Expected TimeoutOption but not found")
	}
	if !hasRetry {
		t.Errorf("Expected RetryOption but not found")
	}
}

func TestWithReliabilityOptions(t *testing.T) {
	options := WithReliabilityOptions()

	// Check that we have the expected number of options
	if len(options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(options))
	}

	// Check that we have the expected option types
	hasHTTPClient := false
	hasTimeout := false
	hasRetry := false

	for _, opt := range options {
		switch opt.(type) {
		case *domain.HTTPClientOption:
			hasHTTPClient = true
		case *domain.TimeoutOption:
			hasTimeout = true
		case *domain.RetryOption:
			hasRetry = true
		}
	}

	if !hasHTTPClient {
		t.Errorf("Expected HTTPClientOption but not found")
	}
	if !hasTimeout {
		t.Errorf("Expected TimeoutOption but not found")
	}
	if !hasRetry {
		t.Errorf("Expected RetryOption but not found")
	}
}

func TestWithOpenAIDefaultOptions(t *testing.T) {
	// Test with organization ID
	options := WithOpenAIDefaultOptions("test-org")

	// Check that we have the expected types of options
	hasOrganization := false
	hasHeaders := false

	for _, opt := range options {
		switch opt.(type) {
		case *domain.OpenAIOrganizationOption:
			hasOrganization = true
		case *domain.HeadersOption:
			hasHeaders = true
		}
	}

	if !hasOrganization {
		t.Errorf("Expected OpenAIOrganizationOption but not found")
	}
	if !hasHeaders {
		t.Errorf("Expected HeadersOption but not found")
	}

	// Test without organization ID
	options = WithOpenAIDefaultOptions("")

	// Check that we don't have the organization option
	hasOrganization = false
	for _, opt := range options {
		if _, ok := opt.(*domain.OpenAIOrganizationOption); ok {
			hasOrganization = true
			break
		}
	}

	if hasOrganization {
		t.Errorf("Did not expect OpenAIOrganizationOption but found it")
	}
}

func TestWithAnthropicDefaultOptions(t *testing.T) {
	// Test with custom system prompt
	options := WithAnthropicDefaultOptions("Custom system prompt")

	// Check that we have the expected types of options
	hasSystemPrompt := false
	hasHeaders := false

	for _, opt := range options {
		switch opt.(type) {
		case *domain.AnthropicSystemPromptOption:
			hasSystemPrompt = true
		case *domain.HeadersOption:
			hasHeaders = true
		}
	}

	if !hasSystemPrompt {
		t.Errorf("Expected AnthropicSystemPromptOption but not found")
	}
	if !hasHeaders {
		t.Errorf("Expected HeadersOption but not found")
	}

	// Test with default system prompt
	options = WithAnthropicDefaultOptions("")

	// Check that we still have the system prompt option
	hasSystemPrompt = false
	for _, opt := range options {
		if _, ok := opt.(*domain.AnthropicSystemPromptOption); ok {
			hasSystemPrompt = true
			break
		}
	}

	if !hasSystemPrompt {
		t.Errorf("Expected AnthropicSystemPromptOption but not found")
	}
}

func TestWithGeminiDefaultOptions(t *testing.T) {
	options := WithGeminiDefaultOptions()

	// Check that we have the expected types of options
	hasGenerationConfig := false
	hasSafetySettings := false

	for _, opt := range options {
		switch opt.(type) {
		case *domain.GeminiGenerationConfigOption:
			hasGenerationConfig = true
		case *domain.GeminiSafetySettingsOption:
			hasSafetySettings = true
		}
	}

	if !hasGenerationConfig {
		t.Errorf("Expected GeminiGenerationConfigOption but not found")
	}
	if !hasSafetySettings {
		t.Errorf("Expected GeminiSafetySettingsOption but not found")
	}
}

func TestWithStreamingOptions(t *testing.T) {
	options := WithStreamingOptions()

	// Check that we have the expected types of options
	hasHTTPClient := false
	hasTimeout := false
	hasHeaders := false

	for _, opt := range options {
		switch opt.(type) {
		case *domain.HTTPClientOption:
			hasHTTPClient = true
		case *domain.TimeoutOption:
			hasTimeout = true
		case *domain.HeadersOption:
			hasHeaders = true
		}
	}

	if !hasHTTPClient {
		t.Errorf("Expected HTTPClientOption but not found")
	}
	if !hasTimeout {
		t.Errorf("Expected TimeoutOption but not found")
	}
	if !hasHeaders {
		t.Errorf("Expected HeadersOption but not found")
	}
}

func TestWithProxyOptions(t *testing.T) {
	// Test with base URL and API key
	options := WithProxyOptions("https://example.com", "test-key")

	// Check that we have the expected types of options
	hasBaseURL := false
	hasHeaders := false

	for _, opt := range options {
		switch opt.(type) {
		case *domain.BaseURLOption:
			hasBaseURL = true
		case *domain.HeadersOption:
			hasHeaders = true
		}
	}

	if !hasBaseURL {
		t.Errorf("Expected BaseURLOption but not found")
	}
	if !hasHeaders {
		t.Errorf("Expected HeadersOption but not found")
	}

	// Test without base URL
	options = WithProxyOptions("", "test-key")

	// Check that we don't have the base URL option
	hasBaseURL = false
	for _, opt := range options {
		if _, ok := opt.(*domain.BaseURLOption); ok {
			hasBaseURL = true
			break
		}
	}

	if hasBaseURL {
		t.Errorf("Did not expect BaseURLOption but found it")
	}

	// Test without API key
	options = WithProxyOptions("https://example.com", "")

	// Check that we don't have the headers option
	hasHeaders = false
	for _, opt := range options {
		if _, ok := opt.(*domain.HeadersOption); ok {
			hasHeaders = true
			break
		}
	}

	if hasHeaders {
		t.Errorf("Did not expect HeadersOption but found it")
	}
}

func TestWithOpenAIStreamingOptions(t *testing.T) {
	options := WithOpenAIStreamingOptions("test-org")

	// Check that we have the expected types of options
	hasOrganization := false
	hasHTTPClient := false
	hasTimeout := false
	hasHeaders := false

	for _, opt := range options {
		switch opt.(type) {
		case *domain.OpenAIOrganizationOption:
			hasOrganization = true
		case *domain.HTTPClientOption:
			hasHTTPClient = true
		case *domain.TimeoutOption:
			hasTimeout = true
		case *domain.HeadersOption:
			hasHeaders = true
		}
	}

	if !hasOrganization {
		t.Errorf("Expected OpenAIOrganizationOption but not found")
	}
	if !hasHTTPClient {
		t.Errorf("Expected HTTPClientOption but not found")
	}
	if !hasTimeout {
		t.Errorf("Expected TimeoutOption but not found")
	}
	if !hasHeaders {
		t.Errorf("Expected HeadersOption but not found")
	}
}

func TestWithAnthropicStreamingOptions(t *testing.T) {
	options := WithAnthropicStreamingOptions("Custom system prompt")

	// Check that we have the expected types of options
	hasSystemPrompt := false
	hasHTTPClient := false
	hasTimeout := false
	hasHeaders := false

	for _, opt := range options {
		switch opt.(type) {
		case *domain.AnthropicSystemPromptOption:
			hasSystemPrompt = true
		case *domain.HTTPClientOption:
			hasHTTPClient = true
		case *domain.TimeoutOption:
			hasTimeout = true
		case *domain.HeadersOption:
			hasHeaders = true
		}
	}

	if !hasSystemPrompt {
		t.Errorf("Expected AnthropicSystemPromptOption but not found")
	}
	if !hasHTTPClient {
		t.Errorf("Expected HTTPClientOption but not found")
	}
	if !hasTimeout {
		t.Errorf("Expected TimeoutOption but not found")
	}
	if !hasHeaders {
		t.Errorf("Expected HeadersOption but not found")
	}
}

func TestCreateOptionFactoryFromEnv(t *testing.T) {
	// Save original environment variables
	origOpenAIOrg := os.Getenv(EnvOpenAIOrganization)
	origAnthropicSystemPrompt := os.Getenv(EnvAnthropicSystemPrompt)
	origHTTPTimeout := os.Getenv(EnvHTTPTimeout)
	origOpenAIUseCase := os.Getenv(EnvOpenAIUseCase)
	origAnthropicUseCase := os.Getenv(EnvAnthropicUseCase)
	origGeminiUseCase := os.Getenv(EnvGeminiUseCase)

	// Clean up environment after test
	defer func() {
		os.Setenv(EnvOpenAIOrganization, origOpenAIOrg)
		os.Setenv(EnvAnthropicSystemPrompt, origAnthropicSystemPrompt)
		os.Setenv(EnvHTTPTimeout, origHTTPTimeout)
		os.Setenv(EnvOpenAIUseCase, origOpenAIUseCase)
		os.Setenv(EnvAnthropicUseCase, origAnthropicUseCase)
		os.Setenv(EnvGeminiUseCase, origGeminiUseCase)
	}()

	tests := []struct {
		name          string
		provider      string
		useCase       string
		envVars       map[string]string
		expectedTypes []string
		minOptions    int
	}{
		{
			name:     "OpenAI Default",
			provider: "openai",
			useCase:  "default",
			envVars: map[string]string{
				EnvOpenAIOrganization: "test-org",
			},
			expectedTypes: []string{"*domain.OpenAIOrganizationOption", "*domain.HeadersOption"},
			minOptions:    2,
		},
		{
			name:     "OpenAI Streaming",
			provider: "openai",
			useCase:  "streaming",
			envVars: map[string]string{
				EnvOpenAIOrganization: "test-org",
			},
			expectedTypes: []string{
				"*domain.OpenAIOrganizationOption",
				"*domain.HeadersOption",
				"*domain.HTTPClientOption",
				"*domain.TimeoutOption",
			},
			minOptions: 4,
		},
		{
			name:     "Anthropic With Env Override",
			provider: "anthropic",
			useCase:  "default",
			envVars: map[string]string{
				EnvAnthropicSystemPrompt: "Custom env prompt",
				EnvHTTPTimeout:           "20",
			},
			expectedTypes: []string{"*domain.AnthropicSystemPromptOption", "*domain.TimeoutOption"},
			minOptions:    2,
		},
		{
			name:     "Gemini Performance",
			provider: "gemini",
			useCase:  "performance",
			envVars:  map[string]string{},
			expectedTypes: []string{
				"*domain.HTTPClientOption",
				"*domain.TimeoutOption",
				"*domain.RetryOption",
				"*domain.GeminiGenerationConfigOption",
				"*domain.GeminiSafetySettingsOption",
			},
			minOptions: 5,
		},
		{
			name:     "Unknown Provider",
			provider: "unknown",
			useCase:  "reliability",
			envVars:  map[string]string{},
			expectedTypes: []string{
				"*domain.HTTPClientOption",
				"*domain.TimeoutOption",
				"*domain.RetryOption",
			},
			minOptions: 3,
		},
		{
			name:     "Use Case from Environment",
			provider: "openai",
			useCase:  "", // No use case in parameter
			envVars: map[string]string{
				EnvOpenAIUseCase:      "streaming", // Use case from environment
				EnvOpenAIOrganization: "test-org",
			},
			expectedTypes: []string{
				"*domain.OpenAIOrganizationOption",
				"*domain.HeadersOption",
				"*domain.HTTPClientOption",
				"*domain.TimeoutOption",
			},
			minOptions: 4,
		},
		{
			name:     "Use Case from Parameter Overrides Environment",
			provider: "anthropic",
			useCase:  "performance", // Use case in parameter takes precedence
			envVars: map[string]string{
				EnvAnthropicUseCase:      "streaming", // This should be ignored
				EnvAnthropicSystemPrompt: "Custom env prompt",
			},
			expectedTypes: []string{
				"*domain.AnthropicSystemPromptOption",
				"*domain.HTTPClientOption",
				"*domain.TimeoutOption",
				"*domain.RetryOption",
			},
			minOptions: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables first
			varsToUnset := []string{
				EnvOpenAIOrganization,
				EnvAnthropicSystemPrompt,
				EnvHTTPTimeout,
				EnvOpenAIUseCase,
				EnvAnthropicUseCase,
				EnvGeminiUseCase,
			}
			for _, v := range varsToUnset {
				os.Unsetenv(v)
			}

			// Set environment variables for test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Call function
			options := CreateOptionFactoryFromEnv(tt.provider, tt.useCase)

			// Check minimum number of options
			if len(options) < tt.minOptions {
				t.Errorf("Expected at least %d options, got %d", tt.minOptions, len(options))
			}

			// Check that expected option types are present
			foundTypes := make(map[string]bool)
			for _, opt := range options {
				typeName := fmt.Sprintf("%T", opt)
				foundTypes[typeName] = true
			}

			for _, expectedType := range tt.expectedTypes {
				if !foundTypes[expectedType] {
					t.Errorf("Expected option type %s but not found", expectedType)
				}
			}
		})
	}
}
