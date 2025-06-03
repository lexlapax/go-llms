# Phase 1: Core Infrastructure Implementation Plan

## Overview

This document provides detailed implementation steps for Phase 1 of the agent architecture restructuring. Phase 1 focuses on establishing the core infrastructure including base interfaces, state management, and event system.

## Directory Structure

```
pkg/agent/
├── domain/           # Core interfaces and types
│   ├── base_agent.go
│   ├── state.go
│   ├── events.go
│   ├── config.go
│   ├── artifact.go
│   ├── errors.go
│   ├── handoff.go       # NEW: Handoff interface
│   ├── guardrails.go    # NEW: Guardrails interface
│   ├── context.go       # NEW: Generic RunContext
│   ├── event_stream.go  # NEW: Event stream operations
│   └── state_validator.go # NEW: State validation
├── core/            # Base implementations
│   ├── base_agent_impl.go
│   ├── state_manager.go
│   ├── event_dispatcher.go
│   ├── agent_registry.go
│   ├── state_transforms.go # NEW: Built-in transforms
│   └── tracing.go        # NEW: OpenTelemetry support
└── utils/           # Utility functions
    ├── state_utils.go
    └── event_utils.go
```

## Detailed Implementation

### 1. Base Agent Interface (`pkg/agent/domain/base_agent.go`)

```go
package domain

import (
    "context"
    "time"
    
    sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// BaseAgent defines the core interface for all agents
type BaseAgent interface {
    // Identification
    ID() string                  // Unique identifier
    Name() string               // Human-readable name
    Description() string        // Agent description
    Type() AgentType           // Agent type (LLM, Sequential, Parallel, etc.)
    
    // Hierarchy Management
    Parent() BaseAgent
    SetParent(parent BaseAgent) error
    SubAgents() []BaseAgent
    AddSubAgent(agent BaseAgent) error
    RemoveSubAgent(name string) error
    FindAgent(name string) BaseAgent
    FindSubAgent(name string) BaseAgent
    
    // Execution
    Run(ctx context.Context, input *State) (*State, error)
    RunAsync(ctx context.Context, input *State) (<-chan Event, error)
    
    // Lifecycle Hooks
    Initialize(ctx context.Context) error
    BeforeRun(ctx context.Context, state *State) error
    AfterRun(ctx context.Context, state *State, result *State, err error) error
    Cleanup(ctx context.Context) error
    
    // Schema Definition
    InputSchema() *sdomain.Schema
    OutputSchema() *sdomain.Schema
    
    // Configuration
    Config() AgentConfig
    WithConfig(config AgentConfig) BaseAgent
    Validate() error
    
    // Metadata
    Metadata() map[string]interface{}
    SetMetadata(key string, value interface{})
}

// AgentType represents the type of agent
type AgentType string

const (
    AgentTypeLLM        AgentType = "llm"
    AgentTypeSequential AgentType = "sequential"
    AgentTypeParallel   AgentType = "parallel"
    AgentTypeConditional AgentType = "conditional"
    AgentTypeLoop       AgentType = "loop"
    AgentTypeCustom     AgentType = "custom"
)

// AgentConfig holds configuration for agents
type AgentConfig struct {
    // Common configuration
    Timeout        time.Duration          `json:"timeout,omitempty"`
    MaxRetries     int                    `json:"max_retries,omitempty"`
    RetryDelay     time.Duration          `json:"retry_delay,omitempty"`
    
    // Execution configuration
    Async          bool                   `json:"async,omitempty"`
    StreamEvents   bool                   `json:"stream_events,omitempty"`
    
    // State configuration
    ShareState     bool                   `json:"share_state,omitempty"`
    IsolateState   bool                   `json:"isolate_state,omitempty"`
    
    // Custom configuration
    Custom         map[string]interface{} `json:"custom,omitempty"`
}
```

### 2. State Management (`pkg/agent/domain/state.go`)

```go
package domain

import (
    "encoding/json"
    "sync"
    "time"
)

// State represents the execution state passed between agents
type State struct {
    mu         sync.RWMutex
    id         string
    created    time.Time
    modified   time.Time
    
    // Core state data
    values     map[string]interface{}
    artifacts  map[string]*Artifact
    messages   []Message
    
    // Metadata
    metadata   map[string]interface{}
    
    // State lineage
    parentID   string
    version    int
}

// NewState creates a new state instance
func NewState() *State {
    return &State{
        id:        generateID(),
        created:   time.Now(),
        modified:  time.Now(),
        values:    make(map[string]interface{}),
        artifacts: make(map[string]*Artifact),
        messages:  make([]Message, 0),
        metadata:  make(map[string]interface{}),
        version:   1,
    }
}

// State methods
func (s *State) Get(key string) (interface{}, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    val, ok := s.values[key]
    return val, ok
}

func (s *State) Set(key string, value interface{}) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.values[key] = value
    s.modified = time.Now()
    s.version++
}

func (s *State) Delete(key string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.values, key)
    s.modified = time.Now()
    s.version++
}

// Artifact management
func (s *State) AddArtifact(artifact *Artifact) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.artifacts[artifact.ID] = artifact
    s.modified = time.Now()
}

func (s *State) GetArtifact(id string) (*Artifact, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    artifact, ok := s.artifacts[id]
    return artifact, ok
}

// Message management
func (s *State) AddMessage(message Message) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.messages = append(s.messages, message)
    s.modified = time.Now()
}

func (s *State) Messages() []Message {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return append([]Message{}, s.messages...)
}

// State operations
func (s *State) Clone() *State {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    newState := &State{
        id:       generateID(),
        created:  time.Now(),
        modified: time.Now(),
        parentID: s.id,
        version:  1,
        values:   make(map[string]interface{}),
        artifacts: make(map[string]*Artifact),
        messages: make([]Message, len(s.messages)),
        metadata: make(map[string]interface{}),
    }
    
    // Deep copy values
    for k, v := range s.values {
        newState.values[k] = deepCopy(v)
    }
    
    // Copy artifacts (shallow copy, artifacts are immutable)
    for k, v := range s.artifacts {
        newState.artifacts[k] = v
    }
    
    // Copy messages
    copy(newState.messages, s.messages)
    
    // Copy metadata
    for k, v := range s.metadata {
        newState.metadata[k] = v
    }
    
    return newState
}

func (s *State) Merge(other *State) {
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
    
    s.modified = time.Now()
    s.version++
}

// Serialization
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

// Message represents a conversation message
type Message struct {
    Role      string                 `json:"role"`
    Content   string                 `json:"content"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
}
```

### 3. Event System (`pkg/agent/domain/events.go`)

```go
package domain

import (
    "encoding/json"
    "time"
)

// Event represents an event during agent execution
type Event struct {
    ID        string                 `json:"id"`
    Type      EventType              `json:"type"`
    AgentID   string                 `json:"agent_id"`
    AgentName string                 `json:"agent_name"`
    Timestamp time.Time              `json:"timestamp"`
    Data      interface{}            `json:"data,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Error     error                  `json:"error,omitempty"`
}

// EventType represents the type of event
type EventType string

const (
    // Lifecycle events
    EventAgentStart     EventType = "agent.start"
    EventAgentComplete  EventType = "agent.complete"
    EventAgentError     EventType = "agent.error"
    
    // Execution events
    EventStateUpdate    EventType = "state.update"
    EventProgress       EventType = "progress"
    EventMessage        EventType = "message"
    
    // Tool events
    EventToolCall       EventType = "tool.call"
    EventToolResult     EventType = "tool.result"
    EventToolError      EventType = "tool.error"
    
    // Workflow events
    EventSubAgentStart  EventType = "subagent.start"
    EventSubAgentEnd    EventType = "subagent.end"
    EventWorkflowStep   EventType = "workflow.step"
)

// NewEvent creates a new event
func NewEvent(eventType EventType, agentID, agentName string, data interface{}) Event {
    return Event{
        ID:        generateID(),
        Type:      eventType,
        AgentID:   agentID,
        AgentName: agentName,
        Timestamp: time.Now(),
        Data:      data,
        Metadata:  make(map[string]interface{}),
    }
}

// EventData types for specific events
type (
    // ProgressEventData represents progress information
    ProgressEventData struct {
        Current int    `json:"current"`
        Total   int    `json:"total"`
        Message string `json:"message"`
    }
    
    // ToolCallEventData represents tool call information
    ToolCallEventData struct {
        ToolName   string                 `json:"tool_name"`
        Parameters interface{}            `json:"parameters"`
        RequestID  string                 `json:"request_id"`
    }
    
    // ToolResultEventData represents tool result information
    ToolResultEventData struct {
        ToolName  string      `json:"tool_name"`
        Result    interface{} `json:"result"`
        RequestID string      `json:"request_id"`
        Duration  time.Duration `json:"duration"`
    }
    
    // StateUpdateEventData represents state update information
    StateUpdateEventData struct {
        Key      string      `json:"key"`
        OldValue interface{} `json:"old_value,omitempty"`
        NewValue interface{} `json:"new_value"`
        Action   string      `json:"action"` // set, delete, merge
    }
)

// EventHandler processes events
type EventHandler interface {
    HandleEvent(event Event) error
}

// EventFilter filters events
type EventFilter func(event Event) bool

// EventDispatcher manages event distribution
type EventDispatcher interface {
    Subscribe(handler EventHandler, filters ...EventFilter) string
    Unsubscribe(subscriptionID string)
    Dispatch(event Event)
    Close()
}
```

### 4. Artifact Management (`pkg/agent/domain/artifact.go`)

```go
package domain

import (
    "io"
    "time"
)

// Artifact represents a file or data artifact
type Artifact struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        ArtifactType           `json:"type"`
    MimeType    string                 `json:"mime_type"`
    Size        int64                  `json:"size"`
    Created     time.Time              `json:"created"`
    Metadata    map[string]interface{} `json:"metadata"`
    
    // Content access
    reader      io.ReadCloser
    data        []byte
}

// ArtifactType represents the type of artifact
type ArtifactType string

const (
    ArtifactTypeFile     ArtifactType = "file"
    ArtifactTypeImage    ArtifactType = "image"
    ArtifactTypeDocument ArtifactType = "document"
    ArtifactTypeData     ArtifactType = "data"
    ArtifactTypeModel    ArtifactType = "model"
    ArtifactTypeCustom   ArtifactType = "custom"
)

// NewArtifact creates a new artifact
func NewArtifact(name string, artifactType ArtifactType, data []byte) *Artifact {
    return &Artifact{
        ID:       generateID(),
        Name:     name,
        Type:     artifactType,
        Size:     int64(len(data)),
        Created:  time.Now(),
        Metadata: make(map[string]interface{}),
        data:     data,
    }
}

// Read returns a reader for the artifact content
func (a *Artifact) Read() (io.ReadCloser, error) {
    if a.reader != nil {
        return a.reader, nil
    }
    if a.data != nil {
        return io.NopCloser(bytes.NewReader(a.data)), nil
    }
    return nil, fmt.Errorf("no content available")
}

// Data returns the artifact data (if loaded in memory)
func (a *Artifact) Data() []byte {
    return a.data
}
```

### 5. Base Agent Implementation (`pkg/agent/core/base_agent_impl.go`)

```go
package core

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// baseAgentImpl provides common functionality for all agents
type baseAgentImpl struct {
    mu           sync.RWMutex
    
    // Identity
    id           string
    name         string
    description  string
    agentType    domain.AgentType
    
    // Hierarchy
    parent       domain.BaseAgent
    subAgents    []domain.BaseAgent
    
    // Configuration
    config       domain.AgentConfig
    inputSchema  *sdomain.Schema
    outputSchema *sdomain.Schema
    
    // Metadata
    metadata     map[string]interface{}
    
    // Event handling
    dispatcher   domain.EventDispatcher
}

// NewBaseAgent creates a new base agent implementation
func NewBaseAgent(name, description string, agentType domain.AgentType) *baseAgentImpl {
    return &baseAgentImpl{
        id:          domain.GenerateID(),
        name:        name,
        description: description,
        agentType:   agentType,
        subAgents:   make([]domain.BaseAgent, 0),
        metadata:    make(map[string]interface{}),
        config:      domain.AgentConfig{
            Timeout:    30 * time.Second,
            MaxRetries: 3,
            RetryDelay: time.Second,
        },
    }
}

// Identification methods
func (a *baseAgentImpl) ID() string           { return a.id }
func (a *baseAgentImpl) Name() string         { return a.name }
func (a *baseAgentImpl) Description() string  { return a.description }
func (a *baseAgentImpl) Type() domain.AgentType { return a.agentType }

// Hierarchy management
func (a *baseAgentImpl) Parent() domain.BaseAgent {
    a.mu.RLock()
    defer a.mu.RUnlock()
    return a.parent
}

func (a *baseAgentImpl) SetParent(parent domain.BaseAgent) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    // Check for circular dependencies
    if parent != nil && a.hasCircularDependency(parent) {
        return fmt.Errorf("circular dependency detected")
    }
    
    a.parent = parent
    return nil
}

func (a *baseAgentImpl) SubAgents() []domain.BaseAgent {
    a.mu.RLock()
    defer a.mu.RUnlock()
    return append([]domain.BaseAgent{}, a.subAgents...)
}

func (a *baseAgentImpl) AddSubAgent(agent domain.BaseAgent) error {
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

func (a *baseAgentImpl) RemoveSubAgent(name string) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    for i, agent := range a.subAgents {
        if agent.Name() == name {
            // Clear parent reference
            agent.SetParent(nil)
            
            // Remove from slice
            a.subAgents = append(a.subAgents[:i], a.subAgents[i+1:]...)
            return nil
        }
    }
    
    return fmt.Errorf("agent %s not found", name)
}

func (a *baseAgentImpl) FindAgent(name string) domain.BaseAgent {
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

func (a *baseAgentImpl) FindSubAgent(name string) domain.BaseAgent {
    a.mu.RLock()
    defer a.mu.RUnlock()
    
    for _, agent := range a.subAgents {
        if agent.Name() == name {
            return agent
        }
    }
    return nil
}

// Configuration
func (a *baseAgentImpl) Config() domain.AgentConfig {
    a.mu.RLock()
    defer a.mu.RUnlock()
    return a.config
}

func (a *baseAgentImpl) WithConfig(config domain.AgentConfig) domain.BaseAgent {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.config = config
    return a
}

// Schema methods
func (a *baseAgentImpl) InputSchema() *sdomain.Schema {
    a.mu.RLock()
    defer a.mu.RUnlock()
    return a.inputSchema
}

func (a *baseAgentImpl) OutputSchema() *sdomain.Schema {
    a.mu.RLock()
    defer a.mu.RUnlock()
    return a.outputSchema
}

// Metadata methods
func (a *baseAgentImpl) Metadata() map[string]interface{} {
    a.mu.RLock()
    defer a.mu.RUnlock()
    
    result := make(map[string]interface{})
    for k, v := range a.metadata {
        result[k] = v
    }
    return result
}

func (a *baseAgentImpl) SetMetadata(key string, value interface{}) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.metadata[key] = value
}

// Lifecycle methods (default implementations)
func (a *baseAgentImpl) Initialize(ctx context.Context) error {
    return nil
}

func (a *baseAgentImpl) BeforeRun(ctx context.Context, state *domain.State) error {
    return nil
}

func (a *baseAgentImpl) AfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) error {
    return nil
}

func (a *baseAgentImpl) Cleanup(ctx context.Context) error {
    return nil
}

func (a *baseAgentImpl) Validate() error {
    if a.name == "" {
        return fmt.Errorf("agent name cannot be empty")
    }
    return nil
}

// Helper methods
func (a *baseAgentImpl) hasCircularDependency(parent domain.BaseAgent) bool {
    current := parent
    for current != nil {
        if current.ID() == a.ID() {
            return true
        }
        current = current.Parent()
    }
    return false
}

// Event emission helpers
func (a *baseAgentImpl) emitEvent(eventType domain.EventType, data interface{}) {
    if a.dispatcher != nil {
        event := domain.NewEvent(eventType, a.id, a.name, data)
        a.dispatcher.Dispatch(event)
    }
}
```

### 6. State Manager (`pkg/agent/core/state_manager.go`)

```go
package core

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// StateManager manages state lifecycle and transformations
type StateManager struct {
    mu          sync.RWMutex
    states      map[string]*domain.State
    transforms  map[string]StateTransform
}

// StateTransform defines a state transformation function
type StateTransform func(ctx context.Context, input *domain.State) (*domain.State, error)

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
    return &StateManager{
        states:     make(map[string]*domain.State),
        transforms: make(map[string]StateTransform),
    }
}

// SaveState stores a state snapshot
func (sm *StateManager) SaveState(state *domain.State) error {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    if state == nil {
        return fmt.Errorf("state cannot be nil")
    }
    
    sm.states[state.ID()] = state.Clone()
    return nil
}

// LoadState retrieves a state snapshot
func (sm *StateManager) LoadState(id string) (*domain.State, error) {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    
    state, ok := sm.states[id]
    if !ok {
        return nil, fmt.Errorf("state %s not found", id)
    }
    
    return state.Clone(), nil
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

// MergeStates merges multiple states according to a strategy
func (sm *StateManager) MergeStates(states []*domain.State, strategy MergeStrategy) (*domain.State, error) {
    if len(states) == 0 {
        return nil, fmt.Errorf("no states to merge")
    }
    
    return strategy(states)
}

// MergeStrategy defines how to merge multiple states
type MergeStrategy func(states []*domain.State) (*domain.State, error)

// Built-in merge strategies
var (
    // MergeStrategyLast takes the last state
    MergeStrategyLast MergeStrategy = func(states []*domain.State) (*domain.State, error) {
        return states[len(states)-1].Clone(), nil
    }
    
    // MergeStrategyMergeAll merges all states in order
    MergeStrategyMergeAll MergeStrategy = func(states []*domain.State) (*domain.State, error) {
        result := domain.NewState()
        for _, state := range states {
            result.Merge(state)
        }
        return result, nil
    }
    
    // MergeStrategyUnion creates a union of all values
    MergeStrategyUnion MergeStrategy = func(states []*domain.State) (*domain.State, error) {
        result := domain.NewState()
        
        // Collect all unique keys and their values
        valueMap := make(map[string][]interface{})
        for _, state := range states {
            for k, v := range state.Values() {
                valueMap[k] = append(valueMap[k], v)
            }
        }
        
        // Store arrays of values for keys that appear in multiple states
        for k, values := range valueMap {
            if len(values) == 1 {
                result.Set(k, values[0])
            } else {
                result.Set(k, values)
            }
        }
        
        return result, nil
    }
)
```

### 7. Event Dispatcher (`pkg/agent/core/event_dispatcher.go`)

```go
package core

import (
    "context"
    "sync"
    
    "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// eventDispatcher implements EventDispatcher
type eventDispatcher struct {
    mu            sync.RWMutex
    subscriptions map[string]*subscription
    eventChan     chan domain.Event
    ctx           context.Context
    cancel        context.CancelFunc
    wg            sync.WaitGroup
}

type subscription struct {
    id       string
    handler  domain.EventHandler
    filters  []domain.EventFilter
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher(bufferSize int) domain.EventDispatcher {
    ctx, cancel := context.WithCancel(context.Background())
    ed := &eventDispatcher{
        subscriptions: make(map[string]*subscription),
        eventChan:     make(chan domain.Event, bufferSize),
        ctx:           ctx,
        cancel:        cancel,
    }
    
    ed.wg.Add(1)
    go ed.processEvents()
    
    return ed
}

func (ed *eventDispatcher) Subscribe(handler domain.EventHandler, filters ...domain.EventFilter) string {
    ed.mu.Lock()
    defer ed.mu.Unlock()
    
    sub := &subscription{
        id:      domain.GenerateID(),
        handler: handler,
        filters: filters,
    }
    
    ed.subscriptions[sub.id] = sub
    return sub.id
}

func (ed *eventDispatcher) Unsubscribe(subscriptionID string) {
    ed.mu.Lock()
    defer ed.mu.Unlock()
    delete(ed.subscriptions, subscriptionID)
}

func (ed *eventDispatcher) Dispatch(event domain.Event) {
    select {
    case ed.eventChan <- event:
    case <-ed.ctx.Done():
    }
}

func (ed *eventDispatcher) Close() {
    ed.cancel()
    close(ed.eventChan)
    ed.wg.Wait()
}

func (ed *eventDispatcher) processEvents() {
    defer ed.wg.Done()
    
    for {
        select {
        case event, ok := <-ed.eventChan:
            if !ok {
                return
            }
            ed.handleEvent(event)
        case <-ed.ctx.Done():
            return
        }
    }
}

func (ed *eventDispatcher) handleEvent(event domain.Event) {
    ed.mu.RLock()
    defer ed.mu.RUnlock()
    
    for _, sub := range ed.subscriptions {
        // Check filters
        if !ed.matchesFilters(event, sub.filters) {
            continue
        }
        
        // Handle event (non-blocking)
        go func(h domain.EventHandler, e domain.Event) {
            if err := h.HandleEvent(e); err != nil {
                // Log error or emit error event
            }
        }(sub.handler, event)
    }
}

func (ed *eventDispatcher) matchesFilters(event domain.Event, filters []domain.EventFilter) bool {
    for _, filter := range filters {
        if !filter(event) {
            return false
        }
    }
    return true
}
```

### 8. Handoff Interface (`pkg/agent/domain/handoff.go`)

```go
package domain

import (
    "context"
)

// Handoff represents a delegation mechanism between agents
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

func NewHandoffBuilder(name, targetAgent string) *HandoffBuilder {
    return &HandoffBuilder{
        name:        name,
        targetAgent: targetAgent,
    }
}

func (hb *HandoffBuilder) WithDescription(desc string) *HandoffBuilder {
    hb.description = desc
    return hb
}

func (hb *HandoffBuilder) WithInputFilter(filter func(*State) *State) *HandoffBuilder {
    hb.inputFilter = filter
    return hb
}

func (hb *HandoffBuilder) WithMessageFilter(filter func([]Message) []Message) *HandoffBuilder {
    hb.messageFilter = filter
    return hb
}

func (hb *HandoffBuilder) Build() Handoff {
    return &handoffImpl{
        name:          hb.name,
        targetAgent:   hb.targetAgent,
        description:   hb.description,
        inputFilter:   hb.inputFilter,
        messageFilter: hb.messageFilter,
    }
}
```

### 9. Guardrails Interface (`pkg/agent/domain/guardrails.go`)

```go
package domain

import (
    "context"
    "time"
)

// GuardrailType represents when the guardrail is applied
type GuardrailType string

const (
    GuardrailTypeInput  GuardrailType = "input"
    GuardrailTypeOutput GuardrailType = "output"
    GuardrailTypeBoth   GuardrailType = "both"
)

// Guardrail validates agent inputs/outputs
type Guardrail interface {
    Name() string
    Type() GuardrailType
    
    // Validation
    Validate(ctx context.Context, state *State) error
    
    // Async validation with timeout
    ValidateAsync(ctx context.Context, state *State, timeout time.Duration) <-chan error
}

// GuardrailChain runs multiple guardrails
type GuardrailChain struct {
    guardrails []Guardrail
    failFast   bool
}

func NewGuardrailChain(failFast bool) *GuardrailChain {
    return &GuardrailChain{
        guardrails: make([]Guardrail, 0),
        failFast:   failFast,
    }
}

func (gc *GuardrailChain) Add(guardrail Guardrail) *GuardrailChain {
    gc.guardrails = append(gc.guardrails, guardrail)
    return gc
}

func (gc *GuardrailChain) Validate(ctx context.Context, state *State) error {
    for _, g := range gc.guardrails {
        if err := g.Validate(ctx, state); err != nil {
            if gc.failFast {
                return err
            }
        }
    }
    return nil
}
```

### 10. Enhanced RunContext (`pkg/agent/domain/context.go`)

```go
package domain

import (
    "context"
    "time"
)

// RunContext provides type-safe dependency injection
type RunContext[D any] struct {
    context.Context
    
    // Dependencies
    Deps D
    
    // Execution metadata
    RunID      string
    Retry      int
    StartTime  time.Time
    
    // State access
    State      *State
    
    // Event emission
    EmitEvent  func(Event)
}

// NewRunContext creates a new RunContext
func NewRunContext[D any](ctx context.Context, deps D, state *State) *RunContext[D] {
    return &RunContext[D]{
        Context:   ctx,
        Deps:      deps,
        RunID:     generateID(),
        StartTime: time.Now(),
        State:     state,
    }
}

// WithRetry creates a new context for retry attempt
func (rc *RunContext[D]) WithRetry(retry int) *RunContext[D] {
    newCtx := *rc
    newCtx.Retry = retry
    return &newCtx
}

// Example usage types
type DatabaseDeps struct {
    DB     interface{} // *sql.DB
    Cache  interface{} // *redis.Client
    Logger interface{} // *slog.Logger
}

type ServiceDeps struct {
    UserService    interface{}
    ProductService interface{}
    OrderService   interface{}
}
```

### 11. Event Stream Interface (`pkg/agent/domain/event_stream.go`)

```go
package domain

import (
    "time"
)

// EventStream provides functional operations on event streams
type EventStream interface {
    // Core operations
    Filter(predicate EventPredicate) EventStream
    Map(transform EventTransform) EventStream
    Reduce(reducer EventReducer, initial interface{}) interface{}
    
    // Stream control
    Take(n int) EventStream
    TakeUntil(predicate EventPredicate) EventStream
    Timeout(duration time.Duration) EventStream
    
    // Consumption
    ForEach(handler EventHandler) error
    Collect() ([]Event, error)
    First() (Event, error)
}

// EventPredicate filters events
type EventPredicate func(Event) bool

// EventTransform transforms events
type EventTransform func(Event) Event

// EventReducer reduces events to a single value
type EventReducer func(interface{}, Event) interface{}

// Common predicates
var (
    IsError EventPredicate = func(e Event) bool {
        return e.Type == EventAgentError || e.Type == EventToolError
    }
    
    IsComplete EventPredicate = func(e Event) bool {
        return e.Type == EventAgentComplete
    }
    
    ByType = func(eventType EventType) EventPredicate {
        return func(e Event) bool {
            return e.Type == eventType
        }
    }
    
    ByAgent = func(agentName string) EventPredicate {
        return func(e Event) bool {
            return e.AgentName == agentName
        }
    }
)
```

### 12. State Validators (`pkg/agent/domain/state_validator.go`)

```go
package domain

import (
    "fmt"
    
    sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// StateValidator validates state
type StateValidator interface {
    Validate(state *State) error
}

// StateValidatorFunc is a function adapter
type StateValidatorFunc func(state *State) error

func (f StateValidatorFunc) Validate(state *State) error {
    return f(state)
}

// Built-in validators
var (
    // RequiredKeysValidator ensures required keys exist
    RequiredKeysValidator = func(keys ...string) StateValidator {
        return StateValidatorFunc(func(state *State) error {
            for _, key := range keys {
                if _, ok := state.Get(key); !ok {
                    return fmt.Errorf("required key missing: %s", key)
                }
            }
            return nil
        })
    }
    
    // SchemaValidator validates against JSON schema
    SchemaValidator = func(schema *sdomain.Schema) StateValidator {
        return StateValidatorFunc(func(state *State) error {
            return schema.Validate(state.Values())
        })
    }
    
    // TypeValidator ensures values are of correct type
    TypeValidator = func(key string, expectedType string) StateValidator {
        return StateValidatorFunc(func(state *State) error {
            val, ok := state.Get(key)
            if !ok {
                return nil // Key doesn't exist, not a type error
            }
            
            // Type checking logic here
            actualType := fmt.Sprintf("%T", val)
            if actualType != expectedType {
                return fmt.Errorf("key %s: expected type %s, got %s", key, expectedType, actualType)
            }
            return nil
        })
    }
    
    // CompositeValidator combines multiple validators
    CompositeValidator = func(validators ...StateValidator) StateValidator {
        return StateValidatorFunc(func(state *State) error {
            for _, v := range validators {
                if err := v.Validate(state); err != nil {
                    return err
                }
            }
            return nil
        })
    }
)
```

### 13. State Transforms (`pkg/agent/core/state_transforms.go`)

```go
package core

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Built-in state transformation functions
var (
    // FilterTransform removes keys matching pattern
    FilterTransform = func(pattern string) StateTransform {
        return func(ctx context.Context, state *domain.State) (*domain.State, error) {
            result := state.Clone()
            for key := range state.Values() {
                if matched, _ := filepath.Match(pattern, key); matched {
                    result.Delete(key)
                }
            }
            return result, nil
        }
    }
    
    // MapTransform applies function to all values
    MapTransform = func(fn func(interface{}) interface{}) StateTransform {
        return func(ctx context.Context, state *domain.State) (*domain.State, error) {
            result := state.Clone()
            for key, value := range state.Values() {
                result.Set(key, fn(value))
            }
            return result, nil
        }
    }
    
    // ValidateTransform ensures state matches schema
    ValidateTransform = func(schema *sdomain.Schema) StateTransform {
        return func(ctx context.Context, state *domain.State) (*domain.State, error) {
            if err := schema.Validate(state.Values()); err != nil {
                return nil, fmt.Errorf("state validation failed: %w", err)
            }
            return state, nil
        }
    }
    
    // PrefixKeysTransform adds prefix to all keys
    PrefixKeysTransform = func(prefix string) StateTransform {
        return func(ctx context.Context, state *domain.State) (*domain.State, error) {
            result := domain.NewState()
            for key, value := range state.Values() {
                result.Set(prefix+key, value)
            }
            return result, nil
        }
    }
    
    // SelectKeysTransform keeps only specified keys
    SelectKeysTransform = func(keys ...string) StateTransform {
        return func(ctx context.Context, state *domain.State) (*domain.State, error) {
            result := domain.NewState()
            for _, key := range keys {
                if value, ok := state.Get(key); ok {
                    result.Set(key, value)
                }
            }
            return result, nil
        }
    }
)
```

### 14. Tracing Support (`pkg/agent/core/tracing.go`)

```go
package core

import (
    "context"
    "fmt"
    
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

// TracingHook provides OpenTelemetry integration
type TracingHook struct {
    tracer trace.Tracer
}

// NewTracingHook creates a new tracing hook
func NewTracingHook(tracerName string) *TracingHook {
    return &TracingHook{
        tracer: otel.Tracer(tracerName),
    }
}

// BeforeRun starts a new span
func (h *TracingHook) BeforeRun(ctx context.Context, agent domain.BaseAgent, state *domain.State) (context.Context, error) {
    ctx, span := h.tracer.Start(ctx, fmt.Sprintf("agent.%s.run", agent.Name()),
        trace.WithAttributes(
            attribute.String("agent.id", agent.ID()),
            attribute.String("agent.type", string(agent.Type())),
            attribute.String("agent.name", agent.Name()),
            attribute.String("state.id", state.ID()),
        ),
    )
    
    // Add state size as attribute
    span.SetAttributes(
        attribute.Int("state.values.count", len(state.Values())),
        attribute.Int("state.messages.count", len(state.Messages())),
    )
    
    return ctx, nil
}

// AfterRun completes the span
func (h *TracingHook) AfterRun(ctx context.Context, agent domain.BaseAgent, state *domain.State, result *domain.State, err error) error {
    span := trace.SpanFromContext(ctx)
    if span == nil {
        return nil
    }
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    } else {
        span.SetStatus(codes.Ok, "")
        if result != nil {
            span.SetAttributes(
                attribute.String("result.id", result.ID()),
                attribute.Int("result.values.count", len(result.Values())),
            )
        }
    }
    
    span.End()
    return nil
}

// ToolCallHook traces tool calls
type ToolCallHook struct {
    tracer trace.Tracer
}

func NewToolCallHook(tracerName string) *ToolCallHook {
    return &ToolCallHook{
        tracer: otel.Tracer(tracerName),
    }
}

func (h *ToolCallHook) BeforeToolCall(ctx context.Context, toolName string, params interface{}) (context.Context, error) {
    ctx, span := h.tracer.Start(ctx, fmt.Sprintf("tool.%s.call", toolName),
        trace.WithAttributes(
            attribute.String("tool.name", toolName),
        ),
    )
    return ctx, nil
}

func (h *ToolCallHook) AfterToolCall(ctx context.Context, toolName string, result interface{}, err error) error {
    span := trace.SpanFromContext(ctx)
    if span == nil {
        return nil
    }
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    } else {
        span.SetStatus(codes.Ok, "")
    }
    
    span.End()
    return nil
}
```

## Testing Strategy

### Unit Tests

1. **State Management Tests** (`state_test.go`)
   - Test state creation, cloning, merging
   - Test concurrent access safety
   - Test serialization/deserialization

2. **Event System Tests** (`events_test.go`)
   - Test event creation and dispatch
   - Test filtering and subscription
   - Test concurrent event handling

3. **Base Agent Tests** (`base_agent_test.go`)
   - Test hierarchy management
   - Test configuration
   - Test lifecycle methods

4. **Handoff Tests** (`handoff_test.go`)
   - Test handoff execution
   - Test input transformation
   - Test message filtering

5. **Guardrail Tests** (`guardrails_test.go`)
   - Test validation logic
   - Test async validation
   - Test guardrail chains

6. **RunContext Tests** (`context_test.go`)
   - Test generic type safety
   - Test dependency injection
   - Test context propagation

7. **Event Stream Tests** (`event_stream_test.go`)
   - Test filtering and mapping
   - Test stream operations
   - Test timeout behavior

8. **State Validator Tests** (`state_validator_test.go`)
   - Test built-in validators
   - Test composite validation
   - Test error handling

### Integration Tests

1. **State Flow Tests**
   - Test state passing between agents
   - Test state isolation modes
   - Test state persistence

2. **Event Flow Tests**
   - Test event propagation through agent hierarchy
   - Test event ordering
   - Test error event handling

3. **Tracing Tests**
   - Test span creation and completion
   - Test error recording
   - Test attribute propagation

## Performance Considerations

1. **Object Pooling**
   - Pool State objects for frequent creation/destruction
   - Pool Event objects
   - Use sync.Pool for temporary objects

2. **Concurrent Execution**
   - Use channels for event streaming
   - Implement proper context cancellation
   - Avoid blocking operations in event handlers

3. **Memory Management**
   - Implement state size limits
   - Clean up artifacts after use
   - Use weak references where appropriate

## Migration Path

### From Current Implementation

```go
// Current
agent := workflow.NewAgent(provider)
result, err := agent.Run(ctx, "input")

// New Phase 1 (prepare for full migration)
// Create adapter that implements BaseAgent
adapter := NewLegacyAgentAdapter(agent)
state := domain.NewState()
state.Set("input", "input")
resultState, err := adapter.Run(ctx, state)
```

## Next Steps

After Phase 1 completion:
1. Implement LLMAgent using the base infrastructure
2. Create workflow agents (Sequential, Parallel)
3. Implement agent-tool conversion
4. Create comprehensive examples
5. Performance benchmarking

## Deliverables

1. Complete interface definitions in `pkg/agent/domain/`
2. Base implementations in `pkg/agent/core/`
3. Comprehensive unit tests
4. Integration test suite
5. Performance benchmarks
6. Migration guide documentation