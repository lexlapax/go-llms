// ABOUTME: Tests for EventBus implementation
// ABOUTME: Validates subscription, publishing, filtering, and thread safety

package events

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestEventBus_Subscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	received := make(chan domain.Event, 1)
	handler := EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		received <- event
		return nil
	})

	subID := bus.Subscribe(handler)
	if subID == "" {
		t.Fatal("Expected subscription ID")
	}

	// Publish event
	event := domain.NewEvent(domain.EventAgentStart, "agent1", "TestAgent", nil)
	bus.Publish(event)

	// Wait for event
	select {
	case e := <-received:
		if e.Type != domain.EventAgentStart {
			t.Errorf("Expected event type %s, got %s", domain.EventAgentStart, e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Event not received")
	}
}

func TestEventBus_SubscribePattern(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	tests := []struct {
		name        string
		pattern     string
		eventType   domain.EventType
		shouldMatch bool
	}{
		{
			name:        "exact match",
			pattern:     "agent.start",
			eventType:   domain.EventAgentStart,
			shouldMatch: true,
		},
		{
			name:        "wildcard match",
			pattern:     "tool.*",
			eventType:   domain.EventToolCall,
			shouldMatch: true,
		},
		{
			name:        "wildcard no match",
			pattern:     "tool.*",
			eventType:   domain.EventAgentStart,
			shouldMatch: false,
		},
		{
			name:        "all events",
			pattern:     ".*",
			eventType:   domain.EventProgress,
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			received := make(chan domain.Event, 1)
			handler := EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
				received <- event
				return nil
			})

			subID, err := bus.SubscribePattern(tt.pattern, handler)
			if err != nil {
				t.Fatalf("Failed to subscribe with pattern: %v", err)
			}
			defer bus.Unsubscribe(subID)

			// Publish event
			event := domain.NewEvent(tt.eventType, "agent1", "TestAgent", nil)
			bus.Publish(event)

			// Check if event was received
			select {
			case <-received:
				if !tt.shouldMatch {
					t.Error("Event received but should not match pattern")
				}
			case <-time.After(50 * time.Millisecond):
				if tt.shouldMatch {
					t.Error("Event not received but should match pattern")
				}
			}
		})
	}
}

func TestEventBus_Filters(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	received := make(chan domain.Event, 10)
	handler := EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		received <- event
		return nil
	})

	// Subscribe with type filter
	typeFilter := EventFilterFunc(func(event domain.Event) bool {
		return event.Type == domain.EventToolCall || event.Type == domain.EventToolResult
	})

	subID := bus.Subscribe(handler, typeFilter)
	defer bus.Unsubscribe(subID)

	// Publish various events
	events := []domain.Event{
		domain.NewEvent(domain.EventAgentStart, "agent1", "TestAgent", nil),
		domain.NewEvent(domain.EventToolCall, "agent1", "TestAgent", nil),
		domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", nil),
		domain.NewEvent(domain.EventToolResult, "agent1", "TestAgent", nil),
	}

	for _, e := range events {
		bus.Publish(e)
	}

	// Should receive only tool events
	timeout := time.After(100 * time.Millisecond)
	count := 0

Loop:
	for {
		select {
		case e := <-received:
			count++
			if e.Type != domain.EventToolCall && e.Type != domain.EventToolResult {
				t.Errorf("Received unexpected event type: %s", e.Type)
			}
		case <-timeout:
			break Loop
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 events, got %d", count)
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	received := make(chan domain.Event, 1)
	handler := EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		received <- event
		return nil
	})

	subID := bus.Subscribe(handler)

	// Unsubscribe
	bus.Unsubscribe(subID)

	// Publish event
	event := domain.NewEvent(domain.EventAgentStart, "agent1", "TestAgent", nil)
	bus.Publish(event)

	// Should not receive event
	select {
	case <-received:
		t.Error("Received event after unsubscribe")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

func TestEventBus_ConcurrentAccess(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var receivedCount int32
	var wg sync.WaitGroup

	// Create multiple subscribers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			handler := EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
				atomic.AddInt32(&receivedCount, 1)
				return nil
			})

			subID := bus.Subscribe(handler)
			defer bus.Unsubscribe(subID)

			// Keep subscription active
			time.Sleep(200 * time.Millisecond)
		}(i)
	}

	// Create multiple publishers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < 10; j++ {
				event := domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", map[string]int{
					"publisher": id,
					"event":     j,
				})
				bus.Publish(event)
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Give some time for all events to be processed
	time.Sleep(100 * time.Millisecond)

	// Each of 10 subscribers should receive all 50 events
	expected := int32(500)
	actual := atomic.LoadInt32(&receivedCount)

	// Allow some tolerance for timing issues
	if actual < expected-50 || actual > expected {
		t.Errorf("Expected around %d events received, got %d", expected, actual)
	}
}

func TestEventBus_BufferOverflow(t *testing.T) {
	// Create bus with small buffer
	bus := NewEventBus(WithBufferSize(2))
	defer bus.Close()

	slowHandler := EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		// Simulate slow processing
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	bus.Subscribe(slowHandler)

	// Publish many events quickly
	for i := 0; i < 10; i++ {
		event := domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", i)
		bus.Publish(event)
	}

	// Should not block or panic
	// Some events may be dropped due to buffer overflow
}

func TestEventBus_Close(t *testing.T) {
	bus := NewEventBus()

	var wg sync.WaitGroup
	received := make(chan bool, 1)

	handler := EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		wg.Add(1)
		defer wg.Done()

		// Simulate processing
		time.Sleep(50 * time.Millisecond)
		received <- true
		return nil
	})

	bus.Subscribe(handler)

	// Publish event
	event := domain.NewEvent(domain.EventAgentStart, "agent1", "TestAgent", nil)
	bus.Publish(event)

	// Wait for handler to start processing
	<-received

	// Close bus
	bus.Close()

	// Should wait for handler to complete
	wg.Wait()

	// Try to publish after close
	bus.Publish(event) // Should not panic

	// Try to subscribe after close
	subID := bus.Subscribe(handler)
	if subID != "" {
		t.Error("Expected empty subscription ID after close")
	}
}

func TestEventBus_GetSubscriptionInfo(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	handler := EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		return nil
	})

	// Subscribe with pattern
	subID, err := bus.SubscribePattern("tool.*", handler,
		NewTypeFilter(domain.EventToolCall, domain.EventToolResult),
		NewAgentFilter("agent1", ""))

	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	pattern, filterCount, found := bus.GetSubscriptionInfo(subID)
	if !found {
		t.Fatal("Subscription not found")
	}

	if pattern != "tool.*" {
		t.Errorf("Expected pattern 'tool.*', got %s", pattern)
	}

	if filterCount != 2 {
		t.Errorf("Expected 2 filters, got %d", filterCount)
	}

	// Check non-existent subscription
	_, _, found = bus.GetSubscriptionInfo("invalid-id")
	if found {
		t.Error("Expected subscription not found")
	}
}
