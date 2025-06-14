// ABOUTME: Test context helpers for creating pre-configured test contexts
// ABOUTME: Provides builders for ToolContext and AgentContext with common test scenarios

package helpers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// ContextOption is a function that configures a test context
type ContextOption func(interface{})

// ToolContextOptions holds configuration for creating a test ToolContext
type ToolContextOptions struct {
	State        *domain.State
	EventEmitter domain.EventEmitter
	AgentInfo    domain.AgentInfo
	RunID        string
	Retry        int
	Context      context.Context
}

// CreateTestToolContext creates a ToolContext for testing with sensible defaults
func CreateTestToolContext(options ...ContextOption) *domain.ToolContext {
	opts := &ToolContextOptions{
		State:        domain.NewState(),
		EventEmitter: mocks.NewMockEventEmitter("test-agent", "Test Agent"),
		AgentInfo: domain.AgentInfo{
			ID:          "test-agent",
			Name:        "Test Agent",
			Description: "Agent for testing",
			Type:        domain.AgentTypeCustom,
			Metadata:    make(map[string]interface{}),
		},
		RunID:   uuid.New().String(),
		Retry:   0,
		Context: context.Background(),
	}

	// Apply options
	for _, opt := range options {
		opt(opts)
	}

	return &domain.ToolContext{
		Context:   opts.Context,
		State:     opts.State,
		RunID:     opts.RunID,
		Retry:     opts.Retry,
		StartTime: time.Now(),
		Events:    opts.EventEmitter,
		Agent:     opts.AgentInfo,
	}
}

// WithTestState sets the state for the test context
func WithTestState(state *domain.State) ContextOption {
	return func(opts interface{}) {
		switch o := opts.(type) {
		case *ToolContextOptions:
			o.State = state
		case *AgentContextOptions:
			o.State = state
		}
	}
}

// WithTestEventEmitter sets the event emitter for the test context
func WithTestEventEmitter(emitter domain.EventEmitter) ContextOption {
	return func(opts interface{}) {
		switch o := opts.(type) {
		case *ToolContextOptions:
			o.EventEmitter = emitter
		case *AgentContextOptions:
			o.EventEmitter = emitter
		}
	}
}

// WithTestRunID sets the run ID for the test context
func WithTestRunID(runID string) ContextOption {
	return func(opts interface{}) {
		switch o := opts.(type) {
		case *ToolContextOptions:
			o.RunID = runID
		case *AgentContextOptions:
			o.RunID = runID
		}
	}
}

// WithTestRetry sets the retry count for the test context
func WithTestRetry(retry int) ContextOption {
	return func(opts interface{}) {
		if o, ok := opts.(*ToolContextOptions); ok {
			o.Retry = retry
		}
	}
}

// WithTestContext sets the Go context for the test context
func WithTestContext(ctx context.Context) ContextOption {
	return func(opts interface{}) {
		switch o := opts.(type) {
		case *ToolContextOptions:
			o.Context = ctx
		case *AgentContextOptions:
			o.Context = ctx
		}
	}
}

// WithTestAgentInfo sets the agent info for the test context
func WithTestAgentInfo(info domain.AgentInfo) ContextOption {
	return func(opts interface{}) {
		if o, ok := opts.(*ToolContextOptions); ok {
			o.AgentInfo = info
		}
	}
}

// AgentContextOptions holds configuration for creating a test agent context
type AgentContextOptions struct {
	State        *domain.State
	EventEmitter domain.EventEmitter
	RunID        string
	Context      context.Context
	Dependencies interface{}
}

// CreateTestAgentContext creates a RunContext for testing agents
func CreateTestAgentContext[D any](deps D, options ...ContextOption) *domain.RunContext[D] {
	opts := &AgentContextOptions{
		State:        domain.NewState(),
		EventEmitter: mocks.NewMockEventEmitter("test-agent", "Test Agent"),
		RunID:        uuid.New().String(),
		Context:      context.Background(),
		Dependencies: deps,
	}

	// Apply options
	for _, opt := range options {
		opt(opts)
	}

	rc := domain.NewRunContextWithState(opts.Context, deps, opts.State)
	rc.RunID = opts.RunID

	if opts.EventEmitter != nil {
		rc = rc.WithEventEmitter(func(event domain.Event) {
			opts.EventEmitter.Emit(event.Type, event.Data)
		})
	}

	return rc
}

// Common pre-configured contexts

// CreateToolContextWithError creates a tool context that simulates error conditions
func CreateToolContextWithError() *domain.ToolContext {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Nanosecond) // Ensure context is canceled

	return CreateTestToolContext(
		WithTestContext(ctx),
		WithTestRetry(3), // Simulate retries
	)
}

// CreateToolContextWithState creates a tool context with pre-populated state
func CreateToolContextWithState(data map[string]interface{}) *domain.ToolContext {
	state := domain.NewState()
	for k, v := range data {
		state.Set(k, v)
	}

	return CreateTestToolContext(WithTestState(state))
}

// CreateToolContextWithTimeout creates a tool context with a specific timeout
func CreateToolContextWithTimeout(timeout time.Duration) *domain.ToolContext {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_ = cancel // The cancel is not needed as the context is used immediately
	return CreateTestToolContext(WithTestContext(ctx))
}

// CreateAgentContextWithDeps creates an agent context with typed dependencies
func CreateAgentContextWithDeps[D any](deps D) *domain.RunContext[D] {
	return CreateTestAgentContext(deps)
}

// CreateAgentContextWithState creates an agent context with pre-populated state
func CreateAgentContextWithState[D any](deps D, data map[string]interface{}) *domain.RunContext[D] {
	state := domain.NewState()
	for k, v := range data {
		state.Set(k, v)
	}

	return CreateTestAgentContext(deps, WithTestState(state))
}

// CreateAgentContextWithTimeout creates an agent context with a specific timeout
func CreateAgentContextWithTimeout[D any](deps D, timeout time.Duration) *domain.RunContext[D] {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_ = cancel // The cancel is not needed as the context is used immediately
	return CreateTestAgentContext(deps, WithTestContext(ctx))
}
