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

### 0.3.5.1: Schema Package Implementations (FOUNDATION) ✅ COMPLETED (June 13, 2025)
- [x] Implement InMemorySchemaRepository
  - [x] Thread-safe in-memory storage for schemas
  - [x] CRUD operations for schema management
  - [x] Schema versioning support
  - [x] Export/import functionality
- [x] Implement FileSchemaRepository
  - [x] File-based persistent schema storage
  - [x] Directory structure for schema organization
  - [x] Schema file format (JSON/YAML)
  - [x] Migration support between versions
- [x] Implement ReflectionSchemaGenerator
  - [x] Generate schemas from Go structs using reflection
  - [x] Handle nested structs and slices
  - [x] Support for custom types
  - [x] Preserve Go type information in schema
- [x] Implement TagSchemaGenerator
  - [x] Generate schemas from struct tags (json, validate, etc.)
  - [x] Support multiple tag formats
  - [x] Custom tag handlers
  - [x] Validation rule extraction
- [x] Add comprehensive tests for all implementations
- [x] Create examples demonstrating schema usage

**DOWNSTREAM REQUIREMENTS SATISFIED**:
- ✅ `pkg/schema/repository/memory.go` - InMemoryRepository with thread-safe schema storage
- ✅ `pkg/schema/repository/file.go` - FileRepository with JSON/YAML format support
- ✅ `pkg/schema/generator/reflection.go` - ReflectionGenerator with configurable options (tagName, includePrivate, maxDepth)
- ✅ Bridge-friendly factory methods: `NewInMemoryRepository()`, `NewFileRepository()`, `NewReflectionGenerator()`
- ✅ All implementations follow the domain interfaces exactly as specified in downstream requirements

### 0.3.5.2: Enhanced Error Handling (FOUNDATION) ✅ PARTIAL - Standalone Package Only (June 13, 2025)
- [x] Serializable Error Package Implementation (pkg/errors)
  - [x] JSON serialization for errors (BaseError with ToJSON)
  - [x] Rich error context (map[string]interface{})
  - [x] Stack trace capture (captureStackTrace with filtering)
  - [x] Variable state at error time (Context field)
- [x] Error Recovery Strategies
  - [x] Built-in recovery strategies (exponential, linear, circuit breaker, fallback)
  - [x] Custom strategy registration (RecoveryRegistry)
  - [x] Retry mechanisms (backoff calculations)
  - [x] Fallback options (FallbackStrategy)
- [x] Error Context Enhancement
  - [x] Automatic context collection (EnrichError functions)
  - [x] Custom context providers (ErrorBuilder, ContextProvider)
  - [x] Context filtering (GetAll returns copy)
  - [x] Sensitive data masking (manual via context control)
- [x] Error handling tests (comprehensive test coverage)
- [x] Recovery strategy examples (enhanced_errors example)
- [ ] Library-Wide Error Serialization (REQUIRED FOR DOWNSTREAM)
  - [ ] Convert all pkg/llm provider errors to use pkg/errors
  - [ ] Convert all pkg/agent errors to use pkg/errors
  - [ ] Convert all pkg/schema errors to use pkg/errors
  - [ ] Convert all pkg/structured errors to use pkg/errors
  - [ ] Update error creation patterns throughout codebase
  - [ ] Ensure all errors are JSON serializable
  - [ ] Add migration guide for error handling

**DOWNSTREAM REQUIREMENTS**: 
- ✅ SerializableError interface with Code(), Message(), Context(), ToJSON(), GetRecoveryStrategy()
- ✅ BaseError implementation with all required fields (code, message, context, cause, recovery)
- ✅ Domain-specific errors (AgentError, etc.) that embed BaseError
- ✅ Recovery strategies: RetryOnce, RetryWithBackoff, Failover
- ✅ WrapError function for API boundary error wrapping
- ⚠️ **CRITICAL**: All go-llms errors must implement SerializableError for bridge compatibility
- ⚠️ **CRITICAL**: Error context must be bridge-friendly (map[string]interface{} serializable)

### 0.3.5.3: Enhanced Tool Discovery System
- [ ] Dynamic Tool Registration (REQUIRED FOR DOWNSTREAM)
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
- [ ] Script-Based Tool Factory (CRITICAL FOR SCRIPTING ENGINES)
  - [ ] Factory interface for creating tools from definitions
  - [ ] Support for multiple scripting languages
  - [ ] Sandboxed execution environment
  - [ ] Tool validation before registration
- [ ] Registry Persistence (DOWNSTREAM REQUIREMENT)
  - [ ] SaveRegistry(writer io.Writer) method implementation
  - [ ] LoadRegistry(reader io.Reader) method implementation
  - [ ] Support for tool definitions from external sources
  - [ ] Multi-tenant tool isolation support
- [ ] Integration tests for dynamic tool management
- [ ] Examples for script-based tool creation

**DOWNSTREAM REQUIREMENTS**:
- 🔥 **CRITICAL**: ToolDiscovery interface must support RegisterTool(info ToolInfo, factory ToolFactory) 
- 🔥 **CRITICAL**: Script tools must be registrable at runtime via bridge layer
- 🔥 **CRITICAL**: defaultDiscovery must handle both built-in and dynamic tools with thread safety
- ⚠️ Tool registry persistence for plugin architectures and multi-tenant scenarios
- ⚠️ Runtime tool loading from files, databases, APIs

### 0.3.5.4: Bridge-Friendly Type System (FOUNDATION)
- [ ] Implement Type Registry (CRITICAL FOR DOWNSTREAM)
  - [ ] Central registry for type conversions
  - [ ] RegisterConverter method for custom converters
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
- [ ] Global DefaultRegistry (REQUIRED BY DOWNSTREAM)
  - [ ] Pre-registered common conversions (Schema ↔ map[string]interface{})
  - [ ] Support for multi-hop conversion through intermediate types
  - [ ] Conversion caching for performance
  - [ ] Reverse conversion support (CanReverse method)
- [ ] Type conversion benchmarks
- [ ] Examples for type bridging

**DOWNSTREAM REQUIREMENTS**:
- 🔥 **CRITICAL**: `pkg/util/types/registry.go` with Registry struct and global DefaultRegistry
- 🔥 **CRITICAL**: Convert(from interface{}, toType reflect.Type) method for bridge layer
- 🔥 **CRITICAL**: Pre-registered converters for domain.Schema ↔ map[string]interface{}
- ⚠️ TypeConverter interface with Convert() and CanReverse() methods
- ⚠️ Multi-hop conversion through common types for complex transformations
- ⚠️ Conversion caching to improve performance in bridge scenarios

### 0.3.5.5: Event System Enhancements
- [ ] Event Serialization (CRITICAL FOR DOWNSTREAM)
  - [ ] SerializableEvent interface with MarshalJSON() method
  - [ ] SerializeEvent helper function for any domain.Event
  - [ ] Automatic event type and timestamp inclusion
  - [ ] Support for custom event data
  - [ ] Event versioning
  - [ ] Compression options
- [ ] Event Filtering (REQUIRED FOR BRIDGE LAYER)
  - [ ] Filter interface with Match(event domain.Event) method
  - [ ] PatternFilter with regex pattern matching (e.g., "tool.*")
  - [ ] Composite filters (AND/OR/NOT)
  - [ ] Field-based filtering
  - [ ] Event type filtering
- [ ] Event Replay System
  - [ ] EventRecorder interface implementation
  - [ ] EventStorage interface for different backends
  - [ ] Time-based replay
  - [ ] Event persistence options
  - [ ] Replay speed control
- [ ] Bridge Integration (DOWNSTREAM REQUIREMENT)
  - [ ] Event subscription with pattern-based filtering
  - [ ] Serialized event delivery to bridge handlers
  - [ ] Event context extraction for debugging
- [ ] Event system integration tests
- [ ] Examples for event filtering and replay

**DOWNSTREAM REQUIREMENTS**:
- 🔥 **CRITICAL**: `pkg/agent/events/serialization.go` with SerializeEvent() function
- 🔥 **CRITICAL**: All events must be serializable to map[string]interface{} for bridge layer
- 🔥 **CRITICAL**: PatternFilter for subscribing to event patterns like "tool.*"
- ⚠️ Event replay capabilities for debugging and testing scenarios
- ⚠️ EventStorage abstraction for different persistence backends

### 0.3.5.6: Workflow Serialization and Templates
- [ ] Workflow Serialization (CRITICAL FOR DOWNSTREAM)
  - [ ] WorkflowSerializer with format selection ("json", "yaml")
  - [ ] Serialize(def *WorkflowDefinition) method
  - [ ] DeserializeDefinition for bridge layer workflow creation
  - [ ] Preserve all workflow metadata
  - [ ] Version compatibility handling
- [ ] Workflow Templates
  - [ ] Pre-built workflow templates
  - [ ] Template customization API
  - [ ] Template registry
  - [ ] Template validation
- [ ] Script-Based Step Definitions (REQUIRED FOR SCRIPTING ENGINES)
  - [ ] ScriptStep struct with Script, Language, Handler fields
  - [ ] ScriptHandler interface with Execute(ctx, state, script) method
  - [ ] Support for multiple languages: "javascript", "lua", "tengo", "expr"
  - [ ] RegisterScriptHandler global function for language registration
  - [ ] Script validation before execution
  - [ ] Error handling for script execution
- [ ] Declarative Workflow Support (DOWNSTREAM REQUIREMENT)
  - [ ] JSON/YAML workflow definition format
  - [ ] Script-based step integration in workflows
  - [ ] Dynamic workflow creation from bridge layer
  - [ ] Workflow versioning and migration support
- [ ] Workflow serialization tests
- [ ] Template examples

**DOWNSTREAM REQUIREMENTS**:
- 🔥 **CRITICAL**: `pkg/agent/workflow/serialization.go` with WorkflowSerializer
- 🔥 **CRITICAL**: ScriptStep support for embedding scripts in workflows
- 🔥 **CRITICAL**: DeserializeDefinition for creating workflows from bridge definitions
- 🔥 **CRITICAL**: RegisterScriptHandler for pluggable script language support
- ⚠️ Declarative workflows enable visual builders and no-code tools
- ⚠️ Workflow storage, versioning, and sharing capabilities

### 0.3.5.7: LLM Provider Metadata and Configuration
- [ ] Provider Metadata API (CRITICAL FOR DOWNSTREAM)
  - [ ] ProviderMetadata interface with Name(), Description(), GetCapabilities(), GetModels(), GetConstraints(), GetConfigSchema()
  - [ ] Capability constants: streaming, function_calling, vision, embeddings
  - [ ] ModelInfo struct for model discovery
  - [ ] Constraints struct for limits and rate information
  - [ ] Configuration schema generation for UI
- [ ] MetadataProvider Interface (REQUIRED FOR BRIDGE LAYER)
  - [ ] All providers must implement MetadataProvider interface
  - [ ] GetMetadata() method returning standardized information
  - [ ] Bridge-friendly provider information format
- [ ] Dynamic Provider Registration (DOWNSTREAM REQUIREMENT)
  - [ ] DynamicRegistry extending domain.ModelRegistry
  - [ ] RegisterProvider method with validation
  - [ ] Provider factory pattern using templates
  - [ ] Provider lifecycle management
  - [ ] Hot-reload support
- [ ] Provider Configuration Templates
  - [ ] GetTemplate(type) function for provider templates
  - [ ] CreateProvider from configuration maps
  - [ ] JSON/YAML configuration templates
  - [ ] Template validation against schemas
  - [ ] Environment variable mapping
  - [ ] Secure credential handling
- [ ] Provider metadata tests
- [ ] Configuration examples

**DOWNSTREAM REQUIREMENTS**:
- 🔥 **CRITICAL**: `pkg/llm/providers/metadata.go` with ProviderMetadata interface
- 🔥 **CRITICAL**: MetadataProvider interface for all providers to enable capability discovery
- 🔥 **CRITICAL**: Dynamic provider registration from script configurations
- 🔥 **CRITICAL**: Provider templates for easy provider creation from config maps
- ⚠️ Capability-based provider selection for optimal LLM choice
- ⚠️ UI generation support via configuration schemas

### 0.3.5.8: Structured Output Support
- [ ] Output Parser Interface (CRITICAL FOR DOWNSTREAM)
  - [ ] Parser interface with Parse() and ParseWithRecovery() methods
  - [ ] Parser registry with JSON, XML, YAML implementations
  - [ ] GetParser(format) function for bridge layer
  - [ ] Custom parser plugin system
  - [ ] Error recovery in parsing
  - [ ] Partial parsing support
- [ ] JSON Parser with Recovery (REQUIRED FOR BRIDGE LAYER)
  - [ ] Standard JSON parsing with schema validation
  - [ ] Extract JSON from markdown code blocks
  - [ ] Common issue fixing (trailing commas, quotes, etc.)
  - [ ] Schema-guided extraction as last resort
  - [ ] Configurable strict mode
- [ ] Output Validator (DOWNSTREAM REQUIREMENT)
  - [ ] Validate() function taking output and schema
  - [ ] ValidationResult with detailed error information
  - [ ] Schema-based validation using domain.Schema
  - [ ] Custom validation rules
  - [ ] Validation error details
  - [ ] Fix suggestions
- [ ] Format Converters
  - [ ] Convert between JSON/XML/YAML
  - [ ] Preserve type information
  - [ ] Custom format support
  - [ ] Streaming conversion
- [ ] Bridge Integration Support
  - [ ] Schema conversion from script format to domain.Schema
  - [ ] Result validation with detailed error reporting
  - [ ] Automatic format detection and recovery
- [ ] Output parsing benchmarks
- [ ] Validation examples

**DOWNSTREAM REQUIREMENTS**:
- 🔥 **CRITICAL**: `pkg/llm/outputs/parser.go` with Parser interface and registry
- 🔥 **CRITICAL**: ParseWithRecovery for handling malformed LLM outputs
- 🔥 **CRITICAL**: Validate() function for output verification against schemas
- 🔥 **CRITICAL**: Schema-guided parsing for maximum reliability
- ⚠️ Multiple format support (JSON, XML, YAML) for different LLM output styles
- ⚠️ Markdown code block extraction for common LLM response patterns

### 0.3.5.9: Testing Infrastructure (FOUNDATION SUPPORT)
- [ ] Mock Implementations (REQUIRED FOR DOWNSTREAM)
  - [ ] MockProvider with configurable responses
  - [ ] MockTool for tool testing scenarios
  - [ ] Mock for every major interface
  - [ ] Configurable mock behaviors
  - [ ] Mock assertion helpers
  - [ ] Mock state verification and call tracking
- [ ] Scenario Builder System (CRITICAL FOR BRIDGE TESTING)
  - [ ] NewScenario() function with fluent API
  - [ ] WithMockProvider() with response mapping
  - [ ] WithTool() and WithAgent() setup methods
  - [ ] WithInput() and ExpectOutput() assertion methods
  - [ ] ExpectToolCall() with argument matching
  - [ ] ExpectEvent() for event-driven testing
  - [ ] Run(t) method for test execution
- [ ] Test Helpers
  - [ ] Agent testing utilities
  - [ ] Tool testing framework
  - [ ] Workflow test harness
  - [ ] Event capture for tests
- [ ] MockProvider Pattern Matching (DOWNSTREAM REQUIREMENT)
  - [ ] Response matching based on input patterns
  - [ ] Regex support for flexible response mapping
  - [ ] Call history tracking with timestamps
  - [ ] Provider behavior simulation
- [ ] Assertion and Matcher System
  - [ ] Matcher interface with Match(value) method
  - [ ] Built-in matchers: Contains, Equals, HasField
  - [ ] Custom matcher support
  - [ ] Detailed assertion failure messages
- [ ] Testing framework examples
- [ ] Best practices documentation

**DOWNSTREAM REQUIREMENTS**:
- 🔥 **CRITICAL**: `pkg/testing/scenario.go` with ScenarioBuilder fluent API
- 🔥 **CRITICAL**: MockProvider with pattern-based response matching
- 🔥 **CRITICAL**: Tool and agent testing utilities for bridge scenarios
- 🔥 **CRITICAL**: Event testing support for workflow validation
- ⚠️ Scenario-based testing reduces boilerplate for complex test setups
- ⚠️ Consistent testing patterns across bridge implementations

### 0.3.5.10: Documentation and API Generation
- [ ] API Documentation Generator (CRITICAL FOR DOWNSTREAM)
  - [ ] Generator interface with GenerateOpenAPI(), GenerateMarkdown(), GenerateJSON()
  - [ ] Documentable interface for auto-documentation support
  - [ ] GenerateOpenAPIForTool() function for bridge integration
  - [ ] Tool capability documentation
  - [ ] Interactive API explorer
  - [ ] Version management
- [ ] Auto-Generated Tool Documentation (DOWNSTREAM REQUIREMENT)
  - [ ] OpenAPI 3.0 spec generation from ToolInfo
  - [ ] Automatic request/response schema documentation
  - [ ] Tool discovery documentation for bridge layers
  - [ ] Markdown documentation generation
- [ ] Schema Documentation
  - [ ] Generate docs from schemas
  - [ ] Schema visualization
  - [ ] Example generation
  - [ ] Validation rule docs
- [ ] Documentation Infrastructure (REQUIRED FOR BRIDGE LAYER)
  - [ ] Documentation struct with Name, Description, Examples, Schema, Metadata
  - [ ] Bridge-friendly documentation format
  - [ ] Multi-format documentation support
  - [ ] Documentation builder pattern
- [ ] Example Repository Enhancement
  - [ ] Comprehensive examples for all features
  - [ ] Categorized example structure
  - [ ] README for each example
  - [ ] CI for example validation
- [ ] Documentation generation tests
- [ ] Meta-documentation (docs about docs)

**DOWNSTREAM REQUIREMENTS**:
- 🔥 **CRITICAL**: `pkg/docs/generator.go` with Generator interface
- 🔥 **CRITICAL**: Auto-generation of OpenAPI specs for tools via GenerateOpenAPIForTool()
- 🔥 **CRITICAL**: Documentable interface for auto-documentation of bridge components
- 🔥 **CRITICAL**: Bridge-friendly documentation format (JSON serializable)
- ⚠️ Documentation stays in sync with code through auto-generation
- ⚠️ Multiple documentation formats for different audiences (API, markdown, JSON)

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
