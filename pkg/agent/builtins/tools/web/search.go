// ABOUTME: Web search tool that provides search capabilities using various search engines
// ABOUTME: Built-in tool for performing web searches with configurable result limits and engines

package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// contextKey is a custom type to avoid collisions in context values
type contextKey string

const searchAPIKeyContext contextKey = "search_api_key"

// WebSearchParams defines parameters for the WebSearch tool
type WebSearchParams struct {
	Query        string `json:"query"`
	Engine       string `json:"engine,omitempty"`         // Search engine to use (default: "duckduckgo")
	EngineAPIKey string `json:"engine_api_key,omitempty"` // Optional API key for the search engine
	MaxResults   int    `json:"max_results,omitempty"`    // Maximum number of results (default: 10)
	SafeSearch   bool   `json:"safe_search,omitempty"`    // Enable safe search (default: true)
	TimeoutSecs  int    `json:"timeout,omitempty"`        // Timeout in seconds (default: 30)

	// Authentication parameters (optional)
	AuthType        string `json:"auth_type,omitempty"`         // "bearer", "basic", "api_key", "oauth2", "custom"
	AuthToken       string `json:"auth_token,omitempty"`        // Bearer token or API token
	AuthUsername    string `json:"auth_username,omitempty"`     // Username for basic auth
	AuthPassword    string `json:"auth_password,omitempty"`     // Password for basic auth
	AuthAPIKey      string `json:"auth_api_key,omitempty"`      // API key value
	AuthKeyName     string `json:"auth_key_name,omitempty"`     // API key name (default: X-API-Key)
	AuthKeyLocation string `json:"auth_key_location,omitempty"` // "header", "query", "cookie" (default: header)
	AuthHeaderName  string `json:"auth_header_name,omitempty"`  // Custom header name for custom auth
	AuthHeaderValue string `json:"auth_header_value,omitempty"` // Custom header value for custom auth
	AuthPrefix      string `json:"auth_prefix,omitempty"`       // Optional prefix for custom auth (e.g., "Token")
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
			Description: "Search engine to use (duckduckgo, brave, tavily, serpapi, serperdev, searx, or custom)",
		},
		"engine_api_key": {
			Type:        "string",
			Description: "Optional API key for the search engine (overrides environment variables)",
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
		"auth_type": {
			Type:        "string",
			Description: "Authentication type: 'bearer', 'basic', 'api_key', 'oauth2', 'custom'",
		},
		"auth_token": {
			Type:        "string",
			Description: "Bearer token or general authentication token",
		},
		"auth_username": {
			Type:        "string",
			Description: "Username for basic authentication",
		},
		"auth_password": {
			Type:        "string",
			Description: "Password for basic authentication",
		},
		"auth_api_key": {
			Type:        "string",
			Description: "API key value",
		},
		"auth_key_name": {
			Type:        "string",
			Description: "API key parameter name (default: X-API-Key)",
		},
		"auth_key_location": {
			Type:        "string",
			Description: "Where to place API key: 'header', 'query', or 'cookie' (default: header)",
		},
		"auth_header_name": {
			Type:        "string",
			Description: "Custom header name for custom authentication",
		},
		"auth_header_value": {
			Type:        "string",
			Description: "Custom header value for custom authentication",
		},
		"auth_prefix": {
			Type:        "string",
			Description: "Optional prefix for custom auth header value (e.g., 'Token')",
		},
	},
	Required: []string{"query"},
}

// webSearchOutputSchema defines the output for the WebSearch tool
var webSearchOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"query": {
			Type:        "string",
			Description: "The search query that was executed",
		},
		"engine": {
			Type:        "string",
			Description: "The search engine that was used",
		},
		"results": {
			Type:        "array",
			Description: "Array of search results",
			Items: &sdomain.Property{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"title": {
						Type:        "string",
						Description: "Result title",
					},
					"url": {
						Type:        "string",
						Description: "Result URL",
					},
					"description": {
						Type:        "string",
						Description: "Result description",
					},
					"snippet": {
						Type:        "string",
						Description: "Optional result snippet or excerpt",
					},
				},
				Required: []string{"title", "url", "description"},
			},
		},
		"total_found": {
			Type:        "number",
			Description: "Total number of results found",
		},
		"time_ms": {
			Type:        "number",
			Description: "Search execution time in milliseconds",
		},
	},
	Required: []string{"query", "engine", "results"},
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

// BraveSearchResponse represents the response from Brave Search API
type BraveSearchResponse struct {
	Query   interface{} `json:"query"` // Can be string or object
	Results struct {
		News   []BraveResult `json:"news"`
		Web    []BraveResult `json:"web"`
		Videos []BraveResult `json:"videos"`
	} `json:"results"`
}

// BraveResult represents a single result from Brave Search
type BraveResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	PageAge     string `json:"page_age,omitempty"`
}

// TavilySearchRequest represents the request to Tavily API
type TavilySearchRequest struct {
	APIKey        string `json:"api_key"`
	Query         string `json:"query"`
	MaxResults    int    `json:"max_results,omitempty"`
	IncludeAnswer bool   `json:"include_answer,omitempty"`
}

// TavilySearchResponse represents the response from Tavily API
type TavilySearchResponse struct {
	Query   string         `json:"query"`
	Answer  string         `json:"answer,omitempty"`
	Results []TavilyResult `json:"results"`
}

// TavilyResult represents a single result from Tavily
type TavilyResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

// SerpapiSearchResponse represents the response from Serpapi API
type SerpapiSearchResponse struct {
	OrganicResults []SerpapiResult `json:"organic_results"`
	SearchMetadata struct {
		TotalResults string `json:"total_results,omitempty"`
	} `json:"search_metadata"`
}

// SerpapiResult represents a single result from Serpapi
type SerpapiResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// SerperDevSearchRequest represents the request to Serper.dev API
type SerperDevSearchRequest struct {
	Q          string `json:"q"`
	Num        int    `json:"num,omitempty"`
	SafeSearch string `json:"safe,omitempty"` // "off" or "active"
}

// SerperDevSearchResponse represents the response from Serper.dev API
type SerperDevSearchResponse struct {
	Organic []SerperDevResult `json:"organic"`
}

// SerperDevResult represents a single result from Serper.dev
type SerperDevResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("web_search", WebSearch(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "web_search",
			Category:    "web",
			Tags:        []string{"search", "web", "query", "internet", "network", "brave", "tavily", "duckduckgo", "serpapi", "serperdev", "google"},
			Description: "Performs web searches using various search engines (DuckDuckGo, Brave, Tavily, Serpapi, Serper.dev)",
			Version:     "2.0.0",
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
				{
					Name:        "Brave search",
					Description: "Use Brave Search API (requires BRAVE_API_KEY)",
					Code:        `WebSearch().Execute(ctx, WebSearchParams{Query: "AI news", Engine: "brave"})`,
				},
				{
					Name:        "Tavily search",
					Description: "Use Tavily Search API optimized for LLMs (requires TAVILY_API_KEY)",
					Code:        `WebSearch().Execute(ctx, WebSearchParams{Query: "quantum computing", Engine: "tavily"})`,
				},
				{
					Name:        "Serpapi search",
					Description: "Use Serpapi Search API for Google results (requires SERPAPI_API_KEY)",
					Code:        `WebSearch().Execute(ctx, WebSearchParams{Query: "latest technology trends", Engine: "serpapi"})`,
				},
				{
					Name:        "Serper.dev search",
					Description: "Use Serper.dev Search API for Google results (requires SERPERDEV_API_KEY)",
					Code:        `WebSearch().Execute(ctx, WebSearchParams{Query: "AI research papers", Engine: "serperdev"})`,
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

// WebSearch creates a tool for performing web searches using various search engines with automatic API key detection.
// It supports multiple search providers (DuckDuckGo, Brave, Tavily, Serpapi, Serper.dev), automatically selects
// the best available engine based on API keys, provides safe search filtering and result limit control,
// and includes authentication support for custom search endpoints with comprehensive error handling.
func WebSearch() domain.Tool {
	builder := atools.NewToolBuilder("web_search", "Performs web searches using various search engines with automatic API key detection").
		WithFunction(func(ctx *domain.ToolContext, params WebSearchParams) (*WebSearchResults, error) {
			startTime := time.Now()

			// Emit start event
			if ctx.Events != nil {
				ctx.Events.EmitMessage(fmt.Sprintf("Starting web search for '%s'", params.Query))
			}

			// Set defaults
			if params.Engine == "" {
				params.Engine = selectDefaultEngine()
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
					ctx.Context = context.WithValue(ctx.Context, searchAPIKeyContext, apiKey)
				}
			}

			// Set safe search default to true
			// nolint:staticcheck // QF1007: Can't merge because we need default true behavior
			safeSearch := true
			if !params.SafeSearch {
				// Only set to false if explicitly disabled
				// TODO: Consider using *bool to distinguish between unset and false
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
			case "brave":
				results, err = searchBrave(ctx, client, params.Query, params.MaxResults, safeSearch, params.EngineAPIKey)
			case "tavily":
				results, err = searchTavily(ctx, client, params.Query, params.MaxResults, safeSearch, params.EngineAPIKey)
			case "serpapi":
				results, err = searchSerpapi(ctx, client, params.Query, params.MaxResults, safeSearch, params.EngineAPIKey)
			case "serperdev":
				results, err = searchSerperDev(ctx, client, params.Query, params.MaxResults, safeSearch, params.EngineAPIKey)
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
					"query":       params.Query,
					"engine":      params.Engine,
					"resultCount": len(results),
					"timeMs":      time.Since(startTime).Milliseconds(),
				})
			}

			return &WebSearchResults{
				Query:      params.Query,
				Engine:     params.Engine,
				Results:    results,
				TotalFound: len(results),
				TimeMs:     time.Since(startTime).Milliseconds(),
			}, nil
		}).
		WithParameterSchema(webSearchParamSchema).
		WithOutputSchema(webSearchOutputSchema).
		WithUsageInstructions(`Use this tool to search the web using various search engines with optional authentication. The tool automatically selects the best available search engine based on API keys.

Available engines:
- duckduckgo: Free, no API key required, limited results
- brave: Comprehensive web search (requires BRAVE_API_KEY)
- tavily: AI-optimized search with summaries (requires TAVILY_API_KEY) - best for LLM applications
- serpapi: Google search results (requires SERPAPI_API_KEY)
- serperdev: Fast Google search results (requires SERPERDEV_API_KEY)
- searx: Privacy-focused metasearch (requires searx_url in state)

The tool will automatically:
- Select the best available engine based on API keys
- Handle rate limiting and retries
- Filter results based on safe search settings
- Limit results to the requested maximum (up to 50)

API Key Management:
- Set API keys via environment variables (BRAVE_API_KEY, TAVILY_API_KEY, etc.)
- Or provide engine_api_key parameter to override environment variables
- Keys in state (search_api_key) also work for backward compatibility

Authentication methods (for custom search endpoints):
- bearer: Sends "Authorization: Bearer <token>" header
- basic: Sends HTTP Basic Authentication with username/password
- api_key: Sends API key in header, query, or cookie
- oauth2: Sends OAuth2 access token as bearer token
- custom: Sends custom header with optional prefix

Authentication can be auto-detected from state or provided via parameters.`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Basic web search",
				Description: "Search for information using default engine",
				Scenario:    "When you need to find information on any topic",
				Input: map[string]interface{}{
					"query": "latest AI developments 2024",
				},
				Output: map[string]interface{}{
					"query":  "latest AI developments 2024",
					"engine": "tavily", // Assuming Tavily is configured
					"results": []map[string]interface{}{
						{
							"title":       "Major AI Breakthroughs in 2024",
							"url":         "https://example.com/ai-2024",
							"description": "Overview of significant AI advancements...",
							"snippet":     "In 2024, artificial intelligence saw unprecedented growth...",
						},
					},
					"total_found": 10,
					"time_ms":     342,
				},
				Explanation: "Automatically selected Tavily for AI-optimized results",
			},
			{
				Name:        "Search with specific engine",
				Description: "Use a specific search engine",
				Scenario:    "When you want to use a particular search provider",
				Input: map[string]interface{}{
					"query":       "python programming tutorials",
					"engine":      "brave",
					"max_results": 5,
				},
				Output: map[string]interface{}{
					"query":  "python programming tutorials",
					"engine": "brave",
					"results": []map[string]interface{}{
						{
							"title":       "Python Tutorial - W3Schools",
							"url":         "https://www.w3schools.com/python/",
							"description": "Well organized and easy to understand Web building tutorials",
						},
					},
					"total_found": 5,
					"time_ms":     215,
				},
				Explanation: "Used Brave Search as requested",
			},
			{
				Name:        "Search with API key override",
				Description: "Provide API key directly",
				Scenario:    "When using a different API key than environment variable",
				Input: map[string]interface{}{
					"query":          "climate change research papers",
					"engine":         "serpapi",
					"engine_api_key": "your-serpapi-key-here",
					"max_results":    20,
				},
				Output: map[string]interface{}{
					"query":       "climate change research papers",
					"engine":      "serpapi",
					"total_found": 20,
					"time_ms":     523,
				},
				Explanation: "Used provided API key instead of environment variable",
			},
			{
				Name:        "Search with safe search disabled",
				Description: "Search without content filtering",
				Scenario:    "When you need unfiltered results",
				Input: map[string]interface{}{
					"query":       "medical procedures",
					"safe_search": false,
				},
				Output: map[string]interface{}{
					"query":       "medical procedures",
					"engine":      "duckduckgo",
					"total_found": 10,
					"time_ms":     189,
				},
				Explanation: "Safe search disabled for medical/scientific content",
			},
			{
				Name:        "Handle missing API keys",
				Description: "Fallback to free engine",
				Scenario:    "When no API keys are configured",
				Input: map[string]interface{}{
					"query": "open source projects",
				},
				Output: map[string]interface{}{
					"query":       "open source projects",
					"engine":      "duckduckgo",
					"total_found": 8,
					"time_ms":     412,
				},
				Explanation: "Automatically fell back to DuckDuckGo (no API key required)",
			},
			{
				Name:        "Search with custom timeout",
				Description: "Set longer timeout for slow connections",
				Scenario:    "When dealing with slow network or complex queries",
				Input: map[string]interface{}{
					"query":   "comprehensive market analysis reports 2024",
					"timeout": 60,
				},
				Output: map[string]interface{}{
					"query":       "comprehensive market analysis reports 2024",
					"engine":      "tavily",
					"total_found": 15,
					"time_ms":     2341,
				},
				Explanation: "Extended timeout allowed for thorough search",
			},
			{
				Name:        "Error: Invalid engine",
				Description: "Handle unsupported engine",
				Scenario:    "When requesting a non-existent search engine",
				Input: map[string]interface{}{
					"query":  "test query",
					"engine": "invalid_engine",
				},
				Output: map[string]interface{}{
					"error": "unsupported search engine: invalid_engine",
				},
				Explanation: "Clear error for invalid engine selection",
			},
			{
				Name:        "Search with bearer token authentication",
				Description: "Search protected custom search endpoint",
				Scenario:    "When using a custom search service requiring bearer token",
				Input: map[string]interface{}{
					"query":      "internal documents",
					"engine":     "custom",
					"auth_type":  "bearer",
					"auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
				},
				Output: map[string]interface{}{
					"query":       "internal documents",
					"engine":      "custom",
					"total_found": 5,
					"time_ms":     452,
				},
				Explanation: "Bearer token added to Authorization header for authentication",
			},
			{
				Name:        "Search with API key authentication",
				Description: "Search with API key in custom header",
				Scenario:    "When custom search service requires API key in header",
				Input: map[string]interface{}{
					"query":         "research papers",
					"engine":        "custom",
					"auth_type":     "api_key",
					"auth_api_key":  "abc123xyz789",
					"auth_key_name": "X-Custom-API-Key",
				},
				Output: map[string]interface{}{
					"query":       "research papers",
					"engine":      "custom",
					"total_found": 12,
					"time_ms":     523,
				},
				Explanation: "API key sent in X-Custom-API-Key header for access",
			},
		}).
		WithConstraints([]string{
			"Maximum 50 results per search (API limitations)",
			"Safe search is enabled by default",
			"DuckDuckGo provides limited results compared to other engines",
			"API keys can be set via environment or engine_api_key parameter",
			"Timeout defaults to 30 seconds if not specified",
			"Some engines require paid API subscriptions",
			"Search results may vary between engines",
			"Rate limits apply based on API provider",
			"Searx requires a running instance URL in state",
			"Authentication is optional and supports bearer, basic, API key, OAuth2, and custom methods",
			"Auth credentials should be kept secure and not logged",
			"State-based auth detection looks for common token patterns",
		}).
		WithErrorGuidance(map[string]string{
			"unsupported search engine":        "Use one of: duckduckgo, brave, tavily, serpapi, serperdev, searx",
			"API key required":                 "Set the appropriate environment variable or use engine_api_key parameter",
			"rate limit exceeded":              "Wait before making more requests or upgrade your API plan",
			"timeout":                          "Increase timeout parameter or try a faster search engine",
			"no results found":                 "Try different search terms or another search engine",
			"invalid API key":                  "Check your API key is correct and has not expired",
			"searx_url not found":              "Set searx_url in agent state to use Searx engine",
			"context deadline exceeded":        "Search took too long - try reducing max_results or increasing timeout",
			"connection refused":               "Check internet connection or firewall settings",
			"API returned status":              "Check API service status or contact provider support",
			"error parsing response":           "API response format may have changed - report this issue",
			"brave Search API returned status": "Check Brave API key permissions and quota",
			"tavily API error":                 "Verify Tavily API key and account status",
			"serpapi request failed":           "Check Serpapi API key and request parameters",
			"serperdev API error":              "Verify Serper.dev API key and quota",
			"authentication failed":            "Check auth credentials are correct and match the expected format",
			"401 Unauthorized":                 "Authentication required - check auth_type and credentials",
			"403 Forbidden":                    "Access denied - credentials may be invalid or insufficient permissions",
		}).
		WithCategory("web").
		WithTags([]string{"web", "search", "internet", "query", "research", "auth", "authentication"}).
		WithVersion("3.0.0").
		WithBehavior(
			false,  // Not deterministic - results change over time
			false,  // Not destructive
			false,  // No confirmation needed
			"fast", // Usually fast, depends on engine and network
		)

	return builder.Build()
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
	defer func() {
		_ = resp.Body.Close()
	}()

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

// getSearchAPIKeys checks for search API keys from environment
func getSearchAPIKeys() (braveKey, tavilyKey, serpapiKey, serperdevKey string) {
	braveKey = os.Getenv("BRAVE_API_KEY")
	tavilyKey = os.Getenv("TAVILY_API_KEY")
	serpapiKey = os.Getenv("SERPAPI_API_KEY")
	serperdevKey = os.Getenv("SERPERDEV_API_KEY")
	return
}

// selectDefaultEngine auto-selects the best engine based on available API keys
func selectDefaultEngine() string {
	braveKey, tavilyKey, serpapiKey, serperdevKey := getSearchAPIKeys()

	if tavilyKey != "" {
		return "tavily" // Best for LLM use cases
	}
	if serperdevKey != "" {
		return "serperdev" // Fast Google search results
	}
	if serpapiKey != "" {
		return "serpapi" // Comprehensive Google search results
	}
	if braveKey != "" {
		return "brave" // Good general web search
	}
	return "duckduckgo" // Limited but no API key required
}

// searchBrave performs a search using Brave Search API
func searchBrave(ctx *domain.ToolContext, client *http.Client, query string, maxResults int, safeSearch bool, apiKey string) ([]SearchResult, error) {
	// Use provided API key first, fallback to environment
	braveKey := apiKey
	if braveKey == "" {
		braveKey, _, _, _ = getSearchAPIKeys()
	}
	if braveKey == "" {
		return nil, fmt.Errorf("brave Search API key required - set BRAVE_API_KEY environment variable")
	}

	// Emit progress
	if ctx.Events != nil {
		ctx.Events.EmitProgress(2, 4, "Querying Brave Search")
	}

	// Build URL with query parameters
	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", maxResults))
	if safeSearch {
		params.Set("safesearch", "strict")
	} else {
		params.Set("safesearch", "off")
	}

	apiURL := "https://api.search.brave.com/res/v1/web/search?" + params.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx.Context, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("X-Subscription-Token", braveKey)
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("brave Search API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Emit progress
	if ctx.Events != nil {
		ctx.Events.EmitProgress(3, 4, "Processing results")
	}

	// Parse response
	var braveResp BraveSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&braveResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Convert to our format
	var results []SearchResult

	// Add web results
	for i, r := range braveResp.Results.Web {
		if i >= maxResults {
			break
		}
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Description,
			Snippet:     r.Description,
		})
	}

	// Add news results if space allows
	remaining := maxResults - len(results)
	for i, r := range braveResp.Results.News {
		if i >= remaining {
			break
		}
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Description,
			Snippet:     r.Description,
		})
	}

	return results, nil
}

// searchTavily performs a search using Tavily API
func searchTavily(ctx *domain.ToolContext, client *http.Client, query string, maxResults int, safeSearch bool, apiKey string) ([]SearchResult, error) {
	// Use provided API key first, fallback to environment
	tavilyKey := apiKey
	if tavilyKey == "" {
		_, tavilyKey, _, _ = getSearchAPIKeys()
	}
	if tavilyKey == "" {
		return nil, fmt.Errorf("tavily API key required - set TAVILY_API_KEY environment variable")
	}

	// Emit progress
	if ctx.Events != nil {
		ctx.Events.EmitProgress(2, 4, "Querying Tavily Search")
	}

	// Build request
	tavilyReq := TavilySearchRequest{
		APIKey:        tavilyKey,
		Query:         query,
		MaxResults:    maxResults,
		IncludeAnswer: true,
	}

	reqBody, err := json.Marshal(tavilyReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx.Context, "POST", "https://api.tavily.com/search", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tavily API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Emit progress
	if ctx.Events != nil {
		ctx.Events.EmitProgress(3, 4, "Processing results")
	}

	// Parse response
	var tavilyResp TavilySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&tavilyResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Convert to our format
	var results []SearchResult

	// Add answer as first result if available
	if tavilyResp.Answer != "" {
		results = append(results, SearchResult{
			Title:       "AI Summary",
			URL:         fmt.Sprintf("tavily:answer:%s", query),
			Description: tavilyResp.Answer,
			Snippet:     tavilyResp.Answer,
		})
	}

	// Add search results
	for i, r := range tavilyResp.Results {
		if len(results) >= maxResults {
			break
		}
		// Skip if we already have too many
		if i >= maxResults {
			break
		}

		// Truncate content if too long
		snippet := r.Content
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}

		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: snippet,
			Snippet:     snippet,
		})
	}

	return results, nil
}

// searchSerpapi performs a search using Serpapi API
func searchSerpapi(ctx *domain.ToolContext, client *http.Client, query string, maxResults int, safeSearch bool, apiKey string) ([]SearchResult, error) {
	// Use provided API key first, fallback to environment
	serpapiKey := apiKey
	if serpapiKey == "" {
		_, _, serpapiKey, _ = getSearchAPIKeys()
	}
	if serpapiKey == "" {
		return nil, fmt.Errorf("serpapi API key required - set SERPAPI_API_KEY environment variable")
	}

	// Emit progress
	if ctx.Events != nil {
		ctx.Events.EmitProgress(2, 4, "Querying Serpapi Search")
	}

	// Create request
	// Serpapi uses GET requests with query parameters
	params := url.Values{}
	params.Set("q", query)
	params.Set("api_key", serpapiKey)
	params.Set("engine", "google")
	params.Set("num", fmt.Sprintf("%d", maxResults))
	if safeSearch {
		params.Set("safe", "active")
	}

	apiURL := "https://serpapi.com/search?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx.Context, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("serpapi API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Emit progress
	if ctx.Events != nil {
		ctx.Events.EmitProgress(3, 4, "Processing results")
	}

	// Parse response
	var serpapiResp SerpapiSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&serpapiResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Convert to our format
	var results []SearchResult
	for i, r := range serpapiResp.OrganicResults {
		if i >= maxResults {
			break
		}
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.Link,
			Description: r.Snippet,
			Snippet:     r.Snippet,
		})
	}

	return results, nil
}

// searchSerperDev performs a search using Serper.dev API
func searchSerperDev(ctx *domain.ToolContext, client *http.Client, query string, maxResults int, safeSearch bool, apiKey string) ([]SearchResult, error) {
	// Use provided API key first, fallback to environment
	serperdevKey := apiKey
	if serperdevKey == "" {
		_, _, _, serperdevKey = getSearchAPIKeys()
	}
	if serperdevKey == "" {
		return nil, fmt.Errorf("serper.dev API key required - set SERPERDEV_API_KEY environment variable")
	}

	// Emit progress
	if ctx.Events != nil {
		ctx.Events.EmitProgress(2, 4, "Querying Serper.dev Search")
	}

	// Build request
	serperdevReq := SerperDevSearchRequest{
		Q:   query,
		Num: maxResults,
	}

	if safeSearch {
		serperdevReq.SafeSearch = "active"
	} else {
		serperdevReq.SafeSearch = "off"
	}

	reqBody, err := json.Marshal(serperdevReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx.Context, "POST", "https://google.serper.dev/search", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("X-API-KEY", serperdevKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("serper.dev API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Emit progress
	if ctx.Events != nil {
		ctx.Events.EmitProgress(3, 4, "Processing results")
	}

	// Parse response
	var serperdevResp SerperDevSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&serperdevResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Convert to our format
	var results []SearchResult
	for i, r := range serperdevResp.Organic {
		if i >= maxResults {
			break
		}
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.Link,
			Description: r.Snippet,
			Snippet:     r.Snippet,
		})
	}

	return results, nil
}

// MustGetWebSearch retrieves the registered WebSearch tool or panics if not found.
// This is a convenience function for users who want to ensure the tool exists
// and prefer a panic over error handling for missing tools in their initialization code.
func MustGetWebSearch() domain.Tool {
	return tools.MustGetTool("web_search")
}
