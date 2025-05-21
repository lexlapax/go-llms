package processor

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
	// Check for markdown code blocks first (common in LLM responses)
	if strings.Contains(s, "```") {
		if matches := markdownCodeRegex.FindStringSubmatch(s); len(matches) > 1 {
			potentialJSON := strings.TrimSpace(matches[1])
			// Verify it starts and ends correctly and is valid JSON
			if (strings.HasPrefix(potentialJSON, "{") && strings.HasSuffix(potentialJSON, "}")) ||
				(strings.HasPrefix(potentialJSON, "[") && strings.HasSuffix(potentialJSON, "]")) {
				if json.Valid([]byte(potentialJSON)) {
					return potentialJSON
				}
			}
		}
	}

	// Fast path: Try to find complete JSON objects in the string
	// This handles the case of multiple JSON objects in the same string
	for i := 0; i < len(s); i++ {
		if s[i] == '{' {
			// Try to find matching closing brace
			level := 0
			inString := false
			escaped := false

			for j := i; j < len(s); j++ {
				if !escaped {
					switch s[j] {
					case '\\':
						escaped = true
						continue
					case '"':
						inString = !inString
					case '{':
						if !inString {
							level++
						}
					case '}':
						if !inString {
							level--
							if level == 0 {
								// Found matching closing brace - check if it's valid JSON
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

	// Fast path: Try to find complete JSON arrays in the string
	for i := 0; i < len(s); i++ {
		if s[i] == '[' {
			// Try to find matching closing bracket
			level := 0
			inString := false
			escaped := false

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
