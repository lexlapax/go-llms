package tools

// This file ensures that the generated metadata is always loaded
// when the tools package is imported, even in test contexts

func init() {
	// Force initialization of the discovery system with generated metadata
	// This ensures that external test packages can access the metadata
	_ = NewDiscovery()
}
