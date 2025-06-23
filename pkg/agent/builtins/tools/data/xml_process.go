// ABOUTME: XMLProcess tool provides XML parsing and querying capabilities
// ABOUTME: This tool enables agents to work with XML data without requiring LLM processing

package data

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// XMLProcessInput represents the input for XML processing operations
type XMLProcessInput struct {
	// The XML data to process (as a string)
	Data string `json:"data" jsonschema:"title=XML Data,description=The XML data to process,required"`

	// Operation to perform: parse, query, to_json
	Operation string `json:"operation" jsonschema:"title=Operation,description=Operation to perform: parse query to_json,enum=parse,enum=query,enum=to_json,required"`

	// XPath-like query for query operation (simplified)
	XPath string `json:"xpath,omitempty" jsonschema:"title=XPath,description=Simplified XPath query for query operation"`

	// Whether to include attributes in the result
	IncludeAttributes bool `json:"include_attributes" jsonschema:"title=Include Attributes,description=Whether to include XML attributes in the result,default=true"`
}

// XMLProcessOutput represents the output of XML processing
type XMLProcessOutput struct {
	// The processed result
	Result interface{} `json:"result"`

	// Error message if any
	Error string `json:"error,omitempty"`

	// Root element name
	RootElement string `json:"root_element,omitempty"`
}

// xmlProcessParamSchema defines parameters for the XMLProcess tool
var xmlProcessParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"data": {
			Type:        "string",
			Description: "The XML data to process",
		},
		"operation": {
			Type:        "string",
			Description: "Operation to perform: parse, query, or to_json",
			Enum:        []string{"parse", "query", "to_json"},
		},
		"xpath": {
			Type:        "string",
			Description: "Simplified XPath query for query operation",
		},
		"include_attributes": {
			Type:        "boolean",
			Description: "Whether to include XML attributes in the result",
		},
	},
	Required: []string{"data", "operation"},
}

// XMLProcess creates a tool for processing XML data
// This tool provides XML manipulation capabilities including parsing,
// querying with simplified XPath expressions, and converting to JSON.
// It handles attributes and nested elements.
func XMLProcess() domain.Tool {
	// Create output schema for XMLProcessOutput
	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"result": {
				Type:        "any",
				Description: "The processed result (can be parsed XML structure, query results, or JSON string)",
			},
			"error": {
				Type:        "string",
				Description: "Error message if any",
			},
			"root_element": {
				Type:        "string",
				Description: "Name of the root XML element",
			},
		},
		Required: []string{},
	}

	builder := atools.NewToolBuilder("xml_process", "Process XML data: parse, query with simplified XPath, or convert to JSON").
		WithFunction(xmlProcessExecute).
		WithParameterSchema(xmlProcessParamSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`Use this tool to process XML data in various ways:

Parse Operation:
- Validates XML syntax and parses the data
- Returns a structured representation of the XML
- Preserves element hierarchy and relationships
- Optionally includes attributes (controlled by include_attributes)

Query Operation (Simplified XPath):
- Extract specific elements using path expressions
- Supports basic XPath-like syntax:
  - /: Path separator
  - element: Select elements by name
  - *: Select all child elements
  - @attribute: Select attribute values
  - element/child: Navigate hierarchy
- Returns matching elements or attribute values

Convert to JSON:
- Transforms XML structure to JSON format
- Preserves element names, attributes, and text content
- Useful for working with XML data in JSON-based systems

XML Structure Representation:
- _name: Element name
- _attributes: Element attributes (if include_attributes is true)
- _text: Text content (for leaf elements)
- Child elements are added as properties

Namespace Support:
- Basic namespace handling (local names are used)
- For complex namespace scenarios, consider parsing and processing manually

State Integration:
- xml_include_attributes_default: Default value for include_attributes option`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Parse simple XML",
				Description: "Parse a basic XML document with attributes",
				Scenario:    "When you need to understand XML structure",
				Input: map[string]interface{}{
					"data": `<book id="123">
  <title>Go Programming</title>
  <author>John Doe</author>
  <year>2024</year>
</book>`,
					"operation":          "parse",
					"include_attributes": true,
				},
				Output: map[string]interface{}{
					"result": map[string]interface{}{
						"_name": "book",
						"_attributes": map[string]string{
							"id": "123",
						},
						"title": map[string]interface{}{
							"_name": "title",
							"_text": "Go Programming",
						},
						"author": map[string]interface{}{
							"_name": "author",
							"_text": "John Doe",
						},
						"year": map[string]interface{}{
							"_name": "year",
							"_text": "2024",
						},
					},
					"root_element": "book",
				},
				Explanation: "XML is parsed into a structured format with element names, attributes, and text content",
			},
			{
				Name:        "Query specific element",
				Description: "Extract book title using XPath-like query",
				Scenario:    "When you need to extract specific data from XML",
				Input: map[string]interface{}{
					"data": `<catalog>
  <book>
    <title>XML Processing</title>
    <price>29.99</price>
  </book>
  <book>
    <title>Data Formats</title>
    <price>34.99</price>
  </book>
</catalog>`,
					"operation": "query",
					"xpath":     "book/title",
				},
				Output: map[string]interface{}{
					"result": []interface{}{
						map[string]interface{}{
							"_name": "title",
							"_text": "XML Processing",
						},
						map[string]interface{}{
							"_name": "title",
							"_text": "Data Formats",
						},
					},
					"root_element": "catalog",
				},
				Explanation: "XPath query returns all matching elements from the document",
			},
			{
				Name:        "Query attribute value",
				Description: "Extract attribute values using @ notation",
				Scenario:    "When you need attribute values from XML elements",
				Input: map[string]interface{}{
					"data": `<users>
  <user id="u1" role="admin">Alice</user>
  <user id="u2" role="user">Bob</user>
</users>`,
					"operation": "query",
					"xpath":     "user/@id",
				},
				Output: map[string]interface{}{
					"result":       []string{"u1", "u2"},
					"root_element": "users",
				},
				Explanation: "Attribute queries return arrays of attribute values",
			},
			{
				Name:        "Convert XML to JSON",
				Description: "Transform XML data into JSON format",
				Scenario:    "When you need to work with XML data as JSON",
				Input: map[string]interface{}{
					"data": `<person>
  <name>Jane Smith</name>
  <email>jane@example.com</email>
  <skills>
    <skill>Python</skill>
    <skill>XML</skill>
    <skill>JSON</skill>
  </skills>
</person>`,
					"operation":          "to_json",
					"include_attributes": false,
				},
				Output: map[string]interface{}{
					"result": `{
  "_name": "person",
  "name": {
    "_name": "name",
    "_text": "Jane Smith"
  },
  "email": {
    "_name": "email",
    "_text": "jane@example.com"
  },
  "skills": {
    "_name": "skills",
    "skill": [
      {
        "_name": "skill",
        "_text": "Python"
      },
      {
        "_name": "skill",
        "_text": "XML"
      },
      {
        "_name": "skill",
        "_text": "JSON"
      }
    ]
  }
}`,
					"root_element": "person",
				},
				Explanation: "XML structure is preserved in JSON format with special keys for metadata",
			},
			{
				Name:        "Handle namespaced XML",
				Description: "Parse XML with namespace declarations",
				Scenario:    "When working with XML that uses namespaces",
				Input: map[string]interface{}{
					"data": `<ns:root xmlns:ns="http://example.com/ns">
  <ns:item>Namespaced content</ns:item>
</ns:root>`,
					"operation": "parse",
				},
				Output: map[string]interface{}{
					"result": map[string]interface{}{
						"_name": "root",
						"item": map[string]interface{}{
							"_name": "item",
							"_text": "Namespaced content",
						},
					},
					"root_element": "root",
				},
				Explanation: "Namespace prefixes are stripped, using local element names",
			},
			{
				Name:        "Query all children with wildcard",
				Description: "Use * to select all child elements",
				Scenario:    "When you need all children of an element",
				Input: map[string]interface{}{
					"data": `<config>
  <database>MySQL</database>
  <cache>Redis</cache>
  <queue>RabbitMQ</queue>
</config>`,
					"operation": "query",
					"xpath":     "*",
				},
				Output: map[string]interface{}{
					"result": []interface{}{
						map[string]interface{}{
							"_name": "database",
							"_text": "MySQL",
						},
						map[string]interface{}{
							"_name": "cache",
							"_text": "Redis",
						},
						map[string]interface{}{
							"_name": "queue",
							"_text": "RabbitMQ",
						},
					},
					"root_element": "config",
				},
				Explanation: "Wildcard * selects all child elements at the current level",
			},
			{
				Name:        "Handle invalid XML gracefully",
				Description: "Error handling for malformed XML",
				Scenario:    "When processing potentially invalid XML data",
				Input: map[string]interface{}{
					"data":      `<root><unclosed>`,
					"operation": "parse",
				},
				Output: map[string]interface{}{
					"error": "invalid XML: XML syntax error on line 1: unexpected EOF",
				},
				Explanation: "The tool provides clear error messages for invalid XML",
			},
		}).
		WithConstraints([]string{
			"XPath support is simplified - no predicates, axes, or functions",
			"Namespace handling uses local names only",
			"Mixed content (text and elements) may not be fully preserved",
			"Large XML documents may impact performance",
			"CDATA sections are treated as regular text",
			"Processing instructions and comments are not preserved",
			"DTD validation is not performed",
			"Attribute order is not guaranteed to be preserved",
		}).
		WithErrorGuidance(map[string]string{
			"invalid XML":                "The provided data is not well-formed XML. Check for unclosed tags, missing quotes, or invalid characters",
			"XPath expression required":  "Query operation requires an 'xpath' parameter. Provide a valid path expression",
			"no elements found for path": "The XPath query didn't match any elements. Check the path and element names",
			"invalid operation":          "Operation must be one of: parse, query, to_json",
			"failed to convert to JSON":  "Unable to convert the XML structure to JSON format",
		}).
		WithCategory("data").
		WithTags([]string{"data", "xml", "parse", "query", "xpath", "transform"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "fast")

	return builder.Build()
}

// xmlProcessExecute is the main processing logic
func xmlProcessExecute(ctx *domain.ToolContext, input XMLProcessInput) (*XMLProcessOutput, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.EmitMessage(fmt.Sprintf("Starting XML processing with operation: %s", input.Operation))
	}

	// Check for XML processing preferences in state
	if ctx.State != nil {
		// Check if attributes should be included by default
		if input.Operation != "query" && !input.IncludeAttributes {
			if includeAttrs, exists := ctx.State.Get("xml_include_attributes_default"); exists {
				if include, ok := includeAttrs.(bool); ok {
					input.IncludeAttributes = include
				}
			}
		}
	}

	var result *XMLProcessOutput
	var err error

	switch input.Operation {
	case "parse":
		result, err = parseXML(input.Data, input.IncludeAttributes)
	case "query":
		if input.XPath == "" {
			err = fmt.Errorf("XPath expression required for query operation")
		} else {
			result, err = queryXML(input.Data, input.XPath, input.IncludeAttributes)
		}
	case "to_json":
		result, err = xmlToJSON(input.Data, input.IncludeAttributes)
	default:
		err = fmt.Errorf("invalid operation: %s", input.Operation)
	}

	// Emit completion or error event
	if ctx.Events != nil {
		if err != nil {
			ctx.Events.EmitError(err)
		} else {
			msg := "XML processing completed successfully"
			if result.RootElement != "" {
				msg = fmt.Sprintf("XML processing completed. Root element: %s", result.RootElement)
			}
			ctx.Events.EmitMessage(msg)
		}
	}

	return result, err
}

// XMLNode represents a parsed XML node
type XMLNode struct {
	XMLName    xml.Name
	Attributes []xml.Attr `xml:",any,attr"`
	Content    string     `xml:",chardata"`
	Children   []XMLNode  `xml:",any"`
}

// parseXML validates and parses XML data
func parseXML(data string, includeAttributes bool) (*XMLProcessOutput, error) {
	var node XMLNode
	if err := xml.Unmarshal([]byte(data), &node); err != nil {
		return &XMLProcessOutput{
			Error: fmt.Sprintf("invalid XML: %v", err),
		}, nil
	}

	result := nodeToMap(&node, includeAttributes)

	return &XMLProcessOutput{
		Result:      result,
		RootElement: node.XMLName.Local,
	}, nil
}

// queryXML performs simplified XPath queries on the data
func queryXML(data string, xpath string, includeAttributes bool) (*XMLProcessOutput, error) {
	var node XMLNode
	if err := xml.Unmarshal([]byte(data), &node); err != nil {
		return &XMLProcessOutput{
			Error: fmt.Sprintf("invalid XML: %v", err),
		}, nil
	}

	// Simplified XPath implementation
	result, err := simpleXPath(&node, xpath, includeAttributes)
	if err != nil {
		return &XMLProcessOutput{
			Error: err.Error(),
		}, nil
	}

	return &XMLProcessOutput{
		Result:      result,
		RootElement: node.XMLName.Local,
	}, nil
}

// simpleXPath implements basic XPath functionality
func simpleXPath(node *XMLNode, xpath string, includeAttributes bool) (interface{}, error) {
	// Remove leading / if present
	xpath = strings.TrimPrefix(xpath, "/")

	if xpath == "" {
		return nodeToMap(node, includeAttributes), nil
	}

	// Split path by /
	parts := strings.Split(xpath, "/")
	current := []*XMLNode{node}

	for _, part := range parts {
		next := []*XMLNode{}

		// Handle attribute selection (@attribute)
		if strings.HasPrefix(part, "@") {
			attrName := strings.TrimPrefix(part, "@")
			results := []string{}
			for _, n := range current {
				for _, attr := range n.Attributes {
					if attr.Name.Local == attrName {
						results = append(results, attr.Value)
					}
				}
			}
			return results, nil
		}

		// Handle element selection
		for _, n := range current {
			if part == "*" {
				// Select all children
				for i := range n.Children {
					next = append(next, &n.Children[i])
				}
			} else if n.XMLName.Local == part {
				// If current node matches, keep it
				next = append(next, n)
			} else {
				// Search in children
				for i := range n.Children {
					if n.Children[i].XMLName.Local == part {
						next = append(next, &n.Children[i])
					}
				}
			}
		}

		if len(next) == 0 {
			return nil, fmt.Errorf("no elements found for path: %s", xpath)
		}

		current = next
	}

	// Convert results to appropriate format
	if len(current) == 1 {
		return nodeToMap(current[0], includeAttributes), nil
	}

	results := []interface{}{}
	for _, n := range current {
		results = append(results, nodeToMap(n, includeAttributes))
	}
	return results, nil
}

// nodeToMap converts an XML node to a map structure
func nodeToMap(node *XMLNode, includeAttributes bool) map[string]interface{} {
	result := make(map[string]interface{})

	// Add element name
	result["_name"] = node.XMLName.Local

	// Add attributes if requested
	if includeAttributes && len(node.Attributes) > 0 {
		attrs := make(map[string]string)
		for _, attr := range node.Attributes {
			attrs[attr.Name.Local] = attr.Value
		}
		result["_attributes"] = attrs
	}

	// Add content if present
	content := strings.TrimSpace(node.Content)
	if content != "" && len(node.Children) == 0 {
		result["_text"] = content
	}

	// Add children
	if len(node.Children) > 0 {
		// Group children by element name
		children := make(map[string][]interface{})
		for _, child := range node.Children {
			childMap := nodeToMap(&child, includeAttributes)
			name := child.XMLName.Local
			if _, exists := children[name]; !exists {
				children[name] = []interface{}{}
			}
			children[name] = append(children[name], childMap)
		}

		// Simplify single-element arrays
		for name, elements := range children {
			if len(elements) == 1 {
				result[name] = elements[0]
			} else {
				result[name] = elements
			}
		}
	}

	return result
}

// xmlToJSON converts XML to JSON
func xmlToJSON(data string, includeAttributes bool) (*XMLProcessOutput, error) {
	// Parse XML first
	parseResult, err := parseXML(data, includeAttributes)
	if err != nil {
		return nil, err
	}
	if parseResult.Error != "" {
		return parseResult, nil
	}

	// Convert to JSON
	jsonBytes, err := json.MarshalIndent(parseResult.Result, "", "  ")
	if err != nil {
		return &XMLProcessOutput{
			Error: fmt.Sprintf("failed to convert to JSON: %v", err),
		}, nil
	}

	return &XMLProcessOutput{
		Result:      string(jsonBytes),
		RootElement: parseResult.RootElement,
	}, nil
}

func init() {
	tools.MustRegisterTool("xml_process", XMLProcess(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "xml_process",
			Category:    "data",
			Tags:        []string{"data", "xml", "parse", "query", "xpath", "transform"},
			Description: "Process XML data: parse, query with simplified XPath, or convert to JSON",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Parse XML",
					Description: "Parse and validate XML data",
					Code:        `XMLProcess().Execute(ctx, XMLProcessInput{Data: xmlStr, Operation: "parse", IncludeAttributes: true})`,
				},
				{
					Name:        "Query with XPath",
					Description: "Extract data using simplified XPath expressions",
					Code:        `XMLProcess().Execute(ctx, XMLProcessInput{Data: xmlStr, Operation: "query", XPath: "book/title"})`,
				},
				{
					Name:        "Convert to JSON",
					Description: "Convert XML to JSON format",
					Code:        `XMLProcess().Execute(ctx, XMLProcessInput{Data: xmlStr, Operation: "to_json", IncludeAttributes: false})`,
				},
			},
		},
		RequiredPermissions: []string{},
		ResourceUsage: tools.ResourceInfo{
			Memory:      "low",
			Network:     false,
			FileSystem:  false,
			Concurrency: true,
		},
	})
}

// MustGetXMLProcess returns the XMLProcess tool or panics if not found
func MustGetXMLProcess() domain.Tool {
	tool, ok := tools.GetTool("xml_process")
	if !ok {
		panic(fmt.Errorf("xml_process tool not found"))
	}
	return tool
}
