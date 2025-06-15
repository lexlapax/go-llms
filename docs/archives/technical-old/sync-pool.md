# Sync.Pool Implementation Guide

> **[Documentation Home](/REFERENCE.md) / [Technical Documentation](/docs/technical/) / Sync.Pool Implementation**

This document provides a deep dive into how Go-LLMs uses `sync.Pool` for memory optimization, including implementation details, best practices, and usage patterns.

*Related: [Performance Optimization](performance.md) | [Caching Mechanisms](caching.md) | [Concurrency Patterns](concurrency.md)*

## Understanding sync.Pool

The `sync.Pool` is a built-in Go type that provides a concurrent-safe way to reuse objects and reduce garbage collection (GC) pressure. It offers:

- Thread-safe object reuse
- Automatic cleanup during garbage collection
- Zero initialization cost
- Low-latency object acquisition

## Key sync.Pool Implementations in Go-LLMs

### Response and Token Pools (pkg/llm/domain/pool.go)

Go-LLMs implements several pools for handling response objects and tokens:

```go
// Global singleton response pool
var (
    globalResponsePool *ResponsePool
    responsePoolOnce   sync.Once
)

// GetResponsePool returns the singleton global response pool
func GetResponsePool() *ResponsePool {
    responsePoolOnce.Do(func() {
        globalResponsePool = NewResponsePool()
    })
    return globalResponsePool
}

// NewResponsePool creates a new response pool
func NewResponsePool() *ResponsePool {
    return &ResponsePool{
        pool: sync.Pool{
            New: func() interface{} {
                // Create a new Response when the pool is empty
                return &Response{}
            },
        },
    }
}
```

This pattern provides:
1. A thread-safe singleton pool via `sync.Once`
2. Lazy initialization (created only when needed)
3. Encapsulation of pool operations within a type

### Channel Pool for Streaming (pkg/llm/domain/pool.go)

The `ChannelPool` efficiently manages channels used for token streaming:

```go
// ChannelPoolSize is the default buffer size for channels from the pool
const ChannelPoolSize = 20

// ChannelPool is a pool of channels that can be reused to reduce memory allocations
// This significantly reduces GC pressure in streaming operations
type ChannelPool struct {
    pool sync.Pool
}

// NewChannelPool creates a new channel pool
func NewChannelPool() *ChannelPool {
    return &ChannelPool{
        pool: sync.Pool{
            New: func() interface{} {
                // Create a new channel when the pool is empty
                // Use a buffered channel with a reasonable size to prevent blocking
                return make(chan Token, ChannelPoolSize)
            },
        },
    }
}

// Get retrieves a channel from the pool
func (p *ChannelPool) Get() chan Token {
    return p.pool.Get().(chan Token)
}

// Put returns a channel to the pool after use
// Make sure the channel is empty and not closed before putting it back
func (p *ChannelPool) Put(ch chan Token) {
    if ch == nil {
        return
    }

    // Drain any remaining tokens to ensure the channel is empty
    // This is a non-blocking operation
    for {
        select {
        case _, ok := <-ch:
            if !ok {
                // Channel is closed, don't put it back
                return
            }
        default:
            // Channel is empty
            p.pool.Put(ch)
            return
        }
    }
}
```

Key optimizations:
1. Uses buffered channels with proper size to prevent blocking
2. Cleans channels before returning them to the pool
3. Safe handling of closed channels

### Validator Pool (pkg/schema/validation/validator.go)

The validator uses multiple pools to optimize validation operations:

```go
// Validator implements schema validation with performance enhancements
type Validator struct {
    // errorBufferPool provides reusable string buffers for errors
    // Uses pointers to slices to avoid allocations during Put
    errorBufferPool sync.Pool

    // validationResultPool provides reusable validation results
    validationResultPool sync.Pool
    
    // enableCoercion controls whether the validator attempts to coerce values to the expected type
    enableCoercion bool
}

// NewValidator creates a new validator with performance enhancements
func NewValidator(options ...func(*Validator)) *Validator {
    v := &Validator{
        errorBufferPool: sync.Pool{
            New: func() interface{} {
                // Preallocate a slice with reasonable capacity to avoid reallocation
                // Return a pointer to avoid allocations during Put
                slice := make([]string, 0, 8)
                return &slice
            },
        },
        validationResultPool: sync.Pool{
            New: func() interface{} {
                return &domain.ValidationResult{
                    Valid:  true,
                    Errors: make([]string, 0, 8),
                }
            },
        },
    }
    
    // Apply options
    for _, option := range options {
        option(v)
    }
    
    return v
}
```

Advanced optimization techniques:
1. Pre-allocation of slices with expected capacity
2. Using pointers to slices to avoid allocations during `Put` operations
3. Proper reset of objects before reuse

## Best Practices for sync.Pool

### 1. Always Clear Object State

Before returning an object to the pool, reset its state to avoid data leaks and unexpected behavior:

```go
// Put returns a Response to the pool after use
func (p *ResponsePool) Put(resp *Response) {
    if resp == nil {
        return
    }

    // Clear the Response fields before returning to the pool
    resp.Content = ""

    p.pool.Put(resp)
}
```

### 2. Use Pointers for Slices and Maps

When pooling objects that contain slices or maps, use pointers to avoid unnecessary allocations:

```go
// Get an error buffer from the pool (pointer to slice)
errorsPtr := v.errorBufferPool.Get().(*[]string)
errors := *errorsPtr
errors = errors[:0] // Reset slice but keep capacity

// ... use the slice ...

// Update the pointer's underlying slice
*errorsPtr = errors

// Return the error buffer to the pool
v.errorBufferPool.Put(errorsPtr)
```

### 3. Pre-allocate with Appropriate Capacity

Pre-allocate slices and maps with a reasonable capacity to avoid resizing:

```go
// Preallocate a slice with reasonable capacity to avoid reallocation
slice := make([]string, 0, 8)
return &slice
```

### 4. Implement Thread-Safe Singletons

Use `sync.Once` to ensure thread-safe initialization of global pools:

```go
var (
    globalChannelPool *ChannelPool
    channelPoolOnce   sync.Once
)

func GetChannelPool() *ChannelPool {
    channelPoolOnce.Do(func() {
        globalChannelPool = NewChannelPool()
    })
    return globalChannelPool
}
```

### 5. Handle Special Resource Types Correctly

For resources like channels, ensure proper cleanup before pooling:

```go
// Put returns a channel to the pool after use
func (p *ChannelPool) Put(ch chan Token) {
    if ch == nil {
        return
    }

    // Drain any remaining tokens
    for {
        select {
        case _, ok := <-ch:
            if !ok {
                // Channel is closed, don't put it back
                return
            }
        default:
            // Channel is empty
            p.pool.Put(ch)
            return
        }
    }
}
```

### 6. Consider Pool Size and GC Behavior

Be aware that:
- `sync.Pool` can be emptied during garbage collection
- Pools should be sized appropriately for your workload
- Very large pools can increase memory consumption

## Advanced Pool Techniques

### Cascading Object Creation

For complex objects, implement cascading pool gets to avoid nested allocations:

```go
func (factory *ObjectFactory) GetComplexObject() *ComplexObject {
    obj := factory.objectPool.Get().(*ComplexObject)
    
    // Get child objects from their respective pools
    obj.ChildA = factory.childAPool.Get().(*ChildA)
    obj.ChildB = factory.childBPool.Get().(*ChildB)
    
    return obj
}

func (factory *ObjectFactory) PutComplexObject(obj *ComplexObject) {
    if obj == nil {
        return
    }
    
    // Return child objects to their pools
    if obj.ChildA != nil {
        factory.childAPool.Put(obj.ChildA)
        obj.ChildA = nil
    }
    
    if obj.ChildB != nil {
        factory.childBPool.Put(obj.ChildB)
        obj.ChildB = nil
    }
    
    // Clear and return the parent object
    factory.objectPool.Put(obj)
}
```

### Pool Performance Monitoring

To optimize pool performance, implement monitoring:

```go
type MonitoredPool struct {
    pool       sync.Pool
    gets       int64
    puts       int64
    misses     int64
    hitRatio   float64
}

func (p *MonitoredPool) Get() interface{} {
    atomic.AddInt64(&p.gets, 1)
    obj := p.pool.Get()
    
    // A newly created object indicates a miss
    if obj.(*poolObject).isNew {
        atomic.AddInt64(&p.misses, 1)
        obj.(*poolObject).isNew = false
    }
    
    // Update hit ratio
    p.hitRatio = 1 - (float64(p.misses) / float64(p.gets))
    
    return obj
}
```

## Pool Usage in Benchmarks

Go-LLMs includes benchmarks that demonstrate pool performance benefits:

```go
// BenchmarkChannelPooling benchmarks the channel pooling for streaming operations
func BenchmarkChannelPooling(b *testing.B) {
    // Benchmark without pooling
    b.Run("StreamingWithoutPool", func(b *testing.B) {
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            ch := make(chan domain.Token, 20)
            // Use the channel...
        }
    })

    // Benchmark with pooling
    b.Run("StreamingWithPool", func(b *testing.B) {
        pool := domain.GetChannelPool()
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            _, ch := pool.GetResponseStream()
            // Use the channel...
        }
    })
}
```

## Common Pitfalls to Avoid

1. **Reusing without clearing**: Always reset objects before returning them to the pool
2. **Pooling tiny objects**: Don't pool small, cheap-to-create objects (strings, ints, etc.)
3. **Closing pooled channels**: Never close a channel unless you won't return it to the pool
4. **Excessive pooling**: Don't pool everything; focus on frequently allocated objects
5. **Thread-unsafe operations**: Ensure all pool operations are thread-safe

## When to Use Object Pooling

Use object pooling when:

1. **Objects are expensive to create**: Complex structs, slices, maps, etc.
2. **Allocation rate is high**: Many objects created and discarded rapidly
3. **GC pressure is a concern**: Visible GC pauses affect your application
4. **Objects have clean lifecycle**: Objects can be easily reset and reused

## Conclusion

The judicious use of `sync.Pool` in Go-LLMs significantly reduces memory allocations and GC pressure, especially in high-throughput scenarios like token streaming. By following the best practices outlined in this guide, you can implement efficient object pooling in your own code that interacts with the library.