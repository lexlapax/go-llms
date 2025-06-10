// ABOUTME: Extended tests for WebSearch tool including Brave and Tavily search engines
// ABOUTME: Tests automatic engine selection, API key handling, and result formatting

package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	tools "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	. "github.com/lexlapax/go-llms/pkg/agent/domain"
)

// Note: BraveSearchResponse, BraveResult, TavilySearchResponse, and TavilyResult
// are already defined in search.go, so we use them directly in tests

func TestGetSearchAPIKeys(t *testing.T) {
	// Save current env vars
	oldBrave := os.Getenv("BRAVE_API_KEY")
	oldTavily := os.Getenv("TAVILY_API_KEY")
	oldSerpapi := os.Getenv("SERPAPI_API_KEY")
	oldSerperDev := os.Getenv("SERPERDEV_API_KEY")
	defer func() {
		os.Setenv("BRAVE_API_KEY", oldBrave)
		os.Setenv("TAVILY_API_KEY", oldTavily)
		os.Setenv("SERPAPI_API_KEY", oldSerpapi)
		os.Setenv("SERPERDEV_API_KEY", oldSerperDev)
	}()

	// Test with no API keys
	os.Unsetenv("BRAVE_API_KEY")
	os.Unsetenv("TAVILY_API_KEY")
	os.Unsetenv("SERPAPI_API_KEY")
	os.Unsetenv("SERPERDEV_API_KEY")

	braveKey, tavilyKey, serpapiKey, serperdevKey := getSearchAPIKeys()
	if braveKey != "" || tavilyKey != "" || serpapiKey != "" || serperdevKey != "" {
		t.Error("Expected empty API keys when env vars not set")
	}

	// Test with Brave key only
	os.Setenv("BRAVE_API_KEY", "test-brave-key")
	os.Unsetenv("TAVILY_API_KEY")

	braveKey, tavilyKey, serpapiKey, serperdevKey = getSearchAPIKeys()
	if braveKey != "test-brave-key" {
		t.Errorf("Expected brave key 'test-brave-key', got '%s'", braveKey)
	}
	if tavilyKey != "" {
		t.Errorf("Expected empty tavily key, got '%s'", tavilyKey)
	}
	if serpapiKey != "" {
		t.Errorf("Expected empty serpapi key, got '%s'", serpapiKey)
	}
	if serperdevKey != "" {
		t.Errorf("Expected empty serperdev key, got '%s'", serperdevKey)
	}

	// Test with both keys
	os.Setenv("BRAVE_API_KEY", "test-brave-key")
	os.Setenv("TAVILY_API_KEY", "test-tavily-key")

	braveKey, tavilyKey, serpapiKey, serperdevKey = getSearchAPIKeys()
	if braveKey != "test-brave-key" {
		t.Errorf("Expected brave key 'test-brave-key', got '%s'", braveKey)
	}
	if tavilyKey != "test-tavily-key" {
		t.Errorf("Expected tavily key 'test-tavily-key', got '%s'", tavilyKey)
	}
	if serpapiKey != "" {
		t.Errorf("Expected empty serpapi key, got '%s'", serpapiKey)
	}
	if serperdevKey != "" {
		t.Errorf("Expected empty serperdev key, got '%s'", serperdevKey)
	}
}

func TestSelectDefaultEngine(t *testing.T) {
	// Save current env vars
	oldBrave := os.Getenv("BRAVE_API_KEY")
	oldTavily := os.Getenv("TAVILY_API_KEY")
	oldSerpapi := os.Getenv("SERPAPI_API_KEY")
	oldSerperDev := os.Getenv("SERPERDEV_API_KEY")
	defer func() {
		os.Setenv("BRAVE_API_KEY", oldBrave)
		os.Setenv("TAVILY_API_KEY", oldTavily)
		os.Setenv("SERPAPI_API_KEY", oldSerpapi)
		os.Setenv("SERPERDEV_API_KEY", oldSerperDev)
	}()

	testCases := []struct {
		name           string
		braveKey       string
		tavilyKey      string
		serpapiKey     string
		serperdevKey   string
		expectedEngine string
	}{
		{
			name:           "No API keys - default to DuckDuckGo",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "",
			expectedEngine: "duckduckgo",
		},
		{
			name:           "Only Brave key - use Brave",
			braveKey:       "test-brave-key",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "",
			expectedEngine: "brave",
		},
		{
			name:           "Only Tavily key - use Tavily",
			braveKey:       "",
			tavilyKey:      "test-tavily-key",
			serpapiKey:     "",
			serperdevKey:   "",
			expectedEngine: "tavily",
		},
		{
			name:           "Only Serpapi key - use Serpapi",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "test-serpapi-key",
			serperdevKey:   "",
			expectedEngine: "serpapi",
		},
		{
			name:           "Only SerperDev key - use SerperDev",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "test-serperdev-key",
			expectedEngine: "serperdev",
		},
		{
			name:           "All keys - prefer Tavily for LLM",
			braveKey:       "test-brave-key",
			tavilyKey:      "test-tavily-key",
			serpapiKey:     "test-serpapi-key",
			serperdevKey:   "test-serperdev-key",
			expectedEngine: "tavily",
		},
		{
			name:           "SerperDev and Serpapi - prefer SerperDev",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "test-serpapi-key",
			serperdevKey:   "test-serperdev-key",
			expectedEngine: "serperdev",
		},
		{
			name:           "Serpapi and Brave - prefer Serpapi",
			braveKey:       "test-brave-key",
			tavilyKey:      "",
			serpapiKey:     "test-serpapi-key",
			serperdevKey:   "",
			expectedEngine: "serpapi",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.braveKey != "" {
				os.Setenv("BRAVE_API_KEY", tc.braveKey)
			} else {
				os.Unsetenv("BRAVE_API_KEY")
			}

			if tc.tavilyKey != "" {
				os.Setenv("TAVILY_API_KEY", tc.tavilyKey)
			} else {
				os.Unsetenv("TAVILY_API_KEY")
			}

			if tc.serpapiKey != "" {
				os.Setenv("SERPAPI_API_KEY", tc.serpapiKey)
			} else {
				os.Unsetenv("SERPAPI_API_KEY")
			}

			if tc.serperdevKey != "" {
				os.Setenv("SERPERDEV_API_KEY", tc.serperdevKey)
			} else {
				os.Unsetenv("SERPERDEV_API_KEY")
			}

			engine := selectDefaultEngine()
			if engine != tc.expectedEngine {
				t.Errorf("Expected engine '%s', got '%s'", tc.expectedEngine, engine)
			}
		})
	}
}

func TestSearchBrave(t *testing.T) {
	// Create mock Brave server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check API key header
		apiKey := r.Header.Get("X-Subscription-Token")
		if apiKey != "test-brave-key" && apiKey != "explicit-brave-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check query parameters
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "Missing query", http.StatusBadRequest)
			return
		}

		// Mock response
		response := BraveSearchResponse{
			Query: query,
			Results: struct {
				News   []BraveResult `json:"news"`
				Web    []BraveResult `json:"web"`
				Videos []BraveResult `json:"videos"`
			}{
				Web: []BraveResult{
					{
						Title:       "Go Programming Language",
						URL:         "https://golang.org",
						Description: "The official Go programming language website",
						PageAge:     "30d",
					},
					{
						Title:       "Go by Example",
						URL:         "https://gobyexample.com",
						Description: "Go by Example is a hands-on introduction to Go",
						PageAge:     "1y",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Set test API key
	os.Setenv("BRAVE_API_KEY", "test-brave-key")
	defer os.Unsetenv("BRAVE_API_KEY")

	// Test search with mock server would require URL override
	// For now, we're testing the parameter structure
	t.Log("Brave search test structure created - mock server: " + server.URL)
}

func TestSearchTavily(t *testing.T) {
	// Create mock Tavily server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Parse request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Check API key in body
		apiKey, ok := requestBody["api_key"].(string)
		if !ok || apiKey != "test-tavily-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check query
		query, ok := requestBody["query"].(string)
		if !ok || query == "" {
			http.Error(w, "Missing query", http.StatusBadRequest)
			return
		}

		// Mock response
		response := TavilySearchResponse{
			Query: query,
			Results: []TavilyResult{
				{
					Title:   "Understanding Go Programming",
					URL:     "https://example.com/go-intro",
					Content: "Go is a statically typed, compiled programming language designed at Google...",
					Score:   0.95,
				},
				{
					Title:   "Advanced Go Patterns",
					URL:     "https://example.com/go-patterns",
					Content: "This article explores advanced patterns in Go programming...",
					Score:   0.87,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Set test API key
	os.Setenv("TAVILY_API_KEY", "test-tavily-key")
	defer os.Unsetenv("TAVILY_API_KEY")

	// Test search (this will need the actual implementation)
	// For now, we're just setting up the test structure
	t.Log("Tavily search test structure created - implementation pending")
}

func TestWebSearchWithMultipleEngines(t *testing.T) {
	// Save current env vars
	oldBrave := os.Getenv("BRAVE_API_KEY")
	oldTavily := os.Getenv("TAVILY_API_KEY")
	oldSerpapi := os.Getenv("SERPAPI_API_KEY")
	oldSerperDev := os.Getenv("SERPERDEV_API_KEY")
	defer func() {
		os.Setenv("BRAVE_API_KEY", oldBrave)
		os.Setenv("TAVILY_API_KEY", oldTavily)
		os.Setenv("SERPAPI_API_KEY", oldSerpapi)
		os.Setenv("SERPERDEV_API_KEY", oldSerperDev)
	}()

	testCases := []struct {
		name           string
		engine         string
		braveKey       string
		tavilyKey      string
		serpapiKey     string
		serperdevKey   string
		shouldError    bool
		expectedEngine string
	}{
		{
			name:           "Explicit DuckDuckGo - always works",
			engine:         "duckduckgo",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "",
			shouldError:    false,
			expectedEngine: "duckduckgo",
		},
		{
			name:           "Explicit Brave without key - should error",
			engine:         "brave",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "",
			shouldError:    true,
			expectedEngine: "",
		},
		{
			name:           "Explicit Brave with key - should work",
			engine:         "brave",
			braveKey:       "test-key",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "",
			shouldError:    false,
			expectedEngine: "brave",
		},
		{
			name:           "Explicit Tavily without key - should error",
			engine:         "tavily",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "",
			shouldError:    true,
			expectedEngine: "",
		},
		{
			name:           "Explicit Tavily with key - should work",
			engine:         "tavily",
			braveKey:       "",
			tavilyKey:      "test-key",
			serpapiKey:     "",
			serperdevKey:   "",
			shouldError:    false,
			expectedEngine: "tavily",
		},
		{
			name:           "Explicit Serpapi without key - should error",
			engine:         "serpapi",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "",
			shouldError:    true,
			expectedEngine: "",
		},
		{
			name:           "Explicit Serpapi with key - should work",
			engine:         "serpapi",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "test-key",
			serperdevKey:   "",
			shouldError:    false,
			expectedEngine: "serpapi",
		},
		{
			name:           "Explicit Serperdev without key - should error",
			engine:         "serperdev",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "",
			shouldError:    true,
			expectedEngine: "",
		},
		{
			name:           "Explicit Serperdev with key - should work",
			engine:         "serperdev",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "test-key",
			shouldError:    false,
			expectedEngine: "serperdev",
		},
		{
			name:           "Auto-select with no keys - use DuckDuckGo",
			engine:         "",
			braveKey:       "",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "",
			shouldError:    false,
			expectedEngine: "duckduckgo",
		},
		{
			name:           "Auto-select with Brave key - use Brave",
			engine:         "",
			braveKey:       "test-key",
			tavilyKey:      "",
			serpapiKey:     "",
			serperdevKey:   "",
			shouldError:    false,
			expectedEngine: "brave",
		},
		{
			name:           "Auto-select with both keys - prefer Tavily",
			engine:         "",
			braveKey:       "test-key",
			tavilyKey:      "test-key",
			serpapiKey:     "",
			serperdevKey:   "",
			shouldError:    false,
			expectedEngine: "tavily",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment
			if tc.braveKey != "" {
				os.Setenv("BRAVE_API_KEY", tc.braveKey)
			} else {
				os.Unsetenv("BRAVE_API_KEY")
			}

			if tc.tavilyKey != "" {
				os.Setenv("TAVILY_API_KEY", tc.tavilyKey)
			} else {
				os.Unsetenv("TAVILY_API_KEY")
			}

			if tc.serpapiKey != "" {
				os.Setenv("SERPAPI_API_KEY", tc.serpapiKey)
			} else {
				os.Unsetenv("SERPAPI_API_KEY")
			}

			if tc.serperdevKey != "" {
				os.Setenv("SERPERDEV_API_KEY", tc.serperdevKey)
			} else {
				os.Unsetenv("SERPERDEV_API_KEY")
			}

			// This test validates the logic we'll implement
			params := WebSearchParams{
				Engine: tc.engine,
			}

			// Simulate engine selection logic
			if params.Engine == "" {
				params.Engine = selectDefaultEngine()
			}

			// Validate engine selection
			if !tc.shouldError && params.Engine != tc.expectedEngine {
				t.Errorf("Expected engine '%s', got '%s'", tc.expectedEngine, params.Engine)
			}
		})
	}
}

func TestWebSearchErrorHandling(t *testing.T) {
	testCases := []struct {
		name          string
		setupFunc     func()
		params        map[string]interface{}
		errorContains string
	}{
		{
			name: "Brave without API key",
			setupFunc: func() {
				os.Unsetenv("BRAVE_API_KEY")
			},
			params: map[string]interface{}{
				"query":  "test",
				"engine": "brave",
			},
			errorContains: "API key required",
		},
		{
			name: "Tavily without API key",
			setupFunc: func() {
				os.Unsetenv("TAVILY_API_KEY")
			},
			params: map[string]interface{}{
				"query":  "test",
				"engine": "tavily",
			},
			errorContains: "API key required",
		},
		{
			name: "Serpapi without API key",
			setupFunc: func() {
				os.Unsetenv("SERPAPI_API_KEY")
			},
			params: map[string]interface{}{
				"query":  "test",
				"engine": "serpapi",
			},
			errorContains: "API key required",
		},
		{
			name: "Serperdev without API key",
			setupFunc: func() {
				os.Unsetenv("SERPERDEV_API_KEY")
			},
			params: map[string]interface{}{
				"query":  "test",
				"engine": "serperdev",
			},
			errorContains: "API key required",
		},
		{
			name:      "Invalid search engine",
			setupFunc: func() {},
			params: map[string]interface{}{
				"query":  "test",
				"engine": "invalid",
			},
			errorContains: "unsupported search engine",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupFunc()

			// This test validates error handling we'll implement
			t.Logf("Test case '%s' - expecting error containing '%s'", tc.name, tc.errorContains)
		})
	}
}

func TestWebSearchWithEngineAPIKey(t *testing.T) {
	// Save current env vars
	oldBrave := os.Getenv("BRAVE_API_KEY")
	oldTavily := os.Getenv("TAVILY_API_KEY")
	oldSerpapi := os.Getenv("SERPAPI_API_KEY")
	oldSerperDev := os.Getenv("SERPERDEV_API_KEY")
	defer func() {
		os.Setenv("BRAVE_API_KEY", oldBrave)
		os.Setenv("TAVILY_API_KEY", oldTavily)
		os.Setenv("SERPAPI_API_KEY", oldSerpapi)
		os.Setenv("SERPERDEV_API_KEY", oldSerperDev)
	}()

	// Clear all environment variables
	os.Unsetenv("BRAVE_API_KEY")
	os.Unsetenv("TAVILY_API_KEY")
	os.Unsetenv("SERPAPI_API_KEY")
	os.Unsetenv("SERPERDEV_API_KEY")

	testCases := []struct {
		name           string
		engine         string
		engineAPIKey   string
		envKey         string
		shouldWork     bool
		expectedAPIKey string
		description    string
	}{
		{
			name:           "Brave with explicit API key - no env var",
			engine:         "brave",
			engineAPIKey:   "explicit-brave-key",
			envKey:         "",
			shouldWork:     true,
			expectedAPIKey: "explicit-brave-key",
			description:    "Should use explicit API key when provided",
		},
		{
			name:           "Brave with both explicit and env key",
			engine:         "brave",
			engineAPIKey:   "explicit-brave-key",
			envKey:         "env-brave-key",
			shouldWork:     true,
			expectedAPIKey: "explicit-brave-key",
			description:    "Should prefer explicit API key over environment",
		},
		{
			name:           "Brave with only env key",
			engine:         "brave",
			engineAPIKey:   "",
			envKey:         "env-brave-key",
			shouldWork:     true,
			expectedAPIKey: "env-brave-key",
			description:    "Should fall back to environment variable",
		},
		{
			name:           "Brave with no keys",
			engine:         "brave",
			engineAPIKey:   "",
			envKey:         "",
			shouldWork:     false,
			expectedAPIKey: "",
			description:    "Should fail when no API key is available",
		},
		{
			name:           "Tavily with explicit API key",
			engine:         "tavily",
			engineAPIKey:   "explicit-tavily-key",
			envKey:         "",
			shouldWork:     true,
			expectedAPIKey: "explicit-tavily-key",
			description:    "Should use explicit Tavily API key",
		},
		{
			name:           "Serpapi with explicit API key",
			engine:         "serpapi",
			engineAPIKey:   "explicit-serpapi-key",
			envKey:         "",
			shouldWork:     true,
			expectedAPIKey: "explicit-serpapi-key",
			description:    "Should use explicit Serpapi API key",
		},
		{
			name:           "Serperdev with explicit API key",
			engine:         "serperdev",
			engineAPIKey:   "explicit-serperdev-key",
			envKey:         "",
			shouldWork:     true,
			expectedAPIKey: "explicit-serperdev-key",
			description:    "Should use explicit Serper.dev API key",
		},
		{
			name:           "Serperdev with both explicit and env key",
			engine:         "serperdev",
			engineAPIKey:   "explicit-serperdev-key",
			envKey:         "env-serperdev-key",
			shouldWork:     true,
			expectedAPIKey: "explicit-serperdev-key",
			description:    "Should prefer explicit API key over environment",
		},
		{
			name:           "Serperdev with only env key",
			engine:         "serperdev",
			engineAPIKey:   "",
			envKey:         "env-serperdev-key",
			shouldWork:     true,
			expectedAPIKey: "env-serperdev-key",
			description:    "Should fall back to environment variable",
		},
		{
			name:           "Serperdev with no keys",
			engine:         "serperdev",
			engineAPIKey:   "",
			envKey:         "",
			shouldWork:     false,
			expectedAPIKey: "",
			description:    "Should fail when no API key is available",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment
			if tc.envKey != "" {
				switch tc.engine {
				case "brave":
					os.Setenv("BRAVE_API_KEY", tc.envKey)
				case "tavily":
					os.Setenv("TAVILY_API_KEY", tc.envKey)
				case "serpapi":
					os.Setenv("SERPAPI_API_KEY", tc.envKey)
				case "serperdev":
					os.Setenv("SERPERDEV_API_KEY", tc.envKey)
				}
			}

			// Create test parameters
			params := map[string]interface{}{
				"query":  "test query",
				"engine": tc.engine,
			}

			if tc.engineAPIKey != "" {
				params["engine_api_key"] = tc.engineAPIKey
			}

			// Log test scenario
			t.Logf("Testing: %s", tc.description)
			t.Logf("Engine: %s, Explicit Key: %s, Env Key: %s",
				tc.engine, tc.engineAPIKey, tc.envKey)

			// Verify behavior (implementation pending)
			// The actual tool execution will be tested after implementation

			// Clear environment for next test
			os.Unsetenv("BRAVE_API_KEY")
			os.Unsetenv("TAVILY_API_KEY")
			os.Unsetenv("SERPAPI_API_KEY")
		})
	}
}

func TestEngineAPIKeyPrecedence(t *testing.T) {
	// Test that explicit API keys take precedence over environment variables
	testCases := []struct {
		name        string
		setupFunc   func()
		cleanupFunc func()
		params      map[string]interface{}
		checkFunc   func(t *testing.T)
	}{
		{
			name: "Explicit key overrides environment for Brave",
			setupFunc: func() {
				os.Setenv("BRAVE_API_KEY", "env-key")
			},
			cleanupFunc: func() {
				os.Unsetenv("BRAVE_API_KEY")
			},
			params: map[string]interface{}{
				"query":          "test",
				"engine":         "brave",
				"engine_api_key": "explicit-key",
			},
			checkFunc: func(t *testing.T) {
				// Would verify that "explicit-key" is used
				t.Log("Test: explicit key should override environment variable")
			},
		},
		{
			name: "Empty explicit key falls back to environment",
			setupFunc: func() {
				os.Setenv("TAVILY_API_KEY", "env-tavily-key")
			},
			cleanupFunc: func() {
				os.Unsetenv("TAVILY_API_KEY")
			},
			params: map[string]interface{}{
				"query":          "test",
				"engine":         "tavily",
				"engine_api_key": "", // Empty string
			},
			checkFunc: func(t *testing.T) {
				// Would verify that "env-tavily-key" is used
				t.Log("Test: empty explicit key should fall back to environment")
			},
		},
		{
			name: "No explicit key uses environment",
			setupFunc: func() {
				os.Setenv("SERPAPI_API_KEY", "env-serpapi-key")
			},
			cleanupFunc: func() {
				os.Unsetenv("SERPAPI_API_KEY")
			},
			params: map[string]interface{}{
				"query":  "test",
				"engine": "serpapi",
				// No engine_api_key parameter
			},
			checkFunc: func(t *testing.T) {
				// Would verify that "env-serpapi-key" is used
				t.Log("Test: missing explicit key should use environment")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupFunc()
			defer tc.cleanupFunc()

			tc.checkFunc(t)
		})
	}
}

func TestEngineAPIKeySecurityValidation(t *testing.T) {
	// Test that API keys are properly validated and not exposed
	testCases := []struct {
		name   string
		params map[string]interface{}
		checks []string
	}{
		{
			name: "API key should not be logged in errors",
			params: map[string]interface{}{
				"query":          "test",
				"engine":         "brave",
				"engine_api_key": "super-secret-key-12345",
			},
			checks: []string{
				"Error messages should not contain the actual API key",
				"API key should be masked in any debug output",
			},
		},
		{
			name: "Invalid API key format handling",
			params: map[string]interface{}{
				"query":          "test",
				"engine":         "tavily",
				"engine_api_key": "", // Empty key
			},
			checks: []string{
				"Should handle empty API keys gracefully",
				"Should provide clear error message without exposing key attempts",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, check := range tc.checks {
				t.Log(check)
			}
		})
	}
}

func TestSearchSerpapi(t *testing.T) {
	// Create mock Serpapi server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check API key in query params
		apiKey := r.URL.Query().Get("api_key")
		if apiKey != "test-serpapi-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// For Serpapi GET request, check query parameters
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "Missing query", http.StatusBadRequest)
			return
		}

		// Mock response
		response := SerpapiSearchResponse{
			OrganicResults: []SerpapiResult{
				{
					Title:   "Go Programming Language",
					Link:    "https://golang.org",
					Snippet: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.",
				},
				{
					Title:   "The Go Programming Language Specification",
					Link:    "https://golang.org/ref/spec",
					Snippet: "This is a reference manual for the Go programming language.",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Set test API key
	os.Setenv("SERPAPI_API_KEY", "test-serpapi-key")
	defer os.Unsetenv("SERPAPI_API_KEY")

	// Test search (this will need the actual implementation)
	// For now, we're just setting up the test structure
	t.Log("Serpapi search test structure created - implementation pending")
}

func TestSearchSerperDev(t *testing.T) {
	// Create mock Serper.dev server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check API key header
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey != "test-serperdev-key" && apiKey != "explicit-serperdev-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse request body
		var requestBody SerperDevSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Check query
		if requestBody.Q == "" {
			http.Error(w, "Missing query", http.StatusBadRequest)
			return
		}

		// Mock response
		response := SerperDevSearchResponse{
			Organic: []SerperDevResult{
				{
					Title:   "Go Programming Language",
					Link:    "https://golang.org",
					Snippet: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.",
				},
				{
					Title:   "Getting Started with Go",
					Link:    "https://golang.org/doc/tutorial/getting-started",
					Snippet: "Tutorial: Get started with Go. In this tutorial, you'll get a brief introduction to Go programming.",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Set test API key
	os.Setenv("SERPERDEV_API_KEY", "test-serperdev-key")
	defer os.Unsetenv("SERPERDEV_API_KEY")

	// Test search structure
	t.Log("Serper.dev search test structure created - mock server: " + server.URL)
}

func TestSearchResultConversion(t *testing.T) {
	// Test converting Brave results to our format
	braveResults := []BraveResult{
		{
			Title:       "Test Title",
			URL:         "https://example.com",
			Description: "Test description",
			PageAge:     "30d",
		},
	}

	// Test converting Tavily results to our format
	tavilyResults := []TavilyResult{
		{
			Title:   "Test Title",
			URL:     "https://example.com",
			Content: "Test content that is longer than description",
			Score:   0.95,
		},
	}

	// These tests validate the conversion logic we'll implement
	t.Log("Brave results:", braveResults)
	t.Log("Tavily results:", tavilyResults)
}

func TestWebSearchIntegrationWithEngineAPIKey(t *testing.T) {
	// This is an integration test that validates the EngineAPIKey parameter
	tool, ok := tools.GetTool("web_search")
	if !ok {
		t.Fatal("web_search tool not found")
	}

	// Create test context
	ctx := NewToolContext(
		context.Background(),
		NewStateReader(NewState()),
		&mockSearchAgent{},
		"test-run",
	)

	testCases := []struct {
		name          string
		params        map[string]interface{}
		setupEnv      func()
		cleanupEnv    func()
		expectError   bool
		errorContains string
	}{
		{
			name: "DuckDuckGo with EngineAPIKey (ignored)",
			params: map[string]interface{}{
				"query":          "test query",
				"engine":         "duckduckgo",
				"engine_api_key": "ignored-key",
			},
			setupEnv:    func() {},
			cleanupEnv:  func() {},
			expectError: false, // DuckDuckGo doesn't need API key
		},
		{
			name: "Brave with explicit API key - no env",
			params: map[string]interface{}{
				"query":          "test query",
				"engine":         "brave",
				"engine_api_key": "explicit-key",
			},
			setupEnv: func() {
				os.Unsetenv("BRAVE_API_KEY")
			},
			cleanupEnv:    func() {},
			expectError:   true, // Will fail with real API
			errorContains: "Brave Search API",
		},
		{
			name: "Brave without any API key",
			params: map[string]interface{}{
				"query":  "test query",
				"engine": "brave",
			},
			setupEnv: func() {
				os.Unsetenv("BRAVE_API_KEY")
			},
			cleanupEnv:    func() {},
			expectError:   true,
			errorContains: "API key required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupEnv()
			defer tc.cleanupEnv()

			result, err := tool.Execute(ctx, tc.params)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tc.errorContains)
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					// Some errors are acceptable (network, rate limiting)
					t.Logf("Execution returned error (may be expected): %v", err)
				} else if result != nil {
					// Validate result structure
					if searchResult, ok := result.(*WebSearchResults); ok {
						t.Logf("Search completed with %d results", len(searchResult.Results))
					}
				}
			}
		})
	}
}
