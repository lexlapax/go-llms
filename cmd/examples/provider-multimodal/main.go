// ABOUTME: Main entry point for the multimodal example showcasing all content types
// ABOUTME: Demonstrates text, image, audio, video and mixed mode interactions with LLM providers

package main

// ABOUTME: Example demonstrating multimodal content support with images and files
// ABOUTME: Shows how to send images and other media to LLM providers

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/lexlapax/go-llms/pkg/llm/provider"
)

// arrayFlags allows multiple -a flags to be provided
type arrayFlags []string

func (a *arrayFlags) String() string {
	return strings.Join(*a, ", ")
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func main() {
	// Define command-line flags
	providerFlag := flag.String("provider", "openai", "LLM provider to use (openai, anthropic, gemini)")
	modeFlag := flag.String("mode", "text", "Content mode (text, image, audio, video, mixed)")
	textFlag := flag.String("text", "", "Text content for the message")
	modelFlag := flag.String("model", "", "Model to use (provider-specific)")

	var attachments arrayFlags
	flag.Var(&attachments, "a", "Attach file(s) to the message (can be used multiple times)")

	flag.Parse()

	// Validate inputs
	if *modeFlag == "text" && *textFlag == "" {
		log.Fatal("Text mode requires -text flag")
	}

	if *modeFlag != "text" && len(attachments) == 0 {
		log.Fatal("Non-text modes require at least one attachment via -a flag")
	}

	if *modeFlag == "mixed" && (*textFlag == "" || len(attachments) == 0) {
		log.Fatal("Mixed mode requires both -text flag and at least one -a flag")
	}

	// Create provider
	llmProvider, err := createProvider(*providerFlag, *modelFlag)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Check provider support for content types
	mimeTypes := []string{}
	for _, path := range attachments {
		mimeTypes = append(mimeTypes, getMimeType(path))
	}

	if err := providerSupportsContent(*providerFlag, *modeFlag, mimeTypes); err != nil {
		log.Fatalf("Provider compatibility error: %v", err)
	}

	// Create message based on mode
	var message domain.Message
	switch *modeFlag {
	case "text":
		message = createTextMessage(*textFlag)
	case "image":
		message, err = createImageMessage(attachments)
	case "audio":
		message, err = createAudioMessage(attachments[0])
	case "video":
		message, err = createVideoMessage(attachments[0])
	case "mixed":
		message, err = createMixedMessage(*textFlag, attachments)
	default:
		log.Fatalf("Unknown mode: %s", *modeFlag)
	}

	if err != nil {
		log.Fatalf("Failed to create message: %v", err)
	}

	// Send message to provider
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Printf("Sending %s message to %s...\n", *modeFlag, *providerFlag)

	startTime := time.Now()
	response, err := llmProvider.GenerateMessage(ctx, []domain.Message{message})
	if err != nil {
		log.Fatalf("Failed to generate response: %v", err)
	}

	fmt.Printf("\nResponse (%.2fs):\n%s\n", time.Since(startTime).Seconds(), response.Content)
}

func createProvider(providerName, modelName string) (domain.Provider, error) {
	switch providerName {
	case "openai":
		if modelName == "" {
			modelName = "gpt-4o-mini"
		}
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
		}
		return provider.NewOpenAIProvider(apiKey, modelName), nil
	case "anthropic":
		if modelName == "" {
			modelName = "claude-3-haiku-20240307"
		}
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
		}
		return provider.NewAnthropicProvider(apiKey, modelName), nil
	case "gemini":
		if modelName == "" {
			modelName = "gemini-1.5-flash"
		}
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
		}
		return provider.NewGeminiProvider(apiKey, modelName), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}
}

// providerSupportsContent checks if a provider supports a specific content type in a given mode
func providerSupportsContent(providerName, mode string, mimeTypes []string) error {
	switch providerName {
	case "openai":
		for _, mimeType := range mimeTypes {
			if strings.HasPrefix(mimeType, "audio/") && mode != "mixed" {
				// OpenAI only supports audio in speech-to-text models, not in chat models
				return fmt.Errorf("OpenAI chat models don't support standalone audio inputs")
			}
			if strings.HasPrefix(mimeType, "video/") {
				return fmt.Errorf("OpenAI doesn't support video inputs")
			}
		}
	case "anthropic":
		for _, mimeType := range mimeTypes {
			if strings.HasPrefix(mimeType, "audio/") {
				return fmt.Errorf("anthropic doesn't support audio inputs")
			}
			if strings.HasPrefix(mimeType, "video/") {
				return fmt.Errorf("anthropic doesn't support video inputs")
			}
		}
	case "gemini":
		// Gemini supports all multimodal content types
		return nil
	}
	return nil
}

func createTextMessage(text string) domain.Message {
	return domain.NewTextMessage(domain.RoleUser, text)
}

func createImageMessage(paths []string) (domain.Message, error) {
	if len(paths) == 0 {
		return domain.Message{}, fmt.Errorf("no image paths provided")
	}

	// For multiple images, we'll create a mixed message
	if len(paths) > 1 {
		parts := []domain.ContentPart{}
		parts = append(parts, domain.ContentPart{
			Type: domain.ContentTypeText,
			Text: "Please analyze these images:",
		})

		for _, path := range paths {
			data, err := readFile(path)
			if err != nil {
				return domain.Message{}, err
			}

			mimeType := getMimeType(path)
			parts = append(parts, domain.ContentPart{
				Type: domain.ContentTypeImage,
				Image: &domain.ImageContent{
					Source: domain.SourceInfo{
						Type:      domain.SourceTypeBase64,
						MediaType: mimeType,
						Data:      base64.StdEncoding.EncodeToString(data),
					},
				},
			})
		}

		return domain.Message{
			Role:    domain.RoleUser,
			Content: parts,
		}, nil
	}

	// Single image
	data, err := readFile(paths[0])
	if err != nil {
		return domain.Message{}, err
	}

	mimeType := getMimeType(paths[0])
	return domain.Message{
		Role: domain.RoleUser,
		Content: []domain.ContentPart{
			{
				Type: domain.ContentTypeText,
				Text: "What's in this image?",
			},
			{
				Type: domain.ContentTypeImage,
				Image: &domain.ImageContent{
					Source: domain.SourceInfo{
						Type:      domain.SourceTypeBase64,
						MediaType: mimeType,
						Data:      base64.StdEncoding.EncodeToString(data),
					},
				},
			},
		},
	}, nil
}

func createAudioMessage(path string) (domain.Message, error) {
	data, err := readFile(path)
	if err != nil {
		return domain.Message{}, err
	}

	mimeType := getMimeType(path)
	return domain.Message{
		Role: domain.RoleUser,
		Content: []domain.ContentPart{
			{
				Type: domain.ContentTypeText,
				Text: "Please transcribe this audio",
			},
			{
				Type: domain.ContentTypeAudio,
				Audio: &domain.AudioContent{
					Source: domain.SourceInfo{
						Type:      domain.SourceTypeBase64,
						MediaType: mimeType,
						Data:      base64.StdEncoding.EncodeToString(data),
					},
				},
			},
		},
	}, nil
}

func createVideoMessage(path string) (domain.Message, error) {
	data, err := readFile(path)
	if err != nil {
		return domain.Message{}, err
	}

	mimeType := getMimeType(path)
	return domain.Message{
		Role: domain.RoleUser,
		Content: []domain.ContentPart{
			{
				Type: domain.ContentTypeText,
				Text: "Please describe what happens in this video",
			},
			{
				Type: domain.ContentTypeVideo,
				Video: &domain.VideoContent{
					Source: domain.SourceInfo{
						Type:      domain.SourceTypeBase64,
						MediaType: mimeType,
						Data:      base64.StdEncoding.EncodeToString(data),
					},
				},
			},
		},
	}, nil
}

func createMixedMessage(text string, paths []string) (domain.Message, error) {
	parts := []domain.ContentPart{
		{
			Type: domain.ContentTypeText,
			Text: text,
		},
	}

	for _, path := range paths {
		data, err := readFile(path)
		if err != nil {
			return domain.Message{}, err
		}

		mimeType := getMimeType(path)

		// Add appropriate content part based on MIME type
		if strings.HasPrefix(mimeType, "image/") {
			parts = append(parts, domain.ContentPart{
				Type: domain.ContentTypeImage,
				Image: &domain.ImageContent{
					Source: domain.SourceInfo{
						Type:      domain.SourceTypeBase64,
						MediaType: mimeType,
						Data:      base64.StdEncoding.EncodeToString(data),
					},
				},
			})
		} else if strings.HasPrefix(mimeType, "audio/") {
			parts = append(parts, domain.ContentPart{
				Type: domain.ContentTypeAudio,
				Audio: &domain.AudioContent{
					Source: domain.SourceInfo{
						Type:      domain.SourceTypeBase64,
						MediaType: mimeType,
						Data:      base64.StdEncoding.EncodeToString(data),
					},
				},
			})
		} else if strings.HasPrefix(mimeType, "video/") {
			parts = append(parts, domain.ContentPart{
				Type: domain.ContentTypeVideo,
				Video: &domain.VideoContent{
					Source: domain.SourceInfo{
						Type:      domain.SourceTypeBase64,
						MediaType: mimeType,
						Data:      base64.StdEncoding.EncodeToString(data),
					},
				},
			})
		} else {
			parts = append(parts, domain.ContentPart{
				Type: domain.ContentTypeFile,
				File: &domain.FileContent{
					FileName: filepath.Base(path),
					FileData: base64.StdEncoding.EncodeToString(data),
					MimeType: mimeType,
				},
			})
		}
	}

	return domain.Message{
		Role:    domain.RoleUser,
		Content: parts,
	}, nil
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return data, nil
}

// staticMimeTypes provides OS-agnostic MIME type detection
var staticMimeTypes = map[string]string{
	// Images
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".webp": "image/webp",
	".ico":  "image/vnd.microsoft.icon",
	".svg":  "image/svg+xml",
	".tiff": "image/tiff",
	".tif":  "image/tiff",

	// Audio
	".mp3":  "audio/mpeg",
	".wav":  "audio/vnd.wave",
	".ogg":  "audio/ogg",
	".aac":  "audio/aac",
	".flac": "audio/flac",
	".m4a":  "audio/mp4",
	".wma":  "audio/x-ms-wma",

	// Video
	".mp4":  "video/mp4",
	".avi":  "video/vnd.avi",
	".mov":  "video/quicktime",
	".wmv":  "video/x-ms-wmv",
	".flv":  "video/x-flv",
	".webm": "video/webm",
	".mkv":  "video/x-matroska",
	".m4v":  "video/mp4",

	// Documents
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".txt":  "text/plain",
	".rtf":  "application/rtf",

	// Archives
	".zip": "application/zip",
	".rar": "application/vnd.rar",
	".tar": "application/x-tar",
	".gz":  "application/gzip",
	".7z":  "application/x-7z-compressed",

	// Web
	".html": "text/html",
	".htm":  "text/html",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".xml":  "application/xml",
}

func getMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	// Use static registry for consistent OS-agnostic results
	if mimeType, exists := staticMimeTypes[ext]; exists {
		return mimeType
	}

	// Fallback to default for unknown extensions
	return "application/octet-stream"
}
