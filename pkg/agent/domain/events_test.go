// ABOUTME: Tests for the event system including event creation, filtering, and marshaling
// ABOUTME: Validates event types, error handling, and event data structures

package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestNewEvent(t *testing.T) {
	event := domain.NewEvent(
		domain.EventAgentStart,
		"agent-123",
		"TestAgent",
		map[string]string{"key": "value"},
	)

	if event.ID == "" {
		t.Error("Event ID should not be empty")
	}

	if event.Type != domain.EventAgentStart {
		t.Errorf("Expected type %s, got %s", domain.EventAgentStart, event.Type)
	}

	if event.AgentID != "agent-123" {
		t.Errorf("Expected agent ID 'agent-123', got %s", event.AgentID)
	}

	if event.AgentName != "TestAgent" {
		t.Errorf("Expected agent name 'TestAgent', got %s", event.AgentName)
	}

	if event.Data == nil {
		t.Error("Event data should not be nil")
	}

	if event.Metadata == nil {
		t.Error("Event metadata should be initialized")
	}

	if event.Timestamp.IsZero() {
		t.Error("Event timestamp should be set")
	}
}

func TestEventWithError(t *testing.T) {
	err := errors.New("test error")
	event := domain.NewEvent(
		domain.EventAgentError,
		"agent-123",
		"TestAgent",
		nil,
	).WithError(err)

	if event.Error != err {
		t.Error("Error should be set")
	}

	if !event.IsError() {
		t.Error("IsError should return true")
	}
}

func TestEventWithMetadata(t *testing.T) {
	event := domain.NewEvent(
		domain.EventProgress,
		"agent-123",
		"TestAgent",
		nil,
	).WithMetadata("key1", "value1").
		WithMetadata("key2", 42)

	if val, ok := event.Metadata["key1"]; !ok || val != "value1" {
		t.Error("Metadata key1 not set correctly")
	}

	if val, ok := event.Metadata["key2"]; !ok || val != 42 {
		t.Error("Metadata key2 not set correctly")
	}
}

func TestEventIsError(t *testing.T) {
	tests := []struct {
		name     string
		event    domain.Event
		expected bool
	}{
		{
			name:     "EventAgentError",
			event:    domain.NewEvent(domain.EventAgentError, "id", "name", nil),
			expected: true,
		},
		{
			name:     "EventToolError",
			event:    domain.NewEvent(domain.EventToolError, "id", "name", nil),
			expected: true,
		},
		{
			name: "Event with error",
			event: domain.NewEvent(domain.EventAgentStart, "id", "name", nil).
				WithError(errors.New("error")),
			expected: true,
		},
		{
			name:     "Normal event",
			event:    domain.NewEvent(domain.EventAgentStart, "id", "name", nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.event.IsError() != tt.expected {
				t.Errorf("Expected IsError=%v, got %v", tt.expected, tt.event.IsError())
			}
		})
	}
}

func TestEventDataTypes(t *testing.T) {
	t.Run("ProgressEventData", func(t *testing.T) {
		data := domain.ProgressEventData{
			Current: 50,
			Total:   100,
			Message: "Processing...",
		}

		event := domain.NewEvent(domain.EventProgress, "agent-123", "TestAgent", data)

		// Verify data is stored correctly
		if eventData, ok := event.Data.(domain.ProgressEventData); ok {
			if eventData.Current != 50 || eventData.Total != 100 {
				t.Error("Progress data not stored correctly")
			}
		} else {
			t.Error("Failed to cast event data to ProgressEventData")
		}
	})

	t.Run("ToolCallEventData", func(t *testing.T) {
		data := domain.ToolCallEventData{
			ToolName:   "Calculator",
			Parameters: map[string]interface{}{"operation": "add", "a": 1, "b": 2},
			RequestID:  "req-123",
		}

		event := domain.NewEvent(domain.EventToolCall, "agent-123", "TestAgent", data)

		if eventData, ok := event.Data.(domain.ToolCallEventData); ok {
			if eventData.ToolName != "Calculator" {
				t.Error("Tool name not stored correctly")
			}
			if eventData.RequestID != "req-123" {
				t.Error("Request ID not stored correctly")
			}
		} else {
			t.Error("Failed to cast event data to ToolCallEventData")
		}
	})

	t.Run("StateUpdateEventData", func(t *testing.T) {
		data := domain.StateUpdateEventData{
			Key:      "config",
			OldValue: "old",
			NewValue: "new",
			Action:   "set",
		}

		event := domain.NewEvent(domain.EventStateUpdate, "agent-123", "TestAgent", data)

		if eventData, ok := event.Data.(domain.StateUpdateEventData); ok {
			if eventData.Key != "config" || eventData.Action != "set" {
				t.Error("State update data not stored correctly")
			}
		} else {
			t.Error("Failed to cast event data to StateUpdateEventData")
		}
	})
}

func TestEventFilters(t *testing.T) {
	// Create test events
	events := []domain.Event{
		domain.NewEvent(domain.EventAgentStart, "agent-1", "Agent1", nil),
		domain.NewEvent(domain.EventAgentComplete, "agent-1", "Agent1", nil),
		domain.NewEvent(domain.EventAgentStart, "agent-2", "Agent2", nil),
		domain.NewEvent(domain.EventToolCall, "agent-1", "Agent1", nil),
		domain.NewEvent(domain.EventAgentError, "agent-2", "Agent2", nil).
			WithError(errors.New("test error")),
	}

	t.Run("FilterByType", func(t *testing.T) {
		filter := domain.FilterByType(domain.EventAgentStart)

		count := 0
		for _, event := range events {
			if filter(event) {
				count++
				if event.Type != domain.EventAgentStart {
					t.Error("Filter should only match EventAgentStart")
				}
			}
		}

		if count != 2 {
			t.Errorf("Expected 2 matches, got %d", count)
		}
	})

	t.Run("FilterByAgent", func(t *testing.T) {
		filter := domain.FilterByAgent("agent-1")

		count := 0
		for _, event := range events {
			if filter(event) {
				count++
				if event.AgentID != "agent-1" {
					t.Error("Filter should only match agent-1")
				}
			}
		}

		if count != 3 {
			t.Errorf("Expected 3 matches, got %d", count)
		}
	})

	t.Run("FilterByAgentName", func(t *testing.T) {
		filter := domain.FilterByAgentName("Agent2")

		count := 0
		for _, event := range events {
			if filter(event) {
				count++
				if event.AgentName != "Agent2" {
					t.Error("Filter should only match Agent2")
				}
			}
		}

		if count != 2 {
			t.Errorf("Expected 2 matches, got %d", count)
		}
	})

	t.Run("FilterErrors", func(t *testing.T) {
		filter := domain.FilterErrors()

		count := 0
		for _, event := range events {
			if filter(event) {
				count++
				if !event.IsError() {
					t.Error("Filter should only match error events")
				}
			}
		}

		if count != 1 {
			t.Errorf("Expected 1 match, got %d", count)
		}
	})

	t.Run("CombineFilters", func(t *testing.T) {
		// Combine filters: agent-1 AND (start OR complete)
		filter := domain.CombineFilters(
			domain.FilterByAgent("agent-1"),
			domain.FilterByType(domain.EventAgentStart, domain.EventAgentComplete),
		)

		count := 0
		for _, event := range events {
			if filter(event) {
				count++
				if event.AgentID != "agent-1" {
					t.Error("Combined filter should match agent-1")
				}
				if event.Type != domain.EventAgentStart && event.Type != domain.EventAgentComplete {
					t.Error("Combined filter should match start or complete events")
				}
			}
		}

		if count != 2 {
			t.Errorf("Expected 2 matches, got %d", count)
		}
	})
}

func TestEventHandlerFunc(t *testing.T) {
	called := false
	var receivedEvent domain.Event

	handler := domain.EventHandlerFunc(func(event domain.Event) error {
		called = true
		receivedEvent = event
		return nil
	})

	event := domain.NewEvent(domain.EventAgentStart, "agent-123", "TestAgent", nil)
	err := handler.HandleEvent(event)

	if err != nil {
		t.Errorf("Handler returned error: %v", err)
	}

	if !called {
		t.Error("Handler function was not called")
	}

	if receivedEvent.ID != event.ID {
		t.Error("Handler received different event")
	}
}

func TestEventMarshalJSON(t *testing.T) {
	// Test event without error
	event1 := domain.NewEvent(
		domain.EventAgentStart,
		"agent-123",
		"TestAgent",
		map[string]string{"key": "value"},
	).WithMetadata("meta", "data")

	data, err := json.Marshal(event1)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	// Verify fields
	if unmarshaled["id"] != event1.ID {
		t.Error("ID not marshaled correctly")
	}

	if unmarshaled["type"] != string(event1.Type) {
		t.Error("Type not marshaled correctly")
	}

	if unmarshaled["agent_id"] != event1.AgentID {
		t.Error("AgentID not marshaled correctly")
	}

	// Test event with error
	event2 := domain.NewEvent(
		domain.EventAgentError,
		"agent-123",
		"TestAgent",
		nil,
	).WithError(errors.New("test error"))

	data2, err := json.Marshal(event2)
	if err != nil {
		t.Fatalf("Failed to marshal event with error: %v", err)
	}

	var unmarshaled2 map[string]interface{}
	err = json.Unmarshal(data2, &unmarshaled2)
	if err != nil {
		t.Fatalf("Failed to unmarshal event with error: %v", err)
	}

	if unmarshaled2["error"] != "test error" {
		t.Error("Error not marshaled correctly")
	}
}

func TestEventTypes(t *testing.T) {
	// Test that all event types are unique
	eventTypes := []domain.EventType{
		domain.EventAgentStart,
		domain.EventAgentComplete,
		domain.EventAgentError,
		domain.EventStateUpdate,
		domain.EventProgress,
		domain.EventMessage,
		domain.EventToolCall,
		domain.EventToolResult,
		domain.EventToolError,
		domain.EventSubAgentStart,
		domain.EventSubAgentEnd,
		domain.EventWorkflowStep,
	}

	seen := make(map[domain.EventType]bool)
	for _, et := range eventTypes {
		if seen[et] {
			t.Errorf("Duplicate event type: %s", et)
		}
		seen[et] = true
	}

	// Verify string values are descriptive
	for _, et := range eventTypes {
		if string(et) == "" {
			t.Error("Event type should not be empty string")
		}
		if len(string(et)) < 5 {
			t.Errorf("Event type %s seems too short", et)
		}
	}
}

func TestComplexEventData(t *testing.T) {
	// Test with complex nested data
	complexData := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"array": []interface{}{1, 2, 3},
				"bool":  true,
				"str":   "nested",
			},
		},
		"number": 42.5,
		"null":   nil,
	}

	event := domain.NewEvent(domain.EventProgress, "agent-123", "TestAgent", complexData)

	// Marshal and unmarshal
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal complex event: %v", err)
	}

	// We can't unmarshal directly to Event due to custom marshaling,
	// but we can verify the JSON structure
	var jsonData map[string]interface{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		t.Fatalf("Failed to unmarshal complex event: %v", err)
	}

	// Verify complex data is preserved
	if eventData, ok := jsonData["data"].(map[string]interface{}); ok {
		if level1, ok := eventData["level1"].(map[string]interface{}); ok {
			if level2, ok := level1["level2"].(map[string]interface{}); ok {
				if arr, ok := level2["array"].([]interface{}); ok {
					if len(arr) != 3 {
						t.Error("Array data not preserved correctly")
					}
				} else {
					t.Error("Array not found in unmarshaled data")
				}
			} else {
				t.Error("level2 not found in unmarshaled data")
			}
		} else {
			t.Error("level1 not found in unmarshaled data")
		}
	} else {
		t.Error("data field not found in unmarshaled event")
	}
}

// Benchmark tests
func BenchmarkNewEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = domain.NewEvent(
			domain.EventAgentStart,
			"agent-123",
			"TestAgent",
			map[string]string{"key": "value"},
		)
	}
}

func BenchmarkEventWithMetadata(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = domain.NewEvent(
			domain.EventAgentStart,
			"agent-123",
			"TestAgent",
			nil,
		).WithMetadata("key1", "value1").
			WithMetadata("key2", "value2").
			WithMetadata("key3", "value3")
	}
}

func BenchmarkEventFilter(b *testing.B) {
	events := make([]domain.Event, 1000)
	for i := 0; i < 1000; i++ {
		eventType := domain.EventAgentStart
		switch i % 3 {
		case 0:
			eventType = domain.EventAgentComplete
		case 1:
			eventType = domain.EventToolCall
		}

		events[i] = domain.NewEvent(
			eventType,
			sprintf("agent-%d", i%10),
			sprintf("Agent%d", i%10),
			nil,
		)
	}

	filter := domain.CombineFilters(
		domain.FilterByType(domain.EventAgentStart, domain.EventAgentComplete),
		domain.FilterByAgent("agent-5"),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		count := 0
		for _, event := range events {
			if filter(event) {
				count++
			}
		}
		_ = count
	}
}
