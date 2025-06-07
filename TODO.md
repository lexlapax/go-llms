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

### Agent Architecture Restructuring
- [x] Phase 1 & 1.5: Core Infrastructure - COMPLETED - see TODO-DONE.md
- [x] Phase 2: LLM Agent Migration - COMPLETED - see TODO-DONE.md
- [x] Phase 3: Workflow Agents - COMPLETED - see TODO-DONE.md
- [x] Phase 4: Agent-Tool Integration - COMPLETED - see TODO-DONE.md
- [x] Phase 5: Multi-Agent System Enhancement - COMPLETED - see TODO-DONE.md

- [ ] Phase 6: Advanced Features (MOVED TO PHASE 7) (low priority)
  - [ ] State persistence and serialization, present plan before implementation
  - [ ] Agent discovery and registry, present plan before implementation
  - [ ] Advanced merge strategies for parallel agents
  - [ ] Streaming support for long-running agents

- [x] Phase 7: Migration and Testing - COMPLETED - see TODO-DONE.md

### Tool System Enhancement with LLM Guidance (HIGHEST PRIORITY - IN PROGRESS)
- [ ] Phase 1: Core Infrastructure (Week 1)
  - [x] Day 1-4: COMPLETED - see TODO-DONE.md
  - [ ] Day 5: API Client Tool & Integration testing
    - [x] Phase 1: Basic REST Client Implementation
      - [x] Create API_CLIENT_TOOL_PLAN.md with detailed design
      - [x] Write tests for basic REST operations (GET, POST, PUT, DELETE)
      - [x] Implement core API client with JSON handling
      - [x] Add basic auth support (API key, Bearer token)
      - [x] Create examples for common API patterns
    - [ ] Phase 2: OpenAPI/Swagger Integration  
      - [ ] Add OpenAPI spec parsing capability
      - [ ] Implement operation discovery from specs
      - [ ] Add request validation against schemas
      - [ ] Create examples using public OpenAPI specs
      - [ ] Add automatic endpoint discovery from specs
    - [ ] Phase 3: GraphQL Support
      - [ ] Implement GraphQL query execution
      - [ ] Add GraphQL mutation support
      - [ ] Support GraphQL variables and fragments
      - [ ] Add GraphQL introspection capabilities
    - [ ] Phase 4: Advanced Authentication
      - [ ] Add OAuth2 flows (client credentials, authorization code)
      - [ ] Implement session/cookie management
      - [ ] Add custom authentication header support
      - [ ] Support JWT token refresh
    - [ ] Phase 5: Advanced Capabilities
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
    - [x] Integration Testing
      - [x] Test tool with LLM Agent
      - [x] Verify MCP export functionality
      - [x] Run performance benchmarks
      - [x] Fix any integration issues
      - [x] Write an example builtins-web-api-client with demonstrates various aspects of the tool
      - [ ] Update tools documentation in docs/ with about this tool.



- [ ] Phase 2: Tool Migration Part 1 (Week 2) - 15 tools
  - [ ] Day 1: Migrate calculator tool (1 tool - template for others)
    - [ ] Write tests for calculator with new interface
    - [ ] Implement comprehensive metadata
    - [ ] Add 5+ examples with scenarios
    - [ ] Test MCP export
  - [ ] Day 2: Migrate system tools (4 tools)
    - [ ] execute_command - Add safety constraints and confirmation
    - [ ] get_environment_variable - Simple migration
    - [ ] get_system_info - Add output examples
    - [ ] process_list - Add filtering guidance
  - [ ] Day 3: Migrate file tools (6 tools)
    - [ ] file_read - Add encoding and size guidance
    - [ ] file_write - Add destructive warnings
    - [ ] file_list - Complex parameter examples
    - [ ] file_delete - Add confirmation requirements
    - [ ] file_move - Add rollback guidance
    - [ ] file_search - Add regex examples
  - [ ] Day 4: Migrate web tools (4 tools)
    - [ ] web_search - Multi-engine examples
    - [ ] web_fetch - Add timeout guidance
    - [ ] web_scrape - Selector examples
    - [ ] http_request - Auth method examples
  - [ ] Day 5: Testing & fixes
    - [ ] Run all migrated tool tests
    - [ ] Fix any issues
    - [ ] Update integration tests

- [ ] Phase 3: Tool Migration Part 2 (Week 3) - 17 tools + examples
  - [ ] Day 1: Migrate data tools (4 tools)
    - [ ] json_process - JSONPath examples
    - [ ] csv_process - Transform examples
    - [ ] xml_process - XPath guidance
    - [ ] data_transform - Operation chains
  - [ ] Day 2: Migrate datetime tools (7 tools)
    - [ ] datetime_now - Timezone examples
    - [ ] datetime_info - Component extraction
    - [ ] datetime_calculate - Business days
    - [ ] datetime_parse - Format patterns
    - [ ] datetime_format - Locale examples
    - [ ] datetime_convert - Zone handling
    - [ ] datetime_compare - Comparison logic
  - [ ] Day 3: Migrate feed tools (6 tools)
    - [ ] feed_fetch - Format detection
    - [ ] feed_discover - Auto-discovery
    - [ ] feed_filter - Complex queries
    - [ ] feed_aggregate - Deduplication
    - [ ] feed_convert - Format examples
    - [ ] feed_extract - Content parsing
  - [ ] Day 4: Update examples (first 15)
    - [ ] agent-calculator - Remove manual prompt
    - [ ] agent-simple-llm - Use auto docs
    - [ ] agent-llm-builtin-tools - Showcase
    - [ ] agent-tools-conversion - Update
    - [ ] builtins-* examples (7) - Update all
    - [ ] Other agent examples (4)
  - [ ] Day 5: Update examples (remaining 16)
    - [ ] Update all remaining examples
    - [ ] Verify all examples work
    - [ ] Update example documentation

- [ ] Phase 4: Documentation & Polish (Week 4)
  - [ ] Day 1-2: Technical documentation
    - [ ] Create docs/technical/tools.md
    - [ ] Document new Tool interface
    - [ ] Add architecture diagrams
    - [ ] Include best practices
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