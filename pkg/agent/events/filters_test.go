// ABOUTME: Tests for event filtering system
// ABOUTME: Validates pattern matching, composite filters, and field filtering

package events

import (
	"fmt"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestPatternFilter(t *testing.T) {
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
			name:        "wildcard suffix",
			pattern:     "tool.*",
			eventType:   domain.EventToolCall,
			shouldMatch: true,
		},
		{
			name:        "wildcard suffix no match",
			pattern:     "tool.*",
			eventType:   domain.EventAgentStart,
			shouldMatch: false,
		},
		{
			name:        "wildcard prefix",
			pattern:     "*.error",
			eventType:   domain.EventToolError,
			shouldMatch: true,
		},
		{
			name:        "workflow events",
			pattern:     "workflow.*",
			eventType:   domain.EventWorkflowStep,
			shouldMatch: true,
		},
		{
			name:        "all events",
			pattern:     "*",
			eventType:   domain.EventProgress,
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewPatternFilter(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to create pattern filter: %v", err)
			}

			event := domain.NewEvent(tt.eventType, "agent1", "TestAgent", nil)

			if filter.Match(event) != tt.shouldMatch {
				t.Errorf("Pattern %q match for %q = %v, want %v",
					tt.pattern, tt.eventType, filter.Match(event), tt.shouldMatch)
			}
		})
	}
}

func TestTypeFilter(t *testing.T) {
	filter := NewTypeFilter(
		domain.EventToolCall,
		domain.EventToolResult,
		domain.EventToolError,
	)

	tests := []struct {
		eventType   domain.EventType
		shouldMatch bool
	}{
		{domain.EventToolCall, true},
		{domain.EventToolResult, true},
		{domain.EventToolError, true},
		{domain.EventAgentStart, false},
		{domain.EventProgress, false},
	}

	for _, tt := range tests {
		event := domain.NewEvent(tt.eventType, "agent1", "TestAgent", nil)
		if filter.Match(event) != tt.shouldMatch {
			t.Errorf("Type filter match for %q = %v, want %v",
				tt.eventType, filter.Match(event), tt.shouldMatch)
		}
	}
}

func TestAgentFilter(t *testing.T) {
	tests := []struct {
		name        string
		filterID    string
		filterName  string
		eventID     string
		eventName   string
		shouldMatch bool
	}{
		{
			name:        "match by ID",
			filterID:    "agent1",
			filterName:  "",
			eventID:     "agent1",
			eventName:   "TestAgent",
			shouldMatch: true,
		},
		{
			name:        "no match by ID",
			filterID:    "agent1",
			filterName:  "",
			eventID:     "agent2",
			eventName:   "TestAgent",
			shouldMatch: false,
		},
		{
			name:        "match by name",
			filterID:    "",
			filterName:  "TestAgent",
			eventID:     "agent1",
			eventName:   "TestAgent",
			shouldMatch: true,
		},
		{
			name:        "match by both",
			filterID:    "agent1",
			filterName:  "TestAgent",
			eventID:     "agent1",
			eventName:   "TestAgent",
			shouldMatch: true,
		},
		{
			name:        "no match when both specified",
			filterID:    "agent1",
			filterName:  "TestAgent",
			eventID:     "agent1",
			eventName:   "OtherAgent",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewAgentFilter(tt.filterID, tt.filterName)
			event := domain.NewEvent(domain.EventAgentStart, tt.eventID, tt.eventName, nil)

			if filter.Match(event) != tt.shouldMatch {
				t.Errorf("Agent filter match = %v, want %v", filter.Match(event), tt.shouldMatch)
			}
		})
	}
}

func TestErrorFilter(t *testing.T) {
	filter := NewErrorFilter()

	tests := []struct {
		name        string
		event       domain.Event
		shouldMatch bool
	}{
		{
			name:        "error event type",
			event:       domain.NewEvent(domain.EventAgentError, "agent1", "TestAgent", nil),
			shouldMatch: true,
		},
		{
			name:        "tool error type",
			event:       domain.NewEvent(domain.EventToolError, "agent1", "TestAgent", nil),
			shouldMatch: true,
		},
		{
			name:        "event with error",
			event:       domain.NewEvent(domain.EventAgentComplete, "agent1", "TestAgent", nil).WithError(fmt.Errorf("timeout error")),
			shouldMatch: true,
		},
		{
			name:        "normal event",
			event:       domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", nil),
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if filter.Match(tt.event) != tt.shouldMatch {
				t.Errorf("Error filter match = %v, want %v", filter.Match(tt.event), tt.shouldMatch)
			}
		})
	}
}

func TestMetadataFilter(t *testing.T) {
	filter := NewMetadataFilter("environment", "production")

	tests := []struct {
		name        string
		metadata    map[string]interface{}
		shouldMatch bool
	}{
		{
			name:        "exact match",
			metadata:    map[string]interface{}{"environment": "production"},
			shouldMatch: true,
		},
		{
			name:        "no match value",
			metadata:    map[string]interface{}{"environment": "development"},
			shouldMatch: false,
		},
		{
			name:        "missing key",
			metadata:    map[string]interface{}{"other": "value"},
			shouldMatch: false,
		},
		{
			name:        "nil metadata",
			metadata:    nil,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", nil)
			event.Metadata = tt.metadata

			if filter.Match(event) != tt.shouldMatch {
				t.Errorf("Metadata filter match = %v, want %v", filter.Match(event), tt.shouldMatch)
			}
		})
	}
}

func TestCompositeFilter(t *testing.T) {
	// Create base filters
	toolFilter := NewTypeFilter(domain.EventToolCall, domain.EventToolResult)
	agentFilter := NewAgentFilter("agent1", "")
	errorFilter := NewErrorFilter()

	t.Run("AND operator", func(t *testing.T) {
		filter := AND(toolFilter, agentFilter)

		// Should match: tool event from agent1
		event1 := domain.NewEvent(domain.EventToolCall, "agent1", "TestAgent", nil)
		if !filter.Match(event1) {
			t.Error("Expected match for tool event from agent1")
		}

		// Should not match: tool event from different agent
		event2 := domain.NewEvent(domain.EventToolCall, "agent2", "TestAgent", nil)
		if filter.Match(event2) {
			t.Error("Expected no match for tool event from agent2")
		}

		// Should not match: non-tool event from agent1
		event3 := domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", nil)
		if filter.Match(event3) {
			t.Error("Expected no match for non-tool event from agent1")
		}
	})

	t.Run("OR operator", func(t *testing.T) {
		filter := OR(toolFilter, errorFilter)

		// Should match: tool event
		event1 := domain.NewEvent(domain.EventToolCall, "agent1", "TestAgent", nil)
		if !filter.Match(event1) {
			t.Error("Expected match for tool event")
		}

		// Should match: error event
		event2 := domain.NewEvent(domain.EventAgentError, "agent1", "TestAgent", nil)
		if !filter.Match(event2) {
			t.Error("Expected match for error event")
		}

		// Should not match: neither tool nor error
		event3 := domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", nil)
		if filter.Match(event3) {
			t.Error("Expected no match for progress event")
		}
	})

	t.Run("NOT operator", func(t *testing.T) {
		filter := NOT(errorFilter)

		// Should not match: error event
		event1 := domain.NewEvent(domain.EventAgentError, "agent1", "TestAgent", nil)
		if filter.Match(event1) {
			t.Error("Expected no match for error event")
		}

		// Should match: non-error event
		event2 := domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", nil)
		if !filter.Match(event2) {
			t.Error("Expected match for non-error event")
		}
	})

	t.Run("nested composite", func(t *testing.T) {
		// (tool events from agent1) OR (error events)
		filter := OR(
			AND(toolFilter, agentFilter),
			errorFilter,
		)

		// Should match: tool event from agent1
		event1 := domain.NewEvent(domain.EventToolCall, "agent1", "TestAgent", nil)
		if !filter.Match(event1) {
			t.Error("Expected match for tool event from agent1")
		}

		// Should match: any error event
		event2 := domain.NewEvent(domain.EventAgentError, "agent2", "TestAgent", nil)
		if !filter.Match(event2) {
			t.Error("Expected match for error event")
		}

		// Should not match: tool event from different agent
		event3 := domain.NewEvent(domain.EventToolCall, "agent2", "TestAgent", nil)
		if filter.Match(event3) {
			t.Error("Expected no match for tool event from agent2")
		}
	})
}

func TestFieldFilter(t *testing.T) {
	tests := []struct {
		name        string
		fieldPath   string
		operator    FieldOperator
		value       interface{}
		event       domain.Event
		shouldMatch bool
	}{
		{
			name:        "equal match",
			fieldPath:   "Type",
			operator:    OpEqual,
			value:       domain.EventToolCall,
			event:       domain.NewEvent(domain.EventToolCall, "agent1", "TestAgent", nil),
			shouldMatch: true,
		},
		{
			name:        "not equal match",
			fieldPath:   "Type",
			operator:    OpNotEqual,
			value:       domain.EventToolCall,
			event:       domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", nil),
			shouldMatch: true,
		},
		{
			name:        "contains match",
			fieldPath:   "AgentName",
			operator:    OpContains,
			value:       "Test",
			event:       domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", nil),
			shouldMatch: true,
		},
		{
			name:        "contains no match",
			fieldPath:   "AgentName",
			operator:    OpContains,
			value:       "Other",
			event:       domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", nil),
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFieldFilter(tt.fieldPath, tt.operator, tt.value)

			if filter.Match(tt.event) != tt.shouldMatch {
				t.Errorf("Field filter match = %v, want %v", filter.Match(tt.event), tt.shouldMatch)
			}
		})
	}
}
