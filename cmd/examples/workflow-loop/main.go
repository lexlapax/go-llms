// ABOUTME: Example demonstrating loop workflow execution with iterative processing
// ABOUTME: Shows while loops, count loops, and advanced loop control features

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
)

func main() {
	// Example 1: Count loop with batch processing
	countLoopExample()

	// Example 2: While loop with convergence condition
	whileLoopExample()

	// Example 3: Data processing pipeline with retry logic
	retryLoopExample()

	// Example 4: Advanced loop features demonstration
	advancedLoopExample()
}

func countLoopExample() {
	fmt.Println("=== Count Loop Example ===")
	fmt.Println("Processing a batch of items with a fixed number of iterations...")
	fmt.Println()

	ctx := context.Background()

	// Create a data processor agent
	var dataProcessor domain.BaseAgent
	if dp, err := core.NewAgentFromString("batch-processor", "claude"); err != nil {
		log.Printf("Using mock batch processor: %v", err)
		dataProcessor = createMockAgent("batch-processor", "Batch item processed", 100*time.Millisecond)
	} else {
		dp.SetSystemPrompt("You are a data processor. Process each batch item and provide analysis.")
		dataProcessor = dp
	}

	// Create count loop that processes 5 batches
	batchLoop := workflow.CountLoop("batch-processing", 5, createAgentStep("process", dataProcessor))

	// Create initial state with batch data
	initialState := domain.NewState()
	initialState.Set("batch_size", 100)
	initialState.Set("total_processed", 0)
	initialState.Set("prompt", "Process the next batch of data items and provide a summary of processing results.")

	// Run workflow
	fmt.Println("Starting batch processing...")
	start := time.Now()

	result, err := batchLoop.Run(ctx, initialState)
	if err != nil {
		log.Fatalf("Batch processing failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("Batch processing completed in %v\n", duration)

	// Display results
	if response, exists := result.Get("response"); exists {
		fmt.Printf("Final batch result: %v\n", response)
	}

	fmt.Printf("Total iterations: %d\n", batchLoop.GetCurrentIteration())
	fmt.Printf("Total duration: %v\n", batchLoop.GetTotalDuration())
	fmt.Println()
}

func whileLoopExample() {
	fmt.Println("=== While Loop Example ===")
	fmt.Println("Iterative optimization until convergence...")
	fmt.Println()

	ctx := context.Background()

	// Create an optimization agent
	var optimizer domain.BaseAgent
	if opt, err := core.NewAgentFromString("optimizer", "gpt-4"); err != nil {
		log.Printf("Using mock optimizer: %v", err)
		optimizer = createOptimizerAgent("optimizer")
	} else {
		opt.SetSystemPrompt("You are an optimization specialist. Analyze the current parameters and suggest improvements.")
		optimizer = opt
	}

	// Create while loop that continues until convergence
	optimizationLoop := workflow.WhileLoop("optimization", func(state *domain.State, iteration int) bool {
		// Continue while error is above threshold and we haven't hit iteration limit
		if errorRate, exists := state.Get("error_rate"); exists {
			return errorRate.(float64) > 0.01 && iteration < 10 // Stop if error < 1% or max iterations
		}
		return iteration < 10 // Fallback limit
	}, createAgentStep("optimize", optimizer))

	// Create initial state with optimization parameters
	initialState := domain.NewState()
	initialState.Set("error_rate", 0.15) // Start with 15% error
	initialState.Set("learning_rate", 0.1)
	initialState.Set("prompt", "Analyze the current optimization parameters and suggest improvements to reduce error rate.")

	// Run workflow
	fmt.Println("Starting optimization...")
	start := time.Now()

	result, err := optimizationLoop.Run(ctx, initialState)
	if err != nil {
		log.Fatalf("Optimization failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("Optimization completed in %v\n", duration)

	// Display results
	if errorRate, exists := result.Get("error_rate"); exists {
		fmt.Printf("Final error rate: %.4f\n", errorRate)
	}
	if response, exists := result.Get("response"); exists {
		fmt.Printf("Final optimization result: %v\n", response)
	}

	fmt.Printf("Total iterations: %d\n", optimizationLoop.GetCurrentIteration())
	fmt.Printf("Convergence achieved: %v\n", optimizationLoop.GetCurrentIteration() < 10)
	fmt.Println()
}

func retryLoopExample() {
	fmt.Println("=== Retry Loop Example ===")
	fmt.Println("API call with retry logic and exponential backoff...")
	fmt.Println()

	ctx := context.Background()

	// Create an API client agent
	var apiClient domain.BaseAgent
	if api, err := core.NewAgentFromString("api-client", "claude"); err != nil {
		log.Printf("Using mock API client: %v", err)
		apiClient = createRetryAgent("api-client")
	} else {
		api.SetSystemPrompt("You are an API client. Attempt to make an API call and return success or failure status.")
		apiClient = api
	}

	// Create retry loop with exponential backoff
	retryLoop := workflow.NewLoopAgent("api-retry").
		SetLoopBody(createAgentStep("api-call", apiClient)).
		WithUntilCondition(func(state *domain.State, iteration int) bool {
			// Break when API call succeeds
			if success, exists := state.Get("api_success"); exists {
				return success.(bool)
			}
			return false
		}).
		WithMaxIterations(5).
		WithBreakOnError(false).                   // Continue on error for retry logic
		WithIterationDelay(100 * time.Millisecond) // Base delay, would normally implement exponential backoff

	// Create initial state
	initialState := domain.NewState()
	initialState.Set("api_success", false)
	initialState.Set("attempt_count", 0)
	initialState.Set("prompt", "Attempt to make an API call. Return success status and any error information.")

	// Run workflow
	fmt.Println("Starting API calls with retry logic...")
	start := time.Now()

	result, err := retryLoop.Run(ctx, initialState)
	if err != nil {
		log.Printf("API retry workflow failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("API retry completed in %v\n", duration)

	// Display results
	if success, exists := result.Get("api_success"); exists {
		fmt.Printf("API call successful: %v\n", success)
	}
	if response, exists := result.Get("response"); exists {
		fmt.Printf("Final API response: %v\n", response)
	}

	fmt.Printf("Total attempts: %d\n", retryLoop.GetCurrentIteration())
	fmt.Printf("Max attempts reached: %v\n", retryLoop.GetCurrentIteration() >= 5)
	fmt.Println()
}

func advancedLoopExample() {
	fmt.Println("=== Advanced Loop Features Example ===")
	fmt.Println("Demonstrating result collection, state management, and loop control...")
	fmt.Println()

	ctx := context.Background()

	// Create a survey processor agent
	var surveyProcessor domain.BaseAgent
	if sp, err := core.NewAgentFromString("survey-processor", "gpt-4"); err != nil {
		log.Printf("Using mock survey processor: %v", err)
		surveyProcessor = createSurveyAgent("survey-processor")
	} else {
		sp.SetSystemPrompt("You are a survey processor. Analyze survey responses and extract insights.")
		surveyProcessor = sp
	}

	// Create advanced loop with multiple features
	surveyLoop := workflow.NewLoopAgent("survey-analysis").
		SetLoopBody(createAgentStep("analyze", surveyProcessor)).
		WithMaxIterations(3).
		WithCollectResults(true).
		WithPassStateThrough(true).
		WithIterationDelay(50 * time.Millisecond)

	// Create initial state with survey data
	initialState := domain.NewState()
	initialState.Set("survey_responses", []string{
		"Very satisfied with the product",
		"Could use better customer support",
		"Excellent value for money",
	})
	initialState.Set("analyzed_count", 0)
	initialState.Set("insights", []string{})
	initialState.Set("prompt", "Analyze the next survey response and extract key insights and sentiment.")

	// Run workflow
	fmt.Println("Starting survey analysis...")
	start := time.Now()

	result, err := surveyLoop.Run(ctx, initialState)
	if err != nil {
		log.Fatalf("Survey analysis failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("Survey analysis completed in %v\n", duration)

	// Display results
	if analyzedCount, exists := result.Get("analyzed_count"); exists {
		fmt.Printf("Surveys analyzed: %v\n", analyzedCount)
	}
	if insights, exists := result.Get("insights"); exists {
		fmt.Printf("Generated insights: %v\n", insights)
	}

	// Display iteration results
	iterationResults := surveyLoop.GetIterationResults()
	fmt.Printf("Collected %d iteration results:\n", len(iterationResults))
	for i, result := range iterationResults {
		resultMap := result.(map[string]interface{})
		fmt.Printf("  Iteration %d: Duration=%v, Success=%v\n",
			i, resultMap["duration"], resultMap["error"] == nil)
	}

	fmt.Printf("Total iterations: %d\n", surveyLoop.GetCurrentIteration())
	fmt.Printf("Total duration: %v\n", surveyLoop.GetTotalDuration())
	fmt.Println()
}

// Helper functions

func createAgentStep(name string, agent domain.BaseAgent) workflow.WorkflowStep {
	return workflow.NewAgentStep(name, agent)
}

func createMockAgent(name, response string, delay time.Duration) domain.BaseAgent {
	return &mockAgent{
		BaseAgent: core.NewBaseAgent(name, "Mock agent", domain.AgentTypeCustom),
		response:  response,
		delay:     delay,
	}
}

func createOptimizerAgent(name string) domain.BaseAgent {
	return &optimizerAgent{
		BaseAgent: core.NewBaseAgent(name, "Optimizer agent", domain.AgentTypeCustom),
	}
}

func createRetryAgent(name string) domain.BaseAgent {
	return &retryAgent{
		BaseAgent: core.NewBaseAgent(name, "Retry agent", domain.AgentTypeCustom),
		attempts:  0,
	}
}

func createSurveyAgent(name string) domain.BaseAgent {
	return &surveyAgent{
		BaseAgent: core.NewBaseAgent(name, "Survey agent", domain.AgentTypeCustom),
	}
}

// Mock agents for different scenarios

type mockAgent struct {
	domain.BaseAgent
	response string
	delay    time.Duration
}

func (m *mockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	newState := state.Clone()
	newState.Set("response", m.response)
	return newState, nil
}

type optimizerAgent struct {
	domain.BaseAgent
}

func (o *optimizerAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	time.Sleep(80 * time.Millisecond) // Simulate optimization computation

	newState := state.Clone()

	// Simulate optimization progress - reduce error rate
	if errorRate, exists := state.Get("error_rate"); exists {
		currentError := errorRate.(float64)
		newError := currentError * 0.7 // Reduce by 30% each iteration
		if newError < 0.01 {
			newError = 0.005 // Converged
		}
		newState.Set("error_rate", newError)
	}

	newState.Set("response", "Optimization step completed")
	return newState, nil
}

type retryAgent struct {
	domain.BaseAgent
	attempts int
}

func (r *retryAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	time.Sleep(50 * time.Millisecond) // Simulate API call

	r.attempts++
	newState := state.Clone()
	newState.Set("attempt_count", r.attempts)

	// Simulate success after 3 attempts
	if r.attempts >= 3 {
		newState.Set("api_success", true)
		newState.Set("response", "API call successful")
	} else {
		newState.Set("api_success", false)
		newState.Set("response", fmt.Sprintf("API call failed (attempt %d)", r.attempts))
	}

	return newState, nil
}

type surveyAgent struct {
	domain.BaseAgent
}

func (s *surveyAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	time.Sleep(60 * time.Millisecond) // Simulate analysis

	newState := state.Clone()

	// Update analyzed count
	analyzedCount := 0
	if count, exists := state.Get("analyzed_count"); exists {
		analyzedCount = count.(int)
	}
	analyzedCount++
	newState.Set("analyzed_count", analyzedCount)

	// Add insights
	insights := []string{}
	if existingInsights, exists := state.Get("insights"); exists {
		insights = existingInsights.([]string)
	}

	newInsight := fmt.Sprintf("Insight %d: Analysis completed for survey response", analyzedCount)
	insights = append(insights, newInsight)
	newState.Set("insights", insights)

	newState.Set("response", fmt.Sprintf("Survey response %d analyzed", analyzedCount))
	return newState, nil
}
