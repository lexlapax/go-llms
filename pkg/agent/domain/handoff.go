// ABOUTME: Defines handoff mechanism for delegating work between agents
// ABOUTME: Provides interfaces for input transformation and message filtering during handoffs

package domain

import (
	"context"
	"fmt"
)

// globalAgentRegistry holds the global agent registry instance
// This is set by the core package during initialization
var globalAgentRegistry AgentRegistry

// SetGlobalAgentRegistry sets the global agent registry.
// This should be called by the core package during initialization
// to enable handoff operations between agents.
func SetGlobalAgentRegistry(registry AgentRegistry) {
	globalAgentRegistry = registry
}

// GetGlobalAgentRegistry returns the global agent registry.
// Used by handoff implementations to look up target agents.
func GetGlobalAgentRegistry() AgentRegistry {
	return globalAgentRegistry
}

// Handoff represents a delegation mechanism between agents.
// Handoffs enable one agent to delegate work to another agent
// with input transformation and message filtering capabilities.
type Handoff interface {
	// Core identification
	Name() string
	Description() string
	TargetAgent() string

	// Handoff execution
	Execute(ctx context.Context, state *State) (*State, error)

	// Input transformation
	TransformInput(state *State) *State
	FilterMessages(messages []Message) []Message
}

// HandoffBuilder provides fluent configuration
type HandoffBuilder struct {
	name          string
	targetAgent   string
	description   string
	inputFilter   func(*State) *State
	messageFilter func([]Message) []Message
}

// NewHandoffBuilder creates a new handoff builder
func NewHandoffBuilder(name, targetAgent string) *HandoffBuilder {
	return &HandoffBuilder{
		name:        name,
		targetAgent: targetAgent,
	}
}

// WithDescription sets the handoff description
func (hb *HandoffBuilder) WithDescription(desc string) *HandoffBuilder {
	hb.description = desc
	return hb
}

// WithInputFilter sets the input transformation function
func (hb *HandoffBuilder) WithInputFilter(filter func(*State) *State) *HandoffBuilder {
	hb.inputFilter = filter
	return hb
}

// WithMessageFilter sets the message filtering function
func (hb *HandoffBuilder) WithMessageFilter(filter func([]Message) []Message) *HandoffBuilder {
	hb.messageFilter = filter
	return hb
}

// Build creates the handoff instance
func (hb *HandoffBuilder) Build() Handoff {
	return &handoffImpl{
		name:          hb.name,
		targetAgent:   hb.targetAgent,
		description:   hb.description,
		inputFilter:   hb.inputFilter,
		messageFilter: hb.messageFilter,
	}
}

// handoffImpl is the default implementation of Handoff
type handoffImpl struct {
	name          string
	targetAgent   string
	description   string
	inputFilter   func(*State) *State
	messageFilter func([]Message) []Message
}

// Name returns the handoff name
func (h *handoffImpl) Name() string {
	return h.name
}

// Description returns the handoff description
func (h *handoffImpl) Description() string {
	return h.description
}

// TargetAgent returns the target agent name
func (h *handoffImpl) TargetAgent() string {
	return h.targetAgent
}

// Execute performs the handoff
func (h *handoffImpl) Execute(ctx context.Context, state *State) (*State, error) {
	// Transform input state
	transformedState := h.TransformInput(state)

	// Get the global registry to find the target agent
	registry := GetGlobalAgentRegistry()
	if registry == nil {
		return nil, fmt.Errorf("global agent registry not available")
	}

	// Find the target agent by name
	targetAgent, err := registry.GetByName(h.targetAgent)
	if err != nil {
		return nil, fmt.Errorf("target agent '%s' not found: %w", h.targetAgent, err)
	}

	// Create a new run context for the target agent
	// Note: The target agent will create its own RunContext based on its requirements
	// We just pass the transformed state

	// Execute the target agent
	resultState, err := targetAgent.Run(ctx, transformedState)
	if err != nil {
		return nil, fmt.Errorf("handoff to agent '%s' failed: %w", h.targetAgent, err)
	}

	// Optionally merge results back to original state structure
	// For now, we return the result state as-is
	// Future enhancement: could add result transformation/merging

	return resultState, nil
}

// TransformInput applies the input filter to the state
func (h *handoffImpl) TransformInput(state *State) *State {
	if h.inputFilter != nil {
		return h.inputFilter(state)
	}
	// Default: return a clone of the state
	return state.Clone()
}

// FilterMessages applies the message filter
func (h *handoffImpl) FilterMessages(messages []Message) []Message {
	if h.messageFilter != nil {
		return h.messageFilter(messages)
	}
	// Default: return all messages
	return messages
}

// Common handoff patterns

// NewSimpleHandoff creates a handoff that passes state unchanged
func NewSimpleHandoff(name, targetAgent string) Handoff {
	return NewHandoffBuilder(name, targetAgent).
		WithDescription(fmt.Sprintf("Simple handoff to %s", targetAgent)).
		Build()
}

// NewFilteredHandoff creates a handoff that filters specific keys
func NewFilteredHandoff(name, targetAgent string, keepKeys ...string) Handoff {
	keyMap := make(map[string]bool)
	for _, key := range keepKeys {
		keyMap[key] = true
	}

	return NewHandoffBuilder(name, targetAgent).
		WithDescription(fmt.Sprintf("Filtered handoff to %s", targetAgent)).
		WithInputFilter(func(state *State) *State {
			filtered := NewState()
			for key, value := range state.Values() {
				if keyMap[key] {
					filtered.Set(key, value)
				}
			}
			// Note: metadata copying would require access to internal state
			// For now, we'll skip metadata copying in filtered handoffs
			return filtered
		}).
		Build()
}

// NewMessagesOnlyHandoff creates a handoff that only passes messages
func NewMessagesOnlyHandoff(name, targetAgent string) Handoff {
	return NewHandoffBuilder(name, targetAgent).
		WithDescription(fmt.Sprintf("Messages-only handoff to %s", targetAgent)).
		WithInputFilter(func(state *State) *State {
			newState := NewState()
			// Only copy messages
			for _, msg := range state.Messages() {
				newState.AddMessage(msg)
			}
			return newState
		}).
		Build()
}

// NewLastNMessagesHandoff creates a handoff that only passes the last N messages
func NewLastNMessagesHandoff(name, targetAgent string, n int) Handoff {
	return NewHandoffBuilder(name, targetAgent).
		WithDescription(fmt.Sprintf("Last %d messages handoff to %s", n, targetAgent)).
		WithMessageFilter(func(messages []Message) []Message {
			if len(messages) <= n {
				return messages
			}
			return messages[len(messages)-n:]
		}).
		WithInputFilter(func(state *State) *State {
			// Clear messages and add filtered ones
			var filteredMessages []Message
			if len(state.Messages()) > n {
				filteredMessages = state.Messages()[len(state.Messages())-n:]
			} else {
				filteredMessages = state.Messages()
			}
			// Create new state with filtered messages
			newState := NewState()
			for key, value := range state.Values() {
				newState.Set(key, value)
			}
			for _, msg := range filteredMessages {
				newState.AddMessage(msg)
			}
			return newState
		}).
		Build()
}
