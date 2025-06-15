# Logging in Go-LLMs

This document describes the comprehensive logging approach used throughout the go-llms project, including the structured agent hook system and conditional debug logging infrastructure.

## Core Principles

1. **Library code should not log** - Following Go best practices, library code in `pkg/` returns errors rather than logging them
2. **Logging is the caller's responsibility** - Users of the library decide how to handle errors and what to log
3. **Examples demonstrate logging patterns** - Example programs show recommended logging approaches
4. **Performance over convenience** - Logging is avoided in hot paths; debug builds are used when detailed logging is needed
5. **Thread safety** - All logging approaches used are concurrent-safe
6. **Structured monitoring** - Agent hooks provide structured logging and metrics collection
7. **Zero-overhead debug** - Debug logging has zero runtime cost in production builds

## Logging by Component Type

### Core Library Code (`pkg/`)

The core library follows the Go convention of not imposing logging on users:

- **No direct logging**: Functions return errors with context using `fmt.Errorf()`
- **No stdout/stderr output**: The library never prints directly
- **No logging dependencies**: Users aren't forced to use specific logging libraries

**Exception - Agent Hooks**: The library provides comprehensive hook system in `pkg/agent/core/` that users can add to agents for structured logging and monitoring using `slog`.

Example:
```go
// Library code returns errors with context
func (p *Provider) Generate(ctx context.Context, prompt string) (string, error) {
    resp, err := p.client.Complete(ctx, prompt)
    if err != nil {
        return "", fmt.Errorf("provider generate failed: %w", err)
    }
    return resp.Content, nil
}
```

### Command Line Tools (`cmd/`)

CLI tools use direct output for user interaction:

- `fmt.Printf/Println` for normal output
- `fmt.Fprintf(os.Stderr, ...)` for error messages
- Exit codes to indicate success (0) or failure (non-zero)

Example:
```go
func main() {
    result, err := processCommand()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Success: %s\n", result)
}
```

### Example Programs (`cmd/examples/`)

Examples use the standard `log` package for simplicity:

- `log.Printf/Println` for informational output
- `log.Fatalf` for fatal errors
- Agent examples may use `slog` when demonstrating the LoggingHook

Example:
```go
func main() {
    log.Println("Starting example...")
    
    provider := llmutil.NewProvider("openai", apiKey, model)
    result, err := provider.Generate(ctx, "Hello, world!")
    if err != nil {
        log.Fatalf("Generation failed: %v", err)
    }
    
    log.Printf("Result: %s", result)
}
```

### Test Code (`*_test.go`)

Tests use the standard testing package logging:

- `t.Logf()` for debug output (only shown with -v flag)
- `t.Errorf()` for test failures
- `t.Fatalf()` for fatal test errors

### Debug Logging Infrastructure

The library includes a sophisticated debug logging infrastructure that provides zero overhead in production builds while enabling detailed debugging when needed.

#### Conditional Compilation Design

The debug system uses Go build tags to provide two completely different implementations:

**With `-tags debug` (Development/Debug builds):**
- Full debug functionality in `pkg/internal/debug/log.go`
- Component-based filtering via `GO_LLMS_DEBUG` environment variable
- Detailed logging with file/line information

**Without debug tags (Production builds):**
- No-op implementations in `pkg/internal/debug/log_nodebug.go`
- Zero runtime overhead - debug calls compile to nothing
- No debug-related dependencies or memory usage

#### Environment Variable Control

Debug logging is controlled by the `GO_LLMS_DEBUG` environment variable:

```bash
# Enable all debug logging
export GO_LLMS_DEBUG=all
export GO_LLMS_DEBUG=*

# Enable specific components
export GO_LLMS_DEBUG=param_cache,schema,agent

# Enable single component
export GO_LLMS_DEBUG=param_cache
```

#### Available Debug Components

Common debug components used throughout the library:

- `param_cache` - Parameter caching and conversion
- `schema` - Schema validation and processing
- `agent` - Agent execution and lifecycle
- `tools` - Tool registration and execution
- `state` - State management operations
- `events` - Event emission and processing
- `hooks` - Hook execution timing
- `auth` - Authentication detection and application

#### How to Enable Debug Logging

1. **Build with debug flag**:
   ```bash
   # Build CLI with debug support
   go build -tags debug -o bin/go-llms-debug ./cmd/main.go
   
   # Use Makefile target
   make build-debug
   
   # Build specific package with debug
   go build -tags debug ./pkg/agent/tools/...
   ```

2. **Test with debug flag**:
   ```bash
   # Run tests with debug logging
   go test -tags debug ./...
   
   # Use Makefile target
   make test-debug
   
   # Test specific package with debug
   GO_LLMS_DEBUG=param_cache go test -tags debug ./pkg/agent/tools/
   ```

3. **Control debug output**:
   ```bash
   # Enable all debug logging during tests
   GO_LLMS_DEBUG=all make test-debug
   
   # Enable specific components for development
   GO_LLMS_DEBUG=param_cache,schema go run -tags debug ./cmd/examples/simple/
   
   # Debug specific workflows
   GO_LLMS_DEBUG=agent,tools ./bin/go-llms-debug chat
   ```

#### Using Debug Logging in Code

```go
import "github.com/lexlapax/go-llms/pkg/internal/debug"

func processParameters(params map[string]interface{}) error {
    debug.Printf("param_cache", "Processing %d parameters", len(params))
    
    for key, value := range params {
        debug.Printf("param_cache", "Processing field: %s = %v", key, value)
    }
    
    debug.Println("param_cache", "Parameter processing complete")
    return nil
}

func validateSchema(schema *Schema) error {
    debug.Println("schema", "Starting schema validation")
    
    if schema == nil {
        debug.Printf("schema", "Validation failed: schema is nil")
        return fmt.Errorf("schema cannot be nil")
    }
    
    debug.Printf("schema", "Schema type: %s, properties: %d", schema.Type, len(schema.Properties))
    return nil
}
```

#### Debug Output Format

Debug logging includes detailed context information:

```
[DEBUG] 2024/01/10 14:30:15 log.go:51: [param_cache] Processing field: operation = add
[DEBUG] 2024/01/10 14:30:15 log.go:65: [schema] Validation started
[DEBUG] 2024/01/10 14:30:15 log.go:51: [agent] Tool call initiated: calculator
```

Format breakdown:
- `[DEBUG]` - Log level indicator
- `2024/01/10 14:30:15` - Timestamp
- `log.go:51` - Source file and line number
- `[param_cache]` - Component name
- Message content

#### Performance Impact

**Production builds (no debug tags):**
```go
// This compiles to absolutely nothing - zero overhead
debug.Printf("component", "message: %s", value)
```

**Debug builds:**
```go
// This checks if component is enabled, then logs if enabled
// Small overhead only when debug is compiled in
debug.Printf("component", "message: %s", value)
```

#### Custom Debug Logger

For advanced use cases, you can replace the default debug logger:

```go
import (
    "log"
    "os"
    "github.com/lexlapax/go-llms/pkg/internal/debug"
)

// Create custom logger with different format
customLogger := log.New(os.Stderr, "[MY-DEBUG] ", log.Ldate|log.Ltime|log.Lmicroseconds)
debug.SetLogger(customLogger)

// Now all debug.Printf/Println calls use the custom logger
```

## Thread Safety and Concurrency

### Guarantees

- **`slog.Logger`**: Designed to be concurrent-safe
- **`log` package**: Thread-safe by default
- **Immutable loggers**: Logger configurations are not modified after creation
- **No shared state**: Each component manages its own logging

### Best Practices for Concurrent Code

1. **Create loggers once**: Initialize during setup, not during runtime
2. **Use structured logging**: Include context fields for correlation
3. **Avoid logging in hot paths**: Use metrics and periodic summaries instead

Example of structured logging in concurrent operations:
```go
logger.Info("processing item",
    slog.String("worker_id", workerID),
    slog.Int("item_id", itemID),
    slog.String("correlation_id", correlationID))
```

### Anti-Patterns to Avoid

```go
// DON'T: Modify global loggers at runtime
var logger = slog.Default()
func ChangeLogLevel(level slog.Level) {
    logger = slog.New(...) // Race condition!
}

// DON'T: Log while holding locks
mu.Lock()
logger.Info("processing...") // Can cause contention
doWork()
mu.Unlock()

// DON'T: Share logging buffers
var buf bytes.Buffer // Shared across goroutines
fmt.Fprintf(&buf, "log: %v", data) // Race condition!
```

## Agent Hook System

The library provides comprehensive hook system for structured logging and metrics collection in agents.

### LoggingHook

The `LoggingHook` provides structured logging with configurable detail levels:

```go
import (
    "log/slog"
    "github.com/lexlapax/go-llms/pkg/agent/core"
)

// Create a structured logger
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// Create an agent with logging
agent := core.NewLLMAgent("my-agent", "AI assistant", provider)
agent.WithHook(core.NewLoggingHook(logger, core.LogLevelDetailed))

// The hook will log:
// - Before/after LLM generation with emojis (🤔/✅/❌)
// - Tool calls and results with emojis (🔧/✅/❌)
// - Message content (configurable detail)
// - Parameters and results (configurable detail)
```

#### LogLevel Configuration

The `LoggingHook` supports three levels of detail:

```go
// LogLevelBasic - Basic information only
core.NewLoggingHook(logger, core.LogLevelBasic)

// LogLevelDetailed - Includes message counts and truncated content
core.NewLoggingHook(logger, core.LogLevelDetailed)

// LogLevelDebug - Full message content, parameters, and results
core.NewLoggingHook(logger, core.LogLevelDebug)
```

**Sample output with LogLevelDetailed:**
```
INFO Generating response emoji=🤔
INFO Message count count=3
INFO Response generated emoji=✅
INFO Response content content=The weather today is sunny with a temperature...
INFO Calling tool tool=weather_api emoji=🔧
DEBUG Tool parameters params=city: London, units: celsius
INFO Tool executed successfully tool=weather_api emoji=✅
DEBUG Tool result result={"temperature": 22, "description": "sunny"}
```

### MetricsHook

The `LLMMetricsHook` collects performance metrics without logging:

```go
// Create metrics hook
metricsHook := core.NewLLMMetricsHook()
agent.WithHook(metricsHook)

// Get metrics after execution
metrics := metricsHook.GetMetrics()
fmt.Printf("Requests: %d\n", metrics.Requests)
fmt.Printf("Tool calls: %d\n", metrics.ToolCalls)
fmt.Printf("Average generation time: %.2fms\n", metrics.AverageGenTimeMs)
fmt.Printf("Error count: %d\n", metrics.ErrorCount)
fmt.Printf("Total tokens: %d\n", metrics.TotalTokens)

// Tool-specific statistics
for tool, stats := range metrics.ToolStats {
    fmt.Printf("Tool %s: %d calls, avg %.2fms\n", tool, stats.Calls, stats.AverageTimeMs)
}
```

### Combining Hooks

Multiple hooks can be used simultaneously:

```go
// Create logger and metrics hooks
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
loggingHook := core.NewLoggingHook(logger, core.LogLevelDetailed)
metricsHook := core.NewLLMMetricsHook()

// Add both hooks to agent
agent := core.NewLLMAgent("assistant", "AI assistant", provider)
agent.WithHook(loggingHook)
agent.WithHook(metricsHook)

// Now you get both structured logging and metrics collection
```

## Error Handling vs Logging

The library follows these patterns for error handling:

1. **Wrap errors with context**: Use `fmt.Errorf("operation failed: %w", err)`
2. **Return errors up the stack**: Let callers decide how to handle them
3. **Provide detailed error messages**: Include relevant context for debugging
4. **Use error types when appropriate**: For errors that callers might want to handle specially

Example:
```go
// Good: Return error with context
if err := validator.Validate(data); err != nil {
    return fmt.Errorf("validation failed for schema %s: %w", schemaID, err)
}

// Bad: Log and return generic error
if err := validator.Validate(data); err != nil {
    log.Printf("Validation error: %v", err) // Don't log in library
    return errors.New("validation failed")   // Lost context
}
```

## Summary

The go-llms logging approach provides a comprehensive system that prioritizes:

### Core Design Principles
- **User control**: Libraries don't impose logging decisions on users
- **Zero overhead**: Production builds have no debug logging overhead
- **Thread safety**: All logging mechanisms are concurrent-safe
- **Structured monitoring**: Agent hooks provide rich observability
- **Component isolation**: Debug logging can be enabled per component

### Key Features

**Agent Hook System:**
- `LoggingHook`: Structured logging with configurable detail levels and emojis
- `LLMMetricsHook`: Performance metrics collection without logging overhead
- Multiple hooks can be combined for comprehensive monitoring
- Integration with `slog` for modern structured logging

**Debug Infrastructure:**
- Conditional compilation with build tags for zero production overhead
- Component-based filtering via environment variables
- Detailed source location information in debug builds
- Custom logger support for advanced use cases

**Component-Specific Patterns:**
- Core library: Error returns with context, no direct logging
- CLI tools: Direct output with fmt package
- Examples: Standard log package for simplicity
- Tests: Built-in testing package logging

### Usage Scenarios

**Development and Debugging:**
```bash
# Enable detailed logging for specific components
GO_LLMS_DEBUG=agent,tools go run -tags debug ./examples/calculator/

# Monitor agent performance with metrics
agent.WithHook(core.NewLLMMetricsHook())
```

**Production Monitoring:**
```go
// Structured logging for production insights
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
agent.WithHook(core.NewLoggingHook(logger, core.LogLevelBasic))
```

**Testing and Integration:**
```bash
# Debug specific test failures
GO_LLMS_DEBUG=param_cache,schema go test -tags debug ./pkg/agent/tools/
```

This comprehensive logging design allows go-llms to integrate seamlessly into any Go application while providing powerful debugging capabilities for development and structured monitoring for production environments.