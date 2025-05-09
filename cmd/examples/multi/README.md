# Multi-Provider Example

This example demonstrates how to use the Multi-Provider functionality of the Go-LLMs library to work with multiple LLM providers simultaneously using different strategies.

## Overview

The Multi-Provider example showcases:

1. Creating and configuring multiple providers (OpenAI, Anthropic, Gemini, and/or mock providers)
2. Using provider-specific options for each provider:
   - OpenAI with OpenAIOrganizationOption
   - Anthropic with AnthropicSystemPromptOption
   - Gemini with GeminiGenerationConfigOption
3. Using different multi-provider strategies:
   - Fastest Strategy - Returns the first response from any provider
   - Primary Strategy - Tries the primary provider first, falls back to others if it fails
4. Working with real API-based providers when credentials are available
5. Simulating provider behavior with mock providers when no credentials are present
6. Streaming responses through the multi-provider system

## Features Demonstrated

- **Fastest Strategy** - Racing multiple providers to get the quickest response
- **Primary Strategy** - Using a preferred provider with fallbacks
- **Provider Weighting** - Configuring different weights for providers
- **Provider-Specific Options** - Using OpenAIOrganizationOption, AnthropicSystemPromptOption, and GeminiGenerationConfigOption
- **API Integration** - Working with real LLM providers (OpenAI, Anthropic, Gemini)
- **Response Timing** - Measuring and comparing response times
- **Mock Providers** - Simulating different response times and behaviors
- **Streaming** - Streaming tokens through the multi-provider system

## Running the Example

To run the example:

```bash
# With OpenAI, Anthropic, and Gemini API keys
export OPENAI_API_KEY=your_openai_key_here
export ANTHROPIC_API_KEY=your_anthropic_key_here
export GEMINI_API_KEY=your_gemini_key_here

# Optional: Set organization ID for OpenAI
export OPENAI_ORGANIZATION=your_organization_id

make example EXAMPLE=multi
./bin/multi

# With just one API key (will add a mock provider as the second provider)
export OPENAI_API_KEY=your_openai_key_here
make example EXAMPLE=multi
./bin/multi

# Without API keys (uses simulated mock providers)
make example EXAMPLE=multi
./bin/multi
```

## Simulated Provider Configuration

When running without API keys, the example creates three mock providers with different performance characteristics:

- **Fast Provider** - 100ms delay, simulating a fast but potentially less capable model
- **Medium Provider** - 300ms delay, simulating a balanced model
- **Slow Provider** - 500ms delay, simulating a slower but potentially more capable model

## Key Components

- **MultiProvider** - Coordinates multiple LLM providers using different strategies
- **ProviderWeight** - Configures the weight and name for each provider
- **ResponseStream** - Handles streaming responses from providers
- **Provider-Specific Options**:
  - **OpenAIOrganizationOption** - Sets the organization ID for OpenAI
  - **AnthropicSystemPromptOption** - Sets system prompt for Anthropic
  - **GeminiGenerationConfigOption** - Configures generation parameters for Gemini

## Real-World Applications

Multi-provider configurations are useful for:

1. **Reliability** - Ensuring continued operation even if one provider has issues
2. **Performance** - Getting the fastest possible response
3. **Cost Optimization** - Using cheaper providers first, falling back to more expensive ones
4. **Quality Control** - Comparing responses from different providers
5. **A/B Testing** - Testing different models and providers