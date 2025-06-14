package tools

import (
	"testing"
)

func TestRegistryInitialization(t *testing.T) {
	// Check that ToolManifest is populated
	if len(ToolManifest) == 0 {
		t.Fatal("ToolManifest is empty - generated metadata not loaded")
	}
	t.Logf("ToolManifest contains %d tools", len(ToolManifest))

	// Check that discovery has metadata after init
	discovery := NewDiscovery()
	toolDiscovery := discovery.(*toolDiscovery)

	// Get tools from current namespace (default)
	tools := toolDiscovery.ListTools()
	if len(tools) == 0 {
		t.Fatal("Discovery has no tools after init")
	}
	t.Logf("Discovery contains %d tools", len(tools))

	// The exact count may vary depending on what's registered
	// Just ensure we have some tools loaded
	if len(tools) == 0 {
		t.Error("Expected some tools to be registered in discovery")
	}
}
