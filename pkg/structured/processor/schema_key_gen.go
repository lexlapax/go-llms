package processor

import (
	"hash"
	"hash/fnv"
	"sort"
	"strconv"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// ImprovedGenerateSchemaKey creates a hash key for a schema with better coverage of schema structures
// This handles conditional schemas, enums, nested structures, and is less sensitive to property order
func ImprovedGenerateSchemaKey(schema *schemaDomain.Schema) uint64 {
	hasher := fnv.New64a() // Use fnv-1a for better distribution and speed

	// Hash schema type
	writeString(hasher, schema.Type)

	// Hash title, description and additionalProperties
	writeString(hasher, schema.Title)
	writeString(hasher, schema.Description)
	if schema.AdditionalProperties != nil {
		writeBool(hasher, *schema.AdditionalProperties)
	}

	// Hash required fields (sorted for consistent ordering)
	if len(schema.Required) > 0 {
		// Sort for consistent key generation regardless of order
		sortedRequired := make([]string, len(schema.Required))
		copy(sortedRequired, schema.Required)
		sort.Strings(sortedRequired)

		for _, req := range sortedRequired {
			writeString(hasher, req)
		}
	}

	// Hash properties (in sorted key order for consistency)
	if len(schema.Properties) > 0 {
		// Get keys in a sorted order
		keys := make([]string, 0, len(schema.Properties))
		for k := range schema.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Process properties in sorted key order
		for _, k := range keys {
			prop := schema.Properties[k]
			writeString(hasher, k) // Property name
			hashProperty(hasher, &prop)
		}
	}

	// Hash conditional schemas
	hashConditionalSchema(hasher, schema)

	return hasher.Sum64()
}

// hashProperty adds a property's fields to the hash
func hashProperty(hasher Hashtable64, prop *schemaDomain.Property) {
	// Hash basic property fields
	writeString(hasher, prop.Type)
	writeString(hasher, prop.Format)
	writeString(hasher, prop.Description)
	writeString(hasher, prop.Pattern)
	writeString(hasher, prop.CustomValidator)

	// Hash numerical constraints
	if prop.Minimum != nil {
		writeString(hasher, strconv.FormatFloat(*prop.Minimum, 'f', -1, 64))
	}
	if prop.Maximum != nil {
		writeString(hasher, strconv.FormatFloat(*prop.Maximum, 'f', -1, 64))
	}
	if prop.ExclusiveMinimum != nil {
		writeString(hasher, strconv.FormatFloat(*prop.ExclusiveMinimum, 'f', -1, 64))
	}
	if prop.ExclusiveMaximum != nil {
		writeString(hasher, strconv.FormatFloat(*prop.ExclusiveMaximum, 'f', -1, 64))
	}

	// Hash integer constraints
	if prop.MinLength != nil {
		writeString(hasher, strconv.Itoa(*prop.MinLength))
	}
	if prop.MaxLength != nil {
		writeString(hasher, strconv.Itoa(*prop.MaxLength))
	}
	if prop.MinItems != nil {
		writeString(hasher, strconv.Itoa(*prop.MinItems))
	}
	if prop.MaxItems != nil {
		writeString(hasher, strconv.Itoa(*prop.MaxItems))
	}
	if prop.UniqueItems != nil {
		writeBool(hasher, *prop.UniqueItems)
	}

	// Hash enum values (sorted for consistent ordering)
	if len(prop.Enum) > 0 {
		// Sort for consistent key generation regardless of order
		sortedEnum := make([]string, len(prop.Enum))
		copy(sortedEnum, prop.Enum)
		sort.Strings(sortedEnum)

		for _, e := range sortedEnum {
			writeString(hasher, e)
		}
	}

	// Hash array item definition
	if prop.Items != nil {
		// Special marker to indicate presence of items
		writeString(hasher, "has_items")
		hashProperty(hasher, prop.Items)
	}

	// Hash nested properties
	if len(prop.Properties) > 0 {
		// Get keys in a sorted order
		keys := make([]string, 0, len(prop.Properties))
		for k := range prop.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Process nested properties in sorted key order
		for _, k := range keys {
			nestedProp := prop.Properties[k]
			writeString(hasher, k) // Property name
			hashProperty(hasher, &nestedProp)
		}
	}

	// Hash required fields for object properties
	if len(prop.Required) > 0 {
		// Sort for consistent key generation regardless of order
		sortedRequired := make([]string, len(prop.Required))
		copy(sortedRequired, prop.Required)
		sort.Strings(sortedRequired)

		for _, req := range sortedRequired {
			writeString(hasher, req)
		}
	}

	// Hash additionalProperties
	if prop.AdditionalProperties != nil {
		writeBool(hasher, *prop.AdditionalProperties)
	}

	// Hash conditional schema elements
	if len(prop.AnyOf) > 0 {
		writeString(hasher, "anyof")
		for _, subSchema := range prop.AnyOf {
			if subSchema != nil {
				hashSchema(hasher, subSchema)
			}
		}
	}

	if len(prop.OneOf) > 0 {
		writeString(hasher, "oneof")
		for _, subSchema := range prop.OneOf {
			if subSchema != nil {
				hashSchema(hasher, subSchema)
			}
		}
	}

	if prop.Not != nil {
		writeString(hasher, "not")
		hashSchema(hasher, prop.Not)
	}
}

// hashConditionalSchema adds conditional validation schemas to the hash
func hashConditionalSchema(hasher Hashtable64, schema *schemaDomain.Schema) {
	// Handle If/Then/Else
	if schema.If != nil {
		writeString(hasher, "if")
		hashSchema(hasher, schema.If)
	}
	if schema.Then != nil {
		writeString(hasher, "then")
		hashSchema(hasher, schema.Then)
	}
	if schema.Else != nil {
		writeString(hasher, "else")
		hashSchema(hasher, schema.Else)
	}

	// Handle AllOf, AnyOf, OneOf, Not
	if len(schema.AllOf) > 0 {
		writeString(hasher, "allof")
		for _, subSchema := range schema.AllOf {
			if subSchema != nil {
				hashSchema(hasher, subSchema)
			}
		}
	}

	if len(schema.AnyOf) > 0 {
		writeString(hasher, "anyof")
		for _, subSchema := range schema.AnyOf {
			if subSchema != nil {
				hashSchema(hasher, subSchema)
			}
		}
	}

	if len(schema.OneOf) > 0 {
		writeString(hasher, "oneof")
		for _, subSchema := range schema.OneOf {
			if subSchema != nil {
				hashSchema(hasher, subSchema)
			}
		}
	}

	if schema.Not != nil {
		writeString(hasher, "not")
		hashSchema(hasher, schema.Not)
	}
}

// hashSchema recursively adds schema values to the hash
func hashSchema(hasher Hashtable64, schema *schemaDomain.Schema) {
	// Add schema type
	writeString(hasher, schema.Type)

	// Add title and description
	writeString(hasher, schema.Title)
	writeString(hasher, schema.Description)

	// Add additionalProperties
	if schema.AdditionalProperties != nil {
		writeBool(hasher, *schema.AdditionalProperties)
	}

	// Add required fields (sorted for consistent ordering)
	if len(schema.Required) > 0 {
		// Sort for consistent key generation regardless of order
		sortedRequired := make([]string, len(schema.Required))
		copy(sortedRequired, schema.Required)
		sort.Strings(sortedRequired)

		for _, req := range sortedRequired {
			writeString(hasher, req)
		}
	}

	// Add properties (in sorted key order for consistency)
	if len(schema.Properties) > 0 {
		// Get keys in a sorted order
		keys := make([]string, 0, len(schema.Properties))
		for k := range schema.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Process properties in sorted key order
		for _, k := range keys {
			prop := schema.Properties[k]
			writeString(hasher, k) // Property name
			hashProperty(hasher, &prop)
		}
	}

	// Recursively add conditional schema elements
	hashConditionalSchema(hasher, schema)
}

// Helper functions to write different types to the hasher
func writeString(hasher Hashtable64, s string) {
	if s != "" {
		hasher.Write([]byte(s))
		// Add a null byte as a separator to prevent collisions
		hasher.Write([]byte{0})
	}
}

func writeBool(hasher Hashtable64, b bool) {
	if b {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
}

// Hashtable64 interface to make testing easier
type Hashtable64 interface {
	hash.Hash
	Sum64() uint64
}