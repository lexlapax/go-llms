// ABOUTME: JSONProcess tool provides JSON parsing, querying with JSONPath, and transformation capabilities
// ABOUTME: This tool enables agents to work with JSON data without requiring LLM processing

package data

import (
	"context"
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
func JSONProcess() domain.Tool {
	return atools.NewTool(
		"json_process",
		"Process JSON data: parse, query with JSONPath, or transform",
		func(ctx context.Context, input JSONProcessInput) (*JSONProcessOutput, error) {
			switch input.Operation {
			case "parse":
				return parseJSON(input.Data)
			case "query":
				if input.JSONPath == "" {
					return nil, fmt.Errorf("JSONPath expression required for query operation")
				}
				return queryJSON(input.Data, input.JSONPath)
			case "transform":
				if input.Transform == "" {
					return nil, fmt.Errorf("transform type required for transform operation")
				}
				return transformJSON(input.Data, input.Transform)
			default:
				return nil, fmt.Errorf("invalid operation: %s", input.Operation)
			}
		},
		jsonProcessParamSchema,
	)
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
