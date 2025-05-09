# OpenAI Provider Example

This example demonstrates how to use the OpenAI provider with the Go-LLMs library to generate text, hold conversations, and create structured outputs.

## Overview

The OpenAI example showcases:

1. Creating and configuring the OpenAI provider with the gpt-4o model
2. Using the OpenAIOrganizationOption for organization-specific contexts
3. Simple text generation with direct prompts
4. Message-based conversation with system and user roles
5. Structured output generation with schema validation
6. Response streaming
7. Graceful fallback to a mock provider when no API key is available

## Features Demonstrated

- **Simple Text Generation** - Basic text generation with a prompt
- **Provider Options** - Using the OpenAIOrganizationOption for organization-specific configurations
- **Conversation** - Using message-based conversation with system and user roles
- **Structured Output** - Generating structured recipe data with schema validation
- **Prompt Enhancement** - Enriching prompts with schema information for better results
- **Response Processing** - Processing raw LLM responses into validated structured data
- **Response Streaming** - Streaming tokens as they're generated

## Running the Example

To run the example:

```bash
# With OpenAI API key
export OPENAI_API_KEY=your_api_key_here
# Optional organization ID
export OPENAI_ORGANIZATION=your_organization_id_here
make example EXAMPLE=openai
./bin/openai

# Without API key (uses mock provider)
make example EXAMPLE=openai
./bin/openai
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

- **OpenAIProvider** - Handles API communication with OpenAI
- **OpenAIOrganizationOption** - Provider-specific option for organization settings
- **StructuredProcessor** - Processes raw responses into structured data
- **PromptEnhancer** - Enriches prompts with schema information
- **Validator** - Validates structured outputs against schemas

## Mock Provider Fallback

When no API key is provided, the example automatically falls back to a mock provider that simulates OpenAI's responses. This is useful for testing and development without API costs.