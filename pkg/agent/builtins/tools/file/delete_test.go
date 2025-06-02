// ABOUTME: Tests for the FileDelete built-in tool
// ABOUTME: Validates safe deletion, confirmation, and edge cases

package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
)

func TestFileDeleteRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("file_delete")
	if !ok {
		t.Fatal("FileDelete tool not registered")
	}
	if tool == nil {
		t.Fatal("FileDelete tool is nil")
	}

	// Test tool name
	if tool.Name() != "file_delete" {
		t.Errorf("Expected tool name 'file_delete', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("file_delete")
	if len(entries) == 0 {
		t.Fatal("FileDelete tool not found in registry")
	}

	meta := entries[0].Metadata
	if meta.Category != "file" {
		t.Errorf("Expected category 'file', got '%s'", meta.Category)
	}
}

func TestFileDeleteBasic(t *testing.T) {
	tool := FileDelete()
	ctx := context.Background()

	// Test 1: Delete a regular file
	tempFile, err := os.CreateTemp("", "test-delete-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tempPath := tempFile.Name()
	_, err = tempFile.Write([]byte("test content"))
	if err != nil {
		tempFile.Close()
		t.Fatal(err)
	}
	tempFile.Close()

	// Verify file exists
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		t.Fatal("Test file was not created")
	}

	result, err := tool.Execute(ctx, map[string]interface{}{
		"path": tempPath,
	})
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	deleteResult := result.(*FileDeleteResult)
	if !deleteResult.Deleted {
		t.Error("File was not deleted")
	}
	if deleteResult.WasDirectory {
		t.Error("File was incorrectly marked as directory")
	}

	// Verify file is gone
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("File still exists after deletion")
	}

	// Test 2: Delete non-existent file
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path": "/non/existent/file.txt",
	})
	if err != nil {
		t.Fatalf("Unexpected error for non-existent file: %v", err)
	}

	deleteResult = result.(*FileDeleteResult)
	if deleteResult.Deleted {
		t.Error("Non-existent file reported as deleted")
	}
	if deleteResult.Message != "Path does not exist" {
		t.Errorf("Expected 'Path does not exist' message, got: %s", deleteResult.Message)
	}
}

func TestFileDeleteDirectory(t *testing.T) {
	tool := FileDelete()
	ctx := context.Background()

	// Test 1: Delete empty directory
	tempDir, err := os.MkdirTemp("", "test-delete-dir-*")
	if err != nil {
		t.Fatal(err)
	}

	result, err := tool.Execute(ctx, map[string]interface{}{
		"path": tempDir,
	})
	if err != nil {
		t.Fatalf("Failed to delete empty directory: %v", err)
	}

	deleteResult := result.(*FileDeleteResult)
	if !deleteResult.Deleted {
		t.Error("Empty directory was not deleted")
	}
	if !deleteResult.WasDirectory {
		t.Error("Directory was not marked as directory")
	}

	// Verify directory is gone
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Directory still exists after deletion")
	}

	// Test 2: Try to delete non-empty directory without recursive
	tempDir2, err := os.MkdirTemp("", "test-delete-dir2-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir2) // Clean up

	// Create a file in the directory
	testFile := filepath.Join(tempDir2, "test.txt")
	err = os.WriteFile(testFile, []byte("content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	result, err = tool.Execute(ctx, map[string]interface{}{
		"path": tempDir2,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	deleteResult = result.(*FileDeleteResult)
	if deleteResult.Deleted {
		t.Error("Non-empty directory should not be deleted without recursive flag")
	}
	if !strings.Contains(deleteResult.Message, "not empty") {
		t.Errorf("Expected 'not empty' in message, got: %s", deleteResult.Message)
	}

	// Test 3: Delete non-empty directory with recursive and confirmation
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":            tempDir2,
		"recursive":       true,
		"require_confirm": tempDir2,
	})
	if err != nil {
		t.Fatalf("Failed to delete directory recursively: %v", err)
	}

	deleteResult = result.(*FileDeleteResult)
	if !deleteResult.Deleted {
		t.Error("Directory was not deleted with recursive flag")
	}

	// Verify directory is gone
	if _, err := os.Stat(tempDir2); !os.IsNotExist(err) {
		t.Error("Directory still exists after recursive deletion")
	}
}

func TestFileDeleteConfirmation(t *testing.T) {
	tool := FileDelete()
	ctx := context.Background()

	// Create a test file
	tempFile, err := os.CreateTemp("", "test-confirm-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath) // Clean up if test fails

	// Test 1: Wrong confirmation
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":            tempPath,
		"require_confirm": "wrong-path",
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	deleteResult := result.(*FileDeleteResult)
	if deleteResult.Deleted {
		t.Error("File should not be deleted with wrong confirmation")
	}
	if !strings.Contains(deleteResult.Message, "Confirmation mismatch") {
		t.Errorf("Expected confirmation mismatch message, got: %s", deleteResult.Message)
	}

	// Test 2: Correct confirmation with full path
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":            tempPath,
		"require_confirm": tempPath,
	})
	if err != nil {
		t.Fatalf("Failed to delete with correct confirmation: %v", err)
	}

	deleteResult = result.(*FileDeleteResult)
	if !deleteResult.Deleted {
		t.Error("File was not deleted with correct confirmation")
	}

	// Test 3: Confirmation with base name
	tempFile2, err := os.CreateTemp("", "test-confirm2-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tempPath2 := tempFile2.Name()
	tempFile2.Close()

	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":            tempPath2,
		"require_confirm": filepath.Base(tempPath2),
	})
	if err != nil {
		t.Fatalf("Failed to delete with base name confirmation: %v", err)
	}

	deleteResult = result.(*FileDeleteResult)
	if !deleteResult.Deleted {
		t.Error("File was not deleted with base name confirmation")
	}
}

func TestFileDeleteSafety(t *testing.T) {
	tool := FileDelete()
	ctx := context.Background()

	// Test critical path protection
	criticalPaths := []string{"/", "/etc", "/usr", "/bin"}
	if runtime.GOOS == "windows" {
		criticalPaths = []string{"C:\\", "C:\\Windows", "C:\\Program Files"}
	}

	for _, critical := range criticalPaths {
		result, err := tool.Execute(ctx, map[string]interface{}{
			"path": critical,
		})
		if err != nil {
			// Error is ok - means it was rejected
			continue
		}

		deleteResult := result.(*FileDeleteResult)
		if deleteResult.Deleted {
			t.Errorf("Critical path %s should not be deletable!", critical)
		}
		if !strings.Contains(deleteResult.Message, "critical") {
			t.Errorf("Expected critical path protection message for %s, got: %s", critical, deleteResult.Message)
		}
	}

	// Test that force can override (but don't actually delete!)
	// Just verify the message changes
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":  "/tmp", // Use /tmp as it's less critical but still protected
		"force": false,
	})
	// We're just testing the safety mechanism, not actually deleting
	if err == nil && result != nil {
		deleteResult := result.(*FileDeleteResult)
		// Should either error or indicate it's protected
		if deleteResult.Deleted {
			t.Error("Should not actually delete even /tmp without proper confirmation")
		}
	}
}

func TestFileDeleteForce(t *testing.T) {
	tool := FileDelete()
	ctx := context.Background()

	// Create directory with files
	tempDir, err := os.MkdirTemp("", "test-force-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Create some files
	for i := 0; i < 3; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("file%d.txt", i))
		if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test 1: Try without force or confirmation (should fail)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":      tempDir,
		"recursive": true,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	deleteResult := result.(*FileDeleteResult)
	if deleteResult.Deleted {
		t.Error("Non-empty directory should not be deleted without force or confirmation")
	}

	// Test 2: Delete with force
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":      tempDir,
		"recursive": true,
		"force":     true,
	})
	if err != nil {
		t.Fatalf("Failed to delete with force: %v", err)
	}

	deleteResult = result.(*FileDeleteResult)
	if !deleteResult.Deleted {
		t.Error("Directory was not deleted with force flag")
	}

	// Verify directory is gone
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Directory still exists after force deletion")
	}
}
