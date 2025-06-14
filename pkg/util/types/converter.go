// ABOUTME: TypeConverter interface and core conversion logic for bridge-friendly type system
// ABOUTME: Enables seamless type conversion between Go types and scripting language types

package types

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

// TypeConverter defines the interface for converting between types
type TypeConverter interface {
	// Convert transforms a value from one type to another
	Convert(from any, toType reflect.Type) (any, error)

	// CanConvert checks if this converter can handle the from->to conversion
	CanConvert(fromType, toType reflect.Type) bool

	// CanReverse checks if this converter supports reverse conversion
	CanReverse(fromType, toType reflect.Type) bool

	// Name returns a human-readable name for this converter
	Name() string

	// Priority returns the priority of this converter (higher = preferred)
	Priority() int
}

// ConversionError represents an error during type conversion
type ConversionError struct {
	FromType reflect.Type
	ToType   reflect.Type
	Value    any
	Reason   string
	Wrapped  error
}

func (e *ConversionError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("conversion error: cannot convert %v (type %v) to %v: %s: %v",
			e.Value, e.FromType, e.ToType, e.Reason, e.Wrapped)
	}
	return fmt.Sprintf("conversion error: cannot convert %v (type %v) to %v: %s",
		e.Value, e.FromType, e.ToType, e.Reason)
}

func (e *ConversionError) Unwrap() error {
	return e.Wrapped
}

// NewConversionError creates a new conversion error
func NewConversionError(fromType, toType reflect.Type, value any, reason string, wrapped error) *ConversionError {
	return &ConversionError{
		FromType: fromType,
		ToType:   toType,
		Value:    value,
		Reason:   reason,
		Wrapped:  wrapped,
	}
}

// ConversionPath represents a path for multi-hop conversion
type ConversionPath struct {
	Steps      []ConversionStep
	TotalCost  int
	Reversible bool
}

// ConversionStep represents a single step in a conversion path
type ConversionStep struct {
	Converter TypeConverter
	FromType  reflect.Type
	ToType    reflect.Type
	Cost      int
}

// ConversionCache stores cached conversion results for performance
type ConversionCache struct {
	cache  map[string]CacheEntry
	hits   uint64
	misses uint64
	mu     sync.RWMutex
}

// CacheEntry represents a cached conversion result
type CacheEntry struct {
	Result any
	Error  error
}

// NewConversionCache creates a new conversion cache
func NewConversionCache() *ConversionCache {
	return &ConversionCache{
		cache: make(map[string]CacheEntry),
	}
}

// Get retrieves a cached conversion result
func (c *ConversionCache) Get(key string) (any, error, bool) {
	c.mu.RLock()
	entry, exists := c.cache[key]
	c.mu.RUnlock()

	if exists {
		atomic.AddUint64(&c.hits, 1)
		return entry.Result, entry.Error, true
	}
	atomic.AddUint64(&c.misses, 1)
	return nil, nil, false
}

// Set stores a conversion result in the cache
func (c *ConversionCache) Set(key string, result any, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = CacheEntry{Result: result, Error: err}
}

// Clear empties the cache
func (c *ConversionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]CacheEntry)
	atomic.StoreUint64(&c.hits, 0)
	atomic.StoreUint64(&c.misses, 0)
}

// Stats returns cache performance statistics
func (c *ConversionCache) Stats() (hits, misses uint64, size int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return atomic.LoadUint64(&c.hits), atomic.LoadUint64(&c.misses), len(c.cache)
}

// generateCacheKey creates a cache key for a conversion
func generateCacheKey(value any, toType reflect.Type) string {
	fromType := reflect.TypeOf(value)
	return fmt.Sprintf("%v->%v:%v", fromType, toType, value)
}
