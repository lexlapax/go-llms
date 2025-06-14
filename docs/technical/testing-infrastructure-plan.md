# Testing Infrastructure Plan for go-llms

## Overview

This document outlines a comprehensive plan for standardizing and enhancing the testing infrastructure in go-llms, with a focus on creating an exportable testing API that can be used by downstream projects and the go-llmspell bridge layer.

## Current State Analysis

### Existing Mock Implementations
- **Provider Mocks**: Located in `pkg/llm/provider/mock.go` and `pkg/testutils/mock_providers.go`
- **Tool Mocks**: Basic implementations in `pkg/testutils/mock_tools.go`
- **Agent Mocks**: Scattered across test files as local implementations
- **Pattern**: Function-based mocking with configurable behavior

### Issues Identified
1. Mock implementations are scattered across packages
2. No standardized mock interface or registry
3. Limited mock assertion and verification capabilities
4. No scenario-based testing framework
5. Mock implementations not easily exportable for downstream use

## Proposed Testing Infrastructure

### 1. Core Testing Package Structure
```
pkg/testutils/
├── mocks/
│   ├── provider.go      # MockProvider implementation
│   ├── tool.go          # MockTool implementation
│   ├── agent.go         # MockAgent implementation
│   ├── state.go         # MockState implementation
│   ├── event.go         # MockEventEmitter implementation
│   └── registry.go      # Mock registry for management
├── scenario/
│   ├── builder.go       # ScenarioBuilder fluent API
│   ├── assertions.go    # Assertion helpers
│   ├── matchers.go      # Matcher interface and implementations
│   └── runner.go        # Scenario execution engine
├── fixtures/
│   ├── providers.go     # Pre-configured provider mocks
│   ├── tools.go         # Pre-configured tool mocks
│   ├── agents.go        # Pre-configured agent mocks
│   └── states.go        # Common test states
└── helpers/
    ├── context.go       # Test context utilities
    ├── events.go        # Event capture and verification
    └── assertions.go    # Common assertion helpers
```

### 2. Mock Implementation Standards

#### MockProvider
```go
type MockProvider struct {
    // Configuration
    Name            string
    ResponsePattern map[string]Response  // Pattern-based responses
    CallHistory     []ProviderCall       // Call tracking
    
    // Behavior hooks
    OnGenerate      func(ctx context.Context, prompt string, options ...domain.Option) (string, error)
    OnStream        func(ctx context.Context, prompt string, options ...domain.Option) (<-chan string, error)
    OnGenerateSchema func(ctx context.Context, schemaName string, schema domain.Schema, options ...domain.Option) (string, error)
    
    // State
    mu              sync.RWMutex
    callCount       int
    lastError       error
}

type Response struct {
    Content  string
    Error    error
    Delay    time.Duration
    Metadata map[string]interface{}
}

type ProviderCall struct {
    Method    string
    Prompt    string
    Options   []domain.Option
    Response  string
    Error     error
    Timestamp time.Time
    Duration  time.Duration
}
```

#### MockTool
```go
type MockTool struct {
    // Configuration
    Info           domain.ToolInfo
    ResponseMap    map[string]interface{}  // Input pattern to response mapping
    CallHistory    []ToolCall
    
    // Behavior hooks
    OnExecute      func(ctx domain.ToolContext, input map[string]interface{}) (map[string]interface{}, error)
    OnValidate     func(input map[string]interface{}) error
    
    // Assertions
    ExpectedCalls  []ExpectedCall
    
    // State
    mu             sync.RWMutex
    executionCount int
}

type ToolCall struct {
    Input     map[string]interface{}
    Output    map[string]interface{}
    Error     error
    Context   domain.ToolContext
    Timestamp time.Time
    Duration  time.Duration
}
```

#### MockAgent
```go
type MockAgent struct {
    // Configuration
    ID             string
    ResponseQueue  []AgentResponse
    SubAgents      map[string]domain.Agent
    
    // Behavior hooks
    OnStart        func(ctx context.Context, state domain.State) (domain.State, error)
    OnStep         func(ctx context.Context, state domain.State) (domain.State, error)
    
    // Event tracking
    EmittedEvents  []domain.Event
    
    // State management
    StateHistory   []domain.State
    
    // State
    mu             sync.RWMutex
    currentIndex   int
}
```

### 3. Scenario Builder API

```go
type ScenarioBuilder struct {
    t              testing.TB
    providers      map[string]*MockProvider
    tools          map[string]*MockTool
    agents         map[string]*MockAgent
    initialState   domain.State
    expectations   []Expectation
    eventCapture   *EventCapture
}

// Fluent API
func NewScenario(t testing.TB) *ScenarioBuilder
func (s *ScenarioBuilder) WithMockProvider(name string, responses map[string]Response) *ScenarioBuilder
func (s *ScenarioBuilder) WithTool(tool *MockTool) *ScenarioBuilder
func (s *ScenarioBuilder) WithAgent(agent *MockAgent) *ScenarioBuilder
func (s *ScenarioBuilder) WithInput(key string, value interface{}) *ScenarioBuilder
func (s *ScenarioBuilder) ExpectOutput(matcher Matcher) *ScenarioBuilder
func (s *ScenarioBuilder) ExpectToolCall(toolName string, inputMatcher Matcher) *ScenarioBuilder
func (s *ScenarioBuilder) ExpectEvent(eventType string, dataMatcher Matcher) *ScenarioBuilder
func (s *ScenarioBuilder) Run() domain.State
```

### 4. Matcher System

```go
type Matcher interface {
    Match(value interface{}) (bool, string)  // Returns match result and error message
    Description() string
}

// Built-in matchers
func Equals(expected interface{}) Matcher
func Contains(substring string) Matcher
func HasField(field string, valueMatcher Matcher) Matcher
func MatchesJSON(pattern string) Matcher
func MatchesRegex(pattern string) Matcher
func AllOf(matchers ...Matcher) Matcher
func AnyOf(matchers ...Matcher) Matcher
func Not(matcher Matcher) Matcher
```

### 5. Event Testing Support

```go
type EventCapture struct {
    events    []domain.Event
    filters   []EventFilter
    mu        sync.RWMutex
}

func (e *EventCapture) CaptureEvent(event domain.Event)
func (e *EventCapture) GetEvents(eventType string) []domain.Event
func (e *EventCapture) AssertEventEmitted(t testing.TB, eventType string, matcher Matcher)
func (e *EventCapture) AssertEventCount(t testing.TB, eventType string, count int)
func (e *EventCapture) AssertNoEvents(t testing.TB, eventType string)
```

### 6. Test Fixtures

```go
// Pre-configured mocks for common scenarios
func ChatGPTMockProvider() *MockProvider
func ClaudeMockProvider() *MockProvider
func CalculatorMockTool() *MockTool
func WebSearchMockTool() *MockTool
func ResearchMockAgent() *MockAgent
func ErrorMockProvider(errorType string) *MockProvider
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)
1. Expand `pkg/testutils` package structure
2. Implement base mock types (MockProvider, MockTool, MockAgent)
3. Add call tracking and history
4. Create mock registry

### Phase 2: Scenario Builder (Week 2)
1. Implement ScenarioBuilder with fluent API
2. Add matcher interface and basic matchers
3. Create assertion helpers
4. Add event capture support

### Phase 3: Migration and Integration (Week 3)
1. Migrate existing mocks to new structure
2. Update existing tests to use new infrastructure
3. Create comprehensive examples
4. Write migration guide

### Phase 4: Documentation and Polish (Week 4)
1. Complete API documentation
2. Create best practices guide
3. Add performance benchmarks
4. Create testing cookbook

## Benefits

1. **Standardization**: Consistent mock implementations across the codebase
2. **Exportability**: Testing utilities available for downstream projects
3. **Bridge Support**: Scenario-based testing for go-llmspell integration
4. **Maintainability**: Centralized testing infrastructure
5. **Developer Experience**: Fluent API for writing expressive tests
6. **Debugging**: Comprehensive call tracking and history

## Success Criteria

1. All existing tests migrated to new infrastructure
2. 100% test coverage for testing package
3. Documentation complete with examples
4. Performance benchmarks showing no regression
5. Positive feedback from downstream users

## Future Enhancements

1. Property-based testing support
2. Fuzzing integration
3. Test data generation
4. Visual test reporting
5. Integration with CI/CD pipelines