// ABOUTME: Tests for file reading and writing tools with all enhanced features
// ABOUTME: Verifies atomic operations, append mode, line reading, and metadata functionality

package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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
	ctx := context.Background()

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
	ctx := context.Background()

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
	ctx := context.Background()

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
	ctx := context.Background()

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

func TestWriteFile_Basic(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "output.txt")
	testContent := "Hello from WriteFile!"

	tool := MustGetWriteFile()
	ctx := context.Background()

	result, err := tool.Execute(ctx, WriteFileParams{
		Path:    testFile,
		Content: testContent,
	})
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	writeResult := result.(*WriteFileResult)
	if !writeResult.Success {
		t.Error("Write operation failed")
	}
	if writeResult.BytesWritten != len(testContent) {
		t.Errorf("Expected %d bytes written, got %d", len(testContent), writeResult.BytesWritten)
	}
	if writeResult.FileExisted {
		t.Error("File should not have existed before write")
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("Expected content %q, got %q", testContent, string(content))
	}
}

func TestWriteFile_Append(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "append.txt")

	// Create initial file
	initialContent := "Initial content\n"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := MustGetWriteFile()
	ctx := context.Background()

	appendContent := "Appended content\n"
	result, err := tool.Execute(ctx, WriteFileParams{
		Path:    testFile,
		Content: appendContent,
		Append:  true,
	})
	if err != nil {
		t.Fatalf("Failed to append to file: %v", err)
	}

	writeResult := result.(*WriteFileResult)
	if !writeResult.FileExisted {
		t.Error("File should have existed before append")
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	expected := initialContent + appendContent
	if string(content) != expected {
		t.Errorf("Expected content %q, got %q", expected, string(content))
	}
}

func TestWriteFile_CreateDirs(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "nested", "dirs", "file.txt")
	testContent := "Content in nested directory"

	tool := MustGetWriteFile()
	ctx := context.Background()

	result, err := tool.Execute(ctx, WriteFileParams{
		Path:       testFile,
		Content:    testContent,
		CreateDirs: true,
	})
	if err != nil {
		t.Fatalf("Failed to write file with directory creation: %v", err)
	}

	writeResult := result.(*WriteFileResult)
	if !writeResult.Success {
		t.Error("Write operation failed")
	}

	// Verify file exists
	if _, err := os.Stat(testFile); err != nil {
		t.Errorf("File was not created: %v", err)
	}
}

func TestWriteFile_Atomic(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "atomic.txt")

	// Create initial file
	initialContent := "Initial atomic content"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := MustGetWriteFile()
	ctx := context.Background()

	newContent := "New atomic content"
	result, err := tool.Execute(ctx, WriteFileParams{
		Path:    testFile,
		Content: newContent,
		Atomic:  true,
	})
	if err != nil {
		t.Fatalf("Failed to write atomically: %v", err)
	}

	writeResult := result.(*WriteFileResult)
	if !writeResult.Success {
		t.Error("Atomic write operation failed")
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != newContent {
		t.Errorf("Expected content %q, got %q", newContent, string(content))
	}
}

func TestWriteFile_Backup(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "backup.txt")

	// Create initial file
	initialContent := "Original content to backup"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Wait a moment to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	tool := MustGetWriteFile()
	ctx := context.Background()

	newContent := "New content after backup"
	result, err := tool.Execute(ctx, WriteFileParams{
		Path:    testFile,
		Content: newContent,
		Backup:  true,
	})
	if err != nil {
		t.Fatalf("Failed to write with backup: %v", err)
	}

	writeResult := result.(*WriteFileResult)
	if writeResult.BackupPath == "" {
		t.Error("Expected backup path, got empty string")
	}

	// Verify backup exists and contains original content
	backupContent, err := os.ReadFile(writeResult.BackupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}
	if string(backupContent) != initialContent {
		t.Errorf("Backup content mismatch: expected %q, got %q", initialContent, string(backupContent))
	}

	// Verify main file has new content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != newContent {
		t.Errorf("Expected content %q, got %q", newContent, string(content))
	}

	// Clean up backup
	os.Remove(writeResult.BackupPath)
}

func TestWriteFile_CustomPermissions(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "perms.txt")

	tool := MustGetWriteFile()
	ctx := context.Background()

	result, err := tool.Execute(ctx, WriteFileParams{
		Path:    testFile,
		Content: "Test permissions",
		Mode:    0600, // Read/write for owner only
	})
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	writeResult := result.(*WriteFileResult)
	if !writeResult.Success {
		t.Error("Write operation failed")
	}

	// Verify permissions
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Expected permissions 0600, got %o", mode)
	}
}
