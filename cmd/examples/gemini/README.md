# Gemini Example

This example demonstrates how to use the Gemini provider with Go-LLMs to interact with Google's Gemini models.

## Features

- Text generation with prompt
- Provider-specific options (GeminiGenerationConfigOption and GeminiSafetySettingsOption)
- Message-based conversation
- Structured data generation with schema validation
- Streaming responses
- Generation parameter comparison (temperature, top-k, top-p)

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
2. **Provider Options**: Configure Gemini with GeminiGenerationConfigOption and GeminiSafetySettingsOption
3. **Conversation**: Create a conversation with multiple messages
4. **Structured Output**: Generate structured data with schema validation
5. **Streaming**: Stream tokens as they're generated
6. **Generation Parameter Comparison**: Compare different generation settings (temperature, top-k, top-p)

## Output

The example produces outputs from the Gemini model for different types of prompts and shows how to use various features of the Go-LLMs library.

## Notes

- By default, this example uses the "gemini-2.0-flash-lite" model
- The example demonstrates using provider-specific options:
  - GeminiGenerationConfigOption for controlling temperature, top-k, top-p, max output tokens
  - GeminiSafetySettingsOption for configuring content filtering settings
- The generation parameter comparison shows how different settings affect the outputs
- Streaming generates incremental results that are displayed as they arrive
- Structured output uses a Recipe schema to demonstrate validation

## Generation Parameters

The example demonstrates various generation parameters:

- **Temperature**: Controls randomness in token selection
  - Higher values (0.7-1.0): More diverse and creative outputs
  - Lower values (0.1-0.3): More focused and deterministic outputs

- **Top-K Sampling**: Limits token selection to the K most likely next tokens
  - Higher values: Broader vocabulary selection
  - Lower values: More conservative output

- **Top-P (Nucleus) Sampling**: Considers only tokens whose cumulative probability exceeds P
  - Works alongside temperature to balance creativity and coherence

- **Safety Settings**: Control content filtering for harmful categories
  - Configurable thresholds for harassment, hate speech, etc.