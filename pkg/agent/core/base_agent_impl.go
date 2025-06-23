// ABOUTME: Base implementation for all agent types providing common functionality
// ABOUTME: including hierarchy management, event handling, and configuration

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

// BaseAgentImpl provides common functionality for all agents.
// It implements the domain.BaseAgent interface and handles identity management,
// hierarchical relationships, event dispatching, and state management.
// This type should be embedded in concrete agent implementations.
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

	// State inheritance configuration
	enableSharedState bool
	inheritMessages   bool
	inheritArtifacts  bool
	inheritMetadata   bool

	// Metadata
	metadata map[string]interface{}

	// Event handling
	dispatcher domain.EventDispatcher

	// State
	initialized bool
}

// NewBaseAgent creates a new base agent implementation with the given properties.
// It generates a unique ID, initializes default configuration with 30s timeout,
// 3 retries, and enables state inheritance by default. The returned agent
// can be embedded in concrete agent types or used directly for simple agents.
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
		// Default state inheritance configuration
		enableSharedState: true,
		inheritMessages:   true,
		inheritArtifacts:  true,
		inheritMetadata:   true,
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

// Hierarchy methods

// Parent returns the parent agent
func (a *BaseAgentImpl) Parent() domain.BaseAgent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.parent
}

// SetParent sets the parent agent
func (a *BaseAgentImpl) SetParent(parent domain.BaseAgent) error {
	if parent != nil && a.hasCircularDependency(parent) {
		return fmt.Errorf("circular dependency detected")
	}

	a.mu.Lock()
	defer a.mu.Unlock()
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

	// Check if agent already exists (with lock)
	a.mu.RLock()
	for _, existing := range a.subAgents {
		if existing.ID() == agent.ID() {
			a.mu.RUnlock()
			return fmt.Errorf("agent with ID %s already exists", agent.ID())
		}
	}
	a.mu.RUnlock()

	// Set parent (without lock to avoid deadlock)
	if err := agent.SetParent(a); err != nil {
		return fmt.Errorf("failed to set parent: %w", err)
	}

	// Add to subAgents (with lock)
	a.mu.Lock()
	a.subAgents = append(a.subAgents, agent)
	a.mu.Unlock()

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
	eventChan := make(chan domain.Event, 100)

	go func() {
		defer close(eventChan)

		// Emit start event
		eventChan <- domain.NewEvent(domain.EventAgentStart, a.id, a.name, nil)

		result, err := a.Run(ctx, input)
		if err != nil {
			// Emit error event
			eventChan <- domain.NewEvent(domain.EventAgentError, a.id, a.name, err)
			return
		}

		// Emit completion event with result
		eventChan <- domain.NewEvent(domain.EventAgentComplete, a.id, a.name, result)
	}()

	return eventChan, nil
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

// Subscribe adds an event handler with optional filters
func (a *BaseAgentImpl) Subscribe(handler domain.EventHandler, filters ...domain.EventFilter) string {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Create dispatcher if it doesn't exist
	if a.dispatcher == nil {
		a.dispatcher = NewEventDispatcher(100) // Default buffer size
	}

	return a.dispatcher.Subscribe(handler, filters...)
}

// Unsubscribe removes an event handler
func (a *BaseAgentImpl) Unsubscribe(subscriptionID string) {
	a.mu.RLock()
	dispatcher := a.dispatcher
	a.mu.RUnlock()

	if dispatcher != nil {
		dispatcher.Unsubscribe(subscriptionID)
	}
}

// OnEvent is a convenience method to subscribe to events with a function handler
func (a *BaseAgentImpl) OnEvent(handler func(event *domain.Event)) string {
	// Convert function to EventHandler interface
	eventHandler := domain.EventHandlerFunc(func(event domain.Event) error {
		handler(&event)
		return nil
	})

	return a.Subscribe(eventHandler)
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

// State inheritance configuration methods

// EnableSharedState enables or disables shared state for sub-agents
func (a *BaseAgentImpl) EnableSharedState(enable bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enableSharedState = enable
}

// ConfigureStateInheritance configures what sub-agents inherit from parent state
func (a *BaseAgentImpl) ConfigureStateInheritance(messages, artifacts, metadata bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.inheritMessages = messages
	a.inheritArtifacts = artifacts
	a.inheritMetadata = metadata
}

// IsSharedStateEnabled returns whether shared state is enabled
func (a *BaseAgentImpl) IsSharedStateEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.enableSharedState
}

// GetStateInheritanceConfig returns the current state inheritance configuration
func (a *BaseAgentImpl) GetStateInheritanceConfig() (messages, artifacts, metadata bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.inheritMessages, a.inheritArtifacts, a.inheritMetadata
}
