// ABOUTME: Web search tool that provides search capabilities using various search engines
// ABOUTME: Built-in tool for performing web searches with configurable result limits and engines

package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// WebSearchParams defines parameters for the WebSearch tool
type WebSearchParams struct {
	Query       string `json:"query"`
	Engine      string `json:"engine,omitempty"`      // Search engine to use (default: "duckduckgo")
	MaxResults  int    `json:"max_results,omitempty"` // Maximum number of results (default: 10)
	SafeSearch  bool   `json:"safe_search,omitempty"` // Enable safe search (default: true)
	TimeoutSecs int    `json:"timeout,omitempty"`     // Timeout in seconds (default: 30)
}

// SearchResult defines a single search result
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Snippet     string `json:"snippet,omitempty"`
}

// WebSearchResults defines the result of the WebSearch tool
type WebSearchResults struct {
	Query      string         `json:"query"`
	Engine     string         `json:"engine"`
	Results    []SearchResult `json:"results"`
	TotalFound int            `json:"total_found,omitempty"`
	TimeMs     int64          `json:"time_ms"`
}

// webSearchParamSchema defines parameters for the WebSearch tool
var webSearchParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"query": {
			Type:        "string",
			Description: "The search query",
		},
		"engine": {
			Type:        "string",
			Description: "Search engine to use (duckduckgo, searx, or custom)",
		},
		"max_results": {
			Type:        "number",
			Description: "Maximum number of results to return (default: 10, max: 50)",
		},
		"safe_search": {
			Type:        "boolean",
			Description: "Enable safe search filtering (default: true)",
		},
		"timeout": {
			Type:        "number",
			Description: "Request timeout in seconds (default: 30)",
		},
	},
	Required: []string{"query"},
}

// DuckDuckGoResult represents a search result from DuckDuckGo API
type DuckDuckGoResult struct {
	FirstURL string `json:"FirstURL"`
	Text     string `json:"Text"`
	Result   string `json:"Result"`
}

// DuckDuckGoResponse represents the response from DuckDuckGo API
type DuckDuckGoResponse struct {
	Abstract       string             `json:"Abstract"`
	AbstractText   string             `json:"AbstractText"`
	AbstractSource string             `json:"AbstractSource"`
	AbstractURL    string             `json:"AbstractURL"`
	Results        []DuckDuckGoResult `json:"Results"`
	RelatedTopics  []interface{}      `json:"RelatedTopics"`
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("web_search", WebSearch(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "web_search",
			Category:    "web",
			Tags:        []string{"search", "web", "query", "internet", "network"},
			Description: "Performs web searches using various search engines",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic search",
					Description: "Search for Go programming information",
					Code:        `WebSearch().Execute(ctx, WebSearchParams{Query: "golang concurrency patterns"})`,
				},
				{
					Name:        "Limited results",
					Description: "Search with limited results",
					Code:        `WebSearch().Execute(ctx, WebSearchParams{Query: "machine learning", MaxResults: 5})`,
				},
			},
		},
		RequiredPermissions: []string{"network:access"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     true,
			FileSystem:  false,
			Concurrency: true,
		},
	})
}

// WebSearch creates a tool for performing web searches
// This is a built-in tool optimized for:
// - Multiple search engine support
// - Context-aware cancellation
// - Safe search filtering
// - Result limit control
func WebSearch() domain.Tool {
	return atools.NewTool(
		"web_search",
		"Performs web searches using various search engines",
		func(ctx *domain.ToolContext, params WebSearchParams) (*WebSearchResults, error) {
			startTime := time.Now()

			// Emit start event
			if ctx.Events != nil {
				ctx.Events.EmitMessage(fmt.Sprintf("Starting web search for '%s'", params.Query))
			}

			// Set defaults
			if params.Engine == "" {
				params.Engine = "duckduckgo"
			}
			if params.MaxResults == 0 {
				params.MaxResults = 10
			} else if params.MaxResults > 50 {
				params.MaxResults = 50 // Cap at 50 results
			}
			if params.TimeoutSecs == 0 {
				params.TimeoutSecs = 30
			}

			// Check state for custom search engine preferences
			if ctx.State != nil {
				if engine, exists := ctx.State.Get("search_engine"); exists {
					if engineStr, ok := engine.(string); ok {
						params.Engine = engineStr
					}
				}
				// Check for API keys or custom search endpoints
				if apiKey, exists := ctx.State.Get("search_api_key"); exists {
					// Store for use in search functions (could be passed via context)
					ctx.Context = context.WithValue(ctx.Context, "search_api_key", apiKey)
				}
			}

			// Set safe search default to true
			safeSearch := true
			if !params.SafeSearch {
				safeSearch = false
			}

			timeout := time.Duration(params.TimeoutSecs) * time.Second
			client := &http.Client{
				Timeout: timeout,
			}

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(1, 4, "Search engine configured")
			}

			var results []SearchResult
			var err error

			switch params.Engine {
			case "duckduckgo":
				results, err = searchDuckDuckGo(ctx, client, params.Query, params.MaxResults, safeSearch)
			case "searx":
				results, err = searchSearx(ctx, client, params.Query, params.MaxResults, safeSearch)
			default:
				return nil, fmt.Errorf("unsupported search engine: %s", params.Engine)
			}

			if err != nil {
				if ctx.Events != nil {
					ctx.Events.EmitError(err)
				}
				return nil, fmt.Errorf("search failed: %w", err)
			}

			// Emit completion event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(4, 4, "Search complete")
				ctx.Events.EmitCustom("search_complete", map[string]interface{}{
					"query":        params.Query,
					"engine":       params.Engine,
					"resultCount":  len(results),
					"timeMs":       time.Since(startTime).Milliseconds(),
				})
			}

			return &WebSearchResults{
				Query:      params.Query,
				Engine:     params.Engine,
				Results:    results,
				TotalFound: len(results),
				TimeMs:     time.Since(startTime).Milliseconds(),
			}, nil
		},
		webSearchParamSchema,
	)
}

// searchDuckDuckGo performs a search using DuckDuckGo Instant Answer API
func searchDuckDuckGo(ctx *domain.ToolContext, client *http.Client, query string, maxResults int, safeSearch bool) ([]SearchResult, error) {
	// Emit progress
	if ctx.Events != nil {
		ctx.Events.EmitProgress(2, 4, "Querying DuckDuckGo")
	}

	// Build DuckDuckGo API URL
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("no_html", "1")
	params.Set("skip_disambig", "1")
	if safeSearch {
		params.Set("safe_search", "1")
	}

	apiURL := "https://api.duckduckgo.com/?" + params.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx.Context, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set user agent
	userAgent := "go-llms/1.0"
	if ctx.State != nil {
		if ua, exists := ctx.State.Get("user_agent"); exists {
			if uaStr, ok := ua.(string); ok {
				userAgent = uaStr
			}
		}
	}
	req.Header.Set("User-Agent", userAgent)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Emit progress
	if ctx.Events != nil {
		ctx.Events.EmitProgress(3, 4, "Processing results")
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Parse response
	var ddgResp DuckDuckGoResponse
	if err := json.Unmarshal(body, &ddgResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Convert to our format
	var results []SearchResult

	// Add abstract if available
	if ddgResp.Abstract != "" && ddgResp.AbstractURL != "" {
		results = append(results, SearchResult{
			Title:       ddgResp.AbstractSource,
			URL:         ddgResp.AbstractURL,
			Description: ddgResp.AbstractText,
			Snippet:     ddgResp.Abstract,
		})
	}

	// Add results
	for i, r := range ddgResp.Results {
		if i >= maxResults {
			break
		}
		if r.FirstURL != "" {
			results = append(results, SearchResult{
				Title:       extractTitle(r.Result),
				URL:         r.FirstURL,
				Description: r.Text,
				Snippet:     r.Text,
			})
		}
	}

	// If no direct results, try to extract from related topics
	if len(results) < maxResults && len(ddgResp.RelatedTopics) > 0 {
		for _, topic := range ddgResp.RelatedTopics {
			if len(results) >= maxResults {
				break
			}
			// Related topics can be nested, handle carefully
			if topicMap, ok := topic.(map[string]interface{}); ok {
				if firstURL, ok := topicMap["FirstURL"].(string); ok && firstURL != "" {
					title := ""
					if text, ok := topicMap["Text"].(string); ok {
						title = extractTitle(text)
					}
					results = append(results, SearchResult{
						Title:       title,
						URL:         firstURL,
						Description: topicMap["Text"].(string),
					})
				}
			}
		}
	}

	return results, nil
}

// searchSearx performs a search using a Searx instance
func searchSearx(ctx *domain.ToolContext, client *http.Client, query string, maxResults int, safeSearch bool) ([]SearchResult, error) {
	// Check state for Searx instance URL
	if ctx.State != nil {
		if searxURL, exists := ctx.State.Get("searx_url"); exists {
			if urlStr, ok := searxURL.(string); ok {
				// Emit progress
				if ctx.Events != nil {
					ctx.Events.EmitProgress(2, 4, "Querying Searx instance")
				}
				
				// Implementation would go here for custom Searx instance
				// For now, still return error
				return nil, fmt.Errorf("searx search implementation pending for URL: %s", urlStr)
			}
		}
	}
	
	// For now, return an error as Searx requires a running instance
	return nil, fmt.Errorf("searx search not implemented - requires searx_url in state")
}

// extractTitle attempts to extract a title from DuckDuckGo result text
func extractTitle(text string) string {
	// DuckDuckGo often returns results in format "Title - Description"
	parts := strings.SplitN(text, " - ", 2)
	if len(parts) > 0 {
		title := strings.TrimSpace(parts[0])
		// Truncate title if it's too long
		if len(title) > 50 {
			return title[:50] + "..."
		}
		return title
	}
	// Fallback to first 50 characters
	if len(text) > 50 {
		return text[:50] + "..."
	}
	return text
}

// MustGetWebSearch retrieves the registered WebSearch tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetWebSearch() domain.Tool {
	return tools.MustGetTool("web_search")
}