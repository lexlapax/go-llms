// ABOUTME: Tests for the ProcessList built-in tool
// ABOUTME: Validates process listing functionality and cross-platform behavior

package system

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
)

func TestProcessListRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("process_list")
	if !ok {
		t.Fatal("ProcessList tool not registered")
	}
	if tool == nil {
		t.Fatal("ProcessList tool is nil")
	}

	// Test tool name
	if tool.Name() != "process_list" {
		t.Errorf("Expected tool name 'process_list', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("process_list")
	if len(entries) == 0 {
		t.Fatal("ProcessList tool not found in registry")
	}

	meta := entries[0].Metadata
	if meta.Category != "system" {
		t.Errorf("Expected category 'system', got '%s'", meta.Category)
	}
}

func TestProcessListBasic(t *testing.T) {
	tool := ProcessList()
	ctx := createTestToolContext()

	// Test basic process listing
	result, err := tool.Execute(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}

	procResult := result.(*ProcessListResult)

	// Should have at least one process
	if procResult.Count == 0 {
		t.Error("Expected at least one process")
	}

	// Platform should match
	if procResult.Platform != runtime.GOOS {
		t.Errorf("Expected platform '%s', got '%s'", runtime.GOOS, procResult.Platform)
	}

	// Timestamp should be set
	if procResult.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}

	// Validate process info
	for i, proc := range procResult.Processes {
		if proc.PID <= 0 {
			t.Errorf("Process %d has invalid PID: %d", i, proc.PID)
		}
		if proc.Name == "" {
			t.Errorf("Process %d has empty name", i)
		}
	}
}

func TestProcessListIncludeSelf(t *testing.T) {
	tool := ProcessList()
	ctx := createTestToolContext()
	currentPID := os.Getpid()

	// Test with include_self = true
	result, err := tool.Execute(ctx, map[string]interface{}{
		"include_self": true,
	})
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}

	procResult := result.(*ProcessListResult)
	foundSelf := false

	for _, proc := range procResult.Processes {
		if proc.PID == currentPID {
			foundSelf = true
			break
		}
	}

	if !foundSelf {
		t.Error("Expected to find current process when include_self is true")
	}

	// Test with include_self = false (default)
	result, err = tool.Execute(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}

	procResult = result.(*ProcessListResult)
	foundSelf = false

	for _, proc := range procResult.Processes {
		if proc.PID == currentPID {
			foundSelf = true
			break
		}
	}

	if foundSelf {
		t.Error("Should not find current process when include_self is false")
	}
}

func TestProcessListFilter(t *testing.T) {
	tool := ProcessList()
	ctx := createTestToolContext()

	// Filter by a process name that should exist (use "go" as it's running tests)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"filter":       "go",
		"include_self": true,
	})
	if err != nil {
		t.Fatalf("Failed to filter processes: %v", err)
	}

	procResult := result.(*ProcessListResult)

	// All results should contain "go"
	for _, proc := range procResult.Processes {
		nameUpper := strings.ToUpper(proc.Name)
		commandUpper := strings.ToUpper(proc.Command)
		if !strings.Contains(nameUpper, "GO") && !strings.Contains(commandUpper, "GO") {
			t.Errorf("Process %s doesn't match filter 'go'", proc.Name)
		}
	}
}

func TestProcessListSort(t *testing.T) {
	tool := ProcessList()
	ctx := createTestToolContext()

	// Test PID sorting
	result, err := tool.Execute(ctx, map[string]interface{}{
		"sort_by": "pid",
		"limit":   10,
	})
	if err != nil {
		t.Fatalf("Failed to sort processes: %v", err)
	}

	procResult := result.(*ProcessListResult)

	// Verify PID order
	for i := 1; i < len(procResult.Processes); i++ {
		if procResult.Processes[i-1].PID > procResult.Processes[i].PID {
			t.Error("Processes not sorted by PID in ascending order")
			break
		}
	}

	// Test name sorting
	result, err = tool.Execute(ctx, map[string]interface{}{
		"sort_by": "name",
		"limit":   10,
	})
	if err != nil {
		t.Fatalf("Failed to sort processes by name: %v", err)
	}

	procResult = result.(*ProcessListResult)

	// Verify name order (case-insensitive)
	for i := 1; i < len(procResult.Processes); i++ {
		if strings.ToLower(procResult.Processes[i-1].Name) > strings.ToLower(procResult.Processes[i].Name) {
			t.Errorf("Processes not sorted by name: %s > %s",
				procResult.Processes[i-1].Name, procResult.Processes[i].Name)
			break
		}
	}
}

func TestProcessListLimit(t *testing.T) {
	tool := ProcessList()
	ctx := createTestToolContext()

	// Test with limit
	limit := 5
	result, err := tool.Execute(ctx, map[string]interface{}{
		"limit": limit,
	})
	if err != nil {
		t.Fatalf("Failed to limit processes: %v", err)
	}

	procResult := result.(*ProcessListResult)

	if len(procResult.Processes) > limit {
		t.Errorf("Expected at most %d processes, got %d", limit, len(procResult.Processes))
	}

	// Count should match actual processes returned
	if procResult.Count != len(procResult.Processes) {
		t.Errorf("Count mismatch: reported %d, actual %d", procResult.Count, len(procResult.Processes))
	}
}

func TestExtractProcessName(t *testing.T) {
	testCases := []struct {
		command  string
		expected string
	}{
		{"/usr/bin/go test", "go"},
		{"C:\\Windows\\System32\\app.exe", "app"},
		{"python3 script.py", "python3"},
		{"./myapp", "myapp"},
		{" /path/to/binary ", "binary"},
		{"", "unknown"},
		{"   ", "unknown"},
	}

	for _, tc := range testCases {
		result := extractProcessName(tc.command)
		if result != tc.expected {
			t.Errorf("Extract process name from '%s': expected '%s', got '%s'",
				tc.command, tc.expected, result)
		}
	}
}

func TestParseCSVLine(t *testing.T) {
	testCases := []struct {
		line     string
		expected []string
	}{
		{
			`"field1","field2","field3"`,
			[]string{"field1", "field2", "field3"},
		},
		{
			`"field with comma, inside","normal field","field3"`,
			[]string{"field with comma, inside", "normal field", "field3"},
		},
		{
			`field1,field2,field3`,
			[]string{"field1", "field2", "field3"},
		},
		{
			`"quoted",unquoted,"quoted again"`,
			[]string{"quoted", "unquoted", "quoted again"},
		},
	}

	for _, tc := range testCases {
		result := parseCSVLine(tc.line)
		if len(result) != len(tc.expected) {
			t.Errorf("CSV parse length mismatch for '%s': expected %d, got %d",
				tc.line, len(tc.expected), len(result))
			continue
		}

		for i, field := range result {
			if field != tc.expected[i] {
				t.Errorf("CSV field %d mismatch: expected '%s', got '%s'",
					i, tc.expected[i], field)
			}
		}
	}
}

func TestSortFunctions(t *testing.T) {
	// Create test processes
	processes := []ProcessInfo{
		{PID: 100, Name: "zebra", CPUPercent: 5.0, MemoryUsage: 1000},
		{PID: 50, Name: "alpha", CPUPercent: 10.0, MemoryUsage: 2000},
		{PID: 75, Name: "beta", CPUPercent: 2.5, MemoryUsage: 500},
	}

	// Test PID sort
	pidCopy := make([]ProcessInfo, len(processes))
	copy(pidCopy, processes)
	sortByPID(pidCopy)

	if pidCopy[0].PID != 50 || pidCopy[1].PID != 75 || pidCopy[2].PID != 100 {
		t.Error("PID sort failed")
	}

	// Test name sort
	nameCopy := make([]ProcessInfo, len(processes))
	copy(nameCopy, processes)
	sortByName(nameCopy)

	if nameCopy[0].Name != "alpha" || nameCopy[1].Name != "beta" || nameCopy[2].Name != "zebra" {
		t.Error("Name sort failed")
	}

	// Test CPU sort (descending)
	cpuCopy := make([]ProcessInfo, len(processes))
	copy(cpuCopy, processes)
	sortByCPU(cpuCopy)

	if cpuCopy[0].CPUPercent != 10.0 || cpuCopy[1].CPUPercent != 5.0 || cpuCopy[2].CPUPercent != 2.5 {
		t.Error("CPU sort failed")
	}

	// Test memory sort (descending)
	memCopy := make([]ProcessInfo, len(processes))
	copy(memCopy, processes)
	sortByMemory(memCopy)

	if memCopy[0].MemoryUsage != 2000 || memCopy[1].MemoryUsage != 1000 || memCopy[2].MemoryUsage != 500 {
		t.Error("Memory sort failed")
	}
}
