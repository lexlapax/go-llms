// ABOUTME: Defines the event system for agent execution monitoring and hooks
// ABOUTME: Provides event types, handlers, and filtering capabilities for agent lifecycle

package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event represents an event during agent execution
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	AgentID   string                 `json:"agent_id"`
	AgentName string                 `json:"agent_name"`
	Timestamp time.Time              `json:"timestamp"`
	Data      interface{}            `json:"data,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Error     error                  `json:"error,omitempty"`
}

// EventType represents the type of event
type EventType string

const (
	// Lifecycle events
	EventAgentStart    EventType = "agent.start"
	EventAgentComplete EventType = "agent.complete"
	EventAgentError    EventType = "agent.error"

	// Execution events
	EventStateUpdate EventType = "state.update"
	EventProgress    EventType = "progress"
	EventMessage     EventType = "message"

	// Tool events
	EventToolCall   EventType = "tool.call"
	EventToolResult EventType = "tool.result"
	EventToolError  EventType = "tool.error"

	// Workflow events
	EventSubAgentStart EventType = "subagent.start"
	EventSubAgentEnd   EventType = "subagent.end"
	EventWorkflowStep  EventType = "workflow.step"
)

// NewEvent creates a new event
func NewEvent(eventType EventType, agentID, agentName string, data interface{}) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		AgentID:   agentID,
		AgentName: agentName,
		Timestamp: time.Now(),
		Data:      data,
		Metadata:  make(map[string]interface{}),
	}
}

// WithError adds an error to the event
func (e Event) WithError(err error) Event {
	e.Error = err
	return e
}

// WithMetadata adds metadata to the event
func (e Event) WithMetadata(key string, value interface{}) Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// IsError returns true if the event represents an error
func (e Event) IsError() bool {
	return e.Error != nil || e.Type == EventAgentError || e.Type == EventToolError
}

// EventData types for specific events

// ProgressEventData represents progress information
type ProgressEventData struct {
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Message string `json:"message"`
}

// ToolCallEventData represents tool call information
type ToolCallEventData struct {
	ToolName   string      `json:"tool_name"`
	Parameters interface{} `json:"parameters"`
	RequestID  string      `json:"request_id"`
}

// ToolResultEventData represents tool result information
type ToolResultEventData struct {
	ToolName  string        `json:"tool_name"`
	Result    interface{}   `json:"result"`
	RequestID string        `json:"request_id"`
	Duration  time.Duration `json:"duration"`
}

// StateUpdateEventData represents state update information
type StateUpdateEventData struct {
	Key      string      `json:"key"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value"`
	Action   string      `json:"action"` // set, delete, merge
}

// WorkflowStepEventData represents workflow step information
type WorkflowStepEventData struct {
	StepName    string `json:"step_name"`
	StepIndex   int    `json:"step_index"`
	TotalSteps  int    `json:"total_steps"`
	Description string `json:"description,omitempty"`
}

// AgentStartEventData represents agent start information
type AgentStartEventData struct {
	InputState  *State                 `json:"input_state,omitempty"`
	Config      AgentConfig            `json:"config,omitempty"`
	ParentAgent string                 `json:"parent_agent,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AgentCompleteEventData represents agent completion information
type AgentCompleteEventData struct {
	OutputState *State                 `json:"output_state,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// EventHandler processes events
type EventHandler interface {
	HandleEvent(event Event) error
}

// EventHandlerFunc is a function adapter for EventHandler
type EventHandlerFunc func(event Event) error

// HandleEvent implements EventHandler
func (f EventHandlerFunc) HandleEvent(event Event) error {
	return f(event)
}

// EventFilter filters events
type EventFilter func(event Event) bool

// Common event filters

// FilterByType returns a filter that matches events by type
func FilterByType(eventTypes ...EventType) EventFilter {
	typeMap := make(map[EventType]bool)
	for _, t := range eventTypes {
		typeMap[t] = true
	}
	return func(event Event) bool {
		return typeMap[event.Type]
	}
}

// FilterByAgent returns a filter that matches events by agent ID
func FilterByAgent(agentID string) EventFilter {
	return func(event Event) bool {
		return event.AgentID == agentID
	}
}

// FilterByAgentName returns a filter that matches events by agent name
func FilterByAgentName(agentName string) EventFilter {
	return func(event Event) bool {
		return event.AgentName == agentName
	}
}

// FilterErrors returns a filter that matches error events
func FilterErrors() EventFilter {
	return func(event Event) bool {
		return event.IsError()
	}
}

// CombineFilters combines multiple filters with AND logic
func CombineFilters(filters ...EventFilter) EventFilter {
	return func(event Event) bool {
		for _, filter := range filters {
			if !filter(event) {
				return false
			}
		}
		return true
	}
}

// EventDispatcher manages event distribution
type EventDispatcher interface {
	// Subscribe adds a handler with optional filters
	Subscribe(handler EventHandler, filters ...EventFilter) string

	// Unsubscribe removes a subscription
	Unsubscribe(subscriptionID string)

	// Dispatch sends an event to all matching subscribers
	Dispatch(event Event)

	// Close shuts down the dispatcher
	Close()
}

// EventStream represents a stream of events
type EventStream interface {
	// Next returns the next event or blocks until one is available
	Next() (Event, error)

	// Close closes the stream
	Close()
}

// MarshalJSON customizes JSON marshaling for Event
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	var errStr string
	if e.Error != nil {
		errStr = e.Error.Error()
	}
	return json.Marshal(&struct {
		*Alias
		Error string `json:"error,omitempty"`
	}{
		Alias: (*Alias)(&e),
		Error: errStr,
	})
}
