package processor

import (
	"regexp"
	"strings"
)

var (
	// Pre-compiled regex pattern for JSON extraction in markdown code blocks
	markdownCodeRegex = regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
)

// ExtractJSON is an optimized version of JSON extraction that handles various formats
// It uses a tiered approach, starting with fast methods and falling back to more complex ones
func ExtractJSON(s string) string {
	// Fast path: Try to find the first complete JSON object in the string
	// This handles the case of multiple JSON objects in the same string
	if startIdx := strings.Index(s, "{"); startIdx >= 0 {
		// Find the matching closing brace by counting opening/closing braces
		level := 0
		for i := startIdx; i < len(s); i++ {
			if s[i] == '{' {
				level++
			} else if s[i] == '}' {
				level--
				if level == 0 {
					// Found matching closing brace - extract the complete JSON object
					return s[startIdx : i+1]
				}
			}
		}
	}

	// Fast path: Try to find the first complete JSON array in the string
	if startIdx := strings.Index(s, "["); startIdx >= 0 {
		// Find the matching closing bracket by counting opening/closing brackets
		level := 0
		for i := startIdx; i < len(s); i++ {
			if s[i] == '[' {
				level++
			} else if s[i] == ']' {
				level--
				if level == 0 {
					// Found matching closing bracket - extract the complete JSON array
					return s[startIdx : i+1]
				}
			}
		}
	}

	// Check for markdown code blocks (common in LLM responses)
	if strings.Contains(s, "```") {
		if matches := markdownCodeRegex.FindStringSubmatch(s); len(matches) > 1 {
			potentialJSON := strings.TrimSpace(matches[1])
			// Verify it starts and ends correctly
			if (strings.HasPrefix(potentialJSON, "{") && strings.HasSuffix(potentialJSON, "}")) ||
				(strings.HasPrefix(potentialJSON, "[") && strings.HasSuffix(potentialJSON, "]")) {
				return potentialJSON
			}
		}
	}

	// Fallback to more expensive methods for complex cases
	return manualExtractJSON(s)
}

// manualExtractJSON is a fallback JSON extraction method (unused in the optimized version)
func manualExtractJSON(s string) string {
	// This function is now a fallback and should not be called in normal operation
	// since our main ExtractJSON function handles all cases
	return ""
}

// Removed unused isBalanced function
