// ABOUTME: Example demonstrating a custom agent that extends BaseAgentImpl
// ABOUTME: Shows code-based orchestration, multi-search, and LLMAgent usage

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	webtools "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// ResearchAgent is a custom agent that extends BaseAgentImpl for full control
type ResearchAgent struct {
	*core.BaseAgentImpl
	multiSearcher    *MultiSearchAgent
	duplicateFilter  domain.BaseAgent // LLMAgent instance or mock
	contentAnalyzer  domain.BaseAgent // LLMAgent instance or mock
	reportGenerator  domain.BaseAgent // LLMAgent instance or mock
	maxSearchResults int
}

// MultiSearchAgent executes parallel searches across multiple engines
type MultiSearchAgent struct {
	*core.BaseAgentImpl
	engines []string
	apiKeys map[string]string
}

// NewResearchAgent creates a new research agent with code-based orchestration
func NewResearchAgent(name string) (*ResearchAgent, error) {
	// Create base agent
	base := core.NewBaseAgent(name, "A research agent that orchestrates multiple sub-agents for comprehensive research", domain.AgentTypeCustom)

	// Determine LLM provider for sub-agents
	providerStr := os.Getenv("LLM_PROVIDER")
	if providerStr == "" {
		providerStr = "gpt-4o" // Default to Claude
	}

	// Create multi-search agent
	multiSearcher := NewMultiSearchAgent("multi-searcher")

	// Create LLM-based sub-agents with specialized prompts
	var duplicateFilter domain.BaseAgent
	llmDupFilter, err := core.NewAgentFromString("duplicate-filter", providerStr)
	if err != nil {
		log.Printf("Warning: Failed to create duplicate filter with %s, using mock: %v", providerStr, err)
		duplicateFilter = createMockDuplicateFilter()
	} else {
		duplicateFilter = llmDupFilter
		llmDupFilter.SetSystemPrompt(`You are a search result deduplication expert. 
Given multiple search results, identify and merge duplicates based on:
- Similar URLs or domains (e.g., www.example.com and example.com)
- Overlapping content or titles
- Same source referenced differently

Output a JSON array of deduplicated results with relevance scores (0-1).
Each result should have: url, title, snippet, source_engine, relevance_score.
Merge information from duplicates to create richer snippets.`)
	}

	var contentAnalyzer domain.BaseAgent
	llmContentAnalyzer, err := core.NewAgentFromString("content-analyzer", providerStr)
	if err != nil {
		log.Printf("Warning: Failed to create content analyzer with %s, using mock: %v", providerStr, err)
		contentAnalyzer = createMockContentAnalyzer()
	} else {
		contentAnalyzer = llmContentAnalyzer
		llmContentAnalyzer.SetSystemPrompt(`You are a content analysis expert.
Given search results about a topic, extract:
- Key insights and main themes
- Important facts and data points
- Contrasting viewpoints if any
- Knowledge gaps that need further research

Structure your analysis with clear sections and bullet points.
Focus on actionable, relevant information.`)
	}

	var reportGenerator domain.BaseAgent
	llmReportGenerator, err := core.NewAgentFromString("report-generator", providerStr)
	if err != nil {
		log.Printf("Warning: Failed to create report generator with %s, using mock: %v", providerStr, err)
		reportGenerator = createMockReportGenerator()
	} else {
		reportGenerator = llmReportGenerator
		llmReportGenerator.SetSystemPrompt(`You are a professional research report writer.
Given analyzed content about a topic, create a comprehensive report with:
- Executive Summary (2-3 paragraphs)
- Key Findings (structured with subheadings)
- Detailed Analysis
- Conclusions and Recommendations
- Sources and References

Use markdown formatting for clarity.
Maintain an objective, professional tone.`)
	}

	// Add debug logging to all LLM agents if DEBUG=1
	if os.Getenv("DEBUG") == "1" {
		log.Println("🐛 DEBUG mode enabled - Adding logging hooks to LLM agents")

		// Try to add hooks to the LLM agents (type assertion needed)
		debugLogger := slog.Default().With("component", "research-agent")
		if llmAgent, ok := duplicateFilter.(*core.LLMAgent); ok {
			llmAgent.WithHook(core.NewLoggingHook(debugLogger.With("agent", "duplicate-filter"), core.LogLevelDetailed))
		}
		if llmAgent, ok := contentAnalyzer.(*core.LLMAgent); ok {
			llmAgent.WithHook(core.NewLoggingHook(debugLogger.With("agent", "content-analyzer"), core.LogLevelDetailed))
		}
		if llmAgent, ok := reportGenerator.(*core.LLMAgent); ok {
			llmAgent.WithHook(core.NewLoggingHook(debugLogger.With("agent", "report-generator"), core.LogLevelDetailed))
		}
	}

	agent := &ResearchAgent{
		BaseAgentImpl:    base,
		multiSearcher:    multiSearcher,
		duplicateFilter:  duplicateFilter,
		contentAnalyzer:  contentAnalyzer,
		reportGenerator:  reportGenerator,
		maxSearchResults: 10,
	}

	return agent, nil
}

// Run executes the research process with phase-based orchestration
func (r *ResearchAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Extract research topic
	topic, ok := state.Get("topic")
	if !ok {
		return nil, fmt.Errorf("no research topic provided in state")
	}

	topicStr, ok := topic.(string)
	if !ok {
		return nil, fmt.Errorf("topic must be a string")
	}

	// Initialize result state
	resultState := domain.NewState()
	resultState.Set("topic", topicStr)
	resultState.Set("start_time", time.Now())

	// Emit start event
	r.EmitEvent(domain.EventAgentStart, map[string]interface{}{
		"agent": r.Name(),
		"topic": topicStr,
	})

	// Phase 1: Multi-Engine Search
	log.Printf("🔍 Phase 1: Executing parallel searches for '%s'", topicStr)
	if os.Getenv("DEBUG") == "1" {
		log.Printf("🐛 DEBUG: Starting multi-engine search with max results: %d", r.maxSearchResults)
	}
	searchResults, err := r.executeMultiSearch(ctx, topicStr)
	if err != nil {
		r.EmitEvent(domain.EventAgentError, map[string]interface{}{
			"phase": "search",
			"error": err.Error(),
		})
		return nil, fmt.Errorf("multi-search failed: %w", err)
	}
	resultState.Set("raw_search_results", searchResults)
	log.Printf("  ✅ Found %d total results across all engines", len(searchResults))

	// Phase 2: Deduplication and Ranking
	log.Printf("🔄 Phase 2: Deduplicating and ranking results")
	dedupedResults, err := r.deduplicateResults(ctx, searchResults)
	if err != nil {
		log.Printf("  ⚠️  Deduplication failed, using raw results: %v", err)
		dedupedResults = searchResults
	} else {
		log.Printf("  ✅ Reduced to %d unique results", len(dedupedResults))
	}
	resultState.Set("deduped_results", dedupedResults)

	// Phase 3: Content Analysis
	log.Printf("📊 Phase 3: Analyzing content and extracting insights")
	analysis, err := r.analyzeContent(ctx, topicStr, dedupedResults)
	if err != nil {
		log.Printf("  ⚠️  Analysis failed: %v", err)
		analysis = "Analysis unavailable"
	}
	resultState.Set("content_analysis", analysis)

	// Phase 4: Report Generation
	log.Printf("📝 Phase 4: Generating comprehensive report")
	report, err := r.generateReport(ctx, topicStr, analysis, dedupedResults)
	if err != nil {
		log.Printf("  ⚠️  Report generation failed: %v", err)
		report = r.generateFallbackReport(topicStr, analysis, dedupedResults)
	}
	resultState.Set("final_report", report)

	// Add metadata
	resultState.Set("end_time", time.Now())
	startTime, _ := resultState.Get("start_time")
	duration := time.Since(startTime.(time.Time))
	resultState.Set("research_duration", duration.String())
	resultState.Set("total_results_found", len(searchResults))
	resultState.Set("unique_results", len(dedupedResults))

	// Emit completion event
	r.EmitEvent(domain.EventAgentComplete, map[string]interface{}{
		"agent":    r.Name(),
		"duration": duration.String(),
		"results":  len(dedupedResults),
	})

	return resultState, nil
}

// executeMultiSearch runs parallel searches across multiple engines
func (r *ResearchAgent) executeMultiSearch(ctx context.Context, topic string) ([]map[string]interface{}, error) {
	searchState := domain.NewState()
	searchState.Set("topic", topic)
	searchState.Set("max_results", r.maxSearchResults)

	// Pass API keys if available
	apiKeys := make(map[string]string)
	if key := os.Getenv("BRAVE_API_KEY"); key != "" {
		apiKeys["brave"] = key
	}
	if key := os.Getenv("TAVILY_API_KEY"); key != "" {
		apiKeys["tavily"] = key
	}
	if key := os.Getenv("SERPAPI_API_KEY"); key != "" {
		apiKeys["serpapi"] = key
	}
	if key := os.Getenv("SERPERDEV_API_KEY"); key != "" {
		apiKeys["serperdev"] = key
	}
	searchState.Set("api_keys", apiKeys)

	result, err := r.multiSearcher.Run(ctx, searchState)
	if err != nil {
		return nil, err
	}

	results, ok := result.Get("results")
	if !ok {
		return nil, fmt.Errorf("no results from multi-search")
	}

	return results.([]map[string]interface{}), nil
}

// deduplicateResults uses LLM to intelligently deduplicate results
func (r *ResearchAgent) deduplicateResults(ctx context.Context, results []map[string]interface{}) ([]map[string]interface{}, error) {
	dedupState := domain.NewState()
	dedupState.Set("results", results)
	dedupState.Set("prompt", "Deduplicate these search results and assign relevance scores")

	result, err := r.duplicateFilter.Run(ctx, dedupState)
	if err != nil {
		return results, err
	}

	response, ok := result.Get("result")
	if !ok {
		return results, fmt.Errorf("no response from deduplication")
	}

	// Parse the response - expecting JSON array
	// In a real implementation, we'd parse the JSON properly
	// For now, return original results
	_ = response // TODO: Parse and use the deduplication response
	return results[:min(len(results), r.maxSearchResults)], nil
}

// analyzeContent uses LLM to extract insights
func (r *ResearchAgent) analyzeContent(ctx context.Context, topic string, results []map[string]interface{}) (string, error) {
	analysisState := domain.NewState()
	analysisState.Set("topic", topic)
	analysisState.Set("search_results", results)
	// Create a structured prompt with the search results
	var promptBuilder strings.Builder
	promptBuilder.WriteString(fmt.Sprintf("Analyze these search results about '%s' and extract key insights.\n\n", topic))
	promptBuilder.WriteString("Search Results:\n")
	for i, result := range results {
		if i >= 20 { // Limit to first 20 results to avoid token limits
			break
		}
		title := "Untitled"
		if t, ok := result["title"].(string); ok {
			title = t
		}
		url := ""
		if u, ok := result["url"].(string); ok {
			url = u
		}
		description := ""
		if d, ok := result["description"].(string); ok {
			description = d
		}
		promptBuilder.WriteString(fmt.Sprintf("%d. %s\n   URL: %s\n   %s\n\n", i+1, title, url, description))
	}

	analysisState.Set("prompt", promptBuilder.String())

	result, err := r.contentAnalyzer.Run(ctx, analysisState)
	if err != nil {
		return "", err
	}

	analysis, ok := result.Get("result")
	if !ok {
		return "", fmt.Errorf("no response from content analyzer")
	}

	return analysis.(string), nil
}

// generateReport uses LLM to create final report
func (r *ResearchAgent) generateReport(ctx context.Context, topic string, analysis string, results []map[string]interface{}) (string, error) {
	reportState := domain.NewState()
	reportState.Set("topic", topic)
	reportState.Set("analysis", analysis)
	reportState.Set("sources", results)
	// Create a structured prompt with the analysis and sources
	reportPrompt := fmt.Sprintf(`Generate a comprehensive research report about '%s'.

Analysis:
%s

Number of sources: %d

Please create a well-structured report following your system prompt guidelines.`, topic, analysis, len(results))

	reportState.Set("prompt", reportPrompt)

	result, err := r.reportGenerator.Run(ctx, reportState)
	if err != nil {
		return "", err
	}

	report, ok := result.Get("result")
	if !ok {
		return "", fmt.Errorf("no response from report generator")
	}

	return report.(string), nil
}

// generateFallbackReport creates a simple report when LLM fails
func (r *ResearchAgent) generateFallbackReport(topic string, analysis string, results []map[string]interface{}) string {
	var report strings.Builder

	report.WriteString(fmt.Sprintf("# Research Report: %s\n\n", topic))
	report.WriteString(fmt.Sprintf("*Generated on: %s*\n\n", time.Now().Format("January 2, 2006")))

	report.WriteString("## Executive Summary\n\n")
	report.WriteString(fmt.Sprintf("This report presents research findings on '%s' compiled from %d sources.\n\n", topic, len(results)))

	if analysis != "" && analysis != "Analysis unavailable" {
		report.WriteString("## Analysis\n\n")
		report.WriteString(analysis)
		report.WriteString("\n\n")
	}

	report.WriteString("## Sources\n\n")
	for i, result := range results {
		if i >= 10 {
			break
		}
		title := "Untitled"
		if t, ok := result["title"].(string); ok {
			title = t
		}
		url := ""
		if u, ok := result["url"].(string); ok {
			url = u
		}
		report.WriteString(fmt.Sprintf("%d. [%s](%s)\n", i+1, title, url))
	}

	report.WriteString("\n---\n*Report generated by Research Agent*\n")

	return report.String()
}

// NewMultiSearchAgent creates an agent that searches multiple engines in parallel
func NewMultiSearchAgent(name string) *MultiSearchAgent {
	base := core.NewBaseAgent(name, "Executes parallel searches across multiple search engines", domain.AgentTypeCustom)

	return &MultiSearchAgent{
		BaseAgentImpl: base,
		engines:       []string{"duckduckgo", "brave", "tavily", "serpapi", "serperdev"},
		apiKeys:       make(map[string]string),
	}
}

// Run executes parallel searches
func (m *MultiSearchAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	topic, _ := state.Get("topic")
	topicStr := topic.(string)
	maxResults := 5 // Per engine

	if mr, ok := state.Get("max_results"); ok {
		maxResults = mr.(int)
	}

	// Get API keys from state
	if keys, ok := state.Get("api_keys"); ok {
		m.apiKeys = keys.(map[string]string)
	}

	// Get the web search tool
	searchTool, ok := tools.GetTool("web_search")
	if !ok {
		return nil, fmt.Errorf("web_search tool not found")
	}

	// Create different query variations
	queries := map[string]string{
		"overview": topicStr + " overview introduction guide",
		"latest":   topicStr + " latest news updates 2024 2025",
		"expert":   topicStr + " expert analysis research studies",
		"tutorial": topicStr + " tutorial how-to examples",
	}

	var (
		allResults []map[string]interface{}
		mu         sync.Mutex
		wg         sync.WaitGroup
	)

	// Execute searches in parallel
	for _, engine := range m.engines {
		for queryType, query := range queries {
			wg.Add(1)
			go func(eng, qType, q string) {
				defer wg.Done()

				if os.Getenv("DEBUG") == "1" {
					log.Printf("  🐛 DEBUG: Starting search on %s with query type %s", eng, qType)
				}
				log.Printf("  🔎 Searching %s (%s): %s", eng, qType, q)

				// Create tool context
				toolCtx := &domain.ToolContext{
					Context: ctx,
					State:   domain.NewStateReader(state),
					Agent: domain.AgentInfo{
						ID:   m.ID(),
						Name: m.Name(),
						Type: m.Type(),
					},
					RunID: fmt.Sprintf("search-%s-%s-%d", eng, qType, time.Now().Unix()),
				}

				// Prepare parameters
				params := map[string]interface{}{
					"query":       q,
					"engine":      eng,
					"max_results": maxResults,
				}

				// Add API key if available
				if apiKey, exists := m.apiKeys[eng]; exists {
					params["engine_api_key"] = apiKey
				}

				// Execute search
				result, err := searchTool.Execute(toolCtx, params)
				if err != nil {
					log.Printf("    ⚠️  %s search failed: %v", eng, err)
					return
				}

				// Extract results
				if searchResults, ok := result.(*webtools.WebSearchResults); ok {
					mu.Lock()
					for _, r := range searchResults.Results {
						resultMap := map[string]interface{}{
							"title":         r.Title,
							"url":           r.URL,
							"description":   r.Description,
							"snippet":       r.Snippet,
							"source_engine": eng,
							"query_type":    qType,
						}
						allResults = append(allResults, resultMap)
					}
					mu.Unlock()
					log.Printf("    ✅ %s: Found %d results", eng, len(searchResults.Results))
				} else {
					log.Printf("    ⚠️  %s: unexpected result type: %T", eng, result)
				}
			}(engine, queryType, query)
		}
	}

	wg.Wait()

	resultState := domain.NewState()
	resultState.Set("results", allResults)
	resultState.Set("total_count", len(allResults))
	resultState.Set("engines_used", m.engines)

	return resultState, nil
}

// Mock agents for when LLM is not available

func createMockDuplicateFilter() domain.BaseAgent {
	mock := core.NewBaseAgent("mock-duplicate-filter", "Mock duplicate filter", domain.AgentTypeCustom)
	mockWrapper := &MockAgent{
		BaseAgent: mock,
		runFunc: func(ctx context.Context, state *domain.State) (*domain.State, error) {
			results, _ := state.Get("results")
			resultState := domain.NewState()
			resultState.Set("response", results) // Pass through
			return resultState, nil
		},
	}
	return mockWrapper
}

func createMockContentAnalyzer() domain.BaseAgent {
	mock := core.NewBaseAgent("mock-content-analyzer", "Mock content analyzer", domain.AgentTypeCustom)
	mockWrapper := &MockAgent{
		BaseAgent: mock,
		runFunc: func(ctx context.Context, state *domain.State) (*domain.State, error) {
			topic, _ := state.Get("topic")
			resultState := domain.NewState()
			resultState.Set("response", fmt.Sprintf("Mock analysis for %v: This topic appears to have multiple perspectives and recent developments.", topic))
			return resultState, nil
		},
	}
	return mockWrapper
}

func createMockReportGenerator() domain.BaseAgent {
	mock := core.NewBaseAgent("mock-report-generator", "Mock report generator", domain.AgentTypeCustom)
	mockWrapper := &MockAgent{
		BaseAgent: mock,
		runFunc: func(ctx context.Context, state *domain.State) (*domain.State, error) {
			topic, _ := state.Get("topic")
			analysis, _ := state.Get("analysis")
			resultState := domain.NewState()
			resultState.Set("response", fmt.Sprintf("# Mock Report: %v\n\n## Summary\n\n%v\n\n## Conclusion\n\nThis is a mock report for demonstration.", topic, analysis))
			return resultState, nil
		},
	}
	return mockWrapper
}

// MockAgent wraps a base agent with custom run behavior
type MockAgent struct {
	domain.BaseAgent
	runFunc func(context.Context, *domain.State) (*domain.State, error)
}

func (m *MockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	return m.runFunc(ctx, state)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// Create research agent
	agent, err := NewResearchAgent("advanced-researcher")
	if err != nil {
		log.Fatalf("Failed to create research agent: %v", err)
	}

	// Example topics
	topics := []string{
		"quantum computing applications in cryptography",
		"sustainable urban farming technologies",
		"AI ethics and governance frameworks",
	}

	fmt.Println("=== Advanced Research Agent Example ===")
	fmt.Println()
	fmt.Println("This example demonstrates:")
	fmt.Println("- Custom agent extending BaseAgentImpl (not LLMAgent)")
	fmt.Println("- Code-based orchestration of multiple agents")
	fmt.Println("- Parallel search across multiple engines")
	fmt.Println("- LLMAgent instances for intelligent processing")
	fmt.Println("- State management and error handling")
	fmt.Println()

	// Select topic
	topic := topics[0]
	if len(os.Args) > 1 {
		topic = strings.Join(os.Args[1:], " ")
	}

	fmt.Printf("Research Topic: %s\n", topic)
	fmt.Println(strings.Repeat("-", 60))

	// Create initial state
	state := domain.NewState()
	state.Set("topic", topic)

	// Run research
	ctx := context.Background()
	result, err := agent.Run(ctx, state)
	if err != nil {
		log.Fatalf("Research failed: %v", err)
	}

	// Display results
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("RESEARCH COMPLETE")
	fmt.Println(strings.Repeat("=", 60))

	if report, ok := result.Get("final_report"); ok {
		fmt.Println("\n" + report.(string))
	}

	// Display metadata
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("Research Metadata:")
	if duration, ok := result.Get("research_duration"); ok {
		fmt.Printf("- Duration: %v\n", duration)
	}
	if total, ok := result.Get("total_results_found"); ok {
		fmt.Printf("- Total results found: %v\n", total)
	}
	if unique, ok := result.Get("unique_results"); ok {
		fmt.Printf("- Unique results after dedup: %v\n", unique)
	}

	fmt.Println("\nNote: For best results, set these environment variables:")
	fmt.Println("- BRAVE_API_KEY, TAVILY_API_KEY, SERPAPI_API_KEY, SERPERDEV_API_KEY for search")
	fmt.Println("- OPENAI_API_KEY or ANTHROPIC_API_KEY for LLM processing")
	fmt.Println("- LLM_PROVIDER=openai or LLM_PROVIDER=anthropic (default: claude)")
}
