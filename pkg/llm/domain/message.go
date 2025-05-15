// ABOUTME: This file defines the core message domain models for LLM providers.
// ABOUTME: It supports both text and multimodal content in messages.

package domain

import (
	"encoding/base64"
)

// Role represents the role of a message sender
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ContentType represents the type of content in a message part
type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
	ContentTypeFile  ContentType = "file"
	ContentTypeVideo ContentType = "video"
	ContentTypeAudio ContentType = "audio"
)

// SourceType represents how the content is sourced
type SourceType string

const (
	SourceTypeBase64 SourceType = "base64"
	SourceTypeURL    SourceType = "url"
)

// SourceInfo represents the source of media content
type SourceInfo struct {
	Type      SourceType `json:"type"`
	MediaType string     `json:"media_type,omitempty"` // MIME type
	Data      string     `json:"data,omitempty"`       // Base64 encoded
	URL       string     `json:"url,omitempty"`
}

// ImageContent represents an image in a message
type ImageContent struct {
	Source SourceInfo `json:"source"`
}

// FileContent represents a file in a message
type FileContent struct {
	FileName string `json:"file_name"`
	FileData string `json:"file_data"` // Base64 encoded
	MimeType string `json:"mime_type"` // MIME type
}

// VideoContent represents a video in a message
type VideoContent struct {
	Source SourceInfo `json:"source"`
}

// AudioContent represents audio in a message
type AudioContent struct {
	Source SourceInfo `json:"source"`
}

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
		Role: role,
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
		Role: role,
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
		Role: role,
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
		Role: role,
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
		Role: role,
		Content: parts,
	}
}

// Token represents a token in a streamed response
type Token struct {
	Text     string `json:"text"`
	Finished bool   `json:"finished"`
}

// Response represents a complete response from an LLM
type Response struct {
	Content string `json:"content"`
}

// ResponseStream represents a stream of tokens from an LLM
type ResponseStream <-chan Token