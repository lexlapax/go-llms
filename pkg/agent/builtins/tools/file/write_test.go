// ABOUTME: Tests for the file writing tool with all enhanced features
// ABOUTME: Verifies atomic operations, append mode, backup creation, and permission handling

package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
	"github.com/lexlapax/go-llms/pkg/agent/domain"
	sdomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// mockWriteAgent implements the BaseAgent interface for testing
type mockWriteAgent struct {
	id          string
	name        string
	description string
	agentType   domain.AgentType
	metadata    map[string]interface{}
}

func (a *mockWriteAgent) ID() string                                { return a.id }
func (a *mockWriteAgent) Name() string                              { return a.name }
func (a *mockWriteAgent) Description() string                       { return a.description }
func (a *mockWriteAgent) Type() domain.AgentType                    { return a.agentType }
func (a *mockWriteAgent) Parent() domain.BaseAgent                  { return nil }
func (a *mockWriteAgent) SetParent(parent domain.BaseAgent) error   { return nil }
func (a *mockWriteAgent) SubAgents() []domain.BaseAgent             { return nil }
func (a *mockWriteAgent) AddSubAgent(agent domain.BaseAgent) error  { return nil }
func (a *mockWriteAgent) RemoveSubAgent(name string) error          { return nil }
func (a *mockWriteAgent) FindAgent(name string) domain.BaseAgent    { return nil }
func (a *mockWriteAgent) FindSubAgent(name string) domain.BaseAgent { return nil }
func (a *mockWriteAgent) Run(ctx context.Context, input *domain.State) (*domain.State, error) {
	return nil, nil
}
func (a *mockWriteAgent) RunAsync(ctx context.Context, input *domain.State) (<-chan domain.Event, error) {
	return nil, nil
}
func (a *mockWriteAgent) Initialize(ctx context.Context) error                     { return nil }
func (a *mockWriteAgent) BeforeRun(ctx context.Context, state *domain.State) error { return nil }
func (a *mockWriteAgent) AfterRun(ctx context.Context, state *domain.State, result *domain.State, err error) error {
	return nil
}
func (a *mockWriteAgent) Cleanup(ctx context.Context) error                     { return nil }
func (a *mockWriteAgent) InputSchema() *sdomain.Schema                          { return nil }
func (a *mockWriteAgent) OutputSchema() *sdomain.Schema                         { return nil }
func (a *mockWriteAgent) Config() domain.AgentConfig                            { return domain.AgentConfig{} }
func (a *mockWriteAgent) WithConfig(config domain.AgentConfig) domain.BaseAgent { return a }
func (a *mockWriteAgent) Validate() error                                       { return nil }
func (a *mockWriteAgent) Metadata() map[string]interface{}                      { return a.metadata }
func (a *mockWriteAgent) SetMetadata(key string, value interface{}) {
	if a.metadata == nil {
		a.metadata = make(map[string]interface{})
	}
	a.metadata[key] = value
}

// createTestToolContextForWrite creates a ToolContext for testing
func createTestToolContextForWrite() *domain.ToolContext {
	ctx := context.Background()
	state := domain.NewState()
	stateReader := domain.NewStateReader(state)
	agent := &mockWriteAgent{
		id:          "test-agent",
		name:        "Test Agent",
		description: "Test agent for file write tests",
		agentType:   domain.AgentTypeCustom,
		metadata:    make(map[string]interface{}),
	}

	tc := domain.NewToolContext(ctx, stateReader, agent, "test-run-id")

	// Create a simple event emitter
	tc = tc.WithEventEmitter(&testWriteEventEmitter{})

	return tc
}

// testWriteEventEmitter implements EventEmitter for testing
type testWriteEventEmitter struct{}

func (e *testWriteEventEmitter) Emit(eventType domain.EventType, data interface{}) {}
func (e *testWriteEventEmitter) EmitProgress(current, total int, message string)   {}
func (e *testWriteEventEmitter) EmitMessage(message string)                        {}
func (e *testWriteEventEmitter) EmitError(err error)                               {}
func (e *testWriteEventEmitter) EmitCustom(eventName string, data interface{})     {}

func TestWriteFileRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("file_write")
	if !ok {
		t.Fatal("WriteFile tool not registered")
	}
	if tool == nil {
		t.Fatal("WriteFile tool is nil")
	}

	// Test tool name
	if tool.Name() != "file_write" {
		t.Errorf("Expected tool name 'file_write', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("file_write")
	if len(entries) == 0 {
		t.Fatal("WriteFile tool not found in registry")
	}

	meta := entries[0].Metadata
	if meta.Category != "file" {
		t.Errorf("Expected category 'file', got '%s'", meta.Category)
	}
}

func TestWriteFile_Basic(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "output.txt")
	testContent := "Hello from WriteFile!"

	tool := MustGetWriteFile()
	ctx := createTestToolContextForWrite()

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
	ctx := createTestToolContextForWrite()

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
	ctx := createTestToolContextForWrite()

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
	ctx := createTestToolContextForWrite()

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
	ctx := createTestToolContextForWrite()

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
	ctx := createTestToolContextForWrite()

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

func TestWriteFile_OverwriteProtection(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "protected.txt")

	// Create initial file
	initialContent := "Existing content"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := MustGetWriteFile()
	ctx := createTestToolContextForWrite()

	// Write without append should overwrite
	newContent := "Overwritten content"
	result, err := tool.Execute(ctx, WriteFileParams{
		Path:    testFile,
		Content: newContent,
	})
	if err != nil {
		t.Fatalf("Failed to overwrite file: %v", err)
	}

	writeResult := result.(*WriteFileResult)
	if !writeResult.FileExisted {
		t.Error("File should have existed before overwrite")
	}

	// Verify content was overwritten
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != newContent {
		t.Errorf("Expected content %q, got %q", newContent, string(content))
	}
}

func TestWriteFile_InvalidPath(t *testing.T) {
	tool := MustGetWriteFile()
	ctx := createTestToolContextForWrite()

	// Test writing to invalid path
	_, err := tool.Execute(ctx, WriteFileParams{
		Path:    "/root/cannot-write-here.txt",
		Content: "test",
	})
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestWriteFile_EmptyContent(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "empty.txt")

	tool := MustGetWriteFile()
	ctx := createTestToolContextForWrite()

	// Write empty content
	result, err := tool.Execute(ctx, WriteFileParams{
		Path:    testFile,
		Content: "",
	})
	if err != nil {
		t.Fatalf("Failed to write empty file: %v", err)
	}

	writeResult := result.(*WriteFileResult)
	if !writeResult.Success {
		t.Error("Write operation failed")
	}
	if writeResult.BytesWritten != 0 {
		t.Errorf("Expected 0 bytes written, got %d", writeResult.BytesWritten)
	}

	// Verify file exists and is empty
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() != 0 {
		t.Errorf("Expected empty file, got size %d", info.Size())
	}
}
