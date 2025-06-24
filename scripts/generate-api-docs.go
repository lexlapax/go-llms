package main

// ABOUTME: Script to regenerate API documentation for all go-llms packages
// ABOUTME: Generates comprehensive documentation in docs/api directory using pkg/docs tools

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/lexlapax/go-llms/pkg/agent/tools"
	"github.com/lexlapax/go-llms/pkg/docs"
)

// Package represents a package to document
type Package struct {
	Name        string
	Path        string
	Description string
	Category    string
}

var packages = []Package{
	// Core packages
	{
		Name:        "llm",
		Path:        "pkg/llm",
		Description: "Language model provider integration - Unified interface for OpenAI, Anthropic, Google Gemini",
		Category:    "Core",
	},
	{
		Name:        "schema",
		Path:        "pkg/schema",
		Description: "JSON Schema validation - Schema definition, validation, and type coercion",
		Category:    "Core",
	},
	{
		Name:        "structured",
		Path:        "pkg/structured",
		Description: "Extract structured data from LLMs - Prompt enhancement and type-safe output processing",
		Category:    "Core",
	},
	// Agent framework
	{
		Name:        "agent",
		Path:        "pkg/agent",
		Description: "Build autonomous agents - Agent lifecycle, state management, and event-driven architecture",
		Category:    "Agent Framework",
	},
	{
		Name:        "tools",
		Path:        "pkg/agent/tools",
		Description: "Create and manage agent tools - ToolBuilder pattern and agent-tool conversion",
		Category:    "Agent Framework",
	},
	{
		Name:        "builtins",
		Path:        "pkg/agent/builtins",
		Description: "Pre-built tool library - 32+ tools across 7 categories with MCP compatibility",
		Category:    "Agent Framework",
	},
	{
		Name:        "workflows",
		Path:        "pkg/agent/workflow",
		Description: "Compose complex agent behaviors - Sequential, parallel, conditional patterns",
		Category:    "Agent Framework",
	},
	// Utilities
	{
		Name:        "testutils",
		Path:        "pkg/testutils",
		Description: "Testing utilities - Mocks, fixtures, and helpers for testing go-llms applications",
		Category:    "Utilities",
	},
	{
		Name:        "utils",
		Path:        "pkg/util",
		Description: "General utilities - Provider configuration, model info, and helper functions",
		Category:    "Utilities",
	},
}

func main() {
	ctx := context.Background()

	log.Println("Go-LLMs API Documentation Generator")
	log.Println("===================================")

	outputDir := "docs/api"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Generate main README
	if err := generateMainREADME(outputDir); err != nil {
		log.Fatalf("Failed to generate main README: %v", err)
	}

	// Generate documentation for each package
	for _, pkg := range packages {
		log.Printf("Generating documentation for %s...", pkg.Name)
		if err := generatePackageDoc(ctx, pkg, outputDir); err != nil {
			log.Printf("Warning: Failed to generate docs for %s: %v", pkg.Name, err)
		}
	}

	// Generate tool documentation
	log.Println("Generating built-in tools documentation...")
	if err := generateToolsDocumentation(ctx, outputDir); err != nil {
		log.Printf("Warning: Failed to generate tools documentation: %v", err)
	}

	log.Println("\n✅ API documentation generation complete!")
	log.Printf("📁 Documentation saved to: %s\n", outputDir)
}

func generateMainREADME(outputDir string) error {
	content := `# API Reference

This section provides comprehensive API documentation for all go-llms packages, organized by functionality.

## Core Packages

### LLM Integration
- **[LLM API](llm.md)** - Language model provider integration
  - Unified interface for OpenAI, Anthropic, Google Gemini, Ollama, OpenRouter
  - Multi-provider strategies for reliability
  - Streaming and structured generation support

### Data Validation
- **[Schema API](schema.md)** - JSON Schema validation
  - Schema definition and validation
  - Type coercion and custom validators
  - Integration with structured outputs

### Structured Output
- **[Structured API](structured.md)** - Extract structured data from LLMs
  - Prompt enhancement with schemas
  - JSON extraction and validation
  - Type-safe output processing

## Agent Framework

### Core Agent System
- **[Agent API](agent.md)** - Build autonomous agents
  - Agent lifecycle and state management
  - Hook system for monitoring
  - Event-driven architecture

### Tools and Extensions
- **[Tools API](tools.md)** - Create and manage agent tools
  - ToolBuilder pattern for rich metadata
  - Agent-tool bidirectional conversion
  - Performance optimizations

- **[Built-in Tools](builtins.md)** - Pre-built tool library
  - 30+ tools across 7 categories
  - MCP compatibility
  - Tool discovery and registry

### Workflows
- **[Workflow API](workflows.md)** - Compose complex agent behaviors
  - Sequential, parallel, conditional, and loop patterns
  - Error handling and recovery
  - State management across steps

## Utilities

### Testing Support
- **[Test Utilities](testutils.md)** - Testing helpers and mocks
  - Provider mocks for unit testing
  - Fixture management
  - Assertion helpers

### General Utilities
- **[Utilities API](utils.md)** - Common utilities
  - Provider configuration parsing
  - Model information management
  - Error handling utilities

## Quick Links

- [Getting Started Guide](/docs/user-guide/getting-started.md)
- [Examples Directory](/cmd/examples/)
- [Technical Documentation](/docs/technical/)
- [Contributing Guide](/CONTRIBUTING.md)

## API Stability

The APIs documented here follow semantic versioning:
- **Stable APIs** (v0.3.x): May have minor changes but no breaking changes
- **Experimental APIs**: Marked with warnings, may change significantly
- **Deprecated APIs**: Marked with deprecation notices and migration guides

## Documentation Format

Each API document includes:
- **Overview**: High-level description and use cases
- **Core Types**: Main interfaces and structs
- **Functions**: Public functions and methods
- **Examples**: Practical usage examples
- **Best Practices**: Recommended patterns and anti-patterns
- **Error Handling**: Common errors and recovery strategies
`

	return os.WriteFile(filepath.Join(outputDir, "README.md"), []byte(content), 0644)
}

func generatePackageDoc(ctx context.Context, pkg Package, outputDir string) error {
	// Create a template for each package documentation
	content := fmt.Sprintf(`# %s API

%s

## Package Information

- **Import Path**: `+"`github.com/lexlapax/go-llms/%s`"+`
- **Category**: %s
- **Stability**: Stable (v0.3.x)

## Overview

%s

## Core Types

`, strings.ToUpper(pkg.Name[:1])+pkg.Name[1:], pkg.Description, pkg.Path, pkg.Category, getDetailedDescription(pkg.Name))

	// Add package-specific content
	content += getPackageSpecificContent(pkg.Name)

	// Add examples section
	content += `
## Examples

`
	content += getPackageExamples(pkg.Name)

	// Add best practices
	content += `
## Best Practices

`
	content += getPackageBestPractices(pkg.Name)

	// Add error handling
	content += `
## Error Handling

`
	content += getPackageErrorHandling(pkg.Name)

	// Write to file
	filename := filepath.Join(outputDir, pkg.Name+".md")
	return os.WriteFile(filename, []byte(content), 0644)
}

func generateToolsDocumentation(ctx context.Context, outputDir string) error {
	// Initialize tool discovery
	discovery := tools.NewDiscovery()

	// Create documentation generator config
	config := docs.GeneratorConfig{
		Title:           "Go-LLMs Built-in Tools",
		Description:     "Complete reference for all built-in tools",
		Version:         "0.3.5",
		GroupBy:         "category",
		IncludeExamples: true,
		IncludeSchemas:  true,
	}

	// Create integrator
	integrator := docs.NewToolDocumentationIntegrator(discovery, config)

	// Generate markdown documentation
	markdown, err := integrator.GenerateMarkdownForAllTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate tools markdown: %w", err)
	}

	// Write to builtins.md
	filename := filepath.Join(outputDir, "builtins.md")
	return os.WriteFile(filename, []byte(markdown), 0644)
}

func getDetailedDescription(pkgName string) string {
	descriptions := map[string]string{
		"llm": `The LLM package provides a unified interface for interacting with various language model providers including OpenAI, Anthropic, Google (Gemini and Vertex AI), Ollama, and OpenRouter. It abstracts provider-specific differences while exposing common functionality like completions, streaming, and function calling.

Key features:
- Provider abstraction with consistent API
- Automatic retry and error handling
- Token counting and rate limiting
- Streaming support for real-time responses
- Function/tool calling capabilities
- Multi-modal support (text and images)`,

		"schema": `The Schema package implements JSON Schema validation (draft 7) with additional features for type coercion and custom validators. It's designed to work seamlessly with LLM outputs and structured data extraction.

Key features:
- Full JSON Schema draft 7 support
- Type coercion for common conversions
- Custom validator registration
- Schema composition and references
- Integration with structured output parsing
- Performance-optimized validation`,

		"structured": `The Structured package enables reliable extraction of structured data from LLM outputs. It enhances prompts with schema information and validates responses to ensure type-safe results.

Key features:
- Automatic prompt enhancement with schemas
- JSON extraction from free-form text
- Validation against expected schemas
- Retry logic for malformed outputs
- Support for complex nested structures
- Integration with multiple LLM providers`,

		"agent": `The Agent package provides a framework for building autonomous agents that can use tools, maintain state, and execute complex workflows. It supports both simple reactive agents and sophisticated multi-agent systems.

Key features:
- Flexible agent architecture
- Tool integration and management
- State persistence and recovery
- Event-driven lifecycle hooks
- Performance monitoring
- Multi-agent coordination`,

		"tools": `The Tools package defines the interfaces and patterns for creating agent tools. It provides the ToolBuilder pattern for rich metadata, automatic documentation generation, and seamless integration with agents.

Key features:
- ToolBuilder for declarative tool creation
- Automatic schema generation
- Tool discovery and registration
- Metadata and documentation support
- Performance tracking
- MCP (Model Context Protocol) compatibility`,

		"workflows": `The Workflows package enables composition of complex agent behaviors through declarative workflow definitions. It supports various execution patterns and provides robust error handling.

Key features:
- Sequential execution pipelines
- Parallel task execution
- Conditional branching
- Loop constructs
- Error handling and recovery
- State management across steps`,
	}

	if desc, ok := descriptions[pkgName]; ok {
		return desc
	}
	return "Detailed package documentation."
}

func getPackageSpecificContent(pkgName string) string {
	// This would be expanded with actual type definitions
	// For now, returning package-specific templates

	switch pkgName {
	case "llm":
		return getLLMPackageContent()
	case "schema":
		return getSchemaPackageContent()
	case "agent":
		return getAgentPackageContent()
	case "tools":
		return getToolsPackageContent()
	default:
		return getGenericPackageContent()
	}
}

func getLLMPackageContent() string {
	return `### Provider Interface

The core abstraction for all LLM providers:

` + "```go" + `
type Provider interface {
    // Complete generates a completion for the given request
    Complete(ctx context.Context, request *CompletionRequest) (*CompletionResponse, error)
    
    // GetCapabilities returns provider capabilities
    GetCapabilities() Capabilities
    
    // GetModels returns available models
    GetModels(ctx context.Context) ([]Model, error)
    
    // Close cleans up resources
    Close() error
}
` + "```" + `

### Streaming Support

For providers that support streaming:

` + "```go" + `
type StreamingProvider interface {
    Provider
    CompleteStream(ctx context.Context, request *CompletionRequest) (<-chan StreamChunk, error)
}
` + "```" + `

### Provider Factory

Creating providers with the factory pattern:

` + "```go" + `
// Create an OpenAI provider
provider, err := llm.NewProvider("openai", llm.ProviderConfig{
    APIKey: "your-api-key",
    Model: "gpt-4",
})

// Create an Anthropic provider
provider, err := llm.NewProvider("anthropic", llm.ProviderConfig{
    APIKey: "your-api-key",
    Model: "claude-3-opus-20240229",
})
` + "```"
}

func getSchemaPackageContent() string {
	return `### Schema Definition

Define schemas using the Schema type:

` + "```go" + `
type Schema struct {
    Type        string                 ` + "`json:\"type,omitempty\"`" + `
    Properties  map[string]*Schema     ` + "`json:\"properties,omitempty\"`" + `
    Required    []string               ` + "`json:\"required,omitempty\"`" + `
    Title       string                 ` + "`json:\"title,omitempty\"`" + `
    Description string                 ` + "`json:\"description,omitempty\"`" + `
}
` + "```" + `

### Validation

Validate data against schemas:

` + "```go" + `
validator := schema.NewValidator()
err := validator.Validate(data, schemaDefinition)
if err != nil {
    // Handle validation errors
}
` + "```" + `

### Type Coercion

Automatic type conversion during validation:

` + "```go" + `
// Register custom coercion rules
schema.RegisterCoercion(reflect.TypeOf(""), reflect.TypeOf(0), func(v interface{}) (interface{}, error) {
    return strconv.Atoi(v.(string))
})
` + "```"
}

func getAgentPackageContent() string {
	return `### Agent Interface

The core agent abstraction:

` + "```go" + `
type Agent interface {
    // Execute runs the agent with given input
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    
    // GetMetadata returns agent metadata
    GetMetadata() AgentMetadata
    
    // SetConfig updates configuration
    SetConfig(config AgentConfig) error
}
` + "```" + `

### Tool-Enabled Agents

Agents that can use tools:

` + "```go" + `
type ToolEnabledAgent interface {
    Agent
    RegisterTool(tool Tool) error
    ExecuteTool(ctx context.Context, toolName string, input interface{}) (interface{}, error)
}
` + "```" + `

### Creating Agents

` + "```go" + `
// Create a simple LLM agent
agent := agent.NewLLMAgent(agent.Config{
    Provider: provider,
    SystemPrompt: "You are a helpful assistant.",
    Tools: []tools.Tool{
        tools.NewHTTPTool(),
        tools.NewFileTool(),
    },
})
` + "```"
}

func getToolsPackageContent() string {
	return `### Tool Interface

All tools implement this interface:

` + "```go" + `
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, input interface{}) (interface{}, error)
    GetInputSchema() *schema.Schema
    GetOutputSchema() *schema.Schema
}
` + "```" + `

### ToolBuilder Pattern

Create tools with rich metadata:

` + "```go" + `
tool := tools.NewToolBuilder("my-tool").
    WithDescription("Does something useful").
    WithInputSchema(inputSchema).
    WithOutputSchema(outputSchema).
    WithExecutor(func(ctx context.Context, input interface{}) (interface{}, error) {
        // Tool logic here
        return result, nil
    }).
    Build()
` + "```" + `

### Tool Registry

Manage and discover tools:

` + "```go" + `
registry := tools.GetGlobalRegistry()
registry.Register(tool)

// Discover tools by category
webTools := registry.GetByCategory("web")
` + "```"
}

func getGenericPackageContent() string {
	return `### Key Interfaces

This package provides essential interfaces and types for its functionality.

### Core Functions

Main functions exposed by this package.

### Configuration

Configuration options and setup.
`
}

func getPackageExamples(pkgName string) string {
	examples := map[string]string{
		"llm": `### Basic Completion

` + "```go" + `
provider, err := openai.New(openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})

response, err := provider.Complete(ctx, &llm.CompletionRequest{
    Messages: []llm.Message{
        {Role: "user", Content: "Hello, how are you?"},
    },
    Model: "gpt-3.5-turbo",
})
` + "```" + `

### Streaming Response

` + "```go" + `
stream, err := provider.CompleteStream(ctx, request)
for chunk := range stream {
    fmt.Print(chunk.Content)
}
` + "```",

		"schema": `### Define and Validate

` + "```go" + `
personSchema := &schema.Schema{
    Type: "object",
    Properties: map[string]*schema.Schema{
        "name": {Type: "string"},
        "age": {Type: "integer", Minimum: &zero},
    },
    Required: []string{"name"},
}

data := map[string]interface{}{
    "name": "John",
    "age": 30,
}

validator := schema.NewValidator()
err := validator.Validate(data, personSchema)
` + "```",

		"agent": `### Simple Agent

` + "```go" + `
agent := agent.NewSimpleAgent(agent.Config{
    Name: "helper",
    Handler: func(ctx context.Context, input interface{}) (interface{}, error) {
        // Process input
        return "Processed: " + input.(string), nil
    },
})

result, err := agent.Execute(ctx, "Hello")
` + "```",
	}

	if example, ok := examples[pkgName]; ok {
		return example
	}
	return "See the examples directory for usage examples."
}

func getPackageBestPractices(pkgName string) string {
	practices := map[string]string{
		"llm": `1. **Always use context**: Pass context for cancellation and timeouts
2. **Handle rate limits**: Implement exponential backoff for rate limit errors
3. **Monitor token usage**: Track token consumption to manage costs
4. **Use appropriate models**: Choose models based on task complexity
5. **Implement fallbacks**: Use multi-provider strategies for reliability`,

		"schema": `1. **Define schemas upfront**: Create reusable schema definitions
2. **Use references**: Leverage $ref for schema composition
3. **Validate early**: Validate data at system boundaries
4. **Handle coercion carefully**: Be explicit about type conversions
5. **Cache validators**: Reuse compiled validators for performance`,

		"agent": `1. **Keep agents focused**: Single responsibility principle
2. **Use appropriate tools**: Only register necessary tools
3. **Implement error handling**: Graceful degradation for tool failures
4. **Monitor performance**: Track execution time and resource usage
5. **Test thoroughly**: Unit test agent logic and integration test with tools`,
	}

	if practice, ok := practices[pkgName]; ok {
		return practice
	}
	return "Follow Go best practices and the patterns shown in examples."
}

func getPackageErrorHandling(pkgName string) string {
	errorHandling := map[string]string{
		"llm": `Common errors and handling strategies:

` + "```go" + `
response, err := provider.Complete(ctx, request)
if err != nil {
    var apiErr *llm.APIError
    if errors.As(err, &apiErr) {
        switch apiErr.Type {
        case llm.ErrTypeRateLimit:
            // Implement backoff and retry
        case llm.ErrTypeInvalidRequest:
            // Fix request and retry
        case llm.ErrTypeAuthentication:
            // Check API key
        }
    }
}
` + "```",

		"schema": `Validation errors provide detailed information:

` + "```go" + `
err := validator.Validate(data, schema)
if err != nil {
    var validationErr *schema.ValidationError
    if errors.As(err, &validationErr) {
        for _, detail := range validationErr.Details {
            log.Printf("Error at %s: %s", detail.Path, detail.Message)
        }
    }
}
` + "```",

		"agent": `Handle agent execution errors:

` + "```go" + `
result, err := agent.Execute(ctx, input)
if err != nil {
    var agentErr *agent.ExecutionError
    if errors.As(err, &agentErr) {
        log.Printf("Agent %s failed: %s", agentErr.Agent, agentErr.Reason)
        // Implement recovery strategy
    }
}
` + "```",
	}

	if handling, ok := errorHandling[pkgName]; ok {
		return handling
	}
	return "Check error types and implement appropriate recovery strategies."
}
