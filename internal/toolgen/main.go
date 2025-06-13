package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var (
		inputDir    = flag.String("input", "pkg/agent/builtins/tools", "Directory containing tool implementations")
		outputFile  = flag.String("output", "pkg/agent/tools/registry_metadata.go", "Output file for generated metadata")
		factoryFile = flag.String("factory", "pkg/agent/tools/registry_factories.go", "Output file for factory implementations")
		verbose     = flag.Bool("v", false, "Verbose output")
	)
	flag.Parse()

	if *verbose {
		log.Printf("Scanning directory: %s", *inputDir)
	}

	// Find all subdirectories with tools
	var allMetadata []ToolMetadata

	entries, err := os.ReadDir(*inputDir)
	if err != nil {
		log.Fatalf("Error reading directory %s: %v", *inputDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip special directories
		if entry.Name() == "testdata" || entry.Name() == "internal" {
			continue
		}

		subdir := filepath.Join(*inputDir, entry.Name())
		if *verbose {
			log.Printf("Processing subdirectory: %s", subdir)
		}

		metadata, err := ParseDirectory(subdir)
		if err != nil {
			log.Printf("Warning: error parsing %s: %v", subdir, err)
			continue
		}

		if *verbose {
			log.Printf("Found %d tools in %s", len(metadata), entry.Name())
		}

		allMetadata = append(allMetadata, metadata...)
	}

	if len(allMetadata) == 0 {
		log.Fatal("No tools found")
	}

	log.Printf("Found %d tools total", len(allMetadata))

	// Generate metadata file
	generator := NewGenerator()
	metadataCode, err := generator.GenerateMetadataFile(allMetadata)
	if err != nil {
		log.Fatalf("Error generating metadata file: %v", err)
	}

	// Write metadata file
	if err := writeFile(*outputFile, metadataCode); err != nil {
		log.Fatalf("Error writing metadata file: %v", err)
	}
	log.Printf("Generated metadata file: %s", *outputFile)

	// Generate factory file
	factoryCode, err := generator.GenerateFactoryFile(allMetadata)
	if err != nil {
		log.Fatalf("Error generating factory file: %v", err)
	}

	// Write factory file
	if err := writeFile(*factoryFile, factoryCode); err != nil {
		log.Fatalf("Error writing factory file: %v", err)
	}
	log.Printf("Generated factory file: %s", *factoryFile)

	// Print summary
	fmt.Println("\nTool Summary:")
	categoryCount := make(map[string]int)
	for _, tool := range allMetadata {
		categoryCount[tool.Category]++
		if *verbose {
			fmt.Printf("  - %s (%s): %s\n", tool.Name, tool.Category, tool.Description)
		}
	}

	fmt.Println("\nTools by category:")
	for category, count := range categoryCount {
		fmt.Printf("  %s: %d\n", category, count)
	}
}

func writeFile(filename, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}
