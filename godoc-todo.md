# Godoc Documentation TODO List

This file tracks all exported items in the pkg directory that need godoc-compatible documentation.

## Summary
**Total Missing**: 39 exported items across 14 packages
**Status**: ✅ ALL COMPLETED (100% - 39/39 items)

## High Priority (Core APIs and Constructors)

### Provider Constructors and Interfaces
- [x] pkg/llm/provider/ollama.go:36 - `type OllamaOption interface` - Add godoc for interface
- [x] pkg/llm/provider/ollama.go:93 - `func NewOllamaProvider` - Add godoc for constructor
- [x] pkg/llm/provider/openrouter.go:23 - `func NewOpenRouterProvider` - Add godoc for constructor (already has godoc)

### Core Agent Functions
- [x] pkg/agent/core/llm_agent.go:108 - `func NewAgentFromString` - Add godoc for agent creation
- [x] pkg/agent/tools/base_tool.go:61 - `func NewTool` - Add godoc for tool constructor

### Schema/Validation
- [x] pkg/schema/validation/validator.go:45 - `func NewValidator` - Add godoc for validator constructor (already has godoc)
- [x] pkg/schema/validation/custom_validator.go:37 - `type ExtendedProperty` - Add godoc for type

## Medium Priority (Tool Functions)

### Built-in Tools - System
- [x] pkg/agent/builtins/tools/system/env_var.go:261 - `func GetEnvironmentVariable` - Add godoc
- [x] pkg/agent/builtins/tools/system/execute.go:409 - `func ExecuteCommand` - Add godoc
- [x] pkg/agent/builtins/tools/system/process_list.go:268 - `func ProcessList` - Add godoc
- [x] pkg/agent/builtins/tools/system/system_info.go:358 - `func GetSystemInfo` - Add godoc

### Built-in Tools - Web
- [x] pkg/agent/builtins/tools/web/fetch.go:226 - `func WebFetch` - Add godoc
- [x] pkg/agent/builtins/tools/web/http_request.go:211 - `func HTTPRequest` - Add godoc
- [x] pkg/agent/builtins/tools/web/scrape.go:319 - `func WebScrape` - Add godoc (already has godoc)
- [x] pkg/agent/builtins/tools/web/search.go:348 - `func WebSearch` - Add godoc

### Built-in Tools - Math
- [x] pkg/agent/builtins/tools/math/calculator.go:162 - `func Calculator` - Add godoc (already has godoc)

### Utility Functions
- [x] pkg/util/llmutil/llmutil.go:133 - `func ProviderFromEnv` - Add godoc
- [x] pkg/util/llmutil/model_inventory.go:63 - `func GetAvailableModels` - Add godoc (already has godoc)
- [x] pkg/util/llmutil/option_factories.go:276 - `func CreateOptionFactoryFromEnv` - Add godoc (already has godoc)
- [x] pkg/util/llmutil/provider_parser.go:18 - `func ParseProviderModelString` - Add godoc

### Model Fetchers
- [x] pkg/util/llmutil/modelinfo/fetchers/google_fetcher.go:28 - `func NewGoogleFetcher` - Add godoc (already has godoc)
- [x] pkg/util/llmutil/modelinfo/fetchers/ollama_fetcher.go:29 - `func NewOllamaFetcher` - Add godoc (already has godoc)
- [x] pkg/util/llmutil/modelinfo/fetchers/openai_fetcher.go:28 - `func NewOpenAIFetcher` - Add godoc (already has godoc)

## Low Priority (Type Definitions and Test Utilities)

### OpenAPI Types (pkg/agent/builtins/tools/web/openapi.go)
- [x] Line 209: `type TagObject` - Add godoc
- [x] Line 215: `type Example` - Add godoc
- [x] Line 222: `type Header` - Add godoc
- [x] Line 235: `type Link` - Add godoc
- [x] Line 244: `type Callback` - Add godoc
- [x] Line 246: `type Encoding` - Add godoc
- [x] Line 254: `type Discriminator` - Add godoc
- [x] Line 259: `type XML` - Add godoc
- [x] Line 267: `type ExternalDocs` - Add godoc
- [x] Line 272: `type OAuthFlows` - Add godoc
- [x] Line 279: `type OAuthFlow` - Add godoc

### Other Types
- [x] pkg/llm/domain/pool.go:236 - `func ZeroString` - Add godoc (already has godoc)
- [x] pkg/llm/provider/metadata_integration.go:103 - `type ProviderComparison` - Add godoc

### Test Utilities (pkg/testutils/fixtures/providers.go)
- [x] Line 593: `type RateLimitError` - Add godoc
- [x] Line 602: `type AuthenticationError` - Add godoc
- [x] Line 610: `type NetworkError` - Add godoc

## Guidelines for Writing Godoc

1. **Format**: Comment must start with the name of the item
   ```go
   // Calculator returns a new calculator tool instance
   func Calculator() domain.Tool {
   ```

2. **Types**: Describe what the type represents
   ```go
   // OllamaOption configures options for the Ollama provider
   type OllamaOption interface {
   ```

3. **Functions**: Describe what the function does, not how
   ```go
   // NewOllamaProvider creates a new Ollama provider with the given options
   func NewOllamaProvider(opts ...OllamaOption) (Provider, error) {
   ```

4. **Methods**: Include the receiver type name
   ```go
   // Execute runs the tool with the given parameters
   func (t *Tool) Execute(ctx context.Context, params interface{}) (interface{}, error) {
   ```

## Notes
- Focus on high-priority items first (core APIs and constructors)
- Group related items for efficient updates
- Test utilities are lowest priority
- Some files may have been updated since this scan