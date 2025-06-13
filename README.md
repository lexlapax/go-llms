# Go-LLMs: Unified Go Library for LLM Integration

A lightweight Go library providing a simplified, unified interface to interact with various LLM providers while offering robust data validation and agent tooling and multi-agent orchestration via workflows and state management.

## Features

- **Unified API** across OpenAI, Anthropic, Google Gemini, Vertex AI, Ollama, OpenRouter, and compatible providers
- **Structured outputs** with JSON schema validation and type coercion
- **Agent system** with state management, hooks, and workflow patterns
- **32 built-in tools** for web, file, system, data, datetime, and feed operations
- **Tool enhancement** with LLM guidance metadata and MCP (Model Context Protocol) support
- **Tool discovery system** (v0.3.4+) with metadata-first exploration and scripting bridge integration
- **Multimodal content** support for text, images, files, videos, and audio
- **Multi-provider strategies** including fastest, primary, and consensus approaches
- **Type-safe configuration** with interface-based provider options
- **Minimal dependencies** leveraging Go's standard library

## What's New in v0.3.4

See [CHANGELOG.md](CHANGELOG.md) for the complete version history.

### v0.3.4 (June 13, 2025) - Tool Discovery System 🔍
- **Metadata-First Discovery**: Explore 33+ tools without imports - perfect for scripting engines
- **Lazy Loading**: Factory pattern with on-demand tool instantiation using build tags
- **Bridge Integration**: Designed for go-llmspell Lua/JavaScript bridges 
- **Rich Metadata Access**: Get schemas, examples, help text without tool instances
- **Performance**: Excellent benchmarks with thread-safe concurrent access
- **Complete Documentation**: Enhanced examples and technical guides

### v0.3.3 (January 11, 2025) - Major Provider Expansion
- **Ollama**: Local model hosting with GPU acceleration and model management
- **OpenRouter**: Access to 400+ models (68 free) from multiple providers via unified API
- **Vertex AI**: Enterprise Google Cloud deployment with Gemini and partner models (Claude)
- Full integration with utilities, CLI, and configuration systems

### v0.3.2 (January 11, 2025) - Documentation Update
- Complete documentation restructuring for better user experience
- Modularized API documentation with dedicated files for each component
- Enhanced technical documentation with new guides for providers and tools

## Installation

```bash
go get github.com/lexlapax/go-llms
```

## Quick Start

### Basic Usage

```go
// Create a provider
provider := provider.NewOpenAIProvider(
    os.Getenv("OPENAI_API_KEY"),
    "gpt-4o",
)

// Generate text
response, err := provider.Generate(context.Background(), "Explain quantum computing")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
```

### Using Agents with Tools

```go
// Create an agent with built-in tools
agent, err := core.NewAgentFromString("assistant", "openai/gpt-4o")
if err != nil {
    log.Fatal(err)
}

// Add built-in tools
agent.AddTool(web.WebSearch())
agent.AddTool(file.FileRead())

// Execute with state
state := domain.NewState()
state.Set("prompt", "Search for Go programming tutorials and save the results")
result, err := agent.Run(context.Background(), state)
```

### Tool Discovery (v0.3.4+)

```go
// Discover tools without imports - perfect for scripting engines
discovery := tools.NewDiscovery()

// List all 33+ available tools
for _, tool := range discovery.ListTools() {
    fmt.Printf("%s (%s): %s\n", tool.Name, tool.Category, tool.Description)
}

// Search tools by keyword, category, or tags
jsonTools := discovery.SearchTools("json")
mathTools := discovery.ListByCategory("math")

// Get rich metadata before creating tools
schema, _ := discovery.GetToolSchema("calculator")
examples, _ := discovery.GetToolExamples("calculator")

// Lazy tool instantiation only when needed
calculator, _ := discovery.CreateTool("calculator")
```

### Structured Output

```go
// Define a schema
schema := &domain.Schema{
    Type: "object",
    Properties: map[string]domain.Property{
        "name":  {Type: "string"},
        "age":   {Type: "integer"},
        "email": {Type: "string", Format: "email"},
    },
    Required: []string{"name", "email"},
}

// Generate structured data
result, err := provider.GenerateWithSchema(
    context.Background(),
    "Generate a person's information",
    schema,
)
```

## Documentation

- **[Complete Documentation](/docs/README.md)** - Full documentation index
- [Getting Started Guide](docs/user-guide/getting-started.md) - Quick start and basic concepts
- [User Guide](docs/user-guide/README.md) - Complete user documentation
- [Tools & Components](docs/user-guide/tools.md) - Built-in tools and components
- [Tool Discovery API](docs/technical/tool-discovery-api.md) - **NEW**: Metadata-first discovery system
- [Examples Gallery](docs/user-guide/examples-gallery.md) - Usage examples
- [API Reference](docs/api/README.md) - Complete API documentation
- [Technical Documentation](docs/technical/README.md) - Architecture and implementation details

## Supported Providers

- **OpenAI** - GPT-4o, GPT-4o-mini, GPT-4 Turbo, GPT-3.5 Turbo
- **Anthropic** - Claude 3.5 Sonnet, Claude 3 Opus, Claude 3 Haiku
- **Google Gemini** - Gemini 2.0 Flash Lite, Gemini Pro, Gemini Pro Vision
- **Google Vertex AI** - Enterprise Gemini models, Claude (partner models), regional deployment
- **Ollama** - Llama 3.2, Mistral, Phi-3, CodeLlama, and more (local hosting)
- **OpenRouter** - Access to 400+ models from various providers (68 free models)
- **OpenAI-Compatible** - LM Studio, vLLM, and any OpenAI-compatible API

## Examples

The `cmd/examples/` directory contains 40+ examples demonstrating various features:

- **Provider examples**: OpenAI, Anthropic, Gemini, OpenRouter, Ollama, multi-provider strategies
- **Agent examples**: Tool usage, workflows, state management, sub-agents
- **Built-in tools**: Web search, file operations, API client, data processing
- **Advanced patterns**: Structured output, multimodal content, custom agents

## Architecture

Go-LLMs follows a clean architecture with vertical feature slicing:

```
pkg/
├── schema/      # JSON schema validation
├── llm/         # Provider implementations
├── structured/  # Output processing
└── agent/       # Agent system with tools and workflows
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Special thanks to the LLM-based coding tools that helped with documentation and testing: Aider, Claude Code, ChatGPT, Claude Desktop, and Gemini Code.