package provider

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ConsensusStrategy defines how to reach consensus among multiple responses
type ConsensusStrategy int

const (
	// ConsensusMajority uses simple majority voting (most common response)
	ConsensusMajority ConsensusStrategy = iota
	// ConsensusSimilarity groups responses by similarity and chooses largest group
	ConsensusSimilarity
	// ConsensusWeighted considers provider weights
	ConsensusWeighted
)

// consensusConfig contains the configuration for consensus algorithms
type consensusConfig struct {
	// Strategy determines which consensus algorithm to use
	Strategy ConsensusStrategy
	// SimilarityThreshold is the minimum similarity score (0.0-1.0) for responses to be considered similar
	// Used only for ConsensusSimilarity strategy
	SimilarityThreshold float64
}

// defaultConsensusConfig returns the default consensus configuration
func defaultConsensusConfig() consensusConfig {
	return consensusConfig{
		Strategy:            ConsensusMajority,
		SimilarityThreshold: 0.7, // Default 70% similarity threshold
	}
}

// selectConsensusTextResult implements enhanced consensus strategies for text results
// This version uses the provider's consensus configuration
func selectConsensusTextResult(results []fallbackResult) (string, error) {
	// Count successful results to see if we have any
	// Pre-allocate with expected capacity
	successfulResults := make([]fallbackResult, 0, len(results))
	for _, result := range results {
		if result.err == nil && result.content != "" {
			successfulResults = append(successfulResults, result)
		}
	}

	if len(successfulResults) == 0 {
		return "", ErrNoSuccessfulCalls
	}

	// Fast path: If only one successful response, return it immediately
	if len(successfulResults) == 1 {
		return successfulResults[0].content, nil
	}

	// Get the current configuration from the MultiProvider
	// We have to use the global config from multi.go since this function
	// doesn't have direct access to the MultiProvider instance
	config := globalConsensusConfig
	if config == nil {
		defaultConfig := defaultConsensusConfig()
		config = &defaultConfig
	}

	// Apply the selected consensus strategy
	switch config.Strategy {
	case ConsensusMajority:
		return selectMajorityConsensus(successfulResults)
	case ConsensusSimilarity:
		return selectSimilarityConsensus(successfulResults, config.SimilarityThreshold)
	case ConsensusWeighted:
		return selectWeightedConsensus(successfulResults)
	default:
		// Default to majority if unknown strategy
		return selectMajorityConsensus(successfulResults)
	}
}

// Global variable to store the current consensus configuration
// This is set by the MultiProvider when it runs operations
var globalConsensusConfig *consensusConfig

// selectMajorityConsensus implements majority voting consensus
// Returns the most common response among all successful responses
func selectMajorityConsensus(results []fallbackResult) (string, error) {
	// Count occurrences of each unique response
	responseCount := make(map[string]int)
	for _, result := range results {
		responseCount[result.content]++
	}

	// Find the most common response
	var mostCommonResponse string
	maxCount := 0

	for response, count := range responseCount {
		if count > maxCount {
			maxCount = count
			mostCommonResponse = response
		}
	}

	return mostCommonResponse, nil
}

// groupCache is used to cache group memberships for similar content
// This avoids recalculating similarity scores for repeated executions
type groupCache struct {
	groups               map[string]int // maps content hash to group index
	groupRepresentatives []string       // representative content for each group
	similarityThreshold  float64        // the threshold used for this cache
	mu                   sync.RWMutex
}

// global group cache (will be reset when threshold changes)
var globalGroupCache = &groupCache{
	groups:              make(map[string]int, 64),
	similarityThreshold: 0.0, // invalid default, will be updated on first use
}

// contentHash creates a simple hash for a content string to use as cache key
func contentHash(content string) string {
	// For simplicity, we'll use first 64 chars + length as a reasonable uniqueness approximation
	prefix := content
	if len(prefix) > 64 {
		prefix = prefix[:64]
	}
	return fmt.Sprintf("%s#%d", prefix, len(content))
}

// resetIfThresholdChanged resets the cache if the similarity threshold has changed
func (c *groupCache) resetIfThresholdChanged(threshold float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.similarityThreshold != threshold {
		// Reset cache when threshold changes
		c.groups = make(map[string]int, 64)
		c.groupRepresentatives = nil
		c.similarityThreshold = threshold
	}
}

// getGroup returns the cached group index for a content string, or -1 if not found
func (c *groupCache) getGroup(content string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hash := contentHash(content)
	groupIndex, found := c.groups[hash]
	if !found {
		return -1
	}
	return groupIndex
}

// addToGroup adds a content string to a group
func (c *groupCache) addToGroup(content string, groupIndex int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := contentHash(content)
	c.groups[hash] = groupIndex
}

// addNewGroup creates a new group with the given content as representative
func (c *groupCache) addNewGroup(content string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	newGroupIndex := len(c.groupRepresentatives)
	c.groupRepresentatives = append(c.groupRepresentatives, content)

	hash := contentHash(content)
	c.groups[hash] = newGroupIndex

	return newGroupIndex
}

// getGroupRepresentative returns the representative content for a group
func (c *groupCache) getGroupRepresentative(groupIndex int) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if groupIndex < 0 || groupIndex >= len(c.groupRepresentatives) {
		return "" // Invalid group index
	}
	return c.groupRepresentatives[groupIndex]
}

// selectSimilarityConsensus implements similarity-based consensus
// Groups responses by similarity and returns the most common group
// This version is optimized with caching and more efficient group comparisons
func selectSimilarityConsensus(results []fallbackResult, similarityThreshold float64) (string, error) {
	// If we only have a few results, fall back to majority consensus
	if len(results) <= 2 {
		return selectMajorityConsensus(results)
	}

	// Reset the group cache if the threshold has changed
	globalGroupCache.resetIfThresholdChanged(similarityThreshold)

	// Preallocate groups array with estimated capacity
	groups := make([][]fallbackResult, 0, min(len(results), 5))

	// Track group counts for faster largest group lookup
	groupCounts := make([]int, 0, 5)

	// Start with the first result as its own group
	firstResult := results[0]
	groups = append(groups, []fallbackResult{firstResult})
	groupCounts = append(groupCounts, 1)
	globalGroupCache.addNewGroup(firstResult.content)

	// Track the largest group
	largestGroupIndex := 0

	// For each remaining result, check if it should join an existing group
	for i := 1; i < len(results); i++ {
		result := results[i]
		foundGroup := false

		// Check if this content is already in the cache
		groupIndex := globalGroupCache.getGroup(result.content)
		if groupIndex >= 0 && groupIndex < len(groups) {
			// Add to the cached group
			groups[groupIndex] = append(groups[groupIndex], result)
			groupCounts[groupIndex]++
			foundGroup = true

			// Update largest group if needed
			if groupCounts[groupIndex] > groupCounts[largestGroupIndex] {
				largestGroupIndex = groupIndex
			}
			continue
		}

		// Try to add to an existing group by comparing with representatives
		for j := range groups {
			// Get the representative content from cache for comparison
			representative := globalGroupCache.getGroupRepresentative(j)

			// Compare with the representative
			similarity := calculateSimilarity(result.content, representative)
			if similarity >= similarityThreshold {
				groups[j] = append(groups[j], result)
				groupCounts[j]++
				globalGroupCache.addToGroup(result.content, j)
				foundGroup = true

				// Update largest group if needed
				if groupCounts[j] > groupCounts[largestGroupIndex] {
					largestGroupIndex = j
				}
				break
			}
		}

		// If no matching group, create a new one
		if !foundGroup {
			newGroupIndex := len(groups)
			groups = append(groups, []fallbackResult{result})
			groupCounts = append(groupCounts, 1)
			globalGroupCache.addNewGroup(result.content)

			// Check if this is now the largest group (unlikely but possible with singleton groups)
			if groupCounts[newGroupIndex] > groupCounts[largestGroupIndex] {
				largestGroupIndex = newGroupIndex
			}
		}
	}

	// Use the largest group
	largestGroup := groups[largestGroupIndex]

	// If we have a group, return the "best" response from it
	if len(largestGroup) > 0 {
		return selectBestFromGroup(largestGroup), nil
	}

	// Fallback to first result if something went wrong
	return results[0].content, nil
}

// selectWeightedConsensus implements weight-based consensus
// Gives more influence to providers with higher weights
// This version is optimized for performance and also considers similarity for better grouping
func selectWeightedConsensus(results []fallbackResult) (string, error) {
	// If we have less than 2 results, no need for consensus
	if len(results) <= 1 {
		if len(results) == 1 {
			return results[0].content, nil
		}
		return "", ErrNoSuccessfulCalls
	}

	// If no weights are present, fall back to majority voting
	hasWeights := false
	for _, result := range results {
		if result.weight > 0 {
			hasWeights = true
			break
		}
	}

	if !hasWeights {
		return selectMajorityConsensus(results)
	}

	// Enhanced weighted consensus considering both exact matches and similar responses
	// First group exactly matching responses (faster than similarity)

	// Group identical responses - pre-allocate with estimated capacity
	type weightedResponse struct {
		content       string
		totalWeight   float64
		resultCount   int
		providerNames []string // Track all providers for debugging
		elapsedTime   float64  // Average elapsed time in milliseconds
		similarTo     []string // For similar response tracking (used in second phase)
	}

	// Pre-allocate with estimated capacity (to avoid map resizing)
	estimatedUnique := min(len(results), 5) // Usually not more than 5 unique responses
	weightedResponses := make(map[string]*weightedResponse, estimatedUnique)

	totalValidWeight := 0.0

	// First pass: group identical responses (exact match)
	for _, result := range results {
		if result.err != nil {
			continue // Skip failed results
		}

		// Skip empty responses
		if result.content == "" {
			continue
		}

		// Default weight to 1.0 if not specified
		weight := result.weight
		if weight <= 0 {
			weight = 1.0
		}

		totalValidWeight += weight

		// Check for exact match in map
		if wr, exists := weightedResponses[result.content]; exists {
			wr.totalWeight += weight
			wr.resultCount++
			wr.providerNames = append(wr.providerNames, result.provider)

			// Update average elapsed time
			resultElapsedMs := float64(result.elapsedTime.Milliseconds())
			wr.elapsedTime = (wr.elapsedTime*float64(wr.resultCount-1) + resultElapsedMs) / float64(wr.resultCount)
		} else {
			weightedResponses[result.content] = &weightedResponse{
				content:       result.content,
				totalWeight:   weight,
				resultCount:   1,
				providerNames: []string{result.provider},
				elapsedTime:   float64(result.elapsedTime.Milliseconds()),
				similarTo:     make([]string, 0, 3), // Pre-allocate for potential similar responses
			}
		}
	}

	// If we collected no valid responses, return error
	if len(weightedResponses) == 0 {
		return "", ErrNoSuccessfulCalls
	}

	// If we only have one unique response, return it immediately
	if len(weightedResponses) == 1 {
		for _, wr := range weightedResponses {
			return wr.content, nil
		}
	}

	// Second phase: Consider similar responses for enhanced grouping
	// Only do similarity analysis if we have multiple distinct responses
	// and the top response doesn't have overwhelming weight

	// First, convert map to slice for processing
	responses := make([]*weightedResponse, 0, len(weightedResponses))
	for _, wr := range weightedResponses {
		responses = append(responses, wr)
	}

	// Sort by weight for quick access to top response
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].totalWeight > responses[j].totalWeight
	})

	// Check if top response has overwhelming weight (over 70%)
	topResponse := responses[0]
	if topResponse.totalWeight > 0.7*totalValidWeight {
		// One response has overwhelming support, no need for similarity analysis
		return topResponse.content, nil
	}

	// Calculate similarity between responses and combine weights for similar responses
	// Using a slightly lower threshold (0.7) for similarity than the default consensus
	const similarityThreshold = 0.65

	// For each pair of responses, check similarity and potentially combine weights
	for i := 0; i < len(responses); i++ {
		for j := i + 1; j < len(responses); j++ {
			// Skip already processed responses
			if len(responses[j].similarTo) > 0 {
				continue
			}

			similarity := calculateSimilarity(responses[i].content, responses[j].content)
			if similarity >= similarityThreshold {
				// Mark as similar
				responses[i].similarTo = append(responses[i].similarTo, responses[j].content)
				responses[j].similarTo = append(responses[j].similarTo, responses[i].content)

				// Combine weights (add j's weight to i)
				responses[i].totalWeight += responses[j].totalWeight * similarity
			}
		}
	}

	// Resort after combining weights
	sort.Slice(responses, func(i, j int) bool {
		// If weights are very close, use result count as tiebreaker
		if abs(responses[i].totalWeight-responses[j].totalWeight) < 0.1 {
			if responses[i].resultCount == responses[j].resultCount {
				// If counts are also equal, use response time as final tiebreaker
				return responses[i].elapsedTime < responses[j].elapsedTime
			}
			return responses[i].resultCount > responses[j].resultCount
		}
		return responses[i].totalWeight > responses[j].totalWeight
	})

	// Return the response with the highest adjusted weight
	return responses[0].content, nil
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// selectBestFromGroup selects the best response from a group of similar responses
// Currently returns the shortest non-empty response as a simple heuristic
// (assuming shorter responses are more concise and focused)
func selectBestFromGroup(group []fallbackResult) string {
	if len(group) == 0 {
		return ""
	}

	if len(group) == 1 {
		return group[0].content
	}

	// Sort by content length (ascending) to find the shortest non-empty response
	sort.Slice(group, func(i, j int) bool {
		// Skip empty responses
		if len(group[i].content) == 0 {
			return false
		}
		if len(group[j].content) == 0 {
			return true
		}
		return len(group[i].content) < len(group[j].content)
	})

	// Return the first non-empty response
	for _, result := range group {
		if len(result.content) > 0 {
			return result.content
		}
	}

	// If all are empty, return the first response
	return group[0].content
}

// similarityCache implements a simple LRU-like cache for similarity calculations
// to avoid recalculating similarity for the same string pairs repeatedly
type similarityCache struct {
	cache      map[string]float64
	maxEntries int
	mu         sync.RWMutex
}

// global similarity cache with a reasonable size limit
var globalSimilarityCache = &similarityCache{
	cache:      make(map[string]float64, 128),
	maxEntries: 128, // Adjust based on expected number of unique comparisons
}

// GetCacheKey creates a deterministic key for two strings regardless of order
// This is exported for testing purposes
func GetCacheKey(a, b string) string {
	// Ensure consistent order by sorting lexicographically
	if a > b {
		a, b = b, a
	}
	return a + "||" + b
}

// getCacheKey is an alias for GetCacheKey (for backward compatibility)
func getCacheKey(a, b string) string {
	return GetCacheKey(a, b)
}

// Get retrieves a cached similarity score if available
func (c *similarityCache) Get(a, b string) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := getCacheKey(a, b)
	score, found := c.cache[key]
	return score, found
}

// Set stores a similarity score in the cache
func (c *similarityCache) Set(a, b string, score float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := getCacheKey(a, b)

	// Simple cache eviction if we exceed max entries
	if len(c.cache) >= c.maxEntries {
		// Remove a random entry when the cache is full
		// This is simpler than a true LRU implementation but works well for this use case
		for k := range c.cache {
			delete(c.cache, k)
			break
		}
	}

	c.cache[key] = score
}

// calculateSimilarity computes a basic similarity score between two strings
// Returns a value between 0.0 (completely different) and 1.0 (identical)
// This version is optimized with caching and fast paths for common cases
func calculateSimilarity(a, b string) float64 {
	// Fast paths for common cases
	// Simple case: exact match
	if a == b {
		return 1.0
	}

	// Very short strings can be handled directly
	if len(a) < 5 || len(b) < 5 {
		// For very short strings, do direct comparison to avoid overhead
		aLower := strings.ToLower(a)
		bLower := strings.ToLower(b)
		if aLower == bLower {
			return 1.0
		}
	}

	// Check cache first - cache lookup is fast so we can do it early
	if score, found := globalSimilarityCache.Get(a, b); found {
		return score
	}

	// Convert to lowercase for more forgiving comparison
	aLower := strings.ToLower(a)
	bLower := strings.ToLower(b)

	// If still exact match after lowercase conversion
	if aLower == bLower {
		score := 1.0
		globalSimilarityCache.Set(a, b, score)
		return score
	}

	// Length-based optimization - if the length difference is too great,
	// the strings are likely very different
	lenA, lenB := len(aLower), len(bLower)
	if float64(min(lenA, lenB))/float64(max(lenA, lenB)) < 0.5 {
		// Lengths differ by more than 50%, likely very different
		score := 0.3
		globalSimilarityCache.Set(a, b, score)
		return score
	}

	// Pre-allocate word sets with estimated capacity
	estimatedWords := (len(aLower) + len(bLower)) / 10 // rough estimate: 1 word per 10 chars
	if estimatedWords < 10 {
		estimatedWords = 10
	}

	// Implement an optimized Jaccard similarity based on word overlap
	// Pre-allocate slices to reduce allocations
	aWords := strings.Fields(aLower)
	bWords := strings.Fields(bLower)

	// Create sets of words with pre-allocation
	aSet := make(map[string]bool, estimatedWords)
	for _, word := range aWords {
		// Filter out very common words that don't contribute much to meaning
		if len(word) > 2 { // Skip very short words
			aSet[word] = true
		}
	}

	// Quick check after stopword filtering
	if len(aSet) == 0 {
		// No meaningful words, treat as low similarity
		score := 0.2
		globalSimilarityCache.Set(a, b, score)
		return score
	}

	bSet := make(map[string]bool, estimatedWords)
	intersection := 0

	// Build bSet and count intersection in one pass
	for _, word := range bWords {
		// Skip very short words
		if len(word) <= 2 {
			continue
		}

		// Add to set if not already there
		if !bSet[word] {
			bSet[word] = true

			// Check if in both sets
			if aSet[word] {
				intersection++
			}
		}
	}

	// Quick check after stopword filtering
	if len(bSet) == 0 {
		// No meaningful words, treat as low similarity
		score := 0.2
		globalSimilarityCache.Set(a, b, score)
		return score
	}

	// Find union size
	union := len(aSet) + len(bSet) - intersection

	// Calculate Jaccard similarity
	if union == 0 {
		return 0.0
	}

	// Store in cache
	score := float64(intersection) / float64(union)
	globalSimilarityCache.Set(a, b, score)
	return score
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
