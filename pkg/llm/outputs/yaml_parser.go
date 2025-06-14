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

	// Try different recovery strategies
	strategies := []func(string) string{
		func(s string) string { return s }, // Original
		p.extractFromMarkdown,
		p.cleanYAML,
		p.fixIndentation,
		p.extractYAMLBlock,
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
		`^---\s*$`,          // Document separator
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

	// Remove trailing spaces
	lines := strings.Split(cleaned, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	cleaned = strings.Join(lines, "\n")

	return cleaned
}

// fixIndentation attempts to fix YAML indentation issues
func (p *YAMLParser) fixIndentation(output string) string {
	lines := strings.Split(p.cleanYAML(output), "\n")
	if len(lines) == 0 {
		return output
	}

	// Find the base indentation
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// Remove base indentation
	if minIndent > 0 {
		for i, line := range lines {
			if len(line) >= minIndent {
				lines[i] = line[minIndent:]
			}
		}
	}

	return strings.Join(lines, "\n")
}

// extractYAMLBlock attempts to extract a YAML block from text
func (p *YAMLParser) extractYAMLBlock(output string) string {
	// Look for YAML document markers
	start := strings.Index(output, "---")
	if start == -1 {
		// Try to find the first line that looks like YAML
		lines := strings.Split(output, "\n")
		for i, line := range lines {
			if p.looksLikeYAMLLine(line) {
				return strings.Join(lines[i:], "\n")
			}
		}
		return ""
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
