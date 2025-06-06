// ABOUTME: Example demonstrating parallel web searches with different API keys
// ABOUTME: Shows how to use the engine_api_key parameter for production scenarios

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	_ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
)

// SearchConfig holds API keys for different search engines
type SearchConfig struct {
	BraveAPIKey     string
	TavilyAPIKey    string
	SerpapiAPIKey   string
	SerperdevAPIKey string
}

// ParallelSearcher performs parallel searches across multiple engines
type ParallelSearcher struct {
	config SearchConfig
	tool   domain.Tool
}

// NewParallelSearcher creates a new parallel searcher
func NewParallelSearcher(config SearchConfig) *ParallelSearcher {
	tool, ok := tools.GetTool("web_search")
	if !ok {
		log.Fatal("web_search tool not found")
	}

	return &ParallelSearcher{
		config: config,
		tool:   tool,
	}
}

// SearchResult holds the result from a search engine
type SearchResult struct {
	Engine  string
	Query   string
	Results interface{}
	Error   error
	TimeMs  int64
}

// SearchAll performs searches on all available engines in parallel
func (ps *ParallelSearcher) SearchAll(ctx context.Context, query string) []SearchResult {
	var wg sync.WaitGroup
	results := make([]SearchResult, 0, 5)
	resultChan := make(chan SearchResult, 5)

	// Search with Brave if API key available
	if ps.config.BraveAPIKey != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()

			// Create a minimal ToolContext without an agent
			state := domain.NewState()
			toolCtx := &domain.ToolContext{
				Context:   ctx,
				State:     domain.NewStateReader(state),
				RunID:     "search-brave",
				StartTime: time.Now(),
			}
			result, err := ps.tool.Execute(toolCtx, map[string]interface{}{
				"query":          query,
				"engine":         "brave",
				"engine_api_key": ps.config.BraveAPIKey,
				"max_results":    5,
			})

			resultChan <- SearchResult{
				Engine:  "brave",
				Query:   query,
				Results: result,
				Error:   err,
				TimeMs:  time.Since(start).Milliseconds(),
			}
		}()
	}

	// Search with Tavily if API key available
	if ps.config.TavilyAPIKey != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()

			// Create a minimal ToolContext without an agent
			state := domain.NewState()
			toolCtx := &domain.ToolContext{
				Context:   ctx,
				State:     domain.NewStateReader(state),
				RunID:     "search-tavily",
				StartTime: time.Now(),
			}
			result, err := ps.tool.Execute(toolCtx, map[string]interface{}{
				"query":          query,
				"engine":         "tavily",
				"engine_api_key": ps.config.TavilyAPIKey,
				"max_results":    5,
			})

			resultChan <- SearchResult{
				Engine:  "tavily",
				Query:   query,
				Results: result,
				Error:   err,
				TimeMs:  time.Since(start).Milliseconds(),
			}
		}()
	}

	// Search with Serpapi if API key available
	if ps.config.SerpapiAPIKey != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()

			// Create a minimal ToolContext without an agent
			state := domain.NewState()
			toolCtx := &domain.ToolContext{
				Context:   ctx,
				State:     domain.NewStateReader(state),
				RunID:     "search-serpapi",
				StartTime: time.Now(),
			}
			result, err := ps.tool.Execute(toolCtx, map[string]interface{}{
				"query":          query,
				"engine":         "serpapi",
				"engine_api_key": ps.config.SerpapiAPIKey,
				"max_results":    5,
			})

			resultChan <- SearchResult{
				Engine:  "serpapi",
				Query:   query,
				Results: result,
				Error:   err,
				TimeMs:  time.Since(start).Milliseconds(),
			}
		}()
	}

	// Search with Serper.dev if API key available
	if ps.config.SerperdevAPIKey != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()

			// Create a minimal ToolContext without an agent
			state := domain.NewState()
			toolCtx := &domain.ToolContext{
				Context:   ctx,
				State:     domain.NewStateReader(state),
				RunID:     "search-serperdev",
				StartTime: time.Now(),
			}
			result, err := ps.tool.Execute(toolCtx, map[string]interface{}{
				"query":          query,
				"engine":         "serperdev",
				"engine_api_key": ps.config.SerperdevAPIKey,
				"max_results":    5,
			})

			resultChan <- SearchResult{
				Engine:  "serperdev",
				Query:   query,
				Results: result,
				Error:   err,
				TimeMs:  time.Since(start).Milliseconds(),
			}
		}()
	}

	// Always include DuckDuckGo as fallback
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()

		// Create a minimal ToolContext without an agent
		state := domain.NewState()
		toolCtx := &domain.ToolContext{
			Context:   ctx,
			State:     domain.NewStateReader(state),
			RunID:     "search-duckduckgo",
			StartTime: time.Now(),
		}
		result, err := ps.tool.Execute(toolCtx, map[string]interface{}{
			"query":       query,
			"engine":      "duckduckgo",
			"max_results": 5,
		})

		resultChan <- SearchResult{
			Engine:  "duckduckgo",
			Query:   query,
			Results: result,
			Error:   err,
			TimeMs:  time.Since(start).Milliseconds(),
		}
	}()

	// Close channel when all searches complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

func main() {
	// Get API keys from environment variables
	// In production, these would come from a secure key management system
	config := SearchConfig{
		BraveAPIKey:     os.Getenv("BRAVE_API_KEY"),
		TavilyAPIKey:    os.Getenv("TAVILY_API_KEY"),
		SerpapiAPIKey:   os.Getenv("SERPAPI_API_KEY"),
		SerperdevAPIKey: os.Getenv("SERPERDEV_API_KEY"),
	}

	// Check if any keys are available
	if config.BraveAPIKey == "" && config.TavilyAPIKey == "" && config.SerpapiAPIKey == "" && config.SerperdevAPIKey == "" {
		fmt.Println("No API keys found in environment variables.")
		fmt.Println("Please set one or more of the following:")
		fmt.Println("  export BRAVE_API_KEY='your-brave-api-key'")
		fmt.Println("  export TAVILY_API_KEY='your-tavily-api-key'")
		fmt.Println("  export SERPAPI_API_KEY='your-serpapi-api-key'")
		fmt.Println("  export SERPERDEV_API_KEY='your-serperdev-api-key'")
		fmt.Println("\nThe example will still run with DuckDuckGo (no API key required).")
		fmt.Println()
	}

	// Show which engines will be used
	fmt.Println("Search engines configured:")
	if config.BraveAPIKey != "" {
		fmt.Println("  ✓ Brave Search (using BRAVE_API_KEY)")
	}
	if config.TavilyAPIKey != "" {
		fmt.Println("  ✓ Tavily Search (using TAVILY_API_KEY)")
	}
	if config.SerpapiAPIKey != "" {
		fmt.Println("  ✓ Serpapi Search (using SERPAPI_API_KEY)")
	}
	if config.SerperdevAPIKey != "" {
		fmt.Println("  ✓ Serper.dev Search (using SERPERDEV_API_KEY)")
	}
	fmt.Println("  ✓ DuckDuckGo (no API key required)")
	fmt.Println()

	searcher := NewParallelSearcher(config)
	ctx := context.Background()

	query := "artificial intelligence latest developments"
	fmt.Printf("Searching for: %s\n\n", query)

	results := searcher.SearchAll(ctx, query)

	// Display results
	for _, result := range results {
		fmt.Printf("Engine: %s\n", result.Engine)
		fmt.Printf("Time: %dms\n", result.TimeMs)

		if result.Error != nil {
			fmt.Printf("Error: %v\n", result.Error)
		} else {
			fmt.Printf("Success: Retrieved search results\n")
			// In a real application, you would process the results here
		}
		fmt.Println()
	}

	// Example: Compare results from different engines
	fmt.Println("Search Summary:")
	fmt.Printf("Total engines searched: %d\n", len(results))

	successCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		}
	}
	fmt.Printf("Successful searches: %d\n", successCount)

	// Find fastest engine
	if len(results) > 0 {
		fastest := results[0]
		for _, r := range results[1:] {
			if r.Error == nil && (fastest.Error != nil || r.TimeMs < fastest.TimeMs) {
				fastest = r
			}
		}
		if fastest.Error == nil {
			fmt.Printf("Fastest engine: %s (%dms)\n", fastest.Engine, fastest.TimeMs)
		}
	}
}
