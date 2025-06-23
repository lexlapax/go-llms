// ABOUTME: Test helper utilities for schema validation tests.
// ABOUTME: Provides pointer creation functions to simplify test setup.

// Package validation provides JSON schema validation with performance optimizations.
// This file contains test helper functions for creating pointer values.
package validation

// Common test helpers to avoid redeclaration issues

// intPtr returns a pointer to an int value
func intPtr(v int) *int {
	return &v
}

// float64Ptr returns a pointer to a float64 value
func float64Ptr(v float64) *float64 {
	return &v
}
