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

Caches marshaled JSON schema to avoid redundant conversions:

```go
// SchemaCache provides caching for schema JSON to avoid repeated marshaling
type SchemaCache struct {
    lock  sync.RWMutex
    cache map[uint64][]byte
}
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

Prompt processing and template expansion optimization includes:

- Schema JSON caching
- Singleton enhancer instance
- Pre-allocated string builders
- Fast paths for common schemas

```go
// Prompt processing optimization
Original:    692,214 ops/s, 1,733 ns/op, 2,005 B/op, 13 allocs/op
Optimized: 3,840,904 ops/s,   297 ns/op,   896 B/op,  1 allocs/op
```

## Benchmarking

The library includes comprehensive benchmarks to measure performance improvements:

```bash
# Run parameter handling benchmarks
go test -bench=ToolParameterHandling ./benchmarks/... -benchmem

# Run all tool optimization benchmarks
go test -bench=. ./benchmarks/optimized_tools_bench_test.go -benchmem

# Run JSON extraction benchmarks
go test -bench=. ./benchmarks/json_extractor_bench_test.go -benchmem

# Run schema validation benchmarks
go test -bench=. ./benchmarks/optimized_schema_bench_test.go -benchmem

# Run agent context initialization benchmarks
go test -bench=BenchmarkAgentContextInit ./benchmarks/... -benchmem

# Run agent tool call extraction benchmarks
go test -bench=BenchmarkAgentToolExtraction ./benchmarks/... -benchmem

# Run provider message handling benchmarks
go test -bench=BenchmarkProviderMessageConversion ./benchmarks/... -benchmem
```

For more detailed information on specific optimization strategies, see:

- [Sync.Pool Implementation Guide](sync-pool.md)
- [Caching Mechanisms](caching.md)
- [Concurrency Patterns](concurrency.md)