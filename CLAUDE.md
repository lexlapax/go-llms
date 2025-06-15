# CLAUDE.md

Project guidance for Claude Code when working with go-llms.

## Project: Go-LLMs v0.3.5

Unified Go interface for LLM providers with agent tooling.

**Status**: v0.3.5 Complete - All scripting engine requirements implemented (June 15, 2025)
**Providers**: OpenAI, Anthropic, Google (Gemini, Vertex AI), Ollama, OpenRouter  
**Current**: v0.3.5 testing infrastructure complete, awaiting downstream integration
**Next**: v0.3.6+ features based on downstream feedback
**Build Status**: make test; make fmt; make vet; make lint - All passing ✅

## v0.3.5 Complete:
All scripting engine integration requirements implemented:
- ✅ Schema Package (repositories, generators)
- ✅ Enhanced Error Handling (serializable errors, recovery strategies)
- ✅ Tool Discovery (dynamic registration, persistence)
- ✅ Bridge-Friendly Types (conversion registry)
- ✅ Event System (serialization, filtering, replay)
- ✅ Workflow Serialization (templates, script steps)
- ✅ Provider Metadata (capabilities, configuration)
- ✅ Structured Output (parsers, validators)
- ✅ Testing Infrastructure (mocks, scenarios, fixtures)
- ✅ Documentation Generation (OpenAPI, Markdown, JSON)

## Test Status:
- All tests passing ✅ (280+ unit tests)
- Comprehensive test infrastructure with mocks, scenarios, fixtures
- Integration tests require API keys

## Ready for Downstream:
All v0.3.5 requirements complete and ready for go-llmspell integration:
- Bridge-compatible error system
- Runtime tool registration
- Type conversion registry
- Event serialization/filtering
- Workflow templates with script steps
- Comprehensive testing utilities

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
- `pkg/errors/` - Serializable error system
- `pkg/testutils/` - Testing infrastructure

## Important Files
- `TODO.md` - Current tasks (v0.3.6+ deferred items)
- `TODO-DONE.md` - Completed v0.3.x features
- `pkg/testutils/` - Comprehensive test infrastructure
- `docs/technical/testing.md` - Testing documentation