// ABOUTME: Tests for the simplified LLMAgent API methods matching Google ADK patterns
// ABOUTME: including constructors with sub-agents, builder methods, and convenience functions

package core

import (
	"context"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestNewLLMAgentWithSubAgents(t *testing.T) {
	// Create mock sub-agents
	subAgent1 := NewBaseAgent("sub1", "Sub Agent 1", domain.AgentTypeLLM)
	subAgent2 := NewBaseAgent("sub2", "Sub Agent 2", domain.AgentTypeLLM)

	// Create main agent with sub-agents
	mockProv := &mockProvider{response: "test response"}
	agent, err := NewLLMAgentWithSubAgents("main", mockProv, subAgent1, subAgent2)

	if err != nil {
		t.Fatalf("Failed to create agent with sub-agents: %v", err)
	}

	// Verify agent was created
	if agent.Name() != "main" {
		t.Errorf("Expected agent name 'main', got %s", agent.Name())
	}

	// Verify sub-agents were added
	subAgents := agent.SubAgents()
	if len(subAgents) != 2 {
		t.Errorf("Expected 2 sub-agents, got %d", len(subAgents))
	}

	// Verify sub-agents are registered as tools
	if _, ok := agent.GetTool("sub1"); !ok {
		t.Error("Sub-agent 'sub1' not registered as tool")
	}
	if _, ok := agent.GetTool("sub2"); !ok {
		t.Error("Sub-agent 'sub2' not registered as tool")
	}

	// Verify transfer_to_agent tool was added
	if _, ok := agent.GetTool("transfer_to_agent"); !ok {
		t.Error("transfer_to_agent tool not registered")
	}
}

func TestNewLLMAgentWithSubAgentsFromString(t *testing.T) {
	// Create mock sub-agents
	subAgent1 := NewBaseAgent("sub1", "Sub Agent 1", domain.AgentTypeLLM)
	subAgent2 := NewBaseAgent("sub2", "Sub Agent 2", domain.AgentTypeLLM)

	// Create main agent with sub-agents from string
	agent, err := NewLLMAgentWithSubAgentsFromString("main", "mock", subAgent1, subAgent2)

	if err != nil {
		t.Fatalf("Failed to create agent with sub-agents from string: %v", err)
	}

	// Verify agent was created
	if agent.Name() != "main" {
		t.Errorf("Expected agent name 'main', got %s", agent.Name())
	}

	// Verify sub-agents were added
	subAgents := agent.SubAgents()
	if len(subAgents) != 2 {
		t.Errorf("Expected 2 sub-agents, got %d", len(subAgents))
	}
}

func TestWithSubAgents(t *testing.T) {
	// Create main agent
	mockProv := &mockProvider{response: "test response"}
	agent := NewAgent("main", mockProv)

	// Create sub-agents
	subAgent1 := NewBaseAgent("sub1", "Sub Agent 1", domain.AgentTypeLLM)
	subAgent2 := NewBaseAgent("sub2", "Sub Agent 2", domain.AgentTypeLLM)

	// Add sub-agents using builder pattern
	agent.WithSubAgents(subAgent1, subAgent2)

	// Verify sub-agents were added
	subAgents := agent.SubAgents()
	if len(subAgents) != 2 {
		t.Errorf("Expected 2 sub-agents, got %d", len(subAgents))
	}

	// Test chaining
	subAgent3 := NewBaseAgent("sub3", "Sub Agent 3", domain.AgentTypeLLM)
	agent.WithSubAgents(subAgent3)

	subAgents = agent.SubAgents()
	if len(subAgents) != 3 {
		t.Errorf("Expected 3 sub-agents after chaining, got %d", len(subAgents))
	}
}

func TestTransferTo(t *testing.T) {
	// Create main agent
	mockProv := &mockProvider{response: "test response"}
	agent := NewAgent("main", mockProv)

	// Create a mock sub-agent that returns specific output
	baseAgent := NewBaseAgent("calculator", "Calculator Agent", domain.AgentTypeLLM)
	subAgent := &mockTransferAgent{
		BaseAgentImpl: baseAgent,
		result:        "42",
	}

	// Add sub-agent
	err := agent.AddSubAgent(subAgent)
	if err != nil {
		t.Fatalf("Failed to add sub-agent: %v", err)
	}

	// Test finding sub-agent works
	found := agent.FindSubAgent("calculator")
	if found == nil {
		t.Fatal("Could not find calculator sub-agent")
	}

	// Test TransferTo with non-existent agent first (doesn't need registry)
	ctx := context.Background()
	_, err = agent.TransferTo(ctx, "non-existent", "", "test")
	if err == nil {
		t.Error("Expected error for non-existent agent")
	}

	// For successful transfer tests, we need to set up the registry
	// Save the current global registry
	originalRegistry := domain.GetGlobalAgentRegistry()

	// Create a new registry to avoid global state issues
	registry := NewAgentRegistry()
	domain.SetGlobalAgentRegistry(registry)
	defer domain.SetGlobalAgentRegistry(originalRegistry) // Restore original

	// Only register the sub-agent to avoid circular dependencies
	// The handoff will look it up by name
	if err := registry.Register(subAgent); err != nil {
		t.Fatal(err)
	}

	// Test TransferTo with string input
	result, err := agent.TransferTo(ctx, "calculator", "Need calculation", "2 + 2")
	if err != nil {
		t.Fatalf("TransferTo failed: %v", err)
	}

	// Verify result
	if output, ok := result.Get("output"); !ok || output != "42" {
		t.Errorf("Expected output '42', got %v", output)
	}

	// Test TransferTo with map input
	mapInput := map[string]interface{}{
		"operation": "multiply",
		"a":         6,
		"b":         7,
	}
	result, err = agent.TransferTo(ctx, "calculator", "Complex calculation", mapInput)
	if err != nil {
		t.Fatalf("TransferTo with map failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result from TransferTo with map input")
	}
}

func TestGetSubAgentByName(t *testing.T) {
	// Create main agent
	mockProv := &mockProvider{response: "test response"}
	agent := NewAgent("main", mockProv)

	// Create and add sub-agents
	subAgent1 := NewBaseAgent("sub1", "Sub Agent 1", domain.AgentTypeLLM)
	subAgent2 := NewBaseAgent("sub2", "Sub Agent 2", domain.AgentTypeLLM)

	if err := agent.AddSubAgent(subAgent1); err != nil {
		t.Fatal(err)
	}
	if err := agent.AddSubAgent(subAgent2); err != nil {
		t.Fatal(err)
	}

	// Test GetSubAgentByName
	found := agent.GetSubAgentByName("sub1")
	if found == nil {
		t.Error("Expected to find sub1")
	}
	if found != nil && found.Name() != "sub1" {
		t.Errorf("Expected agent name 'sub1', got %s", found.Name())
	}

	// Test with non-existent agent
	notFound := agent.GetSubAgentByName("non-existent")
	if notFound != nil {
		t.Error("Expected nil for non-existent agent")
	}
}

func TestDeclarativeAgentCreation(t *testing.T) {
	// Test declarative creation with chaining
	mockProv := &mockProvider{response: "test response"}

	// Create sub-agents
	calculator := NewBaseAgent("calculator", "Math Calculator", domain.AgentTypeLLM)
	researcher := NewBaseAgent("researcher", "Web Researcher", domain.AgentTypeLLM)

	// Create main agent declaratively
	agent := NewAgent("assistant", mockProv).
		SetSystemPrompt("You are a helpful assistant with access to calculator and research tools.").
		WithSubAgents(calculator, researcher).
		WithModel("gpt-4")

	// Verify configuration
	if agent.Name() != "assistant" {
		t.Errorf("Expected agent name 'assistant', got %s", agent.Name())
	}

	if agent.systemPrompt != "You are a helpful assistant with access to calculator and research tools." {
		t.Error("System prompt not set correctly")
	}

	if agent.modelName != "gpt-4" {
		t.Error("Model name not set correctly")
	}

	subAgents := agent.SubAgents()
	if len(subAgents) != 2 {
		t.Errorf("Expected 2 sub-agents, got %d", len(subAgents))
	}
}

// Mock agent for testing transfers
type mockTransferAgent struct {
	*BaseAgentImpl
	result string
}

func (m *mockTransferAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	output := domain.NewState()
	output.Set("output", m.result)

	// Copy input values to demonstrate state passing
	for k, v := range input.Values() {
		output.Set("received_"+k, v)
	}

	return output, nil
}

func (m *mockTransferAgent) ID() string {
	return m.BaseAgentImpl.ID()
}

func (m *mockTransferAgent) Name() string {
	return m.BaseAgentImpl.Name()
}
