// ABOUTME: Type Registry for managing and organizing type converters with multi-hop conversion support
// ABOUTME: Central registry that enables bridge layers to register custom converters and perform complex type transformations

package types

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
)

// Registry manages a collection of type converters with multi-hop conversion support
type Registry struct {
	converters []TypeConverter
	cache      *ConversionCache
	mu         sync.RWMutex

	// Configuration
	enableCache    bool
	maxHops        int
	enableMultiHop bool
}

// RegistryOption configures the registry
type RegistryOption func(*Registry)

// WithCache enables caching for improved performance
func WithCache(enabled bool) RegistryOption {
	return func(r *Registry) {
		r.enableCache = enabled
	}
}

// WithMaxHops sets the maximum number of hops for multi-hop conversion
func WithMaxHops(maxHops int) RegistryOption {
	return func(r *Registry) {
		r.maxHops = maxHops
	}
}

// WithMultiHop enables or disables multi-hop conversion
func WithMultiHop(enabled bool) RegistryOption {
	return func(r *Registry) {
		r.enableMultiHop = enabled
	}
}

// NewRegistry creates a new type converter registry
func NewRegistry(options ...RegistryOption) *Registry {
	registry := &Registry{
		converters:     make([]TypeConverter, 0),
		cache:          NewConversionCache(),
		enableCache:    true,
		maxHops:        3,
		enableMultiHop: true,
	}

	for _, option := range options {
		option(registry)
	}

	return registry
}

// RegisterConverter adds a new type converter to the registry
func (r *Registry) RegisterConverter(converter TypeConverter) error {
	if converter == nil {
		return fmt.Errorf("converter cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.converters = append(r.converters, converter)

	// Sort converters by priority (highest first)
	sort.Slice(r.converters, func(i, j int) bool {
		return r.converters[i].Priority() > r.converters[j].Priority()
	})

	// Clear cache when converters change
	if r.enableCache {
		r.cache.Clear()
	}

	return nil
}

// Convert performs type conversion using registered converters
func (r *Registry) Convert(from any, toType reflect.Type) (any, error) {
	if from == nil {
		return r.convertNil(toType), nil
	}

	fromType := reflect.TypeOf(from)

	// Check if already the target type
	if fromType == toType {
		return from, nil
	}

	// Check cache first
	if r.enableCache {
		cacheKey := generateCacheKey(from, toType)
		if result, err, found := r.cache.Get(cacheKey); found {
			return result, err
		}
	}

	// Try direct conversion
	result, err := r.tryDirectConversion(from, fromType, toType)
	if err == nil {
		if r.enableCache {
			cacheKey := generateCacheKey(from, toType)
			r.cache.Set(cacheKey, result, nil)
		}
		return result, nil
	}

	// Try multi-hop conversion if enabled
	if r.enableMultiHop {
		result, err := r.tryMultiHopConversion(from, fromType, toType)
		if err == nil {
			if r.enableCache {
				cacheKey := generateCacheKey(from, toType)
				r.cache.Set(cacheKey, result, nil)
			}
			return result, nil
		}
	}

	// No converter found
	convErr := NewConversionError(fromType, toType, from, "no suitable converter found", nil)
	if r.enableCache {
		cacheKey := generateCacheKey(from, toType)
		r.cache.Set(cacheKey, nil, convErr)
	}

	return nil, convErr
}

// CanConvert checks if conversion is possible (direct or multi-hop)
func (r *Registry) CanConvert(fromType, toType reflect.Type) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check direct conversion
	for _, converter := range r.converters {
		if converter.CanConvert(fromType, toType) {
			return true
		}
	}

	// Check multi-hop conversion if enabled
	if r.enableMultiHop {
		return r.findConversionPath(fromType, toType, 0, make(map[reflect.Type]bool)) != nil
	}

	return false
}

// CanReverse checks if reverse conversion is possible
func (r *Registry) CanReverse(fromType, toType reflect.Type) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if any converter supports reverse conversion
	for _, converter := range r.converters {
		if converter.CanConvert(fromType, toType) && converter.CanReverse(fromType, toType) {
			return true
		}
	}

	return false
}

// ListConverters returns all registered converters
func (r *Registry) ListConverters() []TypeConverter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]TypeConverter, len(r.converters))
	copy(result, r.converters)
	return result
}

// GetCacheStats returns cache performance statistics
func (r *Registry) GetCacheStats() (hits, misses uint64, size int) {
	if !r.enableCache {
		return 0, 0, 0
	}
	return r.cache.Stats()
}

// ClearCache empties the conversion cache
func (r *Registry) ClearCache() {
	if r.enableCache {
		r.cache.Clear()
	}
}

// tryDirectConversion attempts direct conversion using registered converters
func (r *Registry) tryDirectConversion(from any, fromType, toType reflect.Type) (any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, converter := range r.converters {
		if converter.CanConvert(fromType, toType) {
			return converter.Convert(from, toType)
		}
	}

	return nil, NewConversionError(fromType, toType, from, "no direct converter available", nil)
}

// tryMultiHopConversion attempts conversion through intermediate types
func (r *Registry) tryMultiHopConversion(from any, fromType, toType reflect.Type) (any, error) {
	path := r.findConversionPath(fromType, toType, 0, make(map[reflect.Type]bool))
	if path == nil {
		return nil, NewConversionError(fromType, toType, from, "no conversion path found", nil)
	}

	// Execute the conversion path
	current := from
	for _, step := range path.Steps {
		result, err := step.Converter.Convert(current, step.ToType)
		if err != nil {
			return nil, NewConversionError(fromType, toType, from,
				fmt.Sprintf("multi-hop conversion failed at step %v->%v", step.FromType, step.ToType), err)
		}
		current = result
	}

	return current, nil
}

// findConversionPath finds a path for multi-hop conversion using DFS
func (r *Registry) findConversionPath(fromType, toType reflect.Type, depth int, visited map[reflect.Type]bool) *ConversionPath {
	if depth >= r.maxHops {
		return nil
	}

	if visited[fromType] {
		return nil // Avoid cycles
	}

	visited[fromType] = true
	defer func() { delete(visited, fromType) }()

	for _, converter := range r.converters {
		if !converter.CanConvert(fromType, toType) {
			continue
		}

		// Try direct conversion first
		if converter.CanConvert(fromType, toType) {
			return &ConversionPath{
				Steps: []ConversionStep{{
					Converter: converter,
					FromType:  fromType,
					ToType:    toType,
					Cost:      1,
				}},
				TotalCost:  1,
				Reversible: converter.CanReverse(fromType, toType),
			}
		}

		// Try indirect conversion through this converter's possible outputs
		// This is a simplified approach - in a more sophisticated implementation,
		// we would enumerate all possible intermediate types
		commonTypes := []reflect.Type{
			reflect.TypeOf(map[string]any{}),
			reflect.TypeOf(""),
			reflect.TypeOf(0),
			reflect.TypeOf(0.0),
			reflect.TypeOf(true),
			reflect.TypeOf([]any{}),
		}

		for _, intermediateType := range commonTypes {
			if converter.CanConvert(fromType, intermediateType) {
				subPath := r.findConversionPath(intermediateType, toType, depth+1, visited)
				if subPath != nil {
					// Construct the full path
					fullPath := &ConversionPath{
						Steps: append([]ConversionStep{{
							Converter: converter,
							FromType:  fromType,
							ToType:    intermediateType,
							Cost:      1,
						}}, subPath.Steps...),
						TotalCost:  1 + subPath.TotalCost,
						Reversible: converter.CanReverse(fromType, intermediateType) && subPath.Reversible,
					}
					return fullPath
				}
			}
		}
	}

	return nil
}

// convertNil handles conversion of nil values to target types
func (r *Registry) convertNil(toType reflect.Type) any {
	switch toType.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return reflect.Zero(toType).Interface()
	default:
		return reflect.Zero(toType).Interface()
	}
}
