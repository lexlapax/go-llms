// ABOUTME: FeedAggregate tool for combining multiple feeds into a single unified feed
// ABOUTME: Supports sorting, duplicate removal, and limiting the number of items

package feed

import (
	"crypto/md5" //nolint:gosec // MD5 used for deduplication, not security
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

// feedAggregateExecute is the execution function for feed_aggregate
func feedAggregateExecute(ctx *domain.ToolContext, params FeedAggregateParams) (*FeedAggregateResult, error) {
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
}

// FeedAggregate creates a tool that combines multiple feeds into a single unified feed with advanced sorting and deduplication capabilities.
// The tool supports merging feed metadata, removing duplicate items based on URL or content hash, and sorting by date or title.
// It can limit the number of items in the aggregated feed and preserve feed structure consistency across different formats.
// This is particularly useful for multi-source news aggregation, creating unified podcast feeds, and building curated content feeds.
func FeedAggregate() domain.Tool {
	// Define parameter schema
	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"feeds": {
				Type:        "array",
				Description: "Array of UnifiedFeed objects to aggregate",
				Items: &sdomain.Property{
					Type:        "object",
					Description: "Feed to include in aggregation",
				},
			},
			"sort_by": {
				Type:        "string",
				Description: "Sort field: 'date' or 'title' (default: date)",
				Enum:        []string{"date", "title"},
			},
			"sort_descending": {
				Type:        "boolean",
				Description: "Sort in descending order (default: false - oldest first for date, A-Z for title)",
			},
			"remove_dupes": {
				Type:        "boolean",
				Description: "Remove duplicate items based on URL or content hash (default: false)",
			},
			"max_items": {
				Type:        "number",
				Description: "Maximum number of items in aggregated feed (0 = unlimited)",
			},
			"merge_metadata": {
				Type:        "boolean",
				Description: "Merge feed metadata (title, description) into aggregated feed (default: false)",
			},
		},
		Required: []string{"feeds"},
	}

	// Define output schema
	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"feed": {
				Type:        "object",
				Description: "The aggregated feed combining all input feeds",
			},
			"source_count": {
				Type:        "integer",
				Description: "Number of source feeds aggregated",
			},
			"total_items": {
				Type:        "integer",
				Description: "Total number of items before filtering/limiting",
			},
			"dupes_removed": {
				Type:        "integer",
				Description: "Number of duplicate items removed (if remove_dupes=true)",
			},
		},
		Required: []string{"feed", "source_count", "total_items"},
	}

	builder := atools.NewToolBuilder("feed_aggregate", "Combine multiple feeds into one unified feed with sorting and deduplication").
		WithFunction(feedAggregateExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The feed_aggregate tool combines multiple feeds into a single unified feed:

Aggregation Features:
1. Feed Combination:
   - Merges items from all input feeds
   - Preserves all item metadata
   - Maintains feed structure consistency

2. Sorting Options:
   - By date: Published date (falls back to updated date)
   - By title: Alphabetical sorting
   - Ascending or descending order
   - Items without dates sorted to end

3. Duplicate Removal:
   - Detects duplicates by URL (primary)
   - Falls back to content hash (MD5 of title+description+content)
   - Preserves first occurrence

4. Metadata Merging:
   - Combines feed titles: "Aggregated: Feed1, Feed2, ..."
   - Joins descriptions with " | " separator
   - Uses most recent updated timestamp
   - Preserves first feed's other metadata

5. Result Limiting:
   - Apply max_items after sorting and deduplication
   - Useful for creating "top N" feeds

State Integration:
- feed_aggregate_default_sort: Default sort field (date/title)
- feed_aggregate_max_items: Default item limit

Common Use Cases:
- Multi-source news aggregation
- Creating unified podcast feeds
- Combining team/department blogs
- Building curated content feeds
- Cross-source content monitoring`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Basic feed aggregation",
				Description: "Combine multiple news feeds",
				Scenario:    "When you want to merge different news sources into one feed",
				Input: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"title": "Tech News",
							"items": []map[string]interface{}{
								{
									"title":     "AI Breakthrough",
									"link":      "https://tech.example.com/ai",
									"published": "2024-03-20T10:00:00Z",
								},
								{
									"title":     "New Smartphone",
									"link":      "https://tech.example.com/phone",
									"published": "2024-03-19T10:00:00Z",
								},
							},
						},
						{
							"title": "Science Daily",
							"items": []map[string]interface{}{
								{
									"title":     "Mars Discovery",
									"link":      "https://science.example.com/mars",
									"published": "2024-03-21T10:00:00Z",
								},
							},
						},
					},
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title":     "New Smartphone",
								"link":      "https://tech.example.com/phone",
								"published": "2024-03-19T10:00:00Z",
							},
							{
								"title":     "AI Breakthrough",
								"link":      "https://tech.example.com/ai",
								"published": "2024-03-20T10:00:00Z",
							},
							{
								"title":     "Mars Discovery",
								"link":      "https://science.example.com/mars",
								"published": "2024-03-21T10:00:00Z",
							},
						},
					},
					"source_count":  2,
					"total_items":   3,
					"dupes_removed": 0,
				},
				Explanation: "Combined 2 feeds with 3 total items, sorted by date (oldest first by default)",
			},
			{
				Name:        "Sort by date descending",
				Description: "Aggregate with most recent items first",
				Scenario:    "When you want the latest content at the top",
				Input: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"items": []map[string]interface{}{
								{"title": "Old Post", "published": "2024-01-01T10:00:00Z"},
								{"title": "Recent Post", "published": "2024-03-20T10:00:00Z"},
								{"title": "Yesterday's Post", "published": "2024-03-19T10:00:00Z"},
							},
						},
					},
					"sort_by":         "date",
					"sort_descending": true,
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{"title": "Recent Post", "published": "2024-03-20T10:00:00Z"},
							{"title": "Yesterday's Post", "published": "2024-03-19T10:00:00Z"},
							{"title": "Old Post", "published": "2024-01-01T10:00:00Z"},
						},
					},
					"source_count":  1,
					"total_items":   3,
					"dupes_removed": 0,
				},
				Explanation: "Items sorted by date in descending order (newest first)",
			},
			{
				Name:        "Remove duplicates",
				Description: "Aggregate and remove duplicate items",
				Scenario:    "When feeds might contain the same articles",
				Input: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"title": "Feed A",
							"items": []map[string]interface{}{
								{
									"title": "Shared Article",
									"link":  "https://example.com/article1",
								},
								{
									"title": "Unique to A",
									"link":  "https://example.com/article2",
								},
							},
						},
						{
							"title": "Feed B",
							"items": []map[string]interface{}{
								{
									"title": "Shared Article",
									"link":  "https://example.com/article1",
								},
								{
									"title": "Unique to B",
									"link":  "https://example.com/article3",
								},
							},
						},
					},
					"remove_dupes": true,
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title": "Shared Article",
								"link":  "https://example.com/article1",
							},
							{
								"title": "Unique to A",
								"link":  "https://example.com/article2",
							},
							{
								"title": "Unique to B",
								"link":  "https://example.com/article3",
							},
						},
					},
					"source_count":  2,
					"total_items":   4,
					"dupes_removed": 1,
				},
				Explanation: "Removed 1 duplicate item based on matching URLs",
			},
			{
				Name:        "Merge metadata",
				Description: "Aggregate with combined feed metadata",
				Scenario:    "When you want to preserve source feed information",
				Input: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"title":       "Tech Blog",
							"description": "Latest technology news",
							"items":       []map[string]interface{}{{"title": "Tech Post"}},
						},
						{
							"title":       "Science Blog",
							"description": "Scientific discoveries",
							"items":       []map[string]interface{}{{"title": "Science Post"}},
						},
					},
					"merge_metadata": true,
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"title":       "Aggregated: Tech Blog, Science Blog",
						"description": "Latest technology news | Scientific discoveries",
						"items": []map[string]interface{}{
							{"title": "Tech Post"},
							{"title": "Science Post"},
						},
					},
					"source_count": 2,
					"total_items":  2,
				},
				Explanation: "Merged feed metadata with aggregated title and combined descriptions",
			},
			{
				Name:        "Sort by title",
				Description: "Aggregate and sort alphabetically",
				Scenario:    "When you want items organized by title",
				Input: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"items": []map[string]interface{}{
								{"title": "Zebra Article"},
								{"title": "Apple News"},
								{"title": "Microsoft Update"},
							},
						},
					},
					"sort_by": "title",
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{"title": "Apple News"},
							{"title": "Microsoft Update"},
							{"title": "Zebra Article"},
						},
					},
					"source_count": 1,
					"total_items":  3,
				},
				Explanation: "Items sorted alphabetically by title (A-Z)",
			},
			{
				Name:        "Limit aggregated items",
				Description: "Create a top-N feed",
				Scenario:    "When you only want the most recent items from multiple sources",
				Input: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"items": []map[string]interface{}{
								{"title": "Post 1", "published": "2024-03-20T10:00:00Z"},
								{"title": "Post 2", "published": "2024-03-19T10:00:00Z"},
								{"title": "Post 3", "published": "2024-03-18T10:00:00Z"},
							},
						},
						{
							"items": []map[string]interface{}{
								{"title": "Post 4", "published": "2024-03-21T10:00:00Z"},
								{"title": "Post 5", "published": "2024-03-17T10:00:00Z"},
							},
						},
					},
					"sort_by":         "date",
					"sort_descending": true,
					"max_items":       3,
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{"title": "Post 4", "published": "2024-03-21T10:00:00Z"},
							{"title": "Post 1", "published": "2024-03-20T10:00:00Z"},
							{"title": "Post 2", "published": "2024-03-19T10:00:00Z"},
						},
					},
					"source_count": 2,
					"total_items":  5,
				},
				Explanation: "Limited to 3 most recent items from 5 total items",
			},
			{
				Name:        "Handle items without dates",
				Description: "Aggregate with mixed date availability",
				Scenario:    "When some items lack date information",
				Input: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"items": []map[string]interface{}{
								{"title": "Dated Item", "published": "2024-03-20T10:00:00Z"},
								{"title": "No Date Item"},
								{"title": "Another Dated", "published": "2024-03-19T10:00:00Z"},
							},
						},
					},
					"sort_by":         "date",
					"sort_descending": true,
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{"title": "Dated Item", "published": "2024-03-20T10:00:00Z"},
							{"title": "Another Dated", "published": "2024-03-19T10:00:00Z"},
							{"title": "No Date Item"},
						},
					},
					"source_count": 1,
					"total_items":  3,
				},
				Explanation: "Items without dates sorted to the end",
			},
			{
				Name:        "Empty feeds handling",
				Description: "Aggregate with some empty feeds",
				Scenario:    "When some feeds have no items",
				Input: map[string]interface{}{
					"feeds": []map[string]interface{}{
						{
							"title": "Active Feed",
							"items": []map[string]interface{}{
								{"title": "Only Item"},
							},
						},
						{
							"title": "Empty Feed",
							"items": []map[string]interface{}{},
						},
					},
				},
				Output: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{"title": "Only Item"},
						},
					},
					"source_count": 2,
					"total_items":  1,
				},
				Explanation: "Successfully aggregated feeds even with empty sources",
			},
		}).
		WithConstraints([]string{
			"Feeds must be in UnifiedFeed format",
			"Sort field must be 'date' or 'title'",
			"MaxItems of 0 means no limit",
			"Duplicate detection uses URL first, then content hash",
			"Items without dates are sorted to the end when sorting by date",
			"Title sorting is case-insensitive",
			"Deduplication preserves first occurrence",
			"Empty feeds are handled gracefully",
		}).
		WithErrorGuidance(map[string]string{
			"invalid sort_by field": "Use 'date' or 'title' for sort_by parameter",
			"no feeds provided":     "Provide at least one feed in the feeds array",
		}).
		WithCategory("feed").
		WithTags([]string{"feed", "aggregate", "combine", "merge", "sort", "deduplicate"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "medium") // Deterministic, local operation, medium memory

	return builder.Build()
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
		hash := md5.Sum([]byte(content)) //nolint:gosec // MD5 used for deduplication, not security
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

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("feed_aggregate", FeedAggregate(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_aggregate",
			Category:    "feed",
			Tags:        []string{"feed", "aggregate", "combine", "merge", "sort", "deduplicate"},
			Description: "Combine multiple feeds into one unified feed with sorting and deduplication",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic aggregation",
					Description: "Combine multiple feeds",
					Code:        `FeedAggregate().Execute(ctx, FeedAggregateParams{Feeds: []UnifiedFeed{feed1, feed2}})`,
				},
				{
					Name:        "Sort and deduplicate",
					Description: "Aggregate with sorting and duplicate removal",
					Code:        `FeedAggregate().Execute(ctx, FeedAggregateParams{Feeds: feeds, SortBy: "date", SortDescending: true, RemoveDupes: true})`,
				},
				{
					Name:        "Limited aggregation",
					Description: "Aggregate with item limit",
					Code:        `FeedAggregate().Execute(ctx, FeedAggregateParams{Feeds: feeds, MaxItems: 20, MergeMetadata: true})`,
				},
				{
					Name:        "Sort by title",
					Description: "Aggregate and sort alphabetically",
					Code:        `FeedAggregate().Execute(ctx, FeedAggregateParams{Feeds: feeds, SortBy: "title"})`,
				},
			},
		},
		RequiredPermissions: []string{},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "medium",
			Network:     false,
			FileSystem:  false,
			Concurrency: true,
		},
		UsageInstructions: `The feed_aggregate tool combines multiple feeds:
- Feed merging with metadata combination
- Sorting by date or title (ascending/descending)
- Duplicate removal based on URL or content hash
- Item limiting for top-N feeds
- Metadata merging for aggregated feed info

Perfect for multi-source aggregation, content curation, and feed consolidation.`,
		Constraints: []string{
			"Feeds must be UnifiedFeed format",
			"Sort field must be 'date' or 'title'",
			"MaxItems 0 = no limit",
			"Duplicate detection uses URL first",
			"Items without dates sorted to end",
			"Title sorting is case-insensitive",
		},
		ErrorGuidance: map[string]string{
			"invalid sort_by field": "Use 'date' or 'title'",
			"no feeds provided":     "Provide at least one feed",
		},
		IsDeterministic:      true, // Local operation
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "low",
	})
}
