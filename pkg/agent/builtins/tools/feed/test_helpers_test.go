// ABOUTME: Test helper functions and mocks for feed tool tests
// ABOUTME: Provides mock agent, event emitter, and test context creation

package feed

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// Helper function to create test tool context
func createTestToolContext() *domain.ToolContext {
	ctx := context.Background()
	state := domain.NewState()
	stateReader := domain.NewStateReader(state)
	events := mocks.NewMockEventEmitter("test-agent", "Test Agent")

	agentInfo := domain.AgentInfo{
		ID:          "test-agent",
		Name:        "Test Agent",
		Description: "A test agent",
		Type:        domain.AgentTypeLLM,
	}

	return &domain.ToolContext{
		Context: ctx,
		State:   stateReader,
		Agent:   agentInfo,
		Events:  events,
		RunID:   "test-run-123",
	}
}
