// ABOUTME: Tests for the FeedExtract tool that extracts specific fields from feeds
// ABOUTME: Tests field extraction, flattening, metadata inclusion, and error handling

package feed

import (
	"strings"
	"testing"
	"time"
)

func TestFeedExtractRegistration(t *testing.T) {
	tool := FeedExtract()

	if tool.Name() != "feed_extract" {
		t.Errorf("Expected tool name 'feed_extract', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Tool description should not be empty")
	}
}

func TestFeedExtractBasicFields(t *testing.T) {
	now := time.Now()
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:          "1",
				Title:       "Article 1",
				Link:        "https://example.com/1",
				Description: "Description 1",
				Published:   &now,
			},
			{
				ID:          "2",
				Title:       "Article 2",
				Link:        "https://example.com/2",
				Description: "Description 2",
				Published:   &now,
			},
		},
	}

	tool := FeedExtract()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"feed":   feed,
		"fields": []string{"id", "title", "link"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	extractResult, ok := result.(*FeedExtractResult)
	if !ok {
		t.Fatalf("Expected *FeedExtractResult, got %T", result)
	}

	if len(extractResult.Data) != 2 {
		t.Errorf("Expected 2 extracted items, got %d", len(extractResult.Data))
	}

	if extractResult.Count != 2 {
		t.Errorf("Expected count 2, got %d", extractResult.Count)
	}

	// Check first item
	if len(extractResult.Data) > 0 {
		first := extractResult.Data[0]
		if first["id"] != "1" {
			t.Errorf("Expected id '1', got '%v'", first["id"])
		}
		if first["title"] != "Article 1" {
			t.Errorf("Expected title 'Article 1', got '%v'", first["title"])
		}
		if first["link"] != "https://example.com/1" {
			t.Errorf("Expected link 'https://example.com/1', got '%v'", first["link"])
		}
	}

	// Check fields list
	if len(extractResult.Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(extractResult.Fields))
	}
}

func TestFeedExtractNestedFields(t *testing.T) {
	now := time.Now()
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:    "1",
				Title: "Article 1",
				Author: &Author{
					Name:  "John Doe",
					Email: "john@example.com",
					URL:   "https://johndoe.com",
				},
				Published: &now,
			},
		},
	}

	tool := FeedExtract()
	tc := createTestToolContext()

	// Test nested field extraction without flattening
	result, err := tool.Execute(tc, map[string]interface{}{
		"feed":   feed,
		"fields": []string{"title", "author.name", "author.email"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	extractResult, ok := result.(*FeedExtractResult)
	if !ok {
		t.Fatalf("Expected *FeedExtractResult, got %T", result)
	}

	if len(extractResult.Data) != 1 {
		t.Errorf("Expected 1 extracted item, got %d", len(extractResult.Data))
	}

	// Check nested field values
	if len(extractResult.Data) > 0 {
		first := extractResult.Data[0]
		if first["author.name"] != "John Doe" {
			t.Errorf("Expected author.name 'John Doe', got '%v'", first["author.name"])
		}
		if first["author.email"] != "john@example.com" {
			t.Errorf("Expected author.email 'john@example.com', got '%v'", first["author.email"])
		}
	}
}

func TestFeedExtractWithFlattening(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:    "1",
				Title: "Article 1",
				Author: &Author{
					Name: "Jane Doe",
				},
			},
		},
	}

	tool := FeedExtract()
	tc := createTestToolContext()

	// Test with flattening enabled
	result, err := tool.Execute(tc, map[string]interface{}{
		"feed":    feed,
		"fields":  []string{"title", "author.name"},
		"flatten": true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	extractResult, ok := result.(*FeedExtractResult)
	if !ok {
		t.Fatalf("Expected *FeedExtractResult, got %T", result)
	}

	// Check that nested field is flattened
	if len(extractResult.Data) > 0 {
		first := extractResult.Data[0]
		if _, exists := first["author_name"]; !exists {
			t.Error("Expected flattened field 'author_name' to exist")
		}
		if first["author_name"] != "Jane Doe" {
			t.Errorf("Expected author_name 'Jane Doe', got '%v'", first["author_name"])
		}
	}
}

func TestFeedExtractWithMetadata(t *testing.T) {
	now := time.Now()
	feed := UnifiedFeed{
		Title:       "Test Feed",
		Description: "A test feed",
		Link:        "https://example.com/feed",
		Language:    "en",
		Updated:     &now,
		Author: &Author{
			Name: "Feed Author",
		},
		Items: []FeedItem{
			{
				ID:    "1",
				Title: "Article 1",
			},
		},
	}

	tool := FeedExtract()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"feed":             feed,
		"fields":           []string{"title"},
		"include_metadata": true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	extractResult, ok := result.(*FeedExtractResult)
	if !ok {
		t.Fatalf("Expected *FeedExtractResult, got %T", result)
	}

	if extractResult.Metadata == nil {
		t.Fatal("Expected metadata to be included")
	}

	// Check metadata fields
	if extractResult.Metadata["title"] != "Test Feed" {
		t.Errorf("Expected metadata title 'Test Feed', got '%v'", extractResult.Metadata["title"])
	}
	if extractResult.Metadata["description"] != "A test feed" {
		t.Errorf("Expected metadata description 'A test feed', got '%v'", extractResult.Metadata["description"])
	}
	if extractResult.Metadata["language"] != "en" {
		t.Errorf("Expected metadata language 'en', got '%v'", extractResult.Metadata["language"])
	}
}

func TestFeedExtractMaxItems(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: make([]FeedItem, 10),
	}

	// Create 10 items
	for i := 0; i < 10; i++ {
		feed.Items[i] = FeedItem{
			ID:    string(rune('0' + i)),
			Title: "Article",
		}
	}

	tool := FeedExtract()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"feed":      feed,
		"fields":    []string{"id", "title"},
		"max_items": 5,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	extractResult, ok := result.(*FeedExtractResult)
	if !ok {
		t.Fatalf("Expected *FeedExtractResult, got %T", result)
	}

	if len(extractResult.Data) != 5 {
		t.Errorf("Expected 5 extracted items (max_items limit), got %d", len(extractResult.Data))
	}

	if extractResult.Count != 5 {
		t.Errorf("Expected count 5, got %d", extractResult.Count)
	}
}

func TestFeedExtractDateFields(t *testing.T) {
	published := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)

	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:        "1",
				Title:     "Article 1",
				Published: &published,
				Updated:   &updated,
			},
		},
	}

	tool := FeedExtract()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"feed":   feed,
		"fields": []string{"title", "published", "updated"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	extractResult, ok := result.(*FeedExtractResult)
	if !ok {
		t.Fatalf("Expected *FeedExtractResult, got %T", result)
	}

	if len(extractResult.Data) > 0 {
		first := extractResult.Data[0]

		// Check date formatting
		publishedStr, ok := first["published"].(string)
		if !ok {
			t.Error("Expected published to be a string")
		}
		if !strings.Contains(publishedStr, "2024-01-15") {
			t.Errorf("Expected published date to contain '2024-01-15', got '%s'", publishedStr)
		}

		updatedStr, ok := first["updated"].(string)
		if !ok {
			t.Error("Expected updated to be a string")
		}
		if !strings.Contains(updatedStr, "2024-01-16") {
			t.Errorf("Expected updated date to contain '2024-01-16', got '%s'", updatedStr)
		}
	}
}

func TestFeedExtractArrayFields(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:         "1",
				Title:      "Article 1",
				Categories: []string{"tech", "news", "ai"},
				Enclosures: []Enclosure{
					{
						URL:    "https://example.com/image.jpg",
						Type:   "image/jpeg",
						Length: 12345,
					},
				},
			},
		},
	}

	tool := FeedExtract()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"feed":   feed,
		"fields": []string{"title", "categories", "enclosures"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	extractResult, ok := result.(*FeedExtractResult)
	if !ok {
		t.Fatalf("Expected *FeedExtractResult, got %T", result)
	}

	if len(extractResult.Data) > 0 {
		first := extractResult.Data[0]

		// Check categories
		categories, ok := first["categories"].([]string)
		if !ok {
			t.Error("Expected categories to be []string")
		} else if len(categories) != 3 {
			t.Errorf("Expected 3 categories, got %d", len(categories))
		}

		// Check enclosures
		enclosures, ok := first["enclosures"].([]map[string]interface{})
		if !ok {
			t.Error("Expected enclosures to be []map[string]interface{}")
		} else if len(enclosures) != 1 {
			t.Errorf("Expected 1 enclosure, got %d", len(enclosures))
		}
	}
}

func TestFeedExtractEmptyFeed(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Empty Feed",
		Items: []FeedItem{},
	}

	tool := FeedExtract()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"feed":   feed,
		"fields": []string{"title"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	extractResult, ok := result.(*FeedExtractResult)
	if !ok {
		t.Fatalf("Expected *FeedExtractResult, got %T", result)
	}

	if len(extractResult.Data) != 0 {
		t.Errorf("Expected 0 extracted items from empty feed, got %d", len(extractResult.Data))
	}

	if extractResult.Count != 0 {
		t.Errorf("Expected count 0, got %d", extractResult.Count)
	}
}

func TestFeedExtractNoFields(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{{ID: "1", Title: "Article 1"}},
	}

	tool := FeedExtract()
	tc := createTestToolContext()

	_, err := tool.Execute(tc, map[string]interface{}{
		"feed":   feed,
		"fields": []string{},
	})
	if err == nil {
		t.Error("Expected error for empty fields list")
	}
}

func TestFeedExtractNonExistentFields(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:    "1",
				Title: "Article 1",
			},
		},
	}

	tool := FeedExtract()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"feed":   feed,
		"fields": []string{"nonexistent", "invalid.field"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	extractResult, ok := result.(*FeedExtractResult)
	if !ok {
		t.Fatalf("Expected *FeedExtractResult, got %T", result)
	}

	// Should still succeed but with empty data for non-existent fields
	if len(extractResult.Data) > 0 {
		first := extractResult.Data[0]
		if len(first) != 0 {
			t.Errorf("Expected empty map for non-existent fields, got %d fields", len(first))
		}
	}
}

func TestFeedExtractFullAuthorObject(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:    "1",
				Title: "Article 1",
				Author: &Author{
					Name:  "John Doe",
					Email: "john@example.com",
					URL:   "https://johndoe.com",
				},
			},
		},
	}

	tool := FeedExtract()
	tc := createTestToolContext()

	// Request full author object (no sub-field)
	result, err := tool.Execute(tc, map[string]interface{}{
		"feed":   feed,
		"fields": []string{"title", "author"},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	extractResult, ok := result.(*FeedExtractResult)
	if !ok {
		t.Fatalf("Expected *FeedExtractResult, got %T", result)
	}

	if len(extractResult.Data) > 0 {
		first := extractResult.Data[0]
		author, ok := first["author"].(map[string]string)
		if !ok {
			t.Error("Expected author to be map[string]string")
		} else {
			if author["name"] != "John Doe" {
				t.Errorf("Expected author name 'John Doe', got '%s'", author["name"])
			}
			if author["email"] != "john@example.com" {
				t.Errorf("Expected author email 'john@example.com', got '%s'", author["email"])
			}
			if author["url"] != "https://johndoe.com" {
				t.Errorf("Expected author url 'https://johndoe.com', got '%s'", author["url"])
			}
		}
	}
}
