# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go-LLMs is a Go library that provides a unified interface to interact with various LLM providers (OpenAI, Anthropic, Google Gemini, etc.) with robust data validation and agent tooling. Key features include structured output processing, a consistent provider interface, agent workflows, and multi-provider strategies.

**Current Version**: v0.2.6 (Released January 30, 2025)

**Recent Updates**:
- **Built-in Components Implementation Progress** (January 31, 2025)
  - Phase 1: Core Registry Infrastructure - COMPLETED
  - Phase 2.0-2.3: Tool Migration and Enhancement - COMPLETED
    - Implemented comprehensive registry system with search and discovery
    - Migrated and enhanced all web tools (WebFetch, WebSearch, WebScrape, HTTPRequest)
    - Migrated and enhanced all file tools (ReadFile, WriteFile, FileList, FileDelete, FileMove, FileSearch)
    - Implemented all system tools (ExecuteCommand, GetEnvironmentVariable, GetSystemInfo, ProcessList)
    - Successfully deprecated and removed common_tools.go
    - All tools have comprehensive tests and documentation
  - Phase 2.4: Data Tools - COMPLETED (January 31, 2025)
    - Implemented JSONProcess with parsing, JSONPath querying, and transformations
    - Implemented CSVProcess with parsing, filtering, transformations, and JSON conversion
    - Implemented XMLProcess with parsing, XPath querying, and JSON conversion
    - Implemented DataTransform with filter, map, reduce, sort, group_by, unique, reverse
    - All data tools are pure data processing without LLM dependencies
    - Comprehensive test coverage and proper registry integration
  - Phase 2.5: Date, Time Tools - PENDING
  - Phase 2.6: Feed Process Tools - PENDING
  - Phase 3: Agent Templates - PENDING
  - Phase 4: Workflow Patterns - PENDING
- **Completed All Phases of Logging Strategy Implementation** (January 30, 2025)
  - Phase 1: Documentation - Created comprehensive logging documentation at `docs/technical/logging.md`
  - Phase 2: Standardized Examples - Converted all examples to use consistent logging patterns
  - Phase 3: Debug Infrastructure - Added debug build tags and conditional compilation support
  - Phase 4: Core Library Cleanup - Removed all direct logging from pkg/, improved error messages with context
- Added ABOUTME comments to all Go source files for better code documentation
- Created `CONTRIBUTING.md` with contribution guidelines including logging best practices
- Fixed all linting errors (removed empty branches in model_inventory.go)
- Added Logger interface to profiling package to support optional logging without forcing it on users
- All make targets tested and working (36.7% test coverage)

## Common Development Commands

### Build Commands
```bash
# Build the main binary
make build

# Build with debug logging enabled
make build-debug
# Then run with: GO_LLMS_DEBUG=all ./bin/go-llms-debug

# Build all example binaries
make build-examples

# Build a specific example
make build-example EXAMPLE=simple
```

### Test Commands
```bash
# Run all tests excluding integration, multi-provider, stress tests
make test

# Run tests with debug logging enabled
make test-debug
GO_LLMS_DEBUG=param_cache make test-debug

# Run all tests including integration tests (requires API keys)
make test-all

# Run tests for a specific package
make test-pkg PKG=schema/validation

# Run a specific test function
make test-func PKG=schema/validation FUNC=TestArrayValidation

# Run mock-only integration tests (doesn't require API keys)
make test-integration-mock

# Run only short tests
make test-short
```

### Integration Test Commands
```bash
# Run integration tests (requires API keys for providers)
make test-integration

# Enable specific provider tests with environment variables
ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 go test ./tests/integration/...

# Skip specific provider tests
SKIP_OPEN_ROUTER=1 ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 go test ./tests/integration/...
SKIP_OLLAMA=1 ENABLE_OPENAPI_COMPATIBLE_API_TESTS=1 go test ./tests/integration/...
```

### Benchmark Commands
```bash
# Run all benchmarks
make benchmark

# Run benchmarks for a specific package
make benchmark-pkg PKG=schema/validation

# Run a specific benchmark
make benchmark-specific BENCH=BenchmarkConsensus

# Profile CPU usage
make profile-cpu
```

### Code Quality Commands
```bash
# Format code
make fmt

# Run vet checks
make vet

# Run linting (requires golangci-lint)
make lint

# Install golangci-lint if not present
make install-lint

# Tidy dependencies
make deps-tidy

# Download dependencies
make deps-download

# Combined dependency management (tidy + download)
make deps
```

### Clean Commands
```bash
# Clean build artifacts
make clean

# Clean everything including Go cache
make clean-all
```

### Coverage Commands
```bash
# Generate test coverage report
make coverage

# Generate coverage for specific package
make coverage-pkg PKG=schema/validation

# Generate and view coverage report (opens in browser)
make coverage-view
```

### Helpful CLI Commands
```bash
# List available examples
ls cmd/examples/

# Run the CLI binary directly after building
./bin/go-llms

# Run a specific example binary
./bin/simple

# Run model info fetcher example to check available models
go run cmd/examples/modelinfo/main.go
```

## Core Architecture

Go-LLMs follows a vertical slicing approach where code is organized by feature:

1. **Schema Validation** (`pkg/schema/`):
   - Validates JSON data against predefined schemas
   - Supports type coercion and conditional validation
   - Key interfaces: `Validator`, `SchemaRepository`, `SchemaGenerator`
   - Custom validators and complex validation rules (if/then/else, allOf, anyOf, oneOf)

2. **LLM Integration** (`pkg/llm/`):
   - Provider implementations for OpenAI, Anthropic, Google Gemini
   - Multi-provider strategies (Fastest, Primary, Consensus)
   - Interface-based provider option system for configuration
   - Multimodal content support (text, images, files, videos, audio)
   - Key interfaces: `Provider` (main interface), `ModelRegistry`
   - Message format with support for multimodal content types

3. **Structured Output Processing** (`pkg/structured/`):
   - Extract structured data from LLM responses
   - Validate against schemas and map to Go structs
   - Schema-based prompt enhancement
   - Key interfaces: `Processor`, `PromptEnhancer`
   - JSON extraction with retries and validation

4. **Agent Workflows** (`pkg/agent/`):
   - Tool integration for function calling
   - Message management and context handling
   - Hooks for monitoring and logging
   - Key interfaces: `Agent`, `Tool`, `Hook`
   - Generic `RunContext[D]` for type-safe dependency injection

## Key Design Patterns

1. **Interface-Based Provider Option System**:
   The codebase uses a hierarchical interface system for type-safe provider configuration:
   - Base interface: `ProviderOption`
   - Provider-specific interfaces: `OpenAIOption`, `AnthropicOption`, `GeminiOption`
   - Common interface: `CommonOption` (implements all provider interfaces)
   
   Options are applied via interface methods rather than type switches:
   ```go
   provider := provider.NewOpenAIProvider(
       apiKey,
       modelName,
       domain.NewHTTPClientOption(httpClient),     // Common option
       domain.NewOpenAIOrganizationOption("org"),  // Provider-specific option
   )
   ```

2. **Multi-Provider Strategies**:
   The codebase supports multiple strategies for working with several LLM providers concurrently:
   - **Fastest Strategy**: Returns the first successful response
   - **Primary Strategy**: Tries primary provider first, falls back to others on failure
   - **Consensus Strategy**: Compares results from multiple providers for agreement
   - **Sequential Strategy**: Tries providers in order until one succeeds

3. **Memory Pooling**:
   Extensive use of `sync.Pool` for improved performance:
   - Schema validation objects are pooled and reused
   - String builders and buffers are pooled
   - Provider message objects are cached and pooled

4. **Message Format and Caching**:
   - Unified `Message` type supports multimodal content
   - Provider-specific message conversions are cached
   - Base64 encoding for binary content, URL references supported

## Testing Approach

1. **Unit Tests**: Test individual components in isolation with mocks
2. **Integration Tests**: Test interactions with actual LLM providers (require API keys)
3. **Stress Tests**: Test behavior under high load and concurrency
4. **Benchmark Tests**: Measure performance of key components

Integration tests with real providers are skipped by default unless the corresponding API key environment variables are set (e.g., OPENAI_API_KEY, ANTHROPIC_API_KEY, GEMINI_API_KEY).

## Go Best Practices to Follow

1. Always run `make fmt` and `make vet` before committing changes
2. Follow the existing error handling patterns (returning errors as the last return value)
3. Use sync.Pool for objects that are frequently created and disposed
4. Use context.Context for timeout and cancellation
5. Match the existing code style and patterns when adding new features
6. Add comprehensive tests for new functionality
7. Use benchmark tests to verify performance of optimizations
8. Implement both the mock and real versions for any new provider

## Logging Guidelines

The codebase follows specific logging patterns to maintain performance and give users control:

1. **Library Code (pkg/)**: No direct logging - return errors with context instead
   ```go
   // Good
   return fmt.Errorf("failed to parse response: %w", err)
   
   // Bad - don't log in library
   log.Printf("Error: %v", err)
   ```

2. **Examples**: Use `log` package for consistency
   - Agent examples can use `slog` to demonstrate LoggingHook
   - Don't mix `log` and `fmt` in the same example

3. **CLI Tools**: Use `fmt` for output control
   - `fmt.Printf/Println` for normal output
   - `fmt.Fprintf(os.Stderr, ...)` for errors

4. **Thread Safety**: All logging approaches are concurrent-safe
   - `slog` and `log` packages are thread-safe
   - No shared mutable logging state

5. **Debug Logging**: Use build tags (coming soon)
   - Will replace commented debug prints
   - Build with `-tags debug` for verbose logging

See [docs/technical/logging.md](docs/technical/logging.md) for the complete logging strategy.

## Debug Logging

The codebase includes a debug logging infrastructure that compiles to zero-overhead no-ops in production:

```bash
# Build with debug support
make build-debug

# Run tests with debug logging
make test-debug

# Enable debug logging for specific components
GO_LLMS_DEBUG=param_cache,schema make test-debug

# Enable all debug logging
GO_LLMS_DEBUG=all make test-debug
```

Debug logging is implemented using build tags, so there's no performance impact when not enabled. The debug infrastructure is in `pkg/internal/debug/`.

## Current Development Focus

Based on the TODO.md file, these are the current development priorities:

1. **Model Context Protocol Support**:
   - Add Model Context Protocol Client support for Agents
   - Add Model Context Protocol Server support for Workflows or Agents
   
2. **Architecture & Built-in Components** (Next Priority):
   - P1: Analyze structure for exposing built-in tools, agents, and workflows
   - P2: Build useful built-in tools (research and implementation)
   - P3: Build useful built-in agents with and without tools
   - P4: Build useful multi-agent workflows
   
3. **Performance Optimizations** (Marked for REVISIT):
   - Create benchmark harness for A/B testing optimizations
   - Implement visualization for memory allocation patterns
   - Create real-world test scenarios for end-to-end performance
   - Advanced optimizations including adaptive channel buffer sizing, pool prewarming, etc.
   - Performance validation with metrics and benchmarks
   
4. **Final Documentation and Release**:
   - Fix identified cross-link issues (path inconsistencies, broken links) - REVISIT
   - Perform final consistency check across all documentation - REVISIT
   - API refinement based on usage feedback
   - Final review and preparation for stable release
   
## Completed Development Items

1. **Comprehensive Logging Strategy (Completed in v0.2.6)**:
   - Created comprehensive logging documentation at docs/technical/logging.md
   - Standardized all examples to use consistent logging patterns
   - Added debug infrastructure with build tags and GO_LLMS_DEBUG support
   - Removed all direct logging from library code (pkg/)
   - Improved error messages with context and proper error wrapping
   - Added Logger interface to profiling package for optional logging
   - Verified thread safety in all logging paths
   - Added ABOUTME comments to all Go source files
   
2. **Dependency Reduction Journey (Completed in v0.2.4)**:
   - Successfully migrated from viper/cobra to koanf/kong, then to stdlib
   - Reduced binary size from 14MB to 6.3MB (55% total reduction)  
   - Documentation at docs/technical/dependency_reduction.md
   - Full backward compatibility maintained
   
2. **CLI Examples Enhancement (Completed in v0.2.1-v0.2.3)**:
   - Added comprehensive multimodal example application
   - Improved CLI argument parsing (v0.2.3)
   - Migrated through multiple CLI frameworks to find optimal solution
   
3. **Multimodal Support (Completed in v0.2.0)**:
   - Full implementation with text, images, files, videos, and audio
   - Complete example with CLI interface
   - Comprehensive documentation and tests
   
4. **Documentation Consolidation (Completed)**:
   - All documentation is consistent and properly linked
   - REFERENCE.md updated with all documentation
   - Navigation links verified
   
See TODO-DONE.md for full list of completed tasks

## CLI Migration Notes

The CLI migration has been completed with the following journey:
1. Phase 1: Migrated from viper/cobra to koanf/kong (increased binary size)
2. Phase 2: Analyzed the impact and identified stdlib approach
3. Phase 3: Removed koanf/kong, replaced with stdlib flag package and direct YAML parsing
4. Result: 36% binary size reduction (from 9.9MB to 6.3MB)
5. Config file format remains YAML for backward compatibility
6. Environment variable support is maintained with GO_LLMS_ prefix
7. Shell completion feature was removed in favor of smaller binary size

For the full journey, see docs/technical/dependency_reduction.md

## Environment Variables for Provider Configuration

The library supports automatic provider configuration through environment variables:

### OpenAI Provider
- `GO_LLMS_OPENAI_API_KEY`: API key
- `GO_LLMS_OPENAI_BASE_URL`: Custom base URL
- `GO_LLMS_OPENAI_ORGANIZATION`: Organization ID
- `GO_LLMS_OPENAI_MODEL`: Default model

### Anthropic Provider
- `GO_LLMS_ANTHROPIC_API_KEY`: API key
- `GO_LLMS_ANTHROPIC_BASE_URL`: Custom base URL
- `GO_LLMS_ANTHROPIC_MODEL`: Default model

### Gemini Provider
- `GO_LLMS_GEMINI_API_KEY`: API key
- `GO_LLMS_GEMINI_BASE_URL`: Custom base URL
- `GO_LLMS_GEMINI_MODEL`: Default model

### Using Environment Variables
```go
// Option factories automatically read from environment
options := llmutil.BuildProviderOptions(
    llmutil.GetEnvOptionFactory(),
    // Additional options...
)
```

## Model Information Feature

The library includes a model discovery feature that fetches available models from providers:

```bash
# Run the model info example to see available models
go run cmd/examples/modelinfo/main.go

# Or use the built binary
./bin/modelinfo
```

The model information service (`pkg/util/llmutil/modelinfo/`) provides:
- Automatic discovery of available models from providers
- Caching of model information to reduce API calls
- File-based cache for persistence across runs
- Model capability information (context length, features, etc.)

## Recent Release Status

### v0.2.6 (Current - January 30, 2025)
- Documentation updates and consistency improvements
- Package organization refinements
- Linting and formatting updates

### v0.2.4 (January 17, 2025)
- Complete dependency reduction journey
- Removed all heavy CLI dependencies (koanf, kong)
- 55% total binary size reduction since v0.1.0
- Maintained full backward compatibility

### v0.2.3 (January 16, 2025)
- Intermediate migration from viper/cobra to koanf/kong
- Improved shell completion (later removed for size optimization)

### v0.2.1 (January 15, 2025)
- Added comprehensive multimodal example CLI
- Enhanced documentation and examples

### v0.2.0 (January 14, 2025)
- Full multimodal content support
- Support for text, images, files, videos, and audio

See README.md for the complete changelog