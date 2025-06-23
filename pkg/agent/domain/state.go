// ABOUTME: Defines the State structure for passing data between agents during execution
// ABOUTME: Provides thread-safe state management with values, artifacts, and message history

package domain

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// State represents the execution state passed between agents.
// It provides thread-safe storage for values, artifacts, and message history
// with automatic versioning and lineage tracking for debugging.
type State struct {
	mu       sync.RWMutex
	id       string
	created  time.Time
	modified time.Time

	// Core state data
	values    map[string]interface{}
	artifacts map[string]*Artifact
	messages  []Message

	// Metadata
	metadata map[string]interface{}

	// State lineage
	parentID string
	version  int
}

// NewState creates a new state instance with unique ID.
// Initializes empty collections for values, artifacts, and messages
// with automatic timestamp tracking and version 1.
func NewState() *State {
	return &State{
		id:        uuid.New().String(),
		created:   time.Now(),
		modified:  time.Now(),
		values:    make(map[string]interface{}),
		artifacts: make(map[string]*Artifact),
		messages:  make([]Message, 0),
		metadata:  make(map[string]interface{}),
		version:   1,
	}
}

// ID returns the state's unique identifier
func (s *State) ID() string {
	return s.id
}

// Created returns when the state was created
func (s *State) Created() time.Time {
	return s.created
}

// Modified returns when the state was last modified
func (s *State) Modified() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modified
}

// Get retrieves a value from the state
func (s *State) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.values[key]
	return val, ok
}

// Set stores a value in the state
func (s *State) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
	s.modified = time.Now()
	s.version++
}

// Delete removes a value from the state
func (s *State) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, key)
	s.modified = time.Now()
	s.version++
}

// Has checks if a key exists in the state
func (s *State) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.values[key]
	return exists
}

// Keys returns all keys in the state
func (s *State) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.values))
	for k := range s.values {
		keys = append(keys, k)
	}
	return keys
}

// Values returns a copy of all values in the state
func (s *State) Values() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]interface{})
	for k, v := range s.values {
		result[k] = v
	}
	return result
}

// AddArtifact adds an artifact to the state
func (s *State) AddArtifact(artifact *Artifact) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.artifacts[artifact.ID] = artifact
	s.modified = time.Now()
}

// GetArtifact retrieves an artifact by ID
func (s *State) GetArtifact(id string) (*Artifact, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	artifact, ok := s.artifacts[id]
	return artifact, ok
}

// Artifacts returns all artifacts in the state
func (s *State) Artifacts() map[string]*Artifact {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*Artifact)
	for k, v := range s.artifacts {
		result[k] = v
	}
	return result
}

// AddMessage adds a message to the conversation history
func (s *State) AddMessage(message Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, message)
	s.modified = time.Now()
}

// Messages returns a copy of all messages
func (s *State) Messages() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Message{}, s.messages...)
}

// SetMetadata sets a metadata value
func (s *State) SetMetadata(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metadata[key] = value
	s.modified = time.Now()
}

// GetMetadata retrieves a metadata value
func (s *State) GetMetadata(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.metadata[key]
	return val, ok
}

// GetAllMetadata returns a copy of all metadata
func (s *State) GetAllMetadata() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]interface{})
	for k, v := range s.metadata {
		result[k] = v
	}
	return result
}

// Version returns the current version number
func (s *State) Version() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// ParentID returns the ID of the parent state (if cloned)
func (s *State) ParentID() string {
	return s.parentID
}

// Clone creates a deep copy of the state
func (s *State) Clone() *State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	newState := &State{
		id:        uuid.New().String(),
		created:   time.Now(),
		modified:  time.Now(),
		parentID:  s.id,
		version:   1,
		values:    make(map[string]interface{}),
		artifacts: make(map[string]*Artifact),
		messages:  make([]Message, len(s.messages)),
		metadata:  make(map[string]interface{}),
	}

	// Deep copy values using JSON marshaling for safety
	for k, v := range s.values {
		newState.values[k] = deepCopyValue(v)
	}

	// Copy artifacts (shallow copy, artifacts are immutable)
	for k, v := range s.artifacts {
		newState.artifacts[k] = v
	}

	// Copy messages
	copy(newState.messages, s.messages)

	// Copy metadata
	for k, v := range s.metadata {
		newState.metadata[k] = deepCopyValue(v)
	}

	return newState
}

// Merge merges another state into this state
func (s *State) Merge(other *State) {
	if other == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	// Merge values (other overwrites)
	for k, v := range other.values {
		s.values[k] = v
	}

	// Merge artifacts
	for k, v := range other.artifacts {
		s.artifacts[k] = v
	}

	// Append messages
	s.messages = append(s.messages, other.messages...)

	// Merge metadata
	for k, v := range other.metadata {
		s.metadata[k] = v
	}

	s.modified = time.Now()
	s.version++
}

// MarshalJSON implements json.Marshaler
func (s *State) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return json.Marshal(map[string]interface{}{
		"id":        s.id,
		"created":   s.created,
		"modified":  s.modified,
		"values":    s.values,
		"artifacts": s.artifacts,
		"messages":  s.messages,
		"metadata":  s.metadata,
		"parent_id": s.parentID,
		"version":   s.version,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (s *State) UnmarshalJSON(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var temp struct {
		ID        string                 `json:"id"`
		Created   time.Time              `json:"created"`
		Modified  time.Time              `json:"modified"`
		Values    map[string]interface{} `json:"values"`
		Artifacts map[string]*Artifact   `json:"artifacts"`
		Messages  []Message              `json:"messages"`
		Metadata  map[string]interface{} `json:"metadata"`
		ParentID  string                 `json:"parent_id"`
		Version   int                    `json:"version"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	s.id = temp.ID
	s.created = temp.Created
	s.modified = temp.Modified
	s.values = temp.Values
	s.artifacts = temp.Artifacts
	s.messages = temp.Messages
	s.metadata = temp.Metadata
	s.parentID = temp.ParentID
	s.version = temp.Version

	// Initialize maps if nil
	if s.values == nil {
		s.values = make(map[string]interface{})
	}
	if s.artifacts == nil {
		s.artifacts = make(map[string]*Artifact)
	}
	if s.messages == nil {
		s.messages = make([]Message, 0)
	}
	if s.metadata == nil {
		s.metadata = make(map[string]interface{})
	}

	return nil
}

// Role represents the role of a message sender
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a conversation message
type Message struct {
	Role      Role                   `json:"role"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewMessage creates a new message
func NewMessage(role Role, content string) Message {
	return Message{
		Role:      role,
		Content:   content,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// deepCopyValue performs a deep copy of a value using JSON marshaling
func deepCopyValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	// For simple types, return as is
	switch v.(type) {
	case string, int, int32, int64, float32, float64, bool:
		return v
	}

	// For complex types, use JSON marshaling
	data, err := json.Marshal(v)
	if err != nil {
		return v // Return original on error
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return v // Return original on error
	}

	return result
}
