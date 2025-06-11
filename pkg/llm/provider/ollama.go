package provider

// ABOUTME: Ollama provider convenience wrapper for local LLM hosting
// ABOUTME: Uses OpenAI-compatible API with sensible defaults for Ollama

import (
	"net/http"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
)

const (
	defaultOllamaHost    = "http://localhost:11434"
	defaultOllamaKey     = "ollama"
	defaultOllamaTimeout = 120 * time.Second
)

// OllamaOption represents options specific to Ollama
type OllamaOption interface {
	domain.OpenAIOption
	ApplyToOllama(provider *OpenAIProvider)
}

// ollamaHost implements an option for setting Ollama host
type ollamaHost struct {
	host string
}

// WithOllamaHost creates an option to set a custom Ollama host
func WithOllamaHost(host string) OllamaOption {
	return &ollamaHost{host: host}
}

func (o *ollamaHost) ProviderType() string { return "openai" }

func (o *ollamaHost) ApplyToOpenAI(provider interface{}) {
	if p, ok := provider.(interface{ SetBaseURL(url string) }); ok {
		p.SetBaseURL(o.host)
	}
}

func (o *ollamaHost) ApplyToOllama(provider *OpenAIProvider) {
	provider.SetBaseURL(o.host)
}

// ollamaTimeoutOption implements an option for setting Ollama timeout
type ollamaTimeoutOption struct {
	timeout time.Duration
}

// WithOllamaTimeout creates an option to set a custom timeout for Ollama
func WithOllamaTimeout(timeout time.Duration) OllamaOption {
	return &ollamaTimeoutOption{timeout: timeout}
}

func (o *ollamaTimeoutOption) ProviderType() string { return "openai" }

func (o *ollamaTimeoutOption) ApplyToOpenAI(provider interface{}) {
	if p, ok := provider.(interface{ SetHTTPClient(client *http.Client) }); ok {
		p.SetHTTPClient(&http.Client{Timeout: o.timeout})
	}
}

func (o *ollamaTimeoutOption) ApplyToOllama(provider *OpenAIProvider) {
	provider.SetHTTPClient(&http.Client{Timeout: o.timeout})
}

// NewOllamaProvider creates a new provider configured for Ollama
// This is a convenience wrapper around the OpenAI provider with sensible defaults.
//
// By default, it uses:
//   - Host: http://localhost:11434
//   - API Key: "ollama" (ignored by Ollama server)
//   - Timeout: 120 seconds (suitable for local model inference)
//
// You can override these defaults using options:
//   - WithOllamaHost("http://custom-host:11434")
//   - WithOllamaTimeout(5 * time.Minute)
//   - Any standard OpenAI options (domain.WithTemperature, etc.)
//
// Example:
//
//	provider := NewOllamaProvider("llama3.2:3b",
//	    WithOllamaHost("http://192.168.1.100:11434"),
//	    domain.WithTemperature(0.7),
//	)
func NewOllamaProvider(model string, options ...domain.ProviderOption) *OpenAIProvider {
	// Start with default options
	defaultOptions := []domain.ProviderOption{
		domain.NewBaseURLOption(defaultOllamaHost),
		domain.NewHTTPClientOption(&http.Client{Timeout: defaultOllamaTimeout}),
	}

	// Process options to handle Ollama-specific ones
	finalOptions := make([]domain.ProviderOption, 0, len(defaultOptions)+len(options))
	finalOptions = append(finalOptions, defaultOptions...)

	// Override defaults with any provided options
	finalOptions = append(finalOptions, options...)

	// Create the OpenAI provider with Ollama configuration
	return NewOpenAIProvider(defaultOllamaKey, model, finalOptions...)
}

// OllamaModels represents the response from Ollama's /api/tags endpoint
type OllamaModels struct {
	Models []OllamaModel `json:"models"`
}

// OllamaModel represents a single model in Ollama
type OllamaModel struct {
	Name       string    `json:"name"`
	Model      string    `json:"model"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    struct {
		ParentModel       string   `json:"parent_model"`
		Format            string   `json:"format"`
		Family            string   `json:"family"`
		Families          []string `json:"families"`
		ParameterSize     string   `json:"parameter_size"`
		QuantizationLevel string   `json:"quantization_level"`
	} `json:"details"`
}
