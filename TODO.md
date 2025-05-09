# Go-LLMs Project TODOs

## Features
- [ ] Add Model Context Protocol Client support for Agents
- [ ] Add Model Context Protocol Server support for Workflows or Agents
- [x] Implement interface-based provider option system (see plan below)

## Additional providers
- [x] Add Google Gemini api based provider
--  [ ] fix gemini streaming output
- [ ] Add Ollama local provider which is going to be very similar to OpenAI provider

## Documentation
- [x] Consolidate documentation and make sure it's consistent

## Examples
- [x] Add example demonstrating schema generation from Go structs

## Testing & Performance
- [x] Add benchmarks for consensus algorithms to ensure performance optimization
- [ ] Create comprehensive test suite for error conditions
- [ ] Add benchmarks for remaining components to ensure performance optimization
- [ ] Implement stress tests for high-load scenarios

## Interface-Based Provider Option System Implementation Plan

### Phase 1: Design and Core Interfaces ✅

**Task 1: Design the interface-based provider option system** ✅

Design principles:
- Clean interfaces with no legacy support
- Type safety through Go's interface system
- Composable options that can work across providers
- Provider-specific options for unique features

Core interfaces:
```go
// pkg/llm/domain/options.go

// ProviderOption is the base interface for all provider options
type ProviderOption interface {
    // Identifies which provider type this option is for
    ProviderType() string
}

// Common marker interfaces for provider-specific options
type OpenAIOption interface {
    ProviderOption
    ApplyToOpenAI(*OpenAIProvider)
}

type AnthropicOption interface {
    ProviderOption
    ApplyToAnthropic(*AnthropicProvider)
}

type GeminiOption interface {
    ProviderOption
    ApplyToGemini(*GeminiProvider)
}
```

**Task 2: Write tests for core option interfaces and base implementations** ✅

**Task 3: Implement core option interfaces and base implementations** ✅

Implement common options in `pkg/llm/domain/common_options.go`:
- HTTPClientOption
- BaseURLOption
- TimeoutOption
- RetryOption

### Phase 2: Provider Refactoring ✅

**Tasks 4-5: Refactor OpenAI provider** ✅
- Update constructor to use new options
- Add provider-specific options
- Update tests

**Tasks 6-9: Refactor Anthropic and Gemini providers** ✅
- Update constructors
- Add provider-specific options
- Update tests

### Phase 3: Utility Function Updates ✅

**Task 10: Update llmutil.ModelConfig and CreateProvider** ✅
- Add Options field to ModelConfig
- Update CreateProvider to handle options
- Refactor ProviderFromEnv

**Task 11: Update existing llmutil tests** ✅
- Test option handling in CreateProvider
- Test conversions from ModelConfig

### Phase 4: Examples and Documentation 🔄

**Tasks 12-15: Examples and Documentation** 🔄
- ✅ Create provider_options example
- ✅ Update custom_providers example
- [ ] Update all other examples
- [ ] Update documentation

