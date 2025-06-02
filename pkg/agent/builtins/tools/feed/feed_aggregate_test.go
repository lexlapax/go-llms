// ABOUTME: Tests for the FeedAggregate tool that combines multiple feeds
// ABOUTME: Tests aggregation, sorting, deduplication, and metadata merging

package feed

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestFeedAggregateRegistration(t *testing.T) {
	tool := FeedAggregate()

	if tool.Name() != "feed_aggregate" {
		t.Errorf("Expected tool name 'feed_aggregate', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Tool description should not be empty")
	}
}

func TestFeedAggregateBasic(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	feed1 := UnifiedFeed{
		Title: "Feed 1",
		Items: []FeedItem{
			{
				ID:        "1-1",
				Title:     "Article 1 from Feed 1",
				Published: &now,
			},
			{
				ID:        "1-2",
				Title:     "Article 2 from Feed 1",
				Published: &yesterday,
			},
		},
	}

	feed2 := UnifiedFeed{
		Title: "Feed 2",
		Items: []FeedItem{
			{
				ID:        "2-1",
				Title:     "Article 1 from Feed 2",
				Published: &now,
			},
		},
	}

	tool := FeedAggregate()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"feeds":          []UnifiedFeed{feed1, feed2},
		"merge_metadata": true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	aggResult, ok := result.(*FeedAggregateResult)
	if !ok {
		t.Fatalf("Expected *FeedAggregateResult, got %T", result)
	}

	if len(aggResult.Feed.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(aggResult.Feed.Items))
	}

	if aggResult.SourceCount != 2 {
		t.Errorf("Expected 2 source feeds, got %d", aggResult.SourceCount)
	}

	if aggResult.TotalItems != 3 {
		t.Errorf("Expected 3 total items, got %d", aggResult.TotalItems)
	}

	// Check metadata merging (default)
	if aggResult.Feed.Title != "Aggregated: Feed 1, Feed 2" {
		t.Errorf("Expected aggregated title, got '%s'", aggResult.Feed.Title)
	}
}

func TestFeedAggregateSortByDate(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastWeek := now.AddDate(0, 0, -7)

	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:        "2",
				Title:     "Yesterday's Article",
				Published: &yesterday,
			},
			{
				ID:        "3",
				Title:     "Last Week's Article",
				Published: &lastWeek,
			},
			{
				ID:        "1",
				Title:     "Today's Article",
				Published: &now,
			},
		},
	}

	tool := FeedAggregate()
	ctx := context.Background()

	// Test ascending date sort (default is ascending now)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feeds":   []UnifiedFeed{feed},
		"sort_by": "date",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	aggResult, ok := result.(*FeedAggregateResult)
	if !ok {
		t.Fatalf("Expected *FeedAggregateResult, got %T", result)
	}

	// Should be sorted oldest first (ascending is default)
	if aggResult.Feed.Items[0].ID != "3" {
		t.Errorf("Expected first item to be last week's (ID=3), got ID=%s", aggResult.Feed.Items[0].ID)
	}
	if aggResult.Feed.Items[1].ID != "2" {
		t.Errorf("Expected second item to be yesterday's (ID=2), got ID=%s", aggResult.Feed.Items[1].ID)
	}
	if aggResult.Feed.Items[2].ID != "1" {
		t.Errorf("Expected third item to be today's (ID=1), got ID=%s", aggResult.Feed.Items[2].ID)
	}

	// Test descending date sort
	result2, err := tool.Execute(ctx, map[string]interface{}{
		"feeds":           []UnifiedFeed{feed},
		"sort_by":         "date",
		"sort_descending": true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	aggResult2, ok := result2.(*FeedAggregateResult)
	if !ok {
		t.Fatalf("Expected *FeedAggregateResult, got %T", result2)
	}

	// Should be sorted newest first
	if len(aggResult2.Feed.Items) < 3 {
		t.Fatalf("Expected 3 items, got %d", len(aggResult2.Feed.Items))
	}
	if aggResult2.Feed.Items[0].ID != "1" {
		t.Errorf("Expected first item to be today's (ID=1), got ID=%s", aggResult2.Feed.Items[0].ID)
	}
	if aggResult2.Feed.Items[1].ID != "2" {
		t.Errorf("Expected second item to be yesterday's (ID=2), got ID=%s", aggResult2.Feed.Items[1].ID)
	}
	if aggResult2.Feed.Items[2].ID != "3" {
		t.Errorf("Expected third item to be last week's (ID=3), got ID=%s", aggResult2.Feed.Items[2].ID)
	}
}

func TestFeedAggregateSortByTitle(t *testing.T) {
	now := time.Now()

	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:        "2",
				Title:     "Bravo Article",
				Published: &now,
			},
			{
				ID:        "3",
				Title:     "Charlie Article",
				Published: &now,
			},
			{
				ID:        "1",
				Title:     "Alpha Article",
				Published: &now,
			},
		},
	}

	tool := FeedAggregate()
	ctx := context.Background()

	// Test ascending title sort (default for title)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feeds":   []UnifiedFeed{feed},
		"sort_by": "title",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	aggResult, ok := result.(*FeedAggregateResult)
	if !ok {
		t.Fatalf("Expected *FeedAggregateResult, got %T", result)
	}

	// Should be sorted alphabetically
	if aggResult.Feed.Items[0].ID != "1" {
		t.Errorf("Expected first item to be Alpha (ID=1), got ID=%s", aggResult.Feed.Items[0].ID)
	}
	if aggResult.Feed.Items[1].ID != "2" {
		t.Errorf("Expected second item to be Bravo (ID=2), got ID=%s", aggResult.Feed.Items[1].ID)
	}
	if aggResult.Feed.Items[2].ID != "3" {
		t.Errorf("Expected third item to be Charlie (ID=3), got ID=%s", aggResult.Feed.Items[2].ID)
	}
}

func TestFeedAggregateRemoveDuplicates(t *testing.T) {
	now := time.Now()

	feed1 := UnifiedFeed{
		Title: "Feed 1",
		Items: []FeedItem{
			{
				ID:        "1",
				Title:     "Article 1",
				Link:      "https://example.com/article1",
				Published: &now,
			},
			{
				ID:        "2",
				Title:     "Article 2",
				Link:      "https://example.com/article2",
				Published: &now,
			},
		},
	}

	feed2 := UnifiedFeed{
		Title: "Feed 2",
		Items: []FeedItem{
			{
				ID:        "1-dup",
				Title:     "Article 1 Duplicate",
				Link:      "https://example.com/article1", // Same URL
				Published: &now,
			},
			{
				ID:        "3",
				Title:     "Article 3",
				Link:      "https://example.com/article3",
				Published: &now,
			},
		},
	}

	tool := FeedAggregate()
	ctx := context.Background()

	// Test with duplicate removal
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feeds":        []UnifiedFeed{feed1, feed2},
		"remove_dupes": true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	aggResult, ok := result.(*FeedAggregateResult)
	if !ok {
		t.Fatalf("Expected *FeedAggregateResult, got %T", result)
	}

	if len(aggResult.Feed.Items) != 3 {
		t.Errorf("Expected 3 unique items, got %d", len(aggResult.Feed.Items))
	}

	if aggResult.DupesRemoved != 1 {
		t.Errorf("Expected 1 duplicate removed, got %d", aggResult.DupesRemoved)
	}

	if aggResult.TotalItems != 4 {
		t.Errorf("Expected 4 total items before dedup, got %d", aggResult.TotalItems)
	}
}

func TestFeedAggregateMaxItems(t *testing.T) {
	now := time.Now()

	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: make([]FeedItem, 10),
	}

	// Create 10 items
	for i := 0; i < 10; i++ {
		feed.Items[i] = FeedItem{
			ID:        fmt.Sprintf("%d", i+1),
			Title:     fmt.Sprintf("Article %d", i+1),
			Published: &now,
		}
	}

	tool := FeedAggregate()
	ctx := context.Background()

	// Test max items limit
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feeds":     []UnifiedFeed{feed},
		"max_items": 5,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	aggResult, ok := result.(*FeedAggregateResult)
	if !ok {
		t.Fatalf("Expected *FeedAggregateResult, got %T", result)
	}

	if len(aggResult.Feed.Items) != 5 {
		t.Errorf("Expected 5 items (max_items limit), got %d", len(aggResult.Feed.Items))
	}

	if aggResult.TotalItems != 10 {
		t.Errorf("Expected 10 total items, got %d", aggResult.TotalItems)
	}
}

func TestFeedAggregateNoMetadataMerge(t *testing.T) {
	feed1 := UnifiedFeed{
		Title:       "Feed 1",
		Description: "Description 1",
		Items:       []FeedItem{},
	}

	feed2 := UnifiedFeed{
		Title:       "Feed 2",
		Description: "Description 2",
		Items:       []FeedItem{},
	}

	tool := FeedAggregate()
	ctx := context.Background()

	mergeMetadata := false
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feeds":          []UnifiedFeed{feed1, feed2},
		"merge_metadata": mergeMetadata,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	aggResult, ok := result.(*FeedAggregateResult)
	if !ok {
		t.Fatalf("Expected *FeedAggregateResult, got %T", result)
	}

	// With merge_metadata=false, title should be empty
	if aggResult.Feed.Title != "" {
		t.Errorf("Expected empty title with merge_metadata=false, got '%s'", aggResult.Feed.Title)
	}

	if aggResult.Feed.Description != "" {
		t.Errorf("Expected empty description with merge_metadata=false, got '%s'", aggResult.Feed.Description)
	}
}

func TestFeedAggregateEmptyFeeds(t *testing.T) {
	tool := FeedAggregate()
	ctx := context.Background()

	// Test with empty feeds array
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feeds": []UnifiedFeed{},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	aggResult, ok := result.(*FeedAggregateResult)
	if !ok {
		t.Fatalf("Expected *FeedAggregateResult, got %T", result)
	}

	if len(aggResult.Feed.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(aggResult.Feed.Items))
	}

	if aggResult.SourceCount != 0 {
		t.Errorf("Expected 0 source feeds, got %d", aggResult.SourceCount)
	}
}

func TestFeedAggregateInvalidSortBy(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{},
	}

	tool := FeedAggregate()
	ctx := context.Background()

	// Test invalid sort_by field
	_, err := tool.Execute(ctx, map[string]interface{}{
		"feeds":   []UnifiedFeed{feed},
		"sort_by": "invalid",
	})
	if err == nil {
		t.Error("Expected error for invalid sort_by field")
	}
}

func TestFeedAggregateDuplicatesByContent(t *testing.T) {
	now := time.Now()

	// Create items without URLs (will use content hash)
	feed1 := UnifiedFeed{
		Title: "Feed 1",
		Items: []FeedItem{
			{
				ID:          "1",
				Title:       "Same Title",
				Description: "Same Description",
				Content:     "Same Content",
				Published:   &now,
			},
		},
	}

	feed2 := UnifiedFeed{
		Title: "Feed 2",
		Items: []FeedItem{
			{
				ID:          "2",
				Title:       "Same Title",
				Description: "Same Description",
				Content:     "Same Content",
				Published:   &now,
			},
		},
	}

	tool := FeedAggregate()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"feeds":        []UnifiedFeed{feed1, feed2},
		"remove_dupes": true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	aggResult, ok := result.(*FeedAggregateResult)
	if !ok {
		t.Fatalf("Expected *FeedAggregateResult, got %T", result)
	}

	if len(aggResult.Feed.Items) != 1 {
		t.Errorf("Expected 1 unique item (content-based dedup), got %d", len(aggResult.Feed.Items))
	}

	if aggResult.DupesRemoved != 1 {
		t.Errorf("Expected 1 duplicate removed, got %d", aggResult.DupesRemoved)
	}
}
