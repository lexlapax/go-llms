// ABOUTME: FeedAggregate tool for combining multiple feeds into a single unified feed
// ABOUTME: Supports sorting, duplicate removal, and limiting the number of items

package feed

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// feedAggregateParamSchema defines parameters for the FeedAggregate tool
var feedAggregateParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"feeds": {
			Type:        "array",
			Description: "Array of feeds to aggregate",
		},
		"sort_by": {
			Type:        "string",
			Description: "Sort field: date, title (default: date)",
		},
		"sort_descending": {
			Type:        "boolean",
			Description: "Sort in descending order (default: false)",
		},
		"remove_dupes": {
			Type:        "boolean",
			Description: "Remove duplicate items based on URL or content similarity (default: false)",
		},
		"max_items": {
			Type:        "number",
			Description: "Maximum number of items in aggregated feed",
		},
		"merge_metadata": {
			Type:        "boolean",
			Description: "Merge feed metadata (title, description) into aggregated feed (default: false)",
		},
	},
	Required: []string{"feeds"},
}

func init() {
	tools.MustRegisterTool("feed_aggregate", FeedAggregate(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_aggregate",
			Category:    "feed",
			Tags:        []string{"feed", "aggregate", "combine", "merge", "sort"},
			Description: "Combines multiple feeds into one unified feed",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Combine news feeds",
					Description: "Merge multiple news feeds into one",
					Code:        `FeedAggregate().Execute(ctx, FeedAggregateParams{Feeds: []UnifiedFeed{techFeed, scienceFeed, businessFeed}})`,
				},
				{
					Name:        "Sort by date descending",
					Description: "Aggregate and sort by most recent first",
					Code:        `FeedAggregate().Execute(ctx, FeedAggregateParams{Feeds: feeds, SortBy: "date", SortDescending: true, MaxItems: 100})`,
				},
				{
					Name:        "Remove duplicates",
					Description: "Combine feeds and remove duplicate articles",
					Code:        `FeedAggregate().Execute(ctx, FeedAggregateParams{Feeds: feeds, RemoveDupes: true, MergeMetadata: true})`,
				},
			},
		},
		RequiredPermissions: []string{},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "medium",
			Network:     false,
			FileSystem:  false,
			Concurrency: false,
		},
	})
}

// FeedAggregateParams contains parameters for the FeedAggregate tool
type FeedAggregateParams struct {
	Feeds          []UnifiedFeed `json:"feeds"`
	SortBy         string        `json:"sort_by,omitempty"`
	SortDescending bool          `json:"sort_descending,omitempty"`
	RemoveDupes    bool          `json:"remove_dupes,omitempty"`
	MaxItems       int           `json:"max_items,omitempty"`
	MergeMetadata  bool          `json:"merge_metadata,omitempty"`
}

// FeedAggregateResult contains the result of feed aggregation
type FeedAggregateResult struct {
	Feed         UnifiedFeed `json:"feed"`
	SourceCount  int         `json:"source_count"`
	TotalItems   int         `json:"total_items"`
	DupesRemoved int         `json:"dupes_removed,omitempty"`
}

// FeedAggregate creates a new FeedAggregate tool
func FeedAggregate() domain.Tool {
	return atools.NewTool(
		"feed_aggregate",
		"Combines multiple feeds into one unified feed with sorting and deduplication",
		func(ctx *domain.ToolContext, params FeedAggregateParams) (*FeedAggregateResult, error) {
			// Emit start event
			if ctx.Events != nil {
				ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
					ToolName:   "feed_aggregate",
					Parameters: params,
					RequestID:  ctx.RunID,
				})
			}

			// Check state for default sort options if not provided
			if params.SortBy == "" && ctx.State != nil {
				if val, ok := ctx.State.Get("feed_aggregate_default_sort"); ok {
					if sort, ok := val.(string); ok && sort != "" {
						params.SortBy = sort
					}
				}
			}

			// Check state for default max items if not provided
			if params.MaxItems == 0 && ctx.State != nil {
				if val, ok := ctx.State.Get("feed_aggregate_max_items"); ok {
					if max, ok := val.(int); ok && max > 0 {
						params.MaxItems = max
					}
				}
			}

			result, err := aggregateFeeds(params)
			if err != nil {
				return nil, err
			}

			// Emit result event
			if ctx.Events != nil {
				ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
					ToolName:  "feed_aggregate",
					Result:    result,
					RequestID: ctx.RunID,
				})
			}

			return result, nil
		},
		feedAggregateParamSchema,
	)
}

func aggregateFeeds(params FeedAggregateParams) (*FeedAggregateResult, error) {
	if len(params.Feeds) == 0 {
		return &FeedAggregateResult{
			Feed:        UnifiedFeed{},
			SourceCount: 0,
			TotalItems:  0,
		}, nil
	}

	// Set defaults
	sortBy := params.SortBy
	if sortBy == "" {
		sortBy = "date"
	}

	// Use sort descending as provided (default is false due to omitempty)
	sortDescending := params.SortDescending

	// Default to merging metadata if not explicitly set to false
	mergeMetadata := params.MergeMetadata

	// Validate sort field
	if sortBy != "date" && sortBy != "title" {
		return nil, fmt.Errorf("invalid sort_by field: %s (must be 'date' or 'title')", sortBy)
	}

	// Create aggregated feed
	var aggregated UnifiedFeed

	// Merge metadata if requested
	if mergeMetadata {
		aggregated = mergeFeeds(params.Feeds)
		aggregated.Items = make([]FeedItem, 0)
	} else {
		aggregated = UnifiedFeed{
			Items: make([]FeedItem, 0),
		}
	}

	// Collect all items
	allItems := make([]FeedItem, 0)
	for _, feed := range params.Feeds {
		allItems = append(allItems, feed.Items...)
	}

	totalItems := len(allItems)
	dupesRemoved := 0

	// Remove duplicates if requested
	if params.RemoveDupes {
		uniqueItems, removed := removeDuplicates(allItems)
		allItems = uniqueItems
		dupesRemoved = removed
	}

	// Sort items
	sortItems(allItems, sortBy, sortDescending)

	// Apply max items limit
	if params.MaxItems > 0 && len(allItems) > params.MaxItems {
		allItems = allItems[:params.MaxItems]
	}

	aggregated.Items = allItems

	return &FeedAggregateResult{
		Feed:         aggregated,
		SourceCount:  len(params.Feeds),
		TotalItems:   totalItems,
		DupesRemoved: dupesRemoved,
	}, nil
}

// mergeFeeds merges metadata from multiple feeds
func mergeFeeds(feeds []UnifiedFeed) UnifiedFeed {
	if len(feeds) == 0 {
		return UnifiedFeed{}
	}

	// Start with the first feed
	merged := UnifiedFeed{
		Title:       feeds[0].Title,
		Description: feeds[0].Description,
		Link:        feeds[0].Link,
		Language:    feeds[0].Language,
		Copyright:   feeds[0].Copyright,
		Author:      feeds[0].Author,
	}

	// If multiple feeds, create aggregated metadata
	if len(feeds) > 1 {
		titles := make([]string, 0)
		descriptions := make([]string, 0)

		for _, feed := range feeds {
			if feed.Title != "" {
				titles = append(titles, feed.Title)
			}
			if feed.Description != "" {
				descriptions = append(descriptions, feed.Description)
			}
		}

		if len(titles) > 0 {
			merged.Title = "Aggregated: " + strings.Join(titles, ", ")
		}
		if len(descriptions) > 0 {
			merged.Description = strings.Join(descriptions, " | ")
		}
	}

	// Use the most recent updated time
	for _, feed := range feeds {
		if feed.Updated != nil {
			if merged.Updated == nil || feed.Updated.After(*merged.Updated) {
				merged.Updated = feed.Updated
			}
		}
	}

	return merged
}

// removeDuplicates removes duplicate items based on URL and content similarity
func removeDuplicates(items []FeedItem) ([]FeedItem, int) {
	seen := make(map[string]bool)
	unique := make([]FeedItem, 0)
	dupesRemoved := 0

	for _, item := range items {
		// Create a unique key based on URL or content hash
		key := getItemKey(item)

		if !seen[key] {
			seen[key] = true
			unique = append(unique, item)
		} else {
			dupesRemoved++
		}
	}

	return unique, dupesRemoved
}

// getItemKey creates a unique key for an item
func getItemKey(item FeedItem) string {
	// First try URL
	if item.Link != "" {
		return item.Link
	}

	// If no URL, use content hash
	content := item.Title + item.Description + item.Content
	if content != "" {
		hash := md5.Sum([]byte(content))
		return fmt.Sprintf("%x", hash)
	}

	// Fallback to ID
	return item.ID
}

// sortItems sorts feed items by the specified field
func sortItems(items []FeedItem, sortBy string, descending bool) {
	sort.Slice(items, func(i, j int) bool {
		switch sortBy {
		case "date":
			// Get dates for comparison
			dateI := getItemDate(items[i])
			dateJ := getItemDate(items[j])

			// Handle nil dates
			if dateI == nil && dateJ == nil {
				return false
			}
			if dateI == nil {
				return !descending // nil dates go to end
			}
			if dateJ == nil {
				return descending // nil dates go to end
			}

			if descending {
				return dateI.After(*dateJ)
			}
			return dateI.Before(*dateJ)

		case "title":
			titleI := strings.ToLower(items[i].Title)
			titleJ := strings.ToLower(items[j].Title)

			if descending {
				return titleI > titleJ
			}
			return titleI < titleJ

		default:
			return false
		}
	})
}

// getItemDate returns the best available date for an item
func getItemDate(item FeedItem) *time.Time {
	if item.Published != nil {
		return item.Published
	}
	if item.Updated != nil {
		return item.Updated
	}
	return nil
}
