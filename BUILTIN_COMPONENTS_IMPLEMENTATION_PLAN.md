# Built-in Components Implementation Plan

## Phase 1: Core Registry Infrastructure (Week 1) ✅ COMPLETED

### 1.1 Base Registry Implementation ✅
- [x] Create `pkg/agent/builtins/registry.go` with generic registry interface
- [x] Implement thread-safe registry with caching
- [x] Add search and filtering capabilities
- [x] Create metadata structures with versioning support

### 1.2 Tool Registry ✅
- [x] Create `pkg/agent/builtins/tools/registry.go` 
- [x] Extend base registry with tool-specific features
- [x] Add resource usage tracking
- [x] Implement permission declarations

### 1.3 Agent Registry ✅
- [x] Create `pkg/agent/builtins/agents/registry.go`
- [x] Add template system for pre-configured agents
- [x] Implement agent composition helpers

### 1.4 Workflow Registry ✅
- [x] Create `pkg/agent/builtins/workflows/registry.go`
- [x] Add workflow builder utilities
- [x] Implement routing patterns

## Phase 2: Initial Tool Set (Week 2)

### 2.1 Web Tools
- [x] Migrate existing WebFetch to built-ins with metadata
- [ ] Add WebScrape tool for HTML extraction
- [ ] Add WebSearch tool with configurable search engines
- [ ] Add HTTPRequest tool for advanced HTTP operations

### 2.2 File Tools
- [ ] FileRead - read file contents with encoding support
- [ ] FileWrite - write with atomic operations
- [ ] FileList - directory listing with filters
- [ ] FileSearch - grep-like file content search

### 2.3 Data Tools
- [ ] JSONProcess - parse, query (JSONPath), and transform JSON
- [ ] CSVProcess - read, write, and transform CSV data
- [ ] DataTransform - common transformations (filter, map, reduce)

### 2.4 Text Tools
- [ ] TextSummarize - intelligent summarization using LLM
- [ ] TextExtract - extract structured data from text
- [ ] TextAnalyze - sentiment, entities, keywords

## Phase 3: Agent Templates (Week 3)

### 3.1 Research Agents
- [ ] WebResearcher - web research with source tracking
- [ ] DocumentAnalyzer - analyze documents and PDFs
- [ ] FactChecker - verify claims against sources

### 3.2 Coding Agents
- [ ] CodeReviewer - review code for issues
- [ ] TestGenerator - generate tests from code
- [ ] DocWriter - generate documentation

### 3.3 Data Agents
- [ ] DataAnalyst - analyze datasets and generate insights
- [ ] ReportGenerator - create formatted reports
- [ ] DataCleaner - clean and validate data

## Phase 4: Workflow Patterns (Week 4)

### 4.1 Core Patterns
- [ ] Pipeline - sequential processing
- [ ] MapReduce - parallel processing with aggregation
- [ ] Consensus - multi-agent agreement
- [ ] Retry - with exponential backoff

### 4.2 Example Workflows
- [ ] ResearchWorkflow - research → verify → summarize → report
- [ ] CodeReviewWorkflow - analyze → review → suggest → document
- [ ] DataPipeline - ingest → clean → analyze → visualize

## Phase 5: Documentation and Examples (Week 5)

### 5.1 Documentation
- [ ] Built-in components guide in docs/user-guide/
- [ ] API reference for each component
- [ ] Best practices for extending built-ins
- [ ] Migration guide from custom to built-ins

### 5.2 Examples
- [ ] Example for each built-in tool
- [ ] Example for each agent template
- [ ] Example for each workflow pattern
- [ ] End-to-end application examples

### 5.3 Testing
- [ ] Unit tests for all components
- [ ] Integration tests for workflows
- [ ] Performance benchmarks
- [ ] Example validation tests

## Implementation Priority Order

### Week 1 Focus: Foundation
1. Generic registry implementation
2. Tool registry with metadata
3. Basic discovery mechanisms
4. Initial tests and benchmarks

### Week 2 Focus: Core Tools
1. Web tools (high usage expected)
2. File tools (fundamental operations)
3. JSON tool (most common data format)
4. Text summarization (showcases LLM integration)

### Week 3 Focus: Agents
1. WebResearcher (most requested)
2. CodeReviewer (developer audience)
3. DataAnalyst (broad applicability)

### Week 4 Focus: Workflows
1. Pipeline pattern (simplest, most common)
2. Research workflow (combines multiple agents)
3. Consensus pattern (unique multi-provider feature)

### Week 5 Focus: Polish
1. Comprehensive documentation
2. Example applications
3. Performance optimization
4. Community feedback incorporation

## Success Metrics

1. **Adoption**: Number of imports of built-in packages
2. **Performance**: No regression in benchmarks
3. **Usability**: Time to implement common tasks reduced by 50%
4. **Extensibility**: Community contributions of new built-ins
5. **Documentation**: All components have examples and guides

## Risk Mitigation

1. **Breaking Changes**: All built-ins behind separate imports
2. **Performance**: Extensive benchmarking before release
3. **Complexity**: Start simple, iterate based on feedback
4. **Maintenance**: Clear ownership and contribution guidelines
5. **Security**: Permission model and sandboxing from day one

## Code Organization Best Practices

1. **One tool per file**: Each tool in its own file for clarity
2. **Consistent naming**: ToolName matches filename and function
3. **Comprehensive tests**: Each component has dedicated test file
4. **Examples in tests**: Test files include usage examples
5. **Benchmark everything**: Performance tests for all components

## Next Immediate Steps

1. Create the directory structure
2. Implement generic registry
3. Create first built-in tool (WebFetch) as reference
4. Write tests for registry and first tool
5. Create example showing discovery and usage