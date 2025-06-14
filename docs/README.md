# Go-LLMs Documentation Reference

Welcome to the comprehensive Go-LLMs documentation reference. This directory contains all documentation resources organized to serve different audiences and use cases.

## 📚 Documentation Structure

### [User Guide](user-guide/) 👥
**For developers using Go-LLMs**
- [Getting Started](user-guide/getting-started.md) - Quick start and basic concepts
- [Core Concepts](user-guide/core-concepts.md) - Understanding the library architecture
- [Providers](user-guide/providers.md) - Working with LLM providers
- [Agents](user-guide/agents.md) - Agent system and patterns
- [Tools & Components](user-guide/tools.md) - Built-in tools and components
- [Workflows](user-guide/workflows.md) - Workflow patterns and automation
- [Structured Output](user-guide/structured-output.md) - Schema validation and type coercion
- [Multimodal Content](user-guide/multimodal-content.md) - Working with text, images, files, videos, and audio
- [Error Handling](user-guide/error-handling.md) - Error handling patterns and best practices
- [Examples Gallery](user-guide/examples-gallery.md) - Usage examples and patterns

### [API Reference](api/) 🔧
**Complete API documentation**
- [LLM API](api/llm.md) - LLM provider integration
- [Agent API](api/agent.md) - Agent and workflow functionality
- [Schema API](api/schema.md) - Schema definition and validation
- [Structured API](api/structured.md) - Structured output processing
- [Built-ins API](api/builtins.md) - Built-in tools and components
- [Tools API](api/tools.md) - Tool development interfaces
- [Workflows API](api/workflows.md) - Workflow interfaces
- [Utils API](api/utils.md) - Utility packages
- [Test Utils API](api/testutils.md) - Testing utilities

### [Technical Documentation](technical/) ⚙️
**For contributors and advanced users**

#### Core Architecture
- [Architecture Overview](technical/architecture.md) - System design and component structure
- [Agent Architecture](technical/agents.md) - Complete agent system implementation
- [Tool Development](technical/tool-development.md) - Internal tool architecture and patterns
- [Built-in Components](technical/built-in-components.md) - Registry system and component patterns

#### Implementation Guides
- [Provider Implementation](technical/provider-implementation.md) - How to add new LLM providers
- [Authentication System](technical/authentication.md) - Authentication architecture and patterns
- [Multimodal Content](technical/multimodal-content.md) - Technical implementation of multimodal support

#### Performance & Optimization
- [Performance Optimization](technical/performance.md) - Optimization strategies and benchmarks
- [Caching Mechanisms](technical/caching.md) - Cache implementations and strategies
- [Concurrency Patterns](technical/concurrency.md) - Thread safety and concurrent execution
- [Sync.Pool Implementation](technical/sync-pool.md) - Memory optimization with object pooling

#### Testing & Quality
- [Testing Framework](technical/testing.md) - Testing strategies and patterns
- [Benchmarking Framework](technical/benchmarks.md) - Performance measurement approach

#### Development Practices
- [Logging](technical/logging.md) - Logging patterns and best practices
- [Dependency Reduction](technical/dependency_reduction.md) - Journey from heavy dependencies to stdlib
- [Tools](technical/tools.md) - Tool system architecture and patterns
- [Structured Output Support](technical/structured-output-support.md) - LLM output parsing and validation

### [Archives](archives/) 📦
**Historical documentation**
- [Historical Documentation](archives/README.md) - Preserved documentation for reference

## 🚀 Quick Start Paths

### For New Users
1. Start with [Getting Started](user-guide/getting-started.md)
2. Understand [Core Concepts](user-guide/core-concepts.md)
3. Explore [Examples Gallery](user-guide/examples-gallery.md)

### For API Users
1. Review [LLM API](api/llm.md) for basic usage
2. Check [Agent API](api/agent.md) for advanced features
3. Use [API Reference](api/) for specific interfaces

### For Contributors
1. Read [Architecture Overview](technical/architecture.md)
2. Understand [Testing Framework](technical/testing.md)
3. Follow [Contributing Guidelines](../CONTRIBUTING.md)

### For Tool Developers
1. Start with [Tools & Components](user-guide/tools.md)
2. Study [Tool Development](technical/tool-development.md)
3. Review [Built-in Components](technical/built-in-components.md)

## 🔗 Quick Links

### Documentation Home
- **[Go-LLMs Home](/)** - Project home and quick start
- **[Examples Repository](/cmd/examples/)** - 40+ working examples
- **[CLI Documentation](/cmd/README.md)** - Command line interface
- **[Contributing Guide](../CONTRIBUTING.md)** - How to contribute

### Project Information
- **[Changelog](../CHANGELOG.md)** - Complete version history and release notes
- **[Project Status](../TODO.md)** - Current development status and roadmap
- **[Completed Tasks](../TODO-DONE.md)** - Development history

## 🎯 Examples Index

### Basic Provider Examples
- [Simple Example](/cmd/examples/simple/) - Basic usage with mock providers
- [Provider Anthropic](/cmd/examples/provider-anthropic/) - Integration with Anthropic Claude
- [Provider OpenAI](/cmd/examples/provider-openai/) - Integration with OpenAI models
- [Provider Gemini](/cmd/examples/provider-gemini/) - Integration with Google Gemini
- [Provider OpenAI Compatible](/cmd/examples/provider-openai-compatible/) - Using OpenRouter and Ollama
- [Multi-Provider](/cmd/examples/provider-multi/) - Working with multiple providers
- [Consensus](/cmd/examples/provider-consensus/) - Multi-provider consensus strategies
- [Provider Options](/cmd/examples/provider-options/) - Provider configuration system
- [Convenience](/cmd/examples/provider-convenience/) - Utility functions for common tasks
- [Multimodal](/cmd/examples/provider-multimodal/) - Working with images, audio, and video content

### Built-in Tools Examples
- [Built-in Tools Discovery](/cmd/examples/builtins-discovery/) - Discover and use built-in tools
- [Built-in File Tools](/cmd/examples/builtins-file-tools/) - Enhanced file operations
- [Built-in Web Tools](/cmd/examples/builtins-web-tools/) - Web operations (fetch, search, scrape, HTTP requests)
- [Built-in Web API Client](/cmd/examples/builtins-web-api-client/) - Advanced API client with REST, OpenAPI, and GraphQL support
- [Built-in API Client Auth](/cmd/examples/builtins-api-client-auth/) - Comprehensive authentication examples
- [Built-in OpenAPI Discovery](/cmd/examples/builtins-openapi-discovery/) - OpenAPI spec discovery and automatic configuration
- [Built-in GraphQL Client](/cmd/examples/builtins-graphql-client/) - GraphQL queries with schema introspection
- [Built-ins Web Search Parallel](/cmd/examples/builtins-web-search-parallel/) - Production API key management with parallel web searches
- [Built-in System Tools](/cmd/examples/builtins-system-tools/) - System operations (execute commands, environment variables, process list)
- [Built-in Data Tools](/cmd/examples/builtins-data-tools/) - Data processing (JSON, CSV, XML, transformations)
- [Built-in DateTime Tools](/cmd/examples/builtins-datetime-tools/) - Date and time operations
- [Built-in Feed Tools](/cmd/examples/builtins-feed-tools/) - RSS, Atom, and JSON Feed processing

### Agent Examples
- [Agent Simple LLM](/cmd/examples/agent-simple-llm/) - Ultra-simple agent creation
- [Agent LLM Built-in Tools](/cmd/examples/agent-llm-builtin-tools/) - Using built-in tools with agents
- [Agent Structured Output](/cmd/examples/agent-structured-output/) - Structured output with schemas
- [Agent Calculator](/cmd/examples/agent-calculator/) - Built-in calculator tool with LLM agents
- [Agent Custom Research](/cmd/examples/agent-custom-research/) - Custom agent with sub-agent coordination
- [Agent Sub-Agents](/cmd/examples/agent-sub-agents/) - Multi-agent coordination patterns
- [Agent Multi-Coordination](/cmd/examples/agent-multi-coordination/) - Advanced multi-agent patterns
- [Agent Tools Conversion](/cmd/examples/agent-tools-conversion/) - Converting between tools and agents
- [Agent Workflow as Tool](/cmd/examples/agent-workflow-as-tool/) - Multi-stage research pipeline
- [Agent Advanced Tool Context](/cmd/examples/agent-advanced-toolcontext/) - Advanced tool context management
- [Agent State Persistence](/cmd/examples/agent-state-persistence/) - State management and persistence
- [Agent Error Handling](/cmd/examples/agent-error-handling/) - Error handling in agents
- [Agent Guardrails](/cmd/examples/agent-guardrails/) - Agent safety and constraints
- [Agent Handoff](/cmd/examples/agent-handoff/) - Agent handoff patterns
- [Agent Metrics Tools](/cmd/examples/agent-metrics-tools/) - Performance monitoring

### Workflow Examples
- [Workflow Sequential](/cmd/examples/workflow-sequential/) - Sequential workflow patterns
- [Workflow Parallel](/cmd/examples/workflow-parallel/) - Parallel workflow execution
- [Workflow Conditional](/cmd/examples/workflow-conditional/) - Conditional workflow logic
- [Workflow Loop](/cmd/examples/workflow-loop/) - Loop-based workflows
- [Workflow Composition](/cmd/examples/workflow-composition/) - Complex workflow composition
- [Workflow Multi-Provider](/cmd/examples/workflow-multi-provider/) - Multi-provider workflows
- [Workflow Hooks](/cmd/examples/workflow-hooks/) - Workflow event handling

### Structured Output Examples
- [Structured Schema](/cmd/examples/structured-schema/) - Schema generation from Go structs
- [Structured Coercion](/cmd/examples/structured-coercion/) - Type coercion for validation

### Utility Examples
- [Utils Model Info](/cmd/examples/utils-modelinfo/) - Model discovery and capability information
- [Utils Profiling](/cmd/examples/utils-profiling/) - Performance profiling and monitoring

## 📖 Documentation Versions

This documentation corresponds to **Go-LLMs v0.3.1** (January 2025).

### Version Highlights
- Enhanced ToolBuilder pattern for all 32 built-in tools
- Comprehensive LLM guidance metadata
- MCP (Model Context Protocol) compatibility
- Advanced authentication support
- Performance improvements and documentation restructuring

For release details, see the [Changelog](../CHANGELOG.md).

## 📝 Documentation Feedback

If you find issues with the documentation or have suggestions for improvement:
1. Check the [Contributing Guidelines](../CONTRIBUTING.md)
2. Open an issue on the project repository
3. Submit a pull request with improvements

The documentation is continuously updated to reflect the latest features and best practices.