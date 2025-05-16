# Multimodal Example

This example demonstrates the multimodal capabilities of the go-llms library, showcasing how to work with different content types including text, images, audio, video, and mixed content.

## Features

- Support for multiple LLM providers (OpenAI, Anthropic, Gemini)
- Multiple content modes: text, image, audio, video, and mixed
- Automatic MIME type detection
- Provider compatibility checking
- Multiple file attachments support

## Prerequisites

Set up your API keys as environment variables:
```bash
export OPENAI_API_KEY="your-api-key"
export ANTHROPIC_API_KEY="your-api-key"
export GEMINI_API_KEY="your-api-key"
```

## Usage

### Text Mode
```bash
# Simple text message
go run main.go -provider openai -mode text -text "What is the capital of France?"

# Using a specific model
go run main.go -provider anthropic -mode text -text "Explain quantum physics" -model claude-3-opus-20240229
```

### Image Mode
```bash
# Single image analysis
go run main.go -provider gemini -mode image -a photo.jpg

# Multiple images
go run main.go -provider openai -mode image -a image1.png -a image2.png
```

### Audio Mode
```bash
# Audio transcription (Gemini only)
go run main.go -provider gemini -mode audio -a speech.mp3
```

### Video Mode
```bash
# Video analysis (Gemini only)
go run main.go -provider gemini -mode video -a video.mp4
```

### Mixed Mode
```bash
# Text with images
go run main.go -provider anthropic -mode mixed -text "Compare these images" -a img1.jpg -a img2.jpg

# Text with multiple content types (Gemini only)
go run main.go -provider gemini -mode mixed -text "Analyze this multimedia content" -a image.jpg -a audio.mp3 -a video.mp4
```

## Provider Compatibility

| Provider  | Text | Images | Audio | Video | Mixed |
|-----------|------|--------|-------|-------|-------|
| OpenAI    | ✅   | ✅     | ❌    | ❌    | ✅*   |
| Anthropic | ✅   | ✅     | ❌    | ❌    | ✅*   |
| Gemini    | ✅   | ✅     | ✅    | ✅    | ✅    |

*Mixed mode for OpenAI and Anthropic only supports text + images

## Command-Line Options

- `-provider`: LLM provider to use (openai, anthropic, gemini)
- `-mode`: Content mode (text, image, audio, video, mixed)
- `-text`: Text content for the message
- `-model`: Specific model to use (optional, uses sensible defaults)
- `-a`: Attach file(s) to the message (can be used multiple times)

## Example Files

The example automatically detects MIME types based on file extensions:
- Images: .jpg, .jpeg, .png, .gif
- Audio: .mp3, .wav
- Video: .mp4, .avi, .mov
- Documents: .pdf

## Error Handling

The example includes comprehensive error handling:
- Validates required flags for each mode
- Checks provider compatibility with content types
- Provides clear error messages for unsupported operations

## Building

```bash
# From the project root
make build-example EXAMPLE=multimodal

# Or directly
go build -o multimodal cmd/examples/multimodal/main.go
```

## Testing

```bash
# Run the example tests
go test ./cmd/examples/multimodal/...
```