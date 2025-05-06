package domain

// Option configures LLM provider behavior
type Option func(*ProviderOptions)

// ProviderOptions stores configuration for LLM providers
type ProviderOptions struct {
	Temperature      float64
	MaxTokens        int
	StopSequences    []string
	TopP             float64
	FrequencyPenalty float64
	PresencePenalty  float64
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