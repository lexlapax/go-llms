// Package generator provides schema generation implementations
package generator

// ABOUTME: Enhanced reflection-based schema generator with advanced features
// ABOUTME: Supports custom types, type preservation, and complex struct handling

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ReflectionSchemaGenerator generates schemas from Go structs using reflection
type ReflectionSchemaGenerator struct {
	// Options for generation
	preserveGoTypes bool
	includePrivate  bool
	maxDepth        int

	// Cache for recursive types
	typeCache map[reflect.Type]*domain.Schema
	cacheMu   sync.RWMutex

	// Custom type handlers
	customHandlers map[reflect.Type]TypeHandler
}

// TypeHandler defines custom handling for specific types
type TypeHandler func(t reflect.Type, tag reflect.StructTag) (domain.Property, error)

// NewReflectionSchemaGenerator creates a new reflection-based generator
func NewReflectionSchemaGenerator() *ReflectionSchemaGenerator {
	return &ReflectionSchemaGenerator{
		preserveGoTypes: true,
		includePrivate:  false,
		maxDepth:        10,
		typeCache:       make(map[reflect.Type]*domain.Schema),
		customHandlers:  make(map[reflect.Type]TypeHandler),
	}
}

// WithOptions configures the generator
func (g *ReflectionSchemaGenerator) WithOptions(preserveGoTypes, includePrivate bool, maxDepth int) *ReflectionSchemaGenerator {
	g.preserveGoTypes = preserveGoTypes
	g.includePrivate = includePrivate
	g.maxDepth = maxDepth
	return g
}

// RegisterTypeHandler registers a custom handler for a specific type
func (g *ReflectionSchemaGenerator) RegisterTypeHandler(t reflect.Type, handler TypeHandler) {
	g.customHandlers[t] = handler
}

// GenerateSchema generates a JSON schema from a Go type
func (g *ReflectionSchemaGenerator) GenerateSchema(obj interface{}) (*domain.Schema, error) {
	t := reflect.TypeOf(obj)
	if t == nil {
		return nil, fmt.Errorf("cannot generate schema for nil")
	}

	// Reset cache for new generation
	g.cacheMu.Lock()
	g.typeCache = make(map[reflect.Type]*domain.Schema)
	g.cacheMu.Unlock()

	return g.generateSchemaForType(t, reflect.StructTag(""), 0)
}

// generateSchemaForType generates schema for a specific type with depth tracking
func (g *ReflectionSchemaGenerator) generateSchemaForType(t reflect.Type, tags reflect.StructTag, depth int) (*domain.Schema, error) {
	if depth > g.maxDepth {
		return nil, fmt.Errorf("max depth %d exceeded", g.maxDepth)
	}

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check cache for recursive types
	g.cacheMu.RLock()
	if cached, ok := g.typeCache[t]; ok && depth < g.maxDepth {
		g.cacheMu.RUnlock()
		return cached, nil
	}
	g.cacheMu.RUnlock()

	// Only handle structs at the top level
	if depth == 0 && t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("schema generation only supports structs at top level, got %s", t.Kind())
	}

	if t.Kind() == reflect.Struct {
		schema := &domain.Schema{
			Type:       "object",
			Properties: make(map[string]domain.Property),
		}

		// Cache the schema early to handle recursive types
		g.cacheMu.Lock()
		g.typeCache[t] = schema
		g.cacheMu.Unlock()

		// Add type information if preserving Go types
		if g.preserveGoTypes {
			schema.Title = t.Name()
			if t.PkgPath() != "" {
				schema.Description = fmt.Sprintf("Go type: %s.%s", t.PkgPath(), t.Name())
			}
		}

		// Extract schema-level tags
		if schemaTag := tags.Get("schema"); schemaTag != "" {
			// Parse schema-level attributes
			for _, part := range strings.Split(schemaTag, ",") {
				if strings.HasPrefix(part, "title=") {
					schema.Title = strings.TrimPrefix(part, "title=")
				} else if strings.HasPrefix(part, "description=") {
					schema.Description = strings.TrimPrefix(part, "description=")
				}
			}
		}

		var required []string

		// Process each field
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Skip unexported fields unless includePrivate is true
			if !field.IsExported() && !g.includePrivate {
				continue
			}

			// Get field name
			name := g.getFieldName(field)
			if name == "-" {
				continue // Skip fields with json:"-"
			}

			// Generate property for field
			prop, err := g.generateProperty(field, depth+1)
			if err != nil {
				return nil, fmt.Errorf("error generating property for field %s: %w", field.Name, err)
			}

			schema.Properties[name] = prop

			// Check if required
			if g.isFieldRequired(field) {
				required = append(required, name)
			}
		}

		if len(required) > 0 {
			schema.Required = required
		}

		// Handle additional properties
		if addProp := tags.Get("additionalProperties"); addProp != "" {
			val := addProp == "true"
			schema.AdditionalProperties = &val
		}

		return schema, nil
	}

	// For non-struct types at depth > 0, convert to property then to schema
	prop, err := g.generatePropertyForType(t, tags, depth)
	if err != nil {
		return nil, err
	}

	// Convert property to schema
	schema := &domain.Schema{
		Type: prop.Type,
	}

	// Copy relevant fields
	if prop.Format != "" {
		// Schemas don't have format at top level, but we can add it to description
		schema.Description = fmt.Sprintf("Format: %s", prop.Format)
	}

	return schema, nil
}

// getFieldName extracts the field name from tags
func (g *ReflectionSchemaGenerator) getFieldName(field reflect.StructField) string {
	// Check JSON tag first
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}

	// Check schema tag
	if schemaTag := field.Tag.Get("schema"); schemaTag != "" {
		for _, part := range strings.Split(schemaTag, ",") {
			if strings.HasPrefix(part, "name=") {
				return strings.TrimPrefix(part, "name=")
			}
		}
	}

	return field.Name
}

// isFieldRequired checks if a field is required
func (g *ReflectionSchemaGenerator) isFieldRequired(field reflect.StructField) bool {
	// Check validate tag
	if validateTag := field.Tag.Get("validate"); strings.Contains(validateTag, "required") {
		return true
	}

	// Check JSON tag for omitempty
	if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
		return !strings.Contains(jsonTag, "omitempty")
	}

	// Check schema tag
	if schemaTag := field.Tag.Get("schema"); strings.Contains(schemaTag, "required") {
		return true
	}

	return false
}

// generateProperty creates a property from a struct field
func (g *ReflectionSchemaGenerator) generateProperty(field reflect.StructField, depth int) (domain.Property, error) {
	return g.generatePropertyForType(field.Type, field.Tag, depth)
}

// generatePropertyForType creates a property for a specific type
func (g *ReflectionSchemaGenerator) generatePropertyForType(t reflect.Type, tags reflect.StructTag, depth int) (domain.Property, error) {
	prop := domain.Property{}

	// Check for custom handler
	if handler, ok := g.customHandlers[t]; ok {
		return handler(t, tags)
	}

	// Set description from tags
	if desc := tags.Get("description"); desc != "" {
		prop.Description = desc
	} else if desc := tags.Get("doc"); desc != "" {
		prop.Description = desc
	}

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Process based on kind
	switch t.Kind() {
	case reflect.String:
		prop.Type = "string"
		g.processStringTags(&prop, tags)

	case reflect.Bool:
		prop.Type = "boolean"

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		prop.Type = "integer"
		g.processNumberTags(&prop, tags)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		prop.Type = "integer"
		g.processNumberTags(&prop, tags)
		// Add minimum constraint for unsigned
		if prop.Minimum == nil {
			prop.Minimum = float64Ptr(0)
		}

	case reflect.Float32, reflect.Float64:
		prop.Type = "number"
		g.processNumberTags(&prop, tags)

	case reflect.Slice, reflect.Array:
		prop.Type = "array"
		g.processArrayTags(&prop, tags)

		// Generate schema for elements
		elemProp, err := g.generatePropertyForType(t.Elem(), reflect.StructTag(""), depth)
		if err != nil {
			return prop, fmt.Errorf("error generating array element schema: %w", err)
		}
		prop.Items = &elemProp

	case reflect.Map:
		prop.Type = "object"
		boolTrue := true
		prop.AdditionalProperties = &boolTrue

		// If map value is a known type, we could add pattern properties
		// For now, we'll just allow any additional properties

	case reflect.Struct:
		// Handle special types
		switch t {
		case reflect.TypeOf(time.Time{}):
			prop.Type = "string"
			prop.Format = "date-time"
		case reflect.TypeOf(json.RawMessage{}):
			prop.Type = "object"
			boolTrue := true
			prop.AdditionalProperties = &boolTrue
		default:
			// For nested structs, check if we're at max depth first
			if depth >= g.maxDepth {
				// Max depth reached, use generic object
				prop.Type = "object"
				boolTrue := true
				prop.AdditionalProperties = &boolTrue
				// Don't set properties for max depth
				prop.Properties = nil
			} else {
				// Generate nested schema
				nestedSchema, err := g.generateSchemaForType(t, tags, depth)
				if err != nil {
					return prop, fmt.Errorf("error generating nested struct schema: %w", err)
				}
				prop.Type = "object"
				prop.Properties = nestedSchema.Properties
				prop.Required = nestedSchema.Required
				prop.AdditionalProperties = nestedSchema.AdditionalProperties
			}
		}

	case reflect.Interface:
		// Interfaces can be anything
		prop.Type = "object"
		boolTrue := true
		prop.AdditionalProperties = &boolTrue

	default:
		// Unknown types default to string
		prop.Type = "string"
		if g.preserveGoTypes {
			prop.Description = fmt.Sprintf("Go type: %s", t.String())
		}
	}

	// Process custom validator tag
	if customValidator := tags.Get("customValidator"); customValidator != "" {
		prop.CustomValidator = customValidator
	}

	return prop, nil
}

// processStringTags processes string-specific tags
func (g *ReflectionSchemaGenerator) processStringTags(prop *domain.Property, tags reflect.StructTag) {
	// Format
	if format := tags.Get("format"); format != "" {
		prop.Format = format
	}

	// Pattern
	if pattern := tags.Get("pattern"); pattern != "" {
		prop.Pattern = pattern
	}

	// Length constraints
	if minLen := tags.Get("minLength"); minLen != "" {
		if min, err := strconv.Atoi(minLen); err == nil {
			prop.MinLength = intPtr(min)
		}
	}

	if maxLen := tags.Get("maxLength"); maxLen != "" {
		if max, err := strconv.Atoi(maxLen); err == nil {
			prop.MaxLength = intPtr(max)
		}
	}

	// Enum from validate tag
	g.processEnumTag(prop, tags)

	// Email validation
	if validateTag := tags.Get("validate"); strings.Contains(validateTag, "email") {
		prop.Format = "email"
	}
}

// processNumberTags processes number-specific tags
func (g *ReflectionSchemaGenerator) processNumberTags(prop *domain.Property, tags reflect.StructTag) {
	// Minimum
	if min := tags.Get("minimum"); min != "" {
		if minVal, err := strconv.ParseFloat(min, 64); err == nil {
			prop.Minimum = float64Ptr(minVal)
		}
	}

	// Maximum
	if max := tags.Get("maximum"); max != "" {
		if maxVal, err := strconv.ParseFloat(max, 64); err == nil {
			prop.Maximum = float64Ptr(maxVal)
		}
	}

	// Exclusive minimum
	if exMin := tags.Get("exclusiveMinimum"); exMin != "" {
		if minVal, err := strconv.ParseFloat(exMin, 64); err == nil {
			prop.ExclusiveMinimum = float64Ptr(minVal)
		}
	}

	// Exclusive maximum
	if exMax := tags.Get("exclusiveMaximum"); exMax != "" {
		if maxVal, err := strconv.ParseFloat(exMax, 64); err == nil {
			prop.ExclusiveMaximum = float64Ptr(maxVal)
		}
	}

	// Process validate tag for min/max
	if validateTag := tags.Get("validate"); validateTag != "" {
		for _, part := range strings.Split(validateTag, ",") {
			if strings.HasPrefix(part, "min=") {
				if min, err := strconv.ParseFloat(strings.TrimPrefix(part, "min="), 64); err == nil {
					prop.Minimum = float64Ptr(min)
				}
			} else if strings.HasPrefix(part, "max=") {
				if max, err := strconv.ParseFloat(strings.TrimPrefix(part, "max="), 64); err == nil {
					prop.Maximum = float64Ptr(max)
				}
			}
		}
	}
}

// processArrayTags processes array-specific tags
func (g *ReflectionSchemaGenerator) processArrayTags(prop *domain.Property, tags reflect.StructTag) {
	// Min items
	if minItems := tags.Get("minItems"); minItems != "" {
		if min, err := strconv.Atoi(minItems); err == nil {
			prop.MinItems = intPtr(min)
		}
	}

	// Max items
	if maxItems := tags.Get("maxItems"); maxItems != "" {
		if max, err := strconv.Atoi(maxItems); err == nil {
			prop.MaxItems = intPtr(max)
		}
	}

	// Unique items
	if unique := tags.Get("uniqueItems"); unique == "true" {
		prop.UniqueItems = boolPtr(true)
	}
}

// processEnumTag processes enum values from tags
func (g *ReflectionSchemaGenerator) processEnumTag(prop *domain.Property, tags reflect.StructTag) {
	// Check enum tag
	if enum := tags.Get("enum"); enum != "" {
		prop.Enum = strings.Split(enum, ",")
		return
	}

	// Check validate tag for oneof
	if validateTag := tags.Get("validate"); strings.Contains(validateTag, "oneof=") {
		for _, part := range strings.Split(validateTag, ",") {
			if strings.HasPrefix(part, "oneof=") {
				values := strings.TrimPrefix(part, "oneof=")
				prop.Enum = strings.Split(values, " ")
				break
			}
		}
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}
