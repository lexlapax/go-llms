// ABOUTME: YAML parser implementation with recovery capabilities
// ABOUTME: Handles YAML parsing with error recovery and markdown extraction

package outputs

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLParser implements Parser for YAML format
type YAMLParser struct {
	strictMode bool
}

// NewYAMLParser creates a new YAML parser
func NewYAMLParser() *YAMLParser {
	return &YAMLParser{
		strictMode: false,
	}
}

// Name returns the parser name
func (p *YAMLParser) Name() string {
	return "yaml"
}

// Parse attempts to parse YAML output
func (p *YAMLParser) Parse(ctx context.Context, output string) (interface{}, error) {
	var result interface{}
	if err := yaml.Unmarshal([]byte(output), &result); err != nil {
		if p.strictMode {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}

		// Try basic cleanup
		cleaned := p.cleanYAML(output)
		if err := yaml.Unmarshal([]byte(cleaned), &result); err != nil {
			return nil, fmt.Errorf("failed to parse YAML after cleanup: %w", err)
		}
	}

	return result, nil
}

// ParseWithRecovery attempts to parse with advanced recovery options
func (p *YAMLParser) ParseWithRecovery(ctx context.Context, output string, opts *RecoveryOptions) (interface{}, error) {
	if opts == nil {
		opts = DefaultRecoveryOptions()
	}

	if opts.StrictMode || p.strictMode {
		return p.Parse(ctx, output)
	}

	attempts := 0
	var lastErr error

	// Always clean the YAML first (handles tabs, BOM, etc.)
	output = p.cleanYAML(output)

	// Try different recovery strategies
	strategies := []func(string) string{
		func(s string) string { return s }, // Original (now cleaned)
		p.extractFromMarkdown,
		p.extractYAMLBlock,
		p.fixIndentation,
	}

	for _, strategy := range strategies {
		if attempts >= opts.MaxAttempts {
			break
		}

		processed := strategy(output)
		if processed == "" {
			continue
		}

		var result interface{}
		err := yaml.Unmarshal([]byte(processed), &result)
		if err == nil {
			return result, nil
		}

		lastErr = err
		attempts++
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to parse YAML after %d recovery attempts: %w", attempts, lastErr)
	}

	return nil, fmt.Errorf("failed to parse YAML after %d recovery attempts", attempts)
}

// ParseWithSchema attempts to parse using schema guidance
func (p *YAMLParser) ParseWithSchema(ctx context.Context, output string, schema *OutputSchema) (interface{}, error) {
	result, err := p.ParseWithRecovery(ctx, output, &RecoveryOptions{
		ExtractFromMarkdown: true,
		FixCommonIssues:     true,
		MaxAttempts:         3,
		Schema:              schema,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML with schema guidance: %w", err)
	}

	return result, nil
}

// CanParse checks if the output might be YAML
func (p *YAMLParser) CanParse(output string) bool {
	trimmed := strings.TrimSpace(output)

	// Check for YAML indicators
	yamlPatterns := []string{
		`^---`,              // Document separator
		`^\w+:\s*`,          // Key-value pair at start
		`^\s*-\s+`,          // List item at start
		`^\w+:\s*\n\s+-`,    // Map with list
		`^\w+:\s*\n\s+\w+:`, // Nested maps
	}

	for _, pattern := range yamlPatterns {
		if matched, _ := regexp.MatchString(pattern, trimmed); matched {
			return true
		}
	}

	// Check for markdown code blocks with yaml
	if strings.Contains(output, "```yaml") || strings.Contains(output, "```yml") {
		return true
	}

	return false
}

// extractFromMarkdown extracts YAML from markdown code blocks
func (p *YAMLParser) extractFromMarkdown(output string) string {
	// Try to extract from ```yaml or ```yml blocks
	patterns := []string{
		"```ya?ml\n([^`]+)\n```",
		"```\n([^`]+)\n```",
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			content := strings.TrimSpace(matches[1])
			if p.looksLikeYAML(content) {
				return content
			}
		}
	}

	return ""
}

// cleanYAML performs basic YAML cleanup
func (p *YAMLParser) cleanYAML(output string) string {
	cleaned := strings.TrimSpace(output)

	// Remove BOM if present
	cleaned = strings.TrimPrefix(cleaned, "\xef\xbb\xbf")

	// Fix Windows line endings
	cleaned = strings.ReplaceAll(cleaned, "\r\n", "\n")

	// Replace tabs with spaces (YAML doesn't allow tabs for indentation)
	lines := strings.Split(cleaned, "\n")
	for i, line := range lines {
		// Replace all tabs with spaces (2 spaces per tab)
		lines[i] = strings.ReplaceAll(line, "\t", "  ")
		// Remove trailing spaces
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	cleaned = strings.Join(lines, "\n")

	return cleaned
}

// fixIndentation attempts to fix YAML indentation issues
func (p *YAMLParser) fixIndentation(output string) string {
	// First clean the YAML to convert tabs to spaces
	cleaned := p.cleanYAML(output)

	// Try to parse the cleaned YAML first
	var test interface{}
	if err := yaml.Unmarshal([]byte(cleaned), &test); err == nil {
		// If it parses successfully after cleaning, return it
		return cleaned
	}

	// If it still doesn't parse, try to fix indentation
	lines := strings.Split(cleaned, "\n")
	if len(lines) == 0 {
		return output
	}

	// Build a new YAML with proper indentation
	fixed := make([]string, 0, len(lines))
	indentStack := []int{0}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			fixed = append(fixed, "")
			continue
		}

		// Calculate current indent
		currentIndent := len(line) - len(strings.TrimLeft(line, " "))

		// Adjust indent stack based on current indentation
		for len(indentStack) > 1 && currentIndent <= indentStack[len(indentStack)-1] {
			indentStack = indentStack[:len(indentStack)-1]
		}

		// Determine the proper indentation level
		properIndent := indentStack[len(indentStack)-1]

		// If previous line ended with ':', increase indent for this line
		if i > 0 {
			prevTrimmed := strings.TrimSpace(lines[i-1])
			if strings.HasSuffix(prevTrimmed, ":") && !strings.HasPrefix(trimmed, "-") {
				properIndent = indentStack[len(indentStack)-1] + 2
				if currentIndent > indentStack[len(indentStack)-1] {
					indentStack = append(indentStack, properIndent)
				}
			}
		}

		// Apply the proper indentation
		fixed = append(fixed, strings.Repeat(" ", properIndent)+trimmed)
	}

	return strings.Join(fixed, "\n")
}

// extractYAMLBlock attempts to extract a YAML block from text
func (p *YAMLParser) extractYAMLBlock(output string) string {
	// Look for YAML document markers
	start := strings.Index(output, "---")
	if start == -1 {
		// Try to find the first line that looks like YAML
		lines := strings.Split(output, "\n")
		yamlStart := -1
		yamlEnd := len(lines)

		// Find start of YAML
		for i, line := range lines {
			if p.looksLikeYAMLLine(line) {
				yamlStart = i
				break
			}
		}

		if yamlStart == -1 {
			return ""
		}

		// Find end of YAML (when we hit non-YAML content)
		indentLevel := 0
		for i := yamlStart + 1; i < len(lines); i++ {
			line := lines[i]
			trimmed := strings.TrimSpace(line)

			if trimmed == "" {
				continue
			}

			// Calculate indentation
			currentIndent := len(line) - len(strings.TrimLeft(line, " "))

			// Check if this line looks like YAML continuation
			if strings.HasPrefix(trimmed, "-") || (currentIndent > 0 && currentIndent >= indentLevel) {
				// This is a list item or indented content
				if currentIndent > indentLevel {
					indentLevel = currentIndent
				}
				continue
			}

			// Check if this line looks like YAML
			if !p.looksLikeYAMLLine(trimmed) {
				yamlEnd = i
				break
			}

			// Update indent level for key-value pairs
			if strings.Contains(trimmed, ":") {
				indentLevel = currentIndent
			}
		}

		return strings.Join(lines[yamlStart:yamlEnd], "\n")
	}

	// Extract from --- to ... or end
	output = output[start+3:]
	end := strings.Index(output, "...")
	if end != -1 {
		output = output[:end]
	}

	return strings.TrimSpace(output)
}

// looksLikeYAML checks if content looks like YAML
func (p *YAMLParser) looksLikeYAML(content string) bool {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) == 0 {
		return false
	}

	// Check if first line looks like YAML
	return p.looksLikeYAMLLine(lines[0])
}

// looksLikeYAMLLine checks if a line looks like YAML
func (p *YAMLParser) looksLikeYAMLLine(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Check for YAML patterns
	patterns := []string{
		`^\w+:`,    // Key-value
		`^-\s+`,    // List item
		`^---$`,    // Document separator
		`^\.\.\.$`, // Document end
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, trimmed); matched {
			return true
		}
	}

	return false
}

// init registers the YAML parser
func init() {
	_ = RegisterParser(NewYAMLParser())
}
