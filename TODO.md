# Go-LLMs Project TODOs

## Features
- [x] Implement interface-based provider option system
- [x] Add multimodal content support to the llm core (completed in v0.2.0)
  - [x] Research a common way to provide files via base64 and mime/type encapsulation to the three major provider apis
  - [x] Implement ContentPart structure with support for text, images, files, videos, and audio
  - [x] Create helper functions for creating different message types (NewTextMessage, NewImageMessage, etc.)
  - [x] Write tests to test multimodal content support
  - [x] Implement provider-specific conversions for each provider
  - [x] Integrate multimodal content documentation into main documentation structure
- [ ] Create multimodal example
  - [ ] Design command-line interface with flags for provider, mode, attachments
  - [ ] Implement file reading and MIME type detection
  - [ ] Create demonstrations for each content type (text, image, audio, video)
  - [ ] Implement mixed mode examples (text + images)
  - [ ] Add error handling for unsupported content types per provider
  - [ ] Write comprehensive README with usage examples
  - [ ] Add unit tests for the example
- [ ] Add Model Context Protocol Client support for Agents
- [ ] Add Model Context Protocol Server support for Workflows or Agents

## Library Migration: Viper/Cobra to Koanf/Kong
- [x] Analysis Phase
  - [x] Analyze current usage of viper and cobra in the codebase
  - [x] Create comprehensive analysis documents (VIPER_COBRA_ANALYSIS.md, VIPER_COBRA_API_USAGE.md)
  - [x] Plan migration from viper to koanf
  - [x] Plan migration from cobra to kong/kongplet
  - [x] Create migration plan document with code mappings (MIGRATION_PLAN_VIPER_COBRA_TO_KOANF_KONG.md)
  - [x] Create detailed implementation guide with code examples (IMPLEMENTATION_GUIDE_VIPER_COBRA_TO_KOANF_KONG.md)
  - [x] Identify dependencies and update strategy
  - [x] Create before/after metrics section in migration plan
- [ ] Implementation Phase
  - [ ] Update go.mod with new dependencies (koanf, kong, kongplete)
  - [ ] Remove viper and cobra dependencies from go.mod
  - [ ] Implement configuration migration (viper to koanf)
    - [ ] Create config.go with koanf implementation
    - [ ] Migrate config loading logic
    - [ ] Update environment variable handling
    - [ ] Update default value management
  - [ ] Implement CLI migration (cobra to kong/kongplet)
    - [ ] Create cli.go with kong structures
    - [ ] Convert commands to kong structs
    - [ ] Implement Run() methods for all commands
    - [ ] Set up kongplete for shell completion
  - [ ] Update main.go to use new structure
  - [ ] Adapt tests for new libraries
    - [ ] Update main_test.go for koanf
    - [ ] Add kong parser tests
    - [ ] Test shell completions
  - [ ] Ensure backward compatibility with existing config files
- [ ] Validation Phase
  - [ ] Test all commands with new implementation
  - [ ] Test configuration loading from all sources (file, env, flags)
  - [ ] Test shell completions for all shells
  - [ ] Performance testing and comparison
  - [ ] Measure and update after-migration metrics (binary size, dependencies)
- [ ] Documentation Phase
  - [ ] Update user documentation for new CLI interface
  - [ ] Create migration guide for existing users
  - [ ] Update examples to use new structure
  - [ ] Update CLAUDE.md with new development commands

## Documentation
- [x] Consolidate documentation and make sure it's consistent
  - [x] Update REFERENCE.md with all new documentation
  - [x] Update DOCUMENTATION_CONSOLIDATION.md with recent changes
  - [x] Ensure navigation links work correctly
- [x] Document multimodal content implementation
  - [x] Create technical documentation in docs/technical/multimodal-content.md
  - [x] Update user guide in docs/user-guide/multimodal-content.md
  - [x] Add multimodal content example to README.md
  - [x] Update version to v0.2.0


## Testing & Performance
- [x] Implement stress tests for high-load scenarios
- [x] Implement multimodal content tests
  - [x] Integration tests for multimodal content
  - [x] Provider-specific multimodal tests (OpenAI, Anthropic, Gemini)
  - [x] Edge case tests for different content types
- [ ] Performance profiling and optimization:
  - [ ] Phase 1: Baseline Profiling Infrastructure (Prerequisites)
    - [x] P0: Add CPU and memory profiling hooks to key operations
    - [x] P0: Add monitoring for cache hit rates and pool statistics
    - [ ] P1: Create benchmark harness for A/B testing optimizations
    - [ ] P2: Implement visualization for memory allocation patterns
    - [ ] P2: Create real-world test scenarios for end-to-end performance

  - [ ] Phase 2: High-Impact Optimizations (Quick Wins)
    - [x] P0: Optimize schema JSON marshaling with faster alternatives
    - [x] P0: Improve schema caching with better key generation
    - [x] P0: Optimize object clearing operations for large response objects
    - [x] P1: Add expiration policy to schema cache to prevent unbounded growth
    - [x] P1: Optimize string builder capacity estimation for complex schemas

  - [ ] Phase 3: Advanced Optimizations (After Initial Improvements)
    - [ ] P1: Implement adaptive channel buffer sizing based on usage patterns
    - [ ] P1: Add pool prewarming for high-throughput scenarios
    - [ ] P1: Reduce redundant property iterations in schema processing
    - [ ] P2: Implement more granular locking in cached objects
    - [ ] P2: Optimize zero-initialization patterns for pooled objects
    - [ ] P2: Introduce buffer pooling for string builders

  - [ ] Phase 4: Integration and Validation (Finalization)
    - [ ] P0: Document performance improvements with metrics
    - [ ] P0: Verify optimizations in high-concurrency scenarios
    - [ ] P1: Create benchmark comparison charts for before/after
    - [ ] P1: Implement regression testing to prevent performance degradation
    - [ ] P2: Add performance acceptance criteria to CI pipeline
  [x] Review and preparation for beta release
  - [x] Enhanced Gemini provider documentation (API, examples, and options)
  - [x] Updated OpenAI API Compatible providers documentation (Ollama, OpenRouter, Groq)
  - [x] Documented performance optimizations in technical documentation
    - [x] Schema caching with LRU eviction and TTL expiration
    - [x] Object clearing optimizations for large response objects
  - [x] Verified cross-links between documentation files
  - [ ] Fix identified cross-link issues (path inconsistencies, broken links)
  - [ ] Perform final consistency check across all documentation
- [x] Revisit openai_api_compatible_providers
  - [x] Documented Ollama integration
  - [x] Documented OpenRouter integration
  - [x] Added documentation for Groq integration
- [ ] API refinement based on usage feedback
- [ ] Final review and preparation for stable release

