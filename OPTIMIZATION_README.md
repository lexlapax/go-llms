# Go-LLMs Performance Optimization

This document describes the performance optimizations implemented in the Go-LLMs library as part of the Phase 7 (Performance Optimization and Refinement) work.

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

## Usage

To use the optimized tools, replace:

```go
tool := tools.NewTool(
    "multiply",
    "Multiply two numbers",
    func(params struct {
        A float64 `json:"a"`
        B float64 `json:"b"`
    }) (map[string]interface{}, error) {
        // Function implementation
    },
    paramSchema,
)
```

With:

```go
tool := tools.NewOptimizedTool(
    "multiply",
    "Multiply two numbers",
    func(params struct {
        A float64 `json:"a"`
        B float64 `json:"b"`
    }) (map[string]interface{}, error) {
        // Same function implementation
    },
    paramSchema,
)
```

For common tools, use the optimized versions:

```go
// Instead of:
webFetchTool := tools.WebFetch()

// Use:
webFetchTool := tools.OptimizedWebFetch()
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

## Future Optimizations

Planned future optimizations include:

1. Improving agent context initialization
2. Adding fast paths for LLM provider message handling
3. Optimizing prompt processing and template expansion
4. Adding more extensive caching for repeated operations

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
```

## Usage - Optimized Schema Validation

To use the optimized schema validator, replace:

```go
validator := validation.NewValidator()
result, err := validator.Validate(schema, jsonData)
```

With:

```go
validator := validation.NewOptimizedValidator()
result, err := validator.Validate(schema, jsonData)
```

The optimized validator is particularly beneficial for:
- Validating complex schemas with many constraints
- Validating strings with patterns or formats
- Validating data with validation errors
- High-frequency validation operations