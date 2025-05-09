// Package profiling provides utilities for CPU and memory profiling in the Go-LLMs project.
// It allows profiling of specific operations and components to identify performance bottlenecks.
package profiling

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"
)

// Default directory for storing profiling output
var profileDir = filepath.Join(os.TempDir(), "go-llms-profiles")

// Global profiler instance for easy access
var (
	globalProfiler *Profiler
	globalMu       sync.Mutex
)

// Profiler implements CPU and memory profiling functionality
type Profiler struct {
	name       string
	enabled    bool
	cpuFile    *os.File
	cpuRunning bool
	mu         sync.Mutex
}

// NewProfiler creates a new profiler with the given name
// The name will be used as part of the output file names
func NewProfiler(name string) *Profiler {
	// Create the profiler
	p := &Profiler{
		name:    name,
		enabled: IsProfilingEnabled(),
	}

	// Create profile directory if it doesn't exist
	if p.enabled {
		ensureProfileDir()
	}

	return p
}

// Enable enables profiling
func (p *Profiler) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
	ensureProfileDir()
}

// Disable disables profiling
func (p *Profiler) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Stop CPU profiling if it's running
	if p.cpuRunning && p.cpuFile != nil {
		pprof.StopCPUProfile()
		p.cpuFile.Close()
		p.cpuFile = nil
		p.cpuRunning = false
	}
	
	p.enabled = false
}

// IsEnabled returns true if profiling is enabled
func (p *Profiler) IsEnabled() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.enabled
}

// StartCPUProfile starts CPU profiling and writes to a file named [name].pprof
func (p *Profiler) StartCPUProfile() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if !p.enabled {
		return nil
	}
	
	// Stop any existing CPU profile
	if p.cpuRunning {
		pprof.StopCPUProfile()
		if p.cpuFile != nil {
			p.cpuFile.Close()
			p.cpuFile = nil
		}
		p.cpuRunning = false
	}
	
	// Create the CPU profile file
	cpuFilePath := filepath.Join(profileDir, fmt.Sprintf("%s.pprof", p.name))
	var err error
	p.cpuFile, err = os.Create(cpuFilePath)
	if err != nil {
		return fmt.Errorf("could not create CPU profile: %v", err)
	}
	
	// Start the CPU profile
	if err := pprof.StartCPUProfile(p.cpuFile); err != nil {
		p.cpuFile.Close()
		p.cpuFile = nil
		return fmt.Errorf("could not start CPU profile: %v", err)
	}
	
	p.cpuRunning = true
	return nil
}

// StopCPUProfile stops CPU profiling if it's running
func (p *Profiler) StopCPUProfile() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.cpuRunning {
		pprof.StopCPUProfile()
		if p.cpuFile != nil {
			p.cpuFile.Close()
			p.cpuFile = nil
		}
		p.cpuRunning = false
	}
}

// WriteHeapProfile writes the heap profile to a file named [name].pprof
func (p *Profiler) WriteHeapProfile() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if !p.enabled {
		return nil
	}
	
	// Create the memory profile file
	memFilePath := filepath.Join(profileDir, fmt.Sprintf("%s_mem.pprof", p.name))
	f, err := os.Create(memFilePath)
	if err != nil {
		return fmt.Errorf("could not create memory profile: %v", err)
	}
	defer f.Close()

	// Write the heap profile
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write memory profile: %v", err)
	}
	
	return nil
}

// ProfileOperation profiles a function execution, capturing CPU and memory profiles
// It returns the result of the function and any error that occurred
func (p *Profiler) ProfileOperation(ctx context.Context, opName string, fn func(context.Context) (interface{}, error)) (interface{}, error) {
	if !p.IsEnabled() {
		// If profiling is disabled, just run the function
		return fn(ctx)
	}
	
	// We create operation-specific file names but don't need a separate profiler instance
	// opName is used directly in file paths below
	
	// Start CPU profiling
	cpuFilePath := filepath.Join(profileDir, fmt.Sprintf("%s_%s_cpu.pprof", p.name, opName))
	cpuFile, err := os.Create(cpuFilePath)
	if err != nil {
		// Log the error but continue with the operation
		fmt.Fprintf(os.Stderr, "Warning: could not create CPU profile for %s: %v\n", opName, err)
	} else {
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not start CPU profile for %s: %v\n", opName, err)
			cpuFile.Close()
		} else {
			defer func() {
				pprof.StopCPUProfile()
				cpuFile.Close()
			}()
		}
	}
	
	// Run the operation
	startTime := time.Now()
	result, err := fn(ctx)
	duration := time.Since(startTime)
	
	// Write memory profile
	memFilePath := filepath.Join(profileDir, fmt.Sprintf("%s_%s_mem.pprof", p.name, opName))
	memFile, memErr := os.Create(memFilePath)
	if memErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not create memory profile for %s: %v\n", opName, memErr)
	} else {
		if memErr := pprof.WriteHeapProfile(memFile); memErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not write memory profile for %s: %v\n", opName, memErr)
		}
		memFile.Close()
	}
	
	// Log duration and return result
	fmt.Printf("Operation %s completed in %v\n", opName, duration)
	return result, err
}

// GetGlobalProfiler returns the global profiler instance
// It creates a new global profiler if one doesn't exist yet
func GetGlobalProfiler() *Profiler {
	globalMu.Lock()
	defer globalMu.Unlock()
	
	if globalProfiler == nil {
		globalProfiler = NewProfiler("global")
	}
	
	return globalProfiler
}

// SetProfileDir sets the directory where profile files will be written
// If the directory doesn't exist or isn't writable, it will keep the current directory
func SetProfileDir(dir string) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: profile directory %s doesn't exist\n", dir)
		return
	}
	
	// Check if it's writable
	testFile := filepath.Join(dir, ".profiler_test")
	f, err := os.Create(testFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: profile directory %s isn't writable: %v\n", dir, err)
		return
	}
	f.Close()
	os.Remove(testFile)
	
	// Set the profile directory
	profileDir = dir
}

// IsProfilingEnabled checks if profiling is enabled via environment variable
func IsProfilingEnabled() bool {
	return os.Getenv("GO_LLMS_ENABLE_PROFILING") == "1"
}

// ensureProfileDir ensures the profile directory exists
func ensureProfileDir() {
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		if err := os.MkdirAll(profileDir, 0755); err != nil {
			// If we can't create the profile directory, use the temp directory directly
			profileDir = os.TempDir()
		}
	}
}