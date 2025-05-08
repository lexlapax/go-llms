# Simple Example

This example demonstrates the core features of the Go-LLMs library using a mock provider, making it easy to understand the basic functionality without requiring any API keys.

## Overview

The Simple example showcases:

1. Creating and using a mock LLM provider
2. Defining schemas for structured data validation
3. Generating structured outputs with schema validation
4. Processing raw LLM responses into structured data
5. Streaming responses
6. Enhancing prompts with schema information
7. Working with typed Go structs for structured data

## Features Demonstrated

- **Basic Text Generation** - Simple text generation with a prompt
- **Schema Definition** - Creating JSON Schema compatible schemas for validation
- **Structured Generation** - Generating structured data with schema validation
- **Response Processing** - Extracting and validating JSON data from raw text responses
- **Response Streaming** - Streaming tokens as they're generated
- **Prompt Enhancement** - Enriching prompts with schema information
- **Typed Data Handling** - Working with Go structs for structured data

## Running the Example

To run the example:

```bash
make example EXAMPLE=simple
./bin/simple
```

Note: This example uses a mock provider, so no API keys are required.

## Structured Data Example

The example demonstrates structured data validation using a Person schema:

```go
// Person defines a struct for our schema
type Person struct {
    Name  string `json:"name" validate:"required" description:"Person's name"`
    Age   int    `json:"age" validate:"min=0,max=120" description:"Age in years"`
    Email string `json:"email" validate:"required,email" description:"Email address"`
}
```

The schema includes:
- Required fields (name, email)
- Integer validation with min/max values
- Format validation for email

## Key Components

- **MockProvider** - Simulates LLM provider responses
- **Schema** - Defines the structure and validation rules for data
- **Validator** - Validates structured outputs against schemas
- **StructuredProcessor** - Processes raw responses into structured data
- **PromptEnhancer** - Enriches prompts with schema information

## Learning Path

This example serves as an introduction to the library. After understanding this example, you can explore:

1. [Anthropic Example](../anthropic) - Working with a real LLM provider
2. [Multi Example](../multi) - Using multiple providers together
3. [Agent Example](../agent) - Building an agent with tools for interaction
4. Other specialized examples in the examples directory