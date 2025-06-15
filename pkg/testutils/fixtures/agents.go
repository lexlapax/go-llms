// ABOUTME: Pre-configured mock agents for common testing scenarios
// ABOUTME: Provides ready-to-use agent fixtures with typical behavior patterns

package fixtures

import (
	"context"
	"fmt"
	"time"

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

// TrackingMockAgent creates a mock agent that tracks execution with delays
func TrackingMockAgent(name string, delay time.Duration) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)

	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		// Simulate delay
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
				// Delay completed
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			}
		}

		// Track execution
		newState := state.Clone()
		newState.Set(fmt.Sprintf("%s_result", name), fmt.Sprintf("data_from_%s", name))
		newState.Set(fmt.Sprintf("%s_executed", name), true)
		newState.Set(fmt.Sprintf("%s_timestamp", name), time.Now().Unix())

		// Pass through any existing data
		if data, exists := state.Get("accumulated_data"); exists {
			if accData, ok := data.([]string); ok {
				accData = append(accData, fmt.Sprintf("processed_by_%s", name))
				newState.Set("accumulated_data", accData)
			}
		} else {
			newState.Set("accumulated_data", []string{fmt.Sprintf("processed_by_%s", name)})
		}

		return newState, nil
	}

	return agent
}

// SpecialistMockAgent creates a mock agent with a specific specialty and processing time
func SpecialistMockAgent(name, specialty string, processTime time.Duration) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)
	agent.AgentDescription = fmt.Sprintf("Specialist in %s", specialty)

	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		// Simulate processing time
		if processTime > 0 {
			select {
			case <-time.After(processTime):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		task, _ := state.Get("task")

		output := domain.NewState()
		output.Set("output", fmt.Sprintf("%s specialist processed: %v", specialty, task))
		output.Set("specialty", specialty)
		output.Set("processed_at", time.Now().Unix())
		output.Set("processing_duration", processTime.String())

		// Add specialty-specific results
		switch specialty {
		case "data_analysis":
			output.Set("analysis", map[string]interface{}{
				"patterns_found":  3,
				"confidence":      0.85,
				"recommendations": []string{"optimize query", "add index", "cache results"},
			})
		case "research":
			output.Set("research_findings", []string{
				"Finding 1: Key insight about " + fmt.Sprintf("%v", task),
				"Finding 2: Related work in the field",
				"Finding 3: Future directions",
			})
		case "development":
			output.Set("code_artifacts", map[string]interface{}{
				"files_created": 2,
				"tests_written": 5,
				"coverage":      "85%",
			})
		}

		return output, nil
	}

	return agent
}

// ErrorSimulationMockAgent creates a mock agent that simulates various error scenarios
func ErrorSimulationMockAgent(name string, errorType string, errorAfterCalls int) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)
	callCount := 0

	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		callCount++

		// Return error after specified number of calls
		if errorAfterCalls > 0 && callCount >= errorAfterCalls {
			switch errorType {
			case "timeout":
				return nil, context.DeadlineExceeded
			case "canceled":
				return nil, context.Canceled
			case "network":
				return nil, fmt.Errorf("network error: connection timeout")
			case "rate_limit":
				return nil, fmt.Errorf("rate limit exceeded: retry after 60s")
			case "validation":
				return nil, fmt.Errorf("validation error: invalid input format")
			case "panic":
				panic(fmt.Sprintf("simulated panic from %s", name))
			default:
				return nil, fmt.Errorf("simulated error: %s", errorType)
			}
		}

		// Normal execution before error
		result := domain.NewState()
		result.Set("output", fmt.Sprintf("Success from %s (call %d)", name, callCount))
		result.Set("call_count", callCount)

		return result, nil
	}

	return agent
}

// StateBuilderMockAgent creates a mock agent that builds complex state data
func StateBuilderMockAgent(name string, modifications map[string]interface{}) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)

	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		newState := state.Clone()

		// Apply all modifications
		for key, value := range modifications {
			newState.Set(key, value)
		}

		// Track which agent modified the state
		if history, exists := state.Get("modification_history"); exists {
			if histList, ok := history.([]string); ok {
				histList = append(histList, name)
				newState.Set("modification_history", histList)
			}
		} else {
			newState.Set("modification_history", []string{name})
		}

		// Add metadata about modifications
		newState.Set(fmt.Sprintf("%s_modifications", name), len(modifications))
		newState.Set(fmt.Sprintf("%s_timestamp", name), time.Now().Unix())

		return newState, nil
	}

	return agent
}

// CoordinatorMockAgent creates a mock agent that simulates coordination behavior
func CoordinatorMockAgent(name string) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)
	agent.AgentDescription = "Coordinator agent for multi-agent workflows"
	delegationCount := 0

	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		delegationCount++

		// Analyze task and decide delegation
		task, _ := state.Get("task")

		output := domain.NewState()
		output.Set("output", fmt.Sprintf("Coordinator analyzed task '%v' and delegated to %d sub-agents", task, delegationCount))
		output.Set("delegated", true)
		output.Set("delegation_count", delegationCount)
		output.Set("coordinator", name)

		// Create delegation plan
		delegationPlan := []map[string]interface{}{
			{"agent": "specialist1", "task": "analyze", "priority": "high"},
			{"agent": "specialist2", "task": "process", "priority": "medium"},
			{"agent": "specialist3", "task": "finalize", "priority": "low"},
		}
		output.Set("delegation_plan", delegationPlan)

		// Pass through any sub-agent results
		if subResults, exists := state.Get("sub_results"); exists {
			output.Set("aggregated_results", subResults)
		}

		return output, nil
	}

	return agent
}

// QualityRefinementMockAgent creates a mock agent that simulates iterative quality improvement
func QualityRefinementMockAgent(name string, initialQuality float64, improvementRate float64) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)
	iterationCount := 0
	currentQuality := initialQuality

	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		iterationCount++

		// Improve quality
		currentQuality = currentQuality + (1-currentQuality)*improvementRate
		if currentQuality > 1.0 {
			currentQuality = 1.0
		}

		output := domain.NewState()
		output.Set("iteration", iterationCount)
		output.Set("quality", currentQuality)
		output.Set("output", fmt.Sprintf("Iteration %d: Improved quality to %.2f", iterationCount, currentQuality))

		// Add refinement details
		refinementDetails := map[string]interface{}{
			"improvements_made": []string{
				"Enhanced clarity",
				"Fixed issues",
				"Added details",
			},
			"quality_delta":    improvementRate,
			"time_spent":       fmt.Sprintf("%dms", iterationCount*100),
			"confidence_level": currentQuality * 0.9,
		}
		output.Set("refinement_details", refinementDetails)

		// Pass through content being refined
		if content, exists := state.Get("content"); exists {
			output.Set("content", fmt.Sprintf("%v [refined x%d]", content, iterationCount))
		}

		return output, nil
	}

	return agent
}

// TimeoutMockAgent creates a mock agent that simulates timeout scenarios
func TimeoutMockAgent(name string, timeout time.Duration) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)

	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		// Create a context with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Simulate work that takes longer than timeout
		workDuration := timeout + 100*time.Millisecond

		select {
		case <-time.After(workDuration):
			// This should not happen due to timeout
			result := domain.NewState()
			result.Set("output", "Completed successfully")
			return result, nil
		case <-timeoutCtx.Done():
			// Timeout occurred
			return nil, fmt.Errorf("operation timed out after %v", timeout)
		}
	}

	return agent
}

// SharedDataBuilderMockAgent creates a mock agent that accumulates data in shared_data
func SharedDataBuilderMockAgent(name string, key string, value interface{}) *mocks.MockAgent {
	agent := mocks.NewMockAgent(name)

	agent.OnRun = func(ctx context.Context, state *domain.State) (*domain.State, error) {
		// Get existing shared data
		existingData := make(map[string]interface{})
		if data, exists := state.Get("shared_data"); exists {
			if m, ok := data.(map[string]interface{}); ok {
				// Make a copy to avoid modifying original
				for k, v := range m {
					existingData[k] = v
				}
			}
		}

		// Add our data
		existingData[key] = value

		// Create new state with updated shared data
		newState := state.Clone()
		newState.Set("shared_data", existingData)
		newState.Set("output", fmt.Sprintf("Added %s: %v", key, value))

		return newState, nil
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
