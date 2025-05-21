# Go-LLMs Examples

This directory contains example applications demonstrating various features and capabilities of the Go-LLMs library. Each example is self-contained and includes its own documentation.

## Quick Start

All examples can be run directly with Go:

```bash
# Navigate to any example directory
cd cmd/examples/<example-name>

# Run the example
go run main.go [args...]

# Or build and run the binary
go build -o example .
./example [args...]
```

## Available Examples

### Basic Usage

#### [**Simple**](simple/README.md)
Basic usage patterns with mock providers. Perfect for getting started without API keys.

**Features:**
- Mock provider setup
- Basic text generation
- Error handling examples

```bash
cd simple && go run main.go
```

#### [**Convenience**](convenience/README.md) 
Utility functions and helper patterns for common LLM operations.

**Features:**
- Batch processing
- Retry mechanisms
- Provider pools
- Configuration helpers

```bash
cd convenience && go run main.go
```

### Provider Integration

#### [**OpenAI**](openai/README.md)
Integration with OpenAI's GPT models including GPT-4o and GPT-4 Turbo.

**Features:**
- Text generation
- Structured output
- Streaming responses
- Organization configuration

```bash
export OPENAI_API_KEY="your-key"
cd openai && go run main.go
```

#### [**Anthropic**](anthropic/README.md)
Integration with Anthropic's Claude models including Claude 3.5 Sonnet.

**Features:**
- Conversation handling
- System prompts
- Claude-specific optimizations

```bash
export ANTHROPIC_API_KEY="your-key"
cd anthropic && go run main.go
```

#### [**Gemini**](gemini/README.md)
Integration with Google's Gemini models including Gemini 2.0 Flash.

**Features:**
- Google AI integration
- Multimodal capabilities
- Safety settings

```bash
export GEMINI_API_KEY="your-key"
cd gemini && go run main.go
```

#### [**OpenAI API Compatible Providers**](openai_api_compatible_providers/README.md)
Working with OpenAI-compatible APIs like OpenRouter and Ollama.

**Features:**
- OpenRouter integration
- Ollama local models
- Custom API endpoints

```bash
export OPENROUTER_API_KEY="your-key"
cd openai_api_compatible_providers && go run main.go
```

### Advanced Features

#### [**Multi-Provider**](multi/README.md)
Using multiple LLM providers simultaneously with different strategies.

**Features:**
- Fastest strategy (first to respond)
- Primary with fallback
- Load balancing

```bash
export OPENAI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
cd multi && go run main.go
```

#### [**Consensus**](consensus/README.md)
Advanced multi-provider consensus strategies for improved reliability.

**Features:**
- Similarity-based consensus
- Voting mechanisms
- Quality scoring

```bash
export OPENAI_API_KEY="your-key" 
export ANTHROPIC_API_KEY="your-key"
export GEMINI_API_KEY="your-key"
cd consensus && go run main.go
```

#### [**Agent**](agent/README.md)
Agent workflows with tool integration for complex tasks.

**Features:**
- Tool calling
- Workflow orchestration
- Message history
- Monitoring hooks

```bash
export OPENAI_API_KEY="your-key"
cd agent && go run main.go
```

### Content and Media

#### [**Multimodal**](multimodal/README.md)
Working with text, images, audio, video, and file content.

**Features:**
- Image processing
- File uploads
- Audio/video handling
- URL-based content

```bash
export OPENAI_API_KEY="your-key"
cd multimodal && go run main.go --image image.jpg --text "Describe this image"
```

### Configuration and Options

#### [**Provider Options**](provider_options/README.md)
Demonstration of the provider option system for configuration.

**Features:**
- Common options (HTTP client, timeouts)
- Provider-specific options
- Environment variable configuration

```bash
export OPENAI_API_KEY="your-key"
cd provider_options && go run main.go
```

### Data and Validation

#### [**Schema**](schema/README.md)
Schema generation from Go structs and validation patterns.

**Features:**
- Automatic schema generation
- Struct validation
- Custom validation rules

```bash
cd schema && go run main.go
```

#### [**Coercion**](coercion/README.md)
Type coercion and data conversion for validation.

**Features:**
- String to number conversion
- Array handling
- Flexible validation

```bash
cd coercion && go run main.go
```

### Model Discovery and Information

#### [**Model Info**](modelinfo/README.md) 🆕
**Discover and explore available models from all LLM providers with capability filtering and caching.**

**Features:**
- **Automatic model discovery** from OpenAI, Anthropic, and Google Gemini
- **Capability filtering** (multimodal, function calling, streaming, etc.)
- **Intelligent caching** to reduce API calls and improve performance
- **Detailed model information** including context windows, token limits, and pricing
- **Provider comparison** and model recommendation
- **CLI interface** with flexible filtering options

**Quick Start:**
```bash
# Set API keys for the providers you want to query
export OPENAI_API_KEY="your-openai-key"
export ANTHROPIC_API_KEY="your-anthropic-key"  # Optional - has fallback data
export GEMINI_API_KEY="your-gemini-key"

cd modelinfo

# Get all available models
go run main.go

# Filter by provider
go run main.go --provider=openai

# Find models that support images
go run main.go --capability=image-input

# Get models with large context windows
go run main.go --name="gpt-4" --pretty

# Force fresh data (ignore cache)
go run main.go --fresh --metadata
```

**Use Cases:**
- **Model selection** - Find the best model for your specific needs
- **Capability assessment** - Determine which models support required features
- **Cost optimization** - Compare pricing and context window limits
- **API exploration** - Discover new models as they become available

### Monitoring and Performance

#### [**Metrics**](metrics/README.md)
Performance monitoring and metrics collection.

**Features:**
- Response time tracking
- Token usage monitoring
- Error rate analysis
- Custom metrics

```bash
export OPENAI_API_KEY="your-key"
cd metrics && go run main.go
```

#### [**Profiling**](profiling/README.md)
Performance profiling and optimization techniques.

**Features:**
- CPU profiling
- Memory analysis
- Benchmark comparisons

```bash
cd profiling && go run main.go
```

## Example Categories

### By Complexity
- **Beginner**: `simple`, `schema`, `coercion`
- **Intermediate**: `openai`, `anthropic`, `gemini`, `provider_options`, `multimodal`, `modelinfo`
- **Advanced**: `multi`, `consensus`, `agent`, `metrics`, `profiling`

### By Use Case
- **Text Generation**: `simple`, `openai`, `anthropic`, `gemini`
- **Multimodal**: `multimodal`, `gemini`
- **Configuration**: `provider_options`, `convenience`
- **Data Validation**: `schema`, `coercion`
- **Multiple Providers**: `multi`, `consensus`
- **Tools & Agents**: `agent`
- **Model Discovery**: `modelinfo`
- **Monitoring**: `metrics`, `profiling`

### By Provider
- **OpenAI**: `openai`, `multimodal`, `agent`
- **Anthropic**: `anthropic`
- **Google Gemini**: `gemini`
- **Multiple**: `multi`, `consensus`, `modelinfo`
- **Compatible APIs**: `openai_api_compatible_providers`

## Running Tests

Each example includes tests that can be run with:

```bash
cd <example-directory>
go test ./...
```

Some integration tests require API keys to be set.

## Common Environment Variables

Most examples use these environment variables:

```bash
# LLM Provider API Keys
export OPENAI_API_KEY="your-openai-api-key"
export ANTHROPIC_API_KEY="your-anthropic-api-key"
export GEMINI_API_KEY="your-gemini-api-key"
export OPENROUTER_API_KEY="your-openrouter-api-key"

# Optional Configuration
export OPENAI_ORGANIZATION="your-org-id"
export OPENAI_BASE_URL="https://api.openai.com"  # Custom endpoint
export LLM_HTTP_TIMEOUT="30"                     # Timeout in seconds
export LLM_RETRY_ATTEMPTS="3"                    # Retry attempts
```

## Building All Examples

You can build all examples at once using the provided Makefile from the project root:

```bash
# Build all examples
make build-examples

# Build a specific example
make build-example EXAMPLE=modelinfo

# Clean all built examples
make clean-examples
```

## Getting Help

- Check individual example READMEs for detailed usage instructions
- Review the [User Guides](/docs/user-guide/) for comprehensive documentation
- See the [API Documentation](/docs/api/) for technical reference
- Check [Troubleshooting](/docs/user-guide/error-handling.md) for common issues

## Contributing

When adding new examples:

1. Create a new directory under `cmd/examples/`
2. Include `main.go`, `main_test.go`, and `README.md`
3. Follow the existing patterns and documentation style
4. Add the example to this README.md
5. Update the main project README.md examples section