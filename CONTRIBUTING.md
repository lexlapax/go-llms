# Contributing to Go-LLMs

Thank you for your interest in contributing to Go-LLMs! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

Please be respectful and considerate in all interactions. We aim to maintain a welcoming and inclusive environment for all contributors.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/go-llms.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes following the guidelines below
5. Run tests: `make test`
6. Submit a pull request

## Development Guidelines

### Code Style

- Run `make fmt` before committing to ensure consistent formatting
- Run `make vet` to catch common issues
- Run `make lint` if you have golangci-lint installed
- Follow existing patterns and conventions in the codebase

### Testing

All contributions must include appropriate tests:

- Unit tests for new functionality
- Integration tests for provider implementations
- Benchmark tests for performance-critical code
- Use table-driven tests for functions with multiple scenarios

Run tests with:
```bash
make test              # Unit tests only
make test-all          # All tests including integration
make test-pkg PKG=...  # Specific package tests
```

### Error Handling

- Return errors as the last return value
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Don't panic except in truly unrecoverable situations
- Provide meaningful error messages that help with debugging

### Logging Guidelines

The project follows specific logging patterns:

#### Library Code (pkg/)
- **DO NOT** add logging to library code
- Return errors with context instead
- Exception: The LoggingHook in pkg/agent/workflow/hooks.go

```go
// Good
return fmt.Errorf("failed to parse response: %w", err)

// Bad - don't log in library
log.Printf("Error: %v", err)
```

#### Example Programs
- Use `log` package for consistency
- Agent examples may use `slog` to demonstrate LoggingHook
- Don't mix `log` and `fmt` in the same example

#### CLI Tools
- Use `fmt` for output control
- `fmt.Printf/Println` for normal output  
- `fmt.Fprintf(os.Stderr, ...)` for errors

For complete logging guidelines, see [docs/technical/logging.md](docs/technical/logging.md).

### Documentation

- Add godoc comments to all exported types, functions, and methods
- Update relevant documentation when making changes
- Include examples in documentation where helpful
- Keep README files up to date

### Commit Messages

- Use clear, descriptive commit messages
- Start with a verb in present tense: "Add", "Fix", "Update", etc.
- Keep the first line under 72 characters
- Add detailed description after a blank line if needed

Example:
```
Add structured logging support to agent workflows

- Implement LoggingHook with configurable levels
- Add emoji decorations for better readability
- Support both default and custom slog instances
```

## Submitting Changes

### Pull Request Process

1. Ensure all tests pass: `make test-all`
2. Update documentation as needed
3. Add entries to relevant example programs if applicable
4. Submit PR with clear description of changes
5. Be responsive to review feedback

### PR Guidelines

- Keep PRs focused on a single feature or fix
- Include tests for new functionality
- Update CHANGELOG.md if applicable
- Reference any related issues

## Adding New Features

### New Provider Implementation

See [docs/technical/provider-implementation.md](docs/technical/provider-implementation.md) for detailed instructions on adding a new LLM provider.

### New Tools or Agents

When adding built-in tools or agents:

1. Follow existing patterns in pkg/agent/tools/
2. Include comprehensive tests
3. Add example usage in cmd/examples/
4. Update documentation

### Performance Optimizations

- Include benchmark tests to demonstrate improvements
- Use `make benchmark` to measure impact
- Consider memory allocations and GC pressure
- Document any trade-offs

## Questions?

If you have questions about contributing, feel free to:

- Open an issue for discussion
- Check existing documentation in /docs
- Review existing code for patterns and examples

Thank you for contributing to Go-LLMs!