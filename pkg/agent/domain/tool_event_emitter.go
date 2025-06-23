// ABOUTME: Implements EventEmitter interface for tools to emit events during execution
// ABOUTME: Integrates tool events with the agent's event system for unified observability

package domain

import (
	"fmt"
	"time"
)

// toolEventEmitter implements EventEmitter for tools.
// It automatically populates tool and agent context information
// for events emitted during tool execution.
type toolEventEmitter struct {
	dispatcher EventDispatcher
	toolName   string
	agentID    string
	agentName  string
}

// NewToolEventEmitter creates a new event emitter for tools.
// The emitter is pre-configured with tool name and agent context
// to automatically enrich all emitted events with proper metadata.
func NewToolEventEmitter(dispatcher EventDispatcher, toolName string, agentID string, agentName string) EventEmitter {
	return &toolEventEmitter{
		dispatcher: dispatcher,
		toolName:   toolName,
		agentID:    agentID,
		agentName:  agentName,
	}
}

// Emit sends an event through the dispatcher with tool context.
// Automatically enriches the event with tool name and agent information.
func (te *toolEventEmitter) Emit(eventType EventType, data interface{}) {
	event := Event{
		ID:        generateEventID(),
		Type:      eventType,
		AgentID:   te.agentID,
		AgentName: te.agentName,
		Timestamp: time.Now(),
		Data:      data,
		Metadata: map[string]interface{}{
			"tool_name": te.toolName,
			"source":    "tool",
		},
	}
	te.dispatcher.Dispatch(event)
}

// EmitProgress sends a progress event
func (te *toolEventEmitter) EmitProgress(current, total int, message string) {
	data := ProgressEventData{
		Current: current,
		Total:   total,
		Message: message,
	}
	te.Emit(EventProgress, data)
}

// EmitMessage sends a message event
func (te *toolEventEmitter) EmitMessage(message string) {
	data := MessageEventData{
		Message: message,
		Level:   "info",
	}
	te.Emit(EventMessage, data)
}

// EmitError sends an error event
func (te *toolEventEmitter) EmitError(err error) {
	if err == nil {
		return
	}
	te.Emit(EventToolError, err.Error())
}

// EmitCustom sends a custom event
func (te *toolEventEmitter) EmitCustom(eventName string, data interface{}) {
	// Create a custom event type
	eventType := EventType(fmt.Sprintf("tool.%s.%s", te.toolName, eventName))
	te.Emit(eventType, data)
}

// MessageEventData holds message information
type MessageEventData struct {
	Message string `json:"message"`
	Level   string `json:"level"` // info, warning, error
}

// Helper function to generate event IDs
func generateEventID() string {
	// This is a simplified version - in production, use UUID
	return fmt.Sprintf("event-%d", time.Now().UnixNano())
}
