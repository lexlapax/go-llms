// ABOUTME: Testing utilities for agent-tool conversions
// ABOUTME: Provides helpers for validating conversions and creating test doubles

package tools

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ValidationError represents a validation failure in conversion
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// AssertRoundTripEquivalence tests that an agent can be converted to tool and back
func AssertRoundTripEquivalence(t *testing.T, agent domain.BaseAgent) {
	t.Helper()

	// Convert to tool
	tool := NewAgentTool(agent)
	if tool == nil {
		t.Fatal("Failed to convert agent to tool")
	}

	// Convert back to agent
	resultAgent := NewToolAgent(tool)
	if resultAgent == nil {
		t.Fatal("Failed to convert tool back to agent")
	}

	// Test basic properties
	if agent.Name() != resultAgent.Name() {
		t.Errorf("Name mismatch: expected %s, got %s", agent.Name(), resultAgent.Name())
	}

	if agent.Description() != resultAgent.Description() {
		t.Errorf("Description mismatch: expected %s, got %s", agent.Description(), resultAgent.Description())
	}

	// Test execution equivalence
	testState := domain.NewState()
	testState.Set("test", "value")

	// Run original agent
	originalResult, originalErr := agent.Run(context.Background(), testState)

	// Run round-trip agent
	roundTripResult, roundTripErr := resultAgent.Run(context.Background(), testState)

	// Compare errors
	if (originalErr == nil) != (roundTripErr == nil) {
		t.Errorf("Error mismatch: original=%v, roundtrip=%v", originalErr, roundTripErr)
		return
	}

	if originalErr != nil {
		// Both errored, that's consistent
		return
	}

	// Compare results
	if !statesEqual(originalResult, roundTripResult) {
		t.Errorf("Result state mismatch\nOriginal: %v\nRoundTrip: %v",
			originalResult.Values(), roundTripResult.Values())
	}
}

// CreateMockAgentForTool creates a mock agent that behaves like the given tool
func CreateMockAgentForTool(tool domain.Tool) domain.BaseAgent {
	return &mockAgentForTool{
		BaseAgentImpl: core.NewBaseAgent(
			tool.Name(),
			tool.Description(),
			domain.AgentTypeCustom,
		),
		tool: tool,
	}
}

// mockAgentForTool wraps a tool to behave as an agent
type mockAgentForTool struct {
	*core.BaseAgentImpl
	tool domain.Tool
}

func (m *mockAgentForTool) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Create tool context
	toolCtx := domain.NewToolContext(ctx, domain.NewStateReader(state), m, "mock-run")

	// Extract parameters using default mapper
	params, err := DefaultParamMapper(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("failed to extract parameters: %w", err)
	}

	// Execute tool
	result, err := m.tool.Execute(toolCtx, params)

	// Update state with result
	return DefaultStateUpdater(ctx, state.Clone(), result, err)
}

func (m *mockAgentForTool) InputSchema() *sdomain.Schema {
	return m.tool.ParameterSchema()
}

// ValidateAgentToolConversion validates that an agent-tool conversion is valid
func ValidateAgentToolConversion(agent domain.BaseAgent, tool domain.Tool) []ValidationError {
	var errors []ValidationError

	// Validate basic properties
	if agent == nil {
		errors = append(errors, ValidationError{
			Field:   "agent",
			Message: "agent is nil",
		})
		return errors
	}

	if tool == nil {
		errors = append(errors, ValidationError{
			Field:   "tool",
			Message: "tool is nil",
		})
		return errors
	}

	// Check name consistency
	agentTool, isAgentTool := tool.(*AgentTool)
	if isAgentTool && agent.Name() != agentTool.Name() {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("name mismatch: agent=%s, tool=%s", agent.Name(), agentTool.Name()),
		})
	}

	// Check schema compatibility
	if agent.InputSchema() != nil && tool.ParameterSchema() != nil {
		if err := validateSchemaCompatibility(agent.InputSchema(), tool.ParameterSchema()); err != nil {
			errors = append(errors, ValidationError{
				Field:   "schema",
				Message: err.Error(),
			})
		}
	}

	// Test execution compatibility
	testCtx := context.Background()
	testState := domain.NewState()
	testState.Set("test", "validation")

	// Try agent execution
	agentResult, agentErr := agent.Run(testCtx, testState)

	// Try tool execution
	toolCtx := domain.NewToolContext(testCtx, domain.NewStateReader(testState), agent, "validation")
	toolResult, toolErr := tool.Execute(toolCtx, testState.Values())

	// Check error consistency
	if (agentErr == nil) != (toolErr == nil) {
		errors = append(errors, ValidationError{
			Field:   "execution",
			Message: fmt.Sprintf("error consistency: agent error=%v, tool error=%v", agentErr, toolErr),
		})
	}

	// If both succeeded, check result compatibility
	if agentErr == nil && toolErr == nil {
		if !resultsCompatible(agentResult, toolResult) {
			errors = append(errors, ValidationError{
				Field:   "result",
				Message: "execution results are not compatible",
			})
		}
	}

	return errors
}

// Helper functions

func statesEqual(s1, s2 *domain.State) bool {
	if s1 == nil || s2 == nil {
		return s1 == s2
	}

	// Get all values to compare
	values1 := s1.Values()
	values2 := s2.Values()

	if len(values1) != len(values2) {
		return false
	}

	// Check all values
	for key, v1 := range values1 {
		v2, exists := values2[key]
		if !exists {
			return false
		}

		if !reflect.DeepEqual(v1, v2) {
			return false
		}
	}

	return true
}

func validateSchemaCompatibility(agentSchema, toolSchema *sdomain.Schema) error {
	// Basic type check
	if agentSchema.Type != toolSchema.Type {
		return fmt.Errorf("schema type mismatch: agent=%s, tool=%s", agentSchema.Type, toolSchema.Type)
	}

	// For object schemas, check required fields
	if agentSchema.Type == "object" {
		// Tool schema should have at least the same required fields as agent
		requiredMap := make(map[string]bool)
		for _, field := range toolSchema.Required {
			requiredMap[field] = true
		}

		for _, field := range agentSchema.Required {
			if !requiredMap[field] {
				return fmt.Errorf("agent requires field %s but tool schema doesn't", field)
			}
		}
	}

	return nil
}

func resultsCompatible(agentResult *domain.State, toolResult interface{}) bool {
	// If tool result is a map, compare with agent state
	if resultMap, ok := toolResult.(map[string]interface{}); ok {
		// Check if agent state contains the tool result keys
		for key, value := range resultMap {
			agentValue, exists := agentResult.Get(key)
			if !exists {
				// Special case: check for "result" key in agent state
				if result, hasResult := agentResult.Get("result"); hasResult {
					if resultAsMap, ok := result.(map[string]interface{}); ok {
						if resultAsMap[key] != value {
							return false
						}
						continue
					}
				}
				return false
			}

			if !reflect.DeepEqual(agentValue, value) {
				return false
			}
		}
		return true
	}

	// If tool result is not a map, check if it's stored in agent state
	if result, exists := agentResult.Get("result"); exists {
		return reflect.DeepEqual(result, toolResult)
	}

	return false
}

// TestAgentToolBuilder helps build agents and tools for testing
type TestAgentToolBuilder struct {
	name        string
	description string
	runFunc     func(context.Context, *domain.State) (*domain.State, error)
	execFunc    func(*domain.ToolContext, interface{}) (interface{}, error)
	schema      *sdomain.Schema
}

// NewTestAgentToolBuilder creates a new test builder
func NewTestAgentToolBuilder(name string) *TestAgentToolBuilder {
	return &TestAgentToolBuilder{
		name:        name,
		description: fmt.Sprintf("Test %s", name),
	}
}

// WithDescription sets the description
func (b *TestAgentToolBuilder) WithDescription(desc string) *TestAgentToolBuilder {
	b.description = desc
	return b
}

// WithRunFunc sets the agent run function
func (b *TestAgentToolBuilder) WithRunFunc(f func(context.Context, *domain.State) (*domain.State, error)) *TestAgentToolBuilder {
	b.runFunc = f
	return b
}

// WithExecuteFunc sets the tool execute function
func (b *TestAgentToolBuilder) WithExecuteFunc(f func(*domain.ToolContext, interface{}) (interface{}, error)) *TestAgentToolBuilder {
	b.execFunc = f
	return b
}

// WithSchema sets the schema
func (b *TestAgentToolBuilder) WithSchema(schema *sdomain.Schema) *TestAgentToolBuilder {
	b.schema = schema
	return b
}

// BuildAgent builds a test agent
func (b *TestAgentToolBuilder) BuildAgent() domain.BaseAgent {
	agent := &testBuilderAgent{
		BaseAgentImpl: core.NewBaseAgent(b.name, b.description, domain.AgentTypeCustom),
		runFunc:       b.runFunc,
		schema:        b.schema,
	}

	if agent.runFunc == nil {
		agent.runFunc = func(ctx context.Context, state *domain.State) (*domain.State, error) {
			return state, nil
		}
	}

	return agent
}

// BuildTool builds a test tool
func (b *TestAgentToolBuilder) BuildTool() domain.Tool {
	tool := &testBuilderTool{
		name:        b.name,
		description: b.description,
		execFunc:    b.execFunc,
		schema:      b.schema,
	}

	if tool.execFunc == nil {
		tool.execFunc = func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			return params, nil
		}
	}

	return tool
}

// testBuilderAgent is an agent created by the builder
type testBuilderAgent struct {
	*core.BaseAgentImpl
	runFunc func(context.Context, *domain.State) (*domain.State, error)
	schema  *sdomain.Schema
}

func (a *testBuilderAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	return a.runFunc(ctx, state)
}

func (a *testBuilderAgent) InputSchema() *sdomain.Schema { return a.schema }

// testBuilderTool is a tool created by the builder
type testBuilderTool struct {
	name        string
	description string
	execFunc    func(*domain.ToolContext, interface{}) (interface{}, error)
	schema      *sdomain.Schema
}

func (t *testBuilderTool) Name() string                     { return t.name }
func (t *testBuilderTool) Description() string              { return t.description }
func (t *testBuilderTool) ParameterSchema() *sdomain.Schema { return t.schema }

func (t *testBuilderTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	return t.execFunc(ctx, params)
}
