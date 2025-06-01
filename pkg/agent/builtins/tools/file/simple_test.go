// ABOUTME: Simple tests to debug tool execution
// ABOUTME: Minimal tests to isolate the hanging issue

package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileToolDirect(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Get the tool directly
	tool := ReadFile()
	if tool == nil {
		t.Fatal("ReadFile() returned nil")
	}

	// Execute with minimal parameters
	ctx := context.Background()
	params := map[string]interface{}{
		"path": testFile,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Failed to execute tool: %v", err)
	}

	// Check result type
	readResult, ok := result.(*ReadFileResult)
	if !ok {
		t.Fatalf("Result is not *ReadFileResult, got %T", result)
	}

	if readResult.Content != testContent {
		t.Errorf("Expected content %q, got %q", testContent, readResult.Content)
	}
}

func TestWriteFileToolDirect(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "output.txt")
	testContent := "Test content"

	// Get the tool directly
	tool := WriteFile()
	if tool == nil {
		t.Fatal("WriteFile() returned nil")
	}

	// Execute with minimal parameters
	ctx := context.Background()
	params := map[string]interface{}{
		"path":    testFile,
		"content": testContent,
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("Failed to execute tool: %v", err)
	}

	// Check result type
	writeResult, ok := result.(*WriteFileResult)
	if !ok {
		t.Fatalf("Result is not *WriteFileResult, got %T", result)
	}

	if !writeResult.Success {
		t.Error("Write operation failed")
	}

	// Verify file was created
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("Expected content %q, got %q", testContent, string(content))
	}
}
