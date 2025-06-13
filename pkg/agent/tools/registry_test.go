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

	if len(toolDiscovery.metadata) == 0 {
		t.Fatal("Discovery metadata is empty after init")
	}
	t.Logf("Discovery metadata contains %d tools", len(toolDiscovery.metadata))

	// Verify they match
	if len(toolDiscovery.metadata) != len(ToolManifest) {
		t.Errorf("Mismatch: ToolManifest has %d tools but discovery has %d",
			len(ToolManifest), len(toolDiscovery.metadata))
	}
}
