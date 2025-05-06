# Claude.md - Go-LLMs Project Guide

This document serves as a guide for Claude Code when working on the Go-LLMs project (a port of pydantic-ai to Go).

## Project Overview

Go-LLMs is a Go library for creating LLM-powered applications with structured outputs and type safety. It aims to port the core functionality of pydantic-ai to Go while embracing Go's idioms and strengths.

## Project Structure

The project follows a vertical slicing approach where code is organized by feature:

```
go-llms/
├── cmd/                       # Application entry points
│   └── examples/              # Example applications
├── internal/                  # Internal packages
├── pkg/                       # Public packages
│   ├── schema/                # Schema definition and validation feature
│   │   ├── domain/            # Core domain models
│   │   ├── validation/        # Validation logic
│   │   └── adapter/           # External adapters (JSON, OpenAPI)
│   ├── llm/                   # LLM integration feature
│   │   ├── domain/            # Core LLM domain models
│   │   ├── provider/          # LLM provider implementations
│   │   └── prompt/            # Prompt templating
│   ├── structured/            # Structured output feature
│   │   ├── domain/            # Structured output domain
│   │   ├── processor/         # Output processors
│   │   └── adapter/           # Integration adapters
│   └── agent/                 # Agent feature (tools, workflows)
│       ├── domain/            # Agent domain models
│       ├── tools/             # Tool implementations
│       └── workflow/          # Agent workflows
└── examples/                  # Usage examples
```

## Implementation Plan Summary

1. **Project Setup and Architecture** (Phase 1)
   - Initialize project structure
   - Define core interfaces

2. **Schema Validation Implementation** (Phase 2)
   - Implement schema models
   - Create validator with TDD approach
   - Implement schema generation from Go structs
   - Implement type coercion

3. **LLM Provider Integration** (Phase 3)
   - Implement base interface
   - Create OpenAI provider
   - Create Anthropic provider
   - Implement mock provider for testing

4. **Structured Output Implementation** (Phase 4)
   - Create output processor
   - Implement prompt enhancement for schemas

5. **Agent and Tool Implementation** (Phase 5)
   - Implement tool system
   - Create context for dependency injection
   - Build agent system
   - Add monitoring hooks

6. **Integration and Examples** (Phase 6)
   - Create example applications
   - Write integration tests
   - Create comprehensive documentation

7. **Performance Optimization and Refinement** (Phase 7)
   - Optimize performance
   - Refine API based on feedback
   - Final testing and release

## Go Coding Standards

When working on this project, follow these Go coding standards:

1. Use idiomatic Go patterns (interfaces, composition over inheritance)
2. Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
3. Use interface-based design at domain boundaries
4. Apply clean architecture principles
5. Implement error handling using the standard Go error pattern
6. Ensure proper documentation for all exported types and functions
7. Write tests for all components (test-driven development)
8. Keep packages focused on a single responsibility

## Available Commands

A Makefile has been created for this project. Use these commands:

```
make help              # Show all available commands
make all               # Build and test everything
make build             # Build the main binary
make test              # Run all tests with race detection and coverage
make test-pkg          # Run tests for a specific package (e.g., make test-pkg PKG=schema/validation)
make test-verbose      # Run all tests with verbose output
make test-func         # Run a specific test function (e.g., make test-func PKG=schema/validation FUNC=TestArrayValidation)
make benchmark         # Run benchmarks
make benchmark-pkg     # Run benchmarks for a specific package
make coverage          # Generate test coverage report
make coverage-pkg      # Generate test coverage for a specific package
make lint              # Run linters
make fmt               # Format Go code
make vet               # Run Go vet
make example           # Build a specific example (e.g., make example EXAMPLE=simple)
make clean             # Clean build artifacts
```

## Current Implementation Status

The project has made significant progress. Here's what has been completed:

1. ✅ Project structure and setup
2. ✅ Core domain interfaces for schema validation, LLM providers, and agents
3. ✅ Schema validation implementation (with TDD)
4. ✅ Schema generation from Go structs 
5. ✅ Type coercion system
6. ✅ Mock provider for testing
7. ✅ OpenAI provider implementation
8. ✅ Anthropic provider implementation
9. ✅ Build system via Makefile
10. ✅ Structured output processor
11. ✅ Prompt enhancement for structured outputs
12. ✅ Example applications
13. ✅ Code quality improvements (fixed linting issues)

## Examples

Two example applications are provided:

1. **Simple Example** - Demonstrates core features with mock providers
   - Basic generation
   - Structured generation with schemas
   - Streaming responses
   - Processing raw outputs
   - Prompt enhancement

2. **Anthropic Example** - Shows integration with Anthropic Claude
   - Text generation
   - Message-based conversation
   - Structured recipe generation
   - Response streaming

Run examples:
```bash
# Build and run the simple example
make example EXAMPLE=simple
./bin/simple

# Build and run the Anthropic example (requires API key)
export ANTHROPIC_API_KEY=your_api_key_here
make example EXAMPLE=anthropic
./bin/anthropic
```

## Next Steps

- Implement tool system
- Create context for dependency injection
- Build agent system
- Add integration tests