# Profiling Example

This example demonstrates how to use the profiling utilities in the Go-LLMs project.

## Features

- CPU and memory profiling of operations
- Component-specific profiling
- Profiling output to file
- Environment variable control of profiling

## Running the Example

```bash
# Build the example
make build-example EXAMPLE=profiling

# Run the example
./bin/profiling
```

## How It Works

The example demonstrates:

1. Enabling profiling via environment variables
2. Setting custom profile output directories
3. Profiling specific operations with the global profiler
4. Component-specific profiling with separate profilers
5. Writing CPU and memory profiles to files

## Using Profiling in Your Code

To add profiling to your components:

```go
import "github.com/lexlapax/go-llms/pkg/util/profiling"

// Profile a structured operation
result, _ := profiling.ProfileStructuredOp(ctx, profiling.OpStructuredExtraction, func(ctx context.Context) (interface{}, error) {
    // Your operation code here
    return extractedData, nil
})

// Profile a schema operation
result, _ := profiling.ProfileSchemaOp(ctx, profiling.OpSchemaValidation, func(ctx context.Context) (interface{}, error) {
    // Your validation code here
    return validationResult, nil
})

// Profile a pool operation
result, _ := profiling.ProfilePoolOp(ctx, profiling.OpPoolAllocation, func(ctx context.Context) (interface{}, error) {
    // Your pool allocation code here
    return allocatedObject, nil
})
```

## Analyzing Profiles

After running code with profiling enabled, you can analyze the profile files using:

```bash
go tool pprof [profile_file]
```

This opens an interactive pprof session where you can visualize and analyze the profile data.