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

- [ ] Phase 5: Advanced Features (Week 5) (low priority)
  - [ ] State persistence and serialization, present plan before implementation
  - [ ] Agent discovery and registry, present plan before implementation
  - [ ] Advanced merge strategies for parallel agents
  - [ ] Streaming support for long-running agents

- [ ] Phase 6: Migration and Testing (Week 5-6) (high priority)
  
  ## Week 1: Code Cleanup and Examples
  
  ### Day 1-2: Discovery and Analysis
  - [x] Scan entire codebase for deprecated patterns and create removal list
    - [x] Search for "deprecated", "backward compatibility", "TODO: remove", "legacy" comments
    - [x] Find all references to workflow.Agent, DefaultAgent, UnoptimizedDefaultAgent (NONE FOUND!)
    - [x] Identify backward compatibility shims and workarounds
    - [x] List all test files using old patterns (10 files with workflow_migration tag)
  - [x] Create inventory document of all changes needed (PHASE6_MIGRATION_INVENTORY.md)
  
  ### Day 3-4: Code Removal and Cleanup - COMPLETED
  - [x] Remove deprecated code
    - [x] Clean up legacy comments (multi.go, processor.go) 
    - [x] Remove workflow_migration exclusion from .golangci.yml
    - [x] Remove any remaining workflow.Agent references (NONE FOUND)
    - [x] Remove old agent implementations if any remain (NONE FOUND)
    - [x] Keep backward compatibility code for API stability
    - [x] Migrated all test files (no obsolete files to remove)
    - [x] No unused imports or dead code found
  - [x] Update build tags and remove migration tags (workflow_migration removed)
  - [x] Migrate test files with workflow_migration tag (10 files) - COMPLETED
    - [x] Migrated 3 benchmark files:
      - agent_bench_test.go
      - tools_bench_test.go
      - tools_builtin_bench_test.go
    - [x] Migrated 6 integration test files:
      - agent_test.go
      - agent_edge_cases_test.go
      - agent_errors_test.go
      - anthropic_e2e_test.go
      - e2e_test.go
      - gemini_agent_e2e_test.go
    - [x] Migrated 1 stress test file:
      - agent_stress_test.go
  
  ### Day 5: Documentation Updates - COMPLETED
  - [x] Update all code documentation to reflect new patterns
    - [x] Remove references to old APIs in comments
    - [x] Update package-level documentation (getting-started.md, built-in-components.md, custom-agents.md)
    - [x] Fix any outdated examples in doc comments (agent.md API documentation)
  - [x] Update technical documentation
    - [x] Updated user guide documentation to use core.LLMAgent
    - [x] Updated API documentation to reflect new architecture
    - [x] Updated README.md with correct examples
  
  ## Week 1-2: Examples Overhaul
  
  ### Example Analysis and Categorization - COMPLETED
  - [x] Analyze all examples in cmd/examples/
    - [x] List examples that work as-is with new architecture (25+ examples)
    - [x] List examples that need updates (7 examples identified)
    - [x] List examples that should be removed (3 empty directories)
    - [x] Identify gaps - missing examples for new features (10 new examples needed)
  
  ### Example Updates
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
  - [ ] Update/verify tool examples
    - [ ] agent-tools-conversion - already updated
    - [ ] agent-llm-builtin-tools - already updated
    - [ ] agent-advanced-toolcontext - already updated
    - [ ] builtins-* - verify all updated
  - [x] Update/verify workflow examples
    - [x] workflow-sequential - verified, uses new architecture correctly
    - [x] workflow-parallel - verified, uses new architecture correctly
    - [x] workflow-conditional - fixed, added workflow.NewAgentStep() public API
    - [x] workflow-loop - fixed, added workflow.NewAgentStep() public API
    - [x] workflow-hooks - verified, uses new architecture correctly
    - [x] agent-workflow-as-tool - already updated
  - [x] Update advanced examples
    - [x] provider-multi (renamed from multi) - kept as provider-level example, added note pointing to workflow-multi-provider
    - [x] provider-consensus (renamed from consensus) - kept as provider-level example, added note pointing to workflow-multi-provider
    - [x] agent-structured-output - verified, already updated
    - [x] agent-custom-calculator - verified, already updated
  
  ### New Examples to Add - revamp this based on example updates above
  - [x] Create state persistence example (created agent-state-persistence/)
  - [x] Create advanced error handling example (created agent-error-handling/)
  - [x] Create complex workflow composition example (created workflow-composition/)
  - [x] Create workflow-multi-provider example (created workflow-multi-provider/)
  - [ ] Create multi-agent coordination example
  - [ ] Create agent handoff example
  - [x] Create guardrails example (created agent-guardrails/)
  
  ### Example Cleanup
  - [x] Remove obsolete examples (removed 3 empty directories)
  - [ ] Ensure all examples have proper README.md
  - [ ] Verify all examples compile and run
  - [ ] Update cmd/examples/README.md with changes
  
  ## Week 2: Testing Migration
  
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
  
  ### Benchmark Migration - COMPLETED
  - [x] Move benchmarks/ directory to tests/benchmarks/
  - [x] Update all benchmarks to new architecture (done during migration)
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