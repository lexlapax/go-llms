# Building Research Agents: Information Gathering Systems

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Guides](/docs/user-guide/guides/) / Building Research Agents**

Master the creation of intelligent research agents that can search the web, gather information from multiple sources, analyze content, and synthesize comprehensive reports. Build production systems that handle unreliable data sources with grace.

## Why Research Agents Matter

- **Comprehensive Coverage** - Search multiple engines and sources simultaneously
- **Intelligent Processing** - LLM-powered deduplication, analysis, and synthesis
- **Reliable Operations** - Handle unreliable external sources with fallbacks
- **Scalable Architecture** - From simple searches to complex research workflows
- **Production Ready** - Authentication, rate limiting, and error recovery

## Research Agent Architecture

![Research Agent Flow](../../images/research-agent-flow.svg)

### Core Components
1. **Multi-Engine Search** - Query multiple search engines in parallel
2. **Content Retrieval** - Fetch detailed information from sources
3. **Data Processing** - Extract, clean, and normalize gathered data
4. **Intelligence Layer** - LLM-powered analysis and synthesis
5. **State Management** - Track research progress and findings

### Tool Categories
| Category | Tools | Purpose |
|----------|-------|---------|
| **Web Search** | WebSearch, multi-engine support | Find relevant sources |
| **Content Retrieval** | WebFetch, WebScrape, HTTPRequest | Get detailed information |
| **Data Processing** | JSONProcess, DataTransform | Clean and structure data |
| **Feed Processing** | FeedFetch, FeedFilter | Monitor ongoing sources |
| **File Operations** | ReadFile, WriteFile, FileSearch | Store and retrieve findings |

## Prerequisites

- [Creating Agents guide completed](creating-agents.md) ✅
- [Agent Tools guide helpful](agent-tools.md) ✅
- Understanding of async operations ✅

---

## Level 1: Simple Search Agent
*Build your first research agent in 10 minutes*

### Basic Web Search Agent
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)

func main() {
    fmt.Println("🔍 Simple Research Agent")
    fmt.Println("========================")

    // Create research agent
    agent, err := core.NewAgentFromString("research-agent", "anthropic/claude-3-5-sonnet")
    if err != nil {
        log.Fatalf("Failed to create agent: %v", err)
    }

    agent.SetSystemPrompt(`You are a research assistant with web search capabilities.
    When asked to research a topic:
    1. Use web search to find current information
    2. Analyze the search results for relevance and credibility
    3. Provide a comprehensive summary with key findings
    4. Always cite your sources with URLs
    
    Focus on recent, authoritative sources.`)

    // Add web search tool (supports multiple engines)
    searchTool := web.NewWebSearchTool()
    agent.AddTool(searchTool)

    // Research topics
    topics := []string{
        "Latest developments in Go programming language 2025",
        "Current state of artificial intelligence safety research",
        "Recent breakthroughs in quantum computing",
    }

    for i, topic := range topics {
        fmt.Printf("\n--- Research Task %d ---\n", i+1)
        fmt.Printf("Topic: %s\n", topic)

        // Create research state
        state := domain.NewState()
        state.Set("user_input", fmt.Sprintf("Research this topic comprehensively: %s", topic))
        
        // Optional: Configure search parameters
        state.Set("search_engines", []string{"tavily", "brave", "serpapi"})
        state.Set("max_results", 10)

        // Execute research
        result, err := agent.Run(context.Background(), state)
        if err != nil {
            log.Printf("Research failed: %v", err)
            continue
        }

        if findings, exists := result.Get("response"); exists {
            fmt.Printf("\n📋 Research Findings:\n%v\n", findings)
        }
        
        // Check for search results
        if searchData, exists := result.Get("search_results"); exists {
            fmt.Printf("\n🔗 Sources Found: %d results\n", len(searchData.([]interface{})))
        }
    }
}
```

### Key Features
✅ **Multi-Engine Search** - Automatically tries multiple search engines  
✅ **Intelligent Analysis** - LLM processes and synthesizes results  
✅ **Source Citation** - Maintains links to original sources  
✅ **Configurable Parameters** - Control search depth and engines  

---

## Level 2: Multi-Source Research Agent
*Combine search, fetching, and analysis*

### Comprehensive Research System
```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "sync"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
)

// ResearchFindings represents structured research output
type ResearchFindings struct {
    Topic     string    `json:"topic"`
    Sources   []Source  `json:"sources"`
    KeyPoints []string  `json:"key_points"`
    Summary   string    `json:"summary"`
    Timestamp string    `json:"timestamp"`
}

type Source struct {
    Title       string `json:"title"`
    URL         string `json:"url"`
    Description string `json:"description"`
    Relevance   int    `json:"relevance"`
}

// MultiSourceResearcher combines multiple tools for comprehensive research
type MultiSourceResearcher struct {
    searchAgent    domain.BaseAgent
    analysisAgent  domain.BaseAgent
    synthesizerAgent domain.BaseAgent
    
    searchTool     domain.Tool
    fetchTool      domain.Tool
    scrapeTool     domain.Tool
    jsonTool       domain.Tool
    fileTool       domain.Tool
}

// NewMultiSourceResearcher creates a comprehensive research system
func NewMultiSourceResearcher() (*MultiSourceResearcher, error) {
    // Create specialized agents
    searchAgent, err := core.NewAgentFromString("search-specialist", "gemini/gemini-2.0-flash")
    if err != nil {
        return nil, fmt.Errorf("failed to create search agent: %w", err)
    }
    searchAgent.SetSystemPrompt(`You are a search specialist. Generate effective search queries and evaluate search results for relevance and credibility.`)

    analysisAgent, err := core.NewAgentFromString("content-analyst", "anthropic/claude-3-5-sonnet")
    if err != nil {
        return nil, fmt.Errorf("failed to create analysis agent: %w", err)
    }
    analysisAgent.SetSystemPrompt(`You are a content analyst. Extract key information, facts, and insights from web content. Focus on accuracy and relevance.`)

    synthesizerAgent, err := core.NewAgentFromString("report-synthesizer", "openai/gpt-4o")
    if err != nil {
        return nil, fmt.Errorf("failed to create synthesizer agent: %w", err)
    }
    synthesizerAgent.SetSystemPrompt(`You are a report synthesizer. Create comprehensive, well-structured reports from multiple sources. Ensure balanced perspectives and cite all sources.`)

    // Create tools
    return &MultiSourceResearcher{
        searchAgent:      searchAgent,
        analysisAgent:    analysisAgent,
        synthesizerAgent: synthesizerAgent,
        searchTool:       web.NewWebSearchTool(),
        fetchTool:        web.NewWebFetchTool(),
        scrapeTool:       web.NewWebScrapeTool(),
        jsonTool:         data.NewJSONProcessTool(),
        fileTool:         file.NewWriteFileTool(),
    }, nil
}

// Research conducts comprehensive multi-phase research
func (r *MultiSourceResearcher) Research(ctx context.Context, topic string) (*ResearchFindings, error) {
    fmt.Printf("🔬 Starting comprehensive research on: %s\n", topic)
    
    // Phase 1: Multi-engine search
    searchResults, err := r.searchPhase(ctx, topic)
    if err != nil {
        return nil, fmt.Errorf("search phase failed: %w", err)
    }
    fmt.Printf("📊 Found %d initial sources\n", len(searchResults))

    // Phase 2: Content retrieval and analysis
    analyzedContent, err := r.analysisPhase(ctx, searchResults)
    if err != nil {
        return nil, fmt.Errorf("analysis phase failed: %w", err)
    }
    fmt.Printf("🧠 Analyzed %d content pieces\n", len(analyzedContent))

    // Phase 3: Synthesis and report generation
    findings, err := r.synthesisPhase(ctx, topic, analyzedContent)
    if err != nil {
        return nil, fmt.Errorf("synthesis phase failed: %w", err)
    }

    // Phase 4: Save findings
    err = r.saveFindings(ctx, findings)
    if err != nil {
        log.Printf("Warning: failed to save findings: %v", err)
    }

    return findings, nil
}

// searchPhase performs parallel searches across multiple engines
func (r *MultiSourceResearcher) searchPhase(ctx context.Context, topic string) ([]map[string]interface{}, error) {
    // Generate search variations
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf("Generate 3-5 different search queries for comprehensive research on: %s", topic))
    
    queryResult, err := r.searchAgent.Run(ctx, state)
    if err != nil {
        return nil, fmt.Errorf("query generation failed: %w", err)
    }

    // Extract search queries (simplified - in real implementation, parse structured output)
    queries := []string{
        topic + " overview",
        topic + " latest developments",
        topic + " expert analysis",
        topic + " 2025",
    }

    // Execute parallel searches
    var mu sync.Mutex
    var wg sync.WaitGroup
    allResults := make([]map[string]interface{}, 0)

    for _, query := range queries {
        wg.Add(1)
        go func(q string) {
            defer wg.Done()
            
            // Execute search
            searchParams := map[string]interface{}{
                "query": q,
                "max_results": 5,
                "engines": []string{"tavily", "brave", "serpapi"},
            }
            
            result, err := r.searchTool.Execute(ctx, searchParams)
            if err != nil {
                log.Printf("Search failed for query '%s': %v", q, err)
                return
            }

            mu.Lock()
            if results, ok := result.Output.([]interface{}); ok {
                for _, r := range results {
                    if resultMap, ok := r.(map[string]interface{}); ok {
                        allResults = append(allResults, resultMap)
                    }
                }
            }
            mu.Unlock()
        }(query)
    }

    wg.Wait()

    // Remove duplicates (simplified)
    return r.deduplicateResults(allResults), nil
}

// analysisPhase fetches content and analyzes it
func (r *MultiSourceResearcher) analysisPhase(ctx context.Context, searchResults []map[string]interface{}) ([]map[string]interface{}, error) {
    analyzed := make([]map[string]interface{}, 0)
    
    // Process top results (limit to prevent overload)
    maxResults := 10
    if len(searchResults) < maxResults {
        maxResults = len(searchResults)
    }

    for i := 0; i < maxResults; i++ {
        result := searchResults[i]
        url, ok := result["url"].(string)
        if !ok {
            continue
        }

        // Fetch full content
        fetchParams := map[string]interface{}{
            "url": url,
            "timeout": 10,
            "max_content_length": 50000,
        }

        content, err := r.fetchTool.Execute(ctx, fetchParams)
        if err != nil {
            log.Printf("Failed to fetch %s: %v", url, err)
            continue
        }

        // Analyze content
        state := domain.NewState()
        state.Set("user_input", fmt.Sprintf("Analyze this content for key information and insights. Focus on facts, data, and expert opinions: %v", content.Output))
        
        analysis, err := r.analysisAgent.Run(ctx, state)
        if err != nil {
            log.Printf("Content analysis failed for %s: %v", url, err)
            continue
        }

        // Structure analyzed data
        analyzedResult := map[string]interface{}{
            "url": url,
            "title": result["title"],
            "description": result["description"],
            "content": content.Output,
            "analysis": analysis.Get("response"),
            "timestamp": "2025-01-23", // In real implementation, use time.Now()
        }

        analyzed = append(analyzed, analyzedResult)
    }

    return analyzed, nil
}

// synthesisPhase combines all analyzed content into comprehensive findings
func (r *MultiSourceResearcher) synthesisPhase(ctx context.Context, topic string, analyzedContent []map[string]interface{}) (*ResearchFindings, error) {
    // Prepare synthesis input
    state := domain.NewState()
    synthesizerPrompt := fmt.Sprintf(`Based on the following analyzed content, create a comprehensive research report on "%s".

Structure your response as JSON with the following fields:
- topic: the research topic
- key_points: array of 5-10 key findings
- summary: comprehensive 2-3 paragraph summary
- sources: array of source objects with title, url, description, relevance (1-10)

Analyzed Content:
%v

Ensure all claims are supported by the sources and provide balanced perspectives.`, topic, r.formatContentForSynthesis(analyzedContent))

    state.Set("user_input", synthesizerPrompt)

    result, err := r.synthesizerAgent.Run(ctx, state)
    if err != nil {
        return nil, fmt.Errorf("synthesis failed: %w", err)
    }

    // Parse structured output (simplified - in real implementation, use structured output parser)
    findings := &ResearchFindings{
        Topic:     topic,
        Sources:   r.extractSources(analyzedContent),
        KeyPoints: []string{"Key finding 1", "Key finding 2", "Key finding 3"}, // Extract from LLM response
        Summary:   fmt.Sprintf("%v", result.Get("response")),
        Timestamp: "2025-01-23T10:00:00Z",
    }

    return findings, nil
}

// saveFindings persists research findings
func (r *MultiSourceResearcher) saveFindings(ctx context.Context, findings *ResearchFindings) error {
    // Convert to JSON
    jsonParams := map[string]interface{}{
        "data": findings,
        "operation": "stringify",
        "pretty": true,
    }

    jsonResult, err := r.jsonTool.Execute(ctx, jsonParams)
    if err != nil {
        return fmt.Errorf("JSON formatting failed: %w", err)
    }

    // Save to file
    filename := fmt.Sprintf("research_%s_%s.json", 
        r.sanitizeFilename(findings.Topic), 
        findings.Timestamp[:10])

    fileParams := map[string]interface{}{
        "path": filename,
        "content": jsonResult.Output,
        "create_dirs": true,
    }

    _, err = r.fileTool.Execute(ctx, fileParams)
    if err != nil {
        return fmt.Errorf("file save failed: %w", err)
    }

    fmt.Printf("💾 Research findings saved to: %s\n", filename)
    return nil
}

// Helper methods
func (r *MultiSourceResearcher) deduplicateResults(results []map[string]interface{}) []map[string]interface{} {
    seen := make(map[string]bool)
    unique := make([]map[string]interface{}, 0)
    
    for _, result := range results {
        if url, ok := result["url"].(string); ok {
            if !seen[url] {
                seen[url] = true
                unique = append(unique, result)
            }
        }
    }
    
    return unique
}

func (r *MultiSourceResearcher) formatContentForSynthesis(content []map[string]interface{}) string {
    // Simplified formatting - in real implementation, create structured summary
    return "Analyzed content from multiple sources..."
}

func (r *MultiSourceResearcher) extractSources(content []map[string]interface{}) []Source {
    sources := make([]Source, 0)
    for _, item := range content {
        source := Source{
            Title:       fmt.Sprintf("%v", item["title"]),
            URL:         fmt.Sprintf("%v", item["url"]),
            Description: fmt.Sprintf("%v", item["description"]),
            Relevance:   8, // In real implementation, calculate relevance
        }
        sources = append(sources, source)
    }
    return sources
}

func (r *MultiSourceResearcher) sanitizeFilename(name string) string {
    // Simple sanitization - in real implementation, handle all special characters
    return "research_topic"
}

func main() {
    fmt.Println("🎯 Multi-Source Research Agent")
    fmt.Println("==============================")

    // Create comprehensive researcher
    researcher, err := NewMultiSourceResearcher()
    if err != nil {
        log.Fatalf("Failed to create researcher: %v", err)
    }

    // Research topics
    topics := []string{
        "Impact of AI on software development",
        "Sustainable energy solutions 2025",
        "Remote work productivity trends",
    }

    for _, topic := range topics {
        fmt.Printf("\n🔍 Researching: %s\n", topic)
        fmt.Println(strings.Repeat("=", 50))

        findings, err := researcher.Research(context.Background(), topic)
        if err != nil {
            log.Printf("Research failed: %v", err)
            continue
        }

        // Display results
        fmt.Printf("\n📋 Research Complete!\n")
        fmt.Printf("Topic: %s\n", findings.Topic)
        fmt.Printf("Sources: %d\n", len(findings.Sources))
        fmt.Printf("Key Points: %d\n", len(findings.KeyPoints))
        fmt.Printf("Summary Length: %d characters\n", len(findings.Summary))
        
        fmt.Printf("\n🔗 Top Sources:\n")
        for i, source := range findings.Sources {
            if i >= 3 { break } // Show top 3
            fmt.Printf("  %d. %s (Relevance: %d/10)\n", i+1, source.Title, source.Relevance)
            fmt.Printf("     %s\n", source.URL)
        }
    }
}
```

### Advanced Features
✅ **Multi-Phase Processing** - Search → Fetch → Analyze → Synthesize  
✅ **Parallel Operations** - Concurrent searches and content fetching  
✅ **Specialized Agents** - Different LLMs for different tasks  
✅ **Structured Output** - JSON-formatted research findings  
✅ **Data Persistence** - Save research for future reference  

---

## Level 3: Production Research System
*Handle unreliable sources and large-scale research*

### Enterprise Research Platform
```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed"
)

// ProductionResearchSystem handles enterprise-scale research operations
type ProductionResearchSystem struct {
    config           *ResearchConfig
    searchEngines    []string
    apiKeys          map[string]string
    rateLimiters     map[string]*RateLimiter
    cacheSystem      *CacheSystem
    retryManager     *RetryManager
    alertSystem      *AlertSystem
    
    // Agent pools
    searchAgents     []domain.BaseAgent
    analysisAgents   []domain.BaseAgent
    synthesizerAgents []domain.BaseAgent
}

// ResearchConfig holds production configuration
type ResearchConfig struct {
    MaxConcurrentSearches    int
    MaxResultsPerEngine      int
    ContentFetchTimeout      time.Duration
    AnalysisTimeout          time.Duration
    RetryAttempts           int
    CacheEnabled            bool
    CacheTTL               time.Duration
    AlertsEnabled          bool
    QualityThreshold       float64
    EnableContinuousMonitoring bool
}

// RateLimiter manages API rate limiting
type RateLimiter struct {
    requestsPerMinute int
    window           time.Duration
    requests         []time.Time
    mutex           sync.Mutex
}

// CacheSystem manages research result caching
type CacheSystem struct {
    cache     map[string]CachedResult
    mutex     sync.RWMutex
    ttl       time.Duration
}

type CachedResult struct {
    Data      interface{}
    Timestamp time.Time
    TTL       time.Duration
}

// RetryManager handles intelligent retry logic
type RetryManager struct {
    maxRetries      int
    baseDelay       time.Duration
    maxDelay        time.Duration
    exponentialBase float64
}

// AlertSystem sends notifications for research issues
type AlertSystem struct {
    webhookURL    string
    emailEnabled  bool
    slackEnabled  bool
}

// NewProductionResearchSystem creates enterprise research system
func NewProductionResearchSystem(config *ResearchConfig) (*ProductionResearchSystem, error) {
    // Load API keys from environment
    apiKeys := map[string]string{
        "tavily":   os.Getenv("TAVILY_API_KEY"),
        "serpapi":  os.Getenv("SERPAPI_API_KEY"),
        "brave":    os.Getenv("BRAVE_SEARCH_API_KEY"),
        "serper":   os.Getenv("SERPER_API_KEY"),
    }

    // Initialize components
    system := &ProductionResearchSystem{
        config:        config,
        searchEngines: []string{"tavily", "brave", "serpapi", "serper"},
        apiKeys:       apiKeys,
        rateLimiters:  make(map[string]*RateLimiter),
        cacheSystem:   NewCacheSystem(config.CacheTTL),
        retryManager:  NewRetryManager(config.RetryAttempts),
        alertSystem:   NewAlertSystem(),
    }

    // Initialize rate limiters for each engine
    for engine := range apiKeys {
        system.rateLimiters[engine] = NewRateLimiter(60) // 60 requests per minute
    }

    // Create agent pools
    err := system.initializeAgentPools()
    if err != nil {
        return nil, fmt.Errorf("failed to initialize agent pools: %w", err)
    }

    return system, nil
}

// ExecuteResearchPipeline runs comprehensive research with production features
func (s *ProductionResearchSystem) ExecuteResearchPipeline(ctx context.Context, request *ResearchRequest) (*ResearchReport, error) {
    startTime := time.Now()
    
    fmt.Printf("🏭 Starting production research pipeline\n")
    fmt.Printf("Topic: %s\n", request.Topic)
    fmt.Printf("Depth: %s\n", request.Depth)
    fmt.Printf("Max Sources: %d\n", request.MaxSources)

    // Phase 1: Intelligent search planning
    searchPlan, err := s.planSearchStrategy(ctx, request)
    if err != nil {
        s.alertSystem.SendAlert("Search planning failed", err)
        return nil, fmt.Errorf("search planning failed: %w", err)
    }

    // Phase 2: Parallel multi-engine search with retry
    searchResults, err := s.executeSearchPlan(ctx, searchPlan)
    if err != nil {
        s.alertSystem.SendAlert("Search execution failed", err)
        return nil, fmt.Errorf("search execution failed: %w", err)
    }
    fmt.Printf("📊 Collected %d search results\n", len(searchResults))

    // Phase 3: Content quality assessment and filtering
    qualityResults, err := s.assessContentQuality(ctx, searchResults)
    if err != nil {
        log.Printf("Warning: Quality assessment failed: %v", err)
        qualityResults = searchResults // Fallback to unfiltered results
    }
    fmt.Printf("✅ %d high-quality sources identified\n", len(qualityResults))

    // Phase 4: Concurrent content analysis
    analyzedContent, err := s.analyzeContentConcurrently(ctx, qualityResults)
    if err != nil {
        return nil, fmt.Errorf("content analysis failed: %w", err)
    }
    fmt.Printf("🧠 Analyzed %d content pieces\n", len(analyzedContent))

    // Phase 5: Multi-perspective synthesis
    report, err := s.synthesizeReport(ctx, request, analyzedContent)
    if err != nil {
        return nil, fmt.Errorf("report synthesis failed: %w", err)
    }

    // Phase 6: Quality validation and enhancement
    err = s.validateAndEnhanceReport(ctx, report)
    if err != nil {
        log.Printf("Warning: Report validation failed: %v", err)
    }

    // Add metadata
    report.ExecutionTime = time.Since(startTime)
    report.SourcesAnalyzed = len(analyzedContent)
    report.QualityScore = s.calculateQualityScore(report)
    
    fmt.Printf("✅ Research pipeline completed in %v\n", report.ExecutionTime)
    fmt.Printf("📈 Quality Score: %.2f/10\n", report.QualityScore)

    return report, nil
}

// planSearchStrategy creates intelligent search plan
func (s *ProductionResearchSystem) planSearchStrategy(ctx context.Context, request *ResearchRequest) (*SearchPlan, error) {
    // Use dedicated planning agent
    plannerAgent := s.getAvailableAgent(s.searchAgents)
    
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf(`Create a comprehensive search strategy for: "%s"
    
    Requirements:
    - Depth: %s
    - Max sources: %d
    - Generate 5-8 diverse search queries
    - Consider different perspectives and angles
    - Include recent and historical context
    - Suggest optimal search engines for each query
    
    Return structured search plan.`, request.Topic, request.Depth, request.MaxSources))

    result, err := plannerAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    // Parse search plan (simplified - in real implementation, use structured output)
    plan := &SearchPlan{
        Topic:     request.Topic,
        Queries:   s.generateSearchQueries(request),
        Engines:   s.selectOptimalEngines(request),
        Strategy:  "comprehensive",
        Timestamp: time.Now(),
    }

    return plan, nil
}

// executeSearchPlan runs parallel searches with retry and fallback
func (s *ProductionResearchSystem) executeSearchPlan(ctx context.Context, plan *SearchPlan) ([]SearchResult, error) {
    var mu sync.Mutex
    var wg sync.WaitGroup
    allResults := make([]SearchResult, 0)
    
    // Create work channel for rate limiting
    workChan := make(chan SearchJob, len(plan.Queries)*len(plan.Engines))
    resultChan := make(chan []SearchResult, len(plan.Queries)*len(plan.Engines))

    // Start worker pool
    numWorkers := s.config.MaxConcurrentSearches
    for i := 0; i < numWorkers; i++ {
        go s.searchWorker(ctx, workChan, resultChan)
    }

    // Queue search jobs
    go func() {
        defer close(workChan)
        for _, query := range plan.Queries {
            for _, engine := range plan.Engines {
                workChan <- SearchJob{
                    Query:    query,
                    Engine:   engine,
                    MaxResults: s.config.MaxResultsPerEngine,
                }
            }
        }
    }()

    // Collect results
    go func() {
        defer close(resultChan)
        for i := 0; i < len(plan.Queries)*len(plan.Engines); i++ {
            select {
            case results := <-resultChan:
                mu.Lock()
                allResults = append(allResults, results...)
                mu.Unlock()
            case <-ctx.Done():
                return
            }
        }
    }()

    // Wait for completion
    select {
    case <-time.After(5 * time.Minute): // Search timeout
        return nil, fmt.Errorf("search operation timed out")
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Continue
    }

    // Remove duplicates and rank results
    dedupedResults := s.deduplicateAndRankResults(allResults)
    
    return dedupedResults, nil
}

// searchWorker processes search jobs with rate limiting and retry
func (s *ProductionResearchSystem) searchWorker(ctx context.Context, jobs <-chan SearchJob, results chan<- []SearchResult) {
    searchTool := web.NewWebSearchTool()
    
    for job := range jobs {
        // Check rate limiter
        rateLimiter := s.rateLimiters[job.Engine]
        if !rateLimiter.Allow() {
            log.Printf("Rate limit exceeded for %s, skipping", job.Engine)
            results <- []SearchResult{}
            continue
        }

        // Check cache first
        cacheKey := fmt.Sprintf("%s:%s", job.Query, job.Engine)
        if s.config.CacheEnabled {
            if cached, exists := s.cacheSystem.Get(cacheKey); exists {
                results <- cached.([]SearchResult)
                continue
            }
        }

        // Execute search with retry
        searchResults, err := s.retryManager.ExecuteWithRetry(func() (interface{}, error) {
            params := map[string]interface{}{
                "query": job.Query,
                "engine": job.Engine,
                "max_results": job.MaxResults,
                "timeout": s.config.ContentFetchTimeout.Seconds(),
            }
            
            return searchTool.Execute(ctx, params)
        })

        if err != nil {
            log.Printf("Search failed for query '%s' on engine '%s': %v", job.Query, job.Engine, err)
            results <- []SearchResult{}
            continue
        }

        // Process and cache results
        processedResults := s.processSearchResults(searchResults, job.Engine)
        
        if s.config.CacheEnabled {
            s.cacheSystem.Set(cacheKey, processedResults, s.config.CacheTTL)
        }
        
        results <- processedResults
    }
}

// assessContentQuality filters results based on quality metrics
func (s *ProductionResearchSystem) assessContentQuality(ctx context.Context, results []SearchResult) ([]SearchResult, error) {
    qualityAgent := s.getAvailableAgent(s.analysisAgents)
    
    var highQualityResults []SearchResult
    
    for _, result := range results {
        // Quick quality check
        qualityScore := s.calculateContentQuality(result)
        
        if qualityScore >= s.config.QualityThreshold {
            highQualityResults = append(highQualityResults, result)
        }
    }

    // If too few high-quality results, lower threshold
    if len(highQualityResults) < 5 {
        log.Printf("Insufficient high-quality results, lowering threshold")
        highQualityResults = results[:min(10, len(results))]
    }

    return highQualityResults, nil
}

// analyzeContentConcurrently processes content in parallel
func (s *ProductionResearchSystem) analyzeContentConcurrently(ctx context.Context, results []SearchResult) ([]AnalyzedContent, error) {
    semaphore := make(chan struct{}, s.config.MaxConcurrentSearches)
    var mu sync.Mutex
    var wg sync.WaitGroup
    analyzed := make([]AnalyzedContent, 0)

    for _, result := range results {
        wg.Add(1)
        go func(r SearchResult) {
            defer wg.Done()
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release

            content, err := s.fetchAndAnalyzeContent(ctx, r)
            if err != nil {
                log.Printf("Content analysis failed for %s: %v", r.URL, err)
                return
            }

            mu.Lock()
            analyzed = append(analyzed, content)
            mu.Unlock()
        }(result)
    }

    wg.Wait()
    return analyzed, nil
}

// Supporting types and methods (simplified implementations)
type ResearchRequest struct {
    Topic       string
    Depth       string
    MaxSources  int
    Urgency     string
    Perspective string
}

type SearchPlan struct {
    Topic     string
    Queries   []string
    Engines   []string
    Strategy  string
    Timestamp time.Time
}

type SearchJob struct {
    Query      string
    Engine     string
    MaxResults int
}

type SearchResult struct {
    Title       string
    URL         string
    Description string
    Engine      string
    Timestamp   time.Time
    Quality     float64
}

type AnalyzedContent struct {
    SearchResult
    Content   string
    KeyPoints []string
    Insights  []string
    Entities  []string
    Sentiment float64
}

type ResearchReport struct {
    Topic           string
    Summary         string
    KeyFindings     []string
    Sources         []SearchResult
    Recommendations []string
    QualityScore    float64
    ExecutionTime   time.Duration
    SourcesAnalyzed int
    Metadata        map[string]interface{}
}

// Helper function implementations
func NewCacheSystem(ttl time.Duration) *CacheSystem {
    return &CacheSystem{
        cache: make(map[string]CachedResult),
        ttl:   ttl,
    }
}

func (c *CacheSystem) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    if result, exists := c.cache[key]; exists {
        if time.Since(result.Timestamp) < result.TTL {
            return result.Data, true
        }
        delete(c.cache, key)
    }
    return nil, false
}

func (c *CacheSystem) Set(key string, data interface{}, ttl time.Duration) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    
    c.cache[key] = CachedResult{
        Data:      data,
        Timestamp: time.Now(),
        TTL:       ttl,
    }
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
    return &RateLimiter{
        requestsPerMinute: requestsPerMinute,
        window:           time.Minute,
        requests:         make([]time.Time, 0),
    }
}

func (r *RateLimiter) Allow() bool {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    now := time.Now()
    
    // Clean old requests
    validRequests := make([]time.Time, 0)
    for _, reqTime := range r.requests {
        if now.Sub(reqTime) < r.window {
            validRequests = append(validRequests, reqTime)
        }
    }
    r.requests = validRequests
    
    // Check if we can make a new request
    if len(r.requests) < r.requestsPerMinute {
        r.requests = append(r.requests, now)
        return true
    }
    
    return false
}

func NewRetryManager(maxRetries int) *RetryManager {
    return &RetryManager{
        maxRetries:      maxRetries,
        baseDelay:       time.Second,
        maxDelay:        30 * time.Second,
        exponentialBase: 2.0,
    }
}

func (r *RetryManager) ExecuteWithRetry(fn func() (interface{}, error)) (interface{}, error) {
    var lastErr error
    
    for attempt := 0; attempt < r.maxRetries; attempt++ {
        if attempt > 0 {
            delay := time.Duration(float64(r.baseDelay) * math.Pow(r.exponentialBase, float64(attempt-1)))
            if delay > r.maxDelay {
                delay = r.maxDelay
            }
            time.Sleep(delay)
        }
        
        result, err := fn()
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        log.Printf("Attempt %d failed: %v", attempt+1, err)
    }
    
    return nil, fmt.Errorf("all %d attempts failed, last error: %w", r.maxRetries, lastErr)
}

func NewAlertSystem() *AlertSystem {
    return &AlertSystem{
        webhookURL:   os.Getenv("RESEARCH_WEBHOOK_URL"),
        emailEnabled: os.Getenv("RESEARCH_EMAIL_ALERTS") == "true",
        slackEnabled: os.Getenv("RESEARCH_SLACK_ALERTS") == "true",
    }
}

func (a *AlertSystem) SendAlert(message string, err error) {
    alertMsg := fmt.Sprintf("Research Alert: %s - Error: %v", message, err)
    log.Printf("ALERT: %s", alertMsg)
    
    // In real implementation, send to webhook, email, Slack, etc.
}

// Simplified implementations for example
func (s *ProductionResearchSystem) initializeAgentPools() error {
    // Create agent pools (simplified)
    for i := 0; i < 3; i++ {
        searchAgent, _ := core.NewAgentFromString(fmt.Sprintf("search-%d", i), "gemini/gemini-2.0-flash")
        s.searchAgents = append(s.searchAgents, searchAgent)
        
        analysisAgent, _ := core.NewAgentFromString(fmt.Sprintf("analysis-%d", i), "anthropic/claude-3-5-sonnet")
        s.analysisAgents = append(s.analysisAgents, analysisAgent)
        
        synthesizerAgent, _ := core.NewAgentFromString(fmt.Sprintf("synthesizer-%d", i), "openai/gpt-4o")
        s.synthesizerAgents = append(s.synthesizerAgents, synthesizerAgent)
    }
    return nil
}

func (s *ProductionResearchSystem) getAvailableAgent(pool []domain.BaseAgent) domain.BaseAgent {
    // Simple round-robin selection
    return pool[0]
}

func (s *ProductionResearchSystem) generateSearchQueries(request *ResearchRequest) []string {
    return []string{
        request.Topic,
        request.Topic + " 2025",
        request.Topic + " trends",
        request.Topic + " analysis",
        request.Topic + " expert opinion",
    }
}

func (s *ProductionResearchSystem) selectOptimalEngines(request *ResearchRequest) []string {
    engines := make([]string, 0)
    for engine, apiKey := range s.apiKeys {
        if apiKey != "" {
            engines = append(engines, engine)
        }
    }
    return engines
}

func (s *ProductionResearchSystem) processSearchResults(raw interface{}, engine string) []SearchResult {
    // Simplified processing
    return []SearchResult{}
}

func (s *ProductionResearchSystem) deduplicateAndRankResults(results []SearchResult) []SearchResult {
    // Simplified deduplication
    return results
}

func (s *ProductionResearchSystem) calculateContentQuality(result SearchResult) float64 {
    // Simplified quality calculation
    return 8.0
}

func (s *ProductionResearchSystem) fetchAndAnalyzeContent(ctx context.Context, result SearchResult) (AnalyzedContent, error) {
    // Simplified implementation
    return AnalyzedContent{SearchResult: result}, nil
}

func (s *ProductionResearchSystem) synthesizeReport(ctx context.Context, request *ResearchRequest, content []AnalyzedContent) (*ResearchReport, error) {
    return &ResearchReport{
        Topic:       request.Topic,
        Summary:     "Research summary...",
        KeyFindings: []string{"Finding 1", "Finding 2"},
    }, nil
}

func (s *ProductionResearchSystem) validateAndEnhanceReport(ctx context.Context, report *ResearchReport) error {
    return nil
}

func (s *ProductionResearchSystem) calculateQualityScore(report *ResearchReport) float64 {
    return 8.5
}

func min(a, b int) int {
    if a < b { return a }
    return b
}

func main() {
    fmt.Println("🏭 Production Research System")
    fmt.Println("============================")

    // Create production configuration
    config := &ResearchConfig{
        MaxConcurrentSearches:    5,
        MaxResultsPerEngine:      10,
        ContentFetchTimeout:      30 * time.Second,
        AnalysisTimeout:         60 * time.Second,
        RetryAttempts:           3,
        CacheEnabled:            true,
        CacheTTL:               24 * time.Hour,
        AlertsEnabled:          true,
        QualityThreshold:       7.0,
        EnableContinuousMonitoring: true,
    }

    // Initialize production system
    system, err := NewProductionResearchSystem(config)
    if err != nil {
        log.Fatalf("Failed to initialize production system: %v", err)
    }

    // Execute production research
    request := &ResearchRequest{
        Topic:       "Enterprise AI adoption trends 2025",
        Depth:       "comprehensive",
        MaxSources:  20,
        Urgency:     "high",
        Perspective: "business",
    }

    report, err := system.ExecuteResearchPipeline(context.Background(), request)
    if err != nil {
        log.Fatalf("Research pipeline failed: %v", err)
    }

    // Display results
    fmt.Printf("\n🎯 Production Research Complete\n")
    fmt.Printf("Topic: %s\n", report.Topic)
    fmt.Printf("Quality Score: %.2f/10\n", report.QualityScore)
    fmt.Printf("Execution Time: %v\n", report.ExecutionTime)
    fmt.Printf("Sources Analyzed: %d\n", report.SourcesAnalyzed)
    fmt.Printf("Key Findings: %d\n", len(report.KeyFindings))
}
```

### Production Features
✅ **Rate Limiting** - Respect API limits across all services  
✅ **Caching System** - Reduce API calls and improve performance  
✅ **Retry Logic** - Handle failures with exponential backoff  
✅ **Quality Assessment** - Filter low-quality sources automatically  
✅ **Concurrent Processing** - Parallel content analysis with semaphores  
✅ **Alert System** - Monitor failures and performance issues  
✅ **Agent Pooling** - Distribute load across multiple agents  
✅ **Structured Monitoring** - Track execution time and quality metrics  

---

## Level 4: Continuous Research Monitoring
*Build systems that monitor topics over time*

### Monitoring and Feed Integration
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
)

// ContinuousResearchMonitor tracks topics over time
type ContinuousResearchMonitor struct {
    topics          []MonitoredTopic
    feedProcessor   domain.BaseAgent
    alertAgent      domain.BaseAgent
    updateAgent     domain.BaseAgent
    
    feedTools       []domain.Tool
    webTools        []domain.Tool
    
    updateInterval  time.Duration
    alertThreshold  float64
}

type MonitoredTopic struct {
    Name            string
    Keywords        []string
    FeedSources     []string
    SearchQueries   []string
    LastUpdate      time.Time
    ChangeScore     float64
    TrendDirection  string
    Alerts          []Alert
}

type Alert struct {
    Type        string
    Message     string
    Severity    string
    Timestamp   time.Time
    Topic       string
    Source      string
}

func NewContinuousResearchMonitor() (*ContinuousResearchMonitor, error) {
    // Create specialized agents for monitoring
    feedProcessor, err := core.NewAgentFromString("feed-processor", "gemini/gemini-2.0-flash")
    if err != nil {
        return nil, err
    }
    feedProcessor.SetSystemPrompt(`You process RSS/Atom feeds and news sources to identify relevant updates on monitored topics. Extract key information, assess importance, and detect trends.`)

    alertAgent, err := core.NewAgentFromString("alert-generator", "anthropic/claude-3-5-sonnet")
    if err != nil {
        return nil, err
    }
    alertAgent.SetSystemPrompt(`You generate intelligent alerts for significant developments in monitored topics. Assess severity, context, and actionability.`)

    updateAgent, err := core.NewAgentFromString("trend-analyzer", "openai/gpt-4o")
    if err != nil {
        return nil, err
    }
    updateAgent.SetSystemPrompt(`You analyze trends and changes in monitored topics over time. Identify patterns, predict developments, and provide strategic insights.`)

    return &ContinuousResearchMonitor{
        feedProcessor:   feedProcessor,
        alertAgent:     alertAgent,
        updateAgent:    updateAgent,
        feedTools:      []domain.Tool{feed.NewFeedFetchTool(), feed.NewFeedFilterTool()},
        webTools:       []domain.Tool{web.NewWebSearchTool(), web.NewWebFetchTool()},
        updateInterval: 30 * time.Minute,
        alertThreshold: 7.0,
    }, nil
}

// StartMonitoring begins continuous monitoring of topics
func (m *ContinuousResearchMonitor) StartMonitoring(ctx context.Context, topics []MonitoredTopic) {
    m.topics = topics
    
    fmt.Printf("📡 Starting continuous monitoring of %d topics\n", len(topics))
    
    // Initial baseline collection
    for i := range m.topics {
        err := m.establishBaseline(ctx, &m.topics[i])
        if err != nil {
            log.Printf("Failed to establish baseline for %s: %v", m.topics[i].Name, err)
        }
    }

    // Start monitoring loop
    ticker := time.NewTicker(m.updateInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            m.processUpdates(ctx)
        case <-ctx.Done():
            fmt.Println("🛑 Monitoring stopped")
            return
        }
    }
}

// establishBaseline creates initial data for a topic
func (m *ContinuousResearchMonitor) establishBaseline(ctx context.Context, topic *MonitoredTopic) error {
    fmt.Printf("📊 Establishing baseline for: %s\n", topic.Name)
    
    // Collect feed data
    feedData, err := m.collectFeedData(ctx, topic)
    if err != nil {
        log.Printf("Feed collection failed: %v", err)
    }

    // Collect search data
    searchData, err := m.collectSearchData(ctx, topic)
    if err != nil {
        log.Printf("Search collection failed: %v", err)
    }

    // Analyze baseline
    baseline, err := m.analyzeBaseline(ctx, topic, feedData, searchData)
    if err != nil {
        return fmt.Errorf("baseline analysis failed: %w", err)
    }

    topic.LastUpdate = time.Now()
    topic.ChangeScore = baseline.Score
    topic.TrendDirection = baseline.Direction
    
    fmt.Printf("✅ Baseline established for %s (Score: %.2f, Trend: %s)\n", 
        topic.Name, topic.ChangeScore, topic.TrendDirection)
    
    return nil
}

// processUpdates checks for changes in all monitored topics
func (m *ContinuousResearchMonitor) processUpdates(ctx context.Context) {
    fmt.Printf("\n🔄 Processing updates at %v\n", time.Now().Format("15:04:05"))
    
    for i := range m.topics {
        topic := &m.topics[i]
        
        // Collect new data
        newData, err := m.collectTopicUpdate(ctx, topic)
        if err != nil {
            log.Printf("Update collection failed for %s: %v", topic.Name, err)
            continue
        }

        // Analyze changes
        changes, err := m.analyzeChanges(ctx, topic, newData)
        if err != nil {
            log.Printf("Change analysis failed for %s: %v", topic.Name, err)
            continue
        }

        // Update topic state
        m.updateTopicState(topic, changes)

        // Generate alerts if significant changes detected
        if changes.Significance >= m.alertThreshold {
            alert := m.generateAlert(topic, changes)
            m.processAlert(alert)
        }

        fmt.Printf("📈 %s: Change Score %.2f → %.2f (%s)\n",
            topic.Name, topic.ChangeScore, changes.NewScore, changes.Direction)
    }
}

// collectFeedData gathers information from RSS/Atom feeds
func (m *ContinuousResearchMonitor) collectFeedData(ctx context.Context, topic *MonitoredTopic) ([]FeedItem, error) {
    var allItems []FeedItem
    
    for _, feedURL := range topic.FeedSources {
        // Fetch feed
        fetchParams := map[string]interface{}{
            "url": feedURL,
            "max_items": 50,
            "since": topic.LastUpdate.Format(time.RFC3339),
        }
        
        feedTool := feed.NewFeedFetchTool()
        result, err := feedTool.Execute(ctx, fetchParams)
        if err != nil {
            log.Printf("Failed to fetch feed %s: %v", feedURL, err)
            continue
        }

        // Filter for relevant items
        filterParams := map[string]interface{}{
            "items": result.Output,
            "keywords": topic.Keywords,
            "relevance_threshold": 6.0,
        }
        
        filterTool := feed.NewFeedFilterTool()
        filtered, err := filterTool.Execute(ctx, filterParams)
        if err != nil {
            log.Printf("Failed to filter feed %s: %v", feedURL, err)
            continue
        }

        // Convert to FeedItem structs
        if items, ok := filtered.Output.([]interface{}); ok {
            for _, item := range items {
                feedItem := m.convertToFeedItem(item, feedURL)
                allItems = append(allItems, feedItem)
            }
        }
    }
    
    return allItems, nil
}

// collectSearchData performs targeted searches for updates
func (m *ContinuousResearchMonitor) collectSearchData(ctx context.Context, topic *MonitoredTopic) ([]SearchUpdate, error) {
    var updates []SearchUpdate
    
    for _, query := range topic.SearchQueries {
        // Add time constraint to search
        timeConstrainedQuery := fmt.Sprintf("%s since:%s", query, 
            topic.LastUpdate.Format("2006-01-02"))
        
        searchParams := map[string]interface{}{
            "query": timeConstrainedQuery,
            "max_results": 10,
            "engines": []string{"tavily", "brave"},
        }
        
        searchTool := web.NewWebSearchTool()
        result, err := searchTool.Execute(ctx, searchParams)
        if err != nil {
            log.Printf("Search failed for query '%s': %v", query, err)
            continue
        }

        // Process search results
        if results, ok := result.Output.([]interface{}); ok {
            for _, r := range results {
                update := m.convertToSearchUpdate(r, query)
                updates = append(updates, update)
            }
        }
    }
    
    return updates, nil
}

// analyzeChanges compares new data with baseline
func (m *ContinuousResearchMonitor) analyzeChanges(ctx context.Context, topic *MonitoredTopic, newData *TopicUpdate) (*ChangeAnalysis, error) {
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf(`Analyze changes in monitored topic "%s":

Current Status:
- Last Score: %.2f
- Trend Direction: %s
- Last Update: %s

New Data:
- Feed Items: %d
- Search Results: %d
- Key Developments: %v

Assess:
1. Significance of changes (0-10 scale)
2. New trend direction (up/down/stable)
3. Key developments and their impact
4. Recommended actions

Provide detailed analysis.`, 
        topic.Name, topic.ChangeScore, topic.TrendDirection, 
        topic.LastUpdate.Format("2006-01-02 15:04"),
        len(newData.FeedItems), len(newData.SearchUpdates),
        newData.KeyDevelopments))

    result, err := m.updateAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    // Parse analysis (simplified - in real implementation, use structured output)
    analysis := &ChangeAnalysis{
        Significance:     8.5, // Extract from LLM response
        Direction:        "up",
        NewScore:        topic.ChangeScore + 1.2,
        KeyDevelopments: newData.KeyDevelopments,
        Impact:          "high",
        Recommendations: []string{"Monitor closely", "Consider deep dive analysis"},
    }

    return analysis, nil
}

// generateAlert creates alert from significant changes
func (m *ContinuousResearchMonitor) generateAlert(topic *MonitoredTopic, changes *ChangeAnalysis) Alert {
    severity := "medium"
    if changes.Significance >= 9.0 {
        severity = "high"
    } else if changes.Significance >= 7.0 {
        severity = "medium"
    } else {
        severity = "low"
    }

    return Alert{
        Type:      "trend_change",
        Message:   fmt.Sprintf("Significant change detected in %s (Score: %.2f, Impact: %s)", topic.Name, changes.Significance, changes.Impact),
        Severity:  severity,
        Timestamp: time.Now(),
        Topic:     topic.Name,
        Source:    "continuous_monitor",
    }
}

// processAlert handles alert delivery and logging
func (m *ContinuousResearchMonitor) processAlert(alert Alert) {
    fmt.Printf("🚨 ALERT [%s]: %s\n", strings.ToUpper(alert.Severity), alert.Message)
    
    // In production, send to notification systems
    switch alert.Severity {
    case "high":
        // Send immediate notifications (email, Slack, webhook)
        log.Printf("HIGH PRIORITY ALERT: %s", alert.Message)
    case "medium":
        // Send regular notifications
        log.Printf("MEDIUM PRIORITY ALERT: %s", alert.Message)
    case "low":
        // Log only
        log.Printf("LOW PRIORITY ALERT: %s", alert.Message)
    }
}

// Supporting types and helper methods
type FeedItem struct {
    Title       string
    Description string
    URL         string
    PubDate     time.Time
    Source      string
    Relevance   float64
}

type SearchUpdate struct {
    Title       string
    URL         string
    Snippet     string
    Query       string
    Timestamp   time.Time
    Relevance   float64
}

type TopicUpdate struct {
    FeedItems       []FeedItem
    SearchUpdates   []SearchUpdate
    KeyDevelopments []string
    Timestamp       time.Time
}

type ChangeAnalysis struct {
    Significance     float64
    Direction        string
    NewScore        float64
    KeyDevelopments []string
    Impact          string
    Recommendations []string
}

type BaselineData struct {
    Score     float64
    Direction string
    Items     []interface{}
}

// Helper implementations (simplified)
func (m *ContinuousResearchMonitor) collectTopicUpdate(ctx context.Context, topic *MonitoredTopic) (*TopicUpdate, error) {
    feedItems, _ := m.collectFeedData(ctx, topic)
    searchUpdates, _ := m.collectSearchData(ctx, topic)
    
    return &TopicUpdate{
        FeedItems:     feedItems,
        SearchUpdates: searchUpdates,
        KeyDevelopments: []string{"Development 1", "Development 2"},
        Timestamp:     time.Now(),
    }, nil
}

func (m *ContinuousResearchMonitor) analyzeBaseline(ctx context.Context, topic *MonitoredTopic, feedData []FeedItem, searchData []SearchUpdate) (*BaselineData, error) {
    return &BaselineData{
        Score:     7.5,
        Direction: "stable",
        Items:     []interface{}{feedData, searchData},
    }, nil
}

func (m *ContinuousResearchMonitor) updateTopicState(topic *MonitoredTopic, changes *ChangeAnalysis) {
    topic.ChangeScore = changes.NewScore
    topic.TrendDirection = changes.Direction
    topic.LastUpdate = time.Now()
}

func (m *ContinuousResearchMonitor) convertToFeedItem(item interface{}, source string) FeedItem {
    return FeedItem{
        Title:     "Feed item title",
        Source:    source,
        Timestamp: time.Now(),
        Relevance: 8.0,
    }
}

func (m *ContinuousResearchMonitor) convertToSearchUpdate(item interface{}, query string) SearchUpdate {
    return SearchUpdate{
        Title:     "Search result title",
        Query:     query,
        Timestamp: time.Now(),
        Relevance: 7.5,
    }
}

func main() {
    fmt.Println("📡 Continuous Research Monitoring")
    fmt.Println("=================================")

    monitor, err := NewContinuousResearchMonitor()
    if err != nil {
        log.Fatalf("Failed to create monitor: %v", err)
    }

    // Define topics to monitor
    topics := []MonitoredTopic{
        {
            Name:     "AI Safety Research",
            Keywords: []string{"AI safety", "alignment", "AGI", "existential risk"},
            FeedSources: []string{
                "https://www.anthropic.com/feed.xml",
                "https://openai.com/blog/rss.xml",
                "https://deepmind.google/feed.xml",
            },
            SearchQueries: []string{
                "AI safety research breakthrough",
                "artificial intelligence alignment progress",
                "AGI safety developments",
            },
        },
        {
            Name:     "Go Language Development",
            Keywords: []string{"golang", "go language", "go programming"},
            FeedSources: []string{
                "https://go.dev/blog/feed.atom",
                "https://blog.golang.org/feed.atom",
            },
            SearchQueries: []string{
                "Go language new features",
                "golang performance improvements",
                "go programming updates",
            },
        },
    }

    // Start monitoring (would run indefinitely in production)
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    monitor.StartMonitoring(ctx, topics)
}
```

### Monitoring Features
✅ **Feed Processing** - RSS/Atom feed monitoring and filtering  
✅ **Trend Analysis** - Track changes over time with scoring  
✅ **Smart Alerts** - Context-aware notifications based on significance  
✅ **Baseline Establishment** - Compare current state with historical data  
✅ **Multi-Source Integration** - Combine feeds and search results  
✅ **Relevance Filtering** - Focus on important developments only  

---

## Error Handling and Reliability

### Robust Error Recovery
```go
// Implement comprehensive error handling for production research agents
type ResearchErrorHandler struct {
    retryPolicy     *RetryPolicy
    fallbackSources []string
    errorAggregator *ErrorAggregator
    circuit         *CircuitBreaker
}

// RetryPolicy defines retry behavior for different error types
type RetryPolicy struct {
    NetworkErrors   RetryConfig
    RateLimitErrors RetryConfig
    APIErrors       RetryConfig
    TimeoutErrors   RetryConfig
}

type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Multiplier  float64
    Jitter      bool
}

// Example usage with graceful degradation
func (r *ResearchAgent) SearchWithFallback(ctx context.Context, query string) ([]SearchResult, error) {
    // Try primary search engines
    for _, engine := range r.primaryEngines {
        results, err := r.searchEngine(ctx, engine, query)
        if err == nil {
            return results, nil
        }
        
        // Handle specific error types
        switch {
        case isRateLimitError(err):
            r.handleRateLimit(engine)
            continue
        case isNetworkError(err):
            r.handleNetworkError(engine)
            continue
        case isAPIError(err):
            r.handleAPIError(engine, err)
            continue
        }
    }
    
    // Fallback to cached results if available
    if cached := r.getCachedResults(query); len(cached) > 0 {
        log.Printf("Using cached results for query: %s", query)
        return cached, nil
    }
    
    // Final fallback to mock data for development
    if r.config.EnableMockFallback {
        return r.generateMockResults(query), nil
    }
    
    return nil, fmt.Errorf("all search engines failed")
}
```

## Best Practices

### Research Agent Design Patterns

1. **Progressive Enhancement**
   - Start with simple search
   - Add content fetching
   - Include analysis layers
   - Add monitoring capabilities

2. **Tool Composition**
   - Combine complementary tools
   - Chain operations logically
   - Handle failures gracefully
   - Cache expensive operations

3. **State Management**
   - Track research progress
   - Persist intermediate results
   - Enable resume/retry
   - Version research sessions

4. **Quality Assurance**
   - Validate source credibility
   - Filter duplicate content
   - Assess information freshness
   - Maintain source attribution

## Next Steps

🔍 **Research agents mastered!** Continue with:

- **[Building Automation Agents](building-automation-agents.md)** - Task automation workflows
- **[Agent Communication](agent-communication.md)** - Coordination and handoffs
- **[Multi-Provider Strategies](multi-provider-strategies.md)** - Reliability optimization
- **[Data Validation](data-validation.md)** - Validation and error recovery

### Quick Reference

- **[Built-in Tools Reference](../reference/built-in-tools-reference.md)** - Complete tool catalog
- **[Configuration Reference](../reference/configuration-reference.md)** - All configuration options
- **[Best Practices Checklist](../reference/best-practices-checklist.md)** - Production checklist

---

**Need help?** Check our [troubleshooting guide](../advanced/troubleshooting.md) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).