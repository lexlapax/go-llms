// ABOUTME: Enhanced mock agent implementation with comprehensive testing support
// ABOUTME: Provides response queues, sub-agent management, event tracking, and state history

package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// AgentCall represents a recorded agent execution
type AgentCall struct {
	Input     *domain.State
	Output    *domain.State
	Error     error
	Context   context.Context
	Timestamp time.Time
	Duration  time.Duration
}

// MockAgent is an enhanced mock implementation of BaseAgent
type MockAgent struct {
	// Configuration
	AgentID          string
	AgentName        string
	AgentDescription string
	AgentType        domain.AgentType

	// Response queue for deterministic testing
	ResponseQueue []*domain.State
	ErrorQueue    []error

	// Sub-agent management
	SubAgentList []domain.BaseAgent
	ParentAgent  domain.BaseAgent

	// Behavior hooks
	OnRun        func(ctx context.Context, input *domain.State) (*domain.State, error)
	OnInitialize func(ctx context.Context) error
	OnBeforeRun  func(ctx context.Context, state *domain.State) error
	OnAfterRun   func(ctx context.Context, state *domain.State, result *domain.State, err error) error
	OnCleanup    func(ctx context.Context) error

	// Event tracking
	EmittedEvents []domain.Event
	EventChannel  chan domain.Event

	// State history
	StateHistory []AgentCall

	// Configuration
	AgentConfig     domain.AgentConfig
	InputSchemaVal  *sdomain.Schema
	OutputSchemaVal *sdomain.Schema
	MetadataMap     map[string]interface{}

	// Internal state
	mu              sync.RWMutex
	initialized     bool
	executionCount  int
	queueIndex      int
	errorQueueIndex int
}

// NewMockAgent creates a new mock agent with default configuration
func NewMockAgent(name string) *MockAgent {
	return &MockAgent{
		AgentID:          fmt.Sprintf("mock-agent-%d", time.Now().UnixNano()),
		AgentName:        name,
		AgentDescription: fmt.Sprintf("Mock agent %s", name),
		AgentType:        domain.AgentTypeCustom,
		ResponseQueue:    make([]*domain.State, 0),
		ErrorQueue:       make([]error, 0),
		SubAgentList:     make([]domain.BaseAgent, 0),
		EmittedEvents:    make([]domain.Event, 0),
		StateHistory:     make([]AgentCall, 0),
		MetadataMap:      make(map[string]interface{}),
		EventChannel:     make(chan domain.Event, 100),
	}
}

// ID returns the agent ID
func (m *MockAgent) ID() string {
	return m.AgentID
}

// Name returns the agent name
func (m *MockAgent) Name() string {
	return m.AgentName
}

// Description returns the agent description
func (m *MockAgent) Description() string {
	return m.AgentDescription
}

// Type returns the agent type
func (m *MockAgent) Type() domain.AgentType {
	return m.AgentType
}

// Parent returns the parent agent
func (m *MockAgent) Parent() domain.BaseAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ParentAgent
}

// SetParent sets the parent agent
func (m *MockAgent) SetParent(parent domain.BaseAgent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ParentAgent = parent
	return nil
}

// SubAgents returns the list of sub-agents
func (m *MockAgent) SubAgents() []domain.BaseAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	agents := make([]domain.BaseAgent, len(m.SubAgentList))
	copy(agents, m.SubAgentList)
	return agents
}

// AddSubAgent adds a sub-agent
func (m *MockAgent) AddSubAgent(agent domain.BaseAgent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicates
	for _, existing := range m.SubAgentList {
		if existing.Name() == agent.Name() {
			return fmt.Errorf("sub-agent with name %s already exists", agent.Name())
		}
	}

	m.SubAgentList = append(m.SubAgentList, agent)
	if err := agent.SetParent(m); err != nil {
		return fmt.Errorf("failed to set parent: %w", err)
	}
	return nil
}

// RemoveSubAgent removes a sub-agent by name
func (m *MockAgent) RemoveSubAgent(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, agent := range m.SubAgentList {
		if agent.Name() == name {
			m.SubAgentList = append(m.SubAgentList[:i], m.SubAgentList[i+1:]...)
			if err := agent.SetParent(nil); err != nil {
				return fmt.Errorf("failed to clear parent: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("sub-agent %s not found", name)
}

// FindAgent searches for an agent in the hierarchy
func (m *MockAgent) FindAgent(name string) domain.BaseAgent {
	if m.Name() == name {
		return m
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, agent := range m.SubAgentList {
		if found := agent.FindAgent(name); found != nil {
			return found
		}
	}

	return nil
}

// FindSubAgent searches for a direct sub-agent
func (m *MockAgent) FindSubAgent(name string) domain.BaseAgent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, agent := range m.SubAgentList {
		if agent.Name() == name {
			return agent
		}
	}

	return nil
}

// Run executes the agent
func (m *MockAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	start := time.Now()

	// Check behavior hook first
	if m.OnRun != nil {
		output, err := m.OnRun(ctx, input)
		m.mu.Lock()
		m.recordCall(input, output, err, ctx, start)
		m.mu.Unlock()
		return output, err
	}

	// Use response queue if available
	m.mu.Lock()
	defer m.mu.Unlock()

	var output *domain.State
	var err error

	// Check error queue first
	if m.errorQueueIndex < len(m.ErrorQueue) {
		err = m.ErrorQueue[m.errorQueueIndex]
		m.errorQueueIndex++
	} else if m.queueIndex < len(m.ResponseQueue) {
		// Use response queue
		output = m.ResponseQueue[m.queueIndex]
		m.queueIndex++
	} else {
		// Default response
		output = domain.NewState()
		output.Set("result", fmt.Sprintf("Mock response from %s", m.Name()))
	}

	m.recordCall(input, output, err, ctx, start)
	return output, err
}

// RunAsync executes the agent asynchronously
func (m *MockAgent) RunAsync(ctx context.Context, input *domain.State) (<-chan domain.Event, error) {
	eventChan := make(chan domain.Event)

	go func() {
		defer close(eventChan)

		// Emit start event
		startEvent := domain.Event{
			Type:      "agent.start",
			Timestamp: time.Now(),
			AgentID:   m.ID(),
			Data: map[string]interface{}{
				"agent": m.Name(),
				"input": input,
			},
		}
		eventChan <- startEvent
		m.recordEvent(startEvent)

		// Run the agent
		output, err := m.Run(ctx, input)

		// Emit completion event
		completeEvent := domain.Event{
			Type:      "agent.complete",
			Timestamp: time.Now(),
			AgentID:   m.ID(),
			Data: map[string]interface{}{
				"agent":  m.Name(),
				"output": output,
				"error":  err,
			},
		}
		eventChan <- completeEvent
		m.recordEvent(completeEvent)
	}()

	return eventChan, nil
}

// Initialize initializes the agent
func (m *MockAgent) Initialize(ctx context.Context) error {
	if m.OnInitialize != nil {
		return m.OnInitialize(ctx)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.initialized = true
	return nil
}

// BeforeRun is called before running
func (m *MockAgent) BeforeRun(ctx context.Context, state *domain.State) error {
	if m.OnBeforeRun != nil {
		return m.OnBeforeRun(ctx, state)
	}
	return nil
}

// AfterRun is called after running
func (m *MockAgent) AfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) error {
	if m.OnAfterRun != nil {
		return m.OnAfterRun(ctx, state, result, err)
	}
	return nil
}

// Cleanup cleans up the agent
func (m *MockAgent) Cleanup(ctx context.Context) error {
	if m.OnCleanup != nil {
		return m.OnCleanup(ctx)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.initialized = false
	return nil
}

// InputSchema returns the input schema
func (m *MockAgent) InputSchema() *sdomain.Schema {
	return m.InputSchemaVal
}

// OutputSchema returns the output schema
func (m *MockAgent) OutputSchema() *sdomain.Schema {
	return m.OutputSchemaVal
}

// Config returns the agent configuration
func (m *MockAgent) Config() domain.AgentConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.AgentConfig
}

// WithConfig sets the agent configuration
func (m *MockAgent) WithConfig(config domain.AgentConfig) domain.BaseAgent {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.AgentConfig = config
	return m
}

// Validate validates the agent configuration
func (m *MockAgent) Validate() error {
	if m.Name() == "" {
		return fmt.Errorf("agent name cannot be empty")
	}
	return nil
}

// Metadata returns the agent metadata
func (m *MockAgent) Metadata() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	meta := make(map[string]interface{})
	for k, v := range m.MetadataMap {
		meta[k] = v
	}
	return meta
}

// SetMetadata sets a metadata value
func (m *MockAgent) SetMetadata(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MetadataMap[key] = value
}

// Helper methods for testing

// AddResponse adds a response to the queue
func (m *MockAgent) AddResponse(state *domain.State) *MockAgent {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ResponseQueue = append(m.ResponseQueue, state)
	return m
}

// AddError adds an error to the error queue
func (m *MockAgent) AddError(err error) *MockAgent {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorQueue = append(m.ErrorQueue, err)
	return m
}

// GetCallHistory returns the call history
func (m *MockAgent) GetCallHistory() []AgentCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	history := make([]AgentCall, len(m.StateHistory))
	copy(history, m.StateHistory)
	return history
}

// GetEmittedEvents returns all emitted events
func (m *MockAgent) GetEmittedEvents() []domain.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	events := make([]domain.Event, len(m.EmittedEvents))
	copy(events, m.EmittedEvents)
	return events
}

// Reset clears the mock state
func (m *MockAgent) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ResponseQueue = make([]*domain.State, 0)
	m.ErrorQueue = make([]error, 0)
	m.EmittedEvents = make([]domain.Event, 0)
	m.StateHistory = make([]AgentCall, 0)
	m.queueIndex = 0
	m.errorQueueIndex = 0
	m.executionCount = 0

	// Clear event channel
	for len(m.EventChannel) > 0 {
		<-m.EventChannel
	}
}

// Internal methods

func (m *MockAgent) recordCall(input, output *domain.State, err error, ctx context.Context, start time.Time) {
	call := AgentCall{
		Input:     input,
		Output:    output,
		Error:     err,
		Context:   ctx,
		Timestamp: start,
		Duration:  time.Since(start),
	}

	m.StateHistory = append(m.StateHistory, call)
	m.executionCount++
}

func (m *MockAgent) recordEvent(event domain.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.EmittedEvents = append(m.EmittedEvents, event)

	// Also send to channel if anyone is listening
	select {
	case m.EventChannel <- event:
	default:
		// Channel full, drop event
	}
}
