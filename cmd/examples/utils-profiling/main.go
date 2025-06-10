// This file contains an example of how to use the profiling utilities.
// It is not imported by main.go, but can be compiled separately.

package main

// ABOUTME: Example demonstrating the profiling utilities for performance analysis
// ABOUTME: Shows how to capture CPU and memory profiles using the util/profiling package

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lexlapax/go-llms/pkg/util/profiling"
)

// logAdapter adapts standard log package to profiling.Logger interface
type logAdapter struct{}

func (logAdapter) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (logAdapter) Errorf(format string, v ...interface{}) {
	log.Printf("ERROR: "+format, v...)
}

// Example demonstrating the use of the profiling utilities in the Go-LLMs project
// Compile this separately with: go build -o bin/profile_example cmd/examples/metrics/profile_example.go
func main() {
	// Set up logging for profiling package
	profiling.SetLogger(logAdapter{})
	// Enable profiling via environment variable
	_ = os.Setenv("GO_LLMS_ENABLE_PROFILING", "1")

	// Set a custom profile output directory (optional)
	tempDir, err := os.MkdirTemp("", "profile_example")
	if err != nil {
		log.Printf("Warning: Could not create temp dir: %v\n", err)
	} else {
		profiling.SetProfileDir(tempDir)
		defer func() { _ = os.RemoveAll(tempDir) }() // Clean up when done
	}

	// Get the global profiler
	profiler := profiling.GetGlobalProfiler()

	// Profile some operations using the global profiler
	log.Println("Running simple profiled operations...")
	ctx := context.Background()

	// Profile structured extraction operation
	log.Println("\nProfiling structured extraction:")
	result, err := profiling.ProfileStructuredOp(ctx, profiling.OpStructuredExtraction, func(ctx context.Context) (interface{}, error) {
		// Simulate work - extract a "JSON" from a string
		log.Println("  Extracting JSON...")
		time.Sleep(20 * time.Millisecond) // Simulate work
		inputText := `Here's the data you requested: {"name":"John","age":30,"city":"New York"}`

		// Simple extraction (similar to the actual extractor)
		startIdx := strings.Index(inputText, "{")
		endIdx := strings.LastIndex(inputText, "}")
		if startIdx >= 0 && endIdx > startIdx {
			return inputText[startIdx : endIdx+1], nil
		}
		return "{}", nil
	})

	if err != nil {
		log.Printf("Error in extraction: %v\n", err)
	} else {
		log.Printf("  Extracted: %s\n", result)
	}

	// Profile schema validation operation
	log.Println("\nProfiling schema validation:")
	_, err = profiling.ProfileSchemaOp(ctx, profiling.OpSchemaValidation, func(ctx context.Context) (interface{}, error) {
		// Simulate schema validation
		log.Println("  Validating schema...")
		time.Sleep(30 * time.Millisecond) // Simulate work
		return true, nil
	})
	if err != nil {
		log.Printf("Error profiling schema validation: %v\n", err)
	}

	// Profile a component with the component enabler
	log.Println("\nProfiling a specific component:")
	disableFn := profiling.EnableProfilingForComponent("json_processor")

	// Start CPU profiling for the component
	err = profiler.StartCPUProfile()
	if err != nil {
		log.Printf("Error starting CPU profile: %v\n", err)
	}

	// Simulate component work
	log.Println("  Running JSON processing work...")
	time.Sleep(50 * time.Millisecond)

	// Stop CPU profiling
	profiler.StopCPUProfile()

	// Take a memory profile
	err = profiler.WriteHeapProfile()
	if err != nil {
		log.Printf("Error writing heap profile: %v\n", err)
	}

	// Disable profiling for the component
	disableFn()

	// Display where profiles were saved
	log.Printf("\nProfile files were saved to: %s\n", profiling.GetProfileDir())
	log.Println("You can view the profiles with 'go tool pprof [profile_file]'")
}
