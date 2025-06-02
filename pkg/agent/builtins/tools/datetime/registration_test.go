package datetime

import (
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
)

func TestToolRegistration(t *testing.T) {
	// Test that datetime_now is registered
	tool, found := tools.GetTool("datetime_now")
	if !found {
		t.Error("datetime_now tool should be registered")
	}
	if tool == nil {
		t.Error("datetime_now tool should not be nil")
	}

	// Test registry search
	results := tools.Tools.Search("datetime")
	if len(results) == 0 {
		t.Error("Should find at least one datetime tool")
	}

	// Test category listing
	categoryTools := tools.Tools.ListByCategory("datetime")
	if len(categoryTools) == 0 {
		t.Error("Should find at least one tool in datetime category")
	}
}
