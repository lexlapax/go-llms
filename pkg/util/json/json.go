// Package json provides an optimized JSON implementation with multiple backends
package json

import (
	"bytes"
	"io"

	jsoniter "github.com/json-iterator/go"
)

var (
	// Global configuration that matches standard library behavior
	jsonAPI = jsoniter.ConfigCompatibleWithStandardLibrary
)

// Marshal marshals the interface into a JSON byte array.
// It's a drop-in replacement for encoding/json.Marshal with better performance.
func Marshal(v interface{}) ([]byte, error) {
	return jsonAPI.Marshal(v)
}

// MarshalIndent marshals the interface into a JSON byte array with indentation.
// It's a drop-in replacement for encoding/json.MarshalIndent with better performance.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return jsonAPI.MarshalIndent(v, prefix, indent)
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
// It's a drop-in replacement for encoding/json.Unmarshal with better performance.
func Unmarshal(data []byte, v interface{}) error {
	return jsonAPI.Unmarshal(data, v)
}

// MarshalToString marshals the interface into a JSON string.
// This is more efficient than Marshal followed by string conversion.
func MarshalToString(v interface{}) (string, error) {
	return jsonAPI.MarshalToString(v)
}

// UnmarshalFromString parses the JSON-encoded string and stores the result in the value pointed to by v.
// This is more efficient than string to bytes conversion followed by Unmarshal.
func UnmarshalFromString(data string, v interface{}) error {
	return jsonAPI.UnmarshalFromString(data, v)
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *jsoniter.Encoder {
	return jsonAPI.NewEncoder(w)
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *jsoniter.Decoder {
	return jsonAPI.NewDecoder(r)
}

// Valid reports whether data is a valid JSON encoding.
func Valid(data []byte) bool {
	return jsonAPI.Valid(data)
}

// Get searches a path in json object and returns the JsonIter's Any object.
// Any is a data type that can be used to represent any JSON value with lazy parsing.
func Get(data []byte, path ...interface{}) jsoniter.Any {
	return jsonAPI.Get(data, path...)
}

// MarshalIndentWithBuffer is an optimized version of MarshalIndent that reuses a provided buffer
// to minimize allocations. The buffer will be reset and used for the output.
func MarshalIndentWithBuffer(v interface{}, buf *bytes.Buffer, prefix, indent string) error {
	if buf == nil {
		return nil
	}
	
	buf.Reset()
	enc := jsonAPI.NewEncoder(buf)
	enc.SetIndent(prefix, indent)
	return enc.Encode(v)
}

// MarshalWithBuffer is an optimized version of Marshal that reuses a provided buffer
// to minimize allocations. The buffer will be reset and used for the output.
func MarshalWithBuffer(v interface{}, buf *bytes.Buffer) error {
	if buf == nil {
		return nil
	}
	
	buf.Reset()
	return jsonAPI.NewEncoder(buf).Encode(v)
}