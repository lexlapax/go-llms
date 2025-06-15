// ABOUTME: Comprehensive tests for LLMAgent covering all Phase 1.5 component integrations
// ABOUTME: Validates state-based execution, tool calling, guardrails, and migration functionality

package core

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// Mock types for testing moved to use testutils/mocks infrastructure

// Mock Tool for testing
type mockTool struct {
	name              string
	description       string
	result            any
	err               error
	usageInstructions string
	examples          []domain.ToolExample
	constraints       []string
	errorGuidance     map[string]string
	category          string
	tags              []string
	version           string
	deterministic     bool
	destructive       bool
	requiresConfirm   bool
	latency           string
	paramSchema       *sdomain.Schema
	outputSchema      *sdomain.Schema
}

func (t *mockTool) Name() string                     { return t.name }
func (t *mockTool) Description() string              { return t.description }
func (t *mockTool) ParameterSchema() *sdomain.Schema { return t.paramSchema }
func (t *mockTool) OutputSchema() *sdomain.Schema    { return t.outputSchema }

func (t *mockTool) Execute(ctx *domain.ToolContext, params any) (any, error) {
	if t.err != nil {
		return nil, t.err
	}
	return t.result, nil
}

func (t *mockTool) UsageInstructions() string {
	if t.usageInstructions != "" {
		return t.usageInstructions
	}
	return "Basic usage instructions for " + t.name
}

func (t *mockTool) Examples() []domain.ToolExample {
	if t.examples != nil {
		return t.examples
	}
	return []domain.ToolExample{}
}

func (t *mockTool) Constraints() []string {
	if t.constraints != nil {
		return t.constraints
	}
	return []string{}
}

func (t *mockTool) ErrorGuidance() map[string]string {
	if t.errorGuidance != nil {
		return t.errorGuidance
	}
	return map[string]string{}
}

func (t *mockTool) Category() string {
	if t.category != "" {
		return t.category
	}
	return "test"
}

func (t *mockTool) Tags() []string {
	if t.tags != nil {
		return t.tags
	}
	return []string{"test"}
}

func (t *mockTool) Version() string {
	if t.version != "" {
		return t.version
	}
	return "1.0.0"
}

func (t *mockTool) IsDeterministic() bool {
	return t.deterministic
}

func (t *mockTool) IsDestructive() bool {
	return t.destructive
}

func (t *mockTool) RequiresConfirmation() bool {
	return t.requiresConfirm
}

func (t *mockTool) EstimatedLatency() string {
	if t.latency != "" {
		return t.latency
	}
	return "fast"
}

func (t *mockTool) ToMCPDefinition() domain.MCPToolDefinition {
	return domain.MCPToolDefinition{
		Name:        t.name,
		Description: t.description,
		InputSchema: t.paramSchema,
	}
}

// Test NewAgent factory function (excellent DX)
func TestNewAgent(t *testing.T) {
	provider := mocks.NewMockProvider("test-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Hello"})

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
	provider := mocks.NewMockProvider("test-provider-with-logger")
	provider.WithDefaultResponse(mocks.Response{Content: "Hello"})
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
	provider := mocks.NewMockProvider("test-run-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Hello, World!"})
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
	provider := mocks.NewMockProvider("test-async-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Async response"})
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
	provider := mocks.NewMockProvider("test-tools-provider")
	provider.WithDefaultResponse(mocks.Response{Content: `{"tool": "calculator", "params": {"a": 2, "b": 2}}`})
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
	provider := mocks.NewMockProvider("test-system-prompt-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Response"})
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
	provider := mocks.NewMockProvider("test-model-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Response"})
	agent := NewAgent("test-agent", provider).
		WithModel("gpt-4")

	if agent.modelName != "gpt-4" {
		t.Errorf("Model name not set correctly: %s", agent.modelName)
	}
}

// Test input guardrails
func TestLLMAgent_WithInputGuardrails(t *testing.T) {
	provider := mocks.NewMockProvider("test-guardrails-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Response"})
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
	provider := mocks.NewMockProvider("test-output-guardrails-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Valid response"})
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
	provider := mocks.NewMockProvider("test-transforms-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Transformed response"})

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
	provider := mocks.NewMockProvider("test-handoff-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Response"})
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
	provider := mocks.NewMockProvider("test-tracing-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Response"})
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
	provider := mocks.NewMockProvider("test-event-stream-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Response"})

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
	provider := mocks.NewMockProvider("test-error-provider")
	provider.WithDefaultResponse(mocks.Response{
		Error: fmt.Errorf("provider error"),
	})
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
	provider := mocks.NewMockProvider("test-prompt-extraction")
	provider.WithDefaultResponse(mocks.Response{Content: "test"})
	agent := NewAgent("test-agent", provider)

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
	provider := mocks.NewMockProvider("test-tool-extraction")
	provider.WithDefaultResponse(mocks.Response{Content: "test"})
	agent := NewAgent("test-agent", provider)

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
	provider := mocks.NewMockProvider("benchmark-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Benchmark response"})
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
	provider := mocks.NewMockProvider("benchmark-tools-provider")
	provider.WithDefaultResponse(mocks.Response{Content: `{"tool": "calculator", "params": {"a": 1, "b": 2}}`})
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

// Test enhanced system content generation with full tool metadata
func TestLLMAgent_EnhancedSystemContent(t *testing.T) {
	provider := mocks.NewMockProvider("test-enhanced-system-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Response"})

	// Create a tool with full metadata
	tool := &mockTool{
		name:              "calculator",
		description:       "Perform mathematical calculations",
		usageInstructions: "Use this tool for any mathematical operations. Supports +, -, *, / operations.",
		examples: []domain.ToolExample{
			{
				Name:        "Addition",
				Description: "Add two numbers together",
				Scenario:    "When you need to find the sum of numbers",
				Input:       map[string]any{"operation": "+", "operand1": 5, "operand2": 3},
				Output:      8,
				Explanation: "5 + 3 = 8",
			},
			{
				Name:        "Division",
				Description: "Divide one number by another",
				Scenario:    "When you need to find the quotient",
				Input:       map[string]any{"operation": "/", "operand1": 10, "operand2": 2},
				Output:      5,
				Explanation: "10 / 2 = 5",
			},
		},
		constraints: []string{
			"Division by zero is not allowed",
			"Only numeric operands are supported",
			"Operations limited to +, -, *, /",
		},
		errorGuidance: map[string]string{
			"division_by_zero":  "Cannot divide by zero. Please ensure operand2 is not 0.",
			"invalid_operation": "Operation must be one of: +, -, *, /",
			"invalid_operand":   "Both operands must be numeric values",
		},
		category:        "math",
		tags:            []string{"calculation", "arithmetic"},
		version:         "2.0.0",
		deterministic:   true,
		destructive:     false,
		requiresConfirm: false,
		latency:         "fast",
		paramSchema: &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"operation": {
					Type:        "string",
					Enum:        []string{"+", "-", "*", "/"},
					Description: "The mathematical operation to perform",
				},
				"operand1": {
					Type:        "number",
					Description: "The first operand",
				},
				"operand2": {
					Type:        "number",
					Description: "The second operand",
				},
			},
			Required: []string{"operation", "operand1", "operand2"},
		},
		outputSchema: &sdomain.Schema{
			Type:        "number",
			Description: "The result of the calculation",
		},
	}

	agent := NewAgent("test-agent", provider).
		SetSystemPrompt("You are a helpful math assistant.").
		AddTool(tool)

	// Get system content
	systemContent := agent.getSystemContent()

	// Verify basic content is present
	if !containsString(systemContent, "You are a helpful math assistant.") {
		t.Error("System prompt not found in system content")
	}

	if !containsString(systemContent, "calculator") {
		t.Error("Tool name not found in system content")
	}

	if !containsString(systemContent, "Perform mathematical calculations") {
		t.Error("Tool description not found in system content")
	}

	// Check for enhanced formatting
	if !containsString(systemContent, "## Available Tools") {
		t.Error("Enhanced tool section header not found")
	}

	// Check for enhanced content
	if !containsString(systemContent, "**Usage Instructions:**") {
		t.Error("Usage instructions section not found")
	}

	if !containsString(systemContent, "**Characteristics:**") {
		t.Error("Characteristics section not found")
	}

	if !containsString(systemContent, "**Parameters:**") {
		t.Error("Parameters section not found")
	}

	if !containsString(systemContent, "**Examples:**") {
		t.Error("Examples section not found")
	}

	if !containsString(systemContent, "Division by zero is not allowed") {
		t.Error("Constraints not found in system content")
	}

	if !containsString(systemContent, "### Tool Usage Format") {
		t.Error("Tool usage format section not found")
	}
}

// Test system content with multiple tools
func TestLLMAgent_SystemContentMultipleTools(t *testing.T) {
	provider := mocks.NewMockProvider("test-multiple-tools-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Response"})

	tool1 := &mockTool{
		name:        "tool1",
		description: "First tool for testing",
		category:    "test",
		version:     "1.0.0",
	}

	tool2 := &mockTool{
		name:        "tool2",
		description: "Second tool for testing",
		category:    "test",
		version:     "1.0.0",
	}

	agent := NewAgent("test-agent", provider).
		SetSystemPrompt("Test system prompt").
		AddTool(tool1).
		AddTool(tool2)

	systemContent := agent.getSystemContent()

	// Verify both tools are included
	if !containsString(systemContent, "tool1") {
		t.Error("First tool not found in system content")
	}

	if !containsString(systemContent, "tool2") {
		t.Error("Second tool not found in system content")
	}
}

// Test system content caching
func TestLLMAgent_SystemContentCaching(t *testing.T) {
	provider := mocks.NewMockProvider("test-caching-provider")
	provider.WithDefaultResponse(mocks.Response{Content: "Response"})

	tool := &mockTool{
		name:        "test_tool",
		description: "Test tool",
	}

	agent := NewAgent("test-agent", provider).
		SetSystemPrompt("Initial prompt").
		AddTool(tool)

	// Get system content first time
	content1 := agent.getSystemContent()

	// Get system content second time (should be cached)
	content2 := agent.getSystemContent()

	if content1 != content2 {
		t.Error("System content should be cached and identical")
	}

	// Modify agent (add another tool) - should invalidate cache
	tool2 := &mockTool{
		name:        "another_tool",
		description: "Another test tool",
	}

	agent.AddTool(tool2)

	// Get system content again
	content3 := agent.getSystemContent()

	if content3 == content1 {
		t.Error("System content should change after adding tool")
	}

	if !containsString(content3, "another_tool") {
		t.Error("New tool should be in updated system content")
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && strings.Contains(s, substr)
}
