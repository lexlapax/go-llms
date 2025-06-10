// ABOUTME: Demonstrates hierarchical multi-agent coordination with teams of specialized agents
// ABOUTME: Shows how to build complex agent systems with automatic delegation and result aggregation

package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/agent/workflow"
)

// Research Team Agents

type WebResearchAgent struct {
	*core.BaseAgentImpl
}

func NewWebResearchAgent() *WebResearchAgent {
	return &WebResearchAgent{
		BaseAgentImpl: core.NewBaseAgent("webResearcher", "Searches web for current information", domain.AgentTypeLLM),
	}
}

func (w *WebResearchAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	topic, _ := state.Get("topic")
	log.Printf("[Web Research] Researching: %v\n", topic)

	output := domain.NewState()
	output.Set("output", fmt.Sprintf("Web research on %v: Found 10 recent articles, 3 blog posts, and 2 whitepapers", topic))
	output.Set("sources", []string{"techcrunch.com", "arxiv.org", "medium.com"})
	output.Set("research_type", "web")
	return output, nil
}

type AcademicResearchAgent struct {
	*core.BaseAgentImpl
}

func NewAcademicResearchAgent() *AcademicResearchAgent {
	return &AcademicResearchAgent{
		BaseAgentImpl: core.NewBaseAgent("academicResearcher", "Searches academic papers and journals", domain.AgentTypeLLM),
	}
}

func (a *AcademicResearchAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	topic, _ := state.Get("topic")
	log.Printf("[Academic Research] Researching: %v\n", topic)

	output := domain.NewState()
	output.Set("output", fmt.Sprintf("Academic research on %v: Found 5 peer-reviewed papers, 2 dissertations", topic))
	output.Set("sources", []string{"scholar.google.com", "jstor.org", "ieee.org"})
	output.Set("research_type", "academic")
	return output, nil
}

// Analysis Team Agents

type DataAnalyst struct {
	*core.BaseAgentImpl
}

func NewDataAnalyst() *DataAnalyst {
	return &DataAnalyst{
		BaseAgentImpl: core.NewBaseAgent("dataAnalyst", "Analyzes data and statistics", domain.AgentTypeLLM),
	}
}

func (d *DataAnalyst) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	_, _ = state.Get("research_data")
	log.Printf("[Data Analyst] Analyzing research data\n")

	output := domain.NewState()
	output.Set("output", "Data analysis complete: Identified 3 key trends, 2 anomalies, and strong correlation patterns")
	output.Set("analysis_type", "statistical")
	output.Set("confidence", 0.85)
	return output, nil
}

type TrendAnalyst struct {
	*core.BaseAgentImpl
}

func NewTrendAnalyst() *TrendAnalyst {
	return &TrendAnalyst{
		BaseAgentImpl: core.NewBaseAgent("trendAnalyst", "Identifies trends and patterns", domain.AgentTypeLLM),
	}
}

func (t *TrendAnalyst) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	log.Printf("[Trend Analyst] Analyzing trends\n")

	output := domain.NewState()
	output.Set("output", "Trend analysis: Growing adoption (45% YoY), emerging in 3 new sectors, peak interest expected Q3")
	output.Set("analysis_type", "trends")
	output.Set("growth_rate", 0.45)
	return output, nil
}

// Report Writer Agent

type ReportWriter struct {
	*core.BaseAgentImpl
}

func NewReportWriter() *ReportWriter {
	return &ReportWriter{
		BaseAgentImpl: core.NewBaseAgent("reportWriter", "Writes comprehensive reports", domain.AgentTypeLLM),
	}
}

func (r *ReportWriter) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	log.Printf("[Report Writer] Compiling final report\n")

	// Gather all research and analysis
	var sections []string

	// Check for research data
	if webRes, ok := state.Get("web_research"); ok {
		sections = append(sections, fmt.Sprintf("Web Research:\n%v", webRes))
	}
	if acadRes, ok := state.Get("academic_research"); ok {
		sections = append(sections, fmt.Sprintf("\nAcademic Research:\n%v", acadRes))
	}

	// Check for analysis data
	if dataAnalysis, ok := state.Get("data_analysis"); ok {
		sections = append(sections, fmt.Sprintf("\nData Analysis:\n%v", dataAnalysis))
	}
	if trendAnalysis, ok := state.Get("trend_analysis"); ok {
		sections = append(sections, fmt.Sprintf("\nTrend Analysis:\n%v", trendAnalysis))
	}

	report := fmt.Sprintf("=== Comprehensive Report ===\n\n%s\n\n=== Conclusion ===\nBased on multi-source research and analysis, the findings indicate strong potential with measured risks.",
		strings.Join(sections, "\n"))

	output := domain.NewState()
	output.Set("output", report)
	output.Set("report_sections", len(sections))
	output.Set("timestamp", time.Now().Format(time.RFC3339))
	return output, nil
}

// Coordinator Agents

type ResearchCoordinator struct {
	*core.BaseAgentImpl
}

func NewResearchCoordinator() *ResearchCoordinator {
	return &ResearchCoordinator{
		BaseAgentImpl: core.NewBaseAgent("researchCoordinator", "Coordinates research activities", domain.AgentTypeLLM),
	}
}

func (r *ResearchCoordinator) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	topic, _ := state.Get("topic")
	log.Printf("[Research Coordinator] Coordinating research on: %v\n", topic)

	// In a real implementation, this would use the LLM to decide which researchers to activate
	// For demo, we'll activate both
	output := domain.NewState()
	output.Set("output", "Research coordination complete. Activated web and academic researchers.")
	output.Set("research_plan", []string{"web_research", "academic_research"})
	return output, nil
}

type AnalysisCoordinator struct {
	*core.BaseAgentImpl
}

func NewAnalysisCoordinator() *AnalysisCoordinator {
	return &AnalysisCoordinator{
		BaseAgentImpl: core.NewBaseAgent("analysisCoordinator", "Coordinates analysis activities", domain.AgentTypeLLM),
	}
}

func (a *AnalysisCoordinator) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	log.Printf("[Analysis Coordinator] Coordinating analysis\n")

	output := domain.NewState()
	output.Set("output", "Analysis coordination complete. Both data and trend analysis performed.")
	output.Set("analysis_plan", []string{"data_analysis", "trend_analysis"})
	return output, nil
}

func main() {
	log.Println("=== Multi-Agent Coordination Example ===")
	log.Println()

	// Create individual specialist agents
	webResearcher := NewWebResearchAgent()
	academicResearcher := NewAcademicResearchAgent()
	dataAnalyst := NewDataAnalyst()
	trendAnalyst := NewTrendAnalyst()
	reportWriter := NewReportWriter()

	// Create coordinator agents with their teams
	researchCoordinator := NewResearchCoordinator()
	analysisCoordinator := NewAnalysisCoordinator()

	// Register all agents for handoff
	_ = core.Register(webResearcher)
	_ = core.Register(academicResearcher)
	_ = core.Register(dataAnalyst)
	_ = core.Register(trendAnalyst)
	_ = core.Register(reportWriter)
	_ = core.Register(researchCoordinator)
	_ = core.Register(analysisCoordinator)

	// Create research team using workflow
	researchTeam := workflow.NewParallelAgent("researchTeam").
		AddAgent(webResearcher).
		AddAgent(academicResearcher).
		WithMergeStrategy(workflow.MergeAll)

	// Create analysis team using workflow
	analysisTeam := workflow.NewParallelAgent("analysisTeam").
		AddAgent(dataAnalyst).
		AddAgent(trendAnalyst).
		WithMergeStrategy(workflow.MergeAll)

	// Create the main coordinator with sub-agents
	mainCoordinator, err := core.NewLLMAgentWithSubAgentsFromString(
		"mainCoordinator",
		"mock", // In production: "openai/gpt-4"
		researchCoordinator,
		analysisCoordinator,
		reportWriter,
	)
	if err != nil {
		log.Fatal("Failed to create main coordinator:", err)
	}

	// Configure main coordinator
	mainCoordinator.SetSystemPrompt(`You are the main coordinator for comprehensive research and analysis projects.
You have access to:
- researchCoordinator: Manages research activities
- analysisCoordinator: Manages data analysis
- reportWriter: Creates final reports

Coordinate these teams to produce comprehensive insights.`)

	// Example 1: Simple delegation
	fmt.Println("=== Example 1: Direct Delegation ===")
	ctx := context.Background()

	// Research phase
	researchInput := map[string]interface{}{
		"topic": "quantum computing applications in finance",
	}

	researchResult, err := mainCoordinator.TransferTo(ctx, "researchCoordinator",
		"Starting research phase", researchInput)
	if err != nil {
		log.Fatal("Research coordination failed:", err)
	}
	if output, ok := researchResult.Get("output"); ok {
		fmt.Printf("Research Coordinator: %v\n\n", output)
	}

	// Example 2: Full workflow with teams
	fmt.Println("=== Example 2: Full Multi-Team Workflow ===")

	// Create a complete workflow
	fullWorkflow := workflow.NewSequentialAgent("fullWorkflow").
		// Phase 1: Research (parallel)
		AddAgent(researchTeam).
		// Phase 2: Analysis (parallel)
		AddAgent(analysisTeam).
		// Phase 3: Report writing
		AddAgent(reportWriter)

	// Execute the full workflow
	initialState := domain.NewState()
	initialState.Set("topic", "artificial intelligence in healthcare")

	result, err := fullWorkflow.Run(ctx, initialState)
	if err != nil {
		log.Fatal("Workflow failed:", err)
	}

	if report, ok := result.Get("output"); ok {
		fmt.Printf("Final Report:\n%v\n\n", report)
	}

	// Example 3: Hierarchical coordination
	fmt.Println("=== Example 3: Hierarchical Agent Structure ===")

	// Create sub-coordinators with their teams
	researchCoordWithTeam, _ := core.NewLLMAgentWithSubAgentsFromString(
		"researchLead",
		"mock",
		webResearcher,
		academicResearcher,
	)

	analysisCoordWithTeam, _ := core.NewLLMAgentWithSubAgentsFromString(
		"analysisLead",
		"mock",
		dataAnalyst,
		trendAnalyst,
	)

	// Register the coordinators
	_ = core.Register(researchCoordWithTeam)
	_ = core.Register(analysisCoordWithTeam)

	// Create top-level coordinator
	topCoordinator, _ := core.NewLLMAgentWithSubAgentsFromString(
		"topCoordinator",
		"mock",
		researchCoordWithTeam,
		analysisCoordWithTeam,
		reportWriter,
	)

	fmt.Println("Hierarchical Structure Created:")
	fmt.Println("- Top Coordinator")
	fmt.Println("  - Research Lead")
	fmt.Println("    - Web Researcher")
	fmt.Println("    - Academic Researcher")
	fmt.Println("  - Analysis Lead")
	fmt.Println("    - Data Analyst")
	fmt.Println("    - Trend Analyst")
	fmt.Println("  - Report Writer")
	fmt.Println()

	// Show available tools at top level
	fmt.Println("Top Coordinator's Direct Tools:")
	for _, tool := range topCoordinator.ListTools() {
		if t, ok := topCoordinator.GetTool(tool); ok {
			fmt.Printf("- %s: %s\n", tool, t.Description())
		}
	}
	fmt.Println()

	// Example 4: Conditional multi-path
	fmt.Println("=== Example 4: Conditional Multi-Path Execution ===")

	// Create conditional workflow based on initial assessment
	conditionalWorkflow := workflow.NewConditionalAgent("conditionalWorkflow").
		AddAgent("simple", func(state *domain.State) bool {
			complexity, _ := state.Get("complexity")
			return complexity == "simple"
		}, reportWriter).
		AddAgent("complex", func(state *domain.State) bool {
			complexity, _ := state.Get("complexity")
			return complexity == "complex"
		}, fullWorkflow)

	// Simulate different complexity levels
	simpleState := domain.NewState()
	simpleState.Set("complexity", "simple")
	simpleState.Set("topic", "basic AI overview")

	complexState := domain.NewState()
	complexState.Set("complexity", "complex")
	complexState.Set("topic", "AI ethics and regulation")

	fmt.Println("Processing simple request...")
	simpleResult, _ := conditionalWorkflow.Run(ctx, simpleState)
	if output, ok := simpleResult.Get("output"); ok {
		fmt.Printf("Simple path result: %v\n\n", output)
	}

	fmt.Println("=== Key Patterns Demonstrated ===")
	fmt.Println("1. Hierarchical agent structures with multi-level coordination")
	fmt.Println("2. Parallel execution teams working concurrently")
	fmt.Println("3. Sequential phases with automatic result passing")
	fmt.Println("4. Conditional routing based on request analysis")
	fmt.Println("5. Automatic tool registration for all sub-agents")
	fmt.Println("6. State sharing across agent hierarchies")
}
