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
- [x] Add new file tools: ✅ COMPLETED
  - [x] FileList - directory listing with filters
    - [x] Pattern matching and size/date filters
    - [x] Sorting capabilities (name, size, modified)
    - [x] Recursive directory traversal
    - [x] File extension extraction
  - [x] FileDelete - safe file deletion with confirmation
    - [x] Safety checks for critical system directories
    - [x] Confirmation requirements for destructive operations
    - [x] Support for recursive directory deletion
    - [x] Force flag for advanced users
  - [x] FileMove - move/rename files
    - [x] Atomic moves within same filesystem
    - [x] Cross-device move support (copy then delete)
    - [x] Directory move support
    - [x] Overwrite and directory creation options
  - [x] FileSearch - grep-like file content search
    - [x] Plain text and regex pattern matching
    - [x] Case-sensitive/insensitive search
    - [x] Context lines before/after matches
    - [x] File pattern filtering
    - [x] Binary file detection and skipping
    - [x] Recursive directory search

### 2.3 System Tools ✅ COMPLETED
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
- [x] Add new system tools:
  - [x] GetEnvironmentVariable - read env vars safely ✅ COMPLETED
    - [x] Pattern matching for variable names (prefix*, *suffix, *contains*)
    - [x] Sensitive variable masking (API keys, passwords, tokens)
    - [x] Configurable value inclusion/exclusion with NoValues flag
    - [x] Sorted output for better readability
    - [x] Comprehensive tests covering all features
  - [x] GetSystemInfo - OS, architecture, resources ✅ COMPLETED
    - [x] Basic system information (OS, architecture, CPU count)
    - [x] Memory statistics with runtime allocation info
    - [x] Go runtime information (version, goroutines, GOMAXPROCS)
    - [x] Environment summary (paths, temp dir, env var count)
    - [x] Cross-platform support with platform name mapping
    - [x] Comprehensive tests with full coverage
  - [x] ProcessList - list running processes ✅ COMPLETED
    - [x] Cross-platform process enumeration (Unix/Linux/macOS/Windows)
    - [x] Process filtering by name (case-insensitive contains)
    - [x] Sorting by PID, name, CPU usage, or memory usage
    - [x] Include/exclude current process option
    - [x] Result limiting for performance
    - [x] Process information extraction (PID, name, command, CPU%, memory, user)
    - [x] Comprehensive tests with helper function coverage

### 2.4 Data Tools ✅ COMPLETED
- [x] JSONProcess - parse, query (JSONPath), and transform JSON
  - [x] JSON parsing and validation with error handling
  - [x] JSONPath querying with object navigation and array indexing
  - [x] Transform operations: extract_keys, extract_values, flatten, prettify, minify
  - [x] Type-safe execution with proper result types
  - [x] Comprehensive test coverage
- [x] CSVProcess - read, write, and transform CSV data
  - [x] CSV parsing with configurable delimiter and header support
  - [x] Filtering with multiple operators (eq, ne, contains, starts_with, ends_with, gt, lt, gte, lte)
  - [x] Transform operations: select_columns, sort, count_rows, statistics
  - [x] CSV to JSON conversion with proper type handling
  - [x] Comprehensive test coverage
- [x] XMLProcess - parse and query XML data
  - [x] XML parsing with full attribute support
  - [x] Simplified XPath querying for elements and attributes
  - [x] XML to JSON conversion with configurable attribute inclusion
  - [x] Nested element navigation and array handling
  - [x] Comprehensive test coverage
- [x] DataTransform - common transformations (filter, map, reduce)
  - [x] Filter: complex condition-based filtering with field access
  - [x] Map: extract_field, to_upper, to_lower, to_number, to_string
  - [x] Reduce: sum, count, min, max, average, concat
  - [x] Additional operations: sort, group_by, unique, reverse
  - [x] Nested field access support with dot notation
  - [x] Comprehensive test coverage
- [x] All tools follow consistent built-in tool patterns
- [x] Proper registration with the tools registry
- [x] No LLM dependencies - pure data processing

### 2.5 Date Time Tools
- [ ] research common date, time, timezone, conversion actions and update todo .. read and update this todo list
- [ ] todays date, 
- [ ] date time conversion +- days
- [ ] duration calculation
- [ ] weekday calculations
- [ ] timezone convert
- [ ] parse date time strings
- [ ] format date time strings
- [ ] compare date times
- [ ] business days
- [ ] unix timestamp conversion

### 2.6 Feed Process tools (rss, atom and other feeds)
- [ ] research common feed actions and needed tools  and create todo

## Phase 3: Agent Templates (Week 3)

### 3.1 Text Agents 
- [ ] TextSummarize - intelligent summarization using LLM
- [ ] TextExtract - extract structured data from text
- [ ] TextAnalyze - sentiment, entities, keywords
- [ ] TextTranslate - language translation using LLM

### 3.2 Research Agents
- [ ] WebResearcher - web research with source tracking
- [ ] DocumentAnalyzer - analyze documents and PDFs
- [ ] FactChecker - verify claims against sources

### 3.3 Coding Agents
- [ ] CodeReviewer - review code for issues
- [ ] TestGenerator - generate tests from code
- [ ] DocWriter - generate documentation

### 3.4 Data Agents
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