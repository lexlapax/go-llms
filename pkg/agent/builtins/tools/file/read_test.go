// ABOUTME: Tests for the file reading tool with all enhanced features
// ABOUTME: Verifies metadata extraction, line range reading, and binary detection

package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// createTestToolContextForRead creates a ToolContext for testing
func createTestToolContextForRead() *domain.ToolContext {
	ctx := context.Background()
	state := domain.NewState()
	stateReader := domain.NewStateReader(state)
	agent := mocks.NewMockAgent("Test Agent")

	tc := domain.NewToolContext(ctx, stateReader, agent, "test-run-id")

	// Create a simple event emitter
	tc = tc.WithEventEmitter(&testEventEmitter{})

	return tc
}

// testEventEmitter implements EventEmitter for testing
type testEventEmitter struct{}

func (e *testEventEmitter) Emit(eventType domain.EventType, data interface{}) {}
func (e *testEventEmitter) EmitProgress(current, total int, message string)   {}
func (e *testEventEmitter) EmitMessage(message string)                        {}
func (e *testEventEmitter) EmitError(err error)                               {}
func (e *testEventEmitter) EmitCustom(eventName string, data interface{})     {}

func TestReadFileRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("file_read")
	if !ok {
		t.Fatal("ReadFile tool not registered")
	}
	if tool == nil {
		t.Fatal("ReadFile tool is nil")
	}

	// Test tool name
	if tool.Name() != "file_read" {
		t.Errorf("Expected tool name 'file_read', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("file_read")
	if len(entries) == 0 {
		t.Fatal("ReadFile tool not found in registry")
	}

	meta := entries[0].Metadata
	if meta.Category != "file" {
		t.Errorf("Expected category 'file', got '%s'", meta.Category)
	}
}

func TestReadFile_Basic(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!\nThis is a test file.\nWith multiple lines."

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test basic read
	tool := MustGetReadFile()
	ctx := createTestToolContextForRead()

	result, err := tool.Execute(ctx, ReadFileParams{
		Path: testFile,
	})
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	readResult := result.(*ReadFileResult)
	if readResult.Content != testContent {
		t.Errorf("Expected content %q, got %q", testContent, readResult.Content)
	}
	if readResult.IsBinary {
		t.Error("Expected text file, but detected as binary")
	}
	if readResult.Encoding != "utf-8" {
		t.Errorf("Expected UTF-8 encoding, got %s", readResult.Encoding)
	}
	if readResult.Lines != 3 {
		t.Errorf("Expected 3 lines, got %d", readResult.Lines)
	}
}

func TestReadFile_WithMetadata(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.json")
	testContent := `{"key": "value"}`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := MustGetReadFile()
	ctx := createTestToolContextForRead()

	result, err := tool.Execute(ctx, ReadFileParams{
		Path:        testFile,
		IncludeMeta: true,
	})
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	readResult := result.(*ReadFileResult)
	if readResult.Metadata == nil {
		t.Fatal("Expected metadata, got nil")
	}
	if readResult.Metadata.Extension != ".json" {
		t.Errorf("Expected .json extension, got %s", readResult.Metadata.Extension)
	}
	if readResult.Metadata.Size != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), readResult.Metadata.Size)
	}
}

func TestReadFile_LineRange(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "lines.txt")

	// Create file with numbered lines
	var lines []string
	for i := 1; i <= 10; i++ {
		lines = append(lines, fmt.Sprintf("Line %d", i))
	}
	testContent := strings.Join(lines, "\n")

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := MustGetReadFile()
	ctx := createTestToolContextForRead()

	// Test reading lines 3-5
	result, err := tool.Execute(ctx, ReadFileParams{
		Path:      testFile,
		LineStart: 3,
		LineEnd:   5,
	})
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	readResult := result.(*ReadFileResult)
	expected := "Line 3\nLine 4\nLine 5"
	if readResult.Content != expected {
		t.Errorf("Expected content %q, got %q", expected, readResult.Content)
	}
	if readResult.Lines != 3 {
		t.Errorf("Expected 3 lines read, got %d", readResult.Lines)
	}
}

func TestReadFile_BinaryDetection(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "binary.dat")

	// Create binary file with null bytes
	binaryContent := []byte{0x00, 0xFF, 0x42, 0x00, 0xAB}
	if err := os.WriteFile(testFile, binaryContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := MustGetReadFile()
	ctx := createTestToolContextForRead()

	result, err := tool.Execute(ctx, ReadFileParams{
		Path: testFile,
	})
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	readResult := result.(*ReadFileResult)
	if !readResult.IsBinary {
		t.Error("Expected binary file detection")
	}
	if readResult.Encoding != "binary" {
		t.Errorf("Expected binary encoding, got %s", readResult.Encoding)
	}
}

func TestReadFile_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "large.txt")

	// Create a large file (over 10MB)
	var content strings.Builder
	lineContent := strings.Repeat("x", 100) + "\n" // 101 bytes per line
	for i := 0; i < 120000; i++ {                  // ~12MB
		content.WriteString(lineContent)
	}

	if err := os.WriteFile(testFile, []byte(content.String()), 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	tool := MustGetReadFile()
	ctx := createTestToolContextForRead()

	result, err := tool.Execute(ctx, ReadFileParams{
		Path: testFile,
	})
	if err != nil {
		t.Fatalf("Failed to read large file: %v", err)
	}

	readResult := result.(*ReadFileResult)
	// Should be truncated to approximately 10MB (10 * 1024 * 1024 bytes)
	if len(readResult.Content) > 10*1024*1024 {
		t.Errorf("Content not properly truncated, got %d bytes", len(readResult.Content))
	}
	// Check if warnings contain truncation message
	foundTruncationWarning := false
	for _, warning := range readResult.Warnings {
		if strings.Contains(warning, "truncated") || strings.Contains(warning, "size limit") {
			foundTruncationWarning = true
			break
		}
	}
	if !foundTruncationWarning && len(readResult.Warnings) > 0 {
		t.Logf("Warnings: %v", readResult.Warnings)
	}
}

func TestReadFile_NonExistentFile(t *testing.T) {
	tool := MustGetReadFile()
	ctx := createTestToolContextForRead()

	_, err := tool.Execute(ctx, ReadFileParams{
		Path: "/non/existent/file.txt",
	})
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestReadFile_DirectoryInsteadOfFile(t *testing.T) {
	tempDir := t.TempDir()

	tool := MustGetReadFile()
	ctx := createTestToolContextForRead()

	_, err := tool.Execute(ctx, ReadFileParams{
		Path: tempDir,
	})
	if err == nil {
		t.Error("Expected error when trying to read directory")
	}
}

func TestReadFile_PermissionDenied(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "noperms.txt")

	// Create file with no read permissions
	if err := os.WriteFile(testFile, []byte("secret"), 0000); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := MustGetReadFile()
	ctx := createTestToolContextForRead()

	_, err := tool.Execute(ctx, ReadFileParams{
		Path: testFile,
	})
	if err == nil {
		t.Error("Expected permission denied error")
	}
}

func TestReadFile_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "empty.txt")

	// Create empty file
	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	tool := MustGetReadFile()
	ctx := createTestToolContextForRead()

	result, err := tool.Execute(ctx, ReadFileParams{
		Path: testFile,
	})
	if err != nil {
		t.Fatalf("Failed to read empty file: %v", err)
	}

	readResult := result.(*ReadFileResult)
	if readResult.Content != "" {
		t.Errorf("Expected empty content, got %q", readResult.Content)
	}
	// An empty file is still counted as having 1 line in the implementation
	if readResult.Lines != 1 {
		t.Errorf("Expected 1 line for empty file, got %d", readResult.Lines)
	}
}

func TestReadFile_LineRangeValidation(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "lines.txt")

	// Create file with 5 lines
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := MustGetReadFile()
	ctx := createTestToolContextForRead()

	testCases := []struct {
		name      string
		lineStart int
		lineEnd   int
		shouldErr bool
		expected  string
	}{
		{"Valid range", 2, 4, false, "Line 2\nLine 3\nLine 4"},
		{"Start > End", 4, 2, false, ""},                            // No validation, just returns empty
		{"Negative start", -1, 3, false, "Line 1\nLine 2\nLine 3"},  // Treated as 0
		{"Zero start", 0, 3, false, "Line 1\nLine 2\nLine 3"},       // 0 means start from beginning
		{"End beyond file", 3, 10, false, "Line 3\nLine 4\nLine 5"}, // Should read to end
		{"Start beyond file", 10, 12, false, ""},                    // Should return empty
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tool.Execute(ctx, ReadFileParams{
				Path:      testFile,
				LineStart: tc.lineStart,
				LineEnd:   tc.lineEnd,
			})
			if tc.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil && !tc.shouldErr {
				readResult := result.(*ReadFileResult)
				if readResult.Content != tc.expected {
					t.Errorf("Expected content %q, got %q", tc.expected, readResult.Content)
				}
			}
		})
	}
}
