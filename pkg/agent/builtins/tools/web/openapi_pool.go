// ABOUTME: Object pools for OpenAPI-related operations to reduce memory allocations
// ABOUTME: Provides sync.Pool instances for commonly allocated objects

package web

import (
	"bytes"
	"sync"
)

var (
	// Pool for bytes.Buffer used in HTTP request/response handling
	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	// Pool for string slices used in error collection
	stringSlicePool = sync.Pool{
		New: func() interface{} {
			slice := make([]string, 0, 10)
			return &slice
		},
	}

	// Pool for map[string]interface{} used in responses
	mapPool = sync.Pool{
		New: func() interface{} {
			m := make(map[string]interface{})
			return &m
		},
	}
)

// GetBuffer gets a buffer from the pool
func GetBuffer() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 1024*1024 { // Don't pool buffers larger than 1MB
		return
	}
	bufferPool.Put(buf)
}

// GetStringSlice gets a string slice from the pool
func GetStringSlice() *[]string {
	slice := stringSlicePool.Get().(*[]string)
	*slice = (*slice)[:0] // Reset length but keep capacity
	return slice
}

// PutStringSlice returns a string slice to the pool
func PutStringSlice(slice *[]string) {
	if cap(*slice) > 100 { // Don't pool large slices
		return
	}
	stringSlicePool.Put(slice)
}

// GetMap gets a map from the pool
func GetMap() *map[string]interface{} {
	m := mapPool.Get().(*map[string]interface{})
	// Clear the map
	for k := range *m {
		delete(*m, k)
	}
	return m
}

// PutMap returns a map to the pool
func PutMap(m *map[string]interface{}) {
	if len(*m) > 100 { // Don't pool large maps
		return
	}
	mapPool.Put(m)
}
