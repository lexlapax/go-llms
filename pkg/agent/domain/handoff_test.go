// ABOUTME: Tests for the Handoff interface and implementations that enable agent delegation
// ABOUTME: including input transformation, message filtering, and handoff execution

package domain

import (
	"context"
	"testing"
)

// Mock types for testing

type mockAgentRegistry struct {
	agents map[string]BaseAgent
}

func (m *mockAgentRegistry) Register(agent BaseAgent) error {
	m.agents[agent.Name()] = agent
	return nil
}

func (m *mockAgentRegistry) Get(agentID string) (BaseAgent, error) {
	agent, ok := m.agents[agentID]
	if !ok {
		return nil, ErrAgentNotFound
	}
	return agent, nil
}

func (m *mockAgentRegistry) GetByName(name string) (BaseAgent, error) {
	agent, ok := m.agents[name]
	if !ok {
		return nil, ErrAgentNotFound
	}
	return agent, nil
}

func (m *mockAgentRegistry) List() []BaseAgent {
	agents := make([]BaseAgent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents
}

type mockHandoffAgent struct {
	BaseAgent
	name    string
	runFunc func(context.Context, *State) (*State, error)
}

func (m *mockHandoffAgent) Name() string {
	return m.name
}

func (m *mockHandoffAgent) Run(ctx context.Context, state *State) (*State, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, state)
	}
	return state, nil
}

func TestHandoff_BasicOperations(t *testing.T) {
	tests := []struct {
		name        string
		handoff     Handoff
		description string
		targetAgent string
	}{
		{
			name:        "simple handoff",
			handoff:     NewSimpleHandoff("test", "agent-123"),
			description: "Simple handoff to agent-123",
			targetAgent: "agent-123",
		},
		{
			name:        "filtered handoff",
			handoff:     NewFilteredHandoff("test", "agent-456", "key1", "key2"),
			description: "Filtered handoff to agent-456",
			targetAgent: "agent-456",
		},
		{
			name:        "messages only handoff",
			handoff:     NewMessagesOnlyHandoff("test", "agent-789"),
			description: "Messages-only handoff to agent-789",
			targetAgent: "agent-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.handoff.Name() != "test" {
				t.Errorf("Name() = %v, want %v", tt.handoff.Name(), "test")
			}
			if tt.handoff.Description() != tt.description {
				t.Errorf("Description() = %v, want %v", tt.handoff.Description(), tt.description)
			}
			if tt.handoff.TargetAgent() != tt.targetAgent {
				t.Errorf("TargetAgent() = %v, want %v", tt.handoff.TargetAgent(), tt.targetAgent)
			}
		})
	}
}

func TestHandoff_TransformInput(t *testing.T) {
	// Create a state with some values
	state := NewState()
	state.Set("key1", "value1")
	state.Set("key2", "value2")
	state.Set("key3", "value3")

	tests := []struct {
		name           string
		handoff        Handoff
		expectedKeys   []string
		unexpectedKeys []string
	}{
		{
			name:           "simple handoff clones all state",
			handoff:        NewSimpleHandoff("test", "agent-123"),
			expectedKeys:   []string{"key1", "key2", "key3"},
			unexpectedKeys: []string{},
		},
		{
			name:           "filtered handoff only includes specified keys",
			handoff:        NewFilteredHandoff("test", "agent-456", "key1", "key3"),
			expectedKeys:   []string{"key1", "key3"},
			unexpectedKeys: []string{"key2"},
		},
		{
			name:           "messages only handoff excludes values",
			handoff:        NewMessagesOnlyHandoff("test", "agent-789"),
			expectedKeys:   []string{},
			unexpectedKeys: []string{"key1", "key2", "key3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformed := tt.handoff.TransformInput(state)

			// Check expected keys
			for _, key := range tt.expectedKeys {
				if _, ok := transformed.Get(key); !ok {
					t.Errorf("Expected key %s not found in transformed state", key)
				}
			}

			// Check unexpected keys
			for _, key := range tt.unexpectedKeys {
				if _, ok := transformed.Get(key); ok {
					t.Errorf("Unexpected key %s found in transformed state", key)
				}
			}
		})
	}
}

func TestHandoff_FilterMessages(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "hello"},
		{Role: RoleAssistant, Content: "hi"},
		{Role: RoleSystem, Content: "system message"},
		{Role: RoleUser, Content: "question"},
	}

	tests := []struct {
		name          string
		handoff       Handoff
		inputMessages []Message
		expectedCount int
	}{
		{
			name:          "simple handoff returns all messages",
			handoff:       NewSimpleHandoff("test", "agent-123"),
			inputMessages: messages,
			expectedCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.handoff.FilterMessages(tt.inputMessages)
			if len(result) != tt.expectedCount {
				t.Errorf("FilterMessages() returned %d messages, want %d", len(result), tt.expectedCount)
			}
		})
	}
}

func TestHandoff_LastNMessages(t *testing.T) {
	// Create state with messages
	state := NewState()
	for i := 0; i < 10; i++ {
		state.AddMessage(NewMessage(RoleUser, "message"))
	}

	handoff := NewLastNMessagesHandoff("test", "agent-123", 3)

	if handoff.Name() != "test" {
		t.Errorf("Name() = %v, want %v", handoff.Name(), "test")
	}

	if handoff.TargetAgent() != "agent-123" {
		t.Errorf("TargetAgent() = %v, want %v", handoff.TargetAgent(), "agent-123")
	}

	// Transform input should only keep last 3 messages
	transformed := handoff.TransformInput(state)
	if len(transformed.Messages()) != 3 {
		t.Errorf("Expected 3 messages in transformed state, got %d", len(transformed.Messages()))
	}

	// Filter messages should return last 3
	filtered := handoff.FilterMessages(state.Messages())
	if len(filtered) != 3 {
		t.Errorf("Expected 3 filtered messages, got %d", len(filtered))
	}
}

func TestHandoff_Execute(t *testing.T) {
	ctx := context.Background()
	state := NewState()
	state.Set("input", "test input")

	handoff := NewSimpleHandoff("test", "target-agent")

	// Test when global registry is not set
	originalRegistry := GetGlobalAgentRegistry()
	SetGlobalAgentRegistry(nil)
	defer SetGlobalAgentRegistry(originalRegistry)

	_, err := handoff.Execute(ctx, state)
	if err == nil {
		t.Error("Expected error when global registry is nil")
	}
	if err.Error() != "global agent registry not available" {
		t.Errorf("Execute() error = %v, want 'global agent registry not available'", err.Error())
	}

	// Test when target agent is not found
	mockRegistry := &mockAgentRegistry{
		agents: make(map[string]BaseAgent),
	}
	SetGlobalAgentRegistry(mockRegistry)

	_, err = handoff.Execute(ctx, state)
	if err == nil {
		t.Error("Expected error when target agent not found")
	}
	expectedErr := "target agent 'target-agent' not found: agent not found"
	if err.Error() != expectedErr {
		t.Errorf("Execute() error = %v, want %v", err.Error(), expectedErr)
	}

	// Test successful handoff
	mockTargetAgent := &mockHandoffAgent{
		name: "target-agent",
		runFunc: func(ctx context.Context, state *State) (*State, error) {
			result := NewState()
			result.Set("output", "processed")
			return result, nil
		},
	}
	mockRegistry.agents["target-agent"] = mockTargetAgent

	result, err := handoff.Execute(ctx, state)
	if err != nil {
		t.Errorf("Execute() unexpected error = %v", err)
	}
	if output, ok := result.Get("output"); !ok || output != "processed" {
		t.Errorf("Execute() result output = %v, want 'processed'", output)
	}

	// Test when target agent returns error
	mockTargetAgent.runFunc = func(ctx context.Context, state *State) (*State, error) {
		return nil, context.DeadlineExceeded
	}

	_, err = handoff.Execute(ctx, state)
	if err == nil {
		t.Error("Expected error when target agent fails")
	}
	expectedErr = "handoff to agent 'target-agent' failed: context deadline exceeded"
	if err.Error() != expectedErr {
		t.Errorf("Execute() error = %v, want %v", err.Error(), expectedErr)
	}
}

func TestHandoffBuilder(t *testing.T) {
	// Test builder pattern
	handoff := NewHandoffBuilder("custom", "agent-456").
		WithDescription("Custom handoff").
		WithInputFilter(func(s *State) *State {
			newState := NewState()
			newState.Set("filtered", true)
			return newState
		}).
		WithMessageFilter(func(messages []Message) []Message {
			// Filter out system messages
			var filtered []Message
			for _, msg := range messages {
				if msg.Role != RoleSystem {
					filtered = append(filtered, msg)
				}
			}
			return filtered
		}).
		Build()

	if handoff.Name() != "custom" {
		t.Errorf("Name() = %v, want %v", handoff.Name(), "custom")
	}

	if handoff.Description() != "Custom handoff" {
		t.Errorf("Description() = %v, want %v", handoff.Description(), "Custom handoff")
	}

	if handoff.TargetAgent() != "agent-456" {
		t.Errorf("TargetAgent() = %v, want %v", handoff.TargetAgent(), "agent-456")
	}

	// Test input filter
	state := NewState()
	transformed := handoff.TransformInput(state)
	if val, ok := transformed.Get("filtered"); !ok || val != true {
		t.Error("Input filter did not set filtered=true")
	}

	// Test message filter
	messages := []Message{
		{Role: RoleSystem, Content: "system"},
		{Role: RoleUser, Content: "user"},
		{Role: RoleAssistant, Content: "assistant"},
	}
	filtered := handoff.FilterMessages(messages)
	if len(filtered) != 2 {
		t.Errorf("Message filter should have removed system message, got %d messages", len(filtered))
	}
}
