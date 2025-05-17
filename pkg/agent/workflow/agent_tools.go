// Package workflow provides agent workflow implementations.
package workflow

import (
	"encoding/json"
	"fmt"
	"strings"
)

// getToolsDescription creates a description of available tools - optimized version
func (a *DefaultAgent) getToolsDescription() string {
	if len(a.tools) == 0 {
		return ""
	}

	// Optimization: Use cached description if available
	if a.cachedToolsDescription != "" {
		return a.cachedToolsDescription
	}

	// Estimate the size for pre-allocation (reduce allocations)
	// Base estimate: 500 characters for standard text + ~200 per tool
	estimatedSize := 500 + (len(a.tools) * 200)

	var builder strings.Builder
	// Pre-allocate builder capacity to reduce allocations
	builder.Grow(estimatedSize)
	builder.WriteString("You have access to the following tools:\n\n")

	// Pre-allocate tool definitions to avoid resizing
	toolDefinitions := make([]map[string]interface{}, 0, len(a.tools))

	for name, tool := range a.tools {
		builder.WriteString(fmt.Sprintf("Tool: %s\n", name))
		builder.WriteString(fmt.Sprintf("Description: %s\n", tool.Description()))

		// Add parameter schema if available
		schema := tool.ParameterSchema()
		if schema != nil {
			// Optimization: Use MarshalIndent only once per schema
			paramSchemaJSON, err := json.MarshalIndent(schema, "", "  ")
			if err == nil {
				builder.WriteString(fmt.Sprintf("Parameters: %s\n", string(paramSchemaJSON)))
			}

			// Add to OpenAI-style tool definitions
			toolDefinitions = append(toolDefinitions, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        name,
					"description": tool.Description(),
					"parameters":  schema,
				},
			})
		} else {
			// Add minimal tool definition if no schema is available
			toolDefinitions = append(toolDefinitions, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        name,
					"description": tool.Description(),
					"parameters": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			})
		}

		builder.WriteString("\n")
	}

	// Add usage instructions with proper escaping
	builder.WriteString("\nTo use a tool, respond with one of these formats:\n")
	builder.WriteString("```json\n{\"tool\": \"tool_name\", \"params\": {...}}\n```\n")
	builder.WriteString("\nOR\n\n")
	builder.WriteString("```json\n{\"tool_calls\": [{\"function\": {\"name\": \"tool_name\", \"arguments\": \"{...}\"}}]}\n```\n")

	// Add OpenAI-style tool definitions as JSON
	if len(toolDefinitions) > 0 {
		toolDefsJSON, err := json.MarshalIndent(toolDefinitions, "", "  ")
		if err == nil {
			builder.WriteString("\nTool definitions in OpenAI format:\n```json\n")
			builder.Write(toolDefsJSON)
			builder.WriteString("\n```\n")
		}
	}

	// Cache the result for future calls
	a.cachedToolsDescription = builder.String()
	return a.cachedToolsDescription
}

// getToolsDescription creates a description of available tools - unoptimized version
func (a *UnoptimizedDefaultAgent) getToolsDescription() string {
	if len(a.tools) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("You have access to the following tools:\n\n")

	// Also build OpenAI-style tool definitions for the model
	var toolDefinitions []map[string]interface{}

	for name, tool := range a.tools {
		builder.WriteString(fmt.Sprintf("Tool: %s\n", name))
		builder.WriteString(fmt.Sprintf("Description: %s\n", tool.Description()))

		// Add parameter schema if available
		schema := tool.ParameterSchema()
		if schema != nil {
			paramSchemaJSON, err := json.MarshalIndent(schema, "", "  ")
			if err == nil {
				builder.WriteString(fmt.Sprintf("Parameters: %s\n", string(paramSchemaJSON)))
			}

			// Add to OpenAI-style tool definitions
			toolDefinitions = append(toolDefinitions, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        name,
					"description": tool.Description(),
					"parameters":  schema,
				},
			})
		} else {
			// Add minimal tool definition if no schema is available
			toolDefinitions = append(toolDefinitions, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        name,
					"description": tool.Description(),
					"parameters": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			})
		}

		builder.WriteString("\n")
	}

	builder.WriteString("\nTo use a tool, respond with one of these formats:\n")
	builder.WriteString("```json\n{\"tool\": \"tool_name\", \"params\": {...}}\n```\n")
	builder.WriteString("\nOR\n\n")
	builder.WriteString("```json\n{\"tool_calls\": [{\"function\": {\"name\": \"tool_name\", \"arguments\": \"{...}\"}}]}\n```\n")

	// Add OpenAI-style tool definitions as JSON
	if len(toolDefinitions) > 0 {
		toolDefsJSON, err := json.MarshalIndent(toolDefinitions, "", "  ")
		if err == nil {
			builder.WriteString("\nTool definitions in OpenAI format:\n```json\n")
			builder.Write(toolDefsJSON)
			builder.WriteString("\n```\n")
		}
	}

	return builder.String()
}
