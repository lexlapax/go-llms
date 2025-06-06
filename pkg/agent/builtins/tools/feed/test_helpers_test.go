// ABOUTME: Test helper functions and mocks for feed tool tests
// ABOUTME: Provides mock agent, event emitter, and test context creation

package feed

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

//nolint:unused // Mock agent for testing that implements BaseAgent interface
type mockFeedTestAgent struct{}

func (m *mockFeedTestAgent) ID() string                                { return "test-agent" }
func (m *mockFeedTestAgent) Name() string                              { return "Test Agent" }
func (m *mockFeedTestAgent) Description() string                       { return "A test agent" }
func (m *mockFeedTestAgent) Type() domain.AgentType                    { return domain.AgentTypeLLM }
func (m *mockFeedTestAgent) Parent() domain.BaseAgent                  { return nil }
func (m *mockFeedTestAgent) SetParent(parent domain.BaseAgent) error   { return nil }
func (m *mockFeedTestAgent) SubAgents() []domain.BaseAgent             { return nil }
func (m *mockFeedTestAgent) AddSubAgent(agent domain.BaseAgent) error  { return nil }
func (m *mockFeedTestAgent) RemoveSubAgent(name string) error          { return nil }
func (m *mockFeedTestAgent) FindAgent(name string) domain.BaseAgent    { return nil }
func (m *mockFeedTestAgent) FindSubAgent(name string) domain.BaseAgent { return nil }
func (m *mockFeedTestAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	return input, nil
}
func (m *mockFeedTestAgent) RunAsync(ctx context.Context, input *domain.State) (<-chan domain.Event, error) {
	return nil, nil
}
func (m *mockFeedTestAgent) Initialize(ctx context.Context) error                     { return nil }
func (m *mockFeedTestAgent) BeforeRun(ctx context.Context, state *domain.State) error { return nil }
func (m *mockFeedTestAgent) AfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) error {
	return nil
}
func (m *mockFeedTestAgent) Cleanup(ctx context.Context) error                     { return nil }
func (m *mockFeedTestAgent) InputSchema() *sdomain.Schema                          { return nil }
func (m *mockFeedTestAgent) OutputSchema() *sdomain.Schema                         { return nil }
func (m *mockFeedTestAgent) Config() domain.AgentConfig                            { return domain.AgentConfig{} }
func (m *mockFeedTestAgent) WithConfig(config domain.AgentConfig) domain.BaseAgent { return m }
func (m *mockFeedTestAgent) Validate() error                                       { return nil }
func (m *mockFeedTestAgent) Metadata() map[string]interface{}                      { return nil }
func (m *mockFeedTestAgent) SetMetadata(key string, value interface{})             {}

// Test event emitter
type testEventEmitterFeedTest struct {
	events []domain.Event
}

func (e *testEventEmitterFeedTest) Emit(eventType domain.EventType, data interface{}) {
	e.events = append(e.events, domain.Event{
		Type: eventType,
		Data: data,
	})
}

func (e *testEventEmitterFeedTest) EmitProgress(current, total int, message string) {
	e.events = append(e.events, domain.Event{
		Type: domain.EventProgress,
		Data: map[string]interface{}{
			"current": current,
			"total":   total,
			"message": message,
		},
	})
}

func (e *testEventEmitterFeedTest) EmitMessage(message string) {
	e.events = append(e.events, domain.Event{
		Type: domain.EventMessage,
		Data: message,
	})
}

func (e *testEventEmitterFeedTest) EmitError(err error) {
	e.events = append(e.events, domain.Event{
		Type: domain.EventToolError,
		Data: err,
	})
}

func (e *testEventEmitterFeedTest) EmitCustom(eventName string, data interface{}) {
	e.events = append(e.events, domain.Event{
		Type: domain.EventType(eventName),
		Data: data,
	})
}

func (e *testEventEmitterFeedTest) GetEvents() []domain.Event {
	return e.events
}

// Helper function to create test tool context
func createTestToolContext() *domain.ToolContext {
	ctx := context.Background()
	state := domain.NewState()
	stateReader := domain.NewStateReader(state)
	events := &testEventEmitterFeedTest{}

	agentInfo := domain.AgentInfo{
		ID:          "test-agent",
		Name:        "Test Agent",
		Description: "A test agent",
		Type:        domain.AgentTypeLLM,
	}

	return &domain.ToolContext{
		Context: ctx,
		State:   stateReader,
		Agent:   agentInfo,
		Events:  events,
		RunID:   "test-run-123",
	}
}
