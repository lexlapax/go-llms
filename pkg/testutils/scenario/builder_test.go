package scenario

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	llmdomain "github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// MockTestingT is a mock implementation of testing.TB for testing
type MockTestingT struct {
	testing.TB
	errors []string
}

func (m *MockTestingT) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func (m *MockTestingT) Helper() {}

// Implement other required methods
func (m *MockTestingT) Error(args ...interface{})                 {}
func (m *MockTestingT) Fail()                                     {}
func (m *MockTestingT) FailNow()                                  {}
func (m *MockTestingT) Failed() bool                              { return len(m.errors) > 0 }
func (m *MockTestingT) Fatal(args ...interface{})                 {}
func (m *MockTestingT) Fatalf(format string, args ...interface{}) {}
func (m *MockTestingT) Log(args ...interface{})                   {}
func (m *MockTestingT) Logf(format string, args ...interface{})   {}
func (m *MockTestingT) Name() string                              { return "MockTest" }
func (m *MockTestingT) Skip(args ...interface{})                  {}
func (m *MockTestingT) SkipNow()                                  {}
func (m *MockTestingT) Skipf(format string, args ...interface{})  {}
func (m *MockTestingT) Skipped() bool                             { return false }

func TestScenarioBuilder_Basic(t *testing.T) {
	scenario := NewScenario(t)

	// Test fluent API
	scenario.
		WithInput("name", "test").
		WithInput("value", 42).
		WithTimeout(5 * time.Second)

	state := scenario.GetState()
	val, exists := state.Get("name")
	assert.True(t, exists)
	assert.Equal(t, "test", val)

	val, exists = state.Get("value")
	assert.True(t, exists)
	assert.Equal(t, 42, val)
}

func TestScenarioBuilder_WithProvider(t *testing.T) {
	scenario := NewScenario(t)

	responses := map[string]llmdomain.Response{
		"hello": {Content: "Hello response"},
		"world": {Content: "World response"},
	}

	scenario.WithMockProvider("test-provider", responses)

	provider := scenario.GetProvider("test-provider")
	require.NotNil(t, provider)

	// Verify responses were added
	resp, err := provider.GenerateMessage(context.Background(), []llmdomain.Message{
		{Role: llmdomain.RoleUser, Content: []llmdomain.ContentPart{{Type: llmdomain.ContentTypeText, Text: "hello"}}},
	})
	assert.NoError(t, err)
	assert.Equal(t, "Hello response", resp.Content)
}

func TestScenarioBuilder_WithTool(t *testing.T) {
	scenario := NewScenario(t)

	tool := mocks.NewMockTool("calculator", "Performs calculations")
	tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "result: 42"}, nil
	}

	scenario.WithTool(tool)

	// Execute tool
	result, err := scenario.RunTool("calculator", "1+1")
	assert.NoError(t, err)
	expected := map[string]interface{}{"result": "result: 42"}
	assert.Equal(t, expected, result)

	// Verify tool was called
	history := tool.GetCallHistory()
	assert.Len(t, history, 1)
}

func TestScenarioBuilder_WithAgent(t *testing.T) {
	scenario := NewScenario(t)

	agent := mocks.NewMockAgent("test-agent")
	response := domain.NewState()
	response.Set("result", "success")
	agent.AddResponse(response)

	scenario.WithAgent(agent)

	// Run scenario
	finalState := scenario.Run()

	// Verify agent was called
	val, exists := finalState.Get("result")
	assert.True(t, exists)
	assert.Equal(t, "success", val)
	assert.Len(t, agent.GetCallHistory(), 1)
}

func TestScenarioBuilder_Expectations(t *testing.T) {
	t.Run("ExpectOutput", func(t *testing.T) {
		mockT := &MockTestingT{}
		scenario := NewScenario(mockT)

		scenario.
			WithInput("result", "success").
			ExpectOutput("result", Equals("success")).
			ExpectOutput("missing", Equals("value"))

		scenario.Run()

		// Should have one error for missing output
		assert.Len(t, mockT.errors, 1)
		assert.Contains(t, mockT.errors[0], "missing")
	})

	t.Run("ExpectToolCall", func(t *testing.T) {
		scenario := NewScenario(t)

		tool := mocks.NewMockTool("test-tool", "Test tool")
		tool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{"result": "success"}, nil
		}

		scenario.
			WithTool(tool).
			ExpectToolCall("test-tool", HasField("input", Contains("test input data")))

		// Execute tool
		_, _ = scenario.RunTool("test-tool", "test input data")
		scenario.Run()

		// Should pass - tool was called with matching input
	})

	t.Run("ExpectEvent", func(t *testing.T) {
		scenario := NewScenario(t)

		emitter := scenario.GetEventEmitter()

		scenario.ExpectEvent(string(domain.EventAgentStart), HasField("agentName", Equals("test")))

		// Emit event
		emitter.Emit(domain.EventAgentStart, map[string]interface{}{
			"agentName": "test",
		})

		scenario.Run()

		// Should pass - event was emitted
	})

	t.Run("ExpectError", func(t *testing.T) {
		mockT := &MockTestingT{}
		scenario := NewScenario(mockT)

		agent := mocks.NewMockAgent("error-agent")
		agent.AddError(errors.New("test error"))

		scenario.
			WithAgent(agent).
			ExpectError(Contains("test error"))

		scenario.Run()

		// Should pass - error occurred
		assert.Len(t, mockT.errors, 0)

		// Test with no error expected
		scenario2 := NewScenario(mockT)
		scenario2.ExpectError(Contains("different error"))
		scenario2.Run()

		// Should fail - wrong error
		assert.Len(t, mockT.errors, 1)
	})

	t.Run("ExpectNoError", func(t *testing.T) {
		scenario := NewScenario(t)

		agent := mocks.NewMockAgent("success-agent")
		agent.AddResponse(domain.NewState())

		scenario.
			WithAgent(agent).
			ExpectNoError()

		scenario.Run()

		// Should pass - no error occurred
	})
}

func TestScenarioBuilder_ComplexScenario(t *testing.T) {
	scenario := NewScenario(t)

	// Setup provider
	provider := mocks.NewMockProvider("llm")
	provider.WithPatternResponse("analyze", mocks.Response{Content: "Analysis complete"})

	// Setup tools
	searchTool := mocks.NewMockTool("search", "Web search")
	searchTool.WithResponseMapping("query: test", []string{"result1", "result2"})

	calcTool := mocks.NewMockTool("calculator", "Calculator")
	calcTool.OnExecute = func(ctx *domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": 42}, nil
	}

	// Setup agent
	agent := mocks.NewMockAgent("research-agent")
	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		// Simulate agent using tools
		_, _ = scenario.RunTool("search", "query: test")
		calcResult, _ := scenario.RunTool("calculator", map[string]interface{}{"expression": "40+2"})

		newState := state.Clone()
		newState.Set("search_results", []string{"result1", "result2"})
		if res, ok := calcResult.(map[string]interface{}); ok {
			newState.Set("calculation", res["result"])
		}
		newState.Set("analysis", "Analysis complete")

		return newState, nil
	}

	// Build scenario
	scenario.
		WithMockProvider("llm", map[string]llmdomain.Response{
			"analyze": {Content: "Analysis complete"},
		}).
		WithTool(searchTool).
		WithTool(calcTool).
		WithAgent(agent).
		WithInput("query", "test query").
		ExpectOutput("search_results", HasLength(2)).
		ExpectOutput("calculation", Equals(42)).
		ExpectOutput("analysis", Contains("complete")).
		ExpectToolCall("search", HasField("input", Contains("test"))).
		ExpectToolCall("calculator", HasField("expression", IsNotNil())).
		ExpectNoError()

	// Run scenario
	finalState := scenario.Run()

	// Additional assertions
	searchResults, _ := finalState.Get("search_results")
	assert.Equal(t, []string{"result1", "result2"}, searchResults)

	calculation, _ := finalState.Get("calculation")
	assert.Equal(t, 42, calculation)
}

func TestScenarioBuilder_Reset(t *testing.T) {
	scenario := NewScenario(t)

	// Setup initial scenario
	tool := mocks.NewMockTool("tool", "Test tool")
	agent := mocks.NewMockAgent("agent")

	scenario.
		WithTool(tool).
		WithAgent(agent).
		WithInput("key", "value")

	// Execute something
	_, _ = scenario.RunTool("tool", "input")

	// Verify setup
	assert.Len(t, tool.GetCallHistory(), 1)
	val, exists := scenario.GetState().Get("key")
	assert.True(t, exists)
	assert.Equal(t, "value", val)

	// Reset
	scenario.Reset()

	// Verify reset
	assert.Len(t, tool.GetCallHistory(), 0)
	_, exists2 := scenario.GetState().Get("key")
	assert.False(t, exists2)
}

func TestScenarioBuilder_ProviderCall(t *testing.T) {
	scenario := NewScenario(t)

	scenario.WithMockProvider("test-provider", map[string]llmdomain.Response{
		"hello": {Content: "Hello response"},
	})

	// Manually call provider to test expectation
	provider := scenario.GetProvider("test-provider")
	messages := []llmdomain.Message{
		{Role: llmdomain.RoleUser, Content: []llmdomain.ContentPart{{Type: llmdomain.ContentTypeText, Text: "hello"}}},
	}
	_, _ = provider.GenerateMessage(context.Background(), messages)

	scenario.ExpectProviderCall("test-provider", HasLength(1))
	scenario.Run()
}

func TestScenarioBuilder_EventEmitter(t *testing.T) {
	scenario := NewScenario(t)

	// Use custom event emitter
	customEmitter := mocks.NewMockEventEmitter("custom", "custom-emitter")
	scenario.WithEventEmitter(customEmitter)

	// Emit some events
	customEmitter.Emit(domain.EventAgentStart, map[string]interface{}{
		"agentName": "test-agent",
	})
	customEmitter.EmitProgress(50, 100, "50% complete")

	// Verify events
	events := customEmitter.GetEvents()
	assert.Len(t, events, 2)
	assert.Equal(t, domain.EventAgentStart, events[0].Type)
	assert.Equal(t, domain.EventProgress, events[1].Type)
}
