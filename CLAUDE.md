# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

Go-LLMs is a Go library providing a unified interface for various LLM providers (OpenAI, Anthropic, Google Gemini, OpenRouter, Ollama, etc.) with robust data validation and agent tooling.

**Version**: v0.3.3 (OpenRouter Provider - January 2025)

**Status**: OpenRouter provider integration completed with access to 400+ models. Documentation updated to clarify base URL handling for OpenAI-compatible providers.

## Key Commands

```bash
make build                # Build main binary
make test                 # Run unit tests
make test-all            # Run all tests including integration
make lint                # Run linting
make fmt                 # Format code
```

## Architecture

1. **Schema Validation** (`pkg/schema/`): JSON validation with type coercion
2. **LLM Integration** (`pkg/llm/`): Provider implementations
3. **Structured Output** (`pkg/structured/`): Extract structured data from LLMs
4. **Agent Workflows** (`pkg/agent/`): Tools, state management, workflows

## Development Guidelines

- No backward compatibility until v1.0
- No direct logging in library code (pkg/)
- Use `sync.Pool` for performance
- Follow existing patterns and style
- Add comprehensive tests
- Run `make fmt` and `make vet` before committing

## Current Status

- **v0.3.3**: OpenRouter provider integration complete (January 11, 2025)
- **v0.3.2**: Documentation restructuring complete (January 11, 2025)
- **v0.3.1**: Tool system enhancement complete (January 10, 2025)
- **Next**: v0.3.4 - Built-in Agents Library

See TODO.md for roadmap and TODO-DONE.md for completed items.