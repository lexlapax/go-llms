package llmutil

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lexlapax/go-llms/pkg/modelinfo/cache"
	"github.com/lexlapax/go-llms/pkg/modelinfo/domain"
	"github.com/lexlapax/go-llms/pkg/modelinfo" // Corrected import
)

const (
	defaultCacheDirPerm = 0755
	defaultCacheFile    = "model_inventory.json"
	defaultMaxCacheAge  = 24 * time.Hour
	defaultSubDir       = "go-llms"
)

// GetAvailableModelsOptions specifies options for the GetAvailableModels function.
type GetAvailableModelsOptions struct {
	// UseCache enables or disables the caching mechanism.
	// Defaults to true.
	UseCache bool

	// CachePath specifies the full file path for the cache file.
	// If empty, a default path under the user's cache directory is used
	// (e.g., ~/.cache/go-llms/model_inventory.json on Linux).
	CachePath string

	// MaxCacheAge defines the maximum age for a cache entry to be considered valid.
	// If zero when passed in opts, a default of 24 hours is used.
	MaxCacheAge time.Duration
}

// GetAvailableModels fetches an aggregated inventory of available LLM models
// from various providers. It supports caching to reduce redundant data fetching.
//
// Parameters:
//   opts: An optional *GetAvailableModelsOptions struct to customize behavior.
//         If nil, default options are used (caching enabled, default path and age).
//
// Returns:
//   A *domain.ModelInventory containing the aggregated list of models and metadata.
//   An error if fetching or processing fails and cache is unavailable or invalid.
//
// Caching:
//   - If opts.UseCache is true (default behavior if opts is nil or opts.UseCache is not explicitly false),
//     the function first attempts to load a valid (non-expired) inventory from the cache path.
//   - The cache path is determined by opts.CachePath if provided. Otherwise, a default path
//     is constructed, typically under os.UserCacheDir()/go-llms/model_inventory.json.
//     If os.UserCacheDir() fails, it falls back to a local ".cache/go-llms/model_inventory.json".
//   - The maximum cache age is determined by opts.MaxCacheAge if provided and non-zero.
//     Otherwise, a default of 24 hours (defaultMaxCacheAge) is used.
//   - If fresh data is fetched successfully and caching is enabled, the new inventory
//     is saved to the cache.
func GetAvailableModels(opts *GetAvailableModelsOptions) (*domain.ModelInventory, error) {
	// 1. Handle Options and Defaults
	options := GetAvailableModelsOptions{
		UseCache:    true, // Default to using cache
		MaxCacheAge: defaultMaxCacheAge,
	}

	if opts != nil {
		// Only set UseCache to false if opts explicitly sets it to false.
		// If opts.UseCache is true (its zero value when opts is not nil but UseCache isn't set),
		// it will remain true due to the above default initialization.
		// A more explicit way: if opts contains a field indicating UseCache was set, use its value.
		// For this implementation, we assume if opts is passed, and UseCache is false, it's intentional.
		if !opts.UseCache && opts.UseCache == false { // Check if UseCache is explicitly set to false
			options.UseCache = false
		}

		if opts.CachePath != "" {
			options.CachePath = opts.CachePath
		}
		if opts.MaxCacheAge != 0 {
			options.MaxCacheAge = opts.MaxCacheAge
		}
	}

	if options.CachePath == "" {
		userCacheDir, err := os.UserCacheDir()
		if err != nil {
			// Fallback strategy: try to create a local .cache directory
			cwd, CwdErr := os.Getwd()
			if CwdErr != nil {
				// Very unlikely, but as a last resort, use a relative path from where the binary might be.
				// This might not be writable.
				userCacheDir = filepath.Join(".", ".cache")
				fmt.Printf("Warning: Could not get user cache directory (%v) or CWD (%v), using %s\n", err, CwdErr, userCacheDir)
			} else {
				userCacheDir = filepath.Join(cwd, ".cache")
				fmt.Printf("Warning: Could not get user cache directory: %v. Using local .cache dir: %s\n", err, userCacheDir)
			}
		}
		options.CachePath = filepath.Join(userCacheDir, defaultSubDir, defaultCacheFile)
	}

	// 2. Caching Logic
	if options.UseCache {
		loadedInventory, err := cache.LoadInventory(options.CachePath)
		if err == nil && loadedInventory != nil { // Cache hit
			if cache.IsCacheValid(loadedInventory, options.MaxCacheAge) {
				return &loadedInventory.Inventory, nil
			}
			// Cache is expired or invalid, proceed to fetch fresh data
			fmt.Printf("Cache found but expired or invalid at %s\n", options.CachePath)
		} else if err != os.ErrNotExist {
			// Log error if it's something other than cache file not existing
			fmt.Printf("Error loading cache from %s: %v. Fetching fresh data.\n", options.CachePath, err)
		}
		// If os.ErrNotExist, that's fine, we just need to fetch.
	}

	// 3. Data Fetching
	modelInfoService := modelinfo.NewModelInfoServiceFunc() // Use the func variable
	freshInventoryData, err := modelInfoService.AggregateModels()
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate model data: %w", err)
	}

	// If fetching was successful and caching is enabled, save the fresh data.
	if options.UseCache && freshInventoryData != nil {
		cachedDataToSave := domain.CachedModelInventory{
			Inventory: *freshInventoryData,
			FetchedAt: time.Now(),
		}
		if saveErr := cache.SaveInventory(&cachedDataToSave, options.CachePath); saveErr != nil {
			// Log error during saving but do not fail the whole operation.
			fmt.Printf("Warning: Failed to save fresh model inventory to cache at %s: %v\n", options.CachePath, saveErr)
		} else {
			fmt.Printf("Successfully saved fresh model inventory to cache at %s\n", options.CachePath)
		}
	}

	return freshInventoryData, nil
}
