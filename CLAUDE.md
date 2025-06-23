# CLAUDE.md

Project guidance for Claude Code when working with go-llms.

## Project: Go-LLMs v0.3.6.1 ✅ DOCUMENTATION COMPLETED

Unified Go interface for LLM providers with agent tooling.

**Status**: v0.3.5 Released (June 15, 2025), v0.3.6.1 Documentation Completed (January 23, 2025)
**Providers**: OpenAI, Anthropic, Google (Gemini, Vertex AI), Ollama, OpenRouter  
**Build Status**: All checks passing ✅

## v0.3.6.1 Documentation Improvements:
- ✅ **Complete Godoc Documentation** - All Go files with comprehensive godoc comments
- ✅ **ABOUTME Comments** - Standardized 2-line file descriptions for easy identification
- ✅ **Documentation Style Guide** - CONTRIBUTING-DOCS.md with comprehensive standards
- ✅ **Package Documentation** - Enhanced package-level documentation across all modules
- ✅ **Cross-Reference Updates** - Updated README.md and docs/ to reference style guide

## v0.3.5 Features:
Complete scripting engine integration with:
- Schema Package & JSON validation
- Serializable error system
- Dynamic tool discovery & registration  
- Type conversion registry
- Event system with serialization/filtering
- Workflow templates & script steps
- Provider metadata & capabilities
- Structured output parsers
- Comprehensive testing infrastructure (280+ tests)
- Documentation generation (OpenAPI, Markdown, JSON)

## Commands
```bash
make test fmt lint    # Standard workflow
make test-all         # All tests including integration
make generate         # Generate tool metadata
```

## Development Rules
- No backward compatibility until v1.0
- No logging in pkg/ (library code)  
- Run `make fmt lint` before committing
- Put common test fixtures, mocks, helpers and scenarios in pkg/testutils for reuse
- Follow CONTRIBUTING-DOCS.md for all code documentation
- All Go files must include ABOUTME comments (2 lines starting with "ABOUTME: ")
- All exported functions, types, and interfaces require comprehensive godoc comments


## Architecture
- `pkg/llm/` - Provider implementations
- `pkg/agent/` - Tools, workflows, discovery
- `pkg/schema/` - JSON validation
- `pkg/structured/` - Structured outputs
- `pkg/errors/` - Serializable error system
- `pkg/testutils/` - Testing infrastructure