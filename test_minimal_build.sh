#!/bin/bash

echo "Testing build sizes with different configurations..."

# Current build
echo "1. Current build with all dependencies:"
go build -ldflags="-s -w" -o bin/go-llms-current ./cmd/
ls -lh bin/go-llms-current

# Build without kongplete (comment out in main.go temporarily)
echo -e "\n2. Build without shell completion (manual edit needed):"
echo "   Comment out kongplete imports and usage in main.go"

# Create a minimal go.mod for testing
echo -e "\n3. Creating minimal build..."
mkdir -p cmd_minimal
cat > cmd_minimal/main.go << 'EOF'
package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	version = "minimal"
	provider = flag.String("provider", "openai", "LLM provider")
	model = flag.String("model", "", "Model to use")
	verbose = flag.Bool("verbose", false, "Verbose output")
)

func main() {
	flag.Parse()
	
	if len(os.Args) < 2 {
		fmt.Println("Usage: go-llms [chat|complete|agent|structured]")
		os.Exit(1)
	}
	
	command := os.Args[1]
	fmt.Printf("Minimal build - Command: %s, Provider: %s\n", command, *provider)
}
EOF

go build -ldflags="-s -w" -o bin/go-llms-minimal ./cmd_minimal/
ls -lh bin/go-llms-minimal

# Compare sizes
echo -e "\n=== Size Comparison ==="
ls -lh bin/go-llms*