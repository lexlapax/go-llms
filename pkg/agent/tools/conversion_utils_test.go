// ABOUTME: Tests for bidirectional agent-tool conversion utilities
// ABOUTME: Validates registry integration, schema mapping, and conversion patterns

package tools

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type mockTool struct {
	name        string
	description string
	execFunc    func(ctx *domain.ToolContext, params interface{}) (interface{}, error)
	schema      *sdomain.Schema
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Description() string { return m.description }
func (m *mockTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, params)
	}
	return params, nil
}
func (m *mockTool) ParameterSchema() *sdomain.Schema { return m.schema }

type mockAgent struct {
	*core.BaseAgentImpl
	runFunc      func(ctx context.Context, state *domain.State) (*domain.State, error)
	inputSchema  *sdomain.Schema
	outputSchema *sdomain.Schema
}

func (m *mockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, state)
	}
	return state, nil
}

func (m *mockAgent) InputSchema() *sdomain.Schema  { return m.inputSchema }
func (m *mockAgent) OutputSchema() *sdomain.Schema { return m.outputSchema }

// Tests

func TestRegisterAgentAsTool(t *testing.T) {
	// Create a mock agent
	agent := &mockAgent{
		BaseAgentImpl: core.NewBaseAgent("test-agent", "Test agent", domain.AgentTypeCustom),
	}

	// Create a registry
	registry := builtins.NewRegistry[domain.Tool]()

	// Register the agent as a tool
	err := RegisterAgentAsTool(agent, registry)
	assert.NoError(t, err)

	// Verify the tool was registered
	tool, ok := registry.Get("test-agent")
	assert.True(t, ok)
	assert.NotNil(t, tool)
	assert.Equal(t, "test-agent", tool.Name())
	assert.Equal(t, "Test agent", tool.Description())
}

func TestRegisterAgentAsToolWithPrefix(t *testing.T) {
	agent := &mockAgent{
		BaseAgentImpl: core.NewBaseAgent("calculator", "Calculator agent", domain.AgentTypeCustom),
	}

	registry := builtins.NewRegistry[domain.Tool]()

	// Register with prefix
	err := RegisterAgentAsTool(agent, registry, ConversionOptions{
		NamePrefix: "agent_",
	})
	assert.NoError(t, err)

	// Check it was registered with prefix
	tool, ok := registry.Get("agent_calculator")
	assert.True(t, ok)
	assert.NotNil(t, tool)
}

func TestConvertToolCategoryToAgents(t *testing.T) {
	// Create registry and add some tools
	registry := builtins.NewRegistry[domain.Tool]()

	tool1 := &mockTool{name: "tool1", description: "Tool 1"}
	tool2 := &mockTool{name: "tool2", description: "Tool 2"}

	err := registry.Register("tool1", tool1, builtins.Metadata{
		Name:     "tool1",
		Category: "test-tools",
		Tags:     []string{"test"},
	})
	require.NoError(t, err)

	err = registry.Register("tool2", tool2, builtins.Metadata{
		Name:     "tool2",
		Category: "test-tools",
		Tags:     []string{"test"},
	})
	require.NoError(t, err)

	// Convert category to agents
	agents, err := ConvertToolCategoryToAgents(registry, "test-tools")
	assert.NoError(t, err)
	assert.Len(t, agents, 2)

	// Verify agents
	for _, agent := range agents {
		assert.NotNil(t, agent)
		assert.Contains(t, []string{"tool1", "tool2"}, agent.Name())
	}
}

func TestNewToolAgentWithEvents(t *testing.T) {
	tool := &mockTool{
		name:        "event-tool",
		description: "Tool with events",
		execFunc: func(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
			if ctx.Events != nil {
				ctx.Events.EmitMessage("Tool executed")
			}
			return "result", nil
		},
	}

	// Create event dispatcher
	dispatcher := core.NewEventDispatcher(10)
	defer dispatcher.Close() // Properly close the dispatcher

	// Create ToolAgent with events
	toolAgent := NewToolAgentWithEvents(tool, dispatcher)
	assert.NotNil(t, toolAgent)

	// Subscribe to events
	var eventReceived atomic.Bool
	dispatcher.Subscribe(domain.EventHandlerFunc(func(event domain.Event) error {
		if event.Type == domain.EventMessage {
			eventReceived.Store(true)
		}
		return nil
	}))

	// Run the agent
	state := domain.NewState()
	state.Set("input", "test")

	result, err := toolAgent.Run(context.Background(), state)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Give some time for event to be processed
	// In real usage, events are processed asynchronously
	assert.Eventually(t, func() bool {
		return eventReceived.Load()
	}, 100*time.Millisecond, 10*time.Millisecond, "Event should have been received")
}

func TestCreateEventForwardingToolContext(t *testing.T) {
	agent := &mockAgent{
		BaseAgentImpl: core.NewBaseAgent("test", "Test", domain.AgentTypeCustom),
	}

	dispatcher := core.NewEventDispatcher(10)
	defer dispatcher.Close() // Properly close the dispatcher

	ctx := context.Background()

	toolCtx := CreateEventForwardingToolContext(ctx, dispatcher, agent, "run-123")
	assert.NotNil(t, toolCtx)
	assert.NotNil(t, toolCtx.Events)

	// Test event emission
	var eventReceived atomic.Bool
	dispatcher.Subscribe(domain.EventHandlerFunc(func(event domain.Event) error {
		if event.Type == domain.EventMessage {
			eventReceived.Store(true)
		}
		return nil
	}))

	toolCtx.Events.EmitMessage("Test message")

	// Give some time for event to be processed
	assert.Eventually(t, func() bool {
		return eventReceived.Load()
	}, 100*time.Millisecond, 10*time.Millisecond)
}

func TestDeriveToolSchemaFromAgent(t *testing.T) {
	inputSchema := &sdomain.Schema{
		Type:        "object",
		Description: "Input for test agent",
		Properties: map[string]sdomain.Property{
			"name": {Type: "string", Description: "Name"},
			"age":  {Type: "integer", Description: "Age"},
		},
		Required: []string{"name"},
	}

	agent := &mockAgent{
		BaseAgentImpl: core.NewBaseAgent("test", "Test", domain.AgentTypeCustom),
		inputSchema:   inputSchema,
	}

	toolSchema := DeriveToolSchemaFromAgent(agent)
	assert.NotNil(t, toolSchema)
	assert.Equal(t, "object", toolSchema.Type)
	assert.Contains(t, toolSchema.Description, "test agent")
	assert.Len(t, toolSchema.Properties, 2)
	assert.Equal(t, []string{"name"}, toolSchema.Required)
}

func TestGenerateSmartMappers(t *testing.T) {
	inputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"input": {Type: "string"},
		},
		Required: []string{"input"},
	}

	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"result": {Type: "string"},
		},
		Required: []string{"result"},
	}

	stateMapper, resultMapper := GenerateSmartMappers(inputSchema, outputSchema)
	assert.NotNil(t, stateMapper)
	assert.NotNil(t, resultMapper)

	// Test state mapper
	params := map[string]interface{}{"input": "test"}
	state, err := stateMapper(context.Background(), params)
	assert.NoError(t, err)
	assert.NotNil(t, state)

	val, exists := state.Get("input")
	assert.True(t, exists)
	assert.Equal(t, "test", val)

	// Test missing required field
	_, err = stateMapper(context.Background(), map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required parameter input not found")
}

func TestWrapLLMAgentAsTool(t *testing.T) {
	agent := &core.LLMAgent{}
	// Initialize the agent properly
	baseAgent := core.NewBaseAgent("llm-agent", "LLM Agent", domain.AgentTypeLLM)
	agent.BaseAgentImpl = baseAgent

	tool := WrapLLMAgentAsTool(agent)
	assert.NotNil(t, tool)
	assert.Equal(t, "llm-agent", tool.Name())

	// Test the state mapper by type asserting
	agentTool, ok := tool.(*AgentTool)
	require.True(t, ok, "Tool should be an AgentTool")

	// Test with string parameter
	state, err := agentTool.stateMapper(context.Background(), "test prompt")
	assert.NoError(t, err)
	val, exists := state.Get("prompt")
	assert.True(t, exists)
	assert.Equal(t, "test prompt", val)

	// Test with map parameter
	state, err = agentTool.stateMapper(context.Background(), map[string]interface{}{
		"query": "test query",
	})
	assert.NoError(t, err)
	val, exists = state.Get("prompt")
	assert.True(t, exists)
	assert.Equal(t, "test query", val)
}

func TestCreateToolChainFromAgents(t *testing.T) {
	// Create test agents
	agent1 := &mockAgent{
		BaseAgentImpl: core.NewBaseAgent("agent1", "Agent 1", domain.AgentTypeCustom),
		runFunc: func(ctx context.Context, state *domain.State) (*domain.State, error) {
			newState := state.Clone()
			newState.Set("step1", "completed")
			return newState, nil
		},
	}

	agent2 := &mockAgent{
		BaseAgentImpl: core.NewBaseAgent("agent2", "Agent 2", domain.AgentTypeCustom),
		runFunc: func(ctx context.Context, state *domain.State) (*domain.State, error) {
			newState := state.Clone()
			newState.Set("step2", "completed")
			return newState, nil
		},
	}

	// Create chain tool
	chainTool := CreateToolChainFromAgents(agent1, agent2)
	assert.NotNil(t, chainTool)
	assert.Contains(t, chainTool.Name(), "chain")

	// Execute the chain
	// Create a minimal agent for the tool context
	minAgent := &mockAgent{
		BaseAgentImpl: core.NewBaseAgent("test", "Test", domain.AgentTypeCustom),
	}
	ctx := domain.NewToolContext(context.Background(), nil, minAgent, "test-run")
	result, err := chainTool.Execute(ctx, map[string]interface{}{})
	assert.NoError(t, err)

	// Check result contains both steps
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "completed", resultMap["step1"])
	assert.Equal(t, "completed", resultMap["step2"])
}

func TestRoundTripConvert(t *testing.T) {
	originalAgent := &mockAgent{
		BaseAgentImpl: core.NewBaseAgent("test-agent", "Test agent", domain.AgentTypeCustom),
	}

	// Do round trip conversion
	resultAgent, err := RoundTripConvert(originalAgent)
	assert.NoError(t, err)
	assert.NotNil(t, resultAgent)

	// Verify properties preserved
	assert.Equal(t, originalAgent.Name(), resultAgent.Name())
	assert.Equal(t, originalAgent.Description(), resultAgent.Description())
}

func TestCreatePathMapper(t *testing.T) {
	paths := map[string]string{
		"userName": "user.name",
		"userAge":  "user.age",
		"city":     "address.city",
	}

	mapper := CreatePathMapper(paths)
	assert.NotNil(t, mapper)

	// Test with nested data
	params := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"age":  30,
		},
		"address": map[string]interface{}{
			"city": "New York",
		},
	}

	state, err := mapper(context.Background(), params)
	assert.NoError(t, err)

	// Check extracted values
	val, exists := state.Get("userName")
	assert.True(t, exists)
	assert.Equal(t, "John", val)

	val, exists = state.Get("userAge")
	assert.True(t, exists)
	assert.Equal(t, 30, val)

	val, exists = state.Get("city")
	assert.True(t, exists)
	assert.Equal(t, "New York", val)
}

func TestCreateTypeConversionMapper(t *testing.T) {
	conversions := map[string]func(interface{}) interface{}{
		"age": func(v interface{}) interface{} {
			// Convert string to int
			if str, ok := v.(string); ok {
				if str == "30" {
					return 30
				}
			}
			return v
		},
		"active": func(v interface{}) interface{} {
			// Convert string to bool
			if str, ok := v.(string); ok {
				return str == "true"
			}
			return v
		},
	}

	mapper := CreateTypeConversionMapper(conversions)

	state, err := mapper(context.Background(), map[string]interface{}{
		"age":    "30",
		"active": "true",
		"name":   "John",
	})
	assert.NoError(t, err)

	// Check conversions
	val, _ := state.Get("age")
	assert.Equal(t, 30, val)

	val, _ = state.Get("active")
	assert.Equal(t, true, val)

	val, _ = state.Get("name")
	assert.Equal(t, "John", val) // Unchanged
}

func TestCreateNestedStateMapper(t *testing.T) {
	// Test flatten mode
	mapper := CreateNestedStateMapper(true)

	params := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"settings": map[string]interface{}{
				"theme": "dark",
			},
		},
	}

	state, err := mapper(context.Background(), params)
	assert.NoError(t, err)

	// Check flattened keys
	val, exists := state.Get("user.name")
	assert.True(t, exists)
	assert.Equal(t, "John", val)

	val, exists = state.Get("user.settings.theme")
	assert.True(t, exists)
	assert.Equal(t, "dark", val)
}

func TestErrorHandling(t *testing.T) {
	t.Run("RegisterAgentAsTool with nil agent", func(t *testing.T) {
		registry := builtins.NewRegistry[domain.Tool]()
		err := RegisterAgentAsTool(nil, registry)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent cannot be nil")
	})

	t.Run("RegisterAgentAsTool with nil registry", func(t *testing.T) {
		agent := &mockAgent{
			BaseAgentImpl: core.NewBaseAgent("test", "Test", domain.AgentTypeCustom),
		}
		err := RegisterAgentAsTool(agent, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "registry cannot be nil")
	})

	t.Run("ConvertToolCategoryToAgents with empty category", func(t *testing.T) {
		registry := builtins.NewRegistry[domain.Tool]()
		agents, err := ConvertToolCategoryToAgents(registry, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no tools found in category")
		assert.Nil(t, agents)
	})

	t.Run("CreateToolChainFromAgents with failing agent", func(t *testing.T) {
		failingAgent := &mockAgent{
			BaseAgentImpl: core.NewBaseAgent("failing", "Failing", domain.AgentTypeCustom),
			runFunc: func(ctx context.Context, state *domain.State) (*domain.State, error) {
				return nil, errors.New("agent failed")
			},
		}

		chainTool := CreateToolChainFromAgents(failingAgent)
		minAgent := &mockAgent{
			BaseAgentImpl: core.NewBaseAgent("test", "Test", domain.AgentTypeCustom),
		}
		ctx := domain.NewToolContext(context.Background(), nil, minAgent, "test")

		_, err := chainTool.Execute(ctx, map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent 0 (failing) failed")
	})
}
