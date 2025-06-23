// ABOUTME: FeedExtract tool for extracting specific fields from feed items
// ABOUTME: Supports field selection, flattening, and customizable extraction

package feed

import (
	"fmt"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// FeedExtractParams contains parameters for the FeedExtract tool
type FeedExtractParams struct {
	Feed            UnifiedFeed `json:"feed"`
	Fields          []string    `json:"fields"`
	Flatten         bool        `json:"flatten,omitempty"`
	IncludeMetadata bool        `json:"include_metadata,omitempty"`
	MaxItems        int         `json:"max_items,omitempty"`
}

// FeedExtractResult contains the result of feed extraction
type FeedExtractResult struct {
	Data     []map[string]interface{} `json:"data"`
	Metadata map[string]interface{}   `json:"metadata,omitempty"`
	Count    int                      `json:"count"`
	Fields   []string                 `json:"fields"`
}

// feedExtractExecute is the execution function for feed_extract
func feedExtractExecute(ctx *domain.ToolContext, params FeedExtractParams) (*FeedExtractResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
			ToolName:   "feed_extract",
			Parameters: params,
			RequestID:  ctx.RunID,
		})
	}

	// Check state for default max items if not provided
	if params.MaxItems == 0 && ctx.State != nil {
		if val, ok := ctx.State.Get("feed_extract_max_items"); ok {
			if max, ok := val.(int); ok && max > 0 {
				params.MaxItems = max
			}
		}
	}

	// Check state for default fields if none provided
	if len(params.Fields) == 0 && ctx.State != nil {
		if val, ok := ctx.State.Get("feed_extract_default_fields"); ok {
			if fields, ok := val.([]string); ok && len(fields) > 0 {
				params.Fields = fields
			}
		}
	}

	result, err := extractFromFeed(params)
	if err != nil {
		return nil, err
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
			ToolName:  "feed_extract",
			Result:    result,
			RequestID: ctx.RunID,
		})
	}

	return result, nil
}

// FeedExtract creates a tool that selectively extracts specific fields from feed items for structured data analysis and transformation.
// The tool supports extraction of basic fields, nested fields using dot notation, and media enclosures, with optional field flattening.
// It can include feed-level metadata in results and limit the number of items processed for performance optimization.
// This is perfect for data transformation, creating simplified feed summaries, and preparing feed data for external systems.
func FeedExtract() domain.Tool {
	// Define parameter schema
	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"feed": {
				Type:        "object",
				Description: "Feed data to extract from (UnifiedFeed format)",
			},
			"fields": {
				Type:        "array",
				Description: "List of fields to extract from each item",
				Items: &sdomain.Property{
					Type: "string",
				},
			},
			"flatten": {
				Type:        "boolean",
				Description: "Flatten nested fields like author.name to author_name (default: false)",
			},
			"include_metadata": {
				Type:        "boolean",
				Description: "Include feed metadata in results (default: false)",
			},
			"max_items": {
				Type:        "number",
				Description: "Maximum number of items to extract (0 = all)",
			},
		},
		Required: []string{"feed", "fields"},
	}

	// Define output schema
	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"data": {
				Type:        "array",
				Description: "Extracted data from feed items",
				Items: &sdomain.Property{
					Type:        "object",
					Description: "Extracted fields from a feed item",
				},
			},
			"metadata": {
				Type:        "object",
				Description: "Feed metadata (if include_metadata is true)",
			},
			"count": {
				Type:        "integer",
				Description: "Number of items extracted",
			},
			"fields": {
				Type:        "array",
				Description: "List of fields that were extracted",
				Items: &sdomain.Property{
					Type: "string",
				},
			},
		},
		Required: []string{"data", "count", "fields"},
	}

	builder := atools.NewToolBuilder("feed_extract", "Extract specific fields from feed items for structured data analysis").
		WithFunction(feedExtractExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The feed_extract tool provides selective field extraction from feed data:

Field Extraction:
1. Basic Fields:
   - id, title, description, content, link
   - published, updated (formatted as RFC3339)
   - categories (array of strings)

2. Nested Fields:
   - author.name, author.email, author.url
   - Individual author fields or full author object

3. Media Fields:
   - enclosures (array of media attachments)
   - Each enclosure contains url, type, length

Advanced Features:
- Field Flattening: Convert author.name to author_name
- Metadata Inclusion: Extract feed-level information
- Item Limiting: Control number of items processed
- State Integration: Default fields and limits from state

State Integration:
- feed_extract_max_items: Default maximum items
- feed_extract_default_fields: Default field list

Common Use Cases:
- Data transformation for analytics
- Creating simplified feed summaries
- Extracting specific content types
- Preparing data for external systems
- Content migration and archiving`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Basic field extraction",
				Description: "Extract titles and links from feed items",
				Scenario:    "When you need just the essential information from feeds",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"title": "Tech Blog",
						"items": []map[string]interface{}{
							{
								"id":    "post-1",
								"title": "Latest Technology Trends",
								"link":  "https://blog.example.com/post-1",
							},
							{
								"id":    "post-2",
								"title": "AI Developments",
								"link":  "https://blog.example.com/post-2",
							},
						},
					},
					"fields": []string{"title", "link"},
				},
				Output: map[string]interface{}{
					"data": []map[string]interface{}{
						{
							"title": "Latest Technology Trends",
							"link":  "https://blog.example.com/post-1",
						},
						{
							"title": "AI Developments",
							"link":  "https://blog.example.com/post-2",
						},
					},
					"count":  2,
					"fields": []string{"title", "link"},
				},
				Explanation: "Extracted only title and link fields from each feed item",
			},
			{
				Name:        "Extract with author information",
				Description: "Extract nested author fields",
				Scenario:    "When you need author information from articles",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title": "Article by John",
								"author": map[string]interface{}{
									"name":  "John Doe",
									"email": "john@example.com",
								},
								"published": "2024-03-15T10:00:00Z",
							},
						},
					},
					"fields": []string{"title", "author.name", "author.email", "published"},
				},
				Output: map[string]interface{}{
					"data": []map[string]interface{}{
						{
							"title":        "Article by John",
							"author.name":  "John Doe",
							"author.email": "john@example.com",
							"published":    "2024-03-15T10:00:00Z",
						},
					},
					"count":  1,
					"fields": []string{"title", "author.name", "author.email", "published"},
				},
				Explanation: "Extracted nested author fields using dot notation",
			},
			{
				Name:        "Flattened field extraction",
				Description: "Extract and flatten nested fields",
				Scenario:    "When you want simplified field names for external systems",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title": "Article with Author",
								"author": map[string]interface{}{
									"name": "Jane Smith",
									"url":  "https://jane.example.com",
								},
							},
						},
					},
					"fields":  []string{"title", "author.name", "author.url"},
					"flatten": true,
				},
				Output: map[string]interface{}{
					"data": []map[string]interface{}{
						{
							"title":       "Article with Author",
							"author_name": "Jane Smith",
							"author_url":  "https://jane.example.com",
						},
					},
					"count":  1,
					"fields": []string{"title", "author.name", "author.url"},
				},
				Explanation: "Flattened nested fields from author.name to author_name",
			},
			{
				Name:        "Extract with feed metadata",
				Description: "Include feed-level information in results",
				Scenario:    "When you need both item data and feed context",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"title":       "News Feed",
						"description": "Latest news updates",
						"link":        "https://news.example.com",
						"language":    "en",
						"items": []map[string]interface{}{
							{"title": "Breaking News", "id": "news-1"},
						},
					},
					"fields":           []string{"title", "id"},
					"include_metadata": true,
				},
				Output: map[string]interface{}{
					"data": []map[string]interface{}{
						{"title": "Breaking News", "id": "news-1"},
					},
					"metadata": map[string]interface{}{
						"title":       "News Feed",
						"description": "Latest news updates",
						"link":        "https://news.example.com",
						"language":    "en",
					},
					"count":  1,
					"fields": []string{"title", "id"},
				},
				Explanation: "Included feed metadata alongside extracted item data",
			},
			{
				Name:        "Limited item extraction",
				Description: "Extract from first N items only",
				Scenario:    "When you only need recent items from a large feed",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{"title": "Post 1", "link": "link1"},
							{"title": "Post 2", "link": "link2"},
							{"title": "Post 3", "link": "link3"},
							{"title": "Post 4", "link": "link4"},
						},
					},
					"fields":    []string{"title", "link"},
					"max_items": 2,
				},
				Output: map[string]interface{}{
					"data": []map[string]interface{}{
						{"title": "Post 1", "link": "link1"},
						{"title": "Post 2", "link": "link2"},
					},
					"count":  2,
					"fields": []string{"title", "link"},
				},
				Explanation: "Limited extraction to first 2 items",
			},
			{
				Name:        "Extract media enclosures",
				Description: "Extract podcast or media attachments",
				Scenario:    "When working with podcast feeds or media content",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title": "Episode 42",
								"enclosures": []map[string]interface{}{
									{
										"url":    "https://podcast.example.com/ep42.mp3",
										"type":   "audio/mpeg",
										"length": 25000000,
									},
								},
							},
						},
					},
					"fields": []string{"title", "enclosures"},
				},
				Output: map[string]interface{}{
					"data": []map[string]interface{}{
						{
							"title": "Episode 42",
							"enclosures": []map[string]interface{}{
								{
									"url":    "https://podcast.example.com/ep42.mp3",
									"type":   "audio/mpeg",
									"length": 25000000,
								},
							},
						},
					},
					"count":  1,
					"fields": []string{"title", "enclosures"},
				},
				Explanation: "Extracted media enclosure information for podcast episode",
			},
			{
				Name:        "Categories and tags extraction",
				Description: "Extract taxonomies and content classification",
				Scenario:    "When you need to analyze content categories",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title":      "Tech Article",
								"categories": []string{"technology", "programming", "web"},
								"content":    "Article about web development...",
							},
						},
					},
					"fields": []string{"title", "categories", "content"},
				},
				Output: map[string]interface{}{
					"data": []map[string]interface{}{
						{
							"title":      "Tech Article",
							"categories": []string{"technology", "programming", "web"},
							"content":    "Article about web development...",
						},
					},
					"count":  1,
					"fields": []string{"title", "categories", "content"},
				},
				Explanation: "Extracted content categories and full text content",
			},
			{
				Name:        "Full author object extraction",
				Description: "Extract complete author information",
				Scenario:    "When you need all author details as a structured object",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title": "Expert Opinion",
								"author": map[string]interface{}{
									"name":  "Dr. Sarah Wilson",
									"email": "sarah@university.edu",
									"url":   "https://university.edu/faculty/sarah",
								},
							},
						},
					},
					"fields": []string{"title", "author"},
				},
				Output: map[string]interface{}{
					"data": []map[string]interface{}{
						{
							"title": "Expert Opinion",
							"author": map[string]interface{}{
								"name":  "Dr. Sarah Wilson",
								"email": "sarah@university.edu",
								"url":   "https://university.edu/faculty/sarah",
							},
						},
					},
					"count":  1,
					"fields": []string{"title", "author"},
				},
				Explanation: "Extracted complete author object with all available fields",
			},
		}).
		WithConstraints([]string{
			"Feed must be in UnifiedFeed format",
			"Fields array cannot be empty",
			"MaxItems of 0 extracts from all items",
			"Nested fields use dot notation (author.name)",
			"Date fields are formatted as RFC3339",
			"Missing fields are omitted from results",
			"Field flattening replaces dots with underscores",
		}).
		WithErrorGuidance(map[string]string{
			"no fields specified":   "Provide at least one field in the fields array",
			"invalid feed format":   "Feed must be a valid UnifiedFeed object with items array",
			"field not found":       "Check field names - available fields: id, title, description, content, link, published, updated, author.*, categories, enclosures",
			"no items to extract":   "Feed contains no items or all items were filtered out",
			"max_items must be > 0": "MaxItems must be a positive number or 0 for all items",
		}).
		WithCategory("feed").
		WithTags([]string{"feed", "extract", "transform", "data", "parsing", "analysis"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "low") // Deterministic, local operation

	return builder.Build()
}

func extractFromFeed(params FeedExtractParams) (*FeedExtractResult, error) {
	if len(params.Fields) == 0 {
		return nil, fmt.Errorf("no fields specified for extraction")
	}

	result := &FeedExtractResult{
		Data:   make([]map[string]interface{}, 0),
		Fields: params.Fields,
	}

	// Extract metadata if requested
	if params.IncludeMetadata {
		result.Metadata = extractFeedMetadata(params.Feed)
	}

	// Determine items to process
	items := params.Feed.Items
	if params.MaxItems > 0 && len(items) > params.MaxItems {
		items = items[:params.MaxItems]
	}

	// Extract fields from each item
	for _, item := range items {
		extracted := extractItemFields(item, params.Fields, params.Flatten)
		if len(extracted) > 0 {
			result.Data = append(result.Data, extracted)
		}
	}

	result.Count = len(result.Data)
	return result, nil
}

// extractFeedMetadata extracts metadata from the feed
func extractFeedMetadata(feed UnifiedFeed) map[string]interface{} {
	metadata := make(map[string]interface{})

	if feed.Title != "" {
		metadata["title"] = feed.Title
	}
	if feed.Description != "" {
		metadata["description"] = feed.Description
	}
	if feed.Link != "" {
		metadata["link"] = feed.Link
	}
	if feed.Language != "" {
		metadata["language"] = feed.Language
	}
	if feed.Copyright != "" {
		metadata["copyright"] = feed.Copyright
	}
	if feed.Author != nil {
		metadata["author"] = map[string]string{
			"name":  feed.Author.Name,
			"email": feed.Author.Email,
			"url":   feed.Author.URL,
		}
	}
	if feed.Updated != nil {
		metadata["updated"] = feed.Updated.Format("2006-01-02T15:04:05Z07:00")
	}
	if feed.Published != nil {
		metadata["published"] = feed.Published.Format("2006-01-02T15:04:05Z07:00")
	}

	return metadata
}

// extractItemFields extracts specified fields from a feed item
func extractItemFields(item FeedItem, fields []string, flatten bool) map[string]interface{} {
	extracted := make(map[string]interface{})

	for _, field := range fields {
		value := getFieldValue(item, field)
		if value != nil {
			if flatten {
				field = strings.ReplaceAll(field, ".", "_")
			}
			extracted[field] = value
		}
	}

	return extracted
}

// getFieldValue retrieves a field value from a feed item
func getFieldValue(item FeedItem, field string) interface{} {
	// Handle nested fields
	parts := strings.Split(field, ".")

	switch parts[0] {
	case "id":
		return item.ID
	case "title":
		return item.Title
	case "description":
		return item.Description
	case "content":
		return item.Content
	case "link":
		return item.Link
	case "published":
		if item.Published != nil {
			return item.Published.Format("2006-01-02T15:04:05Z07:00")
		}
		return nil
	case "updated":
		if item.Updated != nil {
			return item.Updated.Format("2006-01-02T15:04:05Z07:00")
		}
		return nil
	case "author":
		if item.Author == nil {
			return nil
		}
		if len(parts) > 1 {
			switch parts[1] {
			case "name":
				return item.Author.Name
			case "email":
				return item.Author.Email
			case "url":
				return item.Author.URL
			}
		}
		// Return full author object if no sub-field specified
		return map[string]string{
			"name":  item.Author.Name,
			"email": item.Author.Email,
			"url":   item.Author.URL,
		}
	case "categories":
		return item.Categories
	case "enclosures":
		if len(item.Enclosures) == 0 {
			return nil
		}
		// Convert enclosures to maps
		enclosures := make([]map[string]interface{}, len(item.Enclosures))
		for i, enc := range item.Enclosures {
			enclosures[i] = map[string]interface{}{
				"url":    enc.URL,
				"type":   enc.Type,
				"length": enc.Length,
			}
		}
		return enclosures
	default:
		return nil
	}
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("feed_extract", FeedExtract(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_extract",
			Category:    "feed",
			Tags:        []string{"feed", "extract", "transform", "data", "parsing", "analysis"},
			Description: "Extract specific fields from feed items for structured data analysis",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic extraction",
					Description: "Extract titles and links from feed items",
					Code:        `FeedExtract().Execute(ctx, FeedExtractParams{Feed: feed, Fields: []string{"title", "link"}})`,
				},
				{
					Name:        "Extract with metadata",
					Description: "Include feed metadata in extraction",
					Code:        `FeedExtract().Execute(ctx, FeedExtractParams{Feed: feed, Fields: []string{"title", "published", "author.name"}, IncludeMetadata: true})`,
				},
				{
					Name:        "Flatten nested fields",
					Description: "Extract and flatten author information",
					Code:        `FeedExtract().Execute(ctx, FeedExtractParams{Feed: feed, Fields: []string{"title", "author.name", "author.email"}, Flatten: true, MaxItems: 20})`,
				},
				{
					Name:        "Extract media",
					Description: "Extract podcast enclosures",
					Code:        `FeedExtract().Execute(ctx, FeedExtractParams{Feed: feed, Fields: []string{"title", "enclosures"}})`,
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
		UsageInstructions: `The feed_extract tool provides selective field extraction from feed data:
- Supports basic fields: id, title, description, content, link, published, updated, categories
- Nested fields with dot notation: author.name, author.email, author.url
- Media fields: enclosures (array of media attachments)
- Field flattening: author.name becomes author_name
- Feed metadata inclusion for context
- Item limiting for large feeds

Use for data transformation, analytics, content migration, and system integration.`,
		Constraints: []string{
			"Feed must be UnifiedFeed format",
			"Fields array cannot be empty",
			"MaxItems of 0 extracts all items",
			"Date fields formatted as RFC3339",
			"Missing fields omitted from results",
		},
		ErrorGuidance: map[string]string{
			"no fields specified": "Provide at least one field",
			"invalid feed format": "Feed must be valid UnifiedFeed",
			"field not found":     "Check available field names",
			"no items to extract": "Feed contains no items",
		},
		IsDeterministic:      true, // Local operation
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "low",
	})
}
