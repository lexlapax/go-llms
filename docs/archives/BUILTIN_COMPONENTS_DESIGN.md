# Built-in Components Design Document

## Overview

This document outlines the design for exposing built-in tools, agents, and workflows in the go-llms library. The goal is to provide a rich set of pre-built components that users can leverage immediately while maintaining the flexibility for custom implementations.

## Design Principles

1. **Discoverability**: Built-in components should be easy to find and understand
2. **Composability**: Components should work well together
3. **Extensibility**: Users should be able to extend or customize built-ins
4. **Performance**: Built-ins should follow established optimization patterns
5. **Consistency**: All built-ins should follow the same patterns as user-defined components

## Proposed Structure

### Directory Organization

```
pkg/agent/builtins/
├── registry.go              # Central registry interfaces
├── tools/
│   ├── categories.go        # Tool categorization system
│   ├── registry.go          # Tool-specific registry
│   ├── web/                 # Web-related tools
│   │   ├── fetch.go         # HTTP fetching
│   │   ├── scrape.go        # Web scraping
│   │   └── search.go        # Web search
│   ├── file/                # File system tools
│   │   ├── read.go          # File reading
│   │   ├── write.go         # File writing
│   │   ├── list.go          # Directory listing
│   │   └── search.go        # File search
│   ├── data/                # Data processing tools
│   │   ├── json.go          # JSON manipulation
│   │   ├── csv.go           # CSV processing
│   │   ├── xml.go           # XML processing
│   │   └── transform.go     # Data transformation
│   ├── text/                # Text processing tools
│   │   ├── summarize.go     # Text summarization
│   │   ├── extract.go       # Information extraction
│   │   └── translate.go     # Translation
│   └── system/              # System tools
│       ├── exec.go          # Command execution
│       ├── env.go           # Environment variables
│       └── time.go          # Time/date operations
├── agents/
│   ├── registry.go          # Agent-specific registry
│   ├── templates.go         # Common agent configurations
│   ├── research/            # Research-focused agents
│   │   ├── web_researcher.go
│   │   ├── doc_analyzer.go
│   │   └── fact_checker.go
│   ├── coding/              # Code-related agents
│   │   ├── code_reviewer.go
│   │   ├── test_generator.go
│   │   ├── doc_writer.go
│   │   └── bug_finder.go
│   ├── data/                # Data processing agents
│   │   ├── data_analyst.go
│   │   ├── report_generator.go
│   │   └── visualizer.go
│   └── creative/            # Creative agents
│       ├── writer.go
│       ├── brainstormer.go
│       └── storyteller.go
└── workflows/
    ├── registry.go          # Workflow-specific registry
    ├── builder.go           # Workflow builder utilities
    ├── patterns/            # Common workflow patterns
    │   ├── map_reduce.go    # Map-reduce pattern
    │   ├── pipeline.go      # Sequential pipeline
    │   ├── consensus.go     # Multi-agent consensus
    │   └── retry.go         # Retry with fallback
    └── examples/            # Example workflows
        ├── research_workflow.go
        ├── code_review_workflow.go
        └── data_pipeline_workflow.go
```

## Core Interfaces

### Registry Interface

```go
package builtins

// Registry provides a central registration and discovery mechanism
type Registry[T any] interface {
    // Register adds a component to the registry
    Register(name string, component T, metadata Metadata) error
    
    // Get retrieves a component by name
    Get(name string) (T, bool)
    
    // List returns all registered components
    List() []RegistryEntry[T]
    
    // ListByCategory returns components in a specific category
    ListByCategory(category string) []RegistryEntry[T]
    
    // ListByTags returns components matching all provided tags
    ListByTags(tags ...string) []RegistryEntry[T]
    
    // Search returns components matching the query
    Search(query string) []RegistryEntry[T]
}

// Metadata describes a registered component
type Metadata struct {
    Name        string
    Category    string
    Tags        []string
    Description string
    Version     string
    Examples    []Example
    Deprecated  bool
    Experimental bool
}

// Example shows how to use a component
type Example struct {
    Name        string
    Description string
    Code        string
}

// RegistryEntry combines a component with its metadata
type RegistryEntry[T any] struct {
    Component T
    Metadata  Metadata
}
```

### Tool Registry Extensions

```go
package tools

// ToolMetadata extends base metadata for tools
type ToolMetadata struct {
    builtins.Metadata
    RequiredPermissions []string  // e.g., "file:read", "network:access"
    ResourceUsage      ResourceInfo
}

// ResourceInfo describes resource requirements
type ResourceInfo struct {
    Memory      string // e.g., "low", "medium", "high"
    Network     bool   // requires network access
    FileSystem  bool   // requires file system access
    Concurrency bool   // thread-safe for concurrent use
}
```

### Agent Templates

```go
package agents

// AgentTemplate provides a pre-configured agent
type AgentTemplate interface {
    // Build creates a new agent instance with the template's configuration
    Build(provider domain.Provider, opts ...AgentOption) domain.Agent
    
    // Metadata returns information about this template
    Metadata() AgentMetadata
}

// AgentMetadata extends base metadata for agents
type AgentMetadata struct {
    builtins.Metadata
    RequiredTools []string // Names of required tools
    OptionalTools []string // Names of optional tools
    Capabilities  []string // What the agent can do
}

// AgentOption allows customization of template agents
type AgentOption func(*agentConfig)
```

### Workflow Patterns

```go
package workflows

// WorkflowPattern defines a reusable workflow structure
type WorkflowPattern interface {
    // Build creates a workflow instance
    Build(opts ...WorkflowOption) Workflow
    
    // Metadata returns information about this pattern
    Metadata() WorkflowMetadata
}

// Workflow represents an executable workflow
type Workflow interface {
    // Execute runs the workflow
    Execute(ctx context.Context, input any) (any, error)
    
    // AddAgent adds an agent to the workflow
    AddAgent(name string, agent domain.Agent) error
    
    // SetRouter configures routing between agents
    SetRouter(router Router) error
}

// Router determines how data flows between agents
type Router interface {
    Route(ctx context.Context, from string, result any) (string, any, error)
}
```

## Implementation Examples

### Built-in Tool Example

```go
package web

import (
    "github.com/yourusername/go-llms/pkg/agent/builtins"
    "github.com/yourusername/go-llms/pkg/agent/domain"
    "github.com/yourusername/go-llms/pkg/agent/tools"
)

func init() {
    // Auto-register on import
    builtins.Tools.Register("web_fetch", WebFetch(), tools.ToolMetadata{
        Metadata: builtins.Metadata{
            Name:        "web_fetch",
            Category:    "web",
            Tags:        []string{"http", "fetch", "download"},
            Description: "Fetches content from a URL",
            Version:     "1.0.0",
            Examples: []builtins.Example{
                {
                    Name:        "Basic fetch",
                    Description: "Fetch a web page",
                    Code:        `WebFetch().Execute(ctx, map[string]any{"url": "https://example.com"})`,
                },
            },
        },
        RequiredPermissions: []string{"network:access"},
        ResourceUsage: tools.ResourceInfo{
            Memory:      "low",
            Network:     true,
            FileSystem:  false,
            Concurrency: true,
        },
    })
}

// WebFetch returns a tool that fetches web content
func WebFetch() domain.Tool {
    return tools.NewTool(
        "web_fetch",
        "Fetches content from a URL",
        webFetchExecute,
        webFetchSchema,
    )
}
```

### Built-in Agent Example

```go
package research

import (
    "github.com/yourusername/go-llms/pkg/agent/builtins"
    "github.com/yourusername/go-llms/pkg/agent/builtins/agents"
    "github.com/yourusername/go-llms/pkg/agent/domain"
)

func init() {
    builtins.Agents.Register("web_researcher", &WebResearcherTemplate{}, agents.AgentMetadata{
        Metadata: builtins.Metadata{
            Name:        "web_researcher",
            Category:    "research",
            Tags:        []string{"research", "web", "analysis"},
            Description: "Agent specialized in web research and fact-checking",
            Version:     "1.0.0",
        },
        RequiredTools: []string{"web_fetch", "web_search"},
        OptionalTools: []string{"summarize", "fact_check"},
        Capabilities:  []string{
            "Web research",
            "Fact verification",
            "Source synthesis",
            "Report generation",
        },
    })
}

type WebResearcherTemplate struct{}

func (t *WebResearcherTemplate) Build(provider domain.Provider, opts ...agents.AgentOption) domain.Agent {
    agent := workflow.NewAgent(provider)
    
    // Configure with research-specific prompt
    agent.SetSystemPrompt(researchSystemPrompt)
    
    // Add required tools
    agent.AddTool(builtins.Tools.MustGet("web_fetch"))
    agent.AddTool(builtins.Tools.MustGet("web_search"))
    
    // Apply custom options
    cfg := &agentConfig{}
    for _, opt := range opts {
        opt(cfg)
    }
    
    // Add optional tools if requested
    if cfg.includeAllTools {
        agent.AddTool(builtins.Tools.MustGet("summarize"))
        agent.AddTool(builtins.Tools.MustGet("fact_check"))
    }
    
    return agent
}
```

### Built-in Workflow Example

```go
package examples

import (
    "github.com/yourusername/go-llms/pkg/agent/builtins"
    "github.com/yourusername/go-llms/pkg/agent/builtins/workflows"
)

func init() {
    builtins.Workflows.Register("research_pipeline", &ResearchPipelinePattern{}, workflows.WorkflowMetadata{
        Metadata: builtins.Metadata{
            Name:        "research_pipeline",
            Category:    "research",
            Tags:        []string{"research", "pipeline", "multi-agent"},
            Description: "Multi-stage research workflow with fact-checking",
            Version:     "1.0.0",
        },
        RequiredAgents: []string{"web_researcher", "fact_checker", "report_writer"},
        Stages: []string{"research", "verify", "synthesize", "report"},
    })
}

type ResearchPipelinePattern struct{}

func (p *ResearchPipelinePattern) Build(opts ...workflows.WorkflowOption) workflows.Workflow {
    builder := workflows.NewBuilder()
    
    // Define stages
    builder.AddStage("research", builtins.Agents.MustGet("web_researcher"))
    builder.AddStage("verify", builtins.Agents.MustGet("fact_checker"))
    builder.AddStage("synthesize", builtins.Agents.MustGet("synthesizer"))
    builder.AddStage("report", builtins.Agents.MustGet("report_writer"))
    
    // Define routing
    builder.SetRouter(workflows.SequentialRouter())
    
    // Apply options
    for _, opt := range opts {
        opt(builder)
    }
    
    return builder.Build()
}
```

## Usage Examples

### Using Built-in Tools

```go
// Get a specific tool
webFetch, _ := builtins.Tools.Get("web_fetch")

// List all web-related tools
webTools := builtins.Tools.ListByCategory("web")

// Search for tools
searchTools := builtins.Tools.Search("json")

// Use with an agent
agent := workflow.NewAgent(provider)
agent.AddTool(webFetch)
```

### Using Built-in Agents

```go
// Create a research agent
researcherTemplate, _ := builtins.Agents.Get("web_researcher")
researcher := researcherTemplate.Build(provider)

// Customize with options
researcher := researcherTemplate.Build(provider,
    agents.WithAdditionalTools("calculator", "translator"),
    agents.WithCustomPrompt("Focus on scientific papers"),
)

// Use the agent
result, _ := researcher.Run(ctx, "Research quantum computing applications")
```

### Using Built-in Workflows

```go
// Get a workflow pattern
pipelinePattern, _ := builtins.Workflows.Get("research_pipeline")
pipeline := pipelinePattern.Build()

// Execute the workflow
result, _ := pipeline.Execute(ctx, map[string]any{
    "topic": "artificial intelligence in healthcare",
    "depth": "comprehensive",
})
```

## Discovery and Documentation

### CLI Discovery

```go
// Add to cmd/main.go
func listBuiltins(cmd *cobra.Command, args []string) {
    fmt.Println("Built-in Tools:")
    for _, entry := range builtins.Tools.List() {
        fmt.Printf("  %s (%s): %s\n", 
            entry.Metadata.Name,
            entry.Metadata.Category,
            entry.Metadata.Description)
    }
    
    fmt.Println("\nBuilt-in Agents:")
    for _, entry := range builtins.Agents.List() {
        fmt.Printf("  %s (%s): %s\n",
            entry.Metadata.Name,
            entry.Metadata.Category,
            entry.Metadata.Description)
    }
}
```

### Programmatic Discovery

```go
// Find all research-related components
researchTools := builtins.Tools.ListByTags("research")
researchAgents := builtins.Agents.ListByCategory("research")
researchWorkflows := builtins.Workflows.ListByTags("research")

// Get detailed information
if tool, ok := builtins.Tools.Get("web_fetch"); ok {
    metadata := tool.Metadata
    fmt.Printf("Tool: %s v%s\n", metadata.Name, metadata.Version)
    fmt.Printf("Permissions: %v\n", metadata.RequiredPermissions)
    
    // Show examples
    for _, example := range metadata.Examples {
        fmt.Printf("\nExample: %s\n%s\n", example.Name, example.Code)
    }
}
```

## Migration Strategy

1. **Phase 1**: Implement registry infrastructure
   - Core registry interfaces
   - Metadata structures
   - Registration mechanisms

2. **Phase 2**: Migrate existing tools
   - Move WebFetch to built-ins
   - Add metadata and examples
   - Maintain backward compatibility

3. **Phase 3**: Create initial built-in set
   - Essential tools for each category
   - Basic agent templates
   - Simple workflow patterns

4. **Phase 4**: Expand and refine
   - Add more specialized components
   - Gather user feedback
   - Optimize performance

## Backward Compatibility

- All existing APIs remain unchanged
- Built-ins use the same interfaces as user-defined components
- Optional import - users only get built-ins they explicitly import
- Version tagging allows for controlled evolution

## Performance Considerations

1. **Lazy Loading**: Components are only initialized when accessed
2. **Registry Caching**: Frequently accessed components are cached
3. **Pooling**: All built-ins follow established pooling patterns
4. **Minimal Dependencies**: Each category can be imported independently

## Security Considerations

1. **Permission Model**: Tools declare required permissions
2. **Sandboxing**: System tools run in restricted contexts
3. **Input Validation**: All tools validate inputs against schemas
4. **Resource Limits**: Tools respect configured resource limits

## Next Steps

1. Implement core registry infrastructure
2. Create initial tool categories with 2-3 tools each
3. Build basic agent templates for common use cases
4. Design workflow builder API
5. Add comprehensive examples and documentation
6. Gather feedback and iterate