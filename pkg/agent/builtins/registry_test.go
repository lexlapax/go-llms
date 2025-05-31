// ABOUTME: Tests for the generic registry implementation
// ABOUTME: Verifies thread-safety, search functionality, and metadata handling

package builtins

import (
	"fmt"
	"sync"
	"testing"
)

// mockComponent is a simple component for testing
type mockComponent struct {
	ID   string
	Data string
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry[*mockComponent]()

	tests := []struct {
		name      string
		compName  string
		component *mockComponent
		metadata  Metadata
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "successful registration",
			compName:  "test1",
			component: &mockComponent{ID: "1", Data: "data1"},
			metadata: Metadata{
				Category:    "test",
				Description: "Test component 1",
			},
			wantErr: false,
		},
		{
			name:      "empty name",
			compName:  "",
			component: &mockComponent{ID: "2", Data: "data2"},
			metadata:  Metadata{},
			wantErr:   true,
			errMsg:    "component name cannot be empty",
		},
		{
			name:      "duplicate registration",
			compName:  "test1",
			component: &mockComponent{ID: "3", Data: "data3"},
			metadata:  Metadata{},
			wantErr:   true,
			errMsg:    "component 'test1' is already registered",
		},
		{
			name:      "metadata name mismatch",
			compName:  "test2",
			component: &mockComponent{ID: "4", Data: "data4"},
			metadata: Metadata{
				Name: "different",
			},
			wantErr: true,
			errMsg:  "metadata name 'different' does not match registration name 'test2'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.compName, tt.component, tt.metadata)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry[*mockComponent]()

	comp1 := &mockComponent{ID: "1", Data: "data1"}
	registry.Register("test1", comp1, Metadata{Category: "test"})

	tests := []struct {
		name      string
		compName  string
		wantFound bool
		wantComp  *mockComponent
	}{
		{
			name:      "existing component",
			compName:  "test1",
			wantFound: true,
			wantComp:  comp1,
		},
		{
			name:      "non-existing component",
			compName:  "test2",
			wantFound: false,
			wantComp:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp, found := registry.Get(tt.compName)
			if found != tt.wantFound {
				t.Errorf("expected found=%v, got %v", tt.wantFound, found)
			}
			if found && comp != tt.wantComp {
				t.Errorf("expected component %v, got %v", tt.wantComp, comp)
			}
		})
	}
}

func TestRegistry_MustGet(t *testing.T) {
	registry := NewRegistry[*mockComponent]()

	comp := &mockComponent{ID: "1", Data: "data1"}
	registry.Register("test1", comp, Metadata{})

	// Test successful MustGet
	result := registry.MustGet("test1")
	if result != comp {
		t.Errorf("expected component %v, got %v", comp, result)
	}

	// Test panic on missing component
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic but got none")
		} else if msg, ok := r.(string); !ok || msg != "component 'missing' not found in registry" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()
	registry.MustGet("missing")
}

func TestRegistry_ListByCategory(t *testing.T) {
	registry := NewRegistry[*mockComponent]()

	registry.Register("web1", &mockComponent{ID: "1"}, Metadata{Category: "web"})
	registry.Register("web2", &mockComponent{ID: "2"}, Metadata{Category: "web"})
	registry.Register("file1", &mockComponent{ID: "3"}, Metadata{Category: "file"})
	registry.Register("data1", &mockComponent{ID: "4"}, Metadata{Category: "data"})

	webComponents := registry.ListByCategory("web")
	if len(webComponents) != 2 {
		t.Errorf("expected 2 web components, got %d", len(webComponents))
	}

	fileComponents := registry.ListByCategory("file")
	if len(fileComponents) != 1 {
		t.Errorf("expected 1 file component, got %d", len(fileComponents))
	}

	// Test case-insensitive search
	webComponentsUpper := registry.ListByCategory("WEB")
	if len(webComponentsUpper) != 2 {
		t.Errorf("expected 2 web components with uppercase search, got %d", len(webComponentsUpper))
	}
}

func TestRegistry_ListByTags(t *testing.T) {
	registry := NewRegistry[*mockComponent]()

	registry.Register("comp1", &mockComponent{ID: "1"}, Metadata{
		Tags: []string{"http", "web", "fetch"},
	})
	registry.Register("comp2", &mockComponent{ID: "2"}, Metadata{
		Tags: []string{"http", "web", "scrape"},
	})
	registry.Register("comp3", &mockComponent{ID: "3"}, Metadata{
		Tags: []string{"file", "read"},
	})

	// Test single tag
	httpComponents := registry.ListByTags("http")
	if len(httpComponents) != 2 {
		t.Errorf("expected 2 http components, got %d", len(httpComponents))
	}

	// Test multiple tags (AND operation)
	webFetchComponents := registry.ListByTags("web", "fetch")
	if len(webFetchComponents) != 1 {
		t.Errorf("expected 1 component with both web and fetch tags, got %d", len(webFetchComponents))
	}

	// Test no matching tags
	nonExistentComponents := registry.ListByTags("nonexistent")
	if len(nonExistentComponents) != 0 {
		t.Errorf("expected 0 components with nonexistent tag, got %d", len(nonExistentComponents))
	}

	// Test empty tags returns all
	allComponents := registry.ListByTags()
	if len(allComponents) != 3 {
		t.Errorf("expected 3 components with no tags filter, got %d", len(allComponents))
	}
}

func TestRegistry_Search(t *testing.T) {
	registry := NewRegistry[*mockComponent]()

	registry.Register("web_fetch", &mockComponent{ID: "1"}, Metadata{
		Category:    "web",
		Description: "Fetches content from URLs",
		Tags:        []string{"http", "download"},
	})
	registry.Register("file_read", &mockComponent{ID: "2"}, Metadata{
		Category:    "file",
		Description: "Reads file contents",
		Tags:        []string{"io", "filesystem"},
	})
	registry.Register("json_parse", &mockComponent{ID: "3"}, Metadata{
		Category:    "data",
		Description: "Parses JSON data",
		Tags:        []string{"json", "parser"},
	})

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{
			name:     "search by name",
			query:    "fetch",
			expected: 1,
		},
		{
			name:     "search by description",
			query:    "content",
			expected: 2, // "Fetches content" and "Reads file contents"
		},
		{
			name:     "search by category",
			query:    "web",
			expected: 1,
		},
		{
			name:     "search by tag",
			query:    "json",
			expected: 1,
		},
		{
			name:     "case insensitive search",
			query:    "JSON",
			expected: 1,
		},
		{
			name:     "empty query returns all",
			query:    "",
			expected: 3,
		},
		{
			name:     "no matches",
			query:    "xyz",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.Search(tt.query)
			if len(results) != tt.expected {
				t.Errorf("expected %d results for query '%s', got %d", tt.expected, tt.query, len(results))
			}
		})
	}
}

func TestRegistry_Categories(t *testing.T) {
	registry := NewRegistry[*mockComponent]()

	registry.Register("comp1", &mockComponent{}, Metadata{Category: "web"})
	registry.Register("comp2", &mockComponent{}, Metadata{Category: "file"})
	registry.Register("comp3", &mockComponent{}, Metadata{Category: "web"})
	registry.Register("comp4", &mockComponent{}, Metadata{Category: "data"})
	registry.Register("comp5", &mockComponent{}, Metadata{}) // No category

	categories := registry.Categories()

	// Should have 3 unique categories
	if len(categories) != 3 {
		t.Errorf("expected 3 categories, got %d", len(categories))
	}

	// Check that all expected categories are present
	expectedCategories := map[string]bool{"web": true, "file": true, "data": true}
	for _, cat := range categories {
		if !expectedCategories[cat] {
			t.Errorf("unexpected category: %s", cat)
		}
		delete(expectedCategories, cat)
	}

	if len(expectedCategories) > 0 {
		t.Errorf("missing categories: %v", expectedCategories)
	}
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry[*mockComponent]()

	// Add some components
	registry.Register("comp1", &mockComponent{}, Metadata{})
	registry.Register("comp2", &mockComponent{}, Metadata{})

	// Verify they exist
	if len(registry.List()) != 2 {
		t.Errorf("expected 2 components before clear, got %d", len(registry.List()))
	}

	// Clear the registry
	registry.Clear()

	// Verify it's empty
	if len(registry.List()) != 0 {
		t.Errorf("expected 0 components after clear, got %d", len(registry.List()))
	}
}

func TestRegistry_ThreadSafety(t *testing.T) {
	registry := NewRegistry[*mockComponent]()

	// Number of goroutines and operations
	numGoroutines := 10
	numOperations := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // 3 types of operations

	// Concurrent registrations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				name := fmt.Sprintf("comp_%d_%d", id, j)
				comp := &mockComponent{ID: fmt.Sprintf("%d-%d", id, j)}
				registry.Register(name, comp, Metadata{
					Category: fmt.Sprintf("cat%d", id%3),
					Tags:     []string{fmt.Sprintf("tag%d", id%5)},
				})
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Try various read operations
				registry.List()
				registry.ListByCategory(fmt.Sprintf("cat%d", j%3))
				registry.ListByTags(fmt.Sprintf("tag%d", j%5))
				registry.Search(fmt.Sprintf("comp_%d", j%numGoroutines))
				registry.Categories()

				// Try to get a specific component
				name := fmt.Sprintf("comp_%d_%d", j%numGoroutines, j%numOperations)
				registry.Get(name)
			}
		}(i)
	}

	// Concurrent searches
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				registry.Search(fmt.Sprintf("%d", j))
			}
		}(i)
	}

	wg.Wait()

	// Verify registry is in a consistent state
	entries := registry.List()
	if len(entries) != numGoroutines*numOperations {
		t.Errorf("expected %d entries, got %d", numGoroutines*numOperations, len(entries))
	}
}

func TestRegistry_Examples(t *testing.T) {
	registry := NewRegistry[*mockComponent]()

	// Register a component with examples
	registry.Register("web_fetch", &mockComponent{}, Metadata{
		Category:    "web",
		Description: "Fetches web content",
		Version:     "1.0.0",
		Examples: []Example{
			{
				Name:        "Basic Usage",
				Description: "Fetch a simple web page",
				Code:        `fetch("https://example.com")`,
			},
			{
				Name:        "With Headers",
				Description: "Fetch with custom headers",
				Code:        `fetch("https://api.example.com", {headers: {"Authorization": "Bearer token"}})`,
			},
		},
	})

	// Get the component and verify examples
	entries := registry.Search("web_fetch")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if len(entry.Metadata.Examples) != 2 {
		t.Errorf("expected 2 examples, got %d", len(entry.Metadata.Examples))
	}

	// Verify first example
	if entry.Metadata.Examples[0].Name != "Basic Usage" {
		t.Errorf("expected first example name 'Basic Usage', got '%s'", entry.Metadata.Examples[0].Name)
	}
}

func TestRegistry_DeprecatedAndExperimental(t *testing.T) {
	registry := NewRegistry[*mockComponent]()

	// Register deprecated component
	registry.Register("old_tool", &mockComponent{}, Metadata{
		Category:   "legacy",
		Deprecated: true,
	})

	// Register experimental component
	registry.Register("new_tool", &mockComponent{}, Metadata{
		Category:     "experimental",
		Experimental: true,
	})

	// Register normal component
	registry.Register("stable_tool", &mockComponent{}, Metadata{
		Category: "stable",
	})

	// Verify flags are preserved
	entries := registry.List()
	for _, entry := range entries {
		switch entry.Metadata.Name {
		case "old_tool":
			if !entry.Metadata.Deprecated {
				t.Error("expected old_tool to be deprecated")
			}
		case "new_tool":
			if !entry.Metadata.Experimental {
				t.Error("expected new_tool to be experimental")
			}
		case "stable_tool":
			if entry.Metadata.Deprecated || entry.Metadata.Experimental {
				t.Error("expected stable_tool to be neither deprecated nor experimental")
			}
		}
	}
}