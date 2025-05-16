# Multimodal Content Support

> **[Documentation Home](/REFERENCE.md) / [User Guide](README.md) / Multimodal Content**

Go-LLMs provides comprehensive support for multimodal content across all supported LLM providers. This guide explains how to use multimodal content types such as images, videos, audio, and files in your applications.

For technical implementation details, see the [Multimodal Content Implementation](../technical/multimodal-content.md) documentation.

## Overview

Modern LLM providers support various types of content beyond just text. Go-LLMs offers a unified way to work with:

- **Text**: Basic text messages
- **Images**: Both base64-encoded images and URL references
- **Files**: Documents and other file types
- **Videos**: Video content in various formats
- **Audio**: Audio content in various formats

Each provider has different capabilities regarding which content types they support:

| Provider  | Text | Images | Files | Videos | Audio |
|-----------|------|--------|-------|--------|-------|
| OpenAI    | ✅   | ✅     | ✅    | ✅     | ✅    |
| Anthropic | ✅   | ✅     | ✅    | ✅     | ✅    |
| Gemini    | ✅   | ✅     | ✅    | ✅     | ✅    |

When using unsupported content types, the library will return a clear error with detailed information.

## Creating Multimodal Messages

### Text Messages

Creating a simple text message:

```go
// Create a text message
textMessage := domain.NewTextMessage(domain.RoleUser, "Hello, world!")
```

### Image Messages

Sending images with text:

```go
// Using base64-encoded image data
imageData, _ := ioutil.ReadFile("image.jpg")
imageMessage := domain.NewImageMessage(
    domain.RoleUser,
    imageData,
    "image/jpeg",
    "What's in this image?"
)

// Using image URL
imageURLMessage := domain.NewImageURLMessage(
    domain.RoleUser,
    "https://example.com/image.jpg",
    "What's in this image?"
)
```

### File Messages

Sending file attachments:

```go
// Read and send a PDF file
fileData, _ := ioutil.ReadFile("document.pdf")
fileMessage := domain.NewFileMessage(
    domain.RoleUser,
    "document.pdf",
    fileData,
    "application/pdf",
    "Please analyze this document."
)
```

### Video Messages

Sending video content:

```go
// Read and send a video file
videoData, _ := ioutil.ReadFile("video.mp4")
videoMessage := domain.NewVideoMessage(
    domain.RoleUser,
    videoData,
    "video/mp4",
    "What's happening in this video?"
)
```

### Audio Messages

Sending audio content:

```go
// Read and send an audio file
audioData, _ := ioutil.ReadFile("audio.mp3")
audioMessage := domain.NewAudioMessage(
    domain.RoleUser,
    audioData,
    "audio/mp3",
    "What is said in this audio?"
)
```

## Handling Provider Limitations

When a content type is not supported by a provider, the library will return an `UnsupportedContentTypeError`:

```go
// Try to send video to Anthropic (unsupported)
videoMessage := domain.NewVideoMessage(
    domain.RoleUser,
    videoData,
    "video/mp4",
    "What's happening in this video?"
)

// This will return an error since Anthropic doesn't support video
response, err := anthropicProvider.GenerateMessage(ctx, []domain.Message{videoMessage})
if err != nil {
    if domain.IsUnsupportedContentTypeError(err) {
        // Handle unsupported content type error gracefully
        fmt.Printf("Provider %s doesn't support content type %s\n", 
            err.(*domain.UnsupportedContentTypeError).Provider,
            err.(*domain.UnsupportedContentTypeError).ContentType)
    }
}
```

## Advanced Usage: Multiple Content Types

You can combine multiple content types in a single conversation:

```go
// Create a conversation with multiple content types
messages := []domain.Message{
    domain.NewTextMessage(domain.RoleSystem, "You are a helpful assistant."),
    domain.NewImageMessage(domain.RoleUser, imageData, "image/jpeg", "What's in this image?"),
    domain.NewTextMessage(domain.RoleAssistant, "I see a cat in the image."),
    domain.NewTextMessage(domain.RoleUser, "What breed is it?"),
}

// Send to the provider
response, err := provider.GenerateMessage(ctx, messages)
```

## Provider-Specific Considerations

### OpenAI

OpenAI supports all content types and offers the most comprehensive multimodal capabilities. For images, OpenAI expects a specific format for base64-encoded images:

```
data:<mime-type>;base64,<base64-data>
```

The library handles this conversion automatically.

### Anthropic

Anthropic Claude supports text and image content. For images, it uses a specific format:

```json
{
  "type": "image",
  "source": {
    "type": "base64",
    "media_type": "image/jpeg",
    "data": "<base64-data>"
  }
}
```

Attempting to use unsupported content types (files, videos, audio) will result in an error.

### Gemini

Google Gemini supports text, image, and video content. It uses a specific format for media:

```json
{
  "inline_data": {
    "mime_type": "image/jpeg",
    "data": "<base64-data>"
  }
}
```

Attempting to use unsupported content types (files, audio) will result in an error.

## Performance Considerations

When working with multimodal content, keep in mind:

1. **Size Limits**: Each provider has different limits on file sizes:
   - OpenAI: 25MB for most content
   - Anthropic: 3.75MB for images
   - Gemini: 20MB for inline data

2. **Base64 Encoding**: Base64 encoding increases the size of binary data by approximately 33%.

3. **Message Caching**: The library implements caching for converted messages to improve performance when sending the same messages multiple times.

## Error Handling

The library provides specific error types for handling multimodal content issues:

```go
// Check for unsupported content type errors
if domain.IsUnsupportedContentTypeError(err) {
    // Handle accordingly
}

// Check for content validation errors
if errors.As(err, &domain.ErrContentTypeValidationFailed{}) {
    // Handle validation errors (e.g., file too large)
}
```

## Examples

For complete examples of working with multimodal content, check out the following example applications:

- [Multimodal Example](/cmd/examples/multimodal/) - Comprehensive multimodal example with support for all content types
- `cmd/examples/openai/multimodal.go` - Using multimodal content with OpenAI
- `cmd/examples/anthropic/vision.go` - Using vision capabilities with Anthropic Claude
- `cmd/examples/gemini/multimodal.go` - Using multimodal content with Google Gemini