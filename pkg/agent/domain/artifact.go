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

// Artifact represents a file or data artifact
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

// ArtifactType represents the type of artifact
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

// NewArtifact creates a new artifact with data
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

// NewArtifactFromReader creates a new artifact from a reader
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

// NewArtifactFromLocation creates a new artifact referencing a file location
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

// WithMimeType sets the MIME type
func (a *Artifact) WithMimeType(mimeType string) *Artifact {
	a.MimeType = mimeType
	return a
}

// WithMetadata adds metadata to the artifact
func (a *Artifact) WithMetadata(key string, value interface{}) *Artifact {
	if a.Metadata == nil {
		a.Metadata = make(map[string]interface{})
	}
	a.Metadata[key] = value
	return a
}

// Read returns a reader for the artifact content
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

// Data returns the artifact data (if loaded in memory)
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

// Location returns the file location (if artifact references a file)
func (a *Artifact) Location() string {
	return a.location
}

// IsInMemory returns true if the artifact data is loaded in memory
func (a *Artifact) IsInMemory() bool {
	return a.data != nil
}

// IsStreaming returns true if the artifact has a reader
func (a *Artifact) IsStreaming() bool {
	return a.reader != nil
}

// IsReference returns true if the artifact references an external file
func (a *Artifact) IsReference() bool {
	return a.location != ""
}

// Clone creates a copy of the artifact metadata (content is shared)
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

// MarshalJSON customizes JSON marshaling
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

// GuessArtifactType attempts to determine artifact type from MIME type
func GuessArtifactType(mimeType string) ArtifactType {
	switch {
	case mimeType == MimeTypePDF || mimeType == MimeTypeHTML || mimeType == MimeTypeMarkdown:
		return ArtifactTypeDocument
	case mimeType == MimeTypePNG || mimeType == MimeTypeJPEG || mimeType == MimeTypeGIF:
		return ArtifactTypeImage
	case mimeType == MimeTypeMP4:
		return ArtifactTypeVideo
	case mimeType == MimeTypeMP3 || mimeType == MimeTypeWAV:
		return ArtifactTypeAudio
	case mimeType == MimeTypeJSON:
		return ArtifactTypeData
	default:
		return ArtifactTypeFile
	}
}
