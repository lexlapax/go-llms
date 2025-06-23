// ABOUTME: Defines the event system for agent execution monitoring and hooks
// ABOUTME: Provides event types, handlers, and filtering capabilities for agent lifecycle

package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event represents an event during agent execution.
// Events capture agent lifecycle, state changes, tool calls, and errors
// with timestamps and optional metadata for observability and debugging.
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

// EventType represents the type of event in the agent system.
// Different types correspond to different phases of execution and operations.
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
	EventWorkflowStart EventType = "workflow.start"
)

// NewEvent creates a new event with the specified type and data.
// Automatically assigns a unique ID and current timestamp.
// Metadata map is initialized empty and can be populated using WithMetadata.
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

// WithError adds an error to the event.
// Returns a copy of the event with the error attached.
func (e Event) WithError(err error) Event {
	e.Error = err
	return e
}

// WithMetadata adds a metadata key-value pair to the event.
// Initializes the metadata map if it doesn't exist.
// Returns a copy of the event with updated metadata.
func (e Event) WithMetadata(key string, value interface{}) Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// IsError returns true if the event represents an error condition.
// Checks both explicit error field and error-type event types.
func (e Event) IsError() bool {
	return e.Error != nil || e.Type == EventAgentError || e.Type == EventToolError
}

// EventData types for specific events

// ProgressEventData represents progress information for long-running operations.
// Includes current/total counts and optional descriptive message.
type ProgressEventData struct {
	Current int    `json:"current"`
	Total   int    `json:"total"`
	Message string `json:"message"`
}

// ToolCallEventData represents information about a tool invocation.
// Captures the tool name, parameters, and request ID for tracing.
type ToolCallEventData struct {
	ToolName   string      `json:"tool_name"`
	Parameters interface{} `json:"parameters"`
	RequestID  string      `json:"request_id"`
}

// ToolResultEventData represents the result of a tool execution.
// Includes the result data, execution duration, and request ID for correlation.
type ToolResultEventData struct {
	ToolName  string        `json:"tool_name"`
	Result    interface{}   `json:"result"`
	RequestID string        `json:"request_id"`
	Duration  time.Duration `json:"duration"`
}

// StateUpdateEventData represents information about state changes.
// Tracks the key, old/new values, and the type of operation performed.
type StateUpdateEventData struct {
	Key      string      `json:"key"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value"`
	Action   string      `json:"action"` // set, delete, merge
}

// WorkflowStepEventData represents information about workflow step execution.
// Provides step identification, position, and optional description.
type WorkflowStepEventData struct {
	StepName    string `json:"step_name"`
	StepIndex   int    `json:"step_index"`
	TotalSteps  int    `json:"total_steps"`
	Description string `json:"description,omitempty"`
}

// AgentStartEventData represents information when an agent begins execution.
// Captures initial state, configuration, and parent agent context.
type AgentStartEventData struct {
	InputState  *State                 `json:"input_state,omitempty"`
	Config      AgentConfig            `json:"config,omitempty"`
	ParentAgent string                 `json:"parent_agent,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AgentCompleteEventData represents information when an agent completes execution.
// Includes final state, execution duration, and completion metadata.
type AgentCompleteEventData struct {
	OutputState *State                 `json:"output_state,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// EventHandler processes events emitted by agents.
// Implementations can log, persist, or react to specific event types.
type EventHandler interface {
	HandleEvent(event Event) error
}

// EventHandlerFunc is a function adapter that implements EventHandler.
// Allows plain functions to be used as event handlers.
type EventHandlerFunc func(event Event) error

// HandleEvent implements EventHandler by calling the underlying function.
// This adapter allows functions to satisfy the EventHandler interface.
func (f EventHandlerFunc) HandleEvent(event Event) error {
	return f(event)
}

// EventFilter filters events based on custom criteria.
// Returns true if the event should be processed, false to skip it.
type EventFilter func(event Event) bool

// Common event filters

// FilterByType returns a filter that matches events by their type.
// Multiple types can be specified to match any of them.
func FilterByType(eventTypes ...EventType) EventFilter {
	typeMap := make(map[EventType]bool)
	for _, t := range eventTypes {
		typeMap[t] = true
	}
	return func(event Event) bool {
		return typeMap[event.Type]
	}
}

// FilterByAgent returns a filter that matches events by agent ID.
// Useful for tracking events from a specific agent instance.
func FilterByAgent(agentID string) EventFilter {
	return func(event Event) bool {
		return event.AgentID == agentID
	}
}

// FilterByAgentName returns a filter that matches events by agent name.
// Useful for tracking events from agents with a specific name.
func FilterByAgentName(agentName string) EventFilter {
	return func(event Event) bool {
		return event.AgentName == agentName
	}
}

// FilterErrors returns a filter that matches only error events.
// Combines both explicit errors and error-type events.
func FilterErrors() EventFilter {
	return func(event Event) bool {
		return event.IsError()
	}
}

// CombineFilters combines multiple filters with AND logic.
// The event must pass all filters to be accepted.
// Returns true only if all filters return true.
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

// EventDispatcher manages event distribution to subscribers.
// Supports filtered subscriptions and handles event routing
// to appropriate handlers based on their filter criteria.
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

// EventStream represents a stream of events that can be consumed sequentially.
// Provides blocking access to events as they become available.
type EventStream interface {
	// Next returns the next event or blocks until one is available
	Next() (Event, error)

	// Close closes the stream
	Close()
}

// MarshalJSON customizes JSON marshaling for Event.
// Converts the error field to a string for JSON serialization
// since error types cannot be directly marshaled to JSON.
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
