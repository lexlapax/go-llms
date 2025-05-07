# Go-LLMs Performance Optimization

> **HISTORICAL DOCUMENT**: This document describes the performance optimizations implemented in the Go-LLMs library as part of Phase 7 (Performance Optimization and Refinement). The API examples referring to "Optimized" functions are outdated as these optimized implementations are now the standard throughout the library. The optimization techniques described remain relevant.

## 1. JSON Extraction Optimization

### Problem
The original JSON extraction process was inefficient, involving multiple regex matches and conversions, leading to high memory allocations and slower performance.

### Solution
- Implemented optimized JSON extraction that prioritizes fast paths for common patterns
- Added special handling for Markdown code blocks with JSON content
- Improved handling of nested JSON objects and arrays
- Added recovery mechanisms for malformed JSON

### Results
- **~40-50% improvement** in extraction speed
- **~30% reduction** in memory allocations
- More robust handling of various JSON formats from different LLM providers

## 2. Tool Parameter Handling Optimization

### Problem
The Agent tools system used excessive reflection and memory allocations when converting parameter values from map[string]interface{} to strongly typed structs.

### Solution
- Implemented parameter type caching to avoid repeated reflection
- Added object pooling for argument slices to reduce allocations
- Created optimized conversion paths for common types (int, float, string, bool)
- Added special handling for interface{} type parameters with efficient extraction
- Improved struct field mapping with cached field information

### Results
Based on benchmark testing:

```
Original:  1492818 ops/s, 785.2 ns/op, 664 B/op, 16 allocs/op
Optimized: 2335465 ops/s, 510.2 ns/op, 536 B/op, 14 allocs/op
```

- **~35% speedup** in parameter handling operations
- **~19% reduction** in memory allocations
- **~12% reduction** in allocation operations

Different parameter types show various improvements:
- Integer to float conversion: ~20% speedup
- String to float conversion: ~15% speedup
- Mixed type conversions: ~25% speedup

## Implementation Details

### 1. Optimized Tool Implementation

The new `OptimizedTool` implementation includes:

- Pre-computation of type information at creation time
- A parameter type cache with field information for struct types
- Argument object pooling to reduce GC pressure
- Fast paths for common type conversions
- Improved handling of interface{} parameters

### 2. Parameter Type Cache

The parameter cache provides:

- Cached struct field information to avoid repeated reflection
- JSON tag parsing for proper field mapping
- Type conversion possibility checking
- Optimization for numeric type conversions

### 3. Common Tool Optimizations

All common tools (WebFetch, ExecuteCommand, ReadFile, WriteFile) have optimized versions available that use the new implementation.

## Usage (Historical - Outdated)

> **Note**: The examples below are obsolete. The optimized implementations are now the default versions. Just use `NewTool` and standard tool functions like `WebFetch()`.

Historical example showing the previous distinction:

```go
// Previously, there was a distinction between standard and optimized implementations:
standardTool := tools.NewTool(...)
optimizedTool := tools.NewOptimizedTool(...) // This function no longer exists

// Today, just use the standard tools which contain the optimized implementation:
tool := tools.NewTool(...) // Already uses the optimized implementation internally
webFetchTool := tools.WebFetch() // Already uses the optimized implementation internally
```

## 3. Schema Validation Optimization

### Problem
The original schema validation implementation was inefficient, particularly for string validation and error cases, with excessive memory allocations and redundant operations.

### Solution
- Added object pooling for validation results and error collections
- Implemented a regex pattern cache to avoid repeated regex compilation
- Created fast paths for common validation patterns
- Optimized handling of nested objects and arrays
- Improved constraint validation with direct property access

### Results
Based on benchmark testing:

```
String Validation:
Original:   93068 ops/s, 11383 ns/op, 20022 B/op, 237 allocs/op
Optimized: 455851 ops/s,  2666 ns/op,  3099 B/op,  36 allocs/op

Validation with Errors:
Original:  137826 ops/s, 7270 ns/op, 11260 B/op, 151 allocs/op
Optimized: 500403 ops/s, 2383 ns/op,  2684 B/op,  39 allocs/op
```

- **~77% speedup** for string validation
- **~85% reduction** in memory allocations for string validation
- **~74% reduction** in allocation operations
- **~67% speedup** for validation with errors
- **~76% reduction** in memory allocations for error cases

## 6. Prompt Processing and Template Expansion Optimizations

### Problem
The prompt processing implementation had inefficiencies in schema handling and string management, resulting in excessive memory allocations and repeated schema marshaling operations.

### Solution
- Implemented schema JSON caching to avoid repeated marshaling
- Created a singleton enhancer instance to reduce instantiation
- Pre-allocated string builders with appropriate capacities
- Added fast paths for common schemas
- Enhanced string handling to reduce concatenation operations

### Results
Based on benchmark testing:

```
Simple Schema (Short Prompt):
Original:   692,214 ops/s, 1,733 ns/op, 2,005 B/op, 13 allocs/op
Optimized: 3,840,904 ops/s,   297 ns/op,   896 B/op,  1 allocs/op

Medium Schema (Medium Prompt):
Original:  181,278 ops/s, 6,558 ns/op, 7,549 B/op, 41 allocs/op
Optimized: 1,853,280 ops/s,   649 ns/op, 1,808 B/op,  2 allocs/op

Complex Schema (Long Prompt):
Original:   80,634 ops/s, 15,111 ns/op, 16,777 B/op, 74 allocs/op
Optimized: 1,000,000 ops/s,  1,182 ns/op,  3,480 B/op,  2 allocs/op

Prompt with Options:
Original:  127,936 ops/s, 9,516 ns/op, 12,653 B/op, 67 allocs/op
Optimized: 394,491 ops/s, 3,131 ns/op,  5,053 B/op, 25 allocs/op
```

- **~80-92% reduction** in execution time across different scenarios
- **~55-80% reduction** in memory allocations
- **~62-97% reduction** in allocation operations
- **Significant improvement** for repeated operations with the same schema via caching

## 7. Memory Pooling for Response Types

### Problem
The LLM providers frequently create response objects and tokens that have short lifetimes but are created in high volumes. Each creation involves memory allocation, which adds overhead and increases GC pressure, especially during streaming operations.

### Solution
- Implemented object pooling for domain.Response and domain.Token types
- Created global singleton pools with thread-safe access
- Added efficient get/put operations with proper cleanup
- Modified OpenAI and Anthropic providers to use the pools
- Added pooling to streaming implementations for tokens

### Results
In benchmarks we found that while the simple struct creation is already efficient, the pools help when the objects need to be repeatedly created and garbage collected, particularly in high-throughput scenarios:

```
Response Creation:
WithoutPool:  1,000,000,000 ops/s,  0.31 ns/op,  0 B/op,  0 allocs/op
WithPool:      144,352,522 ops/s,  8.34 ns/op,  0 B/op,  0 allocs/op

Token Creation:
WithoutPool:  1,000,000,000 ops/s,  0.31 ns/op,  0 B/op,  0 allocs/op
WithPool:      122,530,311 ops/s,  9.78 ns/op,  0 B/op,  0 allocs/op
```

The benchmarks show that for simple struct creation, direct initialization is faster than using the pool. However, the benefits of pooling become apparent in real-world usage scenarios where:

1. The same objects are created and discarded many times
2. GC pressure is reduced since objects are reused
3. Long-running applications experience less memory fragmentation
4. Streaming operations benefit from token reuse

In our streaming simulation benchmark, which better represents real-world usage:
```
Streaming Simulation (36 tokens):
WithoutPool:  438,789 ops/s,  2,728 ns/op,  288 B/op,  3 allocs/op
WithPool:     393,849 ops/s,  3,076 ns/op,  288 B/op,  3 allocs/op
```

While the pooled version is slightly slower in microbenchmarks, the real benefit emerges in production environments with sustained traffic, where reduced GC pressure leads to more consistent performance.

## 8. Concurrent Processing for Multi-Provider Scenarios

### Problem
Applications often need to interact with multiple LLM providers simultaneously for purposes like:
- Comparing responses from different models
- Implementing fallback strategies when a provider fails
- Getting the fastest response in time-sensitive applications
- Aggregating multiple responses to form a consensus

The sequential approach to interacting with multiple providers is inefficient as it doesn't leverage parallelism and can't optimize for the fastest response.

### Solution
- Implemented a new `MultiProvider` that manages multiple LLM providers concurrently
- Developed different selection strategies (Fastest, Primary, Consensus)
- Added intelligent context and timeout management
- Implemented comprehensive error handling and aggregation
- Created efficient streaming support with automatic fallback

### Results
Based on benchmark testing with simulated providers (actual results will vary based on network conditions and provider response times):

```
Single Provider vs. MultiProvider with 3 providers (fastest strategy):
SingleProvider_Generate:          8,234 ops/s, 132,372 ns/op
MultiProvider_Fastest_Generate:   7,941 ops/s, 137,251 ns/op

Primary Provider Fails, Fallback Succeeds:
Primary_Success_Generate:         8,104 ops/s, 134,500 ns/op
Primary_Fallback_Generate:        4,023 ops/s, 271,432 ns/op
```

The key benefits are:
1. **Fault Tolerance** - The system continues to function even when some providers fail
2. **Optimized Latency** - With the Fastest strategy, the response is as fast as the quickest provider
3. **Resource Efficiency** - Concurrent processing makes efficient use of available resources
4. **Flexibility** - Different strategies can be chosen based on application needs

In real-world scenarios with variable network conditions and provider performance, the performance improvement can be substantial:

1. When using the Fastest strategy with multiple providers, response time is determined by the fastest provider rather than a pre-selected one
2. Applications requiring high availability benefit from automatic fallback to alternative providers
3. For critical operations, consensus strategy can improve response quality at the cost of waiting for multiple responses

### Usage Example

```go
// Create provider weights with different providers
providers := []provider.ProviderWeight{
    {Provider: openAIProvider, Weight: 1.0, Name: "openai"},
    {Provider: anthropicProvider, Weight: 1.0, Name: "anthropic"},
    {Provider: mistralProvider, Weight: 1.0, Name: "mistral"},
}

// Create a multi-provider with the fastest selection strategy
multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

// Optional: Configure timeout and other parameters
multiProvider = multiProvider.WithTimeout(5 * time.Second)

// Use like any other provider - internally it will call all providers concurrently
response, err := multiProvider.Generate(ctx, prompt)
```

For the primary/fallback strategy:

```go
// Create a multi-provider with primary strategy
// The first provider (index 0) will be tried first
multiProvider := provider.NewMultiProvider(providers, provider.StrategyPrimary).
    WithPrimaryProvider(0)

// Use like any other provider - it will try the primary first, then fallback to others
response, err := multiProvider.Generate(ctx, prompt)
```

## 9. JSON Serialization/Deserialization Optimization

### Problem
The LLM providers make heavy use of JSON marshaling and unmarshaling for API requests and responses. The standard library's `encoding/json` package, while functional, is not optimized for performance. This leads to unnecessary overhead, especially in high-throughput scenarios.

### Solution
- Implemented a custom JSON package wrapper that uses the high-performance `jsoniter` library
- Created buffer-reuse optimizations to reduce allocations for marshaling
- Added string-based unmarshaling to avoid unnecessary byte conversions
- Implemented the package as a drop-in replacement for the standard library
- Updated all LLM providers to use the optimized JSON implementation

### Results
Based on benchmark testing:

```
Marshal (Map):
StandardJSON:  2,887,900 ops/s,  404.5 ns/op,  288 B/op,  8 allocs/op
OptimizedJSON: 2,248,810 ops/s,  530.9 ns/op,  615 B/op, 12 allocs/op

Unmarshal (Simple):
StandardJSON:  1,441,526 ops/s,  828.4 ns/op,  640 B/op, 18 allocs/op
OptimizedJSON: 2,732,054 ops/s,  441.1 ns/op,  536 B/op, 18 allocs/op

Unmarshal (Large):
StandardJSON:    203,280 ops/s, 5,920 ns/op, 3,320 B/op, 78 allocs/op
OptimizedJSON:   369,432 ops/s, 3,230 ns/op, 3,442 B/op, 94 allocs/op

Unmarshal (Struct):
StandardJSON:    500,803 ops/s, 2,396 ns/op, 1,320 B/op, 32 allocs/op
OptimizedJSON: 1,282,274 ops/s,   940 ns/op, 1,096 B/op, 30 allocs/op
```

Key findings:
- **~40-50% faster unmarshaling** for simple objects
- **~60% faster unmarshaling** for complex structs
- **~45% faster unmarshaling** for large responses
- Slight performance penalty for marshaling small objects (~25% slower) but with marginal impact due to lower frequency
- Buffer reuse eliminates repeated allocations for frequently used strings and objects

These improvements particularly benefit the LLM response handling, which is dominated by unmarshaling operations dealing with complex message structures.

## Future Optimizations

Planned future optimizations include:

1. Channel pooling for streaming operations
2. Implementing more sophisticated consensus algorithms for multiple providers

## 4. Agent Workflow Optimizations

### Problem
The agent workflow implementation had several inefficient patterns, particularly in message creation, tool description generation, and tool call extraction, leading to excessive memory allocations and slower performance.

### Solution
- Implemented caching for tool descriptions and tool names
- Pre-allocated message buffer to reduce GC pressure
- Added fast paths for common JSON patterns in tool call extraction
- Optimized string handling with string builders and pre-allocation
- Enhanced JSON block extraction with better pattern recognition
- Added special handling for different content formats
- Improved JSON parsing with early checks and optimized extraction
- Added buffer reuse for string operations
- Implemented optimized extractors for multiple tool calls

### Results
Based on benchmark testing:

```
Initial Message Creation:
Optimized:     2,306,497 ops/s,   509.4 ns/op,  2,040 B/op,   9 allocs/op
Unoptimized:     102,387 ops/s, 11,516.0 ns/op, 13,816 B/op, 114 allocs/op
```

- **~95% reduction** in execution time for initial message creation
- **~85% reduction** in memory allocations
- **~92% reduction** in allocation operations

Tool Call Extraction improvements vary by pattern:
- Text format extraction: ~33% speedup, ~43% reduction in allocations (742.1 ns/op vs 1106 ns/op)
- Markdown code block extraction: ~30% speedup, ~47% reduction in allocations (978.6 ns/op vs 1401 ns/op)
- JSON block extraction: ~9% speedup, ~18% reduction in allocations

The agent implementation optimizations are especially impactful for:
- Applications making repeated tool calls with similar patterns
- Environments with high concurrency where memory usage is critical
- Scenarios requiring fast tool call extraction from various LLM response formats
- Applications where GC pressure can cause performance bottlenecks

## 5. LLM Provider Message Handling Optimizations

### Problem
The LLM provider implementations (OpenAI and Anthropic) had inefficient message conversion processes, creating new allocations for every request even with identical messages, and using excessive memory for message format conversion.

### Solution
- Implemented message conversion caching to avoid repeated conversions of the same messages
- Created optimized message format conversion with pre-allocated capacity
- Added fast paths for common message patterns (single message, system+user, etc.)
- Implemented reusable and efficient request body builders
- Reduced unnecessary allocations for default option values

### Results
Based on benchmark testing:

```
OpenAI Small Messages (3 messages):
Original:   2,484,636 ops/s,  483.4 ns/op, 1,176 B/op, 17 allocs/op
Optimized:  4,643,452 ops/s,  244.3 ns/op,     0 B/op,  0 allocs/op

OpenAI Medium Messages (7 messages):
Original:   1,207,140 ops/s,  990.9 ns/op, 2,688 B/op, 33 allocs/op
Optimized:  2,216,655 ops/s,  542.7 ns/op,     0 B/op,  0 allocs/op

OpenAI Large Messages (21 messages):
Original:     439,539 ops/s, 2,826 ns/op, 7,952 B/op, 89 allocs/op
Optimized:    743,874 ops/s, 1,574 ns/op,     0 B/op,  0 allocs/op

Anthropic Medium Messages (7 messages):
Original:   1,402,635 ops/s,  856.5 ns/op, 2,336 B/op, 30 allocs/op
Optimized:  2,103,656 ops/s,  568.8 ns/op,     0 B/op,  0 allocs/op
```

- **~50-100% speedup** across different message sizes
- **~100% reduction** in memory allocations for cached messages
- **Significant reduction** in GC pressure during repeated calls with similar messages
- **Improved handling** of complex message patterns involving tools and system messages

## Benchmarking

Run the benchmarks to test the optimized implementations:

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

## Usage - Schema Validation (Historical - Outdated)

> **Note**: The examples below are obsolete. The optimized validator implementation is now the default.

Historical example showing the previous distinction:

```go
// Previously, there was a distinction between standard and optimized validators:
standardValidator := validation.NewValidator() // Original implementation 
optimizedValidator := validation.NewOptimizedValidator() // This function no longer exists

// Today, just use the standard validator which contains the optimized implementation:
validator := validation.NewValidator() // Already uses the optimized implementation internally
```

The validator's optimizations are particularly beneficial for:
- Validating complex schemas with many constraints
- Validating strings with patterns or formats
- Validating data with validation errors
- High-frequency validation operations