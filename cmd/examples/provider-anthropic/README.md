# Anthropic Provider Example

This example demonstrates how to use the Anthropic Claude provider with the Go-LLMs library to generate text, hold conversations, and create structured outputs.

## Overview

The Anthropic example showcases:

1. Creating and configuring the Anthropic provider with the claude-3-5-sonnet-latest model
2. Using the AnthropicSystemPromptOption for consistent system behavior
3. Simple text generation with direct prompts
4. Message-based conversation leveraging the system prompt option
5. Structured output generation with schema validation
6. Response streaming
7. Graceful fallback to a mock provider when no API key is available

## Features Demonstrated

- **Simple Text Generation** - Basic text generation with a prompt
- **Provider Options** - Using the AnthropicSystemPromptOption for consistent system instructions
- **Conversation** - Using message-based conversation with user roles and system prompt option
- **Structured Output** - Generating structured recipe data with schema validation
- **Prompt Enhancement** - Enriching prompts with schema information for better results
- **Response Processing** - Processing raw LLM responses into validated structured data
- **Response Streaming** - Streaming tokens as they're generated

## Running the Example

To run the example:

```bash
# With Anthropic API key
export ANTHROPIC_API_KEY=your_api_key_here
make example EXAMPLE=anthropic
./bin/anthropic

# Without API key (uses mock provider)
make example EXAMPLE=anthropic
./bin/anthropic
```

## Structured Data Example

The example demonstrates structured data generation with a recipe schema:

```go
// Recipe represents a cooking recipe
type Recipe struct {
    Title       string   `json:"title"`
    Ingredients []string `json:"ingredients"`
    Steps       []string `json:"steps"`
    PrepTime    int      `json:"prepTime"`
    CookTime    int      `json:"cookTime"`
    Servings    int      `json:"servings"`
    Difficulty  string   `json:"difficulty"`
}
```

The schema includes validation rules:
- Required fields (title, ingredients, steps, cookTime, servings)
- Integer validation with minimum values
- String enumeration for difficulty (easy, medium, hard)

## Key Components

- **AnthropicProvider** - Handles API communication with Anthropic
- **AnthropicSystemPromptOption** - Provider-specific option for setting a persistent system prompt
- **StructuredProcessor** - Processes raw responses into structured data
- **PromptEnhancer** - Enriches prompts with schema information
- **Validator** - Validates structured outputs against schemas

## Mock Provider Fallback

When no API key is provided, the example automatically falls back to a mock provider that simulates Anthropic's responses. This is useful for testing and development without API costs.