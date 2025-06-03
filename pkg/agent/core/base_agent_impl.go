// ABOUTME: Provides base implementation for agents with common functionality
// ABOUTME: Implements hierarchy management, configuration, and lifecycle methods

package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// BaseAgentImpl provides common functionality for all agents
type BaseAgentImpl struct {
	mu sync.RWMutex

	// Identity
	id          string
	name        string
	description string
	agentType   domain.AgentType

	// Hierarchy
	parent    domain.BaseAgent
	subAgents []domain.BaseAgent

	// Configuration
	config       domain.AgentConfig
	inputSchema  *sdomain.Schema
	outputSchema *sdomain.Schema

	// Metadata
	metadata map[string]interface{}

	// Event handling
	dispatcher domain.EventDispatcher

	// State
	initialized bool
}

// NewBaseAgent creates a new base agent implementation
func NewBaseAgent(name, description string, agentType domain.AgentType) *BaseAgentImpl {
	return &BaseAgentImpl{
		id:          uuid.New().String(),
		name:        name,
		description: description,
		agentType:   agentType,
		subAgents:   make([]domain.BaseAgent, 0),
		metadata:    make(map[string]interface{}),
		config: domain.AgentConfig{
			Timeout:    30 * time.Second,
			MaxRetries: 3,
			RetryDelay: time.Second,
			Custom:     make(map[string]interface{}),
		},
	}
}

// Identification methods

// ID returns the agent's unique identifier
func (a *BaseAgentImpl) ID() string {
	return a.id
}

// Name returns the agent's name
func (a *BaseAgentImpl) Name() string {
	return a.name
}

// Description returns the agent's description
func (a *BaseAgentImpl) Description() string {
	return a.description
}

// Type returns the agent's type
func (a *BaseAgentImpl) Type() domain.AgentType {
	return a.agentType
}

// Hierarchy management

// Parent returns the parent agent
func (a *BaseAgentImpl) Parent() domain.BaseAgent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.parent
}

// SetParent sets the parent agent
func (a *BaseAgentImpl) SetParent(parent domain.BaseAgent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check for circular dependencies
	if parent != nil && a.hasCircularDependency(parent) {
		return domain.ErrCircularDependency
	}

	a.parent = parent
	return nil
}

// SubAgents returns a copy of the sub-agents list
func (a *BaseAgentImpl) SubAgents() []domain.BaseAgent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return append([]domain.BaseAgent{}, a.subAgents...)
}

// AddSubAgent adds a sub-agent
func (a *BaseAgentImpl) AddSubAgent(agent domain.BaseAgent) error {
	if agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if agent already exists
	for _, existing := range a.subAgents {
		if existing.ID() == agent.ID() {
			return fmt.Errorf("agent with ID %s already exists", agent.ID())
		}
	}

	// Set parent
	if err := agent.SetParent(a); err != nil {
		return fmt.Errorf("failed to set parent: %w", err)
	}

	a.subAgents = append(a.subAgents, agent)
	return nil
}

// RemoveSubAgent removes a sub-agent by name
func (a *BaseAgentImpl) RemoveSubAgent(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i, agent := range a.subAgents {
		if agent.Name() == name {
			// Clear parent reference
			_ = agent.SetParent(nil)

			// Remove from slice
			a.subAgents = append(a.subAgents[:i], a.subAgents[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("agent %s not found", name)
}

// FindAgent searches for an agent by name in the hierarchy
func (a *BaseAgentImpl) FindAgent(name string) domain.BaseAgent {
	// Check self
	if a.name == name {
		return a
	}

	// Check sub-agents recursively
	for _, agent := range a.SubAgents() {
		if found := agent.FindAgent(name); found != nil {
			return found
		}
	}

	return nil
}

// FindSubAgent searches for a direct sub-agent by name
func (a *BaseAgentImpl) FindSubAgent(name string) domain.BaseAgent {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, agent := range a.subAgents {
		if agent.Name() == name {
			return agent
		}
	}
	return nil
}

// Execution methods (must be overridden by concrete implementations)

// Run executes the agent synchronously
func (a *BaseAgentImpl) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	return nil, fmt.Errorf("Run method must be implemented by concrete agent type")
}

// RunAsync executes the agent asynchronously
func (a *BaseAgentImpl) RunAsync(ctx context.Context, input *domain.State) (<-chan domain.Event, error) {
	return nil, fmt.Errorf("RunAsync method must be implemented by concrete agent type")
}

// Lifecycle methods

// Initialize initializes the agent
func (a *BaseAgentImpl) Initialize(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.initialized {
		return nil
	}

	// Initialize sub-agents
	for _, agent := range a.subAgents {
		if err := agent.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize sub-agent %s: %w", agent.Name(), err)
		}
	}

	a.initialized = true
	return nil
}

// BeforeRun is called before agent execution
func (a *BaseAgentImpl) BeforeRun(ctx context.Context, state *domain.State) error {
	// Default implementation does nothing
	return nil
}

// AfterRun is called after agent execution
func (a *BaseAgentImpl) AfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) error {
	// Default implementation does nothing
	return nil
}

// Cleanup cleans up agent resources
func (a *BaseAgentImpl) Cleanup(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Cleanup sub-agents
	for _, agent := range a.subAgents {
		if err := agent.Cleanup(ctx); err != nil {
			// Log error but continue cleanup
			_ = err
		}
	}

	a.initialized = false
	return nil
}

// Schema methods

// InputSchema returns the input schema
func (a *BaseAgentImpl) InputSchema() *sdomain.Schema {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.inputSchema
}

// OutputSchema returns the output schema
func (a *BaseAgentImpl) OutputSchema() *sdomain.Schema {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.outputSchema
}

// SetInputSchema sets the input schema
func (a *BaseAgentImpl) SetInputSchema(schema *sdomain.Schema) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.inputSchema = schema
}

// SetOutputSchema sets the output schema
func (a *BaseAgentImpl) SetOutputSchema(schema *sdomain.Schema) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.outputSchema = schema
}

// Configuration methods

// Config returns the agent configuration
func (a *BaseAgentImpl) Config() domain.AgentConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config
}

// WithConfig sets the agent configuration
func (a *BaseAgentImpl) WithConfig(config domain.AgentConfig) domain.BaseAgent {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config = config
	return a
}

// Validate validates the agent configuration
func (a *BaseAgentImpl) Validate() error {
	if a.name == "" {
		return domain.NewValidationError("name", a.name, "agent name cannot be empty")
	}

	// Validate configuration
	if err := domain.ValidateConfig(a.config); err != nil {
		return err
	}

	// Validate sub-agents
	for _, agent := range a.SubAgents() {
		if err := agent.Validate(); err != nil {
			return fmt.Errorf("sub-agent %s validation failed: %w", agent.Name(), err)
		}
	}

	return nil
}

// Metadata methods

// Metadata returns a copy of the agent metadata
func (a *BaseAgentImpl) Metadata() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range a.metadata {
		result[k] = v
	}
	return result
}

// SetMetadata sets a metadata value
func (a *BaseAgentImpl) SetMetadata(key string, value interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.metadata[key] = value
}

// Event handling

// SetEventDispatcher sets the event dispatcher
func (a *BaseAgentImpl) SetEventDispatcher(dispatcher domain.EventDispatcher) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.dispatcher = dispatcher
}

// EmitEvent emits an event if a dispatcher is configured
func (a *BaseAgentImpl) EmitEvent(eventType domain.EventType, data interface{}) {
	if a.dispatcher != nil {
		event := domain.NewEvent(eventType, a.id, a.name, data)
		a.dispatcher.Dispatch(event)
	}
}

// Helper methods

// hasCircularDependency checks if setting the given parent would create a circular dependency
func (a *BaseAgentImpl) hasCircularDependency(parent domain.BaseAgent) bool {
	current := parent
	for current != nil {
		if current.ID() == a.ID() {
			return true
		}
		current = current.Parent()
	}
	return false
}

// ExecuteWithRetry executes a function with retry logic
func (a *BaseAgentImpl) ExecuteWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= a.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(a.config.RetryDelay * time.Duration(attempt)):
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable
		if !domain.IsRetryable(lastErr) {
			return lastErr
		}
	}

	return domain.NewAgentError(a.id, a.name, "execution", domain.ErrMaxRetriesExceeded).
		WithContext("last_error", lastErr)
}

// ExecuteWithTimeout executes a function with timeout
func (a *BaseAgentImpl) ExecuteWithTimeout(ctx context.Context, fn func(context.Context) error) error {
	if a.config.Timeout <= 0 {
		return fn(ctx)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, a.config.Timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn(timeoutCtx)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return domain.NewAgentError(a.id, a.name, "execution", domain.ErrExecutionTimeout)
	}
}
