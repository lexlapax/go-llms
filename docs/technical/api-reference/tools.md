# Tool Interface Documentation

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [API Reference](../../technical/api-reference) / Tools**

Complete API reference for tool interfaces and implementations in Go-LLMs, covering core tool interfaces, built-in tools, tool registry, metadata systems, execution patterns, and tool lifecycle management.

## Core Tool Interfaces

### Tool Interface

The base interface that all tools must implement:

```go
package tools

// Tool defines the interface for all tools in the system
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
    ValidateOutput(output interface{}) error
    
    // Lifecycle
    Initialize(ctx context.Context) error
    Cleanup(ctx context.Context) error
    
    // Metadata
    GetMetadata() ToolMetadata
    GetCapabilities() ToolCapabilities
}
```

#### Methods

##### Name

```go
Name() string
```

Returns the unique name identifier for the tool.

**Returns:**
- `string`: Tool name (e.g., "http_request", "file_reader")

**Example:**
```go
name := tool.Name() // "http_request"
```

##### GetInputSchema

```go
GetInputSchema() *jsonschema.Schema
```

Returns the JSON Schema defining valid input for the tool.

**Returns:**
- `*jsonschema.Schema`: Input schema definition

**Example:**
```go
schema := tool.GetInputSchema()
// Use schema to validate input before execution
```

##### Execute

```go
Execute(ctx context.Context, input interface{}) (interface{}, error)
```

Executes the tool with the provided input.

**Parameters:**
- `ctx`: Context for cancellation and timeout control
- `input`: The input data (must match input schema)

**Returns:**
- `interface{}`: Tool execution result
- `error`: Error if execution fails

**Example:**
```go
result, err := tool.Execute(ctx, map[string]interface{}{
    "url": "https://api.example.com/data",
    "method": "GET",
}
```

### StatefulTool Interface

Extends Tool with state management capabilities:

```go
// StatefulTool maintains state across executions
type StatefulTool interface {
    Tool
    
    // State management
    GetState() ToolState
    SetState(state ToolState) error
    ResetState() error
    
    // State persistence
    SaveState(ctx context.Context) error
    LoadState(ctx context.Context) error
    
    // State versioning
    GetStateVersion() int64
    MigrateState(fromVersion int64) error
}

// ToolState represents tool state
type ToolState interface {
    // Basic operations
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}) error
    Delete(key string) error
    
    // Bulk operations
    GetAll() map[string]interface{}
    Clear() error
    
    // Metadata
    Version() int64
    Modified() time.Time
}
```

### AsyncTool Interface

For tools with asynchronous execution:

```go
// AsyncTool supports asynchronous execution
type AsyncTool interface {
    Tool
    
    // Async execution
    ExecuteAsync(ctx context.Context, input interface{}) (TaskID, error)
    
    // Status checking
    GetStatus(ctx context.Context, taskID TaskID) (TaskStatus, error)
    GetResult(ctx context.Context, taskID TaskID) (interface{}, error)
    
    // Cancellation
    Cancel(ctx context.Context, taskID TaskID) error
    
    // Bulk operations
    ListTasks(ctx context.Context) ([]TaskInfo, error)
    CancelAll(ctx context.Context) error
}

// TaskStatus represents async task status
type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"
    TaskStatusRunning   TaskStatus = "running"
    TaskStatusCompleted TaskStatus = "completed"
    TaskStatusFailed    TaskStatus = "failed"
    TaskStatusCancelled TaskStatus = "cancelled"
)
```

### StreamingTool Interface

For tools that produce streaming output:

```go
// StreamingTool supports streaming output
type StreamingTool interface {
    Tool
    
    // Streaming execution
    ExecuteStream(ctx context.Context, input interface{}) (<-chan StreamChunk, error)
    
    // Stream control
    SupportsBackpressure() bool
    SetBufferSize(size int) error
}

// StreamChunk represents a chunk of streaming data
type StreamChunk struct {
    Data      interface{} `json:"data"`
    Metadata  StreamMeta  `json:"metadata"`
    Error     error       `json:"error,omitempty"`
    Final     bool        `json:"final"`
}

// StreamMeta contains streaming metadata
type StreamMeta struct {
    Sequence  int64     `json:"sequence"`
    Timestamp time.Time `json:"timestamp"`
    Type      string    `json:"type"`
}
```

## Tool Registry

### ToolRegistry Interface

Manages tool discovery and registration:

```go
// ToolRegistry manages tool registration and discovery
type ToolRegistry interface {
    // Registration
    Register(tool Tool) error
    RegisterWithAlias(tool Tool, aliases ...string) error
    Unregister(name string) error
    
    // Discovery
    Get(name string) (Tool, error)
    List() []ToolInfo
    Search(query ToolQuery) []ToolInfo
    
    // Bulk operations
    RegisterBatch(tools []Tool) error
    UnregisterBatch(names []string) error
    
    // Categories
    ListCategories() []string
    GetByCategory(category string) []ToolInfo
    
    // Lifecycle
    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error
}

// ToolQuery defines search criteria
type ToolQuery struct {
    Name        string   `json:"name,omitempty"`
    Category    string   `json:"category,omitempty"`
    Tags        []string `json:"tags,omitempty"`
    Capabilities []string `json:"capabilities,omitempty"`
    Version     string   `json:"version,omitempty"`
}

// ToolInfo provides tool information
type ToolInfo struct {
    Name         string    `json:"name"`
    Description  string    `json:"description"`
    Version      string    `json:"version"`
    Category     string    `json:"category"`
    Tags         []string  `json:"tags"`
    Aliases      []string  `json:"aliases,omitempty"`
    Registered   time.Time `json:"registered"`
}
```

### Global Registry

Access the global tool registry:

```go
// GetGlobalRegistry returns the global tool registry
func GetGlobalRegistry() ToolRegistry

// RegisterTool registers a tool globally
func RegisterTool(tool Tool) error

// GetTool retrieves a tool from the global registry
func GetTool(name string) (Tool, error)

// ListTools lists all registered tools
func ListTools() []ToolInfo
```

## Tool Metadata

### ToolMetadata

Comprehensive tool metadata:

```go
// ToolMetadata provides detailed tool information
type ToolMetadata struct {
    // Basic info
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Version     string    `json:"version"`
    Description string    `json:"description"`
    
    // Classification
    Category    string    `json:"category"`
    Tags        []string  `json:"tags"`
    Keywords    []string  `json:"keywords"`
    
    // Documentation
    Documentation ToolDocs `json:"documentation"`
    Examples      []Example `json:"examples"`
    
    // Technical details
    Author      string    `json:"author,omitempty"`
    License     string    `json:"license,omitempty"`
    Homepage    string    `json:"homepage,omitempty"`
    Repository  string    `json:"repository,omitempty"`
    
    // Dependencies
    Dependencies []Dependency `json:"dependencies,omitempty"`
    Requirements []string     `json:"requirements,omitempty"`
    
    // Timestamps
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// ToolDocs contains documentation
type ToolDocs struct {
    Summary      string            `json:"summary"`
    Description  string            `json:"description"`
    Usage        string            `json:"usage"`
    Parameters   []ParameterDoc    `json:"parameters"`
    Returns      string            `json:"returns"`
    Errors       []ErrorDoc        `json:"errors"`
    SeeAlso      []string          `json:"see_also,omitempty"`
}

// Example shows tool usage
type Example struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Input       interface{} `json:"input"`
    Output      interface{} `json:"output"`
    Code        string      `json:"code,omitempty"`
}
```

### ToolCapabilities

Defines tool capabilities:

```go
// ToolCapabilities describes what a tool can do
type ToolCapabilities struct {
    // Core capabilities
    Async           bool     `json:"async"`
    Streaming       bool     `json:"streaming"`
    Stateful        bool     `json:"stateful"`
    Idempotent      bool     `json:"idempotent"`
    
    // Resource requirements
    RequiresNetwork bool     `json:"requires_network"`
    RequiresFS      bool     `json:"requires_fs"`
    RequiresAuth    bool     `json:"requires_auth"`
    
    // Performance characteristics
    Timeout         time.Duration `json:"timeout"`
    MaxConcurrency  int          `json:"max_concurrency"`
    RateLimit       *RateLimit   `json:"rate_limit,omitempty"`
    
    // Security
    Permissions     []string     `json:"permissions"`
    SecurityLevel   string       `json:"security_level"`
    
    // Features
    Features        []string     `json:"features"`
    Limitations     []string     `json:"limitations,omitempty"`
}

// RateLimit defines rate limiting
type RateLimit struct {
    RequestsPerSecond int           `json:"requests_per_second"`
    BurstSize         int           `json:"burst_size"`
    Window            time.Duration `json:"window"`
}
```

## Built-in Tools

### File System Tools

```go
// FileReader reads file contents
type FileReader struct {
    *BaseTool
    maxSize      int64
    allowedPaths []string
}

// Input schema
type FileReaderInput struct {
    Path     string `json:"path" jsonschema:"required,description=File path to read"`
    Encoding string `json:"encoding,omitempty" jsonschema:"enum=utf8,enum=base64,default=utf8"`
    Lines    *Lines `json:"lines,omitempty" jsonschema:"description=Line range to read"`
}

// FileWriter writes content to files
type FileWriter struct {
    *BaseTool
    allowedPaths []string
    createDirs   bool
}

// DirectoryLister lists directory contents
type DirectoryLister struct {
    *BaseTool
    maxDepth int
    filters  []string
}
```

### HTTP Tools

```go
// HTTPRequest makes HTTP requests
type HTTPRequest struct {
    *BaseTool
    client      *http.Client
    timeout     time.Duration
    maxRedirects int
}

// Input schema
type HTTPRequestInput struct {
    URL     string            `json:"url" jsonschema:"required,format=uri"`
    Method  string            `json:"method,omitempty" jsonschema:"enum=GET,enum=POST,enum=PUT,enum=DELETE,default=GET"`
    Headers map[string]string `json:"headers,omitempty"`
    Body    interface{}       `json:"body,omitempty"`
    Timeout int               `json:"timeout,omitempty" jsonschema:"minimum=1,maximum=300,default=30"`
}

// WebScraper extracts data from web pages
type WebScraper struct {
    *BaseTool
    parser      HTMLParser
    javascriptEnabled bool
}

// APIClient provides API interaction
type APIClient struct {
    *BaseTool
    baseURL     string
    auth        AuthProvider
    rateLimit   *RateLimiter
}
```

### Data Processing Tools

```go
// JSONProcessor manipulates JSON data
type JSONProcessor struct {
    *BaseTool
    validator SchemaValidator
}

// Operations
const (
    OpValidate   = "validate"
    OpTransform  = "transform"
    OpQuery      = "query"
    OpMerge      = "merge"
    OpPatch      = "patch"
)

// CSVProcessor handles CSV data
type CSVProcessor struct {
    *BaseTool
    delimiter    rune
    hasHeader    bool
    maxRows      int
}

// DataTransformer transforms data formats
type DataTransformer struct {
    *BaseTool
    converters map[string]Converter
}
```

### System Tools

```go
// CommandExecutor runs system commands
type CommandExecutor struct {
    *BaseTool
    allowedCommands []string
    workingDir      string
    environment     []string
    timeout         time.Duration
}

// ProcessManager manages system processes
type ProcessManager struct {
    *BaseTool
    maxProcesses int
    tracking     map[int]*Process
}

// EnvironmentReader reads environment variables
type EnvironmentReader struct {
    *BaseTool
    allowedVars []string
    redactKeys  []string
}
```

### Math and Calculation Tools

```go
// Calculator performs calculations
type Calculator struct {
    *BaseTool
    precision int
    functions map[string]MathFunction
}

// StatisticalAnalyzer performs statistical analysis
type StatisticalAnalyzer struct {
    *BaseTool
    methods []string
}

// Operations
type StatsOperation string

const (
    StatsMean       StatsOperation = "mean"
    StatsMedian     StatsOperation = "median"
    StatsMode       StatsOperation = "mode"
    StatsStdDev     StatsOperation = "stddev"
    StatsVariance   StatsOperation = "variance"
    StatsCorrelation StatsOperation = "correlation"
)
```

### Date/Time Tools

```go
// DateTimeFormatter formats dates and times
type DateTimeFormatter struct {
    *BaseTool
    defaultFormat string
    timezone      *time.Location
}

// DateCalculator performs date calculations
type DateCalculator struct {
    *BaseTool
    calendar Calendar
}

// Operations
type DateOperation string

const (
    DateAdd      DateOperation = "add"
    DateSubtract DateOperation = "subtract"
    DateDiff     DateOperation = "difference"
    DateFormat   DateOperation = "format"
    DateParse    DateOperation = "parse"
)

// TimezoneConverter converts between timezones
type TimezoneConverter struct {
    *BaseTool
    supportedZones []string
}
```

## Tool Implementation

### BaseTool

Base implementation for tools:

```go
// BaseTool provides common tool functionality
type BaseTool struct {
    name         string
    version      string
    description  string
    metadata     ToolMetadata
    capabilities ToolCapabilities
    inputSchema  *jsonschema.Schema
    outputSchema *jsonschema.Schema
    initialized  bool
    mu           sync.RWMutex
}

// NewBaseTool creates a new base tool
func NewBaseTool(name, version, description string) *BaseTool {
    return &BaseTool{
        name:        name,
        version:     version,
        description: description,
        metadata: ToolMetadata{
            Name:        name,
            Version:     version,
            Description: description,
            CreatedAt:   time.Now(),
            UpdatedAt:   time.Now(),
        },
    }
}

// Common methods implementation
func (t *BaseTool) Name() string { return t.name }
func (t *BaseTool) Version() string { return t.version }
func (t *BaseTool) Description() string { return t.description }

func (t *BaseTool) ValidateInput(input interface{}) error {
    if t.inputSchema == nil {
        return nil
    }
    return t.inputSchema.Validate(input)
}

func (t *BaseTool) Initialize(ctx context.Context) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    if t.initialized {
        return nil
    }
    
    // Perform initialization
    t.initialized = true
    return nil
}
```

### Creating Custom Tools

Example custom tool implementation:

```go
// WeatherTool fetches weather information
type WeatherTool struct {
    *BaseTool
    apiKey   string
    apiURL   string
    cache    *Cache
    client   *http.Client
}

// NewWeatherTool creates a new weather tool
func NewWeatherTool(apiKey string) *WeatherTool {
    tool := &WeatherTool{
        BaseTool: NewBaseTool("weather", "1.0.0", "Fetches weather information"),
        apiKey:   apiKey,
        apiURL:   "https://api.weather.com/v1",
        client:   &http.Client{Timeout: 10 * time.Second},
    }
    
    // Define input schema
    tool.inputSchema = &jsonschema.Schema{
        Type: "object",
        Properties: map[string]*jsonschema.Schema{
            "location": {
                Type:        "string",
                Description: "City name or coordinates",
            },
            "units": {
                Type:        "string",
                Enum:        []interface{}{"metric", "imperial"},
                Default:     "metric",
            },
        },
        Required: []string{"location"},
    }
    
    // Define output schema
    tool.outputSchema = &jsonschema.Schema{
        Type: "object",
        Properties: map[string]*jsonschema.Schema{
            "temperature": {Type: "number"},
            "humidity":    {Type: "number"},
            "conditions":  {Type: "string"},
            "wind_speed":  {Type: "number"},
        },
    }
    
    return tool
}

// Execute fetches weather data
func (t *WeatherTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Validate input
    if err := t.ValidateInput(input); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }
    
    // Parse input
    params := input.(map[string]interface{})
    location := params["location"].(string)
    units := "metric"
    if u, ok := params["units"].(string); ok {
        units = u
    }
    
    // Check cache
    cacheKey := fmt.Sprintf("%s:%s", location, units)
    if cached, found := t.cache.Get(cacheKey); found {
        return cached, nil
    }
    
    // Make API request
    url := fmt.Sprintf("%s/weather?location=%s&units=%s&apikey=%s",
        t.apiURL, url.QueryEscape(location), units, t.apiKey)
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := t.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("API request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Parse response
    var data map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    // Extract weather data
    result := map[string]interface{}{
        "temperature": data["main"].(map[string]interface{})["temp"],
        "humidity":    data["main"].(map[string]interface{})["humidity"],
        "conditions":  data["weather"].([]interface{})[0].(map[string]interface{})["description"],
        "wind_speed":  data["wind"].(map[string]interface{})["speed"],
    }
    
    // Cache result
    t.cache.Set(cacheKey, result, 10*time.Minute)
    
    return result, nil
}
```

## Tool Execution

### ToolExecutor

Manages tool execution:

```go
// ToolExecutor handles tool execution
type ToolExecutor interface {
    // Execution
    Execute(ctx context.Context, toolName string, input interface{}) (interface{}, error)
    ExecuteWithOptions(ctx context.Context, toolName string, input interface{}, options ExecutionOptions) (interface{}, error)
    
    // Batch execution
    ExecuteBatch(ctx context.Context, requests []ExecutionRequest) ([]ExecutionResult, error)
    
    // Pipeline execution
    ExecutePipeline(ctx context.Context, pipeline Pipeline) (interface{}, error)
    
    // Monitoring
    GetMetrics() ExecutorMetrics
    GetActiveExecutions() []ExecutionInfo
}

// ExecutionOptions configures execution
type ExecutionOptions struct {
    Timeout      time.Duration          `json:"timeout,omitempty"`
    RetryPolicy  *RetryPolicy           `json:"retry_policy,omitempty"`
    CachePolicy  *CachePolicy           `json:"cache_policy,omitempty"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
    TraceID      string                 `json:"trace_id,omitempty"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
    MaxAttempts   int           `json:"max_attempts"`
    InitialDelay  time.Duration `json:"initial_delay"`
    MaxDelay      time.Duration `json:"max_delay"`
    BackoffFactor float64       `json:"backoff_factor"`
    RetryableErrors []string    `json:"retryable_errors,omitempty"`
}
```

### Execution Context

Tool execution context:

```go
// ToolContext provides execution context
type ToolContext struct {
    // Core context
    context.Context
    
    // Execution info
    ExecutionID string
    ToolName    string
    StartTime   time.Time
    
    // Resources
    Logger      Logger
    Metrics     MetricsCollector
    Tracer      Tracer
    
    // Configuration
    Config      map[string]interface{}
    Secrets     SecretProvider
}

// NewToolContext creates a new tool context
func NewToolContext(ctx context.Context, toolName string) *ToolContext {
    return &ToolContext{
        Context:     ctx,
        ExecutionID: generateID(),
        ToolName:    toolName,
        StartTime:   time.Now(),
        Logger:      getLogger(toolName),
        Metrics:     getMetricsCollector(),
    }
}
```

## Tool Lifecycle

### Initialization

```go
// Initialize prepares a tool for use
func (t *BaseTool) Initialize(ctx context.Context) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    if t.initialized {
        return nil
    }
    
    // Validate configuration
    if err := t.validateConfig(); err != nil {
        return fmt.Errorf("configuration validation failed: %w", err)
    }
    
    // Initialize resources
    if err := t.initializeResources(ctx); err != nil {
        return fmt.Errorf("resource initialization failed: %w", err)
    }
    
    // Load dependencies
    if err := t.loadDependencies(ctx); err != nil {
        return fmt.Errorf("dependency loading failed: %w", err)
    }
    
    t.initialized = true
    t.metadata.UpdatedAt = time.Now()
    
    return nil
}
```

### Cleanup

```go
// Cleanup releases tool resources
func (t *BaseTool) Cleanup(ctx context.Context) error {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    if !t.initialized {
        return nil
    }
    
    // Cleanup resources
    var errs []error
    
    if err := t.cleanupResources(ctx); err != nil {
        errs = append(errs, fmt.Errorf("resource cleanup failed: %w", err))
    }
    
    if err := t.closeConnections(ctx); err != nil {
        errs = append(errs, fmt.Errorf("connection cleanup failed: %w", err))
    }
    
    t.initialized = false
    
    if len(errs) > 0 {
        return fmt.Errorf("cleanup errors: %v", errs)
    }
    
    return nil
}
```

## Error Handling

### Tool Errors

```go
// ToolError represents a tool-specific error
type ToolError struct {
    Tool      string                 `json:"tool"`
    Operation string                 `json:"operation"`
    Code      string                 `json:"code"`
    Message   string                 `json:"message"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Cause     error                  `json:"-"`
    Timestamp time.Time              `json:"timestamp"`
}

// Error codes
const (
    ErrCodeInvalidInput     = "invalid_input"
    ErrCodeExecutionFailed  = "execution_failed"
    ErrCodeTimeout          = "timeout"
    ErrCodeResourceNotFound = "resource_not_found"
    ErrCodePermissionDenied = "permission_denied"
    ErrCodeRateLimit        = "rate_limit_exceeded"
    ErrCodeDependencyFailed = "dependency_failed"
)

// IsRetryable checks if error is retryable
func (e *ToolError) IsRetryable() bool {
    switch e.Code {
    case ErrCodeTimeout, ErrCodeRateLimit:
        return true
    default:
        return false
    }
}
```

## Security

### Tool Permissions

```go
// Permission represents a tool permission
type Permission struct {
    Resource  string   `json:"resource"`
    Actions   []string `json:"actions"`
    Condition string   `json:"condition,omitempty"`
}

// SecurityContext provides security context
type SecurityContext struct {
    User        string       `json:"user"`
    Roles       []string     `json:"roles"`
    Permissions []Permission `json:"permissions"`
    Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// CheckPermission verifies permission
func CheckPermission(ctx SecurityContext, resource string, action string) bool {
    for _, perm := range ctx.Permissions {
        if perm.Resource == resource {
            for _, a := range perm.Actions {
                if a == action || a == "*" {
                    return true
                }
            }
        }
    }
    return false
}
```

## Best Practices

### 1. Tool Design

Design tools with single responsibility:

```go
// Good: Focused tool
type EmailSender struct {
    *BaseTool
    smtpConfig SMTPConfig
}

func (t *EmailSender) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Only handles email sending
    email := input.(*EmailMessage)
    return t.sendEmail(ctx, email)
}

// Avoid: Tool doing too much
type CommunicationTool struct {
    // Handles email, SMS, push notifications, etc.
}
```

### 2. Input Validation

Always validate input thoroughly:

```go
func (t *MyTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Schema validation
    if err := t.ValidateInput(input); err != nil {
        return nil, &ToolError{
            Tool:      t.Name(),
            Code:      ErrCodeInvalidInput,
            Message:   "input validation failed",
            Cause:     err,
        }
    }
    
    // Business logic validation
    data := input.(map[string]interface{})
    if err := t.validateBusinessRules(data); err != nil {
        return nil, err
    }
    
    // Execute
    return t.executeCore(ctx, data)
}
```

### 3. Resource Management

Properly manage resources:

```go
type ResourcefulTool struct {
    *BaseTool
    pool     *ResourcePool
    connections map[string]Connection
    mu       sync.RWMutex
}

func (t *ResourcefulTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Acquire resource
    resource, err := t.pool.Acquire(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to acquire resource: %w", err)
    }
    defer t.pool.Release(resource)
    
    // Use resource
    return t.processWithResource(ctx, resource, input)
}
```

### 4. Error Context

Provide rich error context:

```go
func (t *MyTool) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    result, err := t.performOperation(ctx, input)
    if err != nil {
        return nil, &ToolError{
            Tool:      t.Name(),
            Operation: "performOperation",
            Code:      ErrCodeExecutionFailed,
            Message:   fmt.Sprintf("operation failed for input: %v", input),
            Details: map[string]interface{}{
                "input_type": fmt.Sprintf("%T", input),
                "context_deadline": ctx.Deadline(),
            },
            Cause:     err,
            Timestamp: time.Now(),
        }
    }
    return result, nil
}
```

### 5. Testing Tools

Comprehensive tool testing:

```go
func TestMyTool(t *testing.T) {
    tool := NewMyTool()
    
    // Test initialization
    ctx := context.Background()
    err := tool.Initialize(ctx)
    assert.NoError(t, err)
    defer tool.Cleanup(ctx)
    
    // Test valid input
    input := map[string]interface{}{
        "param1": "value1",
        "param2": 42,
    }
    
    result, err := tool.Execute(ctx, input)
    assert.NoError(t, err)
    assert.NotNil(t, result)
    
    // Test invalid input
    invalidInput := map[string]interface{}{
        "param1": 123, // Wrong type
    }
    
    _, err = tool.Execute(ctx, invalidInput)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid_input")
    
    // Test timeout
    timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
    defer cancel()
    
    _, err = tool.Execute(timeoutCtx, input)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "timeout")
}
```

This comprehensive tool API documentation provides all the necessary information for building and integrating tools with Go-LLMs applications.