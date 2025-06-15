# Changelog

All notable changes to the Go-LLMs project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.3.5] - 2025-06-15

### Overview

Go-LLMs v0.3.5 is the Scripting Integration release, providing comprehensive support for go-llmspell integration with enhanced schema capabilities, structured outputs, event system improvements, workflow serialization, and a completely revamped testing infrastructure. This release completes all requirements for seamless scripting engine bridge integration.

### Added

#### Schema System (v0.3.5.1-v0.3.5.3)
- **Schema Repositories**: InMemory and File-based storage with versioning support
- **Schema Generators**: Reflection and Tag-based generation from Go structs
- **Schema Validation**: Comprehensive validation with detailed error reporting
- **Bridge Integration**: Type-safe conversion utilities for go-llmspell
- **Examples**: Complete schema repository and generator examples
- **Documentation**: `docs/technical/schema-system.md` and `docs/technical/schema-package.md`

#### Structured Output Support (v0.3.5.8)
- **Output Parsers**: JSON, XML, and YAML parsers with intelligent recovery
- **Recovery Mechanisms**: Handle malformed LLM outputs automatically
  - Extract from markdown code blocks
  - Fix trailing commas and quote issues
  - Close unclosed tags and fix missing decimals
- **Schema-Based Validation**: Validate outputs against defined schemas
- **Format Conversion**: Convert between JSON, XML, and YAML formats
- **Bridge Adapter**: Seamless integration with go-llmspell
- **Documentation**: `docs/technical/structured-output-support.md`

#### Event System Enhancements (v0.3.5.4-v0.3.5.5)
- **Event Bus**: Thread-safe publish/subscribe with pattern matching
- **Bridge Events**: Specialized events for script integration
  - Bridge lifecycle events (connected, disconnected, ready)
  - Communication events (request, response, callback)
  - Type conversion events with error tracking
  - Script execution events with output capture
- **Event Serialization**: JSON-based serialization for bridge communication
- **Event Storage**: Persistent event storage with filtering
- **Bridge Integration**: `pkg/agent/events/bridge.go` with publisher/listener

#### Workflow Serialization (v0.3.5.6-v0.3.5.7)
- **Workflow Serialization**: Export/import workflows as JSON
- **Script Handlers**: Integration points for scripting engines
- **State Preservation**: Maintain workflow state across serialization
- **Test Coverage**: Comprehensive tests for all serialization paths

#### Testing Infrastructure (v0.3.5.9)
- **Centralized Mocks**: Complete mock implementations in `pkg/testutils/mocks/`
  - MockProvider with pattern-based responses
  - MockAgent with sub-agent management and event tracking
  - MockTool with builder pattern and execution hooks
  - MockState with change tracking and snapshots
  - MockEventEmitter with recording and assertions
- **Fixture Library**: 37+ fixtures for common test scenarios
  - 14 tool fixtures (file, web, calculation, datetime, data)
  - 12 provider fixtures (OpenAI, Anthropic, Gemini patterns)
  - 8 agent fixtures (workflow, multi-agent, error scenarios)
- **Migration**: 47 files migrated to centralized infrastructure
- **Documentation**: Comprehensive testing guide at `docs/technical/testing.md`

#### Documentation and API Generation (v0.3.5.10)
- **Documentation Generation System**: Complete `pkg/docs/` package with auto-generation
  - Generator interface with OpenAPI, Markdown, and JSON output formats
  - Documentable interface for auto-documentation support
  - GenerateOpenAPIForTool() function for individual tool documentation
- **Tool Discovery Integration**: Seamless integration with existing tool discovery
  - Automatic documentation generation for all 33 discovered tools
  - Category-based, tag-based, and search-based filtering
  - Enhanced tool help with rich formatting
- **Bridge-Friendly Design**: All types JSON serializable for go-llmspell
  - Complete OpenAPI 3.0 specification generation (142KB for all tools)
  - Human-readable Markdown documentation (207KB comprehensive docs)
  - Structured JSON output for programmatic access
- **Comprehensive Example**: Working example demonstrating all features
  - Real-world generation from actual tool discovery system
  - Multiple output formats and filtering options

### Enhanced

#### Bridge-Friendly Type System
- Type-safe serialization for all domain objects
- Schema conversion utilities for script engines
- Event types designed for bridge communication
- Metadata-rich tool discovery for dynamic loading

#### Runtime Tool Registration
- Already implemented in v0.3.4, enhanced with schema support
- Tool schemas accessible without instantiation
- Factory pattern enables lazy loading in scripts

### Performance

- **Structured Output Parsing**: Direct parsing before recovery attempts
- **Event Bus**: Efficient pattern matching with minimal allocations
- **Schema Generation**: Reflection caching for repeated generations
- **Mock Performance**: Lightweight mocks with minimal overhead

### Documentation

- Created `docs/technical/schema-system.md` - Overview of schema architecture
- Created `docs/technical/schema-package.md` - Schema implementation details
- Updated `docs/technical/testing.md` - Complete testing infrastructure guide
- Enhanced `docs/technical/structured-output-support.md` - Output parsing guide
- Comprehensive examples for all new features

### Fixed

- Circular import issues in test infrastructure
- Mock implementations now properly isolated
- Event serialization edge cases
- Schema validation error reporting

### Testing

- All tests passing (except integration tests requiring API keys)
- 100% coverage for new mock implementations
- Comprehensive fixture test coverage
- Migration completed without breaking existing tests

## [v0.3.4] - 2025-06-13

### Overview

Go-LLMs v0.3.4 introduces the groundbreaking Tool Discovery System, a metadata-first approach that enables dynamic tool exploration without imports. This release is specifically designed for scripting engines like go-llmspell, providing seamless bridge integration for Lua/JavaScript environments while maintaining excellent performance and comprehensive tooling.

### Added

#### Tool Discovery System (v0.3.4.1)
- **Metadata-First Discovery**: Explore 33+ tools without requiring package imports
- **Lazy Loading**: Factory pattern with on-demand tool instantiation
- **Build Tag Isolation**: Avoid import cycles and compilation bloat using conditional builds
- **Rich Metadata Access**: Get schemas, examples, and help text without tool instances
- **Bridge Integration**: Perfect for go-llmspell Lua/JavaScript bridges

#### Core Discovery API
- `ToolDiscovery` interface in `pkg/agent/tools/discovery.go`
- `ListTools()` - Return all available tools without loading them
- `SearchTools(query)` - Filter by keyword, category, tags, description
- `GetToolSchema(name)` - Detailed parameter/output schemas
- `GetToolExamples(name)` - Usage examples with input/output
- `CreateTool(name)` - Lazy tool instantiation
- `GetToolHelp(name)` - Formatted help text generation

#### Enhanced Registry Features
- Global `GetToolMetadata()` function for bridge access
- Category-based tool grouping with `ListByCategory()`
- Tag-based tool discovery and search capabilities
- Tool factories with actual constructor function names
- Metadata extraction from ToolBuilder calls via AST parsing

#### Code Generation System
- AST-based tool metadata extractor in `internal/toolgen/`
- Generated `pkg/agent/tools/registry_metadata.go` with compile-time metadata
- Generated `pkg/agent/tools/registry_factories.go` with conditional imports
- `//go:generate` directive for automatic regeneration
- `make generate` target for easy metadata updates

### Enhanced

#### Examples and Documentation
- **Enhanced builtins-discovery example**: Comprehensive demonstrations of both new discovery API and legacy registry
- **Technical documentation**: Complete API reference at `docs/technical/tool-discovery-api.md`
- **Integration examples**: Shows metadata-first exploration, lazy loading, and scripting bridge patterns
- **Migration guidance**: Clear comparison between legacy and new approaches

#### Testing and Quality
- **Unit tests**: Comprehensive coverage for metadata extraction and factory pattern
- **Integration tests**: Full discovery API testing (moved to `tests/integration/`)
- **Benchmark tests**: Performance validation ensuring no regression (moved to `tests/benchmarks/`)
- **Code quality**: All lint errors fixed, proper file organization

### Performance

#### Benchmark Results
- **ListTools**: ~5μs per operation with minimal allocations (1 alloc)
- **SearchTools**: ~14μs per operation for keyword searches (36 allocs)
- **ListByCategory**: Very fast at ~550ns per operation (1 alloc)
- **Schema operations**: 35-70μs for complex schema access (241 allocs)
- **Concurrent access**: Excellent scalability with thread-safe operations

### Key Benefits

#### For Scripting Engines (go-llmspell)
- **Zero Imports**: Explore tools without package dependencies
- **Metadata Access**: Get schemas, examples, help without tool instances
- **Lazy Loading**: Create tools only when actually needed
- **Build Tag Safe**: Avoid import cycles and compilation issues
- **Bridge Friendly**: Designed for Lua/JavaScript integration

#### For Developers
- **Dynamic Discovery**: Find tools at runtime based on needs
- **Rich Metadata**: Access parameter schemas, examples, constraints
- **Help Generation**: Get formatted documentation for any tool
- **Search & Filter**: Find tools by keyword, category, or tags
- **Performance**: No upfront tool loading costs

### Technical Architecture

#### Metadata-First Design
- Tool metadata separated from implementation
- JSON-based schema storage for script accessibility
- Factory pattern enables lazy instantiation
- Build tags prevent unwanted imports in scripting environments

#### Bridge Integration Pattern
```go
// Perfect for scripting engines
metadata := tools.GetToolMetadata()
for name, info := range metadata {
    // Expose to Lua/JavaScript bridge
    bridge.ExposeToolMetadata(name, info)
}
```

### Documentation
- Complete API reference with examples and integration patterns
- Enhanced technical documentation index with proper navigation
- Migration guide from legacy registry approach
- Bridge integration examples for go-llmspell

### Changed
- Updated technical documentation index to include Tool Discovery API
- Enhanced examples to demonstrate both legacy and new approaches
- Improved project structure with proper test organization

## [v0.3.3] - 2025-01-11

### Overview

Go-LLMs v0.3.3 is a major provider expansion release that adds support for three new LLM providers: Ollama for local model hosting, OpenRouter for unified access to 400+ models, and Google Vertex AI for enterprise deployments. This release significantly expands the library's reach, enabling users to run models locally, access a vast array of models through a single API, and deploy in enterprise Google Cloud environments.

### Added

#### Ollama Provider Support (v0.3.3.1)
- New `NewOllamaProvider()` convenience function for easy local LLM integration
- Model discovery support via `/api/tags` endpoint
- Full integration with utility systems (env vars, option factories, CLI)
- Ollama-specific error handling including OOM detection
- Example application demonstrating Ollama usage
- Documentation in user guide and provider implementation guide

#### OpenRouter Provider Support (v0.3.3.2)
- New `NewOpenRouterProvider()` with access to 400+ models from various providers
- Automatic model discovery via `/api/v1/models` endpoint
- Support for 68 free models without API costs
- Full streaming support with OpenAI-compatible API
- Cost-optimized model selection and automatic fallbacks
- Privacy-focused options (no logging by default)
- Example demonstrating multi-model usage and free tier access
- Comprehensive integration tests and documentation

#### Google Vertex AI Provider Support (v0.3.3.3)
- New `NewVertexAIProvider()` for enterprise Google Cloud deployments
- OAuth2 authentication with service account and ADC support
- Access to Google's Gemini models and partner models (Claude via Vertex)
- Regional deployment support for data residency requirements
- IAM integration for fine-grained access control
- Full multimodal support with Gemini models
- Streaming responses with Server-Sent Events
- Model discovery with hardcoded catalog (no public API available)
- Comprehensive example showing authentication methods and regional deployment
- Enterprise-focused documentation with IAM setup instructions

### Changed
- Enhanced documentation for OpenAI-compatible providers clarifying base URL usage
- Updated all provider integration points (CLI, config, utilities) to support new providers
- Expanded environment variable support for provider-specific configurations

### Documentation
- Added dedicated sections for each new provider in user guide
- Updated README with expanded provider list
- Enhanced provider implementation guide with new provider examples
- Added enterprise deployment guidance for Vertex AI

## [v0.3.2] - 2025-01-11

### Overview

Go-LLMs v0.3.2 is a documentation update release that significantly improves the organization, clarity, and accessibility of all project documentation. This release restructures documentation to follow a user journey approach and consolidates redundant content.

### Changed

#### Documentation Structure Overhaul
- **API Documentation** (`docs/api/`)
  - Modularized API docs with dedicated files for each major component
  - Added tools.md, workflows.md, builtins.md, utils.md, testutils.md
  - Updated existing docs to cross-reference new modular structure
  - Improved navigation with comprehensive index

- **User Guide** (`docs/user-guide/`)
  - Restructured to follow natural learning progression
  - New getting-started.md and core-concepts.md for beginners
  - Consolidated tool documentation from multiple sources
  - Merged web-search-tool.md into tools.md
  - Enhanced examples-gallery.md with 40+ categorized examples

- **Technical Documentation** (`docs/technical/`)
  - Added provider-implementation.md for custom provider development
  - Added tool-development.md for internal tool architecture
  - Updated logging.md with agent hook system and debug infrastructure
  - Updated tools.md with ToolContext system documentation
  - Removed duplicate content, keeping technical details separate from user guides

- **Archives** (`docs/archives/`)
  - Moved historical documentation from docs/plan/
  - Preserved design documents and migration guides
  - Consistent file naming (underscores to hyphens)

### Added
- Comprehensive CHANGELOG.md consolidating all release notes
- Enhanced docs/README.md as central documentation hub
- Documentation for recent features:
  - Agent hook system (LoggingHook, LLMMetricsHook)
  - ToolContext system with StateReader and EventEmitter
  - Debug infrastructure with build tags

### Removed
- REFERENCE.md (merged into docs/README.md)
- RELEASE_NOTES_v0.3.1.md (consolidated into CHANGELOG.md)
- Redundant documentation files
- docs/plan/ directory (moved to archives)

### Documentation
- Improved cross-linking between all documentation files
- Added quick start paths for different user types
- Enhanced navigation with categorized content
- Updated all example READMEs for consistency
- Fixed broken links throughout documentation

## [v0.3.1] - 2025-01-10

### Overview

Go-LLMs v0.3.1 is a major milestone release that completes the Tool System Enhancement initiative, bringing comprehensive improvements to all 32 built-in tools with the new ToolBuilder pattern. This release provides enhanced LLM integration, better error handling, and full Model Context Protocol (MCP) compatibility.

### Added

#### ToolBuilder Pattern Migration
- All 32 built-in tools migrated to enhanced ToolBuilder pattern
- Rich metadata including usage instructions, constraints, and examples
- LLM-optimized error messages and guidance
- Full MCP (Model Context Protocol) compatibility
- Improved tool discovery and categorization

#### API Client Tool v3.0.0
- GraphQL support with query/mutation execution and variables
- Schema introspection and automatic discovery
- OpenAPI integration with automatic server URL resolution
- Enhanced authentication with automatic credential detection
- Context-aware error messages and guidance

#### Enhanced Tool Metadata
- 3-7 usage examples per tool with input/output
- Comprehensive constraints documentation
- Error guidance mapping for common failures
- Resource usage indicators
- MCP export capability for all tools

#### Calculator Tool v2.0.0
- Extended mathematical constants (phi, tau, sqrt variants)
- Enhanced LLM integration mode as default
- Provider/model information display
- DEBUG environment variable support

### Changed

#### Tool Categories Enhanced

**Web Tools (6 tools)**
- Enhanced authentication support across all web tools
- Improved error handling and retry logic
- Better timeout management

**File Tools (6 tools)**
- Large file support with streaming
- Atomic write operations
- Enhanced search with regex patterns
- Binary file detection

**System Tools (4 tools)**
- Safe command execution with timeouts
- Environment variable pattern matching
- Comprehensive system information
- Process filtering and monitoring

**Data Tools (4 tools)**
- JSONPath query support
- CSV statistics and filtering
- XML to JSON conversion
- Functional transformations (map, filter, reduce)

**DateTime Tools (7 tools)**
- Natural language date parsing
- Business day calculations
- Multi-timezone support
- Localized formatting (6 languages)

**Feed Tools (6 tools)**
- Multi-format support (RSS, Atom, JSON Feed)
- Feed discovery and aggregation
- Content filtering and extraction
- Format conversion between feed types

### Fixed
- Hardcoded URL detection in authentication system
- Token detection patterns for various providers
- Linting issues across feed tools
- Example patterns in agent-calculator

### Documentation
- Created comprehensive tool development guide
- Updated built-in tools documentation
- Added examples gallery with 40+ examples
- Improved cross-linking and navigation
- Enhanced READMEs across all examples

### Performance
- Test Coverage: 44.3% overall with 280+ tests passing
- API Client: ~115μs for simple GET requests
- Tool Execution: ~6.3μs per tool call
- State Operations: ~67ns for get/set operations
- Workflow Execution: ~22μs for sequential workflows

## [v0.3.0] - 2024-12-15

### Added
- Agent Architecture Restructuring with new domain-driven design
- Enhanced core infrastructure with Handoff, Guardrails, and StateValidator
- Workflow agents: Sequential, Parallel, Conditional, and Loop
- Multi-agent system enhancements with automatic sub-agent tools
- Comprehensive hook system for logging and metrics
- Built-in components infrastructure for tools, agents, and workflows

### Changed
- Migrated from DefaultAgent to new LLMAgent architecture
- Ultra-simple agent creation from provider/model strings
- Improved state management with SharedStateContext
- Enhanced tool-agent bidirectional conversion utilities

### Removed
- Deprecated DefaultAgent and UnoptimizedDefaultAgent
- Old workflow package implementation

## [v0.2.0] - 2024-09-20

### Added
- Multimodal content support for all providers
- ContentPart structure supporting text, images, files, videos, and audio
- Helper functions for creating different message types
- Provider-specific multimodal conversions
- Comprehensive multimodal example

### Changed
- Enhanced message structure to support multimodal content
- Updated all providers to handle multimodal inputs

### Documentation
- Technical documentation for multimodal implementation
- User guide for working with multimodal content
- Integration examples for each content type

## [v0.1.0] - 2024-06-15

### Initial Release
- Core LLM provider implementations (OpenAI, Anthropic, Gemini)
- Schema validation with type coercion
- Structured output extraction
- Basic agent implementation
- Tool system foundation
- Workflow patterns (sequential, parallel)
- Comprehensive test suite
- Initial documentation

## [Unreleased]

### Planned for v0.4.0
- Built-in Agents Library
- Text processing agents (summarize, extract, analyze, translate)
- Research agents (web researcher, document analyzer, fact checker)
- Coding agents (code reviewer, test generator, doc writer)

### Planned for v0.4.1
- Built-in Workflow Patterns
- Pipeline, MapReduce, Consensus patterns
- Retry with exponential backoff
- Example workflows for common use cases

### Planned for v0.4.2
- Enhanced Tool Capabilities with API pagination and rate limiting
- Advanced authentication system improvements
- OAuth2 discovery via .well-known endpoints
- Request/response middleware plugin system

---

For detailed migration guides and breaking changes, please refer to the documentation in the `/docs` directory.