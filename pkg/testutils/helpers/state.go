// ABOUTME: State testing utilities for diffing, snapshots, and mutations
// ABOUTME: Provides helpers for testing state changes and validating state transitions

package helpers

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// StateDiff represents the difference between two states
type StateDiff struct {
	Added    map[string]interface{}
	Modified map[string]ModifiedValue
	Removed  []string
}

// ModifiedValue represents a value that was modified
type ModifiedValue struct {
	Old interface{}
	New interface{}
}

// DiffStates compares two states and returns the differences
func DiffStates(old, new *domain.State) *StateDiff {
	diff := &StateDiff{
		Added:    make(map[string]interface{}),
		Modified: make(map[string]ModifiedValue),
		Removed:  make([]string, 0),
	}

	oldValues := old.Values()
	newValues := new.Values()

	// Check for added and modified values
	for key, newVal := range newValues {
		if oldVal, exists := oldValues[key]; exists {
			if !reflect.DeepEqual(oldVal, newVal) {
				diff.Modified[key] = ModifiedValue{
					Old: oldVal,
					New: newVal,
				}
			}
		} else {
			diff.Added[key] = newVal
		}
	}

	// Check for removed values
	for key := range oldValues {
		if _, exists := newValues[key]; !exists {
			diff.Removed = append(diff.Removed, key)
		}
	}

	return diff
}

// IsEmpty returns true if there are no differences
func (d *StateDiff) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Modified) == 0 && len(d.Removed) == 0
}

// String returns a string representation of the diff
func (d *StateDiff) String() string {
	if d.IsEmpty() {
		return "No differences"
	}

	var sb strings.Builder
	sb.WriteString("State Differences:\n")

	if len(d.Added) > 0 {
		sb.WriteString("\nAdded:\n")
		keys := make([]string, 0, len(d.Added))
		for k := range d.Added {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			sb.WriteString(fmt.Sprintf("  + %s: %v\n", key, d.Added[key]))
		}
	}

	if len(d.Modified) > 0 {
		sb.WriteString("\nModified:\n")
		keys := make([]string, 0, len(d.Modified))
		for k := range d.Modified {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			mod := d.Modified[key]
			sb.WriteString(fmt.Sprintf("  ~ %s: %v -> %v\n", key, mod.Old, mod.New))
		}
	}

	if len(d.Removed) > 0 {
		sb.WriteString("\nRemoved:\n")
		sort.Strings(d.Removed)
		for _, key := range d.Removed {
			sb.WriteString(fmt.Sprintf("  - %s\n", key))
		}
	}

	return sb.String()
}

// StateSnapshot captures a state at a point in time
type StateSnapshot struct {
	Values    map[string]interface{}
	Artifacts map[string]*domain.Artifact
	Messages  []domain.Message
	// Metadata is stored within state values, not as a separate field
}

// CaptureSnapshot creates a snapshot of the current state
func CaptureSnapshot(state *domain.State) *StateSnapshot {
	return &StateSnapshot{
		Values:    state.Values(),
		Artifacts: state.Artifacts(),
		Messages:  state.Messages(),
		// Note: Metadata is not a separate field in agent domain state
	}
}

// CompareSnapshots compares two state snapshots
func CompareSnapshots(old, new *StateSnapshot) *SnapshotComparison {
	comp := &SnapshotComparison{
		ValuesDiff:    compareValues(old.Values, new.Values),
		ArtifactsDiff: compareArtifacts(old.Artifacts, new.Artifacts),
		MessagesDiff:  compareMessages(old.Messages, new.Messages),
		// MetadataDiff not applicable
	}
	return comp
}

// SnapshotComparison holds the comparison between two snapshots
type SnapshotComparison struct {
	ValuesDiff    *StateDiff
	ArtifactsDiff *ArtifactsDiff
	MessagesDiff  *MessagesDiff
	// MetadataDiff not applicable for agent domain state
}

// ArtifactsDiff represents differences in artifacts
type ArtifactsDiff struct {
	Added    map[string]*domain.Artifact
	Modified map[string]struct{ Old, New *domain.Artifact }
	Removed  []string
}

// MessagesDiff represents differences in messages
type MessagesDiff struct {
	Added int
	Total int
}

func compareValues(old, new map[string]interface{}) *StateDiff {
	diff := &StateDiff{
		Added:    make(map[string]interface{}),
		Modified: make(map[string]ModifiedValue),
		Removed:  make([]string, 0),
	}

	for key, newVal := range new {
		if oldVal, exists := old[key]; exists {
			if !reflect.DeepEqual(oldVal, newVal) {
				diff.Modified[key] = ModifiedValue{Old: oldVal, New: newVal}
			}
		} else {
			diff.Added[key] = newVal
		}
	}

	for key := range old {
		if _, exists := new[key]; !exists {
			diff.Removed = append(diff.Removed, key)
		}
	}

	return diff
}

func compareArtifacts(old, new map[string]*domain.Artifact) *ArtifactsDiff {
	diff := &ArtifactsDiff{
		Added:    make(map[string]*domain.Artifact),
		Modified: make(map[string]struct{ Old, New *domain.Artifact }),
		Removed:  make([]string, 0),
	}

	for id, newArt := range new {
		if oldArt, exists := old[id]; exists {
			if !artifactsEqual(oldArt, newArt) {
				diff.Modified[id] = struct{ Old, New *domain.Artifact }{Old: oldArt, New: newArt}
			}
		} else {
			diff.Added[id] = newArt
		}
	}

	for id := range old {
		if _, exists := new[id]; !exists {
			diff.Removed = append(diff.Removed, id)
		}
	}

	return diff
}

func compareMessages(old, new []domain.Message) *MessagesDiff {
	return &MessagesDiff{
		Added: len(new) - len(old),
		Total: len(new),
	}
}

func artifactsEqual(a, b *domain.Artifact) bool {
	return a.ID == b.ID &&
		a.Type == b.Type &&
		a.MimeType == b.MimeType &&
		reflect.DeepEqual(a.Data, b.Data) &&
		reflect.DeepEqual(a.Metadata, b.Metadata)
}

// StateMutator provides fluent state mutations for testing
type StateMutator struct {
	state *domain.State
}

// MutateState creates a new state mutator
func MutateState(state *domain.State) *StateMutator {
	return &StateMutator{state: state}
}

// Set sets a value in the state
func (sm *StateMutator) Set(key string, value interface{}) *StateMutator {
	sm.state.Set(key, value)
	return sm
}

// SetMultiple sets multiple values
func (sm *StateMutator) SetMultiple(values map[string]interface{}) *StateMutator {
	for k, v := range values {
		sm.state.Set(k, v)
	}
	return sm
}

// Delete removes a key from the state
func (sm *StateMutator) Delete(key string) *StateMutator {
	sm.state.Delete(key)
	return sm
}

// AddArtifact adds an artifact to the state
func (sm *StateMutator) AddArtifact(artifact *domain.Artifact) *StateMutator {
	sm.state.AddArtifact(artifact)
	return sm
}

// AddMessage adds a message to the state
func (sm *StateMutator) AddMessage(role domain.Role, content string) *StateMutator {
	sm.state.AddMessage(domain.NewMessage(role, content))
	return sm
}

// SetMetadata sets a metadata value
func (sm *StateMutator) SetMetadata(key string, value interface{}) *StateMutator {
	sm.state.SetMetadata(key, value)
	return sm
}

// Clear clears all values from the state
func (sm *StateMutator) Clear() *StateMutator {
	// Clear all values
	for _, key := range sm.state.Keys() {
		sm.state.Delete(key)
	}
	return sm
}

// Done returns the mutated state
func (sm *StateMutator) Done() *domain.State {
	return sm.state
}

// StateValidator provides state validation helpers
type StateValidator struct {
	state  *domain.State
	errors []string
}

// ValidateState creates a new state validator
func ValidateState(state *domain.State) *StateValidator {
	return &StateValidator{
		state:  state,
		errors: make([]string, 0),
	}
}

// HasKey validates that a key exists
func (sv *StateValidator) HasKey(key string) *StateValidator {
	if !sv.state.Has(key) {
		sv.errors = append(sv.errors, fmt.Sprintf("missing required key: %s", key))
	}
	return sv
}

// HasKeys validates that multiple keys exist
func (sv *StateValidator) HasKeys(keys ...string) *StateValidator {
	for _, key := range keys {
		sv.HasKey(key)
	}
	return sv
}

// HasValue validates that a key has a specific value
func (sv *StateValidator) HasValue(key string, expected interface{}) *StateValidator {
	value, exists := sv.state.Get(key)
	if !exists {
		sv.errors = append(sv.errors, fmt.Sprintf("key %s does not exist", key))
	} else if !reflect.DeepEqual(value, expected) {
		sv.errors = append(sv.errors, fmt.Sprintf("key %s: expected %v, got %v", key, expected, value))
	}
	return sv
}

// HasType validates that a key has a specific type
func (sv *StateValidator) HasType(key string, expectedType reflect.Type) *StateValidator {
	value, exists := sv.state.Get(key)
	if !exists {
		sv.errors = append(sv.errors, fmt.Sprintf("key %s does not exist", key))
	} else if reflect.TypeOf(value) != expectedType {
		sv.errors = append(sv.errors, fmt.Sprintf("key %s: expected type %v, got %v",
			key, expectedType, reflect.TypeOf(value)))
	}
	return sv
}

// HasArtifact validates that an artifact exists
func (sv *StateValidator) HasArtifact(id string) *StateValidator {
	if _, exists := sv.state.GetArtifact(id); !exists {
		sv.errors = append(sv.errors, fmt.Sprintf("missing artifact: %s", id))
	}
	return sv
}

// HasMessageCount validates the message count
func (sv *StateValidator) HasMessageCount(expected int) *StateValidator {
	actual := len(sv.state.Messages())
	if actual != expected {
		sv.errors = append(sv.errors, fmt.Sprintf("expected %d messages, got %d", expected, actual))
	}
	return sv
}

// IsValid returns true if no validation errors occurred
func (sv *StateValidator) IsValid() bool {
	return len(sv.errors) == 0
}

// GetErrors returns all validation errors
func (sv *StateValidator) GetErrors() []string {
	return sv.errors
}

// String returns a string representation of all errors
func (sv *StateValidator) String() string {
	if sv.IsValid() {
		return "State validation passed"
	}
	return "State validation failures:\n" + strings.Join(sv.errors, "\n")
}

// Common state creation helpers

// CreateStateWithData creates a state with initial data
func CreateStateWithData(data map[string]interface{}) *domain.State {
	state := domain.NewState()
	for k, v := range data {
		state.Set(k, v)
	}
	return state
}

// CreateStateWithMessages creates a state with messages
func CreateStateWithMessages(messages ...domain.Message) *domain.State {
	state := domain.NewState()
	for _, msg := range messages {
		state.AddMessage(msg)
	}
	return state
}

// CreateStateWithArtifacts creates a state with artifacts
func CreateStateWithArtifacts(artifacts ...*domain.Artifact) *domain.State {
	state := domain.NewState()
	for _, artifact := range artifacts {
		state.AddArtifact(artifact)
	}
	return state
}
