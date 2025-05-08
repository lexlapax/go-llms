package processor

import (
	"testing"
)

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isEmpty  bool
	}{
		{
			name:     "simple_json_object",
			input:    `{"name": "John", "age": 30}`,
			expected: `{"name": "John", "age": 30}`,
			isEmpty:  false,
		},
		{
			name:     "json_in_text",
			input:    `Here is some JSON: {"name": "John", "age": 30} and more text`,
			expected: `{"name": "John", "age": 30}`,
			isEmpty:  false,
		},
		{
			name:     "json_array",
			input:    `[1, 2, 3, 4, 5]`,
			expected: `[1, 2, 3, 4, 5]`,
			isEmpty:  false,
		},
		{
			name:     "json_array_in_text",
			input:    `Here is an array: [1, 2, 3, 4, 5] and more text`,
			expected: `[1, 2, 3, 4, 5]`,
			isEmpty:  false,
		},
		{
			name: "json_in_markdown_code_block",
			input: "```json\n" +
				`{"name": "John", "age": 30}` +
				"\n```",
			expected: `{"name": "John", "age": 30}`,
			isEmpty:  false,
		},
		{
			name: "json_in_generic_code_block",
			input: "```\n" +
				`{"name": "John", "age": 30}` +
				"\n```",
			expected: `{"name": "John", "age": 30}`,
			isEmpty:  false,
		},
		{
			name: "nested_json",
			input: `{
				"person": {
					"name": "John",
					"address": {
						"city": "New York",
						"zip": "10001"
					}
				},
				"orders": [
					{
						"id": 1,
						"items": ["apple", "orange"]
					}
				]
			}`,
			expected: `{
				"person": {
					"name": "John",
					"address": {
						"city": "New York",
						"zip": "10001"
					}
				},
				"orders": [
					{
						"id": 1,
						"items": ["apple", "orange"]
					}
				]
			}`,
			isEmpty: false,
		},
		{
			name:     "no_json",
			input:    "This text contains no JSON",
			expected: "",
			isEmpty:  true,
		},
		{
			name:     "unbalanced_braces",
			input:    `{"name": "John", "age": 30`,
			expected: "",
			isEmpty:  true,
		},
		{
			name:     "partial_json_extraction",
			input:    `Before {"name": "John"} After {"age": 30}`,
			expected: `{"name": "John"}`,
			isEmpty:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractJSON(tt.input)

			if tt.isEmpty && result != "" {
				t.Errorf("Expected empty result, got: %s", result)
			}

			if !tt.isEmpty && result == "" {
				t.Errorf("Expected non-empty result, got empty string")
			}

			if !tt.isEmpty && result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
