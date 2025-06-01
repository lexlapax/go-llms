// ABOUTME: Tests for system tool registration and basic functionality
// ABOUTME: Verifies system tools are properly registered in the registry

package system

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
)

func TestSystemToolsRegistration(t *testing.T) {
	// Get all tools
	allTools := tools.Tools.List()
	systemToolCount := 0
	systemTools := make(map[string]bool)
	
	for _, entry := range allTools {
		if entry.Metadata.Category == "system" {
			systemToolCount++
			systemTools[entry.Metadata.Name] = true
		}
	}
	
	// We now have 4 system tools
	expectedTools := []string{
		"execute_command",
		"get_environment_variable",
		"get_system_info",
		"process_list",
	}
	
	if systemToolCount != len(expectedTools) {
		t.Errorf("Expected %d system tools, got %d", len(expectedTools), systemToolCount)
	}
	
	// Check each expected tool exists
	for _, toolName := range expectedTools {
		if !systemTools[toolName] {
			t.Errorf("Expected system tool '%s' not found", toolName)
		}
		
		// Also verify it can be retrieved
		tool, found := tools.GetTool(toolName)
		if !found {
			t.Errorf("System tool '%s' not retrievable", toolName)
		}
		if tool == nil {
			t.Errorf("System tool '%s' is nil", toolName)
		}
	}
}