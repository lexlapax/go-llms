# CLAUDE.md

Project guidance for Claude Code when working with go-llms.

## Project: Go-LLMs v0.3.3

Unified Go interface for LLM providers with agent tooling.

**Providers**: OpenAI, Anthropic, Google (Gemini, Vertex AI), Ollama, OpenRouter  
**Next**: v0.3.4 - Mistral AI provider, then Built-in Agents Library

## Commands
```bash
make test        # Unit tests
make test-all    # All tests
make lint fmt    # Lint & format
```

## Key Rules
- No backward compatibility until v1.0
- No logging in pkg/ (library code)
- Follow existing patterns
- Run `make fmt lint` before committing

## Architecture
- `pkg/llm/` - Provider implementations
- `pkg/agent/` - Tools, state, workflows
- `pkg/schema/` - JSON validation
- `pkg/structured/` - Structured outputs

See TODO.md for roadmap.