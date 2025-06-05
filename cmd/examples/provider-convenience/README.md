# Provider-Level Convenience Functions Example

This example demonstrates the provider-level convenience functions available in the `llmutil` package. These functions simplify common provider operations without requiring the agent framework.

## Overview

The example showcases provider-level utilities for:
- Creating providers from environment variables
- Batch generation across multiple prompts
- Generation with automatic retry
- Provider pooling with round-robin strategy
- Typed generation with schema validation
- Custom provider configuration

## Key Functions Demonstrated

### 1. ProviderFromEnv()
Automatically creates a provider using environment variables:
- Detects available API keys (OpenAI, Anthropic, Gemini)
- Selects appropriate model defaults
- Returns ready-to-use provider instance

### 2. BatchGenerate()
Processes multiple prompts concurrently:
- Efficient parallel execution
- Individual error handling per prompt
- Useful for bulk operations

### 3. GenerateWithRetry()
Adds resilience with automatic retry logic:
- Configurable max retry attempts
- Handles transient failures
- Uses exponential backoff

### 4. NewProviderPool()
Creates a pool of providers with load balancing:
- Round-robin strategy for distribution
- Useful for rate limit management
- Can mix different providers/models

### 5. ProcessTypedWithProvider()
Generates structured output with schema validation:
- Direct provider-level typed generation
- Schema-based validation
- Type-safe results

### 6. CreateProvider()
Creates providers with custom configuration:
- ModelConfig for detailed setup
- Override default settings
- Add provider-specific options

## Running the Example

```bash
# Set up environment variables for at least one provider
export OPENAI_API_KEY=your_openai_key
# OR
export ANTHROPIC_API_KEY=your_anthropic_key
# OR
export GEMINI_API_KEY=your_gemini_key

# Run the example
go run main.go
```

## When to Use Provider-Level Functions

Use these convenience functions when you:
- Need direct provider access without agent overhead
- Want simple request/response patterns
- Are building your own abstraction layer
- Need fine-grained control over provider behavior

For more complex scenarios involving:
- State management
- Tool calling
- Workflow orchestration
- Event handling

Consider using the agent framework instead (see agent examples).

## Related Examples

- `provider-multi` - Multi-provider strategies at provider level
- `provider-consensus` - Consensus strategies across providers
- `agent-simple-llm` - Using agents for more complex scenarios