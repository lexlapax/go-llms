// ABOUTME: Defines artifacts for storing files, images, and data objects in agent state
// ABOUTME: Provides immutable data containers that can be passed between agents

package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

// Artifact represents a file or data artifact that can be stored in agent state.
// It provides a flexible container for various content types including files, images,
// documents, and arbitrary data with support for both in-memory and streaming access.
type Artifact struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     ArtifactType           `json:"type"`
	MimeType string                 `json:"mime_type"`
	Size     int64                  `json:"size"`
	Created  time.Time              `json:"created"`
	Metadata map[string]interface{} `json:"metadata"`

	// Content storage (one of these will be used)
	data     []byte        // For in-memory data
	reader   io.ReadCloser // For streaming data
	location string        // For file references
}

// ArtifactType represents the type of artifact stored.
// It helps categorize artifacts for processing and display purposes.
type ArtifactType string

const (
	ArtifactTypeFile     ArtifactType = "file"
	ArtifactTypeImage    ArtifactType = "image"
	ArtifactTypeVideo    ArtifactType = "video"
	ArtifactTypeAudio    ArtifactType = "audio"
	ArtifactTypeDocument ArtifactType = "document"
	ArtifactTypeData     ArtifactType = "data"
	ArtifactTypeModel    ArtifactType = "model"
	ArtifactTypeCode     ArtifactType = "code"
	ArtifactTypeLog      ArtifactType = "log"
	ArtifactTypeCustom   ArtifactType = "custom"
)

// NewArtifact creates a new artifact with in-memory data.
// The artifact is assigned a unique ID and the current timestamp.
// Use this for small artifacts that fit comfortably in memory.
func NewArtifact(name string, artifactType ArtifactType, data []byte) *Artifact {
	return &Artifact{
		ID:       uuid.New().String(),
		Name:     name,
		Type:     artifactType,
		Size:     int64(len(data)),
		Created:  time.Now(),
		Metadata: make(map[string]interface{}),
		data:     data,
	}
}

// NewArtifactFromReader creates a new artifact from an io.ReadCloser.
// This is ideal for streaming large files without loading them entirely into memory.
// The size parameter should specify the expected content size if known.
func NewArtifactFromReader(name string, artifactType ArtifactType, reader io.ReadCloser, size int64) *Artifact {
	return &Artifact{
		ID:       uuid.New().String(),
		Name:     name,
		Type:     artifactType,
		Size:     size,
		Created:  time.Now(),
		Metadata: make(map[string]interface{}),
		reader:   reader,
	}
}

// NewArtifactFromLocation creates a new artifact referencing an external file.
// The artifact stores only the file path, not the content itself.
// This is useful for very large files that should not be loaded into memory.
func NewArtifactFromLocation(name string, artifactType ArtifactType, location string, size int64) *Artifact {
	return &Artifact{
		ID:       uuid.New().String(),
		Name:     name,
		Type:     artifactType,
		Size:     size,
		Created:  time.Now(),
		Metadata: make(map[string]interface{}),
		location: location,
	}
}

// WithMimeType sets the MIME type for the artifact.
// Returns the artifact for method chaining.
func (a *Artifact) WithMimeType(mimeType string) *Artifact {
	a.MimeType = mimeType
	return a
}

// WithMetadata adds a key-value metadata entry to the artifact.
// Metadata can store additional context like source, processing status, or custom attributes.
// Returns the artifact for method chaining.
func (a *Artifact) WithMetadata(key string, value interface{}) *Artifact {
	if a.Metadata == nil {
		a.Metadata = make(map[string]interface{})
	}
	a.Metadata[key] = value
	return a
}

// Read returns an io.ReadCloser for accessing the artifact content.
// For in-memory artifacts, it wraps the data in a reader.
// For file references, it returns an error indicating external file access is needed.
func (a *Artifact) Read() (io.ReadCloser, error) {
	// If we have a reader, return it
	if a.reader != nil {
		return a.reader, nil
	}

	// If we have data, wrap it in a reader
	if a.data != nil {
		return io.NopCloser(bytes.NewReader(a.data)), nil
	}

	// If we have a location, caller needs to handle file reading
	if a.location != "" {
		return nil, fmt.Errorf("artifact references external file: %s", a.location)
	}

	return nil, fmt.Errorf("no content available for artifact %s", a.ID)
}

// Data returns the artifact content as a byte slice.
// If the artifact uses a reader, it will read all content into memory.
// Returns an error if the artifact references an external file.
func (a *Artifact) Data() ([]byte, error) {
	if a.data != nil {
		return a.data, nil
	}

	// Try to read from reader if available
	if a.reader != nil {
		data, err := io.ReadAll(a.reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read artifact data: %w", err)
		}
		// Cache the data
		a.data = data
		// Reset reader for future reads
		a.reader = io.NopCloser(bytes.NewReader(data))
		return data, nil
	}

	return nil, fmt.Errorf("artifact data not loaded in memory")
}

// Location returns the file path for artifacts referencing external files.
// Returns an empty string for in-memory or streaming artifacts.
func (a *Artifact) Location() string {
	return a.location
}

// IsInMemory returns true if the artifact data is loaded in memory.
// In-memory artifacts provide fastest access but consume memory.
func (a *Artifact) IsInMemory() bool {
	return a.data != nil
}

// IsStreaming returns true if the artifact uses an io.Reader for content access.
// Streaming artifacts are memory-efficient for large files.
func (a *Artifact) IsStreaming() bool {
	return a.reader != nil
}

// IsReference returns true if the artifact references an external file by path.
// Referenced artifacts require separate file system access to read content.
func (a *Artifact) IsReference() bool {
	return a.location != ""
}

// Clone creates a shallow copy of the artifact.
// The metadata is deep copied but the content (data, reader, or location) is shared.
// The clone retains the same ID as the original artifact.
func (a *Artifact) Clone() *Artifact {
	metadata := make(map[string]interface{})
	for k, v := range a.Metadata {
		metadata[k] = v
	}

	return &Artifact{
		ID:       a.ID, // Keep same ID for clones
		Name:     a.Name,
		Type:     a.Type,
		MimeType: a.MimeType,
		Size:     a.Size,
		Created:  a.Created,
		Metadata: metadata,
		data:     a.data,
		reader:   a.reader,
		location: a.location,
	}
}

// MarshalJSON customizes JSON marshaling to exclude binary content.
// Only metadata and references are included in the JSON representation.
// Binary data and readers are intentionally omitted for efficiency.
func (a *Artifact) MarshalJSON() ([]byte, error) {
	// For JSON serialization, we only include metadata, not content
	return json.Marshal(map[string]interface{}{
		"id":        a.ID,
		"name":      a.Name,
		"type":      a.Type,
		"mime_type": a.MimeType,
		"size":      a.Size,
		"created":   a.Created,
		"metadata":  a.Metadata,
		"location":  a.location,
		// Note: data and reader are not serialized
	})
}

// Common MIME types
const (
	MimeTypeJSON     = "application/json"
	MimeTypeText     = "text/plain"
	MimeTypeHTML     = "text/html"
	MimeTypeMarkdown = "text/markdown"
	MimeTypePDF      = "application/pdf"
	MimeTypePNG      = "image/png"
	MimeTypeJPEG     = "image/jpeg"
	MimeTypeGIF      = "image/gif"
	MimeTypeMP4      = "video/mp4"
	MimeTypeMP3      = "audio/mpeg"
	MimeTypeWAV      = "audio/wav"
	MimeTypeZip      = "application/zip"
	MimeTypeBinary   = "application/octet-stream"
)

// GuessArtifactType attempts to determine the appropriate ArtifactType from a MIME type string.
// It maps common MIME types to artifact categories like document, image, video, etc.
// Returns ArtifactTypeFile for unrecognized MIME types.
func GuessArtifactType(mimeType string) ArtifactType {
	switch mimeType {
	case MimeTypePDF, MimeTypeHTML, MimeTypeMarkdown:
		return ArtifactTypeDocument
	case MimeTypePNG, MimeTypeJPEG, MimeTypeGIF:
		return ArtifactTypeImage
	case MimeTypeMP4:
		return ArtifactTypeVideo
	case MimeTypeMP3, MimeTypeWAV:
		return ArtifactTypeAudio
	case MimeTypeJSON:
		return ArtifactTypeData
	default:
		return ArtifactTypeFile
	}
}
