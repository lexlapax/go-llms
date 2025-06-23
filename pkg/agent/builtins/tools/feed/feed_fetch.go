// ABOUTME: Feed fetching and parsing tool supporting RSS, Atom, JSON Feed formats
// ABOUTME: Built-in tool that provides feed retrieval and parsing capabilities for agents

package feed

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/util/auth"
)

// FeedFetchParams defines parameters for the FeedFetch tool
type FeedFetchParams struct {
	URL        string                 `json:"url"`
	Timeout    int                    `json:"timeout,omitempty"`     // Timeout in seconds, default 30
	MaxItems   int                    `json:"max_items,omitempty"`   // Maximum items to return, 0 = all
	UserAgent  string                 `json:"user_agent,omitempty"`  // Custom user agent
	IfModified string                 `json:"if_modified,omitempty"` // If-Modified-Since header value
	ETag       string                 `json:"etag,omitempty"`        // ETag for conditional requests
	Auth       map[string]interface{} `json:"auth,omitempty"`        // Authentication configuration
	Headers    map[string]string      `json:"headers,omitempty"`     // Additional HTTP headers
}

// FeedFetchResult defines the result of the FeedFetch tool
type FeedFetchResult struct {
	Feed        UnifiedFeed       `json:"feed"`
	Status      int               `json:"status"`
	Headers     map[string]string `json:"headers,omitempty"`
	Format      string            `json:"format"` // RSS2, Atom, JSONFeed, RDF
	NotModified bool              `json:"not_modified,omitempty"`
}

// UnifiedFeed represents a feed in a unified format
type UnifiedFeed struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Link        string     `json:"link,omitempty"`
	Updated     *time.Time `json:"updated,omitempty"`
	Published   *time.Time `json:"published,omitempty"`
	Language    string     `json:"language,omitempty"`
	Copyright   string     `json:"copyright,omitempty"`
	Author      *Author    `json:"author,omitempty"`
	Image       *Image     `json:"image,omitempty"`
	Items       []FeedItem `json:"items"`
}

// FeedItem represents an item in the feed
type FeedItem struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Content     string      `json:"content,omitempty"`
	Link        string      `json:"link,omitempty"`
	Published   *time.Time  `json:"published,omitempty"`
	Updated     *time.Time  `json:"updated,omitempty"`
	Author      *Author     `json:"author,omitempty"`
	Categories  []string    `json:"categories,omitempty"`
	Enclosures  []Enclosure `json:"enclosures,omitempty"`
}

// Author represents author information
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// Image represents feed image
type Image struct {
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
	Link  string `json:"link,omitempty"`
}

// Enclosure represents media attachments
type Enclosure struct {
	URL    string `json:"url"`
	Type   string `json:"type,omitempty"`
	Length int64  `json:"length,omitempty"`
}

// Feed format structures for parsing
type rss2Feed struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel struct {
		Title       string    `xml:"title"`
		Description string    `xml:"description"`
		Link        string    `xml:"link"`
		Language    string    `xml:"language"`
		Copyright   string    `xml:"copyright"`
		PubDate     string    `xml:"pubDate"`
		LastBuild   string    `xml:"lastBuildDate"`
		Image       *rssImage `xml:"image"`
		Items       []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssImage struct {
	URL   string `xml:"url"`
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

type rssItem struct {
	Title       string         `xml:"title"`
	Description string         `xml:"description"`
	Link        string         `xml:"link"`
	GUID        string         `xml:"guid"`
	PubDate     string         `xml:"pubDate"`
	Author      string         `xml:"author"`
	Categories  []string       `xml:"category"`
	Enclosures  []rssEnclosure `xml:"enclosure"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length int64  `xml:"length,attr"`
}

type atomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Links   []atomLink  `xml:"link"`
	Author  *atomAuthor `xml:"author"`
	Entries []atomEntry `xml:"entry"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr"`
}

type atomAuthor struct {
	Name  string `xml:"name"`
	Email string `xml:"email"`
	URI   string `xml:"uri"`
}

type atomEntry struct {
	ID         string         `xml:"id"`
	Title      string         `xml:"title"`
	Updated    string         `xml:"updated"`
	Published  string         `xml:"published"`
	Content    atomContent    `xml:"content"`
	Summary    string         `xml:"summary"`
	Links      []atomLink     `xml:"link"`
	Author     *atomAuthor    `xml:"author"`
	Categories []atomCategory `xml:"category"`
}

type atomContent struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type atomCategory struct {
	Term string `xml:"term,attr"`
}

type jsonFeed struct {
	Version     string          `json:"version"`
	Title       string          `json:"title"`
	HomePageURL string          `json:"home_page_url"`
	FeedURL     string          `json:"feed_url"`
	Description string          `json:"description"`
	Icon        string          `json:"icon"`
	Favicon     string          `json:"favicon"`
	Author      *jsonFeedAuthor `json:"author"`
	Items       []jsonFeedItem  `json:"items"`
}

type jsonFeedAuthor struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Avatar string `json:"avatar"`
}

type jsonFeedItem struct {
	ID            string           `json:"id"`
	URL           string           `json:"url"`
	Title         string           `json:"title"`
	ContentHTML   string           `json:"content_html"`
	ContentText   string           `json:"content_text"`
	Summary       string           `json:"summary"`
	DatePublished string           `json:"date_published"`
	DateModified  string           `json:"date_modified"`
	Author        *jsonFeedAuthor  `json:"author"`
	Tags          []string         `json:"tags"`
	Attachments   []jsonAttachment `json:"attachments"`
}

type jsonAttachment struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size_in_bytes"`
}

// feedFetchExecute is the execution function for feed_fetch
func feedFetchExecute(ctx *domain.ToolContext, params FeedFetchParams) (*FeedFetchResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
			ToolName:   "feed_fetch",
			Parameters: params,
			RequestID:  ctx.RunID,
		})
	}

	// Check state for default timeout if not provided
	timeout := 30 * time.Second
	if params.Timeout > 0 {
		timeout = time.Duration(params.Timeout) * time.Second
	} else if ctx.State != nil {
		if val, ok := ctx.State.Get("feed_fetch_default_timeout"); ok {
			if t, ok := val.(int); ok && t > 0 {
				timeout = time.Duration(t) * time.Second
			}
		}
	}

	// Check state for default user agent if not provided
	userAgent := "go-llms-feed/2.0"
	if params.UserAgent != "" {
		userAgent = params.UserAgent
	} else if ctx.State != nil {
		if val, ok := ctx.State.Get("feed_fetch_user_agent"); ok {
			if ua, ok := val.(string); ok && ua != "" {
				userAgent = ua
			}
		}
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

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx.Context, "GET", params.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set user agent header
	req.Header.Set("User-Agent", userAgent)

	// Apply custom headers
	for k, v := range params.Headers {
		req.Header.Set(k, v)
	}

	// Set conditional headers
	if params.IfModified != "" {
		req.Header.Set("If-Modified-Since", params.IfModified)
	}
	if params.ETag != "" {
		req.Header.Set("If-None-Match", params.ETag)
	}

	// Apply authentication
	if params.Auth != nil {
		if err := auth.ApplyAuth(req, params.Auth); err != nil {
			return nil, fmt.Errorf("auth error: %w", err)
		}
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching feed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle 304 Not Modified
	if resp.StatusCode == http.StatusNotModified {
		result := &FeedFetchResult{
			Status:      resp.StatusCode,
			NotModified: true,
			Headers:     extractHeaders(resp.Header),
		}

		// Emit result event
		if ctx.Events != nil {
			ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
				ToolName:  "feed_fetch",
				Result:    result,
				RequestID: ctx.RunID,
			})
		}

		return result, nil
	}

	// Check status code
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}

	// Read response body with size limit (10MB)
	limitReader := io.LimitReader(resp.Body, 10*1024*1024)
	body, err := io.ReadAll(limitReader)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Detect and parse feed format
	feed, format, err := parseFeed(body)
	if err != nil {
		return nil, fmt.Errorf("error parsing feed: %w", err)
	}

	// Apply max items limit
	if params.MaxItems > 0 && len(feed.Items) > params.MaxItems {
		feed.Items = feed.Items[:params.MaxItems]
	}

	result := &FeedFetchResult{
		Feed:    *feed,
		Status:  resp.StatusCode,
		Headers: extractHeaders(resp.Header),
		Format:  format,
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
			ToolName:  "feed_fetch",
			Result:    result,
			RequestID: ctx.RunID,
		})
	}

	return result, nil
}

// FeedFetch creates a tool that retrieves and parses web feeds in RSS, Atom, or JSON Feed formats with comprehensive authentication support.
// The tool automatically detects feed formats, normalizes them into a unified structure, and handles conditional requests using ETags and If-Modified-Since headers.
// It supports various authentication methods including API key, Bearer token, Basic auth, and OAuth2 for accessing protected feeds.
// This is essential for news aggregation, podcast feed parsing, content monitoring, and any application that needs to consume syndicated content.
func FeedFetch() domain.Tool {
	// Define parameter schema
	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"url": {
				Type:        "string",
				Format:      "uri",
				Description: "The feed URL to fetch",
			},
			"timeout": {
				Type:        "number",
				Description: "Request timeout in seconds (default: 30)",
			},
			"max_items": {
				Type:        "number",
				Description: "Maximum number of items to return (0 = all)",
			},
			"user_agent": {
				Type:        "string",
				Description: "Custom User-Agent header",
			},
			"if_modified": {
				Type:        "string",
				Description: "If-Modified-Since header value for conditional requests",
			},
			"etag": {
				Type:        "string",
				Description: "ETag value for conditional requests",
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
			"feed": {
				Type:        "object",
				Description: "The parsed feed in unified format",
				Properties: map[string]sdomain.Property{
					"title": {
						Type:        "string",
						Description: "Feed title",
					},
					"description": {
						Type:        "string",
						Description: "Feed description",
					},
					"link": {
						Type:        "string",
						Description: "Feed website link",
					},
					"updated": {
						Type:        "string",
						Format:      "date-time",
						Description: "Last update time",
					},
					"published": {
						Type:        "string",
						Format:      "date-time",
						Description: "Publication time",
					},
					"language": {
						Type:        "string",
						Description: "Feed language",
					},
					"copyright": {
						Type:        "string",
						Description: "Copyright information",
					},
					"author": {
						Type:        "object",
						Description: "Feed author",
					},
					"image": {
						Type:        "object",
						Description: "Feed image/logo",
					},
					"items": {
						Type:        "array",
						Description: "Feed items/entries",
						Items: &sdomain.Property{
							Type: "object",
						},
					},
				},
			},
			"status": {
				Type:        "integer",
				Description: "HTTP status code",
			},
			"headers": {
				Type:        "object",
				Description: "Relevant HTTP response headers",
			},
			"format": {
				Type:        "string",
				Description: "Detected feed format: RSS2, Atom, JSONFeed",
			},
			"not_modified": {
				Type:        "boolean",
				Description: "True if feed hasn't changed (304 response)",
			},
		},
		Required: []string{"status"},
	}

	builder := atools.NewToolBuilder("feed_fetch", "Fetch and parse RSS, Atom, or JSON feeds").
		WithFunction(feedFetchExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The feed_fetch tool retrieves and parses web feeds in various formats:

Feed Formats Supported:
1. RSS 2.0:
   - Most common feed format
   - XML-based with channel and item elements
   - Supports enclosures for podcasts/media

2. Atom:
   - IETF standard feed format
   - More structured than RSS
   - Better date/time handling

3. JSON Feed:
   - Modern JSON-based format
   - Easier to parse than XML
   - Native support for content types

Unified Output Format:
- All feeds converted to consistent structure
- Normalized field names across formats
- Proper date/time parsing
- Author information extraction
- Media enclosure support

Authentication Support:
- Automatic detection from state (api_key, bearer_token, etc.)
- Manual auth configuration for protected feeds
- Support for API key, Bearer, Basic, OAuth2, and custom auth
- Works with subscription-based feeds

Conditional Requests:
- ETag support for bandwidth efficiency
- If-Modified-Since header support
- Returns 304 Not Modified when unchanged
- Preserves caching headers in response

State Integration:
- feed_fetch_default_timeout: Default timeout in seconds
- feed_fetch_user_agent: Default User-Agent string
- Authentication auto-detected from state keys

Common Use Cases:
- News aggregation and monitoring
- Podcast feed parsing
- Blog post syndication
- Content change detection
- Feed validation and testing`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Basic RSS fetch",
				Description: "Fetch a public RSS feed",
				Scenario:    "When you need to read articles from a blog",
				Input: map[string]interface{}{
					"url": "https://example.com/rss",
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"title":       "Example Blog",
						"description": "A blog about examples",
						"link":        "https://example.com",
						"items": []map[string]interface{}{
							{
								"id":          "https://example.com/post1",
								"title":       "First Post",
								"description": "This is the first post",
								"link":        "https://example.com/post1",
								"published":   "2024-03-15T10:00:00Z",
							},
						},
					},
					"status": 200,
					"format": "RSS2",
				},
				Explanation: "Successfully fetched and parsed RSS 2.0 feed",
			},
			{
				Name:        "Fetch with authentication",
				Description: "Fetch a protected feed",
				Scenario:    "When accessing a subscription-based feed",
				Input: map[string]interface{}{
					"url": "https://premium.example.com/feed",
					"auth": map[string]interface{}{
						"type":  "bearer",
						"token": "your-access-token",
					},
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"title": "Premium Content Feed",
						"items": []map[string]interface{}{
							{
								"id":      "premium-1",
								"title":   "Exclusive Article",
								"content": "Full premium content here...",
							},
						},
					},
					"status": 200,
					"format": "Atom",
				},
				Explanation: "Used bearer token to access protected Atom feed",
			},
			{
				Name:        "Conditional fetch with ETag",
				Description: "Check if feed has changed",
				Scenario:    "When monitoring feeds for updates efficiently",
				Input: map[string]interface{}{
					"url":  "https://news.example.com/feed",
					"etag": `W/"123456789"`,
				},
				Output: map[string]interface{}{
					"status":       304,
					"not_modified": true,
					"headers": map[string]string{
						"ETag":          `W/"123456789"`,
						"Cache-Control": "max-age=300",
					},
				},
				Explanation: "Feed hasn't changed since last fetch (304 Not Modified)",
			},
			{
				Name:        "Fetch with item limit",
				Description: "Get only recent items",
				Scenario:    "When you only need the latest updates",
				Input: map[string]interface{}{
					"url":       "https://blog.example.com/feed",
					"max_items": 5,
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"title": "Tech Blog",
						"items": []map[string]interface{}{
							{"id": "1", "title": "Latest Post"},
							{"id": "2", "title": "Yesterday's Post"},
							{"id": "3", "title": "Previous Post"},
							{"id": "4", "title": "Older Post"},
							{"id": "5", "title": "Fifth Post"},
						},
					},
					"status": 200,
					"format": "RSS2",
				},
				Explanation: "Limited results to 5 most recent items",
			},
			{
				Name:        "JSON Feed fetch",
				Description: "Fetch a modern JSON feed",
				Scenario:    "When working with JSON-based feeds",
				Input: map[string]interface{}{
					"url": "https://modern.example.com/feed.json",
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"title":       "Modern Blog",
						"description": "A blog using JSON Feed",
						"link":        "https://modern.example.com",
						"author": map[string]interface{}{
							"name": "John Doe",
							"url":  "https://modern.example.com/about",
						},
						"items": []map[string]interface{}{
							{
								"id":      "2024-03-15",
								"title":   "JSON Feeds are Great",
								"content": "Here's why JSON feeds are awesome...",
								"tags":    []string{"json", "feeds", "web"},
							},
						},
					},
					"status": 200,
					"format": "JSONFeed",
				},
				Explanation: "Parsed JSON Feed 1.1 format",
			},
			{
				Name:        "Podcast feed with enclosures",
				Description: "Fetch podcast RSS with media files",
				Scenario:    "When parsing podcast feeds",
				Input: map[string]interface{}{
					"url": "https://podcast.example.com/rss",
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"title": "Tech Podcast",
						"items": []map[string]interface{}{
							{
								"id":          "episode-42",
								"title":       "Episode 42: Feed Processing",
								"description": "Discussion about feed formats",
								"enclosures": []map[string]interface{}{
									{
										"url":    "https://podcast.example.com/episodes/42.mp3",
										"type":   "audio/mpeg",
										"length": 25000000,
									},
								},
							},
						},
					},
					"status": 200,
					"format": "RSS2",
				},
				Explanation: "Parsed podcast feed with audio enclosures",
			},
			{
				Name:        "Custom headers",
				Description: "Fetch with custom HTTP headers",
				Scenario:    "When the feed server requires specific headers",
				Input: map[string]interface{}{
					"url": "https://api.example.com/feed",
					"headers": map[string]string{
						"X-API-Version": "2.0",
						"Accept":        "application/rss+xml",
					},
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"title": "API Feed",
						"items": []map[string]interface{}{},
					},
					"status": 200,
					"format": "RSS2",
				},
				Explanation: "Custom headers included in feed request",
			},
			{
				Name:        "Timeout handling",
				Description: "Set custom timeout for slow feeds",
				Scenario:    "When fetching from slow servers",
				Input: map[string]interface{}{
					"url":     "https://slow.example.com/feed",
					"timeout": 60,
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"title": "Slow Feed",
						"items": []map[string]interface{}{},
					},
					"status": 200,
					"format": "Atom",
				},
				Explanation: "Extended timeout allowed slow feed to complete",
			},
		}).
		WithConstraints([]string{
			"URL must be a valid HTTP or HTTPS URL",
			"Timeout must be positive (default: 30 seconds)",
			"MaxItems of 0 returns all items",
			"Response body limited to 10MB",
			"Supports RSS 2.0, Atom 1.0, and JSON Feed 1.x",
			"Date formats are normalized to RFC3339",
			"Authentication is applied to feed requests",
			"Conditional requests save bandwidth",
		}).
		WithErrorGuidance(map[string]string{
			"error creating request": "Check if the URL is properly formatted",
			"error fetching feed":    "Verify network connectivity and URL accessibility",
			"HTTP error 401":         "Authentication required - provide auth configuration",
			"HTTP error 403":         "Access forbidden - check credentials or permissions",
			"HTTP error 404":         "Feed not found at the specified URL",
			"error parsing feed":     "The content is not a valid RSS, Atom, or JSON feed",
			"unknown feed format":    "Unable to detect feed format - check if URL points to a feed",
			"auth error":             "Authentication configuration is invalid",
			"timeout":                "Increase timeout parameter for slow servers",
		}).
		WithCategory("feed").
		WithTags([]string{"feed", "rss", "atom", "json", "syndication", "news", "podcast", "authentication"}).
		WithVersion("2.0.0").
		WithBehavior(false, false, false, "medium") // Non-deterministic due to network

	return builder.Build()
}

// parseFeed detects and parses the feed format
func parseFeed(data []byte) (*UnifiedFeed, string, error) {
	// Try JSON Feed first (starts with {)
	if len(data) > 0 && data[0] == '{' {
		var jf jsonFeed
		if err := json.Unmarshal(data, &jf); err == nil && jf.Version != "" {
			return convertJSONFeed(&jf), "JSONFeed", nil
		}
	}

	// Try XML formats
	// Try Atom
	var atom atomFeed
	if err := xml.Unmarshal(data, &atom); err == nil && atom.XMLName.Local == "feed" {
		return convertAtomFeed(&atom), "Atom", nil
	}

	// Try RSS 2.0
	var rss rss2Feed
	if err := xml.Unmarshal(data, &rss); err == nil && rss.XMLName.Local == "rss" {
		return convertRSSFeed(&rss), "RSS2", nil
	}

	// Try RDF/RSS 1.0 (simplified - treat as RSS)
	// For now, we'll treat RSS 1.0 similar to RSS 2.0

	return nil, "", fmt.Errorf("unknown feed format")
}

// convertRSSFeed converts RSS feed to unified format
func convertRSSFeed(rss *rss2Feed) *UnifiedFeed {
	feed := &UnifiedFeed{
		Title:       rss.Channel.Title,
		Description: rss.Channel.Description,
		Link:        rss.Channel.Link,
		Language:    rss.Channel.Language,
		Copyright:   rss.Channel.Copyright,
		Items:       make([]FeedItem, 0, len(rss.Channel.Items)),
	}

	// Parse dates
	if rss.Channel.PubDate != "" {
		if t, err := parseRSSDate(rss.Channel.PubDate); err == nil {
			feed.Published = &t
		}
	}
	if rss.Channel.LastBuild != "" {
		if t, err := parseRSSDate(rss.Channel.LastBuild); err == nil {
			feed.Updated = &t
		}
	}

	// Convert image
	if rss.Channel.Image != nil {
		feed.Image = &Image{
			URL:   rss.Channel.Image.URL,
			Title: rss.Channel.Image.Title,
			Link:  rss.Channel.Image.Link,
		}
	}

	// Convert items
	for _, item := range rss.Channel.Items {
		feedItem := FeedItem{
			ID:          item.GUID,
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Categories:  item.Categories,
		}

		if feedItem.ID == "" {
			feedItem.ID = item.Link
		}

		// Parse date
		if item.PubDate != "" {
			if t, err := parseRSSDate(item.PubDate); err == nil {
				feedItem.Published = &t
			}
		}

		// Parse author
		if item.Author != "" {
			feedItem.Author = &Author{Name: item.Author}
		}

		// Convert enclosures
		for _, enc := range item.Enclosures {
			feedItem.Enclosures = append(feedItem.Enclosures, Enclosure(enc))
		}

		feed.Items = append(feed.Items, feedItem)
	}

	return feed
}

// convertAtomFeed converts Atom feed to unified format
func convertAtomFeed(atom *atomFeed) *UnifiedFeed {
	feed := &UnifiedFeed{
		Title: atom.Title,
		Items: make([]FeedItem, 0, len(atom.Entries)),
	}

	// Find alternate link
	for _, link := range atom.Links {
		if link.Rel == "alternate" || link.Rel == "" {
			feed.Link = link.Href
			break
		}
	}

	// Parse updated date
	if atom.Updated != "" {
		if t, err := time.Parse(time.RFC3339, atom.Updated); err == nil {
			feed.Updated = &t
		}
	}

	// Convert author
	if atom.Author != nil {
		feed.Author = &Author{
			Name:  atom.Author.Name,
			Email: atom.Author.Email,
			URL:   atom.Author.URI,
		}
	}

	// Convert entries
	for _, entry := range atom.Entries {
		feedItem := FeedItem{
			ID:    entry.ID,
			Title: entry.Title,
		}

		// Get content or summary
		if entry.Content.Value != "" {
			feedItem.Content = entry.Content.Value
		}
		if entry.Summary != "" {
			feedItem.Description = entry.Summary
		}

		// Find alternate link
		for _, link := range entry.Links {
			if link.Rel == "alternate" || link.Rel == "" {
				feedItem.Link = link.Href
				break
			}
		}

		// Parse dates
		if entry.Published != "" {
			if t, err := time.Parse(time.RFC3339, entry.Published); err == nil {
				feedItem.Published = &t
			}
		}
		if entry.Updated != "" {
			if t, err := time.Parse(time.RFC3339, entry.Updated); err == nil {
				feedItem.Updated = &t
			}
		}

		// Convert author
		if entry.Author != nil {
			feedItem.Author = &Author{
				Name:  entry.Author.Name,
				Email: entry.Author.Email,
				URL:   entry.Author.URI,
			}
		}

		// Convert categories
		for _, cat := range entry.Categories {
			feedItem.Categories = append(feedItem.Categories, cat.Term)
		}

		feed.Items = append(feed.Items, feedItem)
	}

	return feed
}

// convertJSONFeed converts JSON Feed to unified format
func convertJSONFeed(jf *jsonFeed) *UnifiedFeed {
	feed := &UnifiedFeed{
		Title:       jf.Title,
		Description: jf.Description,
		Link:        jf.HomePageURL,
		Items:       make([]FeedItem, 0, len(jf.Items)),
	}

	// Convert author
	if jf.Author != nil {
		feed.Author = &Author{
			Name: jf.Author.Name,
			URL:  jf.Author.URL,
		}
	}

	// Convert icon to image
	if jf.Icon != "" {
		feed.Image = &Image{
			URL: jf.Icon,
		}
	}

	// Convert items
	for _, item := range jf.Items {
		feedItem := FeedItem{
			ID:         item.ID,
			Title:      item.Title,
			Link:       item.URL,
			Categories: item.Tags,
		}

		// Get content
		if item.ContentHTML != "" {
			feedItem.Content = item.ContentHTML
		} else if item.ContentText != "" {
			feedItem.Content = item.ContentText
		}

		if item.Summary != "" {
			feedItem.Description = item.Summary
		}

		// Parse dates
		if item.DatePublished != "" {
			if t, err := time.Parse(time.RFC3339, item.DatePublished); err == nil {
				feedItem.Published = &t
			}
		}
		if item.DateModified != "" {
			if t, err := time.Parse(time.RFC3339, item.DateModified); err == nil {
				feedItem.Updated = &t
			}
		}

		// Convert author
		if item.Author != nil {
			feedItem.Author = &Author{
				Name: item.Author.Name,
				URL:  item.Author.URL,
			}
		}

		// Convert attachments
		for _, att := range item.Attachments {
			feedItem.Enclosures = append(feedItem.Enclosures, Enclosure{
				URL:    att.URL,
				Type:   att.MimeType,
				Length: att.Size,
			})
		}

		feed.Items = append(feed.Items, feedItem)
	}

	return feed
}

// parseRSSDate parses common RSS date formats
func parseRSSDate(dateStr string) (time.Time, error) {
	// Try common RSS date formats
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		"Mon, 02 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z07:00",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// extractHeaders extracts relevant headers from HTTP response
func extractHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	relevantHeaders := []string{
		"Last-Modified",
		"ETag",
		"Content-Type",
		"Content-Length",
		"Cache-Control",
		"Expires",
	}

	for _, key := range relevantHeaders {
		if value := headers.Get(key); value != "" {
			result[key] = value
		}
	}

	return result
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("feed_fetch", FeedFetch(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_fetch",
			Category:    "feed",
			Tags:        []string{"feed", "rss", "atom", "json", "syndication", "news", "podcast"},
			Description: "Fetches and parses feeds in RSS, Atom, or JSON Feed format with authentication support",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic RSS fetch",
					Description: "Fetch an RSS feed",
					Code:        `FeedFetch().Execute(ctx, FeedFetchParams{URL: "https://example.com/rss"})`,
				},
				{
					Name:        "Fetch with auth",
					Description: "Fetch protected feed",
					Code:        `FeedFetch().Execute(ctx, FeedFetchParams{URL: "https://private.example.com/feed", Auth: map[string]interface{}{"type": "bearer", "token": "xyz"}})`,
				},
				{
					Name:        "Fetch with limit",
					Description: "Fetch only recent items",
					Code:        `FeedFetch().Execute(ctx, FeedFetchParams{URL: "https://blog.example.com/feed", MaxItems: 10})`,
				},
				{
					Name:        "Conditional fetch",
					Description: "Fetch only if modified",
					Code:        `FeedFetch().Execute(ctx, FeedFetchParams{URL: "https://news.example.com/atom", ETag: "W/\"123456\""})`,
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
		UsageInstructions: `The feed_fetch tool retrieves and parses web feeds:
- Supports RSS 2.0, Atom, and JSON Feed formats
- Unified output format across all feed types
- Authentication support for protected feeds
- Conditional requests with ETag/If-Modified-Since
- Configurable timeouts and item limits
- Automatic format detection

Use for news aggregation, podcast feeds, and content monitoring.`,
		Constraints: []string{
			"URL must be valid HTTP/HTTPS",
			"Response body limited to 10MB",
			"Timeout defaults to 30 seconds",
			"Dates normalized to RFC3339",
			"Auth applied to all requests",
		},
		ErrorGuidance: map[string]string{
			"error fetching feed": "Check URL and network",
			"HTTP error 401":      "Add authentication",
			"error parsing feed":  "Verify feed format",
			"unknown feed format": "Not a valid feed",
			"auth error":          "Check auth config",
		},
		IsDeterministic:      false, // Network operation
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "medium",
	})
}
