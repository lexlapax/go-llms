package tools

import (
	"testing"
)

func TestToolManifestLoaded(t *testing.T) {
	// Check that ToolManifest is populated
	if len(ToolManifest) == 0 {
		t.Fatal("ToolManifest is empty - generated metadata not loaded")
	}

	// Check a few known tools
	expectedTools := []string{"calculator", "web_search", "file_read", "datetime_now"}
	for _, name := range expectedTools {
		if _, exists := ToolManifest[name]; !exists {
			t.Errorf("Expected tool %s not found in ToolManifest", name)
		}
	}

	t.Logf("ToolManifest contains %d tools", len(ToolManifest))
}

func TestDiscoveryMetadataLoaded(t *testing.T) {
	discovery := NewDiscovery()
	tools := discovery.ListTools()

	if len(tools) == 0 {
		t.Fatal("Discovery returned no tools - metadata not loaded")
	}

	t.Logf("Discovery contains %d tools", len(tools))
}
