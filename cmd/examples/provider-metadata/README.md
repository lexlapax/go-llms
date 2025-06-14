# Provider Metadata Example

This example demonstrates the provider metadata and dynamic registry system in go-llms v0.3.5.7.

## Features Demonstrated

### 1. Provider Templates
- Explore available provider templates (OpenAI, Anthropic, Mock)
- View required and optional configuration fields
- Check environment variable support

### 2. Dynamic Provider Creation
- Create providers from templates with configuration
- Register custom providers with metadata
- Runtime provider instantiation

### 3. Capability-Based Discovery
- Find providers by specific capabilities (streaming, vision, function calling)
- Filter providers based on feature requirements
- Capability-aware provider selection

### 4. Model Comparison
- Compare models across different providers
- View model pricing, context windows, and capabilities
- Identify deprecated models

### 5. Configuration Management
- Export provider configurations
- Import configurations to new registries
- Configuration persistence and portability

### 6. Best Model Selection
- Find optimal models based on constraints:
  - Minimum context window requirements
  - Maximum price limits
  - Required capabilities
- Automatic provider configuration with selected model

## Running the Example

```bash
cd cmd/examples/provider-metadata
go run main.go
```

## Example Output

```
=== Provider Metadata Example ===

1. Available Provider Templates:

  Template: openai
  Type: openai
  Description: OpenAI GPT models
  Required fields:
    - api_key (string): OpenAI API key
      Can use env var: OPENAI_API_KEY

2. Dynamic Provider Creation:
  Created provider 'dynamic-mock'
  Test response: Hello from dynamic mock!
  
  Registered provider 'custom-mock' with metadata

3. Capability-Based Discovery:
  Providers with streaming: [stream-only, vision-capable, full-featured]
  Providers with vision: [vision-capable, full-featured]
  Providers with function calling: [full-featured]

4. Model Comparison:
  Provider Comparison:
  ┌──────────────┬─────────┬──────────┬───────────┬───────────┐
  │ Provider     │ Models  │ Streaming│ Vision    │ Functions │
  ├──────────────┼─────────┼──────────┼───────────┼───────────┤
  │ openai-test  │ 20      │ true     │ true      │ true      │
  │ anthropic-test│ 6       │ true     │ true      │ false     │
  └──────────────┴─────────┴──────────┴───────────┴───────────┘

5. Configuration Management:
  Exported configuration saved
  Successfully imported configuration into new registry

6. Best Model Selection:
  Best model for requirements:
    Min context: 100000 tokens
    Max price: $5.00 per million tokens
    Required: Streaming, Vision

  Selected: GPT-4 Vision (gpt-4-vision-preview)
    Context: 128000 tokens
    Price: $10.00 per 1000000 tokens
```

## Key Concepts

### Provider Metadata
Each provider can expose metadata including:
- Provider name and description
- Supported capabilities
- Available models with pricing
- Configuration schema
- Rate limits and constraints

### Dynamic Registry
The registry supports:
- Runtime provider registration
- Factory-based provider creation
- Capability-based queries
- Configuration import/export
- Event listeners for registry changes

### Provider Factories
Factories enable:
- Template-based provider creation
- Configuration validation
- Dynamic provider instantiation
- Scripting engine integration

## Use Cases

1. **UI/Tool Integration**: Use templates and schemas to build configuration UIs
2. **Cost Optimization**: Select models based on pricing constraints
3. **Capability Matching**: Find providers that support required features
4. **Multi-Provider Apps**: Dynamically switch between providers
5. **Scripting Engines**: Runtime provider discovery and creation

## Integration with go-llmspell

This metadata system is designed to support go-llmspell's requirements:
- Metadata-first exploration without imports
- Runtime provider registration
- Bridge-friendly serialization
- Dynamic capability discovery