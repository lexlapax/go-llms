// ABOUTME: Mock state implementation with history tracking and manipulation helpers
// ABOUTME: Provides state snapshots, change tracking, and deterministic behavior for testing

package mocks

import (
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// StateChange represents a recorded state modification
type StateChange struct {
	Operation string // "set", "delete", "add_artifact", etc.
	Key       string
	OldValue  interface{}
	NewValue  interface{}
	Timestamp time.Time
}

// StateSnapshot represents a point-in-time state capture
type StateSnapshot struct {
	Values    map[string]interface{}
	Artifacts map[string]*domain.Artifact
	Messages  []domain.Message
	Metadata  map[string]interface{}
	Timestamp time.Time
}

// MockState wraps domain.State with testing utilities
type MockState struct {
	*domain.State

	// Change tracking
	changes   []StateChange
	snapshots []StateSnapshot

	// Behavior hooks
	OnGet    func(key string) (interface{}, bool)
	OnSet    func(key string, value interface{})
	OnDelete func(key string)

	// Access tracking
	getCount map[string]int
	setCount map[string]int

	// Failure injection
	failureMode  bool
	failureError error

	mu sync.RWMutex
}

// NewMockState creates a new mock state
func NewMockState() *MockState {
	return &MockState{
		State:     domain.NewState(),
		changes:   make([]StateChange, 0),
		snapshots: make([]StateSnapshot, 0),
		getCount:  make(map[string]int),
		setCount:  make(map[string]int),
	}
}

// NewMockStateWithData creates a mock state with initial data
func NewMockStateWithData(data map[string]interface{}) *MockState {
	mock := NewMockState()
	for k, v := range data {
		mock.Set(k, v)
	}
	return mock
}

// Get retrieves a value with tracking
func (m *MockState) Get(key string) (interface{}, bool) {
	m.mu.Lock()
	m.getCount[key]++
	m.mu.Unlock()

	if m.OnGet != nil {
		return m.OnGet(key)
	}

	if m.failureMode {
		return nil, false
	}

	return m.State.Get(key)
}

// Set stores a value with tracking
func (m *MockState) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Track old value for change history
	oldValue, _ := m.State.Get(key)

	// Record change
	change := StateChange{
		Operation: "set",
		Key:       key,
		OldValue:  oldValue,
		NewValue:  value,
		Timestamp: time.Now(),
	}
	m.changes = append(m.changes, change)
	m.setCount[key]++

	if m.OnSet != nil {
		m.OnSet(key, value)
		return
	}

	if !m.failureMode {
		m.State.Set(key, value)
	}
}

// Delete removes a value with tracking
func (m *MockState) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Track old value for change history
	oldValue, _ := m.State.Get(key)

	// Record change
	change := StateChange{
		Operation: "delete",
		Key:       key,
		OldValue:  oldValue,
		NewValue:  nil,
		Timestamp: time.Now(),
	}
	m.changes = append(m.changes, change)

	if m.OnDelete != nil {
		m.OnDelete(key)
		return
	}

	if !m.failureMode {
		m.State.Delete(key)
	}
}

// TakeSnapshot captures the current state
func (m *MockState) TakeSnapshot() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create deep copies of the state data
	values := make(map[string]interface{})
	for k, v := range m.Values() {
		values[k] = v
	}

	artifacts := make(map[string]*domain.Artifact)
	for k, v := range m.Artifacts() {
		artifacts[k] = v
	}

	messages := make([]domain.Message, len(m.Messages()))
	copy(messages, m.Messages())

	metadata := make(map[string]interface{})
	for k, v := range m.GetAllMetadata() {
		metadata[k] = v
	}

	snapshot := StateSnapshot{
		Values:    values,
		Artifacts: artifacts,
		Messages:  messages,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	m.snapshots = append(m.snapshots, snapshot)
}

// GetSnapshot retrieves a specific snapshot
func (m *MockState) GetSnapshot(index int) (*StateSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if index < 0 || index >= len(m.snapshots) {
		return nil, fmt.Errorf("snapshot index %d out of range", index)
	}

	return &m.snapshots[index], nil
}

// GetSnapshots returns all snapshots
func (m *MockState) GetSnapshots() []StateSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots := make([]StateSnapshot, len(m.snapshots))
	copy(snapshots, m.snapshots)
	return snapshots
}

// GetChanges returns the change history
func (m *MockState) GetChanges() []StateChange {
	m.mu.RLock()
	defer m.mu.RUnlock()

	changes := make([]StateChange, len(m.changes))
	copy(changes, m.changes)
	return changes
}

// GetAccessCount returns how many times each key was accessed
func (m *MockState) GetAccessCount(key string) (gets int, sets int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getCount[key], m.setCount[key]
}

// EnableFailureMode simulates state failures
func (m *MockState) EnableFailureMode(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.failureMode = true
	m.failureError = err
}

// DisableFailureMode disables failure simulation
func (m *MockState) DisableFailureMode() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.failureMode = false
	m.failureError = nil
}

// Reset clears all tracking data
func (m *MockState) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.changes = make([]StateChange, 0)
	m.snapshots = make([]StateSnapshot, 0)
	m.getCount = make(map[string]int)
	m.setCount = make(map[string]int)
	m.failureMode = false
	m.failureError = nil
}

// Helper methods for testing

// AssertKeyAccessed checks if a key was accessed
func (m *MockState) AssertKeyAccessed(key string, minGets, minSets int) error {
	gets, sets := m.GetAccessCount(key)

	if gets < minGets {
		return fmt.Errorf("key %s: expected at least %d gets, got %d", key, minGets, gets)
	}

	if sets < minSets {
		return fmt.Errorf("key %s: expected at least %d sets, got %d", key, minSets, sets)
	}

	return nil
}

// AssertChangeCount verifies the number of changes
func (m *MockState) AssertChangeCount(expected int) error {
	m.mu.RLock()
	actual := len(m.changes)
	m.mu.RUnlock()

	if actual != expected {
		return fmt.Errorf("expected %d changes, got %d", expected, actual)
	}

	return nil
}

// FindChanges returns changes for a specific key
func (m *MockState) FindChanges(key string) []StateChange {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var keyChanges []StateChange
	for _, change := range m.changes {
		if change.Key == key {
			keyChanges = append(keyChanges, change)
		}
	}

	return keyChanges
}

// DiffSnapshots compares two snapshots
func (m *MockState) DiffSnapshots(index1, index2 int) (map[string]interface{}, error) {
	snapshot1, err := m.GetSnapshot(index1)
	if err != nil {
		return nil, err
	}

	snapshot2, err := m.GetSnapshot(index2)
	if err != nil {
		return nil, err
	}

	diff := make(map[string]interface{})

	// Check for additions and modifications
	for k, v2 := range snapshot2.Values {
		v1, exists1 := snapshot1.Values[k]
		if !exists1 {
			diff[k] = map[string]interface{}{
				"type": "added",
				"new":  v2,
			}
		} else if v1 != v2 {
			diff[k] = map[string]interface{}{
				"type": "modified",
				"old":  v1,
				"new":  v2,
			}
		}
	}

	// Check for deletions
	for k, v1 := range snapshot1.Values {
		if _, exists2 := snapshot2.Values[k]; !exists2 {
			diff[k] = map[string]interface{}{
				"type": "deleted",
				"old":  v1,
			}
		}
	}

	return diff, nil
}
