package outputs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONParser_Parse(t *testing.T) {
	ctx := context.Background()
	parser := NewJSONParser()

	testCases := []struct {
		name     string
		input    string
		expected interface{}
		wantErr  bool
	}{
		{
			name:  "Valid JSON object",
			input: `{"name": "test", "value": 42, "active": true}`,
			expected: map[string]interface{}{
				"name":   "test",
				"value":  float64(42),
				"active": true,
			},
		},
		{
			name:  "Valid JSON array",
			input: `[1, 2, 3, "test"]`,
			expected: []interface{}{
				float64(1), float64(2), float64(3), "test",
			},
		},
		{
			name:  "JSON with whitespace",
			input: `  { "key" : "value" }  `,
			expected: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name:    "Invalid JSON",
			input:   `{invalid json}`,
			wantErr: true,
		},
		{
			name:     "Empty object",
			input:    `{}`,
			expected: map[string]interface{}{},
		},
		{
			name:     "Empty array",
			input:    `[]`,
			expected: []interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestJSONParser_ParseWithRecovery(t *testing.T) {
	ctx := context.Background()
	parser := NewJSONParser()

	testCases := []struct {
		name     string
		input    string
		expected interface{}
		opts     *RecoveryOptions
	}{
		{
			name:  "JSON in markdown code block",
			input: "Here is the JSON response:\n```json\n{\n  \"name\": \"test\",\n  \"value\": 42\n}\n```",
			expected: map[string]interface{}{
				"name":  "test",
				"value": float64(42),
			},
		},
		{
			name: "JSON with trailing comma",
			input: `{
				"name": "test",
				"value": 42,
			}`,
			expected: map[string]interface{}{
				"name":  "test",
				"value": float64(42),
			},
		},
		{
			name:  "JSON with single quotes",
			input: `{'name': 'test', 'value': 42}`,
			expected: map[string]interface{}{
				"name":  "test",
				"value": float64(42),
			},
		},
		{
			name:  "JSON with unquoted keys",
			input: `{name: "test", value: 42}`,
			expected: map[string]interface{}{
				"name":  "test",
				"value": float64(42),
			},
		},
		{
			name:  "Extract JSON object from text",
			input: `The response is: {"status": "success", "count": 10} and that's it.`,
			expected: map[string]interface{}{
				"status": "success",
				"count":  float64(10),
			},
		},
		{
			name:  "Extract JSON array from text",
			input: `Results: ["apple", "banana", "orange"] are the fruits.`,
			expected: []interface{}{
				"apple", "banana", "orange",
			},
		},
		{
			name:  "Fix decimal without leading zero",
			input: `{"price": .99, "discount": .1}`,
			expected: map[string]interface{}{
				"price":    float64(0.99),
				"discount": float64(0.1),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := tc.opts
			if opts == nil {
				opts = DefaultRecoveryOptions()
			}

			result, err := parser.ParseWithRecovery(ctx, tc.input, opts)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestJSONParser_CanParse(t *testing.T) {
	parser := NewJSONParser()

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "JSON object",
			input:    `{"key": "value"}`,
			expected: true,
		},
		{
			name:     "JSON array",
			input:    `[1, 2, 3]`,
			expected: true,
		},
		{
			name:     "JSON in markdown",
			input:    "```json\n{}\n```",
			expected: true,
		},
		{
			name:     "Contains JSON-like pattern",
			input:    `The result has "key": "value" in it`,
			expected: true,
		},
		{
			name:     "Plain text",
			input:    "This is just plain text",
			expected: false,
		},
		{
			name:     "YAML format",
			input:    "key: value\nlist:\n  - item",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser.CanParse(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestJSONParser_StrictMode(t *testing.T) {
	ctx := context.Background()
	parser := NewStrictJSONParser()

	// Valid JSON should work
	result, err := parser.Parse(ctx, `{"valid": true}`)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"valid": true}, result)

	// Invalid JSON should fail immediately
	_, err = parser.Parse(ctx, `{invalid json}`)
	assert.Error(t, err)

	// Recovery options with strict mode should not attempt recovery
	opts := &RecoveryOptions{StrictMode: true}
	_, err = parser.ParseWithRecovery(ctx, `{invalid json}`, opts)
	assert.Error(t, err)
}

func TestJSONParser_ExtractFromMarkdown(t *testing.T) {
	parser := NewJSONParser()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSON code block",
			input:    "```json\n{\"test\": true}\n```",
			expected: `{"test": true}`,
		},
		{
			name:     "Generic code block with JSON",
			input:    "```\n[1, 2, 3]\n```",
			expected: `[1, 2, 3]`,
		},
		{
			name:     "No code block",
			input:    "Just text",
			expected: "",
		},
		{
			name:     "Code block without JSON",
			input:    "```\nprint('hello')\n```",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parser.extractFromMarkdown(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
