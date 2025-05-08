# Go-LLMs API Documentation

This documentation provides detailed information about the Go-LLMs public API, organized by packages and features.

## Packages

- [schema](schema.md) - Schema definition and validation
- [llm](llm.md) - LLM provider integration
- [structured](structured.md) - Structured output processing
- [agent](agent.md) - Agent and tool functionality

## Getting Started

For most applications, you'll need to use multiple packages together. Here's a typical usage flow:

1. Define a schema for the structured output you want to receive
2. Create an LLM provider (e.g., OpenAI, Anthropic)
3. Use the structured output processor to generate and validate responses
4. Optionally, use the agent system to add tools and workflow capabilities

Check each package's documentation for detailed examples and guides.
