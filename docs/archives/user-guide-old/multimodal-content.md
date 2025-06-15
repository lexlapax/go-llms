# Working with Multimodal Content

Learn how to use images, audio, video, and files with LLMs in go-llms.

## Overview

Modern LLMs can understand more than just text. With go-llms, you can send:
- **Images** - Photos, diagrams, screenshots
- **Audio** - Voice recordings, music, sounds
- **Video** - Clips and recordings
- **Files** - Documents like PDFs

## Provider Support

Not all providers support all content types:

| Provider  | Text | Images | Audio | Video | Files |
|-----------|------|--------|-------|-------|-------|
| OpenAI    | ✅   | ✅     | ✅    | ✅    | ✅    |
| Anthropic | ✅   | ✅     | ❌    | ❌    | ❌    |
| Gemini    | ✅   | ✅     | ✅    | ✅    | ❌    |

Don't worry - go-llms handles unsupported types gracefully with clear error messages.

## Sending Images

### From a File

```go
import (
    "os"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
)

// Read image file
imageData, err := os.ReadFile("photo.jpg")
if err != nil {
    log.Fatal(err)
}

// Create image message
message := domain.NewImageMessage(
    domain.RoleUser,
    imageData,
    "image/jpeg",
    "What's in this photo?",
)

// Send to LLM
response, err := provider.GenerateMessage(ctx, []domain.Message{message})
```

### From a URL

```go
// Use an image URL directly
message := domain.NewImageURLMessage(
    domain.RoleUser,
    "https://example.com/chart.png",
    "Explain this chart",
)

response, err := provider.GenerateMessage(ctx, []domain.Message{message})
```

### Common Image Use Cases

```go
// Analyze a chart
chartMsg := domain.NewImageMessage(
    domain.RoleUser,
    chartData,
    "image/png",
    "What trends do you see in this chart?",
)

// OCR text from image
textMsg := domain.NewImageMessage(
    domain.RoleUser,
    documentImage,
    "image/jpeg",
    "Extract all text from this image",
)

// Describe a scene
sceneMsg := domain.NewImageMessage(
    domain.RoleUser,
    photoData,
    "image/jpeg",
    "Describe what's happening in this photo in detail",
)
```

## Sending Audio

Audio support lets you transcribe, analyze, and understand audio content:

```go
// Read audio file
audioData, err := os.ReadFile("recording.mp3")
if err != nil {
    log.Fatal(err)
}

// Create audio message
message := domain.NewAudioMessage(
    domain.RoleUser,
    audioData,
    "audio/mp3",
    "Transcribe this recording",
)

// Send to provider
response, err := provider.GenerateMessage(ctx, []domain.Message{message})
```

### Audio Use Cases

```go
// Transcription
transcribeMsg := domain.NewAudioMessage(
    domain.RoleUser,
    audioData,
    "audio/wav",
    "Transcribe this interview",
)

// Analysis
analyzeMsg := domain.NewAudioMessage(
    domain.RoleUser,
    musicData,
    "audio/mp3",
    "What instruments do you hear?",
)

// Translation
translateMsg := domain.NewAudioMessage(
    domain.RoleUser,
    speechData,
    "audio/mp3",
    "Translate this speech to English",
)
```

## Sending Video

Analyze video content for actions, objects, and events:

```go
// Read video file
videoData, err := os.ReadFile("clip.mp4")
if err != nil {
    log.Fatal(err)
}

// Create video message
message := domain.NewVideoMessage(
    domain.RoleUser,
    videoData,
    "video/mp4",
    "Summarize what happens in this video",
)

response, err := provider.GenerateMessage(ctx, []domain.Message{message})
```

### Video Use Cases

```go
// Action recognition
actionMsg := domain.NewVideoMessage(
    domain.RoleUser,
    sportsClip,
    "video/mp4",
    "Describe the techniques shown in this sports clip",
)

// Content moderation
moderateMsg := domain.NewVideoMessage(
    domain.RoleUser,
    userVideo,
    "video/mp4",
    "Is this video appropriate for all ages?",
)

// Tutorial analysis
tutorialMsg := domain.NewVideoMessage(
    domain.RoleUser,
    howToVideo,
    "video/mp4",
    "List the steps shown in this tutorial",
)
```

## Sending Files

Send documents for analysis, summarization, or extraction:

```go
// Read document
pdfData, err := os.ReadFile("report.pdf")
if err != nil {
    log.Fatal(err)
}

// Create file message
message := domain.NewFileMessage(
    domain.RoleUser,
    "report.pdf",
    pdfData,
    "application/pdf",
    "Summarize the key findings in this report",
)

response, err := provider.GenerateMessage(ctx, []domain.Message{message})
```

## Combining Content Types

Mix different content types in conversations:

```go
messages := []domain.Message{
    // System prompt
    domain.NewTextMessage(
        domain.RoleSystem,
        "You are a helpful analyst",
    ),
    
    // Send an image
    domain.NewImageMessage(
        domain.RoleUser,
        chartData,
        "image/png",
        "What does this chart show?",
    ),
    
    // LLM responds
    domain.NewTextMessage(
        domain.RoleAssistant,
        "This chart shows quarterly revenue growth...",
    ),
    
    // Follow-up with document
    domain.NewFileMessage(
        domain.RoleUser,
        "q4-report.pdf",
        reportData,
        "application/pdf",
        "How does this compare to the Q4 report?",
    ),
}

response, err := provider.GenerateMessage(ctx, messages)
```

## Handling Size Limits

Each provider has file size limits:

```go
const (
    OpenAIMaxSize    = 25 * 1024 * 1024  // 25MB
    AnthropicMaxSize = 3.75 * 1024 * 1024 // 3.75MB
    GeminiMaxSize    = 20 * 1024 * 1024  // 20MB
)

// Check file size before sending
fileInfo, _ := os.Stat("large-file.pdf")
if fileInfo.Size() > OpenAIMaxSize {
    log.Fatal("File too large for OpenAI")
}

// Or compress/resize if needed
if fileInfo.Size() > AnthropicMaxSize {
    // Resize image
    resizedData := resizeImage(imageData, 1024, 768)
    message := domain.NewImageMessage(
        domain.RoleUser,
        resizedData,
        "image/jpeg",
        "Analyze this image",
    )
}
```

## Error Handling

Handle unsupported content types gracefully:

```go
// Try to send video to Anthropic
videoMsg := domain.NewVideoMessage(
    domain.RoleUser,
    videoData,
    "video/mp4",
    "What's in this video?",
)

response, err := anthropicProvider.GenerateMessage(ctx, []domain.Message{videoMsg})
if err != nil {
    if domain.IsUnsupportedContentTypeError(err) {
        // Fallback to a different approach
        log.Printf("Video not supported, extracting frames instead...")
        
        // Extract key frames as images
        frames := extractKeyFrames(videoData)
        for i, frame := range frames {
            imgMsg := domain.NewImageMessage(
                domain.RoleUser,
                frame,
                "image/jpeg",
                fmt.Sprintf("Frame %d: What do you see?", i+1),
            )
            // Process each frame
        }
    }
}
```

## Best Practices

### 1. Choose the Right Format
```go
// For text extraction, use high-res images
highResImage := captureScreenshot(300) // 300 DPI

// For general analysis, standard resolution is fine
standardImage := captureScreenshot(72) // 72 DPI
```

### 2. Provide Clear Instructions
```go
// Be specific about what you want
vague := "What's this?"
better := "Identify all objects in this image and their locations"
best := "List each person in this photo with their apparent age, clothing, and activity"
```

### 3. Optimize File Sizes
```go
// Compress before sending
compressedImage := compressJPEG(imageData, 85) // 85% quality

// Convert formats if needed
pngData := convertToPNG(jpegData) // Some providers prefer PNG
```

### 4. Handle Fallbacks
```go
func analyzeContent(provider domain.Provider, content []byte, mimeType string) (string, error) {
    // Try video first
    if strings.HasPrefix(mimeType, "video/") {
        msg := domain.NewVideoMessage(domain.RoleUser, content, mimeType, "Analyze this")
        if response, err := provider.GenerateMessage(ctx, []domain.Message{msg}); err == nil {
            return response.Content, nil
        }
    }
    
    // Fallback to image frames
    if frames := extractFrames(content); len(frames) > 0 {
        msg := domain.NewImageMessage(domain.RoleUser, frames[0], "image/jpeg", "Analyze this")
        if response, err := provider.GenerateMessage(ctx, []domain.Message{msg}); err == nil {
            return response.Content, nil
        }
    }
    
    // Final fallback to text description
    return "Unable to analyze visual content", nil
}
```

## Real-World Examples

### Document Analysis Pipeline
```go
func analyzeDocument(provider domain.Provider, docPath string) (*Analysis, error) {
    // Read document
    docData, err := os.ReadFile(docPath)
    if err != nil {
        return nil, err
    }
    
    // Create conversation
    messages := []domain.Message{
        domain.NewTextMessage(
            domain.RoleSystem,
            "You are a document analyst. Extract key information clearly.",
        ),
        domain.NewFileMessage(
            domain.RoleUser,
            filepath.Base(docPath),
            docData,
            "application/pdf",
            "Extract: 1) Summary 2) Key points 3) Action items 4) Dates",
        ),
    }
    
    // Get analysis
    response, err := provider.GenerateMessage(ctx, messages)
    if err != nil {
        return nil, err
    }
    
    // Parse structured response
    return parseAnalysis(response.Content), nil
}
```

### Visual QA System
```go
func visualQA(provider domain.Provider, imagePath string, question string) (string, error) {
    imageData, err := os.ReadFile(imagePath)
    if err != nil {
        return "", err
    }
    
    messages := []domain.Message{
        domain.NewImageMessage(
            domain.RoleUser,
            imageData,
            "image/jpeg",
            question,
        ),
    }
    
    response, err := provider.GenerateMessage(ctx, messages)
    if err != nil {
        return "", err
    }
    
    return response.Content, nil
}

// Usage
answer, _ := visualQA(provider, "diagram.png", "Explain how this system works")
```

### Multi-Modal Assistant
```go
func processUserInput(provider domain.Provider, input UserInput) (string, error) {
    var messages []domain.Message
    
    // Add system prompt
    messages = append(messages, domain.NewTextMessage(
        domain.RoleSystem,
        "You are a helpful assistant that can analyze various content types.",
    ))
    
    // Add user content based on type
    switch input.Type {
    case "image":
        messages = append(messages, domain.NewImageMessage(
            domain.RoleUser,
            input.Data,
            input.MimeType,
            input.Question,
        ))
    case "audio":
        messages = append(messages, domain.NewAudioMessage(
            domain.RoleUser,
            input.Data,
            input.MimeType,
            input.Question,
        ))
    case "document":
        messages = append(messages, domain.NewFileMessage(
            domain.RoleUser,
            input.Filename,
            input.Data,
            input.MimeType,
            input.Question,
        ))
    default:
        messages = append(messages, domain.NewTextMessage(
            domain.RoleUser,
            input.Question,
        ))
    }
    
    response, err := provider.GenerateMessage(ctx, messages)
    if err != nil {
        return "", fmt.Errorf("failed to process %s: %w", input.Type, err)
    }
    
    return response.Content, nil
}
```

## Next Steps

- See [Providers](providers.md) for provider-specific multimodal features
- Check [Examples Gallery](examples-gallery.md#multimodal) for complete examples
- Learn about [Structured Output](structured-output.md) to extract data from multimodal content

Ready to go beyond text? Start building multimodal applications! 🖼️🎵🎬