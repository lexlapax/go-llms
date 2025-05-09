# Gemini Example

This example demonstrates how to use the Gemini provider with Go-LLMs to interact with Google's Gemini models.

## Features

- Text generation with prompt
- Message-based conversation
- Structured data generation with schema validation
- Streaming responses

## Prerequisites

- Go 1.21 or later
- A Google API key for Gemini

## Running the Example

1. Set your Gemini API key as an environment variable:

```bash
export GEMINI_API_KEY=your_api_key_here
```

2. Build and run the example:

```bash
make example EXAMPLE=gemini
./bin/gemini
```

## What This Example Demonstrates

1. **Simple Text Generation**: Generate text with a single prompt
2. **Conversation**: Create a conversation with multiple messages 
3. **Structured Output**: Generate structured data with schema validation
4. **Streaming**: Stream tokens as they're generated

## Output

The example produces outputs from the Gemini model for different types of prompts and shows how to use various features of the Go-LLMs library.

## Notes

- By default, this example uses the "gemini-2.0-flash-lite" model
- Streaming generates incremental results that are displayed as they arrive
- Structured output uses a Recipe schema to demonstrate validation