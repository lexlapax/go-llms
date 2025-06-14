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

## v0.3.4: Enhanced Tool Capabilities
### 0.3.4.1 Advanced Tool features - Runtime Tool Discovery for Scripting Engines ✅ COMPLETED (June 13, 2025)
  
### 0.3.4.5: Web API Client Advanced Features 
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

### 0.3.4.6: Authentication System Improvements 
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

## v0.3.5: Scripting Engine Integration Support (go-llmspell requirements)

### 0.3.5.1: Schema Package Implementations
- [ ] Implement InMemorySchemaRepository
  - [ ] Thread-safe in-memory storage for schemas
  - [ ] CRUD operations for schema management
  - [ ] Schema versioning support
  - [ ] Export/import functionality
- [ ] Implement FileSchemaRepository
  - [ ] File-based persistent schema storage
  - [ ] Directory structure for schema organization
  - [ ] Schema file format (JSON/YAML)
  - [ ] Migration support between versions
- [ ] Implement ReflectionSchemaGenerator
  - [ ] Generate schemas from Go structs using reflection
  - [ ] Handle nested structs and slices
  - [ ] Support for custom types
  - [ ] Preserve Go type information in schema
- [ ] Implement TagSchemaGenerator
  - [ ] Generate schemas from struct tags (json, validate, etc.)
  - [ ] Support multiple tag formats
  - [ ] Custom tag handlers
  - [ ] Validation rule extraction
- [ ] Add comprehensive tests for all implementations
- [ ] Create examples demonstrating schema usage

### 0.3.5.2: Enhanced Tool Discovery System
- [ ] Dynamic Tool Registration
  - [ ] Add RegisterTool method to ToolDiscovery interface
  - [ ] Add UnregisterTool method for runtime removal
  - [ ] Add GetRegisteredTools for listing all tools
  - [ ] Thread-safe registration/unregistration
  - [ ] Tool versioning support
- [ ] Tool Metadata Repository
  - [ ] Persistent storage for custom tool definitions
  - [ ] Tool definition format (JSON/YAML)
  - [ ] Tool dependency management
  - [ ] Tool lifecycle hooks
- [ ] Script-Based Tool Factory
  - [ ] Factory interface for creating tools from definitions
  - [ ] Support for multiple scripting languages
  - [ ] Sandboxed execution environment
  - [ ] Tool validation before registration
- [ ] Integration tests for dynamic tool management
- [ ] Examples for script-based tool creation

### 0.3.5.3: Bridge-Friendly Type System
- [ ] Implement Type Registry
  - [ ] Central registry for type conversions
  - [ ] RegisterType method for custom converters
  - [ ] Built-in converters for common types
  - [ ] Bidirectional conversion support
- [ ] Generic Type Converter
  - [ ] Configurable converter with plugin system
  - [ ] Handle primitive types automatically
  - [ ] Support for complex type mappings
  - [ ] Error handling and validation
- [ ] Serialization Helpers
  - [ ] JSON serialization for all domain types
  - [ ] YAML serialization support
  - [ ] Custom serialization formats
  - [ ] Schema-aware serialization
- [ ] Type conversion benchmarks
- [ ] Examples for type bridging

### 0.3.5.4: Event System Enhancements
- [ ] Event Serialization
  - [ ] Implement JSON serialization for all event types
  - [ ] Support for custom event data
  - [ ] Event versioning
  - [ ] Compression options
- [ ] Event Filtering
  - [ ] EventFilter interface implementation
  - [ ] Composite filters (AND/OR/NOT)
  - [ ] Field-based filtering
  - [ ] Event type filtering
- [ ] Event Replay System
  - [ ] EventRecorder interface implementation
  - [ ] Time-based replay
  - [ ] Event persistence options
  - [ ] Replay speed control
- [ ] Event system integration tests
- [ ] Examples for event filtering and replay

### 0.3.5.5: Workflow Serialization and Templates
- [ ] Workflow Serialization
  - [ ] WorkflowSerializer interface implementation
  - [ ] Support JSON/YAML formats
  - [ ] Preserve all workflow metadata
  - [ ] Version compatibility handling
- [ ] Workflow Templates
  - [ ] Pre-built workflow templates
  - [ ] Template customization API
  - [ ] Template registry
  - [ ] Template validation
- [ ] Script-Based Step Definitions
  - [ ] ScriptStep implementation
  - [ ] Multiple language support
  - [ ] Script validation
  - [ ] Error handling for script execution
- [ ] Workflow serialization tests
- [ ] Template examples

### 0.3.5.6: LLM Provider Metadata and Configuration
- [ ] Provider Metadata API
  - [ ] ProviderMetadata interface implementation
  - [ ] Capability discovery
  - [ ] Constraint documentation
  - [ ] Configuration schema generation
- [ ] Dynamic Provider Registration
  - [ ] Runtime provider registration
  - [ ] Provider factory pattern
  - [ ] Provider lifecycle management
  - [ ] Hot-reload support
- [ ] Provider Configuration Templates
  - [ ] JSON/YAML configuration templates
  - [ ] Template validation
  - [ ] Environment variable mapping
  - [ ] Secure credential handling
- [ ] Provider metadata tests
- [ ] Configuration examples

### 0.3.5.7: Structured Output Support
- [ ] Output Parser Interface
  - [ ] Implement parsers for JSON, XML, YAML
  - [ ] Custom parser plugin system
  - [ ] Error recovery in parsing
  - [ ] Partial parsing support
- [ ] Output Validator
  - [ ] Schema-based validation
  - [ ] Custom validation rules
  - [ ] Validation error details
  - [ ] Fix suggestions
- [ ] Format Converters
  - [ ] Convert between JSON/XML/YAML
  - [ ] Preserve type information
  - [ ] Custom format support
  - [ ] Streaming conversion
- [ ] Output parsing benchmarks
- [ ] Validation examples

### 0.3.5.8: Enhanced Error Handling
- [ ] Serializable Error Implementation
  - [ ] JSON serialization for all errors
  - [ ] Rich error context
  - [ ] Stack trace capture
  - [ ] Variable state at error time
- [ ] Error Recovery Strategies
  - [ ] Built-in recovery strategies
  - [ ] Custom strategy registration
  - [ ] Retry mechanisms
  - [ ] Fallback options
- [ ] Error Context Enhancement
  - [ ] Automatic context collection
  - [ ] Custom context providers
  - [ ] Context filtering
  - [ ] Sensitive data masking
- [ ] Error handling tests
- [ ] Recovery strategy examples

### 0.3.5.9: Testing Infrastructure
- [ ] Mock Implementations
  - [ ] Mock for every interface
  - [ ] Configurable mock behaviors
  - [ ] Mock assertion helpers
  - [ ] Mock state verification
- [ ] Test Helpers
  - [ ] Agent testing utilities
  - [ ] Tool testing framework
  - [ ] Workflow test harness
  - [ ] Event capture for tests
- [ ] Scenario Builders
  - [ ] DSL for test scenarios
  - [ ] Fluent API design
  - [ ] Assertion library
  - [ ] Test report generation
- [ ] Testing framework examples
- [ ] Best practices documentation

### 0.3.5.10: Documentation and API Generation
- [ ] API Documentation Generator
  - [ ] Generate OpenAPI specs for tools
  - [ ] Tool capability documentation
  - [ ] Interactive API explorer
  - [ ] Version management
- [ ] Schema Documentation
  - [ ] Generate docs from schemas
  - [ ] Schema visualization
  - [ ] Example generation
  - [ ] Validation rule docs
- [ ] Example Repository Enhancement
  - [ ] Comprehensive examples for all features
  - [ ] Categorized example structure
  - [ ] README for each example
  - [ ] CI for example validation
- [ ] Documentation generation tests
- [ ] Meta-documentation (docs about docs)

## v0.3.6: [Reserved for future features]

## v0.3.7: [Reserved for future features]

## v0.3.8: Advanced Agent Features
### 0.3.8.1: State Management 
- [ ] State persistence and serialization design
- [ ] Implement state storage backends (file, database)
- [ ] Add state versioning and migration
- [ ] Create examples with persistent agents

### 0.3.8.2: Agent Discovery
- [ ] Agent discovery and registry design
- [ ] Implement agent metadata and search
- [ ] Add dynamic agent loading
- [ ] Create agent marketplace example

### 0.3.8.3: Advanced Features 
- [ ] Advanced merge strategies for parallel agents
- [ ] Streaming support for long-running agents
- [ ] Agent composition patterns
- [ ] Agent lifecycle management

## v0.3.9: Model Context Protocol (MCP) Support
### 0.3.9.1: MCP Client Support 
- [ ] Research MCP specification and requirements
- [ ] Design MCP client interface for agents
- [ ] Implement MCP client in pkg/agent/mcp/client
- [ ] Add MCP tool discovery and registration
- [ ] Create examples demonstrating MCP client usage
- [ ] Write comprehensive tests
- [ ] Document MCP client usage in user guide

### 0.3.9.2: MCP Server Support 
- [ ] Design MCP server interface for exposing agents/workflows
- [ ] Implement MCP server in pkg/agent/mcp/server
- [ ] Add agent/workflow registration to MCP server
- [ ] Create example MCP server implementations
- [ ] Write comprehensive tests
- [ ] Document MCP server setup and configuration

## v0.3.10: Built-in Agents Library
### 0.3.10.1: Text Processing Agents 
- [ ] TextSummarize - intelligent summarization using LLM
- [ ] TextExtract - extract structured data from text
- [ ] TextAnalyze - sentiment, entities, keywords
- [ ] TextTranslate - language translation using LLM

### 0.3.10.2: Research Agents 
- [ ] WebResearcher - web research with source tracking
- [ ] DocumentAnalyzer - analyze documents and PDFs
- [ ] FactChecker - verify claims against sources

### 0.3.10.3: Coding Agents 
- [ ] CodeReviewer - review code for issues
- [ ] TestGenerator - generate tests from code
- [ ] DocWriter - generate documentation

### 0.3.10.4: Data Agents 
- [ ] DataAnalyst - analyze datasets and generate insights
- [ ] ReportGenerator - create formatted reports
- [ ] DataCleaner - clean and validate data

### 0.3.10.5: Feed Agents 
- [ ] NewsMonitor - monitor news feeds for keywords and topics using LLM
- [ ] FeedAggregator - aggregate and deduplicate content from multiple feeds
- [ ] FeedSummarizer - summarize feed content using LLM
- [ ] ContentCurator - curate and categorize feed content using LLM

## v0.3.11: Built-in Workflow Patterns
### 0.3.11.1: Core Workflow Patterns 
- [ ] Pipeline - sequential processing workflow
- [ ] MapReduce - parallel processing with aggregation
- [ ] Consensus - multi-agent agreement pattern
- [ ] Retry - with exponential backoff

### 0.3.11.2: Example Workflows 
- [ ] ResearchWorkflow - research → verify → summarize → report
- [ ] CodeReviewWorkflow - analyze → review → suggest → document
- [ ] DataPipeline - ingest → clean → analyze → visualize


## v0.3.12: Performance Optimization
### 0.3.12.1: Profiling Infrastructure (REVISIT)
- [ ] Create benchmark harness for A/B testing optimizations
- [ ] Implement visualization for memory allocation patterns
- [ ] Create real-world test scenarios for end-to-end performance

### 0.3.12.2: Advanced Optimizations (REVISIT)
- [ ] Implement adaptive channel buffer sizing based on usage patterns
- [ ] Add pool prewarming for high-throughput scenarios
- [ ] Reduce redundant property iterations in schema processing
- [ ] Implement more granular locking in cached objects
- [ ] Optimize zero-initialization patterns for pooled objects
- [ ] Introduce buffer pooling for string builders

### 0.3.12.3: Performance Validation (REVISIT)
- [ ] Document performance improvements with metrics
- [ ] Verify optimizations in high-concurrency scenarios
- [ ] Create benchmark comparison charts for before/after
- [ ] Implement regression testing to prevent performance degradation
- [ ] Add performance acceptance criteria to CI pipeline

## v0.3.13: Final Polish and Stable Release
### 0.3.13.1: Documentation Polish
- [ ] Fix identified cross-link issues (path inconsistencies, broken links)
- [ ] Perform final consistency check across all documentation
- [ ] Update all examples to showcase v0.3.x features
- [ ] Create migration guide from earlier versions

### 0.3.13.2: API Refinement
- [ ] API refinement based on usage feedback
- [ ] Deprecate old patterns with migration paths
- [ ] Ensure backward compatibility where possible
- [ ] Final review and preparation for v0.4.0 stable release

## Notes
- Tool System Enhancement Phases 1-4: COMPLETED (see TODO-DONE.md)
- Agent Architecture Restructuring Phases 1-7: COMPLETED (see TODO-DONE.md)
- Performance Phase 2 (Quick Wins): COMPLETED (see TODO-DONE.md)

See TODO-DONE.md for all completed tasks
