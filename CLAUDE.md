# Claude.md - Go-LLMs Project Guide

This document serves as a guide for Claude Code when working on the Go-LLMs project.

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

7. **Performance Optimization and Refinement** (Phase 7) 🔄
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
7. ✅ OpenAI provider implementation (using gpt-4o model)
8. ✅ Anthropic provider implementation (using claude-3-5-sonnet-latest model)
9. ✅ Build system via Makefile
10. ✅ Structured output processor
11. ✅ Prompt enhancement for structured outputs
12. ✅ Example applications
13. ✅ Code quality improvements (fixed linting issues)
14. ✅ Documentation consolidation and restructuring
15. ✅ Benchmarks for consensus algorithms
16. ✅ Example demonstrating schema generation from Go structs

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
make example EXAMPLE=openai_api_compatible_providers
./bin/openai_api_compatible_providers
```

## Current Next Steps

The project has successfully implemented all major components from the implementation plan. The remaining work focuses on the following areas:

### Features
- [ ] Add Model Context Protocol Client support for Agents
- [ ] Add Model Context Protocol Server support for Workflows or Agents
- [x] Implement interface-based provider option system (Phases 1-3 complete, Phase 4 in progress)

### Additional Providers
- [x] Add Google Gemini API provider
- [ ] Fix Gemini streaming output issues
- [x] Add Ollama support via OpenAI-compatible API
  - [x] Implemented in openai_api_compatible_providers example
  - [ ] Create dedicated integration test for Ollama

### Phase 4: Examples and Documentation for Provider Options ✅
- [x] Create provider_options example to demonstrate the new option system
- [x] Update openai_api_compatible_providers example (formerly custom_providers)
- [x] Update all other examples for consistency with the new option system
  - [x] Update anthropic example with AnthropicSystemPromptOption
  - [x] Update openai example with OpenAIOrganizationOption
  - [x] Update gemini example with GeminiGenerationConfigOption
  - [x] Update multi example with provider-specific options
- [x] Update documentation to explain the new provider option system
  - [x] Create detailed guide on using the provider option system
  - [x] Document all common and provider-specific options
  - [x] Add examples of combining options across providers

### Phase 5: Provider Options Enhancements 🔄
- [ ] Add support for passing provider options directly in ModelConfig
- [ ] Implement environment variable support for provider-specific options
- [ ] Create option factory functions for common provider configurations

### Phase 7: Testing & Performance Optimization
- [x] Add benchmarks for consensus algorithms
- [ ] Create comprehensive test suite for error conditions
- [ ] Add benchmarks for remaining components
- [ ] Implement stress tests for high-load scenarios
- [ ] Performance profiling and optimization:
  - [ ] Prompt processing and template expansion
  - [ ] Memory pooling for response types
- [ ] API refinement based on usage feedback
- [ ] Additional test coverage for edge cases
- [ ] Final review and preparation for stable release

### Optimizations Already Completed
- ✅ Agent workflow optimization (message creation, tool extraction, JSON parsing)
- ✅ LLM provider message handling optimization (caching, fast paths, reduced allocations)
- ✅ Documentation consolidation and consistency
- ✅ Interface-based provider option system (core interfaces and implementations)
- ✅ Refactored OpenAI, Anthropic, and Gemini providers to use the new option system

## Interface-Based Provider Option System - Implementation Status

The provider option system has been successfully implemented through Phases 1-3:

### ✅ Phase 1: Design and Core Interfaces
- Defined core ProviderOption interface
- Created provider-specific option interfaces (OpenAIOption, AnthropicOption, GeminiOption)
- Implemented common options (HTTPClient, BaseURL, Timeout, etc.)
- Added tests for core option interfaces

### ✅ Phase 2: Provider Refactoring
- Updated OpenAI provider to use the new option system
- Updated Anthropic provider to use the new option system
- Updated Gemini provider to use the new option system
- Created provider-specific options for unique features
- Updated all tests to verify the new option system works correctly

### ✅ Phase 3: Utility Function Updates
- Updated CreateProvider to handle the new options
- Updated ProviderFromEnv to support provider-specific options via environment variables
- Added tests for the updated utility functions

### ✅ Phase 4: Examples and Documentation
- ✅ Create provider_options example demonstrating the new option system
- ✅ Update openai_api_compatible_providers example (formerly custom_providers)
- ✅ Update all other examples for consistency
  - ✅ Add AnthropicSystemPromptOption to anthropic example
  - ✅ Add OpenAIOrganizationOption to openai example
  - ✅ Add GeminiGenerationConfigOption to gemini example
- ✅ Update documentation to explain the new system
  - ✅ Create detailed guide on using the provider option system
  - ✅ Update REFERENCE.md with all new documentation
  - ✅ Update DOCUMENTATION_CONSOLIDATION.md with recent changes

### Option System Design

```go
// Base interface for all provider options
type ProviderOption interface {
    // Identifies which provider type this option is for
    ProviderType() string
}

// Provider-specific option interfaces
type OpenAIOption interface {
    ProviderOption
    ApplyToOpenAI(*OpenAIProvider)
}

type AnthropicOption interface {
    ProviderOption
    ApplyToAnthropic(*AnthropicProvider)
}

type GeminiOption interface {
    ProviderOption
    ApplyToGemini(*GeminiProvider)
}

type MockOption interface {
    ProviderOption
    ApplyToMock(*MockProvider)
}
```

This approach follows Go's idiomatic interface design and enables type-safe, extensible option handling for all providers while supporting provider-specific options. The system has been expanded to include support for Mock providers and OpenAI API compatible providers like Ollama.