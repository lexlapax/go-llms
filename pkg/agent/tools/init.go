// ABOUTME: Package initialization ensuring tool discovery system is loaded on import.
// ABOUTME: Forces metadata initialization even in test contexts for external packages.
package tools

// This file ensures that the generated metadata is always loaded
// when the tools package is imported, even in test contexts

func init() {
	// Force initialization of the discovery system with generated metadata
	// This ensures that external test packages can access the metadata
	_ = NewDiscovery()
}
