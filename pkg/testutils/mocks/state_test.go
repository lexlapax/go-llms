// ABOUTME: Tests for MockState implementation verifying state tracking and manipulation
// ABOUTME: Covers change tracking, snapshots, behavior hooks, and failure simulation

package mocks

import (
	"errors"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockState_Basic(t *testing.T) {
	state := NewMockState()

	// Basic set and get
	state.Set("key1", "value1")
	state.Set("key2", 42)

	val1, exists := state.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", val1)

	val2, exists := state.Get("key2")
	assert.True(t, exists)
	assert.Equal(t, 42, val2)

	// Non-existent key
	_, exists = state.Get("nonexistent")
	assert.False(t, exists)
}

func TestMockState_WithData(t *testing.T) {
	data := map[string]interface{}{
		"name":   "test",
		"count":  10,
		"active": true,
	}

	state := NewMockStateWithData(data)

	// Verify all data was set
	for k, v := range data {
		got, exists := state.Get(k)
		assert.True(t, exists)
		assert.Equal(t, v, got)
	}
}

func TestMockState_ChangeTracking(t *testing.T) {
	state := NewMockState()

	// Make some changes
	state.Set("key1", "initial")
	state.Set("key1", "updated")
	state.Set("key2", 42)
	state.Delete("key1")

	// Get change history
	changes := state.GetChanges()
	assert.Len(t, changes, 4)

	// Verify first change
	assert.Equal(t, "set", changes[0].Operation)
	assert.Equal(t, "key1", changes[0].Key)
	assert.Nil(t, changes[0].OldValue)
	assert.Equal(t, "initial", changes[0].NewValue)

	// Verify update
	assert.Equal(t, "set", changes[1].Operation)
	assert.Equal(t, "key1", changes[1].Key)
	assert.Equal(t, "initial", changes[1].OldValue)
	assert.Equal(t, "updated", changes[1].NewValue)

	// Verify delete
	assert.Equal(t, "delete", changes[3].Operation)
	assert.Equal(t, "key1", changes[3].Key)
	assert.Equal(t, "updated", changes[3].OldValue)
	assert.Nil(t, changes[3].NewValue)

	// Test finding changes for specific key
	key1Changes := state.FindChanges("key1")
	assert.Len(t, key1Changes, 3) // 2 sets + 1 delete
}

func TestMockState_AccessTracking(t *testing.T) {
	state := NewMockState()

	// Set and get multiple times
	state.Set("key1", "value")
	state.Set("key1", "updated")
	state.Get("key1")
	state.Get("key1")
	state.Get("key1")

	gets, sets := state.GetAccessCount("key1")
	assert.Equal(t, 3, gets)
	assert.Equal(t, 2, sets)

	// Test assertion helper
	err := state.AssertKeyAccessed("key1", 3, 2)
	require.NoError(t, err)

	err = state.AssertKeyAccessed("key1", 5, 2) // Too many gets expected
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected at least 5 gets")

	err = state.AssertKeyAccessed("key1", 3, 5) // Too many sets expected
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected at least 5 sets")
}

func TestMockState_Snapshots(t *testing.T) {
	state := NewMockState()

	// Set initial state
	state.Set("key1", "value1")
	state.Set("key2", 42)

	// Add artifact
	artifact := &domain.Artifact{
		ID:       "artifact1",
		Name:     "test-artifact",
		Type:     domain.ArtifactTypeFile,
		MimeType: "text/plain",
		Size:     16,
		Created:  time.Now(),
	}
	state.AddArtifact(artifact)

	// Add message
	msg := domain.Message{
		Role:    "user",
		Content: "test message",
	}
	state.AddMessage(msg)

	// Add metadata
	state.SetMetadata("meta1", "metavalue")

	// Take snapshot
	state.TakeSnapshot()

	// Modify state
	state.Set("key1", "modified")
	state.Delete("key2")
	state.Set("key3", "new")

	// Take another snapshot
	state.TakeSnapshot()

	// Get snapshots
	snapshots := state.GetSnapshots()
	assert.Len(t, snapshots, 2)

	// Verify first snapshot
	snap1, err := state.GetSnapshot(0)
	require.NoError(t, err)
	assert.Equal(t, "value1", snap1.Values["key1"])
	assert.Equal(t, 42, snap1.Values["key2"])
	assert.Len(t, snap1.Artifacts, 1)
	assert.Len(t, snap1.Messages, 1)
	assert.Equal(t, "metavalue", snap1.Metadata["meta1"])

	// Verify second snapshot
	snap2, err := state.GetSnapshot(1)
	require.NoError(t, err)
	assert.Equal(t, "modified", snap2.Values["key1"])
	_, exists := snap2.Values["key2"]
	assert.False(t, exists)
	assert.Equal(t, "new", snap2.Values["key3"])

	// Test diff
	diff, err := state.DiffSnapshots(0, 1)
	require.NoError(t, err)

	// Check modifications
	key1Diff := diff["key1"].(map[string]interface{})
	assert.Equal(t, "modified", key1Diff["type"])
	assert.Equal(t, "value1", key1Diff["old"])
	assert.Equal(t, "modified", key1Diff["new"])

	// Check deletions
	key2Diff := diff["key2"].(map[string]interface{})
	assert.Equal(t, "deleted", key2Diff["type"])
	assert.Equal(t, 42, key2Diff["old"])

	// Check additions
	key3Diff := diff["key3"].(map[string]interface{})
	assert.Equal(t, "added", key3Diff["type"])
	assert.Equal(t, "new", key3Diff["new"])
}

func TestMockState_BehaviorHooks(t *testing.T) {
	state := NewMockState()

	// Test OnGet hook
	getCalled := false
	state.OnGet = func(key string) (interface{}, bool) {
		getCalled = true
		if key == "hooked" {
			return "hooked value", true
		}
		return nil, false
	}

	val, exists := state.Get("hooked")
	assert.True(t, getCalled)
	assert.True(t, exists)
	assert.Equal(t, "hooked value", val)

	// Test OnSet hook
	setCalled := false
	var setKey string
	var setValue interface{}
	state.OnSet = func(key string, value interface{}) {
		setCalled = true
		setKey = key
		setValue = value
	}

	state.Set("test", "value")
	assert.True(t, setCalled)
	assert.Equal(t, "test", setKey)
	assert.Equal(t, "value", setValue)

	// Note: When OnSet is defined, it overrides normal behavior
	// So the value won't actually be stored
	_, exists = state.State.Get("test")
	assert.False(t, exists)

	// Test OnDelete hook
	deleteCalled := false
	state.OnDelete = func(key string) {
		deleteCalled = true
	}

	state.Delete("somekey")
	assert.True(t, deleteCalled)
}

func TestMockState_FailureMode(t *testing.T) {
	state := NewMockState()

	// Set some initial values
	state.Set("key1", "value1")

	// Enable failure mode
	testErr := errors.New("simulated failure")
	state.EnableFailureMode(testErr)

	// Get should return nil, false in failure mode
	val, exists := state.Get("key1")
	assert.False(t, exists)
	assert.Nil(t, val)

	// Set should not store in failure mode
	state.Set("key2", "value2")

	// Disable failure mode
	state.DisableFailureMode()

	// Original value should still be there
	val, exists = state.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", val)

	// New value should not have been stored
	_, exists = state.Get("key2")
	assert.False(t, exists)
}

func TestMockState_ChangeCount(t *testing.T) {
	state := NewMockState()

	// Make some changes
	state.Set("key1", "value1")
	state.Set("key2", "value2")
	state.Delete("key1")

	// Assert change count
	err := state.AssertChangeCount(3)
	require.NoError(t, err)

	err = state.AssertChangeCount(5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 5 changes, got 3")
}

func TestMockState_Reset(t *testing.T) {
	state := NewMockState()

	// Add data and track changes
	state.Set("key1", "value1")
	state.Get("key1")
	state.TakeSnapshot()
	state.EnableFailureMode(errors.New("test"))

	// Reset
	state.Reset()

	// Verify everything is cleared
	assert.Empty(t, state.GetChanges())
	assert.Empty(t, state.GetSnapshots())

	gets, sets := state.GetAccessCount("key1")
	assert.Equal(t, 0, gets)
	assert.Equal(t, 0, sets)

	// Failure mode should be disabled
	val, exists := state.Get("key1")
	assert.True(t, exists) // Original data is preserved
	assert.Equal(t, "value1", val)
}

func TestMockState_SnapshotErrors(t *testing.T) {
	state := NewMockState()

	// Try to get non-existent snapshot
	_, err := state.GetSnapshot(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")

	// Take a snapshot
	state.TakeSnapshot()

	// Try invalid indices
	_, err = state.GetSnapshot(-1)
	assert.Error(t, err)

	_, err = state.GetSnapshot(1)
	assert.Error(t, err)
}
