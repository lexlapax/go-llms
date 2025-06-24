# API Reference

This section provides comprehensive API documentation for all go-llms packages, organized by functionality.

## Core Packages

### LLM Integration
- **[LLM API](llm.md)** - Language model provider integration
  - Unified interface for OpenAI, Anthropic, Google Gemini, Ollama, OpenRouter
  - Multi-provider strategies for reliability
  - Streaming and structured generation support

### Data Validation
- **[Schema API](schema.md)** - JSON Schema validation
  - Schema definition and validation
  - Type coercion and custom validators
  - Integration with structured outputs

### Structured Output
- **[Structured API](structured.md)** - Extract structured data from LLMs
  - Prompt enhancement with schemas
  - JSON extraction and validation
  - Type-safe output processing

## Agent Framework

### Core Agent System
- **[Agent API](agent.md)** - Build autonomous agents
  - Agent lifecycle and state management
  - Hook system for monitoring
  - Event-driven architecture

### Tools and Extensions
- **[Tools API](tools.md)** - Create and manage agent tools
  - ToolBuilder pattern for rich metadata
  - Agent-tool bidirectional conversion
  - Performance optimizations

- **[Built-in Tools](builtins.md)** - Pre-built tool library
  - 30+ tools across 7 categories
  - MCP compatibility
  - Tool discovery and registry

### Workflows
- **[Workflow API](workflows.md)** - Compose complex agent behaviors
  - Sequential, parallel, conditional, and loop patterns
  - Error handling and recovery
  - State management across steps

## Utilities

### Testing Support
- **[Test Utilities](testutils.md)** - Testing helpers and mocks
  - Provider mocks for unit testing
  - Fixture management
  - Assertion helpers

### General Utilities
- **[Utilities API](utils.md)** - Common utilities
  - Provider configuration parsing
  - Model information management
  - Error handling utilities

## Quick Links

- [Getting Started Guide](/docs/user-guide/getting-started.md)
- [Examples Directory](/cmd/examples/)
- [Technical Documentation](/docs/technical/)
- [Contributing Guide](/CONTRIBUTING.md)

## API Stability

The APIs documented here follow semantic versioning:
- **Stable APIs** (v0.3.x): May have minor changes but no breaking changes
- **Experimental APIs**: Marked with warnings, may change significantly
- **Deprecated APIs**: Marked with deprecation notices and migration guides

## Documentation Format

Each API document includes:
- **Overview**: High-level description and use cases
- **Core Types**: Main interfaces and structs
- **Functions**: Public functions and methods
- **Examples**: Practical usage examples
- **Best Practices**: Recommended patterns and anti-patterns
- **Error Handling**: Common errors and recovery strategies
