// ABOUTME: Event serialization for bridge layer integration and persistence
// ABOUTME: Converts events to/from various formats for downstream consumption

package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// SerializableEvent wraps a domain.Event with serialization capabilities
type SerializableEvent struct {
	domain.Event
	Version string `json:"version"`
}

// NewSerializableEvent creates a new serializable event
func NewSerializableEvent(event domain.Event) *SerializableEvent {
	return &SerializableEvent{
		Event:   event,
		Version: "1.0",
	}
}

// MarshalJSON implements json.Marshaler
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

// UnmarshalJSON implements json.Unmarshaler
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

// SerializeEvent converts an event to a bridge-friendly map format
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

// DeserializeEvent converts a map back to an event
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

// EventBatch represents a batch of events for efficient serialization
type EventBatch struct {
	Events    []SerializableEvent `json:"events"`
	BatchID   string              `json:"batch_id"`
	Timestamp time.Time           `json:"timestamp"`
	Count     int                 `json:"count"`
}

// NewEventBatch creates a new event batch
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

// SerializeEventBatch converts multiple events to a batch format
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

// EventSerializer provides different serialization formats
type EventSerializer interface {
	Serialize(event domain.Event) ([]byte, error)
	Deserialize(data []byte) (domain.Event, error)
	Format() string
}

// JSONSerializer serializes events to JSON
type JSONSerializer struct {
	pretty bool
}

// NewJSONSerializer creates a new JSON serializer
func NewJSONSerializer(pretty bool) *JSONSerializer {
	return &JSONSerializer{pretty: pretty}
}

// Serialize implements EventSerializer
func (s *JSONSerializer) Serialize(event domain.Event) ([]byte, error) {
	serializable := NewSerializableEvent(event)

	if s.pretty {
		return json.MarshalIndent(serializable, "", "  ")
	}
	return json.Marshal(serializable)
}

// Deserialize implements EventSerializer
func (s *JSONSerializer) Deserialize(data []byte) (domain.Event, error) {
	var serializable SerializableEvent
	if err := json.Unmarshal(data, &serializable); err != nil {
		return domain.Event{}, err
	}
	return serializable.Event, nil
}

// Format implements EventSerializer
func (s *JSONSerializer) Format() string {
	return "json"
}

// CompactSerializer provides minimal serialization for performance
type CompactSerializer struct{}

// NewCompactSerializer creates a new compact serializer
func NewCompactSerializer() *CompactSerializer {
	return &CompactSerializer{}
}

// Serialize implements EventSerializer with minimal format
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

// Deserialize implements EventSerializer
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

// Format implements EventSerializer
func (s *CompactSerializer) Format() string {
	return "compact"
}

// GetSerializer returns a serializer for the specified format
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
