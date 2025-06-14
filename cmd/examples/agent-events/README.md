# Agent Events Example

This example demonstrates the enhanced event system for agent monitoring, filtering, serialization, and bridge integration.

## Overview

The event system provides:
- **EventBus** for decoupled event distribution
- **Pattern-based subscriptions** (e.g., "tool.*", "agent.*")
- **Advanced filtering** with composite filters
- **Event serialization** for bridge layer integration
- **Event storage and replay** for debugging
- **Bridge-specific events** for go-llmspell integration

## Running the Example

```bash
go run main.go
```

## Features Demonstrated

### 1. Basic Event Subscription
Shows how to subscribe to all events and handle them:
```go
handler := events.EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
    fmt.Printf("Received event: Type=%s, Agent=%s\n", event.Type, event.AgentName)
    return nil
})
subID := bus.Subscribe(handler)
```

### 2. Pattern-based Subscriptions
Subscribe to events matching specific patterns:
```go
// Subscribe to all tool events
subID, err := bus.SubscribePattern("tool.*", handler)

// Other pattern examples:
// "agent.*"     - All agent events
// "workflow.*"  - All workflow events
// "*"           - All events
```

### 3. Event Filtering
Use filters to receive only specific events:
```go
// Composite filter: (tool events OR error events) AND from specific agent
filter := events.AND(
    events.OR(
        events.NewTypeFilter(domain.EventToolCall, domain.EventToolResult),
        events.NewErrorFilter(),
    ),
    events.NewAgentFilter("agent1", ""),
)
```

### 4. Event Serialization
Convert events to bridge-friendly formats:
```go
// Serialize to map for bridge layer
serialized, err := events.SerializeEvent(event)

// Multiple serialization formats
serializers := []string{"json", "json-pretty", "compact"}
for _, format := range serializers {
    serializer := events.GetSerializer(format)
    data, err := serializer.Serialize(event)
}
```

### 5. Event Storage and Replay
Record and replay events for debugging:
```go
// Create storage and recorder
storage := events.NewMemoryStorage()
recorder := events.NewEventRecorder(storage, bus)
recorder.Start()

// ... events occur ...

// Replay at 2x speed
replayer := events.NewEventReplayer(storage, replayBus)
err = replayer.Replay(ctx, query, events.ReplayOptions{
    Speed: 2.0,
})
```

### 6. Bridge Integration
Special event types for scripting engine integration:
```go
// Create bridge event publisher
bridgePublisher := events.NewBridgeEventPublisher(bus, "bridge-001", "session-123")

// Publish bridge events
requestID := bridgePublisher.PublishRequest("executeScript", map[string]interface{}{
    "language": "javascript",
    "code":     "console.log('Hello from bridge');",
})

// Handle bridge events
bridgeHandler := events.BridgeEventHandlerFunc(func(ctx context.Context, event *events.BridgeEvent) error {
    switch events.BridgeEventType(event.Type) {
    case events.BridgeEventRequest:
        // Handle request
    case events.BridgeEventScriptStart:
        // Handle script execution
    }
    return nil
})
```

## Event Types

The system supports all standard domain events:
- `agent.start`, `agent.complete`, `agent.error`
- `tool.call`, `tool.result`, `tool.error`
- `workflow.start`, `workflow.step`, `workflow.complete`
- `progress` events with current/total tracking

And bridge-specific events:
- `bridge.connected`, `bridge.disconnected`
- `bridge.request`, `bridge.response`
- `bridge.script.start`, `bridge.script.complete`
- `bridge.convert`, `bridge.error`

## Key Components

- **EventBus**: Central hub for event distribution
- **EventHandler**: Interface for processing events
- **EventFilter**: Interface for filtering events
- **EventStorage**: Interface for persisting events
- **EventSerializer**: Interface for converting events to different formats
- **BridgeEvent**: Special events for scripting engine integration

## Use Cases

1. **Monitoring**: Track agent execution in real-time
2. **Debugging**: Record and replay event sequences
3. **Integration**: Bridge events to external systems
4. **Analytics**: Filter and analyze specific event patterns
5. **Testing**: Use event replay for test scenarios

## Related Examples

- `agent-state-persistence`: Shows state management with events
- `agent-workflow-hooks`: Demonstrates workflow event hooks
- `agent-metrics-tools`: Uses events for metrics collection