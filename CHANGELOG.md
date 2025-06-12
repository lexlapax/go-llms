# Changelog

All notable changes to the Go-LLMs project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

### Planned for v0.3.2
- Enhanced Tool Capabilities with API pagination and rate limiting
- Advanced authentication system improvements
- OAuth2 discovery via .well-known endpoints
- Request/response middleware plugin system

### Planned for v0.3.3
- Built-in Agents Library
- Text processing agents (summarize, extract, analyze, translate)
- Research agents (web researcher, document analyzer, fact checker)
- Coding agents (code reviewer, test generator, doc writer)

### Planned for v0.3.4
- Built-in Workflow Patterns
- Pipeline, MapReduce, Consensus patterns
- Retry with exponential backoff
- Example workflows for common use cases

---

For detailed migration guides and breaking changes, please refer to the documentation in the `/docs` directory.