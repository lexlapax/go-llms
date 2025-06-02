// ABOUTME: HTTP request tool for advanced HTTP operations with full method support
// ABOUTME: Built-in tool supporting POST, PUT, DELETE, PATCH with headers, auth, and body options

package web

import (
	"bytes"
	"context"
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
		func(ctx context.Context, params HTTPRequestParams) (*HTTPRequestResult, error) {
			startTime := time.Now()

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

			// Prepare request body
			var bodyReader io.Reader
			if params.Body != "" {
				bodyReader = bytes.NewBufferString(params.Body)
			}

			// Create request
			req, err := http.NewRequestWithContext(ctx, params.Method, parsedURL.String(), bodyReader)
			if err != nil {
				return nil, fmt.Errorf("error creating request: %w", err)
			}

			// Set default headers
			req.Header.Set("User-Agent", "go-llms/1.0 (HTTPRequest)")

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
			for key, value := range params.Headers {
				req.Header.Set(key, value)
			}

			// Apply authentication
			if params.AuthType != "" {
				if err := applyAuthFromParams(req, &params); err != nil {
					return nil, fmt.Errorf("error applying authentication: %w", err)
				}
			}

			// Create HTTP client
			client := &http.Client{
				Timeout: time.Duration(params.Timeout) * time.Second,
			}

			// Configure redirect policy
			if !followRedirects {
				client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
			}

			// Execute request
			resp, err := client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("error executing request: %w", err)
			}
			defer resp.Body.Close()

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

			// Build result
			result := &HTTPRequestResult{
				StatusCode:    resp.StatusCode,
				Status:        resp.Status,
				Headers:       headers,
				Body:          string(body),
				ContentType:   resp.Header.Get("Content-Type"),
				ContentLength: resp.ContentLength,
				ResponseTime:  time.Since(startTime).Milliseconds(),
			}

			// Check for redirect
			if location := resp.Header.Get("Location"); location != "" {
				result.RedirectURL = location
			}

			return result, nil
		},
		httpRequestParamSchema,
	)
}

// applyAuthFromParams applies authentication to the request using flattened params
func applyAuthFromParams(req *http.Request, params *HTTPRequestParams) error {
	switch strings.ToLower(params.AuthType) {
	case "basic":
		if params.AuthUsername == "" || params.AuthPassword == "" {
			return fmt.Errorf("basic auth requires auth_username and auth_password")
		}
		req.SetBasicAuth(params.AuthUsername, params.AuthPassword)

	case "bearer":
		if params.AuthToken == "" {
			return fmt.Errorf("bearer auth requires auth_token")
		}
		req.Header.Set("Authorization", "Bearer "+params.AuthToken)

	case "api_key":
		if params.AuthKeyName == "" || params.AuthKeyValue == "" {
			return fmt.Errorf("api_key auth requires auth_key_name and auth_key_value")
		}

		location := strings.ToLower(params.AuthKeyLocation)
		if location == "" {
			location = "header"
		}

		switch location {
		case "header":
			req.Header.Set(params.AuthKeyName, params.AuthKeyValue)
		case "query":
			q := req.URL.Query()
			q.Add(params.AuthKeyName, params.AuthKeyValue)
			req.URL.RawQuery = q.Encode()
		default:
			return fmt.Errorf("invalid auth_key_location: %s (use 'header' or 'query')", location)
		}

	default:
		return fmt.Errorf("unsupported auth type: %s", params.AuthType)
	}

	return nil
}

// MustGetHTTPRequest retrieves the registered HTTPRequest tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetHTTPRequest() domain.Tool {
	return tools.MustGetTool("http_request")
}
