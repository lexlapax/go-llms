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
)

// FeedFetchParams defines parameters for the FeedFetch tool
type FeedFetchParams struct {
	URL        string `json:"url"`
	Timeout    int    `json:"timeout,omitempty"`     // Timeout in seconds, default 30
	MaxItems   int    `json:"max_items,omitempty"`   // Maximum items to return, 0 = all
	UserAgent  string `json:"user_agent,omitempty"`  // Custom user agent
	IfModified string `json:"if_modified,omitempty"` // If-Modified-Since header value
	ETag       string `json:"etag,omitempty"`        // ETag for conditional requests
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

// feedFetchParamSchema defines parameters for the FeedFetch tool
var feedFetchParamSchema = &sdomain.Schema{
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
	},
	Required: []string{"url"},
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("feed_fetch", FeedFetch(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_fetch",
			Category:    "feed",
			Tags:        []string{"feed", "rss", "atom", "json", "syndication", "news"},
			Description: "Fetches and parses feeds in RSS, Atom, or JSON Feed format",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic RSS fetch",
					Description: "Fetch an RSS feed",
					Code:        `FeedFetch().Execute(ctx, FeedFetchParams{URL: "https://example.com/rss"})`,
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
	})
}

// FeedFetch creates a tool for fetching and parsing feeds
func FeedFetch() domain.Tool {
	return atools.NewTool(
		"feed_fetch",
		"Fetches and parses feeds in RSS, Atom, or JSON Feed format",
		func(ctx *domain.ToolContext, params FeedFetchParams) (*FeedFetchResult, error) {
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
			userAgent := "go-llms-feed/1.0"
			if params.UserAgent != "" {
				userAgent = params.UserAgent
			} else if ctx.State != nil {
				if val, ok := ctx.State.Get("feed_fetch_user_agent"); ok {
					if ua, ok := val.(string); ok && ua != "" {
						userAgent = ua
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

			// Set conditional headers
			if params.IfModified != "" {
				req.Header.Set("If-Modified-Since", params.IfModified)
			}
			if params.ETag != "" {
				req.Header.Set("If-None-Match", params.ETag)
			}

			// Execute request
			resp, err := client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("error fetching feed: %w", err)
			}
			defer resp.Body.Close()

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
		},
		feedFetchParamSchema,
	)
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
