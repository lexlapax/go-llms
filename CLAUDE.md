# CLAUDE.md

Project guidance for Claude Code when working with go-llms.

## Project: Go-LLMs v0.3.4

Unified Go interface for LLM providers with agent tooling.

**Status**: v0.3.5.9 Testing Infrastructure (Phase 1 Complete - June 14, 2025)
**Providers**: OpenAI, Anthropic, Google (Gemini, Vertex AI), Ollama, OpenRouter  
**Current**: Phase 2 - Scenario Builder System
**Next**: Complete v0.3.5.9 Testing Infrastructure

## Recent Progress:
### v0.3.5.8 Structured Output Support ✅
- Implemented output parser interface with JSON, XML, YAML parsers
- Recovery mechanisms for malformed LLM outputs
- Schema-based validation
- Format conversion between JSON/XML/YAML
- Bridge adapter for go-llmspell compatibility

### v0.3.5.9 Testing Infrastructure (Phase 1 Complete)
- Enhanced mock implementations with pattern-based responses ✅
- Centralized mock registry ✅
- Call history tracking ✅
- Phase 1 completed:
  - MockAgent with response queues, sub-agent management, event tracking ✅
  - MockState with change tracking, snapshots, behavior hooks ✅
  - MockEventEmitter with recording, filtering, assertions ✅
  - Comprehensive test coverage for all mocks ✅
  - Fixed import cycles by removing circular dependencies ✅
- Next: Phase 2 Scenario Builder System

## Current Test Status:
- All structured output parsers: Tests passing ✅
- Mock implementations: All tests passing ✅
  - MockAgent: Complete with all features tested
  - MockState: Complete with change tracking, snapshots
  - MockEventEmitter: Complete with event assertions
- Unit tests: Passing (except integration/stress tests which require API keys)

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