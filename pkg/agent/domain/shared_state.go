// ABOUTME: Defines SharedStateContext for parent-child state sharing in multi-agent systems
// ABOUTME: Enables child agents to inherit and selectively override parent state values

package domain

import (
	"sync"
)

// SharedStateContext provides a state context that shares data with a parent state.
// It allows child agents to inherit values from parent state while maintaining
// their own local modifications and overrides for hierarchical agent workflows.
type SharedStateContext struct {
	mu sync.RWMutex

	// Parent state (read-only access)
	parent StateReader

	// Local state (read-write access)
	local *State

	// Configuration
	inheritMessages  bool // Whether to inherit parent messages
	inheritArtifacts bool // Whether to inherit parent artifacts
	inheritMetadata  bool // Whether to inherit parent metadata
}

// NewSharedStateContext creates a new shared state context with parent state.
// By default inherits messages, artifacts, and metadata from the parent.
// Local state is initialized empty and can override parent values.
func NewSharedStateContext(parent StateReader) *SharedStateContext {
	return &SharedStateContext{
		parent:           parent,
		local:            NewState(),
		inheritMessages:  true,
		inheritArtifacts: true,
		inheritMetadata:  true,
	}
}

// WithInheritanceConfig configures which parent state components to inherit.
// Allows fine-grained control over messages, artifacts, and metadata inheritance.
// Returns the context for method chaining.
func (ssc *SharedStateContext) WithInheritanceConfig(messages, artifacts, metadata bool) *SharedStateContext {
	ssc.mu.Lock()
	defer ssc.mu.Unlock()

	ssc.inheritMessages = messages
	ssc.inheritArtifacts = artifacts
	ssc.inheritMetadata = metadata
	return ssc
}

// Get retrieves a value using hierarchical lookup.
// Checks local state first for overrides, then falls back to parent state.
// This implements the inheritance and override semantics.
func (ssc *SharedStateContext) Get(key string) (interface{}, bool) {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	// Check local state first
	if val, ok := ssc.local.Get(key); ok {
		return val, true
	}

	// Fall back to parent state
	if ssc.parent != nil {
		return ssc.parent.Get(key)
	}

	return nil, false
}

// Set sets a value in the local state only
func (ssc *SharedStateContext) Set(key string, value interface{}) {
	ssc.mu.Lock()
	defer ssc.mu.Unlock()

	ssc.local.Set(key, value)
}

// Values returns merged values from parent and local state
func (ssc *SharedStateContext) Values() map[string]interface{} {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	// Start with parent values
	merged := make(map[string]interface{})
	if ssc.parent != nil {
		for k, v := range ssc.parent.Values() {
			merged[k] = v
		}
	}

	// Override with local values
	for k, v := range ssc.local.Values() {
		merged[k] = v
	}

	return merged
}

// GetArtifact retrieves an artifact, checking local first, then parent
func (ssc *SharedStateContext) GetArtifact(id string) (*Artifact, bool) {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	// Check local artifacts first
	if artifact, ok := ssc.local.GetArtifact(id); ok {
		return artifact, true
	}

	// Fall back to parent artifacts if inheritance is enabled
	if ssc.inheritArtifacts && ssc.parent != nil {
		return ssc.parent.GetArtifact(id)
	}

	return nil, false
}

// Artifacts returns merged artifacts from parent and local state
func (ssc *SharedStateContext) Artifacts() map[string]*Artifact {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	merged := make(map[string]*Artifact)

	// Add parent artifacts if inheritance is enabled
	if ssc.inheritArtifacts && ssc.parent != nil {
		for id, artifact := range ssc.parent.Artifacts() {
			merged[id] = artifact
		}
	}

	// Override with local artifacts
	for id, artifact := range ssc.local.Artifacts() {
		merged[id] = artifact
	}

	return merged
}

// Messages returns merged messages from parent and local state
func (ssc *SharedStateContext) Messages() []Message {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	var messages []Message

	// Add parent messages if inheritance is enabled
	if ssc.inheritMessages && ssc.parent != nil {
		messages = append(messages, ssc.parent.Messages()...)
	}

	// Add local messages
	messages = append(messages, ssc.local.Messages()...)

	return messages
}

// GetMetadata retrieves metadata, checking local first, then parent
func (ssc *SharedStateContext) GetMetadata(key string) (interface{}, bool) {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	// Check local metadata first
	if val, ok := ssc.local.GetMetadata(key); ok {
		return val, true
	}

	// Fall back to parent metadata if inheritance is enabled
	if ssc.inheritMetadata && ssc.parent != nil {
		return ssc.parent.GetMetadata(key)
	}

	return nil, false
}

// Has checks if a key exists in local or parent state
func (ssc *SharedStateContext) Has(key string) bool {
	_, ok := ssc.Get(key)
	return ok
}

// Keys returns all keys from parent and local state
func (ssc *SharedStateContext) Keys() []string {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	keyMap := make(map[string]bool)

	// Add parent keys
	if ssc.parent != nil {
		for _, key := range ssc.parent.Keys() {
			keyMap[key] = true
		}
	}

	// Add local keys
	for _, key := range ssc.local.Keys() {
		keyMap[key] = true
	}

	// Convert to slice
	keys := make([]string, 0, len(keyMap))
	for key := range keyMap {
		keys = append(keys, key)
	}

	return keys
}

// LocalState returns the local state for direct access
func (ssc *SharedStateContext) LocalState() *State {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()
	return ssc.local
}

// MergeToParent merges local changes back to parent state (if parent is writable)
func (ssc *SharedStateContext) MergeToParent() error {
	ssc.mu.Lock()
	defer ssc.mu.Unlock()

	// For now, we can't merge back to parent since StateReader is read-only
	// This would require the parent to be a full State, not just StateReader
	// In the future, we could add a WritableState interface
	return ErrStateReadOnly
}

// Clone creates a new SharedStateContext with the same parent but fresh local state
func (ssc *SharedStateContext) Clone() *SharedStateContext {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	return &SharedStateContext{
		parent:           ssc.parent,
		local:            NewState(),
		inheritMessages:  ssc.inheritMessages,
		inheritArtifacts: ssc.inheritArtifacts,
		inheritMetadata:  ssc.inheritMetadata,
	}
}

// AsState converts the shared state context to a regular state by merging all data
func (ssc *SharedStateContext) AsState() *State {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	// Create new state with merged data
	state := NewState()

	// Copy all values
	for key, value := range ssc.Values() {
		state.Set(key, value)
	}

	// Copy all artifacts
	for _, artifact := range ssc.Artifacts() {
		state.AddArtifact(artifact)
	}

	// Copy all messages
	for _, message := range ssc.Messages() {
		state.AddMessage(message)
	}

	// Copy metadata from local state only
	// (we can't access parent metadata directly since it's a StateReader)
	localMeta := ssc.local.GetAllMetadata()
	for key, value := range localMeta {
		state.SetMetadata(key, value)
	}

	return state
}
