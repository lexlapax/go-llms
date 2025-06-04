// ABOUTME: Example demonstrating parallel workflow execution with multiple LLM agents
// ABOUTME: Shows concurrent execution, merge strategies, and concurrency control

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
	// Example 1: Parallel analysis with merge all strategy
	parallelAnalysisExample()

	// Example 2: Race to find the best answer (merge first)
	racingAgentsExample()

	// Example 3: Custom merge strategy
	customMergeExample()
}

func parallelAnalysisExample() {
	fmt.Println("=== Parallel Analysis Example ===")
	fmt.Println("Analyzing a topic from multiple perspectives simultaneously...")
	fmt.Println()

	ctx := context.Background()

	// Create agents for different types of analysis
	var technicalAnalyst domain.BaseAgent
	if ta, err := core.NewAgentFromString("technical-analyst", "claude"); err != nil {
		log.Printf("Using mock technical analyst: %v", err)
		technicalAnalyst = createMockAnalyst("technical", "Technical analysis")
	} else {
		ta.SetSystemPrompt("You are a technical analyst. Analyze the technical aspects and implementation details.")
		technicalAnalyst = ta
	}

	var businessAnalyst domain.BaseAgent
	if ba, err := core.NewAgentFromString("business-analyst", "gpt-4"); err != nil {
		log.Printf("Using mock business analyst: %v", err)
		businessAnalyst = createMockAnalyst("business", "Business analysis")
	} else {
		ba.SetSystemPrompt("You are a business analyst. Analyze the business impact and market implications.")
		businessAnalyst = ba
	}

	var ethicalAnalyst domain.BaseAgent
	if ea, err := core.NewAgentFromString("ethical-analyst", "claude"); err != nil {
		log.Printf("Using mock ethical analyst: %v", err)
		ethicalAnalyst = createMockAnalyst("ethical", "Ethical analysis")
	} else {
		ea.SetSystemPrompt("You are an ethics expert. Analyze the ethical implications and societal impact.")
		ethicalAnalyst = ea
	}

	// Create parallel workflow
	analysisWorkflow := workflow.NewParallelAgent("multi-perspective-analysis").
		WithMaxConcurrency(3). // Run all three concurrently
		WithMergeStrategy(workflow.MergeAll).
		AddAgent(technicalAnalyst).
		AddAgent(businessAnalyst).
		AddAgent(ethicalAnalyst)

	// Create initial state with topic
	initialState := domain.NewState()
	initialState.Set("prompt", "Analyze the implications of widespread AI adoption in healthcare")

	// Run workflow
	fmt.Println("Starting parallel analysis...")
	start := time.Now()

	result, err := analysisWorkflow.Run(ctx, initialState)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("\nAnalysis completed in %v\n", duration)

	// Display results
	if parallelResults, exists := result.Get("parallel_results"); exists {
		results := parallelResults.(map[string]interface{})

		for analyst, data := range results {
			fmt.Printf("\n--- %s Analysis ---\n", analyst)
			if resultMap, ok := data.(map[string]interface{}); ok {
				if response, exists := resultMap["response"]; exists {
					fmt.Printf("%v\n", response)
				}
			}
		}
	}
}

func racingAgentsExample() {
	fmt.Println("\n=== Racing Agents Example ===")
	fmt.Println("Multiple agents race to provide the fastest response...")
	fmt.Println()

	ctx := context.Background()

	// Create multiple agents with different models/speeds
	fastAgent := createMockAgentWithDelay("fast-agent", 100*time.Millisecond, "Quick but potentially less detailed response")
	mediumAgent := createMockAgentWithDelay("medium-agent", 300*time.Millisecond, "Balanced response with good detail")
	slowAgent := createMockAgentWithDelay("slow-agent", 500*time.Millisecond, "Comprehensive and detailed response")

	// Create parallel workflow with MergeFirst strategy
	racingWorkflow := workflow.NewParallelAgent("racing-agents").
		WithMergeStrategy(workflow.MergeFirst). // Use first completed result
		WithTimeout(400 * time.Millisecond).    // Timeout before slowest agent
		AddAgent(fastAgent).
		AddAgent(mediumAgent).
		AddAgent(slowAgent)

	// Run workflow
	initialState := domain.NewState()
	initialState.Set("prompt", "What is the capital of France?")

	fmt.Println("Starting race...")
	result, err := racingWorkflow.Run(ctx, initialState)
	if err != nil {
		log.Printf("Workflow error: %v", err)
	}

	// Display winner
	if response, exists := result.Get("response"); exists {
		fmt.Printf("\nFirst response received: %v\n", response)
	}
}

func customMergeExample() {
	fmt.Println("\n=== Custom Merge Example ===")
	fmt.Println("Using custom merge function to combine results...")
	fmt.Println()

	ctx := context.Background()

	// Create agents that return scores
	agent1 := createScoringAgent("scorer1", 85)
	agent2 := createScoringAgent("scorer2", 92)
	agent3 := createScoringAgent("scorer3", 78)

	// Custom merge function that calculates average score
	averageMerge := func(results map[string]*domain.State) *domain.State {
		merged := domain.NewState()
		totalScore := 0
		count := 0

		for agentName, state := range results {
			if score, exists := state.Get("score"); exists {
				if s, ok := score.(int); ok {
					totalScore += s
					count++
				}
			}
			merged.Set(fmt.Sprintf("%s_completed", agentName), true)
		}

		if count > 0 {
			merged.Set("average_score", totalScore/count)
			merged.Set("total_score", totalScore)
			merged.Set("agent_count", count)
		}

		return merged
	}

	// Create workflow with custom merge
	scoringWorkflow := workflow.NewParallelAgent("scoring-workflow").
		WithMergeFunc(averageMerge).
		AddAgent(agent1).
		AddAgent(agent2).
		AddAgent(agent3)

	// Run workflow
	initialState := domain.NewState()
	result, err := scoringWorkflow.Run(ctx, initialState)
	if err != nil {
		log.Fatalf("Workflow failed: %v", err)
	}

	// Display results
	if avg, exists := result.Get("average_score"); exists {
		fmt.Printf("Average score: %v\n", avg)
	}
	if total, exists := result.Get("total_score"); exists {
		fmt.Printf("Total score: %v\n", total)
	}
}

// Helper functions to create mock agents
func createMockAnalyst(name, analysisType string) domain.BaseAgent {
	agent := &mockAgent{
		BaseAgent: core.NewBaseAgent(name, analysisType, domain.AgentTypeCustom),
		response:  fmt.Sprintf("%s: This would contain the actual %s analysis results.", name, analysisType),
	}
	return agent
}

func createMockAgentWithDelay(name string, delay time.Duration, response string) domain.BaseAgent {
	agent := &mockAgent{
		BaseAgent: core.NewBaseAgent(name, "Mock agent with delay", domain.AgentTypeCustom),
		delay:     delay,
		response:  response,
	}
	return agent
}

func createScoringAgent(name string, score int) domain.BaseAgent {
	agent := &mockAgent{
		BaseAgent: core.NewBaseAgent(name, "Scoring agent", domain.AgentTypeCustom),
		score:     score,
	}
	return agent
}

type mockAgent struct {
	domain.BaseAgent
	delay    time.Duration
	response string
	score    int
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

	if m.response != "" {
		newState.Set("response", m.response)
	}

	if m.score > 0 {
		newState.Set("score", m.score)
	}

	return newState, nil
}
