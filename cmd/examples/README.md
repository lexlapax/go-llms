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
Advanced custom agent extending BaseAgentImpl with code-based orchestration.

**Features:**
- Custom agent extending BaseAgentImpl (not LLMAgent)
- Code-based orchestration without library sub-agent features
- Multi-engine parallel search (Tavily, Brave, Serpapi, Serper.dev, DuckDuckGo)
- LLMAgent instances for intelligent processing (dedup, analysis, report)
- Complex state management between phases
- Comprehensive research report generation

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

#### [**Agent Events**](agent-events/README.md)
Enhanced event system for agent monitoring and bridge integration.

**Features:**
- EventBus with pattern-based subscriptions
- Advanced filtering (composite filters)
- Event serialization for bridge layer
- Event storage and replay
- Bridge-specific event types

```bash
cd agent-events && go run main.go
```

#### [**Agent Sub-Agents**](agent-sub-agents/README.md)
Orchestrating multiple sub-agents for complex tasks.

**Features:**
- Sub-agent management
- Task delegation
- Result aggregation
- Hierarchical agent structures

```bash
export OPENAI_API_KEY="your-key"
cd agent-sub-agents && go run main.go
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

#### [**Workflow Serialization**](workflow-serialization/README.md)
Workflow serialization for bridge layer integration.

**Features:**
- JSON/YAML workflow serialization
- Script-based workflow steps
- Multiple scripting language support
- Workflow templates
- Bridge layer deserialization

```bash
cd workflow-serialization && go run main.go
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

#### [**Provider Ollama**](provider-ollama/README.md)
Local model integration with Ollama.

**Features:**
- Local LLM hosting
- Model management
- Streaming support
- No API keys required

```bash
# Make sure Ollama is running locally
cd provider-ollama && go run main.go
```

#### [**Provider OpenRouter**](provider-openrouter/README.md)
Multi-model routing through OpenRouter.

**Features:**
- Access to 100+ models
- Single API for multiple providers
- Cost optimization
- Model comparison

```bash
export OPENROUTER_API_KEY="your-key"
cd provider-openrouter && go run main.go
```

#### [**Provider Vertex AI**](provider-vertexai/README.md)
Google Cloud Vertex AI integration.

**Features:**
- Enterprise-grade deployment
- Regional endpoints
- Service account authentication
- Model versioning

```bash
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account.json"
cd provider-vertexai && go run main.go
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

The built-in tools system provides pre-made, optimized tools that follow standardized interfaces. All tools are discoverable through the registry system and offer enhanced capabilities over custom implementations.

#### Using Built-in Tools

All built-in tools follow the same pattern:

1. **Import the category** to trigger registration:
   ```go
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/file"
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/system"
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/data"
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/datetime"
   import _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/feed"
   ```

2. **Discover available tools**:
   ```go
   tools.Tools.List()                    // All tools
   tools.Tools.ListByCategory("file")    // By category
   tools.Tools.Search("read")            // By search term
   ```

3. **Use the tools**:
   ```go
   tool, _ := tools.GetTool("file_read")
   result, err := tool.Execute(ctx, params)
   ```

#### Benefits Over Custom Tools

- **Consistency**: Standardized interfaces across all projects
- **Features**: Enhanced capabilities (streaming, timeouts, metadata)
- **Discovery**: Easy to find and understand available tools
- **Maintenance**: Updates and fixes handled centrally
- **Performance**: Optimized implementations with pooling

#### [**Built-ins Discovery**](builtins-discovery/README.md)
**Focus**: Registry discovery and basic usage

**Features:**
- How to discover available built-in tools
- Search and filter tools by category/tags
- Basic tool usage with agents
- Migration from custom tools to built-ins

**When to use**: Start here to understand the registry system and available tools.

```bash
cd builtins-discovery && go run main.go
```

#### [**Built-ins File Tools**](builtins-file-tools/README.md)
**Focus**: Deep dive into file tool capabilities

**Features:**
- Enhanced file reading (streaming, metadata, line ranges)
- Atomic file writing with backups
- Binary file detection
- Large file handling
- Agent integration for file operations

**When to use**: When you need to understand the full capabilities of file tools.

```bash
cd builtins-file-tools && go run main.go
```

#### [**Built-ins Web Tools**](builtins-web-tools/README.md)
**Focus**: Web interaction and HTTP operations

**Features:**
- Web fetching with timeouts and headers
- Web search using DuckDuckGo, Brave, Tavily, Serpapi, and Serper.dev
- Web scraping with CSS selectors
- Advanced HTTP requests (all methods, auth, custom headers)
- Response metadata and timing information

**When to use**: When you need to interact with web services, APIs, or scrape web content.

```bash
cd builtins-web-tools && go run main.go
```

#### [**Built-ins Web API Client**](builtins-web-api-client/README.md)
**Focus**: REST API interaction with authentication

**Features:**
- All HTTP methods with JSON support
- Multiple authentication methods (API key, Bearer, Basic)
- Path parameter substitution
- Custom headers and timeouts
- Comprehensive error handling

**When to use**: When you need to interact with REST APIs that require authentication and complex request handling.

```bash
cd builtins-web-api-client && go run main.go
```

#### [**Built-ins GraphQL Client**](builtins-graphql-client/README.md)
**Focus**: GraphQL API interaction with introspection and variable support

**Features:**
- GraphQL query and mutation execution
- Schema introspection and discovery
- Variable support for dynamic queries
- GraphQL-specific error handling
- Support for nested queries and field selection

**When to use**: When you need to interact with GraphQL APIs like GitHub GraphQL, Shopify, or other modern APIs.

```bash
cd builtins-graphql-client && go run main.go
```

#### [**Built-ins API Client Authentication**](builtins-api-client-auth/README.md)
**Focus**: Advanced authentication features for API interactions

**Features:**
- OAuth2 bearer token and access token authentication
- Custom header authentication with prefixes
- Automatic authentication detection from state
- API key in headers, query parameters, or cookies
- Session/cookie management across requests
- OAuth2 configuration and token exchange
- Multiple authentication method fallbacks

**When to use**: When you need to interact with APIs that require advanced authentication methods beyond basic API keys.

```bash
cd builtins-api-client-auth && go run main.go
```

#### [**Built-ins OpenAPI Discovery**](builtins-openapi-discovery/README.md)
**Focus**: OpenAPI/Swagger specification discovery and validation

**Features:**
- OpenAPI 3.0/3.1 spec discovery
- Operation enumeration and metadata extraction
- Request validation against OpenAPI schemas
- LLM-friendly operation guidance
- Support for GitHub, PetStore, and custom APIs

**When to use**: When you need to discover API endpoints, validate requests, or work with OpenAPI-documented APIs.

```bash
cd builtins-openapi-discovery && go run main.go
```

#### [**Built-ins Web Search Parallel**](builtins-web-search-parallel/README.md)
**Focus**: Production web search with explicit API key management

**Features:**
- Parallel searches across multiple engines
- Explicit API key injection (no environment variables)
- Performance comparison between engines
- Multi-tenant and A/B testing patterns
- Error handling and fallback strategies

**When to use**: When you need production-grade web search with secure API key management.

```bash
cd builtins-web-search-parallel && go run main.go
```

#### [**Built-ins System Tools**](builtins-system-tools/README.md)
**Focus**: System interaction and management

**Features:**
- Command execution with safety controls and timeouts
- Environment variable access with pattern matching
- Comprehensive system information gathering
- Process listing and filtering
- Cross-platform compatibility

**When to use**: When you need to interact with the operating system, run commands, or gather system information.

```bash
cd builtins-system-tools && go run main.go
```

#### [**Built-ins Data Tools**](builtins-data-tools/README.md)
**Focus**: Structured data processing

**Features:**
- JSON processing with JSONPath queries
- CSV parsing, filtering, and statistics
- XML parsing with XPath and JSON conversion
- Common data transformations (filter, map, reduce, sort, group)
- Type conversions and aggregations

**When to use**: When you need to process, transform, or analyze structured data in various formats.

```bash
cd builtins-data-tools && go run main.go
```

#### [**Built-ins DateTime Tools**](builtins-datetime-tools/README.md)
**Focus**: Comprehensive date/time operations

**Features:**
- Current time in various formats and timezones
- Date parsing with auto-detection and relative dates
- Date arithmetic and business day calculations
- Formatting with localization support
- Timezone conversions with DST handling
- Date comparisons and sorting

**When to use**: When you need to work with dates, times, and timezones in your applications.

```bash
cd builtins-datetime-tools && go run main.go
```

#### [**Built-ins Feed Tools**](builtins-feed-tools/README.md)
**Focus**: Feed processing and syndication

**Features:**
- Fetching RSS, Atom, and JSON Feed formats
- Auto-discovering feeds from websites
- Filtering feed items by keywords, dates, and categories
- Aggregating multiple feeds into one
- Converting between feed formats
- Extracting specific data for analysis

**When to use**: When you need to work with RSS/Atom feeds, aggregate news sources, or build feed-based applications.

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

#### [**Structured Output**](structured-output/README.md)
LLM output parsing with recovery and validation.

**Features:**
- JSON parsing with recovery from markdown
- Schema validation with detailed errors
- Format conversion (JSON/YAML/XML)
- Bridge integration for go-llmspell

```bash
cd structured-output && go run main.go
```

### Schema Management Examples

#### [**Schema Generator**](schema-generator/README.md)
Advanced schema generation from Go structs.

**Features:**
- Reflection-based schema generation
- Tag-based schema generation
- Custom type handlers
- Nested struct support
- Schema versioning

```bash
cd schema-generator && go run main.go
```

#### [**Schema Repository**](schema-repository/README.md)
Schema storage and versioning.

**Features:**
- In-memory repository with thread safety
- File-based persistent storage
- Schema versioning and migration
- Import/export functionality
- Schema discovery and search

```bash
cd schema-repository && go run main.go
```

### Bridge Integration Examples

#### [**Types Bridge**](types-bridge/README.md)
Type conversion and bridging for scripting engines.

**Features:**
- Type registry with bidirectional conversions
- Schema to map[string]interface{} conversion
- Custom type converters
- Multi-hop conversion support
- Conversion caching

```bash
cd types-bridge && go run main.go
```

#### [**Tools Script Dynamic**](tools-script-dynamic/README.md)
Dynamic tool registration for scripting engines.

**Features:**
- Runtime tool registration
- Script-based tool factories
- Tool persistence and loading
- Multi-tenant tool isolation
- Tool versioning

```bash
cd tools-script-dynamic && go run main.go
```

### Error Handling Examples

#### [**Provider Metadata**](provider-metadata/README.md)
Provider metadata and dynamic registry system.

**Features:**
- Provider capability discovery
- Model information with pricing
- Dynamic provider registration
- Configuration export/import
- Best model selection by requirements

```bash
cd provider-metadata && go run main.go
```

#### [**Errors Serialization**](errors-serialization/README.md)
Enhanced error handling with serialization and recovery.

**Features:**
- JSON serializable errors
- Rich error context with stack traces
- Recovery strategies (exponential backoff, circuit breaker)
- Error aggregation for batch operations
- Error builder pattern

```bash
cd errors-serialization && go run main.go
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