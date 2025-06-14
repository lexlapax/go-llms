// ABOUTME: XML parser implementation with recovery capabilities
// ABOUTME: Handles XML parsing with error recovery and markdown extraction

package outputs

import (
	"context"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// XMLParser implements Parser for XML format
type XMLParser struct {
	strictMode bool
}

// xmlNode represents a generic XML node
type xmlNode struct {
	XMLName    xml.Name
	Attributes []xml.Attr `xml:",any,attr"`
	Content    string     `xml:",chardata"`
	Nodes      []xmlNode  `xml:",any"`
}

// NewXMLParser creates a new XML parser
func NewXMLParser() *XMLParser {
	return &XMLParser{
		strictMode: false,
	}
}

// Name returns the parser name
func (p *XMLParser) Name() string {
	return "xml"
}

// Parse attempts to parse XML output
func (p *XMLParser) Parse(ctx context.Context, output string) (interface{}, error) {
	// For XML, we'll parse into a generic structure
	result, err := p.parseXMLToInterface(output)
	if err != nil {
		if p.strictMode {
			return nil, fmt.Errorf("failed to parse XML: %w", err)
		}

		// Try basic cleanup
		cleaned := p.cleanXML(output)
		result, err = p.parseXMLToInterface(cleaned)
		if err != nil {
			return nil, fmt.Errorf("failed to parse XML after cleanup: %w", err)
		}
	}

	return result, nil
}

// ParseWithRecovery attempts to parse with advanced recovery options
func (p *XMLParser) ParseWithRecovery(ctx context.Context, output string, opts *RecoveryOptions) (interface{}, error) {
	if opts == nil {
		opts = DefaultRecoveryOptions()
	}

	if opts.StrictMode || p.strictMode {
		return p.Parse(ctx, output)
	}

	attempts := 0
	var lastErr error

	// Try different recovery strategies
	strategies := []func(string) string{
		func(s string) string { return s }, // Original
		p.extractFromMarkdown,
		p.cleanXML,
		p.fixUnclosedTags,
		p.extractXMLBlock,
		p.wrapInRoot,
	}

	for _, strategy := range strategies {
		if attempts >= opts.MaxAttempts {
			break
		}

		processed := strategy(output)
		if processed == "" {
			continue
		}

		result, err := p.parseXMLToInterface(processed)
		if err == nil {
			return result, nil
		}

		lastErr = err
		attempts++
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to parse XML after %d recovery attempts: %w", attempts, lastErr)
	}

	return nil, fmt.Errorf("failed to parse XML after %d recovery attempts", attempts)
}

// ParseWithSchema attempts to parse using schema guidance
func (p *XMLParser) ParseWithSchema(ctx context.Context, output string, schema *OutputSchema) (interface{}, error) {
	result, err := p.ParseWithRecovery(ctx, output, &RecoveryOptions{
		ExtractFromMarkdown: true,
		FixCommonIssues:     true,
		MaxAttempts:         3,
		Schema:              schema,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse XML with schema guidance: %w", err)
	}

	return result, nil
}

// CanParse checks if the output might be XML
func (p *XMLParser) CanParse(output string) bool {
	trimmed := strings.TrimSpace(output)

	// Check for XML indicators
	if strings.HasPrefix(trimmed, "<?xml") {
		return true
	}

	// Check for XML-like structure
	if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") {
		// Simple check for balanced tags
		openCount := strings.Count(trimmed, "<")
		closeCount := strings.Count(trimmed, ">")
		return openCount > 0 && openCount == closeCount
	}

	// Check for markdown code blocks with xml
	if strings.Contains(output, "```xml") {
		return true
	}

	return false
}

// parseXMLToInterface parses XML into a generic interface
func (p *XMLParser) parseXMLToInterface(xmlStr string) (interface{}, error) {
	var root xmlNode
	if err := xml.Unmarshal([]byte(xmlStr), &root); err != nil {
		return nil, err
	}

	// Convert to generic map structure
	return p.xmlNodeToMap(&root), nil
}

// xmlNodeToMap converts an xmlNode to a map
func (p *XMLParser) xmlNodeToMap(node *xmlNode) map[string]interface{} {
	result := make(map[string]interface{})

	// Add attributes with @ prefix
	for _, attr := range node.Attributes {
		result["@"+attr.Name.Local] = attr.Value
	}

	// Handle content and child nodes
	if len(node.Nodes) == 0 {
		// Leaf node with just content
		if node.Content != "" && strings.TrimSpace(node.Content) != "" {
			return map[string]interface{}{
				node.XMLName.Local: strings.TrimSpace(node.Content),
			}
		}
		return map[string]interface{}{
			node.XMLName.Local: result,
		}
	}

	// Process child nodes
	children := make(map[string]interface{})
	for _, child := range node.Nodes {
		childMap := p.xmlNodeToMap(&child)
		for k, v := range childMap {
			if existing, exists := children[k]; exists {
				// Convert to array if multiple elements with same name
				switch e := existing.(type) {
				case []interface{}:
					children[k] = append(e, v)
				default:
					children[k] = []interface{}{e, v}
				}
			} else {
				children[k] = v
			}
		}
	}

	// Merge attributes and children
	for k, v := range children {
		result[k] = v
	}

	return map[string]interface{}{
		node.XMLName.Local: result,
	}
}

// extractFromMarkdown extracts XML from markdown code blocks
func (p *XMLParser) extractFromMarkdown(output string) string {
	// Try to extract from ```xml blocks
	re := regexp.MustCompile("```xml\n([^`]+)\n```")
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try generic code blocks
	re = regexp.MustCompile("```\n([^`]+)\n```")
	matches = re.FindStringSubmatch(output)
	if len(matches) > 1 {
		content := strings.TrimSpace(matches[1])
		if p.looksLikeXML(content) {
			return content
		}
	}

	return ""
}

// cleanXML performs basic XML cleanup
func (p *XMLParser) cleanXML(output string) string {
	cleaned := strings.TrimSpace(output)

	// Remove BOM if present
	cleaned = strings.TrimPrefix(cleaned, "\xef\xbb\xbf")

	// Fix common entity issues
	cleaned = strings.ReplaceAll(cleaned, "&", "&amp;")
	cleaned = strings.ReplaceAll(cleaned, "&amp;amp;", "&amp;") // Fix double encoding

	// Ensure proper quotes in attributes
	cleaned = p.fixAttributeQuotes(cleaned)

	return cleaned
}

// fixUnclosedTags attempts to fix unclosed tags
func (p *XMLParser) fixUnclosedTags(output string) string {
	// This is a simple implementation that handles basic cases
	// A production implementation would need a proper XML parser

	// Track open tags
	openTags := []string{}
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		// Find opening tags (but not self-closing)
		openRe := regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9]*)\b[^>]*[^/]>`)
		openMatches := openRe.FindAllStringSubmatch(line, -1)
		for _, match := range openMatches {
			openTags = append(openTags, match[1])
		}

		// Find closing tags
		closeRe := regexp.MustCompile(`</([a-zA-Z][a-zA-Z0-9]*)>`)
		closeMatches := closeRe.FindAllStringSubmatch(line, -1)
		for _, match := range closeMatches {
			// Remove from open tags
			for j := len(openTags) - 1; j >= 0; j-- {
				if openTags[j] == match[1] {
					openTags = append(openTags[:j], openTags[j+1:]...)
					break
				}
			}
		}

		// Find self-closing tags
		selfCloseRe := regexp.MustCompile(`<[a-zA-Z][a-zA-Z0-9]*\b[^>]*/>`)
		selfCloseMatches := selfCloseRe.FindAllString(line, -1)
		_ = selfCloseMatches // These don't need closing

		lines[i] = line
	}

	// Add closing tags for unclosed tags
	for i := len(openTags) - 1; i >= 0; i-- {
		lines = append(lines, fmt.Sprintf("</%s>", openTags[i]))
	}

	return strings.Join(lines, "\n")
}

// extractXMLBlock attempts to extract an XML block from text
func (p *XMLParser) extractXMLBlock(output string) string {
	// Find the first < and last >
	start := strings.Index(output, "<")
	if start == -1 {
		return ""
	}

	end := strings.LastIndex(output, ">")
	if end == -1 || end <= start {
		return ""
	}

	return output[start : end+1]
}

// wrapInRoot wraps content in a root element if needed
func (p *XMLParser) wrapInRoot(output string) string {
	trimmed := strings.TrimSpace(output)

	// Check if it already has a root element
	if strings.HasPrefix(trimmed, "<?xml") {
		// Find where the actual content starts
		idx := strings.Index(trimmed, "?>")
		if idx != -1 {
			trimmed = strings.TrimSpace(trimmed[idx+2:])
		}
	}

	// Count top-level elements
	topLevelCount := 0
	depth := 0

	for _, ch := range trimmed {
		switch ch {
		case '<':
			if depth == 0 {
				topLevelCount++
			}
			depth++
		case '>':
			depth--
		}
	}

	// If multiple top-level elements, wrap in root
	if topLevelCount > 1 {
		return fmt.Sprintf("<root>%s</root>", trimmed)
	}

	return output
}

// fixAttributeQuotes fixes attribute quote issues
func (p *XMLParser) fixAttributeQuotes(xmlStr string) string {
	// Fix attributes without quotes
	re := regexp.MustCompile(`(\s+)([a-zA-Z][a-zA-Z0-9_-]*)=([^"\s>]+)`)
	return re.ReplaceAllString(xmlStr, `$1$2="$3"`)
}

// looksLikeXML checks if content looks like XML
func (p *XMLParser) looksLikeXML(content string) bool {
	trimmed := strings.TrimSpace(content)
	return strings.HasPrefix(trimmed, "<") && strings.Contains(trimmed, ">")
}

// init registers the XML parser
func init() {
	_ = RegisterParser(NewXMLParser())
}
