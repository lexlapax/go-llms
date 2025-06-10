// ABOUTME: GraphQL discovery and introspection for LLM-friendly operation exploration
// ABOUTME: Provides schema exploration, operation discovery, and example generation

package web

import (
	"fmt"
	"sort"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// GraphQLDiscoveryResult represents discovered GraphQL operations for LLM consumption
type GraphQLDiscoveryResult struct {
	Endpoint   string                     `json:"endpoint"`
	Operations GraphQLOperations          `json:"operations"`
	Types      map[string]GraphQLTypeInfo `json:"types"`
}

// GraphQLOperations contains categorized operations
type GraphQLOperations struct {
	Queries       []GraphQLOperationInfo `json:"queries"`
	Mutations     []GraphQLOperationInfo `json:"mutations"`
	Subscriptions []GraphQLOperationInfo `json:"subscriptions,omitempty"`
}

// GraphQLOperationInfo describes a single operation for LLM understanding
type GraphQLOperationInfo struct {
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Example         string                `json:"example"`
	Returns         string                `json:"returns"`
	Arguments       []GraphQLArgumentInfo `json:"arguments,omitempty"`
	RequiredArgs    []string              `json:"required_args,omitempty"`
	AvailableFields []string              `json:"available_fields,omitempty"`
}

// GraphQLArgumentInfo describes an operation argument
type GraphQLArgumentInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// GraphQLTypeInfo describes a GraphQL type
type GraphQLTypeInfo struct {
	Kind        string   `json:"kind"`
	Description string   `json:"description"`
	Fields      []string `json:"fields,omitempty"`
	EnumValues  []string `json:"enum_values,omitempty"`
}

// DiscoverOperations analyzes a GraphQL schema and returns LLM-friendly operation information
func DiscoverOperations(schema *ast.Schema, endpoint string) (*GraphQLDiscoveryResult, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	result := &GraphQLDiscoveryResult{
		Endpoint: endpoint,
		Operations: GraphQLOperations{
			Queries:       []GraphQLOperationInfo{},
			Mutations:     []GraphQLOperationInfo{},
			Subscriptions: []GraphQLOperationInfo{},
		},
		Types: make(map[string]GraphQLTypeInfo),
	}

	// Process Query type
	if schema.Query != nil {
		queries := processOperationType(schema, schema.Query, "query")
		result.Operations.Queries = queries
	}

	// Process Mutation type
	if schema.Mutation != nil {
		mutations := processOperationType(schema, schema.Mutation, "mutation")
		result.Operations.Mutations = mutations
	}

	// Process Subscription type
	if schema.Subscription != nil {
		subscriptions := processOperationType(schema, schema.Subscription, "subscription")
		result.Operations.Subscriptions = subscriptions
	}

	// Process important types
	for _, typeDef := range schema.Types {
		if typeDef == nil || strings.HasPrefix(typeDef.Name, "__") {
			continue // Skip introspection types
		}

		typeInfo := GraphQLTypeInfo{
			Kind:        string(typeDef.Kind),
			Description: typeDef.Description,
		}

		switch typeDef.Kind {
		case ast.Object, ast.Interface:
			var fields []string
			for _, field := range typeDef.Fields {
				fields = append(fields, field.Name)
			}
			sort.Strings(fields)
			typeInfo.Fields = fields

		case ast.Enum:
			var values []string
			for _, enumVal := range typeDef.EnumValues {
				values = append(values, enumVal.Name)
			}
			sort.Strings(values)
			typeInfo.EnumValues = values
		}

		// Only include types that are commonly used
		if isImportantType(typeDef) {
			result.Types[typeDef.Name] = typeInfo
		}
	}

	return result, nil
}

// processOperationType processes fields of a query/mutation/subscription type
func processOperationType(schema *ast.Schema, typeDef *ast.Definition, operationType string) []GraphQLOperationInfo {
	var operations []GraphQLOperationInfo

	for _, field := range typeDef.Fields {
		if field == nil {
			continue
		}

		// Skip introspection fields
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		op := GraphQLOperationInfo{
			Name:        field.Name,
			Description: field.Description,
			Returns:     getTypeName(field.Type),
		}

		// Process arguments
		var requiredArgs []string
		for _, arg := range field.Arguments {
			argInfo := GraphQLArgumentInfo{
				Name:        arg.Name,
				Type:        getTypeName(arg.Type),
				Description: arg.Description,
				Required:    arg.Type.NonNull,
			}
			op.Arguments = append(op.Arguments, argInfo)

			if arg.Type.NonNull {
				requiredArgs = append(requiredArgs, arg.Name)
			}
		}
		op.RequiredArgs = requiredArgs

		// Generate example
		op.Example = generateOperationExample(operationType, field)

		// Get available fields for return type
		if returnType := getBaseType(field.Type); returnType != nil {
			if typeDef := schema.Types[returnType.Name()]; typeDef != nil {
				var fields []string
				for _, f := range typeDef.Fields {
					fields = append(fields, f.Name)
				}
				sort.Strings(fields)
				if len(fields) > 10 {
					fields = fields[:10] // Limit to first 10 fields
				}
				op.AvailableFields = fields
			}
		}

		operations = append(operations, op)
	}

	// Sort operations by name
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].Name < operations[j].Name
	})

	return operations
}

// generateOperationExample creates an example query for an operation
func generateOperationExample(operationType string, field *ast.FieldDefinition) string {
	var example strings.Builder

	example.WriteString(operationType)
	example.WriteString(" { ")
	example.WriteString(field.Name)

	// Add arguments if any
	if len(field.Arguments) > 0 {
		example.WriteString("(")
		for i, arg := range field.Arguments {
			if i > 0 {
				example.WriteString(", ")
			}
			example.WriteString(arg.Name)
			example.WriteString(": ")
			example.WriteString(getExampleValue(arg.Type))
		}
		example.WriteString(")")
	}

	// Add basic field selection
	example.WriteString(" { ")
	if returnType := getBaseType(field.Type); returnType != nil {
		if returnType.Name() == "String" || returnType.Name() == "Int" || returnType.Name() == "Boolean" {
			example.WriteString("# Scalar type - no fields to select")
		} else {
			example.WriteString("# Add fields to select here")
		}
	}
	example.WriteString(" } }")

	return example.String()
}

// getExampleValue returns an example value for a GraphQL type
func getExampleValue(t *ast.Type) string {
	baseType := getBaseType(t)
	if baseType == nil {
		return "null"
	}

	switch baseType.Name() {
	case "String":
		return `"example"`
	case "Int":
		return "123"
	case "Float":
		return "123.45"
	case "Boolean":
		return "true"
	case "ID":
		return `"id123"`
	default:
		return "{...}" // Object or input type
	}
}

// getTypeName returns a human-readable type name
func getTypeName(t *ast.Type) string {
	if t == nil {
		return "Unknown"
	}

	var name string
	var nonNull bool
	var isList bool

	// Unwrap type modifiers
	current := t
	for current != nil {
		if current.NonNull {
			nonNull = true
		}
		if current.Elem != nil {
			isList = true
			current = current.Elem
		} else {
			name = current.Name()
			break
		}
	}

	// Build type string
	result := name
	if isList {
		result = "[" + result + "]"
	}
	if nonNull {
		result = result + "!"
	}

	return result
}

// getBaseType returns the base type without modifiers
func getBaseType(t *ast.Type) *ast.Type {
	if t == nil {
		return nil
	}

	current := t
	for current.Elem != nil {
		current = current.Elem
	}

	return current
}

// isImportantType determines if a type should be included in discovery results
func isImportantType(typeDef *ast.Definition) bool {
	// Skip GraphQL built-in types
	if strings.HasPrefix(typeDef.Name, "__") {
		return false
	}

	// Skip scalar types (they're self-explanatory)
	if typeDef.Kind == ast.Scalar {
		return false
	}

	// Include types with descriptions or multiple fields
	if typeDef.Description != "" {
		return true
	}

	if len(typeDef.Fields) > 1 {
		return true
	}

	if len(typeDef.EnumValues) > 0 {
		return true
	}

	return false
}
