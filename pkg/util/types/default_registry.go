// ABOUTME: Default global type registry with pre-registered converters for common use cases
// ABOUTME: Provides ready-to-use type conversion capabilities for bridge layers and go-llmspell integration

package types

import (
	"reflect"
	"sync"
)

var (
	// DefaultRegistry is the global type converter registry
	DefaultRegistry *Registry
	registryOnce    sync.Once
)

// GetDefaultRegistry returns the global default registry with all built-in converters
func GetDefaultRegistry() *Registry {
	registryOnce.Do(func() {
		DefaultRegistry = NewRegistry(
			WithCache(true),
			WithMaxHops(3),
			WithMultiHop(true),
		)

		// Register all built-in converters
		registerBuiltinConverters(DefaultRegistry)
	})
	return DefaultRegistry
}

// registerBuiltinConverters registers all built-in type converters
func registerBuiltinConverters(registry *Registry) {
	converters := []TypeConverter{
		// High priority converters for specific types
		&SchemaConverter{}, // Priority 200 - Critical for bridge layer

		// Medium priority converters for common types
		&StringConverter{}, // Priority 100
		&NumberConverter{}, // Priority 95
		&SliceConverter{},  // Priority 90
		&MapConverter{},    // Priority 85

		// Low priority fallback converter
		&JSONConverter{}, // Priority 50 - Fallback through JSON
	}

	for _, converter := range converters {
		if err := registry.RegisterConverter(converter); err != nil {
			// In a production system, you might want to log this error
			// For now, we'll silently continue to avoid dependency on logging
			continue
		}
	}
}

// Convert is a convenience function that uses the default registry
func Convert(from any, toType any) (any, error) {
	registry := GetDefaultRegistry()

	var targetType reflect.Type
	switch t := toType.(type) {
	case reflect.Type:
		targetType = t
	default:
		targetType = reflect.TypeOf(toType)
	}

	return registry.Convert(from, targetType)
}

// ConvertTo is a generic convenience function for type conversion
func ConvertTo[T any](from any) (T, error) {
	var zero T
	targetType := reflect.TypeOf(zero)

	registry := GetDefaultRegistry()
	result, err := registry.Convert(from, targetType)
	if err != nil {
		return zero, err
	}

	if converted, ok := result.(T); ok {
		return converted, nil
	}

	return zero, NewConversionError(reflect.TypeOf(from), targetType, from, "type assertion failed", nil)
}

// CanConvert checks if conversion is possible using the default registry
func CanConvert(fromType, toType reflect.Type) bool {
	registry := GetDefaultRegistry()
	return registry.CanConvert(fromType, toType)
}

// CanReverse checks if reverse conversion is possible using the default registry
func CanReverse(fromType, toType reflect.Type) bool {
	registry := GetDefaultRegistry()
	return registry.CanReverse(fromType, toType)
}

// RegisterConverter adds a converter to the default registry
func RegisterConverter(converter TypeConverter) error {
	registry := GetDefaultRegistry()
	return registry.RegisterConverter(converter)
}

// ClearCache clears the conversion cache in the default registry
func ClearCache() {
	registry := GetDefaultRegistry()
	registry.ClearCache()
}

// GetCacheStats returns cache statistics from the default registry
func GetCacheStats() (hits, misses uint64, size int) {
	registry := GetDefaultRegistry()
	return registry.GetCacheStats()
}
