# Event System

> **[Documentation Home](../README.md) / [Advanced Topics](README.md) / Event System**

## Overview

The event system in go-llms provides comprehensive observability into agent operations, tool executions, and state changes. It enables monitoring, debugging, auditing, and building reactive systems on top of agent workflows.

## Prerequisites

Before working with the event system, ensure you have:

- **Go Concurrency**: Understanding of channels, goroutines, and synchronization primitives
- **Event-Driven Architecture**: Familiarity with pub/sub patterns, event sourcing concepts
- **Observability**: Knowledge of logging, metrics, and distributed tracing
- **Performance Optimization**: Understanding of buffering, batching, and backpressure
- **Serialization**: Experience with JSON/Protocol Buffers and custom serialization

**Required Reading**:
- [Agents Overview](../../technical/agents/overview.md) - Agent system that generates events
- [Performance](performance.md) - Performance considerations for high-volume events
- [Core Concepts](../../technical/core-concepts.md) - Fundamental Go-LLMs concepts

## Architecture

### Event Flow

![Event System Flow](../images/event-system-flow.svg)
*Figure 1: Event system architecture showing the complete flow from event creation through processing to storage and replay*

### Core Components

#### Event Structure
```go
type Event struct {
    ID        string                 // Unique event identifier
    Type      EventType             // Event type (enum)
    Timestamp time.Time             // When event occurred
    Source    string                // Event source (agent ID)
    AgentName string                // Human-readable agent name
    Data      interface{}           // Event-specific data
    Metadata  map[string]interface{} // Additional context
}
```

#### Event Types
```go
const (
    // Agent lifecycle events
    EventAgentStart     EventType = "agent.start"
    EventAgentComplete  EventType = "agent.complete"
    EventAgentError     EventType = "agent.error"
    
    // Tool events
    EventToolCall       EventType = "tool.call"
    EventToolResponse   EventType = "tool.response"
    EventToolError      EventType = "tool.error"
    
    // State events
    EventStateChange    EventType = "state.change"
    EventStateSnapshot  EventType = "state.snapshot"
    
    // Workflow events
    EventWorkflowStart  EventType = "workflow.start"
    EventWorkflowStep   EventType = "workflow.step"
    EventWorkflowEnd    EventType = "workflow.end"
    
    // Custom events
    EventCustom         EventType = "custom"
)
```

## Event Handling

### Basic Event Subscription
```go
agent := core.NewLLMAgent("assistant", "gpt-4", deps)

// Subscribe to all events
agent.OnEvent(func(event domain.Event) {
    log.Printf("Event: %s - %s", event.Type, event.Data)
}

// Subscribe to specific event types
agent.OnEventType(domain.EventToolCall, func(event domain.Event) {
    toolCall := event.Data.(ToolCallData)
    log.Printf("Tool called: %s with params: %v", 
        toolCall.Name, toolCall.Parameters)
}
```

### Event Filters
```go
// Filter by event type
typeFilter := filters.NewTypeFilter(
    domain.EventToolCall,
    domain.EventToolResponse,
)

// Filter by agent
agentFilter := filters.NewAgentFilter("research-agent")

// Filter by time range
timeFilter := filters.NewTimeRangeFilter(
    time.Now().Add(-1*time.Hour),
    time.Now(),
)

// Combine filters
combinedFilter := filters.And(typeFilter, agentFilter, timeFilter)

// Apply filter to event stream
filteredEvents := filter.Apply(eventStream, combinedFilter)
```

### Event Handlers
```go
// Logging handler
loggingHandler := handlers.NewLoggingHandler(logger)

// Metrics handler
metricsHandler := handlers.NewMetricsHandler(metricsCollector)

// Storage handler
storageHandler := handlers.NewStorageHandler(eventStore)

// Chain handlers
chainedHandler := handlers.Chain(
    loggingHandler,
    metricsHandler,
    storageHandler,
)

// Attach to agent
agent.OnEvent(chainedHandler.Handle)
```

## Event Emitter

### Basic Emitter
```go
// Create event emitter
emitter := events.NewEventEmitter()

// Subscribe to events
unsubscribe := emitter.On(func(event domain.Event) {
    fmt.Printf("Received: %v\n", event)
}

// Emit event
emitter.Emit(domain.Event{
    Type: domain.EventCustom,
    Data: "Custom event data",
}

// Unsubscribe
unsubscribe()
```

### Buffered Emitter
```go
// Create buffered emitter for high-throughput scenarios
emitter := events.NewBufferedEmitter(
    events.WithBufferSize(1000),
    events.WithFlushInterval(100*time.Millisecond),
    events.WithOverflowStrategy(events.DropOldest),
)

// Handle events in batches
emitter.OnBatch(func(events []domain.Event) {
    // Process batch of events
    processBatch(events)
}
```

## Event Serialization

### JSON Serialization
```go
// Serialize event to JSON
event := domain.Event{
    Type: domain.EventToolCall,
    Data: map[string]interface{}{
        "tool": "calculator",
        "params": map[string]interface{}{
            "operation": "multiply",
            "a": 10,
            "b": 20,
        },
    },
}

jsonData, err := json.Marshal(event)

// Deserialize
var decoded domain.Event
err = json.Unmarshal(jsonData, &decoded)
```

### Custom Serialization
```go
// Register custom type serializer
events.RegisterSerializer(reflect.TypeOf(MyCustomType{}), 
    func(v interface{}) ([]byte, error) {
        custom := v.(MyCustomType)
        return custom.Serialize()
    },
)

// Register deserializer
events.RegisterDeserializer("MyCustomType",
    func(data []byte) (interface{}, error) {
        var custom MyCustomType
        err := custom.Deserialize(data)
        return custom, err
    },
)
```

## Event Storage

### In-Memory Storage
```go
// Create in-memory event store
store := storage.NewMemoryEventStore(
    storage.WithMaxEvents(10000),
    storage.WithTTL(24*time.Hour),
)

// Store events
store.Store(event)

// Query events
events, err := store.Query(
    storage.WithType(domain.EventToolCall),
    storage.WithTimeRange(start, end),
    storage.WithAgent("assistant"),
)
```

### Persistent Storage
```go
// File-based storage
fileStore := storage.NewFileEventStore("events.jsonl")

// Database storage
dbStore := storage.NewDatabaseEventStore(db,
    storage.WithTable("agent_events"),
    storage.WithPartitioning(storage.PartitionByDay),
)

// S3 storage for long-term retention
s3Store := storage.NewS3EventStore(s3Client,
    storage.WithBucket("agent-events"),
    storage.WithCompression(true),
)
```

## Event Replay

### Basic Replay
```go
// Create event replayer
replayer := replay.NewEventReplayer(eventStore)

// Replay events from a time range
err := replayer.Replay(
    context.Background(),
    time.Now().Add(-1*time.Hour),
    time.Now(),
    func(event domain.Event) error {
        // Process replayed event
        return processEvent(event)
    },
)
```

### Workflow Replay
```go
// Replay specific workflow execution
workflowReplayer := replay.NewWorkflowReplayer(eventStore)

events, err := workflowReplayer.ReplayWorkflow(
    workflowID,
    replay.WithSpeed(2.0), // 2x speed
    replay.WithFilters(
        filters.NewTypeFilter(
            domain.EventWorkflowStep,
            domain.EventStateChange,
        ),
    ),
)

// Analyze workflow execution
analyzer := analysis.NewWorkflowAnalyzer()
report := analyzer.Analyze(events)
```

## Event Patterns

### Event Sourcing
```go
// Event-sourced agent state
type EventSourcedAgent struct {
    agent  domain.BaseAgent
    events []domain.Event
}

func (e *EventSourcedAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
    // Record start event
    e.recordEvent(domain.Event{
        Type: domain.EventAgentStart,
        Data: map[string]interface{}{
            "input_state": state.Snapshot(),
        },
}
    
    // Run agent
    result, err := e.agent.Run(ctx, state)
    
    if err != nil {
        e.recordEvent(domain.Event{
            Type: domain.EventAgentError,
            Data: map[string]interface{}{
                "error": err.Error(),
            },
}
        return nil, err
    }
    
    // Record completion
    e.recordEvent(domain.Event{
        Type: domain.EventAgentComplete,
        Data: map[string]interface{}{
            "output_state": result.Snapshot(),
        },
}
    
    return result, nil
}

// Rebuild state from events
func (e *EventSourcedAgent) RebuildState() *domain.State {
    state := domain.NewState()
    
    for _, event := range e.events {
        switch event.Type {
        case domain.EventStateChange:
            change := event.Data.(StateChange)
            state.Set(change.Key, change.Value)
        }
    }
    
    return state
}
```

### Event Aggregation
```go
// Aggregate events for analytics
type EventAggregator struct {
    store storage.EventStore
}

func (a *EventAggregator) GetToolUsageStats(timeRange time.Duration) (map[string]int, error) {
    events, err := a.store.Query(
        storage.WithType(domain.EventToolCall),
        storage.WithTimeRange(
            time.Now().Add(-timeRange),
            time.Now(),
        ),
    )
    
    usage := make(map[string]int)
    for _, event := range events {
        toolCall := event.Data.(ToolCallData)
        usage[toolCall.Name]++
    }
    
    return usage, nil
}
```

### Event-Driven Workflows
```go
// React to events in real-time
type EventDrivenWorkflow struct {
    emitter events.EventEmitter
}

func (w *EventDrivenWorkflow) Start() {
    // React to tool errors
    w.emitter.On(func(event domain.Event) {
        if event.Type == domain.EventToolError {
            w.handleToolError(event)
        }
}
    
    // Monitor performance
    w.emitter.On(func(event domain.Event) {
        if event.Type == domain.EventAgentComplete {
            w.recordPerformanceMetrics(event)
        }
}
    
    // Trigger alerts
    w.emitter.On(func(event domain.Event) {
        if event.Type == domain.EventAgentError {
            w.sendAlert(event)
        }
}
}
```

## Performance Considerations

### Event Volume
```go
// Rate limiting for high-volume events
rateLimiter := events.NewRateLimiter(
    events.WithMaxEventsPerSecond(1000),
    events.WithBurstSize(5000),
)

// Sampling for metrics
sampler := events.NewSampler(
    events.WithSampleRate(0.1), // 10% sampling
    events.WithAlwaysSample(
        domain.EventAgentError,
        domain.EventToolError,
    ),
)
```

### Memory Management
```go
// Bounded event buffer
buffer := events.NewCircularBuffer(10000)

// Event compaction
compactor := events.NewCompactor(
    events.WithCompactionInterval(5*time.Minute),
    events.WithKeepLatest(1000),
    events.WithAggregation(true),
)
```

## Monitoring and Debugging

### Event Metrics
```go
// Prometheus metrics
var (
    eventCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agent_events_total",
            Help: "Total number of agent events",
        },
        []string{"type", "agent"},
    )
    
    eventDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "agent_event_duration_seconds",
            Help: "Event processing duration",
        },
        []string{"type"},
    )
)

// Update metrics in handler
func metricsHandler(event domain.Event) {
    eventCounter.WithLabelValues(
        string(event.Type),
        event.AgentName,
    ).Inc()
    
    // Track duration for completion events
    if event.Type == domain.EventAgentComplete {
        if duration, ok := event.Metadata["duration"].(time.Duration); ok {
            eventDuration.WithLabelValues(string(event.Type)).Observe(duration.Seconds())
        }
    }
}
```

### Debug Logging
```go
// Debug event logger
debugLogger := events.NewDebugLogger(
    events.WithLogLevel(slog.LevelDebug),
    events.WithPrettyPrint(true),
    events.WithIncludeStackTrace(true),
)

// Attach to agent for debugging
if debug {
    agent.OnEvent(debugLogger.Log)
}
```

### Event Tracing
```go
// OpenTelemetry integration
type TracingEventHandler struct {
    tracer trace.Tracer
}

func (h *TracingEventHandler) Handle(event domain.Event) {
    ctx := context.Background()
    
    // Create span for event
    ctx, span := h.tracer.Start(ctx, 
        fmt.Sprintf("event.%s", event.Type),
        trace.WithAttributes(
            attribute.String("event.id", event.ID),
            attribute.String("agent.name", event.AgentName),
        ),
    )
    defer span.End()
    
    // Add event data as span attributes
    if data, err := json.Marshal(event.Data); err == nil {
        span.SetAttributes(attribute.String("event.data", string(data)))
    }
}
```

## Best Practices

### 1. Event Design
- Use consistent event types and data structures
- Include sufficient context in event data
- Keep events immutable
- Use structured data over strings

### 2. Performance
- Buffer events in high-throughput scenarios
- Use sampling for metrics events
- Implement backpressure handling
- Clean up old events regularly

### 3. Error Handling
- Never let event handling errors affect main flow
- Log event handling errors separately
- Implement circuit breakers for handlers
- Have fallback strategies

### 4. Testing
- Test event emission in unit tests
- Verify event sequences in integration tests
- Use event replay for debugging
- Mock event handlers in tests

## Examples

### Complete Event Pipeline
```go
func SetupEventPipeline(agent domain.BaseAgent) {
    // Create event store
    store := storage.NewMemoryEventStore()
    
    // Create processors
    filter := filters.NewTypeFilter(
        domain.EventToolCall,
        domain.EventAgentError,
    )
    
    enricher := processors.NewEventEnricher(
        processors.WithHostInfo(),
        processors.WithTraceID(),
    )
    
    // Create handlers
    handlers := []events.Handler{
        handlers.NewLoggingHandler(logger),
        handlers.NewMetricsHandler(metrics),
        handlers.NewStorageHandler(store),
        handlers.NewAlertHandler(alertManager),
    }
    
    // Build pipeline
    pipeline := events.NewPipeline(
        events.WithFilter(filter),
        events.WithProcessor(enricher),
        events.WithHandlers(handlers...),
        events.WithErrorHandler(func(err error) {
            log.Printf("Event pipeline error: %v", err)
        }),
    )
    
    // Attach to agent
    agent.OnEvent(pipeline.Process)
}
```

## Next Steps

- Explore [Error Handling](error-handling.md) for error event patterns
- See [Performance](performance.md) for event system optimization
- Check [Bridge Integration](bridge-integration.md) for external event systems
- Review event examples in `/examples/events/`