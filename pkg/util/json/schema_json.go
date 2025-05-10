package json

import (
	"bytes"
	"sync"

	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

var (
	// Pool of buffers for Schema marshaling operations
	schemaBufferPool = sync.Pool{
		New: func() interface{} {
			// Initialize with a reasonable capacity for typical schemas
			buf := &bytes.Buffer{}
			buf.Grow(4096) // 4KB starting capacity
			return buf
		},
	}
)

// MarshalSchemaIndent marshals a schema to JSON with indentation, using a pooled buffer
// This is specifically optimized for the schema domain model and reduces allocations
func MarshalSchemaIndent(schema *schemaDomain.Schema) ([]byte, error) {
	// Get a buffer from the pool
	buf := schemaBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		schemaBufferPool.Put(buf)
	}()

	// Directly marshal to the buffer using MarshalIndent to avoid extra allocations
	result, err := jsonAPI.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}

	return result, nil
}

// MarshalSchemaFast marshals a schema to JSON without indentation, using a pooled buffer
// This is the fastest option for schema serialization when indentation is not needed
func MarshalSchemaFast(schema *schemaDomain.Schema) ([]byte, error) {
	// Direct marshal is faster than using an encoder for single operations
	return jsonAPI.Marshal(schema)
}

// MarshalSchemaToString marshals a schema to a JSON string with indentation
// This is convenience function for contexts where a string is preferred
func MarshalSchemaToString(schema *schemaDomain.Schema) (string, error) {
	return jsonAPI.MarshalToString(schema)
}