// ABOUTME: Event testing utilities for capturing, filtering, and asserting events
// ABOUTME: Provides event timeline visualization and common event assertions

package helpers

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// EventCapture captures events for testing purposes
type EventCapture struct {
	events []domain.Event
	mu     sync.RWMutex
}

// NewEventCapture creates a new event capture
func NewEventCapture() *EventCapture {
	return &EventCapture{
		events: make([]domain.Event, 0),
	}
}

// Capture captures an event
func (ec *EventCapture) Capture(event domain.Event) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.events = append(ec.events, event)
}

// GetEvents returns all captured events
func (ec *EventCapture) GetEvents() []domain.Event {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	events := make([]domain.Event, len(ec.events))
	copy(events, ec.events)
	return events
}

// FilterByType filters events by type
func (ec *EventCapture) FilterByType(eventType domain.EventType) []domain.Event {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	var filtered []domain.Event
	for _, event := range ec.events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// FilterByData filters events by data content
func (ec *EventCapture) FilterByData(predicate func(data interface{}) bool) []domain.Event {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	var filtered []domain.Event
	for _, event := range ec.events {
		if predicate(event.Data) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// FilterByTimeRange filters events within a time range
func (ec *EventCapture) FilterByTimeRange(start, end time.Time) []domain.Event {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	var filtered []domain.Event
	for _, event := range ec.events {
		if event.Timestamp.After(start) && event.Timestamp.Before(end) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// Clear clears all captured events
func (ec *EventCapture) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.events = make([]domain.Event, 0)
}

// EventAssertion provides fluent event assertions
type EventAssertion struct {
	events []domain.Event
	errors []string
}

// AssertEvents creates a new event assertion
func AssertEvents(events []domain.Event) *EventAssertion {
	return &EventAssertion{
		events: events,
		errors: make([]string, 0),
	}
}

// HasCount asserts the event count
func (ea *EventAssertion) HasCount(expected int) *EventAssertion {
	if len(ea.events) != expected {
		ea.errors = append(ea.errors, fmt.Sprintf("expected %d events, got %d", expected, len(ea.events)))
	}
	return ea
}

// HasType asserts at least one event of the given type exists
func (ea *EventAssertion) HasType(eventType domain.EventType) *EventAssertion {
	found := false
	for _, event := range ea.events {
		if event.Type == eventType {
			found = true
			break
		}
	}
	if !found {
		ea.errors = append(ea.errors, fmt.Sprintf("no event of type %s found", eventType))
	}
	return ea
}

// HasTypeCount asserts the count of events of a specific type
func (ea *EventAssertion) HasTypeCount(eventType domain.EventType, expected int) *EventAssertion {
	count := 0
	for _, event := range ea.events {
		if event.Type == eventType {
			count++
		}
	}
	if count != expected {
		ea.errors = append(ea.errors, fmt.Sprintf("expected %d events of type %s, got %d", expected, eventType, count))
	}
	return ea
}

// InOrder asserts events occur in a specific order
func (ea *EventAssertion) InOrder(types ...domain.EventType) *EventAssertion {
	if len(types) > len(ea.events) {
		ea.errors = append(ea.errors, fmt.Sprintf("expected at least %d events for order check, got %d", len(types), len(ea.events)))
		return ea
	}

	typeIndex := 0
	for _, event := range ea.events {
		if typeIndex < len(types) && event.Type == types[typeIndex] {
			typeIndex++
		}
	}

	if typeIndex != len(types) {
		ea.errors = append(ea.errors, fmt.Sprintf("events not in expected order: %v", types))
	}
	return ea
}

// NoErrors asserts no error events occurred
func (ea *EventAssertion) NoErrors() *EventAssertion {
	errorTypes := []domain.EventType{
		domain.EventAgentError,
		domain.EventToolError,
	}

	for _, event := range ea.events {
		for _, errorType := range errorTypes {
			if event.Type == errorType {
				ea.errors = append(ea.errors, fmt.Sprintf("found error event: %s", event.Type))
			}
		}
		if event.Error != nil {
			ea.errors = append(ea.errors, fmt.Sprintf("event %s has error: %v", event.Type, event.Error))
		}
	}
	return ea
}

// WithData asserts at least one event has the expected data
func (ea *EventAssertion) WithData(key string, value interface{}) *EventAssertion {
	found := false
	for _, event := range ea.events {
		if data, ok := event.Data.(map[string]interface{}); ok {
			if v, exists := data[key]; exists && v == value {
				found = true
				break
			}
		}
	}
	if !found {
		ea.errors = append(ea.errors, fmt.Sprintf("no event found with data[%s] = %v", key, value))
	}
	return ea
}

// GetErrors returns all assertion errors
func (ea *EventAssertion) GetErrors() []string {
	return ea.errors
}

// IsValid returns true if no assertion errors occurred
func (ea *EventAssertion) IsValid() bool {
	return len(ea.errors) == 0
}

// String returns a string representation of all errors
func (ea *EventAssertion) String() string {
	if ea.IsValid() {
		return "All event assertions passed"
	}
	return "Event assertion failures:\n" + strings.Join(ea.errors, "\n")
}

// EventTimeline provides a visual representation of events
type EventTimeline struct {
	events []domain.Event
	start  time.Time
	end    time.Time
}

// NewEventTimeline creates a new event timeline
func NewEventTimeline(events []domain.Event) *EventTimeline {
	if len(events) == 0 {
		return &EventTimeline{events: events}
	}

	start := events[0].Timestamp
	end := events[0].Timestamp

	for _, event := range events {
		if event.Timestamp.Before(start) {
			start = event.Timestamp
		}
		if event.Timestamp.After(end) {
			end = event.Timestamp
		}
	}

	return &EventTimeline{
		events: events,
		start:  start,
		end:    end,
	}
}

// Visualize returns a string visualization of the event timeline
func (et *EventTimeline) Visualize() string {
	if len(et.events) == 0 {
		return "No events to visualize"
	}

	var sb strings.Builder
	sb.WriteString("Event Timeline:\n")
	sb.WriteString("==============\n\n")

	duration := et.end.Sub(et.start)

	for _, event := range et.events {
		elapsed := event.Timestamp.Sub(et.start)
		percentage := float64(elapsed) / float64(duration) * 100

		// Create visual bar
		barLength := int(percentage / 5) // 20 character bar max
		if barLength < 0 {
			barLength = 0
		}
		if barLength > 20 {
			barLength = 20
		}

		bar := strings.Repeat("█", barLength) + strings.Repeat("░", 20-barLength)

		sb.WriteString(fmt.Sprintf("[%s] %s %s\n",
			bar,
			event.Timestamp.Format("15:04:05.000"),
			event.Type,
		))

		// Add event details
		if event.Error != nil {
			sb.WriteString(fmt.Sprintf("  └─ Error: %v\n", event.Error))
		}
		if data, ok := event.Data.(map[string]interface{}); ok && len(data) > 0 {
			sb.WriteString("  └─ Data: ")
			first := true
			for k, v := range data {
				if !first {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%s=%v", k, v))
				first = false
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Total duration: %v\n", duration))
	sb.WriteString(fmt.Sprintf("Total events: %d\n", len(et.events)))

	return sb.String()
}

// GroupByType groups events by their type
func (et *EventTimeline) GroupByType() map[domain.EventType][]domain.Event {
	groups := make(map[domain.EventType][]domain.Event)

	for _, event := range et.events {
		groups[event.Type] = append(groups[event.Type], event)
	}

	return groups
}

// GetDuration returns the duration of the timeline
func (et *EventTimeline) GetDuration() time.Duration {
	if len(et.events) == 0 {
		return 0
	}
	return et.end.Sub(et.start)
}

// Common event creation helpers

// CreateProgressEvent creates a progress event for testing
func CreateProgressEvent(agentID, agentName string, current, total int, message string) domain.Event {
	return domain.NewEvent(
		domain.EventProgress,
		agentID,
		agentName,
		domain.ProgressEventData{
			Current: current,
			Total:   total,
			Message: message,
		},
	)
}

// CreateToolCallEvent creates a tool call event for testing
func CreateToolCallEvent(agentID, agentName, toolName string, input interface{}) domain.Event {
	return domain.NewEvent(
		domain.EventToolCall,
		agentID,
		agentName,
		map[string]interface{}{
			"tool":  toolName,
			"input": input,
		},
	)
}

// CreateErrorEvent creates an error event for testing
func CreateErrorEvent(agentID, agentName string, err error) domain.Event {
	return domain.NewEvent(
		domain.EventAgentError,
		agentID,
		agentName,
		nil,
	).WithError(err)
}
