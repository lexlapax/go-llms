// ABOUTME: Pointer helper functions for creating pointers to literals in tests
// ABOUTME: Simplifies test setup when APIs require pointer types

package helpers

// Ptr returns a pointer to the given value
func Ptr[T any](v T) *T {
	return &v
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// Int32Ptr returns a pointer to an int32
func Int32Ptr(i int32) *int32 {
	return &i
}

// Int64Ptr returns a pointer to an int64
func Int64Ptr(i int64) *int64 {
	return &i
}

// Float32Ptr returns a pointer to a float32
func Float32Ptr(f float32) *float32 {
	return &f
}

// Float64Ptr returns a pointer to a float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// Deref safely dereferences a pointer, returning the zero value if nil
func Deref[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

// DerefOr returns the dereferenced value or the provided default if nil
func DerefOr[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}
