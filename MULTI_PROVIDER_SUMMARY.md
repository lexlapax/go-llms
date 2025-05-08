# MultiProvider Implementation for Go-LLMs

## Overview

The `MultiProvider` implementation enables concurrent processing across multiple LLM providers. It addresses several key use cases:

1. **Performance Optimization** - Get responses from the fastest provider 
2. **High Availability** - Implement fallback mechanisms when primary providers fail
3. **Response Quality** - Future versions can implement consensus mechanisms across multiple providers
4. **A/B Testing** - Compare responses from different providers for evaluation

## Implementation Details

The implementation consists of the following key components:

### 1. Provider Architecture

- `MultiProvider` struct that implements the `domain.Provider` interface
- Hybrid execution model:
  - Sequential execution for primary strategy (for deterministic behavior)
  - Concurrent execution for fastest and consensus strategies
- Support for all provider interface methods: `Generate`, `GenerateMessage`, `GenerateWithSchema`, `Stream`, and `StreamMessage`

```go
// MultiProvider implements domain.Provider interface and distributes operations
// across multiple underlying providers with fallback and selection strategies
type MultiProvider struct {
    providers        []ProviderWeight
    selectionStrat   SelectionStrategy
    defaultTimeout   time.Duration
    primaryProvider  int // Index of primary provider for StrategyPrimary
    consensusConfig  consensusConfig // Configuration for consensus algorithms
}
```

### 2. Provider Weighting System

The system supports weighted providers through the `ProviderWeight` struct:

```go
// ProviderWeight defines the weight of a provider in a multi-provider setup
type ProviderWeight struct {
    Provider domain.Provider
    Weight   float64 // 0.0 to 1.0, default 1.0
    Name     string  // Optional name for the provider
}
```

### 3. Result Selection Strategies

Three main selection strategies are provided:

1. **StrategyFastest** - Returns the first successful result (fastest provider wins)
2. **StrategyPrimary** - Uses the designated primary provider, falling back to others on failure
   - Implemented using sequential execution for deterministic behavior
3. **StrategyConsensus** - Enhanced implementations with multiple consensus algorithms:
   - **ConsensusMajority** - Simple majority voting
   - **ConsensusSimilarity** - Groups responses by similarity and chooses largest group
   - **ConsensusWeighted** - Considers provider weights when determining consensus

### 4. Comprehensive Error Handling

- Detailed error aggregation when all providers fail
- Context and timeout management to prevent resource leaks
- Graceful handling of provider failures with automatic fallback

### 5. Streaming Support

Support for streaming responses with automatic fallback:

```go
// Stream streams responses token by token from the fastest or primary provider
func (mp *MultiProvider) Stream(ctx context.Context, prompt string, options ...domain.Option) (domain.ResponseStream, error) {
    // Implementation with fallback capabilities
}
```

## Benchmark Results

Benchmarks show that the `MultiProvider` adds minimal overhead while providing significant benefits:

- Fastest strategy successfully selects the quickest provider
- Primary strategy with fallback maintains availability even when primary provider fails
- Timeout management properly handles slow providers, preventing resource leaks

```
BenchmarkProviderTypes/SingleProvider/Generate                 50528744    23.57 ns/op
BenchmarkProviderTypes/MultiProvider_Fastest_Optimal/Generate        22    51048634 ns/op
BenchmarkProviderTypes/MultiProvider_Primary_Fallback/Generate   467302    2505 ns/op
```

## Usage Examples

Basic usage with fastest strategy:

```go
// Create provider weights
providers := []provider.ProviderWeight{
    {Provider: openAIProvider, Weight: 1.0, Name: "openai"},
    {Provider: anthropicProvider, Weight: 1.0, Name: "anthropic"},
}

// Create a multi-provider with the fastest selection strategy
multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

// Use like any regular provider
response, err := multiProvider.Generate(ctx, prompt)
```

Primary with fallback:

```go
// Create a multi-provider with primary strategy
multiProvider := provider.NewMultiProvider(providers, provider.StrategyPrimary).
    WithPrimaryProvider(0)  // First provider is primary
    
// Optional timeout configuration
multiProvider = multiProvider.WithTimeout(5 * time.Second)

// Generate with fallback capability
response, err := multiProvider.Generate(ctx, prompt)
```

## Primary Strategy Implementation Details

### Issue Fixed

The `MultiProvider` implementation had a non-deterministic behavior when using the primary strategy. The issue was due to how results were collected from concurrent execution:

1. Results from concurrent provider calls were collected in a non-deterministic order
2. The `primaryIdx` was used to index into this result array, assuming results were in the same order as providers
3. This led to inconsistent behavior in tests and real-world scenarios where the primary provider might not be selected correctly

### Solution Implemented

We implemented a sequential execution model specifically for the primary provider strategy:

1. **Sequential Execution for Primary**
   - Added `sequentialGenerateForPrimary`, `sequentialGenerateMessageForPrimary`, and `sequentialGenerateWithSchemaForPrimary` methods
   - These methods execute the primary provider first, then fall back to other providers if needed
   - This guarantees deterministic behavior for the primary strategy

2. **Implementation Approach**
   ```go
   // sequentialGenerateForPrimary runs Generate sequentially for the primary provider strategy
   func (mp *MultiProvider) sequentialGenerateForPrimary(ctx context.Context, prompt string, options []domain.Option) (string, error) {
       // Try the primary provider first
       primaryProvider := mp.providers[mp.primaryProvider]
       content, err := primaryProvider.Provider.Generate(ctx, prompt, options...)
       if err == nil {
           return content, nil
       }
       
       // If primary fails, try others sequentially
       for i, pw := range mp.providers {
           if i == mp.primaryProvider {
               continue // Skip the primary we already tried
           }
           content, err := pw.Provider.Generate(ctx, prompt, options...)
           if err == nil {
               return content, nil
           }
       }
       
       // All providers failed
       return "", ErrNoSuccessfulCalls
   }
   ```

3. **Hybrid Execution Model**
   - Primary strategy uses sequential execution for deterministic behavior
   - Fastest and Consensus strategies continue using concurrent execution for optimal performance

4. **Comprehensive Testing**
   - Added deterministic tests with call counters to verify correct provider selection
   - Tests verify that only the primary provider is called when it succeeds
   - Tests also verify fallback behavior when the primary provider fails

This approach ensures deterministic behavior for the primary strategy without sacrificing performance for other strategies.

## Future Enhancements

1. **Enhanced Consensus Strategy** - Further refinement of the consensus algorithms for better quality
2. **Dynamic Provider Selection** - Adjust provider selection based on historical performance
3. **Response Quality Metrics** - Add ability to score and compare responses from different providers
4. **Cost Optimization** - Implement strategies that balance cost and quality across different providers
5. **Response Caching** - Add caching for primary provider to further improve performance