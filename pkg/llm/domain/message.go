// ABOUTME: This file defines the core message domain models for LLM providers.
// ABOUTME: It supports both text and multimodal content in messages.

package domain

import (
	"encoding/base64"
)

// Role represents the role of a message sender in a conversation.
// It determines how the LLM interprets and responds to the message.
type Role string

const (
	// RoleSystem represents system messages that set context or behavior for the conversation.
	// System messages typically contain instructions that guide the model's responses.
	RoleSystem Role = "system"

	// RoleUser represents messages from the human user.
	// These are the primary inputs that the model responds to.
	RoleUser Role = "user"

	// RoleAssistant represents messages from the AI assistant.
	// These are the model's responses to user inputs.
	RoleAssistant Role = "assistant"

	// RoleTool represents results from function or tool calls.
	// Used when the model invokes external tools and receives their output.
	RoleTool Role = "tool"
)

// ContentType represents the type of content in a message part.
// It indicates the format and handling required for multimodal content.
type ContentType string

const (
	// ContentTypeText represents plain text content.
	ContentTypeText ContentType = "text"

	// ContentTypeImage represents image content (JPEG, PNG, etc.).
	ContentTypeImage ContentType = "image"

	// ContentTypeFile represents general file attachments.
	ContentTypeFile ContentType = "file"

	// ContentTypeVideo represents video content.
	ContentTypeVideo ContentType = "video"

	// ContentTypeAudio represents audio content.
	ContentTypeAudio ContentType = "audio"
)

// SourceType indicates how media content is provided to the model.
// Content can be embedded as base64 data or referenced by URL.
type SourceType string

const (
	// SourceTypeBase64 indicates content is embedded as base64-encoded data.
	SourceTypeBase64 SourceType = "base64"

	// SourceTypeURL indicates content is referenced by a URL.
	SourceTypeURL SourceType = "url"
)

// SourceInfo describes how to access media content in a message.
// It supports both embedded data and external URLs.
type SourceInfo struct {
	Type      SourceType `json:"type"`
	MediaType string     `json:"media_type,omitempty"` // MIME type
	Data      string     `json:"data,omitempty"`       // Base64 encoded
	URL       string     `json:"url,omitempty"`
}

// ImageContent represents image data within a multimodal message.
// Images can be provided as base64-encoded data or URLs.
type ImageContent struct {
	Source SourceInfo `json:"source"`
}

// FileContent represents a file attachment in a message.
// Files are embedded as base64-encoded data with metadata.
type FileContent struct {
	FileName string `json:"file_name"`
	FileData string `json:"file_data"` // Base64 encoded
	MimeType string `json:"mime_type"` // MIME type
}

// VideoContent represents video data within a multimodal message.
// Videos can be provided as base64-encoded data or URLs.
type VideoContent struct {
	Source SourceInfo `json:"source"`
}

// AudioContent represents audio data within a multimodal message.
// Audio can be provided as base64-encoded data or URLs.
type AudioContent struct {
	Source SourceInfo `json:"source"`
}

// ContentPart represents a single piece of content within a multimodal message.
// A message can contain multiple content parts of different types (text, images, etc.).
type ContentPart struct {
	Type  ContentType   `json:"type"`
	Text  string        `json:"text,omitempty"`
	Image *ImageContent `json:"image,omitempty"`
	File  *FileContent  `json:"file,omitempty"`
	Video *VideoContent `json:"video,omitempty"`
	Audio *AudioContent `json:"audio,omitempty"`
}

// Message represents a single message in an LLM conversation.
// It supports multimodal content through an array of content parts,
// allowing text, images, and other media to be combined in a single message.
type Message struct {
	Role    Role          `json:"role"`
	Content []ContentPart `json:"content"`
}

// NewTextMessage creates a message with only text content
func NewTextMessage(role Role, text string) Message {
	return Message{
		Role: role,
		Content: []ContentPart{
			{
				Type: ContentTypeText,
				Text: text,
			},
		},
	}
}

// NewImageMessage creates a message with a base64-encoded image and optional text
func NewImageMessage(role Role, imageData []byte, mimeType string, text string) Message {
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	parts := []ContentPart{
		{
			Type: ContentTypeImage,
			Image: &ImageContent{
				Source: SourceInfo{
					Type:      SourceTypeBase64,
					MediaType: mimeType,
					Data:      base64Data,
				},
			},
		},
	}

	if text != "" {
		parts = append(parts, ContentPart{
			Type: ContentTypeText,
			Text: text,
		})
	}

	return Message{
		Role:    role,
		Content: parts,
	}
}

// NewImageURLMessage creates a message with an image URL and optional text
func NewImageURLMessage(role Role, imageURL string, text string) Message {
	parts := []ContentPart{
		{
			Type: ContentTypeImage,
			Image: &ImageContent{
				Source: SourceInfo{
					Type: SourceTypeURL,
					URL:  imageURL,
				},
			},
		},
	}

	if text != "" {
		parts = append(parts, ContentPart{
			Type: ContentTypeText,
			Text: text,
		})
	}

	return Message{
		Role:    role,
		Content: parts,
	}
}

// NewFileMessage creates a message with a file attachment and optional text
func NewFileMessage(role Role, fileName string, fileData []byte, mimeType string, text string) Message {
	base64Data := base64.StdEncoding.EncodeToString(fileData)

	parts := []ContentPart{
		{
			Type: ContentTypeFile,
			File: &FileContent{
				FileName: fileName,
				FileData: base64Data,
				MimeType: mimeType,
			},
		},
	}

	if text != "" {
		parts = append(parts, ContentPart{
			Type: ContentTypeText,
			Text: text,
		})
	}

	return Message{
		Role:    role,
		Content: parts,
	}
}

// NewVideoMessage creates a message with a video attachment and optional text
func NewVideoMessage(role Role, videoData []byte, mimeType string, text string) Message {
	base64Data := base64.StdEncoding.EncodeToString(videoData)

	parts := []ContentPart{
		{
			Type: ContentTypeVideo,
			Video: &VideoContent{
				Source: SourceInfo{
					Type:      SourceTypeBase64,
					MediaType: mimeType,
					Data:      base64Data,
				},
			},
		},
	}

	if text != "" {
		parts = append(parts, ContentPart{
			Type: ContentTypeText,
			Text: text,
		})
	}

	return Message{
		Role:    role,
		Content: parts,
	}
}

// NewAudioMessage creates a message with an audio attachment and optional text
func NewAudioMessage(role Role, audioData []byte, mimeType string, text string) Message {
	base64Data := base64.StdEncoding.EncodeToString(audioData)

	parts := []ContentPart{
		{
			Type: ContentTypeAudio,
			Audio: &AudioContent{
				Source: SourceInfo{
					Type:      SourceTypeBase64,
					MediaType: mimeType,
					Data:      base64Data,
				},
			},
		},
	}

	if text != "" {
		parts = append(parts, ContentPart{
			Type: ContentTypeText,
			Text: text,
		})
	}

	return Message{
		Role:    role,
		Content: parts,
	}
}

// Token represents a single token in a streaming response from an LLM.
// Tokens are sent incrementally as the model generates output, with Finished
// indicating the end of the stream.
type Token struct {
	Text     string `json:"text"`
	Finished bool   `json:"finished"`
}

// Response represents a complete response from an LLM provider.
// It contains the generated content as a single string after generation completes.
type Response struct {
	Content string `json:"content"`
}

// ResponseStream represents a channel that streams tokens from an LLM as they are generated.
// The stream is closed when generation completes or an error occurs.
type ResponseStream <-chan Token
