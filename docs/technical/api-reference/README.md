# API Reference: Complete API Documentation

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / API Reference**

Comprehensive API reference for Go-LLMs v0.3.5+, covering all public interfaces, types, methods, and constants. This reference provides detailed technical documentation for developers building applications with the Go-LLMs library.

## API Organization

The Go-LLMs API is organized into logical packages, each providing specific functionality:

### Core Packages

| Package | Description | Key Interfaces |
|---------|-------------|----------------|
| [`pkg/llm`](./providers.md) | LLM provider implementations | `Provider`, `StreamingProvider`, `EmbeddingProvider` |
| [`pkg/agent`](./agents.md) | Agent framework | `Agent`, `ToolEnabledAgent`, `WorkflowAgent` |
| [`pkg/agent/tools`](./tools.md) | Tool system | `Tool`, `ToolRegistry`, `ToolExecutor` |
| [`pkg/schema`](./types.md#schema) | JSON Schema validation | `Schema`, `SchemaRegistry`, `Validator` |
| [`pkg/structured`](./types.md#structured) | Structured output parsing | `Parser`, `OutputSchema`, `Extractor` |
| [`pkg/errors`](./types.md#errors) | Error handling system | `Error`, `ErrorType`, `ErrorContext` |

### Utility Packages

| Package | Description | Key Types |
|---------|-------------|-----------|
| `pkg/testutils` | Testing utilities | `MockProvider`, `TestFixtures`, `Assertions` |
| `pkg/util/llmutil` | LLM utilities | `ProviderParser`, `ModelInfo`, `TokenCounter` |

## Quick Navigation

- **[Provider APIs](./providers.md)** - LLM provider interfaces and implementations
- **[Agent APIs](./agents.md)** - Agent framework and workflow orchestration
- **[Tool APIs](./tools.md)** - Tool interfaces and built-in tools
- **[Type Definitions](./types.md)** - Core types, structs, and constants

## Core Interfaces Overview

### Provider Interface

The foundation for all LLM provider implementations:

```go
type Provider interface {
    // Complete generates a completion for the given request
    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
    
    // GetCapabilities returns provider capabilities
    GetCapabilities() Capabilities
    
    // GetModels returns available models
    GetModels(ctx context.Context) ([]Model, error)
}

type StreamingProvider interface {
    Provider
    // CompleteStream generates a streaming completion
    CompleteStream(ctx context.Context, request *CompletionRequest) (<-chan StreamChunk, error)
}

type EmbeddingProvider interface {
    // CreateEmbedding generates embeddings for input text
    CreateEmbedding(ctx context.Context, request *EmbeddingRequest) (*EmbeddingResponse, error)
}
```

### Agent Interface

The core abstraction for intelligent agents:

```go
type Agent interface {
    // Execute runs the agent with given input
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    
    // GetMetadata returns agent metadata
    GetMetadata() AgentMetadata
    
    // SetConfig updates agent configuration
    SetConfig(config AgentConfig) error
}

type ToolEnabledAgent interface {
    Agent
    // RegisterTool adds a tool to the agent
    RegisterTool(tool Tool) error
    
    // ExecuteTool runs a specific tool
    ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error)
}

type WorkflowAgent interface {
    Agent
    // AddStep adds a workflow step
    AddStep(step WorkflowStep) error
    
    // ExecuteWorkflow runs the complete workflow
    ExecuteWorkflow(ctx context.Context, input interface{}) (*WorkflowResult, error)
}
```

### Tool Interface

The standard interface for all tools:

```go
type Tool interface {
    // Core identification
    Name() string
    Description() string
    Version() string
    
    // Schema definition
    GetInputSchema() *jsonschema.Schema
    GetOutputSchema() *jsonschema.Schema
    
    // Execution
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    
    // Validation
    ValidateInput(input interface{}) error
    
    // Lifecycle
    Initialize(ctx context.Context) error
    Cleanup(ctx context.Context) error
}
```

## Common Types

### Request/Response Types

```go
// CompletionRequest represents a completion API request
type CompletionRequest struct {
    Messages    []Message              `json:"messages"`
    Model       string                 `json:"model,omitempty"`
    Temperature *float64               `json:"temperature,omitempty"`
    MaxTokens   *int                   `json:"max_tokens,omitempty"`
    Stream      bool                   `json:"stream,omitempty"`
    Tools       []ToolDefinition       `json:"tools,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CompletionResponse represents a completion API response
type CompletionResponse struct {
    ID          string                 `json:"id"`
    Content     string                 `json:"content"`
    Model       string                 `json:"model"`
    Usage       *Usage                 `json:"usage,omitempty"`
    ToolCalls   []ToolCall             `json:"tool_calls,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Message represents a conversation message
type Message struct {
    Role    string `json:"role"`    // "system", "user", "assistant", "tool"
    Content string `json:"content"`
    Name    string `json:"name,omitempty"`
    ToolCallID string `json:"tool_call_id,omitempty"`
}
```

### Configuration Types

```go
// ProviderConfig configures an LLM provider
type ProviderConfig struct {
    Type        string                 `yaml:"type" json:"type"`
    APIKey      string                 `yaml:"api_key" json:"api_key"`
    BaseURL     string                 `yaml:"base_url,omitempty" json:"base_url,omitempty"`
    Model       string                 `yaml:"model" json:"model"`
    Options     map[string]interface{} `yaml:"options,omitempty" json:"options,omitempty"`
}

// AgentConfig configures an agent
type AgentConfig struct {
    Name        string                 `yaml:"name" json:"name"`
    Type        string                 `yaml:"type" json:"type"`
    Provider    string                 `yaml:"provider" json:"provider"`
    Tools       []string               `yaml:"tools,omitempty" json:"tools,omitempty"`
    SystemPrompt string                `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
    Parameters  map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// ToolConfig configures a tool
type ToolConfig struct {
    Name        string                 `yaml:"name" json:"name"`
    Enabled     bool                   `yaml:"enabled" json:"enabled"`
    Timeout     time.Duration          `yaml:"timeout,omitempty" json:"timeout,omitempty"`
    Parameters  map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}
```

## Error Types

```go
// ErrorType categorizes errors
type ErrorType string

const (
    ErrorTypeValidation   ErrorType = "validation"
    ErrorTypeAuth        ErrorType = "authentication"
    ErrorTypeRateLimit   ErrorType = "rate_limit"
    ErrorTypeNetwork     ErrorType = "network"
    ErrorTypeProvider    ErrorType = "provider"
    ErrorTypeInternal    ErrorType = "internal"
)

// Error represents a structured error
type Error struct {
    Type    ErrorType              `json:"type"`
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
    Cause   error                  `json:"-"`
}
```

## Package Import Guide

### Basic Usage

```go
import (
    "github.com/lexlapax/go-llms/pkg/llm"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/agent"
    "github.com/lexlapax/go-llms/pkg/agent/tools"
)
```

### Provider-Specific Imports

```go
import (
    "github.com/lexlapax/go-llms/pkg/llm/provider/openai"
    "github.com/lexlapax/go-llms/pkg/llm/provider/anthropic"
    "github.com/lexlapax/go-llms/pkg/llm/provider/google"
    "github.com/lexlapax/go-llms/pkg/llm/provider/ollama"
)
```

### Utility Imports

```go
import (
    "github.com/lexlapax/go-llms/pkg/util/llmutil"
    "github.com/lexlapax/go-llms/pkg/testutils"
    "github.com/lexlapax/go-llms/pkg/errors"
)
```

## Version Compatibility

This API reference covers Go-LLMs v0.3.5 and later. Key version milestones:

| Version | Release Date | Major Changes |
|---------|--------------|---------------|
| v0.3.5 | June 15, 2025 | Scripting engine integration |
| v0.3.4 | June 13, 2025 | Runtime tool discovery |
| v0.3.3 | January 11, 2025 | Additional providers (Ollama, OpenRouter, Vertex AI) |
| v0.3.2 | January 11, 2025 | Documentation improvements |
| v0.3.1 | January 10, 2025 | Initial stable release |

## API Stability Guarantees

### Stable APIs (v1.0 guarantee)

The following APIs are considered stable and will maintain backward compatibility:

- Core `Provider` interface
- Core `Agent` interface  
- Core `Tool` interface
- Basic request/response types
- Error types and handling

### Experimental APIs

The following APIs are experimental and may change:

- Scripting bridge interfaces
- Advanced workflow patterns
- Custom provider extensions

### Deprecation Policy

- Deprecated APIs will be marked with deprecation notices
- Minimum 2 minor versions before removal
- Migration guides provided for breaking changes

## Usage Examples

### Basic Provider Usage

```go
// Create a provider
provider := provider.NewOpenAIProvider(
    APIKey: "your-api-key",
    Model:  "gpt-4",
}

// Make a completion request
response, err := provider.Complete(ctx, &llm.CompletionRequest{
    Messages: []llm.Message{
        {Role: "user", Content: "Hello, how are you?"},
    },
}
```

### Agent with Tools

```go
// Create an agent
agent := agent.NewToolEnabledAgent(agent.Config{
    Name:     "assistant",
    Provider: provider,
}

// Register tools
agent.RegisterTool(tools.GetTool("http_request"))
agent.RegisterTool(tools.GetTool("json_processor"))

// Execute
result, err := agent.Execute(ctx, "Fetch weather data and parse it")
```

### Workflow Orchestration

```go
// Create workflow
workflow := agent.NewWorkflowAgent(agent.WorkflowConfig{
    Name: "data-pipeline",
}

// Add steps
workflow.AddStep(agent.WorkflowStep{
    Name: "fetch",
    Tool: "http_request",
}

workflow.AddStep(agent.WorkflowStep{
    Name: "process",
    Tool: "json_processor",
    Dependencies: []string{"fetch"},
}

// Execute workflow
result, err := workflow.ExecuteWorkflow(ctx, input)
```

## Best Practices

### Context Usage

Always pass context for cancellation and timeout:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := provider.Complete(ctx, request)
```

### Error Handling

Check error types for appropriate handling:

```go
if err != nil {
    var llmErr *errors.Error
    if errors.As(err, &llmErr) {
        switch llmErr.Type {
        case errors.ErrorTypeRateLimit:
            // Handle rate limiting
        case errors.ErrorTypeAuth:
            // Handle authentication errors
        }
    }
}
```

### Resource Management

Always clean up resources:

```go
tool, err := tools.GetTool("file_processor")
if err != nil {
    return err
}
defer tool.Cleanup(ctx)
```

## Additional Resources

- [User Guide](../../user-guide) - High-level usage documentation
- [Technical Documentation](../../technical) - Architecture and design details
- [Examples](/cmd/examples/) - Complete working examples
- [Contributing Guide](../../technical/development/contributing.md) - Development guidelines

## Support

For questions, bug reports, or feature requests:

- GitHub Issues: [github.com/lexlapax/go-llms/issues](https://github.com/lexlapax/go-llms/issues)
- Documentation: [docs.go-llms.dev](https://docs.go-llms.dev)
- Community: [discord.gg/go-llms](https://discord.gg/go-llms)

---

This API reference is automatically generated from source code. Last updated: June 24, 2025