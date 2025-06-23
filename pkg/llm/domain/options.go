package domain

// ABOUTME: Common LLM generation options for controlling output behavior
// ABOUTME: Includes settings for temperature, tokens, response format, and tools

// Option represents a functional option for configuring LLM provider behavior.
// Options are applied in order to customize generation parameters.
type Option func(*ProviderOptions)

// ProviderOptions contains configuration parameters for LLM generation requests.
// These options control various aspects of text generation including randomness,
// length limits, and sampling behavior.
type ProviderOptions struct {
	Temperature      float64
	MaxTokens        int
	StopSequences    []string
	TopP             float64
	TopK             int
	FrequencyPenalty float64
	PresencePenalty  float64
	Model            string
}

// DefaultOptions returns the default provider options
func DefaultOptions() *ProviderOptions {
	return &ProviderOptions{
		Temperature:      0.7,
		MaxTokens:        1024,
		StopSequences:    []string{},
		TopP:             1.0,
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
	}
}

// WithTemperature sets the temperature for generation
func WithTemperature(temp float64) Option {
	return func(o *ProviderOptions) {
		o.Temperature = temp
	}
}

// WithMaxTokens sets the maximum number of tokens to generate
func WithMaxTokens(tokens int) Option {
	return func(o *ProviderOptions) {
		o.MaxTokens = tokens
	}
}

// WithStopSequences sets sequences that stop generation
func WithStopSequences(sequences []string) Option {
	return func(o *ProviderOptions) {
		o.StopSequences = sequences
	}
}

// WithTopP sets the nucleus sampling probability
func WithTopP(topP float64) Option {
	return func(o *ProviderOptions) {
		o.TopP = topP
	}
}

// WithTopK sets the top-k value for generation
func WithTopK(topK int) Option {
	return func(o *ProviderOptions) {
		o.TopK = topK
	}
}

// WithFrequencyPenalty sets the frequency penalty
func WithFrequencyPenalty(penalty float64) Option {
	return func(o *ProviderOptions) {
		o.FrequencyPenalty = penalty
	}
}

// WithPresencePenalty sets the presence penalty
func WithPresencePenalty(penalty float64) Option {
	return func(o *ProviderOptions) {
		o.PresencePenalty = penalty
	}
}

// WithModel sets the model to use
func WithModel(model string) Option {
	return func(o *ProviderOptions) {
		o.Model = model
	}
}
