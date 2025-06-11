# Examples Gallery

This gallery showcases various examples demonstrating Go-LLMs capabilities, organized by use case.

## Quick Start Examples

### Getting Started Without API Keys

**[Simple](../../cmd/examples/simple/)** - Basic mock provider example
```bash
cd cmd/examples/simple && go run main.go
```
Perfect for understanding the basics without needing API keys.

### Your First LLM Agent

**[Agent Simple LLM](../../cmd/examples/agent-simple-llm/)** - Minimal agent setup
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-simple-llm && go run main.go
```
Shows how to create an agent with just a few lines of code.

## Tool Integration Examples

### Calculator and Math

**[Agent Calculator](../../cmd/examples/agent-calculator/)** - Math operations with LLM
```bash
# LLM mode (default)
cd cmd/examples/agent-calculator && go run main.go

# Direct tool usage
go run main.go direct

# Tool information
go run main.go info
```
Demonstrates the enhanced calculator tool with natural language interface.

### Web Research and APIs

**[Built-ins Web Tools](../../cmd/examples/builtins-web-tools/)** - Web scraping and search
```bash
cd cmd/examples/builtins-web-tools && go run main.go
```
Shows web search, fetch, scrape, and HTTP request tools.

**[Built-ins Web API Client](../../cmd/examples/builtins-web-api-client/)** - REST API interactions
```bash
cd cmd/examples/builtins-web-api-client && go run main.go
```
Advanced API client with authentication and error handling.

**[Built-ins GraphQL Client](../../cmd/examples/builtins-graphql-client/)** - GraphQL APIs
```bash
cd cmd/examples/builtins-graphql-client && go run main.go
```
GraphQL queries, mutations, and introspection.

### File Operations

**[Built-ins File Tools](../../cmd/examples/builtins-file-tools/)** - File management
```bash
cd cmd/examples/builtins-file-tools && go run main.go
```
Read, write, list, search, and manage files safely.

### Data Processing

**[Built-ins Data Tools](../../cmd/examples/builtins-data-tools/)** - Structured data processing
```bash
cd cmd/examples/builtins-data-tools && go run main.go
```
Process JSON, CSV, XML with queries and transformations.

### Date and Time

**[Built-ins DateTime Tools](../../cmd/examples/builtins-datetime-tools/)** - Time operations
```bash
cd cmd/examples/builtins-datetime-tools && go run main.go
```
Parse, format, calculate, and convert dates/times across timezones.

### Feed Processing

**[Built-ins Feed Tools](../../cmd/examples/builtins-feed-tools/)** - RSS/Atom feeds
```bash
cd cmd/examples/builtins-feed-tools && go run main.go
```
Fetch, filter, aggregate, and convert news feeds.

### System Operations

**[Built-ins System Tools](../../cmd/examples/builtins-system-tools/)** - System interaction
```bash
cd cmd/examples/builtins-system-tools && go run main.go
```
Execute commands, read environment, get system info safely.

## Agent Patterns

### Structured Output

**[Agent Structured Output](../../cmd/examples/agent-structured-output/)** - Type-safe LLM responses
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-structured-output && go run main.go
```
Extract structured data from LLM responses with validation.

### Error Handling

**[Agent Error Handling](../../cmd/examples/agent-error-handling/)** - Robust error management
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-error-handling && go run main.go
```
Retry mechanisms, fallbacks, and graceful degradation.

### Custom Agents

**[Agent Custom Research](../../cmd/examples/agent-custom-research/)** - Advanced orchestration
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-custom-research && go run main.go
```
Build a research agent with multiple search engines and sub-agents.

### Agent Tools

**[Agent LLM Built-in Tools](../../cmd/examples/agent-llm-builtin-tools/)** - All tools showcase
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-llm-builtin-tools && go run main.go research "Find Go concurrency patterns"
```
Demonstrates agents using different tool categories.

### Metrics and Monitoring

**[Agent Metrics Tools](../../cmd/examples/agent-metrics-tools/)** - Performance tracking
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-metrics-tools && go run main.go
```
Monitor token usage, response times, and tool execution.

## Workflow Examples

### Sequential Processing

**[Workflow Sequential](../../cmd/examples/workflow-sequential/)** - Step-by-step workflows
```bash
cd cmd/examples/workflow-sequential && go run main.go
```
Chain agents for multi-step processing.

### Parallel Execution

**[Workflow Parallel](../../cmd/examples/workflow-parallel/)** - Concurrent workflows
```bash
cd cmd/examples/workflow-parallel && go run main.go
```
Run multiple agents in parallel with merge strategies.

### Conditional Logic

**[Workflow Conditional](../../cmd/examples/workflow-conditional/)** - Branching workflows
```bash
cd cmd/examples/workflow-conditional && go run main.go
```
Dynamic routing based on conditions.

### Loops and Iteration

**[Workflow Loop](../../cmd/examples/workflow-loop/)** - Iterative processing
```bash
cd cmd/examples/workflow-loop && go run main.go
```
Fixed count, while, and until loops.

### Complex Orchestration

**[Workflow Composition](../../cmd/examples/workflow-composition/)** - Nested workflows
```bash
cd cmd/examples/workflow-composition && go run main.go
```
Compose complex workflows from simpler ones.

## Multi-Provider Examples

### Provider Strategies

**[Provider Multi](../../cmd/examples/provider-multi/)** - Multiple providers
```bash
export OPENAI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
cd cmd/examples/provider-multi && go run main.go
```
Fastest, primary with fallback, load balancing.

### Consensus

**[Provider Consensus](../../cmd/examples/provider-consensus/)** - Agreement strategies
```bash
export OPENAI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
export GEMINI_API_KEY="your-key"
cd cmd/examples/provider-consensus && go run main.go
```
Get consensus from multiple providers.

### Agent-Level Multi-Provider

**[Workflow Multi-Provider](../../cmd/examples/workflow-multi-provider/)** - Agent strategies
```bash
export OPENAI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
cd cmd/examples/workflow-multi-provider && go run main.go
```
Multi-provider patterns at the agent level.

## Advanced Features

### Tool Conversion

**[Agent Tools Conversion](../../cmd/examples/agent-tools-conversion/)** - Tools ↔ Agents
```bash
cd cmd/examples/agent-tools-conversion && go run main.go
```
Convert between tools and agents bidirectionally.

### Workflow as Tool

**[Agent Workflow as Tool](../../cmd/examples/agent-workflow-as-tool/)** - Wrap workflows
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-workflow-as-tool && go run main.go
```
Use complex workflows as simple tools.

### State Management

**[Agent State Persistence](../../cmd/examples/agent-state-persistence/)** - Save/load state
```bash
cd cmd/examples/agent-state-persistence && go run main.go
```
Persist agent state across sessions.

### Guardrails

**[Agent Guardrails](../../cmd/examples/agent-guardrails/)** - Safety controls
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-guardrails && go run main.go
```
Input validation and output filtering.

### Agent Handoff

**[Agent Handoff](../../cmd/examples/agent-handoff/)** - Agent delegation
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-handoff && go run main.go
```
Pass control between specialized agents.

### Multi-Agent Coordination

**[Agent Multi-Coordination](../../cmd/examples/agent-multi-coordination/)** - Agent teams
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-multi-coordination && go run main.go
```
Coordinate multiple agents for complex tasks.

## Provider-Specific Examples

### OpenAI

**[Provider OpenAI](../../cmd/examples/provider-openai/)** - OpenAI integration
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/provider-openai && go run main.go
```
GPT-4, structured output, streaming.

### Anthropic

**[Provider Anthropic](../../cmd/examples/provider-anthropic/)** - Claude integration
```bash
export ANTHROPIC_API_KEY="your-key"
cd cmd/examples/provider-anthropic && go run main.go
```
Claude models with specific optimizations.

### Google Gemini

**[Provider Gemini](../../cmd/examples/provider-gemini/)** - Gemini integration
```bash
export GEMINI_API_KEY="your-key"
cd cmd/examples/provider-gemini && go run main.go
```
Gemini models with multimodal support.

### OpenAI Compatible

**[Provider OpenAI Compatible](../../cmd/examples/provider-openai-compatible/)** - Compatible APIs
```bash
# For OpenRouter
export OPENROUTER_API_KEY="your-key"
cd cmd/examples/provider-openai-compatible && go run main.go

# For Ollama (local)
cd cmd/examples/provider-openai-compatible && go run main.go
```
OpenRouter, Ollama, and custom endpoints.

### Multimodal

**[Provider Multimodal](../../cmd/examples/provider-multimodal/)** - Images and media
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/provider-multimodal && go run main.go -provider openai -mode image -a image.jpg
```
Process images, audio, and video across providers.

## Utility Examples

### Model Discovery

**[Utils Model Info](../../cmd/examples/utils-modelinfo/)** - Model capabilities
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/utils-modelinfo && go run main.go
```
Discover and compare model capabilities.

### Performance Profiling

**[Utils Profiling](../../cmd/examples/utils-profiling/)** - Performance analysis
```bash
cd cmd/examples/utils-profiling && go run main.go
```
CPU and memory profiling tools.

### Schema Validation

**[Structured Schema](../../cmd/examples/structured-schema/)** - Schema generation
```bash
cd cmd/examples/structured-schema && go run main.go
```
Generate and validate JSON schemas.

**[Structured Coercion](../../cmd/examples/structured-coercion/)** - Type coercion
```bash
cd cmd/examples/structured-coercion && go run main.go
```
Flexible type conversion and validation.

## Discovery Examples

### Tool Discovery

**[Built-ins Discovery](../../cmd/examples/builtins-discovery/)** - Find tools
```bash
cd cmd/examples/builtins-discovery && go run main.go
```
Discover available tools by category and search.

### API Discovery

**[Built-ins OpenAPI Discovery](../../cmd/examples/builtins-openapi-discovery/)** - API exploration
```bash
cd cmd/examples/builtins-openapi-discovery && go run main.go
```
Discover API endpoints from OpenAPI specs.

### Authentication Patterns

**[Built-ins API Client Auth](../../cmd/examples/builtins-api-client-auth/)** - Auth methods
```bash
cd cmd/examples/builtins-api-client-auth && go run main.go
```
Various authentication patterns for APIs.

### Parallel Search

**[Built-ins Web Search Parallel](../../cmd/examples/builtins-web-search-parallel/)** - Multi-engine search
```bash
cd cmd/examples/builtins-web-search-parallel && go run main.go
```
Search across multiple engines in parallel.

## Integration Patterns

### Hooks and Monitoring

**[Workflow Hooks](../../cmd/examples/workflow-hooks/)** - Instrumentation
```bash
cd cmd/examples/workflow-hooks && go run main.go
```
Add logging, metrics, and monitoring to workflows.

### Advanced Tool Context

**[Agent Advanced Tool Context](../../cmd/examples/agent-advanced-toolcontext/)** - Tool features
```bash
cd cmd/examples/agent-advanced-toolcontext && go run main.go
```
State access and event emission from tools.

### Sub-Agents

**[Agent Sub-Agents](../../cmd/examples/agent-sub-agents/)** - Agent hierarchy
```bash
export OPENAI_API_KEY="your-key"
cd cmd/examples/agent-sub-agents && go run main.go
```
Agents using other agents as tools.

## Tips for Running Examples

1. **Start Simple**: Begin with the `simple` example to understand basics
2. **Check Requirements**: Some examples need API keys (see each README)
3. **Use Debug Mode**: Set `DEBUG=1` for detailed logging
4. **Read READMEs**: Each example has detailed documentation
5. **Modify and Experiment**: Examples are designed to be modified

## Environment Setup

Most examples use these environment variables:

```bash
# Provider API Keys
export OPENAI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
export GEMINI_API_KEY="your-key"

# Optional: Debug logging
export DEBUG=1

# Optional: Custom endpoints
export OPENAI_BASE_URL="https://api.openai.com"
```

## Building Examples

Build all examples:
```bash
make build-examples
```

Build specific example:
```bash
make build-example EXAMPLE=agent-calculator
```

## Creating Your Own Examples

1. Create directory: `cmd/examples/my-example/`
2. Add `main.go` with ABOUTME comments
3. Add `README.md` with documentation
4. Follow existing patterns
5. Test thoroughly
6. Update this gallery

## Common Patterns

### Creating an Agent
```go
deps := core.LLMDeps{Provider: provider}
agent := core.NewLLMAgent("my-agent", "My Agent", deps)
agent.SetSystemPrompt("You are a helpful assistant")
```

### Adding Tools
```go
agent.AddTool(tools.MustGetTool("web_search"))
agent.AddTool(tools.MustGetTool("file_read"))
```

### Running Agent
```go
state := domain.NewState()
state.Set("prompt", "Your task here")
result, err := agent.Run(ctx, state)
```

### Using Workflows
```go
workflow := workflow.NewSequentialWorkflow("my-workflow")
workflow.AddStep("step1", agent1)
workflow.AddStep("step2", agent2)
result, err := workflow.Run(ctx, state)
```

## Need Help?

- Check example READMEs for detailed documentation
- Review test files for additional usage patterns
- See main documentation for API reference
- Open issues for bugs or feature requests