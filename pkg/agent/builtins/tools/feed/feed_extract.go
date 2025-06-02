// ABOUTME: FeedExtract tool for extracting specific fields from feed items
// ABOUTME: Supports field selection, flattening, and customizable extraction

package feed

import (
	"context"
	"fmt"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// feedExtractParamSchema defines parameters for the FeedExtract tool
var feedExtractParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"feed": {
			Type:        "object",
			Description: "Feed data to extract from",
		},
		"fields": {
			Type:        "array",
			Description: "List of fields to extract from each item (e.g., title, link, published)",
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
			Description: "Maximum number of items to extract from",
		},
	},
	Required: []string{"feed", "fields"},
}

func init() {
	tools.MustRegisterTool("feed_extract", FeedExtract(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_extract",
			Category:    "feed",
			Tags:        []string{"feed", "extract", "transform", "data"},
			Description: "Extracts specific fields from feed items",
		},
	})
}

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

// FeedExtract creates a new FeedExtract tool
func FeedExtract() domain.Tool {
	return atools.NewTool(
		"feed_extract",
		"Extracts specific fields from feed items for structured data analysis",
		func(ctx context.Context, params FeedExtractParams) (*FeedExtractResult, error) {
			return extractFromFeed(params)
		},
		feedExtractParamSchema,
	)
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

