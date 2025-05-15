# Multimodal Content Implementation

## Overview

This document describes the implementation of multimodal content support in the go-llms library. The implementation enables the library to handle various types of content such as text, images, files, videos, and audio in messages sent to and received from different LLM providers.

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

## Testing

Tests have been added to verify:
- Proper conversion between the library's message format and provider-specific formats
- Base64 encoding/decoding of binary data
- Proper handling of different content types

## Usage Examples

### Text Message

```go
message := domain.NewTextMessage(domain.RoleUser, "Hello, how are you?")
```

### Image Message

```go
// From file
imageData, _ := os.ReadFile("image.png")
imageMessage := domain.NewImageMessage(domain.RoleUser, imageData, "image/png", "Describe this image")

// From URL
urlMessage := domain.NewImageURLMessage(domain.RoleUser, "https://example.com/image.jpg", "What's in this picture?")
```

### File Attachment

```go
fileData, _ := os.ReadFile("document.pdf")
fileMessage := domain.NewFileMessage(domain.RoleUser, "document.pdf", fileData, "application/pdf", "Summarize this document")
```

## Future Extensions

The implementation is designed to be extensible for future provider-specific features while maintaining a consistent API across the library.