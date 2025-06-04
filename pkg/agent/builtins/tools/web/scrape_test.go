// ABOUTME: Tests for the WebScrape built-in tool
// ABOUTME: Validates HTML parsing, text extraction, link discovery, and selector support

package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Helper to create test ToolContext
func createTestToolContextForScrape() *domain.ToolContext {
	return domain.NewToolContext(
		context.Background(),
		domain.NewStateReader(domain.NewState()),
		&mockScrapeAgent{},
		"test-run",
	)
}

// mockScrapeAgent implements the minimum required methods for BaseAgent
type mockScrapeAgent struct{}

func (m *mockScrapeAgent) ID() string                                { return "test-agent" }
func (m *mockScrapeAgent) Name() string                              { return "Test Agent" }
func (m *mockScrapeAgent) Description() string                       { return "Mock agent for testing" }
func (m *mockScrapeAgent) Type() domain.AgentType                    { return domain.AgentTypeCustom }
func (m *mockScrapeAgent) Parent() domain.BaseAgent                  { return nil }
func (m *mockScrapeAgent) SetParent(parent domain.BaseAgent) error   { return nil }
func (m *mockScrapeAgent) SubAgents() []domain.BaseAgent             { return nil }
func (m *mockScrapeAgent) AddSubAgent(agent domain.BaseAgent) error  { return nil }
func (m *mockScrapeAgent) RemoveSubAgent(name string) error          { return nil }
func (m *mockScrapeAgent) FindAgent(name string) domain.BaseAgent    { return nil }
func (m *mockScrapeAgent) FindSubAgent(name string) domain.BaseAgent { return nil }
func (m *mockScrapeAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	return nil, nil
}
func (m *mockScrapeAgent) RunAsync(ctx context.Context, input *domain.State) (<-chan domain.Event, error) {
	return nil, nil
}
func (m *mockScrapeAgent) Initialize(ctx context.Context) error                     { return nil }
func (m *mockScrapeAgent) BeforeRun(ctx context.Context, state *domain.State) error { return nil }
func (m *mockScrapeAgent) AfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) error {
	return nil
}
func (m *mockScrapeAgent) Cleanup(ctx context.Context) error                     { return nil }
func (m *mockScrapeAgent) InputSchema() *sdomain.Schema                          { return nil }
func (m *mockScrapeAgent) OutputSchema() *sdomain.Schema                         { return nil }
func (m *mockScrapeAgent) Config() domain.AgentConfig                            { return domain.AgentConfig{} }
func (m *mockScrapeAgent) WithConfig(config domain.AgentConfig) domain.BaseAgent { return m }
func (m *mockScrapeAgent) Validate() error                                       { return nil }
func (m *mockScrapeAgent) Metadata() map[string]interface{}                      { return nil }
func (m *mockScrapeAgent) SetMetadata(key string, value interface{})             {}

func TestWebScrapeRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("web_scrape")
	if !ok {
		t.Fatal("WebScrape tool not registered")
	}
	if tool == nil {
		t.Fatal("WebScrape tool is nil")
	}

	// Test tool name
	if tool.Name() != "web_scrape" {
		t.Errorf("Expected tool name 'web_scrape', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("web_scrape")
	if len(entries) == 0 {
		t.Fatal("WebScrape tool not found in registry")
	}

	meta := entries[0].Metadata
	if meta.Category != "web" {
		t.Errorf("Expected category 'web', got '%s'", meta.Category)
	}
}

func TestWebScrapeExecution(t *testing.T) {
	// Create a test HTML page
	testHTML := `
<!DOCTYPE html>
<html>
<head>
	<title>Test Page Title</title>
	<meta name="description" content="Test page description">
	<meta name="keywords" content="test, scraping, html">
	<meta property="og:title" content="OG Title">
</head>
<body>
	<h1>Main Heading</h1>
	<p>This is a paragraph with some text content.</p>
	<div>
		<p>Another paragraph in a div.</p>
	</div>
	<a href="https://example.com">External Link</a>
	<a href="/internal">Internal Link</a>
	<a href="#anchor">Anchor Link</a>
	<ul>
		<li>List item 1</li>
		<li>List item 2</li>
	</ul>
</body>
</html>`

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testHTML))
	}))
	defer server.Close()

	tool := WebScrape()
	ctx := createTestToolContextForScrape()

	// Test basic scraping
	result, err := tool.Execute(ctx, map[string]interface{}{
		"url": server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to scrape URL: %v", err)
	}

	scrapeResult := result.(*WebScrapeResult)

	// Validate metadata
	if scrapeResult.Title != "Test Page Title" {
		t.Errorf("Expected title 'Test Page Title', got '%s'", scrapeResult.Title)
	}

	if scrapeResult.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", scrapeResult.StatusCode)
	}

	// Validate text extraction
	if !strings.Contains(scrapeResult.Text, "Main Heading") {
		t.Error("Expected text to contain 'Main Heading'")
	}
	if !strings.Contains(scrapeResult.Text, "This is a paragraph") {
		t.Error("Expected text to contain 'This is a paragraph'")
	}

	// Validate that script/style tags are removed
	if strings.Contains(scrapeResult.Text, "<") || strings.Contains(scrapeResult.Text, ">") {
		t.Error("Text should not contain HTML tags")
	}

	// Validate metadata extraction
	if scrapeResult.Metadata == nil {
		t.Error("Expected metadata to be extracted")
	} else {
		if scrapeResult.Metadata["description"] != "Test page description" {
			t.Errorf("Expected description metadata, got '%s'", scrapeResult.Metadata["description"])
		}
		if scrapeResult.Metadata["keywords"] != "test, scraping, html" {
			t.Errorf("Expected keywords metadata, got '%s'", scrapeResult.Metadata["keywords"])
		}
		if scrapeResult.Metadata["og:title"] != "OG Title" {
			t.Errorf("Expected og:title metadata, got '%s'", scrapeResult.Metadata["og:title"])
		}
	}

	// Validate link extraction
	if len(scrapeResult.Links) != 3 {
		t.Errorf("Expected 3 links, got %d", len(scrapeResult.Links))
	}

	// Check link types
	linkTypes := make(map[string]int)
	for _, link := range scrapeResult.Links {
		linkTypes[link.Type]++
	}
	if linkTypes["external"] != 1 {
		t.Errorf("Expected 1 external link, got %d", linkTypes["external"])
	}
	if linkTypes["internal"] != 1 {
		t.Errorf("Expected 1 internal link, got %d", linkTypes["internal"])
	}
	if linkTypes["anchor"] != 1 {
		t.Errorf("Expected 1 anchor link, got %d", linkTypes["anchor"])
	}
}

func TestWebScrapeWithSelectors(t *testing.T) {
	testHTML := `
<!DOCTYPE html>
<html>
<body>
	<h1>Title 1</h1>
	<h1>Title 2</h1>
	<p>Paragraph 1</p>
	<p>Paragraph 2</p>
	<div>Some div content</div>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(testHTML))
	}))
	defer server.Close()

	tool := WebScrape()
	ctx := createTestToolContextForScrape()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"url":       server.URL,
		"selectors": []string{"h1", "p"},
	})
	if err != nil {
		t.Fatalf("Failed to scrape with selectors: %v", err)
	}

	scrapeResult := result.(*WebScrapeResult)

	// Validate selector results
	if scrapeResult.Selectors == nil {
		t.Fatal("Expected selector results")
	}

	// Check h1 results
	if h1Results, ok := scrapeResult.Selectors["h1"]; ok {
		if len(h1Results) != 2 {
			t.Errorf("Expected 2 h1 results, got %d", len(h1Results))
		}
		if len(h1Results) >= 2 {
			if h1Results[0] != "Title 1" {
				t.Errorf("Expected first h1 to be 'Title 1', got '%s'", h1Results[0])
			}
			if h1Results[1] != "Title 2" {
				t.Errorf("Expected second h1 to be 'Title 2', got '%s'", h1Results[1])
			}
		}
	} else {
		t.Error("Expected h1 selector results")
	}

	// Check p results
	if pResults, ok := scrapeResult.Selectors["p"]; ok {
		if len(pResults) != 2 {
			t.Errorf("Expected 2 p results, got %d", len(pResults))
		}
	} else {
		t.Error("Expected p selector results")
	}
}

func TestWebScrapeMetadataOnly(t *testing.T) {
	testHTML := `
<!DOCTYPE html>
<html>
<head>
	<title>Metadata Test</title>
	<meta name="author" content="Test Author">
</head>
<body>
	<p>Body content that should not be extracted</p>
	<a href="/link">Link that should not be extracted</a>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(testHTML))
	}))
	defer server.Close()

	tool := WebScrape()
	ctx := createTestToolContextForScrape()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"url":           server.URL,
		"extract_text":  false,
		"extract_links": false,
		"extract_meta":  true,
	})
	if err != nil {
		t.Fatalf("Failed to scrape metadata only: %v", err)
	}

	scrapeResult := result.(*WebScrapeResult)

	// Text and links should be empty
	if scrapeResult.Text != "" {
		t.Error("Expected text to be empty when extract_text is false")
	}
	if len(scrapeResult.Links) > 0 {
		t.Error("Expected no links when extract_links is false")
	}

	// Metadata should be present
	if len(scrapeResult.Metadata) == 0 {
		t.Error("Expected metadata to be extracted")
	}
}

func TestWebScrapeNonHTMLContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key": "value"}`))
	}))
	defer server.Close()

	tool := WebScrape()
	ctx := createTestToolContextForScrape()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"url": server.URL,
	})
	if err == nil {
		t.Error("Expected error for non-HTML content")
	}
	if !strings.Contains(err.Error(), "not HTML/XML") {
		t.Errorf("Expected error about content type, got: %v", err)
	}
}

func TestWebScrapeInvalidURL(t *testing.T) {
	tool := WebScrape()
	ctx := createTestToolContextForScrape()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"url": "not-a-valid-url",
	})
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestWebScrapeWithCustomSelectors(t *testing.T) {
	tool := WebScrape()

	// Create tool context with custom selectors in state
	state := domain.NewState()
	state.Set("scrape_selectors", []string{"div", "span"})
	ctx := domain.NewToolContext(
		context.Background(),
		domain.NewStateReader(state),
		&mockScrapeAgent{},
		"test-run",
	)

	testHTML := `
<!DOCTYPE html>
<html>
<body>
	<div>Div content</div>
	<span>Span content</span>
	<p>Paragraph content</p>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(testHTML))
	}))
	defer server.Close()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"url":       server.URL,
		"selectors": []string{"p"}, // Should be combined with state selectors
	})
	if err != nil {
		t.Fatalf("Failed to scrape with custom selectors: %v", err)
	}

	scrapeResult := result.(*WebScrapeResult)

	// Should have results for div, span (from state) and p (from params)
	if scrapeResult.Selectors == nil {
		t.Fatal("Expected selector results")
	}

	if _, ok := scrapeResult.Selectors["div"]; !ok {
		t.Error("Expected div selector results from state")
	}
	if _, ok := scrapeResult.Selectors["span"]; !ok {
		t.Error("Expected span selector results from state")
	}
	if _, ok := scrapeResult.Selectors["p"]; !ok {
		t.Error("Expected p selector results from params")
	}
}
