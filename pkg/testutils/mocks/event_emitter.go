// ABOUTME: Mock event emitter with event recording, filtering, and verification
// ABOUTME: Provides comprehensive event testing support including listeners and assertions

package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// EventListener represents a callback for event notifications
type EventListener func(event domain.Event)

// EventFilter represents a filter function for events
type EventFilter func(event domain.Event) bool

// MockEventEmitter implements domain.EventEmitter for testing
type MockEventEmitter struct {
	// Event storage
	events []domain.Event

	// Event listeners
	listeners []EventListener

	// Behavior hooks
	OnEmit         func(eventType domain.EventType, data interface{})
	OnEmitProgress func(current, total int, message string)
	OnEmitMessage  func(message string)
	OnEmitError    func(err error)
	OnEmitCustom   func(eventName string, data interface{})

	// Configuration
	blockEvents bool
	asyncEmit   bool
	eventDelay  time.Duration

	// Agent info for event creation
	agentID   string
	agentName string

	mu sync.RWMutex
}

// NewMockEventEmitter creates a new mock event emitter
func NewMockEventEmitter(agentID, agentName string) *MockEventEmitter {
	return &MockEventEmitter{
		events:    make([]domain.Event, 0),
		listeners: make([]EventListener, 0),
		agentID:   agentID,
		agentName: agentName,
	}
}

// Emit sends an event
func (m *MockEventEmitter) Emit(eventType domain.EventType, data interface{}) {
	if m.OnEmit != nil {
		m.OnEmit(eventType, data)
	}

	if m.blockEvents {
		return
	}

	event := domain.NewEvent(eventType, m.agentID, m.agentName, data)

	if m.asyncEmit {
		go m.recordAndNotify(event)
	} else {
		m.recordAndNotify(event)
	}
}

// EmitProgress sends a progress event
func (m *MockEventEmitter) EmitProgress(current, total int, message string) {
	if m.OnEmitProgress != nil {
		m.OnEmitProgress(current, total, message)
	}

	data := map[string]interface{}{
		"current": current,
		"total":   total,
		"message": message,
		"percent": float64(current) / float64(total) * 100,
	}

	m.Emit(domain.EventProgress, data)
}

// EmitMessage sends a message event
func (m *MockEventEmitter) EmitMessage(message string) {
	if m.OnEmitMessage != nil {
		m.OnEmitMessage(message)
	}

	data := map[string]interface{}{
		"message": message,
	}

	m.Emit(domain.EventMessage, data)
}

// EmitError sends an error event
func (m *MockEventEmitter) EmitError(err error) {
	if m.OnEmitError != nil {
		m.OnEmitError(err)
	}

	event := domain.NewEvent(domain.EventAgentError, m.agentID, m.agentName, nil).
		WithError(err)

	if m.asyncEmit {
		go m.recordAndNotify(event)
	} else {
		m.recordAndNotify(event)
	}
}

// EmitCustom sends a custom event
func (m *MockEventEmitter) EmitCustom(eventName string, data interface{}) {
	if m.OnEmitCustom != nil {
		m.OnEmitCustom(eventName, data)
	}

	// Use a custom event type
	customType := domain.EventType("custom." + eventName)
	m.Emit(customType, data)
}

// Helper methods for testing

// GetEvents returns all recorded events
func (m *MockEventEmitter) GetEvents() []domain.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := make([]domain.Event, len(m.events))
	copy(events, m.events)
	return events
}

// GetEventsByType returns events of a specific type
func (m *MockEventEmitter) GetEventsByType(eventType domain.EventType) []domain.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []domain.Event
	for _, event := range m.events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// GetEventsByFilter returns events matching a filter
func (m *MockEventEmitter) GetEventsByFilter(filter EventFilter) []domain.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []domain.Event
	for _, event := range m.events {
		if filter(event) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// AddListener adds an event listener
func (m *MockEventEmitter) AddListener(listener EventListener) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listeners = append(m.listeners, listener)
}

// RemoveAllListeners removes all listeners
func (m *MockEventEmitter) RemoveAllListeners() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listeners = make([]EventListener, 0)
}

// SetBlockEvents blocks or unblocks event emission
func (m *MockEventEmitter) SetBlockEvents(block bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blockEvents = block
}

// SetAsyncEmit enables/disables async event emission
func (m *MockEventEmitter) SetAsyncEmit(async bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.asyncEmit = async
}

// SetEventDelay sets a delay for event emission
func (m *MockEventEmitter) SetEventDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.eventDelay = delay
}

// Reset clears all events and listeners
func (m *MockEventEmitter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = make([]domain.Event, 0)
	m.listeners = make([]EventListener, 0)
	m.blockEvents = false
	m.asyncEmit = false
	m.eventDelay = 0
}

// Assertion helpers

// AssertEventEmitted checks if an event type was emitted
func (m *MockEventEmitter) AssertEventEmitted(eventType domain.EventType) error {
	events := m.GetEventsByType(eventType)
	if len(events) == 0 {
		return fmt.Errorf("expected event type %s to be emitted, but it wasn't", eventType)
	}
	return nil
}

// AssertEventCount checks the total number of events
func (m *MockEventEmitter) AssertEventCount(expected int) error {
	m.mu.RLock()
	actual := len(m.events)
	m.mu.RUnlock()

	if actual != expected {
		return fmt.Errorf("expected %d events, got %d", expected, actual)
	}
	return nil
}

// AssertEventTypeCount checks the count for a specific event type
func (m *MockEventEmitter) AssertEventTypeCount(eventType domain.EventType, expected int) error {
	events := m.GetEventsByType(eventType)
	actual := len(events)

	if actual != expected {
		return fmt.Errorf("expected %d events of type %s, got %d", expected, eventType, actual)
	}
	return nil
}

// AssertNoErrors checks that no error events were emitted
func (m *MockEventEmitter) AssertNoErrors() error {
	errorEvents := m.GetEventsByType(domain.EventAgentError)
	toolErrors := m.GetEventsByType(domain.EventToolError)

	totalErrors := len(errorEvents) + len(toolErrors)
	if totalErrors > 0 {
		return fmt.Errorf("expected no error events, but found %d", totalErrors)
	}

	// Also check for events with error field
	for _, event := range m.GetEvents() {
		if event.Error != nil {
			return fmt.Errorf("found event with error: %v", event.Error)
		}
	}

	return nil
}

// WaitForEvent waits for a specific event type with timeout
func (m *MockEventEmitter) WaitForEvent(eventType domain.EventType, timeout time.Duration) (*domain.Event, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		events := m.GetEventsByType(eventType)
		if len(events) > 0 {
			return &events[len(events)-1], nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for event type %s", eventType)
		}
	}
	return nil, fmt.Errorf("ticker stopped unexpectedly")
}

// Internal methods

func (m *MockEventEmitter) recordAndNotify(event domain.Event) {
	if m.eventDelay > 0 {
		time.Sleep(m.eventDelay)
	}

	m.mu.Lock()
	m.events = append(m.events, event)
	listeners := make([]EventListener, len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.Unlock()

	// Notify listeners outside the lock
	for _, listener := range listeners {
		listener(event)
	}
}

// CreateMockToolContext creates a ToolContext with mock components
func CreateMockToolContext(state *MockState, emitter *MockEventEmitter) *domain.ToolContext {
	return &domain.ToolContext{
		Context:   context.Background(),
		State:     state.State,
		RunID:     fmt.Sprintf("test-run-%d", time.Now().UnixNano()),
		Retry:     0,
		StartTime: time.Now(),
		Events:    emitter,
		Agent: domain.AgentInfo{
			ID:          "test-agent",
			Name:        "Test Agent",
			Description: "Mock agent for testing",
			Type:        domain.AgentTypeCustom,
			Metadata:    make(map[string]interface{}),
		},
	}
}
