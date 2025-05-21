package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

// ErrCacheMiss is a custom error that can be used if os.ErrNotExist is not specific enough,
// but for this implementation, we'll rely on os.ErrNotExist for simplicity.
// var ErrCacheMiss = errors.New("cache file not found or is invalid")

// SaveInventory saves the model inventory to a cache file.
// The inventory parameter is of type *domain.CachedModelInventory, which includes the FetchedAt timestamp.
func SaveInventory(inventory *domain.CachedModelInventory, cachePath string) error {
	if inventory == nil {
		return os.ErrInvalid // Or a custom error like errors.New("cannot save nil inventory")
	}

	// Ensure the directory for cachePath exists.
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err // Return error from MkdirAll
	}

	// Marshal the inventory into JSON format.
	data, err := json.MarshalIndent(inventory, "", "  ") // Using MarshalIndent for readability
	if err != nil {
		return err // Return error from json.MarshalIndent
	}

	// Write the JSON data to the file specified by cachePath.
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return err // Return error from os.WriteFile
	}

	return nil
}

// LoadInventory loads the model inventory from a cache file.
// It returns a *domain.CachedModelInventory struct.
func LoadInventory(cachePath string) (*domain.CachedModelInventory, error) {
	// Check if the file at cachePath exists.
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil, os.ErrNotExist // Return os.ErrNotExist directly
	}

	// Read the content of the file.
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err // Return error from os.ReadFile
	}

	// Unmarshal the JSON data into a domain.CachedModelInventory struct.
	var inventory domain.CachedModelInventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		return nil, err // Return error from json.Unmarshal
	}

	return &inventory, nil
}

// IsCacheValid checks if the cached inventory is still valid based on maxCacheAge.
// cachedInventory is a *domain.CachedModelInventory, which includes the FetchedAt timestamp.
func IsCacheValid(cachedInventory *domain.CachedModelInventory, maxCacheAge time.Duration) bool {
	if cachedInventory == nil {
		return false // Nil inventory is not valid.
	}

	// Calculate if cachedInventory.FetchedAt is within the maxCacheAge from the current time.
	// If FetchedAt + maxCacheAge is after Now, it's valid.
	// Equivalent to: time.Now().Sub(cachedInventory.FetchedAt) <= maxCacheAge
	return time.Since(cachedInventory.FetchedAt) <= maxCacheAge
}
