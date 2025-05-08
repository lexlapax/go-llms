# Multi-Provider Guide

> **[Documentation Home](/REFERENCE.md) / [User Guide](/docs/user-guide/) / Multi-Provider Guide**

This guide explains how to use the Multi-Provider functionality in Go-LLMs to work with multiple LLM providers simultaneously.

*Related: [Getting Started](getting-started.md) | [Error Handling](error-handling.md) | [Architecture](/docs/technical/architecture.md#multi-provider-strategies) | [API Reference](/docs/api/README.md)*

## Table of Contents

1. [Overview](#overview)
2. [Provider Strategies](#provider-strategies)
3. [Implementation Details](#implementation-details)
4. [Usage Examples](#usage-examples)
5. [Performance Considerations](#performance-considerations)
6. [Future Enhancements](#future-enhancements)

## Overview

The `MultiProvider` implementation enables concurrent processing across multiple LLM providers. It addresses several key use cases:

1. **Performance Optimization** - Get responses from the fastest provider 
2. **High Availability** - Implement fallback mechanisms when primary providers fail
3. **Response Quality** - Implement consensus mechanisms across multiple providers
4. **A/B Testing** - Compare responses from different providers for evaluation

## Provider Strategies

Three main selection strategies are implemented:

### 1. Fastest Strategy

The Fastest strategy sends the request to all providers concurrently and returns the first successful response. This is useful when minimizing latency is the primary goal.

```go
// Create a multi-provider with the fastest strategy
multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

// Generate a response - will return from the quickest provider
response, err := multiProvider.Generate(ctx, prompt)
```

### 2. Primary Strategy

The Primary strategy tries a designated primary provider first, falling back to others if it fails. This ensures consistent responses while maintaining high availability.

```go
// Create a multi-provider with primary strategy
multiProvider := provider.NewMultiProvider(providers, provider.StrategyPrimary).
    WithPrimaryProvider(0)  // Use first provider as primary

// Generate with fallback capability
response, err := multiProvider.Generate(ctx, prompt)
```

The Primary strategy uses sequential execution for deterministic behavior:
- It tries the primary provider first
- If that fails, it tries each fallback provider in order
- This guarantees deterministic behavior and maximum consistency

### 3. Consensus Strategy

The Consensus strategy aggregates responses from multiple providers to determine the best response. This is useful for improving response quality and reliability.

```go
// Create a multi-provider with consensus strategy
consensusProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus).
    WithConsensusStrategy(provider.ConsensusSimilarity).
    WithSimilarityThreshold(0.7)

// Generate with consensus from multiple providers
response, err := consensusProvider.Generate(ctx, prompt)
```

The Consensus strategy includes multiple consensus algorithms:
- **ConsensusMajority** - Simple majority voting
- **ConsensusSimilarity** - Groups responses by similarity and chooses largest group
- **ConsensusWeighted** - Considers provider weights when determining consensus

## Implementation Details

### Provider Architecture

The implementation consists of these key components:

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

// ProviderWeight defines the weight of a provider in a multi-provider setup
type ProviderWeight struct {
    Provider domain.Provider
    Weight   float64 // 0.0 to 1.0, default 1.0
    Name     string  // Optional name for the provider
}
```

### Hybrid Execution Model

The implementation uses different execution models depending on the strategy:
- **Concurrent execution** for Fastest and Consensus strategies (for better performance)
- **Sequential execution** for Primary strategy (for deterministic behavior)

### Result Selection

Each strategy implements a different approach to selecting the final result:

- **Fastest Strategy**: Returns the result from the first provider to respond successfully
- **Primary Strategy**: Returns the primary provider's result, or falls back to others in order
- **Consensus Strategy**: Analyzes all responses to determine the best consensus answer

### Comprehensive Error Handling

The MultiProvider implements sophisticated error handling:
- Detailed error aggregation when all providers fail
- Context and timeout management to prevent resource leaks
- Provider-specific error attribution in multi-provider errors

### Streaming Support

All strategies support streaming responses:

```go
// Stream responses with multi-provider
stream, err := multiProvider.Stream(ctx, prompt)
if err != nil {
    return err
}

for token := range stream {
    fmt.Print(token.Text)
    if token.Finished {
        break
    }
}
```

## Usage Examples

### Basic Setup

```go
// Create provider weights
providers := []provider.ProviderWeight{
    {Provider: openAIProvider, Weight: 1.0, Name: "openai"},
    {Provider: anthropicProvider, Weight: 1.0, Name: "anthropic"},
}

// Create a multi-provider with the fastest selection strategy
multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)

// Optional timeout configuration
multiProvider = multiProvider.WithTimeout(5 * time.Second)

// Use like any regular provider
response, err := multiProvider.Generate(ctx, prompt)
```

### With Mock Providers

For testing or development without API keys:

```go
// Create mock providers with different behaviors
fastMock := provider.NewMockProvider().
    WithDelay(100 * time.Millisecond).
    WithResponse("This is a response from the fast provider")

slowMock := provider.NewMockProvider().
    WithDelay(500 * time.Millisecond).
    WithResponse("This is a response from the slow provider")

// Create provider weights
providers := []provider.ProviderWeight{
    {Provider: fastMock, Weight: 0.5, Name: "fast"},
    {Provider: slowMock, Weight: 1.0, Name: "slow"},
}

// Create a multi-provider
multiProvider := provider.NewMultiProvider(providers, provider.StrategyFastest)
```

### Consensus Strategy Example

```go
// Create a multi-provider with the consensus strategy
multiProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus).
    WithConsensusStrategy(provider.ConsensusSimilarity).
    WithSimilarityThreshold(0.7)

// Generate a response using consensus from all providers
response, err := multiProvider.Generate(ctx, prompt)

// You can also specify a different consensus algorithm
multiProvider = multiProvider.WithConsensusStrategy(provider.ConsensusWeighted)
```

## Performance Considerations

Benchmarks show that the MultiProvider adds minimal overhead:

- **Fastest Strategy**: Adds ~5ms overhead but can significantly reduce overall latency by using the quickest provider
- **Primary Strategy**: Near-zero overhead when primary succeeds, small overhead when fallback is needed
- **Consensus Strategy**: Higher overhead as it waits for multiple responses, but can improve response quality

### Timeout Management

Always set appropriate timeouts to prevent resource leaks:

```go
// Set a default timeout for all providers
multiProvider = multiProvider.WithTimeout(5 * time.Second)

// Or use a context with timeout for individual requests
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

response, err := multiProvider.Generate(ctx, prompt)
```

## Enhanced Consensus Algorithms

The MultiProvider includes several optimized consensus algorithms:

### 1. Similarity-Based Consensus

This algorithm groups responses by semantic similarity:
- Calculates similarity between all pairs of responses
- Groups responses that are above the similarity threshold
- Returns the most commonly occurring response group

Optimizations include:
- Similarity score caching
- Group membership caching
- Fast paths for identical responses
- Pre-allocation to reduce memory pressure

### 2. Weighted Consensus

This algorithm factors in provider weights when determining consensus:
- Gives more influence to higher-weighted providers
- Combines weight with similarity for better results
- Includes fast paths for overwhelming consensus

## Future Enhancements

Potential future enhancements to the MultiProvider include:

1. **Enhanced Consensus Strategy** - Further refinement of the consensus algorithms for better quality
2. **Dynamic Provider Selection** - Adjust provider selection based on historical performance
3. **Response Quality Metrics** - Add ability to score and compare responses from different providers
4. **Cost Optimization** - Implement strategies that balance cost and quality across different providers
5. **Response Caching** - Add caching for primary provider to further improve performance