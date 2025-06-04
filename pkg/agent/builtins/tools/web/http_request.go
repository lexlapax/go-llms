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

// HTTPRequest creates a tool for making HTTP requests
// This is a built-in tool optimized for:
// - Full HTTP method support (GET, POST, PUT, DELETE, PATCH, etc.)
// - Multiple authentication methods
// - Custom headers and query parameters
// - Various body content types
// - Redirect control
// - Comprehensive response information
func HTTPRequest() domain.Tool {
	return atools.NewTool(
		"http_request",
		"Makes HTTP requests with full method and authentication support",
		func(ctx *domain.ToolContext, params HTTPRequestParams) (*HTTPRequestResult, error) {
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
			followRedirects := true
			if !params.FollowRedirects {
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
			defer resp.Body.Close()

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
		},
		httpRequestParamSchema,
	)
}

// MustGetHTTPRequest retrieves the registered HTTPRequest tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetHTTPRequest() domain.Tool {
	return tools.MustGetTool("http_request")
}