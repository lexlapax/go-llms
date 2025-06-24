# Multimodal Content: Images, Audio, and Video

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Guides](/docs/user-guide/guides/) / Multimodal Content**

Master multimodal AI capabilities by processing images, audio, and video content with LLMs. Build intelligent systems that understand and analyze visual, auditory, and multimedia content across different providers.

## Why Multimodal AI Matters

- **Rich Understanding** - Process visual and auditory information beyond text
- **Real-World Applications** - Analyze photos, diagrams, videos, and audio recordings
- **Enhanced Context** - Combine multiple modalities for deeper insights
- **Creative Applications** - Generate descriptions, transcribe content, analyze media
- **Accessibility** - Convert between modalities for improved access

## Multimodal Architecture

![Multimodal Content Flow](../../images/multimodal-architecture.svg)

### Content Types
1. **Images** - Photos, diagrams, screenshots, charts
2. **Audio** - Speech, music, sound effects, recordings
3. **Video** - Clips, streams, presentations, recordings
4. **Files** - Documents, PDFs, general attachments
5. **Mixed Content** - Multiple types in single message

### Provider Capabilities
| Provider | Images | Audio | Video | Files | Notes |
|----------|--------|-------|-------|-------|-------|
| **OpenAI** | ✅ | ❌ | ❌ | ✅ | GPT-4 Vision models |
| **Anthropic** | ✅ | ❌ | ❌ | ❌ | Claude 3 models |
| **Google Gemini** | ✅ | ✅ | ✅ | ❌ | Most comprehensive |
| **Ollama** | ❌* | ❌ | ❌ | ❌ | *Some models support |
| **OpenRouter** | Varies | Varies | Varies | Varies | Model-dependent |

## Prerequisites

- [Provider Setup completed](provider-setup.md) ✅
- Vision-capable models configured ✅
- Basic understanding of MIME types ✅

---

## Level 1: Image Analysis
*Process and understand visual content*

### Basic Image Analysis
```go
package main

import (
    "context"
    "encoding/base64"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"

    "github.com/lexlapax/go-llms/pkg/llm"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// ImageAnalyzer handles image processing with LLMs
type ImageAnalyzer struct {
    llm llm.LLM
    provider string
}

func NewImageAnalyzer(providerStr string) (*ImageAnalyzer, error) {
    // Parse provider string
    providerType, model, err := parseProviderString(providerStr)
    if err != nil {
        return nil, err
    }

    // Create LLM instance
    var llmInstance llm.LLM
    var options []llm.Option

    switch providerType {
    case "openai":
        options = append(options, provider.WithModel(model))
        llmInstance, err = provider.NewOpenAI(options...)
    case "anthropic":
        options = append(options, provider.WithModel(model))
        llmInstance, err = provider.NewAnthropic(options...)
    case "gemini":
        options = append(options, provider.WithModel(model))
        llmInstance, err = provider.NewGemini(options...)
    default:
        return nil, fmt.Errorf("unsupported provider: %s", providerType)
    }

    if err != nil {
        return nil, err
    }

    return &ImageAnalyzer{
        llm: llmInstance,
        provider: providerType,
    }, nil
}

func (ia *ImageAnalyzer) AnalyzeImage(ctx context.Context, imagePath string, prompt string) (string, error) {
    // Check if provider supports vision
    caps := ia.llm.Capabilities()
    hasVision := false
    for _, cap := range caps {
        if cap == llm.CapabilityVision {
            hasVision = true
            break
        }
    }

    if !hasVision {
        return "", fmt.Errorf("provider %s does not support vision", ia.provider)
    }

    // Read image file
    imageData, err := os.ReadFile(imagePath)
    if err != nil {
        return "", fmt.Errorf("failed to read image: %w", err)
    }

    // Detect MIME type
    mimeType := http.DetectContentType(imageData)
    fmt.Printf("📷 Image: %s (type: %s, size: %d bytes)\n", imagePath, mimeType, len(imageData))

    // Create image message
    message := domain.NewImageMessage(
        domain.RoleUser,
        imageData,
        mimeType,
        prompt,
    )

    // Send to LLM
    messages := []domain.Message{message}
    response, err := ia.llm.Generate(ctx, messages, nil)
    if err != nil {
        return "", fmt.Errorf("generation failed: %w", err)
    }

    return response.Content, nil
}

func (ia *ImageAnalyzer) AnalyzeImageURL(ctx context.Context, imageURL string, prompt string) (string, error) {
    fmt.Printf("🌐 Analyzing image from URL: %s\n", imageURL)

    // Create image URL message
    message := domain.NewImageURLMessage(
        domain.RoleUser,
        imageURL,
        prompt,
    )

    // Send to LLM
    messages := []domain.Message{message}
    response, err := ia.llm.Generate(ctx, messages, nil)
    if err != nil {
        return "", fmt.Errorf("generation failed: %w", err)
    }

    return response.Content, nil
}

func (ia *ImageAnalyzer) CompareImages(ctx context.Context, image1Path, image2Path string, prompt string) (string, error) {
    fmt.Printf("🔍 Comparing two images\n")

    // Read both images
    image1Data, err := os.ReadFile(image1Path)
    if err != nil {
        return "", fmt.Errorf("failed to read image1: %w", err)
    }

    image2Data, err := os.ReadFile(image2Path)
    if err != nil {
        return "", fmt.Errorf("failed to read image2: %w", err)
    }

    // Create message with multiple images
    message := domain.Message{
        Role: domain.RoleUser,
        Content: []domain.ContentPart{
            {
                Type: domain.ContentTypeText,
                Text: prompt,
            },
            {
                Type: domain.ContentTypeImage,
                Image: &domain.ImageContent{
                    Data:     base64.StdEncoding.EncodeToString(image1Data),
                    MIMEType: http.DetectContentType(image1Data),
                },
            },
            {
                Type: domain.ContentTypeImage,
                Image: &domain.ImageContent{
                    Data:     base64.StdEncoding.EncodeToString(image2Data),
                    MIMEType: http.DetectContentType(image2Data),
                },
            },
        },
    }

    // Send to LLM
    messages := []domain.Message{message}
    response, err := ia.llm.Generate(ctx, messages, nil)
    if err != nil {
        return "", fmt.Errorf("generation failed: %w", err)
    }

    return response.Content, nil
}

// Helper function to parse provider string
func parseProviderString(providerStr string) (string, string, error) {
    // Simple parsing - in practice use llmutil
    switch {
    case len(providerStr) > 7 && providerStr[:7] == "openai/":
        return "openai", providerStr[7:], nil
    case len(providerStr) > 10 && providerStr[:10] == "anthropic/":
        return "anthropic", providerStr[10:], nil
    case len(providerStr) > 7 && providerStr[:7] == "gemini/":
        return "gemini", providerStr[7:], nil
    default:
        return "", "", fmt.Errorf("invalid provider string: %s", providerStr)
    }
}

func main() {
    fmt.Println("🖼️ Multimodal Content - Image Analysis")
    fmt.Println("=====================================")

    // Create analyzer with different providers
    providers := []string{
        "openai/gpt-4o",          // OpenAI with vision
        "anthropic/claude-3-5-sonnet", // Anthropic with vision
        "gemini/gemini-2.0-flash-exp",     // Gemini with vision
    }

    ctx := context.Background()

    for _, providerStr := range providers {
        fmt.Printf("\n🔧 Testing with %s\n", providerStr)
        fmt.Println("-------------------")

        analyzer, err := NewImageAnalyzer(providerStr)
        if err != nil {
            log.Printf("Failed to create analyzer: %v", err)
            continue
        }

        // Example 1: Analyze a local image
        fmt.Println("\n📸 Example 1: Analyze Local Image")
        
        // Create a sample image (in practice, use real images)
        sampleImagePath := createSampleImage()
        defer os.Remove(sampleImagePath)

        result, err := analyzer.AnalyzeImage(ctx, sampleImagePath, 
            "What do you see in this image? Describe it in detail.")
        if err != nil {
            log.Printf("Image analysis failed: %v", err)
        } else {
            fmt.Printf("Analysis: %s\n", result)
        }

        // Example 2: Analyze image from URL
        fmt.Println("\n🌐 Example 2: Analyze Image URL")
        
        imageURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/0/05/Go_Logo_Blue.svg/512px-Go_Logo_Blue.svg.png"
        
        result2, err := analyzer.AnalyzeImageURL(ctx, imageURL,
            "What programming language logo is this? What can you tell me about its design?")
        if err != nil {
            log.Printf("URL analysis failed: %v", err)
        } else {
            fmt.Printf("Analysis: %s\n", result2)
        }

        // Example 3: Specific analysis tasks
        fmt.Println("\n🔍 Example 3: Specific Analysis Tasks")
        
        analysisPrompts := []string{
            "Extract any text visible in this image",
            "Identify the main colors used in this image",
            "Describe the composition and layout of elements",
            "What emotions or mood does this image convey?",
        }

        for _, prompt := range analysisPrompts {
            fmt.Printf("\nPrompt: %s\n", prompt)
            result3, err := analyzer.AnalyzeImage(ctx, sampleImagePath, prompt)
            if err != nil {
                log.Printf("Analysis failed: %v", err)
            } else {
                fmt.Printf("Result: %s\n", result3)
            }
            break // Just do one to save API calls
        }
    }
}

// createSampleImage creates a simple test image
func createSampleImage() string {
    // In practice, use real images
    // This creates a minimal valid PNG
    pngData := []byte{
        0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
        // ... minimal PNG data ...
    }
    
    tmpFile, _ := os.CreateTemp("", "sample-*.png")
    tmpFile.Write(pngData)
    tmpFile.Close()
    
    return tmpFile.Name()
}
```

---

## Level 2: Audio and Video Processing
*Handle audio recordings and video content*

### Multimodal Content Processor
```go
package main

import (
    "context"
    "encoding/base64"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strings"

    "github.com/lexlapax/go-llms/pkg/llm"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// MultimodalProcessor handles all content types
type MultimodalProcessor struct {
    llm          llm.LLM
    provider     string
    capabilities map[string]bool
}

type ContentCapabilities struct {
    Vision bool
    Audio  bool
    Video  bool
    Files  bool
}

func NewMultimodalProcessor(providerStr string) (*MultimodalProcessor, error) {
    providerType, model, err := parseProviderString(providerStr)
    if err != nil {
        return nil, err
    }

    // Create LLM
    var llmInstance llm.LLM
    var options []llm.Option

    switch providerType {
    case "gemini":
        options = append(options, provider.WithModel(model))
        llmInstance, err = provider.NewGemini(options...)
    default:
        return nil, fmt.Errorf("provider %s has limited multimodal support", providerType)
    }

    if err != nil {
        return nil, err
    }

    // Check capabilities
    caps := make(map[string]bool)
    for _, cap := range llmInstance.Capabilities() {
        switch cap {
        case llm.CapabilityVision:
            caps["vision"] = true
        case llm.CapabilityAudio:
            caps["audio"] = true
        case llm.CapabilityVideo:
            caps["video"] = true
        }
    }

    return &MultimodalProcessor{
        llm:          llmInstance,
        provider:     providerType,
        capabilities: caps,
    }, nil
}

func (mp *MultimodalProcessor) GetCapabilities() ContentCapabilities {
    return ContentCapabilities{
        Vision: mp.capabilities["vision"],
        Audio:  mp.capabilities["audio"],
        Video:  mp.capabilities["video"],
        Files:  mp.provider == "openai", // OpenAI supports general files
    }
}

func (mp *MultimodalProcessor) ProcessAudio(ctx context.Context, audioPath string, prompt string) (string, error) {
    if !mp.capabilities["audio"] {
        return "", fmt.Errorf("provider %s does not support audio", mp.provider)
    }

    fmt.Printf("🎵 Processing audio: %s\n", audioPath)

    // Read audio file
    audioData, err := os.ReadFile(audioPath)
    if err != nil {
        return "", fmt.Errorf("failed to read audio: %w", err)
    }

    // Detect MIME type
    mimeType := detectAudioMIMEType(audioPath)
    fmt.Printf("Audio type: %s, size: %d bytes\n", mimeType, len(audioData))

    // Create audio message
    message := domain.NewAudioMessage(
        domain.RoleUser,
        audioData,
        mimeType,
        prompt,
    )

    // Process
    messages := []domain.Message{message}
    response, err := mp.llm.Generate(ctx, messages, nil)
    if err != nil {
        return "", fmt.Errorf("audio processing failed: %w", err)
    }

    return response.Content, nil
}

func (mp *MultimodalProcessor) ProcessVideo(ctx context.Context, videoPath string, prompt string) (string, error) {
    if !mp.capabilities["video"] {
        return "", fmt.Errorf("provider %s does not support video", mp.provider)
    }

    fmt.Printf("🎥 Processing video: %s\n", videoPath)

    // Read video file (in practice, you might want to stream or chunk large videos)
    videoData, err := os.ReadFile(videoPath)
    if err != nil {
        return "", fmt.Errorf("failed to read video: %w", err)
    }

    // Check size limits (Gemini has limits on video size)
    maxSize := 10 * 1024 * 1024 // 10MB example limit
    if len(videoData) > maxSize {
        return "", fmt.Errorf("video too large: %d bytes (max: %d)", len(videoData), maxSize)
    }

    // Detect MIME type
    mimeType := detectVideoMIMEType(videoPath)
    fmt.Printf("Video type: %s, size: %d bytes\n", mimeType, len(videoData))

    // Create video message
    message := domain.NewVideoMessage(
        domain.RoleUser,
        videoData,
        mimeType,
        prompt,
    )

    // Process
    messages := []domain.Message{message}
    response, err := mp.llm.Generate(ctx, messages, nil)
    if err != nil {
        return "", fmt.Errorf("video processing failed: %w", err)
    }

    return response.Content, nil
}

func (mp *MultimodalProcessor) ProcessMixedContent(ctx context.Context, contents []MixedContent, prompt string) (string, error) {
    fmt.Printf("🎨 Processing mixed content (%d items)\n", len(contents))

    // Build message with multiple content parts
    contentParts := []domain.ContentPart{
        {
            Type: domain.ContentTypeText,
            Text: prompt,
        },
    }

    for i, content := range contents {
        fmt.Printf("  %d. %s: %s\n", i+1, content.Type, content.Path)

        data, err := os.ReadFile(content.Path)
        if err != nil {
            return "", fmt.Errorf("failed to read %s: %w", content.Path, err)
        }

        encodedData := base64.StdEncoding.EncodeToString(data)

        switch content.Type {
        case "image":
            if !mp.capabilities["vision"] {
                fmt.Printf("  ⚠️  Skipping image (not supported)\n")
                continue
            }
            contentParts = append(contentParts, domain.ContentPart{
                Type: domain.ContentTypeImage,
                Image: &domain.ImageContent{
                    Data:     encodedData,
                    MIMEType: content.MIMEType,
                },
            })

        case "audio":
            if !mp.capabilities["audio"] {
                fmt.Printf("  ⚠️  Skipping audio (not supported)\n")
                continue
            }
            contentParts = append(contentParts, domain.ContentPart{
                Type: domain.ContentTypeAudio,
                Audio: &domain.AudioContent{
                    Data:     encodedData,
                    MIMEType: content.MIMEType,
                },
            })

        case "video":
            if !mp.capabilities["video"] {
                fmt.Printf("  ⚠️  Skipping video (not supported)\n")
                continue
            }
            contentParts = append(contentParts, domain.ContentPart{
                Type: domain.ContentTypeVideo,
                Video: &domain.VideoContent{
                    Data:     encodedData,
                    MIMEType: content.MIMEType,
                },
            })
        }
    }

    message := domain.Message{
        Role:    domain.RoleUser,
        Content: contentParts,
    }

    // Process
    messages := []domain.Message{message}
    response, err := mp.llm.Generate(ctx, messages, nil)
    if err != nil {
        return "", fmt.Errorf("mixed content processing failed: %w", err)
    }

    return response.Content, nil
}

type MixedContent struct {
    Type     string // "image", "audio", "video"
    Path     string
    MIMEType string
}

// MIME type detection helpers
func detectAudioMIMEType(path string) string {
    ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
    switch ext {
    case "mp3":
        return "audio/mpeg"
    case "wav":
        return "audio/wav"
    case "ogg":
        return "audio/ogg"
    case "m4a":
        return "audio/mp4"
    case "flac":
        return "audio/flac"
    default:
        return "audio/mpeg" // Default
    }
}

func detectVideoMIMEType(path string) string {
    ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
    switch ext {
    case "mp4":
        return "video/mp4"
    case "avi":
        return "video/x-msvideo"
    case "mov":
        return "video/quicktime"
    case "webm":
        return "video/webm"
    case "mkv":
        return "video/x-matroska"
    default:
        return "video/mp4" // Default
    }
}

// Demonstration functions
func demonstrateAudioProcessing(processor *MultimodalProcessor, ctx context.Context) {
    fmt.Println("\n🎵 Audio Processing Examples")
    fmt.Println("==========================")

    // Example audio analysis prompts
    audioPrompts := []struct {
        name   string
        prompt string
    }{
        {
            name:   "Transcription",
            prompt: "Transcribe this audio recording word for word.",
        },
        {
            name:   "Summary",
            prompt: "Provide a brief summary of what is being said in this audio.",
        },
        {
            name:   "Speaker Analysis",
            prompt: "Analyze the speaker(s) in this audio. How many speakers? What is their tone?",
        },
        {
            name:   "Content Analysis",
            prompt: "What is the main topic or subject matter of this audio recording?",
        },
    }

    // Create sample audio (in practice, use real audio files)
    audioPath := createSampleAudio()
    defer os.Remove(audioPath)

    for _, example := range audioPrompts {
        fmt.Printf("\n📍 %s:\n", example.name)
        
        result, err := processor.ProcessAudio(ctx, audioPath, example.prompt)
        if err != nil {
            log.Printf("Audio processing failed: %v", err)
            continue
        }
        
        fmt.Printf("Result: %s\n", result)
        break // Process just one to save API calls
    }
}

func demonstrateVideoProcessing(processor *MultimodalProcessor, ctx context.Context) {
    fmt.Println("\n🎥 Video Processing Examples")
    fmt.Println("==========================")

    // Example video analysis prompts
    videoPrompts := []struct {
        name   string
        prompt string
    }{
        {
            name:   "Scene Description",
            prompt: "Describe what happens in this video clip. Include details about the scene, actions, and any text visible.",
        },
        {
            name:   "Content Summary",
            prompt: "Provide a brief summary of this video's content and main message.",
        },
        {
            name:   "Technical Analysis",
            prompt: "Analyze the video production quality, camera work, and editing style.",
        },
        {
            name:   "Object Detection",
            prompt: "List all the objects, people, and elements visible in this video.",
        },
    }

    // Create sample video (in practice, use real video files)
    videoPath := createSampleVideo()
    defer os.Remove(videoPath)

    for _, example := range videoPrompts {
        fmt.Printf("\n📍 %s:\n", example.name)
        
        result, err := processor.ProcessVideo(ctx, videoPath, example.prompt)
        if err != nil {
            log.Printf("Video processing failed: %v", err)
            continue
        }
        
        fmt.Printf("Result: %s\n", result)
        break // Process just one to save API calls
    }
}

func demonstrateMixedContent(processor *MultimodalProcessor, ctx context.Context) {
    fmt.Println("\n🎨 Mixed Content Examples")
    fmt.Println("========================")

    // Create sample content
    imagePath := createSampleImage()
    audioPath := createSampleAudio()
    defer os.Remove(imagePath)
    defer os.Remove(audioPath)

    contents := []MixedContent{
        {
            Type:     "image",
            Path:     imagePath,
            MIMEType: "image/png",
        },
        {
            Type:     "audio",
            Path:     audioPath,
            MIMEType: "audio/mpeg",
        },
    }

    prompts := []string{
        "Analyze both the image and audio. How do they relate to each other?",
        "What story emerges from combining this visual and audio content?",
        "Compare and contrast the mood/tone of the image versus the audio.",
    }

    for i, prompt := range prompts {
        fmt.Printf("\n📍 Example %d:\n", i+1)
        fmt.Printf("Prompt: %s\n", prompt)
        
        result, err := processor.ProcessMixedContent(ctx, contents, prompt)
        if err != nil {
            log.Printf("Mixed content processing failed: %v", err)
            continue
        }
        
        fmt.Printf("Result: %s\n", result)
        break // Process just one to save API calls
    }
}

// Sample content creators (minimal implementations)
func createSampleAudio() string {
    // In practice, use real audio files
    // This creates a minimal valid MP3 header
    mp3Data := []byte{0xFF, 0xFB, 0x90, 0x00} // MP3 header
    
    tmpFile, _ := os.CreateTemp("", "sample-*.mp3")
    tmpFile.Write(mp3Data)
    tmpFile.Close()
    
    return tmpFile.Name()
}

func createSampleVideo() string {
    // In practice, use real video files
    // This creates a minimal valid MP4 header
    mp4Data := []byte{
        0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, // ftyp box
    }
    
    tmpFile, _ := os.CreateTemp("", "sample-*.mp4")
    tmpFile.Write(mp4Data)
    tmpFile.Close()
    
    return tmpFile.Name()
}

func main() {
    fmt.Println("🎬 Multimodal Content - Audio & Video")
    fmt.Println("====================================")

    ctx := context.Background()

    // Gemini supports all modalities
    processor, err := NewMultimodalProcessor("gemini/gemini-2.0-flash-exp")
    if err != nil {
        log.Fatalf("Failed to create processor: %v", err)
    }

    // Show capabilities
    caps := processor.GetCapabilities()
    fmt.Printf("\n📊 Provider Capabilities:\n")
    fmt.Printf("  Vision: %v\n", caps.Vision)
    fmt.Printf("  Audio: %v\n", caps.Audio)
    fmt.Printf("  Video: %v\n", caps.Video)
    fmt.Printf("  Files: %v\n", caps.Files)

    // Demonstrate different content types
    demonstrateAudioProcessing(processor, ctx)
    demonstrateVideoProcessing(processor, ctx)
    demonstrateMixedContent(processor, ctx)
}
```

---

## Level 3: Production Multimodal Systems
*Build robust multimodal applications*

### Enterprise Multimodal Platform
```go
package main

import (
    "context"
    "encoding/base64"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "github.com/google/uuid"
    "github.com/lexlapax/go-llms/pkg/llm"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
)

// MultimodalPlatform provides enterprise-grade multimodal processing
type MultimodalPlatform struct {
    providers      map[string]llm.LLM
    router         *ContentRouter
    cache          *ContentCache
    metrics        *ProcessingMetrics
    maxConcurrency int
}

// ContentRouter intelligently routes content to appropriate providers
type ContentRouter struct {
    rules []RoutingRule
    mu    sync.RWMutex
}

type RoutingRule struct {
    ContentType   string
    PreferredProvider string
    FallbackProviders []string
    MaxSize      int64
    Priority     int
}

// ContentCache caches processing results
type ContentCache struct {
    cache map[string]CachedResult
    mu    sync.RWMutex
    ttl   time.Duration
}

type CachedResult struct {
    Result    string
    Timestamp time.Time
    Provider  string
}

// ProcessingMetrics tracks performance and usage
type ProcessingMetrics struct {
    mu              sync.RWMutex
    totalRequests   int64
    successCount    int64
    failureCount    int64
    avgResponseTime time.Duration
    providerUsage   map[string]int64
    contentTypes    map[string]int64
}

// ProcessingRequest represents a multimodal processing job
type ProcessingRequest struct {
    ID          string
    Type        ProcessingType
    Content     []ContentItem
    Prompt      string
    Options     ProcessingOptions
    Metadata    map[string]interface{}
}

type ProcessingType string

const (
    ProcessingTypeAnalysis     ProcessingType = "analysis"
    ProcessingTypeTranscription ProcessingType = "transcription"
    ProcessingTypeDescription   ProcessingType = "description"
    ProcessingTypeComparison    ProcessingType = "comparison"
    ProcessingTypeExtraction    ProcessingType = "extraction"
)

type ContentItem struct {
    Type     domain.ContentType
    Data     []byte
    URL      string
    MIMEType string
    Metadata map[string]interface{}
}

type ProcessingOptions struct {
    PreferredProvider string
    MaxRetries       int
    Timeout          time.Duration
    CacheResults     bool
    DetailLevel      string // "brief", "standard", "detailed"
    OutputFormat     string // "text", "json", "structured"
}

// ProcessingResult contains the processing output
type ProcessingResult struct {
    ID           string                 `json:"id"`
    Success      bool                   `json:"success"`
    Result       string                 `json:"result,omitempty"`
    Error        string                 `json:"error,omitempty"`
    Provider     string                 `json:"provider"`
    ProcessingTime time.Duration        `json:"processing_time"`
    Metadata     map[string]interface{} `json:"metadata"`
}

func NewMultimodalPlatform() *MultimodalPlatform {
    return &MultimodalPlatform{
        providers: make(map[string]llm.LLM),
        router:    NewContentRouter(),
        cache:     NewContentCache(15 * time.Minute),
        metrics:   &ProcessingMetrics{
            providerUsage: make(map[string]int64),
            contentTypes:  make(map[string]int64),
        },
        maxConcurrency: 5,
    }
}

func NewContentRouter() *ContentRouter {
    return &ContentRouter{
        rules: []RoutingRule{
            // Image routing rules
            {
                ContentType:       "image",
                PreferredProvider: "gemini",
                FallbackProviders: []string{"openai", "anthropic"},
                MaxSize:          20 * 1024 * 1024, // 20MB
                Priority:         1,
            },
            // Audio routing rules
            {
                ContentType:       "audio",
                PreferredProvider: "gemini",
                FallbackProviders: []string{}, // No fallbacks for audio
                MaxSize:          50 * 1024 * 1024, // 50MB
                Priority:         2,
            },
            // Video routing rules
            {
                ContentType:       "video",
                PreferredProvider: "gemini",
                FallbackProviders: []string{}, // No fallbacks for video
                MaxSize:          100 * 1024 * 1024, // 100MB
                Priority:         3,
            },
        },
    }
}

func NewContentCache(ttl time.Duration) *ContentCache {
    return &ContentCache{
        cache: make(map[string]CachedResult),
        ttl:   ttl,
    }
}

func (mp *MultimodalPlatform) RegisterProvider(name string, llmInstance llm.LLM) {
    mp.providers[name] = llmInstance
    log.Printf("✅ Registered provider: %s", name)
}

func (mp *MultimodalPlatform) Process(ctx context.Context, request ProcessingRequest) (*ProcessingResult, error) {
    startTime := time.Now()
    mp.recordMetric("request", "")

    result := &ProcessingResult{
        ID:       request.ID,
        Metadata: make(map[string]interface{}),
    }

    // Check cache if enabled
    if request.Options.CacheResults {
        if cached := mp.checkCache(request); cached != nil {
            result.Success = true
            result.Result = cached.Result
            result.Provider = cached.Provider
            result.ProcessingTime = time.Since(startTime)
            result.Metadata["cached"] = true
            mp.recordMetric("cache_hit", "")
            return result, nil
        }
    }

    // Route to appropriate provider
    provider, providerName := mp.routeContent(request)
    if provider == nil {
        result.Success = false
        result.Error = "No suitable provider found"
        mp.recordMetric("failure", "routing")
        return result, fmt.Errorf("no suitable provider found")
    }

    // Build message
    message, err := mp.buildMessage(request)
    if err != nil {
        result.Success = false
        result.Error = err.Error()
        mp.recordMetric("failure", "message_build")
        return result, err
    }

    // Process with timeout
    processCtx, cancel := context.WithTimeout(ctx, request.Options.Timeout)
    defer cancel()

    // Execute processing
    response, err := provider.Generate(processCtx, []domain.Message{message}, nil)
    if err != nil {
        result.Success = false
        result.Error = err.Error()
        mp.recordMetric("failure", providerName)
        
        // Try fallback providers
        if fallbackResult := mp.tryFallbacks(ctx, request, message); fallbackResult != nil {
            return fallbackResult, nil
        }
        
        return result, err
    }

    // Success
    result.Success = true
    result.Result = response.Content
    result.Provider = providerName
    result.ProcessingTime = time.Since(startTime)
    
    // Cache result if enabled
    if request.Options.CacheResults {
        mp.cacheResult(request, result)
    }

    mp.recordMetric("success", providerName)
    mp.updateMetrics(result.ProcessingTime)

    return result, nil
}

func (mp *MultimodalPlatform) buildMessage(request ProcessingRequest) (domain.Message, error) {
    contentParts := []domain.ContentPart{
        {
            Type: domain.ContentTypeText,
            Text: mp.buildPrompt(request),
        },
    }

    for _, item := range request.Content {
        switch item.Type {
        case domain.ContentTypeImage:
            if item.URL != "" {
                contentParts = append(contentParts, domain.ContentPart{
                    Type: domain.ContentTypeImage,
                    Image: &domain.ImageContent{
                        URL: item.URL,
                    },
                })
            } else {
                contentParts = append(contentParts, domain.ContentPart{
                    Type: domain.ContentTypeImage,
                    Image: &domain.ImageContent{
                        Data:     base64.StdEncoding.EncodeToString(item.Data),
                        MIMEType: item.MIMEType,
                    },
                })
            }

        case domain.ContentTypeAudio:
            contentParts = append(contentParts, domain.ContentPart{
                Type: domain.ContentTypeAudio,
                Audio: &domain.AudioContent{
                    Data:     base64.StdEncoding.EncodeToString(item.Data),
                    MIMEType: item.MIMEType,
                },
            })

        case domain.ContentTypeVideo:
            contentParts = append(contentParts, domain.ContentPart{
                Type: domain.ContentTypeVideo,
                Video: &domain.VideoContent{
                    Data:     base64.StdEncoding.EncodeToString(item.Data),
                    MIMEType: item.MIMEType,
                },
            })
        }
    }

    return domain.Message{
        Role:    domain.RoleUser,
        Content: contentParts,
    }, nil
}

func (mp *MultimodalPlatform) buildPrompt(request ProcessingRequest) string {
    basePrompt := request.Prompt

    // Add detail level instructions
    switch request.Options.DetailLevel {
    case "brief":
        basePrompt += "\n\nProvide a brief, concise response focusing only on key points."
    case "detailed":
        basePrompt += "\n\nProvide a comprehensive, detailed analysis covering all aspects."
    }

    // Add output format instructions
    switch request.Options.OutputFormat {
    case "json":
        basePrompt += "\n\nReturn the response as valid JSON with appropriate structure."
    case "structured":
        basePrompt += "\n\nStructure your response with clear sections and bullet points."
    }

    // Add processing type specific instructions
    switch request.Type {
    case ProcessingTypeTranscription:
        basePrompt += "\n\nProvide an accurate transcription of all spoken content."
    case ProcessingTypeDescription:
        basePrompt += "\n\nDescribe the content in detail, noting all important elements."
    case ProcessingTypeComparison:
        basePrompt += "\n\nCompare and contrast the provided content items systematically."
    case ProcessingTypeExtraction:
        basePrompt += "\n\nExtract and list all requested information clearly."
    }

    return basePrompt
}

func (mp *MultimodalPlatform) routeContent(request ProcessingRequest) (llm.LLM, string) {
    // Determine primary content type
    primaryType := mp.determinePrimaryContentType(request.Content)
    
    // Find applicable routing rule
    mp.router.mu.RLock()
    defer mp.router.mu.RUnlock()

    for _, rule := range mp.router.rules {
        if rule.ContentType == primaryType {
            // Check preferred provider
            if request.Options.PreferredProvider != "" {
                if provider, exists := mp.providers[request.Options.PreferredProvider]; exists {
                    if mp.supportsContent(provider, request.Content) {
                        return provider, request.Options.PreferredProvider
                    }
                }
            }

            // Use rule's preferred provider
            if provider, exists := mp.providers[rule.PreferredProvider]; exists {
                if mp.supportsContent(provider, request.Content) {
                    return provider, rule.PreferredProvider
                }
            }

            // Try fallback providers
            for _, fallback := range rule.FallbackProviders {
                if provider, exists := mp.providers[fallback]; exists {
                    if mp.supportsContent(provider, request.Content) {
                        return provider, fallback
                    }
                }
            }
        }
    }

    return nil, ""
}

func (mp *MultimodalPlatform) determinePrimaryContentType(content []ContentItem) string {
    // Simple heuristic: return the most complex content type
    hasVideo := false
    hasAudio := false
    hasImage := false

    for _, item := range content {
        switch item.Type {
        case domain.ContentTypeVideo:
            hasVideo = true
        case domain.ContentTypeAudio:
            hasAudio = true
        case domain.ContentTypeImage:
            hasImage = true
        }
    }

    if hasVideo {
        return "video"
    }
    if hasAudio {
        return "audio"
    }
    if hasImage {
        return "image"
    }
    
    return "text"
}

func (mp *MultimodalPlatform) supportsContent(provider llm.LLM, content []ContentItem) bool {
    caps := provider.Capabilities()
    capMap := make(map[llm.Capability]bool)
    for _, cap := range caps {
        capMap[cap] = true
    }

    for _, item := range content {
        switch item.Type {
        case domain.ContentTypeImage:
            if !capMap[llm.CapabilityVision] {
                return false
            }
        case domain.ContentTypeAudio:
            if !capMap[llm.CapabilityAudio] {
                return false
            }
        case domain.ContentTypeVideo:
            if !capMap[llm.CapabilityVideo] {
                return false
            }
        }
    }

    return true
}

func (mp *MultimodalPlatform) tryFallbacks(ctx context.Context, request ProcessingRequest, message domain.Message) *ProcessingResult {
    // Implementation of fallback logic
    return nil
}

func (mp *MultimodalPlatform) checkCache(request ProcessingRequest) *CachedResult {
    cacheKey := mp.generateCacheKey(request)
    
    mp.cache.mu.RLock()
    defer mp.cache.mu.RUnlock()

    if cached, exists := mp.cache.cache[cacheKey]; exists {
        if time.Since(cached.Timestamp) < mp.cache.ttl {
            return &cached
        }
    }

    return nil
}

func (mp *MultimodalPlatform) cacheResult(request ProcessingRequest, result *ProcessingResult) {
    cacheKey := mp.generateCacheKey(request)
    
    mp.cache.mu.Lock()
    defer mp.cache.mu.Unlock()

    mp.cache.cache[cacheKey] = CachedResult{
        Result:    result.Result,
        Timestamp: time.Now(),
        Provider:  result.Provider,
    }
}

func (mp *MultimodalPlatform) generateCacheKey(request ProcessingRequest) string {
    // Simple cache key generation
    h := sha256.New()
    h.Write([]byte(request.Prompt))
    for _, item := range request.Content {
        h.Write([]byte(item.Type))
        if item.URL != "" {
            h.Write([]byte(item.URL))
        } else {
            h.Write(item.Data[:min(100, len(item.Data))]) // Use first 100 bytes
        }
    }
    return hex.EncodeToString(h.Sum(nil))
}

func (mp *MultimodalPlatform) recordMetric(metricType, provider string) {
    mp.metrics.mu.Lock()
    defer mp.metrics.mu.Unlock()

    switch metricType {
    case "request":
        mp.metrics.totalRequests++
    case "success":
        mp.metrics.successCount++
        if provider != "" {
            mp.metrics.providerUsage[provider]++
        }
    case "failure":
        mp.metrics.failureCount++
    case "cache_hit":
        // Could track cache metrics separately
    }
}

func (mp *MultimodalPlatform) updateMetrics(responseTime time.Duration) {
    mp.metrics.mu.Lock()
    defer mp.metrics.mu.Unlock()

    // Update average response time
    if mp.metrics.avgResponseTime == 0 {
        mp.metrics.avgResponseTime = responseTime
    } else {
        mp.metrics.avgResponseTime = (mp.metrics.avgResponseTime + responseTime) / 2
    }
}

func (mp *MultimodalPlatform) GetMetrics() map[string]interface{} {
    mp.metrics.mu.RLock()
    defer mp.metrics.mu.RUnlock()

    successRate := float64(0)
    if mp.metrics.totalRequests > 0 {
        successRate = float64(mp.metrics.successCount) / float64(mp.metrics.totalRequests) * 100
    }

    return map[string]interface{}{
        "total_requests":    mp.metrics.totalRequests,
        "success_count":     mp.metrics.successCount,
        "failure_count":     mp.metrics.failureCount,
        "success_rate":      fmt.Sprintf("%.2f%%", successRate),
        "avg_response_time": mp.metrics.avgResponseTime.String(),
        "provider_usage":    mp.metrics.providerUsage,
    }
}

// Example usage scenarios
func demonstrateDocumentAnalysis(platform *MultimodalPlatform, ctx context.Context) {
    fmt.Println("\n📄 Document Analysis Workflow")
    fmt.Println("============================")

    // Simulate scanning a document with images and text
    request := ProcessingRequest{
        ID:   uuid.New().String(),
        Type: ProcessingTypeAnalysis,
        Content: []ContentItem{
            {
                Type:     domain.ContentTypeImage,
                URL:      "https://example.com/document-page1.jpg",
                MIMEType: "image/jpeg",
            },
            {
                Type:     domain.ContentTypeImage,
                URL:      "https://example.com/document-page2.jpg",
                MIMEType: "image/jpeg",
            },
        },
        Prompt: `Analyze this scanned document:
1. Extract all text content
2. Identify document type and purpose
3. List key information and data points
4. Note any signatures, stamps, or special markings
5. Assess document quality and legibility`,
        Options: ProcessingOptions{
            DetailLevel:  "detailed",
            OutputFormat: "structured",
            CacheResults: true,
            Timeout:      30 * time.Second,
        },
    }

    result, err := platform.Process(ctx, request)
    if err != nil {
        log.Printf("Document analysis failed: %v", err)
        return
    }

    fmt.Printf("✅ Analysis Result (Provider: %s):\n%s\n", result.Provider, result.Result)
}

func demonstrateMediaContentModeration(platform *MultimodalPlatform, ctx context.Context) {
    fmt.Println("\n🛡️ Media Content Moderation")
    fmt.Println("==========================")

    // Example: Moderate user-uploaded content
    request := ProcessingRequest{
        ID:   uuid.New().String(),
        Type: ProcessingTypeAnalysis,
        Content: []ContentItem{
            // Would include actual media content
        },
        Prompt: `Analyze this media content for moderation purposes:
1. Identify any inappropriate or harmful content
2. Check for violence, explicit material, or hate speech
3. Verify content authenticity (detect obvious manipulation)
4. Assess content rating (G, PG, PG-13, R)
5. Flag any potential policy violations`,
        Options: ProcessingOptions{
            DetailLevel:  "standard",
            OutputFormat: "json",
            CacheResults: false,
            Timeout:      20 * time.Second,
        },
    }

    // Process and handle results
    fmt.Println("Processing content for moderation...")
}

func demonstrateAccessibilityGeneration(platform *MultimodalPlatform, ctx context.Context) {
    fmt.Println("\n♿ Accessibility Content Generation")
    fmt.Println("==================================")

    // Generate alt text for images
    imageRequest := ProcessingRequest{
        ID:   uuid.New().String(),
        Type: ProcessingTypeDescription,
        Content: []ContentItem{
            {
                Type:     domain.ContentTypeImage,
                URL:      "https://example.com/product-image.jpg",
                MIMEType: "image/jpeg",
            },
        },
        Prompt: `Generate comprehensive alt text for accessibility:
1. Describe the main subject and action
2. Include relevant colors, shapes, and composition
3. Note any text visible in the image
4. Keep it concise but informative
5. Follow WCAG guidelines`,
        Options: ProcessingOptions{
            DetailLevel:  "standard",
            OutputFormat: "text",
            CacheResults: true,
            Timeout:      15 * time.Second,
        },
    }

    // Generate video descriptions
    videoRequest := ProcessingRequest{
        ID:   uuid.New().String(),
        Type: ProcessingTypeDescription,
        Content: []ContentItem{
            // Video content
        },
        Prompt: "Generate audio descriptions for visually impaired users, describing key visual elements and actions.",
        Options: ProcessingOptions{
            DetailLevel:  "detailed",
            OutputFormat: "structured",
            CacheResults: true,
            Timeout:      60 * time.Second,
        },
    }

    fmt.Println("Generating accessibility content...")
}

func main() {
    fmt.Println("🏭 Enterprise Multimodal Platform")
    fmt.Println("================================")

    ctx := context.Background()

    // Create platform
    platform := NewMultimodalPlatform()

    // Register providers
    providers := map[string]string{
        "openai":    "gpt-4o",
        "anthropic": "claude-3-5-sonnet",
        "gemini":    "gemini-2.0-flash-exp",
    }

    for name, model := range providers {
        var llmInstance llm.LLM
        var err error

        switch name {
        case "openai":
            llmInstance, err = provider.NewOpenAI(provider.WithModel(model))
        case "anthropic":
            llmInstance, err = provider.NewAnthropic(provider.WithModel(model))
        case "gemini":
            llmInstance, err = provider.NewGemini(provider.WithModel(model))
        }

        if err != nil {
            log.Printf("Failed to create %s provider: %v", name, err)
            continue
        }

        platform.RegisterProvider(name, llmInstance)
    }

    // Demonstrate different use cases
    demonstrateDocumentAnalysis(platform, ctx)
    demonstrateMediaContentModeration(platform, ctx)
    demonstrateAccessibilityGeneration(platform, ctx)

    // Show platform metrics
    fmt.Printf("\n📊 Platform Metrics:\n")
    metrics := platform.GetMetrics()
    for key, value := range metrics {
        fmt.Printf("  %s: %v\n", key, value)
    }
}
```

## Multimodal Best Practices

### 1. Content Preparation
- **Optimize file sizes** - Compress images/videos appropriately
- **Use appropriate formats** - JPEG for photos, PNG for diagrams
- **Consider resolution** - Balance quality with processing time
- **Validate content** - Check MIME types and file integrity

### 2. Provider Selection
- **Match capabilities** - Use providers that support your content
- **Consider costs** - Multimodal APIs often cost more
- **Plan fallbacks** - Have alternatives for unsupported content
- **Test thoroughly** - Verify provider behavior with your content

### 3. Error Handling
- **Content validation** - Check sizes and formats before sending
- **Graceful degradation** - Handle unsupported content types
- **Retry strategies** - Implement intelligent retry logic
- **User feedback** - Provide clear error messages

### 4. Performance Optimization
- **Batch processing** - Group related content when possible
- **Caching** - Cache results for repeated queries
- **Streaming** - Use streaming for large files
- **Concurrent processing** - Process independent items in parallel

## Common Use Cases

### Image Analysis
- **Document processing** - Extract text and data from scans
- **Product analysis** - Identify objects and attributes
- **Medical imaging** - Analyze diagnostic images
- **Quality control** - Detect defects and anomalies

### Audio Processing
- **Transcription** - Convert speech to text
- **Translation** - Transcribe and translate audio
- **Speaker analysis** - Identify speakers and emotions
- **Content moderation** - Detect inappropriate audio

### Video Understanding
- **Content summarization** - Generate video summaries
- **Scene detection** - Identify key moments
- **Action recognition** - Detect activities and events
- **Surveillance analysis** - Security and monitoring

### Mixed Media
- **Presentations** - Analyze slides with audio
- **Social media** - Process posts with multiple media
- **Educational content** - Understand lectures and tutorials
- **Customer support** - Handle screenshots with descriptions

## Troubleshooting

### Common Issues

**"Unsupported content type" errors**
- Verify provider capabilities
- Check content type constants
- Ensure proper MIME type detection
- Use fallback providers

**Large file handling**
- Implement file size checks
- Use compression when appropriate
- Consider chunking for very large files
- Stream content when possible

**Performance issues**
- Optimize image/video resolution
- Implement caching strategies
- Use appropriate timeouts
- Monitor API rate limits

**Quality concerns**
- Use high-quality source material
- Provide clear, specific prompts
- Test with different providers
- Implement result validation

## Next Steps

🎨 **Multimodal mastery achieved!** Continue with:

- **[Data Validation](data-validation.md)** - Validate multimodal responses
- **[Data Pipelines](data-pipelines.md)** - Build media processing pipelines
- **[Building Automation Agents](building-automation-agents.md)** - Automate media workflows
- **[Production Deployment](../advanced/production-deployment.md)** - Deploy at scale

### Quick Reference

- **[Provider Comparison](../reference/provider-comparison.md)** - Multimodal capabilities matrix
- **[Configuration Reference](../reference/configuration-reference.md)** - Provider settings
- **[Best Practices Checklist](../reference/best-practices-checklist.md)** - Multimodal guidelines

---

**Need help with multimodal content?** Check our [examples](../../cmd/examples/provider-multimodal/) or join the discussion on [GitHub](https://github.com/lexlapax/go-llms/discussions).