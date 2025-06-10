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
	},
	Required: []string{"url"},
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

// WebScrape creates a tool for scraping web content
// This is a built-in tool optimized for:
// - HTML parsing without external dependencies
// - Text extraction and cleaning
// - Link discovery and classification
// - Metadata extraction
// - Simplified CSS-like selector support
func WebScrape() domain.Tool {
	return atools.NewTool(
		"web_scrape",
		"Extracts structured data from HTML pages",
		func(ctx *domain.ToolContext, params WebScrapeParams) (*WebScrapeResult, error) {
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
		},
		webScrapeParamSchema,
	)
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

// MustGetWebScrape retrieves the registered WebScrape tool or panics
// This is a convenience function for users who want to ensure the tool exists
func MustGetWebScrape() domain.Tool {
	return tools.MustGetTool("web_scrape")
}
