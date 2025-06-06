// ABOUTME: Example demonstrating a custom agent that extends BaseAgent
// ABOUTME: Shows sub-agent coordination, tool usage, and state management

package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	builtintools "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
	"github.com/lexlapax/go-llms/pkg/agent/core"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
	"github.com/lexlapax/go-llms/pkg/util/llmutil"
)

// ResearchAssistant is a custom agent that conducts research on topics
type ResearchAssistant struct {
	*core.LLMAgent
	webSearcher domain.BaseAgent
	summarizer  domain.BaseAgent
	factChecker domain.BaseAgent
	maxSources  int
}

// NewResearchAssistant creates a new research assistant agent
func NewResearchAssistant(name string) (*ResearchAssistant, error) {
	// Create a provider for the base LLM agent
	llmProvider, providerName, modelName, err := llmutil.ProviderFromEnv()
	if err != nil {
		log.Printf("No LLM provider found from env, using mock: %v", err)
		// Use mock provider if no env vars set
		llmProvider = provider.NewMockProvider()
		providerName = "mock"
		modelName = "mock"
	} else {
		log.Printf("Using LLM provider: %s with model: %s", providerName, modelName)
	}

	// Create LLM agent as base
	base := core.NewAgent(name, llmProvider)
	base.SetSystemPrompt("You are a research assistant that gathers and synthesizes information")

	// Create sub-agents
	webSearcher, err := createWebSearchAgent()
	if err != nil {
		return nil, fmt.Errorf("failed to create web searcher: %w", err)
	}

	summarizer, err := createSummarizerAgent()
	if err != nil {
		return nil, fmt.Errorf("failed to create summarizer: %w", err)
	}

	factChecker, err := createFactCheckerAgent()
	if err != nil {
		return nil, fmt.Errorf("failed to create fact checker: %w", err)
	}

	agent := &ResearchAssistant{
		LLMAgent:    base,
		webSearcher: webSearcher,
		summarizer:  summarizer,
		factChecker: factChecker,
		maxSources:  5,
	}

	// Register tools
	if webSearch, ok := builtintools.GetTool("web_search"); ok {
		base.AddTool(webSearch)
	}
	if webFetch, ok := builtintools.GetTool("web_fetch"); ok {
		base.AddTool(webFetch)
	}

	return agent, nil
}

// Run executes the research process
func (r *ResearchAssistant) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	// Extract research topic
	topic, ok := state.Get("topic")
	if !ok {
		return nil, fmt.Errorf("no research topic provided in state")
	}

	topicStr, ok := topic.(string)
	if !ok {
		return nil, fmt.Errorf("topic must be a string")
	}

	// Create run context
	runCtx := domain.NewRunContextWithState[any](ctx, nil, state)
	runCtx.RunID = fmt.Sprintf("research-%d", time.Now().Unix())

	// Initialize research state
	resultState := domain.NewState()
	resultState.Set("topic", topicStr)
	resultState.Set("research_notes", []string{})
	resultState.Set("sources", []string{})
	resultState.Set("start_time", time.Now())

	// Phase 1: Web Search
	log.Printf("🔍 Phase 1: Searching for information about '%s'", topicStr)
	searchResults, err := r.performWebSearch(runCtx, topicStr)
	if err != nil {
		return nil, fmt.Errorf("web search failed: %w", err)
	}
	resultState.Set("search_results", searchResults)

	// Phase 2: Gather Information from Sources
	log.Printf("📚 Phase 2: Gathering information from %d sources", len(searchResults))
	articles, err := r.gatherInformation(runCtx, searchResults)
	if err != nil {
		log.Printf("Warning: Some sources could not be fetched: %v", err)
	}
	resultState.Set("raw_articles", articles)

	// Phase 3: Summarize Each Article
	log.Printf("📝 Phase 3: Summarizing %d articles", len(articles))
	summaries, err := r.summarizeArticles(runCtx, articles)
	if err != nil {
		return nil, fmt.Errorf("summarization failed: %w", err)
	}
	resultState.Set("article_summaries", summaries)

	// Phase 4: Fact Check Key Claims
	log.Printf("✅ Phase 4: Fact-checking key claims")
	checkedFacts, err := r.factCheckClaims(runCtx, summaries)
	if err != nil {
		log.Printf("Warning: Fact checking partially failed: %v", err)
		checkedFacts = summaries // Fallback to unchecked summaries
	}
	resultState.Set("fact_checked_info", checkedFacts)

	// Phase 5: Synthesize Final Report
	log.Printf("📊 Phase 5: Synthesizing final report")
	report := r.synthesizeReport(topicStr, checkedFacts, searchResults)
	resultState.Set("final_report", report)

	// Add metadata
	resultState.Set("end_time", time.Now())
	startTime, _ := resultState.Get("start_time")
	duration := time.Since(startTime.(time.Time))
	resultState.Set("research_duration", duration.String())
	resultState.Set("sources_count", len(searchResults))

	return resultState, nil
}

// performWebSearch uses the web search tool to find relevant sources
func (r *ResearchAssistant) performWebSearch(ctx *domain.RunContext[any], topic string) ([]string, error) {
	searchTool, ok := r.GetTool("web_search")
	if !ok {
		return nil, fmt.Errorf("web_search tool not found")
	}

	// Create tool context
	toolCtx := &domain.ToolContext{
		Context: ctx.Context(),
		State:   domain.NewStateReader(ctx.State),
		Agent: domain.AgentInfo{
			ID:          r.ID(),
			Name:        r.Name(),
			Description: r.Description(),
			Type:        r.Type(),
			Metadata:    r.Metadata(),
		},
		RunID: ctx.RunID,
	}

	// Execute search
	searchParams := map[string]interface{}{
		"query": topic + " comprehensive guide overview",
	}
	log.Printf("Executing web_search with params: %+v", searchParams)

	result, err := searchTool.Execute(toolCtx, searchParams)
	if err != nil {
		log.Printf("Web search failed: %v", err)
		return nil, err
	}

	log.Printf("Web search result type: %T, value: %+v", result, result)

	// Extract URLs from results
	urls := []string{}
	if searchResults, ok := result.(*web.WebSearchResults); ok {
		for i, res := range searchResults.Results {
			if i >= r.maxSources {
				break
			}
			// Skip the AI summary from Tavily
			if strings.HasPrefix(res.URL, "tavily:answer:") {
				continue
			}
			urls = append(urls, res.URL)
		}
	} else if searchResult, ok := result.(map[string]interface{}); ok {
		if results, ok := searchResult["results"].([]interface{}); ok {
			for i, result := range results {
				if i >= r.maxSources {
					break
				}
				if resultMap, ok := result.(map[string]interface{}); ok {
					if url, ok := resultMap["url"].(string); ok {
						urls = append(urls, url)
					}
				}
			}
		}
	}

	return urls, nil
}

// gatherInformation fetches content from URLs
func (r *ResearchAssistant) gatherInformation(ctx *domain.RunContext[any], urls []string) ([]map[string]string, error) {
	fetchTool, ok := r.GetTool("web_fetch")
	if !ok {
		return nil, fmt.Errorf("web_fetch tool not found")
	}

	articles := []map[string]string{}

	for _, url := range urls {
		log.Printf("  📄 Fetching: %s", url)

		toolCtx := &domain.ToolContext{
			Context: ctx.Context(),
			State:   domain.NewStateReader(ctx.State),
			Agent: domain.AgentInfo{
				ID:          r.ID(),
				Name:        r.Name(),
				Description: r.Description(),
				Type:        r.Type(),
				Metadata:    r.Metadata(),
			},
			RunID: ctx.RunID,
		}

		result, err := fetchTool.Execute(toolCtx, map[string]interface{}{
			"url":    url,
			"prompt": "Extract the main content, key points, and important facts from this page",
		})

		if err != nil {
			log.Printf("  ⚠️  Failed to fetch %s: %v", url, err)
			continue
		}

		if content, ok := result.(string); ok {
			articles = append(articles, map[string]string{
				"url":     url,
				"content": content,
			})
		}
	}

	return articles, nil
}

// summarizeArticles uses the summarizer sub-agent
func (r *ResearchAssistant) summarizeArticles(ctx *domain.RunContext[any], articles []map[string]string) ([]string, error) {
	summaries := []string{}

	for _, article := range articles {
		summaryState := domain.NewState()
		summaryState.Set("text", article["content"])
		summaryState.Set("source_url", article["url"])

		result, err := r.summarizer.Run(ctx.Context(), summaryState)
		if err != nil {
			log.Printf("  ⚠️  Failed to summarize article from %s: %v", article["url"], err)
			continue
		}

		if summary, ok := result.Get("summary"); ok {
			summaries = append(summaries, fmt.Sprintf("Source: %s\n%v", article["url"], summary))
		}
	}

	return summaries, nil
}

// factCheckClaims uses the fact checker sub-agent
func (r *ResearchAssistant) factCheckClaims(ctx *domain.RunContext[any], summaries []string) ([]string, error) {
	checkedInfo := []string{}

	allContent := strings.Join(summaries, "\n\n")

	checkState := domain.NewState()
	checkState.Set("content", allContent)
	checkState.Set("check_contradictions", true)
	checkState.Set("verify_sources", true)

	result, err := r.factChecker.Run(ctx.Context(), checkState)
	if err != nil {
		return summaries, err // Return original if fact checking fails
	}

	if checked, ok := result.Get("verified_content"); ok {
		if checkedStr, ok := checked.(string); ok {
			checkedInfo = strings.Split(checkedStr, "\n\n")
		}
	}

	if len(checkedInfo) == 0 {
		return summaries, nil
	}

	return checkedInfo, nil
}

// synthesizeReport creates the final research report
func (r *ResearchAssistant) synthesizeReport(topic string, information []string, sources []string) string {
	var report strings.Builder

	report.WriteString(fmt.Sprintf("# Research Report: %s\n\n", topic))
	report.WriteString(fmt.Sprintf("*Generated on: %s*\n\n", time.Now().Format("January 2, 2006")))

	report.WriteString("## Executive Summary\n\n")
	report.WriteString(fmt.Sprintf("This report synthesizes information from %d sources about %s.\n\n", len(sources), topic))

	report.WriteString("## Key Findings\n\n")
	for i, info := range information {
		report.WriteString(fmt.Sprintf("### Finding %d\n\n", i+1))
		report.WriteString(info)
		report.WriteString("\n\n")
	}

	report.WriteString("## Sources\n\n")
	for i, source := range sources {
		report.WriteString(fmt.Sprintf("%d. %s\n", i+1, source))
	}

	report.WriteString("\n---\n*Report generated by Research Assistant Agent*\n")

	return report.String()
}

// Helper functions to create sub-agents

func createWebSearchAgent() (domain.BaseAgent, error) {
	// This would typically use an LLM agent, but for demo we'll use a simple wrapper
	mockProvider := provider.NewMockProvider()
	agent := core.NewAgent("web-searcher", mockProvider)
	agent.SetSystemPrompt("You search the web for relevant information")

	if searchTool, ok := builtintools.GetTool("web_search"); ok {
		agent.AddTool(searchTool)
	}

	return agent, nil
}

func createSummarizerAgent() (domain.BaseAgent, error) {
	llmProvider, _, _, err := llmutil.ProviderFromEnv()
	if err != nil {
		// Use a mock summarizer for demo
		return createMockSummarizerAgent(), nil
	}

	agent := core.NewAgent("summarizer", llmProvider)
	agent.SetSystemPrompt("You are an expert at summarizing articles. Extract the key points, main arguments, and important facts. Be concise but comprehensive.")

	return agent, nil
}

func createFactCheckerAgent() (domain.BaseAgent, error) {
	llmProvider, _, _, err := llmutil.ProviderFromEnv()
	if err != nil {
		// Use a mock fact checker for demo
		return createMockFactCheckerAgent(), nil
	}

	agent := core.NewAgent("fact-checker", llmProvider)
	agent.SetSystemPrompt("You are a fact-checking expert. Review the provided information, identify any contradictions, verify claims when possible, and flag any dubious statements. Preserve accurate information while noting concerns.")

	return agent, nil
}

// Mock agents for demonstration when no LLM is available

func createMockSummarizerAgent() domain.BaseAgent {
	mockProvider := provider.NewMockProvider()
	agent := core.NewAgent("mock-summarizer", mockProvider)
	agent.SetSystemPrompt("Mock summarizer for demo")

	// Override Run method
	mockAgent := &MockAgent{
		LLMAgent: agent,
		runFunc: func(ctx context.Context, state *domain.State) (*domain.State, error) {
			text, _ := state.Get("text")
			textStr := fmt.Sprintf("%v", text)
			if len(textStr) > 100 {
				textStr = textStr[:100]
			}
			result := domain.NewState()
			result.Set("summary", fmt.Sprintf("Summary: %s... [truncated for demo]", textStr))
			return result, nil
		},
	}

	return mockAgent
}

func createMockFactCheckerAgent() domain.BaseAgent {
	mockProvider := provider.NewMockProvider()
	agent := core.NewAgent("mock-fact-checker", mockProvider)
	agent.SetSystemPrompt("Mock fact checker for demo")

	mockAgent := &MockAgent{
		LLMAgent: agent,
		runFunc: func(ctx context.Context, state *domain.State) (*domain.State, error) {
			content, _ := state.Get("content")
			result := domain.NewState()
			result.Set("verified_content", fmt.Sprintf("✓ Verified: %v", content))
			return result, nil
		},
	}

	return mockAgent
}

// MockAgent is a simple mock agent for testing
type MockAgent struct {
	*core.LLMAgent
	runFunc func(context.Context, *domain.State) (*domain.State, error)
}

func (m *MockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
	return m.runFunc(ctx, state)
}

// Main function

func main() {
	// Create research assistant
	assistant, err := NewResearchAssistant("research-assistant")
	if err != nil {
		log.Fatalf("Failed to create research assistant: %v", err)
	}

	// Example research topics
	topics := []string{
		"artificial intelligence trends 2025",
		"quantum computing applications",
		"sustainable energy solutions",
	}

	fmt.Println("=== Research Assistant Agent Example ===")
	fmt.Println()
	fmt.Println("This example demonstrates a custom agent that:")
	fmt.Println("- Extends LLMAgent (which implements BaseAgent)")
	fmt.Println("- Uses multiple tools (web search, web fetch)")
	fmt.Println("- Coordinates sub-agents (summarizer, fact checker)")
	fmt.Println("- Manages complex state throughout the research process")
	fmt.Println("- Produces a synthesized research report")
	fmt.Println()

	// Select topic (first one for demo)
	topic := topics[0]
	fmt.Printf("Research Topic: %s\n", topic)
	fmt.Println(strings.Repeat("-", 60))

	// Create initial state
	state := domain.NewState()
	state.Set("topic", topic)

	// Run research
	ctx := context.Background()
	result, err := assistant.Run(ctx, state)
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
	if count, ok := result.Get("sources_count"); ok {
		fmt.Printf("- Sources analyzed: %v\n", count)
	}

	fmt.Println("\nNote: This example works best with API keys set for web search and LLM providers.")
	fmt.Println("Without them, it uses mock agents for demonstration.")
}
