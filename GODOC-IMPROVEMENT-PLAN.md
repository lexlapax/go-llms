# Godoc Improvement Plan

## Philosophy
Focus on godoc-compatible documentation that tools can parse and display. ABOUTME comments are secondary - they can inform godoc but aren't a replacement.

## Critical Package Documentation Gaps

### Priority 1: Core Package doc.go Files
These packages lack package-level documentation entirely:

1. **pkg/llm/provider/** - The main provider package
   ```go
   // Package provider implements LLM provider integrations for OpenAI, Anthropic, Google, and others.
   // It provides a unified interface for interacting with different LLM services.
   ```

2. **pkg/llm/domain/** - Core domain models
   ```go
   // Package domain defines the core types and interfaces for LLM interactions.
   // It includes message formats, options, and provider interfaces.
   ```

3. **pkg/schema/domain/** - Schema validation domain
   ```go
   // Package domain defines types for JSON schema validation and generation.
   // It provides interfaces for validators, repositories, and generators.
   ```

## Critical Interface Documentation

### Provider Interfaces (pkg/llm/provider/)
- `ProviderFactory` - Factory for creating providers
- `ProviderTemplate` - Template for provider configuration
- `ProviderRegistration` - Registration info for providers
- `DynamicRegistry` - Dynamic provider registry
- `RegistryListener` - Listener for registry events

### Domain Interfaces (pkg/llm/domain/)
- `Provider` - Core provider interface
- `ModelRegistry` - Registry for model information

### Agent Interfaces (pkg/agent/domain/)
- `Tool` - Tool execution interface
- `Agent` - Agent behavior interface
- `AgentRegistry` - Registry for agents

## Critical Type Documentation

### Message Types (pkg/llm/domain/)
- `Role` - Message role enumeration
- `ContentType` - Content type enumeration
- `Message` - Chat message structure
- `ContentPart` - Multimodal content
- `Response` - LLM response structure
- `ResponseStream` - Streaming response interface

### Provider Types (pkg/llm/provider/)
- `Capability` - Provider capability enumeration
- `ModelInfo` - Model information structure
- `ProviderMetadata` - Provider metadata interface
- `BaseProviderMetadata` - Base metadata implementation

## Critical Function Documentation

### Global Functions
- `GetGlobalRegistry()` - Returns the global provider registry
- `RegisterDefaultFactories()` - Registers built-in provider factories
- `CreateProviderFromEnvironment()` - Creates provider from env vars

### Constructor Functions
All `New*` functions need godoc explaining:
- What they create
- Required parameters
- Return values and errors

## Constants Documentation

### Role Constants (pkg/llm/domain/)
```go
// RoleSystem represents system messages that set context or behavior
const RoleSystem Role = "system"

// RoleUser represents messages from the user
const RoleUser Role = "user"

// RoleAssistant represents messages from the AI assistant
const RoleAssistant Role = "assistant"

// RoleTool represents tool/function call results
const RoleTool Role = "tool"
```

### ContentType Constants
Similar pattern - each needs a godoc comment explaining its purpose

## Method Documentation Pattern

For methods, include the receiver type:
```go
// RegisterFactory registers a new provider factory with the given name.
// It returns an error if a factory with that name already exists.
func (r *DynamicRegistry) RegisterFactory(name string, factory ProviderFactory) error {
```

## Quality Metrics

Good godoc should:
1. Start with the name of the item being documented
2. Be a complete sentence with proper grammar
3. Explain what something is/does, not how it's implemented
4. For functions: describe parameters, return values, and errors
5. For types: explain the purpose and common usage
6. For interfaces: describe the contract and implementation requirements

## Implementation Strategy

1. **Package docs first** - These set context for everything else
2. **Interfaces second** - These define contracts users depend on
3. **Public types third** - These are what users work with
4. **Functions fourth** - These are the entry points
5. **Constants last** - These need brief explanations

## Anti-patterns to Avoid

❌ `// SomeType is a type` - Redundant, adds no value
❌ `// Does stuff` - Too vague
❌ Missing parameter/return documentation
❌ Implementation details in public API docs
❌ Inconsistent style across package

## Success Criteria

- `go doc` shows meaningful information for all exported items
- IDEs show helpful tooltips for all public APIs
- New users can understand APIs without reading implementation
- Consistent documentation style throughout codebase