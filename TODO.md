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
- [x] Phase 5: Multi-Agent System Enhancement - COMPLETED (June 5, 2025) - see TODO-DONE.md

- [ ] Phase 6: Advanced Features (MOVED TO PHASE 7) (low priority)
  - [ ] State persistence and serialization, present plan before implementation
  - [ ] Agent discovery and registry, present plan before implementation
  - [ ] Advanced merge strategies for parallel agents
  - [ ] Streaming support for long-running agents

- [ ] Phase 7: Migration and Testing (RENAMED FROM PHASE 6) - IN PROGRESS
  
  ## Week 1: Code Cleanup and Examples - COMPLETED (February 5, 2025) - see TODO-DONE.md
  ## Week 1-2: Examples Overhaul - COMPLETED (February 5, 2025) - see TODO-DONE.md
  
  ## Week 2: Testing Migration - IN PROGRESS (June 5, 2025)
  
  ### Integration Tests (tests/integration/) - COMPLETED (June 5, 2025) - see TODO-DONE.md
  
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
- [x] P2: Build useful built-in tools - COMPLETED - see TODO-DONE.md

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