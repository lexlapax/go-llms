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
)

// WebFetchParams defines parameters for the WebFetch tool
type WebFetchParams struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout,omitempty"` // Timeout in seconds, default 30
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
	},
	Required: []string{"url"},
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

// WebFetch creates a tool for fetching web content
// This is a built-in tool optimized for:
// - Context-aware cancellation
// - Customizable timeouts
// - Header capture
// - Proper resource cleanup
func WebFetch() domain.Tool {
	return atools.NewTool(
		"web_fetch",
		"Fetches content from a URL with customizable timeout",
		func(ctx *domain.ToolContext, params WebFetchParams) (*WebFetchResult, error) {
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
			defer resp.Body.Close()

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
		},
		webFetchParamSchema,
	)
}

// MustGetWebFetch retrieves the registered WebFetch tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetWebFetch() domain.Tool {
	return tools.MustGetTool("web_fetch")
}
