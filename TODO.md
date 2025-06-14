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

### 0.3.5.2: Enhanced Error Handling (FOUNDATION) ✅ COMPLETED (June 14, 2025)
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
- [x] Library-Wide Error Serialization (REQUIRED FOR DOWNSTREAM)
  - [x] Convert all pkg/llm provider errors to use pkg/errors
  - [x] Convert all pkg/agent errors to use pkg/errors
  - [x] Convert all pkg/schema errors to use pkg/errors (N/A - no errors in schema pkg)
  - [x] Convert all pkg/structured errors to use pkg/errors (N/A - no errors in structured pkg)
  - [x] Update error creation patterns throughout codebase
  - [x] Ensure all errors are JSON serializable
  - [x] Add migration guide for error handling (docs/technical/error-handling-migration.md)

**DOWNSTREAM REQUIREMENTS**: 
- ✅ SerializableError interface with Code(), Message(), Context(), ToJSON(), GetRecoveryStrategy()
- ✅ BaseError implementation with all required fields (code, message, context, cause, recovery)
- ✅ Domain-specific errors (AgentError, etc.) that embed BaseError
- ✅ Recovery strategies: RetryOnce, RetryWithBackoff, Failover
- ✅ WrapError function for API boundary error wrapping
- ⚠️ **CRITICAL**: All go-llms errors must implement SerializableError for bridge compatibility
- ⚠️ **CRITICAL**: Error context must be bridge-friendly (map[string]interface{} serializable)

### 0.3.5.3: Enhanced Tool Discovery System ✅ COMPLETED (June 14, 2025)
- [x] Dynamic Tool Registration (REQUIRED FOR DOWNSTREAM)
  - [x] Add RegisterTool method to ToolDiscovery interface
  - [x] Add UnregisterTool method for runtime removal
  - [x] Add GetRegisteredTools for listing all tools
  - [x] Thread-safe registration/unregistration
  - [x] Tool versioning support
- [x] Tool Metadata Repository
  - [x] Persistent storage for custom tool definitions
  - [x] Tool definition format (JSON/YAML)
  - [x] Tool dependency management
  - [x] Tool lifecycle hooks
- [x] Script-Based Tool Factory (CRITICAL FOR SCRIPTING ENGINES)
  - [x] Factory interface for creating tools from definitions
  - [x] Support for multiple scripting languages
  - [x] Sandboxed execution environment
  - [x] Tool validation before registration
- [x] Registry Persistence (DOWNSTREAM REQUIREMENT)
  - [x] SaveRegistry(writer io.Writer) method implementation
  - [x] LoadRegistry(reader io.Reader) method implementation
  - [x] Support for tool definitions from external sources
  - [x] Multi-tenant tool isolation support
- [x] update toolgen (internal/toolgen) with new metadata, fields and apis
- [x] Integration tests for dynamic tool management
- [x] Examples for script-based tool creation

**DOWNSTREAM REQUIREMENTS**:
- 🔥 **CRITICAL**: ToolDiscovery interface must support RegisterTool(info ToolInfo, factory ToolFactory) 
- 🔥 **CRITICAL**: Script tools must be registrable at runtime via bridge layer
- 🔥 **CRITICAL**: defaultDiscovery must handle both built-in and dynamic tools with thread safety
- ⚠️ Tool registry persistence for plugin architectures and multi-tenant scenarios
- ⚠️ Runtime tool loading from files, databases, APIs

### 0.3.5.4: Bridge-Friendly Type System (FOUNDATION) ✅ COMPLETED (June 14, 2025)
- [x] Implement Type Registry (CRITICAL FOR DOWNSTREAM)
  - [x] Central registry for type conversions
  - [x] RegisterConverter method for custom converters
  - [x] Built-in converters for common types
  - [x] Bidirectional conversion support
- [x] Generic Type Converter
  - [x] Configurable converter with plugin system
  - [x] Handle primitive types automatically
  - [x] Support for complex type mappings
  - [x] Error handling and validation
- [x] Serialization Helpers
  - [x] JSON serialization for all domain types
  - [x] YAML serialization support
  - [x] Custom serialization formats
  - [x] Schema-aware serialization
- [x] Global DefaultRegistry (REQUIRED BY DOWNSTREAM)
  - [x] Pre-registered common conversions (Schema ↔ map[string]interface{})
  - [x] Support for multi-hop conversion through intermediate types
  - [x] Conversion caching for performance
  - [x] Reverse conversion support (CanReverse method)
- [x] Type conversion benchmarks
- [x] Examples for type bridging

**DOWNSTREAM REQUIREMENTS**:
- 🔥 **CRITICAL**: `pkg/util/types/registry.go` with Registry struct and global DefaultRegistry
- 🔥 **CRITICAL**: Convert(from interface{}, toType reflect.Type) method for bridge layer
- 🔥 **CRITICAL**: Pre-registered converters for domain.Schema ↔ map[string]interface{}
- ⚠️ TypeConverter interface with Convert() and CanReverse() methods
- ⚠️ Multi-hop conversion through common types for complex transformations
- ⚠️ Conversion caching to improve performance in bridge scenarios

### 0.3.5.5: Event System Enhancements ✅ COMPLETED (June 13, 2025)
- [x] Event Serialization (CRITICAL FOR DOWNSTREAM)
  - [x] SerializableEvent interface with MarshalJSON() method
  - [x] SerializeEvent helper function for any domain.Event
  - [x] Automatic event type and timestamp inclusion
  - [x] Support for custom event data
  - [x] Event versioning
  - [x] Compression options (via CompactSerializer)
- [x] Event Filtering (REQUIRED FOR BRIDGE LAYER)
  - [x] Filter interface with Match(event domain.Event) method
  - [x] PatternFilter with regex pattern matching (e.g., "tool.*")
  - [x] Composite filters (AND/OR/NOT)
  - [x] Field-based filtering
  - [x] Event type filtering
- [x] Event Replay System
  - [x] EventRecorder interface implementation
  - [x] EventStorage interface for different backends
  - [x] Time-based replay
  - [x] Event persistence options
  - [x] Replay speed control
- [x] Bridge Integration (DOWNSTREAM REQUIREMENT)
  - [x] Event subscription with pattern-based filtering
  - [x] Serialized event delivery to bridge handlers
  - [x] Event context extraction for debugging
- [x] Event system integration tests
- [x] Examples for event filtering and replay

**DOWNSTREAM REQUIREMENTS SATISFIED**:
- ✅ `pkg/agent/events/serialization.go` with SerializeEvent() and DeserializeEvent() functions
- ✅ All events serializable to map[string]interface{} via SerializableEvent wrapper
- ✅ PatternFilter with wildcard support for patterns like "tool.*", "agent.*"
- ✅ Event replay system with EventRecorder, EventReplayer, and speed control
- ✅ EventStorage interface with MemoryStorage and FileStorage implementations
- ✅ BridgeEvent types and utilities for go-llmspell integration
- ✅ Comprehensive filtering system with composite filters (AND/OR/NOT)
- ✅ Multiple serializers (JSON, JSON-pretty, compact) for different use cases

### 0.3.5.6: Workflow Serialization and Templates ✅ COMPLETED (June 13, 2025)
- [x] Workflow Serialization (CRITICAL FOR DOWNSTREAM)
  - [x] WorkflowSerializer with format selection ("json", "yaml")
  - [x] Serialize(def *WorkflowDefinition) method
  - [x] DeserializeDefinition for bridge layer workflow creation
  - [x] Preserve all workflow metadata
  - [x] Version compatibility handling
- [x] Workflow Templates
  - [x] Pre-built workflow templates
  - [x] Template customization API
  - [x] Template registry
  - [x] Template validation
- [x] Script-Based Step Definitions (REQUIRED FOR SCRIPTING ENGINES)
  - [x] ScriptStep struct with Script, Language, Handler fields
  - [x] ScriptHandler interface with Execute(ctx, state, script) method
  - [x] Support for multiple languages: "javascript", "lua", "tengo", "expr"
  - [x] RegisterScriptHandler global function for language registration
  - [x] Script validation before execution
  - [x] Error handling for script execution
- [x] Declarative Workflow Support (DOWNSTREAM REQUIREMENT)
  - [x] JSON/YAML workflow definition format
  - [x] Script-based step integration in workflows
  - [x] Dynamic workflow creation from bridge layer
  - [x] Workflow versioning and migration support
- [x] Workflow serialization tests
- [x] Template examples

**DOWNSTREAM REQUIREMENTS SATISFIED**:
- ✅ `pkg/agent/workflow/serialization.go` with JSON and YAML WorkflowSerializer implementations
- ✅ ScriptStep in `script_step.go` with full builder pattern and validation
- ✅ DeserializeDefinition accepts map[string]interface{} from bridge layer
- ✅ RegisterScriptHandler with global registry and language discovery
- ✅ Mock handlers for javascript, expr, and json-transform (ready for real implementations)
- ✅ Workflow templates with variable substitution and categorization
- ✅ Comprehensive serialization preserving all metadata and versioning

### 0.3.5.7: LLM Provider Metadata and Configuration ✅ COMPLETED (June 14, 2025)

### 0.3.5.8: Structured Output Support ✅ COMPLETED (June 14, 2025)
- [x] Output Parser Interface (CRITICAL FOR DOWNSTREAM)
  - [x] Parser interface with Parse() and ParseWithRecovery() methods
  - [x] Parser registry with JSON, XML, YAML implementations
  - [x] GetParser(format) function for bridge layer
  - [x] Custom parser plugin system
  - [x] Error recovery in parsing
  - [x] Partial parsing support
- [x] JSON Parser with Recovery (REQUIRED FOR BRIDGE LAYER)
  - [x] Standard JSON parsing with schema validation
  - [x] Extract JSON from markdown code blocks
  - [x] Common issue fixing (trailing commas, quotes, etc.)
  - [x] Schema-guided extraction as last resort
  - [x] Configurable strict mode
- [x] Output Validator (DOWNSTREAM REQUIREMENT)
  - [x] Validate() function taking output and schema
  - [x] ValidationResult with detailed error information
  - [x] Schema-based validation using OutputSchema
  - [x] Custom validation rules
  - [x] Validation error details
  - [x] Fix suggestions
- [x] Format Converters
  - [x] Convert between JSON/XML/YAML
  - [x] Preserve type information
  - [x] Custom format support
  - [x] Streaming conversion
- [x] Bridge Integration Support
  - [x] Schema conversion from script format to OutputSchema
  - [x] Result validation with detailed error reporting
  - [x] Automatic format detection and recovery
- [ ] Output parsing benchmarks (deferred to v0.3.6)
- [x] Validation examples

**DOWNSTREAM REQUIREMENTS SATISFIED**:
- ✅ `pkg/llm/outputs/parser.go` with Parser interface and registry
- ✅ ParseWithRecovery for handling malformed LLM outputs
- ✅ Validate() function for output verification against schemas
- ✅ Schema-guided parsing for maximum reliability
- ✅ Multiple format support (JSON, XML, YAML) for different LLM output styles
- ✅ Markdown code block extraction for common LLM response patterns
- ✅ BridgeAdapter for go-llmspell integration
- ✅ Comprehensive recovery strategies for each format
- ✅ OutputSchema type independent from domain.Schema for flexibility

### 0.3.5.9: Testing Infrastructure (FOUNDATION SUPPORT) ✅ COMPLETED (June 14, 2025)
- [x] Inventory and take stock of testing infrastructure including Mock implementations
- [x] Come up with a comprehensive plan for testing infrastructure including common Mock Implementations in an exportable api
- [x] Update this todo.md list for a more comprehensive task list

#### Phase 1: Core Testing Package Structure ✅ COMPLETED
- [x] Expand pkg/testutils package structure
  - [x] Create mocks/ subdirectory for all mock implementations
  - [x] Create scenario/ subdirectory for scenario builder
  - [x] Create fixtures/ subdirectory for pre-configured mocks
  - [x] Create helpers/ subdirectory for test utilities
- [x] Migrate existing testutils files to new structure
  - [x] Move mock_providers.go content to mocks/provider.go
    - [x] Extract TestMockProvider to mocks/provider.go
    - [x] Extract CustomMockProvider to mocks/provider.go
    - [x] Extract MockStructuredProvider to mocks/provider.go
    - [x] Enhance with pattern-based responses and call tracking
  - [x] Move mock_tools.go content to mocks/tool.go
    - [x] Move MockTool struct to mocks/tool.go
    - [x] Enhance with call history tracking
    - [x] Add response mapping functionality
  - [x] Move pointer_helpers.go to helpers/pointers.go
    - [x] Keep as-is (already well-designed helper functions)
  - [x] Keep fixed_test.go in testutils root
    - [x] This is a test file demonstrating correct usage
  - [x] Create compatibility aliases in original files
    - [x] Type aliases pointing to new locations  
    - [x] Deprecation notices in comments
    - [x] Maintain backward compatibility during migration
- [x] Mock Implementations (REQUIRED FOR DOWNSTREAM)
  - [x] MockProvider with configurable responses
    - [x] Pattern-based response mapping (string patterns to responses)
    - [x] Call history tracking with ProviderCall struct
    - [x] Thread-safe implementation with sync.RWMutex
    - [x] Behavior hooks (OnGenerate, OnStream, OnGenerateSchema)
    - [x] Response delay simulation
  - [x] MockTool for tool testing scenarios
    - [x] Input pattern to response mapping
    - [x] Call history with ToolCall struct
    - [x] Execution count tracking
    - [x] Expected calls verification
    - [x] OnExecute and OnValidate hooks
  - [x] MockAgent implementation
    - [x] Response queue for deterministic testing
    - [x] Sub-agent management
    - [x] Event emission tracking
    - [x] State history recording
    - [x] OnStart and OnStep hooks
  - [x] MockState with state manipulation helpers
    - [x] Change tracking with StateChange history
    - [x] State snapshots with diff functionality
    - [x] Access tracking (get/set counts)
    - [x] Behavior hooks (OnGet, OnSet, OnDelete)
    - [x] Failure mode simulation
  - [x] MockEventEmitter for event testing
    - [x] Event recording and filtering
    - [x] Event listeners with async support
    - [x] Behavior hooks for all emit types
    - [x] Event assertions (count, type, content)
    - [x] WaitForEvent with timeout
  - [x] Mock registry for centralized management
    - [x] Register/unregister mocks
    - [x] Lookup by name/type
    - [x] Reset all mocks functionality
  - [x] Comprehensive test coverage for all mock implementations
    - [x] MockAgent test coverage (all features)
    - [x] MockState test coverage (all features)
    - [x] MockEventEmitter test coverage (all features)
    - [x] CreateMockToolContext helper tested
  - [x] Fixed import cycles and race conditions
    - [x] Removed circular dependency between mocks/tool.go and pkg/agent/tools
    - [x] Fixed race conditions in LLMAgent, workflow agents (Sequential, Parallel, Conditional, Loop)
    - [x] Fixed logic issues in error handling tests

**DOWNSTREAM REQUIREMENTS SATISFIED**:
- ✅ MockProvider with pattern-based response matching and call history
- ✅ MockTool with input pattern mapping and execution tracking
- ✅ MockAgent with response queue, sub-agent management, and event tracking
- ✅ MockState with change tracking, snapshots, and access counting
- ✅ MockEventEmitter with recording, filtering, and assertions
- ✅ Thread-safe implementations for all mocks
- ✅ Comprehensive test coverage demonstrating usage

#### Phase 2: Scenario Builder System (CRITICAL FOR BRIDGE TESTING)
- [ ] Core ScenarioBuilder implementation
  - [ ] NewScenario(t testing.TB) constructor
  - [ ] Fluent API method chaining support
  - [ ] Internal state management
  - [ ] Error accumulation and reporting
- [ ] Configuration methods
  - [ ] WithMockProvider(name string, responses map[string]Response)
  - [ ] WithTool(tool *MockTool)
  - [ ] WithAgent(agent *MockAgent)
  - [ ] WithInput(key string, value interface{})
  - [ ] WithState(state domain.State)
  - [ ] WithTimeout(duration time.Duration)
- [ ] Expectation methods
  - [ ] ExpectOutput(matcher Matcher)
  - [ ] ExpectToolCall(toolName string, inputMatcher Matcher)
  - [ ] ExpectEvent(eventType string, dataMatcher Matcher)
  - [ ] ExpectError(errorMatcher Matcher)
  - [ ] ExpectNoError()
- [ ] Execution and verification
  - [ ] Run() domain.State method
  - [ ] RunWithContext(ctx context.Context)
  - [ ] Automatic assertion execution
  - [ ] Detailed failure reporting

#### Phase 3: Matcher System
- [ ] Core Matcher interface
  - [ ] Match(value interface{}) (bool, string) method
  - [ ] Description() string method
- [ ] Basic matchers
  - [ ] Equals(expected interface{})
  - [ ] Contains(substring string)
  - [ ] HasField(field string, valueMatcher Matcher)
  - [ ] IsNil()
  - [ ] IsNotNil()
- [ ] Advanced matchers
  - [ ] MatchesJSON(pattern string)
  - [ ] MatchesRegex(pattern string)
  - [ ] HasLength(expected int)
  - [ ] IsEmpty()
  - [ ] IsBetween(min, max interface{})
- [ ] Composite matchers
  - [ ] AllOf(matchers ...Matcher)
  - [ ] AnyOf(matchers ...Matcher)
  - [ ] Not(matcher Matcher)
- [ ] Custom matcher support
  - [ ] MatcherFunc type for inline matchers
  - [ ] Matcher builder helpers

#### Phase 4: Test Helpers and Utilities
- [ ] Context helpers
  - [ ] CreateTestToolContext(options ...ContextOption)
  - [ ] CreateTestAgentContext(options ...ContextOption)
  - [ ] WithTestState(state domain.State)
  - [ ] WithTestLogger(logger Logger)
- [ ] Event testing support
  - [ ] EventCapture implementation
  - [ ] Event filtering by type/data
  - [ ] Event assertion helpers
  - [ ] Event timeline visualization
- [ ] State testing utilities
  - [ ] State diff functionality
  - [ ] State snapshot comparison
  - [ ] State mutation helpers
- [ ] Provider testing utilities
  - [ ] Response generation from schemas
  - [ ] Error injection helpers
  - [ ] Streaming simulation

#### Phase 5: Test Fixtures
- [ ] Provider fixtures
  - [ ] ChatGPTMockProvider() with typical responses
  - [ ] ClaudeMockProvider() with typical responses
  - [ ] ErrorMockProvider(errorType string)
  - [ ] SlowMockProvider(delay time.Duration)
  - [ ] StreamingMockProvider()
- [ ] Tool fixtures
  - [ ] CalculatorMockTool() with arithmetic operations
  - [ ] WebSearchMockTool() with sample results
  - [ ] FileMockTool() with virtual filesystem
  - [ ] ErrorMockTool(errorRate float64)
- [ ] Agent fixtures
  - [ ] SimpleMockAgent() for basic testing
  - [ ] ResearchMockAgent() with sub-agents
  - [ ] WorkflowMockAgent() with steps
  - [ ] StatefulMockAgent() with complex state
- [ ] State fixtures
  - [ ] EmptyTestState()
  - [ ] PopulatedTestState(data map[string]interface{})
  - [ ] WithTestData(key string, value interface{})

#### Phase 6: Migration and Integration
- [ ] Migrate existing mocks
  - [ ] Update pkg/llm/provider/mock.go to use new structure
  - [ ] Update pkg/testutils mocks to use new structure
  - [ ] Update scattered test mocks to use fixtures
  - [ ] Deprecation notices for old patterns
- [ ] Update existing tests
  - [ ] Identify tests using old mock patterns
  - [ ] Convert to scenario-based testing where appropriate
  - [ ] Update assertions to use matcher system
  - [ ] Verify no test regressions
- [ ] Integration with existing test commands
  - [ ] Update Makefile test targets
  - [ ] Add testutils enhancements to test coverage
  - [ ] Ensure build tags work correctly

#### Phase 7: Documentation and Examples
- [ ] API documentation
  - [ ] Package-level documentation for pkg/testutils
  - [ ] Godoc for all exported types and functions
  - [ ] Usage examples in documentation
- [ ] Testing guide
  - [ ] Migration guide from old patterns
  - [ ] Best practices for mock usage
  - [ ] Scenario building patterns
  - [ ] Common testing recipes
- [ ] Example test files
  - [ ] Provider testing examples
  - [ ] Tool testing examples
  - [ ] Agent testing examples
  - [ ] Workflow testing examples
  - [ ] Integration testing examples
- [ ] Performance benchmarks
  - [ ] Mock performance benchmarks
  - [ ] Scenario builder overhead measurement
  - [ ] Comparison with old patterns

**DOWNSTREAM REQUIREMENTS**:
- 🔥 **CRITICAL**: `pkg/testutils/scenario/builder.go` with ScenarioBuilder fluent API
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
