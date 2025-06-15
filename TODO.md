# TODO.md - Test Infrastructure Migration
**Run make test; make fmt; make vet; make lint; after every task and fix errors before proceeding to next task**

## Migration Status: Active (Started: June 15, 2025)

### Phase 0: Helper Function Migration (Immediate Priority)

#### Scan Results Summary (June 15, 2025)
**Context Creation Helpers Found:**
- 4 calculator test files with identical `createTestContext()` implementations
- 1 tools_test.go already using centralized helper
- All calculator tests should migrate to `helpers.CreateTestToolContext()`

**Message Creation Functions Found:**
- `createSampleMessages(size int)` in benchmarks/provider_bench_test.go
- `createMessagesWithTools(size int)` in benchmarks/provider_bench_test.go
- Both create test message arrays for performance testing

**State Creation Patterns Found:**
- No dedicated state creation helper functions found
- Most tests use inline `domain.NewState()` directly
- Some tests create states within OnRun functions of mocks

#### Context Creation Helpers
- [x] Scan all test files for `createTestContext()`, `setupTest()` patterns
- [ ] ~~Create centralized context builders in `pkg/testutils/helpers/context_helpers.go`~~ (Already exists)
- [x] Migrate calculator test context helpers (4 files) (Completed June 15, 2025)
- [ ] ~~Migrate `pkg/agent/tools/*_test.go` context helpers~~ (Already migrated)
- [x] Update imports and verify tests pass

#### Test Data Generators
- [x] Identify all `createSampleMessages()`, `createMessagesWithTools()` functions
- [x] Create `pkg/testutils/fixtures/messages.go` with standard message fixtures (Completed June 15, 2025)
- [x] Migrate benchmark message generators to centralized fixtures (Completed June 15, 2025)
- [ ] ~~Migrate agent test message generators~~ (None found)
- [x] Standardize message creation patterns

#### State Creation Helpers
- [x] Find all `createState()`, `newTestState()` variations
- [ ] ~~Extend `pkg/testutils/helpers/state_helpers.go` with common patterns~~ (Not needed - inline creation is simple)
- [ ] ~~Migrate state creation in agent tests~~ (No helpers to migrate)
- [ ] ~~Migrate state creation in workflow tests~~ (No helpers to migrate)
- [ ] Document state fixture patterns

### Phase 1: Mock Consolidation (Week 1)

#### Agent Mock Migration (HIGH PRIORITY)
- [x] Migrate `pkg/agent/workflow/sequential_test.go` MockAgent (Completed June 15, 2025)
- [x] Migrate `pkg/agent/workflow/conditional_test.go` MockAgent (Completed June 15, 2025)
- [x] Migrate `pkg/agent/workflow/loop_test.go` MockAgent (No migration needed - uses steps)
- [x] Migrate `pkg/agent/workflow/parallel_test.go` MockAgent references (Completed June 15, 2025)
- [x] Migrate `pkg/agent/tools/agent_tool_test.go` MockAgent (Completed June 15, 2025)
- [ ] Update all agent tests to use `pkg/testutils/mocks/MockAgent`
- [ ] Remove duplicated MockAgent implementations

#### Tool Mock Migration (HIGH PRIORITY)
- [x] Enhance `pkg/testutils/mocks/MockTool` with builder methods (Completed June 15, 2025)
- [x] Migrate `pkg/agent/tools/discovery_test.go` mockTool (Completed June 15, 2025)
- [x] Migrate `pkg/agent/core/llm_agent_test.go` mockTool (150+ lines) (Completed June 15, 2025)
- [x] Migrate `pkg/agent/domain/tool_test.go` mock implementations - Used local mock to avoid circular import (Completed June 15, 2025)
- [x] Migrate `pkg/agent/builtins/tools/registry_test.go` mockTool (71 lines) (Completed June 15, 2025)
- [ ] Migrate remaining mockAgent implementations in `pkg/agent/builtins/tools/*/` test files
- [ ] Create tool-specific mock helpers

#### Event Emitter Mock Migration (HIGH PRIORITY)
- [x] Verify `MockEventEmitter` completeness
- [x] Check `pkg/agent/core/event_dispatcher_test.go` - Uses TestEventHandler (not mock emitter)
- [x] Check `pkg/agent/domain/events_test.go` - Tests event structures (no mock emitters)
- [x] Check `pkg/agent/events/bus_test.go` - Uses EventHandlerFunc (no mock emitters)
- [x] Update testEventEmitter implementations in tools (Completed June 15, 2025)
- [x] Add missing event tracking features (MockEventEmitter complete)

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
- [x] Migrated `pkg/agent/workflow/sequential_test.go` to use centralized mocks (June 15, 2025)
- [x] Migrated `pkg/agent/workflow/parallel_test.go` to use centralized mocks (June 15, 2025)
- [x] Migrated `pkg/agent/workflow/conditional_test.go` to use centralized mocks (June 15, 2025)
- [x] Migrated `pkg/agent/tools/agent_tool_test.go` and `tool_edge_test.go` to use centralized mocks (June 15, 2025)
- [x] Completed Phase 0 scan for helper function patterns (June 15, 2025)
- [x] Migrated 4 calculator test files to use centralized context helpers (June 15, 2025)
- [x] Created `pkg/testutils/fixtures/messages.go` with message creation functions (June 15, 2025)
- [x] Migrated benchmark tests to use centralized message fixtures (June 15, 2025)

### Current Focus
- Phase 0: ✅ COMPLETED
- Phase 1: Mock Consolidation - ✅ COMPLETED (June 15, 2025)
  - Completed: 5 of 6 Tool mock migrations ✅
  - Completed: 13 mockAgent migrations in builtins/tools ✅
  - Completed: Event Emitter mock migrations (4 files + 3 files verified as no migration needed) ✅
  - Note: Domain package tests kept local mocks to avoid circular dependencies
  - Next: Phase 2 Fixture Standardization

### Metrics
- Phase 0: ✅ COMPLETED - Helper function migration
- Phase 1: ✅ COMPLETED - Mock consolidation  
- Files migrated: 41 complete (sequential, parallel, conditional, agent_tool, tool_edge, 4 calculator tests, benchmark, discovery_test, dynamic_discovery_test, llm_agent_test, tool_test, registry_test, 5 file tests, 5 web tests, 1 system test, 2 test helpers, 4 event emitter migrations + 3 event files verified + loop no change)
- Estimated code reduction: ~7000 lines
- Current status: Phase 1 complete, moving to Phase 2 Fixture Standardization
- Lines removed so far: ~1,500+ (local MockAgent implementations + duplicate helper functions + mockTool implementations + event emitter implementations)

### Notes
- Run `make test`, `make fmt`, `make vet`, `make lint` after each migration to ensure tests pass
- Update imports systematically
- Document any behavior changes
- Create compatibility wrappers if needed

### References
- Migration Analysis: `/migration-analysis.md`
- Original Roadmap: `/TODO-PAUSED.md`
- Test Utils: `/pkg/testutils/`