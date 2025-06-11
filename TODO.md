# Go-LLMs Project TODOs

## Features
- [ ] Add Model Context Protocol Client support for Agents
- [ ] Add Model Context Protocol Server support for Workflows or Agents

## Authentication System Improvements
- [ ] Implement provider registry pattern for auth detection (MEDIUM PRIORITY)
  - [ ] Create AuthProvider interface with Name(), CanHandle(), Configure() methods
  - [ ] Implement AuthRegistry for managing multiple providers
  - [ ] Add configuration file support for custom auth mappings (YAML/JSON)
  - [ ] Support provider patterns:
    - [ ] URL pattern matching (regex support)
    - [ ] Response-based detection (401 + WWW-Authenticate header)
    - [ ] OpenAPI security scheme integration
    - [ ] OAuth2 discovery via .well-known endpoints
  - [ ] Create default providers for common services:
    - [ ] GitHub (including Enterprise)
    - [ ] GitLab (including self-hosted)
    - [ ] Generic Bearer token
    - [ ] Generic API key
    - [ ] Basic auth
  - [ ] Ensure backward compatibility with existing detectURLSpecificAuth
  - [ ] Add examples and documentation
  - [ ] Benefits:
    - [ ] Works with any URL including test servers
    - [ ] Extensible without code changes
    - [ ] User-configurable auth providers
    - [ ] Standards-based approach

## Testing & Performance
- [ ] Performance profiling and optimization:
  - [ ] Phase 1: Baseline Profiling Infrastructure (Prerequisites)
    - [ ] P1: Create benchmark harness for A/B testing optimizations (REVISIT)
    - [ ] P2: Implement visualization for memory allocation patterns (REVISIT)
    - [ ] P2: Create real-world test scenarios for end-to-end performance (REVISIT)

  - Phase 2: High-Impact Optimizations (Quick Wins)
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

### Agent Architecture Restructuring
- Agent Architecture Restructuring - All phases completed - see TODO-DONE.md

- [ ] Phase 6: Advanced Features (MOVED TO PHASE 7) (low priority)
  - [ ] State persistence and serialization, present plan before implementation
  - [ ] Agent discovery and registry, present plan before implementation
  - [ ] Advanced merge strategies for parallel agents
  - [ ] Streaming support for long-running agents

### Tool System Enhancement with LLM Guidance (HIGHEST PRIORITY - IN PROGRESS)
- Phase 1: Core Infrastructure (Week 1)
  - Day 1-4: COMPLETED - see TODO-DONE.md
  - Day 5: API Client Tool & Integration testing
    - Phase 1-4: COMPLETED - see TODO-DONE.md
    - [ ] Phase 5: Advanced Capabilities for web-api-client tool (medium priority)
      - [ ] Auto-Pagination: Automatically follow pagination links
      - [ ] Rate Limiting: Respect rate limit headers with intelligent backoff
      - [ ] Response Caching: Cache responses with configurable TTL
      - [ ] Request Templates: Store and reuse common request patterns
      - [ ] Response Transformation: Extract data using JSONPath or JQ-like queries
      - [ ] Error Recovery: Smart retries with exponential backoff
      - [ ] Mock Mode: Optional mock responses for testing
      - [ ] Streaming Responses: Handle large response streaming
      - [ ] Request/Response Middleware: Plugin system for custom processing
      - [ ] Multi-tenancy: Support multiple API configurations
      - [ ] Request Batching: Batch multiple requests for efficiency
    - Integration Testing - COMPLETED - see TODO-DONE.md

- Phase 2: Tool Migration to Enhanced Interface (Week 2) - COMPLETED - see TODO-DONE.md

- [ ] Phase 3: Tool Migration Part 2 (Week 3) - Continue migration to ToolBuilder pattern
  - [x] Day 1: Migrate data tools (4 tools) - COMPLETED January 10, 2025
    - [x] json_process - Update to ToolBuilder with JSONPath query examples
    - [x] csv_process - Add transformation examples, delimiter options
    - [x] xml_process - Add XPath guidance, namespace handling
    - [x] data_transform - Add operation chains, performance considerations
    - [ ] update all example that use these tools to a pattern similar  to the calculator tool or the web api tool
  - [x] Day 2: Migrate datetime tools (7 tools) - COMPLETED January 10, 2025
    - [x] datetime_now - Add timezone examples, format options
    - [x] datetime_info - Add component extraction examples, week calculations
    - [x] datetime_calculate - Add business days examples, date math
    - [ ] datetime_parse - Add format pattern examples, auto-detection (IN PROGRESS)
    - [ ] datetime_format - Add locale examples, custom formats
    - [ ] datetime_convert - Add timezone conversion examples
    - [ ] datetime_compare - Add comparison logic, relative time examples
    - [ ] update all example that use these tools to a pattern similar  to the calculator tool or the web api tool
  - [ ] Day 3: Migrate feed tools (6 tools)
    - [ ] feed_fetch - Add format detection examples, encoding handling
    - [ ] feed_discover - Add auto-discovery examples, link parsing
    - [ ] feed_filter - Add complex query examples, date filtering
    - [ ] feed_aggregate - Add deduplication examples, merge strategies
    - [ ] feed_convert - Add format conversion examples
    - [ ] feed_extract - Add content extraction patterns
    - [ ] update all example that use these tools to a pattern similar  to the calculator tool or the web api tool
  - [ ] Day 4: Update examples (first 15)
    - [ ] agent-calculator - Remove manual prompt
    - [ ] agent-simple-llm - Use auto docs
    - [ ] agent-llm-builtin-tools - Showcase
    - [ ] agent-tools-conversion - Update
    - [ ] builtins-* examples (7) - Update all
    - [ ] Other agent examples (4)
    - [ ] update all example that use these tools to a pattern similar  to the calculator tool or the web api tool
  - [ ] Day 5: Update examples (remaining 16)
    - [ ] Update all remaining examples
    - [ ] Verify all examples work
    - [ ] Update example documentation

- [ ] Phase 4: Documentation & Polish (Week 4)
  - Day 1-2: Technical documentation - COMPLETED January 9, 2025
    - Technical documentation for tools.md created
    - [ ] Add architecture diagrams (SVG diagrams still needed)
  - [ ] Day 3-4: User guide updates
    - [ ] Create docs/user-guide/tool-development.md
    - [ ] Update docs/user-guide/builtin-tools.md
    - [ ] Add migration guide
    - [ ] Create examples gallery
  - [ ] Day 5: Final testing & release
    - [ ] Run full test suite
    - [ ] Performance validation
    - [ ] Create release notes
    - [ ] Tag release
    - [ ] ensure all the documentation links are updated and correct, root README.md REFERENCE.md , the index READMES in docs/ and examples

### Previous Built-in Components Plan
- P2: Build useful built-in tools - COMPLETED - see TODO-DONE.md

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