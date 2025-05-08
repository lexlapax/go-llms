// Package tools provides implementations of agent tools.
package tools

import (
	// "fmt" - Commented out as we disabled debug prints
	"reflect"
	"strings"
	"sync"
)

// parameterTypeCache caches reflection type information to reduce allocations
// during repeated tool executions with the same parameter types
type parameterTypeCache struct {
	// structFieldCache maps struct types to field information to avoid repeated reflection lookups
	structFieldCache sync.Map // map[reflect.Type][]fieldInfo

	// parameterConversionCache caches common conversion patterns
	parameterConversionCache sync.Map // map[typePair]bool
}

// typePair is a key for the conversion cache
type typePair struct {
	source reflect.Type
	target reflect.Type
}

// fieldInfo caches information about a struct field
type fieldInfo struct {
	index      int
	name       string
	jsonName   string
	fieldType  reflect.Type
	canSet     bool
	isExported bool
}

// globalParamCache is a shared instance of the parameter cache
var globalParamCache = &parameterTypeCache{}

// getStructFields returns cached field information for a struct type
func (c *parameterTypeCache) getStructFields(structType reflect.Type) []fieldInfo {
	if structType.Kind() != reflect.Struct {
		return nil
	}

	// Check if we already have cached this type
	if cachedFields, ok := c.structFieldCache.Load(structType); ok {
		return cachedFields.([]fieldInfo)
	}

	// Build field information
	numFields := structType.NumField()
	fields := make([]fieldInfo, 0, numFields)

	for i := 0; i < numFields; i++ {
		field := structType.Field(i)
		jsonName := field.Name

		// Extract JSON tag name if present
		if tag := field.Tag.Get("json"); tag != "" {
			// Parse the tag to extract the name part
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				jsonName = parts[0]
			}
		}

		fields = append(fields, fieldInfo{
			index:      i,
			name:       field.Name,
			jsonName:   jsonName,
			fieldType:  field.Type,
			canSet:     true, // Will be checked during actual mapping
			isExported: field.PkgPath == "",
		})
	}

	// Debug information (commented for performance)
	/*
		fmt.Printf("DEBUG: Struct type %v fields:\n", structType)
		for _, f := range fields {
			fmt.Printf("  Field: %s, JSON: %s, Index: %d, Exported: %v\n",
				f.name, f.jsonName, f.index, f.isExported)
		}
	*/

	// Cache and return
	c.structFieldCache.Store(structType, fields)
	return fields
}

// canConvert checks if a type can be converted to another type
func (c *parameterTypeCache) canConvert(sourceType, targetType reflect.Type) bool {
	// Direct assignability is fastest
	if sourceType.AssignableTo(targetType) {
		return true
	}

	// Check cache for this conversion pair
	pair := typePair{sourceType, targetType}
	if cached, ok := c.parameterConversionCache.Load(pair); ok {
		return cached.(bool)
	}

	// Determine convertibility based on type characteristics
	canConvert := false

	// String destination type can accept almost anything
	if targetType.Kind() == reflect.String {
		canConvert = true
	} else if targetType.Kind() == reflect.Int || targetType.Kind() == reflect.Int64 {
		// Various numeric types can convert to int
		switch sourceType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String:
			canConvert = true
		}
	} else if targetType.Kind() == reflect.Float32 || targetType.Kind() == reflect.Float64 {
		// Various numeric types can convert to float
		switch sourceType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String:
			canConvert = true
		}
	} else if targetType.Kind() == reflect.Bool {
		// Various types can convert to bool
		switch sourceType.Kind() {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String:
			canConvert = true
		}
	} else if targetType.Kind() == reflect.Slice && sourceType.Kind() == reflect.Slice {
		// Slices can be converted if their elements can be converted
		canConvert = c.canConvert(sourceType.Elem(), targetType.Elem())
	} else if targetType.Kind() == reflect.Map && sourceType.Kind() == reflect.Map {
		// Maps can be converted if their keys and values can be converted
		canConvert = c.canConvert(sourceType.Key(), targetType.Key()) &&
			c.canConvert(sourceType.Elem(), targetType.Elem())
	}

	// Cache the result
	c.parameterConversionCache.Store(pair, canConvert)
	return canConvert
}
