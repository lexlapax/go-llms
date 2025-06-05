# Go-LLMs Project TODOs

## Features
- [ ] Add Model Context Protocol Client support for Agents
- [ ] Add Model Context Protocol Server support for Workflows or Agents

## Testing & Performance
- [ ] Performance profiling and optimization:
  - [ ] Phase 1: Baseline Profiling Infrastructure (Prerequisites)
    - [ ] P1: Create benchmark harness for A/B testing optimizations (REVISIT)
    - [ ] P2: Implement visualization for memory allocation patterns (REVISIT)
    - [ ] P2: Create real-world test scenarios for end-to-end performance (REVISIT)

  - [ ] Phase 2: High-Impact Optimizations (Quick Wins)
    (All P0 and P1 items completed - see TODO-DONE.md)

  - [ ] Phase 3: Advanced Optimizations (After Initial Improvements)
    - [ ] P1: Implement adaptive channel buffer sizing based on usage patterns (REVISIT)
    - [ ] P1: Add pool prewarming for high-throughput scenarios (REVISIT)
    - [ ] P1: Reduce redundant property iterations in schema processing (REVISIT)
    - [ ] P2: Implement more granular locking in cached objects (REVISIT)
    - [ ] P2: Optimize zero-initialization patterns for pooled objects (REVISIT)
    - [ ] P2: Introduce buffer pooling for string builders (REVISIT)

  - [ ] Phase 4: Integration and Validation (Finalization)
    - [ ] P0: Document performance improvements with metrics (REVISIT)
    - [ ] P0: Verify optimizations in high-concurrency scenarios (REVISIT)
    - [ ] P1: Create benchmark comparison charts for before/after (REVISIT)
    - [ ] P1: Implement regression testing to prevent performance degradation (REVISIT)
    - [ ] P2: Add performance acceptance criteria to CI pipeline (REVISIT)

## Architecture & Built-in Components for next release

### Agent Architecture Restructuring (NEW - HIGH PRIORITY)

- [x] Phase 1 & 1.5: Core Infrastructure - COMPLETED (February 3, 2025) - see TODO-DONE.md
- [x] Phase 2: LLM Agent Migration - COMPLETED (February 3, 2025) - see TODO-DONE.md
- [x] Phase 3: Workflow Agents - COMPLETED (February 3, 2025) - see TODO-DONE.md
- [x] Phase 4: Agent-Tool Integration (Week 4) - COMPLETED (February 2025) - see TODO-DONE.md

- [ ] Phase 5: Multi-Agent System Enhancement (HIGH PRIORITY - Inspired by Google ADK)
  
  ## Background
  After analyzing Google's Agent Development Kit (ADK), we identified key features that would significantly improve our multi-agent capabilities:
  - Automatic sub-agent to tool conversion
  - Dynamic agent delegation via LLM
  - Shared state between parent and child agents
  - Simplified API for multi-agent creation
  
  ## Phase 5.1: Core Handoff Implementation (1-2 days)
  - [ ] Complete handoff execution in pkg/agent/domain/handoff.go
    - [ ] Implement Execute() method using agent registry
    - [ ] Add global registry access pattern (GetGlobalRegistry())
    - [ ] Handle state transformation and result merging
    - [ ] Add error handling for missing target agents
    - [ ] Test handoff execution flow with unit tests
  
  ## Phase 5.2: Auto-Tool Registration (1 day)
  - [ ] Modify BaseAgentImpl.AddSubAgent to auto-register sub-agents as tools
    - [ ] Create AgentTool wrapper automatically
    - [ ] Add tool to parent if parent is LLMAgent
    - [ ] Ensure tool names don't conflict
  - [ ] Add built-in "transfer_to_agent" tool to LLMAgent
    - [ ] Tool searches sub-agents by name
    - [ ] Executes handoff to selected sub-agent
    - [ ] Returns sub-agent execution result
  - [ ] Update tool discovery to include sub-agent tools
  
  ## Phase 5.3: Shared State Context (1 day)
  - [ ] Implement SharedStateContext for parent-child state sharing
    - [ ] Create SharedStateContext struct with parent and local state
    - [ ] Implement Get() with fallback to parent state
    - [ ] Add Set() that only affects local state
    - [ ] Add MergeToParent() for explicit parent updates
  - [ ] Update RunContext to support shared state
  - [ ] Modify agent execution to use shared state when available
  - [ ] Add configuration option for state inheritance behavior
  
  ## Phase 5.4: API Simplification (1 day)
  - [ ] Create simplified constructors matching Google ADK patterns
    - [ ] NewLLMAgentWithSubAgents(name, provider, subAgents)
    - [ ] Builder pattern: agent.WithSubAgents(agents...)
  - [ ] Add convenience methods
    - [ ] agent.TransferTo(agentName, reason)
    - [ ] agent.GetSubAgentByName(name)
  - [ ] Update agent creation to be more declarative
  
  ## Phase 5.5: Examples and Documentation (1 day)
  - [ ] Create new example: agent-sub-agents
    - [ ] Show automatic tool registration
    - [ ] Demonstrate transfer_to_agent usage
    - [ ] Show shared state in action
  - [ ] Update agent-handoff example to use new implementation
  - [ ] Create migration guide for existing code
  - [ ] Document new patterns in technical docs
  
  ## Expected Outcomes
  - Sub-agents automatically available as tools to parent agents
  - LLM can dynamically choose which sub-agent to delegate to
  - State automatically shared between parent and children
  - Much simpler API for creating multi-agent systems
  - Feature parity with Google ADK's multi-agent approach

- [ ] Phase 6: Advanced Features (MOVED TO PHASE 7) (low priority)
  - [ ] State persistence and serialization, present plan before implementation
  - [ ] Agent discovery and registry, present plan before implementation
  - [ ] Advanced merge strategies for parallel agents
  - [ ] Streaming support for long-running agents

- [ ] Phase 7: Migration and Testing (RENAMED FROM PHASE 6) - Week 1 COMPLETED (February 5, 2025)
  
  ## Week 1: Code Cleanup and Examples - COMPLETED (February 5, 2025)
  
  ### Day 1-2: Discovery and Analysis - COMPLETED
  - [x] Scan entire codebase for deprecated patterns and create removal list
  - [x] Create inventory document of all changes needed (PHASE6_MIGRATION_INVENTORY.md)
  
  ### Day 3-4: Code Removal and Cleanup - COMPLETED
  - [x] Remove deprecated code
  - [x] Update build tags and remove migration tags
  - [x] Migrate test files with workflow_migration tag (10 files)
  
  ### Day 5: Documentation Updates - COMPLETED
  - [x] Update all code documentation to reflect new patterns
  - [x] Update technical documentation
  
  ## Week 1-2: Examples Overhaul - COMPLETED (February 5, 2025)
  
  ### Example Analysis and Categorization - COMPLETED
  - [x] Analyze all examples in cmd/examples/
  
  ### Example Updates - COMPLETED
  - [x] Update basic examples
    - [x] simple - verified, basic structured output example (no agent updates needed)
    - [x] agent-simple-llm - updated to use correct state fields (user_input/output)
    - [x] provider-convenience (renamed from convenience) - removed agent code, focused on provider-level utilities
  - [x] Update provider examples (verify all use new patterns)
    - [x] provider-openai - verified, uses direct provider API (correct)
    - [x] provider-anthropic - verified, uses direct provider API (correct)
    - [x] provider-gemini - verified, uses direct provider API (correct)
    - [x] provider-openai-compatible - verified, uses direct provider API (correct)
    - [x] provider-multimodal - verified, uses direct provider API (correct)
    - [x] provider-multi (renamed from multi) - kept as provider-level example, added note pointing to workflow-multi-provider
    - [x] provider-consensus (renamed from consensus) - kept as provider-level example, added note pointing to workflow-multi-provider
  - [x] Update/verify workflow examples
    - [x] workflow-sequential - verified, uses new architecture correctly
    - [x] workflow-parallel - verified, uses new architecture correctly
    - [x] workflow-conditional - fixed, added workflow.NewAgentStep() public API
    - [x] workflow-loop - fixed, added workflow.NewAgentStep() public API
    - [x] workflow-hooks - verified, uses new architecture correctly
    - [x] agent-workflow-as-tool - already updated
  - [x] Update advanced examples
    - [x] agent-structured-output - verified, already updated
    - [x] agent-custom-calculator - verified, already updated
    - [x] agent-error-handling - fixed compilation errors, updated for new architecture
    - [x] agent-state-persistence - fixed compilation errors
    - [x] agent-guardrails (renamed from guardrails) - renamed for consistency
  - [x] Rename utility examples for consistency
    - [x] utils-profiling (renamed from profiling) - utility package example
    - [x] utils-modelinfo - already correctly named
  - [x] Create structured output category
    - [x] structured-schema (renamed from schema) - schema generation and validation
    - [x] structured-coercion (renamed from coercion) - type coercion in validation
  
  ### New Examples Added - COMPLETED
  - [x] Create state persistence example (created agent-state-persistence/)
  - [x] Create advanced error handling example (created agent-error-handling/)
  - [x] Create complex workflow composition example (created workflow-composition/)
  - [x] Create workflow-multi-provider example (created workflow-multi-provider/)
  - [x] Create guardrails example (created agent-guardrails/)
  - [x] Create multi-agent coordination example (created agent-multi-coordination/)
  - [x] Create agent handoff example (created agent-handoff/)
  
  ### Example Cleanup - COMPLETED
  - [x] Remove obsolete examples (removed 3 empty directories)
  - [x] Fixed compilation errors in error-handling and state-persistence examples
  - [x] Added workflow.NewAgentStep() public API to fix workflow examples
  - [x] Renamed examples for consistent categorization:
    - agent-* (agent features)
    - workflow-* (workflow patterns)
    - provider-* (provider-level features)
    - builtins-* (built-in tools)
    - utils-* (utility packages)
    - structured-* (structured output features)
  
  ### Example Documentation - COMPLETED
  - [x] Ensure all examples have proper README.md
  - [x] Verify all examples compile and run (most compile, 2 have simplified implementations)
  - [x] Update cmd/examples/README.md with new categorization and examples
  
  ## Week 2: Testing Migration - POSTPONED (Focus on Phase 5 Multi-Agent Enhancement)
  
  ### Integration Tests (tests/integration/)
  - [ ] Analyze current integration tests
    - [ ] List tests using old patterns
    - [ ] Identify missing test coverage for new features
    - [ ] Plan test updates
  - [ ] Update integration tests
    - [ ] agent_test.go - migrate to core.LLMAgent
    - [ ] provider tests - ensure work with new patterns
    - [ ] multimodal tests - verify updated
    - [ ] tool integration tests - verify updated
  - [ ] Add new integration tests
    - [ ] Workflow agent integration tests
    - [ ] Agent-tool conversion tests
    - [ ] State management tests
    - [ ] Hook integration tests
  
  ### Stress Tests (tests/stress/)
  - [ ] Update stress tests to new architecture
    - [ ] agent_stress_test.go - use core.LLMAgent
    - [ ] provider_stress_test.go - verify updated
    - [ ] pool_stress_test.go - verify updated
    - [ ] structured_stress_test.go - verify updated
  - [ ] Add new stress tests
    - [ ] Workflow agent stress tests
    - [ ] Concurrent agent execution tests
    - [ ] Memory leak detection tests
    - [ ] State management stress tests
  
  ### Benchmark Updates
  - [x] Move benchmarks/ directory to tests/benchmarks/ - COMPLETED
  - [x] Update all benchmarks to new architecture (done during migration)
  - [ ] Verify and update specific benchmarks
    - [ ] agent_bench_test.go - use core.LLMAgent
    - [ ] provider_bench_test.go - verify updated
    - [ ] tools benchmarks - verify updated
    - [ ] consensus benchmarks - update if needed
  - [ ] Add new benchmarks
    - [ ] Agent creation performance
    - [ ] State management overhead
    - [ ] Tool execution performance
    - [ ] Workflow agent performance
    - [ ] Hook execution overhead
  
  ### Test Documentation
  - [ ] Update testing documentation
  - [ ] Document new test patterns
  - [ ] Create testing best practices guide

### Previous Built-in Components Plan
- [ ] P2: Build useful built-in tools
  - [x] Phase 2.6: Feed Process Tools (Completed - see TODO-DONE.md)

- [ ] P3: Build useful built-in agents (Phase 3 - POSTPONED until after architecture restructuring)
  - [ ] Text Agents
    - [ ] TextSummarize - intelligent summarization using LLM
    - [ ] TextExtract - extract structured data from text
    - [ ] TextAnalyze - sentiment, entities, keywords
    - [ ] TextTranslate - language translation using LLM
  - [ ] Research Agents:
    - [ ] WebResearcher - web research with source tracking
    - [ ] DocumentAnalyzer - analyze documents and PDFs
    - [ ] FactChecker - verify claims against sources
  - [ ] Coding Agents:
    - [ ] CodeReviewer - review code for issues
    - [ ] TestGenerator - generate tests from code
    - [ ] DocWriter - generate documentation
  - [ ] Data Agents:
    - [ ] DataAnalyst - analyze datasets and generate insights
    - [ ] ReportGenerator - create formatted reports
    - [ ] DataCleaner - clean and validate data
  - [ ] Feed Agents:
    - [ ] NewsMonitor - monitor news feeds for keywords and topics using LLM
    - [ ] FeedAggregator - aggregate and deduplicate content from multiple feeds
    - [ ] FeedSummarizer - summarize feed content using LLM
    - [ ] ContentCurator - curate and categorize feed content using LLM
  
- [ ] P4: Build useful multi-agent workflows (Phase 4 - PENDING)
  - [ ] Core Patterns:
    - [ ] Pipeline - sequential processing
    - [ ] MapReduce - parallel processing with aggregation
    - [ ] Consensus - multi-agent agreement
    - [ ] Retry - with exponential backoff
  - [ ] Example Workflows:
    - [ ] ResearchWorkflow - research → verify → summarize → report
    - [ ] CodeReviewWorkflow - analyze → review → suggest → document
    - [ ] DataPipeline - ingest → clean → analyze → visualize
    
- [ ] Fix identified cross-link issues (path inconsistencies, broken links) (REVISIT)
- [ ] Perform final consistency check across all documentation (REVISIT)
- [ ] API refinement based on usage feedback
- [ ] Final review and preparation for stable release

## Completed Tasks
See TODO-DONE.md for all completed tasks