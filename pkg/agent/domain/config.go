// ABOUTME: Defines configuration types and options for agents
// ABOUTME: Provides flexible configuration management for different agent types

package domain

import (
	"time"
)

// RetryStrategy defines how retries should be handled for failed operations.
// It supports exponential backoff with configurable delays and retry limits.
// Custom retry logic can be provided via the OnRetry callback.
type RetryStrategy struct {
	MaxAttempts     int                          `json:"max_attempts"`
	InitialDelay    time.Duration                `json:"initial_delay"`
	MaxDelay        time.Duration                `json:"max_delay"`
	Multiplier      float64                      `json:"multiplier"`
	RetryableErrors []string                     `json:"retryable_errors,omitempty"`
	OnRetry         func(attempt int, err error) `json:"-"`
}

// DefaultRetryStrategy returns a default retry strategy with sensible defaults.
// Uses 3 attempts with exponential backoff starting at 1 second.
func DefaultRetryStrategy() RetryStrategy {
	return RetryStrategy{
		MaxAttempts:  3,
		InitialDelay: time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// MergeStrategy defines how to merge multiple states from parallel agent executions.
// Different strategies support different use cases like taking the last result or merging all.
type MergeStrategy string

const (
	// MergeStrategyLast takes the last state, ignoring others
	MergeStrategyLast MergeStrategy = "last"

	// MergeStrategyMergeAll merges all states sequentially
	MergeStrategyMergeAll MergeStrategy = "merge_all"

	// MergeStrategyUnion creates a union of all values
	MergeStrategyUnion MergeStrategy = "union"

	// MergeStrategyCustom uses a custom merge function
	MergeStrategyCustom MergeStrategy = "custom"
)

// ParallelConfig holds configuration specific to parallel agent execution.
// Controls concurrency limits, completion behavior, and state merging strategies.
type ParallelConfig struct {
	MaxConcurrency int           `json:"max_concurrency,omitempty"`
	WaitForAll     bool          `json:"wait_for_all"`
	MergeStrategy  MergeStrategy `json:"merge_strategy"`
	FailFast       bool          `json:"fail_fast"`
}

// DefaultParallelConfig returns default configuration for parallel agents.
// Waits for all agents to complete and merges all states by default.
func DefaultParallelConfig() ParallelConfig {
	return ParallelConfig{
		MaxConcurrency: 0, // 0 means unlimited
		WaitForAll:     true,
		MergeStrategy:  MergeStrategyMergeAll,
		FailFast:       false,
	}
}

// SequentialConfig holds configuration specific to sequential agent execution.
// Controls error handling and state passing between sequential steps.
type SequentialConfig struct {
	StopOnError     bool `json:"stop_on_error"`
	PassState       bool `json:"pass_state"`
	ContinueOnError bool `json:"continue_on_error"`
}

// DefaultSequentialConfig returns default configuration for sequential agents.
// Stops on first error and passes state between steps by default.
func DefaultSequentialConfig() SequentialConfig {
	return SequentialConfig{
		StopOnError:     true,
		PassState:       true,
		ContinueOnError: false,
	}
}

// ConditionalConfig holds configuration for conditional agent execution.
// Supports branching logic with optional default branches.
type ConditionalConfig struct {
	DefaultBranch string `json:"default_branch,omitempty"`
	Exhaustive    bool   `json:"exhaustive"`
}

// LoopConfig holds configuration for loop agent execution.
// Prevents infinite loops with iteration limits and timeouts.
type LoopConfig struct {
	MaxIterations int           `json:"max_iterations"`
	Timeout       time.Duration `json:"timeout"`
}

// DefaultLoopConfig returns default configuration for loop agents.
// Limits to 100 iterations with a 5-minute timeout by default.
func DefaultLoopConfig() LoopConfig {
	return LoopConfig{
		MaxIterations: 100,
		Timeout:       5 * time.Minute,
	}
}

// LLMConfig holds configuration specific to LLM agent interactions.
// Includes model parameters, prompts, tools, and response formatting options.
// Supports both streaming and cached responses for different use cases.
type LLMConfig struct {
	Model            string                 `json:"model,omitempty"`
	Temperature      float64                `json:"temperature,omitempty"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	TopP             float64                `json:"top_p,omitempty"`
	SystemPrompt     string                 `json:"system_prompt,omitempty"`
	ResponseFormat   string                 `json:"response_format,omitempty"`
	Tools            []string               `json:"tools,omitempty"`
	ToolChoice       string                 `json:"tool_choice,omitempty"`
	StreamResponses  bool                   `json:"stream_responses"`
	CacheResponses   bool                   `json:"cache_responses"`
	CustomParameters map[string]interface{} `json:"custom_parameters,omitempty"`
}

// DefaultLLMConfig returns default configuration for LLM agents.
// Uses moderate temperature (0.7) without streaming or caching by default.
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Temperature:     0.7,
		MaxTokens:       0, // 0 means use model default
		TopP:            1.0,
		StreamResponses: false,
		CacheResponses:  false,
	}
}

// LoggingConfig defines logging configuration for agent execution.
// Controls log levels, state inclusion, and sensitive data filtering.
type LoggingConfig struct {
	Level         string   `json:"level"` // "debug", "info", "warn", "error"
	IncludeState  bool     `json:"include_state"`
	IncludeEvents bool     `json:"include_events"`
	SensitiveKeys []string `json:"sensitive_keys,omitempty"`
	MaxStateSize  int      `json:"max_state_size,omitempty"`
}

// DefaultLoggingConfig returns default logging configuration.
// Uses info level with event logging and sensitive key filtering.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:         "info",
		IncludeState:  false,
		IncludeEvents: true,
		MaxStateSize:  1024, // 1KB limit for state in logs
		SensitiveKeys: []string{"password", "token", "secret", "key", "api_key"},
	}
}

// ObservabilityConfig defines metrics and tracing configuration.
// Supports custom labels and sampling for performance monitoring.
type ObservabilityConfig struct {
	MetricsEnabled bool              `json:"metrics_enabled"`
	TracingEnabled bool              `json:"tracing_enabled"`
	CustomLabels   map[string]string `json:"custom_labels,omitempty"`
	SampleRate     float64           `json:"sample_rate"`
}

// DefaultObservabilityConfig returns default observability configuration.
// Enables metrics with full sampling, tracing disabled by default.
func DefaultObservabilityConfig() ObservabilityConfig {
	return ObservabilityConfig{
		MetricsEnabled: true,
		TracingEnabled: false,
		SampleRate:     1.0,
	}
}

// SecurityConfig defines security constraints for agent execution.
// Limits resource usage and controls tool access for safe operation.
// Supports sandboxed execution for untrusted workloads.
type SecurityConfig struct {
	MaxStateSize     int64         `json:"max_state_size"`    // Maximum state size in bytes
	MaxArtifactSize  int64         `json:"max_artifact_size"` // Maximum artifact size in bytes
	MaxExecutionTime time.Duration `json:"max_execution_time"`
	AllowedTools     []string      `json:"allowed_tools,omitempty"`
	BlockedTools     []string      `json:"blocked_tools,omitempty"`
	SandboxExecution bool          `json:"sandbox_execution"`
}

// DefaultSecurityConfig returns default security configuration.
// Sets reasonable limits: 10MB states, 100MB artifacts, 30-minute execution.
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		MaxStateSize:     10 * 1024 * 1024,  // 10MB
		MaxArtifactSize:  100 * 1024 * 1024, // 100MB
		MaxExecutionTime: 30 * time.Minute,
		SandboxExecution: false,
	}
}

// CacheConfig defines caching configuration for agent responses.
// Supports multiple cache strategies (LRU, LFU, TTL) with size limits.
type CacheConfig struct {
	Enabled       bool          `json:"enabled"`
	TTL           time.Duration `json:"ttl"`
	MaxSize       int           `json:"max_size"`
	CacheKey      string        `json:"cache_key,omitempty"`
	CacheStrategy string        `json:"cache_strategy"` // "lru", "lfu", "ttl"
}

// DefaultCacheConfig returns default cache configuration.
// Caching disabled by default, uses LRU strategy when enabled.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:       false,
		TTL:           time.Hour,
		MaxSize:       1000,
		CacheStrategy: "lru",
	}
}

// ConfigBuilder provides a fluent interface for building agent configurations.
// Supports method chaining for clean configuration construction.
type ConfigBuilder struct {
	config AgentConfig
}

// NewConfigBuilder creates a new configuration builder with default values.
// The builder starts with sensible defaults that can be customized.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: AgentConfig{
			Timeout:    30 * time.Second,
			MaxRetries: 3,
			RetryDelay: time.Second,
			Custom:     make(map[string]interface{}),
		},
	}
}

// WithTimeout sets the execution timeout for the agent.
// Returns the builder for method chaining.
func (b *ConfigBuilder) WithTimeout(timeout time.Duration) *ConfigBuilder {
	b.config.Timeout = timeout
	return b
}

// WithRetries sets the maximum retry attempts and delay between retries.
// Returns the builder for method chaining.
func (b *ConfigBuilder) WithRetries(maxRetries int, retryDelay time.Duration) *ConfigBuilder {
	b.config.MaxRetries = maxRetries
	b.config.RetryDelay = retryDelay
	return b
}

// WithAsync enables asynchronous execution with optional event streaming.
// Returns the builder for method chaining.
func (b *ConfigBuilder) WithAsync(streamEvents bool) *ConfigBuilder {
	b.config.Async = true
	b.config.StreamEvents = streamEvents
	return b
}

// WithStateSharing configures how state is shared between agents.
// Share enables state sharing, isolate creates separate state contexts.
// Returns the builder for method chaining.
func (b *ConfigBuilder) WithStateSharing(share bool, isolate bool) *ConfigBuilder {
	b.config.ShareState = share
	b.config.IsolateState = isolate
	return b
}

// WithCustom adds a custom configuration key-value pair.
// Custom configurations are passed to specific agent implementations.
// Returns the builder for method chaining.
func (b *ConfigBuilder) WithCustom(key string, value interface{}) *ConfigBuilder {
	b.config.Custom[key] = value
	return b
}

// Build returns the completed agent configuration.
// The configuration is ready to use with agent constructors.
func (b *ConfigBuilder) Build() AgentConfig {
	return b.config
}

// ValidateConfig validates an agent configuration for correctness.
// Checks timeout, retry, and state sharing settings for conflicts.
// Returns a ValidationError if the configuration is invalid.
func ValidateConfig(config AgentConfig) error {
	if config.Timeout < 0 {
		return NewValidationError("timeout", config.Timeout, "timeout must be non-negative")
	}

	if config.MaxRetries < 0 {
		return NewValidationError("max_retries", config.MaxRetries, "max retries must be non-negative")
	}

	if config.RetryDelay < 0 {
		return NewValidationError("retry_delay", config.RetryDelay, "retry delay must be non-negative")
	}

	if config.ShareState && config.IsolateState {
		return NewValidationError("state_config", nil, "cannot both share and isolate state")
	}

	return nil
}
