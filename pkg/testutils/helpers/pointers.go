// ABOUTME: Helper functions for creating pointers to primitive types
// ABOUTME: Simplifies test data setup by avoiding inline pointer creation

package helpers

// IntPtr returns a pointer to an int value
func IntPtr(i int) *int {
	return &i
}

// Int32Ptr returns a pointer to an int32 value
func Int32Ptr(i int32) *int32 {
	return &i
}

// Int64Ptr returns a pointer to an int64 value
func Int64Ptr(i int64) *int64 {
	return &i
}

// Float32Ptr returns a pointer to a float32 value
func Float32Ptr(f float32) *float32 {
	return &f
}

// Float64Ptr returns a pointer to a float64 value
func Float64Ptr(f float64) *float64 {
	return &f
}

// BoolPtr returns a pointer to a bool value
func BoolPtr(b bool) *bool {
	return &b
}

// StringPtr returns a pointer to a string value
func StringPtr(s string) *string {
	return &s
}

// BytePtr returns a pointer to a byte value
func BytePtr(b byte) *byte {
	return &b
}

// RunePtr returns a pointer to a rune value
func RunePtr(r rune) *rune {
	return &r
}
