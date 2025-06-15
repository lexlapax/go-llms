// ABOUTME: Tests for AgentTool wrapper that exposes agents as tools
// ABOUTME: Verifies state mapping, result extraction, and error handling

package tools

import (
	"context"
	"fmt"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test ToolContext
func createTestToolContext() *domain.ToolContext {
	state := domain.NewState()
	agent := mocks.NewMockAgent("test")
	return domain.NewToolContext(
		context.Background(),
		domain.NewStateReader(state),
		agent,
		"test-run",
	)
}

func TestAgentTool_Basic(t *testing.T) {
	// Create a mock agent
	agent := mocks.NewMockAgent("test-agent")
	agent.AgentDescription = "Test agent for testing"
	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		input, _ := state.Get("input")
		result := domain.NewState()
		result.Set("result", fmt.Sprintf("processed: %v", input))
		return result, nil
	}

	// Wrap as tool
	tool := NewAgentTool(agent)

	// Test tool interface
	assert.Equal(t, "test-agent", tool.Name())
	assert.Equal(t, "Test agent for testing", tool.Description())

	// Execute with string input
	toolCtx := createTestToolContext()
	result, err := tool.Execute(toolCtx, "test input")
	require.NoError(t, err)
	assert.Equal(t, "processed: test input", result)
}

func TestAgentTool_MapParameters(t *testing.T) {
	agent := mocks.NewMockAgent("map-agent")
	agent.AgentDescription = "Tests map parameters"
	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		// Echo all values
		result := domain.NewState()
		for k, v := range state.Values() {
			result.Set(k, v)
		}
		result.Set("output", "success")
		return result, nil
	}

	tool := NewAgentTool(agent)

	// Execute with map input
	params := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	toolCtx := createTestToolContext()
	result, err := tool.Execute(toolCtx, params)
	require.NoError(t, err)

	// Should return the output
	assert.Equal(t, "success", result)
}

func TestAgentTool_CustomMappers(t *testing.T) {
	agent := mocks.NewMockAgent("custom-mapper")
	agent.AgentDescription = "Tests custom mappers"
	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		name, _ := state.Get("name")
		age, _ := state.Get("age")
		result := domain.NewState()
		result.Set("formatted", fmt.Sprintf("Name: %v, Age: %v", name, age))
		return result, nil
	}

	// Custom state mapper
	stateMapper := CreateStateMapper(map[string]string{
		"user_name": "name",
		"user_age":  "age",
	})

	// Custom result mapper
	resultMapper := CreateResultMapper("formatted")

	tool := NewAgentTool(agent).
		WithStateMapper(stateMapper).
		WithResultMapper(resultMapper)

	// Execute
	params := map[string]interface{}{
		"user_name": "Alice",
		"user_age":  30,
	}
	toolCtx := createTestToolContext()
	result, err := tool.Execute(toolCtx, params)
	require.NoError(t, err)
	assert.Equal(t, "Name: Alice, Age: 30", result)
}

func TestAgentTool_StateInput(t *testing.T) {
	agent := mocks.NewMockAgent("state-agent")
	agent.AgentDescription = "Tests state input"
	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		// Default behavior - echo input
		result := state.Clone()
		result.Set("result", "processed")
		return result, nil
	}

	tool := NewAgentTool(agent)

	// Execute with State input
	inputState := domain.NewState()
	inputState.Set("test", "value")

	toolCtx := createTestToolContext()
	result, err := tool.Execute(toolCtx, inputState)
	require.NoError(t, err)

	// Should get processed result (DefaultResultMapper returns "result" value directly)
	assert.Equal(t, "processed", result)
}

func TestAgentTool_ErrorHandling(t *testing.T) {
	agent := mocks.NewMockAgent("error-agent")
	agent.AgentDescription = "Tests error handling"
	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		return nil, fmt.Errorf("agent execution failed")
	}

	tool := NewAgentTool(agent)

	toolCtx := createTestToolContext()
	_, err := tool.Execute(toolCtx, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent execution failed")
}

func TestAgentTool_ParameterSchema(t *testing.T) {
	agent := mocks.NewMockAgent("schema-agent")
	agent.AgentDescription = "Tests schema"

	// Test with custom schema
	customSchema := &sdomain.Schema{
		Type:        "object",
		Description: "Custom parameters",
		Properties: map[string]sdomain.Property{
			"name": {Type: "string", Description: "User name"},
			"age":  {Type: "integer", Description: "User age"},
		},
		Required: []string{"name"},
	}

	tool := NewAgentTool(agent).WithParameterSchema(customSchema)

	schema := tool.ParameterSchema()
	assert.NotNil(t, schema)
	assert.Equal(t, "Custom parameters", schema.Description)
	assert.Len(t, schema.Properties, 2)

	// Test without custom schema
	tool2 := NewAgentTool(agent)
	schema2 := tool2.ParameterSchema()
	assert.NotNil(t, schema2)
	assert.Equal(t, "object", schema2.Type)
}

func TestAgentTool_MultipleResultFields(t *testing.T) {
	agent := mocks.NewMockAgent("multi-result")
	agent.AgentDescription = "Tests multiple result fields"
	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		result := domain.NewState()
		result.Set("field1", "value1")
		result.Set("field2", "value2")
		result.Set("field3", "value3")
		return result, nil
	}

	// Test extracting multiple fields
	resultMapper := CreateResultMapper("field1", "field3")
	tool := NewAgentTool(agent).WithResultMapper(resultMapper)

	toolCtx := createTestToolContext()
	result, err := tool.Execute(toolCtx, nil)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "value1", resultMap["field1"])
	assert.Equal(t, "value3", resultMap["field3"])
	assert.NotContains(t, resultMap, "field2")
}

func TestDefaultResultMapper_CommonKeys(t *testing.T) {
	tests := []struct {
		name     string
		stateKey string
		value    interface{}
	}{
		{"result key", "result", "test result"},
		{"output key", "output", "test output"},
		{"response key", "response", "test response"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := domain.NewState()
			state.Set(tt.stateKey, tt.value)

			result, err := DefaultResultMapper(context.Background(), state)
			require.NoError(t, err)
			assert.Equal(t, tt.value, result)
		})
	}
}

func TestCreateStateMapper_UnmappedFields(t *testing.T) {
	mapper := CreateStateMapper(map[string]string{
		"old1": "new1",
		"old2": "new2",
	})

	params := map[string]interface{}{
		"old1":     "value1",
		"old2":     "value2",
		"unmapped": "value3",
	}

	state, err := mapper(context.Background(), params)
	require.NoError(t, err)

	// Mapped fields
	val1, _ := state.Get("new1")
	assert.Equal(t, "value1", val1)
	val2, _ := state.Get("new2")
	assert.Equal(t, "value2", val2)

	// Unmapped field should still be included
	val3, _ := state.Get("unmapped")
	assert.Equal(t, "value3", val3)
}
