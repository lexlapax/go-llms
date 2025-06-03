// ABOUTME: Tests for the FunctionalEventStream interface with functional operations
// ABOUTME: including filter, map, reduce, and stream control operations

package domain

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestFunctionalEventStream_Filter(t *testing.T) {
	ctx := context.Background()

	// Create test events
	events := []Event{
		{Type: EventStateUpdate, AgentID: "agent1", Data: map[string]interface{}{"value": 1}},
		{Type: EventToolCall, AgentID: "agent2", Data: map[string]interface{}{"value": 2}},
		{Type: EventStateUpdate, AgentID: "agent1", Data: map[string]interface{}{"value": 3}},
		{Type: EventAgentStart, AgentID: "agent3", Data: map[string]interface{}{"value": 4}},
	}

	// Create a channel and stream manually to avoid immediate closure
	ch := make(chan Event)
	go func() {
		for _, e := range events {
			ch <- e
		}
		close(ch)
	}()

	stream := NewFunctionalEventStream(ctx, ch)

	// Filter by event type
	filtered := stream.Filter(func(e Event) bool {
		return e.Type == EventStateUpdate
	})

	result, err := filtered.Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Filter by type: expected 2 events, got %d", len(result))
	}

	// Filter by agent ID
	ch2 := make(chan Event)
	go func() {
		for _, e := range events {
			ch2 <- e
		}
		close(ch2)
	}()
	stream2 := NewFunctionalEventStream(ctx, ch2)
	filtered2 := stream2.Filter(func(e Event) bool {
		return e.AgentID == "agent1"
	})

	result2, err := filtered2.Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}
	if len(result2) != 2 {
		t.Errorf("Filter by agent: expected 2 events, got %d", len(result2))
	}

	// Chain filters
	ch3 := make(chan Event)
	go func() {
		for _, e := range events {
			ch3 <- e
		}
		close(ch3)
	}()
	stream3 := NewFunctionalEventStream(ctx, ch3)
	filtered3 := stream3.
		Filter(ByType(EventStateUpdate)).
		Filter(ByAgentID("agent1"))

	result3, err := filtered3.Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}
	if len(result3) != 2 {
		t.Errorf("Chained filters: expected 2 events, got %d", len(result3))
	}
}

func TestFunctionalEventStream_Map(t *testing.T) {
	ctx := context.Background()

	events := []Event{
		{Type: EventStateUpdate, Data: map[string]interface{}{"value": 1}},
		{Type: EventToolCall, Data: map[string]interface{}{"value": 2}},
		{Type: EventStateUpdate, Data: map[string]interface{}{"value": 3}},
	}

	ch := make(chan Event)
	go func() {
		for _, e := range events {
			ch <- e
		}
		close(ch)
	}()
	stream := NewFunctionalEventStream(ctx, ch)

	// Map to extract values
	mapped := stream.Map(func(e Event) Event {
		newEvent := e
		if data, ok := e.Data.(map[string]interface{}); ok {
			if val, ok := data["value"].(int); ok {
				newEvent.Data = map[string]interface{}{"doubled": val * 2}
			}
		}
		return newEvent
	})

	result, err := mapped.Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("Expected 3 mapped events, got %d", len(result))
	}

	// Check mapped values
	for i, event := range result {
		if data, ok := event.Data.(map[string]interface{}); ok {
			doubled, ok := data["doubled"].(int)
			if !ok {
				t.Errorf("Event %d missing doubled value", i)
				continue
			}
			expected := (i + 1) * 2
			if doubled != expected {
				t.Errorf("Event %d: doubled = %d, want %d", i, doubled, expected)
			}
		}
	}
}

func TestFunctionalEventStream_Reduce(t *testing.T) {
	ctx := context.Background()

	events := []Event{
		{Type: EventStateUpdate, Data: map[string]interface{}{"count": 5}},
		{Type: EventToolCall, Data: map[string]interface{}{"count": 3}},
		{Type: EventStateUpdate, Data: map[string]interface{}{"count": 7}},
	}

	stream := EventsFromSlice(ctx, events)

	// Sum all counts
	totalCount := stream.Reduce(func(acc interface{}, e Event) interface{} {
		sum := acc.(int)
		if data, ok := e.Data.(map[string]interface{}); ok {
			if count, ok := data["count"].(int); ok {
				sum += count
			}
		}
		return sum
	}, 0)

	if totalCount != 15 {
		t.Errorf("Reduce sum: expected 15, got %v", totalCount)
	}

	// Count by event type
	stream2 := EventsFromSlice(ctx, events)
	typeCounts := stream2.Reduce(func(acc interface{}, e Event) interface{} {
		counts := acc.(map[EventType]int)
		counts[e.Type]++
		return counts
	}, make(map[EventType]int)).(map[EventType]int)

	if typeCounts[EventStateUpdate] != 2 {
		t.Errorf("StateUpdate count: expected 2, got %d", typeCounts[EventStateUpdate])
	}
	if typeCounts[EventToolCall] != 1 {
		t.Errorf("ToolCall count: expected 1, got %d", typeCounts[EventToolCall])
	}
}

func TestFunctionalEventStream_Take(t *testing.T) {
	ctx := context.Background()

	events := make([]Event, 10)
	for i := range events {
		events[i] = Event{
			Type: EventMessage,
			Data: map[string]interface{}{"index": i},
		}
	}

	stream := EventsFromSlice(ctx, events)

	// Take first 3
	taken := stream.Take(3)
	result, err := taken.Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Take(3): expected 3 events, got %d", len(result))
	}

	// Verify correct events were taken
	for i, event := range result {
		if data, ok := event.Data.(map[string]interface{}); ok {
			if idx, ok := data["index"].(int); !ok || idx != i {
				t.Errorf("Take: event %d has wrong index %v", i, data["index"])
			}
		}
	}

	// Take more than available
	stream2 := EventsFromSlice(ctx, events[:5])
	taken2 := stream2.Take(10)
	result2, err := taken2.Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if len(result2) != 5 {
		t.Errorf("Take(10) from 5: expected 5 events, got %d", len(result2))
	}
}

func TestFunctionalEventStream_TakeUntil(t *testing.T) {
	ctx := context.Background()

	events := []Event{
		{Type: EventAgentStart, Data: map[string]interface{}{"value": 1}},
		{Type: EventStateUpdate, Data: map[string]interface{}{"value": 2}},
		{Type: EventToolCall, Data: map[string]interface{}{"value": 3}},
		{Type: EventAgentComplete, Data: map[string]interface{}{"value": 4}},
		{Type: EventStateUpdate, Data: map[string]interface{}{"value": 5}},
	}

	ch := make(chan Event)
	go func() {
		for _, e := range events {
			ch <- e
		}
		close(ch)
	}()
	stream := NewFunctionalEventStream(ctx, ch)

	// Take until completed event
	taken := stream.TakeUntil(func(e Event) bool {
		return e.Type == EventAgentComplete
	})

	result, err := taken.Collect()
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}
	if len(result) != 4 { // TakeUntil includes the final event
		t.Errorf("TakeUntil completed: expected 4 events (including final), got %d", len(result))
	}

	// Verify last event is the completed event
	if result[len(result)-1].Type != EventAgentComplete {
		t.Error("TakeUntil should include the stopping event")
	}
}

func TestFunctionalEventStream_ForEach(t *testing.T) {
	ctx := context.Background()

	events := []Event{
		{Type: EventStateUpdate, AgentID: "agent1"},
		{Type: EventToolCall, AgentID: "agent2"},
		{Type: EventStateUpdate, AgentID: "agent3"},
	}

	stream := EventsFromSlice(ctx, events)

	// Count events using ForEach
	count := 0
	agentIDs := []string{}

	err := stream.ForEach(EventHandlerFunc(func(e Event) error {
		count++
		agentIDs = append(agentIDs, e.AgentID)
		return nil
	}))
	if err != nil {
		t.Fatalf("ForEach failed: %v", err)
	}

	if count != 3 {
		t.Errorf("ForEach: expected to process 3 events, got %d", count)
	}

	// Verify all agent IDs were collected
	expectedIDs := []string{"agent1", "agent2", "agent3"}
	for i, id := range agentIDs {
		if id != expectedIDs[i] {
			t.Errorf("ForEach: agentID[%d] = %s, want %s", i, id, expectedIDs[i])
		}
	}
}

func TestFunctionalEventStream_Timeout(t *testing.T) {
	ctx := context.Background()

	// Create a channel-based stream that will deliver events over time
	eventChan := make(chan Event, 3)

	// Pre-fill with events
	eventChan <- Event{Type: EventAgentStart, Data: map[string]interface{}{"seq": 1}}
	eventChan <- Event{Type: EventStateUpdate, Data: map[string]interface{}{"seq": 2}}

	// Start a goroutine that will send the third event after a delay
	go func() {
		time.Sleep(150 * time.Millisecond)
		eventChan <- Event{Type: EventAgentComplete, Data: map[string]interface{}{"seq": 3}}
		close(eventChan)
	}()

	// Create stream with timeout
	timeout := 100 * time.Millisecond
	stream := NewFunctionalEventStream(ctx, eventChan)
	timeoutStream := stream.Timeout(timeout)

	result, err := timeoutStream.Collect()
	// Timeout should cause an error
	if err == nil {
		t.Error("Expected timeout error")
	}

	// Should have received the first 2 events before timeout
	if len(result) != 2 {
		t.Errorf("Timeout: expected 2 events before timeout, got %d", len(result))
	}
}

func TestFunctionalEventStream_Predicates(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	events := []Event{
		{Type: EventStateUpdate, AgentID: "agent1", Timestamp: now.Add(-2 * time.Minute)},
		{Type: EventToolCall, AgentID: "agent2", Timestamp: now.Add(-30 * time.Second)},
		{Type: EventStateUpdate, AgentID: "agent1", Timestamp: now},
		{Type: EventAgentError, AgentID: "agent3", Error: fmt.Errorf("test error"), Timestamp: now},
	}

	tests := []struct {
		name      string
		predicate EventPredicate
		expected  int
	}{
		{
			name:      "ByType",
			predicate: ByType(EventStateUpdate),
			expected:  2,
		},
		{
			name:      "ByAgentID",
			predicate: ByAgentID("agent1"),
			expected:  2,
		},
		{
			name:      "HasError",
			predicate: HasError,
			expected:  1,
		},
		{
			name:      "IsError",
			predicate: IsError,
			expected:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan Event)
			go func() {
				for _, e := range events {
					ch <- e
				}
				close(ch)
			}()
			stream := NewFunctionalEventStream(ctx, ch)
			filtered := stream.Filter(tt.predicate)
			result, err := filtered.Collect()
			if err != nil {
				t.Fatalf("Collect failed: %v", err)
			}

			if len(result) != tt.expected {
				t.Errorf("%s: expected %d events, got %d", tt.name, tt.expected, len(result))
			}
		})
	}
}

func TestFunctionalEventStream_ComplexExample_Old(t *testing.T) {
	t.Skip("This test uses removed event stream methods")
}
