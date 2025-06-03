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
- [x] Phase 1: Core Infrastructure (Week 1-2) - COMPLETED
  - [x] Define new interfaces in `pkg/agent/domain/`
    - [x] base_agent.go - Core agent interface
    - [x] state.go - State management
    - [x] events.go - Event system
    - [x] artifact.go - Artifact types
    - [x] errors.go - Domain errors
    - [x] config.go - Configuration
  - [x] Implement base agent functionality
    - [x] pkg/agent/core/base_agent_impl.go
    - [x] State management utilities
    - [x] Event system implementation
    - [x] Agent registry implementation
  - [x] Create comprehensive tests
    - [x] state_test.go - State tests (all passing)
    - [x] events_test.go - Event tests (all passing)
    - [x] state_manager_test.go - State manager tests (all passing)
    - [x] event_dispatcher_test.go - Event dispatcher tests (all passing)

- [ ] Phase 2: LLM Agent Migration (Week 2-3)
  - [ ] Implement new LLMAgent based on current DefaultAgent
  - [ ] Migrate tool integration to new interface
  - [ ] Add state management capabilities
  - [ ] Implement agent hierarchy support
  - [ ] Remove old superfluos code, examples and tests

- [ ] Phase 3: Workflow Agents (Week 3-4)
  - [ ] Implement workflow agent base
  - [ ] Create SequentialAgent
  - [ ] Create ParallelAgent
  - [ ] Create ConditionalAgent
  - [ ] Create LoopAgent

- [ ] Phase 4: Agent-Tool Integration (Week 4)
  - [ ] Implement AgentTool wrapper
  - [ ] Create tool context system
  - [ ] Add bidirectional agent-tool conversion utilities

- [ ] Phase 5: Advanced Features (Week 5)
  - [ ] State persistence and serialization
  - [ ] Agent discovery and registry
  - [ ] Advanced merge strategies for parallel agents
  - [ ] Streaming support for long-running agents

- [ ] Phase 6: Migration and Testing (Week 5-6)
  - [ ] Remove old superfluos code, examples and tests
  - [ ] Create migration guide
  - [ ] Update all examples
  - [ ] Comprehensive testing
  - [ ] Performance benchmarking

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