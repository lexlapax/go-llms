// ABOUTME: Tests for the TracingHook interface and implementations for OpenTelemetry integration
// ABOUTME: including agent, tool, and event tracing with metrics collection

package core

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Mock implementations for testing

// mockSpan implements the Span interface
type mockSpan struct {
	name       string
	ended      bool
	attributes []Attribute
	errors     []error
	status     StatusCode
	statusDesc string
	recording  bool
}

func (s *mockSpan) End() {
	s.ended = true
}

func (s *mockSpan) SetAttributes(attributes ...Attribute) {
	s.attributes = append(s.attributes, attributes...)
}

func (s *mockSpan) RecordError(err error) {
	s.errors = append(s.errors, err)
}

func (s *mockSpan) SetStatus(code StatusCode, description string) {
	s.status = code
	s.statusDesc = description
}

func (s *mockSpan) IsRecording() bool {
	return s.recording
}

// mockTracer implements the Tracer interface
type mockTracer struct {
	spans []*mockSpan
}

func (t *mockTracer) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	span := &mockSpan{
		name:      name,
		recording: true,
	}

	// Apply options
	cfg := &SpanConfig{}
	for _, opt := range opts {
		opt.Apply(cfg)
	}
	span.attributes = cfg.Attributes

	t.spans = append(t.spans, span)
	return ContextWithSpan(ctx, span), span
}

// mockAgent implements domain.BaseAgent for testing
type mockAgent struct {
	id          string
	name        string
	agentType   domain.AgentType
	description string
	parent      domain.BaseAgent
	subAgents   []domain.BaseAgent
	config      domain.AgentConfig
	metadata    map[string]interface{}
}

func (a *mockAgent) ID() string               { return a.id }
func (a *mockAgent) Name() string             { return a.name }
func (a *mockAgent) Type() domain.AgentType   { return a.agentType }
func (a *mockAgent) Description() string      { return a.description }
func (a *mockAgent) Parent() domain.BaseAgent { return a.parent }
func (a *mockAgent) SetParent(p domain.BaseAgent) error {
	a.parent = p
	return nil
}
func (a *mockAgent) SubAgents() []domain.BaseAgent { return a.subAgents }
func (a *mockAgent) AddSubAgent(s domain.BaseAgent) error {
	a.subAgents = append(a.subAgents, s)
	return nil
}
func (a *mockAgent) RemoveSubAgent(name string) error {
	for i, sub := range a.subAgents {
		if sub.Name() == name {
			a.subAgents = append(a.subAgents[:i], a.subAgents[i+1:]...)
			break
		}
	}
	return nil
}
func (a *mockAgent) FindAgent(name string) domain.BaseAgent {
	if a.name == name {
		return a
	}
	for _, sub := range a.subAgents {
		if found := sub.FindAgent(name); found != nil {
			return found
		}
	}
	return nil
}
func (a *mockAgent) FindSubAgent(name string) domain.BaseAgent {
	for _, sub := range a.subAgents {
		if sub.Name() == name {
			return sub
		}
	}
	return nil
}
func (a *mockAgent) Initialize(ctx context.Context) error { return nil }
func (a *mockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	return state, nil
}
func (a *mockAgent) RunAsync(ctx context.Context, state *domain.State) (<-chan domain.Event, error) {
	ch := make(chan domain.Event)
	close(ch)
	return ch, nil
}
func (a *mockAgent) BeforeRun(ctx context.Context, state *domain.State) error { return nil }
func (a *mockAgent) AfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) error {
	return nil
}
func (a *mockAgent) Cleanup(ctx context.Context) error { return nil }
func (a *mockAgent) InputSchema() *sdomain.Schema      { return nil }
func (a *mockAgent) OutputSchema() *sdomain.Schema     { return nil }
func (a *mockAgent) Config() domain.AgentConfig        { return a.config }
func (a *mockAgent) WithConfig(config domain.AgentConfig) domain.BaseAgent {
	a.config = config
	return a
}
func (a *mockAgent) Validate() error { return nil }
func (a *mockAgent) Metadata() map[string]interface{} {
	if a.metadata == nil {
		a.metadata = make(map[string]interface{})
	}
	return a.metadata
}
func (a *mockAgent) SetMetadata(key string, value interface{}) {
	if a.metadata == nil {
		a.metadata = make(map[string]interface{})
	}
	a.metadata[key] = value
}

// Tests

func TestTracingHook(t *testing.T) {
	tracer := &mockTracer{}
	hook := NewTracingHook("test-tracer", tracer)

	agent := &mockAgent{
		id:        "test-agent",
		name:      "TestAgent",
		agentType: domain.AgentTypeLLM,
	}

	state := domain.NewState()
	state.Set("test", "value")

	ctx := context.Background()

	// Test BeforeRun
	newCtx, err := hook.BeforeRun(ctx, agent, state)
	if err != nil {
		t.Errorf("BeforeRun failed: %v", err)
	}

	// Verify span was created
	if len(tracer.spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(tracer.spans))
	}

	span := tracer.spans[0]
	if span.name != "agent.TestAgent.run" {
		t.Errorf("Expected span name 'agent.TestAgent.run', got '%s'", span.name)
	}

	// Check attributes
	expectedAttrs := map[string]interface{}{
		"agent.id":              "test-agent",
		"agent.type":            "llm",
		"agent.name":            "TestAgent",
		"state.id":              state.ID(),
		"state.values.count":    1,
		"state.messages.count":  0,
		"state.artifacts.count": 0,
	}

	for key, expected := range expectedAttrs {
		found := false
		for _, attr := range span.attributes {
			if attr.Key == key {
				found = true
				if fmt.Sprintf("%v", attr.Value) != fmt.Sprintf("%v", expected) {
					t.Errorf("Attribute %s: expected %v, got %v", key, expected, attr.Value)
				}
				break
			}
		}
		if !found {
			t.Errorf("Missing attribute: %s", key)
		}
	}

	// Test AfterRun with success
	result := domain.NewState()
	result.Set("result", "success")

	err = hook.AfterRun(newCtx, agent, state, result, nil)
	if err != nil {
		t.Errorf("AfterRun failed: %v", err)
	}

	if !span.ended {
		t.Error("Expected span to be ended")
	}

	if span.status != StatusCodeOk {
		t.Errorf("Expected status OK, got %v", span.status)
	}

	// Test AfterRun with error
	tracer.spans = nil
	newCtx, _ = hook.BeforeRun(ctx, agent, state)
	testErr := errors.New("test error")
	_ = hook.AfterRun(newCtx, agent, state, nil, testErr)

	span = tracer.spans[0]
	if span.status != StatusCodeError {
		t.Error("Expected error status")
	}
	if len(span.errors) != 1 || span.errors[0] != testErr {
		t.Error("Expected error to be recorded")
	}
}

func TestToolCallTracingHook(t *testing.T) {
	tracer := &mockTracer{}
	hook := NewToolCallTracingHook("test-tracer", tracer)

	ctx := context.Background()

	// Test BeforeToolCall
	params := map[string]interface{}{"input": "test"}
	newCtx, err := hook.BeforeToolCall(ctx, "test-tool", params)
	if err != nil {
		t.Errorf("BeforeToolCall failed: %v", err)
	}

	if len(tracer.spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(tracer.spans))
	}

	span := tracer.spans[0]
	if span.name != "tool.test-tool.call" {
		t.Errorf("Expected span name 'tool.test-tool.call', got '%s'", span.name)
	}

	// Test AfterToolCall with success
	result := map[string]interface{}{"output": "success"}
	err = hook.AfterToolCall(newCtx, "test-tool", result, nil)
	if err != nil {
		t.Errorf("AfterToolCall failed: %v", err)
	}

	if !span.ended {
		t.Error("Expected span to be ended")
	}

	if span.status != StatusCodeOk {
		t.Error("Expected OK status")
	}

	// Test AfterToolCall with error
	tracer.spans = nil
	newCtx, _ = hook.BeforeToolCall(ctx, "test-tool", params)
	testErr := errors.New("tool error")
	_ = hook.AfterToolCall(newCtx, "test-tool", nil, testErr)

	span = tracer.spans[0]
	if span.status != StatusCodeError {
		t.Error("Expected error status")
	}
}

func TestEventTracingHook(t *testing.T) {
	tracer := &mockTracer{}
	hook := NewEventTracingHook("test-tracer", tracer)

	event := domain.Event{
		ID:        "event-1",
		Type:      domain.EventStateUpdate,
		AgentID:   "test-agent",
		AgentName: "TestAgent",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"key": "value"},
	}

	// Test HandleEvent
	err := hook.HandleEvent(event)
	if err != nil {
		t.Errorf("HandleEvent failed: %v", err)
	}

	if len(tracer.spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(tracer.spans))
	}

	span := tracer.spans[0]
	if span.name != "event.state.update" {
		t.Errorf("Expected span name 'event.state.update', got '%s'", span.name)
	}

	if !span.ended {
		t.Error("Expected span to be ended")
	}

	// Test with error event
	tracer.spans = nil
	errorEvent := event
	errorEvent.Error = errors.New("event error")

	err = hook.HandleEvent(errorEvent)
	if err != nil {
		t.Errorf("HandleEvent failed: %v", err)
	}

	span = tracer.spans[0]
	if span.status != StatusCodeError {
		t.Error("Expected error status for error event")
	}
}

func TestCompositeTracingHook(t *testing.T) {
	tracer := &mockTracer{}
	composite := NewCompositeTracingHook("test-tracer", tracer)

	// Verify all hooks are created
	if composite.GetAgentHook() == nil {
		t.Error("Expected agent hook to be created")
	}

	if composite.GetToolHook() == nil {
		t.Error("Expected tool hook to be created")
	}

	if composite.GetEventHandler() == nil {
		t.Error("Expected event handler to be created")
	}

	// Test that they all use the same tracer
	agent := &mockAgent{id: "test", name: "Test", agentType: domain.AgentTypeLLM}
	state := domain.NewState()

	ctx := context.Background()

	// Use agent hook
	newCtx, _ := composite.GetAgentHook().BeforeRun(ctx, agent, state)
	_ = composite.GetAgentHook().AfterRun(newCtx, agent, state, state, nil)

	// Use tool hook
	toolCtx, _ := composite.GetToolHook().BeforeToolCall(ctx, "tool", nil)
	_ = composite.GetToolHook().AfterToolCall(toolCtx, "tool", nil, nil)

	// Use event handler
	_ = composite.GetEventHandler().HandleEvent(domain.Event{
		Type:    domain.EventAgentStart,
		AgentID: "test",
	})

	// Should have 3 spans total
	if len(tracer.spans) != 3 {
		t.Errorf("Expected 3 spans, got %d", len(tracer.spans))
	}
}

func TestNoOpTracer(t *testing.T) {
	tracer := &NoOpTracer{}
	ctx := context.Background()

	// Test that NoOp tracer returns context unchanged and no-op span
	newCtx, span := tracer.Start(ctx, "test")

	if newCtx != ctx {
		t.Error("NoOp tracer should return same context")
	}

	// Test NoOp span methods don't panic
	span.End()
	span.SetAttributes(Attribute{Key: "test", Value: "value"})
	span.RecordError(errors.New("test"))
	span.SetStatus(StatusCodeError, "error")

	if span.IsRecording() {
		t.Error("NoOp span should not be recording")
	}
}

func TestMetricsHook(t *testing.T) {
	hook := NewMetricsHook()

	// Record some executions
	hook.RecordExecution("agent1", 100*time.Millisecond)
	hook.RecordExecution("agent1", 200*time.Millisecond)
	hook.RecordExecution("agent2", 50*time.Millisecond)

	// Record some errors
	hook.RecordError("agent1")
	hook.RecordError("agent2")
	hook.RecordError("agent2")

	// Get metrics
	metrics := hook.GetMetrics()

	avgTimes, ok := metrics["average_execution_times"].(map[string]float64)
	if !ok {
		t.Fatal("Expected average_execution_times to be a map")
	}

	// Check average for agent1: (100ms + 200ms) / 2 = 150ms
	if avg, exists := avgTimes["agent1"]; !exists || avg != float64(150*time.Millisecond) {
		t.Errorf("Expected agent1 average to be 150ms, got %v", avg)
	}

	errorCounts, ok := metrics["error_counts"].(map[string]int)
	if !ok {
		t.Fatal("Expected error_counts to be a map")
	}

	if errorCounts["agent1"] != 1 {
		t.Errorf("Expected agent1 to have 1 error, got %d", errorCounts["agent1"])
	}

	if errorCounts["agent2"] != 2 {
		t.Errorf("Expected agent2 to have 2 errors, got %d", errorCounts["agent2"])
	}
}

func TestHelperFunctions(t *testing.T) {
	span := &mockSpan{recording: true}

	// Test AddStateAttributes
	state := domain.NewState()
	state.Set("key", "value")
	state.AddMessage(domain.NewMessage("user", "test"))

	AddStateAttributes(span, state)

	// Check that state attributes were added
	hasStateID := false
	for _, attr := range span.attributes {
		if attr.Key == "state.id" {
			hasStateID = true
			break
		}
	}
	if !hasStateID {
		t.Error("Expected state.id attribute")
	}

	// Test AddAgentAttributes
	agent := &mockAgent{
		id:          "test-agent",
		name:        "TestAgent",
		agentType:   domain.AgentTypeLLM,
		description: "Test agent",
	}

	AddAgentAttributes(span, agent)

	hasAgentID := false
	for _, attr := range span.attributes {
		if attr.Key == "agent.id" && attr.Value == "test-agent" {
			hasAgentID = true
			break
		}
	}
	if !hasAgentID {
		t.Error("Expected agent.id attribute")
	}

	// Test AddErrorAttributes
	span.errors = nil
	span.attributes = nil
	agentErr := &domain.AgentError{
		AgentID:   "test-agent",
		AgentName: "TestAgent",
		Phase:     "execution",
		Err:       errors.New("test error"),
	}

	AddErrorAttributes(span, agentErr)

	if len(span.errors) != 1 {
		t.Error("Expected error to be recorded")
	}

	hasErrorType := false
	for _, attr := range span.attributes {
		if attr.Key == "error.type" && attr.Value == "agent_error" {
			hasErrorType = true
			break
		}
	}
	if !hasErrorType {
		t.Error("Expected error.type attribute")
	}
}

func TestSpanContext(t *testing.T) {
	ctx := context.Background()
	span := &mockSpan{recording: true}

	// Test adding span to context
	ctxWithSpan := ContextWithSpan(ctx, span)

	// Test retrieving span from context
	retrieved := SpanFromContext(ctxWithSpan)
	if retrieved != span {
		t.Error("Expected to retrieve the same span")
	}

	// Test retrieving from context without span
	noSpan := SpanFromContext(ctx)
	if noSpan != nil {
		t.Error("Expected nil span from context without span")
	}
}

func TestInitializeAndCleanupHooks(t *testing.T) {
	tracer := &mockTracer{}
	hook := NewTracingHook("test-tracer", tracer)

	agent := &mockAgent{
		id:        "test-agent",
		name:      "TestAgent",
		agentType: domain.AgentTypeLLM,
	}

	ctx := context.Background()

	// Test BeforeInitialize
	newCtx, err := hook.BeforeInitialize(ctx, agent)
	if err != nil {
		t.Errorf("BeforeInitialize failed: %v", err)
	}

	if len(tracer.spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(tracer.spans))
	}

	span := tracer.spans[0]
	if span.name != "agent.TestAgent.initialize" {
		t.Errorf("Expected span name 'agent.TestAgent.initialize', got '%s'", span.name)
	}

	// Test AfterInitialize
	err = hook.AfterInitialize(newCtx, agent, nil)
	if err != nil {
		t.Errorf("AfterInitialize failed: %v", err)
	}

	if !span.ended {
		t.Error("Expected span to be ended")
	}

	// Test cleanup hooks
	tracer.spans = nil
	cleanupCtx, err := hook.BeforeCleanup(ctx, agent)
	if err != nil {
		t.Errorf("BeforeCleanup failed: %v", err)
	}

	span = tracer.spans[0]
	if span.name != "agent.TestAgent.cleanup" {
		t.Errorf("Expected span name 'agent.TestAgent.cleanup', got '%s'", span.name)
	}

	err = hook.AfterCleanup(cleanupCtx, agent, nil)
	if err != nil {
		t.Errorf("AfterCleanup failed: %v", err)
	}

	if !span.ended {
		t.Error("Expected cleanup span to be ended")
	}
}
