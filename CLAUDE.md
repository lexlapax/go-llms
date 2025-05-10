# Claude.md - Go-LLMs Project Guide

This document serves as a guide for Claude Code when working on the Go-LLMs project.

> **Note:** See [TODO.md](TODO.md) for the current project task list and implementation status.

## Project Overview

Go-LLMs is a unified Go library for LLM integration that provides a simplified interface to interact with various LLM providers while offering robust data validation and agent tooling. It combines structured output processing (inspired by pydantic-ai), a consistent provider interface, and flexible agent workflows in a single, cohesive package.

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

1. **Project Setup and Architecture** (Phase 1) ✅
   - Initialize project structure
   - Define core interfaces

2. **Schema Validation Implementation** (Phase 2) ✅
   - Implement schema models
   - Create validator with TDD approach
   - Implement schema generation from Go structs
   - Implement type coercion

3. **LLM Provider Integration** (Phase 3) ✅
   - Implement base interface
   - Create OpenAI provider
   - Create Anthropic provider
   - Implement mock provider for testing

4. **Structured Output Implementation** (Phase 4) ✅
   - Create output processor
   - Implement prompt enhancement for schemas

5. **Agent and Tool Implementation** (Phase 5) ✅
   - Implement tool system
   - Create context for dependency injection
   - Build agent system
   - Add monitoring hooks

6. **Integration and Examples** (Phase 6) ✅
   - Create example applications
   - Write integration tests
   - Create comprehensive documentation

7. **Performance Optimization and Refinement** (Phase 7) ✅
   - Optimize performance
   - Refine API based on feedback
   - Final testing and beta release

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
7. ✅ OpenAI provider implementation (using gpt-4o model)
8. ✅ Anthropic provider implementation (using claude-3-5-sonnet-latest model)
9. ✅ Gemini provider implementation
10. ✅ Build system via Makefile
11. ✅ Structured output processor
12. ✅ Prompt enhancement for structured outputs
13. ✅ Example applications
14. ✅ Code quality improvements (fixed linting issues)
15. ✅ Documentation consolidation and restructuring
16. ✅ Benchmarks for consensus algorithms and performance optimization
17. ✅ Comprehensive test suite for error conditions
18. ✅ Stress tests for high-load scenarios
19. ✅ Performance optimizations (schema caching, object clearing, etc.)

## Examples

Multiple example applications are provided:

1. **Simple Example** - Demonstrates core features with mock providers
   - Basic generation
   - Structured generation with schemas
   - Streaming responses
   - Processing raw outputs
   - Prompt enhancement

2. **Anthropic Example** - Shows integration with Anthropic Claude (using claude-3-5-sonnet-latest model)
   - Text generation
   - Message-based conversation
   - Structured recipe generation
   - Response streaming
   - AnthropicSystemPromptOption for consistent system behavior

3. **OpenAI Example** - Shows integration with OpenAI models (using gpt-4o model)
   - Text generation
   - Message-based conversation
   - Structured recipe generation
   - Response streaming
   - OpenAIOrganizationOption for organization-specific configuration

4. **Agent Example** - Demonstrates agent with tools for complex tasks

5. **Multi-Provider Example** - Shows working with multiple providers with provider-specific options

6. **Consensus Example** - Demonstrates multi-provider consensus strategies

7. **OpenAI API Compatible Providers Example** - Shows integration with providers implementing the OpenAI API
   - OpenRouter integration with attribution headers
   - Ollama local model integration with dummy API key
   - Groq integration for high-performance inference
   - Best practices for different endpoint configurations

Run examples:
```bash
# Build and run the simple example
make example EXAMPLE=simple
./bin/simple

# Build and run the Anthropic example (requires API key)
export ANTHROPIC_API_KEY=your_api_key_here
make example EXAMPLE=anthropic
./bin/anthropic

# Build and run the OpenAI API Compatible Providers example
# For OpenRouter
export OPENROUTER_API_KEY=your_api_key_here
# For Ollama
export OLLAMA_HOST=http://localhost:11434
# For Groq
export GROQ_API_KEY=your_api_key_here
make example EXAMPLE=openai_api_compatible_providers
./bin/openai_api_compatible_providers
```

## Beta Release Completed

The beta release of Go-LLMs has been successfully completed with all planned features and optimizations for this stage. Key accomplishments include:

### Features Completed
- [x] Interface-based provider option system
- [x] Enhanced provider options with environment variable support
- [x] Option factory functions for common provider configurations
- [x] Performance optimizations for high-throughput scenarios

### Providers Supported
- [x] OpenAI (with gpt-4o and other models)
- [x] Anthropic (with claude-3-5-sonnet and other models)
- [x] Google Gemini
- [x] OpenAI API Compatible providers
  - [x] OpenRouter
  - [x] Ollama
  - [x] Groq

### Performance Optimizations
- [x] Schema caching with LRU eviction and TTL expiration
- [x] Object clearing optimizations for large response objects
- [x] String builder capacity estimation for complex schemas
- [x] Agent workflow optimizations (message creation, tool extraction, JSON parsing)
- [x] LLM provider message handling optimizations (caching, fast paths, reduced allocations)

### Documentation & Examples
- [x] Comprehensive API documentation
- [x] Provider-specific and use-case options documentation
- [x] Example applications demonstrating all key features
- [x] Enhanced Gemini provider documentation
- [x] Updated OpenAI API Compatible providers documentation
- [x] Technical documentation for performance optimizations

## Future Roadmap (Post-Beta)

The following features are planned for future development:

### Features
- [ ] Model Context Protocol Client support for Agents
- [ ] Model Context Protocol Server support for Workflows or Agents

### Performance Optimizations
This work is organized into phases:

#### Phase 1: Baseline Profiling Infrastructure
- [x] Add CPU and memory profiling hooks to key operations
- [x] Add monitoring for cache hit rates and pool statistics
- [ ] Create benchmark harness for A/B testing optimizations
- [ ] Implement visualization for memory allocation patterns
- [ ] Create real-world test scenarios for end-to-end performance

#### Phase 2: High-Impact Optimizations (Quick Wins)
- [x] Optimize schema JSON marshaling with faster alternatives
- [x] Improve schema caching with better key generation
- [x] Optimize object clearing operations for large response objects
- [x] Add expiration policy to schema cache to prevent unbounded growth
- [x] Optimize string builder capacity estimation for complex schemas

#### Phase 3: Advanced Optimizations
- [ ] Implement adaptive channel buffer sizing based on usage patterns
- [ ] Add pool prewarming for high-throughput scenarios
- [ ] Reduce redundant property iterations in schema processing
- [ ] Implement more granular locking in cached objects
- [ ] Optimize zero-initialization patterns for pooled objects
- [ ] Introduce buffer pooling for string builders

#### Phase 4: Integration and Validation
- [ ] Document performance improvements with metrics
- [ ] Verify optimizations in high-concurrency scenarios
- [ ] Create benchmark comparison charts for before/after
- [ ] Implement regression testing to prevent performance degradation
- [ ] Add performance acceptance criteria to CI pipeline

### Documentation and Testing
- [ ] Fix identified cross-link issues (path inconsistencies, broken links)
- [ ] Perform final consistency check across all documentation
- [ ] API refinement based on usage feedback
- [ ] Final review and preparation for stable release