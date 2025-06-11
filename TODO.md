# Go-LLMs Project TODOs - v0.3.x Roadmap

## v0.3.1 Release ✅ COMPLETED (January 10, 2025)

## v0.3.2 Release (Ready - Documentation Update) COMPLETED (January 11, 2025)

## v0.3.2 Documentation simplification and refactoring ✅ COMPLETED (January 11, 2025)

## v0.3.3 Additional Providers in llm library
### 0.3.3.1: Ollama local hosted provider (https://ollama.com/)
- [x] Research how to add Ollama provider and update this todo.md list
  - **Research Findings**: Ollama already has FULL support through the OpenAI-compatible provider
  - **Implementation**: Use `provider.NewOpenAIProvider()` with custom base URL and dummy API key
  - **Tests**: Already exist in `tests/integration/ollama_integration_test.go`
  - **Example**: Already exists in `cmd/examples/provider-openai-compatible/main.go`
- [ ] Add dedicated `NewOllamaProvider()` convenience function in `pkg/llm/provider/ollama.go`
  - [ ] Create wrapper that uses OpenAI provider with proper defaults (base URL, dummy key, timeout)
  - [ ] Add Ollama-specific options (e.g., WithOllamaHost(), WithOllamaTimeout())
  - [ ] Document that it's a convenience wrapper around OpenAI provider
- [ ] Add model discovery/listing support for Ollama
  - [ ] Implement Ollama's `/api/tags` endpoint to list available models
  - [ ] Add to modelinfo fetchers as `ollama_fetcher.go`
- [ ] Create dedicated Ollama example in `cmd/examples/provider-ollama/`
  - [ ] Show basic usage with the new convenience provider
  - [ ] Demonstrate model listing
  - [ ] Show streaming and multimodal capabilities
- [ ] Enhance existing integration tests
  - [ ] Add tests for the new convenience provider
  - [ ] Test model listing functionality
  - [ ] Add multimodal tests (if Ollama models support it)
- [ ] Update documentation
  - [ ] Add Ollama section to `docs/user-guide/providers.md`
  - [ ] Update `docs/technical/provider-implementation.md` with Ollama details
  - [ ] Document Ollama-specific features and limitations

### 0.3.3.2: openrouter provider (https://openrouter.ai)
### 0.3.3.3: mistral provider (https://mistral.ai/)
### 0.3.3.4: AWS Bedrock Cloud provider (https://aws.amazon.com/bedrock/)
### 0.3.3.5: Azure AI Cloud provider (https://azure.microsoft.com/en-us/products/ai-services/)
### 0.3.3.6: Google Vertex Cloud provider (https://cloud.google.com/vertex-ai)

## v0.3.5: Enhanced Tool Capabilities
### 0.3.5.1: Web API Client Advanced Features 
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

### 0.3.5.2: Authentication System Improvements 
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

## v0.3.6: Advanced Agent Features
### 0.3.6.1: State Management 
- [ ] State persistence and serialization design
- [ ] Implement state storage backends (file, database)
- [ ] Add state versioning and migration
- [ ] Create examples with persistent agents

### 0.3.6.2: Agent Discovery
- [ ] Agent discovery and registry design
- [ ] Implement agent metadata and search
- [ ] Add dynamic agent loading
- [ ] Create agent marketplace example

### 0.3.6.3: Advanced Features 
- [ ] Advanced merge strategies for parallel agents
- [ ] Streaming support for long-running agents
- [ ] Agent composition patterns
- [ ] Agent lifecycle management

## v0.3.7: Model Context Protocol (MCP) Support
### 0.3.7.1: MCP Client Support 
- [ ] Research MCP specification and requirements
- [ ] Design MCP client interface for agents
- [ ] Implement MCP client in pkg/agent/mcp/client
- [ ] Add MCP tool discovery and registration
- [ ] Create examples demonstrating MCP client usage
- [ ] Write comprehensive tests
- [ ] Document MCP client usage in user guide

### 0.3.7.2: MCP Server Support 
- [ ] Design MCP server interface for exposing agents/workflows
- [ ] Implement MCP server in pkg/agent/mcp/server
- [ ] Add agent/workflow registration to MCP server
- [ ] Create example MCP server implementations
- [ ] Write comprehensive tests
- [ ] Document MCP server setup and configuration

## v0.3.8: Built-in Agents Library
### 0.3.8.1: Text Processing Agents 
- [ ] TextSummarize - intelligent summarization using LLM
- [ ] TextExtract - extract structured data from text
- [ ] TextAnalyze - sentiment, entities, keywords
- [ ] TextTranslate - language translation using LLM

### 0.3.8.2: Research Agents 
- [ ] WebResearcher - web research with source tracking
- [ ] DocumentAnalyzer - analyze documents and PDFs
- [ ] FactChecker - verify claims against sources

### 0.3.8.3: Coding Agents 
- [ ] CodeReviewer - review code for issues
- [ ] TestGenerator - generate tests from code
- [ ] DocWriter - generate documentation

### 0.3.8.4: Data Agents 
- [ ] DataAnalyst - analyze datasets and generate insights
- [ ] ReportGenerator - create formatted reports
- [ ] DataCleaner - clean and validate data

### 0.3.8.5: Feed Agents 
- [ ] NewsMonitor - monitor news feeds for keywords and topics using LLM
- [ ] FeedAggregator - aggregate and deduplicate content from multiple feeds
- [ ] FeedSummarizer - summarize feed content using LLM
- [ ] ContentCurator - curate and categorize feed content using LLM

## v0.3.9: Built-in Workflow Patterns
### 0.3.9.1: Core Workflow Patterns 
- [ ] Pipeline - sequential processing workflow
- [ ] MapReduce - parallel processing with aggregation
- [ ] Consensus - multi-agent agreement pattern
- [ ] Retry - with exponential backoff

### 0.3.9.2: Example Workflows 
- [ ] ResearchWorkflow - research → verify → summarize → report
- [ ] CodeReviewWorkflow - analyze → review → suggest → document
- [ ] DataPipeline - ingest → clean → analyze → visualize


## v0.3.10: Performance Optimization
### 0.3.10.1: Profiling Infrastructure (REVISIT)
- [ ] Create benchmark harness for A/B testing optimizations
- [ ] Implement visualization for memory allocation patterns
- [ ] Create real-world test scenarios for end-to-end performance

### 0.3.10.2: Advanced Optimizations (REVISIT)
- [ ] Implement adaptive channel buffer sizing based on usage patterns
- [ ] Add pool prewarming for high-throughput scenarios
- [ ] Reduce redundant property iterations in schema processing
- [ ] Implement more granular locking in cached objects
- [ ] Optimize zero-initialization patterns for pooled objects
- [ ] Introduce buffer pooling for string builders

### 0.3.10.3: Performance Validation (REVISIT)
- [ ] Document performance improvements with metrics
- [ ] Verify optimizations in high-concurrency scenarios
- [ ] Create benchmark comparison charts for before/after
- [ ] Implement regression testing to prevent performance degradation
- [ ] Add performance acceptance criteria to CI pipeline

## v0.3.11: Final Polish and Stable Release
### 0.3.11.1: Documentation Polish
- [ ] Fix identified cross-link issues (path inconsistencies, broken links)
- [ ] Perform final consistency check across all documentation
- [ ] Update all examples to showcase v0.3.x features
- [ ] Create migration guide from earlier versions

### 0.3.11.2: API Refinement
- [ ] API refinement based on usage feedback
- [ ] Deprecate old patterns with migration paths
- [ ] Ensure backward compatibility where possible
- [ ] Final review and preparation for v0.4.0 stable release

## Notes
- Tool System Enhancement Phases 1-4: COMPLETED (see TODO-DONE.md)
- Agent Architecture Restructuring Phases 1-7: COMPLETED (see TODO-DONE.md)
- Performance Phase 2 (Quick Wins): COMPLETED (see TODO-DONE.md)

See TODO-DONE.md for all completed tasks