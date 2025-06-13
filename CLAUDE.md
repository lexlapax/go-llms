# CLAUDE.md

Project guidance for Claude Code when working with go-llms.

## Project: Go-LLMs v0.3.4

Unified Go interface for LLM providers with agent tooling.

**Status**: v0.3.4.1 Tool Discovery System COMPLETED (June 13, 2025)
**Providers**: OpenAI, Anthropic, Google (Gemini, Vertex AI), Ollama, OpenRouter  
**Next**: v0.3.4.5 - Web API Client Advanced Features

## Recent Completion: Tool Discovery System ✅
- Metadata-first tool exploration (33+ tools) without imports
- Perfect for scripting engines (go-llmspell integration)
- Lazy loading with factory pattern and build tag isolation
- Comprehensive documentation and examples
- All tests passing with excellent performance

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