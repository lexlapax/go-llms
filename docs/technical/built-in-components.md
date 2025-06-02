# Built-in Components Architecture

## Overview and Design Philosophy

The built-in components system provides a rich set of pre-built tools, agents, and workflows that users can leverage immediately while maintaining the flexibility for custom implementations. This architecture enables discovery, composition, and extension of components through a unified registry system.

### Design Principles

1. **Discoverability**: Built-in components are easy to find and understand through search, filtering, and rich metadata
2. **Composability**: Components work well together and can be combined to create complex behaviors
3. **Extensibility**: Users can extend or customize built-ins while following established patterns
4. **Performance**: Built-ins follow established optimization patterns with pooling and caching
5. **Consistency**: All built-ins follow the same patterns as user-defined components

### Key Benefits

- **For Library Users**:
  - Better tool discovery through registry search and filtering
  - Rich documentation with examples and resource requirements
  - Version tracking for compatibility
  - Consistent patterns across all components
  - No need to know exact import paths

- **For Contributors**:
  - Clear contribution guidelines
  - Consistent component patterns
  - Automated registration on import
  - Rich metadata requirements
  - Improved testability

## Architecture

### Registry System

The core of the built-in components architecture is a generic registry system that provides centralized registration and discovery:

```go
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
```

### Metadata System

Every component includes comprehensive metadata for discovery and documentation:

```go
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
```

### Component-Specific Registries

Each component type extends the base registry with specific features:

#### Tool Registry
```go
// ToolMetadata extends base metadata for tools
type ToolMetadata struct {
    builtins.Metadata
    RequiredPermissions []string  // e.g., "file:read", "network:access"
    ResourceUsage      ResourceInfo
}

type ResourceInfo struct {
    Memory      string // e.g., "low", "medium", "high"
    Network     bool   // requires network access
    FileSystem  bool   // requires file system access
    Concurrency bool   // thread-safe for concurrent use
}
```

#### Agent Registry
```go
// AgentMetadata extends base metadata for agents
type AgentMetadata struct {
    builtins.Metadata
    RequiredTools []string // Names of required tools
    OptionalTools []string // Names of optional tools
    Capabilities  []string // What the agent can do
}
```

#### Workflow Registry
```go
// WorkflowMetadata extends base metadata for workflows
type WorkflowMetadata struct {
    builtins.Metadata
    RequiredAgents []string // Required agent templates
    Stages         []string // Workflow stages
}
```

### Auto-Registration Pattern

Components use Go's init() function for automatic registration:

```go
func init() {
    tools.MustRegisterTool("web_fetch", WebFetch(), tools.ToolMetadata{
        Metadata: builtins.Metadata{
            Name:        "web_fetch",
            Category:    "web",
            Tags:        []string{"http", "fetch", "download", "web", "network"},
            Description: "Fetches content from a URL with customizable timeout",
            Version:     "1.0.0",
            Examples:    []builtins.Example{...},
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
```

## Implementation Patterns

### Tool Implementation Pattern

Tools follow a consistent structure with parameter schemas and result types:

```go
// Tool implementation
func WebFetch() domain.Tool {
    return tools.NewTool(
        "web_fetch",
        "Fetches content from a URL",
        webFetchExecute,
        webFetchSchema,
    )
}

// Execution function with typed parameters
func webFetchExecute(ctx context.Context, params WebFetchParams) (*WebFetchResult, error) {
    // Implementation with:
    // - Context awareness
    // - Resource limits
    // - Error handling with context
    // - Security considerations
}
```

### Discovery Patterns

The registry enables multiple discovery patterns:

```go
// Get a specific tool
webFetch, _ := tools.Tools.Get("web_fetch")

// List all tools in a category
webTools := tools.Tools.ListByCategory("web")

// Search for tools by keyword
jsonTools := tools.Tools.Search("json")

// Filter by tags
networkTools := tools.Tools.ListByTags("network", "http")

// Check metadata
if entry, ok := tools.Tools.Get("web_fetch"); ok {
    fmt.Println(entry.Metadata.Description)
    fmt.Println(entry.Metadata.RequiredPermissions)
    fmt.Println(entry.Metadata.ResourceUsage)
}
```

### Enhanced Functionality Pattern

During migration to built-ins, tools are enhanced with:

- Better error handling with contextual messages
- Context awareness for cancellation and timeouts
- Resource limits and usage tracking
- Security considerations (permission model, input validation)
- Improved parameter schemas with validation
- Comprehensive examples and documentation

## Remaining Work

### Phase 3: Agent Templates

Build pre-configured agents that combine tools and prompts for specific use cases:

#### Text Processing Agents
- **TextSummarize**: Intelligent summarization using LLM with configurable detail levels
- **TextExtract**: Extract structured data from unstructured text
- **TextAnalyze**: Sentiment analysis, entity extraction, keyword identification
- **TextTranslate**: Multi-language translation with context preservation

#### Research Agents
- **WebResearcher**: Web research with source tracking and fact verification
  - Required tools: web_fetch, web_search, web_scrape
  - Capabilities: Source synthesis, credibility assessment
- **DocumentAnalyzer**: Analyze documents with structure understanding
  - Required tools: file_read, text_extract
  - Capabilities: Section identification, key point extraction
- **FactChecker**: Verify claims against multiple sources
  - Required tools: web_search, web_fetch
  - Capabilities: Source comparison, confidence scoring

#### Coding Agents
- **CodeReviewer**: Review code for issues and best practices
  - Required tools: file_read, file_list
  - Capabilities: Bug detection, style checking, security analysis
- **TestGenerator**: Generate tests from code
  - Required tools: file_read, file_write
  - Capabilities: Test case generation, edge case identification
- **DocWriter**: Generate documentation
  - Required tools: file_read, file_write
  - Capabilities: API docs, README generation, inline comments

#### Data Agents
- **DataAnalyst**: Analyze datasets and generate insights
  - Required tools: csv_process, json_process, data_transform
  - Capabilities: Statistical analysis, trend identification
- **ReportGenerator**: Create formatted reports
  - Required tools: data_transform, file_write
  - Capabilities: Chart generation, summary statistics
- **DataCleaner**: Clean and validate data
  - Required tools: csv_process, json_process, data_transform
  - Capabilities: Deduplication, format standardization

#### Feed Processing Agents
- **NewsMonitor**: Monitor news feeds for keywords and topics
  - Required tools: feed_fetch, feed_filter
  - Capabilities: Keyword alerting, topic categorization
- **FeedAggregator**: Aggregate and deduplicate content
  - Required tools: feed_fetch, feed_aggregate
  - Capabilities: Multi-source aggregation, duplicate detection
- **FeedSummarizer**: Summarize feed content
  - Required tools: feed_fetch, feed_extract
  - Capabilities: Daily digests, topic summaries
- **ContentCurator**: Curate and categorize feed content
  - Required tools: feed_fetch, feed_filter, feed_extract
  - Capabilities: Quality scoring, relevance ranking

### Phase 4: Workflow Patterns

Implement reusable workflow patterns for multi-agent coordination:

#### Core Patterns
- **Pipeline**: Sequential processing with data transformation between stages
  ```go
  type Pipeline interface {
      AddStage(name string, agent Agent) error
      Execute(ctx context.Context, input any) (any, error)
  }
  ```

- **MapReduce**: Parallel processing with aggregation
  ```go
  type MapReduce interface {
      SetMapper(agent Agent) error
      SetReducer(agent Agent) error
      Execute(ctx context.Context, items []any) (any, error)
  }
  ```

- **Consensus**: Multi-agent agreement with configurable thresholds
  ```go
  type Consensus interface {
      AddAgent(agent Agent, weight float64) error
      SetThreshold(threshold float64) error
      Execute(ctx context.Context, input any) (any, error)
  }
  ```

- **RetryWithFallback**: Resilient execution with exponential backoff
  ```go
  type RetryWithFallback interface {
      SetPrimary(agent Agent) error
      AddFallback(agent Agent) error
      SetRetryPolicy(policy RetryPolicy) error
      Execute(ctx context.Context, input any) (any, error)
  }
  ```

#### Example Workflows

- **ResearchWorkflow**: Research → Verify → Synthesize → Report
  - Combines WebResearcher, FactChecker, and ReportGenerator
  - Implements citation tracking and confidence scoring
  
- **CodeReviewWorkflow**: Analyze → Review → Suggest → Document
  - Combines CodeReviewer, TestGenerator, and DocWriter
  - Provides comprehensive code quality assessment
  
- **DataPipeline**: Ingest → Clean → Analyze → Visualize
  - Combines DataCleaner, DataAnalyst, and ReportGenerator
  - Handles various data formats and quality issues

- **ContentCurationWorkflow**: Discover → Filter → Analyze → Publish
  - Combines FeedAggregator, ContentCurator, and FeedSummarizer
  - Provides automated content curation pipeline

## Design Guidelines for Contributors

### Creating New Tools

1. **Follow Established Patterns**:
   - One tool per file with consistent naming
   - Use typed parameters and results
   - Include comprehensive parameter schemas
   - Implement proper error handling with context

2. **Metadata Requirements**:
   - Descriptive name following naming conventions
   - Appropriate category and tags
   - Clear description of functionality
   - Version starting at "1.0.0"
   - At least one usage example
   - Resource usage declaration
   - Required permissions list

3. **Implementation Guidelines**:
   - Use context for cancellation support
   - Validate all inputs against schema
   - Provide meaningful error messages
   - Consider resource limits
   - Make tools concurrent-safe when possible
   - Follow security best practices

4. **Testing Requirements**:
   - Unit tests with >90% coverage
   - Integration tests for external dependencies
   - Benchmark tests for performance
   - Example tests demonstrating usage

### Creating New Agents

1. **Template Structure**:
   ```go
   type AgentTemplate interface {
       Build(provider domain.Provider, opts ...AgentOption) domain.Agent
       Metadata() AgentMetadata
   }
   ```

2. **Required Elements**:
   - Clear system prompt defining agent behavior
   - List of required and optional tools
   - Capability declarations
   - Configuration options
   - Usage examples

3. **Best Practices**:
   - Keep agents focused on specific domains
   - Provide sensible defaults
   - Allow customization through options
   - Document expected inputs and outputs
   - Include error handling strategies

### Creating New Workflows

1. **Pattern Implementation**:
   - Implement the WorkflowPattern interface
   - Define clear stages and routing
   - Support configuration through options
   - Handle partial failures gracefully

2. **Composability**:
   - Design workflows to be composable
   - Use standard interfaces for stages
   - Support different routing strategies
   - Enable monitoring and logging hooks

3. **Documentation**:
   - Describe the workflow purpose
   - Document each stage's role
   - Provide configuration examples
   - Include performance considerations

## Migration Strategy

For users migrating from common_tools.go to built-in components:

### Import Changes
```go
// Old approach
import "github.com/yourusername/go-llms/pkg/agent/tools"
tool := tools.WebFetch()

// New approach with built-ins
import (
    "github.com/yourusername/go-llms/pkg/agent/builtins/tools"
    _ "github.com/yourusername/go-llms/pkg/agent/builtins/tools/web"
)
tool := tools.MustGetTool("web_fetch")
```

### Discovery Benefits
```go
// List all available tools
allTools := tools.Tools.List()

// Find tools by category
webTools := tools.Tools.ListByCategory("web")

// Search for tools
fileTools := tools.Tools.Search("file")

// Examine metadata
entry, _ := tools.Tools.Get("web_fetch")
fmt.Printf("Permissions: %v\n", entry.Metadata.RequiredPermissions)
fmt.Printf("Resource Usage: %v\n", entry.Metadata.ResourceUsage)
```

## Performance Considerations

1. **Lazy Loading**: Components are only initialized when accessed
2. **Registry Caching**: Frequently accessed components are cached
3. **Object Pooling**: All built-ins follow established pooling patterns
4. **Minimal Dependencies**: Each category can be imported independently
5. **Efficient Search**: Registry uses optimized search algorithms

## Security Considerations

1. **Permission Model**: Tools declare required permissions for transparency
2. **Input Validation**: All tools validate inputs against schemas
3. **Resource Limits**: Tools respect configured resource limits
4. **Sandboxing**: System tools run with appropriate restrictions
5. **Safe Defaults**: Security-conscious default configurations

## Future Enhancements

1. **Component Versioning**: Support for multiple versions of components
2. **Component Dependencies**: Declare and resolve component dependencies
3. **Remote Components**: Load components from external sources
4. **Component Marketplace**: Community-contributed components
5. **Visual Workflow Builder**: GUI for composing workflows
6. **Performance Profiling**: Built-in profiling for components
7. **A/B Testing**: Support for comparing component variations