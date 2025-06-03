// ABOUTME: Tests for the State type including thread safety and state operations
// ABOUTME: Validates state management, cloning, merging, and serialization

package domain_test

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestNewState(t *testing.T) {
	state := domain.NewState()

	if state == nil {
		t.Fatal("NewState returned nil")
	}

	if state.ID() == "" {
		t.Error("State ID should not be empty")
	}

	if state.Version() != 1 {
		t.Errorf("Initial version should be 1, got %d", state.Version())
	}

	if state.ParentID() != "" {
		t.Error("Initial state should not have a parent ID")
	}
}

func TestStateGetSet(t *testing.T) {
	state := domain.NewState()

	// Test setting and getting values
	state.Set("key1", "value1")
	state.Set("key2", 42)
	state.Set("key3", true)

	// Test getting existing values
	val1, ok := state.Get("key1")
	if !ok || val1 != "value1" {
		t.Errorf("Expected 'value1', got %v", val1)
	}

	val2, ok := state.Get("key2")
	if !ok || val2 != 42 {
		t.Errorf("Expected 42, got %v", val2)
	}

	val3, ok := state.Get("key3")
	if !ok || val3 != true {
		t.Errorf("Expected true, got %v", val3)
	}

	// Test getting non-existent value
	_, ok = state.Get("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}

	// Test version increment
	if state.Version() != 4 { // 1 initial + 3 sets
		t.Errorf("Expected version 4, got %d", state.Version())
	}
}

func TestStateDelete(t *testing.T) {
	state := domain.NewState()

	state.Set("key1", "value1")
	state.Set("key2", "value2")

	// Delete a key
	state.Delete("key1")

	// Verify it's deleted
	_, ok := state.Get("key1")
	if ok {
		t.Error("Key should have been deleted")
	}

	// Verify other key still exists
	val, ok := state.Get("key2")
	if !ok || val != "value2" {
		t.Error("Other key should still exist")
	}

	// Test version increment
	if state.Version() != 4 { // 1 initial + 2 sets + 1 delete
		t.Errorf("Expected version 4, got %d", state.Version())
	}
}

func TestStateValues(t *testing.T) {
	state := domain.NewState()

	expected := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	for k, v := range expected {
		state.Set(k, v)
	}

	values := state.Values()

	// Check all values are present
	for k, v := range expected {
		if val, ok := values[k]; !ok || val != v {
			t.Errorf("Expected %v for key %s, got %v", v, k, val)
		}
	}

	// Verify it's a copy (modification doesn't affect state)
	values["key4"] = "value4"
	_, ok := state.Get("key4")
	if ok {
		t.Error("Modifying returned values should not affect state")
	}
}

func TestStateArtifacts(t *testing.T) {
	state := domain.NewState()

	// Create artifacts
	artifact1 := domain.NewArtifact("file1.txt", domain.ArtifactTypeFile, []byte("content1"))
	artifact2 := domain.NewArtifact("image.png", domain.ArtifactTypeImage, []byte("content2"))

	// Add artifacts
	state.AddArtifact(artifact1)
	state.AddArtifact(artifact2)

	// Get specific artifact
	retrieved, ok := state.GetArtifact(artifact1.ID)
	if !ok {
		t.Error("Should find artifact by ID")
	}
	if retrieved.Name != artifact1.Name {
		t.Errorf("Expected artifact name %s, got %s", artifact1.Name, retrieved.Name)
	}

	// Get all artifacts
	artifacts := state.Artifacts()
	if len(artifacts) != 2 {
		t.Errorf("Expected 2 artifacts, got %d", len(artifacts))
	}

	// Verify it's a copy
	delete(artifacts, artifact1.ID)
	_, ok = state.GetArtifact(artifact1.ID)
	if !ok {
		t.Error("Modifying returned artifacts should not affect state")
	}
}

func TestStateMessages(t *testing.T) {
	state := domain.NewState()

	// Add messages
	msg1 := domain.NewMessage("user", "Hello")
	msg2 := domain.NewMessage("assistant", "Hi there")

	state.AddMessage(msg1)
	state.AddMessage(msg2)

	// Get messages
	messages := state.Messages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "user" || messages[0].Content != "Hello" {
		t.Error("First message incorrect")
	}

	if messages[1].Role != "assistant" || messages[1].Content != "Hi there" {
		t.Error("Second message incorrect")
	}

	// Verify it's a copy
	messages[0].Content = "Modified"
	retrievedMessages := state.Messages()
	if retrievedMessages[0].Content == "Modified" {
		t.Error("Modifying returned messages should not affect state")
	}
}

func TestStateClone(t *testing.T) {
	original := domain.NewState()

	// Set up original state
	original.Set("key1", "value1")
	original.Set("key2", map[string]interface{}{
		"nested": "value",
	})

	artifact := domain.NewArtifact("file.txt", domain.ArtifactTypeFile, []byte("content"))
	original.AddArtifact(artifact)

	msg := domain.NewMessage("user", "Hello")
	original.AddMessage(msg)

	original.SetMetadata("meta1", "metavalue1")

	// Clone the state
	cloned := original.Clone()

	// Verify clone has different ID
	if cloned.ID() == original.ID() {
		t.Error("Clone should have different ID")
	}

	// Verify parent ID is set
	if cloned.ParentID() != original.ID() {
		t.Errorf("Clone parent ID should be %s, got %s", original.ID(), cloned.ParentID())
	}

	// Verify values are copied
	val1, ok := cloned.Get("key1")
	if !ok || val1 != "value1" {
		t.Error("Clone should have same values")
	}

	// Verify deep copy (modifying original doesn't affect clone)
	original.Set("key1", "modified")
	val1, _ = cloned.Get("key1")
	if val1 != "value1" {
		t.Error("Modifying original should not affect clone")
	}

	// Verify artifacts are copied
	clonedArtifacts := cloned.Artifacts()
	if len(clonedArtifacts) != 1 {
		t.Error("Clone should have same artifacts")
	}

	// Verify messages are copied
	clonedMessages := cloned.Messages()
	if len(clonedMessages) != 1 {
		t.Error("Clone should have same messages")
	}

	// Verify metadata is copied
	metaVal, ok := cloned.GetMetadata("meta1")
	if !ok || metaVal != "metavalue1" {
		t.Error("Clone should have same metadata")
	}

	// Verify version reset
	if cloned.Version() != 1 {
		t.Errorf("Clone should have version 1, got %d", cloned.Version())
	}
}

func TestStateMerge(t *testing.T) {
	state1 := domain.NewState()
	state1.Set("key1", "value1")
	state1.Set("key2", "value2")

	artifact1 := domain.NewArtifact("file1.txt", domain.ArtifactTypeFile, []byte("content1"))
	state1.AddArtifact(artifact1)

	msg1 := domain.NewMessage("user", "Message1")
	state1.AddMessage(msg1)

	state2 := domain.NewState()
	state2.Set("key2", "value2_modified") // Override
	state2.Set("key3", "value3")          // New key

	artifact2 := domain.NewArtifact("file2.txt", domain.ArtifactTypeFile, []byte("content2"))
	state2.AddArtifact(artifact2)

	msg2 := domain.NewMessage("assistant", "Message2")
	state2.AddMessage(msg2)

	// Merge state2 into state1
	state1.Merge(state2)

	// Verify values
	val1, _ := state1.Get("key1")
	if val1 != "value1" {
		t.Error("Unchanged values should remain")
	}

	val2, _ := state1.Get("key2")
	if val2 != "value2_modified" {
		t.Error("Values should be overridden by merge")
	}

	val3, _ := state1.Get("key3")
	if val3 != "value3" {
		t.Error("New values should be added")
	}

	// Verify artifacts are merged
	artifacts := state1.Artifacts()
	if len(artifacts) != 2 {
		t.Errorf("Expected 2 artifacts after merge, got %d", len(artifacts))
	}

	// Verify messages are appended
	messages := state1.Messages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages after merge, got %d", len(messages))
	}

	// Test merging nil state
	state1.Merge(nil) // Should not panic
}

func TestStateThreadSafety(t *testing.T) {
	state := domain.NewState()
	iterations := 100
	numGoroutines := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // 3 operations per goroutine

	// Concurrent sets
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := sprintf("key_%d_%d", id, j)
				state.Set(key, j)
			}
		}(i)
	}

	// Concurrent gets
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := sprintf("key_%d_%d", id, j)
				state.Get(key)
			}
		}(i)
	}

	// Concurrent clones
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				clone := state.Clone()
				// Use the clone to avoid compiler optimization
				_ = clone.ID()
			}
		}()
	}

	wg.Wait()

	// Verify state is consistent
	values := state.Values()
	expectedKeys := numGoroutines * iterations
	if len(values) != expectedKeys {
		t.Errorf("Expected %d keys, got %d", expectedKeys, len(values))
	}
}

func TestStateSerialization(t *testing.T) {
	original := domain.NewState()

	// Set up state with various types
	original.Set("string", "value")
	original.Set("number", 42)
	original.Set("float", 3.14)
	original.Set("bool", true)
	original.Set("array", []interface{}{"a", "b", "c"})
	original.Set("object", map[string]interface{}{
		"nested": "value",
		"number": 123,
	})

	artifact := domain.NewArtifact("file.txt", domain.ArtifactTypeFile, []byte("content"))
	original.AddArtifact(artifact)

	msg := domain.NewMessage("user", "Hello")
	original.AddMessage(msg)

	original.SetMetadata("meta1", "value1")

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal state: %v", err)
	}

	// Unmarshal to new state
	restored := &domain.State{}
	err = json.Unmarshal(data, restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal state: %v", err)
	}

	// Verify all data is restored
	if restored.ID() != original.ID() {
		t.Error("ID should be preserved")
	}

	// Check values
	val, _ := restored.Get("string")
	if val != "value" {
		t.Error("String value not restored correctly")
	}

	val, _ = restored.Get("number")
	// JSON unmarshals numbers as float64
	if num, ok := val.(float64); !ok || num != 42 {
		t.Errorf("Number value not restored correctly: %v (%T)", val, val)
	}

	val, _ = restored.Get("bool")
	if val != true {
		t.Error("Bool value not restored correctly")
	}

	// Check complex types
	val, _ = restored.Get("array")
	if arr, ok := val.([]interface{}); !ok || len(arr) != 3 {
		t.Error("Array not restored correctly")
	}

	val, _ = restored.Get("object")
	if obj, ok := val.(map[string]interface{}); !ok {
		t.Error("Object not restored correctly")
	} else {
		if obj["nested"] != "value" {
			t.Error("Nested value not restored correctly")
		}
	}

	// Check artifacts
	restoredArtifacts := restored.Artifacts()
	if len(restoredArtifacts) != 1 {
		t.Error("Artifacts not restored correctly")
	}

	// Check messages
	restoredMessages := restored.Messages()
	if len(restoredMessages) != 1 {
		t.Error("Messages not restored correctly")
	}

	// Check metadata
	metaVal, ok := restored.GetMetadata("meta1")
	if !ok || metaVal != "value1" {
		t.Error("Metadata not restored correctly")
	}

	// Check version
	if restored.Version() != original.Version() {
		t.Error("Version not restored correctly")
	}
}

func TestStateMetadata(t *testing.T) {
	state := domain.NewState()

	// Set metadata
	state.SetMetadata("key1", "value1")
	state.SetMetadata("key2", 42)
	state.SetMetadata("key3", map[string]interface{}{
		"nested": "value",
	})

	// Get metadata
	val1, ok := state.GetMetadata("key1")
	if !ok || val1 != "value1" {
		t.Error("Metadata key1 not set correctly")
	}

	val2, ok := state.GetMetadata("key2")
	if !ok || val2 != 42 {
		t.Error("Metadata key2 not set correctly")
	}

	val3, ok := state.GetMetadata("key3")
	if obj, ok2 := val3.(map[string]interface{}); !ok || !ok2 || obj["nested"] != "value" {
		t.Error("Metadata key3 not set correctly")
	}

	// Get non-existent metadata
	_, ok = state.GetMetadata("nonexistent")
	if ok {
		t.Error("Should return false for non-existent metadata")
	}
}

func TestStateModificationTime(t *testing.T) {
	state := domain.NewState()

	created := state.Created()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure time difference

	// Modification should update modified time
	state.Set("key", "value")
	modified1 := state.Modified()

	if !modified1.After(created) {
		t.Error("Modified time should be after created time")
	}

	time.Sleep(10 * time.Millisecond)

	// Another modification
	state.Set("key2", "value2")
	modified2 := state.Modified()

	if !modified2.After(modified1) {
		t.Error("Modified time should update with each change")
	}
}

// Benchmark tests
func BenchmarkStateSet(b *testing.B) {
	state := domain.NewState()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		state.Set(sprintf("key%d", i), i)
	}
}

func BenchmarkStateGet(b *testing.B) {
	state := domain.NewState()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		state.Set(sprintf("key%d", i), i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		state.Get(sprintf("key%d", i%1000))
	}
}

func BenchmarkStateClone(b *testing.B) {
	state := domain.NewState()

	// Pre-populate
	for i := 0; i < 100; i++ {
		state.Set(sprintf("key%d", i), i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = state.Clone()
	}
}

func BenchmarkStateConcurrentAccess(b *testing.B) {
	state := domain.NewState()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				state.Set(sprintf("key%d", i), i)
			} else {
				state.Get(sprintf("key%d", i))
			}
			i++
		}
	})
}
