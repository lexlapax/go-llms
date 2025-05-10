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

The `SchemaCache` (`pkg/structured/processor/schema_cache.go`) reduces the overhead of schema marshaling and implements comprehensive caching features including LRU eviction and TTL expiration:

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
    maxSize        int         // Maximum number of entries
    expirationTime time.Duration // How long entries live
    lastCleanup    time.Time    // Last time cache was cleaned up
}

// NewSchemaCache creates a new schema cache with default settings
func NewSchemaCache() *SchemaCache {
    return &SchemaCache{
        cache:          make(map[uint64]CacheEntry),
        maxSize:        1000,                  // Default max entries
        expirationTime: 30 * time.Minute,      // Default TTL
        lastCleanup:    time.Now(),
        metrics:        metrics.NewCacheMetrics("schema_cache"),
    }
}

// Get retrieves a schema JSON from the cache
func (c *SchemaCache) Get(key uint64) ([]byte, bool) {
    c.lock.RLock()
    entry, found := c.cache[key]
    c.lock.RUnlock()

    if !found {
        c.metrics.RecordMiss()
        return nil, false
    }

    // Update last access time and metrics
    c.lock.Lock()
    entry.LastAccess = time.Now()
    c.cache[key] = entry
    c.operations++
    c.lock.Unlock()

    c.metrics.RecordHit()

    // Check for cleanup opportunity (every 1000 operations)
    if c.operations%1000 == 0 {
        go c.CleanupExpired()
    }

    return entry.Value, true
}

// Set stores a schema JSON in the cache
func (c *SchemaCache) Set(key uint64, value []byte) {
    c.lock.Lock()
    defer c.lock.Unlock()

    c.cache[key] = CacheEntry{
        Value:      value,
        LastAccess: time.Now(),
    }
    c.size++

    // Check if we need to evict entries
    if len(c.cache) > c.maxSize {
        c.evictLRU()
    }
}
```

### Improved Schema Key Generation

The key generation function has been enhanced to handle all schema components for better cache key uniqueness:

```go
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

        // Add constraints to hash
        if prop.Minimum != nil {
            binary.Write(hasher, binary.LittleEndian, *prop.Minimum)
        }
        if prop.Maximum != nil {
            binary.Write(hasher, binary.LittleEndian, *prop.Maximum)
        }
        if prop.MinLength != nil {
            binary.Write(hasher, binary.LittleEndian, *prop.MinLength)
        }
        if prop.MaxLength != nil {
            binary.Write(hasher, binary.LittleEndian, *prop.MaxLength)
        }

        // Add enum values to hash
        for _, e := range prop.Enum {
            hasher.Write([]byte(e))
        }

        // Add nested properties recursively
        if prop.Items != nil {
            // Hash item type
            hasher.Write([]byte(prop.Items.Type))

            // Hash item properties (simplified for brevity)
            for itemK, itemProp := range prop.Items.Properties {
                hasher.Write([]byte(itemK))
                hasher.Write([]byte(itemProp.Type))
            }
        }
    }

    // Add title and description to hash
    hasher.Write([]byte(schema.Title))
    hasher.Write([]byte(schema.Description))

    // Add conditional schema components if present
    if schema.If != nil {
        hasher.Write([]byte("if"))
    }
    if schema.Then != nil {
        hasher.Write([]byte("then"))
    }
    if schema.Else != nil {
        hasher.Write([]byte("else"))
    }

    return hasher.Sum64()
}
```

### LRU Eviction and TTL Expiration

The cache implements Least Recently Used (LRU) eviction when capacity is reached and Time-To-Live (TTL) expiration for entries:

```go
// evictLRU removes the least recently used entries from the cache
func (c *SchemaCache) evictLRU() {
    // Convert to slice for sorting
    entries := make([]struct {
        key   uint64
        entry CacheEntry
    }, 0, len(c.cache))

    for k, v := range c.cache {
        entries = append(entries, struct {
            key   uint64
            entry CacheEntry
        }{k, v})
    }

    // Sort by last access time (oldest first)
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].entry.LastAccess.Before(entries[j].entry.LastAccess)
    })

    // Remove oldest 20% of entries
    removeCount := len(entries) / 5
    if removeCount < 1 {
        removeCount = 1
    }

    for i := 0; i < removeCount; i++ {
        delete(c.cache, entries[i].key)
        c.metrics.RecordEviction()
    }
}

// CleanupExpired removes expired entries from the cache
func (c *SchemaCache) CleanupExpired() {
    c.lock.Lock()
    defer c.lock.Unlock()

    // Only run cleanup once every minute maximum
    if time.Since(c.lastCleanup) < time.Minute {
        return
    }

    c.lastCleanup = time.Now()
    now := time.Now()
    expiredCount := 0

    for key, entry := range c.cache {
        if now.Sub(entry.LastAccess) > c.expirationTime {
            delete(c.cache, key)
            expiredCount++
        }
    }

    c.metrics.RecordCleanup(expiredCount)
}
```

### Cache Monitoring with Metrics

The cache includes comprehensive metrics collection:

```go
// CacheMetrics tracks performance statistics for the cache
type CacheMetrics struct {
    Hits        int64 // Number of cache hits
    Misses      int64 // Number of cache misses
    Evictions   int64 // Number of entries evicted due to capacity limits
    Expirations int64 // Number of entries expired due to TTL
    mu          sync.Mutex
}

// RecordHit increments the hit counter
func (m *CacheMetrics) RecordHit() {
    atomic.AddInt64(&m.Hits, 1)
}

// RecordMiss increments the miss counter
func (m *CacheMetrics) RecordMiss() {
    atomic.AddInt64(&m.Misses, 1)
}

// GetHitRate returns the cache hit rate as a percentage
func (m *CacheMetrics) GetHitRate() float64 {
    total := atomic.LoadInt64(&m.Hits) + atomic.LoadInt64(&m.Misses)
    if total == 0 {
        return 0
    }
    return float64(atomic.LoadInt64(&m.Hits)) / float64(total) * 100
}
```

### Key Features

1. **Fast hashing**: Uses the non-cryptographic FNV hash for speed
2. **Binary storage**: Stores pre-marshaled binary JSON data
3. **Efficient key generation**: Includes all schema components for accurate uniqueness
4. **LRU eviction**: Implements Least Recently Used eviction policy to manage capacity
5. **TTL expiration**: Automatically removes stale entries based on a time-to-live
6. **Thread safety**: Uses a read-write mutex for concurrent access
7. **Metrics**: Includes detailed performance metrics for monitoring
8. **Background cleanup**: Performs periodic background cleanup to remove expired entries

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