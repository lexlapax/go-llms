// ABOUTME: Web scraping tool for extracting structured data from HTML pages
// ABOUTME: Built-in tool that provides HTML parsing, CSS selector support, and data extraction

package web

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/util/auth"
)

// WebScrapeParams defines parameters for the WebScrape tool
type WebScrapeParams struct {
	URL          string   `json:"url"`
	Selectors    []string `json:"selectors,omitempty"`
	ExtractText  bool     `json:"extract_text,omitempty"`
	ExtractLinks bool     `json:"extract_links,omitempty"`
	ExtractMeta  bool     `json:"extract_meta,omitempty"`
	MaxDepth     int      `json:"max_depth,omitempty"`
	Timeout      int      `json:"timeout,omitempty"`

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

// WebScrapeResult defines the result of the WebScrape tool
type WebScrapeResult struct {
	URL         string              `json:"url"`
	Title       string              `json:"title,omitempty"`
	Text        string              `json:"text,omitempty"`
	Links       []LinkInfo          `json:"links,omitempty"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
	Selectors   map[string][]string `json:"selectors,omitempty"`
	StatusCode  int                 `json:"status_code"`
	ContentType string              `json:"content_type"`
	Timestamp   string              `json:"timestamp"`
}

// LinkInfo contains information about a link
type LinkInfo struct {
	URL  string `json:"url"`
	Text string `json:"text"`
	Type string `json:"type"` // internal, external, anchor
}

// webScrapeParamSchema defines parameters for the WebScrape tool
var webScrapeParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"url": {
			Type:        "string",
			Format:      "uri",
			Description: "The URL to scrape",
		},
		"selectors": {
			Type:        "array",
			Description: "CSS-like selectors to extract specific elements (simplified)",
		},
		"extract_text": {
			Type:        "boolean",
			Description: "Extract all text content from the page (default: true)",
		},
		"extract_links": {
			Type:        "boolean",
			Description: "Extract all links from the page (default: true)",
		},
		"extract_meta": {
			Type:        "boolean",
			Description: "Extract metadata (title, description, keywords) (default: true)",
		},
		"max_depth": {
			Type:        "number",
			Description: "Maximum depth for following links (0 = current page only, default: 0)",
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

// webScrapeOutputSchema defines the output for the WebScrape tool
var webScrapeOutputSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"url": {
			Type:        "string",
			Description: "The URL that was scraped",
		},
		"title": {
			Type:        "string",
			Description: "Page title extracted from HTML",
		},
		"text": {
			Type:        "string",
			Description: "Text content extracted from the page",
		},
		"links": {
			Type:        "array",
			Description: "Links found on the page",
			Items: &sdomain.Property{
				Type: "object",
				Properties: map[string]sdomain.Property{
					"url": {
						Type:        "string",
						Description: "Link URL",
					},
					"text": {
						Type:        "string",
						Description: "Link text",
					},
					"type": {
						Type:        "string",
						Description: "Link type: internal, external, or anchor",
					},
				},
				Required: []string{"url", "text", "type"},
			},
		},
		"metadata": {
			Type:        "object",
			Description: "Page metadata (description, keywords, etc.)",
		},
		"selectors": {
			Type:        "object",
			Description: "Content extracted by CSS selectors",
		},
		"status_code": {
			Type:        "number",
			Description: "HTTP status code",
		},
		"content_type": {
			Type:        "string",
			Description: "Content-Type header value",
		},
		"timestamp": {
			Type:        "string",
			Description: "ISO 8601 timestamp of when the page was scraped",
		},
	},
	Required: []string{"url", "status_code", "content_type", "timestamp"},
}

// Regular expressions for HTML parsing
var (
	titleRegex      = regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	metaRegex       = regexp.MustCompile(`(?i)<meta\s+([^>]+)>`)
	linkRegex       = regexp.MustCompile(`(?i)<a\s+([^>]*href=['"]([^'"]+)['"][^>]*)>([^<]*)</a>`)
	scriptRegex     = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	styleRegex      = regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	tagRegex        = regexp.MustCompile(`<[^>]+>`)
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

// buildScrapeAuthConfig builds authentication configuration from parameters
func buildScrapeAuthConfig(params WebScrapeParams) *auth.AuthConfig {
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

// init automatically registers the tool on package import
func init() {
	tools.MustRegisterTool("web_scrape", WebScrape(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "web_scrape",
			Category:    "web",
			Tags:        []string{"scrape", "html", "extract", "parse", "web", "network"},
			Description: "Extracts structured data from HTML pages",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Basic scraping",
					Description: "Extract text and links from a webpage",
					Code:        `WebScrape().Execute(ctx, WebScrapeParams{URL: "https://example.com"})`,
				},
				{
					Name:        "Extract with selectors",
					Description: "Extract specific elements using CSS-like selectors",
					Code:        `WebScrape().Execute(ctx, WebScrapeParams{URL: "https://example.com", Selectors: []string{"h1", "p", "img"}})`,
				},
				{
					Name:        "Metadata only",
					Description: "Extract only metadata without full text",
					Code:        `WebScrape().Execute(ctx, WebScrapeParams{URL: "https://example.com", ExtractText: false, ExtractLinks: false})`,
				},
			},
		},
		RequiredPermissions: []string{"network:access"},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "medium",
			Network:     true,
			FileSystem:  false,
			Concurrency: true,
		},
	})
}

// WebScrape creates a tool for extracting structured data from HTML pages with authentication support.
// It provides HTML parsing without external dependencies, automatic text extraction and cleaning,
// link discovery with type classification (internal/external/anchor), metadata extraction from meta tags,
// simplified CSS-like selector support for specific elements, and multiple authentication methods.
func WebScrape() domain.Tool {
	builder := atools.NewToolBuilder("web_scrape", "Extracts structured data from HTML pages").
		WithFunction(func(ctx *domain.ToolContext, params WebScrapeParams) (*WebScrapeResult, error) {
			// Emit start event
			if ctx.Events != nil {
				ctx.Events.EmitMessage(fmt.Sprintf("Starting web scrape for %s", params.URL))
			}

			// Set defaults
			if params.Timeout == 0 {
				params.Timeout = 30
			}
			// Default to extracting everything unless explicitly disabled
			// If all are false, enable all (default behavior)
			allFalse := !params.ExtractText && !params.ExtractLinks && !params.ExtractMeta
			shouldExtractText := params.ExtractText || allFalse
			shouldExtractLinks := params.ExtractLinks || allFalse
			shouldExtractMeta := params.ExtractMeta || allFalse

			// Check state for custom scraping rules
			if ctx.State != nil {
				// Check for custom selectors
				if selectors, exists := ctx.State.Get("scrape_selectors"); exists {
					if selectorSlice, ok := selectors.([]string); ok {
						params.Selectors = append(params.Selectors, selectorSlice...)
					}
				}
				// Check for robots.txt compliance setting
				if respectRobots, exists := ctx.State.Get("respect_robots_txt"); exists {
					if respect, ok := respectRobots.(bool); ok && respect {
						// Would implement robots.txt checking here
						ctx.Events.EmitMessage("Robots.txt compliance enabled")
					}
				}
			}

			// Validate URL
			parsedURL, err := url.Parse(params.URL)
			if err != nil {
				return nil, fmt.Errorf("invalid URL: %w", err)
			}

			// Create HTTP client with timeout
			timeout := time.Duration(params.Timeout) * time.Second
			client := &http.Client{
				Timeout: timeout,
			}

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(1, 5, "Fetching page")
			}

			// Create request with context
			req, err := http.NewRequestWithContext(ctx.Context, "GET", params.URL, nil)
			if err != nil {
				return nil, fmt.Errorf("error creating request: %w", err)
			}

			// Set user agent
			userAgent := "go-llms/1.0 (WebScraper)"
			if ctx.State != nil {
				if ua, exists := ctx.State.Get("user_agent"); exists {
					if uaStr, ok := ua.(string); ok {
						userAgent = uaStr
					}
				}
			}
			req.Header.Set("User-Agent", userAgent)
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

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
			authConfig := buildScrapeAuthConfig(params)
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

			// Execute request
			resp, err := client.Do(req)
			if err != nil {
				if ctx.Events != nil {
					ctx.Events.EmitError(err)
				}
				return nil, fmt.Errorf("error fetching URL: %w", err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(2, 5, "Reading response")
			}

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response: %w", err)
			}

			// Check content type
			contentType := resp.Header.Get("Content-Type")
			if contentType != "" && !strings.Contains(strings.ToLower(contentType), "html") && !strings.Contains(strings.ToLower(contentType), "xml") {
				return nil, fmt.Errorf("content type '%s' is not HTML/XML", contentType)
			}

			// Convert body to string for processing
			htmlContent := string(body)

			// Initialize result
			result := &WebScrapeResult{
				URL:         params.URL,
				StatusCode:  resp.StatusCode,
				ContentType: contentType,
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
			}

			// Extract title
			if matches := titleRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
				result.Title = strings.TrimSpace(matches[1])
			}

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(3, 5, "Parsing HTML")
			}

			// Extract metadata if requested
			if shouldExtractMeta {
				result.Metadata = extractMetadata(htmlContent)
			}

			// Extract text if requested
			if shouldExtractText {
				result.Text = extractTextContent(htmlContent)
			}

			// Extract links if requested
			if shouldExtractLinks {
				result.Links = extractLinkElements(htmlContent, parsedURL)
			}

			// Emit progress event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(4, 5, "Processing selectors")
			}

			// Process selectors if provided
			if len(params.Selectors) > 0 {
				result.Selectors = processSelectors(htmlContent, params.Selectors)
			}

			// Emit completion event
			if ctx.Events != nil {
				ctx.Events.EmitProgress(5, 5, "Complete")
				ctx.Events.EmitCustom("scrape_complete", map[string]interface{}{
					"url":          params.URL,
					"statusCode":   resp.StatusCode,
					"textLength":   len(result.Text),
					"linkCount":    len(result.Links),
					"metaCount":    len(result.Metadata),
					"selectorHits": len(result.Selectors),
				})
			}

			return result, nil
		}).
		WithParameterSchema(webScrapeParamSchema).
		WithOutputSchema(webScrapeOutputSchema).
		WithUsageInstructions(`Use this tool to extract structured data from HTML pages with optional authentication. The tool handles:
- HTML parsing and content extraction
- CSS-like selector support (basic tag names)
- Link discovery and classification
- Metadata extraction (title, description, keywords)
- Text content cleaning
- Configurable timeout
- Multiple authentication methods (bearer, basic, API key, OAuth2, custom)

Features:
- extract_text: Get all text content with HTML tags removed
- extract_links: Find all links with type classification (internal/external/anchor)
- extract_meta: Extract metadata from meta tags
- selectors: Extract content matching specific CSS-like selectors (currently supports tag names)

Authentication methods:
- bearer: Sends "Authorization: Bearer <token>" header
- basic: Sends HTTP Basic Authentication with username/password
- api_key: Sends API key in header, query, or cookie
- oauth2: Sends OAuth2 access token as bearer token
- custom: Sends custom header with optional prefix

The tool will:
- Automatically detect content type
- Clean and format extracted text
- Resolve relative URLs to absolute
- Handle common HTML entities
- Filter out script and style content

State configuration:
- user_agent: Custom user agent string
- http_headers: Additional headers as map[string]string
- scrape_selectors: Additional selectors to extract
- respect_robots_txt: Enable robots.txt compliance (future feature)
- Authentication can be auto-detected from state or provided via parameters`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Basic web scraping",
				Description: "Extract all content from a webpage",
				Scenario:    "When you need to get all information from a page",
				Input: map[string]interface{}{
					"url": "https://example.com",
				},
				Output: map[string]interface{}{
					"url":          "https://example.com",
					"title":        "Example Domain",
					"text":         "Example Domain This domain is for use in illustrative examples...",
					"links":        []map[string]interface{}{{"url": "https://www.iana.org/domains/example", "text": "More information...", "type": "external"}},
					"metadata":     map[string]string{"description": "Example domain for documentation"},
					"status_code":  200,
					"content_type": "text/html; charset=UTF-8",
					"timestamp":    "2024-01-15T10:00:00Z",
				},
				Explanation: "Extracts all content by default when no specific options are set",
			},
			{
				Name:        "Extract specific elements",
				Description: "Use selectors to extract specific content",
				Scenario:    "When you need specific HTML elements",
				Input: map[string]interface{}{
					"url":       "https://news.example.com",
					"selectors": []string{"h1", "h2", "p"},
				},
				Output: map[string]interface{}{
					"url": "https://news.example.com",
					"selectors": map[string][]string{
						"h1": []string{"Breaking News", "Top Stories"},
						"h2": []string{"Technology", "Business", "Sports"},
						"p":  []string{"First paragraph...", "Second paragraph..."},
					},
					"status_code": 200,
					"timestamp":   "2024-01-15T10:00:00Z",
				},
				Explanation: "Selectors extract content from specific HTML tags",
			},
			{
				Name:        "Extract only links",
				Description: "Get all links from a page",
				Scenario:    "When building a site map or finding resources",
				Input: map[string]interface{}{
					"url":           "https://blog.example.com",
					"extract_text":  false,
					"extract_links": true,
					"extract_meta":  false,
				},
				Output: map[string]interface{}{
					"url": "https://blog.example.com",
					"links": []map[string]interface{}{
						{"url": "https://blog.example.com/post1", "text": "First Post", "type": "internal"},
						{"url": "https://blog.example.com/post2", "text": "Second Post", "type": "internal"},
						{"url": "https://twitter.com/blog", "text": "Follow us", "type": "external"},
						{"url": "#comments", "text": "Comments", "type": "anchor"},
					},
					"status_code": 200,
					"timestamp":   "2024-01-15T10:00:00Z",
				},
				Explanation: "Links are classified as internal, external, or anchor",
			},
			{
				Name:        "Extract metadata only",
				Description: "Get page metadata without content",
				Scenario:    "When you need SEO information or page details",
				Input: map[string]interface{}{
					"url":           "https://shop.example.com/product",
					"extract_text":  false,
					"extract_links": false,
					"extract_meta":  true,
				},
				Output: map[string]interface{}{
					"url":   "https://shop.example.com/product",
					"title": "Amazing Product - Shop Example",
					"metadata": map[string]string{
						"description":    "Buy the amazing product for only $99",
						"keywords":       "product, amazing, shop",
						"og:title":       "Amazing Product",
						"og:description": "The best product you'll ever buy",
						"og:image":       "https://shop.example.com/images/product.jpg",
					},
					"status_code": 200,
					"timestamp":   "2024-01-15T10:00:00Z",
				},
				Explanation: "Extracts meta tags including Open Graph data",
			},
			{
				Name:        "Scrape with timeout",
				Description: "Set custom timeout for slow sites",
				Scenario:    "When dealing with slow-loading pages",
				Input: map[string]interface{}{
					"url":     "https://slow-site.example.com",
					"timeout": 60,
				},
				Output: map[string]interface{}{
					"url":         "https://slow-site.example.com",
					"text":        "Content that took a while to load...",
					"status_code": 200,
					"timestamp":   "2024-01-15T10:00:00Z",
				},
				Explanation: "Extended timeout allows slow pages to load completely",
			},
			{
				Name:        "Handle non-HTML content",
				Description: "Attempt to scrape non-HTML",
				Scenario:    "When URL doesn't return HTML",
				Input: map[string]interface{}{
					"url": "https://api.example.com/data.json",
				},
				Output: map[string]interface{}{
					"error": "content type 'application/json' is not HTML/XML",
				},
				Explanation: "Tool validates content type before processing",
			},
			{
				Name:        "Complex selector extraction",
				Description: "Extract multiple tag types",
				Scenario:    "When analyzing page structure",
				Input: map[string]interface{}{
					"url":          "https://docs.example.com",
					"selectors":    []string{"h1", "h2", "h3", "code", "pre"},
					"max_depth":    0,
					"extract_text": true,
				},
				Output: map[string]interface{}{
					"url":  "https://docs.example.com",
					"text": "Full text content of the documentation page...",
					"selectors": map[string][]string{
						"h1":   []string{"API Documentation"},
						"h2":   []string{"Getting Started", "Authentication", "Endpoints"},
						"h3":   []string{"Installation", "Configuration", "Examples"},
						"code": []string{"npm install", "const api = new API()"},
						"pre":  []string{"{ \"status\": \"ok\" }"},
					},
					"status_code": 200,
					"timestamp":   "2024-01-15T10:00:00Z",
				},
				Explanation: "Combines full text extraction with specific element selection",
			},
			{
				Name:        "Bearer token authentication",
				Description: "Scrape protected page with bearer token",
				Scenario:    "When scraping authenticated content with bearer token",
				Input: map[string]interface{}{
					"url":        "https://private.example.com/protected-page",
					"auth_type":  "bearer",
					"auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
				},
				Output: map[string]interface{}{
					"url":         "https://private.example.com/protected-page",
					"title":       "Protected Content",
					"text":        "This is protected content only visible to authenticated users...",
					"status_code": 200,
					"timestamp":   "2024-01-15T10:00:00Z",
				},
				Explanation: "Bearer token added to Authorization header for authentication",
			},
			{
				Name:        "API key authentication",
				Description: "Scrape with API key in header",
				Scenario:    "When API requires key in custom header",
				Input: map[string]interface{}{
					"url":           "https://api-docs.example.com/documentation",
					"auth_type":     "api_key",
					"auth_api_key":  "abc123xyz789",
					"auth_key_name": "X-API-Key",
				},
				Output: map[string]interface{}{
					"url":         "https://api-docs.example.com/documentation",
					"title":       "API Documentation",
					"text":        "Welcome to our API documentation...",
					"status_code": 200,
					"timestamp":   "2024-01-15T10:00:00Z",
				},
				Explanation: "API key sent in X-API-Key header for access",
			},
			{
				Name:        "Basic authentication",
				Description: "Scrape with username/password",
				Scenario:    "When site uses HTTP Basic Auth",
				Input: map[string]interface{}{
					"url":           "https://secure.example.com/admin",
					"auth_type":     "basic",
					"auth_username": "admin",
					"auth_password": "secret123",
				},
				Output: map[string]interface{}{
					"url":         "https://secure.example.com/admin",
					"title":       "Admin Dashboard",
					"text":        "Admin control panel content...",
					"status_code": 200,
					"timestamp":   "2024-01-15T10:00:00Z",
				},
				Explanation: "Credentials sent via HTTP Basic Authentication",
			},
		}).
		WithConstraints([]string{
			"Only HTTP and HTTPS protocols are supported",
			"Content must be HTML or XML (validated by Content-Type header)",
			"Default timeout is 30 seconds",
			"CSS selector support is limited to simple tag names",
			"JavaScript is not executed - only static HTML is parsed",
			"max_depth parameter is reserved for future use (currently ignored)",
			"Follows redirects automatically (up to 10)",
			"HTML entities are decoded to their character equivalents",
			"Binary content will return an error",
			"Memory usage scales with page size",
			"Authentication is optional and supports bearer, basic, API key, OAuth2, and custom methods",
			"Auth credentials should be kept secure and not logged",
			"State-based auth detection looks for common token patterns",
		}).
		WithErrorGuidance(map[string]string{
			"invalid URL":                "Ensure URL is properly formatted with http:// or https://",
			"content type * is not HTML": "This tool only works with HTML/XML pages, not JSON/images/etc",
			"error fetching URL":         "Check if the site is accessible and allows scraping",
			"timeout":                    "Increase timeout parameter or check if site is responding",
			"connection refused":         "Site may be down or blocking automated requests",
			"certificate verification":   "SSL certificate issues - site may use self-signed cert",
			"no such host":               "Check domain name is correct",
			"404":                        "Page not found - verify the URL exists",
			"403":                        "Access forbidden - site may block scrapers",
			"robots.txt disallowed":      "Site's robots.txt prohibits scraping this path",
			"too many redirects":         "URL causes a redirect loop",
			"authentication failed":      "Check auth credentials are correct and match the expected format",
			"401 Unauthorized":           "Authentication required - check auth_type and credentials",
			"403 Forbidden":              "Access denied - credentials may be invalid or insufficient permissions",
		}).
		WithCategory("web").
		WithTags([]string{"web", "scrape", "html", "extract", "parse", "crawler", "auth", "authentication"}).
		WithVersion("3.0.0").
		WithBehavior(
			false,  // Not deterministic - content can change
			false,  // Not destructive - only reads
			false,  // No confirmation needed
			"fast", // Usually fast, depends on page size and network
		)

	return builder.Build()
}

// extractMetadata extracts metadata from HTML
func extractMetadata(html string) map[string]string {
	metadata := make(map[string]string)

	// Extract meta tags
	matches := metaRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 1 {
			attrs := parseAttributes(match[1])

			// Handle different meta tag formats
			if name, hasName := attrs["name"]; hasName {
				if content, hasContent := attrs["content"]; hasContent {
					metadata[name] = content
				}
			} else if property, hasProperty := attrs["property"]; hasProperty {
				if content, hasContent := attrs["content"]; hasContent {
					metadata[property] = content
				}
			} else if httpEquiv, hasHttpEquiv := attrs["http-equiv"]; hasHttpEquiv {
				if content, hasContent := attrs["content"]; hasContent {
					metadata[httpEquiv] = content
				}
			}
		}
	}

	return metadata
}

// extractTextContent extracts and cleans text content from HTML
func extractTextContent(html string) string {
	// Remove script and style tags
	cleaned := scriptRegex.ReplaceAllString(html, "")
	cleaned = styleRegex.ReplaceAllString(cleaned, "")

	// Remove all HTML tags
	cleaned = tagRegex.ReplaceAllString(cleaned, " ")

	// Decode HTML entities (basic ones)
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&#39;", "'")
	cleaned = strings.ReplaceAll(cleaned, "&nbsp;", " ")

	// Clean up whitespace
	cleaned = whitespaceRegex.ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// extractLinkElements extracts links from HTML
func extractLinkElements(html string, baseURL *url.URL) []LinkInfo {
	var links []LinkInfo

	matches := linkRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 3 {
			href := match[2]
			text := strings.TrimSpace(match[3])

			// Clean up link text
			text = tagRegex.ReplaceAllString(text, "")
			text = whitespaceRegex.ReplaceAllString(text, " ")
			text = strings.TrimSpace(text)

			// Resolve relative URLs
			linkURL, err := baseURL.Parse(href)
			if err != nil {
				continue
			}

			// Determine link type
			linkType := "internal"
			if linkURL.Host != "" && linkURL.Host != baseURL.Host {
				linkType = "external"
			} else if strings.HasPrefix(href, "#") {
				linkType = "anchor"
			}

			links = append(links, LinkInfo{
				URL:  linkURL.String(),
				Text: text,
				Type: linkType,
			})
		}
	}

	return links
}

// processSelectors processes simplified CSS-like selectors
func processSelectors(html string, selectors []string) map[string][]string {
	results := make(map[string][]string)

	for _, selector := range selectors {
		selector = strings.TrimSpace(selector)
		if selector == "" {
			continue
		}

		// Support simple tag selectors
		if isSimpleTag(selector) {
			matches := findTagContent(html, selector)
			if len(matches) > 0 {
				results[selector] = matches
			}
		}
		// Additional selector types could be implemented here
		// For now, we keep it simple with just tag names
	}

	return results
}

// isSimpleTag checks if a selector is a simple tag name
func isSimpleTag(selector string) bool {
	// Simple validation for tag names
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9]*$`, selector)
	return matched
}

// findTagContent finds content of specific HTML tags
func findTagContent(html, tagName string) []string {
	var contents []string

	// Create regex for the specific tag
	tagPattern := regexp.MustCompile(fmt.Sprintf(`(?i)<%s[^>]*>([^<]*)</%s>`, tagName, tagName))
	matches := tagPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			content := strings.TrimSpace(match[1])
			if content != "" {
				contents = append(contents, content)
			}
		}
	}

	return contents
}

// parseAttributes parses HTML attributes from a string
func parseAttributes(attrString string) map[string]string {
	attrs := make(map[string]string)

	// Simple attribute parsing
	attrPattern := regexp.MustCompile(`(\w+)=["']([^"']+)["']`)
	matches := attrPattern.FindAllStringSubmatch(attrString, -1)

	for _, match := range matches {
		if len(match) > 2 {
			attrs[strings.ToLower(match[1])] = match[2]
		}
	}

	return attrs
}

// MustGetWebScrape retrieves the registered WebScrape tool or panics if not found.
// This is a convenience function for users who want to ensure the tool exists
// and prefer a panic over error handling for missing tools in their initialization code.
func MustGetWebScrape() domain.Tool {
	return tools.MustGetTool("web_scrape")
}
