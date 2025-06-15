// ABOUTME: Tests for LLMAgent sub-agent auto-tool registration feature
// ABOUTME: Verifies that sub-agents are automatically registered as tools when added to an LLMAgent

package core

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

func TestLLMAgent_SubAgentAutoToolRegistration(t *testing.T) {
	// Test basic sub-agent addition
	t.Run("Add single sub-agent", func(t *testing.T) {
		provider := mocks.NewMockProvider("test-provider")
		provider.WithDefaultResponse(mocks.Response{Content: "Response"})
		parentAgent := NewAgent("parent-agent", provider)

		// Create a sub-agent
		subAgent := NewBaseAgent("sub-agent", "A sub-agent", domain.AgentTypeLLM)

		// Add the sub-agent
		err := parentAgent.AddSubAgent(subAgent)
		if err != nil {
			t.Fatalf("Failed to add sub-agent: %v", err)
		}

		// Verify sub-agent was added
		if len(parentAgent.SubAgents()) != 1 {
			t.Errorf("Expected 1 sub-agent, got %d", len(parentAgent.SubAgents()))
		}

		// Verify tools were added (sub-agent + transfer_to_agent)
		tools := parentAgent.ListTools()
		if len(tools) != 2 {
			t.Errorf("Expected 2 tools, got %d: %v", len(tools), tools)
		}

		// Verify specific tools exist
		hasTransferTool := false
		hasSubAgentTool := false
		for _, toolName := range tools {
			if toolName == "transfer_to_agent" {
				hasTransferTool = true
			}
			if toolName == "sub-agent" {
				hasSubAgentTool = true
			}
		}

		if !hasTransferTool {
			t.Error("Expected transfer_to_agent tool to be present")
		}

		if !hasSubAgentTool {
			t.Error("Expected sub-agent tool to be present")
		}
	})

	// Test multiple sub-agents
	t.Run("Add multiple sub-agents", func(t *testing.T) {
		provider := mocks.NewMockProvider("test-provider")
		provider.WithDefaultResponse(mocks.Response{Content: "Response"})
		parentAgent := NewAgent("parent-agent", provider)

		// Add first sub-agent
		subAgent1 := NewBaseAgent("sub-agent-1", "First sub-agent", domain.AgentTypeLLM)
		err := parentAgent.AddSubAgent(subAgent1)
		if err != nil {
			t.Fatalf("Failed to add first sub-agent: %v", err)
		}

		// Add second sub-agent
		subAgent2 := NewBaseAgent("sub-agent-2", "Second sub-agent", domain.AgentTypeLLM)
		err = parentAgent.AddSubAgent(subAgent2)
		if err != nil {
			t.Fatalf("Failed to add second sub-agent: %v", err)
		}

		// Should have 3 tools (transfer_to_agent + 2 sub-agents)
		tools := parentAgent.ListTools()
		if len(tools) != 3 {
			t.Errorf("Expected 3 tools, got %d: %v", len(tools), tools)
		}
	})

	// Test sub-agent removal
	t.Run("Remove sub-agents", func(t *testing.T) {
		provider := mocks.NewMockProvider("test-provider")
		provider.WithDefaultResponse(mocks.Response{Content: "Response"})
		parentAgent := NewAgent("parent-agent", provider)

		// Add two sub-agents
		subAgent1 := NewBaseAgent("sub-agent-1", "First sub-agent", domain.AgentTypeLLM)
		subAgent2 := NewBaseAgent("sub-agent-2", "Second sub-agent", domain.AgentTypeLLM)

		_ = parentAgent.AddSubAgent(subAgent1)
		_ = parentAgent.AddSubAgent(subAgent2)

		// Remove first sub-agent
		err := parentAgent.RemoveSubAgent("sub-agent-1")
		if err != nil {
			t.Fatalf("Failed to remove sub-agent-1: %v", err)
		}

		// Should have 2 tools (transfer_to_agent + sub-agent-2)
		tools := parentAgent.ListTools()
		if len(tools) != 2 {
			t.Errorf("Expected 2 tools after removal, got %d: %v", len(tools), tools)
		}

		// Remove last sub-agent
		err = parentAgent.RemoveSubAgent("sub-agent-2")
		if err != nil {
			t.Fatalf("Failed to remove sub-agent-2: %v", err)
		}

		// Should have no tools (transfer_to_agent should be removed)
		tools = parentAgent.ListTools()
		if len(tools) != 0 {
			t.Errorf("Expected 0 tools after removing all sub-agents, got %d: %v", len(tools), tools)
		}
	})
}
