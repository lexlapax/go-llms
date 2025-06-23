// ABOUTME: HTTP fetch tool for retrieving web content with timeout and context support
// ABOUTME: Built-in tool that provides web content fetching capabilities for agents

package web

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/util/auth"
)

// WebFetchParams defines parameters for the WebFetch tool
type WebFetchParams struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout,omitempty"` // Timeout in seconds, default 30

	// Authentication parameters (optional)
	AuthType        string `json:"auth_type,omitempty"`         // "bearer", "basic", "api_key", "oauth2", "custom"
	AuthToken       string `json:"auth_token,omitempty"`        // Bearer token or API token
	AuthUsername    string `json:"auth_username,omitempty"`     // Username for basic auth
	AuthPassword    string `json:"auth_password,omitempty"`     // Password for basic auth
	AuthAPIKey      string `json:"auth_api_key,omitempty"`      // API key value
	AuthKeyName     string `json:"auth_key_name,omitempty"`     // API key name (default: X-API-Key)
	AuthKeyLocation string `json:"auth_key_location,omitempty"` // "header", "query", "cookie" (default: header)
	AuthHeaderName  string `json:"auth_header_name,omitempty"`  // Custom header name for custom auth
	AuthHeaderValue string `json:"auth_header_value,omitempty"` // Custom header value for custom auth
	AuthPrefix      string `json:"auth_prefix,omitempty"`       // Optional prefix for custom auth (e.g., "Token")
}

// WebFetchResult defines the result of the WebFetch tool
type WebFetchResult struct {
	Content    string            `json:"content"`
	Status     int               `json:"status"`
	StatusText string            `json:"status_text"`
	Headers    map[string]string `json:"headers,omitempty"`
}

// webFetchParamSchema defines parameters for the WebFetch tool
var webFetchParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"url": {
			Type:        "string",
			Format:      "uri",
			Description: "The URL to fetch content from",
		},
		"timeout": {
			Type:        "number",
			Description: "Request timeout in seconds (default: 30)",
		},
		"auth_type": {
			Type:        "string",
			Description: "Authentication type: 'bearer', 'basic', 'api_key', 'oauth2', 'custom'",
		},
		"auth_token": {
			Type:        "string",
			Description: "Bearer token or general authentication token",
		},
		"auth_username": {
			Type:        "string",
			Description: "Username for basic authentication",
		},
		"auth_password": {
			Type:        "string",
			Description: "Password for basic authentication",
		},
		"auth_api_key": {
			Type:        "string",
			Description: "API key value",
		},
		"auth_key_name": {
			Type:        "string",
			Description: "API key parameter name (default: X-API-Key)",
		},
		"auth_key_location": {
			Type:        "string",
			Description: "Where to place API key: 'header', 'query', or 'cookie' (default: header)",
		},
		"auth_header_name": {
			Type:        "string",
			Description: "Custom header name for custom authentication",
		},
		"auth_header_value": {
			Type:        "string",
			Description: "Custom header value for custom authentication",
		},
		"auth_prefix": {
			Type:        "string",
			Description: "Optional prefix for custom auth header value (e.g., 'Token')",
		},
	},
	Required: []string{"url"},
}

// webFetchOutputSchema defines the output for the WebFetch tool
var webFetchOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"content": {
			Type:        "string",
			Description: "The fetched content from the URL",
		},
		"status_code": {
			Type:        "number",
			Description: "HTTP status code of the response",
		},
		"status_text": {
			Type:        "string",
			Description: "HTTP status text (e.g., '200 OK')",
		},
		"headers": {
			Type:        "object",
			Description: "Response headers",
			// Headers can have any string keys and values
		},
	},
	Required: []string{"content", "status_code", "status_text"},
}

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("web_fetch", WebFetch(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "web_fetch",
			Category:    "web",
			Tags:        []string{"http", "fetch", "download", "web", "network"},
			Description: "Fetches content from a URL with customizable timeout",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic fetch",
					Description: "Fetch a simple web page",
					Code:        `WebFetch().Execute(ctx, WebFetchParams{URL: "https://example.com"})`,
				},
				{
					Name:        "With timeout",
					Description: "Fetch with custom timeout",
					Code:        `WebFetch().Execute(ctx, WebFetchParams{URL: "https://slow-api.com", Timeout: 60})`,
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

// buildAuthConfig builds authentication configuration from parameters
func buildAuthConfig(params WebFetchParams) *auth.AuthConfig {
	if params.AuthType == "" {
		return nil
	}

	data := make(map[string]interface{})

	switch params.AuthType {
	case "bearer":
		if params.AuthToken == "" {
			return nil
		}
		data["token"] = params.AuthToken
	case "basic":
		if params.AuthUsername == "" || params.AuthPassword == "" {
			return nil
		}
		data["username"] = params.AuthUsername
		data["password"] = params.AuthPassword
	case "api_key":
		if params.AuthAPIKey == "" {
			return nil
		}
		data["api_key"] = params.AuthAPIKey
		if params.AuthKeyName != "" {
			data["key_name"] = params.AuthKeyName
		} else {
			data["key_name"] = "X-API-Key"
		}
		if params.AuthKeyLocation != "" {
			data["key_location"] = params.AuthKeyLocation
		} else {
			data["key_location"] = "header"
		}
	case "oauth2":
		if params.AuthToken == "" {
			return nil
		}
		data["access_token"] = params.AuthToken
	case "custom":
		if params.AuthHeaderName == "" || params.AuthHeaderValue == "" {
			return nil
		}
		data["header_name"] = params.AuthHeaderName
		data["header_value"] = params.AuthHeaderValue
		if params.AuthPrefix != "" {
			data["prefix"] = params.AuthPrefix
		}
	default:
		return nil
	}

	return &auth.AuthConfig{
		Type: params.AuthType,
		Data: data,
	}
}

// WebFetch creates a tool for fetching content from web URLs with authentication support.
// It provides HTTP GET functionality with customizable timeouts, multiple authentication methods
// (bearer, basic, API key, OAuth2, custom headers), automatic content decoding, header extraction,
// and proper resource cleanup while supporting context-aware cancellation and state-based configuration.
func WebFetch() domain.Tool {
	builder := atools.NewToolBuilder("web_fetch", "Fetches content from a URL with customizable timeout").
		WithFunction(func(ctx *domain.ToolContext, params WebFetchParams) (*WebFetchResult, error) {
			// Emit start event
			if ctx.Events != nil {
				ctx.Events.EmitMessage(fmt.Sprintf("Starting web fetch for %s", params.URL))
			}

			// Set default timeout
			timeout := 30 * time.Second
			if params.Timeout > 0 {
				timeout = time.Duration(params.Timeout) * time.Second
			}

			// Check state for custom user agent
			userAgent := "go-llms/1.0"
			if ctx.State != nil {
				if ua, exists := ctx.State.Get("user_agent"); exists {
					if uaStr, ok := ua.(string); ok {
						userAgent = uaStr
					}
				}
			}

			// Create HTTP client with timeout
			client := &http.Client{
				Timeout: timeout,
			}

			// Create request with context
			req, err := http.NewRequestWithContext(ctx.Context, "GET", params.URL, nil)
			if err != nil {
				return nil, fmt.Errorf("error creating request: %w", err)
			}

			// Set user agent
			req.Header.Set("User-Agent", userAgent)

			// Check state for additional headers
			if ctx.State != nil {
				if headers, exists := ctx.State.Get("http_headers"); exists {
					if headerMap, ok := headers.(map[string]string); ok {
						for key, value := range headerMap {
							req.Header.Set(key, value)
						}
					}
				}
			}

			// Apply authentication if provided
			authConfig := buildAuthConfig(params)
			if authConfig == nil && ctx.State != nil {
				// Try to detect authentication from state
				authConfig = auth.DetectAuthFromState(ctx.State, params.URL, nil)
			}
			if authConfig != nil {
				authMap := auth.ConvertAuthConfigToMap(authConfig)
				if err := auth.ApplyAuth(req, authMap); err != nil {
					return nil, fmt.Errorf("authentication failed: %w", err)
				}
				if ctx.Events != nil {
					ctx.Events.EmitMessage("Authentication applied")
				}
			}

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(1, 4, "Request prepared")
			}

			// Execute request
			resp, err := client.Do(req)
			if err != nil {
				if ctx.Events != nil {
					ctx.Events.EmitError(err)
				}
				return nil, fmt.Errorf("error fetching URL: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(2, 4, "Response received")
			}

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response: %w", err)
			}

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(3, 4, "Response body read")
			}

			// Extract headers
			headers := make(map[string]string)
			for key, values := range resp.Header {
				if len(values) > 0 {
					headers[key] = values[0]
				}
			}

			// Emit completion event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(4, 4, "Complete")
				ctx.Events.EmitCustom("fetch_complete", map[string]interface{}{
					"status":      resp.StatusCode,
					"contentSize": len(body),
				})
			}

			return &WebFetchResult{
				Content:    string(body),
				Status:     resp.StatusCode,
				StatusText: resp.Status,
				Headers:    headers,
			}, nil
		}).
		WithParameterSchema(webFetchParamSchema).
		WithOutputSchema(webFetchOutputSchema).
		WithUsageInstructions(`Use this tool to fetch content from a URL with optional authentication. The tool handles:
- HTTP/HTTPS URLs
- Customizable timeout (default 30 seconds)
- Multiple authentication methods (bearer, basic, API key, OAuth2, custom)
- Automatic content decoding
- Header extraction
- Proper error handling and status codes

Authentication methods:
- bearer: Sends "Authorization: Bearer <token>" header
- basic: Sends HTTP Basic Authentication with username/password
- api_key: Sends API key in header, query, or cookie
- oauth2: Sends OAuth2 access token as bearer token
- custom: Sends custom header with optional prefix

The tool will follow redirects automatically and handle common web server responses.
User agent can be customized via state (user_agent key).
Authentication can be auto-detected from state or provided via parameters.`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Fetch a web page",
				Description: "Basic web page retrieval",
				Scenario:    "When you need to get the content of a web page",
				Input: map[string]interface{}{
					"url": "https://example.com",
				},
				Output: map[string]interface{}{
					"content":     "<!DOCTYPE html>\n<html>\n<head><title>Example Domain</title>...</html>",
					"status_code": 200,
					"status_text": "200 OK",
					"headers": map[string]string{
						"Content-Type": "text/html; charset=UTF-8",
					},
				},
				Explanation: "Successfully fetched the HTML content",
			},
			{
				Name:        "Fetch API endpoint",
				Description: "Retrieve JSON from an API",
				Scenario:    "When fetching data from a REST API",
				Input: map[string]interface{}{
					"url": "https://api.github.com/users/octocat",
				},
				Output: map[string]interface{}{
					"content":     `{"login":"octocat","id":1,"node_id":"MDQ6VXNlcjE=","avatar_url":"https://github.com/images/error/octocat_happy.gif"...}`,
					"status_code": 200,
					"status_text": "200 OK",
					"headers": map[string]string{
						"Content-Type": "application/json; charset=utf-8",
					},
				},
				Explanation: "Fetched JSON data from GitHub API",
			},
			{
				Name:        "With custom timeout",
				Description: "Fetch with extended timeout",
				Scenario:    "When dealing with slow servers",
				Input: map[string]interface{}{
					"url":     "https://slow-server.example.com/large-file",
					"timeout": 120,
				},
				Output: map[string]interface{}{
					"content":     "[Large file content...]",
					"status_code": 200,
					"status_text": "200 OK",
				},
				Explanation: "Extended timeout allowed slow server to respond",
			},
			{
				Name:        "Handle 404 error",
				Description: "Non-existent page",
				Scenario:    "When the URL doesn't exist",
				Input: map[string]interface{}{
					"url": "https://example.com/does-not-exist",
				},
				Output: map[string]interface{}{
					"content":     "404 page not found",
					"status_code": 404,
					"status_text": "404 Not Found",
				},
				Explanation: "Returns error content with 404 status",
			},
			{
				Name:        "Handle redirect",
				Description: "Follow redirects automatically",
				Scenario:    "When the server redirects to another URL",
				Input: map[string]interface{}{
					"url": "http://github.com", // HTTP redirects to HTTPS
				},
				Output: map[string]interface{}{
					"content":     "[GitHub homepage HTML...]",
					"status_code": 200,
					"status_text": "200 OK",
					"headers": map[string]string{
						"Content-Type": "text/html; charset=utf-8",
					},
				},
				Explanation: "Automatically followed redirect from HTTP to HTTPS",
			},
			{
				Name:        "Timeout error",
				Description: "Request times out",
				Scenario:    "When server doesn't respond in time",
				Input: map[string]interface{}{
					"url":     "https://very-slow-server.example.com",
					"timeout": 5,
				},
				Output: map[string]interface{}{
					"error": "request timeout after 5s",
				},
				Explanation: "Request exceeded the timeout limit",
			},
			{
				Name:        "Bearer token authentication",
				Description: "Fetch with bearer token",
				Scenario:    "When accessing protected API with bearer token",
				Input: map[string]interface{}{
					"url":        "https://api.github.com/user",
					"auth_type":  "bearer",
					"auth_token": "ghp_xxxxxxxxxxxxxxxxxxxx",
				},
				Output: map[string]interface{}{
					"content":     `{"login":"username","id":12345,...}`,
					"status_code": 200,
					"status_text": "200 OK",
				},
				Explanation: "Bearer token added to Authorization header",
			},
			{
				Name:        "API key authentication",
				Description: "Fetch with API key in header",
				Scenario:    "When API requires key in custom header",
				Input: map[string]interface{}{
					"url":           "https://api.example.com/data",
					"auth_type":     "api_key",
					"auth_api_key":  "abc123xyz",
					"auth_key_name": "X-API-Key",
				},
				Output: map[string]interface{}{
					"content":     `{"data":[1,2,3]}`,
					"status_code": 200,
					"status_text": "200 OK",
				},
				Explanation: "API key sent in X-API-Key header",
			},
			{
				Name:        "Basic authentication",
				Description: "Fetch with username/password",
				Scenario:    "When API uses HTTP Basic Auth",
				Input: map[string]interface{}{
					"url":           "https://api.example.com/protected",
					"auth_type":     "basic",
					"auth_username": "user",
					"auth_password": "pass",
				},
				Output: map[string]interface{}{
					"content":     `{"message":"authenticated"}`,
					"status_code": 200,
					"status_text": "200 OK",
				},
				Explanation: "Credentials sent via HTTP Basic Authentication",
			},
			{
				Name:        "Invalid URL",
				Description: "Malformed URL",
				Scenario:    "When URL is not valid",
				Input: map[string]interface{}{
					"url": "not-a-valid-url",
				},
				Output: map[string]interface{}{
					"error": "invalid URL: not-a-valid-url",
				},
				Explanation: "URL validation failed",
			},
		}).
		WithConstraints([]string{
			"Only HTTP and HTTPS protocols are supported",
			"Default timeout is 30 seconds",
			"Follows redirects automatically (up to 10 redirects)",
			"Content size is limited by available memory",
			"Binary content is returned as-is",
			"Encoding is detected from Content-Type header",
			"User agent defaults to 'go-llms/1.0'",
			"Does not execute JavaScript or parse dynamic content",
			"Response headers are normalized to lowercase keys",
			"Authentication is optional and supports bearer, basic, API key, OAuth2, and custom methods",
			"Auth credentials should be kept secure and not logged",
			"State-based auth detection looks for common token patterns",
		}).
		WithErrorGuidance(map[string]string{
			"invalid URL":               "Ensure the URL is properly formatted with http:// or https:// prefix",
			"request timeout":           "Increase the timeout parameter or check if the server is responding",
			"connection refused":        "Check if the server is running and accessible",
			"certificate verification":  "The SSL certificate may be invalid or self-signed",
			"no such host":              "Check the domain name is correct and DNS is resolving",
			"unexpected EOF":            "Server closed connection unexpectedly - try again",
			"too many redirects":        "The URL is causing a redirect loop",
			"unsupported protocol":      "Only HTTP and HTTPS are supported",
			"context deadline exceeded": "Request was cancelled - increase timeout or check network",
			"permission denied":         "Check firewall settings or network permissions",
			"authentication failed":     "Check auth credentials are correct and match the expected format",
			"401 Unauthorized":          "Authentication required - check auth_type and credentials",
			"403 Forbidden":             "Access denied - credentials may be invalid or insufficient permissions",
		}).
		WithCategory("web").
		WithTags([]string{"web", "http", "fetch", "download", "network", "auth", "authentication"}).
		WithVersion("3.0.0").
		WithBehavior(
			false,  // Not deterministic - content can change
			false,  // Not destructive - only reads
			false,  // No confirmation needed
			"fast", // Usually fast, depends on network
		)

	return builder.Build()
}

// MustGetWebFetch retrieves the registered WebFetch tool or panics if not found.
// This is a convenience function for users who want to ensure the tool exists
// and prefer a panic over error handling for missing tools in their initialization code.
func MustGetWebFetch() domain.Tool {
	return tools.MustGetTool("web_fetch")
}
