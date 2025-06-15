# Multimodal Content Implementation

> **[Documentation Home](/docs/README.md) / [Technical Documentation](README.md) / Multimodal Content Implementation**

This document describes the technical implementation of multimodal content support in Go-LLMs. It covers the internal architecture, provider-specific adaptations, and extension points for contributors.

## Core Components

### 1. Message Structure

The core of the implementation is in `pkg/llm/domain/message.go`, which defines:

- `ContentType` for different types of content (text, image, file, video, audio)
- `SourceType` for how content is sourced (base64, URL)
- `ContentPart` for representing parts of a message's content
- Helper functions like `NewTextMessage`, `NewImageMessage`, etc.

```go
// ContentPart represents a part of a message's content
type ContentPart struct {
    Type  ContentType  `json:"type"`
    Text  string       `json:"text,omitempty"`
    Image *ImageContent `json:"image,omitempty"`
    File  *FileContent  `json:"file,omitempty"`
    Video *VideoContent `json:"video,omitempty"`
    Audio *AudioContent `json:"audio,omitempty"`
}

// Message represents a message in a conversation with multimodal support
type Message struct {
    Role    Role         `json:"role"`
    Content []ContentPart `json:"content"`
}
```

### 2. Provider-Specific Implementation

Each provider (OpenAI, Anthropic, Gemini) has been updated to:
- Convert multimodal messages to provider-specific formats
- Handle provider-specific responses into the library's standardized format

### 3. Backward Compatibility

Helper functions make it easy to work with the new structure:

```go
// Create a text-only message
message := NewTextMessage(domain.RoleUser, "Hello, world!")

// Create an image message with optional text
imageMessage := NewImageMessage(domain.RoleUser, imageData, "image/png", "This is an image of...")

// Create a message with an image URL
urlMessage := NewImageURLMessage(domain.RoleUser, "https://example.com/image.jpg", "An image from the web")
```

## Implementation Architecture

### Content Type Handling

The system uses a type-safe approach with dedicated structures for each content type:

```go
type ContentType string

const (
    ContentTypeText  ContentType = "text"
    ContentTypeImage ContentType = "image" 
    ContentTypeFile  ContentType = "file"
    ContentTypeVideo ContentType = "video"
    ContentTypeAudio ContentType = "audio"
)

type SourceType string

const (
    SourceTypeBase64 SourceType = "base64"
    SourceTypeURL    SourceType = "url"
)
```

### Memory Management

For binary content, the implementation uses efficient memory management:

```go
// ImageContent uses string for base64 data to avoid unnecessary []byte allocations
type ImageContent struct {
    Data      string     `json:"data,omitempty"`
    URL       string     `json:"url,omitempty"`
    MimeType  string     `json:"mime_type"`
    Source    SourceType `json:"source"`
}
```

## Testing Strategy

The multimodal implementation includes comprehensive tests:

### Unit Tests
- Message construction and validation
- Content type conversion
- Base64 encoding/decoding efficiency

### Integration Tests  
- Provider-specific format conversion
- Round-trip message handling
- Error cases for unsupported content types

### Test Helpers

```go
// Test helper for creating multimodal messages
func createTestImageMessage(t *testing.T) *domain.Message {
    imageData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG header
    return domain.NewImageMessage(
        domain.RoleUser,
        imageData,
        "image/png",
        "Test image",
    )
}
```

## Provider Implementation Details

### OpenAI

The OpenAI provider implementation:
- Converts library ContentPart objects to OpenAI's content format
- Maps content types to the appropriate OpenAI formats
- Handles base64 encoding for binary data

### Anthropic

The Anthropic provider implementation:
- Maps our ContentPart structure to Anthropic's message format
- Handles image and other media content types according to Anthropic's API requirements

### Gemini

The Gemini provider implementation:
- Converts ContentPart objects to Gemini's content format
- Implements appropriate handling for different media types

## Extension Points

### Adding New Content Types

To add support for new content types:

1. **Define the content type constant**:
```go
const ContentTypeNewType ContentType = "new_type"
```

2. **Create the content structure**:
```go
type NewTypeContent struct {
    Data     string     `json:"data,omitempty"`
    URL      string     `json:"url,omitempty"`
    MimeType string     `json:"mime_type"`
    Source   SourceType `json:"source"`
    // Add type-specific fields
}
```

3. **Update ContentPart**:
```go
type ContentPart struct {
    // ... existing fields
    NewType *NewTypeContent `json:"new_type,omitempty"`
}
```

4. **Add helper functions**:
```go
func NewNewTypeMessage(role Role, data []byte, mimeType, text string) *Message {
    // Implementation
}
```

### Provider-Specific Adaptations

Each provider requires specific adaptations in their respective implementation files:

```go
// In provider implementation (e.g., openai.go)
func (p *openAIProvider) convertContentParts(parts []domain.ContentPart) []openAIContent {
    // Handle provider-specific content format
}
```

## Performance Considerations

1. **Base64 Encoding**: Use string type for base64 data to avoid allocations
2. **Content Validation**: Validate content size limits per provider
3. **Memory Pooling**: Consider using sync.Pool for large binary content
4. **Streaming**: Support streaming for large media files (future)

## Security Considerations

1. **Content Validation**: Validate MIME types and file headers
2. **Size Limits**: Enforce provider-specific size limits
3. **URL Validation**: Validate URLs before fetching content
4. **Sanitization**: Sanitize file names and metadata

## Related Documentation

- [Provider Implementation Guide](provider-implementation.md)
- [Testing Framework](testing.md)
- [Performance Optimization](performance.md)