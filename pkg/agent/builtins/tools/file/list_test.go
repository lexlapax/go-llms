// ABOUTME: Tests for the FileList built-in tool
// ABOUTME: Validates directory listing, filtering, sorting, and edge cases

package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	"github.com/lexlapax/go-llms/pkg/testutils/mocks"
)

// Helper to create test ToolContext
func createTestToolContext() *domain.ToolContext {
	return domain.NewToolContext(
		context.Background(),
		domain.NewStateReader(domain.NewState()),
		mocks.NewMockAgent("Test Agent"),
		"test-run",
	)
}

func TestFileListRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("file_list")
	if !ok {
		t.Fatal("FileList tool not registered")
	}
	if tool == nil {
		t.Fatal("FileList tool is nil")
	}

	// Test tool name
	if tool.Name() != "file_list" {
		t.Errorf("Expected tool name 'file_list', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("file_list")
	if len(entries) == 0 {
		t.Fatal("FileList tool not found in registry")
	}

	meta := entries[0].Metadata
	if meta.Category != "file" {
		t.Errorf("Expected category 'file', got '%s'", meta.Category)
	}
}

func TestFileListBasic(t *testing.T) {
	// Create test directory structure
	tempDir := t.TempDir()

	// Create some test files
	testFiles := []struct {
		path    string
		content string
		isDir   bool
	}{
		{"file1.txt", "content1", false},
		{"file2.txt", "content2", false},
		{"script.sh", "#!/bin/bash", false},
		{"data.json", `{"key": "value"}`, false},
		{"subdir", "", true},
		{"subdir/nested.txt", "nested content", false},
		{"subdir/deep", "", true},
		{"subdir/deep/hidden.txt", "deep content", false},
	}

	for _, tf := range testFiles {
		fullPath := filepath.Join(tempDir, tf.path)
		if tf.isDir {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				t.Fatal(err)
			}
		} else {
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(fullPath, []byte(tf.content), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}

	tool := FileList()
	ctx := createTestToolContext()

	// Test 1: List all files (non-recursive)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path": tempDir,
	})
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	listResult := result.(*FileListResult)

	// Should have 4 files in root directory (not including subdir by default)
	if len(listResult.Files) != 4 {
		t.Errorf("Expected 4 files, got %d", len(listResult.Files))
		for _, f := range listResult.Files {
			t.Logf("  - %s (dir: %v)", f.Name, f.IsDir)
		}
	}

	// Test 2: List with pattern
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":    tempDir,
		"pattern": "*.txt",
	})
	if err != nil {
		t.Fatalf("Failed to list with pattern: %v", err)
	}

	listResult = result.(*FileListResult)
	if len(listResult.Files) != 2 {
		t.Errorf("Expected 2 .txt files, got %d", len(listResult.Files))
	}

	// Test 3: Recursive listing
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":      tempDir,
		"recursive": true,
		"pattern":   "*.txt",
	})
	if err != nil {
		t.Fatalf("Failed to list recursively: %v", err)
	}

	listResult = result.(*FileListResult)
	if len(listResult.Files) != 4 { // 2 in root + 2 in subdirs
		t.Errorf("Expected 4 .txt files recursively, got %d", len(listResult.Files))
	}

	// Test 4: Include directories
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":          tempDir,
		"include_dirs":  true,
		"include_files": false,
	})
	if err != nil {
		t.Fatalf("Failed to list directories: %v", err)
	}

	listResult = result.(*FileListResult)
	dirCount := 0
	for _, f := range listResult.Files {
		if f.IsDir {
			dirCount++
		}
	}
	if dirCount != 1 { // Only subdir in root
		t.Errorf("Expected 1 directory, got %d", dirCount)
	}
}

func TestFileListFiltering(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with different sizes
	smallFile := filepath.Join(tempDir, "small.txt")
	mediumFile := filepath.Join(tempDir, "medium.txt")
	largeFile := filepath.Join(tempDir, "large.txt")

	if err := os.WriteFile(smallFile, []byte("small"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mediumFile, make([]byte, 1024), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(largeFile, make([]byte, 10240), 0644); err != nil {
		t.Fatal(err)
	}

	// Modify times for time-based filtering
	oldTime := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(smallFile, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	tool := FileList()
	ctx := createTestToolContext()

	// Test size filtering
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":     tempDir,
		"min_size": 1000,
	})
	if err != nil {
		t.Fatalf("Failed to filter by size: %v", err)
	}

	listResult := result.(*FileListResult)
	if len(listResult.Files) != 2 {
		t.Errorf("Expected 2 files >= 1KB, got %d", len(listResult.Files))
	}

	// Test time filtering
	yesterday := time.Now().Add(-24 * time.Hour)
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":           tempDir,
		"modified_after": yesterday.Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("Failed to filter by time: %v", err)
	}

	listResult = result.(*FileListResult)
	if len(listResult.Files) != 2 { // medium and large files are recent
		t.Errorf("Expected 2 recent files, got %d", len(listResult.Files))
	}
}

func TestFileListSorting(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with specific attributes
	files := []struct {
		name    string
		content string
		sleep   time.Duration
	}{
		{"zebra.txt", "z", 0},
		{"alpha.txt", "a", 100 * time.Millisecond},
		{"beta.txt", "bb", 100 * time.Millisecond},
	}

	for _, f := range files {
		time.Sleep(f.sleep) // Ensure different modification times
		if err := os.WriteFile(filepath.Join(tempDir, f.name), []byte(f.content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tool := FileList()
	ctx := createTestToolContext()

	// Test sort by name
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":    tempDir,
		"sort_by": "name",
	})
	if err != nil {
		t.Fatalf("Failed to sort by name: %v", err)
	}

	listResult := result.(*FileListResult)
	if len(listResult.Files) < 3 {
		t.Fatal("Not enough files returned")
	}
	if listResult.Files[0].Name != "alpha.txt" {
		t.Errorf("Expected first file to be alpha.txt, got %s", listResult.Files[0].Name)
	}
	if listResult.Files[2].Name != "zebra.txt" {
		t.Errorf("Expected last file to be zebra.txt, got %s", listResult.Files[2].Name)
	}

	// Test sort by size
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":    tempDir,
		"sort_by": "size",
	})
	if err != nil {
		t.Fatalf("Failed to sort by size: %v", err)
	}

	listResult = result.(*FileListResult)
	// Files should be sorted by size: a(1) < z(1) < bb(2)
	// When sizes are equal, falls back to name sorting
	if listResult.Files[0].Name != "alpha.txt" {
		t.Errorf("Expected first file to be alpha.txt (size 1), got %s", listResult.Files[0].Name)
	}
	if listResult.Files[2].Name != "beta.txt" {
		t.Errorf("Expected last file to be beta.txt (size 2), got %s", listResult.Files[2].Name)
	}

	// Test reverse sorting
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":         tempDir,
		"sort_by":      "name",
		"sort_reverse": true,
	})
	if err != nil {
		t.Fatalf("Failed to reverse sort: %v", err)
	}

	listResult = result.(*FileListResult)
	if listResult.Files[0].Name != "zebra.txt" {
		t.Errorf("Expected first file to be zebra.txt in reverse, got %s", listResult.Files[0].Name)
	}
}

func TestFileListErrors(t *testing.T) {
	tool := FileList()
	ctx := createTestToolContext()

	// Test non-existent directory
	_, err := tool.Execute(ctx, map[string]interface{}{
		"path": "/non/existent/directory",
	})
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}

	// Test file instead of directory
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()
	_ = tempFile.Close()

	_, err = tool.Execute(ctx, map[string]interface{}{
		"path": tempFile.Name(),
	})
	if err == nil {
		t.Error("Expected error when path is a file")
	}

	// Test invalid time format
	tempDir := t.TempDir()
	_, err = tool.Execute(ctx, map[string]interface{}{
		"path":           tempDir,
		"modified_after": "invalid-time",
	})
	if err == nil {
		t.Error("Expected error for invalid time format")
	}
}

func TestFileListMaxResults(t *testing.T) {
	tempDir := t.TempDir()

	// Create many files
	for i := 0; i < 20; i++ {
		filename := fmt.Sprintf("file%02d.txt", i)
		if err := os.WriteFile(filepath.Join(tempDir, filename), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tool := FileList()
	ctx := createTestToolContext()

	// Test max results limit
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":        tempDir,
		"max_results": 5,
	})
	if err != nil {
		t.Fatalf("Failed to limit results: %v", err)
	}

	listResult := result.(*FileListResult)
	if len(listResult.Files) != 5 {
		t.Errorf("Expected 5 files with max_results=5, got %d", len(listResult.Files))
	}
	if listResult.TotalCount != 20 {
		t.Errorf("Expected total count of 20, got %d", listResult.TotalCount)
	}
}

func TestFileListExtensions(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with different extensions
	files := []string{
		"document.txt",
		"image.png",
		"script.sh",
		"data.json",
		"archive.tar.gz",
		"noext",
	}

	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tempDir, f), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tool := FileList()
	ctx := createTestToolContext()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"path": tempDir,
	})
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	listResult := result.(*FileListResult)

	// Check extensions are properly extracted
	extMap := make(map[string]string)
	for _, f := range listResult.Files {
		extMap[f.Name] = f.Extension
	}

	expectedExts := map[string]string{
		"document.txt":   "txt",
		"image.png":      "png",
		"script.sh":      "sh",
		"data.json":      "json",
		"archive.tar.gz": "gz",
		"noext":          "",
	}

	for name, expectedExt := range expectedExts {
		if extMap[name] != expectedExt {
			t.Errorf("Expected extension '%s' for %s, got '%s'", expectedExt, name, extMap[name])
		}
	}
}
