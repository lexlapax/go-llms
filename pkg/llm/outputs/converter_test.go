package outputs

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConverter_Convert(t *testing.T) {
	ctx := context.Background()
	converter := NewConverter()

	testData := map[string]interface{}{
		"name":   "test",
		"value":  42,
		"active": true,
		"tags":   []interface{}{"go", "llm", "test"},
		"config": map[string]interface{}{
			"timeout": 30,
			"retries": 3,
		},
	}

	testCases := []struct {
		name     string
		data     interface{}
		from     Format
		to       Format
		validate func(t *testing.T, result interface{})
	}{
		{
			name: "JSON to YAML",
			data: testData,
			from: FormatJSON,
			to:   FormatYAML,
			validate: func(t *testing.T, result interface{}) {
				yamlStr, ok := result.(string)
				require.True(t, ok, "Result should be a string")

				var parsed map[string]interface{}
				err := yaml.Unmarshal([]byte(yamlStr), &parsed)
				require.NoError(t, err)

				assert.Equal(t, "test", parsed["name"])
				assert.Equal(t, 42, parsed["value"])
				assert.Equal(t, true, parsed["active"])
			},
		},
		{
			name: "JSON to XML",
			data: testData,
			from: FormatJSON,
			to:   FormatXML,
			validate: func(t *testing.T, result interface{}) {
				xmlStr, ok := result.(string)
				require.True(t, ok, "Result should be a string")

				assert.Contains(t, xmlStr, "<name>test</name>")
				assert.Contains(t, xmlStr, "<value>42</value>")
				assert.Contains(t, xmlStr, "<active>true</active>")
			},
		},
		{
			name: "YAML to JSON",
			data: `name: test
value: 42
active: true`,
			from: FormatYAML,
			to:   FormatJSON,
			validate: func(t *testing.T, result interface{}) {
				jsonStr, ok := result.(string)
				require.True(t, ok, "Result should be a string")

				var parsed map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &parsed)
				require.NoError(t, err)

				assert.Equal(t, "test", parsed["name"])
				assert.Equal(t, float64(42), parsed["value"])
				assert.Equal(t, true, parsed["active"])
			},
		},
		{
			name: "XML to JSON",
			data: `<root>
				<name>test</name>
				<value>42</value>
				<active>true</active>
			</root>`,
			from: FormatXML,
			to:   FormatJSON,
			validate: func(t *testing.T, result interface{}) {
				jsonStr, ok := result.(string)
				require.True(t, ok, "Result should be a string")

				var parsed map[string]interface{}
				err := json.Unmarshal([]byte(jsonStr), &parsed)
				require.NoError(t, err)

				// Debug output
				t.Logf("Parsed JSON: %+v", parsed)

				root, ok := parsed["root"].(map[string]interface{})
				require.True(t, ok, "Expected 'root' key in parsed JSON")
				assert.Equal(t, "test", root["name"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := converter.Convert(ctx, tc.data, tc.from, tc.to, nil)
			require.NoError(t, err)
			tc.validate(t, result)
		})
	}
}

func TestConverter_ConvertWithOptions(t *testing.T) {
	ctx := context.Background()
	converter := NewConverter()

	t.Run("Pretty print JSON", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "test",
			"nested": map[string]interface{}{
				"value": 42,
			},
		}

		opts := &ConversionOptions{
			Pretty:     true,
			IndentSize: 2,
		}

		result, err := converter.Convert(ctx, data, FormatJSON, FormatJSON, opts)
		require.NoError(t, err)

		jsonStr, ok := result.(string)
		require.True(t, ok)

		// Check for pretty printing
		assert.Contains(t, jsonStr, "\n")
		assert.Contains(t, jsonStr, "  ")
	})

	t.Run("XML with custom root", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "test",
		}

		opts := &ConversionOptions{
			RootElement: "custom",
		}

		result, err := converter.Convert(ctx, data, FormatJSON, FormatXML, opts)
		require.NoError(t, err)

		xmlStr, ok := result.(string)
		require.True(t, ok)
		assert.Contains(t, xmlStr, "<custom>")
		assert.Contains(t, xmlStr, "</custom>")
	})

	t.Run("Preserve types", func(t *testing.T) {
		data := map[string]interface{}{
			"string": "42",
			"number": 42,
			"float":  3.14,
		}

		opts := &ConversionOptions{
			PreserveTypes: true,
		}

		// Convert to YAML and back
		yamlResult, err := converter.Convert(ctx, data, FormatJSON, FormatYAML, opts)
		require.NoError(t, err)

		jsonResult, err := converter.Convert(ctx, yamlResult, FormatYAML, FormatJSON, opts)
		require.NoError(t, err)

		var parsed map[string]interface{}
		err = json.Unmarshal([]byte(jsonResult.(string)), &parsed)
		require.NoError(t, err)

		assert.Equal(t, "42", parsed["string"])
		assert.Equal(t, float64(42), parsed["number"])
		assert.Equal(t, 3.14, parsed["float"])
	})
}

func TestConverter_ArrayHandling(t *testing.T) {
	ctx := context.Background()
	converter := NewConverter()

	t.Run("Simple array", func(t *testing.T) {
		data := []interface{}{"apple", "banana", "orange"}

		result, err := converter.Convert(ctx, data, FormatJSON, FormatYAML, nil)
		require.NoError(t, err)

		yamlStr, ok := result.(string)
		require.True(t, ok)
		assert.Contains(t, yamlStr, "- apple")
		assert.Contains(t, yamlStr, "- banana")
		assert.Contains(t, yamlStr, "- orange")
	})

	t.Run("Array of objects", func(t *testing.T) {
		data := []interface{}{
			map[string]interface{}{"name": "Alice", "age": 30},
			map[string]interface{}{"name": "Bob", "age": 25},
		}

		result, err := converter.Convert(ctx, data, FormatJSON, FormatXML, nil)
		require.NoError(t, err)

		xmlStr, ok := result.(string)
		require.True(t, ok)
		assert.Contains(t, xmlStr, "<name>Alice</name>")
		assert.Contains(t, xmlStr, "<age>30</age>")
		assert.Contains(t, xmlStr, "<name>Bob</name>")
		assert.Contains(t, xmlStr, "<age>25</age>")
	})
}

func TestConverter_EdgeCases(t *testing.T) {
	ctx := context.Background()
	converter := NewConverter()

	t.Run("Empty data", func(t *testing.T) {
		result, err := converter.Convert(ctx, map[string]interface{}{}, FormatJSON, FormatYAML, nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Null values", func(t *testing.T) {
		data := map[string]interface{}{
			"key": nil,
		}

		result, err := converter.Convert(ctx, data, FormatJSON, FormatYAML, nil)
		require.NoError(t, err)

		yamlStr, ok := result.(string)
		require.True(t, ok)
		assert.Contains(t, yamlStr, "null")
	})

	t.Run("Special characters", func(t *testing.T) {
		data := map[string]interface{}{
			"message": "Hello & <world>",
		}

		result, err := converter.Convert(ctx, data, FormatJSON, FormatXML, nil)
		require.NoError(t, err)

		xmlStr, ok := result.(string)
		require.True(t, ok)
		assert.Contains(t, xmlStr, "&amp;")
		assert.Contains(t, xmlStr, "&lt;")
		assert.Contains(t, xmlStr, "&gt;")
	})

	t.Run("Unicode", func(t *testing.T) {
		data := map[string]interface{}{
			"emoji":   "🚀",
			"chinese": "你好",
		}

		// JSON to YAML
		result, err := converter.Convert(ctx, data, FormatJSON, FormatYAML, nil)
		require.NoError(t, err)

		yamlStr, ok := result.(string)
		require.True(t, ok)
		// YAML may encode emoji as Unicode escape sequence
		assert.True(t, strings.Contains(yamlStr, "🚀") || strings.Contains(yamlStr, "\\U0001F680"))
		assert.Contains(t, yamlStr, "你好")
	})
}

func TestConverter_InvalidConversions(t *testing.T) {
	ctx := context.Background()
	converter := NewConverter()

	t.Run("Invalid source format", func(t *testing.T) {
		_, err := converter.Convert(ctx, "test", Format("invalid"), FormatJSON, nil)
		assert.Error(t, err)
	})

	t.Run("Invalid target format", func(t *testing.T) {
		_, err := converter.Convert(ctx, "test", FormatJSON, Format("invalid"), nil)
		assert.Error(t, err)
	})

	t.Run("Malformed input", func(t *testing.T) {
		_, err := converter.Convert(ctx, "{invalid json", FormatJSON, FormatYAML, nil)
		assert.Error(t, err)
	})
}

func TestConverter_DetectFormat(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected Format
		wantErr  bool
	}{
		{
			name:     "JSON object",
			input:    `{"key": "value"}`,
			expected: FormatJSON,
		},
		{
			name:     "JSON array",
			input:    `["item1", "item2"]`,
			expected: FormatJSON,
		},
		{
			name:     "YAML",
			input:    "key: value\nlist:\n  - item",
			expected: FormatYAML,
		},
		{
			name:     "XML",
			input:    `<root><item>value</item></root>`,
			expected: FormatXML,
		},
		{
			name:    "Unknown format",
			input:   "This is just plain text",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			format, err := DetectFormat(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, format)
			}
		})
	}
}
