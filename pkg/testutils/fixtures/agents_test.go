package fixtures

import (
	"context"
	"testing"
	"time"

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

func TestTrackingMockAgent(t *testing.T) {
	agent := TrackingMockAgent("tracker1", 10*time.Millisecond)

	ctx := context.Background()
	state := domain.NewState()
	state.Set("initial_data", "test")

	start := time.Now()
	result, err := agent.Run(ctx, state)
	elapsed := time.Since(start)

	require.NoError(t, err)

	// Verify delay occurred
	assert.GreaterOrEqual(t, elapsed, 10*time.Millisecond)

	// Verify tracking data
	data, exists := result.Get("tracker1_result")
	assert.True(t, exists)
	assert.Equal(t, "data_from_tracker1", data)

	executed, exists := result.Get("tracker1_executed")
	assert.True(t, exists)
	assert.True(t, executed.(bool))

	// Verify accumulated data
	accData, exists := result.Get("accumulated_data")
	assert.True(t, exists)
	assert.Contains(t, accData.([]string), "processed_by_tracker1")
}

func TestSpecialistMockAgent(t *testing.T) {
	tests := []struct {
		name      string
		specialty string
		wantKey   string
	}{
		{
			name:      "data_analysis specialist",
			specialty: "data_analysis",
			wantKey:   "analysis",
		},
		{
			name:      "research specialist",
			specialty: "research",
			wantKey:   "research_findings",
		},
		{
			name:      "development specialist",
			specialty: "development",
			wantKey:   "code_artifacts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := SpecialistMockAgent("specialist", tt.specialty, 50*time.Millisecond)

			ctx := context.Background()
			state := domain.NewState()
			state.Set("task", "analyze customer data")

			result, err := agent.Run(ctx, state)
			require.NoError(t, err)

			// Verify specialty output
			output, _ := result.Get("output")
			assert.Contains(t, output.(string), tt.specialty)

			// Verify specialty-specific data
			_, exists := result.Get(tt.wantKey)
			assert.True(t, exists, "Expected key %s not found", tt.wantKey)
		})
	}
}

func TestErrorSimulationMockAgent(t *testing.T) {
	tests := []struct {
		name          string
		errorType     string
		errorAfter    int
		expectedError string
	}{
		{
			name:          "timeout error",
			errorType:     "timeout",
			errorAfter:    2,
			expectedError: "deadline exceeded",
		},
		{
			name:          "network error",
			errorType:     "network",
			errorAfter:    1,
			expectedError: "network error",
		},
		{
			name:          "rate limit error",
			errorType:     "rate_limit",
			errorAfter:    3,
			expectedError: "rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := ErrorSimulationMockAgent("error-agent", tt.errorType, tt.errorAfter)
			ctx := context.Background()
			state := domain.NewState()

			// Run until error
			var err error
			for i := 0; i < tt.errorAfter; i++ {
				result, runErr := agent.Run(ctx, state)
				if runErr != nil {
					err = runErr
					break
				}
				// Verify successful calls before error
				if i < tt.errorAfter-1 {
					assert.NotNil(t, result)
					callCount, _ := result.Get("call_count")
					assert.Equal(t, i+1, callCount)
				}
			}

			// Verify error occurred
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestStateBuilderMockAgent(t *testing.T) {
	modifications := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": []string{"a", "b", "c"},
	}

	agent := StateBuilderMockAgent("builder1", modifications)

	ctx := context.Background()
	state := domain.NewState()
	state.Set("existing", "data")

	result, err := agent.Run(ctx, state)
	require.NoError(t, err)

	// Verify modifications applied
	for key, expectedValue := range modifications {
		actualValue, exists := result.Get(key)
		assert.True(t, exists)
		assert.Equal(t, expectedValue, actualValue)
	}

	// Verify existing data preserved
	existing, _ := result.Get("existing")
	assert.Equal(t, "data", existing)

	// Verify metadata
	history, _ := result.Get("modification_history")
	assert.Contains(t, history.([]string), "builder1")
}

func TestCoordinatorMockAgent(t *testing.T) {
	agent := CoordinatorMockAgent("coordinator")

	ctx := context.Background()
	state := domain.NewState()
	state.Set("task", "process customer order")

	// First run
	result1, err := agent.Run(ctx, state)
	require.NoError(t, err)

	delegated, _ := result1.Get("delegated")
	assert.True(t, delegated.(bool))

	count1, _ := result1.Get("delegation_count")
	assert.Equal(t, 1, count1)

	plan, _ := result1.Get("delegation_plan")
	assert.Len(t, plan.([]map[string]interface{}), 3)

	// Second run - delegation count should increment
	result2, err := agent.Run(ctx, state)
	require.NoError(t, err)

	count2, _ := result2.Get("delegation_count")
	assert.Equal(t, 2, count2)
}

func TestQualityRefinementMockAgent(t *testing.T) {
	agent := QualityRefinementMockAgent("refiner", 0.3, 0.25)

	ctx := context.Background()
	state := domain.NewState()
	state.Set("content", "Initial content")

	qualities := []float64{}

	// Run multiple iterations
	for i := 0; i < 4; i++ {
		result, err := agent.Run(ctx, state)
		require.NoError(t, err)

		quality, _ := result.Get("quality")
		qualities = append(qualities, quality.(float64))

		iteration, _ := result.Get("iteration")
		assert.Equal(t, i+1, iteration)

		// Update state for next iteration
		state = result
	}

	// Verify quality improvement
	for i := 1; i < len(qualities); i++ {
		assert.Greater(t, qualities[i], qualities[i-1])
	}

	// Verify quality approaches but doesn't exceed 1.0
	assert.LessOrEqual(t, qualities[len(qualities)-1], 1.0)
}

func TestTimeoutMockAgent(t *testing.T) {
	agent := TimeoutMockAgent("timeout-agent", 50*time.Millisecond)

	ctx := context.Background()
	state := domain.NewState()

	start := time.Now()
	_, err := agent.Run(ctx, state)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation timed out")

	// Verify timeout occurred around the expected time
	assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond)
	assert.Less(t, elapsed, 200*time.Millisecond)
}

func TestSharedDataBuilderMockAgent(t *testing.T) {
	ctx := context.Background()

	// Create initial state with some shared data
	state := domain.NewState()
	state.Set("shared_data", map[string]interface{}{
		"existing": "data",
	})

	// Create agent that adds to shared data
	agent1 := SharedDataBuilderMockAgent("builder1", "key1", "value1")
	result1, err := agent1.Run(ctx, state)
	require.NoError(t, err)

	// Verify data was added
	sharedData1, exists := result1.Get("shared_data")
	assert.True(t, exists)
	dataMap1 := sharedData1.(map[string]interface{})
	assert.Equal(t, "data", dataMap1["existing"])
	assert.Equal(t, "value1", dataMap1["key1"])

	// Create another agent that adds more data
	agent2 := SharedDataBuilderMockAgent("builder2", "key2", "value2")
	result2, err := agent2.Run(ctx, result1)
	require.NoError(t, err)

	// Verify both data entries exist
	sharedData2, exists := result2.Get("shared_data")
	assert.True(t, exists)
	dataMap2 := sharedData2.(map[string]interface{})
	assert.Equal(t, "data", dataMap2["existing"])
	assert.Equal(t, "value1", dataMap2["key1"])
	assert.Equal(t, "value2", dataMap2["key2"])
}
