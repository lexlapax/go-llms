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

// MockAgentCompat provides a compatibility wrapper for old tests.
// It bridges the gap between legacy test code and the new mock infrastructure.
//
// Deprecated: Use fixtures.SimpleMockAgent() or mocks.MockAgent directly for new code.
type MockAgentCompat struct {
	*core.BaseAgentImpl
	runFunc func(ctx context.Context, state *domain.State) (*domain.State, error)

	// Keep reference to new mock for advanced features
	underlying *mocks.MockAgent
}

// NewMockAgentCompat creates a backward-compatible mock agent.
// This function helps migrate existing tests to the new infrastructure.
//
// Parameters:
//   - name: The agent name
//   - description: The agent description
//
// Returns a MockAgentCompat instance.
//
// Deprecated: Use fixtures.SimpleMockAgent() for new code.
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

// Run executes the agent (compatibility method).
// It uses the custom run function if set, otherwise delegates to the underlying mock.
//
// Parameters:
//   - ctx: The execution context
//   - state: The input state
//
// Returns the result state or an error.
func (m *MockAgentCompat) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, state)
	}

	// Delegate to new infrastructure
	return m.underlying.Run(ctx, state)
}

// WithRunFunc sets a custom run function.
// This allows customizing the agent's behavior for specific tests.
//
// Parameters:
//   - f: The custom run function
//
// Returns the MockAgentCompat for method chaining.
//
// Deprecated: Use mocks.MockAgent.OnRun for new code.
func (m *MockAgentCompat) WithRunFunc(f func(ctx context.Context, state *domain.State) (*domain.State, error)) *MockAgentCompat {
	m.runFunc = f
	m.underlying.OnRun = f
	return m
}

// CreateTestContextCompat creates a tool context for testing.
// This provides a simple way to create test contexts.
//
// Returns a new ToolContext for testing.
//
// Deprecated: Use helpers.CreateTestToolContext() for new code.
func CreateTestContextCompat() *domain.ToolContext {
	return helpers.CreateTestToolContext()
}

// CreateTestContextWithState creates a tool context with pre-populated state.
// This is useful for tests that need specific state values.
//
// Parameters:
//   - data: Initial state data as key-value pairs
//
// Returns a new ToolContext with the specified state.
//
// Deprecated: Use helpers.CreateToolContextWithState() for new code.
func CreateTestContextWithState(data map[string]interface{}) *domain.ToolContext {
	return helpers.CreateToolContextWithState(data)
}

// Migration helpers for common agent patterns

// CreateSimpleTestAgent creates a simple mock agent for testing.
// This provides a basic agent with minimal configuration.
//
// Parameters:
//   - name: The agent name
//
// Returns a MockAgentCompat configured as a simple test agent.
func CreateSimpleTestAgent(name string) *MockAgentCompat {
	underlying := fixtures.SimpleMockAgent()
	underlying.AgentName = name

	return &MockAgentCompat{
		BaseAgentImpl: core.NewBaseAgent(name, "Test agent", domain.AgentTypeCustom),
		underlying:    underlying,
	}
}

// CreateResearchTestAgent creates a research mock agent for testing.
// This provides an agent configured for research-style operations.
//
// Parameters:
//   - name: The agent name
//
// Returns a MockAgentCompat configured as a research test agent.
func CreateResearchTestAgent(name string) *MockAgentCompat {
	underlying := fixtures.ResearchMockAgent()
	underlying.AgentName = name

	return &MockAgentCompat{
		BaseAgentImpl: core.NewBaseAgent(name, "Research test agent", domain.AgentTypeCustom),
		underlying:    underlying,
	}
}

// CreateWorkflowTestAgent creates a workflow mock agent for testing.
// This provides an agent configured for workflow operations.
//
// Parameters:
//   - name: The agent name
//
// Returns a MockAgentCompat configured as a workflow test agent.
func CreateWorkflowTestAgent(name string) *MockAgentCompat {
	underlying := fixtures.WorkflowMockAgent()
	underlying.AgentName = name

	return &MockAgentCompat{
		BaseAgentImpl: core.NewBaseAgent(name, "Workflow test agent", domain.AgentTypeCustom),
		underlying:    underlying,
	}
}

// CreateStatefulTestAgent creates a stateful mock agent for testing.
// This provides an agent that maintains state between operations.
//
// Parameters:
//   - name: The agent name
//
// Returns a MockAgentCompat configured as a stateful test agent.
func CreateStatefulTestAgent(name string) *MockAgentCompat {
	underlying := fixtures.StatefulMockAgent()
	underlying.AgentName = name

	return &MockAgentCompat{
		BaseAgentImpl: core.NewBaseAgent(name, "Stateful test agent", domain.AgentTypeCustom),
		underlying:    underlying,
	}
}
