// ABOUTME: Test helper functions and mocks for feed tool tests
// ABOUTME: Provides mock agent, event emitter, and test context creation

package feed

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

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
