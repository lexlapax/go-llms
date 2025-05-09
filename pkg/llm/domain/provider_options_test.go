package domain

import (
	"net/http"
	"testing"
	"time"
)

// Since we can't redefine the Apply methods in the test file,
// we'll create an applyOption helper function to simulate applying
// the options to our mock providers.

// Mock provider types to test options
type mockOpenAIProvider struct {
	baseURL    string
	httpClient *http.Client
	timeout    int
	maxRetries int
	retryDelay int
	headers    map[string]string
}

type mockAnthropicProvider struct {
	baseURL    string
	httpClient *http.Client
	timeout    int
	maxRetries int
	retryDelay int
	headers    map[string]string
}

type mockGeminiProvider struct {
	baseURL    string
	httpClient *http.Client
	timeout    int
	maxRetries int
	retryDelay int
	headers    map[string]string
}

// Helper function to apply options to mock providers
func applyBaseURLOption(o *BaseURLOption, provider interface{}) {
	switch p := provider.(type) {
	case *mockOpenAIProvider:
		p.baseURL = o.URL
	case *mockAnthropicProvider:
		p.baseURL = o.URL
	case *mockGeminiProvider:
		p.baseURL = o.URL
	}
}

func applyHTTPClientOption(o *HTTPClientOption, provider interface{}) {
	switch p := provider.(type) {
	case *mockOpenAIProvider:
		p.httpClient = o.Client
	case *mockAnthropicProvider:
		p.httpClient = o.Client
	case *mockGeminiProvider:
		p.httpClient = o.Client
	}
}

func applyTimeoutOption(o *TimeoutOption, provider interface{}) {
	switch p := provider.(type) {
	case *mockOpenAIProvider:
		p.timeout = o.Timeout
	case *mockAnthropicProvider:
		p.timeout = o.Timeout
	case *mockGeminiProvider:
		p.timeout = o.Timeout
	}
}

func applyRetryOption(o *RetryOption, provider interface{}) {
	switch p := provider.(type) {
	case *mockOpenAIProvider:
		p.maxRetries = o.MaxRetries
		p.retryDelay = o.RetryDelay
	case *mockAnthropicProvider:
		p.maxRetries = o.MaxRetries
		p.retryDelay = o.RetryDelay
	case *mockGeminiProvider:
		p.maxRetries = o.MaxRetries
		p.retryDelay = o.RetryDelay
	}
}

func applyHeadersOption(o *HeadersOption, provider interface{}) {
	switch p := provider.(type) {
	case *mockOpenAIProvider:
		p.headers = o.Headers
	case *mockAnthropicProvider:
		p.headers = o.Headers
	case *mockGeminiProvider:
		p.headers = o.Headers
	}
}

// Test that options implement the correct interfaces
func TestOptionInterfaces(t *testing.T) {
	// BaseURLOption
	baseURL := NewBaseURLOption("https://api.example.com")
	if _, ok := interface{}(baseURL).(OpenAIOption); !ok {
		t.Error("BaseURLOption should implement OpenAIOption")
	}
	if _, ok := interface{}(baseURL).(AnthropicOption); !ok {
		t.Error("BaseURLOption should implement AnthropicOption")
	}
	if _, ok := interface{}(baseURL).(GeminiOption); !ok {
		t.Error("BaseURLOption should implement GeminiOption")
	}

	// HTTPClientOption
	httpClient := NewHTTPClientOption(&http.Client{})
	if _, ok := interface{}(httpClient).(OpenAIOption); !ok {
		t.Error("HTTPClientOption should implement OpenAIOption")
	}
	if _, ok := interface{}(httpClient).(AnthropicOption); !ok {
		t.Error("HTTPClientOption should implement AnthropicOption")
	}
	if _, ok := interface{}(httpClient).(GeminiOption); !ok {
		t.Error("HTTPClientOption should implement GeminiOption")
	}

	// TimeoutOption
	timeout := NewTimeoutOption(5000)
	if _, ok := interface{}(timeout).(OpenAIOption); !ok {
		t.Error("TimeoutOption should implement OpenAIOption")
	}
	if _, ok := interface{}(timeout).(AnthropicOption); !ok {
		t.Error("TimeoutOption should implement AnthropicOption")
	}
	if _, ok := interface{}(timeout).(GeminiOption); !ok {
		t.Error("TimeoutOption should implement GeminiOption")
	}

	// RetryOption
	retry := NewRetryOption(3, 1000)
	if _, ok := interface{}(retry).(OpenAIOption); !ok {
		t.Error("RetryOption should implement OpenAIOption")
	}
	if _, ok := interface{}(retry).(AnthropicOption); !ok {
		t.Error("RetryOption should implement AnthropicOption")
	}
	if _, ok := interface{}(retry).(GeminiOption); !ok {
		t.Error("RetryOption should implement GeminiOption")
	}

	// HeadersOption
	headers := NewHeadersOption(map[string]string{"X-API-Key": "test-key"})
	if _, ok := interface{}(headers).(OpenAIOption); !ok {
		t.Error("HeadersOption should implement OpenAIOption")
	}
	if _, ok := interface{}(headers).(AnthropicOption); !ok {
		t.Error("HeadersOption should implement AnthropicOption")
	}
	if _, ok := interface{}(headers).(GeminiOption); !ok {
		t.Error("HeadersOption should implement GeminiOption")
	}
}

// Test applying options to providers
func TestApplyOptions(t *testing.T) {
	t.Run("Apply BaseURLOption", func(t *testing.T) {
		customURL := "https://custom-api.example.com"
		option := NewBaseURLOption(customURL)

		// Test with OpenAI provider
		openAIProvider := &mockOpenAIProvider{}
		applyBaseURLOption(option, openAIProvider)
		if openAIProvider.baseURL != customURL {
			t.Errorf("Expected baseURL to be %s for OpenAI provider, got %s", customURL, openAIProvider.baseURL)
		}

		// Test with Anthropic provider
		anthropicProvider := &mockAnthropicProvider{}
		applyBaseURLOption(option, anthropicProvider)
		if anthropicProvider.baseURL != customURL {
			t.Errorf("Expected baseURL to be %s for Anthropic provider, got %s", customURL, anthropicProvider.baseURL)
		}

		// Test with Gemini provider
		geminiProvider := &mockGeminiProvider{}
		applyBaseURLOption(option, geminiProvider)
		if geminiProvider.baseURL != customURL {
			t.Errorf("Expected baseURL to be %s for Gemini provider, got %s", customURL, geminiProvider.baseURL)
		}
	})

	t.Run("Apply HTTPClientOption", func(t *testing.T) {
		customClient := &http.Client{Timeout: 30 * time.Second}
		option := NewHTTPClientOption(customClient)

		// Test with OpenAI provider
		openAIProvider := &mockOpenAIProvider{}
		applyHTTPClientOption(option, openAIProvider)
		if openAIProvider.httpClient != customClient {
			t.Error("Expected httpClient to be set for OpenAI provider")
		}

		// Test with Anthropic provider
		anthropicProvider := &mockAnthropicProvider{}
		applyHTTPClientOption(option, anthropicProvider)
		if anthropicProvider.httpClient != customClient {
			t.Error("Expected httpClient to be set for Anthropic provider")
		}

		// Test with Gemini provider
		geminiProvider := &mockGeminiProvider{}
		applyHTTPClientOption(option, geminiProvider)
		if geminiProvider.httpClient != customClient {
			t.Error("Expected httpClient to be set for Gemini provider")
		}
	})

	t.Run("Apply TimeoutOption", func(t *testing.T) {
		timeoutMS := 5000
		option := NewTimeoutOption(timeoutMS)

		// Test with OpenAI provider
		openAIProvider := &mockOpenAIProvider{}
		applyTimeoutOption(option, openAIProvider)
		if openAIProvider.timeout != timeoutMS {
			t.Errorf("Expected timeout to be %d for OpenAI provider, got %d", timeoutMS, openAIProvider.timeout)
		}

		// Test with Anthropic provider
		anthropicProvider := &mockAnthropicProvider{}
		applyTimeoutOption(option, anthropicProvider)
		if anthropicProvider.timeout != timeoutMS {
			t.Errorf("Expected timeout to be %d for Anthropic provider, got %d", timeoutMS, anthropicProvider.timeout)
		}

		// Test with Gemini provider
		geminiProvider := &mockGeminiProvider{}
		applyTimeoutOption(option, geminiProvider)
		if geminiProvider.timeout != timeoutMS {
			t.Errorf("Expected timeout to be %d for Gemini provider, got %d", timeoutMS, geminiProvider.timeout)
		}
	})

	t.Run("Apply RetryOption", func(t *testing.T) {
		maxRetries := 3
		retryDelay := 1000
		option := NewRetryOption(maxRetries, retryDelay)

		// Test with OpenAI provider
		openAIProvider := &mockOpenAIProvider{}
		applyRetryOption(option, openAIProvider)
		if openAIProvider.maxRetries != maxRetries {
			t.Errorf("Expected maxRetries to be %d for OpenAI provider, got %d", maxRetries, openAIProvider.maxRetries)
		}
		if openAIProvider.retryDelay != retryDelay {
			t.Errorf("Expected retryDelay to be %d for OpenAI provider, got %d", retryDelay, openAIProvider.retryDelay)
		}

		// Test with Anthropic provider
		anthropicProvider := &mockAnthropicProvider{}
		applyRetryOption(option, anthropicProvider)
		if anthropicProvider.maxRetries != maxRetries {
			t.Errorf("Expected maxRetries to be %d for Anthropic provider, got %d", maxRetries, anthropicProvider.maxRetries)
		}
		if anthropicProvider.retryDelay != retryDelay {
			t.Errorf("Expected retryDelay to be %d for Anthropic provider, got %d", retryDelay, anthropicProvider.retryDelay)
		}

		// Test with Gemini provider
		geminiProvider := &mockGeminiProvider{}
		applyRetryOption(option, geminiProvider)
		if geminiProvider.maxRetries != maxRetries {
			t.Errorf("Expected maxRetries to be %d for Gemini provider, got %d", maxRetries, geminiProvider.maxRetries)
		}
		if geminiProvider.retryDelay != retryDelay {
			t.Errorf("Expected retryDelay to be %d for Gemini provider, got %d", retryDelay, geminiProvider.retryDelay)
		}
	})

	t.Run("Apply HeadersOption", func(t *testing.T) {
		headers := map[string]string{
			"X-API-Key":  "test-key",
			"User-Agent": "go-llms-test",
		}
		option := NewHeadersOption(headers)

		// Test with OpenAI provider
		openAIProvider := &mockOpenAIProvider{}
		applyHeadersOption(option, openAIProvider)
		if len(openAIProvider.headers) != len(headers) {
			t.Errorf("Expected %d headers for OpenAI provider, got %d", len(headers), len(openAIProvider.headers))
		}
		for k, v := range headers {
			if openAIProvider.headers[k] != v {
				t.Errorf("Expected header %s to be %s for OpenAI provider, got %s", k, v, openAIProvider.headers[k])
			}
		}

		// Test with Anthropic provider
		anthropicProvider := &mockAnthropicProvider{}
		applyHeadersOption(option, anthropicProvider)
		if len(anthropicProvider.headers) != len(headers) {
			t.Errorf("Expected %d headers for Anthropic provider, got %d", len(headers), len(anthropicProvider.headers))
		}
		for k, v := range headers {
			if anthropicProvider.headers[k] != v {
				t.Errorf("Expected header %s to be %s for Anthropic provider, got %s", k, v, anthropicProvider.headers[k])
			}
		}

		// Test with Gemini provider
		geminiProvider := &mockGeminiProvider{}
		applyHeadersOption(option, geminiProvider)
		if len(geminiProvider.headers) != len(headers) {
			t.Errorf("Expected %d headers for Gemini provider, got %d", len(headers), len(geminiProvider.headers))
		}
		for k, v := range headers {
			if geminiProvider.headers[k] != v {
				t.Errorf("Expected header %s to be %s for Gemini provider, got %s", k, v, geminiProvider.headers[k])
			}
		}
	})
}

// Test applying multiple options to a provider
func TestApplyMultipleOptions(t *testing.T) {
	// Create options
	customURL := "https://custom-api.example.com"
	baseURLOption := NewBaseURLOption(customURL)

	customClient := &http.Client{Timeout: 30 * time.Second}
	clientOption := NewHTTPClientOption(customClient)

	timeoutMS := 5000
	timeoutOption := NewTimeoutOption(timeoutMS)

	maxRetries := 3
	retryDelay := 1000
	retryOption := NewRetryOption(maxRetries, retryDelay)

	headers := map[string]string{
		"X-API-Key":  "test-key",
		"User-Agent": "go-llms-test",
	}
	headersOption := NewHeadersOption(headers)

	// Create provider
	openAIProvider := &mockOpenAIProvider{}

	// Apply all options
	applyBaseURLOption(baseURLOption, openAIProvider)
	applyHTTPClientOption(clientOption, openAIProvider)
	applyTimeoutOption(timeoutOption, openAIProvider)
	applyRetryOption(retryOption, openAIProvider)
	applyHeadersOption(headersOption, openAIProvider)

	// Verify all options were applied
	if openAIProvider.baseURL != customURL {
		t.Errorf("Expected baseURL to be %s, got %s", customURL, openAIProvider.baseURL)
	}
	if openAIProvider.httpClient != customClient {
		t.Error("Expected httpClient to be set")
	}
	if openAIProvider.timeout != timeoutMS {
		t.Errorf("Expected timeout to be %d, got %d", timeoutMS, openAIProvider.timeout)
	}
	if openAIProvider.maxRetries != maxRetries {
		t.Errorf("Expected maxRetries to be %d, got %d", maxRetries, openAIProvider.maxRetries)
	}
	if openAIProvider.retryDelay != retryDelay {
		t.Errorf("Expected retryDelay to be %d, got %d", retryDelay, openAIProvider.retryDelay)
	}
	if len(openAIProvider.headers) != len(headers) {
		t.Errorf("Expected %d headers, got %d", len(headers), len(openAIProvider.headers))
	}
	for k, v := range headers {
		if openAIProvider.headers[k] != v {
			t.Errorf("Expected header %s to be %s, got %s", k, v, openAIProvider.headers[k])
		}
	}
}
