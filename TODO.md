# TODO.md - Test Infrastructure Migration

## Migration Status: Active (Started: June 15, 2025)

### Phase 0: Helper Function Migration (Immediate Priority)

#### Context Creation Helpers
- [ ] Scan all test files for `createTestContext()`, `setupTest()` patterns
- [ ] Create centralized context builders in `pkg/testutils/helpers/context_helpers.go`
- [ ] Migrate `pkg/agent/core/*_test.go` context helpers
- [ ] Migrate `pkg/agent/tools/*_test.go` context helpers
- [ ] Update imports and verify tests pass

#### Test Data Generators
- [ ] Identify all `createSampleMessages()`, `createMessagesWithTools()` functions
- [ ] Create `pkg/testutils/fixtures/messages.go` with standard message fixtures
- [ ] Migrate provider test message generators
- [ ] Migrate agent test message generators
- [ ] Standardize message creation patterns

#### State Creation Helpers
- [ ] Find all `createState()`, `newTestState()` variations
- [ ] Extend `pkg/testutils/helpers/state_helpers.go` with common patterns
- [ ] Migrate state creation in agent tests
- [ ] Migrate state creation in workflow tests
- [ ] Document state fixture patterns

### Phase 1: Mock Consolidation (Week 1)

#### Agent Mock Migration (HIGH PRIORITY)
- [ ] ~~Migrate `pkg/agent/workflow/sequential_test.go` MockAgent~~ (Started)
- [ ] Migrate `pkg/agent/workflow/conditional_test.go` MockAgent
- [ ] Migrate `pkg/agent/workflow/loop_test.go` MockAgent
- [ ] Migrate `pkg/agent/workflow/parallel_test.go` MockAgent references
- [ ] Migrate `pkg/agent/tools/agent_tool_test.go` MockAgent
- [ ] Update all agent tests to use `pkg/testutils/mocks/MockAgent`
- [ ] Remove duplicated MockAgent implementations

#### Tool Mock Migration (HIGH PRIORITY)
- [ ] Enhance `pkg/testutils/mocks/MockTool` with builder methods
- [ ] Migrate `pkg/agent/tools/discovery_test.go` mockTool
- [ ] Migrate `pkg/agent/core/llm_agent_test.go` mockTool (150+ lines)
- [ ] Migrate `pkg/agent/domain/tool_test.go` mock implementations
- [ ] Migrate built-in tool mocks in `pkg/agent/builtins/tools/*/`
- [ ] Create tool-specific mock helpers

#### Event Emitter Mock Migration (HIGH PRIORITY)
- [ ] Verify `MockEventEmitter` completeness
- [ ] Migrate `pkg/agent/core/event_dispatcher_test.go` mocks
- [ ] Migrate `pkg/agent/domain/events_test.go` emitter mocks
- [ ] Migrate `pkg/agent/events/bus_test.go` mock subscribers
- [ ] Update testEventEmitter implementations in tools
- [ ] Add missing event tracking features

### Phase 2: Fixture Standardization (Week 2)

#### Provider Fixtures (MEDIUM PRIORITY)
- [ ] Audit existing provider fixtures in `pkg/testutils/fixtures/providers.go`
- [ ] Add provider-specific configuration fixtures
- [ ] Create streaming provider fixtures
- [ ] Create error scenario fixtures
- [ ] Migrate inline provider configurations

#### Tool Fixtures (MEDIUM PRIORITY)
- [ ] Extend tool fixtures for each built-in category
- [ ] Create file operation tool fixtures
- [ ] Create web tool fixtures
- [ ] Create data processing tool fixtures
- [ ] Document fixture usage patterns

#### Agent Fixtures (MEDIUM PRIORITY)
- [ ] Create workflow agent fixtures
- [ ] Create stateful agent fixtures
- [ ] Create error handling agent fixtures
- [ ] Create concurrent agent fixtures
- [ ] Migrate inline agent setups

### Phase 3: Scenario Builder Adoption (Week 3)

#### Complex Test Migration (MEDIUM PRIORITY)
- [ ] Identify tests with 5+ setup steps
- [ ] Create scenario templates for common patterns
- [ ] Migrate integration tests to scenario builder
- [ ] Migrate workflow tests to scenario builder
- [ ] Document scenario builder patterns

#### Integration Test Patterns (MEDIUM PRIORITY)
- [ ] Standardize provider integration test setup
- [ ] Create multi-component scenario templates
- [ ] Migrate end-to-end tests
- [ ] Add scenario builder examples
- [ ] Create scenario builder cookbook

### Phase 4: Matcher Standardization (Week 4)

#### Custom Assertion Migration (LOW PRIORITY)
- [ ] Audit custom assertion logic in tests
- [ ] Create domain-specific matchers
- [ ] Migrate string assertions to matchers
- [ ] Migrate state assertions to matchers
- [ ] Migrate event assertions to matchers

#### Event Assertion Patterns (LOW PRIORITY)
- [ ] Standardize event verification
- [ ] Create event sequence matchers
- [ ] Create event data matchers
- [ ] Document matcher usage
- [ ] Add matcher examples

## Progress Tracking

### Completed Items
- [x] Created `pkg/testutils/helpers/agent_helpers.go` with reusable mock agent creators (June 15, 2025)
- [x] Started migration of `pkg/agent/workflow/sequential_test.go` (June 15, 2025)

### Current Focus
- Migrating workflow package MockAgent implementations to use centralized mocks

### Metrics
- Total test files to migrate: 176
- Files migrated: 1 (in progress)
- Estimated code reduction: ~7000 lines
- Current status: 0.5% complete

### Notes
- Run `make test` after each migration to ensure tests pass
- Update imports systematically
- Document any behavior changes
- Create compatibility wrappers if needed

### References
- Migration Analysis: `/migration-analysis.md`
- Original Roadmap: `/TODO-PAUSED.md`
- Test Utils: `/pkg/testutils/`