# Multi-Agent Systems: Coordination and Communication

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [Agents](../../technical/agents) / Multi-Agent Systems**

Comprehensive guide to building and orchestrating multi-agent systems in Go-LLMs, covering agent coordination patterns, communication protocols, consensus mechanisms, distributed task execution, and collaborative problem-solving architectures.

## Multi-Agent System Architecture

### Core Multi-Agent Interfaces

```go
// MultiAgentSystem orchestrates multiple agents
type MultiAgentSystem interface {
    // Agent management
    AddAgent(agent Agent) error
    RemoveAgent(agentID string) error
    GetAgent(agentID string) (Agent, error)
    ListAgents() []AgentInfo
    
    // Coordination
    Coordinate(ctx context.Context, task Task) (*CoordinationResult, error)
    ExecuteDistributed(ctx context.Context, request DistributedRequest) (*DistributedResult, error)
    
    // Communication
    Broadcast(ctx context.Context, message Message) error
    SendMessage(ctx context.Context, fromID, toID string, message Message) error
    Subscribe(agentID string, messageType MessageType) (<-chan Message, error)
    
    // Consensus
    ReachConsensus(ctx context.Context, proposal Proposal) (*ConsensusResult, error)
    Vote(ctx context.Context, ballot Ballot) (*VoteResult, error)
    
    // Monitoring
    GetSystemStatus() SystemStatus
    GetMetrics() SystemMetrics
    
    // Lifecycle
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Pause() error
    Resume() error
}

// AgentCoordinator defines coordination strategies
type AgentCoordinator interface {
    // Task distribution
    DistributeTask(ctx context.Context, task Task, agents []Agent) (*TaskDistribution, error)
    CollectResults(ctx context.Context, taskID string) (*CollectedResults, error)
    
    // Load balancing
    SelectAgent(agents []Agent, criteria SelectionCriteria) (Agent, error)
    BalanceLoad(agents []Agent, tasks []Task) (map[string][]Task, error)
    
    // Synchronization
    Synchronize(ctx context.Context, agents []Agent) error
    WaitForCompletion(ctx context.Context, taskIDs []string) error
    
    // Conflict resolution
    ResolveConflict(conflict Conflict) (*Resolution, error)
    Arbitrate(disputes []Dispute) (*Arbitration, error)
}

// CommunicationBus facilitates agent communication
type CommunicationBus interface {
    // Messaging
    Send(ctx context.Context, message Message) error
    Receive(agentID string) (<-chan Message, error)
    
    // Broadcasting
    Broadcast(ctx context.Context, message Message, filter MessageFilter) error
    Multicast(ctx context.Context, message Message, recipients []string) error
    
    // Routing
    Route(message Message) ([]string, error)
    SetRoutingRule(rule RoutingRule) error
    
    // Quality of Service
    SetQoS(qos QoSConfig) error
    GetDeliveryStatus(messageID string) DeliveryStatus
    
    // Management
    Subscribe(agentID string, topics []string) error
    Unsubscribe(agentID string, topics []string) error
}

// Task represents work to be distributed among agents
type Task struct {
    ID          string                 `json:"id"`
    Type        TaskType               `json:"type"`
    Priority    Priority               `json:"priority"`
    Payload     interface{}            `json:"payload"`
    Requirements TaskRequirements      `json:"requirements"`
    Dependencies []string              `json:"dependencies,omitempty"`
    Deadline     *time.Time            `json:"deadline,omitempty"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type TaskType string

const (
    TaskTypeSimple      TaskType = "simple"      // Single agent task
    TaskTypeComposite   TaskType = "composite"   // Multi-step task
    TaskTypeParallel    TaskType = "parallel"    // Parallel execution
    TaskTypeSequential  TaskType = "sequential"  // Sequential execution
    TaskTypeCollaborative TaskType = "collaborative" // Requires collaboration
    TaskTypeCompetitive TaskType = "competitive" // Agent competition
)

type Priority int

const (
    PriorityLow Priority = iota
    PriorityNormal
    PriorityHigh
    PriorityCritical
    PriorityEmergency
)
```

### Default Multi-Agent System Implementation

```go
// DefaultMultiAgentSystem implements MultiAgentSystem
type DefaultMultiAgentSystem struct {
    // Core components
    agents       map[string]Agent
    coordinator  AgentCoordinator
    communicator CommunicationBus
    consensus    ConsensusEngine
    
    // Task management
    taskQueue    TaskQueue
    scheduler    TaskScheduler
    executor     TaskExecutor
    
    // State management
    state        SystemState
    topology     NetworkTopology
    
    // Monitoring
    monitor      SystemMonitor
    metrics      *SystemMetrics
    logger       *zap.Logger
    
    // Synchronization
    mu           sync.RWMutex
    status       SystemStatus
    
    // Configuration
    config       MultiAgentConfig
}

type MultiAgentConfig struct {
    MaxAgents        int                    `yaml:"max_agents" json:"max_agents"`
    CommunicationConfig CommunicationConfig `yaml:"communication" json:"communication"`
    CoordinationConfig  CoordinationConfig  `yaml:"coordination" json:"coordination"`
    ConsensusConfig     ConsensusConfig     `yaml:"consensus" json:"consensus"`
    FailoverConfig      FailoverConfig      `yaml:"failover" json:"failover"`
    MonitoringConfig    MonitoringConfig    `yaml:"monitoring" json:"monitoring"`
}

func NewMultiAgentSystem(config MultiAgentConfig) *DefaultMultiAgentSystem {
    system := &DefaultMultiAgentSystem{
        agents:       make(map[string]Agent),
        state:        NewSystemState(),
        topology:     NewNetworkTopology(),
        metrics:      NewSystemMetrics(),
        logger:       zap.NewNop(),
        status:       SystemStatusIdle,
        config:       config,
    }
    
    // Initialize components
    system.initializeComponents()
    
    return system
}

func (s *DefaultMultiAgentSystem) initializeComponents() {
    // Initialize coordinator
    s.coordinator = NewDefaultAgentCoordinator(s.config.CoordinationConfig)
    
    // Initialize communication bus
    s.communicator = NewCommunicationBus(s.config.CommunicationConfig)
    
    // Initialize consensus engine
    s.consensus = NewConsensusEngine(s.config.ConsensusConfig)
    
    // Initialize task management
    s.taskQueue = NewPriorityTaskQueue()
    s.scheduler = NewRoundRobinScheduler()
    s.executor = NewDistributedTaskExecutor()
    
    // Initialize monitoring
    s.monitor = NewSystemMonitor(s.config.MonitoringConfig)
}

func (s *DefaultMultiAgentSystem) AddAgent(agent Agent) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if len(s.agents) >= s.config.MaxAgents {
        return fmt.Errorf("maximum number of agents reached: %d", s.config.MaxAgents)
    }
    
    agentID := agent.ID()
    if _, exists := s.agents[agentID]; exists {
        return fmt.Errorf("agent %s already exists", agentID)
    }
    
    // Add agent to system
    s.agents[agentID] = agent
    
    // Register with coordinator
    if err := s.coordinator.RegisterAgent(agent); err != nil {
        delete(s.agents, agentID)
        return fmt.Errorf("failed to register agent with coordinator: %w", err)
    }
    
    // Setup communication
    if err := s.setupAgentCommunication(agent); err != nil {
        delete(s.agents, agentID)
        s.coordinator.UnregisterAgent(agentID)
        return fmt.Errorf("failed to setup communication for agent: %w", err)
    }
    
    // Update topology
    s.topology.AddNode(agentID, agent)
    
    // Start monitoring
    s.monitor.StartMonitoring(agentID, agent)
    
    s.logger.Info("Agent added to multi-agent system",
        zap.String("agent_id", agentID),
        zap.String("agent_type", string(agent.Type())),
        zap.Int("total_agents", len(s.agents)),
    )
    
    return nil
}

func (s *DefaultMultiAgentSystem) setupAgentCommunication(agent Agent) error {
    agentID := agent.ID()
    
    // Create communication channel
    msgChan, err := s.communicator.Receive(agentID)
    if err != nil {
        return err
    }
    
    // Start message handling for agent
    go s.handleAgentMessages(agentID, msgChan)
    
    return nil
}

func (s *DefaultMultiAgentSystem) handleAgentMessages(agentID string, msgChan <-chan Message) {
    for message := range msgChan {
        agent, exists := s.agents[agentID]
        if !exists {
            s.logger.Warn("Received message for unknown agent",
                zap.String("agent_id", agentID),
                zap.String("message_id", message.ID),
            )
            continue
        }
        
        // Handle message based on type
        switch message.Type {
        case MessageTypeTask:
            s.handleTaskMessage(agent, message)
        case MessageTypeCommand:
            s.handleCommandMessage(agent, message)
        case MessageTypeQuery:
            s.handleQueryMessage(agent, message)
        case MessageTypeEvent:
            s.handleEventMessage(agent, message)
        default:
            s.logger.Warn("Unknown message type",
                zap.String("agent_id", agentID),
                zap.String("message_type", string(message.Type)),
            )
        }
    }
}
```

---

## Coordination Patterns

### Task Distribution and Load Balancing

```go
// TaskDistributor handles task distribution strategies
type TaskDistributor interface {
    Distribute(ctx context.Context, task Task, agents []Agent) (*TaskDistribution, error)
    GetStrategy() DistributionStrategy
    SetStrategy(strategy DistributionStrategy)
}

type DistributionStrategy interface {
    SelectAgents(task Task, agents []Agent) ([]Agent, error)
    SplitTask(task Task, agents []Agent) ([]SubTask, error)
    GetName() string
}

// Round-robin distribution strategy
type RoundRobinStrategy struct {
    lastIndex int
    mu        sync.Mutex
}

func NewRoundRobinStrategy() *RoundRobinStrategy {
    return &RoundRobinStrategy{}
}

func (s *RoundRobinStrategy) SelectAgents(task Task, agents []Agent) ([]Agent, error) {
    if len(agents) == 0 {
        return nil, fmt.Errorf("no agents available")
    }
    
    s.mu.Lock()
    defer s.mu.Unlock()
    
    selectedAgent := agents[s.lastIndex%len(agents)]
    s.lastIndex++
    
    return []Agent{selectedAgent}, nil
}

func (s *RoundRobinStrategy) SplitTask(task Task, agents []Agent) ([]SubTask, error) {
    // For round-robin, we don't split the task
    return []SubTask{{
        ID:       task.ID + "_0",
        ParentID: task.ID,
        Payload:  task.Payload,
        AgentID:  "", // Will be assigned during distribution
    }}, nil
}

func (s *RoundRobinStrategy) GetName() string {
    return "round_robin"
}

// Load-based distribution strategy
type LoadBasedStrategy struct {
    loadBalancer LoadBalancer
}

type LoadBalancer interface {
    GetAgentLoad(agentID string) (LoadMetrics, error)
    SelectLeastLoaded(agents []Agent) (Agent, error)
}

func NewLoadBasedStrategy(loadBalancer LoadBalancer) *LoadBasedStrategy {
    return &LoadBasedStrategy{
        loadBalancer: loadBalancer,
    }
}

func (s *LoadBasedStrategy) SelectAgents(task Task, agents []Agent) ([]Agent, error) {
    agent, err := s.loadBalancer.SelectLeastLoaded(agents)
    if err != nil {
        return nil, err
    }
    
    return []Agent{agent}, nil
}

func (s *LoadBasedStrategy) SplitTask(task Task, agents []Agent) ([]SubTask, error) {
    return []SubTask{{
        ID:       task.ID + "_0", 
        ParentID: task.ID,
        Payload:  task.Payload,
    }}, nil
}

func (s *LoadBasedStrategy) GetName() string {
    return "load_based"
}

// Capability-based distribution strategy
type CapabilityBasedStrategy struct {
    matcher CapabilityMatcher
}

type CapabilityMatcher interface {
    Match(requirements TaskRequirements, agent Agent) (float64, error)
    FindBestMatch(requirements TaskRequirements, agents []Agent) (Agent, error)
}

func NewCapabilityBasedStrategy(matcher CapabilityMatcher) *CapabilityBasedStrategy {
    return &CapabilityBasedStrategy{
        matcher: matcher,
    }
}

func (s *CapabilityBasedStrategy) SelectAgents(task Task, agents []Agent) ([]Agent, error) {
    agent, err := s.matcher.FindBestMatch(task.Requirements, agents)
    if err != nil {
        return nil, err
    }
    
    return []Agent{agent}, nil
}

func (s *CapabilityBasedStrategy) SplitTask(task Task, agents []Agent) ([]SubTask, error) {
    return []SubTask{{
        ID:       task.ID + "_0",
        ParentID: task.ID,
        Payload:  task.Payload,
    }}, nil
}

func (s *CapabilityBasedStrategy) GetName() string {
    return "capability_based"
}

// Parallel distribution strategy
type ParallelStrategy struct {
    splitter TaskSplitter
}

type TaskSplitter interface {
    Split(task Task, numParts int) ([]SubTask, error)
    CanSplit(task Task) bool
}

func NewParallelStrategy(splitter TaskSplitter) *ParallelStrategy {
    return &ParallelStrategy{
        splitter: splitter,
    }
}

func (s *ParallelStrategy) SelectAgents(task Task, agents []Agent) ([]Agent, error) {
    if !s.splitter.CanSplit(task) {
        return nil, fmt.Errorf("task cannot be split for parallel execution")
    }
    
    // Select all available agents for parallel execution
    return agents, nil
}

func (s *ParallelStrategy) SplitTask(task Task, agents []Agent) ([]SubTask, error) {
    return s.splitter.Split(task, len(agents))
}

func (s *ParallelStrategy) GetName() string {
    return "parallel"
}

// Task distribution implementation
type DefaultTaskDistributor struct {
    strategy DistributionStrategy
    logger   *zap.Logger
}

func NewDefaultTaskDistributor(strategy DistributionStrategy) *DefaultTaskDistributor {
    return &DefaultTaskDistributor{
        strategy: strategy,
        logger:   zap.NewNop(),
    }
}

func (d *DefaultTaskDistributor) Distribute(ctx context.Context, task Task, agents []Agent) (*TaskDistribution, error) {
    // Select agents for the task
    selectedAgents, err := d.strategy.SelectAgents(task, agents)
    if err != nil {
        return nil, fmt.Errorf("agent selection failed: %w", err)
    }
    
    // Split task if necessary
    subTasks, err := d.strategy.SplitTask(task, selectedAgents)
    if err != nil {
        return nil, fmt.Errorf("task splitting failed: %w", err)
    }
    
    // Assign subtasks to agents
    assignments := make([]TaskAssignment, 0)
    for i, subTask := range subTasks {
        if i < len(selectedAgents) {
            assignment := TaskAssignment{
                SubTask: subTask,
                AgentID: selectedAgents[i].ID(),
                AssignedAt: time.Now(),
            }
            assignments = append(assignments, assignment)
        }
    }
    
    distribution := &TaskDistribution{
        TaskID:      task.ID,
        Strategy:    d.strategy.GetName(),
        Assignments: assignments,
        CreatedAt:   time.Now(),
    }
    
    d.logger.Info("Task distributed",
        zap.String("task_id", task.ID),
        zap.String("strategy", d.strategy.GetName()),
        zap.Int("agents_selected", len(selectedAgents)),
        zap.Int("subtasks_created", len(subTasks)),
    )
    
    return distribution, nil
}

type TaskDistribution struct {
    TaskID      string           `json:"task_id"`
    Strategy    string           `json:"strategy"`
    Assignments []TaskAssignment `json:"assignments"`
    CreatedAt   time.Time        `json:"created_at"`
}

type TaskAssignment struct {
    SubTask    SubTask   `json:"subtask"`
    AgentID    string    `json:"agent_id"`
    AssignedAt time.Time `json:"assigned_at"`
    Status     TaskStatus `json:"status"`
}

type SubTask struct {
    ID       string      `json:"id"`
    ParentID string      `json:"parent_id"`
    Payload  interface{} `json:"payload"`
    AgentID  string      `json:"agent_id,omitempty"`
}

type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"
    TaskStatusAssigned  TaskStatus = "assigned"
    TaskStatusRunning   TaskStatus = "running"
    TaskStatusCompleted TaskStatus = "completed"
    TaskStatusFailed    TaskStatus = "failed"
    TaskStatusCancelled TaskStatus = "cancelled"
)
```

### Synchronization and Coordination

```go
// SynchronizationManager handles agent synchronization
type SynchronizationManager interface {
    CreateBarrier(name string, agentIDs []string) (*Barrier, error)
    WaitAtBarrier(ctx context.Context, barrierName, agentID string) error
    CreateSemaphore(name string, permits int) (*Semaphore, error)
    AcquireSemaphore(ctx context.Context, semaphoreName, agentID string) error
    ReleaseSemaphore(semaphoreName, agentID string) error
    CreateMutex(name string) (*DistributedMutex, error)
    LockMutex(ctx context.Context, mutexName, agentID string) error
    UnlockMutex(mutexName, agentID string) error
}

// Barrier for synchronizing multiple agents
type Barrier struct {
    Name       string    `json:"name"`
    AgentIDs   []string  `json:"agent_ids"`
    WaitingFor []string  `json:"waiting_for"`
    Completed  bool      `json:"completed"`
    CreatedAt  time.Time `json:"created_at"`
    mu         sync.Mutex
    cond       *sync.Cond
}

func NewBarrier(name string, agentIDs []string) *Barrier {
    barrier := &Barrier{
        Name:       name,
        AgentIDs:   make([]string, len(agentIDs)),
        WaitingFor: make([]string, len(agentIDs)),
        Completed:  false,
        CreatedAt:  time.Now(),
    }
    
    copy(barrier.AgentIDs, agentIDs)
    copy(barrier.WaitingFor, agentIDs)
    barrier.cond = sync.NewCond(&barrier.mu)
    
    return barrier
}

func (b *Barrier) Wait(ctx context.Context, agentID string) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    // Check if agent is expected
    found := false
    for _, id := range b.AgentIDs {
        if id == agentID {
            found = true
            break
        }
    }
    
    if !found {
        return fmt.Errorf("agent %s not expected at barrier %s", agentID, b.Name)
    }
    
    // Remove agent from waiting list
    for i, id := range b.WaitingFor {
        if id == agentID {
            b.WaitingFor = append(b.WaitingFor[:i], b.WaitingFor[i+1:]...)
            break
        }
    }
    
    // Check if all agents have arrived
    if len(b.WaitingFor) == 0 {
        b.Completed = true
        b.cond.Broadcast() // Wake up all waiting agents
        return nil
    }
    
    // Wait for other agents
    for !b.Completed {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            b.cond.Wait()
        }
    }
    
    return nil
}

// Distributed semaphore for resource control
type Semaphore struct {
    Name      string    `json:"name"`
    Permits   int       `json:"permits"`
    Available int       `json:"available"`
    Holders   []string  `json:"holders"`
    Queue     []string  `json:"queue"`
    CreatedAt time.Time `json:"created_at"`
    mu        sync.Mutex
    cond      *sync.Cond
}

func NewSemaphore(name string, permits int) *Semaphore {
    semaphore := &Semaphore{
        Name:      name,
        Permits:   permits,
        Available: permits,
        Holders:   make([]string, 0),
        Queue:     make([]string, 0),
        CreatedAt: time.Now(),
    }
    
    semaphore.cond = sync.NewCond(&semaphore.mu)
    
    return semaphore
}

func (s *Semaphore) Acquire(ctx context.Context, agentID string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Add to queue if necessary
    if s.Available == 0 {
        s.Queue = append(s.Queue, agentID)
    }
    
    // Wait for permit
    for s.Available == 0 || (len(s.Queue) > 0 && s.Queue[0] != agentID) {
        select {
        case <-ctx.Done():
            // Remove from queue if context cancelled
            s.removeFromQueue(agentID)
            return ctx.Err()
        default:
            s.cond.Wait()
        }
    }
    
    // Acquire permit
    s.Available--
    s.Holders = append(s.Holders, agentID)
    
    // Remove from queue
    s.removeFromQueue(agentID)
    
    return nil
}

func (s *Semaphore) Release(agentID string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Check if agent holds a permit
    holderIndex := -1
    for i, holder := range s.Holders {
        if holder == agentID {
            holderIndex = i
            break
        }
    }
    
    if holderIndex == -1 {
        return fmt.Errorf("agent %s does not hold a permit", agentID)
    }
    
    // Release permit
    s.Available++
    s.Holders = append(s.Holders[:holderIndex], s.Holders[holderIndex+1:]...)
    
    // Notify waiting agents
    s.cond.Signal()
    
    return nil
}

func (s *Semaphore) removeFromQueue(agentID string) {
    for i, id := range s.Queue {
        if id == agentID {
            s.Queue = append(s.Queue[:i], s.Queue[i+1:]...)
            break
        }
    }
}

// Coordination orchestrator
type CoordinationOrchestrator struct {
    barriers   map[string]*Barrier
    semaphores map[string]*Semaphore
    mutexes    map[string]*DistributedMutex
    events     chan CoordinationEvent
    logger     *zap.Logger
    mu         sync.RWMutex
}

type CoordinationEvent struct {
    Type      string                 `json:"type"`
    Source    string                 `json:"source"`
    Target    string                 `json:"target"`
    Data      map[string]interface{} `json:"data"`
    Timestamp time.Time              `json:"timestamp"`
}

func NewCoordinationOrchestrator() *CoordinationOrchestrator {
    return &CoordinationOrchestrator{
        barriers:   make(map[string]*Barrier),
        semaphores: make(map[string]*Semaphore),
        mutexes:    make(map[string]*DistributedMutex),
        events:     make(chan CoordinationEvent, 1000),
        logger:     zap.NewNop(),
    }
}

func (o *CoordinationOrchestrator) CreateBarrier(name string, agentIDs []string) (*Barrier, error) {
    o.mu.Lock()
    defer o.mu.Unlock()
    
    if _, exists := o.barriers[name]; exists {
        return nil, fmt.Errorf("barrier %s already exists", name)
    }
    
    barrier := NewBarrier(name, agentIDs)
    o.barriers[name] = barrier
    
    o.publishEvent(CoordinationEvent{
        Type:   "barrier_created",
        Source: "orchestrator",
        Target: name,
        Data: map[string]interface{}{
            "agent_ids": agentIDs,
        },
        Timestamp: time.Now(),
}
    
    return barrier, nil
}

func (o *CoordinationOrchestrator) WaitAtBarrier(ctx context.Context, barrierName, agentID string) error {
    o.mu.RLock()
    barrier, exists := o.barriers[barrierName]
    o.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("barrier %s not found", barrierName)
    }
    
    o.publishEvent(CoordinationEvent{
        Type:   "barrier_wait",
        Source: agentID,
        Target: barrierName,
        Data: map[string]interface{}{
            "agent_id": agentID,
        },
        Timestamp: time.Now(),
}
    
    err := barrier.Wait(ctx, agentID)
    
    if err == nil {
        o.publishEvent(CoordinationEvent{
            Type:   "barrier_passed",
            Source: agentID,
            Target: barrierName,
            Data: map[string]interface{}{
                "agent_id": agentID,
            },
            Timestamp: time.Now(),
}
    }
    
    return err
}

func (o *CoordinationOrchestrator) publishEvent(event CoordinationEvent) {
    select {
    case o.events <- event:
    default:
        o.logger.Warn("Event queue full, dropping event",
            zap.String("event_type", event.Type),
            zap.String("source", event.Source),
        )
    }
}

func (o *CoordinationOrchestrator) GetEvents() <-chan CoordinationEvent {
    return o.events
}
```

---

## Communication Protocols

### Message-Based Communication

```go
// MessageBus implementation for agent communication
type DefaultCommunicationBus struct {
    // Routing
    router       MessageRouter
    channels     map[string]chan Message
    subscriptions map[string][]string // agentID -> topics
    
    // Quality of Service
    qos          QoSManager
    delivery     DeliveryTracker
    
    // Message store
    messageStore MessageStore
    
    // Filters and middleware
    filters      []MessageFilter
    middleware   []MessageMiddleware
    
    // Monitoring
    metrics      *CommunicationMetrics
    logger       *zap.Logger
    
    // Synchronization
    mu           sync.RWMutex
    running      bool
}

type MessageRouter interface {
    Route(message Message) ([]string, error)
    AddRoute(route Route) error
    RemoveRoute(routeID string) error
    GetRoutes() []Route
}

type Route struct {
    ID          string           `json:"id"`
    Pattern     string           `json:"pattern"`     // Topic pattern
    Recipients  []string         `json:"recipients"`  // Agent IDs
    Conditions  []RouteCondition `json:"conditions"`  // Routing conditions
    Priority    int              `json:"priority"`
    Enabled     bool             `json:"enabled"`
}

type RouteCondition struct {
    Field    string      `json:"field"`    // Message field to check
    Operator string      `json:"operator"` // eq, ne, contains, regex
    Value    interface{} `json:"value"`    // Expected value
}

func NewCommunicationBus(config CommunicationConfig) *DefaultCommunicationBus {
    bus := &DefaultCommunicationBus{
        router:        NewTopicRouter(),
        channels:      make(map[string]chan Message),
        subscriptions: make(map[string][]string),
        filters:       make([]MessageFilter, 0),
        middleware:    make([]MessageMiddleware, 0),
        metrics:       NewCommunicationMetrics(),
        logger:        zap.NewNop(),
        running:       false,
    }
    
    // Initialize components
    bus.qos = NewQoSManager(config.QoS)
    bus.delivery = NewDeliveryTracker()
    bus.messageStore = NewMessageStore(config.Storage)
    
    return bus
}

func (b *DefaultCommunicationBus) Send(ctx context.Context, message Message) error {
    if !b.running {
        return fmt.Errorf("communication bus not running")
    }
    
    // Apply middleware
    processedMessage := message
    for _, mw := range b.middleware {
        var err error
        processedMessage, err = mw.Process(processedMessage)
        if err != nil {
            return fmt.Errorf("middleware processing failed: %w", err)
        }
    }
    
    // Apply filters
    for _, filter := range b.filters {
        if !filter.Accept(processedMessage) {
            b.logger.Debug("Message filtered out",
                zap.String("message_id", message.ID),
                zap.String("filter", fmt.Sprintf("%T", filter)),
            )
            return nil
        }
    }
    
    // Store message
    if err := b.messageStore.Store(processedMessage); err != nil {
        b.logger.Warn("Failed to store message",
            zap.String("message_id", message.ID),
            zap.Error(err),
        )
    }
    
    // Route message
    recipients, err := b.router.Route(processedMessage)
    if err != nil {
        return fmt.Errorf("message routing failed: %w", err)
    }
    
    // Deliver to recipients
    var deliveryErrors []error
    for _, recipientID := range recipients {
        if err := b.deliverToAgent(ctx, recipientID, processedMessage); err != nil {
            deliveryErrors = append(deliveryErrors, err)
            b.metrics.RecordDeliveryFailure(recipientID)
        } else {
            b.metrics.RecordDeliverySuccess(recipientID)
        }
    }
    
    // Track delivery
    b.delivery.Track(processedMessage.ID, recipients, deliveryErrors)
    
    // Record metrics
    b.metrics.RecordMessage(processedMessage)
    
    if len(deliveryErrors) > 0 {
        return fmt.Errorf("partial delivery failure: %d/%d failed", len(deliveryErrors), len(recipients))
    }
    
    return nil
}

func (b *DefaultCommunicationBus) deliverToAgent(ctx context.Context, agentID string, message Message) error {
    b.mu.RLock()
    channel, exists := b.channels[agentID]
    b.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("agent %s not connected", agentID)
    }
    
    // Apply QoS
    qosConfig := b.qos.GetQoS(agentID)
    timeout := qosConfig.DeliveryTimeout
    if timeout == 0 {
        timeout = 5 * time.Second
    }
    
    // Deliver with timeout
    select {
    case channel <- message:
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("delivery timeout for agent %s", agentID)
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (b *DefaultCommunicationBus) Broadcast(ctx context.Context, message Message, filter MessageFilter) error {
    b.mu.RLock()
    agentIDs := make([]string, 0, len(b.channels))
    for agentID := range b.channels {
        agentIDs = append(agentIDs, agentID)
    }
    b.mu.RUnlock()
    
    // Create broadcast message
    broadcastMessage := message
    broadcastMessage.Type = MessageTypeBroadcast
    broadcastMessage.To = "ALL"
    
    var deliveryErrors []error
    for _, agentID := range agentIDs {
        // Apply filter if provided
        if filter != nil {
            agentMessage := broadcastMessage
            agentMessage.To = agentID
            if !filter.Accept(agentMessage) {
                continue
            }
        }
        
        if err := b.deliverToAgent(ctx, agentID, broadcastMessage); err != nil {
            deliveryErrors = append(deliveryErrors, err)
        }
    }
    
    b.metrics.RecordBroadcast(len(agentIDs), len(deliveryErrors))
    
    if len(deliveryErrors) > 0 {
        return fmt.Errorf("broadcast partial failure: %d/%d failed", len(deliveryErrors), len(agentIDs))
    }
    
    return nil
}

func (b *DefaultCommunicationBus) Subscribe(agentID string, topics []string) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    // Create channel if not exists
    if _, exists := b.channels[agentID]; !exists {
        b.channels[agentID] = make(chan Message, 100) // Buffered channel
    }
    
    // Update subscriptions
    b.subscriptions[agentID] = topics
    
    // Update router
    for _, topic := range topics {
        route := Route{
            ID:         fmt.Sprintf("%s_%s", agentID, topic),
            Pattern:    topic,
            Recipients: []string{agentID},
            Priority:   1,
            Enabled:    true,
        }
        
        b.router.AddRoute(route)
    }
    
    b.logger.Info("Agent subscribed to topics",
        zap.String("agent_id", agentID),
        zap.Strings("topics", topics),
    )
    
    return nil
}

func (b *DefaultCommunicationBus) Receive(agentID string) (<-chan Message, error) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    channel, exists := b.channels[agentID]
    if !exists {
        return nil, fmt.Errorf("agent %s not subscribed", agentID)
    }
    
    return channel, nil
}

// Message middleware for processing
type MessageMiddleware interface {
    Process(message Message) (Message, error)
    Priority() int
}

// Encryption middleware
type EncryptionMiddleware struct {
    encryptor Encryptor
    priority  int
}

type Encryptor interface {
    Encrypt(data []byte) ([]byte, error)
    Decrypt(data []byte) ([]byte, error)
}

func NewEncryptionMiddleware(encryptor Encryptor) *EncryptionMiddleware {
    return &EncryptionMiddleware{
        encryptor: encryptor,
        priority:  100,
    }
}

func (m *EncryptionMiddleware) Process(message Message) (Message, error) {
    // Encrypt message content
    contentBytes, err := json.Marshal(message.Content)
    if err != nil {
        return message, fmt.Errorf("failed to marshal content: %w", err)
    }
    
    encryptedContent, err := m.encryptor.Encrypt(contentBytes)
    if err != nil {
        return message, fmt.Errorf("encryption failed: %w", err)
    }
    
    // Modify message
    processedMessage := message
    processedMessage.Content = base64.StdEncoding.EncodeToString(encryptedContent)
    processedMessage.Metadata["encrypted"] = true
    processedMessage.Metadata["encryption_type"] = "aes256"
    
    return processedMessage, nil
}

func (m *EncryptionMiddleware) Priority() int {
    return m.priority
}

// Compression middleware
type CompressionMiddleware struct {
    compressor Compressor
    threshold  int // Minimum size to compress
    priority   int
}

type Compressor interface {
    Compress(data []byte) ([]byte, error)
    Decompress(data []byte) ([]byte, error)
}

func NewCompressionMiddleware(compressor Compressor, threshold int) *CompressionMiddleware {
    return &CompressionMiddleware{
        compressor: compressor,
        threshold:  threshold,
        priority:   50,
    }
}

func (m *CompressionMiddleware) Process(message Message) (Message, error) {
    contentBytes, err := json.Marshal(message.Content)
    if err != nil {
        return message, fmt.Errorf("failed to marshal content: %w", err)
    }
    
    // Only compress if content is large enough
    if len(contentBytes) < m.threshold {
        return message, nil
    }
    
    compressedContent, err := m.compressor.Compress(contentBytes)
    if err != nil {
        return message, fmt.Errorf("compression failed: %w", err)
    }
    
    // Only use compressed version if it's actually smaller
    if len(compressedContent) >= len(contentBytes) {
        return message, nil
    }
    
    processedMessage := message
    processedMessage.Content = base64.StdEncoding.EncodeToString(compressedContent)
    processedMessage.Metadata["compressed"] = true
    processedMessage.Metadata["compression_type"] = "gzip"
    processedMessage.Metadata["original_size"] = len(contentBytes)
    
    return processedMessage, nil
}

func (m *CompressionMiddleware) Priority() int {
    return m.priority
}
```

---

## Consensus Mechanisms

### Consensus Engine

```go
// ConsensusEngine handles consensus protocols
type ConsensusEngine interface {
    // Consensus operations
    ProposeValue(ctx context.Context, proposal Proposal) (*ConsensusResult, error)
    Vote(ctx context.Context, ballot Ballot) error
    GetConsensusResult(proposalID string) (*ConsensusResult, error)
    
    // Protocol management
    SetProtocol(protocol ConsensusProtocol) error
    GetProtocol() ConsensusProtocol
    
    // Participant management
    AddParticipant(agentID string) error
    RemoveParticipant(agentID string) error
    GetParticipants() []string
}

type ConsensusProtocol interface {
    GetName() string
    ReachConsensus(ctx context.Context, proposal Proposal, participants []string) (*ConsensusResult, error)
    ValidateProposal(proposal Proposal) error
    GetRequiredVotes(totalParticipants int) int
}

type Proposal struct {
    ID          string                 `json:"id"`
    Type        ProposalType           `json:"type"`
    Content     interface{}            `json:"content"`
    ProposerID  string                 `json:"proposer_id"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

type ProposalType string

const (
    ProposalTypeTaskAssignment ProposalType = "task_assignment"
    ProposalTypeResourceAllocation ProposalType = "resource_allocation"
    ProposalTypeSystemChange   ProposalType = "system_change"
    ProposalTypeConflictResolution ProposalType = "conflict_resolution"
    ProposalTypeLeaderElection ProposalType = "leader_election"
)

type Ballot struct {
    ProposalID string      `json:"proposal_id"`
    VoterID    string      `json:"voter_id"`
    Vote       VoteChoice  `json:"vote"`
    Reason     string      `json:"reason,omitempty"`
    Timestamp  time.Time   `json:"timestamp"`
}

type VoteChoice string

const (
    VoteYes     VoteChoice = "yes"
    VoteNo      VoteChoice = "no"
    VoteAbstain VoteChoice = "abstain"
)

type ConsensusResult struct {
    ProposalID   string                 `json:"proposal_id"`
    Decision     Decision               `json:"decision"`
    Votes        []Ballot               `json:"votes"`
    Participants []string               `json:"participants"`
    StartTime    time.Time              `json:"start_time"`
    EndTime      time.Time              `json:"end_time"`
    Duration     time.Duration          `json:"duration"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type Decision string

const (
    DecisionAccepted Decision = "accepted"
    DecisionRejected Decision = "rejected"
    DecisionTimeout  Decision = "timeout"
    DecisionFailed   Decision = "failed"
)

// Majority consensus protocol
type MajorityConsensusProtocol struct {
    threshold float64 // e.g., 0.5 for simple majority, 0.67 for supermajority
    timeout   time.Duration
}

func NewMajorityConsensusProtocol(threshold float64, timeout time.Duration) *MajorityConsensusProtocol {
    return &MajorityConsensusProtocol{
        threshold: threshold,
        timeout:   timeout,
    }
}

func (p *MajorityConsensusProtocol) GetName() string {
    return "majority"
}

func (p *MajorityConsensusProtocol) ReachConsensus(ctx context.Context, proposal Proposal, participants []string) (*ConsensusResult, error) {
    result := &ConsensusResult{
        ProposalID:   proposal.ID,
        Participants: participants,
        StartTime:    time.Now(),
        Votes:        make([]Ballot, 0),
    }
    
    requiredVotes := p.GetRequiredVotes(len(participants))
    votesChan := make(chan Ballot, len(participants))
    
    // Start timeout timer
    timeoutCtx, cancel := context.WithTimeout(ctx, p.timeout)
    defer cancel()
    
    // Collect votes
    votesReceived := 0
    yesVotes := 0
    noVotes := 0
    
    for {
        select {
        case vote := <-votesChan:
            result.Votes = append(result.Votes, vote)
            votesReceived++
            
            switch vote.Vote {
            case VoteYes:
                yesVotes++
            case VoteNo:
                noVotes++
            }
            
            // Check if we have enough votes to make a decision
            if yesVotes >= requiredVotes {
                result.Decision = DecisionAccepted
                result.EndTime = time.Now()
                result.Duration = result.EndTime.Sub(result.StartTime)
                return result, nil
            }
            
            if noVotes > len(participants)-requiredVotes {
                result.Decision = DecisionRejected
                result.EndTime = time.Now()
                result.Duration = result.EndTime.Sub(result.StartTime)
                return result, nil
            }
            
            // Check if all votes received
            if votesReceived >= len(participants) {
                if yesVotes >= requiredVotes {
                    result.Decision = DecisionAccepted
                } else {
                    result.Decision = DecisionRejected
                }
                result.EndTime = time.Now()
                result.Duration = result.EndTime.Sub(result.StartTime)
                return result, nil
            }
            
        case <-timeoutCtx.Done():
            result.Decision = DecisionTimeout
            result.EndTime = time.Now()
            result.Duration = result.EndTime.Sub(result.StartTime)
            return result, fmt.Errorf("consensus timeout")
        }
    }
}

func (p *MajorityConsensusProtocol) ValidateProposal(proposal Proposal) error {
    if proposal.ID == "" {
        return fmt.Errorf("proposal ID cannot be empty")
    }
    
    if proposal.ProposerID == "" {
        return fmt.Errorf("proposer ID cannot be empty")
    }
    
    if proposal.Content == nil {
        return fmt.Errorf("proposal content cannot be nil")
    }
    
    // Check expiration
    if proposal.ExpiresAt != nil && time.Now().After(*proposal.ExpiresAt) {
        return fmt.Errorf("proposal has expired")
    }
    
    return nil
}

func (p *MajorityConsensusProtocol) GetRequiredVotes(totalParticipants int) int {
    return int(math.Ceil(float64(totalParticipants) * p.threshold))
}

// Byzantine Fault Tolerant (BFT) consensus protocol
type BFTConsensusProtocol struct {
    maxFaults int
    timeout   time.Duration
    rounds    int
}

func NewBFTConsensusProtocol(maxFaults int, timeout time.Duration) *BFTConsensusProtocol {
    return &BFTConsensusProtocol{
        maxFaults: maxFaults,
        timeout:   timeout,
        rounds:    3, // Typical for BFT
    }
}

func (p *BFTConsensusProtocol) GetName() string {
    return "bft"
}

func (p *BFTConsensusProtocol) ReachConsensus(ctx context.Context, proposal Proposal, participants []string) (*ConsensusResult, error) {
    if len(participants) <= 3*p.maxFaults {
        return nil, fmt.Errorf("insufficient participants for BFT consensus: need > 3f, have %d", len(participants))
    }
    
    result := &ConsensusResult{
        ProposalID:   proposal.ID,
        Participants: participants,
        StartTime:    time.Now(),
        Votes:        make([]Ballot, 0),
    }
    
    // BFT consensus implementation (simplified)
    // In practice, this would involve multiple rounds of prepare, commit phases
    
    requiredVotes := len(participants) - p.maxFaults
    
    // For simplicity, we'll implement a basic version
    // Real BFT would require cryptographic signatures and multiple rounds
    
    timeoutCtx, cancel := context.WithTimeout(ctx, p.timeout)
    defer cancel()
    
    votesChan := make(chan Ballot, len(participants))
    
    votesReceived := 0
    validVotes := 0
    
    for {
        select {
        case vote := <-votesChan:
            result.Votes = append(result.Votes, vote)
            votesReceived++
            
            // In real BFT, we would verify cryptographic signatures
            if p.validateVote(vote) {
                validVotes++
            }
            
            if validVotes >= requiredVotes {
                result.Decision = DecisionAccepted
                result.EndTime = time.Now()
                result.Duration = result.EndTime.Sub(result.StartTime)
                return result, nil
            }
            
            if votesReceived >= len(participants) {
                if validVotes >= requiredVotes {
                    result.Decision = DecisionAccepted
                } else {
                    result.Decision = DecisionRejected
                }
                result.EndTime = time.Now()
                result.Duration = result.EndTime.Sub(result.StartTime)
                return result, nil
            }
            
        case <-timeoutCtx.Done():
            result.Decision = DecisionTimeout
            result.EndTime = time.Now()
            result.Duration = result.EndTime.Sub(result.StartTime)
            return result, fmt.Errorf("BFT consensus timeout")
        }
    }
}

func (p *BFTConsensusProtocol) validateVote(vote Ballot) bool {
    // In real implementation, this would verify cryptographic signatures
    // and check for Byzantine behavior
    return vote.Vote != "" && vote.VoterID != ""
}

func (p *BFTConsensusProtocol) ValidateProposal(proposal Proposal) error {
    // BFT-specific validation
    if proposal.ID == "" {
        return fmt.Errorf("proposal ID cannot be empty")
    }
    
    // Would include cryptographic signature validation
    return nil
}

func (p *BFTConsensusProtocol) GetRequiredVotes(totalParticipants int) int {
    return totalParticipants - p.maxFaults
}

// Consensus engine implementation
type DefaultConsensusEngine struct {
    protocol     ConsensusProtocol
    participants map[string]bool
    proposals    map[string]*ConsensusResult
    mu           sync.RWMutex
    logger       *zap.Logger
}

func NewDefaultConsensusEngine(protocol ConsensusProtocol) *DefaultConsensusEngine {
    return &DefaultConsensusEngine{
        protocol:     protocol,
        participants: make(map[string]bool),
        proposals:    make(map[string]*ConsensusResult),
        logger:       zap.NewNop(),
    }
}

func (e *DefaultConsensusEngine) ProposeValue(ctx context.Context, proposal Proposal) (*ConsensusResult, error) {
    // Validate proposal
    if err := e.protocol.ValidateProposal(proposal); err != nil {
        return nil, fmt.Errorf("invalid proposal: %w", err)
    }
    
    e.mu.RLock()
    participants := make([]string, 0, len(e.participants))
    for agentID := range e.participants {
        participants = append(participants, agentID)
    }
    e.mu.RUnlock()
    
    if len(participants) == 0 {
        return nil, fmt.Errorf("no participants available for consensus")
    }
    
    e.logger.Info("Starting consensus",
        zap.String("proposal_id", proposal.ID),
        zap.String("protocol", e.protocol.GetName()),
        zap.Int("participants", len(participants)),
    )
    
    // Reach consensus using protocol
    result, err := e.protocol.ReachConsensus(ctx, proposal, participants)
    if err != nil {
        return nil, fmt.Errorf("consensus failed: %w", err)
    }
    
    // Store result
    e.mu.Lock()
    e.proposals[proposal.ID] = result
    e.mu.Unlock()
    
    e.logger.Info("Consensus completed",
        zap.String("proposal_id", proposal.ID),
        zap.String("decision", string(result.Decision)),
        zap.Duration("duration", result.Duration),
    )
    
    return result, nil
}

func (e *DefaultConsensusEngine) AddParticipant(agentID string) error {
    e.mu.Lock()
    defer e.mu.Unlock()
    
    e.participants[agentID] = true
    
    e.logger.Info("Participant added to consensus",
        zap.String("agent_id", agentID),
        zap.Int("total_participants", len(e.participants)),
    )
    
    return nil
}

func (e *DefaultConsensusEngine) RemoveParticipant(agentID string) error {
    e.mu.Lock()
    defer e.mu.Unlock()
    
    delete(e.participants, agentID)
    
    e.logger.Info("Participant removed from consensus",
        zap.String("agent_id", agentID),
        zap.Int("total_participants", len(e.participants)),
    )
    
    return nil
}

func (e *DefaultConsensusEngine) GetParticipants() []string {
    e.mu.RLock()
    defer e.mu.RUnlock()
    
    participants := make([]string, 0, len(e.participants))
    for agentID := range e.participants {
        participants = append(participants, agentID)
    }
    
    return participants
}
```

---

## Advanced Multi-Agent Patterns

### Leader Election

```go
// LeaderElection handles leader selection in multi-agent systems
type LeaderElection interface {
    ElectLeader(ctx context.Context, candidates []string) (*LeaderResult, error)
    GetCurrentLeader() (string, error)
    IsLeader(agentID string) bool
    AbdicateLeadership(agentID string) error
    StartElection(ctx context.Context) error
}

type LeaderResult struct {
    LeaderID    string    `json:"leader_id"`
    Term        int       `json:"term"`
    Votes       []Vote    `json:"votes"`
    ElectedAt   time.Time `json:"elected_at"`
    ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type Vote struct {
    VoterID     string    `json:"voter_id"`
    CandidateID string    `json:"candidate_id"`
    Timestamp   time.Time `json:"timestamp"`
}

// Raft-based leader election
type RaftLeaderElection struct {
    currentTerm   int
    currentLeader string
    votedFor      string
    log           []LogEntry
    commitIndex   int
    lastApplied   int
    
    // Raft state
    state         RaftState
    electionTimer *time.Timer
    heartbeatTimer *time.Timer
    
    // Participants
    peers         map[string]RaftPeer
    
    // Communication
    messageChan   chan RaftMessage
    
    // Configuration
    electionTimeout  time.Duration
    heartbeatTimeout time.Duration
    
    mu     sync.RWMutex
    logger *zap.Logger
}

type RaftState string

const (
    RaftStateFollower  RaftState = "follower"
    RaftStateCandidate RaftState = "candidate"
    RaftStateLeader    RaftState = "leader"
)

type LogEntry struct {
    Term    int         `json:"term"`
    Index   int         `json:"index"`
    Command interface{} `json:"command"`
}

type RaftPeer struct {
    ID           string `json:"id"`
    Address      string `json:"address"`
    NextIndex    int    `json:"next_index"`
    MatchIndex   int    `json:"match_index"`
    LastContact  time.Time `json:"last_contact"`
}

type RaftMessage struct {
    Type        RaftMessageType `json:"type"`
    From        string         `json:"from"`
    To          string         `json:"to"`
    Term        int            `json:"term"`
    CandidateID string         `json:"candidate_id,omitempty"`
    LastLogIndex int           `json:"last_log_index,omitempty"`
    LastLogTerm  int           `json:"last_log_term,omitempty"`
    VoteGranted  bool          `json:"vote_granted,omitempty"`
    Entries      []LogEntry    `json:"entries,omitempty"`
    LeaderCommit int           `json:"leader_commit,omitempty"`
    Success      bool          `json:"success,omitempty"`
}

type RaftMessageType string

const (
    RaftMessageRequestVote     RaftMessageType = "request_vote"
    RaftMessageVoteResponse    RaftMessageType = "vote_response"
    RaftMessageAppendEntries   RaftMessageType = "append_entries"
    RaftMessageAppendResponse  RaftMessageType = "append_response"
    RaftMessageHeartbeat       RaftMessageType = "heartbeat"
)

func NewRaftLeaderElection(nodeID string, peers []string) *RaftLeaderElection {
    raft := &RaftLeaderElection{
        currentTerm:      0,
        currentLeader:    "",
        votedFor:         "",
        log:              make([]LogEntry, 0),
        commitIndex:      0,
        lastApplied:      0,
        state:            RaftStateFollower,
        peers:            make(map[string]RaftPeer),
        messageChan:      make(chan RaftMessage, 100),
        electionTimeout:  time.Duration(150+rand.Intn(150)) * time.Millisecond,
        heartbeatTimeout: 50 * time.Millisecond,
        logger:           zap.NewNop(),
    }
    
    // Initialize peers
    for _, peerID := range peers {
        if peerID != nodeID {
            raft.peers[peerID] = RaftPeer{
                ID:         peerID,
                NextIndex:  1,
                MatchIndex: 0,
            }
        }
    }
    
    return raft
}

func (r *RaftLeaderElection) StartElection(ctx context.Context) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Become candidate
    r.state = RaftStateCandidate
    r.currentTerm++
    r.votedFor = "" // Vote for self
    
    votesReceived := 1 // Vote for self
    
    r.logger.Info("Starting leader election",
        zap.Int("term", r.currentTerm),
        zap.Int("peers", len(r.peers)),
    )
    
    // Send RequestVote to all peers
    for peerID := range r.peers {
        go r.sendRequestVote(peerID)
    }
    
    // Wait for votes or timeout
    electionCtx, cancel := context.WithTimeout(ctx, r.electionTimeout)
    defer cancel()
    
    for {
        select {
        case message := <-r.messageChan:
            if message.Type == RaftMessageVoteResponse && message.Term == r.currentTerm {
                if message.VoteGranted {
                    votesReceived++
                    
                    // Check if we have majority
                    if votesReceived > len(r.peers)/2 {
                        r.becomeLeader()
                        return nil
                    }
                }
            }
            
        case <-electionCtx.Done():
            // Election timeout, become follower
            r.state = RaftStateFollower
            return fmt.Errorf("election timeout")
        }
    }
}

func (r *RaftLeaderElection) sendRequestVote(peerID string) {
    r.mu.RLock()
    lastLogIndex := len(r.log) - 1
    lastLogTerm := 0
    if lastLogIndex >= 0 {
        lastLogTerm = r.log[lastLogIndex].Term
    }
    r.mu.RUnlock()
    
    message := RaftMessage{
        Type:         RaftMessageRequestVote,
        From:         "", // Would be set to current node ID
        To:           peerID,
        Term:         r.currentTerm,
        CandidateID:  "", // Current node ID
        LastLogIndex: lastLogIndex,
        LastLogTerm:  lastLogTerm,
    }
    
    // Send message (implementation depends on communication layer)
    r.sendMessage(message)
}

func (r *RaftLeaderElection) becomeLeader() {
    r.state = RaftStateLeader
    r.currentLeader = "" // Current node ID
    
    r.logger.Info("Became leader",
        zap.Int("term", r.currentTerm),
    )
    
    // Initialize leader state
    for peerID := range r.peers {
        peer := r.peers[peerID]
        peer.NextIndex = len(r.log)
        peer.MatchIndex = 0
        r.peers[peerID] = peer
    }
    
    // Start sending heartbeats
    go r.sendHeartbeats()
}

func (r *RaftLeaderElection) sendHeartbeats() {
    ticker := time.NewTicker(r.heartbeatTimeout)
    defer ticker.Stop()
    
    for range ticker.C {
        r.mu.RLock()
        if r.state != RaftStateLeader {
            r.mu.RUnlock()
            return
        }
        r.mu.RUnlock()
        
        // Send heartbeat to all peers
        for peerID := range r.peers {
            go r.sendAppendEntries(peerID, true)
        }
    }
}

func (r *RaftLeaderElection) sendAppendEntries(peerID string, heartbeat bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    peer := r.peers[peerID]
    prevLogIndex := peer.NextIndex - 1
    prevLogTerm := 0
    
    if prevLogIndex >= 0 && prevLogIndex < len(r.log) {
        prevLogTerm = r.log[prevLogIndex].Term
    }
    
    var entries []LogEntry
    if !heartbeat && peer.NextIndex < len(r.log) {
        entries = r.log[peer.NextIndex:]
    }
    
    message := RaftMessage{
        Type:         RaftMessageAppendEntries,
        From:         "", // Current node ID
        To:           peerID,
        Term:         r.currentTerm,
        Entries:      entries,
        LeaderCommit: r.commitIndex,
    }
    
    r.sendMessage(message)
}

func (r *RaftLeaderElection) sendMessage(message RaftMessage) {
    // Implementation depends on communication layer
    // This would typically send over network
    select {
    case r.messageChan <- message:
    default:
        r.logger.Warn("Message channel full, dropping message")
    }
}

func (r *RaftLeaderElection) GetCurrentLeader() (string, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    if r.currentLeader == "" {
        return "", fmt.Errorf("no current leader")
    }
    
    return r.currentLeader, nil
}

func (r *RaftLeaderElection) IsLeader(agentID string) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    return r.currentLeader == agentID && r.state == RaftStateLeader
}
```

### Agent Swarm Intelligence

```go
// SwarmIntelligence implements collective behavior patterns
type SwarmIntelligence interface {
    // Swarm behavior
    FormSwarm(agents []Agent, objective Objective) (*Swarm, error)
    DissolveSwarm(swarmID string) error
    
    // Collective decision making
    CollectiveDecision(ctx context.Context, swarmID string, problem Problem) (*CollectiveResult, error)
    
    // Emergent behavior
    ObserveEmergentBehavior(swarmID string) (*BehaviorAnalysis, error)
    
    // Swarm optimization
    OptimizeSwarm(swarmID string, criteria OptimizationCriteria) error
}

type Swarm struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Agents      []Agent           `json:"-"`
    Objective   Objective         `json:"objective"`
    Topology    SwarmTopology     `json:"topology"`
    Behavior    SwarmBehavior     `json:"behavior"`
    Metrics     SwarmMetrics      `json:"metrics"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

type Objective struct {
    Type        ObjectiveType      `json:"type"`
    Description string            `json:"description"`
    Target      interface{}       `json:"target"`
    Constraints []Constraint      `json:"constraints"`
    Fitness     FitnessFunction   `json:"-"`
}

type ObjectiveType string

const (
    ObjectiveOptimization ObjectiveType = "optimization"
    ObjectiveExploration  ObjectiveType = "exploration"
    ObjectiveClassification ObjectiveType = "classification"
    ObjectivePrediction   ObjectiveType = "prediction"
    ObjectiveCoordination ObjectiveType = "coordination"
)

type FitnessFunction func(solution interface{}) float64

type SwarmTopology string

const (
    TopologyFullyConnected SwarmTopology = "fully_connected"
    TopologyRing          SwarmTopology = "ring"
    TopologyGrid          SwarmTopology = "grid"
    TopologyHierarchical  SwarmTopology = "hierarchical"
    TopologyRandomGraph   SwarmTopology = "random_graph"
)

// Particle Swarm Optimization (PSO) implementation
type ParticleSwarmOptimizer struct {
    particles       []Particle
    globalBest      Solution
    inertiaWeight   float64
    cognitiveWeight float64
    socialWeight    float64
    maxIterations   int
    tolerance       float64
    logger          *zap.Logger
}

type Particle struct {
    ID           string    `json:"id"`
    Position     []float64 `json:"position"`
    Velocity     []float64 `json:"velocity"`
    PersonalBest Solution  `json:"personal_best"`
    Fitness      float64   `json:"fitness"`
}

type Solution struct {
    Values  []float64 `json:"values"`
    Fitness float64   `json:"fitness"`
}

func NewParticleSwarmOptimizer(numParticles, dimensions int) *ParticleSwarmOptimizer {
    particles := make([]Particle, numParticles)
    
    for i := 0; i < numParticles; i++ {
        particles[i] = Particle{
            ID:       fmt.Sprintf("particle_%d", i),
            Position: make([]float64, dimensions),
            Velocity: make([]float64, dimensions),
            PersonalBest: Solution{
                Values:  make([]float64, dimensions),
                Fitness: math.Inf(-1),
            },
            Fitness: math.Inf(-1),
        }
        
        // Random initialization
        for j := 0; j < dimensions; j++ {
            particles[i].Position[j] = rand.Float64()*2 - 1 // [-1, 1]
            particles[i].Velocity[j] = rand.Float64()*0.1 - 0.05 // [-0.05, 0.05]
        }
    }
    
    return &ParticleSwarmOptimizer{
        particles:       particles,
        globalBest:      Solution{Values: make([]float64, dimensions), Fitness: math.Inf(-1)},
        inertiaWeight:   0.729,
        cognitiveWeight: 1.494,
        socialWeight:    1.494,
        maxIterations:   1000,
        tolerance:       1e-6,
        logger:          zap.NewNop(),
    }
}

func (pso *ParticleSwarmOptimizer) Optimize(ctx context.Context, fitnessFunc FitnessFunction) (*Solution, error) {
    for iteration := 0; iteration < pso.maxIterations; iteration++ {
        select {
        case <-ctx.Done():
            return &pso.globalBest, ctx.Err()
        default:
        }
        
        // Evaluate particles
        for i := range pso.particles {
            fitness := fitnessFunc(pso.particles[i].Position)
            pso.particles[i].Fitness = fitness
            
            // Update personal best
            if fitness > pso.particles[i].PersonalBest.Fitness {
                pso.particles[i].PersonalBest.Fitness = fitness
                copy(pso.particles[i].PersonalBest.Values, pso.particles[i].Position)
            }
            
            // Update global best
            if fitness > pso.globalBest.Fitness {
                pso.globalBest.Fitness = fitness
                pso.globalBest.Values = make([]float64, len(pso.particles[i].Position))
                copy(pso.globalBest.Values, pso.particles[i].Position)
            }
        }
        
        // Update velocities and positions
        for i := range pso.particles {
            pso.updateParticle(&pso.particles[i])
        }
        
        // Check convergence
        if pso.checkConvergence() {
            pso.logger.Info("PSO converged",
                zap.Int("iteration", iteration),
                zap.Float64("fitness", pso.globalBest.Fitness),
            )
            break
        }
        
        if iteration%100 == 0 {
            pso.logger.Debug("PSO progress",
                zap.Int("iteration", iteration),
                zap.Float64("best_fitness", pso.globalBest.Fitness),
            )
        }
    }
    
    return &pso.globalBest, nil
}

func (pso *ParticleSwarmOptimizer) updateParticle(particle *Particle) {
    for j := range particle.Velocity {
        r1, r2 := rand.Float64(), rand.Float64()
        
        // Update velocity
        particle.Velocity[j] = pso.inertiaWeight*particle.Velocity[j] +
            pso.cognitiveWeight*r1*(particle.PersonalBest.Values[j]-particle.Position[j]) +
            pso.socialWeight*r2*(pso.globalBest.Values[j]-particle.Position[j])
        
        // Update position
        particle.Position[j] += particle.Velocity[j]
        
        // Apply bounds if necessary
        if particle.Position[j] > 1 {
            particle.Position[j] = 1
        } else if particle.Position[j] < -1 {
            particle.Position[j] = -1
        }
    }
}

func (pso *ParticleSwarmOptimizer) checkConvergence() bool {
    // Simple convergence check based on global best improvement
    return false // Implement based on specific criteria
}

// Ant Colony Optimization (ACO) for pathfinding
type AntColonyOptimizer struct {
    ants            []Ant
    pheromoneMatrix [][]float64
    distanceMatrix  [][]float64
    alpha           float64 // Pheromone influence
    beta            float64 // Distance influence
    evaporationRate float64
    pheromoneDeposit float64
    numCities       int
    maxIterations   int
    logger          *zap.Logger
}

type Ant struct {
    ID       string  `json:"id"`
    Path     []int   `json:"path"`
    Visited  []bool  `json:"visited"`
    Distance float64 `json:"distance"`
    Current  int     `json:"current"`
}

func NewAntColonyOptimizer(distanceMatrix [][]float64) *AntColonyOptimizer {
    numCities := len(distanceMatrix)
    numAnts := numCities
    
    // Initialize pheromone matrix
    pheromoneMatrix := make([][]float64, numCities)
    for i := range pheromoneMatrix {
        pheromoneMatrix[i] = make([]float64, numCities)
        for j := range pheromoneMatrix[i] {
            pheromoneMatrix[i][j] = 1.0 // Initial pheromone level
        }
    }
    
    // Initialize ants
    ants := make([]Ant, numAnts)
    for i := range ants {
        ants[i] = Ant{
            ID:      fmt.Sprintf("ant_%d", i),
            Path:    make([]int, 0),
            Visited: make([]bool, numCities),
            Current: rand.Intn(numCities),
        }
        ants[i].Path = append(ants[i].Path, ants[i].Current)
        ants[i].Visited[ants[i].Current] = true
    }
    
    return &AntColonyOptimizer{
        ants:            ants,
        pheromoneMatrix: pheromoneMatrix,
        distanceMatrix:  distanceMatrix,
        alpha:           1.0,
        beta:            2.0,
        evaporationRate: 0.5,
        pheromoneDeposit: 1.0,
        numCities:       numCities,
        maxIterations:   1000,
        logger:          zap.NewNop(),
    }
}

func (aco *AntColonyOptimizer) Optimize(ctx context.Context) (*Solution, error) {
    bestPath := make([]int, 0)
    bestDistance := math.Inf(1)
    
    for iteration := 0; iteration < aco.maxIterations; iteration++ {
        select {
        case <-ctx.Done():
            return &Solution{
                Values:  intSliceToFloat64(bestPath),
                Fitness: -bestDistance, // Negative because we want to minimize distance
            }, ctx.Err()
        default:
        }
        
        // Reset ants
        for i := range aco.ants {
            aco.resetAnt(&aco.ants[i])
        }
        
        // Each ant constructs a solution
        for step := 0; step < aco.numCities-1; step++ {
            for i := range aco.ants {
                nextCity := aco.selectNextCity(&aco.ants[i])
                aco.moveAnt(&aco.ants[i], nextCity)
            }
        }
        
        // Complete tours and evaluate
        for i := range aco.ants {
            aco.completeTour(&aco.ants[i])
            if aco.ants[i].Distance < bestDistance {
                bestDistance = aco.ants[i].Distance
                bestPath = make([]int, len(aco.ants[i].Path))
                copy(bestPath, aco.ants[i].Path)
            }
        }
        
        // Update pheromones
        aco.updatePheromones()
        
        if iteration%100 == 0 {
            aco.logger.Debug("ACO progress",
                zap.Int("iteration", iteration),
                zap.Float64("best_distance", bestDistance),
            )
        }
    }
    
    return &Solution{
        Values:  intSliceToFloat64(bestPath),
        Fitness: -bestDistance,
    }, nil
}

func (aco *AntColonyOptimizer) selectNextCity(ant *Ant) int {
    probabilities := make([]float64, aco.numCities)
    sum := 0.0
    
    // Calculate probabilities
    for city := 0; city < aco.numCities; city++ {
        if !ant.Visited[city] {
            pheromone := math.Pow(aco.pheromoneMatrix[ant.Current][city], aco.alpha)
            distance := math.Pow(1.0/aco.distanceMatrix[ant.Current][city], aco.beta)
            probabilities[city] = pheromone * distance
            sum += probabilities[city]
        }
    }
    
    // Normalize probabilities
    for city := range probabilities {
        probabilities[city] /= sum
    }
    
    // Roulette wheel selection
    r := rand.Float64()
    cumulative := 0.0
    
    for city := 0; city < aco.numCities; city++ {
        if !ant.Visited[city] {
            cumulative += probabilities[city]
            if r <= cumulative {
                return city
            }
        }
    }
    
    // Fallback to first unvisited city
    for city := 0; city < aco.numCities; city++ {
        if !ant.Visited[city] {
            return city
        }
    }
    
    return -1 // Should not happen
}

func (aco *AntColonyOptimizer) moveAnt(ant *Ant, nextCity int) {
    ant.Distance += aco.distanceMatrix[ant.Current][nextCity]
    ant.Current = nextCity
    ant.Visited[nextCity] = true
    ant.Path = append(ant.Path, nextCity)
}

func (aco *AntColonyOptimizer) completeTour(ant *Ant) {
    // Return to starting city
    startCity := ant.Path[0]
    ant.Distance += aco.distanceMatrix[ant.Current][startCity]
    ant.Path = append(ant.Path, startCity)
}

func (aco *AntColonyOptimizer) resetAnt(ant *Ant) {
    ant.Path = make([]int, 0)
    ant.Visited = make([]bool, aco.numCities)
    ant.Distance = 0.0
    ant.Current = rand.Intn(aco.numCities)
    ant.Path = append(ant.Path, ant.Current)
    ant.Visited[ant.Current] = true
}

func (aco *AntColonyOptimizer) updatePheromones() {
    // Evaporation
    for i := range aco.pheromoneMatrix {
        for j := range aco.pheromoneMatrix[i] {
            aco.pheromoneMatrix[i][j] *= (1.0 - aco.evaporationRate)
        }
    }
    
    // Deposit pheromones
    for _, ant := range aco.ants {
        deposit := aco.pheromoneDeposit / ant.Distance
        
        for i := 0; i < len(ant.Path)-1; i++ {
            from, to := ant.Path[i], ant.Path[i+1]
            aco.pheromoneMatrix[from][to] += deposit
            aco.pheromoneMatrix[to][from] += deposit // Symmetric
        }
    }
}

func intSliceToFloat64(ints []int) []float64 {
    floats := make([]float64, len(ints))
    for i, v := range ints {
        floats[i] = float64(v)
    }
    return floats
}
```

---

## Best Practices

### 1. System Design
- Design for scalability from the start
- Implement proper resource management
- Use appropriate communication patterns
- Plan for failure scenarios
- Monitor system health continuously

### 2. Coordination
- Choose appropriate coordination strategies
- Implement timeout and retry logic
- Design for partial failures
- Use consensus when necessary
- Avoid central points of failure

### 3. Communication
- Use asynchronous messaging when possible
- Implement proper message routing
- Handle message ordering requirements
- Plan for network partitions
- Implement backpressure mechanisms

### 4. Consensus
- Choose appropriate consensus protocols
- Handle Byzantine failures if needed
- Implement proper timeout handling
- Plan for split-brain scenarios
- Monitor consensus performance

### 5. Performance
- Optimize for the common case
- Use connection pooling
- Implement proper caching
- Monitor resource usage
- Plan for horizontal scaling

---

## Next Steps

- **[State Management](state-management.md)** - Agent state and data flow
- **[LLM Agents](llm-agents.md)** - AI-powered agents with tool support
- **[Workflow Agents](workflow-agents.md)** - Sequential, parallel, and conditional patterns
- **[Agent Overview](overview.md)** - Agent architecture and concepts
- **[Agent API Reference](../../technical/api-reference/agents.md)** - Detailed API documentation