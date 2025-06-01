// ABOUTME: Tests for web tool registration and basic functionality
// ABOUTME: Verifies web tools are properly registered in the registry

package web

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
)

func TestWebToolsRegistration(t *testing.T) {
	// Get all tools
	allTools := tools.Tools.List()
	webToolCount := 0
	webTools := make(map[string]bool)
	
	for _, entry := range allTools {
		if entry.Metadata.Category == "web" {
			webToolCount++
			webTools[entry.Metadata.Name] = true
		}
	}
	
	// We now have 4 web tools
	expectedTools := []string{
		"web_fetch",
		"web_search",
		"web_scrape",
		"http_request",
	}
	
	if webToolCount != len(expectedTools) {
		t.Errorf("Expected %d web tools, got %d", len(expectedTools), webToolCount)
		// List what we found
		t.Log("Found web tools:")
		for name := range webTools {
			t.Logf("  - %s", name)
		}
	}
	
	// Check each expected tool exists
	for _, toolName := range expectedTools {
		if !webTools[toolName] {
			t.Errorf("Expected web tool '%s' not found", toolName)
		}
		
		// Also verify it can be retrieved
		tool, found := tools.GetTool(toolName)
		if !found {
			t.Errorf("Web tool '%s' not retrievable", toolName)
		}
		if tool == nil {
			t.Errorf("Web tool '%s' is nil", toolName)
		}
		
		// Verify it has the correct name
		if found && tool != nil && tool.Name() != toolName {
			t.Errorf("Web tool name mismatch: expected '%s', got '%s'", toolName, tool.Name())
		}
	}
}