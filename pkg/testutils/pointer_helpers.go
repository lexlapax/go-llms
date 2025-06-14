package testutils

// ABOUTME: Compatibility layer for pointer helper functions
// ABOUTME: These are re-exported from the helpers package for backward compatibility

import "github.com/lexlapax/go-llms/pkg/testutils/helpers"

// IntPtr returns a pointer to an int value
// Deprecated: Use helpers.IntPtr instead
func IntPtr(i int) *int {
	return helpers.IntPtr(i)
}

// Float64Ptr returns a pointer to a float64 value
// Deprecated: Use helpers.Float64Ptr instead
func Float64Ptr(f float64) *float64 {
	return helpers.Float64Ptr(f)
}

// BoolPtr returns a pointer to a bool value
// Deprecated: Use helpers.BoolPtr instead
func BoolPtr(b bool) *bool {
	return helpers.BoolPtr(b)
}

// StringPtr returns a pointer to a string value
// Deprecated: Use helpers.StringPtr instead
func StringPtr(s string) *string {
	return helpers.StringPtr(s)
}
