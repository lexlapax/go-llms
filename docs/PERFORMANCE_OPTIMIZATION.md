# Performance Optimization Strategies

This document outlines the performance optimization strategies implemented in the Go-LLMs library. These optimizations focus on reducing memory allocations, efficiently managing object lifecycles, and implementing concurrency patterns to improve throughput and reduce latency.

## Memory Pooling

Memory pooling is a technique used to reduce the overhead of frequent allocations and deallocations by reusing objects. Go-LLMs implements several object pools using the standard library's `sync.Pool` to reduce garbage collection pressure.

### Implementation Examples

#### Response and Token Pools (`pkg/llm/domain/pool.go`)

```go
// ResponsePool is a pool of Response objects that can be reused to reduce memory allocations
type ResponsePool struct {
    pool sync.Pool
}

// Get retrieves a Response from the pool
func (p *ResponsePool) Get() *Response {
    return p.pool.Get().(*Response)
}

// Put returns a Response to the pool after use
// Make sure to clear any sensitive data before putting a Response back
func (p *ResponsePool) Put(resp *Response) {
    if resp == nil {
        return
    }

    // Clear the Response fields before returning to the pool
    resp.Content = ""

    p.pool.Put(resp)
}
```

#### Channel Pool for Stream Processing

The `ChannelPool` manages buffered channels for token streaming, which significantly reduces allocation pressure in high-throughput streaming scenarios:

```go
// ChannelPool is a pool of channels that can be reused
// This significantly reduces GC pressure in streaming operations
type ChannelPool struct {
    pool sync.Pool
}

// GetResponseStream creates a new response stream using the pool
func (p *ChannelPool) GetResponseStream() (ResponseStream, chan Token) {
    ch := p.Get()
    return ch, ch
}
```

#### Message Manager Pool (`pkg/agent/workflow/message_manager.go`)

The `MessageManager` uses a pool to efficiently handle LLM conversation messages:

```go
// Pool for message objects to reduce allocations
messagePool: &sync.Pool{
    New: func() interface{} {
        return &ldomain.Message{}
    },
},
```

### Best Practices for Pooling

1. **Clear object state**: Always reset object state before returning to the pool
2. **Use pointers**: Return pointers to avoid allocations during `Put` operations
3. **Pre-allocation**: Pre-allocate slices with expected capacity
4. **Singleton pools**: Use global singleton pools with `sync.Once` to ensure thread-safety
5. **Size appropriately**: Size channel buffers based on expected token flow (e.g., `ChannelPoolSize = 20`)

## Caching Mechanisms

Go-LLMs implements various caching strategies to avoid redundant computations and improve performance.

### Implementation Examples

#### Response Cache (`pkg/agent/workflow/response_cache.go`)

Caches LLM responses to avoid redundant API calls:

```go
// ResponseCache provides a thread-safe cache for LLM responses
// to avoid redundant API calls for the same input
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

#### Schema Cache (`pkg/structured/processor/schema_cache.go`)

Caches marshaled JSON schema to avoid redundant conversions:

```go
// SchemaCache provides caching for schema JSON to avoid repeated marshaling
type SchemaCache struct {
    lock  sync.RWMutex
    cache map[uint64][]byte
}
```

#### Parameter Type Cache (`pkg/agent/tools/param_cache.go`)

Caches reflection-based type information to speed up parameter conversions:

```go
// parameterTypeCache caches reflection type information to reduce allocations
// during repeated tool executions with the same parameter types
type parameterTypeCache struct {
    // structFieldCache maps struct types to field information to avoid repeated reflection lookups
    structFieldCache sync.Map // map[reflect.Type][]fieldInfo

    // parameterConversionCache caches common conversion patterns
    parameterConversionCache sync.Map // map[typePair]bool
}
```

#### Regex Cache (`pkg/schema/validation/validator.go`)

Caches compiled regular expressions to avoid recompilation:

```go
// RegexCache stores compiled regular expressions to avoid recompilation
var RegexCache = sync.Map{}
```

### Caching Best Practices

1. **Hash-based keys**: Use deterministic hashing for cache keys
2. **Atomic operations**: Ensure thread-safety with mutex or sync.Map
3. **Bounded size**: Implement capacity limits and eviction policies
4. **TTL**: Use time-based expiration for potentially stale data
5. **Cache statistics**: Track hit rates and performance metrics

## Concurrency Patterns

Go-LLMs implements several concurrency patterns to handle parallel execution efficiently.

### Thread-Safety

Thread-safety is ensured through:

1. **Mutex locks**: Using `sync.RWMutex` for reader/writer separation
2. **Atomic initialization**: Using `sync.Once` for safe singleton initialization
3. **Immutable outputs**: Returning copies of values to prevent modification
4. **Synchronized pools**: Using thread-safe object pools

Example from `MessageManager`:

```go
// Thread safety
mu sync.RWMutex

// GetMessages returns the current message history
func (m *MessageManager) GetMessages() []ldomain.Message {
    m.mu.RLock()
    defer m.mu.RUnlock()

    // Return a deep copy to prevent external modification
    result := make([]ldomain.Message, len(m.messages))
    for i, msg := range m.messages {
        result[i] = m.cloneMessage(msg)
    }

    return result
}
```

### Streaming with Channels

The library uses Go channels for efficient token streaming from LLM providers:

```go
// In channel_pool_bench_test.go
b.Run("RealisticStreamingWithPooling", func(b *testing.B) {
    channelPool := domain.GetChannelPool()
    tokenPool := domain.GetTokenPool()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        _, ch := channelPool.GetResponseStream()

        go func() {
            defer close(ch)
            // Stream tokens...
        }()

        // Consume tokens
        var fullText string
        for token := range ch {
            fullText += token.Text
        }
    }
})
```

### Performance Considerations

1. **Buffered channels**: Use buffered channels to prevent blocking
2. **Context cancellation**: Implement timeouts and cancellation
3. **Non-blocking selects**: Use `select` with timeouts for non-blocking operations
4. **Resource cleanup**: Always close channels and release resources properly
5. **Backpressure**: Implement flow control mechanisms for high-throughput scenarios

## Optimized Validation

The validation system uses several optimizations to improve performance:

### Fast Path Optimizations

```go
// Fast path for nil schema
if schema == nil {
    return errors
}

// Fast path for nil values
if value == nil {
    return false
}

// Fast path for identical types
switch a := a.(type) {
case string:
    if b, ok := b.(string); ok {
        return a == b
    }
}
```

### Type Coercion Optimizations

```go
// optimizedConvertValue attempts to convert a value to the target type
// This version is optimized to reduce allocations
func optimizedConvertValue(value reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
    // Fast path: if directly assignable, return as is
    if value.Type().AssignableTo(targetType) {
        return value, true
    }
    
    // Type-specific optimizations...
}
```

## Benchmarking

The library includes comprehensive benchmarks to measure performance improvements:

- `benchmarks/channel_pool_bench_test.go`: Tests channel pooling performance
- `benchmarks/json_marshaling_bench_test.go`: Measures JSON serialization performance
- `benchmarks/schema_bench_test.go`: Evaluates schema validation efficiency

## Performance Monitoring

Hooks are implemented to track performance metrics:

```go
// MetricsHook collects performance metrics for agent operations
type MetricsHook struct {
    TotalCalls        int
    TotalLatency      time.Duration
    ResponseSizes     []int
    ToolExecutions    int
    ToolExecutionTime time.Duration
    // Additional metrics...
}
```

## Conclusion

Go-LLMs uses a comprehensive set of performance optimization techniques:

1. **Memory efficiency** via object pooling and reuse
2. **Reduced CPU usage** through caching and fast paths
3. **Concurrency management** with thread-safe data structures
4. **Resource control** with bounded capacity and TTL mechanisms
5. **Performance monitoring** with hooks and metrics

These optimizations ensure high throughput and low latency even under heavy load, making the library suitable for production applications interacting with LLM providers.