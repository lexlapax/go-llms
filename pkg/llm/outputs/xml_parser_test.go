package outputs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXMLParser_Parse(t *testing.T) {
	ctx := context.Background()
	parser := NewXMLParser()

	testCases := []struct {
		name     string
		input    string
		expected interface{}
		wantErr  bool
	}{
		{
			name: "Valid XML with single element",
			input: `<root>
				<name>test</name>
				<value>42</value>
				<active>true</active>
			</root>`,
			expected: map[string]interface{}{
				"root": map[string]interface{}{
					"name":   "test",
					"value":  "42",
					"active": "true",
				},
			},
		},
		{
			name: "XML with attributes",
			input: `<person id="123" role="admin">
				<name>John</name>
				<age>30</age>
			</person>`,
			expected: map[string]interface{}{
				"person": map[string]interface{}{
					"@id":   "123",
					"@role": "admin",
					"name":  "John",
					"age":   "30",
				},
			},
		},
		{
			name: "XML with nested elements",
			input: `<company>
				<name>Tech Corp</name>
				<employees>
					<employee>
						<name>Alice</name>
						<dept>Engineering</dept>
					</employee>
					<employee>
						<name>Bob</name>
						<dept>Sales</dept>
					</employee>
				</employees>
			</company>`,
			expected: map[string]interface{}{
				"company": map[string]interface{}{
					"name": "Tech Corp",
					"employees": map[string]interface{}{
						"employee": []interface{}{
							map[string]interface{}{
								"name": "Alice",
								"dept": "Engineering",
							},
							map[string]interface{}{
								"name": "Bob",
								"dept": "Sales",
							},
						},
					},
				},
			},
		},
		{
			name: "XML with declaration",
			input: `<?xml version="1.0" encoding="UTF-8"?>
			<message>Hello World</message>`,
			expected: map[string]interface{}{
				"message": "Hello World",
			},
		},
		{
			name:    "Invalid XML",
			input:   `<root><unclosed>`,
			wantErr: true,
		},
		{
			name:  "Empty element",
			input: `<root><empty/></root>`,
			expected: map[string]interface{}{
				"root": map[string]interface{}{
					"empty": map[string]interface{}{},
				},
			},
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

func TestXMLParser_ParseWithRecovery(t *testing.T) {
	ctx := context.Background()
	parser := NewXMLParser()

	testCases := []struct {
		name     string
		input    string
		expected interface{}
		opts     *RecoveryOptions
	}{
		{
			name:  "XML in markdown code block",
			input: "Here is the XML response:\n```xml\n<response>\n  <status>success</status>\n  <code>200</code>\n</response>\n```",
			expected: map[string]interface{}{
				"response": map[string]interface{}{
					"status": "success",
					"code":   "200",
				},
			},
		},
		{
			name: "XML with unclosed tags",
			input: `<root>
				<item>First item</item>
				<item>Second item
			</root>`,
			expected: map[string]interface{}{
				"root": map[string]interface{}{
					"item": []interface{}{
						"First item",
						"Second item",
					},
				},
			},
		},
		{
			name:  "Extract XML from text",
			input: `The response is: <result><success>true</success><count>10</count></result> and that's it.`,
			expected: map[string]interface{}{
				"result": map[string]interface{}{
					"success": "true",
					"count":   "10",
				},
			},
		},
		{
			name:  "XML with unquoted attributes",
			input: `<element id=123 name=test>content</element>`,
			expected: map[string]interface{}{
				"element": map[string]interface{}{
					"@id":     "123",
					"@name":   "test",
					"element": "content",
				},
			},
		},
		{
			name: "Multiple root elements wrapped",
			input: `<item>First</item>
			<item>Second</item>`,
			expected: map[string]interface{}{
				"root": map[string]interface{}{
					"item": []interface{}{"First", "Second"},
				},
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

func TestXMLParser_CanParse(t *testing.T) {
	parser := NewXMLParser()

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "XML declaration",
			input:    `<?xml version="1.0"?>`,
			expected: true,
		},
		{
			name:     "XML element",
			input:    `<root>content</root>`,
			expected: true,
		},
		{
			name:     "XML in markdown",
			input:    "```xml\n<root/>\n```",
			expected: true,
		},
		{
			name:     "HTML-like",
			input:    `<div>content</div>`,
			expected: true,
		},
		{
			name:     "JSON format",
			input:    `{"key": "value"}`,
			expected: false,
		},
		{
			name:     "YAML format",
			input:    "key: value",
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

func TestXMLParser_FixStrategies(t *testing.T) {
	parser := NewXMLParser()

	t.Run("Fix unclosed tags", func(t *testing.T) {
		input := `<root>
			<item>First</item>
			<item>Second
		</root>`

		fixed := parser.fixUnclosedTags(input)
		assert.Contains(t, fixed, "</item>")
	})

	t.Run("Fix attribute quotes", func(t *testing.T) {
		input := `<element id=123 name=test class='style'>`
		fixed := parser.fixAttributeQuotes(input)
		assert.Contains(t, fixed, `id="123"`)
		assert.Contains(t, fixed, `name="test"`)
	})

	t.Run("Extract XML block", func(t *testing.T) {
		input := `Some text before <root>content</root> some text after`
		extracted := parser.extractXMLBlock(input)
		assert.Equal(t, `<root>content</root>`, extracted)
	})

	t.Run("Wrap in root", func(t *testing.T) {
		input := `<item>1</item>
		<item>2</item>`
		wrapped := parser.wrapInRoot(input)
		assert.Contains(t, wrapped, "<root>")
		assert.Contains(t, wrapped, "</root>")
	})
}

func TestXMLParser_ComplexStructures(t *testing.T) {
	ctx := context.Background()
	parser := NewXMLParser()

	testCases := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name: "CDATA sections",
			input: `<article>
				<title>Test Article</title>
				<content><![CDATA[This contains <special> characters & symbols]]></content>
			</article>`,
			expected: map[string]interface{}{
				"article": map[string]interface{}{
					"title":   "Test Article",
					"content": "This contains <special> characters & symbols",
				},
			},
		},
		{
			name: "Namespaces",
			input: `<root xmlns:custom="http://example.com">
				<custom:element>Value</custom:element>
			</root>`,
			expected: map[string]interface{}{
				"root": map[string]interface{}{
					"@xmlns:custom": "http://example.com",
					"element":       "Value",
				},
			},
		},
		{
			name: "Mixed content",
			input: `<paragraph>
				This is <bold>bold</bold> and this is <italic>italic</italic> text.
			</paragraph>`,
			expected: map[string]interface{}{
				"paragraph": map[string]interface{}{
					"bold":   "bold",
					"italic": "italic",
				},
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
