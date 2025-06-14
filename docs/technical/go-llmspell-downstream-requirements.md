# Upstream Feature Requests for go-llms

## Executive Summary

go-llmspell is a Go library that provides scriptable LLM interactions by bridging embedded scripting languages (Lua, JavaScript, Tengo) to go-llms functionality. Our architecture follows a strict "bridge-only" principle: we only wrap existing go-llms functionality without implementing business logic.

This document outlines feature requests that would significantly improve go-llms' extensibility and make it easier for projects like ours (and others) to build on top of go-llms without reimplementing core functionality.

## Context: go-llmspell Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Script Engine  │────▶│  Bridge Layer   │────▶│    go-llms      │
│ (Lua/JS/Tengo) │     │  (go-llmspell)  │     │  (Core Logic)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

Our bridge layer's sole responsibility is type conversion and method delegation. We deliberately avoid implementing any business logic that belongs in go-llms.

## Feature Requests

### 1. Schema Package Implementations

**Current State**: The `schema/domain` package defines excellent interfaces (`SchemaRepository`, `SchemaGenerator`) but provides no implementations.

**How We'd Use It**:
```go
// In our schema bridge
func (b *SchemaBridge) Initialize(ctx context.Context) error {
    // Instead of implementing our own repository
    b.repository = schema.NewInMemoryRepository()
    // Or for persistence
    b.repository = schema.NewFileRepository("/schemas")
    
    // Instead of complex reflection code
    b.generator = schema.NewReflectionGenerator()
    
    return nil
}
```

**Why It Makes Sense**:
- Schema storage and generation are common needs for any LLM application
- Multiple go-llms users are likely implementing similar functionality
- Standardized implementations ensure consistent behavior across the ecosystem

**Implementation Expectations**:
```go
// pkg/schema/repository/memory.go
package repository

type InMemoryRepository struct {
    schemas map[string]*domain.Schema
    mu      sync.RWMutex
}

func NewInMemoryRepository() domain.SchemaRepository {
    return &InMemoryRepository{
        schemas: make(map[string]*domain.Schema),
    }
}

// pkg/schema/repository/file.go
type FileRepository struct {
    basePath string
    format   string // "json" or "yaml"
}

func NewFileRepository(path string, opts ...Option) domain.SchemaRepository {
    // Implementation
}

// pkg/schema/generator/reflection.go
package generator

type ReflectionGenerator struct {
    tagName         string // default: "json"
    includePrivate  bool
    maxDepth        int
}

func NewReflectionGenerator(opts ...Option) domain.SchemaGenerator {
    // Generate schemas from Go structs using reflection
}
```

### 2. Tool Discovery Enhancements

**Current State**: Tool discovery is read-only; tools must be compiled in.

**How We'd Use It**:
```go
// Allow scripts to register tools dynamically
func (b *ToolsBridge) registerCustomTool(name string, schema map[string]interface{}, 
    handler func(args map[string]interface{}) (interface{}, error)) error {
    
    toolInfo := tools.ToolInfo{
        Name:        name,
        Description: schema["description"].(string),
        Schema:      schemaToToolSchema(schema),
    }
    
    factory := func() (domain.Tool, error) {
        return &scriptTool{handler: handler}, nil
    }
    
    // Dynamic registration
    return b.discovery.RegisterTool(toolInfo, factory)
}
```

**Why It Makes Sense**:
- Enables plugin architectures and runtime extensibility
- Allows tools to be loaded from external sources (files, databases, APIs)
- Supports multi-tenant scenarios where different users have different tools

**Implementation Expectations**:
```go
// Extend the existing ToolDiscovery interface
type ToolDiscovery interface {
    // Existing methods...
    
    // New methods for dynamic registration
    RegisterTool(info ToolInfo, factory ToolFactory) error
    UnregisterTool(name string) error
    GetRegisteredTools() []ToolInfo
    
    // For persistence
    SaveRegistry(writer io.Writer) error
    LoadRegistry(reader io.Reader) error
}

// Default implementation should handle both built-in and dynamic tools
type defaultDiscovery struct {
    builtinTools   map[string]ToolInfo
    dynamicTools   map[string]ToolInfo
    factories      map[string]ToolFactory
    mu             sync.RWMutex
}
```

### 3. Bridge-Friendly Type System

**Current State**: Each integration must implement its own type conversions.

**How We'd Use It**:
```go
// Instead of manual conversions everywhere
func (b *SchemaBridge) ExecuteMethod(ctx context.Context, name string, args []interface{}) (interface{}, error) {
    // Use centralized type conversion
    schema, err := typeRegistry.Convert(args[0], reflect.TypeOf(&domain.Schema{}))
    if err != nil {
        return nil, fmt.Errorf("invalid schema: %w", err)
    }
    
    result := b.validator.Validate(schema.(*domain.Schema), data)
    
    // Convert back to script-friendly format
    return typeRegistry.Convert(result, reflect.TypeOf(map[string]interface{}{}))
}
```

**Why It Makes Sense**:
- Type conversion is a common challenge for all go-llms integrations
- Standardized conversions prevent subtle bugs and inconsistencies
- Enables better performance through conversion caching

**Implementation Expectations**:
```go
// pkg/util/types/registry.go
package types

type Registry struct {
    converters map[reflect.Type]map[reflect.Type]TypeConverter
    mu         sync.RWMutex
}

// Global registry with common conversions pre-registered
var DefaultRegistry = NewRegistry().
    RegisterConverter(reflect.TypeOf(&domain.Schema{}), reflect.TypeOf(map[string]interface{}{}), SchemaToMapConverter).
    RegisterConverter(reflect.TypeOf(map[string]interface{}{}), reflect.TypeOf(&domain.Schema{}), MapToSchemaConverter).
    // ... other common conversions

// Bi-directional conversion support
func (r *Registry) Convert(from interface{}, toType reflect.Type) (interface{}, error) {
    fromType := reflect.TypeOf(from)
    
    // Direct conversion
    if converter, ok := r.getConverter(fromType, toType); ok {
        return converter.Convert(from)
    }
    
    // Try reverse conversion
    if converter, ok := r.getConverter(toType, fromType); ok && converter.CanReverse() {
        return converter.Reverse(from)
    }
    
    // Multi-hop conversion through common types
    return r.convertViaIntermediate(from, toType)
}
```

### 4. Event System Enhancements

**Current State**: Events are Go structs without built-in serialization or filtering.

**How We'd Use It**:
```go
// In our event bridge
func (b *EventBridge) subscribeToEvents(pattern string, handler func(event map[string]interface{})) error {
    filter := events.NewPatternFilter(pattern) // e.g., "tool.*" for all tool events
    
    return b.agent.Subscribe(func(event domain.Event) {
        // Built-in serialization
        serialized, err := events.SerializeEvent(event)
        if err != nil {
            return
        }
        
        if filter.Match(event) {
            handler(serialized)
        }
    })
}
```

**Why It Makes Sense**:
- Event serialization is needed for logging, debugging, and external integrations
- Filtering reduces overhead for consumers only interested in specific events
- Event replay enables powerful debugging and testing scenarios

**Implementation Expectations**:
```go
// pkg/agent/events/serialization.go
package events

// All event types should implement this interface
type SerializableEvent interface {
    domain.Event
    MarshalJSON() ([]byte, error)
}

// Helper to serialize any event
func SerializeEvent(event domain.Event) (map[string]interface{}, error) {
    // Use reflection if event doesn't implement SerializableEvent
    // Include event type, timestamp, and all fields
}

// pkg/agent/events/filter.go
type Filter interface {
    Match(event domain.Event) bool
}

type PatternFilter struct {
    pattern *regexp.Regexp
}

// pkg/agent/events/recorder.go
type Recorder struct {
    storage EventStorage // Interface for different backends
}

func (r *Recorder) Record(event domain.Event) error {
    serialized, err := SerializeEvent(event)
    if err != nil {
        return err
    }
    return r.storage.Store(serialized)
}
```

### 5. Workflow Improvements

**Current State**: Workflows are defined in Go code and cannot be easily serialized.

**How We'd Use It**:
```go
// Allow scripts to define workflows declaratively
func (b *WorkflowBridge) createWorkflow(definition map[string]interface{}) (*workflow.WorkflowAgent, error) {
    // Deserialize workflow definition
    workflowDef, err := workflow.DeserializeDefinition(definition)
    if err != nil {
        return nil, err
    }
    
    // Support script-based steps
    for _, step := range workflowDef.Steps {
        if scriptStep, ok := step.(*workflow.ScriptStep); ok {
            scriptStep.Handler = b.createScriptHandler(scriptStep.Script)
        }
    }
    
    return workflow.NewWorkflowAgent(workflowDef)
}
```

**Why It Makes Sense**:
- Declarative workflows are easier to understand and modify
- Enables workflow storage, versioning, and sharing
- Supports visual workflow builders and no-code tools

**Implementation Expectations**:
```go
// pkg/agent/workflow/serialization.go
type WorkflowSerializer struct {
    format string // "json", "yaml"
}

func (s *WorkflowSerializer) Serialize(def *WorkflowDefinition) ([]byte, error) {
    // Convert to intermediate format that preserves all information
    intermediate := workflowToIntermediate(def)
    
    switch s.format {
    case "json":
        return json.Marshal(intermediate)
    case "yaml":
        return yaml.Marshal(intermediate)
    }
}

// pkg/agent/workflow/script_step.go
type ScriptStep struct {
    BaseStep
    Script   string
    Language string // "javascript", "lua", "tengo", "expr"
    Handler  ScriptHandler
}

type ScriptHandler interface {
    Execute(ctx context.Context, state *State, script string) (*State, error)
}

// Register script handlers globally
func RegisterScriptHandler(language string, handler ScriptHandler) {
    scriptHandlers[language] = handler
}
```

### 6. Error Handling Improvements

**Current State**: Errors are standard Go errors without structured context.

**How We'd Use It**:
```go
// In our bridge error handling
func (b *AgentBridge) executeAgent(ctx context.Context, agentID string, input interface{}) (interface{}, error) {
    agent, err := b.registry.GetAgent(agentID)
    if err != nil {
        // Structured error with context
        return nil, errors.NewAgentError("agent_not_found",
            errors.WithContext(map[string]interface{}{
                "agentID": agentID,
                "availableAgents": b.registry.ListAgents(),
            }),
            errors.WithRecovery(errors.RecoveryRetryWithBackoff),
        )
    }
    
    result, err := agent.Execute(ctx, input)
    if err != nil {
        // All go-llms errors should be serializable
        if serErr, ok := err.(errors.SerializableError); ok {
            errorData, _ := serErr.ToJSON()
            b.logger.Error("Agent execution failed", "error", errorData)
            
            // Try recovery
            if recovery := serErr.GetRecoveryStrategy(); recovery != nil {
                return recovery.Recover(ctx, err, func() (interface{}, error) {
                    return agent.Execute(ctx, input)
                })
            }
        }
    }
    
    return result, err
}
```

**Why It Makes Sense**:
- Structured errors improve debugging and error handling
- Serializable errors enable better logging and monitoring
- Recovery strategies provide resilience patterns

**Implementation Expectations**:
```go
// pkg/errors/types.go
package errors

// All go-llms errors should implement this
type SerializableError interface {
    error
    Code() string                      // Machine-readable error code
    Message() string                   // Human-readable message
    Context() map[string]interface{}   // Structured context
    ToJSON() ([]byte, error)          // Full serialization
    GetRecoveryStrategy() RecoveryStrategy
}

// Base implementation that all domain errors embed
type BaseError struct {
    code     string
    message  string
    context  map[string]interface{}
    cause    error
    recovery RecoveryStrategy
}

// Domain-specific errors embed BaseError
type AgentError struct {
    BaseError
    AgentID   string
    AgentType domain.AgentType
}

// pkg/errors/recovery.go
type RecoveryStrategy interface {
    Recover(ctx context.Context, err error, retry func() (interface{}, error)) (interface{}, error)
}

var (
    RecoveryRetryOnce        = &retryOnceStrategy{}
    RecoveryRetryWithBackoff = &exponentialBackoffStrategy{}
    RecoveryFailover         = &failoverStrategy{}
)

// Wrap all errors at API boundaries
func WrapError(err error, code string, opts ...ErrorOption) SerializableError {
    if serErr, ok := err.(SerializableError); ok {
        return serErr // Already wrapped
    }
    
    return &BaseError{
        code:    code,
        message: err.Error(),
        cause:   err,
        context: extractContext(err), // Extract context from error chain
    }
}
```

### 7. LLM Provider Enhancements

**Current State**: Provider capabilities are implicit and not discoverable.

**How We'd Use It**:
```go
// In our LLM bridge
func (b *LLMBridge) getProviderInfo(providerName string) (map[string]interface{}, error) {
    provider, err := b.registry.GetProvider(providerName)
    if err != nil {
        return nil, err
    }
    
    // Get standardized metadata
    metadata := provider.GetMetadata()
    
    return map[string]interface{}{
        "name":         metadata.Name(),
        "capabilities": metadata.GetCapabilities(), // ["streaming", "functions", "vision"]
        "models":       metadata.GetModels(),
        "constraints":  metadata.GetConstraints(),   // max_tokens, rate_limits, etc.
        "configSchema": metadata.GetConfigSchema(), // For UI generation
    }, nil
}

// Dynamic provider registration from scripts
func (b *LLMBridge) registerProvider(config map[string]interface{}) error {
    // Use provider templates
    template := providers.GetTemplate(config["type"].(string))
    provider := template.CreateProvider(config)
    
    return b.registry.RegisterProvider(provider)
}
```

**Why It Makes Sense**:
- Enables dynamic UI generation for provider configuration
- Allows capability-based provider selection
- Supports provider discovery and documentation

**Implementation Expectations**:
```go
// pkg/llm/providers/metadata.go
package providers

type ProviderMetadata interface {
    Name() string
    Description() string
    GetCapabilities() []Capability
    GetModels() []ModelInfo
    GetConstraints() Constraints
    GetConfigSchema() *schema.Schema
}

type Capability string

const (
    CapabilityStreaming      Capability = "streaming"
    CapabilityFunctionCalling Capability = "function_calling"
    CapabilityVision         Capability = "vision"
    CapabilityEmbeddings     Capability = "embeddings"
)

// All providers should implement
type MetadataProvider interface {
    domain.Provider
    GetMetadata() ProviderMetadata
}

// pkg/llm/providers/registry.go
type DynamicRegistry struct {
    *domain.ModelRegistry
    providers map[string]MetadataProvider
    mu        sync.RWMutex
}

func (r *DynamicRegistry) RegisterProvider(provider MetadataProvider) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Validate against schema
    metadata := provider.GetMetadata()
    if err := r.validateProvider(provider, metadata.GetConfigSchema()); err != nil {
        return err
    }
    
    r.providers[metadata.Name()] = provider
    return nil
}
```

### 8. Structured Output Support

**Current State**: No standardized way to parse LLM outputs into structured formats.

**How We'd Use It**:
```go
// In our structured output bridge
func (b *StructuredOutputBridge) parseResponse(response string, outputSchema map[string]interface{}) (interface{}, error) {
    // Convert script schema to domain schema
    schema, err := b.schemaConverter.Convert(outputSchema)
    if err != nil {
        return nil, err
    }
    
    // Use built-in parser
    parser := outputs.GetParser("json") // or "xml", "yaml"
    result, err := parser.Parse(response, schema)
    if err != nil {
        // Try recovery with schema-guided parsing
        result, err = parser.ParseWithRecovery(response, schema)
    }
    
    // Validate output
    validation, err := outputs.Validate(result, schema)
    if err != nil || !validation.Valid {
        return nil, fmt.Errorf("output validation failed: %v", validation.Errors)
    }
    
    return result, nil
}
```

**Why It Makes Sense**:
- Structured outputs are critical for reliable LLM applications
- Standardized parsing reduces errors and improves interoperability
- Schema-guided parsing can recover from minor formatting issues

**Implementation Expectations**:
```go
// pkg/llm/outputs/parser.go
package outputs

type Parser interface {
    Parse(response string, schema *domain.Schema) (interface{}, error)
    ParseWithRecovery(response string, schema *domain.Schema) (interface{}, error)
    Format() string
}

// Registry of parsers
var parsers = map[string]Parser{
    "json": &JSONParser{},
    "xml":  &XMLParser{},
    "yaml": &YAMLParser{},
}

// pkg/llm/outputs/json_parser.go
type JSONParser struct {
    strict bool // Whether to require exact schema compliance
}

func (p *JSONParser) ParseWithRecovery(response string, schema *domain.Schema) (interface{}, error) {
    // Try standard parsing first
    result, err := p.Parse(response, schema)
    if err == nil {
        return result, nil
    }
    
    // Extract JSON from markdown code blocks
    if json := extractJSONFromMarkdown(response); json != "" {
        result, err = p.Parse(json, schema)
        if err == nil {
            return result, nil
        }
    }
    
    // Try to fix common issues
    fixed := p.fixCommonIssues(response)
    result, err = p.Parse(fixed, schema)
    if err == nil {
        return result, nil
    }
    
    // Schema-guided extraction
    return p.extractWithSchema(response, schema)
}

// pkg/llm/outputs/validator.go
func Validate(output interface{}, schema *domain.Schema) (*domain.ValidationResult, error) {
    validator := validation.NewValidator()
    
    // Convert output to JSON for validation
    data, err := json.Marshal(output)
    if err != nil {
        return nil, err
    }
    
    return validator.Validate(schema, string(data))
}
```

### 9. Testing Infrastructure

**Current State**: Testing requires significant boilerplate and setup.

**How We'd Use It**:
```go
// In our bridge tests
func TestToolBridge_ExecuteTool(t *testing.T) {
    scenario := testing.NewScenario("tool_execution").
        WithMockProvider("gpt-4", testing.MockResponses{
            "What's the weather?": "I'll check the weather for you.",
        }).
        WithTool(testing.MockTool("weather", func(args map[string]interface{}) (interface{}, error) {
            return map[string]interface{}{"temp": 72, "conditions": "sunny"}, nil
        })).
        WithAgent(testing.AgentConfig{
            Type:  domain.AgentTypeLLM,
            Tools: []string{"weather"},
        }).
        WithInput("What's the weather?").
        ExpectToolCall("weather", testing.WithArgs(map[string]interface{}{})).
        ExpectOutput(testing.Contains("sunny"))
    
    scenario.Run(t)
}
```

**Why It Makes Sense**:
- Reduces test boilerplate and improves readability
- Enables consistent testing patterns across projects
- Supports complex scenario testing without complex setup

**Implementation Expectations**:
```go
// pkg/testing/scenario.go
package testing

type Scenario struct {
    name      string
    providers map[string]*MockProvider
    tools     map[string]domain.Tool
    agents    map[string]domain.Agent
    inputs    []interface{}
    expects   []Expectation
}

type ScenarioBuilder interface {
    WithMockProvider(name string, responses MockResponses) ScenarioBuilder
    WithTool(tool domain.Tool) ScenarioBuilder
    WithAgent(config AgentConfig) ScenarioBuilder
    WithInput(input interface{}) ScenarioBuilder
    ExpectOutput(matcher Matcher) ScenarioBuilder
    ExpectToolCall(toolName string, opts ...ExpectOption) ScenarioBuilder
    ExpectEvent(eventType domain.EventType, matcher Matcher) ScenarioBuilder
    Run(t *testing.T)
}

// pkg/testing/mocks.go
type MockProvider struct {
    *baseMockProvider
    responses map[string]string
    calls     []ProviderCall
}

func (m *MockProvider) Complete(ctx context.Context, req Request) (Response, error) {
    m.calls = append(m.calls, ProviderCall{Request: req, Time: time.Now()})
    
    // Match response based on input
    for pattern, response := range m.responses {
        if matched, _ := regexp.MatchString(pattern, req.Messages[len(req.Messages)-1].Content); matched {
            return Response{Content: response}, nil
        }
    }
    
    return Response{}, fmt.Errorf("no mock response for input")
}

// pkg/testing/assertions.go
type Matcher interface {
    Match(value interface{}) (bool, string)
}

var (
    Contains = func(substr string) Matcher { return &containsMatcher{substr} }
    Equals   = func(expected interface{}) Matcher { return &equalsMatcher{expected} }
    HasField = func(field string, value interface{}) Matcher { return &fieldMatcher{field, value} }
)
```

### 10. Documentation Infrastructure

**Current State**: Documentation is manual and often out of sync with code.

**How We'd Use It**:
```go
// Auto-generate documentation for our bridges
func (b *ToolsBridge) GenerateDocumentation() (Documentation, error) {
    docs := documentation.NewBuilder()
    
    // Document all available tools
    tools, err := b.discovery.ListTools()
    if err != nil {
        return nil, err
    }
    
    for _, tool := range tools {
        // Auto-generate OpenAPI spec
        openapi := documentation.GenerateOpenAPIForTool(tool)
        docs.AddOpenAPISpec(tool.Name, openapi)
        
        // Generate markdown documentation
        markdown := documentation.GenerateMarkdownForTool(tool)
        docs.AddMarkdown(tool.Name, markdown)
    }
    
    return docs.Build()
}
```

**Why It Makes Sense**:
- Ensures documentation stays in sync with code
- Enables automatic API documentation generation
- Supports multiple documentation formats for different audiences

**Implementation Expectations**:
```go
// pkg/docs/generator.go
package docs

type Generator interface {
    GenerateOpenAPI(items ...Documentable) (*OpenAPISpec, error)
    GenerateMarkdown(items ...Documentable) (string, error)
    GenerateJSON(items ...Documentable) ([]byte, error)
}

type Documentable interface {
    GetDocumentation() Documentation
}

// All major types should implement Documentable
type Documentation struct {
    Name        string
    Description string
    Examples    []Example
    Schema      *domain.Schema
    Metadata    map[string]interface{}
}

// pkg/docs/openapi.go
func GenerateOpenAPIForTool(tool ToolInfo) *OpenAPISpec {
    return &OpenAPISpec{
        OpenAPI: "3.0.0",
        Info: Info{
            Title:       tool.Name,
            Description: tool.Description,
            Version:     "1.0.0",
        },
        Paths: map[string]PathItem{
            "/execute": {
                Post: &Operation{
                    Summary:     fmt.Sprintf("Execute %s tool", tool.Name),
                    Description: tool.Description,
                    RequestBody: &RequestBody{
                        Required: true,
                        Content: map[string]MediaType{
                            "application/json": {
                                Schema: tool.Schema,
                            },
                        },
                    },
                    Responses: generateResponses(tool),
                },
            },
        },
    }
}
```

## Implementation Priority

Based on go-llmspell's immediate needs and broader ecosystem benefits:

1. **High Priority** (Blocks core functionality):
   - Schema Package Implementations
   - Tool Discovery Enhancements
   - Error Handling Improvements

2. **Medium Priority** (Significant quality of life improvements):
   - Bridge-Friendly Type System
   - Structured Output Support
   - Event System Enhancements

3. **Lower Priority** (Nice to have):
   - Workflow Improvements
   - Testing Infrastructure
   - Documentation Infrastructure
   - LLM Provider Enhancements

## Conclusion

These enhancements would transform go-llms from a powerful but rigid framework into a truly extensible platform. By implementing these features, go-llms would:

1. Enable easier integration with various programming languages and environments
2. Support dynamic, runtime-configurable applications
3. Provide better debugging and operational capabilities
4. Foster a richer ecosystem of extensions and integrations

Most importantly, these features benefit not just go-llmspell but any project building on go-llms, making it the go-to choice for LLM application development in Go.