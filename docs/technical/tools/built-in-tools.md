# Built-in Tools: Technical Reference and Architecture

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [Tools](../../technical/tools) / Built-in Tools**

Technical reference for built-in tools architecture, interfaces, registration system, and advanced integration patterns. For user-focused tool documentation, see [Built-in Tools Reference](../../user-guide/reference/built-in-tools-reference.md).

## Tool System Architecture

Go-LLMs provides a comprehensive set of built-in tools that enable LLM agents to interact with various systems and perform complex operations. All tools follow a consistent interface and provide:

- **Structured Parameters**: JSON Schema-based parameter validation
- **Type Safety**: Strongly typed inputs and outputs
- **Error Handling**: Comprehensive error guidance for LLMs
- **Event Support**: Progress tracking and custom events
- **State Integration**: Context-aware execution with state management
- **MCP Compatibility**: Export to Model Context Protocol format

## Tool Interface

All tools implement the `domain.Tool` interface:

```go
type Tool interface {
    // Core functionality
    Name() string
    Description() string
    Execute(ctx *ToolContext, params interface{}) (interface{}, error)
    
    // Schema definitions
    ParameterSchema() *ssdomain.Schema
    OutputSchema() *ssdomain.Schema
    
    // LLM guidance
    UsageInstructions() string
    Examples() []ToolExample
    Constraints() []string
    ErrorGuidance() map[string]string
    
    // Metadata
    Category() string
    Tags() []string
    Version() string
    
    // Behavioral hints
    IsDeterministic() bool
    IsDestructive() bool
    RequiresConfirmation() bool
    EstimatedLatency() string
    
    // MCP compatibility
    ToMCPDefinition() MCPToolDefinition
}
```

## Tool Discovery and Registration

### Registry System

Tools are registered in a global registry with metadata:

```go
// Access the global tool registry
registry := tools.Tools

// Register a new tool
tools.MustRegisterTool("tool_name", tool, tools.ToolMetadata{
    Metadata: builtins.Metadata{
        Name:        "tool_name",
        Category:    "category",
        Tags:        []string{"tag1", "tag2"},
        Description: "Tool description",
        Version:     "1.0.0",
    },
    RequiredPermissions: []string{"permission:action"},
    ResourceUsage: tools.ResourceInfo{
        Memory:      "low", // low, medium, high
        Network:     false,
        FileSystem:  true,
        Concurrency: true,
    },
})
```

### Discovery Methods

```go
// List all tools
allTools := registry.List()

// Get tool by name
tool, found := registry.Get("tool_name")

// Filter by category
categoryTools := registry.ListByCategory("file")

// Filter by permission
permTools := registry.ListByPermission("file:read")

// Filter by resource usage
lowMemTools := registry.ListByResourceUsage(tools.ResourceCriteria{
    MaxMemory: "low",
})
```

## Tool Categories Overview

Go-LLMs includes 30+ built-in tools organized into logical categories:

| Category | Tools | Description |
|----------|-------|-------------|
| [File System](#file-system-tools) | 10 tools | File operations, directory management, search |
| [Web & HTTP](#web-http-tools) | 8 tools | HTTP requests, web scraping, API clients |
| [System](#system-tools) | 7 tools | Process management, environment, system info |
| [Data Processing](#data-processing-tools) | 10 tools | JSON, XML, CSV, template rendering, validation |
| [Math & Computation](#math-computation-tools) | 5 tools | Mathematical operations, statistical analysis |
| [Date & Time](#date-time-tools) | 4 tools | Date parsing, formatting, timezone operations |
| [Text Processing](#text-processing-tools) | 6 tools | String manipulation, regex, encoding |

## Tool Integration Patterns

### Safe Tool Execution

```go
func executeToolSafely(tool domain.Tool, ctx *domain.ToolContext, params interface{}) (interface{}, error) {
    // Pre-execution validation
    if err := validateParams(tool, params); err != nil {
        return nil, fmt.Errorf("invalid parameters: %w", err)
    }
    
    // Execute with timeout
    ctx, cancel := context.WithTimeout(ctx.Context, 30*time.Second)
    defer cancel()
    
    // Execute tool
    result, err := tool.Execute(ctx, params)
    if err != nil {
        // Check error guidance
        for errType, guidance := range tool.ErrorGuidance() {
            if strings.Contains(err.Error(), errType) {
                return nil, fmt.Errorf("%w\nGuidance: %s", err, guidance)
            }
        }
        return nil, err
    }
    
    return result, nil
}
```

### Tool Chaining Pattern

```go
func processDataPipeline(ctx *domain.ToolContext) error {
    // Step 1: Read file
    readTool := file.ReadFile()
    data, err := readTool.Execute(ctx, map[string]interface{}{
        "path": "/data/input.json",
    })
    if err != nil {
        return err
    }
    
    // Step 2: Process JSON
    jsonTool := data.JSONProcess()
    processed, err := jsonTool.Execute(ctx, data.JSONProcessInput{
        Data: data.(*file.ReadFileResult).Content,
        Operation: "transform",
        Transform: "flatten",
    })
    if err != nil {
        return err
    }
    
    // Step 3: Write result
    writeTool := file.WriteFile()
    _, err = writeTool.Execute(ctx, map[string]interface{}{
        "path": "/data/output.json",
        "content": processed.(*data.JSONProcessOutput).Result.(string),
    })
    
    return err
}
```

### Conditional Tool Selection

```go
func selectToolForTask(task string) domain.Tool {
    switch {
    case strings.Contains(task, "calculate"):
        return math.Calculator()
    case strings.Contains(task, "read") && strings.Contains(task, "file"):
        return file.ReadFile()
    case strings.Contains(task, "fetch") && strings.Contains(task, "web"):
        return web.WebFetch()
    case strings.Contains(task, "current time"):
        return datetime.DateTimeNow()
    default:
        return nil
    }
}
```

### Tool Registry Extension

```go
// Create custom tool
type CustomTool struct {
    atools.BaseTool
}

func (t *CustomTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
    // Custom implementation
    return nil, nil
}

// Register custom tool
func RegisterCustomTools() {
    tools.MustRegisterTool("custom_tool", &CustomTool{
        BaseTool: atools.NewToolBuilder("custom_tool", "Custom tool description").
            WithCategory("custom").
            WithTags([]string{"custom", "example"}).
            Build().(*atools.BaseTool),
    }, tools.ToolMetadata{
        Metadata: builtins.Metadata{
            Name:        "custom_tool",
            Category:    "custom",
            Description: "Custom tool for specialized tasks",
            Version:     "1.0.0",
        },
    })
}
```

## Advanced Tool Patterns

### Tool Composition

```go
type ComposedTool struct {
    tools []domain.Tool
}

func (c *ComposedTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
    var results []interface{}
    
    for _, tool := range c.tools {
        result, err := tool.Execute(ctx, params)
        if err != nil {
            return nil, fmt.Errorf("tool %s failed: %w", tool.Name(), err)
        }
        results = append(results, result)
    }
    
    return results, nil
}
```

### Tool Factory Pattern

```go
type ToolFactory interface {
    CreateTool(name string, config map[string]interface{}) (domain.Tool, error)
}

type DefaultToolFactory struct{}

func (f *DefaultToolFactory) CreateTool(name string, config map[string]interface{}) (domain.Tool, error) {
    switch name {
    case "file_reader":
        return &FileReadTool{config: config}, nil
    case "web_fetcher":
        return &WebFetchTool{config: config}, nil
    default:
        return nil, fmt.Errorf("unknown tool: %s", name)
    }
}
```

### Tool Middleware

```go
type ToolMiddleware func(domain.Tool) domain.Tool

func WithLogging(tool domain.Tool) domain.Tool {
    return &LoggingTool{
        Tool: tool,
    }
}

type LoggingTool struct {
    domain.Tool
}

func (l *LoggingTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
    log.Printf("Executing tool: %s", l.Tool.Name())
    start := time.Now()
    
    result, err := l.Tool.Execute(ctx, params)
    
    log.Printf("Tool %s completed in %v", l.Tool.Name(), time.Since(start))
    return result, err
}
```

## Tool Categories Implementation

### File System Tools

Core file operations with security and performance considerations:

- **file_read**: Reads files with size limits and encoding support
- **file_write**: Atomic writes with backup and append modes
- **file_list**: Directory listings with filtering and recursion
- **file_search**: Content and metadata-based file search
- **file_move**: Safe file/directory movement operations
- **file_copy**: File duplication with integrity checks
- **file_delete**: Safe deletion with confirmation requirements

### Web & HTTP Tools

Network operations with rate limiting and error handling:

- **web_fetch**: HTTP requests with comprehensive options
- **web_scrape**: HTML parsing and data extraction
- **api_client**: RESTful API interactions with authentication
- **graphql_client**: GraphQL query execution
- **openapi_discovery**: Automatic API discovery and configuration

### System Tools

System interaction with security constraints:

- **system_execute**: Command execution with sandboxing
- **process_list**: Process monitoring and management
- **env_vars**: Environment variable access
- **system_info**: System resource information

### Data Processing Tools

Format conversion and validation:

- **json_process**: JSON parsing, querying, and transformation
- **csv_process**: CSV reading, writing, and analysis
- **xml_process**: XML parsing and manipulation
- **template_render**: Template processing with data binding

### Math & Computation Tools

Mathematical operations for LLM assistance:

- **calculator**: Basic and advanced mathematical operations
- **statistics**: Statistical analysis and calculations
- **data_analysis**: Dataset processing and insights

### Date & Time Tools

Temporal operations and formatting:

- **datetime_now**: Current time with timezone support
- **datetime_parse**: Flexible date parsing
- **datetime_format**: Localized date formatting
- **datetime_calculate**: Date arithmetic operations

## Performance Considerations

### Tool Execution Optimization

```go
// Pool tool instances for reuse
var toolPool = sync.Pool{
    New: func() interface{} {
        return &FileReadTool{}
    },
}

func getFileReadTool() *FileReadTool {
    return toolPool.Get().(*FileReadTool)
}

func putFileReadTool(tool *FileReadTool) {
    tool.Reset() // Clean state
    toolPool.Put(tool)
}
```

### Concurrent Tool Execution

```go
func executeToolsConcurrently(tools []domain.Tool, ctx *domain.ToolContext, params interface{}) ([]interface{}, error) {
    results := make([]interface{}, len(tools))
    errors := make([]error, len(tools))
    
    var wg sync.WaitGroup
    for i, tool := range tools {
        wg.Add(1)
        go func(idx int, t domain.Tool) {
            defer wg.Done()
            results[idx], errors[idx] = t.Execute(ctx, params)
        }(i, tool)
    }
    
    wg.Wait()
    
    // Check for errors
    for _, err := range errors {
        if err != nil {
            return nil, err
        }
    }
    
    return results, nil
}
```

### Resource Management

```go
type ResourceManager struct {
    semaphore chan struct{}
    metrics   *ToolMetrics
}

func NewResourceManager(maxConcurrent int) *ResourceManager {
    return &ResourceManager{
        semaphore: make(chan struct{}, maxConcurrent),
        metrics:   NewToolMetrics(),
    }
}

func (rm *ResourceManager) Execute(tool domain.Tool, ctx *domain.ToolContext, params interface{}) (interface{}, error) {
    // Acquire resource
    rm.semaphore <- struct{}{}
    defer func() { <-rm.semaphore }()
    
    // Track metrics
    start := time.Now()
    defer func() {
        rm.metrics.RecordExecution(tool.Name(), time.Since(start))
    }()
    
    return tool.Execute(ctx, params)
}
```

## Testing Tools

### Mock Tool Implementation

```go
type MockTool struct {
    name     string
    responses map[string]interface{}
    errors   map[string]error
}

func NewMockTool(name string) *MockTool {
    return &MockTool{
        name:      name,
        responses: make(map[string]interface{}),
        errors:    make(map[string]error),
    }
}

func (m *MockTool) SetResponse(input string, response interface{}) {
    m.responses[input] = response
}

func (m *MockTool) SetError(input string, err error) {
    m.errors[input] = err
}

func (m *MockTool) Execute(ctx *domain.ToolContext, params interface{}) (interface{}, error) {
    key := fmt.Sprintf("%v", params)
    
    if err, exists := m.errors[key]; exists {
        return nil, err
    }
    
    if response, exists := m.responses[key]; exists {
        return response, nil
    }
    
    return nil, fmt.Errorf("no mock response for: %s", key)
}
```

### Tool Testing Utilities

```go
func TestToolWithTimeout(t *testing.T, tool domain.Tool, params interface{}, timeout time.Duration) {
    ctx := &domain.ToolContext{
        Context: context.Background(),
        State:   domain.NewState(),
        Events:  domain.NewEventEmitter(),
    }
    
    done := make(chan struct{})
    var result interface{}
    var err error
    
    go func() {
        result, err = tool.Execute(ctx, params)
        close(done)
    }()
    
    select {
    case <-done:
        assert.NoError(t, err)
        assert.NotNil(t, result)
    case <-time.After(timeout):
        t.Fatalf("Tool execution timed out after %v", timeout)
    }
}
```

## Best Practices

### Tool Design Principles

1. **Single Responsibility**: Each tool should have one clear purpose
2. **Idempotent Operations**: Tools should be safe to retry
3. **Error Guidance**: Provide helpful error messages for LLMs
4. **Resource Limits**: Implement appropriate timeouts and size limits
5. **State Independence**: Tools should not rely on external state

### Security Considerations

1. **Input Validation**: Always validate and sanitize inputs
2. **Permission Checks**: Verify required permissions before execution
3. **Resource Limits**: Prevent resource exhaustion attacks
4. **Safe Defaults**: Use secure defaults for all configurations
5. **Audit Logging**: Log all tool executions for security monitoring

### Performance Guidelines

1. **Lazy Loading**: Load resources only when needed
2. **Connection Pooling**: Reuse network connections
3. **Caching**: Cache expensive operations appropriately
4. **Streaming**: Use streaming for large data operations
5. **Graceful Degradation**: Handle partial failures gracefully

## Migration and Compatibility

### Version Management

```go
type VersionedTool struct {
    domain.Tool
    version string
}

func (v *VersionedTool) Version() string {
    return v.version
}

func (v *VersionedTool) IsCompatible(requiredVersion string) bool {
    return compareVersions(v.version, requiredVersion) >= 0
}
```

### Tool Registry Migration

```go
func MigrateToolRegistry(oldRegistry, newRegistry *ToolRegistry) error {
    for name, tool := range oldRegistry.List() {
        if newTool, exists := newRegistry.Get(name); exists {
            if !newTool.IsCompatible(tool.Version()) {
                return fmt.Errorf("incompatible tool version: %s", name)
            }
        } else {
            // Register legacy tool
            newRegistry.Register(name, tool)
        }
    }
    return nil
}
```

## Next Steps

- Explore [Tool Discovery](tool-discovery.md) for runtime tool management
- See [Creating Tools](creating-tools.md) for custom tool development
- Check [Tool Overview](overview.md) for architectural concepts
- Review tool examples in `/examples/builtins-*/`