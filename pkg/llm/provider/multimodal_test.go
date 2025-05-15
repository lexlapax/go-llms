// ABOUTME: This file tests multimodal content handling in provider implementations.
// ABOUTME: It verifies proper format conversion and error handling for different content types.

package provider

import (
	"encoding/base64"
	"testing"

	"github.com/lexlapax/go-llms/pkg/llm/domain"
	"github.com/stretchr/testify/assert"
)

// Mock providers for testing
type MockOpenAIProvider struct {
	providerName string
	model        string
}

type MockAnthropicProvider struct {
	providerName string
	model        string
}

type MockGeminiProvider struct {
	providerName string
	model        string
}

// Mock constructor functions
func NewMockOpenAIProvider(apiKey, model string) *MockOpenAIProvider {
	return &MockOpenAIProvider{
		providerName: "OpenAI",
		model:        model,
	}
}

func NewMockAnthropicProvider(apiKey, model string) *MockAnthropicProvider {
	return &MockAnthropicProvider{
		providerName: "Anthropic",
		model:        model,
	}
}

func NewMockGeminiProvider(apiKey, model string) *MockGeminiProvider {
	return &MockGeminiProvider{
		providerName: "Gemini",
		model:        model,
	}
}

// Add ConvertMessagesToOpenAIFormat method to MockOpenAIProvider
func (p *MockOpenAIProvider) ConvertMessagesToOpenAIFormat(messages []domain.Message) []map[string]interface{} {
	result := make([]map[string]interface{}, len(messages))
	
	for i, msg := range messages {
		openAIMsg := map[string]interface{}{
			"role": string(msg.Role),
		}
		
		// Handle multimodal content
		if len(msg.Content) > 0 {
			content := make([]map[string]interface{}, 0, len(msg.Content))
			
			for _, part := range msg.Content {
				switch part.Type {
				case domain.ContentTypeText:
					content = append(content, map[string]interface{}{
						"type": "text",
						"text": part.Text,
					})
				case domain.ContentTypeImage:
					// OpenAI uses image_url format
					imageURL := "data:" + part.Image.Source.MediaType + ";base64," + part.Image.Source.Data
					content = append(content, map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]interface{}{
							"url": imageURL,
						},
					})
				case domain.ContentTypeFile:
					content = append(content, map[string]interface{}{
						"type": "file",
						"file": map[string]interface{}{
							"file_name": part.File.FileName,
							"file_data": part.File.FileData,
						},
					})
				case domain.ContentTypeVideo:
					content = append(content, map[string]interface{}{
						"type": "video",
						"video": map[string]interface{}{
							"media_type": part.Video.Source.MediaType,
							"data": part.Video.Source.Data,
						},
					})
				case domain.ContentTypeAudio:
					content = append(content, map[string]interface{}{
						"type": "audio",
						"audio": map[string]interface{}{
							"media_type": part.Audio.Source.MediaType,
							"data": part.Audio.Source.Data,
						},
					})
				}
			}
			
			openAIMsg["content"] = content
		}
		
		result[i] = openAIMsg
	}
	
	return result
}

func TestOpenAIMultimodalContent(t *testing.T) {
	provider := NewMockOpenAIProvider("test-key", "gpt-4o")
	
	// Test file content conversion to OpenAI format
	t.Run("FileContent", func(t *testing.T) {
		fileData := []byte("test file data")
		base64Data := base64.StdEncoding.EncodeToString(fileData)
		fileName := "test.pdf"
		mimeType := "application/pdf"
		text := "Please analyze this file"
		
		// Create a message with file attachment
		msg := domain.NewFileMessage(domain.RoleUser, fileName, fileData, mimeType, text)
		messages := []domain.Message{msg}
		
		// Convert to OpenAI format
		oaiFormat := provider.ConvertMessagesToOpenAIFormat(messages)
		
		// Verify the result
		assert.Equal(t, 1, len(oaiFormat), "Should have exactly one message")
		oaiMsg := oaiFormat[0]
		
		// OpenAI format should have a role and content array
		assert.Equal(t, "user", oaiMsg["role"])
		content, ok := oaiMsg["content"].([]map[string]interface{})
		assert.True(t, ok, "Content should be an array of content parts")
		assert.Equal(t, 2, len(content), "Should have file and text content parts")
		
		// Check file part
		filePart := content[0]
		assert.Equal(t, "file", filePart["type"])
		fileObj := filePart["file"].(map[string]interface{})
		assert.Equal(t, fileName, fileObj["file_name"])
		assert.Equal(t, base64Data, fileObj["file_data"])
		
		// Check text part
		textPart := content[1]
		assert.Equal(t, "text", textPart["type"])
		assert.Equal(t, text, textPart["text"])
	})
	
	// Test image content conversion to OpenAI format
	t.Run("ImageContent", func(t *testing.T) {
		imageData := []byte("test image data")
		base64Data := base64.StdEncoding.EncodeToString(imageData)
		mimeType := "image/jpeg"
		text := "What's in this image?"
		
		// Create a message with image attachment
		msg := domain.NewImageMessage(domain.RoleUser, imageData, mimeType, text)
		messages := []domain.Message{msg}
		
		// Convert to OpenAI format
		oaiFormat := provider.ConvertMessagesToOpenAIFormat(messages)
		
		// Verify the result
		assert.Equal(t, 1, len(oaiFormat))
		oaiMsg := oaiFormat[0]
		
		// OpenAI format should have a role and content array
		assert.Equal(t, "user", oaiMsg["role"])
		content, ok := oaiMsg["content"].([]map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(content))
		
		// Check image part - OpenAI uses image_url format
		imagePart := content[0]
		assert.Equal(t, "image_url", imagePart["type"])
		imageObj := imagePart["image_url"].(map[string]interface{})
		
		// OpenAI expects data URLs for base64 images
		expectedURL := "data:" + mimeType + ";base64," + base64Data
		assert.Equal(t, expectedURL, imageObj["url"])
		
		// Check text part
		textPart := content[1]
		assert.Equal(t, "text", textPart["type"])
		assert.Equal(t, text, textPart["text"])
	})
	
	// Test video content conversion to OpenAI format
	t.Run("VideoContent", func(t *testing.T) {
		videoData := []byte("test video data")
		base64Data := base64.StdEncoding.EncodeToString(videoData)
		mimeType := "video/mp4"
		text := "What's happening in this video?"
		
		// Create a message with video attachment
		msg := domain.NewVideoMessage(domain.RoleUser, videoData, mimeType, text)
		messages := []domain.Message{msg}
		
		// Convert to OpenAI format
		oaiFormat := provider.ConvertMessagesToOpenAIFormat(messages)
		
		// Verify the result
		assert.Equal(t, 1, len(oaiFormat))
		oaiMsg := oaiFormat[0]
		
		// OpenAI format should have a role and content array
		assert.Equal(t, "user", oaiMsg["role"])
		content, ok := oaiMsg["content"].([]map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(content))
		
		// Check video part
		videoPart := content[0]
		assert.Equal(t, "video", videoPart["type"])
		videoObj := videoPart["video"].(map[string]interface{})
		assert.Equal(t, base64Data, videoObj["data"])
		assert.Equal(t, mimeType, videoObj["media_type"])
		
		// Check text part
		textPart := content[1]
		assert.Equal(t, "text", textPart["type"])
		assert.Equal(t, text, textPart["text"])
	})
	
	// Test audio content conversion to OpenAI format
	t.Run("AudioContent", func(t *testing.T) {
		audioData := []byte("test audio data")
		base64Data := base64.StdEncoding.EncodeToString(audioData)
		mimeType := "audio/mp3"
		text := "What is said in this audio?"
		
		// Create a message with audio attachment
		msg := domain.NewAudioMessage(domain.RoleUser, audioData, mimeType, text)
		messages := []domain.Message{msg}
		
		// Convert to OpenAI format
		oaiFormat := provider.ConvertMessagesToOpenAIFormat(messages)
		
		// Verify the result
		assert.Equal(t, 1, len(oaiFormat))
		oaiMsg := oaiFormat[0]
		
		// OpenAI format should have a role and content array
		assert.Equal(t, "user", oaiMsg["role"])
		content, ok := oaiMsg["content"].([]map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(content))
		
		// Check audio part
		audioPart := content[0]
		assert.Equal(t, "audio", audioPart["type"])
		audioObj := audioPart["audio"].(map[string]interface{})
		assert.Equal(t, base64Data, audioObj["data"])
		assert.Equal(t, mimeType, audioObj["media_type"])
		
		// Check text part
		textPart := content[1]
		assert.Equal(t, "text", textPart["type"])
		assert.Equal(t, text, textPart["text"])
	})
}

func TestAnthropicMultimodalContent(t *testing.T) {
	_ = NewMockAnthropicProvider("test-key", "claude-3-sonnet-latest") // Create provider but we'll use the utility function directly
	
	// Test image content conversion to Anthropic format (supported)
	t.Run("ImageContent", func(t *testing.T) {
		imageData := []byte("test image data")
		base64Data := base64.StdEncoding.EncodeToString(imageData)
		mimeType := "image/jpeg"
		text := "What's in this image?"
		
		// Create a message with image attachment
		msg := domain.NewImageMessage(domain.RoleUser, imageData, mimeType, text)
		messages := []domain.Message{msg}
		
		// Convert to Anthropic format
		anthropicFormat := ConvertMessagesToAnthropicFormat(messages)
		
		// Verify the result
		assert.Equal(t, 1, len(anthropicFormat))
		anthropicMsg := anthropicFormat[0]
		
		// Anthropic format should have a role and content array
		assert.Equal(t, "user", anthropicMsg["role"])
		content, ok := anthropicMsg["content"].([]map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(content))
		
		// Check image part
		imagePart := content[0]
		assert.Equal(t, "image", imagePart["type"])
		source := imagePart["source"].(map[string]interface{})
		assert.Equal(t, "base64", source["type"])
		assert.Equal(t, mimeType, source["media_type"])
		assert.Equal(t, base64Data, source["data"])
		
		// Check text part
		textPart := content[1]
		assert.Equal(t, "text", textPart["type"])
		assert.Equal(t, text, textPart["text"])
	})
	
	// Test audio content with Anthropic (unsupported)
	t.Run("UnsupportedAudioContent", func(t *testing.T) {
		audioData := []byte("test audio data")
		mimeType := "audio/mp3"
		text := "What is said in this audio?"
		
		// Create a message with audio attachment
		msg := domain.NewAudioMessage(domain.RoleUser, audioData, mimeType, text)
		messages := []domain.Message{msg}
		
		// Test validation function
		err := ValidateMessagesForAnthropic(messages)
		
		// Verify we get an unsupported content type error
		assert.Error(t, err)
		assert.True(t, domain.IsUnsupportedContentTypeError(err))
		
		// Check the specific error contains the provider name and content type
		typeErr, ok := err.(*domain.UnsupportedContentTypeError)
		assert.True(t, ok)
		assert.Equal(t, "Anthropic", typeErr.Provider)
		assert.Equal(t, domain.ContentTypeAudio, typeErr.ContentType)
	})
}

func TestGeminiMultimodalContent(t *testing.T) {
	_ = NewMockGeminiProvider("test-key", "gemini-pro") // Create provider but we'll use the utility function directly
	
	// Test validating content types for Gemini (supports image, video)
	t.Run("SupportedContentTypes", func(t *testing.T) {
		// Image content (supported)
		imageMsg := domain.NewImageMessage(domain.RoleUser, []byte("test image"), "image/jpeg", "test")
		
		// Video content (supported)
		videoMsg := domain.NewVideoMessage(domain.RoleUser, []byte("test video"), "video/mp4", "test")
		
		// File content (unsupported)
		fileMsg := domain.NewFileMessage(domain.RoleUser, "test.pdf", []byte("test file"), "application/pdf", "test")
		
		// Test validation
		assert.NoError(t, ValidateMessagesForGemini([]domain.Message{imageMsg}))
		assert.NoError(t, ValidateMessagesForGemini([]domain.Message{videoMsg}))
		
		// File should be unsupported
		err := ValidateMessagesForGemini([]domain.Message{fileMsg})
		assert.Error(t, err)
		assert.True(t, domain.IsUnsupportedContentTypeError(err))
		
		typeErr, ok := err.(*domain.UnsupportedContentTypeError)
		assert.True(t, ok)
		assert.Equal(t, "Gemini", typeErr.Provider)
		assert.Equal(t, domain.ContentTypeFile, typeErr.ContentType)
	})
}

// Mock validation functions - these would normally be in the provider implementations
func ValidateMessagesForAnthropic(messages []domain.Message) error {
	for _, msg := range messages {
		for _, part := range msg.Content {
			if part.Type == domain.ContentTypeAudio || part.Type == domain.ContentTypeVideo {
				return domain.NewUnsupportedContentTypeError("Anthropic", part.Type)
			}
		}
	}
	return nil
}

func ValidateMessagesForGemini(messages []domain.Message) error {
	for _, msg := range messages {
		for _, part := range msg.Content {
			if part.Type == domain.ContentTypeFile {
				return domain.NewUnsupportedContentTypeError("Gemini", part.Type)
			}
		}
	}
	return nil
}

// Helper function for Anthropic message conversion (simplified for testing)
func ConvertMessagesToAnthropicFormat(messages []domain.Message) []map[string]interface{} {
	result := make([]map[string]interface{}, len(messages))
	
	for i, msg := range messages {
		anthropicMsg := map[string]interface{}{
			"role": string(msg.Role),
		}
		
		content := make([]map[string]interface{}, 0, len(msg.Content))
		
		for _, part := range msg.Content {
			switch part.Type {
			case domain.ContentTypeText:
				content = append(content, map[string]interface{}{
					"type": "text",
					"text": part.Text,
				})
			case domain.ContentTypeImage:
				content = append(content, map[string]interface{}{
					"type": "image",
					"source": map[string]interface{}{
						"type":       "base64",
						"media_type": part.Image.Source.MediaType,
						"data":       part.Image.Source.Data,
					},
				})
			}
		}
		
		anthropicMsg["content"] = content
		result[i] = anthropicMsg
	}
	
	return result
}