// ABOUTME: Tests for ToolAgent wrapper that exposes tools as agents
// ABOUTME: Verifies parameter extraction, state updates, and lifecycle integration

package tools

import (
	"context"
	"fmt"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolAgent_Basic(t *testing.T) {
	tool := testutils.MockTool{
		ToolName:        "test-tool",
		ToolDescription: "Test tool for testing",
		Executor: func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			if input, ok := params.(string); ok {
				return fmt.Sprintf("processed: %s", input), nil
			}
			return "processed", nil
		},
	}

	agent := NewToolAgent(&tool)

	// Test agent interface
	assert.Equal(t, "test-tool", agent.Name())
	assert.Equal(t, "Test tool for testing", agent.Description())
	assert.Equal(t, domain.AgentTypeCustom, agent.Type())

	// Execute with state
	state := domain.NewState()
	state.Set("input", "test value")

	result, err := agent.Run(context.Background(), state)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check result
	resultVal, exists := result.Get("result")
	assert.True(t, exists)
	assert.Equal(t, "processed: test value", resultVal)
	success, _ := result.Get("success")
	assert.True(t, success.(bool))
}

func TestToolAgent_MapParameters(t *testing.T) {
	tool := testutils.MockTool{
		ToolName: "map-tool",
		Executor: func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			if paramsMap, ok := params.(map[string]interface{}); ok {
				return map[string]interface{}{
					"count": len(paramsMap),
					"keys":  fmt.Sprintf("%v", paramsMap),
				}, nil
			}
			return nil, fmt.Errorf("expected map parameters")
		},
	}

	agent := NewToolAgent(&tool)

	// Execute with multiple state values
	state := domain.NewState()
	state.Set("key1", "value1")
	state.Set("key2", 42)

	result, err := agent.Run(context.Background(), state)
	require.NoError(t, err)

	// Check merged results
	count, _ := result.Get("output_count")
	assert.Equal(t, 2, count)
}

func TestToolAgent_CustomMappers(t *testing.T) {
	tool := testutils.MockTool{
		ToolName: "custom-tool",
		Executor: func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			if paramsMap, ok := params.(map[string]interface{}); ok {
				name := paramsMap["name"]
				age := paramsMap["age"]
				return fmt.Sprintf("User: %v (%v years)", name, age), nil
			}
			return nil, fmt.Errorf("invalid parameters")
		},
	}

	// Custom parameter mapper
	paramMapper := CreateParamMapper(map[string]string{
		"user_name": "name",
		"user_age":  "age",
	})

	// Custom state updater with prefix
	stateUpdater := CreateStateUpdaterWithPrefix("user")

	agent := NewToolAgent(&tool).
		WithParamMapper(paramMapper).
		WithStateUpdater(stateUpdater)

	// Execute
	state := domain.NewState()
	state.Set("user_name", "Bob")
	state.Set("user_age", 25)

	result, err := agent.Run(context.Background(), state)
	require.NoError(t, err)

	// Check prefixed result
	userResult, exists := result.Get("user_result")
	assert.True(t, exists)
	assert.Equal(t, "User: Bob (25 years)", userResult)
	userSuccess, _ := result.Get("user_success")
	assert.True(t, userSuccess.(bool))
}

func TestToolAgent_ErrorHandling(t *testing.T) {
	tool := testutils.MockTool{
		ToolName: "error-tool",
		Executor: func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			return nil, fmt.Errorf("tool execution failed")
		},
	}

	agent := NewToolAgent(&tool)
	state := domain.NewState()
	state.Set("input", "test")

	result, err := agent.Run(context.Background(), state)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool execution failed")
	assert.Nil(t, result)
}

func TestToolAgent_StateUpdateError(t *testing.T) {
	tool := testutils.MockTool{
		ToolName: "success-tool",
		Executor: func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			return "success", nil
		},
	}

	// Custom state updater that always fails
	failingUpdater := func(ctx context.Context, state *domain.State, result interface{}, err error) (*domain.State, error) {
		return nil, fmt.Errorf("state update failed")
	}

	agent := NewToolAgent(&tool).WithStateUpdater(failingUpdater)
	state := domain.NewState()

	_, err := agent.Run(context.Background(), state)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update state")
}

func TestToolAgent_SingleParamMapper(t *testing.T) {
	tool := testutils.MockTool{
		ToolName: "single-param-tool",
		Executor: func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			return fmt.Sprintf("Input was: %v", params), nil
		},
	}

	mapper := CreateSingleParamMapper("message")
	agent := NewToolAgent(&tool).WithParamMapper(mapper)

	// Test with value present
	state := domain.NewState()
	state.Set("message", "Hello, World!")

	result, err := agent.Run(context.Background(), state)
	require.NoError(t, err)

	resultVal, _ := result.Get("result")
	assert.Equal(t, "Input was: Hello, World!", resultVal)

	// Test with missing value
	emptyState := domain.NewState()
	_, err = agent.Run(context.Background(), emptyState)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required state key message not found")
}

func TestDefaultParamMapper_Priority(t *testing.T) {
	tool := testutils.MockTool{
		ToolName: "priority-tool",
		Executor: func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			return params, nil
		},
	}

	agent := NewToolAgent(&tool)

	// Test priority: params > input > full state
	tests := []struct {
		name     string
		setup    func(*domain.State)
		expected interface{}
	}{
		{
			name: "params key takes priority",
			setup: func(s *domain.State) {
				s.Set("params", "from params")
				s.Set("input", "from input")
				s.Set("other", "from other")
			},
			expected: "from params",
		},
		{
			name: "input key when no params",
			setup: func(s *domain.State) {
				s.Set("input", "from input")
				s.Set("other", "from other")
			},
			expected: "from input",
		},
		{
			name: "full state when no params or input",
			setup: func(s *domain.State) {
				s.Set("key1", "value1")
				s.Set("key2", "value2")
			},
			expected: map[string]interface{}{"key1": "value1", "key2": "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := domain.NewState()
			tt.setup(state)

			result, err := agent.Run(context.Background(), state)
			require.NoError(t, err)

			resultVal, _ := result.Get("result")
			assert.Equal(t, tt.expected, resultVal)
		})
	}
}

func TestToolAgent_WithSchema(t *testing.T) {
	schema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"name": {Type: "string"},
		},
	}

	tool := testutils.MockTool{
		ToolName: "schema-tool",
		Schema:   schema,
	}

	agent := NewToolAgent(&tool)

	// The agent should have the tool's schema as input schema
	// Note: This would require adding InputSchema() method to ToolAgent
	// For now, we just verify the agent was created successfully
	assert.NotNil(t, agent)
}
