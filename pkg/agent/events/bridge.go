// ABOUTME: Bridge event types and utilities for go-llmspell integration
// ABOUTME: Provides event types and helpers specifically designed for bridge layer communication

package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// BridgeEventType defines events specific to bridge layer integration
type BridgeEventType string

const (
	// Bridge lifecycle events
	BridgeEventConnected    BridgeEventType = "bridge.connected"
	BridgeEventDisconnected BridgeEventType = "bridge.disconnected"
	BridgeEventReady        BridgeEventType = "bridge.ready"
	BridgeEventError        BridgeEventType = "bridge.error"

	// Bridge communication events
	BridgeEventRequest  BridgeEventType = "bridge.request"
	BridgeEventResponse BridgeEventType = "bridge.response"
	BridgeEventCallback BridgeEventType = "bridge.callback"

	// Type conversion events
	BridgeEventConvert      BridgeEventType = "bridge.convert"
	BridgeEventConvertError BridgeEventType = "bridge.convert.error"

	// Script execution events
	BridgeEventScriptStart  BridgeEventType = "bridge.script.start"
	BridgeEventScriptEnd    BridgeEventType = "bridge.script.end"
	BridgeEventScriptError  BridgeEventType = "bridge.script.error"
	BridgeEventScriptOutput BridgeEventType = "bridge.script.output"
)

// BridgeEvent extends domain.Event with bridge-specific fields
type BridgeEvent struct {
	domain.Event
	BridgeID   string                 `json:"bridge_id"`
	SessionID  string                 `json:"session_id"`
	Language   string                 `json:"language,omitempty"`
	ScriptData map[string]interface{} `json:"script_data,omitempty"`
}

// NewBridgeEvent creates a new bridge event
func NewBridgeEvent(eventType BridgeEventType, bridgeID, sessionID string, data interface{}) *BridgeEvent {
	return &BridgeEvent{
		Event: domain.Event{
			ID:        fmt.Sprintf("bridge_%d", time.Now().UnixNano()),
			Type:      domain.EventType(eventType),
			Timestamp: time.Now(),
			Data:      data,
			Metadata:  make(map[string]interface{}),
		},
		BridgeID:  bridgeID,
		SessionID: sessionID,
	}
}

// WithLanguage sets the scripting language for the event
func (e *BridgeEvent) WithLanguage(language string) *BridgeEvent {
	e.Language = language
	return e
}

// WithScriptData adds script-specific data
func (e *BridgeEvent) WithScriptData(key string, value interface{}) *BridgeEvent {
	if e.ScriptData == nil {
		e.ScriptData = make(map[string]interface{})
	}
	e.ScriptData[key] = value
	return e
}

// AsDomainEvent converts to a standard domain.Event
func (e *BridgeEvent) AsDomainEvent() domain.Event {
	// Add bridge-specific fields to metadata
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata["bridge_id"] = e.BridgeID
	e.Metadata["session_id"] = e.SessionID

	if e.Language != "" {
		e.Metadata["language"] = e.Language
	}

	if len(e.ScriptData) > 0 {
		e.Metadata["script_data"] = e.ScriptData
	}

	return e.Event
}

// BridgeRequestData represents a request from the bridge layer
type BridgeRequestData struct {
	RequestID  string                 `json:"request_id"`
	Method     string                 `json:"method"`
	Parameters map[string]interface{} `json:"parameters"`
	Timestamp  time.Time              `json:"timestamp"`
}

// BridgeResponseData represents a response to the bridge layer
type BridgeResponseData struct {
	RequestID string                 `json:"request_id"`
	Result    interface{}            `json:"result,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ScriptExecutionData represents script execution information
type ScriptExecutionData struct {
	ScriptID  string                 `json:"script_id"`
	Language  string                 `json:"language"`
	Source    string                 `json:"source"`
	Args      map[string]interface{} `json:"args,omitempty"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Output    interface{}            `json:"output,omitempty"`
	Error     string                 `json:"error,omitempty"`
	ExitCode  int                    `json:"exit_code"`
}

// BridgeEventHandler handles bridge-specific events
type BridgeEventHandler interface {
	HandleBridgeEvent(ctx context.Context, event *BridgeEvent) error
}

// BridgeEventHandlerFunc is a function adapter for BridgeEventHandler
type BridgeEventHandlerFunc func(ctx context.Context, event *BridgeEvent) error

// HandleBridgeEvent implements BridgeEventHandler
func (f BridgeEventHandlerFunc) HandleBridgeEvent(ctx context.Context, event *BridgeEvent) error {
	return f(ctx, event)
}

// BridgeEventListener listens for bridge events and converts them
type BridgeEventListener struct {
	bus     *EventBus
	handler BridgeEventHandler
	subID   string
	pattern string
}

// NewBridgeEventListener creates a new bridge event listener
func NewBridgeEventListener(bus *EventBus, handler BridgeEventHandler) *BridgeEventListener {
	return &BridgeEventListener{
		bus:     bus,
		handler: handler,
	}
}

// Listen starts listening for bridge events
func (l *BridgeEventListener) Listen(pattern string) error {
	if l.subID != "" {
		return fmt.Errorf("already listening")
	}

	// Create event handler that converts to bridge events
	handler := EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		// Check if this is a bridge event
		if bridgeID, ok := event.Metadata["bridge_id"].(string); ok {
			bridgeEvent := &BridgeEvent{
				Event:    event,
				BridgeID: bridgeID,
			}

			if sessionID, ok := event.Metadata["session_id"].(string); ok {
				bridgeEvent.SessionID = sessionID
			}

			if language, ok := event.Metadata["language"].(string); ok {
				bridgeEvent.Language = language
			}

			if scriptData, ok := event.Metadata["script_data"].(map[string]interface{}); ok {
				bridgeEvent.ScriptData = scriptData
			}

			return l.handler.HandleBridgeEvent(ctx, bridgeEvent)
		}

		return nil
	})

	var err error
	l.subID, err = l.bus.SubscribePattern(pattern, handler)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	l.pattern = pattern
	return nil
}

// Stop stops listening for events
func (l *BridgeEventListener) Stop() {
	if l.subID != "" {
		l.bus.Unsubscribe(l.subID)
		l.subID = ""
		l.pattern = ""
	}
}

// BridgeEventPublisher publishes bridge events
type BridgeEventPublisher struct {
	bus       *EventBus
	bridgeID  string
	sessionID string
}

// NewBridgeEventPublisher creates a new bridge event publisher
func NewBridgeEventPublisher(bus *EventBus, bridgeID, sessionID string) *BridgeEventPublisher {
	return &BridgeEventPublisher{
		bus:       bus,
		bridgeID:  bridgeID,
		sessionID: sessionID,
	}
}

// PublishRequest publishes a bridge request event
func (p *BridgeEventPublisher) PublishRequest(method string, parameters map[string]interface{}) string {
	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())

	data := &BridgeRequestData{
		RequestID:  requestID,
		Method:     method,
		Parameters: parameters,
		Timestamp:  time.Now(),
	}

	event := NewBridgeEvent(BridgeEventRequest, p.bridgeID, p.sessionID, data)
	p.bus.Publish(event.AsDomainEvent())

	return requestID
}

// PublishResponse publishes a bridge response event
func (p *BridgeEventPublisher) PublishResponse(requestID string, result interface{}, err error, duration time.Duration) {
	data := &BridgeResponseData{
		RequestID: requestID,
		Duration:  duration,
		Metadata:  make(map[string]interface{}),
	}

	if err != nil {
		data.Error = err.Error()
	} else {
		data.Result = result
	}

	event := NewBridgeEvent(BridgeEventResponse, p.bridgeID, p.sessionID, data)
	p.bus.Publish(event.AsDomainEvent())
}

// PublishScriptExecution publishes script execution events
func (p *BridgeEventPublisher) PublishScriptExecution(scriptData *ScriptExecutionData) {
	var eventType BridgeEventType

	if scriptData.EndTime == nil {
		eventType = BridgeEventScriptStart
	} else if scriptData.Error != "" {
		eventType = BridgeEventScriptError
	} else {
		eventType = BridgeEventScriptEnd
	}

	event := NewBridgeEvent(eventType, p.bridgeID, p.sessionID, scriptData)
	event.WithLanguage(scriptData.Language)

	p.bus.Publish(event.AsDomainEvent())
}

// SerializeBridgeEvent converts a bridge event to a map for bridge layer
func SerializeBridgeEvent(event *BridgeEvent) (map[string]interface{}, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bridge event: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	return result, nil
}
