// ABOUTME: Example demonstrating the enhanced event system for agent monitoring
// ABOUTME: Shows event bus usage, filtering, serialization, and bridge integration

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/events"
)

func main() {
	fmt.Println("=== Enhanced Event System Example ===")

	// Create event bus
	bus := events.NewEventBus()
	defer bus.Close()

	// Example 1: Basic event subscription and publishing
	fmt.Println("\n1. Basic Event Subscription:")
	demonstrateBasicSubscription(bus)

	// Example 2: Pattern-based subscriptions
	fmt.Println("\n2. Pattern-based Subscriptions:")
	demonstratePatternSubscriptions(bus)

	// Example 3: Event filtering
	fmt.Println("\n3. Event Filtering:")
	demonstrateFiltering(bus)

	// Example 4: Event serialization for bridge layer
	fmt.Println("\n4. Event Serialization:")
	demonstrateSerialization()

	// Example 5: Event storage and replay
	fmt.Println("\n5. Event Storage and Replay:")
	demonstrateStorageAndReplay()

	// Example 6: Bridge integration
	fmt.Println("\n6. Bridge Integration:")
	demonstrateBridgeIntegration(bus)
}

func demonstrateBasicSubscription(bus *events.EventBus) {
	// Create a simple event handler
	handler := events.EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		fmt.Printf("  Received event: Type=%s, Agent=%s\n", event.Type, event.AgentName)
		return nil
	})

	// Subscribe to all events
	subID := bus.Subscribe(handler)
	defer bus.Unsubscribe(subID)

	// Publish some events
	events := []domain.Event{
		domain.NewEvent(domain.EventAgentStart, "agent1", "DemoAgent", nil),
		domain.NewEvent(domain.EventProgress, "agent1", "DemoAgent", &domain.ProgressEventData{
			Current: 5,
			Total:   10,
			Message: "Processing...",
		}),
		domain.NewEvent(domain.EventAgentComplete, "agent1", "DemoAgent", nil),
	}

	for _, event := range events {
		bus.Publish(event)
	}

	// Give handlers time to process
	time.Sleep(50 * time.Millisecond)
}

func demonstratePatternSubscriptions(bus *events.EventBus) {
	// Subscribe to all tool events
	toolHandler := events.EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		fmt.Printf("  Tool event: %s\n", event.Type)
		if data, ok := event.Data.(*domain.ToolCallEventData); ok {
			fmt.Printf("    Tool: %s, Params: %v\n", data.ToolName, data.Parameters)
		}
		return nil
	})

	subID, err := bus.SubscribePattern("tool.*", toolHandler)
	if err != nil {
		log.Printf("Failed to subscribe with pattern: %v", err)
		return
	}
	defer bus.Unsubscribe(subID)

	// Publish various events
	bus.Publish(domain.NewEvent(domain.EventToolCall, "agent1", "DemoAgent", &domain.ToolCallEventData{
		ToolName:   "calculator",
		Parameters: map[string]interface{}{"operation": "add", "a": 1, "b": 2},
		RequestID:  "req-123",
	}))

	bus.Publish(domain.NewEvent(domain.EventToolResult, "agent1", "DemoAgent", &domain.ToolResultEventData{
		ToolName:  "calculator",
		Result:    3,
		RequestID: "req-123",
		Duration:  100 * time.Millisecond,
	}))

	// This won't match the pattern
	bus.Publish(domain.NewEvent(domain.EventProgress, "agent1", "DemoAgent", nil))

	time.Sleep(50 * time.Millisecond)
}

func demonstrateFiltering(bus *events.EventBus) {
	// Create composite filter: (tool events OR error events) AND from specific agent
	filter := events.AND(
		events.OR(
			events.NewTypeFilter(domain.EventToolCall, domain.EventToolResult),
			events.NewErrorFilter(),
		),
		events.NewAgentFilter("agent1", ""),
	)

	handler := events.EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		fmt.Printf("  Filtered event: Type=%s, IsError=%v\n", event.Type, event.IsError())
		return nil
	})

	subID := bus.Subscribe(handler, filter)
	defer bus.Unsubscribe(subID)

	// Publish events - only some will match
	testEvents := []domain.Event{
		domain.NewEvent(domain.EventToolCall, "agent1", "DemoAgent", nil),   // Matches
		domain.NewEvent(domain.EventToolCall, "agent2", "OtherAgent", nil),  // No match (wrong agent)
		domain.NewEvent(domain.EventAgentError, "agent1", "DemoAgent", nil), // Matches
		domain.NewEvent(domain.EventProgress, "agent1", "DemoAgent", nil),   // No match (wrong type)
	}

	for _, event := range testEvents {
		bus.Publish(event)
	}

	time.Sleep(50 * time.Millisecond)
}

func demonstrateSerialization() {
	// Create an event with complex data
	event := domain.NewEvent(domain.EventWorkflowStep, "workflow1", "DataPipeline", &domain.WorkflowStepEventData{
		StepName:    "DataValidation",
		StepIndex:   2,
		TotalSteps:  5,
		Description: "Validating input data",
	})
	event.Metadata["environment"] = "production"
	event.Metadata["version"] = "1.2.3"

	// Serialize to bridge-friendly format
	serialized, err := events.SerializeEvent(event)
	if err != nil {
		log.Printf("Serialization failed: %v", err)
		return
	}

	fmt.Printf("  Serialized event: %v\n", serialized)

	// Deserialize back
	recovered, err := events.DeserializeEvent(serialized)
	if err != nil {
		log.Printf("Deserialization failed: %v", err)
		return
	}

	fmt.Printf("  Recovered event type: %s\n", recovered.Type)

	// Test different serializers
	fmt.Println("\n  Testing serializers:")

	serializers := []string{"json", "json-pretty", "compact"}
	for _, format := range serializers {
		serializer := events.GetSerializer(format)
		data, err := serializer.Serialize(event)
		if err != nil {
			log.Printf("  %s serialization failed: %v", format, err)
			continue
		}
		fmt.Printf("  %s format (%d bytes)\n", format, len(data))
	}
}

func demonstrateStorageAndReplay() {
	// Create storage and event bus
	storage := events.NewMemoryStorage()
	defer func() {
		if err := storage.Close(); err != nil {
			log.Printf("Failed to close storage: %v", err)
		}
	}()

	bus := events.NewEventBus()
	defer bus.Close()

	// Create recorder
	recorder := events.NewEventRecorder(storage, bus)

	// Start recording all events
	err := recorder.Start()
	if err != nil {
		log.Printf("Failed to start recorder: %v", err)
		return
	}
	defer recorder.Stop()

	// Generate some events
	startTime := time.Now()
	for i := 0; i < 5; i++ {
		event := domain.NewEvent(domain.EventProgress, "agent1", "ReplayDemo", &domain.ProgressEventData{
			Current: i + 1,
			Total:   5,
			Message: fmt.Sprintf("Step %d", i+1),
		})
		bus.Publish(event)
		time.Sleep(100 * time.Millisecond)
	}

	// Stop recording and query stored events
	recorder.Stop()

	query := events.EventQuery{
		StartTime: &startTime,
		Limit:     10,
	}

	storedEvents, err := storage.Query(context.Background(), query)
	if err != nil {
		log.Printf("Failed to query events: %v", err)
		return
	}

	fmt.Printf("  Recorded %d events\n", len(storedEvents))

	// Set up replay handler
	replayHandler := events.EventHandlerFunc(func(ctx context.Context, event domain.Event) error {
		if data, ok := event.Data.(*domain.ProgressEventData); ok {
			fmt.Printf("  Replaying: %s (originally at %s)\n", data.Message, event.Timestamp.Format("15:04:05.000"))
		}
		return nil
	})

	replayBus := events.NewEventBus()
	defer replayBus.Close()
	replayBus.Subscribe(replayHandler)

	// Replay events at 2x speed
	fmt.Println("\n  Replaying events at 2x speed:")
	replayer := events.NewEventReplayer(storage, replayBus)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = replayer.Replay(ctx, query, events.ReplayOptions{
		Speed: 2.0, // 2x speed
	})

	if err != nil {
		log.Printf("Replay failed: %v", err)
	}
}

func demonstrateBridgeIntegration(bus *events.EventBus) {
	// Create bridge event publisher
	bridgePublisher := events.NewBridgeEventPublisher(bus, "bridge-001", "session-123")

	// Create bridge event listener
	bridgeHandler := events.BridgeEventHandlerFunc(func(ctx context.Context, event *events.BridgeEvent) error {
		fmt.Printf("  Bridge event: Type=%s, Bridge=%s, Session=%s\n",
			event.Type, event.BridgeID, event.SessionID)

		// Handle specific bridge event types
		switch events.BridgeEventType(event.Type) {
		case events.BridgeEventRequest:
			if data, ok := event.Data.(*events.BridgeRequestData); ok {
				fmt.Printf("    Request: Method=%s, ID=%s\n", data.Method, data.RequestID)
			}
		case events.BridgeEventScriptStart:
			if data, ok := event.Data.(*events.ScriptExecutionData); ok {
				fmt.Printf("    Script: Language=%s, ID=%s\n", data.Language, data.ScriptID)
			}
		}

		return nil
	})

	listener := events.NewBridgeEventListener(bus, bridgeHandler)
	err := listener.Listen("bridge.*")
	if err != nil {
		log.Printf("Failed to start bridge listener: %v", err)
		return
	}
	defer listener.Stop()

	// Publish bridge events

	// 1. Request event
	requestID := bridgePublisher.PublishRequest("executeScript", map[string]interface{}{
		"language": "javascript",
		"code":     "console.log('Hello from bridge');",
	})

	// 2. Script execution event
	scriptData := &events.ScriptExecutionData{
		ScriptID:  "script-001",
		Language:  "javascript",
		Source:    "console.log('Hello');",
		StartTime: time.Now(),
	}
	bridgePublisher.PublishScriptExecution(scriptData)

	// 3. Response event
	bridgePublisher.PublishResponse(requestID, map[string]interface{}{
		"output": "Hello from bridge",
		"status": "success",
	}, nil, 50*time.Millisecond)

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Show bridge event serialization
	fmt.Println("\n  Bridge event serialization:")
	bridgeEvent := events.NewBridgeEvent(events.BridgeEventConnected, "bridge-001", "session-123", nil)
	bridgeEvent.WithLanguage("javascript").WithScriptData("engine", "v8")

	serialized, err := events.SerializeBridgeEvent(bridgeEvent)
	if err != nil {
		log.Printf("Failed to serialize bridge event: %v", err)
		return
	}

	fmt.Printf("  Serialized bridge event: %v\n", serialized)
}
