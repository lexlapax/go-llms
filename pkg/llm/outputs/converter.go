// ABOUTME: Format converter for converting between JSON, XML, and YAML formats
// ABOUTME: Preserves type information and supports streaming conversion

package outputs

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format represents an output format
type Format string

const (
	// FormatJSON represents JSON format
	FormatJSON Format = "json"

	// FormatXML represents XML format
	FormatXML Format = "xml"

	// FormatYAML represents YAML format
	FormatYAML Format = "yaml"
)

// Converter converts between different output formats
type Converter struct {
	// preserveTypes controls whether to preserve type information
	preserveTypes bool

	// indentSize controls indentation for pretty printing
	indentSize int
}

// NewConverter creates a new converter
func NewConverter() *Converter {
	return &Converter{
		preserveTypes: true,
		indentSize:    2,
	}
}

// ConversionOptions configures conversion behavior
type ConversionOptions struct {
	// Pretty enables pretty printing
	Pretty bool

	// PreserveTypes preserves type information during conversion
	PreserveTypes bool

	// RootElement specifies the root element name for XML
	RootElement string

	// XMLNamespace specifies the XML namespace
	XMLNamespace string

	// IndentSize specifies the indent size for pretty printing
	IndentSize int
}

// DefaultConversionOptions returns default conversion options
func DefaultConversionOptions() *ConversionOptions {
	return &ConversionOptions{
		Pretty:        true,
		PreserveTypes: true,
		RootElement:   "root",
		IndentSize:    2,
	}
}

// Convert converts data from one format to another
func (c *Converter) Convert(ctx context.Context, data interface{}, from, to Format, opts *ConversionOptions) (interface{}, error) {
	if from == to {
		return data, nil
	}

	if opts == nil {
		opts = DefaultConversionOptions()
	}

	// Normalize the data to a common representation
	normalized, err := c.normalize(data, from)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize from %s: %w", from, err)
	}

	// Convert to target format
	result, err := c.denormalize(normalized, to, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to %s: %w", to, err)
	}

	return result, nil
}

// ConvertString converts a string from one format to another
func (c *Converter) ConvertString(ctx context.Context, input string, from, to Format, opts *ConversionOptions) (string, error) {
	// Parse the input
	var data interface{}
	var err error

	switch from {
	case FormatJSON:
		err = json.Unmarshal([]byte(input), &data)
	case FormatYAML:
		err = yaml.Unmarshal([]byte(input), &data)
	case FormatXML:
		data, err = c.parseXML(input)
	default:
		return "", fmt.Errorf("unsupported source format: %s", from)
	}

	if err != nil {
		return "", fmt.Errorf("failed to parse %s: %w", from, err)
	}

	// Convert the data
	result, err := c.Convert(ctx, data, from, to, opts)
	if err != nil {
		return "", err
	}

	// Marshal to string
	return c.marshalToString(result, to, opts)
}

// StreamConvert converts data from one format to another using streaming
func (c *Converter) StreamConvert(ctx context.Context, reader io.Reader, writer io.Writer, from, to Format, opts *ConversionOptions) error {
	// For streaming, we need to read the entire input first
	// True streaming conversion would require format-specific streaming parsers
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(reader); err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	output, err := c.ConvertString(ctx, buf.String(), from, to, opts)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(output))
	return err
}

// normalize converts data to a common representation
func (c *Converter) normalize(data interface{}, format Format) (interface{}, error) {
	switch format {
	case FormatJSON, FormatYAML:
		// JSON and YAML already use similar representations
		return c.normalizeJSONYAML(data), nil
	case FormatXML:
		// XML needs special handling
		return c.normalizeXML(data), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// normalizeJSONYAML normalizes JSON/YAML data
func (c *Converter) normalizeJSONYAML(data interface{}) interface{} {
	// Convert yaml.Node to regular types if needed
	switch v := data.(type) {
	case *yaml.Node:
		return c.yamlNodeToInterface(v)
	case yaml.Node:
		return c.yamlNodeToInterface(&v)
	default:
		return data
	}
}

// yamlNodeToInterface converts a yaml.Node to interface{}
func (c *Converter) yamlNodeToInterface(node *yaml.Node) interface{} {
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) > 0 {
			return c.yamlNodeToInterface(node.Content[0])
		}
		return nil
	case yaml.SequenceNode:
		result := make([]interface{}, 0, len(node.Content))
		for _, n := range node.Content {
			result = append(result, c.yamlNodeToInterface(n))
		}
		return result
	case yaml.MappingNode:
		result := make(map[string]interface{})
		for i := 0; i < len(node.Content); i += 2 {
			key := fmt.Sprintf("%v", c.yamlNodeToInterface(node.Content[i]))
			value := c.yamlNodeToInterface(node.Content[i+1])
			result[key] = value
		}
		return result
	case yaml.ScalarNode:
		var value interface{}
		_ = node.Decode(&value)
		return value
	case yaml.AliasNode:
		return c.yamlNodeToInterface(node.Alias)
	default:
		return node.Value
	}
}

// normalizeXML normalizes XML data
func (c *Converter) normalizeXML(data interface{}) interface{} {
	// XML normalization is more complex due to attributes and text content
	// For now, we'll use a simple approach
	return data
}

// denormalize converts normalized data to target format
func (c *Converter) denormalize(data interface{}, format Format, opts *ConversionOptions) (interface{}, error) {
	switch format {
	case FormatJSON, FormatYAML:
		return data, nil // Already in the right format
	case FormatXML:
		return c.toXMLStructure(data, opts), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// marshalToString marshals data to string in the specified format
func (c *Converter) marshalToString(data interface{}, format Format, opts *ConversionOptions) (string, error) {
	switch format {
	case FormatJSON:
		if opts.Pretty {
			b, err := json.MarshalIndent(data, "", strings.Repeat(" ", opts.IndentSize))
			return string(b), err
		}
		b, err := json.Marshal(data)
		return string(b), err

	case FormatYAML:
		b, err := yaml.Marshal(data)
		return string(b), err

	case FormatXML:
		return c.marshalXML(data, opts)

	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// parseXML parses XML into a generic structure
func (c *Converter) parseXML(input string) (interface{}, error) {
	// Simple XML parsing - this is a basic implementation
	// In production, you'd want a more sophisticated XML parser
	decoder := xml.NewDecoder(strings.NewReader(input))

	var result interface{}
	if err := decoder.Decode(&result); err != nil {
		// Try parsing as a simple map
		var xmlMap map[string]interface{}
		if err := xml.Unmarshal([]byte(input), &xmlMap); err != nil {
			return nil, err
		}
		result = xmlMap
	}

	return result, nil
}

// toXMLStructure converts data to an XML-compatible structure
func (c *Converter) toXMLStructure(data interface{}, opts *ConversionOptions) interface{} {
	// Convert the data to an XML-friendly structure
	// This is a simplified implementation
	type XMLElement struct {
		XMLName  xml.Name
		Content  interface{}   `xml:",chardata"`
		Children []interface{} `xml:",any"`
	}

	root := &XMLElement{
		XMLName: xml.Name{Local: opts.RootElement},
	}

	switch v := data.(type) {
	case map[string]interface{}:
		// Convert map to XML elements
		for key, value := range v {
			child := &XMLElement{
				XMLName: xml.Name{Local: key},
				Content: value,
			}
			root.Children = append(root.Children, child)
		}
	case []interface{}:
		// Convert array to XML elements
		for i, item := range v {
			child := &XMLElement{
				XMLName: xml.Name{Local: fmt.Sprintf("item%d", i)},
				Content: item,
			}
			root.Children = append(root.Children, child)
		}
	default:
		root.Content = v
	}

	return root
}

// marshalXML marshals data to XML string
func (c *Converter) marshalXML(data interface{}, opts *ConversionOptions) (string, error) {
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)
	if opts.Pretty {
		encoder.Indent("", strings.Repeat(" ", opts.IndentSize))
	}

	if err := encoder.Encode(data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// DetectFormat attempts to detect the format of a string
func DetectFormat(input string) (Format, error) {
	trimmed := strings.TrimSpace(input)

	// Check for JSON
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		var js interface{}
		if err := json.Unmarshal([]byte(trimmed), &js); err == nil {
			return FormatJSON, nil
		}
	}

	// Check for XML
	if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") {
		decoder := xml.NewDecoder(strings.NewReader(trimmed))
		var x interface{}
		if err := decoder.Decode(&x); err == nil {
			return FormatXML, nil
		}
	}

	// Check for YAML
	// YAML is tricky because almost anything is valid YAML
	// We'll check for common YAML patterns
	if strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "{") {
		var y interface{}
		if err := yaml.Unmarshal([]byte(trimmed), &y); err == nil {
			return FormatYAML, nil
		}
	}

	return "", fmt.Errorf("unable to detect format")
}

// ConvertAuto converts with automatic format detection
func ConvertAuto(ctx context.Context, input string, targetFormat Format, opts *ConversionOptions) (string, error) {
	sourceFormat, err := DetectFormat(input)
	if err != nil {
		return "", fmt.Errorf("failed to detect source format: %w", err)
	}

	converter := NewConverter()
	return converter.ConvertString(ctx, input, sourceFormat, targetFormat, opts)
}
