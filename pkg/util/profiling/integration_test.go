package profiling

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProfileStructuredOp(t *testing.T) {
	// Create temp dir for profile outputs
	tempDir, err := os.MkdirTemp("", "profiler_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Save and restore original profile dir
	origProfileDir := profileDir
	defer func() { profileDir = origProfileDir }()
	profileDir = tempDir
	
	// Enable global profiler
	globalProfiler = NewProfiler("structured_test")
	globalProfiler.Enable()
	
	// Test profiling structured operation
	ctx := context.Background()
	result, err := ProfileStructuredOp(ctx, OpStructuredExtraction, func(ctx context.Context) (interface{}, error) {
		// Simulate work
		time.Sleep(10 * time.Millisecond)
		return "extracted_data", nil
	})
	
	// Check result
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != "extracted_data" {
		t.Errorf("Expected result to be 'extracted_data', got: %v", result)
	}
	
	// Check profile files were created
	cpuFile := filepath.Join(tempDir, "structured_test_structured_extraction_cpu.pprof")
	memFile := filepath.Join(tempDir, "structured_test_structured_extraction_mem.pprof")
	
	if _, err := os.Stat(cpuFile); os.IsNotExist(err) {
		t.Error("Expected CPU profile file to be created")
	}
	if _, err := os.Stat(memFile); os.IsNotExist(err) {
		t.Error("Expected memory profile file to be created")
	}
}

func TestEnableProfilingForComponent(t *testing.T) {
	// Create temp dir for profile outputs
	tempDir, err := os.MkdirTemp("", "profiler_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Save and restore original profile dir
	origProfileDir := profileDir
	defer func() { profileDir = origProfileDir }()
	profileDir = tempDir
	
	// Enable profiling for a component
	disableFn := EnableProfilingForComponent("test_component")
	
	// Start CPU profiling manually to test
	profiler := GetGlobalProfiler()
	err = profiler.StartCPUProfile()
	if err != nil {
		t.Fatalf("Failed to start CPU profile: %v", err)
	}
	
	// Do some work
	doSomeCPUWork()
	
	// Stop CPU profiling
	profiler.StopCPUProfile()
	
	// Disable profiling
	disableFn()
	
	// Check that the GetProfileDir function returns the correct directory
	dir := GetProfileDir()
	if dir != tempDir {
		t.Errorf("Expected profile dir to be '%s', got '%s'", tempDir, dir)
	}
}