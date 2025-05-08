# Multi-Provider Consensus Strategy Example

This example demonstrates the different consensus strategies available in Go-LLMs for using multiple LLM providers together and finding agreement among their responses.

## Overview

The consensus strategy in Go-LLMs allows combining multiple LLM providers and determining the "best" response through various methods of finding agreement. This is particularly useful for:

1. Improving response quality by filtering out outliers or hallucinations
2. Increasing accuracy for factual questions
3. Handling cases where providers might disagree

## Available Consensus Strategies

The example demonstrates three different consensus strategies:

1. **Majority Voting** (`ConsensusMajority`): Returns the most common response
2. **Similarity-Based** (`ConsensusSimilarity`): Groups responses by similarity and returns the largest group
3. **Weighted Consensus** (`ConsensusWeighted`): Considers provider weights in addition to response similarity

## Running the Example

Build and run the example:

```bash
# Build the example
make example EXAMPLE=consensus

# Run the example
./bin/consensus
```

For best results, provide API keys for multiple providers:

```bash
# Set API keys
export OPENAI_API_KEY=your_openai_key
export ANTHROPIC_API_KEY=your_anthropic_key

# Run the example
./bin/consensus
```

If no API keys are provided, the example will use mock providers with simulated responses.

## Example Structure

The example demonstrates:

1. **Basic Consensus Usage**: Simple demonstration of consensus with a factual question
2. **Different Consensus Strategies**: Comparison of the three consensus strategies
3. **Handling Contradictions**: How consensus can handle contradictory information
4. **Weighted Consensus**: How to give different weights to different providers

## Implementation Details

### Creating a Multi-Provider with Consensus

```go
// Create provider weights
providers := []provider.ProviderWeight{
    {Provider: openAIProvider, Weight: 1.0, Name: "openai"},
    {Provider: anthropicProvider, Weight: 1.0, Name: "anthropic"},
}

// Create a multi-provider with the consensus strategy
consensusProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus)
```

### Configuring Consensus Strategy

```go
// Use similarity-based consensus with a 70% similarity threshold
similarityProvider := provider.NewMultiProvider(providers, provider.StrategyConsensus).
    WithConsensusStrategy(provider.ConsensusSimilarity).
    WithSimilarityThreshold(0.7)
```

### Weighting Providers

```go
// Give higher weight to a specific provider
weightedProviders := []provider.ProviderWeight{
    {Provider: openAIProvider, Weight: 2.0, Name: "openai"}, // Double weight
    {Provider: anthropicProvider, Weight: 1.0, Name: "anthropic"},
}

// Create a weighted consensus provider
weightedConsensusProvider := provider.NewMultiProvider(weightedProviders, provider.StrategyConsensus).
    WithConsensusStrategy(provider.ConsensusWeighted)
```

## Testing

The example includes tests demonstrating the behavior of different consensus strategies:

```bash
# Run tests for the consensus example
make test-examples EXAMPLE=consensus
```