package processor

// ABOUTME: Optimized JSON extraction from LLM responses with markdown support
// ABOUTME: Uses tiered approach from fast simple extraction to complex parsing

import (
	"context"
	"regexp"
	"strings"

	"github.com/lexlapax/go-llms/pkg/util/json"
	"github.com/lexlapax/go-llms/pkg/util/profiling"
)

var (
	// Pre-compiled regex pattern for JSON extraction in markdown code blocks
	markdownCodeRegex = regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
)

// ExtractJSON is an optimized version of JSON extraction that handles various formats
// It uses a tiered approach, starting with fast methods and falling back to more complex ones
func ExtractJSON(s string) string {
	// Skip profiling if profiling is disabled (improves performance)
	if !profiling.IsProfilingEnabled() {
		return extractJSONImpl(s)
	}

	// Use profiling to measure performance with a background context
	result, _ := profiling.ProfileStructuredOp(context.Background(), profiling.OpStructuredExtraction, func(ctx context.Context) (interface{}, error) {
		return extractJSONImpl(s), nil
	})

	// Return the extracted JSON string
	return result.(string)
}

// extractJSONImpl implements the actual JSON extraction logic
// This was extracted from ExtractJSON to allow for profiling
func extractJSONImpl(s string) string {
	// extractJSONImpl extracts the first valid JSON object or array from a string.
	// It handles multiple extraction strategies in order of likelihood:
	// 1. Markdown code blocks (```json ... ```)
	// 2. Bracket matching for objects {}
	// 3. Bracket matching for arrays []
	// The function properly handles nested structures and escaped quotes.

	// Strategy 1: Check for markdown code blocks first (common in LLM responses)
	// Many LLMs wrap JSON in markdown code blocks for formatting
	if strings.Contains(s, "```") {
		if matches := markdownCodeRegex.FindStringSubmatch(s); len(matches) > 1 {
			potentialJSON := strings.TrimSpace(matches[1])
			// Verify it starts and ends correctly before expensive validation
			if (strings.HasPrefix(potentialJSON, "{") && strings.HasSuffix(potentialJSON, "}")) ||
				(strings.HasPrefix(potentialJSON, "[") && strings.HasSuffix(potentialJSON, "]")) {
				if json.Valid([]byte(potentialJSON)) {
					return potentialJSON
				}
			}
		}
	}

	// Strategy 2: Find JSON objects using bracket matching
	// This handles raw JSON mixed with text (e.g., "The result is {\"key\": \"value\"}")
	for i := 0; i < len(s); i++ {
		if s[i] == '{' {
			// Track nesting level and string context
			level := 0
			inString := false
			escaped := false

			// Scan forward to find matching closing brace
			for j := i; j < len(s); j++ {
				if !escaped {
					switch s[j] {
					case '\\':
						// Next character is escaped
						escaped = true
						continue
					case '"':
						// Toggle string context (ignore brackets in strings)
						inString = !inString
					case '{':
						if !inString {
							level++
						}
					case '}':
						if !inString {
							level--
							if level == 0 {
								// Found matching brace - validate the JSON
								candidate := s[i : j+1]
								if json.Valid([]byte(candidate)) {
									return candidate
								}
								// Invalid JSON, continue searching from next position
								break
							}
						}
					}
				}
				escaped = false
			}
		}
	}

	// Strategy 3: Find JSON arrays using bracket matching
	// Same logic as objects but for array structures
	for i := 0; i < len(s); i++ {
		if s[i] == '[' {
			// Track nesting level and string context
			level := 0
			inString := false
			escaped := false

			// Scan forward to find matching closing bracket
			for j := i; j < len(s); j++ {
				if !escaped {
					switch s[j] {
					case '\\':
						escaped = true
						continue
					case '"':
						inString = !inString
					case '[':
						if !inString {
							level++
						}
					case ']':
						if !inString {
							level--
							if level == 0 {
								// Found matching closing bracket - check if it's valid JSON
								candidate := s[i : j+1]
								if json.Valid([]byte(candidate)) {
									return candidate
								}
								// Invalid JSON, continue searching
								break
							}
						}
					}
				}
				escaped = false
			}
		}
	}

	// No valid JSON found
	return ""
}

// Removed unused manualExtractJSON and isBalanced functions
