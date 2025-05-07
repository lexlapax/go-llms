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