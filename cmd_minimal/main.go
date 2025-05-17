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
