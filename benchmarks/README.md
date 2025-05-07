# Go-LLMs Benchmarks and Performance Optimization

This directory contains benchmarks and profiling tools to measure and optimize the performance of Go-LLMs.

## Performance Evaluation Strategy

The benchmarking and optimization process follows these steps:

1. **Establish baselines**: Measure current performance of key components
2. **Identify bottlenecks**: Find the slowest parts of the system
3. **Optimize critical paths**: Improve performance of bottlenecks
4. **Verify improvements**: Compare new performance to baselines
5. **Monitor regressions**: Maintain performance gains

## Running Benchmarks

Use the Makefile to run benchmarks:

```bash
# Run all benchmarks
make benchmark

# Run benchmarks for a specific package
make benchmark-pkg PKG=schema/validation
```

## Key Benchmark Scenarios

The benchmarks focus on these key components:

1. **Schema validation**: Measuring validation performance for different schema types and data sizes
2. **JSON extraction**: Testing performance of JSON extraction from LLM responses
3. **Provider requests**: Benchmarking request/response handling for providers
4. **Agent workflows**: Measuring agent performance with different complexity levels
5. **Tool execution**: Benchmarking tool invocation and response processing

## Profiling

Go's built-in profiling tools are used to identify performance bottlenecks:

```bash
# CPU profiling
go test -cpuprofile cpu.prof -bench=. ./...

# Memory profiling
go test -memprofile mem.prof -bench=. ./...

# Block profiling (for concurrency issues)
go test -blockprofile block.prof -bench=. ./...
```

View profiling results with:
```bash
go tool pprof cpu.prof
```

## Performance Targets

The performance targets for key operations are:

- Schema validation: < 10ms for typical schemas
- JSON extraction: < 5ms for typical responses
- Provider request preparation: < 1ms
- End-to-end agent execution: < 100ms (excluding external API call time)