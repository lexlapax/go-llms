// ABOUTME: JSON parser implementation with recovery capabilities for malformed outputs
// ABOUTME: Handles common LLM output issues like markdown blocks, trailing commas, etc.

package outputs

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// JSONParser implements Parser for JSON format
type JSONParser struct {
	// StrictMode disables all recovery attempts
	strictMode bool
}

// NewJSONParser creates a new JSON parser
func NewJSONParser() *JSONParser {
	return &JSONParser{
		strictMode: false,
	}
}

// NewStrictJSONParser creates a JSON parser with strict mode
func NewStrictJSONParser() *JSONParser {
	return &JSONParser{
		strictMode: true,
	}
}

// Name returns the parser name
func (p *JSONParser) Name() string {
	return "json"
}

// Parse attempts to parse JSON output
func (p *JSONParser) Parse(ctx context.Context, output string) (interface{}, error) {
	// Try direct parsing first
	var result interface{}
	err := json.Unmarshal([]byte(output), &result)
	if err == nil {
		return result, nil
	}

	// If strict mode, return error immediately
	if p.strictMode {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Try basic recovery
	cleaned := p.cleanJSON(output)
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON after cleanup: %w", err)
	}

	return result, nil
}

// ParseWithRecovery attempts to parse with advanced recovery options
func (p *JSONParser) ParseWithRecovery(ctx context.Context, output string, opts *RecoveryOptions) (interface{}, error) {
	if opts == nil {
		opts = DefaultRecoveryOptions()
	}

	// If strict mode is set, use basic parse
	if opts.StrictMode || p.strictMode {
		return p.Parse(ctx, output)
	}

	attempts := 0
	var lastErr error

	// Try different recovery strategies
	strategies := []func(string) string{
		func(s string) string { return s }, // Original
		p.extractFromMarkdown,
		p.cleanJSON,
		p.fixCommonIssues,
		p.extractJSONObject,
		p.extractJSONArray,
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
		err := json.Unmarshal([]byte(processed), &result)
		if err == nil {
			return result, nil
		}

		lastErr = err
		attempts++
	}

	// If we have a schema, try schema-guided extraction
	if opts.Schema != nil {
		result, err := p.ParseWithSchema(ctx, output, opts.Schema)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to parse JSON after %d recovery attempts: %w", attempts, lastErr)
	}

	return nil, fmt.Errorf("failed to parse JSON after %d recovery attempts", attempts)
}

// ParseWithSchema attempts to parse using schema guidance
func (p *JSONParser) ParseWithSchema(ctx context.Context, output string, schema *OutputSchema) (interface{}, error) {
	// First try normal parsing
	result, err := p.ParseWithRecovery(ctx, output, &RecoveryOptions{
		ExtractFromMarkdown: true,
		FixCommonIssues:     true,
		MaxAttempts:         3,
	})
	if err == nil {
		// Validate against schema
		if err := validateAgainstSchema(result, schema); err == nil {
			return result, nil
		}
	}

	// Try to extract based on schema structure
	extracted := p.extractBySchema(output, schema)
	if extracted != nil {
		return extracted, nil
	}

	return nil, fmt.Errorf("failed to parse JSON with schema guidance")
}

// CanParse checks if the output might be JSON
func (p *JSONParser) CanParse(output string) bool {
	trimmed := strings.TrimSpace(output)

	// Check for JSON-like structures
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		return true
	}
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		return true
	}

	// Check for markdown code blocks with json
	if strings.Contains(output, "```json") {
		return true
	}

	// Check for JSON object or array anywhere in the text
	if strings.Contains(output, "{") && strings.Contains(output, "}") {
		// Extract potential JSON and check if it looks valid
		jsonObj := p.extractJSONObject(output)
		if jsonObj != "" && p.looksLikeJSON(jsonObj) {
			return true
		}
	}

	// Check for common JSON patterns
	jsonPatterns := []string{
		`"[^"]+"\s*:\s*`,  // Key-value pairs
		`{\s*"[^"]+"\s*:`, // Object start
		`\[\s*{`,          // Array of objects
	}

	for _, pattern := range jsonPatterns {
		if matched, _ := regexp.MatchString(pattern, output); matched {
			return true
		}
	}

	return false
}

// extractFromMarkdown extracts JSON from markdown code blocks
func (p *JSONParser) extractFromMarkdown(output string) string {
	// Try to extract from ```json blocks
	re := regexp.MustCompile("```json\n([^`]+)\n```")
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try generic code blocks
	re = regexp.MustCompile("```\n([^`]+)\n```")
	matches = re.FindStringSubmatch(output)
	if len(matches) > 1 {
		content := strings.TrimSpace(matches[1])
		if p.looksLikeJSON(content) {
			return content
		}
	}

	return ""
}

// cleanJSON performs basic JSON cleanup
func (p *JSONParser) cleanJSON(output string) string {
	cleaned := strings.TrimSpace(output)

	// Remove BOM if present
	cleaned = strings.TrimPrefix(cleaned, "\xef\xbb\xbf")

	// Remove control characters
	cleaned = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(cleaned, "")

	return cleaned
}

// fixCommonIssues fixes common JSON formatting issues
func (p *JSONParser) fixCommonIssues(output string) string {
	fixed := p.cleanJSON(output)

	// First extract JSON object/array if embedded in text
	if !p.looksLikeJSON(fixed) {
		// Try to extract JSON object
		if jsonObj := p.extractJSONObject(fixed); jsonObj != "" {
			fixed = jsonObj
		} else if jsonArr := p.extractJSONArray(fixed); jsonArr != "" {
			fixed = jsonArr
		}
	}

	// Fix trailing commas in objects
	fixed = regexp.MustCompile(`,(\s*})`).ReplaceAllString(fixed, "$1")

	// Fix trailing commas in arrays
	fixed = regexp.MustCompile(`,(\s*])`).ReplaceAllString(fixed, "$1")

	// Fix single quotes (convert to double quotes)
	// This is a simple approach and might need refinement for edge cases
	fixed = p.fixQuotes(fixed)

	// Fix missing quotes around keys - improved regex to handle unquoted keys
	// Match word boundaries to avoid matching inside strings
	fixed = regexp.MustCompile(`([{,]\s*)([a-zA-Z_][a-zA-Z0-9_]*)(\s*:)`).ReplaceAllString(fixed, `$1"$2"$3`)

	// Fix decimal numbers that might be invalid
	fixed = regexp.MustCompile(`:\s*\.(\d+)`).ReplaceAllString(fixed, `:0.$1`)

	return fixed
}

// fixQuotes attempts to fix quote issues
func (p *JSONParser) fixQuotes(input string) string {
	// This is a simplified approach
	// In production, you'd want a more sophisticated quote fixing algorithm

	// Skip if it looks like valid JSON already
	var test interface{}
	if err := json.Unmarshal([]byte(input), &test); err == nil {
		return input
	}

	// Try replacing single quotes with double quotes
	// This is naive and might break on apostrophes in strings
	if strings.Contains(input, "'") {
		fixed := strings.ReplaceAll(input, "'", "\"")
		if err := json.Unmarshal([]byte(fixed), &test); err == nil {
			return fixed
		}
	}

	return input
}

// extractJSONObject attempts to extract a JSON object from text
func (p *JSONParser) extractJSONObject(output string) string {
	// Find the first { and last }
	start := strings.Index(output, "{")
	if start == -1 {
		return ""
	}

	end := strings.LastIndex(output, "}")
	if end == -1 || end <= start {
		return ""
	}

	return output[start : end+1]
}

// extractJSONArray attempts to extract a JSON array from text
func (p *JSONParser) extractJSONArray(output string) string {
	// Find the first [ and last ]
	start := strings.Index(output, "[")
	if start == -1 {
		return ""
	}

	end := strings.LastIndex(output, "]")
	if end == -1 || end <= start {
		return ""
	}

	return output[start : end+1]
}

// looksLikeJSON performs a quick check if content looks like JSON
func (p *JSONParser) looksLikeJSON(content string) bool {
	trimmed := strings.TrimSpace(content)
	return (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))
}

// extractBySchema attempts to extract JSON based on schema structure
func (p *JSONParser) extractBySchema(output string, schema *OutputSchema) interface{} {
	// This is a placeholder for schema-guided extraction
	// In a full implementation, this would analyze the schema
	// and try to extract matching structures from the output

	// For now, return nil to indicate extraction failed
	return nil
}

// validateAgainstSchema validates parsed data against a schema
func validateAgainstSchema(data interface{}, schema *OutputSchema) error {
	// This would integrate with the schema validation system
	// For now, we'll assume validation passes
	return nil
}

// init registers the JSON parser
func init() {
	_ = RegisterParser(NewJSONParser())
}
