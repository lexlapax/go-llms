# Core Type Definitions

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [API Reference](../../technical/api-reference) / Types**

Complete reference for core types, structs, interfaces, and constants used throughout Go-LLMs, including schema types, error types, structured output types, configuration types, and utility types.

## Schema Package Types

### JSON Schema Types

```go
package schema

// Schema represents a JSON Schema definition
type Schema struct {
    // Core schema properties
    Type             string                    `json:"type,omitempty"`
    Title            string                    `json:"title,omitempty"`
    Description      string                    `json:"description,omitempty"`
    Default          interface{}               `json:"default,omitempty"`
    Examples         []interface{}             `json:"examples,omitempty"`
    
    // Validation keywords
    Enum             []interface{}             `json:"enum,omitempty"`
    Const            interface{}               `json:"const,omitempty"`
    
    // Numeric validation
    MultipleOf       *float64                  `json:"multipleOf,omitempty"`
    Maximum          *float64                  `json:"maximum,omitempty"`
    ExclusiveMaximum *float64                  `json:"exclusiveMaximum,omitempty"`
    Minimum          *float64                  `json:"minimum,omitempty"`
    ExclusiveMinimum *float64                  `json:"exclusiveMinimum,omitempty"`
    
    // String validation
    MaxLength        *int                      `json:"maxLength,omitempty"`
    MinLength        *int                      `json:"minLength,omitempty"`
    Pattern          string                    `json:"pattern,omitempty"`
    Format           string                    `json:"format,omitempty"`
    
    // Array validation
    Items            *Schema                   `json:"items,omitempty"`
    AdditionalItems  *Schema                   `json:"additionalItems,omitempty"`
    MaxItems         *int                      `json:"maxItems,omitempty"`
    MinItems         *int                      `json:"minItems,omitempty"`
    UniqueItems      bool                      `json:"uniqueItems,omitempty"`
    Contains         *Schema                   `json:"contains,omitempty"`
    
    // Object validation
    Properties           map[string]*Schema    `json:"properties,omitempty"`
    PatternProperties    map[string]*Schema    `json:"patternProperties,omitempty"`
    AdditionalProperties interface{}           `json:"additionalProperties,omitempty"`
    Required             []string              `json:"required,omitempty"`
    PropertyNames        *Schema               `json:"propertyNames,omitempty"`
    MaxProperties        *int                  `json:"maxProperties,omitempty"`
    MinProperties        *int                  `json:"minProperties,omitempty"`
    Dependencies         map[string]interface{} `json:"dependencies,omitempty"`
    
    // Conditional validation
    If               *Schema                   `json:"if,omitempty"`
    Then             *Schema                   `json:"then,omitempty"`
    Else             *Schema                   `json:"else,omitempty"`
    
    // Composition
    AllOf            []*Schema                 `json:"allOf,omitempty"`
    AnyOf            []*Schema                 `json:"anyOf,omitempty"`
    OneOf            []*Schema                 `json:"oneOf,omitempty"`
    Not              *Schema                   `json:"not,omitempty"`
    
    // Annotations
    ReadOnly         bool                      `json:"readOnly,omitempty"`
    WriteOnly        bool                      `json:"writeOnly,omitempty"`
    Deprecated       bool                      `json:"deprecated,omitempty"`
    
    // References
    Ref              string                    `json:"$ref,omitempty"`
    Definitions      map[string]*Schema        `json:"definitions,omitempty"`
}

// SchemaType represents JSON Schema types
type SchemaType string

const (
    TypeString  SchemaType = "string"
    TypeNumber  SchemaType = "number"
    TypeInteger SchemaType = "integer"
    TypeBoolean SchemaType = "boolean"
    TypeArray   SchemaType = "array"
    TypeObject  SchemaType = "object"
    TypeNull    SchemaType = "null"
)

// Format represents string format types
type Format string

const (
    FormatDateTime  Format = "date-time"
    FormatDate      Format = "date"
    FormatTime      Format = "time"
    FormatDuration  Format = "duration"
    FormatEmail     Format = "email"
    FormatHostname  Format = "hostname"
    FormatIPv4      Format = "ipv4"
    FormatIPv6      Format = "ipv6"
    FormatURI       Format = "uri"
    FormatURIRef    Format = "uri-reference"
    FormatUUID      Format = "uuid"
    FormatRegex     Format = "regex"
    FormatJSONPtr   Format = "json-pointer"
    FormatRelJSONPtr Format = "relative-json-pointer"
)
```

### Schema Registry Types

```go
// SchemaRegistry manages schema definitions
type SchemaRegistry interface {
    // Registration
    Register(id string, schema *Schema) error
    RegisterWithVersion(id string, version string, schema *Schema) error
    Unregister(id string) error
    
    // Retrieval
    Get(id string) (*Schema, error)
    GetVersion(id string, version string) (*Schema, error)
    GetLatest(id string) (*Schema, error)
    
    // Listing
    List() []SchemaInfo
    ListVersions(id string) []string
    
    // Validation
    Validate(id string, data interface{}) error
    ValidateWithVersion(id string, version string, data interface{}) error
}

// SchemaInfo provides schema metadata
type SchemaInfo struct {
    ID          string    `json:"id"`
    Version     string    `json:"version"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Tags        []string  `json:"tags,omitempty"`
}

// ValidationResult contains validation details
type ValidationResult struct {
    Valid     bool              `json:"valid"`
    Errors    []ValidationError `json:"errors,omitempty"`
    Warnings  []string          `json:"warnings,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
    Path        string      `json:"path"`
    Property    string      `json:"property"`
    Message     string      `json:"message"`
    SchemaPath  string      `json:"schema_path"`
    InvalidValue interface{} `json:"invalid_value,omitempty"`
}
```

### Type Conversion Types

```go
// TypeConverter converts between types
type TypeConverter interface {
    // Conversion
    Convert(from interface{}, toType reflect.Type) (interface{}, error)
    CanConvert(from reflect.Type, to reflect.Type) bool
    
    // Registration
    RegisterConverter(from, to reflect.Type, converter ConversionFunc) error
    
    // Type inference
    InferType(value interface{}) reflect.Type
    InferSchemaType(value interface{}) SchemaType
}

// ConversionFunc defines a type conversion function
type ConversionFunc func(interface{}) (interface{}, error)

// ConversionRegistry manages type conversions
type ConversionRegistry struct {
    converters map[ConversionKey]ConversionFunc
    mu         sync.RWMutex
}

// ConversionKey identifies a conversion
type ConversionKey struct {
    From reflect.Type
    To   reflect.Type
}
```

## Error Package Types

### Core Error Types

```go
package errors

// Error represents a structured error
type Error struct {
    Type       ErrorType              `json:"type"`
    Code       string                 `json:"code"`
    Message    string                 `json:"message"`
    Details    map[string]interface{} `json:"details,omitempty"`
    Context    ErrorContext           `json:"context,omitempty"`
    Cause      error                  `json:"-"`
    StackTrace []StackFrame           `json:"stack_trace,omitempty"`
    Timestamp  time.Time              `json:"timestamp"`
}

// ErrorType categorizes errors
type ErrorType string

const (
    // System errors
    ErrorTypeInternal      ErrorType = "internal"
    ErrorTypeConfiguration ErrorType = "configuration"
    ErrorTypeInitialization ErrorType = "initialization"
    
    // Input/Output errors
    ErrorTypeValidation    ErrorType = "validation"
    ErrorTypeParsing       ErrorType = "parsing"
    ErrorTypeSerialization ErrorType = "serialization"
    
    // Network errors
    ErrorTypeNetwork       ErrorType = "network"
    ErrorTypeTimeout       ErrorType = "timeout"
    ErrorTypeConnection    ErrorType = "connection"
    
    // Resource errors
    ErrorTypeNotFound      ErrorType = "not_found"
    ErrorTypePermission    ErrorType = "permission"
    ErrorTypeQuota         ErrorType = "quota"
    ErrorTypeRateLimit     ErrorType = "rate_limit"
    
    // Provider errors
    ErrorTypeProvider      ErrorType = "provider"
    ErrorTypeAuthentication ErrorType = "authentication"
    ErrorTypeAuthorization ErrorType = "authorization"
)

// ErrorContext provides error context
type ErrorContext struct {
    Component   string                 `json:"component,omitempty"`
    Operation   string                 `json:"operation,omitempty"`
    Resource    string                 `json:"resource,omitempty"`
    User        string                 `json:"user,omitempty"`
    RequestID   string                 `json:"request_id,omitempty"`
    TraceID     string                 `json:"trace_id,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// StackFrame represents a stack trace frame
type StackFrame struct {
    Function string `json:"function"`
    File     string `json:"file"`
    Line     int    `json:"line"`
}
```

### Error Builder

```go
// ErrorBuilder builds structured errors
type ErrorBuilder struct {
    err *Error
}

// NewError creates a new error builder
func NewError(errType ErrorType, code string) *ErrorBuilder {
    return &ErrorBuilder{
        err: &Error{
            Type:      errType,
            Code:      code,
            Timestamp: time.Now(),
        },
    }
}

// Builder methods
func (b *ErrorBuilder) WithMessage(msg string) *ErrorBuilder
func (b *ErrorBuilder) WithDetails(details map[string]interface{}) *ErrorBuilder
func (b *ErrorBuilder) WithCause(cause error) *ErrorBuilder
func (b *ErrorBuilder) WithContext(ctx ErrorContext) *ErrorBuilder
func (b *ErrorBuilder) WithStackTrace() *ErrorBuilder
func (b *ErrorBuilder) Build() *Error
```

### Error Utilities

```go
// IsRetryable checks if error is retryable
func IsRetryable(err error) bool

// GetErrorType extracts error type
func GetErrorType(err error) ErrorType

// WrapError wraps an error with additional context
func WrapError(err error, message string) error

// UnwrapError unwraps to the root cause
func UnwrapError(err error) error

// ErrorChain returns the error chain
func ErrorChain(err error) []error
```

## Structured Output Types

### Parser Types

```go
package structured

// Parser extracts structured data from text
type Parser interface {
    // Parsing
    Parse(text string, schema OutputSchema) (interface{}, error)
    ParseWithOptions(text string, schema OutputSchema, options ParseOptions) (interface{}, error)
    
    // Validation
    Validate(output interface{}, schema OutputSchema) error
    
    // Schema management
    RegisterSchema(name string, schema OutputSchema) error
    GetSchema(name string) (OutputSchema, error)
}

// OutputSchema defines expected output structure
type OutputSchema struct {
    Name        string             `json:"name"`
    Description string             `json:"description"`
    Type        string             `json:"type"`
    Schema      *jsonschema.Schema `json:"schema"`
    Examples    []Example          `json:"examples,omitempty"`
    Extractors  []Extractor        `json:"extractors,omitempty"`
}

// ParseOptions configures parsing
type ParseOptions struct {
    Strict      bool                   `json:"strict"`
    AllowPartial bool                  `json:"allow_partial"`
    Timeout     time.Duration          `json:"timeout,omitempty"`
    MaxRetries  int                    `json:"max_retries,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Example shows input/output examples
type Example struct {
    Input       string      `json:"input"`
    Output      interface{} `json:"output"`
    Description string      `json:"description,omitempty"`
}
```

### Extractor Types

```go
// Extractor extracts specific data from text
type Extractor interface {
    // Core extraction
    Extract(text string) (interface{}, error)
    ExtractAll(text string) ([]interface{}, error)
    
    // Configuration
    GetName() string
    GetPattern() string
    GetOutputType() reflect.Type
}

// RegexExtractor uses regex patterns
type RegexExtractor struct {
    Name       string
    Pattern    *regexp.Regexp
    Groups     []string
    OutputType reflect.Type
}

// JSONExtractor extracts JSON data
type JSONExtractor struct {
    Name         string
    StartMarkers []string
    EndMarkers   []string
    Schema       *jsonschema.Schema
}

// TableExtractor extracts tabular data
type TableExtractor struct {
    Name        string
    Delimiter   string
    HasHeaders  bool
    ColumnTypes []reflect.Type
}
```

### Format Types

```go
// OutputFormat defines output formatting
type OutputFormat string

const (
    FormatJSON       OutputFormat = "json"
    FormatYAML       OutputFormat = "yaml"
    FormatXML        OutputFormat = "xml"
    FormatCSV        OutputFormat = "csv"
    FormatMarkdown   OutputFormat = "markdown"
    FormatPlainText  OutputFormat = "text"
)

// Formatter formats structured output
type Formatter interface {
    Format(data interface{}, format OutputFormat) (string, error)
    FormatWithOptions(data interface{}, format OutputFormat, options FormatOptions) (string, error)
}

// FormatOptions configures formatting
type FormatOptions struct {
    Indent      string                 `json:"indent,omitempty"`
    Pretty      bool                   `json:"pretty"`
    EscapeHTML  bool                   `json:"escape_html"`
    SortKeys    bool                   `json:"sort_keys"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
```

## Configuration Types

### Provider Configuration

```go
// ProviderConfig configures an LLM provider
type ProviderConfig struct {
    Type        string                 `yaml:"type" json:"type"`
    Name        string                 `yaml:"name" json:"name"`
    APIKey      string                 `yaml:"api_key" json:"api_key"`
    BaseURL     string                 `yaml:"base_url,omitempty" json:"base_url,omitempty"`
    Model       string                 `yaml:"model" json:"model"`
    Timeout     time.Duration          `yaml:"timeout,omitempty" json:"timeout,omitempty"`
    MaxRetries  int                    `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`
    Options     map[string]interface{} `yaml:"options,omitempty" json:"options,omitempty"`
}

// ProviderType identifies provider types
type ProviderType string

const (
    ProviderTypeOpenAI     ProviderType = "openai"
    ProviderTypeAnthropic  ProviderType = "anthropic"
    ProviderTypeGoogle     ProviderType = "google"
    ProviderTypeVertexAI   ProviderType = "vertexai"
    ProviderTypeOllama     ProviderType = "ollama"
    ProviderTypeOpenRouter ProviderType = "openrouter"
)
```

### Agent Configuration

```go
// AgentConfig configures an agent
type AgentConfig struct {
    Name         string                 `yaml:"name" json:"name"`
    Type         AgentType              `yaml:"type" json:"type"`
    Description  string                 `yaml:"description,omitempty" json:"description,omitempty"`
    Provider     string                 `yaml:"provider" json:"provider"`
    Model        string                 `yaml:"model,omitempty" json:"model,omitempty"`
    SystemPrompt string                 `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
    Tools        []string               `yaml:"tools,omitempty" json:"tools,omitempty"`
    Temperature  float64                `yaml:"temperature,omitempty" json:"temperature,omitempty"`
    MaxTokens    int                    `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`
    Parameters   map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// AgentType identifies agent types
type AgentType string

const (
    AgentTypeSimple    AgentType = "simple"
    AgentTypeLLM       AgentType = "llm"
    AgentTypeWorkflow  AgentType = "workflow"
    AgentTypeStateful  AgentType = "stateful"
    AgentTypeReactive  AgentType = "reactive"
)
```

### Tool Configuration

```go
// ToolConfig configures a tool
type ToolConfig struct {
    Name        string                 `yaml:"name" json:"name"`
    Type        string                 `yaml:"type" json:"type"`
    Enabled     bool                   `yaml:"enabled" json:"enabled"`
    Version     string                 `yaml:"version,omitempty" json:"version,omitempty"`
    Timeout     time.Duration          `yaml:"timeout,omitempty" json:"timeout,omitempty"`
    MaxRetries  int                    `yaml:"max_retries,omitempty" json:"max_retries,omitempty"`
    RateLimit   *RateLimitConfig       `yaml:"rate_limit,omitempty" json:"rate_limit,omitempty"`
    Permissions []string               `yaml:"permissions,omitempty" json:"permissions,omitempty"`
    Parameters  map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// RateLimitConfig configures rate limiting
type RateLimitConfig struct {
    RequestsPerSecond int           `yaml:"requests_per_second" json:"requests_per_second"`
    BurstSize         int           `yaml:"burst_size" json:"burst_size"`
    Window            time.Duration `yaml:"window" json:"window"`
}
```

### Application Configuration

```go
// AppConfig configures the application
type AppConfig struct {
    Name        string                    `yaml:"name" json:"name"`
    Version     string                    `yaml:"version" json:"version"`
    Environment string                    `yaml:"environment" json:"environment"`
    LogLevel    string                    `yaml:"log_level" json:"log_level"`
    Providers   map[string]ProviderConfig `yaml:"providers" json:"providers"`
    Agents      map[string]AgentConfig    `yaml:"agents" json:"agents"`
    Tools       map[string]ToolConfig     `yaml:"tools" json:"tools"`
    Security    SecurityConfig            `yaml:"security" json:"security"`
    Monitoring  MonitoringConfig          `yaml:"monitoring" json:"monitoring"`
}

// SecurityConfig configures security
type SecurityConfig struct {
    AuthEnabled     bool                   `yaml:"auth_enabled" json:"auth_enabled"`
    AuthProviders   []string               `yaml:"auth_providers" json:"auth_providers"`
    APIKeyHeader    string                 `yaml:"api_key_header" json:"api_key_header"`
    JWTSecret       string                 `yaml:"jwt_secret" json:"jwt_secret"`
    AllowedOrigins  []string               `yaml:"allowed_origins" json:"allowed_origins"`
    RateLimiting    bool                   `yaml:"rate_limiting" json:"rate_limiting"`
    Encryption      EncryptionConfig       `yaml:"encryption" json:"encryption"`
}

// MonitoringConfig configures monitoring
type MonitoringConfig struct {
    Enabled         bool                   `yaml:"enabled" json:"enabled"`
    MetricsEnabled  bool                   `yaml:"metrics_enabled" json:"metrics_enabled"`
    TracingEnabled  bool                   `yaml:"tracing_enabled" json:"tracing_enabled"`
    LoggingEnabled  bool                   `yaml:"logging_enabled" json:"logging_enabled"`
    Exporters       []string               `yaml:"exporters" json:"exporters"`
    SampleRate      float64                `yaml:"sample_rate" json:"sample_rate"`
    Configuration   map[string]interface{} `yaml:"configuration" json:"configuration"`
}
```

## Message and Communication Types

### Message Types

```go
// Message represents a conversation message
type Message struct {
    Role        MessageRole    `json:"role"`
    Content     string         `json:"content"`
    Name        string         `json:"name,omitempty"`
    ToolCalls   []ToolCall     `json:"tool_calls,omitempty"`
    ToolCallID  string         `json:"tool_call_id,omitempty"`
    Metadata    MessageMeta    `json:"metadata,omitempty"`
}

// MessageRole defines message roles
type MessageRole string

const (
    RoleSystem    MessageRole = "system"
    RoleUser      MessageRole = "user"
    RoleAssistant MessageRole = "assistant"
    RoleTool      MessageRole = "tool"
)

// MessageMeta contains message metadata
type MessageMeta struct {
    Timestamp   time.Time              `json:"timestamp"`
    TokenCount  int                    `json:"token_count,omitempty"`
    Model       string                 `json:"model,omitempty"`
    Provider    string                 `json:"provider,omitempty"`
    Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// ToolCall represents a function/tool call
type ToolCall struct {
    ID       string      `json:"id"`
    Type     string      `json:"type"`
    Function FunctionCall `json:"function"`
}

// FunctionCall contains function call details
type FunctionCall struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}
```

### Conversation Types

```go
// Conversation represents a conversation thread
type Conversation struct {
    ID          string                 `json:"id"`
    Title       string                 `json:"title,omitempty"`
    Messages    []Message              `json:"messages"`
    Context     ConversationContext    `json:"context,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// ConversationContext provides conversation context
type ConversationContext struct {
    User        string                 `json:"user,omitempty"`
    Session     string                 `json:"session,omitempty"`
    Application string                 `json:"application,omitempty"`
    Environment string                 `json:"environment,omitempty"`
    Variables   map[string]interface{} `json:"variables,omitempty"`
}

// ConversationState tracks conversation state
type ConversationState struct {
    CurrentTurn    int                    `json:"current_turn"`
    TotalTokens    int                    `json:"total_tokens"`
    LastActivity   time.Time              `json:"last_activity"`
    ActiveTools    []string               `json:"active_tools,omitempty"`
    Memory         map[string]interface{} `json:"memory,omitempty"`
}
```

## Utility Types

### Event Types

```go
// Event represents a system event
type Event struct {
    ID          string                 `json:"id"`
    Type        EventType              `json:"type"`
    Source      string                 `json:"source"`
    Timestamp   time.Time              `json:"timestamp"`
    Data        interface{}            `json:"data"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// EventType categorizes events
type EventType string

const (
    // Lifecycle events
    EventTypeStarted     EventType = "started"
    EventTypeStopped     EventType = "stopped"
    EventTypeInitialized EventType = "initialized"
    EventTypeShutdown    EventType = "shutdown"
    
    // Execution events
    EventTypeExecuting   EventType = "executing"
    EventTypeCompleted   EventType = "completed"
    EventTypeFailed      EventType = "failed"
    EventTypeRetrying    EventType = "retrying"
    
    // State events
    EventTypeStateChange EventType = "state_change"
    EventTypeDataUpdate  EventType = "data_update"
    
    // System events
    EventTypeError       EventType = "error"
    EventTypeWarning     EventType = "warning"
    EventTypeInfo        EventType = "info"
)

// EventHandler processes events
type EventHandler func(event Event) error

// EventFilter filters events
type EventFilter func(event Event) bool
```

### Metrics Types

```go
// Metric represents a measurement
type Metric struct {
    Name        string                 `json:"name"`
    Type        MetricType             `json:"type"`
    Value       float64                `json:"value"`
    Unit        string                 `json:"unit,omitempty"`
    Tags        map[string]string      `json:"tags,omitempty"`
    Timestamp   time.Time              `json:"timestamp"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MetricType categorizes metrics
type MetricType string

const (
    MetricTypeCounter   MetricType = "counter"
    MetricTypeGauge     MetricType = "gauge"
    MetricTypeHistogram MetricType = "histogram"
    MetricTypeSummary   MetricType = "summary"
)

// MetricsCollector collects metrics
type MetricsCollector interface {
    // Recording
    RecordCounter(name string, value float64, tags map[string]string)
    RecordGauge(name string, value float64, tags map[string]string)
    RecordHistogram(name string, value float64, tags map[string]string)
    
    // Retrieval
    GetMetric(name string) (*Metric, error)
    GetMetrics(filter MetricFilter) []Metric
    
    // Management
    Reset()
    Flush() error
}

// MetricFilter filters metrics
type MetricFilter struct {
    Names    []string          `json:"names,omitempty"`
    Types    []MetricType      `json:"types,omitempty"`
    Tags     map[string]string `json:"tags,omitempty"`
    Since    time.Time         `json:"since,omitempty"`
    Until    time.Time         `json:"until,omitempty"`
}
```

### Resource Types

```go
// Resource represents a system resource
type Resource struct {
    ID          string                 `json:"id"`
    Type        ResourceType           `json:"type"`
    Name        string                 `json:"name"`
    Status      ResourceStatus         `json:"status"`
    Owner       string                 `json:"owner,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// ResourceType categorizes resources
type ResourceType string

const (
    ResourceTypeProvider  ResourceType = "provider"
    ResourceTypeAgent     ResourceType = "agent"
    ResourceTypeTool      ResourceType = "tool"
    ResourceTypeWorkflow  ResourceType = "workflow"
    ResourceTypeModel     ResourceType = "model"
    ResourceTypeData      ResourceType = "data"
)

// ResourceStatus represents resource state
type ResourceStatus string

const (
    ResourceStatusActive    ResourceStatus = "active"
    ResourceStatusInactive  ResourceStatus = "inactive"
    ResourceStatusLoading   ResourceStatus = "loading"
    ResourceStatusError     ResourceStatus = "error"
    ResourceStatusUnknown   ResourceStatus = "unknown"
)

// ResourceLimits defines resource constraints
type ResourceLimits struct {
    MaxMemory      int64         `json:"max_memory,omitempty"`
    MaxCPU         float64       `json:"max_cpu,omitempty"`
    MaxDuration    time.Duration `json:"max_duration,omitempty"`
    MaxConcurrency int           `json:"max_concurrency,omitempty"`
}
```

### Task Types

```go
// Task represents an executable task
type Task struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"`
    Name        string                 `json:"name"`
    Description string                 `json:"description,omitempty"`
    Input       interface{}            `json:"input,omitempty"`
    Output      interface{}            `json:"output,omitempty"`
    Status      TaskStatus             `json:"status"`
    Priority    TaskPriority           `json:"priority"`
    Dependencies []string              `json:"dependencies,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    StartedAt   *time.Time             `json:"started_at,omitempty"`
    CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// TaskStatus represents task state
type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"
    TaskStatusQueued    TaskStatus = "queued"
    TaskStatusRunning   TaskStatus = "running"
    TaskStatusCompleted TaskStatus = "completed"
    TaskStatusFailed    TaskStatus = "failed"
    TaskStatusCancelled TaskStatus = "cancelled"
    TaskStatusSkipped   TaskStatus = "skipped"
)

// TaskPriority defines task priority
type TaskPriority int

const (
    PriorityLow    TaskPriority = 0
    PriorityNormal TaskPriority = 1
    PriorityHigh   TaskPriority = 2
    PriorityCritical TaskPriority = 3
)

// TaskResult contains task execution result
type TaskResult struct {
    TaskID      string                 `json:"task_id"`
    Success     bool                   `json:"success"`
    Output      interface{}            `json:"output,omitempty"`
    Error       error                  `json:"error,omitempty"`
    Metrics     TaskMetrics            `json:"metrics"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TaskMetrics contains task performance metrics
type TaskMetrics struct {
    Duration        time.Duration `json:"duration"`
    CPUTime         time.Duration `json:"cpu_time,omitempty"`
    MemoryUsed      int64         `json:"memory_used,omitempty"`
    NetworkIO       int64         `json:"network_io,omitempty"`
    RetryCount      int           `json:"retry_count,omitempty"`
}
```

## Constants

### Common Constants

```go
// Version constants
const (
    VersionMajor = 0
    VersionMinor = 3
    VersionPatch = 5
    Version      = "0.3.5"
)

// Limit constants
const (
    DefaultTimeout         = 30 * time.Second
    DefaultMaxRetries      = 3
    DefaultMaxConcurrency  = 10
    DefaultBufferSize      = 1024
    DefaultCacheSize       = 100
    DefaultBatchSize       = 50
)

// Size constants
const (
    KB = 1024
    MB = 1024 * KB
    GB = 1024 * MB
    
    MaxRequestSize  = 10 * MB
    MaxResponseSize = 50 * MB
    MaxFileSize     = 100 * MB
)

// HTTP constants
const (
    HeaderContentType     = "Content-Type"
    HeaderAuthorization   = "Authorization"
    HeaderUserAgent       = "User-Agent"
    HeaderAccept          = "Accept"
    HeaderAcceptEncoding  = "Accept-Encoding"
    
    ContentTypeJSON       = "application/json"
    ContentTypeXML        = "application/xml"
    ContentTypeText       = "text/plain"
    ContentTypeHTML       = "text/html"
    ContentTypeForm       = "application/x-www-form-urlencoded"
    ContentTypeMultipart  = "multipart/form-data"
)

// Environment variables
const (
    EnvLogLevel        = "GO_LLMS_LOG_LEVEL"
    EnvConfigPath      = "GO_LLMS_CONFIG_PATH"
    EnvDataDir         = "GO_LLMS_DATA_DIR"
    EnvCacheDir        = "GO_LLMS_CACHE_DIR"
    EnvDebugMode       = "GO_LLMS_DEBUG"
    EnvMaxWorkers      = "GO_LLMS_MAX_WORKERS"
    EnvRequestTimeout  = "GO_LLMS_REQUEST_TIMEOUT"
)
```

This comprehensive type definitions documentation provides a complete reference for all core types, structs, interfaces, and constants used throughout the Go-LLMs library.