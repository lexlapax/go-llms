// ABOUTME: This file contains tests for the message domain models.
// ABOUTME: It verifies the creation and functionality of multimodal messages.

package domain

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTextMessage(t *testing.T) {
	// Test creating a simple text message
	msg := NewTextMessage(RoleUser, "Hello, world!")

	assert.Equal(t, RoleUser, msg.Role)
	assert.Equal(t, 1, len(msg.Content))
	assert.Equal(t, ContentTypeText, msg.Content[0].Type)
	assert.Equal(t, "Hello, world!", msg.Content[0].Text)
}

func TestNewImageMessage(t *testing.T) {
	// Test creating an image message
	imageData := []byte("fake image data")
	base64Data := base64.StdEncoding.EncodeToString(imageData)
	mimeType := "image/jpeg"
	text := "This is an image"

	msg := NewImageMessage(RoleUser, imageData, mimeType, text)

	assert.Equal(t, RoleUser, msg.Role)
	assert.Equal(t, 2, len(msg.Content))

	// Check image content
	assert.Equal(t, ContentTypeImage, msg.Content[0].Type)
	assert.NotNil(t, msg.Content[0].Image)
	assert.Equal(t, SourceTypeBase64, msg.Content[0].Image.Source.Type)
	assert.Equal(t, mimeType, msg.Content[0].Image.Source.MediaType)
	assert.Equal(t, base64Data, msg.Content[0].Image.Source.Data)

	// Check text content
	assert.Equal(t, ContentTypeText, msg.Content[1].Type)
	assert.Equal(t, text, msg.Content[1].Text)
}

func TestNewFileMessage(t *testing.T) {
	// Test creating a file message
	fileData := []byte("fake file data")
	base64Data := base64.StdEncoding.EncodeToString(fileData)
	fileName := "test.pdf"
	mimeType := "application/pdf"
	text := "This is a file"

	msg := NewFileMessage(RoleUser, fileName, fileData, mimeType, text)

	assert.Equal(t, RoleUser, msg.Role)
	assert.Equal(t, 2, len(msg.Content))

	// Check file content
	assert.Equal(t, ContentTypeFile, msg.Content[0].Type)
	assert.NotNil(t, msg.Content[0].File)
	assert.Equal(t, fileName, msg.Content[0].File.FileName)
	assert.Equal(t, mimeType, msg.Content[0].File.MimeType)
	assert.Equal(t, base64Data, msg.Content[0].File.FileData)

	// Check text content
	assert.Equal(t, ContentTypeText, msg.Content[1].Type)
	assert.Equal(t, text, msg.Content[1].Text)
}

func TestNewVideoMessage(t *testing.T) {
	// Test creating a video message
	videoData := []byte("fake video data")
	base64Data := base64.StdEncoding.EncodeToString(videoData)
	mimeType := "video/mp4"
	text := "This is a video"

	msg := NewVideoMessage(RoleUser, videoData, mimeType, text)

	assert.Equal(t, RoleUser, msg.Role)
	assert.Equal(t, 2, len(msg.Content))

	// Check video content
	assert.Equal(t, ContentTypeVideo, msg.Content[0].Type)
	assert.NotNil(t, msg.Content[0].Video)
	assert.Equal(t, SourceTypeBase64, msg.Content[0].Video.Source.Type)
	assert.Equal(t, mimeType, msg.Content[0].Video.Source.MediaType)
	assert.Equal(t, base64Data, msg.Content[0].Video.Source.Data)

	// Check text content
	assert.Equal(t, ContentTypeText, msg.Content[1].Type)
	assert.Equal(t, text, msg.Content[1].Text)
}

func TestNewAudioMessage(t *testing.T) {
	// Test creating an audio message
	audioData := []byte("fake audio data")
	base64Data := base64.StdEncoding.EncodeToString(audioData)
	mimeType := "audio/mp3"
	text := "This is audio"

	msg := NewAudioMessage(RoleUser, audioData, mimeType, text)

	assert.Equal(t, RoleUser, msg.Role)
	assert.Equal(t, 2, len(msg.Content))

	// Check audio content
	assert.Equal(t, ContentTypeAudio, msg.Content[0].Type)
	assert.NotNil(t, msg.Content[0].Audio)
	assert.Equal(t, SourceTypeBase64, msg.Content[0].Audio.Source.Type)
	assert.Equal(t, mimeType, msg.Content[0].Audio.Source.MediaType)
	assert.Equal(t, base64Data, msg.Content[0].Audio.Source.Data)

	// Check text content
	assert.Equal(t, ContentTypeText, msg.Content[1].Type)
	assert.Equal(t, text, msg.Content[1].Text)
}

func TestContentTypeString(t *testing.T) {
	// Test ContentType string conversions
	assert.Equal(t, "text", string(ContentTypeText))
	assert.Equal(t, "image", string(ContentTypeImage))
	assert.Equal(t, "file", string(ContentTypeFile))
	assert.Equal(t, "video", string(ContentTypeVideo))
	assert.Equal(t, "audio", string(ContentTypeAudio))
}

func TestSourceTypeString(t *testing.T) {
	// Test SourceType string conversions
	assert.Equal(t, "base64", string(SourceTypeBase64))
	assert.Equal(t, "url", string(SourceTypeURL))
}

func TestURLBasedMessage(t *testing.T) {
	// Test creating a message with URL-based content
	imageURL := "https://example.com/image.jpg"
	text := "This is an image URL"

	msg := NewImageURLMessage(RoleUser, imageURL, text)

	assert.Equal(t, RoleUser, msg.Role)
	assert.Equal(t, 2, len(msg.Content))

	// Check image content
	assert.Equal(t, ContentTypeImage, msg.Content[0].Type)
	assert.NotNil(t, msg.Content[0].Image)
	assert.Equal(t, SourceTypeURL, msg.Content[0].Image.Source.Type)
	assert.Equal(t, imageURL, msg.Content[0].Image.Source.URL)

	// Check text content
	assert.Equal(t, ContentTypeText, msg.Content[1].Type)
	assert.Equal(t, text, msg.Content[1].Text)
}
