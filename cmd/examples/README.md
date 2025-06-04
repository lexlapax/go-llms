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

### Built-in Components

#### [**Built-ins Discovery**](builtins-discovery/)
Introduction to the built-in tools registry system.

**Features:**
- Tool discovery and search
- Registry categories and tags
- Basic agent integration
- Migration from custom tools

```bash
cd builtins-discovery && go run main.go
```

#### [**Built-ins File Tools**](builtins-file-tools/)
Deep dive into enhanced file operation tools.

**Features:**
- Streaming large files
- Atomic writes with backups
- Binary file detection
- Line range reading
- File metadata

```bash
cd builtins-file-tools && go run main.go
```

#### [**Built-ins Web Tools**](builtins-web-tools/)
Comprehensive web interaction tools.

**Features:**
- Web page fetching with smart parsing
- DuckDuckGo search integration
- Advanced web scraping
- Custom HTTP requests

```bash
cd builtins-web-tools && go run main.go
```

#### [**Built-ins System Tools**](builtins-system-tools/)
System information and command execution tools.

**Features:**
- Safe command execution
- Environment variable access
- System information gathering
- Process listing

```bash
cd builtins-system-tools && go run main.go
```

#### [**Built-ins Data Tools**](builtins-data-tools/)
Powerful data processing and transformation tools.

**Features:**
- JSON processing with JSONPath
- CSV parsing and transformation
- XML to JSON conversion
- Generic data transformations

```bash
cd builtins-data-tools && go run main.go
```

#### [**Built-ins DateTime Tools**](builtins-datetime-tools/)
Comprehensive date and time manipulation tools.

**Features:**
- Current time in any timezone
- Date arithmetic and comparisons
- Flexible parsing and formatting
- Business day calculations

```bash
cd builtins-datetime-tools && go run main.go
```

#### [**Built-ins Feed Tools**](builtins-feed-tools/)
RSS, Atom, and JSON Feed processing tools.

**Features:**
- Fetch and parse feeds (RSS 2.0, Atom 1.0, JSON Feed)
- Auto-discover feeds from websites
- Filter items by keywords, dates, authors
- Aggregate multiple feeds
- Convert between feed formats
- Extract specific data from feeds

```bash
cd builtins-feed-tools && go run main.go
```

### Provider Integration

#### [**Provider OpenAI**](provider-openai/README.md)
Integration with OpenAI's GPT models including GPT-4o and GPT-4 Turbo.

**Features:**
- Text generation
- Structured output
- Streaming responses
- Organization configuration

```bash
export OPENAI_API_KEY="your-key"
cd provider-openai && go run main.go
```

#### [**Provider Anthropic**](provider-anthropic/README.md)
Integration with Anthropic's Claude models including Claude 3.5 Sonnet.

**Features:**
- Conversation handling
- System prompts
- Claude-specific optimizations

```bash
export ANTHROPIC_API_KEY="your-key"
cd provider-anthropic && go run main.go
```

#### [**Provider Gemini**](provider-gemini/README.md)
Integration with Google's Gemini models including Gemini 2.0 Flash.

**Features:**
- Google AI integration
- Multimodal capabilities
- Safety settings

```bash
export GEMINI_API_KEY="your-key"
cd provider-gemini && go run main.go
```

#### [**Provider OpenAI Compatible**](provider-openai-compatible/README.md)
Working with OpenAI-compatible APIs like OpenRouter and Ollama.

**Features:**
- OpenRouter integration
- Ollama local models
- Custom API endpoints

```bash
export OPENROUTER_API_KEY="your-key"
cd provider-openai-compatible && go run main.go
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

#### [**Agent Structured Output**](agent-structured-output/README.md)
LLM agents with structured output validation using schemas.

**Features:**
- Schema-driven LLM interactions
- Type-safe processing
- Complex data structures
- Real-world use cases (tasks, meetings, analysis)

```bash
export OPENAI_API_KEY="your-key"
cd agent-structured-output && go run main.go
```

#### [**Agent Simple LLM**](agent-simple-llm/README.md)
Ultra-simple agent creation with string-based provider specification.

**Features:**
- Minimal agent setup
- Provider/model aliases
- String-based configuration
- State-based interface

```bash
export OPENAI_API_KEY="your-key"
cd agent-simple-llm && go run main.go
```

### Custom Agents *(NEW - February 3, 2025)*

Custom agents provide ultimate flexibility for implementing arbitrary orchestration logic beyond predefined workflow patterns. They can coordinate multiple sub-agents, integrate with external systems, and implement complex conditional logic using standard Go language constructs.

#### [**Agent Custom Story**](agent-custom-story/README.md)
Multi-LLM coordination with conditional logic for story generation, review, and editing.

**Features:**
- Sub-agent orchestration
- Conditional logic based on analysis results
- State management across multiple steps
- Event emission for monitoring

```bash
cd agent-custom-story && go run main.go
```

#### [**Agent Custom Calculator**](agent-custom-calculator/README.md)
Pure computational logic without LLM dependencies.

**Features:**
- Stateless computation patterns
- Input validation and error handling
- Simple custom agent implementation
- Integration with workflow agents

```bash
cd agent-custom-calculator && go run main.go
```

#### [**Agent Custom Data Pipeline**](agent-custom-data-pipeline/README.md)
Database operations combined with data validation and processing.

**Features:**
- External system integration (mock database)
- Sub-agent coordination for validation and processing
- Error handling and rollback patterns
- Complex state management

```bash
cd agent-custom-data-pipeline && go run main.go
```

#### [**Agent Custom API Orchestrator**](agent-custom-api-orchestrator/README.md)
Multiple API calls with retries, fallbacks, and aggregation.

**Features:**
- External API integration patterns
- Retry logic with exponential backoff
- Parallel API calls with result aggregation
- Timeout and error handling

```bash
cd agent-custom-api-orchestrator && go run main.go
```

**Key Patterns in Custom Agents:**

1. **Sub-Agent Orchestration** - Coordinate multiple agents with custom logic
2. **External Integration** - Connect with databases, APIs, and external systems
3. **Complex State Management** - Manage state across multiple operations
4. **Error Recovery** - Implement retry logic and rollback patterns

Custom agents integrate seamlessly with workflow agents and can be used as building blocks in larger systems.

### Workflow Agents *(NEW - February 3, 2025)*

The workflow agent system provides sophisticated patterns for complex multi-step processing.

#### [**Sequential Workflow**](workflow-sequential/README.md)
Step-by-step processing with error handling and state management.

**Features:**
- Sequential execution
- Error handling strategies
- State passthrough between steps
- Hook integration for monitoring

```bash
cd workflow-sequential && go run main.go
```

#### [**Parallel Workflow**](workflow-parallel/README.md)
Concurrent processing with configurable merge strategies.

**Features:**
- Concurrent agent execution
- Multiple merge strategies (MergeAll, MergeFirst, MergeCustom)
- Configurable concurrency limits
- Timeout and error handling

```bash
cd workflow-parallel && go run main.go
```

#### [**Conditional Workflow**](workflow-conditional/README.md)
Branch-based execution with priority evaluation and multiple conditions.

**Features:**
- Condition-based branching
- Priority-based evaluation
- Multiple match support
- Default branch handling

```bash
cd workflow-conditional && go run main.go
```

#### [**Loop Workflow**](workflow-loop/README.md)
Iterative processing with count, while, and until loop patterns.

**Features:**
- Count loops for fixed iterations
- While/until loops with conditions
- Result collection and state management
- Iteration delays and error handling

```bash
cd workflow-loop && go run main.go
```

#### [**Workflow Hooks**](workflow-hooks/README.md)
Monitoring and instrumentation for workflow agents using hooks.

**Features:**
- Metrics collection and tracking
- Logging integration
- Hook composition
- Workflow monitoring

```bash
cd workflow-hooks && go run main.go
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

#### [**Provider Options**](provider-options/README.md)
Demonstration of the provider option system for configuration.

**Features:**
- Common options (HTTP client, timeouts)
- Provider-specific options
- Environment variable configuration

```bash
export OPENAI_API_KEY="your-key"
cd provider-options && go run main.go
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
- **Beginner**: `simple`, `schema`, `coercion`, `agent-simple-llm`
- **Intermediate**: `provider-openai`, `provider-anthropic`, `provider-gemini`, `provider-options`, `multimodal`, `modelinfo`, `agent-custom-calculator`
- **Advanced**: `multi`, `consensus`, `agent-structured-output`, `agent-custom-*`, `workflow-*`, `metrics`, `profiling`

### By Use Case
- **Text Generation**: `simple`, `provider-openai`, `provider-anthropic`, `provider-gemini`
- **Multimodal**: `multimodal`, `provider-gemini`
- **Configuration**: `provider-options`, `convenience`
- **Data Validation**: `schema`, `coercion`
- **Multiple Providers**: `multi`, `consensus`
- **Tools & Agents**: `agent-structured-output`, `agent-simple-llm`, `agent-custom-*`, `workflow-sequential`, `workflow-parallel`, `workflow-conditional`, `workflow-loop`, `workflow-hooks`
- **Model Discovery**: `modelinfo`
- **Monitoring**: `metrics`, `profiling`

### By Provider
- **OpenAI**: `provider-openai`, `multimodal`, `agent-structured-output`, `agent-simple-llm`
- **Anthropic**: `provider-anthropic`
- **Google Gemini**: `provider-gemini`
- **Multiple**: `multi`, `consensus`, `modelinfo`
- **Compatible APIs**: `provider-openai-compatible`
- **Mock/Provider Agnostic**: `simple`, `workflow-sequential`, `workflow-parallel`, `workflow-conditional`, `workflow-loop`, `workflow-hooks`

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