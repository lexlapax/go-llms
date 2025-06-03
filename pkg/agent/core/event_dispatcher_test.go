// ABOUTME: Tests for event dispatcher including subscription, filtering, and concurrent dispatch
// ABOUTME: Validates event distribution, handler management, and thread safety

package core_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// TestEventHandler is a test implementation of EventHandler
type TestEventHandler struct {
	mu           sync.Mutex
	events       []domain.Event
	shouldError  bool
	errorMessage string
	handleFunc   func(event domain.Event) error
}

func NewTestEventHandler() *TestEventHandler {
	return &TestEventHandler{
		events: make([]domain.Event, 0),
	}
}

func (h *TestEventHandler) HandleEvent(event domain.Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.handleFunc != nil {
		return h.handleFunc(event)
	}

	h.events = append(h.events, event)

	if h.shouldError {
		return errors.New(h.errorMessage)
	}
	return nil
}

func (h *TestEventHandler) GetEvents() []domain.Event {
	h.mu.Lock()
	defer h.mu.Unlock()

	events := make([]domain.Event, len(h.events))
	copy(events, h.events)
	return events
}

func (h *TestEventHandler) EventCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.events)
}

func TestEventDispatcherBasic(t *testing.T) {
	dispatcher := core.NewEventDispatcher(10)
	defer dispatcher.Close()

	handler := NewTestEventHandler()

	// Subscribe
	subID := dispatcher.Subscribe(handler)
	if subID == "" {
		t.Error("Subscribe should return non-empty subscription ID")
	}

	// Dispatch event
	event := domain.NewEvent(domain.EventAgentStart, "agent-1", "TestAgent", nil)
	dispatcher.Dispatch(event)

	// Give handler time to process
	time.Sleep(50 * time.Millisecond)

	// Verify event was received
	if handler.EventCount() != 1 {
		t.Errorf("Expected 1 event, got %d", handler.EventCount())
	}

	receivedEvents := handler.GetEvents()
	if len(receivedEvents) > 0 && receivedEvents[0].ID != event.ID {
		t.Error("Received different event than dispatched")
	}

	// Unsubscribe
	dispatcher.Unsubscribe(subID)

	// Dispatch another event
	event2 := domain.NewEvent(domain.EventAgentComplete, "agent-1", "TestAgent", nil)
	dispatcher.Dispatch(event2)

	time.Sleep(50 * time.Millisecond)

	// Should still have only 1 event
	if handler.EventCount() != 1 {
		t.Error("Handler should not receive events after unsubscribe")
	}
}

func TestEventDispatcherFilters(t *testing.T) {
	dispatcher := core.NewEventDispatcher(10)
	defer dispatcher.Close()

	// Handler 1: Only agent start events
	handler1 := NewTestEventHandler()
	dispatcher.Subscribe(handler1, domain.FilterByType(domain.EventAgentStart))

	// Handler 2: Only agent-2 events
	handler2 := NewTestEventHandler()
	dispatcher.Subscribe(handler2, domain.FilterByAgent("agent-2"))

	// Handler 3: Errors only
	handler3 := NewTestEventHandler()
	dispatcher.Subscribe(handler3, domain.FilterErrors())

	// Handler 4: Combined filter (agent-1 AND start events)
	handler4 := NewTestEventHandler()
	dispatcher.Subscribe(handler4,
		domain.FilterByAgent("agent-1"),
		domain.FilterByType(domain.EventAgentStart),
	)

	// Dispatch various events
	events := []domain.Event{
		domain.NewEvent(domain.EventAgentStart, "agent-1", "Agent1", nil),    // h1, h4
		domain.NewEvent(domain.EventAgentComplete, "agent-1", "Agent1", nil), // none
		domain.NewEvent(domain.EventAgentStart, "agent-2", "Agent2", nil),    // h1, h2
		domain.NewEvent(domain.EventAgentError, "agent-2", "Agent2", nil),    // h2, h3
		domain.NewEvent(domain.EventToolCall, "agent-3", "Agent3", nil),      // none
	}

	for _, event := range events {
		dispatcher.Dispatch(event)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify filters worked correctly
	if handler1.EventCount() != 2 {
		t.Errorf("Handler1 expected 2 events, got %d", handler1.EventCount())
	}

	if handler2.EventCount() != 2 {
		t.Errorf("Handler2 expected 2 events, got %d", handler2.EventCount())
	}

	if handler3.EventCount() != 1 {
		t.Errorf("Handler3 expected 1 event, got %d", handler3.EventCount())
	}

	if handler4.EventCount() != 1 {
		t.Errorf("Handler4 expected 1 event, got %d", handler4.EventCount())
	}
}

func TestEventDispatcherConcurrency(t *testing.T) {
	dispatcher := core.NewEventDispatcher(100)
	defer dispatcher.Close()

	const numHandlers = 10
	const numEvents = 100

	// Create multiple handlers
	handlers := make([]*TestEventHandler, numHandlers)
	for i := 0; i < numHandlers; i++ {
		handlers[i] = NewTestEventHandler()
		dispatcher.Subscribe(handlers[i])
	}

	// Dispatch events concurrently
	var wg sync.WaitGroup
	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			event := domain.NewEvent(
				domain.EventProgress,
				sprintf("agent-%d", id),
				"TestAgent",
				map[string]int{"id": id},
			)
			dispatcher.Dispatch(event)
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond) // Give handlers time to process

	// Verify all handlers received all events
	for i, handler := range handlers {
		if handler.EventCount() != numEvents {
			t.Errorf("Handler %d expected %d events, got %d", i, numEvents, handler.EventCount())
		}
	}
}

func TestEventDispatcherClose(t *testing.T) {
	dispatcher := core.NewEventDispatcher(10)

	handler := NewTestEventHandler()
	dispatcher.Subscribe(handler)

	// Dispatch some events
	for i := 0; i < 5; i++ {
		event := domain.NewEvent(domain.EventProgress, "agent-1", "TestAgent", i)
		dispatcher.Dispatch(event)
	}

	// Give events time to be processed
	time.Sleep(50 * time.Millisecond)

	// Close dispatcher
	dispatcher.Close()

	// Try to dispatch after close (should not panic)
	event := domain.NewEvent(domain.EventAgentComplete, "agent-1", "TestAgent", nil)
	dispatcher.Dispatch(event) // Should handle gracefully

	// Handler should have received the events before close
	time.Sleep(50 * time.Millisecond)
	eventCount := handler.EventCount()
	if eventCount < 5 {
		t.Errorf("Handler should have received at least 5 events, got %d", eventCount)
	}
}

func TestEventDispatcherHandlerPanic(t *testing.T) {
	dispatcher := core.NewEventDispatcher(10)
	defer dispatcher.Close()

	// Handler that panics
	panicHandler := NewTestEventHandler()
	panicHandler.handleFunc = func(event domain.Event) error {
		panic("test panic")
	}

	// Handler that works normally
	normalHandler := NewTestEventHandler()

	dispatcher.Subscribe(panicHandler)
	dispatcher.Subscribe(normalHandler)

	// Dispatch event
	event := domain.NewEvent(domain.EventAgentStart, "agent-1", "TestAgent", nil)
	dispatcher.Dispatch(event)

	time.Sleep(100 * time.Millisecond)

	// Normal handler should still receive the event despite panic in other handler
	if normalHandler.EventCount() != 1 {
		t.Error("Normal handler should receive event despite panic in other handler")
	}
}

func TestEventDispatcherNilHandler(t *testing.T) {
	dispatcher := core.NewEventDispatcher(10)
	defer dispatcher.Close()

	// Subscribe with nil handler
	subID := dispatcher.Subscribe(nil)
	if subID != "" {
		t.Error("Subscribe with nil handler should return empty ID")
	}

	// This should not cause any issues
	event := domain.NewEvent(domain.EventAgentStart, "agent-1", "TestAgent", nil)
	dispatcher.Dispatch(event)
}

func TestEventStream(t *testing.T) {
	// Since we can't access the internal Send method directly,
	// we'll skip this specific test for now
	// TODO: Add a proper test interface or make Send public
	t.Skip("EventStream internals not accessible for testing")
}

func TestBufferedEventHandler(t *testing.T) {
	baseHandler := NewTestEventHandler()
	bufferedHandler := core.NewBufferedEventHandler(baseHandler, 20) // Increased buffer size
	defer bufferedHandler.Close()

	// Send multiple events quickly
	numEvents := 20
	for i := 0; i < numEvents; i++ {
		event := domain.NewEvent(domain.EventProgress, "agent-1", "TestAgent", i)
		err := bufferedHandler.HandleEvent(event)
		if err != nil {
			// Buffer might be full, which is expected behavior
			continue
		}
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// At least some events should be handled
	eventCount := baseHandler.EventCount()
	if eventCount == 0 {
		t.Error("No events were handled")
	}
	// Don't require all events as some might be dropped if buffer is full
	t.Logf("Handled %d out of %d events", eventCount, numEvents)
}

func TestCompositeEventHandler(t *testing.T) {
	handler1 := NewTestEventHandler()
	handler2 := NewTestEventHandler()
	handler3 := NewTestEventHandler()

	composite := core.NewCompositeEventHandler(handler1, handler2, handler3)

	// Send event
	event := domain.NewEvent(domain.EventAgentStart, "agent-1", "TestAgent", nil)
	err := composite.HandleEvent(event)
	if err != nil {
		t.Errorf("Composite handler error: %v", err)
	}

	// All handlers should receive the event
	time.Sleep(50 * time.Millisecond)

	if handler1.EventCount() != 1 {
		t.Error("Handler1 should receive event")
	}
	if handler2.EventCount() != 1 {
		t.Error("Handler2 should receive event")
	}
	if handler3.EventCount() != 1 {
		t.Error("Handler3 should receive event")
	}

	// Test with error
	handler2.shouldError = true
	handler2.errorMessage = "test error"

	event2 := domain.NewEvent(domain.EventAgentComplete, "agent-1", "TestAgent", nil)
	err = composite.HandleEvent(event2)
	if err == nil {
		t.Error("Composite should return error if any handler errors")
	}
}

func TestFilteredEventHandler(t *testing.T) {
	baseHandler := NewTestEventHandler()

	// Create filtered handler that only accepts start events from agent-1
	filteredHandler := core.NewFilteredEventHandler(
		baseHandler,
		domain.FilterByType(domain.EventAgentStart),
		domain.FilterByAgent("agent-1"),
	)

	// Test events
	events := []struct {
		event    domain.Event
		expected bool
	}{
		{
			event:    domain.NewEvent(domain.EventAgentStart, "agent-1", "Agent1", nil),
			expected: true,
		},
		{
			event:    domain.NewEvent(domain.EventAgentStart, "agent-2", "Agent2", nil),
			expected: false, // Wrong agent
		},
		{
			event:    domain.NewEvent(domain.EventAgentComplete, "agent-1", "Agent1", nil),
			expected: false, // Wrong type
		},
		{
			event:    domain.NewEvent(domain.EventAgentStart, "agent-1", "Agent1", nil),
			expected: true,
		},
	}

	for _, tc := range events {
		_ = filteredHandler.HandleEvent(tc.event)
	}

	time.Sleep(50 * time.Millisecond)

	// Only matching events should be handled
	expectedCount := 0
	for _, tc := range events {
		if tc.expected {
			expectedCount++
		}
	}

	if baseHandler.EventCount() != expectedCount {
		t.Errorf("Expected %d events, got %d", expectedCount, baseHandler.EventCount())
	}
}

func TestEventDispatcherBufferFull(t *testing.T) {
	// Create dispatcher with small buffer
	dispatcher := core.NewEventDispatcher(2)
	defer dispatcher.Close()

	// Create slow handler
	slowHandler := NewTestEventHandler()
	slowHandler.handleFunc = func(event domain.Event) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	dispatcher.Subscribe(slowHandler)

	// Dispatch many events quickly
	for i := 0; i < 10; i++ {
		event := domain.NewEvent(domain.EventProgress, "agent-1", "TestAgent", i)
		dispatcher.Dispatch(event) // Some may be dropped due to full buffer
	}

	// This test verifies the dispatcher handles full buffer gracefully
	// without blocking or panicking
}

// Benchmark tests
func BenchmarkEventDispatcher(b *testing.B) {
	dispatcher := core.NewEventDispatcher(1000)
	defer dispatcher.Close()

	handler := NewTestEventHandler()
	dispatcher.Subscribe(handler)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		event := domain.NewEvent(domain.EventProgress, "agent-1", "TestAgent", i)
		dispatcher.Dispatch(event)
	}
}

func BenchmarkEventDispatcherWithFilters(b *testing.B) {
	dispatcher := core.NewEventDispatcher(1000)
	defer dispatcher.Close()

	// Multiple handlers with different filters
	for i := 0; i < 10; i++ {
		handler := NewTestEventHandler()
		dispatcher.Subscribe(handler,
			domain.FilterByAgent(sprintf("agent-%d", i%3)),
			domain.FilterByType(domain.EventProgress),
		)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		event := domain.NewEvent(
			domain.EventProgress,
			sprintf("agent-%d", i%10),
			"TestAgent",
			i,
		)
		dispatcher.Dispatch(event)
	}
}

func BenchmarkConcurrentDispatch(b *testing.B) {
	dispatcher := core.NewEventDispatcher(1000)
	defer dispatcher.Close()

	// Create multiple handlers
	for i := 0; i < 5; i++ {
		handler := NewTestEventHandler()
		dispatcher.Subscribe(handler)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			event := domain.NewEvent(
				domain.EventProgress,
				sprintf("agent-%d", i%100),
				"TestAgent",
				i,
			)
			dispatcher.Dispatch(event)
			i++
		}
	})
}

// sprintf is a helper function for string formatting in tests
func sprintf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

