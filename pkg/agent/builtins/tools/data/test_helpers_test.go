// ABOUTME: Helper functions for data tool testing
// ABOUTME: Provides mock implementations and ToolContext creation utilities

package data

import (
	"context"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// mockAgent implements the BaseAgent interface for testing
type mockAgent struct {
	id          string
	name        string
	description string
	agentType   domain.AgentType
	metadata    map[string]interface{}
}

func (a *mockAgent) ID() string                       { return a.id }
func (a *mockAgent) Name() string                     { return a.name }
func (a *mockAgent) Description() string              { return a.description }
func (a *mockAgent) Type() domain.AgentType           { return a.agentType }
func (a *mockAgent) Metadata() map[string]interface{} { return a.metadata }

// Hierarchy Management
func (a *mockAgent) Parent() domain.BaseAgent                  { return nil }
func (a *mockAgent) SetParent(parent domain.BaseAgent) error   { return nil }
func (a *mockAgent) SubAgents() []domain.BaseAgent             { return nil }
func (a *mockAgent) AddSubAgent(agent domain.BaseAgent) error  { return nil }
func (a *mockAgent) RemoveSubAgent(name string) error          { return nil }
func (a *mockAgent) FindAgent(name string) domain.BaseAgent    { return nil }
func (a *mockAgent) FindSubAgent(name string) domain.BaseAgent { return nil }

// Execution
func (a *mockAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	return input, nil
}
func (a *mockAgent) RunAsync(ctx context.Context, input *domain.State) (<-chan domain.Event, error) {
	return nil, nil
}

// Lifecycle Hooks
func (a *mockAgent) Initialize(ctx context.Context) error                     { return nil }
func (a *mockAgent) BeforeRun(ctx context.Context, state *domain.State) error { return nil }
func (a *mockAgent) AfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) error {
	return nil
}
func (a *mockAgent) Cleanup(ctx context.Context) error { return nil }

// Schema Definition
func (a *mockAgent) InputSchema() *sdomain.Schema  { return nil }
func (a *mockAgent) OutputSchema() *sdomain.Schema { return nil }

// Configuration
func (a *mockAgent) Config() domain.AgentConfig                            { return domain.AgentConfig{} }
func (a *mockAgent) WithConfig(config domain.AgentConfig) domain.BaseAgent { return a }
func (a *mockAgent) Validate() error                                       { return nil }

// Metadata
func (a *mockAgent) SetMetadata(key string, value interface{}) {
	if a.metadata == nil {
		a.metadata = make(map[string]interface{})
	}
	a.metadata[key] = value
}

// testEventEmitter implements the EventEmitter interface for testing
type testEventEmitter struct {
	events []domain.Event
}

func (e *testEventEmitter) Emit(eventType domain.EventType, data interface{}) {
	e.events = append(e.events, domain.Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	})
}

func (e *testEventEmitter) EmitCustom(eventName string, data interface{}) {
	e.events = append(e.events, domain.Event{
		Type:      domain.EventType(eventName),
		Timestamp: time.Now(),
		Data:      data,
	})
}

func (e *testEventEmitter) EmitProgress(current, total int, message string) {
	e.Emit(domain.EventType("progress"), map[string]interface{}{
		"current": current,
		"total":   total,
		"message": message,
	})
}

func (e *testEventEmitter) EmitMessage(message string) {
	e.Emit(domain.EventType("message"), message)
}

func (e *testEventEmitter) EmitError(err error) {
	e.Emit(domain.EventType("error"), err.Error())
}

func (e *testEventEmitter) Subscribe(eventType domain.EventType, handler domain.EventHandler) {
	// Not needed for tests
}

func (e *testEventEmitter) GetEvents() []domain.Event {
	return e.events
}

// createTestToolContext creates a ToolContext for testing
func createTestToolContext(ctx context.Context) *domain.ToolContext {
	// Create a simple agent for testing
	agent := &mockAgent{
		id:          "test-agent",
		name:        "Test Agent",
		description: "Agent for testing",
		agentType:   domain.AgentTypeLLM,
		metadata:    make(map[string]interface{}),
	}

	// Create a simple state with an immutable reader
	state := domain.NewState()
	stateReader := domain.NewStateReader(state)

	tc := domain.NewToolContext(ctx, stateReader, agent, "test-run-id")

	// Create a simple event emitter
	tc = tc.WithEventEmitter(&testEventEmitter{})

	return tc
}

// createTestToolContextWithState creates a ToolContext with predefined state values
func createTestToolContextWithState(ctx context.Context, values map[string]interface{}) *domain.ToolContext {
	// Create a simple agent for testing
	agent := &mockAgent{
		id:          "test-agent",
		name:        "Test Agent",
		description: "Agent for testing",
		agentType:   domain.AgentTypeLLM,
		metadata:    make(map[string]interface{}),
	}

	// Create a state with the provided values
	state := domain.NewState()
	for k, v := range values {
		state.Set(k, v)
	}
	stateReader := domain.NewStateReader(state)

	tc := domain.NewToolContext(ctx, stateReader, agent, "test-run-id")

	// Create a simple event emitter
	tc = tc.WithEventEmitter(&testEventEmitter{})

	return tc
}
