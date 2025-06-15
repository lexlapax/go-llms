// ABOUTME: Helper functions for data tool testing
// ABOUTME: Provides mock implementations and ToolContext creation utilities

package data

import (
	"context"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

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
	agent := mocks.NewMockAgent("Test Agent")

	// Create a simple state with an immutable reader
	state := domain.NewState()
	stateReader := domain.NewStateReader(state)

	tc := domain.NewToolContext(ctx, stateReader, agent, "test-run-id")

	// Create a simple event emitter
	tc = tc.WithEventEmitter(&testEventEmitter{})

	return tc
}
