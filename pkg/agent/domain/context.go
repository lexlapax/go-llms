// Package domain defines the core domain models and interfaces for agents.
package domain

// ABOUTME: Defines the RunContext type for dependency injection in agent workflows
// ABOUTME: Provides type-safe context management for tool execution and state sharing

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RunContext provides type-safe dependency injection for agent execution.
// It carries dependencies, state, and metadata throughout the agent workflow.
// The generic type D allows custom dependency types while maintaining type safety.
type RunContext[D any] struct {
	ctx  context.Context
	deps D

	// Execution metadata
	RunID     string
	Retry     int
	StartTime time.Time

	// State access
	State *State

	// Shared state context (optional)
	SharedState *SharedStateContext

	// Event emission
	EmitEvent func(Event)
}

// NewRunContext creates a new run context with the provided dependencies.
// Generates a unique run ID and initializes with current timestamp.
// The event emitter defaults to a no-op function.
func NewRunContext[D any](ctx context.Context, deps D) *RunContext[D] {
	return &RunContext[D]{
		ctx:       ctx,
		deps:      deps,
		RunID:     uuid.New().String(),
		StartTime: time.Now(),
		EmitEvent: func(Event) {}, // Default no-op
	}
}

// NewRunContextWithState creates a new run context with initial state.
// Useful when state needs to be pre-populated before agent execution.
func NewRunContextWithState[D any](ctx context.Context, deps D, state *State) *RunContext[D] {
	rc := NewRunContext(ctx, deps)
	rc.State = state
	return rc
}

// NewRunContextWithSharedState creates a new run context with shared state access.
// The shared state allows coordination between multiple agents in a workflow.
// Local state is automatically initialized from the shared state.
func NewRunContextWithSharedState[D any](ctx context.Context, deps D, sharedState *SharedStateContext) *RunContext[D] {
	rc := NewRunContext(ctx, deps)
	rc.SharedState = sharedState
	// Also set the local state from shared state
	rc.State = sharedState.LocalState()
	return rc
}

// Deps returns the typed dependencies stored in the context.
// These dependencies are available throughout the agent execution.
func (r *RunContext[D]) Deps() D {
	return r.deps
}

// Context returns the underlying Go context for cancellation and deadlines.
// This context flows through all agent operations.
func (r *RunContext[D]) Context() context.Context {
	return r.ctx
}

// WithRetry creates a new context for a retry attempt.
// Preserves all other context data while updating the retry count.
func (rc *RunContext[D]) WithRetry(retry int) *RunContext[D] {
	newCtx := *rc
	newCtx.Retry = retry
	return &newCtx
}

// WithState creates a new context with different state.
// Useful for creating isolated state contexts in agent hierarchies.
func (rc *RunContext[D]) WithState(state *State) *RunContext[D] {
	newCtx := *rc
	newCtx.State = state
	return &newCtx
}

// WithSharedState creates a new context with shared state access.
// Updates both shared state reference and local state snapshot.
func (rc *RunContext[D]) WithSharedState(sharedState *SharedStateContext) *RunContext[D] {
	newCtx := *rc
	newCtx.SharedState = sharedState
	newCtx.State = sharedState.LocalState()
	return &newCtx
}

// WithEventEmitter sets the event emission function for the context.
// Events emitted through this function flow to the agent's event stream.
func (rc *RunContext[D]) WithEventEmitter(emitter func(Event)) *RunContext[D] {
	newCtx := *rc
	newCtx.EmitEvent = emitter
	return &newCtx
}

// Elapsed returns the time elapsed since the run started.
// Useful for monitoring execution time and implementing timeouts.
func (rc *RunContext[D]) Elapsed() time.Duration {
	return time.Since(rc.StartTime)
}

// EmitProgress emits a progress event with current/total counts.
// Agent ID and name are filled by the event emitter infrastructure.
func (rc *RunContext[D]) EmitProgress(current, total int, message string) {
	if rc.EmitEvent != nil {
		rc.EmitEvent(NewEvent(
			EventProgress,
			"", // Agent ID will be filled by the emitter
			"", // Agent name will be filled by the emitter
			ProgressEventData{
				Current: current,
				Total:   total,
				Message: message,
			},
		))
	}
}

// EmitMessage emits a simple message event for logging or debugging.
// The message flows through the agent's event stream to observers.
func (rc *RunContext[D]) EmitMessage(message string) {
	if rc.EmitEvent != nil {
		rc.EmitEvent(NewEvent(
			EventMessage,
			"", // Agent ID will be filled by the emitter
			"", // Agent name will be filled by the emitter
			message,
		))
	}
}

// Example dependency types that can be used with RunContext

// DatabaseDeps holds database-related dependencies for data access.
// Interface types allow flexibility in concrete implementations.
type DatabaseDeps struct {
	DB     interface{} // *sql.DB
	Cache  interface{} // *redis.Client
	Logger interface{} // *slog.Logger
}

// ServiceDeps holds service layer dependencies for business logic.
// Commonly used for domain services in layered architectures.
type ServiceDeps struct {
	UserService    interface{}
	ProductService interface{}
	OrderService   interface{}
}

// ToolDeps holds tool-related dependencies for agent tool execution.
// Includes available tools and execution timeout configuration.
type ToolDeps struct {
	Tools       map[string]Tool
	ToolTimeout time.Duration
}

// LLMDeps holds LLM provider dependencies for language model interactions.
// Supports different providers, models, and provider-specific options.
type LLMDeps struct {
	Provider interface{} // llm.Provider
	Model    string
	Options  map[string]interface{}
}

// HTTPDeps holds HTTP client dependencies for external API calls.
// Includes authentication, base URLs, and rate limiting support.
type HTTPDeps struct {
	Client      interface{} // *http.Client
	BaseURL     string
	APIKey      string
	RateLimiter interface{} // rate.Limiter
}

// ObservabilityDeps holds observability dependencies for monitoring.
// Supports distributed tracing, metrics, logging, and sampling.
type ObservabilityDeps struct {
	Tracer  interface{} // trace.Tracer
	Meter   interface{} // metric.Meter
	Logger  interface{} // *slog.Logger
	Sampler interface{} // trace.Sampler
}

// CompositeDeps combines multiple dependency types in a single structure.
// Useful for agents that need access to various infrastructure components.
type CompositeDeps struct {
	DB            DatabaseDeps
	Services      ServiceDeps
	HTTP          HTTPDeps
	Observability ObservabilityDeps
}

// Helper functions for common context operations

// GetFromState safely gets a typed value from state with a default fallback.
// Checks shared state first if available, then falls back to local state.
// Returns defaultVal if the key is not found or type conversion fails.
func GetFromState[T any](rc *RunContext[any], key string, defaultVal T) T {
	// Try shared state first if available
	if rc.SharedState != nil {
		if val, ok := rc.SharedState.Get(key); ok {
			if typed, ok := val.(T); ok {
				return typed
			}
		}
	}

	// Fall back to regular state
	if rc.State == nil {
		return defaultVal
	}

	val, ok := rc.State.Get(key)
	if !ok {
		return defaultVal
	}

	typed, ok := val.(T)
	if !ok {
		return defaultVal
	}

	return typed
}

// MustGetFromState gets a typed value from state or panics if not found.
// Checks shared state first if available, then falls back to local state.
// Panics with descriptive message if key is missing or type is wrong.
func MustGetFromState[T any](rc *RunContext[any], key string) T {
	// Try shared state first if available
	if rc.SharedState != nil {
		if val, ok := rc.SharedState.Get(key); ok {
			if typed, ok := val.(T); ok {
				return typed
			}
			panic("state value has wrong type for key: " + key)
		}
	}

	// Fall back to regular state
	if rc.State == nil {
		panic("state is nil")
	}

	val, ok := rc.State.Get(key)
	if !ok {
		panic("required state key not found: " + key)
	}

	typed, ok := val.(T)
	if !ok {
		panic("state value has wrong type for key: " + key)
	}

	return typed
}
