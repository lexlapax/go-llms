// ABOUTME: XMLProcess tool provides XML parsing and querying capabilities
// ABOUTME: This tool enables agents to work with XML data without requiring LLM processing

package data

import (
	"context"
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
func XMLProcess() domain.Tool {
	return atools.NewTool(
		"xml_process",
		"Process XML data: parse, query with simplified XPath, or convert to JSON",
		func(ctx context.Context, input XMLProcessInput) (*XMLProcessOutput, error) {
			return executeXMLProcess(ctx, input)
		},
		xmlProcessParamSchema,
	)
}

// XMLNode represents a parsed XML node
type XMLNode struct {
	XMLName    xml.Name
	Attributes []xml.Attr `xml:",any,attr"`
	Content    string     `xml:",chardata"`
	Children   []XMLNode  `xml:",any"`
}

// executeXMLProcess processes the XML according to the specified operation
func executeXMLProcess(ctx context.Context, input XMLProcessInput) (*XMLProcessOutput, error) {
	switch input.Operation {
	case "parse":
		return parseXML(input.Data, input.IncludeAttributes)
	case "query":
		if input.XPath == "" {
			return nil, fmt.Errorf("XPath expression required for query operation")
		}
		return queryXML(input.Data, input.XPath, input.IncludeAttributes)
	case "to_json":
		return xmlToJSON(input.Data, input.IncludeAttributes)
	default:
		return nil, fmt.Errorf("invalid operation: %s", input.Operation)
	}
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
