package outputs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLParser_Parse(t *testing.T) {
	ctx := context.Background()
	parser := NewYAMLParser()

	testCases := []struct {
		name     string
		input    string
		expected interface{}
		wantErr  bool
	}{
		{
			name: "Valid YAML object",
			input: `name: test
value: 42
active: true`,
			expected: map[string]interface{}{
				"name":   "test",
				"value":  42,
				"active": true,
			},
		},
		{
			name: "Valid YAML array",
			input: `- apple
- banana
- orange`,
			expected: []interface{}{"apple", "banana", "orange"},
		},
		{
			name: "Nested YAML structure",
			input: `person:
  name: John
  age: 30
  contacts:
    - type: email
      value: john@example.com
    - type: phone
      value: "555-1234"`,
			expected: map[string]interface{}{
				"person": map[string]interface{}{
					"name": "John",
					"age":  30,
					"contacts": []interface{}{
						map[string]interface{}{
							"type":  "email",
							"value": "john@example.com",
						},
						map[string]interface{}{
							"type":  "phone",
							"value": "555-1234",
						},
					},
				},
			},
		},
		{
			name: "YAML with comments",
			input: `# This is a comment
name: test # inline comment
value: 42`,
			expected: map[string]interface{}{
				"name":  "test",
				"value": 42,
			},
		},
		{
			name:    "Invalid YAML",
			input:   `[invalid: yaml`,
			wantErr: true,
		},
		{
			name:     "Empty document",
			input:    ``,
			expected: nil,
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

func TestYAMLParser_ParseWithRecovery(t *testing.T) {
	ctx := context.Background()
	parser := NewYAMLParser()

	testCases := []struct {
		name     string
		input    string
		expected interface{}
		opts     *RecoveryOptions
	}{
		{
			name:  "YAML in markdown code block",
			input: "Here is the YAML configuration:\n```yaml\nserver:\n  host: localhost\n  port: 8080\n```",
			expected: map[string]interface{}{
				"server": map[string]interface{}{
					"host": "localhost",
					"port": 8080,
				},
			},
		},
		{
			name: "YAML with tab indentation",
			input: `name: test
	nested:
		value: 42`,
			expected: map[string]interface{}{
				"name": "test",
				"nested": map[string]interface{}{
					"value": 42,
				},
			},
		},
		{
			name: "Extract YAML from text",
			input: `The configuration is:
name: app
version: 1.0.0
features:
  - auth
  - logging
That's all.`,
			expected: map[string]interface{}{
				"name":    "app",
				"version": "1.0.0",
				"features": []interface{}{
					"auth",
					"logging",
				},
			},
		},
		{
			name: "YAML with missing quotes",
			input: `message: Hello world!
status: all systems go`,
			expected: map[string]interface{}{
				"message": "Hello world!",
				"status":  "all systems go",
			},
		},
		{
			name: "YAML document separator",
			input: `---
name: document1
---
name: document2`,
			expected: map[string]interface{}{
				"name": "document1",
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

func TestYAMLParser_CanParse(t *testing.T) {
	parser := NewYAMLParser()

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "YAML key-value",
			input:    "key: value",
			expected: true,
		},
		{
			name:     "YAML list",
			input:    "- item1\n- item2",
			expected: true,
		},
		{
			name:     "YAML in markdown",
			input:    "```yaml\nkey: value\n```",
			expected: true,
		},
		{
			name:     "YAML document separator",
			input:    "---\nkey: value",
			expected: true,
		},
		{
			name:     "JSON format",
			input:    `{"key": "value"}`,
			expected: false,
		},
		{
			name:     "XML format",
			input:    `<root><key>value</key></root>`,
			expected: false,
		},
		{
			name:     "Plain text",
			input:    "This is just plain text",
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

func TestYAMLParser_ExtractFromMarkdown(t *testing.T) {
	parser := NewYAMLParser()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "YAML code block",
			input: "```yaml\nkey: value\nlist:\n  - item\n```",
			expected: `key: value
list:
  - item`,
		},
		{
			name:  "Generic code block with YAML",
			input: "```\nname: test\nversion: 1.0\n```",
			expected: `name: test
version: 1.0`,
		},
		{
			name:     "No code block",
			input:    "Just text",
			expected: "",
		},
		{
			name:     "Code block without YAML",
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

func TestYAMLParser_ComplexStructures(t *testing.T) {
	ctx := context.Background()
	parser := NewYAMLParser()

	testCases := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name: "Anchors and aliases",
			input: `defaults: &defaults
  timeout: 30
  retries: 3

production:
  <<: *defaults
  server: prod.example.com`,
			expected: map[string]interface{}{
				"defaults": map[string]interface{}{
					"timeout": 30,
					"retries": 3,
				},
				"production": map[string]interface{}{
					"timeout": 30,
					"retries": 3,
					"server":  "prod.example.com",
				},
			},
		},
		{
			name: "Multi-line strings",
			input: `description: |
  This is a multi-line
  string that preserves
  line breaks.
summary: >
  This is a folded string
  that will be joined into
  a single line.`,
			expected: map[string]interface{}{
				"description": "This is a multi-line\nstring that preserves\nline breaks.\n",
				"summary":     "This is a folded string that will be joined into a single line.\n",
			},
		},
		{
			name: "Mixed types",
			input: `string: hello
number: 42
float: 3.14
boolean: true
null_value: null
date: 2023-01-01`,
			expected: map[string]interface{}{
				"string":     "hello",
				"number":     42,
				"float":      3.14,
				"boolean":    true,
				"null_value": nil,
				"date":       "2023-01-01",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
