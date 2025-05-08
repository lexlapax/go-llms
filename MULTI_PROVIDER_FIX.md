# MultiProvider Primary Strategy Fix

## Problem

The current implementation of the `StrategyPrimary` in `MultiProvider` has a design issue that can lead to inconsistent behavior:

1. The results are collected from a channel in no particular order.
2. The primary provider selection is done by index, not by provider identity.
3. This means `results[primaryIdx]` might not actually be from the provider at `providers[primaryIdx]`.

## Proposed Solution

There are two potential fixes:

### Option 1: Maintain Result Provider Mapping

Include the provider index in the `fallbackResult` struct to maintain the mapping:

```go
type fallbackResult struct {
    providerName    string
    providerIndex   int    // Add this field
    content         string
    response        domain.Response
    structured      interface{}
    err             error
    elapsedTime     time.Duration
    weight          float64
}
```

Then when launching goroutines:

```go
go func(idx int, providerWeight ProviderWeight) {
    // ... existing code ...
    
    // Send result with provider index
    resultCh <- fallbackResult{
        provider:      providerName,
        providerIndex: idx,  // Include the index
        content:       content,
        err:           err,
        elapsedTime:   elapsed,
        weight:        providerWeight.Weight,
    }
}(i, pw)
```

Finally, when selecting results:

```go
case StrategyPrimary:
    // Try primary provider first, then fall back
    primaryIdx := mp.primaryProvider
    
    // Find the result from the primary provider by index
    var primaryResult *fallbackResult
    for i := range results {
        if results[i].providerIndex == primaryIdx && results[i].err == nil {
            primaryResult = &results[i]
            break
        }
    }
    
    // If found a valid result from primary provider, use it
    if primaryResult != nil {
        return primaryResult.content, nil
    }
    
    // Primary failed or not found, try others
    for _, result := range results {
        if result.err == nil {
            return result.content, nil
        }
    }
```

### Option 2: Sequential Primary Provider Execution

Another approach is to execute the primary provider sequentially first, and only execute other providers if it fails:

```go
func (mp *MultiProvider) Generate(ctx context.Context, prompt string, options ...domain.Option) (string, error) {
    if len(mp.providers) == 0 {
        return "", ErrNoProviders
    }

    // Apply the configured timeout
    ctx, cancel := applyTimeoutFromContext(ctx, mp.defaultTimeout)
    defer cancel()

    if mp.selectionStrat == StrategyPrimary {
        // For primary strategy, try the primary provider first
        primaryIdx := mp.primaryProvider
        if primaryIdx >= 0 && primaryIdx < len(mp.providers) {
            // Try the primary provider
            primaryResult, err := mp.providers[primaryIdx].Provider.Generate(ctx, prompt, options...)
            if err == nil {
                // Primary succeeded, return its result
                return primaryResult, nil
            }
            
            // Primary failed, fall back to concurrent execution of the rest
            // First collect the non-primary providers
            remainingProviders := make([]ProviderWeight, 0, len(mp.providers)-1)
            for i, p := range mp.providers {
                if i != primaryIdx {
                    remainingProviders = append(remainingProviders, p)
                }
            }
            
            // Execute remaining providers concurrently
            results := mp.concurrentGenerateWithProviders(ctx, remainingProviders, prompt, options)
            
            // Select best result from remaining providers
            for _, result := range results {
                if result.err == nil {
                    return result.content, nil
                }
            }
            
            // All providers failed
            return "", ErrNoSuccessfulCalls
        }
    }

    // For other strategies, continue with concurrent execution
    results := mp.concurrentGenerate(ctx, prompt, options)
    
    // Apply selection strategy as before
    return mp.selectTextResult(results)
}
```

This approach ensures that the primary provider is always tried first, and its results are used if available, making the behavior more deterministic.

## Recommendation

Option 2 is recommended because:
1. It's more efficient - if the primary provider succeeds, we don't need to wait for other providers
2. It's more deterministic - we're explicitly trying the primary provider first
3. It's clearer in intent - the behavior matches the name "primary provider"

However, a hybrid approach may be best - implement the provider index tracking from Option 1 for result selection, while maintaining the concurrent execution for better performance in most cases.