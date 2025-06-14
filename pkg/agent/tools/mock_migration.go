// ABOUTME: Migration utilities for agent and tool mock implementations
// ABOUTME: Provides compatibility layer for transitioning to new mock infrastructure

package tools

import (
	"context"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/fixtures"
	"github.com/lexlapax/go-llms/pkg/testutils/helpers"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// MockAgentCompat compatibility wrapper for old tests
// Deprecated: Use fixtures.SimpleMockAgent() or mocks.MockAgent directly for new code
type MockAgentCompat struct {
	*core.BaseAgentImpl
	runFunc func(ctx context.Context, state *domain.State) (*domain.State, error)

	// Keep reference to new mock for advanced features
	underlying *mocks.MockAgent
}

// NewMockAgentCompat creates a backward-compatible mock agent
// Deprecated: Use fixtures.SimpleMockAgent() for new code
func NewMockAgentCompat(name, description string) *MockAgentCompat {
	// Create using new infrastructure
	underlying := fixtures.SimpleMockAgent()
	underlying.AgentName = name
	underlying.AgentDescription = description

	return &MockAgentCompat{
		BaseAgentImpl: core.NewBaseAgent(name, description, domain.AgentTypeCustom),
		underlying:    underlying,
	}
}

// Run executes the agent (compatibility method)
func (m *MockAgentCompat) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, state)
	}

	// Delegate to new infrastructure
	return m.underlying.Run(ctx, state)
}

// WithRunFunc sets a custom run function
// Deprecated: Use mocks.MockAgent.OnRun for new code
func (m *MockAgentCompat) WithRunFunc(f func(ctx context.Context, state *domain.State) (*domain.State, error)) *MockAgentCompat {
	m.runFunc = f
	m.underlying.OnRun = f
	return m
}

// CreateTestContextCompat creates a tool context for testing
// Deprecated: Use helpers.CreateTestToolContext() for new code
func CreateTestContextCompat() *domain.ToolContext {
	return helpers.CreateTestToolContext()
}

// CreateTestContextWithState creates a tool context with pre-populated state
// Deprecated: Use helpers.CreateToolContextWithState() for new code
func CreateTestContextWithState(data map[string]interface{}) *domain.ToolContext {
	return helpers.CreateToolContextWithState(data)
}

// Migration helpers for common agent patterns

// CreateSimpleTestAgent creates a simple mock agent for testing
func CreateSimpleTestAgent(name string) *MockAgentCompat {
	underlying := fixtures.SimpleMockAgent()
	underlying.AgentName = name

	return &MockAgentCompat{
		BaseAgentImpl: core.NewBaseAgent(name, "Test agent", domain.AgentTypeCustom),
		underlying:    underlying,
	}
}

// CreateResearchTestAgent creates a research mock agent for testing
func CreateResearchTestAgent(name string) *MockAgentCompat {
	underlying := fixtures.ResearchMockAgent()
	underlying.AgentName = name

	return &MockAgentCompat{
		BaseAgentImpl: core.NewBaseAgent(name, "Research test agent", domain.AgentTypeCustom),
		underlying:    underlying,
	}
}

// CreateWorkflowTestAgent creates a workflow mock agent for testing
func CreateWorkflowTestAgent(name string) *MockAgentCompat {
	underlying := fixtures.WorkflowMockAgent()
	underlying.AgentName = name

	return &MockAgentCompat{
		BaseAgentImpl: core.NewBaseAgent(name, "Workflow test agent", domain.AgentTypeCustom),
		underlying:    underlying,
	}
}

// CreateStatefulTestAgent creates a stateful mock agent for testing
func CreateStatefulTestAgent(name string) *MockAgentCompat {
	underlying := fixtures.StatefulMockAgent()
	underlying.AgentName = name

	return &MockAgentCompat{
		BaseAgentImpl: core.NewBaseAgent(name, "Stateful test agent", domain.AgentTypeCustom),
		underlying:    underlying,
	}
}
