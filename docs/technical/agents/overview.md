# Agent Architecture: Overview and Concepts

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [Agents](../../technical/agents) / Overview**

Comprehensive overview of the agent architecture in Go-LLMs, covering agent types, design patterns, execution models, communication protocols, and integration strategies for building intelligent, autonomous systems.

## Agent Architecture Overview

### Core Agent Interface

```go
// Agent defines the fundamental agent interface
type Agent interface {
    // Identity and metadata
    ID() string
    Name() string
    Description() string
    Type() AgentType
    
    // Execution
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    ExecuteStream(ctx context.Context, input interface{}) (<-chan AgentResponse, error)
    
    // State management
    GetState() AgentState
    SetState(state AgentState) error
    
    // Lifecycle
    Initialize(ctx context.Context, config AgentConfig) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() HealthStatus
}

// AgentType defines different categories of agents
type AgentType string

const (
    AgentTypeLLM        AgentType = "llm"        // LLM-powered agents
    AgentTypeWorkflow   AgentType = "workflow"   // Workflow orchestration agents
    AgentTypeComposite  AgentType = "composite"  // Multi-agent compositions
    AgentTypeScript     AgentType = "script"     // Script-based agents
    AgentTypeService    AgentType = "service"    // Service integration agents
    AgentTypeProxy      AgentType = "proxy"      // Proxy/adapter agents
)

// AgentResponse represents agent execution results
type AgentResponse struct {
    ID        string                 `json:"id"`
    AgentID   string                 `json:"agent_id"`
    Content   interface{}            `json:"content"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Error     error                  `json:"error,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
    Final     bool                   `json:"final"`
}

// AgentState represents agent state information
type AgentState interface {
    Get(key string) interface{}
    Set(key string, value interface{})
    Delete(key string)
    Keys() []string
    Snapshot() map[string]interface{}
    Load(snapshot map[string]interface{}) error
}
```

### Agent Hierarchy

```go
// BaseAgent provides common agent functionality
type BaseAgent struct {
    id          string
    name        string
    description string
    agentType   AgentType
    state       AgentState
    config      AgentConfig
    logger      *zap.Logger
    metrics     *AgentMetrics
    
    // Lifecycle
    status      AgentStatus
    startTime   time.Time
    lastUsed    time.Time
    
    // Communication
    eventBus    EventBus
    messageQueue MessageQueue
    
    // Error handling
    errorHandler ErrorHandler
    retryPolicy  RetryPolicy
}

// AgentStatus represents agent operational status
type AgentStatus string

const (
    AgentStatusIdle       AgentStatus = "idle"
    AgentStatusRunning    AgentStatus = "running"
    AgentStatusPaused     AgentStatus = "paused"
    AgentStatusStopped    AgentStatus = "stopped"
    AgentStatusError      AgentStatus = "error"
    AgentStatusShutdown   AgentStatus = "shutdown"
)

// Specialized agent interfaces
type LLMAgent interface {
    Agent
    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
    CompleteStream(ctx context.Context, request *CompletionRequest) (<-chan StreamChunk, error)
    GetProvider() provider.Provider
    GetTools() []domain.Tool
    SetSystemPrompt(prompt string)
    AddTool(tool domain.Tool) error
    RemoveTool(name string) error
}

type WorkflowAgent interface {
    Agent
    AddStep(step WorkflowStep) error
    RemoveStep(name string) error
    GetSteps() []WorkflowStep
    ExecuteStep(ctx context.Context, stepName string, input interface{}) (interface{}, error)
    GetWorkflowState() WorkflowState
}

type CompositeAgent interface {
    Agent
    AddSubAgent(agent Agent) error
    RemoveSubAgent(id string) error
    GetSubAgents() []Agent
    Broadcast(ctx context.Context, message interface{}) error
    GetCoordinator() AgentCoordinator
}
```

---

## Agent Design Patterns

### 1. Simple Agent Pattern

```go
// SimpleAgent implements basic agent functionality
type SimpleAgent struct {
    *BaseAgent
    executor Executor
}

func NewSimpleAgent(name string, executor Executor) *SimpleAgent {
    return &SimpleAgent{
        BaseAgent: NewBaseAgent(name, AgentTypeLLM),
        executor:  executor,
    }
}

func (a *SimpleAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Pre-execution hooks
    if err := a.preExecute(ctx, input); err != nil {
        return nil, err
    }
    
    // Execute with retry logic
    result, err := a.executeWithRetry(ctx, input)
    if err != nil {
        a.handleError(err)
        return nil, err
    }
    
    // Post-execution hooks
    if err := a.postExecute(ctx, result); err != nil {
        return nil, err
    }
    
    return result, nil
}

func (a *SimpleAgent) executeWithRetry(ctx context.Context, input interface{}) (interface{}, error) {
    var lastErr error
    
    for attempt := 0; attempt < a.retryPolicy.MaxAttempts; attempt++ {
        if attempt > 0 {
            backoff := a.retryPolicy.BackoffStrategy.NextBackoff(attempt)
            select {
            case <-time.After(backoff):
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }
        
        result, err := a.executor.Execute(ctx, input)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !a.retryPolicy.IsRetryable(err) {
            break
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

### 2. Stateful Agent Pattern

```go
// StatefulAgent maintains state across executions
type StatefulAgent struct {
    *BaseAgent
    stateManager StateManager
    persistence  StatePersistence
}

func NewStatefulAgent(name string, stateManager StateManager) *StatefulAgent {
    return &StatefulAgent{
        BaseAgent:    NewBaseAgent(name, AgentTypeLLM),
        stateManager: stateManager,
        persistence:  NewMemoryStatePersistence(),
    }
}

func (a *StatefulAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Load state
    if err := a.loadState(ctx); err != nil {
        return nil, fmt.Errorf("failed to load state: %w", err)
    }
    
    // Execute with state context
    result, err := a.executeStateful(ctx, input)
    if err != nil {
        return nil, err
    }
    
    // Save state
    if err := a.saveState(ctx); err != nil {
        a.logger.Warn("Failed to save state", zap.Error(err))
    }
    
    return result, nil
}

func (a *StatefulAgent) executeStateful(ctx context.Context, input interface{}) (interface{}, error) {
    // Update state based on input
    a.state.Set("last_input", input)
    a.state.Set("execution_count", a.getExecutionCount()+1)
    a.state.Set("last_execution", time.Now())
    
    // Execute with state awareness
    return a.stateManager.ExecuteWithState(ctx, input, a.state)
}

func (a *StatefulAgent) loadState(ctx context.Context) error {
    stateData, err := a.persistence.Load(ctx, a.id)
    if err != nil {
        if errors.Is(err, ErrStateNotFound) {
            // Initialize new state
            a.state = NewAgentState()
            return nil
        }
        return err
    }
    
    return a.state.Load(stateData)
}

func (a *StatefulAgent) saveState(ctx context.Context) error {
    stateData := a.state.Snapshot()
    return a.persistence.Save(ctx, a.id, stateData)
}
```

### 3. Event-Driven Agent Pattern

```go
// EventDrivenAgent responds to events and messages
type EventDrivenAgent struct {
    *BaseAgent
    eventHandlers map[string]EventHandler
    messageQueue  MessageQueue
    subscribers   []EventSubscriber
}

type EventHandler interface {
    Handle(ctx context.Context, event Event) error
    CanHandle(event Event) bool
}

type Event struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Source    string                 `json:"source"`
    Data      interface{}            `json:"data"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
}

func NewEventDrivenAgent(name string) *EventDrivenAgent {
    return &EventDrivenAgent{
        BaseAgent:     NewBaseAgent(name, AgentTypeLLM),
        eventHandlers: make(map[string]EventHandler),
        messageQueue:  NewInMemoryMessageQueue(),
        subscribers:   make([]EventSubscriber, 0),
    }
}

func (a *EventDrivenAgent) Start(ctx context.Context) error {
    if err := a.BaseAgent.Start(ctx); err != nil {
        return err
    }
    
    // Start event processing
    go a.processEvents(ctx)
    go a.processMessages(ctx)
    
    return nil
}

func (a *EventDrivenAgent) processEvents(ctx context.Context) {
    for {
        select {
        case event := <-a.eventBus.Subscribe(a.id):
            if err := a.handleEvent(ctx, event); err != nil {
                a.logger.Error("Failed to handle event",
                    zap.String("event_id", event.ID),
                    zap.String("event_type", event.Type),
                    zap.Error(err),
                )
            }
            
        case <-ctx.Done():
            return
        }
    }
}

func (a *EventDrivenAgent) handleEvent(ctx context.Context, event Event) error {
    handler, ok := a.eventHandlers[event.Type]
    if !ok {
        // Check for wildcard handlers
        if wildcardHandler, exists := a.eventHandlers["*"]; exists {
            handler = wildcardHandler
        } else {
            return fmt.Errorf("no handler for event type: %s", event.Type)
        }
    }
    
    if !handler.CanHandle(event) {
        return nil // Skip event
    }
    
    return handler.Handle(ctx, event)
}

func (a *EventDrivenAgent) RegisterEventHandler(eventType string, handler EventHandler) {
    a.eventHandlers[eventType] = handler
}

func (a *EventDrivenAgent) PublishEvent(event Event) error {
    return a.eventBus.Publish(event)
}
```

---

## Agent Communication

### Message Passing

```go
// MessageBus facilitates agent communication
type MessageBus interface {
    Send(ctx context.Context, message Message) error
    SendToAgent(ctx context.Context, agentID string, message Message) error
    Broadcast(ctx context.Context, message Message) error
    Subscribe(agentID string) <-chan Message
    Unsubscribe(agentID string) error
}

type Message struct {
    ID          string                 `json:"id"`
    From        string                 `json:"from"`
    To          string                 `json:"to"`
    Type        MessageType            `json:"type"`
    Content     interface{}            `json:"content"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    ReplyTo     string                 `json:"reply_to,omitempty"`
    Timestamp   time.Time              `json:"timestamp"`
    TTL         time.Duration          `json:"ttl,omitempty"`
}

type MessageType string

const (
    MessageTypeRequest    MessageType = "request"
    MessageTypeResponse   MessageType = "response"
    MessageTypeEvent      MessageType = "event"
    MessageTypeCommand    MessageType = "command"
    MessageTypeQuery      MessageType = "query"
    MessageTypeBroadcast  MessageType = "broadcast"
)

// InMemoryMessageBus implements MessageBus
type InMemoryMessageBus struct {
    channels map[string]chan Message
    mu       sync.RWMutex
    logger   *zap.Logger
}

func NewInMemoryMessageBus() *InMemoryMessageBus {
    return &InMemoryMessageBus{
        channels: make(map[string]chan Message),
        logger:   zap.NewNop(),
    }
}

func (mb *InMemoryMessageBus) Send(ctx context.Context, message Message) error {
    mb.mu.RLock()
    ch, exists := mb.channels[message.To]
    mb.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("agent %s not subscribed", message.To)
    }
    
    select {
    case ch <- message:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(5 * time.Second):
        return fmt.Errorf("timeout sending message to %s", message.To)
    }
}

func (mb *InMemoryMessageBus) Subscribe(agentID string) <-chan Message {
    mb.mu.Lock()
    defer mb.mu.Unlock()
    
    ch := make(chan Message, 100) // Buffered channel
    mb.channels[agentID] = ch
    
    return ch
}

func (mb *InMemoryMessageBus) Broadcast(ctx context.Context, message Message) error {
    mb.mu.RLock()
    channels := make([]chan Message, 0, len(mb.channels))
    for _, ch := range mb.channels {
        channels = append(channels, ch)
    }
    mb.mu.RUnlock()
    
    // Send to all channels concurrently
    var wg sync.WaitGroup
    errCh := make(chan error, len(channels))
    
    for _, ch := range channels {
        wg.Add(1)
        go func(c chan Message) {
            defer wg.Done()
            
            select {
            case c <- message:
            case <-ctx.Done():
                errCh <- ctx.Err()
            case <-time.After(5 * time.Second):
                errCh <- fmt.Errorf("timeout broadcasting message")
            }
        }(ch)
    }
    
    wg.Wait()
    close(errCh)
    
    // Check for errors
    for err := range errCh {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

### Agent Coordination

```go
// AgentCoordinator manages agent interactions
type AgentCoordinator interface {
    RegisterAgent(agent Agent) error
    UnregisterAgent(agentID string) error
    GetAgent(agentID string) (Agent, error)
    ListAgents() []Agent
    
    // Coordination patterns
    ExecuteSequential(ctx context.Context, agents []string, input interface{}) ([]interface{}, error)
    ExecuteParallel(ctx context.Context, agents []string, input interface{}) ([]interface{}, error)
    ExecutePipeline(ctx context.Context, agents []string, input interface{}) (interface{}, error)
    
    // Communication
    SendMessage(ctx context.Context, from, to string, message Message) error
    BroadcastMessage(ctx context.Context, from string, message Message) error
}

type DefaultAgentCoordinator struct {
    agents    map[string]Agent
    messageBus MessageBus
    scheduler Scheduler
    mu        sync.RWMutex
    logger    *zap.Logger
}

func NewDefaultAgentCoordinator(messageBus MessageBus) *DefaultAgentCoordinator {
    return &DefaultAgentCoordinator{
        agents:     make(map[string]Agent),
        messageBus: messageBus,
        scheduler:  NewFIFOScheduler(),
        logger:     zap.NewNop(),
    }
}

func (c *DefaultAgentCoordinator) ExecuteParallel(ctx context.Context, agentIDs []string, input interface{}) ([]interface{}, error) {
    // Prepare agents
    agents := make([]Agent, 0, len(agentIDs))
    for _, id := range agentIDs {
        agent, err := c.GetAgent(id)
        if err != nil {
            return nil, fmt.Errorf("agent %s not found: %w", id, err)
        }
        agents = append(agents, agent)
    }
    
    // Execute in parallel
    results := make([]interface{}, len(agents))
    errors := make([]error, len(agents))
    
    var wg sync.WaitGroup
    for i, agent := range agents {
        wg.Add(1)
        go func(idx int, a Agent) {
            defer wg.Done()
            
            result, err := a.Execute(ctx, input)
            results[idx] = result
            errors[idx] = err
        }(i, agent)
    }
    
    wg.Wait()
    
    // Check for errors
    for i, err := range errors {
        if err != nil {
            return nil, fmt.Errorf("agent %s failed: %w", agentIDs[i], err)
        }
    }
    
    return results, nil
}

func (c *DefaultAgentCoordinator) ExecutePipeline(ctx context.Context, agentIDs []string, input interface{}) (interface{}, error) {
    currentInput := input
    
    for _, agentID := range agentIDs {
        agent, err := c.GetAgent(agentID)
        if err != nil {
            return nil, fmt.Errorf("agent %s not found: %w", agentID, err)
        }
        
        result, err := agent.Execute(ctx, currentInput)
        if err != nil {
            return nil, fmt.Errorf("agent %s failed: %w", agentID, err)
        }
        
        currentInput = result
    }
    
    return currentInput, nil
}
```

---

## Agent Registry and Discovery

### Agent Registry

```go
// AgentRegistry manages agent registration and discovery
type AgentRegistry interface {
    Register(agent Agent, metadata AgentMetadata) error
    Unregister(agentID string) error
    Get(agentID string) (Agent, error)
    List() []AgentInfo
    Search(criteria SearchCriteria) []AgentInfo
    
    // Discovery
    DiscoverByCapability(capabilities ...string) []AgentInfo
    DiscoverByType(agentType AgentType) []AgentInfo
    DiscoverByTags(tags ...string) []AgentInfo
}

type AgentMetadata struct {
    Name         string            `json:"name"`
    Description  string            `json:"description"`
    Version      string            `json:"version"`
    Author       string            `json:"author"`
    Tags         []string          `json:"tags"`
    Capabilities []string          `json:"capabilities"`
    Dependencies []string          `json:"dependencies"`
    Config       map[string]interface{} `json:"config"`
    Status       AgentStatus       `json:"status"`
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    time.Time         `json:"updated_at"`
}

type AgentInfo struct {
    Agent    Agent         `json:"-"`
    Metadata AgentMetadata `json:"metadata"`
    Runtime  RuntimeInfo   `json:"runtime"`
}

type RuntimeInfo struct {
    Uptime       time.Duration `json:"uptime"`
    Executions   int64         `json:"executions"`
    LastUsed     time.Time     `json:"last_used"`
    ErrorCount   int64         `json:"error_count"`
    AverageTime  time.Duration `json:"average_time"`
    MemoryUsage  int64         `json:"memory_usage"`
}

// InMemoryAgentRegistry implements AgentRegistry
type InMemoryAgentRegistry struct {
    agents   map[string]AgentInfo
    indexes  map[string]map[string][]string // capability/type/tag -> agent IDs
    mu       sync.RWMutex
    logger   *zap.Logger
}

func NewInMemoryAgentRegistry() *InMemoryAgentRegistry {
    return &InMemoryAgentRegistry{
        agents: make(map[string]AgentInfo),
        indexes: map[string]map[string][]string{
            "capability": make(map[string][]string),
            "type":       make(map[string][]string),
            "tag":        make(map[string][]string),
        },
        logger: zap.NewNop(),
    }
}

func (r *InMemoryAgentRegistry) Register(agent Agent, metadata AgentMetadata) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    agentID := agent.ID()
    
    // Create agent info
    info := AgentInfo{
        Agent:    agent,
        Metadata: metadata,
        Runtime: RuntimeInfo{
            Uptime:      0,
            Executions:  0,
            LastUsed:    time.Time{},
            ErrorCount:  0,
            AverageTime: 0,
            MemoryUsage: 0,
        },
    }
    
    // Store agent
    r.agents[agentID] = info
    
    // Update indexes
    r.updateIndexes(agentID, metadata)
    
    r.logger.Info("Agent registered",
        zap.String("agent_id", agentID),
        zap.String("name", metadata.Name),
        zap.String("type", string(agent.Type())),
    )
    
    return nil
}

func (r *InMemoryAgentRegistry) updateIndexes(agentID string, metadata AgentMetadata) {
    // Capability index
    for _, capability := range metadata.Capabilities {
        r.indexes["capability"][capability] = append(r.indexes["capability"][capability], agentID)
    }
    
    // Type index
    agentType := string(metadata.Config["type"].(AgentType))
    r.indexes["type"][agentType] = append(r.indexes["type"][agentType], agentID)
    
    // Tag index
    for _, tag := range metadata.Tags {
        r.indexes["tag"][tag] = append(r.indexes["tag"][tag], agentID)
    }
}

func (r *InMemoryAgentRegistry) DiscoverByCapability(capabilities ...string) []AgentInfo {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    if len(capabilities) == 0 {
        return nil
    }
    
    // Find agents with all required capabilities
    var candidateIDs []string
    for i, capability := range capabilities {
        agentIDs := r.indexes["capability"][capability]
        
        if i == 0 {
            candidateIDs = agentIDs
        } else {
            // Intersection
            candidateIDs = r.intersect(candidateIDs, agentIDs)
        }
        
        if len(candidateIDs) == 0 {
            break
        }
    }
    
    // Build result
    var result []AgentInfo
    for _, agentID := range candidateIDs {
        if info, ok := r.agents[agentID]; ok {
            result = append(result, info)
        }
    }
    
    return result
}

func (r *InMemoryAgentRegistry) intersect(a, b []string) []string {
    m := make(map[string]bool)
    for _, item := range a {
        m[item] = true
    }
    
    var result []string
    for _, item := range b {
        if m[item] {
            result = append(result, item)
        }
    }
    
    return result
}
```

---

## Agent Metrics and Monitoring

### Agent Metrics

```go
// AgentMetrics tracks agent performance and behavior
type AgentMetrics struct {
    executions      prometheus.Counter
    executionTime   prometheus.Histogram
    errors          *prometheus.CounterVec
    activeAgents    prometheus.Gauge
    memoryUsage     prometheus.Gauge
    stateSize       prometheus.Gauge
    messagesSent    prometheus.Counter
    messagesReceived prometheus.Counter
}

func NewAgentMetrics(agentID string) *AgentMetrics {
    return &AgentMetrics{
        executions: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "agent_executions_total",
            Help: "Total number of agent executions",
            ConstLabels: prometheus.Labels{"agent_id": agentID},
        }),
        executionTime: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "agent_execution_duration_seconds",
            Help: "Agent execution duration",
            ConstLabels: prometheus.Labels{"agent_id": agentID},
            Buckets: prometheus.DefBuckets,
        }),
        errors: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "agent_errors_total",
            Help: "Total number of agent errors",
            ConstLabels: prometheus.Labels{"agent_id": agentID},
        }, []string{"error_type"}),
        activeAgents: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "agent_active",
            Help: "Whether agent is active",
            ConstLabels: prometheus.Labels{"agent_id": agentID},
        }),
        memoryUsage: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "agent_memory_usage_bytes",
            Help: "Agent memory usage in bytes",
            ConstLabels: prometheus.Labels{"agent_id": agentID},
        }),
        stateSize: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "agent_state_size_bytes",
            Help: "Agent state size in bytes",
            ConstLabels: prometheus.Labels{"agent_id": agentID},
        }),
        messagesSent: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "agent_messages_sent_total",
            Help: "Total number of messages sent by agent",
            ConstLabels: prometheus.Labels{"agent_id": agentID},
        }),
        messagesReceived: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "agent_messages_received_total",
            Help: "Total number of messages received by agent",
            ConstLabels: prometheus.Labels{"agent_id": agentID},
        }),
    }
}

func (m *AgentMetrics) RecordExecution(duration time.Duration) {
    m.executions.Inc()
    m.executionTime.Observe(duration.Seconds())
}

func (m *AgentMetrics) RecordError(errorType string) {
    m.errors.WithLabelValues(errorType).Inc()
}

func (m *AgentMetrics) SetActive(active bool) {
    if active {
        m.activeAgents.Set(1)
    } else {
        m.activeAgents.Set(0)
    }
}

func (m *AgentMetrics) UpdateMemoryUsage(bytes int64) {
    m.memoryUsage.Set(float64(bytes))
}

func (m *AgentMetrics) UpdateStateSize(bytes int64) {
    m.stateSize.Set(float64(bytes))
}

func (m *AgentMetrics) RecordMessageSent() {
    m.messagesSent.Inc()
}

func (m *AgentMetrics) RecordMessageReceived() {
    m.messagesReceived.Inc()
}
```

### Health Monitoring

```go
// AgentHealthMonitor monitors agent health
type AgentHealthMonitor struct {
    agents     map[string]Agent
    checks     []HealthCheck
    interval   time.Duration
    alerter    Alerter
    logger     *zap.Logger
    mu         sync.RWMutex
}

type HealthCheck interface {
    Name() string
    Check(ctx context.Context, agent Agent) HealthCheckResult
    Severity() HealthSeverity
}

type HealthCheckResult struct {
    Healthy   bool          `json:"healthy"`
    Message   string        `json:"message"`
    Latency   time.Duration `json:"latency"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Timestamp time.Time     `json:"timestamp"`
}

type HealthSeverity string

const (
    HealthSeverityInfo     HealthSeverity = "info"
    HealthSeverityWarning  HealthSeverity = "warning"
    HealthSeverityCritical HealthSeverity = "critical"
)

func NewAgentHealthMonitor(interval time.Duration) *AgentHealthMonitor {
    return &AgentHealthMonitor{
        agents:   make(map[string]Agent),
        checks:   make([]HealthCheck, 0),
        interval: interval,
        logger:   zap.NewNop(),
    }
}

func (m *AgentHealthMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(m.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            m.runHealthChecks(ctx)
        case <-ctx.Done():
            return
        }
    }
}

func (m *AgentHealthMonitor) runHealthChecks(ctx context.Context) {
    m.mu.RLock()
    agents := make(map[string]Agent)
    for id, agent := range m.agents {
        agents[id] = agent
    }
    m.mu.RUnlock()
    
    for agentID, agent := range agents {
        for _, check := range m.checks {
            result := check.Check(ctx, agent)
            
            if !result.Healthy {
                m.logger.Warn("Agent health check failed",
                    zap.String("agent_id", agentID),
                    zap.String("check", check.Name()),
                    zap.String("message", result.Message),
                    zap.String("severity", string(check.Severity())),
                )
                
                if m.alerter != nil {
                    m.alerter.Alert(Alert{
                        Type:     "agent_health_check_failed",
                        Severity: check.Severity(),
                        Message:  result.Message,
                        Metadata: map[string]interface{}{
                            "agent_id": agentID,
                            "check":    check.Name(),
                            "result":   result,
                        },
}
                }
            }
        }
    }
}

// Built-in health checks
type ResponsivenessCheck struct {
    timeout time.Duration
}

func NewResponsivenessCheck(timeout time.Duration) *ResponsivenessCheck {
    return &ResponsivenessCheck{timeout: timeout}
}

func (c *ResponsivenessCheck) Name() string {
    return "responsiveness"
}

func (c *ResponsivenessCheck) Severity() HealthSeverity {
    return HealthSeverityCritical
}

func (c *ResponsivenessCheck) Check(ctx context.Context, agent Agent) HealthCheckResult {
    start := time.Now()
    
    // Create timeout context
    checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
    defer cancel()
    
    // Simple ping
    _, err := agent.Execute(checkCtx, map[string]interface{}{
        "type": "health_ping",
        "timestamp": time.Now().Unix(),
}
    
    latency := time.Since(start)
    
    if err != nil {
        return HealthCheckResult{
            Healthy:   false,
            Message:   fmt.Sprintf("Agent not responsive: %v", err),
            Latency:   latency,
            Timestamp: time.Now(),
        }
    }
    
    if latency > c.timeout/2 {
        return HealthCheckResult{
            Healthy: false,
            Message: fmt.Sprintf("Agent responding slowly: %v", latency),
            Latency: latency,
            Timestamp: time.Now(),
        }
    }
    
    return HealthCheckResult{
        Healthy:   true,
        Message:   "Agent responsive",
        Latency:   latency,
        Timestamp: time.Now(),
    }
}
```

---

## Configuration and Deployment

### Agent Configuration

```go
// AgentConfig defines agent configuration
type AgentConfig struct {
    // Basic configuration
    Name        string                 `yaml:"name" json:"name"`
    Type        AgentType              `yaml:"type" json:"type"`
    Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
    
    // Runtime configuration
    Concurrency int                    `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
    Timeout     time.Duration          `yaml:"timeout,omitempty" json:"timeout,omitempty"`
    RetryPolicy *RetryPolicyConfig     `yaml:"retry_policy,omitempty" json:"retry_policy,omitempty"`
    
    // State configuration
    StatePersistence *StatePersistenceConfig `yaml:"state_persistence,omitempty" json:"state_persistence,omitempty"`
    
    // Communication configuration
    MessageBus *MessageBusConfig      `yaml:"message_bus,omitempty" json:"message_bus,omitempty"`
    EventBus   *EventBusConfig        `yaml:"event_bus,omitempty" json:"event_bus,omitempty"`
    
    // Monitoring configuration
    Metrics    *MetricsConfig         `yaml:"metrics,omitempty" json:"metrics,omitempty"`
    Health     *HealthConfig          `yaml:"health,omitempty" json:"health,omitempty"`
    Logging    *LoggingConfig         `yaml:"logging,omitempty" json:"logging,omitempty"`
    
    // Custom configuration
    Custom     map[string]interface{} `yaml:"custom,omitempty" json:"custom,omitempty"`
}

type RetryPolicyConfig struct {
    MaxAttempts     int           `yaml:"max_attempts" json:"max_attempts"`
    InitialBackoff  time.Duration `yaml:"initial_backoff" json:"initial_backoff"`
    MaxBackoff      time.Duration `yaml:"max_backoff" json:"max_backoff"`
    BackoffStrategy string        `yaml:"backoff_strategy" json:"backoff_strategy"` // "linear", "exponential"
    RetryableErrors []string      `yaml:"retryable_errors,omitempty" json:"retryable_errors,omitempty"`
}

type StatePersistenceConfig struct {
    Type       string                 `yaml:"type" json:"type"` // "memory", "file", "redis", "database"
    Connection string                 `yaml:"connection,omitempty" json:"connection,omitempty"`
    Config     map[string]interface{} `yaml:"config,omitempty" json:"config,omitempty"`
}

// Example configuration file
const ExampleAgentConfig = `
name: "document_processor"
type: "llm"
description: "AI agent for processing documents"

concurrency: 5
timeout: 30s

retry_policy:
  max_attempts: 3
  initial_backoff: 1s
  max_backoff: 30s
  backoff_strategy: "exponential"
  retryable_errors:
    - "timeout"
    - "rate_limit"
    - "temporary_failure"

state_persistence:
  type: "redis"
  connection: "redis://localhost:6379"
  config:
    db: 0
    key_prefix: "agent_state:"

message_bus:
  type: "redis"
  connection: "redis://localhost:6379"
  config:
    channel_prefix: "agent_messages:"

metrics:
  enabled: true
  endpoint: "http://prometheus:9090"
  labels:
    environment: "production"
    team: "ai"

health:
  enabled: true
  port: 8080
  checks:
    - name: "responsiveness"
      interval: 30s
      timeout: 5s
    - name: "memory_usage"
      interval: 60s
      threshold: "512MB"

logging:
  level: "info"
  format: "json"
  output: "stdout"

custom:
  provider: "openai"
  model: "gpt-4o"
  tools:
    - "file_read"
    - "web_fetch"
    - "json_process"
`
```

### Agent Factory

```go
// AgentFactory creates agents from configuration
type AgentFactory struct {
    builders map[AgentType]AgentBuilder
    registry AgentRegistry
}

type AgentBuilder interface {
    Build(config AgentConfig) (Agent, error)
    Validate(config AgentConfig) error
    GetRequiredFields() []string
}

func NewAgentFactory(registry AgentRegistry) *AgentFactory {
    factory := &AgentFactory{
        builders: make(map[AgentType]AgentBuilder),
        registry: registry,
    }
    
    // Register built-in builders
    factory.RegisterBuilder(AgentTypeLLM, NewLLMAgentBuilder())
    factory.RegisterBuilder(AgentTypeWorkflow, NewWorkflowAgentBuilder())
    factory.RegisterBuilder(AgentTypeComposite, NewCompositeAgentBuilder())
    
    return factory
}

func (f *AgentFactory) RegisterBuilder(agentType AgentType, builder AgentBuilder) {
    f.builders[agentType] = builder
}

func (f *AgentFactory) CreateAgent(config AgentConfig) (Agent, error) {
    builder, ok := f.builders[config.Type]
    if !ok {
        return nil, fmt.Errorf("no builder for agent type: %s", config.Type)
    }
    
    // Validate configuration
    if err := builder.Validate(config); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    // Build agent
    agent, err := builder.Build(config)
    if err != nil {
        return nil, fmt.Errorf("failed to build agent: %w", err)
    }
    
    // Register agent
    metadata := AgentMetadata{
        Name:        config.Name,
        Description: config.Description,
        Version:     "1.0.0", // Default version
        Config:      config.Custom,
        Status:      AgentStatusIdle,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    if err := f.registry.Register(agent, metadata); err != nil {
        return nil, fmt.Errorf("failed to register agent: %w", err)
    }
    
    return agent, nil
}

func (f *AgentFactory) CreateFromFile(configPath string) (Agent, error) {
    data, err := ioutil.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    var config AgentConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return f.CreateAgent(config)
}
```

---

## Best Practices

### 1. Agent Design
- Keep agents focused on single responsibilities
- Use composition over inheritance for complex behaviors
- Implement proper error handling and recovery
- Design for testability and observability

### 2. State Management
- Minimize stateful agents where possible
- Use immutable state patterns
- Implement proper state persistence and recovery
- Consider state versioning for migrations

### 3. Communication
- Use async messaging for loose coupling
- Implement proper timeout and retry logic
- Design idempotent operations
- Use structured message formats

### 4. Monitoring
- Implement comprehensive health checks
- Track key performance metrics
- Use structured logging
- Set up proper alerting

### 5. Deployment
- Use configuration management
- Implement graceful shutdown
- Design for horizontal scaling
- Plan for disaster recovery

---

## Next Steps

- **[LLM Agents](llm-agents.md)** - AI-powered agents with tool support
- **[Workflow Agents](workflow-agents.md)** - Sequential, parallel, and conditional patterns
- **[Multi-Agent Systems](multi-agent-systems.md)** - Coordination and communication
- **[State Management](state-management.md)** - Agent state and data flow
- **[Agent API Reference](../../technical/api-reference/agents.md)** - Detailed API documentation