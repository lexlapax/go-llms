package generator

// ABOUTME: Tag-based schema generator that prioritizes struct tags
// ABOUTME: Supports multiple tag formats and custom validation rules

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// TagSchemaGenerator generates schemas primarily from struct tags
type TagSchemaGenerator struct {
	// Tag priority order
	tagPriority []string

	// Custom tag parsers
	tagParsers map[string]TagParser

	// Validation extractor
	validationExtractor ValidationExtractor
}

// TagParser parses a specific tag and updates the property
type TagParser func(tagValue string, prop *domain.Property) error

// ValidationExtractor extracts validation rules from tags
type ValidationExtractor func(tags reflect.StructTag) []ValidationRule

// ValidationRule represents an extracted validation rule
type ValidationRule struct {
	Type  string
	Value interface{}
}

// NewTagSchemaGenerator creates a new tag-based generator
func NewTagSchemaGenerator() *TagSchemaGenerator {
	g := &TagSchemaGenerator{
		tagPriority: []string{"schema", "json", "validate", "binding"},
		tagParsers:  make(map[string]TagParser),
	}

	// Register default tag parsers
	g.registerDefaultParsers()

	// Set default validation extractor
	g.validationExtractor = g.defaultValidationExtractor

	return g
}

// SetTagPriority sets the order in which tags are processed
func (g *TagSchemaGenerator) SetTagPriority(tags []string) {
	g.tagPriority = tags
}

// RegisterTagParser registers a custom tag parser
func (g *TagSchemaGenerator) RegisterTagParser(tag string, parser TagParser) {
	g.tagParsers[tag] = parser
}

// SetValidationExtractor sets a custom validation extractor
func (g *TagSchemaGenerator) SetValidationExtractor(extractor ValidationExtractor) {
	g.validationExtractor = extractor
}

// GenerateSchema generates a JSON schema from a Go type using tags
func (g *TagSchemaGenerator) GenerateSchema(obj interface{}) (*domain.Schema, error) {
	t := reflect.TypeOf(obj)
	if t == nil {
		return nil, fmt.Errorf("cannot generate schema for nil")
	}

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Only handle structs
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("schema generation only supports structs, got %s", t.Kind())
	}

	schema := &domain.Schema{
		Type:       "object",
		Properties: make(map[string]domain.Property),
	}

	// Extract schema-level metadata from type
	g.extractSchemaMetadata(t, schema)

	var required []string

	// Process each field
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name and check if it should be included
		name, include := g.getFieldName(field)
		if !include {
			continue
		}

		// Generate property from tags
		prop, err := g.generatePropertyFromTags(field)
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

	return schema, nil
}

// extractSchemaMetadata extracts schema-level metadata from type and tags
func (g *TagSchemaGenerator) extractSchemaMetadata(t reflect.Type, schema *domain.Schema) {
	// Use type name as default title
	schema.Title = t.Name()

	// Look for schema tags on the type itself (if supported in future Go versions)
	// For now, we'll just use the type information
	if t.PkgPath() != "" {
		schema.Description = fmt.Sprintf("Generated from %s.%s", t.PkgPath(), t.Name())
	}
}

// getFieldName extracts the field name from tags
func (g *TagSchemaGenerator) getFieldName(field reflect.StructField) (string, bool) {
	// Check tags in priority order
	for _, tagName := range g.tagPriority {
		if tagValue := field.Tag.Get(tagName); tagValue != "" {
			switch tagName {
			case "json":
				parts := strings.Split(tagValue, ",")
				if parts[0] == "-" {
					return "", false
				}
				if parts[0] != "" {
					return parts[0], true
				}
			case "schema":
				for _, part := range strings.Split(tagValue, ",") {
					if part == "-" {
						return "", false
					}
					if strings.HasPrefix(part, "name=") {
						return strings.TrimPrefix(part, "name="), true
					}
				}
			}
		}
	}

	return field.Name, true
}

// isFieldRequired checks if a field is required based on tags
func (g *TagSchemaGenerator) isFieldRequired(field reflect.StructField) bool {
	// Check each tag type
	for _, tagName := range g.tagPriority {
		tagValue := field.Tag.Get(tagName)
		if tagValue == "" {
			continue
		}

		switch tagName {
		case "validate", "binding":
			if strings.Contains(tagValue, "required") {
				return true
			}
		case "json":
			// If json tag exists and doesn't have omitempty, it's required
			if !strings.Contains(tagValue, "omitempty") {
				return true
			}
			return false
		case "schema":
			if strings.Contains(tagValue, "required") {
				return true
			}
		}
	}

	return false
}

// generatePropertyFromTags creates a property by parsing all relevant tags
func (g *TagSchemaGenerator) generatePropertyFromTags(field reflect.StructField) (domain.Property, error) {
	prop := domain.Property{}

	// First, infer basic type from Go type
	g.inferBasicType(field.Type, &prop)

	// Then, process tags in priority order
	for _, tagName := range g.tagPriority {
		if tagValue := field.Tag.Get(tagName); tagValue != "" {
			if parser, ok := g.tagParsers[tagName]; ok {
				if err := parser(tagValue, &prop); err != nil {
					return prop, fmt.Errorf("error parsing %s tag: %w", tagName, err)
				}
			}
		}
	}

	// Process custom tags not in priority list
	for tagName, parser := range g.tagParsers {
		// Skip if already in priority list
		inPriority := false
		for _, priorityTag := range g.tagPriority {
			if tagName == priorityTag {
				inPriority = true
				break
			}
		}
		if !inPriority {
			if tagValue := field.Tag.Get(tagName); tagValue != "" {
				if err := parser(tagValue, &prop); err != nil {
					return prop, fmt.Errorf("error parsing %s tag: %w", tagName, err)
				}
			}
		}
	}

	// Extract validation rules
	if g.validationExtractor != nil {
		rules := g.validationExtractor(field.Tag)
		g.applyValidationRules(&prop, rules)
	}

	// Process description from various tags
	g.extractDescription(field.Tag, &prop)

	return prop, nil
}

// inferBasicType infers the basic JSON schema type from Go type
func (g *TagSchemaGenerator) inferBasicType(t reflect.Type, prop *domain.Property) {
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		prop.Type = "string"
	case reflect.Bool:
		prop.Type = "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		prop.Type = "integer"
	case reflect.Float32, reflect.Float64:
		prop.Type = "number"
	case reflect.Slice, reflect.Array:
		prop.Type = "array"
	case reflect.Map:
		prop.Type = "object"
		boolTrue := true
		prop.AdditionalProperties = &boolTrue
	case reflect.Struct:
		prop.Type = "object"
	default:
		prop.Type = "string" // Default fallback
	}
}

// registerDefaultParsers registers the default tag parsers
func (g *TagSchemaGenerator) registerDefaultParsers() {
	// Schema tag parser
	g.tagParsers["schema"] = func(tagValue string, prop *domain.Property) error {
		for _, part := range strings.Split(tagValue, ",") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				continue
			}

			key, value := kv[0], kv[1]
			switch key {
			case "type":
				prop.Type = value
			case "format":
				prop.Format = value
			case "pattern":
				prop.Pattern = value
			case "minLength":
				if v, err := strconv.Atoi(value); err == nil {
					prop.MinLength = intPtr(v)
				}
			case "maxLength":
				if v, err := strconv.Atoi(value); err == nil {
					prop.MaxLength = intPtr(v)
				}
			case "minimum":
				if v, err := strconv.ParseFloat(value, 64); err == nil {
					prop.Minimum = float64Ptr(v)
				}
			case "maximum":
				if v, err := strconv.ParseFloat(value, 64); err == nil {
					prop.Maximum = float64Ptr(v)
				}
			case "minItems":
				if v, err := strconv.Atoi(value); err == nil {
					prop.MinItems = intPtr(v)
				}
			case "maxItems":
				if v, err := strconv.Atoi(value); err == nil {
					prop.MaxItems = intPtr(v)
				}
			case "enum":
				prop.Enum = strings.Split(value, "|")
			case "description":
				prop.Description = value
			}
		}
		return nil
	}

	// Validate tag parser
	g.tagParsers["validate"] = func(tagValue string, prop *domain.Property) error {
		for _, rule := range strings.Split(tagValue, ",") {
			rule = strings.TrimSpace(rule)

			// Handle parameterized rules
			if idx := strings.Index(rule, "="); idx > 0 {
				ruleName := rule[:idx]
				ruleValue := rule[idx+1:]

				switch ruleName {
				case "min":
					if v, err := strconv.ParseFloat(ruleValue, 64); err == nil {
						prop.Minimum = float64Ptr(v)
					}
				case "max":
					if v, err := strconv.ParseFloat(ruleValue, 64); err == nil {
						prop.Maximum = float64Ptr(v)
					}
				case "len":
					if v, err := strconv.Atoi(ruleValue); err == nil {
						prop.MinLength = intPtr(v)
						prop.MaxLength = intPtr(v)
					}
				case "oneof":
					prop.Enum = strings.Split(ruleValue, " ")
				}
			} else {
				// Handle simple rules
				switch rule {
				case "email":
					prop.Format = "email"
				case "url", "uri":
					prop.Format = "uri"
				case "uuid":
					prop.Format = "uuid"
				case "date":
					prop.Format = "date"
				case "datetime":
					prop.Format = "date-time"
				}
			}
		}
		return nil
	}

	// JSON tag parser (mainly for metadata)
	g.tagParsers["json"] = func(tagValue string, prop *domain.Property) error {
		// JSON tag doesn't directly affect schema properties
		// but we process it for consistency
		return nil
	}

	// Format tag parser
	g.tagParsers["format"] = func(tagValue string, prop *domain.Property) error {
		prop.Format = tagValue
		return nil
	}

	// Pattern tag parser
	g.tagParsers["pattern"] = func(tagValue string, prop *domain.Property) error {
		prop.Pattern = tagValue
		return nil
	}
}

// defaultValidationExtractor extracts validation rules from common tags
func (g *TagSchemaGenerator) defaultValidationExtractor(tags reflect.StructTag) []ValidationRule {
	var rules []ValidationRule

	// Extract from validate tag
	if validate := tags.Get("validate"); validate != "" {
		for _, rule := range strings.Split(validate, ",") {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			vr := ValidationRule{}
			if idx := strings.Index(rule, "="); idx > 0 {
				vr.Type = rule[:idx]
				vr.Value = rule[idx+1:]
			} else {
				vr.Type = rule
			}
			rules = append(rules, vr)
		}
	}

	// Extract from binding tag (common in web frameworks)
	if binding := tags.Get("binding"); binding != "" {
		for _, rule := range strings.Split(binding, ",") {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			vr := ValidationRule{Type: rule}
			rules = append(rules, vr)
		}
	}

	return rules
}

// applyValidationRules applies extracted validation rules to the property
func (g *TagSchemaGenerator) applyValidationRules(prop *domain.Property, rules []ValidationRule) {
	for _, rule := range rules {
		switch rule.Type {
		case "required":
			// Handled at schema level
		case "email":
			prop.Format = "email"
		case "url", "uri":
			prop.Format = "uri"
		case "min":
			if v, err := strconv.ParseFloat(rule.Value.(string), 64); err == nil {
				prop.Minimum = float64Ptr(v)
			}
		case "max":
			if v, err := strconv.ParseFloat(rule.Value.(string), 64); err == nil {
				prop.Maximum = float64Ptr(v)
			}
		case "len":
			if v, err := strconv.Atoi(rule.Value.(string)); err == nil {
				prop.MinLength = intPtr(v)
				prop.MaxLength = intPtr(v)
			}
		case "minlen":
			if v, err := strconv.Atoi(rule.Value.(string)); err == nil {
				prop.MinLength = intPtr(v)
			}
		case "maxlen":
			if v, err := strconv.Atoi(rule.Value.(string)); err == nil {
				prop.MaxLength = intPtr(v)
			}
		}
	}
}

// extractDescription extracts description from various tags
func (g *TagSchemaGenerator) extractDescription(tags reflect.StructTag, prop *domain.Property) {
	// Check tags in order of preference
	descTags := []string{"description", "desc", "doc", "comment"}
	for _, tag := range descTags {
		if desc := tags.Get(tag); desc != "" {
			prop.Description = desc
			return
		}
	}

	// Check schema tag for description
	if schema := tags.Get("schema"); schema != "" {
		for _, part := range strings.Split(schema, ",") {
			if strings.HasPrefix(part, "description=") {
				prop.Description = strings.TrimPrefix(part, "description=")
				return
			}
		}
	}
}
