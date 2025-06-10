// ABOUTME: Tests for the FeedDiscover tool that automatically discovers feed URLs from web pages
// ABOUTME: Tests HTML parsing, common path checking, and feed verification functionality

package feed

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFeedDiscoverRegistration(t *testing.T) {
	tool := FeedDiscover()

	if tool.Name() != "feed_discover" {
		t.Errorf("Expected tool name 'feed_discover', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Tool description should not be empty")
	}
}

func TestFeedDiscoverFromHTML(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
	<title>Test Blog</title>
	<link rel="alternate" type="application/rss+xml" title="RSS Feed" href="/blog/feed.xml">
	<link rel="alternate" type="application/atom+xml" title="Atom Feed" href="/blog/atom.xml">
	<link rel="alternate" type="application/feed+json" title="JSON Feed" href="/blog/feed.json">
	<link rel="stylesheet" href="/style.css">
</head>
<body>
	<h1>Welcome to Test Blog</h1>
</body>
</html>`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(htmlContent))
		case "/blog/feed.xml":
			w.Header().Set("Content-Type", "application/rss+xml")
			w.WriteHeader(http.StatusOK)
		case "/blog/atom.xml":
			w.Header().Set("Content-Type", "application/atom+xml")
			w.WriteHeader(http.StatusOK)
		case "/blog/feed.json":
			w.Header().Set("Content-Type", "application/feed+json")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tool := FeedDiscover()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"url": server.URL,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	feedResult, ok := result.(*FeedDiscoverResult)
	if !ok {
		t.Fatalf("Expected *FeedDiscoverResult, got %T", result)
	}

	if feedResult.Error != "" {
		t.Fatalf("Unexpected error: %s", feedResult.Error)
	}

	// Should find at least 3 feeds from HTML
	if len(feedResult.Feeds) < 3 {
		t.Errorf("Expected at least 3 feeds, got %d", len(feedResult.Feeds))
	}

	// Check that we found the expected feeds
	foundRSS := false
	foundAtom := false
	foundJSON := false

	for _, feed := range feedResult.Feeds {
		if feed.Source == "link_tag" {
			switch feed.Type {
			case "rss":
				foundRSS = true
				if feed.Title != "RSS Feed" {
					t.Errorf("Expected RSS feed title 'RSS Feed', got '%s'", feed.Title)
				}
			case "atom":
				foundAtom = true
				if feed.Title != "Atom Feed" {
					t.Errorf("Expected Atom feed title 'Atom Feed', got '%s'", feed.Title)
				}
			case "json":
				foundJSON = true
				if feed.Title != "JSON Feed" {
					t.Errorf("Expected JSON feed title 'JSON Feed', got '%s'", feed.Title)
				}
			}
		}
	}

	if !foundRSS {
		t.Error("Did not find RSS feed in link tags")
	}
	if !foundAtom {
		t.Error("Did not find Atom feed in link tags")
	}
	if !foundJSON {
		t.Error("Did not find JSON feed in link tags")
	}
}

func TestFeedDiscoverCommonPaths(t *testing.T) {
	// Create test server that responds to common feed paths
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html><body>Test Site</body></html>"))
		case "/feed", "/rss":
			if r.Method == "HEAD" {
				w.Header().Set("Content-Type", "application/rss+xml")
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		case "/atom.xml":
			if r.Method == "HEAD" {
				w.Header().Set("Content-Type", "application/atom+xml")
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tool := FeedDiscover()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"url": server.URL,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	feedResult, ok := result.(*FeedDiscoverResult)
	if !ok {
		t.Fatalf("Expected *FeedDiscoverResult, got %T", result)
	}

	// Should find feeds from common paths
	foundCommonPath := false
	for _, feed := range feedResult.Feeds {
		if feed.Source == "common_path" {
			foundCommonPath = true
			break
		}
	}

	if !foundCommonPath {
		t.Error("Did not find any feeds from common paths")
	}
}

func TestFeedDiscoverRelativeURLs(t *testing.T) {
	// HTML with relative URLs
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
	<link rel="alternate" type="application/rss+xml" href="feed.xml">
	<link rel="alternate" type="application/atom+xml" href="../atom.xml">
	<link rel="alternate" type="application/feed+json" href="/absolute/feed.json">
</head>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/blog/page.html" {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(htmlContent))
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	tool := FeedDiscover()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"url": server.URL + "/blog/page.html",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	feedResult, ok := result.(*FeedDiscoverResult)
	if !ok {
		t.Fatalf("Expected *FeedDiscoverResult, got %T", result)
	}

	// Check that relative URLs were resolved correctly
	expectedURLs := map[string]bool{
		server.URL + "/blog/feed.xml":      false,
		server.URL + "/atom.xml":           false,
		server.URL + "/absolute/feed.json": false,
	}

	for _, feed := range feedResult.Feeds {
		if feed.Source == "link_tag" {
			if _, ok := expectedURLs[feed.URL]; ok {
				expectedURLs[feed.URL] = true
			}
		}
	}

	for url, found := range expectedURLs {
		if !found {
			t.Errorf("Expected to find URL %s, but it was not discovered", url)
		}
	}
}

func TestFeedDiscoverErrors(t *testing.T) {
	tool := FeedDiscover()
	tc := createTestToolContext()

	// Test invalid URL
	result, err := tool.Execute(tc, map[string]interface{}{
		"url": "not-a-valid-url",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	feedResult, ok := result.(*FeedDiscoverResult)
	if !ok {
		t.Fatalf("Expected *FeedDiscoverResult, got %T", result)
	}

	if feedResult.Error == "" {
		t.Error("Expected error for invalid URL")
	}

	// Test non-existent domain
	result, err = tool.Execute(tc, map[string]interface{}{
		"url": "http://non-existent-domain-12345.com",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	feedResult, ok = result.(*FeedDiscoverResult)
	if !ok {
		t.Fatalf("Expected *FeedDiscoverResult, got %T", result)
	}

	if feedResult.Error == "" {
		t.Error("Expected error for non-existent domain")
	}
}

func TestFeedDiscoverTimeout(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't respond, let it timeout
		<-r.Context().Done()
	}))
	defer server.Close()

	tool := FeedDiscover()
	tc := createTestToolContext()

	result, err := tool.Execute(tc, map[string]interface{}{
		"url":     server.URL,
		"timeout": 1, // 1 second timeout
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	feedResult, ok := result.(*FeedDiscoverResult)
	if !ok {
		t.Fatalf("Expected *FeedDiscoverResult, got %T", result)
	}

	if feedResult.Error == "" {
		t.Error("Expected timeout error")
	}
}

func TestFeedDiscoverNoFollowRedirects(t *testing.T) {
	// Simple test: when follow_redirects is false and we get a redirect,
	// we should get an error in the result
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/page-with-redirect":
			w.Header().Set("Location", "/target")
			w.WriteHeader(http.StatusFound)
		case "/page-with-feeds":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html>
				<head>
					<link rel="alternate" type="application/rss+xml" href="/feed.xml">
				</head>
				<body>Page with feeds</body>
			</html>`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tool := FeedDiscover()
	tc := createTestToolContext()

	// Test 1: Page with feeds should work
	result, err := tool.Execute(tc, map[string]interface{}{
		"url": server.URL + "/page-with-feeds",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	feedResult, ok := result.(*FeedDiscoverResult)
	if !ok {
		t.Fatalf("Expected *FeedDiscoverResult, got %T", result)
	}

	if feedResult.Error != "" {
		t.Errorf("Unexpected error: %s", feedResult.Error)
	}

	if len(feedResult.Feeds) == 0 {
		t.Error("Expected to find feeds")
	}

	// Test 2: Redirect page with follow_redirects=false should give an error
	result2, err := tool.Execute(tc, map[string]interface{}{
		"url":              server.URL + "/page-with-redirect",
		"follow_redirects": false,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	feedResult2, ok := result2.(*FeedDiscoverResult)
	if !ok {
		t.Fatalf("Expected *FeedDiscoverResult, got %T", result2)
	}

	// When we don't follow redirects and get a 302, we should get an error
	if feedResult2.Error == "" {
		t.Error("Expected error for redirect when follow_redirects=false")
	}
}
