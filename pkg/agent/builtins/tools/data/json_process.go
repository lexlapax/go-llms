// ABOUTME: JSONProcess tool provides JSON parsing, querying with JSONPath, and transformation capabilities
// ABOUTME: This tool enables agents to work with JSON data without requiring LLM processing

package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/builtins"
	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	atools "github.com/lexlapax/go-llms/pkg/agent/tools"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// JSONProcessInput represents the input for JSON processing operations
type JSONProcessInput struct {
	// The JSON data to process (as a string)
	Data string `json:"data" jsonschema:"title=JSON Data,description=The JSON data to process,required"`

	// Operation to perform: parse, query, transform
	Operation string `json:"operation" jsonschema:"title=Operation,description=Operation to perform: parse query or transform,enum=parse,enum=query,enum=transform,required"`

	// JSONPath expression for query operation
	JSONPath string `json:"jsonpath,omitempty" jsonschema:"title=JSONPath,description=JSONPath expression for query operation"`

	// Transformation type for transform operation
	Transform string `json:"transform,omitempty" jsonschema:"title=Transform,description=Transformation type: extract_keys extract_values flatten prettify minify,enum=extract_keys,enum=extract_values,enum=flatten,enum=prettify,enum=minify"`
}

// JSONProcessOutput represents the output of JSON processing
type JSONProcessOutput struct {
	// The processed result
	Result interface{} `json:"result"`

	// Error message if any
	Error string `json:"error,omitempty"`

	// Type of the result
	ResultType string `json:"result_type"`
}

// jsonProcessParamSchema defines parameters for the JSONProcess tool
var jsonProcessParamSchema = &sdomain.Schema{
	Type: "object",
	Properties: map[string]sdomain.Property{
		"data": {
			Type:        "string",
			Description: "The JSON data to process",
		},
		"operation": {
			Type:        "string",
			Description: "Operation to perform: parse, query, or transform",
			Enum:        []string{"parse", "query", "transform"},
		},
		"jsonpath": {
			Type:        "string",
			Description: "JSONPath expression for query operation",
		},
		"transform": {
			Type:        "string",
			Description: "Transformation type",
			Enum:        []string{"extract_keys", "extract_values", "flatten", "prettify", "minify"},
		},
	},
	Required: []string{"data", "operation"},
}

// JSONProcess creates a tool for processing JSON data
// This tool provides JSON manipulation capabilities including parsing,
// querying with JSONPath expressions, and various transformations
// like flattening, prettifying, and extracting keys/values.
func JSONProcess() domain.Tool {

	// Create output schema for JSONProcessOutput
	outputSchema := &sdomain.Schema{
		Type: "object",
		Properties: map[string]sdomain.Property{
			"result": {
				Type:        "any",
				Description: "The processed result (can be any JSON type)",
			},
			"error": {
				Type:        "string",
				Description: "Error message if any",
			},
			"result_type": {
				Type:        "string",
				Description: "The type of the result (e.g., 'map[string]interface{}', 'string', '[]interface{}')",
			},
		},
		Required: []string{"result_type"},
	}

	builder := atools.NewToolBuilder("json_process", "Process JSON data: parse, query with JSONPath, or transform").
		WithFunction(jsonProcessExecute).
		WithParameterSchema(jsonProcessParamSchema).
		WithOutputSchema(outputSchema).
		WithUsageInstructions(`Use this tool to process JSON data in various ways:

Parse Operation:
- Validates JSON syntax and parses the data
- Returns the parsed data structure and its type
- Useful for checking if data is valid JSON

Query Operation (JSONPath):
- Extract specific values using JSONPath expressions
- Supports basic JSONPath syntax:
  - $ or empty: Root object
  - .field: Access object field
  - [n]: Array index access
  - .field[n]: Combination of field and array access
  - Nested paths: $.users[0].address.city

Transform Operations:
- extract_keys: Get all keys from the JSON structure (includes nested paths)
- extract_values: Get all leaf values from the JSON
- flatten: Convert nested JSON to flat key-value pairs
- prettify: Format JSON with indentation for readability
- minify: Remove unnecessary whitespace for compact representation

JSONPath Examples:
- $.name: Get the 'name' field from root
- $.users[0]: Get the first user from 'users' array
- $.users[0].email: Get email of the first user
- $.products[*].price: Get all product prices (Note: [*] not fully supported in basic implementation)

For complex JSONPath queries beyond basic field and array access, consider using the result of a parse operation and processing it further.`).
		WithExamples([]domain.ToolExample{
			{
				Name:        "Parse and validate JSON",
				Description: "Check if a string is valid JSON and see its structure",
				Scenario:    "When you receive data and need to verify it's valid JSON",
				Input: map[string]interface{}{
					"data":      `{"name": "John", "age": 30, "city": "New York"}`,
					"operation": "parse",
				},
				Output: map[string]interface{}{
					"result": map[string]interface{}{
						"name": "John",
						"age":  float64(30),
						"city": "New York",
					},
					"result_type": "map[string]interface {}",
				},
				Explanation: "The tool parses JSON and returns the data structure with type information",
			},
			{
				Name:        "Query nested data with JSONPath",
				Description: "Extract specific values from complex JSON structures",
				Scenario:    "When you need to extract data from a specific path in JSON",
				Input: map[string]interface{}{
					"data":      `{"users": [{"id": 1, "name": "Alice", "email": "alice@example.com"}, {"id": 2, "name": "Bob", "email": "bob@example.com"}]}`,
					"operation": "query",
					"jsonpath":  "users[0].email",
				},
				Output: map[string]interface{}{
					"result":      "alice@example.com",
					"result_type": "string",
				},
				Explanation: "JSONPath allows precise extraction of values from nested structures",
			},
			{
				Name:        "Extract all keys from JSON",
				Description: "Get a list of all keys including nested paths",
				Scenario:    "When you need to understand the structure of complex JSON",
				Input: map[string]interface{}{
					"data":      `{"user": {"name": "John", "address": {"city": "NYC", "zip": "10001"}}, "active": true}`,
					"operation": "transform",
					"transform": "extract_keys",
				},
				Output: map[string]interface{}{
					"result":      []string{"user", "user.name", "user.address", "user.address.city", "user.address.zip", "active"},
					"result_type": "[]string",
				},
				Explanation: "Returns all keys with their full paths, useful for understanding data structure",
			},
			{
				Name:        "Flatten nested JSON",
				Description: "Convert nested structure to flat key-value pairs",
				Scenario:    "When you need to work with flat data structures or export to CSV",
				Input: map[string]interface{}{
					"data":      `{"user": {"name": "John", "scores": [85, 90, 78]}, "active": true}`,
					"operation": "transform",
					"transform": "flatten",
				},
				Output: map[string]interface{}{
					"result": map[string]interface{}{
						"user.name":      "John",
						"user.scores[0]": float64(85),
						"user.scores[1]": float64(90),
						"user.scores[2]": float64(78),
						"active":         true,
					},
					"result_type": "map[string]interface {}",
				},
				Explanation: "Flattening creates a single-level object with compound keys",
			},
			{
				Name:        "Pretty print JSON",
				Description: "Format JSON with proper indentation",
				Scenario:    "When you need human-readable JSON output",
				Input: map[string]interface{}{
					"data":      `{"name":"John","age":30,"city":"New York"}`,
					"operation": "transform",
					"transform": "prettify",
				},
				Output: map[string]interface{}{
					"result": `{
  "name": "John",
  "age": 30,
  "city": "New York"
}`,
					"result_type": "string",
				},
				Explanation: "Prettify adds indentation and line breaks for readability",
			},
			{
				Name:        "Handle invalid JSON gracefully",
				Description: "Error handling for malformed JSON",
				Scenario:    "When processing potentially invalid JSON data",
				Input: map[string]interface{}{
					"data":      `{"name": "John", "age": 30,}`, // Trailing comma
					"operation": "parse",
				},
				Output: map[string]interface{}{
					"error":       "invalid JSON: invalid character '}' after object key:value pair",
					"result_type": "",
				},
				Explanation: "The tool provides clear error messages for invalid JSON",
			},
			{
				Name:        "Query array elements",
				Description: "Access specific elements in JSON arrays",
				Scenario:    "When you need to extract data from arrays",
				Input: map[string]interface{}{
					"data":      `{"items": ["apple", "banana", "cherry"], "count": 3}`,
					"operation": "query",
					"jsonpath":  "items[1]",
				},
				Output: map[string]interface{}{
					"result":      "banana",
					"result_type": "string",
				},
				Explanation: "Array indices start at 0, so [1] gets the second element",
			},
		}).
		WithConstraints([]string{
			"JSONPath implementation supports basic path notation only (no wildcards or filters)",
			"Array slicing and wildcard operations are not supported",
			"JSONPath expressions must be valid according to basic JSONPath syntax",
			"Large JSON data may impact performance",
			"Circular references in JSON are not supported",
			"Transform operations preserve the original data types",
			"Pretty printing uses 2-space indentation",
		}).
		WithErrorGuidance(map[string]string{
			"invalid JSON":                 "The provided data is not valid JSON. Check for syntax errors like missing quotes, trailing commas, or unmatched brackets",
			"JSONPath expression required": "Query operation requires a 'jsonpath' parameter. Provide a valid JSONPath expression",
			"transform type required":      "Transform operation requires a 'transform' parameter. Choose from: extract_keys, extract_values, flatten, prettify, minify",
			"invalid operation":            "Operation must be one of: parse, query, transform",
			"field not found":              "The specified field doesn't exist in the JSON. Check the path and use parse operation to see the structure",
			"cannot access field":          "Trying to access a field on a non-object. Ensure the path points to an object before accessing fields",
			"cannot index non-array":       "Array index notation [n] can only be used on arrays. Check that the path points to an array",
			"array index out of bounds":    "The array index is outside the valid range. Arrays are 0-indexed",
			"invalid array index":          "Array indices must be non-negative integers",
			"failed to prettify":           "Unable to format the JSON. The data structure may be too complex",
			"failed to minify":             "Unable to minimize the JSON. The data structure may be too complex",
		}).
		WithCategory("data").
		WithTags([]string{"data", "json", "parse", "query", "transform", "jsonpath"}).
		WithVersion("2.0.0").
		WithBehavior(true, false, false, "fast")

	return builder.Build()
}

// jsonProcessExecute is the main processing logic
func jsonProcessExecute(ctx *domain.ToolContext, input JSONProcessInput) (*JSONProcessOutput, error) {
	// Emit start event
	if ctx.Events != nil {
		ctx.Events.EmitMessage(fmt.Sprintf("Starting JSON processing with operation: %s", input.Operation))
	}

	// Check for any JSON processing configuration in state
	if ctx.State != nil {
		// Check for default indent setting for prettify
		if input.Operation == "transform" && input.Transform == "prettify" {
			if indentSize, exists := ctx.State.Get("json_prettify_indent"); exists {
				// Note: This could be used in the prettify operation if needed
				_ = indentSize // Mark as intentionally unused for now
			}
		}
	}

	var result *JSONProcessOutput
	var err error

	switch input.Operation {
	case "parse":
		result, err = parseJSON(input.Data)
	case "query":
		if input.JSONPath == "" {
			err = fmt.Errorf("JSONPath expression required for query operation")
		} else {
			result, err = queryJSON(input.Data, input.JSONPath)
		}
	case "transform":
		if input.Transform == "" {
			err = fmt.Errorf("transform type required for transform operation")
		} else {
			result, err = transformJSON(input.Data, input.Transform)
		}
	default:
		err = fmt.Errorf("invalid operation: %s", input.Operation)
	}

	// Emit completion or error event
	if ctx.Events != nil {
		if err != nil {
			ctx.Events.EmitError(err)
		} else {
			ctx.Events.EmitMessage("JSON processing completed successfully")
		}
	}

	return result, err
}

// parseJSON validates and parses JSON data
func parseJSON(data string) (*JSONProcessOutput, error) {
	var result interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return &JSONProcessOutput{
			Error: fmt.Sprintf("invalid JSON: %v", err),
		}, nil
	}

	return &JSONProcessOutput{
		Result:     result,
		ResultType: fmt.Sprintf("%T", result),
	}, nil
}

// queryJSON performs JSONPath queries on the data
func queryJSON(data string, jsonPath string) (*JSONProcessOutput, error) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return &JSONProcessOutput{
			Error: fmt.Sprintf("invalid JSON: %v", err),
		}, nil
	}

	// Simple JSONPath implementation for common cases
	result, err := simpleJSONPath(jsonData, jsonPath)
	if err != nil {
		return &JSONProcessOutput{
			Error: err.Error(),
		}, nil
	}

	return &JSONProcessOutput{
		Result:     result,
		ResultType: fmt.Sprintf("%T", result),
	}, nil
}

// simpleJSONPath implements basic JSONPath functionality
func simpleJSONPath(data interface{}, path string) (interface{}, error) {
	// Remove leading $ if present
	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")

	if path == "" {
		return data, nil
	}

	// Split path by dots
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		// Handle array index notation [n]
		if strings.HasSuffix(part, "]") {
			if idx := strings.Index(part, "["); idx >= 0 {
				fieldName := part[:idx]
				indexStr := part[idx+1 : len(part)-1]

				// Navigate to field first if present
				if fieldName != "" {
					if m, ok := current.(map[string]interface{}); ok {
						current = m[fieldName]
					} else {
						return nil, fmt.Errorf("cannot access field %s on non-object", fieldName)
					}
				}

				// Handle array access
				if arr, ok := current.([]interface{}); ok {
					var index int
					if _, err := fmt.Sscanf(indexStr, "%d", &index); err != nil {
						return nil, fmt.Errorf("invalid array index: %s", indexStr)
					}
					if index < 0 || index >= len(arr) {
						return nil, fmt.Errorf("array index out of bounds: %d", index)
					}
					current = arr[index]
				} else {
					return nil, fmt.Errorf("cannot index non-array with [%s]", indexStr)
				}
				continue
			}
		}

		// Handle object field access
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[part]; exists {
				current = val
			} else {
				return nil, fmt.Errorf("field not found: %s", part)
			}
		} else {
			return nil, fmt.Errorf("cannot access field %s on non-object", part)
		}
	}

	return current, nil
}

// transformJSON applies various transformations to JSON data
func transformJSON(data string, transformType string) (*JSONProcessOutput, error) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return &JSONProcessOutput{
			Error: fmt.Sprintf("invalid JSON: %v", err),
		}, nil
	}

	var result interface{}
	var err error

	switch transformType {
	case "extract_keys":
		result = extractKeys(jsonData)
	case "extract_values":
		result = extractValues(jsonData)
	case "flatten":
		result = flatten(jsonData, "")
	case "prettify":
		prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			return &JSONProcessOutput{
				Error: fmt.Sprintf("failed to prettify: %v", err),
			}, nil
		}
		result = string(prettyJSON)
	case "minify":
		minified, err := json.Marshal(jsonData)
		if err != nil {
			return &JSONProcessOutput{
				Error: fmt.Sprintf("failed to minify: %v", err),
			}, nil
		}
		result = string(minified)
	default:
		err = fmt.Errorf("unknown transform type: %s", transformType)
	}

	if err != nil {
		return &JSONProcessOutput{
			Error: err.Error(),
		}, nil
	}

	return &JSONProcessOutput{
		Result:     result,
		ResultType: fmt.Sprintf("%T", result),
	}, nil
}

// extractKeys recursively extracts all keys from JSON
func extractKeys(data interface{}) []string {
	keys := []string{}

	switch v := data.(type) {
	case map[string]interface{}:
		for k, val := range v {
			keys = append(keys, k)
			// Recursively extract keys from nested objects
			subKeys := extractKeys(val)
			for _, subKey := range subKeys {
				keys = append(keys, k+"."+subKey)
			}
		}
	case []interface{}:
		// For arrays, extract keys from all elements
		for i, item := range v {
			subKeys := extractKeys(item)
			for _, subKey := range subKeys {
				keys = append(keys, fmt.Sprintf("[%d].%s", i, subKey))
			}
		}
	}

	return keys
}

// extractValues recursively extracts all leaf values from JSON
func extractValues(data interface{}) []interface{} {
	values := []interface{}{}

	switch v := data.(type) {
	case map[string]interface{}:
		for _, val := range v {
			subValues := extractValues(val)
			values = append(values, subValues...)
		}
	case []interface{}:
		for _, item := range v {
			subValues := extractValues(item)
			values = append(values, subValues...)
		}
	default:
		// Leaf value
		values = append(values, v)
	}

	return values
}

// flatten converts nested JSON to flat key-value pairs
func flatten(data interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	switch v := data.(type) {
	case map[string]interface{}:
		for k, val := range v {
			newKey := k
			if prefix != "" {
				newKey = prefix + "." + k
			}
			flattened := flatten(val, newKey)
			for fk, fv := range flattened {
				result[fk] = fv
			}
		}
	case []interface{}:
		for i, item := range v {
			newKey := fmt.Sprintf("%s[%d]", prefix, i)
			flattened := flatten(item, newKey)
			for fk, fv := range flattened {
				result[fk] = fv
			}
		}
	default:
		result[prefix] = v
	}

	return result
}

func init() {
	tools.MustRegisterTool("json_process", JSONProcess(), tools.ToolMetadata{
		Metadata: builtins.Metadata{
			Name:        "json_process",
			Category:    "data",
			Tags:        []string{"data", "json", "parse", "query", "transform", "jsonpath"},
			Description: "Process JSON data: parse, query with JSONPath, or transform",
			Version:     "1.0.0",
			Examples: []builtins.Example{
				{
					Name:        "Parse JSON",
					Description: "Parse and validate JSON data",
					Code:        `JSONProcess().Execute(ctx, JSONProcessInput{Data: jsonStr, Operation: "parse"})`,
				},
				{
					Name:        "Query with JSONPath",
					Description: "Extract data using JSONPath expressions",
					Code:        `JSONProcess().Execute(ctx, JSONProcessInput{Data: jsonStr, Operation: "query", JSONPath: "$.users[0].name"})`,
				},
				{
					Name:        "Transform JSON",
					Description: "Apply transformations like flatten or prettify",
					Code:        `JSONProcess().Execute(ctx, JSONProcessInput{Data: jsonStr, Operation: "transform", Transform: "flatten"})`,
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

// MustGetJSONProcess retrieves the registered JSONProcess tool or panics
func MustGetJSONProcess() domain.Tool {
	return tools.MustGetTool("json_process")
}
