# Testing Infrastructure Migration Guide

This guide helps you migrate from the old mock implementations to the new comprehensive testing infrastructure introduced in v0.3.5.9.

## Overview

The new testing infrastructure provides:
- **Fixtures**: Pre-configured mock objects for common scenarios
- **Scenario Builder**: Fluent API for complex test setups
- **Matcher System**: Flexible assertion capabilities
- **Helpers**: Utilities for context creation and state management

## Migration Paths

### Provider Mocks

#### Old Way (Deprecated)
```go
// OLD - Using MockProvider directly
mock := provider.NewMockProvider()
mock.WithGenerateFunc(func(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
    return "custom response", nil
})
```

#### New Way (Recommended)
```go
// NEW - Using fixtures
import "github.com/lexlapax/go-llms/pkg/testutils/fixtures"

// For ChatGPT-like responses
provider := fixtures.ChatGPTMockProvider()

// For Claude-like responses  
provider := fixtures.ClaudeMockProvider()

// For error testing
provider := fixtures.ErrorMockProvider("rate_limit")

// For slow response simulation
provider := fixtures.SlowMockProvider(time.Second * 2)

// For streaming responses
provider := fixtures.StreamingMockProvider()
```

#### Using Pattern-Based Responses
```go
// NEW - Pattern-based responses
provider := fixtures.ChatGPTMockProvider()
provider.WithPatternResponse("(?i).*weather.*", mocks.Response{
    Content: "Today is sunny with 75°F",
    Metadata: map[string]interface{}{
        "location": "test-city",
    },
})
```

### Agent Mocks

#### Old Way (Deprecated)
```go
// OLD - Manual mock creation
type MockAgent struct {
    *core.BaseAgentImpl
    runFunc func(ctx context.Context, state *domain.State) (*domain.State, error)
}

func NewMockAgent(name, description string) *MockAgent {
    return &MockAgent{
        BaseAgentImpl: core.NewBaseAgent(name, description, domain.AgentTypeCustom),
    }
}
```

#### New Way (Recommended)
```go
// NEW - Using fixtures
import "github.com/lexlapax/go-llms/pkg/testutils/fixtures"

// For simple testing
agent := fixtures.SimpleMockAgent()

// For research workflows
agent := fixtures.ResearchMockAgent() 

// For workflow execution
agent := fixtures.WorkflowMockAgent()

// For stateful testing
agent := fixtures.StatefulMockAgent()
```

### Tool Testing

#### Old Way (Deprecated)
```go
// OLD - Manual context creation
func createTestContext() *domain.ToolContext {
    mockAgent := &MockAgent{
        BaseAgentImpl: core.NewBaseAgent("test-agent", "Test agent", domain.AgentTypeCustom),
    }
    return domain.NewToolContext(
        context.Background(),
        domain.NewStateReader(domain.NewState()),
        mockAgent,
        "test-run",
    )
}
```

#### New Way (Recommended)
```go
// NEW - Using helpers
import "github.com/lexlapax/go-llms/pkg/testutils/helpers"

func createTestContext() *domain.ToolContext {
    return helpers.CreateTestToolContext()
}

// With custom state
func createTestContextWithState(data map[string]interface{}) *domain.ToolContext {
    return helpers.CreateToolContextWithState(data)
}

// Using tool fixtures
calculator := fixtures.CalculatorMockTool()
webSearch := fixtures.WebSearchMockTool()
fileManager := fixtures.FileMockTool()
```

### State Testing

#### Old Way (Manual)
```go
// OLD - Manual state creation
state := domain.NewState()
state.Set("test_key", "test_value")
// ... manual setup
```

#### New Way (Using Fixtures)
```go
// NEW - Using state fixtures
import "github.com/lexlapax/go-llms/pkg/testutils/fixtures"

// Empty state
state := fixtures.EmptyTestState()

// Basic test data
state := fixtures.BasicTestState()

// Workflow context
state := fixtures.WorkflowTestState()

// Conversation history
state := fixtures.ConversationTestState()

// Error conditions
state := fixtures.ErrorTestState()

// With artifacts
state := fixtures.StateWithArtifacts()

// With metadata
state := fixtures.StateWithMetadata()
```

### Scenario-Based Testing

#### Old Way (Manual Setup)
```go
// OLD - Manual test setup
func TestComplexScenario(t *testing.T) {
    // Lots of manual setup...
    provider := NewMockProvider()
    tool := &MockTool{}
    agent := &MockAgent{}
    
    // Manual execution and assertions...
    result, err := agent.Run(ctx, input)
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

#### New Way (Scenario Builder)
```go
// NEW - Using scenario builder
import "github.com/lexlapax/go-llms/pkg/testutils/scenario"

func TestComplexScenario(t *testing.T) {
    scenario.NewScenario(t).
        WithMockProvider("chatgpt", map[string]mocks.Response{
            "(?i).*hello.*": {Content: "Hello! How can I help?"},
        }).
        WithTool(fixtures.CalculatorMockTool()).
        WithAgent(fixtures.ResearchMockAgent()).
        WithInput("query", "research quantum computing").
        ExpectOutput("task_type", matchers.Equals("research")).
        ExpectOutput("query", matchers.Contains("quantum")).
        ExpectNoError().
        Run()
}
```

### Advanced Patterns

#### Error Injection
```go
// Error injection with specific rates
errorTool := fixtures.ErrorMockTool(0.3) // 30% error rate

// Provider with specific error types
errorProvider := fixtures.ErrorMockProvider("auth")
```

#### Call History Tracking
```go
// Track provider calls
provider := fixtures.ChatGPTMockProvider()
// ... use provider in tests
history := provider.GetCallHistory()
assert.Len(t, history, 3)
```

#### Event Testing
```go
// Event capture and assertion
import "github.com/lexlapax/go-llms/pkg/testutils/helpers"

eventCapture := helpers.NewEventCapture()
// ... run tests that emit events
events := eventCapture.GetEvents()

helpers.AssertEvents(t, events).
    HasType("agent.start").
    HasType("tool.execute").
    HasType("agent.complete").
    InOrder()
```

## Migration Checklist

### Phase 1: Update Imports
- [ ] Add `"github.com/lexlapax/go-llms/pkg/testutils/fixtures"`
- [ ] Add `"github.com/lexlapax/go-llms/pkg/testutils/helpers"`
- [ ] Add `"github.com/lexlapax/go-llms/pkg/testutils/scenario"`
- [ ] Add `"github.com/lexlapax/go-llms/pkg/testutils/matchers"`

### Phase 2: Replace Provider Mocks
- [ ] Replace `provider.NewMockProvider()` with fixtures
- [ ] Convert custom functions to pattern responses
- [ ] Update assertions to use new response format

### Phase 3: Replace Agent Mocks
- [ ] Replace manual MockAgent with fixtures
- [ ] Convert custom run functions to OnRun hooks
- [ ] Update context creation to use helpers

### Phase 4: Replace Tool Mocks
- [ ] Replace manual tool context creation
- [ ] Use pre-built tool fixtures where applicable
- [ ] Convert to scenario-based testing for complex cases

### Phase 5: Update Assertions
- [ ] Replace manual assertions with matchers
- [ ] Use scenario expectations where applicable
- [ ] Add call history verification

### Phase 6: Cleanup
- [ ] Remove unused imports
- [ ] Remove deprecated mock definitions
- [ ] Add migration comments for future reference

## Compatibility Layer

For gradual migration, compatibility wrappers are available:

```go
// Provider compatibility
import "github.com/lexlapax/go-llms/pkg/llm/provider"
adapter := provider.NewMockProviderCompat() // Wraps new infrastructure

// Agent compatibility  
import "github.com/lexlapax/go-llms/pkg/agent/tools"
agent := tools.NewMockAgentCompat("test", "description") // Wraps new infrastructure
```

## Benefits of Migration

1. **Reduced Boilerplate**: Fixtures eliminate repetitive setup code
2. **Better Patterns**: Realistic behavior simulation with pattern matching
3. **Consistent Testing**: Standardized mock behavior across the codebase
4. **Enhanced Debugging**: Call history tracking and detailed assertions
5. **Scenario Support**: Fluent API for complex test scenarios
6. **Future-Proof**: Built for extensibility and downstream integration

## Getting Help

- Check existing fixture implementations in `pkg/testutils/fixtures/`
- Review scenario examples in test files
- See helper documentation in `pkg/testutils/helpers/`
- For complex migrations, consider using the compatibility layer first