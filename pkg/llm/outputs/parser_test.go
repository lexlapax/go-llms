package outputs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserRegistry(t *testing.T) {
	t.Run("Register and Get Parser", func(t *testing.T) {
		registry := &ParserRegistry{
			parsers: make(map[string]Parser),
		}

		parser := NewJSONParser()
		err := registry.Register(parser)
		require.NoError(t, err)

		// Get parser
		retrieved, err := registry.Get("json")
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, "json", retrieved.Name())
	})

	t.Run("Register Nil Parser", func(t *testing.T) {
		registry := &ParserRegistry{
			parsers: make(map[string]Parser),
		}

		err := registry.Register(nil)
		assert.Error(t, err)
	})

	t.Run("Get Non-existent Parser", func(t *testing.T) {
		registry := &ParserRegistry{
			parsers: make(map[string]Parser),
		}

		_, err := registry.Get("nonexistent")
		assert.Error(t, err)
	})

	t.Run("Auto Detect Parser", func(t *testing.T) {
		registry := &ParserRegistry{
			parsers: make(map[string]Parser),
		}

		// Register parsers
		_ = registry.Register(NewJSONParser())
		_ = registry.Register(NewYAMLParser())
		_ = registry.Register(NewXMLParser())

		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{
				name:     "JSON object",
				input:    `{"key": "value"}`,
				expected: "json",
			},
			{
				name:     "JSON array",
				input:    `["item1", "item2"]`,
				expected: "json",
			},
			{
				name:     "YAML",
				input:    "key: value\nlist:\n  - item1\n  - item2",
				expected: "yaml",
			},
			{
				name:     "XML",
				input:    `<root><item>value</item></root>`,
				expected: "xml",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				parser, err := registry.AutoDetect(tc.input)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, parser.Name())
			})
		}
	})
}

func TestParseWithAutoDetection(t *testing.T) {
	ctx := context.Background()

	// Ensure parsers are registered
	_ = RegisterParser(NewJSONParser())
	_ = RegisterParser(NewYAMLParser())
	_ = RegisterParser(NewXMLParser())

	testCases := []struct {
		name     string
		input    string
		expected interface{}
		format   string
	}{
		{
			name:  "JSON object",
			input: `{"name": "test", "value": 42}`,
			expected: map[string]interface{}{
				"name":  "test",
				"value": float64(42),
			},
			format: "json",
		},
		{
			name:  "YAML",
			input: "name: test\nvalue: 42",
			expected: map[string]interface{}{
				"name":  "test",
				"value": 42,
			},
			format: "yaml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseWithAutoDetection(ctx, tc.input, nil)
			require.NoError(t, err)
			assert.Equal(t, tc.format, result.Format)
			assert.Equal(t, tc.expected, result.Data)
		})
	}
}

func TestRecoveryOptions(t *testing.T) {
	opts := DefaultRecoveryOptions()
	assert.True(t, opts.ExtractFromMarkdown)
	assert.True(t, opts.FixCommonIssues)
	assert.False(t, opts.StrictMode)
	assert.Equal(t, 3, opts.MaxAttempts)
}
