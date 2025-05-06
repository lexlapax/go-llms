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

For development, use these commands:

```
go mod init github.com/yourusername/go-llms   # Initialize module (already done)
go mod tidy                                   # Clean up dependencies
go test ./...                                 # Run all tests
go test -v ./pkg/schema/...                   # Run tests for schema package
go test -bench=. ./...                        # Run benchmarks
go run examples/simple/main.go                # Run example
```

## Current Implementation Status

The project is in the initial setup phase. Follow the implementation steps in order:

1. First, create the project structure
2. Implement core domain interfaces
3. Implement schema validation
4. Add LLM providers
5. Build agent and tool systems

## Next Steps

- Initialize the project structure
- Define core interfaces for schema validation, LLM providers, and agents
- Start implementing the schema validation system using TDD