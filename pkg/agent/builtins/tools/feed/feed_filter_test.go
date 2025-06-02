// ABOUTME: Tests for the FeedFilter tool that filters feed items by various criteria
// ABOUTME: Tests keyword matching, date filtering, author filtering, and category filtering

package feed

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestFeedFilterRegistration(t *testing.T) {
	tool := FeedFilter()

	if tool.Name() != "feed_filter" {
		t.Errorf("Expected tool name 'feed_filter', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Tool description should not be empty")
	}
}

func TestFeedFilterByKeywords(t *testing.T) {
	// Create test feed
	now := time.Now()
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:          "1",
				Title:       "Introduction to Go Programming",
				Description: "Learn the basics of Go language",
				Content:     "Go is a statically typed, compiled language",
				Published:   &now,
			},
			{
				ID:          "2",
				Title:       "Python for Data Science",
				Description: "Using Python for data analysis",
				Content:     "Python has great libraries for data science",
				Published:   &now,
			},
			{
				ID:          "3",
				Title:       "JavaScript Frameworks",
				Description: "Modern JS frameworks comparison",
				Content:     "React, Vue, and Angular are popular choices",
				Published:   &now,
			},
		},
	}

	tool := FeedFilter()
	ctx := context.Background()

	// Filter for Go-related items
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feed":     feed,
		"keywords": []string{"Go", "golang"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	filterResult, ok := result.(*FeedFilterResult)
	if !ok {
		t.Fatalf("Expected *FeedFilterResult, got %T", result)
	}

	if len(filterResult.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(filterResult.Items))
	}

	if filterResult.Items[0].ID != "1" {
		t.Errorf("Expected item with ID '1', got '%s'", filterResult.Items[0].ID)
	}

	if filterResult.TotalItems != 3 {
		t.Errorf("Expected 3 total items, got %d", filterResult.TotalItems)
	}

	if filterResult.FilteredOut != 2 {
		t.Errorf("Expected 2 filtered out, got %d", filterResult.FilteredOut)
	}
}

func TestFeedFilterByDate(t *testing.T) {
	// Create test feed with different dates
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastWeek := now.AddDate(0, 0, -7)
	lastMonth := now.AddDate(0, -1, 0)

	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:        "1",
				Title:     "Today's News",
				Published: &now,
			},
			{
				ID:        "2",
				Title:     "Yesterday's News",
				Published: &yesterday,
			},
			{
				ID:        "3",
				Title:     "Last Week's News",
				Published: &lastWeek,
			},
			{
				ID:        "4",
				Title:     "Last Month's News",
				Published: &lastMonth,
			},
		},
	}

	tool := FeedFilter()
	ctx := context.Background()

	// Filter for items from the last 3 days
	threeDaysAgo := now.AddDate(0, 0, -3)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feed":  feed,
		"after": threeDaysAgo.Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	filterResult, ok := result.(*FeedFilterResult)
	if !ok {
		t.Fatalf("Expected *FeedFilterResult, got %T", result)
	}

	if len(filterResult.Items) != 2 {
		t.Errorf("Expected 2 items (today and yesterday), got %d", len(filterResult.Items))
	}

	// Test date range
	result2, err := tool.Execute(ctx, map[string]interface{}{
		"feed":   feed,
		"after":  lastWeek.AddDate(0, 0, -1).Format(time.RFC3339),
		"before": yesterday.AddDate(0, 0, 1).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	filterResult2, ok := result2.(*FeedFilterResult)
	if !ok {
		t.Fatalf("Expected *FeedFilterResult, got %T", result2)
	}

	if len(filterResult2.Items) != 2 {
		t.Errorf("Expected 2 items (yesterday and last week), got %d", len(filterResult2.Items))
	}
}

func TestFeedFilterByAuthor(t *testing.T) {
	// Create test feed with different authors
	now := time.Now()
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:        "1",
				Title:     "Article by John",
				Author:    &Author{Name: "John Doe"},
				Published: &now,
			},
			{
				ID:        "2",
				Title:     "Article by Jane",
				Author:    &Author{Name: "Jane Smith"},
				Published: &now,
			},
			{
				ID:        "3",
				Title:     "Article by Bob",
				Author:    &Author{Name: "Bob Johnson"},
				Published: &now,
			},
			{
				ID:        "4",
				Title:     "Anonymous Article",
				Published: &now,
			},
		},
	}

	tool := FeedFilter()
	ctx := context.Background()

	// Filter by author
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feed":    feed,
		"authors": []string{"John"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	filterResult, ok := result.(*FeedFilterResult)
	if !ok {
		t.Fatalf("Expected *FeedFilterResult, got %T", result)
	}

	if len(filterResult.Items) != 2 {
		t.Errorf("Expected 2 items (John Doe and Bob Johnson), got %d", len(filterResult.Items))
	}
}

func TestFeedFilterByCategories(t *testing.T) {
	// Create test feed with categories
	now := time.Now()
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:         "1",
				Title:      "Tech Article",
				Categories: []string{"Technology", "Programming"},
				Published:  &now,
			},
			{
				ID:         "2",
				Title:      "Science Article",
				Categories: []string{"Science", "Research"},
				Published:  &now,
			},
			{
				ID:         "3",
				Title:      "Tech and Science",
				Categories: []string{"Technology", "Science"},
				Published:  &now,
			},
			{
				ID:        "4",
				Title:     "No Category",
				Published: &now,
			},
		},
	}

	tool := FeedFilter()
	ctx := context.Background()

	// Filter by category
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feed":       feed,
		"categories": []string{"Technology"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	filterResult, ok := result.(*FeedFilterResult)
	if !ok {
		t.Fatalf("Expected *FeedFilterResult, got %T", result)
	}

	if len(filterResult.Items) != 2 {
		t.Errorf("Expected 2 items with Technology category, got %d", len(filterResult.Items))
	}
}

func TestFeedFilterMatchAll(t *testing.T) {
	// Create test feed
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:          "1",
				Title:       "Go Programming Today",
				Description: "Learn Go language",
				Author:      &Author{Name: "John Doe"},
				Categories:  []string{"Programming", "Tutorial"},
				Published:   &now,
			},
			{
				ID:          "2",
				Title:       "Python Programming Yesterday",
				Description: "Learn Python",
				Author:      &Author{Name: "John Doe"},
				Categories:  []string{"Programming", "Tutorial"},
				Published:   &yesterday,
			},
			{
				ID:          "3",
				Title:       "Go News",
				Description: "Latest Go updates",
				Author:      &Author{Name: "Jane Smith"},
				Categories:  []string{"News"},
				Published:   &now,
			},
		},
	}

	tool := FeedFilter()
	ctx := context.Background()

	// Test match_all = true
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feed":       feed,
		"keywords":   []string{"Go"},
		"authors":    []string{"John"},
		"categories": []string{"Programming"},
		"match_all":  true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	filterResult, ok := result.(*FeedFilterResult)
	if !ok {
		t.Fatalf("Expected *FeedFilterResult, got %T", result)
	}

	// Only item 1 should match all criteria
	if len(filterResult.Items) != 1 {
		t.Errorf("Expected 1 item matching all criteria, got %d", len(filterResult.Items))
	}

	if len(filterResult.Items) > 0 && filterResult.Items[0].ID != "1" {
		t.Errorf("Expected item with ID '1', got '%s'", filterResult.Items[0].ID)
	}
}

func TestFeedFilterMaxItems(t *testing.T) {
	// Create test feed with many items
	now := time.Now()
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: make([]FeedItem, 10),
	}

	for i := 0; i < 10; i++ {
		feed.Items[i] = FeedItem{
			ID:        fmt.Sprintf("%d", i+1),
			Title:     fmt.Sprintf("Article %d", i+1),
			Published: &now,
		}
	}

	tool := FeedFilter()
	ctx := context.Background()

	// Test max_items limit
	result, err := tool.Execute(ctx, map[string]interface{}{
		"feed":      feed,
		"max_items": 3,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	filterResult, ok := result.(*FeedFilterResult)
	if !ok {
		t.Fatalf("Expected *FeedFilterResult, got %T", result)
	}

	if len(filterResult.Items) != 3 {
		t.Errorf("Expected 3 items (max_items limit), got %d", len(filterResult.Items))
	}

	if filterResult.TotalItems != 10 {
		t.Errorf("Expected 10 total items, got %d", filterResult.TotalItems)
	}
}

func TestFeedFilterEmptyFeed(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Empty Feed",
		Items: []FeedItem{},
	}

	tool := FeedFilter()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"feed":     feed,
		"keywords": []string{"test"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	filterResult, ok := result.(*FeedFilterResult)
	if !ok {
		t.Fatalf("Expected *FeedFilterResult, got %T", result)
	}

	if len(filterResult.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(filterResult.Items))
	}

	if filterResult.TotalItems != 0 {
		t.Errorf("Expected 0 total items, got %d", filterResult.TotalItems)
	}
}

func TestFeedFilterInvalidDateFormat(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:    "1",
				Title: "Test Article",
			},
		},
	}

	tool := FeedFilter()
	ctx := context.Background()

	// Test invalid after date
	_, err := tool.Execute(ctx, map[string]interface{}{
		"feed":  feed,
		"after": "not-a-date",
	})
	if err == nil {
		t.Error("Expected error for invalid date format")
	}

	// Test invalid before date
	_, err = tool.Execute(ctx, map[string]interface{}{
		"feed":   feed,
		"before": "2024-13-45", // Invalid date
	})
	if err == nil {
		t.Error("Expected error for invalid date format")
	}
}
