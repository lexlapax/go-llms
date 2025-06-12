# Go-LLMs Project TODOs - v0.3.x Roadmap

## v0.3.1 Release ✅ COMPLETED (January 10, 2025)

## v0.3.2 Release (Ready - Documentation Update) ✅ COMPLETED (January 11, 2025)

## v0.3.2 Documentation simplification and refactoring ✅ COMPLETED (January 11, 2025)

## v0.3.3 Additional Providers in llm library
### 0.3.3.1: Ollama local hosted provider (https://ollama.com/) ✅ COMPLETED (January 11, 2025) 

### 0.3.3.2: OpenRouter provider (https://openrouter.ai) ✅ COMPLETED (January 11, 2025)

### 0.3.3.3: Google Vertex AI provider (https://cloud.google.com/vertex-ai) ✅ COMPLETED (January 11, 2025)

### 0.3.3.4: Mistral AI provider (https://mistral.ai/)
- [ ] Research Mistral AI API and update this todo.md list
  - [ ] Investigate API authentication method
  - [ ] Check API format (custom or OpenAI-compatible)
  - [ ] Document available models (Mistral 7B, Mixtral, etc.)
  - [ ] Identify streaming support and special features
  - [ ] Check for model listing/discovery endpoints
- [ ] Add dedicated provider implementation
  - [ ] Create `pkg/llm/provider/mistral.go`
  - [ ] Add Mistral-specific options if needed
  - [ ] Write unit tests in `pkg/llm/provider/mistral_test.go`
- [ ] Add model discovery/listing support (if available)
  - [ ] Implement fetcher in `pkg/util/llmutil/modelinfo/fetchers/mistral_fetcher.go`
  - [ ] Add tests for the fetcher
  - [ ] Integrate with modelinfo service
- [ ] Create dedicated example in `cmd/examples/provider-mistral/`
  - [ ] Show basic usage with the provider
  - [ ] Demonstrate function calling if supported
  - [ ] Add streaming examples
- [ ] Add integration tests
  - [ ] Create `tests/integration/mistral_integration_test.go`
  - [ ] Test all provider methods
  - [ ] Add error handling tests
- [ ] Update provider integration code
  - [ ] Update `pkg/util/llmutil/provider_parser.go` and tests
  - [ ] Update `pkg/util/llmutil/llmutil.go` and tests
  - [ ] Update `pkg/util/llmutil/env_vars.go` and tests
  - [ ] Update `pkg/util/llmutil/option_factories.go` and tests
  - [ ] Update `pkg/llm/provider/errors.go` for Mistral errors
  - [ ] Update `cmd/cli.go` and `cmd/config.go`
- [ ] Update documentation
  - [ ] Add Mistral section to `docs/user-guide/providers.md`
  - [ ] Document available models and features
  - [ ] Update main README.md

### 0.3.3.5: AWS Bedrock provider (https://aws.amazon.com/bedrock/)
- [ ] Research AWS Bedrock API and update this todo.md list
  - [ ] Investigate AWS authentication (IAM roles, access keys)
  - [ ] Study Bedrock's unified API for multiple models
  - [ ] Document available models (Claude, Llama 2, Jurassic, etc.)
  - [ ] Check streaming and function calling support
  - [ ] Identify region availability and restrictions
  - [ ] Determine model listing capabilities
- [ ] Add dedicated provider implementation
  - [ ] Create `pkg/llm/provider/bedrock.go`
  - [ ] Integrate AWS SDK for Go v2
  - [ ] Handle AWS authentication and region selection
  - [ ] Support multiple model families through unified API
  - [ ] Write unit tests in `pkg/llm/provider/bedrock_test.go`
- [ ] Add model discovery/listing support
  - [ ] Implement fetcher in `pkg/util/llmutil/modelinfo/fetchers/bedrock_fetcher.go`
  - [ ] Use AWS SDK to list available models
  - [ ] Handle region-specific model availability
  - [ ] Add tests for the fetcher
- [ ] Create dedicated example in `cmd/examples/provider-bedrock/`
  - [ ] Show AWS authentication setup
  - [ ] Demonstrate usage with different model families
  - [ ] Include region configuration
  - [ ] Show streaming if supported
- [ ] Add integration tests
  - [ ] Create `tests/integration/bedrock_integration_test.go`
  - [ ] Test with different AWS authentication methods
  - [ ] Test multiple model families
  - [ ] Test region-specific functionality
- [ ] Update provider integration code
  - [ ] Update `pkg/util/llmutil/provider_parser.go` and tests
  - [ ] Update `pkg/util/llmutil/llmutil.go` and tests
  - [ ] Update `pkg/util/llmutil/env_vars.go` and tests
  - [ ] Update `pkg/util/llmutil/option_factories.go` and tests
  - [ ] Update `pkg/llm/provider/errors.go` for AWS errors
  - [ ] Update `cmd/cli.go` and `cmd/config.go`
- [ ] Update documentation
  - [ ] Add Bedrock section to `docs/user-guide/providers.md`
  - [ ] Document AWS authentication setup
  - [ ] List supported models and regions
  - [ ] Include IAM permission requirements

### 0.3.3.6: Azure AI provider (https://azure.microsoft.com/en-us/products/ai-services/)
- [ ] Research Azure AI/OpenAI Service and update this todo.md list
  - [ ] Investigate Azure authentication (API keys, Azure AD)
  - [ ] Study Azure OpenAI Service API differences
  - [ ] Document deployment model vs standard model names
  - [ ] Check for Azure-specific features
  - [ ] Identify endpoint format and regions
  - [ ] Determine model/deployment listing capabilities
- [ ] Add dedicated provider implementation
  - [ ] Create `pkg/llm/provider/azure.go`
  - [ ] Handle Azure-specific endpoint format
  - [ ] Support deployment names vs model names
  - [ ] Add Azure AD authentication support
  - [ ] Write unit tests in `pkg/llm/provider/azure_test.go`
- [ ] Add deployment discovery/listing support (if available)
  - [ ] Implement fetcher in `pkg/util/llmutil/modelinfo/fetchers/azure_fetcher.go`
  - [ ] Handle Azure authentication for discovery
  - [ ] Map deployments to model capabilities
  - [ ] Add tests for the fetcher
- [ ] Create dedicated example in `cmd/examples/provider-azure/`
  - [ ] Show API key and Azure AD authentication
  - [ ] Demonstrate deployment configuration
  - [ ] Include endpoint customization
  - [ ] Show Azure-specific features
- [ ] Add integration tests
  - [ ] Create `tests/integration/azure_integration_test.go`
  - [ ] Test both authentication methods
  - [ ] Test deployment name handling
  - [ ] Test region-specific endpoints
- [ ] Update provider integration code
  - [ ] Update `pkg/util/llmutil/provider_parser.go` and tests
  - [ ] Update `pkg/util/llmutil/llmutil.go` and tests
  - [ ] Update `pkg/util/llmutil/env_vars.go` and tests
  - [ ] Update `pkg/util/llmutil/option_factories.go` and tests
  - [ ] Update `pkg/llm/provider/errors.go` for Azure errors
  - [ ] Update `cmd/cli.go` and `cmd/config.go`
- [ ] Update documentation
  - [ ] Add Azure section to `docs/user-guide/providers.md`
  - [ ] Document authentication options
  - [ ] Explain deployment vs model concepts
  - [ ] Include endpoint configuration

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