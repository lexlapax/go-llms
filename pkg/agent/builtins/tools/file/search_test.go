// ABOUTME: Tests for the FileSearch built-in tool
// ABOUTME: Validates pattern searching, regex support, context lines, and file filtering

package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
)

func TestFileSearchRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("file_search")
	if !ok {
		t.Fatal("FileSearch tool not registered")
	}
	if tool == nil {
		t.Fatal("FileSearch tool is nil")
	}

	// Test tool name
	if tool.Name() != "file_search" {
		t.Errorf("Expected tool name 'file_search', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("file_search")
	if len(entries) == 0 {
		t.Fatal("FileSearch tool not found in registry")
	}
	
	meta := entries[0].Metadata
	if meta.Category != "file" {
		t.Errorf("Expected category 'file', got '%s'", meta.Category)
	}
}

func TestFileSearchBasic(t *testing.T) {
	tool := FileSearch()
	ctx := context.Background()

	// Create test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := `Line 1: Hello World
Line 2: This is a test
Line 3: TODO: Fix this bug
Line 4: Another line
Line 5: TODO: Add more tests
Line 6: Final line`
	
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Simple text search
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":    testFile,
		"pattern": "TODO",
	})
	if err != nil {
		t.Fatalf("Failed to search file: %v", err)
	}

	searchResult := result.(*FileSearchResult)
	if searchResult.TotalMatches != 2 {
		t.Errorf("Expected 2 matches, got %d", searchResult.TotalMatches)
	}
	if searchResult.FilesSearched != 1 {
		t.Errorf("Expected 1 file searched, got %d", searchResult.FilesSearched)
	}

	// Check first match
	if len(searchResult.Matches) > 0 {
		match := searchResult.Matches[0]
		if match.LineNumber != 3 {
			t.Errorf("Expected first match on line 3, got %d", match.LineNumber)
		}
		if !strings.Contains(match.Line, "Fix this bug") {
			t.Errorf("Expected match line to contain 'Fix this bug', got: %s", match.Line)
		}
	}

	// Test 2: Case-insensitive search
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":           testFile,
		"pattern":        "hello",
		"case_sensitive": false,
	})
	if err != nil {
		t.Fatalf("Failed case-insensitive search: %v", err)
	}

	searchResult = result.(*FileSearchResult)
	if searchResult.TotalMatches != 1 {
		t.Errorf("Expected 1 case-insensitive match, got %d", searchResult.TotalMatches)
	}

	// Test 3: No matches
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":    testFile,
		"pattern": "NOTFOUND",
	})
	if err != nil {
		t.Fatalf("Failed search with no matches: %v", err)
	}

	searchResult = result.(*FileSearchResult)
	if searchResult.TotalMatches != 0 {
		t.Errorf("Expected 0 matches, got %d", searchResult.TotalMatches)
	}
}

func TestFileSearchRegex(t *testing.T) {
	tool := FileSearch()
	ctx := context.Background()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.log")
	content := `2024-01-31 10:00:00 INFO: Application started
2024-01-31 10:00:01 ERROR: Connection failed
2024-01-31 10:00:02 WARNING: Retry attempt 1
2024-01-31 10:00:03 ERROR: Connection timeout
2024-01-31 10:00:04 INFO: Fallback mode activated`
	
	os.WriteFile(testFile, []byte(content), 0644)

	// Test regex search
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":     testFile,
		"pattern":  "ERROR:.*failed",
		"is_regex": true,
	})
	if err != nil {
		t.Fatalf("Failed regex search: %v", err)
	}

	searchResult := result.(*FileSearchResult)
	if searchResult.TotalMatches != 1 {
		t.Errorf("Expected 1 regex match, got %d", searchResult.TotalMatches)
	}

	// Test regex with line anchors
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":     testFile,
		"pattern":  "^\\d{4}-\\d{2}-\\d{2}.*ERROR",
		"is_regex": true,
	})
	if err != nil {
		t.Fatalf("Failed anchored regex search: %v", err)
	}

	searchResult = result.(*FileSearchResult)
	if searchResult.TotalMatches != 2 {
		t.Errorf("Expected 2 ERROR lines, got %d", searchResult.TotalMatches)
	}
}

func TestFileSearchContext(t *testing.T) {
	tool := FileSearch()
	ctx := context.Background()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := `Line 1
Line 2
Line 3: MATCH
Line 4
Line 5
Line 6: MATCH
Line 7
Line 8`
	
	os.WriteFile(testFile, []byte(content), 0644)

	// Search with context lines
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":          testFile,
		"pattern":       "MATCH",
		"context_lines": 2,
	})
	if err != nil {
		t.Fatalf("Failed search with context: %v", err)
	}

	searchResult := result.(*FileSearchResult)
	if len(searchResult.Matches) != 2 {
		t.Fatalf("Expected 2 matches, got %d", len(searchResult.Matches))
	}

	// Check first match context
	match1 := searchResult.Matches[0]
	if len(match1.ContextBefore) != 2 {
		t.Errorf("Expected 2 context lines before, got %d", len(match1.ContextBefore))
	}
	if len(match1.ContextAfter) != 2 {
		t.Errorf("Expected 2 context lines after, got %d", len(match1.ContextAfter))
	}
	
	// Verify context content
	if len(match1.ContextBefore) >= 2 {
		if match1.ContextBefore[0] != "Line 1" {
			t.Errorf("Expected first context line 'Line 1', got '%s'", match1.ContextBefore[0])
		}
		if match1.ContextBefore[1] != "Line 2" {
			t.Errorf("Expected second context line 'Line 2', got '%s'", match1.ContextBefore[1])
		}
	}
}

func TestFileSearchDirectory(t *testing.T) {
	tool := FileSearch()
	ctx := context.Background()

	// Create directory structure with multiple files
	tempDir := t.TempDir()
	
	files := map[string]string{
		"file1.txt": "TODO: First task",
		"file2.txt": "Nothing here",
		"file3.txt": "TODO: Second task",
		"subdir/file4.txt": "TODO: Nested task",
		"other.log": "TODO: Log task",
	}

	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte(content), 0644)
	}

	// Test 1: Non-recursive search
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":    tempDir,
		"pattern": "TODO",
	})
	if err != nil {
		t.Fatalf("Failed directory search: %v", err)
	}

	searchResult := result.(*FileSearchResult)
	if searchResult.TotalMatches != 3 { // Only root directory files
		t.Errorf("Expected 3 matches in root, got %d", searchResult.TotalMatches)
	}

	// Test 2: Recursive search
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":      tempDir,
		"pattern":   "TODO",
		"recursive": true,
	})
	if err != nil {
		t.Fatalf("Failed recursive search: %v", err)
	}

	searchResult = result.(*FileSearchResult)
	if searchResult.TotalMatches != 4 { // All files including subdirectory
		t.Errorf("Expected 4 matches recursively, got %d", searchResult.TotalMatches)
	}

	// Test 3: File pattern filtering
	result, err = tool.Execute(ctx, map[string]interface{}{
		"path":         tempDir,
		"pattern":      "TODO",
		"file_pattern": "*.txt",
		"recursive":    true,
	})
	if err != nil {
		t.Fatalf("Failed filtered search: %v", err)
	}

	searchResult = result.(*FileSearchResult)
	if searchResult.TotalMatches != 3 { // Only .txt files
		t.Errorf("Expected 3 matches in .txt files, got %d", searchResult.TotalMatches)
	}
}

func TestFileSearchMaxResults(t *testing.T) {
	tool := FileSearch()
	ctx := context.Background()

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	
	// Create file with many matches
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, fmt.Sprintf("Line %d: MATCH", i))
	}
	content := strings.Join(lines, "\n")
	os.WriteFile(testFile, []byte(content), 0644)

	// Search with max results limit
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":        testFile,
		"pattern":     "MATCH",
		"max_results": 10,
	})
	if err != nil {
		t.Fatalf("Failed search with max results: %v", err)
	}

	searchResult := result.(*FileSearchResult)
	if len(searchResult.Matches) != 10 {
		t.Errorf("Expected 10 matches with limit, got %d", len(searchResult.Matches))
	}
}

func TestFileSearchBinaryFiles(t *testing.T) {
	tool := FileSearch()
	ctx := context.Background()

	tempDir := t.TempDir()
	
	// Create a text file
	textFile := filepath.Join(tempDir, "text.txt")
	os.WriteFile(textFile, []byte("FINDME in text"), 0644)
	
	// Create a binary file with null bytes
	binaryFile := filepath.Join(tempDir, "binary.dat")
	binaryContent := []byte{0x00, 0x01, 0x02, 'F', 'I', 'N', 'D', 'M', 'E', 0x00}
	os.WriteFile(binaryFile, binaryContent, 0644)

	// Search should skip binary file
	result, err := tool.Execute(ctx, map[string]interface{}{
		"path":    tempDir,
		"pattern": "FINDME",
	})
	if err != nil {
		t.Fatalf("Failed search: %v", err)
	}

	searchResult := result.(*FileSearchResult)
	if searchResult.TotalMatches != 1 {
		t.Errorf("Expected 1 match (text file only), got %d", searchResult.TotalMatches)
	}
	if searchResult.FilesSearched != 1 {
		t.Errorf("Expected 1 file searched (binary skipped), got %d", searchResult.FilesSearched)
	}
}

func TestFileSearchErrors(t *testing.T) {
	tool := FileSearch()
	ctx := context.Background()

	// Test non-existent path
	_, err := tool.Execute(ctx, map[string]interface{}{
		"path":    "/non/existent/path",
		"pattern": "test",
	})
	if err == nil {
		t.Error("Expected error for non-existent path")
	}

	// Test invalid regex
	tempFile, _ := os.CreateTemp("", "test")
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	_, err = tool.Execute(ctx, map[string]interface{}{
		"path":     tempFile.Name(),
		"pattern":  "[invalid regex",
		"is_regex": true,
	})
	if err == nil {
		t.Error("Expected error for invalid regex")
	}
	if !strings.Contains(err.Error(), "invalid regex") {
		t.Errorf("Expected 'invalid regex' error, got: %v", err)
	}
}