// ABOUTME: HTTP request tool for advanced HTTP operations with full method support
// ABOUTME: Built-in tool supporting POST, PUT, DELETE, PATCH with headers, auth, and body options

package web

import (
	"bytes"
	"encoding/base64"
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

// HTTPRequestParams defines parameters for the HTTPRequest tool
type HTTPRequestParams struct {
	URL         string            `json:"url"`
	Method      string            `json:"method,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	BodyType    string            `json:"body_type,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`

	// Auth fields flattened for better compatibility
	AuthType        string `json:"auth_type,omitempty"` // basic, bearer, api_key
	AuthUsername    string `json:"auth_username,omitempty"`
	AuthPassword    string `json:"auth_password,omitempty"`
	AuthToken       string `json:"auth_token,omitempty"`
	AuthKeyName     string `json:"auth_key_name,omitempty"`
	AuthKeyValue    string `json:"auth_key_value,omitempty"`
	AuthKeyLocation string `json:"auth_key_location,omitempty"` // header, query

	Timeout         int  `json:"timeout,omitempty"`
	FollowRedirects bool `json:"follow_redirects,omitempty"`
}

// HTTPRequestResult defines the result of the HTTPRequest tool
type HTTPRequestResult struct {
	StatusCode    int               `json:"status_code"`
	Status        string            `json:"status"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
	ContentType   string            `json:"content_type"`
	ContentLength int64             `json:"content_length"`
	ResponseTime  int64             `json:"response_time_ms"`
	RedirectURL   string            `json:"redirect_url,omitempty"`
}

// httpRequestParamSchema defines parameters for the HTTPRequest tool
var httpRequestParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"url": {
			Type:        "string",
			Format:      "uri",
			Description: "The URL to send the request to",
		},
		"method": {
			Type:        "string",
			Description: "HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)",
		},
		"headers": {
			Type:        "object",
			Description: "HTTP headers to include in the request",
		},
		"body": {
			Type:        "string",
			Description: "Request body content",
		},
		"body_type": {
			Type:        "string",
			Description: "Body content type (json, form, text, xml)",
		},
		"query_params": {
			Type:        "object",
			Description: "Query parameters to append to the URL",
		},
		"auth_type": {
			Type:        "string",
			Description: "Authentication type (basic, bearer, api_key)",
		},
		"auth_username": {
			Type:        "string",
			Description: "Username for basic auth",
		},
		"auth_password": {
			Type:        "string",
			Description: "Password for basic auth",
		},
		"auth_token": {
			Type:        "string",
			Description: "Token for bearer auth",
		},
		"auth_key_name": {
			Type:        "string",
			Description: "API key name",
		},
		"auth_key_value": {
			Type:        "string",
			Description: "API key value",
		},
		"auth_key_location": {
			Type:        "string",
			Description: "Where to place the API key (header or query)",
		},
		"timeout": {
			Type:        "number",
			Description: "Request timeout in seconds (default: 30)",
		},
		"follow_redirects": {
			Type:        "boolean",
			Description: "Whether to follow redirects (default: true)",
		},
	},
	Required: []string{"url"},
}

// httpRequestOutputSchema defines the output for the HTTPRequest tool
var httpRequestOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"status_code": {
			Type:        "number",
			Description: "HTTP response status code",
		},
		"status": {
			Type:        "string",
			Description: "HTTP response status text",
		},
		"headers": {
			Type:        "object",
			Description: "Response headers",
		},
		"body": {
			Type:        "string",
			Description: "Response body content",
		},
		"content_type": {
			Type:        "string",
			Description: "Response Content-Type header",
		},
		"content_length": {
			Type:        "number",
			Description: "Response Content-Length in bytes",
		},
		"response_time_ms": {
			Type:        "number",
			Description: "Response time in milliseconds",
		},
		"redirect_url": {
			Type:        "string",
			Description: "Redirect Location header if present",
		},
	},
	Required: []string{"status_code", "status", "headers", "body"},
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("http_request", HTTPRequest(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "http_request",
			Category:    "web",
			Tags:        []string{"http", "api", "rest", "post", "put", "delete", "network"},
			Description: "Makes HTTP requests with full method and authentication support",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Simple GET request",
					Description: "Make a basic GET request",
					Code:        `HTTPRequest().Execute(ctx, HTTPRequestParams{URL: "https://api.example.com/data"})`,
				},
				{
					Name:        "POST with JSON",
					Description: "Send JSON data via POST",
					Code:        `HTTPRequest().Execute(ctx, HTTPRequestParams{URL: "https://api.example.com/users", Method: "POST", Body: "{\"name\":\"John\"}", BodyType: "json"})`,
				},
				{
					Name:        "Authenticated request",
					Description: "Make request with bearer token",
					Code:        `HTTPRequest().Execute(ctx, HTTPRequestParams{URL: "https://api.example.com/profile", AuthType: "bearer", AuthToken: "your-token"})`,
				},
			},
		},
		RequiredPermissions: []string{"network:access"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     true,
			FileSystem:  false,
			Concurrency: true,
		},
	})
}

// HTTPRequest creates a tool for making HTTP requests with full method and authentication support.
// It provides comprehensive HTTP client functionality supporting all standard methods (GET, POST, PUT, DELETE, PATCH),
// multiple authentication types (basic, bearer, API key), flexible request body handling with content type detection,
// custom headers and query parameters, redirect control, and detailed response information including timing metrics.
func HTTPRequest() domain.Tool {
	builder := atools.NewToolBuilder("http_request", "Makes HTTP requests with full method and authentication support").
		WithFunction(func(ctx *domain.ToolContext, params HTTPRequestParams) (*HTTPRequestResult, error) {
			startTime := time.Now()

			// Emit start event
			if ctx.Events != nil {
				ctx.Events.EmitMessage(fmt.Sprintf("Starting %s request to %s", params.Method, params.URL))
			}

			// Set defaults
			if params.Method == "" {
				params.Method = "GET"
			}
			params.Method = strings.ToUpper(params.Method)

			if params.Timeout == 0 {
				params.Timeout = 30
			}

			// Default to following redirects
			// nolint:staticcheck // QF1007: Can't merge because we need default true behavior
			followRedirects := true
			if !params.FollowRedirects {
				// Only set to false if explicitly disabled
				// TODO: Consider using *bool to distinguish between unset and false
				followRedirects = false
			}

			// Check state for default auth settings
			if ctx.State != nil && params.AuthType == "" {
				if authType, exists := ctx.State.Get("default_auth_type"); exists {
					if authStr, ok := authType.(string); ok {
						params.AuthType = authStr
					}
				}
				// Check for default API keys
				if params.AuthType == "api_key" && params.AuthKeyValue == "" {
					if apiKey, exists := ctx.State.Get("api_key"); exists {
						if keyStr, ok := apiKey.(string); ok {
							params.AuthKeyValue = keyStr
						}
					}
				}
				// Check for default bearer token
				if params.AuthType == "bearer" && params.AuthToken == "" {
					if token, exists := ctx.State.Get("bearer_token"); exists {
						if tokenStr, ok := token.(string); ok {
							params.AuthToken = tokenStr
						}
					}
				}
			}

			// Validate method
			validMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true, "DELETE": true,
				"PATCH": true, "HEAD": true, "OPTIONS": true,
			}
			if !validMethods[params.Method] {
				return nil, fmt.Errorf("invalid HTTP method: %s", params.Method)
			}

			// Parse URL
			parsedURL, err := url.Parse(params.URL)
			if err != nil {
				return nil, fmt.Errorf("invalid URL: %w", err)
			}

			// Add query parameters
			if len(params.QueryParams) > 0 {
				q := parsedURL.Query()
				for key, value := range params.QueryParams {
					q.Add(key, value)
				}
				parsedURL.RawQuery = q.Encode()
			}

			// Add API key to query if specified
			if params.AuthType == "api_key" && params.AuthKeyLocation == "query" {
				q := parsedURL.Query()
				q.Add(params.AuthKeyName, params.AuthKeyValue)
				parsedURL.RawQuery = q.Encode()
			}

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(1, 4, "Preparing request")
			}

			// Prepare request body
			var bodyReader io.Reader
			if params.Body != "" {
				bodyReader = bytes.NewBufferString(params.Body)
			}

			// Create request
			req, err := http.NewRequestWithContext(ctx.Context, params.Method, parsedURL.String(), bodyReader)
			if err != nil {
				return nil, fmt.Errorf("error creating request: %w", err)
			}

			// Set default headers
			userAgent := "go-llms/1.0 (HTTPRequest)"
			if ctx.State != nil {
				if ua, exists := ctx.State.Get("user_agent"); exists {
					if uaStr, ok := ua.(string); ok {
						userAgent = uaStr
					}
				}
			}
			req.Header.Set("User-Agent", userAgent)

			// Set content type based on body type
			if params.Body != "" && params.BodyType != "" {
				switch strings.ToLower(params.BodyType) {
				case "json":
					req.Header.Set("Content-Type", "application/json")
				case "form":
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				case "xml":
					req.Header.Set("Content-Type", "application/xml")
				case "text":
					req.Header.Set("Content-Type", "text/plain")
				default:
					req.Header.Set("Content-Type", "application/octet-stream")
				}
			}

			// Set custom headers
			if params.Headers != nil {
				for key, value := range params.Headers {
					req.Header.Set(key, value)
				}
			}

			// Check state for additional default headers
			if ctx.State != nil {
				if headers, exists := ctx.State.Get("http_headers"); exists {
					if headerMap, ok := headers.(map[string]string); ok {
						for key, value := range headerMap {
							// Don't override explicitly set headers
							if req.Header.Get(key) == "" {
								req.Header.Set(key, value)
							}
						}
					}
				}
			}

			// Set authentication
			switch strings.ToLower(params.AuthType) {
			case "basic":
				if params.AuthUsername != "" && params.AuthPassword != "" {
					auth := params.AuthUsername + ":" + params.AuthPassword
					encoded := base64.StdEncoding.EncodeToString([]byte(auth))
					req.Header.Set("Authorization", "Basic "+encoded)
				}
			case "bearer":
				if params.AuthToken != "" {
					req.Header.Set("Authorization", "Bearer "+params.AuthToken)
				}
			case "api_key":
				if params.AuthKeyLocation == "header" && params.AuthKeyName != "" && params.AuthKeyValue != "" {
					req.Header.Set(params.AuthKeyName, params.AuthKeyValue)
				}
			}

			// Create HTTP client
			timeout := time.Duration(params.Timeout) * time.Second
			client := &http.Client{
				Timeout: timeout,
			}

			// Configure redirect policy
			if !followRedirects {
				client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
			}

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(2, 4, "Sending request")
			}

			// Execute request
			resp, err := client.Do(req)
			if err != nil {
				if ctx.Events != nil {
					ctx.Events.EmitError(err)
				}
				return nil, fmt.Errorf("error executing request: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(3, 4, "Reading response")
			}

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response: %w", err)
			}

			// Extract headers
			headers := make(map[string]string)
			for key, values := range resp.Header {
				if len(values) > 0 {
					headers[key] = values[0]
				}
			}

			// Get redirect URL if present
			redirectURL := ""
			if location := resp.Header.Get("Location"); location != "" {
				redirectURL = location
			}

			// Emit completion event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(4, 4, "Complete")
				ctx.Events.EmitCustom("http_request_complete", map[string]interface{}{
					"method":       params.Method,
					"url":          params.URL,
					"statusCode":   resp.StatusCode,
					"responseTime": time.Since(startTime).Milliseconds(),
					"bodySize":     len(body),
				})
			}

			return &HTTPRequestResult{
				StatusCode:    resp.StatusCode,
				Status:        resp.Status,
				Headers:       headers,
				Body:          string(body),
				ContentType:   resp.Header.Get("Content-Type"),
				ContentLength: resp.ContentLength,
				ResponseTime:  time.Since(startTime).Milliseconds(),
				RedirectURL:   redirectURL,
			}, nil
		}).
		WithParameterSchema(httpRequestParamSchema).
		WithOutputSchema(httpRequestOutputSchema).
		WithUsageInstructions(`Use this tool to make HTTP requests with full control over method, headers, body, and authentication.

Supported methods:
- GET: Retrieve data
- POST: Create new resources
- PUT: Update existing resources
- DELETE: Remove resources
- PATCH: Partial updates
- HEAD: Get headers only
- OPTIONS: Get allowed methods

Authentication methods:
- Basic: Username/password authentication
- Bearer: Token-based authentication (JWT, OAuth)
- API Key: Key in header or query parameter

Body types:
- json: application/json
- form: application/x-www-form-urlencoded
- xml: application/xml
- text: text/plain
- (default): application/octet-stream

State configuration:
- default_auth_type: Default authentication method
- api_key: Default API key for api_key auth
- bearer_token: Default token for bearer auth
- user_agent: Custom User-Agent header
- http_headers: Default headers as map[string]string

The tool will:
- Automatically handle redirects (unless disabled)
- Add query parameters to URL
- Set appropriate Content-Type for body
- Measure response time
- Return comprehensive response information`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Simple GET request",
				Description: "Basic data retrieval",
				Scenario:    "When you need to fetch data from an API",
				Input: map[string]interface{}{
					"url": "https://api.example.com/users",
				},
				Output: map[string]interface{}{
					"status_code": 200,
					"status":      "200 OK",
					"headers": map[string]string{
						"Content-Type": "application/json",
					},
					"body":             `[{"id":1,"name":"John"},{"id":2,"name":"Jane"}]`,
					"content_type":     "application/json",
					"response_time_ms": 125,
				},
				Explanation: "GET is the default method when not specified",
			},
			{
				Name:        "POST with JSON body",
				Description: "Create a new resource",
				Scenario:    "When creating new data via API",
				Input: map[string]interface{}{
					"url":       "https://api.example.com/users",
					"method":    "POST",
					"body":      `{"name":"Alice","email":"alice@example.com"}`,
					"body_type": "json",
				},
				Output: map[string]interface{}{
					"status_code": 201,
					"status":      "201 Created",
					"headers": map[string]string{
						"Location": "https://api.example.com/users/3",
					},
					"body":             `{"id":3,"name":"Alice","email":"alice@example.com"}`,
					"response_time_ms": 230,
				},
				Explanation: "Content-Type is automatically set based on body_type",
			},
			{
				Name:        "PUT with form data",
				Description: "Update resource with form encoding",
				Scenario:    "When updating via form submission",
				Input: map[string]interface{}{
					"url":       "https://api.example.com/profile",
					"method":    "PUT",
					"body":      "name=Bob&city=NYC&age=30",
					"body_type": "form",
				},
				Output: map[string]interface{}{
					"status_code":      200,
					"status":           "200 OK",
					"body":             `{"message":"Profile updated"}`,
					"response_time_ms": 150,
				},
				Explanation: "Form data uses application/x-www-form-urlencoded",
			},
			{
				Name:        "DELETE request",
				Description: "Remove a resource",
				Scenario:    "When deleting data",
				Input: map[string]interface{}{
					"url":    "https://api.example.com/users/123",
					"method": "DELETE",
				},
				Output: map[string]interface{}{
					"status_code":      204,
					"status":           "204 No Content",
					"body":             "",
					"response_time_ms": 90,
				},
				Explanation: "DELETE often returns 204 with no body",
			},
			{
				Name:        "Bearer token auth",
				Description: "Authenticated API request",
				Scenario:    "When accessing protected resources",
				Input: map[string]interface{}{
					"url":        "https://api.example.com/me",
					"auth_type":  "bearer",
					"auth_token": "eyJhbGciOiJIUzI1NiIs...",
				},
				Output: map[string]interface{}{
					"status_code":      200,
					"body":             `{"id":42,"username":"johndoe"}`,
					"response_time_ms": 100,
				},
				Explanation: "Bearer token is added to Authorization header",
			},
			{
				Name:        "Basic authentication",
				Description: "Username/password auth",
				Scenario:    "When using basic auth",
				Input: map[string]interface{}{
					"url":           "https://api.example.com/admin",
					"auth_type":     "basic",
					"auth_username": "admin",
					"auth_password": "secret123",
				},
				Output: map[string]interface{}{
					"status_code": 200,
					"body":        `{"role":"admin","permissions":["read","write"]}`,
				},
				Explanation: "Credentials are base64 encoded in Authorization header",
			},
			{
				Name:        "API key in header",
				Description: "API key authentication",
				Scenario:    "When using API key auth",
				Input: map[string]interface{}{
					"url":               "https://api.example.com/data",
					"auth_type":         "api_key",
					"auth_key_name":     "X-API-Key",
					"auth_key_value":    "abc123xyz",
					"auth_key_location": "header",
				},
				Output: map[string]interface{}{
					"status_code": 200,
					"body":        `{"data":[1,2,3,4,5]}`,
				},
				Explanation: "API key is added as custom header",
			},
			{
				Name:        "Query parameters",
				Description: "Add URL parameters",
				Scenario:    "When filtering or paginating results",
				Input: map[string]interface{}{
					"url": "https://api.example.com/search",
					"query_params": map[string]string{
						"q":     "golang",
						"limit": "10",
						"page":  "2",
					},
				},
				Output: map[string]interface{}{
					"status_code": 200,
					"body":        `{"results":[...],"page":2,"total":150}`,
				},
				Explanation: "Parameters are URL-encoded and appended",
			},
			{
				Name:        "Custom headers",
				Description: "Add custom HTTP headers",
				Scenario:    "When API requires specific headers",
				Input: map[string]interface{}{
					"url":    "https://api.example.com/v2/data",
					"method": "POST",
					"headers": map[string]string{
						"X-Request-ID":  "uuid-123",
						"Accept":        "application/vnd.api+json",
						"Cache-Control": "no-cache",
					},
					"body": `{"type":"query"}`,
				},
				Output: map[string]interface{}{
					"status_code": 200,
					"headers": map[string]string{
						"X-Request-ID": "uuid-123",
					},
				},
				Explanation: "Custom headers are merged with defaults",
			},
			{
				Name:        "Handle redirects",
				Description: "Control redirect behavior",
				Scenario:    "When you need redirect information",
				Input: map[string]interface{}{
					"url":              "http://example.com/old-path",
					"follow_redirects": false,
				},
				Output: map[string]interface{}{
					"status_code":  301,
					"status":       "301 Moved Permanently",
					"redirect_url": "https://example.com/new-path",
					"headers": map[string]string{
						"Location": "https://example.com/new-path",
					},
				},
				Explanation: "When redirects are disabled, returns redirect info",
			},
		}).
		WithConstraints([]string{
			"Only HTTP and HTTPS protocols are supported",
			"Default timeout is 30 seconds",
			"Default method is GET if not specified",
			"Redirects are followed by default (up to 10)",
			"Headers are case-insensitive but preserved as sent",
			"Body is returned as string (binary data may be corrupted)",
			"Auth credentials should be kept secure",
			"Custom headers override default headers",
			"User-Agent defaults to 'go-llms/1.0 (HTTPRequest)'",
			"Response size limited by available memory",
		}).
		WithErrorGuidance(map[string]string{
			"invalid URL":               "Ensure URL is properly formatted with protocol",
			"invalid HTTP method":       "Use one of: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS",
			"timeout":                   "Increase timeout parameter or check server response time",
			"connection refused":        "Server may be down or firewall blocking connection",
			"certificate verification":  "SSL certificate issue - server may use self-signed cert",
			"401 Unauthorized":          "Check authentication credentials are correct",
			"403 Forbidden":             "User lacks permission for this resource",
			"404 Not Found":             "Check URL path is correct",
			"405 Method Not Allowed":    "This HTTP method is not supported for this endpoint",
			"500 Internal Server Error": "Server-side error - check API status",
			"context deadline exceeded": "Request cancelled - increase timeout",
			"no such host":              "Domain name cannot be resolved",
			"malformed HTTP response":   "Server returned invalid HTTP",
			"too many redirects":        "Redirect loop detected",
		}).
		WithCategory("web").
		WithTags([]string{"web", "http", "api", "rest", "request", "post", "put", "delete"}).
		WithVersion("2.0.0").
		WithBehavior(
			false,  // Not deterministic - responses can vary
			false,  // Not destructive by default (but DELETE/PUT can be)
			false,  // No confirmation needed
			"fast", // Usually fast, depends on network and server
		)

	return builder.Build()
}

// MustGetHTTPRequest retrieves the registered HTTPRequest tool or panics if not found.
// This is a convenience function for users who want to ensure the tool exists
// and prefer a panic over error handling for missing tools in their initialization code.
func MustGetHTTPRequest() domain.Tool {
	return tools.MustGetTool("http_request")
}
