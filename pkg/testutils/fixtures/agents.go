// ABOUTME: Pre-configured mock agents for common testing scenarios
// ABOUTME: Provides ready-to-use agent fixtures with typical behavior patterns

package fixtures

import (
	"context"
	"fmt"

	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// SimpleMockAgent creates a basic mock agent for simple testing
func SimpleMockAgent() *mocks.MockAgent {
	agent := mocks.NewMockAgent("simple_agent")

	// Create a simple response state
	simpleState := domain.NewState()
	simpleState.Set("message", "Simple agent response")
	simpleState.Set("status", "success")
	agent.AddResponse(simpleState)

	// Set up simple execution logic using OnRun
	agent.OnRun = func(ctx context.Context, input *domain.State) (*domain.State, error) {
		response := domain.NewState()
		response.Set("message", "Simple agent response")
		response.Set("agent", "simple_agent")

		// Echo back input data if available
		if input != nil {
			if inputData, exists := input.Get("data"); exists && inputData != nil {
				response.Set("input_data", inputData)
			}
		}

		return response, nil
	}

	return agent
}

// ResearchMockAgent creates a mock agent that simulates research workflows
func ResearchMockAgent() *mocks.MockAgent {
	agent := mocks.NewMockAgent("research_agent")

	agent.OnRun = func(ctx context.Context, input *domain.State) (*domain.State, error) {
		query := "default research query"
		if input != nil {
			if queryData, exists := input.Get("query"); exists && queryData != nil {
				query = fmt.Sprintf("%v", queryData)
			}
		}

		// Create research results state
		results := domain.NewState()
		results.Set("task_type", "research")
		results.Set("query", query)

		// Set search results
		searchResults := []map[string]interface{}{
			{
				"title":   "Research Result 1",
				"summary": "Comprehensive overview of " + query,
				"source":  "academic.example.com",
			},
			{
				"title":   "Research Result 2",
				"summary": "In-depth analysis of " + query,
				"source":  "research.example.com",
			},
		}
		results.Set("search_results", searchResults)

		// Set analysis
		analysis := map[string]interface{}{
			"key_points": []string{
				"Key finding 1 about " + query,
				"Important insight 2 about " + query,
				"Critical observation 3 about " + query,
			},
			"confidence": 0.85,
		}
		results.Set("analysis", analysis)
		results.Set("summary", "Based on comprehensive research, "+query+" shows significant potential with multiple applications.")

		// Set metadata
		metadata := map[string]interface{}{
			"research_time": "2.5 seconds",
			"sources_count": 2,
			"confidence":    0.85,
		}
		results.Set("metadata", metadata)

		return results, nil
	}

	return agent
}

// WorkflowMockAgent creates a mock agent that simulates workflow execution
func WorkflowMockAgent() *mocks.MockAgent {
	agent := mocks.NewMockAgent("workflow_agent")

	agent.OnRun = func(ctx context.Context, input *domain.State) (*domain.State, error) {
		workflowType := "default"
		if input != nil {
			if workflowData, exists := input.Get("workflow"); exists && workflowData != nil {
				workflowType = fmt.Sprintf("%v", workflowData)
			}
		}

		// Create workflow result state
		result := domain.NewState()
		result.Set("type", "workflow")
		result.Set("workflow", workflowType)
		result.Set("status", "completed")
		result.Set("total_time", "0.8s")

		// Simulate different workflow types
		var steps []map[string]interface{}

		switch workflowType {
		case "data_processing":
			steps = []map[string]interface{}{
				{
					"name":        "validate",
					"description": "Validate input data",
					"status":      "completed",
					"duration":    "0.1s",
				},
				{
					"name":        "process",
					"description": "Process validated data",
					"status":      "completed",
					"duration":    "0.5s",
				},
				{
					"name":        "finalize",
					"description": "Finalize processed results",
					"status":      "completed",
					"duration":    "0.2s",
				},
			}
		case "analysis":
			steps = []map[string]interface{}{
				{
					"name":        "collect",
					"description": "Collect data for analysis",
					"status":      "completed",
					"duration":    "0.3s",
				},
				{
					"name":        "analyze",
					"description": "Perform statistical analysis",
					"status":      "completed",
					"duration":    "1.2s",
				},
				{
					"name":        "report",
					"description": "Generate analysis report",
					"status":      "completed",
					"duration":    "0.4s",
				},
			}
		default:
			steps = []map[string]interface{}{
				{
					"name":        "execute",
					"description": "Execute default workflow",
					"status":      "completed",
					"duration":    "0.1s",
				},
			}
		}

		result.Set("steps", steps)

		return result, nil
	}

	return agent
}

// StatefulMockAgent creates a mock agent that maintains internal state
func StatefulMockAgent() *mocks.MockAgent {
	agent := mocks.NewMockAgent("stateful_agent")

	// Internal state - using closure to maintain state
	counter := 0
	totalCalls := 0

	agent.OnRun = func(ctx context.Context, input *domain.State) (*domain.State, error) {
		totalCalls++

		command := "get"
		if input != nil {
			if cmdData, exists := input.Get("command"); exists && cmdData != nil {
				command = fmt.Sprintf("%v", cmdData)
			}
		}

		switch command {
		case "increment":
			counter++
		case "decrement":
			counter--
		case "reset":
			counter = 0
		case "get":
			// Just return current state
		default:
			// Treat unknown commands as increment
			counter++
		}

		// Create response state
		response := domain.NewState()
		response.Set("command", command)
		response.Set("counter", counter)
		response.Set("total_calls", totalCalls)

		// Set nested state info
		stateInfo := map[string]interface{}{
			"counter":     counter,
			"total_calls": totalCalls,
		}
		response.Set("state", stateInfo)

		return response, nil
	}

	return agent
}
