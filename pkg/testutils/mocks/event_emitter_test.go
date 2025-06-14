// ABOUTME: Tests for MockEventEmitter implementation verifying event handling
// ABOUTME: Covers event emission, filtering, listeners, assertions, and async behavior

package mocks

import (
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockEventEmitter_Basic(t *testing.T) {
	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	// Emit basic event
	data := map[string]string{"key": "value"}
	emitter.Emit(domain.EventAgentStart, data)

	// Check event was recorded
	events := emitter.GetEvents()
	assert.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, domain.EventAgentStart, event.Type)
	assert.Equal(t, "test-agent", event.AgentID)
	assert.Equal(t, "Test Agent", event.AgentName)
	assert.Equal(t, data, event.Data)
	assert.NotEmpty(t, event.ID)
	assert.False(t, event.Timestamp.IsZero())
}

func TestMockEventEmitter_SpecializedEmits(t *testing.T) {
	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	// Test EmitProgress
	emitter.EmitProgress(50, 100, "Processing...")

	events := emitter.GetEventsByType(domain.EventProgress)
	assert.Len(t, events, 1)

	progressData := events[0].Data.(map[string]interface{})
	assert.Equal(t, 50, progressData["current"])
	assert.Equal(t, 100, progressData["total"])
	assert.Equal(t, "Processing...", progressData["message"])
	assert.Equal(t, 50.0, progressData["percent"])

	// Test EmitMessage
	emitter.EmitMessage("Test message")

	msgEvents := emitter.GetEventsByType(domain.EventMessage)
	assert.Len(t, msgEvents, 1)

	msgData := msgEvents[0].Data.(map[string]interface{})
	assert.Equal(t, "Test message", msgData["message"])

	// Test EmitError
	testErr := assert.AnError
	emitter.EmitError(testErr)

	errorEvents := emitter.GetEventsByType(domain.EventAgentError)
	assert.Len(t, errorEvents, 1)
	assert.Equal(t, testErr, errorEvents[0].Error)

	// Test EmitCustom
	customData := map[string]int{"count": 42}
	emitter.EmitCustom("my.event", customData)

	customEvents := emitter.GetEventsByType("custom.my.event")
	assert.Len(t, customEvents, 1)
	assert.Equal(t, customData, customEvents[0].Data)
}

func TestMockEventEmitter_Filtering(t *testing.T) {
	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	// Emit various events
	emitter.Emit(domain.EventAgentStart, nil)
	emitter.Emit(domain.EventProgress, map[string]interface{}{"current": 1})
	emitter.Emit(domain.EventProgress, map[string]interface{}{"current": 2})
	emitter.EmitError(assert.AnError)
	emitter.Emit(domain.EventAgentComplete, nil)

	// Test GetEventsByType
	progressEvents := emitter.GetEventsByType(domain.EventProgress)
	assert.Len(t, progressEvents, 2)

	// Test GetEventsByFilter
	errorFilter := func(e domain.Event) bool {
		return e.IsError()
	}
	errorEvents := emitter.GetEventsByFilter(errorFilter)
	assert.Len(t, errorEvents, 1)

	// Custom filter
	progressFilter := func(e domain.Event) bool {
		if e.Type != domain.EventProgress {
			return false
		}
		data := e.Data.(map[string]interface{})
		current := data["current"].(int)
		return current > 1
	}
	filteredProgress := emitter.GetEventsByFilter(progressFilter)
	assert.Len(t, filteredProgress, 1)
}

func TestMockEventEmitter_Listeners(t *testing.T) {
	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	// Add listener
	receivedEvents := make([]domain.Event, 0)
	var mu sync.Mutex

	listener := func(event domain.Event) {
		mu.Lock()
		receivedEvents = append(receivedEvents, event)
		mu.Unlock()
	}

	emitter.AddListener(listener)

	// Emit events
	emitter.Emit(domain.EventAgentStart, nil)
	emitter.Emit(domain.EventAgentComplete, nil)

	// Give listeners time to process
	time.Sleep(10 * time.Millisecond)

	// Check listener received events
	mu.Lock()
	assert.Len(t, receivedEvents, 2)
	mu.Unlock()

	// Remove all listeners
	emitter.RemoveAllListeners()

	// Emit more events
	emitter.Emit(domain.EventMessage, nil)

	// Listener should not receive new events
	time.Sleep(10 * time.Millisecond)
	mu.Lock()
	assert.Len(t, receivedEvents, 2) // Still only 2
	mu.Unlock()
}

func TestMockEventEmitter_BlockEvents(t *testing.T) {
	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	// Block events
	emitter.SetBlockEvents(true)

	// Try to emit
	emitter.Emit(domain.EventAgentStart, nil)

	// Should not be recorded
	assert.Empty(t, emitter.GetEvents())

	// Unblock
	emitter.SetBlockEvents(false)

	// Now should work
	emitter.Emit(domain.EventAgentComplete, nil)
	assert.Len(t, emitter.GetEvents(), 1)
}

func TestMockEventEmitter_AsyncEmit(t *testing.T) {
	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	// Enable async emit
	emitter.SetAsyncEmit(true)

	// Emit event
	emitter.Emit(domain.EventAgentStart, nil)

	// Immediate check might not see it
	// Wait a bit for async processing
	time.Sleep(50 * time.Millisecond)

	events := emitter.GetEvents()
	assert.Len(t, events, 1)
}

func TestMockEventEmitter_EventDelay(t *testing.T) {
	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	// Set delay
	emitter.SetEventDelay(100 * time.Millisecond)

	start := time.Now()
	emitter.Emit(domain.EventAgentStart, nil)
	duration := time.Since(start)

	// Should have delayed
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
}

func TestMockEventEmitter_BehaviorHooks(t *testing.T) {
	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	// Test OnEmit hook
	emitCalled := false
	var emitType domain.EventType
	var emitData interface{}

	emitter.OnEmit = func(eventType domain.EventType, data interface{}) {
		emitCalled = true
		emitType = eventType
		emitData = data
	}

	testData := map[string]string{"test": "data"}
	emitter.Emit(domain.EventAgentStart, testData)

	assert.True(t, emitCalled)
	assert.Equal(t, domain.EventAgentStart, emitType)
	assert.Equal(t, testData, emitData)

	// Test OnEmitProgress hook
	progressCalled := false
	emitter.OnEmitProgress = func(current, total int, message string) {
		progressCalled = true
		assert.Equal(t, 25, current)
		assert.Equal(t, 100, total)
		assert.Equal(t, "Testing", message)
	}

	emitter.EmitProgress(25, 100, "Testing")
	assert.True(t, progressCalled)

	// Test OnEmitMessage hook
	messageCalled := false
	emitter.OnEmitMessage = func(message string) {
		messageCalled = true
		assert.Equal(t, "Test message", message)
	}

	emitter.EmitMessage("Test message")
	assert.True(t, messageCalled)

	// Test OnEmitError hook
	errorCalled := false
	emitter.OnEmitError = func(err error) {
		errorCalled = true
		assert.Equal(t, assert.AnError, err)
	}

	emitter.EmitError(assert.AnError)
	assert.True(t, errorCalled)

	// Test OnEmitCustom hook
	customCalled := false
	emitter.OnEmitCustom = func(eventName string, data interface{}) {
		customCalled = true
		assert.Equal(t, "test.event", eventName)
	}

	emitter.EmitCustom("test.event", nil)
	assert.True(t, customCalled)
}

func TestMockEventEmitter_Assertions(t *testing.T) {
	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	// Emit some events
	emitter.Emit(domain.EventAgentStart, nil)
	emitter.EmitProgress(50, 100, "Processing")
	emitter.EmitMessage("Status update")

	// Test AssertEventEmitted
	err := emitter.AssertEventEmitted(domain.EventAgentStart)
	require.NoError(t, err)

	err = emitter.AssertEventEmitted(domain.EventAgentError)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected event type agent.error to be emitted")

	// Test AssertEventCount
	err = emitter.AssertEventCount(3)
	require.NoError(t, err)

	err = emitter.AssertEventCount(5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 5 events, got 3")

	// Test AssertEventTypeCount
	err = emitter.AssertEventTypeCount(domain.EventProgress, 1)
	require.NoError(t, err)

	err = emitter.AssertEventTypeCount(domain.EventMessage, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 2 events of type message")

	// Test AssertNoErrors
	err = emitter.AssertNoErrors()
	require.NoError(t, err)

	// Emit an error and test again
	emitter.EmitError(assert.AnError)
	err = emitter.AssertNoErrors()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected no error events")
}

func TestMockEventEmitter_WaitForEvent(t *testing.T) {
	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	// Emit event after delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		emitter.Emit(domain.EventAgentComplete, map[string]string{"status": "done"})
	}()

	// Wait for event
	event, err := emitter.WaitForEvent(domain.EventAgentComplete, 200*time.Millisecond)
	require.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, domain.EventAgentComplete, event.Type)

	// Test timeout
	event, err = emitter.WaitForEvent(domain.EventAgentError, 50*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting for event")
	assert.Nil(t, event)

	// Test immediate event (already emitted)
	event, err = emitter.WaitForEvent(domain.EventAgentComplete, 100*time.Millisecond)
	require.NoError(t, err)
	assert.NotNil(t, event)
}

func TestMockEventEmitter_Reset(t *testing.T) {
	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	// Add data
	emitter.Emit(domain.EventAgentStart, nil)
	emitter.AddListener(func(e domain.Event) {})
	emitter.SetBlockEvents(true)
	emitter.SetAsyncEmit(true)
	emitter.SetEventDelay(100 * time.Millisecond)

	// Reset
	emitter.Reset()

	// Verify reset
	assert.Empty(t, emitter.GetEvents())
	assert.Empty(t, emitter.listeners)
	assert.False(t, emitter.blockEvents)
	assert.False(t, emitter.asyncEmit)
	assert.Equal(t, time.Duration(0), emitter.eventDelay)
}

func TestCreateMockToolContext(t *testing.T) {
	state := NewMockState()
	state.Set("key", "value")

	emitter := NewMockEventEmitter("test-agent", "Test Agent")

	ctx := CreateMockToolContext(state, emitter)

	assert.NotNil(t, ctx)
	assert.NotNil(t, ctx.Context)
	assert.Equal(t, state.State, ctx.State)
	assert.NotEmpty(t, ctx.RunID)
	assert.Equal(t, 0, ctx.Retry)
	assert.NotZero(t, ctx.StartTime)
	assert.Equal(t, emitter, ctx.Events)
	assert.Equal(t, "test-agent", ctx.Agent.ID)
	assert.Equal(t, "Test Agent", ctx.Agent.Name)
	assert.Equal(t, domain.AgentTypeCustom, ctx.Agent.Type)
}
