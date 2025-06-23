// ABOUTME: Provides utility functions for event handling and processing
// ABOUTME: Includes helpers for event recording, filtering, and analysis

package utils

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// EventRecorder records events for later analysis.
// It provides a thread-safe circular buffer for event storage,
// automatically removing old events when capacity is reached.
type EventRecorder struct {
	mu      sync.RWMutex
	events  []domain.Event
	maxSize int
}

// NewEventRecorder creates a new event recorder with the specified capacity.
// If maxSize is <= 0, it defaults to 1000 events.
//
// Parameters:
//   - maxSize: Maximum number of events to store
//
// Returns a new EventRecorder instance.
func NewEventRecorder(maxSize int) *EventRecorder {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &EventRecorder{
		events:  make([]domain.Event, 0, maxSize),
		maxSize: maxSize,
	}
}

// HandleEvent implements domain.EventHandler interface.
// It stores the event in the circular buffer, removing the oldest
// event if at capacity.
//
// Parameters:
//   - event: The event to record
//
// Returns nil (always succeeds).
func (r *EventRecorder) HandleEvent(event domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If we're at capacity, remove oldest event
	if len(r.events) >= r.maxSize {
		r.events = r.events[1:]
	}

	r.events = append(r.events, event)
	return nil
}

// GetEvents returns a copy of recorded events.
// The returned slice is a copy to prevent external modifications.
//
// Returns all recorded events in chronological order.
func (r *EventRecorder) GetEvents() []domain.Event {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]domain.Event, len(r.events))
	copy(events, r.events)
	return events
}

// GetEventsByType returns events of a specific type.
// This is useful for analyzing specific event categories.
//
// Parameters:
//   - eventType: The event type to filter by
//
// Returns filtered events of the specified type.
func (r *EventRecorder) GetEventsByType(eventType domain.EventType) []domain.Event {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Event
	for _, event := range r.events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// GetEventsByAgent returns events for a specific agent.
// This helps analyze the behavior of individual agents.
//
// Parameters:
//   - agentID: The agent ID to filter by
//
// Returns events from the specified agent.
func (r *EventRecorder) GetEventsByAgent(agentID string) []domain.Event {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Event
	for _, event := range r.events {
		if event.AgentID == agentID {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// Clear removes all recorded events.
// This resets the recorder to an empty state.
func (r *EventRecorder) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = r.events[:0]
}

// EventAnalyzer analyzes recorded events to extract insights.
// It provides various analysis methods for understanding event patterns,
// agent behavior, and system performance.
type EventAnalyzer struct {
	events []domain.Event
}

// NewEventAnalyzer creates a new event analyzer for the given events.
//
// Parameters:
//   - events: The events to analyze
//
// Returns a new EventAnalyzer instance.
func NewEventAnalyzer(events []domain.Event) *EventAnalyzer {
	return &EventAnalyzer{
		events: events,
	}
}

// GetAgentMetrics returns metrics for each agent.
// It calculates event counts, error rates, and execution times
// for each agent found in the events.
//
// Returns a map of agent ID to metrics.
func (a *EventAnalyzer) GetAgentMetrics() map[string]*AgentMetrics {
	metrics := make(map[string]*AgentMetrics)

	for _, event := range a.events {
		m, ok := metrics[event.AgentID]
		if !ok {
			m = &AgentMetrics{
				AgentID:   event.AgentID,
				AgentName: event.AgentName,
				Events:    make(map[domain.EventType]int),
			}
			metrics[event.AgentID] = m
		}

		m.TotalEvents++
		m.Events[event.Type]++

		if event.IsError() {
			m.ErrorCount++
		}

		// Track timing for start/complete events
		switch event.Type {
		case domain.EventAgentStart:
			m.StartTime = event.Timestamp
		case domain.EventAgentComplete:
			if !m.StartTime.IsZero() {
				m.Duration = event.Timestamp.Sub(m.StartTime)
			}
		}
	}

	return metrics
}

// GetEventTimeline returns events organized by time.
// Events are grouped by second for easier visualization
// and analysis of temporal patterns.
//
// Returns timeline entries sorted by timestamp.
func (a *EventAnalyzer) GetEventTimeline() []TimelineEntry {
	if len(a.events) == 0 {
		return nil
	}

	// Group events by second
	timeline := make(map[int64][]domain.Event)
	for _, event := range a.events {
		key := event.Timestamp.Unix()
		timeline[key] = append(timeline[key], event)
	}

	// Convert to sorted timeline
	var entries []TimelineEntry
	for timestamp, events := range timeline {
		entries = append(entries, TimelineEntry{
			Timestamp: time.Unix(timestamp, 0),
			Events:    events,
		})
	}

	// Sort by timestamp
	sortTimeline(entries)

	return entries
}

// GetErrorSummary returns a summary of errors.
// It aggregates error information by type, agent, and message
// to help identify common failure patterns.
//
// Returns an ErrorSummary with detailed error statistics.
func (a *EventAnalyzer) GetErrorSummary() *ErrorSummary {
	summary := &ErrorSummary{
		ErrorsByType:  make(map[domain.EventType]int),
		ErrorsByAgent: make(map[string]int),
		ErrorMessages: make(map[string]int),
	}

	for _, event := range a.events {
		if event.IsError() {
			summary.TotalErrors++
			summary.ErrorsByType[event.Type]++
			summary.ErrorsByAgent[event.AgentID]++

			if event.Error != nil {
				msg := event.Error.Error()
				summary.ErrorMessages[msg]++
			}
		}
	}

	return summary
}

// AgentMetrics contains metrics for a single agent.
// It tracks event counts, errors, and execution timing
// to measure agent performance and reliability.
type AgentMetrics struct {
	AgentID     string
	AgentName   string
	TotalEvents int
	ErrorCount  int
	Events      map[domain.EventType]int
	StartTime   time.Time
	Duration    time.Duration
}

// TimelineEntry represents events at a point in time.
// It groups events that occurred within the same time window
// for timeline visualization.
type TimelineEntry struct {
	Timestamp time.Time
	Events    []domain.Event
}

// ErrorSummary contains error statistics.
// It provides multiple views of error data to help
// identify and diagnose system issues.
type ErrorSummary struct {
	TotalErrors   int
	ErrorsByType  map[domain.EventType]int
	ErrorsByAgent map[string]int
	ErrorMessages map[string]int
}

// EventFormatter formats events for display.
// It provides customizable formatting options for
// human-readable event presentation.
type EventFormatter struct {
	IncludeData     bool
	IncludeMetadata bool
	TimeFormat      string
}

// NewEventFormatter creates a new event formatter with default settings.
// The default time format is "15:04:05.000" (HH:MM:SS.mmm).
//
// Returns a new EventFormatter instance.
func NewEventFormatter() *EventFormatter {
	return &EventFormatter{
		TimeFormat: "15:04:05.000",
	}
}

// Format formats a single event to a human-readable string.
// The output includes timestamp, type, agent info, and optionally
// data and metadata based on formatter settings.
//
// Parameters:
//   - event: The event to format
//
// Returns the formatted event string.
func (f *EventFormatter) Format(event domain.Event) string {
	timestamp := event.Timestamp.Format(f.TimeFormat)
	base := fmt.Sprintf("[%s] %s - %s (%s)", timestamp, event.Type, event.AgentName, event.AgentID[:8])

	if event.Error != nil {
		base += fmt.Sprintf(" ERROR: %v", event.Error)
	}

	if f.IncludeData && event.Data != nil {
		data, _ := json.MarshalIndent(event.Data, "  ", "  ")
		base += fmt.Sprintf("\n  Data: %s", string(data))
	}

	if f.IncludeMetadata && len(event.Metadata) > 0 {
		metadata, _ := json.MarshalIndent(event.Metadata, "  ", "  ")
		base += fmt.Sprintf("\n  Metadata: %s", string(metadata))
	}

	return base
}

// FormatMultiple formats multiple events.
// Each event is formatted on a separate line.
//
// Parameters:
//   - events: The events to format
//
// Returns the formatted events as a multi-line string.
func (f *EventFormatter) FormatMultiple(events []domain.Event) string {
	var result string
	for i, event := range events {
		if i > 0 {
			result += "\n"
		}
		result += f.Format(event)
	}
	return result
}

// EventMatcher provides complex event matching capabilities.
// It supports multiple criteria including type, agent, time range,
// and custom data matching functions.
type EventMatcher struct {
	Type       *domain.EventType
	AgentID    *string
	AgentName  *string
	HasError   *bool
	TimeAfter  *time.Time
	TimeBefore *time.Time
	DataMatch  func(interface{}) bool
}

// Matches checks if an event matches the criteria.
// All non-nil criteria must match for the event to be selected.
//
// Parameters:
//   - event: The event to check
//
// Returns true if the event matches all criteria.
func (m *EventMatcher) Matches(event domain.Event) bool {
	if m.Type != nil && event.Type != *m.Type {
		return false
	}

	if m.AgentID != nil && event.AgentID != *m.AgentID {
		return false
	}

	if m.AgentName != nil && event.AgentName != *m.AgentName {
		return false
	}

	if m.HasError != nil && event.IsError() != *m.HasError {
		return false
	}

	if m.TimeAfter != nil && event.Timestamp.Before(*m.TimeAfter) {
		return false
	}

	if m.TimeBefore != nil && event.Timestamp.After(*m.TimeBefore) {
		return false
	}

	if m.DataMatch != nil && !m.DataMatch(event.Data) {
		return false
	}

	return true
}

// FilterEvents filters events using a matcher.
// This is a convenience function for applying EventMatcher criteria
// to a slice of events.
//
// Parameters:
//   - events: The events to filter
//   - matcher: The matching criteria
//
// Returns events that match all criteria.
func FilterEvents(events []domain.Event, matcher *EventMatcher) []domain.Event {
	var filtered []domain.Event
	for _, event := range events {
		if matcher.Matches(event) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// EventAggregator aggregates events by different dimensions.
// It provides methods to group events for analysis and reporting.
type EventAggregator struct {
	events []domain.Event
}

// NewEventAggregator creates a new event aggregator.
//
// Parameters:
//   - events: The events to aggregate
//
// Returns a new EventAggregator instance.
func NewEventAggregator(events []domain.Event) *EventAggregator {
	return &EventAggregator{events: events}
}

// ByType groups events by type.
// This helps analyze the distribution of event types.
//
// Returns a map of event type to events.
func (a *EventAggregator) ByType() map[domain.EventType][]domain.Event {
	grouped := make(map[domain.EventType][]domain.Event)
	for _, event := range a.events {
		grouped[event.Type] = append(grouped[event.Type], event)
	}
	return grouped
}

// ByAgent groups events by agent.
// This helps analyze individual agent behavior.
//
// Returns a map of agent ID to events.
func (a *EventAggregator) ByAgent() map[string][]domain.Event {
	grouped := make(map[string][]domain.Event)
	for _, event := range a.events {
		grouped[event.AgentID] = append(grouped[event.AgentID], event)
	}
	return grouped
}

// ByTimeWindow groups events by time window.
// Events are sorted by time and grouped into fixed-size windows
// for temporal analysis.
//
// Parameters:
//   - windowSize: The duration of each time window
//
// Returns time windows with their events.
func (a *EventAggregator) ByTimeWindow(windowSize time.Duration) []TimeWindow {
	if len(a.events) == 0 {
		return nil
	}

	// Sort events by time
	sortedEvents := make([]domain.Event, len(a.events))
	copy(sortedEvents, a.events)
	sortEventsByTime(sortedEvents)

	var windows []TimeWindow
	var currentWindow *TimeWindow

	for _, event := range sortedEvents {
		if currentWindow == nil || event.Timestamp.Sub(currentWindow.Start) >= windowSize {
			if currentWindow != nil {
				windows = append(windows, *currentWindow)
			}
			currentWindow = &TimeWindow{
				Start:  event.Timestamp,
				End:    event.Timestamp.Add(windowSize),
				Events: []domain.Event{event},
			}
		} else {
			currentWindow.Events = append(currentWindow.Events, event)
		}
	}

	if currentWindow != nil {
		windows = append(windows, *currentWindow)
	}

	return windows
}

// TimeWindow represents a time window of events.
// It defines a time range and contains all events
// that occurred within that range.
type TimeWindow struct {
	Start  time.Time
	End    time.Time
	Events []domain.Event
}

// sortTimeline sorts timeline entries by timestamp in ascending order.
// Uses bubble sort for simplicity (suitable for small datasets).
func sortTimeline(entries []TimelineEntry) {
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].Timestamp.After(entries[j].Timestamp) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
}

// sortEventsByTime sorts events by timestamp in ascending order.
// Uses bubble sort for simplicity (suitable for small datasets).
func sortEventsByTime(events []domain.Event) {
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.After(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}
}
