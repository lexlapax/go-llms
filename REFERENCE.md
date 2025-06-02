# Go-LLMs Documentation Reference

Welcome to the Go-LLMs documentation reference. This document provides links to all documentation resources for the Go-LLMs project.

## User Guides

- [Getting Started](/docs/user-guide/getting-started.md) - Introduction and basic usage examples
- [Built-in Components](/docs/user-guide/built-in-components.md) - Using built-in tools, agents, and workflows
- [Provider Options](/docs/user-guide/provider-options.md) - Using the provider option system for configuration
- [Multi-Provider Guide](/docs/user-guide/multi-provider.md) - Working with multiple LLM providers
- [Multimodal Content](/docs/user-guide/multimodal-content.md) - Working with text, images, files, videos, and audio
- [Model Discovery](/docs/user-guide/model-discovery.md) - Model inventory and capability discovery
- [Advanced Validation](/docs/user-guide/advanced-validation.md) - Advanced schema validation features
- [Error Handling](/docs/user-guide/error-handling.md) - Error handling patterns and best practices

## API Reference

- [API Overview](/docs/api/README.md) - Overview of the API
- [Schema API](/docs/api/schema.md) - Schema definition and validation
- [LLM API](/docs/api/llm.md) - LLM provider integration
- [Structured API](/docs/api/structured.md) - Structured output processing
- [Agent API](/docs/api/agent.md) - Agent and tool functionality

## Technical Documentation

- [Architecture](/docs/technical/architecture.md) - Overview of the library architecture
- [Performance Optimization](/docs/technical/performance.md) - Performance optimization strategies
- [Multimodal Content Implementation](/docs/technical/multimodal-content.md) - Implementation details for multimodal support
- [Testing Framework](/docs/technical/testing.md) - Error condition testing and stress testing
- [Benchmarking Framework](/docs/technical/benchmarks.md) - Performance benchmarks and measurement
- [Sync.Pool Implementation](/docs/technical/sync-pool.md) - Detailed guide on sync.Pool usage
- [Caching Mechanisms](/docs/technical/caching.md) - Caching strategies and implementations
- [Concurrency Patterns](/docs/technical/concurrency.md) - Thread safety and concurrent execution
- [Adding a New Provider](/docs/technical/new-provider.md) - Step-by-step guide to implementing and integrating a new LLM provider
- [Dependency Reduction Journey](/docs/technical/dependency_reduction.md) - Chronicle of dependency reduction from viper/cobra to stdlib
- [Logging](/docs/technical/logging.md) - Logging patterns and best practices

## Examples

**[Examples Overview](/cmd/examples/README.md)** - Complete guide to all example applications

### Basic Examples
- [Simple Example](/cmd/examples/simple/README.md) - Basic usage with mock providers
- [Anthropic Example](/cmd/examples/anthropic/README.md) - Integration with Anthropic Claude
- [OpenAI Example](/cmd/examples/openai/README.md) - Integration with OpenAI models
- [Gemini Example](/cmd/examples/gemini/README.md) - Integration with Google Gemini
- [OpenAI API Compatible Providers](/cmd/examples/openai_api_compatible_providers/README.md) - Using OpenRouter and Ollama

### Built-in Tools Examples
- [Built-in Components Examples Guide](/cmd/examples/BUILTINS_EXAMPLES.md) - Complete guide to using built-in components
- [Built-in Tools Discovery](/cmd/examples/builtins-discovery/README.md) - Discover and use built-in tools
- [Built-in File Tools](/cmd/examples/builtins-file-tools/README.md) - Enhanced file operations with built-in tools
- [Built-in Web Tools](/cmd/examples/builtins-web-tools/README.md) - Web operations (fetch, search, scrape, HTTP requests)
- [Built-in System Tools](/cmd/examples/builtins-system-tools/README.md) - System operations (execute commands, environment variables, process list)
- [Built-in Data Tools](/cmd/examples/builtins-data-tools/README.md) - Data processing (JSON, CSV, XML, transformations)
- [Built-in DateTime Tools](/cmd/examples/builtins-datetime-tools/README.md) - Date and time operations (parse, format, calculate, compare)
- [Built-in Feed Tools](/cmd/examples/builtins-feed-tools/README.md) - RSS, Atom, and JSON Feed processing (fetch, filter, aggregate, convert)

### Advanced Examples
- [Agent Example](/cmd/examples/agent/README.md) - Agent with tools
- [Multi-Provider Example](/cmd/examples/multi/README.md) - Working with multiple providers
- [Consensus Example](/cmd/examples/consensus/README.md) - Multi-provider consensus strategies
- [Provider Options Example](/cmd/examples/provider_options/README.md) - Demonstration of provider options system
- [Schema Example](/cmd/examples/schema/README.md) - Schema generation from Go structs
- [Coercion Example](/cmd/examples/coercion/README.md) - Type coercion for validation
- [Convenience Example](/cmd/examples/convenience/README.md) - Utility functions for common tasks
- [Model Info Example](/cmd/examples/modelinfo/README.md) - Model discovery and capability information
- [Metrics Example](/cmd/examples/metrics/README.md) - Performance monitoring and metrics
- [Multimodal Example](/cmd/examples/multimodal/README.md) - Working with images, audio, and video content

## CLI Documentation

- [Command Line Interface](/cmd/README.md) - Documentation for the Go-LLMs CLI

## Project Planning

- [Design Inspirations](/docs/plan/design-inspirations.md) - Key inspirations and design decisions
- [Coding Practices](/docs/plan/coding-practices.md) - Coding standards and guidelines
- [Implementation Plan](/docs/plan/implementation-plan.md) - Detailed implementation plan
- [Project Planning Overview](/docs/plan/README.md) - Overview of planning documents
- [Built-in Components Implementation Plan](/BUILTIN_COMPONENTS_IMPLEMENTATION_PLAN.md) - Detailed plan for built-in components implementation
- [Feed Tools Plan](/FEED_TOOLS_PLAN.md) - Design and implementation plan for feed processing tools

## Contributing

- [Contributing Guidelines](/CONTRIBUTING.md) - How to contribute to the project, including coding standards and logging guidelines