# Caching Mechanisms in Go-LLMs

> **[Documentation Home](/REFERENCE.md) / [Technical Documentation](/docs/technical/) / Caching Mechanisms**

This document explains the various caching mechanisms implemented in Go-LLMs to improve performance, reduce redundant operations, and minimize latency. It covers design patterns, implementation details, and best practices for effective caching.

*Related: [Performance Optimization](performance.md) | [Sync.Pool Implementation](sync-pool.md) | [Concurrency Patterns](concurrency.md)*

## Overview of Caching in Go-LLMs

Go-LLMs implements several types of caches to optimize different operations:

1. **Value caches**: Store results of expensive computations
2. **Response caches**: Avoid redundant API calls to LLM providers  
3. **Reflection caches**: Speed up type handling and parameter binding
4. **Compilation caches**: Avoid recompilation of regular expressions
5. **Schema caches**: Reduce schema marshaling overhead

Each cache is designed to address specific performance bottlenecks in the codebase.

## Response Cache Implementation

The `ResponseCache` (`pkg/agent/workflow/response_cache.go`) provides a thread-safe cache for LLM responses to avoid redundant API calls:

```go
// ResponseCache provides a thread-safe cache for LLM responses
// to avoid redundant API calls for the same input
type ResponseCache struct {
    // Main cache storage - key is hash of the messages + options
    cache    map[string]ResponseCacheEntry
    capacity int           // Maximum number of entries to store
    ttl      time.Duration // Time-to-live for cache entries
    mu       sync.RWMutex  // Thread safety
}

// ResponseCacheEntry represents a cached response
type ResponseCacheEntry struct {
    Response   ldomain.Response
    Timestamp  time.Time
    UsageCount int
    Source     string // Description of where response came from (model, etc.)
}
```

### Key Features

1. **Thread safety**: Uses `sync.RWMutex` for reader/writer separation
2. **Time-to-live (TTL)**: Entries expire after a configurable time period
3. **Capacity management**: Implements LRU (Least Recently Used) eviction
4. **Usage tracking**: Tracks how often each entry is accessed
5. **Deterministic keys**: Uses cryptographic hashing for cache keys

### Usage Pattern

```go
// Creating a cache
cache := NewResponseCache(100, 5*time.Minute)

// Getting from cache
if response, found := cache.Get(messages, options); found {
    return response
}

// Setting in cache
cache.Set(messages, options, response, "model-name")

// Cleaning up expired entries
cache.Cleanup()
```

### Cache Key Generation

The cache uses deterministic key generation based on the input parameters:

```go
// generateCacheKey creates a deterministic key for caching
func (c *ResponseCache) generateCacheKey(messages []ldomain.Message, options []ldomain.Option) string {
    // Normalize messages to avoid inconsistent caching
    var messageData []map[string]string
    for _, msg := range messages {
        messageData = append(messageData, map[string]string{
            "role":    string(msg.Role),
            "content": msg.Content,
        })
    }

    // Add options to the key if present
    var optionsData []map[string]string
    for _, opt := range options {
        // Get option name and value through reflection
        optType := fmt.Sprintf("%T", opt)
        optStr := fmt.Sprintf("%v", opt)

        // Add to options data
        optionsData = append(optionsData, map[string]string{
            "type":  optType,
            "value": optStr,
        })
    }

    // Create a combined structure for hashing
    data := map[string]interface{}{
        "messages": messageData,
        "options":  optionsData,
    }

    // Marshal to JSON and hash
    jsonData, err := json.Marshal(data)
    if err != nil {
        // Fallback to a simpler key if marshaling fails
        var sb strings.Builder
        for _, msg := range messages {
            sb.WriteString(string(msg.Role))
            sb.WriteString(":")
            sb.WriteString(msg.Content)
            sb.WriteString("|")
        }
        return hashString(sb.String())
    }

    return hashString(string(jsonData))
}

// hashString creates a SHA-256 hash of the input string
func hashString(input string) string {
    hash := sha256.Sum256([]byte(input))
    return hex.EncodeToString(hash[:])
}
```

### Cache Cleanup and Eviction

When the cache reaches capacity, older and less-used entries are evicted:

```go
// cleanupLocked applies truncation based on the configuration
func (c *ResponseCache) cleanupLocked() {
    now := time.Now()

    // First pass: remove expired entries
    for key, entry := range c.cache {
        if now.Sub(entry.Timestamp) > c.ttl {
            delete(c.cache, key)
        }
    }

    // If still over capacity, remove least recently used entries
    if len(c.cache) > c.capacity {
        // Convert map to slice for sorting
        entries := make([]struct {
            key   string
            entry ResponseCacheEntry
        }, 0, len(c.cache))

        for k, v := range c.cache {
            entries = append(entries, struct {
                key   string
                entry ResponseCacheEntry
            }{k, v})
        }

        // Sort by usage count and timestamp (older and less used first)
        sort.Slice(entries, func(i, j int) bool {
            // Primary sort by usage count
            if entries[i].entry.UsageCount != entries[j].entry.UsageCount {
                return entries[i].entry.UsageCount < entries[j].entry.UsageCount
            }
            // Secondary sort by timestamp
            return entries[i].entry.Timestamp.Before(entries[j].entry.Timestamp)
        })

        // Remove oldest entries until we're under capacity
        for i := 0; i < len(entries) && len(c.cache) > c.capacity; i++ {
            delete(c.cache, entries[i].key)
        }
    }
}
```

## Schema Cache Implementation

The `SchemaCache` (`pkg/structured/processor/schema_cache.go`) reduces the overhead of schema marshaling:

```go
// SchemaCache provides caching for schema JSON to avoid repeated marshaling
type SchemaCache struct {
    lock  sync.RWMutex
    cache map[uint64][]byte
}

// GenerateSchemaKey creates a hash key for a schema
// This is used for cache lookups to avoid repeated JSON marshaling
func GenerateSchemaKey(schema *schemaDomain.Schema) uint64 {
    hasher := fnv.New64()

    // Add schema type to hash
    hasher.Write([]byte(schema.Type))

    // Add required fields to hash
    for _, req := range schema.Required {
        hasher.Write([]byte(req))
    }

    // Add properties to hash (keys and types are the most important for uniqueness)
    for k, prop := range schema.Properties {
        hasher.Write([]byte(k))
        hasher.Write([]byte(prop.Type))
        // Add description to hash
        hasher.Write([]byte(prop.Description))
    }

    // Add title and description to hash
    hasher.Write([]byte(schema.Title))
    hasher.Write([]byte(schema.Description))

    return hasher.Sum64()
}
```

### Key Features

1. **Fast hashing**: Uses the non-cryptographic FNV hash for speed
2. **Binary storage**: Stores pre-marshaled binary JSON data
3. **Efficient key generation**: Only includes schema components that affect the output

## Regex Compilation Cache

Regular expressions are compiled only once and cached for reuse:

```go
// RegexCache stores compiled regular expressions to avoid recompilation
var RegexCache = sync.Map{}

// Usage example
var re *regexp.Regexp
if cached, found := RegexCache.Load(pattern); found {
    re = cached.(*regexp.Regexp)
} else {
    var err error
    re, err = regexp.Compile(pattern)
    if err != nil {
        errors = append(errors, fmt.Sprintf("invalid pattern: %v", err))
        return errors
    }
    RegexCache.Store(pattern, re)
}
```

### Key Features

1. **Global cache**: Single cache shared across all validation operations
2. **Concurrent map**: Uses `sync.Map` for thread-safe access without locking
3. **Type assertion**: Safely retrieves the compiled regex via type assertion

## Parameter Type Cache

The `parameterTypeCache` (`pkg/agent/tools/param_cache.go`) caches reflection-based type information:

```go
// parameterTypeCache caches reflection type information to reduce allocations
// during repeated tool executions with the same parameter types
type parameterTypeCache struct {
    // structFieldCache maps struct types to field information to avoid repeated reflection lookups
    structFieldCache sync.Map // map[reflect.Type][]fieldInfo

    // parameterConversionCache caches common conversion patterns
    parameterConversionCache sync.Map // map[typePair]bool
}

// fieldInfo caches information about a struct field
type fieldInfo struct {
    index      int
    name       string
    jsonName   string
    fieldType  reflect.Type
    canSet     bool
    isExported bool
}
```

### Key Features

1. **Struct field caching**: Stores field information to avoid reflection lookups
2. **Conversion caching**: Remembers which types can be converted to others
3. **JSON tag awareness**: Captures field names from struct tags for mapping

## Cache-Related Performance Considerations

### 1. Cache Hit Ratio

The effectiveness of a cache depends on its hit ratio. Monitor your cache hit rates to determine if your caching strategy is working well.

### 2. Thread Contention

Under high concurrency, mutex-protected caches can become bottlenecks. Use strategies to reduce contention:

- Use reader/writer locks when reads are more common than writes
- Consider sharded maps for very high concurrency
- Use lock-free structures like `sync.Map` where appropriate

### 3. Memory Usage vs. Performance

Caches trade memory for CPU performance. Monitor memory usage and adjust cache sizes accordingly.

### 4. Garbage Collection Impact

Large caches can impact GC performance. Consider:

- Setting appropriate capacity limits on caches
- Implementing periodic cleanup for stale entries
- Using value types rather than pointer types when items are small

## Best Practices for Effective Caching

### 1. Choose the Right Cache Key

Ensure cache keys capture all relevant inputs that affect the output:

```go
// Example of comprehensive key generation
func generateKey(req Request) string {
    h := sha256.New()
    
    // Include all fields that affect the result
    h.Write([]byte(req.UserID))
    h.Write([]byte(req.Query))
    h.Write([]byte(fmt.Sprintf("%v", req.Filters)))
    h.Write([]byte(fmt.Sprintf("%d", req.Version)))
    
    return hex.EncodeToString(h.Sum(nil))
}
```

### 2. Implement Thread-Safe Access

Always ensure thread-safe access to caches:

```go
// Thread-safe cache access
func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    value, found := c.data[key]
    return value, found
}

func (c *Cache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.data[key] = value
}
```

### 3. Handle Cache Invalidation

Implement mechanisms to invalidate cache entries when the underlying data changes:

```go
// Invalidate a specific entry
func (c *Cache) Invalidate(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    delete(c.data, key)
}

// Invalidate entries matching a pattern
func (c *Cache) InvalidatePattern(pattern string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    re := regexp.MustCompile(pattern)
    for k := range c.data {
        if re.MatchString(k) {
            delete(c.data, k)
        }
    }
}
```

### 4. Use TTL and Size Limits

Always implement TTL and size limits to prevent unbounded growth:

```go
// New cache with TTL and capacity limits
func NewCache(capacity int, ttl time.Duration) *Cache {
    return &Cache{
        data:     make(map[string]cacheEntry, capacity),
        capacity: capacity,
        ttl:      ttl,
    }
}
```

### 5. Implement Statistics and Monitoring

Add monitoring capabilities to track cache effectiveness:

```go
// CacheStats provides metrics about cache performance
type CacheStats struct {
    Size      int
    Capacity  int
    Hits      int64
    Misses    int64
    HitRatio  float64
    AvgAccess time.Duration
}

// GetStats returns current cache statistics
func (c *Cache) GetStats() CacheStats {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    hitRatio := 0.0
    if c.hits+c.misses > 0 {
        hitRatio = float64(c.hits) / float64(c.hits+c.misses)
    }
    
    return CacheStats{
        Size:      len(c.data),
        Capacity:  c.capacity,
        Hits:      c.hits,
        Misses:    c.misses,
        HitRatio:  hitRatio,
        AvgAccess: c.totalAccessTime / time.Duration(c.hits+c.misses),
    }
}
```

## Conclusion

Effective caching is a critical optimization technique in Go-LLMs. By implementing appropriate caching mechanisms with proper key generation, thread safety, and capacity management, the library achieves significant performance improvements, particularly for operations that would otherwise require expensive computation or external API calls.

The various caching implementations in Go-LLMs demonstrate a thoughtful approach to performance optimization, with each cache tailored to the specific requirements of the operation being optimized.