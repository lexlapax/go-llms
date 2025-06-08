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
			Version:     "2.0.0",
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
			"openapi_spec": {
				Type:        "string",
				Description: "URL to OpenAPI/Swagger spec for automatic discovery and validation. When provided, enables operation discovery mode",
			},
			"discover_operations": {
				Type:        "boolean",
				Description: "If true, returns available operations from the OpenAPI spec instead of making an API call",
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
- OpenAPI/Swagger spec discovery and validation

The tool will automatically set appropriate headers and handle responses intelligently.

OpenAPI Discovery Mode:
- Set discover_operations=true and provide openapi_spec URL to discover available endpoints
- The tool will fetch and parse the OpenAPI spec to show all available operations
- Use this to understand what endpoints are available before making calls

OpenAPI Validation Mode:
- Provide openapi_spec URL with regular API calls to enable request validation
- The tool will validate parameters and request body against the spec
- Helps ensure API calls are correctly formatted before sending`).
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
			{
				Name:        "OpenAPI discovery",
				Description: "Discover available operations from OpenAPI spec",
				Scenario:    "When you need to explore what endpoints an API offers",
				Input: map[string]interface{}{
					"base_url":            "https://api.example.com",
					"endpoint":            "/not-used-in-discovery",
					"openapi_spec":        "https://api.example.com/openapi.json",
					"discover_operations": true,
				},
				Output: map[string]interface{}{
					"success": true,
					"operations": []map[string]interface{}{
						{
							"path":        "/users",
							"method":      "GET",
							"summary":     "List users",
							"operationId": "listUsers",
						},
						{
							"path":        "/users/{id}",
							"method":      "GET",
							"summary":     "Get user by ID",
							"operationId": "getUser",
						},
					},
					"spec_info": map[string]interface{}{
						"title":   "Example API",
						"version": "1.0.0",
					},
					"total_operations": 2,
				},
				Explanation: "Fetches and parses OpenAPI spec to show available endpoints",
			},
		}).
		WithConstraints([]string{
			"Only JSON request/response bodies are currently supported",
			"Authentication credentials should be kept secure",
			"Rate limiting is handled by returning 429 status codes",
			"Redirects are followed automatically up to 10 times",
			"OpenAPI specs must be in JSON or YAML format (OpenAPI 3.0/3.1)",
			"OpenAPI discovery requires valid spec URL accessible via HTTP/HTTPS",
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
		WithVersion("2.0.0").
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

	// Check if we're in discovery mode
	if discoverOps, ok := paramMap["discover_operations"].(bool); ok && discoverOps {
		specURL, ok := paramMap["openapi_spec"].(string)
		if !ok || specURL == "" {
			return nil, fmt.Errorf("openapi_spec URL is required when discover_operations is true")
		}

		// Perform operation discovery with caching
		parser := NewOpenAPIParser()
		cache := GetOpenAPICache()

		// Try cache first
		spec, discovery, found := cache.Get(specURL)
		if !found {
			// Fetch and cache
			var err error
			spec, err = parser.FetchSpec(specURL)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch OpenAPI spec: %w", err)
			}
			// Discovery is already cached by FetchSpec
			_, discovery, _ = cache.Get(specURL)
		}

		// Get enhanced operations
		enhancedOps := discovery.EnumerateOperations()

		// Get server URLs and security schemes
		serverURLs := []string{}
		for _, server := range spec.Servers {
			serverURLs = append(serverURLs, server.URL)
		}

		securitySchemes := spec.GetSecuritySchemes()

		// Convert enhanced operations to generic format for JSON serialization
		opsData := make([]map[string]interface{}, len(enhancedOps))
		for i, op := range enhancedOps {
			// Convert to map for JSON serialization
			opMap := map[string]interface{}{
				"path":        op.Path,
				"method":      op.Method,
				"operationId": op.OperationID,
				"summary":     op.Summary,
				"description": op.Description,
				"tags":        op.Tags,
				"deprecated":  op.Deprecated,
			}

			// Add parameter counts
			if len(op.PathParameters) > 0 {
				opMap["pathParameterCount"] = len(op.PathParameters)
			}
			if len(op.QueryParameters) > 0 {
				opMap["queryParameterCount"] = len(op.QueryParameters)
			}
			if len(op.HeaderParameters) > 0 {
				opMap["headerParameterCount"] = len(op.HeaderParameters)
			}
			if op.RequestBodyInfo != nil {
				opMap["hasRequestBody"] = true
			}

			opsData[i] = opMap
		}

		// Return discovery results with enhanced information
		return map[string]interface{}{
			"success":    true,
			"operations": opsData,
			"spec_info": map[string]interface{}{
				"title":       spec.Info.Title,
				"version":     spec.Info.Version,
				"description": spec.Info.Description,
			},
			"servers":          serverURLs,
			"security_schemes": securitySchemes,
			"total_operations": len(enhancedOps),
			"llm_guidance":     generateDiscoveryGuidance(spec, enhancedOps),
		}, nil
	}

	// Initialize variables for OpenAPI support
	var spec *OpenAPISpec
	var specURL string

	// Check if OpenAPI spec is provided
	if specURLParam, ok := paramMap["openapi_spec"].(string); ok && specURLParam != "" {
		specURL = specURLParam
		// Fetch and parse the spec
		parser := NewOpenAPIParser()
		var err error
		spec, err = parser.FetchSpec(specURL)
		if err != nil {
			// Log warning but continue - spec is optional for normal requests
			if ctx.Events != nil {
				ctx.Events.EmitCustom("openapi_parse_warning", map[string]interface{}{
					"error": err.Error(),
					"url":   specURL,
				})
			}
		}
	}

	// Validate required parameters
	baseURL, ok := paramMap["base_url"].(string)
	if !ok || baseURL == "" {
		// Try to get base URL from OpenAPI spec if available
		if spec != nil && len(spec.Servers) > 0 {
			baseURL = spec.Servers[0].URL
			// Handle relative server URLs
			if strings.HasPrefix(baseURL, "/") {
				// Extract host from spec URL
				specParsed, _ := url.Parse(specURL)
				if specParsed != nil {
					baseURL = fmt.Sprintf("%s://%s%s", specParsed.Scheme, specParsed.Host, baseURL)
				}
			}
			if ctx.Events != nil {
				ctx.Events.EmitCustom("auto_base_url", map[string]interface{}{
					"source": "openapi_spec",
					"url":    baseURL,
				})
			}
		} else {
			return nil, fmt.Errorf("base_url is required (no OpenAPI spec or servers found)")
		}
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
	authApplied := false
	if auth, ok := paramMap["auth"].(map[string]interface{}); ok {
		if err := applyAuthentication(req, auth); err != nil {
			return nil, err
		}
		authApplied = true
	}

	// Auto-apply authentication from OpenAPI spec if not already provided
	if !authApplied && spec != nil {
		// Try to auto-detect and apply authentication
		if authConfig := detectAuthFromSpec(spec, endpoint, method, ctx); authConfig != nil {
			if err := applyAuthentication(req, authConfig); err != nil {
				if ctx.Events != nil {
					ctx.Events.EmitCustom("auto_auth_failed", map[string]interface{}{
						"error": err.Error(),
					})
				}
			} else {
				if ctx.Events != nil {
					ctx.Events.EmitCustom("auto_auth_applied", map[string]interface{}{
						"type": authConfig["type"],
					})
				}
			}
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

	// If OpenAPI spec is provided, validate the request
	if specURL, ok := paramMap["openapi_spec"].(string); ok && specURL != "" {
		parser := NewOpenAPIParser()
		cache := GetOpenAPICache()

		// Try cache first
		spec, discovery, found := cache.Get(specURL)
		if !found {
			// Fetch and cache
			spec, err = parser.FetchSpec(specURL)
			if err != nil {
				// Continue without validation - spec is optional
				spec = nil
			} else {
				// Get cached discovery instance
				_, discovery, _ = cache.Get(specURL)
			}
		}

		if spec != nil && discovery != nil {
			// Find the operation by path and method using optimized index
			targetOp, found := discovery.FindOperation(method, endpoint)
			if found && targetOp != nil && targetOp.OperationID != "" {
				// Create validation options
				validationOpts := &ValidationOptions{
					SkipRequired:     false,
					SkipConstraints:  false,
					SkipTypeChecking: false,
					AllowCoercion:    true,
					StrictValidation: false,
				}

				// Validate the request
				var requestBody interface{}
				if bodyParam, exists := paramMap["body"]; exists {
					requestBody = bodyParam
				}
				report, err := discovery.ValidateRequest(targetOp.OperationID, paramMap, requestBody, validationOpts)
				if err == nil && !report.Valid {
					// Return validation errors
					errorMessages := []string{}
					if report.ParameterErrors != nil {
						for param, result := range report.ParameterErrors {
							for _, err := range result.Errors {
								errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", param, err))
							}
						}
					}
					if report.RequestBodyError != nil {
						errorMessages = append(errorMessages, report.RequestBodyError.Errors...)
					}

					return map[string]interface{}{
						"success":           false,
						"error_message":     "Request validation failed",
						"validation_errors": errorMessages,
						"error_guidance":    report.Guidance.Summary,
						"suggestions":       report.Suggestions,
					}, nil
				}
			}
		}
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

		// Add helpful guidance based on status code and OpenAPI spec
		if spec != nil {
			result["error_guidance"] = getOpenAPIErrorGuidance(resp.StatusCode, endpoint, method, spec)
		} else {
			result["error_guidance"] = getErrorGuidance(resp.StatusCode)
		}
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

// generateDiscoveryGuidance generates LLM-friendly guidance based on discovered operations
func generateDiscoveryGuidance(spec *OpenAPISpec, operations []EnhancedOperationInfo) string {
	guidance := fmt.Sprintf("API: %s v%s\n", spec.Info.Title, spec.Info.Version)

	if spec.Info.Description != "" {
		guidance += fmt.Sprintf("Description: %s\n\n", spec.Info.Description)
	}

	// Server information
	if len(spec.Servers) > 0 {
		guidance += "Available servers:\n"
		for _, server := range spec.Servers {
			guidance += fmt.Sprintf("- %s", server.URL)
			if server.Description != "" {
				guidance += fmt.Sprintf(" (%s)", server.Description)
			}
			guidance += "\n"
		}
		guidance += "\n"
	}

	// Security schemes
	schemes := spec.GetSecuritySchemes()
	if len(schemes) > 0 {
		guidance += "Authentication methods:\n"
		for name, scheme := range schemes {
			switch scheme.Type {
			case "apiKey":
				guidance += fmt.Sprintf("- %s: API key in %s '%s'\n", name, scheme.In, scheme.Name)
			case "http":
				if scheme.Scheme == "bearer" {
					guidance += fmt.Sprintf("- %s: Bearer token authentication\n", name)
				} else if scheme.Scheme == "basic" {
					guidance += fmt.Sprintf("- %s: Basic authentication (username/password)\n", name)
				} else {
					guidance += fmt.Sprintf("- %s: HTTP %s authentication\n", name, scheme.Scheme)
				}
			case "oauth2":
				guidance += fmt.Sprintf("- %s: OAuth2 authentication\n", name)
			case "openIdConnect":
				guidance += fmt.Sprintf("- %s: OpenID Connect authentication\n", name)
			}
		}
		guidance += "\n"
	}

	// Operation summary
	guidance += fmt.Sprintf("Total operations: %d\n\n", len(operations))

	// Group operations by tag
	taggedOps := make(map[string][]EnhancedOperationInfo)
	untaggedOps := []EnhancedOperationInfo{}

	for _, op := range operations {
		if len(op.Tags) > 0 {
			for _, tag := range op.Tags {
				taggedOps[tag] = append(taggedOps[tag], op)
			}
		} else {
			untaggedOps = append(untaggedOps, op)
		}
	}

	// Display operations by tag
	if len(taggedOps) > 0 {
		guidance += "Operations by category:\n"
		for tag, ops := range taggedOps {
			guidance += fmt.Sprintf("\n%s:\n", tag)
			for _, op := range ops {
				guidance += formatOperationSummary(op)
			}
		}
	}

	// Display untagged operations
	if len(untaggedOps) > 0 {
		guidance += "\nOther operations:\n"
		for _, op := range untaggedOps {
			guidance += formatOperationSummary(op)
		}
	}

	guidance += "\nTo use an operation, provide the endpoint path and method. The tool will guide you on required parameters."

	return guidance
}

// formatOperationSummary formats a single operation for display
func formatOperationSummary(op EnhancedOperationInfo) string {
	summary := fmt.Sprintf("- %s %s", op.Method, op.Path)
	if op.Summary != "" {
		summary += fmt.Sprintf(" - %s", op.Summary)
	}
	if op.OperationID != "" {
		summary += fmt.Sprintf(" (ID: %s)", op.OperationID)
	}
	if op.Deprecated {
		summary += " [DEPRECATED]"
	}
	summary += "\n"

	// Add parameter information
	paramCount := len(op.PathParameters) + len(op.QueryParameters) + len(op.HeaderParameters)
	if paramCount > 0 || op.RequestBodyInfo != nil {
		summary += "  "
		parts := []string{}
		if len(op.PathParameters) > 0 {
			parts = append(parts, fmt.Sprintf("%d path params", len(op.PathParameters)))
		}
		if len(op.QueryParameters) > 0 {
			parts = append(parts, fmt.Sprintf("%d query params", len(op.QueryParameters)))
		}
		if len(op.HeaderParameters) > 0 {
			parts = append(parts, fmt.Sprintf("%d header params", len(op.HeaderParameters)))
		}
		if op.RequestBodyInfo != nil {
			parts = append(parts, "request body")
		}
		summary += fmt.Sprintf("Requires: %s\n", strings.Join(parts, ", "))
	}

	return summary
}

// detectAuthFromSpec attempts to detect authentication configuration from OpenAPI spec
func detectAuthFromSpec(spec *OpenAPISpec, endpoint, method string, ctx *domain.ToolContext) map[string]interface{} {
	// Find the operation
	var operation *Operation
	for path, pathItem := range spec.Paths {
		if path == endpoint {
			switch strings.ToLower(method) {
			case "get":
				operation = pathItem.Get
			case "post":
				operation = pathItem.Post
			case "put":
				operation = pathItem.Put
			case "delete":
				operation = pathItem.Delete
			case "patch":
				operation = pathItem.Patch
			case "head":
				operation = pathItem.Head
			case "options":
				operation = pathItem.Options
			}
			break
		}
	}

	if operation == nil {
		return nil
	}

	// Get security requirements for this operation (or global if not specified)
	securityReqs := operation.Security
	if len(securityReqs) == 0 && len(spec.Security) > 0 {
		securityReqs = spec.Security
	}

	if len(securityReqs) == 0 {
		return nil
	}

	// Get security schemes
	schemes := spec.GetSecuritySchemes()
	if len(schemes) == 0 {
		return nil
	}

	// Try to find credentials in agent state for the first matching security requirement
	for _, req := range securityReqs {
		for schemeName := range req {
			if scheme, ok := schemes[schemeName]; ok {
				// Try to get credentials from state
				authConfig := getAuthConfigFromState(ctx, schemeName, scheme)
				if authConfig != nil {
					return authConfig
				}
			}
		}
	}

	// If no credentials found, return guidance about required authentication
	if ctx.Events != nil {
		ctx.Events.EmitCustom("auth_required", map[string]interface{}{
			"schemes": getAuthSchemeNames(securityReqs, schemes),
			"message": "Authentication required but no credentials found in state",
		})
	}

	return nil
}

// getAuthConfigFromState attempts to retrieve authentication configuration from agent state
func getAuthConfigFromState(ctx *domain.ToolContext, schemeName string, scheme SecurityScheme) map[string]interface{} {
	state := ctx.State

	switch scheme.Type {
	case "apiKey":
		// Look for API key in state with various common names
		keyNames := []string{
			fmt.Sprintf("%s_api_key", schemeName),
			fmt.Sprintf("%s_key", schemeName),
			"api_key",
			"apiKey",
			scheme.Name, // The actual header/query param name
		}

		for _, keyName := range keyNames {
			if value, exists := state.Get(keyName); exists {
				if apiKey, ok := value.(string); ok && apiKey != "" {
					return map[string]interface{}{
						"type":         "api_key",
						"api_key":      apiKey,
						"key_location": scheme.In,
						"key_name":     scheme.Name,
					}
				}
			}
		}

	case "http":
		if scheme.Scheme == "bearer" {
			// Look for bearer token
			tokenNames := []string{
				fmt.Sprintf("%s_token", schemeName),
				fmt.Sprintf("%s_bearer", schemeName),
				"bearer_token",
				"access_token",
				"token",
			}

			for _, tokenName := range tokenNames {
				if value, exists := state.Get(tokenName); exists {
					if token, ok := value.(string); ok && token != "" {
						return map[string]interface{}{
							"type":  "bearer",
							"token": token,
						}
					}
				}
			}
		} else if scheme.Scheme == "basic" {
			// Look for basic auth credentials
			usernameKeys := []string{
				fmt.Sprintf("%s_username", schemeName),
				"api_username",
				"username",
			}
			passwordKeys := []string{
				fmt.Sprintf("%s_password", schemeName),
				"api_password",
				"password",
			}

			var username, password string
			for _, key := range usernameKeys {
				if value, exists := state.Get(key); exists {
					if u, ok := value.(string); ok && u != "" {
						username = u
						break
					}
				}
			}

			for _, key := range passwordKeys {
				if value, exists := state.Get(key); exists {
					if p, ok := value.(string); ok && p != "" {
						password = p
						break
					}
				}
			}

			if username != "" && password != "" {
				return map[string]interface{}{
					"type":     "basic",
					"username": username,
					"password": password,
				}
			}
		}
	}

	return nil
}

// getAuthSchemeNames extracts human-readable authentication scheme names
func getAuthSchemeNames(securityReqs []SecurityRequirement, schemes map[string]SecurityScheme) []string {
	names := []string{}
	seen := make(map[string]bool)

	for _, req := range securityReqs {
		for schemeName := range req {
			if !seen[schemeName] {
				if scheme, ok := schemes[schemeName]; ok {
					var desc string
					switch scheme.Type {
					case "apiKey":
						desc = fmt.Sprintf("%s (API key in %s '%s')", schemeName, scheme.In, scheme.Name)
					case "http":
						if scheme.Scheme == "bearer" {
							desc = fmt.Sprintf("%s (Bearer token)", schemeName)
						} else if scheme.Scheme == "basic" {
							desc = fmt.Sprintf("%s (Basic auth)", schemeName)
						} else {
							desc = fmt.Sprintf("%s (HTTP %s)", schemeName, scheme.Scheme)
						}
					case "oauth2":
						desc = fmt.Sprintf("%s (OAuth2)", schemeName)
					case "openIdConnect":
						desc = fmt.Sprintf("%s (OpenID Connect)", schemeName)
					default:
						desc = schemeName
					}
					names = append(names, desc)
					seen[schemeName] = true
				}
			}
		}
	}

	return names
}

// getOpenAPIErrorGuidance returns enhanced error guidance based on OpenAPI spec
func getOpenAPIErrorGuidance(statusCode int, endpoint, method string, spec *OpenAPISpec) string {
	// Start with generic guidance
	guidance := getErrorGuidance(statusCode)

	// Find the operation in the spec
	var operation *Operation
	for path, pathItem := range spec.Paths {
		if path == endpoint {
			switch strings.ToLower(method) {
			case "get":
				operation = pathItem.Get
			case "post":
				operation = pathItem.Post
			case "put":
				operation = pathItem.Put
			case "delete":
				operation = pathItem.Delete
			case "patch":
				operation = pathItem.Patch
			case "head":
				operation = pathItem.Head
			case "options":
				operation = pathItem.Options
			}
			break
		}
	}

	if operation == nil {
		return guidance + "\n\nNote: This endpoint is not documented in the OpenAPI spec."
	}

	// Add OpenAPI-specific guidance based on status code
	switch statusCode {
	case 400:
		// Provide parameter-specific guidance
		var paramInfo []string

		// Check path parameters
		for _, param := range operation.Parameters {
			if param.Required && param.In == "path" {
				paramInfo = append(paramInfo, fmt.Sprintf("- %s (path): %s", param.Name, param.Description))
			}
		}

		// Check query parameters
		for _, param := range operation.Parameters {
			if param.Required && param.In == "query" {
				paramInfo = append(paramInfo, fmt.Sprintf("- %s (query): %s", param.Name, param.Description))
			}
		}

		// Check request body
		if operation.RequestBody != nil && operation.RequestBody.Required {
			paramInfo = append(paramInfo, "- Request body is required")
			if operation.RequestBody.Description != "" {
				paramInfo = append(paramInfo, fmt.Sprintf("  Description: %s", operation.RequestBody.Description))
			}
		}

		if len(paramInfo) > 0 {
			guidance += "\n\nRequired parameters for this endpoint:\n" + strings.Join(paramInfo, "\n")
		}

	case 401:
		// Provide authentication-specific guidance
		securityReqs := operation.Security
		if len(securityReqs) == 0 && len(spec.Security) > 0 {
			securityReqs = spec.Security
		}

		if len(securityReqs) > 0 {
			schemes := spec.GetSecuritySchemes()
			authMethods := getAuthSchemeNames(securityReqs, schemes)
			if len(authMethods) > 0 {
				guidance += "\n\nThis endpoint requires one of these authentication methods:\n"
				for _, method := range authMethods {
					guidance += fmt.Sprintf("- %s\n", method)
				}
				guidance += "\nProvide credentials using the 'auth' parameter or store them in agent state."
			}
		}

	case 403:
		// Check if operation has specific security requirements
		if operation.Description != "" {
			guidance += fmt.Sprintf("\n\nEndpoint description: %s", operation.Description)
		}
		if len(operation.Security) > 0 {
			guidance += "\n\nThis endpoint has specific permission requirements. Check your access level."
		}

	case 404:
		// Provide path parameter guidance
		pathParams := []string{}
		for _, param := range operation.Parameters {
			if param.In == "path" {
				desc := param.Name
				if param.Description != "" {
					desc += fmt.Sprintf(" (%s)", param.Description)
				}
				pathParams = append(pathParams, desc)
			}
		}

		if len(pathParams) > 0 {
			guidance += "\n\nThis endpoint requires these path parameters:\n"
			for _, param := range pathParams {
				guidance += fmt.Sprintf("- %s\n", param)
			}
			guidance += "\nEnsure all path parameters are correctly substituted in the URL."
		}

	case 405:
		// List allowed methods for this path
		allowedMethods := []string{}
		for path, pathItem := range spec.Paths {
			if path == endpoint {
				if pathItem.Get != nil {
					allowedMethods = append(allowedMethods, "GET")
				}
				if pathItem.Post != nil {
					allowedMethods = append(allowedMethods, "POST")
				}
				if pathItem.Put != nil {
					allowedMethods = append(allowedMethods, "PUT")
				}
				if pathItem.Delete != nil {
					allowedMethods = append(allowedMethods, "DELETE")
				}
				if pathItem.Patch != nil {
					allowedMethods = append(allowedMethods, "PATCH")
				}
				if pathItem.Head != nil {
					allowedMethods = append(allowedMethods, "HEAD")
				}
				if pathItem.Options != nil {
					allowedMethods = append(allowedMethods, "OPTIONS")
				}
				break
			}
		}

		if len(allowedMethods) > 0 {
			guidance += fmt.Sprintf("\n\nAllowed methods for %s: %s", endpoint, strings.Join(allowedMethods, ", "))
		}
	}

	// Add operation-specific information if available
	if operation.Summary != "" {
		guidance += fmt.Sprintf("\n\nOperation: %s", operation.Summary)
	}

	// Check if there's a specific response description for this status code
	if operation.Responses != nil {
		if response, ok := operation.Responses[fmt.Sprintf("%d", statusCode)]; ok {
			if response.Description != "" {
				guidance += fmt.Sprintf("\n\nAPI documentation for %d response: %s", statusCode, response.Description)
			}
		}
	}

	return guidance
}
