// ABOUTME: FeedConvert tool for converting between different feed formats (RSS, Atom, JSON Feed)
// ABOUTME: Supports conversion with optional pretty-printing for human-readable output

package feed

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// FeedConvertParams contains parameters for the FeedConvert tool
type FeedConvertParams struct {
	Feed           UnifiedFeed `json:"feed"`
	TargetType     string      `json:"target_type"`
	Pretty         bool        `json:"pretty,omitempty"`
	IncludeContent bool        `json:"include_content,omitempty"`
}

// FeedConvertResult contains the result of feed conversion
type FeedConvertResult struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	Format      string `json:"format"`
}

// feedConvertExecute is the execution function for feed_convert
func feedConvertExecute(ctx *domain.ToolContext, params FeedConvertParams) (*FeedConvertResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
			ToolName:   "feed_convert",
			Parameters: params,
			RequestID:  ctx.RunID,
		})
	}

	// Check state for default format preferences
	if params.TargetType == "" && ctx.State != nil {
		if val, ok := ctx.State.Get("feed_convert_default_format"); ok {
			if format, ok := val.(string); ok && format != "" {
				params.TargetType = format
			}
		}
	}

	// Check state for default pretty print setting
	if !params.Pretty && ctx.State != nil {
		if val, ok := ctx.State.Get("feed_convert_pretty_print"); ok {
			if pretty, ok := val.(bool); ok {
				params.Pretty = pretty
			}
		}
	}

	result, err := convertFeed(params)
	if err != nil {
		return nil, err
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
			ToolName:  "feed_convert",
			Result:    result,
			RequestID: ctx.RunID,
		})
	}

	return result, nil
}

// FeedConvert creates a tool that transforms feeds between RSS, Atom, and JSON Feed formats while preserving all standard feed elements.
// The tool handles format-specific features like RSS enclosures, Atom content/summary separation, and JSON Feed attachments.
// It supports pretty-printing for human readability and maintains proper date formatting for each target format.
// This is essential for feed format migration, cross-platform compatibility, and meeting specific API format requirements.
func FeedConvert() domain.Tool {
	// Define parameter schema
	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"feed": {
				Type:        "object",
				Description: "Feed data to convert (UnifiedFeed format)",
			},
			"target_type": {
				Type:        "string",
				Description: "Target format: rss, atom, or json",
				Enum:        []string{"rss", "atom", "json"},
			},
			"pretty": {
				Type:        "boolean",
				Description: "Pretty-print the output for readability (default: false)",
			},
			"include_content": {
				Type:        "boolean",
				Description: "Include full content in output, not just summaries (default: true)",
			},
		},
		Required: []string{"feed", "target_type"},
	}

	// Define output schema
	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"content": {
				Type:        "string",
				Description: "The converted feed content in the target format",
			},
			"content_type": {
				Type:        "string",
				Description: "MIME type of the converted content",
			},
			"format": {
				Type:        "string",
				Description: "The format the feed was converted to (rss, atom, json)",
			},
		},
		Required: []string{"content", "content_type", "format"},
	}

	builder := atools.NewToolBuilder("feed_convert", "Convert feeds between RSS, Atom, and JSON Feed formats").
		WithFunction(feedConvertExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The feed_convert tool transforms feeds between different formats:

Supported Conversions:
1. RSS 2.0:
   - Standard RSS format with channel/item structure
   - Wide compatibility with feed readers
   - Best for podcasts and traditional blogs
   - Content included in description field

2. Atom 1.0:
   - IETF standard with better structure
   - Separate content and summary fields
   - Better date/time handling
   - Required unique IDs for entries

3. JSON Feed 1.1:
   - Modern JSON-based format
   - Easy to parse programmatically
   - Native support for attachments
   - Clean separation of content types

Conversion Features:
- Preserves all standard feed elements
- Maps fields appropriately between formats
- Handles enclosures/attachments conversion
- Maintains author information
- Converts dates to appropriate formats
- Optional pretty-printing for readability

State Integration:
- feed_convert_default_format: Default target format
- feed_convert_pretty_print: Default pretty print setting

Common Use Cases:
- Feed format migration
- Cross-platform compatibility
- API format requirements
- Feed validation and testing
- Format modernization`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "RSS to JSON Feed",
				Description: "Convert traditional RSS to modern JSON format",
				Scenario:    "When you need JSON format for modern APIs",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"title":       "Tech Blog",
						"description": "Latest tech news",
						"link":        "https://example.com",
						"items": []map[string]interface{}{
							{
								"id":          "post-1",
								"title":       "New Technology",
								"description": "Brief summary",
								"content":     "<p>Full article content...</p>",
								"link":        "https://example.com/post-1",
								"published":   "2024-03-15T10:00:00Z",
							},
						},
					},
					"target_type": "json",
					"pretty":      true,
				},
				Output: map[string]interface{}{
					"content": `{
  "version": "https://jsonfeed.org/version/1.1",
  "title": "Tech Blog",
  "home_page_url": "https://example.com",
  "description": "Latest tech news",
  "items": [
    {
      "id": "post-1",
      "title": "New Technology",
      "url": "https://example.com/post-1",
      "content_html": "<p>Full article content...</p>",
      "date_published": "2024-03-15T10:00:00Z"
    }
  ]
}`,
					"content_type": "application/feed+json",
					"format":       "json",
				},
				Explanation: "Converted RSS feed to JSON Feed 1.1 format with pretty printing",
			},
			{
				Name:        "Convert to Atom",
				Description: "Convert any feed to Atom format",
				Scenario:    "When you need Atom format for standards compliance",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"title": "News Feed",
						"link":  "https://news.example.com",
						"author": map[string]interface{}{
							"name":  "News Team",
							"email": "news@example.com",
						},
						"items": []map[string]interface{}{
							{
								"id":          "news-1",
								"title":       "Breaking News",
								"description": "Important update",
								"link":        "https://news.example.com/1",
								"published":   "2024-03-15T12:00:00Z",
								"categories":  []string{"urgent", "world"},
							},
						},
					},
					"target_type":     "atom",
					"include_content": false,
				},
				Output: map[string]interface{}{
					"content":      "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<feed xmlns=\"http://www.w3.org/2005/Atom\">...",
					"content_type": "application/atom+xml",
					"format":       "atom",
				},
				Explanation: "Converted to Atom format with summary only (no full content)",
			},
			{
				Name:        "Convert to RSS",
				Description: "Convert modern formats to classic RSS",
				Scenario:    "When you need RSS for legacy compatibility",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"title":     "Podcast Feed",
						"link":      "https://podcast.example.com",
						"language":  "en-us",
						"copyright": "© 2024 Example",
						"items": []map[string]interface{}{
							{
								"id":          "episode-1",
								"title":       "Episode 1: Introduction",
								"description": "Our first episode",
								"link":        "https://podcast.example.com/1",
								"published":   "2024-03-01T09:00:00Z",
								"enclosures": []map[string]interface{}{
									{
										"url":    "https://podcast.example.com/ep1.mp3",
										"type":   "audio/mpeg",
										"length": 15000000,
									},
								},
							},
						},
					},
					"target_type": "rss",
				},
				Output: map[string]interface{}{
					"content":      "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<rss version=\"2.0\">...",
					"content_type": "application/rss+xml",
					"format":       "rss",
				},
				Explanation: "Converted to RSS 2.0 with podcast enclosure support",
			},
			{
				Name:        "Pretty print JSON",
				Description: "Convert with human-readable formatting",
				Scenario:    "When you need formatted output for debugging or display",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"title": "Simple Feed",
						"items": []map[string]interface{}{
							{"id": "1", "title": "Item 1"},
						},
					},
					"target_type": "json",
					"pretty":      true,
				},
				Output: map[string]interface{}{
					"content": `{
  "version": "https://jsonfeed.org/version/1.1",
  "title": "Simple Feed",
  "items": [
    {
      "id": "1",
      "title": "Item 1"
    }
  ]
}`,
					"content_type": "application/feed+json",
					"format":       "json",
				},
				Explanation: "JSON Feed with indentation for readability",
			},
			{
				Name:        "Convert with full content",
				Description: "Include complete article content",
				Scenario:    "When you want full articles in the converted feed",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"title": "Blog",
						"items": []map[string]interface{}{
							{
								"id":          "post-1",
								"title":       "Full Article",
								"description": "Summary only",
								"content":     "<article><p>This is the complete article with multiple paragraphs...</p></article>",
							},
						},
					},
					"target_type":     "atom",
					"include_content": true,
				},
				Output: map[string]interface{}{
					"content":      "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<feed xmlns=\"http://www.w3.org/2005/Atom\">...",
					"content_type": "application/atom+xml",
					"format":       "atom",
				},
				Explanation: "Atom feed with full HTML content in content element",
			},
			{
				Name:        "Author preservation",
				Description: "Convert while maintaining author information",
				Scenario:    "When author attribution is important",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"title": "Team Blog",
						"author": map[string]interface{}{
							"name": "Editorial Team",
						},
						"items": []map[string]interface{}{
							{
								"id":    "1",
								"title": "Post by John",
								"author": map[string]interface{}{
									"name":  "John Smith",
									"email": "john@example.com",
								},
							},
						},
					},
					"target_type": "json",
				},
				Output: map[string]interface{}{
					"content":      `{"version":"https://jsonfeed.org/version/1.1",...}`,
					"content_type": "application/feed+json",
					"format":       "json",
				},
				Explanation: "Author information preserved at both feed and item level",
			},
			{
				Name:        "Attachment conversion",
				Description: "Convert feeds with media attachments",
				Scenario:    "When converting podcast or media feeds",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"title": "Video Feed",
						"items": []map[string]interface{}{
							{
								"id":    "video-1",
								"title": "Tutorial Video",
								"enclosures": []map[string]interface{}{
									{
										"url":    "https://example.com/video.mp4",
										"type":   "video/mp4",
										"length": 50000000,
									},
								},
							},
						},
					},
					"target_type": "json",
				},
				Output: map[string]interface{}{
					"content":      `{..."attachments":[{"url":"https://example.com/video.mp4","mime_type":"video/mp4","size_in_bytes":50000000}]...}`,
					"content_type": "application/feed+json",
					"format":       "json",
				},
				Explanation: "Enclosures converted to JSON Feed attachments format",
			},
			{
				Name:        "Date format handling",
				Description: "Convert with proper date formatting",
				Scenario:    "When date precision and format matter",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"title":   "Event Feed",
						"updated": "2024-03-15T14:30:00Z",
						"items": []map[string]interface{}{
							{
								"id":        "event-1",
								"title":     "Upcoming Event",
								"published": "2024-03-20T09:00:00Z",
								"updated":   "2024-03-21T10:00:00Z",
							},
						},
					},
					"target_type": "rss",
				},
				Output: map[string]interface{}{
					"content":      "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<rss version=\"2.0\">...<pubDate>Mon, 20 Mar 2024 09:00:00 +0000</pubDate>...",
					"content_type": "application/rss+xml",
					"format":       "rss",
				},
				Explanation: "RFC3339 dates converted to RFC1123 format for RSS",
			},
		}).
		WithConstraints([]string{
			"Feed must be in UnifiedFeed format",
			"Target type must be 'rss', 'atom', or 'json'",
			"RSS uses RFC1123 date format",
			"Atom and JSON Feed use RFC3339 dates",
			"RSS puts content in description field",
			"Atom has separate content and summary",
			"JSON Feed supports content_html and content_text",
			"Author email required for RSS author field",
		}).
		WithErrorGuidance(map[string]string{
			"unsupported target format": "Use 'rss', 'atom', or 'json' as target_type",
			"conversion error":          "Check if feed data is valid UnifiedFeed format",
			"marshal error":             "Feed contains invalid characters or structure",
		}).
		WithCategory("feed").
		WithTags([]string{"feed", "convert", "transform", "rss", "atom", "json", "format"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "low") // Deterministic, local operation

	return builder.Build()
}

func convertFeed(params FeedConvertParams) (*FeedConvertResult, error) {
	// Default to including content
	includeContent := params.IncludeContent

	// Normalize target type
	targetType := strings.ToLower(params.TargetType)

	var content string
	var contentType string
	var err error

	switch targetType {
	case "rss", "rss2":
		content, err = convertToRSS(params.Feed, includeContent, params.Pretty)
		contentType = "application/rss+xml"
		targetType = "rss"

	case "atom":
		content, err = convertToAtom(params.Feed, includeContent, params.Pretty)
		contentType = "application/atom+xml"

	case "json", "jsonfeed":
		content, err = convertToJSONFeed(params.Feed, includeContent, params.Pretty)
		contentType = "application/feed+json"
		targetType = "json"

	default:
		return nil, fmt.Errorf("unsupported target format: %s (must be 'rss', 'atom', or 'json')", params.TargetType)
	}

	if err != nil {
		return nil, fmt.Errorf("conversion error: %w", err)
	}

	return &FeedConvertResult{
		Content:     content,
		ContentType: contentType,
		Format:      targetType,
	}, nil
}

// convertToRSS converts a UnifiedFeed to RSS 2.0 format
func convertToRSS(feed UnifiedFeed, includeContent bool, pretty bool) (string, error) {
	rss := rss2Feed{
		Version: "2.0",
	}
	rss.Channel.Title = feed.Title
	rss.Channel.Link = feed.Link
	rss.Channel.Description = feed.Description
	rss.Channel.Language = feed.Language
	rss.Channel.Copyright = feed.Copyright

	// Set channel dates
	if feed.Updated != nil {
		rss.Channel.LastBuild = feed.Updated.Format(time.RFC1123Z)
	}
	if feed.Published != nil {
		rss.Channel.PubDate = feed.Published.Format(time.RFC1123Z)
	}

	// Convert items
	for _, item := range feed.Items {
		rssItem := rssItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			GUID:        item.ID,
		}

		// Set dates
		if item.Published != nil {
			rssItem.PubDate = item.Published.Format(time.RFC1123Z)
		}

		// Set author
		if item.Author != nil && item.Author.Email != "" {
			rssItem.Author = fmt.Sprintf("%s (%s)", item.Author.Email, item.Author.Name)
		}

		// Set categories
		rssItem.Categories = item.Categories

		// RSS doesn't have a standard content field, so we include it in description if requested
		if includeContent && item.Content != "" {
			rssItem.Description = item.Content
		}

		// Convert enclosures
		for _, enc := range item.Enclosures {
			rssItem.Enclosures = append(rssItem.Enclosures, rssEnclosure(enc))
		}

		rss.Channel.Items = append(rss.Channel.Items, rssItem)
	}

	// Marshal to XML
	var output []byte
	var err error

	if pretty {
		output, err = xml.MarshalIndent(rss, "", "  ")
	} else {
		output, err = xml.Marshal(rss)
	}

	if err != nil {
		return "", err
	}

	// Add XML declaration
	return xml.Header + string(output), nil
}

// convertToAtom converts a UnifiedFeed to Atom 1.0 format
func convertToAtom(feed UnifiedFeed, includeContent bool, pretty bool) (string, error) {
	atom := atomFeed{
		XMLName: xml.Name{Space: "http://www.w3.org/2005/Atom", Local: "feed"},
		Title:   feed.Title,
		ID:      feed.Link,
	}

	// Set feed link
	if feed.Link != "" {
		atom.Links = append(atom.Links, atomLink{
			Href: feed.Link,
			Rel:  "alternate",
		})
	}

	// Set dates
	if feed.Updated != nil {
		atom.Updated = feed.Updated.Format(time.RFC3339)
	} else {
		atom.Updated = time.Now().Format(time.RFC3339)
	}

	// Set author
	if feed.Author != nil {
		atom.Author = &atomAuthor{
			Name:  feed.Author.Name,
			Email: feed.Author.Email,
			URI:   feed.Author.URL,
		}
	}

	// Convert items
	for _, item := range feed.Items {
		entry := atomEntry{
			Title: item.Title,
			ID:    item.ID,
		}

		// Set link
		if item.Link != "" {
			entry.Links = append(entry.Links, atomLink{
				Href: item.Link,
				Rel:  "alternate",
			})
		}

		// Set dates
		if item.Updated != nil {
			entry.Updated = item.Updated.Format(time.RFC3339)
		} else if item.Published != nil {
			entry.Updated = item.Published.Format(time.RFC3339)
		}
		if item.Published != nil {
			entry.Published = item.Published.Format(time.RFC3339)
		}

		// Set author
		if item.Author != nil {
			entry.Author = &atomAuthor{
				Name:  item.Author.Name,
				Email: item.Author.Email,
				URI:   item.Author.URL,
			}
		}

		// Set content
		if includeContent && item.Content != "" {
			entry.Content = atomContent{
				Type:  "html",
				Value: item.Content,
			}
		} else if item.Description != "" {
			entry.Summary = item.Description
		}

		// Set categories
		for _, cat := range item.Categories {
			entry.Categories = append(entry.Categories, atomCategory{Term: cat})
		}

		atom.Entries = append(atom.Entries, entry)
	}

	// Marshal to XML
	var output []byte
	var err error

	if pretty {
		output, err = xml.MarshalIndent(atom, "", "  ")
	} else {
		output, err = xml.Marshal(atom)
	}

	if err != nil {
		return "", err
	}

	// Add XML declaration
	return xml.Header + string(output), nil
}

// convertToJSONFeed converts a UnifiedFeed to JSON Feed 1.1 format
func convertToJSONFeed(feed UnifiedFeed, includeContent bool, pretty bool) (string, error) {
	jf := jsonFeed{
		Version:     "https://jsonfeed.org/version/1.1",
		Title:       feed.Title,
		HomePageURL: feed.Link,
		Description: feed.Description,
	}

	// Set author
	if feed.Author != nil {
		jf.Author = &jsonFeedAuthor{
			Name: feed.Author.Name,
			URL:  feed.Author.URL,
		}
	}

	// Convert items
	for _, item := range feed.Items {
		jItem := jsonFeedItem{
			ID:    item.ID,
			Title: item.Title,
			URL:   item.Link,
		}

		// Set dates
		if item.Published != nil {
			jItem.DatePublished = item.Published.Format(time.RFC3339)
		}
		if item.Updated != nil {
			jItem.DateModified = item.Updated.Format(time.RFC3339)
		}

		// Set content
		if includeContent && item.Content != "" {
			jItem.ContentHTML = item.Content
		} else if item.Description != "" {
			jItem.ContentText = item.Description
		}

		// Set author
		if item.Author != nil {
			jItem.Author = &jsonFeedAuthor{
				Name: item.Author.Name,
				URL:  item.Author.URL,
			}
		}

		// Set tags (categories)
		jItem.Tags = item.Categories

		// Convert enclosures to attachments
		for _, enc := range item.Enclosures {
			jItem.Attachments = append(jItem.Attachments, jsonAttachment{
				URL:      enc.URL,
				MimeType: enc.Type,
				Size:     enc.Length,
			})
		}

		jf.Items = append(jf.Items, jItem)
	}

	// Marshal to JSON
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(jf, "", "  ")
	} else {
		output, err = json.Marshal(jf)
	}

	if err != nil {
		return "", err
	}

	return string(output), nil
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("feed_convert", FeedConvert(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_convert",
			Category:    "feed",
			Tags:        []string{"feed", "convert", "transform", "rss", "atom", "json", "format"},
			Description: "Convert feeds between RSS, Atom, and JSON Feed formats",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Convert to RSS",
					Description: "Convert any feed format to RSS 2.0",
					Code:        `FeedConvert().Execute(ctx, FeedConvertParams{Feed: feed, TargetType: "rss", Pretty: true})`,
				},
				{
					Name:        "Convert to JSON Feed",
					Description: "Convert RSS/Atom to modern JSON Feed format",
					Code:        `FeedConvert().Execute(ctx, FeedConvertParams{Feed: feed, TargetType: "json", IncludeContent: true})`,
				},
				{
					Name:        "Convert to Atom",
					Description: "Convert to Atom without full content",
					Code:        `FeedConvert().Execute(ctx, FeedConvertParams{Feed: feed, TargetType: "atom", IncludeContent: false})`,
				},
				{
					Name:        "Pretty print",
					Description: "Convert with human-readable formatting",
					Code:        `FeedConvert().Execute(ctx, FeedConvertParams{Feed: feed, TargetType: "json", Pretty: true})`,
				},
			},
		},
		RequiredPermissions: []string{},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  false,
			Concurrency: true,
		},
		UsageInstructions: `The feed_convert tool transforms feeds between formats:
- RSS 2.0: Classic format with wide compatibility
- Atom 1.0: IETF standard with better structure
- JSON Feed 1.1: Modern JSON-based format
- Pretty printing for human readability
- Content inclusion options
- Full metadata preservation

Perfect for format migration, API requirements, and cross-platform compatibility.`,
		Constraints: []string{
			"Feed must be UnifiedFeed format",
			"Target: 'rss', 'atom', or 'json'",
			"RSS uses RFC1123 dates",
			"Atom/JSON use RFC3339 dates",
			"RSS content in description",
			"Atom has content/summary split",
		},
		ErrorGuidance: map[string]string{
			"unsupported target format": "Use 'rss', 'atom', or 'json'",
			"conversion error":          "Check feed data format",
			"marshal error":             "Invalid characters/structure",
		},
		IsDeterministic:      true, // Local operation
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "low",
	})
}
