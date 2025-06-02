// ABOUTME: Tests for the FeedConvert tool that converts between feed formats
// ABOUTME: Tests RSS, Atom, and JSON Feed conversions with various options

package feed

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestFeedConvertRegistration(t *testing.T) {
	tool := FeedConvert()

	if tool.Name() != "feed_convert" {
		t.Errorf("Expected tool name 'feed_convert', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Tool description should not be empty")
	}
}

func TestFeedConvertToRSS(t *testing.T) {
	now := time.Now()
	feed := UnifiedFeed{
		Title:       "Test Feed",
		Description: "A test feed",
		Link:        "https://example.com",
		Language:    "en",
		Updated:     &now,
		Author:      &Author{Name: "John Doe", Email: "john@example.com"},
		Items: []FeedItem{
			{
				ID:          "item1",
				Title:       "Test Item",
				Link:        "https://example.com/item1",
				Description: "Test description",
				Content:     "Full content here",
				Published:   &now,
				Author:      &Author{Name: "Jane Doe", Email: "jane@example.com"},
				Categories:  []string{"tech", "news"},
			},
		},
	}

	tool := FeedConvert()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"feed":        feed,
		"target_type": "rss",
		"pretty":      true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	convertResult, ok := result.(*FeedConvertResult)
	if !ok {
		t.Fatalf("Expected *FeedConvertResult, got %T", result)
	}

	if convertResult.ContentType != "application/rss+xml" {
		t.Errorf("Expected content type 'application/rss+xml', got '%s'", convertResult.ContentType)
	}

	if convertResult.Format != "rss" {
		t.Errorf("Expected format 'rss', got '%s'", convertResult.Format)
	}

	// Verify it's valid XML
	var rss rss2Feed
	err = xml.Unmarshal([]byte(convertResult.Content), &rss)
	if err != nil {
		t.Errorf("Failed to parse RSS output: %v", err)
	}

	// Check basic fields
	if rss.Channel.Title != "Test Feed" {
		t.Errorf("Expected title 'Test Feed', got '%s'", rss.Channel.Title)
	}

	if len(rss.Channel.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(rss.Channel.Items))
	}

	// Check categories were converted
	if len(rss.Channel.Items[0].Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(rss.Channel.Items[0].Categories))
	}
}

func TestFeedConvertToAtom(t *testing.T) {
	now := time.Now()
	feed := UnifiedFeed{
		Title:   "Test Feed",
		Link:    "https://example.com",
		Updated: &now,
		Author:  &Author{Name: "John Doe", Email: "john@example.com"},
		Items: []FeedItem{
			{
				ID:        "item1",
				Title:     "Test Item",
				Link:      "https://example.com/item1",
				Content:   "Full content",
				Published: &now,
				Updated:   &now,
			},
		},
	}

	tool := FeedConvert()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"feed":        feed,
		"target_type": "atom",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	convertResult, ok := result.(*FeedConvertResult)
	if !ok {
		t.Fatalf("Expected *FeedConvertResult, got %T", result)
	}

	if convertResult.ContentType != "application/atom+xml" {
		t.Errorf("Expected content type 'application/atom+xml', got '%s'", convertResult.ContentType)
	}

	if convertResult.Format != "atom" {
		t.Errorf("Expected format 'atom', got '%s'", convertResult.Format)
	}

	// Verify it's valid XML and contains Atom namespace
	if !strings.Contains(convertResult.Content, "http://www.w3.org/2005/Atom") {
		t.Error("Atom feed should contain Atom namespace")
	}

	// Basic structure check
	if !strings.Contains(convertResult.Content, "<feed") {
		t.Error("Atom feed should contain <feed> element")
	}

	if !strings.Contains(convertResult.Content, "<entry>") {
		t.Error("Atom feed should contain <entry> elements")
	}
}

func TestFeedConvertToJSONFeed(t *testing.T) {
	now := time.Now()
	feed := UnifiedFeed{
		Title:       "Test Feed",
		Description: "A test feed",
		Link:        "https://example.com",
		Language:    "en",
		Author:      &Author{Name: "John Doe", URL: "https://johndoe.com"},
		Items: []FeedItem{
			{
				ID:          "item1",
				Title:       "Test Item",
				Link:        "https://example.com/item1",
				Description: "Test description",
				Published:   &now,
				Categories:  []string{"tech", "news"},
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

	tool := FeedConvert()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"feed":            feed,
		"target_type":     "json",
		"include_content": false, // Use description only
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	convertResult, ok := result.(*FeedConvertResult)
	if !ok {
		t.Fatalf("Expected *FeedConvertResult, got %T", result)
	}

	if convertResult.ContentType != "application/feed+json" {
		t.Errorf("Expected content type 'application/feed+json', got '%s'", convertResult.ContentType)
	}

	if convertResult.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", convertResult.Format)
	}

	// Verify it's valid JSON
	var jf jsonFeed
	err = json.Unmarshal([]byte(convertResult.Content), &jf)
	if err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
	}

	// Check fields
	if jf.Version != "https://jsonfeed.org/version/1.1" {
		t.Errorf("Expected JSON Feed version 1.1, got '%s'", jf.Version)
	}

	if jf.Title != "Test Feed" {
		t.Errorf("Expected title 'Test Feed', got '%s'", jf.Title)
	}

	if len(jf.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(jf.Items))
	}

	// Check tags (categories)
	if len(jf.Items[0].Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(jf.Items[0].Tags))
	}

	// Check attachments (enclosures)
	if len(jf.Items[0].Attachments) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(jf.Items[0].Attachments))
	}

	// Check that content_text is used (not content_html) since include_content is false
	if jf.Items[0].ContentText != "Test description" {
		t.Errorf("Expected content_text to be 'Test description', got '%s'", jf.Items[0].ContentText)
	}
	if jf.Items[0].ContentHTML != "" {
		t.Error("Expected content_html to be empty when include_content is false")
	}
}

func TestFeedConvertPrettyPrint(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:    "1",
				Title: "Item 1",
			},
		},
	}

	tool := FeedConvert()
	ctx := context.Background()

	// Test with pretty=false
	result1, err := tool.Execute(ctx, map[string]interface{}{
		"feed":        feed,
		"target_type": "json",
		"pretty":      false,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	convertResult1 := result1.(*FeedConvertResult)
	if strings.Contains(convertResult1.Content, "\n  ") {
		t.Error("Non-pretty JSON should not contain indentation")
	}

	// Test with pretty=true
	result2, err := tool.Execute(ctx, map[string]interface{}{
		"feed":        feed,
		"target_type": "json",
		"pretty":      true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	convertResult2 := result2.(*FeedConvertResult)
	if !strings.Contains(convertResult2.Content, "\n  ") {
		t.Error("Pretty JSON should contain indentation")
	}
}

func TestFeedConvertEmptyFeed(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Empty Feed",
		Items: []FeedItem{},
	}

	tool := FeedConvert()
	ctx := context.Background()

	// Test all formats with empty feed
	formats := []string{"rss", "atom", "json"}

	for _, format := range formats {
		result, err := tool.Execute(ctx, map[string]interface{}{
			"feed":        feed,
			"target_type": format,
		})
		if err != nil {
			t.Errorf("Execute failed for format %s: %v", format, err)
			continue
		}

		convertResult, ok := result.(*FeedConvertResult)
		if !ok {
			t.Errorf("Expected *FeedConvertResult for format %s, got %T", format, result)
			continue
		}

		if convertResult.Content == "" {
			t.Errorf("Expected non-empty content for format %s", format)
		}
	}
}

func TestFeedConvertInvalidFormat(t *testing.T) {
	feed := UnifiedFeed{
		Title: "Test Feed",
	}

	tool := FeedConvert()
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"feed":        feed,
		"target_type": "invalid",
	})
	if err == nil {
		t.Error("Expected error for invalid target format")
	}
}

func TestFeedConvertWithoutAuthor(t *testing.T) {
	now := time.Now()
	feed := UnifiedFeed{
		Title: "Feed Without Author",
		Items: []FeedItem{
			{
				ID:        "1",
				Title:     "Item Without Author",
				Published: &now,
			},
		},
	}

	tool := FeedConvert()
	ctx := context.Background()

	// Test all formats - should not fail without author
	formats := []string{"rss", "atom", "json"}

	for _, format := range formats {
		result, err := tool.Execute(ctx, map[string]interface{}{
			"feed":        feed,
			"target_type": format,
		})
		if err != nil {
			t.Errorf("Execute failed for format %s: %v", format, err)
		}

		convertResult, ok := result.(*FeedConvertResult)
		if !ok || convertResult.Content == "" {
			t.Errorf("Expected valid output for format %s without author", format)
		}
	}
}

func TestFeedConvertDateHandling(t *testing.T) {
	// Test with only published date
	published := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	feed1 := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:        "1",
				Title:     "Item with published date",
				Published: &published,
			},
		},
	}

	// Test with only updated date
	_ = time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)

	tool := FeedConvert()
	ctx := context.Background()

	// Convert to Atom (which requires updated date)
	result1, err := tool.Execute(ctx, map[string]interface{}{
		"feed":        feed1,
		"target_type": "atom",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	convertResult1 := result1.(*FeedConvertResult)
	// Should use published date as updated in Atom
	if !strings.Contains(convertResult1.Content, "2024-01-15") {
		t.Error("Expected Atom feed to contain published date as updated")
	}

	// Convert feed without dates
	feed3 := UnifiedFeed{
		Title: "Test Feed",
		Items: []FeedItem{
			{
				ID:    "3",
				Title: "Item without dates",
			},
		},
	}

	result2, err := tool.Execute(ctx, map[string]interface{}{
		"feed":        feed3,
		"target_type": "atom",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should still generate valid Atom feed
	convertResult2 := result2.(*FeedConvertResult)
	if !strings.Contains(convertResult2.Content, "<feed") {
		t.Error("Expected valid Atom feed even without dates")
	}
}

