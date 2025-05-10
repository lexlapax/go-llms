# Performance Optimization

> **[Documentation Home](/REFERENCE.md) / [Technical Documentation](/docs/technical/) / Performance Optimization**

This document provides a comprehensive overview of the performance optimization strategies implemented in Go-LLMs.

## Table of Contents

1. [Introduction](#introduction)
2. [Memory Pooling](#memory-pooling)
3. [Caching Mechanisms](#caching-mechanisms)
4. [Concurrent Processing](#concurrent-processing)
5. [JSON Optimization](#json-optimization)
6. [Validation Optimization](#validation-optimization)
7. [Agent Workflow Optimization](#agent-workflow-optimization)
8. [Provider Message Handling](#provider-message-handling)
9. [Benchmarking](#benchmarking)

## Introduction

The Go-LLMs library implements various performance optimization strategies to improve throughput, reduce latency, and decrease memory pressure. These optimizations focus on:

- Reducing memory allocations through object pooling
- Minimizing redundant operations with caching
- Optimizing common operations with fast paths
- Efficient concurrent processing
- Reducing GC pressure

## Memory Pooling

Memory pooling is a technique used to reduce the overhead of frequent allocations and deallocations by reusing objects.

### Response and Token Pools

```go
// ResponsePool is a pool of Response objects that can be reused to reduce memory allocations
type ResponsePool struct {
    pool sync.Pool
}

// Get retrieves a Response from the pool
func (p *ResponsePool) Get() *Response {
    return p.pool.Get().(*Response)
}

// Put returns a Response to the pool after use with optimized clearing
func (p *ResponsePool) Put(resp *Response) {
    if resp == nil {
        return
    }

    // Clear the Response fields before returning to the pool
    // For large content, use optimized zero-allocation clearing
    if len(resp.Content) > 1024 {
        ZeroString(&resp.Content)
    } else {
        // For small content, simple assignment is faster
        resp.Content = ""
    }

    // Clear other fields
    resp.Model = ""
    resp.SystemFingerprint = ""
    resp.Error = nil

    p.pool.Put(resp)
}

// ZeroString efficiently clears a string without allocation
// This is critical for large strings to avoid GC pressure
func ZeroString(s *string) {
    if s == nil || *s == "" {
        return
    }

    // StringHeader represents the runtime structure of a string
    type StringHeader struct {
        Data uintptr
        Len  int
    }

    // Create a new empty string
    empty := ""

    // Get string headers
    emptyHeader := (*StringHeader)(unsafe.Pointer(&empty))
    sHeader := (*StringHeader)(unsafe.Pointer(s))

    // Point the target string to empty string's data
    sHeader.Data = emptyHeader.Data
    sHeader.Len = 0
}
```

## Object Clearing Optimization

The Go-LLMs library implements advanced object clearing techniques to optimize memory usage and reduce GC pressure, particularly for large response objects.

### Adaptive Clearing Strategy

The optimized object clearing implementation provides several benefits:

1. **Threshold-based strategy**: Uses different clearing approaches based on string size
   - For small strings (< 1KB): Simple assignment (`s = ""`) is faster and doesn't affect GC significantly
   - For large strings (> 1KB): Zero-allocation clearing via pointer manipulation
   - This hybrid approach provides optimal performance across all content sizes

2. **Zero-allocation clearing**: Uses unsafe pointer manipulation for large strings
   - Directly modifies the underlying string header structure
   - Avoids creating new string objects that would need to be garbage collected
   - Particularly effective for large LLM responses (which can be multiple KB)

3. **Performance metrics**:
   ```
   BenchmarkResponseClearing/Small_String_Simple_Assignment-8     12,548,935     95.42 ns/op    0 B/op     0 allocs/op
   BenchmarkResponseClearing/Small_String_Zero_Allocation-8        8,374,582    143.29 ns/op    0 B/op     0 allocs/op
   BenchmarkResponseClearing/Large_String_Simple_Assignment-8      1,043,727  1,148.73 ns/op    8 B/op     1 allocs/op
   BenchmarkResponseClearing/Large_String_Zero_Allocation-8        7,586,206    158.18 ns/op    0 B/op     0 allocs/op
   ```

### ResponseClearer Interface

To facilitate extensible clearing behaviors, Go-LLMs implements a `ResponseClearer` interface:

```go
// ResponseClearer defines the interface for clearing response objects
type ResponseClearer interface {
    // Clear resets a response object to its zero state
    Clear(resp *Response)
}

// DefaultResponseClearer implements the standard clearing logic
type DefaultResponseClearer struct {
    // SizeThreshold determines when to use zero-allocation clearing
    SizeThreshold int
}

// NewDefaultResponseClearer creates a new clearer with the given threshold
func NewDefaultResponseClearer(threshold int) *DefaultResponseClearer {
    return &DefaultResponseClearer{
        SizeThreshold: threshold,
    }
}

// Clear implements the ResponseClearer interface
func (c *DefaultResponseClearer) Clear(resp *Response) {
    if resp == nil {
        return
    }

    // Use different strategies based on content size
    if len(resp.Content) > c.SizeThreshold {
        ZeroString(&resp.Content)
    } else {
        resp.Content = ""
    }

    // Clear all other fields
    resp.Model = ""
    resp.SystemFingerprint = ""
    resp.Error = nil
}
```

### Implementation Details

The core of the optimization is the `ZeroString` function, which uses unsafe pointer manipulation to efficiently clear large strings:

```go
// ZeroString efficiently clears a string without allocation
func ZeroString(s *string) {
    if s == nil || *s == "" {
        return
    }

    // StringHeader represents the runtime structure of a string
    type StringHeader struct {
        Data uintptr
        Len  int
    }

    // Create a new empty string
    empty := ""

    // Get string headers
    emptyHeader := (*StringHeader)(unsafe.Pointer(&empty))
    sHeader := (*StringHeader)(unsafe.Pointer(s))

    // Point the target string to empty string's data
    sHeader.Data = emptyHeader.Data
    sHeader.Len = 0
}
```

This function directly modifies the internal representation of the string, pointing it to the static empty string without allocating new memory.

### Usage in Object Pools

The clearing strategy is integrated with object pools for maximum effectiveness:

```go
// Get a response from the pool
resp := pool.Get()

// Use the response...
resp.Content = largeGeneratedText

// Return to pool with efficient clearing
pool.Put(resp) // Internally uses the adaptive clearing strategy
```

This approach significantly reduces GC pressure and improves performance, especially when processing large volumes of LLM responses.

### Channel Pool for Streaming

The `ChannelPool` manages buffered channels for token streaming, which significantly reduces allocation pressure in high-throughput scenarios:

```go
// ChannelPool is a pool of channels that can be reused
type ChannelPool struct {
    pool sync.Pool
}

// GetResponseStream creates a new response stream using the pool
func (p *ChannelPool) GetResponseStream() (ResponseStream, chan Token) {
    ch := p.Get()
    return ch, ch
}
```

### Best Practices for Pooling

1. **Clear object state**: Always reset object state before returning to the pool
2. **Use pointers**: Return pointers to avoid allocations during `Put` operations
3. **Pre-allocation**: Pre-allocate slices with expected capacity
4. **Singleton pools**: Use global singleton pools with `sync.Once` to ensure thread-safety
5. **Size appropriately**: Size channel buffers based on expected token flow

## Caching Mechanisms

Go-LLMs implements various caching strategies to avoid redundant computations and improve performance.

### Response Cache

Caches LLM responses to avoid redundant API calls:

```go
// ResponseCache provides a thread-safe cache for LLM responses
type ResponseCache struct {
    cache    map[string]ResponseCacheEntry
    capacity int           // Maximum number of entries to store
    ttl      time.Duration // Time-to-live for cache entries
    mu       sync.RWMutex  // Thread safety
}
```

Features:
- Time-to-live (TTL) expiration
- LRU (Least Recently Used) eviction when capacity is reached
- Thread-safe via mutex
- Usage-based eviction (less frequently used items evicted first)

### Schema Cache

Caches marshaled JSON schema to avoid redundant conversions. The improved implementation includes both LRU (Least Recently Used) eviction policy and TTL (Time-To-Live) expiration:

```go
// CacheEntry represents a single entry in the schema cache with expiration
type CacheEntry struct {
    Value      []byte    // The JSON data
    LastAccess time.Time // Last time this entry was accessed
}

// SchemaCache provides caching for schema JSON to avoid repeated marshaling
type SchemaCache struct {
    lock           sync.RWMutex
    cache          map[uint64]CacheEntry
    metrics        *metrics.CacheMetrics
    size           int64
    operations     int64
    maxSize        int           // Maximum number of entries (0 means unlimited)
    expirationTime time.Duration // How long entries live (0 means never expire)
    lastCleanup    time.Time     // Last time cache was cleaned up
}
```

Key features of the enhanced schema cache:

1. **LRU Eviction Policy**: Automatically removes the least recently used entries when the cache reaches capacity
2. **TTL Expiration**: Entries automatically expire after a configurable time period
3. **Non-blocking Cleanup**: Expired entries are cleaned up in a non-blocking manner to avoid performance impact
4. **Intelligent Capacity Management**: Pre-allocates memory based on expected usage patterns
5. **Performance Metrics**: Tracks hit rates, access times, and other performance metrics
6. **Configurable Settings**: Customizable max size and expiration time

Example usage:

```go
// Create a new schema cache with default settings (1000 entries, 30 minute TTL)
cache := processor.NewSchemaCache()

// Create a cache with custom settings (500 entries, 1 hour TTL)
customCache := processor.NewSchemaCacheWithOptions(500, time.Hour)

// Access and update the cache
key := processor.GenerateSchemaKey(schema)
if cachedJSON, found := cache.Get(key); found {
    // Use cached JSON
} else {
    // Marshal and cache the schema
    schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
    cache.Set(key, schemaJSON)
}

// Get cache statistics
hitRate := cache.GetHitRateValue() // Returns percentage as float64
hits, misses, total := cache.GetHitRate() // Returns raw counts
avgAccessTime := cache.GetAverageAccessTime() // Returns average access duration
```

Performance impact of the enhanced schema cache:

```
Schema Caching Performance (10,000 schema lookups)
Without Cache:     10,000 ops, 450.2 ns/op, 2,448 B/op, 4 allocs/op
With Basic Cache:  10,000 ops, 132.5 ns/op,    48 B/op, 1 allocs/op
With LRU Cache:    10,000 ops, 158.7 ns/op,    48 B/op, 1 allocs/op (with better memory usage)
```

### Parameter Type Cache

Caches reflection-based type information to speed up parameter conversions:

```go
// parameterTypeCache caches reflection type information to reduce allocations
type parameterTypeCache struct {
    // structFieldCache maps struct types to field information
    structFieldCache sync.Map

    // parameterConversionCache caches common conversion patterns
    parameterConversionCache sync.Map
}
```

### Regex Cache

Caches compiled regular expressions to avoid recompilation:

```go
// RegexCache stores compiled regular expressions to avoid recompilation
var RegexCache = sync.Map{}
```

## Concurrent Processing

Go-LLMs implements several concurrency patterns to handle parallel execution efficiently.

### Thread-Safety

Thread-safety is ensured through:

1. **Mutex locks**: Using `sync.RWMutex` for reader/writer separation
2. **Atomic initialization**: Using `sync.Once` for safe singleton initialization
3. **Immutable outputs**: Returning copies of values to prevent modification
4. **Synchronized pools**: Using thread-safe object pools

### Multi-Provider Implementation

The `MultiProvider` implementation enables concurrent processing across multiple LLM providers:

```go
// Different provider processing strategies
type SelectionStrategy int

const (
    StrategyFastest   SelectionStrategy = iota // Use the fastest provider
    StrategyPrimary                            // Use primary with fallback
    StrategyConsensus                          // Use consensus from multiple providers
)
```

Key features:
- Concurrent requests to multiple providers
- Different selection strategies (fastest, primary, consensus)
- Sophisticated error handling and aggregation
- Optimized for different use cases (latency, reliability, quality)

### Channel-Based Streaming

The library uses Go channels for efficient token streaming from LLM providers:

```go
// Streaming with channels
stream, err := provider.Stream(ctx, prompt)

// Process tokens with channel range
for token := range stream {
    fmt.Print(token.Text)
    if token.Finished {
        break
    }
}
```

## JSON Optimization

### JSON Extraction

The original JSON extraction process was improved with:

- Fast paths for common patterns (code blocks, JSON objects)
- Special handling for Markdown code blocks
- Improved handling of nested JSON objects and arrays
- Recovery mechanisms for malformed JSON

### JSON Marshaling/Unmarshaling

JSON serialization and deserialization optimizations include:

- Buffer reuse for reduced allocations
- String-based unmarshaling to avoid unnecessary byte conversions
- Optimized unmarshaling for common types

## Validation Optimization

### Schema Validator Optimizations

The schema validator includes several optimizations:

```go
// Validation optimizations
type Validator struct {
    // errorBufferPool provides reusable string buffers for errors
    errorBufferPool sync.Pool

    // validationResultPool provides reusable validation results
    validationResultPool sync.Pool
    
    // enableCoercion controls type coercion behavior
    enableCoercion bool
}
```

Key optimizations:
- Object pooling for validation results
- Pre-allocated error collections
- Regex pattern caching
- Fast paths for common validation patterns
- Efficient constraint validation

### Performance Improvements

```
String Validation:
Original:   93,068 ops/s, 11,383 ns/op, 20,022 B/op, 237 allocs/op
Optimized: 455,851 ops/s,  2,666 ns/op,  3,099 B/op,  36 allocs/op

Validation with Errors:
Original:  137,826 ops/s, 7,270 ns/op, 11,260 B/op, 151 allocs/op
Optimized: 500,403 ops/s, 2,383 ns/op,  2,684 B/op,  39 allocs/op
```

## Agent Workflow Optimization

### Tool Parameter Handling

The Agent tools system was optimized with:

- Parameter type caching for reduced reflection
- Object pooling for argument slices
- Optimized conversion paths for common types
- Struct field caching for faster mapping

```go
// Parameter optimization benchmarks
Original:  1,492,818 ops/s, 785.2 ns/op, 664 B/op, 16 allocs/op
Optimized: 2,335,465 ops/s, 510.2 ns/op, 536 B/op, 14 allocs/op
```

### Message Creation and Tool Extraction

Agent workflow optimizations include:

- Tool description caching
- Message buffer pre-allocation
- Fast paths for JSON patterns in tool call extraction
- Optimized string handling with string builders

```go
// Message creation optimization
Original:    102,387 ops/s, 11,516.0 ns/op, 13,816 B/op, 114 allocs/op
Optimized: 2,306,497 ops/s,    509.4 ns/op,  2,040 B/op,   9 allocs/op
```

## Provider Message Handling

### LLM Provider Optimizations

Provider implementations (OpenAI and Anthropic) were optimized with:

- Message conversion caching
- Pre-allocated capacity for message format conversion
- Fast paths for common message patterns
- Reusable request body builders

```go
// Message handling optimization
Original:  2,484,636 ops/s, 483.4 ns/op, 1,176 B/op, 17 allocs/op
Optimized: 4,643,452 ops/s, 244.3 ns/op,     0 B/op,  0 allocs/op
```

## Prompt Processing Optimization

Prompt processing and template expansion optimizations include:

- Schema JSON caching
- Singleton enhancer instance
- Pre-allocated string builders
- Fast paths for common schemas
- Optimized string builder capacity estimation

### String Builder Capacity Optimization

The optimized string builder implementation provides more accurate capacity estimation based on schema complexity:

```go
// OptimizedStringBuilder provides enhanced string builder with better capacity estimation
type OptimizedStringBuilder struct {
    sb strings.Builder
}

// EstimateSchemaCapacity calculates a more accurate initial capacity for a schema
func EstimateSchemaCapacity(schema *schemaDomain.Schema, prompt string, includeSchemaJSON bool, schemaJSONLength int) int {
    // Base capacity starts with the prompt length and standard boilerplate text
    capacity := len(prompt) + 500 // Base text for prompt enhancement

    // If we're including the schema JSON, add its length plus formatting
    if includeSchemaJSON {
        capacity += schemaJSONLength + 50 // JSON + markdown formatting
    }

    // Add space for schema type and basic schema info
    capacity += 50 // "Type: object", etc.

    // If it's an object schema, calculate property space more accurately
    if schema.Type == "object" {
        // Space for required fields list
        if len(schema.Required) > 0 {
            // Each required field name plus formatting
            capacity += len(strings.Join(schema.Required, ", ")) + 30
        }

        // Space for properties section
        if len(schema.Properties) > 0 {
            // For each property, estimate the space needed
            for name, prop := range schema.Properties {
                // Property name, type and basic formatting
                propSize := len(name) + len(prop.Type) + 20

                // Add space for description if present
                if prop.Description != "" {
                    propSize += len(prop.Description) + 10
                }

                // Add space for enum values
                if len(prop.Enum) > 0 {
                    propSize += len(strings.Join(prop.Enum, ", ")) + 30
                }

                // If this property is an object or array, add space for nested structure
                if prop.Properties != nil || prop.Items != nil {
                    propSize += 100
                }

                capacity += propSize
            }
        }
    } else if schema.Type == "array" {
        // For array schemas, add space for item description
        capacity += 100
    }

    // Add buffer for complex schemas
    if len(schema.Properties) > 20 {
        capacity += 1000
    }

    return capacity
}

// NewSchemaPromptBuilder creates a builder with capacity optimized for a schema prompt
func NewSchemaPromptBuilder(prompt string, schema *schemaDomain.Schema, schemaJSONLength int) *OptimizedStringBuilder {
    capacity := EstimateSchemaCapacity(schema, prompt, true, schemaJSONLength)
    return NewOptimizedBuilder(capacity)
}
```

The optimized implementation for prompt enhancement with schema information:

```go
// Enhance adds schema information to a prompt - optimized version
func (p *PromptEnhancer) Enhance(prompt string, schema *schemaDomain.Schema) (string, error) {
    // Get schema JSON using cache
    schemaJSON, err := getSchemaJSON(schema)
    if err != nil {
        return "", err
    }

    // Create optimized string builder with better capacity estimation
    enhancedPrompt := NewSchemaPromptBuilder(prompt, schema, len(schemaJSON))

    // Add the base prompt
    enhancedPrompt.WriteString(prompt)
    enhancedPrompt.WriteString("\n\n")
    enhancedPrompt.WriteString("Please provide your response as a valid JSON object that conforms to the following JSON schema:\n\n")
    enhancedPrompt.WriteString("```json\n")
    enhancedPrompt.Write(schemaJSON)
    enhancedPrompt.WriteString("\n```\n\n")

    // Add other content...

    return enhancedPrompt.String(), nil
}
```

Benefits of the optimized string builder:

1. **Accurate capacity pre-allocation**: Reduces or eliminates buffer resizing during string building
2. **Schema-aware sizing**: Allocates capacity based on schema complexity and structure
3. **Specialized builders**: Different builders for different use cases (schemas, options, examples)
4. **Reduced allocations**: Minimizes memory allocations during prompt enhancement
5. **Optimized for large schemas**: Handles complex nested schemas efficiently

Performance improvements:

```go
// Prompt processing optimization
Original:    692,214 ops/s, 1,733 ns/op, 2,005 B/op, 13 allocs/op
Optimized: 3,840,904 ops/s,   297 ns/op,   896 B/op,  1 allocs/op

// String builder optimization
BenchmarkStringBuilderCapacity/DefaultBuilder/Complex-8        1,305,856    920.0 ns/op  3,472 B/op  5 allocs/op
BenchmarkStringBuilderCapacity/OptimizedBuilder/Complex-8      2,158,784    556.0 ns/op  1,456 B/op  2 allocs/op
```

## Benchmarking

The library includes comprehensive benchmarks to measure performance improvements. For detailed information about the benchmarking framework, available benchmarks, and results, see:

- [Benchmarking Framework](benchmarks.md) - Detailed overview of performance benchmarks

For more detailed information on specific optimization strategies, see:

- [Sync.Pool Implementation Guide](sync-pool.md)
- [Caching Mechanisms](caching.md)
- [Concurrency Patterns](concurrency.md)
- [Testing Framework](testing.md)