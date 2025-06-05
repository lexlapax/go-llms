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
  - [ ] Update integration tests to use core.LLMAgent (tests/integration/) (REVISIT)
  - [ ] Update benchmarks to use core.LLMAgent (benchmarks/) (REVISIT)
  - [ ] Update stress tests to use core.LLMAgent (tests/stress/) (REVISIT)

- [x] Phase 3: Workflow Agents - COMPLETED (February 3, 2025) - see TODO-DONE.md

- [ ] Phase 4: Agent-Tool Integration (Week 4)
  - [x] Phase 4.1: Revisit existing Tool interface and see if it needs updates based on current agent/workflow
  - [x] Phase 4.2: Implement AgentTool wrapper
  - [x] Phase 4.3: Create tool context system - COMPLETED (February 2025)
  - [x] Phase 4.4: Fix all builtins-* examples to use the redone tools with ToolContext - COMPLETED (February 2025)
  - [x] Phase 4.5: Add bidirectional agent-tool conversion utilities - COMPLETED (February 2025)
    - [x] Create pkg/agent/tools/conversion_utils.go with core utilities
    - [x] Implement Event Dispatcher Integration (HIGH PRIORITY)
      - [x] NewToolAgentWithEvents for full event support
      - [x] CreateEventForwardingToolContext for event forwarding
      - [x] Update ToolAgent to support event dispatcher injection
    - [x] Implement Registry Integration Utilities (HIGH PRIORITY)
      - [x] RegisterAgentAsTool to convert and register agents
      - [x] ConvertToolCategoryToAgents for batch conversion
      - [x] RegisterAgentsAsTools for bulk registration
      - [x] Integration with existing tool registry
    - [x] Implement Automatic Schema Mapping (MEDIUM PRIORITY)
      - [x] DeriveToolSchemaFromAgent for auto-generation
      - [x] ValidateConversionCompatibility for round-trip validation
      - [x] GenerateSmartMappers based on schema analysis
    - [x] Implement Common Conversion Patterns (MEDIUM PRIORITY)
      - [x] WrapLLMAgentAsTool helper
      - [x] WrapWorkflowAgentAsTool helper
      - [x] CreateToolChainFromAgents for chaining
      - [x] RoundTripConvert with validation
    - [x] Implement Advanced Mapping Utilities (LOW PRIORITY)
      - [x] CreatePathMapper for path-based extraction
      - [x] CreateTypeConversionMapper for type conversions
      - [x] CreateNestedStateMapper for complex structures
    - [x] Create Testing Utilities
      - [x] AssertRoundTripEquivalence test helper
      - [x] CreateMockAgentForTool for testing
      - [x] ValidateAgentToolConversion validator
    - [x] Create comprehensive tests for all utilities
    - [x] Create examples demonstrating conversion utilities
    - [x] Update documentation with conversion patterns
  - [x] Phase 4.6: Ensure all current tools in builtins/tools work with agents - COMPLETED (February 2025)
    - [x] Created comprehensive agent-llm-builtin-tools example demonstrating all tool categories
    - [x] Fixed compilation issues with LLMAgent creation and hook implementation
    - [x] Demonstrated proper system prompts for each tool category
    - [x] Included tool call monitoring with hooks
  - [x] Phase 4.7: Create examples demonstrating enhanced tool capabilities - COMPLETED (February 2025)
    - [x] Created agent-advanced-toolcontext example demonstrating Advanced ToolContext features
    - [x] Showed event emission with progress reporting, custom events, and error handling
    - [x] Demonstrated retry mechanism with tools detecting retry attempts
    - [x] Implemented state access allowing tools to read agent state
    - [x] Created comprehensive example combining all ToolContext features
    - [x] Fixed compilation issues with EventHandler and EmitCustom usage
  - [x] Create an example of an agent calling another agent that's wrapped as a tool - COMPLETED (February 2025)
    - [x] Design multi-stage research pipeline with workflow agents
    - [x] Create Analysis Pipeline using SequentialAgent (3 stages)
    - [x] Create Comparison Tool using ParallelAgent (2 branches)
    - [x] Wrap workflow agents as tools using AgentTool
    - [x] Create Research Coordinator LLM agent with all tools
    - [x] Implement custom merge strategy for parallel comparison
    - [x] Add comprehensive example with real research scenario
    - [x] Document the architecture and usage patterns

- [ ] Phase 5: Advanced Features (Week 5)
  - [ ] State persistence and serialization, present plan before implementation
  - [ ] Agent discovery and registry, present plan before implementation
  - [ ] Advanced merge strategies for parallel agents
  - [ ] Streaming support for long-running agents

- [ ] Phase 6: Migration and Testing (Week 5-6)
  - [ ] Remove old superfluos code, examples and tests
  - [ ] scan all product code for *backward compatibility* remove them and fix code to use new code.
  - [ ] no need for migration guide - update documentation to new codebase
  - [ ] scan all examples for Redo
    - [ ] Come up with an example plan based on codebase for product
    - [ ] remove invalid examples
    - [ ] Update Valid examples
    - [ ] Add missing examples
  - [ ] Comprehensive testing
    - [ ] Examine all integration tests to see which ones are needed, unneeded or need to be fixed, or new ones added and create a subtask list in todo.md
  - [ ] Performance benchmarking
    - [ ] move benchmarks directory under tests/
    - [ ] Examine all benchmark tests to see which ones are needed, unneeded or need to be fixed, or new ones added and create a subtask list in todo.md

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