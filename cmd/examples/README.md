# Go-LLMs Examples

This directory contains example applications demonstrating various features and capabilities of the Go-LLMs library. Each example is self-contained and includes its own documentation.

## Example Categories

The examples are organized into the following categories:

- **agent-*** - Agent-specific features and patterns
- **workflow-*** - Workflow agent patterns (sequential, parallel, conditional, loop)
- **provider-*** - Provider-specific features and integrations
- **builtins-*** - Built-in tool demonstrations
- **utils-*** - Utility packages and helpers
- **structured-*** - Structured output and validation features

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

### Agent Examples

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

#### [**Agent Calculator**](agent-calculator/README.md)
Calculator tool usage with LLM agents.

**Features:**
- Built-in calculator tool integration
- Mathematical constant handling
- LLM tool calling patterns
- Debugging with conditional logging

```bash
cd agent-calculator && go run main.go
```

#### [**Agent Custom Research**](agent-custom-research/README.md)
Advanced custom agent extending LLMAgent with sub-agent coordination.

**Features:**
- Custom agent extending LLMAgent
- Multi-phase research pipeline
- Sub-agent coordination (searcher, summarizer, fact-checker)
- Tool usage (web search, web fetch)
- Complex state management
- Research report synthesis

```bash
export OPENAI_API_KEY="your-key"  # Optional for real LLM sub-agents
cd agent-custom-research && go run main.go
```

#### [**Agent LLM Built-in Tools**](agent-llm-builtin-tools/README.md)
Demonstrates using built-in tools with LLM agents.

**Features:**
- Tool integration patterns
- Built-in tool usage
- State management with tools

```bash
export OPENAI_API_KEY="your-key"
cd agent-llm-builtin-tools && go run main.go
```

#### [**Agent Metrics Tools**](agent-metrics-tools/README.md)
Performance monitoring and metrics collection for agents with tools.

**Features:**
- Real LLM provider support (OpenAI, Anthropic, Gemini)
- ToolContext pattern demonstration
- Response time tracking
- Token usage monitoring
- Tool execution statistics
- Error rate analysis

```bash
export OPENAI_API_KEY="your-key"
cd agent-metrics-tools && go run main.go
```

#### [**Agent Tools Conversion**](agent-tools-conversion/README.md)
Bidirectional conversion between agents and tools with registry integration.

**Features:**
- Agent to Tool conversion
- Tool to Agent conversion
- Registry integration
- Event forwarding
- Schema mapping
- Tool chains

```bash
cd agent-tools-conversion && go run main.go
```

#### [**Agent Workflow as Tool**](agent-workflow-as-tool/README.md)
Multi-stage research pipeline demonstrating workflow agents wrapped as tools.

**Features:**
- Sequential workflow as tool
- Parallel workflow as tool
- Custom merge strategies
- LLM agent orchestration
- Real-world use case

```bash
export OPENAI_API_KEY="your-key"
cd agent-workflow-as-tool && go run main.go
```

#### [**Agent Advanced Tool Context**](agent-advanced-toolcontext/README.md)
Advanced tool context features including state access and event emission.

**Features:**
- State access from tools
- Event emission
- Tool context patterns

```bash
cd agent-advanced-toolcontext && go run main.go
```

#### [**Agent Error Handling**](agent-error-handling/README.md)
Error handling patterns for agents.

**Features:**
- Error recovery strategies
- Retry mechanisms
- Graceful degradation

```bash
export OPENAI_API_KEY="your-key"
cd agent-error-handling && go run main.go
```

#### [**Agent State Persistence**](agent-state-persistence/README.md)
State persistence and serialization for agents.

**Features:**
- State saving and loading
- Persistence strategies
- State migration

```bash
cd agent-state-persistence && go run main.go
```

#### [**Agent Guardrails**](agent-guardrails/README.md)
Input and output validation using guardrails.

**Features:**
- Input validation
- Output filtering
- Safety mechanisms

```bash
export OPENAI_API_KEY="your-key"
cd agent-guardrails && go run main.go
```

#### [**Agent Handoff**](agent-handoff/README.md)
Agent-to-agent handoff patterns and delegation.

**Features:**
- Handoff builder pattern
- Agent chain execution
- Conditional routing
- State preservation

```bash
export OPENAI_API_KEY="your-key"
cd agent-handoff && go run main.go
```

#### [**Agent Multi-Coordination**](agent-multi-coordination/README.md)
Multiple agents coordinating to solve complex tasks.

**Features:**
- Multi-agent orchestration
- Parallel coordination
- Conditional routing
- Event monitoring

```bash
export OPENAI_API_KEY="your-key"
cd agent-multi-coordination && go run main.go
```

### Workflow Examples

#### [**Workflow Sequential**](workflow-sequential/README.md)
Step-by-step processing with error handling and state management.

**Features:**
- Sequential execution
- Error handling strategies
- State passthrough between steps
- Hook integration

```bash
cd workflow-sequential && go run main.go
```

#### [**Workflow Parallel**](workflow-parallel/README.md)
Concurrent processing with configurable merge strategies.

**Features:**
- Concurrent agent execution
- Multiple merge strategies
- Configurable concurrency limits
- Timeout and error handling

```bash
cd workflow-parallel && go run main.go
```

#### [**Workflow Conditional**](workflow-conditional/README.md)
Branch-based execution with priority evaluation.

**Features:**
- Condition-based branching
- Priority-based evaluation
- Multiple match support
- Default branch handling

```bash
cd workflow-conditional && go run main.go
```

#### [**Workflow Loop**](workflow-loop/README.md)
Iterative processing with count, while, and until loops.

**Features:**
- Count loops for fixed iterations
- While/until loops with conditions
- Result collection
- Iteration delays

```bash
cd workflow-loop && go run main.go
```

#### [**Workflow Hooks**](workflow-hooks/README.md)
Monitoring and instrumentation for workflow agents.

**Features:**
- Metrics collection
- Logging integration
- Hook composition
- Workflow monitoring

```bash
cd workflow-hooks && go run main.go
```

#### [**Workflow Composition**](workflow-composition/README.md)
Complex workflow composition patterns.

**Features:**
- Nested workflows
- Dynamic composition
- Advanced patterns

```bash
cd workflow-composition && go run main.go
```

#### [**Workflow Multi-Provider**](workflow-multi-provider/README.md)
Agent-level multi-provider patterns using workflows.

**Features:**
- Agent-based provider strategies
- Consensus at agent level
- Fallback patterns

```bash
export OPENAI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
cd workflow-multi-provider && go run main.go
```

### Provider Examples

#### [**Provider OpenAI**](provider-openai/README.md)
Direct integration with OpenAI's GPT models.

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
Direct integration with Anthropic's Claude models.

**Features:**
- Conversation handling
- System prompts
- Claude-specific optimizations

```bash
export ANTHROPIC_API_KEY="your-key"
cd provider-anthropic && go run main.go
```

#### [**Provider Gemini**](provider-gemini/README.md)
Direct integration with Google's Gemini models.

**Features:**
- Google AI integration
- Multimodal capabilities
- Safety settings

```bash
export GEMINI_API_KEY="your-key"
cd provider-gemini && go run main.go
```

#### [**Provider OpenAI Compatible**](provider-openai-compatible/README.md)
Working with OpenAI-compatible APIs.

**Features:**
- OpenRouter integration
- Ollama local models
- Custom API endpoints

```bash
export OPENROUTER_API_KEY="your-key"
cd provider-openai-compatible && go run main.go
```

#### [**Provider Multimodal**](provider-multimodal/README.md)
Comparison of multimodal capabilities across providers.

**Features:**
- Provider-specific multimodal support
- Image processing
- Audio/video handling (Gemini)
- File uploads and URLs

```bash
export OPENAI_API_KEY="your-key"
cd provider-multimodal && go run main.go -provider openai -mode image -a image.jpg
```

#### [**Provider Multi**](provider-multi/README.md)
Provider-level multi-provider strategies.

**Features:**
- Fastest strategy
- Primary with fallback
- Load balancing

Note: For agent-level multi-provider patterns, see workflow-multi-provider example.

```bash
export OPENAI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
cd provider-multi && go run main.go
```

#### [**Provider Consensus**](provider-consensus/README.md)
Provider-level consensus strategies.

**Features:**
- Similarity-based consensus
- Voting mechanisms
- Quality scoring

Note: For agent-level consensus patterns, see workflow-multi-provider example.

```bash
export OPENAI_API_KEY="your-key" 
export ANTHROPIC_API_KEY="your-key"
export GEMINI_API_KEY="your-key"
cd provider-consensus && go run main.go
```

#### [**Provider Options**](provider-options/README.md)
Provider configuration options.

**Features:**
- Common options
- Provider-specific options
- Environment variables

```bash
export OPENAI_API_KEY="your-key"
cd provider-options && go run main.go
```

#### [**Provider Convenience**](provider-convenience/README.md)
Provider-level utility functions and helpers.

**Features:**
- Provider pools
- Retry mechanisms
- Configuration helpers

```bash
cd provider-convenience && go run main.go
```

### Built-in Tools Examples

#### [**Built-ins Discovery**](builtins-discovery/README.md)
Introduction to the built-in tools registry system.

**Features:**
- Tool discovery and search
- Registry categories and tags
- Basic agent integration

```bash
cd builtins-discovery && go run main.go
```

#### [**Built-ins File Tools**](builtins-file-tools/README.md)
Enhanced file operation tools.

**Features:**
- Streaming large files
- Atomic writes with backups
- Binary file detection
- File metadata

```bash
cd builtins-file-tools && go run main.go
```

#### [**Built-ins Web Tools**](builtins-web-tools/README.md)
Web interaction tools.

**Features:**
- Web page fetching
- DuckDuckGo search
- Web scraping
- HTTP requests

```bash
cd builtins-web-tools && go run main.go
```

#### [**Built-ins System Tools**](builtins-system-tools/README.md)
System information and command execution.

**Features:**
- Safe command execution
- Environment variables
- System information
- Process listing

```bash
cd builtins-system-tools && go run main.go
```

#### [**Built-ins Data Tools**](builtins-data-tools/README.md)
Data processing and transformation.

**Features:**
- JSON processing
- CSV parsing
- XML conversion
- Data transformations

```bash
cd builtins-data-tools && go run main.go
```

#### [**Built-ins DateTime Tools**](builtins-datetime-tools/README.md)
Date and time manipulation.

**Features:**
- Timezone handling
- Date arithmetic
- Parsing and formatting
- Business days

```bash
cd builtins-datetime-tools && go run main.go
```

#### [**Built-ins Feed Tools**](builtins-feed-tools/README.md)
RSS, Atom, and JSON Feed processing.

**Features:**
- Feed parsing
- Auto-discovery
- Filtering
- Format conversion

```bash
cd builtins-feed-tools && go run main.go
```

### Structured Output Examples

#### [**Structured Schema**](structured-schema/README.md)
Schema generation and validation.

**Features:**
- Schema generation from Go structs
- Validation patterns
- Custom rules

```bash
cd structured-schema && go run main.go
```

#### [**Structured Coercion**](structured-coercion/README.md)
Type coercion and data conversion.

**Features:**
- Type conversions
- Flexible validation
- Data normalization

```bash
cd structured-coercion && go run main.go
```

### Utility Examples

#### [**Utils Model Info**](utils-modelinfo/README.md)
Model discovery and capability assessment.

**Features:**
- Automatic model discovery
- Capability filtering
- Provider comparison
- Caching

```bash
export OPENAI_API_KEY="your-key"
cd utils-modelinfo && go run main.go
```

#### [**Utils Profiling**](utils-profiling/README.md)
Performance profiling and optimization.

**Features:**
- CPU profiling
- Memory analysis
- Benchmarks

```bash
cd utils-profiling && go run main.go
```

## Environment Variables

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
```

## Building Examples

You can build all examples at once using the provided Makefile:

```bash
# Build all examples
make build-examples

# Build a specific example
make build-example EXAMPLE=agent-simple-llm

# Clean all built examples
make clean-examples
```

## Running Tests

Each example includes tests:

```bash
cd <example-directory>
go test ./...
```

## Contributing

When adding new examples:

1. Create a new directory under `cmd/examples/`
2. Follow the naming convention (agent-*, workflow-*, provider-*, etc.)
3. Include `main.go` and `README.md`
4. Add ABOUTME comments to source files
5. Update this README.md