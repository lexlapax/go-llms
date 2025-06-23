// ABOUTME: Manages state lifecycle, transformations, and persistence for agents
// ABOUTME: Provides utilities for state merging, validation, and snapshot management

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// StateManager manages state lifecycle and transformations for agents.
// It provides thread-safe storage for state snapshots, applies transformations,
// validates state according to rules, and supports state merging strategies.
type StateManager struct {
	mu         sync.RWMutex
	states     map[string]*domain.State
	transforms map[string]StateTransform
	validators map[string]StateValidator
}

// StateTransform defines a state transformation function.
// Transformations are applied to modify state before or after agent execution.
// Common uses include filtering sensitive data, enriching with metadata, or
// converting between formats.
type StateTransform func(ctx context.Context, input *domain.State) (*domain.State, error)

// StateValidator validates state according to rules.
// Validators ensure state meets requirements before processing.
// Return an error if validation fails.
type StateValidator func(state *domain.State) error

// MergeFunc defines a custom merge function for combining multiple states.
// Used when merging states from parent/child agents or multiple sources.
// The function should handle conflicts and produce a coherent merged state.
type MergeFunc func(states []*domain.State) (*domain.State, error)

// NewStateManager creates a new state manager instance.
// It initializes empty collections and registers built-in transforms
// for common state operations like filtering and enrichment.
func NewStateManager() *StateManager {
	sm := &StateManager{
		states:     make(map[string]*domain.State),
		transforms: make(map[string]StateTransform),
		validators: make(map[string]StateValidator),
	}

	// Register built-in transforms
	sm.registerBuiltinTransforms()

	return sm
}

// SaveState stores a state snapshot in the manager.
// The state is cloned to prevent external modifications.
// Thread-safe for concurrent access.
func (sm *StateManager) SaveState(state *domain.State) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.states[state.ID()] = state.Clone()
	return nil
}

// LoadState retrieves a state snapshot by ID.
// Returns a clone of the stored state to prevent modifications.
// Returns an error if the state doesn't exist.
func (sm *StateManager) LoadState(id string) (*domain.State, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	state, ok := sm.states[id]
	if !ok {
		return nil, fmt.Errorf("state %s not found", id)
	}

	return state.Clone(), nil
}

// DeleteState removes a state snapshot from storage.
// Returns an error if the state doesn't exist.
// Thread-safe for concurrent access.
func (sm *StateManager) DeleteState(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.states[id]; !ok {
		return fmt.Errorf("state %s not found", id)
	}

	delete(sm.states, id)
	return nil
}

// ListStates returns all stored state IDs
func (sm *StateManager) ListStates() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ids := make([]string, 0, len(sm.states))
	for id := range sm.states {
		ids = append(ids, id)
	}
	return ids
}

// RegisterTransform registers a state transformation
func (sm *StateManager) RegisterTransform(name string, transform StateTransform) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.transforms[name] = transform
}

// ApplyTransform applies a named transformation to a state
func (sm *StateManager) ApplyTransform(ctx context.Context, name string, state *domain.State) (*domain.State, error) {
	sm.mu.RLock()
	transform, ok := sm.transforms[name]
	sm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("transform %s not found", name)
	}

	return transform(ctx, state)
}

// RegisterValidator registers a state validator
func (sm *StateManager) RegisterValidator(name string, validator StateValidator) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.validators[name] = validator
}

// ValidateState validates a state using a named validator
func (sm *StateManager) ValidateState(name string, state *domain.State) error {
	sm.mu.RLock()
	validator, ok := sm.validators[name]
	sm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("validator %s not found", name)
	}

	return validator(state)
}

// MergeStates merges multiple states according to a strategy
func (sm *StateManager) MergeStates(states []*domain.State, strategy domain.MergeStrategy) (*domain.State, error) {
	if len(states) == 0 {
		return nil, fmt.Errorf("no states to merge")
	}

	switch strategy {
	case domain.MergeStrategyLast:
		return MergeStrategyLast(states)
	case domain.MergeStrategyMergeAll:
		return MergeStrategyMergeAll(states)
	case domain.MergeStrategyUnion:
		return MergeStrategyUnion(states)
	default:
		return nil, fmt.Errorf("unknown merge strategy: %s", strategy)
	}
}

// Built-in merge strategies

// MergeStrategyLast takes the last state
func MergeStrategyLast(states []*domain.State) (*domain.State, error) {
	if len(states) == 0 {
		return nil, fmt.Errorf("no states provided")
	}
	return states[len(states)-1].Clone(), nil
}

// MergeStrategyMergeAll merges all states in order
func MergeStrategyMergeAll(states []*domain.State) (*domain.State, error) {
	if len(states) == 0 {
		return nil, fmt.Errorf("no states provided")
	}

	result := domain.NewState()
	for _, state := range states {
		if state != nil {
			result.Merge(state)
		}
	}
	return result, nil
}

// MergeStrategyUnion creates a union of all values
func MergeStrategyUnion(states []*domain.State) (*domain.State, error) {
	if len(states) == 0 {
		return nil, fmt.Errorf("no states provided")
	}

	result := domain.NewState()

	// Collect all unique keys and their values
	valueMap := make(map[string][]interface{})
	for _, state := range states {
		if state == nil {
			continue
		}

		for k, v := range state.Values() {
			valueMap[k] = append(valueMap[k], v)
		}
	}

	// Store arrays of values for keys that appear in multiple states
	for k, values := range valueMap {
		if len(values) == 1 {
			result.Set(k, values[0])
		} else {
			// Remove duplicates for simple types
			unique := removeDuplicates(values)
			if len(unique) == 1 {
				result.Set(k, unique[0])
			} else {
				result.Set(k, unique)
			}
		}
	}

	// Merge artifacts from all states
	for _, state := range states {
		if state == nil {
			continue
		}

		for _, artifact := range state.Artifacts() {
			result.AddArtifact(artifact)
		}
	}

	// Combine all messages
	for _, state := range states {
		if state == nil {
			continue
		}

		for _, msg := range state.Messages() {
			result.AddMessage(msg)
		}
	}

	return result, nil
}

// Built-in transforms

// registerBuiltinTransforms registers built-in state transformations
func (sm *StateManager) registerBuiltinTransforms() {
	// Filter transform - removes specified keys
	sm.RegisterTransform("filter", func(ctx context.Context, state *domain.State) (*domain.State, error) {
		// Get filter keys from state metadata
		filterKeys, ok := state.GetMetadata("filter_keys")
		if !ok {
			return state, nil
		}

		keys, ok := filterKeys.([]string)
		if !ok {
			return state, nil
		}

		result := state.Clone()
		for _, key := range keys {
			result.Delete(key)
		}

		return result, nil
	})

	// Flatten transform - flattens nested structures
	sm.RegisterTransform("flatten", func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := domain.NewState()

		for k, v := range state.Values() {
			flattened := flattenValue(k, v)
			for fk, fv := range flattened {
				result.Set(fk, fv)
			}
		}

		// Copy artifacts and messages
		for _, artifact := range state.Artifacts() {
			result.AddArtifact(artifact)
		}
		for _, msg := range state.Messages() {
			result.AddMessage(msg)
		}

		return result, nil
	})

	// Sanitize transform - removes sensitive data
	sm.RegisterTransform("sanitize", func(ctx context.Context, state *domain.State) (*domain.State, error) {
		sensitiveKeys := []string{"password", "token", "secret", "key", "api_key", "credential"}

		result := state.Clone()
		for k := range result.Values() {
			for _, sensitive := range sensitiveKeys {
				if containsIgnoreCase(k, sensitive) {
					result.Set(k, "[REDACTED]")
					break
				}
			}
		}

		return result, nil
	})
}

// Helper functions

// removeDuplicates removes duplicate values from a slice
func removeDuplicates(values []interface{}) []interface{} {
	seen := make(map[string]bool)
	result := make([]interface{}, 0, len(values))

	for _, v := range values {
		// Use JSON encoding as a simple way to compare values
		key, err := json.Marshal(v)
		if err != nil {
			// If we can't marshal, include it
			result = append(result, v)
			continue
		}

		keyStr := string(key)
		if !seen[keyStr] {
			seen[keyStr] = true
			result = append(result, v)
		}
	}

	return result
}

// flattenValue recursively flattens a value
func flattenValue(prefix string, value interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	switch v := value.(type) {
	case map[string]interface{}:
		for k, val := range v {
			newKey := prefix + "." + k
			for fk, fv := range flattenValue(newKey, val) {
				result[fk] = fv
			}
		}
	case []interface{}:
		for i, val := range v {
			newKey := fmt.Sprintf("%s[%d]", prefix, i)
			for fk, fv := range flattenValue(newKey, val) {
				result[fk] = fv
			}
		}
	default:
		result[prefix] = value
	}

	return result
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			contains(toLower(s), toLower(substr)))
}

// Simple string utilities to avoid imports
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// StateSnapshot captures the current state of multiple agents
type StateSnapshot struct {
	Timestamp string                   `json:"timestamp"`
	States    map[string]*domain.State `json:"states"`
	Metadata  map[string]interface{}   `json:"metadata"`
}

// CreateSnapshot creates a snapshot of multiple states
func (sm *StateManager) CreateSnapshot(states map[string]*domain.State) *StateSnapshot {
	snapshot := &StateSnapshot{
		Timestamp: time.Now().Format(time.RFC3339),
		States:    make(map[string]*domain.State),
		Metadata:  make(map[string]interface{}),
	}

	for name, state := range states {
		if state != nil {
			snapshot.States[name] = state.Clone()
		}
	}

	return snapshot
}
