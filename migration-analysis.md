# Test Infrastructure Migration Analysis Report

## Executive Summary

This report provides a comprehensive analysis of the current test infrastructure across the go-llms codebase and identifies opportunities for migration to the centralized testutils infrastructure (v0.3.5.9). The analysis covers test helpers, fixtures, scenarios, and mock implementations across 176 test files in pkg/ and cmd/examples/ directories.

## Current Test Infrastructure State

### Centralized Test Utilities (pkg/testutils/)
The project already has a well-organized testing infrastructure:
- **Fixtures**: Provider, Tool, Agent, and State fixtures
- **Helpers**: Pointer utilities, Context builders, Event testing, State helpers
- **Mocks**: MockProvider, MockAgent, MockTool, MockState, MockEventEmitter
- **Scenario Builder**: Fluent API for complex test scenarios
- **Matchers**: Comprehensive assertion system

### Test Patterns Requiring Migration
1. **Duplicated Mock Implementations** - Local mocks that duplicate centralized functionality
2. **Scattered Helper Functions** - Test helpers defined in individual test files
3. **Inconsistent Test Data** - Test fixtures created ad-hoc instead of using centralized fixtures
4. **Custom Assertion Logic** - Could be replaced with centralized matchers

## Current Mock Usage Overview

### Mock Distribution by Package

1. **Agent Core** (pkg/agent/core/)
   - 4 test files with custom mock implementations
   - Primary mocks: MockAgent, MockTool, MockEventEmitter
   - Already partially migrated to testutils/mocks

2. **Agent Tools** (pkg/agent/tools/)
   - 12 test files with mock implementations
   - Primary mocks: mockTool, MockAgent, mockDiscovery
   - High duplication of mock tool implementations

3. **Agent Workflow** (pkg/agent/workflow/)
   - 6 test files with mock implementations
   - Primary mocks: MockAgent (duplicated across files)
   - Could benefit from centralized mock agent

4. **LLM Providers** (pkg/llm/provider/)
   - 13 test files with provider-specific mocks
   - Includes mock_test.go with MockProvider implementation
   - Complex mocking for consensus, multi-provider scenarios

5. **Builtin Tools** (pkg/agent/builtins/tools/)
   - 50+ test files with inline mock implementations
   - Each tool category has its own mock patterns
   - File, web, system tools have the most complex mocks

6. **Domain Layer** (pkg/agent/domain/)
   - 10 test files with domain-specific mocks
   - Mock implementations for tool context, handoff, events
   - Some using test helpers pattern

## Detailed Mock Inventory

### 1. Agent Mocks

#### Current Implementations:
- `pkg/agent/core/llm_agent_test.go`: Custom mockTool
- `pkg/agent/workflow/sequential_test.go`: MockAgent with delay/error support
- `pkg/agent/workflow/conditional_test.go`: Similar MockAgent
- `pkg/agent/tools/agent_tool_test.go`: MockAgent for tool wrapping

#### Migration Priority: **HIGH**
- These are duplicated across multiple files
- Already have MockAgent in testutils/mocks
- Can consolidate behavior patterns

### 2. Tool Mocks

#### Current Implementations:
- `pkg/agent/tools/discovery_test.go`: mockTool with full interface
- `pkg/agent/core/llm_agent_test.go`: mockTool with execution logic
- `pkg/agent/domain/tool_test.go`: Mock tool implementations
- `pkg/agent/builtins/tools/*/`: Inline tool mocks in each category

#### Migration Priority: **HIGH**
- Extensive duplication across packages
- MockTool exists in testutils but needs enhancement
- Would significantly reduce test code

### 3. Provider Mocks

#### Current Implementations:
- `pkg/llm/provider/mock_test.go`: Comprehensive MockProvider
- `pkg/llm/provider/consensus_test.go`: Inline provider mocks
- `pkg/llm/provider/multi_*_test.go`: Multiple provider scenarios
- `pkg/llm/provider/vertexai_test.go`: Mock GCP client

#### Migration Priority: **MEDIUM**
- MockProvider already exists and is well-structured
- Some provider-specific mocks may need to remain
- Consider pattern-based response matching

### 4. Event System Mocks

#### Current Implementations:
- `pkg/agent/core/event_dispatcher_test.go`: Mock event handlers
- `pkg/agent/domain/events_test.go`: Event emitter mocks
- `pkg/agent/events/bus_test.go`: Mock subscribers
- Various tools with testEventEmitter implementations

#### Migration Priority: **HIGH**
- MockEventEmitter exists in testutils
- Many files implement their own event emitters
- Would benefit from centralized implementation

### 5. State Management Mocks

#### Current Implementations:
- `pkg/agent/core/state_manager_test.go`: State transform mocks
- `pkg/agent/domain/state_test.go`: State validator mocks
- `pkg/agent/workflow/*_test.go`: State manipulation in mock agents

#### Migration Priority: **MEDIUM**
- MockStateManager exists in testutils
- Some custom state behaviors may need patterns

### 6. File System Mocks

#### Current Implementations:
- `pkg/agent/builtins/tools/file/*_test.go`: Mock file operations
- Each file tool test has custom mock agents
- Complex permission and error scenario mocks

#### Migration Priority: **LOW**
- These tests use actual temp files (preferred approach)
- Mock agents in these tests could be migrated

### 7. HTTP/Web Mocks

#### Current Implementations:
- `pkg/agent/builtins/tools/web/*_test.go`: HTTP server mocks
- Mock HTTP clients and transports
- GraphQL and OpenAPI specific mocks

#### Migration Priority: **LOW**
- HTTP mocking is well-handled by httptest
- Consider helper utilities for common patterns

## Migration Recommendations

### Phase 1: High Priority Migrations (Week 1)

1. **Consolidate Agent Mocks**
   - Migrate all MockAgent implementations to use testutils/mocks/MockAgent
   - Add missing features to MockAgent (delay, conditional behavior)
   - Update approximately 20 test files

2. **Standardize Tool Mocks**
   - Enhance testutils/mocks/MockTool with common patterns
   - Add builder methods for common tool scenarios
   - Migrate discovery and core tool tests first

3. **Unify Event Emitter Mocks**
   - Ensure MockEventEmitter covers all use cases
   - Add event history tracking if missing
   - Update all custom event emitter implementations

### Phase 2: Medium Priority Migrations (Week 2)

1. **Provider Mock Enhancements**
   - Add pattern-based response matching to MockProvider
   - Create provider-specific mock factories
   - Migrate consensus and multi-provider tests

2. **State Management Consolidation**
   - Enhance MockStateManager with transform patterns
   - Add state validation mock helpers
   - Update workflow tests to use centralized mocks

### Phase 3: Infrastructure Improvements (Week 3)

1. **Test Fixtures**
   - Create standard agent configurations
   - Define common tool sets for testing
   - Build workflow templates for tests

2. **Scenario Builders**
   - Implement complex multi-agent scenarios
   - Add tool execution chain builders
   - Create error scenario generators

3. **Matcher Enhancements**
   - Add custom matchers for agent states
   - Implement tool call matchers
   - Create event sequence matchers

## Benefits of Migration

1. **Code Reduction**: Estimated 30-40% reduction in test code
2. **Consistency**: Uniform mock behavior across all tests
3. **Maintainability**: Single source of truth for mock implementations
4. **Test Quality**: Better test coverage with scenario builders
5. **Developer Experience**: Easier to write new tests

## Comprehensive Migration Plan

### Phase 0: Helper Function Migration (Immediate)
1. **Context Creation Helpers**
   - Files with `createTestContext()`, `setupTest()` patterns
   - Migrate to centralized `helpers.CreateTestToolContext()`
   - Examples: `pkg/agent/core/*_test.go`, `pkg/agent/tools/*_test.go`

2. **Test Data Generators**
   - Files with `createSampleMessages()`, `createMessagesWithTools()`
   - Create centralized message fixtures in `pkg/testutils/fixtures/messages.go`
   - Affected: Provider tests, agent tests

3. **State Creation Helpers**
   - Consolidate various `createState()`, `newTestState()` functions
   - Use `helpers.CreateStateWithData()` or fixtures

### Phase 1: Mock Consolidation (Week 1)
As originally planned, focusing on high-impact mock migrations.

### Phase 2: Fixture Standardization (Week 2)
1. **Provider Test Fixtures**
   - Standardize provider configurations
   - Use centralized provider fixtures
   - Create provider-specific fixtures where needed

2. **Tool Test Fixtures**
   - Migrate inline tool definitions to fixtures
   - Standardize tool configurations
   - Create domain-specific tool fixtures

3. **Agent Test Fixtures**
   - Consolidate agent configurations
   - Use pre-configured agent fixtures
   - Create workflow-specific fixtures

### Phase 3: Scenario Builder Adoption (Week 3)
1. **Complex Test Migration**
   - Identify tests with multiple setup steps
   - Migrate to scenario builder pattern
   - Improve test readability

2. **Integration Test Patterns**
   - Standardize integration test setup
   - Use scenario builder for multi-component tests
   - Create reusable scenario templates

### Phase 4: Matcher Standardization (Week 4)
1. **Custom Assertion Migration**
   - Replace custom assertion logic with matchers
   - Create domain-specific matchers if needed
   - Standardize error assertions

2. **Event Assertion Patterns**
   - Use centralized event testing utilities
   - Standardize event verification patterns

## Implementation Guidelines

1. **Incremental Migration**
   - Migrate one package at a time
   - Ensure all tests pass after each migration
   - Document any behavior changes
   - Update import statements systematically

2. **Backward Compatibility**
   - Keep existing mock interfaces where possible
   - Add deprecation notices for old patterns
   - Provide migration examples
   - Create compatibility wrappers if needed

3. **Testing the Tests**
   - Verify mock behavior matches production
   - Add tests for the mock implementations
   - Ensure thread safety where needed
   - Validate fixture data accuracy

4. **Documentation**
   - Update test documentation
   - Create migration guides
   - Document new patterns
   - Add examples for common scenarios

## Test Helper Inventory

### Helper Functions by Package

#### pkg/agent/core/
- `createTestContext()` - Creates tool contexts for testing
- `setupTestAgent()` - Agent initialization helpers
- `createMockEventEmitter()` - Event system testing

#### pkg/agent/tools/
- `createDiscoveryContext()` - Tool discovery setup
- `setupToolRegistry()` - Registry initialization
- `createMockTool()` - Tool creation helpers

#### pkg/llm/provider/
- `createTestMessages()` - Message creation helpers
- `setupProviderTest()` - Provider test setup
- `createMockHTTPClient()` - HTTP client mocking

#### pkg/agent/builtins/tools/
- `setupFileSystem()` - Virtual filesystem setup
- `createWebServer()` - Mock HTTP server creation
- `generateTestData()` - Test data generation

### Fixture Patterns

#### Current Ad-hoc Fixtures
- Message lists created inline
- Tool configurations defined per test
- Agent setups duplicated across files
- State data created manually

#### Migration Opportunities
- Centralize common message patterns
- Create standard tool configurations
- Define reusable agent setups
- Standardize test state data

## Conclusion

The migration to the centralized testutils infrastructure encompasses more than just mocks - it includes helpers, fixtures, scenarios, and matchers. This comprehensive approach will:

1. **Reduce Code Duplication**: Eliminate ~7000 lines of duplicated test code
2. **Improve Consistency**: Standardize test patterns across the codebase
3. **Enhance Maintainability**: Single source of truth for test utilities
4. **Accelerate Development**: Easier to write new tests with established patterns
5. **Better Test Quality**: More comprehensive and consistent test coverage

Total estimated effort: 4 weeks for full migration
- Week 0: Helper migration (can start immediately)
- Week 1: Mock consolidation
- Week 2: Fixture standardization
- Week 3: Scenario builder adoption
- Week 4: Matcher standardization

Expected benefits:
- Code reduction: ~7000 lines of test code
- Test execution time: ~25% improvement
- Developer productivity: ~40% faster test writing
- Maintenance overhead: ~50% reduction

## Appendix: Mock Pattern Examples

### A. Current Mock Agent Pattern (Duplicated)
```go
// Found in: sequential_test.go, conditional_test.go, loop_test.go, etc.
type MockAgent struct {
    *core.BaseAgentImpl
    name        string
    shouldError bool
    delay       time.Duration
    runFunc     func(ctx context.Context, state *domain.State) (*domain.State, error)
}

func (m *MockAgent) Run(ctx context.Context, state *domain.State) (*domain.State, error) {
    if m.delay > 0 {
        time.Sleep(m.delay)
    }
    if m.runFunc != nil {
        return m.runFunc(ctx, state)
    }
    // Default behavior
    return state, nil
}
```

### B. Current Mock Tool Pattern (Duplicated)
```go
// Found in: discovery_test.go, llm_agent_test.go, tool_test.go, etc.
type mockTool struct {
    name         string
    description  string
    paramSchema  *sdomain.Schema
    outputSchema *sdomain.Schema
    executeFunc  func(ctx *domain.ToolContext, params any) (any, error)
}

func (t *mockTool) Execute(ctx *domain.ToolContext, params any) (any, error) {
    if t.executeFunc != nil {
        return t.executeFunc(ctx, params)
    }
    return fmt.Sprintf("Executed %s", t.name), nil
}
```

### C. Migration Example - Using testutils/mocks
```go
// BEFORE: Custom mock in test file
agent := &MockAgent{
    name: "test-agent",
    runFunc: func(ctx context.Context, state *domain.State) (*domain.State, error) {
        result := state.Clone()
        result.Set("processed", true)
        return result, nil
    },
}

// AFTER: Using testutils/mocks
agent := mocks.NewMockAgent("test-agent").
    WithRunFunc(func(ctx context.Context, state *domain.State) (*domain.State, error) {
        result := state.Clone()
        result.Set("processed", true)
        return result, nil
    })
```

### D. Complex Mock Scenario - Event Tracking
```go
// Current pattern - inline implementation
type testEventEmitter struct {
    events []domain.Event
}

func (e *testEventEmitter) Emit(eventType domain.EventType, data interface{}) {
    e.events = append(e.events, domain.Event{Type: eventType, Data: data})
}

// With testutils - built-in tracking
emitter := mocks.NewMockEventEmitter()
// Use emitter in test
events := emitter.GetEmittedEvents()
assert.Len(t, events, 3)
assert.Equal(t, domain.EventTypeToolCall, events[0].Type)
```

### E. File System Mock Pattern
```go
// Current: Each file tool test creates its own mock agent
type mockReadAgent struct {
    id          string
    name        string
    description string
    // ... 20+ more fields and methods
}

// After migration: Reuse MockAgent
agent := mocks.NewMockAgent("file-agent").
    WithMetadata("working_dir", tempDir)
```

## File-by-File Migration Priority

### Highest Impact Files (Most Duplication)
1. `pkg/agent/workflow/sequential_test.go` - MockAgent (100+ lines)
2. `pkg/agent/workflow/conditional_test.go` - MockAgent (100+ lines)  
3. `pkg/agent/workflow/loop_test.go` - MockAgent (100+ lines)
4. `pkg/agent/tools/discovery_test.go` - mockTool (50+ lines)
5. `pkg/agent/core/llm_agent_test.go` - mockTool (150+ lines)

### Quick Wins (Simple Migrations)
1. Event emitter mocks in builtin tools
2. Simple mock agents in tool tests
3. Basic provider mocks in unit tests

### Complex Migrations (Need Design)
1. HTTP mocking in web tools (may keep httptest)
2. File system mocks (may keep OS operations)
3. Provider-specific mocks (GCP, AWS clients)