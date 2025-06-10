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
)

// feedDiscoverParamSchema defines parameters for the FeedDiscover tool
var feedDiscoverParamSchema = &sdomain.Schema{
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
	},
	Required: []string{"url"},
}

func init() {
	tools.MustRegisterTool("feed_discover", FeedDiscover(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_discover",
			Category:    "feed",
			Tags:        []string{"feed", "discover", "rss", "atom", "json", "auto-discovery"},
			Description: "Automatically discover feed URLs from web pages",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic discovery",
					Description: "Discover feeds from a blog homepage",
					Code:        `FeedDiscover().Execute(ctx, FeedDiscoverParams{URL: "https://blog.example.com"})`,
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
	})
}

// FeedDiscoverParams contains parameters for the FeedDiscover tool
type FeedDiscoverParams struct {
	URL             string `json:"url" jsonschema:"required,description=URL of the web page to discover feeds from"`
	FollowRedirects *bool  `json:"follow_redirects,omitempty" jsonschema:"description=Whether to follow HTTP redirects (default: true)"`
	Timeout         int    `json:"timeout,omitempty" jsonschema:"description=Timeout for the HTTP request in seconds (default: 30)"`
	MaxSize         int64  `json:"max_size,omitempty" jsonschema:"description=Maximum size of the response body in bytes (default: 10MB)"`
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

// FeedDiscover creates a new FeedDiscover tool
func FeedDiscover() domain.Tool {
	return atools.NewTool(
		"feed_discover",
		"Automatically discover feed URLs from a web page by analyzing HTML link tags and common feed paths",
		func(ctx *domain.ToolContext, params FeedDiscoverParams) (*FeedDiscoverResult, error) {
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

			result, err := feedDiscoverExecute(ctx.Context, params)
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
		},
		feedDiscoverParamSchema,
	)
}

func feedDiscoverExecute(ctx context.Context, params FeedDiscoverParams) (*FeedDiscoverResult, error) {
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

	req.Header.Set("User-Agent", "FeedDiscover/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

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
			// Verify the feed exists
			if verifyFeedExists(ctx, feed.URL, client) {
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

// verifyFeedExists checks if a feed URL actually exists
func verifyFeedExists(ctx context.Context, feedURL string, client *http.Client) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", feedURL, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", "FeedDiscover/1.0")

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
