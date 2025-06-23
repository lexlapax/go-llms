// ABOUTME: FeedFilter tool for filtering feed items by date, keywords, author, and categories
// ABOUTME: Provides flexible filtering capabilities to extract relevant items from feeds

package feed

import (
	"fmt"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// FeedFilterParams contains parameters for the FeedFilter tool
type FeedFilterParams struct {
	Feed       UnifiedFeed `json:"feed"`
	Keywords   []string    `json:"keywords,omitempty"`
	Authors    []string    `json:"authors,omitempty"`
	Categories []string    `json:"categories,omitempty"`
	After      string      `json:"after,omitempty"`
	Before     string      `json:"before,omitempty"`
	MaxItems   int         `json:"max_items,omitempty"`
	MatchAll   bool        `json:"match_all,omitempty"`
}

// FeedFilterResult contains the result of feed filtering
type FeedFilterResult struct {
	Items       []FeedItem `json:"items"`
	TotalItems  int        `json:"total_items"`
	FilteredOut int        `json:"filtered_out"`
}

// feedFilterExecute is the execution function for feed_filter
func feedFilterExecute(ctx *domain.ToolContext, params FeedFilterParams) (*FeedFilterResult, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolCall, domain.ToolCallEventData{
			ToolName:   "feed_filter",
			Parameters: params,
			RequestID:  ctx.RunID,
		})
	}

	// Check state for default max items if not provided
	if params.MaxItems == 0 && ctx.State != nil {
		if val, ok := ctx.State.Get("feed_filter_max_items"); ok {
			if max, ok := val.(int); ok && max > 0 {
				params.MaxItems = max
			}
		}
	}

	// Check state for default match mode if not provided
	if !params.MatchAll && ctx.State != nil {
		if val, ok := ctx.State.Get("feed_filter_match_all"); ok {
			if matchAll, ok := val.(bool); ok {
				params.MatchAll = matchAll
			}
		}
	}

	result, err := filterFeed(params)
	if err != nil {
		return nil, err
	}

	// Emit result event
	if ctx.Events != nil {
		ctx.Events.Emit(domain.EventToolResult, domain.ToolResultEventData{
			ToolName:  "feed_filter",
			Result:    result,
			RequestID: ctx.RunID,
		})
	}

	return result, nil
}

// FeedFilter creates a tool that filters feed items based on multiple criteria including keywords, date ranges, authors, and categories.
// The tool supports flexible matching modes (ANY or ALL criteria), case-insensitive partial matching, and sophisticated date filtering.
// It can search across title, description, and content fields, with options to limit the number of results returned.
// This is ideal for content curation, creating topic-specific feeds, finding recent updates, and building personalized content streams.
func FeedFilter() domain.Tool {
	// Define parameter schema
	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"feed": {
				Type:        "object",
				Description: "Feed data to filter (UnifiedFeed format)",
			},
			"keywords": {
				Type:        "array",
				Description: "Keywords to match in title/content (case-insensitive)",
				Items: &sdomain.Property{
					Type: "string",
				},
			},
			"authors": {
				Type:        "array",
				Description: "Filter by author names (partial match, case-insensitive)",
				Items: &sdomain.Property{
					Type: "string",
				},
			},
			"categories": {
				Type:        "array",
				Description: "Filter by categories (partial match, case-insensitive)",
				Items: &sdomain.Property{
					Type: "string",
				},
			},
			"after": {
				Type:        "string",
				Format:      "date-time",
				Description: "Only items published after this date (RFC3339 format)",
			},
			"before": {
				Type:        "string",
				Format:      "date-time",
				Description: "Only items published before this date (RFC3339 format)",
			},
			"max_items": {
				Type:        "number",
				Description: "Maximum number of items to return (0 = unlimited)",
			},
			"match_all": {
				Type:        "boolean",
				Description: "If true, items must match ALL criteria; if false, match ANY (default: false)",
			},
		},
		Required: []string{"feed"},
	}

	// Define output schema
	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"items": {
				Type:        "array",
				Description: "Filtered feed items",
				Items: &sdomain.Property{
					Type:        "object",
					Description: "Feed item that matches filter criteria",
				},
			},
			"total_items": {
				Type:        "integer",
				Description: "Total number of items in original feed",
			},
			"filtered_out": {
				Type:        "integer",
				Description: "Number of items filtered out",
			},
		},
		Required: []string{"items", "total_items", "filtered_out"},
	}

	builder := atools.NewToolBuilder("feed_filter", "Filter feed items based on multiple criteria including keywords, dates, authors, and categories").
		WithFunction(feedFilterExecute).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`The feed_filter tool provides powerful filtering capabilities for feed data:

Filter Types:
1. Keyword Filtering:
   - Searches in title, description, and content
   - Case-insensitive matching
   - Partial match support

2. Date Range Filtering:
   - Filter by published date (falls back to updated date)
   - Supports after/before date ranges
   - RFC3339 date format required

3. Author Filtering:
   - Matches author name field
   - Case-insensitive partial matching
   - Useful for finding posts by specific contributors

4. Category Filtering:
   - Matches against item categories/tags
   - Case-insensitive partial matching
   - Helps find topical content

Matching Modes:
- match_all=false (default): Items matching ANY criteria are included
- match_all=true: Items must match ALL specified criteria
- Date filters are always applied regardless of matching mode

State Integration:
- feed_filter_max_items: Default maximum items limit
- feed_filter_match_all: Default matching mode

Common Use Cases:
- Recent content: Filter by date range
- Topic search: Filter by keywords and categories
- Author archives: Filter by specific authors
- Content curation: Combine multiple filters
- Feed sampling: Use max_items to limit results`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Filter by keywords",
				Description: "Find items containing specific keywords",
				Scenario:    "When searching for content about specific topics",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title":       "AI Revolution in Healthcare",
								"description": "How artificial intelligence is transforming medicine",
								"published":   "2024-03-15T10:00:00Z",
							},
							{
								"title":       "Climate Change Report",
								"description": "Latest findings on global warming",
								"published":   "2024-03-14T10:00:00Z",
							},
							{
								"title":     "Tech Stocks Analysis",
								"content":   "Deep dive into artificial intelligence companies",
								"published": "2024-03-13T10:00:00Z",
							},
						},
					},
					"keywords": []string{"artificial intelligence", "AI"},
				},
				Output: map[string]interface{}{
					"items": []map[string]interface{}{
						{
							"title":       "AI Revolution in Healthcare",
							"description": "How artificial intelligence is transforming medicine",
							"published":   "2024-03-15T10:00:00Z",
						},
						{
							"title":     "Tech Stocks Analysis",
							"content":   "Deep dive into artificial intelligence companies",
							"published": "2024-03-13T10:00:00Z",
						},
					},
					"total_items":  3,
					"filtered_out": 1,
				},
				Explanation: "Found 2 items matching keywords 'artificial intelligence' or 'AI'",
			},
			{
				Name:        "Filter by date range",
				Description: "Get items from the last week",
				Scenario:    "When you need recent content only",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title":     "Today's News",
								"published": "2024-03-20T10:00:00Z",
							},
							{
								"title":     "Last Week's Update",
								"published": "2024-03-10T10:00:00Z",
							},
							{
								"title":     "Old Article",
								"published": "2024-01-01T10:00:00Z",
							},
						},
					},
					"after": "2024-03-15T00:00:00Z",
				},
				Output: map[string]interface{}{
					"items": []map[string]interface{}{
						{
							"title":     "Today's News",
							"published": "2024-03-20T10:00:00Z",
						},
					},
					"total_items":  3,
					"filtered_out": 2,
				},
				Explanation: "Filtered to items published after March 15, 2024",
			},
			{
				Name:        "Filter by author",
				Description: "Find posts by specific authors",
				Scenario:    "When looking for content from particular contributors",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title": "Post by John",
								"author": map[string]interface{}{
									"name": "John Smith",
								},
							},
							{
								"title": "Post by Jane",
								"author": map[string]interface{}{
									"name": "Jane Doe",
								},
							},
							{
								"title": "Another John Post",
								"author": map[string]interface{}{
									"name": "John Williams",
								},
							},
						},
					},
					"authors": []string{"John"},
				},
				Output: map[string]interface{}{
					"items": []map[string]interface{}{
						{
							"title": "Post by John",
							"author": map[string]interface{}{
								"name": "John Smith",
							},
						},
						{
							"title": "Another John Post",
							"author": map[string]interface{}{
								"name": "John Williams",
							},
						},
					},
					"total_items":  3,
					"filtered_out": 1,
				},
				Explanation: "Found 2 items with authors containing 'John'",
			},
			{
				Name:        "Filter by categories",
				Description: "Find items in specific categories",
				Scenario:    "When filtering content by topic tags",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title":      "Tech News",
								"categories": []string{"technology", "news"},
							},
							{
								"title":      "Sports Update",
								"categories": []string{"sports", "news"},
							},
							{
								"title":      "Tech Tutorial",
								"categories": []string{"technology", "tutorial"},
							},
						},
					},
					"categories": []string{"tech"},
				},
				Output: map[string]interface{}{
					"items": []map[string]interface{}{
						{
							"title":      "Tech News",
							"categories": []string{"technology", "news"},
						},
						{
							"title":      "Tech Tutorial",
							"categories": []string{"technology", "tutorial"},
						},
					},
					"total_items":  3,
					"filtered_out": 1,
				},
				Explanation: "Found 2 items with categories containing 'tech'",
			},
			{
				Name:        "Complex filter with match_all",
				Description: "Apply multiple filters with ALL matching",
				Scenario:    "When you need items that meet all specified criteria",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"title":      "AI in Tech by John",
								"content":    "Artificial intelligence applications",
								"categories": []string{"technology", "ai"},
								"author":     map[string]interface{}{"name": "John Smith"},
								"published":  "2024-03-20T10:00:00Z",
							},
							{
								"title":      "AI News",
								"content":    "Latest AI developments",
								"categories": []string{"news"},
								"author":     map[string]interface{}{"name": "Jane Doe"},
								"published":  "2024-03-20T10:00:00Z",
							},
							{
								"title":      "Old Tech Post by John",
								"content":    "Technology trends",
								"categories": []string{"technology"},
								"author":     map[string]interface{}{"name": "John Smith"},
								"published":  "2024-01-01T10:00:00Z",
							},
						},
					},
					"keywords":   []string{"AI"},
					"categories": []string{"tech"},
					"authors":    []string{"John"},
					"after":      "2024-03-01T00:00:00Z",
					"match_all":  true,
				},
				Output: map[string]interface{}{
					"items": []map[string]interface{}{
						{
							"title":      "AI in Tech by John",
							"content":    "Artificial intelligence applications",
							"categories": []string{"technology", "ai"},
							"author":     map[string]interface{}{"name": "John Smith"},
							"published":  "2024-03-20T10:00:00Z",
						},
					},
					"total_items":  3,
					"filtered_out": 2,
				},
				Explanation: "Found 1 item matching ALL criteria: AI keyword, tech category, John author, and after March 1",
			},
			{
				Name:        "Date range with before and after",
				Description: "Filter items within a specific date range",
				Scenario:    "When you need content from a specific time period",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{"title": "January Post", "published": "2024-01-15T10:00:00Z"},
							{"title": "February Post", "published": "2024-02-15T10:00:00Z"},
							{"title": "March Post", "published": "2024-03-15T10:00:00Z"},
							{"title": "April Post", "published": "2024-04-15T10:00:00Z"},
						},
					},
					"after":  "2024-02-01T00:00:00Z",
					"before": "2024-04-01T00:00:00Z",
				},
				Output: map[string]interface{}{
					"items": []map[string]interface{}{
						{"title": "February Post", "published": "2024-02-15T10:00:00Z"},
						{"title": "March Post", "published": "2024-03-15T10:00:00Z"},
					},
					"total_items":  4,
					"filtered_out": 2,
				},
				Explanation: "Filtered to items published between February 1 and April 1, 2024",
			},
			{
				Name:        "Limited results",
				Description: "Filter with max_items limit",
				Scenario:    "When you want to limit the number of results",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{"title": "News 1", "categories": []string{"news"}},
							{"title": "News 2", "categories": []string{"news"}},
							{"title": "News 3", "categories": []string{"news"}},
							{"title": "News 4", "categories": []string{"news"}},
							{"title": "News 5", "categories": []string{"news"}},
						},
					},
					"categories": []string{"news"},
					"max_items":  3,
				},
				Output: map[string]interface{}{
					"items": []map[string]interface{}{
						{"title": "News 1", "categories": []string{"news"}},
						{"title": "News 2", "categories": []string{"news"}},
						{"title": "News 3", "categories": []string{"news"}},
					},
					"total_items":  5,
					"filtered_out": 0,
				},
				Explanation: "Limited results to first 3 matching items",
			},
			{
				Name:        "No date items handling",
				Description: "Handle items without dates gracefully",
				Scenario:    "When feed items lack date information",
				Input: map[string]interface{}{
					"feed": map[string]interface{}{
						"items": []map[string]interface{}{
							{"title": "Dated Post", "published": "2024-03-15T10:00:00Z"},
							{"title": "Undated Post"},
							{"title": "Another Dated", "published": "2024-03-16T10:00:00Z"},
						},
					},
					"after": "2024-03-14T00:00:00Z",
				},
				Output: map[string]interface{}{
					"items": []map[string]interface{}{
						{"title": "Dated Post", "published": "2024-03-15T10:00:00Z"},
						{"title": "Another Dated", "published": "2024-03-16T10:00:00Z"},
					},
					"total_items":  3,
					"filtered_out": 1,
				},
				Explanation: "Items without dates are excluded when date filters are applied",
			},
		}).
		WithConstraints([]string{
			"Feed must be in UnifiedFeed format",
			"All string matching is case-insensitive",
			"Partial matches are supported for keywords, authors, and categories",
			"Date filters use RFC3339 format",
			"Items without dates are excluded when date filters are applied",
			"MaxItems of 0 means no limit",
			"Empty filter arrays are ignored",
			"Date filters are always applied regardless of match mode",
		}).
		WithErrorGuidance(map[string]string{
			"invalid after date format":  "Use RFC3339 format: 2024-03-15T10:00:00Z",
			"invalid before date format": "Use RFC3339 format: 2024-03-15T10:00:00Z",
			"no feed provided":           "Feed parameter is required",
			"invalid feed format":        "Feed must be a valid UnifiedFeed object",
		}).
		WithCategory("feed").
		WithTags([]string{"feed", "filter", "search", "query", "date", "keyword", "content"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "low") // Deterministic, local operation

	return builder.Build()
}

func filterFeed(params FeedFilterParams) (*FeedFilterResult, error) {
	// Parse date filters
	var afterTime, beforeTime *time.Time
	if params.After != "" {
		t, err := time.Parse(time.RFC3339, params.After)
		if err != nil {
			return nil, fmt.Errorf("invalid after date format: %w", err)
		}
		afterTime = &t
	}
	if params.Before != "" {
		t, err := time.Parse(time.RFC3339, params.Before)
		if err != nil {
			return nil, fmt.Errorf("invalid before date format: %w", err)
		}
		beforeTime = &t
	}

	// Normalize filter criteria for case-insensitive matching
	keywordsLower := make([]string, len(params.Keywords))
	for i, kw := range params.Keywords {
		keywordsLower[i] = strings.ToLower(kw)
	}

	authorsLower := make([]string, len(params.Authors))
	for i, auth := range params.Authors {
		authorsLower[i] = strings.ToLower(auth)
	}

	categoriesLower := make([]string, len(params.Categories))
	for i, cat := range params.Categories {
		categoriesLower[i] = strings.ToLower(cat)
	}

	// Filter items
	filteredItems := make([]FeedItem, 0)
	totalItems := len(params.Feed.Items)

	for _, item := range params.Feed.Items {
		// Check if item matches filters
		if params.MatchAll {
			// All criteria must match
			if !matchesAllCriteria(item, keywordsLower, authorsLower, categoriesLower, afterTime, beforeTime) {
				continue
			}
		} else {
			// Any criteria can match (but date filters are always applied)
			if !matchesAnyCriteria(item, keywordsLower, authorsLower, categoriesLower, afterTime, beforeTime) {
				continue
			}
		}

		filteredItems = append(filteredItems, item)

		// Check max items limit
		if params.MaxItems > 0 && len(filteredItems) >= params.MaxItems {
			break
		}
	}

	return &FeedFilterResult{
		Items:       filteredItems,
		TotalItems:  totalItems,
		FilteredOut: totalItems - len(filteredItems),
	}, nil
}

func matchesAllCriteria(item FeedItem, keywords, authors, categories []string, after, before *time.Time) bool {
	// Date filters
	if !matchesDateFilters(item, after, before) {
		return false
	}

	// Keywords (if specified, must match at least one)
	if len(keywords) > 0 && !matchesKeywords(item, keywords) {
		return false
	}

	// Authors (if specified, must match)
	if len(authors) > 0 && !matchesAuthors(item, authors) {
		return false
	}

	// Categories (if specified, must match at least one)
	if len(categories) > 0 && !matchesCategories(item, categories) {
		return false
	}

	return true
}

func matchesAnyCriteria(item FeedItem, keywords, authors, categories []string, after, before *time.Time) bool {
	// Date filters are always applied
	if !matchesDateFilters(item, after, before) {
		return false
	}

	// If no other filters specified, date match is enough
	if len(keywords) == 0 && len(authors) == 0 && len(categories) == 0 {
		return true
	}

	// Check other criteria - any match is sufficient
	if len(keywords) > 0 && matchesKeywords(item, keywords) {
		return true
	}

	if len(authors) > 0 && matchesAuthors(item, authors) {
		return true
	}

	if len(categories) > 0 && matchesCategories(item, categories) {
		return true
	}

	return false
}

func matchesDateFilters(item FeedItem, after, before *time.Time) bool {
	// Use published date if available, otherwise use updated date
	var itemDate *time.Time
	if item.Published != nil {
		itemDate = item.Published
	} else if item.Updated != nil {
		itemDate = item.Updated
	}

	if itemDate == nil {
		// No date available, can't filter by date
		return after == nil && before == nil
	}

	if after != nil && itemDate.Before(*after) {
		return false
	}

	if before != nil && itemDate.After(*before) {
		return false
	}

	return true
}

func matchesKeywords(item FeedItem, keywords []string) bool {
	// Search in title, description, and content
	searchText := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)

	for _, keyword := range keywords {
		if strings.Contains(searchText, keyword) {
			return true
		}
	}

	return false
}

func matchesAuthors(item FeedItem, authors []string) bool {
	if item.Author == nil {
		return false
	}

	authorName := strings.ToLower(item.Author.Name)
	for _, author := range authors {
		if strings.Contains(authorName, author) {
			return true
		}
	}

	return false
}

func matchesCategories(item FeedItem, categories []string) bool {
	if len(item.Categories) == 0 {
		return false
	}

	for _, itemCat := range item.Categories {
		itemCatLower := strings.ToLower(itemCat)
		for _, filterCat := range categories {
			if strings.Contains(itemCatLower, filterCat) {
				return true
			}
		}
	}

	return false
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("feed_filter", FeedFilter(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_filter",
			Category:    "feed",
			Tags:        []string{"feed", "filter", "search", "query", "date", "keyword", "content"},
			Description: "Filter feed items based on multiple criteria including keywords, dates, authors, and categories",
			Version:     "2.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Filter by keywords",
					Description: "Find items containing specific keywords",
					Code:        `FeedFilter().Execute(ctx, FeedFilterParams{Feed: feed, Keywords: []string{"technology", "innovation"}})`,
				},
				{
					Name:        "Filter by date range",
					Description: "Get items from the last week",
					Code:        `FeedFilter().Execute(ctx, FeedFilterParams{Feed: feed, After: "2024-01-01T00:00:00Z", MaxItems: 50})`,
				},
				{
					Name:        "Complex filter",
					Description: "Filter by multiple criteria with all conditions matching",
					Code:        `FeedFilter().Execute(ctx, FeedFilterParams{Feed: feed, Keywords: []string{"AI"}, Categories: []string{"tech"}, Authors: []string{"John"}, MatchAll: true})`,
				},
				{
					Name:        "Filter by author",
					Description: "Find posts by specific authors",
					Code:        `FeedFilter().Execute(ctx, FeedFilterParams{Feed: feed, Authors: []string{"Smith", "Johnson"}})`,
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
		UsageInstructions: `The feed_filter tool provides powerful filtering capabilities:
- Keyword filtering: Searches in title, description, and content (case-insensitive)
- Date filtering: Filter by published/updated date ranges
- Author filtering: Match by author name (partial match)
- Category filtering: Match by categories/tags
- Matching modes: ANY (default) or ALL criteria
- Result limiting with max_items

Perfect for content curation, topic searches, and feed analysis.`,
		Constraints: []string{
			"Feed must be UnifiedFeed format",
			"String matching is case-insensitive",
			"Partial matches supported",
			"Date filters use RFC3339",
			"MaxItems 0 = unlimited",
		},
		ErrorGuidance: map[string]string{
			"invalid date format": "Use RFC3339: 2024-03-15T10:00:00Z",
			"no feed provided":    "Feed parameter is required",
			"invalid feed format": "Must be UnifiedFeed object",
		},
		IsDeterministic:      true, // Local operation
		IsDestructive:        false,
		RequiresConfirmation: false,
		EstimatedLatency:     "low",
	})
}
