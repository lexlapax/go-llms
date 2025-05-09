package profiling

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewProfiler(t *testing.T) {
	profiler := NewProfiler("test")
	if profiler == nil {
		t.Fatal("Expected non-nil profiler")
	}
	
	if profiler.name != "test" {
		t.Errorf("Expected name to be 'test', got '%s'", profiler.name)
	}
	
	if profiler.enabled {
		t.Error("Expected profiler to be disabled by default")
	}
}

func TestEnableDisable(t *testing.T) {
	profiler := NewProfiler("test")
	
	// Test initial state
	if profiler.IsEnabled() {
		t.Error("Expected profiler to be disabled initially")
	}
	
	// Test enabling
	profiler.Enable()
	if !profiler.IsEnabled() {
		t.Error("Expected profiler to be enabled after Enable() call")
	}
	
	// Test disabling
	profiler.Disable()
	if profiler.IsEnabled() {
		t.Error("Expected profiler to be disabled after Disable() call")
	}
}

func TestProfilingWithEnvVar(t *testing.T) {
	// Save original env and restore after test
	origEnv := os.Getenv("GO_LLMS_ENABLE_PROFILING")
	defer os.Setenv("GO_LLMS_ENABLE_PROFILING", origEnv)
	
	// Test with env var enabled
	os.Setenv("GO_LLMS_ENABLE_PROFILING", "1")
	profiler := NewProfiler("test_env")
	if !profiler.IsEnabled() {
		t.Error("Expected profiler to be enabled when GO_LLMS_ENABLE_PROFILING=1")
	}
	
	// Test with env var disabled
	os.Setenv("GO_LLMS_ENABLE_PROFILING", "0")
	profiler = NewProfiler("test_env2")
	if profiler.IsEnabled() {
		t.Error("Expected profiler to be disabled when GO_LLMS_ENABLE_PROFILING=0")
	}
}

func TestStartCPUProfile(t *testing.T) {
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
	
	profiler := NewProfiler("cpu_test")
	profiler.Enable()
	
	// Start CPU profiling
	err = profiler.StartCPUProfile()
	if err != nil {
		t.Fatalf("Failed to start CPU profile: %v", err)
	}
	
	// Check that CPU profile file was created
	cpuFile := filepath.Join(tempDir, "cpu_test.pprof")
	if _, err := os.Stat(cpuFile); os.IsNotExist(err) {
		t.Error("Expected CPU profile file to be created")
	}
	
	// Stop CPU profiling
	profiler.StopCPUProfile()
	
	// Verify file has content
	info, err := os.Stat(cpuFile)
	if err != nil {
		t.Fatalf("Error stating CPU profile file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Expected CPU profile file to have content")
	}
}

func TestMemoryProfile(t *testing.T) {
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
	
	profiler := NewProfiler("mem_test")
	profiler.Enable()
	
	// Take memory profile
	err = profiler.WriteHeapProfile()
	if err != nil {
		t.Fatalf("Failed to write heap profile: %v", err)
	}
	
	// Check that memory profile file was created
	memFile := filepath.Join(tempDir, "mem_test_mem.pprof")
	if _, err := os.Stat(memFile); os.IsNotExist(err) {
		t.Error("Expected memory profile file to be created")
	}

	// Verify file has content
	info, err := os.Stat(memFile)
	if err != nil {
		t.Fatalf("Error stating memory profile file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Expected memory profile file to have content")
	}
}

func TestProfileOperation(t *testing.T) {
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
	
	profiler := NewProfiler("op_test")
	profiler.Enable()
	
	// Test profiling an operation
	ctx := context.Background()
	result, err := profiler.ProfileOperation(ctx, "test_operation", func(ctx context.Context) (interface{}, error) {
		// Simulate work
		time.Sleep(10 * time.Millisecond)
		return "result", nil
	})
	
	// Check operation result
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != "result" {
		t.Errorf("Expected result to be 'result', got: %v", result)
	}
	
	// Check that profile files were created
	cpuFile := filepath.Join(tempDir, "op_test_test_operation_cpu.pprof")
	memFile := filepath.Join(tempDir, "op_test_test_operation_mem.pprof")
	
	if _, err := os.Stat(cpuFile); os.IsNotExist(err) {
		t.Error("Expected CPU profile file to be created")
	}
	if _, err := os.Stat(memFile); os.IsNotExist(err) {
		t.Error("Expected memory profile file to be created")
	}
}

func TestProfileOperationWithDisabledProfiler(t *testing.T) {
	profiler := NewProfiler("disabled_test")
	// Don't enable the profiler
	
	ctx := context.Background()
	result, err := profiler.ProfileOperation(ctx, "test_operation", func(ctx context.Context) (interface{}, error) {
		return "result", nil
	})
	
	// Function should still work even though profiling is disabled
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != "result" {
		t.Errorf("Expected result to be 'result', got: %v", result)
	}
}

func TestGetGlobalProfiler(t *testing.T) {
	// Reset global profiler
	globalProfiler = nil
	
	// First call should create a new one
	p1 := GetGlobalProfiler()
	if p1 == nil {
		t.Fatal("Expected non-nil global profiler")
	}
	
	// Second call should return the same instance
	p2 := GetGlobalProfiler()
	if p1 != p2 {
		t.Error("Expected the same global profiler instance")
	}
}

func TestSetProfileDir(t *testing.T) {
	// Create temp dir for profile outputs
	tempDir, err := os.MkdirTemp("", "profiler_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Test setting profile directory
	SetProfileDir(tempDir)
	if profileDir != tempDir {
		t.Errorf("Expected profile dir to be '%s', got '%s'", tempDir, profileDir)
	}
	
	// Test with invalid directory (should keep original)
	invalidDir := "/path/that/does/not/exist"
	originalDir := profileDir
	SetProfileDir(invalidDir)
	if profileDir != originalDir {
		t.Errorf("Expected profile dir to remain '%s', got '%s'", originalDir, profileDir)
	}
}

func TestIsProfilingEnabled(t *testing.T) {
	// Save original env and restore after test
	origEnv := os.Getenv("GO_LLMS_ENABLE_PROFILING")
	defer os.Setenv("GO_LLMS_ENABLE_PROFILING", origEnv)
	
	// Test with env var enabled
	os.Setenv("GO_LLMS_ENABLE_PROFILING", "1")
	if !IsProfilingEnabled() {
		t.Error("Expected IsProfilingEnabled() to return true when GO_LLMS_ENABLE_PROFILING=1")
	}
	
	// Test with env var disabled
	os.Setenv("GO_LLMS_ENABLE_PROFILING", "0")
	if IsProfilingEnabled() {
		t.Error("Expected IsProfilingEnabled() to return false when GO_LLMS_ENABLE_PROFILING=0")
	}
	
	// Test with env var unset
	os.Unsetenv("GO_LLMS_ENABLE_PROFILING")
	if IsProfilingEnabled() {
		t.Error("Expected IsProfilingEnabled() to return false when GO_LLMS_ENABLE_PROFILING is unset")
	}
}

// Helper test to verify if profiling is actually working
func TestActualProfilingOutput(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping actual profiling test in short mode")
	}
	
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
	
	profiler := NewProfiler("actual_test")
	profiler.Enable()
	
	// Start CPU profiling
	err = profiler.StartCPUProfile()
	if err != nil {
		t.Fatalf("Failed to start CPU profile: %v", err)
	}
	
	// Do some measurable work
	doSomeCPUWork()
	
	// Stop CPU profiling
	profiler.StopCPUProfile()
	
	// Take memory profile
	doSomeMemoryWork()
	err = profiler.WriteHeapProfile()
	if err != nil {
		t.Fatalf("Failed to write heap profile: %v", err)
	}
	
	// Check that profile files have meaningful content
	cpuFile := filepath.Join(tempDir, "actual_test.pprof")
	memFile := filepath.Join(tempDir, "actual_test_mem.pprof")

	// Read CPU profile file
	cpuData, err := os.ReadFile(cpuFile)
	if err != nil {
		t.Fatalf("Error reading CPU profile file: %v", err)
	}

	// Verify the file looks like a pprof output
	if len(cpuData) < 100 {
		t.Error("CPU profile file seems too small to be valid")
	}
	
	// Check memory profile file size
	memInfo, err := os.Stat(memFile)
	if err != nil {
		t.Fatalf("Error stating memory profile file: %v", err)
	}
	if memInfo.Size() < 100 {
		t.Error("Memory profile file seems too small to be valid")
	}
}

// Some CPU-intensive work for testing
func doSomeCPUWork() {
	for i := 0; i < 10000000; i++ {
		_ = i * i
	}
}

// Some memory-intensive work for testing
func doSomeMemoryWork() {
	var data [][]byte
	for i := 0; i < 1000; i++ {
		data = append(data, make([]byte, 1000))
	}
	_ = data
}