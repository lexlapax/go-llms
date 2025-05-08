# Convenience Utilities Example

This example demonstrates the use of Go-LLMs' convenience utilities for common LLM operations.

## Overview

The Go-LLMs library includes utility functions that simplify common operations when working with LLMs, structured outputs, and agents. These utilities help reduce boilerplate code and provide a more streamlined development experience.

## Key Concepts

### Provider Management

- **CreateProvider**: Easily create LLM providers from configuration
- **ProviderPool**: Pool multiple providers for load balancing and failover
- **BatchGenerate**: Generate responses for multiple prompts in parallel
- **GenerateWithRetry**: Automatically retry operations on transient errors

### Structured Output

- **GenerateTyped**: Generate and parse typed responses using Go structs
- **EnhancePromptWithExamples**: Add examples to schema-based prompts
- **SanitizeStructuredOutput**: Validate structured output against a schema

### Agent Workflow

- **CreateAgent**: Easily create agents with common configuration patterns
- **CreateStandardTools**: Create standard tools for common operations
- **AgentWithMetrics**: Create an agent with built-in metrics collection
- **RunWithTimeout**: Run an agent with a timeout

## Example Usage

The example demonstrates several key features:

1. **Provider Creation**: Creating providers from configuration
2. **Batch Generation**: Generating multiple responses in parallel
3. **Retry Logic**: Handling transient errors automatically
4. **Provider Pool**: Load balancing across multiple providers
5. **Typed Generation**: Generating and parsing typed responses
6. **Agent Creation**: Creating and running agents with tools

## Running the Example

To run this example:

```bash
# Build the example
make example EXAMPLE=convenience

# Run with an API key (optional)
export OPENAI_API_KEY=your_api_key_here
# OR
export ANTHROPIC_API_KEY=your_api_key_here

./bin/convenience
```

If no API keys are provided, the example will fall back to using a mock provider.

## Key Files

- **agent.go**: Convenience functions for agent workflows
- **llmutil.go**: Core utilities for LLM provider operations
- **pool.go**: Provider pooling and load balancing
- **structured.go**: Utilities for structured output generation

## Best Practices

- Use pooling for high-throughput applications
- Implement retry logic for handling transient errors
- Use typed generation for structured outputs
- Consider timeout constraints for agent operations
- Add metrics collection for performance monitoring