package integration

import (
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/structured/processor"
	"github.com/stretchr/testify/assert"
)

func TestJSONExtractorEdgeCases(t *testing.T) {
	t.Run("MultipleJSONObjects", func(t *testing.T) {
		// Test extracting the first of multiple JSON objects
		input := `{"first": true} {"second": true} {"third": true}`
		result := processor.ExtractJSON(input)
		assert.JSONEq(t, `{"first": true}`, result)

		// The extractor should prefer the first valid JSON object
		input = `Invalid JSON here { not valid } {"valid": true} {"also": "valid"}`
		result = processor.ExtractJSON(input)
		assert.JSONEq(t, `{"valid": true}`, result)
	})

	t.Run("MalformedButRecoverableJSON", func(t *testing.T) {
		// Extra comma is technically invalid JSON but common in LLM responses
		input := `{"name": "John", "age": 30, }`
		result := processor.ExtractJSON(input)

		// Current behavior: this returns nothing because it's invalid JSON
		// If the processor is enhanced to fix common JSON issues, this test should be updated
		assert.Equal(t, "", result)

		// Single quotes instead of double quotes
		input = `{'name': 'John', 'age': 30}`
		result = processor.ExtractJSON(input)

		// Current behavior: this returns nothing because it's invalid JSON
		// If the processor is enhanced to handle single quotes, this test should be updated
		assert.Equal(t, "", result)
	})

	t.Run("NestedCodeBlocks", func(t *testing.T) {
		// Markdown inside code block with JSON
		input := "```\n" +
			"This is a code block with markdown and JSON:\n" +
			"```json\n" +
			`{"nested": true}` +
			"\n```\n" +
			"```"

		result := processor.ExtractJSON(input)
		assert.JSONEq(t, `{"nested": true}`, result)
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		// JSON with Unicode characters
		unicodeInput := `{"text": "Unicode: 你好, ñ, é, ß, 🚀"}`
		unicodeResult := processor.ExtractJSON(unicodeInput)
		assert.JSONEq(t, unicodeInput, unicodeResult)

		// Escaped characters
		escapedInput := `{"text": "Escaped: \"quotes\" and \\backslashes\\"}`
		escapedResult := processor.ExtractJSON(escapedInput)
		assert.JSONEq(t, escapedInput, escapedResult)

		// Escaped Unicode
		unicodeEscapedInput := `{"text": "\u00F1 \u0259"}`
		unicodeEscapedResult := processor.ExtractJSON(unicodeEscapedInput)
		assert.JSONEq(t, unicodeEscapedInput, unicodeEscapedResult)
	})

	t.Run("VeryLargeJSON", func(t *testing.T) {
		// Create a large JSON object
		var sb strings.Builder
		sb.WriteString(`{"items": [`)

		for i := 0; i < 1000; i++ {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(`{"id":`)
			sb.WriteString(string(rune('0' + i%10)))
			sb.WriteString(`,"value":"`)
			// Add a string with some content to increase size
			for j := 0; j < 20; j++ {
				sb.WriteString("item data ")
			}
			sb.WriteString(`"}`)
		}

		sb.WriteString(`]}`)
		largeJSON := sb.String()

		// Test that large JSON can be extracted
		result := processor.ExtractJSON(largeJSON)
		assert.Greater(t, len(result), 1000, "Large JSON should be extracted")

		// Wrap in text
		largeJSONInText := "Here is some large JSON:\n" + largeJSON + "\nEnd of JSON"
		result = processor.ExtractJSON(largeJSONInText)
		assert.Greater(t, len(result), 1000, "Large JSON should be extracted even when embedded in text")
	})

	t.Run("JSONFragments", func(t *testing.T) {
		// Test with key without value
		input1 := `{"key":}`
		result1 := processor.ExtractJSON(input1)
		assert.Equal(t, "", result1, "Invalid JSON key without value should not be extracted")

		// Test with proper key but missing comma
		input2 := `{"key1": true "key2": false}`
		result2 := processor.ExtractJSON(input2)
		assert.Equal(t, "", result2, "Invalid JSON with missing comma should not be extracted")

		// Test with JSON fragment but not a complete object
		input3 := `"key": "value"`
		result3 := processor.ExtractJSON(input3)
		assert.Equal(t, "", result3, "JSON fragment should not be extracted")
	})

	t.Run("EdgeWhitespace", func(t *testing.T) {
		// Test with whitespace at various positions
		input1 := `
		
		
		{"spaced": true}
		
		
		`
		result1 := processor.ExtractJSON(input1)
		assert.JSONEq(t, `{"spaced": true}`, result1)

		// Test with tabs and other whitespace characters
		input2 := "\t\t\t{\"tabbed\": true}\t\t\t"
		result2 := processor.ExtractJSON(input2)
		assert.JSONEq(t, `{"tabbed": true}`, result2)

		// Test with no whitespace
		input3 := `{"compact":true}`
		result3 := processor.ExtractJSON(input3)
		assert.JSONEq(t, `{"compact": true}`, result3)
	})

	t.Run("ObjectsWithArraysContainingObjects", func(t *testing.T) {
		// Complex nested structure
		input := `{
			"people": [
				{"name": "Alice", "details": {"age": 30, "active": true}},
				{"name": "Bob", "details": {"age": 25, "active": false}},
				{"name": "Charlie", "details": {"age": 35, "active": true}}
			],
			"metadata": {
				"count": 3,
				"tags": ["user", "profile"],
				"version": {
					"major": 1,
					"minor": 2
				}
			}
		}`

		result := processor.ExtractJSON(input)
		assert.NotEmpty(t, result, "Complex nested JSON should be extracted")

		// Verify the extracted JSON is valid and preserves structure
		var validJSON bool
		assert.NotPanics(t, func() {
			validJSON = assert.JSONEq(t, input, result)
		})
		assert.True(t, validJSON, "Extracted JSON should match original structure")
	})
}
