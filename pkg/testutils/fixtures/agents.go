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

// ComplexWorkflowMockAgent creates a mock agent that simulates complex workflow execution
func ComplexWorkflowMockAgent() *mocks.MockAgent {
	agent := mocks.NewMockAgent("complex_workflow_agent")

	agent.OnRun = func(ctx context.Context, input *domain.State) (*domain.State, error) {
		workflowType := "multi_step"
		if input != nil {
			if workflowData, exists := input.Get("workflow_type"); exists && workflowData != nil {
				workflowType = fmt.Sprintf("%v", workflowData)
			}
		}

		// Create comprehensive workflow result state
		result := domain.NewState()
		result.Set("type", "complex_workflow")
		result.Set("workflow_type", workflowType)
		result.Set("status", "completed")
		result.Set("total_duration", "5.2s")

		// Simulate complex workflow with multiple phases
		var phases []map[string]interface{}

		switch workflowType {
		case "research_analysis":
			phases = []map[string]interface{}{
				{
					"phase":       "data_collection",
					"description": "Gather research data from multiple sources",
					"status":      "completed",
					"duration":    "1.2s",
					"tools_used":  []string{"web_search", "document_reader"},
					"output_size": 1024,
				},
				{
					"phase":       "analysis",
					"description": "Analyze collected data for patterns",
					"status":      "completed",
					"duration":    "2.1s",
					"tools_used":  []string{"data_processor", "statistical_analyzer"},
					"output_size": 2048,
				},
				{
					"phase":       "synthesis",
					"description": "Synthesize findings into report",
					"status":      "completed",
					"duration":    "1.5s",
					"tools_used":  []string{"report_generator", "formatter"},
					"output_size": 4096,
				},
				{
					"phase":       "validation",
					"description": "Validate results and conclusions",
					"status":      "completed",
					"duration":    "0.4s",
					"tools_used":  []string{"validator", "fact_checker"},
					"output_size": 512,
				},
			}

		case "content_creation":
			phases = []map[string]interface{}{
				{
					"phase":       "planning",
					"description": "Plan content structure and outline",
					"status":      "completed",
					"duration":    "0.8s",
					"tools_used":  []string{"outliner", "planner"},
					"output_size": 256,
				},
				{
					"phase":       "research",
					"description": "Research topic and gather references",
					"status":      "completed",
					"duration":    "1.8s",
					"tools_used":  []string{"web_search", "reference_manager"},
					"output_size": 1536,
				},
				{
					"phase":       "writing",
					"description": "Generate content based on plan and research",
					"status":      "completed",
					"duration":    "2.2s",
					"tools_used":  []string{"content_generator", "editor"},
					"output_size": 8192,
				},
				{
					"phase":       "review",
					"description": "Review and refine generated content",
					"status":      "completed",
					"duration":    "0.4s",
					"tools_used":  []string{"grammar_checker", "style_checker"},
					"output_size": 512,
				},
			}

		case "problem_solving":
			phases = []map[string]interface{}{
				{
					"phase":       "problem_definition",
					"description": "Define and understand the problem scope",
					"status":      "completed",
					"duration":    "0.6s",
					"tools_used":  []string{"problem_analyzer"},
					"output_size": 384,
				},
				{
					"phase":       "solution_generation",
					"description": "Generate multiple solution approaches",
					"status":      "completed",
					"duration":    "1.8s",
					"tools_used":  []string{"brainstormer", "solution_generator"},
					"output_size": 2048,
				},
				{
					"phase":       "evaluation",
					"description": "Evaluate and rank potential solutions",
					"status":      "completed",
					"duration":    "1.2s",
					"tools_used":  []string{"evaluator", "ranker"},
					"output_size": 1024,
				},
				{
					"phase":       "implementation_plan",
					"description": "Create implementation plan for best solution",
					"status":      "completed",
					"duration":    "1.6s",
					"tools_used":  []string{"planner", "timeline_generator"},
					"output_size": 1536,
				},
			}

		default:
			phases = []map[string]interface{}{
				{
					"phase":       "initialization",
					"description": "Initialize workflow components",
					"status":      "completed",
					"duration":    "0.3s",
					"tools_used":  []string{"initializer"},
					"output_size": 128,
				},
				{
					"phase":       "processing",
					"description": "Process input through workflow",
					"status":      "completed",
					"duration":    "2.0s",
					"tools_used":  []string{"processor", "transformer"},
					"output_size": 1024,
				},
				{
					"phase":       "finalization",
					"description": "Finalize workflow output",
					"status":      "completed",
					"duration":    "0.5s",
					"tools_used":  []string{"finalizer"},
					"output_size": 256,
				},
			}
		}

		result.Set("phases", phases)

		// Set aggregate metrics
		totalOutputSize := 0
		for _, phase := range phases {
			if size, ok := phase["output_size"].(int); ok {
				totalOutputSize += size
			}
		}

		metrics := map[string]interface{}{
			"total_phases":   len(phases),
			"total_tools":    getUniqueToolCount(phases),
			"total_output":   totalOutputSize,
			"success_rate":   1.0,
			"efficiency":     "high",
			"resource_usage": "moderate",
		}
		result.Set("metrics", metrics)

		// Set final summary
		summary := map[string]interface{}{
			"workflow_completed": true,
			"quality_score":      0.95,
			"recommendations":    []string{"Workflow executed successfully", "All phases completed within expected timeframes"},
		}
		result.Set("summary", summary)

		return result, nil
	}

	return agent
}

// ConcurrentMockAgent creates a mock agent that simulates concurrent operations
func ConcurrentMockAgent() *mocks.MockAgent {
	agent := mocks.NewMockAgent("concurrent_agent")

	// Simulate concurrent execution tracking
	var executionCounter int64 = 0

	agent.OnRun = func(ctx context.Context, input *domain.State) (*domain.State, error) {
		// Increment counter atomically (simplified for mock)
		executionCounter++
		currentExecution := executionCounter

		// Get concurrency level from input
		concurrencyLevel := 1
		if input != nil {
			if level, exists := input.Get("concurrency_level"); exists {
				if levelInt, ok := level.(int); ok {
					concurrencyLevel = levelInt
				}
			}
		}

		// Create result state
		result := domain.NewState()
		result.Set("execution_id", currentExecution)
		result.Set("agent_type", "concurrent")
		result.Set("concurrency_level", concurrencyLevel)
		result.Set("status", "completed")

		// Simulate concurrent operations
		operations := make([]map[string]interface{}, concurrencyLevel)
		for i := 0; i < concurrencyLevel; i++ {
			operations[i] = map[string]interface{}{
				"operation_id": fmt.Sprintf("op_%d_%d", currentExecution, i+1),
				"status":       "completed",
				"duration":     fmt.Sprintf("0.%ds", (i%5)+1),
				"result":       fmt.Sprintf("Result from operation %d", i+1),
				"thread_id":    fmt.Sprintf("thread_%d", i+1),
			}
		}

		result.Set("operations", operations)
		result.Set("total_operations", len(operations))

		// Set execution metadata
		metadata := map[string]interface{}{
			"execution_time":  "1.2s",
			"memory_usage":    "moderate",
			"cpu_utilization": "75%",
			"success_rate":    100.0,
		}
		result.Set("execution_metadata", metadata)

		return result, nil
	}

	return agent
}

// ErrorRecoveryMockAgent creates a mock agent that simulates error recovery scenarios
func ErrorRecoveryMockAgent() *mocks.MockAgent {
	agent := mocks.NewMockAgent("error_recovery_agent")

	// Track recovery attempts
	var recoveryAttempts = 0

	agent.OnRun = func(ctx context.Context, input *domain.State) (*domain.State, error) {
		// Get error simulation settings
		shouldSimulateError := false
		errorType := "none"
		maxRetries := 3

		if input != nil {
			if simulate, exists := input.Get("simulate_error"); exists {
				shouldSimulateError = simulate.(bool)
			}
			if errType, exists := input.Get("error_type"); exists {
				errorType = errType.(string)
			}
			if retries, exists := input.Get("max_retries"); exists {
				maxRetries = retries.(int)
			}
		}

		// Create result state
		result := domain.NewState()
		result.Set("agent_type", "error_recovery")
		result.Set("recovery_attempts", recoveryAttempts)

		if shouldSimulateError && recoveryAttempts < maxRetries {
			recoveryAttempts++
			result.Set("status", "error_occurred")
			result.Set("error_type", errorType)
			result.Set("recovery_strategy", "retry_with_backoff")
			result.Set("next_retry_delay", fmt.Sprintf("%ds", recoveryAttempts))

			// Return error based on type
			switch errorType {
			case "network":
				return result, fmt.Errorf("network error: connection timeout after %d attempts", recoveryAttempts)
			case "validation":
				return result, fmt.Errorf("validation error: invalid input format (attempt %d)", recoveryAttempts)
			case "resource":
				return result, fmt.Errorf("resource error: insufficient memory (attempt %d)", recoveryAttempts)
			default:
				return result, fmt.Errorf("unknown error occurred (attempt %d)", recoveryAttempts)
			}
		}

		// Successful execution (either no error or recovered)
		recoveryAttempts = 0
		result.Set("status", "completed")
		result.Set("recovery_successful", true)

		// Set recovery statistics
		recoveryStats := map[string]interface{}{
			"total_attempts":    recoveryAttempts + 1,
			"recovery_time":     fmt.Sprintf("0.%ds", recoveryAttempts*2),
			"error_resolved":    shouldSimulateError,
			"recovery_strategy": "retry_with_exponential_backoff",
		}
		result.Set("recovery_stats", recoveryStats)

		return result, nil
	}

	return agent
}

// Helper function to count unique tools across workflow phases
func getUniqueToolCount(phases []map[string]interface{}) int {
	toolSet := make(map[string]bool)
	for _, phase := range phases {
		if tools, ok := phase["tools_used"].([]string); ok {
			for _, tool := range tools {
				toolSet[tool] = true
			}
		}
	}
	return len(toolSet)
}
