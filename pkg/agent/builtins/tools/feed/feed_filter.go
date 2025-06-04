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

// feedFilterParamSchema defines parameters for the FeedFilter tool
var feedFilterParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"feed": {
			Type:        "object",
			Description: "Feed data to filter (from FeedFetch)",
		},
		"keywords": {
			Type:        "array",
			Description: "Keywords to match in title/content (case-insensitive)",
		},
		"authors": {
			Type:        "array",
			Description: "Filter by author names",
		},
		"categories": {
			Type:        "array",
			Description: "Filter by categories",
		},
		"after": {
			Type:        "string",
			Format:      "date-time",
			Description: "Only items published after this date (ISO 8601 format)",
		},
		"before": {
			Type:        "string",
			Format:      "date-time",
			Description: "Only items published before this date (ISO 8601 format)",
		},
		"max_items": {
			Type:        "number",
			Description: "Maximum number of items to return",
		},
		"match_all": {
			Type:        "boolean",
			Description: "If true, items must match ALL criteria (default: false, matches ANY)",
		},
	},
	Required: []string{"feed"},
}

func init() {
	tools.MustRegisterTool("feed_filter", FeedFilter(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "feed_filter",
			Category:    "feed",
			Tags:        []string{"feed", "filter", "search", "query", "date", "keyword"},
			Description: "Filters feed items by date, keywords, author, and categories",
			Version:     "1.0.0",
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

// FeedFilter creates a new FeedFilter tool
func FeedFilter() domain.Tool {
	return atools.NewTool(
		"feed_filter",
		"Filters feed items based on date, keywords, author, and categories",
		func(ctx *domain.ToolContext, params FeedFilterParams) (*FeedFilterResult, error) {
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
		},
		feedFilterParamSchema,
	)
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
