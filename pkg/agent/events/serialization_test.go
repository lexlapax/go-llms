// ABOUTME: Tests for event serialization functionality
// ABOUTME: Validates JSON serialization, bridge format conversion, and batch handling

package events

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestSerializableEvent(t *testing.T) {
	// Create a test event
	event := domain.NewEvent(domain.EventToolCall, "agent1", "TestAgent", &domain.ToolCallEventData{
		ToolName:   "calculator",
		Parameters: map[string]interface{}{"operation": "add", "a": 1, "b": 2},
		RequestID:  "req-123",
	})
	event.Metadata["environment"] = "test"

	serializable := NewSerializableEvent(event)

	// Test JSON marshaling
	data, err := json.Marshal(serializable)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Check required fields
	if result["id"] != event.ID {
		t.Errorf("Expected ID %s, got %v", event.ID, result["id"])
	}

	if result["type"] != string(event.Type) {
		t.Errorf("Expected type %s, got %v", event.Type, result["type"])
	}

	if result["version"] != "1.0" {
		t.Errorf("Expected version 1.0, got %v", result["version"])
	}

	// Check metadata
	if metadata, ok := result["metadata"].(map[string]interface{}); ok {
		if metadata["environment"] != "test" {
			t.Errorf("Expected environment 'test', got %v", metadata["environment"])
		}
	} else {
		t.Error("Missing or invalid metadata")
	}
}

func TestSerializeEvent(t *testing.T) {
	event := domain.NewEvent(domain.EventProgress, "agent1", "TestAgent", &domain.ProgressEventData{
		Current: 5,
		Total:   10,
		Message: "Processing...",
	})

	// Serialize to map
	data, err := SerializeEvent(event)
	if err != nil {
		t.Fatalf("Failed to serialize event: %v", err)
	}

	// Check structure
	if data["type"] != string(domain.EventProgress) {
		t.Errorf("Expected type %s, got %v", domain.EventProgress, data["type"])
	}

	if data["agent_id"] != "agent1" {
		t.Errorf("Expected agent_id 'agent1', got %v", data["agent_id"])
	}

	// Deserialize back
	recovered, err := DeserializeEvent(data)
	if err != nil {
		t.Fatalf("Failed to deserialize event: %v", err)
	}

	if recovered.Type != event.Type {
		t.Errorf("Expected type %s, got %s", event.Type, recovered.Type)
	}

	if recovered.AgentID != event.AgentID {
		t.Errorf("Expected agent ID %s, got %s", event.AgentID, recovered.AgentID)
	}
}

func TestEventBatch(t *testing.T) {
	events := []domain.Event{
		domain.NewEvent(domain.EventAgentStart, "agent1", "TestAgent", nil),
		domain.NewEvent(domain.EventToolCall, "agent1", "TestAgent", nil),
		domain.NewEvent(domain.EventAgentComplete, "agent1", "TestAgent", nil),
	}

	// Create batch
	batchData, err := SerializeEventBatch(events)
	if err != nil {
		t.Fatalf("Failed to serialize batch: %v", err)
	}

	// Check batch structure
	if batchData["count"] != float64(3) { // JSON numbers are float64
		t.Errorf("Expected count 3, got %v", batchData["count"])
	}

	if batchID, ok := batchData["batch_id"].(string); !ok || batchID == "" {
		t.Error("Missing or invalid batch_id")
	}

	if eventsData, ok := batchData["events"].([]interface{}); ok {
		if len(eventsData) != 3 {
			t.Errorf("Expected 3 events in batch, got %d", len(eventsData))
		}
	} else {
		t.Error("Missing or invalid events array")
	}
}

func TestJSONSerializer(t *testing.T) {
	serializer := NewJSONSerializer(false)

	event := domain.NewEvent(domain.EventStateUpdate, "agent1", "TestAgent", &domain.StateUpdateEventData{
		Key:      "counter",
		OldValue: 1,
		NewValue: 2,
		Action:   "set",
	})

	// Serialize
	data, err := serializer.Serialize(event)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	// Should be valid JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		t.Errorf("Invalid JSON output: %v", err)
	}

	// Deserialize
	recovered, err := serializer.Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if recovered.Type != event.Type {
		t.Errorf("Expected type %s, got %s", event.Type, recovered.Type)
	}

	// Test pretty format
	prettySerializer := NewJSONSerializer(true)
	prettyData, err := prettySerializer.Serialize(event)
	if err != nil {
		t.Fatalf("Failed to serialize with pretty format: %v", err)
	}

	// Pretty JSON should contain newlines and indentation
	if len(prettyData) <= len(data) {
		t.Error("Pretty JSON should be longer than compact JSON")
	}
}

func TestCompactSerializer(t *testing.T) {
	serializer := NewCompactSerializer()

	event := domain.NewEvent(domain.EventToolResult, "agent1", "TestAgent", map[string]interface{}{
		"result": "success",
		"value":  42,
	})
	event.Timestamp = time.Unix(1234567890, 0)

	// Serialize
	data, err := serializer.Serialize(event)
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	// Check compact format
	var compact map[string]interface{}
	if err := json.Unmarshal(data, &compact); err != nil {
		t.Fatalf("Failed to unmarshal compact data: %v", err)
	}

	// Should use short keys
	if _, ok := compact["i"]; !ok {
		t.Error("Missing 'i' (id) in compact format")
	}

	if _, ok := compact["t"]; !ok {
		t.Error("Missing 't' (type) in compact format")
	}

	if _, ok := compact["s"]; !ok {
		t.Error("Missing 's' (timestamp) in compact format")
	}

	// Deserialize
	recovered, err := serializer.Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if recovered.Type != event.Type {
		t.Errorf("Expected type %s, got %s", event.Type, recovered.Type)
	}

	// Timestamp should be preserved (to the second)
	if recovered.Timestamp.Unix() != event.Timestamp.Unix() {
		t.Errorf("Expected timestamp %v, got %v", event.Timestamp, recovered.Timestamp)
	}
}

func TestGetSerializer(t *testing.T) {
	tests := []struct {
		format       string
		expectedType string
	}{
		{"json", "json"},
		{"json-pretty", "json"},
		{"compact", "compact"},
		{"unknown", "json"}, // Default
	}

	for _, tt := range tests {
		serializer := GetSerializer(tt.format)
		if serializer.Format() != tt.expectedType {
			t.Errorf("Format %s: expected serializer type %s, got %s",
				tt.format, tt.expectedType, serializer.Format())
		}
	}
}

func TestEventWithError(t *testing.T) {
	event := domain.NewEvent(domain.EventToolError, "agent1", "TestAgent", nil)
	event = event.WithError(fmt.Errorf("timeout error"))

	serializable := NewSerializableEvent(event)

	// Marshal
	data, err := json.Marshal(serializable)
	if err != nil {
		t.Fatalf("Failed to marshal event with error: %v", err)
	}

	// Check error is serialized
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if errStr, ok := result["error"].(string); !ok || errStr == "" {
		t.Error("Error not properly serialized")
	}

	// Deserialize
	var recovered SerializableEvent
	if err := json.Unmarshal(data, &recovered); err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if recovered.Error == nil {
		t.Error("Error not properly deserialized")
	}
}
