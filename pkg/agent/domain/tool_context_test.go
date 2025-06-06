// ABOUTME: Tests for ToolContext and related components
// ABOUTME: Verifies state reader, event emitter, and context functionality

package domain

import (
	"context"
	"testing"
	"time"

	domain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToolContext(t *testing.T) {
	// Create test components
	ctx := context.Background()
	state := NewState()
	state.Set("key1", "value1")
	state.Set("key2", 42)

	agent := &mockAgent{
		id:          "agent-123",
		name:        "test-agent",
		description: "Test agent",
		agentType:   AgentTypeLLM,
		metadata:    map[string]interface{}{"custom": "data"},
	}

	// Create tool context
	tc := NewToolContext(ctx, NewStateReader(state), agent, "run-456")

	// Verify basic properties
	assert.Equal(t, ctx, tc.Context)
	assert.Equal(t, "run-456", tc.RunID)
	assert.Equal(t, 0, tc.Retry)
	assert.NotZero(t, tc.StartTime)

	// Verify agent info
	assert.Equal(t, "agent-123", tc.Agent.ID)
	assert.Equal(t, "test-agent", tc.Agent.Name)
	assert.Equal(t, "Test agent", tc.Agent.Description)
	assert.Equal(t, AgentTypeLLM, tc.Agent.Type)
	assert.Equal(t, "data", tc.Agent.Metadata["custom"])

	// Verify state reader
	val1, exists := tc.State.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", val1)

	val2, exists := tc.State.Get("key2")
	assert.True(t, exists)
	assert.Equal(t, 42, val2)
}

func TestToolContext_WithRetry(t *testing.T) {
	ctx := context.Background()
	state := NewState()
	agent := &mockAgent{id: "test", name: "test", agentType: AgentTypeCustom}

	tc := NewToolContext(ctx, NewStateReader(state), agent, "run-1")
	assert.Equal(t, 0, tc.Retry)

	tc2 := tc.WithRetry(3)
	assert.Equal(t, 3, tc2.Retry)
	assert.Equal(t, 0, tc.Retry) // Original unchanged
}

func TestToolContext_WithEventEmitter(t *testing.T) {
	ctx := context.Background()
	state := NewState()
	agent := &mockAgent{id: "test", name: "test", agentType: AgentTypeCustom}

	tc := NewToolContext(ctx, NewStateReader(state), agent, "run-1")
	assert.Nil(t, tc.Events)

	emitter := &mockEventEmitter{}
	tc2 := tc.WithEventEmitter(emitter)
	assert.Equal(t, emitter, tc2.Events)
	assert.Nil(t, tc.Events) // Original unchanged
}

func TestToolContext_ContextMethods(t *testing.T) {
	// Test with deadline
	deadline := time.Now().Add(5 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	state := NewState()
	agent := &mockAgent{id: "test", name: "test", agentType: AgentTypeCustom}
	tc := NewToolContext(ctx, NewStateReader(state), agent, "run-1")

	// Test Deadline
	dl, ok := tc.Deadline()
	assert.True(t, ok)
	assert.Equal(t, deadline, dl)

	// Test Done channel
	assert.NotNil(t, tc.Done())

	// Test Value
	type contextKey string
	testKey := contextKey("test-key")
	ctx2 := context.WithValue(ctx, testKey, "test-value")
	tc2 := NewToolContext(ctx2, NewStateReader(state), agent, "run-2")
	assert.Equal(t, "test-value", tc2.Value(testKey))
}

func TestToolContext_ElapsedTime(t *testing.T) {
	ctx := context.Background()
	state := NewState()
	agent := &mockAgent{id: "test", name: "test", agentType: AgentTypeCustom}

	tc := NewToolContext(ctx, NewStateReader(state), agent, "run-1")

	// Sleep a bit
	time.Sleep(10 * time.Millisecond)

	elapsed := tc.ElapsedTime()
	assert.True(t, elapsed >= 10*time.Millisecond)
}

func TestStateReader(t *testing.T) {
	// Create state with test data
	state := NewState()
	state.Set("string", "value")
	state.Set("number", 42)
	state.Set("bool", true)
	state.SetMetadata("meta1", "metadata")

	artifact := NewArtifact("test.txt", ArtifactTypeFile, []byte("content"))
	state.AddArtifact(artifact)

	msg := Message{Role: "user", Content: "Hello"}
	state.AddMessage(msg)

	// Create reader
	reader := NewStateReader(state)

	// Test Get
	val, exists := reader.Get("string")
	assert.True(t, exists)
	assert.Equal(t, "value", val)

	_, exists = reader.Get("nonexistent")
	assert.False(t, exists)

	// Test Values
	values := reader.Values()
	assert.Len(t, values, 3)
	assert.Equal(t, "value", values["string"])
	assert.Equal(t, 42, values["number"])
	assert.Equal(t, true, values["bool"])

	// Test GetArtifact
	art, exists := reader.GetArtifact(artifact.ID)
	assert.True(t, exists)
	assert.Equal(t, "test.txt", art.Name)

	// Test Artifacts
	artifacts := reader.Artifacts()
	assert.Len(t, artifacts, 1)

	// Test Messages
	messages := reader.Messages()
	assert.Len(t, messages, 1)
	assert.Equal(t, "Hello", messages[0].Content)

	// Test GetMetadata
	meta, exists := reader.GetMetadata("meta1")
	assert.True(t, exists)
	assert.Equal(t, "metadata", meta)

	// Test Has
	assert.True(t, reader.Has("string"))
	assert.False(t, reader.Has("nonexistent"))

	// Test Keys
	keys := reader.Keys()
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "string")
	assert.Contains(t, keys, "number")
	assert.Contains(t, keys, "bool")
}

func TestToolEventEmitter(t *testing.T) {
	dispatcher := &mockEventDispatcher{}
	emitter := NewToolEventEmitter(dispatcher, "test-tool", "agent-1", "TestAgent")

	// Test basic emit
	emitter.Emit(EventToolCall, "test data")
	require.Len(t, dispatcher.events, 1)
	event := dispatcher.events[0]
	assert.Equal(t, EventToolCall, event.Type)
	assert.Equal(t, "agent-1", event.AgentID)
	assert.Equal(t, "TestAgent", event.AgentName)
	assert.Equal(t, "test data", event.Data)
	assert.Equal(t, "test-tool", event.Metadata["tool_name"])
	assert.Equal(t, "tool", event.Metadata["source"])

	// Test progress emit
	dispatcher.events = nil
	emitter.EmitProgress(5, 10, "Processing")
	require.Len(t, dispatcher.events, 1)
	event = dispatcher.events[0]
	assert.Equal(t, EventProgress, event.Type)
	progressData := event.Data.(ProgressEventData)
	assert.Equal(t, 5, progressData.Current)
	assert.Equal(t, 10, progressData.Total)
	assert.Equal(t, "Processing", progressData.Message)

	// Test message emit
	dispatcher.events = nil
	emitter.EmitMessage("Info message")
	require.Len(t, dispatcher.events, 1)
	event = dispatcher.events[0]
	assert.Equal(t, EventMessage, event.Type)
	msgData := event.Data.(MessageEventData)
	assert.Equal(t, "Info message", msgData.Message)
	assert.Equal(t, "info", msgData.Level)

	// Test error emit
	dispatcher.events = nil
	emitter.EmitError(assert.AnError)
	require.Len(t, dispatcher.events, 1)
	event = dispatcher.events[0]
	assert.Equal(t, EventToolError, event.Type)
	assert.Equal(t, assert.AnError.Error(), event.Data)

	// Test nil error (should not emit)
	dispatcher.events = nil
	emitter.EmitError(nil)
	assert.Len(t, dispatcher.events, 0)

	// Test custom emit
	dispatcher.events = nil
	emitter.EmitCustom("status", map[string]string{"status": "ready"})
	require.Len(t, dispatcher.events, 1)
	event = dispatcher.events[0]
	assert.Equal(t, EventType("tool.test-tool.status"), event.Type)
	assert.Equal(t, map[string]string{"status": "ready"}, event.Data)
}

// Mock implementations for testing

type mockAgent struct {
	id          string
	name        string
	description string
	agentType   AgentType
	metadata    map[string]interface{}
}

func (m *mockAgent) ID() string                                            { return m.id }
func (m *mockAgent) Name() string                                          { return m.name }
func (m *mockAgent) Description() string                                   { return m.description }
func (m *mockAgent) Type() AgentType                                       { return m.agentType }
func (m *mockAgent) Parent() BaseAgent                                     { return nil }
func (m *mockAgent) SetParent(parent BaseAgent) error                      { return nil }
func (m *mockAgent) SubAgents() []BaseAgent                                { return nil }
func (m *mockAgent) AddSubAgent(agent BaseAgent) error                     { return nil }
func (m *mockAgent) RemoveSubAgent(name string) error                      { return nil }
func (m *mockAgent) FindAgent(name string) BaseAgent                       { return nil }
func (m *mockAgent) FindSubAgent(name string) BaseAgent                    { return nil }
func (m *mockAgent) Run(ctx context.Context, input *State) (*State, error) { return nil, nil }
func (m *mockAgent) RunAsync(ctx context.Context, input *State) (<-chan Event, error) {
	return nil, nil
}
func (m *mockAgent) Initialize(ctx context.Context) error              { return nil }
func (m *mockAgent) BeforeRun(ctx context.Context, state *State) error { return nil }
func (m *mockAgent) AfterRun(ctx context.Context, state *State, result *State, err error) error {
	return nil
}
func (m *mockAgent) Cleanup(ctx context.Context) error         { return nil }
func (m *mockAgent) InputSchema() *domain.Schema               { return nil }
func (m *mockAgent) OutputSchema() *domain.Schema              { return nil }
func (m *mockAgent) Config() AgentConfig                       { return AgentConfig{} }
func (m *mockAgent) WithConfig(config AgentConfig) BaseAgent   { return m }
func (m *mockAgent) Validate() error                           { return nil }
func (m *mockAgent) Metadata() map[string]interface{}          { return m.metadata }
func (m *mockAgent) SetMetadata(key string, value interface{}) {}

type mockEventDispatcher struct {
	events []Event
}

func (m *mockEventDispatcher) Dispatch(event Event) {
	m.events = append(m.events, event)
}

func (m *mockEventDispatcher) Subscribe(handler EventHandler, filters ...EventFilter) string {
	return "sub-1"
}
func (m *mockEventDispatcher) Unsubscribe(subscriptionID string) {}
func (m *mockEventDispatcher) Close()                            {}

type mockEventEmitter struct {
	events []interface{}
}

func (m *mockEventEmitter) Emit(eventType EventType, data interface{}) {
	m.events = append(m.events, data)
}
func (m *mockEventEmitter) EmitProgress(current, total int, message string) {
	m.events = append(m.events, ProgressEventData{Current: current, Total: total, Message: message})
}
func (m *mockEventEmitter) EmitMessage(message string) {
	m.events = append(m.events, message)
}
func (m *mockEventEmitter) EmitError(err error) {
	if err != nil {
		m.events = append(m.events, err)
	}
}
func (m *mockEventEmitter) EmitCustom(eventName string, data interface{}) {
	m.events = append(m.events, data)
}
