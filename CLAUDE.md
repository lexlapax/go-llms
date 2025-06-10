# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go-LLMs is a Go library that provides a unified interface to interact with various LLM providers (OpenAI, Anthropic, Google Gemini, etc.) with robust data validation and agent tooling. Key features include structured output processing, a consistent provider interface, agent workflows, and multi-provider strategies.

Until we reach close to v1.. *no backward compatibility* do not add extra code for backward compatibility. when planning plan for in place replacement and migration of code to new functionality.

**Current Version**: v0.3.1 (Active Development - January 2025)

## Common Development Commands

### Build Commands
```bash
make build                      # Build the main binary
make build-debug               # Build with debug logging enabled
make build-examples            # Build all example binaries
make build-example EXAMPLE=simple  # Build a specific example
```

### Test Commands
```bash
make test                      # Run all tests (excludes integration/stress)
make test-all                  # Run all tests including integration
make test-pkg PKG=schema/validation  # Test specific package
make test-func PKG=schema/validation FUNC=TestArrayValidation  # Test specific function
make test-integration-mock     # Run mock-only integration tests
```

### Code Quality Commands
```bash
make fmt                       # Format code
make vet                       # Run vet checks
make lint                      # Run linting (requires golangci-lint)
make coverage                  # Generate test coverage report
make deps                      # Tidy and download dependencies
```

## Core Architecture

Go-LLMs follows a vertical slicing approach where code is organized by feature:

1. **Schema Validation** (`pkg/schema/`): Validates JSON data against schemas with type coercion
2. **LLM Integration** (`pkg/llm/`): Provider implementations with multi-provider strategies
3. **Structured Output Processing** (`pkg/structured/`): Extract and validate structured data from LLM responses
4. **Agent Workflows** (`pkg/agent/`): Tool integration, state management, hooks, and workflows

## Key Design Patterns

1. **Interface-Based Provider Options**: Type-safe configuration via interface methods
2. **Multi-Provider Strategies**: Fastest, Primary, Consensus, Sequential
3. **Memory Pooling**: Extensive use of `sync.Pool` for performance
4. **State-Based Agents**: Agents operate on state rather than messages

## Testing Approach

1. **Unit Tests**: Test components in isolation with mocks
2. **Integration Tests**: Test with actual providers (require API keys)
3. **Stress Tests**: Test under high load and concurrency
4. **Benchmark Tests**: Measure performance

Integration tests are skipped unless API keys are set (e.g., OPENAI_API_KEY).

## Go Best Practices

1. Always run `make fmt` and `make vet` before committing
2. Follow existing error handling patterns (errors as last return value)
3. Use sync.Pool for frequently created/disposed objects
4. Use context.Context for timeout and cancellation
5. Match existing code style and patterns
6. Add comprehensive tests for new functionality
7. Use benchmark tests to verify optimizations

## Logging Guidelines

1. **Library Code (pkg/)**: No direct logging - return errors with context
2. **Examples**: Use `log` package (agent examples can use `slog` for LoggingHook)
3. **CLI Tools**: Use `fmt` for output control
4. **Debug Logging**: Use build tags with `make build-debug` and `GO_LLMS_DEBUG=all`

## Current Development Focus

1. **Tool System Enhancement Phase 2**: Migrating all built-in tools to use ToolBuilder pattern with enhanced metadata (IN PROGRESS)
2. **Model Context Protocol Support**: Add MCP Client/Server support for Agents
3. **Phase 6: Advanced Features**: State persistence, agent discovery, advanced merge strategies
4. **Built-in Agents**: Text, Research, Coding, Data, Feed agents (POSTPONED)
5. **Performance Optimizations**: Benchmark harness, memory patterns (REVISIT)
6. **Final Documentation**: Cross-link fixes, API refinement

## Environment Variables

### Provider Configuration
- `GO_LLMS_OPENAI_API_KEY`, `GO_LLMS_OPENAI_MODEL`, etc.
- `GO_LLMS_ANTHROPIC_API_KEY`, `GO_LLMS_ANTHROPIC_MODEL`, etc.
- `GO_LLMS_GEMINI_API_KEY`, `GO_LLMS_GEMINI_MODEL`, etc.

### Debug Logging
- `GO_LLMS_DEBUG=all` - Enable all debug logging
- `GO_LLMS_DEBUG=param_cache,schema` - Enable specific components

## Recent Major Completions

- **Agent Architecture Restructuring** (Phases 1-5, 7): Complete rewrite with state-based agents, workflow agents, tool integration
- **Built-in Components** (Phases 1-2.6): Registry system, all web/file/system/data/datetime/feed tools
- **Comprehensive Logging Strategy**: Debug infrastructure, no library logging, ABOUTME comments
- **Dependency Reduction**: 55% binary size reduction (14MB → 6.3MB)

See TODO.md for detailed task tracking and TODO-DONE.md for completed items.