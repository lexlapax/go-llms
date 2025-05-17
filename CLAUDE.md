# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go-LLMs is a Go library that provides a unified interface to interact with various LLM providers (OpenAI, Anthropic, Google Gemini, etc.) with robust data validation and agent tooling. Key features include structured output processing, a consistent provider interface, agent workflows, and multi-provider strategies.

**Current Version**: v0.2.4 (Released January 17, 2025)

## Common Development Commands

### Build Commands
```bash
# Build the main binary
make build

# Build all example binaries
make build-examples

# Build a specific example
make build-example EXAMPLE=simple
```

### Test Commands
```bash
# Run all tests excluding integration, multi-provider, stress tests
make test

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

# Run linting
make lint

# Tidy dependencies
make deps-tidy
```

## Core Architecture

Go-LLMs follows a vertical slicing approach where code is organized by feature:

1. **Schema Validation** (`pkg/schema/`):
   - Validates JSON data against predefined schemas
   - Supports type coercion and conditional validation

2. **LLM Integration** (`pkg/llm/`):
   - Provider implementations for OpenAI, Anthropic, Google Gemini
   - Multi-provider strategies (Fastest, Primary, Consensus)
   - Interface-based provider option system for configuration
   - Multimodal content support (text, images, files, videos, audio)

3. **Structured Output Processing** (`pkg/structured/`):
   - Extract structured data from LLM responses
   - Validate against schemas and map to Go structs
   - Schema-based prompt enhancement

4. **Agent Workflows** (`pkg/agent/`):
   - Tool integration for function calling
   - Message management and context handling
   - Hooks for monitoring and logging

## Key Design Patterns

1. **Provider Option System**:
   The codebase uses an interface-based option system that allows for both common and provider-specific configuration. Options are applied during provider creation.

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
   - Fastest Strategy: Uses the first provider to respond
   - Primary Strategy: Tries primary provider first, with fallbacks
   - Consensus Strategy: Compares results from multiple providers

3. **Memory Pooling**:
   The codebase extensively uses sync.Pool for improved performance and reduced GC pressure, particularly in the schema validation and structured output processing.

4. **Message Caching**:
   Caching mechanisms are used to avoid redundant conversions and processing, especially for provider message format conversions.

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

## Current Development Focus

Based on the TODO.md file, these are the current development priorities:

1. **Model Context Protocol Support**:
   - Add Model Context Protocol Client support for Agents
   - Add Model Context Protocol Server support for Workflows or Agents
   
2. **Performance Optimizations**:
   - Create benchmark harness for A/B testing optimizations
   - Implement visualization for memory allocation patterns
   - Create real-world test scenarios for end-to-end performance
   - Advanced optimizations including adaptive channel buffer sizing, pool prewarming, etc.
   - Performance validation with metrics and benchmarks
   
3. **Final Documentation and Release**:
   - Fix identified cross-link issues (path inconsistencies, broken links)
   - Perform final consistency check across all documentation
   - API refinement based on usage feedback
   - Final review and preparation for stable release
   
## Completed Development Items

1. **Dependency Reduction Journey (Completed in v0.2.4)**:
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

## Recent Release Status

### v0.2.4 (Current - January 17, 2025)
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