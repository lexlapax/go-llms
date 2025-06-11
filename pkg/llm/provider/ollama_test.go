package provider

import (
	"net/http"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewOllamaProvider(t *testing.T) {
	t.Run("DefaultConfiguration", func(t *testing.T) {
		provider := NewOllamaProvider("llama3.2:3b")

		assert.NotNil(t, provider)
		assert.Equal(t, "ollama", provider.apiKey)
		assert.Equal(t, "llama3.2:3b", provider.model)
		assert.Equal(t, defaultOllamaHost, provider.baseURL)
		assert.NotNil(t, provider.httpClient)
		assert.Equal(t, defaultOllamaTimeout, provider.httpClient.Timeout)
	})

	t.Run("WithCustomHost", func(t *testing.T) {
		customHost := "http://192.168.1.100:11434"
		provider := NewOllamaProvider("llama3.2:3b", WithOllamaHost(customHost))

		assert.NotNil(t, provider)
		assert.Equal(t, customHost, provider.baseURL)
	})

	t.Run("WithCustomTimeout", func(t *testing.T) {
		customTimeout := 5 * time.Minute
		provider := NewOllamaProvider("llama3.2:3b", WithOllamaTimeout(customTimeout))

		assert.NotNil(t, provider)
		assert.Equal(t, customTimeout, provider.httpClient.Timeout)
	})

	t.Run("WithMultipleOptions", func(t *testing.T) {
		customHost := "http://custom-ollama:11434"
		customTimeout := 3 * time.Minute

		provider := NewOllamaProvider("mistral:7b",
			WithOllamaHost(customHost),
			WithOllamaTimeout(customTimeout),
			domain.NewHeadersOption(map[string]string{"X-Custom": "header"}),
		)

		assert.NotNil(t, provider)
		assert.Equal(t, "mistral:7b", provider.model)
		assert.Equal(t, customHost, provider.baseURL)
		assert.Equal(t, customTimeout, provider.httpClient.Timeout)
	})

	t.Run("WithStandardOpenAIOptions", func(t *testing.T) {
		// Test that standard OpenAI options work with Ollama provider
		customClient := &http.Client{Timeout: 90 * time.Second}
		provider := NewOllamaProvider("codellama:13b",
			domain.NewHTTPClientOption(customClient),
		)

		assert.NotNil(t, provider)
		assert.Equal(t, customClient, provider.httpClient)
	})

	t.Run("OverrideDefaultsWithBaseURL", func(t *testing.T) {
		// Test that domain.NewBaseURLOption overrides the default Ollama host
		customURL := "http://ollama.example.com"
		provider := NewOllamaProvider("llama3.2:3b",
			domain.NewBaseURLOption(customURL),
		)

		assert.NotNil(t, provider)
		assert.Equal(t, customURL, provider.baseURL)
	})
}

func TestOllamaOptions(t *testing.T) {
	t.Run("WithOllamaHost", func(t *testing.T) {
		opt := WithOllamaHost("http://test:11434")
		assert.Equal(t, "openai", opt.ProviderType())

		// Test ApplyToOpenAI
		provider := &OpenAIProvider{baseURL: "default"}
		opt.ApplyToOpenAI(provider)
		assert.Equal(t, "http://test:11434", provider.baseURL)

		// Test ApplyToOllama
		provider2 := &OpenAIProvider{baseURL: "default"}
		opt.ApplyToOllama(provider2)
		assert.Equal(t, "http://test:11434", provider2.baseURL)
	})

	t.Run("WithOllamaTimeout", func(t *testing.T) {
		timeout := 5 * time.Minute
		opt := WithOllamaTimeout(timeout)
		assert.Equal(t, "openai", opt.ProviderType())

		// Test ApplyToOpenAI
		provider := &OpenAIProvider{httpClient: &http.Client{}}
		opt.ApplyToOpenAI(provider)
		assert.Equal(t, timeout, provider.httpClient.Timeout)

		// Test ApplyToOllama
		provider2 := &OpenAIProvider{httpClient: &http.Client{}}
		opt.ApplyToOllama(provider2)
		assert.Equal(t, timeout, provider2.httpClient.Timeout)
	})
}
