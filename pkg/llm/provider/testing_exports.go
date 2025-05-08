package provider

import "time"

// This file contains exported functions for testing purposes only.
// These functions should not be used in production code.

// FallbackResult is a public version of fallbackResult for testing
type FallbackResult struct {
	Provider    string
	Content     string
	Err         error
	ElapsedTime time.Duration
	Weight      float64
}

// PublicCalculateSimilarity exposes calculateSimilarity for testing
func PublicCalculateSimilarity(a, b string) float64 {
	return calculateSimilarity(a, b)
}

// ResetSimilarityCache resets the global similarity cache for testing
func ResetSimilarityCache() {
	globalSimilarityCache = &similarityCache{
		cache:      make(map[string]float64, 128),
		maxEntries: 128,
	}
}

// GetSimilarityCacheEntries returns the current entries in the similarity cache (for testing)
func GetSimilarityCacheEntries() map[string]float64 {
	globalSimilarityCache.mu.RLock()
	defer globalSimilarityCache.mu.RUnlock()

	// Make a copy to avoid race conditions
	result := make(map[string]float64, len(globalSimilarityCache.cache))
	for k, v := range globalSimilarityCache.cache {
		result[k] = v
	}

	return result
}

// This has been moved to consensus.go for simplicity

// ResetGroupCache resets the global group cache for testing
func ResetGroupCache() {
	globalGroupCache = &groupCache{
		groups:              make(map[string]int, 64),
		similarityThreshold: 0.0, // invalid default, will be updated on first use
	}
}

// PublicSelectSimilarityConsensus exposes selectSimilarityConsensus for testing
func PublicSelectSimilarityConsensus(publicResults []FallbackResult, threshold float64) (string, error) {
	// Convert public results to internal results
	results := make([]fallbackResult, len(publicResults))
	for i, pr := range publicResults {
		results[i] = fallbackResult{
			provider:    pr.Provider,
			content:     pr.Content,
			err:         pr.Err,
			elapsedTime: pr.ElapsedTime,
			weight:      pr.Weight,
		}
	}

	return selectSimilarityConsensus(results, threshold)
}
