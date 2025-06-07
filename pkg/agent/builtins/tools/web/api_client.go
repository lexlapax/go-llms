// ABOUTME: API Client Tool for LLM-friendly REST API interactions with auth and error handling
// ABOUTME: Supports OpenAPI discovery, multiple auth methods, and intelligent error guidance

package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func init() {
	tools.MustRegisterTool("api_client", createAPIClientTool(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "api_client",
			Category:    "web",
			Tags:        []string{"api", "rest", "http", "integration", "client"},
			Description: "Make REST API calls with automatic error handling and authentication support",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Fetch GitHub User",
					Description: "Get public information about a GitHub user",
					Code: `// Fetch GitHub user information
params := map[string]interface{}{
    "base_url": "https://api.github.com",
    "endpoint": "/users/octocat",
    "method": "GET"
}
// Returns: {"success": true, "status_code": 200, "data": {"login": "octocat", "name": "The Octocat"}}`,
				},
				{
					Name:        "Authenticated POST",
					Description: "Create a resource with API key authentication",
					Code: `// Create item with API key
params := map[string]interface{}{
    "base_url": "https://api.example.com",
    "endpoint": "/items",
    "method": "POST",
    "auth": map[string]interface{}{
        "type": "api_key",
        "api_key": "your-key",
        "key_location": "header",
        "key_name": "X-API-Key"
    },
    "body": map[string]interface{}{
        "name": "New Item",
        "description": "Created via API"
    }
}
// Returns: {"success": true, "status_code": 201, "data": {"id": "12345", "created": true}}`,
				},
			},
		},
	})
}

// createAPIClientTool creates and configures the API client tool
func createAPIClientTool() domain.Tool {
	paramSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"base_url": {
				Type:        "string",
				Description: "Base URL of the API (e.g., 'https://api.example.com')",
			},
			"endpoint": {
				Type:        "string",
				Description: "API endpoint path (e.g., '/users/{user_id}'). Use {param} for path parameters",
			},
			"method": {
				Type:        "string",
				Description: "HTTP method (GET, POST, PUT, DELETE, PATCH). Defaults to GET",
				Enum:        []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
			},
			"path_params": {
				Type:        "object",
				Description: "Path parameters to substitute in the endpoint (e.g., {'user_id': '123'})",
			},
			"query_params": {
				Type:        "object",
				Description: "Query parameters to append to the URL",
			},
			"headers": {
				Type:        "object",
				Description: "HTTP headers to include in the request",
			},
			"body": {
				Type:        "object",
				Description: "Request body (will be JSON encoded). Only for POST, PUT, PATCH",
			},
			"auth": {
				Type:        "object",
				Description: "Authentication configuration",
				Properties: map[string]sdomain.Property{
					"type": {
						Type:        "string",
						Description: "Authentication type: 'api_key', 'bearer', 'basic'",
						Enum:        []string{"api_key", "bearer", "basic"},
					},
					"api_key": {
						Type:        "string",
						Description: "API key value (for api_key auth)",
					},
					"key_location": {
						Type:        "string",
						Description: "Where to place the API key: 'header' or 'query'",
						Enum:        []string{"header", "query"},
					},
					"key_name": {
						Type:        "string",
						Description: "Name of the API key parameter (e.g., 'X-API-Key')",
					},
					"token": {
						Type:        "string",
						Description: "Bearer token value (for bearer auth)",
					},
					"username": {
						Type:        "string",
						Description: "Username (for basic auth)",
					},
					"password": {
						Type:        "string",
						Description: "Password (for basic auth)",
					},
				},
			},
			"timeout": {
				Type:        "string",
				Description: "Request timeout (e.g., '30s', '1m'). Default is 30s",
			},
		},
		Required: []string{"base_url", "endpoint"},
	}

	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"success": {
				Type:        "boolean",
				Description: "Whether the API call was successful",
			},
			"status_code": {
				Type:        "number",
				Description: "HTTP status code",
			},
			"data": {
				Type:        "object",
				Description: "Response data (parsed from JSON)",
			},
			"error_message": {
				Type:        "string",
				Description: "Error message if the request failed",
			},
			"error_details": {
				Type:        "object",
				Description: "Detailed error information from the API",
			},
			"error_guidance": {
				Type:        "string",
				Description: "Helpful guidance for resolving the error",
			},
			"headers": {
				Type:        "object",
				Description: "Response headers",
			},
		},
		Required: []string{"success", "status_code"},
	}

	builder := atools.NewToolBuilder("api_client", "Make REST API calls with automatic error handling and authentication support").
		WithFunction(executeAPIClient).
		WithParameterSchema(paramSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`Use this tool to interact with REST APIs. It handles:
- Multiple authentication methods (API key, Bearer token, Basic auth)
- Automatic JSON encoding/decoding
- Path parameter substitution
- Helpful error messages and guidance
- Common HTTP methods (GET, POST, PUT, DELETE, etc.)

The tool will automatically set appropriate headers and handle responses intelligently.`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Simple GET request",
				Description: "Fetch user data from an API",
				Scenario:    "When you need to retrieve data from a REST API",
				Input: map[string]interface{}{
					"base_url": "https://api.github.com",
					"endpoint": "/users/octocat",
					"method":   "GET",
				},
				Output: map[string]interface{}{
					"success":     true,
					"status_code": 200,
					"data": map[string]interface{}{
						"login": "octocat",
						"name":  "The Octocat",
					},
				},
				Explanation: "Makes a GET request to fetch GitHub user information",
			},
			{
				Name:        "POST with authentication",
				Description: "Create a resource with API key authentication",
				Scenario:    "When you need to create data in an API that requires authentication",
				Input: map[string]interface{}{
					"base_url": "https://api.example.com",
					"endpoint": "/items",
					"method":   "POST",
					"auth": map[string]interface{}{
						"type":         "api_key",
						"api_key":      "your-api-key",
						"key_location": "header",
						"key_name":     "X-API-Key",
					},
					"body": map[string]interface{}{
						"name":        "New Item",
						"description": "Created via API",
					},
				},
				Output: map[string]interface{}{
					"success":     true,
					"status_code": 201,
					"data": map[string]interface{}{
						"id":      "12345",
						"created": true,
					},
				},
				Explanation: "Creates a new item with API key authentication",
			},
			{
				Name:        "Path parameters",
				Description: "Use path parameters in the endpoint",
				Scenario:    "When the API endpoint contains variable segments",
				Input: map[string]interface{}{
					"base_url": "https://api.example.com",
					"endpoint": "/users/{user_id}/posts/{post_id}",
					"method":   "GET",
					"path_params": map[string]string{
						"user_id": "alice",
						"post_id": "42",
					},
				},
				Output: map[string]interface{}{
					"success":     true,
					"status_code": 200,
					"data": map[string]interface{}{
						"title":   "My Post",
						"content": "Post content here",
					},
				},
				Explanation: "Substitutes path parameters to create the final URL",
			},
		}).
		WithConstraints([]string{
			"Only JSON request/response bodies are currently supported",
			"Authentication credentials should be kept secure",
			"Rate limiting is handled by returning 429 status codes",
			"Redirects are followed automatically up to 10 times",
		}).
		WithErrorGuidance(map[string]string{
			"invalid_url":       "Ensure the base_url is a valid HTTP/HTTPS URL",
			"connection_failed": "Check if the API server is accessible and the URL is correct",
			"auth_required":     "This endpoint requires authentication. Add auth configuration",
			"invalid_json":      "The response was not valid JSON. The API might return a different format",
			"timeout":           "The request timed out. Try increasing the timeout parameter",
			"invalid_method":    "Use one of: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS",
		}).
		WithCategory("web").
		WithTags([]string{"api", "rest", "http", "integration"}).
		WithVersion("1.0.0").
		WithBehavior(false, false, false, "medium")

	return builder.Build()
}

// NewAPIClientTool creates a new instance of the API client tool
func NewAPIClientTool() domain.Tool {
	return createAPIClientTool()
}

// convertToStringMap converts various map types to map[string]string
func convertToStringMap(input interface{}) map[string]string {
	result := make(map[string]string)

	switch m := input.(type) {
	case map[string]string:
		return m
	case map[string]interface{}:
		for k, v := range m {
			result[k] = fmt.Sprintf("%v", v)
		}
	case map[interface{}]interface{}:
		for k, v := range m {
			result[fmt.Sprintf("%v", k)] = fmt.Sprintf("%v", v)
		}
	}

	return result
}

// executeAPIClient executes the API client request
func executeAPIClient(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
	// Parse parameters
	paramMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("parameters must be a map")
	}

	// Validate required parameters
	baseURL, ok := paramMap["base_url"].(string)
	if !ok || baseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}

	// Parse and validate base URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, fmt.Errorf("invalid base_url: must be a valid HTTP/HTTPS URL")
	}

	endpoint, ok := paramMap["endpoint"].(string)
	if !ok || endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	// Get method (default to GET)
	method := "GET"
	if m, ok := paramMap["method"].(string); ok {
		method = strings.ToUpper(m)
	}

	// Validate method
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "DELETE": true,
		"PATCH": true, "HEAD": true, "OPTIONS": true,
	}
	if !validMethods[method] {
		return nil, fmt.Errorf("invalid method: %s", method)
	}

	// Process path parameters
	if pathParams := paramMap["path_params"]; pathParams != nil {
		pathParamMap := convertToStringMap(pathParams)
		for key, value := range pathParamMap {
			placeholder := fmt.Sprintf("{%s}", key)
			endpoint = strings.ReplaceAll(endpoint, placeholder, value)
		}
	}

	// Build full URL
	fullURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(endpoint, "/")

	// Add query parameters
	if queryParams := paramMap["query_params"]; queryParams != nil {
		queryParamMap := convertToStringMap(queryParams)
		values := url.Values{}
		for key, value := range queryParamMap {
			values.Add(key, value)
		}
		if len(values) > 0 {
			fullURL += "?" + values.Encode()
		}
	}

	// Prepare request body
	var bodyReader io.Reader
	if body, ok := paramMap["body"]; ok && body != nil {
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return nil, fmt.Errorf("body not allowed for %s requests", method)
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to encode body as JSON: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx.Context, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", "go-llms-api-client/1.0")
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Add custom headers
	if headers := paramMap["headers"]; headers != nil {
		headerMap := convertToStringMap(headers)
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	// Handle authentication
	if auth, ok := paramMap["auth"].(map[string]interface{}); ok {
		if err := applyAuthentication(req, auth); err != nil {
			return nil, err
		}
	}

	// Set timeout
	timeout := 30 * time.Second
	if t, ok := paramMap["timeout"].(string); ok {
		if parsed, err := time.ParseDuration(t); err == nil {
			timeout = parsed
		}
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: timeout,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"success":        false,
			"error_message":  fmt.Sprintf("Request failed: %v", err),
			"error_guidance": "Check if the API server is accessible and the URL is correct",
		}, nil
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"success":       false,
			"status_code":   float64(resp.StatusCode),
			"error_message": fmt.Sprintf("Failed to read response: %v", err),
		}, nil
	}

	// Parse response
	result := map[string]interface{}{
		"status_code": float64(resp.StatusCode),
		"success":     resp.StatusCode >= 200 && resp.StatusCode < 300,
		"headers":     extractHeaders(resp.Header),
	}

	// Try to parse as JSON
	var responseData interface{}
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
			// If not JSON, return as string
			result["data"] = string(bodyBytes)
			result["error_message"] = "Response is not valid JSON"
		} else {
			result["data"] = responseData
		}
	}

	// Add error details and guidance for non-success responses
	if !result["success"].(bool) {
		result["error_message"] = fmt.Sprintf("API returned status %d", resp.StatusCode)

		// Add error details if response contains error info
		if errorData, ok := responseData.(map[string]interface{}); ok {
			result["error_details"] = errorData
		}

		// Add helpful guidance based on status code
		result["error_guidance"] = getErrorGuidance(resp.StatusCode)
	}

	return result, nil
}

// applyAuthentication applies the specified authentication to the request
func applyAuthentication(req *http.Request, auth map[string]interface{}) error {
	authType, ok := auth["type"].(string)
	if !ok {
		return fmt.Errorf("auth type is required")
	}

	switch authType {
	case "api_key":
		apiKey, ok := auth["api_key"].(string)
		if !ok || apiKey == "" {
			return fmt.Errorf("api_key is required for api_key auth")
		}

		keyLocation := "header"
		if loc, ok := auth["key_location"].(string); ok {
			keyLocation = loc
		}

		keyName := "X-API-Key"
		if name, ok := auth["key_name"].(string); ok {
			keyName = name
		}

		switch keyLocation {
		case "header":
			req.Header.Set(keyName, apiKey)
		case "query":
			q := req.URL.Query()
			q.Set(keyName, apiKey)
			req.URL.RawQuery = q.Encode()
		default:
			return fmt.Errorf("invalid key_location: %s", keyLocation)
		}

	case "bearer":
		token, ok := auth["token"].(string)
		if !ok || token == "" {
			return fmt.Errorf("token is required for bearer auth")
		}
		req.Header.Set("Authorization", "Bearer "+token)

	case "basic":
		username, ok1 := auth["username"].(string)
		password, ok2 := auth["password"].(string)
		if !ok1 || !ok2 || username == "" {
			return fmt.Errorf("username and password are required for basic auth")
		}
		req.SetBasicAuth(username, password)

	default:
		return fmt.Errorf("unsupported auth type: %s", authType)
	}

	return nil
}

// extractHeaders converts http.Header to a simple map
func extractHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

// getErrorGuidance returns helpful guidance based on HTTP status code
func getErrorGuidance(statusCode int) string {
	switch statusCode {
	case 400:
		return "Bad request. Check that all required parameters are provided and properly formatted."
	case 401:
		return "Authentication required. Provide valid credentials using the 'auth' parameter."
	case 403:
		return "Access forbidden. Your credentials may not have permission for this resource."
	case 404:
		return "Resource not found. Verify the endpoint path and any path parameters."
	case 405:
		return "Method not allowed. This endpoint doesn't support the HTTP method you used."
	case 429:
		return "Rate limit exceeded. Wait before making more requests or implement rate limiting."
	case 500, 502, 503, 504:
		return "Server error. This is a server-side issue. You may retry after a short wait."
	default:
		if statusCode >= 400 && statusCode < 500 {
			return "Client error. Review your request parameters and authentication."
		} else if statusCode >= 500 {
			return "Server error. The API server is experiencing issues."
		}
		return "Unexpected status code. Check the API documentation."
	}
}
