// ABOUTME: Tests for the FeedFetch built-in tool
// ABOUTME: Validates RSS, Atom, and JSON Feed parsing with various edge cases

package feed

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
)

func TestFeedFetchRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("feed_fetch")
	if !ok {
		t.Fatal("FeedFetch tool not registered")
	}
	if tool == nil {
		t.Fatal("FeedFetch tool is nil")
	}

	// Test tool name
	if tool.Name() != "feed_fetch" {
		t.Errorf("Expected tool name 'feed_fetch', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("feed_fetch")
	if len(entries) == 0 {
		t.Fatal("FeedFetch tool not found in registry")
	}

	meta := entries[0].Metadata
	if meta.Category != "feed" {
		t.Errorf("Expected category 'feed', got '%s'", meta.Category)
	}
}

func TestFeedFetchRSS(t *testing.T) {
	// Create test RSS feed
	rssFeed := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test RSS Feed</title>
		<description>A test RSS feed</description>
		<link>https://example.com</link>
		<language>en-us</language>
		<copyright>Copyright 2024</copyright>
		<pubDate>Mon, 30 Jan 2024 12:00:00 GMT</pubDate>
		<lastBuildDate>Mon, 30 Jan 2024 13:00:00 GMT</lastBuildDate>
		<image>
			<url>https://example.com/logo.png</url>
			<title>Example Logo</title>
			<link>https://example.com</link>
		</image>
		<item>
			<title>First Post</title>
			<description>This is the first post</description>
			<link>https://example.com/post1</link>
			<guid>https://example.com/post1</guid>
			<pubDate>Mon, 30 Jan 2024 10:00:00 GMT</pubDate>
			<author>john@example.com</author>
			<category>Technology</category>
			<category>News</category>
			<enclosure url="https://example.com/audio.mp3" type="audio/mpeg" length="1234567"/>
		</item>
		<item>
			<title>Second Post</title>
			<description>This is the second post</description>
			<link>https://example.com/post2</link>
			<guid>post-2</guid>
			<pubDate>Mon, 30 Jan 2024 11:00:00 GMT</pubDate>
		</item>
	</channel>
</rss>`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Header().Set("Last-Modified", "Mon, 30 Jan 2024 13:00:00 GMT")
		w.Header().Set("ETag", `W/"123456"`)
		_, _ = w.Write([]byte(rssFeed))
	}))
	defer server.Close()

	tool := FeedFetch()
	tc := createTestToolContext()

	// Test basic RSS fetch
	result, err := tool.Execute(tc, map[string]interface{}{
		"url": server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to fetch RSS feed: %v", err)
	}

	fetchResult := result.(*FeedFetchResult)

	// Check format detection
	if fetchResult.Format != "RSS2" {
		t.Errorf("Expected format 'RSS2', got '%s'", fetchResult.Format)
	}

	// Check feed metadata
	feed := fetchResult.Feed
	if feed.Title != "Test RSS Feed" {
		t.Errorf("Expected title 'Test RSS Feed', got '%s'", feed.Title)
	}
	if feed.Description != "A test RSS feed" {
		t.Errorf("Expected description 'A test RSS feed', got '%s'", feed.Description)
	}
	if feed.Link != "https://example.com" {
		t.Errorf("Expected link 'https://example.com', got '%s'", feed.Link)
	}
	if feed.Language != "en-us" {
		t.Errorf("Expected language 'en-us', got '%s'", feed.Language)
	}
	if feed.Copyright != "Copyright 2024" {
		t.Errorf("Expected copyright 'Copyright 2024', got '%s'", feed.Copyright)
	}

	// Check dates
	if feed.Published == nil {
		t.Error("Expected published date")
	}
	if feed.Updated == nil {
		t.Error("Expected updated date")
	}

	// Check image
	if feed.Image == nil {
		t.Fatal("Expected feed image")
	}
	if feed.Image.URL != "https://example.com/logo.png" {
		t.Errorf("Expected image URL 'https://example.com/logo.png', got '%s'", feed.Image.URL)
	}

	// Check items
	if len(feed.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(feed.Items))
	}

	// Check first item
	item1 := feed.Items[0]
	if item1.Title != "First Post" {
		t.Errorf("Expected item title 'First Post', got '%s'", item1.Title)
	}
	if item1.ID != "https://example.com/post1" {
		t.Errorf("Expected item ID 'https://example.com/post1', got '%s'", item1.ID)
	}
	if len(item1.Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(item1.Categories))
	}
	if item1.Author == nil || item1.Author.Name != "john@example.com" {
		t.Error("Expected author 'john@example.com'")
	}
	if len(item1.Enclosures) != 1 {
		t.Errorf("Expected 1 enclosure, got %d", len(item1.Enclosures))
	} else {
		enc := item1.Enclosures[0]
		if enc.URL != "https://example.com/audio.mp3" {
			t.Errorf("Expected enclosure URL 'https://example.com/audio.mp3', got '%s'", enc.URL)
		}
		if enc.Type != "audio/mpeg" {
			t.Errorf("Expected enclosure type 'audio/mpeg', got '%s'", enc.Type)
		}
		if enc.Length != 1234567 {
			t.Errorf("Expected enclosure length 1234567, got %d", enc.Length)
		}
	}

	// Check headers
	if fetchResult.Headers["Last-Modified"] != "Mon, 30 Jan 2024 13:00:00 GMT" {
		t.Error("Expected Last-Modified header")
	}
	if fetchResult.Headers["ETag"] != `W/"123456"` {
		t.Error("Expected ETag header")
	}
}

func TestFeedFetchAtom(t *testing.T) {
	// Create test Atom feed
	atomFeed := `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
	<title>Test Atom Feed</title>
	<id>https://example.com/feed</id>
	<updated>2024-01-30T13:00:00Z</updated>
	<link rel="alternate" href="https://example.com"/>
	<link rel="self" href="https://example.com/feed"/>
	<author>
		<name>John Doe</name>
		<email>john@example.com</email>
		<uri>https://example.com/john</uri>
	</author>
	<entry>
		<id>https://example.com/entry1</id>
		<title>First Entry</title>
		<updated>2024-01-30T10:00:00Z</updated>
		<published>2024-01-30T09:00:00Z</published>
		<link rel="alternate" href="https://example.com/entry1"/>
		<content type="html">This is the &lt;b&gt;first&lt;/b&gt; entry</content>
		<summary>First entry summary</summary>
		<author>
			<name>Jane Smith</name>
		</author>
		<category term="Technology"/>
		<category term="Atom"/>
	</entry>
	<entry>
		<id>entry-2</id>
		<title>Second Entry</title>
		<updated>2024-01-30T11:00:00Z</updated>
		<link href="https://example.com/entry2"/>
		<summary>Second entry summary</summary>
	</entry>
</feed>`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		_, _ = w.Write([]byte(atomFeed))
	}))
	defer server.Close()

	tool := FeedFetch()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"url": server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to fetch Atom feed: %v", err)
	}

	fetchResult := result.(*FeedFetchResult)

	// Check format detection
	if fetchResult.Format != "Atom" {
		t.Errorf("Expected format 'Atom', got '%s'", fetchResult.Format)
	}

	// Check feed metadata
	feed := fetchResult.Feed
	if feed.Title != "Test Atom Feed" {
		t.Errorf("Expected title 'Test Atom Feed', got '%s'", feed.Title)
	}
	if feed.Link != "https://example.com" {
		t.Errorf("Expected link 'https://example.com', got '%s'", feed.Link)
	}

	// Check author
	if feed.Author == nil {
		t.Fatal("Expected feed author")
	}
	if feed.Author.Name != "John Doe" {
		t.Errorf("Expected author name 'John Doe', got '%s'", feed.Author.Name)
	}
	if feed.Author.Email != "john@example.com" {
		t.Errorf("Expected author email 'john@example.com', got '%s'", feed.Author.Email)
	}

	// Check items
	if len(feed.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(feed.Items))
	}

	// Check first item
	item1 := feed.Items[0]
	if item1.Title != "First Entry" {
		t.Errorf("Expected item title 'First Entry', got '%s'", item1.Title)
	}
	if item1.Content != "This is the <b>first</b> entry" {
		t.Errorf("Expected HTML content, got '%s'", item1.Content)
	}
	if item1.Description != "First entry summary" {
		t.Errorf("Expected summary, got '%s'", item1.Description)
	}
	if len(item1.Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(item1.Categories))
	}
	if item1.Author == nil || item1.Author.Name != "Jane Smith" {
		t.Error("Expected author 'Jane Smith'")
	}
}

func TestFeedFetchJSONFeed(t *testing.T) {
	// Create test JSON Feed
	jsonFeed := `{
		"version": "https://jsonfeed.org/version/1.1",
		"title": "Test JSON Feed",
		"home_page_url": "https://example.com",
		"feed_url": "https://example.com/feed.json",
		"description": "A test JSON feed",
		"icon": "https://example.com/icon.png",
		"author": {
			"name": "John Doe",
			"url": "https://example.com/john"
		},
		"items": [
			{
				"id": "1",
				"url": "https://example.com/post1",
				"title": "First Post",
				"content_html": "<p>This is the first post</p>",
				"summary": "First post summary",
				"date_published": "2024-01-30T10:00:00Z",
				"date_modified": "2024-01-30T10:30:00Z",
				"author": {
					"name": "Jane Smith"
				},
				"tags": ["json", "feed"],
				"attachments": [
					{
						"url": "https://example.com/file.pdf",
						"mime_type": "application/pdf",
						"size_in_bytes": 123456
					}
				]
			},
			{
				"id": "2",
				"url": "https://example.com/post2",
				"title": "Second Post",
				"content_text": "This is plain text",
				"date_published": "2024-01-30T11:00:00Z"
			}
		]
	}`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(jsonFeed))
	}))
	defer server.Close()

	tool := FeedFetch()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"url": server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to fetch JSON Feed: %v", err)
	}

	fetchResult := result.(*FeedFetchResult)

	// Check format detection
	if fetchResult.Format != "JSONFeed" {
		t.Errorf("Expected format 'JSONFeed', got '%s'", fetchResult.Format)
	}

	// Check feed metadata
	feed := fetchResult.Feed
	if feed.Title != "Test JSON Feed" {
		t.Errorf("Expected title 'Test JSON Feed', got '%s'", feed.Title)
	}
	if feed.Description != "A test JSON feed" {
		t.Errorf("Expected description 'A test JSON feed', got '%s'", feed.Description)
	}
	if feed.Image == nil || feed.Image.URL != "https://example.com/icon.png" {
		t.Error("Expected feed icon")
	}

	// Check items
	if len(feed.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(feed.Items))
	}

	// Check first item
	item1 := feed.Items[0]
	if item1.Title != "First Post" {
		t.Errorf("Expected item title 'First Post', got '%s'", item1.Title)
	}
	if item1.Content != "<p>This is the first post</p>" {
		t.Errorf("Expected HTML content, got '%s'", item1.Content)
	}
	if len(item1.Categories) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(item1.Categories))
	}
	if len(item1.Enclosures) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(item1.Enclosures))
	}

	// Check second item (text content)
	item2 := feed.Items[1]
	if item2.Content != "This is plain text" {
		t.Errorf("Expected text content, got '%s'", item2.Content)
	}
}

func TestFeedFetchMaxItems(t *testing.T) {
	// Create RSS feed with many items
	var items strings.Builder
	for i := 1; i <= 20; i++ {
		items.WriteString(fmt.Sprintf(`
		<item>
			<title>Post %d</title>
			<link>https://example.com/post%d</link>
			<guid>post-%d</guid>
		</item>`, i, i, i))
	}

	rssFeed := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Feed</title>
		<link>https://example.com</link>
		%s
	</channel>
</rss>`, items.String())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(rssFeed))
	}))
	defer server.Close()

	tool := FeedFetch()
	tc := createTestToolContext()

	// Test with max_items limit
	result, err := tool.Execute(tc, map[string]interface{}{
		"url":       server.URL,
		"max_items": 5,
	})
	if err != nil {
		t.Fatalf("Failed to fetch feed: %v", err)
	}

	fetchResult := result.(*FeedFetchResult)
	if len(fetchResult.Feed.Items) != 5 {
		t.Errorf("Expected 5 items with max_items limit, got %d", len(fetchResult.Feed.Items))
	}

	// Verify we got the first 5 items
	for i := 0; i < 5; i++ {
		expectedTitle := fmt.Sprintf("Post %d", i+1)
		if fetchResult.Feed.Items[i].Title != expectedTitle {
			t.Errorf("Expected item %d title '%s', got '%s'", i, expectedTitle, fetchResult.Feed.Items[i].Title)
		}
	}
}

func TestFeedFetchConditionalRequest(t *testing.T) {
	// Create test server that handles conditional requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check If-None-Match header
		if r.Header.Get("If-None-Match") == `W/"123456"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// Check If-Modified-Since header
		if r.Header.Get("If-Modified-Since") == "Mon, 30 Jan 2024 12:00:00 GMT" {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// Return feed
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Feed</title>
		<link>https://example.com</link>
	</channel>
</rss>`))
	}))
	defer server.Close()

	tool := FeedFetch()
	tc := createTestToolContext()

	// Test with ETag
	result, err := tool.Execute(tc, map[string]interface{}{
		"url":  server.URL,
		"etag": `W/"123456"`,
	})
	if err != nil {
		t.Fatalf("Failed to fetch feed: %v", err)
	}

	fetchResult := result.(*FeedFetchResult)
	if !fetchResult.NotModified {
		t.Error("Expected NotModified flag to be true")
	}
	if fetchResult.Status != http.StatusNotModified {
		t.Errorf("Expected status 304, got %d", fetchResult.Status)
	}

	// Test with If-Modified-Since
	result, err = tool.Execute(tc, map[string]interface{}{
		"url":         server.URL,
		"if_modified": "Mon, 30 Jan 2024 12:00:00 GMT",
	})
	if err != nil {
		t.Fatalf("Failed to fetch feed: %v", err)
	}

	fetchResult = result.(*FeedFetchResult)
	if !fetchResult.NotModified {
		t.Error("Expected NotModified flag to be true")
	}
}

func TestFeedFetchErrors(t *testing.T) {
	tool := FeedFetch()
	tc := createTestToolContext()

	// Test invalid URL
	_, err := tool.Execute(tc, map[string]interface{}{
		"url": "not-a-valid-url",
	})
	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	// Test 404 error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err = tool.Execute(tc, map[string]interface{}{
		"url": server.URL,
	})
	if err == nil {
		t.Error("Expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("Expected error to contain '404', got: %v", err)
	}

	// Test invalid feed format
	invalidServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("This is not a feed"))
	}))
	defer invalidServer.Close()

	_, err = tool.Execute(tc, map[string]interface{}{
		"url": invalidServer.URL,
	})
	if err == nil {
		t.Error("Expected error for invalid feed")
	}
	if !strings.Contains(err.Error(), "parsing feed") {
		t.Errorf("Expected error to contain 'parsing feed', got: %v", err)
	}
}

func TestFeedFetchTimeout(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		_, _ = w.Write([]byte("slow response"))
	}))
	defer server.Close()

	tool := FeedFetch()
	tc := createTestToolContext()

	// Test with short timeout
	_, err := tool.Execute(tc, map[string]interface{}{
		"url":     server.URL,
		"timeout": 1, // 1 second timeout
	})
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestFeedFetchCustomUserAgent(t *testing.T) {
	var receivedUserAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgent = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test</title>
		<link>https://example.com</link>
	</channel>
</rss>`))
	}))
	defer server.Close()

	tool := FeedFetch()
	tc := createTestToolContext()

	// Test with custom user agent
	_, err := tool.Execute(tc, map[string]interface{}{
		"url":        server.URL,
		"user_agent": "MyCustomAgent/1.0",
	})
	if err != nil {
		t.Fatalf("Failed to fetch feed: %v", err)
	}

	if receivedUserAgent != "MyCustomAgent/1.0" {
		t.Errorf("Expected user agent 'MyCustomAgent/1.0', got '%s'", receivedUserAgent)
	}

	// Test with default user agent
	_, err = tool.Execute(tc, map[string]interface{}{
		"url": server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to fetch feed: %v", err)
	}

	if receivedUserAgent != "go-llms-feed/1.0" {
		t.Errorf("Expected default user agent 'go-llms-feed/1.0', got '%s'", receivedUserAgent)
	}
}
