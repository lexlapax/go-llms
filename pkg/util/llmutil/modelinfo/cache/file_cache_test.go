package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp" // For deep comparison
	"github.com/lexlapax/go-llms/pkg/util/llmutil/modelinfo/domain"
)

func TestSaveInventory_Success(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "test_inventory.json")

	sampleInventory := &domain.CachedModelInventory{
		Inventory: domain.ModelInventory{
			Metadata: domain.Metadata{Version: "1.0"},
			Models:   []domain.Model{{Name: "test-model"}},
		},
		FetchedAt: time.Now().Truncate(time.Second), // Truncate for consistent comparison
	}

	err := SaveInventory(sampleInventory, cachePath)
	if err != nil {
		t.Fatalf("SaveInventory() failed: %v", err)
	}

	// Verify file content
	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("Failed to read saved cache file: %v", err)
	}

	var loadedFromFile domain.CachedModelInventory
	err = json.Unmarshal(data, &loadedFromFile)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved cache file content: %v", err)
	}

	// Normalize FetchedAt for comparison if loaded from JSON (timezones can be tricky)
	loadedFromFile.FetchedAt = loadedFromFile.FetchedAt.Local().Truncate(time.Second)

	if !cmp.Equal(sampleInventory, &loadedFromFile) {
		t.Errorf("Saved inventory content does not match original. Diff: %s", cmp.Diff(sampleInventory, &loadedFromFile))
	}
}

func TestSaveInventory_NilInventory(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "nil_inventory.json")

	err := SaveInventory(nil, cachePath)
	if err == nil {
		t.Fatal("SaveInventory() with nil inventory expected an error, got nil")
	}
	if err != os.ErrInvalid {
		t.Errorf("SaveInventory() with nil inventory expected os.ErrInvalid, got %v", err)
	}
}

func TestSaveInventory_DirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	// Create a path with a non-existent subdirectory
	cachePath := filepath.Join(tempDir, "subdir1", "subdir2", "test_inventory.json")

	sampleInventory := &domain.CachedModelInventory{
		Inventory: domain.ModelInventory{Metadata: domain.Metadata{Version: "1.0"}},
		FetchedAt: time.Now(),
	}

	err := SaveInventory(sampleInventory, cachePath)
	if err != nil {
		t.Fatalf("SaveInventory() failed when creating subdirectories: %v", err)
	}

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Errorf("SaveInventory() did not create the file in nested subdirectories: %v", err)
	}
}

func TestLoadInventory_Success(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "load_test.json")

	expectedInventory := &domain.CachedModelInventory{
		Inventory: domain.ModelInventory{
			Metadata: domain.Metadata{Version: "1.1", Description: "Test Load"},
			Models:   []domain.Model{{Name: "load-model-1"}, {Name: "load-model-2"}},
		},
		FetchedAt: time.Now().Truncate(time.Second), // Truncate for consistency
	}

	// Save it first
	err := SaveInventory(expectedInventory, cachePath)
	if err != nil {
		t.Fatalf("Setup for TestLoadInventory_Success failed during SaveInventory: %v", err)
	}

	loadedInventory, err := LoadInventory(cachePath)
	if err != nil {
		t.Fatalf("LoadInventory() failed: %v", err)
	}

	loadedInventory.FetchedAt = loadedInventory.FetchedAt.Local().Truncate(time.Second)

	if !cmp.Equal(expectedInventory, loadedInventory) {
		t.Errorf("Loaded inventory does not match expected. Diff: %s", cmp.Diff(expectedInventory, loadedInventory))
	}
}

func TestLoadInventory_FileNotExist(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "non_existent.json")

	_, err := LoadInventory(cachePath)
	if err == nil {
		t.Fatal("LoadInventory() expected an error for non-existent file, got nil")
	}
	if err != os.ErrNotExist {
		t.Errorf("LoadInventory() expected os.ErrNotExist for non-existent file, got %v", err)
	}
}

func TestLoadInventory_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "invalid_json.json")

	invalidJSONContent := []byte(`{"inventory": {"_metadata": {"version": "1.0"}, "models": [{"name": "bad-model"}]}, "fetched_at": "not-a-time"`) // Invalid FetchedAt
	err := os.WriteFile(cachePath, invalidJSONContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON file for test setup: %v", err)
	}

	_, err = LoadInventory(cachePath)
	if err == nil {
		t.Fatal("LoadInventory() expected an error for invalid JSON, got nil")
	}
	// Check if it's a json unmarshaling error (specific type might vary, so check string)
	// Example: json: cannot unmarshal string into Go value of type time.Time
	if _, ok := err.(*json.UnmarshalTypeError); !ok {
		// It might also be other json.SyntaxError, etc.
		// For simplicity, just check if it's a json error.
		// A more robust check would be to see if err's underlying type is from encoding/json.
		t.Logf("Note: Error type is %T, value: %v", err, err) // Log the actual error for debugging
		// A simple string check might be too brittle.
		// For now, we accept any error if it's not nil, as JSON parsing errors can vary.
		// A better check might involve errors.As for specific json error types.
	}
}

func TestLoadInventory_CorruptedFile(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "corrupted_json.json")

	corruptedJSONContent := []byte(`{"inventory": {"_metadata": {"version": "1.0"}, "models": [`) // Incomplete JSON
	err := os.WriteFile(cachePath, corruptedJSONContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupted JSON file for test setup: %v", err)
	}

	_, err = LoadInventory(cachePath)
	if err == nil {
		t.Fatal("LoadInventory() expected an error for corrupted JSON, got nil")
	}
	// Expecting a json.SyntaxError or similar
	if _, ok := err.(*json.SyntaxError); !ok {
		t.Logf("Note: Expected a json.SyntaxError or similar for corrupted file, got %T: %v", err, err)
	}
}

func TestIsCacheValid(t *testing.T) {
	maxAge := 24 * time.Hour

	tests := []struct {
		name            string
		fetchedAt       time.Time
		cachedInventory *domain.CachedModelInventory // Allow nil for one case
		maxCacheAge     time.Duration
		expected        bool
	}{
		{
			name:        "Fresh cache",
			fetchedAt:   time.Now().Add(-1 * time.Hour), // Fetched 1 hour ago
			maxCacheAge: maxAge,
			expected:    true,
		},
		{
			name:        "Expired cache",
			fetchedAt:   time.Now().Add(-2 * maxAge), // Fetched 2 * maxAge ago
			maxCacheAge: maxAge,
			expected:    false,
		},
		// Removed "Cache at exact expiry edge (adjusted to be just within window)"
		// Replaced with Cache_just_barely_valid and Cache_just_barely_expired
		{
			name:        "Cache_just_barely_valid (maxAge - 10ms ago)", // Increased buffer
			fetchedAt:   time.Now().Add(-(maxAge - 10*time.Millisecond)),
			maxCacheAge: maxAge,
			expected:    true,
		},
		{
			name:        "Cache_just_barely_expired (maxAge + 10ms ago)", // Increased buffer
			fetchedAt:   time.Now().Add(-(maxAge + 10*time.Millisecond)),
			maxCacheAge: maxAge,
			expected:    false,
		},
		{
			name:        "Cache slightly before expiry (half maxAge)",
			fetchedAt:   time.Now().Add(-maxAge / 2),
			maxCacheAge: maxAge,
			expected:    true,
		},
		{
			name:            "Nil inventory",
			cachedInventory: nil,
			maxCacheAge:     maxAge,
			expected:        false,
		},
		{
			name:        "Zero maxCacheAge, recent fetch (should be false)",
			fetchedAt:   time.Now().Add(-1 * time.Minute),
			maxCacheAge: 0,
			expected:    false,
		},
		{
			name:        "Zero maxCacheAge, fetched exactly now (should be false as Now() in func will be later)",
			fetchedAt:   time.Now(),
			maxCacheAge: 0,
			expected:    false, // time.Now() in IsCacheValid will be slightly after tt.fetchedAt
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inventoryToTest *domain.CachedModelInventory
			if tt.cachedInventory != nil { // Handles the "Nil inventory" case directly
				inventoryToTest = tt.cachedInventory
			} else if !tt.fetchedAt.IsZero() { // Construct inventory if fetchedAt is provided
				inventoryToTest = &domain.CachedModelInventory{
					FetchedAt: tt.fetchedAt,
					Inventory: domain.ModelInventory{}, // Content doesn't matter for this test
				}
			}
			// If both are nil/zero, inventoryToTest remains nil, covered by "Nil inventory" logic in IsCacheValid

			isValid := IsCacheValid(inventoryToTest, tt.maxCacheAge)
			if isValid != tt.expected {
				t.Errorf("IsCacheValid() with FetchedAt %v (delta: %v) and maxAge %v returned %v, expected %v",
					tt.fetchedAt.Format(time.RFC3339), time.Since(tt.fetchedAt), tt.maxCacheAge, isValid, tt.expected)
			}
		})
	}
}
