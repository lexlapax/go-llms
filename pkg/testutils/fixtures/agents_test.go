package fixtures

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

func TestSimpleMockAgent(t *testing.T) {
	agent := SimpleMockAgent()
	assert.NotNil(t, agent)
	assert.Equal(t, "simple_agent", agent.Name())

	ctx := context.Background()

	// Test basic run with input state
	inputState := domain.NewState()
	inputState.Set("data", "Hello, agent!")

	result, err := agent.Run(ctx, inputState)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check response content
	message, _ := result.Get("message")
	assert.Equal(t, "Simple agent response", message)

	agentName, _ := result.Get("agent")
	assert.Equal(t, "simple_agent", agentName)

	inputData, _ := result.Get("input_data")
	assert.Equal(t, "Hello, agent!", inputData)
}

func TestResearchMockAgent(t *testing.T) {
	agent := ResearchMockAgent()
	assert.NotNil(t, agent)
	assert.Equal(t, "research_agent", agent.Name())

	ctx := context.Background()

	// Test research task
	inputState := domain.NewState()
	inputState.Set("query", "quantum computing")

	result, err := agent.Run(ctx, inputState)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check response content
	taskType, _ := result.Get("task_type")
	assert.Equal(t, "research", taskType)

	query, _ := result.Get("query")
	assert.Equal(t, "quantum computing", query)

	searchResults, _ := result.Get("search_results")
	assert.NotNil(t, searchResults)

	analysis, _ := result.Get("analysis")
	assert.NotNil(t, analysis)

	summary, _ := result.Get("summary")
	assert.NotNil(t, summary)

	metadata, _ := result.Get("metadata")
	assert.NotNil(t, metadata)
}

func TestWorkflowMockAgent(t *testing.T) {
	agent := WorkflowMockAgent()
	assert.NotNil(t, agent)
	assert.Equal(t, "workflow_agent", agent.Name())

	ctx := context.Background()

	// Test workflow execution
	inputState := domain.NewState()
	inputState.Set("workflow", "data_processing")
	inputState.Set("data", []string{"item1", "item2", "item3"})

	result, err := agent.Run(ctx, inputState)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Check response content
	workflowType, _ := result.Get("type")
	assert.Equal(t, "workflow", workflowType)

	workflowName, _ := result.Get("workflow")
	assert.Equal(t, "data_processing", workflowName)

	status, _ := result.Get("status")
	assert.Equal(t, "completed", status)

	stepsRaw, _ := result.Get("steps")
	steps, ok := stepsRaw.([]map[string]interface{})
	require.True(t, ok, "Steps should be a slice of maps")
	assert.Len(t, steps, 3)
	assert.Equal(t, "validate", steps[0]["name"])
	assert.Equal(t, "process", steps[1]["name"])
	assert.Equal(t, "finalize", steps[2]["name"])
}

func TestStatefulMockAgent(t *testing.T) {
	agent := StatefulMockAgent()
	assert.NotNil(t, agent)
	assert.Equal(t, "stateful_agent", agent.Name())

	ctx := context.Background()

	// Test initial state - increment
	inputState1 := domain.NewState()
	inputState1.Set("command", "increment")

	result1, err := agent.Run(ctx, inputState1)
	assert.NoError(t, err)
	assert.NotNil(t, result1)

	command1, _ := result1.Get("command")
	assert.Equal(t, "increment", command1)

	counter1, _ := result1.Get("counter")
	assert.Equal(t, 1, counter1)

	totalCalls1, _ := result1.Get("total_calls")
	assert.Equal(t, 1, totalCalls1)

	// Test state persistence - increment again
	inputState2 := domain.NewState()
	inputState2.Set("command", "increment")

	result2, err := agent.Run(ctx, inputState2)
	assert.NoError(t, err)
	assert.NotNil(t, result2)

	command2, _ := result2.Get("command")
	assert.Equal(t, "increment", command2)

	counter2, _ := result2.Get("counter")
	assert.Equal(t, 2, counter2)

	totalCalls2, _ := result2.Get("total_calls")
	assert.Equal(t, 2, totalCalls2)

	// Test reset
	inputState3 := domain.NewState()
	inputState3.Set("command", "reset")

	result3, err := agent.Run(ctx, inputState3)
	assert.NoError(t, err)
	assert.NotNil(t, result3)

	command3, _ := result3.Get("command")
	assert.Equal(t, "reset", command3)

	counter3, _ := result3.Get("counter")
	assert.Equal(t, 0, counter3)

	totalCalls3, _ := result3.Get("total_calls")
	assert.Equal(t, 3, totalCalls3) // Call count should not reset
}
