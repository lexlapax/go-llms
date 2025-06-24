# Go-LLMs Documentation Reference

Welcome to the comprehensive Go-LLMs documentation reference. This directory contains all documentation resources organized to serve different audiences and use cases.

## 📚 Documentation Structure

### [User Guide](user-guide/) 👥
**For developers using Go-LLMs** - Complete user-focused documentation with visual guides

#### 🚀 Getting Started
- [Quick Start](user-guide/getting-started/quickstart.md) - 5-minute setup with visual guide
- [Installation](user-guide/getting-started/installation.md) - Complete setup and environment configuration
- [Key Concepts](user-guide/getting-started/key-concepts.md) - Essential concepts with visual diagrams
- [First Steps](user-guide/getting-started/first-steps.md) - Progressive tutorial examples
- [Choosing Providers](user-guide/getting-started/choosing-providers.md) - Provider selection guide

#### 📖 Task-Oriented Guides
**Building Applications**
- [Chat Applications](user-guide/guides/building-chat-apps.md) - Complete chat application patterns
- [Data Extractors](user-guide/guides/building-data-extractors.md) - Data extraction workflows
- [Research Agents](user-guide/guides/building-research-agents.md) - Information gathering systems
- [Automation Agents](user-guide/guides/building-automation-agents.md) - Task automation workflows

**Working with Providers**
- [Provider Setup](user-guide/guides/provider-setup.md) - Environment configuration and API keys
- [Provider Selection](user-guide/guides/provider-selection.md) - Choosing the right provider
- [Multi-Provider Strategies](user-guide/guides/multi-provider-strategies.md) - Reliability and optimization
- [Local Providers](user-guide/guides/local-providers.md) - Ollama and local models

**Agent Development**
- [Creating Agents](user-guide/guides/creating-agents.md) - Simple to complex agent patterns
- [Agent Communication](user-guide/guides/agent-communication.md) - Coordination and handoffs
- [Agent Tools](user-guide/guides/agent-tools.md) - Using and creating tools effectively
- [Agent Memory](user-guide/guides/agent-memory.md) - State management patterns

**Data Handling**
- [Structured Data](user-guide/guides/structured-data.md) - Reliable data extraction with schemas
- [Multimodal Content](user-guide/guides/multimodal-content.md) - Images, audio, video
- [Data Validation](user-guide/guides/data-validation.md) - Validation and error recovery
- [Data Pipelines](user-guide/guides/data-pipelines.md) - End-to-end processing workflows

**Integration**
- [Web Applications](user-guide/guides/web-applications.md) - Web framework integration
- [APIs and Services](user-guide/guides/apis-and-services.md) - Building LLM-powered APIs
- [Databases](user-guide/guides/databases.md) - Storing LLM interactions
- [Existing Systems](user-guide/guides/existing-systems.md) - Adding LLM capabilities

#### 💡 Practical Examples
**By Use Case**
- [Customer Support](user-guide/examples/customer-support.md) - Complete support system
- [Content Generation](user-guide/examples/content-generation.md) - Content creation and management
- [Code Analysis](user-guide/examples/code-analysis.md) - Code review systems
- [Research Synthesis](user-guide/examples/research-synthesis.md) - Research and report generation
- [Data Analysis](user-guide/examples/data-analysis.md) - Data insights generation

**By Complexity**
- [Beginner Projects](user-guide/examples/beginner-projects.md) - 5 simple projects to get started
- [Intermediate Projects](user-guide/examples/intermediate-projects.md) - 5 practical applications
- [Advanced Projects](user-guide/examples/advanced-projects.md) - 5 complex multi-agent systems

**By Domain**
- [Business Automation](user-guide/examples/business-automation.md) - Process automation
- [Education Tools](user-guide/examples/education-tools.md) - Educational applications
- [Creative Tools](user-guide/examples/creative-tools.md) - Writing and design assistance
- [Developer Tools](user-guide/examples/developer-tools.md) - Development workflow enhancement

#### 📚 Quick Reference
- [API Quick Reference](user-guide/reference/api-quick-reference.md) - Essential API calls and patterns
- [Provider Comparison](user-guide/reference/provider-comparison.md) - Feature matrix and selection
- [Built-in Tools Reference](BUILT-IN-TOOLS-REFERENCE.md) - Complete guide to 33+ built-in tools
- [Tool Usage Examples](TOOL-USAGE-EXAMPLES.md) - Practical patterns and integration examples
- [Configuration Reference](user-guide/reference/configuration-reference.md) - All configuration options
- [Error Codes](user-guide/reference/error-codes-reference.md) - Complete error handling
- [Best Practices](user-guide/reference/best-practices-checklist.md) - Production readiness checklist

#### 🔬 Advanced Topics
- [Performance Optimization](user-guide/advanced/performance-optimization.md) - Tuning and optimization
- [Production Deployment](user-guide/advanced/production-deployment.md) - Deployment and monitoring
- [Security Considerations](user-guide/advanced/security-considerations.md) - Security best practices
- [Custom Providers](user-guide/advanced/custom-providers.md) - Creating custom providers
- [Custom Tools](user-guide/advanced/custom-tools.md) - Advanced tool development
- [Workflow Orchestration](user-guide/advanced/workflow-orchestration.md) - Complex workflows
- [Testing Strategies](user-guide/advanced/testing-strategies.md) - Testing LLM applications
- [Troubleshooting](user-guide/advanced/troubleshooting.md) - Problem diagnosis

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
**For contributors and advanced users** - Architecture and implementation details

#### 🏗️ Foundation
- [Architecture Overview](technical/architecture.md) - System design and high-level structure
- [Core Concepts](technical/core-concepts.md) - Key abstractions and design patterns

#### 🔧 Core Components
**Providers**
- [Provider Overview](technical/providers/overview.md) - Understanding LLM providers
- [Implementing Providers](technical/providers/implementing-providers.md) - Create custom providers
- [Provider Registry](technical/providers/provider-registry.md) - Dynamic registration and discovery
- [Provider Metadata](technical/providers/metadata.md) - Capabilities and configuration

**Agents**
- [Agent Overview](technical/agents/overview.md) - Agent architecture and concepts
- [LLM Agents](technical/agents/llm-agents.md) - AI-powered agents with tool support
- [Workflow Agents](technical/agents/workflow-agents.md) - Sequential, parallel, conditional, and loop patterns
- [Multi-Agent Systems](technical/agents/multi-agent-systems.md) - Coordination and communication
- [State Management](technical/agents/state-management.md) - Agent state and data flow

**Tools**
- [Tool Overview](technical/tools/overview.md) - Tool architecture and integration
- [Creating Tools](technical/tools/creating-tools.md) - Build custom tools
- [Tool Discovery](technical/tools/tool-discovery.md) - Runtime registration and metadata
- [Built-in Tools](technical/tools/built-in-tools.md) - Available tools and examples

#### 🛠️ Development
- [Contributing](technical/development/contributing.md) - Code organization and style guide
- [Testing](technical/development/testing.md) - Testing infrastructure and best practices
- [API Design](technical/development/api-design.md) - Design principles and patterns

#### 🚀 Advanced Topics
- [Performance](technical/advanced/performance.md) - Optimization strategies and benchmarking
- [Event System](technical/advanced/event-system.md) - Event architecture and serialization
- [Error Handling](technical/advanced/error-handling.md) - Error types and recovery strategies
- [Schema System](technical/advanced/schema-system.md) - JSON Schema validation and type conversion
- [Bridge Integration](technical/advanced/bridge-integration.md) - Scripting engine integration

#### 📖 Reference
- [API Reference](technical/api-reference/README.md) - Complete API documentation
- [Provider APIs](technical/api-reference/providers.md) - Provider interface documentation
- [Agent APIs](technical/api-reference/agents.md) - Agent interface documentation
- [Tool APIs](technical/api-reference/tools.md) - Tool interface documentation
- [Type Definitions](technical/api-reference/types.md) - Core type definitions

### [Archives](archives/) 📦
**Historical documentation**
- [Historical Documentation](archives/README.md) - Preserved documentation for reference

## 🚀 Quick Start Paths

### 🌱 For New Users (Beginner Path)
1. [Quick Start](user-guide/getting-started/quickstart.md) - 5-minute setup
2. [Key Concepts](user-guide/getting-started/key-concepts.md) - Understand the basics
3. [Beginner Projects](user-guide/examples/beginner-projects.md) - Try 5 simple projects
4. [Chat Applications](user-guide/guides/building-chat-apps.md) - Build your first app

### 🚀 For Application Developers (Developer Path)
1. [Provider Setup](user-guide/guides/provider-setup.md) - Professional environment setup
2. [Creating Agents](user-guide/guides/creating-agents.md) - Build your first agents
3. [Agent Tools](user-guide/guides/agent-tools.md) - Add capabilities with tools
4. [Structured Data](user-guide/guides/structured-data.md) - Reliable data extraction
5. [API Quick Reference](user-guide/reference/api-quick-reference.md) - Essential patterns

### 🏗️ For System Architects (Architect Path)
1. [Agent Communication](user-guide/guides/agent-communication.md) - Multi-agent coordination
2. [Multi-Provider Strategies](user-guide/guides/multi-provider-strategies.md) - Robust provider management
3. [Data Pipelines](user-guide/guides/data-pipelines.md) - End-to-end data processing
4. [Performance Optimization](user-guide/advanced/performance-optimization.md) - Scale and optimize
5. [Architecture Overview](technical/architecture.md) - System design details

### 🚀 For Production Deployment (Production Path)
1. [Security Considerations](user-guide/advanced/security-considerations.md) - Secure your application
2. [Production Deployment](user-guide/advanced/production-deployment.md) - Deploy and monitor
3. [Testing Strategies](user-guide/advanced/testing-strategies.md) - Test LLM applications
4. [Best Practices](user-guide/reference/best-practices-checklist.md) - Production checklist
5. [Troubleshooting](user-guide/advanced/troubleshooting.md) - Handle issues

### ⚙️ For Contributors & Advanced Users
1. [Architecture Overview](technical/architecture.md) - System design and structure
2. [Core Concepts](technical/core-concepts.md) - Key abstractions and patterns
3. [Testing Framework](technical/development/testing.md) - Testing infrastructure
4. [Contributing Guidelines](technical/development/contributing.md) - Code organization and style

### 🛠️ For Tool Developers
1. [Built-in Tools Reference](BUILT-IN-TOOLS-REFERENCE.md) - Complete guide to 33+ built-in tools
2. [Tool Usage Examples](TOOL-USAGE-EXAMPLES.md) - Practical patterns and integration
3. [Tool Overview](technical/tools/overview.md) - Tool architecture and integration
4. [Creating Tools](technical/tools/creating-tools.md) - Build custom tools
5. [Tool Discovery](technical/tools/tool-discovery.md) - Runtime registration and metadata

## 🔗 Quick Links

### Documentation Home
- **[Go-LLMs Home](/)** - Project home and quick start
- **[User Guide](user-guide/README.md)** - Complete user documentation with visual guides
- **[Technical Documentation](technical/README.md)** - Architecture and implementation details
- **[Examples Repository](/cmd/examples/)** - 80+ working examples
- **[CLI Documentation](/cmd/README.md)** - Command line interface
- **[Contributing Guide](../CONTRIBUTING.md)** - How to contribute

### Project Information
- **[Changelog](../CHANGELOG.md)** - Complete version history and release notes
- **[Project Status](../TODO.md)** - Current development status and roadmap
- **[Completed Tasks](../TODO-DONE.md)** - Development history
- **[CLAUDE.md](../CLAUDE.md)** - Project guidance for AI assistants
- **[Documentation Style Guide](../CONTRIBUTING-DOCS.md)** - Standards for code documentation

### Visual Resources
- **[Images Directory](images/)** - SVG diagrams and visual guides
- **[Architecture Diagrams](images/)** - System design visualizations
- **[Workflow Patterns](images/)** - Agent coordination patterns
- **[Learning Paths](images/)** - Visual learning guides

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

This documentation corresponds to **Go-LLMs v0.3.5** (June 2025).

### Version Highlights
- ✅ **Complete v0.3.5 Release** - All scripting engine requirements implemented
- 🎨 **Visual Documentation** - Comprehensive SVG diagrams and visual guides
- 📚 **User-Focused Structure** - Task-oriented guides and learning paths
- 🧪 **Testing Infrastructure** - Complete testing utilities with mocks, fixtures, and scenarios
- 🔧 **Bridge Integration** - Enhanced scripting engine compatibility
- 📊 **Structured Output** - Advanced schema validation and type conversion
- 🛠️ **Tool Discovery** - Dynamic tool registration and metadata
- ⚡ **Performance Optimization** - Enhanced error handling and state management
- 🏗️ **Production Ready** - Comprehensive deployment and monitoring guidance

### Recent Documentation Improvements
- New visual learning paths and decision trees
- Comprehensive user guide with 5 learning paths
- 5 new SVG diagrams enhancing key concepts
- Complete testing strategies guide
- Reorganized technical documentation
- Quick reference materials for all skill levels

For release details, see the [Changelog](../CHANGELOG.md).

## 📝 Documentation Feedback

If you find issues with the documentation or have suggestions for improvement:
1. Check the [Contributing Guidelines](../CONTRIBUTING.md)
2. Open an issue on the project repository
3. Submit a pull request with improvements

The documentation is continuously updated to reflect the latest features and best practices.