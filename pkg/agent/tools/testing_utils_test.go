// ABOUTME: Tests for testing utilities
// ABOUTME: Validates the test helpers work correctly

package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/stretchr/testify/assert"
)

func TestAssertRoundTripEquivalence(t *testing.T) {
	// Create a simple agent
	agent := NewTestAgentToolBuilder("test-agent").
		WithDescription("Test agent").
		WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
			result := state.Clone()
			result.Set("processed", true)
			return result, nil
		}).
		BuildAgent()

	// This should pass without errors
	AssertRoundTripEquivalence(t, agent)
}

func TestCreateMockAgentForTool(t *testing.T) {
	// Create a test tool
	tool := NewTestAgentToolBuilder("test-tool").
		WithDescription("Test tool").
		WithExecuteFunc(func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			return map[string]interface{}{
				"status": "success",
				"input":  params,
			}, nil
		}).
		BuildTool()

	// Create mock agent
	mockAgent := CreateMockAgentForTool(tool)

	// Verify properties
	assert.Equal(t, tool.Name(), mockAgent.Name())
	assert.Equal(t, tool.Description(), mockAgent.Description())

	// Test execution
	state := domain.NewState()
	state.Set("test", "value")

	result, err := mockAgent.Run(context.Background(), state)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check result contains expected data
	status, exists := result.Get("status")
	assert.True(t, exists)
	assert.Equal(t, "success", status)
}

func TestValidateAgentToolConversion(t *testing.T) {
	t.Run("Valid conversion", func(t *testing.T) {
		agent := NewTestAgentToolBuilder("valid").
			WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
				return state, nil
			}).
			BuildAgent()

		tool := NewAgentTool(agent)

		errors := ValidateAgentToolConversion(agent, tool)
		assert.Empty(t, errors)
	})

	t.Run("Nil agent", func(t *testing.T) {
		tool := &mockTool{name: "test"}
		errors := ValidateAgentToolConversion(nil, tool)
		assert.Len(t, errors, 1)
		assert.Equal(t, "agent", errors[0].Field)
		assert.Contains(t, errors[0].Message, "nil")
	})

	t.Run("Nil tool", func(t *testing.T) {
		agent := core.NewBaseAgent("test", "Test", domain.AgentTypeCustom)
		errors := ValidateAgentToolConversion(agent, nil)
		assert.Len(t, errors, 1)
		assert.Equal(t, "tool", errors[0].Field)
		assert.Contains(t, errors[0].Message, "nil")
	})

	t.Run("Schema incompatibility", func(t *testing.T) {
		agentSchema := &sdomain.Schema{
			Type:     "object",
			Required: []string{"name", "email"},
		}

		toolSchema := &sdomain.Schema{
			Type:     "object",
			Required: []string{"name"}, // Missing email
		}

		agent := &testAgentWithSchema{
			BaseAgentImpl: core.NewBaseAgent("test", "Test", domain.AgentTypeCustom),
			inputSchema:   agentSchema,
		}

		tool := &testToolWithSchema{
			mockTool: mockTool{name: "test"},
			schema:   toolSchema,
		}

		errors := ValidateAgentToolConversion(agent, tool)
		assert.NotEmpty(t, errors)

		// Find schema error
		var schemaError *ValidationError
		for i := range errors {
			if errors[i].Field == "schema" {
				schemaError = &errors[i]
				break
			}
		}
		assert.NotNil(t, schemaError)
		assert.Contains(t, schemaError.Message, "email")
	})
}

func TestTestAgentToolBuilder(t *testing.T) {
	t.Run("Build agent", func(t *testing.T) {
		runCalled := false
		agent := NewTestAgentToolBuilder("builder-test").
			WithDescription("Built by builder").
			WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
				runCalled = true
				return state, nil
			}).
			BuildAgent()

		assert.Equal(t, "builder-test", agent.Name())
		assert.Equal(t, "Built by builder", agent.Description())

		// Test run
		_, err := agent.Run(context.Background(), domain.NewState())
		assert.NoError(t, err)
		assert.True(t, runCalled)
	})

	t.Run("Build tool", func(t *testing.T) {
		execCalled := false
		tool := NewTestAgentToolBuilder("tool-test").
			WithExecuteFunc(func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
				execCalled = true
				return "result", nil
			}).
			BuildTool()

		assert.Equal(t, "tool-test", tool.Name())

		// Test execute
		// Create a minimal agent for the tool context
		minAgent := core.NewBaseAgent("test", "Test", domain.AgentTypeCustom)
		ctx := domain.NewToolContext(context.Background(), nil, minAgent, "test")
		result, err := tool.Execute(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, "result", result)
		assert.True(t, execCalled)
	})

	t.Run("With schema", func(t *testing.T) {
		schema := &sdomain.Schema{
			Type: "object",
			Properties: map[string]sdomain.Property{
				"test": {Type: "string"},
			},
		}

		agent := NewTestAgentToolBuilder("schema-test").
			WithSchema(schema).
			BuildAgent()

		// Check if agent supports InputSchema
		if schemaAgent, ok := agent.(interface{ InputSchema() *sdomain.Schema }); ok {
			assert.Equal(t, schema, schemaAgent.InputSchema())
		}

		tool := NewTestAgentToolBuilder("schema-test").
			WithSchema(schema).
			BuildTool()

		assert.Equal(t, schema, tool.ParameterSchema())
	})
}

func TestStatesEqual(t *testing.T) {
	t.Run("Equal states", func(t *testing.T) {
		s1 := domain.NewState()
		s1.Set("key1", "value1")
		s1.Set("key2", 42)

		s2 := domain.NewState()
		s2.Set("key1", "value1")
		s2.Set("key2", 42)

		assert.True(t, statesEqual(s1, s2))
	})

	t.Run("Different values", func(t *testing.T) {
		s1 := domain.NewState()
		s1.Set("key", "value1")

		s2 := domain.NewState()
		s2.Set("key", "value2")

		assert.False(t, statesEqual(s1, s2))
	})

	t.Run("Different keys", func(t *testing.T) {
		s1 := domain.NewState()
		s1.Set("key1", "value")

		s2 := domain.NewState()
		s2.Set("key2", "value")

		assert.False(t, statesEqual(s1, s2))
	})

	t.Run("Nil states", func(t *testing.T) {
		assert.True(t, statesEqual(nil, nil))
		assert.False(t, statesEqual(domain.NewState(), nil))
		assert.False(t, statesEqual(nil, domain.NewState()))
	})
}

func TestValidateSchemaCompatibility(t *testing.T) {
	t.Run("Compatible object schemas", func(t *testing.T) {
		agentSchema := &sdomain.Schema{
			Type:     "object",
			Required: []string{"name"},
		}

		toolSchema := &sdomain.Schema{
			Type:     "object",
			Required: []string{"name", "extra"}, // Tool can have extra requirements
		}

		err := validateSchemaCompatibility(agentSchema, toolSchema)
		assert.NoError(t, err)
	})

	t.Run("Type mismatch", func(t *testing.T) {
		agentSchema := &sdomain.Schema{Type: "object"}
		toolSchema := &sdomain.Schema{Type: "array"}

		err := validateSchemaCompatibility(agentSchema, toolSchema)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type mismatch")
	})

	t.Run("Missing required field", func(t *testing.T) {
		agentSchema := &sdomain.Schema{
			Type:     "object",
			Required: []string{"name", "email"},
		}

		toolSchema := &sdomain.Schema{
			Type:     "object",
			Required: []string{"name"}, // Missing email
		}

		err := validateSchemaCompatibility(agentSchema, toolSchema)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
	})
}

func TestResultsCompatible(t *testing.T) {
	t.Run("Map results match", func(t *testing.T) {
		agentState := domain.NewState()
		agentState.Set("key1", "value1")
		agentState.Set("key2", 42)

		toolResult := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		assert.True(t, resultsCompatible(agentState, toolResult))
	})

	t.Run("Tool result in agent result key", func(t *testing.T) {
		agentState := domain.NewState()
		agentState.Set("result", map[string]interface{}{
			"status": "success",
		})

		toolResult := map[string]interface{}{
			"status": "success",
		}

		assert.True(t, resultsCompatible(agentState, toolResult))
	})

	t.Run("Non-map tool result", func(t *testing.T) {
		agentState := domain.NewState()
		agentState.Set("result", "simple string")

		toolResult := "simple string"

		assert.True(t, resultsCompatible(agentState, toolResult))
	})

	t.Run("Incompatible results", func(t *testing.T) {
		agentState := domain.NewState()
		agentState.Set("key", "value1")

		toolResult := map[string]interface{}{
			"key": "value2", // Different value
		}

		assert.False(t, resultsCompatible(agentState, toolResult))
	})
}

// Test helpers

type testAgentWithSchema struct {
	*core.BaseAgentImpl
	inputSchema *sdomain.Schema
}

func (a *testAgentWithSchema) InputSchema() *sdomain.Schema { return a.inputSchema }

func (a *testAgentWithSchema) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	return state, nil
}

type testToolWithSchema struct {
	mockTool
	schema *sdomain.Schema
}

func (t *testToolWithSchema) ParameterSchema() *sdomain.Schema { return t.schema }

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	assert.Equal(t, "test_field: test message", err.Error())
}

// Test error cases in round trip testing
func TestAssertRoundTripEquivalenceWithErrors(t *testing.T) {
	// Create an agent that returns an error
	errorAgent := NewTestAgentToolBuilder("error-agent").
		WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
			return nil, errors.New("test error")
		}).
		BuildAgent()

	// This should handle the error case properly
	// The test passes if both original and round-trip return errors
	AssertRoundTripEquivalence(t, errorAgent)
}

// Test execution error consistency
func TestValidateAgentToolConversionWithExecutionErrors(t *testing.T) {
	// Agent that errors
	errorAgent := NewTestAgentToolBuilder("error-test").
		WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
			return nil, errors.New("agent error")
		}).
		BuildAgent()

	// Tool that succeeds
	successTool := NewTestAgentToolBuilder("success-tool").
		WithExecuteFunc(func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			return "success", nil
		}).
		BuildTool()

	errors := ValidateAgentToolConversion(errorAgent, successTool)

	// Should have execution error
	var execError *ValidationError
	for i := range errors {
		if errors[i].Field == "execution" {
			execError = &errors[i]
			break
		}
	}

	assert.NotNil(t, execError)
	assert.Contains(t, execError.Message, "error consistency")
}
