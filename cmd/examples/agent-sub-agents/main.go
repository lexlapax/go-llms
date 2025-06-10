// ABOUTME: Demonstrates multi-agent systems with automatic tool registration and state sharing
// ABOUTME: Shows how sub-agents become tools and can be delegated to dynamically

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// CalculatorAgent performs mathematical operations
type CalculatorAgent struct {
	*core.BaseAgentImpl
}

func NewCalculatorAgent() *CalculatorAgent {
	return &CalculatorAgent{
		BaseAgentImpl: core.NewBaseAgent("calculator", "Performs mathematical calculations", domain.AgentTypeLLM),
	}
}

func (c *CalculatorAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Get the operation from state
	operation, _ := state.Get("operation")
	a, aOk := state.Get("a")
	b, bOk := state.Get("b")

	// Also check for simple input string
	if input, ok := state.Get("input"); ok && !aOk {
		// Parse simple expressions like "2 + 2"
		expr := fmt.Sprint(input)
		if strings.Contains(expr, "+") {
			fmt.Println("[Calculator] Parsing addition expression:", expr)
			// Simple parsing (production code would use proper expression parser)
			var x, y float64
			_, _ = fmt.Sscanf(expr, "%f + %f", &x, &y)
			result := x + y

			output := domain.NewState()
			output.Set("result", result)
			output.Set("output", fmt.Sprintf("%.2f", result))
			output.Set("explanation", fmt.Sprintf("%.2f + %.2f = %.2f", x, y, result))
			return output, nil
		}
	}

	// Handle structured operations
	if operation != nil && aOk && bOk {
		var result float64
		aFloat, _ := toFloat64(a)
		bFloat, _ := toFloat64(b)

		switch fmt.Sprint(operation) {
		case "add":
			result = aFloat + bFloat
		case "subtract":
			result = aFloat - bFloat
		case "multiply":
			result = aFloat * bFloat
		case "divide":
			if bFloat != 0 {
				result = aFloat / bFloat
			} else {
				return nil, fmt.Errorf("division by zero")
			}
		default:
			return nil, fmt.Errorf("unknown operation: %v", operation)
		}

		output := domain.NewState()
		output.Set("result", result)
		output.Set("output", fmt.Sprintf("%.2f", result))
		output.Set("explanation", fmt.Sprintf("%s %.2f and %.2f = %.2f", operation, aFloat, bFloat, result))

		// Demonstrate state inheritance - check if parent state has context
		if context, ok := state.Get("context"); ok {
			output.Set("context", context)
			fmt.Printf("[Calculator] Inherited context from parent: %v\n", context)
		}

		return output, nil
	}

	return nil, fmt.Errorf("invalid input: need operation and values or expression")
}

// ResearchAgent performs web research (simulated)
type ResearchAgent struct {
	*core.BaseAgentImpl
}

func NewResearchAgent() *ResearchAgent {
	return &ResearchAgent{
		BaseAgentImpl: core.NewBaseAgent("researcher", "Performs web research and fact checking", domain.AgentTypeLLM),
	}
}

func (r *ResearchAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	query, ok := state.Get("query")
	if !ok {
		if input, ok := state.Get("input"); ok {
			query = input
		} else {
			return nil, fmt.Errorf("no query provided")
		}
	}

	fmt.Printf("[Researcher] Searching for: %v\n", query)

	// Simulate research (in real implementation would use web tools)
	output := domain.NewState()
	output.Set("output", fmt.Sprintf("Research results for '%v':\n1. Wikipedia: General information about %v\n2. Recent news: Latest developments\n3. Academic papers: In-depth analysis", query, query))
	output.Set("sources", []string{"wikipedia.org", "news.google.com", "scholar.google.com"})

	// Demonstrate accessing shared state
	if sharedData, ok := state.Get("shared_context"); ok {
		fmt.Printf("[Researcher] Accessed shared context: %v\n", sharedData)
		output.Set("used_context", sharedData)
	}

	return output, nil
}

// SummarizerAgent summarizes text
type SummarizerAgent struct {
	*core.BaseAgentImpl
}

func NewSummarizerAgent() *SummarizerAgent {
	return &SummarizerAgent{
		BaseAgentImpl: core.NewBaseAgent("summarizer", "Summarizes long text into key points", domain.AgentTypeLLM),
	}
}

func (s *SummarizerAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	text, ok := state.Get("text")
	if !ok {
		if input, ok := state.Get("input"); ok {
			text = input
		} else {
			return nil, fmt.Errorf("no text provided")
		}
	}

	// Simple summary (real implementation would use LLM)
	textStr := fmt.Sprint(text)
	wordCount := len(strings.Fields(textStr))

	output := domain.NewState()
	output.Set("output", fmt.Sprintf("Summary: This text contains %d words. Key points: [simulated summary of the content]", wordCount))
	output.Set("word_count", wordCount)

	return output, nil
}

func main() {
	// Create sub-agents
	calculator := NewCalculatorAgent()
	researcher := NewResearchAgent()
	summarizer := NewSummarizerAgent()

	// Create main agent with sub-agents using the new simplified API
	mainAgent, err := core.NewLLMAgentWithSubAgentsFromString(
		"assistant",
		"mock", // In production, use real provider like "openai/gpt-4"
		calculator,
		researcher,
		summarizer,
	)
	if err != nil {
		log.Fatal("Failed to create agent:", err)
	}

	// Configure the main agent
	mainAgent.SetSystemPrompt(`You are a helpful assistant with access to specialized tools:
- calculator: For mathematical calculations
- researcher: For web research and fact checking  
- summarizer: For summarizing long text

You can delegate tasks to these sub-agents using the transfer_to_agent tool.`)

	// Register sub-agents for handoff to work
	// Note: Only register sub-agents to avoid circular registration
	_ = core.Register(calculator)
	_ = core.Register(researcher)
	_ = core.Register(summarizer)

	// Demonstrate automatic tool registration
	fmt.Println("=== Automatic Tool Registration ===")
	fmt.Println("Available tools:")
	for _, toolName := range mainAgent.ListTools() {
		tool, _ := mainAgent.GetTool(toolName)
		fmt.Printf("- %s: %s\n", toolName, tool.Description())
	}
	fmt.Println()

	// Example 1: Direct transfer using convenience method
	fmt.Println("=== Example 1: Direct Transfer ===")
	ctx := context.Background()

	result, err := mainAgent.TransferTo(ctx, "calculator", "Need to calculate", "5 + 3")
	if err != nil {
		log.Fatal("Transfer failed:", err)
	}
	if output, ok := result.Get("output"); ok {
		fmt.Printf("Calculator result: %v\n", output)
	}
	fmt.Println()

	// Example 2: Transfer with structured input
	fmt.Println("=== Example 2: Structured Transfer ===")
	calcInput := map[string]interface{}{
		"operation": "multiply",
		"a":         7,
		"b":         8,
		"context":   "calculating weekly hours",
	}

	result, err = mainAgent.TransferTo(ctx, "calculator", "Complex calculation", calcInput)
	if err != nil {
		log.Fatal("Transfer failed:", err)
	}
	if explanation, ok := result.Get("explanation"); ok {
		fmt.Printf("Calculation: %v\n", explanation)
	}
	fmt.Println()

	// Example 3: Research delegation
	fmt.Println("=== Example 3: Research Delegation ===")
	result, err = mainAgent.TransferTo(ctx, "researcher", "Research needed", "quantum computing applications")
	if err != nil {
		log.Fatal("Transfer failed:", err)
	}
	if output, ok := result.Get("output"); ok {
		fmt.Printf("Research results:\n%v\n", output)
	}
	fmt.Println()

	// Example 4: Demonstrate shared state context
	fmt.Println("=== Example 4: Shared State Context ===")

	// Create a state with shared context
	sharedState := domain.NewState()
	sharedState.Set("shared_context", "project: AI research paper")
	sharedState.Set("query", "latest advances in transformer models")

	// Enable shared state for sub-agents
	mainAgent.EnableSharedState(true)
	mainAgent.ConfigureStateInheritance(true, true, true)

	// Create shared state context
	sharedCtx := domain.NewSharedStateContext(sharedState)

	// Research will have access to shared context
	result = sharedCtx.LocalState()
	result.Set("input", "latest advances in transformer models")

	// Manual delegation showing state inheritance
	researchResult, err := researcher.Run(ctx, result)
	if err != nil {
		log.Fatal("Research failed:", err)
	}

	if output, ok := researchResult.Get("output"); ok {
		fmt.Printf("Research with context:\n%v\n", output)
	}
	if usedContext, ok := researchResult.Get("used_context"); ok {
		fmt.Printf("Used context: %v\n", usedContext)
	}
	fmt.Println()

	// Example 5: Chain multiple agents
	fmt.Println("=== Example 5: Agent Chaining ===")

	// First, research a topic
	researchResult, err = mainAgent.TransferTo(ctx, "researcher", "Initial research", "artificial intelligence history")
	if err != nil {
		log.Fatal("Research failed:", err)
	}

	// Then summarize the research
	if researchOutput, ok := researchResult.Get("output"); ok {
		summaryInput := map[string]interface{}{
			"text": researchOutput,
		}
		summaryResult, err := mainAgent.TransferTo(ctx, "summarizer", "Summarize research", summaryInput)
		if err != nil {
			log.Fatal("Summary failed:", err)
		}

		if summary, ok := summaryResult.Get("output"); ok {
			fmt.Printf("Summary of research:\n%v\n", summary)
		}
	}
	fmt.Println()

	// Demonstrate finding sub-agents
	fmt.Println("=== Sub-Agent Discovery ===")
	calcAgent := mainAgent.GetSubAgentByName("calculator")
	if calcAgent != nil {
		fmt.Printf("Found calculator agent: %s\n", calcAgent.Description())
	}

	// List all sub-agents
	fmt.Println("\nAll sub-agents:")
	for _, subAgent := range mainAgent.SubAgents() {
		fmt.Printf("- %s: %s\n", subAgent.Name(), subAgent.Description())
	}
}

// Helper function to convert interface{} to float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		return f, err
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}
