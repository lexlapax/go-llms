package main

// ABOUTME: Example demonstrating multi-provider strategies using workflow agents
// ABOUTME: Shows how to use parallel workflows to implement fastest and consensus patterns

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
)

func main() {
	ctx := context.Background()

	// Check for API keys
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	geminiKey := os.Getenv("GEMINI_API_KEY")

	if openaiKey == "" && anthropicKey == "" && geminiKey == "" {
		log.Println("No API keys found. Using mock agents for demonstration.")
		runWithMockAgents(ctx)
		return
	}

	fmt.Println("=== Multi-Provider Workflow Examples ===")

	// Example 1: Fastest response pattern
	fastestExample(ctx, openaiKey, anthropicKey, geminiKey)

	// Example 2: Consensus pattern
	consensusExample(ctx, openaiKey, anthropicKey, geminiKey)

	// Example 3: Primary with fallback pattern
	primaryFallbackExample(ctx, openaiKey, anthropicKey, geminiKey)
}

func fastestExample(ctx context.Context, openaiKey, anthropicKey, geminiKey string) {
	fmt.Println("=== Fastest Response Pattern ===")
	fmt.Println("Running multiple providers in parallel and returning the first response...")

	// Create agents for available providers
	var agents []domain.BaseAgent

	if openaiKey != "" {
		agent, err := core.NewAgentFromString("openai-agent", "openai/gpt-4o")
		if err == nil {
			agents = append(agents, agent)
			fmt.Println("- Added OpenAI agent")
		}
	}

	if anthropicKey != "" {
		agent, err := core.NewAgentFromString("anthropic-agent", "anthropic/claude-3-5-sonnet-latest")
		if err == nil {
			agents = append(agents, agent)
			fmt.Println("- Added Anthropic agent")
		}
	}

	if geminiKey != "" {
		agent, err := core.NewAgentFromString("gemini-agent", "gemini/gemini-1.5-flash")
		if err == nil {
			agents = append(agents, agent)
			fmt.Println("- Added Gemini agent")
		}
	}

	if len(agents) < 2 {
		fmt.Println("Need at least 2 providers for multi-provider patterns. Skipping.")
		return
	}

	// Create parallel workflow with MergeFirst strategy (returns first response)
	parallelWorkflow := workflow.NewParallelAgent("fastest-response").
		WithMergeStrategy(workflow.MergeFirst).
		WithMaxConcurrency(len(agents))

	// Add all agents to the workflow
	for _, agent := range agents {
		parallelWorkflow.AddAgent(agent)
	}

	// Create initial state
	state := domain.NewState()
	state.Set("user_input", "Tell me a very short joke (one-liner only).")

	// Run the workflow
	fmt.Println("\nSending request to all providers...")
	start := time.Now()

	result, err := parallelWorkflow.Run(ctx, state)
	if err != nil {
		log.Printf("Workflow error: %v", err)
		return
	}

	duration := time.Since(start)

	// Display result
	if output, exists := result.Get("output"); exists {
		fmt.Printf("\nFirst response received in %v:\n%v\n", duration, output)
	}

	// Check which agent responded first
	if metadata := parallelWorkflow.Status(); metadata != nil {
		for stepName, stepStatus := range metadata.Steps {
			if stepStatus.State == workflow.StepStateCompleted {
				fmt.Printf("Response from: %s\n", stepName)
				break
			}
		}
	}

	fmt.Println()
}

func consensusExample(ctx context.Context, openaiKey, anthropicKey, geminiKey string) {
	fmt.Println("=== Consensus Pattern ===")
	fmt.Println("Running multiple providers and comparing their responses...")

	// Create agents
	var agents []domain.BaseAgent

	if openaiKey != "" {
		agent, err := core.NewAgentFromString("openai-agent", "openai/gpt-4o-mini")
		if err == nil {
			agents = append(agents, agent)
		}
	}

	if anthropicKey != "" {
		agent, err := core.NewAgentFromString("anthropic-agent", "anthropic/claude-3-5-haiku-latest")
		if err == nil {
			agents = append(agents, agent)
		}
	}

	if geminiKey != "" {
		agent, err := core.NewAgentFromString("gemini-agent", "gemini/gemini-1.5-flash")
		if err == nil {
			agents = append(agents, agent)
		}
	}

	if len(agents) < 2 {
		fmt.Println("Need at least 2 providers for consensus. Skipping.")
		return
	}

	// Create parallel workflow with custom merge function for consensus
	consensusWorkflow := workflow.NewParallelAgent("consensus").
		WithMergeFunc(func(results map[string]*domain.State) *domain.State {
			// Custom consensus logic
			responses := make(map[string]int)
			allResponses := []string{}

			// Collect all responses
			for _, state := range results {
				if output, exists := state.Get("output"); exists {
					response := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", output)))
					responses[response]++
					allResponses = append(allResponses, fmt.Sprintf("%v", output))
				}
			}

			// Create result state
			resultState := domain.NewState()

			// Find consensus
			var consensusResponse string
			maxCount := 0
			for response, count := range responses {
				if count > maxCount {
					maxCount = count
					consensusResponse = response
				}
			}

			// Determine consensus level
			consensusLevel := float64(maxCount) / float64(len(results))

			if consensusLevel > 0.5 {
				resultState.Set("consensus", true)
				resultState.Set("consensus_level", fmt.Sprintf("%.0f%%", consensusLevel*100))
				resultState.Set("output", fmt.Sprintf("Consensus reached (%.0f%% agreement): %s",
					consensusLevel*100, consensusResponse))
			} else {
				resultState.Set("consensus", false)
				resultState.Set("output", fmt.Sprintf("No consensus. Responses varied:\n%s",
					strings.Join(allResponses, "\n")))
			}

			resultState.Set("all_responses", allResponses)
			return resultState
		}).
		WithMaxConcurrency(len(agents))

	// Add agents
	for _, agent := range agents {
		consensusWorkflow.AddAgent(agent)
	}

	// Test with a factual question
	state := domain.NewState()
	state.Set("user_input", "What is 2+2? Answer with just the number.")

	fmt.Printf("Asking %d providers: What is 2+2?\n", len(agents))

	result, err := consensusWorkflow.Run(ctx, state)
	if err != nil {
		log.Printf("Workflow error: %v", err)
		return
	}

	// Display results
	if output, exists := result.Get("output"); exists {
		fmt.Printf("\nResult: %v\n", output)
	}

	if allResponses, exists := result.Get("all_responses"); exists {
		fmt.Println("\nIndividual responses:")
		for i, response := range allResponses.([]string) {
			fmt.Printf("- Provider %d: %s\n", i+1, response)
		}
	}

	fmt.Println()
}

func primaryFallbackExample(ctx context.Context, openaiKey, anthropicKey, geminiKey string) {
	fmt.Println("=== Primary with Fallback Pattern ===")
	fmt.Println("Using sequential workflow to try providers in order...")

	// Create a sequential workflow that tries providers in order
	sequentialWorkflow := workflow.NewSequentialAgent("primary-fallback").
		WithStopOnError(false) // Continue on error to try fallbacks

	// Add providers in priority order
	added := 0

	if openaiKey != "" {
		agent, err := core.NewAgentFromString("primary-openai", "openai/gpt-4o")
		if err == nil {
			sequentialWorkflow.AddAgent(agent)
			fmt.Println("- Primary: OpenAI")
			added++
		}
	}

	if anthropicKey != "" && added < 3 {
		agent, err := core.NewAgentFromString("fallback-anthropic", "anthropic/claude-3-5-sonnet-latest")
		if err == nil {
			sequentialWorkflow.AddAgent(agent)
			fmt.Println("- Fallback 1: Anthropic")
			added++
		}
	}

	if geminiKey != "" && added < 3 {
		agent, err := core.NewAgentFromString("fallback-gemini", "gemini/gemini-1.5-pro")
		if err == nil {
			sequentialWorkflow.AddAgent(agent)
			fmt.Println("- Fallback 2: Gemini")
			added++
		}
	}

	if added == 0 {
		fmt.Println("No providers available. Skipping.")
		return
	}

	// Create state with a request
	state := domain.NewState()
	state.Set("user_input", "What's the capital of France? Answer in one word.")

	fmt.Println("\nSending request with fallback chain...")

	result, err := sequentialWorkflow.Run(ctx, state)
	if err != nil {
		log.Printf("All providers failed: %v", err)
		return
	}

	// Display result
	if output, exists := result.Get("output"); exists {
		fmt.Printf("\nResponse: %v\n", output)
	}

	// Show which provider responded
	status := sequentialWorkflow.Status()
	for step, stepStatus := range status.Steps {
		if stepStatus.State == workflow.StepStateCompleted {
			fmt.Printf("Responded by: %s\n", step)
			break
		}
	}

	fmt.Println()
}

// Mock implementation for demonstration
func runWithMockAgents(ctx context.Context) {
	fmt.Println("=== Running with Mock Agents ===")

	// Create mock agents with different delays
	mockAgents := []domain.BaseAgent{
		createMockAgent("mock-fast", "Fast response!", 100*time.Millisecond),
		createMockAgent("mock-medium", "Medium response!", 300*time.Millisecond),
		createMockAgent("mock-slow", "Slow response!", 500*time.Millisecond),
	}

	// Demonstrate fastest response
	fmt.Println("1. Fastest Response Pattern:")
	fastestWorkflow := workflow.NewParallelAgent("fastest-mock").
		WithMergeStrategy(workflow.MergeFirst)

	for _, agent := range mockAgents {
		fastestWorkflow.AddAgent(agent)
	}

	state := domain.NewState()
	state.Set("prompt", "test")

	start := time.Now()
	result, _ := fastestWorkflow.Run(ctx, state)
	duration := time.Since(start)

	if response, exists := result.Get("response"); exists {
		fmt.Printf("   Got: %v (in %v)\n\n", response, duration)
	}

	// Demonstrate consensus
	fmt.Println("2. Consensus Pattern:")
	consensusAgents := []domain.BaseAgent{
		createMockAgent("mock-1", "Answer: 42", 100*time.Millisecond),
		createMockAgent("mock-2", "Answer: 42", 150*time.Millisecond),
		createMockAgent("mock-3", "Answer: 24", 200*time.Millisecond),
	}

	consensusWorkflow := workflow.NewParallelAgent("consensus-mock").
		WithMergeStrategy(workflow.MergeAll)

	for _, agent := range consensusAgents {
		consensusWorkflow.AddAgent(agent)
	}

	result, _ = consensusWorkflow.Run(ctx, state)

	// Count responses
	responses := make(map[string]int)
	for i := 0; i < len(consensusAgents); i++ {
		key := fmt.Sprintf("response_%d", i)
		if resp, exists := result.Get(key); exists {
			responses[fmt.Sprintf("%v", resp)]++
		}
	}

	fmt.Println("   Responses:")
	for resp, count := range responses {
		fmt.Printf("   - %s: %d votes\n", resp, count)
	}
}

// Helper to create mock agents
func createMockAgent(name, response string, delay time.Duration) domain.BaseAgent {
	return &mockAgent{
		BaseAgent: core.NewBaseAgent(name, "Mock agent", domain.AgentTypeCustom),
		response:  response,
		delay:     delay,
	}
}

type mockAgent struct {
	domain.BaseAgent
	response string
	delay    time.Duration
}

func (m *mockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Simulate processing delay
	select {
	case <-time.After(m.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	newState := state.Clone()
	newState.Set("response", m.response)
	newState.Set("output", m.response) // Also set output for compatibility

	return newState, nil
}
