// ABOUTME: Schema converter for bridging between domain.Schema and map[string]interface{} types
// ABOUTME: Critical for go-llmspell and other bridge layers to work with schema objects seamlessly

package types

import (
	"fmt"
	"reflect"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// SchemaConverter handles conversions between domain.Schema and map[string]any
type SchemaConverter struct{}

func (c *SchemaConverter) Name() string  { return "SchemaConverter" }
func (c *SchemaConverter) Priority() int { return 200 } // High priority for schema conversions

func (c *SchemaConverter) CanConvert(fromType, toType reflect.Type) bool {
	schemaType := reflect.TypeOf((*schemaDomain.Schema)(nil)).Elem()
	mapType := reflect.TypeOf(map[string]any{})

	return (fromType == schemaType && toType == mapType) ||
		(fromType == mapType && toType == schemaType) ||
		(fromType == reflect.PointerTo(schemaType) && toType == mapType) ||
		(fromType == mapType && toType == reflect.PointerTo(schemaType))
}

func (c *SchemaConverter) CanReverse(fromType, toType reflect.Type) bool {
	return true // Schema conversions are bidirectional
}

func (c *SchemaConverter) Convert(from any, toType reflect.Type) (any, error) {
	schemaType := reflect.TypeOf((*schemaDomain.Schema)(nil)).Elem()
	schemaPtrType := reflect.PointerTo(schemaType)
	mapType := reflect.TypeOf(map[string]any{})

	fromType := reflect.TypeOf(from)

	// Schema -> map[string]any
	if (fromType == schemaType || fromType == schemaPtrType) && toType == mapType {
		return c.schemaToMap(from)
	}

	// map[string]any -> Schema
	if fromType == mapType && (toType == schemaType || toType == schemaPtrType) {
		return c.mapToSchema(from.(map[string]any), toType)
	}

	return nil, NewConversionError(fromType, toType, from, "unsupported schema conversion", nil)
}

func (c *SchemaConverter) schemaToMap(from any) (map[string]any, error) {
	var schema *schemaDomain.Schema

	switch v := from.(type) {
	case schemaDomain.Schema:
		schema = &v
	case *schemaDomain.Schema:
		schema = v
	default:
		return nil, NewConversionError(reflect.TypeOf(from), reflect.TypeOf(map[string]any{}),
			from, "not a schema type", nil)
	}

	if schema == nil {
		return map[string]any{}, nil
	}

	result := make(map[string]any)

	// Basic fields
	if schema.Type != "" {
		result["type"] = schema.Type
	}
	if schema.Title != "" {
		result["title"] = schema.Title
	}
	if schema.Description != "" {
		result["description"] = schema.Description
	}

	// Additional properties
	if schema.AdditionalProperties != nil {
		result["additionalProperties"] = *schema.AdditionalProperties
	}

	// Properties
	if len(schema.Properties) > 0 {
		properties := make(map[string]any)
		for name, prop := range schema.Properties {
			propMap, err := c.propertyToMap(prop)
			if err != nil {
				return nil, NewConversionError(reflect.TypeOf(from), reflect.TypeOf(map[string]any{}),
					from, "failed to convert property: "+name, err)
			}
			properties[name] = propMap
		}
		result["properties"] = properties
	}

	// Required fields
	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	// Conditional validation schemas
	if schema.If != nil {
		ifMap, err := c.schemaToMap(*schema.If)
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), reflect.TypeOf(map[string]any{}),
				from, "failed to convert if schema", err)
		}
		result["if"] = ifMap
	}

	if schema.Then != nil {
		thenMap, err := c.schemaToMap(*schema.Then)
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), reflect.TypeOf(map[string]any{}),
				from, "failed to convert then schema", err)
		}
		result["then"] = thenMap
	}

	if schema.Else != nil {
		elseMap, err := c.schemaToMap(*schema.Else)
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), reflect.TypeOf(map[string]any{}),
				from, "failed to convert else schema", err)
		}
		result["else"] = elseMap
	}

	// AllOf, AnyOf, OneOf
	if len(schema.AllOf) > 0 {
		allOf := make([]any, len(schema.AllOf))
		for i, subSchema := range schema.AllOf {
			subMap, err := c.schemaToMap(*subSchema)
			if err != nil {
				return nil, NewConversionError(reflect.TypeOf(from), reflect.TypeOf(map[string]any{}),
					from, "failed to convert allOf schema", err)
			}
			allOf[i] = subMap
		}
		result["allOf"] = allOf
	}

	if len(schema.AnyOf) > 0 {
		anyOf := make([]any, len(schema.AnyOf))
		for i, subSchema := range schema.AnyOf {
			subMap, err := c.schemaToMap(*subSchema)
			if err != nil {
				return nil, NewConversionError(reflect.TypeOf(from), reflect.TypeOf(map[string]any{}),
					from, "failed to convert anyOf schema", err)
			}
			anyOf[i] = subMap
		}
		result["anyOf"] = anyOf
	}

	if len(schema.OneOf) > 0 {
		oneOf := make([]any, len(schema.OneOf))
		for i, subSchema := range schema.OneOf {
			subMap, err := c.schemaToMap(*subSchema)
			if err != nil {
				return nil, NewConversionError(reflect.TypeOf(from), reflect.TypeOf(map[string]any{}),
					from, "failed to convert oneOf schema", err)
			}
			oneOf[i] = subMap
		}
		result["oneOf"] = oneOf
	}

	if schema.Not != nil {
		notMap, err := c.schemaToMap(*schema.Not)
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), reflect.TypeOf(map[string]any{}),
				from, "failed to convert not schema", err)
		}
		result["not"] = notMap
	}

	return result, nil
}

func (c *SchemaConverter) propertyToMap(prop schemaDomain.Property) (map[string]any, error) {
	result := make(map[string]any)

	if prop.Type != "" {
		result["type"] = prop.Type
	}
	if prop.Description != "" {
		result["description"] = prop.Description
	}
	if prop.Format != "" {
		result["format"] = prop.Format
	}

	// Numeric constraints
	if prop.Minimum != nil {
		result["minimum"] = *prop.Minimum
	}
	if prop.Maximum != nil {
		result["maximum"] = *prop.Maximum
	}
	if prop.ExclusiveMinimum != nil {
		result["exclusiveMinimum"] = *prop.ExclusiveMinimum
	}
	if prop.ExclusiveMaximum != nil {
		result["exclusiveMaximum"] = *prop.ExclusiveMaximum
	}

	// String constraints
	if prop.MinLength != nil {
		result["minLength"] = *prop.MinLength
	}
	if prop.MaxLength != nil {
		result["maxLength"] = *prop.MaxLength
	}
	if prop.Pattern != "" {
		result["pattern"] = prop.Pattern
	}

	// Array constraints
	if prop.MinItems != nil {
		result["minItems"] = *prop.MinItems
	}
	if prop.MaxItems != nil {
		result["maxItems"] = *prop.MaxItems
	}
	if prop.UniqueItems != nil {
		result["uniqueItems"] = *prop.UniqueItems
	}

	// Enum values (Property.Enum is []string)
	if len(prop.Enum) > 0 {
		result["enum"] = prop.Enum
	}

	// Nested properties
	if len(prop.Properties) > 0 {
		properties := make(map[string]any)
		for name, nestedProp := range prop.Properties {
			nestedPropMap, err := c.propertyToMap(nestedProp)
			if err != nil {
				return nil, err
			}
			properties[name] = nestedPropMap
		}
		result["properties"] = properties
	}

	// Required fields for nested objects
	if len(prop.Required) > 0 {
		result["required"] = prop.Required
	}

	// Additional properties
	if prop.AdditionalProperties != nil {
		result["additionalProperties"] = *prop.AdditionalProperties
	}

	// Items for arrays
	if prop.Items != nil {
		itemsMap, err := c.propertyToMap(*prop.Items)
		if err != nil {
			return nil, err
		}
		result["items"] = itemsMap
	}

	// Custom validator
	if prop.CustomValidator != "" {
		result["customValidator"] = prop.CustomValidator
	}

	return result, nil
}

func (c *SchemaConverter) mapToSchema(from map[string]any, toType reflect.Type) (any, error) {
	schema := &schemaDomain.Schema{}

	// Basic fields
	if v, ok := from["type"].(string); ok {
		schema.Type = v
	}
	if v, ok := from["title"].(string); ok {
		schema.Title = v
	}
	if v, ok := from["description"].(string); ok {
		schema.Description = v
	}

	// Additional properties
	if v, ok := from["additionalProperties"].(bool); ok {
		schema.AdditionalProperties = &v
	}

	// Properties
	if props, ok := from["properties"].(map[string]any); ok {
		schema.Properties = make(map[string]schemaDomain.Property)
		for name, propData := range props {
			if propMap, ok := propData.(map[string]any); ok {
				prop, err := c.mapToProperty(propMap)
				if err != nil {
					return nil, NewConversionError(reflect.TypeOf(from), toType,
						from, "failed to convert property: "+name, err)
				}
				schema.Properties[name] = prop
			}
		}
	}

	// Required fields
	if req, ok := from["required"].([]any); ok {
		schema.Required = make([]string, len(req))
		for i, item := range req {
			if str, ok := item.(string); ok {
				schema.Required[i] = str
			}
		}
	}

	// Conditional validation schemas
	if ifSchema, ok := from["if"].(map[string]any); ok {
		converted, err := c.mapToSchema(ifSchema, reflect.TypeOf(schemaDomain.Schema{}))
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), toType,
				from, "failed to convert if schema", err)
		}
		if schemaVal, ok := converted.(schemaDomain.Schema); ok {
			schema.If = &schemaVal
		}
	}

	if thenSchema, ok := from["then"].(map[string]any); ok {
		converted, err := c.mapToSchema(thenSchema, reflect.TypeOf(schemaDomain.Schema{}))
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), toType,
				from, "failed to convert then schema", err)
		}
		if schemaVal, ok := converted.(schemaDomain.Schema); ok {
			schema.Then = &schemaVal
		}
	}

	if elseSchema, ok := from["else"].(map[string]any); ok {
		converted, err := c.mapToSchema(elseSchema, reflect.TypeOf(schemaDomain.Schema{}))
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), toType,
				from, "failed to convert else schema", err)
		}
		if schemaVal, ok := converted.(schemaDomain.Schema); ok {
			schema.Else = &schemaVal
		}
	}

	// AllOf, AnyOf, OneOf, Not
	if allOf, ok := from["allOf"].([]any); ok {
		schema.AllOf = make([]*schemaDomain.Schema, len(allOf))
		for i, item := range allOf {
			if itemMap, ok := item.(map[string]any); ok {
				converted, err := c.mapToSchema(itemMap, reflect.TypeOf(schemaDomain.Schema{}))
				if err != nil {
					return nil, NewConversionError(reflect.TypeOf(from), toType,
						from, "failed to convert allOf schema", err)
				}
				if schemaVal, ok := converted.(schemaDomain.Schema); ok {
					schema.AllOf[i] = &schemaVal
				}
			}
		}
	}

	if anyOf, ok := from["anyOf"].([]any); ok {
		schema.AnyOf = make([]*schemaDomain.Schema, len(anyOf))
		for i, item := range anyOf {
			if itemMap, ok := item.(map[string]any); ok {
				converted, err := c.mapToSchema(itemMap, reflect.TypeOf(schemaDomain.Schema{}))
				if err != nil {
					return nil, NewConversionError(reflect.TypeOf(from), toType,
						from, "failed to convert anyOf schema", err)
				}
				if schemaVal, ok := converted.(schemaDomain.Schema); ok {
					schema.AnyOf[i] = &schemaVal
				}
			}
		}
	}

	if oneOf, ok := from["oneOf"].([]any); ok {
		schema.OneOf = make([]*schemaDomain.Schema, len(oneOf))
		for i, item := range oneOf {
			if itemMap, ok := item.(map[string]any); ok {
				converted, err := c.mapToSchema(itemMap, reflect.TypeOf(schemaDomain.Schema{}))
				if err != nil {
					return nil, NewConversionError(reflect.TypeOf(from), toType,
						from, "failed to convert oneOf schema", err)
				}
				if schemaVal, ok := converted.(schemaDomain.Schema); ok {
					schema.OneOf[i] = &schemaVal
				}
			}
		}
	}

	if notSchema, ok := from["not"].(map[string]any); ok {
		converted, err := c.mapToSchema(notSchema, reflect.TypeOf(schemaDomain.Schema{}))
		if err != nil {
			return nil, NewConversionError(reflect.TypeOf(from), toType,
				from, "failed to convert not schema", err)
		}
		if schemaVal, ok := converted.(schemaDomain.Schema); ok {
			schema.Not = &schemaVal
		}
	}

	// Return as value or pointer based on target type
	if toType.Kind() == reflect.Ptr {
		return schema, nil
	}
	return *schema, nil
}

func (c *SchemaConverter) mapToProperty(from map[string]any) (schemaDomain.Property, error) {
	prop := schemaDomain.Property{}

	if v, ok := from["type"].(string); ok {
		prop.Type = v
	}
	if v, ok := from["description"].(string); ok {
		prop.Description = v
	}
	if v, ok := from["format"].(string); ok {
		prop.Format = v
	}

	// Numeric constraints
	if v, ok := from["minimum"].(float64); ok {
		prop.Minimum = &v
	}
	if v, ok := from["maximum"].(float64); ok {
		prop.Maximum = &v
	}
	if v, ok := from["exclusiveMinimum"].(float64); ok {
		prop.ExclusiveMinimum = &v
	}
	if v, ok := from["exclusiveMaximum"].(float64); ok {
		prop.ExclusiveMaximum = &v
	}

	// String constraints
	if v, ok := from["minLength"].(float64); ok {
		intVal := int(v)
		prop.MinLength = &intVal
	}
	if v, ok := from["maxLength"].(float64); ok {
		intVal := int(v)
		prop.MaxLength = &intVal
	}
	if v, ok := from["pattern"].(string); ok {
		prop.Pattern = v
	}

	// Array constraints
	if v, ok := from["minItems"].(float64); ok {
		intVal := int(v)
		prop.MinItems = &intVal
	}
	if v, ok := from["maxItems"].(float64); ok {
		intVal := int(v)
		prop.MaxItems = &intVal
	}
	if v, ok := from["uniqueItems"].(bool); ok {
		prop.UniqueItems = &v
	}

	// Enum values (convert []any to []string)
	if enum, ok := from["enum"].([]any); ok {
		stringEnum := make([]string, len(enum))
		for i, item := range enum {
			if str, ok := item.(string); ok {
				stringEnum[i] = str
			} else {
				stringEnum[i] = fmt.Sprintf("%v", item)
			}
		}
		prop.Enum = stringEnum
	}

	// Additional properties
	if v, ok := from["additionalProperties"].(bool); ok {
		prop.AdditionalProperties = &v
	}

	// Custom validator
	if v, ok := from["customValidator"].(string); ok {
		prop.CustomValidator = v
	}

	// Nested properties
	if props, ok := from["properties"].(map[string]any); ok {
		prop.Properties = make(map[string]schemaDomain.Property)
		for name, propData := range props {
			if propMap, ok := propData.(map[string]any); ok {
				nestedProp, err := c.mapToProperty(propMap)
				if err != nil {
					return prop, err
				}
				prop.Properties[name] = nestedProp
			}
		}
	}

	// Required fields
	if req, ok := from["required"].([]any); ok {
		prop.Required = make([]string, len(req))
		for i, item := range req {
			if str, ok := item.(string); ok {
				prop.Required[i] = str
			}
		}
	}

	// Items for arrays
	if items, ok := from["items"].(map[string]any); ok {
		itemsProp, err := c.mapToProperty(items)
		if err != nil {
			return prop, err
		}
		prop.Items = &itemsProp
	}

	return prop, nil
}
