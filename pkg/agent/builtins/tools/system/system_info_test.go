// ABOUTME: Tests for the GetSystemInfo built-in tool
// ABOUTME: Validates system information retrieval and parameter handling

package system

import (
	"context"
	"runtime"
	"testing"

	"github.com/lexlapax/go-llms/pkg/agent/builtins/tools"
)

func TestGetSystemInfoRegistration(t *testing.T) {
	// Test that the tool is registered
	tool, ok := tools.GetTool("get_system_info")
	if !ok {
		t.Fatal("GetSystemInfo tool not registered")
	}
	if tool == nil {
		t.Fatal("GetSystemInfo tool is nil")
	}

	// Test tool name
	if tool.Name() != "get_system_info" {
		t.Errorf("Expected tool name 'get_system_info', got '%s'", tool.Name())
	}

	// Test metadata
	entries := tools.Tools.Search("get_system_info")
	if len(entries) == 0 {
		t.Fatal("GetSystemInfo tool not found in registry")
	}
	
	meta := entries[0].Metadata
	if meta.Category != "system" {
		t.Errorf("Expected category 'system', got '%s'", meta.Category)
	}
}

func TestGetSystemInfoBasic(t *testing.T) {
	tool := GetSystemInfo()
	ctx := context.Background()

	// Test 1: Basic system info (no optional fields)
	result, err := tool.Execute(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to get basic system info: %v", err)
	}

	sysInfo := result.(*SystemInfo)
	
	// Validate required fields
	if sysInfo.OS.Name != runtime.GOOS {
		t.Errorf("Expected OS name '%s', got '%s'", runtime.GOOS, sysInfo.OS.Name)
	}
	
	if sysInfo.Architecture != runtime.GOARCH {
		t.Errorf("Expected architecture '%s', got '%s'", runtime.GOARCH, sysInfo.Architecture)
	}
	
	if sysInfo.CPUs != runtime.NumCPU() {
		t.Errorf("Expected %d CPUs, got %d", runtime.NumCPU(), sysInfo.CPUs)
	}
	
	// Platform name should be set
	if sysInfo.OS.Platform == "" {
		t.Error("Expected platform name to be set")
	}
	
	// Timestamp should be set
	if sysInfo.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}
	
	// Optional fields should be nil when not requested
	if sysInfo.Memory != nil {
		t.Error("Expected memory to be nil when not requested")
	}
	if sysInfo.Runtime != nil {
		t.Error("Expected runtime to be nil when not requested")
	}
	if sysInfo.Environment != nil {
		t.Error("Expected environment to be nil when not requested")
	}
}

func TestGetSystemInfoWithMemory(t *testing.T) {
	tool := GetSystemInfo()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"include_memory": true,
	})
	if err != nil {
		t.Fatalf("Failed to get system info with memory: %v", err)
	}

	sysInfo := result.(*SystemInfo)
	
	// Memory should be included
	if sysInfo.Memory == nil {
		t.Fatal("Expected memory info to be included")
	}
	
	// Validate memory fields
	if sysInfo.Memory.Alloc == 0 {
		t.Error("Expected non-zero allocated memory")
	}
	if sysInfo.Memory.Sys == 0 {
		t.Error("Expected non-zero system memory")
	}
	// TotalAlloc and NumGC could be 0 in a fresh process, so we don't check those
}

func TestGetSystemInfoWithRuntime(t *testing.T) {
	tool := GetSystemInfo()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"include_runtime": true,
	})
	if err != nil {
		t.Fatalf("Failed to get system info with runtime: %v", err)
	}

	sysInfo := result.(*SystemInfo)
	
	// Runtime should be included
	if sysInfo.Runtime == nil {
		t.Fatal("Expected runtime info to be included")
	}
	
	// Validate runtime fields
	if sysInfo.Runtime.Version != runtime.Version() {
		t.Errorf("Expected Go version '%s', got '%s'", runtime.Version(), sysInfo.Runtime.Version)
	}
	if sysInfo.Runtime.Compiler != runtime.Compiler {
		t.Errorf("Expected compiler '%s', got '%s'", runtime.Compiler, sysInfo.Runtime.Compiler)
	}
	if sysInfo.Runtime.NumCPU != runtime.NumCPU() {
		t.Errorf("Expected %d CPUs, got %d", runtime.NumCPU(), sysInfo.Runtime.NumCPU)
	}
	if sysInfo.Runtime.NumGoroutine < 1 {
		t.Error("Expected at least 1 goroutine")
	}
	if sysInfo.Runtime.GOMAXPROCS < 1 {
		t.Error("Expected GOMAXPROCS to be at least 1")
	}
}

func TestGetSystemInfoWithEnvironment(t *testing.T) {
	tool := GetSystemInfo()
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{
		"include_environment": true,
	})
	if err != nil {
		t.Fatalf("Failed to get system info with environment: %v", err)
	}

	sysInfo := result.(*SystemInfo)
	
	// Environment should be included
	if sysInfo.Environment == nil {
		t.Fatal("Expected environment info to be included")
	}
	
	// Validate environment fields
	if sysInfo.Environment.TempDir == "" {
		t.Error("Expected temp directory to be set")
	}
	if sysInfo.Environment.TotalEnvVars == 0 {
		t.Error("Expected at least some environment variables")
	}
	
	// Working directory should normally be set
	if sysInfo.Environment.WorkingDir == "" {
		t.Log("Warning: Working directory is empty")
	}
}

func TestGetSystemInfoFull(t *testing.T) {
	tool := GetSystemInfo()
	ctx := context.Background()

	// Test with all optional fields
	result, err := tool.Execute(ctx, map[string]interface{}{
		"include_memory":      true,
		"include_runtime":     true,
		"include_environment": true,
	})
	if err != nil {
		t.Fatalf("Failed to get full system info: %v", err)
	}

	sysInfo := result.(*SystemInfo)
	
	// All optional fields should be present
	if sysInfo.Memory == nil {
		t.Error("Expected memory info to be included")
	}
	if sysInfo.Runtime == nil {
		t.Error("Expected runtime info to be included")
	}
	if sysInfo.Environment == nil {
		t.Error("Expected environment info to be included")
	}
}

func TestPlatformNames(t *testing.T) {
	testCases := []struct {
		goos     string
		expected string
	}{
		{"darwin", "macOS"},
		{"linux", "Linux"},
		{"windows", "Windows"},
		{"freebsd", "FreeBSD"},
		{"android", "Android"},
		{"unknown", "unknown"}, // Should return as-is
	}

	for _, tc := range testCases {
		result := getPlatformName(tc.goos)
		if result != tc.expected {
			t.Errorf("Platform name for %s: expected %s, got %s", tc.goos, tc.expected, result)
		}
	}
}

func TestCountPathDirs(t *testing.T) {
	testCases := []struct {
		path     string
		goos     string
		expected int
	}{
		{"/usr/bin:/usr/local/bin:/home/user/bin", "linux", 3},
		{"C:\\Windows;C:\\Program Files;C:\\Users", "windows", 3},
		{"/single/path", "linux", 1},
		{"", "linux", 0},
		{":", "linux", 2}, // Empty paths
		{"::", "linux", 3}, // Multiple empty paths
	}

	originalGOOS := runtime.GOOS
	for _, tc := range testCases {
		// This is a bit hacky since we can't actually change runtime.GOOS
		// In real code, we'd pass the separator as a parameter
		var result int
		if tc.goos == "windows" && originalGOOS != "windows" {
			// Simulate Windows path counting
			result = 1
			for i := 0; i < len(tc.path); i++ {
				if tc.path[i] == ';' {
					result++
				}
			}
			if tc.path == "" {
				result = 0
			}
		} else {
			result = countPathDirs(tc.path)
		}
		
		if tc.goos == originalGOOS && result != tc.expected {
			t.Errorf("Path dirs for '%s': expected %d, got %d", tc.path, tc.expected, result)
		}
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test splitLines
	lines := splitLines("line1\nline2\nline3")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
		t.Error("splitLines produced incorrect output")
	}

	// Test hasPrefix
	if !hasPrefix("hello world", "hello") {
		t.Error("hasPrefix failed for valid prefix")
	}
	if hasPrefix("hello", "hello world") {
		t.Error("hasPrefix returned true for invalid prefix")
	}

	// Test trimQuotes
	if trimQuotes(`"quoted"`) != "quoted" {
		t.Error("trimQuotes failed to remove quotes")
	}
	if trimQuotes("unquoted") != "unquoted" {
		t.Error("trimQuotes modified unquoted string")
	}
	if trimQuotes(`"single`) != `"single` {
		t.Error("trimQuotes removed incomplete quotes")
	}
}