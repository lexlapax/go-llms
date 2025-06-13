# Go-LLMs Architecture

> **[Documentation Home](/docs/README.md) / [Technical Documentation](/docs/technical/README.md) / Architecture**

This document provides a comprehensive overview of the Go-LLMs library architecture, including design principles, component structure, data flow patterns, and system interactions.

## Table of Contents

1. [Overview](#overview)
2. [Architecture Principles](#architecture-principles)
3. [Project Structure](#project-structure)
4. [Core Components](#core-components)
5. [Data Flow](#data-flow)
6. [Agent Workflow](#agent-workflow)
7. [Multi-Provider Strategies](#multi-provider-strategies)

## Overview

Go-LLMs is a comprehensive Go library for building LLM-powered applications with strong typing, structured outputs, and a flexible agent-based architecture. The library provides a unified interface across multiple LLM providers while maintaining Go's idioms for simplicity, performance, and reliability.

### Key Components

1. **Provider Layer**: Unified interface for multiple LLM providers (OpenAI, Anthropic, Google Gemini)
2. **Schema System**: JSON Schema-based validation with type coercion
3. **Structured Output**: Extract and validate structured data from LLM responses
4. **Agent Framework**: Tool-enabled agents with state management and workflows
5. **Built-in Components**: Pre-built tools, agents, and workflows for common tasks

![Go-LLMs Architecture Overview](/docs/images/architecture_overview.svg)

## Architecture Principles

The library follows these core architectural principles:

### 1. Vertical Slicing
Code is organized by feature (e.g., `llm/`, `schema/`, `agent/`) rather than technical layers, promoting:
- Better feature cohesion
- Easier navigation and maintenance
- Clear ownership boundaries
- Independent feature development

### 2. Interface-Based Design
All major components are defined through interfaces:
- `domain.Provider` for LLM providers
- `domain.Tool` for agent tools
- `domain.BaseAgent` for all agent types
- Enables easy mocking and testing
- Allows runtime provider switching

### 3. Clean Architecture
Clear separation of concerns:
- **Domain Layer**: Core business logic and interfaces
- **Implementation Layer**: Concrete implementations
- **Adapter Layer**: External integrations
- No circular dependencies between packages

### 4. Dependency Injection
Components receive dependencies explicitly:
- Constructor injection for required dependencies
- Functional options for optional configuration
- No global state or singletons (except registries)
- Promotes testability and flexibility

### 5. Performance-First Design
- Object pooling with `sync.Pool` for frequently allocated objects
- Efficient string building and JSON processing
- Context-aware cancellation and timeouts
- Concurrent provider strategies

## Project Structure

The codebase follows a feature-based organization:

```
go-llms/
├── cmd/                      # Application entry points
│   ├── main.go              # CLI application
│   └── examples/            # 40+ example applications
├── pkg/                      # Public packages
│   ├── llm/                 # LLM provider integration
│   │   ├── domain/          # Provider interfaces and types
│   │   └── provider/        # Provider implementations
│   ├── schema/              # JSON Schema validation
│   │   ├── domain/          # Schema interfaces
│   │   ├── validation/      # Validation with coercion
│   │   └── adapter/         # Go struct reflection
│   ├── structured/          # Structured output extraction
│   │   ├── domain/          # Core interfaces
│   │   └── processor/       # JSON extraction and caching
│   ├── agent/               # Agent framework
│   │   ├── domain/          # Agent interfaces and state
│   │   ├── core/            # Core agent implementations
│   │   ├── tools/           # Tool system and conversions
│   │   ├── workflow/        # Workflow patterns
│   │   └── builtins/        # Built-in components
│   │       ├── tools/       # 32 built-in tools
│   │       └── registry.go  # Component registry
│   └── util/                # Utility packages
│       ├── llmutil/         # Provider utilities
│       ├── auth/            # Authentication
│       └── metrics/         # Performance metrics
└── tests/                    # Test suites
    ├── integration/         # Integration tests
    ├── benchmarks/          # Performance benchmarks
    └── stress/              # Stress tests
```

### Package Responsibilities

- **llm**: Provider abstraction and implementations
- **schema**: JSON Schema validation with type coercion
- **structured**: Extract structured data from LLM responses
- **agent**: Agent orchestration, tools, and workflows
- **util**: Cross-cutting concerns and helpers

## Core Components

### 1. Schema Validation System

The schema package provides comprehensive JSON Schema validation:

**Features**:
- Full JSON Schema draft-07 support
- Type coercion for flexible input handling
- Custom validators for domain-specific rules
- Nested object and array validation
- Format validators (email, URL, date-time, etc.)
- Performance optimized with caching

**Key Types**:
```go
type Schema interface {
    Validate(data interface{}) error
    ValidateWithCoercion(data interface{}) (interface{}, error)
}
```

### 2. LLM Provider Layer

Unified interface across multiple LLM providers:

**Core Interface**:
```go
type Provider interface {
    Generate(ctx context.Context, prompt string, options ...Option) (string, error)
    GenerateMessage(ctx context.Context, messages []Message, options ...Option) (Response, error)
    GenerateWithSchema(ctx context.Context, prompt string, schema *Schema, options ...Option) (interface{}, error)
    Stream(ctx context.Context, prompt string, options ...Option) (ResponseStream, error)
    StreamMessage(ctx context.Context, messages []Message, options ...Option) (ResponseStream, error)
}
```

**Provider Strategies**:
- **Single Provider**: Direct provider usage
- **Multi-Provider**: Concurrent strategies (fastest, consensus, primary-fallback)
- **Mock Provider**: Testing without API calls

**Advanced Features**:
- Multimodal content support (images, audio, video, files)
- Streaming responses with backpressure
- Standardized error types across providers
- Context-aware cancellation
- Rate limiting and retry logic

### 3. Structured Output Processing

Extract and validate structured data from LLM responses:

**Components**:
- **JSON Extractor**: Robust extraction from mixed text
- **Schema Validator**: Ensure extracted data matches schema
- **Prompt Enhancer**: Augment prompts with schema information
- **Result Cache**: Cache validated results

**Processing Flow**:
1. Enhance prompt with schema
2. Send to LLM provider
3. Extract JSON from response
4. Validate against schema
5. Return typed result

### 4. Agent Framework

Comprehensive agent system with tools and workflows:

**Agent Hierarchy**:
```
BaseAgent (interface)
├── LLMAgent         - LLM-powered agent with tools
├── ToolAgent        - Wraps a tool as an agent
└── WorkflowAgent    - Orchestrates multiple agents
    ├── Sequential   - Execute agents in order
    ├── Parallel     - Execute agents concurrently
    ├── Conditional  - Route based on conditions
    └── Loop         - Iterate until condition met
```

**Tool System**:
- Enhanced Tool interface with rich metadata
- ToolBuilder pattern for easy creation
- 32 built-in tools across 7 categories
- Tool-to-agent and agent-to-tool conversion
- MCP (Model Context Protocol) compatibility

**State Management**:
- Thread-safe state with read/write operations
- State transformation and validation
- Shared state across agent hierarchies
- Event-driven state updates

### 5. Built-in Components

Pre-built, production-ready components:

**Tools** (32 built-in):
- File operations (read, write, list, search)
- System tools (execute, environment, processes)
- Web tools (search, fetch, scrape, API client)
- Data processing (JSON, CSV, XML, transformations)
- DateTime operations (parsing, formatting, calculations)
- Feed processing (RSS/Atom fetch, filter, aggregate)
- Math calculations

**Registries**:
- Global tool registry with search and filtering
- Enhanced tool discovery system (v0.3.4+)
  - Metadata-first discovery without imports
  - Dynamic tool loading for scripting engines
  - Schema and example access without instantiation
- Agent template registry (planned)
- Workflow pattern registry (planned)

### 6. Utility Systems

**Authentication**:
- Multi-method support (Bearer, API Key, Basic, OAuth2)
- Automatic detection for known services
- Secure credential storage in agent state

**Model Discovery**:
- Fetch model information from provider APIs
- Capability detection (multimodal, streaming, functions)
- Unified model inventory
- Intelligent caching

**Performance Monitoring**:
- Request/response metrics
- Pool statistics
- Cache hit rates
- Custom metric collectors

## Data Flow Patterns

### Basic Generation Flow

```
Application → Provider → LLM API → Response → Application
```

1. **Request**: Application creates prompt/messages
2. **Provider Processing**: Provider formats for specific API
3. **API Call**: HTTPS request to LLM service
4. **Response**: Raw text or streaming tokens
5. **Return**: Formatted response to application

### Structured Output Flow

![Go-LLMs Data Flow](/docs/images/data_flow.svg)

```
Application → Schema Enhancement → Provider → LLM API
     ↓                                            ↓
Typed Result ← Validation ← JSON Extraction ← Response
```

1. **Schema Definition**: Define expected output structure
2. **Prompt Enhancement**: Add schema instructions to prompt
3. **Generation**: LLM generates response with JSON
4. **Extraction**: Extract JSON from mixed text
5. **Validation**: Validate against schema with coercion
6. **Type Conversion**: Convert to Go structs

### Multi-Provider Flow

```
                 ┌→ Provider A → LLM A ─┐
Application → Multi-Provider → Provider B → LLM B ─┼→ Strategy → Result
                 └→ Provider C → LLM C ─┘
```

Strategies:
- **Fastest**: Return first successful response
- **Consensus**: Compare and select best response
- **Primary-Fallback**: Try primary, fallback on error

## Agent Architecture

### Agent Execution Flow

![Go-LLMs Agent Workflow](/docs/images/agent_workflow.svg)

```
User Input → Agent → State → LLM Decision → Tool Execution → State Update → Response
                ↑                                                    ↓
                └──────────────── Continue if needed ←───────────────┘
```

### Components

#### 1. Agent Core
- **State Management**: Thread-safe state for data flow
- **Tool Registry**: Available tools for the agent
- **Message History**: Conversation context
- **Event System**: Lifecycle and tool events
- **Hook System**: Monitoring and intervention points

#### 2. LLM Integration
```go
// Agent decides to use a tool
LLM: "I need to search for information"
Tool Call: {"tool": "web_search", "params": {"query": "..."}}
Execution: tool.Execute(ctx, params)
Result: {"results": [...]}
LLM: "Based on the search results..."
```

#### 3. Tool Execution
- **Validation**: Parameters validated against schema
- **Context**: Tools receive ToolContext with state access
- **Error Handling**: Errors returned to LLM for recovery
- **Events**: Start/complete/error events emitted

#### 4. State Flow
```go
// State flows through the agent
Initial State → Agent Processing → Tool Updates → Final State
         ↓                                              ↓
    Validations                                    Results
```

### Workflow Patterns

#### Sequential Workflow
```
Agent A → State → Agent B → State → Agent C → Final Result
```
- Each agent processes state in order
- State accumulates results
- Stops on first error (configurable)

#### Parallel Workflow
```
        ┌→ Agent A ─┐
State → ├→ Agent B ─┼→ Merge → Final Result
        └→ Agent C ─┘
```
- Agents run concurrently
- Results merged by strategy
- Configurable error handling

#### Conditional Workflow
```
State → Condition → Route A → Agent A → Result
             ↓
         Route B → Agent B → Result
```
- Dynamic routing based on state
- Multiple condition types
- Default route support

#### Loop Workflow
```
State → Agent → Condition → Continue → Agent (repeat)
                    ↓
                  Exit → Final Result
```
- Iterate until condition met
- Max iteration limits
- State accumulation

## Multi-Provider Architecture

![Go-LLMs Multi-Provider Strategies](/docs/images/multi_provider.svg)

### Strategy Implementations

#### 1. Fastest Strategy
```go
type FastestStrategy struct {
    providers []WeightedProvider
    timeout   time.Duration
}
```
- **Execution**: Concurrent requests to all providers
- **Result**: First successful response wins
- **Cancellation**: Cancel remaining requests
- **Use Case**: Minimize latency

#### 2. Primary-Fallback Strategy
```go
type PrimaryStrategy struct {
    primary    Provider
    fallbacks  []Provider
    retryDelay time.Duration
}
```
- **Execution**: Try primary first
- **Fallback**: On error, try fallbacks in order
- **Consistency**: Deterministic provider selection
- **Use Case**: Cost optimization with reliability

#### 3. Consensus Strategy
```go
type ConsensusStrategy struct {
    providers  []WeightedProvider
    threshold  float64
    comparator ResponseComparator
}
```
- **Execution**: Query all providers
- **Comparison**: Semantic similarity or exact match
- **Weighting**: Provider weights affect voting
- **Use Case**: High-stakes decisions

### Error Handling

```go
// Multi-provider error aggregation
type MultiProviderError struct {
    ProviderErrors map[string]error
    AllFailed      bool
}
```

### Performance Considerations

1. **Connection Pooling**: Reuse HTTP connections
2. **Timeout Management**: Provider-specific timeouts
3. **Resource Limits**: Concurrent request limits
4. **Caching**: Response caching per provider

## Performance Architecture

### Object Pooling

```go
// High-frequency object pooling
var (
    messagePool    = &sync.Pool{New: func() interface{} { return &Message{} }}
    statePool      = &sync.Pool{New: func() interface{} { return &State{} }}
    responsePool   = &sync.Pool{New: func() interface{} { return &Response{} }}
)
```

### Caching Layers

1. **Schema Cache**: Compiled schemas cached indefinitely
2. **Prompt Cache**: Enhanced prompts cached by schema
3. **Model Info Cache**: Provider capabilities cached
4. **Tool Registry**: Pre-computed search indices

### Concurrency Patterns

1. **Provider Pools**: Limited concurrent requests per provider
2. **Worker Pools**: For parallel agent execution
3. **Stream Buffers**: Buffered channels for streaming
4. **State Locks**: Read-write locks for shared state

## Security Architecture

### Authentication Flow

```
Request → Auth Detector → Credential Store → Auth Header → API Call
              ↓                   ↓
         URL Patterns      State Storage
```

### Input Validation

1. **Schema Validation**: All tool inputs validated
2. **Prompt Injection**: Best practices in prompts
3. **File Access**: Restricted to allowed paths
4. **Command Execution**: Sandboxed with timeouts

### Credential Management

- Credentials stored in agent state, not tools
- No credentials in prompts or tool descriptions
- Automatic credential detection for known services
- Support for multiple auth methods per service

## Extension Points

### Adding Providers

1. Implement `domain.Provider` interface
2. Add provider-specific options
3. Register in provider factory
4. Add to multi-provider support

See [Provider Implementation Guide](provider-implementation.md)

### Creating Tools

1. Use ToolBuilder pattern
2. Define parameter/output schemas
3. Implement execution function
4. Register in tool registry

See [Tool Development Guide](tool-development.md)

### Custom Agents

1. Implement `domain.BaseAgent` interface
2. Define state management strategy
3. Add tool integration if needed
4. Support standard agent events

See [Agent Development Guide](agents.md)

## Future Architecture

### Planned Enhancements

1. **Plugin System**: Dynamic loading of providers/tools
2. **Distributed Agents**: Multi-node agent execution
3. **Persistent State**: Database-backed state management
4. **GraphQL Support**: Native GraphQL provider
5. **WebAssembly**: WASM tool execution
6. **Event Streaming**: Kafka/NATS integration