package tools

//go:generate go run ../../../internal/toolgen/. -input ../builtins/tools -output registry_metadata.go -factory registry_factories.go -v

// This file contains the go:generate directive to regenerate tool metadata
// Run `go generate ./pkg/agent/tools` to update the tool registry
