// ABOUTME: FeedConvert tool for converting between different feed formats (RSS, Atom, JSON Feed)
// ABOUTME: Supports conversion with optional pretty-printing for human-readable output

package feed

import (
	"context"
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

// feedConvertParamSchema defines parameters for the FeedConvert tool
var feedConvertParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"feed": {
			Type:        "object",
			Description: "Feed data to convert (from FeedFetch or other tools)",
		},
		"target_type": {
			Type:        "string",
			Description: "Target format: rss, atom, json",
		},
		"pretty": {
			Type:        "boolean",
			Description: "Pretty-print the output (default: false)",
		},
		"include_content": {
			Type:        "boolean",
			Description: "Include full content in output (default: true)",
		},
	},
	Required: []string{"feed", "target_type"},
}

func init() {
	tools.MustRegisterTool("feed_convert", FeedConvert(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_convert",
			Category:    "feed",
			Tags:        []string{"feed", "convert", "transform", "rss", "atom", "json"},
			Description: "Converts feeds between RSS, Atom, and JSON Feed formats",
			Version:     "1.0.0",
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
					Name:        "Minimal Atom conversion",
					Description: "Convert to Atom without full content",
					Code:        `FeedConvert().Execute(ctx, FeedConvertParams{Feed: feed, TargetType: "atom", IncludeContent: false})`,
				},
			},
		},
		RequiredPermissions: []string{},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  false,
			Concurrency: false,
		},
	})
}

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

// FeedConvert creates a new FeedConvert tool
func FeedConvert() domain.Tool {
	return atools.NewTool(
		"feed_convert",
		"Converts feeds between RSS, Atom, and JSON Feed formats",
		func(ctx context.Context, params FeedConvertParams) (*FeedConvertResult, error) {
			return convertFeed(params)
		},
		feedConvertParamSchema,
	)
}

func convertFeed(params FeedConvertParams) (*FeedConvertResult, error) {
	// Default to including content
	includeContent := true
	if !params.IncludeContent {
		includeContent = false
	}

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
