// ABOUTME: Event serialization for bridge layer integration and persistence
// ABOUTME: Converts events to/from various formats for downstream consumption

package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// SerializableEvent wraps a domain.Event with serialization capabilities.
// It adds version information and custom JSON marshaling for bridge compatibility.
type SerializableEvent struct {
	domain.Event
	Version string `json:"version"`
}

// NewSerializableEvent creates a new serializable event with version 1.0.
//
// Parameters:
//   - event: The domain event to wrap
//
// Returns a new SerializableEvent instance.
func NewSerializableEvent(event domain.Event) *SerializableEvent {
	return &SerializableEvent{
		Event:   event,
		Version: "1.0",
	}
}

// MarshalJSON implements json.Marshaler interface.
// It converts the event to a bridge-friendly format with string timestamps
// and optional fields included only when present.
func (e *SerializableEvent) MarshalJSON() ([]byte, error) {
	// Convert to a bridge-friendly format
	data := map[string]interface{}{
		"id":         e.ID,
		"type":       string(e.Type),
		"agent_id":   e.AgentID,
		"agent_name": e.AgentName,
		"timestamp":  e.Timestamp.Format(time.RFC3339Nano),
		"version":    e.Version,
	}

	// Add optional fields
	if e.Data != nil {
		data["data"] = e.Data
	}

	if len(e.Metadata) > 0 {
		data["metadata"] = e.Metadata
	}

	if e.Error != nil {
		data["error"] = e.Error.Error()
	}

	return json.Marshal(data)
}

// UnmarshalJSON implements json.Unmarshaler interface.
// It reconstructs the event from JSON data, handling type conversions
// and optional fields gracefully.
func (e *SerializableEvent) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract required fields
	if id, ok := raw["id"].(string); ok {
		e.ID = id
	}

	if typeStr, ok := raw["type"].(string); ok {
		e.Type = domain.EventType(typeStr)
	}

	if agentID, ok := raw["agent_id"].(string); ok {
		e.AgentID = agentID
	}

	if agentName, ok := raw["agent_name"].(string); ok {
		e.AgentName = agentName
	}

	if timestampStr, ok := raw["timestamp"].(string); ok {
		if ts, err := time.Parse(time.RFC3339Nano, timestampStr); err == nil {
			e.Timestamp = ts
		}
	}

	if version, ok := raw["version"].(string); ok {
		e.Version = version
	}

	// Extract optional fields
	if data, ok := raw["data"]; ok {
		e.Data = data
	}

	if metadata, ok := raw["metadata"].(map[string]interface{}); ok {
		e.Metadata = metadata
	}

	if errStr, ok := raw["error"].(string); ok && errStr != "" {
		e.Error = fmt.Errorf("%s", errStr)
	}

	return nil
}

// SerializeEvent converts an event to a bridge-friendly map format.
// This is useful for integration with external systems that expect map data.
//
// Parameters:
//   - event: The event to serialize
//
// Returns a map representation of the event or an error.
func SerializeEvent(event domain.Event) (map[string]interface{}, error) {
	serializable := NewSerializableEvent(event)

	// Marshal to JSON first
	data, err := json.Marshal(serializable)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	// Unmarshal to map
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	return result, nil
}

// DeserializeEvent converts a map back to an event.
// It handles the reverse operation of SerializeEvent.
//
// Parameters:
//   - data: Map containing event data
//
// Returns the reconstructed event or an error.
func DeserializeEvent(data map[string]interface{}) (domain.Event, error) {
	// Marshal map to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return domain.Event{}, fmt.Errorf("failed to marshal map: %w", err)
	}

	// Unmarshal to SerializableEvent
	var serializable SerializableEvent
	if err := json.Unmarshal(jsonData, &serializable); err != nil {
		return domain.Event{}, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	return serializable.Event, nil
}

// EventBatch represents a batch of events for efficient serialization.
// It groups multiple events together with metadata about the batch.
type EventBatch struct {
	Events    []SerializableEvent `json:"events"`
	BatchID   string              `json:"batch_id"`
	Timestamp time.Time           `json:"timestamp"`
	Count     int                 `json:"count"`
}

// NewEventBatch creates a new event batch from a slice of events.
// It automatically generates a batch ID and timestamp.
//
// Parameters:
//   - events: The events to include in the batch
//
// Returns a new EventBatch instance.
func NewEventBatch(events []domain.Event) *EventBatch {
	serializable := make([]SerializableEvent, len(events))
	for i, event := range events {
		serializable[i] = *NewSerializableEvent(event)
	}

	return &EventBatch{
		Events:    serializable,
		BatchID:   fmt.Sprintf("batch_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Count:     len(events),
	}
}

// SerializeEventBatch converts multiple events to a batch format.
// This is useful for bulk operations and efficient transmission.
//
// Parameters:
//   - events: The events to serialize as a batch
//
// Returns a map representation of the batch or an error.
func SerializeEventBatch(events []domain.Event) (map[string]interface{}, error) {
	batch := NewEventBatch(events)

	// Marshal to JSON
	data, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch: %w", err)
	}

	// Unmarshal to map
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch to map: %w", err)
	}

	return result, nil
}

// EventSerializer provides different serialization formats for events.
// Implementations can provide various formats like JSON, compact, or custom.
type EventSerializer interface {
	Serialize(event domain.Event) ([]byte, error)
	Deserialize(data []byte) (domain.Event, error)
	Format() string
}

// JSONSerializer serializes events to JSON format.
// It supports both compact and pretty-printed output.
type JSONSerializer struct {
	pretty bool
}

// NewJSONSerializer creates a new JSON serializer.
//
// Parameters:
//   - pretty: If true, output will be indented for readability
//
// Returns a new JSONSerializer instance.
func NewJSONSerializer(pretty bool) *JSONSerializer {
	return &JSONSerializer{pretty: pretty}
}

// Serialize implements EventSerializer interface.
// It converts an event to JSON bytes, optionally pretty-printed.
func (s *JSONSerializer) Serialize(event domain.Event) ([]byte, error) {
	serializable := NewSerializableEvent(event)

	if s.pretty {
		return json.MarshalIndent(serializable, "", "  ")
	}
	return json.Marshal(serializable)
}

// Deserialize implements EventSerializer interface.
// It reconstructs an event from JSON bytes.
func (s *JSONSerializer) Deserialize(data []byte) (domain.Event, error) {
	var serializable SerializableEvent
	if err := json.Unmarshal(data, &serializable); err != nil {
		return domain.Event{}, err
	}
	return serializable.Event, nil
}

// Format implements EventSerializer interface.
// It returns "json" as the format identifier.
func (s *JSONSerializer) Format() string {
	return "json"
}

// CompactSerializer provides minimal serialization for performance.
// It uses short field names to reduce payload size.
type CompactSerializer struct{}

// NewCompactSerializer creates a new compact serializer.
// This serializer is optimized for minimal payload size.
//
// Returns a new CompactSerializer instance.
func NewCompactSerializer() *CompactSerializer {
	return &CompactSerializer{}
}

// Serialize implements EventSerializer interface with minimal format.
// It uses abbreviated field names: i=ID, t=Type, a=AgentID, s=timestamp, d=data, e=error.
func (s *CompactSerializer) Serialize(event domain.Event) ([]byte, error) {
	// Create minimal representation
	compact := map[string]interface{}{
		"i": event.ID,
		"t": string(event.Type),
		"a": event.AgentID,
		"s": event.Timestamp.Unix(),
	}

	if event.Data != nil {
		compact["d"] = event.Data
	}

	if event.Error != nil {
		compact["e"] = event.Error.Error()
	}

	return json.Marshal(compact)
}

// Deserialize implements EventSerializer interface.
// It reconstructs an event from compact format.
func (s *CompactSerializer) Deserialize(data []byte) (domain.Event, error) {
	var compact map[string]interface{}
	if err := json.Unmarshal(data, &compact); err != nil {
		return domain.Event{}, err
	}

	event := domain.Event{
		Metadata: make(map[string]interface{}),
	}

	if id, ok := compact["i"].(string); ok {
		event.ID = id
	}

	if typeStr, ok := compact["t"].(string); ok {
		event.Type = domain.EventType(typeStr)
	}

	if agentID, ok := compact["a"].(string); ok {
		event.AgentID = agentID
	}

	if timestamp, ok := compact["s"].(float64); ok {
		event.Timestamp = time.Unix(int64(timestamp), 0)
	}

	if data, ok := compact["d"]; ok {
		event.Data = data
	}

	if errStr, ok := compact["e"].(string); ok {
		event.Error = fmt.Errorf("%s", errStr)
	}

	return event, nil
}

// Format implements EventSerializer interface.
// It returns "compact" as the format identifier.
func (s *CompactSerializer) Format() string {
	return "compact"
}

// GetSerializer returns a serializer for the specified format.
// Supported formats: "json", "json-pretty", "compact".
// Defaults to JSON if format is unrecognized.
//
// Parameters:
//   - format: The desired serialization format
//
// Returns an appropriate EventSerializer implementation.
func GetSerializer(format string) EventSerializer {
	switch format {
	case "json":
		return NewJSONSerializer(false)
	case "json-pretty":
		return NewJSONSerializer(true)
	case "compact":
		return NewCompactSerializer()
	default:
		return NewJSONSerializer(false)
	}
}
