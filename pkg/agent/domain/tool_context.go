// ABOUTME: Defines ToolContext for enhanced tool execution with access to agent state and events
// ABOUTME: Provides interfaces for state reading, event emission, and agent metadata access

package domain

import (
	"context"
	"time"
)

// ToolContext provides rich execution context for tools during execution.
// It combines Go context with agent state access, event emission capabilities,
// and metadata about the current execution for comprehensive tool support.
type ToolContext struct {
	// Standard Go context for cancellation and deadlines
	Context context.Context

	// Read-only access to agent state
	State StateReader

	// Metadata about the current execution
	RunID     string
	Retry     int
	StartTime time.Time

	// Event emission capability
	Events EventEmitter

	// Information about the calling agent
	Agent AgentInfo
}

// StateReader provides read-only access to agent state
type StateReader interface {
	// Get retrieves a value from the state
	Get(key string) (interface{}, bool)

	// Values returns a copy of all values in the state
	Values() map[string]interface{}

	// GetArtifact retrieves an artifact by ID
	GetArtifact(id string) (*Artifact, bool)

	// Artifacts returns all artifacts in the state
	Artifacts() map[string]*Artifact

	// Messages returns a copy of all messages
	Messages() []Message

	// GetMetadata retrieves a metadata value
	GetMetadata(key string) (interface{}, bool)

	// Has checks if a key exists in the state
	Has(key string) bool

	// Keys returns all keys in the state
	Keys() []string
}

// EventEmitter allows tools to emit events
type EventEmitter interface {
	// Emit sends an event
	Emit(eventType EventType, data interface{})

	// EmitProgress sends a progress event
	EmitProgress(current, total int, message string)

	// EmitMessage sends a message event
	EmitMessage(message string)

	// EmitError sends an error event
	EmitError(err error)

	// EmitCustom sends a custom event
	EmitCustom(eventName string, data interface{})
}

// AgentInfo provides information about the calling agent
type AgentInfo struct {
	// Agent identification
	ID          string
	Name        string
	Description string
	Type        AgentType

	// Agent hierarchy
	ParentID   string
	ParentName string
	Depth      int // Depth in agent hierarchy

	// Additional metadata
	Metadata map[string]interface{}
}

// NewToolContext creates a new tool context
func NewToolContext(ctx context.Context, state StateReader, agent BaseAgent, runID string) *ToolContext {
	return &ToolContext{
		Context:   ctx,
		State:     state,
		RunID:     runID,
		Retry:     0,
		StartTime: time.Now(),
		Agent: AgentInfo{
			ID:          agent.ID(),
			Name:        agent.Name(),
			Description: agent.Description(),
			Type:        agent.Type(),
			Metadata:    agent.Metadata(),
		},
	}
}

// WithRetry creates a new context with updated retry count
func (tc *ToolContext) WithRetry(retry int) *ToolContext {
	newTC := *tc
	newTC.Retry = retry
	return &newTC
}

// WithEventEmitter sets the event emitter
func (tc *ToolContext) WithEventEmitter(emitter EventEmitter) *ToolContext {
	newTC := *tc
	newTC.Events = emitter
	return &newTC
}

// Deadline returns the context deadline
func (tc *ToolContext) Deadline() (deadline time.Time, ok bool) {
	return tc.Context.Deadline()
}

// Done returns the context's done channel
func (tc *ToolContext) Done() <-chan struct{} {
	return tc.Context.Done()
}

// Err returns the context's error
func (tc *ToolContext) Err() error {
	return tc.Context.Err()
}

// Value returns a context value
func (tc *ToolContext) Value(key interface{}) interface{} {
	return tc.Context.Value(key)
}

// ElapsedTime returns how long the tool has been executing
func (tc *ToolContext) ElapsedTime() time.Duration {
	return time.Since(tc.StartTime)
}
