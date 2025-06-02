// ABOUTME: Tests for the feed tools example
// ABOUTME: Ensures the example code compiles and basic functionality works

package main

import (
	"testing"
	"time"
)

func TestCreateMockFeedResult(t *testing.T) {
	result := createMockFeedResult()

	// Check that result has expected structure
	feed, ok := result["feed"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected feed in result")
	}

	// Check feed has title
	title, ok := feed["title"].(string)
	if !ok || title == "" {
		t.Fatal("Expected feed to have title")
	}

	// Check feed has items
	items, ok := feed["items"].([]interface{})
	if !ok || len(items) == 0 {
		t.Fatal("Expected feed to have items")
	}

	// Check first item structure
	firstItem, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected first item to be a map")
	}

	// Verify required fields
	requiredFields := []string{"id", "title", "link", "published"}
	for _, field := range requiredFields {
		if _, ok := firstItem[field]; !ok {
			t.Errorf("Expected first item to have field: %s", field)
		}
	}
}

func TestCreateMockDiscoverResult(t *testing.T) {
	result := createMockDiscoverResult()

	// Check that result has feeds
	feeds, ok := result["feeds"].([]interface{})
	if !ok || len(feeds) == 0 {
		t.Fatal("Expected discover result to have feeds")
	}

	// Check first feed structure
	firstFeed, ok := feeds[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected first feed to be a map")
	}

	// Verify required fields
	requiredFields := []string{"url", "type", "title"}
	for _, field := range requiredFields {
		if _, ok := firstFeed[field]; !ok {
			t.Errorf("Expected first feed to have field: %s", field)
		}
	}

	// Check counts
	if total, ok := result["total_discovered"].(int); !ok || total != 3 {
		t.Error("Expected total_discovered to be 3")
	}
}

func TestMockDataTimeFormats(t *testing.T) {
	result := createMockFeedResult()
	feed := result["feed"].(map[string]interface{})

	// Check that updated time can be parsed
	if updatedStr, ok := feed["updated"].(string); ok {
		if _, err := time.Parse(time.RFC3339, updatedStr); err != nil {
			t.Errorf("Failed to parse updated time: %v", err)
		}
	}

	// Check item published dates
	items := feed["items"].([]interface{})
	for i, item := range items {
		itemMap := item.(map[string]interface{})
		if publishedStr, ok := itemMap["published"].(string); ok {
			if _, err := time.Parse(time.RFC3339, publishedStr); err != nil {
				t.Errorf("Failed to parse item %d published time: %v", i, err)
			}
		}
	}
}
