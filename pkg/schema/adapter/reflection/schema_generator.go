package reflection

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Helper function for creating float pointers
func float64Ptr(v float64) *float64 {
	return &v
}

// Helper function for creating int pointers
func intPtr(v int) *int {
	return &v
}

// GenerateSchema creates a schema from a Go struct
func GenerateSchema(obj interface{}) (*domain.Schema, error) {
	t := reflect.TypeOf(obj)
	
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
	
	var required []string
	
	// Process each field
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Get JSON field name or use struct field name
		jsonTag := field.Tag.Get("json")
		name := field.Name
		omitempty := false
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				name = parts[0]
			}
			// Check for omitempty
			for _, opt := range parts[1:] {
				if opt == "omitempty" {
					omitempty = true
					break
				}
			}
		}
		
		// Create property
		prop, err := generateProperty(field)
		if err != nil {
			return nil, fmt.Errorf("error generating property for field %s: %w", field.Name, err)
		}
		
		schema.Properties[name] = prop
		
		// Check if required
		if tag := field.Tag.Get("validate"); strings.Contains(tag, "required") && !omitempty {
			required = append(required, name)
		}
	}
	
	if len(required) > 0 {
		schema.Required = required
	}
	
	return schema, nil
}

// generateProperty creates a property from a struct field
func generateProperty(field reflect.StructField) (domain.Property, error) {
	prop := domain.Property{}
	
	// Set description if available
	if desc := field.Tag.Get("description"); desc != "" {
		prop.Description = desc
	}
	
	// Map Go types to JSON Schema types
	t := field.Type
	
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		prop.Type = "string"
		
		// Handle string formats
		if format := field.Tag.Get("format"); format != "" {
			prop.Format = format
		}
		
		// Handle email validation
		if validateTag := field.Tag.Get("validate"); strings.Contains(validateTag, "email") {
			prop.Format = "email"
		}
		
		// Handle pattern
		if pattern := field.Tag.Get("pattern"); pattern != "" {
			prop.Pattern = pattern
		}
		
		// Handle min/max length
		if minLength := field.Tag.Get("minLength"); minLength != "" {
			if min, err := strconv.Atoi(minLength); err == nil {
				prop.MinLength = intPtr(min)
			}
		}
		
		if maxLength := field.Tag.Get("maxLength"); maxLength != "" {
			if max, err := strconv.Atoi(maxLength); err == nil {
				prop.MaxLength = intPtr(max)
			}
		}
		
		// Handle enum values from validate tag
		if validateTag := field.Tag.Get("validate"); strings.Contains(validateTag, "oneof=") {
			for _, part := range strings.Split(validateTag, ",") {
				if strings.HasPrefix(part, "oneof=") {
					values := strings.TrimPrefix(part, "oneof=")
					prop.Enum = strings.Split(values, " ")
					break
				}
			}
		}
		
	case reflect.Bool:
		prop.Type = "boolean"
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		prop.Type = "integer"
		
		// Handle min/max from validate tag
		validateTag := field.Tag.Get("validate")
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
		
	case reflect.Float32, reflect.Float64:
		prop.Type = "number"
		
		// Handle min/max from validate tag
		validateTag := field.Tag.Get("validate")
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
		
	case reflect.Slice, reflect.Array:
		prop.Type = "array"
		
		// Create items schema based on slice element type
		elemType := t.Elem()
		var elemProp domain.Property
		
		// Handle pointer element type
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
		
		switch elemType.Kind() {
		case reflect.String:
			elemProp = domain.Property{Type: "string"}
		case reflect.Bool:
			elemProp = domain.Property{Type: "boolean"}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			elemProp = domain.Property{Type: "integer"}
		case reflect.Float32, reflect.Float64:
			elemProp = domain.Property{Type: "number"}
		case reflect.Struct:
			// For structs, recursively generate a schema
			if elemType == reflect.TypeOf(time.Time{}) {
				elemProp = domain.Property{Type: "string", Format: "date-time"}
			} else {
				// Create a zero value of the struct to pass to GenerateSchema
				elemValue := reflect.New(elemType).Elem().Interface()
				elemSchema, err := GenerateSchema(elemValue)
				if err != nil {
					return prop, fmt.Errorf("error generating schema for slice element: %w", err)
				}
				
				elemProp = domain.Property{
					Type:       "object",
					Properties: elemSchema.Properties,
					Required:   elemSchema.Required,
				}
			}
		default:
			elemProp = domain.Property{Type: "string"} // Default to string for unknown types
		}
		
		prop.Items = &elemProp
		
	case reflect.Struct:
		// Handle special cases
		if t == reflect.TypeOf(time.Time{}) {
			prop.Type = "string"
			prop.Format = "date-time"
		} else {
			// For other structs, recursively generate a schema
			structValue := reflect.New(t).Elem().Interface()
			structSchema, err := GenerateSchema(structValue)
			if err != nil {
				return prop, fmt.Errorf("error generating schema for nested struct: %w", err)
			}
			
			prop.Type = "object"
			prop.Properties = structSchema.Properties
			prop.Required = structSchema.Required
		}
		
	case reflect.Map:
		// Handle maps as objects with additional properties
		prop.Type = "object"
		
		// Get the value type for the map
		valueType := t.Elem()
		
		// Handle pointer value type
		if valueType.Kind() == reflect.Ptr {
			valueType = valueType.Elem()
		}
		
		// Set additionalProperties flag
		boolTrue := true
		prop.AdditionalProperties = &boolTrue
		
	default:
		// For unknown types, default to string
		prop.Type = "string"
	}
	
	return prop, nil
}