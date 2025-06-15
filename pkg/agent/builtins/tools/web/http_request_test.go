// ABOUTME: Tests for the HTTPRequest built-in tool
// ABOUTME: Validates various HTTP methods, authentication, headers, and body handling

package web

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// createTestToolContextForHTTP creates a test ToolContext for HTTP tests
func createTestToolContextForHTTP() *domain.ToolContext {
	// Create a mock agent
	agent := mocks.NewMockAgent("Test HTTP Agent")

	// Create a test state
	state := domain.NewState()
	stateReader := domain.NewStateReader(state)

	// Create the tool context
	tc := domain.NewToolContext(context.Background(), stateReader, agent, "test-run-id")

	return tc
}

func TestHTTPRequestRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("http_request")
	if !ok {
		t.Fatal("HTTPRequest tool not registered")
	}
	if tool == nil {
		t.Fatal("HTTPRequest tool is nil")
	}

	// Test tool name
	if tool.Name() != "http_request" {
		t.Errorf("Expected tool name 'http_request', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("http_request")
	if len(entries) == 0 {
		t.Fatal("HTTPRequest tool not found in registry")
	}

	meta := entries[0].Metadata
	if meta.Category != "web" {
		t.Errorf("Expected category 'web', got '%s'", meta.Category)
	}
}

func TestHTTPRequestMethods(t *testing.T) {
	// Create test server that echoes request details
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"method":  r.Method,
			"url":     r.URL.String(),
			"headers": r.Header,
			"body":    "",
		}

		if r.Body != nil {
			body, _ := io.ReadAll(r.Body)
			response["body"] = string(body)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tool := HTTPRequest()
	ctx := createTestToolContextForHTTP()

	testCases := []struct {
		name    string
		method  string
		body    string
		hasBody bool
	}{
		{"GET", "GET", "", false},
		{"POST", "POST", `{"test": "data"}`, true},
		{"PUT", "PUT", `{"update": "data"}`, true},
		{"DELETE", "DELETE", "", false},
		{"PATCH", "PATCH", `{"patch": "data"}`, true},
		{"HEAD", "HEAD", "", false},
		{"OPTIONS", "OPTIONS", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := map[string]interface{}{
				"url":    server.URL,
				"method": tc.method,
			}

			if tc.hasBody {
				params["body"] = tc.body
				params["body_type"] = "json"
			}

			result, err := tool.Execute(ctx, params)
			if err != nil {
				t.Fatalf("Failed to execute %s request: %v", tc.method, err)
			}

			httpResult, ok := result.(*HTTPRequestResult)
			if !ok {
				t.Fatalf("Result is not HTTPRequestResult: %T", result)
			}

			// Parse echo response (skip for HEAD requests which have no body)
			if tc.method != "HEAD" {
				var echoResp map[string]interface{}
				if err := json.Unmarshal([]byte(httpResult.Body), &echoResp); err != nil {
					t.Fatalf("Failed to parse echo response: %v", err)
				}

				// Verify method
				if echoResp["method"] != tc.method {
					t.Errorf("Expected method %s, got %s", tc.method, echoResp["method"])
				}

				// Verify body if applicable
				if tc.hasBody {
					if echoResp["body"] != tc.body {
						t.Errorf("Expected body %s, got %s", tc.body, echoResp["body"])
					}
				}
			}
		})
	}
}

func TestHTTPRequestAuthentication(t *testing.T) {
	tool := HTTPRequest()
	ctx := createTestToolContextForHTTP()

	t.Run("BasicAuth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Parse basic auth
			if strings.HasPrefix(auth, "Basic ") {
				encoded := strings.TrimPrefix(auth, "Basic ")
				decoded, _ := base64.StdEncoding.DecodeString(encoded)
				parts := strings.SplitN(string(decoded), ":", 2)

				if len(parts) == 2 && parts[0] == "testuser" && parts[1] == "testpass" {
					_, _ = w.Write([]byte("Authenticated"))
					return
				}
			}

			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		result, err := tool.Execute(ctx, map[string]interface{}{
			"url":           server.URL,
			"auth_type":     "basic",
			"auth_username": "testuser",
			"auth_password": "testpass",
		})

		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}

		httpResult := result.(*HTTPRequestResult)
		if httpResult.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", httpResult.StatusCode)
			t.Logf("Response body: %s", httpResult.Body)
			t.Logf("Response headers: %+v", httpResult.Headers)
		}
		if httpResult.Body != "Authenticated" {
			t.Errorf("Expected 'Authenticated', got '%s'", httpResult.Body)
		}
	})

	t.Run("BearerAuth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "Bearer test-token-123" {
				_, _ = w.Write([]byte("Token Valid"))
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		result, err := tool.Execute(ctx, map[string]interface{}{
			"url":        server.URL,
			"auth_type":  "bearer",
			"auth_token": "test-token-123",
		})

		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}

		httpResult := result.(*HTTPRequestResult)
		if httpResult.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", httpResult.StatusCode)
		}
		if httpResult.Body != "Token Valid" {
			t.Errorf("Expected 'Token Valid', got '%s'", httpResult.Body)
		}
	})

	t.Run("APIKeyInHeader", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "secret-key-123" {
				_, _ = w.Write([]byte("API Key Valid"))
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		result, err := tool.Execute(ctx, map[string]interface{}{
			"url":               server.URL,
			"auth_type":         "api_key",
			"auth_key_name":     "X-API-Key",
			"auth_key_value":    "secret-key-123",
			"auth_key_location": "header",
		})

		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}

		httpResult := result.(*HTTPRequestResult)
		if httpResult.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", httpResult.StatusCode)
		}
	})

	t.Run("APIKeyInQuery", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.URL.Query().Get("api_key")
			if apiKey == "query-key-456" { //nolint:gosec // Test credential
				_, _ = w.Write([]byte("Query Key Valid"))
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		result, err := tool.Execute(ctx, map[string]interface{}{
			"url":               server.URL,
			"auth_type":         "api_key",
			"auth_key_name":     "api_key",
			"auth_key_value":    "query-key-456",
			"auth_key_location": "query",
		})

		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}

		httpResult := result.(*HTTPRequestResult)
		if httpResult.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", httpResult.StatusCode)
		}
	})
}

func TestHTTPRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back custom headers
		response := map[string]string{
			"received_custom_header": r.Header.Get("X-Custom-Header"),
			"received_accept":        r.Header.Get("Accept"),
			"received_user_agent":    r.Header.Get("User-Agent"),
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Response-Header", "test-value")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tool := HTTPRequest()
	ctx := createTestToolContextForHTTP()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"url": server.URL,
		"headers": map[string]interface{}{
			"X-Custom-Header": "custom-value",
			"Accept":          "application/json",
		},
	})

	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	httpResult := result.(*HTTPRequestResult)

	// Check response headers
	if httpResult.Headers["X-Response-Header"] != "test-value" {
		t.Errorf("Expected response header 'test-value', got '%s'", httpResult.Headers["X-Response-Header"])
	}

	// Check echoed headers
	var echoResp map[string]string
	if err := json.Unmarshal([]byte(httpResult.Body), &echoResp); err != nil {
		t.Fatal(err)
	}

	if echoResp["received_custom_header"] != "custom-value" {
		t.Errorf("Custom header not received correctly")
	}
	if echoResp["received_accept"] != "application/json" {
		t.Errorf("Accept header not received correctly")
	}
	if !strings.Contains(echoResp["received_user_agent"], "go-llms") {
		t.Errorf("User-Agent should contain 'go-llms'")
	}
}

func TestHTTPRequestQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo query parameters
		params := r.URL.Query()
		response := map[string][]string(params)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tool := HTTPRequest()
	ctx := createTestToolContextForHTTP()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"url": server.URL + "/api/test",
		"query_params": map[string]interface{}{
			"foo":    "bar",
			"page":   "1",
			"filter": "active",
		},
	})

	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	httpResult := result.(*HTTPRequestResult)

	var queryResp map[string][]string
	if err := json.Unmarshal([]byte(httpResult.Body), &queryResp); err != nil {
		t.Fatal(err)
	}

	// Check query parameters
	if queryResp["foo"][0] != "bar" {
		t.Error("Query param 'foo' not set correctly")
	}
	if queryResp["page"][0] != "1" {
		t.Error("Query param 'page' not set correctly")
	}
	if queryResp["filter"][0] != "active" {
		t.Error("Query param 'filter' not set correctly")
	}
}

func TestHTTPRequestBodyTypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)

		response := map[string]string{
			"content_type": contentType,
			"body":         string(body),
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tool := HTTPRequest()
	ctx := createTestToolContextForHTTP()

	testCases := []struct {
		name         string
		bodyType     string
		body         string
		expectedType string
	}{
		{
			name:         "JSON",
			bodyType:     "json",
			body:         `{"key": "value"}`,
			expectedType: "application/json",
		},
		{
			name:         "Form",
			bodyType:     "form",
			body:         "key1=value1&key2=value2",
			expectedType: "application/x-www-form-urlencoded",
		},
		{
			name:         "XML",
			bodyType:     "xml",
			body:         `<root><key>value</key></root>`,
			expectedType: "application/xml",
		},
		{
			name:         "Text",
			bodyType:     "text",
			body:         "Plain text content",
			expectedType: "text/plain",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, map[string]interface{}{
				"url":       server.URL,
				"method":    "POST",
				"body":      tc.body,
				"body_type": tc.bodyType,
			})

			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}

			httpResult := result.(*HTTPRequestResult)

			var resp map[string]string
			if err := json.Unmarshal([]byte(httpResult.Body), &resp); err != nil {
				t.Fatal(err)
			}

			if resp["content_type"] != tc.expectedType {
				t.Errorf("Expected content type %s, got %s", tc.expectedType, resp["content_type"])
			}
			if resp["body"] != tc.body {
				t.Errorf("Body mismatch: expected %s, got %s", tc.body, resp["body"])
			}
		})
	}
}

func TestHTTPRequestRedirects(t *testing.T) {
	// Create a server that redirects
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/target", http.StatusFound)
			return
		}
		if r.URL.Path == "/target" {
			_, _ = w.Write([]byte("Redirect target reached"))
			return
		}
	}))
	defer server.Close()

	tool := HTTPRequest()
	ctx := createTestToolContextForHTTP()

	t.Run("FollowRedirects", func(t *testing.T) {
		result, err := tool.Execute(ctx, map[string]interface{}{
			"url":              server.URL + "/redirect",
			"follow_redirects": true,
		})

		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}

		httpResult := result.(*HTTPRequestResult)
		if httpResult.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", httpResult.StatusCode)
		}
		if httpResult.Body != "Redirect target reached" {
			t.Errorf("Expected to reach redirect target, got: %s", httpResult.Body)
		}
	})

	t.Run("NoFollowRedirects", func(t *testing.T) {
		result, err := tool.Execute(ctx, map[string]interface{}{
			"url":              server.URL + "/redirect",
			"follow_redirects": false,
		})

		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}

		httpResult := result.(*HTTPRequestResult)
		if httpResult.StatusCode != 302 {
			t.Errorf("Expected status 302, got %d", httpResult.StatusCode)
		}
		if httpResult.RedirectURL != "/target" {
			t.Errorf("Expected redirect URL '/target', got '%s'", httpResult.RedirectURL)
		}
	})
}

func TestHTTPRequestErrors(t *testing.T) {
	tool := HTTPRequest()
	ctx := createTestToolContextForHTTP()

	t.Run("InvalidURL", func(t *testing.T) {
		_, err := tool.Execute(ctx, map[string]interface{}{
			"url": "not-a-valid-url",
		})
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		_, err := tool.Execute(ctx, map[string]interface{}{
			"url":    "https://example.com",
			"method": "INVALID",
		})
		if err == nil {
			t.Error("Expected error for invalid method")
		}
		if !strings.Contains(err.Error(), "invalid HTTP method") {
			t.Errorf("Error should mention invalid method: %v", err)
		}
	})

	t.Run("InvalidAuth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The tool no longer returns an error for invalid auth types
			// It simply doesn't set authentication
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}))
		defer server.Close()

		result, err := tool.Execute(ctx, map[string]interface{}{
			"url":       server.URL,
			"auth_type": "invalid_type",
		})
		// The tool should succeed, just without authentication
		if err != nil {
			t.Errorf("Unexpected error for invalid auth type: %v", err)
		}
		if result != nil {
			httpResult := result.(*HTTPRequestResult)
			if httpResult.StatusCode != 200 {
				t.Errorf("Expected status 200, got %d", httpResult.StatusCode)
			}
		}
	})
}
