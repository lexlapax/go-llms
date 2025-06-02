// ABOUTME: Tests for the feed tools example
// ABOUTME: Ensures the example code compiles and basic functionality works

package main

import (
	"testing"
)

func TestCreateMockFeedFetchResult(t *testing.T) {
	result := createMockFeedFetchResult()

	// Check that result has expected structure
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Check basic fields
	if result.Status != 200 {
		t.Errorf("Expected status 200, got %d", result.Status)
	}

	if result.Format != "RSS2" {
		t.Errorf("Expected format RSS2, got %s", result.Format)
	}

	// Check feed structure
	if result.Feed.Title == "" {
		t.Fatal("Expected feed to have title")
	}

	if len(result.Feed.Items) == 0 {
		t.Fatal("Expected feed to have items")
	}

	// Check first item structure
	firstItem := result.Feed.Items[0]

	// Verify required fields
	if firstItem.ID == "" {
		t.Error("Expected first item to have ID")
	}
	if firstItem.Title == "" {
		t.Error("Expected first item to have title")
	}
	if firstItem.Link == "" {
		t.Error("Expected first item to have link")
	}
	if firstItem.Published == nil {
		t.Error("Expected first item to have published date")
	}

	// Check categories
	if len(firstItem.Categories) == 0 {
		t.Error("Expected first item to have categories")
	}

	// Check author
	if firstItem.Author == nil || firstItem.Author.Name == "" {
		t.Error("Expected first item to have author with name")
	}
}

func TestCreateMockDiscoverResult(t *testing.T) {
	result := createMockDiscoverResult()

	// Check that result has feeds
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if len(result.Feeds) == 0 {
		t.Fatal("Expected discover result to have feeds")
	}

	// Check first feed structure
	firstFeed := result.Feeds[0]

	// Verify required fields
	if firstFeed.URL == "" {
		t.Error("Expected first feed to have URL")
	}
	if firstFeed.Type == "" {
		t.Error("Expected first feed to have type")
	}
	if firstFeed.Title == "" {
		t.Error("Expected first feed to have title")
	}
	if firstFeed.Source == "" {
		t.Error("Expected first feed to have source")
	}

	// Check we have the expected number of feeds
	if len(result.Feeds) != 3 {
		t.Errorf("Expected 3 feeds, got %d", len(result.Feeds))
	}

	// Check that error field is empty for successful mock
	if result.Error != "" {
		t.Errorf("Expected no error, got: %s", result.Error)
	}
}

func TestMockDataTimeFormats(t *testing.T) {
	result := createMockFeedFetchResult()

	// Check that updated time is valid
	if result.Feed.Updated == nil {
		t.Error("Expected feed to have updated time")
	}

	// Check item published dates
	for i, item := range result.Feed.Items {
		if item.Published == nil {
			t.Errorf("Expected item %d to have published time", i)
		}
		if item.Updated == nil {
			t.Errorf("Expected item %d to have updated time", i)
		}
	}
}

func TestMockDataStructure(t *testing.T) {
	result := createMockFeedFetchResult()

	// Validate feed metadata
	if result.Feed.Language != "en" {
		t.Errorf("Expected language 'en', got '%s'", result.Feed.Language)
	}

	if result.Feed.Author == nil {
		t.Error("Expected feed to have author")
	} else {
		if result.Feed.Author.Name == "" {
			t.Error("Expected feed author to have name")
		}
		if result.Feed.Author.Email == "" {
			t.Error("Expected feed author to have email")
		}
	}

	// Validate headers
	if len(result.Headers) == 0 {
		t.Error("Expected result to have headers")
	}

	expectedHeaders := []string{"Content-Type", "Last-Modified", "ETag"}
	for _, header := range expectedHeaders {
		if _, exists := result.Headers[header]; !exists {
			t.Errorf("Expected header '%s' to be present", header)
		}
	}
}

func TestDiscoverResultTypes(t *testing.T) {
	result := createMockDiscoverResult()

	expectedTypes := []string{"RSS", "Atom", "Podcast"}
	if len(result.Feeds) != len(expectedTypes) {
		t.Fatalf("Expected %d feeds, got %d", len(expectedTypes), len(result.Feeds))
	}

	for i, expectedType := range expectedTypes {
		if result.Feeds[i].Type != expectedType {
			t.Errorf("Expected feed %d to have type '%s', got '%s'", 
				i, expectedType, result.Feeds[i].Type)
		}
	}

	expectedSources := []string{"link_tag", "auto_discovery", "common_path"}
	for i, expectedSource := range expectedSources {
		if result.Feeds[i].Source != expectedSource {
			t.Errorf("Expected feed %d to have source '%s', got '%s'", 
				i, expectedSource, result.Feeds[i].Source)
		}
	}
}