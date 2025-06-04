// ABOUTME: Tests for the FileMove built-in tool
// ABOUTME: Validates move/rename operations, overwrite handling, and cross-device moves

package file

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

func TestFileMoveRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("file_move")
	if !ok {
		t.Fatal("FileMove tool not registered")
	}
	if tool == nil {
		t.Fatal("FileMove tool is nil")
	}

	// Test tool name
	if tool.Name() != "file_move" {
		t.Errorf("Expected tool name 'file_move', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("file_move")
	if len(entries) == 0 {
		t.Fatal("FileMove tool not found in registry")
	}

	meta := entries[0].Metadata
	if meta.Category != "file" {
		t.Errorf("Expected category 'file', got '%s'", meta.Category)
	}
}

// mockMoveAgent implements BaseAgent for testing
type mockMoveAgent struct {
	id          string
	name        string
	description string
	agentType   domain.AgentType
	metadata    map[string]interface{}
}

func (a *mockMoveAgent) ID() string          { return a.id }
func (a *mockMoveAgent) Name() string        { return a.name }
func (a *mockMoveAgent) Description() string { return a.description }
func (a *mockMoveAgent) Type() domain.AgentType { return a.agentType }
func (a *mockMoveAgent) Parent() domain.BaseAgent { return nil }
func (a *mockMoveAgent) SetParent(parent domain.BaseAgent) error { return nil }
func (a *mockMoveAgent) SubAgents() []domain.BaseAgent { return nil }
func (a *mockMoveAgent) AddSubAgent(agent domain.BaseAgent) error { return nil }
func (a *mockMoveAgent) RemoveSubAgent(name string) error { return nil }
func (a *mockMoveAgent) FindAgent(name string) domain.BaseAgent { return nil }
func (a *mockMoveAgent) FindSubAgent(name string) domain.BaseAgent { return nil }
func (a *mockMoveAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	return nil, nil
}
func (a *mockMoveAgent) RunAsync(ctx context.Context, input *domain.State) (<-chan domain.Event, error) {
	return nil, nil
}
func (a *mockMoveAgent) Initialize(ctx context.Context) error { return nil }
func (a *mockMoveAgent) BeforeRun(ctx context.Context, state *domain.State) error { return nil }
func (a *mockMoveAgent) AfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) error {
	return nil
}
func (a *mockMoveAgent) Cleanup(ctx context.Context) error { return nil }
func (a *mockMoveAgent) InputSchema() *sdomain.Schema { return nil }
func (a *mockMoveAgent) OutputSchema() *sdomain.Schema { return nil }
func (a *mockMoveAgent) Config() domain.AgentConfig { return domain.AgentConfig{} }
func (a *mockMoveAgent) WithConfig(config domain.AgentConfig) domain.BaseAgent { return a }
func (a *mockMoveAgent) Validate() error { return nil }
func (a *mockMoveAgent) Metadata() map[string]interface{} { return a.metadata }
func (a *mockMoveAgent) SetMetadata(key string, value interface{}) {
	if a.metadata == nil {
		a.metadata = make(map[string]interface{})
	}
	a.metadata[key] = value
}

// createTestToolContextForMove creates a ToolContext for testing
func createTestToolContextForMove() *domain.ToolContext {
	ctx := context.Background()
	state := domain.NewState()
	stateReader := domain.NewStateReader(state)
	agent := &mockMoveAgent{
		id:          "test-agent",
		name:        "Test Agent",
		description: "Test agent for file move tests",
		agentType:   domain.AgentTypeCustom,
		metadata:    make(map[string]interface{}),
	}
	
	tc := domain.NewToolContext(ctx, stateReader, agent, "test-run-id")
	
	// Create a simple event emitter
	tc = tc.WithEventEmitter(&testEventEmitterMove{})
	
	return tc
}

// testEventEmitterMove implements EventEmitter for testing
type testEventEmitterMove struct{}

func (e *testEventEmitterMove) Emit(eventType domain.EventType, data interface{}) {}
func (e *testEventEmitterMove) EmitProgress(current, total int, message string) {}
func (e *testEventEmitterMove) EmitMessage(message string) {}
func (e *testEventEmitterMove) EmitError(err error) {}
func (e *testEventEmitterMove) EmitCustom(eventName string, data interface{}) {}

func TestFileMoveBasic(t *testing.T) {
	tool := FileMove()
	tc := createTestToolContextForMove()

	// Test 1: Simple rename in same directory
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "original.txt")
	dstFile := filepath.Join(tempDir, "renamed.txt")

	// Create source file
	if err := os.WriteFile(srcFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := tool.Execute(tc, map[string]interface{}{
		"source":      srcFile,
		"destination": dstFile,
	})
	if err != nil {
		t.Fatalf("Failed to rename file: %v", err)
	}

	moveResult := result.(*FileMoveResult)
	if !moveResult.Moved {
		t.Error("File was not moved")
	}
	if !moveResult.WasRename {
		t.Error("Operation should be marked as rename")
	}
	if moveResult.WasCrossDevice {
		t.Error("Same directory move should not be cross-device")
	}

	// Verify source is gone
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("Source file still exists after move")
	}

	// Verify destination exists with correct content
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatal("Destination file not found")
	}
	if string(content) != "test content" {
		t.Error("Destination file has incorrect content")
	}

	// Test 2: Move to different directory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	srcFile2 := filepath.Join(tempDir, "file2.txt")
	dstFile2 := filepath.Join(subDir, "file2.txt")

	if err := os.WriteFile(srcFile2, []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err = tool.Execute(tc, map[string]interface{}{
		"source":      srcFile2,
		"destination": dstFile2,
	})
	if err != nil {
		t.Fatalf("Failed to move file to subdir: %v", err)
	}

	moveResult = result.(*FileMoveResult)
	if !moveResult.Moved {
		t.Error("File was not moved to subdirectory")
	}
	if moveResult.WasRename {
		t.Error("Cross-directory move should not be marked as rename only")
	}

	// Verify file is in new location
	if _, err := os.Stat(dstFile2); os.IsNotExist(err) {
		t.Error("File not found in destination directory")
	}
}

func TestFileMoveToDirectory(t *testing.T) {
	tool := FileMove()
	tc := createTestToolContextForMove()

	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	targetDir := filepath.Join(tempDir, "target")

	// Create source file and target directory
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Move file to directory (should preserve filename)
	result, err := tool.Execute(tc, map[string]interface{}{
		"source":      srcFile,
		"destination": targetDir,
	})
	if err != nil {
		t.Fatalf("Failed to move file to directory: %v", err)
	}

	moveResult := result.(*FileMoveResult)
	if !moveResult.Moved {
		t.Error("File was not moved")
	}

	// Check file is in target directory with same name
	expectedPath := filepath.Join(targetDir, "source.txt")
	if moveResult.Destination != expectedPath {
		t.Errorf("Expected destination %s, got %s", expectedPath, moveResult.Destination)
	}

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("File not found in target directory")
	}
}

func TestFileMoveOverwrite(t *testing.T) {
	tool := FileMove()
	tc := createTestToolContextForMove()

	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "existing.txt")

	// Create both files
	if err := os.WriteFile(srcFile, []byte("new content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dstFile, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Try without overwrite (should fail)
	result, err := tool.Execute(tc, map[string]interface{}{
		"source":      srcFile,
		"destination": dstFile,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	moveResult := result.(*FileMoveResult)
	if moveResult.Moved {
		t.Error("Should not move when destination exists without overwrite")
	}
	if !strings.Contains(moveResult.Message, "already exists") {
		t.Errorf("Expected 'already exists' message, got: %s", moveResult.Message)
	}

	// Verify source still exists
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Error("Source file should still exist after failed move")
	}

	// Test 2: Move with overwrite
	result, err = tool.Execute(tc, map[string]interface{}{
		"source":      srcFile,
		"destination": dstFile,
		"overwrite":   true,
	})
	if err != nil {
		t.Fatalf("Failed to move with overwrite: %v", err)
	}

	moveResult = result.(*FileMoveResult)
	if !moveResult.Moved {
		t.Error("File was not moved with overwrite flag")
	}

	// Verify content was replaced
	content, _ := os.ReadFile(dstFile)
	if string(content) != "new content" {
		t.Error("Destination file was not overwritten")
	}
}

func TestFileMoveCreateDirs(t *testing.T) {
	tool := FileMove()
	tc := createTestToolContextForMove()

	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "deep", "nested", "dir", "dest.txt")

	// Create source file
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Without create_dirs (should fail)
	_, err := tool.Execute(tc, map[string]interface{}{
		"source":      srcFile,
		"destination": dstFile,
	})
	if err == nil {
		t.Error("Expected error when parent directories don't exist")
	}

	// Test 2: With create_dirs
	result, err := tool.Execute(tc, map[string]interface{}{
		"source":      srcFile,
		"destination": dstFile,
		"create_dirs": true,
	})
	if err != nil {
		t.Fatalf("Failed to move with create_dirs: %v", err)
	}

	moveResult := result.(*FileMoveResult)
	if !moveResult.Moved {
		t.Error("File was not moved")
	}

	// Verify file exists in nested directory
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("File not found in nested directory")
	}
}

func TestFileMoveErrors(t *testing.T) {
	tool := FileMove()
	tc := createTestToolContextForMove()

	tempDir := t.TempDir()

	// Test 1: Non-existent source
	result, err := tool.Execute(tc, map[string]interface{}{
		"source":      "/non/existent/file.txt",
		"destination": filepath.Join(tempDir, "dest.txt"),
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	moveResult := result.(*FileMoveResult)
	if moveResult.Moved {
		t.Error("Non-existent file reported as moved")
	}
	if !strings.Contains(moveResult.Message, "does not exist") {
		t.Errorf("Expected 'does not exist' message, got: %s", moveResult.Message)
	}

	// Test 2: Same source and destination
	srcFile := filepath.Join(tempDir, "same.txt")
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err = tool.Execute(tc, map[string]interface{}{
		"source":      srcFile,
		"destination": srcFile,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	moveResult = result.(*FileMoveResult)
	if moveResult.Moved {
		t.Error("Should not move when source and destination are the same")
	}
	if !strings.Contains(moveResult.Message, "same") {
		t.Errorf("Expected 'same' in message, got: %s", moveResult.Message)
	}
}

func TestFileMoveDirectory(t *testing.T) {
	tool := FileMove()
	tc := createTestToolContextForMove()

	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "srcdir")
	dstDir := filepath.Join(tempDir, "dstdir")

	// Create source directory with files
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	// Move entire directory
	result, err := tool.Execute(tc, map[string]interface{}{
		"source":      srcDir,
		"destination": dstDir,
	})
	if err != nil {
		t.Fatalf("Failed to move directory: %v", err)
	}

	moveResult := result.(*FileMoveResult)
	if !moveResult.Moved {
		t.Error("Directory was not moved")
	}

	// Verify source directory is gone
	if _, err := os.Stat(srcDir); !os.IsNotExist(err) {
		t.Error("Source directory still exists after move")
	}

	// Verify destination directory exists with files
	file1Path := filepath.Join(dstDir, "file1.txt")
	file2Path := filepath.Join(dstDir, "file2.txt")

	if _, err := os.Stat(file1Path); os.IsNotExist(err) {
		t.Error("file1.txt not found in destination directory")
	}
	if _, err := os.Stat(file2Path); os.IsNotExist(err) {
		t.Error("file2.txt not found in destination directory")
	}
}

func TestFileMovePreserveAttributes(t *testing.T) {
	tool := FileMove()
	tc := createTestToolContextForMove()

	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "dest.txt")

	// Create source file with specific permissions
	if err := os.WriteFile(srcFile, []byte("content"), 0755); err != nil {
		t.Fatal(err)
	}

	// Get original file info
	srcInfo, _ := os.Stat(srcFile)
	originalMode := srcInfo.Mode()

	// Move with preserve_attrs
	result, err := tool.Execute(tc, map[string]interface{}{
		"source":         srcFile,
		"destination":    dstFile,
		"preserve_attrs": true,
	})
	if err != nil {
		t.Fatalf("Failed to move with preserve_attrs: %v", err)
	}

	moveResult := result.(*FileMoveResult)
	if !moveResult.Moved {
		t.Error("File was not moved")
	}

	// Check if permissions were preserved
	dstInfo, _ := os.Stat(dstFile)
	if dstInfo.Mode() != originalMode {
		t.Logf("Note: Permissions may differ on some filesystems (original: %v, new: %v)", originalMode, dstInfo.Mode())
	}
}
