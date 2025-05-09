# Go-LLMs Project TODOs

## Features
- [ ] Add Model Context Protocol Client support for Agents
- [ ] Add Model Context Protocol Server support for Workflows or Agents
- [x] Implement interface-based provider option system

## Additional providers
- [x] Add Google Gemini api based provider
  - [x] Fix gemini streaming output by adding the required `alt=sse` parameter to streaming URLs
- [x] Add Ollama support via OpenAI-compatible provider
  - [x] Implemented in openai_api_compatible_providers example
  - [ ] Create dedicated integration test for Ollama
  - [ ] Fix/update integration tests for OpenAI API compatible providers
    - [ ] Skip openai_api_compatible_providers_test.go in integration tests until fixed

## Provider Options Enhancements (Completed) ✅
- [x] Add support for passing provider options directly in ModelConfig
- [x] Implement environment variable support for provider-specific options
- [x] Create option factory functions for common provider configurations
- [x] Implement environment variable support for use case-specific options
- [x] Add support for merging options from environment variables and option factories
- [x] Improve example documentation for provider options

## Documentation
- [x] Consolidate documentation and make sure it's consistent
  - [x] Update REFERENCE.md with all new documentation
  - [x] Update DOCUMENTATION_CONSOLIDATION.md with recent changes
  - [x] Ensure navigation links work correctly
- [x] Create detailed guide on using the provider option system
- [x] Document all common and provider-specific options
- [x] Add examples of combining options across providers

## Examples
- [x] Add example demonstrating schema generation from Go structs
- [x] Create provider_options example
- [x] Update openai_api_compatible_providers example (formerly custom_providers)
- [x] Update all provider examples with provider-specific options
  - [x] Update anthropic example with AnthropicSystemPromptOption
  - [x] Update openai example with OpenAIOrganizationOption
  - [x] Update gemini example with GeminiGenerationConfigOption
  - [x] Update multi example with provider-specific options

## Testing & Performance
- [x] Add benchmarks for consensus algorithms
- [x] Consolidate and cleanup Makefile
- [x] Fix linting errors throughout the codebase
- [x] Create comprehensive test suite for error conditions
  - [x] Tests for provider error conditions
  - [x] Tests for schema validation error conditions
  - [x] Tests for agent error conditions
  - [x] Fix Gemini provider error tests
- [x] Add benchmarks for remaining components
  - [x] Add benchmarks for Gemini provider message conversion
  - [x] Add benchmarks for prompt template processing
  - [x] Add benchmarks for memory pooling
- [x] Implement stress tests for high-load scenarios
  - [x] Provider stress tests for single and multi-provider setups
  - [x] Agent workflow stress tests including MultiAgent and CachedAgent
  - [x] Structured output processor stress tests with various schema complexities
  - [x] Memory pool stress tests (ResponsePool, TokenPool, ChannelPool)
- [ ] Performance profiling and optimization:
  - [ ] Prompt processing and template expansion
  - [ ] Memory pooling for response types
- [ ] Additional test coverage for edge cases
- [ ] Review and preparation for beta release
  - [ ] documentation consolidation including all README.mds and docs/ documentation
- [ ] API refinement based on usage feedback
- [ ] Final review and preparation for stable release

