// ABOUTME: Example demonstrating bridge-friendly type system for go-llmspell integration
// ABOUTME: Shows how to convert between Go types and scripting language types seamlessly

package main

import (
	"fmt"
	"reflect"
	"time"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
	"github.com/lexlapax/go-llms/pkg/util/types"
)

func main() {
	fmt.Println("=== Bridge-Friendly Type System Example ===")

	// Get the default registry with all built-in converters
	registry := types.GetDefaultRegistry()

	// Example 1: Basic Type Conversions
	fmt.Println("\n1. Basic Type Conversions:")
	demonstrateBasicConversions(registry)

	// Example 2: Schema Conversions (Critical for Bridge Layer)
	fmt.Println("\n2. Schema ↔ Map Conversions:")
	demonstrateSchemaConversions(registry)

	// Example 3: Complex Data Structure Bridging
	fmt.Println("\n3. Complex Data Structure Bridging:")
	demonstrateComplexBridging(registry)

	// Example 4: Multi-hop Conversions
	fmt.Println("\n4. Multi-hop Conversions:")
	demonstrateMultiHopConversions(registry)

	// Example 5: Custom Converter Registration
	fmt.Println("\n5. Custom Converter Registration:")
	demonstrateCustomConverter(registry)

	// Example 6: Performance and Caching
	fmt.Println("\n6. Performance and Caching:")
	demonstratePerformance(registry)

	// Example 7: Bridge Simulation (go-llmspell scenario)
	fmt.Println("\n7. Bridge Simulation (go-llmspell scenario):")
	simulateBridgeScenario(registry)
}

func demonstrateBasicConversions(registry *types.Registry) {
	conversions := []struct {
		name   string
		input  any
		target reflect.Type
	}{
		{"String to Int", "42", reflect.TypeOf(0)},
		{"Int to String", 42, reflect.TypeOf("")},
		{"Float to Int", 3.14, reflect.TypeOf(0)},
		{"String to Float", "3.14", reflect.TypeOf(0.0)},
		{"String to Bool", "true", reflect.TypeOf(false)},
		{"Bool to String", true, reflect.TypeOf("")},
	}

	for _, conv := range conversions {
		result, err := registry.Convert(conv.input, conv.target)
		if err != nil {
			fmt.Printf("  ❌ %s: %v\n", conv.name, err)
		} else {
			fmt.Printf("  ✅ %s: %v (%T) → %v (%T)\n",
				conv.name, conv.input, conv.input, result, result)
		}
	}
}

func demonstrateSchemaConversions(registry *types.Registry) {
	// Create a schema (typical in go-llms)
	schema := schemaDomain.Schema{
		Type:        "object",
		Title:       "User Profile",
		Description: "A user profile schema",
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type:        "string",
				Description: "User's full name",
				MinLength:   intPtr(1),
				MaxLength:   intPtr(100),
			},
			"age": {
				Type:        "integer",
				Description: "User's age",
				Minimum:     floatPtr(0),
				Maximum:     floatPtr(150),
			},
			"email": {
				Type:        "string",
				Description: "User's email address",
				Format:      "email",
			},
			"active": {
				Type:        "boolean",
				Description: "Whether the user is active",
			},
		},
		Required: []string{"name", "email"},
	}

	fmt.Println("  Original Schema:")
	fmt.Printf("    Type: %s, Title: %s\n", schema.Type, schema.Title)
	fmt.Printf("    Properties: %d, Required: %v\n", len(schema.Properties), schema.Required)

	// Convert Schema → map[string]any (for scripting layer)
	schemaMap, err := registry.Convert(schema, reflect.TypeOf(map[string]any{}))
	if err != nil {
		fmt.Printf("  ❌ Schema to map conversion failed: %v\n", err)
		return
	}

	fmt.Println("  ✅ Schema → map[string]any conversion successful")
	mapData := schemaMap.(map[string]any)
	fmt.Printf("    Map keys: %d\n", len(mapData))

	// Show some map contents
	if props, ok := mapData["properties"].(map[string]any); ok {
		fmt.Printf("    Properties in map: %v\n", getMapKeys(props))
	}

	// Convert map[string]any → Schema (from scripting layer back)
	backToSchema, err := registry.Convert(mapData, reflect.TypeOf(schemaDomain.Schema{}))
	if err != nil {
		fmt.Printf("  ❌ Map to schema conversion failed: %v\n", err)
		return
	}

	fmt.Println("  ✅ map[string]any → Schema conversion successful")
	convertedSchema := backToSchema.(schemaDomain.Schema)
	fmt.Printf("    Converted Type: %s, Title: %s\n", convertedSchema.Type, convertedSchema.Title)
	fmt.Printf("    Properties: %d, Required: %v\n", len(convertedSchema.Properties), convertedSchema.Required)
}

func demonstrateComplexBridging(registry *types.Registry) {
	// Simulate complex data that might come from a scripting language
	scriptData := map[string]any{
		"users": []any{
			map[string]any{
				"name": "Alice",
				"age":  30.0, // JSON numbers are float64
				"tags": []any{"admin", "active"},
			},
			map[string]any{
				"name": "Bob",
				"age":  25.0,
				"tags": []any{"user", "active"},
			},
		},
		"config": map[string]any{
			"debug":   true,
			"version": "1.0",
			"timeout": 30.0,
		},
	}

	fmt.Printf("  Original script data: %d top-level keys\n", len(scriptData))

	// Convert to Go-friendly types
	results := make(map[string]any)

	// Convert users array
	if users, ok := scriptData["users"]; ok {
		convertedUsers, err := registry.Convert(users, reflect.TypeOf([]any{}))
		if err != nil {
			fmt.Printf("  ❌ Users conversion failed: %v\n", err)
		} else {
			results["users"] = convertedUsers
			fmt.Printf("  ✅ Users array converted: %d users\n", len(convertedUsers.([]any)))
		}
	}

	// Convert config map
	if config, ok := scriptData["config"]; ok {
		convertedConfig, err := registry.Convert(config, reflect.TypeOf(map[string]any{}))
		if err != nil {
			fmt.Printf("  ❌ Config conversion failed: %v\n", err)
		} else {
			results["config"] = convertedConfig
			fmt.Printf("  ✅ Config map converted: %d settings\n", len(convertedConfig.(map[string]any)))
		}
	}

	// Demonstrate nested conversions
	if users, ok := results["users"].([]any); ok && len(users) > 0 {
		if firstUser, ok := users[0].(map[string]any); ok {
			if age, ok := firstUser["age"]; ok {
				// Convert float64 age to int
				intAge, err := registry.Convert(age, reflect.TypeOf(0))
				if err != nil {
					fmt.Printf("  ❌ Age conversion failed: %v\n", err)
				} else {
					fmt.Printf("  ✅ Age converted: %v (%T) → %v (%T)\n",
						age, age, intAge, intAge)
				}
			}
		}
	}
}

func demonstrateMultiHopConversions(registry *types.Registry) {
	// Define a struct that needs multi-hop conversion
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	person := Person{Name: "Charlie", Age: 35}

	fmt.Printf("  Original struct: %+v\n", person)

	// Convert struct → map[string]any (via JSON)
	personMap, err := registry.Convert(person, reflect.TypeOf(map[string]any{}))
	if err != nil {
		fmt.Printf("  ❌ Struct to map conversion failed: %v\n", err)
		return
	}

	fmt.Println("  ✅ Struct → map[string]any (multi-hop via JSON)")
	mapData := personMap.(map[string]any)
	fmt.Printf("    Map: %v\n", mapData)

	// Convert map[string]any → back to struct
	backToPerson, err := registry.Convert(mapData, reflect.TypeOf(Person{}))
	if err != nil {
		fmt.Printf("  ❌ Map to struct conversion failed: %v\n", err)
		return
	}

	fmt.Println("  ✅ map[string]any → Struct (multi-hop via JSON)")
	convertedPerson := backToPerson.(Person)
	fmt.Printf("    Struct: %+v\n", convertedPerson)
}

// TimeConverter handles conversions between string and time.Time
type TimeConverter struct{}

func (c *TimeConverter) Name() string  { return "TimeConverter" }
func (c *TimeConverter) Priority() int { return 150 }

func (c *TimeConverter) CanConvert(fromType, toType reflect.Type) bool {
	return fromType.Kind() == reflect.String && toType.String() == "time.Time"
}

func (c *TimeConverter) CanReverse(fromType, toType reflect.Type) bool {
	return fromType.String() == "time.Time" && toType.Kind() == reflect.String
}

func (c *TimeConverter) Convert(from any, toType reflect.Type) (any, error) {
	if str, ok := from.(string); ok && toType.String() == "time.Time" {
		// Simple time parsing
		return time.Parse("2006-01-02", str)
	}
	return nil, types.NewConversionError(reflect.TypeOf(from), toType, from, "unsupported time conversion", nil)
}

func demonstrateCustomConverter(registry *types.Registry) {

	// Register the custom converter
	err := registry.RegisterConverter(&TimeConverter{})
	if err != nil {
		fmt.Printf("  ❌ Failed to register custom converter: %v\n", err)
		return
	}

	fmt.Println("  ✅ Custom TimeConverter registered")

	// Test the custom converter
	timeStr := "2023-12-25"
	timeValue, err := registry.Convert(timeStr, reflect.TypeOf(time.Time{}))
	if err != nil {
		fmt.Printf("  ❌ Custom conversion failed: %v\n", err)
	} else {
		fmt.Printf("  ✅ String → time.Time: %s → %v\n", timeStr, timeValue)
	}
}

func demonstratePerformance(registry *types.Registry) {
	// Clear cache to start fresh
	registry.ClearCache()

	// Perform some conversions to populate cache
	conversions := []string{"1", "2", "3", "1", "2", "3"} // Repeat to test cache

	for i, str := range conversions {
		_, err := registry.Convert(str, reflect.TypeOf(0))
		if err != nil {
			fmt.Printf("  ❌ Conversion %d failed: %v\n", i, err)
		}
	}

	// Check cache statistics
	hits, misses, size := registry.GetCacheStats()
	fmt.Printf("  Cache Stats: %d hits, %d misses, %d entries\n", hits, misses, size)

	if hits > 0 {
		fmt.Printf("  ✅ Cache hit ratio: %.2f%%\n", float64(hits)/float64(hits+misses)*100)
	}

	// List registered converters
	converters := registry.ListConverters()
	fmt.Printf("  Registered converters: %d\n", len(converters))
	for _, conv := range converters {
		fmt.Printf("    - %s (priority: %d)\n", conv.Name(), conv.Priority())
	}
}

func simulateBridgeScenario(registry *types.Registry) {
	fmt.Println("  Simulating go-llmspell bridge scenario...")

	// 1. Schema comes from Go side
	schema := schemaDomain.Schema{
		Type: "object",
		Properties: map[string]schemaDomain.Property{
			"query": {
				Type:        "string",
				Description: "Search query",
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum results",
			},
		},
		Required: []string{"query"},
	}

	// 2. Convert to map for script bridge
	scriptSchema, err := registry.Convert(schema, reflect.TypeOf(map[string]any{}))
	if err != nil {
		fmt.Printf("  ❌ Go→Script schema conversion failed: %v\n", err)
		return
	}
	fmt.Println("  ✅ Go Schema → Script Map")

	// 3. Script modifies the schema (simulated)
	scriptMap := scriptSchema.(map[string]any)

	// Add a new property from script side
	if props, ok := scriptMap["properties"].(map[string]any); ok {
		props["timestamp"] = map[string]any{
			"type":        "string",
			"description": "Added by script",
			"format":      "date-time",
		}
	}

	// 4. Convert back to Go schema
	modifiedSchema, err := registry.Convert(scriptMap, reflect.TypeOf(schemaDomain.Schema{}))
	if err != nil {
		fmt.Printf("  ❌ Script→Go schema conversion failed: %v\n", err)
		return
	}
	fmt.Println("  ✅ Script Map → Go Schema")

	finalSchema := modifiedSchema.(schemaDomain.Schema)
	fmt.Printf("  Original properties: %d, Final properties: %d\n",
		len(schema.Properties), len(finalSchema.Properties))

	// 5. Data flow simulation
	scriptData := map[string]any{
		"query":     "search term",
		"limit":     5.0, // JSON number
		"timestamp": "2023-12-25T10:00:00Z",
	}

	// Convert script data to Go-friendly format
	convertedData := make(map[string]any)
	for key, value := range scriptData {
		switch key {
		case "limit":
			// Convert float64 to int
			intValue, err := registry.Convert(value, reflect.TypeOf(0))
			if err != nil {
				fmt.Printf("  ❌ Limit conversion failed: %v\n", err)
			} else {
				convertedData[key] = intValue
			}
		default:
			convertedData[key] = value
		}
	}

	fmt.Printf("  ✅ Bridge data conversion complete: %v\n", convertedData)
	fmt.Println("  🎉 Bridge scenario simulation successful!")
}

// Helper functions
func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
