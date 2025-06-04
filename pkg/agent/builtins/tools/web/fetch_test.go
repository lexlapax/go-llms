// ABOUTME: Tests for the WebFetch built-in tool
// ABOUTME: Validates URL fetching, content extraction, and error handling

package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Helper to create test ToolContext
func createTestToolContext() *domain.ToolContext {
	return domain.NewToolContext(
		context.Background(),
		domain.NewStateReader(domain.NewState()),
		&mockAgent{},
		"test-run",
	)
}

// mockAgent implements the minimum required methods for BaseAgent
type mockAgent struct{}

func (m *mockAgent) ID() string                                { return "test-agent" }
func (m *mockAgent) Name() string                              { return "Test Agent" }
func (m *mockAgent) Description() string                       { return "Mock agent for testing" }
func (m *mockAgent) Type() domain.AgentType                    { return domain.AgentTypeCustom }
func (m *mockAgent) Parent() domain.BaseAgent                  { return nil }
func (m *mockAgent) SetParent(parent domain.BaseAgent) error   { return nil }
func (m *mockAgent) SubAgents() []domain.BaseAgent             { return nil }
func (m *mockAgent) AddSubAgent(agent domain.BaseAgent) error  { return nil }
func (m *mockAgent) RemoveSubAgent(name string) error          { return nil }
func (m *mockAgent) FindAgent(name string) domain.BaseAgent    { return nil }
func (m *mockAgent) FindSubAgent(name string) domain.BaseAgent { return nil }
func (m *mockAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	return nil, nil
}
func (m *mockAgent) RunAsync(ctx context.Context, input *domain.State) (<-chan domain.Event, error) {
	return nil, nil
}
func (m *mockAgent) Initialize(ctx context.Context) error                     { return nil }
func (m *mockAgent) BeforeRun(ctx context.Context, state *domain.State) error { return nil }
func (m *mockAgent) AfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) error {
	return nil
}
func (m *mockAgent) Cleanup(ctx context.Context) error                     { return nil }
func (m *mockAgent) InputSchema() *sdomain.Schema                          { return nil }
func (m *mockAgent) OutputSchema() *sdomain.Schema                         { return nil }
func (m *mockAgent) Config() domain.AgentConfig                            { return domain.AgentConfig{} }
func (m *mockAgent) WithConfig(config domain.AgentConfig) domain.BaseAgent { return m }
func (m *mockAgent) Validate() error                                       { return nil }
func (m *mockAgent) Metadata() map[string]interface{}                      { return nil }
func (m *mockAgent) SetMetadata(key string, value interface{})             {}

func TestWebFetchRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("web_fetch")
	if !ok {
		t.Fatal("WebFetch tool not registered")
	}
	if tool == nil {
		t.Fatal("WebFetch tool is nil")
	}

	// Test tool name
	if tool.Name() != "web_fetch" {
		t.Errorf("Expected tool name 'web_fetch', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("web_fetch")
	if len(entries) == 0 {
		t.Fatal("WebFetch tool not found in registry")
	}

	meta := entries[0].Metadata
	if meta.Category != "web" {
		t.Errorf("Expected category 'web', got '%s'", meta.Category)
	}
}

func TestWebFetchBasic(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check User-Agent
		if !strings.Contains(r.UserAgent(), "go-llms") {
			t.Errorf("Expected User-Agent to contain 'go-llms', got '%s'", r.UserAgent())
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
	<meta name="description" content="A test page for WebFetch">
</head>
<body>
	<h1>Hello World</h1>
	<p>This is a test page with some content.</p>
	<a href="/link1">Link 1</a>
	<a href="https://example.com/link2">Link 2</a>
</body>
</html>`))
	}))
	defer server.Close()

	tool := WebFetch()
	ctx := createTestToolContext()

	// Test basic fetch
	result, err := tool.Execute(ctx, map[string]interface{}{
		"url": server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to fetch URL: %v", err)
	}

	fetchResult := result.(*WebFetchResult)

	// Validate result
	if fetchResult.Status != 200 {
		t.Errorf("Expected status 200, got %d", fetchResult.Status)
	}

	if fetchResult.StatusText != "200 OK" {
		t.Errorf("Expected status text '200 OK', got '%s'", fetchResult.StatusText)
	}

	if !strings.Contains(fetchResult.Content, "Hello World") {
		t.Error("Expected content to contain 'Hello World'")
	}

	if !strings.Contains(fetchResult.Content, "This is a test page") {
		t.Error("Expected content to contain 'This is a test page'")
	}

	// The content should include the full HTML
	if !strings.Contains(fetchResult.Content, "<title>Test Page</title>") {
		t.Error("Expected content to contain HTML title tag")
	}

	// Check headers
	if fetchResult.Headers["Content-Type"] != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type header, got '%s'", fetchResult.Headers["Content-Type"])
	}
}

func TestWebFetchTimeout(t *testing.T) {
	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than our timeout
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Too late"))
	}))
	defer server.Close()

	tool := WebFetch()
	ctx := createTestToolContext()

	// Test with short timeout (1 second, while server sleeps for 2)
	_, err := tool.Execute(ctx, map[string]interface{}{
		"url":     server.URL,
		"timeout": 1, // 1 second timeout
	})

	if err == nil {
		t.Error("Expected timeout error")
	}

	if err != nil && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestWebFetchErrorHandling(t *testing.T) {
	tool := WebFetch()
	ctx := createTestToolContext()

	// Test 1: Invalid URL
	_, err := tool.Execute(ctx, map[string]interface{}{
		"url": "not-a-valid-url",
	})
	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	// Test 2: Non-existent domain
	_, err = tool.Execute(ctx, map[string]interface{}{
		"url": "http://this-domain-definitely-does-not-exist-12345.com",
	})
	if err == nil {
		t.Error("Expected error for non-existent domain")
	}
}

func TestWebFetchStatusCodes(t *testing.T) {
	testCases := []struct {
		statusCode int
		expectErr  bool
	}{
		{200, false},
		{201, false},
		{301, false}, // Redirects are followed by default
		{404, false}, // We still return the result, just with status
		{500, false}, // We still return the result, just with status
	}

	for _, tc := range testCases {
		t.Run(string(rune(tc.statusCode))+" status", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte("Response body"))
			}))
			defer server.Close()

			tool := WebFetch()
			ctx := createTestToolContext()

			result, err := tool.Execute(ctx, map[string]interface{}{
				"url": server.URL,
			})

			if tc.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil {
				fetchResult := result.(*WebFetchResult)
				if fetchResult.Status != tc.statusCode {
					t.Errorf("Expected status %d, got %d", tc.statusCode, fetchResult.Status)
				}
			}
		})
	}
}

func TestWebFetchContentTypes(t *testing.T) {
	testCases := []struct {
		contentType string
		body        string
		shouldParse bool
	}{
		{"text/html", "<html><body>HTML content</body></html>", true},
		{"text/plain", "Plain text content", true},
		{"application/json", `{"key": "value"}`, true},
		{"application/xml", `<?xml version="1.0"?><root>XML content</root>`, true},
		{"image/png", "binary-image-data", false},
		{"application/pdf", "binary-pdf-data", false},
	}

	for _, tc := range testCases {
		t.Run(tc.contentType, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tc.contentType)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			tool := WebFetch()
			ctx := createTestToolContext()

			result, err := tool.Execute(ctx, map[string]interface{}{
				"url": server.URL,
			})
			if err != nil {
				t.Fatalf("Failed to fetch: %v", err)
			}

			fetchResult := result.(*WebFetchResult)

			if tc.shouldParse {
				if fetchResult.Content == "" {
					t.Error("Expected content to be extracted")
				}
			}

			if fetchResult.Headers["Content-Type"] != tc.contentType {
				t.Errorf("Expected Content-Type '%s', got '%s'", tc.contentType, fetchResult.Headers["Content-Type"])
			}
		})
	}
}

func TestWebFetchWithCustomUserAgent(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check custom User-Agent
		if r.UserAgent() != "CustomBot/1.0" {
			t.Errorf("Expected User-Agent 'CustomBot/1.0', got '%s'", r.UserAgent())
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	tool := WebFetch()

	// Create tool context with custom user agent in state
	state := domain.NewState()
	state.Set("user_agent", "CustomBot/1.0")
	ctx := domain.NewToolContext(
		context.Background(),
		domain.NewStateReader(state),
		&mockAgent{},
		"test-run",
	)

	// Test fetch with custom user agent
	_, err := tool.Execute(ctx, map[string]interface{}{
		"url": server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to fetch URL: %v", err)
	}
}

func TestWebFetchWithAdditionalHeaders(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check custom headers
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Errorf("Expected X-Custom-Header 'custom-value', got '%s'", r.Header.Get("X-Custom-Header"))
		}
		if r.Header.Get("Authorization") != "Bearer token123" {
			t.Errorf("Expected Authorization 'Bearer token123', got '%s'", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	tool := WebFetch()

	// Create tool context with additional headers in state
	state := domain.NewState()
	state.Set("http_headers", map[string]string{
		"X-Custom-Header": "custom-value",
		"Authorization":   "Bearer token123",
	})
	ctx := domain.NewToolContext(
		context.Background(),
		domain.NewStateReader(state),
		&mockAgent{},
		"test-run",
	)

	// Test fetch with additional headers
	_, err := tool.Execute(ctx, map[string]interface{}{
		"url": server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to fetch URL: %v", err)
	}
}
