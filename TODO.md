# Go-LLMs Project TODOs

## Features
- [ ] Add Model Context Protocol Client support for Agents
- [ ] Add Model Context Protocol Server support for Workflows or Agents
- [x] Implement interface-based provider option system

## Documentation
- [x] Consolidate documentation and make sure it's consistent
  - [x] Update REFERENCE.md with all new documentation
  - [x] Update DOCUMENTATION_CONSOLIDATION.md with recent changes
  - [x] Ensure navigation links work correctly


## Testing & Performance
- [x] Implement stress tests for high-load scenarios
- [ ] Performance profiling and optimization:
  - [ ] Phase 1: Baseline Profiling Infrastructure (Prerequisites)
    - [x] P0: Add CPU and memory profiling hooks to key operations
    - [ ] P0: Add monitoring for cache hit rates and pool statistics
    - [ ] P1: Create benchmark harness for A/B testing optimizations
    - [ ] P2: Implement visualization for memory allocation patterns
    - [ ] P2: Create real-world test scenarios for end-to-end performance

  - [ ] Phase 2: High-Impact Optimizations (Quick Wins)
    - [ ] P0: Optimize schema JSON marshaling with faster alternatives
    - [ ] P0: Improve schema caching with better key generation
    - [ ] P0: Optimize object clearing operations for large response objects
    - [ ] P1: Add expiration policy to schema cache to prevent unbounded growth
    - [ ] P1: Optimize string builder capacity estimation for complex schemas

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
  [ ] Review and preparation for beta release
  - [ ] documentation consolidation including all README.mds and docs/ documentation
- [ ] Revisit openai_api_compatible_providers
  - [ ] redo ollama
  - [ ] redo openrouter
  - [ ] add groq.com
- [ ] API refinement based on usage feedback
- [ ] Final review and preparation for stable release

