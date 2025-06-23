// ABOUTME: Implements StateReader interface to provide read-only access to State
// ABOUTME: Ensures tools cannot modify agent state while providing full read access

package domain

// stateReaderImpl wraps State to provide read-only access.
// This implementation ensures tools cannot modify agent state
// while providing full read access to values, artifacts, and metadata.
type stateReaderImpl struct {
	state *State
}

// NewStateReader creates a new StateReader from a State.
// Returns a read-only wrapper that implements the StateReader interface
// for safe state access in tool execution contexts.
func NewStateReader(state *State) StateReader {
	return &stateReaderImpl{state: state}
}

// Get retrieves a value from the state
func (sr *stateReaderImpl) Get(key string) (interface{}, bool) {
	return sr.state.Get(key)
}

// Values returns a copy of all values in the state
func (sr *stateReaderImpl) Values() map[string]interface{} {
	return sr.state.Values()
}

// GetArtifact retrieves an artifact by ID
func (sr *stateReaderImpl) GetArtifact(id string) (*Artifact, bool) {
	return sr.state.GetArtifact(id)
}

// Artifacts returns all artifacts in the state
func (sr *stateReaderImpl) Artifacts() map[string]*Artifact {
	return sr.state.Artifacts()
}

// Messages returns a copy of all messages
func (sr *stateReaderImpl) Messages() []Message {
	return sr.state.Messages()
}

// GetMetadata retrieves a metadata value
func (sr *stateReaderImpl) GetMetadata(key string) (interface{}, bool) {
	return sr.state.GetMetadata(key)
}

// Has checks if a key exists in the state
func (sr *stateReaderImpl) Has(key string) bool {
	_, exists := sr.state.Get(key)
	return exists
}

// Keys returns all keys in the state
func (sr *stateReaderImpl) Keys() []string {
	values := sr.state.Values()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	return keys
}
