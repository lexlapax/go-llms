// ABOUTME: Provides a global registry for agent discovery and management
// ABOUTME: Enables agents to find and interact with other agents in the system

package core

import (
	"fmt"
	"sync"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// init sets up the global registry in the domain package
func init() {
	// Set the global registry in the domain package to enable handoffs
	domain.SetGlobalAgentRegistry(globalRegistry)
}

// AgentRegistry manages registered agents for discovery and coordination.
// It maintains indexes by ID and type, tracks parent-child relationships,
// and provides thread-safe access to agent instances.
type AgentRegistry struct {
	mu     sync.RWMutex
	agents map[string]domain.BaseAgent
	// Index by type for faster lookups
	agentsByType map[domain.AgentType]map[string]domain.BaseAgent
	// Parent-child relationships
	children map[string][]string
}

// globalRegistry is the singleton instance used throughout the application.
// It enables agents to discover and interact with each other.
var globalRegistry = NewAgentRegistry()

// GetGlobalRegistry returns the global agent registry instance.
// This registry is shared across the application for agent discovery.
func GetGlobalRegistry() *AgentRegistry {
	return globalRegistry
}

// NewAgentRegistry creates a new agent registry instance.
// The registry starts empty and agents can be registered using the Register method.
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents:       make(map[string]domain.BaseAgent),
		agentsByType: make(map[domain.AgentType]map[string]domain.BaseAgent),
		children:     make(map[string][]string),
	}
}

// Register registers an agent with the registry.
// The agent must have a unique ID. If an agent with the same ID already exists,
// it will be replaced. The registry maintains indexes by type for efficient lookups.
func (r *AgentRegistry) Register(agent domain.BaseAgent) error {
	if agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := agent.ID()
	if _, exists := r.agents[id]; exists {
		return fmt.Errorf("agent with ID %s already registered", id)
	}

	// Add to main registry
	r.agents[id] = agent

	// Add to type index
	agentType := agent.Type()
	if r.agentsByType[agentType] == nil {
		r.agentsByType[agentType] = make(map[string]domain.BaseAgent)
	}
	r.agentsByType[agentType][id] = agent

	// Register children recursively
	for _, child := range agent.SubAgents() {
		if err := r.Register(child); err != nil {
			// Continue on error to register other children
			continue
		}
		r.children[id] = append(r.children[id], child.ID())
	}

	return nil
}

// Unregister removes an agent from the registry.
// It also removes all child agents recursively and updates parent-child relationships.
// Returns an error if the agent is not found.
func (r *AgentRegistry) Unregister(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	// Remove from main registry
	delete(r.agents, agentID)

	// Remove from type index
	agentType := agent.Type()
	if typeMap, ok := r.agentsByType[agentType]; ok {
		delete(typeMap, agentID)
		if len(typeMap) == 0 {
			delete(r.agentsByType, agentType)
		}
	}

	// Remove children recursively
	if childIDs, ok := r.children[agentID]; ok {
		for _, childID := range childIDs {
			r.unregisterInternal(childID)
		}
		delete(r.children, agentID)
	}

	// Remove from parent's children list
	for parentID, childIDs := range r.children {
		newChildren := make([]string, 0, len(childIDs))
		for _, childID := range childIDs {
			if childID != agentID {
				newChildren = append(newChildren, childID)
			}
		}
		if len(newChildren) > 0 {
			r.children[parentID] = newChildren
		} else {
			delete(r.children, parentID)
		}
	}

	return nil
}

// unregisterInternal removes an agent without lock (must be called with lock held).
// This is used internally for recursive removal of child agents.
func (r *AgentRegistry) unregisterInternal(agentID string) {
	agent, exists := r.agents[agentID]
	if !exists {
		return
	}

	// Remove from main registry
	delete(r.agents, agentID)

	// Remove from type index
	agentType := agent.Type()
	if typeMap, ok := r.agentsByType[agentType]; ok {
		delete(typeMap, agentID)
		if len(typeMap) == 0 {
			delete(r.agentsByType, agentType)
		}
	}

	// Remove children recursively
	if childIDs, ok := r.children[agentID]; ok {
		for _, childID := range childIDs {
			r.unregisterInternal(childID)
		}
		delete(r.children, agentID)
	}
}

// Get retrieves an agent by ID.
// Returns domain.ErrAgentNotFound if the agent doesn't exist.
func (r *AgentRegistry) Get(agentID string) (domain.BaseAgent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, domain.ErrAgentNotFound
	}

	return agent, nil
}

// GetByName retrieves an agent by name.
// Returns domain.ErrAgentNotFound if no agent with the given name exists.
// Note: Agent names are not guaranteed to be unique.
func (r *AgentRegistry) GetByName(name string) (domain.BaseAgent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, agent := range r.agents {
		if agent.Name() == name {
			return agent, nil
		}
	}

	return nil, domain.ErrAgentNotFound
}

// GetByType retrieves all agents of a specific type.
// Returns an empty slice if no agents of the given type are registered.
func (r *AgentRegistry) GetByType(agentType domain.AgentType) []domain.BaseAgent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	typeMap, ok := r.agentsByType[agentType]
	if !ok {
		return nil
	}

	agents := make([]domain.BaseAgent, 0, len(typeMap))
	for _, agent := range typeMap {
		agents = append(agents, agent)
	}

	return agents
}

// List returns all registered agents.
// The returned slice is a copy and can be safely modified.
func (r *AgentRegistry) List() []domain.BaseAgent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]domain.BaseAgent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}

	return agents
}

// ListIDs returns all registered agent IDs.
// The returned slice contains unique agent identifiers.
func (r *AgentRegistry) ListIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.agents))
	for id := range r.agents {
		ids = append(ids, id)
	}

	return ids
}

// GetChildren returns the direct children of an agent.
// Only returns immediate children, not descendants. Returns an empty slice
// if the agent has no children or doesn't exist.
func (r *AgentRegistry) GetChildren(agentID string) []domain.BaseAgent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	childIDs, ok := r.children[agentID]
	if !ok {
		return nil
	}

	children := make([]domain.BaseAgent, 0, len(childIDs))
	for _, childID := range childIDs {
		if child, exists := r.agents[childID]; exists {
			children = append(children, child)
		}
	}

	return children
}

// GetParent returns the parent of an agent.
// Returns domain.ErrAgentNotFound if the agent doesn't exist or has no parent.
func (r *AgentRegistry) GetParent(agentID string) (domain.BaseAgent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, domain.ErrAgentNotFound
	}

	parent := agent.Parent()
	if parent == nil {
		return nil, domain.ErrAgentNotFound
	}

	return parent, nil
}

// FindByMetadata finds agents with matching metadata.
// Returns all agents where the specified metadata key has the given value.
// Returns an empty slice if no matching agents are found.
func (r *AgentRegistry) FindByMetadata(key string, value interface{}) []domain.BaseAgent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []domain.BaseAgent
	for _, agent := range r.agents {
		metadata := agent.Metadata()
		if v, ok := metadata[key]; ok && v == value {
			matches = append(matches, agent)
		}
	}

	return matches
}

// Clear removes all agents from the registry
func (r *AgentRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.agents = make(map[string]domain.BaseAgent)
	r.agentsByType = make(map[domain.AgentType]map[string]domain.BaseAgent)
	r.children = make(map[string][]string)
}

// Size returns the number of registered agents
func (r *AgentRegistry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// Global registry functions

// Register registers an agent with the global registry
func Register(agent domain.BaseAgent) error {
	return globalRegistry.Register(agent)
}

// Unregister removes an agent from the global registry
func Unregister(agentID string) error {
	return globalRegistry.Unregister(agentID)
}

// Get retrieves an agent from the global registry by ID
func Get(agentID string) (domain.BaseAgent, error) {
	return globalRegistry.Get(agentID)
}

// GetByName retrieves an agent from the global registry by name
func GetByName(name string) (domain.BaseAgent, error) {
	return globalRegistry.GetByName(name)
}

// GetByType retrieves agents from the global registry by type
func GetByType(agentType domain.AgentType) []domain.BaseAgent {
	return globalRegistry.GetByType(agentType)
}

// List returns all agents in the global registry
func List() []domain.BaseAgent {
	return globalRegistry.List()
}

// Clear clears the global registry
func Clear() {
	globalRegistry.Clear()
}

// AgentQuery provides a fluent interface for querying agents
type AgentQuery struct {
	registry *AgentRegistry
	filters  []func(domain.BaseAgent) bool
}

// NewQuery creates a new agent query
func (r *AgentRegistry) NewQuery() *AgentQuery {
	return &AgentQuery{
		registry: r,
		filters:  make([]func(domain.BaseAgent) bool, 0),
	}
}

// WithType filters by agent type
func (q *AgentQuery) WithType(agentType domain.AgentType) *AgentQuery {
	q.filters = append(q.filters, func(agent domain.BaseAgent) bool {
		return agent.Type() == agentType
	})
	return q
}

// WithName filters by agent name
func (q *AgentQuery) WithName(name string) *AgentQuery {
	q.filters = append(q.filters, func(agent domain.BaseAgent) bool {
		return agent.Name() == name
	})
	return q
}

// WithMetadata filters by metadata
func (q *AgentQuery) WithMetadata(key string, value interface{}) *AgentQuery {
	q.filters = append(q.filters, func(agent domain.BaseAgent) bool {
		metadata := agent.Metadata()
		if v, ok := metadata[key]; ok {
			return v == value
		}
		return false
	})
	return q
}

// WithParent filters by parent agent
func (q *AgentQuery) WithParent(parentID string) *AgentQuery {
	q.filters = append(q.filters, func(agent domain.BaseAgent) bool {
		parent := agent.Parent()
		return parent != nil && parent.ID() == parentID
	})
	return q
}

// Execute runs the query and returns matching agents
func (q *AgentQuery) Execute() []domain.BaseAgent {
	q.registry.mu.RLock()
	defer q.registry.mu.RUnlock()

	var results []domain.BaseAgent
	for _, agent := range q.registry.agents {
		match := true
		for _, filter := range q.filters {
			if !filter(agent) {
				match = false
				break
			}
		}
		if match {
			results = append(results, agent)
		}
	}

	return results
}

// Count returns the number of matching agents
func (q *AgentQuery) Count() int {
	return len(q.Execute())
}
