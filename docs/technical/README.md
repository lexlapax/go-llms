# Technical Documentation

> **[Documentation Home](/docs/README.md) / Technical Documentation**

This directory contains technical documentation for the Go-LLMs library, intended for contributors and advanced users who need to understand the internal architecture, implementation patterns, and optimization strategies.

*Related: [User Guide](/docs/user-guide/README.md) | [API Reference](/docs/api/README.md)*

## Core Architecture

- [Architecture Overview](architecture.md) - System design, component structure, and data flow patterns
- [Agent Architecture](agents.md) - Agent system implementation details and patterns
- [Tool Development](tool-development.md) - Internal tool architecture and development patterns
- [Tool Discovery API](tool-discovery-api.md) - Metadata-first tool discovery system for scripting engines
- [Built-in Components](built-in-components.md) - Registry system and component patterns

## Implementation Guides

- [Provider Implementation](provider-implementation.md) - How to add new LLM providers
- [Authentication System](authentication.md) - Authentication architecture and patterns
- [Multimodal Content](multimodal-content.md) - Technical implementation of multimodal support

## Performance & Optimization

- [Performance Optimization](performance.md) - Optimization strategies and benchmarks
- [Caching Mechanisms](caching.md) - Cache implementations and strategies
- [Concurrency Patterns](concurrency.md) - Thread safety and concurrent execution
- [Sync.Pool Implementation](sync-pool.md) - Memory optimization with object pooling

## Testing & Quality

- [Testing Framework](testing.md) - Testing strategies and patterns
- [Benchmarking Framework](benchmarks.md) - Performance measurement approach

## Development Practices

- [Logging](logging.md) - Logging patterns and best practices
- [Dependency Reduction](dependency_reduction.md) - Journey from heavy dependencies to stdlib
- [Tools](tools.md) - Tool system architecture and patterns

## Navigation

### For Contributors
1. Start with [Architecture Overview](architecture.md)
2. Review relevant implementation guides
3. Understand [Testing Framework](testing.md)
4. Follow [Development Practices](#development-practices)

### For Provider Implementers
1. Read [Provider Implementation](provider-implementation.md)
2. Understand [Authentication System](authentication.md)
3. Review [Testing Framework](testing.md)

### For Tool Developers
1. Study [Tool Development](tool-development.md)
2. Learn [Tool Discovery API](tool-discovery-api.md) for scripting integration
3. Review [Built-in Components](built-in-components.md)
4. Understand [Agent Architecture](agents.md)

### For Performance Optimization
1. Read [Performance Optimization](performance.md)
2. Understand [Caching Mechanisms](caching.md)
3. Review [Concurrency Patterns](concurrency.md)

## Contributing

When contributing technical documentation:
1. Focus on implementation details, not usage
2. Include code examples from actual implementation
3. Cross-reference related documentation
4. Keep content up-to-date with code changes