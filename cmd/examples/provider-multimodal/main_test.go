// ABOUTME: Unit tests for the multimodal example
// ABOUTME: Tests command-line argument validation and content type support checking

package main

import (
	"testing"
)

func TestProviderSupportsContent(t *testing.T) {
	tests := []struct {
		name         string
		provider     string
		mode         string
		mimeTypes    []string
		expectError  bool
		errorMessage string
	}{
		// OpenAI tests
		{
			name:        "OpenAI text mode",
			provider:    "openai",
			mode:        "text",
			mimeTypes:   []string{},
			expectError: false,
		},
		{
			name:        "OpenAI image mode",
			provider:    "openai",
			mode:        "image",
			mimeTypes:   []string{"image/jpeg", "image/png"},
			expectError: false,
		},
		{
			name:         "OpenAI audio mode not supported",
			provider:     "openai",
			mode:         "audio",
			mimeTypes:    []string{"audio/mp3"},
			expectError:  true,
			errorMessage: "OpenAI chat models don't support standalone audio inputs",
		},
		{
			name:         "OpenAI video mode not supported",
			provider:     "openai",
			mode:         "video",
			mimeTypes:    []string{"video/mp4"},
			expectError:  true,
			errorMessage: "OpenAI doesn't support video inputs",
		},
		{
			name:        "OpenAI mixed mode with text and images",
			provider:    "openai",
			mode:        "mixed",
			mimeTypes:   []string{"image/jpeg"},
			expectError: false,
		},
		{
			name:        "OpenAI mixed mode with audio allowed",
			provider:    "openai",
			mode:        "mixed",
			mimeTypes:   []string{"audio/mp3"},
			expectError: false, // Audio is allowed in mixed mode
		},

		// Anthropic tests
		{
			name:        "Anthropic text mode",
			provider:    "anthropic",
			mode:        "text",
			mimeTypes:   []string{},
			expectError: false,
		},
		{
			name:        "Anthropic image mode",
			provider:    "anthropic",
			mode:        "image",
			mimeTypes:   []string{"image/jpeg"},
			expectError: false,
		},
		{
			name:         "Anthropic audio mode not supported",
			provider:     "anthropic",
			mode:         "audio",
			mimeTypes:    []string{"audio/mp3"},
			expectError:  true,
			errorMessage: "Anthropic doesn't support audio inputs",
		},
		{
			name:         "Anthropic video mode not supported",
			provider:     "anthropic",
			mode:         "video",
			mimeTypes:    []string{"video/mp4"},
			expectError:  true,
			errorMessage: "Anthropic doesn't support video inputs",
		},

		// Gemini tests
		{
			name:        "Gemini supports all content types",
			provider:    "gemini",
			mode:        "mixed",
			mimeTypes:   []string{"image/jpeg", "audio/mp3", "video/mp4"},
			expectError: false,
		},
		{
			name:        "Gemini audio mode",
			provider:    "gemini",
			mode:        "audio",
			mimeTypes:   []string{"audio/mp3"},
			expectError: false,
		},
		{
			name:        "Gemini video mode",
			provider:    "gemini",
			mode:        "video",
			mimeTypes:   []string{"video/mp4"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := providerSupportsContent(tt.provider, tt.mode, tt.mimeTypes)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetMimeType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"image.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"picture.png", "image/png"},
		{"animation.gif", "image/gif"},
		{"audio.mp3", "audio/mpeg"},
		{"sound.wav", "audio/x-wav"},
		{"video.mp4", "video/mp4"},
		{"movie.avi", "video/x-msvideo"},
		{"clip.mov", "video/quicktime"},
		{"document.pdf", "application/pdf"},
		{"unknown.xyz", "chemical/x-xyz"},
		{"file", "application/octet-stream"}, // No extension
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getMimeType(tt.path)
			if result != tt.expected {
				t.Errorf("Expected MIME type '%s' for path '%s', got '%s'", tt.expected, tt.path, result)
			}
		})
	}
}

func TestArrayFlags(t *testing.T) {
	var flags arrayFlags

	// Test adding values
	err := flags.Set("file1.txt")
	if err != nil {
		t.Errorf("Failed to set first value: %v", err)
	}

	err = flags.Set("file2.txt")
	if err != nil {
		t.Errorf("Failed to set second value: %v", err)
	}

	// Test string representation
	expected := "file1.txt, file2.txt"
	if flags.String() != expected {
		t.Errorf("Expected string '%s', got '%s'", expected, flags.String())
	}

	// Test length
	if len(flags) != 2 {
		t.Errorf("Expected 2 values, got %d", len(flags))
	}

	// Test values
	if flags[0] != "file1.txt" || flags[1] != "file2.txt" {
		t.Errorf("Unexpected values: %v", flags)
	}
}
