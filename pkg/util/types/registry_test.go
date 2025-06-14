// ABOUTME: Comprehensive tests for the type registry and conversion system
// ABOUTME: Tests core functionality, built-in converters, multi-hop conversion, and bridge scenarios

package types

import (
	"fmt"
	"reflect"
	"testing"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestRegistry_Basic(t *testing.T) {
	registry := NewRegistry()

	// Test empty registry
	if registry.CanConvert(reflect.TypeOf(""), reflect.TypeOf(0)) {
		t.Error("Empty registry should not be able to convert anything")
	}

	// Register a converter
	converter := &StringConverter{}
	err := registry.RegisterConverter(converter)
	if err != nil {
		t.Fatalf("Failed to register converter: %v", err)
	}

	// Test conversion
	result, err := registry.Convert("123", reflect.TypeOf(0))
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	if result != 123 {
		t.Errorf("Expected 123, got %v", result)
	}
}

func TestStringConverter(t *testing.T) {
	converter := &StringConverter{}

	tests := []struct {
		name     string
		input    any
		toType   reflect.Type
		expected any
		wantErr  bool
	}{
		{
			name:     "int to string",
			input:    42,
			toType:   reflect.TypeOf(""),
			expected: "42",
			wantErr:  false,
		},
		{
			name:     "string to int",
			input:    "42",
			toType:   reflect.TypeOf(0),
			expected: 42,
			wantErr:  false,
		},
		{
			name:     "float to string",
			input:    3.14,
			toType:   reflect.TypeOf(""),
			expected: "3.14",
			wantErr:  false,
		},
		{
			name:     "string to float",
			input:    "3.14",
			toType:   reflect.TypeOf(0.0),
			expected: 3.14,
			wantErr:  false,
		},
		{
			name:     "bool to string",
			input:    true,
			toType:   reflect.TypeOf(""),
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "string to bool",
			input:    "true",
			toType:   reflect.TypeOf(false),
			expected: true,
			wantErr:  false,
		},
		{
			name:    "invalid string to int",
			input:   "not a number",
			toType:  reflect.TypeOf(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert(tt.input, tt.toType)

			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.expected {
				t.Errorf("Convert() result = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNumberConverter(t *testing.T) {
	converter := &NumberConverter{}

	tests := []struct {
		name     string
		input    any
		toType   reflect.Type
		expected any
	}{
		{
			name:     "int to float64",
			input:    42,
			toType:   reflect.TypeOf(0.0),
			expected: 42.0,
		},
		{
			name:     "float64 to int",
			input:    42.7,
			toType:   reflect.TypeOf(0),
			expected: 42,
		},
		{
			name:     "int32 to int64",
			input:    int32(42),
			toType:   reflect.TypeOf(int64(0)),
			expected: int64(42),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert(tt.input, tt.toType)
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("Convert() result = %v (type %T), expected %v (type %T)",
					result, result, tt.expected, tt.expected)
			}
		})
	}
}

func TestSliceConverter(t *testing.T) {
	converter := &SliceConverter{}

	// Test slice to []any
	input := []int{1, 2, 3}
	result, err := converter.Convert(input, reflect.TypeOf([]any{}))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	expected := []any{1, 2, 3}
	resultSlice := result.([]any)

	if len(resultSlice) != len(expected) {
		t.Errorf("Result slice length = %v, expected %v", len(resultSlice), len(expected))
	}

	for i, v := range expected {
		if resultSlice[i] != v {
			t.Errorf("Result[%d] = %v, expected %v", i, resultSlice[i], v)
		}
	}
}

func TestMapConverter(t *testing.T) {
	converter := &MapConverter{}

	// Test map to map[string]any
	input := map[string]int{"a": 1, "b": 2}
	result, err := converter.Convert(input, reflect.TypeOf(map[string]any{}))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	resultMap := result.(map[string]any)

	if resultMap["a"] != 1 || resultMap["b"] != 2 {
		t.Errorf("Unexpected map conversion result: %v", resultMap)
	}
}

func TestSchemaConverter(t *testing.T) {
	converter := &SchemaConverter{}

	// Create test schema
	schema := schemaDomain.Schema{
		Type:        "object",
		Title:       "Test Schema",
		Description: "A test schema",
		Properties: map[string]schemaDomain.Property{
			"name": {
				Type:        "string",
				Description: "Name field",
			},
			"age": {
				Type:        "integer",
				Description: "Age field",
				Minimum:     floatPtr(0),
			},
		},
		Required: []string{"name"},
	}

	// Test Schema -> map[string]any
	result, err := converter.Convert(schema, reflect.TypeOf(map[string]any{}))
	if err != nil {
		t.Fatalf("Schema to map conversion failed: %v", err)
	}

	resultMap := result.(map[string]any)

	if resultMap["type"] != "object" {
		t.Errorf("Expected type 'object', got %v", resultMap["type"])
	}

	if resultMap["title"] != "Test Schema" {
		t.Errorf("Expected title 'Test Schema', got %v", resultMap["title"])
	}

	properties, ok := resultMap["properties"].(map[string]any)
	if !ok {
		t.Fatal("Properties not converted correctly")
	}

	nameProperty, ok := properties["name"].(map[string]any)
	if !ok {
		t.Fatal("Name property not converted correctly")
	}

	if nameProperty["type"] != "string" {
		t.Errorf("Expected name property type 'string', got %v", nameProperty["type"])
	}

	// Test map[string]any -> Schema
	backToSchema, err := converter.Convert(resultMap, reflect.TypeOf(schemaDomain.Schema{}))
	if err != nil {
		t.Fatalf("Map to schema conversion failed: %v", err)
	}

	convertedSchema := backToSchema.(schemaDomain.Schema)

	if convertedSchema.Type != "object" {
		t.Errorf("Expected type 'object', got %v", convertedSchema.Type)
	}

	if convertedSchema.Title != "Test Schema" {
		t.Errorf("Expected title 'Test Schema', got %v", convertedSchema.Title)
	}

	if len(convertedSchema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(convertedSchema.Properties))
	}
}

func TestDefaultRegistry(t *testing.T) {
	registry := GetDefaultRegistry()

	// Test basic conversions
	tests := []struct {
		name     string
		input    any
		toType   reflect.Type
		expected any
	}{
		{
			name:     "string to int",
			input:    "42",
			toType:   reflect.TypeOf(0),
			expected: 42,
		},
		{
			name:     "int to string",
			input:    42,
			toType:   reflect.TypeOf(""),
			expected: "42",
		},
		{
			name:     "slice to []any",
			input:    []int{1, 2, 3},
			toType:   reflect.TypeOf([]any{}),
			expected: []any{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := registry.Convert(tt.input, tt.toType)
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}

			// For slice comparison, check elements individually
			if reflect.TypeOf(tt.expected).Kind() == reflect.Slice {
				expectedSlice := reflect.ValueOf(tt.expected)
				resultSlice := reflect.ValueOf(result)

				if expectedSlice.Len() != resultSlice.Len() {
					t.Errorf("Slice length mismatch: expected %d, got %d",
						expectedSlice.Len(), resultSlice.Len())
					return
				}

				for i := 0; i < expectedSlice.Len(); i++ {
					if expectedSlice.Index(i).Interface() != resultSlice.Index(i).Interface() {
						t.Errorf("Slice element %d: expected %v, got %v",
							i, expectedSlice.Index(i).Interface(), resultSlice.Index(i).Interface())
					}
				}
			} else {
				if result != tt.expected {
					t.Errorf("Convert() result = %v, expected %v", result, tt.expected)
				}
			}
		})
	}
}

func TestConversionCache(t *testing.T) {
	cache := NewConversionCache()

	// Test cache miss
	_, _, found := cache.Get("test-key")
	if found {
		t.Error("Expected cache miss for new key")
	}

	// Test cache set and hit
	cache.Set("test-key", "test-value", nil)
	value, err, found := cache.Get("test-key")
	if !found {
		t.Error("Expected cache hit")
	}

	if value != "test-value" {
		t.Errorf("Expected 'test-value', got %v", value)
	}

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test cache stats
	hits, misses, size := cache.Stats()
	if hits != 1 {
		t.Errorf("Expected 1 hit, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("Expected 1 miss, got %d", misses)
	}
	if size != 1 {
		t.Errorf("Expected cache size 1, got %d", size)
	}

	// Test cache clear
	cache.Clear()
	hits, misses, size = cache.Stats()
	if hits != 0 || misses != 0 || size != 0 {
		t.Error("Cache not properly cleared")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Test Convert function
	result, err := Convert("42", reflect.TypeOf(0))
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	if result != 42 {
		t.Errorf("Expected 42, got %v", result)
	}

	// Test ConvertTo function
	converted, err := ConvertTo[int]("42")
	if err != nil {
		t.Fatalf("ConvertTo() error = %v", err)
	}

	if converted != 42 {
		t.Errorf("Expected 42, got %v", converted)
	}

	// Test CanConvert function
	if !CanConvert(reflect.TypeOf(""), reflect.TypeOf(0)) {
		t.Error("Expected CanConvert to return true for string->int")
	}
}

func TestMultiHopConversion(t *testing.T) {
	registry := NewRegistry(WithMultiHop(true), WithMaxHops(2))

	// Register converters that enable multi-hop conversion
	err := registry.RegisterConverter(&StringConverter{})
	if err != nil {
		t.Fatalf("Failed to register StringConverter: %v", err)
	}
	err = registry.RegisterConverter(&JSONConverter{})
	if err != nil {
		t.Fatalf("Failed to register JSONConverter: %v", err)
	}

	// Test conversion that requires multiple hops
	// complex struct -> JSON string -> target type
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	input := TestStruct{Name: "John", Age: 30}
	result, err := registry.Convert(input, reflect.TypeOf(map[string]any{}))

	if err != nil {
		t.Fatalf("Multi-hop conversion failed: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", resultMap["name"])
	}
}

func TestErrorHandling(t *testing.T) {
	registry := NewRegistry()

	// Test conversion with no registered converters
	_, err := registry.Convert("test", reflect.TypeOf(0))
	if err == nil {
		t.Error("Expected error for conversion with no converters")
	}

	convErr, ok := err.(*ConversionError)
	if !ok {
		t.Errorf("Expected ConversionError, got %T", err)
	}

	if convErr.FromType != reflect.TypeOf("") {
		t.Errorf("Expected string type, got %v", convErr.FromType)
	}

	if convErr.ToType != reflect.TypeOf(0) {
		t.Errorf("Expected int type, got %v", convErr.ToType)
	}
}

func TestConcurrentAccess(t *testing.T) {
	registry := GetDefaultRegistry()

	// Test concurrent conversions
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(n int) {
			defer func() { done <- true }()

			for j := 0; j < 100; j++ {
				input := fmt.Sprintf("%d", n*100+j)
				_, err := registry.Convert(input, reflect.TypeOf(0))
				if err != nil {
					t.Errorf("Concurrent conversion failed: %v", err)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Helper function
func floatPtr(f float64) *float64 {
	return &f
}
