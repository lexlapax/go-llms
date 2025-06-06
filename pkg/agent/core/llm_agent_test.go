// ABOUTME: Comprehensive tests for LLMAgent covering all Phase 1.5 component integrations
// ABOUTME: Validates state-based execution, tool calling, guardrails, and migration functionality

package core

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	ldomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Mock Provider for testing
type mockProvider struct {
	response string
	err      error
}

func (m *mockProvider) Generate(ctx context.Context, prompt string, options ...ldomain.Option) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *mockProvider) GenerateMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.Response, error) {
	if m.err != nil {
		return ldomain.Response{}, m.err
	}
	return ldomain.Response{Content: m.response}, nil
}

func (m *mockProvider) GenerateWithSchema(ctx context.Context, prompt string, schema *sdomain.Schema, options ...ldomain.Option) (any, error) {
	if m.err != nil {
		return nil, m.err
	}
	return map[string]any{"result": m.response}, nil
}

func (m *mockProvider) Stream(ctx context.Context, prompt string, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	return nil, fmt.Errorf("streaming not implemented in mock")
}

func (m *mockProvider) StreamMessage(ctx context.Context, messages []ldomain.Message, options ...ldomain.Option) (ldomain.ResponseStream, error) {
	return nil, fmt.Errorf("streaming not implemented in mock")
}

// Mock Tool for testing
type mockTool struct {
	name        string
	description string
	result      any
	err         error
}

func (t *mockTool) Name() string                     { return t.name }
func (t *mockTool) Description() string              { return t.description }
func (t *mockTool) ParameterSchema() *sdomain.Schema { return nil }

func (t *mockTool) Execute(ctx *domain.ToolContext, params any) (any, error) {
	if t.err != nil {
		return nil, t.err
	}
	return t.result, nil
}

// Test NewAgent factory function (excellent DX)
func TestNewAgent(t *testing.T) {
	provider := &mockProvider{response: "Hello"}

	agent := NewAgent("test-agent", provider)

	if agent == nil {
		t.Fatal("NewAgent returned nil")
	}

	if agent.Name() != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", agent.Name())
	}

	if agent.Type() != domain.AgentTypeLLM {
		t.Errorf("Expected type LLM, got %s", agent.Type())
	}
}

// Test NewAgentWithLogger factory function
func TestNewAgentWithLogger(t *testing.T) {
	provider := &mockProvider{response: "Hello"}
	logger := slog.Default()

	agent := NewAgentWithLogger("test-agent", provider, logger)

	if agent == nil {
		t.Fatal("NewAgentWithLogger returned nil")
	}

	if agent.deps.Logger != logger {
		t.Error("Logger was not set correctly")
	}
}

// Test basic Run functionality
func TestLLMAgent_Run(t *testing.T) {
	provider := &mockProvider{response: "Hello, World!"}
	agent := NewAgent("test-agent", provider)

	state := domain.NewState()
	state.Set("prompt", "Say hello")

	result, err := agent.Run(context.Background(), state)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	resultValue, exists := result.Get("result")
	if !exists {
		t.Fatal("Result does not contain 'result' key")
	}

	if resultValue != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%v'", resultValue)
	}
}

// Test RunAsync functionality
func TestLLMAgent_RunAsync(t *testing.T) {
	provider := &mockProvider{response: "Async response"}
	agent := NewAgent("test-agent", provider)

	state := domain.NewState()
	state.Set("prompt", "Test async")

	eventChan, err := agent.RunAsync(context.Background(), state)
	if err != nil {
		t.Fatalf("RunAsync failed: %v", err)
	}

	// Wait for completion event
	select {
	case event := <-eventChan:
		if event.Type != domain.EventAgentComplete {
			t.Errorf("Expected complete event, got %s", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for async completion")
	}
}

// Test tool integration
func TestLLMAgent_WithTools(t *testing.T) {
	provider := &mockProvider{response: `{"tool": "calculator", "params": {"a": 2, "b": 2}}`}
	tool := &mockTool{
		name:        "calculator",
		description: "Simple calculator",
		result:      "4",
	}

	agent := NewAgent("test-agent", provider).
		AddTool(tool)

	state := domain.NewState()
	state.Set("prompt", "Calculate 2 + 2")

	result, err := agent.Run(context.Background(), state)
	if err != nil {
		t.Fatalf("Run with tools failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// Verify tool was added
	if len(agent.ListTools()) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(agent.ListTools()))
	}

	if agent.ListTools()[0] != "calculator" {
		t.Errorf("Expected tool 'calculator', got '%s'", agent.ListTools()[0])
	}
}

// Test system prompt configuration
func TestLLMAgent_SetSystemPrompt(t *testing.T) {
	provider := &mockProvider{response: "Response"}
	agent := NewAgent("test-agent", provider).
		SetSystemPrompt("You are a helpful assistant")

	// Verify system prompt is set
	systemContent := agent.getSystemContent()
	if systemContent != "You are a helpful assistant" {
		t.Errorf("System prompt not set correctly: %s", systemContent)
	}
}

// Test model configuration
func TestLLMAgent_WithModel(t *testing.T) {
	provider := &mockProvider{response: "Response"}
	agent := NewAgent("test-agent", provider).
		WithModel("gpt-4")

	if agent.modelName != "gpt-4" {
		t.Errorf("Model name not set correctly: %s", agent.modelName)
	}
}

// Test input guardrails
func TestLLMAgent_WithInputGuardrails(t *testing.T) {
	provider := &mockProvider{response: "Response"}
	guardrail := domain.RequiredKeysGuardrail("input-validation", "prompt")

	agent := NewAgent("test-agent", provider).
		WithInputGuardrails(guardrail)

	// Test with valid state (should pass)
	state := domain.NewState()
	state.Set("prompt", "Valid input")

	result, err := agent.Run(context.Background(), state)
	if err != nil {
		t.Fatalf("Run with valid input failed: %v", err)
	}
	if result == nil {
		t.Fatal("Result is nil")
	}

	// Test with invalid state (should fail guardrails)
	invalidState := domain.NewState()
	// Missing required "prompt" key

	_, err = agent.Run(context.Background(), invalidState)
	if err == nil {
		t.Fatal("Expected guardrails to fail with invalid input")
	}
}

// Test output guardrails
func TestLLMAgent_WithOutputGuardrails(t *testing.T) {
	provider := &mockProvider{response: "Valid response"}
	guardrail := domain.RequiredKeysGuardrail("output-validation", "result")

	agent := NewAgent("test-agent", provider).
		WithOutputGuardrails(guardrail)

	state := domain.NewState()
	state.Set("prompt", "Test output validation")

	result, err := agent.Run(context.Background(), state)
	if err != nil {
		t.Fatalf("Run with output guardrails failed: %v", err)
	}
	if result == nil {
		t.Fatal("Result is nil")
	}
}

// Test state transforms
func TestLLMAgent_WithStateTransforms(t *testing.T) {
	provider := &mockProvider{response: "Transformed response"}

	// Create a simple transform that adds metadata
	inputTransform := func(ctx context.Context, state *domain.State) (*domain.State, error) {
		newState := state.Clone()
		newState.SetMetadata("transformed", true)
		return newState, nil
	}

	agent := NewAgent("test-agent", provider).
		WithInputTransforms(inputTransform)

	state := domain.NewState()
	state.Set("prompt", "Test transform")

	result, err := agent.Run(context.Background(), state)
	if err != nil {
		t.Fatalf("Run with transforms failed: %v", err)
	}

	// Note: Since we use input state for processing, the transform effect
	// would be visible during execution but not necessarily in final result
	if result == nil {
		t.Fatal("Result is nil")
	}
}

// Test handoff capability
func TestLLMAgent_WithHandoff(t *testing.T) {
	provider := &mockProvider{response: "Response"}
	handoff := domain.NewSimpleHandoff("test-handoff", "target-agent")

	agent := NewAgent("test-agent", provider).
		WithHandoff("delegation", handoff)

	// Verify handoff was added
	if len(agent.handoffs) != 1 {
		t.Errorf("Expected 1 handoff, got %d", len(agent.handoffs))
	}

	if agent.handoffs["delegation"] != handoff {
		t.Error("Handoff not set correctly")
	}
}

// Test tracing integration
func TestLLMAgent_WithTracing(t *testing.T) {
	provider := &mockProvider{response: "Response"}
	mockTracer := &mockTracerImpl{}
	tracingHook := NewTracingHook("test-tracer", mockTracer)

	agent := NewAgent("test-agent", provider).
		WithTracing(tracingHook)

	state := domain.NewState()
	state.Set("prompt", "Test tracing")

	result, err := agent.Run(context.Background(), state)
	if err != nil {
		t.Fatalf("Run with tracing failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// Verify tracing hook was called
	if !mockTracer.startCalled {
		t.Error("Tracing hook Start was not called")
	}
}

// Test event stream integration
func TestLLMAgent_WithEventStream(t *testing.T) {
	provider := &mockProvider{response: "Response"}

	eventChan := make(chan domain.Event, 10)
	eventStream := domain.NewFunctionalEventStream(context.Background(), eventChan)

	agent := NewAgent("test-agent", provider).
		WithEventStream(eventStream)

	if agent.eventStream != eventStream {
		t.Error("Event stream not set correctly")
	}
}

// Test error handling
func TestLLMAgent_ErrorHandling(t *testing.T) {
	provider := &mockProvider{err: fmt.Errorf("provider error")}
	agent := NewAgent("test-agent", provider)

	state := domain.NewState()
	state.Set("prompt", "This should fail")

	_, err := agent.Run(context.Background(), state)
	if err == nil {
		t.Fatal("Expected error from provider")
	}

	if err.Error() != "LLM generation failed: provider error" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

// Test prompt extraction strategies
func TestLLMAgent_PromptExtraction(t *testing.T) {
	agent := NewAgent("test-agent", &mockProvider{})

	tests := []struct {
		name     string
		stateKey string
		value    string
		expected string
		wantErr  bool
	}{
		{"prompt key", "prompt", "Hello", "Hello", false},
		{"input key", "input", "Hi there", "Hi there", false},
		{"message key", "message", "Test message", "Test message", false},
		{"query key", "query", "Test query", "Test query", false},
		{"text key", "text", "Test text", "Test text", false},
		{"no valid key", "invalid", "value", "", true},
		{"empty prompt", "prompt", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := domain.NewState()
			state.Set(tt.stateKey, tt.value)

			result, err := agent.extractPromptFromState(state)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Test tool call extraction

func TestLLMAgent_ToolCallExtraction(t *testing.T) {
	agent := NewAgent("test-agent", &mockProvider{})

	tests := []struct {
		name     string
		content  string
		expected []string
		wantCall bool
	}{
		{
			name:     "simple JSON tool call",
			content:  `{"tool": "calculator", "params": {"a": 1, "b": 2}}`,
			expected: []string{"calculator"},
			wantCall: true,
		},
		{
			name:     "OpenAI format tool call",
			content:  `{"tool_calls": [{"id": "call_1", "type": "function", "function": {"name": "weather", "arguments": "{\"city\": \"NYC\"}"}}]}`,
			expected: []string{"weather"},
			wantCall: true,
		},
		{
			name:     "no tool call",
			content:  "Just a regular response",
			expected: nil,
			wantCall: false,
		},
		{
			name:     "markdown JSON block",
			content:  "Here's the tool call:\n```json\n{\"tool\": \"search\", \"params\": {\"query\": \"test\"}}\n```",
			expected: []string{"search"},
			wantCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools, _, found := agent.extractToolCalls(tt.content)

			if found != tt.wantCall {
				t.Errorf("Expected wantCall %v, got %v", tt.wantCall, found)
			}

			if !found {
				return
			}

			if len(tools) != len(tt.expected) {
				t.Errorf("Expected %d tools, got %d", len(tt.expected), len(tools))
				return
			}

			for i, expected := range tt.expected {
				if tools[i] != expected {
					t.Errorf("Expected tool '%s', got '%s'", expected, tools[i])
				}
			}
		})
	}
}

// Mock tracer for testing
type mockTracerImpl struct {
	startCalled bool
}

func (m *mockTracerImpl) Start(ctx context.Context, name string, opts ...SpanOption) (context.Context, Span) {
	m.startCalled = true
	return ctx, &mockTestSpan{}
}

type mockTestSpan struct{}

func (s *mockTestSpan) End()                                          {}
func (s *mockTestSpan) SetAttributes(attributes ...Attribute)         {}
func (s *mockTestSpan) RecordError(err error)                         {}
func (s *mockTestSpan) SetStatus(code StatusCode, description string) {}
func (s *mockTestSpan) IsRecording() bool                             { return true }

// Benchmark tests
func BenchmarkLLMAgent_Run(b *testing.B) {
	provider := &mockProvider{response: "Benchmark response"}
	agent := NewAgent("benchmark-agent", provider)

	state := domain.NewState()
	state.Set("prompt", "Benchmark test")

	b.ResetTimer()
	for range b.N {
		_, err := agent.Run(context.Background(), state)
		if err != nil {
			b.Fatalf("Run failed: %v", err)
		}
	}
}

func BenchmarkLLMAgent_RunWithTools(b *testing.B) {
	provider := &mockProvider{response: `{"tool": "calculator", "params": {"a": 1, "b": 2}}`}
	tool := &mockTool{name: "calculator", result: "3"}

	agent := NewAgent("benchmark-agent", provider).AddTool(tool)

	state := domain.NewState()
	state.Set("prompt", "Calculate 1 + 2")

	b.ResetTimer()
	for range b.N {
		_, err := agent.Run(context.Background(), state)
		if err != nil {
			b.Fatalf("Run with tools failed: %v", err)
		}
	}
}
