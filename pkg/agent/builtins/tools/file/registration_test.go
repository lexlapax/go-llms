// ABOUTME: Tests for file tool registration and basic functionality
// ABOUTME: Verifies tools are properly registered in the registry

package file

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
)

func TestFileToolsRegistration(t *testing.T) {
	// Check that file_read is registered
	readTool, found := tools.GetTool("file_read")
	if !found {
		t.Error("file_read tool not found in registry")
	}
	if readTool == nil {
		t.Error("file_read tool is nil")
	}
	if readTool.Name() != "file_read" {
		t.Errorf("Expected tool name 'file_read', got '%s'", readTool.Name())
	}

	// Check that file_write is registered
	writeTool, found := tools.GetTool("file_write")
	if !found {
		t.Error("file_write tool not found in registry")
	}
	if writeTool == nil {
		t.Error("file_write tool is nil")
	}
	if writeTool.Name() != "file_write" {
		t.Errorf("Expected tool name 'file_write', got '%s'", writeTool.Name())
	}

	// Check metadata
	// Note: ListByCategory is not available on the ToolRegistry interface
	// We'll check by listing all tools instead
	allTools := tools.Tools.List()
	fileToolCount := 0
	for _, entry := range allTools {
		if entry.Metadata.Category == "file" {
			fileToolCount++
		}
	}
	// We now have 6 file tools: read, write, list, delete, move, search
	if fileToolCount != 6 {
		t.Errorf("Expected 6 file tools, got %d", fileToolCount)
	}

	// Verify permissions
	for _, entry := range allTools {
		if entry.Metadata.Name == "file_read" {
			// This would check permissions if we stored full metadata
			if entry.Metadata.Description == "" {
				t.Error("file_read missing description")
			}
		}
	}
}
