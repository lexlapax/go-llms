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

## Phase 2: Initial Tool Set - Migration and Enhancement (Week 2)

### 2.0 Migration Analysis and Strategy ✅ COMPLETED
- [x] Analyze existing tools in common_tools.go
- [x] Create migration plan with enhancements
- [x] Update examples to use built-in tools instead of common_tools.go
  - [x] Reorganized examples with clear naming (builtins-discovery, builtins-file-tools)
  - [x] Added migration guide showing differences
  - [x] Created benchmarks for built-in tools performance
  - [x] Documented tool discovery features
- [x] Deprecate common_tools.go after migration complete
  - [x] Updated all benchmarks to use built-in tools
  - [x] Removed tests that depended on common_tools.go
  - [x] Successfully removed common_tools.go from codebase

### 2.1 Web Tools
- [x] Migrate existing WebFetch to built-ins with enhancements:
  - [x] Custom timeout support
  - [x] Header capture
  - [x] Resource usage metadata
- [x] Implement WebSearch tool (schema exists, needs implementation) ✅ COMPLETED
  - [x] DuckDuckGo search engine support
  - [x] Configurable result limits
  - [x] Safe search filtering
  - [x] Timeout configuration
  - [x] Comprehensive tests
- [x] Add WebScrape tool for HTML extraction ✅ COMPLETED
  - [x] HTML parsing without external dependencies
  - [x] Text extraction with script/style removal
  - [x] Link discovery and classification (internal/external/anchor)
  - [x] Metadata extraction (meta tags, title, og tags)
  - [x] Simplified CSS-like selector support (tag, class, id)
  - [x] Configurable extraction options
  - [x] Comprehensive tests
- [x] Add HTTPRequest tool for advanced HTTP operations ✅ COMPLETED
  - [x] Full HTTP method support (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
  - [x] Multiple authentication methods (basic, bearer, API key)
  - [x] Custom headers and query parameters
  - [x] Various body content types (JSON, form, XML, text)
  - [x] Redirect control with follow/no-follow options
  - [x] Timeout configuration
  - [x] Comprehensive response information with timing
  - [x] Flattened auth fields for tool framework compatibility
  - [x] Comprehensive tests covering all features

### 2.2 File Tools (Priority: Migrate existing tools) ✅ COMPLETED
- [x] Migrate and enhance FileRead from common_tools.go:
  - [x] Large file handling (streaming with 4KB buffer)
  - [x] Binary file detection
  - [x] Encoding detection (UTF-8/binary)
  - [x] File metadata (size, permissions, modified time)
  - [x] Line range reading (start/end line numbers)
  - [x] Size limits with truncation warnings
- [x] Migrate and enhance FileWrite from common_tools.go:
  - [x] Append mode support
  - [x] Custom permissions (file mode)
  - [x] Directory creation option
  - [x] Atomic write support (write to temp, then rename)
  - [x] Backup creation with timestamps
- [ ] Add new file tools:
  - [ ] FileList - directory listing with filters
  - [ ] FileDelete - safe file deletion with confirmation
  - [ ] FileMove - move/rename files
  - [ ] FileSearch - grep-like file content search

### 2.3 System Tools (Priority: Migrate ExecuteCommand) ✅ COMPLETED
- [x] Migrate and enhance ExecuteCommand from common_tools.go:
  - [x] Environment variable support
  - [x] Working directory configuration
  - [x] Stdin support
  - [x] Separate stdout/stderr capture
  - [x] Command sanitization options (safe mode with allowlist/blocklist)
  - [x] Timeout control (max 5 minutes)
  - [x] Shell selection (sh, bash, zsh, or direct execution)
  - [x] Comprehensive safety checks and dangerous command blocking
  - [x] Exit code and success tracking
  - [x] Duration metrics
- [ ] Add new system tools:
  - [ ] GetEnvironmentVariable - read env vars safely
  - [ ] GetSystemInfo - OS, architecture, resources
  - [ ] ProcessList - list running processes

### 2.4 Data Tools
- [ ] JSONProcess - parse, query (JSONPath), and transform JSON
- [ ] CSVProcess - read, write, and transform CSV data
- [ ] XMLProcess - parse and query XML data
- [ ] DataTransform - common transformations (filter, map, reduce)

### 2.5 Text Tools
- [ ] TextSummarize - intelligent summarization using LLM
- [ ] TextExtract - extract structured data from text
- [ ] TextAnalyze - sentiment, entities, keywords
- [ ] TextTranslate - language translation using LLM

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

## Migration Guide from common_tools.go

### For Library Users
When migrating from common_tools.go to built-in tools:

1. **Import Changes**:
   ```go
   // Old
   import "github.com/lexlapax/go-llms/pkg/agent/tools"
   tool := tools.WebFetch()
   
   // New
   import (
       "github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
       _ "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
   )
   tool := tools.MustGetTool("web_fetch")
   ```

2. **Discovery Benefits**:
   ```go
   // List all available tools
   allTools := tools.Tools.List()
   
   // Find tools by category
   webTools := tools.Tools.ListByCategory("web")
   
   // Search for tools
   fileTools := tools.Tools.Search("file")
   ```

3. **Enhanced Metadata**:
   - Version tracking
   - Resource usage information
   - Permission requirements
   - Examples and documentation

### For Tool Implementers
When migrating a tool to the built-in structure:

1. Create appropriate category directory
2. Follow the pattern in `web/fetch.go`
3. Add comprehensive metadata
4. Include init() function for auto-registration
5. Add tests and benchmarks
6. Document enhancements made

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