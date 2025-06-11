# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

Go-LLMs is a Go library providing a unified interface for various LLM providers (OpenAI, Anthropic, Google Gemini, Ollama, etc.) with robust data validation and agent tooling.

**Version**: v0.3.2 (Documentation Update - January 2025)

**Status**: Documentation completely restructured following user journey patterns. All 32 tools support ToolBuilder pattern with MCP compatibility. Ollama provider integration completed (v0.3.3.1).

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

- **v0.3.3.1**: Ollama provider integration complete (January 11, 2025)
- **v0.3.2**: Documentation restructuring complete (January 11, 2025)
- **v0.3.1**: Tool system enhancement complete (January 10, 2025)
- **Next**: v0.3.3.2+ - Additional providers (OpenRouter, Vertex AI, Mistral, AWS Bedrock, Azure AI)

See TODO.md for roadmap and TODO-DONE.md for completed items.