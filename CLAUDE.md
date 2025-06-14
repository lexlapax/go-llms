# CLAUDE.md

Project guidance for Claude Code when working with go-llms.

## Project: Go-LLMs v0.3.4

Unified Go interface for LLM providers with agent tooling.

**Status**: v0.3.5.9 Testing Infrastructure (In Progress - June 14, 2025)
**Providers**: OpenAI, Anthropic, Google (Gemini, Vertex AI), Ollama, OpenRouter  
**Current Issue**: XML parser test failing for multiple root elements case
**Next**: Complete v0.3.5.9 Testing Infrastructure

## Recent Progress:
### v0.3.5.8 Structured Output Support ✅
- Implemented output parser interface with JSON, XML, YAML parsers
- Recovery mechanisms for malformed LLM outputs
- Schema-based validation
- Format conversion between JSON/XML/YAML
- Bridge adapter for go-llmspell compatibility

### v0.3.5.9 Testing Infrastructure (In Progress)
- Enhanced mock implementations with pattern-based responses
- Centralized mock registry
- Call history tracking
- Currently fixing: XML parser test case for multiple root elements

## Current Test Status:
- YAML parser: All tests passing ✅
- JSON parser: All tests passing ✅
- Converter: All tests passing ✅
- Validator: All tests passing ✅
- Bridge adapter: All tests passing ✅
- XML parser: 1 test failing (Multiple root elements wrapped) ❌
  - Issue: Parser only returns first element when multiple root elements exist
  - The `hasMultipleRootElements` detection is working, but wrapping strategy not applied correctly

## Upcoming: v0.3.5 Scripting Integration
Major focus on go-llmspell requirements:
- Schema implementations (repositories, generators)
- Enhanced tool discovery with runtime registration
- Bridge-friendly type system
- Event system improvements
- Workflow serialization
- Testing infrastructure

## Commands
```bash
make test        # Unit tests
make test-all    # All tests  
make lint fmt    # Lint & format
make generate    # Generate tool metadata
```

## Key Rules
- No backward compatibility until v1.0
- No logging in pkg/ (library code)
- Follow existing patterns
- Run `make fmt lint` before committing

## Architecture
- `pkg/llm/` - Provider implementations
- `pkg/agent/` - Tools, state, workflows, discovery
- `pkg/schema/` - JSON validation
- `pkg/structured/` - Structured outputs

See TODO.md for roadmap.