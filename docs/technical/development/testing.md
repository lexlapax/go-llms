# Testing Guide

> **[Documentation Home](../README.md) / [Development](README.md) / Testing**

## Overview

This guide covers the testing philosophy, infrastructure, and best practices for go-llms. We maintain high test coverage and use a comprehensive testing approach that includes unit tests, integration tests, and end-to-end tests.

## Testing Philosophy

### Principles
1. **Test at Multiple Levels** - Unit, integration, and e2e tests serve different purposes
2. **Mock External Dependencies** - Use mocks for predictable unit tests
3. **Real Integration Tests** - Test with actual providers when possible
4. **Fast Feedback** - Unit tests should run quickly
5. **Comprehensive Coverage** - Aim for >80% coverage on critical paths

### Test Categories

| Category | Purpose | Dependencies | Speed | When to Run |
|----------|---------|--------------|-------|-------------|
| Unit | Test individual components | Mocked | Fast (<1s) | Every commit |
| Integration | Test component interactions | Some real | Medium | Before merge |
| E2E | Test complete workflows | All real | Slow | Release |

## Testing Infrastructure

### Directory Structure
```
tests/
├── unit/                 # Unit tests (if separate from packages)
├── integration/         # Integration tests
│   ├── provider_test.go
│   ├── agent_test.go
│   └── workflow_test.go
└── e2e/                # End-to-end tests

pkg/
└── */                  # Each package contains its own unit tests
    ├── something.go
    └── something_test.go
```

### Test Utilities
Located in `pkg/testutils/`:

![Testing Infrastructure](../images/testing-infrastructure.svg)
*Figure 1: Testing infrastructure showing the organization of mocks, fixtures, scenarios, and utilities*

## Writing Tests

### Unit Tests

#### Basic Structure
```go
package mypackage_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMyFunction(t *testing.T) {
    // Arrange
    input := "test input"
    expected := "expected output"
    
    // Act
    result, err := MyFunction(input)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

#### Table-Driven Tests
```go
func TestCalculator(t *testing.T) {
    tests := []struct {
        name      string
        operation string
        a, b      float64
        expected  float64
        wantErr   bool
    }{
        {
            name:      "addition",
            operation: "add",
            a:         5,
            b:         3,
            expected:  8,
        },
        {
            name:      "division by zero",
            operation: "divide",
            a:         10,
            b:         0,
            wantErr:   true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Calculate(tt.operation, tt.a, tt.b)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.expected, result)
}
    }
}
```

#### Testing with Mocks
```go
func TestAgentWithMockProvider(t *testing.T) {
    // Create mock provider
    mockProvider := mocks.NewMockProvider()
    mockProvider.WithGenerateFunc(func(ctx context.Context, prompt string, opts ...Option) (Response, error) {
        assert.Contains(t, prompt, "weather")
        return Response{Content: "Sunny, 72°F"}, nil
}
    
    // Create agent with mock
    agent := core.NewLLMAgent("test", "model", core.LLMDeps{
        Provider: mockProvider,
}
    
    // Test agent behavior
    state := domain.NewState()
    state.Set("user_input", "What's the weather?")
    
    result, err := agent.Run(context.Background(), state)
    require.NoError(t, err)
    
    output, _ := result.Get("output")
    assert.Contains(t, output, "Sunny")
    
    // Verify mock was called
    assert.Equal(t, 1, mockProvider.CallCount())
}
```

### Integration Tests

#### Provider Integration Test
```go
func TestOpenAIProviderIntegration(t *testing.T) {
    // Skip if no API key
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        t.Skip("OPENAI_API_KEY not set, skipping integration test")
    }
    
    // Create real provider
    provider := provider.NewOpenAIProvider(apiKey, "gpt-3.5-turbo")
    
    // Test with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Test generation
    response, err := provider.Generate(ctx, "Say hello in JSON format")
    require.NoError(t, err)
    assert.NotEmpty(t, response.Content)
    
    // Verify response is JSON
    var data map[string]interface{}
    err = json.Unmarshal([]byte(response.Content), &data)
    assert.NoError(t, err)
}
```

#### Multi-Component Integration
```go
func TestAgentWithToolsIntegration(t *testing.T) {
    // Create provider (mock or real based on env)
    provider := getTestProvider(t)
    
    // Create agent
    agent := core.NewLLMAgent("assistant", "model", core.LLMDeps{
        Provider: provider,
}
    
    // Add tools
    agent.AddTool(createTestCalculatorTool())
    agent.AddTool(createTestWeatherTool())
    
    // Test complex interaction
    state := domain.NewState()
    state.Set("user_input", "What's 25 times 4? Also, what's the weather in NYC?")
    
    result, err := agent.Run(context.Background(), state)
    require.NoError(t, err)
    
    // Verify both tools were used
    output := result.Get("output").(string)
    assert.Contains(t, output, "100") // Calculator result
    assert.Contains(t, output, "NYC") // Weather mention
}
```

### End-to-End Tests

```go
func TestEndToEndWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping e2e test in short mode")
    }
    
    // Use real providers
    openai := provider.NewOpenAIProvider(os.Getenv("OPENAI_API_KEY"), "gpt-4")
    anthropic := provider.NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"), "claude-3")
    
    // Create workflow
    workflow := workflow.NewSequentialAgent("research-workflow")
    
    // Research agent (OpenAI)
    researcher := core.NewLLMAgent("researcher", "gpt-4", core.LLMDeps{
        Provider: openai,
}
    researcher.SetSystemPrompt("You are a research assistant")
    
    // Analysis agent (Anthropic)
    analyst := core.NewLLMAgent("analyst", "claude-3", core.LLMDeps{
        Provider: anthropic,
}
    analyst.SetSystemPrompt("You are a data analyst")
    
    // Build workflow
    workflow.AddAgent(researcher)
    workflow.AddAgent(analyst)
    
    // Execute
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()
    
    state := domain.NewState()
    state.Set("topic", "Impact of AI on software development")
    
    result, err := workflow.Run(ctx, state)
    require.NoError(t, err)
    
    // Verify workflow completed
    assert.NotNil(t, result)
    assert.Contains(t, result.Get("output"), "AI")
}
```

## Testing Patterns

### Testing Async Operations
```go
func TestAsyncAgent(t *testing.T) {
    agent := createTestAgent()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    eventChan, err := agent.RunAsync(ctx, domain.NewState())
    require.NoError(t, err)
    
    // Collect events
    var events []domain.Event
    for event := range eventChan {
        events = append(events, event)
    }
    
    // Verify event sequence
    require.Len(t, events, 3)
    assert.Equal(t, domain.EventAgentStart, events[0].Type)
    assert.Equal(t, domain.EventStateChange, events[1].Type)
    assert.Equal(t, domain.EventAgentComplete, events[2].Type)
}
```

### Testing Error Scenarios
```go
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name        string
        setupMock   func(*mocks.MockProvider)
        expectedErr string
    }{
        {
            name: "provider error",
            setupMock: func(m *mocks.MockProvider) {
                m.WithError(errors.NewProviderError("test", errors.ErrorTypeRateLimit, "rate limited"))
            },
            expectedErr: "rate limited",
        },
        {
            name: "context timeout",
            setupMock: func(m *mocks.MockProvider) {
                m.WithDelay(10 * time.Second)
            },
            expectedErr: "context deadline exceeded",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := mocks.NewMockProvider()
            tt.setupMock(mock)
            
            agent := createAgentWithProvider(mock)
            
            ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
            defer cancel()
            
            _, err := agent.Run(ctx, domain.NewState())
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectedErr)
}
    }
}
```

### Testing Concurrency
```go
func TestConcurrentAgentExecution(t *testing.T) {
    agent := createTestAgent()
    
    // Run multiple agents concurrently
    const numGoroutines = 10
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            state := domain.NewState()
            state.Set("id", id)
            
            _, err := agent.Run(context.Background(), state)
            if err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for errors
    for err := range errors {
        t.Errorf("Concurrent execution error: %v", err)
    }
}
```

## Test Fixtures and Helpers

### Creating Test Fixtures
```go
// In pkg/testutils/fixtures/providers.go
func OpenAIProviderWithResponses(responses ...string) *provider.OpenAIProvider {
    mock := mocks.NewMockHTTPClient()
    
    for i, response := range responses {
        mock.OnRequest(i).Return(mockOpenAIResponse(response))
    }
    
    return provider.NewOpenAIProvider("test-key", "gpt-4",
        domain.WithHTTPClient(mock),
    )
}

// Usage in tests
func TestWithFixture(t *testing.T) {
    provider := fixtures.OpenAIProviderWithResponses(
        "First response",
        "Second response",
    )
    
    // Use provider in test
}
```

### Test Helpers
```go
// In pkg/testutils/helpers/state.go
func AssertStateContains(t *testing.T, state *domain.State, key string, expected interface{}) {
    t.Helper()
    
    actual, exists := state.Get(key)
    require.True(t, exists, "State missing key: %s", key)
    assert.Equal(t, expected, actual)
}

// Usage
func TestStateManipulation(t *testing.T) {
    state := domain.NewState()
    state.Set("result", 42)
    
    helpers.AssertStateContains(t, state, "result", 42)
}
```

## Running Tests

### Command Line
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...

# Run only unit tests (exclude integration)
go test -short ./...

# Run specific package
go test ./pkg/agent/core

# Run specific test
go test -run TestAgentExecution ./pkg/agent/core

# Verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Makefile Targets
```bash
# Run unit tests
make test

# Run all tests (including integration)
make test-all

# Run with race detector
make test-race

# Generate coverage report
make coverage

# Run benchmarks
make bench
```

### Environment Variables
```bash
# Skip integration tests
SKIP_INTEGRATION=1 go test ./...

# Use specific providers
OPENAI_API_KEY=sk-... go test ./tests/integration

# Enable debug logging
DEBUG=1 go test -v ./...

# Set test timeout
TEST_TIMEOUT=10m go test ./...
```

## Best Practices

### 1. Test Organization
- Keep tests close to the code they test
- Use `_test` package for black-box testing
- Group related tests using subtests
- Use descriptive test names

### 2. Test Independence
- Each test should be independent
- Clean up resources in defer statements
- Don't rely on test execution order
- Use fresh state for each test

### 3. Mock Usage
- Mock external dependencies
- Don't mock what you own
- Keep mocks simple and focused
- Verify mock expectations

### 4. Assertions
- Use `require` for critical checks
- Use `assert` for non-critical checks
- Provide meaningful error messages
- Test one thing per assertion

### 5. Performance
- Keep unit tests fast (<100ms)
- Use `t.Parallel()` where appropriate
- Skip slow tests with `-short` flag
- Benchmark critical paths

## Debugging Tests

### Verbose Logging
```go
func TestWithLogging(t *testing.T) {
    if testing.Verbose() {
        log.SetLevel(log.DebugLevel)
    }
    
    // Test code with debug logging
}
```

### Test Helpers
```go
func TestWithHelper(t *testing.T) {
    // t.Helper() marks function as test helper
    assertSomething(t, value)
}

func assertSomething(t *testing.T, value interface{}) {
    t.Helper() // Errors report caller's line, not this line
    
    if value == nil {
        t.Error("Value should not be nil")
    }
}
```

### Debugging Goroutines
```go
func TestGoroutineLeaks(t *testing.T) {
    defer goleak.VerifyNone(t)
    
    // Test code that might leak goroutines
}
```

## Continuous Integration

### GitHub Actions Configuration
```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: |
        go test -race -coverprofile=coverage.out ./...
        
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

## Next Steps

- Review [Contributing Guide](../../../CONTRIBUTING.md) for development workflow
- See [Documentation Style Guide](../../../CONTRIBUTING-DOCS.md) for documentation standards
- Check test examples in `/tests` directory
- Explore mock implementations in `/pkg/testutils/mocks`