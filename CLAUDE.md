# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

Go-LLMs is a Go library providing a unified interface for various LLM providers (OpenAI, Anthropic, Google Gemini, etc.) with robust data validation and agent tooling.

**Version**: v0.3.1 (Active Development - January 2025)

**Status**: Tool System Enhancement complete, ready for v0.3.1 release. All 32 tools migrated to ToolBuilder pattern with MCP compatibility.

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

## Current Focus (v0.3.x Roadmap)

1. **v0.3.1**: Tag release
2. **v0.3.2**: Enhanced Tool Capabilities (API Client, Auth)
3. **v0.3.3**: Built-in Agents Library
4. **v0.3.4**: Built-in Workflow Patterns
5. **v0.3.5**: Advanced Agent Features
6. **v0.3.6**: Model Context Protocol Support
7. **v0.3.7**: Performance Optimization
8. **v0.3.8**: Final Polish for v0.4.0

See TODO.md for detailed task tracking and TODO-DONE.md for completed items.