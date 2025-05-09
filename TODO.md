# Go-LLMs Project TODOs

## Features
- [ ] Add Model Context Protocol Client support for Agents
- [ ] Add Model Context Protocol Server support for Workflows or Agents
- [x] Implement interface-based provider option system

## Additional providers
- [x] Add Google Gemini api based provider
  - [ ] Fix gemini streaming output
- [x] Add Ollama support via OpenAI-compatible provider
  - [x] Implemented in openai_api_compatible_providers example
  - [ ] Create dedicated integration test for Ollama

## Provider Options Enhancements
- [x] Add support for passing provider options directly in ModelConfig
- [x] Implement environment variable support for provider-specific options
- [x] Create option factory functions for common provider configurations

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
- [ ] Consolidate and cleanup Makefile
- [ ] Create comprehensive test suite for error conditions
- [ ] Add benchmarks for remaining components
- [ ] Implement stress tests for high-load scenarios
- [ ] Performance profiling and optimization:
  - [ ] Prompt processing and template expansion
  - [ ] Memory pooling for response types
- [ ] API refinement based on usage feedback
- [ ] Additional test coverage for edge cases
- [ ] Final review and preparation for stable release

