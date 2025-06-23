// ABOUTME: FeedDiscover tool for automatically discovering feed URLs from web pages
// ABOUTME: Searches HTML content for RSS, Atom, and JSON feed links in <link> tags and common patterns

package feed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/util/auth"
)

// FeedDiscoverParams contains parameters for the FeedDiscover tool
type FeedDiscoverParams struct {
	URL             string                 `json:"url" jsonschema:"required,description=URL of the web page to discover feeds from"`
	FollowRedirects *bool                  `json:"follow_redirects,omitempty" jsonschema:"description=Whether to follow HTTP redirects (default: true)"`
	Timeout         int                    `json:"timeout,omitempty" jsonschema:"description=Timeout for the HTTP request in seconds (default: 30)"`
	MaxSize         int64                  `json:"max_size,omitempty" jsonschema:"description=Maximum size of the response body in bytes (default: 10MB)"`
	Auth            map[string]interface{} `json:"auth,omitempty" jsonschema:"description=Authentication configuration"`
	Headers         map[string]string      `json:"headers,omitempty" jsonschema:"description=Additional HTTP headers"`
}

// FeedDiscoverResult contains the result of feed discovery
type FeedDiscoverResult struct {
	Feeds []DiscoveredFeed `json:"feeds"`
	Error string           `json:"error,omitempty"`
}

// DiscoveredFeed represents a discovered feed
type DiscoveredFeed struct {
	URL    string `json:"url"`
	Title  string `json:"title,omitempty"`
	Type   string `json:"type"`
	Source string `json:"source"` // "link_tag", "auto_discovery", "common_path"
}

// feedDiscoverExecute is the execution function for feed_discover
func feedDiscoverExecute(ctx *domain.ToolContext, params FeedDiscoverParams) (*FeedDiscoverResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
			ToolName:   "feed_discover",
			Parameters: params,
			RequestID:  ctx.RunID,
		})
	}

	// Check state for default values
	if params.Timeout <= 0 {
		timeout := 30 // default 30 seconds
		if ctx.State != nil {
			if val, ok := ctx.State.Get("feed_discover_timeout"); ok {
				if t, ok := val.(int); ok && t > 0 {
					timeout = t
				}
			}
		}
		params.Timeout = timeout
	}

	if params.MaxSize <= 0 {
		maxSize := int64(10 * 1024 * 1024) // default 10MB
		if ctx.State != nil {
			if val, ok := ctx.State.Get("feed_discover_max_size"); ok {
				if size, ok := val.(int64); ok && size > 0 {
					maxSize = size
				} else if size, ok := val.(int); ok && size > 0 {
					maxSize = int64(size)
				}
			}
		}
		params.MaxSize = maxSize
	}

	// Default to following redirects if not specified
	if params.FollowRedirects == nil {
		followRedirects := true
		if ctx.State != nil {
			if val, ok := ctx.State.Get("feed_discover_follow_redirects"); ok {
				if follow, ok := val.(bool); ok {
					followRedirects = follow
				}
			}
		}
		params.FollowRedirects = &followRedirects
	}

	// Auto-detect auth if not provided
	if params.Auth == nil && ctx.State != nil {
		if authConfig := auth.DetectAuthFromState(ctx.State, params.URL, nil); authConfig != nil {
			params.Auth = map[string]interface{}{
				"type": authConfig.Type,
			}
			for k, v := range authConfig.Data {
				params.Auth[k] = v
			}
		}
	}

	result, err := feedDiscoverExecuteCore(ctx.Context, params)
	if err != nil {
		return nil, err
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
			ToolName:  "feed_discover",
			Result:    result,
			RequestID: ctx.RunID,
		})
	}

	return result, nil
}

// FeedDiscover creates a tool that automatically finds RSS, Atom, and JSON feed URLs from web pages using multiple discovery methods.
// The tool searches HTML link tags for feed declarations, checks common feed paths like /feed and /rss, and supports feed auto-discovery standards.
// It includes authentication support for protected sites and validates discovered feeds with HEAD requests to minimize bandwidth usage.
// This is invaluable for finding all available feeds on blogs, discovering podcast feeds, and aggregating feeds from multiple sites.
func FeedDiscover() domain.Tool {
	// Define parameter schema
	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"url": {
				Type:        "string",
				Format:      "uri",
				Description: "URL of the web page to discover feeds from",
			},
			"follow_redirects": {
				Type:        "boolean",
				Description: "Whether to follow HTTP redirects (default: true)",
			},
			"timeout": {
				Type:        "number",
				Description: "Request timeout in seconds (default: 30)",
			},
			"max_size": {
				Type:        "number",
				Description: "Maximum size of the response body in bytes (default: 10MB)",
			},
			"auth": {
				Type:        "object",
				Description: "Authentication configuration",
				Properties: map[string]sdomain.Property{
					"type": {
						Type:        "string",
						Description: "Authentication type: api_key, bearer, basic, oauth2, custom",
						Enum:        []string{"api_key", "bearer", "basic", "oauth2", "custom"},
					},
				},
			},
			"headers": {
				Type:        "object",
				Description: "Additional HTTP headers to send with the request",
			},
		},
		Required: []string{"url"},
	}

	// Define output schema
	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"feeds": {
				Type:        "array",
				Description: "List of discovered feeds",
				Items: &sdomain.Property{
					Type: "object",
					Properties: map[string]sdomain.Property{
						"url": {
							Type:        "string",
							Description: "Feed URL",
						},
						"title": {
							Type:        "string",
							Description: "Feed title (if found)",
						},
						"type": {
							Type:        "string",
							Description: "Feed type: rss, atom, or json",
						},
						"source": {
							Type:        "string",
							Description: "Discovery source: link_tag, auto_discovery, or common_path",
						},
					},
				},
			},
			"error": {
				Type:        "string",
				Description: "Error message if discovery failed",
			},
		},
	}

	builder := atools.NewToolBuilder("feed_discover", "Automatically discover feed URLs from web pages").
		WithFunction(feedDiscoverExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The feed_discover tool automatically finds RSS, Atom, and JSON feed URLs from web pages:

Discovery Methods:
1. HTML Link Tags:
   - Searches for <link rel="alternate"> tags
   - Detects type attributes like application/rss+xml, application/atom+xml
   - Extracts feed titles from link tags
   - Resolves relative URLs to absolute

2. Common Feed Paths:
   - Checks standard feed locations like /feed, /rss, /atom.xml
   - Verifies feed existence with HEAD requests
   - Validates content types

3. Auto-Discovery:
   - Follows feed auto-discovery standards
   - Supports RSS 2.0, Atom 1.0, and JSON Feed
   - Handles multiple feeds per page

Authentication Support:
- Automatic detection from state (api_key, bearer_token, etc.)
- Manual auth configuration for protected sites
- Support for API key, Bearer, Basic, OAuth2, and custom auth
- Auth applied to both discovery and verification requests

State Integration:
- feed_discover_timeout: Default timeout in seconds
- feed_discover_max_size: Default max response size
- feed_discover_follow_redirects: Default redirect behavior
- Authentication auto-detected from state keys

Common Use Cases:
- Find all feeds on a blog or news site
- Discover podcast feeds
- Find JSON feeds for modern applications
- Aggregate feeds from multiple sites
- Verify feed availability before subscription`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Basic feed discovery",
				Description: "Discover feeds from a blog homepage",
				Scenario:    "When you want to find all available feeds on a website",
				Input: map[string]interface{}{
					"url": "https://blog.example.com",
				},
				Output: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"url":    "https://blog.example.com/feed",
							"title":  "Example Blog RSS Feed",
							"type":   "rss",
							"source": "link_tag",
						},
						{
							"url":    "https://blog.example.com/atom.xml",
							"title":  "Example Blog Atom Feed",
							"type":   "atom",
							"source": "link_tag",
						},
					},
				},
				Explanation: "Found RSS and Atom feeds declared in the HTML head section",
			},
			{
				Name:        "Discovery with authentication",
				Description: "Discover feeds from a protected site",
				Scenario:    "When the site requires authentication to access",
				Input: map[string]interface{}{
					"url": "https://private.example.com",
					"auth": map[string]interface{}{
						"type":  "bearer",
						"token": "your-access-token",
					},
				},
				Output: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"url":    "https://private.example.com/api/feed.json",
							"type":   "json",
							"source": "link_tag",
						},
					},
				},
				Explanation: "Used bearer token to access protected site and discover JSON feed",
			},
			{
				Name:        "Discovery with timeout",
				Description: "Set custom timeout for slow sites",
				Scenario:    "When discovering feeds from a slow-responding website",
				Input: map[string]interface{}{
					"url":     "https://slow.example.com",
					"timeout": 60,
				},
				Output: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"url":    "https://slow.example.com/rss",
							"type":   "rss",
							"source": "common_path",
						},
					},
				},
				Explanation: "Extended timeout allowed discovery from slow site",
			},
			{
				Name:        "Multiple feed types",
				Description: "Discover various feed formats",
				Scenario:    "When a site offers multiple feed formats",
				Input: map[string]interface{}{
					"url": "https://news.example.com",
				},
				Output: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"url":    "https://news.example.com/rss.xml",
							"title":  "News RSS 2.0",
							"type":   "rss",
							"source": "link_tag",
						},
						{
							"url":    "https://news.example.com/feed.atom",
							"title":  "News Atom 1.0",
							"type":   "atom",
							"source": "link_tag",
						},
						{
							"url":    "https://news.example.com/feed.json",
							"title":  "News JSON Feed",
							"type":   "json",
							"source": "link_tag",
						},
					},
				},
				Explanation: "Site offers multiple feed formats for different client preferences",
			},
			{
				Name:        "No redirects follow",
				Description: "Discover without following redirects",
				Scenario:    "When you need to discover feeds only from the exact URL",
				Input: map[string]interface{}{
					"url":              "https://redirect.example.com",
					"follow_redirects": false,
				},
				Output: map[string]interface{}{
					"feeds": []map[string]interface{}{},
					"error": "HTTP error: 301 Moved Permanently",
				},
				Explanation: "Did not follow redirect as requested",
			},
			{
				Name:        "Common path discovery",
				Description: "Find feeds via common URL patterns",
				Scenario:    "When feeds aren't declared in HTML but exist at standard paths",
				Input: map[string]interface{}{
					"url": "https://simple.example.com",
				},
				Output: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"url":    "https://simple.example.com/feed",
							"type":   "rss",
							"source": "common_path",
						},
						{
							"url":    "https://simple.example.com/rss",
							"type":   "rss",
							"source": "common_path",
						},
					},
				},
				Explanation: "Found feeds at common paths even though not declared in HTML",
			},
			{
				Name:        "Custom headers",
				Description: "Include custom headers in discovery request",
				Scenario:    "When the site requires specific headers",
				Input: map[string]interface{}{
					"url": "https://api.example.com",
					"headers": map[string]string{
						"X-Client-ID": "my-app",
						"Accept":      "text/html,application/xml",
					},
				},
				Output: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"url":    "https://api.example.com/v1/feed",
							"type":   "rss",
							"source": "link_tag",
						},
					},
				},
				Explanation: "Custom headers allowed access to API-based feed discovery",
			},
			{
				Name:        "Size-limited discovery",
				Description: "Limit response size for large pages",
				Scenario:    "When discovering from pages with lots of content",
				Input: map[string]interface{}{
					"url":      "https://huge.example.com",
					"max_size": 1048576, // 1MB
				},
				Output: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"url":    "https://huge.example.com/feed.xml",
							"title":  "Main Feed",
							"type":   "rss",
							"source": "link_tag",
						},
					},
				},
				Explanation: "Limited page download to 1MB while still finding feeds in head section",
			},
		}).
		WithConstraints([]string{
			"URL must be a valid HTTP or HTTPS URL",
			"Timeout must be positive (default: 30 seconds)",
			"MaxSize must be positive (default: 10MB)",
			"Only discovers feeds declared in HTML or at common paths",
			"Verification requests use HEAD method to minimize bandwidth",
			"Relative URLs are resolved against the base URL",
			"Content-Type validation ensures discovered URLs are actually feeds",
			"Authentication is applied to all HTTP requests if configured",
		}).
		WithErrorGuidance(map[string]string{
			"invalid URL":            "Ensure the URL is properly formatted with http:// or https:// scheme",
			"error creating request": "Check if the URL is accessible and properly formatted",
			"error fetching page":    "Verify network connectivity and that the site is accessible",
			"HTTP error":             "Check the status code - may need authentication or the page doesn't exist",
			"error reading response": "The response may be too large or network connection was interrupted",
			"timeout":                "Increase the timeout parameter for slow-responding sites",
			"auth required":          "Add authentication configuration if the site requires login",
		}).
		WithCategory("feed").
		WithTags([]string{"feed", "discover", "rss", "atom", "json", "auto-discovery", "web", "authentication"}).
		WithVersion("2.0.0").
		WithBehavior(false, false, false, "medium") // Non-deterministic due to network

	return builder.Build()
}

func feedDiscoverExecuteCore(ctx context.Context, params FeedDiscoverParams) (*FeedDiscoverResult, error) {
	// Parse base URL
	baseURL, err := url.Parse(params.URL)
	if err != nil {
		return &FeedDiscoverResult{
			Feeds: []DiscoveredFeed{},
			Error: fmt.Sprintf("invalid URL: %v", err),
		}, nil
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: time.Duration(params.Timeout) * time.Second,
	}

	if params.FollowRedirects != nil && !*params.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", params.URL, nil)
	if err != nil {
		return &FeedDiscoverResult{
			Feeds: []DiscoveredFeed{},
			Error: fmt.Sprintf("error creating request: %v", err),
		}, nil
	}

	req.Header.Set("User-Agent", "FeedDiscover/2.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	// Apply custom headers
	for k, v := range params.Headers {
		req.Header.Set(k, v)
	}

	// Apply authentication
	if params.Auth != nil {
		if err := auth.ApplyAuth(req, params.Auth); err != nil {
			return &FeedDiscoverResult{
				Feeds: []DiscoveredFeed{},
				Error: fmt.Sprintf("auth error: %v", err),
			}, nil
		}
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return &FeedDiscoverResult{
			Feeds: []DiscoveredFeed{},
			Error: fmt.Sprintf("error fetching page: %v", err),
		}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return &FeedDiscoverResult{
			Feeds: []DiscoveredFeed{},
			Error: fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status),
		}, nil
	}

	// Read response body with size limit
	limitReader := io.LimitReader(resp.Body, params.MaxSize)
	body, err := io.ReadAll(limitReader)
	if err != nil {
		return &FeedDiscoverResult{
			Feeds: []DiscoveredFeed{},
			Error: fmt.Sprintf("error reading response: %v", err),
		}, nil
	}

	// Discover feeds
	feeds := make([]DiscoveredFeed, 0)
	feedURLs := make(map[string]bool) // To avoid duplicates

	// 1. Parse HTML and look for link tags
	htmlFeeds := discoverFromHTML(body, baseURL)
	for _, feed := range htmlFeeds {
		if !feedURLs[feed.URL] {
			feeds = append(feeds, feed)
			feedURLs[feed.URL] = true
		}
	}

	// 2. Check common feed paths
	commonFeeds := discoverFromCommonPaths(baseURL)
	for _, feed := range commonFeeds {
		if !feedURLs[feed.URL] {
			// Verify the feed exists with auth
			if verifyFeedExistsWithAuth(ctx, feed.URL, client, params.Auth) {
				feeds = append(feeds, feed)
				feedURLs[feed.URL] = true
			}
		}
	}

	return &FeedDiscoverResult{
		Feeds: feeds,
	}, nil
}

// discoverFromHTML parses HTML and looks for feed links using regex
func discoverFromHTML(body []byte, baseURL *url.URL) []DiscoveredFeed {
	feeds := make([]DiscoveredFeed, 0)
	bodyStr := string(body)

	// Regex to find link tags
	linkRegex := regexp.MustCompile(`<link[^>]*>`)
	links := linkRegex.FindAllString(bodyStr, -1)

	for _, link := range links {
		// Extract attributes
		rel := extractAttribute(link, "rel")
		if !strings.Contains(rel, "alternate") {
			continue
		}

		href := extractAttribute(link, "href")
		if href == "" {
			continue
		}

		title := extractAttribute(link, "title")
		feedType := extractAttribute(link, "type")

		var discoveredType string
		switch feedType {
		case "application/rss+xml", "application/rdf+xml":
			discoveredType = "rss"
		case "application/atom+xml":
			discoveredType = "atom"
		case "application/json":
			if strings.Contains(href, "feed") {
				discoveredType = "json"
			}
		case "application/feed+json":
			discoveredType = "json"
		}

		if discoveredType != "" {
			// Resolve relative URLs
			feedURL := resolveURL(baseURL, href)
			feeds = append(feeds, DiscoveredFeed{
				URL:    feedURL,
				Title:  title,
				Type:   discoveredType,
				Source: "link_tag",
			})
		}
	}

	return feeds
}

// extractAttribute extracts an attribute value from an HTML tag
func extractAttribute(tag, attrName string) string {
	// Try to match attribute="value" or attribute='value'
	doubleQuoteRegex := regexp.MustCompile(attrName + `\s*=\s*"([^"]*)"`)
	singleQuoteRegex := regexp.MustCompile(attrName + `\s*=\s*'([^']*)'`)

	if matches := doubleQuoteRegex.FindStringSubmatch(tag); len(matches) > 1 {
		return matches[1]
	}
	if matches := singleQuoteRegex.FindStringSubmatch(tag); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// discoverFromCommonPaths checks common feed paths
func discoverFromCommonPaths(baseURL *url.URL) []DiscoveredFeed {
	commonPaths := []struct {
		path     string
		feedType string
	}{
		{"/feed", "rss"},
		{"/feed.xml", "rss"},
		{"/feeds", "rss"},
		{"/rss", "rss"},
		{"/rss.xml", "rss"},
		{"/rss2.xml", "rss"},
		{"/atom.xml", "atom"},
		{"/feed.atom", "atom"},
		{"/feed.json", "json"},
		{"/index.xml", "rss"},
		{"/blog/feed", "rss"},
		{"/blog/rss", "rss"},
		{"/news/feed", "rss"},
		{"/news/rss", "rss"},
	}

	feeds := make([]DiscoveredFeed, 0)

	for _, cp := range commonPaths {
		feedURL := baseURL.Scheme + "://" + baseURL.Host + cp.path
		feeds = append(feeds, DiscoveredFeed{
			URL:    feedURL,
			Type:   cp.feedType,
			Source: "common_path",
		})
	}

	return feeds
}

// resolveURL resolves a relative URL against a base URL
func resolveURL(base *url.URL, ref string) string {
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(refURL).String()
}

// verifyFeedExistsWithAuth checks if a feed URL actually exists with authentication
func verifyFeedExistsWithAuth(ctx context.Context, feedURL string, client *http.Client, authConfig map[string]interface{}) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", feedURL, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", "FeedDiscover/2.0")

	// Apply authentication if configured
	if authConfig != nil {
		if err := auth.ApplyAuth(req, authConfig); err != nil {
			return false
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	// Check if status is OK and content type suggests a feed
	if resp.StatusCode == http.StatusOK {
		contentType := resp.Header.Get("Content-Type")
		return isFeedContentType(contentType)
	}

	return false
}

// isFeedContentType checks if a content type indicates a feed
func isFeedContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	feedTypes := []string{
		"application/rss+xml",
		"application/atom+xml",
		"application/rdf+xml",
		"application/feed+json",
		"application/json",
		"text/xml",
		"text/rss+xml",
		"text/atom+xml",
	}

	for _, ft := range feedTypes {
		if strings.Contains(contentType, ft) {
			return true
		}
	}

	// Also check for generic XML that might be a feed
	if strings.Contains(contentType, "xml") {
		return true
	}

	return false
}

func init() {
	tools.MustRegisterTool("feed_discover", FeedDiscover(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_discover",
			Category:    "feed",
			Tags:        []string{"feed", "discover", "rss", "atom", "json", "auto-discovery", "web"},
			Description: "Automatically discover feed URLs from web pages with authentication support",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic discovery",
					Description: "Discover feeds from a blog homepage",
					Code:        `FeedDiscover().Execute(ctx, FeedDiscoverParams{URL: "https://blog.example.com"})`,
				},
				{
					Name:        "Discovery with auth",
					Description: "Discover feeds from protected site",
					Code:        `FeedDiscover().Execute(ctx, FeedDiscoverParams{URL: "https://private.example.com", Auth: map[string]interface{}{"type": "bearer", "token": "xyz"}})`,
				},
				{
					Name:        "Discovery with timeout",
					Description: "Set custom timeout for slow sites",
					Code:        `FeedDiscover().Execute(ctx, FeedDiscoverParams{URL: "https://news.example.com", Timeout: 60})`,
				},
				{
					Name:        "No redirects",
					Description: "Discover feeds without following redirects",
					Code:        `FeedDiscover().Execute(ctx, FeedDiscoverParams{URL: "https://site.example.com", FollowRedirects: false})`,
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
		UsageInstructions: `The feed_discover tool finds RSS, Atom, and JSON feed URLs from web pages:
- Searches HTML for <link rel="alternate"> tags
- Checks common feed paths like /feed, /rss, /atom.xml
- Verifies feed existence with HEAD requests
- Supports authentication for protected sites
- Resolves relative URLs to absolute

The tool handles various feed formats and auto-discovery standards.`,
		Constraints: []string{
			"URL must be valid HTTP/HTTPS",
			"Timeout must be positive",
			"MaxSize must be positive",
			"Only discovers declared feeds or common paths",
			"Auth applied to all requests",
		},
		ErrorGuidance: map[string]string{
			"invalid URL": "Check URL format",
			"HTTP error":  "May need authentication",
			"timeout":     "Increase timeout for slow sites",
			"auth error":  "Check auth configuration",
		},
		IsDeterministic:      false, // Network operation
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "medium",
	})
}
