package domain

import (
	"net/http"
)

// ProviderOption is the base interface for all provider options
type ProviderOption interface {
	// ProviderType returns the type of provider this option is for
	// Can be "openai", "anthropic", "gemini", or "all"
	ProviderType() string
}

// OpenAIOption is an interface for options specific to the OpenAI provider
type OpenAIOption interface {
	ProviderOption
	// ApplyToOpenAI applies the option to an OpenAI provider
	// The actual OpenAIProvider type will be defined in the provider package
	ApplyToOpenAI(provider interface{})
}

// AnthropicOption is an interface for options specific to the Anthropic provider
type AnthropicOption interface {
	ProviderOption
	// ApplyToAnthropic applies the option to an Anthropic provider
	// The actual AnthropicProvider type will be defined in the provider package
	ApplyToAnthropic(provider interface{})
}

// GeminiOption is an interface for options specific to the Gemini provider
type GeminiOption interface {
	ProviderOption
	// ApplyToGemini applies the option to a Gemini provider
	// The actual GeminiProvider type will be defined in the provider package
	ApplyToGemini(provider interface{})
}

// MockOption is an interface for options specific to the Mock provider
type MockOption interface {
	ProviderOption
	// ApplyToMock applies the option to a Mock provider
	// The actual MockProvider type will be defined in the provider package
	ApplyToMock(provider interface{})
}

// CommonOption is an interface for options that apply to all providers
type CommonOption interface {
	OpenAIOption
	AnthropicOption
	GeminiOption
	MockOption
}

// BaseURLOption sets a custom base URL for the provider API
type BaseURLOption struct {
	URL string
}

// NewBaseURLOption creates a new BaseURLOption
func NewBaseURLOption(url string) *BaseURLOption {
	return &BaseURLOption{URL: url}
}

func (o *BaseURLOption) ProviderType() string { return "all" }

func (o *BaseURLOption) ApplyToOpenAI(provider interface{}) {
	// Use type assertion to check if provider implements SetBaseURL method
	if p, ok := provider.(interface{ SetBaseURL(url string) }); ok {
		p.SetBaseURL(o.URL)
	}
}

func (o *BaseURLOption) ApplyToAnthropic(provider interface{}) {
	if p, ok := provider.(interface{ SetBaseURL(url string) }); ok {
		p.SetBaseURL(o.URL)
	}
}

func (o *BaseURLOption) ApplyToGemini(provider interface{}) {
	if p, ok := provider.(interface{ SetBaseURL(url string) }); ok {
		p.SetBaseURL(o.URL)
	}
}

func (o *BaseURLOption) ApplyToMock(provider interface{}) {
	if p, ok := provider.(interface{ SetBaseURL(url string) }); ok {
		p.SetBaseURL(o.URL)
	}
}

// HTTPClientOption sets a custom HTTP client for the provider
type HTTPClientOption struct {
	Client *http.Client
}

// NewHTTPClientOption creates a new HTTPClientOption
func NewHTTPClientOption(client *http.Client) *HTTPClientOption {
	return &HTTPClientOption{Client: client}
}

func (o *HTTPClientOption) ProviderType() string { return "all" }

func (o *HTTPClientOption) ApplyToOpenAI(provider interface{}) {
	if p, ok := provider.(interface{ SetHTTPClient(client *http.Client) }); ok {
		p.SetHTTPClient(o.Client)
	}
}

func (o *HTTPClientOption) ApplyToAnthropic(provider interface{}) {
	if p, ok := provider.(interface{ SetHTTPClient(client *http.Client) }); ok {
		p.SetHTTPClient(o.Client)
	}
}

func (o *HTTPClientOption) ApplyToGemini(provider interface{}) {
	if p, ok := provider.(interface{ SetHTTPClient(client *http.Client) }); ok {
		p.SetHTTPClient(o.Client)
	}
}

func (o *HTTPClientOption) ApplyToMock(provider interface{}) {
	if p, ok := provider.(interface{ SetHTTPClient(client *http.Client) }); ok {
		p.SetHTTPClient(o.Client)
	}
}

// TimeoutOption sets a timeout for API requests
type TimeoutOption struct {
	Timeout int // timeout in milliseconds
}

// NewTimeoutOption creates a new TimeoutOption
func NewTimeoutOption(timeoutMS int) *TimeoutOption {
	return &TimeoutOption{Timeout: timeoutMS}
}

func (o *TimeoutOption) ProviderType() string { return "all" }

func (o *TimeoutOption) ApplyToOpenAI(provider interface{}) {
	// We'll implement the actual functionality when refactoring the OpenAI provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *TimeoutOption) ApplyToAnthropic(provider interface{}) {
	// We'll implement the actual functionality when refactoring the Anthropic provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *TimeoutOption) ApplyToGemini(provider interface{}) {
	// We'll implement the actual functionality when refactoring the Gemini provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *TimeoutOption) ApplyToMock(provider interface{}) {
	// We'll implement the actual functionality when refactoring the Mock provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

// RetryOption sets retry behavior for API requests
type RetryOption struct {
	MaxRetries int
	RetryDelay int // delay in milliseconds
}

// NewRetryOption creates a new RetryOption
func NewRetryOption(maxRetries, retryDelayMS int) *RetryOption {
	return &RetryOption{
		MaxRetries: maxRetries,
		RetryDelay: retryDelayMS,
	}
}

func (o *RetryOption) ProviderType() string { return "all" }

func (o *RetryOption) ApplyToOpenAI(provider interface{}) {
	// We'll implement the actual functionality when refactoring the OpenAI provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *RetryOption) ApplyToAnthropic(provider interface{}) {
	// We'll implement the actual functionality when refactoring the Anthropic provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *RetryOption) ApplyToGemini(provider interface{}) {
	// We'll implement the actual functionality when refactoring the Gemini provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *RetryOption) ApplyToMock(provider interface{}) {
	// We'll implement the actual functionality when refactoring the Mock provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

// HeadersOption sets custom HTTP headers for API requests
type HeadersOption struct {
	Headers map[string]string
}

// NewHeadersOption creates a new HeadersOption
func NewHeadersOption(headers map[string]string) *HeadersOption {
	return &HeadersOption{Headers: headers}
}

func (o *HeadersOption) ProviderType() string { return "all" }

func (o *HeadersOption) ApplyToOpenAI(provider interface{}) {
	// We'll implement the actual functionality when refactoring the OpenAI provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *HeadersOption) ApplyToAnthropic(provider interface{}) {
	// We'll implement the actual functionality when refactoring the Anthropic provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *HeadersOption) ApplyToGemini(provider interface{}) {
	// We'll implement the actual functionality when refactoring the Gemini provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *HeadersOption) ApplyToMock(provider interface{}) {
	if p, ok := provider.(interface {
		SetHeaders(headers map[string]string)
	}); ok {
		p.SetHeaders(o.Headers)
	}
}

// ModelOption sets the model for the provider
type ModelOption struct {
	Model string
}

// NewModelOption creates a new ModelOption
func NewModelOption(model string) *ModelOption {
	return &ModelOption{Model: model}
}

func (o *ModelOption) ProviderType() string { return "all" }

func (o *ModelOption) ApplyToOpenAI(provider interface{}) {
	// We'll implement the actual functionality when refactoring the OpenAI provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *ModelOption) ApplyToAnthropic(provider interface{}) {
	// We'll implement the actual functionality when refactoring the Anthropic provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *ModelOption) ApplyToGemini(provider interface{}) {
	// We'll implement the actual functionality when refactoring the Gemini provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

func (o *ModelOption) ApplyToMock(provider interface{}) {
	// We'll implement the actual functionality when refactoring the Mock provider
	// For now, we leave this as a stub that will be accessed by reflection in tests
}

//
// Provider-specific options
//

// OpenAI-specific options

// OpenAIOrganizationOption sets the organization for OpenAI API calls
type OpenAIOrganizationOption struct {
	Organization string
}

// NewOpenAIOrganizationOption creates a new OpenAIOrganizationOption
func NewOpenAIOrganizationOption(organization string) *OpenAIOrganizationOption {
	return &OpenAIOrganizationOption{Organization: organization}
}

func (o *OpenAIOrganizationOption) ProviderType() string { return "openai" }

func (o *OpenAIOrganizationOption) ApplyToOpenAI(provider interface{}) {
	if p, ok := provider.(interface{ SetOrganization(org string) }); ok {
		p.SetOrganization(o.Organization)
	}
}

// OpenAILogitBiasOption sets the logit bias for OpenAI API calls
type OpenAILogitBiasOption struct {
	LogitBias map[string]float64
}

// NewOpenAILogitBiasOption creates a new OpenAILogitBiasOption
func NewOpenAILogitBiasOption(logitBias map[string]float64) *OpenAILogitBiasOption {
	return &OpenAILogitBiasOption{LogitBias: logitBias}
}

func (o *OpenAILogitBiasOption) ProviderType() string { return "openai" }

func (o *OpenAILogitBiasOption) ApplyToOpenAI(provider interface{}) {
	if p, ok := provider.(interface {
		SetLogitBias(logitBias map[string]float64)
	}); ok {
		p.SetLogitBias(o.LogitBias)
	}
}

// Anthropic-specific options

// AnthropicSystemPromptOption sets the system prompt for Anthropic API calls
type AnthropicSystemPromptOption struct {
	SystemPrompt string
}

// NewAnthropicSystemPromptOption creates a new AnthropicSystemPromptOption
func NewAnthropicSystemPromptOption(systemPrompt string) *AnthropicSystemPromptOption {
	return &AnthropicSystemPromptOption{SystemPrompt: systemPrompt}
}

func (o *AnthropicSystemPromptOption) ProviderType() string { return "anthropic" }

func (o *AnthropicSystemPromptOption) ApplyToAnthropic(provider interface{}) {
	if p, ok := provider.(interface{ SetSystemPrompt(prompt string) }); ok {
		p.SetSystemPrompt(o.SystemPrompt)
	}
}

// AnthropicMetadataOption sets the metadata for Anthropic API calls
type AnthropicMetadataOption struct {
	Metadata map[string]string
}

// NewAnthropicMetadataOption creates a new AnthropicMetadataOption
func NewAnthropicMetadataOption(metadata map[string]string) *AnthropicMetadataOption {
	return &AnthropicMetadataOption{Metadata: metadata}
}

func (o *AnthropicMetadataOption) ProviderType() string { return "anthropic" }

func (o *AnthropicMetadataOption) ApplyToAnthropic(provider interface{}) {
	if p, ok := provider.(interface {
		SetMetadata(metadata map[string]string)
	}); ok {
		p.SetMetadata(o.Metadata)
	}
}

// Gemini-specific options

// GeminiGenerationConfigOption sets the generation config for Gemini API calls
type GeminiGenerationConfigOption struct {
	Temperature      *float64
	TopP             *float64
	TopK             *int
	MaxOutputTokens  *int
	CandidateCount   *int
	StopSequences    []string
	PresencePenalty  *float64
	FrequencyPenalty *float64
}

// NewGeminiGenerationConfigOption creates a new GeminiGenerationConfigOption
func NewGeminiGenerationConfigOption() *GeminiGenerationConfigOption {
	return &GeminiGenerationConfigOption{}
}

// WithTemperature sets the temperature
func (o *GeminiGenerationConfigOption) WithTemperature(temperature float64) *GeminiGenerationConfigOption {
	o.Temperature = &temperature
	return o
}

// WithTopP sets the top-p value
func (o *GeminiGenerationConfigOption) WithTopP(topP float64) *GeminiGenerationConfigOption {
	o.TopP = &topP
	return o
}

// WithTopK sets the top-k value
func (o *GeminiGenerationConfigOption) WithTopK(topK int) *GeminiGenerationConfigOption {
	o.TopK = &topK
	return o
}

// WithMaxOutputTokens sets the max output tokens
func (o *GeminiGenerationConfigOption) WithMaxOutputTokens(maxOutputTokens int) *GeminiGenerationConfigOption {
	o.MaxOutputTokens = &maxOutputTokens
	return o
}

// WithCandidateCount sets the candidate count
func (o *GeminiGenerationConfigOption) WithCandidateCount(candidateCount int) *GeminiGenerationConfigOption {
	o.CandidateCount = &candidateCount
	return o
}

// WithStopSequences sets the stop sequences
func (o *GeminiGenerationConfigOption) WithStopSequences(stopSequences []string) *GeminiGenerationConfigOption {
	o.StopSequences = stopSequences
	return o
}

// WithPresencePenalty sets the presence penalty
func (o *GeminiGenerationConfigOption) WithPresencePenalty(presencePenalty float64) *GeminiGenerationConfigOption {
	o.PresencePenalty = &presencePenalty
	return o
}

// WithFrequencyPenalty sets the frequency penalty
func (o *GeminiGenerationConfigOption) WithFrequencyPenalty(frequencyPenalty float64) *GeminiGenerationConfigOption {
	o.FrequencyPenalty = &frequencyPenalty
	return o
}

func (o *GeminiGenerationConfigOption) ProviderType() string { return "gemini" }

func (o *GeminiGenerationConfigOption) ApplyToGemini(provider interface{}) {
	// Set topK if configured
	if o.TopK != nil {
		if p, ok := provider.(interface{ SetTopK(topK int) }); ok {
			p.SetTopK(*o.TopK)
		}
	}

	// Other Gemini-specific generation parameters can be implemented
	// when the GeminiProvider supports them
}

// GeminiSafetySettingsOption sets the safety settings for Gemini API calls
type GeminiSafetySettingsOption struct {
	Settings []map[string]interface{}
}

// NewGeminiSafetySettingsOption creates a new GeminiSafetySettingsOption
func NewGeminiSafetySettingsOption(settings []map[string]interface{}) *GeminiSafetySettingsOption {
	return &GeminiSafetySettingsOption{Settings: settings}
}

func (o *GeminiSafetySettingsOption) ProviderType() string { return "gemini" }

func (o *GeminiSafetySettingsOption) ApplyToGemini(provider interface{}) {
	if p, ok := provider.(interface {
		SetSafetySettings(settings []map[string]interface{})
	}); ok {
		p.SetSafetySettings(o.Settings)
	}
}
