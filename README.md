# Go-LLMs: A Go Implementation of Pydantic-AI

Go-LLMs is a Go library for creating LLM-powered applications with structured outputs and type safety. It aims to port the core functionality of pydantic-ai to Go while embracing Go's idioms and strengths.

## Features

- **Structured responses**: Validates LLM outputs against predefined schemas
- **Model-agnostic interface**: Provides a unified API across different LLM providers
- **Type safety**: Leverages Go's type system for better developer experience
- **Dependency injection**: Enables passing data and services into agents
- **Tool integration**: Allows LLMs to interact with external systems through function calls

## Project Goals

1. Create an idiomatic Go implementation of pydantic-ai
2. Minimize external dependencies by leveraging Go's standard library
3. Support modern LLM providers (OpenAI, Anthropic, etc.)
4. Provide comprehensive validation for LLM outputs
5. Follow clean architecture principles with vertical feature slices

## Project Structure

The project follows a vertical slicing approach where code is organized by feature:

```
go-llms/
├── cmd/                       # Application entry points
│   └── examples/              # Example applications
├── internal/                  # Internal packages
├── pkg/                       # Public packages
│   ├── schema/                # Schema definition and validation
│   ├── llm/                   # LLM integration
│   ├── structured/            # Structured output
│   └── agent/                 # Agent feature
└── examples/                  # Usage examples
```

## Installation

```bash
go get github.com/lexlapax/go-llms
```

## Basic Usage

```go
// Example code will be provided as the library develops
```

## Development Status

This project is currently in active development. The API is unstable and subject to change.

## License

This project is licensed under the MIT License - see the LICENSE file for details.