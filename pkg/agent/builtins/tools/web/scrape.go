// ABOUTME: Web scraping tool for extracting structured data from HTML pages
// ABOUTME: Built-in tool that provides HTML parsing, CSS selector support, and data extraction

package web

import (
	"context"
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
		func(ctx context.Context, params WebScrapeParams) (*WebScrapeResult, error) {
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

			// Create request with context
			req, err := http.NewRequestWithContext(ctx, "GET", params.URL, nil)
			if err != nil {
				return nil, fmt.Errorf("error creating request: %w", err)
			}

			// Set user agent
			req.Header.Set("User-Agent", "go-llms/1.0 (WebScraper)")
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

			// Execute request
			resp, err := client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("error fetching URL: %w", err)
			}
			defer resp.Body.Close()

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

			// Process selectors if provided
			if len(params.Selectors) > 0 {
				result.Selectors = processSelectors(htmlContent, params.Selectors)
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
			if name, ok := attrs["name"]; ok {
				if content, ok := attrs["content"]; ok {
					metadata[name] = content
				}
			} else if property, ok := attrs["property"]; ok {
				if content, ok := attrs["content"]; ok {
					metadata[property] = content
				}
			} else if httpEquiv, ok := attrs["http-equiv"]; ok {
				if content, ok := attrs["content"]; ok {
					metadata[httpEquiv] = content
				}
			}
		}
	}

	return metadata
}

// extractTextContent extracts and cleans text content from HTML
func extractTextContent(html string) string {
	// Remove scripts and styles
	cleaned := scriptRegex.ReplaceAllString(html, " ")
	cleaned = styleRegex.ReplaceAllString(cleaned, " ")

	// Remove HTML tags
	cleaned = tagRegex.ReplaceAllString(cleaned, " ")

	// Clean up whitespace
	cleaned = whitespaceRegex.ReplaceAllString(cleaned, " ")

	// Trim and return
	return strings.TrimSpace(cleaned)
}

// extractLinkElements extracts all links from HTML
func extractLinkElements(html string, baseURL *url.URL) []LinkInfo {
	var links []LinkInfo
	linkMap := make(map[string]bool) // Deduplicate links

	matches := linkRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 3 {
			href := match[2]
			linkText := strings.TrimSpace(match[3])

			// Skip empty or anchor-only links
			if href == "" || href == "#" {
				continue
			}

			// Parse and resolve the link
			linkURL, err := url.Parse(href)
			if err != nil {
				continue
			}

			// Resolve relative URLs
			absoluteURL := baseURL.ResolveReference(linkURL)
			urlStr := absoluteURL.String()

			// Skip if we've already seen this link
			if linkMap[urlStr] {
				continue
			}
			linkMap[urlStr] = true

			// Determine link type
			linkType := "internal"
			if strings.HasPrefix(href, "#") {
				linkType = "anchor"
			} else if absoluteURL.Host != baseURL.Host {
				linkType = "external"
			}

			links = append(links, LinkInfo{
				URL:  urlStr,
				Text: linkText,
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
		selector = strings.TrimSpace(strings.ToLower(selector))

		// Simple tag selector support
		if isSimpleTag(selector) {
			results[selector] = extractTagContent(html, selector)
		} else if strings.HasPrefix(selector, ".") {
			// Simple class selector
			className := selector[1:]
			results[selector] = extractByClass(html, className)
		} else if strings.HasPrefix(selector, "#") {
			// Simple ID selector
			id := selector[1:]
			results[selector] = extractByID(html, id)
		} else {
			// For complex selectors, return empty
			results[selector] = []string{"Complex selectors not supported in this implementation"}
		}
	}

	return results
}

// isSimpleTag checks if the selector is a simple HTML tag
func isSimpleTag(selector string) bool {
	// List of common HTML tags
	commonTags := map[string]bool{
		"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
		"p": true, "div": true, "span": true, "a": true, "img": true,
		"ul": true, "ol": true, "li": true, "table": true, "tr": true, "td": true, "th": true,
		"form": true, "input": true, "button": true, "textarea": true, "select": true,
		"header": true, "footer": true, "nav": true, "main": true, "article": true, "section": true,
	}
	return commonTags[selector]
}

// extractTagContent extracts content from specific HTML tags
func extractTagContent(html, tag string) []string {
	var results []string

	// Create regex for the specific tag
	tagPattern := regexp.MustCompile(fmt.Sprintf(`(?i)<%s[^>]*>(.*?)</%s>`, tag, tag))
	matches := tagPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			// Clean the content
			content := tagRegex.ReplaceAllString(match[1], " ")
			content = whitespaceRegex.ReplaceAllString(content, " ")
			content = strings.TrimSpace(content)
			if content != "" {
				results = append(results, content)
			}
		}
	}

	return results
}

// extractByClass extracts elements by class name
func extractByClass(html, className string) []string {
	var results []string

	// Simplified class extraction - finds elements with the specified class
	classPattern := regexp.MustCompile(fmt.Sprintf(`(?i)<[^>]+class=['"][^'"]*\b%s\b[^'"]*['"][^>]*>(.*?)</[^>]+>`, className))
	matches := classPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			content := tagRegex.ReplaceAllString(match[1], " ")
			content = whitespaceRegex.ReplaceAllString(content, " ")
			content = strings.TrimSpace(content)
			if content != "" {
				results = append(results, content)
			}
		}
	}

	return results
}

// extractByID extracts element by ID
func extractByID(html, id string) []string {
	var results []string

	// Simplified ID extraction - finds element with the specified ID
	idPattern := regexp.MustCompile(fmt.Sprintf(`(?i)<[^>]+id=['"]%s['"][^>]*>(.*?)</[^>]+>`, id))
	matches := idPattern.FindAllStringSubmatch(html, -1)

	if len(matches) > 0 && len(matches[0]) > 1 {
		content := tagRegex.ReplaceAllString(matches[0][1], " ")
		content = whitespaceRegex.ReplaceAllString(content, " ")
		content = strings.TrimSpace(content)
		if content != "" {
			results = append(results, content)
		}
	}

	return results
}

// parseAttributes parses HTML attributes from a string
func parseAttributes(attrStr string) map[string]string {
	attrs := make(map[string]string)

	// Simple attribute parsing
	attrPattern := regexp.MustCompile(`(\w+)=['"]([^'"]+)['"]`)
	matches := attrPattern.FindAllStringSubmatch(attrStr, -1)

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
